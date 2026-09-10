package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

type changedLineRange struct{ Start, End int }
type changedCoverageResult struct {
	Covered, Total int
	Percent        float64
}

const (
	homeStateCoverageCommitSubject            = "feat(state): add guarded home-state rollout (t592)"
	homeStateCoverageRemediationCommitSubject = "fix(state): stabilize committed coverage evidence (t592)"
	homeStateCoverageDeltaCommitSubject       = "fix(state): isolate committed coverage deltas (t592)"
)

type homeStateCoverageChangeSet struct {
	Base, Tip           string
	Native, Disposition []string
	Ranges              map[string][]changedLineRange
}

func measureChangedSurfaceCoverage(ctx context.Context, root string) (float64, error) {
	result, err := measureChangedSurfaceCoverageResultWith(ctx, root, runChangedSurfaceCoverageSuite)
	return result.Percent, err
}

func runChangedSurfaceCoverageSuite(ctx context.Context, root, path string) error {
	pattern := "^(TestHomeState.*|TestHomeLayout.*|TestProject(Layout|Dir).*|TestRuntimeCensus.*|TestAdmission.*|TestFactory.*|TestResolveFactory.*|TestResume.*|TestExpireResume.*|TestImportLegacy.*|TestHandoff.*|TestClaim.*|TestClaimThenInject.*|TestConcurrentConsume.*|TestSQLiteClaim.*|TestNonceFallback.*|TestManualMode.*|TestNonClearSource.*|TestDegradeToGuidance.*|TestFailOpen_CorruptPending.*|TestStaleTTL.*|TestBranchTable.*|TestRenderHandoff.*|TestHandle_.*|TestIsHex.*|TestInjectionHeader.*|TestNoUserInteraction.*|TestThreeHandler.*|TestProfile.*|TestCleanHome.*|TestScanHomeCleanable.*|TestSecureHomeDirectories.*|TestContinue.*|TestCC.*|TestCharacterize_CC.*|TestRunCC.*|TestRunGLM.*|TestSession(Start|End).*|TestPersistedHomeStateEvidence.*|TestPlatformProcessIdentity.*|TestParseChanged.*|TestChangedProduction.*|TestCommittedCoverage.*)$"
	coverpkg := "./internal/cli,./internal/homestate,./internal/hook/handoff,./internal/hook,./internal/kanban"
	parts := []struct {
		name string
		args []string
	}{
		{"cli", []string{"./internal/cli", "-run", pattern}},
		{"homestate", []string{"./internal/homestate"}},
		{"handoff", []string{"./internal/hook/handoff"}},
		{"hook", []string{"./internal/hook", "-run", pattern}},
		{"kanban", []string{"./internal/kanban", "-run", pattern}},
	}
	var merged strings.Builder
	merged.WriteString("mode: set\n")
	for _, part := range parts {
		partPath := path + "." + part.name
		args := append([]string{"test"}, part.args...)
		args = append(args, "-count=1", "-coverpkg="+coverpkg, "-coverprofile="+partPath)
		cmd := exec.CommandContext(ctx, "go", args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "MOAI_HOME_STATE_COVERAGE_CHILD=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("coverage %s tests: %w\n%s", part.name, err, out)
		}
		raw, err := os.ReadFile(partPath)
		if err != nil {
			return err
		}
		lines := strings.SplitN(string(raw), "\n", 2)
		if len(lines) != 2 || !strings.HasPrefix(lines[0], "mode:") {
			return fmt.Errorf("coverage %s profile malformed", part.name)
		}
		merged.WriteString(lines[1])
		_ = os.Remove(partPath)
	}
	return os.WriteFile(path, []byte(merged.String()), 0o600)
}

func measureChangedSurfaceCoverageWith(ctx context.Context, root string, run func(context.Context, string, string) error) (float64, error) {
	result, err := measureChangedSurfaceCoverageResultWith(ctx, root, run)
	return result.Percent, err
}

func measureChangedSurfaceCoverageResultWith(ctx context.Context, root string, run func(context.Context, string, string) error) (changedCoverageResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Minute)
	defer cancel()
	profile, err := os.CreateTemp("", "moai-home-state-coverage-*.out")
	if err != nil {
		return changedCoverageResult{}, err
	}
	path := profile.Name()
	if err := profile.Close(); err != nil {
		return changedCoverageResult{}, err
	}
	defer func() { _ = os.Remove(path) }()
	if err := run(ctx, root, path); err != nil {
		return changedCoverageResult{}, err
	}
	changeSet, err := resolveHomeStateCoverageChangeSet(root)
	if err != nil {
		return changedCoverageResult{}, err
	}
	result, err := parseChangedLineCoverage(path, changeSet.Ranges)
	if err != nil {
		return changedCoverageResult{}, err
	}
	return result, nil
}

