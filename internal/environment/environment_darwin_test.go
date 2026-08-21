//go:build darwin

package environment

import (
	"os"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/gitexec"
)

func TestNativeTerminalProbeRejectsArbitraryCharacterDevice(t *testing.T) {
	f, err := os.Open("/dev/null")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if isTerminal(f) {
		t.Fatal("/dev/null must not satisfy the interactive-terminal contract")
	}
}

func TestNativeCheckObservesVersionProbeThroughScrubbedGitBoundary(t *testing.T) {
	original := gitexec.Observe
	defer func() { gitexec.Observe = original }()
	var args, env []string
	gitexec.Observe = func(gotArgs, gotEnv []string) { args, env = gotArgs, gotEnv }
	f, err := os.Open("/dev/null")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_ = Check(t.TempDir(), f)
	if len(args) != 1 || args[0] != "--version" {
		t.Fatalf("unexpected Git version argv: %q", args)
	}
	if len(env) != 3 || !strings.HasPrefix(env[0], "PATH=") || !strings.HasPrefix(env[1], "HOME=") || env[2] != "LANG=C.UTF-8" {
		t.Fatalf("unexpected Git version environment: %q", env)
	}
}
