package memory

import (
	"context"
	"time"
)

// MemoryType representa o tipo de memória
type MemoryType string

const (
	// ShortTerm representa memória de curto prazo
	ShortTerm MemoryType = "short_term"
	// LongTerm representa memória de longo prazo
	LongTerm MemoryType = "long_term"
)

// Memory representa uma unidade de memória
type Memory struct {
	ID         string        `json:"id" bson:"_id"`
	AgentID    string        `json:"agent_id" bson:"agent_id"`
	Content    string        `json:"content" bson:"content"`
	Type       MemoryType    `json:"type" bson:"type"`
	Importance float64       `json:"importance" bson:"importance"`
	Timestamp  time.Time     `json:"timestamp" bson:"timestamp"`
	TTL        time.Duration `json:"ttl" bson:"ttl"`
	Tags       []string      `json:"tags" bson:"tags"`
	Metadata   interface{}   `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// MemoryConfig contém a configuração para o sistema de memória
type MemoryConfig struct {
	// Redis
	RedisURL string `json:"redis_url" yaml:"redis_url"`

	// MongoDB
	MongoURL   string `json:"mongo_url" yaml:"mongo_url"`
	MongoDB    string `json:"mongo_db" yaml:"mongo_db"`
	Collection string `json:"collection" yaml:"collection"`

	// Weaviate
	WeaviateURL       string `json:"weaviate_url" yaml:"weaviate_url"`
	WeaviateAPIKey    string `json:"weaviate_api_key" yaml:"weaviate_api_key"`
	WeaviateClass     string `json:"weaviate_class" yaml:"weaviate_class"`
	WeaviateBatchSize int    `json:"weaviate_batch_size" yaml:"weaviate_batch_size"`

	// Configurações gerais
	ImportanceThreshold float64       `json:"importance_threshold" yaml:"importance_threshold"`
	ShortTermTTL        time.Duration `json:"short_term_ttl" yaml:"short_term_ttl"`
}

// DefaultMemoryConfig retorna uma configuração padrão
func DefaultMemoryConfig() *MemoryConfig {
	return &MemoryConfig{
		RedisURL:            "localhost:6379",
		MongoURL:            "mongodb://localhost:27017",
		MongoDB:             "agent_memory",
		Collection:          "memories",
		WeaviateURL:         "http://localhost:8080",
		WeaviateClass:       "Memory",
		WeaviateBatchSize:   100,
		ImportanceThreshold: 0.7,
		ShortTermTTL:        24 * time.Hour,
	}
}

// MemoryManager define a interface para gerenciamento de memória
type MemoryManager interface {
	// StoreMemory armazena uma nova memória
	StoreMemory(ctx context.Context, memory *Memory) error

	// GetMemory recupera uma memória específica
	GetMemory(ctx context.Context, agentID, memoryID string) (*Memory, error)

	// SearchMemories busca memórias por tags
	SearchMemories(ctx context.Context, agentID string, tags []string) ([]*Memory, error)

	// SearchSimilarMemories busca memórias semanticamente similares
	SearchSimilarMemories(ctx context.Context, query string, limit int) ([]*Memory, error)

	// UpdateMemory atualiza uma memória existente
	UpdateMemory(ctx context.Context, memory *Memory) error

	// DeleteMemory remove uma memória específica
	DeleteMemory(ctx context.Context, agentID, memoryID string) error

	// ConsolidateMemories move memórias importantes para o armazenamento de longo prazo
	ConsolidateMemories(ctx context.Context, agentID string) error

	// PruneMemories remove memórias antigas ou irrelevantes
	PruneMemories(ctx context.Context, agentID string) error

	// Close fecha as conexões com os bancos de dados
	Close(ctx context.Context) error
}
