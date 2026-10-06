package runtime

// audit_counter.go — the per-SPEC plan-audit round counter
// (SPEC-AUDIT-CEILING-001 REQ-ACE-001). The count is derived from durable
// iteration evidence on disk, never from in-process memory: the convention
// family (plan-audit[-iter<N>].md) and the legacy stream
// (<SPEC-ID>-review-<N>.md), with the dedupe identity being the bare
// (SPEC id, iteration number) pair — the same iteration recorded in both
// families and across card directories counts once, and different iteration
// numbers count even when the verdict label, score, and audited hash are
// identical. A file whose iteration number is unreadable is never collapsed
// into another file (fail-counted). The daily run-history file
// (.moai/reports/plan-audit/…) is never counted (plan.md §G).

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/auditverdict"
)

// RoundEvidence is the per-SPEC audit-round evidence the ceiling engine and
// the admission seams read.
type RoundEvidence struct {
	// Count is the number of distinct (SPEC id, iteration number) pairs.
	Count int
	// Sources lists every evidence file consulted.
	Sources []string
	// Latest is the parsed highest-iteration verdict, nil when none.
	Latest *auditverdict.Fields
	// LatestPath is the file Latest was parsed from.
	LatestPath string
}

var (
	// conventionFile is the plan-audit family: a bare single-iteration file
	// (iteration 1), one file per iteration (plan-audit-iter<N>.md), or the
	// numbered shape plan-audit-<N>.md the kickoff reader and the factory
	// card record also accept — the counter and the reader must agree on
	// what an iteration file is, or the ceiling never sees the verdict the
	// seam judges (card-review F2).
	conventionFile = regexp.MustCompile(`^plan-audit(?:-iter)?-?([0-9]+)?\.md$`)
	// conventionUnnumbered is the family shape with an iteration number the
	// counter cannot read — fail-counted, never collapsed (REQ-ACE-001).
	conventionUnnumbered = regexp.MustCompile(`^plan-audit-iter[^/]*\.md$`)
	// specIDToken matches a SPEC identifier anywhere; the FIRST token in a
	// report body is its header's SPEC attribution (card-review F5).
	specIDToken = regexp.MustCompile(`SPEC-[A-Z][A-Z0-9]*(?:-[A-Z0-9]+)*`)
)

// maxReportFileSize bounds one evidence-file read, mirroring the homestate
// readBoundedFile discipline (card_evidence_readers.go, same 4 MiB value).
// The constant is redeclared rather than imported: homestate already imports
// this package, so the dependency runs one way only.
const maxReportFileSize = int64(4 << 20)

// readReportFile reads one evidence file only when it is a regular file of
// at most maxReportFileSize bytes. Card-review round-2 P1: os.ReadFile on a
// FIFO blocks until a writer appears, which hung kickoff and card
// transitions; the Stat check refuses non-regular files without opening
// them, so no evidence path can block the caller.
func readReportFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxReportFileSize {
		return nil, fmt.Errorf("not a regular file of at most %d bytes", maxReportFileSize)
	}
	return os.ReadFile(path)
}

