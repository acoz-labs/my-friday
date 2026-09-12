package portable

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSSHRemoteSyntax(t *testing.T) {
	for _, remote := range []string{"ssh://git@example.test/bank.git", "ssh://example.test:2222/bank.git", "ssh://git@[::1]/bank.git", "git@example.test:bank.git", "example.test:/private/bank.git"} {
		if !isSSHRemote(remote) {
			t.Errorf("valid SSH remote rejected: %s", remote)
		}
	}
	for _, remote := range []string{"https://example.test/bank.git", "http://example.test/bank", "file:///bank", "ext::sh anything", "ssh://git:secret@example.test/bank", "ssh://-oProxyCommand=bad/bank", "ssh://-user@example.test/bank", "ssh://example.test", "ssh://example.test/bank?x=y", "ssh://example.test/bank#fragment", "ssh://example.test/bank%0aevil", "-host:bank", "git@-host:bank", "host:", "host:-option", "host:bank\nevil", "user@@host:bank", "/tmp/with:colon"} {
		if isSSHRemote(remote) {
			t.Errorf("unsafe/non-SSH remote accepted: %q", remote)
		}
	}
}

func TestSyncSSHRemote(t *testing.T) {
	for _, style := range []string{"url", "scp"} {
		t.Run(style, func(t *testing.T) {
			s := fixtureStore(t)
			ctx := context.Background()
			if err := s.InitGit(ctx); err != nil {
				t.Fatal(err)
			}
			bare := filepath.Join(t.TempDir(), "remote.git")
			gitTest(t, "", "init", "--bare", "--initial-branch=main", bare)
			// Only emulate the SSH transport, running real Git upload/receive-pack
			// against the disposable bare repository; no network or credentials.
			helper := filepath.Join(t.TempDir(), "ssh-fixture")
			if err := os.WriteFile(helper, []byte("#!/bin/sh\nfor arg do request=$arg; done\ncase \"$request\" in\n  'git-upload-pack '*|'git-receive-pack '*) exec /bin/sh -c \"$request\" ;;\n  *) exit 2 ;;\nesac\n"), 0700); err != nil {
				t.Fatal(err)
			}
			gitTest(t, s.Root, "config", "core.sshCommand", shellQuote(helper))
			gitTest(t, s.Root, "config", "ssh.variant", "ssh")
			remote := "ssh://fixture@example.test" + bare
			if style == "scp" {
				remote = "fixture@example.test:" + bare
			}
			gitTest(t, s.Root, "remote", "add", "origin", remote)
			state, err := s.Sync(ctx)
			if err != nil || state.State != "synced" {
				t.Fatalf("SSH sync: %+v %v", state, err)
			}
			if strings.TrimSpace(gitTest(t, bare, "rev-parse", "main")) != state.Head {
				t.Fatal("remote did not receive checkpoint")
			}
			if err := s.Put(revision("revision-offline-ssh")); err != nil {
				t.Fatal(err)
			}
			gitTest(t, s.Root, "config", "core.sshCommand", "/usr/bin/false")
			state, err = s.Sync(ctx)
			if err != nil || state.State != "pending" || gitTest(t, s.Root, "status", "--porcelain") != "" {
				t.Fatalf("failed SSH lost local checkpoint: %+v %v", state, err)
			}
		})
	}
}
