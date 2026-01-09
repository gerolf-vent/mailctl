package test

import "database/sql"

type userdbRow struct {
	QuotaStorageSize sql.NullString
}

func (r *userdbRow) Equals(other userdbRow) bool {
	return r.QuotaStorageSize.Valid == other.QuotaStorageSize.Valid &&
		(!r.QuotaStorageSize.Valid || r.QuotaStorageSize.String == other.QuotaStorageSize.String)
}
