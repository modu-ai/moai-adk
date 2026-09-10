# SPEC-CODEX-SKILL-PATH-READBACK-001 — M3 run evidence (card t562)

Worktree `.claude/worktrees/t562` · branch `WT-codex-read-inverse` · HEAD at measurement time
`4aa8915ee` · measured 2026-09-08 on darwin/arm64, go1.26.4.

M3 is plan.md §E M3 exactly: the doctor guard re-run + baseline diff, the four named mutants, the
seam-out and scope pins, the cross-platform build, and the scoped suite with its executed-test
control. **The production diff of this milestone is ZERO lines** — every code edit below was a
deliberate temporary mutant, injected, measured, and reverted, with the tree proven byte-clean after
each revert.

---

## 1. Claim

| # | Claim |
|---|---|
| C1 | AC-CSRB-006 — the doctor regression guard re-run post-M2 reproduces the M1 pre-change baseline exactly; every counter phrase is byte-identical. |
| C2 | All four named mutants were injected, executed against a test whose execution reached the mutated line, and **all four were CAUGHT**; the tree is byte-clean afterwards. |
| C3 | AC-CSRB-007 — this card's diff adds ZERO `osStatFn` tokens to `internal/cli/doctor_codex.go`, with a positive control proving the pattern matches the token when one exists. |
| C4 | AC-CSRB-008 — with `CARD_BASE` re-derived at read time, every changed path is inside the plan.md §G allowlist, the `internal/codexwiring/skills.go` probe is empty against a non-zero control, and the four PRESERVE files carry zero card-authored change after the absorb merge. |
| C5 | AC-CSRB-009 — `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` both exit 0. Build-level only; no Windows runtime behaviour is claimed. |
| C6 | AC-CSRB-010 — the scoped `internal/cli` suite exits 0 with 6938 `--- PASS: ` and 0 `--- FAIL: `, above the BEFORE baseline of 6886. |
| C7 | Quality gates: `go vet ./internal/cli/...` rc=0; `golangci-lint run ./internal/cli/...` → `0 issues.` |

---

## 2. Evidence

Every grep below is `/usr/bin/grep` explicitly — the interactive shell's `grep` here is a ugrep
wrapper that skips files silently.

### 2.1 AC-CSRB-006 — doctor regression guard re-run vs. the M1 baseline

```
$ go test ./internal/cli/ -run 'TestCodexStaleSkillFinding_ShapeReadbackBaselineGuard' -count=1 -timeout 1800s -v \
    > .moai/reports/t562/ac-csrb-006-post-m2.log 2>&1; echo "rc=$?"
rc=0
```

Post-M2 output (`ac-csrb-006-post-m2.log`), verbatim:

```
=== RUN   TestCodexStaleSkillFinding_ShapeReadbackBaselineGuard
    codex_stale_skill_readback_test.go:55: summary: ~/.codex/config.toml: 1 stale skill entry
    codex_stale_skill_readback_test.go:56: detail: /var/folders/kt/.../TestCodexStaleSkillFinding_ShapeReadbackBaselineGuard1184892123/003/.codex/config.toml declares 4 [[skills.config]] entries; 1 with a path that no longer exists (1 enabled, 0 disabled, 0 unspecified, 0 non-boolean) — remove the stale entries or restore the skill files; 1 relative entry (not checked: the resolution base is not observed); 1 oddly-formed entry (not checked: backslash or ~other-user shape)
--- PASS: TestCodexStaleSkillFinding_ShapeReadbackBaselineGuard (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.109s
```

Diff against the M1 baseline `ac-csrb-006-baseline.log`:

```
$ diff ac-csrb-006-baseline.log ac-csrb-006-post-m2.log   # rc=1 — two differing lines
3c3   <detail line: TempDir nonce ...Guard3502885156... vs ...Guard1184892123...>
6c6   < ok  ... internal/cli	0.843s
      > ok  ... internal/cli	1.109s

$ diff <(norm baseline) <(norm post-m2)   # norm masks the TempDir nonce and elapsed seconds
      (no output)   rc=0
```

