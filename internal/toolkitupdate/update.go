// Package toolkitupdate installs explicitly approved toolkit artifacts. It is
// independent of assistant source synchronization and never reads credentials.
package toolkitupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"time"
)

const releaseAPI = "https://api.github.com/repos/acoz-labs/my-friday/releases/latest"
const releasePrefix = "https://github.com/acoz-labs/my-friday/releases/download/"
const maxBinary = 128 << 20

var digestPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var safeName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

type Artifact struct {
	OS      string `json:"os"`
	Arch    string `json:"arch"`
	Name    string `json:"name"`
	SHA256  string `json:"sha256"`
	Version string `json:"-"`
	URL     string `json:"-"`
}
type Manifest struct {
	SchemaVersion      int        `json:"schema_version"`
	PortableFormat     int        `json:"portable_format"`
	ManagementProtocol int        `json:"management_protocol"`
	Version            string     `json:"version"`
	Artifacts          []Artifact `json:"artifacts"`
}
type Client struct {
	get func(context.Context, string, int64) ([]byte, error)
}

func fetch(ctx context.Context, c *http.Client, address string, limit int64) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "my-friday-updater")
	request.Header.Set("Accept", "application/vnd.github+json")
	response, err := c.Do(request)
	if err != nil {
		return nil, errors.New("update server unavailable; current installation is unchanged")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update server returned HTTP %d; retry later", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("update response exceeds size limit")
	}
	return data, nil
}
func (c Client) read(ctx context.Context, address string, limit int64) ([]byte, error) {
	if c.get != nil {
		return c.get(ctx, address, limit)
	}
	client := &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		host := req.URL.Host
		if len(via) >= 5 || req.URL.Scheme != "https" || req.URL.User != nil || (host != "github.com" && host != "release-assets.githubusercontent.com" && host != "objects.githubusercontent.com") {
			return errors.New("unexpected update redirect")
		}
		return nil
	}}
	return fetch(ctx, client, address, limit)
}

