package lib

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"tele-uni/model"

	"github.com/gocolly/colly"
)

func GetJoke() (string, error) {
	resp, err := http.Get("https://official-joke-api.appspot.com/random_joke")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var joke model.JokeResponse
	if err := json.NewDecoder(resp.Body).Decode(&joke); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s\n\n%s", joke.Setup, joke.Punchline), nil
}

func GetQuote() (string, error) {
	resp, err := http.Get("https://zenquotes.io/api/random")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var quote []model.QuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quote); err != nil {
		return "", err
	}

	return fmt.Sprintf("\"%s\" - %s", quote[0].Q, quote[0].A), nil
}

func ScrapeKurikulumBySemester(semester string) ([][]string, error) {
	url := "https://if.unila.ac.id/kurikulum-2020-program-studi-s1-teknik-informatika-universitas-lampung/"
	log.Println("Starting to scrape the curriculum page:", url)
	// Create a new Colly collector
	c := colly.NewCollector(
		colly.AllowedDomains("if.unila.ac.id"),
		colly.AllowURLRevisit(),
	)

	// Log request information
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting:", r.URL.String())
	})

	// Log responses
	c.OnResponse(func(r *colly.Response) {
		log.Printf("Response received with status %d from %s\n", r.StatusCode, r.Request.URL)
	})

	// Log errors
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error: %v for URL: %s\n", err, r.Request.URL)
	})

	// Store scraped data
	var tableData [][]string
	var semesterFound bool

	// Extract relevant tables
	c.OnHTML("table", func(e *colly.HTMLElement) {
		log.Println("Found a table element on the page")

		e.ForEach("tbody > tr", func(i int, row *colly.HTMLElement) {
			rowText := strings.TrimSpace(row.Text)
			log.Printf("Processing row %d: %s\n", i, rowText)

			// Check for semester header
			if strings.Contains(strings.ToLower(rowText), "semester "+strings.ToLower(semester)) {
				log.Printf("Found header for %s\n", semester)
				semesterFound = true
				return
			}

			// If semesterFound, process rows
			if semesterFound {
				// Break when the next semester header is reached
				if strings.Contains(strings.ToLower(rowText), "semester ") && !strings.Contains(strings.ToLower(rowText), "semester "+strings.ToLower(semester)) {
					log.Printf("Next semester detected, stopping collection for %s\n", semester)
					semesterFound = false
					return
				}

				// Extract and log table data
				var rowData []string
				row.ForEach("td", func(_ int, cell *colly.HTMLElement) {
					cellText := strings.TrimSpace(cell.Text)
					log.Printf("Extracted cell data: %s\n", cellText)
					rowData = append(rowData, cellText)
				})

				// Append valid rows
				if len(rowData) > 0 {
					log.Printf("Appending row data: %v\n", rowData)
					tableData = append(tableData, rowData)
				}
			}
		})
	})

	// Visit the target URL
	err := c.Visit(url)
	if err != nil {
		log.Printf("Failed to visit URL: %v\n", err)
		return nil, err
	}

	if len(tableData) == 0 {
		log.Printf("No data found for semester: %s\n", semester)
		return nil, fmt.Errorf("semester %s not found", semester)
	}

	log.Printf("Scraping completed successfully for %s. Extracted %d rows.\n", semester, len(tableData))
	return tableData, nil
}
