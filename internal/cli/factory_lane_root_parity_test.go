package cli

// factory_lane_root_parity_test.go — sync-audit N2 (card t1663; codex P1; decision d-20261010T164151Z-c591).
//
// In a linked worktree, each lane verb must read the factory record and its lease keys from one root
// whether the verb or the MCP tool runs it. The factory root of a linked worktree is its primary checkout,
// where the record is kept (factoryRootOf). The MCP handlers for next, stage, and complete opened the
// factory at the caller's raw project_root instead, so the same card could be acted on under a different
// workflow.factory_lease.duration on each surface.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// laneParityFixture builds a primary checkout whose workflow.factory_lease.duration is duration, a linked
// worktree of it, and one card (t1) leased by lane-1 in run fcRun. The primary's keys are written after
// the fixture has opened the factory, so a 0s duration cannot fail the fixture itself.
func laneParityFixture(t *testing.T, duration string) (root, wt string) {
	t.Helper()
	root, _ = sdMoaiFixture(t)
	wt = addGitWorktree(t, root)
	sdRegisterLane(t, root, "lane-1")
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardLeased, Stage: homestate.CardRun,
		OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: fcNow.Add(5 * time.Minute).Format(time.RFC3339Nano),
		WorktreePath: wt})
	factoryRenewWriteLeaseDuration(t, root, duration)
	sdLaneEnv(t, "lane-1", "")
	t.Setenv("CLAUDE_PROJECT_DIR", wt)
	t.Setenv(config.EnvClaudeCodeSessionID, "lane-parity-session")
	return root, wt
}

// laneParityRun runs one lane verb on a fresh fixture through the surface named by surface ("verb" or
// "MCP tool") and returns the outcome as one comparable string. A refusal carries its text, with the tool
// prefix removed and the checkout paths replaced by tokens, then the card's state, version, and lease
// expiry. A success carries the card's state, version, and lease expiry.
func laneParityRun(t *testing.T, verb, duration, surface string) string {
	t.Helper()
	root, wt := laneParityFixture(t, duration)
	tool := "factory_" + verb
	var callErr error
	if surface == "verb" {
		args := map[string][]string{
			"next":     {"next"},
			"stage":    {"stage", "t1", "run", "--run", fcRun},
			"complete": {"complete", "t1", "--run", fcRun},
		}[verb]
		_, _, callErr = runFactory(t, args...)
	} else {
		args := map[string]any{projectRootArg: wt}
		switch verb {
		case "stage":
			args["card"], args["state"], args["run"] = "t1", "run", fcRun
		case "complete":
			args["card"], args["run"] = "t1", fcRun
		}
		text, err := sdCallTool(t, factoryRenewMCPHandler(t, tool), args)
		callErr = err
		// REQ-MN-001: a policy refusal is a structured result (refused:true), not an
		// error result — read it back as the refusal the verb prints.
		if callErr == nil {
			var payload struct {
				Refused bool   `json:"refused"`
				Reason  string `json:"reason"`
			}
			if json.Unmarshal([]byte(text), &payload) == nil && payload.Refused {
				callErr = errors.New(payload.Reason)
			}
		}
	}
	card := laneParityCard(t, root)
	if callErr != nil {
		msg := laneParityNormalize(strings.TrimPrefix(callErr.Error(), tool+": "),
			map[string]string{root: "<primary>", wt: "<worktree>"})
		return fmt.Sprintf("refused %q state=%s version=%d lease=%s", msg, card.State, card.Version, card.LeaseExpiresAt)
	}
	return fmt.Sprintf("ok state=%s version=%d lease=%s", card.State, card.Version, card.LeaseExpiresAt)
}

// laneParityCard reads card t1 from the factory record of root. It opens the database by path, so that a
// primary whose duration OpenFactory refuses (0s) can still be read after the verb ran against it.
func laneParityCard(t *testing.T, root string) homestate.Card {
	t.Helper()
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatalf("factory record path: %v", err)
	}
	db, err := homestate.OpenFactoryPath(path)
	if err != nil {
		t.Fatalf("open factory record: %v", err)
	}
	defer func() { _ = db.Close() }()
	card, err := db.LoadCard(context.Background(), fcRun, "t1")
	if err != nil {
		t.Fatalf("load t1: %v", err)
	}
	return card
}

// laneParityNormalize replaces each path with its token, so that two fixtures built in different temporary
// directories compare equal when their outcomes are equal. The symlink-resolved spelling is replaced first:
// on macOS it contains the raw spelling (/private/var contains /var), so the raw spelling must not go first.
func laneParityNormalize(msg string, tokens map[string]string) string {
	for path, token := range tokens {
		if real, err := filepath.EvalSymlinks(path); err == nil {
			msg = strings.ReplaceAll(msg, real, token)
		}
		msg = strings.ReplaceAll(msg, path, token)
	}
	return msg
}

// TestLaneVerbsMatchMCPLeaseDuration is the equality table for sync-audit N2. For each lane verb (next,
// stage, complete) and each primary duration (20m, 0s), the verb and the MCP tool run on identical
// fixtures from a linked worktree, and their outcomes and the card's resulting lease expiry must be equal.
// Each row is one subtest, and -v prints the table: one line per row.
func TestLaneVerbsMatchMCPLeaseDuration(t *testing.T) {
	for _, verb := range []string{"next", "stage", "complete"} {
		for _, duration := range []string{"20m", "0s"} {
			t.Run(verb+"/primary-"+duration, func(t *testing.T) {
				verbOut := laneParityRun(t, verb, duration, "verb")
				mcpOut := laneParityRun(t, verb, duration, "MCP tool")
				t.Logf("| %-8s | primary %-3s | verb: %s | MCP: %s | equal: %t |",
					verb, duration, verbOut, mcpOut, verbOut == mcpOut)
				if verbOut != mcpOut {
					t.Errorf("%s with primary duration %s: the verb and the MCP tool disagree", verb, duration)
				}
			})
		}
	}
}
