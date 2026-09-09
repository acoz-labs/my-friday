package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/portable"
	"github.com/acoz-labs/my-friday/internal/toolkitupdate"
)

func TestCompiledToolkitInstallRebindAndRollback(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "my-friday")
	build := exec.Command("go", "build", "-o", binary, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("%s %v", out, err)
	}
	if err := verifyToolkit(binary); err != nil {
		t.Fatal(err)
	}
	old := filepath.Join(directory, "previous")
	data, _ := os.ReadFile(binary)
	os.WriteFile(old, data, 0700)
	home := t.TempDir()
	hash, err := toolkitupdate.Digest(binary)
	if err != nil {
		t.Fatal(err)
	}
	installed, err := toolkitupdate.Stage(home, binary, hash)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := toolkitupdate.Activate(home, installed, binary); err != nil {
		t.Fatal(err)
	}
	s, err := portable.Create(filepath.Join(home, "source"), "pilot", "codex", "device-upgrade", "Fixture")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.InitGit(context.Background()); err != nil {
		t.Fatal(err)
	}
	i, err := portable.Bind(s, filepath.Join(home, "instance"), "pilot", old, "device-upgrade")
	if err != nil {
		t.Fatal(err)
	}
	launcher := filepath.Join(home, ".local/bin/pilot")
	if err := i.InstallLauncher(launcher); err != nil {
		t.Fatal(err)
	}
	native := filepath.Join(i.Root, "codex/config.toml")
	original, _ := os.ReadFile(native)
	raw, err := runToolkit(installed, "toolkit", "use", "--instance", i.Root, "--launcher", launcher)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Backup string `json:"backup"`
	}
	if err := json.Unmarshal(raw, &result); err != nil || result.Backup == "" {
		t.Fatalf("%s %v", raw, err)
	}
	after, _, err := portable.LoadInstance(i.Root)
	if err != nil || after.Binary != installed {
		t.Fatalf("%+v %v", after, err)
	}
	if _, err := runToolkit(old, "toolkit", "check-instance", "--instance", i.Root); err != nil {
		t.Fatal(err)
	}
	if err := after.RollbackToolkit(s, result.Backup); err != nil {
		t.Fatal(err)
	}
	after, _, _ = portable.LoadInstance(i.Root)
	if after.Binary != old {
		t.Fatal("rollback did not restore pin")
	}
	actual, _ := os.ReadFile(native)
	if !bytes.Equal(actual, original) {
		t.Fatal("native config changed")
	}
	// Changing source compatibility prevents old writers from claiming readiness.
	os.WriteFile(filepath.Join(s.Root, ".my-friday/sync.json"), []byte(`{"schema_version":999}`), 0600)
	if _, err := runToolkit(old, "toolkit", "check-instance", "--instance", i.Root); err == nil {
		t.Fatal("accepted incompatible source")
	}
}

func TestUpdateMenuDeclineDoesNotInstall(t *testing.T) {
	home := t.TempDir()
	var out bytes.Buffer
	u := managementUI{promptReader(strings.NewReader("2\n/not/read\nnot-a-digest\nno\n0\n")), &out, home}
	if err := u.updates(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".local")); !os.IsNotExist(err) {
		t.Fatal("decline created files")
	}
}

func TestManagementCommandsHelpAndVersionArguments(t *testing.T) {
	for _, command := range []string{"menu", "toolkit", "version"} {
		var out bytes.Buffer
		if err := runPortable([]string{command, "--help"}, strings.NewReader(""), &out, &out); err != nil || !strings.Contains(out.String(), "Usage:") {
			t.Fatalf("%s: %s %v", command, out.String(), err)
		}
	}
	var out bytes.Buffer
	if err := runPortable([]string{"version", "unexpected"}, strings.NewReader(""), &out, &out); err == nil {
		t.Fatal("silently ignored version argument")
	}
}