**Both differing lines are per-run nonces**, not behaviour: the `t.TempDir()` directory name and the
package elapsed time printed by `go test`. Every counter phrase — `declares 4 [[skills.config]]
entries`, `1 with a path that no longer exists (1 enabled, 0 disabled, 0 unspecified, 0
non-boolean)`, `1 relative entry`, `1 oddly-formed entry` — is byte-identical. Expected on darwin:
`configPathSeparator` is `'/'` at production settings, so the M2 conversion is the identity; a moved
counter would have meant a darwin-visible behaviour change.

The instruction anticipated the TempDir nonce as the only permitted difference; the elapsed-seconds
line is a second nonce of the same kind and is named here rather than folded into the first.

### 2.2 Mutant reachability probes (run BEFORE crediting any mutant)

An unreached mutant and a real survivor print the same `ok`, so each mutation site was first proven
executed by inserting `panic("MX-PROBE-…")` at the site and running the test that will judge the
mutant. All three sites panicked — reached.

| Probe | Site | Test run | Result | Log |
|---|---|---|---|---|
| P1 | `codex_skills_prune.go:83` (the absolute-branch conversion) | `TestJudgeCodexSkillEntry_SeparatorConversion` | `panic: MX-PROBE-P1-CONVERSION-SITE`, stack frame `judgeCodexSkillEntry(...) codex_skills_prune.go:83` → **REACHED** | `probe-p1.log` |
| P2 | `codex_skills_prune.go:101` (the stat call) | `TestJudgeCodexSkillEntry_HomeRelativeStatTargetStaysNative` | `panic: MX-PROBE-P2-STAT-SITE` → **REACHED** | `probe-p2.log` |
| P3 | `codex_skills_prune.go:76` (the `classifyCodexSkillPath` switch) | `TestJudgeCodexSkillEntry_ClassifiesDeclaredFormBeforeConversion` | `panic: MX-PROBE-P3-CLASSIFY-SITE` → **REACHED** | `probe-p3.log` |

The seam mutant (M4 below) has no Go-test judge — its check is a grep over a diff — so its
reachability analogue is the pattern actually matching the injected token; that transition is shown
in §2.3.

### 2.3 The four mutants

**M1 — bypass mutant** (`statPath = fromConfigPath(e.Path, configPathSeparator)` → `statPath = e.Path`),
judged by AC-CSRB-002 → **CAUGHT**. `mutant-bypass.log`, rc=1:

```
    codex_skills_prune_readback_test.go:90: stat target = "/var/folders/.../gone/SKILL.md", want the converted form "\\var\\folders\\...\\gone\\SKILL.md" (declared "/var/folders/.../gone/SKILL.md")
--- FAIL: TestJudgeCodexSkillEntry_SeparatorConversion (0.00s)
```

**M2 — blanket-wrap mutant** (branch conversion removed; `osStatFn(fromConfigPath(statPath,
configPathSeparator))` at the stat site instead), judged by AC-CSRB-004 → **CAUGHT**.
`mutant-blanket-wrap.log`, rc=1:

```
    codex_skills_prune_readback_test.go:139: stat target = "\\var\\folders\\...\\x\\SKILL.md", want the native Join product "/var/folders/.../x/SKILL.md" unrewritten
--- FAIL: TestJudgeCodexSkillEntry_HomeRelativeStatTargetStaysNative (0.00s)
```

This is the §H risk-row-1 shape: the home-relative `filepath.Join` product gets rewritten. The guard
sees it.

**M3 — reorder mutant** (`e.Path = fromConfigPath(e.Path, configPathSeparator)` inserted immediately
BEFORE `switch classifyCodexSkillPath(e.Path)`), judged by AC-CSRB-003 → **CAUGHT**.
`mutant-reorder.log`, rc=1:

```
    codex_skills_prune_readback_test.go:114: SkipReason = "oddly-formed path — not resolvable here", want "relative path — no observed resolution base" (classification ran on the declared form)
```

The observed failure string is exactly the discriminator acceptance.md predicted for the
convert-first order.

**M4 — seam mutant** (a card-authored `osStatFn` line added to `internal/cli/doctor_codex.go`),
judged by the AC-CSRB-007 delta predicate → **CAUGHT**. `mutant-seam.log`:

```
# pre-mutant control (clean tree)
count=0
# post-mutant, line injected at doctor_codex.go:861
		if _, mserr := osStatFn(statPath); mserr != nil { _ = mserr } // MUTANT-SEAM: card-authored osStatFn line
added-osStatFn-line count=1
verdict: pin FAILS iff count >= 1 -> FAILS (count=1) — mutant CAUGHT
build check: build rc=0
```

