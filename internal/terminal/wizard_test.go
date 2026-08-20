package terminal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/transaction"
)

func TestDefaultExitHasNoMutation(t *testing.T) {
	root := t.TempDir()
	in := strings.NewReader("\nFriday\n\nHelp me work\n\n\n" + root + "\n\n")
	var out bytes.Buffer
	result, err := Run(in, &out, root)
	if err != nil {
		t.Fatal(err)
	}
	if result != "Exit" {
		t.Fatal(result)
	}
	if !strings.Contains(out.String(), "No changes made") || strings.Contains(out.String(), "\x1b[") {
		t.Fatal(out.String())
	}
	if matches, err := filepath.Glob(filepath.Join(root, "my-friday-*")); err != nil || len(matches) != 0 {
		t.Fatalf("unexpected writes: %v err=%v", matches, err)
	}
}

func TestBackFromStylePreservesIdentityAnswers(t *testing.T) {
	root := t.TempDir()
	in := strings.NewReader("\nFriday\nBoss\nHelp\nb\n\n\n\n2\n\n" + root + "\nCreate\n")
	var out bytes.Buffer
	result, err := Run(in, &out, root)
	if err != nil || result != "Complete" {
		t.Fatalf("result=%s err=%v\n%s", result, err, out.String())
	}
	b, err := os.ReadFile(filepath.Join(root, "my-friday-runtime", "assistant", "profile.json"))
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Identity struct {
			DisplayName string `json:"display_name"`
			Purpose     string `json:"purpose"`
		} `json:"identity"`
	}
	if err = json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Identity.DisplayName != "Friday" || got.Identity.Purpose != "Help" {
		t.Fatalf("answers not preserved: %+v", got.Identity)
	}
}

func TestBackFromFirstIdentityReturnsToScope(t *testing.T) {
	root := t.TempDir()
	in := strings.NewReader("\nb\nq\n")
	var out bytes.Buffer
	result, err := Run(in, &out, root)
	if err != nil || result != "Exit" {
		t.Fatalf("result=%s err=%v", result, err)
	}
	if strings.Count(out.String(), "Step 1 of 7: Scope") != 2 {
		t.Fatal(out.String())
	}
}
func TestExplicitCreate(t *testing.T) {
	root := t.TempDir()
	in := strings.NewReader("\nFriday\nBoss\nHelp me work\n2\n\n" + root + "\nCreate\n")
	var out bytes.Buffer
	result, err := Run(in, &out, root)
	if err != nil {
		t.Fatal(err)
	}
	if result != "Complete" {
		t.Fatal(result)
	}
	if err := ValidatePair(filepath.Join(root, "my-friday-runtime"), filepath.Join(root, "my-friday-memory")); err != nil {
		t.Fatal(err)
	}
	for _, status := range []string{"Journaled", "Reserved", "Staged runtime", "Staged memory", "Validated", "Promoted runtime", "Promoted memory", "Verified", "Complete"} {
		if !strings.Contains(out.String(), status) {
			t.Fatalf("missing live status %q\n%s", status, out.String())
		}
	}
	for _, fact := range []string{"Normalized identity: Friday", "Normalized style: concise", "Runtime initial state: absent", "Memory initial state: absent"} {
		if !strings.Contains(out.String(), fact) {
			t.Fatalf("missing preview fact %q\n%s", fact, out.String())
		}
	}
}

