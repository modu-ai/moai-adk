# Acceptance Criteria — SPEC-CCSYNC-DYNWF-001

All criteria are grep/content checks on target files plus the template test suite. No behavioral Go test is
introduced. All checks anchor on section HEADINGS, never line numbers (per `plan.md § D.4`).

## D. Acceptance Criteria Matrix

| AC | REQ | Target file | Check type | Severity |
|----|-----|-------------|------------|----------|
| AC-DYNWF-001 | REQ-1 | `dynamic-workflows.md § How a Workflow Runs` | content | Blocking |
| AC-DYNWF-002 | REQ-2 | `CLAUDE.md § 10` | content | Blocking |
| AC-DYNWF-003 | REQ-2 | `moai-domain-research/SKILL.md` | content | Blocking |
| AC-DYNWF-004 | REQ-3 | `dynamic-workflows.md § When to Use a Dynamic Workflow` | content | Blocking |
| AC-DYNWF-005 | REQ-4 | `dynamic-workflows.md § MoAI Integration Notes` | content | Blocking |
| AC-DYNWF-006 | all | template mirror (4 files) | mirror parity | Blocking |
| AC-DYNWF-007 | all | template bodies | neutrality / leak | Blocking |
| AC-DYNWF-008 | all | `go test ./internal/template/...` | regression gate | Blocking |

## Scenarios (Given-When-Then)

### AC-DYNWF-001 — REQ-1 determinism note (Blocking)

- **Given** a SPEC author reading `dynamic-workflows.md § How a Workflow Runs`,
- **When** they look for runtime constraints on the workflow script body,
- **Then** the section states the script body must not call wall-clock or random-number functions (because
  non-determinism breaks resume caching) and that timestamps/random values must be injected via the script's
  input arguments or stamped onto results after the run returns.

Verification:
```bash
# Under the "## How a Workflow Runs" heading, the determinism statement is present.
awk '/^## How a Workflow Runs/{f=1} f; /^## /{if(f && !/How a Workflow Runs/)exit}' \
  .claude/rules/moai/workflow/dynamic-workflows.md | grep -iE "determinist|wall-clock|random"
# Expect: a non-empty match referencing determinism / wall-clock / random and resume caching.
```

### AC-DYNWF-002 — REQ-2 `/deep-research` in CLAUDE.md §10 (Blocking)

- **Given** a reader following `CLAUDE.md § 10` (Web Search Protocol),
- **When** single-pass WebSearch is insufficient for a research-heavy question,
- **Then** §10 cross-references `/deep-research <question>` as the multi-source fan-out + cross-check + claim-vote
  path, and carries: (a) requires the WebSearch tool, (b) a workflow run spends meaningfully more tokens,
  (c) the orchestrator collects/refines the question before launch (AskUserQuestion boundary holds).

Verification:
```bash
awk '/^## 10\. Web Search Protocol/{f=1} f; /^## 11\./{if(f)exit}' CLAUDE.md \
  | grep -iE "deep-research"
awk '/^## 10\. Web Search Protocol/{f=1} f; /^## 11\./{if(f)exit}' CLAUDE.md \
  | grep -iE "WebSearch tool|more token|before launch|mid-run"
# Expect: deep-research present AND the three required facts present.
```

### AC-DYNWF-003 — REQ-2 `/deep-research` in research skill (Blocking)

- **Given** a reader of the `moai-domain-research` skill body,
- **When** they need a heavier multi-source research path than the skill's own parallel WebSearch + Context7,
- **Then** the skill body cross-references `/deep-research` for the same purpose, carrying the same three facts
  (requires WebSearch tool; higher token cost; orchestrator collects the question before launch).

Verification:
```bash
grep -iE "deep-research" .claude/skills/moai-domain-research/SKILL.md
grep -iE "WebSearch tool|more token|before launch|mid-run" .claude/skills/moai-domain-research/SKILL.md
# Expect: deep-research present AND the three required facts present.
```

### AC-DYNWF-004 — REQ-3 routing heuristic (Blocking)

- **Given** an orchestrator choosing among the three runtime primitives,
- **When** it reads `dynamic-workflows.md`,
- **Then** a concise decision heuristic distinguishes dynamic workflow vs Agent Teams vs sequential subagents with
  rough quantitative anchors (dynamic workflow → dozens-to-hundreds of mostly read-only items; Agent Teams → a small
  number of coordinated long-running peers; sequential subagents → default for coding-heavy run-phase), and the
  heuristic reuses (does not contradict) the existing three-primitive table.

