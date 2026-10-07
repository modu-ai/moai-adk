// card_file.go — SPEC-TODO-CARD-ISSUANCE-001 M3 (REQ-TCI-016), broadened by
// SPEC-GRAPH-CARD-SQUASH-EDGE-001: the card→file edge layer. Edges derive
// ONLY from committed evidence — commits whose subject attributes them to
// exactly one card (via the attribution function the caller supplies, wired
// to the engine's single attribution point) and that are reachable from HEAD
// by ANY parent path — plus the files each such landing brought in
// (first-parent diff). Landing commits include squash landings (single
// parent) as well as merge commits. The queue is never read: the layer's
// output is identical whether or not a queue exists (REQ-TCI-016 — no
// unlanded card id, no expected file, no finding).
//
// Cycles terminate through the visit set; absorb-direction merges attribute
// nothing through the attribution function's own non-attribution rule, so
// the walk's breadth (all parent paths) cannot make an absorb merge produce
// edges.
package graph

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// CardFileAttributor attributes one commit subject to exactly one card id,
// or "" when the subject attributes nothing. The caller wires this to the
// engine's single attribution point; card_file.go never guesses.
type CardFileAttributor func(subject, landedBranch string) string

// CardFileEdge is one card→file edge derived from a card-attributed landing
// commit (a squash landing or a merge).
type CardFileEdge struct {
	Card string
	File string
	SHA  string // the landing commit's abbreviated SHA (the evidence pointer)
}

// commitInfo is one commit of the reachable walk: full SHA, subject, and parents.
type commitInfo struct {
	sha, subject, parents string
}

// walkCardCommits lists every commit reachable from HEAD by ANY parent path —
// merge and single-parent (squash) commits alike — in git log order (newest
// first). Subject attribution is the only card filter; the walk carries no
// shape filter of its own (SPEC-GRAPH-CARD-SQUASH-EDGE-001 REQ-GCSE-001).
// `git log` (not `rev-list --format`) keeps one line per commit —
// NUL-separated so a subject carrying spaces or format metacharacters still
// parses.
func walkCardCommits(repoRoot string) ([]commitInfo, error) {
	out, err := gitIn(repoRoot, "log", "--format=%H%x00%s%x00%P", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("card_file: log: %w", err)
	}
	var commits []commitInfo
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, "\x00", 3)
		if len(fields) != 3 || fields[0] == "" {
			continue
		}
		commits = append(commits, commitInfo{sha: fields[0], subject: fields[1], parents: fields[2]})
	}
	return commits, nil
}

// CardFileEdges walks every commit reachable from HEAD by ANY parent path —
// merge and single-parent (squash) commits alike — keeps those whose subject
// the attributor maps to exactly one card, and returns one edge per
// (landing, changed file) pair against the landing's FIRST PARENT — the
// contribution the landing brought in. A parentless (root) commit contributes
// no edge and no error, matching the failed ^1 diff. Native Git batches the
// independent landing diffs; a batch failure falls back to isolated per-commit
// diffs. Deterministic: two runs over the same tree and reachable history
// return byte-identical output. Absorb-direction merges attribute nothing
// through the attributor.
//
// repoRoot is the worktree root the git commands run in; landedBranch is
// the integration branch name the attribution rule compares against.
func CardFileEdges(repoRoot, landedBranch string, attribute CardFileAttributor) ([]CardFileEdge, error) {
	if attribute == nil {
		return nil, fmt.Errorf("card_file: attribution function is required")
	}
	commits, err := walkCardCommits(repoRoot)
	if err != nil {
		return nil, err
	}
	var selected []commitInfo
	cards := make(map[string]string)
	var sentinel string
	for _, m := range commits {
		if m.parents == "" {
			sentinel = m.sha
		}
		if card := attribute(m.subject, landedBranch); card != "" {
			selected = append(selected, m)
			cards[m.sha] = card
		}
	}
	filesBySHA, batchErr := cardFileBatch(repoRoot, selected, sentinel)
	var edges []CardFileEdge
	for _, m := range selected {
		cardID := cards[m.sha]
		// First-parent diff: the contribution THIS landing brought in,
		// independent of which path it was reached by. For a parentless root
		// the ^1 revision fails and the guard below skips the commit — no
		// edge, no error. NUL-separated: git's core.quotepath default
		// C-escapes a non-ASCII path in the line output, and splitting on
		// newlines stored the escape as the file (card t1454 card-review r2
		// finding 16).
		files := filesBySHA[m.sha]
		if batchErr != nil {
			// Preserve per-commit failure isolation when batch Git is unavailable
			// or its output is incomplete (including older Git versions).
			files, err = gitIn(repoRoot, "diff", "--name-only", "-z", m.sha+"^1", m.sha)
			if err != nil {
				continue
			}
		}
		for _, f := range strings.Split(files, "\x00") {
			if f == "" {
				continue
			}
			edges = append(edges, CardFileEdge{Card: cardID, File: f, SHA: m.sha[:9]})
		}
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Card != edges[j].Card {
			return edges[i].Card < edges[j].Card
		}
		if edges[i].File != edges[j].File {
			return edges[i].File < edges[j].File
		}
		return edges[i].SHA < edges[j].SHA
	})
	return edges, nil
}

