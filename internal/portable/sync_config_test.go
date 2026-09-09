package portable

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSyncConfigIsOptionalAndProviderNeutral(t *testing.T) {
	s := fixtureStore(t)
	config, err := s.syncConfiguration()
	if err != nil || len(config.CredentialHelper) != 0 || config.Author != nil {
		t.Fatalf("unexpected default provider: %+v %v", config, err)
	}
	if err := writeNewJSON(filepath.Join(s.Root, ".my-friday/sync.json"), SyncConfig{Version: 1, CredentialHelper: []string{"./capabilities/source-auth/helper", "literal argument"}, Author: &GitAuthor{Name: "Example Agent", Email: "agent@example.invalid"}}); err != nil {
		t.Fatal(err)
	}
	if err := s.InitGit(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := gitTest(t, s.Root, "log", "-1", "--format=%an|%ae"); strings.TrimSpace(got) != "Example Agent|agent@example.invalid" {
		t.Fatalf("wrong attribution: %q", got)
	}
}

func TestSyncConfigRejectsMalformedProvider(t *testing.T) {
	for _, data := range []string{`{}`, `{"schema_version":2}`, `{"schema_version":1,"credential_helper":[""]}`, `{"schema_version":1,"author":{"name":"Missing email"}}`, `{"schema_version":1,"account_role":"prescribed-role"}`, `{"schema_version":1,"github_source":{"repository":"https://github.com/o/r"}}`, `{"schema_version":1,"github_source":{"repository":"owner/repo"},"credential_helper":["helper"]}`} {
		s := fixtureStore(t)
		if err := os.WriteFile(filepath.Join(s.Root, ".my-friday/sync.json"), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if err := s.Validate(); err == nil {
			t.Fatalf("invalid sync configuration accepted: %s", data)
		}
	}
}

func TestPrivateGitHelperDoesNotFallBackToAmbientConfiguration(t *testing.T) {
	s := fixtureStore(t)
	if err := s.InitGit(context.Background()); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(t.TempDir(), "ambient-invoked")
	ambient := filepath.Join(t.TempDir(), "ambient-helper")
	if err := os.WriteFile(ambient, []byte("#!/bin/sh\nprintf invoked > "+shellQuote(marker)+"\nprintf 'username=unintended\\npassword=synthetic\\n'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	gitTest(t, s.Root, "config", "credential.helper", "!"+shellQuote(ambient))
	t.Setenv("GIT_ASKPASS", ambient)
	failing := filepath.Join(t.TempDir(), "failing-helper")
	if err := os.WriteFile(failing, []byte("#!/bin/sh\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	s.gitSettings = &SyncConfig{Version: 1, CredentialHelper: []string{failing}}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append(s.gitArguments(), "credential", "fill")...)
	cmd.Env = cleanGitEnvironment()
	cmd.Stdin = strings.NewReader("protocol=https\nhost=code.example.test\n\n")
	if err := cmd.Run(); err == nil {
		t.Fatal("unconfigured credential fallback succeeded")
	}
	if _, err := os.Lstat(marker); !os.IsNotExist(err) {
		t.Fatal("ambient credential helper or askpass executed")
	}
}

func TestFreshAgentContainsNoServiceCapabilities(t *testing.T) {
	s := fixtureStore(t)
	caps, err := s.Capabilities()
	if err != nil || len(caps) != 0 {
		t.Fatalf("fresh agent has service capabilities: %+v %v", caps, err)
	}
	entries, err := os.ReadDir(filepath.Join(s.Root, "integrations"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() != ".gitkeep" {
			t.Fatalf("seeded integration: %s", entry.Name())
		}
	}
	// Service configuration is opaque to core and validated by its capability.
	if err := os.WriteFile(filepath.Join(s.Root, "integrations/example-service.json"), []byte(`{"private_setting":"example"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestPrivateGitHelperUsesStandardProtocolAndLiteralArguments(t *testing.T) {
	s := fixtureStore(t)
	// This is a synthetic helper, not a bundled service integration.
	dir := filepath.Join(s.Root, "extensions", "helper with ' quote")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	helper := filepath.Join(dir, "credentials")
	script := "#!/bin/sh\n[ \"$1\" = 'literal $(never-execute)' ] || exit 1\n[ \"$2\" = get ] || exit 1\nrequest=$( /bin/cat )\ncase \"$request\" in *host=code.example.test*) printf 'username=fixture\\npassword=synthetic-value\\n';; *) exit 1;; esac\n"
	if err := os.WriteFile(helper, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	config := SyncConfig{Version: 1, CredentialHelper: []string{"./extensions/helper with ' quote/credentials", "literal $(never-execute)"}}
	s.gitSettings = &config
	cmd := exec.Command("git", append(s.gitArguments(), "credential", "fill")...)
	cmd.Env = cleanGitEnvironment()
	cmd.Stdin = strings.NewReader("protocol=https\nhost=code.example.test\n\n")
	out, err := cmd.Output()
	if err != nil || !strings.Contains(string(out), "password=synthetic-value") {
		t.Fatal("private credential helper protocol failed")
	}
}

func TestNetworkSyncWithoutProviderStaysLocallyDurable(t *testing.T) {
	s := fixtureStore(t)
	if err := s.InitGit(context.Background()); err != nil {
		t.Fatal(err)
	}
	gitTest(t, s.Root, "remote", "add", "origin", "https://code.example.test/team/agent.git")
	if err := s.Put(revision("revision-pending-provider")); err != nil {
		t.Fatal(err)
	}
	status, err := s.Sync(context.Background())
	if err != nil || status.State != "pending" || !strings.Contains(status.Detail, "credential helper") {
		t.Fatalf("unexpected missing-provider result: %+v %v", status, err)
	}
	if out := gitTest(t, s.Root, "status", "--porcelain"); out != "" {
		t.Fatal("local changes not committed")
	}
}
