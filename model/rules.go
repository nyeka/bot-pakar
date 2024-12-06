package model

type Rules struct {
	ID           int    `gorm:"primaryKey;autoIncrement"` // Auto-incrementing primary key
	Condition    string `gorm:"type:varchar(255);not null"`
	Then         string `gorm:"type:varchar(255);not null"`
	IDIfRight    string `gorm:"type:varchar(255);not null"`
	IDIfNotRight string `gorm:"type:varchar(255);not null"`
}
