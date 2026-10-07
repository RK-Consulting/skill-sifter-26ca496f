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

// TenantUser is the minimal tenant-local account created during provisioning.
// It is deliberately not a control-plane user record.
type TenantUser struct {
    ID       int
    Username string
    Email    string
    Password string
    Role     string
}

// ProvisionTenantDatabase creates or resumes a deterministic tenant database,
// initializes the final tenant schema, and optionally creates the initial
// tenant-local user. It never copies users from the control plane.
func ProvisionTenantDatabase(controlDB *sql.DB, tenantID string, initialUser *TenantUser) (databaseName string, tenantUserID int, err error) {
    if tenantID == "" {
        return "", 0, fmt.Errorf("tenant id is required")
    }

    lockConn, err := controlDB.Conn(context.Background())
    if err != nil { return "", 0, fmt.Errorf("acquire provisioning connection: %w", err) }
    defer lockConn.Close()
    if _, err = lockConn.ExecContext(context.Background(), "SELECT pg_advisory_lock(hashtext($1)::bigint)", "skill-sifter:provision:"+tenantID); err != nil {
        return "", 0, fmt.Errorf("acquire provisioning lock: %w", err)
    }
    defer func() { _, _ = lockConn.ExecContext(context.Background(), "SELECT pg_advisory_unlock(hashtext($1)::bigint)", "skill-sifter:provision:"+tenantID) }()

    databaseName = TenantDatabaseName(tenantID)
    if _, err = controlDB.Exec(
        `UPDATE platform_tenants
         SET provisioning_status='PROVISIONING', tenant_database=$1, updated_at=NOW()
         WHERE tenant_id=$2 AND provisioning_status IN ('PENDING','PROVISIONING','FAILED','READY')`,
        databaseName, tenantID,
    ); err != nil {
        return "", 0, fmt.Errorf("mark tenant provisioning: %w", err)
    }

    defer func() {
        if err != nil {
            _, _ = controlDB.Exec(
                `UPDATE platform_tenants
                 SET provisioning_status='FAILED', updated_at=NOW()
                 WHERE tenant_id=$1`, tenantID)
        }
    }()

    adminDB, err := OpenDatabase(GetEnv("DB_NAME", "postgres"))
    if err != nil {
        return "", 0, fmt.Errorf("open control database for provisioning: %w", err)
    }
    defer adminDB.Close()

    var exists bool
    if err = adminDB.QueryRow(
        `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)`, databaseName,
    ).Scan(&exists); err != nil {
        return "", 0, fmt.Errorf("check tenant database: %w", err)
    }

    if !exists {
        if _, err = adminDB.Exec(`CREATE DATABASE ` + databaseName); err != nil &&
            !strings.Contains(err.Error(), "already exists") {
            return "", 0, fmt.Errorf("create tenant database: %w", err)
        }
    }

    tenantDB, err := OpenDatabase(databaseName)
    if err != nil {
        return "", 0, fmt.Errorf("open tenant database: %w", err)
    }
    defer tenantDB.Close()

    if err = InitializeTenantSchema(tenantDB); err != nil {
        return "", 0, err
    }

    if initialUser != nil {
        if initialUser.Username == "" || initialUser.Email == "" || initialUser.Password == "" || initialUser.Role == "" {
            return "", 0, fmt.Errorf("initial tenant user is incomplete")
        }

        if initialUser.ID > 0 {
            err = tenantDB.QueryRow(
                `INSERT INTO users(id, username, email, password, role, tenant_id)
                 VALUES($1,$2,$3,$4,$5,$6)
                 ON CONFLICT(id) DO UPDATE
                 SET username=EXCLUDED.username,
                     email=EXCLUDED.email,
                     password=EXCLUDED.password,
                     role=EXCLUDED.role,
                     tenant_id=EXCLUDED.tenant_id
                 RETURNING id`,
                initialUser.ID, initialUser.Username, initialUser.Email,
                initialUser.Password, initialUser.Role, tenantID,
            ).Scan(&tenantUserID)
        } else {
            err = tenantDB.QueryRow(
                `INSERT INTO users(username, email, password, role, tenant_id)
                 VALUES($1,$2,$3,$4,$5)
                 ON CONFLICT(email) DO UPDATE
                 SET username=EXCLUDED.username,
                     password=EXCLUDED.password,
                     role=EXCLUDED.role,
                     tenant_id=EXCLUDED.tenant_id
                 RETURNING id`,
                initialUser.Username, initialUser.Email, initialUser.Password,
                initialUser.Role, tenantID,
            ).Scan(&tenantUserID)
        }
        if err != nil {
            return "", 0, fmt.Errorf("create tenant user: %w", err)
        }

        if _, err = tenantDB.Exec(
            `SELECT setval(pg_get_serial_sequence('users','id'), GREATEST(COALESCE(MAX(id),1),1), true) FROM users`,
        ); err != nil {
            return "", 0, fmt.Errorf("sync tenant user sequence: %w", err)
        }
    }

    if _, err = controlDB.Exec(
        `UPDATE platform_tenants
         SET provisioning_status='READY', tenant_database=$1, updated_at=NOW()
         WHERE tenant_id=$2`,
        databaseName, tenantID,
    ); err != nil {
        return "", 0, fmt.Errorf("mark tenant ready: %w", err)
    }

    return databaseName, tenantUserID, nil
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
