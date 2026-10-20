package game

import (
	"testing"
	"time"
)

func TestEvolution(t *testing.T) {
	// 1. Bulbasaur nível 12 não deve evoluir
	pkLow := PokeItemPokemon{ID: 101, PokeID: 1, Level: 12, Nome: "Bulbasaur"}
	if _, ok := CheckPokemonEvolution(pkLow); ok {
		t.Errorf("Bulbasaur Nv. 12 não deveria evoluir")
	}

	// 2. Bulbasaur nível 45 deve evoluir para Ivysaur (ID 2)
	pkHigh := PokeItemPokemon{ID: 102, PokeID: 1, Level: 45, Nome: "Bulbasaur"}
	evo, ok := CheckPokemonEvolution(pkHigh)
	if !ok {
		t.Fatalf("Bulbasaur Nv. 45 deveria estar elegível para evoluir")
	}
	if evo.TargetID != 2 || evo.TargetName != "Ivysaur" {
		t.Errorf("Esperava Ivysaur (ID 2), obteve %s (ID %d)", evo.TargetName, evo.TargetID)
	}

	// 3. Ivysaur nível 45 deve evoluir para Venusaur (ID 3)
	ivy := PokeItemPokemon{ID: 103, PokeID: 2, Level: 45, Nome: "Ivysaur"}
	evoIvy, ok := CheckPokemonEvolution(ivy)
	if !ok {
		t.Fatalf("Ivysaur Nv. 45 deveria estar elegível para evoluir para Venusaur")
	}
	if evoIvy.TargetID != 3 || evoIvy.TargetName != "Venusaur" {
		t.Errorf("Esperava Venusaur (ID 3), obteve %s (ID %d)", evoIvy.TargetName, evoIvy.TargetID)
	}

	// 4. Variante Outland não deve evoluir
	outland := PokeItemPokemon{ID: 104, PokeID: 2005, Level: 50, Nome: "Ancient Charizard"}
	if _, ok := CheckPokemonEvolution(outland); ok {
		t.Errorf("Variante Outland não deveria evoluir")
	}

	// 5. Verificação em lote
	pokemons := map[int64]PokeItemPokemon{
		101: pkLow,
		102: pkHigh,
	}
	list := FindEvolutionsToPerform(pokemons)
	if len(list) != 1 || list[0].Pokemon.ID != 102 {
		t.Errorf("Esperava 1 evolução pendente para o ID 102, obteve %d", len(list))
	}

	msg := BuildEvolutionMessage(pkHigh.ID, evo.TargetID)
	if msg.T != "pokemon.evoluir" || msg.PokemonID != 102 || msg.AlvoID != 2 {
		t.Errorf("Mensagem de evolução malformada: %+v", msg)
	}
}

func TestBossRush(t *testing.T) {
	bm := NewBossManager()
	vitorias := make(map[string]int)

	// Treinador Nv. 45 deve conseguir enfrentar Erika (min level 40)
	boss, found := bm.PickBestBoss(45, vitorias)
	if !found {
		t.Fatalf("Deveria encontrar boss para treinador Nv. 45")
	}
	if boss.Key != "erika" {
		t.Errorf("Esperava 'erika', obteve '%s'", boss.Key)
	}

	// Treinador Nv. 95 deve pegar Gary (Champion)
	bossGary, found := bm.PickBestBoss(95, vitorias)
	if !found || bossGary.Key != "gary" {
		t.Errorf("Esperava 'gary', obteve '%s'", bossGary.Key)
	}

	msg := BuildBossAutoMessage(boss.Key, true)
	if msg.T != "boss.auto" || msg.Key != "erika" || !msg.Ativo {
		t.Errorf("Mensagem boss.auto malformada: %+v", msg)
	}
}

func TestGymManager(t *testing.T) {
	gm := NewGymManager()
	info := &PokeGinasioInfo{
		NivelMin:      10,
		PodeRegistrar: true,
		Ginasios: []PokeGinasioEntry{
			{Tipo: TypeNormal, SouLider: true},  // Já lidera: deve pular
			{Tipo: TypeFire, SouLider: false},  // Não lidera: deve desafiar
		},
	}
	gm.UpdateFromInfo(info)

	tipo, ok := gm.PickNextGymToChallenge(25)
	if !ok {
		t.Fatalf("Deveria encontrar ginásio para desafiar")
	}
	// O primeiro que não lidera deve ser FIRE (ou próximo na lista)
	if tipo == TypeNormal {
		t.Errorf("Não deveria desafiar NORMAL pois já lidera")
	}

	// Cooldown de 45 segundos deve bloquear desafio consecutivo imediato
	_, okImediato := gm.PickNextGymToChallenge(25)
	if okImediato {
		t.Errorf("Deveria bloquear desafio de ginásio consecutivo antes do cooldown")
	}
}

