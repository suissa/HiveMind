package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"HiveMind/agents"
	"HiveMind/agents/memory"
)

func main() {
	ctx := context.Background()

	// Configuração do gerenciador de memória
	memConfig := memory.DefaultMemoryConfig()
	memManager, err := memory.NewHybridMemoryManager(ctx, memConfig)
	if err != nil {
		log.Fatalf("Erro ao criar gerenciador de memória: %v", err)
	}
	defer memManager.Close(ctx)

	// Criação do agente cognitivo
	agent := agents.NewCognitiveAgent(
		"agent-001",
		"Agente de Pesquisa",
		"Agente especializado em pesquisa e análise de dados",
		10, // maxRounds
		"gpt-4",
		"pesquisador",
		"Analisar e sintetizar informações",
		memManager,
	)

	// Configuração do agente
	agent.Temperature = 0.7
	agent.MaxTokens = 2048
	agent.ContextWindow = 4096
	agent.SetBackstory("Sou um agente especializado em pesquisa e análise de dados, " +
		"com foco em encontrar padrões e insights relevantes.")

	// Adiciona alguns templates de prompt
	agent.AddPromptTemplate("pesquisa", "Analise os seguintes dados e identifique padrões: {{.dados}}")
	agent.AddPromptTemplate("resumo", "Resuma as principais descobertas da análise: {{.descobertas}}")

	// Adiciona conhecimento base
	agent.AddToKnowledgeBase("dominio", "análise de dados")
	agent.AddToKnowledgeBase("especialidade", "identificação de padrões")

	// Simula algumas tarefas e memorização
	tarefas := []struct {
		conteudo    map[string]interface{}
		importancia float64
		tags        []string
		isLongTerm  bool
	}{
		{
			conteudo: map[string]interface{}{
				"tipo":      "análise",
				"dados":     "Padrão de consumo de energia em diferentes horários",
				"resultado": "Pico de consumo entre 18h e 21h",
			},
			importancia: 0.8,
			tags:        []string{"energia", "análise", "padrões"},
			isLongTerm:  true,
		},
		{
			conteudo: map[string]interface{}{
				"tipo":      "observação",
				"dados":     "Temperatura ambiente vs consumo de energia",
				"resultado": "Correlação positiva forte",
			},
			importancia: 0.6,
			tags:        []string{"energia", "temperatura", "correlação"},
			isLongTerm:  false,
		},
		{
			conteudo: map[string]interface{}{
				"tipo":      "insight",
				"dados":     "Comportamento de usuários em horários de pico",
				"resultado": "Maior uso de aparelhos de alto consumo",
			},
			importancia: 0.9,
			tags:        []string{"comportamento", "usuários", "pico"},
			isLongTerm:  true,
		},
	}

	// Processa as tarefas
	for _, tarefa := range tarefas {
		err := agent.Memorize(ctx, tarefa.conteudo, tarefa.importancia, tarefa.tags, tarefa.isLongTerm)
		if err != nil {
			log.Printf("Erro ao memorizar: %v", err)
			continue
		}
		log.Printf("Memorizado: %v", tarefa.conteudo["tipo"])
	}

	// Simula treinamento
	config := agents.TrainingConfig{
		UseHistorical: true,
		BatchSize:     5,
		LearningRate:  0.001,
	}

	metrics, err := agent.Train(ctx, config)
	if err != nil {
		log.Printf("Erro no treinamento: %v", err)
	} else {
		log.Printf("Métricas de treinamento: Acurácia=%.2f, Loss=%.2f", metrics.Accuracy, metrics.Loss)
	}

	// Busca memórias relacionadas a energia
	memories, err := agent.Remember(ctx, []string{"energia"})
	if err != nil {
		log.Printf("Erro ao buscar memórias: %v", err)
	} else {
		log.Printf("Encontradas %d memórias relacionadas a energia", len(memories))
		for _, mem := range memories {
			log.Printf("Memória: %v (Importância: %.2f)", mem.Content["tipo"], mem.Importance)
		}
	}

	// Consolida memórias importantes
	if err := agent.ConsolidateMemories(ctx); err != nil {
		log.Printf("Erro ao consolidar memórias: %v", err)
	}

	// Remove memórias antigas/irrelevantes
	if err := agent.ForgetOldMemories(ctx); err != nil {
		log.Printf("Erro ao remover memórias antigas: %v", err)
	}

	// Exibe estatísticas de performance
	stats := agent.GetPerformanceStats()
	log.Printf("Estatísticas de performance:")
	for metric, value := range stats {
		log.Printf("- %s: %.2f", metric, value)
	}

	// Aguarda sinais de interrupção
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Monitora o agente
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Valida o estado do agente
			if err := agent.Validate(ctx); err != nil {
				log.Printf("Erro na validação do agente: %v", err)
			}

			// Exibe algumas estatísticas
			stats := agent.GetPerformanceStats()
			log.Printf("Taxa de sucesso atual: %.2f", stats["success_rate"])

		case sig := <-sigChan:
			log.Printf("Recebido sinal %v, encerrando...", sig)
			return
		}
	}
}
