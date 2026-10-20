package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

func main() {
	data, err := os.ReadFile("app.js.tmp")
	if err != nil {
		fmt.Println("Erro lendo arquivo:", err)
		return
	}

	content := string(data)

	// Encontra todos os pacotes 't: "..."'
	rePacket := regexp.MustCompile(`t:\s*['"]([a-zA-Z0-9_.-]+)['"]`)
	matches := rePacket.FindAllStringSubmatch(content, -1)

	packetMap := make(map[string]int)
	for _, m := range matches {
		if len(m) > 1 {
			packetMap[m[1]]++
		}
	}

	var packets []string
	for p := range packetMap {
		packets = append(packets, p)
	}
	sort.Strings(packets)

	fmt.Printf("Total de tipos de pacotes 't': %d\n\n", len(packets))

	// Categoriza pacotes
	categories := map[string][]string{
		"Batalha e Caça":           {},
		"Shop e Economia":          {},
		"Equipe e Pokémons":        {},
		"Ginásios e Arenas":        {},
		"Bosses e Eventos":         {},
		"PVP e Duelos":             {},
		"Torre / Desafios":         {},
		"Guild / Clã":              {},
		"Market / Trocas":          {},
		"Pokédex / Bestiário":      {},
		"Passe / Diárias / Missões": {},
		"Social / Chat / Amigos":   {},
		"Outros":                   {},
	}

	for _, p := range packets {
		lower := strings.ToLower(p)
		switch {
		case strings.HasPrefix(lower, "hunt") || strings.Contains(lower, "atacar") || strings.Contains(lower, "captur") || strings.Contains(lower, "fugir"):
			categories["Batalha e Caça"] = append(categories["Batalha e Caça"], p)
		case strings.HasPrefix(lower, "shop") || strings.Contains(lower, "compr") || strings.Contains(lower, "vend"):
			categories["Shop e Economia"] = append(categories["Shop e Economia"], p)
		case strings.HasPrefix(lower, "team") || strings.Contains(lower, "poke") || strings.Contains(lower, "evolu") || strings.Contains(lower, "held") || strings.Contains(lower, "item"):
			categories["Equipe e Pokémons"] = append(categories["Equipe e Pokémons"], p)
		case strings.Contains(lower, "ginasio") || strings.Contains(lower, "gym"):
			categories["Ginásios e Arenas"] = append(categories["Ginásios e Arenas"], p)
		case strings.Contains(lower, "boss"):
			categories["Bosses e Eventos"] = append(categories["Bosses e Eventos"], p)
		case strings.Contains(lower, "pvp") || strings.Contains(lower, "duelo") || strings.Contains(lower, "arena"):
			categories["PVP e Duelos"] = append(categories["PVP e Duelos"], p)
		case strings.Contains(lower, "torre") || strings.Contains(lower, "tower"):
			categories["Torre / Desafios"] = append(categories["Torre / Desafios"], p)
		case strings.Contains(lower, "guild") || strings.Contains(lower, "cla"):
			categories["Guild / Clã"] = append(categories["Guild / Clã"], p)
		case strings.Contains(lower, "market") || strings.Contains(lower, "trade") || strings.Contains(lower, "troca"):
			categories["Market / Trocas"] = append(categories["Market / Trocas"], p)
		case strings.Contains(lower, "dex") || strings.Contains(lower, "pokedex"):
			categories["Pokédex / Bestiário"] = append(categories["Pokédex / Bestiário"], p)
		case strings.Contains(lower, "passe") || strings.Contains(lower, "diari") || strings.Contains(lower, "miss") || strings.Contains(lower, "conquista"):
			categories["Passe / Diárias / Missões"] = append(categories["Passe / Diárias / Missões"], p)
		case strings.Contains(lower, "chat") || strings.Contains(lower, "amigo") || strings.Contains(lower, "msg"):
			categories["Social / Chat / Amigos"] = append(categories["Social / Chat / Amigos"], p)
		default:
			categories["Outros"] = append(categories["Outros"], p)
		}
	}

	for cat, list := range categories {
		if len(list) > 0 {
			fmt.Printf("=== %s (%d) ===\n", cat, len(list))
			for _, item := range list {
				fmt.Printf("  • %s (ocorrências: %d)\n", item, packetMap[item])
			}
			fmt.Println()
		}
	}
}
