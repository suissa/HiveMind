package main

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

type UserInfo struct {
	Level      string `json:"level"`
	Profession string `json:"profession"`
	Age        int    `json:"age"`
}

type ChapterRequest struct {
	Tema     string   `json:"tema"`
	UserInfo UserInfo `json:"user_info"`
}

func handleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func publishChapterRequest() {
	// Conectar ao RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:1234/")
	handleError(err, "Falha ao conectar ao RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	handleError(err, "Falha ao abrir o canal")
	defer ch.Close()

	// Declarar a fila
	q, err := ch.QueueDeclare(
		"chapter.creation.queue", // nome
		true,                     // durable
		false,                    // delete when unused
		false,                    // exclusive
		false,                    // no-wait
		nil,                      // arguments
	)
	handleError(err, "Falha ao declarar a fila")

	// Criar a mensagem
	message := ChapterRequest{
		Tema: "História da Computação",
		UserInfo: UserInfo{
			Level:      "Intermediário",
			Profession: "teacher",
			Age:        30,
		},
	}

	// Converter para JSON
	body, err := json.Marshal(message)
	handleError(err, "Falha ao converter mensagem para JSON")

	// Publicar a mensagem
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	handleError(err, "Falha ao publicar mensagem")

	log.Printf(" [x] Mensagem enviada com sucesso: %s", body)
}

func init() {
	publishChapterRequest()
}
