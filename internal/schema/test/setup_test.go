package test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/schema"
	"github.com/gerolf-vent/mailctl/internal/schema/test/mockdata"
)

var (
	testDB   *sql.DB
	fixtures *mockdata.Fixture
)

func TestMain(m *testing.M) {
	var err error
	testDB, err = db.Connect()
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer testDB.Close()

	latest, err := schema.GetLatestAvailableVersion()
	if err != nil {
		log.Fatalf("latest schema: %v", err)
	}

	if err := schema.Purge(testDB); err != nil {
		log.Fatalf("purge schema: %v", err)
	}
	if err := schema.Upgrade(testDB, latest); err != nil {
		log.Fatalf("upgrade schema: %v", err)
	}

	tx, err := testDB.Begin()
	if err != nil {
		log.Fatalf("begin seed tx: %v", err)
	}
	fixtures, err = mockdata.SeedAll(context.Background(), tx)
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("seed fixtures: %v", err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatalf("commit seed tx: %v", err)
	}

	if _, err := testDB.Exec(`CREATE EXTENSION IF NOT EXISTS pg_stat_statements`); err != nil {
		log.Printf("warning: create pg_stat_statements extension: %v", err)
	}
	if _, err := testDB.Exec(`SELECT pg_stat_statements_reset()`); err != nil {
		log.Printf("warning: reset pg_stat_statements: %v", err)
	}
	if _, err := testDB.Exec(`SET track_functions = 'all'`); err != nil {
		log.Printf("warning: set track_functions: %v", err)
	}

	code := m.Run()

	os.Exit(code)
}
