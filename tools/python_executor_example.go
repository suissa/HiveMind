package tools

import (
	"fmt"
	"log"
)

// ExamplePythonExecutor demonstra o uso do executor Python
func ExamplePythonExecutor() {
	// Criar executor
	executor, err := NewPythonExecutor()
	if err != nil {
		log.Fatal(err)
	}

	// Verificar versão do Python
	version, err := executor.GetPythonVersion()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Versão do Python: %s\n", version)

	// Exemplo simples
	result, err := executor.Execute(PythonExecutionOptions{
		Script: `
import math

def calculate_circle_area(radius):
    return math.pi * radius ** 2

radius = 5
result = {
    'radius': radius,
    'area': calculate_circle_area(radius),
    'pi': math.pi
}
`,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Execução completada em %s\n", result.Duration)
	fmt.Printf("Memória utilizada: %d bytes\n", result.MemoryUsed)

	if result.Value != nil {
		fmt.Printf("Resultado: %+v\n", result.Value)
	}

	// Instalar e usar pacotes externos
	if err := executor.InstallPackage("numpy", ""); err != nil {
		log.Fatal(err)
	}

	result, err = executor.Execute(PythonExecutionOptions{
		Script: `
import numpy as np

# Criar array e realizar operações
arr = np.array([1, 2, 3, 4, 5])
mean = np.mean(arr)
std = np.std(arr)

result = {
    'array': arr.tolist(),
    'mean': mean,
    'std': std
}
`,
	})
	if err != nil {
		log.Fatal(err)
	}

	if result.Value != nil {
		fmt.Printf("\nResultados NumPy:\n%+v\n", result.Value)
	}

	// Exemplo com variáveis no contexto
	context, err := executor.CreateContext(map[string]interface{}{
		"input_data": map[string]interface{}{
			"x": []float64{1, 2, 3, 4, 5},
			"y": []float64{2, 4, 6, 8, 10},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	result, err = executor.Execute(PythonExecutionOptions{
		Script: `
import json

x = input_data['x']
y = input_data['y']

# Calcular correlação
def calculate_correlation(x, y):
    n = len(x)
    sum_x = sum(x)
    sum_y = sum(y)
    sum_xy = sum(x[i] * y[i] for i in range(n))
    sum_x2 = sum(x[i] ** 2 for i in range(n))
    sum_y2 = sum(y[i] ** 2 for i in range(n))
    
    correlation = (n * sum_xy - sum_x * sum_y) / (
        ((n * sum_x2 - sum_x ** 2) * (n * sum_y2 - sum_y ** 2)) ** 0.5
    )
    return correlation

result = {
    'correlation': calculate_correlation(x, y),
    'x_mean': sum(x) / len(x),
    'y_mean': sum(y) / len(y)
}
`,
		Context: context,
	})
	if err != nil {
		log.Fatal(err)
	}

	if result.Value != nil {
		fmt.Printf("\nAnálise de Correlação:\n%+v\n", result.Value)
	}
}

// ExamplePythonExecutorAdvanced demonstra recursos avançados do executor
func ExamplePythonExecutorAdvanced() {
	executor, err := NewPythonExecutor()
	if err != nil {
		log.Fatal(err)
	}

	// Exemplo com tratamento de erros
	result, err := executor.Execute(PythonExecutionOptions{
		Script: `
def divide(a, b):
    if b == 0:
        raise ValueError("Divisão por zero não permitida")
    return a / b

try:
    result = divide(10, 0)
except Exception as e:
    print(f"Erro capturado: {e}")
    raise
`,
		Debug: true,
	})

	if result.Error != nil {
		fmt.Printf("\nErro na execução:\n")
		fmt.Printf("Tipo: %s\n", result.Error.Type)
		fmt.Printf("Mensagem: %s\n", result.Error.Message)
		if result.Error.Traceback != "" {
			fmt.Printf("Traceback:\n%s\n", result.Error.Traceback)
		}
	}

	// Exemplo com processamento de dados
	result, err = executor.Execute(PythonExecutionOptions{
		Script: `
import json
from datetime import datetime

# Dados de exemplo
data = [
    {"date": "2024-01-01", "value": 100},
    {"date": "2024-01-02", "value": 150},
    {"date": "2024-01-03", "value": 120},
    {"date": "2024-01-04", "value": 200}
]

# Processar dados
def process_data(data):
    total = 0
    dates = []
    values = []
    
    for item in data:
        date = datetime.strptime(item['date'], '%Y-%m-%d')
        value = item['value']
        
        total += value
        dates.append(date.strftime('%Y-%m-%d'))
        values.append(value)
    
    return {
        'total': total,
        'average': total / len(data),
        'dates': dates,
        'values': values,
        'count': len(data)
    }

result = process_data(data)
`,
	})
	if err != nil {
		log.Fatal(err)
	}

	if result.Value != nil {
		fmt.Printf("\nProcessamento de Dados:\n%+v\n", result.Value)
	}

	// Exemplo com avaliação de expressão
	value, err := executor.EvaluateExpression("sum(range(1, 101))")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nSoma dos números de 1 a 100: %v (tipo: %s)\n", value.Value, value.Type)

	// Listar pacotes instalados
	packages, err := executor.GetInstalledPackages()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nPacotes instalados:\n")
	for _, pkg := range packages {
		fmt.Printf("- %s\n", pkg)
	}
} 