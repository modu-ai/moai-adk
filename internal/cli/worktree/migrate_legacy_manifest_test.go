package worktree

// migrate_legacy_manifest_test.go — card t2053: the manifest algorithm of base
// 67b4734d6b (full SHA-256 of every untracked file INCLUDING ignored ones, the
// pre-IgnoredDigest JSON shape), copied verbatim so a journal written by that
// build can be reproduced exactly. Only the identifiers are renamed.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type legacyTreeManifestT2053 struct {
	HEAD          string
	Branch        string
	Unpushed      string
	Status        string
	Diff          string
	IgnoredCount  int
	ReportsCount  int
	ContentDigest string
}

// legacyCaptureManifestT2053 digests the working state of the tree at path through the
// full AC-WHR-017a manifest.
func legacyCaptureManifestT2053(tree string) string {
	m := legacyCaptureTreeManifestT2053(tree)
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// legacyCaptureTreeManifestT2053 reads the AC-WHR-017a manifest fields of the tree at
// path. Read failures degrade to zero values inside the manifest — the
// pre/post comparison then notices the change of shape, exactly as an
// unreadable field should fail a field-for-field comparison.
func legacyCaptureTreeManifestT2053(tree string) legacyTreeManifestT2053 {
	var m legacyTreeManifestT2053
	m.HEAD, _ = listGitOutput(tree, "rev-parse", "HEAD")
	m.HEAD = strings.TrimSpace(m.HEAD)
	m.Branch, _ = listGitOutput(tree, "rev-parse", "--abbrev-ref", "HEAD")
	m.Branch = strings.TrimSpace(m.Branch)
	m.Unpushed, _ = listGitOutput(tree, "rev-list", "HEAD", "--not", "--remotes")
	m.Status, _ = listGitOutput(tree, "status", "--porcelain")
	// The tracked-change content (REQ-WHR-017's content digest over the
	// tracked-change diff — card t1640 review F7): a content mutation that
	// preserves the status shape must change the manifest. Both the
	// unstaged and the staged forms ride the manifest.
	staged, _ := listGitOutput(tree, "diff", "--cached")
	unstaged, _ := listGitOutput(tree, "diff")
	m.Diff = staged + unstaged

	var untracked, ignored []string
	if out, err := listGitOutput(tree, "ls-files", "--others", "-z"); err == nil {
		untracked = splitNul(out)
	}
	if out, err := listGitOutput(tree, "ls-files", "--others", "--ignored", "--exclude-standard", "-z"); err == nil {
		ignored = splitNul(out)
	}
	m.IgnoredCount = len(ignored)

	seen := map[string]bool{}
	var paths []string
	for _, p := range append(append([]string{}, untracked...), ignored...) {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		paths = append(paths, p)
	}
	sort.Strings(paths)
	digest := sha256.New()
	for _, p := range paths {
		_, _ = digest.Write([]byte(p))
		_, _ = digest.Write([]byte{0})
		full := filepath.Join(tree, filepath.FromSlash(p))
		if info, lerr := os.Lstat(full); lerr == nil {
			// Symlinks hash their TARGET (card t1640 review F11): os.ReadFile
			// would follow the link, and a FIFO or device would block the
			// apply while the migration lock is held. Only regular files are
			// read; everything else hashes its type marker.
			switch {
			case info.Mode()&os.ModeSymlink != 0:
				if target, rerr := os.Readlink(full); rerr == nil {
					_, _ = digest.Write([]byte("link:" + target))
				} else {
					_, _ = digest.Write([]byte("\x00unreadable-link"))
				}
			case info.Mode().IsRegular():
				if body, rerr := os.ReadFile(full); rerr == nil {
					_, _ = digest.Write(body)
				} else {
					_, _ = digest.Write([]byte("\x00unreadable"))
				}
			default:
				_, _ = digest.Write([]byte("special:" + info.Mode().String()))
			}
		} else {
			_, _ = digest.Write([]byte("\x00unreadable"))
		}
		_, _ = digest.Write([]byte{0})
	}
	m.ContentDigest = hex.EncodeToString(digest.Sum(nil))

	reports := filepath.Join(tree, ".moai", "reports")
	_ = filepath.WalkDir(reports, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		m.ReportsCount++
		return nil
	})
	return m
}
