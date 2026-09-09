package main

import (
	"bytes"
	"encoding/json"
	"github.com/acoz-labs/my-friday/internal/portable"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupResumeLocalPreservesInstallation(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "agent")
	state := filepath.Join(base, "instance")
	var out bytes.Buffer
	if err := runPortable([]string{"setup", "--repository", root, "--state", state, "--name", "friday", "--device-label", "Fixture", "--no-launcher"}, strings.NewReader(""), &out, &out); err != nil {
		t.Fatal(err)
	}
	i, _, err := portable.LoadInstance(state)
	if err != nil {
		t.Fatal(err)
	}
	binding, _ := os.ReadFile(filepath.Join(state, "binding.json"))
	native := filepath.Join(state, "codex", "auth.json")
	os.WriteFile(native, []byte("synthetic unchanged login"), 0600)
	out.Reset()
	if err := runPortable([]string{"setup", "--instance", state}, strings.NewReader("local\n"), &out, &out); err != nil {
		t.Fatal(err)
	}
	after, _, _ := portable.LoadInstance(state)
	got, _ := os.ReadFile(filepath.Join(state, "binding.json"))
	auth, _ := os.ReadFile(native)
	if !bytes.Equal(got, binding) || after.DeviceID != i.DeviceID || string(auth) != "synthetic unchanged login" {
		t.Fatal("resume changed native state")
	}
	if !strings.Contains(out.String(), "local") {
		t.Fatal(out.String())
	}
	out.Reset()
	if err := runPortable([]string{"setup", "--instance", state}, strings.NewReader(""), &out, &out); err == nil {
		t.Fatal("EOF treated as confirmation")
	}
	var packet map[string]any
	json.Unmarshal(binding, &packet)
	packet["binary"] = "/missing/old-toolkit"
	changed, _ := json.Marshal(packet)
	os.WriteFile(filepath.Join(state, "binding.json"), changed, 0600)
	out.Reset()
	if err := runPortable([]string{"setup", "--instance", state}, strings.NewReader("local\n"), &out, &out); err == nil {
		t.Fatal("setup accepted another bound toolkit")
	}
}

func TestNewInteractiveSetupOffersRemoteStep(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "source")
	state := filepath.Join(base, "instance")
	var out bytes.Buffer
	input := "new\nfriday\n" + root + "\ncodex\nFixture laptop\nlocal\n"
	err := runPortable([]string{"setup", "--state", state, "--no-launcher"}, strings.NewReader(input), &out, &out)
	if err != nil {
		t.Fatal(err, out.String())
	}
	if !strings.Contains(out.String(), "Remote setup:") {
		t.Fatal("interactive setup omitted source hosting")
	}
	if _, _, err := portable.LoadInstance(state); err != nil {
		t.Fatal(err)
	}
}
