package utils

import (
	"database/sql"
	"time"
)

// ToBoolPtr converts various types to a *bool.
// Supported are: bool, *bool, sql.NullBool.
func ToBoolPtr(value any) (*bool, bool) {
	switch v := value.(type) {
	case bool:
		return &v, true
	case *bool:
		return v, true
	case sql.NullBool:
		if v.Valid {
			return &v.Bool, true
		}
		return nil, true
	}
	return nil, false
}

// ToStringPtr converts various types to a *string.
// Supported are: string, *string, sql.NullString.
func ToStringPtr(value any) (*string, bool) {
	switch v := value.(type) {
	case string:
		return &v, true
	case *string:
		return v, true
	case sql.NullString:
		if v.Valid {
			return &v.String, true
		}
		return nil, true
	}
	return nil, false
}

// ToTimePtr converts various types to a *time.Time.
// Supported are: time.Time, *time.Time, sql.NullTime.
func ToTimePtr(value any) (*time.Time, bool) {
	switch v := value.(type) {
	case time.Time:
		return &v, true
	case *time.Time:
		return v, true
	case sql.NullTime:
		if v.Valid {
			return &v.Time, true
		}
		return nil, true
	}
	return nil, false
}

// ToUint64Ptr converts various types to a *uint64.
// Supported are:
// - uint64, *uint64, int64, *int64, sql.NullInt64
// - uint32, *uint32, int32, *int32, sql.NullInt32
// - uint16, *uint16, int16, *int16, sql.NullInt16
// - uint8, *uint8, int8, *int8
// - uint, *uint, int, *int
func ToUint64Ptr(value any) (*uint64, bool) {
	switch v := value.(type) {
	case uint64:
		return &v, true
	case *uint64:
		return v, true
	case int64:
		vv := uint64(v)
		return &vv, true
	case *int64:
		if v != nil {
			vv := uint64(*v)
			return &vv, true
		}
		return nil, true
	case sql.NullInt64:
		if v.Valid {
			vv := uint64(v.Int64)
			return &vv, true
		}
		return nil, true
	case uint32:
		vv := uint64(v)
		return &vv, true
	case *uint32:
		if v != nil {
			vv := uint64(*v)
			return &vv, true
		}
		return nil, true
	case int32:
		vv := uint64(v)
		return &vv, true
	case *int32:
		if v != nil {
			vv := uint64(*v)
			return &vv, true
		}
		return nil, true
	case sql.NullInt32:
		if v.Valid {
			vv := uint64(v.Int32)
			return &vv, true
		}
		return nil, true
	case uint16:
		vv := uint64(v)
		return &vv, true
	case *uint16:
		if v != nil {
			vv := uint64(*v)
			return &vv, true
		}
		return nil, true
	case int16:
		vv := uint64(v)
		return &vv, true
	case *int16:
		if v != nil {
			vv := uint64(*v)
			return &vv, true
		}
		return nil, true
	case sql.NullInt16:
		if v.Valid {
			vv := uint64(v.Int16)
			return &vv, true
		}
		return nil, true
	case uint8:
		vv := uint64(v)
		return &vv, true
	case *uint8:
		if v != nil {
			vv := uint64(*v)
			return &vv, true
		}
		return nil, true
	case int8:
		vv := uint64(v)
		return &vv, true
	case *int8:
		if v != nil {
			vv := uint64(*v)
			return &vv, true
		}
		return nil, true
	case int:
		vv := uint64(v)
		return &vv, true
	case *int:
		if v != nil {
			vv := uint64(*v)
			return &vv, true
		}
		return nil, true
	case uint:
		vv := uint64(v)
		return &vv, true
	case *uint:
		if v != nil {
			vv := uint64(*v)
			return &vv, true
		}
		return nil, true
	}
	return nil, false
}
