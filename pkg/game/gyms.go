package game

import (
	"fmt"
	"time"
)

// GymTypes lista os 18 tipos elementais oficiais que possuem ginásios
var GymTypes = []string{
	TypeNormal, TypeFire, TypeWater, TypeGrass, TypeElectric,
	TypeIce, TypeFighting, TypePoison, TypeGround, TypeFlying,
	TypePsychic, TypeBug, TypeRock, TypeGhost, TypeDragon,
	TypeDark, TypeSteel, TypeFairy,
}

// GymManager cuida da consulta e desafios automáticos aos 18 Ginásios Elementais
type GymManager struct {
	LastQuery       time.Time
	LastChallenge   time.Time
	CurrentIndex    int
	GymsStatus      map[string]PokeGinasioEntry
	PodeRegistrar   bool
	NivelMin        int
}

// NewGymManager cria um novo gerenciador de ginásios
func NewGymManager() *GymManager {
	return &GymManager{
		GymsStatus: make(map[string]PokeGinasioEntry),
	}
}

// UpdateFromInfo atualiza o estado interno a partir da resposta ginasio.info
func (gm *GymManager) UpdateFromInfo(info *PokeGinasioInfo) {
	if info == nil {
		return
	}
	gm.PodeRegistrar = info.PodeRegistrar
	gm.NivelMin = info.NivelMin
	for _, g := range info.Ginasios {
		gm.GymsStatus[g.Tipo] = g
	}
}

// ShouldQueryInfo verifica se já é hora de atualizar os ginásios (a cada 10 minutos)
func (gm *GymManager) ShouldQueryInfo() bool {
	return time.Since(gm.LastQuery) > 10*time.Minute
}

// BuildQueryInfoMessage monta a requisição de informações dos ginásios
func BuildQueryInfoMessage() WSMessage {
	return WSMessage{
		T: "ginasio.info",
	}
}

// PickNextGymToChallenge escolhe o próximo ginásio que o jogador não lidera para desafiar
func (gm *GymManager) PickNextGymToChallenge(trainerLevel int) (string, bool) {
	// Respeita intervalo mínimo entre desafios de ginásio (mínimo 45 segundos para segurança)
	if time.Since(gm.LastChallenge) < 45*time.Second {
		return "", false
	}
	if gm.NivelMin > 0 && trainerLevel < gm.NivelMin {
		return "", false
	}

	for i := 0; i < len(GymTypes); i++ {
		idx := (gm.CurrentIndex + i) % len(GymTypes)
		tipo := GymTypes[idx]
		entry, ok := gm.GymsStatus[tipo]
		if !ok || !entry.SouLider {
			gm.CurrentIndex = (idx + 1) % len(GymTypes)
			gm.LastChallenge = time.Now()
			return tipo, true
		}
	}

	return "", false
}

// BuildGymChallengeMessage monta o pacote de desafio
func BuildGymChallengeMessage(tipo string) WSMessage {
	return WSMessage{
		T:    "ginasio.desafiar",
		Tipo: tipo,
	}
}

// BuildGymSaveTeamMessage monta o pacote para salvar equipe defensiva
func BuildGymSaveTeamMessage(tipo string, pokemonIDs []int64) WSMessage {
	return WSMessage{
		T:          "ginasio.salvar",
		Tipo:       tipo,
		PokemonIDs: pokemonIDs,
	}
}

// FormatGymChallengeLog formata log para a interface
func FormatGymChallengeLog(tipo string) string {
	return fmt.Sprintf("🏛️ [Ginásio Elemental] Desafiando o Líder do Ginásio de %s!", tipo)
}
