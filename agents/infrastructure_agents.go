package agents

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

// Thresholds para escalonamento
const (
	cpuThreshold    = 80.0 // 80% de uso de CPU
	memoryThreshold = 80.0 // 80% de uso de memória
	tasksThreshold  = 100  // 100 tarefas na fila
	errorThreshold  = 5.0  // 5% de taxa de erro

	// Período de cooldown entre escalonamentos (em segundos)
	cooldownPeriod = 300 // 5 minutos
)

// AgentMetrics representa as métricas de um agente
type AgentMetrics struct {
	AgentName    string  `json:"agent_name"`
	CPU          float64 `json:"cpu_usage"`
	Memory       uint64  `json:"memory_usage"`
	TasksInQueue int     `json:"tasks_in_queue"`
	ResponseTime float64 `json:"response_time"`
	ErrorRate    float64 `json:"error_rate"`
	LastUpdated  int64   `json:"last_updated"`
}

// SystemMetrics representa as métricas do sistema como um todo
type SystemMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	TaskCount   int     `json:"task_count"`
	ErrorRate   float64 `json:"error_rate"`
}

// ObserverInfrastructureAgent monitora e coleta métricas dos agentes
type ObserverInfrastructureAgent struct {
	*AgentStruct
	metricsMap     map[string]*AgentMetrics
	metricsMapLock sync.RWMutex
	rabbitmqConn   *amqp.Connection
	rabbitmqCh     *amqp.Channel
	systemMetrics  *SystemMetrics
}

// OrchestratorInfrastructureAgent gerencia o escalonamento dos agentes
type OrchestratorInfrastructureAgent struct {
	*AgentStruct
	instances     map[string][]*AgentInstance
	instancesLock sync.RWMutex
	rabbitmqConn  *amqp.Connection
	rabbitmqCh    *amqp.Channel
	lastScaleTime time.Time
	metrics       *SystemMetrics
}

// AgentInstance representa uma instância de um agente
type AgentInstance struct {
	Agent      *AgentStruct
	LastScaled time.Time
	Metrics    *AgentMetrics
}

// NewObserverInfrastructureAgent cria uma nova instância do ObserverInfrastructureAgent
func NewObserverInfrastructureAgent() *ObserverInfrastructureAgent {
	agent := &ObserverInfrastructureAgent{
		AgentStruct: &AgentStruct{
			Name:            "Observer Infrastructure Agent",
			Role:            "Monitor de Infraestrutura",
			Goal:            "Monitorar métricas de todos os agentes e publicar eventos de telemetria",
			AllowDelegation: false,
			Model:           "gpt-4o-mini",
			Backstory:       "Um especialista em monitoramento e telemetria que observa o comportamento dos agentes",
		},
		metricsMap:    make(map[string]*AgentMetrics),
		systemMetrics: &SystemMetrics{},
	}

	// Inicializar conexão com RabbitMQ
	conn, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@%s:%d/", RABBITMQ_HOST, RABBITMQ_PORT))
	if err != nil {
		log.Fatalf("Falha ao conectar ao RabbitMQ: %v", err)
	}
	agent.rabbitmqConn = conn

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Falha ao abrir canal: %v", err)
	}
	agent.rabbitmqCh = ch

	return agent
}

// NewOrchestratorInfrastructureAgent cria uma nova instância do OrchestratorInfrastructureAgent
func NewOrchestratorInfrastructureAgent() *OrchestratorInfrastructureAgent {
	agent := &OrchestratorInfrastructureAgent{
		AgentStruct: &AgentStruct{
			Name:            "Orchestrator Infrastructure Agent",
			Role:            "Gerenciador de Infraestrutura",
			Goal:            "Gerenciar a escalabilidade dinâmica dos agentes cognitivos",
			AllowDelegation: false,
			Model:           "gpt-4-mini",
			Backstory:       "Um especialista em infraestrutura que monitora e escala recursos automaticamente",
		},
		instances:     make(map[string][]*AgentInstance),
		lastScaleTime: time.Now(),
		metrics:       &SystemMetrics{},
	}

	// Inicializar conexão com RabbitMQ
	conn, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@%s:%d/", RABBITMQ_HOST, RABBITMQ_PORT))
	if err != nil {
		log.Fatalf("Falha ao conectar ao RabbitMQ: %v", err)
	}
	agent.rabbitmqConn = conn

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Falha ao abrir canal: %v", err)
	}
	agent.rabbitmqCh = ch

	return agent
}

