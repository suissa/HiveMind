package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// SpacyNER implementa a interface NERTool usando Spacy
type SpacyNER struct {
	apiURL string
	client *http.Client
}

// NewSpacyNER cria uma nova instância do SpacyNER
func NewSpacyNER() (*SpacyNER, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar arquivo .env: %v", err)
	}

	apiURL := os.Getenv("SPACY_API_URL")
	if apiURL == "" {
		return nil, fmt.Errorf("SPACY_API_URL não encontrado no arquivo .env")
	}

	return &SpacyNER{
		apiURL: apiURL,
		client: &http.Client{},
	}, nil
}

// ExtractEntities extrai entidades nomeadas do texto usando Spacy
func (s *SpacyNER) ExtractEntities(options NEROptions) (*NERResult, error) {
	// Preparar payload
	payload := map[string]interface{}{
		"text":     options.Text,
		"language": options.Language,
	}

	if len(options.Types) > 0 {
		payload["entity_types"] = options.Types
	}
	if options.MinScore > 0 {
		payload["min_score"] = options.MinScore
	}

	// Converter payload para JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter payload para JSON: %v", err)
	}

	// Fazer requisição para a API
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/ner", s.apiURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro na API: status %d", resp.StatusCode)
	}

	// Decodificar resposta
	var result NERResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	// Normalizar entidades se solicitado
	if options.Normalize {
		for i := range result.Entities {
			result.Entities[i].Normalized = normalizeEntity(result.Entities[i])
		}
	}

	return &result, nil
}

// GetSupportedEntityTypes retorna os tipos de entidades suportados
func (s *SpacyNER) GetSupportedEntityTypes() []string {
	return []string{
		"PERSON",      // Pessoas
		"ORG",         // Organizações
		"LOC",         // Localizações
		"GPE",         // Países, cidades, estados
		"PRODUCT",     // Produtos
		"EVENT",       // Eventos
		"WORK_OF_ART", // Obras de arte, livros, etc
		"LAW",         // Leis e documentos legais
		"LANGUAGE",    // Idiomas
		"DATE",        // Datas
		"TIME",        // Horários
		"MONEY",       // Valores monetários
		"PERCENT",     // Porcentagens
		"QUANTITY",    // Quantidades
	}
}

// GetSupportedLanguages retorna os idiomas suportados
func (s *SpacyNER) GetSupportedLanguages() []string {
	return []string{
		"pt",    // Português
		"en",    // Inglês
		"es",    // Espanhol
		"fr",    // Francês
		"de",    // Alemão
		"it",    // Italiano
		"nl",    // Holandês
		"el",    // Grego
		"xx",    // Multi-idioma
	}
}

// normalizeEntity normaliza o texto da entidade
func normalizeEntity(entity Entity) string {
	text := strings.TrimSpace(entity.Text)
	text = strings.ToLower(text)

	switch entity.Type {
	case "PERSON":
		// Capitalizar nomes próprios
		words := strings.Fields(text)
		for i, word := range words {
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[0:1]) + word[1:]
			}
		}
		text = strings.Join(words, " ")
	
	case "ORG":
		// Remover sufixos comuns de organizações
		suffixes := []string{" ltda", " s.a.", " inc", " corp"}
		for _, suffix := range suffixes {
			text = strings.TrimSuffix(text, suffix)
		}
		text = strings.ToUpper(text)

	case "DATE":
		// Padronizar formato de data (simplificado)
		// Aqui você pode adicionar mais lógica de normalização de datas
		text = strings.ReplaceAll(text, "/", "-")
	
	case "MONEY":
		// Padronizar valores monetários
		text = strings.ReplaceAll(text, "R$", "")
		text = strings.ReplaceAll(text, " ", "")
	}

	return text
} 