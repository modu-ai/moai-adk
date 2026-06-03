# Acceptance Criteria — SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001

Each AC is individually grep- or test-verifiable. The `Verify` column gives the canonical command.
Blocking ACs gate completion; informational ACs are observed but do not block.

## A. Marker Installation (B1)

### AC-HAW-001 (Blocking) — InjectMarker has a live call path
**Binds**: REQ-HAW-001
**Given** the harness activation wiring is implemented,
**When** the codebase is searched for non-test callers of `InjectMarker`,
**Then** at least one non-test caller exists (the function is no longer orphaned).
**Verify**:
```bash
grep -rn "InjectMarker(" --include="*.go" internal/ | grep -v "_test.go" | grep -v "func InjectMarker"
# Expected: ≥1 match (a caller, not the definition)
```

### AC-HAW-002 (Blocking) — Marker block is idempotent and paired
**Binds**: REQ-HAW-002
**Given** a project CLAUDE.md,
**When** the marker install runs twice,
**Then** CLAUDE.md contains exactly one `<!-- moai:harness-start` and one `<!-- moai:harness-end -->`.
**Verify**: Go test asserting `strings.Count(content,"moai:harness-start")==1 && strings.Count(content,"moai:harness-end")==1` after a double install (extends `internal/harness/layer3_test.go` idempotency coverage).

### AC-HAW-003 (Blocking) — No AskUserQuestion on the marker-install path
**Binds**: REQ-HAW-003
**Given** any CLI/Go surface that installs the marker,
**When** that source is scanned,
**Then** no `AskUserQuestion` / `mcp__askuser` invocation appears.
**Verify**:
```bash
grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/ internal/harness/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"
# Expected: no matches on the marker-install surface
```

