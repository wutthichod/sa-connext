package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/wutthichod/sa-connext/shared/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewLogger() logger.Interface {
	// Create a standard Go logger that writes to stdout
	stdLogger := log.New(os.Stdout, "\r\n", log.LstdFlags)

	// Configure GORM logger
	newLogger := logger.New(
		stdLogger,
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, // highlight queries slower than 200ms
			LogLevel:                  logger.Info,            // log all queries, errors, warnings
			IgnoreRecordNotFoundError: true,                   // don't log ErrRecordNotFound
			Colorful:                  true,                   // enable colors for easier reading
		},
	)

	return newLogger
}
func InitDatabase(cfg config.Database) (*gorm.DB, error) {
	const (
		maxRetries = 10
		retryDelay = 2 * time.Second
		timeout    = 10 * time.Second
	)

	var db *gorm.DB
	var err error

	// Retry database connection with exponential backoff
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Attempting to connect to database (attempt %d/%d)...", attempt, maxRetries)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		// Use a channel to handle the connection attempt with timeout
		done := make(chan error, 1)
		go func() {
			var connErr error
			db, connErr = gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
				Logger: NewLogger(),
			})
			if connErr != nil {
				done <- connErr
				return
			}

			sqlDB, connErr := db.DB()
			if connErr != nil {
				done <- connErr
				return
			}

			// Set connection pool settings
			sqlDB.SetMaxIdleConns(10)
			sqlDB.SetMaxOpenConns(100)
			sqlDB.SetConnMaxLifetime(time.Hour)

			// Test connection with ping
			pingErr := sqlDB.Ping()
			done <- pingErr
		}()

		select {
		case err = <-done:
			cancel()
			if err == nil {
				log.Println("Successfully connected to database")
				return db, nil
			}
			log.Printf("Database connection attempt %d failed: %v", attempt, err)
		case <-ctx.Done():
			cancel()
			err = ctx.Err()
			log.Printf("Database connection attempt %d timed out", attempt)
		}

		if attempt < maxRetries {
			backoff := retryDelay * time.Duration(attempt)
			log.Printf("Retrying in %v...", backoff)
			time.Sleep(backoff)
		}
	}

	return nil, err
}
