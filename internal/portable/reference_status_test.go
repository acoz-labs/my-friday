package portable

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReferenceAvailabilityReadOnlyAndRecovery(t *testing.T) {
	s, i, root := referenceFixture(t)
	check := func(want string) {
		t.Helper()
		status, err := i.CheckReference(s, "old-notes")
		if err != nil || status.State != want {
			t.Fatalf("%+v %v; want %s", status, err, want)
		}
	}
	check("available")
	p := filepath.Join(i.Root, "references/old-notes.json")
	before, _ := os.ReadFile(p)
	check("available")
	after, _ := os.ReadFile(p)
	if string(before) != string(after) {
		t.Fatal("check changed binding")
	}
	if err := os.Rename(root, root+"-moved"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Rename(root+"-moved", root) })
	check("unavailable")
	if err := i.BindReference(s, "old-notes", root+"-moved"); err != nil {
		t.Fatal(err)
	}
	check("available")
	libs, _ := s.ReferenceLibraries()
	reviewed := libs[0]
	libs[0].Purpose = "Revised purpose requiring review"
	data, _ := json.Marshal(libs[0])
	if err := os.WriteFile(filepath.Join(s.Root, ".my-friday/references/old-notes.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	check("stale")
	if err := i.BindReference(s, "old-notes", root+"-moved", reviewed); err == nil {
		t.Fatal("unseen description acknowledged")
	}
	check("stale")
	if err := i.BindReference(s, "old-notes", root+"-moved"); err != nil {
		t.Fatal(err)
	}
	check("available")
	if err := os.WriteFile(p, []byte("bad json"), 0600); err != nil {
		t.Fatal(err)
	}
	check("invalid")
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	check("unbound")
}
