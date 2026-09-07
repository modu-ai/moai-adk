---
id: SPEC-CODEX-COMMAND-SKILLS-001
title: "Acceptance Criteria — Codex Command-to-Skill Publication Emitter"
version: "0.1.0"
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
tier: M
---

# acceptance.md — SPEC-CODEX-COMMAND-SKILLS-001

Conventions: `<EMIT>` = the regeneration make target, `<EMIT>-CHECK` = the read-only drift
make target (names fixed at M2 per plan.md D4). Every AC is binary-testable by its named
command. Release-blocking ACs whose RED cannot be observed on the pre-implementation tree
are classified `regression-guard` per `verification-completeness.md` §2.1 (noted per AC).

## §D AC Matrix

### AC-001 — Exactly 16 published skills (R-001)

**Given** the 16 command sources under `templates/.claude/commands/moai/`
**When** the emitter runs (`make <EMIT>`)
**Then** the published-skills root contains exactly 16 `SKILL.md` files, one per command.

Verify: `find internal/template/templates/.agents/skills -name SKILL.md | wc -l` → `16`.

### AC-002 — Command sources untouched (R-002)

**Given** a clean tree at a pinned HEAD
**When** the emitter runs
**Then** `templates/.claude/commands/moai/` is byte-identical to pre-run.

Verify: `git status --porcelain internal/template/templates/.claude/commands/moai/` → empty,
after running `make <EMIT>`. (Regression-guard class at plan-phase: the red is observable
only after the emitter exists.)

### AC-003 — Derived identity `moai-<command>` (R-003a)

**Given** the published tree
**When** each `SKILL.md` frontmatter is read
**Then** every `name:` equals `moai-<command>` where `<command>` is its directory name, and
the 16 directory names are exactly the 16 command names prefixed.

Verify (run-phase fixture test in the emitter package asserting the set equality; plus a
spot check): `grep -r '^name: moai-' internal/template/templates/.agents/skills/*/SKILL.md | wc -l` → `16`.

### AC-004 — English description, template-syntax-free (R-003b)

**Given** the published tree
**When** each `SKILL.md` description is read
**Then** it equals the English (`{{else}}`) variant of the source command description
(`todo.md`: its plain description), and no emitted file contains Go template syntax.

