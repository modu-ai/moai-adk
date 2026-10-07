// Package cli — codex_audit_launch.go
//
// The Codex audit launcher: it runs a role whose permission contract says
// `sandbox: read-only` as ONE top-level `codex exec -s read-only` process and
// writes the role's returned text to the destination its caller named.
//
// Why a separate top-level process: Codex applies the PARENT session's
// sandbox to a subagent started with spawn_agent, ignoring the role file's
// own sandbox_mode, so a read-only role spawned from a writing session can
// write. A top-level `codex exec -s read-only` is measured to deny the
// model's writes even when the project config asks for workspace-write. The
// auditor therefore cannot write its verdict itself; the launcher writes the
// returned text for it, byte for byte.
//
// The launcher refuses early and writes nothing when the working root is not
// the caller's own registered worktree, when the destination leaves the
// report tree, when the role is not a read-only contract role, or when the
// final instruction argument would exceed the shared ceiling. Only an
// invocation that actually starts the audit leaves a launch record.
package cli

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/template/agentemit"
)

// Route names recorded in a launch record: how the launcher was reached.
const (
	codexAuditRouteDirect = "direct" // called in-process (tests)
	codexAuditRouteShell  = "shell"  // the `moai codex audit` verb
	codexAuditRouteMCP    = "mcp"    // the codex_role_audit MCP tool
)

const (
	codexAuditSandbox       = "read-only"
	codexAuditRoleDir       = ".codex/agents/moai"
	codexAuditReportsDir    = ".moai/reports"
	codexAuditRecordSubdir  = "codex-audit"
	codexAuditRecordVersion = 1
	codexAuditDisabledURL   = "https://mcp-disabled.invalid"
	codexAuditDisabledCmd   = "__moai_audit_mcp_disabled__"
	codexAuditCovers        = "The read-only guarantee covers the commands and edits the model issues, which the Codex sandbox governs; it does not cover writers outside that sandbox."
)

// codexAuditUnsupported lists the writers the Codex sandbox does not govern.
var codexAuditUnsupported = []string{"codex-home-session-files", "project-hook-commands"}

// codexAuditServerName is the MCP server-name shape the launcher can disable
// through a dotted `-c mcp_servers.<name>.enabled=false` override. A name
// outside it cannot be addressed safely, so the launcher refuses to run.
var codexAuditServerName = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// codexAuditRequest is one launch intent.
type codexAuditRequest struct {
	Role        string        // role name, e.g. plan-auditor
	ProjectRoot string        // the serving checkout's root (repository boundary)
	Root        string        // requested working root
	Out         string        // verdict destination; empty returns on stdout
	Route       string        // direct | shell | mcp
	Program     string        // codex binary; empty resolves "codex" on PATH
	Task        io.Reader     // task text, delivered on the audit's stdin
	Timeout     time.Duration // audit bound; zero uses the configured default
	Stdout      io.Writer
	Stderr      io.Writer
}

// codexAuditResult is the observable outcome of one launch.
type codexAuditResult struct {
	ExitCode    int
	Message     string // returned text (success only)
	RecordPath  string // launch record, relative to the root
	VerdictPath string // verdict, relative to the root
}

// codexAuditRecord is launch record schema version 1.
type codexAuditRecord struct {
	SchemaVersion int      `json:"schema_version"`
	Role          string   `json:"role"`
	Route         string   `json:"route"`
	StartedAt     string   `json:"started_at"`
	EndedAt       string   `json:"ended_at"`
	Argv          []string `json:"argv"`
	Sandbox       string   `json:"sandbox"`
	MCPServers    string   `json:"mcp_servers"`
	Covers        string   `json:"covers"`
	Unsupported   []string `json:"unsupported"`
	ExitCode      int      `json:"exit_code"`
	FailureReason *string  `json:"failure_reason"`
	VerdictPath   *string  `json:"verdict_path"`
	VerdictSHA256 *string  `json:"verdict_sha256"`
}

// Test seams.
var (
	codexAuditRename       = codexAuditRenameIn
	codexAuditRecordSuffix = func() string {
		var b [4]byte
		_, _ = rand.Read(b[:])
		return hex.EncodeToString(b[:])
	}
	codexAuditNow = time.Now
)

