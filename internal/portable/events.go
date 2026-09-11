package portable

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type JournalEntry struct {
	Version    int        `json:"schema_version"`
	ID         string     `json:"id"`
	Kind       string     `json:"kind"`
	Summary    string     `json:"summary"`
	RecordedAt string     `json:"recorded_at"`
	Authorship Authorship `json:"authorship"`
}

func (s *Store) RecordEvent(kind, summary string, author Authorship) (JournalEntry, error) {
	entry := JournalEntry{Version: 1, ID: NewID("event"), Kind: kind, Summary: summary, RecordedAt: time.Now().UTC().Format(time.RFC3339Nano), Authorship: author}
	if strings.TrimSpace(kind) == "" || strings.TrimSpace(summary) == "" {
		return entry, errors.New("event kind and concise summary required")
	}
	if err := s.deviceExists(author.DeviceID); err != nil {
		return entry, err
	}
	err := s.withLock(func() error {
		dir := filepath.Join(s.Root, "memory/events", time.Now().UTC().Format("2006/01"))
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
		return writeNewJSON(filepath.Join(dir, entry.ID+".json"), entry)
	})
	return entry, err
}

func (s *Store) CheckCapability(id string) ([]HandlerResult, error) {
	return s.checkCapability(id, "", "")
}

// CheckCapability resolves machine-local context from a current, valid binding.
// Source-only checks use Store.CheckCapability and never inherit instance state.
func (i Instance) CheckCapability(id string) ([]HandlerResult, error) {
	current, s, err := LoadInstance(i.Root)
	if err != nil {
		return nil, err
	}
	if current != i {
		return nil, errors.New("instance binding changed; reload before checking capability")
	}
	return s.checkCapability(id, current.Root, current.DeviceID)
}

func (s *Store) checkCapability(id, instance, device string) ([]HandlerResult, error) {
	if !identifier.MatchString(id) {
		return nil, errors.New("invalid capability ID")
	}
	caps, err := s.Capabilities()
	if err != nil {
		return nil, err
	}
	for _, c := range caps {
		if c.ID == id {
			for _, sub := range c.Subscriptions {
				if _, err := orderedHandlers(caps, sub.Event); err != nil {
					return nil, err
				}
			}
			results := []HandlerResult{{ID: id, Success: true, Detail: "Manifest and subscription ordering valid."}}
			root, cleanup, err := s.snapshotCapability(id)
			if err != nil {
				return nil, err
			}
			defer cleanup()
			for _, command := range c.Checks {
				if len(command) == 0 {
					return results, errors.New("empty capability check command")
				}
				ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
				_, err := runHandler(ctx, root, s, device, instance, command, nil, true)
				cancel()
				if err != nil {
					return append(results, HandlerResult{ID: id, Success: false, Detail: "Executable check failed; raw output suppressed."}), err
				}
				results = append(results, HandlerResult{ID: id, Success: true, Detail: "Executable check passed."})
			}
			return results, nil
		}
	}
	return nil, errors.New("capability not found")
}
