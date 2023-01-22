package database

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
	"github.com/openchainxyz/openchainxyz-monorepo/internal/database"
	"github.com/openchainxyz/openchainxyz-monorepo/services/signature-database-srv/client"
	"io"
	"regexp"
	"strings"
)

var signatureLens = map[client.SignatureType]int{
	client.SignatureTypeFunction: 4,
	client.SignatureTypeEvent:    32,
}

var saveSignatureQueries = map[client.SignatureType]string{
	client.SignatureTypeFunction: `INSERT INTO fourbyte (name, hash) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
	client.SignatureTypeEvent:    `INSERT INTO thirtytwobyte (name, hash) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
}

var loadSignatureQueries = map[client.SignatureType]string{
	client.SignatureTypeFunction: `SELECT name, hash FROM fourbyte where hash = ANY($1)`,
	client.SignatureTypeEvent:    `SELECT name, hash FROM thirtytwobyte where hash = ANY($1)`,
}

var querySignatureQueries = map[client.SignatureType]string{
	client.SignatureTypeFunction: `SELECT name, hash FROM fourbyte WHERE name LIKE $1 LIMIT $2`,
	client.SignatureTypeEvent:    `SELECT name, hash FROM thirtytwobyte WHERE name LIKE $1 LIMIT $2`,
}

var countSignatureQueries = map[client.SignatureType]string{
	client.SignatureTypeFunction: `SELECT COUNT(*) FROM fourbyte`,
	client.SignatureTypeEvent:    `SELECT COUNT(*) FROM thirtytwobyte`,
}

func (d *Database) SaveSignatures(typ client.SignatureType, names []string) (*client.ImportResponseDetails, error) {
	result := client.NewImportResponseDetails()

	if err := d.db.ExecTx(func(tx *database.Tx) error {
		return tx.ExecBatch(func(stmt *database.Stmt) error {
			for _, name := range names {
				sig := crypto.Keccak256([]byte(name))[:signatureLens[typ]]
				hexSig := "0x" + hex.EncodeToString(sig)

				res, err := stmt.Exec(context.Background(), name, sig)
				if err != nil {
					return fmt.Errorf("failed to insert: %w", err)
				}

				rows := res.RowsAffected()
				if rows > 0 {
					result.Imported[name] = hexSig
				} else {
					result.Duplicated[name] = hexSig
				}
			}

			return nil
		}, saveSignatureQueries[typ])
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (d *Database) ExportData(w io.Writer) error {
	if err := d.db.QuerySimple(func(r pgx.Rows) error {
		var (
			name string
			hash []byte
		)
		for r.Next() {
			if err := r.Scan(&name, &hash); err != nil {
				return err
			}
			if _, err := io.WriteString(w, fmt.Sprintf("0x%x,%s\n", hash, name)); err != nil {
				return err
			}
		}
		return nil
	}, `SELECT * FROM fourbyte ORDER BY hash`); err != nil {
		return err
	}
	if err := d.db.QuerySimple(func(r pgx.Rows) error {
		var (
			name string
			hash []byte
		)
		for r.Next() {
			if err := r.Scan(&name, &hash); err != nil {
				return err
			}
			if _, err := io.WriteString(w, fmt.Sprintf("0x%x,%s\n", hash, name)); err != nil {
				return err
			}
		}
		return nil
	}, `SELECT * FROM thirtytwobyte ORDER BY hash`); err != nil {
		return err
	}

	return nil
}

var isValidQuery = regexp.MustCompile(`^[a-zA-Z0-9$_()\[\],*?]+$`).MatchString

func (d *Database) sanitizeQuery(name string) (string, error) {
	if !isValidQuery(name) {
		return "", fmt.Errorf("invalid query: %s", name)
	}

	name = strings.ReplaceAll(name, "_", "\\_")
	name = strings.ReplaceAll(name, "*", "%")
	name = strings.ReplaceAll(name, "?", "_")
	return name, nil
}

func (d *Database) QuerySignatures(query string) (map[client.SignatureType]map[string][]*client.SignatureData, error) {
	sanitizedQuery, err := d.sanitizeQuery(query)
	if err != nil {
		return nil, err
	}

	result := make(map[client.SignatureType]map[string][]*client.SignatureData)

	if err := d.db.ExecTx(func(tx *database.Tx) error {
		for _, typ := range client.SignatureTypes() {
			result[typ] = make(map[string][]*client.SignatureData)

			if err := tx.QuerySimple(func(r pgx.Rows) error {
				for r.Next() {
					var (
						name string
						hash []byte
					)
					if err := r.Scan(&name, &hash); err != nil {
						return fmt.Errorf("failed to scan: %w", err)
					}

					sel := "0x" + hex.EncodeToString(hash)
					result[typ][sel] = append(result[typ][sel], &client.SignatureData{
						Name: name,
					})
				}

				return nil
			}, querySignatureQueries[typ], sanitizedQuery, 100); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (d *Database) LoadSignatures(typ client.SignatureType, sels []string) (map[string][]*client.SignatureData, error) {
	result := make(map[string][]*client.SignatureData)

	var arr [][]byte
	for _, sel := range sels {
		b, err := hexutil.Decode(sel)
		if err != nil {
			return nil, err
		}
		arr = append(arr, b)
	}

	if err := d.db.QuerySimple(func(rows pgx.Rows) error {
		for rows.Next() {
			var (
				name string
				sel  []byte
			)
			if err := rows.Scan(&name, &sel); err != nil {
				return fmt.Errorf("failed to scan: %w", err)
			}

			h := hexutil.Encode(sel)

			result[h] = append(result[h], &client.SignatureData{
				Name: name,
			})
		}
		return nil
	}, loadSignatureQueries[typ], pq.ByteaArray(arr)); err != nil {
		return nil, err
	}

	for _, v := range sels {
		if _, ok := result[v]; !ok {
			result[v] = []*client.SignatureData{}
		}
	}

	return result, nil
}

func (d *Database) CountSignatures(typ client.SignatureType) (int, error) {
	var count int
	if err := d.db.QuerySimpleOne(func(r pgx.Rows) error {
		return r.Scan(&count)
	}, countSignatureQueries[typ]); err != nil {
		return 0, err
	}
	return count, nil
}
