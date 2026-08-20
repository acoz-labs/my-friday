//go:build darwin

package environment

import (
	"os"
	"testing"
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
