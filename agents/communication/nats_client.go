package communication

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// NatsClient implementa a interface CommunicationClient usando NATS
type NatsClient struct {
	conn          *nats.Conn
	config        *ConnectionConfig
	status        *ClientStatus
	subscriptions map[string]*nats.Subscription
	handlers      map[string]MessageHandler
	mu            sync.RWMutex
}

// NewNatsClient cria uma nova instância do cliente NATS
func NewNatsClient(config *ConnectionConfig) *NatsClient {
	return &NatsClient{
		config: config,
		status: &ClientStatus{
			Connected: false,
		},
		subscriptions: make(map[string]*nats.Subscription),
		handlers:      make(map[string]MessageHandler),
	}
}

// Connect estabelece a conexão com o servidor NATS
func (nc *NatsClient) Connect(ctx context.Context) error {
	opts := []nats.Option{
		nats.Name("HiveMind NATS Client"),
		nats.ReconnectWait(time.Second * 5),
		nats.MaxReconnects(-1),
		nats.DisconnectHandler(func(_ *nats.Conn) {
			nc.mu.Lock()
			nc.status.Connected = false
			nc.mu.Unlock()
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			nc.mu.Lock()
			nc.status.Connected = true
			nc.status.LastConnection = time.Now().Unix()
			nc.mu.Unlock()
		}),
		nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
			nc.mu.Lock()
			nc.status.LastError = err.Error()
			nc.mu.Unlock()
		}),
	}

	if nc.config.Username != "" {
		opts = append(opts, nats.UserInfo(nc.config.Username, nc.config.Password))
	}

	if nc.config.TLS {
		opts = append(opts, nats.Secure())
	}

	url := fmt.Sprintf("nats://%s:%d", nc.config.Host, nc.config.Port)
	conn, err := nats.Connect(url, opts...)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao NATS: %v", err)
	}

	nc.conn = conn
	nc.status.Connected = true
	nc.status.LastConnection = time.Now().Unix()

	return nil
}

// Disconnect fecha a conexão com o servidor NATS
func (nc *NatsClient) Disconnect() error {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	if nc.conn != nil {
		nc.conn.Close()
		nc.status.Connected = false
	}

	return nil
}

// Subscribe registra um handler para receber mensagens de um tópico
func (nc *NatsClient) Subscribe(subject string, handler MessageHandler) error {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	if _, exists := nc.subscriptions[subject]; exists {
		return fmt.Errorf("já existe uma inscrição para o tópico %s", subject)
	}

	sub, err := nc.conn.Subscribe(subject, func(msg *nats.Msg) {
		if handler != nil {
			ctx := context.Background()
			if err := handler(ctx, msg.Subject, msg.Data); err != nil {
				// Log do erro ou tratamento adequado
				fmt.Printf("Erro ao processar mensagem do tópico %s: %v\n", subject, err)
			}
		}
	})

	if err != nil {
		return fmt.Errorf("erro ao se inscrever no tópico %s: %v", subject, err)
	}

	nc.subscriptions[subject] = sub
	nc.handlers[subject] = handler
	nc.status.Subscriptions++

	return nil
}

// Unsubscribe remove a inscrição de um tópico
func (nc *NatsClient) Unsubscribe(subject string) error {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	sub, exists := nc.subscriptions[subject]
	if !exists {
		return fmt.Errorf("não existe inscrição para o tópico %s", subject)
	}

	if err := sub.Unsubscribe(); err != nil {
		return fmt.Errorf("erro ao cancelar inscrição do tópico %s: %v", subject, err)
	}

	delete(nc.subscriptions, subject)
	delete(nc.handlers, subject)
	nc.status.Subscriptions--

	return nil
}

// Publish envia uma mensagem para um tópico
func (nc *NatsClient) Publish(ctx context.Context, subject string, data []byte) error {
	if err := nc.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("erro ao publicar mensagem no tópico %s: %v", subject, err)
	}

	nc.mu.Lock()
	nc.status.BytesSent += int64(len(data))
	nc.mu.Unlock()

	return nil
}

// Request envia uma mensagem e aguarda resposta
func (nc *NatsClient) Request(ctx context.Context, subject string, data []byte, timeout int) ([]byte, error) {
	msg, err := nc.conn.Request(subject, data, time.Duration(timeout)*time.Millisecond)
	if err != nil {
		if err == nats.ErrTimeout {
			return nil, fmt.Errorf("timeout ao aguardar resposta do tópico %s", subject)
		}
		return nil, fmt.Errorf("erro ao fazer request no tópico %s: %v", subject, err)
	}

	nc.mu.Lock()
	nc.status.BytesSent += int64(len(data))
	nc.status.BytesReceived += int64(len(msg.Data))
	nc.mu.Unlock()

	return msg.Data, nil
}

// GetStatus retorna o estado atual do cliente
func (nc *NatsClient) GetStatus() *ClientStatus {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return nc.status
}

// GetSubscriptions retorna a lista de inscrições ativas
func (nc *NatsClient) GetSubscriptions() []string {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	subs := make([]string, 0, len(nc.subscriptions))
	for subject := range nc.subscriptions {
		subs = append(subs, subject)
	}
	return subs
}
