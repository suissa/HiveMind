package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

// SemanticMemoryConfig contém a configuração para o gerenciador de memória semântica
type SemanticMemoryConfig struct {
	WeaviateURL string
	APIKey      string
	Class       string
	BatchSize   int
}

// SemanticMemoryManager gerencia memórias usando Weaviate para busca semântica
type SemanticMemoryManager struct {
	client *weaviate.Client
	config *SemanticMemoryConfig
}

// NewSemanticMemoryManager cria um novo gerenciador de memória semântica
func NewSemanticMemoryManager(config *SemanticMemoryConfig) (*SemanticMemoryManager, error) {
	cfg := weaviate.Config{
		Host:   config.WeaviateURL,
		Scheme: "http",
		Headers: map[string]string{
			"X-OpenAI-Api-Key": config.APIKey,
		},
	}

	client, err := weaviate.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente Weaviate: %v", err)
	}

	manager := &SemanticMemoryManager{
		client: client,
		config: config,
	}

	// Garante que a classe existe
	if err := manager.ensureClass(); err != nil {
		return nil, fmt.Errorf("erro ao configurar classe: %v", err)
	}

	return manager, nil
}

// ensureClass garante que a classe necessária existe no Weaviate
func (m *SemanticMemoryManager) ensureClass() error {
	classExists, err := m.client.Schema().ClassExistenceChecker().WithClassName(m.config.Class).Do(context.Background())
	if err != nil {
		return fmt.Errorf("erro ao verificar existência da classe: %v", err)
	}

	if !classExists {
		class := &models.Class{
			Class: m.config.Class,
			Properties: []*models.Property{
				{
					Name:     "content",
					DataType: []string{"text"},
				},
				{
					Name:     "agentId",
					DataType: []string{"string"},
				},
				{
					Name:     "memoryId",
					DataType: []string{"string"},
				},
				{
					Name:     "importance",
					DataType: []string{"number"},
				},
				{
					Name:     "timestamp",
					DataType: []string{"date"},
				},
				{
					Name:     "tags",
					DataType: []string{"string[]"},
				},
			},
		}

		err = m.client.Schema().ClassCreator().WithClass(class).Do(context.Background())
		if err != nil {
			return fmt.Errorf("erro ao criar classe: %v", err)
		}
	}

	return nil
}

// StoreMemory armazena uma memória no Weaviate
func (m *SemanticMemoryManager) StoreMemory(ctx context.Context, memory *Memory) error {
	properties := map[string]interface{}{
		"content":    memory.Content,
		"agentId":    memory.AgentID,
		"memoryId":   memory.ID,
		"importance": memory.Importance,
		"timestamp":  memory.Timestamp.Format(time.RFC3339),
		"tags":       memory.Tags,
	}

	_, err := m.client.Data().Creator().
		WithClassName(m.config.Class).
		WithProperties(properties).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("erro ao armazenar memória no Weaviate: %v", err)
	}

	return nil
}

// SearchSimilarMemories busca memórias semanticamente similares
func (m *SemanticMemoryManager) SearchSimilarMemories(ctx context.Context, query string, limit int) ([]*Memory, error) {
	fields := []graphql.Field{
		{Name: "content"},
		{Name: "agentId"},
		{Name: "memoryId"},
		{Name: "importance"},
		{Name: "timestamp"},
		{Name: "tags"},
	}

	nearText := m.client.GraphQL().NearTextArgBuilder().
		WithConcepts([]string{query})

	result, err := m.client.GraphQL().Get().
		WithClassName(m.config.Class).
		WithFields(fields...).
		WithNearText(nearText).
		WithLimit(limit).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("erro ao buscar memórias similares: %v", err)
	}

	var memories []*Memory
	for _, obj := range result.Data {
		data := obj.(map[string]interface{})
		memory := &Memory{
			Content:    data["content"].(string),
			AgentID:    data["agentId"].(string),
			ID:         data["memoryId"].(string),
			Importance: data["importance"].(float64),
			Tags:       make([]string, 0),
		}

		timestamp, err := time.Parse(time.RFC3339, data["timestamp"].(string))
		if err != nil {
			return nil, fmt.Errorf("erro ao converter timestamp: %v", err)
		}
		memory.Timestamp = timestamp

		if tags, ok := data["tags"].([]interface{}); ok {
			for _, tag := range tags {
				memory.Tags = append(memory.Tags, tag.(string))
			}
		}

		memories = append(memories, memory)
	}

	return memories, nil
}

// UpdateMemory atualiza uma memória existente
func (m *SemanticMemoryManager) UpdateMemory(ctx context.Context, memory *Memory) error {
	properties := map[string]interface{}{
		"content":    memory.Content,
		"importance": memory.Importance,
		"timestamp":  memory.Timestamp.Format(time.RFC3339),
		"tags":       memory.Tags,
	}

	err := m.client.Data().Updater().
		WithClassName(m.config.Class).
		WithID(memory.ID).
		WithProperties(properties).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("erro ao atualizar memória: %v", err)
	}

	return nil
}

// DeleteMemory remove uma memória
func (m *SemanticMemoryManager) DeleteMemory(ctx context.Context, memoryID string) error {
	err := m.client.Data().Deleter().
		WithClassName(m.config.Class).
		WithID(memoryID).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("erro ao deletar memória: %v", err)
	}

	return nil
}

// Close fecha a conexão com o Weaviate
func (m *SemanticMemoryManager) Close(ctx context.Context) error {
	// O cliente Weaviate não requer fechamento explícito
	return nil
}
