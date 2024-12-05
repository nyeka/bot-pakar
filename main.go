package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/gocolly/colly"
	"github.com/ledongthuc/pdf"
	"github.com/sahilm/fuzzy"
	"gopkg.in/telebot.v3"
)

var activeSessions = make(map[int64]bool) // To track users who are searching for mahasiswa
var activecommand = make(map[int64]string)

var (
	SheetID = "1Y6RMu5Ceb2ZCSpmthYHlv-g7r4DCrUSdRP0PrK32k_Y" // Replace with your Sheet ID
	// Replace with your Google Sheets ID
	SheetRange            = "Untitled spreadsheet!A2:E52" // Adjust range as needed
	ConditionResponseData [][]string                      // Global cache for condition-response data
)

var conditions = []struct {
	Condition []string
	Action    string
}{
	{[]string{"Semester = 1"}, "Tetapkan KRS dalam paket mata kuliah dengan total 20 SKS"},
	{[]string{"Semester = 2"}, "Tetapkan KRS dalam paket mata kuliah dengan total 20 SKS"},
	{[]string{"Semester ≥ 3 DAN IPK ≥ 3.0"}, "Izinkan mengambil maksimum 24 SKS"},
	{[]string{"Semester ≥ 3 DAN IPK antara 2.5 dan 2.99"}, "Batasi SKS menjadi 21"},
	{[]string{"Berencana mengambil >24 SKS"}, "Memerlukan persetujuan dari dosen pembimbing akademik"},
	{[]string{"Mengambil mata kuliah pada semester pendek"}, "Batasi SKS maksimum 9 dengan persetujuan fakultas"},
	{[]string{"Semester 7 atau lebih tinggi DAN semua mata kuliah inti telah lulus"}, "Izinkan pemilihan mata kuliah pilihan"},
	{[]string{"Mengambil PKM dengan proposal yang lolos tahap 1"}, "Konversi 1-2 SKS ke mata kuliah sesuai program studi"},
	{[]string{"Mengikuti Kuliah Kerja Nyata (KKN)"}, "Batasi SKS untuk semester tersebut hingga maksimal 20 SKS"},
	{[]string{"Semester antara (IPK program sosial ≥ 3.5 atau eksakta ≥ 3.2)"}, "Boleh ambil maksimal 9 SKS untuk semester tersebut"},
	{[]string{"Mendapat tugas akhir tetapi memiliki 2 mata kuliah wajib belum lulus"}, "Izinkan kuliah khusus atau studi terbimbing dengan persetujuan akademik"},
	{[]string{"Mengulang mata kuliah dengan nilai D atau E"}, "Perbarui nilai akhir dengan hasil pengulangan"},
}

// Function to find the matching action
func findAction(userMessage string) string {
	// Convert the user message to lowercase for case-insensitive matching
	lowerUserMessage := strings.ToLower(userMessage)

	// Iterate over conditions
	for _, condition := range conditions {
		matches := true
		// Check if all keywords in the condition are present in the user message
		for _, keyword := range condition.Condition {
			fmt.Println("Is keyword present:", strings.Contains(lowerUserMessage, strings.ToLower(keyword)))
			if !strings.Contains(lowerUserMessage, strings.ToLower(keyword)) {
				matches = false
				break
			}
		}

		// If all keywords match, return the action
		if matches {
			return condition.Action
		}
	}

	// Default response if no condition matches
	return "Maaf, saya tidak menemukan aturan yang sesuai dengan pernyataan Anda."
}

