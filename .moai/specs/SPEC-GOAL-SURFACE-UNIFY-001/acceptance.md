---
id: SPEC-GOAL-SURFACE-UNIFY-001
title: Unify the goal surface on /moai goal and relocate goal presentation to the Implementation Kickoff Approval gate
version: 1.3.0
status: completed
created: 2026-07-25
updated: 2026-07-27
author: manager-spec
priority: HIGH
phase: "v3.1.0"
module: doctrine
lifecycle: spec-anchored
tags: "goal, doctrine, session-handoff, slash-command, template-mirror"
tier: L
---

## §A Baseline Provenance

Every judgment command below was **executed** in the worktree at `origin/main` = `e306e21a9`, branch `feat/SPEC-GOAL-SURFACE-UNIFY-001`, working tree clean (0 dirty files). The `Recorded baseline` column carries the verbatim observed output. Every AC **FAILS** at this baseline and PASSES only after its milestone lands.

Three conventions, each chosen after running the alternatives and comparing:

- **Two detectors, deliberately.** Most ACs use the **open form** `` `/goal `` (backtick + `/goal`, no closing backtick) — chosen over the closed `` `/goal` `` because it also catches `` `/goal ac_converge` `` and `` `/goal`-turn ``. AC-GSU-002 and AC-GSU-031 instead use the **union form** `/goal([^a-zA-Z0-9_./-]|$)`, which places no constraint on the preceding character and so catches unbackticked references the open form cannot see. Counts under the two detectors are **not comparable** (`goal-directive.md`: 64 open-form vs 69 union). Where an AC's baseline changed at audit iteration 1, both figures are recorded.
  - Why the union form is not used everywhere: it is broader, so on any file containing a retention surface it must be paired with a section-scope or file-scope carve-out. AC-GSU-002 pairs it with `awk` section scope; AC-GSU-031 is scoped to five files that carry no retention surface. Applying it unscoped would make a sweep AC satisfiable only by deleting retained content.
  - Right-boundary class: `[^a-zA-Z0-9_./-]` excludes `/`, `.`, `-`, so `.moai/state/goal/`, `internal/goal/…`, and `stop-goal` do **not** match. Verified against the five AC-GSU-031 files: `grep -cF 'internal/goal'` → `0` in each.
- **Occurrence vs line counts** — `grep -o … | wc -l` (occurrences) is used throughout, never bare `grep -c` (lines). The two differ materially here: `session-handoff.md` has 12 *lines* but 18 *occurrences* of `post-paste`.
- **Retention-tested before recording** — every sweep-style AC was tested against all three rows of the `plan.md` §A.2 retention register before its baseline was recorded. An AC satisfiable by deleting a retained surface was rewritten (AC-GSU-028, rewritten at iteration 1, then moved to `SPEC-GOAL-DOCS-RETIRE-001` at iteration 2).

Portability note: commands use only POSIX-compatible `awk` / `sed` / `grep`. The macOS `awk` has no 3-arg `match()`, so the self-check count extraction (AC-GSU-008) uses `grep -o` + `sed -n` range extraction instead.

Shell prelude assumed by the commands below (run from the worktree root):

```bash
G=.claude/rules/moai/workflow/goal-directive.md
SH=.claude/rules/moai/workflow/session-handoff.md
EX=.claude/rules/moai/workflow/session-handoff-examples.md
OS=.claude/output-styles/moai/moai.md
GW=.claude/skills/moai/workflows/goal.md
NIM=.claude/rules/moai/workflow/native-invocation-model.md
```

---

## §B AC Matrix — 30 criteria

### M1 — `goal-directive.md` rewrite in place

**AC-GSU-001** — The rule's title line names `/moai goal` as the doctrine subject.

```bash
head -1 "$G" | grep -c '/moai goal'
```

- Recorded baseline: `0` (title reads `# Goal Directive (\`/goal\`) — Autonomous Continuation`)
- Target: `1`

**AC-GSU-002** — Every residual native-`/goal` reference in the file is confined to the retained prohibition section; none survives anywhere else. Section-scoped, so it cannot be satisfied by deleting the prohibition rationale.

