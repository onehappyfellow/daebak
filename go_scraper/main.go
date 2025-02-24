package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"onehappyfellow.com/daebak/sources"
)

// TranslationResponse represents the JSON response from Claude
type TranslationResponse struct {
	Translations []string `json:"translations"`
}

// AnthropicRequest represents the request structure for the Anthropic API
type AnthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Temp      float64   `json:"temperature"`
	Messages  []Message `json:"messages"`
}

// Message represents a message in the Anthropic API request
type Message struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

// ContentBlock represents a content block in a message
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func main() {
	// Define flags
	sourceFlag := flag.String("source", "all", "Source to scrape (chosun, imaeil, or all)")
	translations := flag.Bool("translations", false, "Translate headlines using Claude API")
	modelFlag := flag.String("model", "haiku", "Claude model to use (haiku or opus)")
	emailFlag := flag.Bool("email", false, "Send articles via email")
	
	// Parse flags
	flag.Parse()
	
	// Determine which model to use
	var model string
	switch strings.ToLower(*modelFlag) {
	case "opus":
		model = "claude-3-opus-20240229"
	case "haiku":
		fallthrough
	default:
		model = "claude-3-5-haiku-20241022"
	}
	
	// Process source flag
	source := strings.ToLower(*sourceFlag)
	scrapeImaeil := source == "imaeil" || source == "all"
	scrapeChosun := source == "chosun" || source == "all"
	
	// If an invalid source is specified, default to all
	if !scrapeImaeil && !scrapeChosun {
		log.Printf("Invalid source specified: %s. Defaulting to 'all'", source)
		scrapeImaeil = true
		scrapeChosun = true
	}
	
	// Collect articles from different sources
	var allArticles []sources.Article
	
	// Scrape Maeil news
	if scrapeImaeil {
		log.Println("Scraping Maeil Index...")
		maeilArticles, err := sources.ScrapeImaeil()
		if err != nil {
			log.Printf("Error scraping Maeil: %v", err)
		} else {
			log.Printf("Found %d articles from Maeil Index", len(maeilArticles))
			allArticles = append(allArticles, maeilArticles...)
		}
	}
	
	// Scrape Chosun news
	if scrapeChosun {
		log.Println("Scraping Chosun Ilbo...")
		chosunArticles, err := sources.ScrapeChosun()
		if err != nil {
			log.Printf("Error scraping Chosun: %v", err)
		} else {
			log.Printf("Found %d articles from Chosun Ilbo", len(chosunArticles))
			allArticles = append(allArticles, chosunArticles...)
		}
	}
	
	if len(allArticles) == 0 {
		log.Println("No articles were collected")
		return
	}
	
	log.Printf("Collected a total of %d articles", len(allArticles))
	
	// Translate if requested
	var articlesToUse []sources.Article
	if *translations {
		log.Printf("Translating headlines using Claude model: %s", model)
		translatedArticles, err := translateArticles(allArticles, model)
		if err != nil {
			log.Printf("Translation error: %v", err)
			articlesToUse = allArticles
		} else {
			articlesToUse = translatedArticles
		}
	} else {
		articlesToUse = allArticles
	}

	// Print the articles
	prettyPrint(articlesToUse)

	// Send email if requested
	if *emailFlag {
		log.Println("Sending articles via email...")
		emailBody := formatArticlesForEmail(articlesToUse)
		emailSubject := fmt.Sprintf("Korean News Update - %s", time.Now().Format("Jan 2, 2006"))
		err := sendEmail(emailSubject, emailBody)
		if err != nil {
			log.Printf("Error sending email: %v", err)
		} else {
			log.Println("Email sent successfully")
		}
	}
}