func changedProductionLineRanges(root string, files []string) (map[string][]changedLineRange, error) {
	changeSet, err := resolveHomeStateCoverageChangeSet(root)
	if err != nil {
		return nil, err
	}
	ranges := changeSet.Ranges
	wanted := map[string]bool{}
	for _, file := range files {
		wanted[file] = true
	}
	for file := range ranges {
		if !wanted[file] {
			delete(ranges, file)
		}
	}
	for _, file := range files {
		if _, ok := ranges[file]; !ok {
			return nil, fmt.Errorf("changed production file missing from diff: %s", file)
		}
	}
	return ranges, nil
}

func parseUnifiedZeroDiff(diff string) (map[string][]changedLineRange, error) {
	result := map[string][]changedLineRange{}
	current := ""
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+++ ") {
			path := strings.TrimPrefix(line, "+++ ")
			if path == "/dev/null" {
				return nil, fmt.Errorf("deleted production file")
			}
			if !strings.HasPrefix(path, "b/") {
				return nil, fmt.Errorf("malformed diff target %q", path)
			}
			current = filepath.ToSlash(strings.TrimPrefix(path, "b/"))
			if strings.HasSuffix(current, ".go") && !strings.HasSuffix(current, "_test.go") {
				if _, ok := result[current]; !ok {
					result[current] = nil
				}
			}
			continue
		}
		if !strings.HasPrefix(line, "@@ ") {
			continue
		}
		if current == "" {
			return nil, fmt.Errorf("diff hunk without target")
		}
		plus := strings.Index(line, "+")
		if plus < 0 {
			return nil, fmt.Errorf("malformed diff hunk %q", line)
		}
		end := strings.Index(line[plus:], " @@")
		if end < 0 {
			return nil, fmt.Errorf("malformed diff hunk %q", line)
		}
		spec := line[plus+1 : plus+end]
		parts := strings.Split(spec, ",")
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("malformed diff hunk %q", line)
		}
		count := 1
		if len(parts) == 2 {
			count, err = strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("malformed diff hunk %q", line)
			}
		}
		if len(parts) > 2 || count < 0 {
			return nil, fmt.Errorf("malformed diff hunk %q", line)
		}
		if count > 0 {
			result[current] = append(result[current], changedLineRange{Start: start, End: start + count - 1})
		}
	}
	return result, nil
}

func changedProductionFiles(root string) ([]string, []string, error) {
	changeSet, err := resolveHomeStateCoverageChangeSet(root)
	return changeSet.Native, changeSet.Disposition, err
}

