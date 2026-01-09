package schema

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

const (
	schemaVersionTableName = "schema_version"
)

//go:embed files/**/*.sql
var schemaFiles embed.FS

// Returns latest available schema version available
func GetLatestAvailableVersion() (latestVersion int, err error) {
	entries, err := schemaFiles.ReadDir("files")
	if err != nil {
		return 0, fmt.Errorf("failed to read schema files: %w", err)
	}

	for _, entry := range entries {
		var version int
		_, err = fmt.Sscanf(entry.Name(), "v%d", &version)
		if err == nil && version > latestVersion {
			latestVersion = version
		}
	}

	return
}

// Returns current schema version in the database
func GetCurrentVersion(db *sql.DB) (version int, err error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		err = errors.Join(err, tx.Rollback())
	}()

	version, err = getCurrentVersion(tx)
	return
}

// Upgrades the database schema to the latest available version
func Upgrade(db *sql.DB, targetVersion int) (err error) {
	if targetVersion <= 0 {
		return fmt.Errorf("invalid target schema version: %d", targetVersion)
	}

	latestVersion, err := GetLatestAvailableVersion()
	if err != nil {
		return fmt.Errorf("failed to get latest available schema version: %w", err)
	}

	if targetVersion > latestVersion {
		return fmt.Errorf("target schema version %d is greater than latest available version %d", targetVersion, latestVersion)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		rbErr := tx.Rollback()
		if !errors.Is(rbErr, sql.ErrTxDone) {
			err = errors.Join(err, rbErr)
		}
	}()

	currentVersion, err := getCurrentVersion(tx)
	if err != nil {
		err = fmt.Errorf("failed to get current schema version: %w", err)
		return
	}

	// Ensure schema version table exists
	_, err = tx.Exec(fmt.Sprintf(`
	 	CREATE SCHEMA IF NOT EXISTS meta;
		CREATE TABLE IF NOT EXISTS meta.%s (
			ID INTEGER PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);
	`, pq.QuoteIdentifier(schemaVersionTableName)))
	if err != nil {
		err = fmt.Errorf("failed to create schema version table: %w", err)
		return
	}

	for v := currentVersion + 1; v <= latestVersion; v++ {
		// Read file names for the version
		var dirEntries []fs.DirEntry
		dirEntries, err = schemaFiles.ReadDir(fmt.Sprintf("files/v%d", v))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue // No files for this version, skip
			}
			err = fmt.Errorf("failed to read schema files for version %d: %w", v, err)
			return
		}

		// Sort files by name to ensure consistent order
		slices.SortFunc(dirEntries, func(a fs.DirEntry, b fs.DirEntry) int {
			return strings.Compare(a.Name(), b.Name())
		})

		// Apply each SQL file
		for _, entry := range dirEntries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
				continue
			}

			var sql []byte
			sql, err = schemaFiles.ReadFile(fmt.Sprintf("files/v%d/%s", v, entry.Name()))
			if err != nil {
				err = fmt.Errorf("failed to read schema file %s for version %d: %w", entry.Name(), v, err)
				return
			}

			_, err = tx.Exec(string(sql))
			if err != nil {
				err = fmt.Errorf("failed to apply schema for version %d (%s): %w", v, entry.Name(), err)
				return
			}

		}

		// Update schema version table
		_, err = sq.
			Insert("meta." + schemaVersionTableName).
			Columns("ID").
			Values(v).
			PlaceholderFormat(sq.Dollar).
			RunWith(tx).
			Exec()
		if err != nil {
			err = fmt.Errorf("failed to update schema version to %d: %w", v, err)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		err = fmt.Errorf("failed to commit schema upgrade transaction: %w", err)
		return
	}

	return nil
}

// Purges the database schema (drops all tables)
func Purge(db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		rbErr := tx.Rollback()
		if !errors.Is(rbErr, sql.ErrTxDone) {
			err = errors.Join(err, rbErr)
		}
	}()

	var q string

	tables := []string{"postfix", "dovecot", "stalwart", "audit", "public", "shared", "meta"}
	for _, table := range tables {
		q += fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;", pq.QuoteIdentifier(table))
	}
	q += "CREATE SCHEMA public;"

	_, err = tx.Exec(q)
	if err != nil {
		err = fmt.Errorf("failed to purge schema: %w", err)
		return
	}

	err = tx.Commit()
	if err != nil {
		err = fmt.Errorf("failed to commit schema purge transaction: %w", err)
		return
	}

	return
}

// Returns current schema version in the database
func getCurrentVersion(tx *sql.Tx) (version int, err error) {
	// Check if schema version table exists
	var exists bool
	err = sq.Select("EXISTS").
		Suffix("(?)", sq.
			Select("true").
			From("information_schema.tables").
			Where(sq.Eq{
				"table_schema": "meta",
				"table_name":   schemaVersionTableName,
			}),
		).
		PlaceholderFormat(sq.Dollar).
		RunWith(tx).
		QueryRow().
		Scan(&exists)
	if err != nil {
		err = fmt.Errorf("failed to check schema_version table existence: %w", err)
		return
	}

	if !exists {
		return
	}

	// Get current version from schema version table
	err = sq.
		Select("ID").
		From("meta." + schemaVersionTableName).
		OrderBy("ID DESC").
		Limit(1).
		RunWith(tx).
		QueryRow().
		Scan(&version)
	if err == sql.ErrNoRows {
		return
	}
	if err != nil {
		err = fmt.Errorf("failed to get current schema version: %w", err)
		return
	}

	return
}
