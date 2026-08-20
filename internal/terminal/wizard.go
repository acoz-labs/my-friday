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
	values := []string{"", "", ""}
	preset, guidance := "", ""
scopeStep:
	fmt.Fprintln(output, "Step 1 of 7: Scope\nMy Friday creates two separate local Git repositories. It does not install Codex, use the network, access secrets, import content, create commits, or configure remotes.")
	v, e := line("Press Return to continue")
	if e != nil && e != io.EOF {
		return "", e
	}
	if v == "q" || v == "b" {
		return exit(output), nil
	}
identityStep:
	fmt.Fprintln(output, "Step 2 of 7: Identity")
	validatedText := func(prompt, current string, limit int, required bool) (string, string, error) {
		for {
			displayPrompt := prompt
			if current != "" {
				displayPrompt += " (Return keeps current value)"
			}
			value, readErr := line(displayPrompt)
			if readErr != nil && readErr != io.EOF {
				return "", "", readErr
			}
			if value == "q" {
				return "", "quit", nil
			}
			if value == "b" {
				return "", "back", nil
			}
			if value == "" && current != "" {
				value = current
			}
			normalized, validationErr := profile.Normalize(value, limit, required)
			if validationErr != nil {
				fmt.Fprintf(output, "Invalid input: %v. Try again.\n", validationErr)
				continue
			}
			return normalized, "", nil
		}
	}
	prompts := []string{"Assistant display name", "How should the assistant address you? (optional)", "Assistant purpose"}
	limits := []int{60, 60, 240}
	required := []bool{true, false, true}
	for field := 0; field < len(values); {
		value, action, readErr := validatedText(prompts[field], values[field], limits[field], required[field])
		if readErr != nil {
			return "", readErr
		}
		if action == "quit" {
			return exit(output), nil
		}
		if action == "back" {
			if field == 0 {
				goto scopeStep
			}
			field--
			continue
		}
		values[field] = value
		field++
	}
	name, address, purpose := values[0], values[1], values[2]

