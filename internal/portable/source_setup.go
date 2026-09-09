package portable

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Shared operation used after the human wizard's confirmation or an explicit
// noninteractive apply request. No prompts or account defaults live here.
type SourceSetupOptions struct {
	Mode         string
	Repository   string
	SetupAccount string
	SyncAccount  string
	Create       bool
	Remote       string
}

func (s *Store) SetupSource(ctx context.Context, options SourceSetupOptions) (SyncStatus, error) {
	var result SyncStatus
	state, err := s.RemoteSetupStatus(ctx)
	if err != nil {
		return result, err
	}
	remote := options.Remote
	switch options.Mode {
	case "github":
		if !validGitHubRepository(options.Repository) || !githubLogin.MatchString(options.SetupAccount) || !githubLogin.MatchString(options.SyncAccount) || options.Remote != "" {
			return result, errors.New("GitHub source setup requires explicit repository, setup account and sync account, with no remote override")
		}
		if state.CustomHelper {
			return result, errors.New("custom helper already configured; preserve it through existing-remote setup")
		}
		remote = "https://github.com/" + options.Repository + ".git"
		if state.Origin != "" && state.Origin != remote {
			return result, errors.New("existing origin differs; nothing replaced")
		}
		client := GitHubSourceClient{}
		if err := client.EnsureRepository(ctx, options.SetupAccount, options.Repository, options.Create); err != nil {
			return result, err
		}
		if options.SyncAccount != options.SetupAccount {
			if err := client.EnsureRepository(ctx, options.SyncAccount, options.Repository, false); err != nil {
				return result, fmt.Errorf("repository may exist, but ongoing account access is not ready; grant/accept its source-repository access separately and resume: %w", err)
			}
		}
		if err := s.ConfigureGitHubSource(ctx, options.Repository, options.SyncAccount); err != nil {
			return result, fmt.Errorf("repository may exist; source configuration needs attention: %w", err)
		}
		if err := s.InitGit(ctx); err != nil {
			return result, fmt.Errorf("source settings saved; checkpoint incomplete, preserve and resume: %w", err)
		}
	case "existing":
		if !validRemoteURL(remote) || options.Repository != "" || options.SetupAccount != "" || options.SyncAccount != "" || options.Create {
			return result, errors.New("existing mode requires only a credential-free HTTPS or absolute local remote")
		}
		if strings.HasPrefix(remote, "https://") && !state.CustomHelper && state.GitHubSource == nil {
			return result, errors.New("configure a private credential helper and checkpoint it before existing HTTPS setup")
		}
	default:
		return result, errors.New("source setup mode must be github or existing")
	}
	if err := s.ConnectRemote(ctx, remote); err != nil {
		return result, fmt.Errorf("remote not connected/verified; any created repo or saved settings are retained for recovery: %w", err)
	}
	return s.Sync(ctx)
}
