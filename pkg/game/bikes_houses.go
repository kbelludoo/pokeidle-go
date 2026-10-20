package game

import (
	"fmt"
)

const (
	ItemFragmentoBicicleta = 70013
	ItemFragmentoCasaChave = 70011
	ItemXpShareHeld        = 70050
)

// RarityOrder ordem de valor de raridades
var RarityOrder = map[string]int{
	"comum":    1,
	"incomum":  2,
	"rara":     3,
	"mitica":   4,
	"lendaria": 5,
}

// CheckBicycleActions avalia se é possível fabricar, fundir (craftar) ou equipar uma melhor bicicleta
func CheckBicycleActions(bici *PokeBicicletaState, items map[string]int) (WSMessage, string, bool) {
	// 1. Fabricar bicicleta se tiver 10+ fragmentos
	if items != nil && items[fmt.Sprintf("%d", ItemFragmentoBicicleta)] >= 10 {
		return WSMessage{T: "bicicleta.fabricar"}, "🚲 [Bicicleta] Fabricou nova bicicleta a partir de 10 fragmentos!", true
	}

	if bici == nil || len(bici.Lista) == 0 {
		return WSMessage{}, "", false
	}

	// 2. Fundir 5 bicicletas de mesma raridade em uma superior (craftar)
	porRaridade := make(map[string][]any)
	for _, b := range bici.Lista {
		porRaridade[b.Raridade] = append(porRaridade[b.Raridade], b.ID)
	}

	for rar, ids := range porRaridade {
		if rar != "lendaria" && len(ids) >= 5 {
			var cinco []any
			for i := 0; i < 5; i++ {
				cinco = append(cinco, ids[i])
			}
			return WSMessage{
				T:            "bicicleta.craftar",
				BicicletaIDs: cinco,
			}, fmt.Sprintf("⚙️ [Bicicleta] Fundindo 5 bicicletas de raridade '%s' em uma de tier superior!", rar), true
		}
	}

	// 3. Equipar a melhor bicicleta disponível se ainda não estiver equipada
	var melhorID any
	melhorScore := -1
	melhorRar := ""

	for _, b := range bici.Lista {
		score := RarityOrder[b.Raridade]
		if score > melhorScore {
			melhorScore = score
			melhorID = b.ID
			melhorRar = b.Raridade
		}
	}

	if melhorID != nil && melhorID != bici.Equipada {
		return WSMessage{
			T:           "bicicleta.equipar",
			BicicletaID: melhorID,
		}, fmt.Sprintf("🚴 [Bicicleta] Equipou bicicleta '%s' para acelerar a caçada e abates/minuto!", melhorRar), true
	}

	return WSMessage{}, "", false
}

// CheckHouseXpShareActions preenche postos vagos de XP Share na casa com Pokémons da reserva
func CheckHouseXpShareActions(casa *PokeCasaState, xpShare *PokeXpShareState, pokemons map[int64]PokeItemPokemon, activeID int64) (WSMessage, string, bool) {
	if xpShare == nil || len(xpShare.Slots) == 0 {
		return WSMessage{}, "", false
	}

	casaID := xpShare.CasaID
	if casaID == nil && casa != nil {
		casaID = casa.Dentro
	}
	if casaID == nil {
		return WSMessage{}, "", false
	}

	// Identifica quais pokémons já estão alocados em treino
	jaEmTreino := make(map[int64]bool)
	for _, s := range xpShare.Slots {
		if s.ID > 0 {
			jaEmTreino[s.ID] = true
		}
	}

	// Procura primeiro slot vago
	for idx, s := range xpShare.Slots {
		if s.ID == 0 {
			// Encontra um Pokémon da reserva para treinar
			for _, p := range pokemons {
				if p.ID != activeID && p.Slot == nil && !jaEmTreino[p.ID] {
					pName := p.Nome
					if pName == "" {
						pName = p.Name
					}
					return WSMessage{
						T:         "xpshare.escolher",
						CasaID:    casaID,
						Slot:      &idx,
						PokemonID: p.ID,
					}, fmt.Sprintf("🏠 [XP Share de Casa] Alocou %s (Nv. %d) no posto de treino da casa para receber XP passivo das lutas!", pName, p.Level), true
				}
			}
		}
	}

	return WSMessage{}, "", false
}