// codexAuditRole is the part of an emitted role file the launcher delivers.
// Per-agent effort left with the emitter's frontmatter strip
// (SPEC-AGENT-MODEL-INHERIT-001): the launcher no longer reads
// model_reasoning_effort from the role file — the audit pin
// (workflow.audit.codex.effort) is the surviving effort authority, and an
// unpinned launch omits the directive rather than fabricating a value.
type codexAuditRole struct {
	instructions string
}

// codexAuditPlan is a validated launch: every refusal has already happened,
// the launch record is reserved, and only the audit process remains to run.
type codexAuditPlan struct {
	req        codexAuditRequest
	root       string
	dest       string
	program    string
	role       codexAuditRole
	argv       []string
	instrToken string
	recRel     string
	recFile    *os.File
	recDir     *os.Root // the checked record directory; the verdict may never land in it
	started    time.Time
}

// @MX:ANCHOR: [AUTO] single entry for every read-only role launch (verb, MCP tool, LIVE tests)
// @MX:REASON: the argv contract, confinement, and verbatim write must be identical on every route; a second path would reopen the spawn_agent write hole
// runCodexAudit validates, launches, and records one read-only audit. It
// returns a non-zero ExitCode (never a Go error) for every refusal or
// failure, with the reason on req.Stderr.
func runCodexAudit(ctx context.Context, req codexAuditRequest) (codexAuditResult, error) {
	plan := prepareCodexAudit(ctx, req)
	if plan == nil {
		return codexAuditResult{ExitCode: 1}, nil
	}
	return plan.run(ctx), nil
}

// prepareCodexAudit performs every check that refuses a launch and reserves
// the launch record. A refusal writes its reason to req.Stderr, creates no
// file, and returns nil; the returned plan owns the reserved record.
func prepareCodexAudit(ctx context.Context, req codexAuditRequest) *codexAuditPlan {
	fail := func(format string, args ...any) *codexAuditPlan {
		_, _ = fmt.Fprintf(req.Stderr, "codex audit %s: "+format+"\n", append([]any{req.Role}, args...)...)
		return nil
	}

	root, err := codexAuditValidateRoot(ctx, req.ProjectRoot, req.Root)
	if err != nil {
		return fail("working root rejected: %v", err)
	}
	dest, err := codexAuditValidateDest(root, req.Out)
	if err != nil {
		return fail("destination rejected: %v", err)
	}
	role, err := codexAuditLoadRole(req.Role)
	if err != nil {
		return fail("%v", err)
	}
	encoded, _ := json.Marshal(role.instructions) // a string is always encodable
	instrToken := "developer_instructions=" + string(encoded)
	if err := checkCodexInstructionSize(len(instrToken), "developer_instructions argument"); err != nil {
		return fail("%v", err)
	}
	program := req.Program
	if program == "" {
		if program, err = exec.LookPath("codex"); err != nil {
			return fail("codex binary not found: %v", err)
		}
	}
	disableArgs, err := codexAuditMCPDisableArgs(ctx, program, root)
	if err != nil {
		return fail("cannot list MCP servers to disable: %v", err)
	}

	// Effort rides the audit pin or nothing: pin > absent (codex applies its
	// own default). No per-agent source remains to fall back to. A workflow.yaml
	// that cannot be read or parsed refuses the launch — the error is surfaced
	// on stderr, never folded into an absent pin (SPEC-AUDIT-CEILING-002
	// REQ-ACR-006).
	pins, pinErr := workflowAuditPins(root)
	if pinErr != nil {
		return fail("workflow.audit pins unreadable: %v", pinErr)
	}
	effort := pins.Codex.Effort
	if effort != "" && !codexAuditServerName.MatchString(effort) {
		return fail("workflow.audit.codex effort %q is not a launchable value", effort)
	}
	argv := []string{"exec", "-s", codexAuditSandbox,
		"-c", `approval_policy="never"`}
	if effort != "" {
		argv = append(argv, "-c", `model_reasoning_effort="`+effort+`"`)
	}
	argv = append(argv, "-c", instrToken)
	argv = append(argv, disableArgs...)
	argv = append(argv, "-C", root, "--json", "-")

	started := codexAuditNow().UTC()
	recRel, recFile, recDir, err := codexAuditReserveRecord(root, req.Role, started)
	if err != nil {
		return fail("cannot reserve launch record: %v", err)
	}
	return &codexAuditPlan{req: req, root: root, dest: dest, program: program, role: role,
		argv: argv, instrToken: instrToken, recRel: recRel, recFile: recFile, recDir: recDir, started: started}
}

