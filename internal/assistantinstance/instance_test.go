package assistantinstance

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
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
