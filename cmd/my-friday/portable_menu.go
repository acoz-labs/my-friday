package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/acoz-labs/my-friday/internal/console"
	"github.com/acoz-labs/my-friday/internal/portable"
)

var errMenuBack = errors.New("back to menu")

func promptReader(input io.Reader) *bufio.Reader {
	if reader, ok := input.(*bufio.Reader); ok {
		return reader
	}
	return bufio.NewReader(input)
}

type managementUI struct {
	reader *bufio.Reader
	out    io.Writer
	home   string
	tui    *console.Console
	screen io.Writer
}

func (u managementUI) block(b console.Block) {
	if u.tui != nil {
		u.tui.Block(b)
		return
	}
	fmt.Fprint(u.out, console.RenderBlock(b, console.Theme{}, console.ReportWidth(u.screen)))
}

func (u managementUI) section(title, body string, fields ...console.Field) {
	u.block(console.Block{Title: title, Body: body, Fields: fields, Tone: console.Plain})
}

func (u managementUI) review(change, preserved, sessions string, fields ...console.Field) (bool, error) {
	u.section("What will change", change, fields...)
	u.section("What stays untouched", preserved)
	u.section("Session guidance", sessions)
	return u.confirm("")
}

func (u managementUI) line(tone console.Tone, text string) {
	if u.tui != nil {
		u.tui.Println(tone, text)
		return
	}
	fmt.Fprintln(u.out, text)
}

func (u managementUI) ask(label, def string) (string, error) {
	if u.tui != nil {
		value, err := u.tui.Input(label, def)
		if errors.Is(err, console.ErrBack) || value == ":back" {
			return "", errMenuBack
		}
		return value, err
	}
	if def != "" {
		fmt.Fprintf(u.out, "%s [%s]: ", label, def)
	} else {
		fmt.Fprintf(u.out, "%s: ", label)
	}
	line, err := u.reader.ReadString('\n')
	if err != nil {
		return "", io.EOF
	}
	line = strings.TrimSpace(line)
	if line == ":back" {
		return "", errMenuBack
	}
	if line == "" {
		line = def
	}
	return line, nil
}
func (u managementUI) choose(title string, items []string, back string) (int, error) {
	if u.tui != nil {
		index, err := u.tui.Select(title, append(append([]string{}, items...), back), 0)
		if errors.Is(err, console.ErrBack) || (err == nil && index == len(items)) {
			return 0, nil
		}
		return index + 1, err
	}
	for {
		fmt.Fprintf(u.out, "\n%s\n\n", title)
		for n, item := range items {
			fmt.Fprintf(u.out, "  %d. %s\n", n+1, item)
		}
		fmt.Fprintf(u.out, "  0. %s\n\n", back)
		text, err := u.ask("Choose a number", "0")
		if errors.Is(err, errMenuBack) {
			return 0, nil
		}
		if err != nil {
			return 0, err
		}
		n, err := strconv.Atoi(text)
		if err == nil && n >= 0 && n <= len(items) {
			return n, nil
		}
		fmt.Fprintln(u.out, "Please choose one of the listed numbers.")
	}
}
func (u managementUI) confirm(summary string) (bool, error) {
	if summary != "" {
		u.section("Review", summary)
	}
	if u.tui != nil {
		index, err := u.tui.Select("Continue?", []string{"No, go back", "Yes, continue"}, 0)
		if errors.Is(err, console.ErrBack) {
			return false, errMenuBack
		}
		return index == 1 && err == nil, err
	}
	for {
		answer, err := u.ask("Continue? (yes/no)", "no")
		if err != nil {
			return false, err
		}
		if answer == "yes" {
			return true, nil
		}
		if answer == "no" {
			return false, nil
		}
		fmt.Fprintln(u.out, "Enter yes or no.")
	}
}
func (u managementUI) problem(err error) error {
	if errors.Is(err, io.EOF) {
		return io.EOF
	}
	if err != nil && !errors.Is(err, errMenuBack) {
		u.block(console.Block{Title: "Needs attention", Body: err.Error(), Tone: console.Failure})
		u.section("Next step", "Inspect the reported issue before retrying. Earlier successful changes are retained.")
	}
	return nil
}

