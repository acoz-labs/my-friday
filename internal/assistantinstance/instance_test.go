package assistantinstance

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func fixture(t *testing.T) (string, string, string) {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".local", "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(home, "my-friday")
	codex := filepath.Join(home, "codex")
	for path, body := range map[string]string{exe: "launcher", codex: "codex"} {
		if err := os.WriteFile(path, []byte(body), 0700); err != nil {
			t.Fatal(err)
		}
	}
	return home, exe, codex
}

func TestFindCodexResolvesPATHSymlink(t *testing.T) {
	home, _, codex := fixture(t)
	bin := filepath.Join(home, "path-bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(codex, filepath.Join(bin, "codex")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	got, err := FindCodex()
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.EvalSymlinks(codex)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %q want resolved %q", got, want)
	}
}

func TestNameAndPaths(t *testing.T) {
	for _, valid := range []string{"alfred", "friday-2", "a"} {
		if err := ValidateName(valid); err != nil {
			t.Errorf("%s: %v", valid, err)
		}
	}
	for _, invalid := range []string{"", "Alfred", "álfred", "-alfred", "alfred-", "a/b", "..", "codex", strings.Repeat("a", 33)} {
		if ValidateName(invalid) == nil {
			t.Errorf("accepted %q", invalid)
		}
	}
	p, err := Derive("/Users/test", "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if p.Root != "/Users/test/.my-friday/assistants/alfred" || p.Launcher != "/Users/test/.local/bin/alfred" {
		t.Fatalf("wrong paths: %#v", p)
	}
}

func TestCreateVerifyRemovePreservesSiblings(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission contract")
	}
	home, exe, codex := fixture(t)
	sibling := filepath.Join(home, ".local", "bin", "keep")
	if err := os.WriteFile(sibling, []byte("keep"), 0700); err != nil {
		t.Fatal(err)
	}
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	m, err := Verify(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if m.Root != p.Paths.Root || m.Launcher != p.Paths.Launcher {
		t.Fatal("manifest paths changed")
	}
	if _, err = PlanCreate(home, "alfred", exe, codex); err == nil {
		t.Fatal("repeat create did not refuse collision")
	}
	rp, err := PlanRemove(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if err = Remove(rp); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(sibling); err != nil || string(b) != "keep" {
		t.Fatal("launcher sibling changed")
	}
	if _, err = os.Stat(p.Paths.Root); !os.IsNotExist(err) {
		t.Fatal("root remains")
	}
	entries, err := os.ReadDir(filepath.Join(home, ".my-friday", "assistants"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), "alfred") {
			t.Fatalf("per-name artifact remains after removal: %s", entry.Name())
		}
	}
}

