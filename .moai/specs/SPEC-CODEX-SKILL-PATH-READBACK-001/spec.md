---
id: SPEC-CODEX-SKILL-PATH-READBACK-001
title: "Codex ghost-skill readers consume the t540 path seam — declared config paths convert to host form before stat"
version: "0.1.0"
status: in-progress
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tier: M
depends_on: [SPEC-CODEX-SKILL-PATH-SLASH-001]
related_specs: [SPEC-CODEX-SKILL-PATH-001, SPEC-CODEX-GHOST-SKILLS-MEASURE-001, SPEC-CODEX-GHOST-SKILLS-PRUNE-001]
tags: "codex, skills-config, windows, path-conversion, readback, prune, doctor, t562"
---

# SPEC-CODEX-SKILL-PATH-READBACK-001 — Codex ghost-skill readers consume the t540 path seam

## HISTORY

- 2026-09-08 (plan-phase, v0.1.0) Initial authoring. Card t562, worktree `.claude/worktrees/t562`, branch `WT-codex-read-inverse`, base `bce6d7e08` (= origin/develop at entry; dated anchor, NOT an AC range edge — AC ranges re-derive `CARD_BASE=$(git merge-base origin/develop HEAD)` at read time). Tier M, Class C, cycle_type tdd (`quality.yaml development_mode: tdd`, read this run). This card implements the reader-side half of SPEC-CODEX-SKILL-PATH-SLASH-001 (t540) that t540's run deliberately deferred: its `progress.md §E.3` records `run_status: partial — M1 and M2 complete, M0 and M3 open` with AC-CSPS-004 arms A/A'/C and AC-CSPS-005 OPEN.
- 2026-09-08 (in-place amendment, mid-run, lead-approved) **AC-CSRB-007 re-anchored: absolute file count → card-diff delta.** Trigger: absorb merge `2d1dad058` of origin/develop `c72dc1baf` carrying t563 (SPEC-DOCTOR-STAT-SEAM-001) — `internal/cli/doctor_codex.go` now legitimately holds TWO `osStatFn` occurrences, absorb-provenance, not card-authored. Decision: lead-approved re-anchor — the pin counts what THIS CARD's diff adds (`git show c007e5409 --format='' -- internal/cli/doctor_codex.go | /usr/bin/grep -c 'osStatFn'` → 0, M2 SHA; measured three ways in `.moai/reports/t562/green-csrb-002-m2.md`), never what the file holds. The prior absolute form is recorded here, not silently rewritten: it read "`/usr/bin/grep -c 'osStatFn' internal/cli/doctor_codex.go` → 0". No other AC touched.

## §A. User Story

**As a** MoAI-ADK user on Windows whose `~/.codex/config.toml` carries `[[skills.config]]` paths in the config's forward-slash form (the form t540's publisher now writes), **I want** the ghost-skill readers (`moai clean --codex-skills` and the `moai doctor` stale-path check) to stat the declared path in my host's own separator form, **so that** a healthy registration is not misread as absent — the misread that, on Windows, opens a destructive prune branch against a registration I wrote.

**As a** MoAI-ADK maintainer, **I want** the readers to consume the seam t540 already landed (`toConfigPath` / `fromConfigPath` / `configPathSeparator`) rather than inventing a second conversion, **so that** publisher and readers share one separator contract and the next platform question has exactly one place to be answered in.

## §B. Context and Background

### §B.1 Where this card sits in the t540 family

SPEC-CODEX-SKILL-PATH-SLASH-001 (t540, branch `WT-codex-path-escape`, FROZEN at `e5df637bc`) landed the separator seam and the publisher side only. Measured on the frozen branch this run:

