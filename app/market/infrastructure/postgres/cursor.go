package postgres

import (
	"encoding/base64"
	"strconv"
)

// encodeCursor encodes an int64 id into a URL-safe base64 cursor.
func encodeCursor(id int64) string {
	return base64.URLEncoding.EncodeToString([]byte(strconv.FormatInt(id, 10)))
}

// decodeCursor decodes a previously-issued cursor. An empty string returns 0.
func decodeCursor(raw string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	data, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(string(data), 10, 64)
}