**Rescoped at audit iteration 1 (D5).** The original pattern `` /`\/goal/ `` required a backtick immediately before `/goal`, so it could not see `goal-directive.md:86`'s `` `claude -p "/goal <condition>"` `` (preceded by `"`). The detector is now the **union form** `/goal([^a-zA-Z0-9_./-]|$)` — no constraint on the preceding character, so it matches backticked and unbackticked alike. The right-boundary class excludes `/`, `.`, `-`, and word characters, so path fragments (`.moai/state/goal/`, `internal/goal`) and `stop-goal` do not match.

```bash
awk '/^## /{sec=$0} {n=gsub(/\/goal([^a-zA-Z0-9_.\/-]|$)/,"&"); if(n>0 && sec !~ /Native .\/goal. Prohibition/) c+=n} END{print c+0}' "$G"
```

- Recorded baseline: `69` (the union form; the superseded backtick-only form read `33`)
- Target: `0`
- Retention test: the prohibition section is excluded by `awk` section scope, so retention register row 1 is safe. Verified by construction — the guard clause is the section name itself.

**AC-GSU-003** — The retained prohibition section exists exactly once (not zero — the rationale must survive; not twice — one home only).

```bash
grep -c '^## Native `/goal` Prohibition$' "$G"
```

- Recorded baseline: `0`
- Target: `1`

**AC-GSU-004** — Compound: the W2 codification home exists as a named sub-section **and** that section states the arming-never-authorizes invariant. Strengthened at audit iteration 2 (S-new-1): the heading-only form left REQ-GSU-011 green even if the invariant sentence were omitted.

```bash
grep -c '^### Goal-Presentation Timing$' "$G"
sed -n '/^### Goal-Presentation Timing$/,/^###/p' "$G" | grep -ciE 'does not authorize|never authorizes'
```

- Recorded baseline: `0` / `0`
- Target: `1` / `>= 1`

**AC-GSU-005** — The rejected composite alternative is recorded by name (REQ-GSU-012).

```bash
grep -cF 'moai goal --run' "$G"
```

- Recorded baseline: `0`
- Target: `>= 1`

### M2 — Two-step-mechanism removal + Block 5 gap closure

**AC-GSU-006** — The `post-paste` mechanism token is eliminated from all three M2 surfaces. Case-insensitive (the token appears as both `Post-Paste` in headings and `post-paste` in prose).

```bash
for f in "$SH" "$EX" "$OS"; do grep -oiF 'post-paste' "$f" | wc -l | tr -d ' '; done
```

- Recorded baseline: `18` / `6` / `4` (sum 28)
- Target: `0` / `0` / `0`

**AC-GSU-007** — The Localization Table instruction-line row is removed from all three surfaces.

```bash
for f in "$SH" "$EX" "$OS"; do grep -ciF 'Post-paste /goal instruction line' "$f"; done
```

- Recorded baseline: `1` / `1` / `1`
- Target: `0` / `0` / `0`

**AC-GSU-008** — Compound: the `session-handoff.md` pre-emit self-check no longer carries the retired follow-up-block item, **and** its stated item count equals its actual item count (REQ-GSU-007). The count-consistency half already holds at baseline; the compound fails because the retired item is still present, so the AC is falsifiable and detects a stale count after removal.

```bash
STATED=$(grep -o 'Pre-emit self-check (paste-ready budget) — [0-9]* items' "$SH" | grep -o '[0-9]*' | head -1)
ACTUAL=$(sed -n '/^### Pre-emit self-check (paste-ready budget)/,/^### /p' "$SH" | grep -c '^- \[ \]')
RETIRED=$(sed -n '/^### Pre-emit self-check (paste-ready budget)/,/^### /p' "$SH" | grep -ciF 'Post-paste')
echo "stated=$STATED actual=$ACTUAL retired=$RETIRED"
```

- Recorded baseline: `stated=10 actual=10 retired=1`
- Target: `stated=9 actual=9 retired=0`

**AC-GSU-009** — The arm-only property is stated in `session-handoff.md`, closing the B6 rule gap (REQ-GSU-008).

```bash
grep -ciF 'arm-only' "$SH"
```

- Recorded baseline: `0`
- Target: `>= 1`

**AC-GSU-010** — `session-handoff.md` forward-references the progression-mode axis, making the W2 relationship reachable from the handoff side (half of the bidirectional linkage; the other half is AC-GSU-014).

