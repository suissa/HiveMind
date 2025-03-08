package tools

// JSValue representa um valor retornado pela execução do JavaScript
type JSValue struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// JSError representa um erro de execução JavaScript
type JSError struct {
	Message    string `json:"message"`
	LineNumber int    `json:"line_number"`
	Column     int    `json:"column"`
	Stack      string `json:"stack,omitempty"`
	Source     string `json:"source,omitempty"`
}

// JSResult representa o resultado da execução de código JavaScript
type JSResult struct {
	Value      *JSValue `json:"value,omitempty"`
	Error      *JSError `json:"error,omitempty"`
	Duration   string   `json:"duration"`
	MemoryUsed int64    `json:"memory_used"`
}

// JSContext representa o contexto de execução JavaScript
type JSContext struct {
	Globals    map[string]interface{} `json:"globals"`
	Modules    []string               `json:"modules,omitempty"`
	Timeout    int                    `json:"timeout"`
	MemoryLimit int64                 `json:"memory_limit"`
}

// JSExecutionOptions representa as opções para execução de JavaScript
type JSExecutionOptions struct {
	Script      string                 `json:"script"`
	Context     *JSContext            `json:"context,omitempty"`
	AsyncMode   bool                   `json:"async_mode"`
	Debug       bool                   `json:"debug"`
	Environment map[string]interface{} `json:"environment,omitempty"`
}

// JSExecutor é a interface que todas as ferramentas de execução JS devem implementar
type JSExecutor interface {
	// Execute executa um script JavaScript
	Execute(options JSExecutionOptions) (*JSResult, error)

	// EvaluateExpression avalia uma expressão JavaScript
	EvaluateExpression(expression string) (*JSValue, error)

	// CreateContext cria um novo contexto de execução
	CreateContext(globals map[string]interface{}) (*JSContext, error)

	// LoadModule carrega um módulo JavaScript
	LoadModule(name, source string) error

	// GetAvailableModules retorna os módulos disponíveis
	GetAvailableModules() []string
} 