# Acceptance — SPEC-PROJECT-NAVIGATOR-001

> Each AC is a binary-testable Given-When-Then scenario. REQs live in `spec.md` §C; this file is the verification layer. Severity: M (Milestone-blocking) unless noted. Traceability: AC-PN-XXX ↔ REQ-PN-XXX.

## §D. AC Matrix

### AC-PN-001 — Three Navigator files produced, no extras

**Given** a fixture project with ≥3 SPECs across mixed statuses (draft, in-progress, completed)
**When** the Navigator regeneration procedure runs
**Then** exactly three files exist under `.moai/project/navigator/`: `navigator.md`, `capability-map.md`, `progress-map.md` — and no other top-level Navigator file is present.

### AC-PN-002 — Every row carries provenance

**Given** `capability-map.md` and `progress-map.md` have been regenerated
**When** each row is inspected
**Then** every row carries a `commit-sha` (40-char hex) AND an ISO-8601 `captured-at` timestamp, both drawn from `git log` for that row's owning file's last commit. A grep for rows missing either field returns zero matches.

### AC-PN-003 — `/moai sync` regenerates Navigator before sync-commit

**Given** a SPEC in the fixture project is at the run→sync boundary
**When** `/moai sync` runs to completion
**Then** the Navigator files in the sync commit reflect the post-sync SPEC state (the just-synced SPEC appears with updated status), verified by diffing the Navigator files in the sync commit against the pre-sync state.

### AC-PN-004 — `/moai project` regenerates Navigator alongside product docs

**Given** the fixture project
**When** `/moai project` is invoked
**Then** `product.md`, `structure.md`, `tech.md`, AND the three Navigator files are all regenerated in the same invocation (single command leaves the full project-context surface current).

### AC-PN-005 — Idempotent regeneration

**Given** the fixture project at commit SHA `C`
**When** the Navigator regeneration procedure runs twice in succession with no intervening commit
**Then** `diff` between the two outputs is empty (byte-identical).

### AC-PN-006 — Empty-project resilience

**Given** a fixture project with zero SPECs and zero commits (freshly initialized)
**When** the Navigator regeneration procedure runs
**Then** it exits 0 and emits a minimal `navigator.md` containing the literal "no features tracked yet" placeholder; `capability-map.md` and `progress-map.md` contain only their header row.

### AC-PN-007 — Malformed-frontmatter tolerance

**Given** the fixture project's 3 SPECs, one has a deliberately malformed `spec.md` frontmatter (missing required field)
**When** the Navigator regeneration procedure runs
**Then** the malformed SPEC does NOT appear in `capability-map.md`; the other two SPECs DO appear; a warning line naming the skipped SPEC is appended to `.moai/logs/navigator-warnings.log`; exit code 0.

### AC-PN-008 — Atomic-rename write strategy (deterministic fixture)

**Given** the regeneration procedure uses an atomic-rename write strategy (write `<file>.tmp` then `mv` into place)
**When** a deterministic test fixture synchronously pauses the procedure AFTER `<file>.tmp` is created but BEFORE the `mv` lands (via a synchronized barrier — e.g. a channel rendezous or a injected pre-rename hook), and a reader observes the target path `navigator.md` at that instant
**Then** the reader observes either the previous version OR the new version (post-`mv`), NEVER a partial / truncated markdown document. The fixture releases the barrier and the read is repeated post-`mv` to confirm the new version lands. Verified by a deterministic concurrency fixture (no millisecond polling, no timing-dependent flakiness).

### AC-PN-009 — SessionStart ambient auto-brief fires

**Given** a fixture project with a populated `.moai/project/navigator/navigator.md`
**When** a new Claude Code session starts in that project
**Then** the session's additionalContext (emitted via `hookSpecificOutput.additionalContext` from `handle-session-start-navigator.sh`, role: ambient auto-brief) contains the current-frontier line, the next-task line, and a link to the full `navigator.md` file; and the emitted additionalContext is at most 500 tokens.

### AC-PN-010 — Hook fails open

**Given** the SessionStart Navigator hook is registered
**When** the hook is invoked but (a) `navigator.md` does not exist, OR (b) the read exceeds its deadline, OR (c) the hook script is missing
**Then** the hook exits 0 and emits no `additionalContext` (or an empty one); the session starts normally without any block.

### AC-PN-011 — `/moai project --brief` loads full brief

**Given** a fixture project with a populated Navigator
**When** `/moai project --brief` is invoked
**Then** the active session context receives the full `navigator.md` entry brief AND the current-frontier section of `progress-map.md` as a structured reorientation brief; a follow-up prompt asking "what is the current frontier?" is answered correctly without further file reads.

### AC-PN-012 — Staleness advisory (N=3 hard-coded default)

