package tools

import (
	"fmt"
	"log"
)

// ExampleMeilisearch demonstra o uso do cliente Meilisearch
func ExampleMeilisearch() {
	// Criar cliente Meilisearch
	client, err := NewMeilisearchClient()
	if err != nil {
		log.Fatal(err)
	}

	// Exemplo de indexação de documentos a partir de um arquivo JSON
	result, err := client.Index("movies", "movies.json", "id")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Indexação iniciada: TaskID=%d, Status=%s\n", result.TaskID, result.Status)

	// Exemplo de indexação de documentos a partir de uma estrutura de dados
	movies := []map[string]interface{}{
		{
			"id":    1,
			"title": "O Poderoso Chefão",
			"year":  1972,
		},
		{
			"id":    2,
			"title": "Matrix",
			"year":  1999,
		},
	}
	result, err = client.Index("movies", movies, "id")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Indexação iniciada: TaskID=%d, Status=%s\n", result.TaskID, result.Status)

	// Exemplo de busca simples
	searchResult, err := client.Search(SearchEngineOptions{
		IndexName: "movies",
		Query:    "matrix",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Encontrados %d resultados em %dms\n", searchResult.Total, searchResult.ProcessingTime)
	for _, hit := range searchResult.Hits {
		fmt.Printf("- %s (%d)\n", hit["title"], hit["year"])
	}

	// Exemplo de busca com filtros
	searchResult, err = client.Search(SearchEngineOptions{
		IndexName: "movies",
		Query:    "",
		Filters: map[string]interface{}{
			"year": map[string]interface{}{
				"$gte": 1990,
			},
		},
		Sort: []string{"year:desc"},
		Limit: 10,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nFilmes após 1990:\n")
	for _, hit := range searchResult.Hits {
		fmt.Printf("- %s (%d)\n", hit["title"], hit["year"])
	}

	// Exemplo de obtenção de estatísticas
	stats, err := client.GetStats("movies")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nEstatísticas do índice:\n")
	fmt.Printf("Total de documentos: %v\n", stats["numberOfDocuments"])
	fmt.Printf("Última atualização: %v\n", stats["lastUpdate"])

	// Exemplo de deleção de índice
	err = client.DeleteIndex("movies")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nÍndice deletado com sucesso")
}

// ExampleMeilisearchAdvanced demonstra recursos avançados do Meilisearch
func ExampleMeilisearchAdvanced() {
	client, err := NewMeilisearchClient()
	if err != nil {
		log.Fatal(err)
	}

	// Busca com múltiplos filtros e parâmetros avançados
	searchResult, err := client.Search(SearchEngineOptions{
		IndexName: "movies",
		Query:    "action",
		Filters: map[string]interface{}{
			"$and": []map[string]interface{}{
				{"year": map[string]interface{}{"$gte": 2000}},
				{"rating": map[string]interface{}{"$gt": 4.0}},
			},
		},
		Sort:   []string{"rating:desc", "year:desc"},
		Limit:  20,
		Offset: 0,
		SearchParams: map[string]interface{}{
			"matchingStrategy": "all",
			"attributesToRetrieve": []string{
				"id",
				"title",
				"year",
				"rating",
			},
			"attributesToHighlight": []string{"title"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Busca avançada:\n")
	fmt.Printf("Total: %d, Tempo: %dms\n", searchResult.Total, searchResult.ProcessingTime)
	for _, hit := range searchResult.Hits {
		fmt.Printf("- %s (%d) - Rating: %.1f\n",
			hit["title"],
			hit["year"],
			hit["rating"],
		)
	}
} 