package tools

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// SpreadsheetProcessor implementa a interface SpreadsheetTool
type SpreadsheetProcessor struct{}

// NewSpreadsheetProcessor cria uma nova instância do SpreadsheetProcessor
func NewSpreadsheetProcessor() *SpreadsheetProcessor {
	return &SpreadsheetProcessor{}
}

// Process implementa a interface SpreadsheetTool
func (p *SpreadsheetProcessor) Process(options SpreadsheetOptions) (*SpreadsheetResult, error) {
	result := &SpreadsheetResult{
		Sheets: make([]SheetData, 0),
	}

	// Verificar se o arquivo existe
	if _, err := os.Stat(options.FilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("arquivo não encontrado: %s", options.FilePath)
	}

	// Determinar o tipo de arquivo pela extensão
	ext := strings.ToLower(filepath.Ext(options.FilePath))
	switch ext {
	case ".xlsx", ".xlsm", ".xls":
		return p.processExcel(options)
	case ".csv":
		return p.processCSV(options)
	default:
		return nil, fmt.Errorf("formato de arquivo não suportado: %s", ext)
	}
}

// processExcel processa arquivos Excel (XLSX, XLS)
func (p *SpreadsheetProcessor) processExcel(options SpreadsheetOptions) (*SpreadsheetResult, error) {
	result := &SpreadsheetResult{
		Sheets: make([]SheetData, 0),
	}

	// Abrir arquivo Excel
	f, err := excelize.OpenFile(options.FilePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo Excel: %v", err)
	}
	defer f.Close()

	// Obter lista de planilhas
	sheets := f.GetSheetList()

	// Filtrar planilhas se especificado
	if len(options.SheetNames) > 0 {
		sheetMap := make(map[string]bool)
		for _, name := range options.SheetNames {
			sheetMap[name] = true
		}
		filteredSheets := make([]string, 0)
		for _, sheet := range sheets {
			if sheetMap[sheet] {
				filteredSheets = append(filteredSheets, sheet)
			}
		}
		sheets = filteredSheets
	}

	// Processar cada planilha
	for _, sheetName := range sheets {
		// Obter todas as células da planilha
		rows, err := f.GetRows(sheetName)
		if err != nil {
			continue
		}

		sheetData := SheetData{
			Name:     sheetName,
			Headers:  make([]string, 0),
			Rows:     make([][]CellValue, 0),
			Metadata: make(map[string]string),
		}

		// Processar linhas
		startRow := options.SkipRows
		if options.HasHeaders {
			if len(rows) > startRow {
				sheetData.Headers = rows[startRow]
				startRow++
			}
		}

		// Determinar número máximo de linhas
		endRow := len(rows)
		if options.MaxRows > 0 && startRow+options.MaxRows < endRow {
			endRow = startRow + options.MaxRows
		}

		// Processar células
		for rowIndex := startRow; rowIndex < endRow; rowIndex++ {
			if rowIndex >= len(rows) {
				break
			}

			row := rows[rowIndex]
			processedRow := make([]CellValue, len(row))

			for colIndex, cellValue := range row {
				cell := p.processCellValue(cellValue, options)
				processedRow[colIndex] = cell
			}

			sheetData.Rows = append(sheetData.Rows, processedRow)
		}

		result.Sheets = append(result.Sheets, sheetData)
	}

	return result, nil
}

// processCSV processa arquivos CSV
func (p *SpreadsheetProcessor) processCSV(options SpreadsheetOptions) (*SpreadsheetResult, error) {
	result := &SpreadsheetResult{
		Sheets: make([]SheetData, 0),
	}

	// Abrir arquivo CSV
	file, err := os.Open(options.FilePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo CSV: %v", err)
	}
	defer file.Close()

	// Configurar leitor CSV
	reader := csv.NewReader(file)
	if options.Delimiter != "" {
		reader.Comma = rune(options.Delimiter[0])
	}

	// Ler todas as linhas
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo CSV: %v", err)
	}

	sheetData := SheetData{
		Name:     "Sheet1",
		Headers:  make([]string, 0),
		Rows:     make([][]CellValue, 0),
		Metadata: make(map[string]string),
	}

	// Processar linhas
	startRow := options.SkipRows
	if options.HasHeaders && len(rows) > startRow {
		sheetData.Headers = rows[startRow]
		startRow++
	}

	// Determinar número máximo de linhas
	endRow := len(rows)
	if options.MaxRows > 0 && startRow+options.MaxRows < endRow {
		endRow = startRow + options.MaxRows
	}

	// Processar células
	for rowIndex := startRow; rowIndex < endRow; rowIndex++ {
		if rowIndex >= len(rows) {
			break
		}

		row := rows[rowIndex]
		processedRow := make([]CellValue, len(row))

		for colIndex, cellValue := range row {
			cell := p.processCellValue(cellValue, options)
			processedRow[colIndex] = cell
		}

		sheetData.Rows = append(sheetData.Rows, processedRow)
	}

	result.Sheets = append(result.Sheets, sheetData)
	return result, nil
}

// processCellValue processa o valor de uma célula e determina seu tipo
func (p *SpreadsheetProcessor) processCellValue(value string, options SpreadsheetOptions) CellValue {
	// Tentar converter para número
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		return CellValue{
			Type:  "number",
			Value: num,
		}
	}

	// Tentar converter para booleano
	lower := strings.ToLower(value)
	if lower == "true" || lower == "false" {
		return CellValue{
			Type:  "boolean",
			Value: lower == "true",
		}
	}

	// Tentar converter para data
	if options.DateFormat != "" {
		if date, err := time.Parse(options.DateFormat, value); err == nil {
			return CellValue{
				Type:  "date",
				Value: date,
			}
		}
	}

	// Se nenhuma conversão funcionar, retornar como string
	return CellValue{
		Type:  "string",
		Value: value,
	}
} 