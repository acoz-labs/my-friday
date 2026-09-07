package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/acoz-labs/my-friday/internal/portable"
)

func portableFlags(name string, out io.Writer) *flag.FlagSet {
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.SetOutput(out)
	return f
}
func outputJSON(out io.Writer, value any) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
func parseFlags(f *flag.FlagSet, args []string) error {
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(f.Args(), " "))
	}
	return nil
}

func runPortable(args []string, input io.Reader, out, errout io.Writer) (err error) {
	defer func() {
		if errors.Is(err, flag.ErrHelp) {
			err = nil
		}
	}()
	if len(args) == 0 {
		args = []string{"setup"}
	}
	if len(args) == 1 && helpFlag(args[0]) {
		return printPortableHelp("", out)
	}
	if args[0] == "help" {
		topic := strings.Join(args[1:], " ")
		if _, ok := portableHelpTopics[topic]; ok {
			return printPortableHelp(topic, out)
		}
		// Delegate option help to the command's own flag definitions.
		if (len(args) == 2 && (args[1] == "setup" || args[1] == "sync" || args[1] == "hook")) ||
			(len(args) == 3 && (args[1] == "agent" || args[1] == "memory")) {
			return runPortable(append(append([]string{}, args[1:]...), "--help"), input, out, out)
		}
		return printPortableHelp(topic, out)
	}
	if (args[0] == "agent" || args[0] == "memory") && (len(args) == 1 || (len(args) == 2 && helpFlag(args[1]))) {
		return printPortableHelp(args[0], out)
	}
	if (len(args) == 2 || len(args) == 3) && helpFlag(args[len(args)-1]) {
		errout = out
	}
	switch args[0] {
	case "setup":
		return portableSetup(args[1:], input, out, errout)
	case "memory":
		return portableMemory(args[1:], input, out, errout)
	case "sync":
		f := portableFlags("sync", errout)
		root := f.String("repository", os.Getenv("MY_FRIDAY_ASSISTANT_ROOT"), "Assistant repository")
		if err := parseFlags(f, args[1:]); err != nil {
			return err
		}
		s, err := portable.Open(*root)
		if err != nil {
			return err
		}
		status, err := s.Sync(context.Background())
		if err != nil {
			return err
		}
		return outputJSON(out, status)
	case "agent":
		return portableAgent(args[1:], input, out, errout)
	case "hook":
		return portableHook(args[1:], input, out, errout)
	default:
		return errors.New("usage: my-friday <setup|agent|memory|sync|hook>")
	}
}

func portableSetup(args []string, input io.Reader, out, errout io.Writer) error {
	f := portableFlags("setup", errout)
	root := f.String("repository", "", "New assistant repository directory")
	existing := f.String("import", "", "Existing local assistant repository")
	state := f.String("state", "", "Local instance directory")
	name := f.String("name", "", "Launcher name")
	harness := f.String("harness", "codex", "Default harness for new assistants")
	label := f.String("device-label", "", "Readable machine label")
	noLauncher := f.Bool("no-launcher", false, "Prepare an instance without installing its named launcher")
	launcher := f.String("launcher", "", "Launcher path")
	if err := parseFlags(f, args); err != nil {
		return err
	}
	home, err := realHome()
	if err != nil {
		return err
	}
	reader := bufio.NewReader(input)
	ask := func(prompt, def string) (string, error) {
		fmt.Fprintf(out, "%s [%s]: ", prompt, def)
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", errors.New("setup input ended; supply explicit flags for noninteractive setup")
		}
		line = strings.TrimSpace(line)
		if line == "" {
			line = def
		}
		return line, nil
	}
	if *root == "" && *existing == "" {
		mode, err := ask("Create a new agent or import an existing local repository (new/import)", "new")
		if err != nil {
			return err
		}
		if mode == "import" {
			*existing, err = ask("Existing assistant repository", "")
			if err != nil {
				return err
			}
		} else if mode != "new" {
			return errors.New("choose new or import")
		}
	}
	if *name == "" {
		*name, err = ask("Agent launcher name", "friday")
		if err != nil {
			return err
		}
	}
	if err := portable.ValidateLauncherName(*name); err != nil {
		return err
	}
	if *root == "" && *existing == "" {
		*root, err = ask("Assistant repository", filepath.Join(home, ".local/share/my-friday/repositories", *name))
		if err != nil {
			return err
		}
		*harness, err = ask("Default harness (codex/pi)", *harness)
		if err != nil {
			return err
		}
	}
	if *label == "" {
		host, _ := os.Hostname()
		*label, err = ask("Machine label", host)
		if err != nil {
			return err
		}
	}
	if *state == "" {
		*state = filepath.Join(home, ".local/share/my-friday/instances", *name)
	}
	if *launcher == "" {
		*launcher = filepath.Join(home, ".local/bin", *name)
	}
	// Check all obvious installation collisions before creating source state.
	if _, err = os.Lstat(*state); !os.IsNotExist(err) {
		return errors.New("instance directory already exists or cannot be inspected")
	}
	if !*noLauncher {
		if _, err = os.Lstat(*launcher); !os.IsNotExist(err) {
			return errors.New("launcher already exists or cannot be inspected")
		}
	}
	device := portable.NewID("device")
	var s *portable.Store
	if *existing != "" {
		if *root != "" {
			return errors.New("choose --repository or --import")
		}
		s, err = portable.Open(*existing)
		if err != nil {
			return err
		}
		if err = s.Validate(); err != nil {
			return err
		}
		if err = s.AddDevice(portable.Device{Version: 1, ID: device, Label: *label}); err != nil {
			return err
		}
	} else {
		s, err = portable.Create(*root, *name, *harness, device, *label)
		if err != nil {
			return err
		}
	}
	if err = s.InitGit(context.Background()); err != nil {
		return err
	}
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	binary, err = filepath.EvalSymlinks(binary)
	if err != nil {
		return err
	}
	instance, err := portable.Bind(s, *state, *name, binary, device)
	if err != nil {
		return err
	}
	if !*noLauncher {
		if err = instance.InstallLauncher(*launcher); err != nil {
			return err
		}
	}
	return outputJSON(out, map[string]any{"assistant_id": s.Agent.ID, "repository": s.Root, "instance": instance.Root, "device_id": device, "launcher_installed": !*noLauncher, "default_harness": s.Agent.DefaultHarness, "next": "Authenticate each harness in its private instance home before first use."})
}

