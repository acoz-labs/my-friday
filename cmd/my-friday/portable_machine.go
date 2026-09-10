package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/acoz-labs/my-friday/internal/portable"
)

func portableMachine(args []string, out, errout io.Writer) error {
	if len(args) == 0 {
		return printPortableHelp("machine", out)
	}
	action := args[0]
	if action != "status" && action != "check" && action != "prepare" {
		return errors.New("unknown machine command; use help machine")
	}
	f := portableFlags("machine "+action, errout)
	instance := f.String("instance", os.Getenv("MY_FRIDAY_INSTANCE"), "Absolute machine-local instance directory")
	capability := f.String("capability", "", "Explicit capability ID")
	requirement := f.String("requirement", "", "Explicit requirement ID")
	apply := f.Bool("apply", false, "Execute preparation; otherwise print a no-execution preview")
	expected := f.String("expect-sha256", "", "Reviewed capability fingerprint; required with --apply")
	if err := parseFlags(f, args[1:]); err != nil {
		return err
	}
	if !filepath.IsAbs(*instance) {
		return errors.New("machine commands require an absolute --instance or MY_FRIDAY_INSTANCE")
	}
	if action != "prepare" && (*apply || *expected != "") {
		return errors.New("--apply and --expect-sha256 are only for machine prepare")
	}
	i, s, err := portable.LoadInstance(*instance)
	if err != nil {
		return err
	}
	if action == "status" {
		if *capability != "" || *requirement != "" {
			return errors.New("machine status lists all requirements; omit capability/requirement flags")
		}
		status, err := i.MachineStatus(s)
		if err != nil {
			return err
		}
		return outputJSON(out, map[string]any{"requirements": status, "notice": "Metadata only. Ready is a matching past check, not current authentication or network verification. No scripts run."})
	}
	if *capability == "" || *requirement == "" {
		return errors.New("select both --capability and --requirement; use machine status")
	}
	p, err := i.MachinePlan(s, *capability, *requirement)
	if err != nil {
		return err
	}
	if action == "prepare" {
		if !*apply {
			return outputJSON(out, map[string]any{"state": "preview", "plan": p, "notice": "No scripts run. Review private instructions and commands, then use --apply --expect-sha256 with this capability fingerprint. Preparation checks first, skips an already-satisfied installer, then verifies. Scripts have full user access. No rollback is promised."})
		}
		if *expected == "" || *expected != p.SHA256 {
			return errors.New("preparation requires the current reviewed --expect-sha256; obtain a fresh preview")
		}
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	r, runErr := i.RunMachine(ctx, s, p, action)
	if r.ID != "" {
		if err := outputJSON(out, r); err != nil {
			return err
		}
	}
	if runErr != nil {
		return runErr
	}
	if r.State != "ready" {
		return errors.New("machine requirement needs preparation; inspect the receipt and private instructions")
	}
	return nil
}
