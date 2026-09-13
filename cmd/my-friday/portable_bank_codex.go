package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/acoz-labs/my-friday/internal/console"
	"github.com/acoz-labs/my-friday/internal/memorybank"
	"github.com/acoz-labs/my-friday/internal/memoryinstall"
	"github.com/acoz-labs/my-friday/internal/portable"
)

func codexConnectionDefaults(home string) memoryinstall.Options {
	nativeHome := os.Getenv("CODEX_HOME")
	if nativeHome == "" {
		nativeHome = filepath.Join(home, ".codex")
	}
	codex, _ := exec.LookPath("codex")
	binary, _ := os.Executable()
	binding, _ := memorybank.DefaultBindingPath()
	return memoryinstall.Options{Home: home, CodexHome: nativeHome, Codex: codex, Binary: binary, Binding: binding}
}

func memoryCodexConnectionCLI(args []string, out, errout io.Writer) error {
	home, err := realHome()
	if err != nil {
		return err
	}
	o := codexConnectionDefaults(home)
	f := portableFlags("bank "+args[0], errout)
	f.StringVar(&o.Home, "installation-home", o.Home, "Home under which immutable My Friday artifacts are retained")
	f.StringVar(&o.CodexHome, "codex-home", o.CodexHome, "Existing native Codex profile; does not copy authentication")
	f.StringVar(&o.Codex, "codex", o.Codex, "Absolute native Codex executable")
	apply := false
	refresh := false
	if args[0] == "connect-codex" {
		f.StringVar(&o.Binding, "binding", o.Binding, "Selected machine-local memory bank binding")
		f.StringVar(&o.Binary, "binary", o.Binary, "Trusted My Friday executable to pin")
		f.BoolVar(&apply, "apply", false, "Install/reconnect after reviewing the default read-only JSON preview")
		f.BoolVar(&refresh, "refresh", false, "Repair with a fresh generated plugin copy; retain previous copies and edits")
	}
	if err = parseFlags(f, args[1:]); err != nil {
		return err
	}
	if args[0] == "doctor-codex" {
		r := memoryinstall.Doctor(context.Background(), o)
		if err = outputJSON(out, r); err != nil {
			return err
		}
		if !r.Healthy {
			return errors.New("Codex memory connection needs attention; see checks")
		}
		return nil
	}
	if refresh {
		o.Generation = portable.NewID("repair")
	}
	p, err := memoryinstall.Prepare(o)
	if err != nil {
		return err
	}
	if !apply {
		return outputJSON(out, map[string]any{"state": "preview", "connection": p, "requires_fresh_session": true, "notice": "No changes made. With --apply: retain a pinned runtime/plugin, register its marketplace and install through native Codex. No memory writes/sync or authentication/shell changes. Review hook trust in a fresh session. Explicit memory environment overrides take precedence."})
	}
	r, err := memoryinstall.Apply(context.Background(), p)
	if e := outputJSON(out, r); e != nil {
		return e
	}
	return err
}

func (u managementUI) memoryConnectCodex(check, refresh bool) error {
	o := codexConnectionDefaults(u.home)
	var err error
	title := "Connect or update Codex memory"
	if check {
		title = "Check Codex memory connection"
	}
	if refresh {
		title = "Repair Codex memory connection"
	}
	u.section(title, "Uses your native Codex profile. No agent launcher or replacement login/settings is created.")
	o.CodexHome, err = u.ask("Native Codex profile directory", o.CodexHome)
	if err != nil {
		return err
	}
	o.Codex, err = u.ask("Codex executable", o.Codex)
	if err != nil {
		return err
	}
	if check {
		r := memoryinstall.Doctor(context.Background(), o)
		for _, c := range r.Checks {
			tone := console.Success
			if !c.OK {
				tone = console.Warning
			}
			u.block(console.Block{Title: c.Name, Body: c.Detail, Tone: tone})
		}
		u.section("Check limits", r.Notice)
		if !r.Healthy {
			u.section("Next step", "Use Connect or update Codex memory after resolving the reported problem. Existing edited files are preserved, not silently overwritten.")
		}
		return nil
	}
	if refresh {
		r := memoryinstall.Doctor(context.Background(), o)
		if r.Connection != nil {
			o.Binding = r.Connection.Binding
		}
		o.Generation = portable.NewID("repair")
		u.section("Repair scope", "Create a fresh generated plugin copy using this running toolkit. Retain prior copies, edits and cached versions for inspection. Confirm the selected bank below; no memory repair or migration is performed.")
	}
	o.Binding, err = u.ask("Memory binding file", o.Binding)
	if err != nil {
		return err
	}
	p, err := memoryinstall.Prepare(o)
	if err != nil {
		return err
	}
	yes, err := u.review("Retain this My Friday runtime and a machine-local plugin copy; use native Codex to install or reconnect only the My Friday memory plugin.", "Memory, Git remotes, native login, other plugins, shell startup files, existing Alfred installations and retained runtime versions.", "Close affected Codex sessions before changing their connection. Start a fresh session afterward and review hook trust/MCP status. Explicit MY_FRIDAY_MEMORY_BIN/BINDING environment overrides still win.", console.Field{Label: "Bank", Value: p.BankID}, console.Field{Label: "Binding", Value: p.Binding}, console.Field{Label: "Native profile", Value: p.CodexHome}, console.Field{Label: "Pinned runtime", Value: p.Runtime}, console.Field{Label: "Plugin", Value: p.PluginID})
	if err != nil || !yes {
		return err
	}
	r, err := memoryinstall.Apply(context.Background(), p)
	if err != nil {
		return err
	}
	u.block(console.Block{Title: "Codex memory connected", Body: r.Notice, Tone: console.Success})
	u.section("Next step", "Launch Codex normally in any project. In a fresh session, review /hooks and check /mcp. If using a non-default native profile, select that same CODEX_HOME. No custom assistant command is required.")
	return nil
}
