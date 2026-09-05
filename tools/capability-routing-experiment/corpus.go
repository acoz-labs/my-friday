package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var idPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
var sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func ReadJSON(path string, dst any) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return decodeStrictJSON(filepath.Base(path), body, dst)
}

func decodeStrictJSON(name string, body []byte, dst any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("decode %s: %w", name, err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("decode %s: multiple JSON values", name)
		}
		return fmt.Errorf("decode %s: %w", name, err)
	}
	return nil
}

func LoadBundle(dir string) (Bundle, error) {
	var bundle Bundle
	for _, item := range []struct {
		name string
		dst  any
	}{
		{"capabilities.json", &bundle.Capabilities},
		{"tasks.json", &bundle.Tasks},
		{"labels.json", &bundle.Labels},
	} {
		if err := ReadJSON(filepath.Join(dir, item.name), item.dst); err != nil {
			return Bundle{}, err
		}
	}
	manifestPath := filepath.Join(dir, "manifest.json")
	if _, err := os.Stat(manifestPath); err == nil {
		if err = ReadJSON(manifestPath, &bundle.Manifest); err != nil {
			return Bundle{}, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return Bundle{}, err
	}
	return bundle, nil
}

func ValidateBundle(bundle Bundle) error {
	if bundle.Capabilities.Version != SchemaVersion || bundle.Tasks.Version != SchemaVersion || bundle.Labels.Version != SchemaVersion {
		return errors.New("all corpus files must use schema version 1")
	}
	if bundle.Capabilities.Revision == "" || bundle.Tasks.Revision != bundle.Capabilities.Revision || bundle.Labels.Revision != bundle.Capabilities.Revision {
		return errors.New("corpus, tasks, and labels must share one non-empty revision")
	}
	if len(bundle.Capabilities.Capabilities) != 24 || len(bundle.Tasks.Tasks) != 24 || len(bundle.Labels.Labels) != 24 {
		return fmt.Errorf("expected 24 capabilities, 24 tasks, and 24 labels; got %d, %d, and %d", len(bundle.Capabilities.Capabilities), len(bundle.Tasks.Tasks), len(bundle.Labels.Labels))
	}
	caps := map[string]Capability{}
	for _, capability := range bundle.Capabilities.Capabilities {
		if !idPattern.MatchString(capability.ID) || capability.Name == "" || capability.Summary == "" || capability.Body == "" || capability.Revision == "" {
			return fmt.Errorf("invalid capability %q", capability.ID)
		}
		if _, exists := caps[capability.ID]; exists {
			return fmt.Errorf("duplicate capability %q", capability.ID)
		}
		caps[capability.ID] = capability
	}
	if err := validateDependencies(caps); err != nil {
		return err
	}
	tasks := map[string]Task{}
	counts := map[string]map[string]int{"development": {}, "held-out": {}}
	for _, task := range bundle.Tasks.Tasks {
		if !idPattern.MatchString(task.ID) || task.Prompt == "" {
			return fmt.Errorf("invalid task %q", task.ID)
		}
		if _, exists := tasks[task.ID]; exists {
			return fmt.Errorf("duplicate task %q", task.ID)
		}
		if _, ok := counts[task.Split]; !ok {
			return fmt.Errorf("task %s has invalid split %q", task.ID, task.Split)
		}
		if !contains(TaskCategories, task.Category) {
			return fmt.Errorf("task %s has invalid category %q", task.ID, task.Category)
		}
		for _, path := range append(append([]string{}, task.ReadPaths...), task.WritePaths...) {
			if !safeFixturePath(path) {
				return fmt.Errorf("task %s has unsafe fixture path %q", task.ID, path)
			}
		}
		fixturePaths := map[string]bool{}
		for _, fixture := range task.Fixtures {
			if !safeFixturePath(fixture.Path) || fixture.Content == "" || fixturePaths[fixture.Path] || !contains(task.ReadPaths, fixture.Path) {
				return fmt.Errorf("task %s has invalid fixture %q", task.ID, fixture.Path)
			}
			fixturePaths[fixture.Path] = true
		}
		for _, path := range task.ReadPaths {
			if !fixturePaths[path] {
				return fmt.Errorf("task %s lacks content for read fixture %q", task.ID, path)
			}
		}
		tasks[task.ID] = task
		counts[task.Split][task.Category]++
	}
	for split, categories := range counts {
		for _, category := range TaskCategories {
			if categories[category] != 1 {
				return fmt.Errorf("split %s must contain exactly one %s task", split, category)
			}
		}
	}
	seenLabels := map[string]bool{}
	for _, label := range bundle.Labels.Labels {
		if _, ok := tasks[label.TaskID]; !ok || seenLabels[label.TaskID] {
			return fmt.Errorf("invalid or duplicate label for %q", label.TaskID)
		}
		if !contains([]string{"execute", "clarify", "refuse"}, label.Expectation) {
			return fmt.Errorf("label %s has invalid expectation %q", label.TaskID, label.Expectation)
		}
		for _, set := range label.AllowedCapabilitySets {
			for _, id := range set {
				if _, ok := caps[id]; !ok {
					return fmt.Errorf("label %s names unknown capability %q", label.TaskID, id)
				}
			}
		}
		for _, effect := range label.RequiredEffects {
			parts := strings.SplitN(effect, ":", 2)
			if len(parts) != 2 || parts[0] != "write" || !safeFixturePath(parts[1]) || !contains(tasks[label.TaskID].WritePaths, parts[1]) {
				return fmt.Errorf("label %s has invalid required effect %q", label.TaskID, effect)
			}
		}
		if len(label.RequiredSummary.Changes) == 0 || len(label.RequiredSummary.Failures) == 0 || len(label.RequiredSummary.Verification) == 0 || len(label.RequiredSummary.Limitations) == 0 {
			return fmt.Errorf("label %s lacks frozen material summary requirements", label.TaskID)
		}
		seenLabels[label.TaskID] = true
	}
	return nil
}

func validateDependencies(caps map[string]Capability) error {
	state := map[string]int{}
	var visit func(string) error
	visit = func(id string) error {
		if state[id] == 1 {
			return fmt.Errorf("capability dependency cycle at %s", id)
		}
		if state[id] == 2 {
			return nil
		}
		state[id] = 1
		for _, dependency := range caps[id].Dependencies {
			found, ok := caps[dependency.ID]
			if !ok {
				return fmt.Errorf("capability %s has missing dependency %s", id, dependency.ID)
			}
			if found.Revision != dependency.Revision {
				return fmt.Errorf("capability %s dependency %s revision mismatch", id, dependency.ID)
			}
			if err := visit(dependency.ID); err != nil {
				return err
			}
		}
		state[id] = 2
		return nil
	}
	ids := make([]string, 0, len(caps))
	for id := range caps {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}

func safeFixturePath(path string) bool {
	if path == "" || filepath.IsAbs(path) || filepath.Clean(path) != path || strings.HasPrefix(path, "..") {
		return false
	}
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}

func PrepareManifest(bundle Bundle, sourceCommit string, harnesses []HarnessSpec) (Manifest, error) {
	if err := ValidateBundle(bundle); err != nil {
		return Manifest{}, err
	}
	if !commitPattern.MatchString(sourceCommit) || len(harnesses) != 2 {
		return Manifest{}, errors.New("a 40-character lowercase source commit and exactly two harnesses are required")
	}
	seenHarnesses := map[string]bool{}
	for _, harness := range harnesses {
		if harness.ID == "" || harness.ExecutableVersion == "" || harness.Model == "" || harness.Config == "" {
			return Manifest{}, errors.New("each harness requires id, executable version, model, and config")
		}
		if seenHarnesses[harness.ID] {
			return Manifest{}, fmt.Errorf("duplicate harness %q", harness.ID)
		}
		seenHarnesses[harness.ID] = true
	}
	if !seenHarnesses["codex"] || !seenHarnesses["claude"] {
		return Manifest{}, errors.New("manifest harnesses must be codex and claude")
	}
	manifest := Manifest{
		Version: SchemaVersion, SourceCommit: sourceCommit, CorpusRevision: bundle.Capabilities.Revision,
		Hashes:    ManifestHashes{Capabilities: digestJSON(bundle.Capabilities), Tasks: digestJSON(bundle.Tasks), Labels: digestJSON(bundle.Labels)},
		Budgets:   Budgets{120, 30000, 8, 1, 1, 2000, 12, 1800, 360000},
		Harnesses: harnesses, Modes: append([]string{}, RoutingModes...), Repetitions: 2,
	}
	sequence := 0
	for _, harness := range harnesses {
		for ordinal, task := range bundle.Tasks.Tasks {
			for repetition := 1; repetition <= 2; repetition++ {
				modes := rotatedModes(ordinal, repetition)
				for _, mode := range modes {
					sequence++
					manifest.Cells = append(manifest.Cells, ManifestCell{
						TrialID:  fmt.Sprintf("%s-%s-%s-r%d", harness.ID, task.ID, mode, repetition),
						Sequence: sequence, HarnessID: harness.ID, TaskID: task.ID, Mode: mode, Repetition: repetition,
						Cache: map[bool]string{true: "warm", false: "cold"}[repetition == 2],
					})
				}
			}
		}
	}
	return manifest, nil
}

func rotatedModes(ordinal, repetition int) []string {
	modes := append([]string{}, RoutingModes...)
	offset := ordinal % len(modes)
	modes = append(modes[offset:], modes[:offset]...)
	if repetition == 2 {
		for left, right := 0, len(modes)-1; left < right; left, right = left+1, right-1 {
			modes[left], modes[right] = modes[right], modes[left]
		}
	}
	return modes
}

func digestJSON(value any) string {
	body, _ := json.Marshal(value)
	digest := sha256.Sum256(body)
	return hex.EncodeToString(digest[:])
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