// run starts the audit process, writes the verdict on success, and completes
// the launch record whatever the outcome.
func (p *codexAuditPlan) run(ctx context.Context) codexAuditResult {
	req := p.req
	defer func() { _ = p.recFile.Close(); _ = p.recDir.Close() }()

	message, runErr := codexAuditExec(ctx, p.program, p.argv, p.root, req.Task, req.Stderr, req.Timeout)
	result := codexAuditResult{RecordPath: p.recRel}
	rec := codexAuditRecord{
		SchemaVersion: codexAuditRecordVersion,
		Role:          req.Role,
		Route:         req.Route,
		StartedAt:     p.started.Format(time.RFC3339),
		Argv:          codexAuditRedactArgv(p.argv, p.instrToken, p.role.instructions),
		Sandbox:       codexAuditSandbox,
		MCPServers:    "disabled",
		Covers:        codexAuditCovers,
		Unsupported:   append([]string(nil), codexAuditUnsupported...),
	}
	var failure string
	switch {
	case runErr != nil:
		failure = runErr.Error()
		rec.ExitCode = codexAuditExitCode(runErr)
	case message == "":
		failure = "empty final message"
	default:
		if p.dest != "" {
			rel, werr := codexAuditWriteVerdict(p.root, p.dest, []byte(message), p.recDir)
			if werr != nil {
				failure = "verdict write failed: " + werr.Error()
				break
			}
			sum := sha256.Sum256([]byte(message))
			hexSum := hex.EncodeToString(sum[:])
			rec.VerdictPath, rec.VerdictSHA256 = &rel, &hexSum
			result.VerdictPath = rel
		} else {
			_, _ = io.WriteString(req.Stdout, message)
		}
		result.Message = message
	}
	if failure != "" {
		if rec.ExitCode == 0 {
			rec.ExitCode = 1
		}
		rec.FailureReason = &failure
		result.ExitCode = 1
	}
	rec.EndedAt = codexAuditNow().UTC().Format(time.RFC3339)
	body, _ := json.MarshalIndent(rec, "", "  ")
	if _, werr := p.recFile.Write(append(body, '\n')); werr != nil && failure == "" {
		failure = "launch record write failed: " + werr.Error()
		result.ExitCode = 1
	}
	_, _ = fmt.Fprintf(req.Stderr, "LAUNCH_RECORD %s\n", p.recRel)
	if failure != "" {
		_, _ = fmt.Fprintf(req.Stderr, "codex audit %s: audit failed: %s\n", req.Role, failure)
	}
	return result
}

// codexAuditValidateRoot accepts root only when, after symlink resolution, it
// is a worktree registered in the same repository as projectRoot — the serving
// checkout. The caller presents its own tree explicitly: a sibling worktree
// and the primary checkout are accepted when registered, and the root is never
// required to equal the toplevel of the directory this server process started
// in (SPEC-CODEX-ROLE-AUDIT-ROOT-001 REQ-001). An absent root is refused, so
// no call ever falls back to a default tree (REQ-002).
func codexAuditValidateRoot(ctx context.Context, projectRoot, root string) (string, error) {
	if root == "" {
		return "", errors.New("worktree root is required")
	}
	if projectRoot == "" {
		return "", errors.New("project root is required")
	}
	resolved, err := codexAuditResolve(root)
	if err != nil {
		return "", fmt.Errorf("%s: %w", root, err)
	}
	rootCommon, err := codexAuditGit(ctx, resolved, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("%s is not a git worktree: %w", resolved, err)
	}
	projectCommon, err := codexAuditGit(ctx, projectRoot, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("project root is not a git repository: %w", err)
	}
	a, errA := codexAuditResolve(rootCommon)
	b, errB := codexAuditResolve(projectCommon)
	if errA != nil || errB != nil || a != b {
		return "", fmt.Errorf("%s belongs to a different repository", resolved)
	}
	list, err := codexAuditGit(ctx, projectRoot, "worktree", "list", "--porcelain")
	if err != nil {
		return "", fmt.Errorf("cannot list worktrees: %w", err)
	}
	for _, line := range strings.Split(list, "\n") {
		path, ok := strings.CutPrefix(line, "worktree ")
		if !ok {
			continue
		}
		if p, err := codexAuditResolve(path); err == nil && p == resolved {
			return resolved, nil
		}
	}
	return "", fmt.Errorf("%s is not a registered worktree of this repository", resolved)
}