func TestUpgradeV1InstanceProjectsManifestBoundBuilder(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(p.Paths.Root, "manifest.json")
	body, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var m Manifest
	if err = json.Unmarshal(body, &m); err != nil {
		t.Fatal(err)
	}
	m.ContractVersion = 1
	m.Owned = []string{"codex", "dependencies", "memory", "runtime", "workspace"}
	m.CapabilityBuilder = ""
	m.CapabilityBuilderSHA256 = ""
	m.CapabilityPolicySHA256 = ""
	body, _ = json.MarshalIndent(m, "", "  ")
	body = append(body, '\n')
	if err = os.WriteFile(manifestPath, body, 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.RemoveAll(filepath.Join(p.Paths.Root, "capabilities")); err != nil {
		t.Fatal(err)
	}
	if err = os.RemoveAll(filepath.Join(p.Paths.Root, "workspace", ".agents")); err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(home, "alfred"); err != nil {
		t.Fatalf("v1 compatibility: %v", err)
	}
	upgrade, err := PlanUpgrade(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if err = Upgrade(upgrade); err != nil {
		t.Fatal(err)
	}
	upgraded, err := Verify(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if upgraded.ContractVersion != 2 || upgraded.CapabilityBuilderSHA256 == "" {
		t.Fatalf("not upgraded: %#v", upgraded)
	}
	if err = os.WriteFile(filepath.Join(upgraded.CapabilityBuilder, "SKILL.md"), []byte("drift"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(home, "alfred"); err == nil || !strings.Contains(err.Error(), "builder drift") {
		t.Fatalf("builder drift accepted: %v", err)
	}
}

func TestForeignLauncherAndDriftFailClosed(t *testing.T) {
	home, exe, codex := fixture(t)
	foreign := filepath.Join(home, ".local", "bin", "alfred")
	if err := os.WriteFile(foreign, []byte("foreign"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := PlanCreate(home, "alfred", exe, codex); err == nil {
		t.Fatal("foreign launcher accepted")
	}
	if _, err := os.Stat(filepath.Join(home, ".my-friday")); !os.IsNotExist(err) {
		t.Fatal("collision mutated root")
	}
	if err := os.Remove(foreign); err != nil {
		t.Fatal(err)
	}
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(p.Paths.Launcher, []byte("drift"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err = PlanRemove(home, "alfred"); err == nil {
		t.Fatal("drifted launcher granted deletion authority")
	}
	if _, err = os.Stat(p.Paths.Root); err != nil {
		t.Fatal("drift removed root")
	}
}

func TestPostPreviewForeignCollisionLeavesNoInstanceControlArtifact(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(p.Paths.Launcher, []byte("foreign"), 0700); err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err == nil {
		t.Fatal("post-preview collision accepted")
	}
	if _, err = os.Stat(filepath.Join(home, ".my-friday")); !os.IsNotExist(err) {
		t.Fatalf("collision created control state: %v", err)
	}
}

func TestRemovePathReplacementPreservesForeignRootExactly(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	removePlan, err := PlanRemove(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	original := p.Paths.Root + ".original"
	if err = os.Rename(p.Paths.Root, original); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(p.Paths.Root, 0700); err != nil {
		t.Fatal(err)
	}
	canary := filepath.Join(p.Paths.Root, "foreign-canary")
	if err = os.WriteFile(canary, []byte("unchanged"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = Remove(removePlan); err == nil {
		t.Fatal("replaced root accepted")
	}
	entries, err := os.ReadDir(p.Paths.Root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "foreign-canary" {
		t.Fatalf("foreign entries changed: %v", entries)
	}
	if b, err := os.ReadFile(canary); err != nil || string(b) != "unchanged" {
		t.Fatalf("foreign content changed: %q %v", b, err)
	}
}

func TestRecoverForeignRootPreservesEntriesAndContent(t *testing.T) {
	home, _, _ := fixture(t)
	p, err := Derive(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(p.Root, 0700); err != nil {
		t.Fatal(err)
	}
	canary := filepath.Join(p.Root, "foreign-canary")
	if err = os.WriteFile(canary, []byte("unchanged"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = Recover(home, "alfred"); err == nil {
		t.Fatal("foreign recovery root accepted")
	}
	entries, err := os.ReadDir(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "foreign-canary" {
		t.Fatalf("foreign entries changed: %v", entries)
	}
	if b, err := os.ReadFile(canary); err != nil || string(b) != "unchanged" {
		t.Fatalf("foreign content changed: %q %v", b, err)
	}
}

func TestForgedCodexExecutableIsDenied(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(p.Paths.Root, "manifest.json")
	b, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var m Manifest
	if err = json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	m.CodexExecutable = "/bin/echo"
	b, err = json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(manifestPath, append(b, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(home, "alfred"); err == nil || !strings.Contains(err.Error(), "manifest ownership contract mismatch") {
		t.Fatalf("forged executable accepted: %v", err)
	}
}

func TestTwoInstancesAreIndependent(t *testing.T) {
	home, exe, codex := fixture(t)
	for _, name := range []string{"alfred", "robin"} {
		p, err := PlanCreate(home, name, exe, codex)
		if err != nil {
			t.Fatal(err)
		}
		if err = Create(p, exe, codex); err != nil {
			t.Fatal(err)
		}
	}
	rp, err := PlanRemove(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if err = Remove(rp); err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(home, "robin"); err != nil {
		t.Fatal("sibling instance changed:", err)
	}
}

func TestCreateTrustsOnlyExactInstanceWorkspaceWithTOMLEscaping(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "home \"quoted\" \\ path")
	if err := os.MkdirAll(filepath.Join(home, ".local", "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	exe, codex := filepath.Join(home, "my-friday"), filepath.Join(home, "codex")
	for path, body := range map[string]string{exe: "launcher", codex: "codex"} {
		if err := os.WriteFile(path, []byte(body), 0700); err != nil {
			t.Fatal(err)
		}
	}
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	config, err := os.ReadFile(filepath.Join(p.Paths.Root, "codex", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	escapedWorkspace := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(filepath.Join(p.Paths.Root, "workspace"))
	want := "[projects.\"" + escapedWorkspace + "\"]\ntrust_level = \"trusted\"\n"
	if string(config) != want {
		t.Fatalf("unexpected instance trust config\nwant: %q\n got: %q", want, config)
	}
	for _, forbidden := range []string{home, p.Paths.Root, "*", "approval_policy", "sandbox_mode"} {
		if forbidden != filepath.Join(p.Paths.Root, "workspace") && strings.Contains(string(config), "projects.\""+forbidden+"\"") {
			t.Fatalf("broader trust scope rendered: %q", forbidden)
		}
	}
}

func TestInstanceWorkspaceTrustIsSeparatedAndTamperFailsClosed(t *testing.T) {
	home, exe, codex := fixture(t)
	paths := make(map[string]Paths)
	for _, name := range []string{"alfred", "robin"} {
		p, err := PlanCreate(home, name, exe, codex)
		if err != nil {
			t.Fatal(err)
		}
		if err = Create(p, exe, codex); err != nil {
			t.Fatal(err)
		}
		paths[name] = p.Paths
	}
	alfredConfig, err := os.ReadFile(filepath.Join(paths["alfred"].Root, "codex", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(alfredConfig), paths["robin"].Root) {
		t.Fatal("one instance trusted another instance")
	}
	configPath := filepath.Join(paths["alfred"].Root, "codex", "config.toml")
	if err = os.WriteFile(configPath, append(alfredConfig, []byte("[projects.\"/tmp\"]\ntrust_level = \"trusted\"\n")...), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(home, "alfred"); err == nil || !strings.Contains(err.Error(), "config") {
		t.Fatalf("tampered trust config accepted: %v", err)
	}
	if _, err = PlanRemove(home, "alfred"); err == nil {
		t.Fatal("tampered trust config granted removal authority")
	}
	if err = os.Remove(paths["alfred"].Launcher); err != nil {
		t.Fatal(err)
	}
	if _, err = Recover(home, "alfred"); err == nil || !strings.Contains(err.Error(), "config") {
		t.Fatalf("tampered trust config granted recovery authority: %v", err)
	}
	if _, err = Verify(home, "robin"); err != nil {
		t.Fatalf("sibling instance changed: %v", err)
	}
}

func TestForgedManagedConfigManifestIsDenied(t *testing.T) {
	for _, mutate := range []func(*Manifest){
		func(m *Manifest) { m.CodexConfig = filepath.Join(m.Root, "codex", "other.toml") },
		func(m *Manifest) { m.CodexConfigSHA256 = strings.Repeat("a", 64) },
	} {
		home, exe, codex := fixture(t)
		p, err := PlanCreate(home, "alfred", exe, codex)
		if err != nil {
			t.Fatal(err)
		}
		if err = Create(p, exe, codex); err != nil {
			t.Fatal(err)
		}
		manifestPath := filepath.Join(p.Paths.Root, "manifest.json")
		b, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatal(err)
		}
		var m Manifest
		if err = json.Unmarshal(b, &m); err != nil {
			t.Fatal(err)
		}
		mutate(&m)
		b, err = json.MarshalIndent(m, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(manifestPath, append(b, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err = Verify(home, "alfred"); err == nil || !strings.Contains(err.Error(), "manifest ownership contract mismatch") {
			t.Fatalf("forged config manifest accepted: %v", err)
		}
	}
}

func TestRecoverCompletesInterruptedRemoval(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(p.Paths.Launcher); err != nil {
		t.Fatal(err)
	}
	result, err := Recover(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if result != "restored absent state" {
		t.Fatal(result)
	}
	if _, err = os.Stat(p.Paths.Root); !os.IsNotExist(err) {
		t.Fatal("recovery retained root")
	}
}

func TestSameNameCreateIsSerialized(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}, 2), make(chan struct{})
	mutationHook = func(phase string) {
		if phase == "create-preflight" {
			entered <- struct{}{}
			<-release
		}
	}
	defer func() { mutationHook = nil }()
	results := make(chan error, 2)
	go func() { results <- Create(p, exe, codex) }()
	<-entered
	go func() { results <- Create(p, exe, codex) }()
	select {
	case <-entered:
		t.Fatal("second same-name create entered mutation while first held lock")
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	err1, err2 := <-results, <-results
	if (err1 == nil) == (err2 == nil) {
		t.Fatalf("want one success and one collision, got %v and %v", err1, err2)
	}
	if _, err = Verify(home, "alfred"); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(p.Paths.Root + ".creating"); !os.IsNotExist(err) {
		t.Fatal("shared stage remains")
	}
}

func TestMigrationSuccessAndCleanupFault(t *testing.T) {
	for _, tc := range []struct {
		name       string
		cleanupErr error
	}{{"success", nil}, {"cleanup-fault", errors.New("injected legacy cleanup fault")}} {
		t.Run(tc.name, func(t *testing.T) {
			home, exe, codex := fixture(t)
			p, err := PlanCreate(home, "alfred", exe, codex)
			if err != nil {
				t.Fatal(err)
			}
			cleanupCalled := false
			err = Migrate(p, exe, codex, func() error {
				cleanupCalled = true
				return tc.cleanupErr
			})
			if !cleanupCalled {
				t.Fatal("legacy cleanup was not called")
			}
			if tc.cleanupErr == nil && err != nil {
				t.Fatal(err)
			}
			if tc.cleanupErr != nil && (err == nil || !strings.Contains(err.Error(), tc.cleanupErr.Error())) {
				t.Fatalf("cleanup fault not retained: %v", err)
			}
			if _, verifyErr := Verify(home, "alfred"); verifyErr != nil {
				t.Fatalf("named replacement not retained: %v", verifyErr)
			}
		})
	}
}

func TestLauncherEnvironmentAndArguments(t *testing.T) {
	if os.Getenv("MY_FRIDAY_LAUNCH_HELPER") == "1" {
		if err := Launch(os.Getenv("MY_FRIDAY_TEST_HOME"), "alfred", []string{"hello world", "--cd", "/foreign"}); err != nil {
			t.Fatal(err)
		}
		return
	}
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".local", "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	capture := filepath.Join(home, "capture")
	codex := filepath.Join(home, "codex-stub")
	stub := "#!/bin/sh\nprintf '%s\\n' \"$HOME\" \"$CODEX_HOME\" \"$@\" > \"$MY_FRIDAY_CAPTURE\"\n"
	if err = os.WriteFile(codex, []byte(stub), 0700); err != nil {
		t.Fatal(err)
	}
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, "-test.run=TestLauncherEnvironmentAndArguments")
	cmd.Env = append(os.Environ(), "MY_FRIDAY_LAUNCH_HELPER=1", "MY_FRIDAY_TEST_HOME="+home, "MY_FRIDAY_CAPTURE="+capture, "HOME="+home, "CODEX_HOME=/foreign/codex")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("launch helper: %v: %s", err, out)
	}
	b, err := os.ReadFile(capture)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Join([]string{home, filepath.Join(p.Paths.Root, "codex"), "--cd", filepath.Join(p.Paths.Root, "workspace"), "--", "hello world", "--cd", "/foreign", ""}, "\n")
	if string(b) != want {
		t.Fatalf("launcher contract\nwant: %q\n got: %q", want, b)
	}
}
