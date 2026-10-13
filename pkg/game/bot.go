package game

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"bot-template-go/pkg/auth"
	"bot-template-go/pkg/engine"
	"bot-template-go/pkg/kick"
	"bot-template-go/pkg/network"
	"bot-template-go/pkg/twitch"
)

type Bot struct {
	config  AccountConfig
	mu      sync.RWMutex
	status  engine.AccountStatus
	socket  *network.SocketClient
	httpCli *network.HTTPClient

	// Twitch IRC lurker (para manter +15% a +20% XP ativo 24/7 sem navegador)
	twitchCli *twitch.Client

	// Kick WebSocket lurker (para farmar pontos de canal 24/7 sem navegador)
	kickCli *kick.Client

	// Otimizador de Hunt e Fraquezas Elementais
	optimizer *HuntOptimizer

	// Dados do PokeIdle
	hunts       []PokeHuntArea
	currentHunt string
	noCentro    bool
	catches     int
	startXP     int64
	startedAt   time.Time
	activeID    int64
	pokemons    map[int64]PokeItemPokemon
	balls       map[string]int
	pokedex     map[string]PokeDexEntry

	ticketChan chan string
}

func NewBot(cfg AccountConfig) *Bot {
	st := engine.NewStatus(cfg.ID, cfg.Name)
	st.CurrentArea = cfg.Settings.HuntTarget

	return &Bot{
		config:      cfg,
		status:      st,
		httpCli:     network.NewHTTPClient("https://pokeidle.io", 15*time.Second),
		ticketChan:  make(chan string, 1),
		currentHunt: cfg.Settings.HuntTarget,
		startedAt:   time.Now(),
		pokemons:    make(map[int64]PokeItemPokemon),
		balls:       make(map[string]int),
		pokedex:     make(map[string]PokeDexEntry),
		twitchCli:   twitch.NewClient(cfg.Twitch),
		kickCli:     kick.NewClient(cfg.Kick),
		optimizer:   NewHuntOptimizer(),
	}
}

func (b *Bot) ID() string {
	return b.config.ID
}

func (b *Bot) Name() string {
	return b.config.Name
}

func (b *Bot) Init(ctx context.Context) error {
	b.updateStatus(func(s *engine.AccountStatus) {
		s.Status = "Autenticando..."
	})

	// 1. Se já tem token configurado em accounts.json, valida diretamente
	if b.config.Credentials.Token != "" {
		log.Printf("[%s] Token encontrado no accounts.json. Pronto para conectar.", b.ID())
		b.updateStatus(func(s *engine.AccountStatus) {
			s.Status = "Token configurado"
		})
		return nil
	}

	// 2. Se não tem token, mas tem e-mail e senha, tenta capturar via CDP local
	userOrEmail := b.config.Credentials.Email
	if userOrEmail == "" {
		userOrEmail = b.config.Name
	}

	if userOrEmail != "" && b.config.Credentials.Password != "" {
		log.Printf("[%s] Token ausente. Tentando login automático via navegador (CDP) para %s...", b.ID(), userOrEmail)
		b.updateStatus(func(s *engine.AccountStatus) {
			s.Status = "Aguardando login no navegador..."
		})

		sess, err := auth.LoginPokeIdleViaBrowserCDP(userOrEmail, b.config.Credentials.Password)
		if err == nil && sess != nil && sess.Token != "" {
			b.mu.Lock()
			b.config.Credentials.Token = sess.Token
			if sess.Nick != "" {
				b.config.Name = sess.Nick
				b.status.Name = sess.Nick
			}
			b.mu.Unlock()

			// Salva o token no accounts.json para não precisar abrir o navegador de novo
			b.saveTokenToDisk(sess.Token, sess.Nick)
			log.Printf("[%s] 🎉 Token capturado e salvo com sucesso em accounts.json!", b.ID())

			b.updateStatus(func(s *engine.AccountStatus) {
				s.Status = "Pronto para conectar"
			})
			return nil
		}
		log.Printf("[%s] Aviso no login automático via CDP: %v", b.ID(), err)
	}

	// 3. Caso não consiga automaticamente, orienta o usuário
	b.updateStatus(func(s *engine.AccountStatus) {
		s.Status = "Aguardando token (execute pegar-token.bat)"
	})
	b.addLog("Token não configurado. Execute pegar-token.bat ou insira em accounts.json.")
	return fmt.Errorf("token não configurado para conta %s", b.ID())
}

