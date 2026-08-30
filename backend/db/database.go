package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

// Database connection
var DB *sql.DB

// Initialize the database connection
func InitDB() {
	// Load .env file if exists
	godotenv.Load()

	host := GetEnv("DB_HOST", "localhost")
	port := GetEnv("DB_PORT", "5432")
	user := GetEnv("DB_USER", "skillsifter")
	password := GetEnv("DB_PASSWORD", "ROOT")
	dbname := GetEnv("DB_NAME", "postgres")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s search_path=public sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Could not ping database: %v", err)
	}

	fmt.Println("Successfully connected to database")
}
