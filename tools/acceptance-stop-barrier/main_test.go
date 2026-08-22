package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestValidJournalRequiresMatchingRecoverableTransaction(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "transaction.json")
	info, _ := os.Stat(dir)
	stat := info.Sys().(*syscall.Stat_t)
	proof := `{"exists":true,"sha256":"2d711642b726b04401627ca9fbac32f5c8530fb1903cc4db02258717921a4881","bytes":"eA=="}`
	absent := `{"exists":false}`
	valid := fmt.Sprintf(`{"contract_version":1,"action":"upgrade","phase":"mutating","root":{"device":%d,"inode":%d},"before":{"projection":%s,"manifest":%s,"canonical":%s,"previous":%s},"after":{"projection":%s,"manifest":%s,"canonical":%s,"previous":%s}}`, stat.Dev, stat.Ino, proof, proof, proof, absent, proof, proof, proof, proof)
	if err := os.WriteFile(path, []byte(valid), 0o600); err != nil {
		t.Fatal(err)
	}
	if !validJournal([]string{path}, "upgrade", dir) {
		t.Fatal("matching durable journal was not accepted")
	}
	if validJournal([]string{path}, "install", dir) {
		t.Fatal("journal for another action was accepted")
	}
	committed := []byte(strings.Replace(valid, `"phase":"mutating"`, `"phase":"committed"`, 1))
	if err := os.WriteFile(path, committed, 0o600); err != nil {
		t.Fatal(err)
	}
	if validJournal([]string{path}, "upgrade", dir) {
		t.Fatal("committed journal was accepted as an interruption point")
	}
	for name, mutation := range map[string]string{
		"wrong root":    strings.Replace(valid, fmt.Sprintf(`"inode":%d`, stat.Ino), `"inode":999`, 1),
		"bad digest":    strings.Replace(valid, "2d711642", "0d711642", 1),
		"unknown field": strings.Replace(valid, `"contract_version":1`, `"contract_version":1,"extra":true`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(mutation), 0o600); err != nil {
				t.Fatal(err)
			}
			if validJournal([]string{path}, "upgrade", dir) {
				t.Fatal("malformed journal accepted")
			}
		})
	}
}
