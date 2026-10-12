package twitch

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Enabled    bool   `json:"enabled"`
	Username   string `json:"username"`
	OAuthToken string `json:"oauthToken"` // Formato: "oauth:xxxxxxxxxx" ou apenas o token
}

type Client struct {
	cfg       Config
	mu        sync.RWMutex
	conn      net.Conn
	reader    *bufio.Reader
	writer    *bufio.Writer
	channels  map[string]bool
	activeChs map[string]bool
	running   bool
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg:       cfg,
		channels:  make(map[string]bool),
		activeChs: make(map[string]bool),
	}
}

// Start inicia o loop de conexão ao chat IRC da Twitch
func (c *Client) Start(ctx context.Context) {
	if !c.cfg.Enabled || c.cfg.Username == "" || c.cfg.OAuthToken == "" {
		return
	}

	token := c.cfg.OAuthToken
	if !strings.HasPrefix(token, "oauth:") {
		token = "oauth:" + token
	}

	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}
	c.running = true
	c.mu.Unlock()

	go func() {
		log.Printf("[Twitch Lurker] Iniciando worker de chat IRC para o usuário '%s'...", c.cfg.Username)
		for {
			select {
			case <-ctx.Done():
				c.Close()
				return
			default:
				if err := c.connectAndListen(ctx, token); err != nil {
					log.Printf("[Twitch Lurker] Conexão IRC encerrada (%v). Reconectando em 15s...", err)
					time.Sleep(15 * time.Second)
				}
			}
		}
	}()
}

func (c *Client) connectAndListen(ctx context.Context, token string) error {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", "irc.chat.twitch.tv:6697", nil)
	if err != nil {
		return fmt.Errorf("erro ao conectar no irc.chat.twitch.tv: %w", err)
	}
	defer conn.Close()

	c.mu.Lock()
	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.writer = bufio.NewWriter(conn)
	c.activeChs = make(map[string]bool)
	c.mu.Unlock()

	// Handshake IRC com a Twitch
	if err := c.sendRaw(fmt.Sprintf("PASS %s", token)); err != nil {
		return err
	}
	if err := c.sendRaw(fmt.Sprintf("NICK %s", strings.ToLower(c.cfg.Username))); err != nil {
		return err
	}

	log.Printf("[Twitch Lurker] Autenticado com sucesso na Twitch como '%s'!", c.cfg.Username)

	// Entra nos canais configurados
	c.mu.RLock()
	targetChannels := make([]string, 0, len(c.channels))
	for ch := range c.channels {
		targetChannels = append(targetChannels, ch)
	}
	c.mu.RUnlock()

	// Por padrão, sempre entra no canal oficial faasii se não houver canais definidos
	if len(targetChannels) == 0 {
		targetChannels = append(targetChannels, "faasii")
	}

	for _, ch := range targetChannels {
		_ = c.joinChannel(ch)
	}

	// Loop de leitura de mensagens e keepalive (PING/PONG)
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "PING ") {
			pongMsg := "PONG " + line[5:]
			_ = c.sendRaw(pongMsg)
		}
	}
}

func (c *Client) SetTargetChannels(channels []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	newMap := make(map[string]bool)
	for _, ch := range channels {
		clean := strings.ToLower(strings.TrimPrefix(ch, "#"))
		if clean != "" {
			newMap[clean] = true
			if !c.activeChs[clean] && c.conn != nil {
				go func(canal string) {
					_ = c.joinChannel(canal)
				}(clean)
			}
		}
	}
	c.channels = newMap
}

func (c *Client) joinChannel(ch string) error {
	ch = strings.ToLower(strings.TrimPrefix(ch, "#"))
	err := c.sendRaw(fmt.Sprintf("JOIN #%s", ch))
	if err == nil {
		c.mu.Lock()
		c.activeChs[ch] = true
		c.mu.Unlock()
		log.Printf("[Twitch Lurker] Entrou no chat do canal oficial: #%s (+15%% XP ativo)", ch)
	}
	return err
}

func (c *Client) sendRaw(cmd string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil || c.writer == nil {
		return fmt.Errorf("conexão não estabelecida")
	}

	_, err := c.writer.WriteString(cmd + "\r\n")
	if err != nil {
		return err
	}
	return c.writer.Flush()
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.running = false
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}