func (b *Bot) Run(ctx context.Context) error {
	b.mu.RLock()
	token := b.config.Credentials.Token
	nick := b.config.Name
	b.mu.RUnlock()

	if token == "" {
		return fmt.Errorf("token vazio ao iniciar Run")
	}

	wsURL := "wss://pokeidle.io"
	headers := make(http.Header)
	headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	headers.Set("Origin", "https://pokeidle.io")

	disconnected := make(chan struct{}, 1)

	b.socket = network.NewSocketClient(
		wsURL,
		headers,
		func(data []byte) {
			b.handleMessage(data)
		},
		func(err error) {
			log.Printf("[%s] WebSocket desconectado: %v", b.ID(), err)
			b.updateStatus(func(s *engine.AccountStatus) {
				s.Status = "reconnecting"
			})
			select {
			case disconnected <- struct{}{}:
			default:
			}
		},
	)

	log.Printf("[%s] Conectando ao WebSocket do Pokéidle (%s)...", b.ID(), wsURL)
	if err := b.socket.Connect(); err != nil {
		return fmt.Errorf("falha ao conectar WebSocket: %w", err)
	}
	defer b.socket.Close()

	// Enviar handshake 'hello'
	handshake := WSMessage{
		T:           "hello",
		Nick:        nick,
		Token:       token,
		Delta:       1,
		Dispositivo: fmt.Sprintf("d%d", time.Now().UnixMilli()),
	}

	if err := b.socket.SendJSON(handshake); err != nil {
		return fmt.Errorf("falha ao enviar handshake: %w", err)
	}

	b.updateStatus(func(s *engine.AccountStatus) {
		s.Status = "hunting"
	})
	b.addLog("Conectado ao PokeIdle. Sessão iniciada.")

	// Inicia worker lurker da Twitch em background se configurado
	if b.twitchCli != nil && b.config.Twitch.Enabled {
		b.twitchCli.Start(ctx)
	}

	// Inicia worker lurker da Kick em background se configurado
	if b.kickCli != nil && b.config.Kick.Enabled {
		b.kickCli.Start(ctx)
	}

	// Tickers para keepalive, manutenção e auto-heal
	keepaliveTicker := time.NewTicker(20 * time.Second)
	defer keepaliveTicker.Stop()

	routineTicker := time.NewTicker(45 * time.Second)
	defer routineTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-disconnected:
			return fmt.Errorf("conexão perdida com o servidor")

		case <-keepaliveTicker.C:
			if b.socket.IsConnected() {
				_ = b.socket.SendJSON(map[string]string{"t": "ping"})
			}

		case <-routineTicker.C:
			b.performPeriodicRoutine()
		}
	}
}

func (b *Bot) handleMessage(data []byte) {
	var msg WSMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	switch msg.T {
	case "pong":
		// Keepalive OK
		return

	case "sessao.renovada":
		if msg.Token != "" {
			b.mu.Lock()
			b.config.Credentials.Token = msg.Token
			b.mu.Unlock()
			b.saveTokenToDisk(msg.Token, b.config.Name)
			b.addLog("Sessão renovada pelo servidor por mais 30 dias.")
		}

	case "welcome":
		b.mu.Lock()
		if len(msg.Hunts) > 0 {
			b.hunts = msg.Hunts
		}
		b.mu.Unlock()

		if msg.Estado != nil {
			b.applyState(msg.Estado)
		}

		// Configurações iniciais recomendadas
		_ = b.socket.SendJSON(map[string]any{"t": "cliente.economia", "ativo": true})
		_ = b.socket.SendJSON(WSMessage{
			T:                   "auto.set",
			AutoBallSemParar:    true,
			AutoBallAteCapturar: false,
			BallIDs:             []int{1, 2, 3, 4},
			AutoRevive:          true,
			AutoPotion:          true,
		})
		_ = b.socket.SendJSON(WSMessage{
			T:          "shop.autoVendaLoot",
			Ativo:      true,
			AvisoVisto: true,
		})
		_ = b.socket.SendJSON(map[string]any{"t": "shop.autoLockShiny", "ativo": true})
		_ = b.socket.SendJSON(map[string]any{"t": "shop.autoLockP5", "ativo": true})
		_ = b.socket.SendJSON(map[string]any{"t": "passe.resgatar"})

		// Escolhe ou reativa a área de caça
		b.ensureHunting()

		// Otimiza pokémon ativo e equipe inicial
		go b.optimizeTeam()

	case "estado":
		if msg.Estado != nil {
			b.applyState(msg.Estado)
		}

	case "batalha":
		b.handleBattleEvents(msg.Ev)
	}
}

