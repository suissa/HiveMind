package communication

import (
	"context"
	"testing"
	"time"
)

func TestNatsClient(t *testing.T) {
	config := &ConnectionConfig{
		Host: "localhost",
		Port: 4222,
	}

	client := NewNatsClient(config)
	ctx := context.Background()

	// Testa conexão
	if err := client.Connect(ctx); err != nil {
		t.Skipf("NATS não disponível: %v", err)
		return
	}
	defer client.Disconnect()

	// Testa status inicial
	status := client.GetStatus()
	if !status.Connected {
		t.Error("Cliente deveria estar conectado")
	}
	if status.LastConnection == 0 {
		t.Error("LastConnection deveria estar definido")
	}

	// Testa publicação e inscrição
	msgReceived := make(chan []byte, 1)
	handler := func(ctx context.Context, subject string, data []byte) error {
		msgReceived <- data
		return nil
	}

	if err := client.Subscribe("test.topic", handler); err != nil {
		t.Fatalf("Erro ao se inscrever: %v", err)
	}

	testMsg := []byte("teste")
	if err := client.Publish(ctx, "test.topic", testMsg); err != nil {
		t.Fatalf("Erro ao publicar: %v", err)
	}

	select {
	case received := <-msgReceived:
		if string(received) != string(testMsg) {
			t.Errorf("Mensagem recebida incorreta. Esperado: %s, Recebido: %s", testMsg, received)
		}
	case <-time.After(time.Second):
		t.Error("Timeout ao aguardar mensagem")
	}

	// Testa request/reply
	go func() {
		handler := func(ctx context.Context, subject string, data []byte) error {
			return client.Publish(ctx, subject, []byte("pong"))
		}
		if err := client.Subscribe("test.request", handler); err != nil {
			t.Errorf("Erro ao se inscrever para request: %v", err)
		}
	}()

	response, err := client.Request(ctx, "test.request", []byte("ping"), 1000)
	if err != nil {
		t.Fatalf("Erro na requisição: %v", err)
	}
	if string(response) != "pong" {
		t.Errorf("Resposta incorreta. Esperado: pong, Recebido: %s", string(response))
	}
}

func TestWebSocketClient(t *testing.T) {
	config := &ConnectionConfig{
		Host: "localhost",
		Port: 8080,
	}

	client := NewWebSocketClient(config)
	ctx := context.Background()

	// Testa conexão
	if err := client.Connect(ctx); err != nil {
		t.Skipf("WebSocket não disponível: %v", err)
		return
	}
	defer client.Disconnect()

	// Testa status inicial
	status := client.GetStatus()
	if !status.Connected {
		t.Error("Cliente deveria estar conectado")
	}
	if status.LastConnection == 0 {
		t.Error("LastConnection deveria estar definido")
	}

	// Testa publicação e inscrição
	msgReceived := make(chan []byte, 1)
	handler := func(ctx context.Context, subject string, data []byte) error {
		msgReceived <- data
		return nil
	}

	if err := client.Subscribe("test.topic", handler); err != nil {
		t.Fatalf("Erro ao se inscrever: %v", err)
	}

	testMsg := []byte("teste")
	if err := client.Publish(ctx, "test.topic", testMsg); err != nil {
		t.Fatalf("Erro ao publicar: %v", err)
	}

	select {
	case received := <-msgReceived:
		if string(received) != string(testMsg) {
			t.Errorf("Mensagem recebida incorreta. Esperado: %s, Recebido: %s", testMsg, received)
		}
	case <-time.After(time.Second):
		t.Error("Timeout ao aguardar mensagem")
	}

	// Testa request/reply
	go func() {
		handler := func(ctx context.Context, subject string, data []byte) error {
			return client.Publish(ctx, subject, []byte("pong"))
		}
		if err := client.Subscribe("test.request", handler); err != nil {
			t.Errorf("Erro ao se inscrever para request: %v", err)
		}
	}()

	response, err := client.Request(ctx, "test.request", []byte("ping"), 1000)
	if err != nil {
		t.Fatalf("Erro na requisição: %v", err)
	}
	if string(response) != "pong" {
		t.Errorf("Resposta incorreta. Esperado: pong, Recebido: %s", string(response))
	}
}

func TestConnectionConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   *ConnectionConfig
		wantHost string
		wantPort int
	}{
		{
			name: "Configuração básica",
			config: &ConnectionConfig{
				Host: "localhost",
				Port: 4222,
			},
			wantHost: "localhost",
			wantPort: 4222,
		},
		{
			name: "Configuração com credenciais",
			config: &ConnectionConfig{
				Host:     "example.com",
				Port:     8080,
				Username: "user",
				Password: "pass",
			},
			wantHost: "example.com",
			wantPort: 8080,
		},
		{
			name: "Configuração com TLS",
			config: &ConnectionConfig{
				Host: "secure.example.com",
				Port: 443,
				TLS:  true,
			},
			wantHost: "secure.example.com",
			wantPort: 443,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Host != tt.wantHost {
				t.Errorf("Host incorreto. Esperado: %s, Recebido: %s", tt.wantHost, tt.config.Host)
			}
			if tt.config.Port != tt.wantPort {
				t.Errorf("Porta incorreta. Esperado: %d, Recebido: %d", tt.wantPort, tt.config.Port)
			}
		})
	}
}
