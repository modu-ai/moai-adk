// mcp_factory_card.go — the four factory card MCP tools
// (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-014/-024): factory_next,
// factory_stage, factory_complete, factory_decide. Each is a thin wrapper
// over the SAME function the cobra RunE body calls (design.md §3 — one
// implementation per verb): factoryNextLeaseOnce + factoryEnsureCardWorktree
// + factoryNextWriteOutput, factoryStageCard, factoryCompleteCard,
// factoryDecideCards. No core logic is forked here, and every refusal the
// CLI prints is printed by the same guard function, so the two surfaces
// cannot drift (AC-SD-014).
//
// REQ-SD-024: the three lane verbs REQUIRE the caller-supplied project_root
// (missing → rejected naming the argument; not a MoAI project root →
// rejected naming the path, exactly like the existing project_root tools),
// and factory_next's parent-checkout check evaluates that argument. The
// lane predicates read the server process's environment — the same
// environment the Codex MCP env_vars allowlist forwards — so a lane session
// and a Codex lane session behave identically on both surfaces. The
// merge-ready → merging edge is refused on this path by the same
// factoryRefuseCodexMergeEdge check the CLI verbs run: it rides inside
// factoryStageCard and factoryCompleteCard, so no separate wiring lives
// here (AC-SD-024's MCP half consumes it through the same call).
package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/modu-ai/moai-adk/internal/config"
)

