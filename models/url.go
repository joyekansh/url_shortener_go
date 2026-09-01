package models

import "gorm.io/gorm"

type URL struct {
	gorm.Model
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url" gorm:"uniqueIndex"`
}
