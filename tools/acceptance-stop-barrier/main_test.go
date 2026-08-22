package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func proof(body []byte) fileProof {
	return fileProof{Exists: true, SHA256: fmt.Sprintf("%x", sha256.Sum256(body)), Bytes: body}
}

func generation(t *testing.T, projection []byte, previous fileProof) map[string]fileProof {
	t.Helper()
	p := proof(projection)
	m := installedManifest{ContractVersion: 1, AssistantID: "assistant", ProjectionSHA256: p.SHA256, SourcePath: "/runtime", SourceSHA256: strings.Repeat("a", 64), CanonicalSHA256: p.SHA256, PreviousSHA256: previous.SHA256}
	body, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	body = append(body, '\n')
	return map[string]fileProof{"projection": p, "manifest": proof(body), "canonical": p, "previous": previous}
}

func writeProof(t *testing.T, path string, p fileProof) {
	t.Helper()
	if p.Exists {
		if err := os.WriteFile(path, p.Bytes, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func journalFixture(t *testing.T) (string, string, journal) {
	t.Helper()
	home := t.TempDir()
	control := filepath.Join(home, ".my-friday")
	if err := os.Mkdir(control, 0o700); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(home)
	stat := info.Sys().(*syscall.Stat_t)
	before := generation(t, []byte("before\n"), fileProof{})
	after := generation(t, []byte("after\n"), before["projection"])
	j := journal{ContractVersion: 1, Action: "upgrade", Phase: "mutating", Root: rootProof{Device: uint64(stat.Dev), Inode: stat.Ino}, Before: before, After: after}
	writeProof(t, filepath.Join(home, "AGENTS.md"), before["projection"])
	writeProof(t, filepath.Join(control, "installed-baseline.json"), before["manifest"])
	writeProof(t, filepath.Join(control, "canonical-AGENTS.md"), before["canonical"])
	path := filepath.Join(control, "transaction.json")
	writeJournal(t, path, j)
	return home, path, j
}

func writeJournal(t *testing.T, path string, j journal) {
	t.Helper()
	body, err := json.Marshal(j)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, body, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestValidJournalRequiresFilesystemLinkedRecoverableTransaction(t *testing.T) {
	home, path, j := journalFixture(t)
	if !validJournal([]string{path}, "upgrade", home) {
		t.Fatal("matching durable journal was not accepted")
	}
	if validJournal([]string{path}, "install", home) {
		t.Fatal("journal for another action was accepted")
	}
	j.Phase = "committed"
	writeJournal(t, path, j)
	if validJournal([]string{path}, "upgrade", home) {
		t.Fatal("committed journal was accepted")
	}
}

func TestJournalRejectsEmbeddedProofWithoutMatchingFilesystemState(t *testing.T) {
	home, path, _ := journalFixture(t)
	if err := os.WriteFile(filepath.Join(home, "AGENTS.md"), []byte("foreign\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if validJournal([]string{path}, "upgrade", home) {
		t.Fatal("journal accepted embedded bytes that do not match the projection filesystem")
	}
}

func TestJournalRejectsMalformedAndForeignStaging(t *testing.T) {
	home, path, j := journalFixture(t)
	j.Root.Inode++
	writeJournal(t, path, j)
	if validJournal([]string{path}, "upgrade", home) {
		t.Fatal("wrong-root journal was accepted")
	}
	home, path, _ = journalFixture(t)
	if err := os.WriteFile(filepath.Join(home, ".my-friday", "foreign.next"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if validJournal([]string{path}, "upgrade", home) {
		t.Fatal("foreign staging file was accepted")
	}
}

func TestJournalRejectsUnknownFieldsAndManifestSemanticForgery(t *testing.T) {
	home, path, j := journalFixture(t)
	body, _ := json.Marshal(j)
	body = []byte(strings.Replace(string(body), `"contract_version":1`, `"contract_version":1,"extra":true`, 1))
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatal(err)
	}
	if validJournal([]string{path}, "upgrade", home) {
		t.Fatal("unknown journal field was accepted")
	}
	home, path, j = journalFixture(t)
	manifest := j.Before["manifest"]
	manifest.Bytes = []byte(`{"contract_version":1,"assistant_id":"assistant","projection_sha256":"bad","source_path":"/runtime","source_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","canonical_sha256":"bad"}`)
	manifest.SHA256 = fmt.Sprintf("%x", sha256.Sum256(manifest.Bytes))
	j.Before["manifest"] = manifest
	writeJournal(t, path, j)
	if validJournal([]string{path}, "upgrade", home) {
		t.Fatal("semantically forged manifest proof was accepted")
	}
}
