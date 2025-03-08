package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

const EXA_API_URL = "https://api.exa.ai/search"

// ExaTool implementa a interface SearchTool para a API Exa
type ExaTool struct {
	apiToken string
}

// ExaRequest representa a estrutura de requisição específica da API Exa
type ExaRequest struct {
	Query string `json:"query"`
	Text  bool   `json:"text"`
}

// ExaResponse representa a estrutura de resposta específica da API Exa
type ExaResponse struct {
	Results []struct {
		Content string   `json:"content"`
		URL     string   `json:"url"`
	} `json:"results"`
}

// NewExaTool cria uma nova instância da ferramenta Exa
func NewExaTool() (*ExaTool, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar arquivo .env: %v", err)
	}

	apiToken := os.Getenv("EXA_API_TOKEN")
	if apiToken == "" {
		return nil, fmt.Errorf("EXA_API_TOKEN não encontrado no arquivo .env")
	}

	return &ExaTool{
		apiToken: apiToken,
	}, nil
}

// Search implementa a interface SearchTool
func (e *ExaTool) Search(options SearchOptions) (*SearchResult, error) {
	// Converter as opções genéricas para o formato específico do Exa
	exaReq := ExaRequest{
		Query: options.Query,
		Text:  true, // Sempre true para obter resultados em texto
	}

	// Converter a requisição para JSON
	jsonData, err := json.Marshal(exaReq)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter requisição para JSON: %v", err)
	}

	// Criar a requisição HTTP
	req, err := http.NewRequest("POST", EXA_API_URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição HTTP: %v", err)
	}

	// Adicionar headers
	req.Header.Set("Authorization", "Bearer "+e.apiToken)
	req.Header.Set("Content-Type", "application/json")

	// Executar a requisição
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao executar requisição: %v", err)
	}
	defer resp.Body.Close()

	// Ler a resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro na API Exa: %s", string(body))
	}

	// Converter a resposta do Exa
	var exaResp ExaResponse
	err = json.Unmarshal(body, &exaResp)
	if err != nil {
		return nil, fmt.Errorf("erro ao processar resposta: %v", err)
	}

	// Converter para o formato genérico SearchResult
	result := &SearchResult{
		RawContent: "", // Será preenchido com o conteúdo concatenado
		URLs:       make([]string, 0),
	}

	// Processar os resultados
	for _, r := range exaResp.Results {
		if options.IncludeRawContent {
			if result.RawContent != "" {
				result.RawContent += "\n\n"
			}
			result.RawContent += r.Content
		}
		result.URLs = append(result.URLs, r.URL)
	}

	return result, nil
} 