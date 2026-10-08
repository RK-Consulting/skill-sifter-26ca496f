package db

import (
    "context"
    "crypto/sha256"
    "database/sql"
    "encoding/hex"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "sort"
    "strconv"
)

var schemaSeqPattern = regexp.MustCompile(`^(\d+)_`)
const controlSchemaLockKey = "skill-sifter:control-schema"
const tenantSchemaLockKey = "skill-sifter:tenant-schema"

type schemaDefinition struct {
    version int
    name    string
    path    string
}

type appliedSchemaVersion struct {
    version  int
    name     string
    checksum string
}

func schemaDefinitionsDir(kind string) (string, error) {
    candidates := []string{
        filepath.Join("database", kind),
        filepath.Join("backend", "database", kind),
    }
    for _, c := range candidates {
        if info, err := os.Stat(c); err == nil && info.IsDir() {
            return c, nil
        }
    }
    return "", fmt.Errorf("could not locate %s schema directory (tried: %v)", kind, candidates)
}

func discoverSchemaDefinitions(dir string) ([]schemaDefinition, error) {
    entries, err := os.ReadDir(dir)
    if err != nil {
        return nil, fmt.Errorf("could not read schema definitions directory %q: %w", dir, err)
    }

    files := make([]schemaDefinition, 0, len(entries))
    seen := map[int]string{}

    for _, e := range entries {
        if e.IsDir() || filepath.Ext(e.Name()) != ".sql" {
            continue
        }

        m := schemaSeqPattern.FindStringSubmatch(e.Name())
        if m == nil {
            return nil, fmt.Errorf("schema definition %q does not start with a numeric sequence prefix", e.Name())
        }

        version, err := strconv.Atoi(m[1])
        if err != nil {
            return nil, fmt.Errorf("schema definition %q has an invalid sequence prefix: %w", e.Name(), err)
        }
        if prior, exists := seen[version]; exists {
            return nil, fmt.Errorf("duplicate schema definition sequence %d: %q and %q", version, prior, e.Name())
        }
        seen[version] = e.Name()

        files = append(files, schemaDefinition{version: version, name: e.Name(), path: filepath.Join(dir, e.Name())})
    }

    sort.Slice(files, func(i, j int) bool { return files[i].version < files[j].version })
    return files, nil
}

func schemaDefinitionChecksum(path string) (string, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return "", err
    }
    sum := sha256.Sum256(data)
    return hex.EncodeToString(sum[:]), nil
}

func ensureSchemaVersionsTable(conn *sql.DB) error {
    _, err := conn.Exec(`
        CREATE TABLE IF NOT EXISTS schema_versions (
            version INTEGER PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            checksum VARCHAR(64) NOT NULL,
            applied_at TIMESTAMP NOT NULL DEFAULT NOW()
        )`)
    if err != nil {
        return fmt.Errorf("could not create schema_versions table: %w", err)
    }
    return nil
}

func appliedSchemaVersions(conn *sql.DB) (map[int]appliedSchemaVersion, error) {
    rows, err := conn.Query(`
        SELECT version, name, checksum
        FROM schema_versions
        ORDER BY version`)
    if err != nil {
        return nil, fmt.Errorf("could not read schema_versions: %w", err)
    }
    defer rows.Close()

    applied := map[int]appliedSchemaVersion{}
    for rows.Next() {
        var v appliedSchemaVersion
        if err := rows.Scan(&v.version, &v.name, &v.checksum); err != nil {
            return nil, fmt.Errorf("could not scan schema_versions row: %w", err)
        }
        applied[v.version] = v
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("could not read schema_versions rows: %w", err)
    }
    return applied, nil
}

