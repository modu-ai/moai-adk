# SPEC-INVOCATION-MODEL-002 — Acceptance Criteria

All AC are measurable and grep- or diff-verifiable. Traceability to REQ-IM2-001..011 is noted per AC. Verification commands assume project root `/Users/goos/MoAI/moai-adk-go`.

## Deliverable 1 — Divergence reconciliation

### AC-IM2-001 — Authoritative record present (→ REQ-IM2-001)

**Given** this SPEC's `spec.md`, **When** the §A + §C classification is read, **Then** the corrected 9-command matrix is recorded with exactly 6 PROGRAMMATIC and 3 HUMAN-ONLY, and `/security-review` + `/review` appear in the PROGRAMMATIC set.

```bash
# PROGRAMMATIC count (expect 6) + HUMAN-ONLY count (expect 3) in the §A authoritative-record table
grep -c "| PROGRAMMATIC |" .moai/specs/SPEC-INVOCATION-MODEL-002/spec.md   # >= 1 table region; manual: 6 rows
grep -E "/security-review.*PROGRAMMATIC|/review.*PROGRAMMATIC" .moai/specs/SPEC-INVOCATION-MODEL-002/spec.md   # both present
```
PASS: the §A table lists `/code-review`, `/simplify`, `/loop`, `/deep-research`, `/security-review`, `/review` as PROGRAMMATIC and `/goal`, `/clear`, `/compact` as HUMAN-ONLY.

### AC-IM2-002 — Single errata-pointer line appended (→ REQ-IM2-002)

**Given** the closed `SPEC-INVOCATION-MODEL-001/spec.md`, **When** run-phase completes, **Then** exactly one errata-pointer line exists, naming `/security-review` + `/review` as PROGRAMMATIC and pointing to `native-invocation-model.md` + this follow-up SPEC.

```bash
grep -c "Errata (SPEC-INVOCATION-MODEL-002)" .moai/specs/SPEC-INVOCATION-MODEL-001/spec.md   # == 1
```
PASS: count is exactly 1; the line is the last content line of the file.

### AC-IM2-003 — Closed body immutability (→ REQ-IM2-003)

**Given** the closed `spec.md`, **When** the run-phase diff is inspected, **Then** the ONLY change is the single appended errata line; §A body lines (~31-41) are byte-unchanged.

```bash
git -C /Users/goos/MoAI/moai-adk-go diff -- .moai/specs/SPEC-INVOCATION-MODEL-001/spec.md
# Expect: exactly ONE added line ('+ > Errata ...') at end of file; zero removed lines; zero modified §A lines.
```
PASS: `git diff` shows `1 insertion(+), 0 deletions(-)` and the added hunk is at the file tail.

### AC-IM2-004 — Rule file untouched (→ REQ-IM2-004)

**Given** the run-phase changeset, **When** `native-invocation-model.md` (local + template) is checked, **Then** it is absent from the changeset (no classification edit, SPEC-ID cross-ref scoped out).

```bash
git -C /Users/goos/MoAI/moai-adk-go diff --stat -- \
  .claude/rules/moai/workflow/native-invocation-model.md \
  internal/template/templates/.claude/rules/moai/workflow/native-invocation-model.md
# Expect: no output (file not in changeset).
```
PASS: neither path appears in the diff stat.

## Deliverable 2 — Axis-A alignment

### AC-IM2-005 — clean untouched + scope-out recorded (→ REQ-IM2-005)

**Given** the run-phase changeset, **When** `clean.md` (local + template) is checked, **Then** neither is modified; AND this SPEC's `spec.md` §E records the clean↔/simplify scope-out with the capability-mismatch rationale.

```bash
git -C /Users/goos/MoAI/moai-adk-go diff --stat -- \
  .claude/skills/moai/workflows/clean.md \
  internal/template/templates/.claude/skills/moai/workflows/clean.md
# Expect: no output.
grep -i "scoped OUT\|scoped-OUT\|capability mismatch" .moai/specs/SPEC-INVOCATION-MODEL-002/spec.md   # >= 1
grep -i "dead-code" .moai/specs/SPEC-INVOCATION-MODEL-002/spec.md                                     # >= 1
```
PASS: no clean.md change; §E contains the scope-out + dead-code rationale.