// codexAuditValidateDest confines the destination to the root's report tree,
// outside the launcher's own record directory, with no .git component. An
// empty destination means "return on stdout" and is always accepted.
func codexAuditValidateDest(root, out string) (string, error) {
	if out == "" {
		return "", nil
	}
	p := out
	if !filepath.IsAbs(p) {
		p = filepath.Join(root, p)
	}
	p = filepath.Clean(p)
	if codexAuditHasGitComponent(p) {
		return "", fmt.Errorf("%s has a .git path component", out)
	}
	resolved, err := codexAuditResolveMaybeMissing(p)
	if err != nil {
		return "", fmt.Errorf("%s: %w", out, err)
	}
	reports := filepath.Join(root, filepath.FromSlash(codexAuditReportsDir))
	rel, err := filepath.Rel(reports, resolved)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("%s is outside the report tree", out)
	}
	// Names are compared case-insensitively because the default macOS and
	// Windows filesystems are: CODEX-AUDIT and codex-audit are one directory.
	parts := strings.Split(rel, string(filepath.Separator))
	if strings.EqualFold(parts[0], codexAuditRecordSubdir) || codexAuditSameEntry(filepath.Join(reports, parts[0]), filepath.Join(reports, codexAuditRecordSubdir)) {
		return "", fmt.Errorf("%s is inside the launcher's record directory", out)
	}
	if codexAuditHasGitComponent(resolved) {
		return "", fmt.Errorf("%s resolves through a .git path component", out)
	}
	if info, err := os.Stat(resolved); err == nil && info.IsDir() {
		return "", fmt.Errorf("%s is a directory", out)
	}
	return resolved, nil
}

func codexAuditHasGitComponent(p string) bool {
	for _, part := range strings.Split(filepath.ToSlash(p), "/") {
		if strings.EqualFold(part, ".git") {
			return true
		}
	}
	return false
}

// codexAuditSameEntry reports whether two existing paths name the same
// file-system object, which catches aliases a string comparison cannot see
// (case folding, Unicode normalization). A missing path is never the same.
func codexAuditSameEntry(a, b string) bool {
	fa, errA := os.Lstat(a)
	fb, errB := os.Lstat(b)
	return errA == nil && errB == nil && os.SameFile(fa, fb)
}

// codexAuditOpenDir opens root/rel as a directory handle, creating missing
// components. Each component is looked up with Lstat through its parent's
// handle and refused when it is a symlink or not a directory; it is then
// opened from that same parent handle and its identity compared with the
// Lstat result, so a component swapped between the check and the open is
// refused. Every write that follows goes through the returned handle, never
// through the path name, so a later swap of any component cannot redirect
// it. A component that is the same directory as forbid is refused.
//
// @MX:WARN: [AUTO] confinement of launcher writes; a name-based re-walk here would reopen the symlink-swap escape
// @MX:REASON: the audit runs for minutes while a sandboxed writer can replace report directories with symlinks
func codexAuditOpenDir(root, rel string, forbid *os.Root) (*os.Root, error) {
	var forbidInfo os.FileInfo
	if forbid != nil {
		fi, err := forbid.Stat(".")
		if err != nil {
			return nil, err
		}
		forbidInfo = fi
	}
	cur, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	walked := "."
	for _, name := range strings.Split(filepath.ToSlash(filepath.Clean(rel)), "/") {
		if name == "." || name == "" {
			continue
		}
		walked = filepath.Join(walked, name)
		fi, err := cur.Lstat(name)
		if errors.Is(err, fs.ErrNotExist) {
			if err = cur.Mkdir(name, 0o755); err == nil || errors.Is(err, fs.ErrExist) {
				fi, err = cur.Lstat(name)
			}
		}
		if err != nil {
			_ = cur.Close()
			return nil, err
		}
		if fi.Mode()&os.ModeSymlink != 0 || !fi.IsDir() {
			_ = cur.Close()
			return nil, fmt.Errorf("%s is not a plain directory (symlink or other file)", filepath.ToSlash(walked))
		}
		next, err := cur.OpenRoot(name)
		_ = cur.Close()
		if err != nil {
			return nil, err
		}
		got, err := next.Stat(".")
		if err != nil || !os.SameFile(fi, got) {
			_ = next.Close()
			return nil, fmt.Errorf("%s changed while it was being opened", filepath.ToSlash(walked))
		}
		if forbidInfo != nil && os.SameFile(got, forbidInfo) {
			_ = next.Close()
			return nil, fmt.Errorf("%s is the launcher's record directory", filepath.ToSlash(walked))
		}
		cur = next
	}
	return cur, nil
}

