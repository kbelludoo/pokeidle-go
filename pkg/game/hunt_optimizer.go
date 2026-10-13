package game

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// HuntMetric armazena métricas medidas de desempenho em uma rota de caça
type HuntMetric struct {
	Slug        string
	Name        string
	Level       int
	Kills       int
	Catches     int
	XPGained    int64
	TimeSeconds float64
	ExpPerHour  float64
	LastUpdated time.Time
}

// HuntOptimizer gerencia a otimização inteligente de hunts e análise de fraquezas
type HuntOptimizer struct {
	mu          sync.RWMutex
	metrics     map[string]*HuntMetric
	currentSlug string
	huntStart   time.Time
	huntStartXP int64
}

// NewHuntOptimizer cria um novo otimizador de caça
func NewHuntOptimizer() *HuntOptimizer {
	return &HuntOptimizer{
		metrics: make(map[string]*HuntMetric),
	}
}

// OnHuntChanged registra troca de área de caça para medir o delta de XP
func (o *HuntOptimizer) OnHuntChanged(newSlug string, currentXP int64) {
	o.mu.Lock()
	defer o.mu.Unlock()

	now := time.Now()
	// Salva a métrica da rota anterior se houve tempo decorrido
	if o.currentSlug != "" && !o.huntStart.IsZero() && currentXP > o.huntStartXP {
		elapsed := now.Sub(o.huntStart).Seconds()
		if elapsed > 15 {
			xpDelta := currentXP - o.huntStartXP
			m, exists := o.metrics[o.currentSlug]
			if !exists {
				m = &HuntMetric{Slug: o.currentSlug}
				o.metrics[o.currentSlug] = m
			}
			m.XPGained += xpDelta
			m.TimeSeconds += elapsed
			if m.TimeSeconds > 0 {
				m.ExpPerHour = float64(m.XPGained) / (m.TimeSeconds / 3600.0)
			}
			m.LastUpdated = now
		}
	}

	o.currentSlug = newSlug
	o.huntStart = now
	o.huntStartXP = currentXP
}

// RecordKill atualiza abates na rota atual
func (o *HuntOptimizer) RecordKill(slug string, xpDelta int64) {
	o.mu.Lock()
	defer o.mu.Unlock()

	m, exists := o.metrics[slug]
	if !exists {
		m = &HuntMetric{Slug: slug, LastUpdated: time.Now()}
		o.metrics[slug] = m
	}
	m.Kills++
	m.XPGained += xpDelta
	if !o.huntStart.IsZero() {
		m.TimeSeconds = time.Since(o.huntStart).Seconds()
		if m.TimeSeconds > 0 {
			m.ExpPerHour = float64(m.XPGained) / (m.TimeSeconds / 3600.0)
		}
	}
}

// GetHuntDefenderTypes infere os tipos elementais dos monstros de uma hunt
func GetHuntDefenderTypes(h PokeHuntArea) []string {
	// 1. Tenta extrair a partir do slug ou nome
	slug := strings.ToLower(h.Slug)
	nome := strings.ToLower(h.Nome)

	combined := slug + " " + nome

	if strings.Contains(combined, "geodude") || strings.Contains(combined, "onix") || strings.Contains(combined, "rocha") || strings.Contains(combined, "rock") {
		return []string{TypeRock, TypeGround}
	}
	if strings.Contains(combined, "diglett") || strings.Contains(combined, "sandshrew") || strings.Contains(combined, "terra") || strings.Contains(combined, "cave") {
		return []string{TypeGround}
	}
	if strings.Contains(combined, "pidgey") || strings.Contains(combined, "spearow") || strings.Contains(combined, "voador") {
		return []string{TypeNormal, TypeFlying}
	}
	if strings.Contains(combined, "caterpie") || strings.Contains(combined, "weedle") || strings.Contains(combined, "floresta") || strings.Contains(combined, "viridian") {
		return []string{TypeBug, TypeGrass}
	}
	if strings.Contains(combined, "zubat") || strings.Contains(combined, "ekans") || strings.Contains(combined, "veneno") {
		return []string{TypePoison}
	}
	if strings.Contains(combined, "magikarp") || strings.Contains(combined, "poliwag") || strings.Contains(combined, "agua") || strings.Contains(combined, "lago") || strings.Contains(combined, "water") {
		return []string{TypeWater}
	}
	if strings.Contains(combined, "charmander") || strings.Contains(combined, "vulpix") || strings.Contains(combined, "fogo") || strings.Contains(combined, "vulcao") {
		return []string{TypeFire}
	}
	if strings.Contains(combined, "gastly") || strings.Contains(combined, "torre") || strings.Contains(combined, "fantasma") {
		return []string{TypeGhost, TypePoison}
	}
	if strings.Contains(combined, "machop") || strings.Contains(combined, "mankey") || strings.Contains(combined, "lutador") {
		return []string{TypeFighting}
	}
	if strings.Contains(combined, "pikachu") || strings.Contains(combined, "usina") || strings.Contains(combined, "eletrico") {
		return []string{TypeElectric}
	}
	if strings.Contains(combined, "abra") || strings.Contains(combined, "drowzee") || strings.Contains(combined, "psiquico") {
		return []string{TypePsychic}
	}

	return []string{TypeNormal}
}

