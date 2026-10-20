package game

import (
	"fmt"
)

// EvolutionPath define o requisito de nível e espécie destino para evolução
type EvolutionPath struct {
	SourceID   int    // Species / PokeID de origem
	TargetID   int    // Species / PokeID de destino
	TargetName string // Nome da evolução
	MinLevel   int    // Nível mínimo exigido
	StoneID    int    // Item ID da pedra (0 se evolução apenas por nível)
}

// EvolutionTable mapeia as principais cadeias evolutivas (Gerações 1 a 9)
var EvolutionTable = map[int]EvolutionPath{
	// Gen 1 Iniciais & Comuns
	1:  {SourceID: 1, TargetID: 2, TargetName: "Ivysaur", MinLevel: 16},
	2:  {SourceID: 2, TargetID: 3, TargetName: "Venusaur", MinLevel: 32},
	4:  {SourceID: 4, TargetID: 5, TargetName: "Charmeleon", MinLevel: 16},
	5:  {SourceID: 5, TargetID: 6, TargetName: "Charizard", MinLevel: 36},
	7:  {SourceID: 7, TargetID: 8, TargetName: "Wartortle", MinLevel: 16},
	8:  {SourceID: 8, TargetID: 9, TargetName: "Blastoise", MinLevel: 36},
	10: {SourceID: 10, TargetID: 11, TargetName: "Metapod", MinLevel: 7},
	11: {SourceID: 11, TargetID: 12, TargetName: "Butterfree", MinLevel: 10},
	13: {SourceID: 13, TargetID: 14, TargetName: "Kakuna", MinLevel: 7},
	14: {SourceID: 14, TargetID: 15, TargetName: "Beedrill", MinLevel: 10},
	16: {SourceID: 16, TargetID: 17, TargetName: "Pidgeotto", MinLevel: 18},
	17: {SourceID: 17, TargetID: 18, TargetName: "Pidgeot", MinLevel: 36},
	19: {SourceID: 19, TargetID: 20, TargetName: "Raticate", MinLevel: 20},
	21: {SourceID: 21, TargetID: 22, TargetName: "Fearow", MinLevel: 20},
	23: {SourceID: 23, TargetID: 24, TargetName: "Arbok", MinLevel: 22},
	25: {SourceID: 25, TargetID: 26, TargetName: "Raichu", MinLevel: 20},
	27: {SourceID: 27, TargetID: 28, TargetName: "Sandslash", MinLevel: 22},
	29: {SourceID: 29, TargetID: 30, TargetName: "Nidorina", MinLevel: 16},
	30: {SourceID: 30, TargetID: 31, TargetName: "Nidoqueen", MinLevel: 36},
	32: {SourceID: 32, TargetID: 33, TargetName: "Nidorino", MinLevel: 16},
	33: {SourceID: 33, TargetID: 34, TargetName: "Nidoking", MinLevel: 36},
	41: {SourceID: 41, TargetID: 42, TargetName: "Golbat", MinLevel: 22},
	42: {SourceID: 42, TargetID: 169, TargetName: "Crobat", MinLevel: 36},
	43: {SourceID: 43, TargetID: 44, TargetName: "Gloom", MinLevel: 21},
	44: {SourceID: 44, TargetID: 45, TargetName: "Vileplume", MinLevel: 36},
	46: {SourceID: 46, TargetID: 47, TargetName: "Parasect", MinLevel: 24},
	48: {SourceID: 48, TargetID: 49, TargetName: "Venomoth", MinLevel: 31},
	50: {SourceID: 50, TargetID: 51, TargetName: "Dugtrio", MinLevel: 26},
	52: {SourceID: 52, TargetID: 53, TargetName: "Persian", MinLevel: 28},
	54: {SourceID: 54, TargetID: 55, TargetName: "Golduck", MinLevel: 33},
	56: {SourceID: 56, TargetID: 57, TargetName: "Primeape", MinLevel: 28},
	60: {SourceID: 60, TargetID: 61, TargetName: "Poliwhirl", MinLevel: 25},
	61: {SourceID: 61, TargetID: 62, TargetName: "Poliwrath", MinLevel: 36},
	63: {SourceID: 63, TargetID: 64, TargetName: "Kadabra", MinLevel: 16},
	64: {SourceID: 64, TargetID: 65, TargetName: "Alakazam", MinLevel: 36},
	66: {SourceID: 66, TargetID: 67, TargetName: "Machoke", MinLevel: 28},
	67: {SourceID: 67, TargetID: 68, TargetName: "Machamp", MinLevel: 36},
	69: {SourceID: 69, TargetID: 70, TargetName: "Weepinbell", MinLevel: 21},
	70: {SourceID: 70, TargetID: 71, TargetName: "Victreebel", MinLevel: 36},
	72: {SourceID: 72, TargetID: 73, TargetName: "Tentacruel", MinLevel: 30},
	74: {SourceID: 74, TargetID: 75, TargetName: "Graveler", MinLevel: 25},
	75: {SourceID: 75, TargetID: 76, TargetName: "Golem", MinLevel: 36},
	77: {SourceID: 77, TargetID: 78, TargetName: "Rapidash", MinLevel: 40},
	79: {SourceID: 79, TargetID: 80, TargetName: "Slowbro", MinLevel: 37},
	81: {SourceID: 81, TargetID: 82, TargetName: "Magneton", MinLevel: 30},
	82: {SourceID: 82, TargetID: 462, TargetName: "Magnezone", MinLevel: 45},
	84: {SourceID: 84, TargetID: 85, TargetName: "Dodrio", MinLevel: 31},
	86: {SourceID: 86, TargetID: 87, TargetName: "Dewgong", MinLevel: 34},
	88: {SourceID: 88, TargetID: 89, TargetName: "Muk", MinLevel: 38},
	92: {SourceID: 92, TargetID: 93, TargetName: "Haunter", MinLevel: 25},
	93: {SourceID: 93, TargetID: 94, TargetName: "Gengar", MinLevel: 36},
	96: {SourceID: 96, TargetID: 97, TargetName: "Hypno", MinLevel: 26},
	98: {SourceID: 98, TargetID: 99, TargetName: "Kingler", MinLevel: 28},
	100: {SourceID: 100, TargetID: 101, TargetName: "Electrode", MinLevel: 30},
	104: {SourceID: 104, TargetID: 105, TargetName: "Marowak", MinLevel: 28},
	109: {SourceID: 109, TargetID: 110, TargetName: "Weezing", MinLevel: 35},
	111: {SourceID: 111, TargetID: 112, TargetName: "Rhydon", MinLevel: 42},
	112: {SourceID: 112, TargetID: 464, TargetName: "Rhyperior", MinLevel: 55},
	116: {SourceID: 116, TargetID: 117, TargetName: "Seadra", MinLevel: 32},
	117: {SourceID: 117, TargetID: 230, TargetName: "Kingdra", MinLevel: 45},
	118: {SourceID: 118, TargetID: 119, TargetName: "Seaking", MinLevel: 33},
	129: {SourceID: 129, TargetID: 130, TargetName: "Gyarados", MinLevel: 20},
	133: {SourceID: 133, TargetID: 134, TargetName: "Vaporeon", MinLevel: 25},
	138: {SourceID: 138, TargetID: 139, TargetName: "Omastar", MinLevel: 40},
	140: {SourceID: 140, TargetID: 141, TargetName: "Kabutops", MinLevel: 40},
	147: {SourceID: 147, TargetID: 148, TargetName: "Dragonair", MinLevel: 30},
	148: {SourceID: 148, TargetID: 149, TargetName: "Dragonite", MinLevel: 55},

	// Gen 2 Iniciais & Comuns
	152: {SourceID: 152, TargetID: 153, TargetName: "Bayleef", MinLevel: 16},
	153: {SourceID: 153, TargetID: 154, TargetName: "Meganium", MinLevel: 32},
	155: {SourceID: 155, TargetID: 156, TargetName: "Quilava", MinLevel: 14},
	156: {SourceID: 156, TargetID: 157, TargetName: "Typhlosion", MinLevel: 36},
	158: {SourceID: 158, TargetID: 159, TargetName: "Croconaw", MinLevel: 18},
	159: {SourceID: 159, TargetID: 160, TargetName: "Feraligatr", MinLevel: 30},
	161: {SourceID: 161, TargetID: 162, TargetName: "Furret", MinLevel: 15},
	163: {SourceID: 163, TargetID: 164, TargetName: "Noctowl", MinLevel: 20},
	165: {SourceID: 165, TargetID: 166, TargetName: "Ledian", MinLevel: 18},
	167: {SourceID: 167, TargetID: 168, TargetName: "Ariados", MinLevel: 22},
	170: {SourceID: 170, TargetID: 171, TargetName: "Lanturn", MinLevel: 27},
	179: {SourceID: 179, TargetID: 180, TargetName: "Flaaffy", MinLevel: 15},
	180: {SourceID: 180, TargetID: 181, TargetName: "Ampharos", MinLevel: 30},
	183: {SourceID: 183, TargetID: 184, TargetName: "Azumarill", MinLevel: 18},
	194: {SourceID: 194, TargetID: 195, TargetName: "Quagsire", MinLevel: 20},
	204: {SourceID: 204, TargetID: 205, TargetName: "Forretress", MinLevel: 31},
	216: {SourceID: 216, TargetID: 217, TargetName: "Ursaring", MinLevel: 30},
	220: {SourceID: 220, TargetID: 221, TargetName: "Piloswine", MinLevel: 33},
	228: {SourceID: 228, TargetID: 229, TargetName: "Houndoom", MinLevel: 24},
	231: {SourceID: 231, TargetID: 232, TargetName: "Donphan", MinLevel: 25},
	246: {SourceID: 246, TargetID: 247, TargetName: "Pupitar", MinLevel: 30},
	247: {SourceID: 247, TargetID: 248, TargetName: "Tyranitar", MinLevel: 55},

	// Gen 3 Iniciais & Comuns
	252: {SourceID: 252, TargetID: 253, TargetName: "Grovyle", MinLevel: 16},
	253: {SourceID: 253, TargetID: 254, TargetName: "Sceptile", MinLevel: 36},
	255: {SourceID: 255, TargetID: 256, TargetName: "Combusken", MinLevel: 16},
	256: {SourceID: 256, TargetID: 257, TargetName: "Blaziken", MinLevel: 36},
	258: {SourceID: 258, TargetID: 259, TargetName: "Marshtomp", MinLevel: 16},
	259: {SourceID: 259, TargetID: 260, TargetName: "Swampert", MinLevel: 36},
	261: {SourceID: 261, TargetID: 262, TargetName: "Mightyena", MinLevel: 18},
	263: {SourceID: 263, TargetID: 264, TargetName: "Linoone", MinLevel: 20},
	276: {SourceID: 276, TargetID: 277, TargetName: "Swellow", MinLevel: 22},
	280: {SourceID: 280, TargetID: 281, TargetName: "Kirlia", MinLevel: 20},
	281: {SourceID: 281, TargetID: 282, TargetName: "Gardevoir", MinLevel: 30},
	287: {SourceID: 287, TargetID: 288, TargetName: "Vigoroth", MinLevel: 18},
	288: {SourceID: 288, TargetID: 289, TargetName: "Slaking", MinLevel: 36},
	304: {SourceID: 304, TargetID: 305, TargetName: "Lairon", MinLevel: 32},
	305: {SourceID: 305, TargetID: 306, TargetName: "Aggron", MinLevel: 42},
	328: {SourceID: 328, TargetID: 329, TargetName: "Vibrava", MinLevel: 35},
	329: {SourceID: 329, TargetID: 330, TargetName: "Flygon", MinLevel: 45},
	371: {SourceID: 371, TargetID: 372, TargetName: "Shelgon", MinLevel: 30},
	372: {SourceID: 372, TargetID: 373, TargetName: "Salamence", MinLevel: 50},
	374: {SourceID: 374, TargetID: 375, TargetName: "Metang", MinLevel: 20},
	375: {SourceID: 375, TargetID: 376, TargetName: "Metagross", MinLevel: 45},

	// Gen 4 Iniciais & Comuns
	387: {SourceID: 387, TargetID: 388, TargetName: "Grotle", MinLevel: 18},
	388: {SourceID: 388, TargetID: 389, TargetName: "Torterra", MinLevel: 32},
	390: {SourceID: 390, TargetID: 391, TargetName: "Monferno", MinLevel: 14},
	391: {SourceID: 391, TargetID: 392, TargetName: "Infernape", MinLevel: 36},
	393: {SourceID: 393, TargetID: 394, TargetName: "Prinplup", MinLevel: 16},
	394: {SourceID: 394, TargetID: 395, TargetName: "Empoleon", MinLevel: 36},
	396: {SourceID: 396, TargetID: 397, TargetName: "Staravia", MinLevel: 14},
	397: {SourceID: 397, TargetID: 398, TargetName: "Staraptor", MinLevel: 34},
	403: {SourceID: 403, TargetID: 404, TargetName: "Luxio", MinLevel: 15},
	404: {SourceID: 404, TargetID: 405, TargetName: "Luxray", MinLevel: 30},
	443: {SourceID: 443, TargetID: 444, TargetName: "Gabite", MinLevel: 24},
	444: {SourceID: 444, TargetID: 445, TargetName: "Garchomp", MinLevel: 48},
	447: {SourceID: 447, TargetID: 448, TargetName: "Lucario", MinLevel: 30},

	// Gen 5 Iniciais
	495: {SourceID: 495, TargetID: 496, TargetName: "Servine", MinLevel: 17},
	496: {SourceID: 496, TargetID: 497, TargetName: "Serperior", MinLevel: 36},
	498: {SourceID: 498, TargetID: 499, TargetName: "Pignite", MinLevel: 17},
	499: {SourceID: 499, TargetID: 500, TargetName: "Emboar", MinLevel: 36},
	501: {SourceID: 501, TargetID: 502, TargetName: "Dewott", MinLevel: 17},
	502: {SourceID: 502, TargetID: 503, TargetName: "Samurott", MinLevel: 36},
	633: {SourceID: 633, TargetID: 634, TargetName: "Zweilous", MinLevel: 50},
	634: {SourceID: 634, TargetID: 635, TargetName: "Hydreigon", MinLevel: 64},

	// Gen 6 Iniciais
	650: {SourceID: 650, TargetID: 651, TargetName: "Quilladin", MinLevel: 16},
	651: {SourceID: 651, TargetID: 652, TargetName: "Chesnaught", MinLevel: 36},
	653: {SourceID: 653, TargetID: 654, TargetName: "Braixen", MinLevel: 16},
	654: {SourceID: 654, TargetID: 655, TargetName: "Delphox", MinLevel: 36},
	656: {SourceID: 656, TargetID: 657, TargetName: "Frogadier", MinLevel: 16},
	657: {SourceID: 657, TargetID: 658, TargetName: "Greninja", MinLevel: 36},
	704: {SourceID: 704, TargetID: 705, TargetName: "Sliggoo", MinLevel: 40},
	705: {SourceID: 705, TargetID: 706, TargetName: "Goodra", MinLevel: 50},

	// Gen 7 Iniciais
	722: {SourceID: 722, TargetID: 723, TargetName: "Dartrix", MinLevel: 17},
	723: {SourceID: 723, TargetID: 724, TargetName: "Decidueye", MinLevel: 34},
	725: {SourceID: 725, TargetID: 726, TargetName: "Torracat", MinLevel: 17},
	726: {SourceID: 726, TargetID: 727, TargetName: "Incineroar", MinLevel: 34},
	728: {SourceID: 728, TargetID: 729, TargetName: "Brionne", MinLevel: 17},
	729: {SourceID: 729, TargetID: 730, TargetName: "Primarina", MinLevel: 34},

	// Gen 8 Iniciais
	810: {SourceID: 810, TargetID: 811, TargetName: "Thwackey", MinLevel: 16},
	811: {SourceID: 811, TargetID: 812, TargetName: "Rillaboom", MinLevel: 35},
	813: {SourceID: 813, TargetID: 814, TargetName: "Raboot", MinLevel: 16},
	814: {SourceID: 814, TargetID: 815, TargetName: "Cinderace", MinLevel: 35},
	816: {SourceID: 816, TargetID: 817, TargetName: "Drizzile", MinLevel: 16},
	817: {SourceID: 817, TargetID: 818, TargetName: "Inteleon", MinLevel: 35},
	885: {SourceID: 885, TargetID: 886, TargetName: "Drakloak", MinLevel: 50},
	886: {SourceID: 886, TargetID: 887, TargetName: "Dragapult", MinLevel: 60},

	// Gen 9 Iniciais
	906: {SourceID: 906, TargetID: 907, TargetName: "Floragato", MinLevel: 16},
	907: {SourceID: 907, TargetID: 908, TargetName: "Meowscarada", MinLevel: 36},
	909: {SourceID: 909, TargetID: 910, TargetName: "Crocalor", MinLevel: 16},
	910: {SourceID: 910, TargetID: 911, TargetName: "Skeledirge", MinLevel: 36},
	912: {SourceID: 912, TargetID: 913, TargetName: "Quaxwell", MinLevel: 16},
	913: {SourceID: 913, TargetID: 914, TargetName: "Quaquaval", MinLevel: 36},
	996: {SourceID: 996, TargetID: 997, TargetName: "Arctibax", MinLevel: 35},
	997: {SourceID: 997, TargetID: 998, TargetName: "Baxcalibur", MinLevel: 54},
}

