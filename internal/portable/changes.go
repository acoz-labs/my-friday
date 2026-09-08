package portable

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

// SourceChange records an observation of the local Git index before a commit,
// not a claim that the observer authored every file or that commit succeeded.
// Memory has its own richer authorship/supersession records and is excluded.
type SourceChange struct {
	Version    int                `json:"schema_version"`
	ID         string             `json:"id"`
	RecordedAt string             `json:"recorded_at"`
	BaseCommit string             `json:"base_commit"`
	Observer   *Authorship        `json:"checkpoint_observer"`
	Files      []SourceFileChange `json:"files"`
}

type SourceFileChange struct {
	Path   string         `json:"path"`
	Before *SourceVersion `json:"before"`
	After  *SourceVersion `json:"after"`
}

type SourceVersion struct {
	ObjectID string `json:"git_object_id"`
	Mode     string `json:"git_mode"`
}

var gitObjectID = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)

// WithCheckpointObserver binds attribution to this operation, never to a
// repository-wide default or an ambient host name. Unbound callers stay unknown.
func (s *Store) WithCheckpointObserver(observer Authorship) *Store {
	work := *s
	work.checkpointObserver = &observer
	return &work
}

func (s *Store) validateCheckpointObserver(observer *Authorship) error {
	if observer == nil {
		return nil
	}
	if strings.TrimSpace(observer.Actor) == "" || strings.TrimSpace(observer.Harness) == "" {
		return errors.New("checkpoint observer requires actor and harness")
	}
	return s.deviceExists(observer.DeviceID)
}

func sourceChangePath(p string) bool {
	return p != "" && utf8.ValidString(p) && !strings.ContainsRune(p, '\x00') &&
		!path.IsAbs(p) && path.Clean(p) == p && p != "." && p != ".." &&
		!strings.HasPrefix(p, "../") && p != ".git" && !strings.HasPrefix(p, ".git/")
}

func sourceChangeTracked(p string) bool {
	return p != "memory" && !strings.HasPrefix(p, "memory/") &&
		p != "provenance" && !strings.HasPrefix(p, "provenance/")
}

func sourceChangeKey(c SourceChange) [32]byte {
	// Stable retries with the same observer/base/file versions reuse the first
	// observation. A different device/session is a separate observation.
	c.ID, c.RecordedAt = "", ""
	data, _ := json.Marshal(c)
	return sha256.Sum256(data)
}

func (s *Store) validateSourceChange(c SourceChange) error {
	if c.Version != 1 || !identifier.MatchString(c.ID) || !strings.HasPrefix(c.ID, "change-") || len(c.Files) == 0 {
		return errors.New("invalid source-change version, ID or files")
	}
	if _, err := time.Parse(time.RFC3339Nano, c.RecordedAt); err != nil {
		return err
	}
	if c.BaseCommit != "" && !gitObjectID.MatchString(c.BaseCommit) {
		return errors.New("invalid source-change base commit")
	}
	if err := s.validateCheckpointObserver(c.Observer); err != nil {
		return err
	}
	previous := ""
	for _, f := range c.Files {
		if !sourceChangePath(f.Path) || !sourceChangeTracked(f.Path) || f.Path <= previous {
			return errors.New("invalid, duplicate or unordered source-change path")
		}
		previous = f.Path
		if f.Before == nil && f.After == nil {
			return errors.New("source change has no file versions")
		}
		if f.Before != nil && f.After != nil && *f.Before == *f.After {
			return errors.New("source change has identical versions")
		}
		for _, v := range []*SourceVersion{f.Before, f.After} {
			if v == nil {
				continue
			}
			if !gitObjectID.MatchString(v.ObjectID) || strings.Trim(v.ObjectID, "0") == "" {
				return errors.New("invalid source-change object ID")
			}
			switch v.Mode {
			case "100644", "100755", "120000", "160000":
			default:
				return errors.New("invalid source-change Git mode")
			}
		}
	}
	return nil
}

