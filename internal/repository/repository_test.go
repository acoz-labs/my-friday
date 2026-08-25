package repository

import (
	"errors"
	"github.com/acoz-labs/my-friday/internal/gitexec"
	"github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/profile"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionGitCommandsUseExactArgvAndEnvironment(t *testing.T) {
	original := gitexec.Observe
	defer func() { gitexec.Observe = original }()
	var observed [][]string
	gitexec.Observe = func(args, env []string) {
		if len(env) != 3 || !strings.HasPrefix(env[0], "PATH=") || !strings.HasPrefix(env[1], "HOME=") || env[2] != "LANG=C.UTF-8" {
			t.Fatalf("unexpected Git environment: %q", env)
		}
		observed = append(observed, args)
	}
	root := t.TempDir()
	p, _ := profile.New("Friday", "", "Help", "balanced", "")
	pl, _ := plan.Build(p, filepath.Join(root, "runtime"), filepath.Join(root, "memory"))
	if err := Create(pl, pl.Targets.Runtime, pl.Targets.Memory); err != nil {
		t.Fatal(err)
	}
	if !ExactBaseline(pl, pl.Targets.Runtime, pl.Targets.Memory) {
		t.Fatal("baseline validation failed")
	}
	if len(observed) == 0 {
		t.Fatal("no Git commands captured")
	}
	for _, args := range observed {
		allowed := exactAllowedGitArgs(args)
		if !allowed {
			t.Fatalf("non-allowlisted Git argv: %q", args)
		}
	}
}

func exactAllowedGitArgs(args []string) bool {
	if len(args) == 5 && args[0] == "init" && args[1] == "--quiet" && args[2] == "--initial-branch=main" && strings.HasPrefix(args[3], "--template=") && args[4] != "" {
		return true
	}
	if len(args) < 3 || args[0] != "-C" || args[1] == "" {
		return false
	}
	tail := strings.Join(args[2:], "\x00")
	switch tail {
	case "rev-parse\x00--git-dir", "symbolic-ref\x00--short\x00HEAD", "rev-parse\x00--verify\x00HEAD", "remote", "config\x00--local\x00--null\x00--list":
		return true
	}
	if len(args) == 6 && args[2] == "config" && args[3] == "--local" {
		expected := map[string]string{
			"core.repositoryformatversion": "0", "core.filemode": "true", "core.bare": "false",
			"core.logallrefupdates": "true", "core.ignorecase": "false", "core.precomposeunicode": "false",
		}
		return expected[args[4]] == args[5]
	}
	return false
}

func TestCreateAndValidatePairWithoutCommitOrRemote(t *testing.T) {
	root := t.TempDir()
	p, _ := profile.New("Friday", "", "Help", "balanced", "")
	pl, _ := plan.Build(p, filepath.Join(root, "runtime"), filepath.Join(root, "memory"))
	if err := Create(pl, pl.Targets.Runtime, pl.Targets.Memory); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePair(pl.Targets.Runtime, pl.Targets.Memory); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{pl.Targets.Runtime, pl.Targets.Memory} {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
			t.Fatal(err)
		}
		out, _ := exec.Command("git", "-C", dir, "remote").Output()
		if len(out) != 0 {
			t.Fatal("unexpected remote")
		}
		if exec.Command("git", "-C", dir, "rev-parse", "HEAD").Run() == nil {
			t.Fatal("unexpected commit")
		}
	}
}

func TestCapabilitySourceRollbackRequiresEmptySource(t *testing.T) {
	root := t.TempDir()
	p, _ := profile.New("Friday", "", "Help", "balanced", "")
	pl, _ := plan.Build(p, filepath.Join(root, "runtime"), filepath.Join(root, "memory"))
	if err := Create(pl, pl.Targets.Runtime, pl.Targets.Memory); err != nil {
		t.Fatal(err)
	}
	if err := RollbackCapabilities(pl.Targets.Runtime); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(pl.Targets.Runtime, "skills", ".gitkeep")); err != nil {
		t.Fatal(err)
	}
	if err := InitializeCapabilities(pl.Targets.Runtime); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(pl.Targets.Runtime, "skills", "daily-brief"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := RollbackCapabilities(pl.Targets.Runtime); err == nil || !strings.Contains(err.Error(), "source exists") {
		t.Fatalf("non-empty rollback accepted: %v", err)
	}
}

