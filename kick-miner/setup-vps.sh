#!/bin/bash
# Script de instalacao e inicio do Kick Channel Points Miner no VPS Linux

echo "=== Configurando Kick Channel Points Miner ==="

if command -v docker &> /dev/null; then
    echo "[+] Docker detectado. Iniciando via Docker Compose..."
    docker compose up -d || docker-compose up -d
    echo "[✓] Minerador rodando em background no Docker!"
    echo "Painel Web disponivel em: http://IP_DO_SEU_VPS:5000"
else
    echo "[!] Docker nao encontrado. Configurando via Python..."
    if ! command -v git &> /dev/null; then
        sudo apt update && sudo apt install -y git python3 python3-pip python3-venv
    fi

    if [ ! -d "app" ]; then
        git clone https://github.com/Baillora/Kick_Channel_Points_Miner.git app
    fi

    cd app
    python3 -m venv venv
    source venv/bin/activate
    pip install -r requirements.txt
    cp ../config.json config.json

    echo "[+] Iniciando minerador em background via nohup..."
    nohup python3 main.py > ../miner.log 2>&1 &
    echo "[✓] Minerador rodando via Python em background!"
fi
