package tools

// Entity representa uma entidade nomeada extraída do texto
type Entity struct {
	Text       string  `json:"text"`
	Type       string  `json:"type"`
	Start      int     `json:"start"`
	End        int     `json:"end"`
	Score      float64 `json:"score,omitempty"`
	Category   string  `json:"category,omitempty"`
	Normalized string  `json:"normalized,omitempty"`
}

// NERResult representa o resultado da extração de entidades
type NERResult struct {
	Entities      []Entity `json:"entities"`
	ProcessedText string   `json:"processed_text"`
	Language      string   `json:"language,omitempty"`
	Error         string   `json:"error,omitempty"`
}

// NEROptions representa as opções para extração de entidades
type NEROptions struct {
	Text            string   `json:"text"`
	Types           []string `json:"types,omitempty"`
	MinScore        float64  `json:"min_score,omitempty"`
	Language        string   `json:"language,omitempty"`
	Normalize       bool     `json:"normalize"`
	IncludeSpans    bool     `json:"include_spans"`
	IncludeSubtypes bool     `json:"include_subtypes"`
}

// NERTool é a interface que todas as ferramentas de NER devem implementar
type NERTool interface {
	ExtractEntities(options NEROptions) (*NERResult, error)
	GetSupportedEntityTypes() []string
	GetSupportedLanguages() []string
} 