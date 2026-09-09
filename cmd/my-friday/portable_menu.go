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
}

func (u managementUI) ask(label, def string) (string, error) {
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
	fmt.Fprintln(u.out, summary)
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
		fmt.Fprintf(u.out, "\nNeeds attention: %s\nYou can correct the issue and retry from this menu.\n", err)
	}
	return nil
}

func managementMenu(home string, input io.Reader, out io.Writer) error {
	u := managementUI{reader: promptReader(input), out: out, home: home}
	fmt.Fprintln(out, "My Friday — agent management\nNothing changes just by opening this menu. At a prompt, :back cancels that step.")
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

func (u managementUI) setup(importing bool) error {
	title := "New agent"
	if importing {
		title = "Import an existing agent"
	}
	fmt.Fprintf(u.out, "\n%s\n", title)
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
	ok, err := u.confirm(fmt.Sprintf("Agent: %s\nSource: %s\nInstance: %s\nCommand: %s\nExisting paths will not be overwritten.", name, source, instance, launcher))
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
	fmt.Fprintf(u.out, "\n%s is installed. Its native harness login is separate; no credentials were copied.\nConfigure repository synchronization below when ready.\n", name)
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
		n, err := u.choose(i.Name, []string{"View status", "Configure repository and synchronization", "Change default harness", "Check installation health", "Repair installation", "Use this toolkit version for this agent", "Roll back the last toolkit change"}, "Back")
		if err != nil || n == 0 {
			return err
		}
		switch n {
		case 1:
			fmt.Fprintf(u.out, "\nAgent: %s\nDefault harness: %s\nSource: %s\nPersistent memory: %s\nInstance: %s\nPinned toolkit: %s\n", i.Name, s.Agent.DefaultHarness, s.Root, filepath.Join(s.Root, "memory"), i.Root, i.Binary)
			state, e := s.RemoteSetupStatus(context.Background())
			if e != nil {
				fmt.Fprintf(u.out, "Source setup status needs attention: %s\n", e)
			} else if state.Origin == "" {
				fmt.Fprintln(u.out, "Source backup: local only (no remote).")
			} else {
				fmt.Fprintf(u.out, "Configured remote: %s\nThis is configuration, not proof of a successful recent sync.\n", state.Origin)
			}
		case 2:
			err = remoteSetupWizardUI(u.reader, u.out, i, s, true)
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
				ok, err = u.confirm("This changes the portable default on all machines after synchronization. Existing sessions are unaffected; native login remains separate.")
				if err == nil && ok {
					err = s.WithCheckpointObserver(portable.Authorship{DeviceID: i.DeviceID, Actor: s.Agent.Name, Harness: "management"}).SetDefaultHarness(context.Background(), harness)
					if err == nil {
						fmt.Fprintln(u.out, "Default harness saved and checkpointed locally. It will synchronize at the next normal sync.")
					}
				}
			}
		case 4:
			u.doctor(i, s)
		case 5:
			var ok bool
			ok, err = u.confirm("Close active agent sessions first. Repair refreshes managed instructions/hooks only; memory, credentials, native settings and sessions are preserved. It does not fix network/authentication failures.")
			if err == nil && ok {
				err = s.Validate()
				if err == nil {
					err = i.Project(s)
				}
				if err == nil {
					fmt.Fprintln(u.out, "Managed files refreshed. Start a fresh agent session.")
					launcher := filepath.Join(u.home, ".local/bin", i.Name)
					if _, missing := os.Lstat(launcher); os.IsNotExist(missing) {
						create, askErr := u.confirm("The default launcher is missing at " + launcher + ". Create it? Custom launchers elsewhere will not be changed.")
						if askErr != nil {
							err = askErr
						} else if create {
							err = i.InstallLauncher(launcher)
						}
					}
					u.doctor(i, s)
				}
			}
		case 6:
			err = u.adopt(i, s)
		case 7:
			err = u.rollback(i, s)
		}
		if err := u.problem(err); err != nil {
			return err
		}
	}
}

func (u managementUI) doctor(i portable.Instance, s *portable.Store) {
	r := i.Doctor(s, "")
	passed := 0
	for _, check := range r.Checks {
		if check.OK {
			passed++
		}
	}
	fmt.Fprintf(u.out, "\nInstallation health: %d of %d local checks passed.\n", passed, len(r.Checks))
	for _, check := range r.Checks {
		if check.OK {
			continue
		}
		fmt.Fprintf(u.out, "  Needs attention — %s: %s\n", check.Name, check.Detail)
		if check.Remedy != "" {
			fmt.Fprintf(u.out, "    Next: %s\n", check.Remedy)
		}
	}
	if r.Healthy {
		fmt.Fprintln(u.out, "Local structural checks passed.")
	} else {
		fmt.Fprintln(u.out, "Repair installation can refresh managed instructions/hooks; other findings need the remedy shown above.")
	}
	current, _ := os.Executable()
	current, _ = filepath.EvalSymlinks(current)
	if current != i.Binary {
		fmt.Fprintln(u.out, "This compares against the running toolkit; the agent is pinned to another executable. Use this toolkit version to adopt it explicitly.")
	}
	fmt.Fprintln(u.out, "Authentication and remote synchronization were not tested. Doctor is read-only.")
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
	ok, err := u.confirm(fmt.Sprintf("Close active %s sessions first.\nPrevious toolkit: %s\nSelected toolkit: %s\nBack up and update this agent's binding, managed files and selected launcher. Source and native settings/logins stay unchanged.", i.Name, i.Binary, binary))
	if err != nil || !ok {
		return err
	}
	backup, err := i.UseToolkit(s, binary, launcher)
	if err != nil {
		return err
	}
	fmt.Fprintf(u.out, "Agent toolkit updated. Rollback files: %s\nStart a fresh agent session. Other agents retain their existing versions.\n", backup)
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
	ok, err := u.confirm("Close active agent sessions. Verify the retained executable against today's source, then restore only unchanged managed files from this checkpoint. Memory/source and native credentials are never rolled back.")
	if err != nil || !ok {
		return err
	}
	if _, err := runToolkit(r.Before.Binary, "toolkit", "check-instance", "--instance", i.Root); err != nil {
		return errors.New("previous toolkit cannot verify compatibility with this installation; automatic rollback refused, backup preserved")
	}
	if err := i.RollbackToolkit(s, backups[n-1]); err != nil {
		return err
	}
	fmt.Fprintln(u.out, "Previous agent toolkit restored. Start a fresh session. The My Friday management command was not downgraded.")
	return nil
}
