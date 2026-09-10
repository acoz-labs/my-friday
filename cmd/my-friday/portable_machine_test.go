package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/portable"
)

func commandMachineFixture(t *testing.T) (portable.Instance, *portable.Store) {
	t.Helper()
	i, s := apiFixture(t)
	writeCommandMachineFixture(t, s)
	return i, s
}

func writeCommandMachineFixture(t *testing.T, s *portable.Store) {
	t.Helper()
	dir := filepath.Join(s.Root, "capabilities/fixture-tool")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	c := portable.Capability{Version: 1, ID: "fixture-tool", Description: "Synthetic machine tool", Subscriptions: []portable.Subscription{}, MachineRequirements: []portable.MachineRequirement{{ID: "runtime", Description: "Synthetic runtime", Check: []string{"sh", "run.sh", "check"}, Prepare: []string{"sh", "run.sh", "prepare"}, Verify: []string{"sh", "run.sh", "verify"}}}}
	b, _ := json.Marshal(c)
	if err := os.WriteFile(filepath.Join(dir, "capability.json"), b, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "run.sh"), []byte(`echo RAW_OUTPUT_CANARY
case "$1" in
check) test -f "$MY_FRIDAY_MACHINE_STATE/ready" && exit 0; exit 10 ;;
prepare) touch "$MY_FRIDAY_MACHINE_STATE/ready" ;;
verify) test -f "$MY_FRIDAY_MACHINE_STATE/ready" ;;
esac
`), 0700); err != nil {
		t.Fatal(err)
	}
}

func TestMachineCLIPlanApplyAndNoPrompt(t *testing.T) {
	i, s := commandMachineFixture(t)
	var out, errout bytes.Buffer
	args := []string{"machine", "prepare", "--instance", i.Root, "--capability", "fixture-tool", "--requirement", "runtime"}
	if err := runPortable(args, strings.NewReader(""), &out, &errout); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"state": "preview"`) {
		t.Fatal(out.String())
	}
	if _, err := os.Stat(filepath.Join(i.Root, "machine")); !os.IsNotExist(err) {
		t.Fatal("preview created state")
	}
	if err := runPortable(append(args, "--apply"), strings.NewReader(""), &out, &errout); err == nil {
		t.Fatal("applied without reviewed fingerprint")
	}
	p, _ := i.MachinePlan(s, "fixture-tool", "runtime")
	out.Reset()
	if err := runPortable(append(args, "--apply", "--expect-sha256", p.SHA256), strings.NewReader(""), &out, &errout); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"state": "ready"`) || strings.Contains(out.String()+errout.String(), "RAW_OUTPUT_CANARY") {
		t.Fatal(out.String(), errout.String())
	}
	for _, args := range [][]string{{"machine", "status", "--instance", i.Root, "--apply"}, {"machine", "check", "--instance", i.Root, "--capability", "fixture-tool", "--requirement", "runtime", "--interactive"}} {
		if err := runPortable(args, strings.NewReader(""), &out, &errout); err == nil {
			t.Fatal("accepted irrelevant flag")
		}
	}
}

func TestMachineMenuReviewCancellationAndExplicitPreparation(t *testing.T) {
	i, s := commandMachineFixture(t)
	var out bytes.Buffer
	// Requirement, prepare, decline, leave.
	u := newManagementUI(t.TempDir(), strings.NewReader("1\n2\nno\n0\n"), &out, true)
	if err := u.machine(i.Root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(i.Root, "machine")); !os.IsNotExist(err) {
		t.Fatal("cancel ran scripts")
	}
	u = newManagementUI(t.TempDir(), strings.NewReader("1\n2\nyes\n0\n"), &out, true)
	if err := u.machine(i.Root); err != nil {
		t.Fatal(err)
	}
	status, err := i.MachineStatus(s)
	if err != nil || status[0].State != "ready" {
		t.Fatalf("%+v %v", status, err)
	}
	for _, want := range []string{"Prepare this machine", "What will change", "ready"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %s", want)
		}
	}
	if strings.Contains(out.String(), "RAW_OUTPUT_CANARY") {
		t.Fatal("raw script output reached menu")
	}
}
