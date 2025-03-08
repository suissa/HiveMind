package tools

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// ExampleScanner demonstra o uso do scanner de vulnerabilidades
func ExampleScanner() {
	// Criar scanner
	scanner, err := NewNmapScanner()
	if err != nil {
		log.Fatal(err)
	}

	// Scan básico do localhost
	result, err := scanner.Scan(ScanOptions{
		Timeout:          300,
		ServiceDetection: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Scan concluído em %s\n", result.Duration)
	fmt.Printf("Alvo: %s\n\n", result.Target)

	// Exibir portas abertas e serviços
	fmt.Println("Portas abertas e serviços:")
	for _, service := range result.Services {
		fmt.Printf("- Porta %d (%s): %s", service.Port, service.Protocol, service.Name)
		if service.Version != "" {
			fmt.Printf(" (%s)", service.Version)
		}
		fmt.Println()
	}

	// Exibir informações do SO
	if result.OSInfo.Name != "" {
		fmt.Printf("\nSistema Operacional: %s (Família: %s)\n",
			result.OSInfo.Name, result.OSInfo.Family)
	}

	// Exibir vulnerabilidades encontradas
	if len(result.Vulnerabilities) > 0 {
		fmt.Println("\nVulnerabilidades encontradas:")
		for _, vuln := range result.Vulnerabilities {
			fmt.Printf("\n[%s] %s\n", vuln.Severity, vuln.Name)
			fmt.Printf("ID: %s (CVSS: %.1f)\n", vuln.ID, vuln.CVSS)
			fmt.Printf("Descrição: %s\n", vuln.Description)
			if vuln.Solution != "" {
				fmt.Printf("Solução: %s\n", vuln.Solution)
			}
		}
	}

	// Salvar resultado em JSON
	saveResult(result, "scan_result.json")
}

// ExampleScannerAdvanced demonstra recursos avançados do scanner
func ExampleScannerAdvanced() {
	scanner, err := NewNmapScanner()
	if err != nil {
		log.Fatal(err)
	}

	// Atualizar base de vulnerabilidades
	fmt.Println("Atualizando base de vulnerabilidades...")
	if err := scanner.UpdateVulnerabilityDatabase(); err != nil {
		log.Fatal(err)
	}

	// Exibir informações da base
	info, err := scanner.GetVulnerabilityDatabase()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Base de vulnerabilidades atualizada em: %s\n", info["last_update"])

	// Scan avançado com scripts específicos
	result, err := scanner.Scan(ScanOptions{
		Target:           "localhost",
		Ports:           []string{"80", "443", "3306", "22"},
		Timeout:         600,
		Aggressive:      true,
		ServiceDetection: true,
		Scripts:         []string{"vuln", "auth", "ssl-enum"},
		Categories:      []string{"critical", "high"},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Agrupar vulnerabilidades por severidade
	vulnsBySeverity := make(map[string][]Vulnerability)
	for _, vuln := range result.Vulnerabilities {
		vulnsBySeverity[vuln.Severity] = append(vulnsBySeverity[vuln.Severity], vuln)
	}

	// Exibir resultados agrupados
	fmt.Printf("\nResultados do scan (%s):\n", result.Duration)
	for severity, vulns := range vulnsBySeverity {
		fmt.Printf("\n[%s] - %d vulnerabilidades encontradas:\n", severity, len(vulns))
		for _, vuln := range vulns {
			fmt.Printf("- %s (CVE: %s)\n", vuln.Name, vuln.CVE)
			fmt.Printf("  Porta: %d/%s\n", vuln.Port, vuln.Protocol)
			fmt.Printf("  CVSS: %.1f\n", vuln.CVSS)
		}
	}

	// Salvar resultado detalhado
	saveResult(result, "scan_result_detailed.json")
}

// saveResult salva o resultado do scan em um arquivo JSON
func saveResult(result *ScanResult, filename string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao converter para JSON: %v", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %v", err)
	}

	fmt.Printf("\nResultado salvo em %s\n", filename)
	return nil
} 