func managementMenu(home string, input io.Reader, out io.Writer) error {
	return managementMenuMode(home, input, out, false)
}

func managementMenuMode(home string, input io.Reader, out io.Writer, plain bool) error {
	u := newManagementUI(home, input, out, plain)
	u.line(console.Accent, "My Friday — agent management")
	u.line(console.Muted, "Nothing changes just by opening this menu.")
	if u.tui == nil {
		fmt.Fprintln(out, "Plain mode: choose a number; :back cancels a prompt.")
	}
	for {
		n, err := u.choose("My Friday", []string{"Set up a new agent", "Import an existing agent", "Manage an installed agent", "Update My Friday"}, "Exit")
		if errors.Is(err, io.EOF) || (err == nil && n == 0) {
			return nil
		}
		if err != nil {
			return err
		}
		switch n {
		case 1:
			err = u.setup(false)
		case 2:
			err = u.setup(true)
		case 3:
			err = u.agents()
		case 4:
			err = u.updates()
		}
		if errors.Is(err, errToolkitActivated) {
			return nil
		}
		if errors.Is(u.problem(err), io.EOF) {
			return nil
		}
	}
}

func newManagementUI(home string, input io.Reader, out io.Writer, plain bool) managementUI {
	u := managementUI{reader: promptReader(input), out: console.SafeWriter{Output: out}, home: home, screen: out}
	if !plain {
		u.tui = console.New(input, out)
	}
	return u
}

func (u managementUI) setup(importing bool) error {
	title := "New agent"
	if importing {
		title = "Import an existing agent"
	}
	u.line(console.Heading, "\n"+title)
	source := ""
	var err error
	if importing {
		fmt.Fprintln(u.out, "Choose an already-cloned My Friday agent repository. Legacy memory belongs in reference libraries, not direct import. Remote cloning is not yet provided here.")
		source, err = u.ask("Local agent repository", "")
		if err != nil {
			return err
		}
	}
	var name string
	for {
		name, err = u.ask("Agent command name", "friday")
		if err != nil {
			return err
		}
		if err = portable.ValidateLauncherName(name); err == nil {
			break
		}
		fmt.Fprintln(u.out, err)
	}
	harness := "codex"
	if !importing {
		for {
			harness, err = u.ask("Default harness (codex/pi)", harness)
			if err != nil {
				return err
			}
			if harness == "codex" || harness == "pi" {
				break
			}
			fmt.Fprintln(u.out, "Choose codex or pi.")
		}
		source = filepath.Join(u.home, ".local/share/my-friday/repositories", name)
	}
	instance := filepath.Join(u.home, ".local/share/my-friday/instances", name)
	launcher := filepath.Join(u.home, ".local/bin", name)
	host, _ := os.Hostname()
	label, err := u.ask("Machine label", host)
	if err != nil {
		return err
	}
	change := "Create a new agent source, instance and launcher."
	if importing {
		change = "Register this machine in the imported source, checkpoint it, and create an instance and launcher."
	}
	ok, err := u.review(change, "Existing destination paths will not be overwritten. Native logins and credentials are not copied.", "Launch the agent after setup; native harness login remains separate.", console.Field{Label: "Agent", Value: name}, console.Field{Label: "Source", Value: source}, console.Field{Label: "Instance", Value: instance}, console.Field{Label: "Command", Value: launcher})
	if err != nil || !ok {
		return err
	}
	args := []string{"--name", name, "--state", instance, "--launcher", launcher, "--device-label", label}
	if importing {
		args = append(args, "--import", source)
	} else {
		args = append(args, "--repository", source, "--harness", harness)
	}
	// Noninteractive primitive retains its JSON contract; the menu presents a
	// human summary and uses the created binding for subsequent management.
	if err := portableSetup(args, u.reader, io.Discard, u.out); err != nil {
		return err
	}
	u.block(console.Block{Title: "Installed", Body: name + " is installed. No credentials were copied.", Tone: console.Success})
	u.section("Next step", "Complete native harness login if needed, then configure repository synchronization. For imported capability prerequisites, choose Prepare this machine; setup has not executed private preparation scripts.")
	return u.agent(instance)
}

