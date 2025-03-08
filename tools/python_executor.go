package tools

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// PythonExecutorImpl implementa a interface PythonExecutor
type PythonExecutorImpl struct {
	pythonPath  string
	venvPath    string
	pipPath     string
	mu          sync.Mutex
}

// NewPythonExecutor cria uma nova instância do PythonExecutor
func NewPythonExecutor() (*PythonExecutorImpl, error) {
	// Encontrar Python no sistema
	pythonPath, err := findPython()
	if err != nil {
		return nil, err
	}

	executor := &PythonExecutorImpl{
		pythonPath: pythonPath,
	}

	// Criar ambiente virtual se necessário
	if err := executor.setupVirtualEnv(); err != nil {
		return nil, err
	}

	return executor, nil
}

// Execute executa um script Python
func (e *PythonExecutorImpl) Execute(options PythonExecutionOptions) (*PythonResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	startTime := time.Now()

	// Criar arquivo temporário para o script
	tmpFile, err := os.CreateTemp("", "python_script_*.py")
	if err != nil {
		return nil, fmt.Errorf("erro ao criar arquivo temporário: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Preparar script com wrapper para capturar saída e resultado
	wrappedScript := e.wrapScript(options.Script)
	if _, err := tmpFile.Write([]byte(wrappedScript)); err != nil {
		return nil, fmt.Errorf("erro ao escrever script: %v", err)
	}
	tmpFile.Close()

	// Preparar comando
	cmd := e.createPythonCommand(tmpFile.Name(), options)

	// Capturar saída
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Executar com timeout se especificado
	var err error
	if options.Context != nil && options.Context.Timeout > 0 {
		err = e.executeWithTimeout(cmd, options.Context.Timeout)
	} else {
		err = cmd.Run()
	}

	// Processar resultado
	result := &PythonResult{
		Duration:   time.Since(startTime).String(),
		MemoryUsed: e.getMemoryUsage(),
	}

	if err != nil {
		// Verificar se é erro de timeout
		if err.Error() == "timeout" {
			return nil, fmt.Errorf("timeout após %d segundos", options.Context.Timeout)
		}

		// Capturar erro Python
		errOutput := stderr.String()
		result.Error = e.parseError(errOutput)
		return result, nil
	}

	// Processar saída
	output := stdout.String()
	result.Output = output

	// Tentar extrair valor retornado (última linha em JSON)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > 0 {
		lastLine := lines[len(lines)-1]
		if strings.HasPrefix(lastLine, "PYTHON_RESULT:") {
			jsonStr := strings.TrimPrefix(lastLine, "PYTHON_RESULT:")
			var value PythonValue
			if err := json.Unmarshal([]byte(jsonStr), &value); err == nil {
				result.Value = &value
				// Remover linha do resultado do output
				result.Output = strings.Join(lines[:len(lines)-1], "\n")
			}
		}
	}

	return result, nil
}

// EvaluateExpression avalia uma expressão Python
func (e *PythonExecutorImpl) EvaluateExpression(expression string) (*PythonValue, error) {
	script := fmt.Sprintf("print('PYTHON_RESULT:' + json.dumps({'type': str(type(%s).__name__), 'value': %s}))",
		expression, expression)

	result, err := e.Execute(PythonExecutionOptions{
		Script: fmt.Sprintf("import json\n%s", script),
	})
	if err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, fmt.Errorf("erro ao avaliar expressão: %s", result.Error.Message)
	}

	return result.Value, nil
}

// CreateContext cria um novo contexto de execução
func (e *PythonExecutorImpl) CreateContext(variables map[string]interface{}) (*PythonContext, error) {
	context := &PythonContext{
		Variables:   variables,
		Timeout:    30, // 30 segundos por padrão
		MemoryLimit: 128 * 1024 * 1024, // 128MB por padrão
		PythonPath: e.pythonPath,
		VirtualEnv: e.venvPath,
	}

	return context, nil
}

// InstallPackage instala um pacote Python
func (e *PythonExecutorImpl) InstallPackage(name string, version string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	pip := e.pipPath
	if pip == "" {
		pip = "pip"
	}

	pkg := name
	if version != "" {
		pkg = fmt.Sprintf("%s==%s", name, version)
	}

	cmd := exec.Command(pip, "install", pkg)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("erro ao instalar pacote: %s\n%s", err, output)
	}

	return nil
}

