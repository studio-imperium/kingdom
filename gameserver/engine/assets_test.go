package engine

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var testAssets *httptest.Server

func TestMain(m *testing.M) {
	testAssets = httptest.NewServer(http.FileServer(http.Dir("../../gateway/assets")))
	code := m.Run()
	testAssets.Close()
	os.Exit(code)
}

func TestAssetLoadFailure(t *testing.T) {
	for _, status := range []int{http.StatusServiceUnavailable, http.StatusOK} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			w.Write([]byte("not JSON"))
		}))
		err := InitAssets(server.URL + "/")
		server.Close()
		if err == nil {
			t.Fatalf("accepted invalid asset response (HTTP %d)", status)
		}
	}
}
