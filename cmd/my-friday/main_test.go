package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/acoz-labs/my-friday/internal/capability"
	"github.com/acoz-labs/my-friday/internal/capabilityworkshop"
	"github.com/acoz-labs/my-friday/internal/codexhome"
)

func TestWorkshopSignalHelper(t *testing.T) {
	if os.Getenv("MY_FRIDAY_WORKSHOP_SIGNAL_HELPER") != "1" {
		return
	}
	instance := os.Getenv("MY_FRIDAY_WORKSHOP_SIGNAL_INSTANCE")
	if err := capability.InitializeInstance(instance); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(70)
	}
	if err := os.MkdirAll(filepath.Join(instance, "runtime", "skills"), 0700); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(70)
	}
	plan, err := capabilityworkshop.Plan(instance, filepath.Join(instance, "runtime", "skills", "daily-brief"), "daily-brief")
	if err == nil {
		err = runCapabilityWorkshop(plan, os.Stdin, os.Stdout)
	}
	code, _ := classifyError([]string{"my-friday", "capability", "workshop"}, err)
	if err == nil {
		code = 0
	}
	os.Exit(code)
}

func TestRealWorkshopSignalsInterruptPromptWithoutWriting(t *testing.T) {
	if _, err := os.Stat("/usr/bin/expect"); err != nil {
		t.Skip("native workshop signal integration requires Expect")
	}
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	script := filepath.Join("..", "..", "config", "acceptance", "capability-workshop-signal-real.exp")
	for _, mode := range []string{"int", "term"} {
		t.Run(mode, func(t *testing.T) {
			instance := filepath.Join(root, mode)
			cmd := exec.Command("/usr/bin/expect", script, mode, os.Args[0], instance)
			cmd.Env = append(os.Environ(), "MY_FRIDAY_WORKSHOP_SIGNAL_HELPER=1", "MY_FRIDAY_WORKSHOP_SIGNAL_INSTANCE="+instance)
			if output, runErr := cmd.CombinedOutput(); runErr != nil {
				t.Fatalf("real workshop %s signal failed: %v\n%s", mode, runErr, output)
			}
			for _, path := range []string{filepath.Join(instance, "runtime", "skills", "daily-brief"), filepath.Join(instance, "capabilities", ".workshop-daily-brief.json")} {
				if _, statErr := os.Lstat(path); !os.IsNotExist(statErr) {
					t.Fatalf("%s signal changed %s: %v", mode, path, statErr)
				}
			}
		})
	}
}

func TestRealHomeIgnoresCallerHOME(t *testing.T) {
	want, err := realHome()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", t.TempDir())
	got, err := realHome()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("caller HOME became authority: got %q want %q", got, want)
	}
	if os.Getenv("HOME") == want {
		t.Fatal("test did not override child HOME environment")
	}
}

func TestCapabilityTestCLIRefusesContradictorySuite(t *testing.T) {
	pkg := capability.Package{Manifest: capability.Manifest{Slug: "daily-brief", Triggers: []string{"prepare my daily brief"}}, Cases: capability.Cases{ContractVersion: 1, PositiveTriggers: []string{"prepare my daily brief"}, NonTriggers: []string{"prepare my daily brief"}}}
	var out strings.Builder
	if err := reportCapabilityTest(&out, pkg); err == nil {
		t.Fatal("CLI reported contradictory suite passed")
	}
	if out.Len() != 0 {
		t.Fatalf("CLI emitted success output: %q", out.String())
	}
}

func TestAssistantLegacyMigrationHomeUsesAccountBoundary(t *testing.T) {
	accountHome, err := realHome()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEX_HOME", filepath.Join(os.Getenv("HOME"), ".codex"))
	if _, err = codexHomeWithin(accountHome); err == nil {
		t.Fatal("foreign caller HOME granted migration cleanup authority")
	}
}

