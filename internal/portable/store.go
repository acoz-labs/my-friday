// Package portable implements the versioned, harness-independent assistant store.
package portable

import (
	"bytes"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const FormatVersion = 1

//go:embed schemas/*.json
var schemas embed.FS

var identifier = regexp.MustCompile(`^[a-z][a-z0-9-]{2,127}$`)

type Agent struct {
	FormatVersion  int    `json:"format_version"`
	ID             string `json:"id"`
	Name           string `json:"name"`
	DefaultHarness string `json:"default_harness"`
}
type Device struct {
	Version int    `json:"schema_version"`
	ID      string `json:"id"`
	Label   string `json:"label"`
}
type Source struct {
	Version    int    `json:"schema_version"`
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Summary    string `json:"summary"`
	DeviceID   string `json:"device_id"`
	RecordedAt string `json:"recorded_at"`
}
type Store struct {
	Root               string
	Agent              Agent
	gitSettings        *SyncConfig
	checkpointObserver *Authorship
}

func NewID(prefix string) string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return prefix + "-" + hex.EncodeToString(b)
}

func Create(root, name, harness, deviceID, label string) (*Store, error) {
	if strings.TrimSpace(name) == "" || strings.ContainsAny(name, "\x00\r\n") {
		return nil, errors.New("agent name is required and must be one line")
	}
	if harness != "codex" && harness != "pi" {
		return nil, errors.New("harness must be codex or pi")
	}
	if !identifier.MatchString(deviceID) || strings.TrimSpace(label) == "" {
		return nil, errors.New("valid device ID and label required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if _, err = os.Lstat(abs); !os.IsNotExist(err) {
		return nil, fmt.Errorf("target already exists or cannot be inspected: %s", abs)
	}
	if err = os.MkdirAll(filepath.Dir(abs), 0700); err != nil {
		return nil, err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(abs))
	if err != nil {
		return nil, err
	}
	abs = filepath.Join(parent, filepath.Base(abs))
	staging, err := os.MkdirTemp(parent, ".my-friday-create-")
	if err != nil {
		return nil, err
	}
	// Only this newly created temporary tree is owned by this operation.
	defer os.RemoveAll(staging)
	s := &Store{Root: staging, Agent: Agent{FormatVersion: FormatVersion, ID: NewID("assistant"), Name: name, DefaultHarness: harness}}
	for _, dir := range []string{".my-friday", "instructions", "memory/records", "memory/events", "memory/sources", "provenance/devices", "provenance/changes", "capabilities", "integrations", "extensions"} {
		if err = os.MkdirAll(filepath.Join(staging, dir), 0700); err != nil {
			return nil, err
		}
		if err = os.WriteFile(filepath.Join(staging, dir, ".gitkeep"), nil, 0600); err != nil {
			return nil, err
		}
	}
	if err = writeNewJSON(filepath.Join(staging, "agent.json"), s.Agent); err != nil {
		return nil, err
	}
	if err = s.AddDevice(Device{Version: 1, ID: deviceID, Label: label}); err != nil {
		return nil, err
	}
	files := map[string]string{
		".gitignore":                ".my-friday/write.lock\n.my-friday/local/\n.DS_Store\n",
		"instructions/identity.md":  "# Identity\n\nYour name is " + name + ". Be candid, concise, and useful.\n",
		"instructions/operating.md": "# Working behavior\n\nCarry authorized tasks through to a verified outcome. Investigate available APIs, CLI tools, browser tools, and computer use when a route fails. Use the tools actually available on this installation.\n\nCurrent explicit user direction supersedes older remembered user guidance within its scope. Distinguish one-task exceptions from ongoing changes. Record meaningful outcomes, corrections, preferences, and reusable procedures automatically. Keep inference distinguishable from user direction.\n\nYou may improve and verify executable capabilities and subscriptions in this private assistant repository as part of authorized work. Preserve history and recovery. A code rollback does not undo external effects. Keep credentials out of memory and transcripts.\n",
	}
	for path, content := range files {
		if err = os.WriteFile(filepath.Join(staging, path), []byte(content), 0600); err != nil {
			return nil, err
		}
	}
	if err = s.Validate(); err != nil {
		return nil, err
	}
	if err = renameNewDirectory(staging, abs); err != nil {
		return nil, err
	}
	s.Root = abs
	return s, nil
}

func Open(root string) (*Store, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("assistant root must be a real directory")
	}
	s := &Store{Root: abs}
	if err = readJSON(filepath.Join(abs, "agent.json"), &s.Agent); err != nil {
		return nil, err
	}
	if s.Agent.FormatVersion != FormatVersion {
		return nil, fmt.Errorf("unsupported assistant format %d (writer supports %d)", s.Agent.FormatVersion, FormatVersion)
	}
	if !identifier.MatchString(s.Agent.ID) || strings.TrimSpace(s.Agent.Name) == "" {
		return nil, errors.New("invalid assistant identity")
	}
	if s.Agent.DefaultHarness != "codex" && s.Agent.DefaultHarness != "pi" {
		return nil, errors.New("unsupported default harness")
	}
	return s, nil
}