// cardFileBatch asks native Git for independent first-parent diffs without
// walking again or spawning one process per landing. The empty NUL token marks
// a commit boundary; paths are never trimmed or split on newlines. A root's
// files are deliberately ignored, matching the failed ^1 diff in the fallback.
// A reachable root is appended last as an ignored terminal record: seeing its
// complete header proves all preceding landing records reached the reader.
func cardFileBatch(root string, commits []commitInfo, sentinel string) (map[string]string, error) {
	files := make(map[string]string, len(commits))
	if len(commits) == 0 {
		return files, nil
	}
	if sentinel == "" {
		return nil, fmt.Errorf("card_file: missing batch sentinel")
	}
	var input strings.Builder
	for _, c := range commits {
		if c.sha == sentinel {
			continue
		}
		input.WriteString(c.sha)
		input.WriteByte('\n')
	}
	input.WriteString(sentinel + "\n")
	cmd := exec.Command("git", "-C", root, "log", "--no-walk=unsorted", "--stdin",
		"--diff-merges=first-parent", "--name-only", "-z", "--format=%x00%H%x00%P")
	cmd.Stdin = strings.NewReader(input.String())
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	if len(out) == 0 || out[len(out)-1] != 0 {
		return nil, fmt.Errorf("card_file: unterminated batch output")
	}
	tokens := strings.Split(string(out), "\x00")
	var lastSHA string
	for i := 0; i < len(tokens)-1; {
		// The parent field needs its own terminator; Split's final empty
		// token alone cannot prove a complete empty-parent root header.
		if tokens[i] != "" || i+3 >= len(tokens) {
			return nil, fmt.Errorf("card_file: invalid batch header")
		}
		sha, parents := tokens[i+1], tokens[i+2]
		lastSHA = sha
		i += 3
		var paths []string
		for first := true; i < len(tokens) && tokens[i] != ""; i++ {
			path := tokens[i]
			if first {
				if !strings.HasPrefix(path, "\n") {
					return nil, fmt.Errorf("card_file: invalid batch path separator")
				}
				path = strings.TrimPrefix(path, "\n")
				first = false
			}
			paths = append(paths, path)
		}
		if parents != "" && len(paths) != 0 {
			files[sha] = strings.Join(paths, "\x00")
		} else {
			files[sha] = ""
		}
	}
	if lastSHA != sentinel {
		return nil, fmt.Errorf("card_file: missing terminal batch sentinel")
	}
	for _, c := range commits {
		if _, ok := files[c.sha]; !ok {
			return nil, fmt.Errorf("card_file: incomplete batch output")
		}
	}
	return files, nil
}

// CardAttributedMergeSHAs returns the sorted, deduplicated full SHAs of the
// commits the attributor maps to a card — the freshness input of the
// card-file layer. The list covers squash landings (single-parent commits)
// as well as merge commits. A new card-attributed landing changes the list;
// an absorb-direction merge landing does not (it attributes nothing), so the
// fingerprint tracks edges content, not all of history.
func CardAttributedMergeSHAs(repoRoot, landedBranch string, attribute CardFileAttributor) ([]string, error) {
	if attribute == nil {
		return nil, fmt.Errorf("card_file: attribution function is required")
	}
	commits, err := walkCardCommits(repoRoot)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(commits))
	var shas []string
	for _, m := range commits {
		if attribute(m.subject, landedBranch) == "" {
			continue
		}
		if !seen[m.sha] {
			seen[m.sha] = true
			shas = append(shas, m.sha)
		}
	}
	sort.Strings(shas)
	return shas, nil
}

// CardMergeFingerprint hashes the card-attributed landing SHA list (squash
// and merge landings alike) into the stable fingerprint
// SourceFingerprintsForEdges stamps and re-checks. An empty list hashes to
// the empty-input digest — a stable, comparable state like the other source
// sets' absent-dir form.
func CardMergeFingerprint(repoRoot, landedBranch string, attribute CardFileAttributor) (string, error) {
	shas, err := CardAttributedMergeSHAs(repoRoot, landedBranch, attribute)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(strings.Join(shas, "\n")))
	return hex.EncodeToString(sum[:]), nil
}

// cardFileLandedBranch resolves the integration branch the attribution rule
// compares against, through the same three-level ref chain the landed
// question asks. One resolution, shared by the layer and the fingerprint.
func cardFileLandedBranch(projectRoot string) string {
	return factory.LandedBranchFromRef(factory.LandedRefFor(projectRoot))
}

func gitIn(root string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	out, err := cmd.Output()
	return string(out), err
}
