package communication

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

// KafkaClient implementa a interface CommunicationClient usando Kafka
type KafkaClient struct {
	producer    sarama.SyncProducer
	consumer    sarama.ConsumerGroup
	config      *ConnectionConfig
	status      *ClientStatus
	handlers    map[string]MessageHandler
	groupID     string
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	consumeWait sync.WaitGroup
}

// NewKafkaClient cria uma nova instância do cliente Kafka
func NewKafkaClient(config *ConnectionConfig, groupID string) *KafkaClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &KafkaClient{
		config:   config,
		groupID:  groupID,
		status:   &ClientStatus{Connected: false},
		handlers: make(map[string]MessageHandler),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Connect estabelece a conexão com o servidor Kafka
func (kc *KafkaClient) Connect(ctx context.Context) error {
	// Configuração do Kafka
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	// Configura autenticação se necessário
	if kc.config.Username != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = kc.config.Username
		config.Net.SASL.Password = kc.config.Password
	}

	// Configura TLS se necessário
	if kc.config.TLS {
		config.Net.TLS.Enable = true
	}

	// Cria o produtor
	brokers := []string{fmt.Sprintf("%s:%d", kc.config.Host, kc.config.Port)}
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return fmt.Errorf("erro ao criar produtor Kafka: %v", err)
	}
	kc.producer = producer

	// Cria o consumidor
	consumer, err := sarama.NewConsumerGroup(brokers, kc.groupID, config)
	if err != nil {
		kc.producer.Close()
		return fmt.Errorf("erro ao criar consumidor Kafka: %v", err)
	}
	kc.consumer = consumer

	kc.status.Connected = true
	kc.status.LastConnection = time.Now().Unix()

	return nil
}

// Disconnect fecha a conexão com o servidor Kafka
func (kc *KafkaClient) Disconnect() error {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	kc.cancel()
	kc.consumeWait.Wait()

	if kc.producer != nil {
		if err := kc.producer.Close(); err != nil {
			return fmt.Errorf("erro ao fechar produtor Kafka: %v", err)
		}
	}

	if kc.consumer != nil {
		if err := kc.consumer.Close(); err != nil {
			return fmt.Errorf("erro ao fechar consumidor Kafka: %v", err)
		}
	}

	kc.status.Connected = false
	return nil
}

// consumerHandler implementa a interface sarama.ConsumerGroupHandler
type consumerHandler struct {
	handlers map[string]MessageHandler
	mu       sync.RWMutex
	status   *ClientStatus
}

func (h *consumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		h.mu.RLock()
		handler, exists := h.handlers[message.Topic]
		h.mu.RUnlock()

		if exists && handler != nil {
			if err := handler(session.Context(), message.Topic, message.Value); err != nil {
				h.mu.Lock()
				h.status.LastError = err.Error()
				h.mu.Unlock()
			}
		}

		h.mu.Lock()
		h.status.BytesReceived += int64(len(message.Value))
		h.mu.Unlock()

		session.MarkMessage(message, "")
	}
	return nil
}

// Subscribe registra um handler para receber mensagens de um tópico
func (kc *KafkaClient) Subscribe(subject string, handler MessageHandler) error {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	if _, exists := kc.handlers[subject]; exists {
		return fmt.Errorf("já existe um handler para o tópico %s", subject)
	}

	kc.handlers[subject] = handler
	kc.status.Subscriptions++

	// Inicia uma goroutine para consumir mensagens
	kc.consumeWait.Add(1)
	go func() {
		defer kc.consumeWait.Done()
		h := &consumerHandler{
			handlers: kc.handlers,
			status:   kc.status,
		}

		for {
			select {
			case <-kc.ctx.Done():
				return
			default:
				if err := kc.consumer.Consume(kc.ctx, []string{subject}, h); err != nil {
					kc.mu.Lock()
					kc.status.LastError = err.Error()
					kc.mu.Unlock()
					time.Sleep(time.Second) // Espera antes de tentar novamente
				}
			}
		}
	}()

	return nil
}

// Unsubscribe remove a inscrição de um tópico
func (kc *KafkaClient) Unsubscribe(subject string) error {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	if _, exists := kc.handlers[subject]; !exists {
		return fmt.Errorf("não existe handler para o tópico %s", subject)
	}

	delete(kc.handlers, subject)
	kc.status.Subscriptions--

	return nil
}

// Publish envia uma mensagem para um tópico
func (kc *KafkaClient) Publish(ctx context.Context, subject string, data []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: subject,
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := kc.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("erro ao publicar mensagem no tópico %s: %v", subject, err)
	}

	kc.mu.Lock()
	kc.status.BytesSent += int64(len(data))
	kc.mu.Unlock()

	// Log opcional para debug
	fmt.Printf("Mensagem enviada para tópico=%s partition=%d offset=%d\n",
		subject, partition, offset)

	return nil
}

// Request envia uma mensagem e aguarda resposta
// Nota: Kafka não tem suporte nativo para request/reply, então implementamos usando tópicos temporários
func (kc *KafkaClient) Request(ctx context.Context, subject string, data []byte, timeout int) ([]byte, error) {
	// Cria um tópico temporário para a resposta
	replyTopic := fmt.Sprintf("%s.reply.%d", subject, time.Now().UnixNano())
	responseChan := make(chan []byte, 1)
	errorChan := make(chan error, 1)

	// Inscreve-se no tópico de resposta
	handler := func(ctx context.Context, subject string, data []byte) error {
		select {
		case responseChan <- data:
		default:
		}
		return nil
	}

	if err := kc.Subscribe(replyTopic, handler); err != nil {
		return nil, fmt.Errorf("erro ao se inscrever no tópico de resposta: %v", err)
	}
	defer kc.Unsubscribe(replyTopic)

	// Adiciona o tópico de resposta à mensagem
	requestMsg := &sarama.ProducerMessage{
		Topic: subject,
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("reply_to"),
				Value: []byte(replyTopic),
			},
		},
		Value: sarama.ByteEncoder(data),
	}

	// Envia a requisição
	if _, _, err := kc.producer.SendMessage(requestMsg); err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %v", err)
	}

	kc.mu.Lock()
	kc.status.BytesSent += int64(len(data))
	kc.mu.Unlock()

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
func (kc *KafkaClient) GetStatus() *ClientStatus {
	kc.mu.RLock()
	defer kc.mu.RUnlock()
	return kc.status
}

// GetSubscriptions retorna a lista de inscrições ativas
func (kc *KafkaClient) GetSubscriptions() []string {
	kc.mu.RLock()
	defer kc.mu.RUnlock()

	subs := make([]string, 0, len(kc.handlers))
	for subject := range kc.handlers {
		subs = append(subs, subject)
	}
	return subs
}
