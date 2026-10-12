package kick

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Config struct {
	Enabled     bool     `json:"enabled"`
	BearerToken string   `json:"bearerToken"`
	Streamers   []string `json:"streamers,omitempty"`
}

type Client struct {
	cfg        Config
	mu         sync.RWMutex
	httpCli    *http.Client
	streamers  map[string]bool
	activeMap  map[string]*streamWatcher
	running    bool
	cancelFunc context.CancelFunc
}

func NewClient(cfg Config) *Client {
	c := &Client{
		cfg: cfg,
		httpCli: &http.Client{
			Timeout: 15 * time.Second,
		},
		streamers: make(map[string]bool),
		activeMap: make(map[string]*streamWatcher),
	}

	for _, s := range cfg.Streamers {
		clean := strings.ToLower(strings.TrimSpace(s))
		if clean != "" {
			c.streamers[clean] = true
		}
	}

	return c
}

func (c *Client) Start(ctx context.Context) {
	if !c.cfg.Enabled || c.cfg.BearerToken == "" {
		return
	}

	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}
	c.running = true
	ctx, c.cancelFunc = context.WithCancel(ctx)
	c.mu.Unlock()

	log.Printf("[Kick Lurker] Minerador nativo Kick em Go ativo (0 navegadores, 0 Python, ~2MB RAM).")

	go c.runLoop(ctx)
}

func (c *Client) SetTargetStreamers(streamers []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, s := range streamers {
		clean := strings.ToLower(strings.TrimSpace(s))
		if clean != "" {
			c.streamers[clean] = true
		}
	}
}

func (c *Client) runLoop(ctx context.Context) {
	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()

	// Checagem imediata inicial
	c.checkAndWatchStreams(ctx)

	for {
		select {
		case <-ctx.Done():
			c.Close()
			return
		case <-ticker.C:
			c.checkAndWatchStreams(ctx)
		}
	}
}

func (c *Client) checkAndWatchStreams(ctx context.Context) {
	c.mu.RLock()
	targets := make([]string, 0, len(c.streamers))
	for s := range c.streamers {
		targets = append(targets, s)
	}

	// Contar quantos estão ativos atualmente (limitar a 2 simultâneos para evitar rate-limits)
	activeCount := 0
	for _, w := range c.activeMap {
		if w.isAlive() {
			activeCount++
		}
	}
	c.mu.RUnlock()

	if len(targets) == 0 {
		targets = []string{"tftoddy", "faasii", "axtlol", "xdaren_", "cogu"}
	}

	if activeCount >= 2 {
		return
	}

	for _, streamer := range targets {
		c.mu.RLock()
		watcher, exists := c.activeMap[streamer]
		c.mu.RUnlock()

		if exists && watcher.isAlive() {
			continue
		}

		c.tryWatchStreamer(ctx, streamer)

		// Verifica se atingiu limite
		c.mu.RLock()
		currActive := 0
		for _, w := range c.activeMap {
			if w.isAlive() {
				currActive++
			}
		}
		c.mu.RUnlock()
		if currActive >= 2 {
			break
		}
	}
}

func (c *Client) tryWatchStreamer(ctx context.Context, streamer string) {
	info, err := c.getChannelInfo(streamer)
	if err != nil || info == nil || !info.IsLive {
		return
	}

	c.mu.Lock()
	if existing, ok := c.activeMap[streamer]; ok && existing.isAlive() {
		c.mu.Unlock()
		return
	}

	watcher := newStreamWatcher(c, c.cfg.BearerToken, streamer, info.ChannelID, info.UserID, info.StreamID)
	c.activeMap[streamer] = watcher
	c.mu.Unlock()

	go watcher.run(ctx)
}

type channelInfo struct {
	ChannelID int64
	UserID    int64
	StreamID  int64
	IsLive    bool
}

func (c *Client) getChannelInfo(streamer string) (*channelInfo, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://kick.com/api/v2/channels/%s", streamer), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Authorization", "Bearer "+c.cfg.BearerToken)
	req.Header.Set("Origin", "https://kick.com")
	req.Header.Set("Referer", fmt.Sprintf("https://kick.com/%s/", streamer))

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	chIDFloat, ok := data["id"].(float64)
	if !ok {
		return nil, fmt.Errorf("sem id de canal")
	}

	userIDFloat, _ := data["user_id"].(float64)
	livestreamMap, isLiveMap := data["livestream"].(map[string]any)

	info := &channelInfo{
		ChannelID: int64(chIDFloat),
		UserID:    int64(userIDFloat),
		IsLive:    isLiveMap && livestreamMap != nil,
	}

	if info.IsLive {
		if streamIDFloat, ok := livestreamMap["id"].(float64); ok {
			info.StreamID = int64(streamIDFloat)
		}
	}

	return info, nil
}