func TestPreviewShowsModeNormalizationAndSymlinkMapping(t *testing.T) {
	root := t.TempDir()
	realParent := filepath.Join(root, "real")
	if err := os.Mkdir(realParent, 0700); err != nil {
		t.Fatal(err)
	}
	linkParent := filepath.Join(root, "linked")
	if err := os.Symlink(realParent, linkParent); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(realParent, "my-friday-runtime"), 0750); err != nil {
		t.Fatal(err)
	}
	in := strings.NewReader("\nFriday\n\nHelp\n\n\n" + linkParent + "\n\n")
	var out bytes.Buffer
	result, err := Run(in, &out, root)
	if err != nil || result != "Exit" {
		t.Fatalf("result=%s err=%v\n%s", result, err, out.String())
	}
	for _, fact := range []string{"Runtime initial state: empty directory mode 0750; will normalize to 0700", "Runtime symlink mapping:", linkParent, realParent} {
		if !strings.Contains(out.String(), fact) {
			t.Fatalf("missing preview fact %q\n%s", fact, out.String())
		}
	}
}

func TestAlreadyCompleteIsSeparateNoWriteResult(t *testing.T) {
	root := t.TempDir()
	answers := "\nFriday\nBoss\nHelp me work\n2\n\n" + root + "\nCreate\n"
	if result, err := Run(strings.NewReader(answers), &bytes.Buffer{}, root); err != nil || result != "Complete" {
		t.Fatalf("result=%s err=%v", result, err)
	}
	var out bytes.Buffer
	result, err := Run(strings.NewReader(answers), &out, root)
	if err != nil || result != "Already complete" {
		t.Fatalf("result=%s err=%v\n%s", result, err, out.String())
	}
	if strings.Contains(out.String(), "Promoted runtime") || !strings.Contains(out.String(), "Already complete") || !strings.Contains(out.String(), "exact completed repository mode 0700; no write needed") || strings.Contains(out.String(), "will normalize") {
		t.Fatal(out.String())
	}
}

func TestPreviewDistinguishesUnrelatedNonEmptyCollision(t *testing.T) {
	root := t.TempDir()
	runtime := filepath.Join(root, "runtime")
	memory := filepath.Join(root, "memory")
	if err := os.MkdirAll(runtime, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runtime, "foreign.txt"), []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	in := strings.NewReader("\nFriday\n\nHelp\n\n2\n" + runtime + "\n" + memory + "\n\n")
	var out bytes.Buffer
	_, _ = Run(in, &out, root)
	if !strings.Contains(out.String(), "Runtime initial state: unrelated non-empty collision mode 0700") || strings.Contains(out.String(), "Runtime initial state: empty directory") {
		t.Fatal(out.String())
	}
}

func TestPreMutationNavigationMatrixHasNoWrites(t *testing.T) {
	prefixes := map[string]string{
		"scope": "", "name": "\n", "address": "\nFriday\n", "purpose": "\nFriday\nBoss\n",
		"style": "\nFriday\nBoss\nHelp\n", "custom": "\nFriday\nBoss\nHelp\n4\n",
		"location-mode": "\nFriday\nBoss\nHelp\n2\n", "parent": "\nFriday\nBoss\nHelp\n2\n\n",
		"runtime": "\nFriday\nBoss\nHelp\n2\n2\n", "memory": "\nFriday\nBoss\nHelp\n2\n2\nRUNTIME\n",
		"confirmation": "\nFriday\nBoss\nHelp\n2\n\nPARENT\n",
	}
	backDestination := map[string]string{
		"name": "Step 1 of 7: Scope", "address": "Assistant display name", "purpose": "How should the assistant address you?",
		"style": "Step 2 of 7: Identity", "custom": "Step 3 of 7: Communication style",
		"location-mode": "Step 3 of 7: Communication style", "parent": "Step 4 of 7: Locations",
		"runtime": "Step 4 of 7: Locations", "memory": "Runtime target", "confirmation": "Step 4 of 7: Locations",
	}
	for prompt, prefix := range prefixes {
		for _, action := range []string{"q\n", "b\nq\n", ""} {
			name := prompt + "/" + map[string]string{"q\n": "q", "b\nq\n": "b", "": "eof"}[action]
			t.Run(name, func(t *testing.T) {
				root := t.TempDir()
				input := strings.ReplaceAll(strings.ReplaceAll(prefix, "RUNTIME", filepath.Join(root, "runtime")), "PARENT", root) + action
				before := snapshotTree(t, root)
				var out bytes.Buffer
				result, err := Run(strings.NewReader(input), &out, root)
				if err != nil || result != "Exit" || !strings.Contains(out.String(), "No changes made") {
					t.Fatalf("result=%q err=%v\n%s", result, err, out.String())
				}
				if after := snapshotTree(t, root); after != before {
					t.Fatalf("adjacent filesystem mutated\nbefore=%s\nafter=%s", before, after)
				}
				if action == "b\nq\n" && backDestination[prompt] != "" && !strings.Contains(out.String(), backDestination[prompt]) {
					t.Fatalf("back did not navigate to %q\n%s", backDestination[prompt], out.String())
				}
			})
		}
	}
}

