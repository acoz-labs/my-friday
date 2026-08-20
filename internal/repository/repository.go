package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/profile"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

var observeGitCommand = func([]string, []string) {}

func gitCommand(args ...string) *exec.Cmd {
	env := []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "LANG=C.UTF-8"}
	observeGitCommand(append([]string(nil), args...), append([]string(nil), env...))
	cmd := exec.Command("git", args...)
	cmd.Env = env
	return cmd
}

func Create(pl plan.CreationPlan, runtime, memory string) error {
	return CreateWithCheckpoint(pl, runtime, memory, nil)
}

func CreateWithCheckpoint(pl plan.CreationPlan, runtime, memory string, checkpoint func(string) error) error {
	for _, target := range []struct{ role, path string }{{"runtime", runtime}, {"memory", memory}} {
		if err := os.MkdirAll(target.path, 0700); err != nil {
			return err
		}
		if err := os.Chmod(target.path, 0700); err != nil {
			return err
		}
		for _, f := range pl.Files {
			if f.Role != target.role {
				continue
			}
			dst := filepath.Join(target.path, filepath.FromSlash(f.Path))
			if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
				return err
			}
			if err := os.WriteFile(dst, f.Bytes, 0600); err != nil {
				return err
			}
		}
		if checkpoint != nil {
			if err := checkpoint(target.role + "-files"); err != nil {
				return err
			}
		}
		tmpl, err := os.MkdirTemp(filepath.Dir(target.path), ".my-friday-empty-template-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpl)
		cmd := gitCommand("init", "--quiet", "--initial-branch=main", "--template="+tmpl, target.path)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git init %s: %w: %s", target.role, err, strings.TrimSpace(string(out)))
		}
		for _, setting := range [][2]string{
			{"core.repositoryformatversion", "0"},
			{"core.filemode", "true"},
			{"core.bare", "false"},
			{"core.logallrefupdates", "true"},
			{"core.ignorecase", "false"},
			{"core.precomposeunicode", "false"},
		} {
			key, value := setting[0], setting[1]
			config := gitCommand("-C", target.path, "config", "--local", key, value)
			if out, err := config.CombinedOutput(); err != nil {
				return fmt.Errorf("git config %s %s: %w: %s", target.role, key, err, strings.TrimSpace(string(out)))
			}
		}
		if err := filepath.WalkDir(filepath.Join(target.path, ".git"), func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			mode := os.FileMode(0600)
			if entry.IsDir() {
				mode = 0700
			}
			return os.Chmod(path, mode)
		}); err != nil {
			return fmt.Errorf("normalize Git metadata modes for %s: %w", target.role, err)
		}
		if checkpoint != nil {
			if err := checkpoint(target.role + "-git"); err != nil {
				return err
			}
		}
	}
	return nil
}

func ValidatePair(runtime, memory string) error {
	_, _, err := validatePair(runtime, memory, false, false)
	return err
}

func ValidateFreshPair(runtime, memory string) error {
	_, _, err := validatePair(runtime, memory, true, true)
	return err
}

