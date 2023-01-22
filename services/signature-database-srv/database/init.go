package database

import (
	"embed"
	"github.com/openchainxyz/openchainxyz-monorepo/internal/database"
)

//go:embed migrations
var migrations embed.FS

type Database struct {
	db *database.Database
}

func New(host string, port int, dbname string, user string, pass string) (*Database, error) {
	db, err := database.New(host, port, dbname,
		database.WithAuth(user, pass),
		database.WithMigrations(&migrations),
	)
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func NewWithDatabase(db *database.Database) *Database {
	return &Database{db: db}
}
