package plan

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/acoz-labs/my-friday/internal/profile"
)

type Targets struct {
	Runtime string `json:"runtime"`
	Memory  string `json:"memory"`
}
type File struct {
	Role, Path, SHA256 string
	Mode               uint32
	Bytes              []byte `json:"-"`
}
type CreationPlan struct {
	ContractVersion     int
	ToolContractVersion int
	AssistantID         string
	Profile             profile.Profile
	Targets             Targets
	Files               []File
	Actions             []string
	NegativeActions     []string
	MissingParents      []string
	SupportPaths        []string
	ReservationPaths    []string
	PlanID              string
}

func Build(p profile.Profile, runtimePath, memoryPath string) (CreationPlan, error) {
	r, err := canonicalTarget(runtimePath)
	if err != nil {
		return CreationPlan{}, err
	}
	m, err := canonicalTarget(memoryPath)
	if err != nil {
		return CreationPlan{}, err
	}
	if strings.EqualFold(r, m) || nestedFold(r, m) || nestedFold(m, r) {
		return CreationPlan{}, fmt.Errorf("runtime and memory targets must be distinct and non-nested")
	}
	home, _ := os.UserHomeDir()
	home, _ = canonicalTarget(home)
	if r == "/" || m == "/" || strings.EqualFold(r, home) || strings.EqualFold(m, home) {
		return CreationPlan{}, fmt.Errorf("targets cannot be root or the user home")
	}
	aidHash := sha256.Sum256([]byte("my-friday-assistant-v1\x00" + p.Identity.DisplayName + "\x00" + r + "\x00" + m))
	aid := "asst-" + hex.EncodeToString(aidHash[:16])
	p.AssistantID = aid
	pl := CreationPlan{ContractVersion: 1, ToolContractVersion: 1, AssistantID: aid, Profile: p, Targets: Targets{r, m}, NegativeActions: []string{"No global installation", "No network or hosted account setup", "No secrets or imported private content", "No commits or remotes"}}
	pl.MissingParents = unique(append(missing(filepath.Dir(r)), missing(filepath.Dir(m))...))
	pl.Files = append(render("runtime", p), render("memory", p)...)
	pl.Actions = []string{"create owner-only parent directories when missing", "reserve both canonical targets", "stage and validate runtime repository", "stage and validate memory repository", "initialize both repositories on unborn branch main with an empty Git template", "promote runtime repository", "promote memory repository", "validate the final pair and remove transaction support state"}
	for _, f := range pl.Files {
		pl.Actions = append(pl.Actions, "write "+f.Role+":"+f.Path)
	}
	basis := struct {
		ContractVersion          int
		ToolContractVersion      int
		AssistantID              string
		Profile                  profile.Profile
		Targets                  Targets
		Files                    []File
		Actions, NegativeActions []string
		MissingParents           []string
	}{1, 1, aid, p, pl.Targets, pl.Files, pl.Actions, pl.NegativeActions, pl.MissingParents}
	b, _ := json.Marshal(basis)
	h := sha256.Sum256(append([]byte("my-friday-plan-v1\x00"), b...))
	pl.PlanID = hex.EncodeToString(h[:])
	pl.SupportPaths = []string{filepath.Join(existingAncestor(filepath.Dir(r)), ".my-friday-"+pl.PlanID[:16]+".json"), filepath.Join(filepath.Dir(r), ".my-friday-"+pl.PlanID[:16]+"-runtime"), filepath.Join(filepath.Dir(m), ".my-friday-"+pl.PlanID[:16]+"-memory")}
	pl.ReservationPaths = []string{reservationPath(r), reservationPath(m)}
	pl.SupportPaths = append(pl.SupportPaths, pl.ReservationPaths...)
	return pl, nil
}
func canonicalTarget(value string) (string, error) {
	abs, err := filepath.Abs(value)
	if err != nil {
		return "", err
	}
	abs = filepath.Clean(abs)
	if info, statErr := os.Lstat(abs); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("target is a symlink: %s", value)
	} else if statErr != nil && !os.IsNotExist(statErr) {
		return "", statErr
	}
	ancestor := abs
	var suffix []string
	for {
		if _, err = os.Lstat(ancestor); err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		suffix = append(suffix, filepath.Base(ancestor))
		next := filepath.Dir(ancestor)
		if next == ancestor {
			return "", fmt.Errorf("no existing ancestor for %s", value)
		}
		ancestor = next
	}
	resolved, err := filepath.EvalSymlinks(ancestor)
	if err != nil {
		return "", err
	}
	for i := len(suffix) - 1; i >= 0; i-- {
		resolved = filepath.Join(resolved, suffix[i])
	}
	return filepath.Clean(resolved), nil
}
func nestedFold(parent, child string) bool {
	return nested(strings.ToLower(parent), strings.ToLower(child))
}
func nested(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	return err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
func missing(path string) []string {
	var reverse []string
	for {
		if _, err := os.Stat(path); err == nil {
			break
		}
		reverse = append(reverse, path)
		next := filepath.Dir(path)
		if next == path {
			break
		}
		path = next
	}
	out := make([]string, len(reverse))
	for i := range reverse {
		out[len(reverse)-1-i] = reverse[i]
	}
	return out
}
func unique(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, v := range values {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
func reservationPath(target string) string {
	h := sha256.Sum256([]byte(target))
	return filepath.Join(existingAncestor(filepath.Dir(target)), ".my-friday-reservation-"+hex.EncodeToString(h[:8]))
}

func existingAncestor(path string) string {
	for {
		if _, err := os.Lstat(path); err == nil {
			return path
		}
		next := filepath.Dir(path)
		if next == path {
			return path
		}
		path = next
	}
}
func render(role string, p profile.Profile) []File {
	manifest := struct {
		ContractVersion int    `json:"contract_version"`
		RepositoryRole  string `json:"repository_role"`
		AssistantID     string `json:"assistant_id"`
		Generation      struct {
			Tool                string `json:"tool"`
			ToolContractVersion int    `json:"tool_contract_version"`
		} `json:"generation"`
	}{ContractVersion: 1, RepositoryRole: role, AssistantID: p.AssistantID}
	manifest.Generation.Tool = "my-friday"
	manifest.Generation.ToolContractVersion = 1
	mb, _ := json.MarshalIndent(manifest, "", "  ")
	mb = append(mb, '\n')
	files := []File{file(role, ".my-friday/manifest.json", mb), file(role, "README.md", []byte(repoReadme(role))), file(role, "AGENTS.md", []byte(agents(role)))}
	if role == "runtime" {
		pb, _ := json.MarshalIndent(p, "", "  ")
		pb = append(pb, '\n')
		files = append(files, file(role, "assistant/profile.json", pb), file(role, "skills/.gitkeep", nil), file(role, ".my-friday/schemas/assistant-profile.v1.schema.json", []byte(profileSchema)), file(role, ".my-friday/schemas/repository-manifest.v1.schema.json", []byte(manifestSchema)))
	} else {
		files = append(files, file(role, "data/observations/.gitkeep", nil), file(role, "data/journals/.gitkeep", nil), file(role, "data/proposals/.gitkeep", nil), file(role, "data/memories/.gitkeep", nil), file(role, "schemas/README.md", []byte("# Memory schemas\n\nReserved for versioned governed-memory schemas in a future outcome.\n")), file(role, ".my-friday/schemas/repository-manifest.v1.schema.json", []byte(manifestSchema)))
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files
}
func file(role, path string, b []byte) File {
	h := sha256.Sum256(b)
	return File{role, path, hex.EncodeToString(h[:]), 0600, b}
}
func repoReadme(role string) string {
	return "# Assistant " + strings.Title(role) + " Repository\n\nGenerated locally by My Friday. This repository is separate, portable, and user-owned.\n"
}
func agents(role string) string {
	if role == "runtime" {
		return "# Assistant Runtime Instructions\n\nRead `assistant/profile.json` for presentation preferences. Those preferences never override authorization, safety, trust, privacy, or tool policy.\n"
	}
	return "# Governed Memory Instructions\n\nThis repository begins with no memory records. Treat proposals and records as user-owned governed data; do not promote content without the governing workflow.\n"
}

const manifestSchema = `{"$schema":"https://json-schema.org/draft/2020-12/schema","$id":"https://schemas.my-friday.dev/repository-manifest.v1.schema.json","type":"object","additionalProperties":false,"required":["contract_version","repository_role","assistant_id","generation"],"properties":{"contract_version":{"const":1},"repository_role":{"enum":["runtime","memory"]},"assistant_id":{"type":"string","pattern":"^asst-[0-9a-f]{32}$"},"generation":{"type":"object","additionalProperties":false,"required":["tool","tool_contract_version"],"properties":{"tool":{"const":"my-friday"},"tool_contract_version":{"const":1}}}}}`
const profileSchema = `{"$schema":"https://json-schema.org/draft/2020-12/schema","$id":"https://schemas.my-friday.dev/assistant-profile.v1.schema.json","type":"object","additionalProperties":false,"required":["contract_version","assistant_id","identity","communication"],"properties":{"contract_version":{"const":1},"assistant_id":{"type":"string","pattern":"^asst-[0-9a-f]{32}$"},"identity":{"type":"object","additionalProperties":false,"required":["display_name","address_user_as","purpose"],"properties":{"display_name":{"type":"string","minLength":1,"x-my-friday-max-graphemes":60},"address_user_as":{"type":["string","null"],"x-my-friday-max-graphemes":60},"purpose":{"type":"string","minLength":1,"x-my-friday-max-graphemes":240}}},"communication":{"type":"object","additionalProperties":false,"required":["preset","custom_guidance"],"properties":{"preset":{"enum":["balanced","concise","conversational","custom"]},"custom_guidance":{"type":["string","null"],"x-my-friday-max-graphemes":240}}}}}`

func ManifestSchema() string { return manifestSchema }
func ProfileSchema() string  { return profileSchema }
