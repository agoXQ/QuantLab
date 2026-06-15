// Package pagination provides cursor-based and offset-based pagination utilities.
package pagination

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
)

// Cursor encodes pagination cursor data.
type Cursor struct {
	NextID    int64  `json:"n"`
	NextValue string `json:"v,omitempty"`
}

// EncodeCursor encodes a Cursor to a base64 string.
func EncodeCursor(c Cursor) string {
	data, _ := json.Marshal(c)
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeCursor decodes a base64 cursor string.
func DecodeCursor(raw string) (Cursor, error) {
	data, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		return Cursor{}, err
	}
	var c Cursor
	if err := json.Unmarshal(data, &c); err != nil {
		return Cursor{}, err
	}
	return c, nil
}

// IDCursor creates a simple cursor from an int64 ID.
func IDCursor(id int64) string {
	return EncodeCursor(Cursor{NextID: id})
}

// PageMeta contains pagination metadata for API responses.
type PageMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// Limit validates and normalizes a limit parameter.
func Limit(raw string, defaultVal, maxVal int) int {
	if raw == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return defaultVal
	}
	if n > maxVal {
		return maxVal
	}
	return n
}
