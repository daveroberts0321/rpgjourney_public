// models/user.go
package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email    string `gorm:"unique"` // unique email for authentication
	Password string
	Role     string
	Active   bool // active status
}

type UserProfile struct {
	gorm.Model
	UserID     uint
	Username   string // unique username for display
	FirstName  string
	Level      uint // user level ( completed quest /20 = level )
	LastName   string
	ProfileImg string
	Email      string
	Phone      string
	Address    string
	City       string
	State      string
	Zip        string
	Image      string
	Notes      string
}
