package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kirillmashkov/gophermart/internal/config"
	"go.uber.org/zap"
)

const migrateDir = "migrations"

type Database struct {
	cfg    *config.ServerConfig
	conn   *pgx.Conn
	Dbpool *pgxpool.Pool
	logger *zap.Logger
}

func NewDatabase(config *config.ServerConfig, logger *zap.Logger) *Database {
	return &Database{cfg: config, logger: logger}
}

func (d *Database) Open() error {
	var err error
	d.Dbpool, _ = pgxpool.New(context.Background(), d.cfg.Connection)
	return err
}

func (d *Database) Close() {
	d.Dbpool.Close()
}

func (d *Database) Migrate() error {
	m, err := migrate.New("file://"+migrateDir, d.cfg.Connection)
	if err != nil {
		fmt.Printf("%s", migrateDir)
		d.logger.Error("Can't initialize migrations", zap.Error(err))
		return err
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			d.logger.Info("No migrations need")
			return nil
		}
		d.logger.Error("Something went wrong while migrations", zap.Error(err))
		return err
	}
	return nil
}
