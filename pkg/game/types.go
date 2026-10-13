package game

import (
	"bot-template-go/pkg/kick"
	"bot-template-go/pkg/twitch"
)

// AccountConfig modelo de configuração lido do accounts.json
type AccountConfig struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Enabled     bool            `json:"enabled"`
	Credentials GameCredentials `json:"credentials"`
	Settings    GameSettings    `json:"settings"`
	Twitch      twitch.Config   `json:"twitch,omitempty"`
	Kick        kick.Config     `json:"kick,omitempty"`
}

type GameCredentials struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	SessionCookie string `json:"sessionCookie"`
	Token         string `json:"token"`
}

type GameSettings struct {
	HuntTarget    string `json:"huntTarget"` // Slug da hunt (ex: "pallet-town", "route-1")
	AutoHunt      bool   `json:"autoHunt"`
	AutoSell      bool   `json:"autoSell"`
	AutoHeal      bool   `json:"autoHeal"`
	AutoDaily     bool   `json:"autoDaily"`
	AutoCatch     bool   `json:"autoCatch"`
	FocusBestiary bool   `json:"focusBestiary"` // Priorizar preenchimento do Bestiário / Pokédex
	Lure          int    `json:"lure"`
}

// PokeIdle Inbound & Outbound Packets
type WSMessage struct {
	T string `json:"t"`

	// Handshake
	Nick        string `json:"nick,omitempty"`
	Token       string `json:"token,omitempty"`
	Delta       int    `json:"delta,omitempty"`
	Dispositivo string `json:"dispositivo,omitempty"`

	// Hunt
	Slug string `json:"slug,omitempty"`

	// Automation
	AutoBallSemParar    bool  `json:"autoBallSemParar,omitempty"`
	AutoBallAteCapturar bool  `json:"autoBallAteCapturar,omitempty"`
	BallIDs             []int `json:"ballIds,omitempty"`
	AutoRevive          bool  `json:"autoRevive,omitempty"`
	AutoPotion          bool  `json:"autoPotion,omitempty"`

	// Shop / Loot / Team
	Kind       string `json:"kind,omitempty"`
	ItemID     any    `json:"id,omitempty"`
	Qty        int    `json:"qty,omitempty"`
	PokemonID  int64  `json:"pokemonId,omitempty"`
	Slot       *int   `json:"slot,omitempty"`
	Ativo      bool   `json:"ativo,omitempty"`
	AvisoVisto bool   `json:"avisoVisto,omitempty"`

	// Server responses
	Estado *PokeState      `json:"estado,omitempty"`
	Hunts  []PokeHuntArea  `json:"hunts,omitempty"`
	Ev     []PokeGameEvent `json:"ev,omitempty"`
}

type PokeState struct {
	Cheio      bool               `json:"cheio,omitempty"`
	Nick       string             `json:"nick"`
	Level      int                `json:"level"`
	XP         int64              `json:"xp"`
	XPNivel    int64              `json:"xpNivel"`
	XPProximo  int64              `json:"xpProximo"`
	Gold       int64              `json:"gold"`
	Diamonds   int64              `json:"diamonds"`
	Orbs       int64              `json:"orbs"`
	ActiveID   int64              `json:"activeId"`
	HuntSlug   string             `json:"huntSlug"`
	NoCentro   bool               `json:"noCentro"`
	Pokemons   []PokeItemPokemon  `json:"pokemons,omitempty"`
	PkMud      []PokeItemPokemon  `json:"pkMud,omitempty"`
	PkFora     []int64            `json:"pkFora,omitempty"`
	Items      map[string]int          `json:"items,omitempty"`
	Balls      map[string]int          `json:"balls,omitempty"`
	Automation PokeAutomationPref      `json:"automation,omitempty"`
	Pokedex    map[string]PokeDexEntry `json:"pokedex,omitempty"`
	Boss       *PokeBossState          `json:"boss,omitempty"`
	Twitch     *PokeTwitchState        `json:"twitch,omitempty"`
	Kick       *PokeKickState          `json:"kick,omitempty"`
}

type PokeBossState struct {
	Vitorias map[string]int `json:"vitorias,omitempty"`
	Tokens   int            `json:"tokens,omitempty"`
}

type PokeDexEntry struct {
	K  int `json:"k"`  // Kills (derrotas de monstros daquela espécie)
	C  int `json:"c"`  // Capturas (registradas na Pokébola)
	SV int `json:"sv"` // Shinys vistos no selvagem
	SC int `json:"sc"` // Shinys capturados
}

// GenerationInfo armazena dados de uma geração / região específica da Pokédex
type GenerationInfo struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Region   string  `json:"region"`
	MinID    int     `json:"minId"`
	MaxID    int     `json:"maxId"`
	Total    int     `json:"total"`
	Captured int     `json:"captured"`
	Seen     int     `json:"seen"`
	Pct      float64 `json:"pct"`
	Complete bool    `json:"complete"`
}

