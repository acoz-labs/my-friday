package portable

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type InstallationEntry struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Problem string `json:"problem,omitempty"`
}

func (s *Store) ValidateToolkitCompatibility() error {
	if err := s.Validate(); err != nil {
		return err
	}
	_, err := s.syncConfiguration()
	return err
}

// Discovery reads only the known instances directory, never scans home/source.
// Broken entries remain visible so a damaged installation does not disappear.
func DiscoverInstances(directory string) ([]InstallationEntry, error) {
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	result := []InstallationEntry{}
	for _, entry := range entries {
		if !entry.IsDir() && entry.Type()&os.ModeSymlink == 0 {
			continue
		}
		item := InstallationEntry{Path: filepath.Join(directory, entry.Name()), Name: entry.Name()}
		i, _, err := LoadInstance(item.Path)
		if err != nil {
			item.Problem = "Installation needs inspection"
		} else {
			item.Name = i.Name
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *Store) SetDefaultHarness(parent context.Context, harness string) error {
	if harness != "codex" && harness != "pi" {
		return errors.New("choose codex or pi")
	}
	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()
	work := *s
	s = &work
	return s.withLock(func() error {
		if err := s.gitBoundary(ctx); err != nil {
			return err
		}
		if err := s.Validate(); err != nil {
			return err
		}
		dirty, err := s.git(ctx, "status", "--porcelain")
		if err != nil {
			return err
		}
		if dirty != "" {
			return errors.New("source has uncommitted work; checkpoint it before changing settings")
		}
		cfg, err := s.syncConfiguration()
		if err != nil {
			return err
		}
		s.gitSettings = &cfg
		fresh, err := Open(s.Root)
		if err != nil {
			return err
		}
		s.Agent = fresh.Agent
		if s.Agent.DefaultHarness == harness {
			return nil
		}
		s.Agent.DefaultHarness = harness
		data, _ := json.MarshalIndent(s.Agent, "", "  ")
		if err := replaceProjectionFile(filepath.Join(s.Root, "agent.json"), append(data, '\n')); err != nil {
			return err
		}
		if err := s.checkpoint(ctx); err != nil {
			return fmt.Errorf("harness saved locally; checkpoint needs attention: %w", err)
		}
		return nil
	})
}

func launcherBytes(i Instance) []byte {
	return []byte("#!/bin/sh\nexec " + shellQuote(i.Binary) + " agent launch --instance " + shellQuote(i.Root) + " \"$@\"\n")
}

type toolkitSnapshot struct {
	Name   string `json:"name"`
	Data   []byte `json:"data,omitempty"`
	Mode   uint32 `json:"mode"`
	Exists bool   `json:"exists"`
	After  string `json:"after_sha256"`
}
type ToolkitReceipt struct {
	Version  int               `json:"schema_version"`
	Before   Instance          `json:"before"`
	After    Instance          `json:"after"`
	Launcher string            `json:"launcher,omitempty"`
	Files    []toolkitSnapshot `json:"files"`
}

func digestBytes(data []byte) string { h := sha256.Sum256(data); return hex.EncodeToString(h[:]) }

// UseToolkit adopts the running toolkit. Backup only managed files, never native
// config, credentials, sessions or source. Rollback copies remain private/local.
func (i Instance) UseToolkit(s *Store, binary, launcher string) (string, error) {
	current, err := os.Executable()
	if err != nil {
		return "", err
	}
	current, err = filepath.EvalSymlinks(current)
	if err != nil {
		return "", err
	}
	binary, err = filepath.EvalSymlinks(binary)
	if err != nil || binary != current {
		return "", errors.New("adopt a toolkit from its own management menu")
	}
	var backup string
	err = s.withLock(func() error {
		fresh, _, err := LoadInstance(i.Root)
		if err != nil {
			return err
		}
		if fresh != i {
			return errors.New("installation binding changed; reopen its menu")
		}
		if err := s.Validate(); err != nil {
			return err
		}
		if _, err := s.syncConfiguration(); err != nil {
			return err
		}
		if err := i.preflightProjection(s); err != nil {
			return err
		}
		if launcher != "" {
			if err := ValidateInstallationPaths(s.Root, i.Root, launcher); err != nil {
				return err
			}
			st, err := os.Lstat(launcher)
			if err != nil || !st.Mode().IsRegular() {
				return errors.New("launcher must be an existing regular managed file; use repair for a missing launcher")
			}
			old, err := os.ReadFile(launcher)
			if err != nil || !bytes.Equal(old, launcherBytes(i)) {
				return errors.New("launcher is customized or belongs elsewhere; it was not replaced")
			}
		}
		next := i
		next.Binary = binary
		generated, err := next.projection(s)
		if err != nil {
			return err
		}
		data, _ := json.MarshalIndent(next, "", "  ")
		generated["binding.json"] = append(data, '\n')
		names := append([]string{"binding.json"}, projectionFiles...)
		if launcher != "" {
			names = append(names, "launcher")
			generated["launcher"] = launcherBytes(next)
		}
		receipt := ToolkitReceipt{Version: 1, Before: i, After: next, Launcher: launcher}
		for _, name := range names {
			target := filepath.Join(i.Root, name)
			if name == "launcher" {
				target = launcher
			}
			snap := toolkitSnapshot{Name: name, After: digestBytes(generated[name])}
			st, err := os.Lstat(target)
			if err == nil {
				if !st.Mode().IsRegular() {
					return errors.New("managed target is not a regular file")
				}
				snap.Exists = true
				snap.Mode = uint32(st.Mode().Perm())
				snap.Data, err = os.ReadFile(target)
			} else if os.IsNotExist(err) {
				err = nil
			}
			if err != nil {
				return err
			}
			receipt.Files = append(receipt.Files, snap)
		}
		parent := filepath.Join(i.Root, "toolkit-updates")
		if err := referenceDirectory(parent, true); err != nil {
			return err
		}
		backup, err = os.MkdirTemp(parent, time.Now().UTC().Format("20060102T150405.000000000")+"-")
		if err != nil {
			return err
		}
		if err := writeNewJSON(filepath.Join(backup, "receipt.json"), receipt); err != nil {
			return err
		}
		for _, snap := range receipt.Files {
			target := filepath.Join(i.Root, snap.Name)
			if snap.Name == "launcher" {
				target = launcher
			}
			if err = os.MkdirAll(filepath.Dir(target), 0700); err == nil {
				err = replaceProjectionFile(target, generated[snap.Name])
			}
			if err == nil && snap.Name == "launcher" {
				err = os.Chmod(target, 0700)
			}
			if err != nil {
				if restoreErr := restoreToolkit(i.Root, receipt); restoreErr != nil {
					return fmt.Errorf("update interrupted; rollback also needs attention; retain %s", backup)
				}
				return fmt.Errorf("update failed; previous managed files restored; backup: %s", backup)
			}
		}
		return nil
	})
	return backup, err
}

func restoreToolkit(root string, r ToolkitReceipt) error {
	for _, snap := range r.Files {
		target := filepath.Join(root, snap.Name)
		if snap.Name == "launcher" {
			target = r.Launcher
		}
		if !snap.Exists {
			if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
				return err
			}
			continue
		}
		if err := replaceProjectionFile(target, snap.Data); err != nil {
			return err
		}
		if err := os.Chmod(target, os.FileMode(snap.Mode)); err != nil {
			return err
		}
	}
	return nil
}

func (i Instance) ReadToolkitReceipt(backup string) (ToolkitReceipt, error) {
	var r ToolkitReceipt
	parent := filepath.Join(i.Root, "toolkit-updates")
	if filepath.Dir(filepath.Clean(backup)) != parent {
		return r, errors.New("choose a backup from this instance")
	}
	if err := referenceDirectory(parent, false); err != nil {
		return r, err
	}
	if err := referenceDirectory(backup, false); err != nil {
		return r, err
	}
	if err := readJSON(filepath.Join(backup, "receipt.json"), &r); err != nil {
		return r, err
	}
	if r.Version != 1 || r.After.AssistantID != i.AssistantID || r.Before.AssistantID != i.AssistantID || r.After.Repository != i.Repository || r.Before.Repository != i.Repository || r.After.Name != i.Name || r.Before.Name != i.Name || r.Before.DeviceID != i.DeviceID || r.After.DeviceID != i.DeviceID || !filepath.IsAbs(r.Before.Binary) {
		return r, errors.New("rollback receipt does not match this installation")
	}
	allowed := map[string]bool{"binding.json": true}
	for _, name := range projectionFiles {
		allowed[name] = true
	}
	if r.Launcher != "" {
		allowed["launcher"] = true
	}
	for _, f := range r.Files {
		if !allowed[f.Name] {
			return r, errors.New("invalid rollback target")
		}
		delete(allowed, f.Name)
	}
	if len(allowed) != 0 {
		return r, errors.New("incomplete rollback receipt")
	}
	if r.Launcher != "" {
		if err := ValidateInstallationPaths(i.Repository, i.Root, r.Launcher); err != nil {
			return r, err
		}
	}
	return r, nil
}

// Caller must verify the previous executable can validate today's source before
// explicit rollback. Refuse to overwrite any subsequent managed-file edits.
func (i Instance) RollbackToolkit(s *Store, backup string) error {
	return s.withLock(func() error {
		r, err := i.ReadToolkitReceipt(backup)
		if err != nil {
			return err
		}
		if err := i.preflightProjection(s); err != nil {
			return err
		}
		for _, f := range r.Files {
			target := filepath.Join(i.Root, f.Name)
			if f.Name == "launcher" {
				target = r.Launcher
			}
			st, err := os.Lstat(target)
			if err != nil || !st.Mode().IsRegular() {
				return errors.New("rollback target missing or not regular; inspect backup manually")
			}
			data, err := os.ReadFile(target)
			if err != nil || digestBytes(data) != f.After {
				return errors.New("managed files changed since update; preserve changes and inspect the backup before rollback")
			}
		}
		return restoreToolkit(i.Root, r)
	})
}
