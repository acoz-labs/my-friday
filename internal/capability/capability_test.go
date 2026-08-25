package capability

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writePackage(t *testing.T, root, slug, version string) string {
	t.Helper()
	p := filepath.Join(root, "skills", slug)
	for _, d := range []string{"skill", "tests"} {
		if err := os.MkdirAll(filepath.Join(p, d), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	manifest := `{"contract_version":1,"slug":"` + slug + `","version":"` + version + `","display_name":"Daily brief","summary":"Prepare a concise daily brief","profile":"instruction-only","codex_compatibility":"skills-v1","triggers":["prepare my daily brief"],"inputs":["topics"],"outputs":["brief"],"success_behavior":"Return a concise brief","failure_behavior":"Explain missing input","scripts":"none","dependencies":"none","network":"none","credentials":"none","background":"none","durable_data":"none","publishing":"none"}` + "\n"
	skill := "---\nname: " + slug + "\ndescription: Prepare a concise daily brief when explicitly requested.\n---\n\n# Daily brief\n\nAsk for missing topics, then return a concise brief.\n"
	cases := `{"contract_version":1,"positive_triggers":["prepare my daily brief"],"non_triggers":["hello"],"examples":[{"input":"prepare my daily brief","output_contains":["brief"]}],"required_facts":["explicit invocation only"],"forbidden_effects":["network","credentials","durable-data"]}` + "\n"
	for name, body := range map[string]string{"capability.json": manifest, "skill/SKILL.md": skill, "tests/cases.json": cases} {
		if err := os.WriteFile(filepath.Join(p, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return p
}

func TestValidateStrictInstructionOnlyPackage(t *testing.T) {
	p := writePackage(t, t.TempDir(), "daily-brief", "1.0.0")
	pkg, err := Validate(p)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Manifest.Slug != "daily-brief" || len(pkg.Files) != 3 || pkg.SourceDigest == "" || pkg.ProjectionDigest == "" {
		t.Fatalf("unexpected package: %#v", pkg)
	}
	if pkg.Projection["agents/openai.yaml"] == nil || !strings.Contains(string(pkg.Projection["agents/openai.yaml"]), "allow_implicit_invocation: false") {
		t.Fatal("fixed explicit-invocation policy missing")
	}
}

func TestValidateRejectsProhibitedAndUnsafeEntries(t *testing.T) {
	for _, tc := range []struct {
		name, path string
		body       []byte
		mode       os.FileMode
	}{
		{"scripts", "skill/scripts/run.sh", []byte("exit 0"), 0o600},
		{"user policy", "skill/agents/openai.yaml", []byte("policy: {}"), 0o600},
		{"executable", "skill/assets/run", []byte("x"), 0o700},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := writePackage(t, t.TempDir(), "daily-brief", "1.0.0")
			if err := os.MkdirAll(filepath.Dir(filepath.Join(p, tc.path)), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(p, tc.path), tc.body, tc.mode); err != nil {
				t.Fatal(err)
			}
			if _, err := Validate(p); err == nil {
				t.Fatal("unsafe package accepted")
			}
		})
	}
	t.Run("symlink", func(t *testing.T) {
		p := writePackage(t, t.TempDir(), "daily-brief", "1.0.0")
		if err := os.Symlink("SKILL.md", filepath.Join(p, "skill", "linked")); err != nil {
			t.Fatal(err)
		}
		if _, err := Validate(p); err == nil {
			t.Fatal("symlink accepted")
		}
	})
}

func TestLifecyclePreservesSourceAndRefusesDrift(t *testing.T) {
	root := t.TempDir()
	p := writePackage(t, root, "daily-brief", "1.0.0")
	instance := filepath.Join(root, "instance")
	if err := os.MkdirAll(filepath.Join(instance, "workspace"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	plan, err := Plan(instance, p, ActionInstall)
	if err != nil {
		t.Fatal(err)
	}
	if err = Execute(plan); err != nil {
		t.Fatal(err)
	}
	status, err := Inspect(instance, p)
	if err != nil || status.State != StateInstalledHealthy {
		t.Fatalf("status=%#v err=%v", status, err)
	}
	if err = os.WriteFile(filepath.Join(p, "skill", "SKILL.md"), []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	status, err = Inspect(instance, p)
	if err == nil || status.State != StateDraftInvalid {
		t.Fatalf("invalid source status=%#v err=%v", status, err)
	}
	// Restore source, then prove projection drift is never overwritten.
	p = writePackage(t, root, "daily-brief", "1.1.0")
	projection := filepath.Join(instance, "workspace", ".agents", "skills", "daily-brief", "SKILL.md")
	if err = os.WriteFile(projection, []byte("foreign drift"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err = Plan(instance, p, ActionUpgrade); err == nil {
		t.Fatal("managed drift accepted")
	}
	if err = os.Remove(projection); err != nil {
		t.Fatal(err)
	}
	if _, err = Plan(instance, p, ActionDisable); err == nil {
		t.Fatal("missing managed projection accepted")
	}
	if _, err = os.Stat(p); err != nil {
		t.Fatal("source was not preserved")
	}
}

func TestDisableEnableRemoveCompleteReversal(t *testing.T) {
	root := t.TempDir()
	p := writePackage(t, root, "daily-brief", "1.0.0")
	instance := filepath.Join(root, "instance")
	if err := os.MkdirAll(filepath.Join(instance, "workspace"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	for _, action := range []Action{ActionInstall, ActionDisable, ActionEnable, ActionRemove} {
		pl, err := Plan(instance, p, action)
		if err != nil {
			t.Fatalf("%s: %v", action, err)
		}
		if err = Execute(pl); err != nil {
			t.Fatalf("%s: %v", action, err)
		}
	}
	if _, err := os.Stat(filepath.Join(instance, "workspace", ".agents", "skills", "daily-brief")); !os.IsNotExist(err) {
		t.Fatal("projection remains")
	}
	if _, err := os.Stat(filepath.Join(instance, "capabilities", "daily-brief")); !os.IsNotExist(err) {
		t.Fatal("control state remains")
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatal("source removed")
	}
}
