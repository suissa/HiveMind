package tools

import (
	"encoding/json"
	"fmt"
	"log"
)

// ExampleJSExecutor demonstra o uso do executor JavaScript
func ExampleJSExecutor() {
	// Criar executor
	executor, err := NewV8Executor()
	if err != nil {
		log.Fatal(err)
	}

	// Exemplo simples
	result, err := executor.Execute(JSExecutionOptions{
		Script: `
			const message = "Olá do JavaScript!";
			const number = 42;
			const array = [1, 2, 3];
			const obj = { name: "teste", value: 123 };
			
			console.log(message);
			
			({ message, number, array, obj })
		`,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Execução completada em %s\n", result.Duration)
	fmt.Printf("Memória utilizada: %d bytes\n", result.MemoryUsed)

	if result.Value != nil {
		printJSValue(result.Value)
	}

	// Exemplo com contexto personalizado
	context, err := executor.CreateContext(map[string]interface{}{
		"config": map[string]interface{}{
			"apiKey": "123456",
			"debug":  true,
		},
		"data": []interface{}{1, 2, 3, 4, 5},
	})
	if err != nil {
		log.Fatal(err)
	}

	result, err = executor.Execute(JSExecutionOptions{
		Script: `
			console.log("Config:", config);
			console.log("Data:", data);
			
			const sum = data.reduce((a, b) => a + b, 0);
			const apiKey = config.apiKey;
			
			({ sum, apiKey })
		`,
		Context: context,
	})
	if err != nil {
		log.Fatal(err)
	}

	if result.Value != nil {
		printJSValue(result.Value)
	}

	// Exemplo com módulo
	moduleCode := `
		export function calculate(x, y) {
			return x * y + Math.pow(2, 3);
		}

		export const constants = {
			PI: Math.PI,
			E: Math.E,
		};
	`

	if err := executor.LoadModule("math", moduleCode); err != nil {
		log.Fatal(err)
	}

	result, err = executor.Execute(JSExecutionOptions{
		Script: `
			const math = require('./math');
			const result = math.calculate(5, 3);
			const pi = math.constants.PI;
			
			({ result, pi })
		`,
		Context: &JSContext{
			Modules: []string{"math"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	if result.Value != nil {
		printJSValue(result.Value)
	}
}

// ExampleJSExecutorAdvanced demonstra recursos avançados do executor
func ExampleJSExecutorAdvanced() {
	executor, err := NewV8Executor()
	if err != nil {
		log.Fatal(err)
	}

	// Exemplo com async/await
	result, err := executor.Execute(JSExecutionOptions{
		Script: `
			async function fetchData() {
				// Simulando uma chamada assíncrona
				await new Promise(resolve => setTimeout(resolve, 1000));
				return { data: "Dados obtidos com sucesso!" };
			}

			async function main() {
				console.log("Iniciando...");
				const result = await fetchData();
				console.log("Resultado:", result);
				return result;
			}

			main();
		`,
		AsyncMode: true,
		Context: &JSContext{
			Timeout: 5, // 5 segundos
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	if result.Value != nil {
		printJSValue(result.Value)
	}

	// Exemplo com tratamento de erros
	result, err = executor.Execute(JSExecutionOptions{
		Script: `
			function processData(data) {
				if (!Array.isArray(data)) {
					throw new Error("Entrada deve ser um array");
				}
				return data.map(x => x * 2);
			}

			try {
				const result = processData("não é um array");
				console.log(result);
			} catch (error) {
				console.log("Erro:", error.message);
				throw error;
			}
		`,
		Debug: true,
	})

	if result.Error != nil {
		fmt.Printf("\nErro na execução:\n")
		fmt.Printf("Mensagem: %s\n", result.Error.Message)
		fmt.Printf("Linha: %d, Coluna: %d\n", result.Error.LineNumber, result.Error.Column)
		if result.Error.Stack != "" {
			fmt.Printf("Stack trace:\n%s\n", result.Error.Stack)
		}
	}

	// Exemplo com avaliação de expressão
	value, err := executor.EvaluateExpression(`
		(function() {
			const x = 10;
			const y = 20;
			return x + y;
		})()
	`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nResultado da expressão: %v (tipo: %s)\n", value.Value, value.Type)
}

// printJSValue imprime um valor JavaScript de forma formatada
func printJSValue(value *JSValue) {
	fmt.Printf("\nValor JavaScript (tipo: %s):\n", value.Type)
	
	switch value.Type {
	case "object":
		if data, err := json.MarshalIndent(value.Value, "", "  "); err == nil {
			fmt.Println(string(data))
		}
	case "array":
		if arr, ok := value.Value.([]interface{}); ok {
			fmt.Println(arr)
		}
	default:
		fmt.Println(value.Value)
	}
} 