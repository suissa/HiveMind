# Clientes de Comunicação

Este pacote fornece implementações de clientes de comunicação para NATS e WebSocket, permitindo a troca de mensagens entre agentes do HiveMind.

## Características

- Interface comum para diferentes protocolos de comunicação
- Suporte para publicação/inscrição de mensagens
- Suporte para requisições síncronas (request/reply)
- Monitoramento de status da conexão
- Gerenciamento automático de reconexão
- Suporte para TLS e autenticação

## Uso

### Configuração

```go
config := &communication.ConnectionConfig{
    Host:     "localhost",
    Port:     4222,      // Para NATS
    // Port:  8080,      // Para WebSocket
    Username: "user",    // Opcional
    Password: "pass",    // Opcional
    TLS:     false,      // Opcional
}
```

### Criando um Cliente

```go
// Cliente NATS
natsClient := communication.NewNatsClient(config)

// Cliente WebSocket
wsClient := communication.NewWebSocketClient(config)
```

### Conectando

```go
ctx := context.Background()
err := client.Connect(ctx)
if err != nil {
    log.Fatalf("Erro ao conectar: %v", err)
}
defer client.Disconnect()
```

### Publicando Mensagens

```go
err := client.Publish(ctx, "topic.name", []byte("mensagem"))
if err != nil {
    log.Printf("Erro ao publicar: %v", err)
}
```

### Inscrevendo-se em Tópicos

```go
handler := func(ctx context.Context, subject string, data []byte) error {
    fmt.Printf("Mensagem recebida em %s: %s\n", subject, string(data))
    return nil
}

err := client.Subscribe("topic.name", handler)
if err != nil {
    log.Printf("Erro ao se inscrever: %v", err)
}
```

### Fazendo Requisições

```go
response, err := client.Request(ctx, "service.name", []byte("request"), 1000) // timeout em ms
if err != nil {
    log.Printf("Erro na requisição: %v", err)
} else {
    fmt.Printf("Resposta: %s\n", string(response))
}
```

### Monitorando Status

```go
status := client.GetStatus()
fmt.Printf("Conectado: %v\n", status.Connected)
fmt.Printf("Última conexão: %v\n", time.Unix(status.LastConnection, 0))
fmt.Printf("Último erro: %v\n", status.LastError)
fmt.Printf("Bytes enviados: %d\n", status.BytesSent)
fmt.Printf("Bytes recebidos: %d\n", status.BytesReceived)
fmt.Printf("Inscrições: %d\n", status.Subscriptions)
```

## Exemplo Completo

Veja o arquivo `examples/communication/main.go` para um exemplo completo de uso dos clientes.

## Considerações de Uso

1. **Tratamento de Erros**: Sempre verifique os erros retornados pelos métodos.
2. **Contexto**: Use contexto para controle de timeout e cancelamento.
3. **Desconexão**: Sempre chame `Disconnect()` ao finalizar o uso do cliente.
4. **Concorrência**: Os clientes são thread-safe e podem ser usados em goroutines.
5. **Reconexão**: Os clientes tentarão reconectar automaticamente em caso de falha.

## Dependências

- github.com/nats-io/nats.go v1.33.1
- github.com/gorilla/websocket v1.5.1 