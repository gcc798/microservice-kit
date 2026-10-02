package commonutil

import (
	"database/sql"
	"encoding/json"
	"time"
)

func NullString(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func NullInt64(v sql.NullInt64) int64 {
	if v.Valid {
		return v.Int64
	}
	return 0
}

func NullBool(v sql.NullBool) bool {
	return v.Valid && v.Bool
}

func NullTime(v sql.NullTime) string {
	if v.Valid {
		return v.Time.Format("2006-01-02 15:04:05")
	}
	return ""
}

func TimePtrString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func NormalizePage(pageNum, pageSize int64) (int64, int64) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 1000 {
		pageSize = 1000
	}
	return pageNum, pageSize
}

func PageData(records interface{}, total, pageNum, pageSize int64) map[string]interface{} {
	pages := int64(0)
	if pageSize > 0 {
		pages = (total + pageSize - 1) / pageSize
	}
	return map[string]interface{}{
		"records": records,
		"total":   total,
		"size":    pageSize,
		"current": pageNum,
		"pages":   pages,
	}
}

func InterfaceToJSONString(v interface{}) (string, error) {
	if v == nil {
		return "", nil
	}
	switch vv := v.(type) {
	case string:
		return vv, nil
	case []byte:
		return string(vv), nil
	default:
		bs, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(bs), nil
	}
}

func JSONStringToValue(v string) interface{} {
	if v == "" {
		return nil
	}
	var data interface{}
	if err := json.Unmarshal([]byte(v), &data); err == nil {
		return data
	}
	return v
}