func (b *Bot) applyState(e *PokeState) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if e.Level > 0 {
		b.status.Level = e.Level
	}
	if e.Gold > 0 || e.Cheio {
		b.status.Gold = e.Gold
	}
	b.noCentro = e.NoCentro
	if e.HuntSlug != "" {
		if b.optimizer != nil && (b.currentHunt != e.HuntSlug || b.currentHunt == "") {
			b.optimizer.OnHuntChanged(e.HuntSlug, e.XP)
		}
		b.currentHunt = e.HuntSlug
		b.status.CurrentArea = e.HuntSlug
	}

	if b.startXP == 0 && e.XP > 0 {
		b.startXP = e.XP
	} else if e.XP > b.startXP {
		elapsedHours := time.Since(b.startedAt).Hours()
		if elapsedHours > 0.005 {
			b.status.ExpPerHour = float64(e.XP-b.startXP) / elapsedHours
		}
	}

	b.status.DurationSec = int(time.Since(b.startedAt).Seconds())
	b.status.UpdatedAt = time.Now().UnixMilli()

	// Atualiza Pokémons e Deltas
	if b.pokemons == nil {
		b.pokemons = make(map[int64]PokeItemPokemon)
	}

	if e.Cheio || len(e.Pokemons) > 0 {
		if len(e.Pokemons) > 0 {
			b.pokemons = make(map[int64]PokeItemPokemon)
			for _, p := range e.Pokemons {
				b.pokemons[p.ID] = p
			}
		}
	}

	for _, p := range e.PkMud {
		b.pokemons[p.ID] = p
	}

	for _, id := range e.PkFora {
		delete(b.pokemons, id)
	}

	if e.ActiveID > 0 {
		b.activeID = e.ActiveID
	}

	// Procura o pokémon ativo em campo
	var activePk *PokeItemPokemon
	if b.activeID > 0 {
		if pk, ok := b.pokemons[b.activeID]; ok {
			activePk = &pk
		}
	}

	if activePk == nil && len(b.pokemons) > 0 {
		for _, p := range b.pokemons {
			if p.Slot != nil {
				activePk = &p
				break
			}
		}
	}

	if activePk != nil {
		pkName := activePk.Nome
		if pkName == "" {
			pkName = activePk.Name
		}
		if activePk.Shiny {
			pkName = "✨ " + pkName
		}
		b.status.Vocation = fmt.Sprintf("%s (Nv. %d)", pkName, activePk.Level)
		b.status.HP = activePk.HP
		b.status.MaxHP = activePk.MaxHP
		b.status.Mana = activePk.Poder
		b.status.MaxMana = activePk.Poder
	}

	if e.Diamonds > 0 || e.Cheio {
		b.status.Extra["diamonds"] = e.Diamonds
	}
	if e.Orbs > 0 || e.Cheio {
		b.status.Extra["orbs"] = e.Orbs
	}
	b.status.Extra["catches"] = b.catches
	b.status.Extra["noCentro"] = e.NoCentro

	if e.Balls != nil {
		if b.balls == nil {
			b.balls = make(map[string]int)
		}
		for k, v := range e.Balls {
			b.balls[k] = v
		}
	}

	if e.Pokedex != nil {
		if b.pokedex == nil {
			b.pokedex = make(map[string]PokeDexEntry)
		}
		for k, v := range e.Pokedex {
			b.pokedex[k] = v
		}
	}

	// Telemetria do Bestiário / Pokédex
	captured, seen, pct := CountBestiaryProgress(b.pokedex)
	b.status.Extra["bestiaryCaptured"] = captured
	b.status.Extra["bestiarySeen"] = seen
	b.status.Extra["bestiaryTotal"] = 151
	b.status.Extra["bestiaryProgressPct"] = fmt.Sprintf("%.1f%%", pct)
	b.status.Extra["focusBestiary"] = b.config.Settings.FocusBestiary
	b.status.Extra["huntHasUnregistered"] = HuntHasUnregisteredSpecies(b.getCurrentHuntArea(), b.pokedex)

	// Sincroniza canais de streamers oficiais no ar para o lurker IRC da Twitch
	if e.Twitch != nil && b.twitchCli != nil {
		var liveChannels []string
		for _, l := range e.Twitch.Lives {
			if l.Login != "" {
				liveChannels = append(liveChannels, l.Login)
			}
		}
		for _, o := range e.Twitch.Oficiais {
			if o.Login != "" {
				liveChannels = append(liveChannels, o.Login)
			}
		}
		if len(liveChannels) > 0 {
			b.twitchCli.SetTargetChannels(liveChannels)
		}
	}

	// Sincroniza canais de streamers oficiais no ar da Kick
	if e.Kick != nil {
		var liveKickStreamers []string
		for _, o := range e.Kick.Oficiais {
			if o.AoVivo && o.Ativo && o.Slug != "" {
				liveKickStreamers = append(liveKickStreamers, o.Slug)
			}
		}
		if len(liveKickStreamers) > 0 {
			b.status.Extra["kickLives"] = liveKickStreamers
			if b.kickCli != nil {
				b.kickCli.SetTargetStreamers(liveKickStreamers)
			}
			go b.syncKickMinerConfig(liveKickStreamers)
		}
	}
}