func (u managementUI) agents() error {
	for {
		entries, err := portable.DiscoverInstances(filepath.Join(u.home, ".local/share/my-friday/instances"))
		if err != nil {
			return err
		}
		labels := []string{}
		for _, item := range entries {
			label := item.Name
			if item.Problem != "" {
				label += " — " + item.Problem
			}
			labels = append(labels, label)
		}
		labels = append(labels, "Open an installation at another path")
		if len(entries) == 0 {
			fmt.Fprintln(u.out, "No agents found in the standard installation directory.")
		}
		n, err := u.choose("Installed agents", labels, "Back")
		if err != nil || n == 0 {
			return err
		}
		path := ""
		if n == len(labels) {
			path, err = u.ask("Instance directory", "")
		} else {
			path = entries[n-1].Path
		}
		if err == nil {
			err = u.agent(path)
		}
		if err := u.problem(err); err != nil {
			return err
		}
	}
}

func (u managementUI) agent(path string) error {
	for {
		i, s, err := portable.LoadInstance(path)
		if err != nil {
			return fmt.Errorf("cannot load installation %s; preserve its files and inspect binding/source: %w", path, err)
		}
		if u.tui != nil {
			pin := "custom path (see status)"
			parent := filepath.Dir(i.Binary)
			if filepath.Base(filepath.Dir(parent)) == "releases" {
				pin = filepath.Base(parent)
			}
			u.tui = u.tui.WithContext(s.Agent.DefaultHarness + " · pinned: " + pin)
		}
		n, err := u.choose(i.Name, []string{"View status", "Configure repository and synchronization", "Change default harness", "Check installation health", "Repair installation", "Use this toolkit version for this agent", "Roll back the last toolkit change", "Reference sources", "Prepare this machine"}, "Back")
		if err != nil || n == 0 {
			return err
		}
		switch n {
		case 1:
			u.section("Agent: "+i.Name, "", console.Field{Label: "Default harness", Value: s.Agent.DefaultHarness}, console.Field{Label: "Pinned toolkit", Value: i.Binary})
			u.section("Storage", "", console.Field{Label: "Source", Value: s.Root}, console.Field{Label: "Persistent memory", Value: filepath.Join(s.Root, "memory")}, console.Field{Label: "Instance", Value: i.Root})
			state, e := s.RemoteSetupStatus(context.Background())
			if e != nil {
				u.block(console.Block{Title: "Source backup — needs attention", Body: e.Error(), Tone: console.Warning})
				u.section("Next step", "Resolve the source setup finding before configuring or retrying synchronization.")
			} else if state.Origin == "" {
				u.block(console.Block{Title: "Source backup", Body: "Local only (no remote).", Tone: console.Warning})
				u.section("Next step", "Choose Configure repository and synchronization when you are ready to enable remote backup.")
			} else {
				u.section("Source backup", "Configured, not verified by this status view.", console.Field{Label: "Remote", Value: state.Origin})
				u.section("Next step", "Use source synchronization to verify fetch/push; a configured URL does not establish a current remote backup.")
			}
		case 2:
			err = remoteSetupWizardUI(u.reader, u.out, i, s, true, &u)
		case 3:
			var harness string
			for {
				harness, err = u.ask("Default harness (codex/pi)", s.Agent.DefaultHarness)
				if err != nil || harness == "codex" || harness == "pi" {
					break
				}
				fmt.Fprintln(u.out, "Choose codex or pi.")
			}
			if err == nil {
				var ok bool
				ok, err = u.review("Checkpoint the portable default locally; other machines receive it after synchronization.", "Existing sessions and native logins remain unchanged.", "The new default applies to future launches, not this conversation.", console.Field{Label: "Previous harness", Value: s.Agent.DefaultHarness}, console.Field{Label: "Selected harness", Value: harness})
				if err == nil && ok {
					err = s.WithCheckpointObserver(portable.Authorship{DeviceID: i.DeviceID, Actor: s.Agent.Name, Harness: "management"}).SetDefaultHarness(context.Background(), harness)
					if err == nil {
						u.block(console.Block{Title: "Default saved", Body: "Default harness saved and checkpointed locally.", Tone: console.Success})
						u.section("Next step", "Launch a fresh agent session to use the default. It will propagate at the next normal sync.")
					}
				}
			}
		case 4:
			u.doctor(i, s, false)
		case 5:
			var ok bool
			ok, err = u.review("Refresh managed instructions and hooks. Network and authentication failures are not repaired.", "Memory, credentials, native settings and stored sessions are preserved.", "Close active agent sessions first; start a fresh session after repair.")
			if err == nil && ok {
				err = s.Validate()
				if err == nil {
					err = i.Project(s)
				}
				if err == nil {
					u.block(console.Block{Title: "Repair complete", Body: "Managed files refreshed.", Tone: console.Success})
					launcher := filepath.Join(u.home, ".local/bin", i.Name)
					if _, missing := os.Lstat(launcher); os.IsNotExist(missing) {
						create, askErr := u.review("Create the missing default launcher.", "Custom launchers elsewhere will not be changed.", "Start a fresh agent session after repair.", console.Field{Label: "Launcher", Value: launcher})
						if askErr != nil {
							err = askErr
						} else if create {
							err = i.InstallLauncher(launcher)
						}
					}
					u.doctor(i, s, true)
				}
			}
		case 6:
			err = u.adopt(i, s)
		case 7:
			err = u.rollback(i, s)
		case 8:
			err = u.references(i.Root)
		case 9:
			err = u.machine(i.Root)
		}
		if err := u.problem(err); err != nil {
			return err
		}
	}
}

