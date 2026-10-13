package game

import (
	"fmt"
	"sort"
	"strconv"
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

// PokeGenerations define todas as 11 faixas/gerações da Pokédex oficial do Pokéidle World
var PokeGenerations = []GenerationInfo{
	{ID: 1, Name: "Geração 1", Region: "Kanto", MinID: 1, MaxID: 151, Total: 151},
	{ID: 2, Name: "Geração 2", Region: "Johto", MinID: 152, MaxID: 251, Total: 100},
	{ID: 3, Name: "Geração 3", Region: "Hoenn", MinID: 252, MaxID: 386, Total: 135},
	{ID: 4, Name: "Geração 4", Region: "Sinnoh", MinID: 387, MaxID: 493, Total: 107},
	{ID: 5, Name: "Geração 5", Region: "Unova", MinID: 494, MaxID: 649, Total: 156},
	{ID: 6, Name: "Geração 6", Region: "Kalos", MinID: 650, MaxID: 721, Total: 72},
	{ID: 7, Name: "Geração 7", Region: "Alola", MinID: 722, MaxID: 809, Total: 88},
	{ID: 8, Name: "Geração 8", Region: "Galar", MinID: 810, MaxID: 905, Total: 96},
	{ID: 9, Name: "Geração 9", Region: "Paldea", MinID: 906, MaxID: 1025, Total: 120},
	{ID: 10, Name: "Outland", Region: "Outland", MinID: 2001, MaxID: 2058, Total: 58},
	{ID: 11, Name: "Mega Evoluções", Region: "Megas", MinID: 3001, MaxID: 3999, Total: 48},
}

// GetGenerationForPokeID retorna a geração e região correspondente a um pokeId
func GetGenerationForPokeID(pokeID int) *GenerationInfo {
	for i := range PokeGenerations {
		if pokeID >= PokeGenerations[i].MinID && pokeID <= PokeGenerations[i].MaxID {
			return &PokeGenerations[i]
		}
	}
	return nil
}

// CalculateAllBestiaries gera o relatório completo de todos os bestiários (Nacional, 1-9, Outland, Megas, Shiny e Bosses)
func CalculateAllBestiaries(pokedex map[string]PokeDexEntry, bossState *PokeBossState) BestiaryFullReport {
	report := BestiaryFullReport{
		NationalTotal: 1131, // Total de espécies catalogadas no Pokéidle
		BossesTotal:   87,   // 87 chefes mundiais da Arena
	}

	gens := make([]GenerationInfo, len(PokeGenerations))
	copy(gens, PokeGenerations)

	if pokedex != nil {
		for key, entry := range pokedex {
			id, err := strconv.Atoi(key)
			if err != nil {
				continue
			}

			report.TotalKills += int64(entry.K)
			report.TotalCaptures += int64(entry.C)

			isCaptured := entry.C > 0
			isSeen := entry.K > 0 || entry.C > 0

			if isCaptured {
				report.NationalCaptured++
			}
			if isSeen {
				report.NationalSeen++
			}

			// Shiny Dex
			if entry.SC > 0 {
				report.ShinysCaptured++
			}
			if entry.SV > 0 || entry.SC > 0 {
				report.ShinysSeen++
			}

			// Atualiza a geração correspondente
			for i := range gens {
				if id >= gens[i].MinID && id <= gens[i].MaxID {
					if isCaptured {
						gens[i].Captured++
					}
					if isSeen {
						gens[i].Seen++
					}
					break
				}
			}
		}
	}

	// Porcentagens por geração
	for i := range gens {
		if gens[i].Total > 0 {
			gens[i].Pct = (float64(gens[i].Captured) / float64(gens[i].Total)) * 100.0
			if gens[i].Pct > 100.0 {
				gens[i].Pct = 100.0
			}
			gens[i].Complete = gens[i].Captured >= gens[i].Total
		}
	}
	report.Generations = gens

	// Porcentagem nacional global
	if report.NationalTotal > 0 {
		report.NationalPct = (float64(report.NationalCaptured) / float64(report.NationalTotal)) * 100.0
		if report.NationalPct > 100.0 {
			report.NationalPct = 100.0
		}
	}

	// Shiny Dex Pct
	if report.NationalTotal > 0 {
		report.ShinysPct = (float64(report.ShinysCaptured) / float64(report.NationalTotal)) * 100.0
	}

	// Boss Dex
	if bossState != nil && bossState.Vitorias != nil {
		report.BossesDefeated = len(bossState.Vitorias)
		if report.BossesTotal > 0 {
			report.BossesPct = (float64(report.BossesDefeated) / float64(report.BossesTotal)) * 100.0
			if report.BossesPct > 100.0 {
				report.BossesPct = 100.0
			}
		}
	}

	return report
}

// CountBestiaryProgress calcula o total de espécies capturadas, vistas e porcentagem nacional
func CountBestiaryProgress(pokedex map[string]PokeDexEntry) (int, int, float64) {
	if pokedex == nil {
		return 0, 0, 0.0
	}
	captured := 0
	seen := 0
	for _, entry := range pokedex {
		if entry.C > 0 {
			captured++
		}
		if entry.K > 0 || entry.C > 0 {
			seen++
		}
	}
	const totalNacional = 1131
	pct := (float64(captured) / float64(totalNacional)) * 100.0
	if pct > 100.0 {
		pct = 100.0
	}
	return captured, seen, pct
}

// HuntHasUnregisteredSpecies verifica se uma rota contém monstros que ainda não foram capturados no Bestiário
func HuntHasUnregisteredSpecies(h PokeHuntArea, pokedex map[string]PokeDexEntry) bool {
	if pokedex == nil {
		return false
	}
	// Se a rota tiver espécies explícitas enviadas pelo servidor
	if len(h.Especies) > 0 {
		for _, esp := range h.Especies {
			key := fmt.Sprintf("%d", esp.PokeID)
			entry, exists := pokedex[key]
			if !exists || entry.C == 0 {
				return true // Espécie ainda pendente de registro de captura no Bestiário!
			}
		}
		return false
	}

	// Fallback por slug se o monstro estiver no nome da hunt
	slugLower := strings.ToLower(h.Slug + " " + h.Nome)
	unregisteredKeyNames := map[string]int{
		"geodude": 74, "pidgey": 16, "caterpie": 10, "rattata": 19,
		"zubat": 41, "diglett": 50, "magikarp": 129, "poliwag": 60,
		"charmander": 4, "bulbasaur": 1, "squirtle": 7, "pikachu": 25,
		"gastly": 92, "machop": 66, "abra": 63, "onix": 95,
	}
	for name, id := range unregisteredKeyNames {
		if strings.Contains(slugLower, name) {
			key := fmt.Sprintf("%d", id)
			entry, exists := pokedex[key]
			if !exists || entry.C == 0 {
				return true
			}
		}
	}
	return false
}

// FindBestHuntForMaxXP avalia todas as hunts acessíveis e escolhe a que maximiza EXP/h com bônus de Bestiário
func (o *HuntOptimizer) FindBestHuntForMaxXP(hunts []PokeHuntArea, pokemons map[int64]PokeItemPokemon, trainerLevel int, pokedex map[string]PokeDexEntry, focusBestiary bool) (PokeHuntArea, MatchupResult, string) {
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

		// BÔNUS DO BESTIÁRIO: Prioriza rotas que têm monstros ainda não registrados
		hasPending := HuntHasUnregisteredSpecies(h, pokedex)
		if hasPending {
			if focusBestiary {
				score += 350.0 // Prioridade máxima para fechar o Bestiário!
			} else {
				score += 75.0 // Bônus natural de descoberta e registro
			}
		}

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
			bestiaryTag := ""
			if hasPending {
				bestiaryTag = " • 📖 Monstro Novo p/ Bestiário!"
			}
			rationale = fmt.Sprintf("Hunt '%s' (Nv. %d): Vantagem %s com %s (Dano Efetivo: %.0f)%s",
				h.Nome, h.Nivel, matchup.MatchupLabel, matchup.PokemonName, matchup.EffectivePower, bestiaryTag)
		}
	}

	return bestHunt, bestMatchup, rationale
}

