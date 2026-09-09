package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/acoz-labs/my-friday/internal/portable"
	"github.com/charmbracelet/x/term"
	"golang.org/x/sys/unix"
)

// The PTY subprocess uses a disposable home explicitly; it never discovers or
// operates on the developer's real installations or native configuration.
func TestManagementPTYHelper(t *testing.T) {
	if os.Getenv("MY_FRIDAY_TEST_UI_HELPER") != "1" {
		return
	}
	home := os.Getenv("MY_FRIDAY_TEST_UI_HOME")
	if !filepath.IsAbs(home) {
		t.Fatal("missing isolated test home")
	}
	before, err := term.GetState(os.Stdin.Fd())
	if err != nil {
		t.Fatal(err)
	}
	s, err := portable.Create(filepath.Join(home, ".local/share/my-friday/repositories/fixture"), "fixture", "codex", "device-tui", "Fixture")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.InitGit(context.Background()); err != nil {
		t.Fatal(err)
	}
	binary, _ := os.Executable()
	if _, err := portable.Bind(s, filepath.Join(home, ".local/share/my-friday/instances/fixture"), "fixture", binary, "device-tui"); err != nil {
		t.Fatal(err)
	}
	if err := managementMenu(home, os.Stdin, os.Stdout); err != nil {
		t.Fatal(err)
	}
	after, err := term.GetState(os.Stdin.Fd())
	// macOS sets PENDIN when returning to canonical input. It is transient
	// kernel input state, not an echo/raw-mode setting; compare all other bits.
	if after != nil {
		before.Lflag &^= unix.PENDIN
		after.Lflag &^= unix.PENDIN
	}
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("terminal mode not restored: before=%+v after=%+v error=%v", before, after, err)
	}
	fmt.Println("FRIDAY_TUI_TERMINAL_RESTORED")
}

func TestManagementPTYNavigationFormsSignalsAndResize(t *testing.T) {
	if _, err := os.Stat("/usr/bin/expect"); err != nil {
		t.Skip("expect unavailable; pure model tests still run")
	}
	binary, _ := os.Executable()
	for _, mode := range []string{"vim", "form", "interrupt", "terminate", "resize"} {
		t.Run(mode, func(t *testing.T) {
			home := t.TempDir()
			ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			defer cancel()
			script := filepath.Join("..", "..", "config", "testing", "management-tui.exp")
			cmd := exec.CommandContext(ctx, "/usr/bin/expect", script, binary, mode)
			cmd.Env = []string{"PATH=/usr/bin:/bin", "TERM=xterm-256color", "MY_FRIDAY_TEST_UI_HELPER=1", "MY_FRIDAY_TEST_UI_HOME=" + home}
			data, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s\n%v", data, err)
			}
			if _, err := os.Stat(filepath.Join(home, ".local/share/my-friday/repositories/hjkl-agent")); !os.IsNotExist(err) {
				t.Fatal("cancelled form created agent")
			}
		})
	}
}
