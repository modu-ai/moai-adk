# SPEC-FALSE-ALLCLEAR-GUARD-001 — Acceptance Criteria

> **Revision v0.2.0** — rewritten after plan-audit iteration 1 (FAIL 0.66, Testability 0.45). Ten criteria failed to discriminate. Every criterion below now carries a **measured baseline**: the judging command was executed against the unmodified tree at `6763aff3b` and its observed output is recorded inline. A criterion whose command already produces the passing observation before any implementation exists is defective by construction — that is how the previous revision's vacuous criteria were produced.

## §A Judging conventions

Every criterion is a runnable command plus its expected observation, plus **what would have to break for it to fail** and the **baseline observed on the unmodified tree**.

### A.1 Mandatory command shapes

**Shape 1 — test selection.** `go test -run` takes a **non-anchored regex** and exits 0 when it matches nothing. Any criterion selecting tests by name MUST prove the test exists AND observe an actual `--- PASS:` line:

```bash
go test -list '^<TestName>$' ./<pkg>/ | grep -c '^<TestName>$'                              # expect: 1
go test -count=1 -v -run '^<TestName>$' ./<pkg>/ 2>&1 | grep -c '^--- PASS: <TestName>'      # expect: >= 1
```

Neither half alone suffices. Verified working: `go test -list '^TestQualityGate_Run_AstGrep' ./internal/hook/quality/ | grep -c ...` → **2**, matching `--- PASS:` count → **2**.

**Shape 2 — `sg`-absence construction (CORRECTED).** The previous revision used `env PATH="$EMPTY" command -v sg`. That is **inoperative**: `command` is a shell builtin, `env` cannot exec it, and it returns exit 127 identically whether or not `sg` exists. Measured on this host (where `sg` IS installed at `/opt/homebrew/bin/sg`):

```
$ env PATH="$EMPTY" command -v sg
env: command: No such file or directory      → exit 127   (INOPERATIVE — never use)
```

The obvious repair `env PATH="$EMPTY" sh -c 'command -v sg'` is **also broken** — `env` cannot find `sh` under a stripped `PATH`. Measured: `env: sh: No such file or directory`, exit 127. Recorded so it is not re-introduced as a "fix".

The canonical form uses an assignment prefix on the builtin, which execs nothing:

```bash
EMPTY=$(mktemp -d)
PATH="$EMPTY" command -v sg     # expect: no output, non-zero exit   ← sg removed
command -v sg                    # expect: a path, exit 0            ← positive control
```

**Measured discrimination proof** (this host, unmodified tree): stripped → no output, `exit=1`; normal → `/opt/homebrew/bin/sg`, `exit=0`. The prefix assignment is temporary — PATH restores afterwards (verified). **Both lines are mandatory**: the positive control is what proves the strip is the cause. On a host where `sg` is genuinely absent the positive control fails and the criterion is reported **NOT RUNNABLE**, never as a pass.

Binaries under test are invoked by **absolute path**, so `env PATH="$EMPTY" "$BIN" ...` still execs correctly (verified: `ast-grep` ran and produced `no findings`).

**Shape 3 — stream separation.** stdout and stderr captured to separate files, asserted independently. `2>&1` is prohibited in any criterion judging which stream carried a message.

**Shape 4 — exit code after a pipe.** `$?` after a pipeline reports the **last** stage. Measured: `sh -c 'exit 3' 2>&1 | tail -5; echo $?` → **0**. Redirect first, read `$?`, then tail:

```bash
<cmd> > /tmp/<slug>.log 2>&1; echo "exit=$?"; tail -30 /tmp/<slug>.log
```

Measured on the corrected form: `sh -c 'exit 3' > /tmp/f.log 2>&1; echo $?` → **3**.

**Shape 5 — anchored diffs.** `git diff --stat -- <path>` with no ref compares working tree to index, so it is empty **by construction** once work is committed and can never detect a committed change. Every diff-based criterion names the base commit `6763aff3b` explicitly.

**Shape 6 — Go-test `sg`-absence lever.** `Scanner.isSGAvailable` calls `exec.LookPath`, which reads the process `PATH`. In a Go test the lever is `t.Setenv("PATH", t.TempDir())`. `t.Setenv` is incompatible with `t.Parallel()`, so every test using this lever is **non-parallel** — stated once here so it is not rediscovered per test.

**Shape 8 — isolated mutation round trip (N3).** Three criteria (AC-FAG-004, AC-FAG-006, AC-FAG-011) prove falsifiability by *removing* a change and observing the result flip. None of them may edit tracked source in place: this repository has 50+ live worktrees sharing one git dir, and an interrupted round trip would leave `root.go` / `pre_tool.go` / `gate.go` broken on the feature branch. Every round trip runs in a throwaway detached worktree:

```bash
PROBE=$(mktemp -d)/probe
git worktree add --detach "$PROBE" <ref>      # <ref> = 6763aff3b for AC-004; HEAD for AC-006 / AC-011
#   ... mutate and build/test INSIDE "$PROBE" only ...
git worktree remove --force "$PROBE"
```

`<ref>` differs by purpose: AC-FAG-004 compares against the **base commit** (`6763aff3b`) because it needs the pre-M1 binary; AC-FAG-006 and AC-FAG-011 branch from **HEAD** because they need the implementation *and* its new tests present before mutating them. `git worktree add` is the procedure `main-checkout-branch-guard.md` explicitly sanctions; `git stash` is prohibited there.

