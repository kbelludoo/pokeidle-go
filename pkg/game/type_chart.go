package game

import (
	"strings"
)

// Pokemon Types
const (
	TypeNormal   = "normal"
	TypeFire     = "fogo"
	TypeWater    = "agua"
	TypeGrass    = "grama"
	TypeElectric = "eletrico"
	TypeIce      = "gelo"
	TypeFighting = "lutador"
	TypePoison   = "veneno"
	TypeGround   = "terra"
	TypeFlying   = "voador"
	TypePsychic  = "psiquico"
	TypeBug      = "inseto"
	TypeRock     = "pedra"
	TypeGhost    = "fantasma"
	TypeDragon   = "dragao"
	TypeSteel    = "aco"
	TypeDark     = "sombrio"
	TypeFairy    = "fada"
)

// typeChart[attacker][defender] -> multiplier
var typeChart = map[string]map[string]float64{
	TypeNormal: {
		TypeRock:  0.5,
		TypeGhost: 0.0,
		TypeSteel: 0.5,
	},
	TypeFire: {
		TypeFire:     0.5,
		TypeWater:    0.5,
		TypeGrass:    2.0,
		TypeIce:      2.0,
		TypeBug:      2.0,
		TypeRock:     0.5,
		TypeDragon:   0.5,
		TypeSteel:    2.0,
	},
	TypeWater: {
		TypeFire:   2.0,
		TypeWater:  0.5,
		TypeGrass:  0.5,
		TypeGround: 2.0,
		TypeRock:   2.0,
		TypeDragon: 0.5,
	},
	TypeGrass: {
		TypeFire:     0.5,
		TypeWater:    2.0,
		TypeGrass:    0.5,
		TypePoison:   0.5,
		TypeGround:   2.0,
		TypeFlying:   0.5,
		TypeBug:      0.5,
		TypeRock:     2.0,
		TypeDragon:   0.5,
		TypeSteel:    0.5,
	},
	TypeElectric: {
		TypeWater:    2.0,
		TypeGrass:    0.5,
		TypeElectric: 0.5,
		TypeGround:   0.0,
		TypeFlying:   2.0,
		TypeDragon:   0.5,
	},
	TypeIce: {
		TypeFire:   0.5,
		TypeWater:  0.5,
		TypeGrass:  2.0,
		TypeIce:    0.5,
		TypeGround: 2.0,
		TypeFlying: 2.0,
		TypeDragon: 2.0,
		TypeSteel:  0.5,
	},
	TypeFighting: {
		TypeNormal:   2.0,
		TypeIce:      2.0,
		TypePoison:   0.5,
		TypeFlying:   0.5,
		TypePsychic:  0.5,
		TypeBug:      0.5,
		TypeRock:     2.0,
		TypeGhost:    0.0,
		TypeDark:     2.0,
		TypeSteel:    2.0,
		TypeFairy:    0.5,
	},
	TypePoison: {
		TypeGrass:  2.0,
		TypePoison: 0.5,
		TypeGround: 0.5,
		TypeRock:   0.5,
		TypeGhost:  0.5,
		TypeSteel:  0.0,
		TypeFairy:  2.0,
	},
	TypeGround: {
		TypeFire:     2.0,
		TypeGrass:    0.5,
		TypeElectric: 2.0,
		TypePoison:   2.0,
		TypeFlying:   0.0,
		TypeBug:      0.5,
		TypeRock:     2.0,
		TypeSteel:    2.0,
	},
	TypeFlying: {
		TypeGrass:    2.0,
		TypeElectric: 0.5,
		TypeFighting: 2.0,
		TypeBug:      2.0,
		TypeRock:     0.5,
		TypeSteel:    0.5,
	},
	TypePsychic: {
		TypeFighting: 2.0,
		TypePoison:   2.0,
		TypePsychic:  0.5,
		TypeDark:     0.0,
		TypeSteel:    0.5,
	},
	TypeBug: {
		TypeFire:     0.5,
		TypeGrass:    2.0,
		TypeFighting: 0.5,
		TypePoison:   0.5,
		TypeFlying:   0.5,
		TypePsychic:  2.0,
		TypeGhost:    0.5,
		TypeDark:     2.0,
		TypeSteel:    0.5,
		TypeFairy:    0.5,
	},
	TypeRock: {
		TypeFire:     2.0,
		TypeIce:      2.0,
		TypeFighting: 0.5,
		TypeGround:   0.5,
		TypeFlying:   2.0,
		TypeBug:      2.0,
		TypeSteel:    0.5,
	},
	TypeGhost: {
		TypeNormal:  0.0,
		TypePsychic: 2.0,
		TypeGhost:   2.0,
		TypeDark:    0.5,
	},
	TypeDragon: {
		TypeDragon: 2.0,
		TypeSteel:  0.5,
		TypeFairy:  0.0,
	},
	TypeSteel: {
		TypeFire:     0.5,
		TypeWater:    0.5,
		TypeElectric: 0.5,
		TypeIce:      2.0,
		TypeRock:     2.0,
		TypeSteel:    0.5,
		TypeFairy:    2.0,
	},
	TypeDark: {
		TypeFighting: 0.5,
		TypePsychic:  2.0,
		TypeGhost:    2.0,
		TypeDark:     0.5,
		TypeFairy:    0.5,
	},
	TypeFairy: {
		TypeFire:     0.5,
		TypeFighting: 2.0,
		TypePoison:   0.5,
		TypeDragon:   2.0,
		TypeDark:     2.0,
		TypeSteel:    0.5,
	},
}

