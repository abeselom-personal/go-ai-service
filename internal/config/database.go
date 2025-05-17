// internal/config/database.go
package config

import (
	"fmt"
	"time"

	"github.com/abeselom-personal/go-ai-service/pkg/logger"
	"github.com/caarlos0/env/v10"
	"github.com/go-playground/validator/v10"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseConfig struct {
	Host     string `env:"DB_HOST" envDefault:"localhost" validate:"required"`
	Port     string `env:"DB_PORT" envDefault:"5432" validate:"numeric"`
	User     string `env:"DB_USER" envDefault:"postgres" validate:"required"`
	Password string `env:"DB_PASSWORD" envDefault:"postgres" validate:"required"`
	Name     string `env:"DB_NAME" envDefault:"postgres" validate:"required"`
	SSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`
}

func (c *DatabaseConfig) GetConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}

func InitDBWithLogger(log *logger.Logger) (*gorm.DB, error) {
	var cfg DatabaseConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("env parse error: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	var db *gorm.DB
	err := retry(5, 2*time.Second, func() error {
		var err error
		db, err = gorm.Open(postgres.Open(cfg.GetConnectionString()), &gorm.Config{
			PrepareStmt: true,
			Logger:      logger.NewGormLogger(log),
		})
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("connection pool error: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Info("database connection established")
	return db, nil
}

func Close(db *gorm.DB) {
	if db == nil {
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		return
	}

	if err := sqlDB.Close(); err != nil {
		// Log close error if needed
	}
}

func retry(attempts int, sleep time.Duration, f func() error) error {
	for i := 0; ; i++ {
		err := f()
		if err == nil {
			return nil
		}

		if i >= (attempts - 1) {
			return fmt.Errorf("after %d attempts: %w", attempts, err)
		}

		time.Sleep(sleep)
	}
}
