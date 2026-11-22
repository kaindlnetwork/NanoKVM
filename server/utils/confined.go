package utils

import (
	"errors"
	"os"
)

// The helpers below confine file access to one directory with os.Root: the
// kernel-side lookup refuses "..", absolute paths and symlinks that lead
// outside dir. The name is additionally checked by SafeFileName, so only a
// single path element is ever accepted.

func ConfinedReadFile(dir, name string) ([]byte, error) {
	if _, err := SafeFileName(name); err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()
	return root.ReadFile(name)
}

func ConfinedWriteFile(dir, name string, data []byte, perm os.FileMode) error {
	if _, err := SafeFileName(name); err != nil {
		return err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	f, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func ConfinedRemove(dir, name string) error {
	if _, err := SafeFileName(name); err != nil {
		return err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	return root.Remove(name)
}

// ConfinedRegularFile reports an error unless name is a regular file directly
// inside dir. Symlinks are rejected even if they point inside dir, because a
// path handed to exec or python is resolved again outside of os.Root.
func ConfinedRegularFile(dir, name string) error {
	if _, err := SafeFileName(name); err != nil {
		return err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	info, err := root.Lstat(name)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return ErrNotRegularFile
	}
	return nil
}

var ErrNotRegularFile = errors.New("not a regular file")