// FindBestHuntForPokemon encontra a melhor rota especificamente vantajosa para um Pokémon individual com foco em Bestiário
func (o *HuntOptimizer) FindBestHuntForPokemon(hunts []PokeHuntArea, p PokeItemPokemon, trainerLevel int, pokedex map[string]PokeDexEntry, focusBestiary bool) (PokeHuntArea, MatchupResult, bool) {
	if len(hunts) == 0 {
		return PokeHuntArea{}, MatchupResult{}, false
	}

	var unlocked []PokeHuntArea
	for _, h := range hunts {
		if trainerLevel >= h.Nivel {
			unlocked = append(unlocked, h)
		}
	}
	if len(unlocked) == 0 {
		unlocked = append(unlocked, hunts[0])
	}

	bestScore := -999.0
	var bestHunt PokeHuntArea
	var bestMatchup MatchupResult
	found := false

	for _, h := range unlocked {
		defenderTypes := GetHuntDefenderTypes(h)
		m := EvaluatePokemonAgainstHunt(p, defenderTypes)

		// Se a hunt for muito alta para o nível atual do Pokémon, não arrisca morrer
		if h.Nivel > p.Level+4 && !p.Starter {
			continue
		}

		score := float64(h.Nivel) * 10.0
		if m.Multiplier >= 2.0 {
			score += 150.0 // Super Efetivo! Vantagem máxima
		} else if m.Multiplier > 1.0 {
			score += 50.0
		} else if m.Multiplier < 1.0 {
			score -= 80.0 // Evita desvantagem de tipo
		}

		// Bônus de Bestiário
		if HuntHasUnregisteredSpecies(h, pokedex) {
			if focusBestiary {
				score += 250.0
			} else {
				score += 60.0
			}
		}

		// Considera métrica empírica se já houver
		o.mu.RLock()
		if metric, exists := o.metrics[h.Slug]; exists && metric.ExpPerHour > 0 {
			score += metric.ExpPerHour * 0.05
		}
		o.mu.RUnlock()

		if score > bestScore {
			bestScore = score
			bestHunt = h
			bestMatchup = m
			found = true
		}
	}

	return bestHunt, bestMatchup, found
}

