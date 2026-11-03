package cli

// factory_release_route.go — card t2013, AFF S2 (SPEC-FACTORY-COMPLETION-
// RECOVERY-001 REQ-FCR-026, spec.md section D.5): the release route of
// `moai factory complete`.
//
// Under agentic-trunk, with the fixed release branch (release_branch_prefix +
// "main-batch") checked out in a worktree other than the primary checkout, a
// completed card lands through the local integration window into THAT branch
// — never main, never the pull-request edge — after the fast required check
// (`moai verify run`) has passed against the card's own tree. This file holds
// the route predicate, the idempotence probe, the fast check, and the
// card-level hold record; the window, the gates and the merge are the
// existing complete body's, called once, not re-implemented.

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/closure/gitio"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// factoryFastCheckID labels the fast required check's records so they can be
// filtered from other `moai verify run` records.
const factoryFastCheckID = "factory-complete-fast"

// factoryReleaseRoute reports the release route's branch and worktree when every
// condition of REQ-FCR-026 holds: an agentic-trunk profile, a non-empty
// release_branch_prefix, and the fixed release branch checked out in a worktree
// that is not the primary checkout. Any condition missing — including an
// unprovisioned branch (decision-index Q9) — answers false and leaves the
// existing routes in force.
func factoryReleaseRoute(root string) (branch, tree string, ok bool) {
	branch = config.LoadAgenticTrunkReleaseBranch(root)
	if branch == "" {
		return "", "", false
	}
	tree = factoryWorktreeForBranchIn(root, branch)
	if tree == "" {
		return "", "", false
	}
	primary, _, err := identifyPrimaryCheckout(root)
	if err != nil || factorySameTree(tree, primary) {
		return "", "", false
	}
	return branch, tree, true
}

// factoryReleaseAlreadyMerged returns the card's recorded merge SHA when the
// card is merged-local and that SHA is a merge commit whose second parent is
// the card's CURRENT tip and which the release branch still carries; "" in
// every other case. A card whose HEAD moved after the merge is not "already
// merged" and takes the ordinary gates.
func factoryReleaseAlreadyMerged(card homestate.Card, branch, tree string) string {
	if card.State != homestate.CardMergedLocal || card.MergeSHA == "" {
		return ""
	}
	cardTip := factoryBranchTip(card.WorktreePath, factoryBranchOfWorktree(card.WorktreePath))
	parents := factoryCommitParents(tree, card.MergeSHA)
	if cardTip == "" || len(parents) != 3 || parents[2] != cardTip {
		return ""
	}
	if ok, err := gitio.IsAncestor(tree, card.MergeSHA, "refs/heads/"+branch); err != nil || !ok {
		return ""
	}
	return card.MergeSHA
}

// factoryReleaseStop records the card-level hold of a release-route stop. It is
// inert (enabled false) on every other route, which keeps their stops exactly
// as they were.
type factoryReleaseStop struct {
	enabled                       bool
	db                            *homestate.FactoryDB
	runID                         string
	card                          homestate.Card
	lane                          string
	branch, integTree, cardBranch string
}

// hold writes the hold (reason class, detail, card HEAD, check key, the paths
// both sides changed) and returns ret unchanged, so a stop keeps its own error
// and exit code. A failed hold write is reported, never swallowed, and never
// replaces ret.
func (s factoryReleaseStop) hold(ctx context.Context, out io.Writer, reason string, detail error, checkKey string, ret error) error {
	if !s.enabled {
		return ret
	}
	cardTip := factoryBranchTip(s.card.WorktreePath, s.cardBranch)
	h := homestate.CardHold{
		CardID: s.card.CardID, Reason: reason, CardHead: cardTip, CheckKey: checkKey,
		Branch: s.branch, Actor: s.lane,
	}
	if detail != nil {
		h.Detail = factoryBoundedTail(detail.Error(), 2000)
	}
	if reason != "check-failed" && reason != "dirty-card-tree" && reason != "card-tree-unreadable" {
		h.ConflictPaths = factoryDivergedPaths(s.integTree, factoryBranchTip(s.integTree, s.branch), cardTip)
	}
	if err := s.db.RecordCardHold(ctx, s.runID, h, factoryCardNow()); err != nil {
		_, _ = fmt.Fprintf(out, "  writing the card hold failed (%v) — the stop reason is the refusal above\n", err)
	}
	return ret
}

