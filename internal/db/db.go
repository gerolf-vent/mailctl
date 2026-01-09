package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type Config struct {
	Host      string
	Port      string
	User      string
	DBName    string
	SSLMode   string
	Password  string
	TLSCert   string
	TLSKey    string
	TLSCACert string
}

func (c Config) DSN() string {
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		c.Host,
		c.Port,
		c.User,
		c.DBName,
		c.SSLMode,
		c.Password,
	)

	// Add CA certificate path if provided and SSL mode requires verification
	if c.TLSCACert != "" && (c.SSLMode == "verify-ca" || c.SSLMode == "verify-full") {
		dsn += fmt.Sprintf(" sslrootcert=%s", c.TLSCACert)
	}

	if c.TLSCert != "" && c.TLSKey != "" {
		dsn += fmt.Sprintf(" sslcert=%s sslkey=%s", c.TLSCert, c.TLSKey)
	}

	return dsn
}

func GetConfig() Config {
	return Config{
		Host:      getEnv("DB_HOST", "localhost"),
		Port:      getEnv("DB_PORT", "5432"),
		User:      getEnv("DB_USER", "mail"),
		DBName:    getEnv("DB_NAME", "mail"),
		SSLMode:   getEnv("DB_SSLMODE", "disable"),
		Password:  os.Getenv("DB_PASSWORD"),
		TLSCert:   os.Getenv("DB_TLSCERT"),
		TLSKey:    os.Getenv("DB_TLSKEY"),
		TLSCACert: os.Getenv("DB_TLSCACERT"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Connect() (*sql.DB, error) {
	config := GetConfig()

	db, err := sql.Open("postgres", config.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
