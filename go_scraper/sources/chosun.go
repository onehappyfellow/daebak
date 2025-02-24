package sources

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

const YYYYMD = "2006/1/2"

// ScrapeChosun scrapes economy news headlines from chosun.com
func ScrapeChosun() ([]Article, error) {
	// Initialize the collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.chosun.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"),
	)

	articles := make(map[string]Article)

	// Set up the collector to find news headline links
	c.OnHTML("a.story-card__headline", func(e *colly.HTMLElement) {
		url := e.Attr("href")
		headline := e.Text
		
		if validateUrl(url) {
			if _, exists := articles[url]; !exists {
				articles[url] = Article{
					Source: fmt.Sprintf("https://www.chosun.com%s", url),
					Title:  headline,
				}
			}
		}
	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Request URL: %s failed with response: %v\nError: %s", r.Request.URL, r, err)
	})

	// Start the scraping
	err := c.Visit("https://www.chosun.com/economy/")
	if err != nil {
		return nil, fmt.Errorf("failed to visit URL: %w", err)
	}

	// Wait until the collector has completed its job
	c.Wait()

	// Convert map to slice
	var articlesList []Article
	for _, article := range articles {
		articlesList = append(articlesList, article)
	}

	if len(articlesList) == 0 {
		return nil, fmt.Errorf("no articles found")
	}
	
	log.Printf("Scraped Chosun Ilbo - %d articles found", len(articlesList))
	return articlesList, nil
}

// validateUrl checks if the URL contains today's date (in Seoul timezone)
// and doesn't contain advertising link attribution
func validateUrl(url string) bool {
	// function will return true for urls that contain the current date
	// assuming seoul timezone (must run before 9:00am MST)
	r := regexp.MustCompile(`\d{4}/\d{1,2}/\d{1,2}`)
	
	// screen out links that contain advertising link attribution
	if strings.Contains(url, "utm_medium") {
		return false
	}
	
	// validate current date
	if r.MatchString(url) {
		dateStr := r.FindString(url)
		seoul, err := time.LoadLocation("Asia/Seoul")
		if err != nil {
			log.Printf("Error loading Seoul timezone: %v", err)
			return false
		}
		
		// if set to 24 hours, must run before 8am MST
		hours := 48 * time.Hour
		now := time.Now().In(seoul)
		pub, err := time.ParseInLocation(YYYYMD, dateStr, seoul)
		if err != nil {
			log.Printf("Cannot validate as date YYYY/M/D, %s: %s", dateStr, err)
			return false
		}
		return now.Sub(pub) < hours
	}
	
	// url does not have date, default to invalid
	return false
}