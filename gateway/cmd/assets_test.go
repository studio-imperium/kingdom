package main

import (
	"bytes"
	"encoding/json"
	"gateway/assets"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublicAssets(t *testing.T) {
	api := createApiSurface()
	for _, name := range []string{"tiles", "items", "npcs", "animations", "projectiles", "bombs", "spawns", "loot"} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/assets/"+name+".json", nil)
		request.Header.Set("Origin", "http://localhost:8080")
		api.ServeHTTP(response, request)
		want, _ := assets.Files.ReadFile(name + ".json")
		if response.Code != http.StatusOK || !json.Valid(response.Body.Bytes()) || !bytes.Equal(response.Body.Bytes(), want) || response.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Fatalf("asset %s failed: status %d", name, response.Code)
		}
	}
	for path, status := range map[string]int{"/assets/tiles.json": http.StatusOK, "/assets/missing.json": http.StatusNotFound, "/assets/assets.go": http.StatusNotFound} {
		response := httptest.NewRecorder()
		api.ServeHTTP(response, httptest.NewRequest(http.MethodHead, path, nil))
		if response.Code != status || (status == http.StatusOK && response.Body.Len() != 0) {
			t.Fatalf("HEAD %s: status %d, bytes %d", path, response.Code, response.Body.Len())
		}
	}
}
