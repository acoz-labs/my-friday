package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/acoz-labs/my-friday/internal/memorybank"
	"github.com/acoz-labs/my-friday/internal/memorycodex"
	"github.com/acoz-labs/my-friday/internal/memorymcp"
	"github.com/acoz-labs/my-friday/internal/portable"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func memoryBankCLI(args []string, input io.Reader, out, errout io.Writer) error {
	if len(args) == 0 || helpFlag(args[0]) {
		return printPortableHelp("bank", out)
	}
	command := args[0]
	f := portableFlags("bank "+command, errout)
	if command == "create" || command == "bind" {
		root := f.String("repository", "", "Memory-only repository directory")
		label := f.String("device-label", "", "Readable name of this machine")
		if command == "create" {
			name := f.String("name", "", "Memory bank display name")
			if err := parseFlags(f, args[1:]); err != nil {
				return err
			}
			s, err := portable.CreateMemoryBank(*root, *name, portable.NewID("device"), *label)
			if err != nil {
				return err
			}
			if err := s.InitGit(context.Background()); err != nil {
				return err // Preserve created bank for diagnosis; no automatic cleanup.
			}
			return outputJSON(out, map[string]any{"bank_id": s.Agent.ID, "repository": s.Root, "created": true, "synchronization": "local-only"})
		}
		defaultPath, err := memorybank.DefaultBindingPath()
		if err != nil {
			return err
		}
		binding := f.String("binding", defaultPath, "New machine-local binding file outside the bank")
		actor := f.String("actor", "", "Attribution name, not an authentication identity")
		if err := parseFlags(f, args[1:]); err != nil {
			return err
		}
		b, err := memorybank.Bind(*root, *binding, *label, *actor)
		if err != nil {
			return err
		}
		return outputJSON(out, b)
	}
	defaultPath, err := memorybank.DefaultBindingPath()
	if err != nil {
		return err
	}
	binding := f.String("binding", defaultPath, "Machine-local memory binding")
	harness := f.String("harness", "cli", "Authorship harness label")
	var query, scopeKind, scopeID, recordID string
	var limit, offset, budget int
	switch command {
	case "recall":
		f.StringVar(&query, "query", "", "Words or identifiers to recall")
		f.StringVar(&scopeKind, "scope-kind", "", "Explicit stored scope kind")
		f.StringVar(&scopeID, "scope-id", "", "Explicit stored scope ID")
		f.IntVar(&limit, "limit", 5, "Maximum returned hits")
		f.IntVar(&budget, "budget-bytes", 8192, "Compact JSON context budget")
	case "scopes", "history":
		f.IntVar(&limit, "limit", 5, "Maximum page size")
		f.IntVar(&offset, "offset", 0, "Page offset")
		if command == "history" {
			f.StringVar(&recordID, "record", "", "Record ID to inspect")
		}
	case "journal":
		f.StringVar(&query, "query", "", "Search journal summaries")
		f.IntVar(&limit, "limit", 5, "Maximum entries")
	case "remember", "journal-append", "sync", "doctor":
	default:
		return errors.New("usage: my-friday bank <create|bind|doctor|recall|remember|history|scopes|journal|journal-append|sync>")
	}
	if err := parseFlags(f, args[1:]); err != nil {
		return err
	}
	s, err := memorybank.OpenBinding(*binding, *harness)
	if err != nil {
		return err
	}
	var result any
	switch command {
	case "recall":
		var scope *portable.Scope
		if scopeKind != "" || scopeID != "" {
			if scopeKind == "" || scopeID == "" {
				return errors.New("scope-kind and scope-id must be supplied together")
			}
			scope = &portable.Scope{Kind: scopeKind, ID: scopeID}
		}
		result, err = s.Recall(query, scope, limit, budget)
	case "scopes":
		result, err = s.ScopePage(offset, limit)
	case "history":
		result, err = s.HistoryPage(recordID, offset, limit)
	case "journal":
		result, err = s.JournalPage(query, limit)
	case "remember":
		var v memorybank.Write
		if err := memoryInput(input, &v); err != nil {
			return err
		}
		result, err = s.Remember(v)
	case "journal-append":
		var v memorymcp.JournalWrite
		if err := memoryInput(input, &v); err != nil {
			return err
		}
		result, err = s.AppendJournal(v.Kind, v.Summary)
	case "sync":
		result, err = s.Sync(context.Background())
	case "doctor":
		var store *portable.Store
		store, err = portable.Open(s.Root())
		if err == nil {
			err = store.Validate()
		}
		result = map[string]any{"healthy": err == nil, "bank_id": s.ID(), "notice": "Read-only memory structure and binding checks; remote freshness, credentials, native plugin loading, and model behavior are not tested."}
		if err != nil {
			_ = outputJSON(out, result)
		}
	}
	if err != nil {
		return err
	}
	return outputJSON(out, result)
}

func memoryInput(input io.Reader, value any) error {
	b, err := io.ReadAll(io.LimitReader(input, 32769))
	if err != nil {
		return err
	}
	if len(b) > 32768 {
		return errors.New("memory input exceeds 32 KiB")
	}
	// Reuse strict JSON handling: unknown fields and trailing values are errors.
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		return err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return errors.New("memory input contains trailing data")
	}
	return nil
}

type memoryOutput struct{ io.Writer }

func (memoryOutput) Close() error { return nil }

func memoryMCPCLI(args []string, input io.Reader, out, errout io.Writer) error {
	f := portableFlags("mcp", errout)
	defaultPath, err := memorybank.DefaultBindingPath()
	if err != nil {
		return err
	}
	binding := f.String("binding", defaultPath, "Machine-local memory bank binding")
	harness := f.String("harness", "mcp", "Native harness name for provenance")
	if err := parseFlags(f, args); err != nil {
		return err
	}
	s, err := memorybank.OpenBinding(*binding, *harness)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return memorymcp.New(s).Run(ctx, &mcp.IOTransport{Reader: io.NopCloser(input), Writer: memoryOutput{out}})
}

func memoryCodexHookCLI(args []string, input io.Reader, out, errout io.Writer) error {
	f := portableFlags("codex-memory-hook", errout)
	defaultPath, err := memorybank.DefaultBindingPath()
	if err != nil {
		return err
	}
	binding := f.String("binding", defaultPath, "Machine-local memory bank binding")
	if err := parseFlags(f, args); err != nil {
		return err
	}
	return memorycodex.Run(*binding, input, out)
}
