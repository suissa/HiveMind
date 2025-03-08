package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/streadway/amqp"
)

type GroqRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	// Conectar ao RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:1234/")
	failOnError(err, "Falha ao conectar ao RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Falha ao abrir o canal")
	defer ch.Close()

	// Declarar as filas
	inputQueue, err := ch.QueueDeclare(
		"start_queue", // nome
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	failOnError(err, "Falha ao declarar a fila de entrada")

	outputQueue, err := ch.QueueDeclare(
		"task.completed", // nome
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Falha ao declarar a fila de saída")

	msgs, err := ch.Consume(
		inputQueue.Name, // queue
		"",              // consumer
		true,            // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	failOnError(err, "Falha ao registrar um consumidor")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			if string(d.Body) == "chapter.creation.queue" {
				// Preparar a requisição para a Groq API
				groqReq := GroqRequest{
					Model: "llama-3.3-70b-versatile",
					Messages: []Message{
						{
							Role:    "user",
							Content: "Explain the importance of fast language models",
						},
					},
				}

				jsonData, err := json.Marshal(groqReq)
				if err != nil {
					log.Printf("Erro ao criar JSON: %s", err)
					continue
				}

				// Fazer a requisição para a Groq API
				req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
				if err != nil {
					log.Printf("Erro ao criar requisição: %s", err)
					continue
				}

				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+os.Getenv("GROQ_API_KEY"))

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					log.Printf("Erro ao fazer requisição: %s", err)
					continue
				}
				defer resp.Body.Close()

				// Publicar mensagem de conclusão
				err = ch.Publish(
					"",               // exchange
					outputQueue.Name, // routing key
					false,            // mandatory
					false,            // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte("Tarefa concluída"),
					})
				if err != nil {
					log.Printf("Erro ao publicar mensagem: %s", err)
				}

				log.Printf("Tarefa processada com sucesso")
			}
		}
	}()

	log.Printf(" [*] Aguardando mensagens. Para sair pressione CTRL+C")
	<-forever
}
