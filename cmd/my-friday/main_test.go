package main

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/codexhome"
)

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