// translateArticles translates the titles of all articles using Claude API
func translateArticles(articles []sources.Article, model string) ([]sources.Article, error) {
	if len(articles) == 0 {
		return articles, nil
	}
	
	// Build prompt for translation
	prompt := "Translate these news headlines into English:\n"
	for i, article := range articles {
		prompt += fmt.Sprintf("%d. %s\n", i+1, article.Title)
	}
	
	// Add final instruction for Claude
	prompt += `Respond only in JSON with this format: {"translations": [ordered list of translations]}`
	
	// Get translations from Claude
	translations, err := callClaudeAPI(prompt, model)
	if err != nil {
		return articles, err
	}
	
	// Create a new slice for translated articles
	translatedArticles := make([]sources.Article, len(articles))
	copy(translatedArticles, articles)
	
	// Add translations to articles
	for i, translation := range translations {
		if i < len(translatedArticles) {
			translatedArticles[i].Translation = translation
		}
	}
	
	return translatedArticles, nil
}

// callClaudeAPI sends a request to the Claude API and parses the response
func callClaudeAPI(prompt string, model string) ([]string, error) {
	// Get API key from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}
	
	// Create the Anthropic API request
	apiRequest := AnthropicRequest{
		Model:     model,
		MaxTokens: 5000,
		Temp:      1.0,
		Messages: []Message{
			{
				Role: "user",
				Content: []ContentBlock{
					{
						Type: "text",
						Text: prompt,
					},
				},
			},
		},
	}

	// Marshal the request to JSON
	requestBody, err := json.Marshal(apiRequest)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	// Parse the response
	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("empty response content")
	}

	// Extract translations from the JSON response
	responseText := response.Content[0].Text
	// Clean the response text if needed
	responseText = strings.TrimSpace(responseText)

	var translationResponse TranslationResponse
	if err := json.Unmarshal([]byte(responseText), &translationResponse); err != nil {
		return nil, fmt.Errorf("error parsing translations JSON: %w", err)
	}

	return translationResponse.Translations, nil
}

// prettyPrint formats and prints the articles
func prettyPrint(articles []sources.Article) {
	fmt.Println("\nArticles:")
	for i, article := range articles {
		fmt.Printf("%d. Source: %s\n", i+1, article.Source)
		fmt.Printf("   Title: %s\n", article.Title)
		if article.Translation != "" {
			fmt.Printf("   Translation: %s\n", article.Translation)
		}
		fmt.Println()
	}
}

func formatArticlesForEmail(articles []sources.Article) string {
	var buffer bytes.Buffer
	
	buffer.WriteString("Today's Korean News Headlines\n")
	buffer.WriteString("==========================\n\n")
	
	for i, article := range articles {
		buffer.WriteString(fmt.Sprintf("%d. %s\n", i+1, article.Title))
		if article.Translation != "" {
			buffer.WriteString(fmt.Sprintf("   %s\n", article.Translation))
		}
		buffer.WriteString(fmt.Sprintf("   Link: %s\n\n", article.Source))
	}
	
	buffer.WriteString("\n--\n")
	buffer.WriteString("This email was automatically generated by the Korean News Scraper.\n")
	buffer.WriteString(fmt.Sprintf("Generated on: %s\n", time.Now().Format(time.RFC1123Z)))
	
	return buffer.String()
}

// sendEmail sends an email with the given subject and body
func sendEmail(subject, body string) error {
	// SMTP server configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	
	// Sender details
	from := "jonathan.droege@gmail.com"
	
	// Get password from environment
	password := os.Getenv("GMAIL_APP_PASSWORD")
	if password == "" {
		return fmt.Errorf("missing required environment variable GMAIL_APP_PASSWORD")
	}
	
	// Recipient(s)
	to := []string{
		"jonathan.droege+donotreply@gmail.com",
	}
	
	// Format recipient list for the header
	toHeader := strings.Join(to, ", ")
	
	// Build the email header
	header := make(map[string]string)
	header["From"] = fmt.Sprintf("Korean News Scraper <%s>", from)
	header["To"] = toHeader
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=UTF-8"
	header["Content-Transfer-Encoding"] = "8bit"
	header["Date"] = time.Now().Format(time.RFC1123Z)
	
	// Build the message
	var message bytes.Buffer
	for key, value := range header {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	message.WriteString("\r\n")
	message.WriteString(body)
	
	// Authenticate
	auth := smtp.PlainAuth("", from, password, smtpHost)
	
	// Send the email
	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", smtpHost, smtpPort),
		auth,
		from,
		to,
		message.Bytes(),
	)
	
	return err
}