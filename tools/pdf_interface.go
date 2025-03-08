package tools

// PDFResult representa o resultado do processamento de um PDF
type PDFResult struct {
	Text       string            `json:"text,omitempty"`
	Pages      []string         `json:"pages,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Images     []PDFImage       `json:"images,omitempty"`
	Error      string           `json:"error,omitempty"`
}

// PDFImage representa uma imagem extraída do PDF
type PDFImage struct {
	Page     int    `json:"page"`
	Data     []byte `json:"data,omitempty"`
	Text     string `json:"text,omitempty"` // Texto extraído via OCR
	MimeType string `json:"mime_type,omitempty"`
}

// PDFOptions representa as opções de processamento do PDF
type PDFOptions struct {
	FilePath        string `json:"file_path"`
	UseOCR          bool   `json:"use_ocr"`
	ExtractImages   bool   `json:"extract_images"`
	StartPage       int    `json:"start_page,omitempty"`
	EndPage         int    `json:"end_page,omitempty"`
	Language        string `json:"language,omitempty"` // Idioma para OCR (ex: "por", "eng")
	DPI             int    `json:"dpi,omitempty"`     // DPI para OCR
}

// PDFTool é a interface que todas as ferramentas de processamento de PDF devem implementar
type PDFTool interface {
	Process(options PDFOptions) (*PDFResult, error)
} 