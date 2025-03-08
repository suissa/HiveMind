package publishers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

const (
	RABBITMQ_HOST = "localhost"
	RABBITMQ_PORT = 1234
)

// PublishEvent publica uma mensagem em uma fila específica do RabbitMQ
func PublishEvent(queueName string, message interface{}) error {
	// Conectar ao RabbitMQ
	conn, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@%s:%d/", RABBITMQ_HOST, RABBITMQ_PORT))
	if err != nil {
		return fmt.Errorf("falha ao conectar ao RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Criar canal
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("falha ao abrir o canal: %v", err)
	}
	defer ch.Close()

	// Declarar a fila
	_, err = ch.QueueDeclare(
		queueName, // nome
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("falha ao declarar a fila: %v", err)
	}

	// Converter a mensagem para JSON
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("falha ao converter mensagem para JSON: %v", err)
	}

	// Publicar a mensagem
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
	if err != nil {
		return fmt.Errorf("falha ao publicar mensagem: %v", err)
	}

	log.Printf("📨 Mensagem enviada para %s: %s", queueName, string(body))
	return nil
}
