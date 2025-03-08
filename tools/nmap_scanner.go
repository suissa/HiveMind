package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// NmapScanner implementa a interface SecurityScanner usando Nmap
type NmapScanner struct {
	nmapPath     string
	scriptsPath  string
	vulnDBPath   string
	lastUpdate   time.Time
}

// NewNmapScanner cria uma nova instância do NmapScanner
func NewNmapScanner() (*NmapScanner, error) {
	// Verificar se o Nmap está instalado
	nmapPath, err := exec.LookPath("nmap")
	if err != nil {
		return nil, fmt.Errorf("nmap não encontrado no sistema: %v", err)
	}

	scanner := &NmapScanner{
		nmapPath:    nmapPath,
		scriptsPath: "/usr/share/nmap/scripts",
		vulnDBPath:  "/usr/share/nmap/scripts/vulscan",
	}

	// Verificar se a base de vulnerabilidades está atualizada
	if err := scanner.checkVulnDB(); err != nil {
		return nil, err
	}

	return scanner, nil
}

// Scan realiza uma varredura de segurança
func (s *NmapScanner) Scan(options ScanOptions) (*ScanResult, error) {
	startTime := time.Now()

	// Usar localhost se nenhum alvo for especificado
	target := "localhost"
	if options.Target != "" {
		target = options.Target
	}

	// Construir comando Nmap base
	args := []string{
		"-sV",                // Detecção de versão
		"-sC",                // Scripts padrão
		"-O",                 // Detecção de SO
		"--stats-every", "5s", // Estatísticas a cada 5 segundos
		"--max-retries", "2",  // Máximo de tentativas
	}

	// Adicionar portas se especificadas
	if len(options.Ports) > 0 {
		args = append(args, "-p", strings.Join(options.Ports, ","))
	}

	// Modo agressivo
	if options.Aggressive {
		args = append(args, "-A", "-T4")
	}

	// Scripts específicos
	if len(options.Scripts) > 0 {
		args = append(args, "--script", strings.Join(options.Scripts, ","))
	}

	// Timeout
	if options.Timeout > 0 {
		args = append(args, "--host-timeout", fmt.Sprintf("%ds", options.Timeout))
	}

	// Adicionar alvo
	args = append(args, target)

	// Executar Nmap
	cmd := exec.Command(s.nmapPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("erro ao executar nmap: %v", err)
	}

	// Processar resultado
	result := &ScanResult{
		Target:    target,
		StartTime: startTime.Format(time.RFC3339),
		EndTime:   time.Now().Format(time.RFC3339),
		Duration:  time.Since(startTime).String(),
	}

	// Parsear saída do Nmap
	if err := s.parseNmapOutput(string(output), result); err != nil {
		return nil, err
	}

	// Adicionar vulnerabilidades encontradas
	if err := s.addVulnerabilities(result); err != nil {
		return nil, err
	}

	return result, nil
}

// parseNmapOutput processa a saída do Nmap
func (s *NmapScanner) parseNmapOutput(output string, result *ScanResult) error {
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		// Portas abertas
		if strings.Contains(line, "open") && strings.Contains(line, "/tcp") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				port, _ := strconv.Atoi(strings.Split(parts[0], "/")[0])
				result.OpenPorts = append(result.OpenPorts, port)

				// Adicionar serviço
				service := Service{
					Port:     port,
					Protocol: "tcp",
				}

				if len(parts) >= 3 {
					service.Name = parts[2]
				}
				if len(parts) >= 4 {
					service.Version = strings.Join(parts[3:], " ")
				}

				result.Services = append(result.Services, service)
			}
		}

		// Informações do SO
		if strings.Contains(line, "OS details:") {
			osDetails := strings.TrimPrefix(line, "OS details:")
			result.OSInfo = OSInfo{
				Name:       strings.TrimSpace(osDetails),
				Family:     detectOSFamily(osDetails),
				Confidence: 100,
			}
		}
	}

	return nil
}

// addVulnerabilities adiciona vulnerabilidades encontradas
func (s *NmapScanner) addVulnerabilities(result *ScanResult) error {
	// Para cada serviço, verificar vulnerabilidades conhecidas
	for _, service := range result.Services {
		vulns, err := s.checkServiceVulnerabilities(service)
		if err != nil {
			return err
		}
		result.Vulnerabilities = append(result.Vulnerabilities, vulns...)
	}

	return nil
}

