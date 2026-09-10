package portable

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func machineFixture(t *testing.T) (*Store, Instance) {
	t.Helper()
	s := fixtureStore(t)
	i, err := Bind(s, filepath.Join(t.TempDir(), "instance"), "friday", "/fixture/my-friday", "device-laptop")
	if err != nil {
		t.Fatal(err)
	}
	fixtureCapability(t, s, "local-tool", nil, `#!/bin/sh
echo SECRET_OUTPUT_CANARY
echo SECRET_ERROR_CANARY >&2
case "$1" in
check) test -f "$MY_FRIDAY_MACHINE_STATE/ready" && exit 0; exit 10 ;;
prepare) printf 'prepared\n' >> "$MY_FRIDAY_MACHINE_STATE/count"; touch "$MY_FRIDAY_MACHINE_STATE/ready" ;;
verify) test -f "$MY_FRIDAY_MACHINE_STATE/ready" ;;
esac
`)
	caps, _ := s.Capabilities()
	c := caps[0]
	c.MachineRequirements = []MachineRequirement{{ID: "runtime", Description: "Synthetic local runtime", Check: []string{"sh", "scripts/run.sh", "check"}, Prepare: []string{"sh", "scripts/run.sh", "prepare"}, Verify: []string{"sh", "scripts/run.sh", "verify"}}}
	if err := writeLocalJSON(filepath.Join(s.Root, "capabilities/local-tool/capability.json"), c); err != nil {
		t.Fatal(err)
	}
	return s, i
}

func TestMachinePreviewPrepareRepeatAndStale(t *testing.T) {
	s, i := machineFixture(t)
	ctx := context.Background()
	status, err := i.MachineStatus(s)
	if err != nil || len(status) != 1 || status[0].State != "unknown" {
		t.Fatalf("%+v %v", status, err)
	}
	if _, err := os.Stat(filepath.Join(i.Root, "machine")); !os.IsNotExist(err) {
		t.Fatal("status wrote local state")
	}
	p, err := i.MachinePlan(s, "local-tool", "runtime")
	if err != nil || p.SHA256 == "" {
		t.Fatalf("%+v %v", p, err)
	}
	r, err := i.RunMachine(ctx, s, p, "check")
	if err != nil || r.State != "needs-preparation" {
		t.Fatalf("%+v %v", r, err)
	}
	for range 2 {
		r, err = i.RunMachine(ctx, s, p, "prepare")
		if err != nil || r.State != "ready" {
			t.Fatalf("%+v %v", r, err)
		}
	}
	b, _ := os.ReadFile(filepath.Join(p.StateDirectory, "count"))
	if string(b) != "prepared\n" {
		t.Fatalf("installer repeated: %s", b)
	}
	status, err = i.MachineStatus(s)
	if err != nil || status[0].State != "ready" || status[0].CheckedAt == "" {
		t.Fatalf("%+v %v", status, err)
	}
	_ = filepath.WalkDir(filepath.Join(i.Root, "machine/receipts"), func(path string, d os.DirEntry, e error) error {
		if e == nil && !d.IsDir() {
			b, _ := os.ReadFile(path)
			if strings.Contains(string(b), "SECRET_") {
				t.Error("retained raw script output")
			}
		}
		return e
	})
	if err := os.WriteFile(filepath.Join(s.Root, "capabilities/local-tool/instructions.md"), []byte("Changed implementation contract."), 0600); err != nil {
		t.Fatal(err)
	}
	status, err = i.MachineStatus(s)
	if err != nil || status[0].State != "stale" {
		t.Fatalf("%+v %v", status, err)
	}
	if _, err := i.RunMachine(ctx, s, p, "prepare"); err == nil {
		t.Fatal("executed changed source after review")
	}
}

func TestMachineFailureDoesNotBecomePermissionToInstall(t *testing.T) {
	for _, code := range []string{"1", "127"} {
		t.Run(code, func(t *testing.T) {
			s, i := machineFixture(t)
			path := filepath.Join(s.Root, "capabilities/local-tool/scripts/run.sh")
			if err := os.WriteFile(path, []byte("#!/bin/sh\n[ \"$1\" = check ] && exit "+code+"\ntouch \"$MY_FRIDAY_MACHINE_STATE/should-not-exist\"\n"), 0700); err != nil {
				t.Fatal(err)
			}
			p, _ := i.MachinePlan(s, "local-tool", "runtime")
			r, err := i.RunMachine(context.Background(), s, p, "prepare")
			if err == nil || r.State != "failed" || len(r.Phases) != 1 {
				t.Fatalf("%+v %v", r, err)
			}
			if _, err := os.Stat(filepath.Join(p.StateDirectory, "should-not-exist")); !os.IsNotExist(err) {
				t.Fatal("failed probe ran preparation")
			}
		})
	}
}