func TestBicycles(t *testing.T) {
	// 1. Fragmentos suficientes para fabricar
	items := map[string]int{"70013": 12}
	msg, _, ok := CheckBicycleActions(nil, items)
	if !ok || msg.T != "bicicleta.fabricar" {
		t.Errorf("Deveria fabricar bicicleta com 12 fragmentos")
	}

	// 2. Fundir 5 bicicletas de mesma raridade
	biciState := &PokeBicicletaState{
		Equipada:    1,
		EquipadaRar: "comum",
		Lista: []PokeBicicletaItem{
			{ID: 1, Raridade: "comum"},
			{ID: 2, Raridade: "comum"},
			{ID: 3, Raridade: "comum"},
			{ID: 4, Raridade: "comum"},
			{ID: 5, Raridade: "comum"},
		},
	}
	msgCraft, _, okCraft := CheckBicycleActions(biciState, map[string]int{})
	if !okCraft || msgCraft.T != "bicicleta.craftar" || len(msgCraft.BicicletaIDs) != 5 {
		t.Errorf("Deveria fundir 5 bicicletas comuns: %+v", msgCraft)
	}

	// 3. Equipar bicicleta superior
	biciStateUpgrade := &PokeBicicletaState{
		Equipada:    1,
		EquipadaRar: "comum",
		Lista: []PokeBicicletaItem{
			{ID: 1, Raridade: "comum"},
			{ID: 10, Raridade: "rara"},
		},
	}
	msgEquip, _, okEquip := CheckBicycleActions(biciStateUpgrade, map[string]int{})
	if !okEquip || msgEquip.T != "bicicleta.equipar" || msgEquip.BicicletaID != 10 {
		t.Errorf("Deveria equipar bicicleta rara (ID 10): %+v", msgEquip)
	}
}

func TestIncubator(t *testing.T) {
	choc := &PokeChocadeiraState{
		Total: 2,
		Ovos: []PokeEggSlot{
			{Slot: 0, Ovo: "p2"},
		},
	}
	items := map[string]int{"shiny": 1, "p3": 2}

	msg, _, ok := CheckIncubatorActions(choc, items)
	if !ok {
		t.Fatalf("Deveria encontrar vaga e ovo para chocar")
	}
	if msg.T != "chocadeira.colocar" || *msg.Slot != 1 || msg.Ovo != "shiny" {
		t.Errorf("Deveria colocar ovo shiny no slot 1: %+v", msg)
	}

	// Chocadeira cheia não deve colocar
	chocCheia := &PokeChocadeiraState{
		Total: 1,
		Ovos: []PokeEggSlot{
			{Slot: 0, Ovo: "p2"},
		},
	}
	if _, _, okCheia := CheckIncubatorActions(chocCheia, items); okCheia {
		t.Errorf("Chocadeira cheia não deveria aceitar mais ovos")
	}
}

func TestMystic(t *testing.T) {
	// Com Mystic Ticket
	items := map[string]int{"70014": 1}
	msg, _, ok := CheckMysticEntry(items, nil)
	if !ok || msg.T != "mistico.entrar" {
		t.Errorf("Deveria entrar na arena mística com ticket")
	}

	// Seleção de melhor bola
	balls := map[string]int{"1": 10, "3": 5, "4": 2} // Tem Great Ball e Ultra Ball
	bestBall := PickBestBallForMystic(balls)
	if bestBall != 4 {
		t.Errorf("Esperava Ultra Ball (ID 4), obteve ID %d", bestBall)
	}
}

func TestOffering(t *testing.T) {
	om := NewOfferingManager()

	// Menos de 10 pokemons não deve rodar
	if om.ShouldRunOffering(5) {
		t.Errorf("Não deveria rodar oferenda com menos de 10 pokémons")
	}

	// 15 pokemons e primeira execução deve rodar
	if !om.ShouldRunOffering(15) {
		t.Errorf("Deveria rodar oferenda com 15 pokémons")
	}

	// Imediatamente após não deve rodar pelo cooldown
	if om.ShouldRunOffering(15) {
		t.Errorf("Não deveria rodar oferenda imediatamente após disparo")
	}

	// Simulando passagem do tempo
	om.LastOffering = time.Now().Add(-10 * time.Minute)
	if !om.ShouldRunOffering(15) {
		t.Errorf("Deveria rodar oferenda após expirar o cooldown de 7 min")
	}
}
