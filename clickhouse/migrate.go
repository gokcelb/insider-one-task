package clickhouse

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/insider/event-ingestion/config"
)

func RunMigrations(cfg config.ClickHouseConfig, migrationsPath string) error {
	dsn := fmt.Sprintf(
		"clickhouse://%s:%d?database=%s&username=%s&password=%s&x-multi-statement=true",
		cfg.Host, cfg.Port, cfg.Database, cfg.Username, cfg.Password,
	)

	m, err := migrate.New("file://"+migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}
