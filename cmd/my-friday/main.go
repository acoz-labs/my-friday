package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/acoz-labs/my-friday/internal/repository"
	"github.com/acoz-labs/my-friday/internal/terminal"
	"github.com/acoz-labs/my-friday/internal/transaction"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: my-friday init | validate <runtime> <memory> | recover <journal>")
	}
	switch os.Args[1] {
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
		if len(os.Args) != 4 {
			return fmt.Errorf("usage: my-friday validate <runtime> <memory>")
		}
		if e := repository.ValidatePair(os.Args[2], os.Args[3]); e != nil {
			return e
		}
		fmt.Fprintln(os.Stdout, "Valid repository pair")
		return nil
	case "recover":
		if len(os.Args) != 3 {
			return fmt.Errorf("usage: my-friday recover <journal>")
		}
		return transaction.Recover(os.Args[2])
	default:
		return fmt.Errorf("unknown command %q", os.Args[1])
	}
}