// codexAuditResolve returns the absolute, symlink-free form of an existing path.
func codexAuditResolve(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

// codexAuditResolveMaybeMissing resolves the longest existing prefix of p and
// re-attaches the missing tail. A dangling symlink anywhere is refused.
func codexAuditResolveMaybeMissing(p string) (string, error) {
	var tail []string
	cur := p
	for {
		if _, err := os.Lstat(cur); err == nil {
			resolved, err := filepath.EvalSymlinks(cur)
			if err != nil {
				return "", fmt.Errorf("cannot resolve %s: %w", cur, err)
			}
			for i := len(tail) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, tail[i])
			}
			return resolved, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", fmt.Errorf("no existing ancestor for %s", p)
		}
		tail = append(tail, filepath.Base(cur))
		cur = parent
	}
}

// codexAuditRoleFS is the source of both role eligibility and role content:
// the binary's embedded templates. A project's .codex/ copy is not consulted,
// because a Claude-profile project never refreshes it on update and a stale
// copy would refuse a role the binary considers read-only (GitHub #1735).
// Tests replace it to size the instructions without rebuilding the binary.
var codexAuditRoleFS = template.EmbeddedTemplates

// codexAuditLaunchableRoles derives the launchable set from the permission
// contract: every emitted role whose contract sandbox is read-only.
func codexAuditLaunchableRoles() (map[string]bool, error) {
	man, err := agentemit.LoadManifest()
	if err != nil {
		return nil, err
	}
	if man.PermissionContract == nil {
		return nil, errors.New("the Codex mapping manifest carries no permission contract")
	}
	fsys, err := codexAuditRoleFS()
	if err != nil {
		return nil, err
	}
	entries, err := fs.ReadDir(fsys, codexAuditRoleDir)
	if err != nil {
		return nil, err
	}
	set := map[string]bool{}
	for _, e := range entries {
		name, ok := strings.CutSuffix(e.Name(), ".toml")
		if ok && man.PermissionContract.ContractSandbox(name) == codexAuditSandbox {
			set[name] = true
		}
	}
	return set, nil
}

// codexAuditLoadRole checks eligibility and reads the role file from the
// embedded templates (codexAuditRoleFS).
func codexAuditLoadRole(role string) (codexAuditRole, error) {
	set, err := codexAuditLaunchableRoles()
	if err != nil {
		return codexAuditRole{}, fmt.Errorf("cannot derive launchable roles: %w", err)
	}
	if !set[role] {
		return codexAuditRole{}, fmt.Errorf("role %q is not a read-only contract role", role)
	}
	fsys, err := codexAuditRoleFS()
	if err != nil {
		return codexAuditRole{}, fmt.Errorf("role %q: cannot open embedded templates: %w", role, err)
	}
	src, err := fs.ReadFile(fsys, codexAuditRoleDir+"/"+role+".toml")
	if err != nil {
		return codexAuditRole{}, fmt.Errorf("role %q has no emitted role file: %w", role, err)
	}
	fields, err := codexAuditParseRoleTOML(string(src))
	if err != nil {
		return codexAuditRole{}, fmt.Errorf("role %q file: %w", role, err)
	}
	if fields["sandbox_mode"] != codexAuditSandbox {
		return codexAuditRole{}, fmt.Errorf("role %q file declares sandbox_mode %q, not %s", role, fields["sandbox_mode"], codexAuditSandbox)
	}
	instr, ok := fields["developer_instructions"]
	if !ok || instr == "" {
		return codexAuditRole{}, fmt.Errorf("role %q file carries no developer_instructions", role)
	}
	return codexAuditRole{instructions: instr}, nil
}