| # | Address (symbol-first; line numbers are hints measured this run) | Fact as read |
|---|---|---|
| M1 | `internal/cli/codex_config_path.go` (frozen branch, new file) | `toConfigPath(p string, sep rune) string`, `fromConfigPath(p string, sep rune) string`, `var configPathSeparator = filepath.Separator`. Identity when `sep == '/'`; `strings.ReplaceAll` otherwise. [HARD] separator-aware — an unconditional backslash replacement is forbidden (its REQ-CSPS-010). |
| M2 | `internal/cli/codex_skills_disable.go` (frozen branch) | `:247` `skillPath = toConfigPath(skillPath, configPathSeparator)`; `:277` comparison on the `toConfigPath` form of both sides. Publisher conversion LANDED. |
| M3 | `internal/cli/codex_skills_prune.go` (in-tree at `bce6d7e08`) | `judgeCodexSkillEntry`: absolute branch sets `statPath = e.Path` verbatim (~:77-78); stat via the injectable seam `osStatFn(statPath)` (~:96). `/usr/bin/grep -n 'fromConfigPath\|osStatFn\|toConfigPath'` on the frozen branch matches `:96` only — **the reader conversion is NOT implemented**. |
| M4 | `internal/cli/doctor_codex.go` (in-tree at `bce6d7e08`) | `codexStaleSkillFinding`: absolute branch sets `statPath = e.Path` verbatim (~:834-835); stat is a DIRECT `os.Stat(statPath)` (~:857) — NOT routed through `osStatFn`. `classifyCodexSkillPath` (~:667) runs `filepath.IsAbs` FIRST and treats any residual `\` as `codexPathOddlyFormed`. `expandCodexHomeRelativePath` (~:686) returns a `filepath.Join` product. Measured: `os.Stat` appears at exactly two sites (~:459, ~:857); `osStatFn` appears ZERO times in this file. Historical at that tree — the absorbed t563 seam has since moved the file's count to 2 (see HISTORY 2026-09-08 amendment); the row stands as a dated measurement, not current state. |
| M5 | `internal/cli/codex_skills_prune_test.go` | The seam-override discipline is established: `:10-11` "Tests here reassign package-level seams … MUST NOT call t.Parallel()"; `:55-57` `orig := osStatFn; t.Cleanup(func(){ osStatFn = orig }); osStatFn = …`. Fixtures under `t.TempDir()`; `"~"` expansion rides `codexUserHomeDir`. |
| M6 | t540 `progress.md §E.2/§E.3` (frozen branch) | Run scope was **M1 + M2 only** (operator decision: "M1+M2 first, landing deferred"). Commits `2d66ebad4` (M1), `dc71e8ef9` (M2). AC-CSPS-004 arms A/A'/C and AC-CSPS-005 recorded OPEN — "codex_skills_prune.go / doctor_codex.go untouched". |
| M7 | `.moai/specs/SPEC-CODEX-GHOST-SKILLS-MEASURE-001/acceptance.md:181,198,208` (in-tree) | t533's layer-2 ACs — AC-CGM-011 (prune dry-run), AC-CGM-012 (`--force` write + backup), AC-CGM-013 (handoff self-evidence) — exist and gate on prune execution. |

So this card IS t540's deferred M3, promoted to its own card. It is the CONSUMER of the seam; **re-creating the seam is forbidden** (t540 owns `codex_config_path.go`; this card does not modify it).

### §B.2 The defect, precisely

After the absorb merge (the run-phase absorb — `plan.md` M1), a Windows user's config declares paths in the config's forward-slash form — `path = "C:/Users/u/.agents/skills/foo"` — because that is what t540's publisher writes. Both readers then classify that declaration on its declared form and stat it raw:

- `classifyCodexSkillPath("C:/Users/…")` on Windows returns `codexPathAbsolute` (`filepath.IsAbs` accepts `/` on Windows — structural inference from Go's contract; see §B.3), so the absolute branch hands the slash-form string to stat **unconverted**. On Windows this still often resolves (the OS tolerates `/`), but the readers' contract and the publisher's output now disagree about form, and any host or resolver that does not tolerate the mismatch reads a healthy registration as absent.
- When the stat does report `fs.ErrNotExist`, the prune verb's destructive branch opens against a registration that may be perfectly healthy — reachable only under the platform-conditional truth of §B.3, carried by REQ-CSRB-004.

**INFERRED (I2 lineage — not observed on a Windows runtime in this tree):** the Windows classification and the Windows stat tolerance are structural inferences from Go's `filepath` contract. What IS measurable on this darwin host is the **conversion half** — that the stat target handed to the stat call is the `fromConfigPath`-converted form of the declaration — because t540's seam takes the separator as a parameter. That measurable/structural split is the axis this card's test strategy is built on (§D.4).

### §B.3 MEASURED vs INFERRED

**MEASURED — read directly from source this run, at `bce6d7e08` (in-tree) and `e5df637bc` (frozen branch):** the M1-M7 table above.

**INFERRED — not observed on a Windows runtime:**

| # | From | Inference |
|---|---|---|
| I-1 | `classifyCodexSkillPath` + Go's `filepath.IsAbs` contract | a `C:/…` (or `C:\…`) declaration classifies `codexPathAbsolute` on Windows; the same string on a `/`-separator host classifies `codexPathRelative` (slash form) or `codexPathOddlyFormed` (backslash form) |
| I-2 | M3/M4 + I-1 | the unconverted stat target and any resulting eligibility are Windows-reachable behaviours this host cannot execute |

**[HARD] Platform-conditional destructiveness — the exact truth, not the broad one.** `Eligible: true` in `judgeCodexSkillEntry` requires BOTH a classification of `codexPathAbsolute` or `codexPathHomeRelative` AND a stat error of `fs.ErrNotExist` (M3 addresses ~:76-94 and ~:102). A Windows-form declaration classifies `absolute` on Windows → the destructive branch is reachable there; the same declaration on a `/`-separator host classifies `oddly-formed` or `relative` → non-destructive skip. The phrase "destructive on all platforms" is FALSE and does not appear in this SPEC; the platform-conditional statement is the one carried by REQ-CSRB-004.

### §B.4 Card ordering — this card unblocks t533 layer 2

t533 (SPEC-CODEX-GHOST-SKILLS-MEASURE-001) layer 2 — its prune-execution ACs AC-CGM-011/012/013 (M7) — stays OPEN until this card lands. The dependency chain is: t540 (seam + publisher) → **t562 (this card, reader readback)** → t533 layer 2 (prune execution against measured ghosts). This card does not execute `prune` against any real config (§F); it makes the reader half correct so that a later execution is safe to reason about.

Run-gate note on `depends_on`: `SPEC-CODEX-SKILL-PATH-SLASH-001` is absent from this tree until M1's absorb merge, so a run-phase Depends_on pre-flight reads it unfulfilled. The override path is the M1 absorb merge itself — the merge IS the fulfilment event; record it as satisfied at M1 completion, do not treat the pre-flight reading as a blocker.

### §B.5 Why the conversion site is the absolute branch, not the stat call

The two lead constraints that fix this design, applied to the source:

1. **Classification precedes conversion** (`classifyCodexSkillPath` runs `filepath.IsAbs` on the DECLARED string first). Converting before classification would feed `classifyCodexSkillPath` a native-form string and misclassify legitimate Windows declarations — on a `/`-separator host a `C:\…` declaration must classify `codexPathOddlyFormed` (non-destructive skip), and only the declared form produces that.
2. **The home-relative branch is already native.** `expandCodexHomeRelativePath` returns a `filepath.Join` product (M4); `filepath.Join`/`Clean` normalize to the host separator. Routing that product through `fromConfigPath` is a category error — converting a string that never carried the config form — and is FORBIDDEN.

Together these dictate: the conversion applies at the branch where `statPath` is set from the declaration verbatim — the `codexPathAbsolute` branch — in BOTH readers. A blanket wrap at the stat call site (`osStatFn(fromConfigPath(statPath, …))`) would satisfy constraint 1 but violate constraint 2, and is named as a mutant this card's tests must catch (AC-CSRB-004).

## §C. Requirements (GEARS)

- **REQ-CSRB-001** (Event-driven) — When a stat-ing reader resolves a declared `[[skills.config]]` path whose classification is `codexPathAbsolute`, the reader shall pass `fromConfigPath(e.Path, configPathSeparator)` — the host-form conversion of the declared path — as the stat target, in both `judgeCodexSkillEntry` (`internal/cli/codex_skills_prune.go`) and `codexStaleSkillFinding` (`internal/cli/doctor_codex.go`).
- **REQ-CSRB-002** (Unwanted) — The home-relative branch's expanded stat target (the `expandCodexHomeRelativePath` `filepath.Join` product, and the bare-home product for a `~` declaration) shall NOT be routed through `fromConfigPath`.
- **REQ-CSRB-003** (Unwanted, ordering) — The classification (`classifyCodexSkillPath`) shall run on the DECLARED form in every reader, and no conversion shall run before or inside the classification call.
- **REQ-CSRB-004** (Unwanted, safety) — The destructive branch of `judgeCodexSkillEntry` shall open only under classification ∈ {`codexPathAbsolute`, `codexPathHomeRelative`} AND stat error `fs.ErrNotExist`; on a host whose path separator is `/`, a backslash-bearing non-absolute declaration shall classify `codexPathOddlyFormed` and be skipped non-destructively. This gating is unchanged by this card — the card only corrects the stat target fed into it.
- **REQ-CSRB-005** (Ubiquitous) — `codexStaleSkillFinding`'s outputs for already-native-form inputs shall be unchanged by this card (regression guard), and the `inspectSkillMirror` stat site (~:459, which stats mirror-directory walk results and never reads a config declaration) shall be untouched.
- **REQ-CSRB-006** (Ubiquitous, scope pin) — `internal/codexwiring/skills.go`, `internal/cli/codex_config_path.go`, and `internal/cli/codex_skills_disable.go` shall not be modified by this card's own commits (the first two are t540's read-only surface; the last is t540's landed publisher — absorbed, not edited). No stat injection seam (`osStatFn`) shall be ADDED BY THIS CARD's own commits to `internal/cli/doctor_codex.go`; the file's present `osStatFn` occurrences are absorb-provenance (t563 / SPEC-DOCTOR-STAT-SEAM-001, via absorb `2d1dad058` ← `c72dc1baf`), not card-authored (re-anchored 2026-09-08, see HISTORY).
- **REQ-CSRB-007** (Ubiquitous, verification honesty) — Windows behaviour shall be asserted at build level (`GOOS=windows` compile) and structurally (source reading) only; no AC of this SPEC shall claim a Windows runtime observation, and no test of this card shall perform an actual filesystem deletion.
- **REQ-CSRB-008** (Event-driven, isolation) — When any experiment or test writes a Codex config file, it shall run under an isolated `CODEX_HOME` (or `t.TempDir()` fixtures through the established seams); the real `~/.codex/config.toml` shall never be written.

### §C.1 REQ → AC traceability

Full Given-When-Then bodies live in `acceptance.md`; this table is the coverage map only.

| REQ | Covering AC |
|---|---|
| REQ-CSRB-001 | AC-CSRB-002 (prune, behavioral), AC-CSRB-006 (doctor, structural + guard) |
| REQ-CSRB-002 | AC-CSRB-004 |
| REQ-CSRB-003 | AC-CSRB-003 |
| REQ-CSRB-004 | AC-CSRB-005 |
| REQ-CSRB-005 | AC-CSRB-006 |
| REQ-CSRB-006 | AC-CSRB-007, AC-CSRB-008 |
| REQ-CSRB-007 | AC-CSRB-009, AC-CSRB-002 (eligibility arm — classification asserted, no deletion performed) |
| REQ-CSRB-008 | AC-CSRB-006 (fixture isolation), AC-CSRB-002/004/005 (all fixtures under `t.TempDir()`) |

## §D. Decisions and Verdicts

### §D.1 SPEC-ID choice

`SPEC-CODEX-SKILL-PATH-READBACK-001` — it names the card's act (reading the declared config form BACK into host form before stat) and sits in the t540 family namespace (`…-PATH-001` doctor path shapes → `…-PATH-SLASH-001` publisher/seam → `…-PATH-READBACK-001` reader readback). Uniqueness measured this run: `ls .moai/specs/ | grep READBACK` → 0, with a working positive control on the same pattern (`CODEX-SKILL-PATH` → `SPEC-CODEX-SKILL-PATH-001`). ID regex self-check: PASS (executed this run, verbatim in the authoring record).

### §D.2 Conversion site — absolute branch only (decision, with its mutant)

Per §B.5: `statPath = fromConfigPath(e.Path, configPathSeparator)` inside the `codexPathAbsolute` case of BOTH readers; the `codexPathHomeRelative` case is left exactly as-is. The named mutant — a blanket `fromConfigPath` wrap at the stat call site — is behaviorally detectable on this host: with `configPathSeparator` overridden to `'\\'`, a darwin `filepath.Join` product contains `/`, so the wrap rewrites it and AC-CSRB-004 fails. That the mutant is caught on the test host is what makes the design decision enforced rather than advisory.

### §D.3 The two read sides are not test-symmetric — the asymmetry is a design input, not an accident

- **Prune side** (`judgeCodexSkillEntry`): stat rides the `osStatFn` injection seam (M3), so a recorder can observe the stat target. This side carries the behavioral RED test and its bypass mutant.
- **Doctor side** (`codexStaleSkillFinding`): stat was a DIRECT `os.Stat` when this card's plan was authored (M4, measured at `bce6d7e08`; `osStatFn` count in `doctor_codex.go` = 0 — a dated fact, since moved by the absorbed t563 seam). On that basis no recorder was assumed there, and this card's doctor-side verification stayed guard-based. **A card-authored seam at ~:857 to manufacture symmetry is FORBIDDEN** — t563 owns that seam, and it entered this branch by absorption, not by this card (AC-CSRB-007 pins the CARD-DIFF DELTA at zero with a positive control; re-anchored 2026-09-08, see HISTORY).

Consequence, stated rather than hidden: the doctor half's conversion is verified **structurally** (source reading: the same one-line conversion shape as the verified prune side) + **build-level** (`GOOS=windows`) + by an **output-level regression guard** (REQ-CSRB-005). It has no behavioral test on this host, and this SPEC claims none. This inherits t540's §G gap 5 disposition and hands the seam question to t563 unchanged.

### §D.4 Test strategy — recorder RED on the prune side, guards elsewhere

- **Behavioral RED (the only one):** AC-CSRB-002 — recorder via `osStatFn`, `configPathSeparator` overridden to `'\\'` (`t.Cleanup` restore, test non-parallel — M5's discipline, extended to the new seam), entry declaring a HOST-absolute slash path so it classifies `codexPathAbsolute` on darwin (`IsAbs` true). Before the change the recorder observes the declared form (RED); after, the converted form (GREEN). Mutant: bypass/remove the conversion call → test FAILS (the test catches the defect class, in the REQ-CSPS-010 mutant style).
- **Ordering guard:** AC-CSRB-003 — a slash-form Windows declaration (`C:/Users/…`) with the separator overridden classifies on the DECLARED form and skips as `codexPathRelative` with stat never called; the forbidden convert-first order would produce the `codexPathOddlyFormed` skip reason instead. The two skip-reason strings differ, so the ordering is darwin-observable. No RED claimed (green before by construction); its mutant is the reorder.
- **No-double-conversion guard:** AC-CSRB-004 — home-relative declaration with a stubbed `codexUserHomeDir`; the recorder must observe the `filepath.Join` product EXACTLY. Mutant: the §D.2 blanket wrap.
- **Platform-conditional safety pins:** AC-CSRB-005 — backslash declaration on this `/`-separator host → `codexPathOddlyFormed` skip, stat count 0; absolute declaration + recorder `ErrNotExist` → `Eligible: true`; stat resolves → skip "the path resolves". These pin REQ-CSRB-004's exact gating (the platform-conditional truth of §B.3).
- **Doctor regression guard:** AC-CSRB-006 — `CODEX_HOME`-scoped fixture declaring an absolute path to an EXISTING file; the finding must not count it missing and the classification counters must be unchanged. No RED claimed; the pre-change pass IS the baseline the post-change pass is compared against.
- **Build-level Windows:** AC-CSRB-009 — `GOOS=windows GOARCH=amd64 go build ./...` rc=0. Structural, never runtime-observed (REQ-CSRB-007).

### §D.5 Relation to the predecessor's open ACs — traceability, not closure

t540's `progress.md §E.3` holds AC-CSPS-004 arms A/A'/C and AC-CSPS-005 as OPEN pending its M3. This card's run produces exactly that M3's evidence. **Closing those ACs in the predecessor SPEC is not this card's act** — t540's `progress.md` is owned by its own lane's manager-develop, and this card does not edit frozen-branch artifacts. This SPEC records the mapping (its AC-CSRB-002/004/005/006 correspond to AC-CSPS-004 arms A/B'/C-shaped evidence and AC-CSPS-005) so the shared lead can make the cross-card bookkeeping decision on read evidence, not on memory.

## §E. Constraints

- **Hard-to-reverse writes.** `~/.codex/config.toml` is never written by this card's tests or experiments; fixtures ride `t.TempDir()` + `CODEX_HOME` + the `codexUserHomeDir` seam (REQ-CSRB-008; the SPEC-CODEX-SKILLCONFIG-SHAPE-001 harness).
- **No actual deletion is reproduced.** Tests assert eligibility classification and stat-target conversion; the destructive prune branch is never fired against a filesystem in this card's tests (REQ-CSRB-007). This is a deliberate non-goal, stated as a boundary: the card makes the read half correct; it does not demonstrate a deletion.
- **Verification is scoped.** Touched packages only (`internal/cli`); never `go test ./...` locally. `internal/cli` runs with `-timeout 1800s` (t540's measured floor: a lane measured 702.5s on this package, 2026-09-07). Full-suite judgment belongs to CI.
- **Absence claims use `/usr/bin/grep`** (the shell's `grep` is a ugrep wrapper that skips silently), and every zero-count claim carries a positive control on a sibling file.
- **Line numbers decay.** Addresses in this SPEC are symbol-first; line numbers are hints measured at `bce6d7e08` / `e5df637bc` on 2026-09-08 and are never cited as evidence by themselves.
- **No time predictions.** Priority labels and phase ordering only.
- **Every commit message on this branch contains `t562`.** Evidence path: `.moai/reports/t562/`.
- **Worktree discipline.** All work in `.claude/worktrees/t562` on `WT-codex-read-inverse`; the frozen branch `WT-codex-path-escape` is read-only (`git show`), absorbed by merge at M1 and never written.

## §F. Exclusions — what this SPEC does NOT build

This section states what is out of scope for this SPEC.

### Out of Scope — the seam itself

- `internal/cli/codex_config_path.go` is not created, modified, or extended by this card's own commits. It arrives via the absorb merge and is consumed as-is. Re-creating `toConfigPath`/`fromConfigPath`/`configPathSeparator` anywhere is forbidden.
- The publisher side (`codex_skills_disable.go` conversion + comparison) is t540's landed work; this card edits neither.

### Out of Scope — the doctor stat seam (card t563's scope)

- This card's own commits introduce no `osStatFn` (or any other injection seam) into `internal/cli/doctor_codex.go`. AC-CSRB-007 pins the card-diff delta at zero with a positive control (re-anchored 2026-09-08: t563's seam IS in the file by absorption — the pin counts what THIS CARD adds, not what the file holds; see HISTORY).
- Making `classifyCodexSkillPath` separator-injectable. Its `filepath.IsAbs` behaviour stays GOOS-fixed; the Windows half of classification remains inference I-1 (inherited from t540 §F, unchanged).

### Out of Scope — prune execution and t533 layer 2

- This card does not execute `prune` against any real config, dry-run or forced. t533's AC-CGM-011/012/013 own prune execution and stay OPEN until this card lands (§B.4) — this card only unblocks them.
- No actual filesystem deletion is performed or reproduced by any test of this card.

### Out of Scope — classification semantics and pathological forms

- `classifyCodexSkillPath`'s classification outcomes are not changed. The pre-existing mixed-form edge (a declaration like `~/x\SKILL.md` — forward slash after `~`, backslashes later — classifies home-relative and keeps a literal-backslash component in its expansion on a `/`-separator host) is pre-existing behaviour, not introduced or repaired here; recorded as residual-risk.
- The `~user` (another user's home) form stays `codexPathOddlyFormed`.

### Out of Scope — parser, exit codes, duplicate collapse

- `internal/codexwiring/skills.go` is not modified (t540 REQ-CSPS-007 carries forward; AC-CSRB-008 re-pins it).
- No `skip` branch's exit code is changed (t502 F5 remains a separate policy card).
- Collapsing duplicate entries stays with `moai clean --codex-skills`'s existing surface; the 49 ghost entries measured by t533 are not counted, re-counted, or removed here.

### Out of Scope — new surfaces

- No new CLI verb, flag, config key, or finding text. `moai doctor`'s stale-path finding wording is unchanged; only its stat target's form is corrected.

## §G. Gaps — what was NOT observed

1. **Windows execution has NOT been observed.** Every Windows-behaviour statement in this SPEC (I-1, I-2) is structural inference from Go's `filepath` contract plus source reading. The conversion half is measurable on darwin only through t540's separator seam (the reason the seam exists); the classification half is not measurable here at all.
2. **The doctor half's conversion has no behavioral test on this host.** `codexStaleSkillFinding` stats directly (M4), so no recorder exists; AC-CSRB-006 is a regression guard and AC's evidence for the doctor conversion is structural + build-level (§D.3). The seam that would make it measurable is t563's scope, deliberately not taken here.
3. **The destructive branch is not reproduced.** No test fires an actual deletion; eligibility is asserted on the verdict object, not on filesystem state (REQ-CSRB-007 boundary).
4. **`classifyCodexSkillPath`'s Windows behaviour is not measurable on this host** — inherited unchanged from t540 §G gap 4.
5. **The mixed home-relative form residual** (`~/x\SKILL.md` shape, §F) is read from source and reasoned about, not fixture-tested here; it is pre-existing behaviour outside this card's classification scope.

## §H. Cross-references

- SPEC-CODEX-SKILL-PATH-SLASH-001 (t540, frozen branch `WT-codex-path-escape` @ `e5df637bc`) — the seam owner and predecessor; its `progress.md §E.3` records the open M3 this card implements
- `.moai/reports/t540/reader-census.md` (frozen branch) — the reader census and its 2026-09-08 platform-conditionality corrections, which this SPEC's §B.3 carries forward
- SPEC-CODEX-SKILL-PATH-001 — the doctor path-shape semantics (`classifyCodexSkillPath` family) this card consumes unchanged
- SPEC-CODEX-GHOST-SKILLS-PRUNE-001 — the prune verb's REQ-CGP-022 disqualification-beats-eligibility contract, unchanged here
- SPEC-CODEX-GHOST-SKILLS-MEASURE-001 (t533) — its layer-2 ACs AC-CGM-011/012/013 wait on this card (§B.4)
- Card t563 — owner of the doctor stat seam this card explicitly excludes (§F)