func readJSON(path string, out any) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file: %s", path)
	}
	if info.Size() > 4<<20 {
		return fmt.Errorf("record exceeds 4 MiB: %s", path)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err = d.Decode(out); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	var extra any
	if err = d.Decode(&extra); err != io.EOF {
		return fmt.Errorf("trailing JSON in %s", path)
	}
	return nil
}

func writeNewJSON(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	f, err := os.CreateTemp(filepath.Dir(path), ".write-")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Link(tmp, path); err != nil {
		return fmt.Errorf("immutable file already exists or cannot be created: %w", err)
	}
	d, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}

func (s *Store) withLock(fn func() error) error {
	if _, err := Open(s.Root); err != nil {
		return err
	}
	if err := s.checkDirectories(); err != nil {
		return err
	}
	path := filepath.Join(s.Root, ".my-friday", "write.lock")
	fd, err := syscall.Open(path, syscall.O_CREAT|syscall.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)
	if err = syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return errors.New("assistant writer busy; retry the operation")
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)
	return fn()
}

func (s *Store) checkDirectories() error {
	for _, dir := range []string{".my-friday", "memory", "memory/records", "memory/events", "memory/sources", "provenance", "provenance/devices", "provenance/changes", "capabilities", "integrations", "instructions"} {
		info, err := os.Lstat(filepath.Join(s.Root, dir))
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("invalid assistant directory: %s", dir)
		}
	}
	return nil
}

func (s *Store) AddDevice(d Device) error {
	if d.Version != 1 || !identifier.MatchString(d.ID) || strings.TrimSpace(d.Label) == "" {
		return errors.New("invalid device record")
	}
	return s.withLock(func() error { return writeNewJSON(filepath.Join(s.Root, "provenance/devices", d.ID+".json"), d) })
}
func (s *Store) AddSource(source Source) error {
	if source.Version != 1 || !identifier.MatchString(source.ID) || strings.TrimSpace(source.Summary) == "" || strings.TrimSpace(source.Kind) == "" {
		return errors.New("invalid source record")
	}
	if _, err := time.Parse(time.RFC3339Nano, source.RecordedAt); err != nil {
		return err
	}
	if err := s.deviceExists(source.DeviceID); err != nil {
		return err
	}
	return s.withLock(func() error { return writeNewJSON(filepath.Join(s.Root, "memory/sources", source.ID+".json"), source) })
}
func (s *Store) deviceExists(id string) error {
	if !identifier.MatchString(id) {
		return errors.New("invalid device ID")
	}
	var d Device
	if err := readJSON(filepath.Join(s.Root, "provenance/devices", id+".json"), &d); err != nil {
		return fmt.Errorf("unknown device %s: %w", id, err)
	}
	if d.ID != id || d.Version != 1 || strings.TrimSpace(d.Label) == "" {
		return errors.New("invalid device identity")
	}
	return nil
}

func memorySchema() (*jsonschema.Schema, error) {
	b, err := schemas.ReadFile("schemas/memory-revision.schema.json")
	if err != nil {
		return nil, err
	}
	var value any
	if err = json.Unmarshal(b, &value); err != nil {
		return nil, err
	}
	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()
	if err = compiler.AddResource("memory.json", value); err != nil {
		return nil, err
	}
	return compiler.Compile("memory.json")
}

func (s *Store) revisions() ([]Revision, error) {
	if err := s.checkDirectories(); err != nil {
		return nil, err
	}
	records := []Revision{}
	err := filepath.WalkDir(filepath.Join(s.Root, "memory/records"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink in memory: %s", path)
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == ".gitkeep" {
			return nil
		}
		if !strings.HasSuffix(path, ".json") {
			return fmt.Errorf("unexpected memory file: %s", path)
		}
		var r Revision
		if err := readJSON(path, &r); err != nil {
			return err
		}
		if filepath.Base(filepath.Dir(path)) != r.RecordID || filepath.Base(path) != r.ID+".json" {
			return errors.New("memory ID/path mismatch")
		}
		records = append(records, r)
		return nil
	})
	return records, err
}

func (s *Store) Validate() error {
	current, err := Open(s.Root)
	if err != nil {
		return err
	}
	if current.Agent.ID != s.Agent.ID {
		return errors.New("assistant identity changed during operation")
	}
	records, err := s.revisions()
	if err != nil {
		return err
	}
	if err := s.validateGraph(records); err != nil {
		return err
	}
	if _, err := s.SourceChanges(""); err != nil {
		return err
	}
	caps, err := s.Capabilities()
	if err != nil {
		return err
	}
	for _, c := range caps {
		for _, sub := range c.Subscriptions {
			if _, err := orderedHandlers(caps, sub.Event); err != nil {
				return err
			}
		}
	}
	_, err = s.syncConfiguration()
	return err
}
