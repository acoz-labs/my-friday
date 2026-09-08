package portable

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type SyncStatus struct {
	State      string `json:"state"`
	Head       string `json:"head"`
	RemoteHead string `json:"remote_head,omitempty"`
	CheckedAt  string `json:"checked_at"`
	Detail     string `json:"detail,omitempty"`
}

func cleanGitEnvironment() []string {
	env := []string{}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "GIT_") || key == "SSH_ASKPASS" {
			continue
		}
		env = append(env, entry)
	}
	return append(env, "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0")
}

func (s *Store) git(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", append(s.gitArguments(), args...)...)
	cmd.Env = cleanGitEnvironment()
	configureCommandCancellation(cmd)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("assistant Git operation %s failed: %w", args[0], err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (s *Store) InitGit(ctx context.Context) error {
	work := *s
	s = &work
	return s.withLock(func() error {
		config, err := s.syncConfiguration()
		if err != nil {
			return err
		}
		s.gitSettings = &config
		if _, err := os.Lstat(filepath.Join(s.Root, ".git")); os.IsNotExist(err) {
			if _, err = s.git(ctx, "init", "--initial-branch=main", "--template="); err != nil {
				return err
			}
		}
		return s.checkpoint(ctx)
	})
}

func (s *Store) gitBoundary(ctx context.Context) error {
	info, err := os.Lstat(filepath.Join(s.Root, ".git"))
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("assistant requires its own .git directory")
	}
	top, err := s.git(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return err
	}
	a, _ := filepath.EvalSymlinks(top)
	b, _ := filepath.EvalSymlinks(s.Root)
	if a != b {
		return errors.New("Git root differs from assistant root")
	}
	branch, err := s.git(ctx, "symbolic-ref", "--short", "HEAD")
	if err != nil || branch != "main" {
		return errors.New("assistant synchronization requires its main branch")
	}
	for _, state := range []string{"MERGE_HEAD", "CHERRY_PICK_HEAD", "rebase-merge", "rebase-apply"} {
		if _, err := os.Lstat(filepath.Join(s.Root, ".git", state)); err == nil {
			return errors.New("finish the existing Git operation before synchronization")
		}
	}
	return nil
}

func (s *Store) checkpoint(ctx context.Context) error {
	if err := s.gitBoundary(ctx); err != nil {
		return err
	}
	if err := s.Validate(); err != nil {
		return err
	}
	if err := s.validateCheckpointObserver(s.checkpointObserver); err != nil {
		return err
	}
	base, headErr := s.git(ctx, "rev-parse", "--verify", "HEAD")
	if headErr == nil {
		if err := s.validateAppendOnly(ctx, "HEAD"); err != nil {
			return err
		}
	}
	if _, err := s.git(ctx, "add", "--all", "--", "."); err != nil {
		return err
	}
	status, err := s.git(ctx, "status", "--porcelain")
	if err != nil {
		return err
	}
	if status == "" {
		return nil
	}
	if err := s.captureSourceChanges(ctx, base); err != nil {
		return err
	}
	_, err = s.git(ctx, "commit", "-m", "Record assistant changes")
	return err
}

// Existing durable evidence is never edited or deleted by automatic sync.
// Corrections must append a revision that explicitly supersedes its predecessor.
func (s *Store) validateAppendOnly(ctx context.Context, revisions ...string) error {
	args := append([]string{"diff", "--name-only", "--no-renames", "--diff-filter=DMRTUXB"}, revisions...)
	args = append(args, "--", "memory", "provenance")
	changed, err := s.git(ctx, args...)
	if err != nil {
		return err
	}
	if changed != "" {
		return errors.New("durable evidence was modified or deleted; append a superseding revision instead")
	}
	return nil
}

// Sync reconciles append-only records without resolving semantic disagreements.
// Network/authentication failures are pending status, never loss of local work.
func (s *Store) Sync(parent context.Context) (SyncStatus, error) {
	work := *s
	s = &work
	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()
	result := SyncStatus{CheckedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	err := s.withLock(func() error {
		config, err := s.syncConfiguration()
		if err != nil {
			return err
		}
		s.gitSettings = &config
		if err := s.checkpoint(ctx); err != nil {
			return err
		}
		result.Head, err = s.git(ctx, "rev-parse", "HEAD")
		if err != nil {
			return err
		}
		remote, err := s.git(ctx, "remote", "get-url", "origin")
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			result.State = "local-only"
			return nil
		}
		if !filepath.IsAbs(remote) && !strings.HasPrefix(remote, "file://") {
			parsed, parseErr := url.Parse(remote)
			if parseErr != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
				result.State = "pending"
				result.Detail = "This transport is not configured; use a local remote or HTTPS with a private credential helper."
				return nil
			}
			if len(config.CredentialHelper) == 0 {
				result.State = "pending"
				result.Detail = "Configure a private credential helper in .my-friday/sync.json; local changes are committed."
				return nil
			}
		}
		return s.syncRemote(ctx, &result)
	})
	return result, err
}

func (s *Store) syncRemote(ctx context.Context, result *SyncStatus) error {
	var err error
	if _, err = s.git(ctx, "fetch", "--prune", "origin"); err != nil {
		result.State = "pending"
		result.Detail = "Remote unavailable; local changes are committed."
		return nil
	}
	remoteHead, err := s.git(ctx, "rev-parse", "--verify", "refs/remotes/origin/main")
	if err == nil {
		result.RemoteHead = remoteHead
		if remoteHead != result.Head {
			if _, err = s.git(ctx, "merge-base", "--is-ancestor", remoteHead, result.Head); err != nil {
				candidate := remoteHead
				if _, err = s.git(ctx, "merge-base", "--is-ancestor", result.Head, remoteHead); err != nil {
					tree, mergeErr := s.git(ctx, "merge-tree", "--write-tree", result.Head, remoteHead)
					if mergeErr != nil {
						result.State = "conflict"
						result.Detail = "Concurrent file changes require reconciliation; both commits are preserved."
						return nil
					}
					candidate, err = s.git(ctx, "commit-tree", tree, "-p", result.Head, "-p", remoteHead, "-m", "Reconcile assistant changes")
					if err != nil {
						return err
					}
				}
				if err = s.validateCandidate(ctx, candidate); err != nil {
					result.State = "conflict"
					result.Detail = "Remote candidate failed assistant validation; local state is preserved."
					return nil
				}
				if _, err = s.git(ctx, "merge", "--ff-only", candidate); err != nil {
					return err
				}
				result.Head = candidate
			}
		}
	}
	if _, err = s.git(ctx, "push", "origin", "HEAD:refs/heads/main"); err != nil {
		result.State = "pending"
		result.Detail = "Push not confirmed; local history is preserved. Retry synchronization."
		return nil
	}
	result.State = "synced"
	result.RemoteHead = result.Head
	return nil
}

func (s *Store) validateCandidate(ctx context.Context, commit string) error {
	if err := s.validateAppendOnly(ctx, "HEAD", commit); err != nil {
		return err
	}
	parent, err := os.MkdirTemp("", "my-friday-verify-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(parent)
	candidate := filepath.Join(parent, "candidate")
	if _, err = s.git(ctx, "worktree", "add", "--detach", candidate, commit); err != nil {
		return err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.git(cleanup, "worktree", "remove", "--force", candidate)
	}()
	other, err := Open(candidate)
	if err != nil {
		return err
	}
	if other.Agent.ID != s.Agent.ID {
		return errors.New("remote changed assistant identity")
	}
	return other.Validate()
}
