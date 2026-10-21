package cli

// M3 acceptance tests for the codex-review-entry gate-ON shell fast path
// (SPEC-HOOK-SYNC-ASYNC-001, plan §F M3). The wrapper
// .claude/hooks/moai/handle-codex-review-entry.sh gains a shell-layer receipt
// probe that resolves the common ALLOW states without the synchronous moai
// cold start, and fires the same marker-guarded background kick the Stop gate
// uses on cache-MISS states. These tests pin:
//
//   - the key parity (AC-HSA-008): the shell key computation must reproduce
//     verify.Key byte-for-byte, on both the `shasum` and `openssl` tool paths;
//   - the golden runtime-config prefix lists (plan M3 4a): the shell list is
//     byte-pinned to the Go constants so the probe can only narrow, never
//     widen, the nothing-reviewable detector's answer;
//   - the behavioral probe matrix (plan §F M3 list): gate off / no receipt /
//     stale receipt / unparseable snapshot / fresh pass / fresh inconclusive /
//     fresh FAIL / clean tree / config-only-only tree / behind binary /
//     HEAD-match binary / no-commit-metadata binary / missing git / missing
//     shasum / WT- branch / tree_scope: skip / primary-checkout session —
//     each asserting exit 0 plus the expected kick-or-fall-through observed
//     through an argv-recording stub that distinguishes `hook
//     codex-review-entry` (the synchronous handler, forbidden on ALLOW
//     states) from `verify codex-review` (the permitted background kick);
//   - the single-kick-per-bucket marker contract (AC-HSA-015);
//   - the live/template mirror byte parity (AC-HSA-006: the probe exists in
//     BOTH wrapper surfaces).
//
// The fixtures run the REAL wrapper under bash against a real git fixture
// tree with a REAL receipt written through the domain's own producer
// (verify.RecordReceipt), so the shell is judged against the artifacts the
// Go side actually writes.

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
	"github.com/modu-ai/moai-adk/internal/verify"
)

const (
	codexEntryWrapperRel  = "../../.claude/hooks/moai/handle-codex-review-entry.sh"
	codexEntryTemplateRel = "../../internal/template/templates/.claude/hooks/moai/handle-codex-review-entry.sh"

	// Stub-version placeholders newCodexEntryFixture resolves: "@PREV@" names
	// the first commit's short SHA (an ancestor of HEAD, not HEAD) and
	// "@HEAD@" the current HEAD's — the behind-binary and HEAD-match
	// fixtures of the M3 matrix.
	codexEntryPrevPlaceholder = "@PREV@"
	codexEntryHeadPlaceholder = "@HEAD@"
)

// codexEntryFullTools is the D3 dependency set the fast-path-eligible
// fixtures put on PATH. The wrapper's PATH is masked to exactly this set, so
// a probe removing an entry is the "missing <tool>" fixture of the M3 matrix.
var codexEntryFullTools = []string{"git", "awk", "shasum", "mkdir", "rm", "cat", "readlink"}

type codexEntryFixtureOpts struct {
	enabled      bool   // workflow.codex.review_gate.enabled
	treeScope    string // workflow.codex.review_gate.tree_scope ("" = key absent)
	primaryScope string // workflow.codex.review_gate.primary_scope ("" = key absent)
	branch       string // fixture branch name (WT- prefix ⇒ card scope)
	dirty        bool   // uncommitted tracked modification (reviewable change)
	configOnly   bool   // the ONLY change is .moai/config/ (config-only turn)
	twoCommits   bool   // base + second commit, for the behind-binary fixture
	stubVersion  string // what the stub `moai version` prints ("@PREV@"/"@HEAD@" resolved)
	stubTools    []string
	receipt      *codexEntryReceiptSpec // nil = no snapshot store entry
	receiptRaw   string                 // raw snapshot JSON body (overrides receipt; the R6 unbound-fields shape)
	logFileDir   bool                   // plant .moai/logs/codex-review-bg.log as a DIRECTORY (the R4 unopenable-log shape)
	logFileFifo  bool                   // plant the log path as a FIFO (the C1 foreground-block shape)
	corruptSnap  bool                   // overwrite the snapshot with garbage
}

// codexEntryReceiptSpec writes the receipt through the domain's own producer.
// cfgDigest/toolVersion default to the values the shell recomputes today (the
// real ConfigDigest over the tree-scope state map and the real `codex
// --version` output); the mismatch fixtures override exactly one of them.
type codexEntryReceiptSpec struct {
	verdict     string
	recordedAt  time.Time
	cfgDigest   string
	toolVersion string
}

type codexEntryFixture struct {
	dir       string
	toolsDir  string
	headShort string
	headFull  string
	prevShort string // first commit when twoCommits, else the same commit
}

func codexEntryFixtureGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, stderr.String())
	}
	return strings.TrimSpace(string(out))
}