### AC-IM2-006 — review compose present, composition preserved (→ REQ-IM2-006)

**Given** `review.md` (local + template), **When** the Phase 2 region is read, **Then** the compose note invokes native `/code-review` via `Skill("code-review")` as one Phase 2 finding source, AND Perspective 1 (Security), Phase 3 (`@MX` compliance), Perspective 4 (UX), and Phase 4.5 (design) headings are still present.

> **Discriminating** — each novel-literal grep below returns **0 on the unedited baseline** (whole-file `code-review` matches only frontmatter L13; whole-file `Skill(` matches only L106 `Skill("moai")`), so this AC FAILS pre-edit and PASSES only after run-phase inserts the canonical compose note (plan.md §A.2). The Phase-2 `awk` scope excludes the frontmatter tags line.

```bash
R=.claude/skills/moai/workflows/review.md
# (1) compose note present — Phase-2-scoped (excludes frontmatter L13); baseline = 0
awk '/^## Phase 2/,/^## Phase 3/' $R | grep -c 'code-review'          # >= 1 (0 on baseline)
# (2) exact invocation literal; baseline = 0 (baseline only has Skill("moai"))
grep -c 'Skill("code-review")' $R                                     # >= 1 (0 on baseline)
# (3) exact compose-note heading; baseline = 0
grep -c '### Native /code-review compose (Axis A)' $R                 # == 1 (0 on baseline)
# (4) composition preserved (unchanged headings)
grep -c "### Perspective 1: Security Review" $R    # == 1 (preserved)
grep -c "## Phase 3: MX Tag Compliance Check" $R   # == 1 (preserved)
grep -c "### Perspective 4: UX Review" $R          # == 1 (preserved)
grep -c "## Phase 4.5: Design Review" $R           # == 1 (preserved)
```
PASS: (1)-(3) each ≥ 1 / == 1 post-edit (all 0 on baseline → discriminating); all four composition headings preserved.

### AC-IM2-007 — conditional-PROGRAMMATIC caveat + fallback (→ REQ-IM2-007)

**Given** `review.md` Phase 2, **When** the compose note is read, **Then** the Phase 2 region states runtime auto-invocability verification (a caveat term) AND a fallback to the existing sync-auditor path — all three confirmations co-located in the Phase 2 `awk` region.

> **Discriminating** — all three greps are scoped to the Phase 2 `awk` region (`/^## Phase 2/,/^## Phase 3/`). The two novel tokens are the discriminators (baseline Phase 2 = 0 for both): whole-file `fallback` matches only L275 (Team Mode, past Phase 3) and `disable-model-invocation` is absent entirely. `sync-auditor` pre-exists in Phase 2 (L64/L68), so it is a **co-location confirmation** that the fallback names the correct agent, NOT the discriminator.

```bash
R=.claude/skills/moai/workflows/review.md
# (1) caveat term — DISCRIMINATOR, Phase-2-scoped; baseline = 0
awk '/^## Phase 2/,/^## Phase 3/' $R | grep -Ec 'disable-model-invocation|disableBundledSkills'   # >= 1 (0 on baseline)
# (2) fallback trigger literal — DISCRIMINATOR, Phase-2-scoped; baseline = 0
awk '/^## Phase 2/,/^## Phase 3/' $R | grep -c 'not auto-invocable'    # >= 1 (0 on baseline)
# (3) fallback names sync-auditor — co-location confirmation (pre-exists in Phase 2, not the discriminator)
awk '/^## Phase 2/,/^## Phase 3/' $R | grep -c 'sync-auditor'          # >= 1
```
PASS: (1) and (2) each ≥ 1 post-edit (both 0 on baseline → discriminating); (3) confirms the fallback agent is named inside the Phase 2 region.

### AC-IM2-008 — composition not weakened (→ REQ-IM2-008)