The mutant was exercised as an **uncommitted working-tree change**, measured with the same predicate
the pin uses (`git diff -- internal/cli/doctor_codex.go | /usr/bin/grep -c '^+.*osStatFn'`), because
committing a mutant is prohibited. The 0 → 1 transition is what demonstrates the predicate can fail;
see Gaps G3 for the limit this leaves.

**Mutant hygiene.** `git status --porcelain` and `git rev-parse --short HEAD` were re-read
immediately before each injection and immediately after each revert; HEAD never moved from
`4aa8915ee`. After the last revert:

```
$ git status --porcelain -- internal/cli/codex_skills_prune.go internal/cli/doctor_codex.go \
      internal/cli/codex_skills_prune_readback_test.go internal/cli/codex_stale_skill_readback_test.go
      (no output)
$ git diff --stat -- internal/
      (no output)
$ git status --porcelain            # only new, untracked evidence logs under .moai/reports/t562/
```

No mutant was committed and none survived between steps.

### 2.4 AC-CSRB-007 — doctor seam-out pin (`ac-csrb-007.log`)

```
$ git show c007e5409 --format='' -- internal/cli/doctor_codex.go | /usr/bin/grep -c 'osStatFn'
0
$ git show c007e5409 --format='' -- internal/cli/doctor_codex.go | /usr/bin/grep -c '^+.*osStatFn'
0
$ /usr/bin/grep -c 'osStatFn' internal/cli/codex_skills_prune.go        # POSITIVE CONTROL
1
$ /usr/bin/grep -c 'osStatFn' internal/cli/doctor_codex.go              # provenance context
2
$ git show c007e5409 -- internal/cli/doctor_codex.go | /usr/bin/grep -c 'osStatFn'   # WITHOUT --format=''
2
```

The control at 1 is what makes the 0 mean something: a zero from a broken pattern is
indistinguishable from a true zero, and neither operand of this comparison is empty. The last line is
the methodology witness the plan calls load-bearing — without `--format=''`, `git show` prints the
commit message, whose own attribution prose mentions `osStatFn` twice, yielding a spurious 2 that is
message text and not diff lines.

Generalized across **every** card-authored commit, not just the M2 one:

```
  4aa8915ee -> added-osStatFn-lines=0
  835215bab -> added-osStatFn-lines=0
  c007e5409 -> added-osStatFn-lines=0
  b78d2e425 -> added-osStatFn-lines=0
  f1654c924 -> added-osStatFn-lines=0
  fe2c8f51f -> added-osStatFn-lines=0
# the absorb merge 2d1dad058 (first-parent diff) -> 2
```

Clean attribution: the file's two occurrences arrive with the absorb of `c72dc1baf` (t563 /
SPEC-DOCTOR-STAT-SEAM-001) and with nothing this card wrote.

### 2.5 AC-CSRB-008 — scope pin (`ac-csrb-008.log`)

```
$ git fetch origin develop
$ CARD_BASE=$(git merge-base origin/develop HEAD)
CARD_BASE=c72dc1bafe8023628b7f9fdc746bc168bf93e6ad   (origin/develop=91d25bc61, HEAD=4aa8915ee)
$ git diff --name-only "$CARD_BASE"..HEAD | wc -l
31
$ git diff --name-only "$CARD_BASE"..HEAD -- internal/codexwiring/skills.go | wc -l
0
```

The left edge is the re-derived merge-base, never a pinned SHA (the t543 discipline). The full
31-path listing is the non-zero control for the empty probe — both operands are non-degenerate.

