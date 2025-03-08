package tools

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
)

// WeaviateClient implementa a interface SemanticSearchTool
type WeaviateClient struct {
	client *weaviate.Client
	ctx    context.Context
}

// NewWeaviateClient cria uma nova instância do WeaviateClient
func NewWeaviateClient() (*WeaviateClient, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar arquivo .env: %v", err)
	}

	url := os.Getenv("WEAVIATE_URL")
	if url == "" {
		return nil, fmt.Errorf("WEAVIATE_URL não encontrado no arquivo .env")
	}

	apiKey := os.Getenv("WEAVIATE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("WEAVIATE_API_KEY não encontrado no arquivo .env")
	}

	cfg := weaviate.Config{
		Host:       url,
		Scheme:     "https",
		AuthConfig: auth.ApiKey{Value: apiKey},
	}

	client, err := weaviate.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente Weaviate: %v", err)
	}

	return &WeaviateClient{
		client: client,
		ctx:    context.Background(),
	}, nil
}

// Search implementa a busca semântica no Weaviate
func (w *WeaviateClient) Search(options SemanticSearchOptions) (*SemanticSearchResult, error) {
	startTime := time.Now()

	// Preparar campos para retornar
	fields := []graphql.Field{
		{Name: "_additional { id distance score vector }"},
	}
	if len(options.Properties) > 0 {
		for _, prop := range options.Properties {
			fields = append(fields, graphql.Field{Name: prop})
		}
	}

	// Construir query
	whereFilter := filters.Where{}
	if len(options.Filters) > 0 {
		for key, value := range options.Filters {
			whereFilter.WithPath([]string{key}).WithOperator(filters.Equal).WithValueString(fmt.Sprint(value))
		}
	}

	// Configurar busca
	query := w.client.GraphQL().Get()
	query = query.WithClassName(options.Class)
	query = query.WithFields(fields...)

	if options.Limit > 0 {
		query = query.WithLimit(options.Limit)
	}
	if options.Offset > 0 {
		query = query.WithOffset(options.Offset)
	}

	// Adicionar busca por texto se fornecido
	if options.Query != "" {
		query = query.WithNearText(w.client.GraphQL().NearTextArgBuilder().
			WithConcepts([]string{options.Query}))
	}

	// Adicionar busca por vetor se fornecido
	if options.NearVector != nil {
		query = query.WithNearVector(w.client.GraphQL().NearVectorArgBuilder().
			WithVector(options.NearVector))
	}

	// Executar busca
	result, err := query.Do(w.ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao executar busca: %v", err)
	}

	// Processar resultados
	searchResult := &SemanticSearchResult{
		Results:        make([]SemanticDocument, 0),
		ProcessingTime: time.Since(startTime).Milliseconds(),
	}

	// Extrair resultados do GraphQL
	if data, ok := result.Data["Get"].(map[string]interface{}); ok {
		if items, ok := data[options.Class].([]interface{}); ok {
			for _, item := range items {
				if obj, ok := item.(map[string]interface{}); ok {
					doc := SemanticDocument{
						Class:      options.Class,
						Properties: make(map[string]interface{}),
					}

					// Extrair propriedades adicionais
					if additional, ok := obj["_additional"].(map[string]interface{}); ok {
						doc.ID = fmt.Sprint(additional["id"])
						doc.Score = additional["score"].(float64)
						if options.IncludeVector {
							if vector, ok := additional["vector"].([]interface{}); ok {
								doc.Vector = make([]float32, len(vector))
								for i, v := range vector {
									doc.Vector[i] = float32(v.(float64))
								}
							}
						}
						if dist, ok := additional["distance"].(float64); ok {
							doc.Distance = dist
						}
					}

					// Extrair propriedades regulares
					for key, value := range obj {
						if key != "_additional" {
							doc.Properties[key] = value
						}
					}

					searchResult.Results = append(searchResult.Results, doc)
				}
			}
		}
	}

	searchResult.Total = len(searchResult.Results)
	return searchResult, nil
}

// AddDocument adiciona um documento ao Weaviate
func (w *WeaviateClient) AddDocument(class string, properties map[string]interface{}, vector []float32) error {
	_, err := w.client.Data().Creator().
		WithClassName(class).
		WithProperties(properties).
		WithVector(vector).
		Do(w.ctx)

	if err != nil {
		return fmt.Errorf("erro ao adicionar documento: %v", err)
	}

	return nil
}

// DeleteDocument remove um documento do Weaviate
func (w *WeaviateClient) DeleteDocument(class string, id string) error {
	err := w.client.Data().Deleter().
		WithClassName(class).
		WithID(id).
		Do(w.ctx)

	if err != nil {
		return fmt.Errorf("erro ao deletar documento: %v", err)
	}

	return nil
}

// CreateClass cria uma nova classe no Weaviate
func (w *WeaviateClient) CreateClass(class string, properties map[string]interface{}) error {
	classObj := &models.Class{
		Class:      class,
		Vectorizer: "text2vec-openai", // Pode ser configurável
		Properties: make([]*models.Property, 0),
	}

	for name, prop := range properties {
		if propMap, ok := prop.(map[string]interface{}); ok {
			property := &models.Property{
				Name:     name,
				DataType: []string{fmt.Sprint(propMap["type"])},
			}
			classObj.Properties = append(classObj.Properties, property)
		}
	}

	err := w.client.Schema().ClassCreator().WithClass(classObj).Do(w.ctx)
	if err != nil {
		return fmt.Errorf("erro ao criar classe: %v", err)
	}

	return nil
}

// DeleteClass remove uma classe do Weaviate
func (w *WeaviateClient) DeleteClass(class string) error {
	err := w.client.Schema().ClassDeleter().WithClassName(class).Do(w.ctx)
	if err != nil {
		return fmt.Errorf("erro ao deletar classe: %v", err)
	}

	return nil
} 