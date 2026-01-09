package db

import (
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
)

var (
	ErrAffectedRowsMismatch = errors.New("affected rows do not match expectation")
)

func Exec(db sq.BaseRunner, q sq.Sqlizer, expectAffectedRows int64) error {
	var result sql.Result
	var err error

	switch q := q.(type) {
	case sq.InsertBuilder:
		result, err = q.
			PlaceholderFormat(sq.Dollar).
			RunWith(db).
			Exec()
	case sq.UpdateBuilder:
		result, err = q.
			PlaceholderFormat(sq.Dollar).
			RunWith(db).
			Exec()
	case sq.DeleteBuilder:
		result, err = q.
			PlaceholderFormat(sq.Dollar).
			RunWith(db).
			Exec()
	default:
		sql, args, err := q.ToSql()
		if err != nil {
			return err
		}
		_, err = db.Exec(sql, args...)
		return err
	}

	if err != nil {
		return err
	}

	if expectAffectedRows > 0 {
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected != expectAffectedRows {
			return ErrAffectedRowsMismatch
		}
	}

	return nil
}
