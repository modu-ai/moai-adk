package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// The citations axis (SPEC-CODEMAPS-ACCURACY-001 REQ-CMA-002, design D1):
// every positive citation of a source path in the codemaps docs must name a
// path that exists in the working tree. LayerCitations and
// MetricPositiveCitedPathAbsence are declared in check.go's constant blocks
// alongside the three freshness layers' tokens.

// citedPathPattern matches a source-path citation token: an `internal/`,
// `pkg/`, or `cmd/` prefixed path, matching the t432 full-census extraction
// shape. Code fences and mermaid blocks are scanned under the same rule —
// citations live inside them too (the ListActive mermaid node was a real
// instance) — only blockquote lines are exempt (D2).
var citedPathPattern = regexp.MustCompile(`\b(?:internal|pkg|cmd)/[A-Za-z0-9_/.-]*`)

// cmdMainPathMap carries the call-chain notation map from the t432
// normalization rules (spec §1.1 P8): `cmd/moai/main` denotes the function
// chain rooted at `cmd/moai/main.go`, not a missing path.
var cmdMainPathMap = map[string]string{
	"cmd/moai/main": "cmd/moai/main.go",
}

// citedPathTrailingPunct is the punctuation a markdown sentence can glue to
// the end of a path token (`.`, `,`, `;`, `:`, `)` and quote forms). The
// extraction charset stops at other characters on its own.
const citedPathTrailingPunct = ".,;:)]}\"'"

// checkCitations measures the cited-path-existence axis over the codemaps
// docs under projectRoot: extract cited source-path tokens from every
// non-blockquote line of `.moai/project/codemaps/*.md`, normalize each with
// the t432 rules, and count the unique positively-cited paths absent from
// the working tree. Threshold is 0 by construction — one positive phantom is
// red (REQ-CMA-002). The row rides the same LayerReport consumption path
// (Failed()/OffendingLayers()) as the three freshness layers, so the CLI's
// exit-code contract (0/1/2) is unchanged.
//
// A missing codemaps directory or an unreadable doc leaves the layer
// unjudgeable — absent with a reason, never silently fresh (the existing
// layer contract). Read errors on individual docs likewise degrade the whole
// layer to absent: a partially-measured green would assert accuracy over
// files the walk never saw.
//
// @MX:NOTE: [AUTO] citations layer — accuracy axis distinct from the 3 freshness layers; blockquote exemption is deliberate (D2), documented as a false-negative risk in spec.md D2
func checkCitations(projectRoot string) LayerReport {
	rep := LayerReport{Layer: LayerCitations, Metric: MetricPositiveCitedPathAbsence, Threshold: 0}

	dir := filepath.Join(projectRoot, ".moai", "project", "codemaps")
	entries, err := os.ReadDir(dir)
	if err != nil {
		rep.Verdict = VerdictAbsent
		rep.Reason = "codemaps directory missing"
		return rep
	}

	absent := map[string]string{} // normalized absent path -> first citing doc
	var docs int
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		docs++
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			rep.Verdict = VerdictAbsent
			rep.Reason = fmt.Sprintf("codemaps doc %s unreadable: %v", e.Name(), err)
			return rep
		}
		for _, cited := range positiveCitedPaths(string(data)) {
			normalized := normalizeCitedPath(projectRoot, cited)
			if _, statErr := os.Lstat(filepath.Join(projectRoot, filepath.FromSlash(normalized))); statErr != nil {
				if _, seen := absent[normalized]; !seen {
					absent[normalized] = e.Name()
				}
			}
		}
	}
	if docs == 0 {
		rep.Verdict = VerdictAbsent
		rep.Reason = "no codemaps documents to check"
		return rep
	}

	paths := make([]string, 0, len(absent))
	for p := range absent {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	rep.Value = len(paths)
	if len(paths) > 0 {
		rep.Verdict = VerdictStale
		rep.Reason = fmt.Sprintf("%d positively-cited path(s) absent from the tree (first cited in %s)", len(paths), absent[paths[0]])
		// The driving-path bound contract (REQ-GFC-008) is reused verbatim:
		// a red naming 80 paths stays readable, overflow declared.
		attachDrivingPaths(&rep, paths)
	} else {
		rep.Verdict = VerdictFresh
	}
	return rep
}

// positiveCitedPaths extracts cited source-path tokens from the doc, skipping
// blockquote (`>`-prefixed) lines entirely: paths on those lines are
// negative-context citations — warning notes recording removals and renames —
// and are exempt by design (D2). The exemption is a line-level, syntactic
// discriminator only; no intent inference.
func positiveCitedPaths(doc string) []string {
	var out []string
	seen := map[string]bool{}
	for _, line := range strings.Split(doc, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), ">") {
			continue
		}
		for _, m := range citedPathPattern.FindAllString(line, -1) {
			if !seen[m] {
				seen[m] = true
				out = append(out, m)
			}
		}
	}
	return out
}

// normalizeCitedPath applies the t432 normalization rules (REQ-CMA-002):
// trailing punctuation trim, trailing-slash strip, the `cmd/moai/main`
// call-chain map, and the `.go`-suffix restore for tokens whose separating
// period was stripped (a `xxxgo` artifact). No rule invents existence — an
// unknown path passes through unchanged and fails the existence test.
func normalizeCitedPath(projectRoot, raw string) string {
	p := strings.TrimRight(raw, citedPathTrailingPunct)
	p = strings.TrimRight(p, "/")
	if mapped, ok := cmdMainPathMap[p]; ok {
		return mapped
	}
	if _, err := os.Lstat(filepath.Join(projectRoot, filepath.FromSlash(p))); err == nil {
		return p
	}
	// `.go`-suffix restore: `internal/graph/checkgo`-shaped artifacts resolve
	// to the real `internal/graph/check.go`.
	if strings.HasSuffix(p, "go") && len(p) > 2 {
		restored := p[:len(p)-2] + ".go"
		if _, err := os.Lstat(filepath.Join(projectRoot, filepath.FromSlash(restored))); err == nil {
			return restored
		}
	}
	return p
}
