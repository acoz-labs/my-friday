//go:build darwin

package environment

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

func Check(path string, input *os.File) error {
	if runtime.GOARCH != "arm64" {
		return fmt.Errorf("unsupported architecture %s; Apple silicon is required", runtime.GOARCH)
	}
	version, err := syscall.Sysctl("kern.osproductversion")
	if err != nil {
		return fmt.Errorf("read macOS version: %w", err)
	}
	major, _ := strconv.Atoi(strings.Split(version, ".")[0])
	if major < 14 {
		return fmt.Errorf("macOS 14 or later is required; found %s", version)
	}
	if info, err := input.Stat(); err != nil || info.Mode()&os.ModeCharDevice == 0 {
		return fmt.Errorf("init requires an interactive terminal")
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return fmt.Errorf("inspect target filesystem: %w", err)
	}
	fs := byteString(stat.Fstypename[:])
	if fs != "apfs" {
		return fmt.Errorf("local APFS is required; found %s", fs)
	}
	out, err := exec.Command("git", "--version").Output()
	if err != nil {
		return fmt.Errorf("Git 2.28 or later is required: %w", err)
	}
	var maj, min int
	if _, err = fmt.Sscanf(strings.TrimSpace(string(out)), "git version %d.%d", &maj, &min); err != nil || maj < 2 || maj == 2 && min < 28 {
		return fmt.Errorf("Git 2.28 or later is required; found %s", strings.TrimSpace(string(out)))
	}
	return nil
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