// CountAuditRounds computes the round count of specID from the given report
// directories (see RoundReportDirs). A directory that does not exist is
// skipped; an unreadable directory is an error — a count the caller cannot
// verify must not silently read as zero.
//
// @MX:NOTE: [AUTO] the round count derives from durable disk evidence only — dedupe identity is the (SPEC id, iteration number) pair; in-process memory is never consulted
// @MX:SPEC: SPEC-AUDIT-CEILING-001
func CountAuditRounds(specID string, reportDirs []string) (RoundEvidence, error) {
	var ev RoundEvidence
	seen := map[int]bool{}
	failCounted := 0
	bestN := -1
	legacyFile := regexp.MustCompile(`^` + regexp.QuoteMeta(specID) + `-review-([0-9]+)\.md$`)

	for _, dir := range reportDirs {
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return ev, fmt.Errorf("read report directory %s: %w", dir, err)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			n := -1
			fail := false
			convention := false
			if m := conventionFile.FindStringSubmatch(name); m != nil {
				convention = true
				n = 1
				if m[1] != "" {
					if v, perr := strconv.Atoi(m[1]); perr == nil {
						n = v
					}
				}
			} else if m := legacyFile.FindStringSubmatch(name); m != nil {
				n, _ = strconv.Atoi(m[1])
			} else if conventionUnnumbered.MatchString(name) {
				convention = true
				fail = true
			} else {
				continue
			}
			path := filepath.Join(dir, name)
			// A convention-family file belongs to the SPEC named in its
			// report header (REQ-ACE-001), compared exactly on the FIRST
			// SPEC identifier in the body — the header's attribution. A
			// foreign SPEC's report that merely mentions the target in prose
			// is not its round (card-review F5); the legacy family names the
			// SPEC in its file name.
			if convention {
				data, rerr := readReportFile(path)
				if rerr != nil {
					return ev, fmt.Errorf("read plan-audit evidence %s: %w", path, rerr)
				}
				if head := specIDToken.Find(data); head == nil || string(head) != specID {
					continue
				}
			}
			ev.Sources = append(ev.Sources, path)
			if fail {
				failCounted++
				continue
			}
			seen[n] = true
			if n > bestN {
				bestN = n
				ev.LatestPath = path
			}
		}
	}
	ev.Count = len(seen) + failCounted
	if ev.LatestPath != "" {
		data, err := readReportFile(ev.LatestPath)
		if err != nil {
			return ev, fmt.Errorf("read plan-audit evidence %s: %w", ev.LatestPath, err)
		}
		f := auditverdict.Parse(data)
		ev.Latest = &f
	}
	return ev, nil
}

// RoundReportDirs enumerates the report directories that may carry specID's
// round evidence: the SPEC-scoped directory, the card's directory, and every
// other report directory whose plan-audit iteration files name the SPEC
// (REQ-ACE-001's SPEC attribution rule).
func RoundReportDirs(projectRoot, specID, cardID string) []string {
	reportsRoot := filepath.Join(projectRoot, ".moai", "reports")
	specDir := filepath.Join(reportsRoot, specID)
	dirs := []string{specDir}
	seen := map[string]bool{specDir: true}
	if cardID != "" {
		d := filepath.Join(reportsRoot, cardID)
		if !seen[d] {
			seen[d] = true
			dirs = append(dirs, d)
		}
	}
	entries, err := os.ReadDir(reportsRoot)
	if err != nil {
		return dirs
	}
	for _, e := range entries {
		if !e.IsDir() || e.Name() == "plan-audit" {
			continue // the daily run-history directory is never evidence
		}
		d := filepath.Join(reportsRoot, e.Name())
		if seen[d] {
			continue
		}
		if cardDirNamesSpec(d, specID) {
			seen[d] = true
			dirs = append(dirs, d)
		}
	}
	return dirs
}

// cardDirNamesSpec reports whether dir carries round evidence for specID: a
// plan-audit iteration file whose report header names the SPEC, or a legacy
// stream file whose FILE NAME names it (card-review F6 — a legacy-only
// directory is evidence too).
func cardDirNamesSpec(dir, specID string) bool {
	legacyFile := regexp.MustCompile(`^` + regexp.QuoteMeta(specID) + `-review-[0-9]+\.md$`)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if legacyFile.MatchString(name) {
			return true
		}
		if !conventionFile.MatchString(name) && !conventionUnnumbered.MatchString(name) {
			continue
		}
		data, err := readReportFile(filepath.Join(dir, name))
		if err == nil {
			if head := specIDToken.Find(data); head != nil && string(head) == specID {
				return true
			}
		}
	}
	return false
}

// auditedSHAOf reads the audited_sha machine line of one verdict file.
func auditedSHAOf(path string) string {
	data, err := readReportFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "audited_sha:"); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