Verification:
```bash
grep -iE "dozens-to-hundreds" .claude/rules/moai/workflow/dynamic-workflows.md
grep -iE "sequential subagents" .claude/rules/moai/workflow/dynamic-workflows.md
grep -iE "Agent Teams" .claude/rules/moai/workflow/dynamic-workflows.md
# Expect: all three primitive names appear in a heuristic near the "When to Use" section,
# with the dozens-to-hundreds anchor for the dynamic-workflow branch.
```

### AC-DYNWF-005 — REQ-4 `ultracode` resume pairing (Blocking)

- **Given** a session resumed via a paste-ready resume message,
- **When** that session needs `ultracode` auto-orchestration,
- **Then** `dynamic-workflows.md § MoAI Integration Notes` states that `ultracode` resets on a new session and is
  NOT restored by the `ultrathink.` resume-message opener (which restores reasoning effort only); the session must
  explicitly re-issue `/effort ultracode`, parallel to the re-set `/goal` note.

Verification:
```bash
awk '/^## MoAI Integration Notes/{f=1} f; /^## /{if(f && !/MoAI Integration Notes/)exit}' \
  .claude/rules/moai/workflow/dynamic-workflows.md \
  | grep -iE "ultracode" | grep -iE "resume|new session|ultrathink|re-issue|/effort ultracode"
# Expect: the resume-pairing note for ultracode is present in the section.
```

### AC-DYNWF-006 — Template mirror parity (Blocking)

- **Given** every edited template-distributed file,
- **When** the run-phase edit is complete and `make build` has run,
- **Then** the working copy and its `internal/template/templates/<same-path>` mirror are consistent (the
  mirror-drift test does not fail).

Verification:
```bash
# Each edited file's working copy and template mirror carry the same new content.
for f in \
  ".claude/rules/moai/workflow/dynamic-workflows.md" \
  "CLAUDE.md" \
  ".claude/skills/moai-domain-research/SKILL.md"; do
  diff <(grep -iE "deep-research|ultracode|dozens-to-hundreds|determinist" "$f") \
       <(grep -iE "deep-research|ultracode|dozens-to-hundreds|determinist" "internal/template/templates/$f") \
    && echo "PARITY OK: $f" || echo "DRIFT: $f"
done
# Expect: PARITY OK for every edited file (relative to the REQ this file carries).
```

### AC-DYNWF-007 — Neutrality / no internal-content leak (Blocking)

- **Given** the template-distributed bodies after edit,
- **When** the neutrality / leak guard runs,
- **Then** no internal SPEC ID, REQ/AC token, audit citation, internal date, commit SHA, macOS-bias path, or
  `CLAUDE.local.md` reference appears in any edited template body.

Verification:
```bash
grep -rn "SPEC-CCSYNC-DYNWF" internal/template/templates/ && echo "LEAK — FAIL" || echo "clean — no SPEC-ID leak"
go test ./internal/template/... -run 'TestTemplateNeutralityAudit|TestInternalContentLeak'
# Expect: no SPEC-ID leak AND neutrality/leak tests green.
```

### AC-DYNWF-008 — Template test-suite regression gate (Blocking)

- **Given** all run-phase edits complete and `make build` run,
- **When** the template test suite executes,
- **Then** it passes (mirror-drift + neutrality + leak + embedded-parity), confirming documentation-only changes
  introduced no regression.

Verification:
```bash
go test ./internal/template/...
# Expect: ok (all template tests green).
```

## D.1 Severity Summary

- All eight ACs are **Blocking** (must-pass). There are no nice-to-have criteria for this Tier S doc-seam SPEC.

## Definition of Done

- [ ] AC-DYNWF-001..005 content checks pass on the working-copy target files.
- [ ] AC-DYNWF-006 mirror parity holds for every edited file (working copy == template mirror for new content).
- [ ] AC-DYNWF-007 no internal-content leak; neutrality test green.
- [ ] AC-DYNWF-008 `go test ./internal/template/...` green.
- [ ] `make build` run after all edits (embedded.go regenerated).
- [ ] No Go behavior code modified; no new behavioral Go test added.
- [ ] Pre-existing unrelated working-tree changes (docs-site, CHANGELOG, README.ko.md) NOT included in any commit.
