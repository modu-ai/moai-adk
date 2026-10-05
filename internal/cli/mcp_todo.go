// mcp_todo.go — the `todo_add` and `todo_list` MCP tools
// (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-014/-024). Thin wrappers over the
// SAME functions the `moai todo` cobra RunE bodies call (design.md §3 — one
// implementation per verb): runTodoAddAppendRoot and runTodoListRoot, each
// anchored at the caller-supplied project_root the existing project_root
// tools resolve (mcp_project_root.go). No core logic is forked here.
//
// The REQ-SD-015 lane boundary rides in the handler: the CLI path gets it
// from the todo tree's PersistentPreRunE, which an MCP call never crosses,
// so todo_add refuses with the SAME one-line refusal the CLI guard prints
// (todoLaneMutationRefusalText) and leaves the queue file byte-identical.
// todo_list is on the read-only allowlist and needs no guard.
package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

// registerTodoMCPTools registers the todo surface's two tools on the shared
// add gate (per-tool enablement map + catalog equality guard).
func registerTodoMCPTools(add func(name string, tool mcp.Tool, handler server.ToolHandlerFunc)) {
	add("todo_add", mcp.NewTool(
		"todo_add",
		mcp.WithDescription("Append one card to the backlog queue. Same implementation as `moai todo add` (without --pick/--force). "+projectRootDesc),
		mcp.WithString("text", mcp.Required(), mcp.Description("The card text.")),
		projectRootOption(),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleTodoAdd)

	add("todo_list", mcp.NewTool(
		"todo_list",
		mcp.WithDescription("Render the backlog queue (lock-free; the default view — live cards plus the dropped count). Same implementation as `moai todo list`. "+projectRootDesc),
		projectRootOption(),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleTodoList)

	add("todo_claim", mcp.NewTool(
		"todo_claim",
		mcp.WithDescription("Claim the oldest queued card under a lease (atomic CAS + lease), or renew a held card's lease with --renew. Same implementation as `moai todo claim`. "+projectRootDesc),
		mcp.WithString("lane", mcp.Description("Attribute the claim to this operator/leader-supplied lane label (optional).")),
		mcp.WithString("renew", mcp.Description("Renew the addressed card's lease (id) instead of claiming a new card (optional).")),
		projectRootOption(),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleTodoClaim)
}

// handleTodoClaim wraps runTodoClaimRoot (todo_claim.go) — the same body
// `moai todo claim` runs, anchored at the resolved root. The REQ-SD-015
// refusal rides in the handler (an MCP call never crosses the todo tree's
// PersistentPreRunE) and is the SAME predicate and the SAME one-line text
// the CLI guard prints — the flag form grants nothing here either
// (SPEC-TODO-CLAIM-LEASE-001 REQ-TCL-013; the surface name comes from the
// claim command itself, the value todoRefuseLaneMutation reads off the
// executing subcommand).
func handleTodoClaim(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root, err := resolveToolProjectRoot(req)
	if err != nil {
		return toolErr("todo_claim", err), nil
	}
	if factoryLaneRefusal() {
		return toolErr("todo_claim", errors.New(todoLaneMutationRefusalText(newTodoClaimCmd().Name()))), nil
	}
	lane := strings.TrimSpace(req.GetString("lane", ""))
	renew := strings.TrimSpace(req.GetString("renew", ""))
	out, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	cmd := newBufferedCommand(out, errBuf)
	if err := runTodoClaimRoot(root, cmd, lane, renew); err != nil {
		// runTodoClaimRoot writes its own output (reclaim lines, the no-card
		// message) before the error returns — surface both faithfully.
		if out.Len() > 0 {
			_ = errBuf
			return toolErr("todo_claim", fmt.Errorf("%s%v", out.String(), err)), nil
		}
		return toolErr("todo_claim", err), nil
	}
	return mcp.NewToolResultText(strings.TrimRight(out.String(), "\n")), nil
}

// handleTodoAdd wraps runTodoAddAppendRoot (todo.go) — the same body
// `moai todo add` runs, anchored at the resolved root.
func handleTodoAdd(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root, err := resolveToolProjectRoot(req)
	if err != nil {
		return toolErr("todo_add", err), nil
	}
	// REQ-SD-015: the queue-mutation refusal holds on a lane marker, a lane
	// label, or the Codex backend — the same predicate and the same one-line
	// refusal the CLI guard prints, and no queue byte moves. The surface
	// name comes from the add command itself, the same value
	// todoRefuseLaneMutation reads off the executing subcommand.
	if factoryLaneRefusal() {
		return toolErr("todo_add", errors.New(todoLaneMutationRefusalText(newTodoAddCmd().Name()))), nil
	}
	text := strings.TrimSpace(req.GetString("text", ""))
	if text == "" {
		return toolErr("todo_add", errors.New("todo add: text must be non-empty")), nil
	}
	out, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	presentation, err := runTodoAddAppendRoot(root, newBufferedCommand(out, errBuf), text, false, todoCardDecider, nil)
	if err != nil {
		return toolErr("todo_add", err), nil
	}
	// SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-005: the result text's first line
	// stays "<id> <pos>"; a non-empty presentation follows after one blank
	// line. An empty presentation adds nothing — the text stays the CLI
	// stdout, byte for byte (MU-21).
	result := strings.TrimRight(out.String(), "\n")
	if presentation != "" {
		result += "\n\n" + presentation
	}
	return mcp.NewToolResultText(result), nil
}

// handleTodoList wraps runTodoListRoot (todo.go) — the same default render
// `moai todo list` prints, anchored at the resolved root.
func handleTodoList(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root, err := resolveToolProjectRoot(req)
	if err != nil {
		return toolErr("todo_list", err), nil
	}
	out, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	if err := runTodoListRoot(root, newBufferedCommand(out, errBuf), false, false, todoListDefaultLimit); err != nil {
		return toolErr("todo_list", err), nil
	}
	return mcp.NewToolResultText(strings.TrimRight(out.String(), "\n")), nil
}

// newBufferedCommand is the minimal cobra command shell the verb cores
// write through (OutOrStdout/ErrOrStderr); nothing executes and no flag is
// defined on it.
func newBufferedCommand(out, errBuf *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.SetOut(out)
	cmd.SetErr(errBuf)
	return cmd
}
