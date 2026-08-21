# acceptance.md — SPEC-CODEX-DUAL-AGENTS-001 (Codex Dual Harness M5)

> Verification layer. Each criterion is an `AC-XXX` labeled Given-When-Then, binary-testable.
> `AC-P01..AC-P06` are run-phase probe records, outside the Tier M AC budget (§D.1) — they
> measure codex-cli's acceptance behavior, not the deliverable (executed through the t91
  harness pattern; evidence lands in progress.md §E.2). GEARS obligations live in spec.md §C
  — they are not restated here.

## §D AC Matrix

### AC-001 — Per-file byte-identity golden test (all 11 `.md`)

**Given** the committed template files `internal/template/templates/.claude/agents/moai/*.md`
(11 files) **When** the emitter publishes the agent set **Then** each published `.md` output is
byte-identical to its committed source (sha256 equality), 11 of 11, and the test fails naming
any mismatched file.

### AC-002 — No modification of the committed `.md` tree

**Given** a clean working tree **When** the emitter runs to completion (including failure
paths) **Then** `git status` shows zero modifications under
`internal/template/templates/.claude/agents/moai/` (the `.md` inputs are never rewritten).

### AC-003 — TOML structural validation + verbatim body

**Given** the 11 emitted TOML files **When** each is parsed by a spec-compliant TOML parser
**Then** every file carries non-empty `name` and `description` and a `developer_instructions`
whose decoded value is byte-equal to the corresponding `.md` body text, and `name` equals the
`.md` frontmatter `name` (11 of 11).

### AC-004 — Determinism

**Given** identical inputs (same `.md` set + same manifest) **When** the emitter runs twice in
separate processes **Then** all emitted bytes are identical (pairwise byte comparison; no
timestamps, absolute paths, or environment-derived values anywhere in the output).

### AC-005 — Fail-closed: unknown tool token

**Given** a fixture `.md` whose `tools:` CSV contains a token outside the manifest's mapped
classes **When** emission runs **Then** the emitter exits non-zero, the diagnostic names the
fixture file and the offending token, and no artifact set is partially written.

### AC-006 — Fail-closed: unmapped effort value

**Given** a fixture `.md` with `effort: <value not in the manifest's measured enumeration>`
**When** emission runs **Then** the emitter exits non-zero naming the file and value (never a
silent downgrade — grounded in the M0 silent-ignore caveat, t91 §1).

### AC-007 — MCP server mapping (7 of 11)

**Given** the inventory (7 agents carry `mcp__moai__*` tools: manager-develop, manager-docs,
manager-lead, manager-spec, plan-auditor, super-advisor, sync-auditor; 4 do not:
builder-harness, e2e-tester, manager-design, manager-git) **When** the TOMLs are emitted
**Then** exactly those 7 declare `mcp_servers` containing `"moai"` and the other 4 carry no
`mcp_servers` key.

### AC-008 — Effort mapping per manifest

**Given** the shipped effort values (low ×3, medium ×2, high ×5, xhigh ×1 — spec plan §A.2)
**When** the TOMLs are emitted **Then** each `model_reasoning_effort` equals the manifest's
mapped value for its source effort (identity mapping per plan §A.3 class 10, as locked by
probe P-02), 11 of 11.

### AC-009 — Model omission

**Given** 10 agents declaring `model: inherit` and manager-git declaring `model: sonnet`
**When** the TOMLs are emitted **Then** zero of the 11 files contain a `model` key, and the
manager-git pin is recorded as a documented drop in the mapping manifest.

### AC-010 — Embed and distribute

**Given** the 11 committed TOMLs under `internal/template/templates/.codex/agents/moai/`
**When** `make build` runs **Then** the embedded template FS exposes all 11 TOML paths; and
**When** the deploy path runs against a `t.TempDir()` fixture project **Then** the fixture
contains `.codex/agents/moai/` with the 11 TOMLs byte-equal to the committed sources.

### AC-011 — Template neutrality of emitted files

**Given** the 11 emitted TOML files (new files under `templates/`) **When** the
template-neutrality checks run **Then** the emitted files contain no SPEC-{...} IDs, no
internal dates, and no commit SHAs (grep guard; CI workflow passes on the new paths).

### AC-012 — M0 fact citations present

**Given** spec.md and plan.md **When** audited **Then** the t91 M0 facts used as premises
(agents-TOML works on 0.147.0; silent-ignore caveat; MCP server + 21 tools; effort/sandbox
fields work) are each cited with their t91 report section, and every unmeasured semantic is
marked as a probe item rather than stated as fact.

### AC-013 — Documented-drop completeness

