package database

import (
	"log"
	"os"
	"time"

	"github.com/wutthichod/sa-connext/services/user-service/pkg/config"
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
	// newLogger := logger.New(
	// 	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	// 	logger.Config{
	// 		SlowThreshold: time.Second, // Slow SQL threshold
	// 		LogLevel:      logger.Info, // Log level
	// 		Colorful:      true,        // Enable color
	// 	},
	// )

	// db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
	// 	Logger: newLogger,
	// })
	// if err != nil {
	// 	return nil, err
	// }
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger: NewLogger(),
	})
	if err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
