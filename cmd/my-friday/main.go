package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/acoz-labs/my-friday/internal/codexhome"
	"github.com/acoz-labs/my-friday/internal/repository"
	"github.com/acoz-labs/my-friday/internal/terminal"
)

func main() {
	if err := run(); err != nil {
		code, stable := classifyError(os.Args, err)
		fmt.Fprintf(os.Stderr, "Error [%s]: %v\n", stable, err)
		os.Exit(code)
	}
}

func classifyError(args []string, err error) (int, string) {
	message := err.Error()
	command := "init"
	if len(args) > 1 {
		command = args[1]
	}
	if strings.HasPrefix(message, "usage:") || strings.Contains(message, "supports macOS") || strings.Contains(message, "required") && command == "init" {
		return 2, "input.invalid"
	}
	if command == "validate" {
		return 6, "contract.validation"
	}
	if command == "recover" || strings.Contains(message, "recovery required") {
		return 5, "transaction.recovery_required"
	}
	if command == "codex" {
		if strings.Contains(message, "collision") || strings.Contains(message, "drift") || strings.Contains(message, "refused") {
			return 3, "codex.state_denied"
		}
		return 2, "codex.input_invalid"
	}
	if strings.Contains(message, "rolled back") {
		return 4, "transaction.rolled_back"
	}
	if strings.Contains(message, "target") || strings.Contains(message, "reserved") || strings.Contains(message, "path") {
		return 3, "path.denied"
	}
	return 2, "input.invalid"
}
func run() error {
	command := "init"
	if len(os.Args) >= 2 {
		command = os.Args[1]
	}
	switch command {
	case "init":
		if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
			return fmt.Errorf("init supports macOS on Apple silicon only")
		}
		cwd, e := os.Getwd()
		if e != nil {
			return e
		}
		_, e = terminal.Run(os.Stdin, os.Stdout, cwd)
		return e
	case "validate":
		if len(os.Args) != 6 || os.Args[2] != "--runtime" || os.Args[4] != "--memory" {
			return fmt.Errorf("usage: my-friday validate --runtime PATH --memory PATH")
		}
		if e := repository.ValidatePair(os.Args[3], os.Args[5]); e != nil {
			return e
		}
		fmt.Fprintln(os.Stdout, "Valid repository pair")
		return nil
	case "recover":
		if len(os.Args) != 4 || os.Args[2] != "--transaction" {
			return fmt.Errorf("usage: my-friday recover --transaction PATH")
		}
		_, e := terminal.Recover(os.Args[3], os.Stdout)
		return e
	case "codex":
		return runCodex()
	case "version":
		if len(os.Args) != 2 {
			return fmt.Errorf("usage: my-friday version")
		}
		fmt.Fprintln(os.Stdout, "my-friday development (tool contract 1; repository contract 1)")
		return nil
	default:
		return fmt.Errorf("unknown command %q", os.Args[1])
	}
}

func codexHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	home, err = filepath.EvalSymlinks(home)
	if err != nil {
		return "", err
	}
	if value := os.Getenv("CODEX_HOME"); value != "" {
		absolute, err := filepath.Abs(value)
		if err != nil {
			return "", err
		}
		relative, err := filepath.Rel(home, absolute)
		if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("CODEX_HOME must be a descendant of the current user's real home")
		}
		return absolute, nil
	}
	return filepath.Join(home, ".codex"), nil
}

func runCodex() error {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		return fmt.Errorf("Codex baseline lifecycle supports macOS on Apple silicon only")
	}
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: my-friday codex <install|verify|repair|upgrade|rollback|uninstall|recover>")
	}
	home, err := codexHome()
	if err != nil {
		return err
	}
	sub := os.Args[2]
	if sub == "verify" {
		if len(os.Args) != 3 {
			return fmt.Errorf("usage: my-friday codex verify")
		}
		status, err := codexhome.Inspect("", home)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Codex baseline: %s — %s\n", status.State, status.Detail)
		return nil
	}
	if sub == "recover" {
		if len(os.Args) != 5 || os.Args[3] != "--transaction" {
			return fmt.Errorf("usage: my-friday codex recover --transaction PATH")
		}
		expected := filepath.Join(home, ".my-friday", "transaction.json")
		provided, _ := filepath.Abs(os.Args[4])
		if filepath.Clean(provided) != filepath.Clean(expected) {
			return fmt.Errorf("recovery refused: transaction path is not the effective Codex-home journal")
		}
		if err := codexhome.Recover(home); err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, "Codex baseline recovery complete")
		return nil
	}
	action := codexhome.Action(sub)
	runtimePath := ""
	if sub == "install" || sub == "upgrade" {
		if len(os.Args) != 5 || os.Args[3] != "--runtime" {
			return fmt.Errorf("usage: my-friday codex %s --runtime PATH", sub)
		}
		runtimePath = os.Args[4]
	} else if len(os.Args) != 3 {
		return fmt.Errorf("usage: my-friday codex %s", sub)
	}
	p, err := codexhome.Plan(action, runtimePath, home)
	if err != nil {
		return err
	}
	fmt.Fprint(os.Stdout, p.String())
	token := strings.ToUpper(sub[:1]) + sub[1:]
	fmt.Fprintf(os.Stdout, "Type %s to continue: ", token)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && strings.TrimSpace(line) == "" {
		return err
	}
	if strings.TrimSpace(line) != token {
		fmt.Fprintln(os.Stdout, "No changes made")
		return nil
	}
	if err := codexhome.Execute(p); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Codex baseline %s complete\n", sub)
	return nil
}
