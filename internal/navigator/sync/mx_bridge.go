package sync

import (
	"path/filepath"
	"sort"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// MxBridgeSpec is one observed `@MX:SPEC` association bridged into the
// Navigator graph (REQ-NS-005). It is the consumer-side reshape of the
// mx-scanner's per-tag SpecAssociator output: the code-side location
// (tag.File + tag.Line) joined to the associated SPEC ID.
type MxBridgeSpec struct {
	SpecID     string // the associated SPEC ID (e.g. SPEC-AUTH-001)
	SourcePath string // the code-side path where the association was observed
	LineNumber int    // 1-indexed line in SourcePath
}

// BridgeMxAssociations consumes the existing internal/mx SpecAssociator
// output (REQ-NS-005): it walks the project source tree with mx.Scanner,
// builds the SPEC ID → module-paths map via mx.LoadSpecModules, runs the
// associator per tag, and returns the inverse `tag location → SPEC ID`
// reshape for graph consumption.
//
// This function does NOT modify internal/mx/ (REQ-NS-005); it imports and
// consumes its public API. Fail-open: any mx-scanner error returns the
// records collected so far + the error.
func BridgeMxAssociations(projectRoot string) ([]MxBridgeSpec, error) {
	specModules, err := mx.LoadSpecModules(projectRoot)
	if err != nil {
		return nil, err
	}
	scanner := mx.NewScanner()
	// Seed the scanner's ignore list with the shared default so the walk skips
	// `.git/`, `.claude/`, `.moai/`, `vendor/`, `node_modules/`, and the rest
	// of DefaultScanIgnore. Without this call, `NewScanner()` leaves
	// ignorePatterns nil and the walk descends into every directory — the same
	// ignore set the `moai mx scan` CLI and the SessionStart-hook background
	// scan use (internal/cli/mx_scan.go, internal/hook/session_start.go).
	scanner.SetIgnorePatterns(mx.DefaultScanIgnore)

	// Scan the project source tree for @MX tags. The mx-scanner walks
	// .moai/specs/ for body-based association (already aggregated by the
	// associator) plus the source tree. We scope the walk to the same roots
	// the mx command uses, bounded to projectRoot to avoid escaping.
	tags, err := scanner.ScanDir(projectRoot)
	if err != nil {
		return nil, err
	}

	// @MX:DEBT: path-based association (source (a) of AssociateWithDiagnostics)
	// fans out far wider than it is useful, because SPEC `module:` declarations
	// are coarse. Measured on this repo: 863 tags produce ~28.8k bridge records
	// across 259 SPECs, and each record materializes both a symbol node and a
	// spec-edge — a ~7 MB nav-graph.json. The driver is declaration breadth, not
	// the matching rule: 34 SPECs declare `module: internal` (the whole tree) and
	// 65 declare `internal/cli`, so one tag under internal/cli/ edges to all of
	// them. The blast radius is LATENT today — join.go's capability gate
	// suppresses the artifact while capability-map.md is absent.
	// @MX:CEILING: no algorithmic rule fixes this. Measured alternatives:
	// separator-boundary matching removes 77 records (false sibling matches
	// only), and most-specific-module-wins removes ~43% but leaves ~16.4k,
	// because the 65 SPECs sharing `internal/cli` tie at the same specificity.
	// @MX:UPGRADE: narrow the `module:` declarations themselves — a SPEC whose
	// module is the entire internal tree carries no navigational signal. Until
	// then, treat spec-edge density as an artifact of declaration breadth.
	assoc := mx.NewSpecAssociator(specModules)
	seen := map[string]bool{}
	var out []MxBridgeSpec
	for _, tag := range tags {
		// Normalize the tag's file path to project-relative BEFORE association.
		// `scanner.ScanDir` was called with an ABSOLUTE projectRoot, so tag.File
		// is absolute; but `mx.LoadSpecModules` returns module paths verbatim
		// from spec.md frontmatter (project-relative). The associator's
		// path-based source (a) does a plain `strings.HasPrefix(filePath,
		// modulePath)` (internal/mx/spec_association.go), so an absolute tag.File
		// never matches a relative module path and path-based association
		// silently fails. Relativizing here — on a per-iteration copy, NOT
		// mutating the shared tag — makes path-based association work and keeps
		// SourcePath consistent with the rest of the graph (project-relative).
		// Consumer-only: this bridge does NOT modify internal/mx (REQ-NS-005).
		// `filepath.Rel` returns OS-native separators, so on Windows the result
		// is `internal\mx\tagged.go` while `mx.LoadSpecModules` returns module
		// paths verbatim from frontmatter YAML (always forward-slash). Without
		// ToSlash the HasPrefix comparison below fails on Windows exactly as it
		// did before relativization, and `mx:` node identifiers would differ
		// between a graph built on Windows and one built on macOS/Linux.
		assocFile := filepath.ToSlash(tag.File)
		if rel, err := filepath.Rel(projectRoot, tag.File); err == nil {
			assocFile = filepath.ToSlash(rel)
		}
		assocTag := tag
		assocTag.File = assocFile
		specIDs, _ := assoc.AssociateWithDiagnostics(assocTag)
		for _, specID := range specIDs {
			key := specID + "|" + assocFile + "|" + itoa(tag.Line)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, MxBridgeSpec{
				SpecID:     specID,
				SourcePath: assocFile,
				LineNumber: tag.Line,
			})
		}
	}
	sortMxBridge(out)
	return out, nil
}

// sortMxBridge orders the bridge slice deterministically.
func sortMxBridge(b []MxBridgeSpec) {
	sort.SliceStable(b, func(i, j int) bool {
		if b[i].SpecID != b[j].SpecID {
			return b[i].SpecID < b[j].SpecID
		}
		if b[i].SourcePath != b[j].SourcePath {
			return b[i].SourcePath < b[j].SourcePath
		}
		return b[i].LineNumber < b[j].LineNumber
	})
}

// itoa is a small strconv-free int→string helper to keep imports lean.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
