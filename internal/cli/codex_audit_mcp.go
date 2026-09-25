// Package cli — codex_audit_mcp.go
//
// The MCP route to the Codex audit launcher. A Codex lane session cannot start
// the launcher through its shell: a nested `codex exec` under the lane's
// writing sandbox fails before it reaches the model. The MCP server runs
// outside that sandbox and starts in the lane's own worktree, so it can.
//
// Three tools share one in-process job table:
//
//   - codex_role_audit starts a read-only role through the SAME core the
//     `moai codex audit` verb uses (prepareCodexAudit / run). Every refusal
//     happens before a job exists and writes nothing; an accepted call returns
//     a job id at once, so the host's tool timeout never cuts an audit short.
//   - codex_role_audit_status and codex_role_audit_result read the job.
//
// The worktree root is an explicit input and is confined by the core: it must
// be the worktree this server process started in.
package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// Tool names; the registration and the catalog use these values.
const (
	codexRoleAuditToolName       = "codex_role_audit"
	codexRoleAuditStatusToolName = "codex_role_audit_status"
	codexRoleAuditResultToolName = "codex_role_audit_result"
)

// Job states.
const (
	codexRoleAuditRunning   = "running"
	codexRoleAuditCompleted = "completed"
	codexRoleAuditFailed    = "failed"
)

// codexRoleAuditServerDir is the directory this server process started in; it
// names the caller's own worktree. A seam so tests can place the server.
var codexRoleAuditServerDir = os.Getwd

// codexRoleAuditJob is one background launch.
type codexRoleAuditJob struct {
	mu        sync.Mutex
	id        string
	role      string
	state     string
	startedAt time.Time
	endedAt   time.Time
	result    codexAuditResult
	output    bytes.Buffer // the role's text when no destination was named
	diag      bytes.Buffer // launcher diagnostics
}

// lockedWriter serializes writes from the job goroutine against readers.
type lockedWriter struct {
	mu  *sync.Mutex
	buf *bytes.Buffer
}

func (w lockedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Write(p)
}

var (
	codexRoleAuditJobsMu sync.Mutex
	codexRoleAuditJobs   = map[string]*codexRoleAuditJob{}
)

// handleCodexRoleAudit validates synchronously, then runs the audit in the
// background and returns its job id.
//
// @MX:WARN: [AUTO] the audit goroutine outlives the tool call by design
// @MX:REASON: a host tool timeout must not kill a minutes-long audit; the launcher's own bound and process-group cleanup end it
func handleCodexRoleAudit(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	role := strings.TrimSpace(req.GetString("role", ""))
	root := strings.TrimSpace(req.GetString("worktree_root", ""))
	task := req.GetString("task", "")
	out := strings.TrimSpace(req.GetString("out", ""))
	if role == "" || root == "" || strings.TrimSpace(task) == "" {
		return toolErr(codexRoleAuditToolName, errors.New("role, worktree_root, and task are required")), nil
	}
	serverDir, err := codexRoleAuditServerDir()
	if err != nil {
		return toolErr(codexRoleAuditToolName, fmt.Errorf("cannot read the server's start directory: %w", err)), nil
	}
	top, err := codexAuditGit(ctx, serverDir, "rev-parse", "--show-toplevel")
	if err != nil {
		return toolErr(codexRoleAuditToolName, errors.New("the server did not start inside a git worktree")), nil
	}

	id, err := newCodexJobID()
	if err != nil {
		return toolErr(codexRoleAuditToolName, err), nil
	}
	job := &codexRoleAuditJob{id: id, role: role, state: codexRoleAuditRunning, startedAt: codexAuditNow().UTC()}
	plan := prepareCodexAudit(ctx, codexAuditRequest{
		Role: role, ProjectRoot: top, CallerDir: serverDir, Root: root, Out: out,
		Route: codexAuditRouteMCP, Task: strings.NewReader(task),
		Stdout: lockedWriter{&job.mu, &job.output}, Stderr: lockedWriter{&job.mu, &job.diag},
	})
	if plan == nil {
		return toolErr(codexRoleAuditToolName, errors.New(strings.TrimSpace(job.diag.String()))), nil
	}

	codexRoleAuditJobsMu.Lock()
	codexRoleAuditJobs[id] = job
	codexRoleAuditJobsMu.Unlock()
	go func() {
		res := plan.run(context.Background())
		job.mu.Lock()
		defer job.mu.Unlock()
		job.result = res
		job.endedAt = codexAuditNow().UTC()
		job.state = codexRoleAuditCompleted
		if res.ExitCode != 0 {
			job.state = codexRoleAuditFailed
		}
	}()
	return toolJSON(codexRoleAuditToolName, map[string]any{
		"job_id": id, "role": role, "state": codexRoleAuditRunning, "record_path": plan.recRel,
	}), nil
}

