package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/acoz-labs/my-friday/internal/console"
	"github.com/acoz-labs/my-friday/internal/portable"
)

func (u managementUI) machine(path string) error {
	u.section("Prepare this machine", "Requirements and scripts belong to private capabilities. Opening this view runs no scripts. Readiness is a past observation, not a live authentication test. Toolkit updates never run preparation automatically.")
	for {
		i, s, err := portable.LoadInstance(path)
		if err != nil {
			return err
		}
		status, err := i.MachineStatus(s)
		if err != nil {
			return err
		}
		labels := []string{}
		for _, st := range status {
			labels = append(labels, st.CapabilityID+" / "+st.Requirement.ID+" — "+st.State)
		}
		if len(status) == 0 {
			u.section("No machine requirements", "Ask your agent to use agent capability-guide when designing a capability that needs local tools, services or credentials. Nothing is installed by this view.")
		}
		n, err := u.choose("Machine requirements", labels, "Back")
		if err != nil || n == 0 {
			return err
		}
		p := status[n-1].MachinePlan
		u.section("Requirement", p.Requirement.Description, console.Field{Label: "Capability", Value: p.CapabilityID}, console.Field{Label: "Requirement", Value: p.Requirement.ID}, console.Field{Label: "Source fingerprint", Value: p.SHA256}, console.Field{Label: "Local state", Value: p.StateDirectory})
		u.section("Before running", "Read this capability's instructions and scripts. Check/verify must avoid external changes; they still run with your full access and write a local receipt. Preparation may install or configure software. No script output is retained. Secret enrollment, if needed, uses the capability's separate private setup helper; never enter secrets in this menu or an agent transcript.")
		n, err = u.choose("Requirement action", []string{"Check readiness", "Prepare / reconcile this requirement"}, "Back")
		if err != nil {
			return err
		}
		if n == 0 {
			continue
		}
		action := "check"
		change := "Run this private requirement's check and, if already prepared, verification. Save machine-local receipts; no source synchronization."
		if n == 2 {
			action = "prepare"
			change = "Check this requirement, run its private preparation script only when the check requests it, then verify. External effects can survive failure. Preparation is noninteractive; complete any private secret-enrollment step first."
		}
		ok, err := u.review(change, "My Friday does not rewrite source, store credential values, or change native harness settings. Private scripts must preserve existing credentials and unrelated configuration.", "No agent session is launched. New PATH or service settings may require a fresh terminal/session.")
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		u.section("Running machine "+action, "Waiting for the selected private scripts. Ctrl+C cancels; external effects are not undone.")
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		r, runErr := i.RunMachine(ctx, s, p, action)
		cancel()
		if r.ID != "" {
			tone := console.Warning
			if r.State == "ready" {
				tone = console.Success
			}
			u.block(console.Block{Title: "Machine result", Body: r.State, Tone: tone, Fields: []console.Field{{Label: "Receipt", Value: r.ID}, {Label: "Observed at", Value: r.CheckedAt}, {Label: "Phases", Value: fmt.Sprint(r.Phases)}}})
		}
		if err := u.problem(runErr); err != nil {
			return err
		}
	}
}
