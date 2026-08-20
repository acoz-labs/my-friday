//go:build darwin

package environment

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

func Check(path string, input *os.File) error {
	version, err := syscall.Sysctl("kern.osproductversion")
	if err != nil {
		return fmt.Errorf("read macOS version: %w", err)
	}
	info, inputErr := input.Stat()
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return fmt.Errorf("inspect target filesystem: %w", err)
	}
	fs := byteString(stat.Fstypename[:])
	out, err := exec.Command("git", "--version").Output()
	if err != nil {
		return fmt.Errorf("Git 2.28 or later is required: %w", err)
	}
	return validateContract(runtime.GOARCH, version, fs, string(out), inputErr == nil && info.Mode()&os.ModeCharDevice != 0)
}

func byteString(value []int8) string {
	b := make([]byte, 0, len(value))
	for _, v := range value {
		if v == 0 {
			break
		}
		b = append(b, byte(v))
	}
	return string(b)
}