// Normaliza tipo para string limpa
func normalizeType(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	t = strings.ReplaceAll(t, "á", "a")
	t = strings.ReplaceAll(t, "é", "e")
	t = strings.ReplaceAll(t, "í", "i")
	t = strings.ReplaceAll(t, "ç", "c")
	switch t {
	case "grass":
		return TypeGrass
	case "water":
		return TypeWater
	case "fire":
		return TypeFire
	case "electric":
		return TypeElectric
	case "poison":
		return TypePoison
	case "ground":
		return TypeGround
	case "rock":
		return TypeRock
	case "flying":
		return TypeFlying
	case "bug":
		return TypeBug
	case "psychic":
		return TypePsychic
	case "fighting":
		return TypeFighting
	case "ghost":
		return TypeGhost
	case "ice":
		return TypeIce
	case "dragon":
		return TypeDragon
	case "steel":
		return TypeSteel
	case "fairy":
		return TypeFairy
	case "dark":
		return TypeDark
	default:
		return t
	}
}

// pokedexTypes armazena os tipos primários e secundários conhecidos por pokeId
var pokedexTypes = map[int][]string{
	1:  {TypeGrass, TypePoison},  // Bulbasaur
	2:  {TypeGrass, TypePoison},  // Ivysaur
	3:  {TypeGrass, TypePoison},  // Venusaur
	4:  {TypeFire},               // Charmander
	5:  {TypeFire},               // Charmeleon
	6:  {TypeFire, TypeFlying},   // Charizard
	7:  {TypeWater},              // Squirtle
	8:  {TypeWater},              // Wartortle
	9:  {TypeWater},              // Blastoise
	10: {TypeBug},                // Caterpie
	11: {TypeBug},                // Metapod
	12: {TypeBug, TypeFlying},    // Butterfree
	13: {TypeBug, TypePoison},    // Weedle
	14: {TypeBug, TypePoison},    // Kakuna
	15: {TypeBug, TypePoison},    // Beedrill
	16: {TypeNormal, TypeFlying}, // Pidgey
	17: {TypeNormal, TypeFlying}, // Pidgeotto
	18: {TypeNormal, TypeFlying}, // Pidgeot
	19: {TypeNormal},             // Rattata
	20: {TypeNormal},             // Raticate
	21: {TypeNormal, TypeFlying}, // Spearow
	22: {TypeNormal, TypeFlying}, // Fearow
	23: {TypePoison},             // Ekans
	24: {TypePoison},             // Arbok
	25: {TypeElectric},           // Pikachu
	26: {TypeElectric},           // Raichu
	27: {TypeGround},             // Sandshrew
	28: {TypeGround},             // Sandslash
	29: {TypePoison},             // Nidoran F
	30: {TypePoison},             // Nidorina
	31: {TypePoison, TypeGround}, // Nidoqueen
	32: {TypePoison},             // Nidoran M
	33: {TypePoison},             // Nidorino
	34: {TypePoison, TypeGround}, // Nidoking
	35: {TypeFairy},              // Clefairy
	36: {TypeFairy},              // Clefable
	37: {TypeFire},               // Vulpix
	38: {TypeFire},               // Ninetales
	39: {TypeNormal, TypeFairy},  // Jigglypuff
	40: {TypeNormal, TypeFairy},  // Wigglytuff
	41: {TypePoison, TypeFlying}, // Zubat
	42: {TypePoison, TypeFlying}, // Golbat
	43: {TypeGrass, TypePoison},  // Oddish
	44: {TypeGrass, TypePoison},  // Gloom
	45: {TypeGrass, TypePoison},  // Vileplume
	46: {TypeBug, TypeGrass},     // Paras
	47: {TypeBug, TypeGrass},     // Parasect
	48: {TypeBug, TypePoison},    // Venonat
	49: {TypeBug, TypePoison},    // Venomoth
	50: {TypeGround},             // Diglett
	51: {TypeGround},             // Dugtrio
	52: {TypeNormal},             // Meowth
	53: {TypeNormal},             // Persian
	54: {TypeWater},              // Psyduck
	55: {TypeWater},              // Golduck
	56: {TypeFighting},           // Mankey
	57: {TypeFighting},           // Primeape
	58: {TypeFire},               // Growlithe
	59: {TypeFire},               // Arcanine
	60: {TypeWater},              // Poliwag
	61: {TypeWater},              // Poliwhirl
	62: {TypeWater, TypeFighting},// Poliwrath
	63: {TypePsychic},            // Abra
	64: {TypePsychic},            // Kadabra
	65: {TypePsychic},            // Alakazam
	66: {TypeFighting},           // Machop
	67: {TypeFighting},           // Machoke
	68: {TypeFighting},           // Machamp
	69: {TypeGrass, TypePoison},  // Bellsprout
	70: {TypeGrass, TypePoison},  // Weepinbell
	71: {TypeGrass, TypePoison},  // Victreebel
	72: {TypeWater, TypePoison},  // Tentacool
	73: {TypeWater, TypePoison},  // Tentacruel
	74: {TypeRock, TypeGround},   // Geodude
	75: {TypeRock, TypeGround},   // Graveler
	76: {TypeRock, TypeGround},   // Golem
	77: {TypeFire},               // Ponyta
	78: {TypeFire},               // Rapidash
	79: {TypeWater, TypePsychic}, // Slowpoke
	80: {TypeWater, TypePsychic}, // Slowbro
	81: {TypeElectric, TypeSteel},// Magnemite
	82: {TypeElectric, TypeSteel},// Magneton
	83: {TypeNormal, TypeFlying}, // Farfetch'd
	84: {TypeNormal, TypeFlying}, // Doduo
	85: {TypeNormal, TypeFlying}, // Dodrio
	86: {TypeWater},              // Seel
	87: {TypeWater, TypeIce},     // Dewgong
	88: {TypePoison},             // Grimer
	89: {TypePoison},             // Muk
	90: {TypeWater},              // Shellder
	91: {TypeWater, TypeIce},     // Cloyster
	92: {TypeGhost, TypePoison},  // Gastly
	93: {TypeGhost, TypePoison},  // Haunter
	94: {TypeGhost, TypePoison},  // Gengar
	95: {TypeRock, TypeGround},   // Onix
	96: {TypePsychic},            // Drowzee
	97: {TypePsychic},            // Hypno
	98: {TypeWater},              // Krabby
	99: {TypeWater},              // Kingler
	100: {TypeElectric},          // Voltorb
	101: {TypeElectric},          // Electrode
	102: {TypeGrass, TypePsychic},// Exeggcute
	103: {TypeGrass, TypePsychic},// Exeggutor
	104: {TypeGround},            // Cubone
	105: {TypeGround},            // Marowak
	106: {TypeFighting},          // Hitmonlee
	107: {TypeFighting},          // Hitmonchan
	108: {TypeNormal},            // Lickitung
	109: {TypePoison},            // Koffing
	110: {TypePoison},            // Weezing
	111: {TypeGround, TypeRock},  // Rhyhorn
	112: {TypeGround, TypeRock},  // Rhydon
	113: {TypeNormal},            // Chansey
	114: {TypeGrass},             // Tangela
	115: {TypeNormal},            // Kangaskhan
	116: {TypeWater},             // Horsea
	117: {TypeWater},             // Seadra
	118: {TypeWater},             // Goldeen
	119: {TypeWater},             // Seaking
	120: {TypeWater},             // Staryu
	121: {TypeWater, TypePsychic},// Starmie
	122: {TypePsychic, TypeFairy},// Mr. Mime
	123: {TypeBug, TypeFlying},   // Scyther
	124: {TypeIce, TypePsychic},  // Jynx
	125: {TypeElectric},          // Electabuzz
	126: {TypeFire},              // Magmar
	127: {TypeBug},               // Pinsir
	128: {TypeNormal},            // Tauros
	129: {TypeWater},             // Magikarp
	130: {TypeWater, TypeFlying}, // Gyarados
	131: {TypeWater, TypeIce},    // Lapras
	132: {TypeNormal},            // Ditto
	133: {TypeNormal},            // Eevee
	134: {TypeWater},             // Vaporeon
	135: {TypeElectric},          // Jolteon
	136: {TypeFire},              // Flareon
	137: {TypeNormal},            // Porygon
	138: {TypeRock, TypeWater},   // Omanyte
	139: {TypeRock, TypeWater},   // Omastar
	140: {TypeRock, TypeWater},   // Kabuto
	141: {TypeRock, TypeWater},   // Kabutops
	142: {TypeRock, TypeFlying},  // Aerodactyl
	143: {TypeNormal},            // Snorlax
	144: {TypeIce, TypeFlying},   // Articuno
	145: {TypeElectric, TypeFlying}, // Zapdos
	146: {TypeFire, TypeFlying},  // Moltres
	147: {TypeDragon},            // Dratini
	148: {TypeDragon},            // Dragonair
	149: {TypeDragon, TypeFlying},// Dragonite
	150: {TypePsychic},           // Mewtwo
	151: {TypePsychic},           // Mew
}

