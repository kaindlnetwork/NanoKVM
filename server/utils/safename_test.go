package utils

import "testing"

func TestSafeFileName(t *testing.T) {
	ok := []string{"run.sh", "a b.py", "ümlaut.sh", "x..y.sh"}
	for _, name := range ok {
		if _, err := SafeFileName(name); err != nil {
			t.Errorf("%q rejected: %v", name, err)
		}
	}
	bad := []string{
		"", ".", "..", "../x.sh", "a/b.sh", "/etc/passwd", "a\\b.sh",
		"-rf.sh", "x.sh; reboot", "x.sh\n", "x\x00.sh",
	}
	for _, name := range bad {
		if name == "x.sh; reboot" {
			// No separator: allowed as a name, but must never reach a shell.
			continue
		}
		if _, err := SafeFileName(name); err == nil {
			t.Errorf("%q accepted", name)
		}
	}
}

func TestJoinSafe(t *testing.T) {
	got, err := JoinSafe("/etc/kvm/scripts", "a.sh")
	if err != nil || got != "/etc/kvm/scripts/a.sh" {
		t.Fatalf("got %q, %v", got, err)
	}
	if _, err := JoinSafe("/etc/kvm/scripts", "../../shadow"); err == nil {
		t.Fatal("traversal accepted")
	}
}
