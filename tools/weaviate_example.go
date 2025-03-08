package tools

import (
	"fmt"
	"log"
)

// ExampleWeaviate demonstra o uso do cliente Weaviate
func ExampleWeaviate() {
	// Criar cliente Weaviate
	client, err := NewWeaviateClient()
	if err != nil {
		log.Fatal(err)
	}

	// Criar uma classe para artigos
	articleProperties := map[string]interface{}{
		"title": map[string]interface{}{
			"type": "string",
		},
		"content": map[string]interface{}{
			"type": "text",
		},
		"category": map[string]interface{}{
			"type": "string",
		},
		"publishDate": map[string]interface{}{
			"type": "date",
		},
	}

	err = client.CreateClass("Article", articleProperties)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Classe Article criada com sucesso")

	// Adicionar alguns documentos
	articles := []map[string]interface{}{
		{
			"title":       "Introdução à Inteligência Artificial",
			"content":     "A IA é um campo da computação que busca criar máquinas inteligentes...",
			"category":    "Tecnologia",
			"publishDate": "2023-01-01T00:00:00Z",
		},
		{
			"title":       "Machine Learning na Prática",
			"content":     "Machine Learning é uma subárea da IA que foca em aprendizado automático...",
			"category":    "Tecnologia",
			"publishDate": "2023-02-01T00:00:00Z",
		},
	}

	for _, article := range articles {
		err = client.AddDocument("Article", article, nil) // Vetor será gerado automaticamente pelo vectorizer
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Documentos adicionados com sucesso")

	// Busca semântica simples
	result, err := client.Search(SemanticSearchOptions{
		Class:      "Article",
		Query:      "como funciona inteligência artificial",
		Properties: []string{"title", "content", "category"},
		Limit:      5,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nResultados da busca (%d encontrados):\n", result.Total)
	for _, doc := range result.Results {
		fmt.Printf("- %s (Score: %.2f)\n", doc.Properties["title"], doc.Score)
		fmt.Printf("  Categoria: %s\n", doc.Properties["category"])
	}

	// Busca com filtros
	result, err = client.Search(SemanticSearchOptions{
		Class: "Article",
		Query: "machine learning",
		Filters: map[string]interface{}{
			"category": "Tecnologia",
		},
		Properties: []string{"title", "publishDate"},
		Sort:       []string{"publishDate"},
		Limit:      10,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nArtigos sobre Machine Learning em Tecnologia:\n")
	for _, doc := range result.Results {
		fmt.Printf("- %s (Publicado em: %s)\n",
			doc.Properties["title"],
			doc.Properties["publishDate"],
		)
	}
}

// ExampleWeaviateAdvanced demonstra recursos avançados do Weaviate
func ExampleWeaviateAdvanced() {
	client, err := NewWeaviateClient()
	if err != nil {
		log.Fatal(err)
	}

	// Busca por similaridade vetorial
	vector := []float32{0.1, 0.2, 0.3, 0.4} // Exemplo de vetor
	result, err := client.Search(SemanticSearchOptions{
		Class:       "Article",
		NearVector:  vector,
		Properties:  []string{"title", "content"},
		Limit:       5,
		Distance:    0.7,
		IncludeVector: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nBusca por similaridade vetorial:\n")
	for _, doc := range result.Results {
		fmt.Printf("- %s (Distância: %.3f)\n",
			doc.Properties["title"],
			doc.Distance,
		)
	}

	// Busca híbrida (texto + vetor)
	result, err = client.Search(SemanticSearchOptions{
		Class:      "Article",
		Query:      "inteligência artificial",
		NearVector: vector,
		Properties: []string{"title", "content"},
		SearchParams: map[string]interface{}{
			"hybrid": map[string]interface{}{
				"alpha": 0.5, // Peso entre busca textual e vetorial
			},
		},
		Limit: 5,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nBusca híbrida:\n")
	for _, doc := range result.Results {
		fmt.Printf("- %s (Score: %.3f)\n",
			doc.Properties["title"],
			doc.Score,
		)
	}
} 