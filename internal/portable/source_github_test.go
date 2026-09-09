package portable

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func sourceResponse(status int, body string) []byte {
	return []byte(fmtStatus(status) + "\r\nContent-Type: application/json\r\n\r\n" + body)
}
func fmtStatus(status int) string {
	if status == 404 {
		return "HTTP/2.0 404 Not Found"
	}
	if status == 201 {
		return "HTTP/2.0 201 Created"
	}
	return "HTTP/2.0 200 OK"
}
func repoPacket(name string, private bool) string {
	data, _ := json.Marshal(map[string]any{"full_name": name, "private": private, "permissions": map[string]bool{"push": true}})
	return string(data)
}

func TestGitHubCreateUsesExplicitOwnerAccountAndPrivateOnly(t *testing.T) {
	for _, repository := range []string{"fixture-user/assistant", "fixture-org/assistant"} {
		t.Run(repository, func(t *testing.T) {
			created := false
			posted := ""
			c := GitHubSourceClient{run: func(ctx context.Context, token string, args ...string) ([]byte, error) {
				if args[0] == "auth" {
					if strings.Join(args, " ") != "auth token --hostname github.com --user fixture-user" || token != "" {
						t.Fatal("implicit account", args)
					}
					return []byte("synthetic-token\n"), nil
				}
				if token != "synthetic-token" {
					t.Fatal("wrong API credential")
				}
				endpoint := args[4]
				if endpoint == "user" {
					return sourceResponse(200, `{"login":"fixture-user"}`), nil
				}
				if endpoint == "repos/"+repository {
					if !created {
						return sourceResponse(404, `{}`), errors.New("not found")
					}
					return sourceResponse(200, repoPacket(repository, true)), nil
				}
				posted = strings.Join(args, " ")
				created = true
				if !strings.Contains(posted, "private=true") || !strings.Contains(posted, "auto_init=false") || !strings.Contains(posted, "--method POST") {
					t.Fatal("unsafe creation", posted)
				}
				return sourceResponse(201, `{}`), nil
			}}
			if err := c.EnsureRepository(context.Background(), "fixture-user", repository, true); err != nil {
				t.Fatal(err)
			}
			expected := "user/repos"
			if strings.HasPrefix(repository, "fixture-org/") {
				expected = "orgs/fixture-org/repos"
			}
			if !strings.Contains(posted, expected) {
				t.Fatal("wrong owner endpoint", posted)
			}
			posted = ""
			if err := c.EnsureRepository(context.Background(), "fixture-user", repository, true); err != nil {
				t.Fatal(err)
			}
			if posted != "" {
				t.Fatal("resume repeated creation")
			}
		})
	}
}

func TestGitHubRejectsMismatchedIdentityPublicAndUnknownFailures(t *testing.T) {
	for _, mode := range []string{"wrong-user", "public", "unavailable", "missing"} {
		t.Run(mode, func(t *testing.T) {
			c := GitHubSourceClient{run: func(ctx context.Context, token string, args ...string) ([]byte, error) {
				if args[0] == "auth" {
					return []byte("synthetic-token"), nil
				}
				if args[4] == "user" {
					login := "fixture-user"
					if mode == "wrong-user" {
						login = "other-user"
					}
					return sourceResponse(200, `{"login":"`+login+`"}`), nil
				}
				if len(args) > 5 {
					t.Fatal("unexpected mutation")
				}
				if mode == "unavailable" {
					return []byte("synthetic-token accidental raw diagnostic"), errors.New("synthetic-token")
				}
				if mode == "missing" {
					return sourceResponse(404, `{}`), errors.New("not found")
				}
				return sourceResponse(200, repoPacket("fixture-user/assistant", false)), nil
			}}
			err := c.EnsureRepository(context.Background(), "fixture-user", "fixture-user/assistant", false)
			if err == nil || strings.Contains(err.Error(), "synthetic-token") {
				t.Fatalf("unsafe result %v", err)
			}
		})
	}
}