func resolveHomeStateCoverageChangeSet(root string) (homeStateCoverageChangeSet, error) {
	var result homeStateCoverageChangeSet
	subjects := []string{
		homeStateCoverageCommitSubject,
		homeStateCoverageRemediationCommitSubject,
		homeStateCoverageDeltaCommitSubject,
	}
	commitsBySubject := make(map[string]string, len(subjects))
	log, err := gitCoverageOutput(root, "log", "--format=%H%x09%s", "HEAD")
	if err != nil {
		return result, err
	}
	for _, line := range strings.Split(strings.TrimSpace(log), "\n") {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		for _, subject := range subjects {
			if parts[1] != subject {
				continue
			}
			if commitsBySubject[subject] != "" {
				return result, fmt.Errorf("ambiguous home-state coverage evidence marker: %s", subject)
			}
			commitsBySubject[subject] = parts[0]
		}
	}
	originalTip := commitsBySubject[homeStateCoverageCommitSubject]
	if originalTip == "" {
		return result, fmt.Errorf("home-state coverage evidence commit not found")
	}
	result.Base, err = gitCoverageOutput(root, "rev-parse", originalTip+"^1")
	if err != nil {
		return result, fmt.Errorf("resolve home-state coverage base: %w", err)
	}
	var auditedCommits []string
	previous := ""
	missingEarlier := false
	for _, subject := range subjects {
		commit := commitsBySubject[subject]
		if commit == "" {
			missingEarlier = true
			continue
		}
		if missingEarlier {
			return result, fmt.Errorf("home-state coverage evidence chain has a missing predecessor: %s", subject)
		}
		if previous != "" {
			if err := gitCoverageRun(root, "merge-base", "--is-ancestor", previous, commit); err != nil {
				return result, fmt.Errorf("home-state coverage evidence does not descend from prior marker: %s: %w", subject, err)
			}
		}
		auditedCommits = append(auditedCommits, commit)
		previous = commit
	}
	result.Tip = auditedCommits[len(auditedCommits)-1]
	if err := gitCoverageRun(root, "merge-base", "--is-ancestor", result.Tip, "HEAD"); err != nil {
		return result, fmt.Errorf("home-state coverage tip is not an ancestor of HEAD: %w", err)
	}

	result.Ranges = map[string][]changedLineRange{}
	committedFiles := map[string]bool{}
	expectedBlobs := map[string]string{}
	for _, commit := range auditedCommits {
		parent, err := gitCoverageOutput(root, "rev-parse", commit+"^1")
		if err != nil {
			return result, fmt.Errorf("resolve coverage marker parent: %w", err)
		}
		status, err := gitCoverageOutput(root, "diff", "--name-status", "-M", parent, commit, "--", ":(glob)**/*.go")
		if err != nil {
			return result, err
		}
		files, err := productionFilesFromNameStatus(status)
		if err != nil {
			return result, err
		}
		diff, err := gitCoverageOutput(root, "diff", "--unified=0", "--no-color", parent, commit, "--", ":(glob)**/*.go")
		if err != nil {
			return result, err
		}
		ranges, err := parseUnifiedZeroDiff(diff)
		if err != nil {
			return result, err
		}
		for file := range files {
			committedFiles[file] = true
			blob, err := gitCoverageOutput(root, "rev-parse", commit+":"+file)
			if err != nil {
				return result, fmt.Errorf("read audited production blob %s: %w", file, err)
			}
			expectedBlobs[file] = blob
		}
		for file, changed := range ranges {
			if isProductionCoverageFile(file) {
				result.Ranges[file] = append(result.Ranges[file], changed...)
			}
		}
	}
	for file, expected := range expectedBlobs {
		headBlob, err := gitCoverageOutput(root, "rev-parse", "HEAD:"+file)
		if err != nil || headBlob != expected {
			return result, fmt.Errorf("audited production file changed after coverage tip: %s", file)
		}
	}

	dirtyStatus, err := gitCoverageOutput(root, "diff", "--name-status", "-M", "HEAD", "--", ":(glob)**/*.go")
	if err != nil {
		return result, err
	}
	dirtyFiles, err := productionFilesFromNameStatus(dirtyStatus)
	if err != nil {
		return result, err
	}
	dirtyDiff, err := gitCoverageOutput(root, "diff", "--unified=0", "--no-color", "HEAD", "--", ":(glob)**/*.go")
	if err != nil {
		return result, err
	}
	dirtyRanges, err := parseUnifiedZeroDiff(dirtyDiff)
	if err != nil {
		return result, err
	}
	for file, ranges := range dirtyRanges {
		if isProductionCoverageFile(file) {
			result.Ranges[file] = append(result.Ranges[file], ranges...)
		}
	}

	untracked, err := gitCoverageOutput(root, "ls-files", "--others", "--exclude-standard", "--", ":(glob)**/*.go")
	if err != nil {
		return result, err
	}
	all := committedFiles
	for file := range dirtyFiles {
		all[file] = true
	}
	for _, file := range strings.Fields(untracked) {
		file = filepath.ToSlash(file)
		if !isProductionCoverageFile(file) {
			continue
		}
		all[file] = true
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
		if err != nil {
			return result, err
		}
		result.Ranges[file] = []changedLineRange{{Start: 1, End: 1 + strings.Count(string(raw), "\n")}}
	}
	for file := range all {
		if _, ok := result.Ranges[file]; !ok {
			return result, fmt.Errorf("changed production file missing from diff: %s", file)
		}
		if strings.HasSuffix(file, "_windows.go") && runtime.GOOS != "windows" || strings.HasSuffix(file, "_unix.go") && runtime.GOOS == "windows" {
			result.Disposition = append(result.Disposition, file+":cross-compile")
			delete(result.Ranges, file)
		} else {
			result.Native = append(result.Native, file)
		}
	}
	if len(result.Native) == 0 {
		return result, fmt.Errorf("zero changed production files")
	}
	sort.Strings(result.Native)
	sort.Strings(result.Disposition)
	return result, nil
}

