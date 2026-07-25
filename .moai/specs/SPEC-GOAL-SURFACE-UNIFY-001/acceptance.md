---
id: SPEC-GOAL-SURFACE-UNIFY-001
title: Unify the goal surface on /moai goal and relocate goal presentation to the Implementation Kickoff Approval gate
version: 1.0.0
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: HIGH
phase: plan
module: doctrine
lifecycle: spec-anchored
tags: [goal, doctrine, session-handoff, slash-command, template-mirror]
tier: L
---

## §A Baseline Provenance

Every judgment command below was **executed** in the worktree at `origin/main` = `e306e21a9`, branch `feat/SPEC-GOAL-SURFACE-UNIFY-001`, working tree clean (0 dirty files). The `Recorded baseline` column carries the verbatim observed output. Every AC **FAILS** at this baseline and PASSES only after its milestone lands.

Two conventions, chosen after running both forms and comparing:

- **Detector literal** — the open form `` `/goal `` (backtick + `/goal`, no closing backtick). Chosen over the closed form `` `/goal` `` because the open form also catches `` `/goal ac_converge` `` and `` `/goal`-turn ``, which the closed form misses. Counts in this file are therefore not comparable to a closed-form count.
- **Occurrence vs line counts** — `grep -o … | wc -l` (occurrences) is used throughout, never bare `grep -c` (lines). The two differ materially here: `session-handoff.md` has 12 *lines* but 18 *occurrences* of `post-paste`.

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

## §B AC Matrix — 29 criteria

### M1 — `goal-directive.md` rewrite in place

**AC-GSU-001** — The rule's title line names `/moai goal` as the doctrine subject.

```bash
head -1 "$G" | grep -c '/moai goal'
```

- Recorded baseline: `0` (title reads `# Goal Directive (\`/goal\`) — Autonomous Continuation`)
- Target: `1`

**AC-GSU-002** — Every residual native-`/goal` reference in the file is confined to the retained prohibition section; none survives anywhere else. Section-scoped, so it cannot be satisfied by deleting the prohibition rationale.

```bash
awk '/^## /{sec=$0} /`\/goal/{if (sec !~ /Native .\/goal. Prohibition/) c++} END{print c+0}' "$G"
```

- Recorded baseline: `33`
- Target: `0`

**AC-GSU-003** — The retained prohibition section exists exactly once (not zero — the rationale must survive; not twice — one home only).

```bash
grep -c '^## Native `/goal` Prohibition$' "$G"
```

- Recorded baseline: `0`
- Target: `1`

**AC-GSU-004** — The W2 codification home exists as a named sub-section.

```bash
grep -c '^### Goal-Presentation Timing$' "$G"
```

- Recorded baseline: `0`
- Target: `1`

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

**AC-GSU-015** — All 8 emission-path files carry zero native-`/goal`. The two retained surfaces (§A.2 of `plan.md`) are deliberately excluded from this file list.

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

**AC-GSU-016** — **Reachability, not text presence.** The new wrapper is actually embedded in the binary's template FS and passes the thin-command audit. `TestCommandsThinPattern` walks the *embedded* FS, so this fails unless `make build` ran; and the subtest is named after the path, so a match proves the file was both embedded and audited.

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

**AC-GSU-018** — The local wrapper routes to the `goal` subcommand (the routing target, not merely a file that exists).

```bash
grep -cF 'arguments: goal' .claude/commands/moai/goal.md
```

- Recorded baseline: `ugrep: warning: .claude/commands/moai/goal.md: No such file or directory` (non-zero exit — file absent)
- Target: `1`

### M6 — Template mirror sweep

**AC-GSU-019** — Compound parity: all 13 `.claude/**` pairs are byte-identical **and** the template mirrors carry zero retired-mechanism token. Byte-parity alone already holds at baseline, so the compound is what makes this falsifiable — it verifies parity was *preserved across* the edit rather than merely observed before it.

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
         .claude/skills/moai/workflows/run.md; do
  t="internal/template/templates/$f"
  if diff -q "$f" "$t" >/dev/null 2>&1; then SAME=$((SAME+1)); else DIFF=$((DIFF+1)); fi
  n=$(grep -oiF 'post-paste' "$t" 2>/dev/null | wc -l | tr -d ' '); LEAK=$((LEAK+n))
done
echo "same=$SAME diff=$DIFF post_paste_in_templates=$LEAK"
```

- Recorded baseline: `same=13 diff=0 post_paste_in_templates=34`
- Target: `same=13 diff=0 post_paste_in_templates=0`

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

**AC-GSU-023** — **The RED half of the TDD cycle.** A four-locale renderer regression test exists and passes. No such test exists at baseline (`ls internal/hook/handoff_inject_render*_test.go` → no matches), so this is the guard M7 must author before changing behaviour (REQ-GSU-021).

```bash
go test -run 'TestHandoffInjectRender' ./internal/hook/ -v 2>&1 | grep -c -- '--- PASS'
```

- Recorded baseline: `0`
- Target: `>= 4` (one PASS line per locale)

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

### Sync phase — public and internal documentation (owner `manager-docs`)

**AC-GSU-028** — The 13 affected documentation files carry zero native-`/goal` emission reference.

```bash
grep -rohF '`/goal' \
  docs-site/content/*/advanced/autonomous-loops.md \
  docs-site/content/*/cli-reference/handoff.md \
  docs-site/content/*/advanced/self-evolving.md \
  .moai/docs/autonomous-workflow-strategy.md | wc -l | tr -d ' '
