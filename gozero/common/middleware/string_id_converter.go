package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var stringIDFields = map[string]struct{}{
	"id": {}, "ids": {}, "userId": {}, "userIds": {}, "orgId": {}, "orgIds": {},
	"roleId": {}, "roleIds": {}, "menuId": {}, "menuIds": {}, "configId": {}, "configIds": {},
	"dictId": {}, "dictIds": {}, "envId": {}, "envIds": {}, "parentId": {}, "parentIds": {},
	"storageId": {}, "storageIds": {}, "attachmentId": {}, "attachmentIds": {},
	"clientId": {}, "clientIds": {}, "createBy": {}, "updateBy": {},
}

func StringIDConverter(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch) &&
			strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			body, err := io.ReadAll(r.Body)
			if err == nil {
				r.Body = io.NopCloser(bytes.NewReader(body))
				if converted, ok := convertStringIDs(body); ok {
					r.Body = io.NopCloser(bytes.NewReader(converted))
					r.ContentLength = int64(len(converted))
				}
			}
		}
		next(w, r)
	}
}

func convertStringIDs(body []byte) ([]byte, bool) {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if decoder.Decode(&value) != nil || !convertIDValue(value) {
		return body, false
	}
	converted, err := json.Marshal(value)
	return converted, err == nil
}

func convertIDValue(value any) bool {
	converted := false
	switch value := value.(type) {
	case map[string]any:
		for key, field := range value {
			if _, ok := stringIDFields[key]; ok {
				switch field := field.(type) {
				case string:
					if number, err := strconv.ParseInt(field, 10, 64); err == nil {
						value[key], converted = number, true
					}
				case []any:
					for index, item := range field {
						if text, ok := item.(string); ok {
							if number, err := strconv.ParseInt(text, 10, 64); err == nil {
								field[index], converted = number, true
							}
						}
					}
				}
			}
			if convertIDValue(field) {
				converted = true
			}
		}
	case []any:
		for _, item := range value {
			if convertIDValue(item) {
				converted = true
			}
		}
	}
	return converted
}
