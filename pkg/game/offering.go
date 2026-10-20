package game

import (
	"time"
)

// OfferingManager gerencia os ciclos de sacrifício de Pokémons excedentes na Roleta da Oferenda
type OfferingManager struct {
	LastOffering time.Time
}

// NewOfferingManager cria uma nova instância
func NewOfferingManager() *OfferingManager {
	return &OfferingManager{}
}

// ShouldRunOffering verifica se é hora de enviar oferenda.tudo (a cada 7 minutos e se tiver pokémons no depot)
func (om *OfferingManager) ShouldRunOffering(totalPokemons int) bool {
	if totalPokemons < 10 {
		return false
	}
	if time.Since(om.LastOffering) > 7*time.Minute {
		om.LastOffering = time.Now()
		return true
	}
	return false
}

// BuildOfferingMessage monta o pacote para girar a roleta da oferenda
func BuildOfferingMessage() WSMessage {
	return WSMessage{
		T: "oferenda.tudo",
	}
}

// FormatOfferingLog gera log informativo para a interface
func FormatOfferingLog() string {
	return "🎰 [Oferenda de Lixo] Sacrificou duplicatas fracas do depot para girar a Roleta por Pedras de Evolução e Mega Stones!"
}