// codexAuditParseRoleTOML reads the root keys of an emitted role file. It
// accepts only the emitter's grammar: comments, `key = "basic"` values
// without escapes, and `key = ”'` multi-line literals. Parsing stops at the
// first table header.
func codexAuditParseRoleTOML(src string) (map[string]string, error) {
	out := map[string]string{}
	lines := strings.Split(src, "\n")
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			break
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("line %d is not key = value", i+1)
		}
		key, val = strings.TrimSpace(key), strings.TrimSpace(val)
		switch {
		case val == "'''":
			body := strings.Join(lines[i+1:], "\n")
			end := strings.Index(body, "'''")
			if end < 0 {
				return nil, fmt.Errorf("%s literal is not closed", key)
			}
			run := 3
			for end+run < len(body) && body[end+run] == '\'' && run < 5 {
				run++
			}
			end += run - 3 // up to two apostrophes may precede the delimiter
			out[key] = body[:end]
			i += strings.Count(body[:end+3], "\n") + 1
		case len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' && !strings.ContainsAny(val[1:len(val)-1], `"\`):
			out[key] = val[1 : len(val)-1]
		default:
			return nil, fmt.Errorf("%s has a value outside the emitted grammar", key)
		}
	}
	return out, nil
}

// codexAuditMCPDisableArgs disables every server Codex resolves for root.
// Plugin transports are resolved after bootstrap config validation: an
// enabled=false override alone creates an invalid, transport-less table.
// Preserve the transport kind with inert descriptors for bootstrap validation.
// Never forward real URLs or commands: these may contain credentials and would
// leak through process arguments even when launch records redact them.
func codexAuditMCPDisableArgs(ctx context.Context, program, root string) ([]string, error) {
	lctx, cancel := context.WithTimeout(ctx, config.DefaultCodexAuditListTimeout)
	defer cancel()
	cmd := exec.CommandContext(lctx, program, "mcp", "list", "--json")
	cmd.Dir = root
	cmd.Env = codexChildEnv()
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	configureClaudeAuditProcess(cmd)
	cmd.WaitDelay = 2 * time.Second
	if err := runClaudeAuditProcess(cmd); err != nil {
		return nil, fmt.Errorf("mcp list: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var servers []struct {
		Name      string `json:"name"`
		Transport struct {
			Type    string `json:"type"`
			Command string `json:"command"`
			URL     string `json:"url"`
		} `json:"transport"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &servers); err != nil {
		return nil, fmt.Errorf("mcp list output is not a JSON server list: %w", err)
	}
	seen := map[string]bool{}
	args := make([]string, 0, len(servers)*4)
	for _, s := range servers {
		if !codexAuditServerName.MatchString(s.Name) {
			return nil, fmt.Errorf("mcp list names server %q, which cannot be disabled by name", s.Name)
		}
		field, value := "", ""
		switch s.Transport.Type {
		case "stdio":
			field, value = "command", s.Transport.Command
		case "streamable_http":
			field, value = "url", s.Transport.URL
		default:
			return nil, fmt.Errorf("mcp list server %q has an unsupported transport", s.Name)
		}
		if value == "" || strings.ContainsRune(value, 0) {
			return nil, fmt.Errorf("mcp list server %q has an invalid transport", s.Name)
		}
		if seen[s.Name] {
			return nil, fmt.Errorf("mcp list duplicates server %q", s.Name)
		}
		seen[s.Name] = true
		if field == "command" {
			value = codexAuditDisabledCmd
		} else {
			value = codexAuditDisabledURL
		}
		encoded, _ := json.Marshal(value)
		prefix := "mcp_servers." + s.Name
		args = append(args, "-c", prefix+"."+field+"="+string(encoded), "-c", prefix+".enabled=false")
	}
	return args, nil
}

// codexAuditReserveRecord creates the launch record exclusively before the
// audit starts, so a taken name stops the launch instead of losing a record.
// The record directory is opened through codexAuditOpenDir, so neither it nor
// .moai/reports may be a symlink; the returned handle stays open for the run.
func codexAuditReserveRecord(root, role string, at time.Time) (string, *os.File, *os.Root, error) {
	dir, err := codexAuditOpenDir(root, codexAuditReportsDir+"/"+codexAuditRecordSubdir, nil)
	if err != nil {
		return "", nil, nil, err
	}
	name := fmt.Sprintf("%s-%s-%s.launch.json", role, at.Format("20060102T150405Z"), codexAuditRecordSuffix())
	f, err := dir.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		_ = dir.Close()
		return "", nil, nil, err
	}
	return codexAuditReportsDir + "/" + codexAuditRecordSubdir + "/" + name, f, dir, nil
}

// codexAuditExec runs the audit process under a bound and returns its final
// agent message from the --json event stream.
//
// @MX:WARN: [AUTO] kills the whole process group on timeout
// @MX:REASON: an audit that outlives its bound must not leave codex or its children running
func codexAuditExec(ctx context.Context, program string, argv []string, root string, task io.Reader, stderr io.Writer, timeout time.Duration) (string, error) {
	if timeout <= 0 {
		timeout = config.DefaultCodexAuditTimeout
	}
	if task == nil {
		task = strings.NewReader("")
	}
	ectx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ectx, program, argv...)
	cmd.Dir = root
	cmd.Env = codexChildEnv()
	cmd.Stdin = task
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = stderr
	configureClaudeAuditProcess(cmd)
	cmd.WaitDelay = 2 * time.Second
	err := runClaudeAuditProcess(cmd)
	if errors.Is(ectx.Err(), context.DeadlineExceeded) {
		return "", fmt.Errorf("timed out after %s", timeout)
	}
	if err != nil {
		return "", err
	}
	return codexAuditFinalMessage(stdout.Bytes()), nil
}

// codexAuditFinalMessage returns the text of the last agent_message item.
func codexAuditFinalMessage(stream []byte) string {
	var last string
	sc := bufio.NewScanner(bytes.NewReader(stream))
	sc.Buffer(make([]byte, 0, 64*1024), 64*1024*1024)
	for sc.Scan() {
		var ev struct {
			Type string `json:"type"`
			Item struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"item"`
		}
		if json.Unmarshal(sc.Bytes(), &ev) != nil {
			continue
		}
		if ev.Type == "item.completed" && ev.Item.Type == "agent_message" {
			last = ev.Item.Text
		}
	}
	return last
}

