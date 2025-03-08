package tools

// SearchResult representa o resultado de uma busca
type SearchResult struct {
	Answer      string   `json:"answer,omitempty"`
	RawContent  string   `json:"raw_content,omitempty"`
	URLs        []string `json:"urls,omitempty"`
	Images      []string `json:"images,omitempty"`
	Error       string   `json:"error,omitempty"`
}

// SearchOptions representa as opções comuns de busca
type SearchOptions struct {
	Query                  string   `json:"query"`
	MaxResults            int      `json:"max_results"`
	IncludeAnswer         bool     `json:"include_answer"`
	IncludeRawContent     bool     `json:"include_raw_content"`
	IncludeImages         bool     `json:"include_images"`
	IncludeDomains        []string `json:"include_domains"`
	ExcludeDomains        []string `json:"exclude_domains"`
}

// SearchTool é a interface que todas as ferramentas de busca devem implementar
type SearchTool interface {
	Search(options SearchOptions) (*SearchResult, error)
}

// SearchEngineResult representa o resultado de uma operação de busca
type SearchEngineResult struct {
	Hits          []map[string]interface{} `json:"hits,omitempty"`
	Total         int                      `json:"total"`
	ProcessingTime int64                   `json:"processing_time_ms"`
	Query         string                   `json:"query"`
	Error         string                   `json:"error,omitempty"`
}

// IndexResult representa o resultado de uma operação de indexação
type IndexResult struct {
	TaskID        int64  `json:"taskId"`
	IndexUID      string `json:"indexUid"`
	Status        string `json:"status"`
	ProcessedDocs int    `json:"processed_documents,omitempty"`
	Error         string `json:"error,omitempty"`
}

// SearchEngineOptions representa as opções para operações de busca
type SearchEngineOptions struct {
	IndexName    string                 `json:"index_name"`
	Query        string                 `json:"query,omitempty"`
	Filters      map[string]interface{} `json:"filters,omitempty"`
	Sort         []string              `json:"sort,omitempty"`
	Limit        int                   `json:"limit,omitempty"`
	Offset       int                   `json:"offset,omitempty"`
	PrimaryKey   string                `json:"primary_key,omitempty"`
	SearchParams map[string]interface{} `json:"search_params,omitempty"`
}

// SearchEngineTool é a interface que todas as ferramentas de busca devem implementar
type SearchEngineTool interface {
	Search(options SearchEngineOptions) (*SearchEngineResult, error)
	Index(indexName string, documents interface{}, primaryKey string) (*IndexResult, error)
	DeleteIndex(indexName string) error
	GetStats(indexName string) (map[string]interface{}, error)
}

// SemanticSearchResult representa o resultado de uma busca semântica
type SemanticSearchResult struct {
	Results        []SemanticDocument `json:"results"`
	Total          int               `json:"total"`
	ProcessingTime int64             `json:"processing_time_ms"`
	Vector         []float32         `json:"vector,omitempty"`
	Error          string            `json:"error,omitempty"`
}

// SemanticDocument representa um documento no resultado da busca
type SemanticDocument struct {
	ID           string                 `json:"id"`
	Class        string                 `json:"class"`
	Properties   map[string]interface{} `json:"properties"`
	Score        float64               `json:"score"`
	Vector       []float32             `json:"vector,omitempty"`
	Distance     float64               `json:"distance,omitempty"`
}

// SemanticSearchOptions representa as opções para busca semântica
type SemanticSearchOptions struct {
	Class            string                 `json:"class"`
	Query            string                 `json:"query,omitempty"`
	Vector           []float32             `json:"vector,omitempty"`
	Properties       []string              `json:"properties,omitempty"`
	Filters          map[string]interface{} `json:"filters,omitempty"`
	Limit            int                   `json:"limit,omitempty"`
	Offset           int                   `json:"offset,omitempty"`
	NearVector       []float32             `json:"near_vector,omitempty"`
	NearObject       string                `json:"near_object,omitempty"`
	Distance         float64               `json:"distance,omitempty"`
	IncludeVector    bool                  `json:"include_vector"`
	ConsistencyLevel string                `json:"consistency_level,omitempty"`
}

// SemanticSearchTool é a interface que todas as ferramentas de busca semântica devem implementar
type SemanticSearchTool interface {
	Search(options SemanticSearchOptions) (*SemanticSearchResult, error)
	AddDocument(class string, properties map[string]interface{}, vector []float32) error
	DeleteDocument(class string, id string) error
	CreateClass(class string, properties map[string]interface{}) error
	DeleteClass(class string) error
} 