func codexRoleAuditLookup(tool string, req mcp.CallToolRequest) (*codexRoleAuditJob, *mcp.CallToolResult) {
	id := strings.TrimSpace(req.GetString("job_id", ""))
	codexRoleAuditJobsMu.Lock()
	job, ok := codexRoleAuditJobs[id]
	codexRoleAuditJobsMu.Unlock()
	if !ok {
		return nil, toolErr(tool, fmt.Errorf("unknown job %q (jobs live only in the server process that started them)", id))
	}
	return job, nil
}

// handleCodexRoleAuditStatus reads a job's lifecycle fields.
func handleCodexRoleAuditStatus(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	job, errRes := codexRoleAuditLookup(codexRoleAuditStatusToolName, req)
	if errRes != nil {
		return errRes, nil
	}
	job.mu.Lock()
	defer job.mu.Unlock()
	status := map[string]any{
		"job_id": job.id, "role": job.role, "state": job.state,
		"started_at": job.startedAt.Format(time.RFC3339),
	}
	if !job.endedAt.IsZero() {
		status["ended_at"] = job.endedAt.Format(time.RFC3339)
	}
	return toolJSON(codexRoleAuditStatusToolName, status), nil
}

// handleCodexRoleAuditResult returns a finished job's outcome, or its current
// state without blocking while it still runs.
func handleCodexRoleAuditResult(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	job, errRes := codexRoleAuditLookup(codexRoleAuditResultToolName, req)
	if errRes != nil {
		return errRes, nil
	}
	job.mu.Lock()
	defer job.mu.Unlock()
	if job.state == codexRoleAuditRunning {
		return toolJSON(codexRoleAuditResultToolName, map[string]any{"job_id": job.id, "state": job.state}), nil
	}
	return toolJSON(codexRoleAuditResultToolName, map[string]any{
		"job_id":       job.id,
		"state":        job.state,
		"exit_code":    job.result.ExitCode,
		"output":       job.output.String(),
		"verdict_path": job.result.VerdictPath,
		"record_path":  job.result.RecordPath,
		"diagnostics":  job.diag.String(),
	}), nil
}

// codexRoleAuditTools returns the three tool declarations for registration.
func codexRoleAuditTools() []struct {
	tool    mcp.Tool
	handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
} {
	jobArg := mcp.WithString("job_id", mcp.Required(), mcp.Description("The job id returned by codex_role_audit."))
	return []struct {
		tool    mcp.Tool
		handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
	}{
		{mcp.NewTool(codexRoleAuditToolName,
			mcp.WithDescription("Start a role whose permission contract is read-only (plan-auditor, sync-auditor, mission-governor, super-advisor) as one top-level codex exec process with the read-only sandbox and every MCP server disabled. Returns a job id at once; read it with codex_role_audit_status and codex_role_audit_result. When out is given, the launcher writes that file with exactly the returned text. The worktree root must be the worktree this server started in; the destination must stay under its .moai/reports/ directory. Use this instead of spawn_agent for these roles."),
			mcp.WithString("role", mcp.Required(), mcp.Description("A read-only contract role name.")),
			mcp.WithString("worktree_root", mcp.Required(), mcp.Description("Your own worktree root (git rev-parse --show-toplevel).")),
			mcp.WithString("task", mcp.Required(), mcp.Description("The task text given to the role on its stdin.")),
			mcp.WithString("out", mcp.Description("Optional verdict or report path under the worktree's .moai/reports/ directory.")),
			mcp.WithReadOnlyHintAnnotation(false),
		), handleCodexRoleAudit},
		{mcp.NewTool(codexRoleAuditStatusToolName,
			mcp.WithDescription("Read a codex_role_audit job's state (running, completed, failed) and timestamps."),
			jobArg,
			mcp.WithReadOnlyHintAnnotation(true),
		), handleCodexRoleAuditStatus},
		{mcp.NewTool(codexRoleAuditResultToolName,
			mcp.WithDescription("Read a finished codex_role_audit job's outcome: exit code, returned text when no destination was named, verdict and launch record paths, and diagnostics. A running job returns its state without blocking."),
			jobArg,
			mcp.WithReadOnlyHintAnnotation(true),
		), handleCodexRoleAuditResult},
	}
}
