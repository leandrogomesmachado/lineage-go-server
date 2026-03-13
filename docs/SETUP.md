# Setup L2Raptors-Go

Guia completo de instalação e configuração do L2Raptors-Go.

##  Pré-requisitos

### 1. Go Language
```powershell
# Instalar Go 1.23+
winget install GoLang.Go

# Verificar instalação
go version
```

### 2. MongoDB
```powershell
# Instalar MongoDB
winget install MongoDB.Server

# Criar diretório de dados
mkdir C:\data\db

# Iniciar MongoDB
mongod --dbpath C:\data\db
```

### 3. Cliente Lineage 2
- Cliente compatível com o protocolo do servidor
- Configurar para conectar em `127.0.0.1:2106` (LoginServer)

##  Instalação

### 1. Clonar o Repositório
```bash
cd C:\dev\l2raptors
git clone https://github.com/leandrogomesmachado/l2raptors-go.git
cd l2raptors-go
```

### 2. Instalar Dependências
```bash
go mod download
```

### 3. Configurar MongoDB

Abrir MongoDB shell:
```bash
mongo
```

Executar:
```javascript
use l2raptors

// Criar usuário admin
db.createUser({
  user: "l2admin",
  pwd: "senha123",
  roles: ["readWrite"]
})

// Criar índices
db.accounts.createIndex({ "login": 1 }, { unique: true })
db.characters.createIndex({ "account_id": 1 })
db.characters.createIndex({ "char_name": 1 }, { unique: true })

exit
```

### 4. Configurar Servidores

Editar `configs/loginserver.yaml`:
```yaml
database:
  uri: "mongodb://l2admin:senha123@localhost:27017"
```

Editar `configs/gameserver.yaml`:
```yaml
database:
  uri: "mongodb://l2admin:senha123@localhost:27017"
  
datapack:
  path: "../l2raptors-java/raptors_datapack"
```

## Build

```powershell
# Compilar ambos os servidores
.\scripts\build.ps1
```

Isso criará:
- `bin/loginserver.exe`
- `bin/gameserver.exe`

## Executar

### Opção 1: Scripts PowerShell (Recomendado)

```powershell
# Terminal 1 - LoginServer
.\scripts\run-loginserver.ps1

# Terminal 2 - GameServer
.\scripts\run-gameserver.ps1
```

### Opção 2: Executáveis Diretos

```powershell
# LoginServer
.\bin\loginserver.exe

# GameServer
.\bin\gameserver.exe
```

### Opção 3: Go Run (Desenvolvimento)

```powershell
# LoginServer
go run cmd/loginserver/main.go

# GameServer
go run cmd/gameserver/main.go
```

## Testar

### 1. Verificar LoginServer

O LoginServer deve exibir:
```
========================================
  L2Raptors LoginServer - Go Edition
========================================

MongoDB conectado com sucesso
LoginServer iniciado em 0.0.0.0:2106
Aguardando conexoes...
```

### 2. Verificar GameServer

O GameServer deve exibir:
```
========================================
  L2Raptors GameServer - Go Edition
========================================

Server ID: 1
Server Name: Raptors
Max Players: 1000

MongoDB conectado com sucesso
Datapack path: ../l2raptors-java/raptors_datapack
GameServer pronto para conexoes!
```

### 3. Conectar com Cliente

1. Abrir cliente Lineage 2
2. Configurar para conectar em `127.0.0.1`
3. Criar conta (auto-criação habilitada por padrão)
4. Login deve ser bem-sucedido

## Desenvolvimento

### Estrutura de Diretórios

```
l2raptors-go/
├── cmd/                    # Executáveis
│   ├── loginserver/       # LoginServer main
│   └── gameserver/        # GameServer main
├── internal/              # Código interno
│   ├── loginserver/      # Lógica LoginServer
│   └── gameserver/       # Lógica GameServer
├── pkg/                   # Bibliotecas compartilhadas
│   ├── config/           # Configuração
│   ├── logger/           # Logging
│   ├── protocol/         # Protocolo L2
│   └── database/         # MongoDB
├── configs/              # Arquivos YAML
├── scripts/              # Scripts de build/run
└── docs/                 # Documentação
```

### Adicionar Nova Funcionalidade

1. Criar código em `internal/`
2. Adicionar testes em `*_test.go`
3. Atualizar documentação
4. Commit e push

### Rodar Testes

```bash
# Todos os testes
go test ./...

# Com coverage
go test -cover ./...

# Teste específico
go test ./internal/loginserver/auth
```

## Troubleshooting

### MongoDB não conecta

```bash
# Verificar se MongoDB está rodando
tasklist | findstr mongod

# Iniciar MongoDB manualmente
mongod --dbpath C:\data\db
```

### Porta já em uso

```powershell
# Verificar porta 2106 (LoginServer)
netstat -ano | findstr :2106

# Verificar porta 7777 (GameServer)
netstat -ano | findstr :7777

# Matar processo
taskkill /PID <PID> /F
```

### Erro de import

```bash
# Limpar cache do Go
go clean -modcache

# Reinstalar dependências
go mod download
```

## Monitoramento

### Logs

Logs são salvos em `logs/`:
- `logs/loginserver.log` - Logs do LoginServer
- `logs/gameserver.log` - Logs do GameServer

### MongoDB

```bash
# Conectar ao MongoDB
mongo

# Ver estatísticas
use l2raptors
db.stats()

# Ver contas
db.accounts.find().pretty()

# Ver personagens
db.characters.find().pretty()
```

## Segurança

### Produção

Para produção, alterar:

1. **MongoDB**: Usar autenticação forte
2. **Firewall**: Abrir apenas portas necessárias
3. **Logs**: Configurar rotação de logs
4. **Configs**: Não commitar senhas

### Exemplo de configuração segura:

```yaml
database:
  uri: "mongodb://user:strong_password@localhost:27017/?authSource=admin"
  
security:
  auto_create_accounts: false  # Desabilitar em produção
  max_login_attempts: 3
  ban_duration_minutes: 60
```

## Links Úteis

- [Documentação Go](https://golang.org/doc/)
- [MongoDB Go Driver](https://www.mongodb.com/docs/drivers/go/current/)

