package mx

import (
	"strings"
	"testing"
)

// TestSpecRefDiscard_UnpairedWarnBlock locks the scanner's handling of an
// @MX:SPEC sub-line whose owning tag is an unpaired WARN (t620).
//
// The judgment this test records: discarding the SpecRef along with its carrier
// is CORRECT, not a silent loss.
//
//  1. Every path that drops an unpaired WARN also emits MissingReasonForWarn at
//     the WARN's own line, and the @MX:SPEC sub-line sits within the WARN's
//     sub-line window — so the block IS reported, at the coordinate whose single
//     repair (adding @MX:REASON) restores the tag, its SpecRef, and the
//     downstream spec-edge built in internal/graph.
//  2. The alternative — falling back to an earlier tag — is the defect t612
//     removed: the SPEC would attach to a tag the author never wrote it under.
//     The "no fallback" rows below are what keeps that fix in place.
//
// A second diagnostic naming the discarded SpecRef is therefore NOT added: it
// would report one repair twice, 1-3 lines apart, inside a warning list already
// truncated at internal/cli.maxScannerWarningsShown.
func TestSpecRefDiscard_UnpairedWarnBlock(t *testing.T) {
	const specID = "SPEC-FIXTURE-620"

	tests := []struct {
		name string
		body string
		// wantSpecRefOn is the Kind of the single tag expected to carry specID,
		// or "" when no surviving tag may carry it.
		wantSpecRefOn TagKind
		wantTagKinds  []TagKind
		wantWarnings  []string
		denyWarnings  []string
	}{
		{
			// Control 1: the discard case. The WARN never pairs, so it and its
			// SpecRef both leave the output; MissingReasonForWarn is the signal.
			name: "unpaired WARN drops its SpecRef with the tag",
			body: `// @MX:WARN: unguarded goroutine
// @MX:SPEC: ` + specID + `
func f() {}
`,
			wantSpecRefOn: "",
			wantTagKinds:  nil,
			wantWarnings:  []string{"MissingReasonForWarn"},
			denyWarnings:  []string{"DanglingSpecRef", "UnresolvedSpecRef"},
		},
		{
			// Control 2: the healthy block. REASON pairs the WARN, so the tag is
			// emitted and carries the SpecRef.
			name: "paired WARN keeps its SpecRef",
			body: `// @MX:WARN: unguarded goroutine
// @MX:SPEC: ` + specID + `
// @MX:REASON: lifetime is bounded by the caller
func f() {}
`,
			wantSpecRefOn: MXWarn,
			wantTagKinds:  []TagKind{MXWarn},
			wantWarnings:  nil,
			denyWarnings:  []string{"MissingReasonForWarn", "DanglingSpecRef"},
		},
		{
			// Control 3: no preceding tag at all — the sub-line is dangling at
			// capture time, which is a distinct diagnostic from control 1.
			name: "SPEC sub-line with no preceding tag is dangling",
			body: `// @MX:SPEC: ` + specID + `
func f() {}
`,
			wantSpecRefOn: "",
			wantTagKinds:  nil,
			wantWarnings:  []string{"DanglingSpecRef"},
			denyWarnings:  []string{"MissingReasonForWarn"},
		},
		{
			// The asymmetry that makes control 1 correct rather than lossy: an
			// unpaired WARN owns the sub-line and takes it down with it. It must
			// NOT fall back to the NOTE above — that is the t612 defect.
			name: "unpaired WARN does not hand its SpecRef back to an earlier tag",
			body: `// @MX:NOTE: a context note
// @MX:WARN: unguarded goroutine
// @MX:SPEC: ` + specID + `
func f() {}
`,
			wantSpecRefOn: "",
			wantTagKinds:  []TagKind{MXNote},
			wantWarnings:  []string{"MissingReasonForWarn"},
			denyWarnings:  []string{"DanglingSpecRef"},
		},
		{
			// Baseline for the row above: with no WARN in between, the same
			// sub-line does attach to the NOTE. The delta between these two rows
			// is the whole behavior under test.
			name: "without the WARN the same sub-line attaches to the NOTE",
			body: `// @MX:NOTE: a context note
// @MX:SPEC: ` + specID + `
func f() {}
`,
			wantSpecRefOn: MXNote,
			wantTagKinds:  []TagKind{MXNote},
			wantWarnings:  nil,
			denyWarnings:  []string{"MissingReasonForWarn", "DanglingSpecRef"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := writeSubLineFixture(t, dir, "discard.go", tt.body)

			scanner := NewScanner()
			tags, err := scanner.ScanFile(path)
			if err != nil {
				t.Fatalf("ScanFile: %v", err)
			}

			gotKinds := make([]TagKind, 0, len(tags))
			for _, tag := range tags {
				gotKinds = append(gotKinds, tag.Kind)
			}
			if len(gotKinds) != len(tt.wantTagKinds) {
				t.Fatalf("tag kinds: want %v, got %v (%+v)", tt.wantTagKinds, gotKinds, tags)
			}
			for i, want := range tt.wantTagKinds {
				if gotKinds[i] != want {
					t.Fatalf("tag kinds: want %v, got %v", tt.wantTagKinds, gotKinds)
				}
			}

			// Reachability: the fixture's SPEC ID must be well-formed enough for
			// extractSpecRef to see it at all — otherwise every "no SpecRef"
			// assertion below would pass for the wrong reason.
			if tt.wantSpecRefOn != "" {
				found := false
				for _, tag := range tags {
					if tag.SpecRef == specID {
						if tag.Kind != tt.wantSpecRefOn {
							t.Errorf("SpecRef carrier: want %s, got %s", tt.wantSpecRefOn, tag.Kind)
						}
						found = true
					}
				}
				if !found {
					t.Errorf("no emitted tag carries %q (tags=%+v)", specID, tags)
				}
			} else {
				for _, tag := range tags {
					if tag.SpecRef == specID {
						t.Errorf("no emitted tag may carry %q, but %s at line %d does",
							specID, tag.Kind, tag.Line)
					}
				}
			}

			warnings := scanner.GetWarnings()
			for _, want := range tt.wantWarnings {
				if !containsWarning(warnings, want) {
					t.Errorf("missing %s warning; got %v", want, warnings)
				}
			}
			for _, deny := range tt.denyWarnings {
				if containsWarning(warnings, deny) {
					t.Errorf("unexpected %s warning; got %v", deny, warnings)
				}
			}
		})
	}
}

func containsWarning(warnings []string, category string) bool {
	for _, w := range warnings {
		if strings.Contains(w, category) {
			return true
		}
	}
	return false
}
