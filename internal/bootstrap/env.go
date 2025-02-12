package bootstrap

import (
	"log"

	"github.com/joho/godotenv"
)

func Loadenv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
}
