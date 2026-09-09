package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/acoz-labs/my-friday/internal/portable"
)

func remoteSetupWizard(reader *bufio.Reader, out io.Writer, instance portable.Instance, s *portable.Store) error {
	current, err := os.Executable()
	if err != nil {
		return err
	}
	current, err = filepath.EvalSymlinks(current)
	if err != nil {
		return err
	}
	bound, err := filepath.EvalSymlinks(instance.Binary)
	if err != nil || current != bound {
		return errors.New("resume setup with the instance's bound toolkit; deliberately upgrade its executable binding and launcher first if needed")
	}
	ctx := context.Background()
	state, err := s.RemoteSetupStatus(ctx)
	if err != nil {
		return err
	}
	ask := func(prompt, def string) (string, error) {
		fmt.Fprintf(out, "%s [%s]: ", prompt, def)
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", errors.New("setup input ended; nothing further approved; resume with setup --instance PATH")
		}
		line = strings.TrimSpace(line)
		if line == "" {
			line = def
		}
		return line, nil
	}
	fmt.Fprintf(out, "\nSource synchronization for %s\nInstance and native logins are preserved. Source must be clean before setup.\n", s.Agent.Name)
	if state.Origin != "" {
		fmt.Fprintf(out, "Existing origin: %s (will not be replaced)\n", state.Origin)
	}
	mode, err := ask("Remote setup: local (leave unchanged), github, or existing (URL + configured helper)", "local")
	if err != nil {
		return err
	}
	if mode == "local" {
		return outputJSON(out, map[string]any{"state": "unchanged", "origin": state.Origin, "notice": "Local installation preserved. Existing remote settings, if any, are not disabled. Resume with setup --instance PATH."})
	}
	if mode != "github" && mode != "existing" {
		return errors.New("choose local, github, or existing")
	}
	s = s.WithCheckpointObserver(portable.Authorship{DeviceID: instance.DeviceID, Actor: s.Agent.Name, Harness: "setup"})
	var remote string
	if mode == "github" {
		if state.CustomHelper {
			return errors.New("custom sync helper already configured; use existing mode to preserve it")
		}
		client := portable.GitHubSourceClient{}
		accounts, err := client.Accounts(ctx)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Stored github.com accounts: %s\nNo active-account switch or token copying is performed.\n", strings.Join(accounts, ", "))
		account, err := ask("Account for repository setup/creation", accounts[0])
		if err != nil {
			return err
		}
		known := false
		for _, a := range accounts {
			if a == account {
				known = true
			}
		}
		if !known {
			return errors.New("select a listed account, or authenticate the intended account separately and resume")
		}
		target := account + "/" + s.Agent.Name
		if state.GitHubSource != nil {
			target = state.GitHubSource.Repository
		} else if strings.HasPrefix(state.Origin, "https://github.com/") {
			target = strings.TrimSuffix(strings.TrimPrefix(state.Origin, "https://github.com/"), ".git")
		}
		repository, err := ask("Repository owner/name (your personal account or an organization)", target)
		if err != nil {
			return err
		}
		if !portable.ValidGitHubRepository(repository) {
			return errors.New("expected owner/repository, without URL or .git suffix")
		}
		remote = "https://github.com/" + repository + ".git"
		if state.Origin != "" && state.Origin != remote {
			return errors.New("selected repository differs from existing origin; nothing replaced")
		}
		action, err := ask("Create private repository if absent, or connect only (create/connect)", "connect")
		if err != nil {
			return err
		}
		if action != "create" && action != "connect" {
			return errors.New("choose create or connect")
		}
		syncAccount, err := ask("Account authorized for ONGOING source sync (may differ from setup account)", account)
		if err != nil {
			return err
		}
		known = false
		for _, a := range accounts {
			if a == syncAccount {
				known = true
			}
		}
		if !known {
			return errors.New("authenticate the intended source-sync account separately and resume")
		}
		fmt.Fprintf(out, "\nPlan: %s PRIVATE %s using setup account %s; use %s for ongoing source synchronization on this machine.\nUpload this assistant's complete committed history, including private memory and capabilities.\nReview the source/history for secrets before proceeding. No secret scanner or history rewrite is provided.\nOther machines must explicitly bind their source account. Existing remote must be empty or this same assistant.\n", action, repository, account, syncAccount)
		confirmed, err := ask("To approve, type the exact owner/name", "")
		if err != nil {
			return err
		}
		if confirmed != repository {
			return errors.New("repository setup not approved; nothing created or configured")
		}
		if err := client.EnsureRepository(ctx, account, repository, action == "create"); err != nil {
			return err
		}
		if syncAccount != account {
			if err := client.EnsureRepository(ctx, syncAccount, repository, false); err != nil {
				return fmt.Errorf("repository may exist, but ongoing account access is not ready; grant/accept its source-repository access separately and resume: %w", err)
			}
		}
		if err := s.ConfigureGitHubSource(ctx, repository, syncAccount); err != nil {
			return fmt.Errorf("repository may exist; source configuration needs attention: %w", err)
		}
		// Checkpoint only wizard-owned configuration; preflight refused unrelated dirt.
		if err := s.InitGit(ctx); err != nil {
			return fmt.Errorf("source settings saved; checkpoint incomplete, preserve and resume: %w", err)
		}
	} else {
		fmt.Fprintln(out, "Existing mode uses your already configured private helper for HTTPS. It does not install credentials or verify hosting-service visibility. Review privacy yourself; local remotes also work.")
		remote, err = ask("Existing remote HTTPS URL or absolute local path", state.Origin)
		if err != nil {
			return err
		}
		if !portable.ValidSourceRemote(remote) {
			return errors.New("use a credential-free HTTPS URL or an absolute local remote path")
		}
		if strings.HasPrefix(remote, "https://") && !state.CustomHelper && state.GitHubSource == nil {
			return errors.New("configure a private credential_helper in .my-friday/sync.json, checkpoint it, then resume (or use github mode)")
		}
		fmt.Fprintf(out, "This will verify origin and upload complete committed assistant history to %s.\n", remote)
		confirmation, err := ask("Approve source synchronization (yes/no)", "no")
		if err != nil {
			return err
		}
		if confirmation != "yes" {
			return errors.New("source synchronization not approved")
		}
	}
	if err := s.ConnectRemote(ctx, remote); err != nil {
		return fmt.Errorf("remote not connected/verified; any created repo or saved settings are retained for recovery: %w", err)
	}
	status, err := s.Sync(ctx)
	if err != nil {
		return err
	}
	if err := outputJSON(out, status); err != nil {
		return err
	}
	if status.State != "synced" {
		return errors.New("synchronization not confirmed; local history preserved; resolve the reported status and resume")
	}
	fmt.Fprintln(out, "Verified fetch/push synchronization. No launcher, instance binding, native login, memory claim or agent capability was recreated.")
	return nil
}

func portableSourceCredential(args []string, input io.Reader, out, errout io.Writer) error {
	f := portableFlags("source-credential", errout)
	root := f.String("repository", "", "Exact assistant source; internal Git-helper endpoint")
	if err := f.Parse(args); err != nil {
		return err
	}
	if *root == "" || f.NArg() != 1 {
		return errors.New("internal Git helper requires --repository PATH and get/store/erase; do not invoke get manually")
	}
	s, err := portable.Open(*root)
	if err != nil {
		return err
	}
	return s.GitHubSourceCredential(context.Background(), portable.GitHubSourceClient{}, f.Arg(0), input, out)
}
