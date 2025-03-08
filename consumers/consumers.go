package consumers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"groq-consumer/agents"

	"github.com/streadway/amqp"
)

const (
	RABBITMQ_HOST = "localhost"
	RABBITMQ_PORT = 1234
)

var (
	orchestrator   = agents.NewOrchestratorAgent()
	quizAgent      = agents.NewQuizAgent()
	challengeAgent = agents.NewChallengeAgent()
)

func init() {
	// Atribuir os agentes cognitivos ao orquestrador
	orchestrator.AssignCognitiveAgents(quizAgent, challengeAgent)
}

type UserInfo struct {
	Level      string `json:"level"`
	Profession string `json:"profession"`
	Age        int    `json:"age"`
}

type ChapterCreationMessage struct {
	Tema     string   `json:"tema"`
	UserInfo UserInfo `json:"user_info"`
}

type TaskMessage struct {
	Task string `json:"task"`
	Tema string `json:"tema"`
}

type ApprovalMessage struct {
	Type    string `json:"type"`
	Prompt  string `json:"prompt"`
	Content string `json:"content"`
	Score   int    `json:"score"`
}

type FinishedMessage struct {
	Status     string `json:"status"`
	TotalScore int    `json:"total_score"`
}

func handleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func publishEvent(queueName string, message interface{}) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@%s:%d/", RABBITMQ_HOST, RABBITMQ_PORT))
	handleError(err, "Falha ao conectar ao RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	handleError(err, "Falha ao abrir o canal")
	defer ch.Close()

	body, err := json.Marshal(message)
	handleError(err, "Falha ao converter mensagem para JSON")

	err = ch.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: 2, // mensagem persistente
		})
	handleError(err, "Falha ao publicar mensagem")
	log.Printf("📨 Mensagem enviada para %s: %s", queueName, string(body))
}

func startChapter(tema string, userInfo UserInfo) {
	log.Printf("\n📖 Iniciando criação do capítulo: %s", tema)
	log.Printf("👤 Informações do usuário: %+v", userInfo)

	// Publicar eventos para criação de quizzes
	for i := 0; i < 3; i++ {
		publishEvent("quiz.creation.queue", TaskMessage{
			Task: "generate_quiz",
			Tema: tema,
		})
	}

	// Publicar eventos para criação de desafios
	for i := 0; i < 2; i++ {
		publishEvent("challenge.creation.queue", TaskMessage{
			Task: "generate_challenge",
			Tema: tema,
		})
	}

	// Loop de espera para aguardar a conclusão do capítulo
	go func() {
		for !orchestrator.IsWorkflowComplete() {
			progress := orchestrator.GetProgress()
			log.Printf("⏳ Aguardando... Progresso: %.1f%%", progress)
			time.Sleep(2 * time.Second)
		}
		log.Printf("🎉 Capítulo finalizado com sucesso!")
	}()
}

func generateQuiz(tema string) {
	score := rand.Intn(451) + 50 // Gera número entre 50 e 500

	task := &agents.Task{
		Description:    fmt.Sprintf("Crie um quiz sobre '%s'. Sugestão de pontuação: %d pontos.", tema, score),
		ExpectedOutput: "Um conjunto de perguntas e respostas relacionadas ao tema.",
	}

	result, err := orchestrator.DelegateTask(quizAgent, task)
	if err != nil {
		log.Printf("Erro ao gerar quiz: %v", err)
		return
	}

	// Enviar para aprovação
	publishEvent("chapter.approval.queue", ApprovalMessage{
		Type:    "quiz",
		Prompt:  task.Description,
		Content: result,
		Score:   score,
	})
}

func generateChallenge(tema string) {
	score := rand.Intn(451) + 50 // Gera número entre 50 e 500

	task := &agents.Task{
		Description:    fmt.Sprintf("Crie um desafio sobre '%s'. Sugestão de pontuação: %d pontos.", tema, score),
		ExpectedOutput: "Um desafio interativo que teste o conhecimento do usuário sobre o tema.",
	}

	result, err := orchestrator.DelegateTask(challengeAgent, task)
	if err != nil {
		log.Printf("Erro ao gerar desafio: %v", err)
		return
	}

	// Enviar para aprovação
	publishEvent("chapter.approval.queue", ApprovalMessage{
		Type:    "challenge",
		Prompt:  task.Description,
		Content: result,
		Score:   score,
	})
}

func processChapterCreation(body []byte) {
	var message ChapterCreationMessage
	err := json.Unmarshal(body, &message)
	handleError(err, "Erro ao decodificar mensagem")

	log.Printf("\n🔔 Criando um novo Capítulo...")
	startChapter(message.Tema, message.UserInfo)
}

func processGenerateQuiz(body []byte) {
	var message TaskMessage
	err := json.Unmarshal(body, &message)
	handleError(err, "Erro ao decodificar mensagem")

	log.Printf("\n🔔 Criando um novo Quiz sobre '%s'...", message.Tema)
	generateQuiz(message.Tema)
}

func processGenerateChallenge(body []byte) {
	var message TaskMessage
	err := json.Unmarshal(body, &message)
	handleError(err, "Erro ao decodificar mensagem")

	log.Printf("\n🔔 Criando um novo Desafio sobre '%s'...", message.Tema)
	generateChallenge(message.Tema)
}

func processChapterApproval(body []byte) {
	var message ApprovalMessage
	err := json.Unmarshal(body, &message)
	handleError(err, "Erro ao decodificar mensagem")

	log.Printf("\n🔔 Aprovando %s...", message.Type)

	approved, err := orchestrator.EvaluateContent(message.Content, message.Score)
	if err != nil {
		log.Printf("Erro ao avaliar conteúdo: %v", err)
		return
	}

	if approved && orchestrator.IsWorkflowComplete() {
		publishEvent("chapter.finished.queue", FinishedMessage{
			Status:     "completed",
			TotalScore: orchestrator.TotalScore,
		})
	}
}

func startConsumer(queueName string, handler func([]byte)) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@%s:%d/", RABBITMQ_HOST, RABBITMQ_PORT))
	handleError(err, "Falha ao conectar ao RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	handleError(err, "Falha ao abrir o canal")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName, // nome
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	handleError(err, "Falha ao declarar a fila")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	handleError(err, "Falha ao registrar um consumidor")

	log.Printf("📡 Aguardando mensagens na fila `%s`...", queueName)

	for d := range msgs {
		handler(d.Body)
	}
}

func StartConsumers() {
	rand.Seed(time.Now().UnixNano())

	var wg sync.WaitGroup

	consumers := []struct {
		queueName string
		handler   func([]byte)
	}{
		{"chapter.creation.queue", processChapterCreation},
		{"quiz.creation.queue", processGenerateQuiz},
		{"challenge.creation.queue", processGenerateChallenge},
		{"chapter.approval.queue", processChapterApproval},
	}

	for _, consumer := range consumers {
		wg.Add(1)
		go func(queueName string, handler func([]byte)) {
			defer wg.Done()
			startConsumer(queueName, handler)
		}(consumer.queueName, consumer.handler)
	}

	wg.Wait()
}
