package portable

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceChangesCaptureVersionsAndCheckpointDevice(t *testing.T) {
	s := fixtureStore(t).WithCheckpointObserver(Authorship{DeviceID: "device-laptop", Actor: "Friday", Harness: "cli"})
	ctx := context.Background()
	if err := s.InitGit(ctx); err != nil {
		t.Fatal(err)
	}
	changes, err := s.SourceChanges("instructions/identity.md")
	if err != nil || len(changes) != 1 {
		t.Fatalf("initial provenance: %+v %v", changes, err)
	}
	first := changes[0]
	if first.Observer == nil || first.Observer.DeviceID != "device-laptop" || first.BaseCommit != "" {
		t.Fatalf("initial observer: %+v", first)
	}
	path := filepath.Join(s.Root, "instructions/identity.md")
	if err := os.WriteFile(path, []byte("Changed identity\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	changes, err = s.SourceChanges("instructions/identity.md")
	if err != nil || len(changes) != 2 {
		t.Fatalf("changed provenance: %+v %v", changes, err)
	}
	var update SourceChange
	for _, c := range changes {
		if c.BaseCommit != "" {
			update = c
		}
	}
	if len(update.Files) != 1 || update.Files[0].Before == nil || update.Files[0].After == nil || update.Files[0].Before.ObjectID == update.Files[0].After.ObjectID {
		t.Fatalf("missing before/after: %+v", update)
	}
	want := strings.TrimSpace(gitTest(t, s.Root, "rev-parse", "HEAD:instructions/identity.md"))
	if update.Files[0].After.ObjectID != want {
		t.Fatalf("wrong captured blob: %+v", update.Files)
	}
	if _, err := s.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := s.SourceChanges("")
	if len(after) != 2 {
		t.Fatalf("no-op sync recorded a change: %+v", after)
	}
	if err := s.Put(revision("revision-memory-only")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ = s.SourceChanges("")
	if len(after) != 2 {
		t.Fatal("memory-only write created redundant source provenance")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ = s.SourceChanges("instructions/identity.md")
	foundDelete := false
	for _, c := range after {
		for _, f := range c.Files {
			if f.Path == "instructions/identity.md" && f.After == nil {
				foundDelete = f.Before != nil
			}
		}
	}
	if !foundDelete {
		t.Fatal("deletion not captured")
	}
}

func TestSourceChangesSyncPreservesOriginalObserver(t *testing.T) {
	a, b, _ := syncedPair(t)
	ctx := context.Background()
	if err := b.AddDevice(Device{Version: 1, ID: "device-desktop", Label: "Desktop"}); err != nil {
		t.Fatal(err)
	}
	a = a.WithCheckpointObserver(Authorship{DeviceID: "device-laptop", Actor: "Friday", Harness: "codex"})
	b = b.WithCheckpointObserver(Authorship{DeviceID: "device-desktop", Actor: "Friday", Harness: "pi"})
	for _, s := range []*Store{a, b} {
		file := "left.md"
		if s == b {
			file = "right.md"
		}
		if err := os.WriteFile(filepath.Join(s.Root, "instructions", file), []byte("Synthetic instructions"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := a.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	if status, err := b.Sync(ctx); err != nil || status.State != "synced" {
		t.Fatalf("merge: %+v %v", status, err)
	}
	if _, err := a.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	for _, s := range []*Store{a, b} {
		for file, device := range map[string]string{"left.md": "device-laptop", "right.md": "device-desktop"} {
			changes, err := s.SourceChanges("instructions/" + file)
			if err != nil || len(changes) != 1 || changes[0].Observer == nil || changes[0].Observer.DeviceID != device {
				t.Fatalf("reattributed remote change: %+v %v", changes, err)
			}
		}
	}
}

func TestSourceChangesUnboundSyncDoesNotInventDevice(t *testing.T) {
	s := fixtureStore(t)
	if err := s.InitGit(context.Background()); err != nil {
		t.Fatal(err)
	}
	changes, err := s.SourceChanges("")
	if err != nil || len(changes) != 1 || changes[0].Observer != nil {
		t.Fatalf("invented observer: %+v %v", changes, err)
	}
	s = s.WithCheckpointObserver(Authorship{DeviceID: "device-unknown", Actor: "Friday", Harness: "cli"})
	if _, err := s.Sync(context.Background()); err == nil {
		t.Fatal("unknown observer accepted")
	}
}

func TestSourceChangeValidationAndAppendOnly(t *testing.T) {
	for _, kind := range []string{"rewrite", "bad path", "bad timestamp", "unknown device", "unknown field", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			s := fixtureStore(t)
			if err := s.InitGit(context.Background()); err != nil {
				t.Fatal(err)
			}
			changes, err := s.SourceChanges("")
			if err != nil || len(changes) != 1 {
				t.Fatalf("capture: %+v %v", changes, err)
			}
			c := changes[0]
			path := filepath.Join(s.Root, "provenance/changes", c.ID+".json")
			switch kind {
			case "rewrite":
				c.RecordedAt = "2026-01-01T00:00:00Z"
			case "bad path":
				c.Files[0].Path = "../outside"
			case "bad timestamp":
				c.RecordedAt = "not a date"
			case "unknown device":
				c.Observer = &Authorship{DeviceID: "device-unknown", Actor: "Friday", Harness: "cli"}
			}
			data, _ := json.Marshal(c)
			if kind == "unknown field" {
				data = append(data[:len(data)-1], []byte(`,"unknown":true}`)...)
			}
			if kind == "symlink" {
				if err := os.Rename(path, filepath.Join(t.TempDir(), "original.json")); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("/dev/null", path); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := s.Sync(context.Background()); err == nil {
				t.Fatal("invalid or rewritten provenance accepted")
			}
		})
	}
}

func TestSourceChangeRawPathsAndModes(t *testing.T) {
	s := fixtureStore(t)
	ctx := context.Background()
	if err := s.InitGit(ctx); err != nil {
		t.Fatal(err)
	}
	rel := "extensions/tab\tline\nscript.sh"
	if err := os.WriteFile(filepath.Join(s.Root, rel), []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	changes, err := s.SourceChanges(rel)
	if err != nil || len(changes) != 1 || changes[0].Files[0].Path != rel || changes[0].Files[0].After.Mode != "100755" {
		t.Fatalf("filename or executable mode lost: %+v %v", changes, err)
	}
	for _, path := range []string{"../outside", "/absolute", "instructions/../instructions/identity.md"} {
		if _, err := s.SourceChanges(path); err == nil {
			t.Fatalf("invalid filter accepted: %q", path)
		}
	}
}

func TestSourceChangesRetryKeepsFirstObservation(t *testing.T) {
	s := fixtureStore(t).WithCheckpointObserver(Authorship{DeviceID: "device-laptop", Actor: "Friday", Harness: "cli"})
	ctx := context.Background()
	if err := s.InitGit(ctx); err != nil {
		t.Fatal(err)
	}
	rel := "extensions/retry.txt"
	if err := os.WriteFile(filepath.Join(s.Root, rel), []byte("Retry fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	base := strings.TrimSpace(gitTest(t, s.Root, "rev-parse", "HEAD"))
	// Simulate cancellation after capture but before commit. No deletion or
	// rewrite is needed to resume the next checkpoint.
	if err := s.withLock(func() error {
		if _, err := s.git(ctx, "add", "--all", "--", "."); err != nil {
			return err
		}
		return s.captureSourceChanges(ctx, base)
	}); err != nil {
		t.Fatal(err)
	}
	before, err := s.SourceChanges(rel)
	if err != nil || len(before) != 1 {
		t.Fatalf("capture: %+v %v", before, err)
	}
	gitTest(t, s.Root, "remote", "add", "origin", filepath.Join(t.TempDir(), "unavailable.git"))
	if status, err := s.Sync(ctx); err != nil || status.State != "pending" {
		t.Fatalf("offline retry: %+v %v", status, err)
	}
	after, err := s.SourceChanges(rel)
	if err != nil || len(after) != 1 || before[0].ID != after[0].ID || before[0].RecordedAt != after[0].RecordedAt {
		t.Fatalf("retry duplicated or rewrote observation: %+v %v", after, err)
	}
	if status := gitTest(t, s.Root, "status", "--porcelain"); status != "" {
		t.Fatalf("retry not committed: %s", status)
	}
}

func TestSourceChangesIdenticalUnboundEditsMerge(t *testing.T) {
	a, b, _ := syncedPair(t)
	ctx := context.Background()
	rel := "instructions/same.md"
	for _, s := range []*Store{a, b} {
		if err := os.WriteFile(filepath.Join(s.Root, rel), []byte("Same independent edit"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := a.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	if status, err := b.Sync(ctx); err != nil || status.State != "synced" {
		t.Fatalf("identical changes conflicted: %+v %v", status, err)
	}
	changes, err := b.SourceChanges(rel)
	if err != nil || len(changes) != 2 || changes[0].ID == changes[1].ID {
		t.Fatalf("lost independent observations: %+v %v", changes, err)
	}
}

func TestSourceChangesRejectInvalidRemoteRecord(t *testing.T) {
	a, b, _ := syncedPair(t)
	path := filepath.Join(b.Root, "provenance/changes/change-invalid.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":99,"id":"change-invalid"}`), 0600); err != nil {
		t.Fatal(err)
	}
	gitTest(t, b.Root, "add", ".")
	gitTest(t, b.Root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.test", "commit", "-m", "Invalid external record")
	gitTest(t, b.Root, "push", "origin", "main")
	before := gitTest(t, a.Root, "rev-parse", "HEAD")
	if status, err := a.Sync(context.Background()); err != nil || status.State != "conflict" {
		t.Fatalf("invalid remote adopted: %+v %v", status, err)
	}
	if after := gitTest(t, a.Root, "rev-parse", "HEAD"); before != after {
		t.Fatal("invalid remote changed local HEAD")
	}
}
