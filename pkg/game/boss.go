package game

import (
	"fmt"
	"sort"
)

// BossManager cuida da automação do Auto-Boss Rush utilizando os tokens acumulados
type BossManager struct {
	PlayableBosses []PokeBossEntry
}

// NewBossManager inicializa o gerenciador de bosses
func NewBossManager() *BossManager {
	return &BossManager{
		PlayableBosses: DefaultPlayableBosses(),
	}
}

// SetPlayableBosses atualiza os chefes disponíveis a partir do catálogo do welcome
func (bm *BossManager) SetPlayableBosses(bosses []PokeBossEntry) {
	if len(bosses) > 0 {
		bm.PlayableBosses = bosses
	}
}

// DefaultPlayableBosses catálogo de fallback com os chefes clássicos de arena
func DefaultPlayableBosses() []PokeBossEntry {
	return []PokeBossEntry{
		{Key: "brock", Name: "Brock", Level: 15, MinNivelTreinador: 10, Type1: TypeRock, Type2: TypeGround},
		{Key: "misty", Name: "Misty", Level: 25, MinNivelTreinador: 20, Type1: TypeWater},
		{Key: "surge", Name: "Lt. Surge", Level: 35, MinNivelTreinador: 30, Type1: TypeElectric},
		{Key: "erika", Name: "Erika", Level: 45, MinNivelTreinador: 40, Type1: TypeGrass, Type2: TypePoison},
		{Key: "koga", Name: "Koga", Level: 55, MinNivelTreinador: 50, Type1: TypePoison},
		{Key: "sabrina", Name: "Sabrina", Level: 65, MinNivelTreinador: 60, Type1: TypePsychic},
		{Key: "blaine", Name: "Blaine", Level: 75, MinNivelTreinador: 70, Type1: TypeFire},
		{Key: "giovanni", Name: "Giovanni", Level: 85, MinNivelTreinador: 80, Type1: TypeGround},
		{Key: "lorelei", Name: "Lorelei", Level: 90, MinNivelTreinador: 85, Type1: TypeIce, Type2: TypeWater},
		{Key: "bruno", Name: "Bruno", Level: 92, MinNivelTreinador: 88, Type1: TypeFighting, Type2: TypeRock},
		{Key: "agatha", Name: "Agatha", Level: 94, MinNivelTreinador: 90, Type1: TypeGhost, Type2: TypePoison},
		{Key: "lance", Name: "Lance", Level: 96, MinNivelTreinador: 92, Type1: TypeDragon, Type2: TypeFlying},
		{Key: "gary", Name: "Champion Gary", Level: 100, MinNivelTreinador: 95, Type1: TypeNormal},
	}
}

// PickBestBoss seleciona o chefe de nível mais alto que o treinador já pode enfrentar
func (bm *BossManager) PickBestBoss(trainerLevel int, vitorias map[string]int) (PokeBossEntry, bool) {
	if len(bm.PlayableBosses) == 0 {
		return PokeBossEntry{}, false
	}

	var eligible []PokeBossEntry
	for _, b := range bm.PlayableBosses {
		minLevel := b.MinNivelTreinador
		if minLevel == 0 {
			minLevel = b.Level
		}
		if trainerLevel >= minLevel || minLevel <= 10 {
			eligible = append(eligible, b)
		}
	}

	if len(eligible) == 0 {
		return bm.PlayableBosses[0], true
	}

	// Ordena decrescente por nível do chefe para maximizar drops e XP
	sort.Slice(eligible, func(i, j int) bool {
		lvlI := eligible[i].MinNivelTreinador
		if lvlI == 0 {
			lvlI = eligible[i].Level
		}
		lvlJ := eligible[j].MinNivelTreinador
		if lvlJ == 0 {
			lvlJ = eligible[j].Level
		}
		if lvlI != lvlJ {
			return lvlI > lvlJ
		}
		// Desempate por menos vitórias (para completar o Bestiário de Bosses)
		return vitorias[eligible[i].Key] < vitorias[eligible[j].Key]
	})

	return eligible[0], true
}

// BuildBossAutoMessage monta o pacote de ativação do Auto-Boss Rush
func BuildBossAutoMessage(bossKey string, ativo bool) WSMessage {
	return WSMessage{
		T:     "boss.auto",
		Ativo: ativo,
		Key:   bossKey,
	}
}

// FormatBossLog gera log explicativo para a interface
func FormatBossLog(boss PokeBossEntry, tokens int, ativo bool) string {
	name := boss.Nome
	if name == "" {
		name = boss.Name
	}
	if ativo {
		return fmt.Sprintf("👑 [Auto-Boss Rush] Ativou ciclo contínuo contra '%s' (Nv. %d) gastando %d tokens de chefe!",
			name, boss.Level, tokens)
	}
	return fmt.Sprintf("⏸️ [Auto-Boss Rush] Pausou loop do chefe '%s' (Tokens esgotados).", name)
}
