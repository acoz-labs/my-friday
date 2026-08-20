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
		cmd := exec.Command("git", "init", "--quiet", "--initial-branch=main", "--template="+tmpl, target.path)
		cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "LANG=C.UTF-8"}
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git init %s: %w: %s", target.role, err, strings.TrimSpace(string(out)))
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
	cmd := exec.Command("git", "-C", root, "rev-parse", "--git-dir")
	if out, err := cmd.Output(); err != nil || strings.TrimSpace(string(out)) != ".git" {
		return fmt.Errorf("repository is not a local Git repository")
	}
	return nil
}

func validateFreshGit(root string) error {
	cmd := exec.Command("git", "-C", root, "symbolic-ref", "--short", "HEAD")
	out, e := cmd.Output()
	if e != nil || strings.TrimSpace(string(out)) != "main" {
		return fmt.Errorf("repository is not an unborn main branch")
	}
	if exec.Command("git", "-C", root, "rev-parse", "--verify", "HEAD").Run() == nil {
		return fmt.Errorf("repository must have no commits")
	}
	if out, e = exec.Command("git", "-C", root, "remote").Output(); e != nil || len(bytes.TrimSpace(out)) != 0 {
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
	allowedValues := map[string]map[string]bool{
		"core.repositoryformatversion": {"0": true},
		"core.filemode":                {"true": true, "false": true},
		"core.bare":                    {"false": true},
		"core.logallrefupdates":        {"true": true},
		"core.ignorecase":              {"true": true, "false": true},
		"core.precomposeunicode":       {"true": true, "false": true},
	}
	out, err := exec.Command("git", "-C", root, "config", "--local", "--null", "--list").Output()
	if err != nil {
		return false
	}
	for _, entry := range bytes.Split(bytes.TrimSuffix(out, []byte{0}), []byte{0}) {
		parts := bytes.SplitN(entry, []byte{'\n'}, 2)
		if len(parts) != 2 || !allowedValues[string(parts[0])][string(parts[1])] {
			return false
		}
	}
	allowedPaths := map[string]bool{
		".": true, "HEAD": true, "config": true, "branches": true, "hooks": true,
		"info": true, "objects": true, "objects/info": true, "objects/pack": true,
		"refs": true, "refs/heads": true, "refs/tags": true,
	}
	ok := true
	_ = filepath.WalkDir(filepath.Join(root, ".git"), func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			ok = false
			return walkErr
		}
		rel, _ := filepath.Rel(filepath.Join(root, ".git"), path)
		if !allowedPaths[filepath.ToSlash(rel)] {
			ok = false
		}
		return nil
	})
	return ok
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