func portableMemory(args []string, input io.Reader, out, errout io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: my-friday memory <template|write|recall|scopes|history|source|event>")
	}
	f := portableFlags("memory "+args[0], errout)
	root := f.String("repository", os.Getenv("MY_FRIDAY_ASSISTANT_ROOT"), "Assistant repository")
	device := f.String("device", os.Getenv("MY_FRIDAY_DEVICE_ID"), "Registered originating device")
	query := f.String("query", "", "Recall query")
	scopeKind := f.String("scope-kind", "assistant", "assistant, project, account, or task")
	scopeID := f.String("scope-id", "", "Scoped entity ID")
	record := f.String("record", "", "Record ID")
	file := f.String("input", "", "Revision document, or - for stdin")
	summary := f.String("summary", "", "Concise source or journal summary")
	kind := f.String("kind", "user-direction", "Source or journal kind")
	if err := parseFlags(f, args[1:]); err != nil {
		return err
	}
	if *root == "" {
		return errors.New("assistant repository required via --repository or named launcher")
	}
	s, err := portable.Open(*root)
	if err != nil {
		return err
	}
	if *scopeID == "" && *scopeKind == "assistant" {
		*scopeID = s.Agent.ID
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	author := portable.Authorship{DeviceID: *device, Actor: s.Agent.Name, Harness: os.Getenv("MY_FRIDAY_HARNESS")}
	if author.Harness == "" {
		author.Harness = "cli"
	}
	if value := os.Getenv("MY_FRIDAY_SESSION_ID"); value != "" {
		author.SessionID = &value
	}
	if value := os.Getenv("PI_MODEL"); value != "" {
		author.Model = &value
	}
	switch args[0] {
	case "scopes":
		scopes, err := s.Scopes()
		if err != nil {
			return err
		}
		return outputJSON(out, scopes)
	case "template":
		return outputJSON(out, portable.Revision{Version: 1, ID: portable.NewID("revision"), RecordID: portable.NewID("record"), Kind: "preference", Scope: portable.Scope{Kind: *scopeKind, ID: *scopeID}, Summary: "Describe the enduring claim", Body: "Explain its meaning and boundaries", Sensitivity: "private", Volatility: "stable", RecordedAt: now, EffectiveFrom: now, Authorship: author, Evidence: portable.Evidence{Basis: "user-direction", Confidence: "high", SourceRefs: []string{}}, Supersedes: []string{}, ChangeReason: "Describe why this claim was created or changed"})
	case "recall":
		packet, err := s.Recall(portable.Query{Text: *query, Scope: portable.Scope{Kind: *scopeKind, ID: *scopeID}}, time.Now())
		if err != nil {
			return err
		}
		return outputJSON(out, packet)
	case "history":
		history, err := s.History(*record)
		if err != nil {
			return err
		}
		return outputJSON(out, history)
	case "write":
		var reader io.Reader = input
		if *file == "" {
			return errors.New("--input is required")
		}
		if *file != "-" {
			r, err := os.Open(*file)
			if err != nil {
				return err
			}
			defer r.Close()
			reader = r
		}
		decoder := json.NewDecoder(io.LimitReader(reader, 4<<20))
		decoder.DisallowUnknownFields()
		var r portable.Revision
		if err = decoder.Decode(&r); err != nil {
			return err
		}
		var extra any
		if err = decoder.Decode(&extra); err != io.EOF {
			return errors.New("trailing revision JSON")
		}
		r.Authorship = author
		r.RecordedAt = now
		if r.EffectiveFrom == "" {
			r.EffectiveFrom = now
		}
		if err = s.Put(r); err != nil {
			return err
		}
		status, err := s.Sync(context.Background())
		if err != nil {
			return fmt.Errorf("revision %s saved; checkpoint needs recovery: %w", r.ID, err)
		}
		return outputJSON(out, map[string]any{"revision_id": r.ID, "sync": status})
	case "source":
		source := portable.Source{Version: 1, ID: portable.NewID("source"), Kind: *kind, Summary: *summary, DeviceID: *device, RecordedAt: now}
		if err = s.AddSource(source); err != nil {
			return err
		}
		status, err := s.Sync(context.Background())
		if err != nil {
			return err
		}
		return outputJSON(out, map[string]any{"source": source, "sync": status})
	case "event":
		entry, err := s.RecordEvent(*kind, *summary, author)
		if err != nil {
			return err
		}
		status, err := s.Sync(context.Background())
		if err != nil {
			return err
		}
		return outputJSON(out, map[string]any{"event": entry, "sync": status})
	default:
		return errors.New("unknown memory command")
	}
}

