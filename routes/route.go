package routes

import (
	"fmt"
	"strings"

	"tele-uni/handller"
	lib "tele-uni/lib"
	"tele-uni/model"

	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

var activeSessions = make(map[int64]bool) // To track users who are searching for mahasiswa
var activecommand = make(map[int64]string)
var userResponses = make(map[int64]string) // Track user's response (true/false)

func containsKeyword(message string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(message, keyword) {
			return true
		}
	}
	return false
}

func Handler(bot *telebot.Bot, db *gorm.DB) *telebot.Bot {

	// Define a handler for incoming messages
	bot.Handle(telebot.OnText, func(c telebot.Context) error {

		if activeSessions[c.Sender().ID] {
			userMessage := strings.ToLower(c.Text()) // Convert text to lowercase for easier matching

			if activecommand[c.Sender().ID] == "kurikulum" {
				// Get the name entered by the user
				semester := strings.TrimSpace(c.Text())

				// Scrape the curriculum data for the given semester
				data, err := lib.ScrapeKurikulumBySemester(semester)
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

			if activecommand[c.Sender().ID] == "pakar" {

				processRuleBasedQuestions(c, userMessage, db)

				if userResponses[c.Sender().ID] == "" {
					return handller.HandleTrueFalseResponse(c, db, activeSessions, activecommand)
				}

			}

			var rule model.Rules
			err := db.Where("LOWER(condition) LIKE ?", "%"+userMessage+"%").First(&rule).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return c.Send("Maaf, saya tidak mengerti pesan Anda. Bisa Anda coba lebih spesifik?")
				}
				return c.Send("Terjadi kesalahan saat mencari aturan. Silakan coba lagi.")
			}

			// Return the response from the "Then" field of the rule
			return c.Send(rule.Then)
		}
		// Echo the received message
		userMessage := strings.ToLower(c.Text()) // Convert text to lowercase for easier matching

		var message model.Messages

		jokeKeywords := []string{"joke", "lelucon", "lawakan"}

		// Check if the message contains any joke-related keywords
		if containsKeyword(userMessage, jokeKeywords) {
			// Fetch and return a joke
			joke, err := lib.GetJoke() // Assume GetJoke() is a function that fetches a random joke
			if err != nil {
				return c.Send("Failed to fetch a joke. Please try again!")
			}
			return c.Send(joke)
		}

		quoteKeywords := []string{"quote", "kutipan", "inspirasi"}
		// Check if the message contains any quote-related keywords
		if containsKeyword(userMessage, quoteKeywords) {
			// Fetch and return a quote
			quote, err := lib.GetQuote() // Assume GetQuote() is a function that fetches a random quote
			if err != nil {
				return c.Send("Failed to fetch a quote. Please try again!")
			}
			return c.Send(quote)
		}
		// Check for specific keywords or phrases in the message
		userWords := strings.Fields(userMessage)

		minMatches := 2

		var rule model.Rules
		query := "SELECT * FROM rules WHERE LOWER(condition) ILIKE ?"

		for _, word := range userWords {
			// Execute the raw SQL query to find rules based on the user's input
			err := db.Raw(query, "%"+word+"%").Scan(&rule).Error
			if err == nil {
				matchCount := 0
				// Check how many words in the condition match the user's words
				for _, conditionWord := range strings.Fields(strings.ToLower(rule.Condition)) {
					for _, userWord := range userWords {
						if conditionWord == userWord {
							matchCount++
						}
					}
				}

				// If the match count reaches the threshold (minMatches), respond with the "Then" field of the rule
				if matchCount >= minMatches {
					return c.Send(rule.Then)
				}
			}
		}

		// If no rule matches, check the Messages table for a direct response
		err := db.Where("LOWER(msg) = ?", userMessage).First(&message).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Send("Maaf, saya tidak mengerti pesan Anda. Bisa Anda coba lebih spesifik?")
			}
			// Handle database errors
			return c.Send("Terjadi kesalahan saat mencari respons. Silakan coba lagi.")
		}

		// Send the response from the database message
		return c.Send(message.Response)
	})

	bot.Handle("/kurikulum", func(c telebot.Context) error {
		activeSessions[c.Sender().ID] = true // Mark user as in "searching" mode
		activecommand[c.Sender().ID] = "kurikulum"
		return c.Send("kirim semester untuk dilakukan pengecekan, (contoh: '1')")
	})

	bot.Handle("/start", func(c telebot.Context) error {
		return c.Send("welcome to the fantasy world!")
	})

	bot.Handle("/joke", func(c telebot.Context) error {
		joke, err := lib.GetJoke()
		if err != nil {
			return c.Send("Failed to fetch a joke. Please try again!")
		}
		return c.Send(joke)
	})

	bot.Handle("/quote", func(c telebot.Context) error {
		quote, err := lib.GetQuote()
		if err != nil {
			return c.Send("Failed to fetch a quote. Please try again!")
		}
		return c.Send(quote)
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
		return c.Send("You selected Search Rules. (This simulates the /pakar command)")
	})

	bot.Handle(&telebot.InlineButton{Unique: "btn1"}, func(c telebot.Context) error {
		return c.Send("You clicked Button 1")
	})

	bot.Handle(telebot.OnPhoto, func(c telebot.Context) error {
		return c.Send("Nice photo!")
	})
	return bot
}

func processRuleBasedQuestions(c telebot.Context, userMessage string, db *gorm.DB) error {
	// Fetch all the rules
	var rules []model.Rules
	err := db.Find(&rules).Error
	if err != nil {
		return c.Send("Failed to fetch rules. Please try again!")
	}

	// Look for matching rules based on user input
	for _, rule := range rules {
		if strings.Contains(userMessage, rule.Condition) {
			// Ask the user if the condition is true or false
			activeSessions[c.Sender().ID] = true
			userResponses[c.Sender().ID] = "" // Reset previous response
			c.Send(fmt.Sprintf("Apakah pernyataan ini benar? %s", rule.Condition))

			// Wait for the user's response ("true" or "false")
			// We'll handle the true/false response in the OnText handler below
			return nil
		}
	}

	return c.Send("Maaf, saya tidak mengerti kondisi tersebut.")
}
