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
)

const (
	codexAuditSandbox       = "read-only"
	codexAuditRoleDir       = ".codex/agents/moai"
	codexAuditReportsDir    = ".moai/reports"
	codexAuditRecordSubdir  = "codex-audit"
	codexAuditRecordVersion = 1
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
	ProjectRoot string        // the launcher's own project root
	CallerDir   string        // a directory inside the caller's worktree
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
	codexAuditRename       = renameWithRetry
	codexAuditRecordSuffix = func() string {
		var b [4]byte
		_, _ = rand.Read(b[:])
		return hex.EncodeToString(b[:])
	}
	codexAuditNow = time.Now
)

// codexAuditRole is the part of an emitted role file the launcher delivers.
type codexAuditRole struct {
	instructions string
	effort       string
}

// @MX:ANCHOR: [AUTO] single entry for every read-only role launch (verb, MCP tool, LIVE tests)
// @MX:REASON: the argv contract, confinement, and verbatim write must be identical on every route; a second path would reopen the spawn_agent write hole
// runCodexAudit validates, launches, and records one read-only audit. It
// returns a non-zero ExitCode (never a Go error) for every refusal or
// failure, with the reason on req.Stderr.
func runCodexAudit(ctx context.Context, req codexAuditRequest) (codexAuditResult, error) {
	fail := func(format string, args ...any) (codexAuditResult, error) {
		_, _ = fmt.Fprintf(req.Stderr, "codex audit %s: "+format+"\n", append([]any{req.Role}, args...)...)
		return codexAuditResult{ExitCode: 1}, nil
	}

	root, err := codexAuditValidateRoot(ctx, req.ProjectRoot, req.CallerDir, req.Root)
	if err != nil {
		return fail("working root rejected: %v", err)
	}
	dest, err := codexAuditValidateDest(root, req.Out)
	if err != nil {
		return fail("destination rejected: %v", err)
	}
	role, err := codexAuditLoadRole(root, req.Role)
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
	names, err := codexAuditMCPServerNames(ctx, program, root)
	if err != nil {
		return fail("cannot list MCP servers to disable: %v", err)
	}

	argv := []string{"exec", "-s", codexAuditSandbox,
		"-c", `approval_policy="never"`,
		"-c", `model_reasoning_effort="` + role.effort + `"`,
		"-c", instrToken}
	for _, n := range names {
		argv = append(argv, "-c", "mcp_servers."+n+".enabled=false")
	}
	argv = append(argv, "-C", root, "--json", "-")

	started := codexAuditNow().UTC()
	recRel, recFile, err := codexAuditReserveRecord(root, req.Role, started)
	if err != nil {
		return fail("cannot reserve launch record: %v", err)
	}
	defer func() { _ = recFile.Close() }()

	message, runErr := codexAuditExec(ctx, program, argv, root, req.Task, req.Stderr, req.Timeout)
	result := codexAuditResult{RecordPath: recRel}
	rec := codexAuditRecord{
		SchemaVersion: codexAuditRecordVersion,
		Role:          req.Role,
		Route:         req.Route,
		StartedAt:     started.Format(time.RFC3339),
		Argv:          codexAuditRedactArgv(argv, instrToken, role.instructions),
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
		if dest != "" {
			if werr := codexAuditWriteVerdict(dest, []byte(message)); werr != nil {
				failure = "verdict write failed: " + werr.Error()
				break
			}
			rel := codexAuditRel(root, dest)
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
	if _, werr := recFile.Write(append(body, '\n')); werr != nil && failure == "" {
		failure = "launch record write failed: " + werr.Error()
		result.ExitCode = 1
	}
	_, _ = fmt.Fprintf(req.Stderr, "LAUNCH_RECORD %s\n", recRel)
	if failure != "" {
		_, _ = fmt.Fprintf(req.Stderr, "codex audit %s: audit failed: %s\n", req.Role, failure)
	}
	return result, nil
}

// codexAuditValidateRoot accepts root only when, after symlink resolution, it
// is the top of the caller's own worktree and a worktree registered in the
// same repository as projectRoot. A sibling worktree or the primary checkout
// is refused even though it is registered.
func codexAuditValidateRoot(ctx context.Context, projectRoot, callerDir, root string) (string, error) {
	if root == "" || callerDir == "" || projectRoot == "" {
		return "", errors.New("root, caller directory, and project root are all required")
	}
	resolved, err := codexAuditResolve(root)
	if err != nil {
		return "", fmt.Errorf("%s: %w", root, err)
	}
	callerTop, err := codexAuditGit(ctx, callerDir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("caller directory is not inside a git worktree: %w", err)
	}
	if callerTop, err = codexAuditResolve(callerTop); err != nil {
		return "", err
	}
	if resolved != callerTop {
		return "", fmt.Errorf("%s is not the caller's own worktree (%s)", resolved, callerTop)
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
	parts := strings.Split(rel, string(filepath.Separator))
	if parts[0] == codexAuditRecordSubdir {
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
		if part == ".git" {
			return true
		}
	}
	return false
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
	fsys, err := template.EmbeddedTemplates()
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

// codexAuditLoadRole checks eligibility and reads the role file in root.
func codexAuditLoadRole(root, role string) (codexAuditRole, error) {
	set, err := codexAuditLaunchableRoles()
	if err != nil {
		return codexAuditRole{}, fmt.Errorf("cannot derive launchable roles: %w", err)
	}
	if !set[role] {
		return codexAuditRole{}, fmt.Errorf("role %q is not a read-only contract role", role)
	}
	path := filepath.Join(root, filepath.FromSlash(codexAuditRoleDir), role+".toml")
	src, err := os.ReadFile(path)
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
	effort := fields["model_reasoning_effort"]
	if !codexAuditServerName.MatchString(effort) {
		return codexAuditRole{}, fmt.Errorf("role %q file carries no usable model_reasoning_effort", role)
	}
	return codexAuditRole{instructions: instr, effort: effort}, nil
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

// codexAuditMCPServerNames asks codex which MCP servers are declared, across
// every configuration layer codex itself resolves for root.
func codexAuditMCPServerNames(ctx context.Context, program, root string) ([]string, error) {
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
		Name string `json:"name"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &servers); err != nil {
		return nil, fmt.Errorf("mcp list output is not a JSON server list: %w", err)
	}
	seen := map[string]bool{}
	names := make([]string, 0, len(servers))
	for _, s := range servers {
		if !codexAuditServerName.MatchString(s.Name) {
			return nil, fmt.Errorf("mcp list names server %q, which cannot be disabled by name", s.Name)
		}
		if !seen[s.Name] {
			seen[s.Name] = true
			names = append(names, s.Name)
		}
	}
	return names, nil
}

// codexAuditReserveRecord creates the launch record exclusively before the
// audit starts, so a taken name stops the launch instead of losing a record.
func codexAuditReserveRecord(root, role string, at time.Time) (string, *os.File, error) {
	dir := filepath.Join(root, filepath.FromSlash(codexAuditReportsDir), codexAuditRecordSubdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", nil, err
	}
	name := fmt.Sprintf("%s-%s-%s.launch.json", role, at.Format("20060102T150405Z"), codexAuditRecordSuffix())
	f, err := os.OpenFile(filepath.Join(dir, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", nil, err
	}
	return codexAuditReportsDir + "/" + codexAuditRecordSubdir + "/" + name, f, nil
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
// or the complete new one, never a partial file.
func codexAuditWriteVerdict(dest string, data []byte) error {
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".codex-audit-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer func() { _ = os.Remove(name) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return codexAuditRename(name, dest)
}

// codexAuditRedactArgv replaces the instruction token with its digest.
func codexAuditRedactArgv(argv []string, token, instructions string) []string {
	sum := sha256.Sum256([]byte(instructions))
	out := make([]string, len(argv))
	for i, a := range argv {
		if a == token {
			a = fmt.Sprintf("developer_instructions=<redacted sha256=%s bytes=%d>", hex.EncodeToString(sum[:]), len(instructions))
		}
		out[i] = a
	}
	return out
}

func codexAuditRel(root, p string) string {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return filepath.ToSlash(p)
	}
	return filepath.ToSlash(rel)
}

// codexAuditGit runs a read-only git query and returns trimmed stdout.
func codexAuditGit(ctx context.Context, dir string, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
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
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			top, err := codexAuditGit(cmd.Context(), cwd, "rev-parse", "--show-toplevel")
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "codex audit %s: current directory is not inside a git worktree\n", args[0])
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
				Role: args[0], ProjectRoot: top, CallerDir: cwd, Root: top, Out: dest,
				Route: codexAuditRouteShell, Task: cmd.InOrStdin(),
				Stdout: cmd.OutOrStdout(), Stderr: cmd.ErrOrStderr(),
			})
			if res.ExitCode != 0 {
				return &exitCodeError{code: res.ExitCode}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "write the returned text to this path instead of stdout")
	return cmd
}

func init() {
	codexCmd.AddCommand(newCodexAuditCmd())
}
