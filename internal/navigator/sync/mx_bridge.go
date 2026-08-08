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

	// Scan the project source tree for @MX tags. The mx-scanner walks
	// .moai/specs/ for body-based association (already aggregated by the
	// associator) plus the source tree. We scope the walk to the same roots
	// the mx command uses, bounded to projectRoot to avoid escaping.
	tags, err := scanner.ScanDir(projectRoot)
	if err != nil {
		return nil, err
	}

	assoc := mx.NewSpecAssociator(specModules)
	seen := map[string]bool{}
	var out []MxBridgeSpec
	for _, tag := range tags {
		// Bug-3 fix: the scanner stores an ABSOLUTE tag.File (the walk path),
		// while mx.LoadSpecModules returns RELATIVE module paths verbatim from
		// SPEC frontmatter. isFileUnderModules does a plain strings.HasPrefix,
		// so HasPrefix(absPath, relModule) was always false and path-based
		// association never fired. Normalize to project-relative here so the
		// associator sees the same shape the frontmatter carries.
		relFile, err := filepath.Rel(projectRoot, tag.File)
		if err != nil {
			relFile = tag.File // fall back to the raw path on error (fail-open)
		}
		tag.File = filepath.ToSlash(relFile)
		specIDs, _ := assoc.AssociateWithDiagnostics(tag)
		for _, specID := range specIDs {
			key := specID + "|" + tag.File + "|" + itoa(tag.Line)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, MxBridgeSpec{
				SpecID:     specID,
				SourcePath: tag.File,
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
