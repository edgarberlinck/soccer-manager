#!/bin/bash

# Script de setup do sistema de bots

set -e

echo "🤖 Setup do Sistema de Bots"
echo "============================"
echo ""

# 1. Rodar migrations
echo "1️⃣ Rodando migrations..."
cd api
if command -v goose &> /dev/null; then
    echo "Usando goose para migrations..."
    # Usuário deve configurar a connection string
    echo "⚠️  Configure sua DATABASE_URL no .env antes de rodar as migrations"
    echo "   Exemplo: goose -dir internal/infrastructure/database/migrations postgres \$DATABASE_URL up"
else
    echo "⚠️  Goose não encontrado. Instale com: go install github.com/pressly/goose/v3/cmd/goose@latest"
fi
echo ""

# 2. Gerar código sqlc
echo "2️⃣ Gerando código sqlc..."
if command -v sqlc &> /dev/null; then
    sqlc generate
    echo "✅ Código gerado com sucesso!"
else
    echo "⚠️  sqlc não encontrado. Instale com: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"
fi
echo ""

# 3. Instalar dependências Go
echo "3️⃣ Instalando dependências Go..."
go mod download
go mod tidy
echo "✅ Dependências instaladas!"
echo ""

# 4. Compilar comandos
echo "4️⃣ Compilando comandos..."
echo "   - create-bots"
go build -o bin/create-bots cmd/create-bots/main.go
echo "   - bot-scheduler"
go build -o bin/bot-scheduler cmd/bot-scheduler/main.go
echo "✅ Comandos compilados em ./bin/"
echo ""

echo "✅ Setup concluído!"
echo ""
echo "📋 Próximos passos:"
echo "   1. Configure o arquivo .env com DATABASE_URL"
echo "   2. Rode as migrations: goose -dir internal/infrastructure/database/migrations postgres \$DATABASE_URL up"
echo "   3. Crie os bots: ./bin/create-bots"
echo "   4. Inicie o scheduler: ./bin/bot-scheduler"
echo ""
echo "📖 Consulte Bots.md para mais detalhes"