func TestCapabilityMigrationFaultsRequireAndCompleteRecovery(t *testing.T) {
	root := t.TempDir()
	p, _ := profile.New("Friday", "", "Help", "balanced", "")
	pl, _ := plan.Build(p, filepath.Join(root, "runtime"), filepath.Join(root, "memory"))
	if err := Create(pl, pl.Targets.Runtime, pl.Targets.Memory); err != nil {
		t.Fatal(err)
	}
	capabilityMigrationHook = func(phase string) error {
		if phase == "rollback-applied" {
			return errors.New("injected stop")
		}
		return nil
	}
	if err := RollbackCapabilities(pl.Targets.Runtime); err == nil {
		t.Fatal("fault missing")
	}
	capabilityMigrationHook = nil
	if _, err := ValidateRuntime(pl.Targets.Runtime); err == nil {
		t.Fatal("journal did not block validation")
	}
	if err := RecoverCapabilities(pl.Targets.Runtime); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(pl.Targets.Runtime, "skills", ".gitkeep")); err != nil {
		t.Fatal(err)
	}
	capabilityMigrationHook = func(phase string) error {
		if phase == "initialize-applied" {
			return errors.New("injected stop")
		}
		return nil
	}
	if err := InitializeCapabilities(pl.Targets.Runtime); err == nil {
		t.Fatal("fault missing")
	}
	capabilityMigrationHook = nil
	if err := RecoverCapabilities(pl.Targets.Runtime); err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateRuntime(pl.Targets.Runtime); err != nil {
		t.Fatal(err)
	}
}

func TestValidationAuthenticatesSchemaBeforeCompilationAndChecksGit(t *testing.T) {
	root := t.TempDir()
	p, _ := profile.New("Friday", "", "Help", "balanced", "")
	pl, _ := plan.Build(p, filepath.Join(root, "runtime"), filepath.Join(root, "memory"))
	if err := Create(pl, pl.Targets.Runtime, pl.Targets.Memory); err != nil {
		t.Fatal(err)
	}
	schemaPath := filepath.Join(pl.Targets.Runtime, ".my-friday/schemas/repository-manifest.v1.schema.json")
	if err := os.WriteFile(schemaPath, []byte(`{"$ref":"https://example.invalid/foreign.json"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePair(pl.Targets.Runtime, pl.Targets.Memory); err == nil || !strings.Contains(err.Error(), "differs from the embedded") {
		t.Fatalf("err=%v", err)
	}
	if err := os.WriteFile(schemaPath, []byte(plan.ManifestSchema()), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(pl.Targets.Runtime, ".git")); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePair(pl.Targets.Runtime, pl.Targets.Memory); err == nil || !strings.Contains(err.Error(), "not a local Git") {
		t.Fatalf("err=%v", err)
	}
}

func TestOrdinaryValidationRoutesRetainedCreationMarkerToRecovery(t *testing.T) {
	root := t.TempDir()
	p, _ := profile.New("Friday", "", "Help", "balanced", "")
	pl, _ := plan.Build(p, filepath.Join(root, "runtime"), filepath.Join(root, "memory"))
	if err := Create(pl, pl.Targets.Runtime, pl.Targets.Memory); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{pl.Targets.Runtime, pl.Targets.Memory} {
		marker := filepath.Join(target, ".my-friday", "creation-state.json")
		if err := os.WriteFile(marker, []byte(pl.PlanID+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := ValidatePair(pl.Targets.Runtime, pl.Targets.Memory); err == nil || !strings.Contains(err.Error(), "unknown owned contract path") {
		t.Fatalf("err=%v", err)
	}
	if err := ValidateFreshPair(pl.Targets.Runtime, pl.Targets.Memory); err != nil {
		t.Fatalf("transaction validation: %v", err)
	}
}

func TestExactBaselineRejectsGitTypeAndModeDrift(t *testing.T) {
	root := t.TempDir()
	p, _ := profile.New("Friday", "", "Help", "balanced", "")
	pl, _ := plan.Build(p, filepath.Join(root, "runtime"), filepath.Join(root, "memory"))
	if err := Create(pl, pl.Targets.Runtime, pl.Targets.Memory); err != nil {
		t.Fatal(err)
	}
	if !ExactBaseline(pl, pl.Targets.Runtime, pl.Targets.Memory) {
		t.Fatal("fresh repositories did not match exact baseline")
	}
	head := filepath.Join(pl.Targets.Runtime, ".git", "HEAD")
	if err := os.Chmod(head, 0700); err != nil {
		t.Fatal(err)
	}
	if ExactBaseline(pl, pl.Targets.Runtime, pl.Targets.Memory) {
		t.Fatal("executable Git HEAD accepted as exact baseline")
	}
	if err := os.Chmod(head, 0600); err != nil {
		t.Fatal(err)
	}
	hooks := filepath.Join(pl.Targets.Runtime, ".git", "hooks")
	if err := os.Symlink("objects", hooks); err != nil {
		t.Fatal(err)
	}
	if ExactBaseline(pl, pl.Targets.Runtime, pl.Targets.Memory) {
		t.Fatal("symlinked Git directory accepted as exact baseline")
	}
}
