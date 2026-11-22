package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// A symlink inside the directory must not lead the helpers outside of it.
func TestConfinedHelpersDoNotFollowSymlinkOut(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "scripts")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(base, "secret")
	if err := os.WriteFile(secret, []byte("s3cret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(secret, filepath.Join(dir, "link.sh")); err != nil {
		t.Fatal(err)
	}

	if data, err := ConfinedReadFile(dir, "link.sh"); err == nil {
		t.Fatalf("read through symlink: %q", data)
	}
	if err := ConfinedWriteFile(dir, "link.sh", []byte("x"), 0o600); err == nil {
		t.Fatal("wrote through symlink")
	}
	if got, _ := os.ReadFile(secret); string(got) != "s3cret" {
		t.Fatalf("target changed: %q", got)
	}
}

func TestConfinedHelpersRoundTrip(t *testing.T) {
	dir := t.TempDir()
	if err := ConfinedWriteFile(dir, "a.sh", []byte("echo hi"), 0o700); err != nil {
		t.Fatal(err)
	}
	data, err := ConfinedReadFile(dir, "a.sh")
	if err != nil || string(data) != "echo hi" {
		t.Fatalf("read back %q, %v", data, err)
	}
	if err := ConfinedRemove(dir, "a.sh"); err != nil {
		t.Fatal(err)
	}
	if err := ConfinedRemove(dir, "../a.sh"); err == nil {
		t.Fatal("traversal accepted")
	}
}
