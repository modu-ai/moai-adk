// template_neutrality_audit_test.go: CI guard for
// SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001 (REQ-TNA-009).
//
// Scans internal/template/templates/** for developer-local / dev-incident
// traces that would leak into 16-language user-distributed templates. This
// audit is scoped to the NEUTRALITY-unique "kept" classes ONLY:
//
//	C1 macOS-bias absolute path placeholder (/Users/)        — binary FAIL
//	C2 bare-narrative V3R[0-9] version sigil                 — advisory WARN
//	C4 feedback_ / memory.md substring reference             — advisory WARN
//	C5 CLAUDE.local.md maintainer-only local file reference  — binary FAIL
//	C6 PR #N specific pull-request number reference          — binary FAIL
//	C8 GOOS=<os> Go cross-compile env var (false positive)   — PRESERVE
//
// [SCOPE — disjoint from internal_content_leak_test.go]
// The date class (C3) and commit-hash class (C7) are DEFERRED to the sibling
// SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001, whose
// internal_content_leak_test.go strict-tier S1-internal-date /
// S2-short-sha-sentence-final classes own them. This audit deliberately does
// NOT scan C3/C7 — re-scanning would create two test files in the same Go
// package with two divergent allow-lists for one pattern class (the
// dual-allow-list drift the v0.1.2 rescope eliminated). The pattern set here
// is therefore DISJOINT from internal_content_leak_test.go: no class is
// enforced by both files.
//
// [C2 narrowed to bare-narrative — v0.1.2]
// The broad V3R[0-9] regex collides with the SPEC-ID / CONST-registry-ID /
// REQ-ID domain owned by internal_content_leak_test.go's C1-spec-id class.
// C2 here owns ONLY the bare-narrative version sigil (a V3R[0-9] NOT part of a
// larger SPEC-/CONST-/REQ- identifier). Go's regexp (RE2) lacks lookbehind, so
// bare-narrative detection is an explicit TWO-PASS exclusion: Pass 1 matches
// \bV3R[0-9]; Pass 2 drops hits whose immediately-preceding token is an
// ID-prefix (SPEC-/CONST-/REQ-) or whose preceding rune continues an
// identifier ([A-Za-z0-9-]). This produces the same file set as the perl PCRE
// negative-lookbehind form `grep -rlP '(?<![A-Za-z0-9-])V3R[0-9]'`.
//
// Sentinel on failure: TEMPLATE_NEUTRALITY_VIOLATION (binary) /
// TEMPLATE_NEUTRALITY_WARN (advisory).
//
// Verified in ISOLATION via:
//
//	go test ./internal/template/... -run TestTemplateNeutralityAudit
//
// Package-wide green is NOT a precondition (the internal/template package has
// pre-existing failures unrelated to this SPEC — see spec.md §3.4). This test
// is designed to PASS standalone.
package template_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// neutralityTemplatesRoot is the scan root, relative to this package
// directory (internal/template). A distinct name avoids collision with the
// package-internal `templatesRoot` const in internal_content_leak_test.go
// (which is in `package template`, while this file is `package template_test`).
const neutralityTemplatesRoot = "templates"

// neutralitySeverity classifies an audit class.
type neutralitySeverity int

const (
	neutralityBinary  neutralitySeverity = iota // 1+ hit outside allow-list → test FAIL
	neutralityWarn                              // hit > allow-list → t.Logf WARN, test passes
	neutralityExclude                           // false positive — recorded, never a violation
)

// neutralityClass is one kept-category detector.
type neutralityClass struct {
	name     string
	severity neutralitySeverity
	pattern  *regexp.Regexp
	// allowList enumerates relative paths (forward-slash, under
	// neutralityTemplatesRoot) permitted to contain hits. For binary classes
	// with an empty allow-list, ANY hit is a violation.
	allowList map[string]struct{}
}

// neutralityScannedExts are the text formats that ship verbatim to user
// projects and are therefore the documented leak surface.
var neutralityScannedExts = map[string]struct{}{
	".md":   {},
	".tmpl": {},
	".yaml": {},
	".yml":  {},
	".sh":   {},
	".json": {},
}

// c2BarePass1 finds all V3R[0-9] candidates (RE2-compatible word boundary).
var c2BarePass1 = regexp.MustCompile(`\bV3R[0-9]`)

// c2IDEmbeddedPrefixes are the ID-prefix forms that, when they immediately
// precede a V3R[0-9] hit, mark it as ID-embedded (owned by the leak test's
// C1-spec-id class) rather than bare-narrative. The Pass-2 exclusion also
// drops any hit whose preceding rune is in [A-Za-z0-9-] (continues an
// identifier token).
var c2IDEmbeddedRe = regexp.MustCompile(`(SPEC|CONST|REQ|AC|BC|HARNESS|PROJECT|HL)-V3R[0-9]`)

