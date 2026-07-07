package tokenusage

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// SectionIHeading is the canonical progress.md §I Token Accounting heading.
// It is a fresh top-level letter (§I) deliberately chosen so it does NOT
// collide with the era.go-parsed §E.* namespace (era.go matches only §E.2,
// §E.3, §E.4, §E.5 headings and the sync_commit_sha / mx_commit_sha fields).
const SectionIHeading = "## §I Token Accounting"

// BuildSectionI renders the §I Token Accounting section body from an
// Attribution. The output begins with the canonical §I heading followed by
// the nine token-accounting fields as `key: value` bullet lines, matching the
// schema documented in the SPEC progress.md §I placeholder. The returned
// string ends with a single trailing newline.
//
// Pure (no I/O): callers can assert on the rendered text directly, and
// WriteSectionI composes it with the read/apply/write step.
//
// @MX:NOTE: [AUTO] the nine field keys mirror the Attribution struct JSON tags (tokens_spent, token_attribution, ...) so the rendered section round-trips 1:1 with the struct the audit surface (M4) will parse
// @MX:SPEC: SPEC-TOKEN-ACCOUNTING-001
func BuildSectionI(attr Attribution) string {
	var b strings.Builder
	b.WriteString(SectionIHeading)
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "- tokens_spent: %d\n", attr.TokensSpent)
	fmt.Fprintf(&b, "- tokens_input: %d\n", attr.TokensInput)
	fmt.Fprintf(&b, "- tokens_output: %d\n", attr.TokensOutput)
	fmt.Fprintf(&b, "- tokens_cache_creation: %d\n", attr.TokensCacheCreation)
	fmt.Fprintf(&b, "- tokens_cache_read: %d\n", attr.TokensCacheRead)
	fmt.Fprintf(&b, "- cache_hit_ratio: %s\n", formatRatio(attr.CacheHitRatio))
	fmt.Fprintf(&b, "- token_attribution: %s\n", attr.AttributionMethod)
	fmt.Fprintf(&b, "- token_attribution_confidence: %s\n", attr.Confidence)
	fmt.Fprintf(&b, "- token_session_count: %d\n", attr.SessionCount)
	return b.String()
}

// WriteSectionI reads the progress.md at progressPath, replaces (or, if the
// §I section is absent, appends) the `## §I Token Accounting` section with a
// freshly built field block derived from attr, and writes the result back.
//
// Idempotent: a second call with the same attr overwrites the first cleanly.
// Scope-safe: it MUST NOT touch any other section (§E.*, §F, §G, §H, ...);
// only the span from the §I heading to the next top-level (## ) heading, or
// EOF, is replaced. Parser-safe by construction: the §I heading and field
// names are not among the tokens era.go greps (§E.{2,3,4,5} + the two SHA
// fields), so writing §I cannot change era classification (AC-TA-009).
//
// @MX:ANCHOR: [AUTO] progress.md §I sync-close writer — public write entry point consumed by manager-docs (sync-close) and the M4 audit surface
// @MX:REASON: [AUTO] the idempotent replace-or-append invariant + the scope-safe (sibling-section-preserving) invariant are the load-bearing contracts; breaking either silently corrupts progress.md structure or clobbers §E.* parser-load-bearing content
// @MX:SPEC: SPEC-TOKEN-ACCOUNTING-001
func WriteSectionI(progressPath string, attr Attribution) error {
	content, err := os.ReadFile(progressPath)
	if err != nil {
		return fmt.Errorf("tokenusage: read progress for §I write %q: %w", progressPath, err)
	}
	updated := applySectionI(string(content), attr)
	if err := os.WriteFile(progressPath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("tokenusage: write progress §I %q: %w", progressPath, err)
	}
	return nil
}

// applySectionI performs the idempotent in-memory replace-or-append of the §I
// section. It is separated from WriteSectionI so the transformation is
// exercisable without a filesystem path.
//
// Section span: from the canonical §I heading line to the next line beginning
// with "## " (a top-level heading) or EOF, inclusive. The span is replaced
// wholesale with BuildSectionI(attr); when no §I heading exists the section
// is appended at the end with a single blank separator line. Exactly one
// trailing newline is guaranteed.
func applySectionI(content string, attr Attribution) string {
	section := BuildSectionI(attr)
	lines := splitContentLines(content)
	sectionLines := splitContentLines(section)

	headingIdx := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == SectionIHeading {
			headingIdx = i
			break
		}
	}

	if headingIdx < 0 {
		// Append: ensure exactly one blank separator line before §I.
		out := append([]string{}, lines...)
		if len(out) == 0 || out[len(out)-1] != "" {
			out = append(out, "")
		}
		out = append(out, sectionLines...)
		return joinContentLines(out)
	}

	// Replace: find the next top-level heading after the §I heading, or EOF.
	end := len(lines)
	for i := headingIdx + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}

	out := append([]string{}, lines[:headingIdx]...)
	if len(out) > 0 && out[len(out)-1] != "" {
		out = append(out, "") // blank separator before §I heading
	}
	out = append(out, sectionLines...)
	if end < len(lines) {
		if out[len(out)-1] != "" {
			out = append(out, "") // blank separator after §I before next heading
		}
		out = append(out, lines[end:]...)
	}
	return joinContentLines(out)
}

// formatRatio renders a cache-hit ratio as a fixed 4-decimal string in [0,1].
// A fixed precision (rather than %v) keeps the rendered value grep-friendly
// and parseable by the downstream audit surface.
func formatRatio(r float64) string {
	return strconv.FormatFloat(r, 'f', 4, 64)
}

// splitContentLines splits s on "\n" and drops the single trailing empty
// element produced by a terminal newline, so the returned slice carries no
// trailing "". joinContentLines re-appends exactly one trailing newline.
func splitContentLines(s string) []string {
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// joinContentLines joins lines with "\n", collapses any run of trailing
// blank lines, and guarantees exactly one trailing newline (the conventional
// progress.md EOF shape).
func joinContentLines(lines []string) string {
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