func (b *Bot) handleBattleEvents(evs []PokeGameEvent) {
	for _, ev := range evs {
		switch ev.K {
		case "morte":
			if ev.Quem == "selvagem" {
				b.mu.Lock()
				b.status.Kills++
				b.status.UpdatedAt = time.Now().UnixMilli()
				if b.optimizer != nil {
					b.optimizer.RecordKill(b.currentHunt, 120)
				}
				b.mu.Unlock()

				var dropNames []string
				for _, d := range ev.Drops {
					if d.Qtd > 1 {
						dropNames = append(dropNames, fmt.Sprintf("%dx %s", d.Qtd, d.Nome))
					} else {
						dropNames = append(dropNames, d.Nome)
					}
				}
				if len(dropNames) > 0 {
					b.addLog(fmt.Sprintf("Derrotou %s! Loot: %s", ev.Nome, strings.Join(dropNames, ", ")))
				}
			} else if ev.Quem == "jogador" {
				b.addLog("Pokémon desmaiou! Curando no Centro Pokémon...")
				b.healAndResumeHunt()
			}

		case "capturado":
			b.mu.Lock()
			b.catches++
			b.status.Extra["catches"] = b.catches
			b.mu.Unlock()

			if ev.Pokemon != nil {
				shinyTag := ""
				if ev.Pokemon.Shiny {
					shinyTag = "✨ SHINY "
				}
				b.addLog(fmt.Sprintf("🎉 %s%s capturado com sucesso! (Nv. %d)", shinyTag, ev.Pokemon.Nome, ev.Pokemon.Level))

				// Checagem de novo registro no Bestiário / Pokédex
				b.mu.Lock()
				pokeIDStr := fmt.Sprintf("%d", ev.Pokemon.PokeID)
				entry := b.pokedex[pokeIDStr]
				isNewBestiary := entry.C == 0
				entry.C++
				b.pokedex[pokeIDStr] = entry
				captured, _, pct := CountBestiaryProgress(b.pokedex)
				b.status.Extra["bestiaryCaptured"] = captured
				b.status.Extra["bestiaryProgressPct"] = fmt.Sprintf("%.1f%%", pct)
				b.mu.Unlock()

				if isNewBestiary {
					b.addLog(fmt.Sprintf("📖 [Bestiário] NOVO REGISTRO DESBLOQUEADO: %s (#%d)! Progresso da Pokédex: %d/151 (%.1f%%)",
						ev.Pokemon.Nome, ev.Pokemon.PokeID, captured, pct))

					// Se focado em completar o Bestiário, checa se a rota atual foi concluída
					if b.config.Settings.FocusBestiary {
						currentArea := b.getCurrentHuntArea()
						if !HuntHasUnregisteredSpecies(currentArea, b.pokedex) {
							b.addLog(fmt.Sprintf("🏆 [Bestiário] 100%% das espécies de '%s' capturadas! Procurando próxima rota com espécies pendentes...", currentArea.Nome))
							go b.ensureHunting()
						}
					}
				}
			}
			go b.optimizeTeam()

		case "levelup":
			if ev.Quem == "treinador" {
				b.addLog(fmt.Sprintf("🌟 Treinador subiu para o Nível %d!", ev.Level))
				go b.ensureHunting()
			} else {
				b.addLog(fmt.Sprintf("⬆️ Pokémon subiu para o Nível %d!", ev.Level))
			}
			go b.optimizeTeam()

		case "evolucao":
			b.addLog(fmt.Sprintf("🌟 Pokémon evoluiu para %s!", ev.Nome))
		}
	}
}

