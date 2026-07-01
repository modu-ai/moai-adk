# Acceptance Criteria — SPEC-INVOCATION-MODEL-001

> Every AC is observable (grep match, file existence, `diff` parity, or `go test` output). Traceability column maps each AC to its REQ.

---

## §D. AC Matrix

| AC ID | REQ | Verification | Pass condition |
|-------|-----|--------------|----------------|
| AC-IM-001 | REQ-IM-001 | grep the doctrine rule for the taxonomy | Rule file contains both `PROGRAMMATIC` and `HUMAN-ONLY` category names, the `Skill()`/`Workflow()` auto-invocation statement, and the `[Skill]`/`[Workflow]`-marker heuristic |
| AC-IM-002 | REQ-IM-001 | grep for the official citation | Rule file cites `skills.md` with the verbatim "Unlike most built-in commands, which execute fixed logic directly, bundled skills are prompt-based" quote (or a faithful trimmed quotation) |
| AC-IM-003 | REQ-IM-002 | grep the classification matrix | All 9 native commands (`/code-review`, `/simplify`, `/loop`, `/deep-research`, `/security-review`, `/goal`, `/review`, `/clear`, `/compact`) appear in the matrix, each tagged PROGRAMMATIC or HUMAN-ONLY |
| AC-IM-004 | REQ-IM-003 | grep for the thesis | Rule file states the automation-justification thesis with both Axis A (reuse via `Skill()`) and Axis B (reimplement HUMAN-ONLY) |
| AC-IM-005 | REQ-IM-004 | grep per-subcommand notes | Rule contains a per-subcommand justification note for `feedback` (no native equivalent — targets remote tool repo) and at least one HUMAN-ONLY-automation example (e.g. review --security) |
| AC-IM-006 | REQ-IM-005 | grep for the caveat | Rule documents that `disable-model-invocation: true` OR `disableBundledSkills` removes auto-invocability, requiring runtime verification |
| AC-IM-007 | REQ-IM-006 | grep feedback.md (local + template) | Feedback workflow specifies auto-collection of the 2 guaranteed items — `moai version`, `uname` (OS) — in the issue body (the testable contract). `go version` (best-effort, tool build-provenance per the §15 neutrality note) AND the optional orchestrator-mediated error context are best-effort; their absence is NOT an AC failure |
| AC-IM-008 | REQ-IM-007 | grep feedback.md | Feedback workflow states the tool-diagnostic-only boundary and prohibits attaching arbitrary user file contents |
| AC-IM-009 | REQ-IM-008 | grep feedback.md | Feedback workflow includes the `gh issue list --repo <target> --search` duplicate-detection step producing a structured report |
| AC-IM-010 | REQ-IM-009 | grep feedback.md | Feedback workflow includes a gh-failure fallback path (detect `gh auth status` failure / rate-limit → guide user → offer local save of drafted body) |
| AC-IM-011 | REQ-IM-010 | `go test ./internal/config/...` | Feedback repo config resolution unit test PASSES: absent/empty `.moai/config/sections/feedback.yaml` (or missing `feedback.repository` key) resolves to default `modu-ai/moai-adk` |
| AC-IM-012 | REQ-IM-011 | `go test ./internal/config/...` | Feedback repo config resolution is verified by an **integration/file-load test** (NOT a hand-constructed struct): the test writes a real `feedback.yaml` with a `feedback.repository` override into a temp dir and loads it through `Loader.Load()` (exercising `feedbackFileWrapper` + `loadFeedbackSection` registration), asserting the override value resolves |
| AC-IM-013 | REQ-IM-010/011 | grep feedback.md | The hardcoded `--repo modu-ai/moai-adk` is replaced by a resolved-config reference (no bare hardcode remains as the sole target source) |
| AC-IM-014 | REQ-IM-012 | `diff` parity | `diff .claude/skills/moai/workflows/feedback.md internal/template/templates/.claude/skills/moai/workflows/feedback.md` → IDENTICAL |
| AC-IM-015 | REQ-IM-012 | `diff` parity | `diff .claude/rules/moai/workflow/native-invocation-model.md internal/template/templates/.claude/rules/moai/workflow/native-invocation-model.md` → IDENTICAL |
| AC-IM-016 | REQ-IM-012 | `diff` parity | `diff .moai/config/sections/feedback.yaml internal/template/templates/.moai/config/sections/feedback.yaml` → IDENTICAL |
| AC-IM-017 | REQ-IM-012 | `make build` + `go build ./...` | `make build` succeeds and the binary compiles with the embedded new templates |
| AC-IM-018 | REQ-IM-013 | `go test ./internal/template/...` | Template-neutrality guard (`internal_content_leak_test.go`) PASSES — no internal SPEC IDs / REQ tokens / dates / SHAs in the template-bound doctrine rule, feedback.md, or feedback.yaml |
| AC-IM-019 | REQ-IM-014 | boundary grep | `grep -n 'AskUserQuestion' .claude/skills/moai/workflows/feedback.md` shows NO new AskUserQuestion call added for duplicate resolution (existing feedback-collection AskUserQuestion usage is orchestrator-level and unchanged) |
| AC-IM-020 | (schema) | `moai spec lint` | The SPEC passes frontmatter schema + OutOfScopeRule lint (12 canonical fields present; `### Out of Scope —` sub-headings present) |
| AC-IM-021 | REQ-IM-002 | grep the doctrine rule per-command | EACH of the 9 command classifications in the doctrine rule carries an inline citation anchor — an official `skills.md`/`commands.md` quote OR the observed `[Skill]`/`[Workflow]` marker (present for bundled, absent for built-in). The `/loop` row + a cross-reference correcting `goal-directive.md`'s "time-interval scheduler" framing are present |
| AC-IM-022 | REQ-IM-002 | run-phase evidence (progress.md §E.2) | Run-phase RECORDS an official-docs verification of the classification — a WebFetch of `commands.md`/`skills.md` — as evidence in progress.md §E.2 before the doctrine rule's classification matrix is committed (verification-claim-integrity §1.1: no unobserved classification claim) |