```bash
grep -cF 'Progression Mode' "$SH"
```

- Recorded baseline: `0`
- Target: `>= 1`

**AC-GSU-011** — The `moai.md` §8 render-surface self-check stated count is updated after the item removal.

```bash
grep -o 'Pre-emit self-check ([0-9]* items)' "$OS" | head -1
```

- Recorded baseline: `Pre-emit self-check (12 items)`
- Target: `Pre-emit self-check (11 items)`

### M3 — W2 wiring + retention semantics

**AC-GSU-012** — Compound: `workflows/goal.md` no longer cross-references the removed section, **and** points at the surviving Block 5 rule instead. The second half prevents satisfying this by deleting the cross-reference outright.

```bash
grep -cF 'Post-Paste' "$GW"; grep -cF 'Canonical Format' "$GW"
```

- Recorded baseline: `1` / `0`
- Target: `0` / `>= 1`

**AC-GSU-013** — Compound over-deletion guard: `native-invocation-model.md` retains its HUMAN-ONLY classification, **and** gains the prohibition cross-reference. A blind sweep that stripped the classification would drive the first number down and fail.

```bash
grep -oF 'HUMAN-ONLY' "$NIM" | wc -l | tr -d ' '
grep -cF 'Native `/goal` Prohibition' "$NIM"
```

- Recorded baseline: `17` / `0`
- Target: `>= 17` / `1`

**AC-GSU-014** — The back-reference lives **inside** `workflows/goal.md` § Progression Mode, not merely somewhere in the file. Section-scoped, so a stray mention in § Cross-references does not satisfy it (that mention already exists at baseline, and the AC still reads 0).

```bash
sed -n '/^## Progression Mode/,/^## Safety Invariants/p' "$GW" | grep -cF 'session-handoff'
```

- Recorded baseline: `0`
- Target: `>= 1`

### M4 — Emission-path sweep

**AC-GSU-015** — All 9 M4-owned emission-path files carry zero native-`/goal`. The three retained surfaces (`plan.md` §A.2) are deliberately excluded from this file list.

```bash
grep -ohF '`/goal' \
  .claude/skills/moai/workflows/run.md \
  .claude/skills/moai/workflows/harness-builder.md \
  .claude/skills/moai/workflows/harness-build-entry.md \
  .claude/skills/moai/workflows/moai.md \
  .claude/skills/moai/SKILL.md \
  .claude/rules/moai/workflow/orchestration-mode-selection.md \
  .claude/rules/moai/workflow/dynamic-workflows.md \
  CLAUDE.md | wc -l | tr -d ' '
```

- Recorded baseline: `38`
- Target: `0`

### M5 — W3 wrapper

**AC-GSU-016** — **Reachability, not text presence.** The new wrapper is actually present in the template source tree and passes the thin-command audit. `TestCommandsThinPattern` walks the embedded template FS at test-binary compile time (`//go:embed all:templates` re-embeds from the source tree on every `go test`, independent of whether `make build` was separately run); the subtest is named after the path, so a match proves the file both exists in the source tree and passes the thin-command audit.

```bash
go test -run TestCommandsThinPattern ./internal/template/ -v 2>&1 | grep -c 'commands/moai/goal.md.tmpl'
```

- Recorded baseline: `0`
- Target: `>= 1`
- Control observation (proves the detector works): the same command with `run.md.tmpl` substituted returns `2` at baseline (a `=== RUN` line and a `--- PASS` line).

**AC-GSU-017** — The wrapper exists in both trees and the wrapper count rises by exactly one on each side.

```bash
ls .claude/commands/moai/*.md | wc -l | tr -d ' '
ls internal/template/templates/.claude/commands/moai/*.md.tmpl | wc -l | tr -d ' '
```

- Recorded baseline: `14` / `14`
- Target: `15` / `15`

**AC-GSU-018** — Compound: the local wrapper routes to the `goal` subcommand **and** its `argument-hint` advertises only delivered verbs. Strengthened at audit iteration 2 (S-new-1): the routing-line-only form left REQ-GSU-015 green even if `argument-hint` advertised the undelivered `resume` verb.

