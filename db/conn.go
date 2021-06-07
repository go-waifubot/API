package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Config struct {
	User, Database, Password, Host string
}

// Init initialises the connetion with the db
func Init(c Config) (*Queries, error) {
	connStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s sslmode=disable", c.User, c.Database, c.Password, c.Host)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return New(db), nil
}
