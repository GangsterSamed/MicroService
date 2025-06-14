package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log/slog"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	"time"
)

func initDatabase(cfg *config.UserConfig, logger *slog.Logger) (*sql.DB, error) {
	var db *sql.DB
	var err error
	dbDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", dbDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}

		// Проверка подключения
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			logger.Warn("Failed to ping database, retrying...", "attempt", i+1, "error", err)
			time.Sleep(2 * time.Second)
			continue
		}
	}

	// Применяем миграции
	if err := applyMigrations(db, logger); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	logger.Info("Database connection established")
	return db, nil
}

func applyMigrations(db *sql.DB, logger *slog.Logger) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	logger.Info("Applying database migrations...")
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	logger.Info("Migrations applied successfully")
	return nil
}