styleStep:
	fmt.Fprintln(output, "Step 3 of 7: Communication style\n1 Balanced (default)\n2 Concise\n3 Conversational\n4 Custom")
	for {
		choice, _ := line("Choose 1-4")
		if choice == "q" {
			return exit(output), nil
		}
		if choice == "b" {
			goto identityStep
		}
		if choice == "" && preset != "" {
			break
		}
		selected := map[string]string{"": "balanced", "1": "balanced", "2": "concise", "3": "conversational", "4": "custom"}[choice]
		if selected == "" {
			fmt.Fprintln(output, "Invalid input: choose 1, 2, 3, or 4. Try again.")
			continue
		}
		preset = selected
		if preset != "custom" {
			guidance = ""
		}
		break
	}
	if preset == "custom" {
		var action string
		guidance, action, e = validatedText("Custom guidance", guidance, 240, true)
		if e != nil {
			return "", e
		}
		if action == "quit" {
			return exit(output), nil
		}
		if action == "back" {
			goto styleStep
		}
	}
	p, err := profile.New(name, address, purpose, preset, guidance)
	if err != nil {
		return "", err
	}
	locationMode, parentValue, runtimeValue, memoryValue := "", "", "", ""
	for {
		fmt.Fprintln(output, "Step 4 of 7: Locations\n1 One parent with stable names (default)\n2 Two separate targets")
		choice, _ := line("Choose 1-2")
		if choice == "" && locationMode != "" {
			choice = locationMode
		}
		locationMode = choice
		if locationMode == "q" {
			return exit(output), nil
		}
		if locationMode == "b" {
			goto styleStep
		}
		var runtime, memory string
		if locationMode == "" || locationMode == "1" {
			parent, _ := line("Parent directory (default: invocation directory)")
			if parent == "q" {
				return exit(output), nil
			}
			if parent == "b" {
				locationMode = ""
				continue
			}
			if parent == "" && parentValue != "" {
				parent = parentValue
			}
			if parent == "" {
				parent = invocationDir
			}
			parentValue = parent
			parent, err = resolve(parent, invocationDir)
			if err != nil {
				fmt.Fprintf(output, "Invalid input: %v. Try again.\n", err)
				continue
			}
			runtime, memory = filepath.Join(parent, "my-friday-runtime"), filepath.Join(parent, "my-friday-memory")
		} else if locationMode == "2" {
			for {
				runtime, _ = line("Runtime target")
				if runtime == "q" {
					return exit(output), nil
				}
				if runtime == "b" {
					locationMode = ""
					break
				}
				if runtime == "" && runtimeValue != "" {
					runtime = runtimeValue
				}
				runtimeValue = runtime
				runtime, err = resolve(runtime, invocationDir)
				if err != nil {
					fmt.Fprintf(output, "Invalid input: %v. Try again.\n", err)
					continue
				}
				memory, _ = line("Memory target")
				if memory == "q" {
					return exit(output), nil
				}
				if memory == "b" {
					continue
				}
				if memory == "" && memoryValue != "" {
					memory = memoryValue
				}
				memoryValue = memory
				memory, err = resolve(memory, invocationDir)
				if err != nil {
					fmt.Fprintf(output, "Invalid input: %v. Try again.\n", err)
					continue
				}
				break
			}
			if locationMode == "" {
				continue
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
			fmt.Fprintf(output, "Invalid input: %v. Try again.\n", err)
			continue
		}
		fmt.Fprintln(output, "Step 5 of 7: Preview")
		fmt.Fprintf(output, "Plan: %s\nAssistant: %s\nRuntime entered: %s\nRuntime canonical: %s\nMemory entered: %s\nMemory canonical: %s\n", pl.PlanID, pl.AssistantID, runtimeValueOr(runtimeValue, parentValue, runtime), pl.Targets.Runtime, runtimeValueOr(memoryValue, parentValue, memory), pl.Targets.Memory)
		fmt.Fprintf(output, "Normalized identity: %s\nNormalized address: %s\nNormalized purpose: %s\nNormalized style: %s\n", p.Identity.DisplayName, optionalPreview(p.Identity.AddressUserAs), p.Identity.Purpose, stylePreview(p))
		printTargetPreview(output, "Runtime", runtime, pl.Targets.Runtime)
		printTargetPreview(output, "Memory", memory, pl.Targets.Memory)
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
		if journalPath, phase, interrupted := transaction.Interrupted(pl); interrupted {
			fmt.Fprintf(output, "Interrupted creation found at phase %s.\nRecovery command: my-friday recover --transaction %s\n", phase, journalPath)
			recoverNow, _ := line("Type r to recover now; default Exit")
			if recoverNow != "r" {
				return exit(output), nil
			}
			recoveryResult, err := transaction.RecoverWithResult(journalPath)
			if err != nil {
				return "", err
			}
			fmt.Fprintln(output, "Step 7 of 7: Result\n"+recoveryResult)
			return recoveryResult, nil
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
		result, err := transaction.ExecuteWithProgress(pl, nil, func(status string) { fmt.Fprintln(output, status) })
		if err != nil {
			fmt.Fprintln(output, "Step 7 of 7: Recovery required")
			fmt.Fprintf(output, "Plan: %s\nError: %v\nCompleted phase: see status lines above\n", pl.PlanID, err)
			if journalPath, phase, interrupted := transaction.Interrupted(pl); interrupted {
				fmt.Fprintf(output, "Retained phase: %s\nRecovery command: my-friday recover --transaction %s\n", phase, journalPath)
			} else {
				fmt.Fprintln(output, "Rollback restored the pre-run state; correct the reported error and run init again.")
			}
			return "", err
		}
		fmt.Fprintln(output, "Step 7 of 7: Result")
		fmt.Fprintln(output, result)
		runtimeMode := actualMode(pl.Targets.Runtime)
		memoryMode := actualMode(pl.Targets.Memory)
		fmt.Fprintf(output, "Runtime: %s mode %04o\nMemory: %s mode %04o\nAssistant: %s\nContracts: repository v1, tool v1\nGit: unborn main; no commits; no remotes\nNext: inspect the repositories, then use my-friday validate when needed\n", pl.Targets.Runtime, runtimeMode, pl.Targets.Memory, memoryMode, pl.AssistantID)
		for _, parent := range pl.MissingParents {
			fmt.Fprintf(output, "Created parent: %s mode %04o\n", parent, actualMode(parent))
		}
		return result, nil
	}
}
func actualMode(path string) os.FileMode {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Mode().Perm()
}
func runtimeValueOr(value, parent, fallback string) string {
	if value != "" {
		return value
	}
	if parent != "" {
		return filepath.Join(parent, filepath.Base(fallback))
	}
	return fallback
}
func optionalPreview(value *string) string {
	if value == nil {
		return "(none)"
	}
	return *value
}
func stylePreview(p profile.Profile) string {
	if p.Communication.CustomGuidance == nil {
		return p.Communication.Preset
	}
	return p.Communication.Preset + ": " + *p.Communication.CustomGuidance
}
func printTargetPreview(w io.Writer, role, entered, canonical string) {
	info, err := os.Lstat(canonical)
	switch {
	case os.IsNotExist(err):
		fmt.Fprintf(w, "%s initial state: absent\n", role)
	case err != nil:
		fmt.Fprintf(w, "%s initial state: unavailable (%v)\n", role, err)
	case info.IsDir() && info.Mode()&os.ModeSymlink == 0:
		fmt.Fprintf(w, "%s initial state: empty directory mode %04o; will normalize to 0700\n", role, info.Mode().Perm())
	default:
		fmt.Fprintf(w, "%s initial state: collision (%s)\n", role, info.Mode().Type())
	}
	if filepath.Clean(entered) != canonical {
		fmt.Fprintf(w, "%s symlink mapping: %s -> %s\n", role, filepath.Clean(entered), canonical)
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
