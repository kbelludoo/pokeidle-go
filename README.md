# ⚡ Pokéidle Go — Autonomous Go Bot, Twitch & Kick Lurker

Bot autônomo de altíssima performance desenvolvido em **Go (Golang)** para o jogo web idle **[Pokéidle](https://pokeidle.io/app)**, equipado com lurker nativo de **Twitch IRC** (+15% a +20% EXP) e **Kick WebSocket Points Miner** (farm contínuo de Pontos de Canal 24/7 sem abrir navegadores).

🌐 **Command Center & Live Dashboard (GitHub Pages):**  
👉 **[https://kbelludoo.github.io/pokeidle-go/](https://kbelludoo.github.io/pokeidle-go/)**

---

## 🌟 Destaques e Arquitetura

- 🪶 **Ultra-leve (Consumo < 3 MB RAM):** Enquanto soluções em Python/Docker demandam de 45 a 100 MB de RAM e navegadores passam de 1 GB, o executável Go compila tudo de forma estática e roda com apenas **~2 a 3 MB de heap**.
- 🎮 **Automação Completa Pokéidle:**
  - Caça contínua na melhor área disponível.
  - Cura automática no Centro Pokémon quando a vida cai.
  - Auto-loot e venda automática de itens excedentes.
  - Otimização do time de combate (mantém 5 slots cheios para zerar a penalidade de dano de boss).
  - Renovação automática de sessão JWT (`sessao.renovada`).
- 🟣 **Twitch IRC Lurker Nativo:**
  - Conecta-se diretamente aos servidores IRC da Twitch (`irc.chat.twitch.tv:6667`).
  - Entra nas salas dos canais parceiros oficiais online (`#tftoddy`, `#faasii`, `#axtlol`, `#xdaren_`, `#cogu`, etc.).
  - Mantém o bônus de **+15% a +20% de XP** ativo 24 horas por dia com zero navegadores.
- 🟢 **Kick WebSocket Viewer Lurker Nativo:**
  - Engenharia reversa completa do protocolo de visualização da Kick (`websockets.kick.com`).
  - Envio contínuo de `channel_handshake` e `tracking.user.watch.livestream` com ID de canal e live.
  - Credita **+10 Pontos de Canal a cada 5 minutos** automaticamente.
  - Monitoramento de saldo em tempo real via API `https://kick.com/api/v2/channels/{canal}/points`.
- 📊 **Dashboard Web & GitHub Pages:**
  - Servidor local HTTP embutido (`http://localhost:3006`) com métricas em tempo real.
  - Página estática no GitHub Pages com Command Center integrado da VPS.

---

## 📁 Estrutura do Projeto

```text
pokeidle-go/
├── accounts.example.json       # Modelo sanitizado de configuração
├── accounts.json               # Suas credenciais reais (ignorado pelo git via .gitignore)
├── main.go                     # Ponto de entrada do executável
├── go.mod / go.sum             # Dependências Go nativas
│
├── docs/                       # 🌐 GITHUB PAGES (COMMAND CENTER AO VIVO)
│   ├── index.html              # Dashboard Cyberpunk agregador
│   └── status.json             # Telemetria em tempo real
│
├── pkg/
│   ├── engine/                 # Motor supervisor, métricas e ciclo de vida
│   ├── game/                   # Protocolo WebSocket do Pokéidle e regras de combate
│   ├── twitch/                 # Lurker nativo Twitch IRC
│   ├── kick/                   # Minerador nativo Kick WebSocket
│   ├── network/                # Conexão WebSocket thread-safe e cliente HTTP
│   └── auth/                   # Captura e renovação de tokens
│
└── scripts/
    ├── build-linux.bat         # Compilação cruzada para Linux VPS
    └── bot.service             # Template systemd para rodar como daemon na VPS
```

---

## 🚀 Como Executar

### 1. Clonar e Configurar
```bash
git clone https://github.com/kbelludoo/pokeidle-go.git
cd pokeidle-go
cp accounts.example.json accounts.json
```

Edite o arquivo `accounts.json` com suas credenciais:
```json
[
  {
    "id": "acc-1",
    "name": "Seu Nick",
    "enabled": true,
    "credentials": {
      "token": "seu_token_pokeidle_aqui"
    },
    "settings": {
      "huntTarget": "",
      "autoHunt": true,
      "autoSell": true,
      "autoHeal": true,
      "autoDaily": true,
      "lure": 2
    },
    "twitch": {
      "enabled": true,
      "username": "seu_usuario_twitch",
      "oauthToken": "oauth:seu_oauth_token_twitch"
    },
    "kick": {
      "enabled": true,
      "bearerToken": "seu_bearer_token_kick",
      "streamers": ["tftoddy", "faasii", "axtlol", "xdaren_", "cogu"]
    }
  }
]
```

### 2. Rodar no Windows
```bash
go run .
# ou execute iniciar-bot.bat
```

### 3. Compilar para Linux VPS
No Windows:
```cmd
cmd /c "set GOOS=linux&& set GOARCH=amd64&& go build -ldflags=\"-s -w\" -o bot-poke-linux ."
```

Envie para a VPS:
```bash
scp bot-poke-linux ubuntu@sua-vps-ip:/home/ubuntu/pokeidle-go/
```

Execute na VPS:
```bash
chmod +x bot-poke-linux
./bot-poke-linux
```

---

## 🔒 Segurança

- O arquivo `accounts.json` e chaves privadas estão protegidos pelo `.gitignore` e **nunca** são enviados ao repositório público.
- O executável não realiza polling abusivo e respeita os tempos de handshake exigidos pelos servidores da Twitch, Kick e Pokéidle.