func newCodexEntryFixture(t *testing.T, opts codexEntryFixtureOpts) *codexEntryFixture {
	t.Helper()

	for _, b := range []string{"bash", "git", "awk"} {
		if _, err := exec.LookPath(b); err != nil {
			t.Skipf("%s not available on this machine", b)
		}
	}

	dir := t.TempDir()

	// The stub binary, its argv record, the masked tools dir, and the
	// receipt/kick state live inside the fixture project; the committed
	// .gitignore keeps them out of every porcelain/ls-files read so they
	// never perturb the tree key — the .moai lines mirror every real MoAI
	// project (the repo's own .gitignore carries .moai/logs/, .moai/state/),
	// without them a written receipt would move the key it was written under.
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("/bin/\n/tools/\n/moai.argv\n/.moai/state/\n/.moai/logs/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	enabled := "true"
	if !opts.enabled {
		enabled = "false"
	}
	yaml := "# fixture workflow config\nworkflow:\n  codex:\n    review_gate:\n      enabled: " + enabled + "\n"
	if opts.treeScope != "" {
		yaml += "      tree_scope: " + opts.treeScope + "\n"
	}
	if opts.primaryScope != "" {
		yaml += "      primary_scope: " + opts.primaryScope + "\n"
	}
	if err := os.WriteFile(filepath.Join(dir, ".moai", "config", "sections", "workflow.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	branch := opts.branch
	if branch == "" {
		branch = "main"
	}
	codexEntryFixtureGit(t, dir, "init", "-b", branch)
	codexEntryFixtureGit(t, dir, "add", ".")
	codexEntryFixtureGit(t, dir, "-c", "user.email=t@fixture", "-c", "user.name=t", "commit", "-m", "base")
	f := &codexEntryFixture{dir: dir}
	f.headFull = codexEntryFixtureGit(t, dir, "rev-parse", "HEAD")
	f.headShort = codexEntryFixtureGit(t, dir, "rev-parse", "--short", "HEAD")
	f.prevShort = f.headShort

	if opts.twoCommits {
		if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("second\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		codexEntryFixtureGit(t, dir, "add", ".")
		codexEntryFixtureGit(t, dir, "-c", "user.email=t@fixture", "-c", "user.name=t", "commit", "-m", "second")
		f.headFull = codexEntryFixtureGit(t, dir, "rev-parse", "HEAD")
		f.headShort = codexEntryFixtureGit(t, dir, "rev-parse", "--short", "HEAD")
	}

	switch {
	case opts.configOnly:
		// The tracked runtime-config file is the ONLY change: the tree-scope
		// config-only narrowing must suppress the kick.
		if err := os.WriteFile(filepath.Join(dir, ".moai", "config", "sections", "workflow.yaml"), []byte(yaml+"# local edit\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	case opts.dirty:
		if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("edited\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// The argv-recording stub moai (the E-L07 fixture pattern): every
	// invocation appends its joined argv, so the probe distinguishes the
	// synchronous handler call (`hook codex-review-entry`, forbidden on
	// ALLOW states post-M3) from the background kick (`verify codex-review`,
	// expected there).
	stubVersion := opts.stubVersion
	stubVersion = strings.ReplaceAll(stubVersion, codexEntryPrevPlaceholder, f.prevShort)
	stubVersion = strings.ReplaceAll(stubVersion, codexEntryHeadPlaceholder, f.headShort)
	// The argv record path is baked in ABSOLUTE: the stub runs under the
	// wrapper's masked PATH, where dirname/coreutils may not exist.
	stub := "#!/bin/bash\nprintf '%s\\n' \"$*\" >> " + codexEntryQuoteShell(filepath.Join(dir, "moai.argv")) + "\n[ \"$1\" = \"version\" ] && printf '%s\\n' " + codexEntryQuoteShell(stubVersion) + "\nexit 0\n"
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bin", "moai"), []byte(stub), 0o755); err != nil {
		t.Fatal(err)
	}

	f.toolsDir = filepath.Join(dir, "tools")
	if err := os.MkdirAll(f.toolsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	tools := opts.stubTools
	if tools == nil {
		tools = codexEntryFullTools
	}
	if opts.receipt != nil || opts.receiptRaw != "" {
		// The R6 binding recomputation resolves the reviewer the way the Go
		// entry does; the masked PATH must carry it for receipt-present runs.
		tools = append(append([]string{}, tools...), "codex")
	}
	for _, name := range tools {
		real, err := exec.LookPath(name)
		if err != nil {
			t.Skipf("%s not available on this machine", name)
		}
		if err := os.Symlink(real, filepath.Join(f.toolsDir, name)); err != nil {
			t.Fatal(err)
		}
	}

	if opts.logFileDir {
		// The R4 unopenable-log shape: the background kick's log path exists
		// as a DIRECTORY, so the `>>` redirect dies before exec and strands
		// the marker unless the wrapper routes the failure through the
		// /dev/null arm.
		if err := os.MkdirAll(filepath.Join(dir, ".moai", "logs", "codex-review-bg.log"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if opts.logFileFifo {
		// The C1 foreground-block shape: a FIFO at the log path — opening it
		// for write blocks until a reader appears. The wrapper must reject
		// non-regular log paths before the probe.
		if err := os.MkdirAll(filepath.Join(dir, ".moai", "logs"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := exec.Command("mkfifo", filepath.Join(dir, ".moai", "logs", "codex-review-bg.log")).Run(); err != nil {
			t.Skipf("mkfifo unavailable: %v", err)
		}
	}
	if opts.receipt != nil || opts.receiptRaw != "" || opts.corruptSnap {
		key, err := verify.Key(context.Background(), dir)
		if err != nil {
			t.Fatalf("verify.Key over the fixture: %v", err)
		}
		if opts.receipt != nil {
			// The bound fields default to what the wrapper recomputes today:
			// the real ConfigDigest over the tree-scope state map (golden-
			// pinned by TestCodexEntryFastpath_GoldenConfigDigest) and the
			// real `codex --version` output (the codexVersionProbe mirror).
			cfgDigest := opts.receipt.cfgDigest
			if cfgDigest == "" {
				cfgDigest = verify.ConfigDigest(codexEntryStateConfigMap())
			}
			toolVersion := opts.receipt.toolVersion
			if toolVersion == "" {
				toolVersion = codexEntryRealToolVersion(t)
			}
			head, digest, _ := strings.Cut(key, ":")
			err = verify.RecordReceipt(dir, verify.Receipt{
				CheckID:      "codex-review",
				Head:         head,
				TreeDigest:   digest,
				ConfigDigest: cfgDigest,
				Command:      codexwiring.CodexReviewReceiptCommand,
				ToolVersion:  toolVersion,
				ExitCode:     0,
				Verdict:      opts.receipt.verdict,
				RecordedAt:   opts.receipt.recordedAt,
			})
			if err != nil {
				t.Fatalf("verify.RecordReceipt: %v", err)
			}
		}
		if opts.receiptRaw != "" {
			// PLACEHOLDER stands in for the fixture's real key — the F7
			// key-mismatch case writes a divergent literal instead and stays
			// unreplaced.
			body := strings.ReplaceAll(opts.receiptRaw, "PLACEHOLDER", key)
			if err := os.WriteFile(verify.SnapshotPath(dir, key), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if opts.corruptSnap {
			if err := os.WriteFile(verify.SnapshotPath(dir, key), []byte("}}} not json {{{"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return f
}

// codexEntryStateConfigMap is the tree-scope receipt binding's config map,
// byte-identical to reviewReceiptStateFrom's tree arm — the single source the
// fixture receipts, the golden test, and the wrapper's recomputation share.
func codexEntryStateConfigMap() map[string]string {
	return map[string]string{
		"gate":   codexReviewCheckID,
		"method": codexMethodReviewStart,
		"target": codexTargetUncommitted,
	}
}

// codexEntryRealToolVersion resolves the reviewer version exactly as the
// receipt producer does (codexVersionProbe: TrimSpace of `codex --version`).
func codexEntryRealToolVersion(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("codex"); err != nil {
		t.Skip("codex not available on this machine — the receipt-binding cases need the reviewer")
	}
	out, err := exec.Command("codex", "--version").Output()
	if err != nil {
		t.Skipf("codex --version failed: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func codexEntryQuoteShell(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\'`+"'") + "'"
}

type codexEntryRun struct {
	stdout   string
	stderr   string
	exitCode int
}

func runCodexEntryWrapper(t *testing.T, f *codexEntryFixture) codexEntryRun {
	t.Helper()
	return runCodexEntryWrapperStdin(t, f, fmt.Sprintf(`{"cwd": %q}`, f.dir))
}

func runCodexEntryWrapperStdin(t *testing.T, f *codexEntryFixture, stdin string) codexEntryRun {
	t.Helper()
	wrapper, err := filepath.Abs(codexEntryWrapperRel)
	if err != nil {
		t.Fatal(err)
	}
	// The hook runs on a 30s UserPromptSubmit budget; a wrapper that blocks
	// past this bound is a defect (the C1 shape) and must fail the test, not
	// hang the suite.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", wrapper)
	// A blocking child (the C1 FIFO shape) survives the context kill and
	// holds the inherited stdout/stderr pipes open; WaitDelay bounds Wait so
	// the deadline is actually observed.
	cmd.WaitDelay = 3 * time.Second
	cmd.Dir = f.dir
	cmd.Env = []string{
		"CLAUDE_PROJECT_DIR=" + f.dir,
		"PATH=" + f.toolsDir,
		"HOME=" + f.dir,
	}
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("the wrapper blocked past the 30s hook budget (foreground blocking probe?)\nstderr:\n%s", stderr.String())
	}
	code := 0
	if runErr != nil {
		ee, ok := runErr.(*exec.ExitError)
		if !ok {
			t.Fatalf("running the wrapper: %v\nstderr:\n%s", runErr, stderr.String())
		}
		code = ee.ExitCode()
	}
	return codexEntryRun{stdout: stdout.String(), stderr: stderr.String(), exitCode: code}
}

// codexEntryArgvCounts reads the stub's argv record and counts the two
// invocation classes the matrix distinguishes.
func codexEntryArgvCounts(t *testing.T, f *codexEntryFixture) (forward, kick int, argv string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(f.dir, "moai.argv"))
	if err != nil {
		return 0, 0, ""
	}
	argv = string(b)
	for _, line := range strings.Split(argv, "\n") {
		switch {
		case line == "hook codex-review-entry":
			forward++
		case strings.HasPrefix(line, "verify codex-review"):
			kick++
		}
	}
	return forward, kick, argv
}

// codexEntryAwaitArgv waits out the detached-kick race: the background review
// child is async by design and may append its argv line after the wrapper
// (and this test's exec) has returned. For expected invocations it polls up
// to 2s for the counts to be reached; for a want-0 state it gives the child a
// fixed settle window before the caller asserts absence.
func codexEntryAwaitArgv(t *testing.T, f *codexEntryFixture, wantFwd, wantKick int) (int, int, string) {
	t.Helper()
	if wantFwd == 0 && wantKick == 0 {
		time.Sleep(500 * time.Millisecond)
		return codexEntryArgvCounts(t, f)
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		fwd, kick, argv := codexEntryArgvCounts(t, f)
		if fwd >= wantFwd && kick >= wantKick {
			return fwd, kick, argv
		}
		if time.Now().After(deadline) {
			return fwd, kick, argv
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// TestCodexEntryFastpath_KeyParity pins AC-HSA-008: the shell key computation
// reproduces verify.Key byte-for-byte, on both sha256 tool paths (E6).
func TestCodexEntryFastpath_KeyParity(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, content string) {
		t.Helper()
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	codexEntryFixtureGit(t, dir, "init", "-b", "main")
	write(".gitignore", "/ignored-dir/\n")
	codexEntryFixtureGit(t, dir, "add", ".")
	codexEntryFixtureGit(t, dir, "-c", "user.email=t@fixture", "-c", "user.name=t", "commit", "-m", "base")

	// A rich untracked set: plain, nested, empty, binary (NUL bytes), and an
	// ignored path that must NOT enter the hash. Symlinks are NOT here: the
	// R3 repair declines them (the readlink byte semantics are not portable),
	// and they carry their own decline test below.
	write("u.txt", "untracked\n")
	write("sub/nested.txt", "nested\n")
	write("empty.txt", "")
	write("blob.bin", "bin\x00ary\xff bytes")
	write("ignored-dir/x.txt", "must not enter the hash\n")
	// A dirty tracked file exercises the porcelain + diff legs.
	write("a.txt", "edited\n")

	goKey, err := verify.Key(context.Background(), dir)
	if err != nil {
		t.Fatalf("verify.Key: %v", err)
	}
	goHead, goDigest, _ := strings.Cut(goKey, ":")

	wrapper, err := filepath.Abs(codexEntryWrapperRel)
	if err != nil {
		t.Fatal(err)
	}
	// Sourcing: CODEX_ENTRY_SOURCE_ONLY=1 makes the wrapper yield right after
	// the helper definitions (its normal body exits — an executed hook always
	// terminates), so the caller can invoke codex_entry_key directly.
	gateHome := t.TempDir()
	shellKey := func(shaPipeline string) string {
		t.Helper()
		script := `export CODEX_ENTRY_SOURCE_ONLY=1
. "$1" </dev/null
codex_entry_key "$2" ` + shaPipeline + ` | codex_entry_sha_extract`
		cmd := exec.Command("bash", "-c", script, "bash", wrapper, dir)
		cmd.Env = []string{"CLAUDE_PROJECT_DIR=" + gateHome, "PATH=" + os.Getenv("PATH"), "HOME=" + gateHome}
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("shell key computation failed: %v", err)
		}
		return strings.TrimSpace(string(out))
	}

	for _, tool := range []struct {
		name     string
		pipeline string
	}{
		{"shasum", "| shasum -a 256"},
		{"openssl", "| openssl dgst -sha256"},
	} {
		hex := shellKey(tool.pipeline)
		if len(hex) != 64 {
			t.Fatalf("%s path: shell digest %q is not 64 hex chars", tool.name, hex)
		}
		if hex[:16] != goDigest {
			t.Fatalf("%s path: key digest mismatch: shell %q vs go %q (full go key %q)", tool.name, hex[:16], goDigest, goKey)
		}
	}
	if goHead == "" {
		t.Fatal("go key carried an empty head")
	}
	if want := codexEntryFixtureGit(t, dir, "rev-parse", "HEAD"); goHead != want {
		t.Fatalf("go key head %q != rev-parse HEAD %q", goHead, want)
	}
}

// TestCodexEntryFastpath_KeyDeclinesUnreproducibleEntries pins the R2/R3
// repairs: the shell key loop must DECLINE (non-zero) on entries whose
// verify.Key legs it cannot reproduce byte-exactly — untracked directory
// entries (an embedded repo is listed as the directory itself; WalkDir hashes
// every descendant after the 'd' marker) and symlink entries (the
// os.Readlink bytes are not portably readable). A near-miss key silently
// kills the kick arm (plan §G) — declining hands the tree to Go instead.
func TestCodexEntryFastpath_KeyDeclinesUnreproducibleEntries(t *testing.T) {
	baseTree := func(t *testing.T) string {
		t.Helper()
		dir := t.TempDir()
		codexEntryFixtureGit(t, dir, "init", "-b", "main")
		if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("base\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		codexEntryFixtureGit(t, dir, "add", ".")
		codexEntryFixtureGit(t, dir, "-c", "user.email=t@fixture", "-c", "user.name=t", "commit", "-m", "base")
		return dir
	}
	assertDeclines := func(t *testing.T, dir, what string) {
		t.Helper()
		if _, err := verify.Key(context.Background(), dir); err != nil {
			t.Fatalf("Go Key over the %s fixture: %v", what, err)
		}
		wrapper, err := filepath.Abs(codexEntryWrapperRel)
		if err != nil {
			t.Fatal(err)
		}
		script := `export CODEX_ENTRY_SOURCE_ONLY=1
. "$1" </dev/null
codex_entry_key "$2" >/dev/null`
		cmd := exec.Command("bash", "-c", script, "bash", wrapper, dir)
		cmd.Env = []string{"CLAUDE_PROJECT_DIR=" + dir, "PATH=" + os.Getenv("PATH"), "HOME=" + dir}
		if err := cmd.Run(); err == nil {
			t.Fatalf("codex_entry_key succeeded on a tree with %s — a near-miss key would silently kill the kick arm; it must decline", what)
		}
	}

	t.Run("untracked directory entry (embedded repo)", func(t *testing.T) {
		dir := baseTree(t)
		if err := os.MkdirAll(filepath.Join(dir, "em"), 0o755); err != nil {
			t.Fatal(err)
		}
		codexEntryFixtureGit(t, dir, "init", "-q", "-b", "main", filepath.Join(dir, "em"))
		if err := os.WriteFile(filepath.Join(dir, "em", "f.txt"), []byte("inner\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		assertDeclines(t, dir, "an untracked directory")
	})

	t.Run("plain symlink entry", func(t *testing.T) {
		dir := baseTree(t)
		if err := os.Symlink("a.txt", filepath.Join(dir, "link.txt")); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		assertDeclines(t, dir, "a symlink")
	})

	t.Run("newline-terminated symlink target", func(t *testing.T) {
		dir := baseTree(t)
		if err := os.Symlink("a.txt\n", filepath.Join(dir, "nl-link")); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		assertDeclines(t, dir, "a newline-terminated symlink target")
	})
}

// TestCodexEntryFastpath_GoldenConfigDigest pins the R6 config-digest
// recomputation: the shell's length-prefixed emission must hash to exactly
// verify.ConfigDigest's output for the tree-scope state map (CheckReceipt's
// config_digest bound field).
func TestCodexEntryFastpath_GoldenConfigDigest(t *testing.T) {
	goDigest := verify.ConfigDigest(codexEntryStateConfigMap())
	if len(goDigest) != 16 {
		t.Fatalf("go config digest %q is not 16 hex chars", goDigest)
	}
	wrapper, err := filepath.Abs(codexEntryWrapperRel)
	if err != nil {
		t.Fatal(err)
	}
	script := `export CODEX_ENTRY_SOURCE_ONLY=1
. "$1" </dev/null
codex_entry_config_digest | shasum -a 256 | codex_entry_sha_extract`
	cmd := exec.Command("bash", "-c", script, "bash", wrapper)
	cmd.Env = []string{"CLAUDE_PROJECT_DIR=" + t.TempDir(), "PATH=" + os.Getenv("PATH")}
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("shell config-digest computation failed: %v", err)
	}
	hex := strings.TrimSpace(string(out))
	if len(hex) != 64 {
		t.Fatalf("shell config digest %q is not 64 hex chars", hex)
	}
	if hex[:16] != goDigest {
		t.Fatalf("config digest mismatch: shell %q vs go %q", hex[:16], goDigest)
	}
}

// TestCodexEntryFastpath_GoldenPrefixes pins the shell runtime-config prefix
// lists to the Go constants they mirror (plan M3 4a): the config-only probe
// can only narrow the nothing-reviewable detector's answer, never widen it.
func TestCodexEntryFastpath_GoldenPrefixes(t *testing.T) {
	src, err := os.ReadFile(codexEntryWrapperRel)
	if err != nil {
		t.Fatalf("reading the live wrapper: %v", err)
	}
	extract := func(varName string) []string {
		t.Helper()
		re := regexp.MustCompile(`(?m)^` + varName + `="([^"]*)"$`)
		m := re.FindSubmatch(src)
		if m == nil {
			t.Fatalf("the wrapper does not declare %s — the golden pin has nothing to bind", varName)
		}
		return strings.Fields(string(m[1]))
	}
	assertSameStrings := func(what string, got, want []string) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("%s: shell carries %d entries (%v), Go constants carry %d (%v)", what, len(got), got, len(want), want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("%s: shell entry %d = %q, Go constant = %q", what, i, got[i], want[i])
			}
		}
	}
	assertSameStrings("runtime prefixes", extract("CODEX_ENTRY_RUNTIME_PREFIXES"), reviewGateRuntimePrefixes)
	assertSameStrings("config exact paths", extract("CODEX_ENTRY_CONFIG_EXACT"), reviewGateTreeConfigExactPaths)
	assertSameStrings("config dir prefixes", extract("CODEX_ENTRY_CONFIG_DIR"), reviewGateTreeConfigDirPrefixes)
}

// TestCodexEntryFastpath_ProbeMatrix walks the plan §F M3 behavioral probe
// list. Every state asserts exit 0 (the entry hook NEVER blocks — REQ-RG-001)
// plus the expected kick-or-fall-through observation.
func TestCodexEntryFastpath_ProbeMatrix(t *testing.T) {
	now := time.Now()

	// handWrittenReceipt builds a raw snapshot body the way verify.Save's
	// MarshalIndent output is shaped; the PLACEHOLDER key is replaced with
	// the fixture's real key by newCodexEntryFixture. `toolVersion` and
	// `storedKey` are inserted verbatim (the C2/F7 fixtures need escapes and
	// divergent keys the producer would never write).
	handWrittenReceipt := func(t *testing.T, at time.Time, toolVersion, storedKey string) string {
		t.Helper()
		return `{
  "key": "` + storedKey + `",
  "recorded_at": "` + at.Format(time.RFC3339Nano) + `",
  "checks": [
    {
      "check_id": "codex-review",
      "command": "moai verify codex-review",
      "exit_code": 0,
      "recorded_at": "` + at.Format(time.RFC3339Nano) + `",
      "duration_ms": 0,
      "config_digest": "` + verify.ConfigDigest(codexEntryStateConfigMap()) + `",
      "tool_version": "` + toolVersion + `",
      "verdict": "pass"
    }
  ]
}
`
	}
	escapedToolReceipt := func(t *testing.T, at time.Time) string {
		t.Helper()
		// A FAIL receipt whose stored tool_version carries a JSON-escaped
		// newline (the C2 shape: a multi-line reviewer version); the exit
		// code 2 mirrors the producer's inconclusive-failure convention for
		// a recorded fail.
		body := handWrittenReceipt(t, at, `1.0\n2`, "PLACEHOLDER")
		return strings.Replace(body, `"exit_code": 0`, `"exit_code": 2`, 1)
	}
	truncatedReceipt := func(t *testing.T, at time.Time) string {
		t.Helper()
		body := handWrittenReceipt(t, at, codexEntryRealToolVersion(t), "PLACEHOLDER")
		// Cut the document right after the review entry's closing brace —
		// the F7 shape the round-1 awk parser accepted as a fresh pass.
		idx := strings.LastIndex(body, "  ]")
		if idx < 0 {
			t.Fatal("receipt body has no checks-array close")
		}
		return body[:idx]
	}
	mismatchedKeyReceipt := func(t *testing.T, at time.Time) string {
		t.Helper()
		return handWrittenReceipt(t, at, codexEntryRealToolVersion(t), "mismatchedtreehead000:0000000000000000")
	}
	// reviewableKickable is a miss-state reviewable fixture whose stub stamp
	// carries the HEAD commit (StatusFresh), so the binary-age precondition
	// proceeds and the expected kick is reached on every post-R5 path.
	reviewableKickable := func(mut func(*codexEntryFixtureOpts)) codexEntryFixtureOpts {
		opts := codexEntryFixtureOpts{enabled: true, primaryScope: "review", dirty: true,
			stubVersion: "moai-adk v3.2.0-1-g" + codexEntryHeadPlaceholder + "-dirty"}
		if mut != nil {
			mut(&opts)
		}
		return opts
	}
	reviewableMiss := reviewableKickable
	cases := []struct {
		name     string
		opts     codexEntryFixtureOpts
		wantKick int
		wantFwd  int
		stdin    string // override payload; "" = {"cwd": <fixture dir>}
	}{
		{
			name: "gate off exits 0 with no binary resolution",
			opts: codexEntryFixtureOpts{enabled: false},
		},
		{
			name:     "no receipt on a reviewable tree: ALLOW with one kick, no forward",
			opts:     reviewableKickable(nil),
			wantKick: 1,
		},
		{
			name: "stale receipt: miss arm, one kick",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.receipt = &codexEntryReceiptSpec{verdict: "pass", recordedAt: now.Add(-25 * time.Hour)}
			}),
			wantKick: 1,
		},
		{
			name: "unparseable snapshot: receipt-absent, miss arm, one kick",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.receipt = &codexEntryReceiptSpec{verdict: "pass", recordedAt: now}
				o.corruptSnap = true
			}),
			wantKick: 1,
		},
		{
			name: "fresh pass: exit 0, no kick, no forward",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.receipt = &codexEntryReceiptSpec{verdict: "pass", recordedAt: now}
			}),
		},
		{
			name: "fresh inconclusive: exit 0, no kick, no forward",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.receipt = &codexEntryReceiptSpec{verdict: "inconclusive", recordedAt: now}
			}),
		},
		{
			name: "fresh FAIL: fall through to the synchronous handler",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.receipt = &codexEntryReceiptSpec{verdict: "fail", recordedAt: now}
			}),
			wantFwd: 1,
		},
		{
			// R6: a receipt whose tool_version differs from the current
			// reviewer is NOT valid for this state — CheckReceipt's bound
			// field mismatch reads as miss, and the miss arm kicks.
			name: "tool_version mismatch: bound-field miss arm, one kick",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.receipt = &codexEntryReceiptSpec{verdict: "pass", recordedAt: now, toolVersion: "codex-cli 0.0.1-mismatch"}
			}),
			wantKick: 1,
		},
		{
			// R6: a config_digest mismatch is the same bound-field miss.
			name: "config_digest mismatch: bound-field miss arm, one kick",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.receipt = &codexEntryReceiptSpec{verdict: "pass", recordedAt: now, cfgDigest: "aaaaaaaaaaaaaaaa"}
			}),
			wantKick: 1,
		},
		{
			// R6: an entry written without config_digest/tool_version (the
			// omitempty shapes) is UNBOUND — never evidence — so the shell
			// falls through and lets CheckReceipt's own reading decide.
			name: "receipt without bound fields: fall through",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.receipt = &codexEntryReceiptSpec{verdict: "pass", recordedAt: now}
				o.receiptRaw = `{
  "key": "PLACEHOLDER",
  "recorded_at": "` + now.Format(time.RFC3339Nano) + `",
  "checks": [
    {
      "check_id": "codex-review",
      "command": "moai verify codex-review",
      "exit_code": 0,
      "recorded_at": "` + now.Format(time.RFC3339Nano) + `",
      "duration_ms": 0,
      "verdict": "pass"
    }
  ]
}`
			}),
			wantFwd: 1,
		},
		{
			name: "clean tree: nothing reviewable, no kick, no forward",
			opts: codexEntryFixtureOpts{enabled: true, primaryScope: "review", stubVersion: "v"},
		},
		{
			name: "config-only-only tree: no kick, no forward",
			opts: codexEntryFixtureOpts{enabled: true, primaryScope: "review", configOnly: true, stubVersion: "v"},
		},
		{
			name: "behind binary: decidable ancestor-hit suppresses the early kick",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.twoCommits = true
				o.stubVersion = "moai-adk v3.2.0-1-g" + codexEntryPrevPlaceholder + "-dirty"
			}),
		},
		{
			name:     "HEAD-match binary: StatusFresh, the kick proceeds",
			opts:     reviewableKickable(nil),
			wantKick: 1,
		},
		{
			// R5: without a -g commit component the shell cannot distinguish
			// "tag-only build" from "unparseable output" — it must fall
			// through and let Go's staleBinarySkipApplies own the leniency
			// with the real ldflags commit.
			name: "no commit component in the stamp: fall through, Go owns leniency",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.stubVersion = "moai-adk v3.2.0 (tag only, no -g component)"
			}),
			wantFwd: 1,
		},
		{
			// R1: the stdin session dir outranks CLAUDE_PROJECT_DIR
			// (reviewScopeSessionDir); when they disagree the shell's tree
			// premise is wrong and it must fall through.
			name:    "payload cwd diverges from CLAUDE_PROJECT_DIR: fall through",
			opts:    reviewableKickable(nil),
			wantFwd: 1,
			stdin:   `{"cwd": "/tmp/t1913-some-other-session-tree"}`,
		},
		{
			// R1's precedence mirror: an empty cwd falls back to project_dir.
			name:    "payload project_dir diverges: fall through",
			opts:    reviewableKickable(nil),
			wantFwd: 1,
			stdin:   `{"project_dir": "/tmp/t1913-another-session-tree"}`,
		},
		{
			// R4: an unopenable log path must not strand the marker without a
			// review behind it — the kick routes through the /dev/null arm.
			name: "unopenable log path: the kick still fires",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.logFileDir = true
			}),
			wantKick: 1,
		},
		{
			// C1: a FIFO at the log path blocks a foreground `>>` probe until
			// a reader appears (repro exit 124) — the wrapper must reject
			// non-regular log paths before the probe and still kick.
			name: "FIFO log path: no blocking, the kick still fires",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.logFileFifo = true
			}),
			wantKick: 1,
		},
		{
			// C2: a stored tool_version carrying JSON escapes (backslash in
			// the raw text — multi-line reviewer versions) cannot be decoded
			// byte-exactly in the shell; the binding must decline so the Go
			// handler — which JSON-decodes before comparing — delivers the
			// valid fresh FAIL instead of the shell mis-classing it as a
			// cache miss (delivery skipped, re-kick every bucket).
			name: "escaped tool_version: the binding declines, Go decides",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.receipt = &codexEntryReceiptSpec{verdict: "fail", recordedAt: now}
				o.receiptRaw = escapedToolReceipt(t, now)
			}),
			wantFwd: 1,
		},
		{
			// F7: a snapshot TRUNCATED right after the review entry's closing
			// brace is not a parsable JSON document — verify.Load would
			// return absent, so the shell must not read it as a fresh pass.
			name: "truncated snapshot: cache-miss, one kick",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.receipt = &codexEntryReceiptSpec{verdict: "pass", recordedAt: now}
				o.receiptRaw = truncatedReceipt(t, now)
			}),
			wantKick: 1,
		},
		{
			// F7's second leg: verify.Load treats a STORED-KEY MISMATCH as
			// absent (the filename-collision defense) — the shell must
			// compare the stored key against the key it computed.
			name: "stored key mismatch: cache-miss, one kick",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.receipt = &codexEntryReceiptSpec{verdict: "pass", recordedAt: now}
				o.receiptRaw = mismatchedKeyReceipt(t, now)
			}),
			wantKick: 1,
		},
		{
			name: "missing git: probe failure falls through",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.stubVersion = "v"
				o.stubTools = []string{"awk", "shasum", "mkdir", "rm", "cat", "readlink"}
			}),
			wantFwd: 1,
		},
		{
			name: "missing shasum: the openssl fallback still kicks",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.stubTools = []string{"git", "awk", "openssl", "mkdir", "rm", "cat", "readlink"}
			}),
			wantKick: 1,
		},
		{
			name: "both sha256 tools absent: fall through",
			opts: reviewableKickable(func(o *codexEntryFixtureOpts) {
				o.stubVersion = "v"
				o.stubTools = []string{"git", "awk", "mkdir", "rm", "cat", "readlink"}
			}),
			wantFwd: 1,
		},
		{
			name: "WT- branch: card scope reaches the Go handler once",
			opts: reviewableMiss(func(o *codexEntryFixtureOpts) {
				o.branch = "WT-fixture-card"
				o.stubVersion = "v"
			}),
			wantFwd: 1,
		},
		{
			name: "tree_scope skip config: fall through to the Go handler",
			opts: reviewableMiss(func(o *codexEntryFixtureOpts) {
				o.treeScope = "skip"
				o.stubVersion = "v"
			}),
			wantFwd: 1,
		},
		{
			name: "primary-checkout session under the skip default: fall through",
			opts: reviewableMiss(func(o *codexEntryFixtureOpts) {
				o.primaryScope = "" // the skip default; the fixture IS the primary working tree
				o.stubVersion = "v"
			}),
			wantFwd: 1,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			f := newCodexEntryFixture(t, tc.opts)
			var run codexEntryRun
			if tc.stdin != "" {
				run = runCodexEntryWrapperStdin(t, f, tc.stdin)
			} else {
				run = runCodexEntryWrapper(t, f)
			}
			if run.exitCode != 0 {
				t.Fatalf("exit = %d, want 0 (the entry hook never blocks)\nstderr:\n%s", run.exitCode, run.stderr)
			}
			fwd, kick, argv := codexEntryAwaitArgv(t, f, tc.wantFwd, tc.wantKick)
			if fwd != tc.wantFwd || kick != tc.wantKick {
				t.Fatalf("argv counts (forward=%d kick=%d) != want (forward=%d kick=%d)\nargv:\n%s\nstderr:\n%s", fwd, kick, tc.wantFwd, tc.wantKick, argv, run.stderr)
			}
		})
	}
}

