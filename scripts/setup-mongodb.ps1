# Script de Setup do MongoDB - L2Raptors Go

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  MongoDB Setup - L2Raptors Go" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$mongoScript = @"
use l2raptors

// Criar indices para accounts
db.accounts.createIndex({ "login": 1 }, { unique: true })
db.accounts.createIndex({ "last_ip": 1 })
db.accounts.createIndex({ "banned_until": 1 })

// Criar indices para characters (para o futuro)
db.characters.createIndex({ "account_id": 1 })
db.characters.createIndex({ "char_name": 1 }, { unique: true })

// Verificar indices criados
print("Indices de accounts:")
db.accounts.getIndexes()

print("\nIndices de characters:")
db.characters.getIndexes()

print("\nMongoDB configurado com sucesso!")
"@

# Salvar script temporario
$mongoScript | Out-File -FilePath "temp_mongo_setup.js" -Encoding UTF8

Write-Host "Executando configuracao do MongoDB..." -ForegroundColor Yellow

# Executar script no MongoDB
mongosh --quiet --file temp_mongo_setup.js

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "  MongoDB Configurado com Sucesso!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Cyan
} else {
    Write-Host ""
    Write-Host "ERRO: Falha ao configurar MongoDB" -ForegroundColor Red
}

# Remover arquivo temporario
Remove-Item "temp_mongo_setup.js" -ErrorAction SilentlyContinue

Write-Host ""
