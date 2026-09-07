# SPEC-CODEX-SKILL-PATH-READBACK-001 — Acceptance Criteria

Card t562 · Tier M · cycle_type tdd. Every AC is binary-testable. Guards (no-RED-by-design ACs)
carry their discriminating power as named, executed mutants; every zero-count claim carries a
positive control. Line numbers in this file are hints measured 2026-09-08 at `bce6d7e08`
(in-tree) / `e5df637bc` (frozen branch) and are never the evidence itself — symbols are.

**Two-cell discipline.** Each AC records a RED-now cell (what was observed red, on which tree) and a
green-path cell (which milestone flips it, to what output) — or, for guards, an explicit
"green-by-construction" cell plus the executed mutant that demonstrates the test can fail. An AC
whose failure was never observed on a known failing input is not adopted (verification-completeness
§1.1/§2).

---

### AC-CSRB-001 — Absorb gate: the seam is present and the tree builds (REQ-CSRB-001 precondition)

**Given** the card worktree at `WT-codex-read-inverse` **When** M1's absorb merge
(`git merge WT-codex-path-escape`) has landed **Then** `go build ./...` exits 0,
`internal/cli/codex_config_path.go` defines `toConfigPath`, `fromConfigPath`, and
`configPathSeparator`, and `codex_skills_disable.go` contains the publisher conversion (locate by
symbol) — the consumer precondition of the whole card.

- **RED-now**: before the merge, `internal/cli/codex_config_path.go` does not exist in this tree
  (measured: the file is absent at `bce6d7e08`; `git show WT-codex-path-escape:internal/cli/codex_config_path.go`
  is the only access path).
- **Green path**: M1. Evidence: build rc + the symbol grep, recorded under `.moai/reports/t562/`.

### AC-CSRB-002 — Prune stat target is the converted form (behavioral RED; maps REQ-CSRB-001, REQ-CSRB-007)

**Given** `configPathSeparator` overridden to `'\\'` (`t.Cleanup` restore, test non-parallel) and
`osStatFn` replaced by a recorder capturing its argument **When** `judgeCodexSkillEntry` judges an
entry whose `Path` is a HOST-absolute slash path (e.g. `filepath.Join(t.TempDir(), "gone",
"SKILL.md")`, which classifies `codexPathAbsolute` on this host) **Then** the recorder's captured
argument equals `fromConfigPath(declared, '\\')` — the backslash-converted form, not the declared
form.

- **RED-now**: on the absorbed, pre-implementation tree the recorder observes the DECLARED form
  (`statPath = e.Path` verbatim) and the assertion FAILS. Command (single selector — Go applies only
  the last `-run` flag, so a two-flag form would leave the first selector silently dead):
  `go test ./internal/cli/... -run 'TestJudgeCodexSkillEntry_SeparatorConversion' -timeout 1800s -v`.
  The name is a placeholder until M1 authors the test — pin the final test name in this AC's
  evidence record during M1.
- **Green path**: M2's one-line branch conversion flips it to the converted form.
- **Mutant (mandatory)**: remove/bypass the conversion call → the test must FAIL again. Record
  caught/missed under `.moai/reports/t562/`.
- **Boundary (REQ-CSRB-007)**: the test asserts the stat TARGET and the verdict; it performs no
  filesystem deletion.

### AC-CSRB-003 — Classification precedes conversion (ordering guard; maps REQ-CSRB-003)

**Given** `configPathSeparator` overridden to `'\\'` **When** `judgeCodexSkillEntry` judges an entry
declaring `C:/Users/u/SKILL.md` (a slash-form Windows declaration; on darwin `IsAbs` is false and
the declared string carries no backslash) **Then** the verdict's SkipReason is the RELATIVE skip
("relative path — no observed resolution base") and the stat recorder counted zero calls.

- **Why this detects the ordering violation**: the forbidden convert-first order would classify
  `C:\Users\u\SKILL.md` (post-conversion) as `codexPathOddlyFormed` and emit the oddly-formed skip
  reason instead — a distinct string, so the ordering is observable on this host.
