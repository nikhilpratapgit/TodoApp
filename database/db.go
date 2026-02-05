package database

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	Todo *sqlx.DB
)

const (
	SSLModeDisable SSLMode = "disable"
)

type SSLMode string

func ConnectandMigrate(host, port, databaseName, user, password string, sslMode SSLMode) error {
	connectionStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", host, port, databaseName, user, password, sslMode)

	DB, err := sqlx.Open("postgres", connectionStr)
	if err != nil {
		return err
	}
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("database ping failed %w", err)
	}
	fmt.Println("Database connected successfully")
	Todo = DB
	return migrateUp(DB)
}

func migrateUp(db *sqlx.DB) error {
	fmt.Println("Starting database migrations...")
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://database/migrations",
		"postgres", driver)

	if err != nil {
		return err
	}
	//_ = m.Force(1)

	//if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
	//	return err
	//}
	//return nil
	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("Migrations applied successfully")
	return nil
}

//	func ShutdownDatabase() error {
//		return Todo.Close()
//	}
//func Tx(fn func(tx *sqlx.Tx) error) error {
//	tx, err := Todo.Beginx()
//	if err != nil {
//		return fmt.Errorf("failed to start a transaction: %v", err)
//	}
//	defer func() {
//		if err != nil {
//			if rollBackErr := tx.Rollback(); rollBackErr != nil {
//				fmt.Println("failed to rollback tx : %s", rollBackErr)
//			}
//			return
//		}
//		if commitErr := tx.Commit(); commitErr != nil {
//			fmt.Println("failed to commit: %s", commitErr)
//		}
//	}()
//	err = fn(tx)
//	return err
//}
