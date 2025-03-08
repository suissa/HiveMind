package tools

import (
	"fmt"
	"strings"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

// SeleniumScraper implementa a interface ScraperTool usando o Selenium
type SeleniumScraper struct {
	service *selenium.Service
	driver  selenium.WebDriver
}

// NewSeleniumScraper cria uma nova instância do SeleniumScraper
func NewSeleniumScraper() (*SeleniumScraper, error) {
	// Configurar o serviço do ChromeDriver
	service, err := selenium.NewChromeDriverService("chromedriver", 4444)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar serviço ChromeDriver: %v", err)
	}

	// Configurar as opções do Chrome
	caps := selenium.Capabilities{
		"browserName": "chrome",
	}

	// Configurar opções específicas do Chrome
	chromeCaps := chrome.Capabilities{
		Args: []string{
			"--headless",              // Modo headless
			"--no-sandbox",            // Necessário para alguns ambientes
			"--disable-dev-shm-usage", // Necessário para alguns ambientes
		},
	}
	caps.AddChrome(chromeCaps)

	// Criar o WebDriver
	driver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 4444))
	if err != nil {
		service.Stop()
		return nil, fmt.Errorf("erro ao criar WebDriver: %v", err)
	}

	return &SeleniumScraper{
		service: service,
		driver:  driver,
	}, nil
}

// Close fecha o serviço e o driver do Selenium
func (s *SeleniumScraper) Close() error {
	if s.driver != nil {
		s.driver.Quit()
	}
	if s.service != nil {
		return s.service.Stop()
	}
	return nil
}

// Scrape implementa a interface ScraperTool
func (s *SeleniumScraper) Scrape(options ScraperOptions) (*ScraperResult, error) {
	result := &ScraperResult{
		Content:  "",
		Metadata: make(map[string]string),
		Links:    make([]string, 0),
		Images:   make([]string, 0),
	}

	// Navegar até a URL
	err := s.driver.Get(options.URL)
	if err != nil {
		return nil, fmt.Errorf("erro ao acessar URL: %v", err)
	}

	// Aguardar pelo seletor específico se fornecido
	if options.WaitForSelector != "" {
		_, err := s.driver.FindElement(selenium.ByCSSSelector, options.WaitForSelector)
		if err != nil {
			return nil, fmt.Errorf("erro ao aguardar pelo seletor %s: %v", options.WaitForSelector, err)
		}
	}

	// Aguardar tempo adicional se especificado
	if options.WaitTime > 0 {
		time.Sleep(time.Duration(options.WaitTime) * time.Second)
	}

	// Coletar conteúdo baseado nos seletores
	if len(options.Selectors) > 0 {
		for _, selector := range options.Selectors {
			elements, err := s.driver.FindElements(selenium.ByCSSSelector, selector)
			if err != nil {
				continue
			}

			for _, element := range elements {
				text, err := element.Text()
				if err == nil && text != "" {
					if result.Content != "" {
						result.Content += "\n\n"
					}
					result.Content += strings.TrimSpace(text)
				}
			}
		}
	} else {
		// Se não houver seletores específicos, pegar o texto do body
		body, err := s.driver.FindElement(selenium.ByTagName, "body")
		if err == nil {
			text, err := body.Text()
			if err == nil {
				result.Content = strings.TrimSpace(text)
			}
		}
	}

	// Coletar metadados
	metaTags, err := s.driver.FindElements(selenium.ByTagName, "meta")
	if err == nil {
		for _, meta := range metaTags {
			name, _ := meta.GetAttribute("name")
			content, _ := meta.GetAttribute("content")
			if name != "" && content != "" {
				result.Metadata[name] = content
			}
		}
	}

	// Coletar links se solicitado
	if options.CollectLinks {
		links, err := s.driver.FindElements(selenium.ByTagName, "a")
		if err == nil {
			for _, link := range links {
				href, err := link.GetAttribute("href")
				if err == nil && href != "" && !strings.HasPrefix(href, "#") {
					result.Links = append(result.Links, href)
				}
			}
		}
	}

	// Coletar imagens se solicitado
	if options.CollectImages {
		images, err := s.driver.FindElements(selenium.ByTagName, "img")
		if err == nil {
			for _, img := range images {
				src, err := img.GetAttribute("src")
				if err == nil && src != "" {
					result.Images = append(result.Images, src)
				}
			}
		}
	}

	return result, nil
} 