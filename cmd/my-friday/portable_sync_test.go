package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/portable"
)

// Exercise setup/import and write/checkpoint through the command layer, not
// just two Store objects. All repositories, identities and outages are local
// synthetic fixtures; this is not an authenticated-network acceptance test.
func TestPortableImportOfflineRecoveryAcrossDevices(t *testing.T) {
	base := t.TempDir()
	a, b := filepath.Join(base, "first agent"), filepath.Join(base, "second agent")
	sa, sb := filepath.Join(base, "first instance"), filepath.Join(base, "second instance")
	remote := filepath.Join(base, "remote.git")
	t.Chdir(t.TempDir()) // Caller cwd must never select the assistant repository.
	call := func(input string, target any, args ...string) {
		t.Helper()
		var out, errs bytes.Buffer
		if err := runPortable(args, strings.NewReader(input), &out, &errs); err != nil {
			t.Fatalf("%v: %v; %s", args, err, errs.String())
		}
		if target != nil {
			if err := json.Unmarshal(out.Bytes(), target); err != nil {
				t.Fatalf("%v output: %s; %v", args, out.String(), err)
			}
		}
	}
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "core.hooksPath=/dev/null"}, args...)...)
		for _, env := range os.Environ() {
			if !strings.HasPrefix(env, "GIT_") && !strings.HasPrefix(env, "SSH_ASKPASS=") {
				cmd.Env = append(cmd.Env, env)
			}
		}
		cmd.Env = append(cmd.Env, "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v; %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	sync := func(root, want string) {
		t.Helper()
		var status portable.SyncStatus
		call("", &status, "sync", "--repository", root)
		if status.State != want {
			t.Fatalf("sync %s: %+v, want %s", root, status, want)
		}
	}
	call("", nil, "setup", "--repository", a, "--state", sa, "--name", "sync-pilot", "--device-label", "First fixture device", "--no-launcher")
	ia, first, err := portable.LoadInstance(sa)
	if err != nil {
		t.Fatal(err)
	}
	git("init", "--bare", "--initial-branch=main", "--template=", remote)
	git("-C", a, "remote", "add", "origin", remote)
	sync(a, "synced")
	var original portable.Revision
	call("", &original, "memory", "template", "--repository", a, "--device", ia.DeviceID, "--scope-kind", "project", "--scope-id", "project-example")
	original.Summary, original.Body = "Synthetic project is named First", "First is the fictional project name."
	write := func(root, device, harness string, r portable.Revision, want string) {
		t.Helper()
		t.Setenv("MY_FRIDAY_HARNESS", harness)
		data, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		var result struct {
			Sync portable.SyncStatus `json:"sync"`
		}
		call(string(data), &result, "memory", "write", "--repository", root, "--device", device, "--input", "-")
		if result.Sync.State != want || result.Sync.Head == "" {
			t.Fatalf("write lost checkpoint/status: %+v", result)
		}
	}
	write(a, ia.DeviceID, "codex", original, "synced")
	git("clone", "--template=", remote, b)
	call("", nil, "setup", "--import", b, "--state", sb, "--name", "sync-pilot", "--device-label", "Second fixture device", "--no-launcher")
	ib, second, err := portable.LoadInstance(sb)
	if err != nil {
		t.Fatal(err)
	}
	if ia.AssistantID != ib.AssistantID || ia.DeviceID == ib.DeviceID || ia.Repository == ib.Repository || ia.Root == ib.Root {
		t.Fatal("import did not preserve agent identity with distinct device/source/state bindings")
	}
	git("-C", b, "remote", "set-url", "origin", filepath.Join(base, "unavailable.git"))
	updated := original
	updated.ID = portable.NewID("revision")
	updated.Supersedes = []string{original.ID}
	updated.Summary, updated.Body = "Synthetic project is named Second", "Second replaces First; preserve the old name as history."
	write(b, ib.DeviceID, "pi", updated, "pending")
	if git("-C", b, "status", "--porcelain") != "" {
		t.Fatal("offline write was not committed")
	}
	// Create independent online work so recovery must merge divergent histories.
	t.Setenv("MY_FRIDAY_HARNESS", "codex")
	var online struct {
		Event portable.JournalEntry `json:"event"`
	}
	call("", &online, "memory", "event", "--repository", a, "--device", ia.DeviceID, "--kind", "test-checkpoint", "--summary", "Independent synthetic online work")
	git("-C", b, "remote", "set-url", "origin", remote)
	sync(b, "synced")
	sync(a, "synced")
	for _, store := range []*portable.Store{first, second} {
		if err := store.Validate(); err != nil {
			t.Fatal(err)
		}
		var packet portable.Packet
		call("", &packet, "memory", "recall", "--repository", store.Root, "--scope-kind", original.Scope.Kind, "--scope-id", original.Scope.ID)
		if len(packet.Current) != 1 || packet.Current[0].ID != updated.ID || len(packet.Conflicts) != 0 {
			t.Fatalf("recovery did not retain the current correction: %+v", packet)
		}
		history, err := store.History(original.RecordID)
		if err != nil || len(history) != 2 || history[0].Authorship.DeviceID != ia.DeviceID || history[1].Authorship.DeviceID != ib.DeviceID || history[1].Authorship.Harness != "pi" {
			t.Fatalf("cross-device provenance/history lost: %+v %v", history, err)
		}
		for _, id := range []string{ia.DeviceID, ib.DeviceID} {
			if _, err := os.Stat(filepath.Join(store.Root, "provenance/devices", id+".json")); err != nil {
				t.Fatal("device registration did not synchronize:", err)
			}
		}
		entries, err := filepath.Glob(filepath.Join(store.Root, "memory/events/*/*", online.Event.ID+".json"))
		if err != nil || len(entries) != 1 {
			t.Fatalf("independent online journal entry lost: %v %v", entries, err)
		}
		data, err := os.ReadFile(entries[0])
		if err != nil {
			t.Fatal(err)
		}
		var entry portable.JournalEntry
		if err := json.Unmarshal(data, &entry); err != nil || entry.Summary != online.Event.Summary || entry.Authorship.DeviceID != ia.DeviceID || entry.Authorship.Harness != "codex" {
			t.Fatalf("independent online work changed: %+v %v", entry, err)
		}
		if git("-C", store.Root, "status", "--porcelain") != "" {
			t.Fatal("recovery left a dirty worktree")
		}
	}
	head := git("-C", a, "rev-parse", "HEAD")
	if git("-C", b, "rev-parse", "HEAD") != head || git("--git-dir", remote, "rev-parse", "main") != head {
		t.Fatal("installations and remote did not converge")
	}
}
