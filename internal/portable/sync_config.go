package portable

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// SyncConfig contains only source-synchronization policy. Credential storage,
// provider selection, and account verification belong to the private helper.
type SyncConfig struct {
	Version          int        `json:"schema_version"`
	CredentialHelper []string   `json:"credential_helper,omitempty"`
	Author           *GitAuthor `json:"author,omitempty"`
}

type GitAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (s *Store) syncConfiguration() (SyncConfig, error) {
	config := SyncConfig{Version: 1}
	path := filepath.Join(s.Root, ".my-friday/sync.json")
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return config, nil
	} else if err != nil {
		return config, err
	}
	config = SyncConfig{}
	if err := readJSON(path, &config); err != nil {
		return config, err
	}
	if config.Version != 1 {
		return config, errors.New("unsupported sync configuration version")
	}
	for n, arg := range config.CredentialHelper {
		if strings.ContainsAny(arg, "\x00\r\n") || (n == 0 && strings.TrimSpace(arg) == "") {
			return config, errors.New("invalid sync credential helper argument")
		}
	}
	if a := config.Author; a != nil {
		if strings.TrimSpace(a.Name) == "" || strings.TrimSpace(a.Email) == "" || strings.ContainsAny(a.Name+a.Email, "\x00\r\n") {
			return config, errors.New("sync author requires a name and email")
		}
	}
	return config, nil
}

func (s *Store) gitArguments() []string {
	args := []string{"-C", s.Root, "-c", "core.hooksPath=/dev/null", "-c", "user.name=" + s.Agent.Name, "-c", "user.email=my-friday@localhost", "-c", "commit.gpgsign=false", "-c", "credential.helper=", "-c", "core.askPass="}
	if s.gitSettings == nil {
		return args
	}
	if a := s.gitSettings.Author; a != nil {
		args = append(args, "-c", "user.name="+a.Name, "-c", "user.email="+a.Email)
	}
	if len(s.gitSettings.CredentialHelper) > 0 {
		command := append([]string{}, s.gitSettings.CredentialHelper...)
		if strings.ContainsRune(command[0], '/') && !filepath.IsAbs(command[0]) {
			command[0] = filepath.Join(s.Root, command[0])
		}
		for n := range command {
			command[n] = shellQuote(command[n])
		}
		// Git appends its operation (get/store/erase). Quote argv separately;
		// neither repository paths nor configuration arguments become shell code.
		helper := "!f() { " + strings.Join(command, " ") + " \"$@\"; }; f"
		args = append(args, "-c", "credential.helper="+helper, "-c", "credential.useHttpPath=true")
	}
	return args
}
