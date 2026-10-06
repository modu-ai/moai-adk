// Package embedemit emits the template embed allowlist manifest
// (internal/template/embed_manifest_gen.go, card t1539).
//
// //go:embed all:templates embeds every file on disk under templates/,
// including git-ignored and untracked additions — the boundary defect behind
// the 2026-07-07 trace-*.jsonl incident and the August 2026 preservation set
// (config-cache.json, github/counts.json, handoff/pending.json). This
// package replaces that directive with one explicit //go:embed pattern per
// approved file, emitted from the git-tracked file set: a file on disk that
// is not tracked is by definition not approved to ship, and no longer
// reaches the binary even under a direct `go build`.
//
// Regeneration stays behind the explicit `make embed-manifest` verb
// (EMBED_MANIFEST_UPDATE=1 golden test); `make embed-manifest-check` judges
// the committed manifest read-only ahead of every build, in the same
// position as agents-emit-check.
package embedemit

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strings"
)

// templatesDir is the tracked template tree, repo-root relative.
const templatesDir = "internal/template/templates"

// embedPrefix is the directive path prefix shared by every boundary entry,
// relative to the internal/template package directory.
const embedPrefix = "templates/"

// catalogDirective pins the catalog.yaml sibling into the same embed.FS.
// catalog.yaml lives outside templates/ (SPEC-V3R4-CATALOG-001 T-023), so it
// keeps its own fixed directive regardless of the file set.
const catalogDirective = "//go:embed catalog.yaml"

// ApprovedPaths returns the embed allowlist: every git-tracked file under
// internal/template/templates/ in repoRoot, as go:embed patterns relative to
// the internal/template package directory ("templates/..."), sorted. The
// tracked file set is the shipment contract — untracked and git-ignored
// files are by definition not approved.
func ApprovedPaths(repoRoot string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoRoot, "ls-files", "-z", "--", templatesDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list tracked template files: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	var paths []string
	seen := make(map[string]struct{})
	for p := range strings.SplitSeq(string(out), "\x00") {
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, templatesDir+"/") {
			continue
		}
		pattern := embedPrefix + strings.TrimPrefix(p, templatesDir+"/")
		if err := validatePattern(pattern); err != nil {
			return nil, fmt.Errorf("tracked template file %q: %w", p, err)
		}
		if _, dup := seen[pattern]; dup {
			continue
		}
		seen[pattern] = struct{}{}
		paths = append(paths, pattern)
	}
	sort.Strings(paths)
	return paths, nil
}

// DirectiveRe matches one //go:embed directive line in a rendered manifest.
var DirectiveRe = regexp.MustCompile(`(?m)^//go:embed (.+)$`)

// ParseDirectives extracts the go:embed patterns from a rendered manifest,
// in emission order. The generator emits exactly one pattern per directive;
// any other shape is a generator-contract violation, not something to parse
// around. Both the emitter's golden test and the consumer-side boundary
// audit parse through this one function so the directive grammar has a
// single reading.
func ParseDirectives(raw []byte) ([]string, error) {
	var out []string
	for _, m := range DirectiveRe.FindAllStringSubmatch(string(raw), -1) {
		fields := strings.Fields(m[1])
		if len(fields) != 1 {
			return nil, fmt.Errorf("directive %q carries %d patterns; the manifest emits exactly one path per line", m[1], len(fields))
		}
		out = append(out, fields[0])
	}
	return out, nil
}

// validatePattern rejects patterns go:embed cannot carry or that would break
// the boundary contract: the emitter writes one explicit path per directive,
// so anything but a single clean slash-separated path under templates/ is a
// generator-contract violation, not a pattern to escape.
func validatePattern(p string) error {
	switch {
	case p == "":
		return fmt.Errorf("empty pattern")
	case !strings.HasPrefix(p, embedPrefix):
		return fmt.Errorf("pattern %q is outside %q", p, embedPrefix)
	case strings.Contains(p, ".."):
		return fmt.Errorf("pattern %q contains a parent-directory element", p)
	case strings.ContainsRune(p, '\\'):
		return fmt.Errorf("pattern %q contains a backslash; go:embed patterns are slash-separated", p)
	case strings.ContainsAny(p, " :*?\"<>|[]{}"):
		return fmt.Errorf("pattern %q contains a character go:embed patterns cannot carry or interpret as a glob (brackets are character classes, braces are alternation)", p)
	}
	if path.Clean(p) != p {
		return fmt.Errorf("pattern %q is not in clean form", p)
	}
	return nil
}

// Render returns the bytes of embed_manifest_gen.go for the given paths.
// The output is deterministic: paths are validated, deduplicated and sorted,
// and no timestamp or environment value is rendered.
func Render(paths []string) ([]byte, error) {
	seen := make(map[string]struct{}, len(paths))
	unique := make([]string, 0, len(paths))
	for _, p := range paths {
		if err := validatePattern(p); err != nil {
			return nil, err
		}
		if _, dup := seen[p]; dup {
			continue
		}
		seen[p] = struct{}{}
		unique = append(unique, p)
	}
	sort.Strings(unique)

	var b strings.Builder
	b.WriteString(`// Code generated by internal/template/embedemit; DO NOT EDIT.
//
// The //go:embed directives below are the embed ALLOWLIST (card t1539):
// one explicit per-file pattern per approved template file, emitted from
// the git-tracked file set under internal/template/templates/. A file on
// disk that is not listed here is NOT embedded — an ignored or untracked
// addition can no longer reach the shipped binary, even under a direct
// go build.
//
// Regenerate with ` + "`make embed-manifest`" + `. ` + "`make embed-manifest-check`" + `
// (wired ahead of build) judges this file against a fresh emission and the
// templates/ tree.

package template

import "embed"

// embeddedRaw holds the raw embedded filesystem with the "templates/"
// prefix: one explicit directive per approved file — the allowlist boundary
// (see internal/template/embed.go for the accessor) — plus the catalog.yaml
// sibling.
//
`)
	b.WriteString(catalogDirective)
	b.WriteString("\n")
	for _, p := range unique {
		b.WriteString("//go:embed ")
		b.WriteString(p)
		b.WriteString("\n")
	}
	b.WriteString("var embeddedRaw embed.FS\n")
	return []byte(b.String()), nil
}
