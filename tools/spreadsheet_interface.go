package tools

// CellValue representa o valor de uma célula com seu tipo
type CellValue struct {
	Type  string      `json:"type"`  // string, number, boolean, date, error
	Value interface{} `json:"value"` // valor da célula
}

// SheetData representa os dados de uma planilha
type SheetData struct {
	Name     string               `json:"name"`
	Headers  []string            `json:"headers,omitempty"`
	Rows     [][]CellValue      `json:"rows"`
	Metadata map[string]string   `json:"metadata,omitempty"`
}

// SpreadsheetResult representa o resultado do processamento de uma planilha
type SpreadsheetResult struct {
	Sheets []SheetData         `json:"sheets"`
	Error  string             `json:"error,omitempty"`
}

// SpreadsheetOptions representa as opções de processamento da planilha
type SpreadsheetOptions struct {
	FilePath     string   `json:"file_path"`
	SheetNames   []string `json:"sheet_names,omitempty"`    // Se vazio, processa todas as planilhas
	HasHeaders   bool     `json:"has_headers"`              // Se true, primeira linha é cabeçalho
	SkipRows     int      `json:"skip_rows,omitempty"`      // Número de linhas para pular
	MaxRows      int      `json:"max_rows,omitempty"`       // Máximo de linhas para ler (0 = sem limite)
	Delimiter    string   `json:"delimiter,omitempty"`       // Para arquivos CSV
	DateFormat   string   `json:"date_format,omitempty"`    // Formato de data para parse
	NumberFormat string   `json:"number_format,omitempty"`  // Formato de número para parse
}

// SpreadsheetTool é a interface que todas as ferramentas de processamento de planilha devem implementar
type SpreadsheetTool interface {
	Process(options SpreadsheetOptions) (*SpreadsheetResult, error)
} 