**Shape 7 — Go-test non-sentinel-error lever.** `ValidateBinary` (`internal/astgrep/scanner.go:139-173`) accepts a bare name only when it is `sg` or `ast-grep`, and an absolute path only under `trustedBinaryPrefixes()` = `/usr/bin/`, `/usr/local/bin/`, `/opt/homebrew/bin/`, `~/go/bin/`, `~/.local/bin/`, `~/.cargo/bin/`; anything else falls through to `return ErrUntrustedBinary`. A `t.TempDir()` path (`/var/folders/...` on macOS, `/tmp/...` on Linux) matches none, so `ScannerConfig{SGBinary: filepath.Join(t.TempDir(), "sg")}` yields a **non-sentinel** `Scan` error with no fake binary and no rules fixture.

### A.2 Gaps — assumptions NOT verified

- **G-1** `sg` presence on the run-phase host — neutralized by Shape 2's mandatory positive control.
- **G-2** The tests calling `InitDependencies()` directly were not read; M1 may perturb them.
- **G-3** Existing assertions that `QualityGate.Run` returns `""` on success were not enumerated (`gate_test.go` is 26 KB).
- **G-5** AC-FAG-004's revert-flips-result outcome was derived from code reading, not executed. The criterion is written as a mandatory executed round trip.
- **G-6 — RETRACTED.** The previous revision claimed the config-loaded path leaves `AstGrepGate.Enabled` at zero-value `false`, conflicting with `DefaultGateConfig()`. **That premise was false.** The chain resolves it: `internal/config/loader.go:36` seeds `cfg := NewDefaultConfig()` (→ `defaults.go:297` `Enabled: true, WarnOnlyMode: true`), `loader.go:89` calls `loadGateSection`, and `loader_gate.go:20-21` seeds its wrapper from the already-populated `cfg.Gate`, with the doc comment stating that an absent or partial `gate.yaml` yields `Enabled=true`. There is no conflict; both paths yield `true`. (`CLAUDE.local.md` §2.2 asserts the opposite and is **stale** relative to `loader_gate.go` — not a valid source.) AC-FAG-011/012 still construct their config explicitly, for **test determinism** — a test must not depend on ambient project config — not because of any default ambiguity.
- **G-7 — SPLIT AND PARTLY CLOSED (plan-audit iter-2).** The prior blanket deferral of AC-FAG-017/018 baselines was wrong in one half and is now corrected. **The rule going forward: a deferral is justified by cost, not by not having run it.**
  - **Closed — `go vet`**: measured, **exit 0**. Folded into AC-FAG-018.
  - **Closed — `gofmt`**: measured, **107 unclean files** repo-wide. Deferring this concealed a criterion-breaking fact — the old AC-FAG-018 expected 0, which no correct implementation of this SPEC could ever produce. See AC-FAG-018 for the re-scoped criterion and the 107-file out-of-scope observation.
  - **Still deferred (accepted) — `go test ./...` skip count**: expensive, and it runs against a shared tree. It remains a Definition-of-Done gate that MUST be measured before M1 begins.
  - **Still deferred — Windows cross-build**: not run at plan-phase; measured at run-phase per AC-FAG-018.
- **G-8 (new)** Shape 2 is POSIX-only. The shell criteria carry no Windows coverage; Windows is covered only by AC-FAG-018's cross-build.

---

## §B M1 — Logging scope (6 criteria)

### AC-FAG-001 — exactly one `slog.SetDefault` site remains

Covers REQ-FAG-001.

```bash
grep -rn 'slog\.SetDefault' internal/ cmd/ --include='*.go' | grep -v '_test\.go' | wc -l
```

Expect **1**. **Measured baseline: 2** — `internal/cli/deps.go:91`, `internal/cli/root.go:106`. A measured 2 → 1 delta, not an absolute assertion.

*Fails when*: either original site survives, or a third is introduced.

### AC-FAG-002 — level resolution table

Covers REQ-FAG-004, 005, 006, 007. Shape 1.

```bash
go test -list '^TestResolveLogLevel$' ./internal/cli/ | grep -c '^TestResolveLogLevel$'                                     # expect: 1
go test -count=1 -v -run '^TestResolveLogLevel$' ./internal/cli/ 2>&1 | grep -c '^    --- PASS: TestResolveLogLevel/'       # expect: >= 5
```

The table MUST cover: unset → `warn`; `debug` → `debug`; `info` → `info`; `error` → `error`; unrecognized → `warn`. Counting **subtest** PASS lines means a table that silently loses rows fails.

**Measured baseline: 0 / 0** — the test does not exist.

Env-name constant discipline:

```bash
grep -rln 'config\.EnvLogLevel' internal/cli/ --include='*.go' | wc -l                                    # expect: >= 1
grep -rn '"MOAI_LOG_LEVEL"' internal/cli/ --include='*.go' | grep -v '_test\.go' | wc -l                  # expect: 0
```

**Measured baselines: 0 and 0.** The first flips 0 → ≥1 (discriminating); the second is a standing prohibition that must not regress.

*Fails when*: a level is mis-mapped, the unrecognized-value fallback errors instead of defaulting, a table row is deleted, or the env name is inlined.

### AC-FAG-003 — hook path discards, non-hook path does not

Covers REQ-FAG-002, 003. Shape 1.

