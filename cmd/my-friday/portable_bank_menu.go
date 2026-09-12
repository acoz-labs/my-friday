package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/acoz-labs/my-friday/internal/console"
	"github.com/acoz-labs/my-friday/internal/memorybank"
	"github.com/acoz-labs/my-friday/internal/portable"
)

func memoryMenuMode(home string, input io.Reader, out io.Writer, plain bool) error {
	u := newManagementUI(home, input, out, plain)
	u.line(console.Accent, "My Friday — durable memory")
	u.line(console.Muted, "Bring your agent. Nothing changes just by opening this menu.")
	u.line(console.Muted, "Previous assistant installations: my-friday menu --legacy")
	for {
		n, err := u.choose("Memory banks", []string{"Create a memory bank", "Connect an existing bank", "Check a memory bank", "Synchronize a memory bank"}, "Exit")
		if errors.Is(err, io.EOF) || (err == nil && n == 0) {
			return nil
		}
		if err != nil {
			return err
		}
		switch n {
		case 1, 2:
			err = u.memorySetup(n == 2)
		case 3, 4:
			err = u.memoryInspectOrSync(n == 4)
		}
		if errors.Is(u.problem(err), io.EOF) {
			return nil
		}
	}
}

func (u managementUI) memorySetup(existing bool) error {
	name := ""
	var err error
	if existing {
		u.section("Connect an existing bank", "Choose an already-cloned memory-only repository. This enrolls this machine; it does not copy threads or migrate an older assistant repository.")
	} else {
		name, err = u.ask("Memory bank name", "Personal memory")
		if err != nil {
			return err
		}
	}
	root, err := u.ask("Absolute memory bank directory", filepath.Join(u.home, ".local/share/my-friday/banks/personal"))
	if err != nil {
		return err
	}
	if !filepath.IsAbs(root) {
		return errors.New("use an absolute memory bank directory")
	}
	host, _ := os.Hostname()
	label, err := u.ask("This machine's label", host)
	if err != nil {
		return err
	}
	actor, err := u.ask("Writer name (attribution, not authentication)", os.Getenv("USER"))
	if err != nil {
		return err
	}
	defaultBinding, err := memorybank.DefaultBindingPath()
	if err != nil {
		return err
	}
	binding, err := u.ask("New machine-local binding file", defaultBinding)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(binding) {
		return errors.New("use an absolute binding file path outside the bank")
	}
	change := "Create a memory-only repository with local Git history, then connect this machine. No remote will be configured."
	if existing {
		change = "Add this machine's provenance record to the existing bank and create a local connection file. No synchronization will run."
	}
	approved, err := u.review(change, "Native agent settings, authentication, skills, existing assistant installations and unrelated files.", "The bank is independent of the directory from which you launch your agent. Configure its native memory plugin separately.", console.Field{Label: "Bank", Value: root}, console.Field{Label: "Local binding", Value: binding})
	if err != nil || !approved {
		return err
	}
	// Catch an ordinary collision before creating a bank, not only after the
	// first half of setup. Bind itself still enforces no-replacement publication.
	if err := memorybank.ValidateBindingDestination(root, binding); err != nil {
		return err
	}
	if !existing {
		s, err := portable.CreateMemoryBank(root, name, portable.NewID("device"), label)
		if err != nil {
			return err
		}
		if err := s.InitGit(context.Background()); err != nil {
			return fmt.Errorf("bank created at %s, but Git setup failed; preserve it and diagnose before retrying: %w", s.Root, err)
		}
	}
	b, err := memorybank.Bind(root, binding, label, actor)
	if err != nil {
		if !existing {
			return fmt.Errorf("bank created at %s, but connection failed; use Connect an existing bank after correcting the problem: %w", root, err)
		}
		return err
	}
	u.block(console.Block{Title: "Memory bank connected", Tone: console.Success, Fields: []console.Field{{Label: "Bank", Value: b.Root}, {Label: "Bank ID", Value: b.BankID}, {Label: "Local binding", Value: binding}, {Label: "Device", Value: b.DeviceID}}})
	u.section("Next step", "Connect your native agent's memory plugin. It must use this binding file; non-default files are selected with MY_FRIDAY_MEMORY_BINDING. No agent was installed or launched.")
	return nil
}

func (u managementUI) memoryInspectOrSync(sync bool) error {
	defaultBinding, err := memorybank.DefaultBindingPath()
	if err != nil {
		return err
	}
	binding, err := u.ask("Machine-local binding file", defaultBinding)
	if err != nil {
		return err
	}
	s, err := memorybank.OpenBinding(binding, "menu")
	if err != nil {
		return err
	}
	if sync {
		approved, err := u.review("Checkpoint local memory changes and synchronize the bank's already-configured Git remote, if any.", "Native agent configuration and other banks. No remote or authentication will be configured.", "Offline/pending means locally durable, not confirmed on another machine.", console.Field{Label: "Bank", Value: s.Root()})
		if err != nil || !approved {
			return err
		}
		status, err := s.Sync(context.Background())
		if err != nil {
			return err
		}
		u.section("Memory synchronization", status.Detail, console.Field{Label: "State", Value: status.State}, console.Field{Label: "Local commit", Value: status.Head}, console.Field{Label: "Checked", Value: status.CheckedAt})
		return nil
	}
	store, err := portable.Open(s.Root())
	if err != nil {
		return err
	}
	if err := store.Validate(); err != nil {
		return err
	}
	u.block(console.Block{Title: "Memory structure is healthy", Tone: console.Success, Body: "Read-only bank structure and binding checks passed. This does not verify remote freshness, authentication, native plugin loading, or model behavior.", Fields: []console.Field{{Label: "Bank", Value: s.Root()}, {Label: "Bank ID", Value: s.ID()}}})
	return nil
}
