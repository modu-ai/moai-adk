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

	// WarnRepoRelativeLink reports a repo-relative link target carried by
	// an index (SPEC-MEMORY-FOLD-BUDGET-001 REQ-MFB-011). The target names
	// a path outside the store, so it can never resolve here: it is not
	// dangling, and the link-repair follow-up card owns any rewrite — the
	// doctor only classifies and reports (spec.md §4).
	WarnRepoRelativeLink AuditCode = "MEMORY_REPO_RELATIVE_LINK"
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
// and a topic file citing a couple of relatives is not one of them.
//
// The threshold counts LINK MATCHES, not lines carrying a link, and the two
// give different answers: a card record citing two siblings on one line reads
// as 1 by line and 2 by match. Measured over the live store by match, the real
// index files carry 42 links at the low end while the ordinary topic files
// that cite anything at all carry 1 or 2 — so 3 sits in a gap fourteen times
// wider than the boundary case it has to exclude.
//
// Counting by line put the threshold at 2, which admitted exactly one card
// record and hid the two siblings it cited: both were reachable from nothing
// else, so the orphan finding went silent on two real losses. A false negative
// here is worse than a false positive — an unreported loss is the failure this
// audit exists to catch — which is why the margin is spent on that side.
const secondaryIndexLinkThreshold = 3

// markdownLinkTarget captures the target of a markdown link. The index format
// is one `- [Title](file.md) — hook` line per memory, so the targets are the
// set of files the index can reach.
var markdownLinkTarget = regexp.MustCompile(`\]\(([^)]+\.md)\)`)

// SecondaryIndexLinkThreshold exposes the secondary-index threshold so every
// reader of the rule — the doctor above and the fold's reachability checker
// (SPEC-MEMORY-FOLD-BUDGET-001 §1.5) — reads one number. The checker builds
// its index set from this accessor, so the threshold cannot drift between the
// doctor and the checker, and the boundary test is written against it rather
// than against a literal.
func SecondaryIndexLinkThreshold() int {
	return secondaryIndexLinkThreshold
}

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

// indexTargets returns the set of .md files the index links to, keyed by
// base name — the doctor's resolution rule for orphans and duplicates — plus
// the index's raw link targets in order of appearance, which the class-aware
// dangling pass classifies by full text (REQ-MFB-011).
func indexTargets(indexPath string) (map[string]bool, []string, error) {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]bool{}, nil, nil
		}
		return nil, nil, fmt.Errorf("taxonomy: read index %s: %w", indexPath, err)
	}

	targets := map[string]bool{}
	for _, m := range markdownLinkTarget.FindAllStringSubmatch(string(data), -1) {
		// Index entries are written as bare filenames; tolerate a relative
		// prefix so a hand-edited `./name.md` still resolves.
		targets[filepath.Base(m[1])] = true
	}
	return targets, ExtractLinkTargets(string(data)), nil
}

// secondaryIndex is one promoted index: the sibling files it actually reaches,
// and the links it carries that no file answers.
type secondaryIndex struct {
	reaches map[string]bool
	missing []string
	// repoRelative holds the distinct repo-relative targets the index
	// carries, full target text, sorted (REQ-MFB-011). They are reported
	// under their own code and never as dangling; the qualification above
	// is unchanged and keeps resolving by base name.
	repoRelative []string
}

// secondaryIndexTargets returns, for each topic file that qualifies as a
// secondary index, what it reaches and what it only promises. A file that
// cannot be read is treated as an ordinary memory rather than as an error: a
// store the audit cannot fully read should report what it can, not refuse to
// report.
//
// Only links with a file behind them count toward the threshold. Counting raw
// matches let a file citing absent siblings buy index status, and the purchase
// was self-concealing: the links that bought it were also the links no check
// read, so a real memory the same file linked went silent behind an index that
// reached nothing. Applying the existing threshold to resolved links closes
// that without a second number to justify — a promotion is a claim to carry
// reachability, and a link to a file that does not exist carries none.
//
// Classification is by FULL target text (REQ-MFB-011): only a store-local
// target resolves against the store, so only store-local targets feed
// reaches/missing — a repo-relative target that happens to share a base name
// with a store file neither buys promotion nor silences the store-local
// dangling warning for that name.
func secondaryIndexTargets(dir string, names []string, present map[string]bool) map[string]secondaryIndex {
	out := map[string]secondaryIndex{}
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		seenLocal := map[string]bool{}
		idx := secondaryIndex{reaches: map[string]bool{}}
		for _, m := range markdownLinkTarget.FindAllStringSubmatch(string(data), -1) {
			if ClassifyLinkTarget(m[1]) != LinkStoreLocal {
				continue
			}
			base := filepath.Base(m[1])
			// A file linking itself reaches nothing, and the index a session
			// already loads is not a secondary one.
			if base == name || base == indexFileName || seenLocal[base] {
				continue
			}
			seenLocal[base] = true
			if present[base] {
				idx.reaches[base] = true
			} else {
				idx.missing = append(idx.missing, base)
			}
		}
		idx.repoRelative = indexRepoRelativeTargets(data)

		if len(idx.reaches) >= secondaryIndexLinkThreshold {
			sort.Strings(idx.missing)
			out[name] = idx
		}
	}
	return out
}