```bash
grep -cF 'arguments: goal' .claude/commands/moai/goal.md
grep '^argument-hint:' .claude/commands/moai/goal.md | grep -c 'resume'
```

- Recorded baseline: `ugrep: warning: .claude/commands/moai/goal.md: No such file or directory` (non-zero exit — file absent) for both
- Target: `1` / `0`

### M6 — Template mirror sweep

**AC-GSU-019** — Compound parity: all 14 `.claude/**` pairs are byte-identical **and** the template mirrors carry zero retired-mechanism token. Byte-parity alone already holds at baseline, so the compound is what makes this falsifiable — it verifies parity was *preserved across* the edit rather than merely observed before it.

```bash
SAME=0; DIFF=0; LEAK=0
for f in .claude/output-styles/moai/moai.md \
         .claude/rules/moai/workflow/dynamic-workflows.md \
         .claude/rules/moai/workflow/goal-directive.md \
         .claude/rules/moai/workflow/native-invocation-model.md \
         .claude/rules/moai/workflow/orchestration-mode-selection.md \
         .claude/rules/moai/workflow/session-handoff-examples.md \
         .claude/rules/moai/workflow/session-handoff.md \
         .claude/skills/moai/SKILL.md \
         .claude/skills/moai/workflows/goal.md \
         .claude/skills/moai/workflows/harness-build-entry.md \
         .claude/skills/moai/workflows/harness-builder.md \
         .claude/skills/moai/workflows/moai.md \
         .claude/skills/moai/workflows/run.md \
         .claude/skills/moai-meta-harness/SKILL.md; do
  t="internal/template/templates/$f"
  if diff -q "$f" "$t" >/dev/null 2>&1; then SAME=$((SAME+1)); else DIFF=$((DIFF+1)); fi
  n=$(grep -oiF 'post-paste' "$t" 2>/dev/null | wc -l | tr -d ' '); LEAK=$((LEAK+n))
done
echo "same=$SAME diff=$DIFF post_paste_in_templates=$LEAK"
```

- Recorded baseline: `same=14 diff=0 post_paste_in_templates=34`
- Target: `same=14 diff=0 post_paste_in_templates=0`
- **The `moai-meta-harness` pair was added at audit iteration 2 (N4).** M6 owns 15 mirrors; this loop covers 14 and AC-GSU-020 covers the 15th (`CLAUDE.md`, by clause rather than byte-identity). The `moai-meta-harness` pair — added in iteration 1 to close D6 — had no guard at all: `plan.md` B1 names it among the pairs with no CI parity gate, so a one-sided M4 edit would have shipped as silent drift with every criterion green.

**AC-GSU-020** — The intentionally-divergent `CLAUDE.md` pair carries the identical edited stage-⑤ clause on both sides. Byte-identity is deliberately NOT asserted here (REQ-GSU-018).

```bash
grep -coF 'when a goal is armed (`/moai goal`)' CLAUDE.md internal/template/templates/CLAUDE.md
```

- Recorded baseline: `CLAUDE.md:0` / `internal/template/templates/CLAUDE.md:0`
- Target: `CLAUDE.md:1` / `internal/template/templates/CLAUDE.md:1`

**AC-GSU-021** — Compound neutrality: the template tree carries the new doctrine **and** zero internal-content leak (REQ-GSU-017). The leak half already holds at baseline; the compound fails because the doctrine has not propagated, and it stays failing if the M6 sweep copies SPEC-annotated text.

```bash
grep -rlE 'SPEC-GOAL-SURFACE-UNIFY-001|REQ-GSU-|AC-GSU-' internal/template/templates/ 2>/dev/null | wc -l | tr -d ' '
grep -cF 'Goal-Presentation Timing' internal/template/templates/.claude/rules/moai/workflow/goal-directive.md
```

- Recorded baseline: `0` / `0`
- Target: `0` / `>= 1`

### M7 — Go emission paths (`cycle_type: tdd`)

**AC-GSU-022** — The auto-injected-resume renderer names `/moai goal` in all four locale blocks. Both halves asserted so a deletion (rather than a substitution) cannot pass.

```bash
grep -coF '• /goal ' internal/hook/handoff_inject_render.go
grep -coF '• /moai goal ' internal/hook/handoff_inject_render.go
```

