package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

// This file owns the application's deterministic database schema setup.
// Schema definitions are ordered by their numeric filename prefix, recorded
// in schema_versions, and protected by content checksums. Concurrent
// application startups are serialized with a PostgreSQL advisory lock.
var schemaSeqPattern = regexp.MustCompile(`^(\d+)_`)

const schemaLockKey = "skill-sifter:schema"

// schemaDefinitionsDir locates the authoritative schema definitions directory.
func schemaDefinitionsDir() (string, error) {
	candidates := []string{"database/migrations", "backend/database/migrations"}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("could not locate schema definitions directory (tried: %v)", candidates)
}

// ensureSchemaVersionsTable creates the schema version table for a clean
// installation. The table is intentionally defined in its final form: there
// is no compatibility or upgrade path for an older schema-version table.
func ensureSchemaVersionsTable() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_versions (
			version    INTEGER PRIMARY KEY,
			name       VARCHAR(255) NOT NULL,
			checksum   VARCHAR(64) NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`)
	if err != nil {
		return fmt.Errorf("could not create schema_versions table: %w", err)
	}
	return nil
}

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

// discoverSchemaDefinitions reads every SQL schema definition, validates its
// numeric sequence, and returns the files in ascending execution order.
func discoverSchemaDefinitions(dir string) ([]schemaDefinition, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read schema definitions directory %q: %w", dir, err)
	}

	var files []schemaDefinition
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
			return nil, fmt.Errorf("schema definition %q has an unparseable sequence prefix: %w", e.Name(), err)
		}

		if prior, exists := seen[version]; exists {
			return nil, fmt.Errorf("duplicate schema definition sequence %d: both %q and %q claim it", version, prior, e.Name())
		}
		seen[version] = e.Name()

		files = append(files, schemaDefinition{
			version: version,
			name:    e.Name(),
			path:    filepath.Join(dir, e.Name()),
		})
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

func appliedSchemaVersions() (map[int]appliedSchemaVersion, error) {
	rows, err := DB.Query(`
		SELECT version, name, checksum
		FROM schema_versions
		ORDER BY version
	`)
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

// withSchemaLock serializes schema initialization across application
// instances. The lock is session-scoped and is released on the same
// dedicated database connection.
func withSchemaLock(fn func() error) error {
	conn, err := DB.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("could not acquire schema database connection: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(context.Background(),
		`SELECT pg_advisory_lock(hashtext($1)::bigint)`,
		schemaLockKey,
	); err != nil {
		return fmt.Errorf("could not acquire schema lock: %w", err)
	}

	defer func() {
		_, _ = conn.ExecContext(context.Background(),
			`SELECT pg_advisory_unlock(hashtext($1)::bigint)`,
			schemaLockKey,
		)
	}()

	return fn()
}

// InitializeSchema discovers the authoritative schema definitions, verifies
// recorded definitions, applies pending definitions in order, and records
// each successful definition atomically.
func InitializeSchema() error {
	dir, err := schemaDefinitionsDir()
	if err != nil {
		return err
	}
	return initializeSchemaFromDir(dir)
}

func initializeSchemaFromDir(dir string) error {
	return withSchemaLock(func() error {
		if err := ensureSchemaVersionsTable(); err != nil {
			return err
		}

		files, err := discoverSchemaDefinitions(dir)
		if err != nil {
			return err
		}

		applied, err := appliedSchemaVersions()
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
				return fmt.Errorf("schema definition %d (%s): could not calculate checksum: %w", version, f.name, err)
			}
			if checksum != recorded.checksum {
				return fmt.Errorf("schema definition %d (%s) checksum mismatch: database=%s filesystem=%s", version, f.name, recorded.checksum, checksum)
			}
		}

		for _, f := range files {
			if _, alreadyApplied := applied[f.version]; alreadyApplied {
				continue
			}
			if err := applySchemaDefinition(f); err != nil {
				return err
			}
			fmt.Printf("Applied schema definition %d (%s)\n", f.version, f.name)
		}

		return nil
	})
}

func applySchemaDefinition(f schemaDefinition) error {
	data, err := os.ReadFile(f.path)
	if err != nil {
		return fmt.Errorf("schema definition %d (%s): could not read file: %w", f.version, f.name, err)
	}

	sum := sha256.Sum256(data)
	checksum := hex.EncodeToString(sum[:])

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("schema definition %d (%s): could not start transaction: %w", f.version, f.name, err)
	}

	if _, err := tx.Exec(string(data)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("schema definition %d (%s) failed: %w", f.version, f.name, err)
	}

	if _, err := tx.Exec(
		`INSERT INTO schema_versions (version, name, checksum) VALUES ($1, $2, $3)`,
		f.version, f.name, checksum,
	); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("schema definition %d (%s): could not record success: %w", f.version, f.name, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("schema definition %d (%s): could not commit: %w", f.version, f.name, err)
	}

	return nil
}
