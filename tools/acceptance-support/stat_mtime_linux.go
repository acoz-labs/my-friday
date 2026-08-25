//go:build linux

package main

import "syscall"

func statMtime(st *syscall.Stat_t) (int64, int64) {
	return st.Mtim.Sec, st.Mtim.Nsec
}
