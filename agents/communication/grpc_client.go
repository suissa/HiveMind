package communication

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClient implementa a interface CommunicationClient usando gRPC
type GRPCClient struct {
	conn     *grpc.ClientConn
	client   MessagingClient
	config   *ConnectionConfig
	status   *ClientStatus
	handlers map[string]MessageHandler
	mu       sync.RWMutex
	stream   Messaging_SubscribeClient
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewGRPCClient cria uma nova instância do cliente gRPC
func NewGRPCClient(config *ConnectionConfig) *GRPCClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &GRPCClient{
		config:   config,
		status:   &ClientStatus{Connected: false},
		handlers: make(map[string]MessageHandler),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Connect estabelece a conexão com o servidor gRPC
func (gc *GRPCClient) Connect(ctx context.Context) error {
	var opts []grpc.DialOption

	// Configura TLS se necessário
	if gc.config.TLS {
		creds, err := credentials.NewClientTLSFromFile("cert.pem", "")
		if err != nil {
			return fmt.Errorf("erro ao carregar certificado TLS: %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Configura autenticação se necessário
	if gc.config.Username != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(&authCreds{
			username: gc.config.Username,
			password: gc.config.Password,
		}))
	}

	// Estabelece a conexão
	addr := fmt.Sprintf("%s:%d", gc.config.Host, gc.config.Port)
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao servidor gRPC: %v", err)
	}

	gc.conn = conn
	gc.client = NewMessagingClient(conn)
	gc.status.Connected = true
	gc.status.LastConnection = time.Now().Unix()

	// Inicia o stream de mensagens
	stream, err := gc.client.Subscribe(gc.ctx, &SubscribeRequest{})
	if err != nil {
		gc.conn.Close()
		return fmt.Errorf("erro ao iniciar stream de mensagens: %v", err)
	}
	gc.stream = stream

	// Inicia goroutine para processar mensagens recebidas
	go gc.processMessages()

	return nil
}

// processMessages processa mensagens recebidas do stream
func (gc *GRPCClient) processMessages() {
	for {
		select {
		case <-gc.ctx.Done():
			return
		default:
			msg, err := gc.stream.Recv()
			if err != nil {
				gc.mu.Lock()
				gc.status.LastError = err.Error()
				gc.mu.Unlock()
				time.Sleep(time.Second) // Espera antes de tentar novamente
				continue
			}

			gc.mu.RLock()
			handler, exists := gc.handlers[msg.Subject]
			gc.mu.RUnlock()

			if exists && handler != nil {
				if err := handler(gc.ctx, msg.Subject, msg.Data); err != nil {
					gc.mu.Lock()
					gc.status.LastError = err.Error()
					gc.mu.Unlock()
				}
			}

			gc.mu.Lock()
			gc.status.BytesReceived += int64(len(msg.Data))
			gc.mu.Unlock()
		}
	}
}

// Disconnect fecha a conexão com o servidor gRPC
func (gc *GRPCClient) Disconnect() error {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	gc.cancel()

	if gc.conn != nil {
		if err := gc.conn.Close(); err != nil {
			return fmt.Errorf("erro ao fechar conexão gRPC: %v", err)
		}
	}

	gc.status.Connected = false
	return nil
}

// Subscribe registra um handler para receber mensagens de um tópico
func (gc *GRPCClient) Subscribe(subject string, handler MessageHandler) error {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	if _, exists := gc.handlers[subject]; exists {
		return fmt.Errorf("já existe um handler para o tópico %s", subject)
	}

	// Envia requisição de inscrição
	req := &SubscribeRequest{
		Subject: subject,
	}

	if _, err := gc.client.AddSubscription(gc.ctx, req); err != nil {
		return fmt.Errorf("erro ao se inscrever no tópico %s: %v", subject, err)
	}

	gc.handlers[subject] = handler
	gc.status.Subscriptions++

	return nil
}

// Unsubscribe remove a inscrição de um tópico
func (gc *GRPCClient) Unsubscribe(subject string) error {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	if _, exists := gc.handlers[subject]; !exists {
		return fmt.Errorf("não existe handler para o tópico %s", subject)
	}

	// Envia requisição de cancelamento de inscrição
	req := &UnsubscribeRequest{
		Subject: subject,
	}

	if _, err := gc.client.RemoveSubscription(gc.ctx, req); err != nil {
		return fmt.Errorf("erro ao cancelar inscrição do tópico %s: %v", subject, err)
	}

	delete(gc.handlers, subject)
	gc.status.Subscriptions--

	return nil
}

// Publish envia uma mensagem para um tópico
func (gc *GRPCClient) Publish(ctx context.Context, subject string, data []byte) error {
	msg := &Message{
		Subject:   subject,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	if _, err := gc.client.Publish(ctx, msg); err != nil {
		return fmt.Errorf("erro ao publicar mensagem no tópico %s: %v", subject, err)
	}

	gc.mu.Lock()
	gc.status.BytesSent += int64(len(data))
	gc.mu.Unlock()

	return nil
}

// Request envia uma mensagem e aguarda resposta
func (gc *GRPCClient) Request(ctx context.Context, subject string, data []byte, timeout int) ([]byte, error) {
	msg := &RequestMessage{
		Subject:   subject,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()

	response, err := gc.client.Request(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("erro na requisição para o tópico %s: %v", subject, err)
	}

	gc.mu.Lock()
	gc.status.BytesSent += int64(len(data))
	gc.status.BytesReceived += int64(len(response.Data))
	gc.mu.Unlock()

	return response.Data, nil
}

// GetStatus retorna o estado atual do cliente
func (gc *GRPCClient) GetStatus() *ClientStatus {
	gc.mu.RLock()
	defer gc.mu.RUnlock()
	return gc.status
}

// GetSubscriptions retorna a lista de inscrições ativas
func (gc *GRPCClient) GetSubscriptions() []string {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	subs := make([]string, 0, len(gc.handlers))
	for subject := range gc.handlers {
		subs = append(subs, subject)
	}
	return subs
}

// authCreds implementa as credenciais de autenticação para gRPC
type authCreds struct {
	username string
	password string
}

func (c *authCreds) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"username": c.username,
		"password": c.password,
	}, nil
}

func (c *authCreds) RequireTransportSecurity() bool {
	return true
}