// BestiaryFullReport engloba todas as facetas do Bestiário (Nacional, Gerações 1-9, Outland, Megas, Shiny Dex e Boss Dex)
type BestiaryFullReport struct {
	NationalCaptured int              `json:"nationalCaptured"`
	NationalSeen     int              `json:"nationalSeen"`
	NationalTotal    int              `json:"nationalTotal"`
	NationalPct      float64          `json:"nationalPct"`
	ShinysCaptured   int              `json:"shinysCaptured"`
	ShinysSeen       int              `json:"shinysSeen"`
	ShinysPct        float64          `json:"shinysPct"`
	BossesDefeated   int              `json:"bossesDefeated"`
	BossesTotal      int              `json:"bossesTotal"`
	BossesPct        float64          `json:"bossesPct"`
	TotalKills       int64            `json:"totalKills"`
	TotalCaptures    int64            `json:"totalCaptures"`
	Generations      []GenerationInfo `json:"generations"`
}

type PokeTwitchState struct {
	Login        string           `json:"login,omitempty"`
	AoVivo       bool             `json:"aoVivo,omitempty"`
	Assistindo   bool             `json:"assistindo,omitempty"`
	PctAtual     float64          `json:"pctAtual,omitempty"`
	Canal        string           `json:"canal,omitempty"`
	Lives        []PokeTwitchLive `json:"lives,omitempty"`
	Oficiais     []PokeTwitchLive `json:"oficiais,omitempty"`
	AssistindoEm []string         `json:"assistindoEm,omitempty"`
}

type PokeTwitchLive struct {
	Login string `json:"login"`
	Nome  string `json:"nome,omitempty"`
	Bonus bool   `json:"bonus,omitempty"`
}

type PokeKickState struct {
	Vinculado   bool              `json:"vinculado,omitempty"`
	Login       string            `json:"login,omitempty"`
	Pct         float64           `json:"pct,omitempty"`
	Minutos     int               `json:"minutos,omitempty"`
	Bonus       map[string]int64  `json:"bonus,omitempty"`
	Oficiais    []PokeKickOficial `json:"oficiais,omitempty"`
	Recompensas []PokeKickReward  `json:"recompensas,omitempty"`
}

type PokeKickOficial struct {
	Slug   string `json:"slug"`
	Nome   string `json:"nome,omitempty"`
	AoVivo bool   `json:"aoVivo,omitempty"`
	Ativo  bool   `json:"ativo,omitempty"`
}

type PokeKickReward struct {
	Tipo  string `json:"tipo"`
	Custo int    `json:"custo"`
}

type PokeItemPokemon struct {
	ID        int64  `json:"id"`
	PokeID    int    `json:"pokeId"`
	Nome      string `json:"nome"`
	Name      string `json:"name"`
	Level     int    `json:"level"`
	HP        int    `json:"hp"`
	MaxHP     int    `json:"maxHp"`
	Poder     int    `json:"poder"`
	Shiny     bool   `json:"shiny"`
	Starter   bool   `json:"starter,omitempty"`
	Slot      *int   `json:"slot"`
	XP        int64  `json:"xp"`
	XPProximo int64  `json:"xpProximo"`
}

type PokeHuntArea struct {
	Slug     string            `json:"slug"`
	Nome     string            `json:"nome"`
	Nivel    int               `json:"nivel"`
	Regiao   string            `json:"regiao,omitempty"`
	Especies []PokeHuntSpecies `json:"especies,omitempty"`
}

type PokeHuntSpecies struct {
	PokeID int `json:"pokeId"`
	Pontos int `json:"pontos,omitempty"`
}

type PokeGameEvent struct {
	K             string          `json:"k"` // "morte", "ataque", "capturado", "levelup", "evolucao", "bola", "fugiu", etc.
	Quem          string          `json:"quem,omitempty"`
	Por           string          `json:"por,omitempty"`
	Nome          string          `json:"nome,omitempty"`
	Level         int             `json:"level,omitempty"`
	Drops         []PokeDrop      `json:"drops,omitempty"`
	Pokemon       *PokeCapturedPk `json:"pokemon,omitempty"`
	LevelSelvagem int             `json:"levelSelvagem,omitempty"`
}

type PokeDrop struct {
	Nome   string `json:"nome"`
	ItemID any    `json:"itemId"`
	Qtd    int    `json:"qtd"`
}

type PokeCapturedPk struct {
	ID      int64  `json:"id"`
	PokeID  int    `json:"pokeId,omitempty"`
	Nome    string `json:"nome"`
	Level   int    `json:"level"`
	Shiny   bool   `json:"shiny"`
	Quality string `json:"quality,omitempty"`
}

type PokeAutomationPref struct {
	PvpAutoFila         bool  `json:"pvpAutoFila,omitempty"`
	AutoRevive          bool  `json:"autoRevive,omitempty"`
	AutoPotion          bool  `json:"autoPotion,omitempty"`
	AutoBallSemParar    bool  `json:"autoBallSemParar,omitempty"`
	AutoBallAteCapturar bool  `json:"autoBallAteCapturar,omitempty"`
	BallIDs             []int `json:"ballIds,omitempty"`
}
