package test

import "database/sql"

type passdbRow struct {
	Password sql.NullString
	NoLogin  sql.NullBool
	Reason   sql.NullString
}

func (r *passdbRow) Equals(other passdbRow) bool {
	return r.Password.Valid == other.Password.Valid &&
		(!r.Password.Valid || r.Password.String == other.Password.String) &&
		r.NoLogin.Valid == other.NoLogin.Valid &&
		(!r.NoLogin.Valid || r.NoLogin.Bool == other.NoLogin.Bool) &&
		r.Reason.Valid == other.Reason.Valid &&
		(!r.Reason.Valid || r.Reason.String == other.Reason.String)
}