```bash
go test -list '^TestLoggingHandlerSelection$' ./internal/cli/ | grep -c '^TestLoggingHandlerSelection$'                             # expect: 1
go test -count=1 -v -run '^TestLoggingHandlerSelection$' ./internal/cli/ 2>&1 | grep -c '^    --- PASS: TestLoggingHandlerSelection/'  # expect: >= 4
```

Subtests MUST assert: `["hook","pre-tool"]` → discarding handler; `["doctor"]`, `["ast-grep","."]`, `["update"]` → stderr handler.

**Measured baseline: 0 / 0.**

*Fails when*: the hook carve-out is dropped (hook records leak), or it over-matches and silences non-hook commands.

### AC-FAG-004 — falsification: the scanner warning becomes observable, and reverting removes it

Covers REQ-FAG-003 and REQ-FAG-030. **Evaluation point: after the M1 commit, before any M2 change.** The comparison is anchored to base commit `6763aff3b`, so it works whether or not M1 is committed.

```bash
# Setup
BINDIR=$(mktemp -d); BIN="$BINDIR/moai"; go build -o "$BIN" ./cmd/moai
EMPTY=$(mktemp -d)
PATH="$EMPTY" command -v sg     # expect: no output, non-zero exit   (Shape 2)
command -v sg                    # expect: a path, exit 0            (positive control)

# (a) after M1 — the scanner's pre-existing slog.Warn is visible on stderr
env PATH="$EMPTY" "$BIN" ast-grep ./internal/astgrep 1>/tmp/fag4.out 2>/tmp/fag4.err
wc -c < /tmp/fag4.err                                   # expect: > 0
grep -c 'ast-grep (sg) CLI not found' /tmp/fag4.err      # expect: >= 1

# (b) revert round trip — NON-MUTATING, commit-safe. NOT git stash.
BASE=$(mktemp -d)/base
git worktree add --detach "$BASE" 6763aff3b
go -C "$BASE" build -o "$BASE/moai-base" ./cmd/moai
env PATH="$EMPTY" "$BASE/moai-base" ast-grep ./internal/astgrep 1>/dev/null 2>/tmp/fag4-base.err
wc -c < /tmp/fag4-base.err                              # expect: 0
git worktree remove "$BASE"
```

**Measured baseline (unmodified tree — exactly the (b) condition): stderr = 0 bytes, exit 0, stdout `no findings`, `ast-grep (sg) CLI not found` count = 0.** So (b)'s target is the real current behavior; (a) is what M1 must create.

**Why not `git stash`.** The previous revision used a `git stash push` / `pop` round trip. That is prohibited: `main-checkout-branch-guard.md` forbids `git stash` because the stash is **repository-global** and this repository has 50+ live worktrees sharing one git dir, so a push/pop pair can silently absorb another session's uncommitted work. It also becomes a no-op once M1 is committed, producing a spurious FAIL and a failing `pop`. `git worktree add --detach` is the sanctioned isolation procedure in that same rule.

*Fails when*: M1 does not change which records are emitted ((a) stays at 0 bytes), or the emission comes from something other than the repaired logging path ((b) is non-zero).

*Per G-5*: both halves MUST be executed and both byte counts recorded. A derived expectation is not evidence.

### AC-FAG-005 — no `slog` record reaches stdout

Covers REQ-FAG-008. Same setup as AC-FAG-004.

```bash
grep -c 'level=' /tmp/fag4.out    # expect: 0
grep -c 'msg=' /tmp/fag4.out      # expect: 0
```

**Measured baseline: 0 and 0** (stdout is `no findings`). This is a **regression guard**, not a delta — it holds today and must still hold after M1. Its falsification is constructive: build M1 with the handler over `os.Stdout` and both counts go non-zero.

*Fails when*: the non-hook handler is constructed over stdout, corrupting every `--format=json` consumer.

### AC-FAG-006 — trivial fast path still bypasses full initialization

Covers REQ-FAG-009 and REQ-FAG-030 clause (b). **Discriminator replaced (plan-audit iter-2 N2).**

The previous revision's round trip asserted that deleting the trivial branch makes `--version` emit warn records to stderr. That is **not a property of the change** — it is a property of the host's tooling. All five warn-or-above sites reachable on the full-init path are conditional on a *failure* state, and none fires on a healthy host:

| Site | Fires only when |
|------|-----------------|
| `deps.go:112` | `gopls.LoadConfig` errors |
| `deps.go:120` | `gopls.NewBridge` errors |
| `deps.go:174` | `goplsBridge == nil` |
| `deps.go:143` | `deps.Config.Load(cwd)` errors |
| `internal/config/loader.go:41` | the config sections dir is absent |

Verified host state: `gopls` present at `/Users/goos/go/bin/gopls`, `.moai/config/sections/lsp.yaml` present (8244 bytes) and enabled, sections dir present. On this state none fires, stderr stays 0 bytes, and the mandatory round trip would never flip.

**Chosen replacement: assert the `deps` global directly** (coordinator option (c), a controlled in-test probe). Justification: `var deps *Dependencies` (`internal/cli/deps.go:76`) is assigned **only** by `InitDependencies` (`deps.go:132`, `deps.go:271`). Its nil-ness is therefore an exact, binary, host-independent witness of whether full initialization ran — which is precisely what REQ-FAG-009 asserts. Options (a) and (b) were rejected: (a) would require M1 to introduce a warn record purely to be observed, adding production code for a test's benefit; (b) asserts handler wiring, which is already AC-FAG-003's job and would not prove the branch was *taken*.

