package database

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"math/big"
)

func Nullable(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func CheckAffected(res pgconn.CommandTag, want int64) error {
	rows := res.RowsAffected()
	if rows != want {
		return fmt.Errorf("expected to modify %d rows, but modified %d instead", want, rows)
	}
	return nil
}

type SQLBigInt big.Int

func (i *SQLBigInt) Value() (driver.Value, error) {
	return (*big.Int)(i).String(), nil
}
