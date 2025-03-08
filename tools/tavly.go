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

const TAVILY_API_URL = "https://api.tavily.com/search"

// TavilyTool implementa a interface SearchTool para a API Tavily
type TavilyTool struct {
	apiToken string
}

// TavilyRequest representa a estrutura de requisição específica da API Tavily
type TavilyRequest struct {
	Query                   string   `json:"query"`
	Topic                   string   `json:"topic"`
	SearchDepth            string   `json:"search_depth"`
	MaxResults             int      `json:"max_results"`
	TimeRange              *string  `json:"time_range"`
	Days                   int      `json:"days"`
	IncludeAnswer          bool     `json:"include_answer"`
	IncludeRawContent      bool     `json:"include_raw_content"`
	IncludeImages          bool     `json:"include_images"`
	IncludeImageDescriptions bool   `json:"include_image_descriptions"`
	IncludeDomains         []string `json:"include_domains"`
	ExcludeDomains         []string `json:"exclude_domains"`
}

// NewTavilyTool cria uma nova instância da ferramenta Tavily
func NewTavilyTool() (*TavilyTool, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar arquivo .env: %v", err)
	}

	apiToken := os.Getenv("TAVILY_API_TOKEN")
	if apiToken == "" {
		return nil, fmt.Errorf("TAVILY_API_TOKEN não encontrado no arquivo .env")
	}

	return &TavilyTool{
		apiToken: apiToken,
	}, nil
}

// Search implementa a interface SearchTool
func (t *TavilyTool) Search(options SearchOptions) (*SearchResult, error) {
	// Converter as opções genéricas para o formato específico do Tavily
	tavilyReq := TavilyRequest{
		Query:                    options.Query,
		Topic:                    "general",
		SearchDepth:             "basic",
		MaxResults:              options.MaxResults,
		TimeRange:               nil,
		Days:                    3,
		IncludeAnswer:           options.IncludeAnswer,
		IncludeRawContent:       options.IncludeRawContent,
		IncludeImages:           options.IncludeImages,
		IncludeImageDescriptions: false,
		IncludeDomains:          options.IncludeDomains,
		ExcludeDomains:          options.ExcludeDomains,
	}

	// Converter a requisição para JSON
	jsonData, err := json.Marshal(tavilyReq)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter requisição para JSON: %v", err)
	}

	// Criar a requisição HTTP
	req, err := http.NewRequest("POST", TAVILY_API_URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição HTTP: %v", err)
	}

	// Adicionar headers
	req.Header.Set("Authorization", "Bearer "+t.apiToken)
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
		return nil, fmt.Errorf("erro na API Tavily: %s", string(body))
	}

	// Converter a resposta para o formato genérico
	var result SearchResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("erro ao processar resposta: %v", err)
	}

	return &result, nil
} 