// GetPokemonTypes retorna os tipos elementais de um Pokémon dado seu pokeId ou nome
func GetPokemonTypes(pokeID int, name string) []string {
	if types, exists := pokedexTypes[pokeID]; exists && len(types) > 0 {
		return types
	}

	nameLower := strings.ToLower(name)
	for id, types := range pokedexTypes {
		_ = id
		for _, t := range types {
			if strings.Contains(nameLower, t) {
				return []string{t}
			}
		}
	}

	// Heurísticas por nome comuns
	if strings.Contains(nameLower, "bulba") || strings.Contains(nameLower, "ivy") || strings.Contains(nameLower, "venu") {
		return []string{TypeGrass, TypePoison}
	}
	if strings.Contains(nameLower, "charm") || strings.Contains(nameLower, "chariz") {
		return []string{TypeFire}
	}
	if strings.Contains(nameLower, "squirt") || strings.Contains(nameLower, "wart") || strings.Contains(nameLower, "blast") {
		return []string{TypeWater}
	}
	if strings.Contains(nameLower, "pika") || strings.Contains(nameLower, "rai") {
		return []string{TypeElectric}
	}
	if strings.Contains(nameLower, "geodude") || strings.Contains(nameLower, "gravel") || strings.Contains(nameLower, "golem") || strings.Contains(nameLower, "onix") {
		return []string{TypeRock, TypeGround}
	}
	if strings.Contains(nameLower, "pidgey") || strings.Contains(nameLower, "spearow") {
		return []string{TypeNormal, TypeFlying}
	}
	if strings.Contains(nameLower, "caterpie") || strings.Contains(nameLower, "weedle") {
		return []string{TypeBug}
	}
	if strings.Contains(nameLower, "rattata") {
		return []string{TypeNormal}
	}

	return []string{TypeNormal}
}