// MatchupResult detalha a eficácia de um Pokémon contra uma hunt
type MatchupResult struct {
	PokemonID       int64
	PokemonName     string
	PokemonTypes    []string
	DefenderTypes   []string
	Multiplier      float64
	MatchupLabel    string
	BasePower       int
	EffectivePower  float64
}

// EvaluatePokemonAgainstHunt calcula a força efetiva de um Pokémon contra os tipos da hunt
func EvaluatePokemonAgainstHunt(p PokeItemPokemon, defenderTypes []string) MatchupResult {
	name := p.Nome
	if name == "" {
		name = p.Name
	}
	atkTypes := GetPokemonTypes(p.PokeID, name)
	mult, label := MatchupDescription(atkTypes, defenderTypes)

	effectivePower := float64(p.Poder) * mult
	if p.Shiny {
		effectivePower *= 1.20 // Bônus tático por velocidade de shiny
	}

	return MatchupResult{
		PokemonID:      p.ID,
		PokemonName:    name,
		PokemonTypes:   atkTypes,
		DefenderTypes:  defenderTypes,
		Multiplier:     mult,
		MatchupLabel:   label,
		BasePower:      p.Poder,
		EffectivePower: effectivePower,
	}
}

// PickBestPokemonForHunt escolhe o Pokémon com maior DANO EFETIVO considerando fraqueza
func PickBestPokemonForHunt(pokemons map[int64]PokeItemPokemon, hunt PokeHuntArea, trainerLevel int) (PokeItemPokemon, MatchupResult, bool) {
	defenderTypes := GetHuntDefenderTypes(hunt)

	var candidates []PokeItemPokemon
	for _, p := range pokemons {
		// Regra de uso do Pokéidle: Starter sempre utilizável; demais até trainerLevel + 5
		if p.Starter || (trainerLevel > 0 && p.Level <= trainerLevel+5) || trainerLevel == 0 {
			candidates = append(candidates, p)
		}
	}

	if len(candidates) == 0 {
		return PokeItemPokemon{}, MatchupResult{}, false
	}

	bestScore := -1.0
	var bestPk PokeItemPokemon
	var bestMatchup MatchupResult

	for _, p := range candidates {
		m := EvaluatePokemonAgainstHunt(p, defenderTypes)
		if m.EffectivePower > bestScore {
			bestScore = m.EffectivePower
			bestPk = p
			bestMatchup = m
		}
	}

	return bestPk, bestMatchup, true
}

// FindBestHuntForMaxXP avalia todas as hunts acessíveis e escolhe a que maximiza EXP/h
func (o *HuntOptimizer) FindBestHuntForMaxXP(hunts []PokeHuntArea, pokemons map[int64]PokeItemPokemon, trainerLevel int) (PokeHuntArea, MatchupResult, string) {
	if len(hunts) == 0 {
		return PokeHuntArea{}, MatchupResult{}, "Nenhuma hunt disponível"
	}

	var unlockedHunts []PokeHuntArea
	for _, h := range hunts {
		if trainerLevel >= h.Nivel {
			unlockedHunts = append(unlockedHunts, h)
		}
	}

	if len(unlockedHunts) == 0 {
		// Se nenhuma desbloqueada por nível, usa a primeira
		unlockedHunts = append(unlockedHunts, hunts[0])
	}

	bestScore := -1.0
	var bestHunt PokeHuntArea
	var bestMatchup MatchupResult
	var rationale string

	for _, h := range unlockedHunts {
		_, matchup, ok := PickBestPokemonForHunt(pokemons, h, trainerLevel)
		if !ok {
			continue
		}

		// Score preditivo de XP/h:
		// Base level da hunt * Dano Efetivo * Fator de velocidade
		levelFactor := float64(h.Nivel)
		if levelFactor < 1 {
			levelFactor = 1
		}

		// Penalidade se a hunt for muito baixa (XP muito fraco)
		if trainerLevel > 15 && h.Nivel < trainerLevel-8 {
			levelFactor *= 0.5
		}

		// Bônus elemental: Se o jogador tem 2x ou mais de vantagem, o monstro morre na metade do tempo
		speedBonus := 1.0
		if matchup.Multiplier >= 2.0 {
			speedBonus = 2.2 // Super efetivo duplica a velocidade de farm
		} else if matchup.Multiplier > 1.0 {
			speedBonus = 1.3
		} else if matchup.Multiplier < 1.0 {
			speedBonus = 0.5 // Desvantagem de tipo torna a hunt lenta e perigosa
		}

		score := (levelFactor * 100.0) + (matchup.EffectivePower * 10.0 * speedBonus)

		// Se já temos medição empírica real de EXP/h dessa rota, dá peso prioritário
		o.mu.RLock()
		if m, exists := o.metrics[h.Slug]; exists && m.ExpPerHour > 0 {
			score += m.ExpPerHour * 0.1
		}
		o.mu.RUnlock()

		if score > bestScore {
			bestScore = score
			bestHunt = h
			bestMatchup = matchup
			rationale = fmt.Sprintf("Hunt '%s' (Nv. %d): Vantagem %s com %s (Dano Efetivo: %.0f)",
				h.Nome, h.Nivel, matchup.MatchupLabel, matchup.PokemonName, matchup.EffectivePower)
		}
	}

	return bestHunt, bestMatchup, rationale
}
