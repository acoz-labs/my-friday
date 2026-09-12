package memorybank

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/acoz-labs/my-friday/internal/portable"
)

// Binding is machine-local. Each clone is enrolled independently; moving the
// bank preserves its ID, while replacing it with another bank requires rebind.
type Binding struct {
	Version  int    `json:"schema_version"`
	BankID   string `json:"bank_id"`
	Root     string `json:"root"`
	DeviceID string `json:"device_id"`
	Actor    string `json:"actor"`
}

func DefaultBindingPath() (string, error) {
	if path := os.Getenv("MY_FRIDAY_MEMORY_BINDING"); path != "" {
		return path, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "my-friday", "memory.json"), nil
}

// prospectivePath resolves the existing ancestor before any directories are
// created, including when a symlink points a proposed local path into the bank.
func prospectivePath(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved, nil
	}
	if !os.IsNotExist(err) || filepath.Dir(path) == path {
		return "", err
	}
	parent, err := prospectivePath(filepath.Dir(path))
	if err != nil {
		return "", err
	}
	return filepath.Join(parent, filepath.Base(path)), nil
}

func Bind(root, path, label, actor string) (Binding, error) {
	var b Binding
	s, err := portable.Open(root)
	if err != nil {
		return b, err
	}
	if !s.IsMemoryBank() {
		return b, errors.New("only a memory bank can be bound")
	}
	if !textWithin(label, 256) || !textWithin(actor, 256) || path == "" {
		return b, errors.New("binding path, machine label, and actor required")
	}
	path, err = prospectivePath(path)
	if err != nil {
		return b, err
	}
	rel, err := filepath.Rel(s.Root, path)
	if err != nil || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
		return b, errors.New("machine binding must be outside the Git-backed bank")
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		return b, errors.New("binding already exists or cannot be inspected; use a new path, preserving the existing binding")
	}
	if err := s.Validate(); err != nil {
		return b, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return b, err
	}
	b = Binding{Version: 1, BankID: s.Agent.ID, Root: s.Root, DeviceID: portable.NewID("device"), Actor: actor}
	f, err := os.CreateTemp(filepath.Dir(path), ".memory-binding-*")
	if err != nil {
		return Binding{}, err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if err := json.NewEncoder(f).Encode(b); err != nil {
		return Binding{}, err
	}
	if err := f.Sync(); err != nil {
		return Binding{}, err
	}
	if err := f.Close(); err != nil {
		return Binding{}, err
	}
	if err := s.AddDevice(portable.Device{Version: 1, ID: b.DeviceID, Label: label}); err != nil {
		return Binding{}, err
	}
	// A failed publication may leave an unused device record. It must not
	// overwrite another binding, nor be reported as successful enrollment.
	if err := os.Link(f.Name(), path); err != nil {
		return Binding{}, err
	}
	return b, nil
}

func OpenBinding(path, harness string) (*Service, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 16384 {
		return nil, errors.New("binding must be a regular JSON file under 16 KiB")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	d := json.NewDecoder(io.LimitReader(f, 16385))
	d.DisallowUnknownFields()
	var b Binding
	if err := d.Decode(&b); err != nil {
		return nil, err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return nil, errors.New("binding contains trailing data")
	}
	if b.Version != 1 || !filepath.IsAbs(b.Root) || !textWithin(harness, 64) {
		return nil, errors.New("invalid binding version, root, or harness")
	}
	s, err := Open(b.Root, portable.Authorship{DeviceID: b.DeviceID, Actor: b.Actor, Harness: harness})
	if err != nil {
		return nil, err
	}
	if s.ID() != b.BankID {
		return nil, errors.New("bound bank identity changed; select the intended bank explicitly")
	}
	return s, nil
}
