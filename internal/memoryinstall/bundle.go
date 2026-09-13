// Package memoryinstall connects a memory bank through Codex's native plugin
// installer. It does not own an agent home, authentication, or memory semantics.
package memoryinstall

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/acoz-labs/my-friday/internal/memorybank"
	"github.com/acoz-labs/my-friday/internal/toolkitupdate"
	codexplugin "github.com/acoz-labs/my-friday/plugins/codex"
)

type Options struct {
	Home       string `json:"installation_home"`
	CodexHome  string `json:"codex_home"`
	Codex      string `json:"codex_binary"`
	Binary     string `json:"source_binary"`
	Binding    string `json:"binding"`
	Generation string `json:"generation,omitempty"`
}

type Plan struct {
	Options
	SchemaVersion int    `json:"schema_version"`
	BankID        string `json:"bank_id"`
	BinarySHA256  string `json:"binary_sha256"`
	Runtime       string `json:"runtime"`
	Root          string `json:"marketplace_root"`
	Marketplace   string `json:"marketplace"`
	PluginID      string `json:"plugin_id"`
	Version       string `json:"plugin_version"`
}

func hash(b []byte) string { v := sha256.Sum256(b); return hex.EncodeToString(v[:]) }
func inside(root, path string) bool {
	r, e := filepath.Rel(root, path)
	return e == nil && r != ".." && !strings.HasPrefix(r, ".."+string(filepath.Separator))
}

// Resolve an absent destination through its existing ancestor without writing.
func canonical(path string) (string, error) {
	if !filepath.IsAbs(path) || strings.ContainsAny(path, "\x00\r\n") {
		return "", errors.New("absolute paths without control characters required")
	}
	path = filepath.Clean(path)
	if _, err := os.Lstat(path); err == nil {
		return filepath.EvalSymlinks(path)
	} else if !os.IsNotExist(err) {
		return "", err
	}
	parent, err := canonical(filepath.Dir(path))
	if err != nil {
		return "", err
	}
	return filepath.Join(parent, filepath.Base(path)), nil
}

func packageFiles() (map[string][]byte, error) {
	files := map[string][]byte{}
	err := fs.WalkDir(codexplugin.Files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := codexplugin.Files.ReadFile(path)
		if err == nil {
			files[path] = b
		}
		return err
	})
	return files, err
}

// Prepare only reads local inputs. Applying a plan executes the explicitly
// selected native Codex executable and copies the selected trusted runtime.
func Prepare(o Options) (Plan, error) {
	var p Plan
	var err error
	for _, s := range []*string{&o.Home, &o.CodexHome, &o.Codex, &o.Binary, &o.Binding} {
		*s, err = canonical(*s)
		if err != nil {
			return p, err
		}
	}
	s, err := memorybank.OpenBinding(o.Binding, "installation")
	if err != nil {
		return p, err
	}
	if inside(s.Root(), o.Home) || inside(s.Root(), o.CodexHome) {
		return p, errors.New("installation and native profile must be outside the Git-backed bank")
	}
	digest, err := toolkitupdate.Digest(o.Binary)
	if err != nil {
		return p, err
	}
	files, err := packageFiles()
	if err != nil {
		return p, err
	}
	var marketplace struct {
		Name string `json:"name"`
	}
	if err = json.Unmarshal(files[".agents/plugins/marketplace.json"], &marketplace); err != nil || marketplace.Name == "" {
		return p, errors.New("invalid bundled marketplace")
	}
	p = Plan{Options: o, SchemaVersion: 1, BankID: s.ID(), BinarySHA256: digest, Marketplace: marketplace.Name, PluginID: "my-friday-memory@" + marketplace.Name}
	p.Runtime = filepath.Join(o.Home, ".local/share/my-friday/releases", "sha256-"+digest, "my-friday")
	identity, _ := json.Marshal(struct {
		Options Options
		Digest  string
		Files   map[string][]byte
	}{o, digest, files})
	key := hash(identity)
	p.Root = filepath.Join(managedRoot(o), key)
	var manifest struct {
		Version string `json:"version"`
	}
	if json.Unmarshal(files["plugins/my-friday-memory/.codex-plugin/plugin.json"], &manifest) != nil || manifest.Version == "" {
		return Plan{}, errors.New("invalid plugin version")
	}
	base, _, _ := strings.Cut(manifest.Version, "+")
	p.Version = base + "+codex." + key
	if inside(s.Root(), p.Root) || inside(s.Root(), p.Runtime) {
		return Plan{}, errors.New("managed runtime/plugin would be inside the bank")
	}
	// Existing managed ancestors must not redirect writes, including into a bank.
	resolved, err := canonical(p.Root)
	if err != nil || resolved != p.Root {
		return Plan{}, errors.New("managed plugin destination is redirected; preserve it and select a real installation home")
	}
	return p, nil
}

func managedRoot(o Options) string {
	return filepath.Join(o.Home, ".local/share/my-friday/memory-codex", hash([]byte(o.CodexHome)))
}
func quote(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }

