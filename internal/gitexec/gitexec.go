package gitexec

import (
	"os"
	"os/exec"
)

// Observe exposes the complete, scrubbed production Git boundary to tests.
var Observe = func([]string, []string) {}

func Command(args ...string) *exec.Cmd {
	env := []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "LANG=C.UTF-8"}
	Observe(append([]string(nil), args...), append([]string(nil), env...))
	cmd := exec.Command("git", args...)
	cmd.Env = env
	return cmd
}
