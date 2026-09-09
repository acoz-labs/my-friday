package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/acoz-labs/my-friday/internal/portable"
	"github.com/acoz-labs/my-friday/internal/toolkitupdate"
)

// One bounded request per invocation: no terminal prompts, shell evaluation,
// background daemon or hidden memory of an earlier approval.
type apiRequest struct {
	SchemaVersion int                        `json:"schema_version"`
	ID            string                     `json:"id,omitempty"`
	Action        string                     `json:"action"`
	Apply         bool                       `json:"apply,omitempty"`
	Params        map[string]json.RawMessage `json:"params,omitempty"`
}
type apiProblem struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}
type apiResponse struct {
	SchemaVersion        int         `json:"schema_version"`
	ID                   string      `json:"id,omitempty"`
	Action               string      `json:"action,omitempty"`
	OK                   bool        `json:"ok"`
	State                string      `json:"state"`
	Result               any         `json:"result,omitempty"`
	Error                *apiProblem `json:"error,omitempty"`
	Effects              []string    `json:"effects,omitempty"`
	RequiresFreshSession bool        `json:"requires_fresh_session,omitempty"`
	SideEffectsPossible  bool        `json:"side_effects_may_have_occurred,omitempty"`
}

// main must not append human error prose after an already-written API envelope.
type apiExit struct {
	status int
	code   string
}

func (e apiExit) Error() string { return e.code }

type apiParameter struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}
type apiSpec struct {
	Action                  string                  `json:"action"`
	Description             string                  `json:"description"`
	Mutates                 bool                    `json:"mutates"`
	Network                 bool                    `json:"network"`
	RequiresStoppedSessions bool                    `json:"requires_stopped_sessions"`
	Parameters              map[string]apiParameter `json:"parameters"`
	Required                []string                `json:"required"`
	Effects                 []string                `json:"effects,omitempty"`
	ParameterSchema         map[string]any          `json:"parameter_schema"`
}