- **Cells**: green-by-construction (no conversion exists before M2, so classification trivially runs
  on the declared form); discriminating power = the **reorder mutant** (convert before
  `classifyCodexSkillPath` → test must FAIL), executed and recorded in M3. No RED claimed.

### AC-CSRB-004 — Home-relative stat target is NOT converted (guard; maps REQ-CSRB-002)

**Given** `configPathSeparator` overridden to `'\\'`, `codexUserHomeDir` stubbed to a temp home,
and the recorder in place **When** `judgeCodexSkillEntry` judges an entry declaring `~/x/SKILL.md`
**Then** the recorder's captured argument equals `filepath.Join(home, "x/SKILL.md")` EXACTLY — the
native `filepath.Join` product, unrewritten.

- **Why this detects the forbidden shape**: the blanket-wrap mutant
  (`osStatFn(fromConfigPath(statPath, configPathSeparator))` at the stat site) rewrites the Join
  product on this host — with `'\\'` injected, the product's `/` characters are replaced and the
  captured argument changes.
- **Cells**: green-by-construction; discriminating power = the **blanket-wrap mutant** (must FAIL
  the test), executed and recorded in M3. No RED claimed.

### AC-CSRB-005 — Eligibility gating unchanged and platform-conditional (pins; maps REQ-CSRB-004)

Three arms, one table-driven test (production separators unless stated):

| Arm | Given | When | Then |
|---|---|---|---|
| a | absolute declaration; recorder returns `fs.ErrNotExist` | judged | `Eligible: true` (classification + ErrNotExist — the exact two-condition gate) |
| b | absolute declaration; recorder returns nil | judged | skip, reason "the path resolves" |
| c | backslash declaration `C:\Users\u\SKILL.md`, production separators | judged | `codexPathOddlyFormed` skip, stat count 0 — the non-destructive `/`-separator-host behaviour of §B.3 |

- **Cells**: green-by-construction pins (the gate exists today); discriminating power = any
  regression that widens the gate or converts before classifying flips arm c's reason or arm a/b's
  verdict. The phrase "destructive on all platforms" appears nowhere in this SPEC or its evidence.
- **Boundary**: no arm performs a deletion; verdicts only.

### AC-CSRB-006 — Doctor regression guard (maps REQ-CSRB-001, REQ-CSRB-005, REQ-CSRB-008)

**Given** `CODEX_HOME` pointed at a temp tree holding a `config.toml` fixture whose entry declares
an absolute path to an EXISTING file (plus, in a second fixture, an entry already in native form
and one backslash declaration) **When** `codexStaleSkillFinding` runs **Then** the existing-file
entry is not counted missing, the backslash declaration counts as oddly-formed, and the
relative/oddly-formed/indeterminate counters match the PRE-CHANGE baseline exactly.

