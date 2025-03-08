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
	memConfig := &memory.MemoryConfig{
		RedisURL:            "redis://localhost:4567",
		MongoURL:            "mongodb://suissa:dc0b410b23dd26da2d423375437cceb4@195.35.19.148:27017/",
		MongoDB:             "HiveMind",
		Collection:          "memories",
		ShortTermTTL:        1 * time.Hour,
		ImportanceThreshold: 0.7,
		WeaviateURL:         "195.35.19.148:1111",
		WeaviateClass:       "Memory",
		WeaviateBatchSize:   100,
	}

	memManager, err := memory.NewHybridMemoryManager(ctx, memConfig)
	if err != nil {
		log.Fatalf("Erro ao criar gerenciador de memória: %v", err)
	}
	defer memManager.Close(ctx)

	// Cria uma equipe de treinamento
	crew := agents.NewTrainingCrew(memManager)

	// Registra listener para todos os eventos
	crew.OnAnyEvent(func(event agents.Event) {
		log.Printf("\n🔔 === NOVO EVENTO [%s] ===\n%s\n========================", event.Type, event.ToJSON())
	})

	// Configura os agentes
	trainingAgent := agents.NewCognitiveAgent(
		"training-agent",
		"Training Agent",
		"Agente responsável por criar e gerenciar treinamentos",
		100,
		"gpt-4",
		"training",
		"Criar experiências de treinamento envolventes e eficazes",
		memManager,
	)
	trainingAgent.SetBackstory("Sou um especialista em design instrucional e narrativas de aprendizado.")

	chapterAgent := agents.NewCognitiveAgent(
		"chapter-agent",
		"Chapter Agent",
		"Agente responsável por criar e gerenciar capítulos dos treinamentos",
		100,
		"gpt-4",
		"chapter",
		"Criar capítulos interativos e desafiadores",
		memManager,
	)
	chapterAgent.SetBackstory("Sou especializado em criar conteúdo educacional estruturado e desafios envolventes.")

	feedbackAgent := agents.NewCognitiveAgent(
		"feedback-agent",
		"Player Feedback Agent",
		"Agente responsável por gerar feedback personalizado",
		100,
		"gpt-4",
		"feedback",
		"Fornecer feedback construtivo e personalizado",
		memManager,
	)
	feedbackAgent.SetBackstory("Sou um analista especializado em avaliar desempenho e fornecer recomendações personalizadas.")

	accountAgent := agents.NewCognitiveAgent(
		"account-agent",
		"Account Agent",
		"Agente responsável por gerenciar dados da conta",
		100,
		"gpt-4",
		"account",
		"Otimizar a experiência de treinamento em nível organizacional",
		memManager,
	)
	accountAgent.SetBackstory("Sou um estrategista focado em análise de dados e recomendações em nível organizacional.")

	// Adiciona os agentes à equipe
	crew.AddAgent(trainingAgent)
	crew.AddAgent(chapterAgent)
	crew.AddAgent(feedbackAgent)
	crew.AddAgent(accountAgent)

	// Define detalhes do projeto de treinamento
	projectDetails := &agents.TrainingProject{
		Name:        "Curso de Programação em Go",
		Description: "Um curso interativo para ensinar programação em Go",
		Objectives: []string{
			"Entender os fundamentos da linguagem Go",
			"Aprender boas práticas de programação",
			"Desenvolver projetos práticos",
		},
		TargetAudience: []string{
			"Desenvolvedores iniciantes",
			"Programadores de outras linguagens",
		},
		Duration:   30 * 24 * time.Hour,
		Difficulty: "Intermediário",
		Prerequisites: []string{
			"Conhecimento básico de programação",
			"Familiaridade com linha de comando",
		},
	}

	// Executa o workflow
	results, err := crew.ExecuteWorkflow(projectDetails)
	if err != nil {
		log.Printf("Erro na execução do workflow: %v", err)
	} else {
		log.Printf("\nResultados do workflow:")
		log.Printf("Treinamento: %s", results.Training)
		log.Printf("Capítulos: %s", results.Chapters)
		log.Printf("Feedback: %s", results.Feedback)
	}

	// Configura handler para sinais de término
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Monitora o projeto
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status := crew.GetProjectStatus()
			log.Printf("\nStatus do projeto:")
			log.Printf("Progresso: %.2f%%", status.Progress)
			log.Printf("Tarefas completadas: %d/%d", status.CompletedTasks, status.TotalTasks)

		case sig := <-sigChan:
			log.Printf("Recebido sinal %v, encerrando...", sig)
			return
		}
	}
}
