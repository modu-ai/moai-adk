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
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Linkage and capacity warning codes.
const (
	WarnOrphanNotIndexed  AuditCode = "MEMORY_ORPHAN_NOT_INDEXED"
	WarnDanglingIndexLink AuditCode = "MEMORY_DANGLING_INDEX_LINK"
	WarnTopicCountOverCap AuditCode = "MEMORY_TOPIC_COUNT_OVER_CAP"
)

// indexFileName is the index a session actually loads.
const indexFileName = "MEMORY.md"

// archiveDirName holds retired topic files. Archiving is the remedy the cap
// prescribes, so its contents are excluded from both audits — otherwise the
// remedy would trip the alarm it is meant to clear.
const archiveDirName = "_archive"

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

// AuditLinkage reports topic files the index cannot reach (orphans) and index
// links with no file behind them (dangling). Both directions matter: an orphan
// is a memory that will never be recalled, and a dangling link is an index
// promising context that cannot be loaded.
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

	var findings []AuditFinding
	for _, n := range names {
		if !targets[n] {
			findings = append(findings, AuditFinding{
				Code:   WarnOrphanNotIndexed,
				Path:   filepath.Join(dir, n),
				Detail: fmt.Sprintf("%s is not linked from %s — a session loads the index, so this memory is never recalled", n, indexFileName),
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