- **Derivation order (load-bearing)**: the expected counter values are derived from the PRE-CHANGE
  run (captured during M1's window, before M2 lands) — the guard asserts UNCHANGED behaviour, not
  new behaviour.
- **Cells**: green-by-construction regression guard; no RED claimed (t540 §G gap 5 disposition
  carries forward). The doctor half of REQ-CSRB-001 is additionally evidenced structurally (the
  M2 diff is the same one-line shape as the behaviorally-verified prune side) and at build level
  (AC-CSRB-009). This is the honest ceiling of doctor-side verification until t563's seam exists.
- **Isolation (REQ-CSRB-008)**: fixture under `t.TempDir()`; the real `~/.codex/config.toml` is
  never read or written.

### AC-CSRB-007 — Doctor seam-out pin (maps REQ-CSRB-006)

**Given** the card complete **When**
`/usr/bin/grep -c 'osStatFn' internal/cli/doctor_codex.go` runs **Then** it prints `0`, with the
positive control `/usr/bin/grep -c 'osStatFn' internal/cli/codex_skills_prune.go` printing `>= 1`
in the same verification batch.

- **Why the control**: a zero from a broken pattern is indistinguishable from a true zero; the
  control proves the pattern matches the seam when one exists.
- **Mutant (mandatory)**: add an `osStatFn` declaration to `doctor_codex.go` → the count becomes 1
  and the pin FAILS; run, record, revert. Its presence would also be a DoD failure (t563's scope).

### AC-CSRB-008 — Scope pin via re-derived CARD_BASE (maps REQ-CSRB-006)

**Given** the card's commits on `WT-codex-read-inverse` **When** `CARD_BASE=$(git merge-base
origin/develop HEAD)` is re-derived at read time and `git diff --name-only "$CARD_BASE"..HEAD` runs
**Then** every path is inside the §G allowlist of `plan.md` (this card's prune/doctor/test files +
the four absorbed t540 files + `.moai/` artifacts), and the probe
`git diff --name-only "$CARD_BASE"..HEAD -- internal/codexwiring/skills.go` is empty against the
non-zero control (the full diff listing).

- **Cells**: the probe-empty arm is absence-shaped and carries the non-zero control + a token
  demonstration (touching `skills.go` in a scratch mutant flips the probe non-empty; run, record,
  revert).
- **Note**: the dated anchor `bce6d7e08` (= origin/develop, 2026-09-08) is context only; the AC's
  left edge is the re-derived merge-base, never the pinned SHA.

### AC-CSRB-009 — Cross-platform build, structural only (maps REQ-CSRB-007)

**Given** the card complete **When** `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...`
run **Then** both exit 0.

- **Cells**: build-level verification only. No Windows runtime behaviour is claimed, observed, or
  implied by this AC; every Windows statement in this SPEC is structural inference (§G gap 1).

### AC-CSRB-010 — Package suite with executed-test control (REQ-CSRB-005/006 verification layer)

**Given** the BASE pass count captured pre-flight (before the absorb merge, same command) **When**
`go test ./internal/cli/... -timeout 1800s -v` runs post-M3 **Then** exit code is 0, `--- FAIL: `
count is 0, and the `--- PASS: ` count A satisfies `A >= B AND A > 0` where B is the BASE count —
with the expected delta accounted (t540's absorbed tests + this card's new tests), and any shortfall
below the expected delta reported as "not measurable", never as a pass.

- **Why the control**: a selector matching zero tests and a package that failed to compile both
  present as quiet green; the count is the liveness discriminator.
- **Scope**: `internal/cli` only. The full suite is CI's, on the pushed head.

---

## Edge cases

- **Pure `~` declaration**: expands to the home itself (native) — covered under AC-CSRB-004's
  no-conversion contract.
- **Directory at the declared path**: resolves → skip "the path resolves" (arm b) — unchanged.
- **Stat error that is not `ErrNotExist`** (permission, symlink loop, I/O): indeterminate skip —
  unchanged by this card; not re-tested here (existing REQ-CGP coverage holds).
- **Slash-form Windows declaration on darwin** (`C:/…`): classifies relative on this host — the
  AC-CSRB-003 discriminator; its Windows classification (absolute) is inference I-1.
- **Mixed home-relative form** (`~/x\SKILL.md`): pre-existing residual, out of scope (`spec.md §F/§G`).

## Quality gates

- `go vet` clean on touched packages; `golangci-lint run` clean (0 new findings).
- Scoped package suite per AC-CSRB-010; `internal/cli` `-timeout 1800s` floor respected.
- `GOOS=windows` build per AC-CSRB-009.
- Every mutant record names what it mutated, which AC caught it, and where the record lives
  (`.moai/reports/t562/`).

## Definition of Done

1. All ten ACs PASS, or any OPEN AC names its owner and reason explicitly (an OPEN AC with no named
   owner is a FAIL).
2. The production diff is exactly the two branch conversions (`spec.md §D.2`); nothing else.
3. `osStatFn` count in `doctor_codex.go` is 0 (AC-CSRB-007) — the t563 seam was NOT added.
4. `internal/codexwiring/skills.go`, `codex_config_path.go`, and `codex_skills_disable.go` carry no
   card-authored modifications beyond the absorb merge.
5. No test performed a filesystem deletion; no experiment wrote the real `~/.codex/config.toml`.
6. Evidence under `.moai/reports/t562/`; every commit message contains `t562`.
7. The cross-card notes are traceable: t540's open M3/ACs map to this card's ACs (`spec.md §D.5`);
   t533 layer 2 (AC-CGM-011/012/013) is named as the unblock target (`spec.md §B.4`).
