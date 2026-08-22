package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidJournalRequiresMatchingRecoverableTransaction(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "transaction.json")
	valid := `{"contract_version":1,"action":"upgrade","phase":"mutating","root":{"device":1,"inode":2},"before":{"projection":{},"manifest":{},"canonical":{},"previous":{}},"after":{"projection":{},"manifest":{},"canonical":{},"previous":{}}}`
	if err := os.WriteFile(path, []byte(valid), 0o600); err != nil {
		t.Fatal(err)
	}
	if !validJournal([]string{path}, "upgrade") {
		t.Fatal("matching durable journal was not accepted")
	}
	if validJournal([]string{path}, "install") {
		t.Fatal("journal for another action was accepted")
	}
	committed := []byte(`{"contract_version":1,"action":"upgrade","phase":"committed","root":{"device":1,"inode":2},"before":{"projection":{},"manifest":{},"canonical":{},"previous":{}},"after":{"projection":{},"manifest":{},"canonical":{},"previous":{}}}`)
	if err := os.WriteFile(path, committed, 0o600); err != nil {
		t.Fatal(err)
	}
	if validJournal([]string{path}, "upgrade") {
		t.Fatal("committed journal was accepted as an interruption point")
	}
}
