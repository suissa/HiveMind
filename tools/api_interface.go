package tools

import "time"

// APIResponse representa a resposta de uma requisição à API
type APIResponse struct {
	StatusCode    int               `json:"status_code"`
	Headers      map[string]string `json:"headers"`
	Body         interface{}       `json:"body"`
	RawBody      []byte           `json:"raw_body,omitempty"`
	ResponseTime time.Duration    `json:"response_time"`
	Error        string           `json:"error,omitempty"`
}

// APIAuth representa as opções de autenticação
type APIAuth struct {
	Type        string            `json:"type"`        // basic, bearer, api_key, custom
	Username    string            `json:"username,omitempty"`
	Password    string            `json:"password,omitempty"`
	Token       string            `json:"token,omitempty"`
	KeyName     string            `json:"key_name,omitempty"`
	KeyValue    string            `json:"key_value,omitempty"`
	HeaderName  string            `json:"header_name,omitempty"`
	HeaderValue string            `json:"header_value,omitempty"`
}

// APIOptions representa as opções para uma requisição à API
type APIOptions struct {
	Method      string            `json:"method"`       // GET, POST, PUT, DELETE, etc.
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers,omitempty"`
	QueryParams map[string]string `json:"query_params,omitempty"`
	Body        interface{}       `json:"body,omitempty"`
	Auth        *APIAuth          `json:"auth,omitempty"`
	Timeout     time.Duration    `json:"timeout,omitempty"`
	RetryCount  int              `json:"retry_count,omitempty"`
	RetryDelay  time.Duration    `json:"retry_delay,omitempty"`
}

// APITool é a interface que todas as ferramentas de requisição à API devem implementar
type APITool interface {
	Request(options APIOptions) (*APIResponse, error)
} 