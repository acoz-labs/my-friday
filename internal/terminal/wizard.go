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
	name, e := line("Assistant display name")
	if e != nil && e != io.EOF {
		return "", e
	}
	if name == "q" {
		return exit(output), nil
	}
	address, _ := line("How should the assistant address you? (optional)")
	if address == "q" {
		return exit(output), nil
	}
	purpose, _ := line("Assistant purpose")
	if purpose == "q" {
		return exit(output), nil
	}
	fmt.Fprintln(output, "Step 3 of 7: Communication style\n1 Balanced (default)\n2 Concise\n3 Conversational\n4 Custom")
	choice, _ := line("Choose 1-4")
	if choice == "q" {
		return exit(output), nil
	}
	preset := map[string]string{"": "balanced", "1": "balanced", "2": "concise", "3": "conversational", "4": "custom"}[choice]
	if preset == "" {
		return "", fmt.Errorf("communication style: choose 1, 2, 3, or 4")
	}
	guidance := ""
	if preset == "custom" {
		guidance, _ = line("Custom guidance")
	}
	p, err := profile.New(name, address, purpose, preset, guidance)
	if err != nil {
		return "", err
	}
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
		return "", fmt.Errorf("location mode: choose 1 or 2")
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