- Recorded baseline: `4` / `0`
- Target: `0` / `4`

**AC-GSU-023** — A renderer regression test exists and passes **with one subtest per locale**. No such test exists at baseline (`ls internal/hook/handoff_inject_render*_test.go` → no matches).

**Rescoped at audit iteration 1 (S6).** The original command counted bare `--- PASS` lines, which four *unrelated* subtests would satisfy. The command now requires each of the four locale identifiers to appear as a subtest name, so the count cannot be met without locale-scoped subtests.

```bash
go test -run 'TestHandoffInjectRender' ./internal/hook/ -v 2>&1 | grep -c -- '--- PASS: TestHandoffInjectRender/\(ko\|ja\|zh\|en\)'
```

- Recorded baseline: `0`
- Target: `4`

**Scope limit, stated rather than overclaimed (S5).** This AC proves a locale-named test exists and is green **at judgment time**. It does NOT prove the RED-before-GREEN ordering that REQ-GSU-021 requires — a change-then-write-test order yields the same `4`. The ordering is process discipline carried by REQ-GSU-021 and made auditable by AC-GSU-033, which requires the observed RED output to be recorded in `progress.md` §E.2. AC-GSU-022 independently pins the four rendered literals, so a test that passes without asserting them still cannot let a wrong string ship.

**AC-GSU-024** — Compound: the `PrimitiveGoal` token is changed **and** the occupancy precondition still holds at 0, so the hard rename remains safe (REQ-GSU-022). The occupancy half already holds at baseline; the compound fails because the token is unchanged.

```bash
grep -c 'PrimitiveGoal              = "/moai goal"' internal/harness/v4manifest/schema.go
grep -rl '"/goal"' .claude/commands/harness/ .claude/workflows/ 2>/dev/null | wc -l | tr -d ' '
```

- Recorded baseline: `0` / `0`
- Target: `1` / `0`
- Occupancy scope measured: 2 manifests (`oss-docs`, `release-update`) + 5 workflow scripts; zero declare the primitive.

**AC-GSU-025** — The runner-template doc comment and dispatch arm both switch. Both halves asserted.

```bash
grep -coF 'case "/goal":' internal/harness/v4manifest/runner_template.go
grep -coF 'case "/moai goal":' internal/harness/v4manifest/runner_template.go
```

- Recorded baseline: `2` / `0`
- Target: `0` / `2`

**AC-GSU-026** — Compound CLI-contract guard: the `--goal` flag **name** is preserved **and** only its help string changes (REQ-GSU-023). An AC on the help string alone would let a rename pass; this one would fail on a rename.

```bash
grep -c '"goal", ""' internal/cli/handoff.go
grep -coF 'a /moai goal condition' internal/cli/handoff.go
```

- Recorded baseline: `1` / `0`
- Target: `1` / `>= 1`

**AC-GSU-027** — Compound over-deletion guard for the Go retention surface: the native-`/goal` yield invariant survives **and** the M7 emission literals are gone (REQ-GSU-024). The pin half already holds at baseline, so the compound is what makes this falsifiable — and a blind sweep that stripped `evaluate.go` would drive the pin down and fail.

```bash
grep -coF 'NativeGoalActive' internal/goal/evaluate.go
grep -coF 'native /goal' internal/goal/evaluate.go
grep -rnoF '"/goal"' --include='*.go' internal/ | grep -v '_test.go' | wc -l | tr -d ' '
```

- Recorded baseline: `2` / `3` / `3`
- Target: `>= 2` / `>= 3` / `0`

### Sync phase — MOVED to `SPEC-GOAL-DOCS-RETIRE-001`

AC-GSU-028 (split-surface emission + retention pins), AC-GSU-029 (sync retention pins), and AC-GSU-032 (strategy-record superseding note) moved with the public-documentation scope after plan-audit iteration 2 emitted STOP. They are re-authored there as AC-GDR-001..007 with **locale-invariant** detectors, closing audit finding N2. The identifier gaps here are deliberate — see `spec.md` §B.7.

### Added at plan-audit iteration 1

