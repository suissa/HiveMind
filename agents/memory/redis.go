package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisMemoryManager gerencia memórias usando Redis
type RedisMemoryManager struct {
	client *redis.Client
}

// NewRedisMemoryManager cria um novo gerenciador de memória Redis
func NewRedisMemoryManager(redisURL string) (*RedisMemoryManager, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao analisar URL do Redis: %v", err)
	}

	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("erro ao conectar ao Redis: %v", err)
	}

	return &RedisMemoryManager{
		client: client,
	}, nil
}

// StoreMemory armazena uma memória no Redis
func (m *RedisMemoryManager) StoreMemory(ctx context.Context, memory *Memory) error {
	key := fmt.Sprintf("memory:%s:%s", memory.AgentID, memory.ID)
	data, err := json.Marshal(memory)
	if err != nil {
		return fmt.Errorf("erro ao serializar memória: %v", err)
	}

	// Define o TTL padrão de 24 horas se não especificado
	ttl := memory.TTL
	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	err = m.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("erro ao armazenar memória: %v", err)
	}

	// Adiciona à lista de memórias do agente
	agentKey := fmt.Sprintf("agent:%s:memories", memory.AgentID)
	err = m.client.SAdd(ctx, agentKey, memory.ID).Err()
	if err != nil {
		return fmt.Errorf("erro ao adicionar à lista de memórias: %v", err)
	}

	// Adiciona índices para as tags
	for _, tag := range memory.Tags {
		tagKey := fmt.Sprintf("tag:%s:%s", memory.AgentID, tag)
		err = m.client.SAdd(ctx, tagKey, memory.ID).Err()
		if err != nil {
			return fmt.Errorf("erro ao adicionar índice de tag: %v", err)
		}
	}

	return nil
}

// GetMemory recupera uma memória específica
func (m *RedisMemoryManager) GetMemory(ctx context.Context, agentID, memoryID string) (*Memory, error) {
	key := fmt.Sprintf("memory:%s:%s", agentID, memoryID)
	data, err := m.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("memória não encontrada")
		}
		return nil, fmt.Errorf("erro ao recuperar memória: %v", err)
	}

	var memory Memory
	if err := json.Unmarshal(data, &memory); err != nil {
		return nil, fmt.Errorf("erro ao deserializar memória: %v", err)
	}

	return &memory, nil
}

// SearchMemories busca memórias por tags
func (m *RedisMemoryManager) SearchMemories(ctx context.Context, agentID string, tags []string) ([]*Memory, error) {
	var memoryIDs []string
	if len(tags) > 0 {
		// Busca a interseção de memórias com todas as tags
		var keys []string
		for _, tag := range tags {
			keys = append(keys, fmt.Sprintf("tag:%s:%s", agentID, tag))
		}
		memoryIDs, _ = m.client.SInter(ctx, keys...).Result()
	} else {
		// Se não houver tags, retorna todas as memórias do agente
		agentKey := fmt.Sprintf("agent:%s:memories", agentID)
		memoryIDs, _ = m.client.SMembers(ctx, agentKey).Result()
	}

	var memories []*Memory
	for _, id := range memoryIDs {
		memory, err := m.GetMemory(ctx, agentID, id)
		if err == nil {
			memories = append(memories, memory)
		}
	}

	return memories, nil
}

// UpdateMemory atualiza uma memória existente
func (m *RedisMemoryManager) UpdateMemory(ctx context.Context, memory *Memory) error {
	// Recupera o TTL restante da memória existente
	key := fmt.Sprintf("memory:%s:%s", memory.AgentID, memory.ID)
	ttl, err := m.client.TTL(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("erro ao recuperar TTL: %v", err)
	}

	// Se a memória não existe ou expirou, retorna erro
	if ttl < 0 {
		return fmt.Errorf("memória não encontrada ou expirada")
	}

	// Atualiza a memória mantendo o TTL original
	data, err := json.Marshal(memory)
	if err != nil {
		return fmt.Errorf("erro ao serializar memória: %v", err)
	}

	err = m.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("erro ao atualizar memória: %v", err)
	}

	return nil
}

// DeleteMemory remove uma memória
func (m *RedisMemoryManager) DeleteMemory(ctx context.Context, agentID, memoryID string) error {
	key := fmt.Sprintf("memory:%s:%s", agentID, memoryID)

	// Remove a memória
	err := m.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("erro ao deletar memória: %v", err)
	}

	// Remove da lista de memórias do agente
	agentKey := fmt.Sprintf("agent:%s:memories", agentID)
	err = m.client.SRem(ctx, agentKey, memoryID).Err()
	if err != nil {
		return fmt.Errorf("erro ao remover da lista de memórias: %v", err)
	}

	return nil
}

// PruneMemories remove memórias expiradas
func (m *RedisMemoryManager) PruneMemories(ctx context.Context, agentID string) error {
	// O Redis já remove automaticamente as chaves expiradas
	// Esta função é mantida para compatibilidade com a interface
	return nil
}

// Close fecha a conexão com o Redis
func (m *RedisMemoryManager) Close(ctx context.Context) error {
	if err := m.client.Close(); err != nil {
		return fmt.Errorf("erro ao fechar conexão com Redis: %v", err)
	}
	return nil
}
