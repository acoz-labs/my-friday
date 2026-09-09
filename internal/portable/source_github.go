package portable

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var githubLogin = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{0,38}$`)
var githubRepoName = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,99}$`)

func validGitHubRepository(value string) bool {
	parts := strings.Split(value, "/")
	return len(parts) == 2 && githubLogin.MatchString(parts[0]) && githubRepoName.MatchString(parts[1]) && !strings.HasSuffix(parts[1], ".git")
}
func ValidGitHubRepository(value string) bool { return validGitHubRepository(value) }

type gitHubSourceBinding struct {
	Version     int    `json:"schema_version"`
	AssistantID string `json:"assistant_id"`
	Repository  string `json:"repository"`
	Account     string `json:"account"`
}

// GitHub CLI is used only for assistant-source hosting. No PR/issue/mail or
// password-manager integration is exposed through this adapter.
type GitHubSourceClient struct {
	run func(context.Context, string, ...string) ([]byte, error)
}

func sourceGHEnvironment(token string) []string {
	result := []string{}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if (strings.HasPrefix(key, "GH_") && key != "GH_CONFIG_DIR") || strings.HasPrefix(key, "GITHUB_") {
			continue
		}
		result = append(result, entry)
	}
	result = append(result, "GH_HOST=github.com", "GH_PROMPT_DISABLED=1", "GH_PAGER=cat")
	if token != "" {
		result = append(result, "GH_TOKEN="+token)
	}
	return result
}

func runSourceGH(parent context.Context, token string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Env = sourceGHEnvironment(token)
	configureCommandCancellation(cmd)
	output := &boundedOutput{Limit: (1 << 20) + 1}
	cmd.Stdout = output
	cmd.Stderr = io.Discard
	err := cmd.Run()
	if output.Len() > 1<<20 {
		return nil, errors.New("GitHub response exceeded the source-setup limit")
	}
	if err != nil {
		return output.Bytes(), errors.New("GitHub CLI source operation failed; check selected account access and connectivity")
	}
	return output.Bytes(), nil
}
func (c GitHubSourceClient) invoke(ctx context.Context, token string, args ...string) ([]byte, error) {
	if c.run != nil {
		return c.run(ctx, token, args...)
	}
	return runSourceGH(ctx, token, args...)
}

func (c GitHubSourceClient) Accounts(ctx context.Context) ([]string, error) {
	raw, err := c.invoke(ctx, "", "auth", "status", "--hostname", "github.com", "--json", "hosts")
	if err != nil {
		return nil, err
	}
	var packet struct {
		Hosts map[string][]struct {
			Login string `json:"login"`
		} `json:"hosts"`
	}
	if json.Unmarshal(raw, &packet) != nil {
		return nil, errors.New("could not read GitHub account metadata")
	}
	accounts := []string{}
	for _, a := range packet.Hosts["github.com"] {
		if githubLogin.MatchString(a.Login) {
			accounts = append(accounts, a.Login)
		}
	}
	if len(accounts) == 0 {
		return nil, errors.New("no stored github.com account found; run gh auth login separately, then resume setup")
	}
	return accounts, nil
}

func (c GitHubSourceClient) api(ctx context.Context, token string, args ...string) (int, []byte, error) {
	raw, runErr := c.invoke(ctx, token, append([]string{"api", "--hostname", "github.com", "--include"}, args...)...)
	raw = bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
	header, body, ok := bytes.Cut(raw, []byte("\n\n"))
	fields := strings.Fields(strings.SplitN(string(header), "\n", 2)[0])
	if !ok || len(fields) < 2 || !strings.HasPrefix(fields[0], "HTTP/") {
		return 0, nil, errors.New("GitHub response unavailable; retry setup without changing accounts")
	}
	status, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, nil, errors.New("invalid GitHub response status")
	}
	if status >= 200 && status < 300 && runErr == nil {
		return status, body, nil
	}
	if status == 404 {
		return status, nil, nil
	}
	return status, nil, errors.New("GitHub request was not confirmed; check access and resume (no remote is deleted on failure)")
}

func (c GitHubSourceClient) selectedToken(ctx context.Context, account string) (string, error) {
	if !githubLogin.MatchString(account) {
		return "", errors.New("invalid GitHub account login")
	}
	raw, err := c.invoke(ctx, "", "auth", "token", "--hostname", "github.com", "--user", account)
	if err != nil {
		return "", errors.New("selected GitHub account token unavailable; authenticate it on this machine and resume")
	}
	token := strings.TrimSpace(string(raw))
	if token == "" || len(token) > 16384 || strings.ContainsAny(token, " \t\r\n\x00") {
		return "", errors.New("selected account credential was invalid")
	}
	status, body, err := c.api(ctx, token, "user")
	var user struct {
		Login string `json:"login"`
	}
	if err != nil || status != 200 || json.Unmarshal(body, &user) != nil || !strings.EqualFold(user.Login, account) {
		return "", errors.New("selected credential does not verify as the requested GitHub account")
	}
	return token, nil
}

