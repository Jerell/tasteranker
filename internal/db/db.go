package db

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
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
        connStr := fmt.Sprintf(
            "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
            cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
        )
        db, err = sql.Open("postgres", connStr)
    }

    if err != nil {
        return nil, fmt.Errorf("error opening database: %v", err)
    }

    fmt.Printf("Database connection opened, attempting ping\n") 

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("error connecting to database: %v", err)
    }

    fmt.Printf("Database ping successful\n") 
    return db, nil
}

func RunMigrations(db *sql.DB) error {
    fmt.Printf("Starting migrations...\n") 

    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        fmt.Printf("Driver creation error: %v\n", err) 
        return fmt.Errorf("could not create database driver: %w", err)
    }

    fmt.Printf("Driver created successfully, attempting to create migrate instance\n") 

    m, err := migrate.NewWithDatabaseInstance(
        "file:///app/migrations",
        "postgres", 
        driver,
    )
    if err != nil {
        fmt.Printf("Migrate instance creation error: %v\n", err) 
        return fmt.Errorf("could not create migrate instance: %w", err)
    }
    defer m.Close()

    fmt.Printf("Running migrations...\n") 

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        fmt.Printf("Migration error: %v\n", err) 
        return fmt.Errorf("failed to run migrations: %w", err)
    }
    
    fmt.Printf("Migrations completed successfully\n") 
    return nil
}
