package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

const YYYYMD = "2006/1/2"

type Article struct {
	Url, Headline string
}

func main() {
	fmt.Println("Scraping Chosun Ilbo")

	c := colly.NewCollector(
		colly.AllowedDomains("www.chosun.com"),
	)

	articles := make(map[string]Article)

	c.OnHTML("a.story-card__headline", func(e *colly.HTMLElement) {
		a := Article{}
		a.Url = e.Attr("href")
		a.Headline = e.Text
		if validateUrl(a.Url) {
			if _, exists := articles[a.Url]; !exists {
				articles[a.Url] = a
				r := regexp.MustCompile(`\d{4}/\d{1,2}/\d{1,2}`)
				dateStr := r.FindString(a.Url)
				fmt.Println(dateStr, a.Headline)
			}
		}
	})

	c.Visit("https://www.chosun.com/economy/")

	fmt.Printf("Page lists %d current articles", len(articles))

}

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
		seoul, _ := time.LoadLocation("Asia/Seoul")
		hours := 24 * time.Hour

		now := time.Now().In(seoul)
		pub, err := time.ParseInLocation(YYYYMD, dateStr, seoul)
		if err != nil {
			fmt.Printf("Cannot validate as date YYYY/M/D, %s: %s\n", dateStr, err)
			return false
		}

		return now.Sub(pub) < hours
	}

	// url does not have date, default to invalid
	return false
}
