package tools

import (
	"fmt"
	"time"
)

// ExampleFormFiller demonstra o uso básico do preenchedor de formulários
func ExampleFormFiller() {
	// Criar nova instância do preenchedor
	filler, err := NewFormFiller()
	if err != nil {
		fmt.Printf("Erro ao criar preenchedor: %v\n", err)
		return
	}

	// Criar um formulário de exemplo (cadastro de cliente)
	form := Form{
		ID:          "cadastro_cliente",
		Name:        "Cadastro de Cliente",
		Description: "Formulário para cadastro de novos clientes",
		Fields: []FormField{
			{
				ID:          "nome",
				Name:        "nome",
				Label:       "Nome Completo",
				Type:        FieldTypeText,
				Required:    true,
				Placeholder: "Digite seu nome completo",
				Validations: []ValidationRule{
					{
						Type:    "min",
						Value:   3,
						Message: "Nome deve ter pelo menos 3 caracteres",
					},
				},
			},
			{
				ID:          "email",
				Name:        "email",
				Label:       "E-mail",
				Type:        FieldTypeEmail,
				Required:    true,
				Placeholder: "Digite seu e-mail",
				Validations: []ValidationRule{
					{
						Type:    "email",
						Message: "E-mail inválido",
					},
				},
			},
			{
				ID:          "telefone",
				Name:        "telefone",
				Label:       "Telefone",
				Type:        FieldTypePhone,
				Required:    true,
				Placeholder: "(00) 00000-0000",
				Mask:        "(##) #####-####",
				Validations: []ValidationRule{
					{
						Type:    "phone",
						Message: "Telefone inválido",
					},
				},
			},
			{
				ID:          "data_nascimento",
				Name:        "data_nascimento",
				Label:       "Data de Nascimento",
				Type:        FieldTypeDate,
				Required:    true,
				Validations: []ValidationRule{
					{
						Type:    "max",
						Value:   time.Now(),
						Message: "Data não pode ser futura",
					},
				},
			},
			{
				ID:    "tipo_cliente",
				Name:  "tipo_cliente",
				Label: "Tipo de Cliente",
				Type:  FieldTypeSelect,
				Options: []FieldOption{
					{Value: "pf", Label: "Pessoa Física", Selected: true},
					{Value: "pj", Label: "Pessoa Jurídica"},
				},
				Required: true,
			},
			{
				ID:       "newsletter",
				Name:     "newsletter",
				Label:    "Receber Newsletter",
				Type:     FieldTypeCheckbox,
				Required: false,
			},
		},
		Groups: []FormGroup{
			{
				ID:          "dados_pessoais",
				Name:        "Dados Pessoais",
				Description: "Informações básicas do cliente",
				Fields:      []string{"nome", "email", "telefone", "data_nascimento"},
			},
			{
				ID:          "preferencias",
				Name:        "Preferências",
				Description: "Configurações e preferências",
				Fields:      []string{"tipo_cliente", "newsletter"},
			},
		},
	}

	// Adicionar formulário
	if err := filler.CreateForm(form); err != nil {
		fmt.Printf("Erro ao criar formulário: %v\n", err)
		return
	}

	// Preencher formulário com dados
	data := map[string]interface{}{
		"nome":            "João da Silva",
		"email":          "joao@exemplo.com.br",
		"telefone":       "(11) 98765-4321",
		"data_nascimento": "1990-01-01",
		"tipo_cliente":   "pf",
		"newsletter":     true,
	}

	options := FillOptions{
		ValidateOnFill: true,
		AutoComplete:   true,
		Locale:        "pt-BR",
	}

	// Realizar preenchimento
	result, err := filler.FillForm(form.ID, data, options)
	if err != nil {
		fmt.Printf("Erro ao preencher formulário: %v\n", err)
		return
	}

	// Verificar erros de validação
	if len(result.Errors) > 0 {
		fmt.Println("\nErros de validação encontrados:")
		for _, err := range result.Errors {
			fmt.Printf("Campo %s: %s\n", err.FieldID, err.Message)
		}
		return
	}

	// Imprimir resultado
	fmt.Println("\nFormulário preenchido com sucesso!")
	fmt.Printf("ID do Formulário: %s\n", result.FormID)
	fmt.Printf("Data/Hora: %s\n", result.Timestamp.Format("02/01/2006 15:04:05"))
	fmt.Println("\nDados preenchidos:")
	for fieldID, value := range result.Data {
		fmt.Printf("%s: %v\n", fieldID, value)
	}
}

