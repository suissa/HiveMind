package tools

// Vulnerability representa uma vulnerabilidade encontrada
type Vulnerability struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"` // "CRITICAL", "HIGH", "MEDIUM", "LOW", "INFO"
	Type        string  `json:"type"`
	Target      string  `json:"target"`
	Port        int     `json:"port,omitempty"`
	Protocol    string  `json:"protocol,omitempty"`
	CVE         string  `json:"cve,omitempty"`
	CVSS        float64 `json:"cvss,omitempty"`
	Solution    string  `json:"solution,omitempty"`
	References  []string `json:"references,omitempty"`
}

// ScanResult representa o resultado de uma varredura
type ScanResult struct {
	Target           string          `json:"target"`
	StartTime        string          `json:"start_time"`
	EndTime          string          `json:"end_time"`
	Duration         string          `json:"duration"`
	Vulnerabilities  []Vulnerability `json:"vulnerabilities"`
	OpenPorts        []int           `json:"open_ports"`
	Services         []Service       `json:"services"`
	OSInfo           OSInfo          `json:"os_info"`
	Error            string          `json:"error,omitempty"`
}

// Service representa um serviço detectado
type Service struct {
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	Name        string `json:"name"`
	Version     string `json:"version,omitempty"`
	Banner      string `json:"banner,omitempty"`
	CPE         string `json:"cpe,omitempty"`
}

// OSInfo representa informações do sistema operacional
type OSInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Family      string `json:"family"`
	CPE         string `json:"cpe,omitempty"`
	Confidence  int    `json:"confidence"`
}

// ScanOptions representa as opções de varredura
type ScanOptions struct {
	Target            string   `json:"target"`           // IP ou hostname (opcional, usa localhost se vazio)
	Ports            []string `json:"ports,omitempty"`  // Portas específicas para scan
	Timeout          int      `json:"timeout"`          // Timeout em segundos
	Aggressive       bool     `json:"aggressive"`       // Modo agressivo de scan
	ServiceDetection bool     `json:"service_detection"` // Detectar versões de serviços
	Scripts          []string `json:"scripts,omitempty"` // Scripts NSE específicos para executar
	ExcludePorts    []string `json:"exclude_ports,omitempty"` // Portas para excluir
	Categories      []string `json:"categories,omitempty"`    // Categorias de vulnerabilidades
}

// SecurityScanner é a interface que todas as ferramentas de scan devem implementar
type SecurityScanner interface {
	// Scan realiza uma varredura de segurança
	Scan(options ScanOptions) (*ScanResult, error)
	
	// GetVulnerabilityDatabase retorna informações sobre a base de vulnerabilidades
	GetVulnerabilityDatabase() (map[string]interface{}, error)
	
	// UpdateVulnerabilityDatabase atualiza a base de vulnerabilidades
	UpdateVulnerabilityDatabase() error
	
	// GetSupportedScripts retorna os scripts de varredura suportados
	GetSupportedScripts() []string
	
	// GetSupportedCategories retorna as categorias de vulnerabilidades suportadas
	GetSupportedCategories() []string
}

// Scan básico
scanner, err := NewNmapScanner()
result, err := scanner.Scan(ScanOptions{
    Timeout: 300,
    ServiceDetection: true,
})

// Scan avançado
result, err := scanner.Scan(ScanOptions{
    Target:     "192.168.1.1", // opcional, usa localhost se vazio
    Ports:      []string{"80", "443", "22"},
    Aggressive: true,
    Scripts:    []string{"vuln", "auth"},
}) 