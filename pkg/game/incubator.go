package game

import (
	"fmt"
)

// EggPriority ordem de preferência para chocar ovos
var EggPriority = []string{"shiny", "p4", "p3", "p2"}

// CheckIncubatorActions verifica se há vagas na chocadeira e ovos no inventário para colocar
func CheckIncubatorActions(choc *PokeChocadeiraState, items map[string]int) (WSMessage, string, bool) {
	if choc == nil || choc.Total <= 0 {
		return WSMessage{}, "", false
	}

	// 1. Fabricar Ovo Shiny se tiver fragmentos suficientes
	if items != nil && items["70012"] >= 10 { // Fragmento de Shiny / Ovo Shiny
		return WSMessage{T: "ovoShiny.fabricar"}, "✨ [Chocadeira] Fabricou um Ovo Shiny garantido na bancada!", true
	}

	// Identifica slots ocupados
	slotsOcupados := make(map[int]bool)
	for _, o := range choc.Ovos {
		slotsOcupados[o.Slot] = true
	}

	// Procura primeiro slot livre
	slotLivre := -1
	for s := 0; s < choc.Total; s++ {
		if !slotsOcupados[s] {
			slotLivre = s
			break
		}
	}

	if slotLivre == -1 {
		return WSMessage{}, "", false
	}

	// Procura ovos no inventário
	for _, eggType := range EggPriority {
		// Checa se o item está presente
		temOvo := false
		if items != nil {
			for k, v := range items {
				if v > 0 && (k == eggType || k == fmt.Sprintf("ovo-%s", eggType)) {
					temOvo = true
					break
				}
			}
		}

		if temOvo {
			return WSMessage{
				T:    "chocadeira.colocar",
				Slot: &slotLivre,
				Ovo:  eggType,
			}, fmt.Sprintf("🥚 [Chocadeira] Colocou ovo '%s' no slot livre #%d para incubar!", eggType, slotLivre+1), true
		}
	}

	return WSMessage{}, "", false
}