// ExampleFormFillerAdvanced demonstra recursos avançados do preenchedor
func ExampleFormFillerAdvanced() {
	filler, err := NewFormFiller()
	if err != nil {
		fmt.Printf("Erro ao criar preenchedor: %v\n", err)
		return
	}

	// Criar formulário com validações customizadas
	form := Form{
		ID:          "pedido",
		Name:        "Pedido de Compra",
		Description: "Formulário para realização de pedidos",
		Fields: []FormField{
			{
				ID:          "produto",
				Name:        "produto",
				Label:       "Produto",
				Type:        FieldTypeSelect,
				Required:    true,
				Options: []FieldOption{
					{Value: "1", Label: "Produto A - R$ 100,00"},
					{Value: "2", Label: "Produto B - R$ 200,00"},
					{Value: "3", Label: "Produto C - R$ 300,00"},
				},
			},
			{
				ID:          "quantidade",
				Name:        "quantidade",
				Label:       "Quantidade",
				Type:        FieldTypeNumber,
				Required:    true,
				Validations: []ValidationRule{
					{
						Type:    "min",
						Value:   1,
						Message: "Quantidade mínima é 1",
					},
					{
						Type:    "max",
						Value:   10,
						Message: "Quantidade máxima é 10",
					},
				},
			},
			{
				ID:          "cupom",
				Name:        "cupom",
				Label:       "Cupom de Desconto",
				Type:        FieldTypeText,
				Required:    false,
				Validations: []ValidationRule{
					{
						Type:    "regex",
						Value:   "^[A-Z0-9]{6}$",
						Message: "Cupom deve ter 6 caracteres alfanuméricos",
					},
				},
			},
			{
				ID:          "endereco_entrega",
				Name:        "endereco_entrega",
				Label:       "Endereço de Entrega",
				Type:        FieldTypeTextarea,
				Required:    true,
				Validations: []ValidationRule{
					{
						Type:    "min",
						Value:   10,
						Message: "Endereço muito curto",
					},
				},
			},
		},
	}

	if err := filler.CreateForm(form); err != nil {
		fmt.Printf("Erro ao criar formulário: %v\n", err)
		return
	}

	// Testar diferentes cenários de preenchimento
	testCases := []struct {
		name    string
		data    map[string]interface{}
		options FillOptions
	}{
		{
			name: "Pedido válido",
			data: map[string]interface{}{
				"produto":          "1",
				"quantidade":       5,
				"cupom":           "ABC123",
				"endereco_entrega": "Rua Exemplo, 123 - Bairro - Cidade - Estado - CEP 12345-678",
			},
			options: FillOptions{
				ValidateOnFill: true,
			},
		},
		{
			name: "Pedido com erros",
			data: map[string]interface{}{
				"produto":          "1",
				"quantidade":       15, // Maior que o máximo permitido
				"cupom":           "123", // Formato inválido
				"endereco_entrega": "Rua X", // Muito curto
			},
			options: FillOptions{
				ValidateOnFill: true,
			},
		},
		{
			name: "Pedido com auto-complete",
			data: map[string]interface{}{
				"produto": "2",
			},
			options: FillOptions{
				ValidateOnFill: true,
				AutoComplete:   true,
			},
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\nTestando: %s\n", tc.name)
		fmt.Println("----------------------------------------")

		result, err := filler.FillForm(form.ID, tc.data, tc.options)
		if err != nil {
			fmt.Printf("Erro ao preencher formulário: %v\n", err)
			continue
		}

		if len(result.Errors) > 0 {
			fmt.Println("Erros de validação:")
			for _, err := range result.Errors {
				fmt.Printf("- Campo %s: %s\n", err.FieldID, err.Message)
			}
		} else {
			fmt.Println("Dados preenchidos:")
			for fieldID, value := range result.Data {
				fmt.Printf("- %s: %v\n", fieldID, value)
			}
		}
	}

	// Exportar dados em diferentes formatos
	formData, _ := filler.GetFormData(form.ID)
	if formData != nil {
		fmt.Println("\nExportando dados:")
		
		jsonData, _ := filler.ExportFormData(form.ID, "json")
		fmt.Printf("\nJSON:\n%s\n", string(jsonData))

		csvData, _ := filler.ExportFormData(form.ID, "csv")
		if csvData != nil {
			fmt.Printf("\nCSV:\n%s\n", string(csvData))
		}
	}

	// Imprimir estatísticas
	stats, _ := filler.GetStatistics()
	fmt.Println("\nEstatísticas:")
	for key, value := range stats {
		fmt.Printf("%s: %v\n", key, value)
	}
} 