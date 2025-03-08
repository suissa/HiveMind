package tools

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// APIClient implementa a interface APITool
type APIClient struct {
	client *http.Client
}

// NewAPIClient cria uma nova instância do APIClient
func NewAPIClient() *APIClient {
	return &APIClient{
		client: &http.Client{},
	}
}

// Request implementa a interface APITool
func (c *APIClient) Request(options APIOptions) (*APIResponse, error) {
	startTime := time.Now()

	// Validar método HTTP
	method := strings.ToUpper(options.Method)
	if method == "" {
		method = "GET"
	}

	// Construir URL com query params
	reqURL, err := c.buildURL(options.URL, options.QueryParams)
	if err != nil {
		return nil, fmt.Errorf("erro ao construir URL: %v", err)
	}

	// Preparar corpo da requisição
	var bodyReader io.Reader
	if options.Body != nil {
		bodyData, err := json.Marshal(options.Body)
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar corpo da requisição: %v", err)
		}
		bodyReader = bytes.NewReader(bodyData)
	}

	// Criar requisição
	req, err := http.NewRequest(method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %v", err)
	}

	// Adicionar headers padrão se não existirem
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}

	// Adicionar headers customizados
	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}

	// Configurar autenticação
	if options.Auth != nil {
		err = c.setAuthentication(req, options.Auth)
		if err != nil {
			return nil, fmt.Errorf("erro ao configurar autenticação: %v", err)
		}
	}

	// Configurar timeout
	if options.Timeout > 0 {
		c.client.Timeout = options.Timeout
	}

	// Executar requisição com retry
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= options.RetryCount; attempt++ {
		if attempt > 0 && options.RetryDelay > 0 {
			time.Sleep(options.RetryDelay)
		}

		resp, err = c.client.Do(req)
		if err == nil {
			break
		}
		lastErr = err
	}

	if err != nil {
		return &APIResponse{
			Error:        fmt.Sprintf("erro após %d tentativas: %v", options.RetryCount+1, lastErr),
			ResponseTime: time.Since(startTime),
		}, nil
	}
	defer resp.Body.Close()

	// Ler corpo da resposta
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler corpo da resposta: %v", err)
	}

	// Preparar headers da resposta
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}

	// Preparar resposta
	apiResp := &APIResponse{
		StatusCode:   resp.StatusCode,
		Headers:      headers,
		RawBody:      respBody,
		ResponseTime: time.Since(startTime),
	}

	// Tentar fazer parse do corpo como JSON
	var jsonBody interface{}
	if len(respBody) > 0 {
		err = json.Unmarshal(respBody, &jsonBody)
		if err == nil {
			apiResp.Body = jsonBody
		} else {
			apiResp.Body = string(respBody)
		}
	}

	return apiResp, nil
}

// buildURL constrói a URL final com os query params
func (c *APIClient) buildURL(baseURL string, params map[string]string) (string, error) {
	if len(params) == 0 {
		return baseURL, nil
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	query := parsedURL.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

// setAuthentication configura a autenticação na requisição
func (c *APIClient) setAuthentication(req *http.Request, auth *APIAuth) error {
	switch strings.ToLower(auth.Type) {
	case "basic":
		if auth.Username == "" || auth.Password == "" {
			return fmt.Errorf("usuário e senha são necessários para autenticação basic")
		}
		authStr := base64.StdEncoding.EncodeToString([]byte(auth.Username + ":" + auth.Password))
		req.Header.Set("Authorization", "Basic "+authStr)

	case "bearer":
		if auth.Token == "" {
			return fmt.Errorf("token é necessário para autenticação bearer")
		}
		req.Header.Set("Authorization", "Bearer "+auth.Token)

	case "api_key":
		if auth.KeyName == "" || auth.KeyValue == "" {
			return fmt.Errorf("nome e valor da chave são necessários para autenticação api_key")
		}
		req.Header.Set(auth.KeyName, auth.KeyValue)

	case "custom":
		if auth.HeaderName == "" || auth.HeaderValue == "" {
			return fmt.Errorf("nome e valor do header são necessários para autenticação custom")
		}
		req.Header.Set(auth.HeaderName, auth.HeaderValue)

	default:
		return fmt.Errorf("tipo de autenticação não suportado: %s", auth.Type)
	}

	return nil
} 