// TestCodexEntryFastpath_SingleKickPerBucket pins AC-HSA-015: the entry-side
// kick honors the shared kick marker — two entries (or an entry after the
// Stop arm's kick) over the same tree state in the same bucket kick exactly
// one background review; and a fresh-pass receipt produces NO kick at all.
func TestCodexEntryFastpath_SingleKickPerBucket(t *testing.T) {
	base := func(mut func(*codexEntryFixtureOpts)) codexEntryFixtureOpts {
		opts := codexEntryFixtureOpts{enabled: true, primaryScope: "review", dirty: true,
			stubVersion: "moai-adk v3.2.0-1-g" + codexEntryHeadPlaceholder + "-dirty"}
		if mut != nil {
			mut(&opts)
		}
		return opts
	}

	t.Run("two entries in one bucket kick once", func(t *testing.T) {
		f := newCodexEntryFixture(t, base(nil))
		if run := runCodexEntryWrapper(t, f); run.exitCode != 0 {
			t.Fatalf("first run exit = %d\nstderr:\n%s", run.exitCode, run.stderr)
		}
		if run := runCodexEntryWrapper(t, f); run.exitCode != 0 {
			t.Fatalf("second run exit = %d\nstderr:\n%s", run.exitCode, run.stderr)
		}
		_, kick, argv := codexEntryAwaitArgv(t, f, 0, 1)
		if kick != 1 {
			t.Fatalf("kick count = %d, want exactly 1 across both entries\nargv:\n%s", kick, argv)
		}
	})

	t.Run("a pre-existing Stop-arm marker suppresses the entry kick", func(t *testing.T) {
		f := newCodexEntryFixture(t, base(nil))
		key, err := verify.Key(context.Background(), f.dir)
		if err != nil {
			t.Fatal(err)
		}
		head, digest, _ := strings.Cut(key, ":")
		bucket := time.Now().Unix() / 900
		marker := filepath.Join(f.dir, ".moai", "state", "verify", "codex-review",
			head+"-"+digest+".kick-"+strconv.FormatInt(bucket, 10))
		if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(marker, []byte("stop-arm"), 0o644); err != nil {
			t.Fatal(err)
		}
		if run := runCodexEntryWrapper(t, f); run.exitCode != 0 {
			t.Fatalf("exit = %d\nstderr:\n%s", run.exitCode, run.stderr)
		}
		_, kick, argv := codexEntryAwaitArgv(t, f, 0, 0)
		if kick != 0 {
			t.Fatalf("kick count = %d, want 0 with the Stop arm's marker in flight\nargv:\n%s", kick, argv)
		}
	})

	t.Run("fresh pass kicks nothing", func(t *testing.T) {
		f := newCodexEntryFixture(t, base(func(o *codexEntryFixtureOpts) {
			o.receipt = &codexEntryReceiptSpec{verdict: "pass", recordedAt: time.Now()}
		}))
		if run := runCodexEntryWrapper(t, f); run.exitCode != 0 {
			t.Fatalf("exit = %d\nstderr:\n%s", run.exitCode, run.stderr)
		}
		fwd, kick, argv := codexEntryAwaitArgv(t, f, 0, 0)
		if kick != 0 || fwd != 0 {
			t.Fatalf("fresh pass produced forward=%d kick=%d, want 0/0\nargv:\n%s", fwd, kick, argv)
		}
	})
}

