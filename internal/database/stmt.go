package database

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"math/rand"
	"strings"
)

type Stmt struct {
	Conn *pgx.Conn
	name string
}

const (
	Uppercase    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Lowercase    = "abcdefghijklmnopqrstuvwxyz"
	Alphabetic   = Uppercase + Lowercase
	Numeric      = "0123456789"
	Alphanumeric = Alphabetic + Numeric
)

func randomString(length uint8, charsets ...string) string {
	charset := strings.Join(charsets, "")
	if charset == "" {
		charset = Alphanumeric
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Int63()%int64(len(charset))]
	}
	return string(b)
}

func (t *Tx) ExecBatch(apply func(*Stmt) error, query string) error {
	stmtName := randomString(32)

	_, err := t.Prepare(context.Background(), stmtName, query)
	if err != nil {
		return err
	}

	defer t.Conn().Deallocate(context.Background(), stmtName)

	return apply(&Stmt{Conn: t.Conn(), name: stmtName})
}

func (s *Stmt) Exec(context context.Context, args ...any) (pgconn.CommandTag, error) {
	return s.Conn.Exec(context, s.name, args...)
}

func (s *Stmt) ExecSimple(rows int64, args ...any) error {
	res, err := s.Conn.Exec(context.Background(), s.name, args...)
	if err != nil {
		return err
	}
	if err := CheckAffected(res, rows); err != nil {
		return err
	}
	return nil
}

func (s *Stmt) QuerySimple(apply func(r pgx.Rows) error, args ...any) error {
	rows, err := s.Conn.Query(context.Background(), s.name, args...)
	if err != nil {
		return err
	}

	defer rows.Close()

	return apply(rows)
}

func (s *Stmt) QuerySimpleOne(apply func(r pgx.Rows) error, args ...any) error {
	return s.QuerySimple(func(r pgx.Rows) error {
		if !r.Next() {
			return pgx.ErrNoRows
		}

		return apply(r)
	}, args...)
}

func (s *Stmt) QueryRowSimple(apply RowScanner, args ...any) error {
	return apply(s.Conn.QueryRow(context.Background(), s.name, args...))
}
