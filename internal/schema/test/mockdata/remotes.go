package mockdata

import (
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
)

// RemotesVariant captures the inserted remote and its config.
type RemotesVariant struct {
	ID        int
	Name      string
	Password  sql.NullString
	Enabled   bool
	DeletedAt sql.NullTime
}

func (b *Builder) seedRemotes() error {
	passOptions := []sql.NullString{{}, {String: "bcrypt:remote", Valid: true}}
	enabledOptions := []bool{false, true}
	deletedOptions := []sql.NullTime{{}, {Time: b.now.Add(-1 * time.Hour), Valid: true}}

	var variants []RemotesVariant
	stmt := sq.Insert("remotes").Columns("name", "password_hash", "enabled", "deleted_at")

	for _, pass := range passOptions {
		for _, enabled := range enabledOptions {
			for _, del := range deletedOptions {
				name := fmt.Sprintf("remote-%d", len(variants)+1)
				pwd := pass
				if pass.Valid {
					pwd.String = fmt.Sprintf("%s-%d", pass.String, len(variants)+1)
				}

				stmt = stmt.Values(name, pwd, enabled, del)
				variants = append(variants, RemotesVariant{
					Name:      name,
					Password:  pwd,
					Enabled:   enabled,
					DeletedAt: del,
				})
			}
		}
	}

	ids, err := b.insertIDs(stmt)
	if err != nil {
		return err
	}

	for i, id := range ids {
		variants[i].ID = id
		b.f.Remotes[id] = variants[i]
	}

	return nil
}
