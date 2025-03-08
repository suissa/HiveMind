package memory

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoMemoryManager gerencia memórias usando MongoDB
type MongoMemoryManager struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewMongoMemoryManager cria um novo gerenciador de memória MongoDB
func NewMongoMemoryManager(ctx context.Context, mongoURL, database, collection string) (*MongoMemoryManager, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao MongoDB: %v", err)
	}

	// Verifica a conexão
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar conexão com MongoDB: %v", err)
	}

	// Cria índices
	coll := client.Database(database).Collection(collection)
	_, err = coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "agent_id", Value: 1},
				{Key: "timestamp", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tags", Value: 1},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao criar índices: %v", err)
	}

	return &MongoMemoryManager{
		client:     client,
		collection: coll,
	}, nil
}

// StoreMemory armazena uma memória no MongoDB
func (m *MongoMemoryManager) StoreMemory(ctx context.Context, memory *Memory) error {
	memory.Timestamp = time.Now()
	_, err := m.collection.InsertOne(ctx, memory)
	if err != nil {
		return fmt.Errorf("erro ao armazenar memória: %v", err)
	}
	return nil
}

// GetMemory recupera uma memória específica
func (m *MongoMemoryManager) GetMemory(ctx context.Context, agentID, memoryID string) (*Memory, error) {
	var memory Memory
	err := m.collection.FindOne(ctx, bson.M{
		"_id":      memoryID,
		"agent_id": agentID,
	}).Decode(&memory)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("memória não encontrada")
		}
		return nil, fmt.Errorf("erro ao buscar memória: %v", err)
	}

	return &memory, nil
}

// SearchMemories busca memórias por tags
func (m *MongoMemoryManager) SearchMemories(ctx context.Context, agentID string, tags []string) ([]*Memory, error) {
	filter := bson.M{"agent_id": agentID}
	if len(tags) > 0 {
		filter["tags"] = bson.M{"$in": tags}
	}

	cursor, err := m.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar memórias: %v", err)
	}
	defer cursor.Close(ctx)

	var memories []*Memory
	for cursor.Next(ctx) {
		var memory Memory
		if err := cursor.Decode(&memory); err != nil {
			return nil, fmt.Errorf("erro ao decodificar memória: %v", err)
		}
		memories = append(memories, &memory)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar sobre resultados: %v", err)
	}

	return memories, nil
}

// UpdateMemory atualiza uma memória existente
func (m *MongoMemoryManager) UpdateMemory(ctx context.Context, memory *Memory) error {
	memory.Timestamp = time.Now()
	_, err := m.collection.UpdateOne(ctx,
		bson.M{"_id": memory.ID, "agent_id": memory.AgentID},
		bson.M{"$set": memory},
	)
	if err != nil {
		return fmt.Errorf("erro ao atualizar memória: %v", err)
	}
	return nil
}

// DeleteMemory remove uma memória
func (m *MongoMemoryManager) DeleteMemory(ctx context.Context, agentID, memoryID string) error {
	_, err := m.collection.DeleteOne(ctx, bson.M{
		"_id":      memoryID,
		"agent_id": agentID,
	})
	if err != nil {
		return fmt.Errorf("erro ao deletar memória: %v", err)
	}
	return nil
}

// PruneMemories remove memórias antigas
func (m *MongoMemoryManager) PruneMemories(ctx context.Context, agentID string) error {
	cutoff := time.Now().Add(-24 * time.Hour)
	_, err := m.collection.DeleteMany(ctx, bson.M{
		"agent_id":  agentID,
		"timestamp": bson.M{"$lt": cutoff},
	})
	if err != nil {
		return fmt.Errorf("erro ao limpar memórias antigas: %v", err)
	}
	return nil
}

// Close fecha a conexão com o MongoDB
func (m *MongoMemoryManager) Close(ctx context.Context) error {
	if err := m.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("erro ao fechar conexão com MongoDB: %v", err)
	}
	return nil
}