// c8GoosRe is the C8 false-positive PRESERVE pattern. Hits matching it are
// excluded from C2/etc. consideration on that span (and recorded for the
// C8 preservation subtest).
var c8GoosRe = regexp.MustCompile(`GOOS=(linux|windows|darwin|freebsd|openbsd|netbsd)`)

// allowListSet builds a relative-path set from a slice.
func allowListSet(paths ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		m[p] = struct{}{}
	}
	return m
}

// neutralityClasses defines the kept-class detectors. C2/C4 allow-lists are
// the migration-matrix.md §C2/§C4 file-path enumerations (mirrored here as Go
// constants per REQ-TNA-009 "Go test reads .md or duplicated Go constant").
// Keys are relative paths under neutralityTemplatesRoot (forward-slash).
var neutralityClasses = []neutralityClass{
	{
		name:      "C1-macos-bias-path",
		severity:  neutralityBinary,
		pattern:   regexp.MustCompile(`/Users/`),
		allowList: allowListSet(), // empty — binary FAIL on any hit
	},
	{
		name:     "C2-bare-narrative-v3r",
		severity: neutralityWarn,
		// pattern is handled specially (two-pass) in scanC2BareNarrative;
		// this field is the Pass-1 candidate matcher.
		pattern: c2BarePass1,
		allowList: allowListSet(
			".claude/rules/moai/core/zone-registry.md",
			".claude/agents/moai/manager-spec.md",
			".claude/skills/moai-harness-learner/SKILL.md",
			".claude/skills/moai/SKILL.md",
			".claude/skills/moai/workflows/harness.md",
			".claude/skills/moai-meta-harness/SKILL.md",
		),
	},
	{
		name:     "C4-feedback-memory-ref",
		severity: neutralityWarn,
		pattern:  regexp.MustCompile(`feedback_|memory\.md`),
		allowList: allowListSet(
			"CLAUDE.md",
			".claude/rules/moai/workflow/context-window-management.md",
			".claude/rules/moai/workflow/session-handoff.md",
			".claude/rules/moai/workflow/spec-workflow.md",
			".claude/rules/moai/workflow/worktree-state-guard.md",
			".claude/rules/moai/workflow/verification-batch-pattern.md",
			".claude/skills/moai/team/run.md",
		),
	},
	{
		name:      "C5-claude-local-ref",
		severity:  neutralityBinary,
		pattern:   regexp.MustCompile(`CLAUDE\.local\.md`),
		allowList: allowListSet(), // empty — binary FAIL on any hit
	},
	{
		name:      "C6-pr-number-ref",
		severity:  neutralityBinary,
		pattern:   regexp.MustCompile(`PR #[0-9]+`),
		allowList: allowListSet(), // empty — binary FAIL on any hit
	},
}

// findNeutralityRoot locates the scan root by walking upward from the current
// working directory until a "templates" subdir alongside go.mod is found. This
// mirrors the project-root discovery used by rule_template_mirror_test.go while
// keeping this test self-contained (no shared package-internal helper).
func findNeutralityRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		candidate := filepath.Join(dir, neutralityTemplatesRoot)
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			// Confirm this is the internal/template/templates dir by checking
			// for a known top-level template artifact.
			if _, claudeErr := os.Stat(filepath.Join(candidate, ".claude")); claudeErr == nil {
				return candidate
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("templates root not found from cwd; expected internal/template/templates")
		}
		dir = parent
	}
}

// scanC2BareNarrative returns the relative paths (under root) that contain at
// least one bare-narrative V3R[0-9] hit — i.e. a V3R[0-9] that is NOT part of a
// larger SPEC-/CONST-/REQ-/identifier token. Implements the two-pass exclusion
// described in the file header.
func scanC2BareNarrative(root string) (map[string]struct{}, error) {
	hits := make(map[string]struct{})
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if _, ok := neutralityScannedExts[filepath.Ext(path)]; !ok {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		text := string(content)
		// Pass 1: locate all V3R[0-9] candidates by byte offset.
		for _, loc := range c2BarePass1.FindAllStringIndex(text, -1) {
			start := loc[0]
			// Pass 2a: drop ID-embedded hits (SPEC-V3R, CONST-V3R, REQ-V3R, ...).
			// Inspect a small left window for an ID-prefix form ending at start.
			windowStart := start - 16
			if windowStart < 0 {
				windowStart = 0
			}
			window := text[windowStart : loc[1]]
			if c2IDEmbeddedRe.MatchString(window) {
				continue
			}
			// Pass 2b: drop hits whose immediately-preceding rune continues an
			// identifier token ([A-Za-z0-9-]).
			if start > 0 {
				prev := text[start-1]
				if (prev >= 'A' && prev <= 'Z') ||
					(prev >= 'a' && prev <= 'z') ||
					(prev >= '0' && prev <= '9') ||
					prev == '-' {
					continue
				}
			}
			// Bare-narrative hit confirmed.
			rel := relUnderRoot(root, path)
			hits[rel] = struct{}{}
			break // one bare hit per file is enough to flag the file
		}
		return nil
	})
	return hits, err
}

