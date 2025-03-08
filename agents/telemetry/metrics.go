package telemetry

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

var (
	meter metric.Meter

	// Métricas do agente
	taskProcessingDuration metric.Float64Histogram
	taskSuccessCounter     metric.Int64Counter
	taskFailureCounter     metric.Int64Counter
	activeTasksGauge       metric.Int64UpDownCounter
	memoryUsageGauge       metric.Float64UpDownCounter
	cpuUsageGauge          metric.Float64UpDownCounter
)

// InitTelemetry inicializa o sistema de telemetria
func InitTelemetry(serviceName string) error {
	// Criar exportador OTLP
	exporter, err := otlpmetricgrpc.New(context.Background())
	if err != nil {
		return fmt.Errorf("erro ao criar exportador OTLP: %v", err)
	}

	// Criar provedor de métricas
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(exporter,
				sdkmetric.WithInterval(3*time.Second),
			),
		),
	)
	otel.SetMeterProvider(provider)

	// Criar medidor
	meter = provider.Meter(serviceName)

	// Inicializar métricas
	taskProcessingDuration, err = meter.Float64Histogram(
		"task.processing.duration",
		metric.WithDescription("Duração do processamento de tarefas em segundos"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("erro ao criar métrica de duração: %v", err)
	}

	taskSuccessCounter, err = meter.Int64Counter(
		"task.success",
		metric.WithDescription("Número de tarefas concluídas com sucesso"),
	)
	if err != nil {
		return fmt.Errorf("erro ao criar contador de sucesso: %v", err)
	}

	taskFailureCounter, err = meter.Int64Counter(
		"task.failure",
		metric.WithDescription("Número de tarefas que falharam"),
	)
	if err != nil {
		return fmt.Errorf("erro ao criar contador de falhas: %v", err)
	}

	activeTasksGauge, err = meter.Int64UpDownCounter(
		"tasks.active",
		metric.WithDescription("Número atual de tarefas ativas"),
	)
	if err != nil {
		return fmt.Errorf("erro ao criar gauge de tarefas ativas: %v", err)
	}

	memoryUsageGauge, err = meter.Float64UpDownCounter(
		"system.memory.usage",
		metric.WithDescription("Uso de memória em bytes"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		return fmt.Errorf("erro ao criar gauge de memória: %v", err)
	}

	cpuUsageGauge, err = meter.Float64UpDownCounter(
		"system.cpu.usage",
		metric.WithDescription("Uso de CPU em porcentagem"),
		metric.WithUnit("%"),
	)
	if err != nil {
		return fmt.Errorf("erro ao criar gauge de CPU: %v", err)
	}

	log.Printf("✅ Telemetria inicializada para o serviço: %s", serviceName)
	return nil
}

// RecordTaskStart registra o início de uma tarefa
func RecordTaskStart(ctx context.Context, agentName, taskID, taskType string) {
	attrs := []attribute.KeyValue{
		attribute.String("agent", agentName),
		attribute.String("task_id", taskID),
		attribute.String("task_type", taskType),
	}

	activeTasksGauge.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordTaskCompletion registra a conclusão de uma tarefa
func RecordTaskCompletion(ctx context.Context, agentName, taskID, taskType string, duration time.Duration, success bool) {
	attrs := []attribute.KeyValue{
		attribute.String("agent", agentName),
		attribute.String("task_id", taskID),
		attribute.String("task_type", taskType),
	}

	taskProcessingDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	activeTasksGauge.Add(ctx, -1, metric.WithAttributes(attrs...))

	if success {
		taskSuccessCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	} else {
		taskFailureCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

// RecordResourceUsage registra o uso de recursos do sistema
func RecordResourceUsage(ctx context.Context, agentName string, memoryBytes float64, cpuPercent float64) {
	attrs := []attribute.KeyValue{
		attribute.String("agent", agentName),
	}

	memoryUsageGauge.Add(ctx, memoryBytes, metric.WithAttributes(attrs...))
	cpuUsageGauge.Add(ctx, cpuPercent, metric.WithAttributes(attrs...))
}

// GetMetrics retorna as métricas atuais
func GetMetrics() metric.Meter {
	return meter
}
