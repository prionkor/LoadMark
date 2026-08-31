package k6

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func encodeBody(body map[string]string, contentType string) (string, error) {
	switch contentType {
	case "application/x-www-form-urlencoded":
		values := url.Values{}

		for key, value := range body {
			values.Set(key, value)
		}
		return values.Encode(), nil

	case "application/json":
		data, err := json.Marshal(body)
		if err != nil {
			return "", fmt.Errorf("failed to encode body to JSON: %w", err)
		}
		return string(data), nil

	default:
		return "", fmt.Errorf("unsupported content type: %s", contentType)
	}
}