// SelectNextPokemonForBalancedTeam identifica qual membro da equipe precisa de treinamento para manter todos no mesmo nível
func SelectNextPokemonForBalancedTeam(pokemons map[int64]PokeItemPokemon, trainerLevel int, activeID int64) (PokeItemPokemon, bool) {
	var usable []PokeItemPokemon
	for _, p := range pokemons {
		if p.Starter || (trainerLevel > 0 && p.Level <= trainerLevel+5) || trainerLevel == 0 {
			usable = append(usable, p)
		}
	}
	if len(usable) == 0 {
		return PokeItemPokemon{}, false
	}

	// Filtra os pokémons da equipe (ou starter)
	var teamMembers []PokeItemPokemon
	for _, p := range usable {
		if p.Slot != nil || p.ID == activeID || p.Starter {
			teamMembers = append(teamMembers, p)
		}
	}
	if len(teamMembers) == 0 {
		teamMembers = usable
	}

	var currentActive *PokeItemPokemon
	for _, p := range teamMembers {
		if p.ID == activeID {
			cp := p
			currentActive = &cp
			break
		}
	}

	// Ordena membros do menor nível para o maior
	sort.Slice(teamMembers, func(i, j int) bool {
		return teamMembers[i].Level < teamMembers[j].Level
	})

	lowestMember := teamMembers[0]

	// Se o ativo já encostou no nível do treinador e outro membro está mais de 1 nível atrás:
	// rotaciona para o menor subir!
	if currentActive != nil && currentActive.Level >= trainerLevel && lowestMember.Level < currentActive.Level {
		return lowestMember, true
	}

	// Se o ativo está quase no teto (+4), força a troca para não bater na trava de desobediência
	if currentActive != nil && currentActive.Level >= trainerLevel+4 && lowestMember.ID != currentActive.ID {
		return lowestMember, true
	}

	if currentActive != nil {
		return *currentActive, true
	}

	return lowestMember, true
}