Allowlist check over all 31 paths: 8 under `internal/cli/` — `codex_skills_prune.go`,
`codex_skills_prune_readback_test.go`, `codex_stale_skill_readback_test.go`, `doctor_codex.go` (this
card's four) plus the four absorbed t540 files `codex_config_path.go`, `codex_config_path_test.go`,
`codex_skills_disable.go`, `codex_skills_disable_path_test.go` — and 23 under `.moai/` (this card's
and t540's SPEC artifacts and evidence). Nothing outside the plan.md §G allowlist.

PRESERVE strengthened — the four absorbed files carry **zero card-authored change after the absorb**:

```
$ git diff --stat 2d1dad058..HEAD -- <the four PRESERVE files>
      (no output)   line count = 0
$ git diff --stat 2d1dad058..HEAD | tail -1        # NON-ZERO control, same range
 10 files changed, 227 insertions(+), 30 deletions(-)
```

### 2.6 AC-CSRB-009 — cross-platform build (`ac-csrb-009.log`)

```
$ go build ./...                                  rc=0   (no output)
$ GOOS=windows GOARCH=amd64 go build ./...        rc=0   (no output)
$ GOOS=windows GOARCH=amd64 go vet ./internal/cli/   rc=0   (no output)
```

**Stated limit**: `go build` compiles non-test packages only — a GOOS cross-build does **not** compile
`*_test.go` files, so the windows build alone says nothing about this card's test files. The `go vet`
line is added as a stronger witness because vet *does* type-check test files under the target GOOS;
it is still compile-level. No Windows runtime behaviour is claimed, observed, or implied.

### 2.7 AC-CSRB-010 — scoped suite with executed-test control (`ac-csrb-010.log`, raw `ac-010.log`)

```
$ go test ./internal/cli/... -timeout 1800s -v > .moai/reports/t562/ac-010.log 2>&1; echo "rc=$?"
rc=0
$ /usr/bin/grep -c -- '--- PASS: ' .moai/reports/t562/ac-010.log      # AFTER
6938
$ /usr/bin/grep -c -- '--- FAIL: ' .moai/reports/t562/ac-010.log
0
$ /usr/bin/grep -c '^FAIL' .moai/reports/t562/ac-010.log
0
$ /usr/bin/grep -c -- '--- SKIP: ' .moai/reports/t562/ac-010.log
30
$ /usr/bin/grep -c -- '--- PASS: ' .moai/reports/t540/ac-006-base.log  # BEFORE, see baseline-attribution
6886
```

Predicate: `AFTER >= BEFORE AND AFTER > 0` → `6938 >= 6886` and `6938 > 0` → **PASS**. The trailing
space in `'--- PASS: '` is load-bearing: go appends ` (0.06s)`, so a `$` anchor would count 0 on every
line and read as a silent zero.

This card's own contribution inside that count is 8 `--- PASS: ` lines (5 top-level + 3 subtests),
all present in the AFTER log:

```
--- PASS: TestJudgeCodexSkillEntry_SeparatorConversion
--- PASS: TestJudgeCodexSkillEntry_ClassifiesDeclaredFormBeforeConversion
--- PASS: TestJudgeCodexSkillEntry_HomeRelativeStatTargetStaysNative
--- PASS: TestJudgeCodexSkillEntry_EligibilityGatingPins   (+3 subtests)
--- PASS: TestCodexStaleSkillFinding_ShapeReadbackBaselineGuard
```

AC-CSRB-003 / 004 / 005 are confirmed GREEN in this same full-suite run, so their green state is not
carried over from the M2 scoped run.

### 2.8 Quality gates (`quality-gates-m3.log`)

```
$ go vet ./internal/cli/...                              rc=0   (no output)
$ golangci-lint run --timeout=5m ./internal/cli/...      0 issues.   rc=0
```

---

## 3. Baseline-attribution

Everything above was measured **in this run, against this tree** — worktree
`.claude/worktrees/t562`, branch `WT-codex-read-inverse`, HEAD `4aa8915ee`, `git status` carrying no
tracked modification at each measurement point — except the two carried baselines below, whose
provenance is named rather than assumed:

| Carried value | What it is | Where it was measured |
|---|---|---|
| AC-CSRB-006 baseline (`ac-csrb-006-baseline.log`) | the PRE-CHANGE doctor counters | captured in the M1 window on tree `69cfdce6f`, **before** M2 landed — by construction it cannot be re-measured now, which is why M1 captured it |
| AC-CSRB-010 BEFORE = 6886 (`.moai/reports/t540/ac-006-base.log`) | a `--- PASS: ` count of the same command `go test ./internal/cli/... -timeout 1800s -v` | t540's pre-flight, tree `b4ce67468`, per t540 `run-m1-m2.md` §2.7 |

The AC-CSRB-010 BEFORE is a carried figure, and it is used as a **lower bound**, not as this card's
exact pre-card count — see Gap G1. `origin/develop` was re-fetched at read time for AC-CSRB-008 and
stood at `91d25bc61`; `CARD_BASE` was re-derived from it, never pinned.

---

## 4. Gaps — what was NOT observed

**G1 — no t562-own AC-CSRB-010 BEFORE baseline exists.** plan.md §C required capturing
`.moai/reports/t562/ac-010-base.log` at pre-flight, before the absorb merge, with the same command.
That file does not exist in this tree and was not captured in the M1 or M2 windows; it cannot be
captured now, because the pre-absorb tree state is gone. The BEFORE used here (6886) is **t540's**
pre-flight base at `b4ce67468`, which predates t540's own new tests AND t563's tests that arrived via
the `c72dc1baf` absorb. It is therefore a strict lower bound on the true pre-card count. Consequence:
the observed delta of +52 is **not decomposed** into t540 / t563 / this-card contributions beyond the
8 lines attributable to this card by name. The AC's `A >= B AND A > 0` predicate holds against the
lower bound, and the "expected delta accounted" clause of acceptance.md is only partially satisfied —
this is reported as a gap, not folded into the pass.

**G2 — no Windows runtime observation.** AC-CSRB-009 is compile-level. The Windows classification of
a `C:/…` declaration remains inference I-1; nothing here observes it.

**G3 — the seam mutant was not committed.** M4 was exercised as a working-tree change, so what was
demonstrated is that the *delta predicate* flips 0 → 1 when a card-authored `osStatFn` line exists,
not that a committed card-authored seam flips the specific `git show c007e5409 …` invocation (it
could not — that command names a fixed commit and is insensitive to any later commit). The
generalized per-commit sweep in §2.4 covers the committed axis for the commits that actually exist.

**G4 — the full repository suite was not run locally.** Scope was `internal/cli` per plan.md §G. The
full-suite verdict belongs to CI on the pushed head; this run says nothing about packages outside
`internal/cli/...`.

**G5 — no `-race` run.** Not required by any AC and not performed; the card touches no goroutine or
channel.

**G6 — the AC-CSRB-006 guard exercises the darwin path only.** With `configPathSeparator == '/'` the
M2 doctor conversion is the identity, so the guard proves the conversion did not *disturb* darwin
behaviour; it does not exercise a non-identity conversion on the doctor side. That asymmetry is the
declared honest ceiling of doctor-side verification until t563's seam exists (`spec.md §D.3`), not a
new finding.

**G7 — no independent check of the `--- SKIP: ` 30.** The skipped tests were counted but not
enumerated or attributed; nothing here claims they are the same 30 as any prior run.

**G8 — two measurements were tree-attributed by witness rather than re-executed.** After the lead's
CWD-drift warning, every recorded value was re-measured with the tree pinned absolutely (§7.3) except
two: the AC-CSRB-010 suite run and the probe/mutant runs. Neither was re-executed. Their tree is
established by a positive witness inside the artifact itself — a test name that exists nowhere in the
primary checkout, and go panic stack frames printing the worktree path verbatim (§7.4). That is
attribution evidence, not a fresh execution: it establishes WHERE the recorded run happened, and says
nothing further about whether a fresh run today would reproduce it. For AC-CSRB-010 specifically, a
re-run would in any case be measured against the same lower-bound BEFORE described in G1.

---

## 5. Residual-risk

- **R1 (from G1).** If the pre-card count was in fact higher than 6886 — it must have been, since
  t540 and t563 both added tests — then a silent removal or de-registration of up to 52 pre-existing
  tests would still satisfy `AFTER >= BEFORE`. The `--- FAIL: ` and `^FAIL` zeros do not close this;
  only an exact pre-card baseline would.
- **R2.** The mutant set is the four the plan named. A conversion defect outside those four shapes —
  for example converting in the home-relative branch as well, which none of the four mutates — is not
  covered by an executed mutant; AC-CSRB-004 asserts the home-relative target directly, which is
  narrower than a mutant demonstration.
- **R3.** The doctor half rests on structural argument (identical one-line shape) plus a
  behaviour-unchanged guard, not on a behavioural assertion of the converted target. Until t563's
  seam lands, a doctor-side conversion defect that is invisible on a `'/'`-separator host would not be
  caught here.
- **R4.** `origin/develop` moved to `91d25bc61` after this card's `c72dc1baf` absorb. Nothing in this
  milestone was measured against the merged tree, so a semantic clash with what landed on develop in
  between is possible and belongs to the integration window, not to this run.
- **R5.** The revert-proof is `git status` / `git diff` against `HEAD`, which is byte-level for
  tracked files. It would not detect a stray untracked file left outside the evidence directory; the
  full `git status --porcelain` listing in §2.3 shows only the expected evidence logs, which bounds
  this to what that listing covers.

---

## 6. AC matrix — M3

| AC | Verdict | Evidence |
|---|---|---|
| AC-CSRB-003 | PASS (guard; reorder mutant CAUGHT) | `mutant-reorder.log`, `probe-p3.log`, `ac-010.log` |
| AC-CSRB-004 | PASS (guard; blanket-wrap mutant CAUGHT) | `mutant-blanket-wrap.log`, `probe-p2.log`, `ac-010.log` |
| AC-CSRB-005 | PASS (3 arms green in the full-suite run) | `ac-010.log`, `ac-csrb-010.log` |
| AC-CSRB-006 | PASS (normalized diff empty; only run nonces differ) | `ac-csrb-006-post-m2.log` vs `ac-csrb-006-baseline.log` |
| AC-CSRB-007 | PASS (0 with control 1; all 6 card commits 0) | `ac-csrb-007.log`, `mutant-seam.log` |
| AC-CSRB-008 | PASS (31 paths all allowlisted; probe empty vs non-zero control; PRESERVE 0) | `ac-csrb-008.log` |
| AC-CSRB-009 | PASS (both builds rc=0; limit stated) | `ac-csrb-009.log` |
| AC-CSRB-010 | PASS against a lower-bound BEFORE (see G1) | `ac-csrb-010.log`, `ac-010.log` |

Mutant ledger: **4 injected, 4 caught, 0 missed, 0 committed, 0 surviving in the tree.**

---

## 7. Addendum — tree attribution (lead's CWD hazard warning, and its correction)

The lead first reported the hazard as two checkouts differing by one letter's case, then corrected
the premise: the two spellings are **one directory**, and the real axis is worktree → primary drift.
The prescription was unchanged, so all four items are applied here. **Re-measured in this run, with
every read pinned to an absolute tree path** (`tree-attribution-remeasure.log`).

### 7.1 The three paths, measured

```
$ stat -f '%d:%i  %N' <worktree> /Users/goos/MoAI/moai-adk-go /Users/goos/moai/moai-adk-go
16777231:3137769831  /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t562
16777231:253706617   /Users/goos/MoAI/moai-adk-go
16777231:253706617   /Users/goos/moai/moai-adk-go
```

The two spellings share a dev:inode — one directory, two names, exactly as the correction says. The
worktree is a distinct directory. Their git identities differ, and that difference is the
discriminator every measurement below leans on:

| Path | branch | HEAD |
|---|---|---|
| `.../.claude/worktrees/t562` | `WT-codex-read-inverse` | `757ef601f` (M3 commit; the mutant window ran at `4aa8915ee`) |
| `/Users/goos/MoAI/moai-adk-go` (both spellings) | `main` | `7ad9f8534` |

### 7.2 Where drift would have been silent, measured rather than assumed

```
internal/cli/codex_skills_prune.go                primary=ABSENT   worktree=EXISTS
internal/cli/codex_skills_prune_readback_test.go  primary=ABSENT   worktree=EXISTS
internal/cli/codex_config_path.go                 primary=ABSENT   worktree=EXISTS
internal/cli/doctor_codex.go                      primary=EXISTS   worktree=EXISTS   ← the silent one
```

`doctor_codex.go` is the one file this card measures that exists in **both** trees with **different
content**: `/usr/bin/grep -c 'osStatFn'` gives **0** in the primary and **2** in the worktree. A
drifted provenance-context read would therefore have returned a plausible `0` with no error — the
exact failure shape the lead names. The other reads fail loudly (file absent), so they were never at
risk of quiet corruption.

This cuts the other way as attribution evidence: the recorded values **1** (control, on a file the
primary does not have) and **2** (provenance, where the primary holds 0) are producible only in the
worktree.

### 7.3 Re-measurement results — every value reproduced

| Measurement | Originally recorded | Re-measured with absolute pinning | Match |
|---|---|---|---|
| AC-CSRB-007 delta (`git -C <wt> show c007e5409 --format='' -- doctor_codex.go \| grep -c osStatFn`) | 0 | 0 | ✓ |
| AC-CSRB-007 added-lines-only | 0 | 0 | ✓ |
| AC-CSRB-007 positive control (`codex_skills_prune.go`) | 1 | 1 | ✓ |
| AC-CSRB-007 provenance context (`doctor_codex.go`) | 2 | 2 | ✓ |
| AC-CSRB-007 per-commit sweep (6 card commits) | all 0 | all 0 | ✓ |
| absorb merge `2d1dad058` first-parent | 2 | 2 | ✓ |
| AC-CSRB-008 `CARD_BASE` | `c72dc1baf` | `c72dc1baf` | ✓ |
| AC-CSRB-008 changed-path count | 31 | 31 | ✓ |
| AC-CSRB-008 `codexwiring` probe | empty | 0 lines | ✓ |
| AC-CSRB-008 paths outside `internal/cli/` + `.moai/` | (implied 0) | **0**, measured directly | ✓ |
| AC-CSRB-008 PRESERVE diff `2d1dad058..4aa8915ee` | empty | 0 lines, control 10 | ✓ |
| AC-CSRB-006 guard (`go -C <wt> test …`) | PASS, normalized diff empty | PASS, normalized diff empty | ✓ |
| AC-CSRB-009 native build | rc=0 | rc=0 | ✓ |
| AC-CSRB-009 `GOOS=windows` build | rc=0 | rc=0 | ✓ |
| AC-CSRB-009 `GOOS=windows` vet | rc=0 | rc=0 | ✓ |

The AC-CSRB-008 row "paths outside the two roots = 0" is new: the original record listed all 31
paths and judged them by reading, which is a human comparison. It is now a counted predicate.

### 7.4 Attribution of the two measurements NOT re-executed

**AC-CSRB-010's suite run** (`ac-010.log`, 18,820 lines) was not re-executed — a ~12-minute run —
because a positive witness settles its tree without one:

```
TestJudgeCodexSkillEntry_SeparatorConversion  in PRIMARY tree  = 0 files
                                               in WORKTREE     = 1 file
                                               in ac-010.log   = 2 occurrences
```

The test name does not exist anywhere in the primary checkout, so a drifted suite could not have
emitted it. The log contains it, therefore the log came from this worktree.

**The probe and mutant runs** carry their tree in the recorded output itself — the go panic stack
frames print absolute paths:

```
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t562/internal/cli/codex_skills_prune.go:83
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t562/internal/cli/codex_skills_prune.go:101
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t562/internal/cli/codex_skills_prune.go:76
```

That is the worktree path verbatim, in the artifact, for all three probe sites.

### 7.5 Line-citation provenance (prescription item 3)

Every line citation in this document — `codex_skills_prune.go:76` (classify switch), `:83` (absolute-
branch conversion), `:101` (stat call), and `doctor_codex.go:861` (the seam-mutant injection point) —
is measured in:

> **tree** `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t562` · **branch** `WT-codex-read-inverse`
> · **SHA** `4aa8915ee` (the mutant-window HEAD; M3 added zero production lines, so these line
> numbers are unchanged at the M3 commit `757ef601f`)

A line citation decays like a HEAD reading; it is pinned here rather than left bare.

### 7.6 Why no earlier measurement had drifted

The session's working directory was never changed — no `cd` was issued at any point — and the first
command of the run printed `git rev-parse --show-toplevel` as the worktree. More usefully, every
batch that touched git printed `git rev-parse --short HEAD` as `4aa8915ee` and
`git branch --show-current` as `WT-codex-read-inverse`; a drifted command would have printed
`7ad9f8534` / `main`. That discriminator was present in the record before the warning arrived, and
the re-measurement above confirms it independently rather than resting on it.

**Nothing was quietly kept and nothing was quietly dropped**: every value was re-measured except the
two in §7.4, whose tree is established by a positive witness in the artifact itself, and both
exceptions are named here rather than left implicit.
