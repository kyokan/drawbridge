package db

import (
	"database/sql"
	"strings"
	"github.com/go-errors/errors"
	_ "github.com/lib/pq"
)

type DB struct {
	Outputs  Outputs
	dbUrl    string
	db       *sql.DB
}

func NewDB(dbUrl string) (*DB, error) {
	parts := strings.Split(dbUrl, "://")

	if len(parts) != 2 {
		return nil, errors.New("mal-formed database URL")
	}

	if parts[0] != "postgres" {
		return nil, errors.New("only postgres databases are supported right now")
	}

	db, err := sql.Open("postgres", dbUrl)

	if err != nil {
		return nil, err
	}

	return &DB{
		Outputs: &PostgresOutputs{
			db: db,
		},
		dbUrl: dbUrl,
		db:    db,
	}, nil
}

func (db *DB) Connect() error {
	return db.db.Ping()
}

func (db *DB) Close() error {
	return db.db.Close()
}