// GetInstalledPackages retorna os pacotes instalados
func (e *PythonExecutorImpl) GetInstalledPackages() ([]string, error) {
	pip := e.pipPath
	if pip == "" {
		pip = "pip"
	}

	cmd := exec.Command(pip, "list", "--format=json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("erro ao listar pacotes: %v", err)
	}

	var packages []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(output, &packages); err != nil {
		return nil, fmt.Errorf("erro ao decodificar lista de pacotes: %v", err)
	}

	result := make([]string, len(packages))
	for i, pkg := range packages {
		result[i] = fmt.Sprintf("%s==%s", pkg.Name, pkg.Version)
	}

	return result, nil
}

// GetPythonVersion retorna a versão do Python
func (e *PythonExecutorImpl) GetPythonVersion() (string, error) {
	cmd := exec.Command(e.pythonPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("erro ao obter versão do Python: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// Funções auxiliares

func findPython() (string, error) {
	// Tentar Python 3 primeiro
	if path, err := exec.LookPath("python3"); err == nil {
		return path, nil
	}

	// Tentar apenas Python (pode ser Python 3 em alguns sistemas)
	if path, err := exec.LookPath("python"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("Python 3 não encontrado no sistema")
}

func (e *PythonExecutorImpl) setupVirtualEnv() error {
	// Criar diretório para ambiente virtual
	venvPath := filepath.Join(".", "venv")
	e.venvPath = venvPath

	// Verificar se já existe
	if _, err := os.Stat(venvPath); err == nil {
		// Ambiente virtual já existe
		e.updatePaths()
		return nil
	}

	// Criar novo ambiente virtual
	cmd := exec.Command(e.pythonPath, "-m", "venv", venvPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erro ao criar ambiente virtual: %v", err)
	}

	e.updatePaths()
	return nil
}

func (e *PythonExecutorImpl) updatePaths() {
	if runtime.GOOS == "windows" {
		e.pythonPath = filepath.Join(e.venvPath, "Scripts", "python.exe")
		e.pipPath = filepath.Join(e.venvPath, "Scripts", "pip.exe")
	} else {
		e.pythonPath = filepath.Join(e.venvPath, "bin", "python")
		e.pipPath = filepath.Join(e.venvPath, "bin", "pip")
	}
}

func (e *PythonExecutorImpl) wrapScript(script string) string {
	return fmt.Sprintf(`
import sys
import json
import traceback

def main():
    try:
%s
        result = locals().get('result', None)
        if result is not None:
            print('PYTHON_RESULT:' + json.dumps({
                'type': str(type(result).__name__),
                'value': result
            }))
    except Exception as e:
        traceback.print_exc()
        sys.exit(1)

if __name__ == '__main__':
    main()
`, indentScript(script))
}

func (e *PythonExecutorImpl) createPythonCommand(scriptPath string, options PythonExecutionOptions) *exec.Cmd {
	cmd := exec.Command(e.pythonPath, scriptPath)

	// Configurar ambiente
	env := os.Environ()
	if options.Environment != nil {
		for k, v := range options.Environment {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}
	cmd.Env = env

	// Configurar PYTHONPATH
	if options.Context != nil && options.Context.PythonPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PYTHONPATH=%s", options.Context.PythonPath))
	}

	return cmd
}

func (e *PythonExecutorImpl) executeWithTimeout(cmd *exec.Cmd, timeout int) error {
	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(time.Duration(timeout) * time.Second):
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("erro ao matar processo: %v", err)
		}
		return fmt.Errorf("timeout")
	}
}

func (e *PythonExecutorImpl) parseError(errOutput string) *PythonError {
	lines := strings.Split(errOutput, "\n")
	if len(lines) == 0 {
		return &PythonError{
			Type:    "Unknown",
			Message: "Erro desconhecido",
		}
	}

	error := &PythonError{
		Traceback: errOutput,
	}

	// Tentar extrair tipo e mensagem do erro
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, ": ") {
			parts := strings.SplitN(line, ": ", 2)
			error.Type = parts[0]
			error.Message = parts[1]
			break
		}
	}

	// Tentar extrair número da linha
	for _, line := range lines {
		if strings.Contains(line, "line") {
			if _, err := fmt.Sscanf(line, "  File \"%s\", line %d", &error.Source, &error.LineNumber); err == nil {
				break
			}
		}
	}

	return error
}

func (e *PythonExecutorImpl) getMemoryUsage() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Alloc)
}

func indentScript(script string) string {
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(script))
	for scanner.Scan() {
		result.WriteString("    " + scanner.Text() + "\n")
	}
	return result.String()
} 