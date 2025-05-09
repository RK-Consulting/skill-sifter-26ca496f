
package db

import (
	"database/sql"
	"fmt"
	"log"
	//"os"
	"reflect"
	"strings"
	//"time"

	"github.com/RK-Consulting/skill-sifter/models"
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

// One-time schema initialization function
func InitializeSchema() error {
	// Check if schema_version table exists
	var exists bool
	err := DB.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'schema_version'
		)
	`).Scan(&exists)
	
	if err != nil {
		return fmt.Errorf("error checking schema_version table: %v", err)
	}
	
	// If schema_version table doesn't exist, create it and initialize schema
	if !exists {
		tx, err := DB.Begin()
		if err != nil {
			return fmt.Errorf("error starting transaction: %v", err)
		}
		defer tx.Rollback()
		
		// Create schema_version table
		_, err = tx.Exec(`
			CREATE TABLE public.schema_version (
				version INTEGER PRIMARY KEY,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			return fmt.Errorf("error creating schema_version table: %v", err)
		}
		
		// Create tables in correct order based on dependencies
		fmt.Println("Initializing database schema...")
		
		// 1. Companies table (no dependencies)
		if err := createTableFromStruct(tx, models.Company{}, "companies"); err != nil {
			return err
		}
		
		// 2. Roles table (no dependencies)
		if err := createTableFromStruct(tx, models.Role{}, "roles"); err != nil {
			return err
		}
		
		// 3. Users table (depends on roles and companies)
		if err := createTableFromStruct(tx, models.User{}, "users"); err != nil {
			return err
		}
		
		// 4. Candidates table (depends on companies)
		if err := createTableFromStruct(tx, models.Candidate{}, "candidates"); err != nil {
			return err
		}
		
		// 5. Jobs table (depends on companies)
		if err := createTableFromStruct(tx, models.Job{}, "jobs"); err != nil {
			return err
		}
		
		// 6. Daily Jobs table (depends on companies and users)
		if err := createTableFromStruct(tx, models.DailyJob{}, "daily_jobs"); err != nil {
			return err
		}
		
		// 7. Interviews table (depends on companies and candidates)
		if err := createTableFromStruct(tx, models.Interview{}, "interviews"); err != nil {
			return err
		}
		
		// Create indexes for performance
		_, err = tx.Exec(`
			CREATE INDEX IF NOT EXISTS idx_candidates_company ON candidates(company_name);
			CREATE INDEX IF NOT EXISTS idx_jobs_company ON jobs(company_name);
			CREATE INDEX IF NOT EXISTS idx_daily_jobs_company ON daily_jobs(company_name);
			CREATE INDEX IF NOT EXISTS idx_interviews_company ON interviews(company_name);
			CREATE INDEX IF NOT EXISTS idx_users_company ON users(company_name);
			CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
		`)
		
		if err != nil {
			return fmt.Errorf("error creating indexes: %v", err)
		}
		
		// Insert default roles
		_, err = tx.Exec(`
			INSERT INTO roles (name, permissions, created_at) VALUES 
			('admin', '["all"]', NOW()),
			('manager', '["manage_candidates", "manage_jobs", "manage_interviews", "view_reports"]', NOW()),
			('recruiter', '["view_candidates", "add_candidates", "view_jobs", "schedule_interviews"]', NOW()),
			('team_leader', '["view_candidates", "add_candidates", "view_jobs", "manage_team"]', NOW())
			ON CONFLICT (name) DO NOTHING;
		`)
		
		if err != nil {
			return fmt.Errorf("error inserting default roles: %v", err)
		}
		
		// Insert schema version
		_, err = tx.Exec("INSERT INTO schema_version (version) VALUES (1)")
		if err != nil {
			return fmt.Errorf("error inserting schema version: %v", err)
		}
		
		// Commit the transaction
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("error committing schema transaction: %v", err)
		}
		
		fmt.Println("Database schema initialized successfully!")
	} else {
		fmt.Println("Database schema already initialized, skipping...")
	}
	
	return nil
}

// Helper function to create a table from a struct
func createTableFromStruct(tx *sql.Tx, model interface{}, tableName string) error {
	// Check if table already exists
	var exists bool
	err := tx.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)
	`, tableName).Scan(&exists)
	
	if err != nil {
		return fmt.Errorf("error checking if %s table exists: %v", tableName, err)
	}
	
	if exists {
		fmt.Printf("Table %s already exists, skipping creation\n", tableName)
		return nil
	}
	
	fmt.Printf("Creating table %s...\n", tableName)
	
	t := reflect.TypeOf(model)
	var columns []string
	var foreignKeys []string
	
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		
		if dbTag == "" {
			continue // Skip fields without db tag
		}
		
		parts := strings.Split(dbTag, ",")
		colName := parts[0]
		
		if colName == "-" {
			continue // Skip explicitly ignored fields
		}
		
		colType := getPostgresType(field.Type)
		colDef := fmt.Sprintf("%s %s", colName, colType)
		
		// Add constraints from tags
		for _, part := range parts[1:] {
			switch {
			case part == "primarykey":
				colDef += " PRIMARY KEY"
			case part == "autoincrement":
				// For Postgres, we change int to SERIAL when it has autoincrement
				if strings.Contains(colType, "INTEGER") {
					colDef = fmt.Sprintf("%s SERIAL PRIMARY KEY", colName)
				}
			case part == "notnull":
				colDef += " NOT NULL"
			case part == "unique":
				colDef += " UNIQUE"
			case strings.HasPrefix(part, "default:"):
				defaultVal := strings.TrimPrefix(part, "default:")
				colDef += fmt.Sprintf(" DEFAULT %s", defaultVal)
			case strings.HasPrefix(part, "foreignkey:"):
				fkRef := strings.TrimPrefix(part, "foreignkey:")
				foreignKeys = append(foreignKeys, fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s", colName, fkRef))
				continue // Don't add this part to the column definition
			case strings.HasPrefix(part, "type:"):
				// Override detected type with explicit type
				typeVal := strings.TrimPrefix(part, "type:")
				colDef = fmt.Sprintf("%s %s", colName, typeVal)
			}
		}
		
		columns = append(columns, colDef)
	}
	
	// Add foreign key constraints at the end
	columns = append(columns, foreignKeys...)
	
	createSQL := fmt.Sprintf("CREATE TABLE %s (\n\t%s\n)", tableName, strings.Join(columns, ",\n\t"))
	_, err = tx.Exec(createSQL)
	
	if err != nil {
		return fmt.Errorf("error creating %s table: %v\nSQL: %s", tableName, err, createSQL)
	}
	
	return nil
}

// Helper function to convert Go types to PostgreSQL types
func getPostgresType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "INTEGER"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "INTEGER"
	case reflect.Float32, reflect.Float64:
		return "NUMERIC"
	case reflect.String:
		return "VARCHAR(255)"
	default:
		// Handle special types
		typeName := t.String()
		switch typeName {
		case "time.Time":
			return "TIMESTAMP"
		case "[]string":
			return "TEXT[]"
		case "map[string]interface {}", "map[string]interface{}":
			return "JSONB"
		default:
			return "TEXT"
		}
	}
}
