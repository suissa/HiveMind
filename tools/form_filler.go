package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

// FormFillerImpl implementa a interface FormFiller
type FormFillerImpl struct {
	forms      map[string]Form
	formData   map[string]FormData
	dataPath   string
	statistics map[string]interface{}
	mu         sync.RWMutex
}

// NewFormFiller cria uma nova instância do FormFiller
func NewFormFiller() (*FormFillerImpl, error) {
	filler := &FormFillerImpl{
		forms:      make(map[string]Form),
		formData:   make(map[string]FormData),
		dataPath:   "form_data",
		statistics: make(map[string]interface{}),
	}

	// Criar diretório de dados se não existir
	if err := os.MkdirAll(filler.dataPath, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de dados: %v", err)
	}

	// Carregar formulários existentes
	if err := filler.loadForms(); err != nil {
		return nil, err
	}

	return filler, nil
}

// CreateForm cria um novo formulário
func (f *FormFillerImpl) CreateForm(form Form) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.forms[form.ID]; exists {
		return fmt.Errorf("formulário já existe: %s", form.ID)
	}

	form.CreatedAt = time.Now()
	form.UpdatedAt = time.Now()

	f.forms[form.ID] = form
	return f.saveForm(form.ID)
}

// UpdateForm atualiza um formulário existente
func (f *FormFillerImpl) UpdateForm(formID string, form Form) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.forms[formID]; !exists {
		return fmt.Errorf("formulário não encontrado: %s", formID)
	}

	form.UpdatedAt = time.Now()
	f.forms[formID] = form
	return f.saveForm(formID)
}

// DeleteForm remove um formulário
func (f *FormFillerImpl) DeleteForm(formID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.forms[formID]; !exists {
		return fmt.Errorf("formulário não encontrado: %s", formID)
	}

	delete(f.forms, formID)
	return os.Remove(filepath.Join(f.dataPath, formID+".json"))
}

// GetForm retorna um formulário específico
func (f *FormFillerImpl) GetForm(formID string) (*Form, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	form, exists := f.forms[formID]
	if !exists {
		return nil, fmt.Errorf("formulário não encontrado: %s", formID)
	}

	return &form, nil
}

// ListForms lista todos os formulários disponíveis
func (f *FormFillerImpl) ListForms() ([]Form, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	forms := make([]Form, 0, len(f.forms))
	for _, form := range f.forms {
		forms = append(forms, form)
	}
	return forms, nil
}

// FillForm preenche um formulário com dados
func (f *FormFillerImpl) FillForm(formID string, data map[string]interface{}, options FillOptions) (*FormData, error) {
	form, err := f.GetForm(formID)
	if err != nil {
		return nil, err
	}

	formData := &FormData{
		FormID:    formID,
		Data:      make(map[string]interface{}),
		Files:     make(map[string][]byte),
		Timestamp: time.Now(),
	}

	// Preencher campos
	for _, field := range form.Fields {
		// Pular campos configurados para serem ignorados
		if contains(options.SkipFields, field.ID) {
			continue
		}

		// Obter valor do campo
		value := f.getFieldValue(field, data, options)
		if value != nil {
			formData.Data[field.ID] = value
		}
	}

	// Validar se necessário
	if options.ValidateOnFill {
		if errors, err := f.ValidateForm(formID, *formData); err != nil {
			return formData, err
		} else {
			formData.Errors = errors
			formData.ValidatedAt = time.Now()
		}
	}

	// Salvar dados do formulário
	f.mu.Lock()
	f.formData[formID] = *formData
	f.mu.Unlock()

	return formData, nil
}

// ValidateForm valida os dados de um formulário
func (f *FormFillerImpl) ValidateForm(formID string, data FormData) ([]ValidationError, error) {
	form, err := f.GetForm(formID)
	if err != nil {
		return nil, err
	}

	errors := make([]ValidationError, 0)

	for _, field := range form.Fields {
		value, exists := data.Data[field.ID]

		// Verificar campo obrigatório
		if field.Required && (!exists || isEmpty(value)) {
			errors = append(errors, ValidationError{
				FieldID: field.ID,
				Message: "Campo obrigatório não preenchido",
				Rule:    "required",
			})
			continue
		}

		// Validar regras específicas
		for _, rule := range field.Validations {
			if err := f.validateRule(field, value, rule); err != nil {
				errors = append(errors, ValidationError{
					FieldID: field.ID,
					Message: err.Error(),
					Rule:    rule.Type,
				})
			}
		}
	}

	return errors, nil
}

// GetFormData retorna os dados preenchidos de um formulário
func (f *FormFillerImpl) GetFormData(formID string) (*FormData, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	data, exists := f.formData[formID]
	if !exists {
		return nil, fmt.Errorf("dados não encontrados para o formulário: %s", formID)
	}

	return &data, nil
}

