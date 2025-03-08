package communication

import (
	"context"
)

// MessageHandler é a função que processa mensagens recebidas
type MessageHandler func(ctx context.Context, subject string, data []byte) error

// CommunicationClient define a interface base para comunicação
type CommunicationClient interface {
	// Connect estabelece a conexão com o servidor
	Connect(ctx context.Context) error
	// Disconnect fecha a conexão com o servidor
	Disconnect() error
	// Subscribe registra um handler para receber mensagens de um tópico
	Subscribe(subject string, handler MessageHandler) error
	// Unsubscribe remove a inscrição de um tópico
	Unsubscribe(subject string) error
	// Publish envia uma mensagem para um tópico
	Publish(ctx context.Context, subject string, data []byte) error
	// Request envia uma mensagem e aguarda resposta
	Request(ctx context.Context, subject string, data []byte, timeout int) ([]byte, error)
	// GetStatus retorna o estado atual do cliente
	GetStatus() *ClientStatus
	// GetSubscriptions retorna a lista de inscrições ativas
	GetSubscriptions() []string
}

// ConnectionConfig define as configurações de conexão
type ConnectionConfig struct {
	Host     string            // Endereço do servidor
	Port     int               // Porta do servidor
	Username string            // Nome de usuário (opcional)
	Password string            // Senha (opcional)
	TLS      bool              // Usar TLS
	Headers  map[string]string // Headers adicionais
}

// Message representa uma mensagem trocada entre os agentes
type Message struct {
	Subject   string                 `json:"subject"`   // Tópico da mensagem
	Data      []byte                 `json:"data"`      // Conteúdo da mensagem
	Metadata  map[string]interface{} `json:"metadata"`  // Metadados adicionais
	Timestamp int64                  `json:"timestamp"` // Timestamp da mensagem
	ID        string                 `json:"id"`        // ID único da mensagem
	ReplyTo   string                 `json:"reply_to"`  // Tópico para resposta (opcional)
}

// ClientStatus representa o estado atual do cliente
type ClientStatus struct {
	Connected      bool   `json:"connected"`       // Se está conectado
	LastError      string `json:"last_error"`      // Último erro ocorrido
	LastConnection int64  `json:"last_connection"` // Timestamp da última conexão
	Subscriptions  int    `json:"subscriptions"`   // Número de inscrições ativas
	BytesSent      int64  `json:"bytes_sent"`      // Total de bytes enviados
	BytesReceived  int64  `json:"bytes_received"`  // Total de bytes recebidos
}
