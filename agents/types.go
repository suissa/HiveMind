package agents

import "time"

// ProjectStatus representa o status atual de um projeto
type ProjectStatus struct {
	Progress       float64
	CompletedTasks int
	TotalTasks     int
	ElapsedTime    time.Duration
	RemainingTime  time.Duration
}