// Only reserve launcher options; all other arguments belong to the harness.
// A literal -- ends My Friday parsing and is preserved for the harness.
func splitLaunchArgs(args []string) (owned, forwarded []string, err error) {
	for n := 0; n < len(args); n++ {
		arg := args[n]
		if arg == "--" {
			forwarded = append(forwarded, args[n:]...)
			break
		}
		key, _, equals := strings.Cut(arg, "=")
		if key != "--instance" && key != "--harness" {
			forwarded = append(forwarded, arg)
			continue
		}
		owned = append(owned, arg)
		if !equals {
			n++
			if n == len(args) || strings.HasPrefix(args[n], "--") {
				return nil, nil, fmt.Errorf("%s requires a value", key)
			}
			owned = append(owned, args[n])
		}
	}
	return owned, forwarded, nil
}

func portableAgent(args []string, input io.Reader, out, errout io.Writer) error {
	if len(args) == 0 {
		return printPortableHelp("agent", out)
	}
	switch args[0] {
	case "launch", "inspect", "validate", "capabilities", "check", "capability-guide", "capability-template":
	default:
		return errors.New("unknown agent command; use my-friday help agent")
	}
	f := portableFlags("agent "+args[0], errout)
	root := f.String("repository", os.Getenv("MY_FRIDAY_ASSISTANT_ROOT"), "Assistant repository")
	state := f.String("instance", os.Getenv("MY_FRIDAY_INSTANCE"), "Instance directory")
	harness := f.String("harness", "", "Override default harness")
	capabilityID := f.String("capability", "", "Capability ID")
	description := f.String("description", "", "Capability template description")
	parseArgs := args[1:]
	var forwarded []string
	if args[0] == "launch" {
		var err error
		parseArgs, forwarded, err = splitLaunchArgs(parseArgs)
		if err != nil {
			return err
		}
	}
	if err := f.Parse(parseArgs); err != nil {
		return err
	}
	if args[0] == "launch" {
		instance, s, err := portable.LoadInstance(*state)
		if err != nil {
			return err
		}
		status, err := s.Sync(context.Background())
		if err != nil {
			return err
		}
		if status.State != "synced" && status.State != "local-only" {
			fmt.Fprintf(errout, "My Friday: %s — %s\n", status.State, status.Detail)
		}
		// The pull may have changed the default harness or agent display name.
		instance, s, err = portable.LoadInstance(*state)
		if err != nil {
			return err
		}
		if err = instance.Project(s); err != nil {
			return err
		}
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		plan, err := instance.Plan(s, *harness, cwd, forwarded)
		if err != nil {
			return err
		}
		cmd := exec.Command(plan.Executable, plan.Arguments...)
		cmd.Dir = plan.Directory
		cmd.Env = plan.Environment
		cmd.Stdin = input
		cmd.Stdout = out
		cmd.Stderr = errout
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("%s session exited: %w", plan.Executable, err)
		}
		return nil
	}
	if f.NArg() != 0 {
		return errors.New("unexpected agent arguments")
	}
	if args[0] == "capability-guide" {
		_, err := io.WriteString(out, portable.CapabilityGuide)
		return err
	}
	if args[0] == "capability-template" {
		template, err := portable.CapabilityTemplate(*capabilityID, *description)
		if err != nil {
			return err
		}
		return outputJSON(out, template)
	}
	if args[0] == "check" && *capabilityID == "" {
		return errors.New("agent check requires --capability ID; use agent capabilities to list IDs or agent capability-guide to design one")
	}
	s, err := portable.Open(*root)
	if err != nil {
		return err
	}
	switch args[0] {
	case "inspect":
		return outputJSON(out, s.Agent)
	case "validate":
		if err = s.Validate(); err != nil {
			return err
		}
		return outputJSON(out, map[string]bool{"valid": true})
	case "capabilities":
		caps, err := s.DiscoverCapabilities()
		if err != nil {
			return err
		}
		return outputJSON(out, caps)
	case "check":
		results, err := s.CheckCapability(*capabilityID)
		if err != nil {
			return err
		}
		return outputJSON(out, results)
	default:
		return errors.New("unknown agent command")
	}
}

