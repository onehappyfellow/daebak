package sources

import (
	"fmt"
	"log"

	"github.com/gocolly/colly"
)

// Article represents a news article with its source URL, original title, and translation
type Article struct {
	Source      string `json:"src"`
	Title       string `json:"title"`
	Translation string `json:"translation,omitempty"`
}

// ScrapeImaeil scrapes news headlines from imaeil.com
func ScrapeImaeil() ([]Article, error) {
	// Initialize the collector
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"),
	)

	var articles []Article

	// Set up the collector to find news headline links
	c.OnHTML(".wcms_bestnews_day a", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		title := e.Text
		
		// Add to articles slice
		articles = append(articles, Article{
			Source: href,
			Title:  title,
		})
	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Request URL: %s failed with response: %v\nError: %s", r.Request.URL, r, err)
	})

	// Start the scraping
	err := c.Visit("https://www.imaeil.com/")
	if err != nil {
		return nil, fmt.Errorf("failed to visit URL: %w", err)
	}

	// Wait until the collector has completed its job
	c.Wait()

	if len(articles) == 0 {
		return nil, fmt.Errorf("no articles found")
	}
	
	return articles, nil
}