Verify: `grep -rl '{{' internal/template/templates/.agents/skills/` → no output. Content
equality asserted by the golden test (AC-011's dual-mode comparison pins the bytes).

### AC-005 — Collision refusal (R-004) `[RED-first required]`

**Given** a fixture tree where a derived name (`moai-plan`) collides with an existing
canonical skill directory
**When** the emitter runs against the fixture
**Then** emission fails with a diagnostic naming the derived name and the colliding skill,
and no partial artifact set is written.

Verify: `go test ./internal/template/<emitterpkg>/... -run TestCollisionRefused` — GREEN
after implementation; RED observed first on the pre-guard fixture (E8 evidence). Never
suffixes/overwrites: asserted by the same test (no `moai-plan-2`-style artifact).

### AC-006 — Verbatim body (R-005a)

**Given** the published tree
**When** each `SKILL.md` body (bytes after the frontmatter delimiter) is compared with its
command source body
**Then** all 16 pairs are byte-identical.

Verify: emitter-package test comparing body bytes for all 16 pairs (golden test also pins
this transitively); spot check `diff <(tail -n +5 internal/template/templates/.claude/commands/moai/todo.md) <(tail -n +5 internal/template/templates/.agents/skills/moai-todo/SKILL.md)` → empty (exact offsets fixed at M1; the test is the binding check).

### AC-007 — Boundary flag recorded, not repaired (R-005b)

**Given** the emitted report/golden output
**When** the emission runs
**Then** each published skill's report carries a boundary note that the body references
Claude-only tooling, and the published body still contains the verbatim
`Use Skill("moai")` line (proving no repair happened).

Verify: `grep -rl 'Use Skill("moai")' internal/template/templates/.agents/skills/ | wc -l` → `16`;
boundary-flag presence asserted in the emitter-package test.

### AC-008 — Layout coexistence (R-006)

**Given** a deploy fixture containing both mirror entries (`.agents/skills/<canonical-skill>`
symlinks) and the 16 published directories
**When** mirror creation runs
**Then** no published directory is modified or removed, and no mirror entry is
created/overwritten under a `moai-<command>` name.

Verify: `go test ./internal/template/... -run TestMirrorCommandSkillCoexistence` (GREEN;
RED-first on a constructed fixture where the mirror would clobber).

### AC-009 — Deploy-side user-file safety + fresh-init distribution (R-011, R-010)

**Given** a fresh project initialized from the built binary (`t.TempDir()`)
**When** init completes
**Then** all 16 published skills exist at `<root>/.agents/skills/moai-<command>/SKILL.md`;
and where a user-owned entry pre-occupies one such path, deploy leaves it and reports the
skip.

Verify: init fixture test in `internal/template` or `internal/cli` (fresh `t.TempDir()`
project); occupancy case via the mirror's existing skip semantics or the R-011 addition —
if a `deploy.go` change proves necessary, that is a blocker report first (plan.md M3).

### AC-010 — Drift check: red observed, then green (R-007) `[RED-first required]`

**Given** one committed published artifact hand-mutated
**When** `make <EMIT>-CHECK` runs
**Then** it exits 1 with a diagnostic naming the drift and pointing at the regeneration
verb; after `make <EMIT>`, the same check exits 0; and the check never writes (tree
unchanged by a check-only run on a clean tree).

Verify: the three observations recorded in E1/E8 (mutate → rc=1 + message; regen → rc=0;
`git status --porcelain` after a check-only run → empty).

### AC-011 — Idempotent regeneration + golden pin (R-008)

**Given** a clean tree
**When** `make <EMIT>` runs twice consecutively
**Then** the second run produces a byte-identical tree.

Verify: `make <EMIT> && git status --porcelain internal/template/templates/ | wc -l` → `0`
(after the first run committed the artifacts, the second changes nothing); golden
comparison test (`TestGoldenCommittedArtifactsMatchEmission` pattern) GREEN.

### AC-012 — Template neutrality over the emitted subtree (R-009)

**Given** the emitted files
**When** the neutrality audit runs
**Then** no forbidden class (SPEC IDs, internal dates, commit SHAs, `/Users/` paths,
`CLAUDE.local.md` refs) is present in `templates/.agents/skills/**`.

Verify: `go test ./internal/template/... -run 'TestTemplateNeutralityAudit'` → PASS
(the workflow's path filter already covers the new subtree).

### AC-013 — Cross-platform build stays green (B1)

**Given** the emitter added
**When** the tree builds for both host and windows/amd64
**Then** both builds exit 0.

Verify: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.

## §D.1 Quality gates

- New emitter package coverage ≥ 85% (`go test -cover ./internal/template/<emitterpkg>/...`).
- `golangci-lint run` — no NEW issues attributable to this SPEC (baseline distinguished).
- Emitter is fail-closed like its sibling: a bad input (missing frontmatter, unparseable
  description, collision) aborts emission with a named diagnostic and zero partial output.

## §D.2 Edge cases

- `todo.md` has no `.tmpl` suffix and a plain (non-templated) description — extraction must
  handle both shapes (AC-003/004 cover it via the 16-count and the golden).
- A command description containing a colon or quotes must survive YAML frontmatter quoting
  in `SKILL.md` (golden test pins the emitted bytes).
- A future command added/renamed: drift check goes red (AC-010 shape) — regeneration is the
  documented remedy; stale-artifact cleanup is out of scope (spec.md).

## §D.3 Definition of Done

All 13 ACs PASS with E1 matrix evidence (command + verbatim output + tree SHA); RED-first
evidence present for AC-005, AC-008, AC-010; neutrality audit green; no byte changed under
the §A.5 PRESERVE list.
