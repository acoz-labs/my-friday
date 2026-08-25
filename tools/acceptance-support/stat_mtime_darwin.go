//go:build darwin

package main

import "syscall"

func statMtime(st *syscall.Stat_t) (int64, int64) {
	return st.Mtimespec.Sec, st.Mtimespec.Nsec
}
