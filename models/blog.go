package models

import (
	"html/template"

	"gorm.io/gorm"
)

type BlogPost struct {
	gorm.Model
	UserID     uint
	Title      string
	Content    template.HTML `gorm:"type:text"`
	ProfileImg string
	BlogImg    string
	Username   string
}
