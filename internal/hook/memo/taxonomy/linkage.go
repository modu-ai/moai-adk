// linkage.go — index-linkage and topic-count audits.
//
// AuditFile/AuditIndex/AuditDuplicates validate a memory file's own shape and
// the index's length. Neither reads the index's links, so a topic file written
// without its index line passes every existing check while being unreachable:
// a session loads MEMORY.md, not the directory, so an unlinked file is stored
// and never recalled. AuditLinkage closes that gap in both directions, and
// AuditTopicCount gives the per-project ceiling a checker.
package taxonomy

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Linkage and capacity warning codes.
const (
	WarnOrphanNotIndexed    AuditCode = "MEMORY_ORPHAN_NOT_INDEXED"
	WarnDanglingIndexLink   AuditCode = "MEMORY_DANGLING_INDEX_LINK"
	WarnIndexDuplicateEntry AuditCode = "MEMORY_INDEX_DUPLICATE_ENTRY"
	WarnTopicCountOverCap   AuditCode = "MEMORY_TOPIC_COUNT_OVER_CAP"
)

// indexFileName is the index a session actually loads.
const indexFileName = "MEMORY.md"

// archiveDirName holds retired topic files. Archiving is the remedy the cap
// prescribes, so its contents are excluded from both audits — otherwise the
// remedy would trip the alarm it is meant to clear.
const archiveDirName = "_archive"

// secondaryIndexLinkThreshold separates an index from a memory that happens to
// cite a sibling. Archiving in practice folds an entry out of MEMORY.md into a
// secondary index that sits beside the topic files rather than moving the file
// into a subdirectory, so reachability has to be read from those files too —
// and a topic file citing one relative is not one of them. Measured over the
// live store, the separation is clean rather than marginal: the real index
// files carry 3 links at the low end, while the ordinary topic files that
// carry any sibling link at all carry exactly 1.
const secondaryIndexLinkThreshold = 2

// markdownLinkTarget captures the target of a markdown link. The index format
// is one `- [Title](file.md) — hook` line per memory, so the targets are the
// set of files the index can reach.
var markdownLinkTarget = regexp.MustCompile(`\]\(([^)]+\.md)\)`)

// topicFiles returns the memory topic files directly under dir: .md files,
// excluding the index itself. A missing directory yields no files and no
// error — a project that has not written a memory yet is not a defect.
func topicFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("taxonomy: read memory dir %s: %w", dir, err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || e.Name() == indexFileName {
			continue
		}
		names = append(names, e.Name())
	}
	return names, nil
}

// indexTargets returns the set of .md files the index links to.
func indexTargets(indexPath string) (map[string]bool, error) {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]bool{}, nil
		}
		return nil, fmt.Errorf("taxonomy: read index %s: %w", indexPath, err)
	}

	targets := map[string]bool{}
	for _, m := range markdownLinkTarget.FindAllStringSubmatch(string(data), -1) {
		// Index entries are written as bare filenames; tolerate a relative
		// prefix so a hand-edited `./name.md` still resolves.
		targets[filepath.Base(m[1])] = true
	}
	return targets, nil
}

// secondaryIndexTargets returns, for each topic file that qualifies as a
// secondary index, the set of sibling files it links. A file that cannot be
// read is treated as an ordinary memory rather than as an error: a store the
// audit cannot fully read should report what it can, not refuse to report.
func secondaryIndexTargets(dir string, names []string) map[string]map[string]bool {
	out := map[string]map[string]bool{}
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		targets := map[string]bool{}
		for _, m := range markdownLinkTarget.FindAllStringSubmatch(string(data), -1) {
			base := filepath.Base(m[1])
			// A file linking itself reaches nothing, and the index a session
			// already loads is not a secondary one.
			if base == name || base == indexFileName {
				continue
			}
			targets[base] = true
		}
		if len(targets) >= secondaryIndexLinkThreshold {
			out[name] = targets
		}
	}
	return out
}

// AuditLinkage reports topic files no index can reach (orphans), index links
// with no file behind them (dangling), and files carried by two indexes at
// once (duplicate entries). All three matter: an orphan is a memory that will
// never be recalled, a dangling link is an index promising context that cannot
// be loaded, and a duplicate entry is the residue of a revived memory whose
// archive line was never removed — reachable twice, so silent in both of the
// other directions.
func AuditLinkage(dir string) ([]AuditFinding, error) {
	names, err := topicFiles(dir)
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		// No topic files: a dangling link cannot be distinguished from a
		// directory that was never populated, so stay silent.
		return nil, nil
	}

	indexPath := filepath.Join(dir, indexFileName)
	targets, err := indexTargets(indexPath)
	if err != nil {
		return nil, err
	}

	present := make(map[string]bool, len(names))
	for _, n := range names {
		present[n] = true
	}

	secondary := secondaryIndexTargets(dir, names)

	var findings []AuditFinding
	for _, n := range names {
		// Deterministic order: a finding naming an arbitrary one of several
		// secondary indexes would change between runs on the same store.
		var carriers []string
		for idx, reach := range secondary {
			if idx != n && reach[n] {
				carriers = append(carriers, idx)
			}
		}
		sort.Strings(carriers)

		switch {
		case !targets[n] && len(carriers) == 0:
			findings = append(findings, AuditFinding{
				Code:   WarnOrphanNotIndexed,
				Path:   filepath.Join(dir, n),
				Detail: fmt.Sprintf("%s is linked from no index — a session loads an index, not the directory, so this memory is never recalled", n),
			})
		case targets[n] && len(carriers) > 0:
			findings = append(findings, AuditFinding{
				Code:   WarnIndexDuplicateEntry,
				Path:   filepath.Join(dir, n),
				Detail: fmt.Sprintf("%s is linked from both %s and %s — one of the two entries is stale", n, indexFileName, strings.Join(carriers, ", ")),
			})
		}
	}
	for target := range targets {
		if !present[target] {
			findings = append(findings, AuditFinding{
				Code:   WarnDanglingIndexLink,
				Path:   indexPath,
				Detail: fmt.Sprintf("index links %s but no such file exists", target),
			})
		}
	}
	return findings, nil
}

// AuditTopicCount reports when the directory holds more topic files than cap.
// A non-positive cap falls back to the configured default.
func AuditTopicCount(dir string, cap int) ([]AuditFinding, error) {
	if cap <= 0 {
		cap = config.DefaultMemoryTopicFileCap
	}

	names, err := topicFiles(dir)
	if err != nil {
		return nil, err
	}
	if len(names) <= cap {
		return nil, nil
	}

	return []AuditFinding{{
		Code:   WarnTopicCountOverCap,
		Path:   dir,
		Detail: fmt.Sprintf("%d topic files; cap is %d — archive %d into %s/", len(names), cap, len(names)-cap, archiveDirName),
	}}, nil
}
