// todo_issuance.go — SPEC-TODO-CARD-ISSUANCE-001 M1: the CLI side of the
// issuance presentation. The factory layer (backlog_issuance.go) is pure;
// this file owns the three things that touch the world: the renderer, the
// time-bounded completed-SPEC directory read, and the production lane-branch
// probe. All of them run OUTSIDE the queue's cross-process write lock
// (REQ-TCI-003).
package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// todoLaneFilesProbe is the production lane-branch probe seam. Tests swap
// it; the add path reads it once per presentation build.
var todoLaneFilesProbe = productionLaneFilesProbe

// todoCompletedSpecsReader is the completed-SPEC directory read seam.
var todoCompletedSpecsReader = loadCompletedSpecs

// renderIssuanceText renders the presentation to its line form. An empty
// presentation renders as "" — the CLI prints nothing and the MCP result
// text gains no blank line (AC-TCI-006 (b), MU-21).
func renderIssuanceText(p factory.IssuancePresentation) string {
	if p.Empty() {
		return ""
	}
	var b strings.Builder
	fmt.Fprintln(&b, "issuance: advisory presentation (read-only, blocks nothing)")
	for _, n := range p.Neighbors {
		label, measure := "similar", "measure=token-set-jaccard"
		if n.Exact {
			label, measure = "exact", "measure=normalized-equal"
		}
		preview := []rune(n.Text)
		if len(preview) > 60 {
			preview = preview[:60]
		}
		fmt.Fprintf(&b, "  %-9s %-6s %-9s %.2f  %s  %q", label, n.ID, n.State, n.Score, measure, string(preview))
		if n.Reason != "" {
			fmt.Fprintf(&b, " (reason: %s)", n.Reason)
		}
		fmt.Fprintln(&b)
	}
	for _, c := range p.Components {
		fmt.Fprintf(&b, "  %-9s %-6s %-9s %s  %s\n", "component", c.ID, c.State, c.Key, "measure=component")
	}
	for _, it := range p.Overlap.Items {
		fmt.Fprintf(&b, "  %-9s %-6s %-9s %s  %s\n", "overlap", it.CardID, it.Lane, it.Path, "measure=file-overlap")
	}
	if p.Overlap.Unmeasured != "" {
		fmt.Fprintf(&b, "  %-9s unmeasured (%s)\n", "overlap", p.Overlap.Unmeasured)
	}
	for _, s := range p.Specs {
		fmt.Fprintf(&b, "  %-9s %-24s %-9s %s  heuristic\n", "spec", s.ID, s.Status, "measure=spec-heuristic")
	}
	return strings.TrimRight(b.String(), "\n")
}

// todoIssuancePresentation builds the presentation for one candidate text
// from a fresh LoadPure snapshot: queue snapshot, completed-SPEC directory
// read and lane probes all run here, before the caller acquires the queue
// lock for the write.
func todoIssuancePresentation(root, text string) factory.IssuancePresentation {
	return todoIssuancePresentationFloor(root, text, -1)
}

// todoIssuancePresentationFloor is todoIssuancePresentation with an explicit
// display floor; floor < 0 lifts the floor (the --dry-run view shows the
// top-3 neighbors regardless of score, design §3.2).
func todoIssuancePresentationFloor(root, text string, floor float64) factory.IssuancePresentation {
	store := todoStoreAt(root)
	rec, err := store.LoadPure()
	if err != nil {
		rec = &factory.BacklogRecord{}
	}
	specs := todoCompletedSpecsReader(root, time.Now().Add(factory.IssuanceProbeTimeBound))
	if floor < 0 {
		return factory.BuildIssuancePresentation(text, rec, specs, todoLaneFilesProbe)
	}
	return factory.BuildIssuancePresentationFloor(text, rec, specs, todoLaneFilesProbe, floor)
}

var issuanceFrontmatterField = regexp.MustCompile(`^([a-z_]+):\s*(.*)$`)

