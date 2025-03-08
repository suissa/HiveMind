package tools

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
)

// CollyScraper implementa a interface ScraperTool usando o Colly
type CollyScraper struct {
	collector *colly.Collector
	mu        sync.Mutex
}

// NewCollyScraper cria uma nova instância do CollyScraper
func NewCollyScraper() *CollyScraper {
	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.MaxDepth(3),
	)

	return &CollyScraper{
		collector: c,
	}
}

// Scrape implementa a interface ScraperTool
func (s *CollyScraper) Scrape(options ScraperOptions) (*ScraperResult, error) {
	result := &ScraperResult{
		Content:  "",
		Metadata: make(map[string]string),
		Links:    make([]string, 0),
		Images:   make([]string, 0),
	}

	// Configurar o collector
	s.collector.MaxDepth = options.MaxDepth
	if !options.FollowLinks {
		s.collector.MaxDepth = 1
	}

	// Mutex para sincronizar acesso aos dados do resultado
	var mu sync.Mutex

	// Handler para o conteúdo HTML
	s.collector.OnHTML("html", func(e *colly.HTMLElement) {
		mu.Lock()
		defer mu.Unlock()

		// Coletar metadados
		e.ForEach("meta", func(_ int, el *colly.HTMLElement) {
			name := el.Attr("name")
			content := el.Attr("content")
			if name != "" && content != "" {
				result.Metadata[name] = content
			}
		})

		// Coletar conteúdo baseado nos seletores
		if len(options.Selectors) > 0 {
			for _, selector := range options.Selectors {
				e.ForEach(selector, func(_ int, el *colly.HTMLElement) {
					text := strings.TrimSpace(el.Text)
					if text != "" {
						if result.Content != "" {
							result.Content += "\n\n"
						}
						result.Content += text
					}
				})
			}
		} else {
			// Se não houver seletores específicos, pegar o texto do body
			text := strings.TrimSpace(e.DOM.Find("body").Text())
			if text != "" {
				result.Content = text
			}
		}

		// Coletar links se solicitado
		if options.CollectLinks {
			e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
				link := el.Attr("href")
				if link != "" && !strings.HasPrefix(link, "#") {
					result.Links = append(result.Links, el.Request.AbsoluteURL(link))
				}
			})
		}

		// Coletar imagens se solicitado
		if options.CollectImages {
			e.ForEach("img[src]", func(_ int, el *colly.HTMLElement) {
				src := el.Attr("src")
				if src != "" {
					result.Images = append(result.Images, el.Request.AbsoluteURL(src))
				}
			})
		}
	})

	// Handler para erros
	s.collector.OnError(func(r *colly.Response, err error) {
		result.Error = fmt.Sprintf("Erro ao acessar %s: %v", r.Request.URL, err)
	})

	// Iniciar o scraping
	err := s.collector.Visit(options.URL)
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar scraping: %v", err)
	}

	// Aguardar a conclusão
	s.collector.Wait()

	return result, nil
} 