func (b *Bot) getCurrentHuntArea() PokeHuntArea {
	for _, h := range b.hunts {
		if h.Slug == b.currentHunt {
			return h
		}
	}
	return PokeHuntArea{Slug: b.currentHunt, Nome: b.currentHunt, Nivel: 1}
}

func (b *Bot) ensureHunting() {
	b.mu.RLock()
	target := b.config.Settings.HuntTarget
	current := b.currentHunt
	userLevel := b.status.Level
	hunts := b.hunts
	pokemons := b.pokemons
	pokedex := b.pokedex
	focusBestiary := b.config.Settings.FocusBestiary
	b.mu.RUnlock()

	slugToHunt := target
	if slugToHunt == "" || slugToHunt == "auto" {
		if len(hunts) > 0 && b.optimizer != nil {
			bestHunt, _, rationale := b.optimizer.FindBestHuntForMaxXP(hunts, pokemons, userLevel, pokedex, focusBestiary)
			if bestHunt.Slug != "" {
				slugToHunt = bestHunt.Slug
				b.addLog(fmt.Sprintf("🎯 [Otimizador de XP/h] Rota recomendada: %s (Nv. %d) — %s", bestHunt.Nome, bestHunt.Nivel, rationale))
			}
		}
		if slugToHunt == "" || slugToHunt == "auto" {
			if current != "" {
				slugToHunt = current
			} else if len(hunts) > 0 {
				bestSlug := hunts[0].Slug
				bestLevel := 0
				for _, h := range hunts {
					if userLevel >= h.Nivel && h.Nivel >= bestLevel {
						bestLevel = h.Nivel
						bestSlug = h.Slug
					}
				}
				slugToHunt = bestSlug
			}
		}
	}

	if slugToHunt != "" && b.socket != nil && b.socket.IsConnected() {
		_ = b.socket.SendJSON(WSMessage{
			T:    "hunt.select",
			Slug: slugToHunt,
		})
		b.mu.Lock()
		b.currentHunt = slugToHunt
		b.status.CurrentArea = slugToHunt
		b.mu.Unlock()
	}
}

func (b *Bot) healAndResumeHunt() {
	if b.socket == nil || !b.socket.IsConnected() {
		return
	}
	_ = b.socket.SendJSON(map[string]any{"t": "centro.curar"})
	time.Sleep(1 * time.Second)
	b.ensureHunting()
}

