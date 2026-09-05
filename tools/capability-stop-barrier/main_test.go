package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestStablePostMutationRequiresOwnedCanonicalJournalAndAbsentProjection(t *testing.T) {
	root := t.TempDir()
	control := filepath.Join(root, "capabilities", "daily-brief")
	if err := os.MkdirAll(control, 0o700); err != nil {
		t.Fatal(err)
	}
	r := map[string]any{"contract_version": 1, "slug": "daily-brief", "version": "1.0.0", "source_path": "/source", "source_digest": fmt.Sprintf("%064x", 1), "projection_digest": fmt.Sprintf("%064x", 2), "state": "installed-healthy", "generation": 1, "files": []string{"SKILL.md"}, "generations": []string{fmt.Sprintf("%064x", 2)}}
	rb, _ := json.MarshalIndent(r, "", "  ")
	rb = append(rb, '\n')
	compact, _ := json.Marshal(r)
	prior := sha256.Sum256(compact)
	tx := transaction{ContractVersion: 1, Action: "disable", Slug: "daily-brief", SourceDigest: r["source_digest"].(string), ProjectionDigest: r["projection_digest"].(string), PriorReceiptDigest: fmt.Sprintf("%x", prior)}
	tb, _ := json.MarshalIndent(tx, "", "  ")
	tb = append(tb, '\n')
	if err := os.WriteFile(filepath.Join(control, "receipt.json"), rb, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(control, "transaction.json"), tb, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, ok := stablePostMutation(root, "daily-brief"); !ok {
		t.Fatal("valid post-mutation state refused")
	}
	projection := filepath.Join(root, "workspace", ".agents", "skills", "daily-brief")
	if err := os.MkdirAll(projection, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, ok := stablePostMutation(root, "daily-brief"); ok {
		t.Fatal("pre-mutation projection accepted")
	}
	if err := os.RemoveAll(filepath.Join(root, "workspace")); err != nil {
		t.Fatal(err)
	}
	tb[0] = ' '
	if err := os.WriteFile(filepath.Join(control, "transaction.json"), tb, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, ok := stablePostMutation(root, "daily-brief"); ok {
		t.Fatal("noncanonical journal accepted")
	}
}

func TestReadStoppedPIDRejectsInvalidMarker(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "candidate-stopped")
	if err := os.WriteFile(marker, []byte("not-a-pid\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readStoppedPID(marker, os.Getpid()); err == nil {
		t.Fatal("accepted invalid stopped-capability marker")
	}
}
