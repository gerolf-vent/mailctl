package schema

import (
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/lib/pq"
)

type User struct {
	Name     string
	Password string
}

// EnsureUser creates or updates a single database user and assigns the
// minimum required privileges for the given user type.
func EnsureUser(db *sql.DB, dbName, userType string, user User) error {
	if user.Name == "" {
		return fmt.Errorf("username is required")
	}
	if user.Password == "" {
		return fmt.Errorf("password is required")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := ensureUser(tx, dbName, user.Name, user.Password); err != nil {
		return err
	}

	var grantFn func(*sql.Tx, string) error
	switch userType {
	case "manager":
		grantFn = ensureManagerGrants
	case "postfix":
		grantFn = ensureIntegrationSchemaGrants("postfix")
	case "dovecot":
		grantFn = ensureIntegrationSchemaGrants("dovecot")
	case "stalwart":
		grantFn = ensureIntegrationSchemaGrants("stalwart")
	default:
		return fmt.Errorf("invalid user type %q", userType)
	}

	if err := grantFn(tx, user.Name); err != nil {
		return fmt.Errorf("failed to create grants for user %s: %w", user.Name, err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func DropUser(db *sql.DB, dbName, dbUserName, userName string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	exists, err := checkUserExists(tx, userName)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check existence of user %s: %w", userName, err)
	}

	if !exists || errors.Is(err, sql.ErrNoRows) {
		return nil // User does not exist, nothing to do
	}

	var isManager bool
	err = sq.
		Select("granted").
		Suffix("FROM has_schema_privilege(?, ?, ?) AS granted", userName, "public", "USAGE").
		PlaceholderFormat(sq.Dollar).
		RunWith(tx).
		QueryRow().
		Scan(&isManager)
	if err != nil {
		return fmt.Errorf("failed to check if user %s is a manager: %w", userName, err)
	}

	var reassignUserName string
	if isManager {
		if dbUserName == userName {
			return fmt.Errorf("cannot drop user %s because it is the connected database user", userName)
		}
		reassignUserName = dbUserName
	}

	if err := dropUser(tx, dbName, userName, reassignUserName); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func checkUserExists(tx *sql.Tx, userName string) (bool, error) {
	var exists bool
	err := sq.
		Select("true").
		From("pg_catalog.pg_user").
		Where(sq.Eq{
			"usename": userName,
		}).
		PlaceholderFormat(sq.Dollar).
		RunWith(tx).
		QueryRow().
		Scan(&exists)
	return exists, err
}

func ensureUser(tx *sql.Tx, dbName, userName, password string) error {
	exists, err := checkUserExists(tx, userName)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check existence of user %s: %w", userName, err)
	}

	var q string
	if !exists {
		// Create user
		q = fmt.Sprintf("CREATE USER %s", pq.QuoteIdentifier(userName))
		if password != "" {
			q += fmt.Sprintf(" WITH PASSWORD %s", pq.QuoteLiteral(password))
		}
	} else {
		// Update password if needed
		if password != "" {
			q = fmt.Sprintf("ALTER USER %s WITH PASSWORD %s",
				pq.QuoteIdentifier(userName),
				pq.QuoteLiteral(password),
			)
		} else {
			q = fmt.Sprintf("ALTER USER %s WITH PASSWORD NULL",
				pq.QuoteIdentifier(userName),
			)
		}
	}

	err = db.Exec(tx, sq.Expr(q), 1)
	if err != nil {
		return fmt.Errorf("failed to ensure user %s: %w", userName, err)
	}

	q = fmt.Sprintf("REVOKE ALL ON DATABASE %s FROM %s; GRANT CONNECT ON DATABASE %s TO %s;",
		pq.QuoteIdentifier(dbName),
		pq.QuoteIdentifier(userName),
		pq.QuoteIdentifier(dbName),
		pq.QuoteIdentifier(userName),
	)
	if _, err := tx.Exec(q); err != nil {
		return fmt.Errorf("failed to ensure basic privileges for %s: %w", userName, err)
	}

	return nil
}

func dropUser(tx *sql.Tx, dbName, userName, reassignUserName string) error {
	// Reassign owned objects if needed
	if reassignUserName != "" {
		q := fmt.Sprintf("REASSIGN OWNED BY %s TO %s;",
			pq.QuoteIdentifier(userName),
			pq.QuoteIdentifier(reassignUserName),
		)
		if _, err := tx.Exec(q); err != nil {
			return fmt.Errorf("failed to reassign owned objects for %s: %w", userName, err)
		}
	}

	// Drop owned objects
	q := fmt.Sprintf("DROP OWNED BY %s;", pq.QuoteIdentifier(userName))
	if _, err := tx.Exec(q); err != nil {
		return fmt.Errorf("failed to drop owned objects for %s: %w", userName, err)
	}

	// Drop user
	q = fmt.Sprintf("DROP USER IF EXISTS %s", pq.QuoteIdentifier(userName))
	_, err := tx.Exec(q)
	if err != nil {
		return fmt.Errorf("failed to drop user %s: %w", userName, err)
	}

	return nil
}

func ensureManagerGrants(tx *sql.Tx, userName string) error {
	// Allow usage on shared schema
	q := fmt.Sprintf("GRANT USAGE ON SCHEMA meta TO %s", pq.QuoteIdentifier(userName))
	if _, err := tx.Exec(q); err != nil {
		return err
	}

	// Allow read access on tables in public meta
	q = fmt.Sprintf("GRANT SELECT ON ALL TABLES IN SCHEMA meta TO %s", pq.QuoteIdentifier(userName))
	if _, err := tx.Exec(q); err != nil {
		return err
	}

	// Allow usage on shared schema
	q = fmt.Sprintf("GRANT USAGE ON SCHEMA shared TO %s", pq.QuoteIdentifier(userName))
	if _, err := tx.Exec(q); err != nil {
		return err
	}

	// Allow usage on sequences in shared schema
	q = fmt.Sprintf("GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA shared TO %s", pq.QuoteIdentifier(userName))
	if _, err := tx.Exec(q); err != nil {
		return err
	}

	// Allow usage on public schema
	q = fmt.Sprintf("GRANT USAGE ON SCHEMA public TO %s", pq.QuoteIdentifier(userName))
	if _, err := tx.Exec(q); err != nil {
		return err
	}

	// Allow usage on sequences in public schema
	q = fmt.Sprintf("GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO %s", pq.QuoteIdentifier(userName))
	if _, err := tx.Exec(q); err != nil {
		return err
	}

	// Allow full access on tables in public schema
	q = fmt.Sprintf("GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO %s", pq.QuoteIdentifier(userName))
	if _, err := tx.Exec(q); err != nil {
		return err
	}

	// Allow usage on postfix schema
	if err := ensureIntegrationSchemaGrants("postfix")(tx, userName); err != nil {
		return err
	}

	// Allow usage on dovecot schema
	if err := ensureIntegrationSchemaGrants("dovecot")(tx, userName); err != nil {
		return err
	}

	// Allow usage on stalwart schema
	if err := ensureIntegrationSchemaGrants("stalwart")(tx, userName); err != nil {
		return err
	}

	return nil
}

func ensureIntegrationSchemaGrants(schema string) func(*sql.Tx, string) error {
	return func(tx *sql.Tx, username string) error {
		// Allow usage on schema
		q := fmt.Sprintf("GRANT USAGE ON SCHEMA %s TO %s", pq.QuoteIdentifier(schema), pq.QuoteIdentifier(username))
		if _, err := tx.Exec(q); err != nil {
			return err
		}

		// Allow execute functions in schema
		q = fmt.Sprintf("GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA %s TO %s", pq.QuoteIdentifier(schema), pq.QuoteIdentifier(username))
		if _, err := tx.Exec(q); err != nil {
			return err
		}

		return nil
	}
}
