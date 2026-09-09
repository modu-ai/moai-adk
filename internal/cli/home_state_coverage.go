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

func measureChangedSurfaceCoverage(ctx context.Context, root string) (float64, error) {
	result, err := measureChangedSurfaceCoverageResultWith(ctx, root, runChangedSurfaceCoverageSuite)
	return result.Percent, err
}

func runChangedSurfaceCoverageSuite(ctx context.Context, root, path string) error {
	pattern := "^(TestHomeState.*|TestHomeLayout.*|TestProject(Layout|Dir).*|TestRuntimeCensus.*|TestAdmission.*|TestFactory.*|TestResolveFactory.*|TestResume.*|TestExpireResume.*|TestImportLegacy.*|TestHandoff.*|TestClaim.*|TestClaimThenInject.*|TestConcurrentConsume.*|TestSQLiteClaim.*|TestNonceFallback.*|TestManualMode.*|TestNonClearSource.*|TestDegradeToGuidance.*|TestFailOpen_CorruptPending.*|TestStaleTTL.*|TestBranchTable.*|TestRenderHandoff.*|TestHandle_.*|TestIsHex.*|TestInjectionHeader.*|TestNoUserInteraction.*|TestThreeHandler.*|TestProfile.*|TestCleanHome.*|TestScanHomeCleanable.*|TestSecureHomeDirectories.*|TestContinue.*|TestCC.*|TestCharacterize_CC.*|TestRunCC.*|TestRunGLM.*|TestSession(Start|End).*|TestPersistedHomeStateEvidence.*|TestPlatformProcessIdentity.*|TestParseChanged.*|TestChangedProduction.*)$"
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
	defer os.Remove(path)
	if err := run(ctx, root, path); err != nil {
		return changedCoverageResult{}, err
	}
	files, _, err := changedProductionFiles(root)
	if err != nil {
		return changedCoverageResult{}, err
	}
	ranges, err := changedProductionLineRanges(root, files)
	if err != nil {
		return changedCoverageResult{}, err
	}
	result, err := parseChangedLineCoverage(path, ranges)
	if err != nil {
		return changedCoverageResult{}, err
	}
	return result, nil
}

func changedProductionLineRanges(root string, files []string) (map[string][]changedLineRange, error) {
	out, err := exec.Command("git", "-C", root, "diff", "--unified=0", "--no-color", "HEAD", "--", ":(glob)**/*.go").Output()
	if err != nil {
		return nil, err
	}
	ranges, err := parseUnifiedZeroDiff(string(out))
	if err != nil {
		return nil, err
	}
	wanted := map[string]bool{}
	for _, file := range files {
		wanted[file] = true
	}
	for file := range ranges {
		if !wanted[file] {
			delete(ranges, file)
		}
	}
	nameStatus, err := exec.Command("git", "-C", root, "diff", "--name-status", "-M", "HEAD", "--", ":(glob)**/*.go").Output()
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(strings.TrimSpace(string(nameStatus)), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 3 && strings.HasPrefix(fields[0], "R") && wanted[filepath.ToSlash(fields[2])] {
			if _, ok := ranges[filepath.ToSlash(fields[2])]; !ok {
				ranges[filepath.ToSlash(fields[2])] = nil
			}
		}
	}
	tracked := map[string]bool{}
	for file := range ranges {
		tracked[file] = true
	}
	for _, file := range files {
		if _, ok := ranges[file]; ok {
			continue
		}
		status, err := exec.Command("git", "-C", root, "ls-files", "--others", "--exclude-standard", "--", file).Output()
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(string(status)) == file {
			raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
			if err != nil {
				return nil, err
			}
			lines := 1 + strings.Count(string(raw), "\n")
			ranges[file] = []changedLineRange{{Start: 1, End: lines}}
			continue
		}
		if !tracked[file] {
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
	commands := [][]string{{"diff", "--name-only", "--diff-filter=ACMR", "HEAD", "--", ":(glob)**/*.go"}, {"ls-files", "--others", "--exclude-standard", "--", ":(glob)**/*.go"}}
	all := map[string]bool{}
	for _, args := range commands {
		out, err := exec.Command("git", append([]string{"-C", root}, args...)...).Output()
		if err != nil {
			return nil, nil, err
		}
		for _, file := range strings.Fields(string(out)) {
			file = filepath.ToSlash(file)
			if strings.Contains(file, "/testdata/") || strings.HasSuffix(file, "_test.go") || !strings.HasPrefix(file, "internal/") {
				continue
			}
			all[file] = true
		}
	}
	statusOut, err := exec.Command("git", "-C", root, "diff", "--name-status", "-M", "HEAD", "--", ":(glob)**/*.go").Output()
	if err != nil {
		return nil, nil, err
	}
	var deleted []string
	for _, line := range strings.Split(strings.TrimSpace(string(statusOut)), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "D" {
			deleted = append(deleted, fields[1])
		}
	}
	if len(deleted) != 0 {
		return nil, nil, fmt.Errorf("deleted production coverage target: %s", strings.Join(deleted, ","))
	}
	var native, disposition []string
	for file := range all {
		if strings.HasSuffix(file, "_windows.go") && runtime.GOOS != "windows" || strings.HasSuffix(file, "_unix.go") && runtime.GOOS == "windows" {
			disposition = append(disposition, file+":cross-compile")
		} else {
			native = append(native, file)
		}
	}
	if len(native) == 0 {
		return nil, nil, fmt.Errorf("zero changed production files")
	}
	sort.Strings(native)
	sort.Strings(disposition)
	return native, disposition, nil
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
	defer f.Close()
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
