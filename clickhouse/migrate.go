package clickhouse

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/insider/event-ingestion/config"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(cfg config.ClickHouseConfig) error {
	dsn := fmt.Sprintf(
		"clickhouse://%s:%d?database=%s&username=%s&password=%s&x-multi-statement=true",
		cfg.Host, cfg.Port, cfg.Database, cfg.Username, cfg.Password,
	)

	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create source driver: %w", err)
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", sourceDriver, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}
