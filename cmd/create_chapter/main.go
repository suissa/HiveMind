package main

import (
	"log"

	"groq-consumer/publishers"
)

func main() {
	// Criar mensagem para um novo capítulo sobre IA
	chapterMessage := map[string]interface{}{
		"tema": "Inteligência Artificial: Fundamentos e Aplicações",
		"user_info": map[string]interface{}{
			"level":      "Iniciante",
			"profession": "estudante",
			"age":        25,
		},
	}

	log.Printf("🚀 Iniciando criação de capítulo sobre IA...")

	// Publicar a mensagem
	err := publishers.PublishEvent("chapter.creation.queue", chapterMessage)
	if err != nil {
		log.Fatalf("❌ Erro ao publicar mensagem: %v", err)
	}

	log.Printf("✅ Solicitação de criação de capítulo enviada com sucesso!")
	log.Printf("ℹ️  O sistema irá:")
	log.Printf("   1. Gerar 3 quizzes sobre o tema")
	log.Printf("   2. Criar 2 desafios práticos")
	log.Printf("   3. Avaliar e pontuar cada conteúdo")
	log.Printf("   4. Finalizar quando atingir 1000 pontos")
}