// StartMonitoring inicia o monitoramento dos agentes
func (o *ObserverInfrastructureAgent) StartMonitoring(agents ...*AgentStruct) {
	log.Printf("🔍 Iniciando monitoramento de %d agentes", len(agents))

	// Inicializar métricas para cada agente
	for _, agent := range agents {
		o.metricsMap[agent.GetName()] = &AgentMetrics{
			AgentName: agent.GetName(),
		}
	}

	// Iniciar coleta de métricas em background
	go o.collectMetrics()
}

// collectMetrics coleta métricas periodicamente
func (o *ObserverInfrastructureAgent) collectMetrics() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		o.metricsMapLock.Lock()

		var totalCPU float64
		var totalMemory uint64
		var totalTasks int
		var totalErrors float64
		var agentCount int

		for agentName, metrics := range o.metricsMap {
			// Coletar métricas do agente
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			metrics.Memory = m.Alloc
			metrics.CPU = getCPUUsage()
			metrics.LastUpdated = time.Now().Unix()

			// Acumular métricas do sistema
			totalCPU += metrics.CPU
			totalMemory += metrics.Memory
			totalTasks += metrics.TasksInQueue
			totalErrors += metrics.ErrorRate
			agentCount++

			// Publicar métricas no RabbitMQ
			metricsJSON, err := json.Marshal(metrics)
			if err != nil {
				log.Printf("Erro ao converter métricas para JSON: %v", err)
				continue
			}

			// Publicar cada métrica separadamente
			o.publishMetric(agentName, "cpu", metrics.CPU)
			o.publishMetric(agentName, "memory", float64(metrics.Memory))
			o.publishMetric(agentName, "tasks_in_queue", float64(metrics.TasksInQueue))
			o.publishMetric(agentName, "response_time", metrics.ResponseTime)
			o.publishMetric(agentName, "error_rate", metrics.ErrorRate)

			log.Printf("📊 Métricas coletadas para %s: %s", agentName, string(metricsJSON))
		}

		// Atualizar métricas do sistema
		if agentCount > 0 {
			o.systemMetrics.CPUUsage = totalCPU / float64(agentCount)
			o.systemMetrics.MemoryUsage = float64(totalMemory) / float64(agentCount)
			o.systemMetrics.TaskCount = totalTasks
			o.systemMetrics.ErrorRate = totalErrors / float64(agentCount)
		}

		o.metricsMapLock.Unlock()
	}
}

// publishMetric publica uma métrica específica no RabbitMQ
func (o *ObserverInfrastructureAgent) publishMetric(agentName, metricName string, value float64) {
	queueName := fmt.Sprintf("metrics.%s.%s", agentName, metricName)

	// Declarar a fila para a métrica
	_, err := o.rabbitmqCh.QueueDeclare(
		queueName, // nome
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Printf("Erro ao declarar fila %s: %v", queueName, err)
		return
	}

	// Criar mensagem com a métrica
	message := struct {
		Value     float64 `json:"value"`
		Timestamp int64   `json:"timestamp"`
	}{
		Value:     value,
		Timestamp: time.Now().Unix(),
	}

	body, err := json.Marshal(message)
	if err != nil {
		log.Printf("Erro ao converter mensagem para JSON: %v", err)
		return
	}

	// Publicar mensagem
	err = o.rabbitmqCh.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: 2, // mensagem persistente
		})
	if err != nil {
		log.Printf("Erro ao publicar métrica %s: %v", queueName, err)
	}
}