// No prereleases, branch heads or legacy releases without the portable contract.
// Digests provide integrity relative to the official HTTPS manifest, not a
// separate signing authority or protection against publisher compromise.
func (c Client) Latest(ctx context.Context) (Artifact, error) {
	var result Artifact
	data, err := c.read(ctx, releaseAPI, 1<<20)
	if err != nil {
		return result, err
	}
	var release struct {
		Tag        string `json:"tag_name"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
		Assets     []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if json.Unmarshal(data, &release) != nil || !safeName.MatchString(release.Tag) || release.Draft || release.Prerelease {
		return result, errors.New("latest release metadata is not eligible")
	}
	manifestURL := releasePrefix + release.Tag + "/my-friday-update.json"
	found := 0
	for _, asset := range release.Assets {
		if asset.Name == "my-friday-update.json" && asset.URL == manifestURL {
			found++
		}
	}
	if found != 1 {
		return result, errors.New("no compatible portable release is published as latest; legacy releases are not installed")
	}
	data, err = c.read(ctx, manifestURL, 1<<20)
	if err != nil {
		return result, err
	}
	var manifest Manifest
	if json.Unmarshal(data, &manifest) != nil || manifest.SchemaVersion != 1 || manifest.PortableFormat != 1 || manifest.ManagementProtocol != 1 || manifest.Version != release.Tag {
		return result, errors.New("release requires an unsupported format or management protocol; installation unchanged")
	}
	matches := 0
	for _, asset := range manifest.Artifacts {
		if asset.OS == runtime.GOOS && asset.Arch == runtime.GOARCH {
			if !safeName.MatchString(asset.Name) || !digestPattern.MatchString(asset.SHA256) {
				return result, errors.New("invalid release artifact metadata")
			}
			matches++
			result = asset
			result.Version = manifest.Version
			result.URL = releasePrefix + release.Tag + "/" + asset.Name
		}
	}
	if matches != 1 {
		return result, errors.New("release has no unambiguous artifact for this operating system and architecture")
	}
	// The exact artifact must also belong to the selected release.
	found = 0
	for _, asset := range release.Assets {
		if asset.Name == result.Name && asset.URL == result.URL {
			found++
		}
	}
	if found != 1 {
		return result, errors.New("release artifact is missing or ambiguous")
	}
	return result, nil
}

func (c Client) Download(ctx context.Context, home string, a Artifact) (string, error) {
	expected := releasePrefix + a.Version + "/" + a.Name
	parsed, err := url.Parse(a.URL)
	if err != nil || parsed.RawQuery != "" || !safeName.MatchString(a.Version) || !safeName.MatchString(a.Name) || a.URL != expected || !digestPattern.MatchString(a.SHA256) {
		return "", errors.New("invalid approved artifact")
	}
	data, err := c.read(ctx, a.URL, maxBinary)
	if err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp("", "my-friday-download-")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		return "", err
	}
	if err = tmp.Close(); err != nil {
		return "", err
	}
	return Stage(home, tmp.Name(), a.SHA256)
}

func Digest(path string) (string, error) {
	st, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !st.Mode().IsRegular() || st.Size() > maxBinary {
		return "", errors.New("artifact must be a regular file under 128 MiB")
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, maxBinary+1))
	if err != nil || n > maxBinary {
		return "", errors.New("cannot hash artifact within size limit")
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func realDirectory(path string) error {
	// Inspect existing ancestors before mkdir; never create through a redirected
	// managed path, even when its final directory does not exist yet.
	ancestor := path
	for {
		if _, err := os.Lstat(ancestor); err == nil {
			resolved, err := filepath.EvalSymlinks(ancestor)
			if err != nil || resolved != ancestor {
				return errors.New("managed installation directory is redirected by a symlink")
			}
			break
		} else if !os.IsNotExist(err) {
			return err
		}
		next := filepath.Dir(ancestor)
		if next == ancestor {
			return errors.New("invalid installation directory")
		}
		ancestor = next
	}
	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	// Canonical home paths are supplied by the menu; reject redirection below it.
	if resolved != abs {
		return errors.New("managed installation directory is redirected by a symlink")
	}
	return nil
}

// Stage is content-addressed and never overwrites a prior executable.
func Stage(home, source, expected string) (string, error) {
	home, err := filepath.EvalSymlinks(home)
	if err != nil {
		return "", err
	}
	if !digestPattern.MatchString(expected) {
		return "", errors.New("enter the full 64-character SHA-256 from a trusted artifact record")
	}
	actual, err := Digest(source)
	if err != nil {
		return "", err
	}
	if actual != expected {
		return "", errors.New("artifact checksum mismatch; nothing activated")
	}
	directory := filepath.Join(home, ".local/share/my-friday/releases", "sha256-"+expected)
	if err := realDirectory(directory); err != nil {
		return "", err
	}
	target := filepath.Join(directory, "my-friday")
	if _, err := os.Lstat(target); err == nil {
		got, err := Digest(target)
		if err != nil || got != expected {
			return "", errors.New("existing immutable artifact differs; preserve it for inspection")
		}
		return target, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	in, err := os.Open(source)
	if err != nil {
		return "", err
	}
	defer in.Close()
	tmp, err := os.CreateTemp(directory, ".download-")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	if _, err = io.Copy(tmp, io.LimitReader(in, maxBinary+1)); err != nil {
		tmp.Close()
		return "", err
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return "", err
	}
	if err = tmp.Close(); err != nil {
		return "", err
	}
	got, err := Digest(tmp.Name())
	if err != nil || got != expected {
		return "", errors.New("artifact changed during installation")
	}
	if err = os.Chmod(tmp.Name(), 0700); err != nil {
		return "", err
	}
	if err = os.Link(tmp.Name(), target); err != nil {
		return "", err
	}
	return target, nil
}

// Activate changes only the convenience symlink. Agent pins remain explicit.
func Activate(home, target, current string) (string, error) {
	home, err := filepath.EvalSymlinks(home)
	if err != nil {
		return "", err
	}
	directory := filepath.Join(home, ".local/bin")
	if err := realDirectory(directory); err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(target)
	if err != nil {
		return "", err
	}
	target = resolved
	link := filepath.Join(directory, "my-friday")
	previous := ""
	if st, err := os.Lstat(link); err == nil {
		if st.Mode()&os.ModeSymlink == 0 {
			return "", errors.New("my-friday command is not a managed symlink; existing command preserved")
		}
		previous, err = filepath.EvalSymlinks(link)
		if err != nil {
			return "", err
		}
		current, err = filepath.EvalSymlinks(current)
		if err != nil || (previous != current && previous != target) {
			return "", errors.New("my-friday command points to another toolkit; reopen that menu before updating")
		}
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if previous == target {
		return previous, nil
	}
	backupDir := filepath.Join(home, ".local/share/my-friday/toolkit-updates")
	if err := realDirectory(backupDir); err != nil {
		return "", err
	}
	backup, err := os.MkdirTemp(backupDir, "activation-")
	if err != nil {
		return "", err
	}
	if previous != "" {
		if err := os.Symlink(previous, filepath.Join(backup, "previous-my-friday")); err != nil {
			return "", err
		}
	}
	// Temporary link is on the same filesystem as the destination.
	stage, err := os.MkdirTemp(directory, ".my-friday-activate-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(stage)
	if err := os.Symlink(target, filepath.Join(stage, "my-friday")); err != nil {
		return "", err
	}
	if err := os.Rename(filepath.Join(stage, "my-friday"), link); err != nil {
		return "", err
	}
	return backup, nil
}
