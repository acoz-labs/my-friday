package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/portable"
)

func checkContextFixture(t *testing.T, s *portable.Store, instance, device string) string {
	t.Helper()
	dir := filepath.Join(s.Root, "capabilities", "context-probe")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(t.TempDir(), "executed")
	manifest, err := json.Marshal(map[string]any{
		"schema_version": 1, "id": "context-probe", "description": "Synthetic context check", "subscriptions": []any{},
		"checks": [][]string{{"sh", "probe.sh", s.Root, instance, device, binary, marker}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{
		"capability.json": manifest,
		"instructions.md": []byte("Synthetic check only."),
		"probe.sh": []byte(`set -eu
echo executed > "$5"
echo private-output-canary
echo private-error-canary >&2
test "$MY_FRIDAY_ASSISTANT_ROOT" = "$1"
test "${MY_FRIDAY_INSTANCE-}" = "$2"
test "$MY_FRIDAY_DEVICE_ID" = "$3"
test "$MY_FRIDAY_BIN" = "$4"
test "$MY_FRIDAY_ASSISTANT_ID" = "$MY_FRIDAY_DISPATCH_ACTIVE"
test -z "${MY_FRIDAY_FORGED-}"
test -z "${MY_FRIDAY_MACHINE_STATE-}"
test "$(pwd -P)" != "$1/capabilities/context-probe"
test -f probe.sh
`),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return marker
}

func TestAgentCheckInstanceContext(t *testing.T) {
	for _, mode := range []string{"explicit", "instance-only", "environment", "override-environment", "matching-repository-alias", "source-only", "explicit-unbound"} {
		t.Run(mode, func(t *testing.T) {
			i, s := apiFixture(t)
			t.Setenv("MY_FRIDAY_ASSISTANT_ROOT", s.Root)
			t.Setenv("MY_FRIDAY_INSTANCE", "")
			t.Setenv("MY_FRIDAY_DEVICE_ID", "forged-device")
			t.Setenv("MY_FRIDAY_BIN", "forged-binary")
			t.Setenv("MY_FRIDAY_FORGED", "must-not-survive")
			t.Setenv("MY_FRIDAY_MACHINE_STATE", "must-not-survive")
			args := []string{"agent", "check", "--capability", "context-probe"}
			instance, device := i.Root, i.DeviceID
			switch mode {
			case "explicit":
				args = append(args, "--instance", i.Root, "--repository", s.Root)
			case "instance-only":
				t.Setenv("MY_FRIDAY_ASSISTANT_ROOT", "")
				args = append(args, "--instance", i.Root)
			case "environment":
				t.Setenv("MY_FRIDAY_INSTANCE", i.Root)
			case "override-environment":
				t.Setenv("MY_FRIDAY_INSTANCE", filepath.Join(t.TempDir(), "stale instance"))
				t.Setenv("MY_FRIDAY_ASSISTANT_ROOT", filepath.Join(t.TempDir(), "stale source"))
				args = append(args, "--instance", i.Root)
			case "matching-repository-alias":
				alias := filepath.Join(t.TempDir(), "source alias")
				if err := os.Symlink(filepath.Dir(s.Root), alias); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--instance", i.Root, "--repository", filepath.Join(alias, filepath.Base(s.Root)))
			case "source-only":
				instance, device = "", ""
			case "explicit-unbound":
				t.Setenv("MY_FRIDAY_INSTANCE", i.Root)
				args = append(args, "--instance", "", "--repository", s.Root)
				instance, device = "", ""
			}
			marker := checkContextFixture(t, s, instance, device)
			t.Chdir(t.TempDir())
			var out, errout bytes.Buffer
			if err := runPortable(args, strings.NewReader(""), &out, &errout); err != nil {
				t.Fatalf("check failed: %v; %s; %s", err, &out, &errout)
			}
			if strings.Contains(out.String()+errout.String(), "canary") || errout.Len() != 0 {
				t.Fatal("raw check output escaped")
			}
			var results []portable.HandlerResult
			if err := json.Unmarshal(out.Bytes(), &results); err != nil || len(results) != 2 || !results[1].Success {
				t.Fatalf("unexpected results: %s; %v", &out, err)
			}
			if _, err := os.Stat(marker); err != nil {
				t.Fatal("check never executed")
			}
		})
	}
}

func TestAgentCheckRejectsInvalidInstanceContextBeforeExecution(t *testing.T) {
	for _, mode := range []string{"missing", "invalid", "wrong-identity", "wrong-device", "explicit-source-mismatch", "ambient-source-mismatch", "repository-does-not-override-instance"} {
		t.Run(mode, func(t *testing.T) {
			i, s := apiFixture(t)
			t.Setenv("MY_FRIDAY_ASSISTANT_ROOT", s.Root)
			t.Setenv("MY_FRIDAY_INSTANCE", i.Root)
			marker := checkContextFixture(t, s, i.Root, i.DeviceID)
			args := []string{"agent", "check", "--capability", "context-probe"}
			switch mode {
			case "missing":
				args = append(args, "--instance", filepath.Join(t.TempDir(), "missing"))
			case "invalid", "wrong-identity", "wrong-device":
				binding := i
				switch mode {
				case "invalid":
					binding.Version = 999
				case "wrong-identity":
					binding.AssistantID = "assistant-other"
				case "wrong-device":
					binding.DeviceID = "device-other"
				}
				data, _ := json.Marshal(binding)
				if err := os.WriteFile(filepath.Join(i.Root, "binding.json"), data, 0600); err != nil {
					t.Fatal(err)
				}
			default:
				other, otherStore := apiFixture(t)
				switch mode {
				case "explicit-source-mismatch":
					args = append(args, "--instance", other.Root, "--repository", s.Root)
				case "ambient-source-mismatch":
					t.Setenv("MY_FRIDAY_ASSISTANT_ROOT", otherStore.Root)
				case "repository-does-not-override-instance":
					t.Setenv("MY_FRIDAY_INSTANCE", other.Root)
					args = append(args, "--repository", s.Root)
				}
			}
			var out, errout bytes.Buffer
			err := runPortable(args, strings.NewReader(""), &out, &errout)
			if err == nil {
				t.Fatal("invalid context accepted")
			}
			if (strings.Contains(mode, "mismatch") || mode == "repository-does-not-override-instance") && !strings.Contains(err.Error(), "mismatch") {
				t.Fatalf("did not reject the mismatched context: %v", err)
			}
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatal("check executed before binding validation")
			}
		})
	}
}

func TestAgentCheckBoundFailureSuppressesRawOutput(t *testing.T) {
	i, s := apiFixture(t)
	checkContextFixture(t, s, i.Root, i.DeviceID)
	path := filepath.Join(s.Root, "capabilities/context-probe/probe.sh")
	if err := os.WriteFile(path, []byte("echo private-output-canary\necho private-error-canary >&2\nexit 20\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var out, errout bytes.Buffer
	err := runPortable([]string{"agent", "check", "--instance", i.Root, "--capability", "context-probe"}, strings.NewReader(""), &out, &errout)
	if err == nil {
		t.Fatal("failed check claimed success")
	}
	if strings.Contains(out.String()+errout.String()+err.Error(), "canary") {
		t.Fatal("raw failure output escaped")
	}
	if _, err := os.Stat(filepath.Join(i.Root, "machine")); !os.IsNotExist(err) {
		t.Fatal("capability check created machine state")
	}
}
