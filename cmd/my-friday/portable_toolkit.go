package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/acoz-labs/my-friday/internal/console"
	"github.com/acoz-labs/my-friday/internal/portable"
	"github.com/acoz-labs/my-friday/internal/toolkitupdate"
)

type toolkitBuild struct {
	Product            string `json:"product"`
	PortableFormat     int    `json:"portable_format"`
	ManagementProtocol int    `json:"management_protocol"`
	AgentAPIVersion    int    `json:"agent_api_version,omitempty"`
	Revision           string `json:"revision"`
	Modified           bool   `json:"modified"`
	OS                 string `json:"os"`
	Arch               string `json:"arch"`
}

func toolkitVersion() toolkitBuild {
	v := toolkitBuild{Product: "my-friday", PortableFormat: portable.FormatVersion, ManagementProtocol: 1, AgentAPIVersion: 1, Revision: "development", OS: runtime.GOOS, Arch: runtime.GOARCH}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				v.Revision = setting.Value
			}
			if setting.Key == "vcs.modified" {
				v.Modified = setting.Value == "true"
			}
		}
	}
	return v
}

type toolkitOutput struct{ bytes.Buffer }

func (b *toolkitOutput) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 1<<20 {
		return 0, errors.New("toolkit response too large")
	}
	return b.Buffer.Write(p)
}
func runToolkit(binary string, args ...string) ([]byte, error) {
	return runToolkitContext(context.Background(), binary, args...)
}

func runToolkitContext(parent context.Context, binary string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = time.Second
	var output toolkitOutput
	cmd.Stdout = &output
	cmd.Stderr = io.Discard
	cmd.Stdin = nil
	if err := cmd.Run(); err != nil {
		return nil, errors.New("toolkit compatibility check failed or timed out")
	}
	return output.Bytes(), nil
}
func verifyToolkit(path string) error {
	return verifyToolkitContext(context.Background(), path)
}

func verifyToolkitContext(ctx context.Context, path string) error {
	data, err := runToolkitContext(ctx, path, "--version")
	if err != nil {
		return err
	}
	var v toolkitBuild
	if json.Unmarshal(data, &v) != nil || v.Product != "my-friday" || v.PortableFormat != 1 || v.ManagementProtocol != 1 || v.OS != runtime.GOOS || v.Arch != runtime.GOARCH {
		return errors.New("artifact is not a compatible My Friday management toolkit for this machine")
	}
	return nil
}

