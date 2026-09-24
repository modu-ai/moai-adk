//go:build !windows

// codex_audit_launch_test.go — deterministic criteria for the Codex audit
// launcher (argv contract, role eligibility, verbatim write, failure paths,
// root/destination confinement, instruction ceiling, launch record).
//
// Every test drives the launcher against a fake `codex` placed first on PATH:
// a shell script that records each invocation's argv (NUL-separated, one file
// per call), answers `mcp list --json` from a canned file, and plays back a
// canned `exec --json` event stream. No model is ever called. The file is
// excluded from Windows because the fake is a POSIX shell script; the
// launcher's Windows compile is covered by the cross-build gate instead.
package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/template/agentemit"
)

// ─── fixtures ──────────────────────────────────────────────────────────────

// auditGit runs git in dir with a neutral identity and fails the test on error.
func auditGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@example.invalid", "-c", "commit.gpgsign=false"}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// auditTempDir returns a symlink-free temp dir (macOS /var → /private/var).
func auditTempDir(t *testing.T) string {
	t.Helper()
	d, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// auditRepo is a primary checkout A with two registered worktrees A1 and A2.
type auditRepo struct {
	base, a, a1, a2 string
}

func newAuditRepo(t *testing.T) auditRepo {
	t.Helper()
	base := auditTempDir(t)
	a := filepath.Join(base, "A")
	if err := os.MkdirAll(a, 0o755); err != nil {
		t.Fatal(err)
	}
	auditGit(t, a, "init", "-q", "-b", "main")
	auditGit(t, a, "commit", "-q", "--allow-empty", "-m", "init")
	a1 := filepath.Join(base, "A1")
	a2 := filepath.Join(base, "A2")
	auditGit(t, a, "worktree", "add", "-q", "-b", "wt-one", a1)
	auditGit(t, a, "worktree", "add", "-q", "-b", "wt-two", a2)
	for _, root := range []string{a, a1, a2} {
		installAuditRoles(t, root)
	}
	return auditRepo{base: base, a: a, a1: a1, a2: a2}
}

// installAuditRoles copies the emitted Codex role files into root.
func installAuditRoles(t *testing.T, root string) {
	t.Helper()
	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, ".codex", "agents", "moai")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	entries, err := fs.ReadDir(fsys, ".codex/agents/moai")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		b, err := fs.ReadFile(fsys, ".codex/agents/moai/"+e.Name())
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dst, e.Name()), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// emittedRoleNames lists the role names the emitter publishes.
func emittedRoleNames(t *testing.T) []string {
	t.Helper()
	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatal(err)
	}
	entries, err := fs.ReadDir(fsys, ".codex/agents/moai")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".toml") {
			names = append(names, strings.TrimSuffix(e.Name(), ".toml"))
		}
	}
	sort.Strings(names)
	return names
}

// roleLiteral extracts a multi-line literal value from an emitted role file
// independently of the launcher's own reader.
func roleLiteral(t *testing.T, src, key string) string {
	t.Helper()
	open := key + " = '''\n"
	i := strings.Index(src, open)
	if i < 0 {
		t.Fatalf("role file has no %s literal", key)
	}
	rest := src[i+len(open):]
	j := strings.Index(rest, "'''\n")
	if j < 0 {
		t.Fatalf("role file %s literal is not closed", key)
	}
	// Up to two apostrophes may precede the closing delimiter.
	for j+3 < len(rest) && rest[j+3] == '\'' {
		j++
	}
	return rest[:j]
}

func roleBasic(t *testing.T, src, key string) string {
	t.Helper()
	m := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + ` = "([^"]*)"$`).FindStringSubmatch(src)
	if m == nil {
		t.Fatalf("role file has no %s", key)
	}
	return m[1]
}

// fakeCodex is the PATH-first `codex` stand-in.
type fakeCodex struct{ dir string }

const fakeCodexScript = `#!/bin/sh
d="$FAKE_CODEX_DIR"
n=$(ls "$d/calls" | wc -l | tr -d ' ')
n=$((n+1))
printf '%s\0' "$@" > "$d/calls/$n"
if [ "$1" = "mcp" ]; then
  [ -f "$d/mcp.out" ] && cat "$d/mcp.out"
  exit "$(cat "$d/mcp.rc" 2>/dev/null || echo 0)"
fi
cat > "$d/stdin.$n"
if [ -f "$d/sleep" ]; then
  sleep 30 &
  echo $! > "$d/child.pid"
  wait
fi
if [ -f "$d/delay" ]; then
  sleep "$(cat "$d/delay")"
fi
[ -f "$d/exec.out" ] && cat "$d/exec.out"
exit "$(cat "$d/exec.rc" 2>/dev/null || echo 0)"
`