**AC-GSU-030** (D3) — Compound: the `v4manifest` package suite exits 0 **and** the test fixture's marker literal moved with the token. The suite passes at baseline, so the compound is what makes this falsifiable; without the second half, M7 could leave the suite red while AC-GSU-027 (which excludes `_test.go`) still read `0`.

```bash
go test ./internal/harness/v4manifest/ >/dev/null 2>&1; echo "exit=$?"
grep -coF 'case "/moai goal"' internal/harness/v4manifest/runner_template_test.go
```

- Recorded baseline: `exit=0` / `0`
- Target: `exit=0` / `1`

**AC-GSU-031** (D5) — Union-detector sweep over the five owned files that carry **no** retention surface, closing the eight blind spots the backtick-anchored detectors could not see. Target 0 is achievable precisely because none of these five files is a retention surface — tested against all three register rows (post-split; see `spec.md` §B.7) before recording.

```bash
grep -rhoE '/goal([^a-zA-Z0-9_./-]|$)' \
  .claude/skills/moai/workflows/run.md \
  .claude/skills/moai/workflows/harness-builder.md \
  .claude/rules/moai/workflow/session-handoff.md \
  .claude/rules/moai/workflow/session-handoff-examples.md \
  .claude/skills/moai-meta-harness/SKILL.md \
  .claude/rules/moai/workflow/orchestration-mode-selection.md | wc -l | tr -d ' '
```

- Recorded baseline: `112` — per file: `run.md` 13, `harness-builder.md` 10, `session-handoff.md` 47, `session-handoff-examples.md` 28, `moai-meta-harness/SKILL.md` 1, `orchestration-mode-selection.md` 13
- Target: `0`
- **`orchestration-mode-selection.md` added at audit iteration 2 (N3).** `plan.md` previously asserted this AC guarded that file's mixed-form residue (`(/goal ac_converge)` at lines 18/145/204/205) while the file was absent from the list — the guard did not exist. Its union count is 13 against a backtick count of 9, so the unguarded delta was exactly the 4 residue tokens. Retention-tested before adding: the file carries no register row, so target `0` is safe.
- **Graceful-degradation component (S10).** M4 rewrites `run.md`'s availability condition for `stop-goal`, and the v2.1.139 floor is a property of the *native* command's Stop-hook wrapper, not of `moai hook stop-goal`. A transposition would satisfy every other AC, so this AC carries an additional component:

```bash
grep -c '2\.1\.139' .claude/skills/moai/workflows/run.md
```

  Recorded baseline: `1` · Target: `0`

- This subsumes AC-GSU-015 on the four files they share; AC-GSU-015 is retained because it also covers `harness-build-entry.md`, `workflows/moai.md`, `SKILL.md`, `orchestration-mode-selection.md`, `dynamic-workflows.md`, and `CLAUDE.md`.

**AC-GSU-033** (S5) — The TDD ordering REQ-GSU-021 requires is made auditable: the observed RED failure output is recorded in `progress.md` §E.2 before the GREEN change. This is the verifiable artefact that AC-GSU-023 cannot supply.

```bash
sed -n '/^## §E.2 Run-phase Evidence/,/^## §E.3/p' .moai/specs/SPEC-GOAL-SURFACE-UNIFY-001/progress.md | grep -c 'FAIL: TestHandoffInjectRender'
```

- Recorded baseline: `0` (§E.2 is the `_<pending run-phase>_` placeholder)
- Target: `>= 1`
- **Self-attestation limit, stated rather than implied (finding A-7).** The artefact is a string the run phase writes into its own `progress.md`, so this criterion attests to itself: its truthfulness is bound by `verification-claim-integrity.md` §1.1 surface 2 (manager-agent self-verification), not by an independent observation. It is the best mechanism available at this layer — the RED state is transient and leaves no other durable trace — and it is a real improvement over asserting AC-GSU-023 blocks the ordering, which it does not. It is **not** mechanical proof, and must not be cited as such.

---

## §C Held-Out Gates (NOT acceptance criteria)

These already pass at baseline, so they are regression gates rather than ACs — an AC that already passes is not an AC. They must still be green at close.