```

- Recorded baseline: `89` (autonomous-loops 52 + handoff 4 + self-evolving 8 + strategy 25)
- Target: `0`

**AC-GSU-029** — Compound retention guard: every sync-phase retention group keeps its exact occurrence count **and** the affected set is clean. The pins already hold at baseline, so the compound is what makes this falsifiable; a sweep that over-reached into the `claude-code/**` pages would drive a pin off its value and fail.

```bash
grep -rohF '`/goal' docs-site/content/*/claude-code/ | wc -l | tr -d ' '
grep -rohF '`/goal' docs-site/content/*/cli-reference/goal.md | wc -l | tr -d ' '
grep -rohF '`/goal' docs-site/content/*/utility-commands/moai-goal.md | wc -l | tr -d ' '
grep -rohF '`/goal' docs-site/content/*/advanced/hooks-reference.md | wc -l | tr -d ' '
grep -rohF '`/goal' .moai/research/ | wc -l | tr -d ' '
```

- Recorded baseline: `80` / `12` / `4` / `2` / `4`
- Target: `80` / `12` / `4` / `2` / `4` — unchanged, **and** AC-GSU-028 at `0`

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

- All 29 ACs transitioned from their recorded baseline to their target (held-in).
- All four §C gates green (held-out).
- No file under `internal/template/templates/` carries a SPEC ID, REQ/AC token, internal date, or commit SHA.
- Exactly **four** surfaces retain native-`/goal` references, per `plan.md` §A.2.
- Working tree contains no changes outside the 35 run-phase paths in `plan.md` §F.1 (sync-phase paths land in the sync commit, not the run commits).

---

## §E Edge Cases

- **Removing the prohibition section to satisfy AC-GSU-002.** Blocked: AC-GSU-003 requires the section to exist exactly once, and AC-GSU-002 is scoped to occurrences *outside* it.
- **Stripping `native-invocation-model.md` in a blind sweep.** Blocked by AC-GSU-013's `>= 17` HUMAN-ONLY floor.
- **Deleting the `workflows/goal.md` cross-reference instead of repointing it.** Blocked by AC-GSU-012's second half.
- **Adding the wrapper without `make build`.** Blocked by AC-GSU-016 (embedded-FS reachability).
- **Editing local files without mirrors.** Blocked by AC-GSU-019's `same=13` requirement.
- **Removing the self-check item but leaving the count at 10.** Blocked by AC-GSU-008's `stated == actual` equality.
- **Changing the renderer before writing its test.** Blocked by AC-GSU-023 requiring ≥ 4 PASS lines from a test that does not exist at baseline; a change-first order leaves the assertion unwritten and the AC failing.
- **Renaming the `--goal` flag while updating its help string.** Blocked by AC-GSU-026's first half (`"goal", ""` must still be present).
- **Deleting `internal/goal/evaluate.go`'s native-`/goal` yield references as part of a Go sweep.** Blocked by AC-GSU-027's `>= 2` / `>= 3` pins — this would remove the no-double-block safety invariant.
- **Hard-renaming `PrimitiveGoal` after a manifest starts declaring the old token.** AC-GSU-024 re-measures occupancy at judgment time; a non-zero count fails the compound and forces the back-compat path (REQ-GSU-022).
- **Sweeping `docs-site/content/*/claude-code/**` along with the MoAI surface.** Blocked by AC-GSU-029's `80` pin.
- **Treating `cli-reference/loop.md` as in scope.** Its `/goal` match is the URL path `/en/cli-reference/goal`, not a command reference; it appears in no AC file list.
