package portable

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// MemoryBank is the independent memory format. It contains no harness choice,
// assistant instructions, capability registry or native installation state.
type MemoryBank struct {
	Version int    `json:"schema_version"`
	ID      string `json:"id"`
	Name    string `json:"name"`
}

func CreateMemoryBank(root, name, deviceID, label string) (*Store, error) {
	return createStore(root, name, "", deviceID, label, true)
}

func (s *Store) IsMemoryBank() bool { return s.memoryOnly }

// ValidateAuthorship checks a caller binding without performing a write.
func (s *Store) ValidateAuthorship(author Authorship) error {
	return s.validateCheckpointObserver(&author)
}

// PutSourced validates the entire proposed graph before writing either file,
// under the same lock as synchronization. A disk failure between the writes
// can leave unreferenced evidence, but never a revision with missing evidence.
func (s *Store) PutSourced(r Revision, source Source) error {
	return s.withLock(func() error {
		if err := s.validateSource(source); err != nil {
			return err
		}
		if len(r.Evidence.SourceRefs) != 1 || r.Evidence.SourceRefs[0] != source.ID || r.Authorship.DeviceID != source.DeviceID {
			return errors.New("revision and source must share evidence and authorship")
		}
		records, err := s.revisions()
		if err != nil {
			return err
		}
		if err := s.validateGraphWithSources(append(records, r), map[string]Source{source.ID: source}); err != nil {
			return err
		}
		dir := filepath.Join(s.Root, "memory/records", r.RecordID)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
		info, err := os.Lstat(dir)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("invalid record directory")
		}
		if err := writeNewJSON(filepath.Join(s.Root, "memory/sources", source.ID+".json"), source); err != nil {
			return err
		}
		return writeNewJSON(filepath.Join(dir, r.ID+".json"), r)
	})
}

func openMemoryBank(s *Store) (*Store, error) {
	if _, err := os.Lstat(filepath.Join(s.Root, "agent.json")); !os.IsNotExist(err) {
		return nil, errors.New("ambiguous memory bank and assistant format")
	}
	var bank MemoryBank
	if err := readJSON(filepath.Join(s.Root, "bank.json"), &bank); err != nil {
		return nil, err
	}
	if bank.Version != 1 || !identifier.MatchString(bank.ID) || !strings.HasPrefix(bank.ID, "bank-") || strings.TrimSpace(bank.Name) == "" || strings.ContainsAny(bank.Name, "\x00\r\n") {
		return nil, errors.New("invalid or unsupported memory bank")
	}
	s.memoryOnly = true
	// Reuse the established memory graph and sync primitives. This internal
	// identity projection does not create an agent.json or configure a harness.
	s.Agent = Agent{FormatVersion: bank.Version, ID: bank.ID, Name: bank.Name}
	return s, nil
}

func (s *Store) validateBankFiles() error {
	entries, err := os.ReadDir(s.Root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		switch entry.Name() {
		case "bank.json", ".gitignore", "README.md", ".DS_Store", ".git", ".my-friday", "memory", "provenance":
		default:
			return fmt.Errorf("unexpected memory bank entry %s; preserve it outside the bank before synchronization", entry.Name())
		}
		if entry.Type()&os.ModeSymlink != 0 || (!entry.IsDir() && !entry.Type().IsRegular()) {
			return errors.New("memory bank entries must be regular files or directories")
		}
	}
	return nil
}

// Journal returns recent matching semantic journal entries, never raw harness
// transcripts. Validation covers all entries even when the result is limited.
func (s *Store) Journal(query string, limit int) ([]JournalEntry, error) {
	if limit < 1 || limit > 100 {
		return nil, errors.New("journal limit must be between 1 and 100")
	}
	if err := s.checkDirectories(); err != nil {
		return nil, err
	}
	entries := []JournalEntry{}
	seen := map[string]bool{}
	err := filepath.WalkDir(filepath.Join(s.Root, "memory/events"), func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.Type()&os.ModeSymlink != 0 {
			return errors.New("symlink in memory journal")
		}
		if info.IsDir() || info.Name() == ".gitkeep" {
			return nil
		}
		var entry JournalEntry
		if err := readJSON(path, &entry); err != nil {
			return err
		}
		at, err := time.Parse(time.RFC3339Nano, entry.RecordedAt)
		if err != nil || entry.Version != 1 || !identifier.MatchString(entry.ID) || seen[entry.ID] || strings.TrimSpace(entry.Kind) == "" || strings.TrimSpace(entry.Summary) == "" {
			return errors.New("invalid journal entry")
		}
		if path != filepath.Join(s.Root, "memory/events", at.UTC().Format("2006/01"), entry.ID+".json") {
			return errors.New("journal ID/date/path mismatch")
		}
		if err := s.validateCheckpointObserver(&entry.Authorship); err != nil {
			return err
		}
		seen[entry.ID] = true
		if strings.TrimSpace(query) == "" || relevance(Revision{Summary: entry.Summary, Body: entry.Kind}, query) > 0 {
			entries = append(entries, entry)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].RecordedAt == entries[j].RecordedAt {
			return entries[i].ID < entries[j].ID
		}
		return entries[i].RecordedAt > entries[j].RecordedAt
	})
	if len(entries) > limit {
		entries = entries[:limit]
	}
	return entries, nil
}
