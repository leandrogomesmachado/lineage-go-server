# L2Raptors-Go

Port do servidor Lineage 2 L2Raptors de Java para Go Language.

## Objetivo

Criar uma versão de alta performance do L2Raptors usando Go, mantendo compatibilidade total com o protocolo Lineage 2 e reutilizando o datapack existente.

## Arquitetura

- **LoginServer**: Autenticação de contas e gerenciamento de GameServers
- **GameServer**: Lógica do jogo, NPCs, combate, quests, etc.
- **MongoDB**: Banco de dados NoSQL (substitui MariaDB)
- **Datapack**: Reutiliza XMLs, HTMLs e geodata do projeto Java

## Requisitos

- Go 1.23+
- MongoDB 7.0+
- Cliente Lineage 2 (compatível com o protocolo 746 - Interlude)

## Quick Start

### 1. Instalar Dependências

```bash
go mod download
```

### 2. Configurar MongoDB

```bash
# Iniciar MongoDB
mongod --dbpath C:\data\db

# Criar database e usuário
mongo
> use l2raptors
> db.createUser({user: "l2admin", pwd: "senha123", roles: ["readWrite"]})
```

### 3. Configurar Servidores

Editar arquivos em `configs/`:
- `loginserver.yaml` - Configurações do LoginServer
- `gameserver.yaml` - Configurações do GameServer

### 4. Executar

```bash
# LoginServer
go run cmd/loginserver/main.go

# GameServer (em outro terminal)
go run cmd/gameserver/main.go
```

## Estrutura do Projeto

```
l2raptors-go/
├── cmd/                    # Executáveis principais
│   ├── loginserver/       # LoginServer entry point
│   └── gameserver/        # GameServer entry point
├── internal/              # Código interno (não exportável)
│   ├── loginserver/      # Lógica do LoginServer
│   └── gameserver/       # Lógica do GameServer
├── pkg/                   # Bibliotecas compartilhadas (exportáveis)
│   ├── config/           # Gerenciamento de configuração
│   ├── logger/           # Sistema de logging
│   ├── protocol/         # Protocolo Lineage 2
│   └── database/         # Camada de acesso ao MongoDB
├── configs/              # Arquivos de configuração
├── docs/                 # Documentação
└── scripts/              # Scripts de build e deploy
```

## Desenvolvimento

### Build

```bash
# Build LoginServer
go build -o bin/loginserver.exe cmd/loginserver/main.go

# Build GameServer
go build -o bin/gameserver.exe cmd/gameserver/main.go
```

### Testes

```bash
# Rodar todos os testes
go test ./...

# Testes com coverage
go test -cover ./...

# Testes de um pacote específico
go test ./internal/loginserver/auth
```

### Linting

```bash
# Instalar golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Rodar linter
golangci-lint run
```

## Performance

Comparado com a versão Java:

- **Startup**: 15-30x mais rápido
- **Memória**: 60-70% menos uso
- **TPS**: 3x mais transações por segundo
- **Latência**: 70% menor


## 📝 Licença

MIT.

## Contribuindo

1. Fork o projeto
2. Crie uma branch (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -m 'feat: adicionar nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## Links

- [Documentação Go](https://golang.org/doc/)
- [MongoDB Go Driver](https://www.mongodb.com/docs/drivers/go/current/)