func apiSpecs() []apiSpec {
	makeSpec := func(action, description, params, required string, mutates, network, stopped bool, effects ...string) apiSpec {
		p := map[string]apiParameter{}
		for _, name := range strings.Fields(params) {
			kind := "string"
			if name == "no_launcher" || name == "create" || name == "sessions_stopped" || name == "activate_command" {
				kind = "boolean"
			}
			p[name] = apiParameter{Type: kind}
		}
		properties := map[string]any{}
		for name, param := range p {
			property := map[string]any{"type": param.Type}
			if param.Type == "string" {
				property["minLength"] = 1
				property["maxLength"] = 4096
			}
			if name == "harness" {
				property["enum"] = []string{"codex", "pi"}
			}
			if name == "sha256" {
				property["pattern"] = "^[a-f0-9]{64}$"
			}
			properties[name] = property
		}
		return apiSpec{Action: action, Description: description, Parameters: p, Required: strings.Fields(required), Mutates: mutates, Network: network, RequiresStoppedSessions: stopped, Effects: effects, ParameterSchema: map[string]any{"type": "object", "properties": properties, "required": strings.Fields(required), "additionalProperties": false}}
	}
	return []apiSpec{
		makeSpec("system.describe", "Discover this API; request IDs correlate responses, not deduplicate actions.", "", "", false, false, false),
		makeSpec("system.version", "Inspect the running executable's build and compatibility.", "", "", false, false, false),
		makeSpec("agent.list", "Discover local instances, including damaged entries. Defaults to the standard instances directory.", "directory", "", false, false, false),
		makeSpec("agent.inspect", "Inspect an explicitly selected agent; no provider authentication or sync.", "instance", "instance", false, false, false),
		makeSpec("agent.doctor", "Read-only structural checks, not authentication/network verification.", "instance harness", "instance", false, false, false),
		makeSpec("agent.setup", "Create or import local source; exactly one of repository/import. Set no_launcher=true or supply launcher. Requires a machine label; no credential or reference migration.", "instance name repository import harness device_label launcher no_launcher", "instance name device_label", true, false, false, "Create source or register a device in imported source; checkpoint; create instance and optional launcher. Native logins remain separate."),
		makeSpec("agent.repair", "Refresh managed files; optional launcher creates only a missing launcher.", "instance launcher", "instance", true, false, false, "Regenerate managed instructions/hooks. Existing native settings, credentials, sessions and source are preserved; start a fresh session."),
		makeSpec("agent.harness", "Change the portable default; refuse unrelated dirty source and checkpoint locally.", "instance harness", "instance harness", true, false, false, "Change agent.json and record a local source checkpoint; propagate on later sync, not into a running conversation."),
		makeSpec("source.status", "Inspect clean-source setup configuration without authentication or network checks.", "instance", "instance", false, false, false),
		makeSpec("source.accounts", "List stored source-hosting accounts; not proof of valid credentials.", "", "", false, true, false),
		makeSpec("source.configure", "Configure source hosting: github requires target/setup_account/sync_account; existing requires remote and a configured helper. Explicit accounts, no active-account switch.", "instance mode target setup_account sync_account create remote", "instance mode", true, true, false, "May create a private remote, save source settings, checkpoint, attach origin and upload complete committed source history. Created repos/settings are retained on later failure."),
		makeSpec("source.sync", "Checkpoint and synchronize the selected source. Local-only/pending/conflict are explicit states, not remote success.", "instance", "instance", true, true, false, "Checkpoint ALL nonignored source, fetch, reconcile and push; never place secrets/unrelated work in source. Offline work is retained."),
		makeSpec("toolkit.latest", "Check the official latest release; refuse incompatible legacy releases.", "", "", false, true, false),
		makeSpec("toolkit.install", "Install a trusted local artifact by expected digest; optionally activate the management command. Agent pins stay unchanged.", "binary sha256 activate_command", "binary sha256", true, false, false, "Stage immutable artifact and execute its compatibility check. Optionally replace only the managed convenience symlink, retaining the previous pointer."),
		makeSpec("toolkit.install-release", "Install an explicitly selected version/digest only if it is still the eligible latest release. No automatic branch/prerelease selection.", "version sha256 activate_command", "version sha256", true, true, false, "Download and verify the approved artifact; execute compatibility check; optionally activate management command. Agent pins remain unchanged."),
		makeSpec("toolkit.use", "Adopt the RUNNING executable for one agent; launcher optional, otherwise unchanged. Active self-upgrade is refused.", "instance launcher sessions_stopped", "instance", true, false, true, "Back up and replace only instance binding, managed files and the explicitly selected managed launcher. Native data and source are preserved."),
		makeSpec("toolkit.backups", "List valid local adoption receipts, not native credentials or sessions.", "instance", "instance", false, false, false),
		makeSpec("toolkit.rollback", "Verify previous toolkit against today's source and restore only unchanged managed files from a selected receipt.", "instance backup sessions_stopped", "instance backup", true, false, true, "Restore managed files/pin; refuse later edits. Memory and external actions are never rolled back."),
	}
}

func strictJSON(data []byte, target any) error {
	if err := uniqueJSON(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("expected one JSON value")
	}
	return nil
}

// encoding/json otherwise silently accepts duplicate fields, including apply.
func uniqueJSON(data []byte) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	var value func(int) error
	value = func(depth int) error {
		if depth > 64 {
			return errors.New("JSON nesting limit exceeded")
		}
		token, err := d.Token()
		if err != nil {
			return err
		}
		switch token {
		case json.Delim('{'):
			seen := map[string]bool{}
			for d.More() {
				key, err := d.Token()
				if err != nil {
					return err
				}
				name, ok := key.(string)
				if !ok || seen[name] {
					return errors.New("duplicate JSON field")
				}
				seen[name] = true
				if err := value(depth + 1); err != nil {
					return err
				}
			}
			_, err = d.Token()
			return err
		case json.Delim('['):
			for d.More() {
				if err := value(depth + 1); err != nil {
					return err
				}
			}
			_, err = d.Token()
			return err
		}
		return nil
	}
	if err := value(0); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return errors.New("expected one JSON value")
	}
	return nil
}
func apiString(r apiRequest, name string) string {
	var value string
	_ = json.Unmarshal(r.Params[name], &value)
	return value
}
func apiBool(r apiRequest, name string) bool {
	var value bool
	_ = json.Unmarshal(r.Params[name], &value)
	return value
}

