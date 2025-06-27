package util

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

func EncodeCursor(ts int64, id string) string {
	cursor := fmt.Sprintf("%d:%s", ts, id)
	return base64.RawURLEncoding.EncodeToString([]byte(cursor))
}

func DecodeCursor(encoded string) (int64, string, error) {
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return 0, "", err
	}
	parts := strings.SplitN(string(data), ":", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid cursor format")
	}
	ts, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", err
	}
	return ts, parts[1], nil
}
