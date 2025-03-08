package tools

// ScraperResult representa o resultado de um scraping
type ScraperResult struct {
	Content     string            `json:"content,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Links       []string          `json:"links,omitempty"`
	Images      []string          `json:"images,omitempty"`
	Error       string            `json:"error,omitempty"`
}

// ScraperOptions representa as opções comuns de scraping
type ScraperOptions struct {
	URL                string   `json:"url"`
	Selectors         []string `json:"selectors,omitempty"`
	MaxDepth          int      `json:"max_depth"`
	FollowLinks       bool     `json:"follow_links"`
	WaitForSelector   string   `json:"wait_for_selector,omitempty"`
	WaitTime         int      `json:"wait_time,omitempty"`
	CollectLinks     bool     `json:"collect_links"`
	CollectImages    bool     `json:"collect_images"`
}

// ScraperTool é a interface que todas as ferramentas de scraping devem implementar
type ScraperTool interface {
	Scrape(options ScraperOptions) (*ScraperResult, error)
} 