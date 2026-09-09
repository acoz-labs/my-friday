package portable

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RemoteSetupState struct {
	Origin       string              `json:"origin,omitempty"`
	GitHubSource *GitHubSourceConfig `json:"github_source,omitempty"`
	CustomHelper bool                `json:"custom_helper"`
}

func validRemoteURL(remote string) bool {
	if remote == "" || strings.ContainsAny(remote, "\x00\r\n") {
		return false
	}
	if filepath.IsAbs(remote) {
		return true
	}
	u, err := url.Parse(remote)
	return err == nil && u.Scheme == "https" && u.Host != "" && u.User == nil && u.RawQuery == "" && u.Fragment == "" && u.Opaque == ""
}

func ValidSourceRemote(remote string) bool { return validRemoteURL(remote) }

func (s *Store) remoteSetupState(ctx context.Context) (RemoteSetupState, error) {
	var result RemoteSetupState
	if err := s.gitBoundary(ctx); err != nil {
		return result, err
	}
	if err := s.Validate(); err != nil {
		return result, err
	}
	dirty, err := s.git(ctx, "status", "--porcelain")
	if err != nil {
		return result, err
	}
	if dirty != "" {
		return result, errors.New("source has uncommitted work; checkpoint or reconcile it before remote setup")
	}
	cfg, err := s.syncConfiguration()
	if err != nil {
		return result, err
	}
	result.GitHubSource = cfg.GitHubSource
	result.CustomHelper = len(cfg.CredentialHelper) > 0
	remotes, err := s.git(ctx, "remote")
	if err != nil {
		return result, err
	}
	for _, name := range strings.Fields(remotes) {
		if name != "origin" {
			continue
		}
		raw, err := s.git(ctx, "config", "--get-all", "remote.origin.url")
		if err != nil {
			return result, err
		}
		if len(strings.Split(raw, "\n")) != 1 || !validRemoteURL(raw) {
			return result, errors.New("origin must have one local or credential-free HTTPS URL; review its configuration manually")
		}
		if push, _ := s.git(ctx, "config", "--get-all", "remote.origin.pushurl"); push != "" {
			return result, errors.New("origin has a separate push URL; review it manually before setup")
		}
		if mirror, _ := s.git(ctx, "config", "--get", "remote.origin.mirror"); mirror != "" && mirror != "false" {
			return result, errors.New("origin uses mirror settings; review it manually before setup")
		}
		if fetch, _ := s.git(ctx, "config", "--get-all", "remote.origin.fetch"); fetch != "+refs/heads/*:refs/remotes/origin/*" {
			return result, errors.New("origin uses a nonstandard fetch mapping; review it manually before setup")
		}
		result.Origin = raw
	}
	return result, nil
}

// Read-only preflight. Never checkpoint unrelated dirty work as part of setup.
func (s *Store) RemoteSetupStatus(parent context.Context) (RemoteSetupState, error) {
	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()
	var result RemoteSetupState
	err := s.withLock(func() error { var err error; result, err = s.remoteSetupState(ctx); return err })
	return result, err
}

// ConnectRemote verifies an empty remote or this same assistant before recording
// origin. It never replaces an existing origin or pushes into an unrelated repo.
// Failed verification may leave fetched Git objects/FETCH_HEAD, not source edits.
func (s *Store) ConnectRemote(parent context.Context, remote string) error {
	if !validRemoteURL(remote) {
		return errors.New("use a credential-free HTTPS URL or an absolute local remote path")
	}
	work := *s
	s = &work
	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()
	return s.withLock(func() error {
		state, err := s.remoteSetupState(ctx)
		if err != nil {
			return err
		}
		if state.Origin != "" && state.Origin != remote {
			return errors.New("origin already points elsewhere; setup will not replace it")
		}
		cfg, err := s.syncConfiguration()
		if err != nil {
			return err
		}
		s.gitSettings = &cfg
		refs, err := s.git(ctx, "ls-remote", "--refs", remote)
		if err != nil {
			return errors.New("remote access could not be verified; source and origin preserved")
		}
		if refs != "" {
			hasMain := false
			for _, line := range strings.Split(refs, "\n") {
				if strings.HasSuffix(line, "\trefs/heads/main") {
					hasMain = true
				}
			}
			if !hasMain {
				return errors.New("nonempty remote has no main branch; refusing to attach it")
			}
			if _, err := s.git(ctx, "fetch", "--no-tags", "--no-recurse-submodules", remote, "refs/heads/main"); err != nil {
				return errors.New("could not inspect remote main; origin preserved")
			}
			data, err := s.git(ctx, "show", "FETCH_HEAD:agent.json")
			if err != nil {
				return errors.New("remote is not a My Friday assistant")
			}
			var agent Agent
			if json.Unmarshal([]byte(data), &agent) != nil || agent.ID != s.Agent.ID {
				return errors.New("remote belongs to another assistant; import it into a separate installation instead")
			}
			// Full reconciliation/validation happens through Sync after attachment.
		}
		if state.Origin == "" {
			_, err = s.git(ctx, "remote", "add", "origin", remote)
		}
		return err
	})
}

func (s *Store) configureGitHubSource(ctx context.Context, repository, account string) error {
	if !validGitHubRepository(repository) || !githubLogin.MatchString(account) {
		return errors.New("invalid GitHub repository or account")
	}
	return s.withLock(func() error {
		state, err := s.remoteSetupState(ctx)
		if err != nil {
			return err
		}
		target := "https://github.com/" + repository + ".git"
		if state.Origin != "" && state.Origin != target {
			return errors.New("existing origin differs from the selected repository")
		}
		cfg, err := s.syncConfiguration()
		if err != nil {
			return err
		}
		if len(cfg.CredentialHelper) > 0 {
			return errors.New("custom credential helper exists; preserve it and use the existing-remote flow")
		}
		local := filepath.Join(s.Root, ".my-friday/local")
		if err := referenceDirectory(local, true); err != nil {
			return err
		}
		ignored, err := s.git(ctx, "check-ignore", "--no-index", "--", ".my-friday/local/source-github.json")
		if err != nil || ignored == "" {
			return errors.New("machine-local source credentials metadata must be Git-ignored")
		}
		if tracked, _ := s.git(ctx, "ls-files", "--", ".my-friday/local"); tracked != "" {
			return errors.New("local state is tracked; fix the source boundary first")
		}
		p := filepath.Join(local, "source-github.json")
		if info, err := os.Lstat(p); err == nil && !info.Mode().IsRegular() {
			return errors.New("source account binding must be a regular file")
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}
		binding := gitHubSourceBinding{Version: 1, AssistantID: s.Agent.ID, Repository: repository, Account: account}
		data, _ := json.MarshalIndent(binding, "", "  ")
		if err := replaceProjectionFile(p, append(data, '\n')); err != nil {
			return err
		}
		cfg.GitHubSource = &GitHubSourceConfig{Repository: repository}
		data, _ = json.MarshalIndent(cfg, "", "  ")
		return replaceProjectionFile(filepath.Join(s.Root, ".my-friday/sync.json"), append(data, '\n'))
	})
}
