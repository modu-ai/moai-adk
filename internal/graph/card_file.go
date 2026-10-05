// card_file.go — SPEC-TODO-CARD-ISSUANCE-001 M3 (REQ-TCI-016): the
// card→file edge layer. Edges derive ONLY from committed evidence — merge
// commits whose subject attributes them to exactly one card (via the
// attribution function the caller supplies, wired to the engine's single
// attribution point) and that are reachable from HEAD by ANY parent path —
// plus the files each such merge brought in (first-parent diff). The queue
// is never read: the layer's output is identical whether or not a queue
// exists (REQ-TCI-016 — no unlanded card id, no expected file, no finding).
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

// CardFileAttributor attributes one merge subject to exactly one card id,
// or "" when the subject attributes nothing. The caller wires this to the
// engine's single attribution point; card_file.go never guesses.
type CardFileAttributor func(subject, landedBranch string) string

// CardFileEdge is one card→file edge derived from a card-attributed merge.
type CardFileEdge struct {
	Card string
	File string
	SHA  string // the merge commit's abbreviated SHA (the evidence pointer)
}

// mergeInfo is one merge commit of the reachable walk: full SHA and subject.
type mergeInfo struct {
	sha, subject string
}

// walkCardMerges lists every merge commit reachable from HEAD by ANY parent
// path, in git log order (newest first). `git log` (not `rev-list --format`)
// keeps one line per commit — NUL-separated so a subject carrying spaces or
// format metacharacters still parses.
func walkCardMerges(repoRoot string) ([]mergeInfo, error) {
	out, err := gitIn(repoRoot, "log", "--merges", "--format=%H%x00%s", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("card_file: log merges: %w", err)
	}
	var merges []mergeInfo
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		sha, subject, ok := strings.Cut(line, "\x00")
		if !ok || sha == "" {
			continue
		}
		merges = append(merges, mergeInfo{sha: sha, subject: subject})
	}
	return merges, nil
}

// CardFileEdges walks every merge commit reachable from HEAD by ANY parent
// path, keeps those whose subject the attributor maps to exactly one card,
// and returns one edge per (merge, changed file) pair against the merge's
// FIRST PARENT — the contribution the merge brought in. Deterministic: two
// runs over the same tree and reachable history return byte-identical
// output. Absorb-direction merges attribute nothing through the attributor.
//
// repoRoot is the worktree root the git commands run in; landedBranch is
// the integration branch name the attribution rule compares against.
func CardFileEdges(repoRoot, landedBranch string, attribute CardFileAttributor) ([]CardFileEdge, error) {
	if attribute == nil {
		return nil, fmt.Errorf("card_file: attribution function is required")
	}
	merges, err := walkCardMerges(repoRoot)
	if err != nil {
		return nil, err
	}
	var edges []CardFileEdge
	for _, m := range merges {
		cardID := attribute(m.subject, landedBranch)
		if cardID == "" {
			continue
		}
		// First-parent diff: the contribution THIS merge brought in,
		// independent of which path it was reached by.
		files, err := gitIn(repoRoot, "diff", "--name-only", m.sha+"^1", m.sha)
		if err != nil {
			continue // an unreachable or shallow-clone merge contributes no edges
		}
		for _, f := range strings.Split(strings.TrimSpace(files), "\n") {
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
// merge commits the attributor maps to a card — the freshness input of the
// card-file layer. A new card-attributed merge landing changes the list; an
// absorb-direction merge landing does not (it attributes nothing), so the
// fingerprint tracks edges content, not all of history.
func CardAttributedMergeSHAs(repoRoot, landedBranch string, attribute CardFileAttributor) ([]string, error) {
	if attribute == nil {
		return nil, fmt.Errorf("card_file: attribution function is required")
	}
	merges, err := walkCardMerges(repoRoot)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(merges))
	var shas []string
	for _, m := range merges {
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

// CardMergeFingerprint hashes the card-attributed merge SHA list into the
// stable fingerprint SourceFingerprintsForEdges stamps and re-checks. An
// empty list hashes to the empty-input digest — a stable, comparable state
// like the other source sets' absent-dir form.
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