---

## §D.1 Given-When-Then Scenarios

### Scenario 1 — Feedback repo defaults to the tool channel (config resolution)

- **Given** a project with NO `.moai/config/sections/feedback.yaml` (or an empty `feedback:` section)
- **When** the feedback repo config accessor is resolved
- **Then** it returns `modu-ai/moai-adk` (the default tool feedback channel) — verified by `go test ./internal/config/...` (AC-IM-011).

### Scenario 2 — Fork maintainer overrides the feedback repo

- **Given** a `.moai/config/sections/feedback.yaml` with `feedback.repository: myfork/moai-adk`
- **When** the feedback repo config accessor is resolved
- **Then** it returns `myfork/moai-adk` (the override), and the feedback workflow's `gh issue create --repo <resolved>` targets the fork — verified by `go test ./internal/config/...` (AC-IM-012) + AC-IM-013 grep.

### Scenario 3 — Duplicate detection surfaces candidates without in-skill prompting

- **Given** a user submits a feedback title matching an existing open issue
- **When** the feedback workflow runs the `gh issue list --search` step
- **Then** it emits a structured "possible duplicates" report for the orchestrator (NOT an in-skill AskUserQuestion) — verified by AC-IM-009 + AC-IM-019.

### Scenario 4 — gh unauthenticated triggers graceful fallback

- **Given** `gh auth status` fails (unauthenticated or rate-limited)
- **When** the user invokes `/moai feedback`
- **Then** the workflow guides the user to resolve auth/rate-limit and offers to save the drafted issue body locally so no draft is lost — verified by AC-IM-010.

---

## §D.2 Edge Cases

- **EC-1** — Empty `feedback:` section key present but `repository:` missing → resolver falls back to default `modu-ai/moai-adk` (covered by AC-IM-011).
- **EC-2** — Config loader uses an explicit section allowlist rather than a directory scan → new `feedback.yaml` must be registered; if unregistered, AC-IM-011/012 FAIL, surfacing the gap (plan.md R2 mitigation).
- **EC-3** — `go` toolchain absent when collecting diagnostics (the COMMON case — most users run a prebuilt `moai` binary with no Go toolchain) → `go version` collection degrades gracefully; the 2 guaranteed items (MoAI version + OS/uname) still attach. `go version` is best-effort (REQ-IM-006 + §15 neutrality note), so its absence is NOT an AC failure.
- **EC-4** — Doctrine rule cites a native command whose live classification differs from the plan.md draft table → run-phase MUST re-verify against official docs before committing (plan.md R1).

---

## §D.3 Quality Gate Criteria

- `go test ./...` — full suite GREEN (no regressions from the new config struct/accessor).
- `go test ./internal/config/...` — includes the new feedback-repo resolution test, GREEN.
- `go test ./internal/template/...` — template-neutrality + parity guards GREEN.
- `golangci-lint run` — clean on the modified `internal/config` files.
- `moai spec lint` — SPEC frontmatter + OutOfScopeRule clean.
- Coverage: the new `internal/config` feedback accessor is covered by the new unit test (target ≥ 85% for the touched resolver path).

---

## §D.4 Definition of Done

- [ ] Doctrine rule created at the confirmed path + template mirror, with per-command citation anchors + `/loop` framing correction (AC-IM-001..006, 015, 021) and run-phase official-docs verification recorded (AC-IM-022).
- [ ] Feedback workflow enhanced with all 4 items + template mirror (AC-IM-007..010, 013, 014).
- [ ] Feedback repo config-ization implemented — `FeedbackConfig` + `feedbackFileWrapper` + `loadFeedbackSection` + `Loader.Load()` wiring — with a passing integration/file-load test (AC-IM-011, 012).
- [ ] `.moai/config/sections/feedback.yaml` created + template mirror (AC-IM-016).
- [ ] `make build` succeeds; binary compiles (AC-IM-017).
- [ ] Template-neutrality guard passes (AC-IM-018).
- [ ] AskUserQuestion boundary preserved in feedback.md (AC-IM-019).
- [ ] Full test suite + lint + spec-lint GREEN (AC-IM-020 + §D.3).
- [ ] All 3 design decisions fixed (plan.md §H Resolved Design Decisions) — rule path, feedback config location/key, HYBRID diagnostic scope — reflected across spec.md / plan.md / acceptance.md.