func TestGitHubSourceCredentialBoundToAccountAssistantAndExactPath(t *testing.T) {
	s := fixtureStore(t)
	ctx := context.Background()
	if err := s.InitGit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.ConfigureGitHubSource(ctx, "fixture-user/assistant", "fixture-user"); err != nil {
		t.Fatal(err)
	}
	calls := 0
	c := GitHubSourceClient{run: func(ctx context.Context, token string, args ...string) ([]byte, error) {
		calls++
		if args[0] == "auth" {
			return []byte("synthetic-token"), nil
		}
		if args[4] == "user" {
			return sourceResponse(200, `{"login":"fixture-user"}`), nil
		}
		return sourceResponse(200, repoPacket("fixture-user/assistant", true)), nil
	}}
	request := "protocol=https\nhost=github.com\npath=fixture-user/assistant.git\n\n"
	var out bytes.Buffer
	if err := s.GitHubSourceCredential(ctx, c, "get", strings.NewReader(request), &out); err != nil {
		t.Fatal(err)
	}
	if out.String() != "username=fixture-user\npassword=synthetic-token\n\n" {
		t.Fatal("incorrect credential protocol")
	}
	out.Reset()
	extended := strings.Replace(request, "\n\n", "\nwwwauth[]=one\nwwwauth[]=two\n\n", 1)
	if err := s.GitHubSourceCredential(ctx, c, "get", strings.NewReader(extended), &out); err != nil || out.Len() == 0 {
		t.Fatal("valid Git extension fields rejected", err)
	}
	out.Reset()
	duplicate := strings.Replace(request, "\n\n", "\nhost=elsewhere.test\n\n", 1)
	if err := s.GitHubSourceCredential(ctx, c, "get", strings.NewReader(duplicate), &out); err == nil || out.Len() != 0 {
		t.Fatal("ambiguous credential authority accepted")
	}
	for _, wrong := range []string{strings.Replace(request, "github.com", "elsewhere.test", 1), strings.Replace(request, "assistant.git", "other.git", 1), strings.Replace(request, "https", "http", 1), strings.Replace(request, "\n\n", "\nusername=other-user\n\n", 1)} {
		before := calls
		out.Reset()
		if err := s.GitHubSourceCredential(ctx, c, "get", strings.NewReader(wrong), &out); err != nil || out.Len() != 0 || calls != before {
			t.Fatal("foreign request received credential or contacted provider")
		}
	}
	before := calls
	out.Reset()
	for _, op := range []string{"store", "erase"} {
		if err := s.GitHubSourceCredential(ctx, c, op, strings.NewReader(request), &out); err != nil {
			t.Fatal(err)
		}
	}
	if calls != before || out.Len() != 0 {
		t.Fatal("credential storage mutated")
	}
	if err := s.InitGit(ctx); err != nil {
		t.Fatal(err)
	}
	tracked := gitTest(t, s.Root, "ls-files", ".my-friday/local")
	if tracked != "" {
		t.Fatal("local account binding committed")
	}
	other := filepath.Join(t.TempDir(), "clone")
	gitTest(t, "", "clone", s.Root, other)
	clone, err := Open(other)
	if err != nil {
		t.Fatal(err)
	}
	if err := clone.GitHubSourceCredential(ctx, c, "get", strings.NewReader(request), &out); err == nil {
		t.Fatal("account inherited across clone")
	}
	if calls != before {
		t.Fatal("unbound clone accessed provider")
	}
	path := filepath.Join(s.Root, ".my-friday/local/source-github.json")
	os.Remove(path)
	os.Symlink(filepath.Join(t.TempDir(), "unavailable"), path)
	if err := s.GitHubSourceCredential(ctx, c, "get", strings.NewReader(request), &out); err == nil {
		t.Fatal("symlink binding accepted")
	}
}

func TestSourceGHEnvironmentAndBoundedFailure(t *testing.T) {
	t.Setenv("GH_TOKEN", "ambient-secret")
	t.Setenv("GITHUB_TOKEN", "other-secret")
	t.Setenv("GH_DEBUG", "api")
	t.Setenv("GH_HOST", "wrong.test")
	t.Setenv("GH_CONFIG_DIR", "explicit-local-store")
	env := strings.Join(sourceGHEnvironment("selected-synthetic"), "\n")
	if strings.Contains(env, "ambient-secret") || strings.Contains(env, "other-secret") || strings.Contains(env, "GH_DEBUG") || strings.Contains(env, "wrong.test") || !strings.Contains(env, "GH_CONFIG_DIR=explicit-local-store") {
		t.Fatal("unsafe environment")
	}
	bin := t.TempDir()
	script := "#!/bin/sh\nprintf sensitive-diagnostic >&2\nsleep 2\n"
	os.WriteFile(filepath.Join(bin, "gh"), []byte(script), 0700)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := runSourceGH(ctx, "", "auth", "status")
	if err == nil || strings.Contains(err.Error(), "sensitive") {
		t.Fatal("raw diagnostics leaked or timeout ignored")
	}
}