**Given** the mapping manifest's class table (plan §A.3) **When** reviewed **Then** every
semantic class carries exactly one disposition (emit / consequence / documented-drop /
deferred-to-M1 / correspondence-note), and every documented drop carries a rationale — no
class is silently discarded.

## Run-phase probe criteria (t91 harness pattern: isolated `CODEX_HOME`, codex-cli 0.147.0,
user's real `~/.codex` untouched; verbatim command + output recorded in progress.md §E.2)

### AC-P01 — sandbox_mode value set

**Given** candidate `sandbox_mode` values **When** probe agent TOMLs are loaded and a
delegation is attempted per value **Then** the accepted value set is recorded and the
manifest's emitted value belongs to it (or the field ships omitted per the ship-omitted
fallback).

### AC-P02 — model_reasoning_effort enumeration

**Given** candidate effort values (`low`, `medium`, `high`, `xhigh` at minimum) **When**
probed individually **Then** the accepted enumeration is recorded and the manifest's mapping
emits only confirmed values.

### AC-P03 — model field omission semantics

**Given** one probe TOML with no `model` key and one with a bogus `model` string **When** both
are loaded **Then** the omission path is confirmed to inherit the Codex default (and the
bogus-string result demonstrates the silent-ignore behavior or its absence), justifying R-011.

### AC-P04 — agents-dir layout

**Given** identical probe agents placed at `.codex/agents/moai/x.toml` and
`.codex/agents/x.toml` **When** delegation lists available agents **Then** the scanned layout
is recorded and the manifest's layout knob matches it (subdir preferred; flat + `moai-` prefix
fallback).

### AC-P05 — skills.config value set (M1-deferred; optional)

**Given** the optional `skills.config` field **When** probed (only if cheap) **Then** the
observation is recorded for M1's consumption; M5 emits no skills field regardless.

### AC-P06 — per-agent MCP tool filtering (optional)

**Given** an agent TOML with `mcp_servers = ["moai"]` **When** the agent runs **Then** it is
recorded whether the server's 21 tools are all exposed (expected: yes — coarse grant), closing
the documented-drop premise of R-009.

## §D.1 Severity and closure gates

- **Must-pass (blocks close)**: AC-001, AC-002, AC-003, AC-004, AC-005, AC-006, AC-007,
  AC-008, AC-009, AC-010.
- **Should-pass**: AC-011, AC-012, AC-013 (documentation/CI guards — a failure is a real
  defect but does not require redesign).
- **AC budget (Tier M)**: the budgeted AC set is AC-001..AC-013 (13 ≤ 16). AC-P01..AC-P06
  are **probe records**, not budgeted ACs — they measure codex-cli's acceptance behavior
  (environment facts that feed the mapping manifest), not the deliverable. P-01..P-04 are
  required records (verbatim command + output in progress.md §E.2) because the manifest
  enums depend on them; P-05/P-06 are optional records.
- **R-003 rationale anchor** (relocated from the requirement text per audit D5): the
  by-construction guarantee holds because the `.md` is the neutral source itself (Option A,
  plan.md §A.5) — the emitter never re-renders it, so the regression ban cannot degrade into
  a test-only guard.
- **Definition of Done**: all must-pass ACs green with verbatim command+output evidence in
  progress.md §E.2; probe records P-01..P-04 filed with the manifest enums locked or the
  affected fields omitted; committed artifacts (11 TOML + emitter + manifest + tests)
  merged; zero diff under `templates/.claude/agents/moai/`.

## §D.2 Edge cases to cover in tests

- Body containing the TOML multi-line literal delimiter (writer must fail closed or escape —
  never emit a structurally broken file).
- Description block scalar with multi-line content and special characters (round-trips into
  the TOML `description` string).
- Agent with no `skills:` field (plan-auditor) and agents with 2 skills — parser must accept
  both.
- Duplicate agent `name` across files (emitter fails closed — Codex namespace collision).
- Empty tools CSV token (e.g. trailing comma) — parser normalizes or fails closed, never emits
  an empty class lookup.

## §D.3 Indirect verification

- Codex-side acceptance of the real emitted TOMLs: one manual smoke delegation per layout
  (post-AC-P04) citing the t91 `T91PROBE-OK` pattern — recorded in §E.2 as supplementary
  evidence (not a close gate; the binary tests above are the gate).

## §D.7 Forward-looking checks

- If codex-cli upgrades past 0.147.0, re-run P-01..P-04 before trusting the manifest enums
  (manifest records the measured version).
- When M1 lands, class-6 (Skill) disposition changes from deferred to emitted — re-run
  AC-013 completeness against the updated manifest.