// factoryDivergedPaths lists the paths both the integration tip and the card
// tip changed since their merge base — the files a merge could conflict on.
func factoryDivergedPaths(dir, tip, cardTip string) []string {
	if tip == "" || cardTip == "" {
		return nil
	}
	base, err := gitio.MergeBase(dir, tip, cardTip)
	if err != nil {
		return nil
	}
	changed := func(to string) map[string]bool {
		out, err := cliGitIn(dir, "diff", "--name-only", base, to)
		set := map[string]bool{}
		if err != nil {
			return set
		}
		for _, p := range strings.Fields(out) {
			set[p] = true
		}
		return set
	}
	mine, theirs := changed(cardTip), changed(tip)
	var both []string
	for p := range mine {
		if theirs[p] {
			both = append(both, p)
		}
	}
	sort.Strings(both)
	return both
}

// factoryBoundedTail keeps the last n bytes of s (a failing run's tail is the
// part that names the failure), cut at a line start when one is in reach.
func factoryBoundedTail(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	s = s[len(s)-n:]
	if i := strings.IndexByte(s, '\n'); i >= 0 && i < len(s)-1 {
		s = s[i+1:]
	}
	return s
}

// factoryDirtyPathLimit bounds how many offending paths a dirty-card-tree hold
// names.
const factoryDirtyPathLimit = 10

// factoryCardTreeDirt reports whether the card worktree differs from its
// committed HEAD: any modified, staged or untracked path (ignored files are not
// changes). The fast check runs in the working copy while the merge step merges
// the committed HEAD, so the check speaks for the merge only on a tree that is
// exactly its HEAD (REQ-FCR-026 (b); sync-audit F1). It returns "" when the tree
// is clean; otherwise the hold reason — dirty-card-tree, or card-tree-unreadable
// when the status query itself fails (fail closed) — with the hold detail.
func factoryCardTreeDirt(cardTree string) (reason string, detail error) {
	out, err := cliGitIn(cardTree, "status", "--porcelain", "-z", "--untracked-files=all")
	if err != nil {
		return "card-tree-unreadable", fmt.Errorf("git status in the card worktree %s failed (%v); commit or remove the changes in the card worktree, then run moai factory complete again", cardTree, err)
	}
	var paths []string
	entries := strings.Split(out, "\x00")
	for i := 0; i < len(entries); i++ {
		e := entries[i]
		if len(e) < 4 {
			continue
		}
		paths = append(paths, e[3:])
		if e[0] == 'R' || e[0] == 'C' || e[1] == 'R' || e[1] == 'C' {
			i++ // a rename or copy entry is followed by its source path
		}
	}
	if len(paths) == 0 {
		return "", nil
	}
	shown, more := paths, ""
	if len(shown) > factoryDirtyPathLimit {
		more = fmt.Sprintf(" and %d more", len(shown)-factoryDirtyPathLimit)
		shown = shown[:factoryDirtyPathLimit]
	}
	return "dirty-card-tree", fmt.Errorf("the card worktree %s has %d uncommitted path(s): %s%s; the merge lands the committed HEAD, so the check cannot speak for it; commit or remove the changes in the card worktree, then run moai factory complete again", cardTree, len(paths), strings.Join(shown, ", "), more)
}

