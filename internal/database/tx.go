package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

type Tx struct {
	pgx.Tx
}

func (d *Database) Tx(apply func(*Tx) (any, error)) (_ any, rerr error) {
	tx, err := d.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer func() {
		if rerr == nil {
			if err := tx.Commit(context.Background()); err != nil {
				rerr = fmt.Errorf("failed to commit tx: %w", err)
			}
		} else {
			if err := tx.Rollback(context.Background()); err != nil {
				rerr = fmt.Errorf("failed to rollback while handling failed tx: %w (%v)", rerr, err)
			}
		}
	}()

	ret, err := apply(&Tx{tx})
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (d *Database) ExecTx(apply func(*Tx) error) error {
	_, err := d.Tx(func(tx *Tx) (any, error) {
		return nil, apply(tx)
	})
	return err
}

func (t *Tx) ExecSimple(rows int64, query string, args ...any) error {
	res, err := t.Exec(context.Background(), query, args...)
	if err != nil {
		return err
	}
	if err := CheckAffected(res, rows); err != nil {
		return err
	}
	return nil
}

func (t *Tx) QuerySimple(apply RowsScanner, query string, args ...any) error {
	rows, err := t.Query(context.Background(), query, args...)
	if err != nil {
		return err
	}

	defer rows.Close()

	if err := apply(rows); err != nil {
		return err
	}

	return nil
}

func (t *Tx) QueryRowSimple(apply RowScanner, query string, args ...any) error {
	return apply(t.Tx.QueryRow(context.Background(), query, args...))
}