func portableAPI(args []string, input io.Reader, out io.Writer) error {
	response := apiResponse{SchemaVersion: 1}
	finish := func(code int) error {
		if err := outputJSON(out, response); err != nil {
			return err
		}
		if code != 0 {
			return apiExit{code, response.Error.Code}
		}
		return nil
	}
	fail := func(code, message string) error {
		response.State = "invalid_request"
		response.Error = &apiProblem{Code: code, Message: message}
		return finish(2)
	}
	var r apiRequest
	if len(args) == 1 && args[0] == "describe" {
		r = apiRequest{SchemaVersion: 1, Action: "system.describe"}
	} else {
		if len(args) != 2 || args[0] != "--input" {
			return fail("input.invalid", "Use api describe or api --input FILE (use - for stdin).")
		}
		var reader io.Reader = input
		if args[1] != "-" {
			f, err := os.Open(args[1])
			if err != nil {
				return fail("input.unreadable", "Cannot open request file.")
			}
			defer f.Close()
			reader = f
		}
		data, err := io.ReadAll(io.LimitReader(reader, (1<<20)+1))
		if err != nil || len(data) > 1<<20 {
			return fail("input.invalid", "Request must be readable and at most 1 MiB.")
		}
		if strictJSON(data, &r) != nil {
			return fail("input.invalid", "Expected one strict JSON request; unknown fields and trailing values are not accepted.")
		}
		var fields map[string]json.RawMessage
		_ = json.Unmarshal(data, &fields)
		allowed := map[string]bool{"schema_version": true, "id": true, "action": true, "apply": true, "params": true}
		for name, raw := range fields {
			if !allowed[name] {
				return fail("input.invalid", "Request field names must match the documented spelling exactly.")
			}
			if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				return fail("input.invalid", "Request fields cannot be null; omit optional fields instead.")
			}
		}
	}
	if r.SchemaVersion != 1 || len(r.ID) > 128 || strings.ContainsAny(r.ID, "\r\n\x00") {
		return fail("input.version", "Use schema_version 1 and an optional single-line ID of at most 128 bytes.")
	}
	response.ID = r.ID
	response.Action = r.Action
	var spec *apiSpec
	for _, item := range apiSpecs() {
		if item.Action == r.Action {
			copy := item
			spec = &copy
			break
		}
	}
	if spec == nil {
		return fail("action.unknown", "Unknown action; use api describe. Credential getter endpoints are not part of this API.")
	}
	if err := validateAPIParams(r, *spec); err != nil {
		return fail("params.invalid", err.Error())
	}
	if !spec.Mutates && r.Apply {
		return fail("params.invalid", "Read-only actions do not accept apply=true.")
	}
	response.Effects = spec.Effects
	if spec.Mutates && !r.Apply {
		response.OK = true
		response.State = "planned"
		response.Result = map[string]any{"params": r.Params, "notice": "Preview only: parameter validation, not permission, authentication, full preflight or reserved state. Re-read state before applying; no operation was executed."}
		return finish(0)
	}
	if spec.RequiresStoppedSessions {
		if !apiBool(r, "sessions_stopped") {
			return fail("sessions.active", "Close affected agent sessions and explicitly set sessions_stopped=true before applying this operation.")
		}
		active := os.Getenv("MY_FRIDAY_INSTANCE")
		target := apiString(r, "instance")
		if active != "" {
			a, _ := filepath.EvalSymlinks(active)
			b, _ := filepath.EvalSymlinks(target)
			if a != "" && a == b {
				return fail("sessions.self_update", "This invocation belongs to the affected active agent. Stage/inspect now; perform adoption or rollback from a separate management session after it exits.")
			}
		}
	}
	home, err := realHome()
	if err != nil {
		return fail("environment.home", "Cannot resolve this machine's user home.")
	}
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(signalCtx, 60*time.Second)
	defer cancel()
	result, state, err := executeAPI(ctx, home, r)
	response.Result = result
	response.State = state
	if err != nil {
		response.SideEffectsPossible = spec.Mutates
		response.Error = &apiProblem{Code: "operation.failed", Message: err.Error(), Retryable: false}
		if state == "needs_attention" {
			response.Error.Code = "installation.unhealthy"
		}
		if state == "pending" {
			response.Error.Code = "sync.pending"
			response.Error.Retryable = true
		}
		if state == "conflict" {
			response.Error.Code = "sync.conflict"
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			response.Error.Code = "operation.interrupted"
			response.Error.Message = "Operation stopped; inspect target state before retrying."
		}
		if response.State == "" || response.State == "completed" {
			response.State = "failed"
		}
		return finish(3)
	}
	response.OK = true
	if response.State == "" {
		response.State = "completed"
	}
	response.RequiresFreshSession = r.Action == "agent.repair" || r.Action == "toolkit.use" || r.Action == "toolkit.rollback"
	return finish(0)
}

