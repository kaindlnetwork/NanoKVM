package application

import (
	"os"
	"path/filepath"
	"testing"
)

// Update packages run as root and their SHA-512 comes from the same server
// as the package. Over plain http anyone on the path can replace both.
func TestCustomUpdateServerRequiresHTTPS(t *testing.T) {
	if _, err := normalizeUpdateServerURL("http://updates.example.com/nanokvm"); err == nil {
		t.Fatal("plain http update server accepted")
	}
	got, err := normalizeUpdateServerURL("https://updates.example.com/nanokvm/")
	if err != nil || got != "https://updates.example.com/nanokvm" {
		t.Fatalf("https update server: %q, %v", got, err)
	}
}

// A stored http server from before the fix must not be used for updates.
func TestStoredHTTPUpdateServerIsNotUsed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "application-update.json")
	if err := os.WriteFile(path, []byte(`{"enabled":true,"url":"http://updates.example.com/nanokvm"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := loadUpdateServerConfigFromPath(path)
	if err == nil && cfg.Enabled && cfg.URL == "http://updates.example.com/nanokvm" {
		t.Fatal("stored http update server would be used")
	}
}