**Given** the run-phase `review.md`, **When** compared to baseline, **Then** no Security / `@MX` / UX / design content is removed; native `/code-review` is additive only.

```bash
# The four composition headings (AC-IM2-006) remain == 1 each; the Security dependency-scan + secrets-scan blocks remain.
grep -c "Dependency Vulnerability Scan" .claude/skills/moai/workflows/review.md   # == 1 (preserved)
grep -c "Secrets Scan (Full Git History)" .claude/skills/moai/workflows/review.md # == 1 (preserved)
```
PASS: security sub-sections preserved; the compose note uses "augment"/"compose" framing, never "replace".

## Cross-cutting

### AC-IM2-009 — local ↔ template parity + build (→ REQ-IM2-009)

**Given** the edited `review.md`, **When** local and template mirror are compared and the build runs, **Then** the edited region is identical in both and `make build` exits 0.

> **Strict parity** — local and template `review.md` are byte-identical at baseline (`diff` exits 0 today), so the compose edit MUST keep them identical. The escape clause was dropped: a non-empty `diff` after the edit is a FAIL, full stop.

```bash
diff .claude/skills/moai/workflows/review.md \
     internal/template/templates/.claude/skills/moai/workflows/review.md
# Expect: STRICT — empty output, exit 0. Any diff = FAIL.
make build   # exit 0
```
PASS: `diff` produces empty output and exits 0 (byte-identical across both trees); `make build` succeeds.

### AC-IM2-010 — no runtime mechanism added (→ REQ-IM2-010)

**Given** the full changeset, **When** inspected, **Then** no new hook, lint rule, or Go runtime file is added; the changeset is doc/prose only.

```bash
git -C /Users/goos/MoAI/moai-adk-go diff --stat | grep -E "\.claude/hooks/|internal/spec/lint.*\.go|\.go " || echo "NO-RUNTIME-CHANGE"
# Expect: NO-RUNTIME-CHANGE (changeset limited to review.md x2 + closed spec.md + this SPEC's artifacts).
```
PASS: no `.go`, no hook, no lint file in the changeset.

### AC-IM2-011 — template neutrality (→ REQ-IM2-011)

**Given** the template `review.md`, **When** scanned, **Then** it contains no SPEC ID, no REQ token, no ISO date, no commit SHA introduced by this edit.

> **Real test name** — the neutrality guard is `TestTemplateNoInternalContentLeak` (verified present at `internal/template/internal_content_leak_test.go:449`). The prior `TestInternalContentLeak|TestNeutrality` pattern matched ZERO tests (`go test -run` on an unmatched pattern exits 0, silently recording PASS without running the guard). Use the exact name (or the substring `InternalContentLeak`, which also matches).

```bash
grep -c "SPEC-INVOCATION" internal/template/templates/.claude/skills/moai/workflows/review.md   # == 0
grep -Ec "REQ-IM2|2026-07-01" internal/template/templates/.claude/skills/moai/workflows/review.md   # == 0
# Real test — verify it actually runs (not a silent no-match PASS):
go test ./internal/template/ -run 'TestTemplateNoInternalContentLeak' -v 2>&1 | grep -E 'RUN|PASS|FAIL|ok' | tail -5   # RUN + PASS
```
PASS: zero SPEC-ID/REQ/date leaks in the template `review.md`; `TestTemplateNoInternalContentLeak` actually executes (shows `--- PASS`) and passes.

## Definition of Done

- [ ] AC-IM2-001..011 all PASS.
- [ ] Closed `SPEC-INVOCATION-MODEL-001/spec.md`: exactly one errata line appended; §A immutable.
- [ ] `review.md` (local + template) carries the native `/code-review` compose + conditional caveat + sync-auditor fallback; Security/`@MX`/UX/design composition preserved.
- [ ] `clean.md` and `native-invocation-model.md` untouched.
- [ ] `make build` exits 0; local↔template parity holds; template neutrality guard passes.
- [ ] No hook/lint/runtime mechanism introduced.
