# Script de Build - L2Raptors Go

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  L2Raptors Go - Build Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$ErrorActionPreference = "Stop"

# Criar diretorio bin
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

# Criar diretorio logs
if (-not (Test-Path "logs")) {
    New-Item -ItemType Directory -Path "logs" | Out-Null
}

Write-Host "Building LoginServer..." -ForegroundColor Yellow
go build -o bin/loginserver.exe cmd/loginserver/main.go
if ($LASTEXITCODE -eq 0) {
    Write-Host "  OK: LoginServer compilado" -ForegroundColor Green
} else {
    Write-Host "  ERRO: Falha ao compilar LoginServer" -ForegroundColor Red
    exit 1
}

Write-Host "Building GameServer..." -ForegroundColor Yellow
go build -o bin/gameserver.exe cmd/gameserver/main.go
if ($LASTEXITCODE -eq 0) {
    Write-Host "  OK: GameServer compilado" -ForegroundColor Green
} else {
    Write-Host "  ERRO: Falha ao compilar GameServer" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Build Concluido com Sucesso!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Executaveis criados em:" -ForegroundColor Yellow
Write-Host "  bin/loginserver.exe" -ForegroundColor White
Write-Host "  bin/gameserver.exe" -ForegroundColor White
Write-Host ""
