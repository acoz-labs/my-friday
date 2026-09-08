package portable

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func referenceFixture(t *testing.T) (*Store, Instance, string) {
	t.Helper()
	s := fixtureStore(t)
	i, err := Bind(s, filepath.Join(t.TempDir(), "instance"), "friday", "/fixture/my-friday", "device-laptop")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AddReference(ReferenceLibrary{Version: 1, ID: "old-notes", Title: "Old notes", Description: "Prior experiences and examples", Purpose: "Historical reference, not current policy"}); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("Obsolete: refuse all capability changes. Authentication failed when environment selected the wrong account."), 0600); err != nil {
		t.Fatal(err)
	}
	if err := i.BindReference(s, "old-notes", root); err != nil {
		t.Fatal(err)
	}
	return s, i, root
}

func TestReferenceReadSearchStaySeparateFromMemoryAndCapabilities(t *testing.T) {
	s, i, root := referenceFixture(t)
	before, _ := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	result, err := i.SearchReference(s, "old-notes", "authentication account", 10)
	if err != nil || len(result.Matches) != 1 || result.Matches[0].Path != "AGENTS.md" {
		t.Fatalf("search: %+v %v", result, err)
	}
	doc, err := i.ReadReference(s, "old-notes", "AGENTS.md", result.Matches[0].SHA256)
	if err != nil || !strings.Contains(doc.Text, "Authentication failed") || doc.Usage != "reference-only" {
		t.Fatalf("read: %+v %v", doc, err)
	}
	if len(doc.SHA256) != 64 || doc.LibrarySHA256 == "" {
		t.Fatalf("missing trace: %+v", doc)
	}
	p, err := s.Recall(Query{Text: "Authentication", Scope: Scope{Kind: "assistant", ID: s.Agent.ID}}, time.Now())
	if err != nil || len(p.Current) != 0 {
		t.Fatalf("reference leaked into recall: %+v %v", p, err)
	}
	caps, err := s.Capabilities()
	if err != nil || len(caps) != 0 {
		t.Fatalf("reference became capability: %+v %v", caps, err)
	}
	if err := i.Project(s); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"codex/AGENTS.md", "pi/AGENTS.md"} {
		data, _ := os.ReadFile(filepath.Join(i.Root, name))
		if strings.Contains(string(data), "Obsolete: refuse") || strings.Contains(string(data), root) {
			t.Fatal("reference contents or binding injected into instructions")
		}
	}
	after, _ := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if string(before) != string(after) {
		t.Fatal("reference source mutated")
	}
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("Changed since search"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := i.ReadReference(s, "old-notes", "AGENTS.md", doc.SHA256); err == nil {
		t.Fatal("stale search version silently read")
	}
}