| Gate | Command | Baseline observed |
|---|---|---|
| Template guard suite (mirror parity, neutrality, command audit) | `go test ./internal/template/ 2>&1 \| tail -3` | `ok  github.com/modu-ai/moai-adk/internal/template  0.238s` |
| Full suite | `go test ./... 2>&1 \| tail -5` | not re-run at plan time — run-phase obligation |
| M7 package suites | `go test ./internal/hook/ ./internal/harness/v4manifest/ ./internal/cli/ ./internal/goal/` | not re-run at plan time — M7 obligation |
| Cross-platform build (M7 touches Go) | `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` | not re-run at plan time — M7 obligation |

---

## §D Definition of Done

- All 30 ACs transitioned from their recorded baseline to their target (held-in).
- All four §C gates green (held-out).
- No file under `internal/template/templates/` carries a SPEC ID, REQ/AC token, internal date, or commit SHA.
- Exactly **three** surfaces retain native-`/goal` references, per the `plan.md` §A.2 retention register.
- Working tree contains no changes outside the 37 run-phase paths in `plan.md` §F.1. This SPEC has no sync-phase paths.

---

## §E Edge Cases

- **Removing the prohibition section to satisfy AC-GSU-002.** Blocked: AC-GSU-003 requires the section to exist exactly once, and AC-GSU-002 is scoped to occurrences *outside* it.
- **Stripping `native-invocation-model.md` in a blind sweep.** Blocked by AC-GSU-013's `>= 17` HUMAN-ONLY floor.
- **Deleting the `workflows/goal.md` cross-reference instead of repointing it.** Blocked by AC-GSU-012's second half.
- **Adding the wrapper without `make build`.** Blocked by AC-GSU-016 (embedded-FS reachability).
- **Editing local files without mirrors.** Blocked by AC-GSU-019's `same=14` requirement.
- **Removing the self-check item but leaving the count at 10.** Blocked by AC-GSU-008's `stated == actual` equality.
- **Changing the renderer before writing its test.** Blocked by **AC-GSU-033**, which requires the observed RED failure output to be recorded in `progress.md` §E.2. AC-GSU-023 does NOT block this — a change-then-write-test order yields the same `4` — and the earlier claim that it did was removed at audit iteration 2 (N5).
- **Renaming the `--goal` flag while updating its help string.** Blocked by AC-GSU-026's first half (`"goal", ""` must still be present).
- **Deleting `internal/goal/evaluate.go`'s native-`/goal` yield references as part of a Go sweep.** Blocked by AC-GSU-027's `>= 2` / `>= 3` pins — this would remove the no-double-block safety invariant.
- **Hard-renaming `PrimitiveGoal` after a manifest starts declaring the old token.** AC-GSU-024 re-measures occupancy at judgment time; a non-zero count fails the compound and forces the back-compat path (REQ-GSU-022).
- **Sweeping `docs-site/content/*/claude-code/**`.** Out of this SPEC's scope entirely (`spec.md` §C excludes `docs-site/content/**`); owned by `SPEC-GOAL-DOCS-RETIRE-001` AC-GDR-008's `cc=80` pin.
- **Treating `cli-reference/loop.md` as in scope.** Its `/goal` match is the URL path `/en/cli-reference/goal`, not a command reference; it appears in no AC file list.
- **Sweeping `autonomous-loops.md` to zero.** Out of scope here; owned by `SPEC-GOAL-DOCS-RETIRE-001` AC-GDR-006's per-locale `h3/h2/row` pins.
- **Sweeping the strategy record.** Out of scope here; owned by `SPEC-GOAL-DOCS-RETIRE-001` AC-GDR-007's `25` content pin.
- **Leaving the v4manifest test fixture behind.** Blocked by AC-GSU-030's marker half; the suite-exit half alone would pass at baseline.
- **Swapping only the backticked token on a mixed-form line** (`orchestration-mode-selection.md:18,145,204,205`). Blocked by AC-GSU-031's union detector on the files it covers.

---

## §F REQ ↔ AC Traceability Matrix

Every one of the 28 requirements maps to at least one acceptance criterion. Where a REQ is a constraint rather than an observable, the AC that would fail if the constraint were violated is named, and the mechanism is stated.

