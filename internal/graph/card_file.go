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

// commitInfo is one commit of the reachable walk: full SHA and subject.
type commitInfo struct {
	sha, subject string
}

// walkCardCommits lists every commit reachable from HEAD by ANY parent path —
// merge and single-parent (squash) commits alike — in git log order (newest
// first). Subject attribution is the only card filter; the walk carries no
// shape filter of its own (SPEC-GRAPH-CARD-SQUASH-EDGE-001 REQ-GCSE-001).
// `git log` (not `rev-list --format`) keeps one line per commit —
// NUL-separated so a subject carrying spaces or format metacharacters still
// parses.
func walkCardCommits(repoRoot string) ([]commitInfo, error) {
	out, err := gitIn(repoRoot, "log", "--format=%H%x00%s", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("card_file: log: %w", err)
	}
	var commits []commitInfo
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		sha, subject, ok := strings.Cut(line, "\x00")
		if !ok || sha == "" {
			continue
		}
		commits = append(commits, commitInfo{sha: sha, subject: subject})
	}
	return commits, nil
}

// CardFileEdges walks every commit reachable from HEAD by ANY parent path —
// merge and single-parent (squash) commits alike — keeps those whose subject
// the attributor maps to exactly one card, and returns one edge per
// (landing, changed file) pair against the landing's FIRST PARENT — the
// contribution the landing brought in. The ^1 diff of a parentless (root)
// commit fails, and the per-commit guard below turns that into no edge and
// no error. Deterministic: two runs over the same tree and reachable history
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
	var edges []CardFileEdge
	for _, m := range commits {
		cardID := attribute(m.subject, landedBranch)
		if cardID == "" {
			continue
		}
		// First-parent diff: the contribution THIS landing brought in,
		// independent of which path it was reached by. For a parentless root
		// the ^1 revision fails and the guard below skips the commit — no
		// edge, no error. NUL-separated: git's core.quotepath default
		// C-escapes a non-ASCII path in the line output, and splitting on
		// newlines stored the escape as the file (card t1454 card-review r2
		// finding 16).
		files, err := gitIn(repoRoot, "diff", "--name-only", "-z", m.sha+"^1", m.sha)
		if err != nil {
			continue // an unreachable, shallow-clone, or parentless commit contributes no edges
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
