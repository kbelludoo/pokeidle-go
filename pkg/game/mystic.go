package game

import (
	"fmt"
)

const (
	ItemMysticTicket = 70014
)

// BestBallsOrder lista de Pokébolas da melhor para a mais básica para captura de lendários
var BestBallsOrder = []int{5, 4, 3, 2, 1}

// CheckMysticEntry avalia se o jogador possui Mystic Tickets para desafiar a Masmorra Mística
func CheckMysticEntry(items map[string]int, mistico *PokeMisticoState) (WSMessage, string, bool) {
	// Se já estiver dentro da arena ou arremessando, não entra de novo
	if mistico != nil && mistico.SpeciesID > 0 {
		return WSMessage{}, "", false
	}

	temTicket := false
	if items != nil {
		if items[fmt.Sprintf("%d", ItemMysticTicket)] > 0 || items["mysticTicket"] > 0 {
			temTicket = true
		}
	}

	if temTicket {
		return WSMessage{
			T: "mistico.entrar",
		}, "🌌 [Masmorra Mística] Entrou na arena mística para desafiar um Pokémon Lendário com Mystic Ticket!", true
	}

	return WSMessage{}, "", false
}

// PickBestBallForMystic escolhe a melhor bola disponível no inventário
func PickBestBallForMystic(balls map[string]int) int {
	for _, bID := range BestBallsOrder {
		if balls != nil && balls[fmt.Sprintf("%d", bID)] > 0 {
			return bID
		}
	}
	return 1 // Fallback Pokébola comum
}

// BuildMysticBallMessage monta o arremesso de bola no lendário caído
func BuildMysticBallMessage(ballID int) WSMessage {
	return WSMessage{
		T:      "mistico.bola",
		BallID: ballID,
	}
}

// FormatMysticLog gera log de captura ou derrota
func FormatMysticLog(nome string, ballID int) string {
	return fmt.Sprintf("🎯 [Masmorra Mística] Arremessou bola #%d para capturar %s Lendário!", ballID, nome)
}
