package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// Database connection
var DB *sql.DB

var tenantDBs sync.Map

// WithTenantDB stores the authenticated tenant database in the request context.
func WithTenantDB(ctx context.Context, tenantDB *sql.DB) context.Context {
	return context.WithValue(ctx, tenantDBKey{}, tenantDB)
}

type tenantDBKey struct{}

// RequestDB returns the request-scoped tenant database. Protected HTTP routes
// install tenant routing before tenant-owned handlers execute; the control DB
// fallback preserves direct handler tests and non-routed control-plane handlers.
func RequestDB(r *http.Request) *sql.DB {
	if tenantDB, ok := r.Context().Value(tenantDBKey{}).(*sql.DB); ok && tenantDB != nil {
		return tenantDB
	}
	return DB
}

// TenantDB resolves a READY tenant database from the control plane and caches
// its connection pool for reuse across requests.
func TenantDB(controlDB *sql.DB, tenantID string) (*sql.DB, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant id is required")
	}
	if cached, ok := tenantDBs.Load(tenantID); ok {
		return cached.(*sql.DB), nil
	}

	var databaseName, status string
	if err := controlDB.QueryRow(
		`SELECT COALESCE(tenant_database,''), provisioning_status
		 FROM platform_tenants WHERE tenant_id=$1`, tenantID,
	).Scan(&databaseName, &status); err != nil {
		return nil, fmt.Errorf("resolve tenant database: %w", err)
	}
	if status != "READY" || databaseName == "" {
		return nil, fmt.Errorf("tenant database is not ready")
	}

	conn, err := OpenDatabase(databaseName)
	if err != nil {
		return nil, fmt.Errorf("open tenant database: %w", err)
	}
	actual, loaded := tenantDBs.LoadOrStore(tenantID, conn)
	if loaded {
		conn.Close()
		return actual.(*sql.DB), nil
	}
	return conn, nil
}

// CloseTenantDatabases closes all cached tenant pools during shutdown.
func CloseTenantDatabases() {
	tenantDBs.Range(func(_, value interface{}) bool {
		value.(*sql.DB).Close()
		return true
	})
	tenantDBs = sync.Map{}
}

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

func setTenantProvisioningStatus(controlDB *sql.DB, tenantID, status, databaseName string) {
	_, _ = controlDB.Exec(
		`UPDATE platform_tenants
		 SET provisioning_status=$1, tenant_database=NULLIF($2,''), updated_at=NOW()
		 WHERE tenant_id=$3`,
		status, databaseName, tenantID,
	)
}

// ProvisionTenantDatabase creates or resumes a deterministic tenant database,
// initializes tenant-owned schema definitions, and mirrors the tenant's users.
// It is safe to call repeatedly.
func ProvisionTenantDatabase(controlDB *sql.DB, tenantID, companyName string) (databaseName string, err error) {
	if tenantID == "" || companyName == "" {
		return "", fmt.Errorf("tenant id and company name are required")
	}

	databaseName = TenantDatabaseName(tenantID)
	setTenantProvisioningStatus(controlDB, tenantID, "PROVISIONING", databaseName)

	defer func() {
		if err != nil {
			setTenantProvisioningStatus(controlDB, tenantID, "FAILED", databaseName)
		}
	}()

	adminDB, err := OpenDatabase(GetEnv("DB_NAME", "postgres"))
	if err != nil {
		return "", fmt.Errorf("open control database for provisioning: %w", err)
	}
	defer adminDB.Close()

	var exists bool
	if err = adminDB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)`,
		databaseName,
	).Scan(&exists); err != nil {
		return "", fmt.Errorf("check tenant database: %w", err)
	}

	if !exists {
		if _, err = adminDB.Exec(`CREATE DATABASE ` + databaseName); err != nil && !strings.Contains(err.Error(), "already exists") {
			return "", fmt.Errorf("create tenant database: %w", err)
		}
	}

	tenantDB, err := OpenDatabase(databaseName)
	if err != nil {
		return "", fmt.Errorf("open tenant database: %w", err)
	}
	defer tenantDB.Close()

	if err = InitializeTenantSchema(tenantDB); err != nil {
		return "", err
	}

	rows, err := controlDB.Query(
		`SELECT id, username, email, password, role, tenant_id, company_name, created_at
		 FROM users WHERE tenant_id=$1 ORDER BY id`,
		tenantID,
	)
	if err != nil {
		return "", fmt.Errorf("read tenant users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var username, email, password, role, userTenantID, userCompany string
		var createdAt time.Time
		if err := rows.Scan(&id, &username, &email, &password, &role, &userTenantID, &userCompany, &createdAt); err != nil {
			return "", fmt.Errorf("read tenant user: %w", err)
		}
		if _, err := tenantDB.Exec(
			`INSERT INTO users(id,username,email,password,role,tenant_id,company_name,created_at)
			 VALUES($1,$2,$3,$4,$5,$6,$7,$8)
			 ON CONFLICT(id) DO UPDATE
			 SET username=EXCLUDED.username,
			     email=EXCLUDED.email,
			     password=EXCLUDED.password,
			     role=EXCLUDED.role,
			     tenant_id=EXCLUDED.tenant_id,
			     company_name=EXCLUDED.company_name`,
			id, username, email, password, role, userTenantID, userCompany, createdAt,
		); err != nil {
			return "", fmt.Errorf("seed tenant user %d: %w", id, err)
		}
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("read tenant users: %w", err)
	}

	if _, err = tenantDB.Exec(
		`SELECT setval(pg_get_serial_sequence('users','id'), COALESCE(MAX(id),1), true) FROM users`,
	); err != nil {
		return "", fmt.Errorf("sync tenant user sequence: %w", err)
	}

	setTenantProvisioningStatus(controlDB, tenantID, "READY", databaseName)
	return databaseName, nil
}

// DeleteTenantDatabase permanently removes the tenant database and is intended
// only for expired, unsubscribed trial tenants after the retention window.
func DeleteTenantDatabase(controlDB *sql.DB, tenantID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant id is required")
	}
	databaseName := TenantDatabaseName(tenantID)
	if cached, ok := tenantDBs.Load(tenantID); ok {
		cached.(*sql.DB).Close()
		tenantDBs.Delete(tenantID)
	}

	adminDB, err := OpenDatabase(GetEnv("DB_NAME", "postgres"))
	if err != nil {
		return fmt.Errorf("open control database for tenant deletion: %w", err)
	}
	defer adminDB.Close()

	if _, err := adminDB.Exec("DROP DATABASE IF EXISTS " + databaseName + " WITH (FORCE)"); err != nil {
		return fmt.Errorf("drop tenant database: %w", err)
	}
	return nil
}
