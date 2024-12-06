package model

type Messages struct {
	ID       int    `gorm:"primaryKey;autoIncrement"` // Auto-incrementing primary key
	Msg      string `gorm:"type:text;not null"`       // Text field for the message
	Response string `gorm:"type:text;not null"`       // Text field for the response
}

type JokeResponse struct {
	Setup    string `json:"setup"`
	Punchline string `json:"punchline"`
}

// QuoteResponse represents the quote API response
type QuoteResponse struct {
	Q string `json:"q"` // Quote
	A string `json:"a"` // Author
}