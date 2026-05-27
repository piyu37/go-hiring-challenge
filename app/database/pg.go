package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mytheresa/go-hiring-challenge/app/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(ctx context.Context, cfg config.Database) (*gorm.DB, func() error, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)

	var (
		db  *gorm.DB
		err error
	)

	for attempt := 1; attempt <= cfg.ConnectRetries; attempt++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			sqlDB, dbErr := db.DB()
			if dbErr != nil {
				return nil, nil, fmt.Errorf("get sql db: %w", dbErr)
			}

			if pingErr := ping(ctx, sqlDB); pingErr == nil {
				configurePool(sqlDB, cfg)
				return db, sqlDB.Close, nil
			}

			err = ping(ctx, sqlDB)
			_ = sqlDB.Close()
		}

		if attempt == cfg.ConnectRetries {
			break
		}

		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(cfg.ConnectRetryDelay):
		}
	}

	return nil, nil, fmt.Errorf("connect database after %d attempts: %w", cfg.ConnectRetries, err)
}

func Ping(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return ping(ctx, sqlDB)
}

func ping(ctx context.Context, sqlDB *sql.DB) error {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return sqlDB.PingContext(pingCtx)
}

func configurePool(sqlDB *sql.DB, cfg config.Database) {
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
}