func TestExitCategories(t *testing.T) {
	cases := []struct {
		args    []string
		message string
		code    int
	}{
		{[]string{"my-friday", "validate"}, "manifest schema: invalid", 6},
		{[]string{"my-friday", "recover"}, "journal mismatch", 5},
		{[]string{"my-friday", "init"}, "target is not empty", 3},
		{[]string{"my-friday", "init"}, "creation failed and was rolled back", 4},
		{[]string{"my-friday", "version", "extra"}, "usage: my-friday version", 2},
	}
	for _, tc := range cases {
		if got, _ := classifyError(tc.args, errors.New(tc.message)); got != tc.code {
			t.Errorf("%v: got %d want %d", tc.args, got, tc.code)
		}
	}
}

func TestVerifyStatusRequiresHealthyState(t *testing.T) {
	states := []codexhome.State{
		codexhome.StateNotInstalled,
		codexhome.StateCollision,
		codexhome.StateDrift,
		codexhome.StateSourceDrift,
		codexhome.StateInterrupted,
	}
	for _, state := range states {
		t.Run(string(state), func(t *testing.T) {
			err := verifyStatus(codexhome.Status{State: state, Detail: "test detail"})
			if err == nil {
				t.Fatal("expected unhealthy state to fail verification")
			}
			if code, stable := classifyError([]string{"my-friday", "codex", "verify"}, err); code != 3 || stable != "codex.state_denied" {
				t.Fatalf("got (%d, %q), want (3, %q)", code, stable, "codex.state_denied")
			}
		})
	}

	if err := verifyStatus(codexhome.Status{State: codexhome.StateHealthy, Detail: "test detail"}); err != nil {
		t.Fatalf("healthy state failed verification: %v", err)
	}
}

func TestReadConfirmationRequiresExactCaseSensitiveNewlineTerminatedToken(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		confirmed bool
		wantErr   error
	}{
		{name: "exact", input: "Install\n", confirmed: true},
		{name: "padded whitespace", input: " Install \n"},
		{name: "wrong case", input: "install\n"},
		{name: "EOF without newline", input: "Install", wantErr: io.EOF},
		{name: "noninteractive empty input", input: "", wantErr: io.EOF},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confirmed, err := readConfirmation(strings.NewReader(tt.input), "Install")
			if confirmed != tt.confirmed {
				t.Fatalf("confirmed = %v, want %v", confirmed, tt.confirmed)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestWorkshopSignalNeverHidesCommitOrRecoveryError(t *testing.T) {
	interrupted := make(chan os.Signal, 1)
	interrupted <- os.Interrupt
	want := errors.New("source recovery required")
	if got := workshopResult(want, interrupted); !errors.Is(got, want) {
		t.Fatalf("signal replaced transaction error: %v", got)
	}
}

func TestWorkshopSignalAfterSuccessfulCommitReturnsStableInterruption(t *testing.T) {
	interrupted := make(chan os.Signal, 1)
	interrupted <- os.Interrupt
	got := workshopResult(nil, interrupted)
	if got == nil || got.Error() != "capability workshop interrupted after source transaction completed" {
		t.Fatalf("successful signaled commit error = %v", got)
	}
	if code, stable := classifyError([]string{"my-friday", "capability", "workshop"}, got); code != 130 || stable != "workshop.interrupted" {
		t.Fatalf("successful signal classification = (%d, %q)", code, stable)
	}
}

func TestWorkshopSignalClassifiesStableStatuses(t *testing.T) {
	for _, test := range []struct {
		signal os.Signal
		code   int
	}{{os.Interrupt, 130}, {syscall.SIGTERM, 143}} {
		code, stable := classifyError([]string{"my-friday", "capability", "workshop"}, workshopInterruptedError{signal: test.signal})
		if code != test.code || stable != "workshop.interrupted" {
			t.Fatalf("%s classification = (%d, %q), want (%d, workshop.interrupted)", test.signal, code, stable, test.code)
		}
	}
}
