package template

// The classification engine of the github-flow sweep guard (SPEC-GITHUB-FLOW-DEFAULT-001
// design D-9, as amended by D-27; card t1453 M4 step 1).
//
// STUB — every entry point returns "clean". This is the observed-RED stand-in:
// the fixture tests run against it first so the missing engine is a failing
// assertion, not a build error (verification-completeness §1.1).

// sweepClass is the verdict class of one finding.
type sweepClass string

const (
	sweepStrong  sweepClass = "strong"
	sweepWeak    sweepClass = "weak"
	sweepExempt  sweepClass = "exempt"
	sweepRatchet sweepClass = "ratchet"
)

// sweepFinding is one classified line.
type sweepFinding struct {
	File     string
	Line     int
	Text     string
	Class    sweepClass
	Rule     string
	Why      string
	AllowIdx int
}

func (f sweepFinding) violation() bool { return f.Class == sweepStrong || f.Class == sweepWeak }

// sweepAllow is one whole-line allow-list entry.
type sweepAllow struct {
	File    string
	Literal string
	Why     string
}

// sweepSubtree is one scoped surface subtree and its visited-count floor.
type sweepSubtree struct {
	Name         string
	Files        []string
	Dir          string
	Exts         []string
	NonRecursive bool
	Floor        int
}

type sweepCounts struct {
	MarkerLines   int
	MarkerH1Lines int
	PCLines       int
}

type sweepCeilings struct {
	MarkerLines   int
	MarkerH1Lines int
	PCLines       int
}

func sweepScan(file, content string, allow []sweepAllow) []sweepFinding { return nil }

func sweepValidateAllow(entries []sweepAllow, cap int) error { return nil }

func sweepCountsOf(findings []sweepFinding) sweepCounts { return sweepCounts{} }

func sweepRatchetProblems(c sweepCounts, ceil sweepCeilings) []string { return nil }

func sweepVisitProblems(visited map[string]int, subtrees []sweepSubtree, totalFloor int) []string {
	return nil
}
