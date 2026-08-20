package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/acoz-labs/my-friday/internal/repository"
	"github.com/acoz-labs/my-friday/internal/terminal"
	"github.com/acoz-labs/my-friday/internal/transaction"
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
		result, e := transaction.RecoverWithResult(os.Args[3])
		if e == nil {
			fmt.Fprintln(os.Stdout, result)
		}
		return e
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