func (u managementUI) doctor(i portable.Instance, s *portable.Store, repaired bool) {
	r := i.Doctor(s, "")
	passed := 0
	for _, check := range r.Checks {
		if check.OK {
			passed++
		}
	}
	tone, summary := console.Warning, "Needs attention"
	if r.Healthy {
		tone, summary = console.Success, "Healthy"
	}
	u.block(console.Block{Title: "Installation health", Body: fmt.Sprintf("%s — %d of %d local checks passed.", summary, passed, len(r.Checks)), Tone: tone})
	for _, check := range r.Checks {
		if check.OK {
			continue
		}
		fields := []console.Field{}
		if check.Remedy != "" {
			fields = append(fields, console.Field{Label: "Remedy", Value: check.Remedy})
		}
		u.block(console.Block{Title: "Needs attention — " + check.Name, Body: check.Detail, Fields: fields, Tone: console.Warning})
	}
	scope := "Read-only local structural checks. Authentication and remote synchronization were not tested."
	current, _ := os.Executable()
	current, _ = filepath.EvalSymlinks(current)
	if current != i.Binary {
		scope += " This compares against the running toolkit; the agent is pinned to another executable. Adoption is a separate action."
	}
	u.section("Scope", scope)
	if len(r.MachineRequirements) > 0 {
		u.section("Machine readiness", "Historical receipts only; no private scripts were run. Use Prepare this machine for an explicit check or targeted preparation. Structural repair does not install dependencies.")
		for _, st := range r.MachineRequirements {
			u.section(st.CapabilityID+" / "+st.Requirement.ID, st.State, console.Field{Label: "Last observation", Value: st.CheckedAt})
		}
	}
	if repaired {
		u.section("Next step", "Start a fresh agent session after resolving any remaining findings above.")
	} else if r.Healthy {
		u.section("Next step", "No structural repair is needed. Return to the agent menu.")
	} else {
		u.section("Next step", "Follow the remedies above. Repair installation refreshes managed instructions/hooks; other findings need their specific remedy.")
	}
}

