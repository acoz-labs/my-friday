package terminal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/acoz-labs/my-friday/internal/environment"
	"github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/profile"
	"github.com/acoz-labs/my-friday/internal/repository"
	"github.com/acoz-labs/my-friday/internal/transaction"
)

func Run(input io.Reader, output io.Writer, invocationDir string) (string, error) {
	r := bufio.NewReader(input)
	line := func(prompt string) (string, error) {
		fmt.Fprint(output, prompt+" [b back; q quit]: ")
		s, e := r.ReadString('\n')
		if e == io.EOF && s == "" {
			return "q", nil
		}
		return strings.TrimSuffix(strings.TrimSuffix(s, "\n"), "\r"), e
	}
	fmt.Fprintln(output, "Step 1 of 7: Scope\nMy Friday creates two separate local Git repositories. It does not install Codex, use the network, access secrets, import content, create commits, or configure remotes.")
	v, e := line("Press Return to continue")
	if e != nil && e != io.EOF {
		return "", e
	}
	if v == "q" || v == "b" {
		return exit(output), nil
	}
	fmt.Fprintln(output, "Step 2 of 7: Identity")
	validatedText := func(prompt string, limit int, required bool) (string, bool, error) {
		for {
			value, readErr := line(prompt)
			if readErr != nil && readErr != io.EOF {
				return "", false, readErr
			}
			if value == "q" {
				return "", true, nil
			}
			if value == "b" {
				fmt.Fprintln(output, "Back is unavailable at this field; enter a value or q to quit.")
				continue
			}
			normalized, validationErr := profile.Normalize(value, limit, required)
			if validationErr != nil {
				fmt.Fprintf(output, "Invalid input: %v. Try again.\n", validationErr)
				continue
			}
			return normalized, false, nil
		}
	}
	name, quit, e := validatedText("Assistant display name", 60, true)
	if e != nil {
		return "", e
	}
	if quit {
		return exit(output), nil
	}
	address, quit, e := validatedText("How should the assistant address you? (optional)", 60, false)
	if e != nil {
		return "", e
	}
	if quit {
		return exit(output), nil
	}
	purpose, quit, e := validatedText("Assistant purpose", 240, true)
	if e != nil {
		return "", e
	}
	if quit {
		return exit(output), nil
	}
	fmt.Fprintln(output, "Step 3 of 7: Communication style\n1 Balanced (default)\n2 Concise\n3 Conversational\n4 Custom")
	preset := ""
	for preset == "" {
		choice, _ := line("Choose 1-4")
		if choice == "q" {
			return exit(output), nil
		}
		preset = map[string]string{"": "balanced", "1": "balanced", "2": "concise", "3": "conversational", "4": "custom"}[choice]
		if preset == "" {
			fmt.Fprintln(output, "Invalid input: choose 1, 2, 3, or 4. Try again.")
		}
	}
	guidance := ""
	if preset == "custom" {
		guidance, quit, e = validatedText("Custom guidance", 240, true)
		if e != nil {
			return "", e
		}
		if quit {
			return exit(output), nil
		}
	}
	p, err := profile.New(name, address, purpose, preset, guidance)
	if err != nil {
		return "", err
	}
	for {
		fmt.Fprintln(output, "Step 4 of 7: Locations\n1 One parent with stable names (default)\n2 Two separate targets")
		locationMode, _ := line("Choose 1-2")
		if locationMode == "q" {
			return exit(output), nil
		}
		var runtime, memory string
		if locationMode == "" || locationMode == "1" {
			parent, _ := line("Parent directory (default: invocation directory)")
			if parent == "q" {
				return exit(output), nil
			}
			if parent == "" {
				parent = invocationDir
			}
			parent, err = resolve(parent, invocationDir)
			if err != nil {
				return "", err
			}
			runtime, memory = filepath.Join(parent, "my-friday-runtime"), filepath.Join(parent, "my-friday-memory")
		} else if locationMode == "2" {
			runtime, _ = line("Runtime target")
			memory, _ = line("Memory target")
			if runtime == "q" || memory == "q" {
				return exit(output), nil
			}
			runtime, err = resolve(runtime, invocationDir)
			if err != nil {
				return "", err
			}
			memory, err = resolve(memory, invocationDir)
			if err != nil {
				return "", err
			}
		} else {
			fmt.Fprintln(output, "Invalid input: choose 1 or 2. Try again.")
			continue
		}
		if f, ok := input.(*os.File); ok {
			if err := environment.Check(existingAncestor(runtime), f); err != nil {
				return "", err
			}
			if err := environment.Check(existingAncestor(memory), f); err != nil {
				return "", err
			}
		}
		pl, err := plan.Build(p, runtime, memory)
		if err != nil {
			return "", err
		}
		fmt.Fprintln(output, "Step 5 of 7: Preview")
		fmt.Fprintf(output, "Plan: %s\nAssistant: %s\nRuntime: %s\nMemory: %s\n", pl.PlanID, pl.AssistantID, runtime, memory)
		for _, parent := range pl.MissingParents {
			fmt.Fprintln(output, "- create parent", parent, "mode 0700")
		}
		for _, support := range pl.SupportPaths {
			fmt.Fprintln(output, "- temporary support", support, "(removed after success)")
		}
		for _, f := range pl.Files {
			fmt.Fprintf(output, "- file %s:%s mode %04o sha256 %s\n", f.Role, f.Path, f.Mode, f.SHA256)
		}
		for _, a := range pl.Actions {
			fmt.Fprintln(output, "-", a)
		}
		for _, a := range pl.NegativeActions {
			fmt.Fprintln(output, "-", a)
		}
		fmt.Fprintln(output, "Step 6 of 7: Creation and verification")
		confirm, _ := line("Create these two repositories? [type Create; default Exit]")
		if confirm == "b" {
			continue
		}
		if confirm != "Create" {
			return exit(output), nil
		}
		fmt.Fprintln(output, "Preflight")
		result, err := transaction.Execute(pl, nil)
		if err != nil {
			return "", err
		}
		fmt.Fprintln(output, "Reserved\nStaged runtime\nStaged memory\nValidated\nPromoted runtime\nPromoted memory\nVerified")
		fmt.Fprintln(output, "Step 7 of 7: Result")
		fmt.Fprintln(output, result)
		fmt.Fprintln(output, runtime)
		fmt.Fprintln(output, memory)
		return result, nil
	}
}
func exit(w io.Writer) string { fmt.Fprintln(w, "No changes made"); return "Exit" }
func resolve(value, cwd string) (string, error) {
	if strings.Contains(value, "$") || strings.HasPrefix(value, "~") && value != "~" && !strings.HasPrefix(value, "~/") {
		return "", fmt.Errorf("path uses unsupported expansion")
	}
	if value == "~" || strings.HasPrefix(value, "~/") {
		h, e := os.UserHomeDir()
		if e != nil {
			return "", e
		}
		value = filepath.Join(h, strings.TrimPrefix(value, "~/"))
	}
	if !filepath.IsAbs(value) {
		value = filepath.Join(cwd, value)
	}
	return filepath.Clean(value), nil
}
func existingAncestor(path string) string {
	for {
		if _, err := os.Stat(path); err == nil {
			return path
		}
		next := filepath.Dir(path)
		if next == path {
			return path
		}
		path = next
	}
}
func ValidatePair(runtime, memory string) error { return repository.ValidatePair(runtime, memory) }
