package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

// MeilisearchClient implementa a interface SearchEngineTool
type MeilisearchClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewMeilisearchClient cria uma nova instância do MeilisearchClient
func NewMeilisearchClient() (*MeilisearchClient, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar arquivo .env: %v", err)
	}

	baseURL := os.Getenv("MEILISEARCH_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("MEILISEARCH_URL não encontrado no arquivo .env")
	}

	apiKey := os.Getenv("MEILISEARCH_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("MEILISEARCH_API_KEY não encontrado no arquivo .env")
	}

	return &MeilisearchClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Search implementa a busca no Meilisearch
func (m *MeilisearchClient) Search(options SearchEngineOptions) (*SearchEngineResult, error) {
	url := fmt.Sprintf("%s/indexes/%s/search", m.baseURL, options.IndexName)

	// Preparar corpo da requisição
	searchParams := map[string]interface{}{
		"q": options.Query,
	}

	if options.Limit > 0 {
		searchParams["limit"] = options.Limit
	}
	if options.Offset > 0 {
		searchParams["offset"] = options.Offset
	}
	if len(options.Sort) > 0 {
		searchParams["sort"] = options.Sort
	}
	if len(options.Filters) > 0 {
		searchParams["filter"] = options.Filters
	}
	// Adicionar parâmetros adicionais de busca
	for k, v := range options.SearchParams {
		searchParams[k] = v
	}

	body, err := json.Marshal(searchParams)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar parâmetros de busca: %v", err)
	}

	// Criar requisição
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	// Executar requisição
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao executar requisição: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro na API do Meilisearch: %s", string(body))
	}

	// Decodificar resposta
	var result SearchEngineResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return &result, nil
}

// Index implementa a indexação de documentos no Meilisearch
func (m *MeilisearchClient) Index(indexName string, documents interface{}, primaryKey string) (*IndexResult, error) {
	url := fmt.Sprintf("%s/indexes/%s/documents", m.baseURL, indexName)
	if primaryKey != "" {
		url += "?primaryKey=" + primaryKey
	}

	var body []byte
	var err error

	// Se documents é um arquivo JSON
	if docPath, ok := documents.(string); ok && filepath.Ext(docPath) == ".json" {
		body, err = os.ReadFile(docPath)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler arquivo JSON: %v", err)
		}
	} else {
		// Se documents é uma estrutura de dados
		body, err = json.Marshal(documents)
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar documentos: %v", err)
		}
	}

	// Criar requisição
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	// Executar requisição
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao executar requisição: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro na API do Meilisearch: %s", string(body))
	}

	// Decodificar resposta
	var result IndexResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return &result, nil
}

// DeleteIndex implementa a deleção de um índice no Meilisearch
func (m *MeilisearchClient) DeleteIndex(indexName string) error {
	url := fmt.Sprintf("%s/indexes/%s", m.baseURL, indexName)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao executar requisição: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("erro na API do Meilisearch: %s", string(body))
	}

	return nil
}

// GetStats obtém estatísticas de um índice no Meilisearch
func (m *MeilisearchClient) GetStats(indexName string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/indexes/%s/stats", m.baseURL, indexName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao executar requisição: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro na API do Meilisearch: %s", string(body))
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return stats, nil
} 