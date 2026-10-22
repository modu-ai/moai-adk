// mcp_factory_card.go — the five factory card MCP tools
// (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-014/-024): factory_next,
// factory_stage, factory_complete, factory_decide, and factory_renew
// (SPEC-FACTORY-LEASE-RENEW-001 REQ-FLR-002). Each is a thin wrapper
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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// registerFactoryCardMCPTools registers the four factory card tools on the
// shared add gate (per-tool enablement map + catalog equality guard).
func registerFactoryCardMCPTools(add func(name string, tool mcp.Tool, handler server.ToolHandlerFunc)) {
	add("factory_next", mcp.NewTool(
		"factory_next",
		mcp.WithDescription("Lease the lane's next card through the factory record — the MCP form of `moai factory next` (no --wait). Lane session only; runs from the parent checkout named by project_root. Same implementation, same record changes, same refusals. With card it leases exactly that card or returns the refusal line `factory next: refused <token>: <detail>` as an error result."),
		mcp.WithString("run", mcp.Description("Factory run id (default: the single active run). A run that is not the active run refuses stale-run.")),
		mcp.WithString("card", mcp.Description("Optional: the card id to lease (the session's own judged pick). Leased through the same version-checked edges, or refused with a reason token; the quota hold and the Codex skip apply as to any new card. Absent: the unnominated priority-order lease.")),
		optionalFactoryActorOption(),
		requiredProjectRootOption(),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleFactoryNext)

	add("factory_stage", mcp.NewTool(
		"factory_stage",
		mcp.WithDescription("Apply a card's next F1 stage transition with its evidence and renew the lease — the MCP form of `moai factory stage`. Lane session only; a Codex lane is refused the merge-ready → merging edge. Same implementation, same record changes, same refusals."),
		mcp.WithString("card", mcp.Required(), mcp.Description("The card id.")),
		mcp.WithString("state", mcp.Required(), mcp.Description("The F1 state to transition to (e.g. plan-audit).")),
		mcp.WithString("evidence", mcp.Description("Optional evidence: `<sha>` where the edge reads a commit only, `<sha>:<repo-relative-artifact>` where it also names the artifact.")),
		mcp.WithString("run", mcp.Description("Factory run id (default: the single active run). A run that is not the active run refuses stale-run.")),
		optionalFactoryActorOption(),
		requiredProjectRootOption(),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleFactoryStage)

	add("factory_complete", mcp.NewTool(
		"factory_complete",
		mcp.WithDescription("Take a merge-ready card through merging to merged-local by the F1 merge gate — the MCP form of `moai factory complete` (no re-measure positional; complete records the merge evidence itself). Lane session only; a Codex lane is refused. The integration window stays held — release is the lane's next step. Under github-flow the same call delivers by pull request instead: it pushes the card branch, opens a PR against the integration target and records pr-open (no local merge, no window); a later call records merged-pr once the PR is observed merged."),
		mcp.WithString("card", mcp.Required(), mcp.Description("The card id.")),
		mcp.WithString("run", mcp.Description("Factory run id (default: the single active run). A run that is not the active run refuses stale-run.")),
		optionalFactoryActorOption(),
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

	add("factory_renew", mcp.NewTool(
		"factory_renew",
		mcp.WithDescription("Extend the lane's card lease by the configured duration — the MCP form of `moai factory renew --json` (SPEC-FACTORY-LEASE-RENEW-001 REQ-FLR-002). Lane session only. Returns the same JSON outcome as the verb: result (renewed, refused-expired, refused-not-holder, or error) and expires_at. A refused renewal writes nothing and is returned as that result, not as a tool error."),
		mcp.WithString("card", mcp.Required(), mcp.Description("The card id.")),
		mcp.WithString("run", mcp.Description("Factory run id (default: the single active run). A run that is not the active run refuses stale-run.")),
		optionalFactoryActorOption(),
		requiredProjectRootOption(),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleFactoryRenew)
}

// optionalFactoryActorOption declares the actor argument (SPEC-MCP-NEUTRALITY-001
// REQ-MN-003): the lane label the call acts as, with the spawn environment's
// factory-worker key (internal/config/envkeys.go EnvMoaiFactoryWorker) as the
// default only. An explicit actor lets a caller name its lane when the
// environment does not carry it.
func optionalFactoryActorOption() mcp.ToolOption {
	return mcp.WithString("actor", mcp.Description(
		"Optional: the lane label this call acts as. Default: the spawn environment's "+config.EnvMoaiFactoryWorker+". An empty value with no environment label refuses as unknown attribution."))
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
//
// @MX:ANCHOR: [AUTO] REQUIRED project_root resolver for the factory MCP tools.
// @MX:REASON: 3 call sites (factory_next/stage/complete handlers); a new tool handler skipping it would act on the server's working directory instead of the caller's tree (REQ-SD-024).
// @MX:SPEC: SPEC-FACTORY-SELF-DISPATCH-001
func mcpRequiredProjectRoot(req mcp.CallToolRequest, tool string) (string, error) {
	raw := strings.TrimSpace(req.GetString(projectRootArg, ""))
	if raw == "" {
		// The handler's toolErr adds the "tool: " prefix — the error stays
		// unprefixed so the wire text reads "factory_stage: project_root is
		// required — ..." (one prefix, not two; the t1632 absorb merge).
		return "", errors.New("project_root is required — pass the tree to act on (your git rev-parse --show-toplevel)")
	}
	return validateProjectRoot(raw)
}

// handleFactoryNext is the MCP form of `moai factory next` (no --wait): one
// selection+lease for the lane, then the REQ-SD-011 worktree step. The policy
// gates refuse as structured results (SPEC-MCP-NEUTRALITY-001 REQ-MN-001):
// serial-slot / owned / raced carry their closed-set codes, the identity and
// parent-checkout classes carry refused:true, and none rides the error
// channel. The actor argument (REQ-MN-003) names whose label acts; the spawn
// environment is the default only. An explicit run that is not the active
// run refuses stale-run before any row is read.
func handleFactoryNext(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	launch, err := mcpRequiredProjectRoot(req, "factory_next")
	if err != nil {
		return toolErr("factory_next", err), nil
	}
	if !factoryLaneAdmission() {
		return toolRefusal(mcpIdentityRefusal(factoryNotALaneError("next").Error())), nil
	}
	// The same lane-label resolution the verb runs (factoryLaneLabel, the t1914
	// parity contract, card t1632 main-absorb r9): the actor argument wins, the
	// environment is the default, and BOTH the empty case and a legacy spelling
	// refuse on the error channel with the verb's exact wording.
	lane, err := factoryLaneLabel("next", mcpFactoryActor(req))
	if err != nil {
		return toolErr("factory_next", err), nil
	}
	// The parent-checkout assertion names the caller's own tree, so it stays on the raw
	// project_root: a linked worktree is refused here, as the verb refuses it. A check run
	// on the factory root instead would pass every worktree (its root IS the parent) and
	// silently disable the refusal. The record and the lease keys come from the factory root
	// of that tree, which after the assertion is the parent checkout itself (sync-audit N2).
	if err := factoryAssertParentCheckout(launch); err != nil {
		return mcpParentCheckoutRefusal(err), nil
	}
	root := factoryRootOf(launch)
	explicitRun := req.GetString("run", "")
	if res := mcpStaleRunCheck(ctx, root, explicitRun, "factory_next"); res != nil {
		return res, nil
	}
	runID, err := resolveFactoryCardRun(ctx, root, explicitRun)
	if err != nil {
		return toolErr("factory_next", fmt.Errorf("factory next: %w", err)), nil
	}
	// A card argument present with a blank value is an error, never a silent
	// fall-back to the priority-order choice the session did not make.
	nominee := strings.TrimSpace(req.GetString("card", ""))
	if _, present := req.GetArguments()["card"]; present && nominee == "" {
		return toolErr("factory_next", errors.New("factory next: card needs a card id")), nil
	}
	// The same quota gate as the CLI verb, through the same shared evaluation.
	held, holdLine := (&factoryQuotaLatch{}).evaluate(root)
	// The open-PR query runs before the lease, as in the CLI verb, so the output
	// below issues no query after the queue lock is released (AC-FAL-015 (vii)).
	openPRs := factoryNextPRSnapshot(ctx, root, runID, lane, nominee, held)
	var card homestate.Card
	leased := false
	// The gate's refusals of the last pass (SPEC-CARD-ELIGIBLE-PULL-001
	// REQ-CEP-005): rendered through the same shared text the CLI verb prints.
	var gateRefusals []factoryEligibleRefusal
	if nominee != "" {
		// The nominated lease: the same implementation the CLI verb calls, the
		// same refusal line (design.md §3 — one implementation per verb); the
		// closed-set refusal tokens surface as structured results.
		quotaHold := ""
		if held {
			quotaHold = holdLine
		}
		card, err = factoryNextNominate(ctx, root, runID, lane, nominee, quotaHold)
		var refusal *factoryNominateRefusal
		switch {
		case err == nil:
			leased = true
		case errors.As(err, &refusal):
			if payload, ok := mcpNominateRefusalPayload(refusal); ok {
				return toolRefusal(payload), nil
			}
			return toolErr("factory_next", refusal), nil
		default:
			return toolErr("factory_next", fmt.Errorf("factory next: %w", err)), nil
		}
	} else {
		card, leased, gateRefusals, err = factoryNextLeaseOnceEligibility(ctx, root, runID, lane, held)
		if err != nil {
			return toolErr("factory_next", fmt.Errorf("factory next: %w", err)), nil
		}
	}
	if !leased {
		if held {
			// A hold is the whole, non-error result text — never the empty-queue text.
			return mcp.NewToolResultText(holdLine), nil
		}
		return mcp.NewToolResultText(factoryNextNoCardText(gateRefusals)), nil
	}
	wt, _, err := factoryEnsureCardWorktree(ctx, root, runID, card, lane, io.Discard)
	if err != nil {
		// The lease already committed (card t1696, the lane-20 shape): the
		// refusal must say so, or the lane retries a card it already holds.
		return toolErr("factory_next", fmt.Errorf("factory next: %s was leased (the lease stands in the record; do not re-run next for it — settle the worktree step instead): %w", card.CardID, err)), nil
	}
	card.WorktreePath = wt
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	if err := factoryNextWriteOutput(out, errOut, root, card, openPRs); err != nil {
		return toolErr("factory_next", err), nil
	}
	return mcp.NewToolResultText(strings.TrimRight(out.String(), "\n")), nil
}

// handleFactoryStage is the MCP form of `moai factory stage`. The actor
// argument (REQ-MN-003) names whose label applies the edge; a stale run
// refuses before any row is read, and the F1 holder refusal surfaces as the
// owned policy code (SPEC-MCP-NEUTRALITY-001 REQ-MN-001/003).
func handleFactoryStage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	launch, err := mcpRequiredProjectRoot(req, "factory_stage")
	if err != nil {
		return toolErr("factory_stage", err), nil
	}
	if !factoryLaneAdmission() {
		return toolRefusal(mcpIdentityRefusal(factoryNotALaneError("stage").Error())), nil
	}
	cardID := strings.TrimSpace(req.GetString("card", ""))
	state := strings.TrimSpace(req.GetString("state", ""))
	if cardID == "" || state == "" {
		return toolErr("factory_stage", errors.New("factory stage: card and state are required")), nil
	}
	// The factory root of the caller's tree, as the verb resolves its own (factoryCardRoot), so a
	// linked worktree's lease keys come from the primary checkout on both surfaces (sync-audit N2).
	root := factoryRootOf(launch)
	explicitRun := req.GetString("run", "")
	if res := mcpStaleRunCheck(ctx, root, explicitRun, "factory_stage"); res != nil {
		return res, nil
	}
	// The same lane-label resolution the verb runs (factoryLaneLabel, the t1914
	// parity contract, card t1632 main-absorb r9) — the error channel carries
	// the verb's exact wording for both the empty case and a legacy spelling.
	lane, err := factoryLaneLabel("stage", mcpFactoryActor(req))
	if err != nil {
		return toolErr("factory_stage", err), nil
	}
	card, err := factoryStageCard(ctx, root, cardID, state, req.GetString("evidence", ""), explicitRun, lane)
	if err != nil {
		if res, ok := mcpLeaseHolderRefusal(err); ok {
			return res, nil
		}
		return toolErr("factory_stage", err), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("%s %s v%d lease renewed", card.CardID, card.State, card.Version)), nil
}

// handleFactoryComplete is the MCP form of `moai factory complete`. The actor
// argument (REQ-MN-003) names the completing lane; a stale run refuses before
// any row is read (REQ-MN-003).
func handleFactoryComplete(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	launch, err := mcpRequiredProjectRoot(req, "factory_complete")
	if err != nil {
		return toolErr("factory_complete", err), nil
	}
	if !factoryLaneAdmission() {
		return toolRefusal(mcpIdentityRefusal(factoryNotALaneError("complete").Error())), nil
	}
	// The same lane-label resolution the verb runs (factoryLaneLabel, the t1914
	// parity contract, card t1632 main-absorb r9) — the error channel carries
	// the verb's exact wording for both the empty case and a legacy spelling.
	lane, err := factoryLaneLabel("complete", mcpFactoryActor(req))
	if err != nil {
		return toolErr("factory_complete", err), nil
	}
	cardID := strings.TrimSpace(req.GetString("card", ""))
	if cardID == "" {
		return toolErr("factory_complete", errors.New("factory complete: card is required")), nil
	}
	root := factoryRootOf(launch)
	explicitRun := req.GetString("run", "")
	if res := mcpStaleRunCheck(ctx, root, explicitRun, "factory_complete"); res != nil {
		return res, nil
	}
	out := &bytes.Buffer{}
	// The record and the git-flow config live at the factory root of the caller's tree, the parent checkout
	// (sync-audit N2, card t1680 (b)). The integration window and the re-measure record live at the
	// repository common root (factory.IntegrationCommonRoot, card t1680 (b) F1): both surfaces resolve the
	// lock root to the repository common root, so a CLAUDE_PROJECT_DIR naming a linked worktree no longer
	// splits them. The handler cannot call integrationLockRoot() itself — it reads the process environment and
	// working directory, which this long-lived server cannot use for a caller's tree.
	if err := factoryCompleteCard(ctx, out, root, factory.IntegrationCommonRoot(launch), cardID, "", explicitRun, lane); err != nil {
		if res, ok := mcpLeaseHolderRefusal(err); ok {
			return res, nil
		}
		return toolErr("factory_complete", err), nil
	}
	return mcp.NewToolResultText(strings.TrimRight(out.String(), "\n")), nil
}

// handleFactoryDecide is the MCP form of `moai factory decide` (one card per
// call). The REQ-MN-003 operator gate is the ONE predicate the CLI verb runs
// (factoryDecideGate, M2 Part A absorb): a positive operator mark admits, a
// lane keeps the boundary refusal, an unmarked session is unknown
// attribution. Both identity classes refuse as structured results
// (SPEC-MCP-NEUTRALITY-001 REQ-MN-001) carrying the CLI-identical text.
func handleFactoryDecide(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	launch, err := resolveToolProjectRoot(req)
	if err != nil {
		return toolErr("factory_decide", err), nil
	}
	// The factory root of the caller's tree, as the verb resolves its own (factoryCardRoot): the record,
	// the git-flow config and the queue are read at the parent checkout on both surfaces (card t1680 (a)).
	root := factoryRootOf(launch)
	// The REQ-MN-003 operator gate is the ONE predicate both surfaces run
	// (factoryDecideGate, M2 Part A absorb), split here by CHANNEL (card t1632
	// main-absorb r9, the t1914 parity contract winning on the boundary class):
	// a lane session keeps the CLI-identical boundary refusal (factoryDecideLaneRefusal)
	// on the ERROR channel, exactly as main's handler and the CLI verb surface it,
	// while an unmarked session's unknown-attribution refusal stays a structured
	// result (REQ-MN-001). A positive operator mark admits.
	if factoryLaneRefusal() {
		return toolErr("factory_decide", factoryDecideLaneRefusal()), nil
	}
	if !mcpOperatorAuthority() {
		return toolRefusal(mcpIdentityRefusal(factoryUnknownAttributionError("factory decide").Error())), nil
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

// handleFactoryRenew is the MCP form of `moai factory renew --json`: the same
// factoryRenewCard the verb calls (SPEC-FACTORY-LEASE-RENEW-001 REQ-FLR-002). Every
// renewal outcome is returned as its JSON text, so that a refusal is data the watchdog
// reads and not a tool error. The outcome carries the same detail the verb's error
// carries. Only the lane gates, the actor validation, the stale-run check, and the
// card argument are tool errors or structured refusals.
func handleFactoryRenew(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	launch, err := mcpRequiredProjectRoot(req, "factory_renew")
	if err != nil {
		return toolErr("factory_renew", err), nil
	}
	if !factoryLaneAdmission() {
		return toolRefusal(mcpIdentityRefusal(factoryNotALaneError("renew").Error())), nil
	}
	// The same lane-label resolution the verb runs (factoryLaneLabel, the t1914
	// parity contract, card t1632 main-absorb r9): the actor argument wins, the
	// environment is the default, and the error channel carries the verb's
	// exact wording for both the empty case and a legacy spelling (AC-SD-021).
	lane, err := factoryLaneLabel("renew", mcpFactoryActor(req))
	if err != nil {
		return toolErr("factory_renew", err), nil
	}
	cardID := strings.TrimSpace(req.GetString("card", ""))
	if cardID == "" {
		return toolErr("factory_renew", errors.New("factory renew: card is required")), nil
	}
	// The factory root of the caller's launch tree, by the rule the verb applies to its own
	// launch directory (factoryRootOf). A linked-worktree caller then reads the same lease keys
	// as the verb (sync-audit F2).
	root := factoryRootOf(launch)
	explicitRun := req.GetString("run", "")
	if res := mcpStaleRunCheck(ctx, root, explicitRun, "factory_renew"); res != nil {
		return res, nil
	}
	outcome, _ := factoryRenewCard(ctx, root, cardID, explicitRun, lane)
	body, err := json.Marshal(outcome)
	if err != nil {
		return toolErr("factory_renew", err), nil
	}
	return mcp.NewToolResultText(string(body)), nil
}