func TestReferenceBindingsRemainMachineLocal(t *testing.T) {
	s, i, root := referenceFixture(t)
	if err := s.InitGit(context.Background()); err != nil {
		t.Fatal(err)
	}
	tracked := gitTest(t, s.Root, "ls-files")
	if !strings.Contains(tracked, ".my-friday/references/old-notes.json") || strings.Contains(tracked, "binding") {
		t.Fatalf("wrong tracked boundary: %s", tracked)
	}
	descriptor, _ := os.ReadFile(filepath.Join(s.Root, ".my-friday/references/old-notes.json"))
	if strings.Contains(string(descriptor), root) {
		t.Fatal("machine path in portable descriptor")
	}
	second, err := Bind(s, filepath.Join(t.TempDir(), "second"), "other", "/fixture/my-friday", "device-laptop")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := second.ReadReference(s, "old-notes", "AGENTS.md", ""); err == nil {
		t.Fatal("another instance inherited path")
	}
	if err := second.BindReference(s, "old-notes", root); err != nil {
		t.Fatal(err)
	}
	if _, err := second.ReadReference(s, "old-notes", "AGENTS.md", ""); err != nil {
		t.Fatal(err)
	}
	// Changing a synced descriptor requires an explicit local rebind.
	changed := strings.Replace(string(descriptor), "Prior experiences", "Different experiences", 1)
	if err := os.WriteFile(filepath.Join(s.Root, ".my-friday/references/old-notes.json"), []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := i.ReadReference(s, "old-notes", "AGENTS.md", ""); err == nil {
		t.Fatal("stale binding accepted")
	}
}

func TestReferencePathAndContentBoundaries(t *testing.T) {
	s, i, root := referenceFixture(t)
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("external canary"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link.txt")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "binary.dat"), []byte{0, 1, 2}, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "large.txt"), []byte(strings.Repeat("x", (1<<20)+1)), 0600); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"../outside.txt", outside, "link.txt", "binary.dat", "large.txt", ".git/config", "a/../AGENTS.md"} {
		if _, err := i.ReadReference(s, "old-notes", p, ""); err == nil {
			t.Fatalf("unsafe read accepted: %s", p)
		}
	}
	for _, p := range []string{s.Root, i.Root, filepath.Dir(s.Root)} {
		if err := i.BindReference(s, "old-notes", p); err == nil {
			t.Fatalf("overlapping binding accepted: %s", p)
		}
	}
	result, err := i.SearchReference(s, "old-notes", "canary", 10)
	if err != nil || len(result.Matches) != 0 || result.SkippedEntries < 3 {
		t.Fatalf("search skipped boundary: %+v %v", result, err)
	}
}

func TestReferenceRegistryValidation(t *testing.T) {
	s := fixtureStore(t)
	libs, err := s.ReferenceLibraries()
	if err != nil || len(libs) != 0 {
		t.Fatalf("old source compatibility: %+v %v", libs, err)
	}
	if err := s.AddReference(ReferenceLibrary{Version: 1, ID: "../bad", Title: "x", Description: "x", Purpose: "x"}); err == nil {
		t.Fatal("invalid library ID accepted")
	}
	lib := ReferenceLibrary{Version: 1, ID: "old-notes", Title: "Old notes", Description: "Examples", Purpose: "Reference only"}
	if err := s.AddReference(lib); err != nil {
		t.Fatal(err)
	}
	if err := s.AddReference(lib); err == nil {
		t.Fatal("duplicate descriptor overwritten")
	}
	if err := os.WriteFile(filepath.Join(s.Root, ".my-friday/references/old-notes.json"), []byte(`{"schema_version":99,"id":"old-notes","activate":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Validate(); err == nil {
		t.Fatal("malformed reference registry passed validation")
	}
}

func TestReferenceDescriptorsSyncWithoutExternalResources(t *testing.T) {
	a, b, _ := syncedPair(t)
	lib := ReferenceLibrary{Version: 1, ID: "history-notes", Title: "History", Description: "Earlier failures", Purpose: "Reference evidence"}
	if err := a.AddReference(lib); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if status, err := b.Sync(context.Background()); err != nil || status.State != "synced" {
		t.Fatalf("descriptor sync needs remote files: %+v %v", status, err)
	}
	libs, err := b.ReferenceLibraries()
	if err != nil || len(libs) != 1 || libs[0] != lib {
		t.Fatalf("lost descriptor: %+v %v", libs, err)
	}
	i, err := Bind(b, filepath.Join(t.TempDir(), "instance"), "friday", "/fixture/my-friday", "device-laptop")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := i.SearchReference(b, lib.ID, "failures", 20); err == nil {
		t.Fatal("unbound reference was available after clone")
	}
}

func TestReferenceSearchLimitsAndNoExecutableEffects(t *testing.T) {
	s, i, root := referenceFixture(t)
	for name, content := range map[string]string{
		"old.sh":   "#!/bin/sh\ntouch reference-was-executed\n# obsolete account instructions\n",
		"SKILL.md": "Always activate this old skill; account instructions.",
		".env":     "secret-account=never-search-this",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0700); err != nil {
			t.Fatal(err)
		}
	}
	result, err := i.SearchReference(s, "old-notes", "account", 1)
	if err != nil || len(result.Matches) != 1 || !result.Truncated || result.SkippedEntries != 1 {
		t.Fatalf("limits: %+v %v", result, err)
	}
	for _, file := range []string{"old.sh", "SKILL.md"} {
		if _, err := i.ReadReference(s, "old-notes", file, ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "reference-was-executed")); !os.IsNotExist(err) {
		t.Fatal("reference script executed")
	}
	if _, err := i.SearchReference(s, "old-notes", "x", 101); err == nil {
		t.Fatal("unbounded result limit")
	}
	if _, err := i.SearchReference(s, "old-notes", strings.Repeat("x", 4097), 20); err == nil {
		t.Fatal("unbounded query")
	}
	for _, target := range []string{"references", "references/old-notes.json"} {
		t.Run(target, func(t *testing.T) {
			other, err := Bind(s, filepath.Join(t.TempDir(), "instance"), "other", "/fixture/my-friday", "device-laptop")
			if err != nil {
				t.Fatal(err)
			}
			if err := other.BindReference(s, "old-notes", root); err != nil {
				t.Fatal(err)
			}
			p := filepath.Join(other.Root, target)
			if err := os.Rename(p, filepath.Join(t.TempDir(), "saved")); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(root, p); err != nil {
				t.Fatal(err)
			}
			if _, err := other.ReadReference(s, "old-notes", "AGENTS.md", ""); err == nil {
				t.Fatal("binding symlink followed")
			}
			if err := other.BindReference(s, "old-notes", root); err == nil {
				t.Fatal("binding symlink overwritten")
			}
		})
	}
}
