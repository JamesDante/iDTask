package utils

import (
	"database/sql"
	"fmt"
	"time"
)

// FormatNullInt converts sql.NullInt64 to string
func FormatNullInt(n sql.NullInt64) string {
	if n.Valid {
		return fmt.Sprintf("%d", n.Int64)
	}
	return "null"
}

// FormatNullString converts sql.NullString to string
func FormatNullString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return "null"
}

// FormatNullTime converts sql.NullTime to RFC3339 string
func FormatNullTime(t sql.NullTime) string {
	if t.Valid {
		return t.Time.Format(time.RFC3339)
	}
	return "null"
}