func (u managementUI) adopt(i portable.Instance, s *portable.Store) error {
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	binary, err = filepath.EvalSymlinks(binary)
	if err != nil {
		return err
	}
	if binary == i.Binary {
		fmt.Fprintln(u.out, "This agent already uses the running toolkit. Use Repair for generated-file problems.")
		return nil
	}
	launcher, err := u.ask("Managed launcher path (or none for an installation without a launcher)", filepath.Join(u.home, ".local/bin", i.Name))
	if err != nil {
		return err
	}
	if launcher == "none" {
		launcher = ""
	}
	launcherLabel := launcher
	if launcherLabel == "" {
		launcherLabel = "Unchanged (no launcher selected)"
	}
	ok, err := u.review("Back up and update this agent's binding, managed files and selected launcher.", "Source, native settings/logins and other agent pins stay unchanged.", "Close active "+i.Name+" sessions first; start a fresh session after adoption.", console.Field{Label: "Previous toolkit", Value: i.Binary}, console.Field{Label: "Selected toolkit", Value: binary}, console.Field{Label: "Launcher", Value: launcherLabel})
	if err != nil || !ok {
		return err
	}
	backup, err := i.UseToolkit(s, binary, launcher)
	if err != nil {
		return err
	}
	u.block(console.Block{Title: "Toolkit adopted", Body: "Agent toolkit updated. Other agents retain their existing versions.", Tone: console.Success, Fields: []console.Field{{Label: "Rollback files", Value: backup}}})
	u.section("Next step", "Start a fresh agent session.")
	return nil
}

func (u managementUI) rollback(i portable.Instance, s *portable.Store) error {
	directory := filepath.Join(i.Root, "toolkit-updates")
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		fmt.Fprintln(u.out, "No toolkit rollback checkpoints for this agent.")
		return nil
	}
	if err != nil {
		return err
	}
	backups := []string{}
	labels := []string{}
	for n := len(entries) - 1; n >= 0; n-- {
		entry := entries[n]
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		r, err := i.ReadToolkitReceipt(path)
		if err != nil {
			continue
		}
		backups = append(backups, path)
		labels = append(labels, entry.Name()+" → "+r.Before.Binary)
	}
	n, err := u.choose("Toolkit rollback checkpoints", labels, "Back")
	if err != nil || n == 0 {
		return err
	}
	r, err := i.ReadToolkitReceipt(backups[n-1])
	if err != nil {
		return err
	}
	ok, err := u.review("Verify the retained executable against today's source, then restore only unchanged managed files from this checkpoint.", "Memory, source, native credentials and the management command are not rolled back.", "Close active agent sessions first; start a fresh session after rollback.", console.Field{Label: "Checkpoint", Value: backups[n-1]}, console.Field{Label: "Restore toolkit", Value: r.Before.Binary})
	if err != nil || !ok {
		return err
	}
	if _, err := runToolkit(r.Before.Binary, "toolkit", "check-instance", "--instance", i.Root); err != nil {
		return errors.New("previous toolkit cannot verify compatibility with this installation; automatic rollback refused, backup preserved")
	}
	if err := i.RollbackToolkit(s, backups[n-1]); err != nil {
		return err
	}
	u.block(console.Block{Title: "Rollback complete", Body: "Previous agent toolkit restored. The My Friday management command was not downgraded.", Tone: console.Success})
	u.section("Next step", "Start a fresh agent session.")
	return nil
}