func TestMachineCancellationAndVerifyFailure(t *testing.T) {
	for _, script := range []string{
		"[ \"$1\" = check ] && sleep 30\n",
		"[ \"$1\" = check ] && exit 10\n[ \"$1\" = prepare ] && exit 0\nexit 1\n",
	} {
		s, i := machineFixture(t)
		if err := os.WriteFile(filepath.Join(s.Root, "capabilities/local-tool/scripts/run.sh"), []byte(script), 0700); err != nil {
			t.Fatal(err)
		}
		p, _ := i.MachinePlan(s, "local-tool", "runtime")
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		r, err := i.RunMachine(ctx, s, p, "prepare")
		cancel()
		if err == nil || r.State != "failed" {
			t.Fatalf("%+v %v", r, err)
		}
		status, err := i.MachineStatus(s)
		if err != nil || status[0].State != "failed" {
			t.Fatalf("%+v %v", status, err)
		}
	}
}

func TestMachineManifestAndLocalPathBoundaries(t *testing.T) {
	s, i := machineFixture(t)
	p, _ := i.MachinePlan(s, "local-tool", "runtime")
	external := t.TempDir()
	if err := os.Symlink(external, filepath.Join(i.Root, "machine")); err != nil {
		t.Fatal(err)
	}
	if _, err := i.RunMachine(context.Background(), s, p, "prepare"); err == nil {
		t.Fatal("followed local state symlink")
	}
	entries, _ := os.ReadDir(external)
	if len(entries) != 0 {
		t.Fatal("wrote outside instance")
	}
	caps, _ := s.Capabilities()
	for _, mutate := range []func(*Capability){
		func(c *Capability) { c.MachineRequirements[0].Check = nil },
		func(c *Capability) { c.MachineRequirements[0].TimeoutSeconds = 3601 },
		func(c *Capability) { c.MachineRequirements = append(c.MachineRequirements, c.MachineRequirements[0]) },
		func(c *Capability) { c.MachineRequirements[0].Verify = []string{""} },
	} {
		b, _ := json.Marshal(caps[0])
		var c Capability
		_ = json.Unmarshal(b, &c)
		mutate(&c)
		if err := writeLocalJSON(filepath.Join(s.Root, "capabilities/local-tool/capability.json"), c); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Capabilities(); err == nil {
			t.Fatal("accepted invalid machine contract")
		}
	}
}

func TestMachineDoctorRepairAndAnotherInstanceDoNotPrepare(t *testing.T) {
	s, i := machineFixture(t)
	r := i.Doctor(s, "codex")
	if len(r.MachineRequirements) != 1 || r.MachineRequirements[0].State != "unknown" {
		t.Fatalf("%+v", r)
	}
	if err := i.Project(s); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(i.Root, "machine")); !os.IsNotExist(err) {
		t.Fatal("doctor or projection ran preparation")
	}
	p, _ := i.MachinePlan(s, "local-tool", "runtime")
	if _, err := i.RunMachine(context.Background(), s, p, "prepare"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddDevice(Device{Version: 1, ID: "device-second", Label: "Second synthetic machine"}); err != nil {
		t.Fatal(err)
	}
	other, err := Bind(s, filepath.Join(t.TempDir(), "instance"), "friday", "/fixture/my-friday", "device-second")
	if err != nil {
		t.Fatal(err)
	}
	status, err := other.MachineStatus(s)
	if err != nil || len(status) != 1 || status[0].State != "unknown" {
		t.Fatalf("another instance inherited readiness: %+v %v", status, err)
	}
	if status[0].StateDirectory == p.StateDirectory {
		t.Fatal("instances share state")
	}
	if err := i.Project(s); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(p.StateDirectory, "count"))
	if string(b) != "prepared\n" {
		t.Fatal("repair changed local state")
	}
}

func TestMachineBusyRecursiveAndCorruptReceiptRefused(t *testing.T) {
	s, i := machineFixture(t)
	p, _ := i.MachinePlan(s, "local-tool", "runtime")
	dir, err := i.machineDirectory("", true)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(filepath.Join(dir, "run.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}
	if _, err := i.RunMachine(context.Background(), s, p, "prepare"); err == nil || !strings.Contains(err.Error(), "active") {
		t.Fatalf("busy not refused: %v", err)
	}
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	t.Setenv("MY_FRIDAY_MACHINE_ACTIVE", "1")
	if _, err := i.RunMachine(context.Background(), s, p, "prepare"); err == nil || !strings.Contains(err.Error(), "recursive") {
		t.Fatalf("recursive not refused: %v", err)
	}
	t.Setenv("MY_FRIDAY_MACHINE_ACTIVE", "")
	dir, err = i.machineDirectory("receipts/local-tool", true)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "runtime.json")
	if err := os.WriteFile(path, []byte("corrupt-but-preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := i.RunMachine(context.Background(), s, p, "prepare"); err == nil {
		t.Fatal("overwrote unreadable receipt")
	}
	b, _ := os.ReadFile(path)
	if string(b) != "corrupt-but-preserve" {
		t.Fatal("receipt destroyed")
	}
}