// checkServiceVulnerabilities verifica vulnerabilidades para um serviço
func (s *NmapScanner) checkServiceVulnerabilities(service Service) ([]Vulnerability, error) {
	vulns := []Vulnerability{}

	// Verificar base de dados local
	if service.Version != "" {
		// Exemplo: verificar CVEs conhecidas para a versão do serviço
		cves, err := s.searchCVEs(service.Name, service.Version)
		if err != nil {
			return nil, err
		}

		for _, cve := range cves {
			vuln := Vulnerability{
				ID:          cve.ID,
				Name:        cve.Title,
				Description: cve.Description,
				Severity:    cve.Severity,
				Type:        "CVE",
				Target:      service.Name,
				Port:        service.Port,
				Protocol:    service.Protocol,
				CVE:        cve.ID,
				CVSS:       cve.CVSS,
				Solution:   cve.Solution,
				References: cve.References,
			}
			vulns = append(vulns, vuln)
		}
	}

	return vulns, nil
}

// GetVulnerabilityDatabase retorna informações sobre a base de vulnerabilidades
func (s *NmapScanner) GetVulnerabilityDatabase() (map[string]interface{}, error) {
	info := map[string]interface{}{
		"last_update": s.lastUpdate.Format(time.RFC3339),
		"location":    s.vulnDBPath,
		"scripts":     len(s.GetSupportedScripts()),
	}
	return info, nil
}

// UpdateVulnerabilityDatabase atualiza a base de vulnerabilidades
func (s *NmapScanner) UpdateVulnerabilityDatabase() error {
	// Atualizar scripts NSE
	cmd := exec.Command("nmap", "--script-updatedb")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erro ao atualizar scripts NSE: %v", err)
	}

	// Atualizar base de vulnerabilidades
	if err := s.updateVulnDB(); err != nil {
		return err
	}

	s.lastUpdate = time.Now()
	return nil
}

// GetSupportedScripts retorna os scripts de varredura suportados
func (s *NmapScanner) GetSupportedScripts() []string {
	return []string{
		"vuln",           // Vulnerabilidades conhecidas
		"auth",           // Autenticação
		"default",        // Scripts padrão
		"discovery",      // Descoberta de serviços
		"version",        // Detecção de versão
		"safe",           // Scripts seguros
		"exploit",        // Exploits conhecidos
		"brute",         // Força bruta
		"dos",           // Denial of Service
		"fuzzer",        // Fuzzing
		"malware",       // Detecção de malware
		"backdoor",      // Backdoors conhecidos
		"ssl-enum",      // Enumeração SSL/TLS
		"http-enum",     // Enumeração HTTP
		"smb-enum",      // Enumeração SMB
	}
}

// GetSupportedCategories retorna as categorias de vulnerabilidades suportadas
func (s *NmapScanner) GetSupportedCategories() []string {
	return []string{
		"critical",
		"high",
		"medium",
		"low",
		"info",
		"auth",
		"dos",
		"exploit",
		"fuzzer",
		"scan",
		"malware",
		"backdoor",
		"brute",
		"intrusive",
	}
}

// Funções auxiliares

func (s *NmapScanner) checkVulnDB() error {
	if _, err := os.Stat(s.vulnDBPath); os.IsNotExist(err) {
		return s.updateVulnDB()
	}
	return nil
}

func (s *NmapScanner) updateVulnDB() error {
	// Implementar atualização da base de vulnerabilidades
	// Por exemplo, baixar CVEs mais recentes
	return nil
}

func (s *NmapScanner) searchCVEs(service, version string) ([]struct {
	ID          string
	Title       string
	Description string
	Severity    string
	CVSS        float64
	Solution    string
	References  []string
}, error) {
	// Implementar busca de CVEs
	// Esta é uma implementação simplificada
	return []struct {
		ID          string
		Title       string
		Description string
		Severity    string
		CVSS        float64
		Solution    string
		References  []string
	}{}, nil
}

func detectOSFamily(osDetails string) string {
	osDetails = strings.ToLower(osDetails)
	if strings.Contains(osDetails, "linux") {
		return "Linux"
	} else if strings.Contains(osDetails, "windows") {
		return "Windows"
	} else if strings.Contains(osDetails, "mac") || strings.Contains(osDetails, "darwin") {
		return "MacOS"
	}
	return "Unknown"
} 