package sysservicelogic

import "database/sql"

func nullString(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func nullInt64(v sql.NullInt64) int64 {
	if v.Valid {
		return v.Int64
	}
	return 0
}

func nullTime(v sql.NullTime) string {
	if v.Valid {
		return v.Time.Format("2006-01-02 15:04:05")
	}
	return ""
}
