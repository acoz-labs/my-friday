package portable

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstallationPathsKeepStateOutOfSource(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "agent")
	for _, state := range []string{repo, filepath.Join(repo, "instance"), base} {
		if err := ValidateInstallationPaths(repo, state, ""); err == nil {
			t.Fatalf("accepted overlapping source/state: %s", state)
		}
	}
	if err := ValidateInstallationPaths(repo, filepath.Join(base, "agent-state"), filepath.Join(repo, "launcher")); err == nil {
		t.Fatal("accepted machine-local launcher inside source")
	}
	if err := ValidateInstallationPaths(repo, filepath.Join(base, "agent-state"), filepath.Join(base, "bin/friday")); err != nil {
		t.Fatal("rejected sibling paths:", err)
	}
	if runtime.GOOS == "darwin" {
		if err := ValidateInstallationPaths(repo, filepath.Join(base, "AGENT/state"), ""); err == nil {
			t.Fatal("case-only overlap accepted")
		}
	}
	if err := os.Mkdir(repo, 0700); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(base, "alias")
	if err := os.Symlink(repo, alias); err != nil {
		t.Fatal(err)
	}
	if err := ValidateInstallationPaths(repo, filepath.Join(alias, "not-created-yet/state"), ""); err == nil {
		t.Fatal("symlink alias bypassed separation")
	}
}

func TestLoadInstanceRefusesPreviouslyNestedState(t *testing.T) {
	s := fixtureStore(t)
	state := filepath.Join(s.Root, "old-local-instance")
	if err := os.Mkdir(state, 0700); err != nil {
		t.Fatal(err)
	}
	i := Instance{Version: 1, Name: "friday", AssistantID: s.Agent.ID, Repository: s.Root, DeviceID: "device-laptop", Binary: "/fixture/my-friday"}
	data, err := json.Marshal(i)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(state, "binding.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := LoadInstance(state); err == nil {
		t.Fatal("legacy nested binding accepted before sync")
	}
}

func TestBindRefusesNestedStateBeforeCreatingIt(t *testing.T) {
	s := fixtureStore(t)
	state := filepath.Join(s.Root, "local-instance")
	if _, err := Bind(s, state, "friday", "/fixture/my-friday", "device-laptop"); err == nil {
		t.Fatal("nested instance accepted")
	}
	if _, err := os.Lstat(state); !os.IsNotExist(err) {
		t.Fatal("invalid binding left instance files")
	}
}

func TestProjectionRejectsSymlinkTargetsWithoutTouchingOtherFiles(t *testing.T) {
	for _, target := range []string{"codex", "pi/extensions", "codex/config.toml", "pi/AGENTS.md", "pi/extensions/my-friday.ts"} {
		t.Run(target, func(t *testing.T) {
			s := fixtureStore(t)
			i, err := Bind(s, filepath.Join(t.TempDir(), "state"), "friday", "/fixture/my-friday", "device-laptop")
			if err != nil {
				t.Fatal(err)
			}
			original := filepath.Join(i.Root, target)
			saved := original + ".saved"
			if err := os.Rename(original, saved); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(saved, original); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(filepath.Join(i.Root, "codex/AGENTS.md"))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(s.Root, "instructions/identity.md"), []byte("New identity, not yet projected"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := i.Project(s); err == nil {
				t.Fatal("projection followed symlink")
			}
			after, err := os.ReadFile(filepath.Join(i.Root, "codex/AGENTS.md"))
			if err != nil || string(before) != string(after) {
				t.Fatal("projection partially wrote before rejecting target")
			}
		})
	}
}

func TestProjectionRefreshPreservesNativeStateAndReplacesInodes(t *testing.T) {
	s := fixtureStore(t)
	i, err := Bind(s, filepath.Join(t.TempDir(), "state"), "friday", "/fixture/my-friday", "device-laptop")
	if err != nil {
		t.Fatal(err)
	}
	for _, harness := range []string{"codex", "pi"} {
		// Synthetic canaries, not real credentials.
		if err := os.WriteFile(filepath.Join(i.Root, harness, "auth.json"), []byte("synthetic-auth-canary"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	config := filepath.Join(i.Root, "codex/config.toml")
	old := filepath.Join(t.TempDir(), "old-config")
	if err := os.Link(config, old); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(old, []byte("old generated config"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(i.Root, "pi/extensions/my-friday.ts")); err != nil {
		t.Fatal(err)
	}
	if err := i.Project(s); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(old)
	if string(data) != "old generated config" {
		t.Fatal("refresh truncated the old inode")
	}
	for _, harness := range []string{"codex", "pi"} {
		data, err := os.ReadFile(filepath.Join(i.Root, harness, "auth.json"))
		if err != nil || string(data) != "synthetic-auth-canary" {
			t.Fatal("native auth changed")
		}
	}
	data, err = os.ReadFile(filepath.Join(i.Root, "pi/extensions/my-friday.ts"))
	if err != nil || !strings.Contains(string(data), "before_agent_start") {
		t.Fatal("missing projection not repaired")
	}
}