func portableHook(args []string, input io.Reader, out, errout io.Writer) error {
	f := portableFlags("hook", errout)
	state := f.String("instance", "", "Instance directory")
	harness := f.String("harness", "", "Harness")
	native := f.String("native", "", "Native lifecycle event")
	if err := parseFlags(f, args); err != nil {
		return err
	}
	instance, s, err := portable.LoadInstance(*state)
	if err != nil {
		return err
	}
	if *harness != "codex" && *harness != "pi" {
		return errors.New("invalid hook harness")
	}
	var payload map[string]json.RawMessage
	decoder := json.NewDecoder(io.LimitReader(input, 1<<20))
	if err = decoder.Decode(&payload); err != nil && err != io.EOF {
		return err
	}
	field := func(name string) string { var value string; _ = json.Unmarshal(payload[name], &value); return value }
	eventID := field("event_id")
	if eventID == "" {
		eventID = portable.NewID("event")
	}
	raw, _ := json.Marshal(payload)
	event := portable.Event{Version: 1, ID: eventID, Name: portable.NormalizeEvent(*harness, *native), AssistantID: s.Agent.ID, DeviceID: instance.DeviceID, SessionID: field("session_id"), RequestID: field("turn_id"), NativeEvent: *native, WorkingDirectory: field("cwd"), Payload: raw}
	contextParts := []string{}
	synchronize := func() {
		status, syncErr := s.Sync(context.Background())
		if syncErr != nil {
			contextParts = append(contextParts, "My Friday synchronization needs attention; local memory is preserved.")
		} else if status.State != "synced" && status.State != "local-only" {
			contextParts = append(contextParts, "My Friday sync status: "+status.State+". "+status.Detail)
		}
	}
	if event.Name == "session.started" || event.Name == "request.received" {
		synchronize()
	}
	if event.Name == "request.received" {
		packet, err := s.Recall(portable.Query{Text: field("prompt"), Scope: portable.Scope{Kind: "assistant", ID: s.Agent.ID}}, time.Now())
		if err != nil {
			return err
		}
		data, _ := json.Marshal(packet)
		contextParts = append(contextParts, string(data))
	}
	dispatched, dispatchErr := s.Dispatch(context.Background(), event)
	contextParts = append(contextParts, dispatched.Context...)
	if dispatchErr != nil {
		contextParts = append(contextParts, "My Friday subscription requires attention: "+dispatchErr.Error())
	}
	if event.Name == "request.completed" {
		// Include changes made by completion subscribers in this checkpoint.
		synchronize()
	}
	text := strings.Join(contextParts, "\n")
	if *harness == "pi" {
		response := map[string]any{"additional_context": text}
		if dispatchErr != nil {
			response["error"] = dispatchErr.Error()
		}
		return outputJSON(out, response)
	}
	response := map[string]any{}
	switch *native {
	case "SessionStart", "UserPromptSubmit", "PreToolUse", "PostToolUse", "PreCompact", "PostCompact", "SubagentStart":
		if text != "" {
			response["hookSpecificOutput"] = map[string]string{"hookEventName": *native, "additionalContext": text}
		}
	default:
		if text != "" {
			response["systemMessage"] = text
		}
	}
	return outputJSON(out, response)
}