func portableToolkit(args []string, input io.Reader, out, errout io.Writer) error {
	if len(args) == 0 {
		home, err := realHome()
		if err != nil {
			return err
		}
		err = newManagementUI(home, input, out, false).updates()
		if errors.Is(err, errToolkitActivated) || errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	f := portableFlags("toolkit "+args[0], errout)
	instance := f.String("instance", "", "Agent instance")
	launcher := f.String("launcher", "", "Exact existing managed launcher; omitted leaves launchers unchanged")
	binary := f.String("binary", "", "For manifest: approved executable")
	release := f.String("release", "", "For manifest: release tag")
	if err := parseFlags(f, args[1:]); err != nil {
		return err
	}
	switch args[0] {
	case "check-instance":
		_, s, err := portable.LoadInstance(*instance)
		if err != nil {
			return err
		}
		if err := s.ValidateToolkitCompatibility(); err != nil {
			return err
		}
		return outputJSON(out, map[string]bool{"compatible": true})
	case "use":
		i, s, err := portable.LoadInstance(*instance)
		if err != nil {
			return err
		}
		current, err := os.Executable()
		if err != nil {
			return err
		}
		backup, err := i.UseToolkit(s, current, *launcher)
		if err != nil {
			return err
		}
		return outputJSON(out, map[string]string{"backup": backup, "notice": "Agent re-bound to running toolkit. Source/native settings preserved; start a fresh session."})
	case "manifest":
		if *release == "" || *binary == "" {
			return errors.New("manifest requires --binary and --release; this prints metadata only and never publishes")
		}
		if err := verifyToolkit(*binary); err != nil {
			return err
		}
		hash, err := toolkitupdate.Digest(*binary)
		if err != nil {
			return err
		}
		return outputJSON(out, toolkitupdate.Manifest{SchemaVersion: 1, PortableFormat: 1, ManagementProtocol: 1, Version: *release, Artifacts: []toolkitupdate.Artifact{{OS: runtime.GOOS, Arch: runtime.GOARCH, Name: "my-friday-" + runtime.GOOS + "-" + runtime.GOARCH, SHA256: hash}}})
	default:
		return errors.New("usage: my-friday toolkit [check-instance|use|manifest] (no subcommand opens update menu)")
	}
}

func (u managementUI) updates() error {
	for {
		v := toolkitVersion()
		u.section("Running My Friday", "", console.Field{Label: "Revision", Value: v.Revision}, console.Field{Label: "Platform", Value: v.OS + "/" + v.Arch})
		n, err := u.choose("Update My Friday", []string{"Check latest compatible release", "Install an approved local artifact", "Switch to a retained toolkit version"}, "Back")
		if err != nil || n == 0 {
			return err
		}
		switch n {
		case 1:
			err = u.latestRelease()
		case 2:
			err = u.localArtifact()
		case 3:
			err = u.retainedToolkit()
		}
		if errors.Is(err, errToolkitActivated) {
			return err
		}
		if err := u.problem(err); err != nil {
			return err
		}
	}
}

var errToolkitActivated = errors.New("toolkit activated")

func (u managementUI) activate(path string) error {
	if err := verifyToolkit(path); err != nil {
		return err
	}
	current, err := os.Executable()
	if err != nil {
		return err
	}
	backup, err := toolkitupdate.Activate(u.home, path, current)
	if err != nil {
		return err
	}
	u.block(console.Block{Title: "Management command updated", Body: "Agents keep their existing pins.", Tone: console.Success, Fields: []console.Field{{Label: "Selected toolkit", Value: path}, {Label: "Previous command checkpoint", Value: backup}}})
	u.section("Next step", "Exit and run my-friday again to use that version. To update one agent, choose Manage an installed agent → Use this toolkit version.")
	return errToolkitActivated
}
func (u managementUI) latestRelease() error {
	fmt.Fprintln(u.out, "Checking the official latest release (read-only; no account credentials used)…")
	client := toolkitupdate.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	a, err := client.Latest(ctx)
	if err != nil {
		return err
	}
	ok, err := u.review("Download, verify and execute the artifact's compatibility check, then update the My Friday command.", "Agent versions stay pinned. Agent source and native logins are not changed.", "Reopen my-friday after installation. Agent adoption is a separate action.", console.Field{Label: "Release", Value: a.Version}, console.Field{Label: "Artifact", Value: a.Name}, console.Field{Label: "SHA-256", Value: a.SHA256})
	if err != nil || !ok {
		return err
	}
	// The user may spend arbitrarily long deciding; give download its own deadline.
	downloadCtx, stop := context.WithTimeout(context.Background(), 45*time.Second)
	defer stop()
	path, err := client.Download(downloadCtx, u.home, a)
	if err != nil {
		return err
	}
	return u.activate(path)
}
func (u managementUI) localArtifact() error {
	u.section("Local artifact", "For a separately approved development/offline artifact. A checksum verifies bytes, not publisher trust. Only select an executable you trust to run.")
	path, err := u.ask("Artifact file", "")
	if err != nil {
		return err
	}
	hash, err := u.ask("Expected SHA-256 from your trusted artifact record", "")
	if err != nil {
		return err
	}
	ok, err := u.review("Verify and install this artifact, execute its compatibility check, and update the My Friday command.", "Existing agent versions remain pinned. Agent source and native logins are not changed.", "Reopen my-friday after installation. Agent adoption is a separate action.", console.Field{Label: "Artifact", Value: path}, console.Field{Label: "Expected SHA-256", Value: hash})
	if err != nil || !ok {
		return err
	}
	installed, err := toolkitupdate.Stage(u.home, path, hash)
	if err != nil {
		return err
	}
	return u.activate(installed)
}
func (u managementUI) retainedToolkit() error {
	directory := filepath.Join(u.home, ".local/share/my-friday/releases")
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		fmt.Fprintln(u.out, "No retained versions.")
		return nil
	}
	if err != nil {
		return err
	}
	paths := []string{}
	names := []string{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(directory, entry.Name(), "my-friday")
		if st, err := os.Lstat(path); err == nil && st.Mode().IsRegular() {
			paths = append(paths, path)
			names = append(names, entry.Name())
		}
	}
	n, err := u.choose("Retained toolkits", names, "Back")
	if err != nil || n == 0 {
		return err
	}
	ok, err := u.review("Execute this retained toolkit's compatibility check and point the My Friday command to it.", "No agent, memory or external action is rolled back.", "Reopen my-friday to use the selected toolkit. Agent adoption is separate.", console.Field{Label: "Selected toolkit", Value: paths[n-1]})
	if err != nil || !ok {
		return err
	}
	return u.activate(paths[n-1])
}
