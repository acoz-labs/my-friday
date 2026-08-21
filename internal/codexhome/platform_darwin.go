//go:build darwin

package codexhome

import (
	"strings"

	"golang.org/x/sys/unix"
)

func validatePlatformHome(fd, euid int) error {
	var fs unix.Statfs_t
	if err := unix.Fstatfs(fd, &fs); err != nil {
		return err
	}
	fsType := strings.TrimRight(string(fs.Fstypename[:]), "\x00")
	return validateHomeEnvironment(euid, fsType, fs.Flags&unix.MNT_LOCAL != 0)
}

func renameExclusive(fromFD int, from string, toFD int, to string) error {
	return unix.RenameatxNp(fromFD, from, toFD, to, unix.RENAME_EXCL)
}

func renameSwap(fromFD int, from string, toFD int, to string) error {
	return unix.RenameatxNp(fromFD, from, toFD, to, unix.RENAME_SWAP)
}
