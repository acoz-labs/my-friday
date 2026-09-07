package portable

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func gitTest(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = cleanGitEnvironment()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}

func syncedPair(t *testing.T) (*Store, *Store, string) {
	t.Helper()
	a := fixtureStore(t)
	ctx := context.Background()
	if err := a.InitGit(ctx); err != nil {
		t.Fatal(err)
	}
	bare := filepath.Join(t.TempDir(), "remote.git")
	gitTest(t, "", "init", "--bare", "--initial-branch=main", bare)
	gitTest(t, a.Root, "remote", "add", "origin", bare)
	if state, err := a.Sync(ctx); err != nil || state.State != "synced" {
		t.Fatalf("initial sync: %+v %v", state, err)
	}
	other := filepath.Join(t.TempDir(), "other")
	gitTest(t, "", "clone", bare, other)
	b, err := Open(other)
	if err != nil {
		t.Fatal(err)
	}
	if err = b.Validate(); err != nil {
		t.Fatal(err)
	}
	return a, b, bare
}

func TestSyncIndependentOfflineWrites(t *testing.T) {
	a, b, _ := syncedPair(t)
	ctx := context.Background()
	first := revision("revision-initial")
	if err := a.Put(first); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	second := revision("revision-other")
	second.RecordID = "record-other"
	if err := b.Put(second); err != nil {
		t.Fatal(err)
	}
	if state, err := b.Sync(ctx); err != nil || state.State != "synced" {
		t.Fatalf("merge independent additions: %+v %v", state, err)
	}
	if _, err := a.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	p, err := a.Recall(Query{Text: "drafts", Scope: first.Scope}, time.Now())
	if err != nil || len(p.Current) != 2 {
		t.Fatalf("lost writes: %+v %v", p, err)
	}
}

func TestSyncRefusesRewritingPublishedRevision(t *testing.T) {
	a, b, _ := syncedPair(t)
	ctx := context.Background()
	r := revision("revision-published")
	r.Supersedes = []string{}
	if err := a.Put(r); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	r.Body = "A silent replacement is not a supersession."
	data, _ := json.Marshal(r)
	path := filepath.Join(b.Root, "memory/records", r.RecordID, r.ID+".json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Sync(ctx); err == nil {
		t.Fatal("checkpoint accepted rewritten history")
	}
	// A remote edited outside My Friday must be rejected as well.
	gitTest(t, b.Root, "add", ".")
	gitTest(t, b.Root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.test", "commit", "-m", "rewrite")
	gitTest(t, b.Root, "push", "origin", "main")
	if state, err := a.Sync(ctx); err != nil || state.State != "conflict" {
		t.Fatalf("remote rewrite accepted: %+v %v", state, err)
	}
	history, err := a.History(r.RecordID)
	if err != nil || len(history) != 1 || history[0].Body == r.Body {
		t.Fatalf("local history changed: %+v %v", history, err)
	}
}

func TestSyncRetainsSemanticConflict(t *testing.T) {
	a, b, _ := syncedPair(t)
	ctx := context.Background()
	root := revision("revision-root")
	if err := a.Put(root); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	for i, s := range []*Store{a, b} {
		r := revision([]string{"revision-left", "revision-right"}[i], root.ID)
		if err := s.Put(r); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := a.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	p, err := b.Recall(Query{Scope: root.Scope}, time.Now())
	if err != nil || len(p.Conflicts) != 1 || len(p.Current) != 0 {
		t.Fatalf("conflict discarded: %+v %v", p, err)
	}
}

func TestOfflineSyncCommitsAndReturnsPending(t *testing.T) {
	s := fixtureStore(t)
	ctx := context.Background()
	if err := s.InitGit(ctx); err != nil {
		t.Fatal(err)
	}
	gitTest(t, s.Root, "remote", "add", "origin", filepath.Join(t.TempDir(), "unavailable.git"))
	if err := s.Put(revision("revision-offline")); err != nil {
		t.Fatal(err)
	}
	state, err := s.Sync(ctx)
	if err != nil || state.State != "pending" {
		t.Fatalf("offline failed local work: %+v %v", state, err)
	}
	if out := gitTest(t, s.Root, "status", "--porcelain"); out != "" {
		t.Fatalf("not committed: %s", out)
	}
}

func TestSyncDoesNotTouchCallerGitRepository(t *testing.T) {
	s := fixtureStore(t)
	foreign := t.TempDir()
	gitTest(t, foreign, "init", "--initial-branch=main")
	canary := filepath.Join(foreign, "untouched")
	if err := os.WriteFile(canary, []byte("canary"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_DIR", filepath.Join(foreign, ".git"))
	t.Setenv("GIT_WORK_TREE", foreign)
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "user.name")
	t.Setenv("GIT_CONFIG_VALUE_0", "Wrong actor")
	if err := s.InitGit(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if out := gitTest(t, foreign, "status", "--porcelain"); out != "?? untouched\n" {
		t.Fatalf("foreign repo changed: %q", out)
	}
}

func TestSyncConflictingConfigurationPreservesBothCommits(t *testing.T) {
	a, b, _ := syncedPair(t)
	ctx := context.Background()
	for i, s := range []*Store{a, b} {
		if err := os.WriteFile(filepath.Join(s.Root, "instructions/identity.md"), []byte([]string{"Left identity\n", "Right identity\n"}[i]), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := a.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	state, err := b.Sync(ctx)
	if err != nil || state.State != "conflict" {
		t.Fatalf("configuration conflict: %+v %v", state, err)
	}
	data, _ := os.ReadFile(filepath.Join(b.Root, "instructions/identity.md"))
	if string(data) != "Right identity\n" {
		t.Fatalf("local configuration altered: %s", data)
	}
	if out := gitTest(t, b.Root, "status", "--porcelain"); out != "" {
		t.Fatalf("left merge in progress: %s", out)
	}
}