func validatePair(runtime, memory string, fresh, allowMarker bool) (string, string, error) {
	rid, err := validate(runtime, "runtime", allowMarker)
	if err != nil {
		return "", "", err
	}
	mid, err := validate(memory, "memory", allowMarker)
	if err != nil {
		return "", "", err
	}
	if rid != mid {
		return "", "", fmt.Errorf("assistant identifiers do not match")
	}
	if fresh {
		if err = validateFreshGit(runtime); err != nil {
			return "", "", err
		}
		if err = validateFreshGit(memory); err != nil {
			return "", "", err
		}
	}
	return rid, mid, nil
}
func validate(root, role string, allowMarker bool) (string, error) {
	mb, err := os.ReadFile(filepath.Join(root, ".my-friday/manifest.json"))
	if err != nil {
		return "", err
	}
	sb, err := os.ReadFile(filepath.Join(root, ".my-friday/schemas/repository-manifest.v1.schema.json"))
	if err != nil {
		return "", err
	}
	if !bytes.Equal(sb, []byte(plan.ManifestSchema())) {
		return "", fmt.Errorf("repository manifest schema differs from the embedded v1 contract")
	}
	if err = validateJSON([]byte(plan.ManifestSchema()), mb); err != nil {
		return "", fmt.Errorf("manifest schema: %w", err)
	}
	var m struct {
		ContractVersion int    `json:"contract_version"`
		RepositoryRole  string `json:"repository_role"`
		AssistantID     string `json:"assistant_id"`
	}
	if err = json.Unmarshal(mb, &m); err != nil {
		return "", err
	}
	if m.ContractVersion != 1 || m.RepositoryRole != role {
		return "", fmt.Errorf("invalid %s repository role or version", role)
	}
	if role == "runtime" {
		pb, e := os.ReadFile(filepath.Join(root, "assistant/profile.json"))
		if e != nil {
			return "", e
		}
		ps, e := os.ReadFile(filepath.Join(root, ".my-friday/schemas/assistant-profile.v1.schema.json"))
		if e != nil {
			return "", e
		}
		if !bytes.Equal(ps, []byte(plan.ProfileSchema())) {
			return "", fmt.Errorf("assistant profile schema differs from the embedded v1 contract")
		}
		if e = validateJSON([]byte(plan.ProfileSchema()), pb); e != nil {
			return "", fmt.Errorf("profile schema: %w", e)
		}
		var p profile.Profile
		if e = json.Unmarshal(pb, &p); e != nil || p.AssistantID != m.AssistantID {
			return "", fmt.Errorf("profile assistant identifier mismatch")
		}
		if e = profile.Validate(p); e != nil {
			return "", fmt.Errorf("profile semantics: %w", e)
		}
	}
	allowed := map[string]bool{"manifest.json": true, "schemas": true, "schemas/repository-manifest.v1.schema.json": true}
	if allowMarker {
		allowed["creation-state.json"] = true
	}
	if role == "runtime" {
		allowed["schemas/assistant-profile.v1.schema.json"] = true
	}
	if err = filepath.WalkDir(filepath.Join(root, ".my-friday"), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, _ := filepath.Rel(filepath.Join(root, ".my-friday"), path)
		if rel != "." && !allowed[filepath.ToSlash(rel)] {
			return fmt.Errorf("unknown owned contract path: .my-friday/%s", filepath.ToSlash(rel))
		}
		return nil
	}); err != nil {
		return "", err
	}
	if err = validateGitRepository(root); err != nil {
		return "", err
	}
	return m.AssistantID, nil
}

func validateGitRepository(root string) error {
	cmd := gitCommand("-C", root, "rev-parse", "--git-dir")
	if out, err := cmd.Output(); err != nil || strings.TrimSpace(string(out)) != ".git" {
		return fmt.Errorf("repository is not a local Git repository")
	}
	return nil
}

func validateFreshGit(root string) error {
	cmd := gitCommand("-C", root, "symbolic-ref", "--short", "HEAD")
	out, e := cmd.Output()
	if e != nil || strings.TrimSpace(string(out)) != "main" {
		return fmt.Errorf("repository is not an unborn main branch")
	}
	if gitCommand("-C", root, "rev-parse", "--verify", "HEAD").Run() == nil {
		return fmt.Errorf("repository must have no commits")
	}
	if out, e = gitCommand("-C", root, "remote").Output(); e != nil || len(bytes.TrimSpace(out)) != 0 {
		return fmt.Errorf("repository must have no remotes")
	}
	return nil
}
func validateJSON(schema, doc []byte) error {
	var schemaValue any
	if err := json.Unmarshal(schema, &schemaValue); err != nil {
		return err
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource("schema.json", schemaValue); err != nil {
		return err
	}
	s, err := c.Compile("schema.json")
	if err != nil {
		return err
	}
	var v any
	if err = json.Unmarshal(doc, &v); err != nil {
		return err
	}
	return s.Validate(v)
}

func ExactBaseline(pl plan.CreationPlan, runtime, memory string) bool {
	return exactBaseline(pl, runtime, memory, false)
}

// ExactTransactionBaseline proves that both repositories contain only the
// planned baseline plus the transaction ownership marker used before cleanup.
func ExactTransactionBaseline(pl plan.CreationPlan, runtime, memory string) bool {
	return exactBaseline(pl, runtime, memory, true)
}

func exactBaseline(pl plan.CreationPlan, runtime, memory string, allowMarker bool) bool {
	if ValidateFreshPair(runtime, memory) != nil {
		return false
	}
	for _, f := range pl.Files {
		root := runtime
		if f.Role == "memory" {
			root = memory
		}
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f.Path)))
		if err != nil || !bytes.Equal(b, f.Bytes) {
			return false
		}
	}
	return noUnexpected(runtime, pl, "runtime", allowMarker) && noUnexpected(memory, pl, "memory", allowMarker) && exactFreshGitMetadata(runtime) && exactFreshGitMetadata(memory)
}

