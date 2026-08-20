//go:build darwin

package environment

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"unsafe"
)

func Check(path string, input *os.File) error {
	version, err := syscall.Sysctl("kern.osproductversion")
	if err != nil {
		return fmt.Errorf("read macOS version: %w", err)
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return fmt.Errorf("inspect target filesystem: %w", err)
	}
	fs := byteString(stat.Fstypename[:])
	git := exec.Command("git", "--version")
	git.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "LANG=C.UTF-8"}
	out, err := git.Output()
	if err != nil {
		return fmt.Errorf("Git 2.28 or later is required: %w", err)
	}
	return validateContract(runtime.GOARCH, version, fs, string(out), isTerminal(input))
}

func isTerminal(input *os.File) bool {
	var state syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, input.Fd(), uintptr(syscall.TIOCGETA), uintptr(unsafe.Pointer(&state)))
	return errno == 0
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
