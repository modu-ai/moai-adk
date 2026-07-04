package spec

import (
	"os"
	"path/filepath"
	"sort"
)

// SPEC-WEB-CONSOLE-011 M5 (REQ-WC11-041) — exported, read-only catalog lister.
//
// ListDocs is the exported wrapper around the unexported discoverSPECs +
// parseSPECDoc helpers. It powers the READ-ONLY web SPEC board: it discovers the
// SPEC documents under a project root and returns their parsed frontmatter
// records WITHOUT the heavy Body/Criteria/REQ payload of SPECDoc. Like every
// other rule/parser in this package it is observation-only — it MUST NOT mutate
// any file.

// DocRecord is a per-SPEC frontmatter record returned by ListDocs. It carries the
// source path, the parsed frontmatter, and any per-document parse error (surfaced
// per-record so one malformed spec.md never aborts the whole scan).
type DocRecord struct {
	Path        string
	Frontmatter SPECFrontmatter
	ParseError  error
}

// ListDocs discovers every SPEC document under baseDir/.moai/specs/SPEC-*/spec.md
// and returns their parsed frontmatter records, sorted by path (i.e. by SPEC-ID
// directory name) for deterministic output.
//
// baseDir is the PROJECT ROOT; the ".moai/specs" suffix is appended internally,
// matching the Audit() BaseDir convention so a single project-root value can feed
// both ListDocs and Audit. An empty baseDir defaults to ".". A missing specs
// directory returns an empty slice (not an error) so a fresh project renders an
// empty board rather than failing.
func ListDocs(baseDir string) ([]DocRecord, error) {
	if baseDir == "" {
		baseDir = "."
	}
	specsDir := filepath.Join(baseDir, ".moai", "specs")

	// Graceful empty result when the specs directory does not exist yet — mirrors
	// Audit()'s os.IsNotExist handling so a fresh project is not an error.
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		return []DocRecord{}, nil
	}

	paths, err := discoverSPECs(specsDir)
	if err != nil {
		return nil, err
	}

	records := make([]DocRecord, 0, len(paths))
	for _, p := range paths {
		doc := parseSPECDoc(p)
		records = append(records, DocRecord{
			Path:        p,
			Frontmatter: doc.Frontmatter,
			ParseError:  doc.ParseError,
		})
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Path < records[j].Path })
	return records, nil
}