func exactFreshGitMetadata(root string) bool {
	expectedValues := map[string]string{
		"core.repositoryformatversion": "0",
		"core.filemode":                "true",
		"core.bare":                    "false",
		"core.logallrefupdates":        "true",
		"core.ignorecase":              "false",
		"core.precomposeunicode":       "false",
	}
	out, err := gitCommand("-C", root, "config", "--local", "--null", "--list").Output()
	if err != nil {
		return false
	}
	seen := map[string]bool{}
	for _, entry := range bytes.Split(bytes.TrimSuffix(out, []byte{0}), []byte{0}) {
		parts := bytes.SplitN(entry, []byte{'\n'}, 2)
		if len(parts) != 2 || seen[string(parts[0])] || expectedValues[string(parts[0])] != string(parts[1]) {
			return false
		}
		seen[string(parts[0])] = true
	}
	if len(seen) != len(expectedValues) {
		return false
	}
	expectedPaths := map[string]struct {
		dir  bool
		mode fs.FileMode
	}{
		".": {true, 0700}, "HEAD": {false, 0600}, "config": {false, 0600},
		"objects": {true, 0700}, "objects/info": {true, 0700}, "objects/pack": {true, 0700},
		"refs": {true, 0700}, "refs/heads": {true, 0700}, "refs/tags": {true, 0700},
	}
	ok := true
	seenPaths := map[string]bool{}
	_ = filepath.WalkDir(filepath.Join(root, ".git"), func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			ok = false
			return walkErr
		}
		rel, _ := filepath.Rel(filepath.Join(root, ".git"), path)
		rel = filepath.ToSlash(rel)
		expected, exists := expectedPaths[rel]
		info, infoErr := os.Lstat(path)
		if !exists || infoErr != nil || info.Mode()&os.ModeSymlink != 0 || info.IsDir() != expected.dir || info.Mode().Perm() != expected.mode {
			ok = false
		}
		seenPaths[rel] = true
		return nil
	})
	return ok && len(seenPaths) == len(expectedPaths)
}
func noUnexpected(root string, pl plan.CreationPlan, role string, allowMarker bool) bool {
	allowed := map[string]bool{".": true, ".git": true}
	if allowMarker {
		allowed[".my-friday/creation-state.json"] = true
	}
	for _, f := range pl.Files {
		if f.Role != role {
			continue
		}
		p := filepath.FromSlash(f.Path)
		for p != "." {
			allowed[p] = true
			p = filepath.Dir(p)
		}
	}
	ok := true
	filepath.WalkDir(root, func(p string, d fs.DirEntry, e error) error {
		if e != nil {
			ok = false
			return e
		}
		rel, _ := filepath.Rel(root, p)
		if rel == ".git" || strings.HasPrefix(rel, ".git"+string(filepath.Separator)) {
			return filepath.SkipDir
		}
		if !allowed[rel] {
			ok = false
		}
		return nil
	})
	return ok
}
