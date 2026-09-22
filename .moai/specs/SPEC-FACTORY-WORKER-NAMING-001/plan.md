# plan.md — SPEC-FACTORY-WORKER-NAMING-001 (card t1085)

## §A Context

- Branch: `WT-worker-rename` (already exists — do NOT create a branch), HEAD at plan time `3f3ffbb57`.
- Tier: M. Harness level: standard. Cycle: per `quality.yaml` (DDD/TDD as configured).
- Three work items: (1) GTD todo naming-axis closure record — independent, no dependency; (2) factory `-f agent` join token → `-f worker`; (3) `lane-N` notation → `worker-N`. Items 2–3 serialize behind card t1074 (SPEC-FACTORY-MIXED-HOOK-001).
- Vocabulary sites measured on this tree (baseline: HEAD 3f3ffbb57, this session):
  - `internal/cli/factory.go` — 18 `lane-` hits; `factoryAgentRoleToken = "agent"` (line 55); help text lines 60–84, 148–156, 185–232 region; `resolveFactoryWorkerName` (line 355) delegating to `kanban.ClaimFactoryWorkerName` (`internal/kanban/factory_slots.go:117`).
  - `internal/hook/session_start_factory_i18n.go` — 4-locale status strings naming `lane-1..lane-%d` and `agent-<n>` plus the `-f lane-<n>` hint (en/ko/ja/zh at lines 62/66, 91/95, 117/121, 143/147).
  - Tests referencing lane- vocabulary: `internal/cli/factory_test.go`, `codex_factory_test.go`, `goal_mission_test.go`, `goal_blocked_question_regression_test.go`, `doctor_jev_test.go`, `gtd_compat_test.go`.
  - Doc twins (all four present, currently byte-identical per pair): `.claude/rules/moai/workflow/kanban-dispatch.md` (2 `lane-` hits) + `kanban-dispatch-detail.md` (3 hits), and the template mirrors under `internal/template/templates/.claude/rules/moai/workflow/` (2 + 3 hits).
- SPEC-ID self-check executed: `SPEC-FACTORY-WORKER-NAMING-001` → Bash regex `PASS`; dedup confirmed (no existing directory, no references in `.moai/specs/`).

## §B Known Issues

- SPEC-FACTORY-MIXED-HOOK-001 (t1074) is `in-progress`, run_status fail (11/15 AC), actively repaired; its worktree carries uncommitted `internal/factorymsg/` files using `agent`/`agent-%d` slot vocabulary and an already-modified `factory.go` (20,142 bytes vs develop 18,473). Any parallel rename collides — hence the M2 gate.
- spec-lint REQ collection requires leading list markers on REQ entries — satisfied in spec.md §B (verified lesson from t1020/t1057).
- Ownership lint is blind before commit — commit plan artifacts, then re-run lint (t1047/t1049 lesson).

## §C Pre-flight

- [ ] Worktree confirmed: `git rev-parse --show-toplevel` = this worktree; branch `WT-worker-rename`.
- [ ] No new branch creation; no commit outside `.moai/specs/SPEC-FACTORY-WORKER-NAMING-001/`.
- [ ] GTD closure vehicle decided (see M1 below).
- [ ] Template neutrality check awareness: renamed rule-doc content is programming-language neutral (§15/§25 of CLAUDE.local.md) — no SPEC IDs, internal dates, or commit SHAs leak into the template mirror text.

## §D Constraints

