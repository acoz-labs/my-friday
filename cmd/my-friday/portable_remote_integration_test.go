package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Real compiled CLI, credential-helper subprocess, Git commits and a local bare
// remote. Only gh/API and network transport are replaced with synthetic fixtures.
func TestCompiledSourceWizardAndCredentialRoundTrip(t *testing.T) {
	base := t.TempDir()
	binary := filepath.Join(base, "my-friday")
	bin := filepath.Join(base, "bin")
	source := filepath.Join(base, "agent")
	state := filepath.Join(base, "instance")
	remote := filepath.Join(base, "remote.git")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, output)
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	bare := exec.Command(realGit, "init", "--bare", "--initial-branch=main", "--template=", remote)
	if output, err := bare.CombinedOutput(); err != nil {
		t.Fatalf("bare: %v %s", err, output)
	}
	gh := `#!/bin/sh
set -eu
test -z "${GITHUB_TOKEN:-}"
test -z "${GH_DEBUG:-}"
if test "$1" = auth; then
  test -z "${GH_TOKEN:-}"
  case "$2" in
    status) printf '%s\n' '{"hosts":{"github.com":[{"login":"fixture-user"},{"login":"fixture-agent"}]}}' ;;
    token)
      case "$6" in
        fixture-user) printf '%s\n' synthetic-token ;;
        fixture-agent) printf '%s\n' synthetic-agent-token ;;
        *) exit 2 ;;
      esac ;;
    *) exit 2 ;;
  esac
  exit 0
fi
case "${GH_TOKEN:-}" in synthetic-token|synthetic-agent-token) ;; *) exit 2 ;; esac
case "$5" in
  user)
    if test "$GH_TOKEN" = synthetic-token; then
      printf 'HTTP/2.0 200 OK\n\n%s\n' '{"login":"fixture-user"}'
    else
      printf 'HTTP/2.0 200 OK\n\n%s\n' '{"login":"fixture-agent"}'
    fi ;;
  user/repos)
    test "$GH_TOKEN" = synthetic-token
    test "$6" = --method; test "$7" = POST
    touch "$FRIDAY_FIXTURE_CREATED"
    printf 'HTTP/2.0 201 Created\n\n{}\n' ;;
  repos/fixture-user/assistant)
    if test ! -e "$FRIDAY_FIXTURE_CREATED"; then printf 'HTTP/2.0 404 Not Found\n\n{}\n'; exit 1; fi
    printf 'HTTP/2.0 200 OK\n\n%s\n' '{"full_name":"fixture-user/assistant","private":true,"permissions":{"push":true}}' ;;
  *) exit 2 ;;
esac
`
	git := `#!/bin/sh
set -eu
operation=
for arg do
  case "$arg" in ls-remote|fetch|push) operation="$arg"; break ;; esac
done
if test -z "$operation"; then exec "$FRIDAY_FIXTURE_GIT" "$@"; fi
credential=$(printf 'protocol=https\nhost=github.com\npath=fixture-user/assistant.git\n\n' | "$FRIDAY_FIXTURE_BINARY" source-credential --repository "$FRIDAY_FIXTURE_SOURCE" get)
expected=$(printf 'username=fixture-agent\npassword=synthetic-agent-token')
test "$credential" = "$expected"
printf '%s\n' "$operation" >> "$FRIDAY_FIXTURE_TRANSPORT"
case "$operation" in
 ls-remote) exec "$FRIDAY_FIXTURE_GIT" ls-remote --refs "$FRIDAY_FIXTURE_REMOTE" ;;
 fetch) exec "$FRIDAY_FIXTURE_GIT" -C "$FRIDAY_FIXTURE_SOURCE" fetch --no-tags "$FRIDAY_FIXTURE_REMOTE" '+refs/heads/*:refs/remotes/origin/*' ;;
 push) exec "$FRIDAY_FIXTURE_GIT" -C "$FRIDAY_FIXTURE_SOURCE" push "$FRIDAY_FIXTURE_REMOTE" HEAD:refs/heads/main ;;
esac
`
	for name, data := range map[string]string{"gh": gh, "git": git} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte(data), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	for key, value := range map[string]string{"FRIDAY_FIXTURE_CREATED": filepath.Join(base, "created"), "FRIDAY_FIXTURE_GIT": realGit, "FRIDAY_FIXTURE_BINARY": binary, "FRIDAY_FIXTURE_SOURCE": source, "FRIDAY_FIXTURE_REMOTE": remote, "FRIDAY_FIXTURE_TRANSPORT": filepath.Join(base, "transport"), "GH_TOKEN": "ambient-token", "GITHUB_TOKEN": "ambient-other", "GH_DEBUG": "api", "GIT_CONFIG_GLOBAL": "/dev/null", "GIT_CONFIG_NOSYSTEM": "1"} {
		t.Setenv(key, value)
	}
	run := func(input string, args ...string) (string, error) {
		cmd := exec.Command(binary, args...)
		cmd.Stdin = strings.NewReader(input)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	if out, err := run("", "setup", "--repository", source, "--state", state, "--name", "assistant", "--device-label", "Fixture", "--no-launcher"); err != nil {
		t.Fatalf("setup %v %s", err, out)
	}
	binding, _ := os.ReadFile(filepath.Join(state, "binding.json"))
	native := filepath.Join(state, "codex", "auth.json")
	os.WriteFile(native, []byte("synthetic-login"), 0600)
	// Decline must not even create the remote.
	if _, err := run("github\nfixture-user\nfixture-user/assistant\ncreate\nfixture-user\nno\n", "setup", "--instance", state); err == nil {
		t.Fatal("unapproved setup succeeded")
	}
	if _, err := os.Stat(filepath.Join(base, "created")); !os.IsNotExist(err) {
		t.Fatal("decline created a repository")
	}
	input := "github\nfixture-user\nfixture-user/assistant\ncreate\nfixture-agent\nfixture-user/assistant\n"
	out, err := run(input, "setup", "--instance", state)
	if err != nil {
		t.Fatalf("wizard %v %s", err, out)
	}
	if !strings.Contains(out, `"state": "synced"`) || strings.Contains(out, "synthetic-token") || strings.Contains(out, "synthetic-agent-token") || strings.Contains(out, "ambient-token") {
		t.Fatal("wizard did not verify sync or leaked credentials", out)
	}
	after, _ := os.ReadFile(filepath.Join(state, "binding.json"))
	auth, _ := os.ReadFile(native)
	if !bytes.Equal(binding, after) || string(auth) != "synthetic-login" {
		t.Fatal("native state changed")
	}
	if out, err := run("", "sync", "--repository", source); err != nil || !strings.Contains(out, `"state": "synced"`) {
		t.Fatalf("future sync %v %s", err, out)
	}
	// Resume an already connected repo with the same selected account.
	input = strings.Replace(input, "create\n", "connect\n", 1)
	if out, err := run(input, "setup", "--instance", state); err != nil {
		t.Fatalf("resume %v %s", err, out)
	}
	inspect := exec.Command(realGit, "-C", source, "ls-files", ".my-friday/local")
	tracked, _ := inspect.Output()
	if len(tracked) != 0 {
		t.Fatal("account binding uploaded")
	}
	transport, _ := os.ReadFile(filepath.Join(base, "transport"))
	if !strings.Contains(string(transport), "fetch") || !strings.Contains(string(transport), "push") {
		t.Fatal("transport not exercised")
	}
}