**Half 1 — runtime regression guard (shell).**

```bash
"$BIN" --version > /tmp/v.out 2>/tmp/v.err; echo "exit=$?"    # expect: exit=0
wc -l < /tmp/v.out                                            # expect: 1
wc -c < /tmp/v.err                                            # expect: 0
```

**Measured baseline: exit=0, stdout 1 line, stderr 0 bytes.** This half is a guard (holds today, must still hold), not a delta. It is explicitly **not** the discriminator.

**Half 2 — the discriminator (Go test, host-independent).** Shape 1.

```bash
go test -list '^TestTrivialPathSkipsFullInit$' ./internal/cli/ | grep -c '^TestTrivialPathSkipsFullInit$'                              # expect: 1
go test -count=1 -v -run '^TestTrivialPathSkipsFullInit$' ./internal/cli/ 2>&1 | grep -c '^--- PASS: TestTrivialPathSkipsFullInit'      # expect: >= 1
```

**Measured baseline: 0 / 0** — the test does not exist.

The test is in `package cli` (same package, so the unexported `deps` global is visible). It MUST:

1. save and restore `os.Args`; set `os.Args = []string{"moai", "--version"}`;
2. reset `deps = nil`;
3. call `Execute()`;
4. assert `deps == nil` — full initialization did **not** run;
5. assert `isTrivialCommand([]string{"--version"})` is true and `isTrivialCommand([]string{"doctor"})` is false.

*Falsification (Shape 8 — MUST be executed and recorded)*:

```bash
PROBE=$(mktemp -d)/probe
git worktree add --detach "$PROBE" HEAD
#   in "$PROBE" only: delete the isTrivialCommand branch from Execute()
#   so every invocation calls InitDependencies()
go -C "$PROBE" test -count=1 -run '^TestTrivialPathSkipsFullInit$' ./internal/cli/
# expect: FAIL at assertion 4 (deps is non-nil)
git worktree remove --force "$PROBE"
```

The SPEC worktree's tracked source is never mutated. The failure is deterministic: it depends only on whether `InitDependencies()` ran, not on whether any warn record happened to fire.

*What would make this fail*: the trivial branch being deleted, bypassed, or reordered after `InitDependencies()`; or `InitDependencies` ceasing to assign `deps` (which would break the witness — if the run phase changes that assignment, this criterion must be re-derived, not silently kept).

### §B evaluation gate

AC-FAG-004 (a) MUST be green before any M2 file is edited. It is the only criterion whose observable — the pre-existing `slog.Warn` at `scanner.go:236` — exists solely in the M1-done / M2-not-started window.

---

## §C M2 — ast-grep sentinel, gate, doctor, docs (10 criteria)

### AC-FAG-007 — sentinel is returned, matchable, and carries guidance

Covers REQ-FAG-010, 011, 012, 013. Shape 1 + Shape 6 + Shape 7.

The previous revision judged REQ-FAG-012's message-content clauses with `go doc ... | grep -c ErrScannerUnavailable`, which only proves the symbol is exported and cannot observe the runtime message. Replaced: the message assertions move **into the test**.

```bash
go test -list '^TestScannerScan_UnavailableSentinel$' ./internal/astgrep/ | grep -c '^TestScannerScan_UnavailableSentinel$'                              # expect: 1
go test -count=1 -v -run '^TestScannerScan_UnavailableSentinel$' ./internal/astgrep/ 2>&1 | grep -c '^    --- PASS: TestScannerScan_UnavailableSentinel/'  # expect: >= 3
```

The three mandatory subtests:

