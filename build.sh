#!/bin/bash

# Criar diretório bin se não existir
mkdir -p bin

echo "🔨 Compilando os consumidores..."
go build -o bin/consumer cmd/consume/main.go

echo "🔨 Compilando o criador de capítulo..."
go build -o bin/create_chapter cmd/create_chapter/main.go

echo "✅ Compilação concluída!"
echo ""
echo "Para executar:"
echo "1. Configure a chave da API Groq:"
echo "   export GROQ_API_KEY=\"sua-chave-aqui\""
echo ""
echo "2. Em um terminal, inicie os consumidores:"
echo "   ./bin/consumer"
echo ""
echo "3. Em outro terminal, crie um capítulo:"
echo "   ./bin/create_chapter" 