//go:build linux

package codexhome

import "golang.org/x/sys/unix"

// Linux is a source-validation surface only. The command layer refuses the
// installed-baseline lifecycle outside Darwin before opening a Codex home.
func validatePlatformHome(_ int, _ int) error { return nil }

func renameExclusive(fromFD int, from string, toFD int, to string) error {
	return unix.Renameat2(fromFD, from, toFD, to, unix.RENAME_NOREPLACE)
}

func renameSwap(fromFD int, from string, toFD int, to string) error {
	return unix.Renameat2(fromFD, from, toFD, to, unix.RENAME_EXCHANGE)
}
