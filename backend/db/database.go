package db

import (
	"database/sql"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

// Database connection
var DB *sql.DB

// OpenDatabase opens a PostgreSQL connection using the application's database credentials.
func OpenDatabase(dbname string) (*sql.DB, error) {
	host := GetEnv("DB_HOST", "localhost")
	port := GetEnv("DB_PORT", "5432")
	user := GetEnv("DB_USER", "skillsifter")
	password := GetEnv("DB_PASSWORD", "ROOT")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s search_path=public sslmode=disable",
		host, port, user, password, dbname)

	conn, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

// TenantDatabaseName returns a deterministic, safe PostgreSQL database name.
func TenantDatabaseName(tenantID string) string {
	sum := sha256.Sum256([]byte(tenantID))
	return "skillsifter_t_" + hex.EncodeToString(sum[:])[:20]
}

// InitDB initializes the application's control-plane database connection.
func InitDB() {
	godotenv.Load()

	var err error
	DB, err = OpenDatabase(GetEnv("DB_NAME", "postgres"))
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	fmt.Println("Successfully connected to database")
}