// CalculateEffectiveness calcula o dano multiplicador de um tipo atacante contra tipos defensores
func CalculateEffectiveness(attackerType string, defenderTypes []string) float64 {
	atk := normalizeType(attackerType)
	multiplier := 1.0

	for _, def := range defenderTypes {
		d := normalizeType(def)
		if row, ok := typeChart[atk]; ok {
			if val, exists := row[d]; exists {
				multiplier *= val
			}
		}
	}
	return multiplier
}

// GetBestAttackMultiplier encontra a melhor eficácia de um atacante com múltiplos tipos
func GetBestAttackMultiplier(attackerTypes []string, defenderTypes []string) float64 {
	if len(attackerTypes) == 0 {
		return 1.0
	}
	best := 0.0
	for _, atkType := range attackerTypes {
		mult := CalculateEffectiveness(atkType, defenderTypes)
		if mult > best {
			best = mult
		}
	}
	return best
}

// MatchupDescription retorna a eficácia e um rótulo legível (ex: "Super Efetivo x2.0")
func MatchupDescription(attackerTypes []string, defenderTypes []string) (float64, string) {
	mult := GetBestAttackMultiplier(attackerTypes, defenderTypes)
	if mult >= 2.0 {
		return mult, "Super Efetivo x2.0 💥"
	} else if mult > 1.0 {
		return mult, "Efetivo com Vantagem ⚔️"
	} else if mult == 1.0 {
		return mult, "Neutro (x1.0)"
	} else if mult > 0.0 {
		return mult, "Pouco Efetivo (x0.5) 🛡️"
	}
	return mult, "Sem Efeito (x0.0) 🚫"
}
