package template

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

// ComputeDirTreeHash hashes a catalog directory entry as a whole tree: every
// regular file under dir is read, normalized (NormalizeForHash), and hashed;
// each file contributes its slash path relative to dir plus its digest; the
// entries are sorted by path and the joined list is hashed once more.
//
// Deployed through //go:embed all:templates, everything under a skill
// directory reaches the user — but the v1 catalog hash covered SKILL.md only,
// so the sub-files (modules/, references/, scripts/, workflows/, schemas/, ...)
// were delivered while sitting entirely outside the catalog's integrity claim
// (card t323). This function gives the claim and the deployment the same
// boundary.
//
// Both consumers of the catalog hash share this one implementation: the
// generator (scripts/gen-catalog-hashes.go, over os.DirFS of the templates
// tree) and the audit test (TestManifestHashFormat in catalog_tier_audit_test.go,
// over the embedded FS) — so "what is hashed" has a single answer that cannot
// drift into a pair.
func ComputeDirTreeHash(fsys fs.FS, dir string) (string, error) {
	type entry struct {
		rel string
		sum string
	}
	var entries []entry
	walkErr := fs.WalkDir(fsys, dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !d.Type().IsRegular() {
			return nil
		}
		raw, readErr := fs.ReadFile(fsys, p)
		if readErr != nil {
			return fmt.Errorf("read %q: %w", p, readErr)
		}
		sum := sha256.Sum256(NormalizeForHash(raw))
		rel := strings.TrimPrefix(p, dir+"/")
		entries = append(entries, entry{rel: rel, sum: hex.EncodeToString(sum[:])})
		return nil
	})
	if walkErr != nil {
		return "", fmt.Errorf("walk %q: %w", dir, walkErr)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].rel < entries[j].rel })
	h := sha256.New()
	for _, e := range entries {
		fmt.Fprintf(h, "%s:%s\n", e.rel, e.sum)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
