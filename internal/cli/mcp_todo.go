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
		mcp.WithDescription("Append one card to the kanban backlog queue. Same implementation as `moai todo add` (without --pick/--force). "+projectRootDesc),
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
	if err := runTodoAddAppendRoot(root, newBufferedCommand(out, errBuf), text, false); err != nil {
		return toolErr("todo_add", err), nil
	}
	return mcp.NewToolResultText(strings.TrimRight(out.String(), "\n")), nil
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
