package models

import "gorm.io/gorm"

type Quest struct {
	gorm.Model
	UserID      uint
	Category    string // strength, health, wealth, wisdom, skills
	Title       string
	Description string
	Goalmet     uint    // number of times the target has been met
	Unit        string  // unit of measurement (e.g. lbs, reps, miles, minutes, dollars, etc.)
	Ascending   bool    // true if the quest is ascending (e.g. weight lifted, reps, miles, etc.), false if descending (e.g. time, weight loss, etc.)
	Starting    float32 // starting value
	Target      float32 // target value
	Current     float32 // current value
	Ending      float32 // ending value
	Completed   bool
	Notes       string // notes or comments about the quest
}

type QuestLog struct {
	gorm.Model
	UserID   uint
	Strength uint // strength quest completed/10 = strength level
	Health   uint // health quest completed
	Wealth   uint // wealth quest completed
	Wisdom   uint // wisdom quest completed
	Skills   uint // skills quest completed
	Notes    string
}

type History struct {
	gorm.Model
	QuestID uint
	UserID  uint
	Unit    string
	Amount  float32
	Notes   string
}
