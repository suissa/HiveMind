package tools

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"

	"github.com/gen2brain/go-fitz"
	"github.com/otiai10/gosseract/v2"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

// PDFProcessor implementa a interface PDFTool
type PDFProcessor struct {
	tesseract *gosseract.Client
}

// NewPDFProcessor cria uma nova instância do PDFProcessor
func NewPDFProcessor() (*PDFProcessor, error) {
	client := gosseract.NewClient()
	
	return &PDFProcessor{
		tesseract: client,
	}, nil
}

// Close libera os recursos do processador
func (p *PDFProcessor) Close() error {
	if p.tesseract != nil {
		return p.tesseract.Close()
	}
	return nil
}

// Process implementa a interface PDFTool
func (p *PDFProcessor) Process(options PDFOptions) (*PDFResult, error) {
	result := &PDFResult{
		Text:     "",
		Pages:    make([]string, 0),
		Metadata: make(map[string]string),
		Images:   make([]PDFImage, 0),
	}

	// Verificar se o arquivo existe
	if _, err := os.Stat(options.FilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("arquivo não encontrado: %s", options.FilePath)
	}

	// Extrair metadados do PDF
	ctx := pdfcpu.NewDefaultConfiguration()
	meta, err := api.Info(options.FilePath, ctx)
	if err == nil {
		result.Metadata = meta.Values
	}

	// Abrir o documento com go-fitz para extração de texto e imagens
	doc, err := fitz.New(options.FilePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir PDF: %v", err)
	}
	defer doc.Close()

	// Determinar páginas a processar
	startPage := options.StartPage
	if startPage < 1 {
		startPage = 1
	}
	endPage := options.EndPage
	if endPage <= 0 || endPage > doc.NumPages() {
		endPage = doc.NumPages()
	}

	// Configurar Tesseract se OCR estiver habilitado
	if options.UseOCR {
		if options.Language != "" {
			err = p.tesseract.SetLanguage(options.Language)
			if err != nil {
				return nil, fmt.Errorf("erro ao configurar idioma OCR: %v", err)
			}
		}
		if options.DPI > 0 {
			p.tesseract.SetSourceResolution(options.DPI)
		}
	}

	// Processar cada página
	for pageNum := startPage - 1; pageNum < endPage; pageNum++ {
		// Extrair texto da página
		text, err := doc.Text(pageNum)
		if err != nil {
			continue
		}

		// Adicionar texto à lista de páginas
		result.Pages = append(result.Pages, text)
		
		// Concatenar texto ao resultado geral
		if result.Text != "" {
			result.Text += "\n\n"
		}
		result.Text += text

		// Processar imagens se solicitado
		if options.ExtractImages || options.UseOCR {
			img, err := doc.Image(pageNum)
			if err != nil {
				continue
			}

			// Criar diretório temporário para salvar imagem
			tmpDir, err := os.MkdirTemp("", "pdf_images")
			if err != nil {
				continue
			}
			defer os.RemoveAll(tmpDir)

			// Salvar imagem temporariamente
			imgPath := filepath.Join(tmpDir, fmt.Sprintf("page_%d.png", pageNum+1))
			imgFile, err := os.Create(imgPath)
			if err != nil {
				continue
			}

			err = png.Encode(imgFile, img)
			imgFile.Close()
			if err != nil {
				continue
			}

			// Realizar OCR se solicitado
			var ocrText string
			if options.UseOCR {
				p.tesseract.SetImage(imgPath)
				ocrText, err = p.tesseract.Text()
				if err == nil && ocrText != "" {
					if result.Text != "" {
						result.Text += "\n\n"
					}
					result.Text += ocrText
				}
			}

			// Adicionar imagem ao resultado se solicitado
			if options.ExtractImages {
				imgData, err := os.ReadFile(imgPath)
				if err == nil {
					result.Images = append(result.Images, PDFImage{
						Page:     pageNum + 1,
						Data:     imgData,
						Text:     ocrText,
						MimeType: "image/png",
					})
				}
			}
		}
	}

	return result, nil
}

// extractImageFromPage extrai uma imagem de uma página específica
func (p *PDFProcessor) extractImageFromPage(img image.Image) ([]byte, error) {
	// Criar um buffer para armazenar a imagem
	var buf []byte
	writer := io.NewBuffer(buf)

	// Codificar a imagem como PNG
	err := png.Encode(writer, img)
	if err != nil {
		return nil, err
	}

	return writer.Bytes(), nil
} 