// TestCodexEntryFastpath_BothSurfacesCarryProbe pins AC-HSA-006's both-
// surfaces clause: the shell receipt probe exists in the live wrapper AND its
// template mirror, and the two surfaces carry the SAME bytes (the Template-
// First pair, C2).
func TestCodexEntryFastpath_BothSurfacesCarryProbe(t *testing.T) {
	live, err := os.ReadFile(codexEntryWrapperRel)
	if err != nil {
		t.Fatalf("reading the live wrapper: %v", err)
	}
	tmpl, err := os.ReadFile(codexEntryTemplateRel)
	if err != nil {
		t.Fatalf("reading the template wrapper: %v", err)
	}
	for _, probe := range []string{"snapshots", "codex_entry_key", "codex_entry_fastpath"} {
		if !strings.Contains(string(live), probe) {
			t.Fatalf("the live wrapper carries no %q — the shell receipt probe is missing", probe)
		}
		if !strings.Contains(string(tmpl), probe) {
			t.Fatalf("the template wrapper carries no %q — the shell receipt probe is missing from the mirror", probe)
		}
	}
	if !bytes.Equal(live, tmpl) {
		t.Fatalf("live and template wrappers diverged: live %d bytes vs template %d bytes (Template-First pair must be byte-identical)", len(live), len(tmpl))
	}
}