// ExportFormData exporta os dados do formulário em diferentes formatos
func (f *FormFillerImpl) ExportFormData(formID string, format string) ([]byte, error) {
	data, err := f.GetFormData(formID)
	if err != nil {
		return nil, err
	}

	switch format {
	case "json":
		return json.MarshalIndent(data, "", "  ")
	case "csv":
		return f.exportToCSV(data)
	case "xml":
		return f.exportToXML(data)
	default:
		return nil, fmt.Errorf("formato não suportado: %s", format)
	}
}

// GetStatistics retorna estatísticas do preenchedor
func (f *FormFillerImpl) GetStatistics() (map[string]interface{}, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	stats := make(map[string]interface{})
	for k, v := range f.statistics {
		stats[k] = v
	}
	return stats, nil
}

// Funções auxiliares

func (f *FormFillerImpl) loadForms() error {
	files, err := os.ReadDir(f.dataPath)
	if err != nil {
		return fmt.Errorf("erro ao ler diretório de dados: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(f.dataPath, file.Name()))
		if err != nil {
			return fmt.Errorf("erro ao ler arquivo %s: %v", file.Name(), err)
		}

		var form Form
		if err := json.Unmarshal(data, &form); err != nil {
			return fmt.Errorf("erro ao decodificar formulário %s: %v", file.Name(), err)
		}

		f.forms[form.ID] = form
	}

	return nil
}

func (f *FormFillerImpl) saveForm(formID string) error {
	form := f.forms[formID]
	data, err := json.MarshalIndent(form, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao codificar formulário: %v", err)
	}

	filename := filepath.Join(f.dataPath, formID+".json")
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("erro ao salvar formulário: %v", err)
	}

	return nil
}

func (f *FormFillerImpl) getFieldValue(field FormField, data map[string]interface{}, options FillOptions) interface{} {
	// Tentar obter valor dos dados fornecidos
	if value, exists := data[field.ID]; exists {
		return value
	}

	// Tentar obter valor padrão das opções
	if value, exists := options.DefaultValues[field.ID]; exists {
		return value
	}

	// Usar valor padrão do campo
	if field.DefaultValue != nil {
		return field.DefaultValue
	}

	// Auto-completar se habilitado
	if options.AutoComplete {
		return f.autoCompleteField(field, options.Locale)
	}

	return nil
}

func (f *FormFillerImpl) validateRule(field FormField, value interface{}, rule ValidationRule) error {
	switch rule.Type {
	case "required":
		if isEmpty(value) {
			return fmt.Errorf("campo obrigatório")
		}
	case "min":
		if !validateMin(value, rule.Value) {
			return fmt.Errorf("valor menor que o mínimo permitido")
		}
	case "max":
		if !validateMax(value, rule.Value) {
			return fmt.Errorf("valor maior que o máximo permitido")
		}
	case "regex":
		if !validateRegex(value, rule.Value.(string)) {
			return fmt.Errorf("valor não corresponde ao padrão esperado")
		}
	case "email":
		if !validateEmail(value.(string)) {
			return fmt.Errorf("email inválido")
		}
	case "phone":
		if !validatePhone(value.(string)) {
			return fmt.Errorf("telefone inválido")
		}
	}

	return nil
}

func (f *FormFillerImpl) autoCompleteField(field FormField, locale string) interface{} {
	switch field.Type {
	case FieldTypeEmail:
		return "usuario@exemplo.com.br"
	case FieldTypePhone:
		return "(11) 99999-9999"
	case FieldTypeDate:
		return time.Now().Format("2006-01-02")
	case FieldTypeNumber:
		return 0
	case FieldTypeText:
		return "Texto exemplo"
	}
	return nil
}

func (f *FormFillerImpl) exportToCSV(data *FormData) ([]byte, error) {
	// Implementar exportação para CSV
	return nil, fmt.Errorf("exportação para CSV não implementada")
}

func (f *FormFillerImpl) exportToXML(data *FormData) ([]byte, error) {
	// Implementar exportação para XML
	return nil, fmt.Errorf("exportação para XML não implementada")
}

// Funções utilitárias

func isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	switch v := value.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	}
	return false
}

func validateMin(value, min interface{}) bool {
	switch v := value.(type) {
	case int:
		return float64(v) >= min.(float64)
	case float64:
		return v >= min.(float64)
	case string:
		return len(v) >= int(min.(float64))
	}
	return false
}

func validateMax(value, max interface{}) bool {
	switch v := value.(type) {
	case int:
		return float64(v) <= max.(float64)
	case float64:
		return v <= max.(float64)
	case string:
		return len(v) <= int(max.(float64))
	}
	return false
}

func validateRegex(value interface{}, pattern string) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	match, err := regexp.MatchString(pattern, str)
	return err == nil && match
}

func validateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return validateRegex(email, pattern)
}

func validatePhone(phone string) bool {
	pattern := `^\(\d{2}\)\s\d{4,5}-\d{4}$`
	return validateRegex(phone, pattern)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
} 