// SourceChanges is a read-only listing, optionally filtered by an exact relative
// path. Time ordering is for display, not conflict resolution or supersession.
func (s *Store) SourceChanges(filter string) ([]SourceChange, error) {
	if filter != "" && !sourceChangePath(filter) {
		return nil, errors.New("source-change path must be canonical and repository-relative")
	}
	if err := s.checkDirectories(); err != nil {
		return nil, err
	}
	dir := filepath.Join(s.Root, "provenance/changes")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	result := []SourceChange{}
	for _, entry := range entries {
		if entry.Name() == ".gitkeep" && entry.Type().IsRegular() {
			continue
		}
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".json") {
			return nil, errors.New("unexpected source-change file")
		}
		var c SourceChange
		if err := readJSON(filepath.Join(dir, entry.Name()), &c); err != nil {
			return nil, err
		}
		if entry.Name() != c.ID+".json" {
			return nil, errors.New("source-change ID/path mismatch")
		}
		if err := s.validateSourceChange(c); err != nil {
			return nil, err
		}
		match := filter == ""
		for _, f := range c.Files {
			if f.Path == filter {
				match = true
			}
		}
		if match {
			result = append(result, c)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		a, _ := time.Parse(time.RFC3339Nano, result[i].RecordedAt)
		b, _ := time.Parse(time.RFC3339Nano, result[j].RecordedAt)
		if a.Equal(b) {
			return result[i].ID < result[j].ID
		}
		return a.Before(b)
	})
	return result, nil
}

func parseSourceChanges(raw string) ([]SourceFileChange, error) {
	files := []SourceFileChange{}
	for raw != "" {
		header, rest, ok := strings.Cut(raw, "\x00")
		if !ok {
			return nil, errors.New("invalid source-change Git header")
		}
		p, rest, ok := strings.Cut(rest, "\x00")
		if !ok || !sourceChangePath(p) {
			return nil, errors.New("invalid source-change Git path")
		}
		raw = rest
		fields := strings.Fields(header)
		if len(fields) != 5 || !strings.HasPrefix(fields[0], ":") || (fields[4] != "A" && fields[4] != "M" && fields[4] != "D" && fields[4] != "T") {
			return nil, errors.New("unsupported source-change Git diff")
		}
		if !sourceChangeTracked(p) {
			continue
		}
		f := SourceFileChange{Path: p}
		if fields[0] != ":000000" {
			f.Before = &SourceVersion{ObjectID: fields[2], Mode: strings.TrimPrefix(fields[0], ":")}
		}
		if fields[1] != "000000" {
			f.After = &SourceVersion{ObjectID: fields[3], Mode: fields[1]}
		}
		files = append(files, f)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

// captureSourceChanges runs under the source lock after staging and before
// committing or fetching. It never attributes incoming remote work locally.
func (s *Store) captureSourceChanges(ctx context.Context, base string) error {
	raw, err := s.git(ctx, "diff", "--cached", "--raw", "--no-abbrev", "--no-renames", "--no-ext-diff", "-z", "--", ".")
	if err != nil {
		return err
	}
	files, err := parseSourceChanges(raw)
	if err != nil || len(files) == 0 {
		return err
	}
	c := SourceChange{Version: 1, ID: NewID("change"), RecordedAt: time.Now().UTC().Format(time.RFC3339Nano), BaseCommit: base, Observer: s.checkpointObserver, Files: files}
	// Random IDs keep identical edits on independent machines mergeable, even
	// for unbound callers. Reuse a matching observation already present locally
	// when a previous attempt stopped before commit.
	existing, err := s.SourceChanges("")
	if err != nil {
		return err
	}
	key := sourceChangeKey(c)
	for _, observed := range existing {
		if sourceChangeKey(observed) == key {
			c = observed
			break
		}
	}
	if err := s.validateSourceChange(c); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if len(data)+1 > 4<<20 {
		return errors.New("source-change record exceeds 4 MiB; split the local changes into smaller checkpoints")
	}
	rel := "provenance/changes/" + c.ID + ".json"
	p := filepath.Join(s.Root, filepath.FromSlash(rel))
	if _, err := os.Lstat(p); os.IsNotExist(err) {
		if err := writeNewJSON(p, c); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	// Existing records were validated before staging; retry preserves their time.
	_, err = s.git(ctx, "add", "--", rel)
	return err
}