1. **unavailable** — `t.Setenv("PATH", t.TempDir())` (Shape 6), `SGBinary: "sg"`: `errors.Is(err, astgrep.ErrScannerUnavailable)` is **true**, AND `strings.Contains(err.Error(), "sg")` is true, AND the message contains the install-guidance substring (REQ-FAG-012's two clauses, both asserted on the real runtime error).
2. **non-sentinel** — `SGBinary: filepath.Join(t.TempDir(), "sg")` (Shape 7): `err != nil` AND `errors.Is(err, astgrep.ErrScannerUnavailable)` is **false**.
3. **clean scan** — `sg` resolvable: `err == nil` and `errors.Is(...)` is false. Guarded by the Shape-2 positive control; when `sg` is absent on the host this subtest skips and the skip is **recorded as a gap**, never counted as a pass.

**Measured baseline: 0 / 0** — the test does not exist.

*Fails when*: `Scan` still returns nil on the unavailable path; the error is a bare `errors.New` that `errors.Is` cannot match through the wrap; the message omits the binary name or the guidance; or the sentinel over-matches and swallows unrelated errors (subtest 2).

### AC-FAG-008 — `moai ast-grep` exits non-zero with stderr guidance

Covers REQ-FAG-014, 015. Shape 2 + Shape 3.

```bash
env PATH="$EMPTY" "$BIN" ast-grep ./internal/astgrep 1>/tmp/fag8.out 2>/tmp/fag8.err; echo "exit=$?"
# expect: exit=1
grep -ci 'install' /tmp/fag8.err     # expect: >= 1
grep -c 'no findings' /tmp/fag8.out  # expect: 0
```

**Measured baseline: exit=0, stderr install-mentions 0, stdout `no findings` count 1.** All three flip — the strongest criterion in the set.

*Positive control* (mandatory, not optional):

```bash
command -v sg > /dev/null && { "$BIN" ast-grep ./internal/astgrep > /dev/null 2>&1; echo "with-sg exit=$?"; } \
  || echo "NOT RUNNABLE: sg absent on host — recorded as a gap"
```

Expect `with-sg exit=0` or `1` (findings-dependent) — never the unavailable path. `sg` is present on the authoring host, so this control is runnable.

*Fails when*: exit stays 0 (the original false all-clear), or the command still claims a clean result it never produced.

### AC-FAG-009 — the message never reaches stdout, under any format

Covers REQ-FAG-016.

```bash
for f in text json sarif; do
  env PATH="$EMPTY" "$BIN" ast-grep --format=$f ./internal/astgrep 1>/tmp/fag9-$f.out 2>/dev/null
  echo "$f stdout install-mentions: $(grep -ci install /tmp/fag9-$f.out)"
done
# expect: 0 for all three
```

**Measured baseline: 0 / 0 / 0.** A regression guard (holds today, must still hold), not a delta. Falsification is constructive: write the guidance with `cmd.OutOrStdout()` and all three go non-zero.

*Fails when*: guidance goes to stdout, breaking JSON and SARIF consumers.

### AC-FAG-010 — `ast-edit` behavior is unchanged

Covers REQ-FAG-017. Shape 5.

```bash
env PATH="$EMPTY" "$BIN" ast-edit --pattern 'x' --rewrite 'y' --lang go ./internal/astgrep 1>/tmp/fag10.out 2>/dev/null; echo "exit=$?"
# expect: exit=0
grep -c 'nothing to apply' /tmp/fag10.out    # expect: >= 1

git diff --stat 6763aff3b -- internal/cli/astedit.go | wc -l    # expect: 0
```

**Measured baselines: exit=0, `nothing to apply` count 1, anchored diff 0.** The previous revision used an unanchored `git diff --stat --`, empty by construction once committed; the base-anchored form stays meaningful after the run-phase commits.

*Fails when*: the deliberate D-7 asymmetry is "tidied up" into symmetry, turning a genuine mutator no-op into a spurious failure.

### AC-FAG-011 — the skip reason survives all three frames (reachability)

Covers REQ-FAG-018, 021, 022, 023, 030. **The make-or-break criterion**: changing `astgrep_gate.go` alone leaves the reason inert, because `gate.go:284` and `pre_tool.go:390` both discard it on the pass path.

The previous revision cited `TestQualityGate_Run_AstGrepGuardReachable` as its construction precedent. **That citation is withdrawn** — that test's own header says *"No sg binary is required"*; it exercises the pure-Go suppression branch and never reaches `Scanner.Scan`. It contributes only the marker/skip-tests trick, cited as such below.

**Construction — every gate on the path, enumerated and verified:**

| # | Gate | Construction | Verified at |
|---|------|--------------|-------------|
| 1 | `QualityGate.Run` returns early when `detectToolchain()` is nil | Write `build.zig` into the temp project. The Zig toolchain entry has **only** a `testStep` — no `vetSteps`, no `lintSteps` | `gate.go:220-224` |
| 2 | vet + lint steps must pass before the ast-grep step | Satisfied by #1 (Zig has none); additionally set `SkipTests: true` | `gate.go:265-277` |
| 3 | `RunAstGrepGateV2` runs the suppression sweep **before** the scan; any unpaired `ast-grep-ignore` returns `(false, …)` → deny path | Write no `ast-grep-ignore` anywhere under the temp project | `astgrep_gate.go:27-39` |
| 4 | `sg` must be absent | `t.Setenv("PATH", t.TempDir())` — Shape 6; test is **non-parallel** | `scanner.go:215-222` |
| 5 | `pre_tool.go` reaches the gate only for `ToolName == "Bash"` && `IsGitCommit(command)` | `HookInput{ToolName: "Bash", ToolInput: {"command":"git commit -m \"x\""}}` — `IsGitCommit` is a regex over the command string | `pre_tool.go:388`, `gate.go:625` |
| 6 | `projectDir` must be the temp project | Construct `&preToolHandler{cfg: …, policy: …, projectDir: tmpDir}` directly — the hook tests are `package hook` (89 files) and existing tests already use this literal | `pre_tool.go:314-327`, precedent `coverage_boost_test.go:216` |
| 7 | The ast-grep sub-gate must be enabled | Set `AstGrepGate: &AstGrepGateConfig{Enabled: true, RulesDir: …}` explicitly, for test determinism (per retracted G-6, the default is already `true`) | `defaults.go:297`, `loader_gate.go:20` |

No production seam is required — every gate is reachable from a test.

```bash
go test -list '^TestPreTool_AstGrepSkipReasonSurfaces$' ./internal/hook/ | grep -c '^TestPreTool_AstGrepSkipReasonSurfaces$'                              # expect: 1
go test -count=1 -v -run '^TestPreTool_AstGrepSkipReasonSurfaces$' ./internal/hook/ 2>&1 | grep -c '^--- PASS: TestPreTool_AstGrepSkipReasonSurfaces'     # expect: >= 1
```

**Measured baseline: 0 / 0.**

The test's three mandatory assertions:

1. the handler's returned `HookOutput.SystemMessage` is non-empty and names the skip;
2. the permission decision is **not** `deny`;
3. the reason observed at the handler equals the reason the gate step produced — the same string traversed all three frames.

**Dual falsification — both forms MUST be executed and recorded. Both run in a detached probe worktree at `HEAD` per Shape 8 (N3); neither edits tracked source in the SPEC worktree.**

- *Form A (call deleted)*: in the probe worktree, remove the `SystemMessage` assignment in `pre_tool.go` → this test fails there. Proves the emission is load-bearing.
- *Form B (body neutered)*: in a fresh probe worktree, restore `gate.go`'s pass branch to discard `out`, leaving `pre_tool.go` intact → this test fails there. Proves the **propagation**, not merely the final assignment, is load-bearing.

A guard that only proves the function was called does not prove it did its job; Form B is what distinguishes the two. Remove each probe worktree (`git worktree remove --force`) after recording the result.

### AC-FAG-012 — unavailable and other-error reasons are distinct

Covers REQ-FAG-019. **Restated** from the previous revision, which required a non-sentinel error to reach `RunAstGrepGateV2` — unreachable, because that function hard-codes `SGBinary: "sg"` (`astgrep_gate.go:45`) and offers no injection seam. Asserting it would have forced an unplanned production seam or a fake-`sg` fixture neither artifact planned for.

Two mechanically-checkable halves, each at the frame where it is observable:

```bash
go test -list '^TestAstGrepGateReasons$' ./internal/hook/quality/ | grep -c '^TestAstGrepGateReasons$'                              # expect: 1
go test -count=1 -v -run '^TestAstGrepGateReasons$' ./internal/hook/quality/ 2>&1 | grep -c '^    --- PASS: TestAstGrepGateReasons/'  # expect: >= 2
```

1. **Reachable half (E2E)** — with `t.Setenv("PATH", t.TempDir())` (Shape 6) and an explicitly-enabled config, `RunAstGrepGateV2` returns `(true, r)` with `r` non-empty and naming the skip.
2. **Distinctness half (direct)** — assert both reason values are non-empty and differ from each other. Per REQ-FAG-032 the two values are package-level named constants, so the test asserts on the constants directly rather than round-tripping through the gate. This is what prevents the collapse-to-one-string failure. (N4: the named-constant shape is no longer smuggled in as an AC-only prescription — it is now a stated requirement with its testability rationale in `spec.md` §B.4.)

**Measured baseline: 0 / 0.**

*Explicitly NOT covered*: the non-sentinel branch has **no end-to-end path** through `RunAstGrepGateV2` without a fake `sg` on `PATH` plus a populated rules directory. That fixture is named in `plan.md` §F as an optional deepening and is deliberately not required here. The scanner-level classification it would exercise IS covered — by AC-FAG-007 subtest 2, via the Shape-7 lever. Recorded as residual risk, not silently omitted.

*Fails when*: both classes collapse to one reason string, or either returns the pre-change empty string.

### AC-FAG-013 — the gate never denies because `sg` is absent

Covers REQ-FAG-020. Shape 1 — the previous revision used a bare `grep -c '^--- FAIL'` with no existence proof, so a renamed or non-matching selector produced 0 FAIL lines and "passed".

```bash
go test -list '^TestQualityGate_Run_AstGrep' ./internal/hook/quality/ | grep -c '^TestQualityGate_Run_AstGrep'                              # expect: >= 2
go test -count=1 -v -run '^TestQualityGate_Run_AstGrep' ./internal/hook/quality/ 2>&1 | grep -c '^--- PASS: TestQualityGate_Run_AstGrep'      # expect: >= 2
go test -count=1 -v -run '^TestQualityGate_Run_AstGrep' ./internal/hook/quality/ 2>&1 | grep -c '^--- FAIL'                                   # expect: 0
```

**Measured baselines: `-list` → 2, `--- PASS:` → 2, `--- FAIL` → 0.** The existing `TestQualityGate_Run_AstGrepGuardReachable` and `…GuardDisabled` must still pass unchanged — a regression there means pass/deny semantics moved, not just the reason. The `-list` count is what makes a rename fail the criterion instead of silently satisfying it.

AC-FAG-011 assertion 2 covers the same property at the handler frame.

*Fails when*: the sentinel branch returns `false`, converting a missing optional tool into a blocked commit.

### AC-FAG-014 — docs updated in all four locales

Covers REQ-FAG-028, 029. **Fully replaced** — both of the previous revision's greps were non-discriminating: `grep -lc 'ast-grep.github.io' … | wc -l` already returned **4** before any edit (and `-l` suppresses `-c`, so it counted files, not matches), and the stale-claim grep's baseline was **1**, not 4 — only the `en` page carries the English string — so an en-only edit satisfied both. That contradicted risk R-8's claimed mitigation.

```bash
# (a) per-locale change proof — 4/4 must be non-zero  (Shape 5)
for f in docs-site/content/*/cli-reference/ast-grep.md; do
  printf "%s -> %s\n" "$(basename $(dirname $(dirname $f)))" "$(git diff --stat 6763aff3b -- "$f" | wc -l | tr -d ' ')"
done
# expect: en, ja, ko, zh all > 0

# (b) locale-neutral content anchor — 4/4 must carry the install target
grep -l 'guide/quick-start' docs-site/content/*/cli-reference/ast-grep.md | wc -l    # expect: 4

# (c) the stale English claim is gone from the en page
grep -c 'exits without an error' docs-site/content/en/cli-reference/ast-grep.md      # expect: 0
```

**Measured baselines: (a) en 0, ja 0, ko 0, zh 0 — all four flip. (b) 0 files carry `guide/quick-start` today. (c) 1.** An en-only edit now fails (a) on three locales and fails (b) with a count of 1.

The anchor in (b) is the deep install URL (`ast-grep.github.io/guide/quick-start`), chosen because the bare domain already appears in all four files and is therefore useless as an anchor.

*Human obligation (not a criterion)*: (c) is English-only by construction. The run phase MUST additionally **read** all four pages and confirm each states the asymmetric behavior (`ast-grep` non-zero, `ast-edit` zero). A grep cannot judge translated prose; the reading is recorded in the Definition of Done, not counted as a mechanical check.

*Fails when*: one locale is edited and three are not (R-8), or the install target is omitted from any locale.

### AC-FAG-015 — `moai doctor` reports `sg`

Covers REQ-FAG-024, 026. The judging surface is the exported diagnostics JSON, **not** the process exit code — `moai doctor` returns nil regardless of failure count (`doctor.go:99-124`), so an exit-code assertion would be meaningless. (Refinement per audit: `doctor.go` can return non-nil from an `exportDiagnostics` failure; the load-bearing fact — failure count never influences the return — holds.)

```bash
EXPORT=$(mktemp -d)/diag.json
env PATH="$EMPTY" "$BIN" doctor --export "$EXPORT" >/dev/null 2>&1
grep -ci 'ast-grep\|"sg"' "$EXPORT"     # expect: >= 1

grep -c 'exec.LookPath' internal/cli/doctor.go     # expect: >= 3
```

**Measured baselines: JSON grep 0** (22 checks, statuses `['fail','ok','warn']`, no check name contains `sg`); **`exec.LookPath` count 2** (git, gh).

The second command is the registration guard: a `checkAstGrep` written but never added to `systemChecks` would leave the JSON count at 0.

*Fails when*: the check is defined but not registered — the registration-omission failure mode.

### AC-FAG-016 — the `sg` check warns, with guidance in its message

Covers REQ-FAG-025, 027. Extended per audit: the previous revision asserted only `status != 'fail'`, which a permanently-`ok` check would also satisfy, leaving REQ-FAG-025's "non-OK status" and "install guidance" clauses unjudged.

```bash
python3 - "$EXPORT" <<'EOF'
import json,sys
checks=json.load(open(sys.argv[1]))
rows=[c for c in checks if 'sg' in c['name'].lower() or 'ast-grep' in c['name'].lower()]
assert len(rows)==1, f"expected exactly 1 sg check, got {len(rows)}"
r=rows[0]
assert r['status']=='warn', f"sg-absent must be warn (not fail, not ok), got {r['status']}"
msg=(r.get('message','')+' '+r.get('detail','')).lower()
assert 'install' in msg or 'ast-grep.github.io' in msg, f"no install guidance in: {r}"
print("OK", r['name'], r['status'])
EOF
# expect: OK <name> warn
```

`uikit.CheckStatus` is a string type (`internal/cli/uikit/types.go:9`, values `"ok"`/`"warn"`/`"fail"`), so the comparison is a real string comparison. Run under the stripped `PATH` of AC-FAG-015 so the absent branch is the one exercised.

**Measured baseline: 0 matching rows** — the harness aborts on the first assertion today.

*Fails when*: the check reports `fail` (misrepresenting an optional tool as a broken install, and inflating the `--fix` list with an unfixable entry), reports a constant `ok` regardless of `sg`, or omits guidance.

---

## §D Cross-cutting criteria (2)

### AC-FAG-017 — full suite green, no new skips

Covers REQ-FAG-031. Shape 4 — the previous revision read `$?` after a pipe into `tail`, which always reports 0.

```bash
go test ./... > /tmp/fag17.log 2>&1; echo "exit=$?"; tail -30 /tmp/fag17.log
# expect: exit=0

grep -c '^--- SKIP' /tmp/fag17.log
# expect: <= the pre-M1 baseline
```

**Baseline: NOT MEASURED (G-7).** It MUST be measured on the unmodified tree **before M1 begins** and recorded in `progress.md` §E.2. A comparison against an unmeasured baseline is not a comparison, and this criterion is not judgeable until that number exists.

The redirect also satisfies the file-redirect contract in `agent-common-protocol.md` § File-redirect contract.

*Fails when*: a test is converted to `t.Skip` to accommodate the change rather than updated.

### AC-FAG-018 — vet, format, and cross-platform build clean

Covers the repo's standing constraints. This criterion does **not** cover REQ-FAG-030 — `go vet`, a Windows build, and `gofmt` cannot observe revert-behaviour; REQ-FAG-030 traces to AC-FAG-004, AC-FAG-006, and AC-FAG-011.

**Format half re-scoped (plan-audit iter-2 N1).** The previous revision expected `gofmt -l internal/ cmd/ | wc -l` → 0. The measured repo-wide value is **107**, so a *correct* implementation of this SPEC would fail the criterion, and satisfying it would require reformatting 107 unrelated files — contradicting this SPEC's own §C exclusions. Per user decision, the check is scoped to the files this SPEC touches, with the list stated **explicitly** rather than derived from an unpinned `git diff`.

```bash
go vet ./... > /tmp/fag18-vet.log 2>&1; echo "vet exit=$?"                                   # expect: 0
GOOS=windows GOARCH=amd64 go build ./... > /tmp/fag18-win.log 2>&1; echo "win exit=$?"       # expect: 0

# Format — scoped to this SPEC's change set (explicit list, M1 + M2)
gofmt -l \
  internal/cli/deps.go \
  internal/cli/root.go \
  internal/cli/logging.go \
  internal/astgrep/scanner.go \
  internal/cli/astgrep.go \
  internal/hook/quality/astgrep_gate.go \
  internal/hook/quality/gate.go \
  internal/hook/pre_tool.go \
  internal/cli/doctor.go \
  | wc -l
# expect: 0
```

`internal/cli/logging.go` is the new M1 file — substitute the actual filename if the run phase names it differently, and record the substitution. New `_test.go` files are appended to this list as they are created.

**Measured baselines:**

| Command | Baseline | Expected |
|---|---|---|
| `go vet ./...` | **exit 0** | 0 |
| `GOOS=windows GOARCH=amd64 go build ./...` | NOT MEASURED — deferred to run-phase | 0 |
| scoped `gofmt -l` (the 8 files that exist today) | **1** — `internal/cli/root.go` | 0 |
| repo-wide `gofmt -l internal cmd` | **107** — out of scope, see below | (not asserted) |

**The scoped baseline is 1, not 0 — that is what makes this criterion discriminate.** `internal/cli/root.go` is currently unclean and is a file M1 edits. Its entire `gofmt` delta is 6 lines of whitespace alignment in the `trivialCommands` map literal (`root.go:43-51`), which `gofmt` widens to accommodate the longer `"completion":` key. Bringing that one file to clean is in scope because M1 edits it anyway; the resulting hunk is small and self-explanatory. The other 7 existing files were verified clean **per-file** (a directory-level `gofmt -l` and a per-file `gofmt -l` were cross-checked to confirm the count).

**Out-of-scope observation (do NOT act on):** 106 further files under `internal/` + `cmd/` are `gofmt`-unclean. They are not generated and not a toolchain skew (`go1.26.4`, matching `go.mod`). The repository's actual formatter is **`gofumpt`** (`Makefile:61` — `gofumpt -l -w .`), and `.github/workflows/` contains **no** format guard at all, which is how the debt accumulated silently. `gofumpt` is **not installed on this host**, so no criterion here may depend on it. Cleaning those 106 files is explicitly NOT part of this SPEC and MUST NOT be bundled into its diff — see `spec.md` §C.

*Fails when*: the change introduces a platform-specific construct; unformatted code lands in one of the listed files; or `root.go` is edited by M1 without being brought to `gofmt`-clean.

*Reproducibility note*: `gofmt` resolves to `/opt/homebrew/bin/gofmt` on the authoring host with `go1.26.4`. If a run-phase host reports a different scoped baseline on the unmodified tree, record it before treating any delta as this SPEC's doing.

---

## §E Definition of Done

- [ ] All 18 criteria executed, verbatim command output recorded in `progress.md` §E.2.
- [ ] Shape-2 positive control (`command -v sg` → a path) executed on the run host; if `sg` is absent, every Shape-2 criterion is recorded **NOT RUNNABLE**, never passed.
- [ ] AC-FAG-017 skip-count baseline measured on the unmodified tree **before** M1 begins (the remaining half of G-7; `go vet` = exit 0 and scoped `gofmt` = 1 are already measured and recorded in AC-FAG-018).
- [ ] All three falsification round trips (AC-FAG-004, AC-FAG-006, AC-FAG-011 Form A + Form B) executed in **detached probe worktrees** per Shape 8 — **no `git stash`, no in-place edit of tracked source** — and every probe worktree removed afterward.
- [ ] AC-FAG-004 both halves executed with both byte counts recorded.
- [ ] AC-FAG-006 Half 2 falsification recorded (probe worktree: branch deleted → test FAILs at assertion 4).
- [ ] AC-FAG-011 both falsification forms (call-deleted AND body-neutered) executed and recorded.
- [ ] `internal/cli/root.go` brought to `gofmt`-clean as part of M1 (scoped baseline 1 → 0); the other 106 unclean files left untouched.
- [ ] AC-FAG-007 subtest 3 either executed or recorded as a skip-with-reason gap.
- [ ] All four docs-site locale pages **read** (not merely grepped) and confirmed to describe the asymmetric behavior.
- [ ] Every gap in §A.2 either closed with evidence or restated as residual risk in the run-phase report.
- [ ] No file outside the surfaces named in `spec.md` §E modified.

## §F Quality gates

| Gate | Threshold |
|------|-----------|
| Test suite | `go test ./...` exit 0 (Shape 4 form) |
| Static analysis | `go vet ./...` exit 0 |
| Formatting | scoped `gofmt -l <this SPEC's file list>` = 0 (AC-FAG-018; repo-wide 107-file debt is out of scope) |
| Coverage — `internal/astgrep` | not below the pre-change measurement |
| Coverage — `internal/hook/quality` | not below the pre-change measurement |
| Cross-platform | `GOOS=windows GOARCH=amd64 go build ./...` exit 0 |
| Locale parity | 4/4 per-locale anchored diffs non-empty (AC-FAG-014a) |

Coverage thresholds are relative because no per-package figure was measured during plan-phase. The run phase measures the baseline first, then compares.
