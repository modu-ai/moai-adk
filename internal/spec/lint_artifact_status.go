package spec

// lint_artifact_status.go — ArtifactStatusFieldForbiddenRule
// (SPEC-ARTIFACT-STATELESS-001 M2, REQ-AST-001-004..007). An `error`-severity
// SPEC-artifact lint rule that rejects a `status:` field in the YAML
// frontmatter of a non-spec.md SPEC artifact.
//
// The declaration it enforces: the canonical 12-field obligation binds
// `spec.md` only, and the four sibling artifacts — plan.md, acceptance.md,
// design.md, research.md — are stateless ON THE STATUS AXIS
// (`.claude/rules/moai/development/spec-frontmatter-schema.md`
// § Artifact Statelessness). A SPEC's lifecycle state lives in exactly one
// file, which is the file every lint, audit, and close path reads it from
// (`audit.go` reads spec.md; `discoverSPECs` globs SPEC-*/spec.md;
// `closer.go` writes spec.md + progress.md).
//
// THE AXIS IS STATUS, NOT FRONTMATTER. Frontmatter itself stays permitted in
// those four artifacts; `id`, `title`, `version`, and `created` are outside
// this rule entirely. Widening the predicate to "frontmatter forbidden" would
// re-create the very mismatch this SPEC exists to close — the corpus cleanup
// (D1) removes the `status:` line and nothing else, so a checker that reads a
// wider axis than the cleanup writes would fire on files the cleanup left
// deliberately intact.
//
// The predicate tracks the cleanup's own shell predicate (acceptance.md
// snippet (B)):
//
//	awk 'NR==1&&/^---/{p=1;next} p&&/^---/{exit} p' "$f" | grep -qE '^status:[[:space:]]'
//
// Two properties of that shell form are load-bearing and are reproduced exactly
// rather than "cleaned up" in Go:
//
//  1. The frontmatter block opens ONLY at line 1. A `---` appearing later in
//     the body is body text, not a frontmatter fence.
//  2. `/^---/` is a PREFIX match, so a `----` rule line closes the block.
//
// The third property is deliberately NOT reproduced. `^status:[[:space:]]`
// requires whitespace after the colon, so `status:draft` and a bare `status:`
// escape it; this rule matches the `status:` prefix alone, making it a strict
// SUPERSET of the counting predicate. The widening is safe in exactly one
// direction and measured, not assumed: the corpus carries ZERO such lines in
// the four governed artifacts (`grep -rnE '^status:[^ \t]' --include=plan.md
// --include=acceptance.md --include=design.md --include=research.md .moai/specs/`
// → no output, run at the M2 HEAD), so the two predicates select the same set
// today. They can diverge only on a line a future author writes, which is the
// case a RECURRENCE-prevention rule exists to catch.
//
// The cleanup carries the same widened predicate, which is what keeps this
// safe. A checker stricter than its cleanup is the dangerous direction: it
// leaves residual findings the cleanup cannot clear, reddening the corpus the
// moment the rule lands — the outcome REQ-AST-001-010 (same-SPEC landing, no
// era carve-out) exists to prevent.
//
// Severity is `error`, and the code is deliberately ABSENT from
// `eraDemotableCodes` (REQ-AST-001-006). That absence is a design decision,
// not an omission, and its precondition is that the D1 cleanup lands in the
// same SPEC: with corpus violations at 0 on landing, `error` gates nothing
// retroactively. If the cleanup is ever split into a separate card, this
// decision inverts and the code MUST be era-demoted — plan.md §B2 records the
// pair.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// statelessArtifacts is the closed set of non-spec.md SPEC artifacts governed
// by the statelessness declaration (REQ-AST-001-004). The list is fixed rather
// than derived from a directory scan: `spec.md` carries the canonical status
// field per the schema, and `progress.md` records phase progress in body
// sections rather than frontmatter — both sit outside this rule by
// REQ-AST-001-005/011, and enumerating the governed set is what keeps them out
// by construction instead of by an exclusion that could be forgotten.
var statelessArtifacts = []string{"plan.md", "acceptance.md", "design.md", "research.md"}

// ArtifactStatusFieldForbiddenRule implements the Rule interface for
// REQ-AST-001-004.
//
// It reads SIBLING artifacts rather than `doc.Body`: `SPECDoc` carries spec.md
// alone, and spec.md is precisely the file this rule does NOT govern. The
// precedent for a rule reaching past the single parsed document is
// MovingRefUnpinnedRule (same per-SPEC shape, so `lint.skip` and era demotion
// both continue to apply) and, before it, HaikuResidualRule.
type ArtifactStatusFieldForbiddenRule struct{}

// Code returns the stable lint finding code.
func (r *ArtifactStatusFieldForbiddenRule) Code() string { return "ArtifactStatusFieldForbidden" }

// Check scans the four governed artifacts in the SPEC's own directory.
func (r *ArtifactStatusFieldForbiddenRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	if doc == nil || doc.Path == "" {
		return nil
	}
	dir := filepath.Dir(doc.Path)

	var findings []Finding
	for _, name := range statelessArtifacts {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path) // #nosec G304 -- path is derived from the discovered SPEC directory
		if err != nil {
			continue // absent artifact is the dominant corpus shape, not a defect
		}
		line, value, found := frontmatterStatusLine(strings.Split(string(data), "\n"))
		if !found {
			continue
		}
		findings = append(findings, Finding{
			File:     path,
			Line:     line,
			Severity: SeverityError,
			Code:     r.Code(),
			Message: fmt.Sprintf(
				"`%s` carries `status: %s` in its frontmatter. Non-`spec.md` SPEC artifacts are stateless on the status axis: "+
					"a SPEC's lifecycle state lives in `spec.md` alone, which is the file every lint, audit, and close path reads it from. "+
					"Remove the `status:` line; every other frontmatter field (`id`, `title`, `version`, `created`) stays — "+
					"frontmatter itself is permitted here, and this rule governs the status axis only. "+
					"See `.claude/rules/moai/development/spec-frontmatter-schema.md` § Artifact Statelessness.",
				name, value),
		})
	}
	return findings
}

// frontmatterStatusLine reports the 1-based line number and the value of the
// first `status:` field inside the document's leading frontmatter block.
//
// It reproduces the cleanup's block-scoping rules exactly and widens only the
// field match — see the file header for why each property is load-bearing, and
// for the measurement showing the widening selects the same corpus set today.
// Desynchronizing the checker from the cleanup in the OTHER direction (checker
// stricter) is the failure this rule's whole design is arranged around.
func frontmatterStatusLine(lines []string) (line int, value string, found bool) {
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "---") {
		return 0, "", false // property 1: the block opens only at line 1
	}
	for i := 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "---") {
			return 0, "", false // property 2: prefix match closes the block
		}
		rest, ok := strings.CutPrefix(lines[i], "status:")
		if !ok {
			continue
		}
		// Widened past the counting predicate on purpose: no whitespace is
		// required after the colon, so `status:draft` is caught too.
		return i + 1, strings.TrimSpace(rest), true
	}
	return 0, "", false
}
