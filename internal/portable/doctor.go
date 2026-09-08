package portable

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
)

type Diagnostic struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
	Remedy string `json:"remedy,omitempty"`
}

type DoctorReport struct {
	Healthy  bool         `json:"healthy"`
	Instance string       `json:"instance"`
	Harness  string       `json:"harness"`
	Checks   []Diagnostic `json:"checks"`
	Notice   string       `json:"notice"`
}

// Doctor is deliberately local and read-only: no commands, credential reads,
// sync helpers, native initialization, or repair. Healthy is structural only.
func (i Instance) Doctor(s *Store, harness string) DoctorReport {
	if harness == "" {
		harness = s.Agent.DefaultHarness
	}
	r := DoctorReport{Healthy: true, Instance: i.Root, Harness: harness, Checks: []Diagnostic{}, Notice: "Read-only structural checks; authentication, network access, launcher placement, harness version compatibility, active-session context and external capability effects are not tested. Generated files are compared with this running toolkit, not the bound executable's version."}
	toolkit, err := os.Executable()
	if err != nil {
		toolkit = i.Binary
	}
	repair := shellQuote(toolkit) + " agent repair --instance " + shellQuote(i.Root)
	add := func(name string, ok bool, detail, remedy string) {
		if !ok {
			r.Healthy = false
		} else {
			remedy = ""
		}
		r.Checks = append(r.Checks, Diagnostic{Name: name, OK: ok, Detail: detail, Remedy: remedy})
	}
	if err := s.Validate(); err != nil {
		// Source errors can contain user-authored content; avoid echoing it.
		add("source", false, "Assistant source validation failed.", "Run agent validate --repository PATH and reconcile source before repair.")
		return r
	}
	add("source", true, "Assistant source and memory structure valid.", "")
	gitInfo, gitErr := os.Lstat(filepath.Join(s.Root, ".git"))
	add("source-git", gitErr == nil && gitInfo.IsDir(), "Assistant requires its own .git directory for synchronization.", "Restore the source repository's Git metadata from a known backup, or import a separate verified clone; preserve uncommitted local records. Repair does not rebuild Git history.")
	info, err := os.Stat(i.Binary)
	add("binary", err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0, "Bound toolkit: "+i.Binary, "Restore a compatible executable at the bound path. Repair does not replace the binary.")
	if harness != "codex" && harness != "pi" {
		add("harness", false, "Unsupported harness selection.", "Select codex or pi.")
	} else {
		path, err := exec.LookPath(harness)
		add("harness", err == nil, "Selected harness on this process's PATH: "+path, "Make the selected harness available on PATH in the terminal used to launch the assistant.")
	}
	if err := i.preflightProjection(s); err != nil {
		add("projection-paths", false, err.Error(), "Inspect the conflicting path; preserve its contents. Repair refuses symlink and non-regular targets.")
		return r
	}
	add("projection-paths", true, "Generated paths have no detected type or overlap conflicts.", "")
	files, err := i.projection(s)
	if err != nil {
		add("projection", false, "Could not render generated instructions.", "Inspect source instructions before repair.")
		return r
	}
	for _, name := range projectionFiles {
		// Preflight checked all parent types. This is not a same-UID race defense.
		data, err := os.ReadFile(filepath.Join(i.Root, name))
		ok := err == nil && bytes.Equal(data, files[name])
		detail := "Matches the running toolkit and current assistant instructions."
		if err != nil {
			detail = "Generated file is missing or unreadable."
		} else if !ok {
			detail = "Generated file differs from the running toolkit or current assistant instructions."
		}
		add("projection:"+name, ok, detail, "Run: "+repair+" ; then start a fresh session. This regenerates all managed files, not native authentication or sessions.")
	}
	return r
}
