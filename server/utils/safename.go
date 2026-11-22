package utils

import (
	"errors"
	"path/filepath"
	"strings"
)

var ErrUnsafeFileName = errors.New("unsafe file name")

// SafeFileName accepts a single path element that is safe to join onto a
// fixed directory: no separators, no traversal, no control characters and no
// leading dash (so it can never be read as a command-line option).
func SafeFileName(name string) (string, error) {
	if name == "" || name == "." || name == ".." || len(name) > 255 {
		return "", ErrUnsafeFileName
	}
	if strings.ContainsAny(name, "/\\") || strings.HasPrefix(name, "-") {
		return "", ErrUnsafeFileName
	}
	for _, r := range name {
		if r < 0x20 || r == 0x7f {
			return "", ErrUnsafeFileName
		}
	}
	if filepath.Base(name) != name {
		return "", ErrUnsafeFileName
	}
	return name, nil
}

// JoinSafe joins a validated file name onto dir.
func JoinSafe(dir, name string) (string, error) {
	safe, err := SafeFileName(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, safe), nil
}