func codexAuditExitCode(err error) int {
	var ee *exec.ExitError
	if errors.As(err, &ee) && ee.ExitCode() > 0 {
		return ee.ExitCode()
	}
	return 1
}

// codexAuditWriteVerdict replaces dest atomically: either the previous file
// or the complete new one, never a partial file. The destination's directory
// is opened as a handle by walking from root without following symlinks
// (codexAuditOpenDir); the temporary file is created and renamed inside that
// handle, so the file lands in the directory that was checked even if a path
// component is swapped meanwhile. It returns the worktree-relative path that
// was actually written, for the launch record.
func codexAuditWriteVerdict(root, dest string, data []byte, recDir *os.Root) (string, error) {
	rel, err := filepath.Rel(root, dest)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("%s is not below the working root", dest)
	}
	dir, err := codexAuditOpenDir(root, filepath.Dir(rel), recDir)
	if err != nil {
		return "", err
	}
	defer func() { _ = dir.Close() }()
	base := filepath.Base(rel)
	var rnd [8]byte
	_, _ = rand.Read(rnd[:])
	name := ".codex-audit-" + hex.EncodeToString(rnd[:]) + ".tmp"
	tmp, err := dir.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", err
	}
	defer func() { _ = dir.Remove(name) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return "", err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := codexAuditRename(dir, name, base); err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

// codexAuditRedactArgv replaces instructions and transport values with digests.
func codexAuditRedactArgv(argv []string, token, instructions string) []string {
	sum := sha256.Sum256([]byte(instructions))
	out := make([]string, len(argv))
	for i, a := range argv {
		if a == token {
			a = fmt.Sprintf("developer_instructions=<redacted sha256=%s bytes=%d>", hex.EncodeToString(sum[:]), len(instructions))
		} else if key, value, ok := strings.Cut(a, "="); ok && strings.HasPrefix(key, "mcp_servers.") && (strings.HasSuffix(key, ".url") || strings.HasSuffix(key, ".command")) {
			digest := sha256.Sum256([]byte(value))
			a = fmt.Sprintf("%s=<redacted sha256=%s bytes=%d>", key, hex.EncodeToString(digest[:]), len(value))
		}
		out[i] = a
	}
	return out
}

// codexAuditRenameIn renames within one directory handle. Windows can refuse
// a rename briefly while another process holds the target open, so it is
// retried there, as renameWithRetry does for path-based renames.
func codexAuditRenameIn(dir *os.Root, oldname, newname string) error {
	err := dir.Rename(oldname, newname)
	if err == nil || runtime.GOOS != "windows" {
		return err
	}
	for range 10 {
		time.Sleep(5 * time.Millisecond)
		if err = dir.Rename(oldname, newname); err == nil {
			return nil
		}
	}
	return err
}

// codexAuditGit runs a read-only git query and returns trimmed stdout.
func codexAuditGit(ctx context.Context, dir string, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// runCodexAuditVerb is the shared launcher path of the two CLI verbs
// (`moai codex audit` and `moai codex role-audit`): the working root is the
// git worktree containing the current directory, and every refusal travels
// through the same codexAuditValidateRoot as the MCP route.
func runCodexAuditVerb(cmd *cobra.Command, role string, out string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	top, err := codexAuditGit(cmd.Context(), cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "codex audit %s: current directory is not inside a git worktree\n", role)
		return &exitCodeError{code: 1}
	}
	dest := out
	if dest != "" {
		if dest, err = filepath.Abs(dest); err != nil {
			return err
		}
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	res, _ := runCodexAudit(ctx, codexAuditRequest{
		Role: role, ProjectRoot: top, Root: top, Out: dest,
		Route: codexAuditRouteShell, Task: cmd.InOrStdin(),
		Stdout: cmd.OutOrStdout(), Stderr: cmd.ErrOrStderr(),
	})
	if res.ExitCode != 0 {
		return &exitCodeError{code: res.ExitCode}
	}
	return nil
}

// newCodexAuditCmd builds `moai codex audit <role> [--out <path>]`.
func newCodexAuditCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "audit <role>",
		Short: "Run a read-only role as a top-level read-only Codex process",
		Long: "Run a role whose permission contract is read-only as one top-level\n" +
			"codex exec process with the read-only sandbox, every MCP server\n" +
			"disabled, and the task text read from stdin. The role's returned\n" +
			"text is written verbatim to --out, or printed when --out is absent.\n" +
			"The working root is the git worktree containing the current\n" +
			"directory; --out must stay inside that worktree's report tree.",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCodexAuditVerb(cmd, args[0], out)
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "write the returned text to this path instead of stdout")
	return cmd
}

// newCodexRoleAuditCmd builds `moai codex role-audit <role> [--out <path>]`
// (SPEC-CODEX-ROLE-AUDIT-ROOT-001 REQ-004): the CLI twin of the MCP
// codex_role_audit tool, so a session in its own linked-worktree working
// directory can launch the read-only role audit directly, without the MCP
// server, under the same refusal semantics.
func newCodexRoleAuditCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "role-audit <role>",
		Short: "Run a read-only role against this worktree (the codex_role_audit MCP tool's CLI twin)",
		Long: "The CLI twin of the codex_role_audit MCP tool: run a role whose\n" +
			"permission contract is read-only as one top-level codex exec process\n" +
			"with the read-only sandbox and every MCP server disabled, task text\n" +
			"read from stdin. The working root is the git worktree containing the\n" +
			"current directory, verified against the same registration rules as\n" +
			"the MCP route; --out must stay inside that worktree's report tree.",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCodexAuditVerb(cmd, args[0], out)
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "write the returned text to this path instead of stdout")
	return cmd
}

func init() {
	codexCmd.AddCommand(newCodexAuditCmd())
	codexCmd.AddCommand(newCodexRoleAuditCmd())
}
