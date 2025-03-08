package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// EventLog representa um evento de mensagem padronizado
type EventLog struct {
	Timestamp   int64                  `json:"timestamp"`
	Protocol    string                 `json:"protocol"` // NATS, Kafka ou gRPC
	Type        string                 `json:"type"`     // publish, subscribe, request, response
	Subject     string                 `json:"subject"`
	Data        []byte                 `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
	Status      string                 `json:"status"` // success, error
	Error       string                 `json:"error,omitempty"`
	ProcessedAt int64                  `json:"processed_at"`
}

// Observer monitora todas as filas e persiste eventos
type Observer struct {
	clients    []CommunicationClient
	logger     *log.Logger
	druidConn  *avatica.Connection
	eventsChan chan *EventLog
	mu         sync.RWMutex
}

// NewObserver cria uma nova instância do Observer
func NewObserver(druidURL string) (*Observer, error) {
	// Configura logger com formato personalizado
	logger := log.New(os.Stdout, "[OBSERVER] ", log.Ldate|log.Ltime|log.LUTC)

	// Conecta ao Apache Druid via Avatica
	conn, err := avatica.NewConnection(druidURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao Druid: %v", err)
	}

	// Cria tabela de eventos se não existir
	if err := createEventsTable(conn); err != nil {
		conn.Close()
		return nil, err
	}

	return &Observer{
		clients:    make([]CommunicationClient, 0),
		logger:     logger,
		druidConn:  conn,
		eventsChan: make(chan *EventLog, 1000),
	}, nil
}

// AddClient adiciona um cliente para ser monitorado
func (o *Observer) AddClient(client CommunicationClient) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.clients = append(o.clients, client)
}

// Start inicia o monitoramento
func (o *Observer) Start(ctx context.Context) error {
	// Handler para todas as mensagens
	handler := func(ctx context.Context, subject string, data []byte) error {
		event := &EventLog{
			Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
			Type:      "message",
			Subject:   subject,
			Data:      data,
			Status:    "success",
		}
		o.eventsChan <- event
		return nil
	}

	// Inscreve em todos os tópicos de todos os clientes
	for _, client := range o.clients {
		status := client.GetStatus()
		if !status.Connected {
			continue
		}

		// Identifica o protocolo baseado no tipo do cliente
		protocol := "unknown"
		switch client.(type) {
		case *NatsClient:
			protocol = "NATS"
		case *KafkaClient:
			protocol = "Kafka"
		case *GRPCClient:
			protocol = "gRPC"
		}

		// Inscreve em todos os tópicos ativos
		subs := client.GetSubscriptions()
		for _, topic := range subs {
			if err := client.Subscribe(topic, handler); err != nil {
				o.logger.Printf("Erro ao se inscrever no tópico %s: %v", topic, err)
			}
		}

		// Monitora novos tópicos
		go o.monitorNewTopics(ctx, client, protocol, handler)
	}

	// Processa e persiste eventos
	go o.processEvents(ctx)

	return nil
}

// monitorNewTopics monitora novos tópicos que surgem
func (o *Observer) monitorNewTopics(ctx context.Context, client CommunicationClient, protocol string, handler MessageHandler) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	knownTopics := make(map[string]bool)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			subs := client.GetSubscriptions()
			for _, topic := range subs {
				if !knownTopics[topic] {
					if err := client.Subscribe(topic, handler); err != nil {
						o.logger.Printf("Erro ao se inscrever no novo tópico %s: %v", topic, err)
						continue
					}
					knownTopics[topic] = true
					o.logger.Printf("Novo tópico detectado e monitorado: %s (%s)", topic, protocol)
				}
			}
		}
	}
}

// processEvents processa e persiste eventos
func (o *Observer) processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-o.eventsChan:
			// Log no terminal
			eventJSON, _ := json.MarshalIndent(event, "", "  ")
			o.logger.Printf("Evento recebido:\n%s\n", string(eventJSON))

			// Persiste no Druid
			if err := o.persistEvent(event); err != nil {
				o.logger.Printf("Erro ao persistir evento: %v", err)
			}
		}
	}
}

// persistEvent persiste um evento no Druid
func (o *Observer) persistEvent(event *EventLog) error {
	query := `
		INSERT INTO events (
			timestamp,
			protocol,
			type,
			subject,
			data,
			metadata,
			status,
			error,
			processed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("erro ao serializar metadata: %v", err)
	}

	_, err = o.druidConn.Exec(
		query,
		event.Timestamp,
		event.Protocol,
		event.Type,
		event.Subject,
		event.Data,
		string(metadataJSON),
		event.Status,
		event.Error,
		event.ProcessedAt,
	)

	if err != nil {
		return fmt.Errorf("erro ao inserir evento: %v", err)
	}

	return nil
}

// createEventsTable cria a tabela de eventos no Druid
func createEventsTable(conn *avatica.Connection) error {
	query := `
		CREATE TABLE IF NOT EXISTS events (
			timestamp BIGINT,
			protocol VARCHAR,
			type VARCHAR,
			subject VARCHAR,
			data VARBINARY,
			metadata VARCHAR,
			status VARCHAR,
			error VARCHAR,
			processed_at BIGINT
		) PARTITIONED BY (
			FLOOR(timestamp TO DAY)
		)
		STORED AS json
		TBLPROPERTIES (
			"druid.segment.granularity" = "DAY",
			"druid.query.granularity" = "SECOND"
		)
	`

	_, err := conn.Exec(query)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela de eventos: %v", err)
	}

	return nil
}

// Close fecha o observer e suas conexões
func (o *Observer) Close() error {
	if o.druidConn != nil {
		return o.druidConn.Close()
	}
	return nil
}
