package portable

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoteSetupResumesWithoutChangingIdentity(t *testing.T) {
	s := fixtureStore(t)
	ctx := context.Background()
	if err := s.InitGit(ctx); err != nil {
		t.Fatal(err)
	}
	before := strings.TrimSpace(gitTest(t, s.Root, "rev-parse", "HEAD"))
	remote := filepath.Join(t.TempDir(), "source.git")
	gitTest(t, "", "init", "--bare", "--initial-branch=main", remote)
	if err := s.ConnectRemote(ctx, remote); err != nil {
		t.Fatal(err)
	}
	if status, err := s.Sync(ctx); err != nil || status.State != "synced" {
		t.Fatalf("%+v %v", status, err)
	}
	if err := s.ConnectRemote(ctx, remote); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(gitTest(t, s.Root, "rev-parse", "HEAD")); got != before {
		t.Fatal("resume changed source")
	}
	if err := s.ConnectRemote(ctx, filepath.Join(t.TempDir(), "different.git")); err == nil {
		t.Fatal("replaced origin")
	}
}

func TestRemoteSetupRejectsForeignAndDirtySources(t *testing.T) {
	ctx := context.Background()
	a, _, remote := syncedPair(t)
	b := fixtureStore(t)
	if err := b.InitGit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := b.ConnectRemote(ctx, remote); err == nil {
		t.Fatal("attached foreign assistant")
	}
	if got := gitTest(t, b.Root, "remote"); got != "" {
		t.Fatal("failed attach left origin")
	}
	if err := os.WriteFile(filepath.Join(a.Root, "instructions", "draft.md"), []byte("unfinished"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := a.RemoteSetupStatus(ctx); err == nil {
		t.Fatal("dirty source accepted")
	}
}

func TestRemoteSetupRejectsNoMainAndUnsafeURLs(t *testing.T) {
	ctx := context.Background()
	s := fixtureStore(t)
	if err := s.InitGit(ctx); err != nil {
		t.Fatal(err)
	}
	for _, remote := range []string{"https://token@github.com/owner/repo", "https://github.com/owner/repo?secret=x", "git@github.com:owner/repo", "-upload-pack=evil"} {
		if err := s.ConnectRemote(ctx, remote); err == nil {
			t.Fatalf("accepted %s", remote)
		}
	}
	remote := filepath.Join(t.TempDir(), "only-other.git")
	gitTest(t, "", "init", "--bare", remote)
	gitTest(t, s.Root, "push", remote, "HEAD:refs/heads/other")
	if err := s.ConnectRemote(ctx, remote); err == nil {
		t.Fatal("nonempty remote without main accepted")
	}
}

func TestRemoteSetupRejectsAlternatePushAndFetchTargets(t *testing.T) {
	for _, setting := range []string{"remote.origin.pushurl", "remote.origin.mirror", "remote.origin.fetch"} {
		t.Run(setting, func(t *testing.T) {
			s := fixtureStore(t)
			ctx := context.Background()
			if err := s.InitGit(ctx); err != nil {
				t.Fatal(err)
			}
			gitTest(t, s.Root, "remote", "add", "origin", filepath.Join(t.TempDir(), "remote.git"))
			value := "unexpected"
			if setting == "remote.origin.mirror" {
				value = "true"
			}
			gitTest(t, s.Root, "config", setting, value)
			if _, err := s.RemoteSetupStatus(ctx); err == nil {
				t.Fatal("unsafe origin configuration accepted")
			}
		})
	}
}
