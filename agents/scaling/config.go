package scaling

// Thresholds define os limites para escalabilidade
const (
	CPUThreshold    = 80.0 // 80% de uso de CPU
	MemoryThreshold = 85.0 // 85% de uso de memória
	TasksThreshold  = 100  // 100 tarefas na fila
	ErrorThreshold  = 0.05 // 5% de taxa de erro
	CooldownPeriod  = 300  // 5 minutos de cooldown entre escalas
)
