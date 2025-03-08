package memory

import (
	"context"
	"fmt"
)

// HybridMemoryManager combina Redis (curto prazo), MongoDB (longo prazo) e Weaviate (semântica)
type HybridMemoryManager struct {
	shortTerm *RedisMemoryManager
	longTerm  *MongoMemoryManager
	semantic  *SemanticMemoryManager
	config    *MemoryConfig
}

// NewHybridMemoryManager cria um novo gerenciador de memória híbrido
func NewHybridMemoryManager(ctx context.Context, config *MemoryConfig) (*HybridMemoryManager, error) {
	// Inicializa Redis para memória de curto prazo
	shortTerm, err := NewRedisMemoryManager(config.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar Redis: %v", err)
	}

	// Inicializa MongoDB para memória de longo prazo
	longTerm, err := NewMongoMemoryManager(ctx, config.MongoURL, config.MongoDB, config.Collection)
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar MongoDB: %v", err)
	}

	// Inicializa Weaviate para memória semântica
	semantic, err := NewSemanticMemoryManager(&SemanticMemoryConfig{
		WeaviateURL: config.WeaviateURL,
		APIKey:      config.WeaviateAPIKey,
		Class:       config.WeaviateClass,
		BatchSize:   config.WeaviateBatchSize,
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar Weaviate: %v", err)
	}

	return &HybridMemoryManager{
		shortTerm: shortTerm,
		longTerm:  longTerm,
		semantic:  semantic,
		config:    config,
	}, nil
}

// StoreMemory armazena uma memória no sistema apropriado
func (m *HybridMemoryManager) StoreMemory(ctx context.Context, memory *Memory) error {
	// Armazena na memória semântica para busca por similaridade
	if err := m.semantic.StoreMemory(ctx, memory); err != nil {
		return fmt.Errorf("erro ao armazenar na memória semântica: %v", err)
	}

	// Decide onde armazenar com base na importância
	if memory.Importance >= m.config.ImportanceThreshold {
		// Memória importante vai para o armazenamento de longo prazo
		if err := m.longTerm.StoreMemory(ctx, memory); err != nil {
			return fmt.Errorf("erro ao armazenar na memória de longo prazo: %v", err)
		}
	} else {
		// Memória menos importante vai para o armazenamento de curto prazo
		if err := m.shortTerm.StoreMemory(ctx, memory); err != nil {
			return fmt.Errorf("erro ao armazenar na memória de curto prazo: %v", err)
		}
	}

	return nil
}

// GetMemory recupera uma memória específica
func (m *HybridMemoryManager) GetMemory(ctx context.Context, agentID, memoryID string) (*Memory, error) {
	// Tenta primeiro na memória de curto prazo
	memory, err := m.shortTerm.GetMemory(ctx, agentID, memoryID)
	if err == nil {
		return memory, nil
	}

	// Se não encontrou, tenta na memória de longo prazo
	memory, err = m.longTerm.GetMemory(ctx, agentID, memoryID)
	if err == nil {
		return memory, nil
	}

	return nil, fmt.Errorf("memória não encontrada")
}

// SearchMemories busca memórias por tags
func (m *HybridMemoryManager) SearchMemories(ctx context.Context, agentID string, tags []string) ([]*Memory, error) {
	var allMemories []*Memory

	// Busca em todas as camadas
	shortTermMemories, _ := m.shortTerm.SearchMemories(ctx, agentID, tags)
	longTermMemories, _ := m.longTerm.SearchMemories(ctx, agentID, tags)

	// Combina os resultados
	allMemories = append(allMemories, shortTermMemories...)
	allMemories = append(allMemories, longTermMemories...)

	return allMemories, nil
}

// SearchSimilarMemories busca memórias semanticamente similares
func (m *HybridMemoryManager) SearchSimilarMemories(ctx context.Context, query string, limit int) ([]*Memory, error) {
	return m.semantic.SearchSimilarMemories(ctx, query, limit)
}

