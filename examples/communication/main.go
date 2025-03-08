package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"HiveMind/agents/communication"
)

func main() {
	// Configurações dos clientes
	natsConfig := &communication.ConnectionConfig{
		Host: "localhost",
		Port: 4222,
	}

	kafkaConfig := &communication.ConnectionConfig{
		Host: "localhost",
		Port: 9092,
	}

	grpcConfig := &communication.ConnectionConfig{
		Host: "localhost",
		Port: 50051,
	}

	// Cria os clientes
	natsClient := communication.NewNatsClient(natsConfig)
	kafkaClient := communication.NewKafkaClient(kafkaConfig, "test-group")
	grpcClient := communication.NewGRPCClient(grpcConfig)

	// Contexto com cancelamento
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Conecta os clientes
	if err := natsClient.Connect(ctx); err != nil {
		log.Printf("Erro ao conectar ao NATS: %v", err)
	} else {
		defer natsClient.Disconnect()
	}

	if err := kafkaClient.Connect(ctx); err != nil {
		log.Printf("Erro ao conectar ao Kafka: %v", err)
	} else {
		defer kafkaClient.Disconnect()
	}

	if err := grpcClient.Connect(ctx); err != nil {
		log.Printf("Erro ao conectar ao gRPC: %v", err)
	} else {
		defer grpcClient.Disconnect()
	}

	// Cria e inicia o observer
	observer, err := communication.NewObserver("http://localhost:8082")
	if err != nil {
		log.Fatalf("Erro ao criar observer: %v", err)
	}
	defer observer.Close()

	// Adiciona os clientes ao observer
	observer.AddClient(natsClient)
	observer.AddClient(kafkaClient)
	observer.AddClient(grpcClient)

	// Inicia o observer
	if err := observer.Start(ctx); err != nil {
		log.Fatalf("Erro ao iniciar observer: %v", err)
	}

	// Handler para mensagens
	messageHandler := func(ctx context.Context, subject string, data []byte) error {
		fmt.Printf("Mensagem recebida no tópico %s: %s\n", subject, string(data))
		return nil
	}

	// Lista de clientes ativos
	var activeClients []communication.CommunicationClient
	if natsClient.GetStatus().Connected {
		activeClients = append(activeClients, natsClient)
	}
	if kafkaClient.GetStatus().Connected {
		activeClients = append(activeClients, kafkaClient)
	}
	if grpcClient.GetStatus().Connected {
		activeClients = append(activeClients, grpcClient)
	}

	// Inscreve-se em tópicos para cada cliente ativo
	for _, client := range activeClients {
		if err := client.Subscribe("test.topic", messageHandler); err != nil {
			log.Printf("Erro ao se inscrever no tópico: %v", err)
		}
	}

	// Publica mensagens periodicamente
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				message := fmt.Sprintf("Teste %d", time.Now().Unix())

				// Publica em todos os clientes ativos
				for _, client := range activeClients {
					if err := client.Publish(ctx, "test.topic", []byte(message)); err != nil {
						log.Printf("Erro ao publicar mensagem: %v", err)
					}

					// Faz uma requisição
					response, err := client.Request(ctx, "test.request", []byte("ping"), 1000)
					if err != nil {
						log.Printf("Erro na requisição: %v", err)
					} else {
						fmt.Printf("Resposta: %s\n", string(response))
					}
				}
			}
		}
	}()

	// Monitora o status dos clientes
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fmt.Println("\n=== Status dos Clientes ===")

				if natsClient.GetStatus().Connected {
					status := natsClient.GetStatus()
					fmt.Printf("\nNATS:\n")
					fmt.Printf("Conectado: %v\n", status.Connected)
					fmt.Printf("Última conexão: %v\n", time.Unix(status.LastConnection, 0))
					fmt.Printf("Último erro: %v\n", status.LastError)
					fmt.Printf("Bytes enviados: %d\n", status.BytesSent)
					fmt.Printf("Bytes recebidos: %d\n", status.BytesReceived)
					fmt.Printf("Inscrições: %d\n", status.Subscriptions)
				}

				if kafkaClient.GetStatus().Connected {
					status := kafkaClient.GetStatus()
					fmt.Printf("\nKafka:\n")
					fmt.Printf("Conectado: %v\n", status.Connected)
					fmt.Printf("Última conexão: %v\n", time.Unix(status.LastConnection, 0))
					fmt.Printf("Último erro: %v\n", status.LastError)
					fmt.Printf("Bytes enviados: %d\n", status.BytesSent)
					fmt.Printf("Bytes recebidos: %d\n", status.BytesReceived)
					fmt.Printf("Inscrições: %d\n", status.Subscriptions)
				}

				if grpcClient.GetStatus().Connected {
					status := grpcClient.GetStatus()
					fmt.Printf("\ngRPC:\n")
					fmt.Printf("Conectado: %v\n", status.Connected)
					fmt.Printf("Última conexão: %v\n", time.Unix(status.LastConnection, 0))
					fmt.Printf("Último erro: %v\n", status.LastError)
					fmt.Printf("Bytes enviados: %d\n", status.BytesSent)
					fmt.Printf("Bytes recebidos: %d\n", status.BytesReceived)
					fmt.Printf("Inscrições: %d\n", status.Subscriptions)
				}
			}
		}
	}()

	// Aguarda sinal de interrupção
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nEncerrando...")
}