// factoryFastCheck runs the fast required check against cardTree: build the
// module, lint and test the packages the card changed since the release tip.
// Each step is one `moai verify run` in the card tree, so a passing step for
// the same tree key is reused within its TTL (the reuse is printed with its key
// and recorded_at — a prior observation, not one made in this run). It returns
// the card tree's verify key and, on failure, the failing step and the tail of
// its output. A card tree that is not a Go module has no check defined here and
// fails closed.
func factoryFastCheck(ctx context.Context, out io.Writer, cardTree, releaseTip string) (key, tail string, err error) {
	if k, keyErr := verify.Key(ctx, cardTree); keyErr == nil {
		key = k
	}
	steps, err := factoryFastCheckSteps(cardTree, releaseTip)
	if err != nil {
		return key, err.Error(), err
	}
	for _, argv := range steps {
		var buf bytes.Buffer
		c := &cobra.Command{}
		c.SetContext(ctx)
		c.SetOut(&buf)
		c.SetErr(&buf)
		runErr := verifyRun(c, verifyRunParams{
			root: cardTree, argv: argv, checkID: factoryFastCheckID, ttl: 0,
			envNames:           []string{"GOFLAGS", "GOOS", "GOARCH", "CGO_ENABLED"},
			toolVersionCmd:     []string{"go", "version"},
			toolVersionTimeout: verifyRunToolVersionTimeout,
			timeout:            verifyRunDefaultTimeout,
		})
		if runErr != nil {
			cmdline := strings.Join(argv, " ")
			return key, cmdline + ": " + factoryBoundedTail(buf.String(), 1500), fmt.Errorf("%s: %w", cmdline, runErr)
		}
		for _, line := range strings.Split(buf.String(), "\n") {
			if strings.Contains(line, "verify run: reuse") {
				_, _ = fmt.Fprintf(out, "  fast check %s: %s (output not re-observed)\n", strings.Join(argv, " "), strings.TrimSpace(line))
			}
		}
	}
	return key, "", nil
}

// factoryCheckLeaseRenewEvery overrides the period at which the card lease is
// renewed while the fast check runs; zero means the configured
// workflow.factory_lease.renew_interval. Tests shorten it.
var factoryCheckLeaseRenewEvery time.Duration

// factoryFastCheckKeepingLease runs factoryFastCheck while renewing the lane's
// card lease every `every` through factoryRenewCard, the one renewal
// implementation: a lapsed lease cannot be renewed afterwards, and the check may
// outlast it. One more renewal follows the check so the merge step starts with a
// full lease. leaseErr is the first renewal failure; it is only meaningful when
// the check itself succeeded.
func factoryFastCheckKeepingLease(ctx context.Context, out io.Writer, root, runID, cardID, lane string, every time.Duration, cardTree, releaseTip string) (key, tail string, checkErr, leaseErr error) {
	renew := func() error {
		_, err := factoryRenewCard(ctx, root, cardID, runID, lane)
		return err
	}
	stop, finished := make(chan struct{}), make(chan struct{})
	var renewErr error // written by the goroutine only, read after <-finished
	go func() {
		defer close(finished)
		ticker := time.NewTicker(every)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				if renewErr = renew(); renewErr != nil {
					return
				}
			}
		}
	}()
	key, tail, checkErr = factoryFastCheck(ctx, out, cardTree, releaseTip)
	close(stop)
	<-finished
	if renewErr == nil && checkErr == nil {
		renewErr = renew()
	}
	return key, tail, checkErr, renewErr
}

// factoryFastCheckSteps names the check as argv steps (no shell, so no sh on
// Windows): `go build ./...`, the lint of the changed packages, and their
// tests. Changed packages come from the card's change against the release tip.
func factoryFastCheckSteps(cardTree, releaseTip string) ([][]string, error) {
	if _, err := os.Stat(filepath.Join(cardTree, "go.mod")); err != nil {
		return nil, fmt.Errorf("no go.mod in the card tree %s: the fast required check is defined for Go modules only", cardTree)
	}
	pkgs := factoryFastCheckPackages(cardTree, releaseTip)
	steps := [][]string{{"go", "build", "./..."}}
	if len(pkgs) == 0 {
		return steps, nil
	}
	lint := append([]string{"go", "vet"}, pkgs...)
	if _, err := exec.LookPath("golangci-lint"); err == nil && factoryHasGolangciConfig(cardTree) {
		lint = append([]string{"golangci-lint", "run", "--allow-parallel-runners"}, pkgs...)
	}
	return append(steps, lint, append([]string{"go", "test"}, pkgs...)), nil
}

