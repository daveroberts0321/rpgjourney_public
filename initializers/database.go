package initializers

import (
	"log"
	"os"

	"github.com/daveroberts0321/rpgbackend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("SUPABASE_DSN")
	var err error
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
}

func SyncDB() {
	DB.AutoMigrate(&models.User{})
	DB.AutoMigrate(&models.UserProfile{})
	DB.AutoMigrate(&models.Quest{})
	DB.AutoMigrate(&models.QuestLog{})
	DB.AutoMigrate(&models.History{})
	DB.AutoMigrate(&models.BlogPost{})
}
