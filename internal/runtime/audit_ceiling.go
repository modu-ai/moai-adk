// audit_ceiling.go — SPEC-AUDIT-CEILING-002: the CLI-counted plan-audit
// iteration ceiling. The counter derives a SPEC's plan-audit round count from
// durable iteration evidence on disk (one recorded file one round), the
// ceiling resolves from the typed harness configuration, and the latest
// verdict is selected by the recorded iteration number — never from
// in-process memory or session state.
//
// The C5 exemption holds: no syscall and no build-tag literal exist in this
// file's scope.
package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

// planAuditRoundFile is one recorded plan-audit file: plan-audit.md ranks as
// iteration 0 and each plan-audit-iter<N>.md carries its parsed N.
type planAuditRoundFile struct {
	path string
	n    int
}

// iterFileName matches the plan-audit-iter<N>.md family; the captured suffix
// must parse as a positive integer. The suffix may be empty (an
// iter-with-no-number file still carries the family shape, so it fails the
// count closed rather than skipping silently).
var iterFileName = regexp.MustCompile(`^plan-audit-iter(.*)\.md$`)

// countPlanAuditRounds is the shared scanner behind CountPlanAuditRounds and
// SelectLatestVerdict: it walks the SPEC-scoped directory and the explicitly
// listed directories and returns the plan-audit round files each holds.
//
// Directory-class rules (REQ-ACR-001/013): an absent SPEC-scoped directory
// contributes 0 (an unaudited SPEC is not an error); an explicitly LISTED
// directory that does not exist is an error naming the missing path — a typo
// must never silently count 0. Identical listed paths are deduplicated before
// counting (R6); a path listed explicitly keeps its error-on-absent class even
// when it equals the SPEC-scoped one (fail-closed direction).
func scanEvidenceDirs(specDir string, listedDirs []string) ([]planAuditRoundFile, error) {
	type dirInput struct {
		path   string
		listed bool
	}
	var inputs []dirInput
	seen := map[string]bool{}
	// Listed directories register first so their error-on-absent class wins a
	// dedupe collision against the SPEC-scoped class.
	for _, d := range listedDirs {
		if d == "" {
			continue
		}
		clean := filepath.Clean(d)
		if seen[clean] {
			continue
		}
		seen[clean] = true
		inputs = append(inputs, dirInput{path: clean, listed: true})
	}
	if specDir != "" {
		clean := filepath.Clean(specDir)
		if !seen[clean] {
			seen[clean] = true
			inputs = append(inputs, dirInput{path: clean, listed: false})
		}
	}

	var out []planAuditRoundFile
	for _, in := range inputs {
		entries, err := os.ReadDir(in.path)
		if err != nil {
			if in.listed || !os.IsNotExist(err) {
				return nil, fmt.Errorf("read evidence directory %q: %w", in.path, err)
			}
			continue // SPEC-scoped and absent → contributes 0
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if name == "plan-audit.md" {
				out = append(out, planAuditRoundFile{path: filepath.Join(in.path, name), n: 0})
				continue
			}
			m := iterFileName.FindStringSubmatch(name)
			if m == nil {
				continue // not in the plan-audit round family
			}
			n, perr := strconv.Atoi(m[1])
			if perr != nil || n <= 0 {
				return nil, fmt.Errorf("round file %q has an iteration suffix that does not parse as a positive integer", filepath.Join(in.path, name))
			}
			out = append(out, planAuditRoundFile{path: filepath.Join(in.path, name), n: n})
		}
	}
	return out, nil
}

// CountPlanAuditRounds computes a SPEC's plan-audit round count from durable
// iteration evidence on disk, one recorded file one round: plan-audit.md
// contributes one round and each plan-audit-iter<N>.md one round. specDir is
// the SPEC-scoped evidence directory (.moai/reports/<SPEC-ID>); listedDirs are
// the evidence directories explicitly listed in the invocation. The count is
// never derived from in-process memory or session state.
func CountPlanAuditRounds(specDir string, listedDirs []string) (int, error) {
	files, err := scanEvidenceDirs(specDir, listedDirs)
	if err != nil {
		return 0, err
	}
	return len(files), nil
}

// SelectLatestVerdict returns the recorded file with the highest parsed
// iteration number — plan-audit.md ranking as iteration 0 — the verdict the
// disposition branches read. The same highest number appearing in two evidence
// directories is a selection error rather than a silent pick; no evidence at
// all selects nothing (empty path, no error).
func SelectLatestVerdict(specDir string, listedDirs []string) (string, error) {
	files, err := scanEvidenceDirs(specDir, listedDirs)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", nil
	}
	sort.Slice(files, func(i, j int) bool { return files[i].n < files[j].n })
	best := files[len(files)-1]
	var tied []string
	for _, f := range files {
		if f.n == best.n {
			tied = append(tied, f.path)
		}
	}
	if len(tied) > 1 {
		return "", fmt.Errorf("latest plan-audit verdict is ambiguous: iteration %d recorded in %d evidence files (%q and %q); list one evidence directory to disambiguate",
			best.n, len(tied), tied[0], tied[1])
	}
	return best.path, nil
}

// ResolvePlanAuditCeiling resolves the SPEC's ceiling from the configured
// plan_audit_tier_ceilings map. The tier resolves under the shared predicate's
// rule (auditverdict.SpecTier): an absent or unknown tier resolves to L. A
// resolved ceiling that is missing from the map or non-positive is a
// configuration error rather than a ceiling of zero (REQ-ACR-002).
func ResolvePlanAuditCeiling(tier string, ceilings map[string]int) (int, error) {
	resolved := tier
	if resolved != "S" && resolved != "M" && resolved != "L" {
		resolved = "L" // auditverdict.SpecTier's rule: absent or unknown → L
	}
	ceiling, ok := ceilings[resolved]
	if !ok || ceiling <= 0 {
		return 0, fmt.Errorf("plan_audit_tier_ceilings carries no positive ceiling for tier %q (resolved to %q); a mis-configured ceilings map must never resolve to a ceiling of zero", tier, resolved)
	}
	return ceiling, nil
}
