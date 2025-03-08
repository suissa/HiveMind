package main

import (
	"log"

	"groq-consumer/publishers"
)

func main() {
	// Exemplo de envio de mensagem para acionar o CrewAI
	message := map[string]interface{}{
		"data": "Analise as tendências de vendas do último mês.",
	}

	err := publishers.PublishEvent("agent.tasks.queue", message)
	if err != nil {
		log.Fatalf("Erro ao publicar mensagem: %v", err)
	}

	// Exemplo de envio de mensagem para criar um capítulo
	chapterMessage := map[string]interface{}{
		"tema": "História da Computação",
		"user_info": map[string]interface{}{
			"level":      "Intermediário",
			"profession": "teacher",
			"age":        30,
		},
	}

	err = publishers.PublishEvent("chapter.creation.queue", chapterMessage)
	if err != nil {
		log.Fatalf("Erro ao publicar mensagem: %v", err)
	}
}
