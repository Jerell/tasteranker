package db

import (
	"database/sql"
	"fmt"
	"os"
	_ "github.com/lib/pq"
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewConfig() *Config {
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		return &Config{
			Host: dbURL, // Store the full URL in Host field
		}
	}

	return &Config{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "5432"),
		User:     getEnvOrDefault("DB_USER", "postgres"),
		Password: getEnvOrDefault("DB_PASSWORD", "postgres"),
		DBName:   getEnvOrDefault("DB_NAME", "tasteranker-dev"),
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func NewConnection(cfg *Config) (*sql.DB, error) {
	var db *sql.DB
	var err error

	if cfg.Host != "localhost" && cfg.Port == "" {
		db, err = sql.Open("postgres", cfg.Host)
	} else {
		// Local development connection string
		connStr := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
		)
		db, err = sql.Open("postgres", connStr)
	}

	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	return db, nil
}

func RunMigrations(db *sql.DB) error {
    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        return fmt.Errorf("could not create database driver: %w", err)
    }

    m, err := migrate.NewWithDatabaseInstance(
        "file:///app/migrations",  // migration files location
        "postgres", 
        driver,
    )
    if err != nil {
        return fmt.Errorf("could not create migrate instance: %w", err)
    }
    defer m.Close()

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("failed to run migrations: %w", err)
    }
    
    return nil
}