// UpdateMetrics atualiza as métricas de um agente específico
func (o *ObserverInfrastructureAgent) UpdateMetrics(agentName string, metrics *AgentMetrics) {
	o.metricsMapLock.Lock()
	defer o.metricsMapLock.Unlock()

	o.metricsMap[agentName] = metrics
}

// GetMetrics retorna as métricas de um agente específico
func (o *ObserverInfrastructureAgent) GetMetrics(agentName string) *AgentMetrics {
	o.metricsMapLock.RLock()
	defer o.metricsMapLock.RUnlock()

	return o.metricsMap[agentName]
}

// GetSystemMetrics retorna as métricas do sistema como um todo
func (o *ObserverInfrastructureAgent) GetSystemMetrics() *SystemMetrics {
	o.metricsMapLock.RLock()
	defer o.metricsMapLock.RUnlock()

	return o.systemMetrics
}

// RegisterAgent registra um novo agente para monitoramento
func (o *OrchestratorInfrastructureAgent) RegisterAgent(agent *AgentStruct) {
	o.instancesLock.Lock()
	defer o.instancesLock.Unlock()

	instance := &AgentInstance{
		Agent:      agent,
		LastScaled: time.Now(),
		Metrics:    &AgentMetrics{AgentName: agent.GetName()},
	}

	o.instances[agent.GetName()] = append(o.instances[agent.GetName()], instance)
	log.Printf("✅ Agente %s registrado para monitoramento", agent.GetName())
}

// CheckScaling verifica se é necessário escalar o sistema
func (o *OrchestratorInfrastructureAgent) CheckScaling() bool {
	if time.Since(o.lastScaleTime).Seconds() < float64(cooldownPeriod) {
		return false
	}

	if o.metrics.CPUUsage > cpuThreshold ||
		o.metrics.MemoryUsage > memoryThreshold ||
		o.metrics.TaskCount > tasksThreshold ||
		o.metrics.ErrorRate > errorThreshold {
		o.lastScaleTime = time.Now()
		return true
	}

	return false
}

// UpdateMetrics atualiza as métricas do sistema
func (o *OrchestratorInfrastructureAgent) UpdateMetrics(metrics *SystemMetrics) {
	o.metrics = metrics
}

// ScaleSystem escala o sistema baseado nas métricas atuais
func (o *OrchestratorInfrastructureAgent) ScaleSystem() error {
	if !o.CheckScaling() {
		return nil
	}

	o.instancesLock.Lock()
	defer o.instancesLock.Unlock()

	// Escalar cada tipo de agente
	for agentType, instances := range o.instances {
		// Verificar métricas do tipo de agente
		var totalCPU float64
		var totalMemory float64
		var totalTasks int
		var totalErrors float64
		var instanceCount int

		for _, instance := range instances {
			metrics := instance.Metrics
			totalCPU += metrics.CPU
			totalMemory += float64(metrics.Memory)
			totalTasks += metrics.TasksInQueue
			totalErrors += metrics.ErrorRate
			instanceCount++
		}

		// Calcular médias
		avgCPU := totalCPU / float64(instanceCount)
		avgMemory := totalMemory / float64(instanceCount)
		avgTasks := float64(totalTasks) / float64(instanceCount)
		avgErrors := totalErrors / float64(instanceCount)

		// Decidir se precisa escalar
		if avgCPU > cpuThreshold ||
			avgMemory > memoryThreshold ||
			avgTasks > float64(tasksThreshold) ||
			avgErrors > errorThreshold {
			// Criar nova instância do agente
			baseInstance := instances[0]
			newAgent := baseInstance.Agent.Clone()

			// Registrar nova instância
			instance := &AgentInstance{
				Agent:      newAgent,
				LastScaled: time.Now(),
				Metrics:    &AgentMetrics{AgentName: newAgent.GetName()},
			}

			o.instances[agentType] = append(o.instances[agentType], instance)
			log.Printf("🔄 Escalando agente %s: nova instância criada", agentType)
		}
	}

	return nil
}

// getCPUUsage retorna o uso atual de CPU (simulado)
func getCPUUsage() float64 {
	return 50.0 // Valor simulado de 50% de uso de CPU
}
