package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
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
	if os.Getenv("MY_FRIDAY_TEST_UI_REFERENCES") == "1" {
		if err := s.AddReference(portable.ReferenceLibrary{Version: 1, ID: "prior-work", Title: "Prior work", Description: "Earlier experiences", Purpose: "Reference only"}); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(home, "external"), 0700); err != nil {
			t.Fatal(err)
		}
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
	for _, mode := range []string{"vim", "no-color", "form", "reports", "references", "interrupt", "terminate", "resize"} {
		t.Run(mode, func(t *testing.T) {
			home := t.TempDir()
			ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			defer cancel()
			script := filepath.Join("..", "..", "config", "testing", "management-tui.exp")
			cmd := exec.CommandContext(ctx, "/usr/bin/expect", script, binary, mode)
			cmd.Env = []string{"PATH=/usr/bin:/bin", "TERM=xterm-256color", "MY_FRIDAY_TEST_UI_HELPER=1", "MY_FRIDAY_TEST_UI_HOME=" + home}
			transcript := filepath.Join(home, "terminal.txt")
			cmd.Env = append(cmd.Env, "MY_FRIDAY_TEST_UI_TRANSCRIPT="+transcript)
			if mode == "references" {
				cmd.Env = append(cmd.Env, "MY_FRIDAY_TEST_UI_REFERENCES=1")
			}
			if mode == "no-color" {
				cmd.Env = append(cmd.Env, "NO_COLOR=1")
			}
			data, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s\n%v", data, err)
			}
			raw, err := os.ReadFile(transcript)
			if err != nil {
				t.Fatal(err)
			}
			styled := regexp.MustCompile(`\x1b\[(1;36|32|33|31|2|1)m`).Match(raw)
			if mode == "no-color" && styled {
				t.Fatal("NO_COLOR emitted styling")
			}
			if mode == "vim" && !styled {
				t.Fatal("interactive terminal emitted no styling")
			}
			if (mode == "vim" || mode == "no-color") && !strings.Contains(string(raw), "pinned:") {
				t.Fatal("missing agent context")
			}
			if _, err := os.Stat(filepath.Join(home, ".local/share/my-friday/repositories/hjkl-agent")); !os.IsNotExist(err) {
				t.Fatal("cancelled form created agent")
			}
		})
	}
}