func TestInterruptedRecoveryPromptExitMatrixDoesNotMutate(t *testing.T) {
	for name, action := range map[string]string{"q": "q\n", "b": "b\n", "eof": ""} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			parent := filepath.Join(root, "targets")
			if err := os.Mkdir(parent, 0700); err != nil {
				t.Fatal(err)
			}
			answers := "\nFriday\nBoss\nHelp\n2\n\n" + parent + "\n"
			original := executeWithProgress
			executeWithProgress = func(pl plan.CreationPlan, _ transaction.Fault, progress func(string)) (string, error) {
				return transaction.ExecuteWithProgress(pl, func(phase string) error {
					if phase == "verified" {
						return fmt.Errorf("injected interruption")
					}
					return nil
				}, progress)
			}
			_, _ = Run(strings.NewReader(answers+"Create\n"), &bytes.Buffer{}, root)
			executeWithProgress = original
			before := snapshotTree(t, root)
			var out bytes.Buffer
			result, err := Run(strings.NewReader(answers+action), &out, root)
			if err != nil || result != "Exit" || !strings.Contains(out.String(), "Interrupted creation found") {
				t.Fatalf("result=%q err=%v\n%s", result, err, out.String())
			}
			if after := snapshotTree(t, root); before != after {
				t.Fatalf("recovery exit mutated state\nbefore=%s\nafter=%s", before, after)
			}
		})
	}
}

func snapshotTree(t *testing.T, root string) string {
	t.Helper()
	var result strings.Builder
	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		fmt.Fprintf(&result, "%s:%s:%04o", rel, info.Mode().Type(), info.Mode().Perm())
		if info.Mode().IsRegular() {
			b, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			fmt.Fprintf(&result, ":%x", b)
		}
		result.WriteByte('\n')
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return result.String()
}

func TestBackFromConfirmationPreservesProfile(t *testing.T) {
	root := t.TempDir()
	in := strings.NewReader("\nFriday\nBoss\nHelp me work\n2\n\n" + root + "\nb\n\n" + root + "\nCreate\n")
	var out bytes.Buffer
	result, err := Run(in, &out, root)
	if err != nil {
		t.Fatal(err)
	}
	if result != "Complete" {
		t.Fatal(result)
	}
	if strings.Count(out.String(), "Step 4 of 7: Locations") != 2 {
		t.Fatal("confirmation back did not return to locations")
	}
}

func TestInvalidProfileAndChoicesRepromptWithoutMutation(t *testing.T) {
	root := t.TempDir()
	in := strings.NewReader("\n\nFriday\n\nHelp\n9\n2\n9\n\n" + root + "\n\n")
	var out bytes.Buffer
	result, err := Run(in, &out, root)
	if err != nil || result != "Exit" {
		t.Fatalf("result=%s err=%v", result, err)
	}
	if strings.Count(out.String(), "Invalid input:") < 3 {
		t.Fatal(out.String())
	}
	if matches, _ := filepath.Glob(filepath.Join(root, "my-friday-*")); len(matches) != 0 {
		t.Fatalf("unexpected writes: %v", matches)
	}
}