// CheckPokemonEvolution avalia se um dado Pokémon está elegível para evoluir
func CheckPokemonEvolution(pk PokeItemPokemon) (EvolutionPath, bool) {
	// Espécie base (pokeId ou speciesId)
	spID := pk.PokeID
	if pk.SpeciesID > 0 {
		spID = pk.SpeciesID
	}

	// Não evolui variantes Outland (IDs >= 2000 no catálogo especial)
	if spID >= 2000 && spID < 3000 {
		return EvolutionPath{}, false
	}

	evo, found := EvolutionTable[spID]
	if !found {
		return EvolutionPath{}, false
	}

	if pk.Level >= evo.MinLevel {
		return evo, true
	}

	return EvolutionPath{}, false
}

// FindEvolutionsToPerform varre a lista de Pokémons da conta e retorna os que devem evoluir
func FindEvolutionsToPerform(pokemons map[int64]PokeItemPokemon) []struct {
	Pokemon PokeItemPokemon
	Evo     EvolutionPath
} {
	var list []struct {
		Pokemon PokeItemPokemon
		Evo     EvolutionPath
	}

	for _, pk := range pokemons {
		if evo, ok := CheckPokemonEvolution(pk); ok {
			list = append(list, struct {
				Pokemon PokeItemPokemon
				Evo     EvolutionPath
			}{
				Pokemon: pk,
				Evo:     evo,
			})
		}
	}

	return list
}

// BuildEvolutionMessage monta o pacote WebSocket de evolução
func BuildEvolutionMessage(pokemonID int64, targetSpeciesID int) WSMessage {
	return WSMessage{
		T:         "pokemon.evoluir",
		PokemonID: pokemonID,
		AlvoID:    targetSpeciesID,
	}
}

// FormatEvolutionLog gera mensagem descritiva para a telemetria
func FormatEvolutionLog(pk PokeItemPokemon, evo EvolutionPath) string {
	name := pk.Nome
	if name == "" {
		name = pk.Name
	}
	prefix := ""
	if pk.Shiny {
		prefix = "✨ "
	}
	return fmt.Sprintf("🌟 [Auto-Evolução] %s%s (Nv. %d) atingiu os requisitos e está evoluindo para %s (Nv. mín %d)!",
		prefix, name, pk.Level, evo.TargetName, evo.MinLevel)
}