// relUnderRoot returns the forward-slash relative path of path under root.
func relUnderRoot(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

// scanSimpleClass returns the relative paths containing at least one hit of the
// class pattern, EXCLUDING spans covered by the C8 GOOS false-positive pattern.
func scanSimpleClass(root string, class neutralityClass) (map[string]struct{}, error) {
	hits := make(map[string]struct{})
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if _, ok := neutralityScannedExts[filepath.Ext(path)]; !ok {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		text := string(content)
		locs := class.pattern.FindAllStringIndex(text, -1)
		if len(locs) == 0 {
			return nil
		}
		// C8 exclusion: drop any class hit that overlaps a GOOS=<os> span.
		// (None of C1/C5/C6 patterns can match inside a GOOS= span, so this is
		// a defensive no-op for those classes; retained for symmetry with the
		// REQ-TNA-008 PRESERVE contract.)
		goosSpans := c8GoosRe.FindAllStringIndex(text, -1)
		for _, loc := range locs {
			if spanOverlapsAny(loc, goosSpans) {
				continue
			}
			rel := relUnderRoot(root, path)
			hits[rel] = struct{}{}
			break
		}
		return nil
	})
	return hits, err
}

// spanOverlapsAny reports whether span [a[0],a[1]) overlaps any span in others.
func spanOverlapsAny(a []int, others [][]int) bool {
	for _, o := range others {
		if a[0] < o[1] && o[0] < a[1] {
			return true
		}
	}
	return false
}

// TestTemplateNeutralityAudit enforces the kept-class neutrality contract over
// the user-distributed template tree. Binary classes (C1/C5/C6) FAIL on any
// hit outside their allow-list; advisory classes (C2/C4) emit a WARN log when
// hits exceed the allow-list size but do NOT fail the test.
//
// Sentinel: TEMPLATE_NEUTRALITY_VIOLATION (binary) / TEMPLATE_NEUTRALITY_WARN.
func TestTemplateNeutralityAudit(t *testing.T) {
	t.Parallel()

	root := findNeutralityRoot(t)

	for _, class := range neutralityClasses {
		class := class
		t.Run(class.name, func(t *testing.T) {
			t.Parallel()

			var hits map[string]struct{}
			var err error
			if class.name == "C2-bare-narrative-v3r" {
				hits, err = scanC2BareNarrative(root)
			} else {
				hits, err = scanSimpleClass(root, class)
			}
			if err != nil {
				t.Fatalf("scan error for %s: %v", class.name, err)
			}

			// Partition hits into allow-listed vs violations.
			var violations []string
			for rel := range hits {
				if _, ok := class.allowList[rel]; ok {
					continue
				}
				violations = append(violations, rel)
			}

			switch class.severity {
			case neutralityBinary:
				if len(violations) > 0 {
					for _, v := range violations {
						t.Errorf("TEMPLATE_NEUTRALITY_VIOLATION: class=%s file=%s "+
							"(binary class — not permitted in distributed templates)",
							class.name, v)
					}
				}
			case neutralityWarn:
				// Advisory: hits beyond the allow-list are a WARN, not a FAIL.
				// AC-TNA-002 / AC-TNA-004 require actual ≤ allow-list size, so we
				// also enforce the count ceiling as a soft assertion via WARN.
				if len(violations) > 0 {
					for _, v := range violations {
						t.Logf("TEMPLATE_NEUTRALITY_WARN: class=%s file=%s "+
							"(advisory — exceeds migration-matrix allow-list)",
							class.name, v)
					}
				}
				// Belt-and-suspenders: the run-phase reduced both advisory
				// classes to exactly their allow-list, so any residual
				// violation indicates regression. Surface it as a WARN log only
				// (CI green with annotations per REQ-TNA-010), NOT a FAIL.
			}
		})
	}
}

// TestTemplateNeutralityAuditC8Preserve verifies the C8 GOOS=<os> false-positive
// PRESERVE contract (REQ-TNA-008 / AC-TNA-011): the Go cross-compile env var
// MUST be preserved in the template tree AND MUST NOT be emitted as a
// neutrality violation. Exactly 3 files carry the GOOS= substring.
func TestTemplateNeutralityAuditC8Preserve(t *testing.T) {
	t.Parallel()

	root := findNeutralityRoot(t)

	preserved := make(map[string]struct{})
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if _, ok := neutralityScannedExts[filepath.Ext(path)]; !ok {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if c8GoosRe.Match(content) {
			preserved[relUnderRoot(root, path)] = struct{}{}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("C8 scan error: %v", err)
	}

	if len(preserved) != 3 {
		var files []string
		for f := range preserved {
			files = append(files, f)
		}
		t.Errorf("C8 GOOS= PRESERVE expected 3 files, got %d: %s",
			len(preserved), strings.Join(files, ", "))
	}
}