### AC-HAW-004 (Blocking) — Structured error on CLAUDE.md write failure
**Binds**: REQ-HAW-004
**Given** a CLAUDE.md path that cannot be written (absent file / read-only),
**When** the marker install runs,
**Then** it returns a non-nil wrapped error and does NOT report success.
**Verify**: Go test passing a bogus path to the install call path; assert error returned (reuses `InjectMarker`'s existing `read %s: %w` error wrapping).

## B. main.md Router Generation (B2)

### AC-HAW-005 (Blocking) — ScaffoldHarnessDir has a live call path
**Binds**: REQ-HAW-005
**Given** the harness activation wiring is implemented,
**When** the codebase is searched for non-test callers of `ScaffoldHarnessDir`,
**Then** at least one non-test caller exists.
**Verify**:
```bash
grep -rn "ScaffoldHarnessDir(" --include="*.go" internal/ | grep -v "_test.go" | grep -v "func ScaffoldHarnessDir"
# Expected: ≥1 caller
```

### AC-HAW-006 (Blocking) — main.md is a task-shape router manifest
**Binds**: REQ-HAW-006
**Given** a freshly scaffolded `.moai/harness/main.md`,
**When** its content is inspected,
**Then** it contains a domain summary, a routing table mapping task-shapes to harness specialists, and a Linked Files section.
**Verify**: Go test asserting `mainMD()` output contains a routing-table heading (e.g. `## Routing` / `## Task-Shape Routing`) AND a `## Linked Files` section AND a `**Domain**:` line.

### AC-HAW-007 (Blocking) — No agents-without-entry-point skeleton
**Binds**: REQ-HAW-007
**Given** a generated harness with ≥1 `.claude/agents/harness/*.md` agent,
**When** the smoke gate runs,
**Then** the gate fails if `.moai/harness/main.md` is absent.
**Verify**: `go test ./internal/cli/ -run TestRunHarnessCheck` case: agents present + main.md removed → status FAIL.

## C. Generated-Artifact Self-Activation (B4)

### AC-HAW-008 (Blocking) — Generated agents declare skills: preload
**Binds**: REQ-HAW-008
**Given** the generated-agent emission contract (meta-harness skill body + project workflow),
**When** the emission template/instruction is inspected,
**Then** it mandates a `skills:` frontmatter field referencing the companion `my-harness-*` skill.
**Verify**:
```bash
grep -n "skills:" .claude/skills/moai-meta-harness/SKILL.md .claude/skills/moai/workflows/project/meta-harness.md
# Expected: ≥1 match documenting the skills: preload requirement on generated agents
```

### AC-HAW-009 (Blocking) — Generated agents carry non-empty descriptions
**Binds**: REQ-HAW-009
**Given** a generated harness,
**When** the smoke gate runs,
**Then** it fails if any `.claude/agents/harness/*.md` has an empty `description` frontmatter field.
**Verify**: `go test ./internal/cli/ -run TestRunHarnessCheck` case: agent with empty description → status FAIL.

## D. Phase-6 Smoke Gate (REQ-HAW-010..014)

### AC-HAW-010 (Blocking) — Gate fails when main.md absent
**Binds**: REQ-HAW-010
**Verify**: `go test ./internal/cli/ -run TestRunHarnessCheck` case: `.moai/harness/` present, `main.md` removed → FAIL (the existing L5 check already covers `main.md`; assert it surfaces in the smoke gate).

### AC-HAW-011 (Blocking) — Gate fails when CLAUDE.md markers absent/unpaired
**Binds**: REQ-HAW-011
**Verify**: `go test ./internal/cli/ -run TestRunHarnessCheck` case: CLAUDE.md with 0 markers → L3 FAIL (existing `checkLayer3Marker` asserts `1 start / 1 end`).

### AC-HAW-012 (Blocking) — Gate fails on empty agent description
**Binds**: REQ-HAW-012
**Verify**: new `doctor_harness_test.go` case: generated agent with `description:` empty → smoke gate FAIL with a description-specific detail message.

### AC-HAW-013 (Blocking) — Gate fails on dangling skill reference
**Binds**: REQ-HAW-013
**Verify**: new `doctor_harness_test.go` case: generated agent declares `skills:` referencing a `my-harness-X` dir absent on disk → smoke gate FAIL with a dangling-reference detail message.

### AC-HAW-015 (Blocking) — Gate fails when a generated agent OMITS skills: preload (runtime enforcement of REQ-HAW-008)
**Binds**: REQ-HAW-008, REQ-HAW-013b
**Given** a generated harness with ≥1 `.claude/agents/harness/*.md` agent,
**When** the post-generation smoke gate runs and any such agent has NO `skills:` frontmatter key at all,
**Then** the gate FAILs with a missing-`skills:`-preload detail message — so a `skills:`-less generated agent cannot pass silently (the auto-discovery failure mode the SPEC exists to close).
**Verify**: new `doctor_harness_test.go` case: generated agent with frontmatter that lacks any `skills:` key → smoke gate status FAIL with a detail string naming the offending agent and the missing `skills:` key.
**Note**: distinct from AC-HAW-013 (which covers a `skills:` key that is PRESENT but dangling). AC-HAW-015 covers a `skills:` key that is ABSENT. Both are enforced by the same `doctor harness`-extending smoke gate (REQ-HAW-014).

### AC-HAW-014 (Blocking) — Smoke gate extends doctor harness, preserves L1-L5
**Binds**: REQ-HAW-014
**Given** the smoke gate implementation,
**When** `doctor_harness.go` is inspected,
**Then** the new checks live inside / alongside `runHarnessCheck` and the existing L1-L5 status semantics remain (no parallel diagnosis engine, no removed layer).
**Verify**:
```bash
grep -n "L1:\|L2:\|L3:\|L4:\|L5:\|runHarnessCheck" internal/cli/doctor_harness.go
go test ./internal/cli/ -run TestRunHarnessCheck   # all existing L1-L5 cases still pass
```

## E. Process Constraints

### AC-HAW-PROC-1 (Blocking) — Template-first mirror consistency
**Binds**: REQ-HAW-015
**Given** all `.claude/**` edits,
**When** the mirror-drift test runs,
**Then** the working `.claude/` copy matches `internal/template/templates/.claude/...`.
**Verify**: `go test ./internal/template/... -run TestRuleTemplateMirror` (and `make build` ran).

### AC-HAW-PROC-2 (Blocking) — §6.4 prefix correction target equals `my-harness-*` (NOT the EX-1 migration)
**Binds**: REQ-HAW-016
**Disambiguation (EX-1 boundary)**: the §6.4 prefix-correction step (plan.md M4) corrects the stale
`moai-harness-<domain>-*` reference to the **code-side `my-harness-*` prefix that the Go code + generator
actually use today** — it does NOT advance the `my-harness-*` → `harness-*` migration (EX-1, OUT OF
SCOPE). `meta-harness SKILL.md:168` documents the doctrine(`harness-*`)-vs-code(`my-harness-*`) drift; this
SPEC deliberately stays on the code-side `my-harness-*`.
**Given** the generated-artifact emission contract,
**When** the §6.4 Expected Outputs prefix-correction is applied and the emission templates are scanned,
**Then** every corrected generated/referenced skill prefix equals exactly `my-harness-` AND no new
`harness-*`-rename directive (a bare `harness-` prefix without the `my-` lead) is introduced anywhere in
the emission contract.
**Verify**:
```bash
# (a) the §6.4 corrected references MUST read my-harness-, not moai-harness- or bare harness-
grep -nE "my-harness-[a-z]" .claude/skills/moai/workflows/project/meta-harness.md
# Expected: ≥1 match — corrected references carry the my-harness- prefix

# (b) NO new bare harness-* generation directive (a "harness-" not preceded by "my-" or "moai-")
#     in the generated-agent / skill emission contract sections.
grep -nE "(generate|emit|create).*[^a-z-]harness-[a-z]" \
  .claude/skills/moai/workflows/project/meta-harness.md \
  .claude/skills/moai-meta-harness/SKILL.md \
  | grep -vE "my-harness-|moai-harness-"
# Expected: no output — the AC FAILs (and the EX-1 boundary is crossed) if a bare harness-* rename appears
```
This phrasing makes it impossible to read the AC as endorsing the migration: PASS requires `my-harness-`
equality (check a) AND the absence of any bare `harness-*` generation directive (check b).

## F. Definition of Done

- [ ] All Blocking ACs (AC-HAW-001..015 + AC-HAW-PROC-1..2) verified green.
- [ ] `InjectMarker` and `ScaffoldHarnessDir` both have a live (non-test) call path OR a documented
      orchestrated path; the dead-code recurrence is guarded by the smoke gate.
- [ ] `project/meta-harness.md` Phase 7 (5-Layer Activation) body present (was absent).
- [ ] `doctor harness` smoke gate extended with agent-description + dangling-skill checks; L1-L5 preserved.
- [ ] `go test ./...` passes (cascade-clean); `golangci-lint run` baseline clean.
- [ ] Template-first ordering honored; mirror consistent; `make build` ran.
- [ ] Retrofit note for external incomplete harnesses (MINK et al.) documented (no external-project edits).
- [ ] `progress.md` §E.2 run-phase evidence + §E.5 Mx signal populated at close.

## G. Edge Cases

- **EC-1** — Project with no `.moai/harness/` directory: smoke gate is a no-op (existing `runHarnessCheck`
  returns OK with "not present"); no false failure.
- **EC-2** — Re-running generation on a harness that already has a marker: idempotent replace (AC-HAW-002),
  no duplicate block.
- **EC-3** — Windows path handling on the CLI install surface (if M2=Option A): `filepath.Abs` for
  user-supplied paths, cross-platform build verified (`internal/cli/CLAUDE.md` absolute-path rule).
- **EC-4** — Generated agent with a valid `skills:` reference but the skill dir is the
  template-distributed `moai-*` (not `my-harness-*`): not a dangling reference; gate passes.
