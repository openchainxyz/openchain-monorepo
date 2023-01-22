package database

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"net/url"
	"reflect"
)

type Database struct {
	*pgxpool.Pool
}

type Scannable interface {
	MakeScanner() RowScanner
}

type RowScanner func(row pgx.Row) error
type RowsScanner func(rows pgx.Rows) error

func ScanInto(outs ...any) RowScanner {
	type unmarshalContext struct {
		idx  int
		data []byte
	}
	return func(row pgx.Row) error {
		var remappedOuts []any

		var toUnmarshal []*unmarshalContext

		for idx, out := range outs {
			typ := reflect.TypeOf(out)
			if typ.Kind() != reflect.Pointer {
				return fmt.Errorf("expected pointer")
			}

			elem := typ.Elem()
			if elem.Kind() == reflect.Pointer {
				elem = elem.Elem()
			}

			if elem.Kind() == reflect.Struct || elem.Kind() == reflect.Map {
				var isJson bool
				if elem.Kind() == reflect.Map {
					isJson = true
				} else if elem.Kind() == reflect.Struct {
					for i := 0; i < elem.NumField(); i++ {
						if _, ok := elem.Field(i).Tag.Lookup("json"); ok {
							isJson = true
						}
					}
				}

				if isJson {
					// we want to unmarshal the jsonb
					ctx := &unmarshalContext{
						idx: idx,
					}
					remappedOuts = append(remappedOuts, &ctx.data)
					toUnmarshal = append(toUnmarshal, ctx)
				} else {
					remappedOuts = append(remappedOuts, out)
				}
			} else {
				remappedOuts = append(remappedOuts, out)
			}
		}
		if err := row.Scan(remappedOuts...); err != nil {
			return err
		}
		for _, unmarshal := range toUnmarshal {
			if err := json.Unmarshal(unmarshal.data, outs[unmarshal.idx]); err != nil {
				return fmt.Errorf("failed to unmarshal while scanning: %w", err)
			}
		}
		return nil
	}
}

type Querier interface {
	QuerySimple(apply RowsScanner, query string, args ...any) error
	QueryRowSimple(apply RowScanner, query string, args ...any) error
}

func NewDatabase(host string, port int, dbname string, username string, password string) (*Database, error) {
	return New(host, port, dbname, WithAuth(username, password))
}

func (d *Database) QuerySimple(apply RowsScanner, query string, args ...any) error {
	rows, err := d.Query(context.Background(), query, args...)
	if err != nil {
		return err
	}

	defer rows.Close()

	return apply(rows)
}

func (d *Database) QuerySimpleOne(apply func(r pgx.Rows) error, query string, args ...any) error {
	return d.QuerySimple(func(r pgx.Rows) error {
		if !r.Next() {
			return pgx.ErrNoRows
		}

		return apply(r)
	}, query, args...)
}

func (d *Database) QueryRowSimple(apply RowScanner, query string, args ...any) error {
	return apply(d.QueryRow(context.Background(), query, args...))
}

type databaseBuilder struct {
	host       string
	port       int
	database   string
	username   string
	password   string
	migrations *embed.FS
}

type Option func(db *databaseBuilder)

func WithAuth(username string, password string) Option {
	return func(db *databaseBuilder) {
		db.username = username
		db.password = password
	}
}

func WithMigrations(migrations *embed.FS) Option {
	return func(db *databaseBuilder) {
		db.migrations = migrations
	}
}

func New(host string, port int, database string, apply ...Option) (*Database, error) {
	builder := &databaseBuilder{
		host:     host,
		port:     port,
		database: database,
	}
	for _, fn := range apply {
		fn(builder)
	}

	pgurl := url.URL{}
	pgurl.Scheme = "postgres"
	if builder.username != "" {
		if builder.password != "" {
			pgurl.User = url.UserPassword(builder.username, builder.password)
		} else {
			pgurl.User = url.User(builder.username)
		}
	}
	pgurl.Host = fmt.Sprintf("%s:%d", builder.host, builder.port)
	pgurl.Path = builder.database

	pool, err := pgxpool.New(context.Background(), pgurl.String())
	if err != nil {
		return nil, err
	}

	if builder.migrations != nil {
		sqlDb := stdlib.OpenDB(*pool.Config().ConnConfig)

		driver, err := postgres.WithInstance(sqlDb, &postgres.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to create postgres: %w", err)
		}

		data, err := iofs.New(builder.migrations, "migrations")
		if err != nil {
			return nil, fmt.Errorf("failed to create bindata: %w", err)
		}

		m, err := migrate.NewWithInstance("iofs", data, "postgres", driver)
		if err != nil {
			return nil, fmt.Errorf("failed to create migrate: %w", err)
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return nil, fmt.Errorf("failed to apply migrations: %w", err)
		}

		if err := sqlDb.Close(); err != nil {
			return nil, fmt.Errorf("failed to close db: %w", err)
		}
	}

	return &Database{
		Pool: pool,
	}, nil
}
