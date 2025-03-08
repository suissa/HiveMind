package tools

import "time"

// FieldType define os tipos de campos suportados
type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypeNumber   FieldType = "number"
	FieldTypeEmail    FieldType = "email"
	FieldTypePhone    FieldType = "phone"
	FieldTypeDate     FieldType = "date"
	FieldTypeSelect   FieldType = "select"
	FieldTypeCheckbox FieldType = "checkbox"
	FieldTypeRadio    FieldType = "radio"
	FieldTypeFile     FieldType = "file"
	FieldTypeTextarea FieldType = "textarea"
)

// ValidationRule define uma regra de validação
type ValidationRule struct {
	Type        string      `json:"type"`         // "required", "min", "max", "regex", etc
	Value       interface{} `json:"value"`        // Valor para comparação
	Message     string      `json:"message"`      // Mensagem de erro
	IsCustom    bool        `json:"is_custom"`    // Se é uma validação customizada
	CustomFunc  string      `json:"custom_func"`  // Nome da função customizada
}

// FormField representa um campo do formulário
type FormField struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Label        string           `json:"label"`
	Type         FieldType        `json:"type"`
	Value        interface{}      `json:"value"`
	DefaultValue interface{}      `json:"default_value"`
	Options      []FieldOption    `json:"options,omitempty"`
	Validations  []ValidationRule `json:"validations,omitempty"`
	Required     bool             `json:"required"`
	Disabled     bool             `json:"disabled"`
	Placeholder  string           `json:"placeholder,omitempty"`
	Mask         string           `json:"mask,omitempty"`
	Group        string           `json:"group,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// FieldOption representa uma opção para campos select, radio, etc
type FieldOption struct {
	Value    interface{} `json:"value"`
	Label    string      `json:"label"`
	Selected bool        `json:"selected"`
	Disabled bool        `json:"disabled"`
}

// Form representa um formulário completo
type Form struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Fields      []FormField           `json:"fields"`
	Groups      []FormGroup           `json:"groups,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// FormGroup representa um grupo de campos relacionados
type FormGroup struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Fields      []string `json:"fields"` // IDs dos campos
}

// ValidationError representa um erro de validação
type ValidationError struct {
	FieldID string `json:"field_id"`
	Message string `json:"message"`
	Rule    string `json:"rule"`
}

// FormData representa os dados preenchidos de um formulário
type FormData struct {
	FormID      string                 `json:"form_id"`
	Data        map[string]interface{} `json:"data"`
	Files       map[string][]byte      `json:"files,omitempty"`
	Timestamp   time.Time             `json:"timestamp"`
	ValidatedAt time.Time             `json:"validated_at,omitempty"`
	Errors      []ValidationError     `json:"errors,omitempty"`
}

// FillOptions representa as opções para preenchimento
type FillOptions struct {
	ValidateOnFill bool                   `json:"validate_on_fill"`
	AutoComplete   bool                   `json:"auto_complete"`
	DefaultValues  map[string]interface{} `json:"default_values,omitempty"`
	SkipFields     []string              `json:"skip_fields,omitempty"`
	Locale         string                `json:"locale"`
}

// FormFiller é a interface que todas as ferramentas de preenchimento devem implementar
type FormFiller interface {
	// CreateForm cria um novo formulário
	CreateForm(form Form) error

	// UpdateForm atualiza um formulário existente
	UpdateForm(formID string, form Form) error

	// DeleteForm remove um formulário
	DeleteForm(formID string) error

	// GetForm retorna um formulário específico
	GetForm(formID string) (*Form, error)

	// ListForms lista todos os formulários disponíveis
	ListForms() ([]Form, error)

	// FillForm preenche um formulário com dados
	FillForm(formID string, data map[string]interface{}, options FillOptions) (*FormData, error)

	// ValidateForm valida os dados de um formulário
	ValidateForm(formID string, data FormData) ([]ValidationError, error)

	// GetFormData retorna os dados preenchidos de um formulário
	GetFormData(formID string) (*FormData, error)

	// ExportFormData exporta os dados do formulário em diferentes formatos
	ExportFormData(formID string, format string) ([]byte, error)

	// GetStatistics retorna estatísticas do preenchedor
	GetStatistics() (map[string]interface{}, error)
} 