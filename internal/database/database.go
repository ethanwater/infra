package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"vivian.infra/internal/utils"
)

const (
	MaxIdleConns int = 10
	MaxOpenConns int = 20
)

type T interface {
	Init(context.Context) error
	PingDBConnection(context.Context) error
	FetchAccount(context.Context, string) (Account, error)
	AddAccount(context.Context, Account) error
}

type ConfigSQL struct {
	Database *sql.DB
	Driver   string
	Source   string
}

func (config *ConfigSQL) InitDatabase(ctx context.Context, s *utils.VivianLogger) error {
	Database, _ := sql.Open(config.Driver, config.Source)
	config.Database = Database
	config.Database.SetMaxIdleConns(MaxIdleConns)
	config.Database.SetMaxOpenConns(MaxOpenConns)

	ping := config.Database.Ping()
	return ping
}

type Account struct {
	ID       int
	Email    string
	Password string
}

func FetchAccount(database *sql.DB, email string) (Account, error) {
	var mux sync.Mutex
	var account Account

	mux.Lock()
	defer mux.Unlock()

	_, err := database.Exec("USE user_schema")
	if err != nil {
		log.Fatal("Error selecting Database:", err)
	}

	stmt, err := database.Prepare("SELECT * FROM users WHERE email = ?")
	if err != nil {
		return Account{}, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(email).Scan(&account.ID, &account.Email, &account.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return Account{}, fmt.Errorf("no account found for email: %w", err)
		}
		return Account{}, fmt.Errorf("failed to fetch account: %w", err)
	}
	return account, nil
}
