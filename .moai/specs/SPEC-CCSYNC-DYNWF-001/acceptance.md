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
- **Then** the skill body cross-references `/deep-research` (under the `## Works Well With` section)
  for the same purpose, carrying the same three facts (requires WebSearch tool; higher token cost;
  orchestrator collects the question before launch).

Verification (heading-anchored per plan.md §D.4 — bound the grep to the `## Works Well With` section
of the research skill, consistent with the awk-bounded pattern used by AC-001/004/005, rather than a
bare file-wide grep):
```bash
SECTION=$(awk '/^## Works Well With/{f=1} f; /^## /{if(f && !/Works Well With/)exit}' \
  .claude/skills/moai-domain-research/SKILL.md)
echo "$SECTION" | grep -iqE "deep-research" \
  && echo "$SECTION" | grep -iqE "WebSearch tool|more token|before launch|mid-run" \
  && echo "AC-003 PASS — /deep-research cross-ref + facts present in Works Well With section" \
  || echo "AC-003 FAIL — cross-ref absent within bounded section"
# Expect: AC-003 PASS only once REQ-2's cross-ref lands in the bounded section.
# On the un-edited file this prints FAIL (true discriminator).
```

### AC-DYNWF-004 — REQ-3 routing heuristic (Blocking)

- **Given** an orchestrator choosing among the three runtime primitives,
- **When** it reads `dynamic-workflows.md`,
- **Then** a concise decision heuristic distinguishes dynamic workflow vs Agent Teams vs sequential subagents with
  rough quantitative anchors (dynamic workflow → dozens-to-hundreds of mostly read-only items; Agent Teams → a small
  number of coordinated long-running peers; sequential subagents → default for coding-heavy run-phase), and the
  heuristic reuses (does not contradict) the existing three-primitive table.

Verification (section-bounded — NOT file-wide; the un-edited file already has
`sequential subagents` at the §When-NOT-to-Use prose and `Agent Teams` in the L20 three-primitive
table, so a file-wide grep false-positive PASSes 2/3 before REQ-3 lands; the discriminator REQUIRES
all three primitive names AND the `dozens-to-hundreds` anchor to co-occur WITHIN the bounded
`## When to Use a Dynamic Workflow` heuristic block):
```bash
# Extract only the "## When to Use a Dynamic Workflow" section body (awk-bounded, mirrors AC-001/AC-005).
SECTION=$(awk '/^## When to Use a Dynamic Workflow/{f=1} f; /^## /{if(f && !/When to Use a Dynamic Workflow/)exit}' \
  .claude/rules/moai/workflow/dynamic-workflows.md)
# All four signals MUST be present WITHIN that bounded section:
echo "$SECTION" | grep -iqE "dozens-to-hundreds" \
  && echo "$SECTION" | grep -iqE "sequential subagents" \
  && echo "$SECTION" | grep -iqE "Agent Teams" \
  && echo "$SECTION" | grep -iqE "dynamic workflow" \
  && echo "AC-004 PASS — heuristic co-occurs within bounded section" \
  || echo "AC-004 FAIL — heuristic absent or incomplete within bounded section"
# Expect: AC-004 PASS only once REQ-3's heuristic (3 primitive names + dozens-to-hundreds anchor)
# lands inside the bounded section. On the un-edited file this prints FAIL (true discriminator).
```

### AC-DYNWF-005 — REQ-4 `ultracode` resume pairing (Blocking)

- **Given** a session resumed via a paste-ready resume message,
- **When** that session needs `ultracode` auto-orchestration,
- **Then** `dynamic-workflows.md § MoAI Integration Notes` states that `ultracode` resets on a new session and is
  NOT restored by the `ultrathink.` resume-message opener (which restores reasoning effort only); the session must
  explicitly re-issue `/effort ultracode`, parallel to the re-set `/goal` note.

Verification (the pre-existing `ultracode` bullet in §MoAI Integration Notes already contains
`Reverts on a new session`, so a grep that accepts `resume|new session` false-positive PASSes today;
the discriminator REQUIRES the net-new REQ-4 pairing text — `ultracode` AND `ultrathink` AND
(`re-issue` OR `/effort ultracode`) — to co-occur within the bounded section. None of `ultrathink`,
`re-issue`, or `/effort ultracode` appear in the current bullet, so this only passes once REQ-4 lands):
```bash
SECTION=$(awk '/^## MoAI Integration Notes/{f=1} f; /^## /{if(f && !/MoAI Integration Notes/)exit}' \
  .claude/rules/moai/workflow/dynamic-workflows.md)
echo "$SECTION" | grep -iqE "ultracode" \
  && echo "$SECTION" | grep -iqE "ultrathink" \
  && echo "$SECTION" | grep -iqE "re-issue|/effort ultracode" \
  && echo "AC-005 PASS — ultracode resume-pairing co-occurs within bounded section" \
  || echo "AC-005 FAIL — resume-pairing text absent within bounded section"
# Expect: AC-005 PASS only once REQ-4's net-new pairing text lands. On the un-edited file
# this prints FAIL (true discriminator — the L58 bullet has 'new session' but not 'ultrathink'/'re-issue').
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