// indexRepoRelativeTargets returns the distinct repo-relative link targets
// one index file carries, full target text, sorted. Collection is separate
// from the base-name loop above so the qualification rule (reaches counts
// resolved store-local links by base name) stays byte-for-byte what it was.
func indexRepoRelativeTargets(data []byte) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range markdownLinkTarget.FindAllStringSubmatch(string(data), -1) {
		if ClassifyLinkTarget(m[1]) != LinkRepoRelative || seen[m[1]] {
			continue
		}
		seen[m[1]] = true
		out = append(out, m[1])
	}
	sort.Strings(out)
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

	// The index is read regardless of topic files (REQ-MFB-011): an index
	// with links but an empty directory around it is exactly a store of
	// dangling promises, and repo-relative targets are classified whether or
	// not any topic file exists. A missing directory yields empty names and
	// an unreadable index — silence falls out of the empty sets below.
	indexPath := filepath.Join(dir, indexFileName)
	targets, ordered, err := indexTargets(indexPath)
	if err != nil {
		return nil, err
	}

	present := make(map[string]bool, len(names))
	for _, n := range names {
		present[n] = true
	}
	// The index itself is present — a MEMORY.md link naming MEMORY.md (a
	// store whose only file is the index) resolves; reporting it as
	// dangling reported a file that exists (codex-review round 2).
	present[indexFileName] = true

	secondary := secondaryIndexTargets(dir, names, present)

	var findings []AuditFinding
	for _, n := range names {
		// Deterministic order: a finding naming an arbitrary one of several
		// secondary indexes would change between runs on the same store.
		var carriers []string
		for idx, sec := range secondary {
			if idx != n && sec.reaches[n] {
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
	// The class-aware dangling pass over the index a session loads
	// (REQ-MFB-011): a store-local target with no file behind it dangles as
	// before; a repo-relative target is reported once per distinct target,
	// full text, under its own code and never as dangling.
	findings = append(findings, linkFindings(indexPath, ordered, present)...)

	// The same direction, one carrier over: a secondary index promises files
	// too, and reading only MEMORY.md left those promises unchecked. Reported
	// under the same code because it is the same defect — an index line with
	// no file behind it — differing only in which index carries it.
	var carriers []string
	for idx := range secondary {
		carriers = append(carriers, idx)
	}
	sort.Strings(carriers)
	for _, idx := range carriers {
		sec := secondary[idx]
		// missing holds store-local targets only (full-path classification
		// in secondaryIndexTargets), so every entry here is a true dangling
		// link — no basename suppression needed.
		for _, target := range sec.missing {
			findings = append(findings, AuditFinding{
				Code:   WarnDanglingIndexLink,
				Path:   filepath.Join(dir, idx),
				Detail: fmt.Sprintf("index links %s but no such file exists", target),
			})
		}
		for _, target := range sec.repoRelative {
			findings = append(findings, AuditFinding{
				Code:   WarnRepoRelativeLink,
				Path:   filepath.Join(dir, idx),
				Detail: fmt.Sprintf("link target %s is repo-relative — it names no file inside the store; reported as classified, never rewritten", target),
			})
		}
	}
	return findings, nil
}

// linkFindings walks one index's link targets in order of appearance and
// returns the class-aware per-target findings (REQ-MFB-011): a store-local
// target with no file behind it dangles under the existing code and text; a
// repo-relative target is reported once per distinct target, full text, under
// MEMORY_REPO_RELATIVE_LINK and never as dangling; an absolute target is
// classified, and REQ-MFB-011 prescribes no finding for it.
func linkFindings(carrierPath string, targets []string, present map[string]bool) []AuditFinding {
	var findings []AuditFinding
	seen := map[string]bool{}
	for _, t := range targets {
		if seen[t] {
			continue
		}
		seen[t] = true
		switch ClassifyLinkTarget(t) {
		case LinkStoreLocal:
			n := storeLocalName(t)
			if !present[n] {
				findings = append(findings, AuditFinding{
					Code:   WarnDanglingIndexLink,
					Path:   carrierPath,
					Detail: fmt.Sprintf("index links %s but no such file exists", n),
				})
			}
		case LinkRepoRelative:
			findings = append(findings, AuditFinding{
				Code:   WarnRepoRelativeLink,
				Path:   carrierPath,
				Detail: fmt.Sprintf("link target %s is repo-relative — it names no file inside the store; reported as classified, never rewritten", t),
			})
		case LinkAbsolute:
			// Classified; no finding is prescribed.
		}
	}
	return findings
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
