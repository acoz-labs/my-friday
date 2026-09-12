package portable

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMemoryBankNeedsNoAssistantInstallation(t *testing.T) {
	s, err := CreateMemoryBank(filepath.Join(t.TempDir(), "bank with spaces"), "Personal memory", "device-laptop", "Laptop")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"agent.json", "instructions", "capabilities", "integrations", "codex", "pi"} {
		if _, err := os.Lstat(filepath.Join(s.Root, forbidden)); !os.IsNotExist(err) {
			t.Fatalf("bank requires assistant component %s", forbidden)
		}
	}
	fresh, err := Open(s.Root)
	if err != nil || !fresh.IsMemoryBank() || fresh.Agent.ID != s.Agent.ID || fresh.Agent.DefaultHarness != "" {
		t.Fatalf("bank identity: %+v %v", fresh, err)
	}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	r := revision("revision-bank")
	if err := s.Put(r); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RecordEvent("decision", "Selected the document policy.", r.Authorship); err != nil {
		t.Fatal(err)
	}
	entries, err := s.Journal("document", 10)
	if err != nil || len(entries) != 1 || entries[0].Authorship.DeviceID != "device-laptop" {
		t.Fatalf("journal: %+v %v", entries, err)
	}
	if _, err := s.Sync(context.Background()); err == nil {
		t.Fatal("sync accepted an uninitialized bank")
	}
	if _, err := os.Lstat(filepath.Join(s.Root, ".git")); !os.IsNotExist(err) {
		t.Fatal("sync initialized Git implicitly")
	}
	if err := s.InitGit(context.Background()); err != nil {
		t.Fatal(err)
	}
	status, err := s.Sync(context.Background())
	if err != nil || status.State != "local-only" {
		t.Fatalf("local bank: %+v %v", status, err)
	}
	if _, err := CreateMemoryBank(s.Root, "Overwrite", "device-laptop", "Laptop"); err == nil {
		t.Fatal("overwrote bank")
	}
	if _, err := Bind(s, filepath.Join(t.TempDir(), "instance"), "example", "/bin/true", "device-laptop"); err == nil {
		t.Fatal("memory bank accepted as assistant")
	}
}

func TestMemoryBankCloneCorrectionAndConcurrentKnowledge(t *testing.T) {
	ctx := context.Background()
	a, err := CreateMemoryBank(filepath.Join(t.TempDir(), "bank"), "Shared memory", "device-laptop", "Laptop")
	if err != nil {
		t.Fatal(err)
	}
	if err := a.InitGit(ctx); err != nil {
		t.Fatal(err)
	}
	remote := filepath.Join(t.TempDir(), "remote.git")
	gitTest(t, "", "init", "--bare", "--initial-branch=main", remote)
	gitTest(t, a.Root, "remote", "add", "origin", remote)
	first := revision("revision-first")
	if err := a.Put(first); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	clone := filepath.Join(t.TempDir(), "second-machine")
	gitTest(t, "", "clone", remote, clone)
	b, err := Open(clone)
	if err != nil {
		t.Fatal(err)
	}
	if err := b.AddDevice(Device{Version: 1, ID: "device-second", Label: "Second machine"}); err != nil {
		t.Fatal(err)
	}
	correction := revision("revision-corrected", first.ID)
	correction.Body = "Move drafts into Pending instead of Review."
	correction.Authorship.DeviceID = "device-second"
	if err := b.Put(correction); err != nil {
		t.Fatal(err)
	}
	for _, s := range []*Store{b, a} {
		if status, err := s.Sync(ctx); err != nil || status.State != "synced" {
			t.Fatalf("sync: %+v %v", status, err)
		}
	}
	packet, err := a.Recall(Query{Text: "drafts", Scope: first.Scope}, time.Now())
	if err != nil || len(packet.Current) != 1 || packet.Current[0].ID != correction.ID {
		t.Fatalf("correction lost: %+v %v", packet, err)
	}
	history, err := a.History(first.RecordID)
	if err != nil || len(history) != 2 || history[0].Authorship.DeviceID != "device-laptop" || history[1].Authorship.DeviceID != "device-second" {
		t.Fatalf("origin changed: %+v %v", history, err)
	}
	// Concurrent corrections remain conflicts; later wall time cannot pick truth.
	left, right := revision("revision-left", correction.ID), revision("revision-right", correction.ID)
	if err := a.Put(left); err != nil {
		t.Fatal(err)
	}
	if err := b.Put(right); err != nil {
		t.Fatal(err)
	}
	for _, s := range []*Store{a, b} {
		if _, err := s.Sync(ctx); err != nil {
			t.Fatal(err)
		}
	}
	packet, err = b.Recall(Query{Text: "drafts", Scope: first.Scope}, time.Now())
	if err != nil || len(packet.Current) != 0 || len(packet.Conflicts) != 1 {
		t.Fatalf("concurrent decisions became guidance: %+v %v", packet, err)
	}
}

func TestMemoryBankRefusesAmbiguousFormatAndUnexpectedSource(t *testing.T) {
	s, err := CreateMemoryBank(filepath.Join(t.TempDir(), "bank"), "Memory", "device-laptop", "Laptop")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.Root, "agent.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(s.Root); err == nil {
		t.Fatal("ambiguous bank/assistant format accepted")
	}
	if err := os.Remove(filepath.Join(s.Root, "agent.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.Root, "unrelated-data"), []byte("preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.InitGit(context.Background()); err == nil {
		t.Fatal("unknown bank content would be committed")
	}
	data, _ := os.ReadFile(filepath.Join(s.Root, "unrelated-data"))
	if string(data) != "preserve" {
		t.Fatal("unknown content changed")
	}
}
