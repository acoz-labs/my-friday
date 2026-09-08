package portable

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"sort"
	"strings"
	"syscall"
	"unicode/utf8"
)

const referenceMaxFile = 1 << 20
const referenceMaxEntries = 2000
const referenceMaxBytes = 32 << 20

type ReferenceMatch struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int    `json:"bytes"`
	Score  int    `json:"score"`
}
type ReferenceSearch struct {
	Usage          string           `json:"usage"`
	Library        ReferenceLibrary `json:"library"`
	LibrarySHA256  string           `json:"library_sha256"`
	Matches        []ReferenceMatch `json:"matches"`
	ScannedFiles   int              `json:"scanned_files"`
	SkippedEntries int              `json:"skipped_entries"`
	Truncated      bool             `json:"truncated"`
	Notice         string           `json:"notice"`
}
type ReferenceDocument struct {
	Usage         string           `json:"usage"`
	Library       ReferenceLibrary `json:"library"`
	LibrarySHA256 string           `json:"library_sha256"`
	Path          string           `json:"path"`
	SHA256        string           `json:"sha256"`
	Text          string           `json:"text"`
	Notice        string           `json:"notice"`
}

func referencePath(p string) bool {
	if !sourceChangePath(p) {
		return false
	}
	for _, part := range strings.Split(p, "/") {
		if strings.HasPrefix(part, ".") || part == "node_modules" {
			return false
		}
	}
	return true
}

func readReferenceFile(root *os.Root, p string) ([]byte, error) {
	if !referencePath(p) {
		return nil, errors.New("reference file must be a canonical relative path; hidden paths and node_modules are excluded")
	}
	parts := strings.Split(p, "/")
	for n := range parts {
		info, err := root.Lstat(strings.Join(parts[:n+1], "/"))
		if err != nil {
			return nil, errors.New("reference file unavailable")
		}
		if n < len(parts)-1 && !info.IsDir() {
			return nil, errors.New("reference traversal through a symlink or non-directory is not allowed")
		}
		if n == len(parts)-1 && (!info.Mode().IsRegular() || info.Size() > referenceMaxFile) {
			return nil, errors.New("reference file must be regular text of at most 1 MiB")
		}
	}
	// Root confines traversal; nonblocking/no-follow avoids special-file or
	// final-symlink replacement races. No source is executed or written.
	f, err := root.OpenFile(p, os.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, errors.New("reference file cannot be opened safely")
	}
	defer f.Close()
	before, err := f.Stat()
	if err != nil || !before.Mode().IsRegular() || before.Size() > referenceMaxFile {
		return nil, errors.New("reference file must be regular text of at most 1 MiB")
	}
	data, err := io.ReadAll(io.LimitReader(f, referenceMaxFile+1))
	if err != nil {
		return nil, errors.New("reference file read failed")
	}
	if len(data) > referenceMaxFile || !utf8.Valid(data) || strings.ContainsRune(string(data), '\x00') {
		return nil, errors.New("reference file is oversized or not UTF-8 text")
	}
	after, err := f.Stat()
	if err != nil || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		return nil, errors.New("reference file changed while reading; retry")
	}
	return data, nil
}

func (i Instance) ReadReference(s *Store, id, p, expectedSHA256 string) (ReferenceDocument, error) {
	result := ReferenceDocument{Usage: "reference-only", Path: p, Notice: ReferenceNotice}
	root, lib, err := i.openReference(s, id)
	if err != nil {
		return result, err
	}
	defer root.Close()
	result.Library, result.LibrarySHA256 = lib, libraryHash(lib)
	data, err := readReferenceFile(root, p)
	if err != nil {
		return result, err
	}
	result.SHA256 = referenceHash(data)
	if expectedSHA256 != "" && expectedSHA256 != result.SHA256 {
		return result, errors.New("reference content changed since the selected version; search again or explicitly read the current version")
	}
	result.Text = string(data)
	return result, nil
}

func (i Instance) SearchReference(s *Store, id, query string, limit int) (ReferenceSearch, error) {
	result := ReferenceSearch{Usage: "reference-only", Matches: []ReferenceMatch{}, Notice: ReferenceNotice + " Search is bounded lexical discovery, not an exhaustive audit. Hidden paths, node_modules, symlinks, non-text, oversized and unreadable files are skipped. An empty result does not prove relevant experience is absent. Read selected files with --sha256 before using them."}
	if limit < 1 || limit > 100 {
		return result, errors.New("reference search limit must be 1-100")
	}
	if len(query) > 4096 {
		return result, errors.New("reference query exceeds 4096 bytes")
	}
	if strings.TrimSpace(query) != "" && len(terms(query)) == 0 {
		return result, errors.New("reference query needs searchable words or an empty query for listing")
	}
	root, lib, err := i.openReference(s, id)
	if err != nil {
		return result, err
	}
	defer root.Close()
	result.Library, result.LibrarySHA256 = lib, libraryHash(lib)
	entries, totalBytes := 0, 0
	err = fs.WalkDir(root.FS(), ".", func(p string, entry fs.DirEntry, walkErr error) error {
		if p == "." && walkErr != nil {
			return walkErr
		}
		if p == "." {
			return nil
		}
		entries++
		if entries > referenceMaxEntries || totalBytes >= referenceMaxBytes {
			result.Truncated = true
			return fs.SkipAll
		}
		if walkErr != nil {
			result.SkippedEntries++
			return nil
		}
		if !referencePath(p) || entry.Type()&os.ModeSymlink != 0 {
			result.SkippedEntries++
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		data, err := readReferenceFile(root, p)
		if err != nil {
			result.SkippedEntries++
			return nil
		}
		totalBytes += len(data)
		result.ScannedFiles++
		score := relevance(Revision{Summary: p, Body: string(data)}, query)
		if score > 0 {
			result.Matches = append(result.Matches, ReferenceMatch{Path: p, SHA256: referenceHash(data), Bytes: len(data), Score: score})
		}
		return nil
	})
	if err != nil {
		return result, errors.New("reference directory scan failed")
	}
	sort.Slice(result.Matches, func(a, b int) bool {
		if result.Matches[a].Score == result.Matches[b].Score {
			return result.Matches[a].Path < result.Matches[b].Path
		}
		return result.Matches[a].Score > result.Matches[b].Score
	})
	if len(result.Matches) > limit {
		result.Truncated = true
		result.Matches = result.Matches[:limit]
	}
	return result, nil
}