// registerFactoryCardMCPTools registers the four factory card tools on the
// shared add gate (per-tool enablement map + catalog equality guard).
func registerFactoryCardMCPTools(add func(name string, tool mcp.Tool, handler server.ToolHandlerFunc)) {
	add("factory_next", mcp.NewTool(
		"factory_next",
		mcp.WithDescription("Lease the lane's next card through the factory record — the MCP form of `moai factory next` (no --wait). Lane session only; runs from the parent checkout named by project_root. Same implementation, same record changes, same refusals."),
		mcp.WithString("run", mcp.Description("Factory run id (default: the single active run).")),
		requiredProjectRootOption(),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleFactoryNext)

	add("factory_stage", mcp.NewTool(
		"factory_stage",
		mcp.WithDescription("Apply a card's next F1 stage transition with its evidence and renew the lease — the MCP form of `moai factory stage`. Lane session only; a Codex lane is refused the merge-ready → merging edge. Same implementation, same record changes, same refusals."),
		mcp.WithString("card", mcp.Required(), mcp.Description("The card id.")),
		mcp.WithString("state", mcp.Required(), mcp.Description("The F1 state to transition to (e.g. plan-audit).")),
		mcp.WithString("evidence", mcp.Description("Optional evidence: `<sha>` where the edge reads a commit only, `<sha>:<repo-relative-artifact>` where it also names the artifact.")),
		mcp.WithString("run", mcp.Description("Factory run id (default: the single active run).")),
		requiredProjectRootOption(),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleFactoryStage)

	add("factory_complete", mcp.NewTool(
		"factory_complete",
		mcp.WithDescription("Take a merge-ready card through merging to merged-local by the F1 merge gate — the MCP form of `moai factory complete` (no re-measure positional; complete records the merge evidence itself). Lane session only; a Codex lane is refused. The integration window stays held — release is the lane's next step."),
		mcp.WithString("card", mcp.Required(), mcp.Description("The card id.")),
		mcp.WithString("run", mcp.Description("Factory run id (default: the single active run).")),
		requiredProjectRootOption(),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleFactoryComplete)

	add("factory_decide", mcp.NewTool(
		"factory_decide",
		mcp.WithDescription("Record one operator decision — the MCP form of `moai factory decide` (one card per call). Refused for a lane session: decide records the operator's decisions. "+projectRootDesc),
		mcp.WithString("card", mcp.Required(), mcp.Description("The card id.")),
		mcp.WithString("gate", mcp.Description("Decision gate: kickoff or push. Empty with a lifecycle choice decides the card's pause state.")),
		mcp.WithString("choice", mcp.Description("approve|reject (kickoff), or resume|block|unblock|abandon.")),
		mcp.WithString("run", mcp.Description("Factory run id (default: the single active run).")),
		projectRootOption(),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleFactoryDecide)
}

// requiredProjectRootOption declares project_root on a tool where the
// argument is REQUIRED (REQ-SD-024): the lane verbs must be told which tree
// to act on — the server's own resolution cannot follow the caller's
// worktree, and guessing is the failure this argument exists to prevent.
func requiredProjectRootOption() mcp.ToolOption {
	return mcp.WithString(projectRootArg, mcp.Required(), mcp.Description(
		"Required project or worktree root to act on. Supply your own `git rev-parse --show-toplevel`. "+
			"Missing is rejected naming this argument; a path that is not a MoAI project root is rejected naming the path; "+
			"an accepted path is canonicalized — symlinks resolved — so the call acts on the real directory."))
}

// mcpRequiredProjectRoot resolves the REQUIRED project_root argument
// (REQ-SD-024): absent is a rejection naming the argument, present runs the
// same validation the existing project_root tools run (validateProjectRoot,
// which rejects a non-MoAI-project-root naming the path).
func mcpRequiredProjectRoot(req mcp.CallToolRequest, tool string) (string, error) {
	raw := strings.TrimSpace(req.GetString(projectRootArg, ""))
	if raw == "" {
		return "", fmt.Errorf("%s: project_root is required — pass the tree to act on (your git rev-parse --show-toplevel)", tool)
	}
	return validateProjectRoot(raw)
}

// handleFactoryNext is the MCP form of `moai factory next` (no --wait): one
// selection+lease for the lane, then the REQ-SD-011 worktree step.
func handleFactoryNext(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root, err := mcpRequiredProjectRoot(req, "factory_next")
	if err != nil {
		return toolErr("factory_next", err), nil
	}
	if !factoryLaneAdmission() {
		return toolErr("factory_next", factoryNotALaneError("next")), nil
	}
	lane := strings.TrimSpace(os.Getenv(config.EnvMoaiKanbanLabel))
	if lane == "" {
		return toolErr("factory_next", fmt.Errorf("factory next: %s is empty — a lane session carries its lane label there", config.EnvMoaiKanbanLabel)), nil
	}
	if err := factoryAssertParentCheckout(root); err != nil {
		return toolErr("factory_next", err), nil
	}
	runID, err := resolveFactoryCardRun(ctx, root, req.GetString("run", ""))
	if err != nil {
		return toolErr("factory_next", fmt.Errorf("factory next: %w", err)), nil
	}
	card, leased, err := factoryNextLeaseOnce(ctx, root, runID, lane)
	if err != nil {
		return toolErr("factory_next", fmt.Errorf("factory next: %w", err)), nil
	}
	if !leased {
		return mcp.NewToolResultText("no card is available"), nil
	}
	wt, _, err := factoryEnsureCardWorktree(ctx, root, runID, card, lane, io.Discard)
	if err != nil {
		return toolErr("factory_next", fmt.Errorf("factory next: %w", err)), nil
	}
	card.WorktreePath = wt
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	if err := factoryNextWriteOutput(out, errOut, root, card); err != nil {
		return toolErr("factory_next", err), nil
	}
	return mcp.NewToolResultText(strings.TrimRight(out.String(), "\n")), nil
}

// handleFactoryStage is the MCP form of `moai factory stage`.
func handleFactoryStage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root, err := mcpRequiredProjectRoot(req, "factory_stage")
	if err != nil {
		return toolErr("factory_stage", err), nil
	}
	if !factoryLaneAdmission() {
		return toolErr("factory_stage", factoryNotALaneError("stage")), nil
	}
	cardID := strings.TrimSpace(req.GetString("card", ""))
	state := strings.TrimSpace(req.GetString("state", ""))
	if cardID == "" || state == "" {
		return toolErr("factory_stage", errors.New("factory stage: card and state are required")), nil
	}
	card, err := factoryStageCard(ctx, root, cardID, state, req.GetString("evidence", ""), req.GetString("run", ""), strings.TrimSpace(os.Getenv(config.EnvMoaiKanbanLabel)))
	if err != nil {
		return toolErr("factory_stage", err), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("%s %s v%d lease renewed", card.CardID, card.State, card.Version)), nil
}

// handleFactoryComplete is the MCP form of `moai factory complete`.
func handleFactoryComplete(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root, err := mcpRequiredProjectRoot(req, "factory_complete")
	if err != nil {
		return toolErr("factory_complete", err), nil
	}
	if !factoryLaneAdmission() {
		return toolErr("factory_complete", factoryNotALaneError("complete")), nil
	}
	lane := strings.TrimSpace(os.Getenv(config.EnvMoaiKanbanLabel))
	if lane == "" {
		return toolErr("factory_complete", fmt.Errorf("factory complete: %s is empty — a lane session carries its lane label there", config.EnvMoaiKanbanLabel)), nil
	}
	cardID := strings.TrimSpace(req.GetString("card", ""))
	if cardID == "" {
		return toolErr("factory_complete", errors.New("factory complete: card is required")), nil
	}
	out := &bytes.Buffer{}
	if err := factoryCompleteCard(ctx, out, root, root, cardID, "", req.GetString("run", ""), lane); err != nil {
		return toolErr("factory_complete", err), nil
	}
	return mcp.NewToolResultText(strings.TrimRight(out.String(), "\n")), nil
}

// handleFactoryDecide is the MCP form of `moai factory decide` (one card per
// call). The REQ-SD-016 lane refusal — a lane by marker, label, or Codex
// backend — is the same refusal the CLI guard returns.
func handleFactoryDecide(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root, err := resolveToolProjectRoot(req)
	if err != nil {
		return toolErr("factory_decide", err), nil
	}
	if factoryLaneRefusal() {
		return toolErr("factory_decide", factoryDecideLaneRefusal()), nil
	}
	cardID := strings.TrimSpace(req.GetString("card", ""))
	if cardID == "" {
		return toolErr("factory_decide", errors.New("factory decide: card is required")), nil
	}
	gate := req.GetString("gate", "")
	choice := req.GetString("choice", "")
	switch {
	case gate == "kickoff" && (choice == "approve" || choice == "reject"):
	case gate == "push" && choice == "":
	case gate == "" && (choice == "resume" || choice == "block" || choice == "unblock" || choice == "abandon"):
	default:
		return toolErr("factory_decide", errors.New("factory decide: want --gate kickoff --choice approve|reject, --gate push, or --choice resume|block|unblock|abandon")), nil
	}
	out := &bytes.Buffer{}
	if err := factoryDecideCards(ctx, root, out, []string{cardID}, gate, choice, req.GetString("run", "")); err != nil {
		return toolErr("factory_decide", err), nil
	}
	return mcp.NewToolResultText(strings.TrimRight(out.String(), "\n")), nil
}
