package tools

import (
	"fmt"
	"log"
)

// ExampleNER demonstra o uso do SpacyNER
func ExampleNER() {
	// Criar cliente NER
	ner, err := NewSpacyNER()
	if err != nil {
		log.Fatal(err)
	}

	// Exemplo de texto para análise
	text := `A Microsoft anunciou hoje que Bill Gates doou R$ 10 milhões para projetos 
	         de inteligência artificial na Universidade de São Paulo em 15 de março de 2024. 
	         O evento aconteceu na sede da empresa em Redmond, Washington.`

	// Extrair todas as entidades
	result, err := ner.ExtractEntities(NEROptions{
		Text:         text,
		Language:     "pt",
		Normalize:    true,
		IncludeSpans: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Todas as entidades encontradas:")
	for _, entity := range result.Entities {
		fmt.Printf("- Texto: %s\n", entity.Text)
		fmt.Printf("  Tipo: %s\n", entity.Type)
		fmt.Printf("  Posição: %d-%d\n", entity.Start, entity.End)
		if entity.Normalized != "" {
			fmt.Printf("  Normalizado: %s\n", entity.Normalized)
		}
		fmt.Println()
	}

	// Buscar apenas pessoas e organizações
	result, err = ner.ExtractEntities(NEROptions{
		Text:     text,
		Language: "pt",
		Types:    []string{"PERSON", "ORG"},
		MinScore: 0.5,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nPessoas e Organizações:")
	for _, entity := range result.Entities {
		fmt.Printf("- %s (%s)\n", entity.Text, entity.Type)
	}

	// Listar tipos de entidades suportados
	fmt.Println("\nTipos de entidades suportados:")
	for _, tipo := range ner.GetSupportedEntityTypes() {
		fmt.Printf("- %s\n", tipo)
	}

	// Listar idiomas suportados
	fmt.Println("\nIdiomas suportados:")
	for _, lang := range ner.GetSupportedLanguages() {
		fmt.Printf("- %s\n", lang)
	}
}

// ExampleNERAdvanced demonstra recursos avançados do NER
func ExampleNERAdvanced() {
	ner, err := NewSpacyNER()
	if err != nil {
		log.Fatal(err)
	}

	// Texto com múltiplos tipos de entidades
	text := `O presidente da Apple, Tim Cook, participou de uma reunião com 
	         representantes do Banco do Brasil e da Petrobras no dia 20/04/2024 
	         para discutir investimentos de US$ 500 milhões em tecnologia verde. 
	         O encontro ocorreu no Palácio do Planalto, em Brasília.`

	// Extrair entidades com configurações específicas
	result, err := ner.ExtractEntities(NEROptions{
		Text:            text,
		Language:        "pt",
		Types:           []string{"PERSON", "ORG", "GPE", "MONEY", "DATE"},
		MinScore:        0.7,
		Normalize:       true,
		IncludeSpans:    true,
		IncludeSubtypes: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Agrupar entidades por tipo
	entityGroups := make(map[string][]Entity)
	for _, entity := range result.Entities {
		entityGroups[entity.Type] = append(entityGroups[entity.Type], entity)
	}

	fmt.Println("Entidades agrupadas por tipo:")
	for tipo, entities := range entityGroups {
		fmt.Printf("\n%s:\n", tipo)
		for _, entity := range entities {
			fmt.Printf("- %s", entity.Text)
			if entity.Normalized != "" {
				fmt.Printf(" (normalizado: %s)", entity.Normalized)
			}
			if entity.Score > 0 {
				fmt.Printf(" [confiança: %.2f]", entity.Score)
			}
			fmt.Println()
		}
	}
} 