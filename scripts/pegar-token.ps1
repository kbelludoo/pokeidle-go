param()

Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "     POKEIDLE.IO - CAPTURADOR / CONFIGURADOR DE TOKEN       " -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host ""

$root = Resolve-Path "$PSScriptRoot\.."
$accountsFile = Join-Path $root "accounts.json"

$token = ""
$nick = "kbelludoo"

try {
    $clip = [System.Windows.Forms.Clipboard]::GetText()
    if (-not $clip) {
        $clip = (Get-Clipboard -Raw)
    }
    if ($clip) {
        $clip = $clip.Trim()
        # Se for JSON completo {"nick":"...","token":"..."}
        if ($clip -like "*token*" -and $clip -like "*{*") {
            try {
                $obj = ConvertFrom-Json $clip
                if ($obj.token) {
                    $token = $obj.token
                    if ($obj.nick) { $nick = $obj.nick }
                    Write-Host "[!] Sessão detectada na área de transferência!" -ForegroundColor Green
                }
            } catch {}
        } elseif ($clip.Length -gt 20 -and -not ($clip -like "* *")) {
            Write-Host "[!] Token detectado na área de transferência!" -ForegroundColor Green
            $token = $clip
        }
    }
} catch {}

if (-not $token) {
    Write-Host "Como pegar sua sessão/token no navegador (leva 5 segundos):" -ForegroundColor Yellow
    Write-Host "  1. Abra o jogo logado: https://pokeidle.io/app" -ForegroundColor White
    Write-Host "  2. Pressione F12 -> Console" -ForegroundColor White
    Write-Host "  3. Cole o comando abaixo e aperte Enter:" -ForegroundColor White
    Write-Host '     copy(localStorage.getItem("sessao"))' -ForegroundColor Green
    Write-Host ""
    Write-Host "Ou se preferir apenas o token:" -ForegroundColor White
    Write-Host '     copy(JSON.parse(localStorage.getItem("sessao")).token)' -ForegroundColor Green
    Write-Host ""
    $inputToken = Read-Host "Cole aqui o conteúdo copiado ou o Token"
    if ($inputToken) {
        $inputToken = $inputToken.Trim()
        if ($inputToken -like "*token*" -and $inputToken -like "*{*") {
            try {
                $obj = ConvertFrom-Json $inputToken
                if ($obj.token) {
                    $token = $obj.token
                    if ($obj.nick) { $nick = $obj.nick }
                }
            } catch {}
        } else {
            $token = $inputToken
        }
    }
}

if (-not $token) {
    Write-Host ""
    Write-Host "[ERRO] Token não informado." -ForegroundColor Red
    return
}

# Salvar em accounts.json
try {
    $raw = @'
[
  {
    "id": "acc-1",
    "name": "Conta Principal",
    "enabled": true,
    "credentials": {
      "email": "kbelludoo@gmail.com",
      "password": "ECTbb,,0209",
      "sessionCookie": "",
      "token": ""
    },
    "settings": {
      "huntTarget": "pallet-town",
      "autoHunt": true,
      "autoSell": true,
      "autoHeal": true,
      "autoDaily": true,
      "lure": 2
    }
  }
]
'@
    if (Test-Path $accountsFile) {
        $content = Get-Content $accountsFile -Raw
        if ($content) { $raw = $content }
    }

    $json = ConvertFrom-Json $raw
    $json[0].credentials.token = $token
    if ($nick) { $json[0].name = $nick }
    $out = ConvertTo-Json $json -Depth 10
    Set-Content -Path $accountsFile -Value $out -Encoding UTF8
    Write-Host "[OK] Token salvo com sucesso em accounts.json para o treinador '$nick'!" -ForegroundColor Green
} catch {
    Write-Host "[ERRO] Erro ao salvar accounts.json: $_" -ForegroundColor Red
}

Write-Host ""
$start = Read-Host "Deseja iniciar o bot agora? (S/N) [Padrao: S]"
if ($start -eq '' -or $start -eq 'S' -or $start -eq 's') {
    Start-Process 'go' -ArgumentList 'run .' -WorkingDirectory $root
}