func validateAPIParams(r apiRequest, spec apiSpec) error {
	for name, raw := range r.Params {
		parameter, ok := spec.Parameters[name]
		if !ok {
			return errors.New("Unknown action parameter; consult api describe.")
		}
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return errors.New("Parameters must have their declared type, not null.")
		}
		if parameter.Type == "boolean" {
			var value bool
			if json.Unmarshal(raw, &value) != nil {
				return errors.New("Expected a boolean parameter.")
			}
		} else {
			var value string
			if json.Unmarshal(raw, &value) != nil || strings.TrimSpace(value) == "" || len(value) > 4096 || strings.ContainsAny(value, "\r\n\x00") {
				return errors.New("Expected a nonempty single-line string parameter (at most 4096 bytes).")
			}
			switch name {
			case "instance", "repository", "import", "launcher", "directory", "binary", "backup":
				if !filepath.IsAbs(value) {
					return errors.New("Filesystem parameters must be absolute paths; cwd is never a target default.")
				}
			}
		}
	}
	for _, name := range spec.Required {
		if _, ok := r.Params[name]; !ok {
			return fmt.Errorf("Missing required parameter: %s.", name)
		}
	}
	if harness := apiString(r, "harness"); harness != "" && harness != "codex" && harness != "pi" {
		return errors.New("Harness must be codex or pi.")
	}
	if hash := apiString(r, "sha256"); hash != "" {
		if _, err := hex.DecodeString(hash); err != nil || len(hash) != 64 || hash != strings.ToLower(hash) {
			return errors.New("SHA-256 must be exactly 64 lowercase hexadecimal characters from the trusted artifact record.")
		}
	}
	if r.Action == "agent.setup" {
		if apiString(r, "import") != "" && apiString(r, "harness") != "" {
			return errors.New("Imported source retains its default harness; use agent.harness afterward to change it.")
		}
		if (apiString(r, "repository") == "") == (apiString(r, "import") == "") {
			return errors.New("Setup requires exactly one of repository or import.")
		}
		if apiBool(r, "no_launcher") == (apiString(r, "launcher") != "") {
			return errors.New("Choose no_launcher=true or an explicit launcher, not both.")
		}
		if err := portable.ValidateLauncherName(apiString(r, "name")); err != nil {
			return err
		}
	}
	if r.Action == "source.configure" {
		mode := apiString(r, "mode")
		if mode != "github" && mode != "existing" {
			return errors.New("Source mode must be github or existing.")
		}
		if mode == "github" {
			if !portable.ValidGitHubRepository(apiString(r, "target")) || apiString(r, "setup_account") == "" || apiString(r, "sync_account") == "" || apiString(r, "remote") != "" {
				return errors.New("GitHub mode requires target owner/repo, setup_account and sync_account, with no remote override.")
			}
		} else if !portable.ValidSourceRemote(apiString(r, "remote")) || apiString(r, "target") != "" || apiString(r, "setup_account") != "" || apiString(r, "sync_account") != "" || apiBool(r, "create") {
			return errors.New("Existing mode requires only a credential-free HTTPS or absolute local remote; no provider/account overrides.")
		}
	}
	return nil
}

