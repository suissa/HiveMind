package tools

// PythonValue representa um valor retornado pela execução do Python
type PythonValue struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// PythonError representa um erro de execução Python
type PythonError struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	LineNumber int    `json:"line_number"`
	Traceback  string `json:"traceback,omitempty"`
	Source     string `json:"source,omitempty"`
}

// PythonResult representa o resultado da execução de código Python
type PythonResult struct {
	Value      *PythonValue `json:"value,omitempty"`
	Error      *PythonError `json:"error,omitempty"`
	Output     string       `json:"output,omitempty"`
	Duration   string       `json:"duration"`
	MemoryUsed int64        `json:"memory_used"`
}

// PythonContext representa o contexto de execução Python
type PythonContext struct {
	Variables   map[string]interface{} `json:"variables"`
	Packages    []string              `json:"packages,omitempty"`
	VirtualEnv  string                `json:"virtual_env,omitempty"`
	PythonPath  string                `json:"python_path,omitempty"`
	Timeout     int                   `json:"timeout"`
	MemoryLimit int64                 `json:"memory_limit"`
}

// PythonExecutionOptions representa as opções para execução de Python
type PythonExecutionOptions struct {
	Script      string                 `json:"script"`
	Context     *PythonContext        `json:"context,omitempty"`
	Interactive bool                   `json:"interactive"`
	Debug       bool                   `json:"debug"`
	Environment map[string]string      `json:"environment,omitempty"`
}

// PythonExecutor é a interface que todas as ferramentas de execução Python devem implementar
type PythonExecutor interface {
	// Execute executa um script Python
	Execute(options PythonExecutionOptions) (*PythonResult, error)

	// EvaluateExpression avalia uma expressão Python
	EvaluateExpression(expression string) (*PythonValue, error)

	// CreateContext cria um novo contexto de execução
	CreateContext(variables map[string]interface{}) (*PythonContext, error)

	// InstallPackage instala um pacote Python
	InstallPackage(name string, version string) error

	// GetInstalledPackages retorna os pacotes instalados
	GetInstalledPackages() ([]string, error)

	// GetPythonVersion retorna a versão do Python
	GetPythonVersion() (string, error)
} 