| REQ | Covered by | Note |
|---|---|---|
| REQ-GSU-001 (single goal surface) | AC-GSU-001, AC-GSU-015, AC-GSU-031 | Title + emission sweep + union sweep |
| REQ-GSU-002 (emission specifies `/moai goal`) | AC-GSU-015, AC-GSU-022, AC-GSU-025, AC-GSU-031 | Doctrine + Go emission |
| REQ-GSU-003 (pipeline emits no native `/goal`) | AC-GSU-002, AC-GSU-015, AC-GSU-027, AC-GSU-031 | Union + section-carved + Go |
| REQ-GSU-004 (retention surfaces) | AC-GSU-002, AC-GSU-003, AC-GSU-013, AC-GSU-027 | One guard per register row (three rows post-split) |
| REQ-GSU-005 (rewrite in place) | AC-GSU-001, AC-GSU-019 | Path unchanged; a rename would break the mirror pair |
| REQ-GSU-006 (Post-Paste removal) | AC-GSU-006, AC-GSU-007, AC-GSU-008, AC-GSU-011, AC-GSU-012 | Section + row + self-check + render surface |
| REQ-GSU-007 (stated count == actual) | AC-GSU-008, AC-GSU-011 | Both surfaces |
| REQ-GSU-008 (arm-only stated) | AC-GSU-004, AC-GSU-009 | |
| REQ-GSU-009 (Block 5 keeps `/moai run`) | AC-GSU-009, AC-GSU-010 | arm-only clause + progression-mode xref |
| REQ-GSU-010 (presentation at Kickoff) | AC-GSU-004, AC-GSU-010, AC-GSU-014 | Bidirectional linkage |
| REQ-GSU-011 (arming ≠ authorization) | AC-GSU-004 | Second component asserts the invariant sentence inside the section |
| REQ-GSU-012 (rejected alternative recorded) | AC-GSU-005 | |
| REQ-GSU-013 (wrapper exists) | AC-GSU-016, AC-GSU-017, AC-GSU-018 | |
| REQ-GSU-014 (template-first + `make build`) | AC-GSU-016 | Source-tree presence + thin-command-audit reachability; `go:embed` re-embeds from source at every `go test`, so this AC does not observe whether `make build` specifically ran |
| REQ-GSU-015 (`argument-hint` matches verbs) | AC-GSU-018 | Second component asserts `argument-hint` omits the undelivered `resume` verb |
| REQ-GSU-016 (mirror byte-identity) | AC-GSU-019, AC-GSU-020 | 14 pairs + the divergent `CLAUDE.md` pair |
| REQ-GSU-017 (no template leak) | AC-GSU-021 | |
| REQ-GSU-018 (divergent pair handled) | AC-GSU-020 | |
| REQ-GSU-019 (Go emits `/moai goal`) | AC-GSU-022, AC-GSU-025, AC-GSU-026, AC-GSU-027 | |
| REQ-GSU-020 (all four locale blocks) | AC-GSU-022, AC-GSU-023 | Literals + locale-named subtests |
| REQ-GSU-021 (test before change) | AC-GSU-023, AC-GSU-033 | AC-023 proves existence; AC-033 proves the RED ordering |
| REQ-GSU-022 (occupancy precondition) | AC-GSU-024 | Re-measured at judgment time |
| REQ-GSU-023 (flag NAME preserved) | AC-GSU-026 | First half fails on a rename |
| REQ-GSU-024 (no removal of yield invariant / impl) | AC-GSU-027, AC-GSU-030 | Symbol pins + suite green |
| REQ-GSU-028 (test fixture moves with token) | AC-GSU-030 | Added at audit iteration 1 (D3) |

REQ-GSU-025..027 moved to `SPEC-GOAL-DOCS-RETIRE-001` (`spec.md` §B.7); their rows left this matrix with them.

Coverage: **25 / 25 REQs** cited by ≥ 1 AC. Reverse direction: every one of the 30 retained ACs appears in at least one row.

**Proxy-coverage rows closed at audit iteration 2 (S-new-1).** Two rows previously credited a REQ to an AC whose command could not fail on violation. Both ACs were strengthened rather than the rows re-labelled: AC-GSU-004 gained a section-scoped assertion for the arming-never-authorizes invariant (REQ-GSU-011), and AC-GSU-018 gained an `argument-hint`-scoped assertion for the delivered-verb constraint (REQ-GSU-015). Every row's Note now holds.
