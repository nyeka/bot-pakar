package handller

import (
	"strings"
	"tele-uni/model"

	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

func HandleTrueFalseResponse(c telebot.Context, db *gorm.DB, activeSessions map[int64]bool, userResponses map[int64]string) error {
	// Check if the user is in a session waiting for true/false response
	if activeSessions[c.Sender().ID] && userResponses[c.Sender().ID] == "" {
		// Get the user's response (either "true" or "false")
		userResponse := strings.ToLower(c.Text())

		if userResponse == "true" || userResponse == "false" {
			userResponses[c.Sender().ID] = userResponse

			var rule model.Rules
			err := db.Where("condition LIKE ?", "%"+userResponses[c.Sender().ID]+"%").First(&rule).Error
			if err != nil {
				return c.Send("Something went wrong with processing the rule.")
			}

			if userResponses[c.Sender().ID] == "true" {
				// Process the "true" path (using IDIfRight)
				return c.Send(rule.IDIfRight)
			} else {
				// Process the "false" path (using IDIfNotRight)
				return c.Send(rule.IDIfNotRight)
			}
		} else {
			return c.Send("Please reply with 'true' or 'false'.")
		}
	}

	return nil
}