func (b *Bot) performPeriodicRoutine() {
	if b.socket == nil || !b.socket.IsConnected() {
		return
	}

	b.mu.RLock()
	noCentro := b.noCentro
	autoHeal := b.config.Settings.AutoHeal
	autoSell := b.config.Settings.AutoSell
	autoDaily := b.config.Settings.AutoDaily
	hp := b.status.HP
	gold := b.status.Gold
	ballCount := b.balls["1"]
	b.mu.RUnlock()

	// 1. Se estiver no Centro Pokémon ou com HP zerado, cura e retoma caça
	if (noCentro || hp == 0) && autoHeal {
		b.healAndResumeHunt()
	}

	// 2. Venda automática de itens acumulados
	if autoSell {
		_ = b.socket.SendJSON(map[string]any{"t": "shop.sellAllItems"})
	}

	// 3. Resgate diário de passe
	if autoDaily {
		_ = b.socket.SendJSON(map[string]any{"t": "passe.resgatar"})
	}

	// 4. Auto-Restock de Pokébolas se a quantidade for baixa e tiver ouro
	minBalls := 15
	currentHuntArea := b.getCurrentHuntArea()
	hasUnregistered := HuntHasUnregisteredSpecies(currentHuntArea, b.pokedex)
	if b.config.Settings.FocusBestiary || hasUnregistered {
		minBalls = 30
	}
	if ballCount < minBalls && gold >= 3000 {
		_ = b.socket.SendJSON(WSMessage{
			T:      "shop.buy",
			Kind:   "ball",
			ItemID: 1,
			Qty:    30,
		})
		b.addLog("Reabasteceu 30x Pokébolas no Shop automaticamente para garantir capturas.")
	}

	// 5. Se estiver caçando em rota com monstros ainda não registrados, força captura contínua
	if hasUnregistered || b.config.Settings.FocusBestiary {
		_ = b.socket.SendJSON(WSMessage{
			T:                   "auto.set",
			AutoBallSemParar:    true,
			AutoBallAteCapturar: true,
			BallIDs:             []int{1, 2, 3, 4},
			AutoRevive:          true,
			AutoPotion:          true,
		})
	}

	// 6. Otimização contínua do time e do Pokémon ativo
	b.optimizeTeam()
}