func main() {
	// Replace this with your bot token from BotFather
	botToken := "7561704710:AAGGHbhrPZuf0h5HDlaDvlDTtcZ-uAvQiXo"

	// Create a bot instance
	bot, err := telebot.NewBot(telebot.Settings{
		Token: botToken,
	})
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	bot.Handle("/kurikulum", func(c telebot.Context) error {
		activeSessions[c.Sender().ID] = true // Mark user as in "searching" mode
		activecommand[c.Sender().ID] = "kurikulum"
		return c.Send("kirim semester untuk dilakukan pengecekan, (contoh: '1')")
	})

	// Define a handler for incoming messages
	bot.Handle(telebot.OnText, func(c telebot.Context) error {

		if activeSessions[c.Sender().ID] {
			userMessage := strings.ToLower(c.Text()) // Convert text to lowercase for easier matching

			if activecommand[c.Sender().ID] == "kurikulum" {
				// Get the name entered by the user
				semester := strings.TrimSpace(c.Text())

				// Scrape the curriculum data for the given semester
				data, err := scrapeKurikulumBySemester(semester)
				if err != nil {
					delete(activeSessions, c.Sender().ID) // End the session on error
					return c.Send(fmt.Sprintf("Gagal mendapatkan data untuk %s: %v", semester, err))
				}

				// Format the data into a readable table
				var response string
				for _, row := range data {
					response += strings.Join(row, " | ") + "\n"
				}

				delete(activeSessions, c.Sender().ID) // End the session after sending the data
				return c.Send(fmt.Sprintf("Berikut adalah data kurikulum untuk semester %s:\n\n%s", semester, response))
			}

			response := findAction(userMessage) // Find the appropriate response
			return c.Send(response)

		}
		// Echo the received message
		userMessage := strings.ToLower(c.Text()) // Convert text to lowercase for easier matching

		// Check for specific keywords or phrases in the message
		switch {
		case strings.Contains(userMessage, "pembimbing 1"):
			return c.Send("Kamu bisa membuka laman Unila untuk informasi tentang pembimbing 1.")
		case strings.Contains(userMessage, "nyoman ganteng"):
			return c.Send("Betul sekali.")
		case strings.Contains(userMessage, "halo"):
			return c.Send("Hai.")
		case strings.Contains(userMessage, "cara mendapatkan pembimbing"):
			return c.Send("Untuk mendapatkan pembimbing, hubungi akademik atau cek sistem online kampus.")
		case strings.Contains(userMessage, "skripsi"):
			return c.Send("Skripsi adalah langkah terakhir dalam studi Anda. Apakah Anda butuh bantuan lain?")
		default:
			return c.Send("Maaf, saya tidak mengerti pesan Anda. Bisa Anda coba lebih spesifik?")
		}
	})

	bot.Handle("/start", func(c telebot.Context) error {
		return c.Send("welcome to the fantasy world!")
	})

	btnKurikulum := telebot.InlineButton{Unique: "kurikulum", Text: "📘 Kurikulum"}
	btnSearchRules := telebot.InlineButton{Unique: "search_rules", Text: "🔍 Search Rules"}

	bot.Handle("/menu", func(c telebot.Context) error {
		buttons := [][]telebot.InlineButton{
			{btnKurikulum},
			{btnSearchRules},
		}

		// Create the inline keyboard
		markup := &telebot.ReplyMarkup{InlineKeyboard: buttons}

		// Send the menu message with the inline keyboard
		return c.Send("Choose an option:", markup)
	})

	bot.Handle(&btnKurikulum, func(c telebot.Context) error {
		// Simulate sending the /kurikulum command
		activeSessions[c.Sender().ID] = true // Mark user as in "searching" mode
		activecommand[c.Sender().ID] = "kurikulum"
		return c.Send("kirim semester untuk dilakukan pengecekan, (contoh: '1')")
	})

	bot.Handle("/pakar", func(c telebot.Context) error {
		// Simulate sending the /kurikulum command
		activeSessions[c.Sender().ID] = true // Mark user as in "searching" mode
		activecommand[c.Sender().ID] = "pakar"
		return c.Send("kirim perintah untuk dilakukan pengecekan, (contoh: 'syarat pendaftaran skripsi')")
	})

	// Handle callback for Search Rules button
	bot.Handle(&btnSearchRules, func(c telebot.Context) error {
		// Simulate sending the /search_rules command
		return c.Send("You selected Search Rules. (This simulates the /search_rules command)")
	})

	bot.Handle("/search-rules", func(c telebot.Context) error {
		// PDF path
		pdfPath := "Peraturan.pdf"

		// Extract the PDF content
		content, err := extractTextFromPDF(pdfPath)
		if err != nil {
			log.Println("Error reading PDF:", err)
			return c.Send("Gagal membaca file PDF.")
		}

		// Get the user's query
		query := c.Message().Payload
		if query == "" {
			return c.Send("Masukkan pertanyaan Anda, contoh: 'bagaimana kita tahu bahwa mahasiswa naik tingkat'.")
		}

		// Perform fuzzy search on the PDF content
		result := fuzzySearchInPDF(content, query)

		// Send the search result back to the user
		return c.Send(result)
	})

	bot.Handle(&telebot.InlineButton{Unique: "btn1"}, func(c telebot.Context) error {
		return c.Send("You clicked Button 1")
	})

	bot.Handle(telebot.OnPhoto, func(c telebot.Context) error {
		return c.Send("Nice photo!")
	})

	log.Println("Bot is running...")
	bot.Start()
}

func extractTextFromPDF(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		log.Println("Error opening PDF:", err)
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		log.Println("Error extracting plain text from PDF:", err)
		return "", err
	}

	buf.ReadFrom(b)
	return buf.String(), nil
}

func fuzzySearchInPDF(content string, query string) string {
	// Split the content into sections for better context matching
	sections := strings.Split(content, "\n\n") // Assume sections are separated by double newlines.

	// Use fuzzy matching to find the closest section
	matches := fuzzy.Find(strings.ToLower(query), sections)

	if len(matches) == 0 {
		return "Maaf, tidak ditemukan informasi yang relevan untuk pertanyaan Anda."
	}

	// Return the most relevant section
	bestMatch := matches[0].Str
	return strings.TrimSpace(bestMatch)
}

func scrapeKurikulumBySemester(semester string) ([][]string, error) {
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