**Given** the Navigator's `last-regen-commit.txt` points to a commit that is more than 3 sync cycles behind HEAD (the hard-coded default; overridable via the `navigator.staleness_cycles` config key)
**When** the SessionStart hook emits its additionalContext
**Then** the additionalContext includes a staleness advisory naming the gap (e.g. "Navigator is 3+ sync cycles behind HEAD — recent work may be missing"). Override verification: setting `navigator.staleness_cycles: 1` in config causes the advisory to fire at 2 cycles behind.

### AC-PN-013 — Non-duplication

**Given** a regenerated `capability-map.md`
**When** a row's content is inspected
**Then** the row contains only reference fields (spec-id, title, status, implementation-path, commit-sha, captured-at) and does NOT contain a copy of the SPEC's §A user story, requirements, or any body content exceeding ~200 characters per row.

### AC-PN-014 — Template neutrality

**Given** the template-distributed Navigator surfaces (`moai-workflow-project/SKILL.md`, `references/navigator.md`, `handle-session-start-navigator.sh`, `settings.json.tmpl`)
**When** the CI template-neutrality guard runs (`internal/template/internal_content_leak_test.go` extended with `SPEC-PROJECT-NAVIGATOR-` sentinel)
**Then** the guard finds zero matches for: internal SPEC IDs, REQ tokens (`REQ-PN-`), internal dates, commit SHAs (C2/C3/C7 forbidden classes per CLAUDE.local.md §25.1).

### AC-PN-015 — 16-language neutrality

**Given** a fixture project whose primary language is NOT Go (e.g. Python-only, with `pyproject.toml` and `.py` sources)
**When** the Navigator regeneration procedure runs
**Then** it succeeds without any Go-specific assumption (no `go list`, no `go doc`, no `.go` path bias); the output format is identical to the Go-fixture case modulo the project's own SPECs.

### AC-PN-016 — LSEL boundary non-overlap

**Given** the Navigator regeneration procedure has run
**When** the set of files read and written by the procedure is enumerated
**Then** the read set contains NO LSEL surface (`.moai/lessons-inbox.jsonl`, `.moai/state/lsel/`, `memory/feedback_*.md`, `hns-lsel-*`) AND the write set contains NO LSEL surface.

### AC-PN-017 — `/moai plan` Phase 1 consults Navigator

**Given** a fixture project with a populated `.moai/project/navigator/navigator.md` AND a candidate SPEC description that overlaps an existing capability already tracked in `capability-map.md`
**When** `/moai plan` enters its Phase 1 context-load step
**Then** the skill consults `navigator.md` AND surfaces the overlap (e.g. proposes to amend the existing SPEC rather than create a duplicate), demonstrating that the new SPEC's boundary is drawn against the current frontier.

### AC-PN-018 — `/moai run` consults brief before implementation

**Given** a fixture SPEC at the start of run-phase AND a populated Navigator
**When** `/moai run` begins
**Then** the implementing agent's context contains the Navigator brief (current frontier + the owning SPEC's `progress-map.md` row) before the first implementation action; verified by inspecting the agent's loaded-context log and confirming no re-derivation file reads precede the first work-action.

## §D.1 Severity classification

- **Milestone-blocking (M)**: AC-PN-001, 002, 003, 005, 008, 009, 010, 013, 014, 016 — core invariants (file shape, provenance, sync chain, idempotence, atomicity, ambient brief, fail-open, non-duplication, neutrality, boundary).
- **Important (I)**: AC-PN-004, 011, 012, 015, 017, 018 — feature completeness (project regen chain, --brief full mode, staleness, language neutrality, plan/run consultation).

## §D.2 Indirect verification

- REQ-PN-002 provenance: indirectly verified by AC-PN-002 (grep for missing field) AND AC-PN-005 (idempotence implies determinism implies the timestamp is sourced consistently).
- REQ-PN-010 fail-open: verified by AC-PN-010 across three failure modes (missing file / timeout / missing script).

## §D.3 Closure gates (Definition of Done)

- All 16 ACs PASS with observed evidence (command + output cited per verification-claim-integrity.md §2).
- `make build` clean after template mirror.
- CI template-neutrality guard green.
- plan-auditor PASS on this artifact set.
- Implementation Kickoff Approval human gate passed (three iter-1 open questions resolved per plan.md §B.3).
- Sync-phase merge of the implementation PR lands the Navigator feature on `main`.

## §D.4 Forward-looking checks (deferred to SPEC-002 / SPEC-003)

- `--audit` drift output correctness → SPEC-002 ACs.
- tree-sitter auto-derived capability rows → SPEC-003 ACs.
- These are listed here so reviewers know the boundary; they are NOT in scope for 001's Definition of Done.