func installFakeCodex(t *testing.T) *fakeCodex {
	t.Helper()
	dir := auditTempDir(t)
	bin := filepath.Join(dir, "bin")
	for _, d := range []string{bin, filepath.Join(dir, "calls")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(bin, "codex"), []byte(fakeCodexScript), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_CODEX_DIR", dir)
	f := &fakeCodex{dir: dir}
	f.setMCP(`[]`, 0)
	f.setExec("ok", 0)
	t.Cleanup(func() {
		if b, err := os.ReadFile(filepath.Join(dir, "child.pid")); err == nil {
			if pid, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil && pid > 0 {
				_ = syscall.Kill(pid, syscall.SIGKILL)
			}
		}
	})
	return f
}

func (f *fakeCodex) write(t *testing.T, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(f.dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (f *fakeCodex) setMCP(out string, rc int) {
	_ = os.WriteFile(filepath.Join(f.dir, "mcp.out"), []byte(out), 0o644)
	_ = os.WriteFile(filepath.Join(f.dir, "mcp.rc"), []byte(strconv.Itoa(rc)), 0o644)
}

// setExec plays back a JSONL stream whose LAST agent message is msg (a
// preamble message precedes it, as in a real run).
func (f *fakeCodex) setExec(msg string, rc int) {
	f.setExecRaw(fakeExecStream(msg), rc)
}

func (f *fakeCodex) setExecRaw(stream string, rc int) {
	_ = os.WriteFile(filepath.Join(f.dir, "exec.out"), []byte(stream), 0o644)
	_ = os.WriteFile(filepath.Join(f.dir, "exec.rc"), []byte(strconv.Itoa(rc)), 0o644)
}

func fakeExecStream(msg string) string {
	line := func(v any) string { b, _ := json.Marshal(v); return string(b) + "\n" }
	var b strings.Builder
	b.WriteString(line(map[string]any{"type": "thread.started", "thread_id": "fake"}))
	b.WriteString(line(map[string]any{"type": "turn.started"}))
	b.WriteString(line(map[string]any{"type": "item.completed", "item": map[string]any{"id": "item_0", "type": "agent_message", "text": "working on it\n"}}))
	b.WriteString(line(map[string]any{"type": "item.completed", "item": map[string]any{"id": "item_1", "type": "agent_message", "text": msg}}))
	b.WriteString(line(map[string]any{"type": "turn.completed"}))
	return b.String()
}

// calls returns every recorded invocation's argv, in order.
func (f *fakeCodex) calls(t *testing.T) [][]string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(f.dir, "calls"))
	if err != nil {
		t.Fatal(err)
	}
	out := make([][]string, len(entries))
	for _, e := range entries {
		n, err := strconv.Atoi(e.Name())
		if err != nil || n < 1 || n > len(entries) {
			t.Fatalf("unexpected call record %q", e.Name())
		}
		b, err := os.ReadFile(filepath.Join(f.dir, "calls", e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		args := strings.Split(strings.TrimSuffix(string(b), "\x00"), "\x00")
		if len(b) == 0 {
			args = nil
		}
		out[n-1] = args
	}
	return out
}

func (f *fakeCodex) execCalls(t *testing.T) [][]string {
	var ex [][]string
	for _, c := range f.calls(t) {
		if len(c) > 0 && c[0] == "exec" {
			ex = append(ex, c)
		}
	}
	return ex
}

// auditRun invokes the launcher core with test defaults.
type auditRun struct {
	res    codexAuditResult
	err    error
	stdout string
	stderr string
}

func runAudit(t *testing.T, req codexAuditRequest) auditRun {
	t.Helper()
	var out, errb bytes.Buffer
	req.Stdout, req.Stderr = &out, &errb
	if req.Route == "" {
		req.Route = codexAuditRouteDirect
	}
	if req.Task == nil {
		req.Task = strings.NewReader("audit task body")
	}
	res, err := runCodexAudit(context.Background(), req)
	return auditRun{res: res, err: err, stdout: out.String(), stderr: errb.String()}
}

// auditSnapshotTree maps every path under root to a content fingerprint.
func auditSnapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()
	snap := map[string]string{}
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			target, _ := os.Readlink(p)
			snap[p] = "link:" + target
		case info.IsDir():
			snap[p] = "dir"
		default:
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			snap[p] = sha256Hex(b)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return snap
}

func auditDiffSnapshots(before, after map[string]string) []string {
	var d []string
	for k, v := range after {
		if before[k] != v {
			d = append(d, "changed/added "+k)
		}
	}
	for k := range before {
		if _, ok := after[k]; !ok {
			d = append(d, "removed "+k)
		}
	}
	sort.Strings(d)
	return d
}

func auditFileSHA(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func launchRecordLines(stderr string) []string {
	var out []string
	for _, l := range strings.Split(stderr, "\n") {
		if strings.HasPrefix(l, "LAUNCH_RECORD ") {
			out = append(out, strings.TrimPrefix(l, "LAUNCH_RECORD "))
		}
	}
	return out
}

// ─── AC-CAR-001 ────────────────────────────────────────────────────────────

func TestCodexAuditLaunchArgv(t *testing.T) {
	repo := newAuditRepo(t)
	fake := installFakeCodex(t)
	fake.setMCP(`[{"name":"alpha","enabled":true},{"name":"beta","enabled":false},{"name":"moai","enabled":true}]`, 0)

	roleSrc, err := os.ReadFile(filepath.Join(repo.a1, ".codex", "agents", "moai", "plan-auditor.toml"))
	if err != nil {
		t.Fatal(err)
	}
	wantInstr := roleLiteral(t, string(roleSrc), "developer_instructions")
	wantInstrJSON, _ := json.Marshal(wantInstr)
	wantEffort := roleBasic(t, string(roleSrc), "model_reasoning_effort")

	r := runAudit(t, codexAuditRequest{Role: "plan-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1, Task: strings.NewReader("check SPEC X")})
	if r.err != nil || r.res.ExitCode != 0 {
		t.Fatalf("launch failed: err=%v code=%d stderr=%s", r.err, r.res.ExitCode, r.stderr)
	}
	calls := fake.calls(t)
	if len(calls) != 2 {
		t.Fatalf("fake codex called %d times, want 2 (mcp list, exec): %q", len(calls), calls)
	}
	if strings.Join(calls[0], " ") != "mcp list --json" {
		t.Fatalf("first call = %q, want mcp list --json", calls[0])
	}
	if b, _ := os.ReadFile(filepath.Join(fake.dir, "stdin.2")); string(b) != "check SPEC X" {
		t.Errorf("exec stdin = %q, want the task body", b)
	}
	if violations := codexAuditArgvViolations(calls[1], repo.a1, string(wantInstrJSON), `"`+wantEffort+`"`, []string{"alpha", "beta", "moai"}); len(violations) > 0 {
		t.Fatalf("exec argv outside the allowlist:\n%s", strings.Join(violations, "\n"))
	}

	// Mutant guard: the allowlist judge must reject argv the launcher must never send.
	for name, mutate := range map[string]func([]string) []string{
		"second sandbox":  func(a []string) []string { return append(a[:len(a)-1:len(a)-1], "-s", "read-only", "-") },
		"workspace-write": func(a []string) []string { c := append([]string{}, a...); c[2] = "workspace-write"; return c },
		"table override": func(a []string) []string {
			return append(append([]string{}, a[:len(a)-1]...), "-c", "mcp_servers={}", "-")
		},
		"missing disable": func(a []string) []string { return removeTokenPair(a, "mcp_servers.beta.enabled=false") },
		"bypass flag": func(a []string) []string {
			return append(append([]string{}, a[:len(a)-1]...), "--dangerously-bypass-approvals-and-sandbox", "-")
		},
		"ignore-user-config": func(a []string) []string {
			return append(append([]string{}, a[:len(a)-1]...), "--ignore-user-config", "-")
		},
		"sandbox_mode key": func(a []string) []string {
			return append(append([]string{}, a[:len(a)-1]...), "-c", `sandbox_mode="read-only"`, "-")
		},
		"other root": func(a []string) []string { return replaceAfter(a, "-C", repo.a2) },
		"duplicated disable": func(a []string) []string {
			return append(append([]string{}, a[:len(a)-1]...), "-c", "mcp_servers.alpha.enabled=false", "-")
		},
		"altered instruction": func(a []string) []string { return replaceKV(a, "developer_instructions", `"x"`) },
	} {
		if v := codexAuditArgvViolations(mutate(append([]string{}, calls[1]...)), repo.a1, string(wantInstrJSON), `"`+wantEffort+`"`, []string{"alpha", "beta", "moai"}); len(v) == 0 {
			t.Errorf("allowlist judge accepted mutant %q", name)
		}
	}

	// Negative cases: a failing or malformed `mcp list` stops before exec and writes nothing.
	for name, set := range map[string]func(){
		"mcp list exit 1":   func() { fake.setMCP(`[]`, 1) },
		"mcp list not json": func() { fake.setMCP(`not json`, 0) },
		"mcp list bad name": func() { fake.setMCP(`[{"name":"a.b"}]`, 0) },
	} {
		t.Run(name, func(t *testing.T) {
			set()
			before := len(fake.calls(t))
			snap := auditSnapshotTree(t, repo.base)
			r := runAudit(t, codexAuditRequest{Role: "plan-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1, Out: filepath.Join(repo.a1, ".moai", "reports", "neg", "v.md")})
			if r.res.ExitCode == 0 {
				t.Fatalf("launcher succeeded despite %s", name)
			}
			if got := fake.execCalls(t); len(fake.calls(t)) != before+1 || len(got) != 1 {
				t.Fatalf("exec was started after %s (calls %d → %d)", name, before, len(fake.calls(t)))
			}
			if !strings.Contains(r.stderr, "mcp") {
				t.Errorf("stderr names no reason: %q", r.stderr)
			}
			if d := auditDiffSnapshots(snap, auditSnapshotTree(t, repo.base)); len(d) > 0 {
				t.Fatalf("files written after %s: %v", name, d)
			}
			if len(launchRecordLines(r.stderr)) != 0 {
				t.Errorf("LAUNCH_RECORD printed after %s", name)
			}
		})
	}

	t.Run("empty list adds no mcp overrides", func(t *testing.T) {
		fake.setMCP(`[]`, 0)
		r := runAudit(t, codexAuditRequest{Role: "plan-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1})
		if r.res.ExitCode != 0 {
			t.Fatalf("launch failed: %s", r.stderr)
		}
		ex := fake.execCalls(t)
		last := ex[len(ex)-1]
		for _, a := range last {
			if strings.HasPrefix(a, "mcp_servers") {
				t.Fatalf("empty list produced override %q", a)
			}
		}
		if v := codexAuditArgvViolations(last, repo.a1, string(wantInstrJSON), `"`+wantEffort+`"`, nil); len(v) > 0 {
			t.Fatalf("argv violations with empty list: %v", v)
		}
	})
}

// codexAuditArgvViolations is the allowlist judge for one exec argv.
func codexAuditArgvViolations(argv []string, root, instrJSON, effort string, names []string) []string {
	var v []string
	if len(argv) == 0 || argv[0] != "exec" {
		return []string{"first token is not exec"}
	}
	wantNames := map[string]int{}
	for _, n := range names {
		wantNames[n] = 0
	}
	sandboxes, roots, jsons, stdins := 0, 0, 0, 0
	keys := map[string]int{}
	for i := 1; i < len(argv); i++ {
		tok := argv[i]
		for _, bad := range []string{"workspace-write", "danger-full-access", "spawn_agent"} {
			if strings.Contains(tok, bad) {
				v = append(v, "forbidden string "+bad+" in "+tok)
			}
		}
		switch tok {
		case "-s", "--sandbox":
			sandboxes++
			if i+1 >= len(argv) || argv[i+1] != "read-only" {
				v = append(v, "sandbox value is not read-only")
			}
			i++
		case "-c":
			if i+1 >= len(argv) {
				v = append(v, "dangling -c")
				continue
			}
			i++
			key, val, ok := strings.Cut(argv[i], "=")
			if !ok {
				v = append(v, "-c without key=value: "+argv[i])
				continue
			}
			keys[key]++
			switch {
			case key == "approval_policy":
				if val != `"never"` {
					v = append(v, "approval_policy="+val)
				}
			case key == "model_reasoning_effort":
				if val != effort {
					v = append(v, "model_reasoning_effort="+val)
				}
			case key == "developer_instructions":
				if val != instrJSON {
					v = append(v, "developer_instructions differs from the role file")
				}
			case strings.HasPrefix(key, "mcp_servers.") && strings.HasSuffix(key, ".enabled"):
				n := strings.TrimSuffix(strings.TrimPrefix(key, "mcp_servers."), ".enabled")
				if _, ok := wantNames[n]; !ok || val != "false" {
					v = append(v, "unexpected mcp override "+argv[i])
				} else {
					wantNames[n]++
				}
			default:
				v = append(v, "-c key outside allowlist: "+key)
			}
		case "-C":
			roots++
			if i+1 >= len(argv) || argv[i+1] != root {
				v = append(v, "-C is not the caller root")
			}
			i++
		case "--json":
			jsons++
		case "-o":
			if i+1 >= len(argv) || strings.HasPrefix(argv[i+1], root) || !strings.HasPrefix(argv[i+1], os.TempDir()) {
				v = append(v, "-o outside an OS temp path")
			}
			i++
		case "-":
			stdins++
			if i != len(argv)-1 {
				v = append(v, "stdin marker is not last")
			}
		default:
			v = append(v, "token outside allowlist: "+tok)
		}
	}
	if sandboxes != 1 {
		v = append(v, fmt.Sprintf("%d sandbox tokens, want 1", sandboxes))
	}
	if roots != 1 || jsons != 1 || stdins != 1 {
		v = append(v, fmt.Sprintf("-C=%d --json=%d -=%d, want 1 each", roots, jsons, stdins))
	}
	for _, k := range []string{"approval_policy", "model_reasoning_effort", "developer_instructions"} {
		if keys[k] != 1 {
			v = append(v, fmt.Sprintf("-c %s appears %d times, want 1", k, keys[k]))
		}
	}
	for n, c := range wantNames {
		if c != 1 {
			v = append(v, fmt.Sprintf("mcp_servers.%s.enabled=false appears %d times, want 1", n, c))
		}
	}
	return v
}

func removeTokenPair(a []string, value string) []string {
	var out []string
	for i := 0; i < len(a); i++ {
		if a[i] == "-c" && i+1 < len(a) && a[i+1] == value {
			i++
			continue
		}
		out = append(out, a[i])
	}
	return out
}

func replaceAfter(a []string, flag, value string) []string {
	c := append([]string{}, a...)
	for i := range c {
		if c[i] == flag && i+1 < len(c) {
			c[i+1] = value
		}
	}
	return c
}

func replaceKV(a []string, key, value string) []string {
	c := append([]string{}, a...)
	for i := range c {
		if strings.HasPrefix(c[i], key+"=") {
			c[i] = key + "=" + value
		}
	}
	return c
}

// ─── AC-CAR-002 ────────────────────────────────────────────────────────────

func TestCodexAuditLaunchRoleEligibility(t *testing.T) {
	man, err := agentemit.LoadManifest()
	if err != nil {
		t.Fatal(err)
	}
	if man.PermissionContract == nil {
		t.Fatal("manifest carries no permission contract")
	}
	roles := emittedRoleNames(t)
	want := map[string]bool{}
	for _, r := range roles {
		if man.PermissionContract.ContractSandbox(r) == "read-only" {
			want[r] = true
		}
	}
	if len(want) == 0 {
		t.Fatal("contract-derived read-only role set is empty — vacuous")
	}

	repo := newAuditRepo(t)
	fake := installFakeCodex(t)
	got := map[string]bool{}
	for _, role := range append(append([]string{}, roles...), "no-such-role") {
		before := len(fake.calls(t))
		r := runAudit(t, codexAuditRequest{Role: role, ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1})
		after := len(fake.calls(t))
		if r.res.ExitCode == 0 {
			got[role] = true
			if after-before != 2 {
				t.Errorf("%s: %d codex calls, want 2", role, after-before)
			}
			continue
		}
		if after != before {
			t.Errorf("%s: rejected role still called codex (%d → %d)", role, before, after)
		}
		if !strings.Contains(r.stderr, role) {
			t.Errorf("%s: diagnostic does not name the role: %q", role, r.stderr)
		}
	}
	if fmt.Sprint(auditSortedKeys(got)) != fmt.Sprint(auditSortedKeys(want)) {
		t.Fatalf("launchable set = %v, contract read-only set = %v", auditSortedKeys(got), auditSortedKeys(want))
	}

	// The contract, not the role file, decides eligibility: a writing role
	// whose file was edited to claim read-only is still refused.
	var writer string
	for _, r := range roles {
		if !want[r] {
			writer = r
			break
		}
	}
	if writer == "" {
		t.Fatal("no writing role to tamper with — vacuous")
	}
	writerFile := filepath.Join(repo.a1, ".codex", "agents", "moai", writer+".toml")
	writerSrc, err := os.ReadFile(writerFile)
	if err != nil {
		t.Fatal(err)
	}
	tampered := regexp.MustCompile(`(?m)^sandbox_mode = "[^"]*"$`).ReplaceAllString(string(writerSrc), `sandbox_mode = "read-only"`)
	if err := os.WriteFile(writerFile, []byte(tampered), 0o644); err != nil {
		t.Fatal(err)
	}
	before := len(fake.calls(t))
	if r := runAudit(t, codexAuditRequest{Role: writer, ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1}); r.res.ExitCode == 0 || len(fake.calls(t)) != before {
		t.Fatalf("%s with a tampered read-only file was launched", writer)
	}

	// A read-only contract role whose file drifted to a writing sandbox is refused.
	var reader string
	for _, r := range roles {
		if want[r] {
			reader = r
			break
		}
	}
	readerFile := filepath.Join(repo.a1, ".codex", "agents", "moai", reader+".toml")
	readerSrc, err := os.ReadFile(readerFile)
	if err != nil {
		t.Fatal(err)
	}
	drifted := regexp.MustCompile(`(?m)^sandbox_mode = "[^"]*"$`).ReplaceAllString(string(readerSrc), `sandbox_mode = "workspace-write"`)
	if err := os.WriteFile(readerFile, []byte(drifted), 0o644); err != nil {
		t.Fatal(err)
	}
	if r := runAudit(t, codexAuditRequest{Role: reader, ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1}); r.res.ExitCode == 0 || len(fake.calls(t)) != before {
		t.Fatalf("%s with a drifted writing file was launched", reader)
	}
	if err := os.WriteFile(readerFile, readerSrc, 0o644); err != nil {
		t.Fatal(err)
	}

	// A contract read-only role whose role file is missing is rejected too.
	missing := reader
	if err := os.Remove(filepath.Join(repo.a1, ".codex", "agents", "moai", missing+".toml")); err != nil {
		t.Fatal(err)
	}
	r := runAudit(t, codexAuditRequest{Role: missing, ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1})
	if r.res.ExitCode == 0 || len(fake.calls(t)) != before || !strings.Contains(r.stderr, missing) {
		t.Fatalf("role with no emitted file: code=%d calls %d→%d stderr=%q", r.res.ExitCode, before, len(fake.calls(t)), r.stderr)
	}
}

func auditSortedKeys(m map[string]bool) []string {
	var k []string
	for x := range m {
		k = append(k, x)
	}
	sort.Strings(k)
	return k
}

// ─── AC-CAR-003 ────────────────────────────────────────────────────────────

func TestCodexAuditLaunchVerbatimWrite(t *testing.T) {
	repo := newAuditRepo(t)
	fake := installFakeCodex(t)
	msg := "## 판정\n\nPASS — 모든 항목 통과\n\n  들여쓴 줄\t탭\n마지막 줄\n"
	fake.setExec(msg, 0)
	want := sha256.Sum256([]byte(msg))
	dest := filepath.Join(repo.a1, ".moai", "reports", "x", "v.md")

	t.Run("fresh destination", func(t *testing.T) {
		r := runAudit(t, codexAuditRequest{Role: "sync-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1, Out: dest})
		if r.res.ExitCode != 0 {
			t.Fatalf("launch failed: %s", r.stderr)
		}
		if got := auditFileSHA(t, dest); got != hex.EncodeToString(want[:]) {
			t.Fatalf("verdict sha %s, returned text sha %s", got, hex.EncodeToString(want[:]))
		}
	})

	t.Run("existing destination is replaced whole", func(t *testing.T) {
		if err := os.WriteFile(dest, []byte(strings.Repeat("old verdict line\n", 500)), 0o644); err != nil {
			t.Fatal(err)
		}
		r := runAudit(t, codexAuditRequest{Role: "sync-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1, Out: dest})
		if r.res.ExitCode != 0 {
			t.Fatalf("launch failed: %s", r.stderr)
		}
		if got := auditFileSHA(t, dest); got != hex.EncodeToString(want[:]) {
			t.Fatalf("verdict sha %s after overwrite, want %s", got, hex.EncodeToString(want[:]))
		}
	})

	t.Run("interrupted write leaves the previous file", func(t *testing.T) {
		prev := []byte("previous verdict\n")
		if err := os.WriteFile(dest, prev, 0o644); err != nil {
			t.Fatal(err)
		}
		orig := codexAuditRename
		codexAuditRename = func(string, string) error { return errors.New("simulated interruption") }
		t.Cleanup(func() { codexAuditRename = orig })
		r := runAudit(t, codexAuditRequest{Role: "sync-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1, Out: dest})
		codexAuditRename = orig
		if r.res.ExitCode == 0 {
			t.Fatal("launcher reported success although the write was interrupted")
		}
		if got, _ := os.ReadFile(dest); !bytes.Equal(got, prev) {
			t.Fatalf("destination changed after an interrupted write: %q", got)
		}
		entries, _ := os.ReadDir(filepath.Dir(dest))
		if len(entries) != 1 {
			t.Fatalf("temporary files left beside the destination: %v", entries)
		}
	})
}

// ─── AC-CAR-004 ────────────────────────────────────────────────────────────

func TestCodexAuditLaunchFailureWritesNothing(t *testing.T) {
	repo := newAuditRepo(t)
	fake := installFakeCodex(t)
	modes := []struct {
		name, reason string
		setup        func()
		timeout      time.Duration
	}{
		{"non-zero exit", "exit", func() { fake.setExec("partial", 3) }, 0},
		{"timeout", "timed out", func() { fake.setExec("late", 0); fake.write(t, "sleep", "1") }, 500 * time.Millisecond},
		{"empty final message", "empty", func() { fake.setExec("", 0) }, 0},
	}
	for _, m := range modes {
		for _, existing := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/existing=%v", m.name, existing), func(t *testing.T) {
				_ = os.Remove(filepath.Join(fake.dir, "sleep"))
				_ = os.Remove(filepath.Join(fake.dir, "child.pid"))
				// Cleanup-guaranteed: whatever this subtest's fake started is
				// killed even when an assertion stops the subtest early.
				t.Cleanup(func() {
					if b, err := os.ReadFile(filepath.Join(fake.dir, "child.pid")); err == nil {
						if pid, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil && pid > 0 {
							_ = syscall.Kill(pid, syscall.SIGKILL)
						}
					}
				})
				m.setup()
				dest := filepath.Join(repo.a1, ".moai", "reports", "fail", strings.ReplaceAll(m.name, " ", "-")+fmt.Sprint(existing)+".md")
				var beforeSHA string
				if existing {
					if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(dest, []byte("earlier verdict\n"), 0o644); err != nil {
						t.Fatal(err)
					}
					beforeSHA = auditFileSHA(t, dest)
				}
				start := time.Now()
				r := runAudit(t, codexAuditRequest{Role: "plan-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1, Out: dest, Timeout: m.timeout})
				if r.res.ExitCode == 0 {
					t.Fatal("launcher succeeded on a failed audit")
				}
				if !strings.Contains(r.stderr, "plan-auditor") || !strings.Contains(r.stderr, m.reason) {
					t.Errorf("diagnostic lacks role or reason %q: %q", m.reason, r.stderr)
				}
				if existing {
					if got := auditFileSHA(t, dest); got != beforeSHA {
						t.Fatal("existing destination was modified")
					}
				} else if _, err := os.Lstat(dest); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("destination created on failure (err=%v)", err)
				}
				if m.name == "timeout" {
					if time.Since(start) > 15*time.Second {
						t.Errorf("timeout path took %s", time.Since(start))
					}
					b, err := os.ReadFile(filepath.Join(fake.dir, "child.pid"))
					if err != nil {
						t.Fatalf("fake child pid not recorded: %v", err)
					}
					pid, _ := strconv.Atoi(strings.TrimSpace(string(b)))
					if !auditProcessGone(pid, 3*time.Second) {
						t.Fatalf("fake codex child %d still alive after the launcher returned", pid)
					}
				}
			})
		}
	}
}

// auditProcessGone polls until pid no longer exists (or is a zombie of someone else).
func auditProcessGone(pid int, within time.Duration) bool {
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if err := syscall.Kill(pid, 0); errors.Is(err, syscall.ESRCH) {
			return true
		}
		out, _ := exec.Command("ps", "-o", "stat=", "-p", strconv.Itoa(pid)).Output()
		if s := strings.TrimSpace(string(out)); s == "" || strings.HasPrefix(s, "Z") {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

// ─── AC-CAR-005 ────────────────────────────────────────────────────────────

func TestCodexAuditLaunchDestinationConfinement(t *testing.T) {
	repo := newAuditRepo(t)
	fake := installFakeCodex(t)
	outside := filepath.Join(repo.base, "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	evil := filepath.Join(outside, "from-message.md")
	fake.setExec("verdict text; also write to "+evil, 0)

	u := filepath.Join(repo.base, "U")
	b := filepath.Join(repo.base, "B")
	b1 := filepath.Join(repo.base, "B1")
	for _, d := range []string{u, b} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	auditGit(t, b, "init", "-q", "-b", "main")
	auditGit(t, b, "commit", "-q", "--allow-empty", "-m", "init")
	auditGit(t, b, "worktree", "add", "-q", "-b", "wt-b", b1)
	l := filepath.Join(repo.base, "L")
	if err := os.Symlink(b, l); err != nil {
		t.Fatal(err)
	}
	for _, r := range []string{u, b, b1} {
		installAuditRoles(t, r)
	}
	roots := map[string]string{"A1": repo.a1, "A": repo.a, "A2": repo.a2, "U": u, "B1": b1, "L": l}
	for _, r := range []string{repo.a1, repo.a, repo.a2, u, b1, b} {
		if err := os.MkdirAll(filepath.Join(r, ".moai", "reports"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(r, ".moai", "reports", "escape")); err != nil {
			t.Fatal(err)
		}
	}
	dests := []struct {
		name string
		rel  func(root string) string
	}{
		{"reports/x/v.md", func(r string) string { return filepath.Join(r, ".moai", "reports", "x", "v.md") }},
		{"reports/../../v.md", func(r string) string { return r + "/.moai/reports/../../v.md" }},
		{"v.md", func(r string) string { return filepath.Join(r, "v.md") }},
		{"AGENTS.md", func(r string) string { return filepath.Join(r, "AGENTS.md") }},
		{".codex/config.toml", func(r string) string { return filepath.Join(r, ".codex", "config.toml") }},
		{".git/v.md", func(r string) string { return filepath.Join(r, ".git", "v.md") }},
		{"reports/x/.git/v.md", func(r string) string { return filepath.Join(r, ".moai", "reports", "x", ".git", "v.md") }},
		{"reports/codex-audit/v.md", func(r string) string { return filepath.Join(r, ".moai", "reports", "codex-audit", "v.md") }},
		{"outside absolute", func(string) string { return filepath.Join(outside, "abs.md") }},
		{"reports symlink escape", func(r string) string { return filepath.Join(r, ".moai", "reports", "escape", "v.md") }},
	}
	accepted := 0
	for rootName, root := range roots {
		for _, d := range dests {
			t.Run(rootName+"/"+d.name, func(t *testing.T) {
				dest := d.rel(root)
				before := len(fake.calls(t))
				snap := auditSnapshotTree(t, repo.base)
				r := runAudit(t, codexAuditRequest{Role: "plan-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: root, Out: dest})
				if rootName == "A1" && d.name == "reports/x/v.md" {
					accepted++
					if r.res.ExitCode != 0 {
						t.Fatalf("the one legal combination was rejected: %s", r.stderr)
					}
					if len(fake.calls(t))-before != 2 {
						t.Fatalf("legal combination: %d codex calls, want 2", len(fake.calls(t))-before)
					}
					if _, err := os.Stat(dest); err != nil {
						t.Fatalf("verdict not written: %v", err)
					}
					if len(launchRecordLines(r.stderr)) != 1 {
						t.Fatalf("no launch record reported: %q", r.stderr)
					}
					return
				}
				if r.res.ExitCode == 0 {
					t.Fatal("illegal combination accepted")
				}
				if n := len(fake.calls(t)) - before; n != 0 {
					t.Fatalf("illegal combination called codex %d times", n)
				}
				if d := auditDiffSnapshots(snap, auditSnapshotTree(t, repo.base)); len(d) > 0 {
					t.Fatalf("illegal combination changed files: %v", d)
				}
				if r.stdout != "" || len(launchRecordLines(r.stderr)) != 0 || strings.TrimSpace(r.stderr) == "" {
					t.Fatalf("rejection not reported on stderr only: stdout=%q stderr=%q", r.stdout, r.stderr)
				}
			})
		}
	}
	if accepted != 1 {
		t.Fatalf("legal combination ran %d times, want 1", accepted)
	}
	if _, err := os.Lstat(evil); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("a path named in the returned message was created: %v", err)
	}
}

// ─── AC-CAR-006 ────────────────────────────────────────────────────────────

func TestCodexAuditLaunchInstructionCeiling(t *testing.T) {
	repo := newAuditRepo(t)
	fake := installFakeCodex(t)
	ceiling := config.DefaultCodexInstructionArgBytes
	const prefix = "developer_instructions="
	writeRole := func(n int) int {
		body := strings.Repeat("a", n)
		enc, _ := json.Marshal(body)
		role := "# Generated by the MoAI agent dual-publication emitter; regenerate, do not edit.\n" +
			"name = \"plan-auditor\"\n" +
			"description = '''\nsized role\n'''\n" +
			"developer_instructions = '''\n" + body + "'''\n" +
			"model_reasoning_effort = \"high\"\n" +
			"sandbox_mode = \"read-only\"\n"
		if err := os.WriteFile(filepath.Join(repo.a1, ".codex", "agents", "moai", "plan-auditor.toml"), []byte(role), 0o644); err != nil {
			t.Fatal(err)
		}
		return len(prefix) + len(enc)
	}
	base := ceiling - len(prefix) - 2

	if got := writeRole(base + 1); got != ceiling+1 {
		t.Fatalf("over-ceiling fixture token is %d bytes, want %d", got, ceiling+1)
	}
	before := len(fake.calls(t))
	r := runAudit(t, codexAuditRequest{Role: "plan-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1})
	if r.res.ExitCode == 0 {
		t.Fatal("over-ceiling instructions were launched")
	}
	if len(fake.calls(t)) != before {
		t.Fatal("codex was called for an over-ceiling role")
	}
	if !strings.Contains(r.stderr, strconv.Itoa(ceiling+1)) || !strings.Contains(r.stderr, strconv.Itoa(ceiling)) {
		t.Fatalf("diagnostic lacks measured length and ceiling: %q", r.stderr)
	}

	if got := writeRole(base); got != ceiling {
		t.Fatalf("at-ceiling fixture token is %d bytes, want %d", got, ceiling)
	}
	r = runAudit(t, codexAuditRequest{Role: "plan-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1})
	if r.res.ExitCode != 0 {
		t.Fatalf("at-ceiling instructions rejected: %s", r.stderr)
	}
	if ex := fake.execCalls(t); len(ex) != 1 {
		t.Fatalf("exec called %d times, want 1", len(ex))
	}
}

// ─── AC-CAR-014 ────────────────────────────────────────────────────────────

func TestCodexAuditLaunchRecord(t *testing.T) {
	repo := newAuditRepo(t)
	fake := installFakeCodex(t)
	fake.setMCP(`[{"name":"moai"}]`, 0)
	recordDir := filepath.Join(repo.a1, ".moai", "reports", "codex-audit")

	// Rejected invocations first: none may create the record directory.
	t.Run("rejections leave no file", func(t *testing.T) {
		big := strings.Repeat("b", config.DefaultCodexInstructionArgBytes)
		cases := map[string]codexAuditRequest{
			"destination": {Role: "plan-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1, Out: filepath.Join(repo.a1, "AGENTS.md")},
			"eligibility": {Role: "manager-docs", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1},
			"ceiling":     {Role: "mission-governor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1},
		}
		roleFile := filepath.Join(repo.a1, ".codex", "agents", "moai", "mission-governor.toml")
		orig, err := os.ReadFile(roleFile)
		if err != nil {
			t.Fatal(err)
		}
		for name, req := range cases {
			if name == "ceiling" {
				oversized := strings.Replace(string(orig), "developer_instructions = '''\n", "developer_instructions = '''\n"+big+"\n", 1)
				if err := os.WriteFile(roleFile, []byte(oversized), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			snap := auditSnapshotTree(t, repo.base)
			r := runAudit(t, req)
			if r.res.ExitCode == 0 {
				t.Fatalf("%s: rejected invocation succeeded", name)
			}
			if len(launchRecordLines(r.stderr)) != 0 {
				t.Fatalf("%s: LAUNCH_RECORD printed for a rejection", name)
			}
			if d := auditDiffSnapshots(snap, auditSnapshotTree(t, repo.base)); len(d) > 0 {
				t.Fatalf("%s: files changed: %v", name, d)
			}
		}
		if err := os.WriteFile(roleFile, orig, 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(recordDir); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("record directory exists after rejections only: %v", err)
		}
		if n := len(fake.calls(t)); n != 0 {
			t.Fatalf("codex called %d times by rejected invocations", n)
		}
	})

	roleSrc, err := os.ReadFile(filepath.Join(repo.a1, ".codex", "agents", "moai", "plan-auditor.toml"))
	if err != nil {
		t.Fatal(err)
	}
	instr := roleLiteral(t, string(roleSrc), "developer_instructions")
	instrSum := sha256.Sum256([]byte(instr))

	readRecord := func(t *testing.T, r auditRun) (string, map[string]any, []byte) {
		t.Helper()
		lines := launchRecordLines(r.stderr)
		if len(lines) != 1 {
			t.Fatalf("want exactly one LAUNCH_RECORD line, got %d: %q", len(lines), r.stderr)
		}
		rel := lines[0]
		if filepath.IsAbs(rel) || !strings.HasPrefix(rel, ".moai/reports/codex-audit/") {
			t.Fatalf("LAUNCH_RECORD path %q is not relative under the record directory", rel)
		}
		if !regexp.MustCompile(`^\.moai/reports/codex-audit/plan-auditor-[0-9]{8}T[0-9]{6}Z-[0-9a-f]{8}\.launch\.json$`).MatchString(rel) {
			t.Fatalf("record name %q does not follow <role>-<UTC>-<8hex>.launch.json", rel)
		}
		raw, err := os.ReadFile(filepath.Join(repo.a1, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		var rec map[string]any
		if err := json.Unmarshal(raw, &rec); err != nil {
			t.Fatalf("record is not JSON: %v", err)
		}
		return rel, rec, raw
	}
	checkCommon := func(t *testing.T, rec map[string]any, raw []byte) {
		t.Helper()
		if rec["schema_version"] != float64(1) || rec["role"] != "plan-auditor" || rec["route"] != "direct" {
			t.Errorf("header fields: %v %v %v", rec["schema_version"], rec["role"], rec["route"])
		}
		for _, k := range []string{"started_at", "ended_at"} {
			s, _ := rec[k].(string)
			ts, err := time.Parse(time.RFC3339, s)
			if err != nil || !strings.HasSuffix(s, "Z") || ts.Location() != time.UTC {
				t.Errorf("%s = %q is not RFC 3339 UTC", k, s)
			}
		}
		if rec["sandbox"] != "read-only" || rec["mcp_servers"] != "disabled" {
			t.Errorf("sandbox=%v mcp_servers=%v", rec["sandbox"], rec["mcp_servers"])
		}
		if s, _ := rec["covers"].(string); strings.TrimSpace(s) == "" {
			t.Error("covers is empty")
		}
		var uns []string
		for _, v := range rec["unsupported"].([]any) {
			uns = append(uns, v.(string))
		}
		sort.Strings(uns)
		if fmt.Sprint(uns) != "[codex-home-session-files project-hook-commands]" {
			t.Errorf("unsupported = %v", uns)
		}
		argv, _ := rec["argv"].([]any)
		redacted := 0
		want := fmt.Sprintf("developer_instructions=<redacted sha256=%s bytes=%d>", hex.EncodeToString(instrSum[:]), len(instr))
		for _, a := range argv {
			s := a.(string)
			if strings.HasPrefix(s, "developer_instructions=") {
				if s != want {
					t.Errorf("argv instruction token = %q, want %q", s, want)
				}
				redacted++
			}
		}
		if redacted != 1 || len(argv) == 0 || argv[0] != "exec" {
			t.Errorf("argv = %v", argv)
		}
		probe := instr[len(instr)/2 : len(instr)/2+120]
		if bytes.Contains(raw, []byte(probe)) {
			t.Error("record carries instruction text")
		}
	}

	t.Run("success", func(t *testing.T) {
		fake.setExec("PASS\n", 0)
		dest := filepath.Join(repo.a1, ".moai", "reports", "rec", "v.md")
		r := runAudit(t, codexAuditRequest{Role: "plan-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1, Out: dest})
		if r.res.ExitCode != 0 {
			t.Fatalf("launch failed: %s", r.stderr)
		}
		_, rec, raw := readRecord(t, r)
		checkCommon(t, rec, raw)
		if rec["exit_code"] != float64(0) || rec["failure_reason"] != nil {
			t.Errorf("exit_code=%v failure_reason=%v", rec["exit_code"], rec["failure_reason"])
		}
		if rec["verdict_path"] != ".moai/reports/rec/v.md" || rec["verdict_sha256"] != auditFileSHA(t, dest) {
			t.Errorf("verdict_path=%v verdict_sha256=%v", rec["verdict_path"], rec["verdict_sha256"])
		}
	})

	t.Run("failure", func(t *testing.T) {
		fake.setExec("half", 2)
		dest := filepath.Join(repo.a1, ".moai", "reports", "rec", "fail.md")
		r := runAudit(t, codexAuditRequest{Role: "plan-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1, Out: dest})
		if r.res.ExitCode == 0 {
			t.Fatal("failing audit reported success")
		}
		_, rec, raw := readRecord(t, r)
		checkCommon(t, rec, raw)
		if code, _ := rec["exit_code"].(float64); code == 0 {
			t.Errorf("exit_code = %v", rec["exit_code"])
		}
		if s, _ := rec["failure_reason"].(string); s == "" {
			t.Error("failure_reason empty")
		}
		if rec["verdict_path"] != nil || rec["verdict_sha256"] != nil {
			t.Errorf("failure record names a verdict: %v %v", rec["verdict_path"], rec["verdict_sha256"])
		}
	})

	t.Run("exclusive create", func(t *testing.T) {
		fake.setExec("PASS\n", 0)
		origName := codexAuditRecordSuffix
		codexAuditRecordSuffix = func() string { return "deadbeef" }
		origNow := codexAuditNow
		fixed := time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)
		codexAuditNow = func() time.Time { return fixed }
		t.Cleanup(func() { codexAuditRecordSuffix, codexAuditNow = origName, origNow })
		taken := filepath.Join(recordDir, "plan-auditor-20300102T030405Z-deadbeef.launch.json")
		if err := os.WriteFile(taken, []byte("keep"), 0o644); err != nil {
			t.Fatal(err)
		}
		execBefore := len(fake.execCalls(t))
		r := runAudit(t, codexAuditRequest{Role: "plan-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1})
		if r.res.ExitCode == 0 {
			t.Fatal("launch succeeded although the record name was taken")
		}
		if b, _ := os.ReadFile(taken); string(b) != "keep" {
			t.Fatalf("existing record overwritten: %q", b)
		}
		if len(fake.execCalls(t)) != execBefore {
			t.Fatal("exec started although no record could be reserved")
		}
	})
}

// ─── verb wiring ───────────────────────────────────────────────────────────

func TestCodexAuditVerbRegistered(t *testing.T) {
	var found bool
	for _, c := range codexCmd.Commands() {
		if c.Name() == "audit" {
			found = true
			if c.Flags().Lookup("out") == nil {
				t.Error("audit verb has no --out flag")
			}
		}
	}
	if !found {
		t.Fatal("moai codex audit is not registered")
	}
}

func TestCodexAuditVerbRunsInCallerWorktree(t *testing.T) {
	repo := newAuditRepo(t)
	fake := installFakeCodex(t)
	fake.setExec("verb verdict\n", 0)
	t.Chdir(repo.a1)
	var out, errb bytes.Buffer
	cmd := newCodexAuditCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&errb)
	cmd.SetIn(strings.NewReader("verb task"))
	cmd.SetArgs([]string{"sync-auditor", "--out", ".moai/reports/verb/v.md"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("verb failed: %v\n%s", err, errb.String())
	}
	got, err := os.ReadFile(filepath.Join(repo.a1, ".moai", "reports", "verb", "v.md"))
	if err != nil || string(got) != "verb verdict\n" {
		t.Fatalf("verb verdict = %q (%v)", got, err)
	}
	ex := fake.execCalls(t)
	if len(ex) != 1 || replaceAfter(ex[0], "-C", "X")[0] != "exec" {
		t.Fatalf("exec calls = %v", ex)
	}
	lines := launchRecordLines(errb.String())
	if len(lines) != 1 {
		t.Fatalf("verb printed %d LAUNCH_RECORD lines", len(lines))
	}
	raw, err := os.ReadFile(filepath.Join(repo.a1, filepath.FromSlash(lines[0])))
	if err != nil {
		t.Fatal(err)
	}
	var rec map[string]any
	_ = json.Unmarshal(raw, &rec)
	if rec["route"] != "shell" {
		t.Errorf("verb route = %v, want shell", rec["route"])
	}
}