func bundle(p Plan) (map[string][]byte, error) {
	files, err := packageFiles()
	if err != nil {
		return nil, err
	}
	prefix := "plugins/my-friday-memory/"
	// Pin machine defaults without rewriting shell startup files. Explicit
	// per-process overrides remain available for deliberate bank selection.
	files[prefix+"scripts/connection.sh"] = []byte("#!/bin/sh\n" +
		"if [ -z \"${MY_FRIDAY_MEMORY_BIN:-}\" ]; then MY_FRIDAY_MEMORY_BIN=" + quote(p.Runtime) + "; fi\n" +
		"if [ -z \"${MY_FRIDAY_MEMORY_BINDING:-}\" ]; then MY_FRIDAY_MEMORY_BINDING=" + quote(p.Binding) + "; fi\n" +
		"export MY_FRIDAY_MEMORY_BIN MY_FRIDAY_MEMORY_BINDING\n" +
		"exec /bin/sh \"$(dirname \"$0\")/run-memory.sh\" \"$@\"\n")
	for _, name := range []string{prefix + ".mcp.json", prefix + "hooks/hooks.json"} {
		files[name] = bytes.ReplaceAll(files[name], []byte("scripts/run-memory.sh"), []byte("scripts/connection.sh"))
	}
	var manifest map[string]any
	if err = json.Unmarshal(files[prefix+".codex-plugin/plugin.json"], &manifest); err != nil {
		return nil, err
	}
	// The source's base version is preserved. The digest identifies the exact
	// bundled bytes, selected runtime and local connection for native caching.
	base, _, _ := strings.Cut(manifest["version"].(string), "+")
	p.Version = base + "+codex." + filepath.Base(p.Root)
	manifest["version"] = p.Version
	files[prefix+".codex-plugin/plugin.json"], err = json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	files["connection.json"], err = json.MarshalIndent(p, "", "  ")
	return files, err
}

func verifyBundle(p Plan) error {
	files, err := bundle(p)
	if err != nil {
		return err
	}
	count := 0
	err = filepath.WalkDir(p.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return errors.New("managed bundle contains a symlink")
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(p.Root, path)
		if err != nil {
			return err
		}
		expected, ok := files[filepath.ToSlash(rel)]
		if !ok {
			return errors.New("managed bundle contains an unexpected file")
		}
		st, err := d.Info()
		if err != nil || !st.Mode().IsRegular() || st.Size() != int64(len(expected)) {
			return errors.New("managed bundle file type or size changed")
		}
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, expected) {
			return errors.New("managed bundle changed; preserve edits before reconnecting")
		}
		count++
		return nil
	})
	if err == nil && count != len(files) {
		return errors.New("managed bundle is incomplete")
	}
	return err
}

func verifyCache(p Plan) error {
	files, err := bundle(p)
	if err != nil {
		return err
	}
	prefix := "plugins/my-friday-memory/"
	root := filepath.Join(p.CodexHome, "plugins/cache", p.Marketplace, "my-friday-memory", p.Version)
	resolved, err := canonical(root)
	if err != nil || resolved != root {
		return errors.New("native cache path redirected; inspect native installation")
	}
	for name, expected := range files {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		path := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(name, prefix)))
		resolved, err := canonical(path)
		if err != nil || resolved != path {
			return errors.New("native plugin cache file path redirected; inspect before repairing")
		}
		st, err := os.Lstat(path)
		if err != nil {
			return errors.New("native plugin cache is incomplete; use repair and then start a fresh session")
		}
		if !st.Mode().IsRegular() || st.Size() != int64(len(expected)) {
			return errors.New("native plugin cache file changed; use repair, retaining previous edits")
		}
		data, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(data, expected) {
			return errors.New("native plugin cache differs from the selected connection; use repair")
		}
	}
	return nil
}

func publish(p Plan) error {
	if _, err := os.Lstat(p.Root); err == nil {
		return verifyBundle(p)
	} else if !os.IsNotExist(err) {
		return err
	}
	parent := filepath.Dir(p.Root)
	resolved, err := canonical(parent)
	if err != nil || resolved != parent {
		return errors.New("managed bundle parent redirected")
	}
	if err = os.MkdirAll(parent, 0700); err != nil {
		return err
	}
	temp, err := os.MkdirTemp(parent, ".prepare-")
	if err != nil {
		return err
	}
	// Only this function's newly allocated temporary directory is removed.
	defer os.RemoveAll(temp)
	files, err := bundle(p)
	if err != nil {
		return err
	}
	for name, data := range files {
		path := filepath.Join(temp, filepath.FromSlash(name))
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return err
		}
		if err = os.WriteFile(path, data, 0600); err != nil {
			return err
		}
	}
	if err = os.Rename(temp, p.Root); err != nil {
		return fmt.Errorf("bundle publication failed; existing files preserved: %w", err)
	}
	return verifyBundle(p)
}