1. [HARD] Compatibility retirement of `-f agent` / `lane-N` forms is decided AFTER M2 measurement — never a priori. No silent destruction (spec REQ-008).
2. [HARD] Template twins are intentionally allowed to diverge; judge each twin file individually; every template-mirror edit is followed by `make build`; verify embed freshness (`make embed-check` or `moai doctor --check "Agent Emit Embed"` awareness — for rule docs the `//go:embed all:templates` refresh via `make build` is the required step).
3. [HARD] GTD 8-tree disposal is t1084's — exclude.
4. [HARD] Serialization: M3/M4 blocked behind the t1074 develop-landing gate; only M1 (and M2's measurement reading of the CURRENT tree) may proceed before the gate.
5. Affected-package testing only (never full `go test ./...` locally — load 413 incident): `go test ./internal/cli/... ./internal/kanban/... ./internal/hook/...`, then CI for the full verdict.

## §E Self-Verification

Plan-phase self-checks (this phase): SPEC-ID regex PASS; frontmatter 12-field schema conformance; dedup confirmed; REQ list markers present; Out of Scope H3 headings with bullets present. Run-phase verification lives in acceptance.md and §F milestones below.

## §F Milestones

Ordering note: per card dispatch, M1 is first (independent). The most decision-loaded work (the rename contract, M3/M4) is deliberately gated behind measurement (M2) so the highest-change-likelihood decisions get human review first; mechanical fallout (test-file edits) is sub-split inside M3.

### M1 — GTD todo naming-axis closure record (independent — proceed immediately)

- Priority: High.
- Vehicle (decided): a closure note at `.moai/specs/SPEC-FACTORY-WORKER-NAMING-001/gtd-todo-naming-closure.md`, committed with this SPEC's artifacts. Rationale: it documents a decision this SPEC's scope statement depends on, and the GTD SPEC directories themselves (SPEC-GTD-AUTONOMY-001, SPEC-GTD-CANON-BODY-001, both `status: completed`) must not be touched — their disposition is t1084's.
- Content: operator decision 2026-09-22 (`moai todo` keeps its name; NOT renamed to GTD); GTD family (t855 naming, t867 canon, t899 landed-store, t939–t941, autonomy ×2) closed as investigation-only; t855 zero-work-commits note; disposal boundary → t1084; pointer to the two completed GTD SPECs (no status transitions performed here).
- Deliverable: closure note file + one-line pointer from spec.md §A (already present).

### M2 — t1074 landing gate + old-token inventory measurement (gate + read-only measurement)

- Priority: High.
- Step 1 (gate, verifiable): read the develop tree — confirm `internal/factorymsg/` exists AND `SPEC-FACTORY-MIXED-HOOK-001` frontmatter status from develop ∈ {implemented, completed}. Record both observations (command + output) in progress.md §E.2-bound notes. FAIL → blocker report, M3/M4 halt (REQ-003).
- Step 2 (inventory, read-only): enumerate every live use of the old tokens: `grep -rn "factoryAgentRoleToken\\|lane-" internal/cli/ internal/kanban/ internal/hook/`; `grep -rn -- "-f agent" . --include="*.go" --include="*.md" --include="*.sh" --include="*.yaml" --include="*.json"` across repo (excluding `.git`, worktrees, `_archive`); `grep -rln "lane-" scripts/ .claude/ internal/template/templates/ 2>/dev/null`; plus a user-script surface sweep (operator-side scripts are out of write reach — they inform the keep-alias decision). Record counts per surface class in progress.md.
- Deliverable: measured inventory table (the M4 decision input).

### M3 — vocabulary rename (agent→worker + lane-N→worker-N), sub-split by M2 inventory

- Priority: High.
- M3a — code: `factoryAgentRoleToken` value `"agent"` → `"worker"`; help text; error strings; `lane-` → `worker-` in user-facing factory CLI surfaces and the join-label construction path (verify the exact default naming format produced by the kanban claim path during run — measured on the landed tree, not assumed from this plan).
- M3b — i18n: all four locale strings in `internal/hook/session_start_factory_i18n.go` updated in lockstep; each locale carries `worker-1..worker-%d`, the `-f worker` join phrasing, and the `worker-<n>` agent-lane replacement; per-locale `-f lane-<n>` hint → `-f worker-<n>` (or the M4-decided alias form).
- M3c — tests: update the six measured test files to worker vocabulary; retain legacy-form cases only where M4's decision says keep-alias (those cases move under the alias coverage, not deleted silently).
- M3d — doc twins: rename `lane-N` notation in both local and template copies of `kanban-dispatch.md` + `kanban-dispatch-detail.md`, judging each file individually (no blind copy); after EACH template-mirror edit run `make build`; prose word "lane" in Kanban-Mode-only contexts judged per line (see Out of Scope).
- Deliverable: renamed surfaces; grep evidence appended to progress.md notes.

### M4 — compatibility decision executed + template embed + verification

- Priority: High (decision) / Medium (verification tail).
- Execute the REQ-007 decision from the M2 inventory: for each old form (`-f agent`, `-f lane-<n>`, `--name lane-<n>`), record keep-alias (parse + deprecation-hint) or remove, with the measured user count as basis. Implement the decision; no form disappears without its inventory row.
- Template embed: `make build` after final template-mirror state; embed freshness check.
- Verification (serial, one at a time): `go vet ./internal/cli/... ./internal/kanban/... ./internal/hook/...`; `golangci-lint run` on affected packages; `go test ./internal/cli/... ./internal/kanban/... ./internal/hook/...`; GTD compat spot-run (`go test ./internal/cli/ -run TestGTD`); final grep sweep vs the AC-010 exception list.
- Deliverable: decision record + verification evidence paths in progress.md.

## §G Anti-Patterns

- Renaming vocabulary before the M2 gate (collides with t1074's live worktree).
- Blind byte-copy between doc twins (they are intentionally allowed to diverge — judge per file).
- Editing a template mirror without `make build` (stale embed ships old vocabulary).
- Deleting old-token test cases without an M4 inventory row (silent destruction).
- Full-suite local test runs (load discipline — affected packages only).

## §H Cross-References

- Trails: SPEC-FACTORY-MIXED-HOOK-001 (card t1074) — serialization precondition.
- Related: SPEC-GTD-AUTONOMY-001, SPEC-GTD-CANON-BODY-001 (completed; naming-axis context; disposition via t1084).
- Acceptance: `acceptance.md` (§D AC Matrix). Closure record: `gtd-todo-naming-closure.md`. Evidence: `progress.md`.