func factoryHasGolangciConfig(dir string) bool {
	for _, name := range []string{".golangci.yml", ".golangci.yaml", ".golangci.toml", ".golangci.json"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	return false
}

// factoryChangedGoPackages lists, as ./dir arguments (or "." for the module
// root), the Go package directories that own what the card changed since its
// merge base with releaseTip.
func factoryChangedGoPackages(cardTree, releaseTip string) []string {
	return factoryOwningGoPackages(cardTree, factoryChangedPaths(cardTree, releaseTip))
}

// factoryChangedPaths lists the paths the card changed since its merge base
// with releaseTip; nil when the tip is empty or the diff cannot be read.
func factoryChangedPaths(cardTree, releaseTip string) []string {
	if releaseTip == "" {
		return nil
	}
	// NUL-delimited: git quotes non-ASCII paths in line mode, and a path may
	// hold spaces, so neither a newline nor a whitespace split is safe.
	out, err := cliGitIn(cardTree, "diff", "--name-only", "-z", releaseTip+"...HEAD")
	if err != nil {
		return nil
	}
	var paths []string
	for _, p := range strings.Split(out, "\x00") {
		if p != "" {
			paths = append(paths, p)
		}
	}
	return paths
}

// factoryFastCheckPackages decides the packages to lint and test from the
// card's changed file list. A module input (go.mod, go.sum, go.work,
// go.work.sum) can break any package and owns none, so it widens the scope to
// the whole module; so does a changed .go file no package directory claims
// (for example a deleted file in a vanished directory).
func factoryFastCheckPackages(cardTree, releaseTip string) []string {
	changed := factoryChangedPaths(cardTree, releaseTip)
	pkgs := factoryOwningGoPackages(cardTree, changed)
	for _, p := range changed {
		switch path.Base(p) {
		case "go.mod", "go.sum", "go.work", "go.work.sum":
			return []string{"./..."}
		}
		if strings.HasSuffix(p, ".go") && len(pkgs) == 0 {
			return []string{"./..."}
		}
	}
	return pkgs
}

// factoryOwningGoPackages maps each changed path (relative to cardTree) to the
// nearest ancestor directory holding at least one .go file, walking up from the
// path's directory to the module root. A testdata fixture or a go:embed asset
// therefore selects the package that owns it; testdata directories are never
// packages themselves; a deleted path whose directory vanished selects its
// nearest existing ancestor. The module root is selected (as ".") only when it
// holds .go files. A pure function of the path list and the tree's directories.
func factoryOwningGoPackages(cardTree string, changed []string) []string {
	seen := map[string]bool{}
	var pkgs []string
	for _, p := range changed {
		if !filepath.IsLocal(p) {
			continue
		}
		for dir := filepath.Dir(p); ; {
			if !factoryInTestdata(dir) && factoryDirHasGoFiles(filepath.Join(cardTree, dir)) {
				arg := "."
				if dir != "." {
					arg = "./" + filepath.ToSlash(dir)
				}
				if !seen[arg] {
					seen[arg] = true
					pkgs = append(pkgs, arg)
				}
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	sort.Strings(pkgs)
	return pkgs
}

// factoryInTestdata reports whether the slash-relative dir sits in or under a
// testdata directory.
func factoryInTestdata(dir string) bool {
	return strings.Contains("/"+filepath.ToSlash(dir)+"/", "/testdata/")
}

// factoryDirHasGoFiles reports whether dir exists and directly holds a .go file.
func factoryDirHasGoFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
			return true
		}
	}
	return false
}
