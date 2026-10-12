package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type SessionResult struct {
	Nick  string `json:"nick"`
	Token string `json:"token"`
}

type CDPTarget struct {
	ID                   string `json:"id"`
	Title                string `json:"title"`
	Type                 string `json:"type"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

type CDPMessage struct {
	ID     int64          `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

type CDPResponse struct {
	ID     int64 `json:"id"`
	Result struct {
		Result struct {
			Type  string `json:"type"`
			Value any    `json:"value"`
		} `json:"result"`
	} `json:"result"`
	Error any `json:"error,omitempty"`
}

// LoginPokeIdleViaBrowserCDP abre temporariamente o navegador local para autenticar no pokeidle.io e extrair o token
func LoginPokeIdleViaBrowserCDP(usernameOrEmail, password string) (*SessionResult, error) {
	browserPath, err := FindBrowserPath()
	if err != nil {
		return nil, err
	}

	port := 9222
	userDataDir, err := os.MkdirTemp("", "pokeidle_cdp_login_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(userDataDir)

	args := []string{
		fmt.Sprintf("--remote-debugging-port=%d", port),
		fmt.Sprintf("--user-data-dir=%s", userDataDir),
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-blink-features=AutomationControlled",
		"--window-size=1100,820",
		"https://pokeidle.io/app",
	}

	if runtime.GOOS != "windows" {
		args = append(args, "--headless=new", "--no-sandbox", "--disable-gpu")
	}

	log.Printf("[CDP Login] Iniciando navegador (%s) em https://pokeidle.io/app ...", browserPath)
	cmd := exec.Command(browserPath, args...)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("falha ao iniciar navegador: %w", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	// Aguardar endpoint CDP ficar pronto
	cdpURL := fmt.Sprintf("http://127.0.0.1:%d/json/list", port)
	var wsURL string

	for i := 0; i < 35; i++ {
		time.Sleep(300 * time.Millisecond)
		resp, err := http.Get(cdpURL)
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			var targets []CDPTarget
			if json.Unmarshal(body, &targets) == nil {
				for _, t := range targets {
					if t.Type == "page" && t.WebSocketDebuggerURL != "" && (strings.Contains(t.URL, "pokeidle.io") || strings.Contains(t.Title, "Pokeidle")) {
						wsURL = t.WebSocketDebuggerURL
						break
					}
				}
				if wsURL == "" {
					for _, t := range targets {
						if t.Type == "page" && t.WebSocketDebuggerURL != "" {
							wsURL = t.WebSocketDebuggerURL
							break
						}
					}
				}
				if wsURL != "" {
					break
				}
			}
		}
	}

	if wsURL == "" {
		return nil, fmt.Errorf("timeout aguardando WebSocket do navegador")
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar no CDP: %w", err)
	}
	defer conn.Close()

	log.Println("[CDP Login] Conectado ao PokeIdle. Aguardando carregamento e formulário de login...")

	var reqID int64 = 1

	fillScript := fmt.Sprintf(`(function() {
		let idInput = document.getElementById('login-id');
		let passInput = document.getElementById('login-senha');
		if (!idInput || !passInput) return false;

		if (idInput.value !== %q) {
			idInput.value = %q;
			idInput.dispatchEvent(new Event('input', { bubbles: true }));
			idInput.dispatchEvent(new Event('change', { bubbles: true }));
		}

		if (passInput.value !== %q) {
			passInput.value = %q;
			passInput.dispatchEvent(new Event('input', { bubbles: true }));
			passInput.dispatchEvent(new Event('change', { bubbles: true }));
		}

		// Se o Cloudflare Turnstile carregou antes do app.js, força o render
		if (window.turnstile) {
			if (typeof window.onTurnstilePronto === 'function') {
				try { window.onTurnstilePronto(); } catch(e) {}
			}
			let wrap = document.getElementById('turnstile-entrar');
			if (wrap && !wrap.hasChildNodes()) {
				try {
					window.turnstile.render(wrap, { sitekey: '0x4AAAAAAEccJgXg19_AZoDr', size: 'flexible' });
				} catch(e) {}
			}
		}

		return true;
	})()`, usernameOrEmail, usernameOrEmail, password, password)

	checkSessionScript := `(function() {
		let raw = localStorage.getItem('sessao');
		return raw || "";
	})()`

	timeout := time.After(90 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	filled := false
	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("tempo limite excedido aguardando login no pokeidle.io")
		case <-ticker.C:
			// 1. Verifica se já está logado ou se sessão já foi gerada
			val := evalCDP(conn, &reqID, checkSessionScript)
			if val != "" && strings.Contains(val, "token") {
				var res SessionResult
				if err := json.Unmarshal([]byte(val), &res); err == nil && res.Token != "" {
					log.Printf("[CDP Login] 🎉 Sessão capturada com sucesso! Nick: %s", res.Nick)
					return &res, nil
				}
			}

			// 2. Preenche os campos quando estiverem disponíveis
			if !filled {
				resFill := evalCDP(conn, &reqID, fillScript)
				if resFill == "true" {
					log.Println("[CDP Login] Credenciais preenchidas. Resolvendo Turnstile...")
					filled = true
				}
			} else {
				// Tenta obter coordenadas do Turnstile e clicar nele automaticamente
				coordsStr := evalCDP(conn, &reqID, `(function() {
					let iframe = document.querySelector('#turnstile-entrar iframe');
					if (iframe) {
						let r = iframe.getBoundingClientRect();
						return JSON.stringify({ x: r.left + 35, y: r.top + 32, visible: r.width > 0 });
					}
					let wrap = document.getElementById('turnstile-entrar');
					if (wrap) {
						let r = wrap.getBoundingClientRect();
						return JSON.stringify({ x: r.left + 35, y: r.top + 32, visible: r.width > 0 });
					}
					return "";
				})()`)

				if coordsStr != "" {
					var coords struct {
						X       float64 `json:"x"`
						Y       float64 `json:"y"`
						Visible bool    `json:"visible"`
					}
					if json.Unmarshal([]byte(coordsStr), &coords) == nil && coords.Visible && coords.X > 0 && coords.Y > 0 {
						dispatchClick(conn, &reqID, coords.X, coords.Y)
					}
				}

				diagStr := evalCDP(conn, &reqID, `(function() {
					let err = document.getElementById('login-erro')?.textContent || "";
					let btn = document.getElementById('entrar');
					let token = document.querySelector('input[name="cf-turnstile-response"]')?.value || "";
					return JSON.stringify({ err, btnDisabled: btn ? btn.disabled : true, hasToken: token.length > 0 });
				})()`)
				log.Printf("[CDP Login Status] %s (Coords: %s)", diagStr, coordsStr)

				// Só clica no botão se o Turnstile já tiver gerado o token de verificação
				evalCDP(conn, &reqID, `(function() {
					let token = document.querySelector('input[name="cf-turnstile-response"]')?.value || "";
					let btn = document.getElementById('entrar');
					if (token.length > 0 && btn && !btn.disabled) {
						btn.click();
						return true;
					}
					return false;
				})()`)
			}
		}
	}
}

func evalCDP(conn *websocket.Conn, reqID *int64, expr string) string {
	*reqID++
	id := *reqID
	msg := CDPMessage{
		ID:     id,
		Method: "Runtime.evaluate",
		Params: map[string]any{
			"expression":    expr,
			"returnByValue": true,
		},
	}
	_ = conn.WriteJSON(msg)

	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		_, respData, err := conn.ReadMessage()
		if err != nil {
			return ""
		}
		var resp CDPResponse
		if json.Unmarshal(respData, &resp) == nil && resp.ID == id {
			if strVal, ok := resp.Result.Result.Value.(string); ok {
				return strVal
			}
			return fmt.Sprintf("%v", resp.Result.Result.Value)
		}
	}
}

func dispatchClick(conn *websocket.Conn, reqID *int64, x, y float64) {
	*reqID++
	_ = conn.WriteJSON(map[string]any{
		"id":     *reqID,
		"method": "Input.dispatchMouseEvent",
		"params": map[string]any{
			"type":       "mousePressed",
			"x":          x,
			"y":          y,
			"button":     "left",
			"clickCount": 1,
		},
	})
	time.Sleep(60 * time.Millisecond)
	*reqID++
	_ = conn.WriteJSON(map[string]any{
		"id":     *reqID,
		"method": "Input.dispatchMouseEvent",
		"params": map[string]any{
			"type":       "mouseReleased",
			"x":          x,
			"y":          y,
			"button":     "left",
			"clickCount": 1,
		},
	})
}