// optimizeTeam seleciona os melhores pokémons com base em Fraquezas Elementais, Poder e nível do treinador
func (b *Bot) optimizeTeam() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.pokemons) == 0 || b.socket == nil || !b.socket.IsConnected() {
		return
	}

	trainerLevel := b.status.Level

	// Regra de uso do Pokéidle: Starter sempre utilizável; demais até trainerLevel + 5
	var usable []PokeItemPokemon
	for _, p := range b.pokemons {
		if p.Starter || (trainerLevel > 0 && p.Level <= trainerLevel+5) || trainerLevel == 0 {
			usable = append(usable, p)
		}
	}

	if len(usable) == 0 {
		return
	}

	// 1. BALANCEAMENTO DE NÍVEL DA EQUIPE COM TROCA DE HUNT:
	// Identifica se há algum membro da equipe ficando para trás e precisando de treinamento
	targetPk, shouldRotate := SelectNextPokemonForBalancedTeam(b.pokemons, trainerLevel, b.activeID)
	if shouldRotate && targetPk.ID != b.activeID && (b.config.Settings.HuntTarget == "" || b.config.Settings.HuntTarget == "auto") {
		// Encontra a melhor rota especificamente vantajosa para os tipos desse Pokémon e com prioridade no Bestiário
		bestHunt, matchup, found := b.optimizer.FindBestHuntForPokemon(b.hunts, targetPk, trainerLevel, b.pokedex, b.config.Settings.FocusBestiary)
		if found && bestHunt.Slug != "" && bestHunt.Slug != b.currentHunt {
			_ = b.socket.SendJSON(WSMessage{
				T:    "hunt.select",
				Slug: bestHunt.Slug,
			})
			b.currentHunt = bestHunt.Slug
			b.status.CurrentArea = bestHunt.Slug
			pkName := targetPk.Nome
			if pkName == "" {
				pkName = targetPk.Name
			}
			b.status.Logs = append(b.status.Logs, fmt.Sprintf("🔄 [Troca de Hunt Automática] Migrou para '%s' (Nv. %d) para dar vantagem %s ao %s (Nv. %d)!",
				bestHunt.Nome, bestHunt.Nivel, matchup.MatchupLabel, pkName, targetPk.Level))
		}

		// Equipa o Pokémon no campo de batalha
		_ = b.socket.SendJSON(WSMessage{
			T:         "team.active",
			PokemonID: targetPk.ID,
		})
		b.activeID = targetPk.ID
		pkName := targetPk.Nome
		if pkName == "" {
			pkName = targetPk.Name
		}
		b.status.Logs = append(b.status.Logs, fmt.Sprintf("🐾 [Equipe Nivelada] Escalou %s (Nv. %d) para subir de nível e acompanhar o time!",
			pkName, targetPk.Level))
	}

	currentHuntArea := b.getCurrentHuntArea()
	defenderTypes := GetHuntDefenderTypes(currentHuntArea)

	// Avalia todos os utilizáveis contra a rota atual considerando fraqueza elemental
	type scoredPokemon struct {
		pk      PokeItemPokemon
		matchup MatchupResult
	}
	var scoredList []scoredPokemon
	for _, p := range usable {
		m := EvaluatePokemonAgainstHunt(p, defenderTypes)
		scoredList = append(scoredList, scoredPokemon{pk: p, matchup: m})
	}

	// Ordena por Dano Efetivo (Poder * Multiplicador de Fraqueza) decrescente
	sort.Slice(scoredList, func(i, j int) bool {
		if scoredList[i].matchup.EffectivePower != scoredList[j].matchup.EffectivePower {
			return scoredList[i].matchup.EffectivePower > scoredList[j].matchup.EffectivePower
		}
		if scoredList[i].pk.Shiny != scoredList[j].pk.Shiny {
			return scoredList[i].pk.Shiny
		}
		return scoredList[i].pk.Level > scoredList[j].pk.Level
	})

	bestChoice := scoredList[0]
	bestPk := bestChoice.pk
	bestMatchup := bestChoice.matchup

	// Salva telemetria de matchup nos extras da conta
	if b.status.Extra == nil {
		b.status.Extra = make(map[string]any)
	}
	b.status.Extra["matchup"] = bestMatchup.MatchupLabel
	b.status.Extra["matchupMultiplier"] = bestMatchup.Multiplier
	b.status.Extra["effectivePower"] = int(bestMatchup.EffectivePower)
	b.status.Extra["pokemonTypes"] = strings.Join(bestMatchup.PokemonTypes, " / ")
	b.status.Extra["huntDefenders"] = strings.Join(defenderTypes, " / ")

	// 2. Se não estava em rotação forçada e o ativo não for o mais forte com vantagem da rota:
	if !shouldRotate && bestPk.ID != b.activeID {
		_ = b.socket.SendJSON(WSMessage{
			T:         "team.active",
			PokemonID: bestPk.ID,
		})
		b.activeID = bestPk.ID
		pkName := bestPk.Nome
		if pkName == "" {
			pkName = bestPk.Name
		}
		b.status.Logs = append(b.status.Logs, fmt.Sprintf("⚔️ [Auto-Fraqueza] Equipou %s (%s) com %s contra a rota (Dano Efetivo: %.0f)!",
			pkName, strings.Join(bestMatchup.PokemonTypes, "/"), bestMatchup.MatchupLabel, bestMatchup.EffectivePower))
	}

	// 2. Preenche os 5 slots da equipe com os melhores Pokémon utilizáveis
	const maxEquipe = 5
	topLimit := len(scoredList)
	if topLimit > maxEquipe {
		topLimit = maxEquipe
	}

	// Identifica quais slots (0 a 4) já estão ocupados
	slotsOcupados := make(map[int]int64)
	for _, p := range b.pokemons {
		if p.Slot != nil {
			slotsOcupados[*p.Slot] = p.ID
		}
	}

	// Para cada um dos top pokémons, se não estiver em nenhum slot, move para o primeiro vago
	for i := 0; i < topLimit; i++ {
		topP := scoredList[i].pk
		if topP.Slot == nil {
			for slot := 0; slot < maxEquipe; slot++ {
				if _, ocupado := slotsOcupados[slot]; !ocupado {
					slotIdx := slot
					_ = b.socket.SendJSON(WSMessage{
						T:         "team.move",
						PokemonID: topP.ID,
						Slot:      &slotIdx,
					})
					slotsOcupados[slot] = topP.ID
					topP.Slot = &slotIdx
					b.pokemons[topP.ID] = topP
					break
				}
			}
		}
	}
}

func (b *Bot) GetStatus() engine.AccountStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status
}

