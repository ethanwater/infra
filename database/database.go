package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"vivian.infra/models"
	"vivian.infra/utils"
)

const (
	MaxIdleConns int = 10
	MaxOpenConns int = 20
)

type T interface {
	InitDatabase(context.Context) error
	FetchAccount(context.Context, string) (models.Account, error)
}

type ConfigSQL struct {
	Database *sql.DB
	Driver   string
	Source   string
}

var VivianServerLogger *utils.VivianLogger

func (config *ConfigSQL) InitDatabase(ctx context.Context, s *utils.VivianLogger) error {
	db, err := sql.Open(config.Driver, config.Source)
	if err != nil {
		return err
	}
	config.Database, VivianServerLogger = db, s

	return config.Database.Ping()
}

func FetchAccount(db *sql.DB, alias string) (models.Account, error) {
	var account models.Account

	_, err := db.Exec("USE user_schema")
	if err != nil {
		VivianServerLogger.LogFatal("error searching database")
	}

	stmt, err := db.Prepare("SELECT * FROM users WHERE alias = ?")
	if err != nil {
		return models.Account{}, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(alias).Scan(&account.ID, &account.Alias, &account.Email, &account.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Account{}, fmt.Errorf("no account found for email: %w", err)
		}
		return models.Account{}, fmt.Errorf("failed to fetch account: %w", err)
	}
	return account, nil
}
