package gateway

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
)

type SessionToken [sha256.Size]byte

func post(token SessionToken, path, contentType string, body []byte, result any) error {
	req, err := http.NewRequest(http.MethodPost, gateway_address+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("gateway %s: %w", path, err)
	}
	req.Header.Set("Authorization", "Bearer "+hex.EncodeToString(token[:]))
	req.Header.Set("Content-Type", contentType)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("gateway %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gateway %s: HTTP %d", path, resp.StatusCode)
	}
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}