func (b *Bot) ExecuteAction(action string, payload map[string]any) (any, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.socket == nil || !b.socket.IsConnected() {
		return nil, fmt.Errorf("bot não está conectado ao jogo")
	}

	switch action {
	case "huntTarget":
		if target, ok := payload["slug"].(string); ok && target != "" {
			b.config.Settings.HuntTarget = target
			b.currentHunt = target
			b.status.CurrentArea = target
			_ = b.socket.SendJSON(WSMessage{T: "hunt.select", Slug: target})
			b.status.Logs = append(b.status.Logs, fmt.Sprintf("Área alterada para: %s", target))
			return fmt.Sprintf("Área alterada para %s", target), nil
		}
		return nil, fmt.Errorf("slug inválido")

	case "heal":
		_ = b.socket.SendJSON(map[string]any{"t": "centro.curar"})
		b.status.Logs = append(b.status.Logs, "Comando de cura enviado.")
		return "Cura enviada!", nil

	case "sell":
		_ = b.socket.SendJSON(map[string]any{"t": "shop.sellAllItems"})
		b.status.Logs = append(b.status.Logs, "Venda de itens solicitada.")
		return "Venda solicitada!", nil

	case "claimPass":
		_ = b.socket.SendJSON(map[string]any{"t": "passe.resgatar"})
		b.status.Logs = append(b.status.Logs, "Resgate do passe solicitado.")
		return "Resgate solicitado!", nil

	case "optimizeTeam":
		go b.optimizeTeam()
		b.status.Logs = append(b.status.Logs, "Otimização de equipe solicitada.")
		return "Equipe otimizada com sucesso!", nil

	default:
		return nil, fmt.Errorf("ação desconhecida: %s", action)
	}
}

func (b *Bot) SupplyTicket(ticket, wsURL string) error {
	log.Printf("[%s] Recebido novo token externo: %s", b.ID(), ticket)
	b.mu.Lock()
	b.config.Credentials.Token = ticket
	b.mu.Unlock()
	b.saveTokenToDisk(ticket, b.config.Name)
	return nil
}

func (b *Bot) saveTokenToDisk(token, nick string) {
	configFile := "accounts.json"
	content, err := os.ReadFile(configFile)
	if err != nil {
		return
	}

	var accs []AccountConfig
	if err := json.Unmarshal(content, &accs); err != nil {
		return
	}

	for i := range accs {
		if accs[i].ID == b.config.ID {
			accs[i].Credentials.Token = token
			if nick != "" {
				accs[i].Name = nick
			}
			break
		}
	}

	if out, err := json.MarshalIndent(accs, "", "  "); err == nil {
		_ = os.WriteFile(configFile, out, 0644)
	}
}

func (b *Bot) updateStatus(fn func(s *engine.AccountStatus)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	fn(&b.status)
	b.status.UpdatedAt = time.Now().UnixMilli()
}

func (b *Bot) addLog(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status.Logs = append(b.status.Logs, msg)
	if len(b.status.Logs) > 10 {
		b.status.Logs = b.status.Logs[1:]
	}
}

func (b *Bot) syncKickMinerConfig(liveStreamers []string) {
	configFile := "kick-miner/config.json"
	data, err := os.ReadFile(configFile)
	if err != nil {
		return
	}

	var cfg struct {
		Accounts []struct {
			Username  string   `json:"username"`
			Token     string   `json:"token"`
			Streamers []string `json:"streamers"`
		} `json:"accounts"`
		Settings map[string]any `json:"settings"`
	}

	if err := json.Unmarshal(data, &cfg); err != nil || len(cfg.Accounts) == 0 {
		return
	}

	seen := make(map[string]bool)
	var updatedList []string
	for _, s := range liveStreamers {
		clean := strings.ToLower(strings.TrimSpace(s))
		if clean != "" && !seen[clean] {
			seen[clean] = true
			updatedList = append(updatedList, clean)
		}
	}
	for _, s := range cfg.Accounts[0].Streamers {
		clean := strings.ToLower(strings.TrimSpace(s))
		if clean != "" && !seen[clean] {
			seen[clean] = true
			updatedList = append(updatedList, clean)
		}
	}

	if len(updatedList) > 0 {
		cfg.Accounts[0].Streamers = updatedList
		if out, err := json.MarshalIndent(cfg, "", "  "); err == nil {
			_ = os.WriteFile(configFile, out, 0644)
		}
	}
}

