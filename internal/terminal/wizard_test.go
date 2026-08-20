package terminal

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	if strings.Contains(out.String(), "Promoted runtime") || !strings.Contains(out.String(), "Already complete") {
		t.Fatal(out.String())
	}
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
