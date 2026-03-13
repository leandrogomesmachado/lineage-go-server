# Script para executar GameServer

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  L2Raptors GameServer" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Verificar se o executavel existe
if (-not (Test-Path "bin/gameserver.exe")) {
    Write-Host "ERRO: GameServer nao encontrado" -ForegroundColor Red
    Write-Host "Execute: .\scripts\build.ps1" -ForegroundColor Yellow
    exit 1
}

# Criar diretorio de logs se nao existir
if (-not (Test-Path "logs")) {
    New-Item -ItemType Directory -Path "logs" | Out-Null
}

# Executar GameServer
.\bin\gameserver.exe