func gitCoverageOutput(root string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func gitCoverageRun(root string, args ...string) error {
	_, err := gitCoverageOutput(root, args...)
	return err
}

func productionFilesFromNameStatus(status string) (map[string]bool, error) {
	files := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(status), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("malformed git name-status row: %q", line)
		}
		if fields[0] == "D" {
			if isProductionCoverageFile(fields[1]) {
				return nil, fmt.Errorf("deleted production coverage target: %s", fields[1])
			}
			continue
		}
		file := fields[len(fields)-1]
		if isProductionCoverageFile(file) {
			files[filepath.ToSlash(file)] = true
		}
	}
	return files, nil
}

func isProductionCoverageFile(file string) bool {
	file = filepath.ToSlash(file)
	return strings.HasPrefix(file, "internal/") && strings.HasSuffix(file, ".go") && !strings.HasSuffix(file, "_test.go") && !strings.Contains(file, "/testdata/")
}

func parseChangedSurfaceCoverage(profilePath string, files []string) (float64, error) {
	ranges := map[string][]changedLineRange{}
	for _, file := range files {
		ranges[file] = []changedLineRange{{Start: 1, End: int(^uint(0) >> 1)}}
	}
	result, err := parseChangedLineCoverage(profilePath, ranges)
	return result.Percent, err
}

func parseChangedLineCoverage(profilePath string, ranges map[string][]changedLineRange) (changedCoverageResult, error) {
	f, err := os.Open(profilePath)
	if err != nil {
		return changedCoverageResult{}, err
	}
	defer func() { _ = f.Close() }()
	wanted := ranges
	type block struct{ statements, count int }
	blocks := map[string]block{}
	seenFiles := map[string]bool{}
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() || !strings.HasPrefix(scanner.Text(), "mode:") {
		return changedCoverageResult{}, fmt.Errorf("invalid coverage profile header")
	}
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 3 {
			return changedCoverageResult{}, fmt.Errorf("invalid coverage row")
		}
		colon := strings.LastIndex(fields[0], ":")
		if colon < 0 {
			return changedCoverageResult{}, fmt.Errorf("invalid coverage location")
		}
		file := filepath.ToSlash(fields[0][:colon])
		matched := ""
		for candidate := range wanted {
			if strings.HasSuffix(file, candidate) {
				matched = candidate
				break
			}
		}
		if matched == "" {
			continue
		}
		seenFiles[matched] = true
		location := fields[0][colon+1:]
		comma := strings.Index(location, ",")
		dot1 := strings.Index(location, ".")
		dot2 := -1
		if comma >= 0 {
			dot2 = strings.Index(location[comma+1:], ".")
		}
		if comma < 0 || dot1 < 0 || dot2 < 0 {
			return changedCoverageResult{}, fmt.Errorf("invalid coverage range")
		}
		startLine, e1 := strconv.Atoi(location[:dot1])
		endLine, e2 := strconv.Atoi(location[comma+1 : comma+1+dot2])
		if e1 != nil || e2 != nil || startLine < 1 || endLine < startLine {
			return changedCoverageResult{}, fmt.Errorf("invalid coverage range")
		}
		statements, err1 := strconv.Atoi(fields[1])
		count, err2 := strconv.Atoi(fields[2])
		if err1 != nil || err2 != nil || statements < 0 || count < 0 {
			return changedCoverageResult{}, fmt.Errorf("invalid coverage counters: %q", scanner.Text())
		}
		overlaps := false
		for _, changed := range wanted[matched] {
			if startLine <= changed.End && endLine >= changed.Start {
				overlaps = true
				break
			}
		}
		if !overlaps {
			continue
		}
		key := matched + fields[0][colon:]
		prior := blocks[key]
		if prior.statements != 0 && prior.statements != statements {
			return changedCoverageResult{}, fmt.Errorf("tampered duplicate coverage block")
		}
		if count > prior.count {
			prior.count = count
		}
		prior.statements = statements
		blocks[key] = prior
	}
	if err := scanner.Err(); err != nil {
		return changedCoverageResult{}, err
	}
	for file := range wanted {
		if !seenFiles[file] {
			return changedCoverageResult{}, fmt.Errorf("coverage profile missing %s", file)
		}
	}
	total, covered := 0, 0
	for _, row := range blocks {
		total += row.statements
		if row.count > 0 {
			covered += row.statements
		}
	}
	if total == 0 {
		return changedCoverageResult{}, fmt.Errorf("zero changed-surface statements")
	}
	return changedCoverageResult{Covered: covered, Total: total, Percent: 100 * float64(covered) / float64(total)}, nil
}
