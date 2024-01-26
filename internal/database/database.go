package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"vivian.infra/internal/utils"
)

const (
	MaxIdleConns int = 10
	MaxOpenConns int = 20
)

type T interface {
	InitDatabase(context.Context) error
	FetchAccount(context.Context, string) (Account, error)
}

type ConfigSQL struct {
	Database *sql.DB
	Driver   string
	Source   string
}

func (config *ConfigSQL) InitDatabase(ctx context.Context, s *utils.VivianLogger) error {
	db, err := sql.Open(config.Driver, config.Source)
	if err != nil {
		return err
	}
	config.Database = db

	return config.Database.Ping()
}

type Account struct {
	ID       int
	Alias    string
	Email    string
	Password string
}

func FetchAccount(db *sql.DB, alias string) (Account, error) {
	var account Account

	_, err := db.Exec("USE user_schema")
	if err != nil {
		log.Fatal("Error selecting Database:", err)
	}

	stmt, err := db.Prepare("SELECT * FROM users WHERE alias = ?")
	if err != nil {
		return Account{}, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(alias).Scan(&account.ID, &account.Alias, &account.Email, &account.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return Account{}, fmt.Errorf("no account found for email: %w", err)
		}
		return Account{}, fmt.Errorf("failed to fetch account: %w", err)
	}
	return account, nil
}