// ConsolidateMemories move memórias importantes para o armazenamento de longo prazo
func (m *HybridMemoryManager) ConsolidateMemories(ctx context.Context, agentID string) error {
	// Busca todas as memórias de curto prazo
	memories, err := m.shortTerm.SearchMemories(ctx, agentID, []string{})
	if err != nil {
		return fmt.Errorf("erro ao buscar memórias para consolidação: %v", err)
	}

	// Avalia cada memória
	for _, memory := range memories {
		if memory.Importance >= m.config.ImportanceThreshold {
			// Move para memória de longo prazo
			if err := m.longTerm.StoreMemory(ctx, memory); err != nil {
				return fmt.Errorf("erro ao consolidar memória: %v", err)
			}

			// Remove da memória de curto prazo
			if err := m.shortTerm.DeleteMemory(ctx, agentID, memory.ID); err != nil {
				return fmt.Errorf("erro ao remover memória consolidada: %v", err)
			}
		}
	}

	return nil
}

// PruneMemories remove memórias antigas ou irrelevantes
func (m *HybridMemoryManager) PruneMemories(ctx context.Context, agentID string) error {
	// Remove memórias antigas da memória de curto prazo
	if err := m.shortTerm.PruneMemories(ctx, agentID); err != nil {
		return fmt.Errorf("erro ao limpar memórias de curto prazo: %v", err)
	}

	// Remove memórias pouco importantes da memória de longo prazo
	if err := m.longTerm.PruneMemories(ctx, agentID); err != nil {
		return fmt.Errorf("erro ao limpar memórias de longo prazo: %v", err)
	}

	return nil
}

// DeleteMemory remove uma memória de todos os sistemas de armazenamento
func (m *HybridMemoryManager) DeleteMemory(ctx context.Context, agentID, memoryID string) error {
	var errors []error

	// Remove da memória semântica
	if err := m.semantic.DeleteMemory(ctx, memoryID); err != nil {
		errors = append(errors, fmt.Errorf("erro ao remover da memória semântica: %v", err))
	}

	// Remove da memória de curto prazo
	if err := m.shortTerm.DeleteMemory(ctx, agentID, memoryID); err != nil {
		errors = append(errors, fmt.Errorf("erro ao remover da memória de curto prazo: %v", err))
	}

	// Remove da memória de longo prazo
	if err := m.longTerm.DeleteMemory(ctx, agentID, memoryID); err != nil {
		errors = append(errors, fmt.Errorf("erro ao remover da memória de longo prazo: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("erros ao remover memória: %v", errors)
	}

	return nil
}

// UpdateMemory atualiza uma memória existente
func (m *HybridMemoryManager) UpdateMemory(ctx context.Context, memory *Memory) error {
	// Atualiza na memória semântica
	if err := m.semantic.UpdateMemory(ctx, memory); err != nil {
		return fmt.Errorf("erro ao atualizar na memória semântica: %v", err)
	}

	// Decide onde atualizar com base na importância
	if memory.Importance >= m.config.ImportanceThreshold {
		// Memória importante vai para o armazenamento de longo prazo
		if err := m.longTerm.UpdateMemory(ctx, memory); err != nil {
			return fmt.Errorf("erro ao atualizar na memória de longo prazo: %v", err)
		}
	} else {
		// Memória menos importante vai para o armazenamento de curto prazo
		if err := m.shortTerm.UpdateMemory(ctx, memory); err != nil {
			return fmt.Errorf("erro ao atualizar na memória de curto prazo: %v", err)
		}
	}

	return nil
}

// Close fecha todas as conexões
func (m *HybridMemoryManager) Close(ctx context.Context) error {
	var errors []error

	if err := m.shortTerm.Close(ctx); err != nil {
		errors = append(errors, fmt.Errorf("erro ao fechar Redis: %v", err))
	}

	if err := m.longTerm.Close(ctx); err != nil {
		errors = append(errors, fmt.Errorf("erro ao fechar MongoDB: %v", err))
	}

	if err := m.semantic.Close(ctx); err != nil {
		errors = append(errors, fmt.Errorf("erro ao fechar Weaviate: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("erros ao fechar conexões: %v", errors)
	}

	return nil
}
