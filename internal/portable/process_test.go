package portable

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGitDeadlineCancelsOrdinaryDescendants(t *testing.T) {
	s := fixtureStore(t)
	bin := t.TempDir()
	leak := filepath.Join(t.TempDir(), "late-effect")
	script := "#!/bin/sh\n(sleep 1; printf unexpected > " + shellQuote(leak) + ") &\nwait\n"
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := s.git(ctx, "fetch"); err == nil {
		t.Fatal("cancelled Git command reported success")
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("Git descendants kept pipes open past cancellation")
	}
	time.Sleep(1100 * time.Millisecond)
	if _, err := os.Stat(leak); !os.IsNotExist(err) {
		t.Fatal("Git child survived deadline and performed a later effect")
	}
}
