package main

import (
	"errors"
	"fmt"
	"sort"
)

var ErrTransportCallLimit = errors.New("transport call limit exceeded")

type CountedTransport struct {
	bundle        Bundle
	task          Task
	calls         int
	fallbackUsed  bool
	fixtures      map[string]string
	primaryLoaded map[string]bool
}

func NewCountedTransport(bundle Bundle, task Task) (*CountedTransport, error) {
	if err := ValidateBundle(bundle); err != nil {
		return nil, err
	}
	if task.IndexRevision != bundle.Capabilities.Revision {
		return nil, ErrStaleIndex
	}
	known := false
	for _, candidate := range bundle.Tasks.Tasks {
		if candidate.ID == task.ID {
			known = true
			break
		}
	}
	if !known {
		return nil, errors.New("transport task is not in the frozen corpus")
	}
	fixtures := map[string]string{}
	for _, fixture := range task.Fixtures {
		fixtures[fixture.Path] = fixture.Content
	}
	return &CountedTransport{bundle: bundle, task: task, fixtures: fixtures, primaryLoaded: map[string]bool{}}, nil
}

func (transport *CountedTransport) count() error {
	transport.calls++
	if transport.calls > 8 {
		return ErrTransportCallLimit
	}
	return nil
}

func (transport *CountedTransport) Calls() int { return transport.calls }

func (transport *CountedTransport) Lookup(query string) ([]CapabilityMetadata, error) {
	if err := transport.count(); err != nil {
		return nil, err
	}
	results := NewBM25Index(transport.bundle.Capabilities.Capabilities).Query(query, 3)
	metadata := make([]CapabilityMetadata, 0, len(results))
	byID := map[string]Capability{}
	for _, capability := range transport.bundle.Capabilities.Capabilities {
		byID[capability.ID] = capability
	}
	for _, result := range results {
		capability := byID[result.ID]
		metadata = append(metadata, CapabilityMetadata{ID: capability.ID, Name: capability.Name, Summary: capability.Summary, Revision: capability.Revision})
	}
	return metadata, nil
}

func (transport *CountedTransport) Load(ids []string) ([]Capability, error) {
	if err := transport.count(); err != nil {
		return nil, err
	}
	primary := map[string]bool{}
	for id := range transport.primaryLoaded {
		primary[id] = true
	}
	for _, id := range ids {
		primary[id] = true
	}
	requested := make([]string, 0, len(primary))
	for id := range primary {
		requested = append(requested, id)
	}
	if len(requested) > 2 {
		return nil, errors.New("cumulative primary capability limit exceeded")
	}
	loaded, err := SelectCapabilitySet(transport.bundle, transport.task, requested)
	if err != nil {
		return nil, err
	}
	transport.primaryLoaded = primary
	return loaded, nil
}

func (transport *CountedTransport) Fallback() ([]CapabilityMetadata, error) {
	if err := transport.count(); err != nil {
		return nil, err
	}
	if transport.fallbackUsed {
		return nil, errors.New("broader metadata fallback already used")
	}
	transport.fallbackUsed = true
	metadata := make([]CapabilityMetadata, 0, len(transport.bundle.Capabilities.Capabilities))
	for _, capability := range transport.bundle.Capabilities.Capabilities {
		metadata = append(metadata, CapabilityMetadata{ID: capability.ID, Name: capability.Name, Summary: capability.Summary, Revision: capability.Revision})
	}
	sort.Slice(metadata, func(i, j int) bool { return metadata[i].ID < metadata[j].ID })
	return metadata, nil
}

func (transport *CountedTransport) ReadFixture(path string) (string, error) {
	if err := transport.count(); err != nil {
		return "", err
	}
	if !safeFixturePath(path) || !contains(transport.task.ReadPaths, path) {
		return "", fmt.Errorf("read path %q is outside task authority", path)
	}
	content, exists := transport.fixtures[path]
	if !exists {
		return "", fmt.Errorf("fixture %q is absent", path)
	}
	return content, nil
}

func (transport *CountedTransport) WriteFixture(path, content string) error {
	if err := transport.count(); err != nil {
		return err
	}
	if !safeFixturePath(path) || !contains(transport.task.WritePaths, path) {
		return fmt.Errorf("write path %q is outside task authority", path)
	}
	transport.fixtures[path] = content
	return nil
}

func (transport *CountedTransport) Snapshot() []FixtureSnapshot {
	paths := make([]string, 0, len(transport.fixtures))
	for path := range transport.fixtures {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	result := make([]FixtureSnapshot, 0, len(paths))
	for _, path := range paths {
		result = append(result, FixtureSnapshot{Path: path, Content: transport.fixtures[path]})
	}
	return result
}