// loadCompletedSpecs reads .moai/specs/*/spec.md frontmatter under root.
// The read is budget-guarded: past the deadline it returns what it has
// (REQ-TCI-003 — a slow directory read degrades the spec item, never the
// admission).
func loadCompletedSpecs(root string, deadline time.Time) []factory.IssuanceCompletedSpec {
	dir := filepath.Join(root, ".moai", "specs")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []factory.IssuanceCompletedSpec
	for _, e := range entries {
		if time.Now().After(deadline) {
			break
		}
		if !e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name(), "spec.md"))
		if err != nil {
			continue
		}
		spec := parseIssuanceFrontmatter(e.Name(), string(data))
		if spec != nil {
			out = append(out, *spec)
		}
	}
	return out
}

// parseIssuanceFrontmatter extracts the five fields the coverage heuristic
// reads; a spec.md without an id line yields nil.
func parseIssuanceFrontmatter(dirName, data string) *factory.IssuanceCompletedSpec {
	spec := &factory.IssuanceCompletedSpec{ID: dirName}
	found := 0
	for _, line := range strings.Split(data, "\n") {
		if line == "---" && found > 0 {
			break
		}
		m := issuanceFrontmatterField.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		switch m[1] {
		case "id":
			spec.ID = strings.Trim(m[2], `"`)
			found++
		case "status":
			spec.Status = strings.Trim(m[2], `"`)
			found++
		case "module":
			spec.Module = strings.Trim(m[2], `"`)
			found++
		case "title":
			spec.Title = strings.Trim(m[2], `"`)
			found++
		case "tags":
			spec.Tags = strings.Trim(m[2], `"`)
			found++
		}
	}
	if spec.ID == "" || spec.Status == "" {
		return nil
	}
	return spec
}

// productionLaneFilesProbe resolves one in-flight card's lane-branch changed
// files: the card worktree convention names the worktree directory after the
// card id, so one porcelain listing locates its branch and the diff runs
// against that branch's merge-base with develop. Any failure is "no input",
// which the presentation reports as unmeasured rather than none.
func productionLaneFilesProbe(cardID, lane string) ([]string, bool) {
	root := resolveTodoQueueRoot()
	_, branch, ok := worktreeBranchForCard(root, cardID)
	if !ok {
		return nil, false
	}
	base := issuanceGitOneLine(root, "merge-base", "develop", branch)
	if base == "" {
		return nil, false
	}
	out := issuanceGitOut(root, "diff", "--name-only", base+"..."+branch)
	files := strings.Fields(out)
	if len(files) == 0 {
		return nil, false
	}
	return files, true
}

// worktreeBranchForCard finds the worktree whose directory is named after
// the card id (the card-worktree convention) and returns its branch.
func worktreeBranchForCard(root, cardID string) (path, branch string, ok bool) {
	list := issuanceGitOut(root, "worktree", "list", "--porcelain")
	suffix := "/worktrees/" + cardID
	var cur string
	for _, line := range strings.Split(list, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			cur = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "branch refs/heads/"):
			if strings.HasSuffix(filepath.ToSlash(cur), suffix) {
				return cur, strings.TrimPrefix(line, "branch refs/heads/"), true
			}
		}
	}
	return "", "", false
}

func issuanceGitOut(root string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func issuanceGitOneLine(root string, args ...string) string {
	return strings.TrimSpace(issuanceGitOut(root, args...))
}

// todoEngagePresentation prints the presentation for an engaged card's text
// to the command's error stream. Read-only end to end: no finding, no
// refusal, no queue write (REQ-TCI-005).
func todoEngagePresentation(cmd *cobra.Command, store *factory.BacklogStore, cardID string) {
	rec, err := store.LoadPure()
	if err != nil {
		return
	}
	var text string
	for _, it := range rec.Items {
		if it.ID == cardID {
			text = it.Text
			break
		}
	}
	if text == "" {
		return
	}
	p := todoIssuancePresentation(resolveTodoQueueRoot(), text)
	if presText := renderIssuanceText(p); presText != "" {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), presText)
	}
}
