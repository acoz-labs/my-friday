package assistantinstance_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/acoz-labs/my-friday/internal/assistantinstance"
	"github.com/acoz-labs/my-friday/internal/codexhome"
	bootstrap "github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/profile"
	"github.com/acoz-labs/my-friday/internal/repository"
)

func TestMigrationCreatesVerifiedInstanceBeforeManifestProvenLegacyUninstall(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	if err := os.MkdirAll(filepath.Join(home, ".local", "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	runtimeRoot, memoryRoot := filepath.Join(home, "source-runtime"), filepath.Join(home, "source-memory")
	configured, err := profile.New("Friday", "Anthony", "Help with careful work", "balanced", "")
	if err != nil {
		t.Fatal(err)
	}
	repositoryPlan, err := bootstrap.Build(configured, runtimeRoot, memoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err = repository.Create(repositoryPlan, runtimeRoot, memoryRoot); err != nil {
		t.Fatal(err)
	}

	legacyHome := filepath.Join(home, ".codex")
	if err = os.Mkdir(legacyHome, 0700); err != nil {
		t.Fatal(err)
	}
	install, err := codexhome.Plan(codexhome.ActionInstall, runtimeRoot, legacyHome)
	if err != nil {
		t.Fatal(err)
	}
	if err = codexhome.Execute(install); err != nil {
		t.Fatal(err)
	}
	uninstall, err := codexhome.Plan(codexhome.ActionUninstall, "", legacyHome)
	if err != nil {
		t.Fatal(err)
	}

	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	codex := filepath.Join(home, "codex-stub")
	if err = os.WriteFile(codex, []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(home, "codex-code-mode-host"), []byte("host"), 0700); err != nil {
		t.Fatal(err)
	}
	instancePlan, err := assistantinstance.PlanCreate(home, "alfred", executable, codex)
	if err != nil {
		t.Fatal(err)
	}
	instancePlan, err = assistantinstance.WithRepositories(instancePlan, runtimeRoot, memoryRoot, repositoryPlan.AssistantID)
	if err != nil {
		t.Fatal(err)
	}
	if err = assistantinstance.Migrate(instancePlan, executable, codex, func() error {
		return codexhome.Execute(uninstall)
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = assistantinstance.Verify(home, "alfred"); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(legacyHome, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("legacy projection remains: %v", err)
	}
}
