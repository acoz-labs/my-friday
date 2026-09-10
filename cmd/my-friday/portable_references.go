package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/acoz-labs/my-friday/internal/portable"
)

func portableReference(args []string, out, errout io.Writer) error {
	if len(args) == 0 {
		return printPortableHelp("reference", out)
	}
	switch args[0] {
	case "add", "list", "bind", "status", "search", "read":
	default:
		return errors.New("unknown reference command; use my-friday help reference")
	}
	f := portableFlags("reference "+args[0], errout)
	repository := f.String("repository", os.Getenv("MY_FRIDAY_ASSISTANT_ROOT"), "Assistant repository for add/list; must match instance if supplied")
	state := f.String("instance", os.Getenv("MY_FRIDAY_INSTANCE"), "Local instance; required for bind/status/search/read")
	device := f.String("device", os.Getenv("MY_FRIDAY_DEVICE_ID"), "Checkpoint device for unbound add; instance binding takes precedence")
	id := f.String("library", "", "Reference library ID")
	title := f.String("title", "", "Human-readable library title")
	description := f.String("description", "", "What the library contains; no machine paths or secrets")
	purpose := f.String("purpose", "", "Intended reference use, not an instruction to follow")
	path := f.String("path", "", "Bind: external directory; read: exact relative file path")
	query := f.String("query", "", "Lexical search words; empty lists eligible files")
	limit := f.Int("limit", 20, "Maximum search results (1-100)")
	hash := f.String("sha256", "", "Read only if content still matches this search-result hash")
	if err := parseFlags(f, args[1:]); err != nil {
		return err
	}
	var s *portable.Store
	var instance portable.Instance
	var err error
	if *state != "" {
		instance, s, err = portable.LoadInstance(*state)
		if err != nil {
			return err
		}
		if *repository != "" {
			root, err := filepath.EvalSymlinks(*repository)
			if err != nil {
				return err
			}
			root, err = filepath.Abs(root)
			if err != nil || root != s.Root {
				return errors.New("reference repository does not match the instance")
			}
		}
	} else {
		if args[0] != "add" && args[0] != "list" {
			return errors.New("reference operation requires --instance or a named agent session; bind paths separately on each machine")
		}
		s, err = portable.Open(*repository)
		if err != nil {
			return err
		}
	}
	switch args[0] {
	case "status":
		status, err := instance.CheckReference(s, *id)
		if err != nil {
			return err
		}
		return outputJSON(out, status)
	case "add":
		lib := portable.ReferenceLibrary{Version: 1, ID: *id, Title: *title, Description: *description, Purpose: *purpose}
		if err := s.AddReference(lib); err != nil {
			return err
		}
		if *state != "" {
			*device = instance.DeviceID
		}
		if *device != "" {
			s = s.WithCheckpointObserver(portableAuthorship(*device, s.Agent.Name))
		}
		status, err := s.Sync(context.Background())
		if err != nil {
			return fmt.Errorf("reference descriptor saved; checkpoint needs recovery: %w", err)
		}
		return outputJSON(out, map[string]any{"library": lib, "sync": status, "notice": "Descriptor registered; no content loaded or path bound. Use reference bind on each instance."})
	case "list":
		libs, err := s.ReferenceLibraries()
		if err != nil {
			return err
		}
		return outputJSON(out, map[string]any{"usage": "reference-only", "libraries": libs, "notice": portable.ReferenceNotice + " This lists portable descriptions, not verified local availability. Bind selected libraries separately on this instance."})
	case "bind":
		if err := instance.BindReference(s, *id, *path); err != nil {
			return err
		}
		return outputJSON(out, map[string]any{"bound": true, "library_id": *id, "instance": instance.Root, "device_id": instance.DeviceID, "notice": "Machine-local binding saved; no source files copied, fetched, executed or promoted."})
	case "search":
		result, err := instance.SearchReference(s, *id, *query, *limit)
		if err != nil {
			return err
		}
		return outputJSON(out, result)
	case "read":
		result, err := instance.ReadReference(s, *id, *path, *hash)
		if err != nil {
			return err
		}
		return outputJSON(out, result)
	}
	return nil
}