func executeAPI(ctx context.Context, home string, r apiRequest) (any, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "failed", err
	}
	switch r.Action {
	case "system.describe":
		return map[string]any{"schema_version": 1, "actions": apiSpecs(), "request": map[string]any{"schema_version": 1, "id": "optional correlation ID, not an idempotency key", "action": "action name", "params": "typed object; unknown fields rejected", "apply": "mutations default to preview; true explicitly executes"}, "execution": "One JSON request/response per invocation; never drive the TUI. Nonzero exit means inspect error and result. Failure can follow partial work. No automatic retries.", "other_tooling": "Memory, reference and capability operations retain their existing JSON CLI. Use my-friday help memory / reference / agent; no general command or credential proxy is exposed."}, "completed", nil
	case "system.version":
		return toolkitVersion(), "completed", nil
	case "agent.list":
		directory := apiString(r, "directory")
		if directory == "" {
			directory = filepath.Join(home, ".local/share/my-friday/instances")
		}
		items, err := portable.DiscoverInstances(directory)
		return map[string]any{"directory": directory, "instances": items}, "completed", err
	case "agent.setup":
		args := []string{"--state", apiString(r, "instance"), "--name", apiString(r, "name"), "--device-label", apiString(r, "device_label")}
		for _, name := range []string{"repository", "import", "harness", "launcher"} {
			if value := apiString(r, name); value != "" {
				args = append(args, "--"+name, value)
			}
		}
		if apiBool(r, "no_launcher") {
			args = append(args, "--no-launcher")
		}
		var result bytes.Buffer
		if err := portableSetupContext(ctx, args, strings.NewReader(""), &result, io.Discard); err != nil {
			return nil, "failed", err
		}
		var data any
		if err := json.Unmarshal(result.Bytes(), &data); err != nil {
			return nil, "failed", errors.New("Setup completed but its result could not be decoded; inspect the explicit target before retrying.")
		}
		return data, "completed", nil
	case "source.accounts":
		items, err := (portable.GitHubSourceClient{}).Accounts(ctx)
		return map[string]any{"accounts": items}, "completed", err
	case "toolkit.latest":
		a, err := (toolkitupdate.Client{}).Latest(ctx)
		return map[string]any{"version": a.Version, "sha256": a.SHA256, "os": a.OS, "arch": a.Arch, "artifact": a.Name}, "completed", err
	case "toolkit.install", "toolkit.install-release":
		var path string
		var err error
		if r.Action == "toolkit.install" {
			path, err = toolkitupdate.Stage(home, apiString(r, "binary"), apiString(r, "sha256"))
		} else {
			client := toolkitupdate.Client{}
			a, e := client.Latest(ctx)
			if e != nil {
				return nil, "failed", e
			}
			if a.Version != apiString(r, "version") || a.SHA256 != apiString(r, "sha256") {
				return nil, "failed", errors.New("Eligible release differs from the explicitly approved version/digest; inspect again before applying.")
			}
			path, err = client.Download(ctx, home, a)
		}
		if err != nil {
			return nil, "failed", err
		}
		if err := verifyToolkitContext(ctx, path); err != nil {
			return map[string]any{"staged_binary": path}, "failed", err
		}
		backup := ""
		if apiBool(r, "activate_command") {
			if err := ctx.Err(); err != nil {
				return map[string]any{"staged_binary": path}, "failed", err
			}
			current, e := os.Executable()
			if e != nil {
				return nil, "failed", e
			}
			backup, err = toolkitupdate.Activate(home, path, current)
		}
		return map[string]any{"binary": path, "command_activated": apiBool(r, "activate_command") && err == nil, "previous_command_checkpoint": backup, "notice": "Agent pins unchanged. Use the selected executable in a new management process for adoption."}, "completed", err
	}
	i, s, err := portable.LoadInstance(apiString(r, "instance"))
	if err != nil {
		return nil, "failed", err
	}
	s = s.WithCheckpointObserver(portable.Authorship{DeviceID: i.DeviceID, Actor: s.Agent.Name, Harness: "management-api"})
	switch r.Action {
	case "agent.inspect":
		return map[string]any{"instance": i, "instance_path": i.Root, "agent": s.Agent, "memory": filepath.Join(s.Root, "memory")}, "completed", nil
	case "agent.doctor":
		report := i.Doctor(s, apiString(r, "harness"))
		if !report.Healthy {
			return report, "needs_attention", errInstallationNeedsAttention
		}
		return report, "completed", nil
	case "agent.repair":
		var result bytes.Buffer
		args := []string{"repair", "--instance", i.Root}
		if launcher := apiString(r, "launcher"); launcher != "" {
			args = append(args, "--launcher", launcher)
		}
		err := portableAgent(args, strings.NewReader(""), &result, io.Discard)
		if err != nil {
			return nil, "failed", err
		}
		var data any
		_ = json.Unmarshal(result.Bytes(), &data)
		return data, "completed", nil
	case "agent.harness":
		err := s.SetDefaultHarness(ctx, apiString(r, "harness"))
		return map[string]any{"default_harness": apiString(r, "harness"), "notice": "Local checkpoint only; later sync propagates this default. Active sessions do not switch harnesses."}, "completed", err
	case "source.status":
		state, err := s.RemoteSetupStatus(ctx)
		return state, "completed", err
	case "source.configure":
		current, err := os.Executable()
		if err != nil {
			return nil, "failed", err
		}
		current, _ = filepath.EvalSymlinks(current)
		bound, _ := filepath.EvalSymlinks(i.Binary)
		if current != bound {
			return nil, "failed", errors.New("Source configuration requires the instance's bound toolkit; adopt a compatible version first.")
		}
		result, err := s.SetupSource(ctx, portable.SourceSetupOptions{Mode: apiString(r, "mode"), Repository: apiString(r, "target"), SetupAccount: apiString(r, "setup_account"), SyncAccount: apiString(r, "sync_account"), Create: apiBool(r, "create"), Remote: apiString(r, "remote")})
		if err != nil {
			return result, "failed", err
		}
		return apiSyncResult(result)
	case "source.sync":
		result, err := s.Sync(ctx)
		if err != nil {
			return result, "failed", err
		}
		return apiSyncResult(result)
	case "toolkit.use":
		binary, err := os.Executable()
		if err != nil {
			return nil, "failed", err
		}
		backup, err := i.UseToolkit(s, binary, apiString(r, "launcher"))
		return map[string]any{"backup": backup, "binary": binary}, "completed", err
	case "toolkit.backups":
		directory := filepath.Join(i.Root, "toolkit-updates")
		entries, err := os.ReadDir(directory)
		if os.IsNotExist(err) {
			return []any{}, "completed", nil
		}
		if err != nil {
			return nil, "failed", err
		}
		results := []any{}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			path := filepath.Join(directory, entry.Name())
			receipt, err := i.ReadToolkitReceipt(path)
			if err == nil {
				results = append(results, map[string]any{"backup": path, "previous_binary": receipt.Before.Binary, "selected_binary": receipt.After.Binary})
			}
		}
		return results, "completed", nil
	case "toolkit.rollback":
		receipt, err := i.ReadToolkitReceipt(apiString(r, "backup"))
		if err != nil {
			return nil, "failed", err
		}
		if _, err := runToolkitContext(ctx, receipt.Before.Binary, "toolkit", "check-instance", "--instance", i.Root); err != nil {
			return nil, "failed", errors.New("Previous toolkit cannot verify current source compatibility; rollback refused, backup preserved.")
		}
		if err := ctx.Err(); err != nil {
			return nil, "failed", err
		}
		err = i.RollbackToolkit(s, apiString(r, "backup"))
		return map[string]any{"binary": receipt.Before.Binary}, "completed", err
	}
	return nil, "failed", errors.New("Action implementation unavailable.")
}
func apiSyncResult(result portable.SyncStatus) (any, string, error) {
	switch result.State {
	case "synced":
		return result, "synced", nil
	case "local-only":
		return result, "local_only", nil
	case "pending":
		return result, "pending", errors.New("Remote synchronization pending; local work preserved. Inspect status before retrying.")
	default:
		return result, result.State, errors.New("Synchronization needs reconciliation; no remote success is claimed.")
	}
}
