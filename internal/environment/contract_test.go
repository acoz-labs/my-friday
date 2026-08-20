package environment

import "testing"

func TestSupportedEnvironmentContract(t *testing.T) {
	if err := validateContract("arm64", "14.0", "apfs", "git version 2.28.0", true); err != nil {
		t.Fatal(err)
	}
}

func TestEnvironmentContractDenials(t *testing.T) {
	cases := []struct {
		name, arch, os, fs, git string
		tty                     bool
	}{
		{"architecture", "amd64", "14.0", "apfs", "git version 2.28.0", true},
		{"macOS", "arm64", "13.6", "apfs", "git version 2.28.0", true},
		{"terminal", "arm64", "14.0", "apfs", "git version 2.28.0", false},
		{"filesystem", "arm64", "14.0", "nfs", "git version 2.28.0", true},
		{"git", "arm64", "14.0", "apfs", "git version 2.27.9", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateContract(tc.arch, tc.os, tc.fs, tc.git, tc.tty); err == nil {
				t.Fatal("expected denial")
			}
		})
	}
}
