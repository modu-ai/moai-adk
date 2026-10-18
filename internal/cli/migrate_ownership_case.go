package cli

// migrate_ownership_case.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 (card t1865),
// REQ-OWR-007 and design.md D-4: a path that differs from a manifest key only in
// letter case.
//
// The question "are these two spellings one file?" is answered by observation,
// never by a file-system-mode probe: both spellings are stat-ed and compared with
// os.SameFile. On a case-insensitive file system they are the same file and the
// registered entry's verdict applies to it; on a case-sensitive one the second
// spelling does not exist or is another file, the comparison is false, and the
// two stay two files. The update never renames. (The probe probeCaseSensitiveFS
// writes into the project root, so it cannot be used on a path that must leave
// the tree unchanged under --dry-run, and no operational code calls it.)

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

// caseFS is the file-system view resolveCaseAlias judges identity with. It is a
// struct of two functions so a test can supply either case mode.
type caseFS struct {
	stat     func(name string) (fs.FileInfo, error)
	sameFile func(a, b fs.FileInfo) bool
}

// hostCaseFS is the operating file-system view. Like removalEnabled it is a
// package variable only so a test can inject a case mode; no operational code
// replaces it.
var hostCaseFS = caseFS{stat: os.Stat, sameFile: os.SameFile}

// resolveCaseAlias returns the single key in keys that differs from rel only in
// letter case (same directory) and names the same file as rel under root, as the
// given file-system view observes it. Two or more matching keys are ambiguous and
// give no alias.
func resolveCaseAlias(c caseFS, root, rel string, keys []string) (string, bool) {
	dir := path.Dir(rel)
	var relInfo fs.FileInfo
	match, matches := "", 0
	for _, k := range keys {
		if k == rel || path.Dir(k) != dir || !strings.EqualFold(k, rel) {
			continue
		}
		if relInfo == nil {
			info, err := c.stat(filepath.Join(root, filepath.FromSlash(rel)))
			if err != nil {
				return "", false
			}
			relInfo = info
		}
		other, err := c.stat(filepath.Join(root, filepath.FromSlash(k)))
		if err != nil || !c.sameFile(relInfo, other) {
			continue
		}
		match, matches = k, matches+1
	}
	if matches != 1 {
		return "", false
	}
	return match, true
}

// manifestKeysByDir indexes the manifest's keys by their directory, so the case
// alias lookup of one walked path touches only its own directory's keys.
func manifestKeysByDir(mgr manifest.Manager) map[string][]string {
	byDir := map[string][]string{}
	for k := range mgr.Manifest().Files {
		d := path.Dir(k)
		byDir[d] = append(byDir[d], k)
	}
	return byDir
}

// shippedCaseSpelling returns the shipped template path in the same directory as
// rel that differs from it only in letter case, or "" when rel is spelled as
// shipped, no such file ships, or the match is ambiguous.
func shippedCaseSpelling(embedded fs.FS, rel string) string {
	dir, base := path.Split(rel)
	entries, err := fs.ReadDir(embedded, strings.TrimSuffix(dir, "/"))
	if err != nil {
		return ""
	}
	match := ""
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == base {
			return ""
		}
		if strings.EqualFold(e.Name(), base) {
			if match != "" {
				return ""
			}
			match = e.Name()
		}
	}
	if match == "" {
		return ""
	}
	return dir + match
}

// canonicalCounterpartRel maps projectRel to the project path whose user-root
// counterpart key is spelled exactly as the user manifest records it, when the
// user manifest holds a key that differs from the counterpart's key only in
// letter case and names the same file.
func canonicalCounterpartRel(userManifest *userassets.Manifest, homeDir, projectRel string) (string, bool) {
	slug, rest, ok := userCounterpartLocation(projectRel)
	if !ok {
		return "", false
	}
	dir := userassets.RootBySlugDir(homeDir, slug)
	if dir == "" {
		return "", false
	}
	prefix := string(slug) + "/"
	var keys []string
	for k := range userManifest.Files {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, strings.TrimPrefix(k, prefix))
		}
	}
	alias, ok := resolveCaseAlias(hostCaseFS, dir, rest, keys)
	if !ok {
		return "", false
	}
	return strings.TrimSuffix(projectRel, rest) + alias, true
}

// counterpartConfirmedCaseAware is userCounterpartConfirmed with one addition: a
// project spelling that the user manifest and the shipped tree spell differently
// in letter case, and that names the same file on the observed file system, is
// confirmed against the user manifest's spelling.
func counterpartConfirmedCaseAware(userManifest *userassets.Manifest, cat *template.Catalog, embedded fs.FS, homeDir, projectRel string) bool {
	if userCounterpartConfirmed(userManifest, cat, embedded, homeDir, projectRel) {
		return true
	}
	canon, ok := canonicalCounterpartRel(userManifest, homeDir, projectRel)
	return ok && userCounterpartConfirmed(userManifest, cat, embedded, homeDir, canon)
}
