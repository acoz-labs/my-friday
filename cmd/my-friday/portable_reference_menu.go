package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/acoz-labs/my-friday/internal/console"
	"github.com/acoz-labs/my-friday/internal/portable"
)

func (u managementUI) references(path string) error {
	u.section("Reference sources", "Link external local directories as reference-only material. Descriptions travel with agent source; directory bindings are machine-local. Reference contents are not cloned, fetched, executed or imported as current memory.")
	for {
		i, s, err := portable.LoadInstance(path)
		if err != nil {
			return err
		}
		libs, err := s.ReferenceLibraries()
		if err != nil {
			return err
		}
		labels := []string{"Register a reference source"}
		for _, lib := range libs {
			labels = append(labels, lib.Title+" ("+lib.ID+")")
		}
		if len(libs) == 0 {
			u.section("No references registered", "Register a description first. You can bind an existing directory now or later on each machine.")
		}
		n, err := u.choose("Reference sources", labels, "Back")
		if err != nil || n == 0 {
			return err
		}
		if n == 1 {
			err = u.addReference(i, s)
		} else {
			err = u.referenceLibrary(path, libs[n-2].ID)
		}
		if err := u.problem(err); err != nil {
			return err
		}
	}
}

func (u managementUI) addReference(i portable.Instance, s *portable.Store) error {
	u.section("Register a reference source", "Use an existing local directory or separately cloned Git repository. Git URLs and automatic updates are not supported here. Keep credentials and machine paths out of the portable description.")
	lib := portable.ReferenceLibrary{Version: 1}
	var err error
	for _, field := range []struct {
		label string
		value *string
	}{{"Library ID (3–128 lowercase letters/digits/hyphens; start with a letter)", &lib.ID}, {"Title", &lib.Title}, {"Description — what it contains", &lib.Description}, {"Purpose — how it should inform future work", &lib.Purpose}} {
		*field.value, err = u.ask(field.label, "")
		if err != nil {
			return err
		}
	}
	ok, err := u.review("Save a portable description, checkpoint source and attempt normal source synchronization. The source must be clean first. Sync may upload complete committed agent history; review it for secrets.", "The external library, local bindings, active memory and capabilities are unchanged. Old instructions do not become current policy.", "Avoid concurrent source edits during registration. No session restart is required for later explicit reference reads.", console.Field{Label: "Library ID", Value: lib.ID}, console.Field{Label: "Title", Value: lib.Title}, console.Field{Label: "Description", Value: lib.Description}, console.Field{Label: "Purpose", Value: lib.Purpose})
	if err != nil || !ok {
		return err
	}
	// Refuse unrelated uncommitted work before the existing add/checkpoint flow.
	if _, err := s.RemoteSetupStatus(context.Background()); err != nil {
		return err
	}
	if err := s.AddReference(lib); err != nil {
		return err
	}
	status, err := s.WithCheckpointObserver(portable.Authorship{DeviceID: i.DeviceID, Actor: s.Agent.Name, Harness: "management"}).Sync(context.Background())
	if err != nil {
		return fmt.Errorf("reference descriptor saved; inspect source/checkpoint state before retrying registration: %w", err)
	}
	tone := console.Success
	if status.State != "synced" && status.State != "local-only" {
		tone = console.Warning
	}
	u.block(console.Block{Title: "Reference registered", Body: "Description saved. No directory is bound and no reference content was loaded.", Tone: tone, Fields: []console.Field{{Label: "Source sync", Value: status.State}}})
	if status.Detail != "" {
		u.section("Synchronization detail", status.Detail)
	}
	u.section("Next step", "Open the registered source and choose Bind/rebind local directory. Local-only means no remote backup; pending or conflict needs source reconciliation, not another registration.")
	return nil
}

func (u managementUI) referenceLibrary(path, id string) error {
	for {
		i, s, err := portable.LoadInstance(path)
		if err != nil {
			return err
		}
		libs, err := s.ReferenceLibraries()
		if err != nil {
			return err
		}
		var lib portable.ReferenceLibrary
		for _, candidate := range libs {
			if candidate.ID == id {
				lib = candidate
				break
			}
		}
		if lib.ID == "" {
			return fmt.Errorf("reference source no longer registered; return to the list")
		}
		n, err := u.choose("Reference: "+lib.Title, []string{"View description", "Bind/rebind local directory", "Check availability"}, "Back")
		if err != nil || n == 0 {
			return err
		}
		switch n {
		case 1:
			u.referenceDescription(lib)
		case 2:
			err = u.bindReference(i, s, lib)
		case 3:
			var status portable.ReferenceAvailability
			status, err = i.CheckReference(s, id)
			if err == nil {
				u.referenceAvailability(status)
			}
		}
		if err := u.problem(err); err != nil {
			return err
		}
	}
}

func (u managementUI) referenceDescription(lib portable.ReferenceLibrary) {
	u.section("Reference description", "Reference-only material, not current instructions, verified facts or an installed capability.", console.Field{Label: "Library ID", Value: lib.ID}, console.Field{Label: "Title", Value: lib.Title}, console.Field{Label: "Description", Value: lib.Description}, console.Field{Label: "Purpose", Value: lib.Purpose})
	u.section("Portability", "This description travels in agent source Git. Bind a separate local directory on each machine. Updating a referenced Git repository is separate from agent-source synchronization.")
}

func (u managementUI) bindReference(i portable.Instance, s *portable.Store, lib portable.ReferenceLibrary) error {
	u.referenceDescription(lib)
	status, err := i.CheckReference(s, lib.ID)
	if err != nil {
		return err
	}
	u.referenceAvailability(status)
	root, err := u.ask("Existing external directory (absolute path)", status.Root)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(root) {
		return fmt.Errorf("choose an absolute local directory, not a Git URL or relative path")
	}
	ok, err := u.review("Save or replace this library's directory binding on this machine, acknowledging its current description.", "External files, source Git, active memory, capabilities and other machines' bindings are untouched. No scripts or old instructions are activated.", "Future explicit reference reads use this binding. No context is automatically imported.", console.Field{Label: "Library ID", Value: lib.ID}, console.Field{Label: "Directory", Value: root})
	if err != nil || !ok {
		return err
	}
	if err := i.BindReference(s, lib.ID, root, lib); err != nil {
		return err
	}
	u.block(console.Block{Title: "Local binding saved", Body: "No source files copied, fetched, executed or promoted.", Tone: console.Success})
	u.section("Next step", "Choose Check availability. Refresh an external Git checkout separately when needed; agent sync does not fetch it.")
	return nil
}

func (u managementUI) referenceAvailability(status portable.ReferenceAvailability) {
	tone := console.Warning
	if status.State == "available" {
		tone = console.Success
	}
	fields := []console.Field{{Label: "State", Value: status.State}}
	if status.Root != "" {
		fields = append(fields, console.Field{Label: "Configured directory", Value: status.Root})
	}
	u.block(console.Block{Title: "Reference availability", Body: status.Detail, Fields: fields, Tone: tone})
	if status.State == "available" {
		u.section("Next step", "The agent can search/read this library as reference-only evidence. Git freshness and document contents were not checked.")
	} else {
		u.section("Next step", "Review the description and directory, then bind/rebind explicitly. Preserve invalid metadata for inspection; missing or stale references do not block normal agent use.")
	}
}
