package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/acoz-labs/my-friday/internal/assistantinstance"
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
		if strings.Contains(message, "collision") || strings.Contains(message, "drift") || strings.Contains(message, "refused") || strings.Contains(message, "unhealthy") {
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

func verifyStatus(status codexhome.Status) error {
	if status.State != codexhome.StateHealthy {
		return fmt.Errorf("Codex baseline unhealthy: %s — %s", status.State, status.Detail)
	}
	return nil
}

func readConfirmation(input io.Reader, token string) (bool, error) {
	line, err := bufio.NewReader(input).ReadString('\n')
	if err != nil {
		return false, err
	}
	return line == token+"\n", nil
}

func run() error {
	if name := filepath.Base(os.Args[0]); assistantinstance.ValidateName(name) == nil {
		executable, err := os.Executable()
		if err != nil {
			return err
		}
		executable, err = filepath.EvalSymlinks(executable)
		if err != nil {
			return err
		}
		launcherDir := filepath.Dir(executable)
		if filepath.Base(launcherDir) == "bin" && filepath.Base(filepath.Dir(launcherDir)) == ".local" {
			home, err := realHome()
			if err != nil {
				return err
			}
			paths, err := assistantinstance.Derive(home, name)
			if err != nil {
				return err
			}
			if executable == paths.Launcher {
				return assistantinstance.Launch(home, name, os.Args[1:])
			}
		}
	}
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
	case "assistant":
		return runAssistant()
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

func realHome() (string, error) {
	account, err := user.Current()
	if err != nil {
		return "", err
	}
	if account.HomeDir == "" {
		return "", errors.New("current account home is unavailable")
	}
	return filepath.EvalSymlinks(account.HomeDir)
}

func runAssistant() error {
	if len(os.Args) < 4 {
		return fmt.Errorf("usage: my-friday assistant <create|migrate|verify|remove|recover> NAME")
	}
	home, err := realHome()
	if err != nil {
		return err
	}
	action, name := os.Args[2], os.Args[3]
	switch action {
	case "verify":
		if len(os.Args) != 4 {
			return fmt.Errorf("usage: my-friday assistant verify NAME")
		}
		m, err := assistantinstance.Verify(home, name)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Assistant %s healthy\nRoot: %s\nLauncher: %s\n", name, m.Root, m.Launcher)
		return nil
	case "create":
		if len(os.Args) != 8 || os.Args[4] != "--runtime" || os.Args[6] != "--memory" {
			return fmt.Errorf("usage: my-friday assistant create NAME --runtime PATH --memory PATH")
		}
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		exe, err = filepath.EvalSymlinks(exe)
		if err != nil {
			return err
		}
		codex, err := assistantinstance.FindCodex()
		if err != nil {
			return fmt.Errorf("Codex executable required on PATH: %w", err)
		}
		codex, err = filepath.EvalSymlinks(codex)
		if err != nil {
			return err
		}
		p, err := assistantinstance.PlanCreate(home, name, exe, codex)
		if err != nil {
			return err
		}
		assistantID, err := repository.ValidateRuntime(os.Args[5])
		if err != nil {
			return err
		}
		if err = repository.ValidatePair(os.Args[5], os.Args[7]); err != nil {
			return err
		}
		p, err = assistantinstance.WithRepositories(p, os.Args[5], os.Args[7], assistantID)
		if err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, p.String())
		fmt.Fprint(os.Stdout, "Type Create to continue: ")
		ok, err := readConfirmation(os.Stdin, "Create")
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(os.Stdout, "No changes made")
			return nil
		}
		if err = assistantinstance.Create(p, exe, codex); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Assistant %s created\n", name)
		return nil
	case "migrate":
		if len(os.Args) != 8 || os.Args[4] != "--runtime" || os.Args[6] != "--memory" {
			return fmt.Errorf("usage: my-friday assistant migrate NAME --runtime PATH --memory PATH")
		}
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		exe, err = filepath.EvalSymlinks(exe)
		if err != nil {
			return err
		}
		codex, err := assistantinstance.FindCodex()
		if err != nil {
			return err
		}
		codex, err = filepath.EvalSymlinks(codex)
		if err != nil {
			return err
		}
		p, err := assistantinstance.PlanCreate(home, name, exe, codex)
		if err != nil {
			return err
		}
		assistantID, err := repository.ValidateRuntime(os.Args[5])
		if err != nil {
			return err
		}
		if err = repository.ValidatePair(os.Args[5], os.Args[7]); err != nil {
			return err
		}
		p, err = assistantinstance.WithRepositories(p, os.Args[5], os.Args[7], assistantID)
		if err != nil {
			return err
		}
		oldHome, err := codexHomeWithin(home)
		if err != nil {
			return err
		}
		oldPlan, err := codexhome.Plan(codexhome.ActionUninstall, "", oldHome)
		if err != nil {
			return fmt.Errorf("prior projection proof refused: %w", err)
		}
		fmt.Fprint(os.Stdout, p.String())
		fmt.Fprintln(os.Stdout, "After the named instance verifies, remove only this proven prior projection:")
		fmt.Fprint(os.Stdout, oldPlan.String())
		fmt.Fprint(os.Stdout, "Type Migrate to continue: ")
		ok, err := readConfirmation(os.Stdin, "Migrate")
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(os.Stdout, "No changes made")
			return nil
		}
		if err = assistantinstance.Migrate(p, exe, codex, func() error { return codexhome.Execute(oldPlan) }); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Assistant %s migrated and prior projection removed\n", name)
		return nil
	case "remove":
		if len(os.Args) != 4 {
			return fmt.Errorf("usage: my-friday assistant remove NAME")
		}
		p, err := assistantinstance.PlanRemove(home, name)
		if err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, p.String())
		fmt.Fprint(os.Stdout, "Type Remove to continue: ")
		ok, err := readConfirmation(os.Stdin, "Remove")
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(os.Stdout, "No changes made")
			return nil
		}
		if err = assistantinstance.Remove(p); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Assistant %s removed\n", name)
		return nil
	case "recover":
		if len(os.Args) != 4 {
			return fmt.Errorf("usage: my-friday assistant recover NAME")
		}
		result, err := assistantinstance.Recover(home, name)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Assistant %s recovery: %s\n", name, result)
		return nil
	default:
		return fmt.Errorf("usage: my-friday assistant <create|migrate|verify|remove|recover> NAME")
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
	return codexHomeWithin(home)
}

func codexHomeWithin(home string) (string, error) {
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
		return verifyStatus(status)
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
	confirmed, err := readConfirmation(os.Stdin, token)
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Fprintln(os.Stdout, "No changes made")
		return nil
	}
	if err := codexhome.Execute(p); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Codex baseline %s complete\n", sub)
	return nil
}