func withSchemaLock(conn *sql.DB, lockKey string, fn func() error) error {
    dedicated, err := conn.Conn(context.Background())
    if err != nil {
        return fmt.Errorf("could not acquire schema database connection: %w", err)
    }
    defer dedicated.Close()

    if _, err := dedicated.ExecContext(context.Background(),
        `SELECT pg_advisory_lock(hashtext($1)::bigint)`, lockKey); err != nil {
        return fmt.Errorf("could not acquire schema lock: %w", err)
    }
    defer func() {
        _, _ = dedicated.ExecContext(context.Background(),
            `SELECT pg_advisory_unlock(hashtext($1)::bigint)`, lockKey)
    }()

    return fn()
}

func initializeSchema(conn *sql.DB, dir, lockKey string) error {
    return withSchemaLock(conn, lockKey, func() error {
        if err := ensureSchemaVersionsTable(conn); err != nil {
            return err
        }

        files, err := discoverSchemaDefinitions(dir)
        if err != nil {
            return err
        }

        applied, err := appliedSchemaVersions(conn)
        if err != nil {
            return err
        }

        filesByVersion := make(map[int]schemaDefinition, len(files))
        for _, f := range files {
            filesByVersion[f.version] = f
        }

        for version, recorded := range applied {
            f, exists := filesByVersion[version]
            if !exists {
                return fmt.Errorf("applied schema definition %d (%s) is missing", version, recorded.name)
            }
            if f.name != recorded.name {
                return fmt.Errorf("schema definition %d filename mismatch: database records %q, filesystem contains %q", version, recorded.name, f.name)
            }
            checksum, err := schemaDefinitionChecksum(f.path)
            if err != nil {
                return fmt.Errorf("schema definition %d (%s): checksum calculation failed: %w", version, f.name, err)
            }
            if checksum != recorded.checksum {
                return fmt.Errorf("schema definition %d (%s) checksum mismatch: database=%s filesystem=%s", version, f.name, recorded.checksum, checksum)
            }
        }

        for _, f := range files {
            if _, alreadyApplied := applied[f.version]; alreadyApplied {
                continue
            }
            if err := applySchemaDefinition(conn, f); err != nil {
                return err
            }
            fmt.Printf("Applied schema definition %d (%s) from %s\n", f.version, f.name, dir)
        }
        return nil
    })
}

func applySchemaDefinition(conn *sql.DB, f schemaDefinition) error {
    data, err := os.ReadFile(f.path)
    if err != nil {
        return fmt.Errorf("schema definition %d (%s): could not read file: %w", f.version, f.name, err)
    }

    sum := sha256.Sum256(data)
    checksum := hex.EncodeToString(sum[:])

    tx, err := conn.Begin()
    if err != nil {
        return fmt.Errorf("schema definition %d (%s): could not start transaction: %w", f.version, f.name, err)
    }

    if _, err := tx.Exec(string(data)); err != nil {
        _ = tx.Rollback()
        return fmt.Errorf("schema definition %d (%s) failed: %w", f.version, f.name, err)
    }

    if _, err := tx.Exec(
        `INSERT INTO schema_versions(version, name, checksum) VALUES($1, $2, $3)`,
        f.version, f.name, checksum); err != nil {
        _ = tx.Rollback()
        return fmt.Errorf("schema definition %d (%s): could not record success: %w", f.version, f.name, err)
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("schema definition %d (%s): could not commit: %w", f.version, f.name, err)
    }
    return nil
}

// InitializeSchema initializes the authoritative control-plane schema.
func InitializeSchema() error {
    dir, err := schemaDefinitionsDir("control-plane")
    if err != nil {
        return err
    }
    return initializeSchema(DB, dir, controlSchemaLockKey)
}

// InitializeControlSchema is the explicit name for control-plane initialization.
func InitializeControlSchema() error {
    return InitializeSchema()
}

// InitializeTenantSchema initializes the authoritative tenant-plane schema
// against the supplied tenant database. It never consults the control DB schema.
func InitializeTenantSchema(tenantDB *sql.DB) error {
    if tenantDB == nil {
        return fmt.Errorf("tenant database connection is required")
    }
    dir, err := schemaDefinitionsDir("tenant-plane")
    if err != nil {
        return err
    }
    return initializeSchema(tenantDB, dir, tenantSchemaLockKey)
}
