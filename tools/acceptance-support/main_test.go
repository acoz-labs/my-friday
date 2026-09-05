package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/acoz-labs/my-friday/internal/assistantinstance"
	"github.com/acoz-labs/my-friday/internal/codexhome"
	bootstrap "github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/profile"
	"github.com/acoz-labs/my-friday/internal/repository"
	"golang.org/x/sys/unix"
)

const testCleanupCandidate = "0123456789abcdef0123456789abcdef01234567"
const testCleanupRunID = "receipt-test-run"

func managedNamedFixture(t *testing.T) (string, assistantinstance.Paths) {
	return managedNamedFixtureWithCodex(t, "")
}

func managedNamedFixtureWithCodex(t *testing.T, codex string) (string, assistantinstance.Paths) {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".local", "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	runtimeRoot, memoryRoot := filepath.Join(home, "source-runtime"), filepath.Join(home, "source-memory")
	p, err := profile.New("Acceptance Assistant", "", "Return only FIXTURE_PURPOSE", "concise", "")
	if err != nil {
		t.Fatal(err)
	}
	repositories, err := bootstrap.Build(p, runtimeRoot, memoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err = repository.Create(repositories, runtimeRoot, memoryRoot); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if codex == "" {
		codex = filepath.Join(home, "codex-stub")
		if err = os.WriteFile(codex, []byte("codex"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	host := filepath.Join(filepath.Dir(codex), "codex-code-mode-host")
	if _, statErr := os.Lstat(host); errors.Is(statErr, os.ErrNotExist) {
		if err = os.WriteFile(host, []byte("host"), 0o700); err != nil {
			t.Fatal(err)
		}
	} else if statErr != nil {
		t.Fatal(statErr)
	}
	instance, err := assistantinstance.PlanCreate(home, "primary", executable, codex)
	if err != nil {
		t.Fatal(err)
	}
	instance, err = assistantinstance.WithRepositories(instance, runtimeRoot, memoryRoot, repositories.AssistantID)
	if err != nil {
		t.Fatal(err)
	}
	if err = assistantinstance.Create(instance, executable, codex); err != nil {
		t.Fatal(err)
	}
	return home, instance.Paths
}

func TestCleanupNamedRemovesOnlySafeDisposableAuthBeforeInstance(t *testing.T) {
	home, paths := managedNamedFixture(t)
	auth := filepath.Join(paths.Root, "codex", "auth.json")
	if err := os.WriteFile(auth, []byte("opaque-fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := cleanupNamed(home, []string{"primary"}, nil, nil); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{auth, paths.Root, paths.Launcher} {
		if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("cleanup retained %s", path)
		}
	}
}

func TestCleanupNamedRemovesReceiptBoundGeneratedCodexState(t *testing.T) {
	home, paths := managedNamedFixture(t)
	auth := filepath.Join(paths.Root, "codex", "auth.json")
	if err := os.WriteFile(auth, []byte("opaque-fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	generatedDir := filepath.Join(paths.Root, "codex", "sessions")
	if err := os.Mkdir(generatedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(generatedDir, "turn.jsonl"), []byte("generated-state"), 0o600); err != nil {
		t.Fatal(err)
	}
	receipt, err := captureCodexCleanupReceipt(home, "primary", testCleanupCandidate, testCleanupRunID)
	if err != nil {
		t.Fatal(err)
	}
	if err = cleanupNamed(home, []string{"primary"}, nil, map[string]codexCleanupReceipt{"primary": receipt}, testCleanupCandidate, testCleanupRunID); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Lstat(paths.Root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("receipt-bound generated state survived reversal: %v", err)
	}
}

func TestCleanupNamedRemovesReceiptBoundPrivatePluginCacheModes(t *testing.T) {
	home, paths := managedNamedFixture(t)
	auth := filepath.Join(paths.Root, "codex", "auth.json")
	generatedDir := filepath.Join(paths.Root, "codex", "plugins", "cache", "fixture")
	if err := os.MkdirAll(generatedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(auth, []byte("opaque-fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	for name, mode := range map[string]os.FileMode{"plugin.json": 0o664, "check.sh": 0o755} {
		path := filepath.Join(generatedDir, name)
		if err := os.WriteFile(path, []byte("generated-state"), mode); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, mode); err != nil {
			t.Fatal(err)
		}
	}
	receipt, err := captureCodexCleanupReceipt(home, "primary", testCleanupCandidate, testCleanupRunID)
	if err != nil {
		t.Fatal(err)
	}
	if err = cleanupNamed(home, []string{"primary"}, nil, map[string]codexCleanupReceipt{"primary": receipt}, testCleanupCandidate, testCleanupRunID); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Lstat(paths.Root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("receipt-bound private plugin cache survived reversal: %v", err)
	}
}

func TestCleanupNamedRemovesReceiptBoundCodexArg0Symlink(t *testing.T) {
	home, paths := managedNamedFixture(t)
	auth := filepath.Join(paths.Root, "codex", "auth.json")
	generatedDir := filepath.Join(paths.Root, "codex", "tmp", "arg0", "codex-fixture")
	if err := os.MkdirAll(generatedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(auth, []byte("opaque-fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest, err := assistantinstance.Verify(home, "primary")
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(generatedDir, "apply_patch")
	if err = os.Symlink(manifest.CodexExecutable, link); err != nil {
		t.Fatal(err)
	}
	receipt, err := captureCodexCleanupReceipt(home, "primary", testCleanupCandidate, testCleanupRunID)
	if err != nil {
		t.Fatal(err)
	}
	if err = cleanupNamed(home, []string{"primary"}, nil, map[string]codexCleanupReceipt{"primary": receipt}, testCleanupCandidate, testCleanupRunID); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Lstat(paths.Root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("receipt-bound arg0 symlink survived reversal: %v", err)
	}
}

func TestGeneratedCodexReceiptPreservesChangedArg0Symlink(t *testing.T) {
	home, paths := managedNamedFixture(t)
	auth := filepath.Join(paths.Root, "codex", "auth.json")
	generatedDir := filepath.Join(paths.Root, "codex", "tmp", "arg0", "codex-fixture")
	if err := os.MkdirAll(generatedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(auth, []byte("opaque-fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest, err := assistantinstance.Verify(home, "primary")
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(generatedDir, "apply_patch")
	if err = os.Symlink(manifest.CodexExecutable, link); err != nil {
		t.Fatal(err)
	}
	receipt, err := captureCodexCleanupReceipt(home, "primary", testCleanupCandidate, testCleanupRunID)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink(filepath.Join(paths.Root, "foreign"), link); err != nil {
		t.Fatal(err)
	}
	if err = cleanupNamed(home, []string{"primary"}, nil, map[string]codexCleanupReceipt{"primary": receipt}, testCleanupCandidate, testCleanupRunID); err == nil {
		t.Fatal("changed generated Codex symlink was accepted")
	}
	if target, readErr := os.Readlink(link); readErr != nil || target != filepath.Join(paths.Root, "foreign") {
		t.Fatalf("changed generated symlink was not preserved: %q %v", target, readErr)
	}
	if body, readErr := os.ReadFile(auth); readErr != nil || string(body) != "opaque-fixture" {
		t.Fatalf("credential changed: %q %v", body, readErr)
	}
}

func TestGeneratedCodexReceiptRefusesManagedCodexSymlinkOutsideArg0Namespace(t *testing.T) {
	home, paths := managedNamedFixture(t)
	generatedDir := filepath.Join(paths.Root, "codex", "sessions")
	if err := os.Mkdir(generatedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	manifest, err := assistantinstance.Verify(home, "primary")
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(generatedDir, "apply_patch")
	if err = os.Symlink(manifest.CodexExecutable, link); err != nil {
		t.Fatal(err)
	}
	if _, err = captureCodexCleanupReceipt(home, "primary", testCleanupCandidate, testCleanupRunID); err == nil {
		t.Fatal("managed Codex symlink outside tmp/arg0 was accepted")
	}
	if target, readErr := os.Readlink(link); readErr != nil || target != manifest.CodexExecutable {
		t.Fatalf("out-of-namespace symlink was not preserved: %q %v", target, readErr)
	}
}

func TestCodexArg0HelperSymlinkNamespaceIsExact(t *testing.T) {
	for _, path := range []string{
		"tmp/arg0/codex-a1/apply_patch",
		"tmp/arg0/codex-Z9/applypatch",
		"tmp/arg0/codex-fixture/codex-execve-wrapper",
	} {
		if !validCodexArg0HelperSymlink(path) {
			t.Fatalf("supported helper path refused: %s", path)
		}
	}
	for _, path := range []string{
		"sessions/codex-a1/apply_patch",
		"tmp/arg0/apply_patch",
		"tmp/arg0/codex-/apply_patch",
		"tmp/arg0/codex-a_b/apply_patch",
		"tmp/arg0/codex-a1/foreign",
		"tmp/arg0/codex-a1/nested/apply_patch",
	} {
		if validCodexArg0HelperSymlink(path) {
			t.Fatalf("unsupported helper path accepted: %s", path)
		}
	}
}

func TestAmbientProtectedMetadataExcludesOnlyObservedLiveCodexState(t *testing.T) {
	for _, path := range []string{
		"codex/sessions",
		"codex/sessions/2026/rollout.jsonl",
		"codex/logs_2.sqlite",
		"codex/logs_2.sqlite-shm",
		"codex/logs_2.sqlite-wal",
		"codex/state_5.sqlite",
		"codex/state_5.sqlite-shm",
		"codex/state_5.sqlite-wal",
	} {
		if !volatileAmbientCodexMetadata(path) {
			t.Fatalf("live Codex metadata path was not excluded: %s", path)
		}
	}
	for _, path := range []string{
		"codex/auth.json",
		"codex/config.toml",
		"codex/skills/example/SKILL.md",
		"codex/logs.sqlite-wal",
		"codex/logs_x.sqlite-wal",
		"codex/logs_2.sqlite-journal",
		"codex/state_5.sqlite.backup",
		"runtime/assistant/profile.json",
	} {
		if volatileAmbientCodexMetadata(path) {
			t.Fatalf("stable protected metadata path was excluded: %s", path)
		}
	}
}

func TestCleanupNamedPreservesChangedGeneratedCodexState(t *testing.T) {
	home, paths := managedNamedFixture(t)
	auth := filepath.Join(paths.Root, "codex", "auth.json")
	generated := filepath.Join(paths.Root, "codex", "state.sqlite")
	if err := os.WriteFile(auth, []byte("opaque-fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(generated, []byte("generated-state"), 0o600); err != nil {
		t.Fatal(err)
	}
	receipt, err := captureCodexCleanupReceipt(home, "primary", testCleanupCandidate, testCleanupRunID)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(generated, generated+".original"); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(generated, []byte("replacement"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = cleanupNamed(home, []string{"primary"}, nil, map[string]codexCleanupReceipt{"primary": receipt}, testCleanupCandidate, testCleanupRunID); err == nil {
		t.Fatal("changed generated Codex state was accepted")
	}
	for path, want := range map[string]string{auth: "opaque-fixture", generated: "replacement", generated + ".original": "generated-state"} {
		if body, readErr := os.ReadFile(path); readErr != nil || string(body) != want {
			t.Fatalf("changed generated state was not preserved at %s: %q %v", path, body, readErr)
		}
	}
}

func TestGeneratedCodexReceiptRejectsMalformedAndDuplicateAuthority(t *testing.T) {
	home, paths := managedNamedFixture(t)
	generated := filepath.Join(paths.Root, "codex", "state.sqlite")
	if err := os.WriteFile(generated, []byte("generated-state"), 0o600); err != nil {
		t.Fatal(err)
	}
	receipt, err := captureCodexCleanupReceipt(home, "primary", testCleanupCandidate, testCleanupRunID)
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]func(*codexCleanupReceipt){
		"candidate": func(r *codexCleanupReceipt) { r.Candidate = strings.Repeat("f", 40) },
		"run":       func(r *codexCleanupReceipt) { r.RunID = "../foreign" },
		"root":      func(r *codexCleanupReceipt) { r.RootInode++ },
		"duplicate": func(r *codexCleanupReceipt) { r.Entries = append(r.Entries, r.Entries[0]) },
		"path":      func(r *codexCleanupReceipt) { r.Entries[0].Path = "../escape" },
		"uid":       func(r *codexCleanupReceipt) { r.Entries[0].UID++ },
		"nlink":     func(r *codexCleanupReceipt) { r.Entries[0].Nlink = 2 },
		"type":      func(r *codexCleanupReceipt) { r.Entries[0].Mode = unix.S_IFLNK | 0o777 },
		"mode":      func(r *codexCleanupReceipt) { r.Entries[0].Mode = unix.S_IFREG | 0o666 },
		"size":      func(r *codexCleanupReceipt) { r.Entries[0].Size++ },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			changed := receipt
			changed.Entries = append([]codexCleanupEntry(nil), receipt.Entries...)
			mutate(&changed)
			if _, err := validateCodexCleanupReceipt(home, "primary", testCleanupCandidate, testCleanupRunID, &changed); err == nil {
				t.Fatal("malformed generated-state receipt was accepted")
			}
		})
	}
}

func TestCleanupNamedRequiresExactNameReceiptAuthoritySet(t *testing.T) {
	home, _ := managedNamedFixture(t)
	receipt, err := captureCodexCleanupReceipt(home, "primary", testCleanupCandidate, testCleanupRunID)
	if err != nil {
		t.Fatal(err)
	}
	valid := map[string]codexCleanupReceipt{"primary": receipt}
	tests := map[string]struct {
		names    []string
		receipts map[string]codexCleanupReceipt
	}{
		"duplicate-name":     {[]string{"primary", "primary"}, valid},
		"missing-receipt":    {[]string{"primary"}, nil},
		"unused-receipt":     {nil, valid},
		"wrong-receipt-name": {[]string{"secondary"}, valid},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := cleanupNamed(home, test.names, nil, test.receipts, testCleanupCandidate, testCleanupRunID); err == nil {
				t.Fatal("non-exact name/receipt authority set was accepted")
			}
		})
	}
	body, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = decodeCodexCleanupReceipts([]string{string(body), string(body)}, testCleanupCandidate, testCleanupRunID); err == nil {
		t.Fatal("duplicate receipt identity was accepted")
	}
}

func TestGeneratedCodexReceiptRejectsFilesystemDrift(t *testing.T) {
	tests := map[string]func(*testing.T, string){
		"missing": func(t *testing.T, path string) { t.Helper(); mustRemove(t, path) },
		"extra": func(t *testing.T, path string) {
			t.Helper()
			mustWrite(t, filepath.Join(filepath.Dir(path), "extra"), 0o600)
		},
		"inode": func(t *testing.T, path string) { t.Helper(); mustRemove(t, path); mustWrite(t, path, 0o600) },
		"size": func(t *testing.T, path string) {
			t.Helper()
			if err := os.WriteFile(path, []byte("different-size"), 0o600); err != nil {
				t.Fatal(err)
			}
		},
		"mtime": func(t *testing.T, path string) {
			t.Helper()
			if err := os.Chtimes(path, time.Unix(10, 0), time.Unix(10, 0)); err != nil {
				t.Fatal(err)
			}
		},
		"mode": func(t *testing.T, path string) {
			t.Helper()
			if err := os.Chmod(path, 0o640); err != nil {
				t.Fatal(err)
			}
		},
		"type": func(t *testing.T, path string) {
			t.Helper()
			mustRemove(t, path)
			if err := os.Mkdir(path, 0o700); err != nil {
				t.Fatal(err)
			}
		},
		"symlink": func(t *testing.T, path string) {
			t.Helper()
			mustRemove(t, path)
			if err := os.Symlink("target", path); err != nil {
				t.Fatal(err)
			}
		},
		"hardlink": func(t *testing.T, path string) {
			t.Helper()
			if err := os.Link(path, path+".link"); err != nil {
				t.Fatal(err)
			}
		},
		"nested-extra": func(t *testing.T, path string) {
			t.Helper()
			mustWrite(t, filepath.Join(filepath.Dir(path), "nested-extra"), 0o600)
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			home, paths := managedNamedFixture(t)
			auth := filepath.Join(paths.Root, "codex", "auth.json")
			nested := filepath.Join(paths.Root, "codex", "sessions", "2026")
			if err := os.MkdirAll(nested, 0o700); err != nil {
				t.Fatal(err)
			}
			generated := filepath.Join(nested, "turn.jsonl")
			mustWrite(t, auth, 0o600)
			mustWrite(t, generated, 0o600)
			receipt, err := captureCodexCleanupReceipt(home, "primary", testCleanupCandidate, testCleanupRunID)
			if err != nil {
				t.Fatal(err)
			}
			mutate(t, generated)
			if err = cleanupNamed(home, []string{"primary"}, nil, map[string]codexCleanupReceipt{"primary": receipt}, testCleanupCandidate, testCleanupRunID); err == nil {
				t.Fatal("generated Codex filesystem drift was accepted")
			}
			if _, err = os.Lstat(paths.Root); err != nil {
				t.Fatalf("instance was not preserved: %v", err)
			}
			if body, readErr := os.ReadFile(auth); readErr != nil || string(body) != "fixture" {
				t.Fatalf("credential changed: %q %v", body, readErr)
			}
		})
	}
}

func TestGeneratedCodexReceiptRefusesDriftImmediatelyBeforeRemoval(t *testing.T) {
	home, paths := managedNamedFixture(t)
	auth := filepath.Join(paths.Root, "codex", "auth.json")
	generated := filepath.Join(paths.Root, "codex", "state.sqlite")
	mustWrite(t, auth, 0o600)
	mustWrite(t, generated, 0o600)
	receipt, err := captureCodexCleanupReceipt(home, "primary", testCleanupCandidate, testCleanupRunID)
	if err != nil {
		t.Fatal(err)
	}
	late := filepath.Join(paths.Root, "codex", "late-state")
	previousHook := cleanupMutationHook
	cleanupMutationHook = func(phase string) {
		if phase != "auth-before-final-removal-validation" {
			return
		}
		cleanupMutationHook = nil
		mustWrite(t, late, 0o600)
	}
	defer func() { cleanupMutationHook = previousHook }()
	if err = cleanupNamed(home, []string{"primary"}, nil, map[string]codexCleanupReceipt{"primary": receipt}, testCleanupCandidate, testCleanupRunID); err == nil {
		t.Fatal("late generated-state drift was accepted")
	}
	if body, readErr := os.ReadFile(late); readErr != nil || string(body) != "fixture" {
		t.Fatalf("late state was not preserved: %q %v", body, readErr)
	}
	if _, err = os.Lstat(paths.Root); err != nil {
		t.Fatalf("instance was not preserved: %v", err)
	}
}

func mustWrite(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte("fixture"), mode); err != nil {
		t.Fatal(err)
	}
}

func mustRemove(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}

func TestCleanupNamedRefusesUnsafeDisposableAuthAndUnexpectedEntries(t *testing.T) {
	tests := map[string]func(*testing.T, assistantinstance.Paths) string{
		"symlink": func(t *testing.T, paths assistantinstance.Paths) string {
			target := filepath.Join(paths.Root, "auth-target")
			if err := os.WriteFile(target, []byte("preserve"), 0o600); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(paths.Root, "codex", "auth.json")
			if err := os.Symlink(target, path); err != nil {
				t.Fatal(err)
			}
			return path
		},
		"hardlink": func(t *testing.T, paths assistantinstance.Paths) string {
			target := filepath.Join(paths.Root, "auth-target")
			if err := os.WriteFile(target, []byte("preserve"), 0o600); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(paths.Root, "codex", "auth.json")
			if err := os.Link(target, path); err != nil {
				t.Fatal(err)
			}
			return path
		},
		"wrong-mode": func(t *testing.T, paths assistantinstance.Paths) string {
			path := filepath.Join(paths.Root, "codex", "auth.json")
			if err := os.WriteFile(path, []byte("preserve"), 0o640); err != nil {
				t.Fatal(err)
			}
			return path
		},
		"alternate-path": func(t *testing.T, paths assistantinstance.Paths) string {
			path := filepath.Join(paths.Root, "codex", "auth-copy.json")
			if err := os.WriteFile(path, []byte("preserve"), 0o600); err != nil {
				t.Fatal(err)
			}
			return path
		},
		"unrelated-extra": func(t *testing.T, paths assistantinstance.Paths) string {
			if err := os.WriteFile(filepath.Join(paths.Root, "codex", "auth.json"), []byte("opaque-fixture"), 0o600); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(paths.Root, "codex", "unrelated")
			if err := os.WriteFile(path, []byte("preserve"), 0o600); err != nil {
				t.Fatal(err)
			}
			return path
		},
		"unrelated-root-extra": func(t *testing.T, paths assistantinstance.Paths) string {
			if err := os.WriteFile(filepath.Join(paths.Root, "codex", "auth.json"), []byte("opaque-fixture"), 0o600); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(paths.Root, "unrelated")
			if err := os.WriteFile(path, []byte("preserve"), 0o600); err != nil {
				t.Fatal(err)
			}
			return path
		},
	}
	for name, arrange := range tests {
		t.Run(name, func(t *testing.T) {
			home, paths := managedNamedFixture(t)
			preserved := arrange(t, paths)
			if err := cleanupNamed(home, []string{"primary"}, nil, nil); err == nil {
				t.Fatal("unsafe disposable credential state was removed")
			}
			if _, err := os.Lstat(paths.Root); err != nil {
				t.Fatalf("instance root changed: %v", err)
			}
			if _, err := os.Lstat(preserved); err != nil {
				t.Fatalf("unsafe entry was not preserved: %v", err)
			}
		})
	}
}

func TestDisposableAuthOwnershipPredicateRejectsWrongOwner(t *testing.T) {
	if disposableAuthStatSafe(0o100600, uint32(os.Getuid()+1), 1) {
		t.Fatal("foreign owner accepted")
	}
}

func TestCleanupNamedPreservesAuthReplacementRace(t *testing.T) {
	home, paths := managedNamedFixture(t)
	auth := filepath.Join(paths.Root, "codex", "auth.json")
	original := filepath.Join(paths.Root, "original-auth")
	if err := os.WriteFile(auth, []byte("verified-original"), 0o600); err != nil {
		t.Fatal(err)
	}
	previousHook := cleanupMutationHook
	cleanupMutationHook = func(phase string) {
		if phase != "auth-before-quarantine" {
			return
		}
		cleanupMutationHook = nil
		if err := os.Rename(auth, original); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(auth, []byte("foreign-replacement"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	defer func() { cleanupMutationHook = previousHook }()
	if err := cleanupNamed(home, []string{"primary"}, nil, nil); err == nil {
		t.Fatal("replacement race was accepted")
	}
	for path, want := range map[string]string{auth: "foreign-replacement", original: "verified-original"} {
		if body, err := os.ReadFile(path); err != nil || string(body) != want {
			t.Fatalf("race entry changed at %s: %q %v", path, body, err)
		}
	}
	if _, err := os.Lstat(paths.Root); err != nil {
		t.Fatalf("instance root changed: %v", err)
	}
}

func TestCleanupNamedPreservesCodexDirectoryReplacementRace(t *testing.T) {
	home, paths := managedNamedFixture(t)
	codexRoot := filepath.Join(paths.Root, "codex")
	auth := filepath.Join(codexRoot, "auth.json")
	originalCodex := filepath.Join(paths.Root, "codex-original")
	if err := os.WriteFile(auth, []byte("verified-original"), 0o600); err != nil {
		t.Fatal(err)
	}
	previousHook := cleanupMutationHook
	cleanupMutationHook = func(phase string) {
		if phase != "auth-before-quarantine" {
			return
		}
		cleanupMutationHook = nil
		if err := os.Rename(codexRoot, originalCodex); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(codexRoot, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(codexRoot, "auth.json"), []byte("foreign-directory-auth"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	defer func() { cleanupMutationHook = previousHook }()
	if err := cleanupNamed(home, []string{"primary"}, nil, nil); err == nil {
		t.Fatal("directory replacement race was accepted")
	}
	for path, want := range map[string]string{
		filepath.Join(codexRoot, "auth.json"):     "foreign-directory-auth",
		filepath.Join(originalCodex, "auth.json"): "verified-original",
	} {
		if body, err := os.ReadFile(path); err != nil || string(body) != want {
			t.Fatalf("directory race entry changed at %s: %q %v", path, body, err)
		}
	}
}

func TestCleanupNamedPreservesCodexDirectoryReplacementAfterQuarantine(t *testing.T) {
	home, paths := managedNamedFixture(t)
	codexRoot := filepath.Join(paths.Root, "codex")
	auth := filepath.Join(codexRoot, "auth.json")
	originalCodex := filepath.Join(paths.Root, "codex-original")
	if err := os.WriteFile(auth, []byte("verified-original"), 0o600); err != nil {
		t.Fatal(err)
	}
	previousHook := cleanupMutationHook
	cleanupMutationHook = func(phase string) {
		if phase != "auth-quarantine-verified" {
			return
		}
		cleanupMutationHook = nil
		if err := os.Rename(codexRoot, originalCodex); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(codexRoot, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(codexRoot, "auth.json"), []byte("foreign-directory-auth"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	defer func() { cleanupMutationHook = previousHook }()
	if err := cleanupNamed(home, []string{"primary"}, nil, nil); err == nil {
		t.Fatal("post-quarantine directory replacement race was accepted")
	}
	for path, want := range map[string]string{
		filepath.Join(codexRoot, "auth.json"):     "foreign-directory-auth",
		filepath.Join(originalCodex, "auth.json"): "verified-original",
	} {
		if body, err := os.ReadFile(path); err != nil || string(body) != want {
			t.Fatalf("post-quarantine directory race changed %s: %q %v", path, body, err)
		}
	}
	entries, err := os.ReadDir(paths.Root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".auth.json.my-friday-cleanup-") {
			t.Fatalf("credential stranded in root quarantine: %s", entry.Name())
		}
	}
}

func authQuarantineNameForTest(t *testing.T, path string) string {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	st := info.Sys().(*syscall.Stat_t)
	return fmt.Sprintf("%s%x-%x", authQuarantinePrefix, uint64(st.Dev), st.Ino)
}

func TestCleanupNamedPreservesFinalPathReplacementAfterNeutralization(t *testing.T) {
	home, paths := managedNamedFixture(t)
	auth := filepath.Join(paths.Root, "codex", "auth.json")
	if err := os.WriteFile(auth, []byte("verified-original"), 0o600); err != nil {
		t.Fatal(err)
	}
	quarantine := filepath.Join(paths.Root, authQuarantineNameForTest(t, auth))
	preservedOriginal := filepath.Join(paths.Root, "neutralized-original")
	previousHook := cleanupMutationHook
	cleanupMutationHook = func(phase string) {
		if phase != "auth-before-neutralize" {
			return
		}
		cleanupMutationHook = nil
		if err := os.Rename(quarantine, preservedOriginal); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(quarantine, []byte("foreign-replacement"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	defer func() { cleanupMutationHook = previousHook }()
	if err := cleanupNamed(home, []string{"primary"}, nil, nil); err == nil {
		t.Fatal("final path replacement was accepted")
	}
	if body, err := os.ReadFile(quarantine); err != nil || string(body) != "foreign-replacement" {
		t.Fatalf("replacement was changed: %q %v", body, err)
	}
	if body, err := os.ReadFile(preservedOriginal); err != nil || string(body) != "verified-original" {
		t.Fatalf("verified credential changed before refused disposal: %q %v", body, err)
	}
	if _, err := os.Lstat(paths.Root); err != nil {
		t.Fatalf("instance changed after refusal: %v", err)
	}
}

func TestCleanupNamedRecoversInterruptedAuthQuarantines(t *testing.T) {
	for _, phase := range []string{"auth-after-codex-quarantine", "auth-after-root-quarantine", "auth-after-neutralize"} {
		t.Run(phase, func(t *testing.T) {
			home, paths := managedNamedFixture(t)
			auth := filepath.Join(paths.Root, "codex", "auth.json")
			if err := os.WriteFile(auth, []byte("verified-original"), 0o600); err != nil {
				t.Fatal(err)
			}
			previousHook := cleanupMutationHook
			cleanupMutationHook = func(observed string) {
				if observed == phase {
					panic("simulated interruption")
				}
			}
			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("cleanup did not reach interruption hook")
					}
				}()
				_ = cleanupNamed(home, []string{"primary"}, nil, nil)
			}()
			cleanupMutationHook = nil
			defer func() { cleanupMutationHook = previousHook }()
			if err := cleanupNamed(home, []string{"primary"}, nil, nil); err != nil {
				t.Fatalf("retry did not recover %s: %v", phase, err)
			}
			if _, err := os.Lstat(paths.Root); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("retry did not fully reverse instance: %v", err)
			}
		})
	}
}

func TestCleanupNamedPreservesRootDestinationCollisionAfterCodexQuarantine(t *testing.T) {
	home, paths := managedNamedFixture(t)
	auth := filepath.Join(paths.Root, "codex", "auth.json")
	if err := os.WriteFile(auth, []byte("verified-original"), 0o600); err != nil {
		t.Fatal(err)
	}
	name := authQuarantineNameForTest(t, auth)
	previousHook := cleanupMutationHook
	cleanupMutationHook = func(phase string) {
		if phase != "auth-after-codex-quarantine" {
			return
		}
		cleanupMutationHook = nil
		if err := os.WriteFile(filepath.Join(paths.Root, name), []byte("foreign-root-collision"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	defer func() { cleanupMutationHook = previousHook }()
	if err := cleanupNamed(home, []string{"primary"}, nil, nil); err == nil {
		t.Fatal("root destination collision was accepted")
	}
	for path, want := range map[string]string{
		filepath.Join(paths.Root, "codex", name): "verified-original",
		filepath.Join(paths.Root, name):          "foreign-root-collision",
	} {
		if body, err := os.ReadFile(path); err != nil || string(body) != want {
			t.Fatalf("destination collision changed at %s: %q %v", path, body, err)
		}
	}
}

func TestCleanupNamedRefusesRequiredProjectionDisappearanceBeforeMutation(t *testing.T) {
	tests := []struct {
		phase            string
		requiredRelative string
	}{
		{"auth-before-quarantine", "manifest.json"},
		{"auth-after-codex-quarantine", "codex/config.toml"},
		{"auth-quarantine-verified", "codex/AGENTS.md"},
		{"auth-after-root-quarantine", "dependencies"},
		{"auth-before-neutralize", "workspace"},
	}
	for _, test := range tests {
		t.Run(test.phase, func(t *testing.T) {
			home, paths := managedNamedFixture(t)
			auth := filepath.Join(paths.Root, "codex", "auth.json")
			if err := os.WriteFile(auth, []byte("verified-original"), 0o600); err != nil {
				t.Fatal(err)
			}
			quarantine := authQuarantineNameForTest(t, auth)
			required := filepath.Join(paths.Root, filepath.FromSlash(test.requiredRelative))
			preservedRequired := required + ".removed-by-test"
			previousHook := cleanupMutationHook
			cleanupMutationHook = func(phase string) {
				if phase != test.phase {
					return
				}
				cleanupMutationHook = nil
				if err := os.Rename(required, preservedRequired); err != nil {
					t.Fatal(err)
				}
			}
			defer func() { cleanupMutationHook = previousHook }()
			if err := cleanupNamed(home, []string{"primary"}, nil, nil); err == nil {
				t.Fatal("required projection disappearance was accepted")
			}
			preserved := false
			for _, credential := range []string{
				filepath.Join(paths.Root, "codex", "auth.json"),
				filepath.Join(paths.Root, "codex", quarantine),
				filepath.Join(paths.Root, quarantine),
			} {
				if body, err := os.ReadFile(credential); err == nil && string(body) == "verified-original" {
					preserved = true
				}
			}
			if !preserved {
				t.Fatal("credential was not preserved after required projection disappeared")
			}
		})
	}
}

func TestCleanupNamedRequiresExactEntrySetsAfterManifestVerification(t *testing.T) {
	for _, requiredRelative := range []string{
		"manifest.json",
		"dependencies",
		"codex/config.toml",
		"codex/AGENTS.md",
	} {
		t.Run(strings.ReplaceAll(requiredRelative, "/", "-"), func(t *testing.T) {
			home, paths := managedNamedFixture(t)
			auth := filepath.Join(paths.Root, "codex", "auth.json")
			if err := os.WriteFile(auth, []byte("verified-original"), 0o600); err != nil {
				t.Fatal(err)
			}
			required := filepath.Join(paths.Root, filepath.FromSlash(requiredRelative))
			preservedRequired := required + ".removed-by-test"
			previousHook := cleanupMutationHook
			cleanupMutationHook = func(phase string) {
				if phase != "authority-after-manifest-verify" {
					return
				}
				cleanupMutationHook = nil
				if err := os.Rename(required, preservedRequired); err != nil {
					t.Fatal(err)
				}
			}
			defer func() { cleanupMutationHook = previousHook }()
			if err := cleanupNamed(home, []string{"primary"}, nil, nil); err == nil {
				t.Fatal("incomplete descriptor entry set was accepted")
			}
			if body, err := os.ReadFile(auth); err != nil || string(body) != "verified-original" {
				t.Fatalf("credential mutated after required entry disappeared: %q %v", body, err)
			}
			if _, err := os.Lstat(preservedRequired); err != nil {
				t.Fatalf("removed required entry was further changed: %v", err)
			}
		})
	}
}

func TestCleanupNamedPreservesAmbiguousQuarantineCollisions(t *testing.T) {
	for _, location := range []string{"codex", "root"} {
		t.Run(location, func(t *testing.T) {
			home, paths := managedNamedFixture(t)
			auth := filepath.Join(paths.Root, "codex", "auth.json")
			if err := os.WriteFile(auth, []byte("verified-original"), 0o600); err != nil {
				t.Fatal(err)
			}
			name := authQuarantineNameForTest(t, auth)
			dir := paths.Root
			if location == "codex" {
				dir = filepath.Join(paths.Root, "codex")
			}
			collision := filepath.Join(dir, name)
			if err := os.WriteFile(collision, []byte("foreign-collision"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := cleanupNamed(home, []string{"primary"}, nil, nil); err == nil {
				t.Fatal("ambiguous quarantine collision was accepted")
			}
			for path, want := range map[string]string{auth: "verified-original", collision: "foreign-collision"} {
				if body, err := os.ReadFile(path); err != nil || string(body) != want {
					t.Fatalf("collision state changed at %s: %q %v", path, body, err)
				}
			}
		})
	}
}

func exactLeafForTest(t *testing.T, path string) exactLeaf {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	st := info.Sys().(*syscall.Stat_t)
	return exactLeaf{Path: path, Device: uint64(st.Dev), Inode: st.Ino, SHA256: fmt.Sprintf("%x", sha256.Sum256(body))}
}

func TestCleanupNamedCoversEveryPostCreatePhase(t *testing.T) {
	for _, phase := range []string{"primary-created", "secondary-created", "collision-created", "recovery-complete", "auth-copied", "smoke-complete"} {
		t.Run(phase, func(t *testing.T) {
			home := t.TempDir()
			if err := os.MkdirAll(filepath.Join(home, ".local", "bin"), 0o755); err != nil {
				t.Fatal(err)
			}
			exe, codex := filepath.Join(home, "candidate"), filepath.Join(home, "codex")
			for _, path := range []string{exe, codex, filepath.Join(home, "codex-code-mode-host")} {
				if err := os.WriteFile(path, []byte(path), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			create := func(name string) {
				p, err := assistantinstance.PlanCreate(home, name, exe, codex)
				if err != nil {
					t.Fatal(err)
				}
				if err = assistantinstance.Create(p, exe, codex); err != nil {
					t.Fatal(err)
				}
			}
			create("primary")
			if phase != "primary-created" {
				create("secondary")
			}
			preReceipts := make(map[string]codexCleanupReceipt)
			for _, receiptName := range []string{"primary", "secondary"} {
				if _, statErr := os.Lstat(filepath.Join(home, ".my-friday", "assistants", receiptName)); statErr != nil {
					continue
				}
				receipt, receiptErr := captureCodexCleanupReceipt(home, receiptName, testCleanupCandidate, testCleanupRunID)
				if receiptErr != nil {
					t.Fatal(receiptErr)
				}
				preReceipts[receiptName] = receipt
			}
			var leaves []exactLeaf
			if phase == "collision-created" || phase == "recovery-complete" || phase == "auth-copied" || phase == "smoke-complete" {
				for _, name := range []string{"mfac-collision", "mfac-sibling"} {
					path := filepath.Join(home, ".local", "bin", name)
					if err := os.WriteFile(path, []byte(name), 0o700); err != nil {
						t.Fatal(err)
					}
					leaves = append(leaves, exactLeafForTest(t, path))
				}
			}
			if phase == "recovery-complete" || phase == "auth-copied" || phase == "smoke-complete" {
				if err := os.Remove(filepath.Join(home, ".local", "bin", "secondary")); err != nil {
					t.Fatal(err)
				}
			}
			source := filepath.Join(home, "source-auth.json")
			if err := os.WriteFile(source, []byte("ambient-secret"), 0o600); err != nil {
				t.Fatal(err)
			}
			if phase == "auth-copied" || phase == "smoke-complete" {
				body, _ := os.ReadFile(source)
				if err := os.WriteFile(filepath.Join(home, ".my-friday", "assistants", "primary", "codex", "auth.json"), body, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if phase == "smoke-complete" {
				mustWrite(t, filepath.Join(home, ".my-friday", "assistants", "primary", "codex", "state.sqlite"), 0o600)
			}
			var cleanupNames []string
			receipts := make(map[string]codexCleanupReceipt)
			for _, cleanupName := range []string{"primary", "secondary"} {
				if _, statErr := os.Lstat(filepath.Join(home, ".my-friday", "assistants", cleanupName)); statErr != nil {
					continue
				}
				receipt, err := captureCodexCleanupReceipt(home, cleanupName, testCleanupCandidate, testCleanupRunID)
				if err != nil {
					receipt = preReceipts[cleanupName]
				}
				if receipt.Name == "" {
					t.Fatal(err)
				}
				cleanupNames = append(cleanupNames, cleanupName)
				receipts[cleanupName] = receipt
			}
			if err := cleanupNamed(home, cleanupNames, leaves, receipts, testCleanupCandidate, testCleanupRunID); err != nil {
				t.Fatal(err)
			}
			for _, path := range []string{filepath.Join(home, ".my-friday", "assistants", "primary"), filepath.Join(home, ".my-friday", "assistants", "secondary"), filepath.Join(home, ".local", "bin", "primary"), filepath.Join(home, ".local", "bin", "secondary"), filepath.Join(home, ".local", "bin", "mfac-collision"), filepath.Join(home, ".local", "bin", "mfac-sibling")} {
				if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("survived: %s", path)
				}
			}
			if body, err := os.ReadFile(source); err != nil || string(body) != "ambient-secret" {
				t.Fatal("ambient auth changed")
			}
		})
	}
}

func TestCleanupNamedRefusesReceiptOutsideAcceptanceLauncherScope(t *testing.T) {
	home := t.TempDir()
	foreign := filepath.Join(home, "foreign")
	if err := os.WriteFile(foreign, []byte("preserve"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := cleanupNamed(home, nil, []exactLeaf{exactLeafForTest(t, foreign)}, nil); err == nil {
		t.Fatal("out-of-scope receipt accepted")
	}
	if body, err := os.ReadFile(foreign); err != nil || string(body) != "preserve" {
		t.Fatal("foreign leaf changed")
	}
}

func TestCleanupNamedPreservesDriftedLeafButRemovesInstanceAndCredential(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".local", "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	exe, codex := filepath.Join(home, "candidate"), filepath.Join(home, "codex")
	for _, path := range []string{exe, codex, filepath.Join(home, "codex-code-mode-host")} {
		if err := os.WriteFile(path, []byte(path), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	p, err := assistantinstance.PlanCreate(home, "primary", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = assistantinstance.Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	auth := filepath.Join(home, ".my-friday", "assistants", "primary", "codex", "auth.json")
	if err = os.WriteFile(auth, []byte("copied-credential"), 0o600); err != nil {
		t.Fatal(err)
	}
	leaf := filepath.Join(home, ".local", "bin", "mfac-sibling")
	if err = os.WriteFile(leaf, []byte("original"), 0o700); err != nil {
		t.Fatal(err)
	}
	receipt := exactLeafForTest(t, leaf)
	exactLeafPath := filepath.Join(home, ".local", "bin", "mfac-collision")
	if err = os.WriteFile(exactLeafPath, []byte("exact-delete"), 0o700); err != nil {
		t.Fatal(err)
	}
	exactReceipt := exactLeafForTest(t, exactLeafPath)
	if err = os.WriteFile(leaf, []byte("drifted-preserve"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err = cleanupNamed(home, []string{"primary"}, []exactLeaf{receipt, exactReceipt}, nil); err == nil {
		t.Fatal("drifted leaf was not reported")
	}
	for _, path := range []string{filepath.Join(home, ".my-friday", "assistants", "primary"), filepath.Join(home, ".local", "bin", "primary"), auth} {
		if _, statErr := os.Lstat(path); !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("manifest-owned state survived: %s", path)
		}
	}
	if body, readErr := os.ReadFile(leaf); readErr != nil || string(body) != "drifted-preserve" {
		t.Fatal("drifted foreign leaf changed")
	}
	if _, statErr := os.Lstat(exactLeafPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatal("exact foreign leaf was not independently cleaned")
	}
}

func TestFixtureIsValidAndRendersToken(t *testing.T) {
	root := t.TempDir()
	runtime := filepath.Join(root, "runtime")
	memory := filepath.Join(root, "memory")
	cmd := exec.Command("go", "run", ".", "fixture", "--runtime", runtime, "--memory", memory, "--token", "TOKEN_ONE")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture: %v: %s", err, out)
	}
	codex := filepath.Join(root, "home", ".codex")
	if err := os.MkdirAll(codex, 0o700); err != nil {
		t.Fatal(err)
	}
	p, err := codexhome.Plan(codexhome.ActionInstall, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = codexhome.Execute(p); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(codex, "AGENTS.md"))
	if err != nil || !strings.Contains(string(body), "TOKEN_ONE") {
		t.Fatalf("rendered token missing: %v %q", err, body)
	}
	cmd = exec.Command("go", "run", ".", "update", "--runtime", runtime, "--token", "TOKEN_TWO")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("update: %v: %s", err, out)
	}
	p, err = codexhome.Plan(codexhome.ActionUpgrade, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(p.String(), "replace") {
		t.Fatal("valid profile update did not plan an upgrade")
	}
}

func TestSchemeStringEscapesQuotesBackslashesAndNewlines(t *testing.T) {
	got := strconv.Quote("/tmp/a\"b\\c\n")
	if got != `"/tmp/a\"b\\c\n"` {
		t.Fatalf("unsafe Scheme string: %s", got)
	}
}

func TestSecureRootsRefusesCollision(t *testing.T) {
	home := t.TempDir()
	cmd := exec.Command("go", "run", ".", "secure-roots", "--home", home, "--run-id", "run")
	if out, err := cmd.CombinedOutput(); err != nil || !strings.Contains(string(out), "device") {
		t.Fatalf("secure roots: %v %s", err, out)
	}
	cmd = exec.Command("go", "run", ".", "secure-roots", "--home", home, "--run-id", "run")
	if err := cmd.Run(); err == nil {
		t.Fatal("secure root collision was accepted")
	}
}

func TestNoFollowReadRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	secret := filepath.Join(root, "secret")
	link := filepath.Join(root, "link")
	if err := os.WriteFile(secret, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(secret, link); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := noFollowRead(link); err == nil || ok {
		t.Fatal("symlinked protected content was accepted")
	}
}

func TestNoFollowReadRejectsIntermediateSymlink(t *testing.T) {
	root := t.TempDir()
	realDir := filepath.Join(root, "real")
	if err := os.Mkdir(realDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(realDir, "protected"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realDir, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := noFollowRead(filepath.Join(root, "link", "protected")); err == nil || ok {
		t.Fatal("intermediate symlink in protected path was accepted")
	}
}

func TestResolveExecutableAcceptsOwnedSymlinkChain(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "codex-real")
	current := filepath.Join(root, "current")
	link := filepath.Join(root, "codex")
	if err := os.WriteFile(target, []byte("binary"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, current); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(current, link); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "run", ".", "resolve-executable", link)
	out, err := cmd.CombinedOutput()
	expected, _ := filepath.EvalSymlinks(target)
	if err != nil || strings.TrimSpace(string(out)) != expected {
		t.Fatalf("safe installed symlink was not resolved: %v %s", err, out)
	}
}

func TestRenderProfileDoesNotInterpretReplacementMetacharacters(t *testing.T) {
	root := t.TempDir()
	template := filepath.Join(root, "profile.in")
	if err := os.WriteFile(template, []byte("(subpath @@VOLUME@@)\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	value := "/tmp/a&b|c\\d\nq\"r"
	cmd := exec.Command("go", "run", ".", "render-profile", "--template", template, "--value", value)
	out, err := cmd.CombinedOutput()
	if err != nil || string(out) != "(subpath "+strconv.Quote(value)+")\n" || strings.Contains(string(out), "@@VOLUME@@") {
		t.Fatalf("profile was not rendered literally: %v %q", err, out)
	}
}

func TestSecureRootsRejectsSymlinkedAncestor(t *testing.T) {
	root := t.TempDir()
	realParent := filepath.Join(root, "real")
	if err := os.Mkdir(realParent, 0o700); err != nil {
		t.Fatal(err)
	}
	linkParent := filepath.Join(root, "link")
	if err := os.Symlink(realParent, linkParent); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(linkParent, "home")
	if err := os.Mkdir(filepath.Join(realParent, "home"), 0o700); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "run", ".", "secure-roots", "--home", home, "--run-id", "run")
	if err := cmd.Run(); err == nil {
		t.Fatal("secure roots accepted symlinked ancestry")
	}
}

func TestCleanupRootsIsReceiptAndMarkerBound(t *testing.T) {
	home := t.TempDir()
	secure := exec.Command("go", "run", ".", "secure-roots", "--home", home, "--run-id", "run")
	receipt, err := secure.CombinedOutput()
	if err != nil {
		t.Fatalf("secure roots: %v %s", err, receipt)
	}
	marker := []byte("marker\n")
	markerSHA := fmt.Sprintf("%x", sha256.Sum256(marker))
	for _, parent := range []string{".my-friday-acceptance", ".my-friday-acceptance-evidence"} {
		child := filepath.Join(home, parent, "run")
		if err = os.WriteFile(filepath.Join(child, "marker.json"), marker, 0o600); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(child, "owned"), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cleanup := exec.Command("go", "run", ".", "cleanup-roots", "--home", home, "--run-id", "run", "--receipt", strings.TrimSpace(string(receipt)), "--marker-sha256", markerSHA,
		"--expected-entry", ".my-friday-acceptance:marker.json", "--expected-entry", ".my-friday-acceptance:owned",
		"--expected-entry", ".my-friday-acceptance-evidence:marker.json", "--expected-entry", ".my-friday-acceptance-evidence:owned")
	if out, runErr := cleanup.CombinedOutput(); runErr != nil {
		t.Fatalf("cleanup: %v %s", runErr, out)
	}
	for _, parent := range []string{".my-friday-acceptance", ".my-friday-acceptance-evidence"} {
		if _, err = os.Stat(filepath.Join(home, parent, "run")); !os.IsNotExist(err) {
			t.Fatal("run child survived cleanup")
		}
		if _, err = os.Stat(filepath.Join(home, parent)); err != nil {
			t.Fatal("fixed parent was removed")
		}
	}
}

func TestCleanupRootsPreservesUnexpectedEntry(t *testing.T) {
	home := t.TempDir()
	receipt, err := exec.Command("go", "run", ".", "secure-roots", "--home", home, "--run-id", "run").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("marker\n")
	markerSHA := fmt.Sprintf("%x", sha256.Sum256(marker))
	for _, parent := range []string{".my-friday-acceptance", ".my-friday-acceptance-evidence"} {
		child := filepath.Join(home, parent, "run")
		if err = os.WriteFile(filepath.Join(child, "marker.json"), marker, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	unexpected := filepath.Join(home, ".my-friday-acceptance", "run", "unexpected")
	if err = os.WriteFile(unexpected, []byte("preserve"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "run", ".", "cleanup-roots", "--home", home, "--run-id", "run", "--receipt", strings.TrimSpace(string(receipt)), "--marker-sha256", markerSHA,
		"--expected-entry", ".my-friday-acceptance:marker.json", "--expected-entry", ".my-friday-acceptance-evidence:marker.json")
	if err = cmd.Run(); err == nil {
		t.Fatal("cleanup accepted an unexpected entry")
	}
	if body, readErr := os.ReadFile(unexpected); readErr != nil || string(body) != "preserve" {
		t.Fatal("unexpected entry was not preserved")
	}
	if body, readErr := os.ReadFile(filepath.Join(home, ".my-friday-acceptance", "run", "marker.json")); readErr != nil || string(body) != "marker\n" {
		t.Fatal("cleanup mutated expected state before refusing the unexpected entry")
	}
}

func TestCleanupRootsPrevalidatesBothRootsBeforeMutation(t *testing.T) {
	home := t.TempDir()
	receipt, err := exec.Command("go", "run", ".", "secure-roots", "--home", home, "--run-id", "run").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("marker\n")
	markerSHA := fmt.Sprintf("%x", sha256.Sum256(marker))
	for _, parent := range []string{".my-friday-acceptance", ".my-friday-acceptance-evidence"} {
		if err = os.WriteFile(filepath.Join(home, parent, "run", "marker.json"), marker, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	unexpected := filepath.Join(home, ".my-friday-acceptance-evidence", "run", "unexpected")
	if err = os.WriteFile(unexpected, []byte("preserve"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "run", ".", "cleanup-roots", "--home", home, "--run-id", "run", "--receipt", strings.TrimSpace(string(receipt)), "--marker-sha256", markerSHA,
		"--expected-entry", ".my-friday-acceptance:marker.json", "--expected-entry", ".my-friday-acceptance-evidence:marker.json")
	if err = cmd.Run(); err == nil {
		t.Fatal("cleanup accepted an evidence-root surprise")
	}
	for _, parent := range []string{".my-friday-acceptance", ".my-friday-acceptance-evidence"} {
		if body, readErr := os.ReadFile(filepath.Join(home, parent, "run", "marker.json")); readErr != nil || string(body) != "marker\n" {
			t.Fatalf("cleanup mutated %s before both roots validated", parent)
		}
	}
}

func TestSandboxDiagnosticAllowlistV1(t *testing.T) {
	exact := "sandbox-exec: warning: sandbox-exec is deprecated and will be removed in a future release."
	for name, test := range map[string]struct {
		input string
		want  bool
	}{
		"empty": {"", true}, "exact": {exact, true}, "suffix": {exact + " attacker", false},
		"multiline": {exact + "\nunexpected", false}, "duplicate": {exact + "\n" + exact, false},
	} {
		t.Run(name, func(t *testing.T) {
			if got := validSandboxDiagnostic("v1", test.input); got != test.want {
				t.Fatalf("got %v want %v", got, test.want)
			}
		})
	}
}

func TestSandboxDiagnosticRejectsUnknownVersionBeforeContent(t *testing.T) {
	for name, input := range map[string]string{"empty": "", "nonempty": "sandbox diagnostic"} {
		t.Run(name, func(t *testing.T) {
			if validSandboxDiagnostic("v2", input) {
				t.Fatal("unknown allowlist version was accepted")
			}
		})
	}
}

func TestCopyAuthNoFollowAndSourceSwapRefusal(t *testing.T) {
	root := t.TempDir()
	sourceDir, destination := filepath.Join(root, "source"), filepath.Join(root, "destination")
	if err := os.Mkdir(sourceDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(destination, 0o700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(sourceDir, "auth.json")
	if err := os.WriteFile(source, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	receipt, err := copyAuthNoFollow(source, destination)
	if err != nil || receipt["schema"] != "auth-copy-receipt-v1" {
		t.Fatalf("copy=%v %v", receipt, err)
	}
	if body, readErr := os.ReadFile(filepath.Join(destination, "auth.json")); readErr != nil || string(body) != "secret" {
		t.Fatal("copied bytes differ")
	}
	if err = os.Remove(filepath.Join(destination, "auth.json")); err != nil {
		t.Fatal(err)
	}
	copyAuthHook = func(phase string) {
		if phase == "source-opened" {
			if renameErr := os.Rename(source, source+".old"); renameErr != nil {
				t.Fatal(renameErr)
			}
			if writeErr := os.WriteFile(source, []byte("replacement"), 0o600); writeErr != nil {
				t.Fatal(writeErr)
			}
		}
	}
	defer func() { copyAuthHook = nil }()
	if _, err = copyAuthNoFollow(source, destination); err == nil {
		t.Fatal("source pathname swap was accepted")
	}
	if _, statErr := os.Stat(filepath.Join(destination, "auth.json")); !os.IsNotExist(statErr) {
		t.Fatal("refused copy left destination")
	}
}

func TestCopyAuthRefusesLinkedAncestryAndDestinationCollision(t *testing.T) {
	root := t.TempDir()
	sourceDir := filepath.Join(root, "source")
	destination := filepath.Join(root, "destination")
	if err := os.Mkdir(sourceDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(destination, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "auth.json"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	linked := filepath.Join(root, "linked")
	if err := os.Symlink(sourceDir, linked); err != nil {
		t.Fatal(err)
	}
	if _, err := copyAuthNoFollow(filepath.Join(linked, "auth.json"), destination); err == nil {
		t.Fatal("linked ancestry accepted")
	}
	if err := os.WriteFile(filepath.Join(destination, "auth.json"), []byte("foreign"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := copyAuthNoFollow(filepath.Join(sourceDir, "auth.json"), destination); err == nil {
		t.Fatal("destination collision accepted")
	}
	if body, _ := os.ReadFile(filepath.Join(destination, "auth.json")); string(body) != "foreign" {
		t.Fatal("destination collision changed")
	}
}

func TestMountedDeviceSelectsRequestedMountpointAfterPhysicalStore(t *testing.T) {
	fixture := `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0"><dict><key>system-entities</key><array>
<dict><key>dev-entry</key><string>/dev/disk6s1</string><key>content-hint</key><string>Apple_APFS</string></dict>
<dict><key>dev-entry</key><string>/dev/disk7s1</string><key>mount-point</key><string>/private/tmp/workshop/mount</string></dict>
</array></dict></plist>`
	device, err := mountedDeviceFromPlist(strings.NewReader(fixture), "/private/tmp/workshop/mount")
	if err != nil || device != "/dev/disk7s1" {
		t.Fatalf("device=%q err=%v", device, err)
	}
}

func TestMountedDeviceRefusesMissingAndAmbiguousMountpoints(t *testing.T) {
	missing := `<plist><dict><key>system-entities</key><array><dict><key>dev-entry</key><string>/dev/disk6s1</string></dict></array></dict></plist>`
	if _, err := mountedDeviceFromPlist(strings.NewReader(missing), "/private/tmp/workshop/mount"); err == nil {
		t.Fatal("missing mount point was accepted")
	}
	ambiguous := `<plist><dict><key>system-entities</key><array><dict><key>dev-entry</key><string>/dev/disk7s1</string><key>mount-point</key><string>/private/tmp/workshop/mount</string></dict><dict><key>dev-entry</key><string>/dev/disk8s1</string><key>mount-point</key><string>/private/tmp/workshop/mount</string></dict></array></dict></plist>`
	if _, err := mountedDeviceFromPlist(strings.NewReader(ambiguous), "/private/tmp/workshop/mount"); err == nil {
		t.Fatal("ambiguous mount point was accepted")
	}
}