type sourceRepository struct {
	FullName    string `json:"full_name"`
	Private     bool   `json:"private"`
	Archived    bool   `json:"archived"`
	Disabled    bool   `json:"disabled"`
	Permissions struct {
		Push bool `json:"push"`
	} `json:"permissions"`
}

func verifySourceRepository(body []byte, repository string) error {
	var r sourceRepository
	if json.Unmarshal(body, &r) != nil || !strings.EqualFold(r.FullName, repository) || !r.Private || r.Archived || r.Disabled || !r.Permissions.Push {
		return errors.New("source repository must be private, writable, active and match the selected owner/name")
	}
	return nil
}

// Create only after the wizard's exact-target confirmation. If a prior attempt
// created the repo but stopped later, inspection allows safe resumption.
func (c GitHubSourceClient) EnsureRepository(ctx context.Context, account, repository string, create bool) error {
	if !validGitHubRepository(repository) {
		return errors.New("use GitHub owner/repository, without a URL or .git suffix")
	}
	token, err := c.selectedToken(ctx, account)
	if err != nil {
		return err
	}
	status, body, err := c.api(ctx, token, "repos/"+repository)
	if err != nil {
		return err
	}
	if status == 404 {
		if !create {
			return errors.New("repository not found or not visible to the selected account; nothing created")
		}
		owner, name, _ := strings.Cut(repository, "/")
		endpoint := "user/repos"
		if !strings.EqualFold(owner, account) {
			endpoint = "orgs/" + owner + "/repos"
		}
		_, _, err = c.api(ctx, token, endpoint, "--method", "POST", "-f", "name="+name, "-F", "private=true", "-F", "auto_init=false")
		if err != nil {
			return err
		}
		// Re-read permissions/privacy instead of assuming successful POST means ready.
		status, body, err = c.api(ctx, token, "repos/"+repository)
	}
	if err != nil {
		return err
	}
	if status != 200 {
		return errors.New("private repository creation/access not confirmed; inspect it and resume")
	}
	return verifySourceRepository(body, repository)
}

func (s *Store) ConfigureGitHubSource(ctx context.Context, repository, account string) error {
	return s.configureGitHubSource(ctx, repository, account)
}

// Implements Git credential get only for the explicitly authorized source path.
// store/erase are no-ops: credentials remain in the user's GitHub CLI store.
func (s *Store) GitHubSourceCredential(ctx context.Context, c GitHubSourceClient, operation string, input io.Reader, out io.Writer) error {
	if operation == "store" || operation == "erase" {
		return nil
	}
	if operation != "get" {
		return errors.New("unsupported Git credential operation")
	}
	raw, err := io.ReadAll(io.LimitReader(input, 8193))
	if err != nil || len(raw) > 8192 {
		return errors.New("invalid Git credential request")
	}
	fields := map[string]string{}
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			return errors.New("invalid Git credential request")
		}
		// Git can repeat extension/array fields (for example wwwauth[]).
		// Only authority-bearing scalar fields participate in credential routing.
		if k != "protocol" && k != "host" && k != "path" && k != "username" {
			continue
		}
		if _, exists := fields[k]; exists {
			return errors.New("duplicate Git credential field")
		}
		fields[k] = v
	}
	cfg, err := s.syncConfiguration()
	if err != nil {
		return err
	}
	if cfg.GitHubSource == nil {
		return errors.New("GitHub source authentication is not configured")
	}
	if fields["protocol"] != "https" || fields["host"] != "github.com" || strings.TrimSuffix(fields["path"], ".git") != cfg.GitHubSource.Repository {
		return nil
	}
	local := filepath.Join(s.Root, ".my-friday/local")
	if err := referenceDirectory(local, false); err != nil {
		return errors.New("source account is not bound on this machine; resume setup")
	}
	var binding gitHubSourceBinding
	if err := readJSON(filepath.Join(local, "source-github.json"), &binding); err != nil {
		return errors.New("source account binding unavailable; resume setup")
	}
	if binding.Version != 1 || binding.AssistantID != s.Agent.ID || binding.Repository != cfg.GitHubSource.Repository || !githubLogin.MatchString(binding.Account) {
		return errors.New("source account binding no longer matches; resume setup")
	}
	if fields["username"] != "" && fields["username"] != binding.Account {
		return nil
	}
	token, err := c.selectedToken(ctx, binding.Account)
	if err != nil {
		return err
	}
	status, body, err := c.api(ctx, token, "repos/"+binding.Repository)
	if err != nil || status != 200 {
		return errors.New("source repository access could not be verified")
	}
	if err := verifySourceRepository(body, binding.Repository); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "username=%s\npassword=%s\n\n", binding.Account, token)
	return err
}
