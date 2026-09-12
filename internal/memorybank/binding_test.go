package memorybank

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBindingIsLocalAndPinsBankIdentity(t *testing.T) {
	s := fixture(t)
	path := filepath.Join(t.TempDir(), "configuration", "memory.json")
	b, err := Bind(s.Root(), path, "Second machine", "Example user")
	if err != nil {
		t.Fatal(err)
	}
	if b.DeviceID == "device-test" || b.BankID != s.ID() {
		t.Fatalf("bad local binding: %+v", b)
	}
	opened, err := OpenBinding(path, "codex")
	if err != nil || opened.ID() != s.ID() {
		t.Fatalf("binding open: %v", err)
	}
	if _, err := Bind(s.Root(), path, "Third machine", "Example"); err == nil {
		t.Fatal("existing binding overwritten")
	}
	if _, err := Bind(s.Root(), filepath.Join(s.Root(), "binding.json"), "Test", "Example"); err == nil {
		t.Fatal("machine-local binding placed inside bank")
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("binding permission: %v %v", info, err)
	}
	other := fixture(t)
	if err := os.Rename(s.Root(), s.Root()+"-saved"); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(other.Root(), s.Root()); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenBinding(path, "codex"); err == nil {
		t.Fatal("replacement bank silently accepted")
	}
	if _, err := opened.Scopes(); err == nil {
		t.Fatal("already-running service read a replacement bank")
	}
}

func TestBindingPreflightRefusesUnresolvedSymlinkIntoFutureBank(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "future-bank")
	alias := filepath.Join(base, "alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatal(err)
	}
	if err := ValidateBindingDestination(root, filepath.Join(alias, "binding.json")); err == nil {
		t.Fatal("unresolved symlink hid a future in-bank binding")
	}
	if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatal("preflight created the bank")
	}
}
