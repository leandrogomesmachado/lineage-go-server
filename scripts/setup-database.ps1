# Script para configurar banco de dados MongoDB do L2Raptors
# Executa todos os scripts de criação de collections

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  L2Raptors - Setup MongoDB" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$mongoScripts = @(
    "mongodb\01-create-accounts.js",
    "mongodb\02-create-characters.js",
    "mongodb\03-create-gameservers.js"
)

$mongoUri = "mongodb://localhost:27017"
$database = "l2raptors"

Write-Host "Conectando ao MongoDB em $mongoUri..." -ForegroundColor Yellow

foreach ($script in $mongoScripts) {
    $scriptPath = Join-Path $PSScriptRoot $script
    
    if (Test-Path $scriptPath) {
        Write-Host ""
        Write-Host "Executando: $script" -ForegroundColor Green
        
        try {
            mongosh $mongoUri --quiet --file $scriptPath
            
            if ($LASTEXITCODE -eq 0) {
                Write-Host "  OK: $script executado com sucesso" -ForegroundColor Green
            } else {
                Write-Host "  ERRO: Falha ao executar $script" -ForegroundColor Red
                exit 1
            }
        }
        catch {
            Write-Host "  ERRO: $($_.Exception.Message)" -ForegroundColor Red
            exit 1
        }
    }
    else {
        Write-Host "  AVISO: Script nao encontrado: $scriptPath" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Setup Concluido com Sucesso!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Collections criadas:" -ForegroundColor Green
Write-Host "  - accounts (LoginServer)" -ForegroundColor White
Write-Host "  - characters (GameServer)" -ForegroundColor White
Write-Host "  - gameservers (LoginServer)" -ForegroundColor White
Write-Host ""
Write-Host "Banco de dados: $database" -ForegroundColor Yellow
Write-Host "URI: $mongoUri" -ForegroundColor Yellow
