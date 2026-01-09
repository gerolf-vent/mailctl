package mockdata

import (
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
)

// TransportsVariant captures the inserted transport and its config.
type TransportsVariant struct {
	ID        int
	Method    string
	Host      string
	Port      sql.NullInt32
	MxLookup  bool
	DeletedAt sql.NullTime
}

func (b *Builder) seedTransports() error {
	methods := []string{"smtp", "lmtp"}
	portOptions := []sql.NullInt32{{}, {Int32: 25, Valid: true}}
	mxOptions := []bool{false, true}
	deletedOptions := []sql.NullTime{{}, {Time: b.now.Add(-1 * time.Hour), Valid: true}}

	var variants []TransportsVariant
	stmt := sq.Insert("transports").Columns("name", "method", "host", "port", "mx_lookup", "deleted_at")

	for _, method := range methods {
		for _, port := range portOptions {
			for _, mx := range mxOptions {
				for _, del := range deletedOptions {
					host := fmt.Sprintf("%s-%d.example", method, len(variants)+1)
					name := fmt.Sprintf("transport-%s-%d", method, len(variants)+1)

					stmt = stmt.Values(name, method, host, port, mx, del)
					variants = append(variants, TransportsVariant{
						Method:    method,
						Host:      host,
						Port:      port,
						MxLookup:  mx,
						DeletedAt: del,
					})
				}
			}
		}
	}

	ids, err := b.insertIDs(stmt)
	if err != nil {
		return err
	}

	for i, id := range ids {
		variants[i].ID = id
		b.f.Transports[id] = variants[i]
	}

	return nil
}