func (c *Client) FetchPoints(streamer string) (int, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://kick.com/api/v2/channels/%s/points", streamer), nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Authorization", "Bearer "+c.cfg.BearerToken)
	req.Header.Set("Origin", "https://kick.com")
	req.Header.Set("Referer", fmt.Sprintf("https://kick.com/%s/", streamer))

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var res struct {
		Data struct {
			Points int `json:"points"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return 0, err
	}

	return res.Data.Points, nil
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.running = false
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
	for _, w := range c.activeMap {
		w.close()
	}
	c.activeMap = make(map[string]*streamWatcher)
}

// streamWatcher gerencia a conexão de visualização ao vivo no WebSocket da Kick
type streamWatcher struct {
	client    *Client
	token     string
	streamer  string
	channelID int64
	userID    int64
	streamID  int64
	mu        sync.RWMutex
	conn      *websocket.Conn
	alive     bool
}

func newStreamWatcher(cli *Client, token, streamer string, chID, uID, sID int64) *streamWatcher {
	return &streamWatcher{
		client:    cli,
		token:     token,
		streamer:  streamer,
		channelID: chID,
		userID:    uID,
		streamID:  sID,
		alive:     true,
	}
}

func (w *streamWatcher) isAlive() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.alive
}

func (w *streamWatcher) close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.alive = false
	if w.conn != nil {
		_ = w.conn.Close()
		w.conn = nil
	}
}

func (w *streamWatcher) run(ctx context.Context) {
	defer func() {
		w.mu.Lock()
		w.alive = false
		if w.conn != nil {
			_ = w.conn.Close()
			w.conn = nil
		}
		w.mu.Unlock()
		log.Printf("[Kick Lurker] Conexão com live de '%s' finalizada.", w.streamer)
	}()

	// 1. Obter Viewer WebSocket Token
	wsToken, err := w.fetchViewerToken()
	if err != nil || wsToken == "" {
		log.Printf("[Kick Lurker] Erro ao obter WS token para '%s': %v", w.streamer, err)
		return
	}

	// 2. Conectar ao WebSocket
	wsURL := fmt.Sprintf("wss://websockets.kick.com/viewer/v1/connect?token=%s", wsToken)
	headers := make(http.Header)
	headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	headers.Set("Origin", "https://kick.com")

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(wsURL, headers)
	if err != nil {
		log.Printf("[Kick Lurker] Falha ao conectar WS de '%s': %v", w.streamer, err)
		return
	}
	defer conn.Close()

	w.mu.Lock()
	w.conn = conn
	w.mu.Unlock()

	// Consulta e loga saldo inicial de pontos
	if pts, err := w.client.FetchPoints(w.streamer); err == nil {
		log.Printf("[Kick Lurker] Conectado à live de '%s' na Kick! Pontos acumulados: %d (+10 a cada 5m)!", w.streamer, pts)
	} else {
		log.Printf("[Kick Lurker] Conectado à live de '%s' na Kick (ganhando pontos a cada 5m)!", w.streamer)
	}

	// 3. Handshake inicial + ping
	_ = conn.WriteJSON(map[string]any{
		"type": "channel_handshake",
		"data": map[string]any{
			"message": map[string]any{
				"channelId": w.channelID,
			},
		},
	})
	_ = conn.WriteJSON(map[string]any{"type": "ping"})

	// Tickers para manter presença e checagem periódica
	handshakeTicker := time.NewTicker(30 * time.Second)
	defer handshakeTicker.Stop()

	trackTicker := time.NewTicker(10 * time.Second)
	defer trackTicker.Stop()

	pointsTicker := time.NewTicker(5 * time.Minute)
	defer pointsTicker.Stop()

	disconnectChan := make(chan struct{}, 1)

	// Goroutine de envio contínuo de tracking e handshakes
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-disconnectChan:
				return
			case <-handshakeTicker.C:
				w.mu.Lock()
				if w.conn != nil {
					_ = w.conn.WriteJSON(map[string]any{
						"type": "channel_handshake",
						"data": map[string]any{
							"message": map[string]any{
								"channelId": w.channelID,
							},
						},
					})
					_ = w.conn.WriteJSON(map[string]any{"type": "ping"})
				}
				w.mu.Unlock()

			case <-trackTicker.C:
				w.mu.Lock()
				if w.conn != nil {
					err := w.conn.WriteJSON(map[string]any{
						"type": "user_event",
						"data": map[string]any{
							"message": map[string]any{
								"name":          "tracking.user.watch.livestream",
								"channel_id":    w.channelID,
								"livestream_id": w.streamID,
							},
						},
					})
					if err != nil {
						w.mu.Unlock()
						return
					}
				}
				w.mu.Unlock()

			case <-pointsTicker.C:
				if pts, err := w.client.FetchPoints(w.streamer); err == nil {
					log.Printf("[Kick Lurker] Live de '%s': saldo atualizado = %d pontos (+10 creditados)!", w.streamer, pts)
				}
			}
		}
	}()

	// Loop de leitura de mensagens e resposta a pings do servidor
	for {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		_, msg, err := conn.ReadMessage()
		if err != nil {
			select {
			case disconnectChan <- struct{}{}:
			default:
			}
			break
		}
		var parsed map[string]any
		if err := json.Unmarshal(msg, &parsed); err == nil {
			if parsed["type"] == "ping" {
				w.mu.Lock()
				if w.conn != nil {
					_ = w.conn.WriteJSON(map[string]any{"type": "pong"})
				}
				w.mu.Unlock()
			}
		}
	}
}

func (w *streamWatcher) fetchViewerToken() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://websockets.kick.com/viewer/v1/token", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Authorization", "Bearer "+w.token)
	req.Header.Set("Origin", "https://kick.com")
	req.Header.Set("Referer", fmt.Sprintf("https://kick.com/%s/", w.streamer))
	req.Header.Set("X-Chatroom", fmt.Sprintf("%d", w.channelID))
	req.Header.Set("X-User-Id", fmt.Sprintf("%d", w.userID))
	req.Header.Set("X-Client-Token", "e1393935a959b4020a4491574f6490129f678acdaa92760471263db43487f823")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var parsed struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
		Token string `json:"token"`
	}
	_ = json.Unmarshal(body, &parsed)

	if parsed.Data.Token != "" {
		return parsed.Data.Token, nil
	}
	return parsed.Token, nil
}
