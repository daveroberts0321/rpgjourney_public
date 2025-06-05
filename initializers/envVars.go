package initializers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func LoadEnvVars() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// generate a new crypto key( only used when rolling a new key)
func GenerateAESKey() ([]byte, error) {
	// AES encryption 256 bit encrypt/deccrypy functions in helpers/cryptohelpers
	key := make([]byte, 32) // For AES-256
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	fmt.Println("AES Key:", base64.StdEncoding.EncodeToString(key))
	return key, nil
}
