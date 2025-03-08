package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketClient implementa a interface CommunicationClient usando WebSocket
type WebSocketClient struct {
	conn     *websocket.Conn
	config   *ConnectionConfig
	status   *ClientStatus
	handlers map[string]MessageHandler
	mu       sync.RWMutex
	done     chan struct{}
}

// NewWebSocketClient cria uma nova instância do cliente WebSocket
func NewWebSocketClient(config *ConnectionConfig) *WebSocketClient {
	return &WebSocketClient{
		config: config,
		status: &ClientStatus{
			Connected: false,
		},
		handlers: make(map[string]MessageHandler),
		done:     make(chan struct{}),
	}
}

// Connect estabelece a conexão com o servidor WebSocket
func (wc *WebSocketClient) Connect(ctx context.Context) error {
	scheme := "ws"
	if wc.config.TLS {
		scheme = "wss"
	}

	u := url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%d", wc.config.Host, wc.config.Port),
		Path:   "/ws",
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	if wc.config.Username != "" {
		u.User = url.UserPassword(wc.config.Username, wc.config.Password)
	}

	conn, _, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao WebSocket: %v", err)
	}

	wc.conn = conn
	wc.status.Connected = true
	wc.status.LastConnection = time.Now().Unix()

	// Inicia goroutine para leitura de mensagens
	go wc.readPump()

	return nil
}

// readPump processa mensagens recebidas do WebSocket
func (wc *WebSocketClient) readPump() {
	defer func() {
		wc.conn.Close()
		wc.status.Connected = false
	}()

	for {
		select {
		case <-wc.done:
			return
		default:
			_, message, err := wc.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					wc.status.LastError = err.Error()
				}
				return
			}

			// Processa a mensagem recebida
			var msg struct {
				Subject string          `json:"subject"`
				Data    json.RawMessage `json:"data"`
			}

			if err := json.Unmarshal(message, &msg); err != nil {
				wc.status.LastError = fmt.Sprintf("erro ao decodificar mensagem: %v", err)
				continue
			}

			wc.mu.RLock()
			handler, exists := wc.handlers[msg.Subject]
			wc.mu.RUnlock()

			if exists && handler != nil {
				ctx := context.Background()
				if err := handler(ctx, msg.Subject, []byte(msg.Data)); err != nil {
					wc.status.LastError = fmt.Sprintf("erro ao processar mensagem do tópico %s: %v", msg.Subject, err)
				}
			}

			wc.mu.Lock()
			wc.status.BytesReceived += int64(len(message))
			wc.mu.Unlock()
		}
	}
}

// Disconnect fecha a conexão com o servidor WebSocket
func (wc *WebSocketClient) Disconnect() error {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	close(wc.done)
	if wc.conn != nil {
		return wc.conn.Close()
	}
	return nil
}

// Subscribe registra um handler para receber mensagens de um tópico
func (wc *WebSocketClient) Subscribe(subject string, handler MessageHandler) error {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	if _, exists := wc.handlers[subject]; exists {
		return fmt.Errorf("já existe um handler para o tópico %s", subject)
	}

	// Envia mensagem de inscrição para o servidor
	msg := struct {
		Action  string `json:"action"`
		Subject string `json:"subject"`
	}{
		Action:  "subscribe",
		Subject: subject,
	}

	if err := wc.conn.WriteJSON(msg); err != nil {
		return fmt.Errorf("erro ao enviar inscrição para o tópico %s: %v", subject, err)
	}

	wc.handlers[subject] = handler
	wc.status.Subscriptions++

	return nil
}

// Unsubscribe remove a inscrição de um tópico
func (wc *WebSocketClient) Unsubscribe(subject string) error {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	if _, exists := wc.handlers[subject]; !exists {
		return fmt.Errorf("não existe handler para o tópico %s", subject)
	}

	// Envia mensagem de cancelamento de inscrição para o servidor
	msg := struct {
		Action  string `json:"action"`
		Subject string `json:"subject"`
	}{
		Action:  "unsubscribe",
		Subject: subject,
	}

	if err := wc.conn.WriteJSON(msg); err != nil {
		return fmt.Errorf("erro ao cancelar inscrição do tópico %s: %v", subject, err)
	}

	delete(wc.handlers, subject)
	wc.status.Subscriptions--

	return nil
}

// Publish envia uma mensagem para um tópico
func (wc *WebSocketClient) Publish(ctx context.Context, subject string, data []byte) error {
	msg := struct {
		Action  string          `json:"action"`
		Subject string          `json:"subject"`
		Data    json.RawMessage `json:"data"`
	}{
		Action:  "publish",
		Subject: subject,
		Data:    data,
	}

	if err := wc.conn.WriteJSON(msg); err != nil {
		return fmt.Errorf("erro ao publicar mensagem no tópico %s: %v", subject, err)
	}

	wc.mu.Lock()
	wc.status.BytesSent += int64(len(data))
	wc.mu.Unlock()

	return nil
}

// Request envia uma mensagem e aguarda resposta
func (wc *WebSocketClient) Request(ctx context.Context, subject string, data []byte, timeout int) ([]byte, error) {
	responseChan := make(chan []byte, 1)
	errorChan := make(chan error, 1)

	// Gera um ID único para a requisição
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Registra um handler temporário para receber a resposta
	wc.mu.Lock()
	wc.handlers[requestID] = func(ctx context.Context, subject string, data []byte) error {
		select {
		case responseChan <- data:
		default:
		}
		return nil
	}
	wc.mu.Unlock()

	// Remove o handler temporário ao finalizar
	defer func() {
		wc.mu.Lock()
		delete(wc.handlers, requestID)
		wc.mu.Unlock()
	}()

	// Envia a requisição
	msg := struct {
		Action    string          `json:"action"`
		Subject   string          `json:"subject"`
		Data      json.RawMessage `json:"data"`
		RequestID string          `json:"request_id"`
	}{
		Action:    "request",
		Subject:   subject,
		Data:      data,
		RequestID: requestID,
	}

	if err := wc.conn.WriteJSON(msg); err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição para o tópico %s: %v", subject, err)
	}

	wc.mu.Lock()
	wc.status.BytesSent += int64(len(data))
	wc.mu.Unlock()

	// Aguarda a resposta com timeout
	select {
	case response := <-responseChan:
		return response, nil
	case err := <-errorChan:
		return nil, err
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		return nil, fmt.Errorf("timeout ao aguardar resposta do tópico %s", subject)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetStatus retorna o estado atual do cliente
func (wc *WebSocketClient) GetStatus() *ClientStatus {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	return wc.status
}

// GetSubscriptions retorna a lista de inscrições ativas
func (wc *WebSocketClient) GetSubscriptions() []string {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	subs := make([]string, 0, len(wc.handlers))
	for subject := range wc.handlers {
		subs = append(subs, subject)
	}
	return subs
}
