# Script para executar LoginServer

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  L2Raptors LoginServer" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Verificar se o executavel existe
if (-not (Test-Path "bin/loginserver.exe")) {
    Write-Host "ERRO: LoginServer nao encontrado" -ForegroundColor Red
    Write-Host "Execute: .\scripts\build.ps1" -ForegroundColor Yellow
    exit 1
}

# Criar diretorio de logs se nao existir
if (-not (Test-Path "logs")) {
    New-Item -ItemType Directory -Path "logs" | Out-Null
}

# Executar LoginServer
.\bin\loginserver.exe
