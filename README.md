# Sistema de Análise com RouteLLM

Este sistema implementa um orquestrador que utiliza o RouteLLM para quebrar tarefas em subtarefas e distribuí-las entre agents especializados.

## Estrutura do Sistema

```
HiveMind/
├── agents/
│   └── llm_agent.go       # Implementação dos agents
├── config/
│   └── rabbitmq.go        # Configuração do RabbitMQ
├── orchestrator/
│   ├── llm_router.go      # Implementação do router
│   └── task_types.go      # Definição dos tipos de tarefas
└── cmd/
    └── main.go            # Arquivo principal
```

## Componentes

1. **LLMRouter**:
   - Recebe tarefas via RabbitMQ
   - Usa RouteLLM para quebrar em subtarefas
   - Distribui subtarefas para os agents

2. **LLMAgents**:
   - 5 tipos diferentes de agents
   - 2 instâncias de cada tipo (10 total)
   - Processamento assíncrono
   - Especialização por tipo de tarefa

3. **Filas RabbitMQ**:
   - `llm_input`: Recebe tarefas principais
   - `llm_tasks`: Distribui subtarefas
   - `llm_results`: Coleta resultados

## Pré-requisitos

1. Go 1.21 ou superior
2. RabbitMQ 3.x
3. Variáveis de ambiente configuradas

## Configuração

1. Clone o repositório
2. Copie `.env.example` para `.env`
3. Configure as variáveis do RabbitMQ:
   ```env
   RABBITMQ_HOST=localhost
   RABBITMQ_PORT=5672
   RABBITMQ_USER=guest
   RABBITMQ_PASSWORD=guest
   ```

## Instalação

```bash
# Instalar dependências
go mod tidy

# Compilar
go build -o llm_system cmd/main.go
```

## Uso

```bash
# Iniciar o sistema
./llm_system
```

## Tipos de Agents

1. **Analysis Agent**:
   - Análise de requisitos e contexto
   - Prioridade: Alta

2. **Research Agent**:
   - Pesquisa e coleta de informações
   - Prioridade: Alta

3. **Development Agent**:
   - Desenvolvimento da solução
   - Prioridade: Alta

4. **Validation Agent**:
   - Validação e testes
   - Prioridade: Média

5. **Documentation Agent**:
   - Documentação e relatórios
   - Prioridade: Média

## Exemplo de Tarefa

```json
{
  "id": "uuid",
  "description": "Analisar o repositório RouteLLM",
  "parameters": {
    "repository": "https://github.com/lm-sys/RouteLLM",
    "priority": "high",
    "context": "Análise técnica e funcional"
  }
}
```

## Monitoramento

O sistema usa logs com emojis para melhor visualização:
- 🚀 Início de operações
- 🤖 Atividade dos agents
- 📥 Recebimento de tarefas
- 🔄 Processamento
- ✅ Conclusão
- ❌ Erros

## Graceful Shutdown

O sistema suporta graceful shutdown com SIGINT/SIGTERM:
1. Cancela o contexto principal
2. Aguarda conclusão das tarefas em andamento
3. Fecha conexões com RabbitMQ
4. Encerra os agents ordenadamente

## Extensões Possíveis

1. Implementar integração real com RouteLLM
2. Adicionar persistência de dados
3. Implementar retry policies
4. Adicionar métricas e monitoramento
5. Implementar balanceamento de carga
6. Adicionar testes automatizados