# Acceptance Criteria — SPEC-SKILLPORT-CLAIM-CHECK-001

## §D. Acceptance Criteria Matrix

| AC ID | REQ | Criterion | Verification |
|-------|-----|-----------|--------------|
| AC-DCC-001 | REQ-DCC-001 | Skill exists template-first + local-synced | BOTH commands exit 0: `test -f internal/template/templates/.claude/skills/moai-workflow-docs-claim-check/SKILL.md` AND `test -f .claude/skills/moai-workflow-docs-claim-check/SKILL.md` |
| AC-DCC-002 | REQ-DCC-002, REQ-DCC-017 | Frontmatter valid + read-only tool set | TWO halves, both required. **Half 1 — REQ-DCC-002 (Level-1 description):** the `description` field summarizes the trigger and names the read-only, public-docs-only scope (reviewer reads and confirms), and its length is MEASURED and reported against the ~100-token Level-1 metadata budget (≈ 400 characters) — report the measured number; an estimate FAILS this half. **Half 2 — REQ-DCC-017 (frontmatter conventions), all three clauses:** (a) `allowed-tools` is a comma-separated STRING (not a YAML array) holding exactly the read-only set, verified by an EXACT-LINE match scoped to that one field — `awk '/^---$/{c++;next} c==1 && /^allowed-tools:/' internal/template/templates/.claude/skills/moai-workflow-docs-claim-check/SKILL.md` → prints exactly `allowed-tools: Read, Grep, Glob` and nothing else. An exact match on the mandated read-only set subsumes the negative check: a line equal to that string cannot contain Write, Edit, Bash, or AskUserQuestion. (b) if a `skills:` key is present it is a YAML ARRAY (`- item` block or `[a, b]` flow form), never a CSV string — if the key is absent, this clause is vacuously satisfied and MUST be reported as "absent, N/A" rather than silently passed; (c) every value under `metadata:` is a QUOTED string. **A whole-frontmatter `grep` for Write/Edit/Bash/AskUserQuestion is PROHIBITED as the verification here**: REQ-DCC-017 binds the `allowed-tools` value only, and a whole-frontmatter tool-token grep MEASURED 10 of the 28 existing template skills as false positives, matching prose in `description` / `when_to_use`. The skill's read-only BEHAVIOR is a separate concern verified by AC-DCC-008, not here. |
| AC-DCC-003 | REQ-DCC-003 | Progressive disclosure structure | SKILL.md body within the Level-2 budget of **≤ ~5K tokens** (MEASURED, not estimated — `wc -c` on the template SKILL.md, converted at the ~4 bytes/token English-markdown ratio, i.e. ≤ ~20,000 bytes); AND `test -d internal/template/templates/.claude/skills/moai-workflow-docs-claim-check/references` exits 0 with the decision-tree tables + worked examples living there, not inline |
| AC-DCC-004 | REQ-DCC-004 | 3 ordered phases present | grep body: Preflight, Claim Triage, Validation in order |
| AC-DCC-005 | REQ-DCC-006 | Composite→atomic decomposition documented | body describes atomic decomposition, one label per atomic claim |
| AC-DCC-006 | REQ-DCC-007 | 4-label taxonomy + ordered decision tree | body enumerates exactly the 4 labels + ordered tree |
| AC-DCC-007 | REQ-DCC-008 | unsupported-reason set present | body lists {missing-evidence, contradicted, insufficient-coverage} |
| AC-DCC-008 | REQ-DCC-009, REQ-DCC-010, REQ-DCC-011 | HARD read-only boundaries stated | body states: no command execution, no patches, no code/security review |
| AC-DCC-009 | REQ-DCC-012 | 3-section output contract + no-exec certification | body defines Input Scope Reviewed / Claim Assessments / Boundary Notes + literal "no commands executed" certification |
| AC-DCC-010 | REQ-DCC-013 | Policy cross-reference, no duplication | MECHANICAL: `grep -c 'verification-claim-integrity' <template SKILL.md>` ≥ 1. JUDGMENT RESIDUE (explicitly labelled — no executable check exists): a reviewer confirms the body POINTS AT the policy rather than restating it — the body carries no reproduction of the policy's §1 invariant text, §2 baseline-attribution text, or the §3 5-section report format. A word-count proxy is deliberately NOT used: a short duplication would pass it and a long legitimate framing would fail it. |
| AC-DCC-011 | REQ-DCC-014 | English body | no non-English prose in SKILL.md/references |
| AC-DCC-012 | REQ-DCC-015 | Template neutrality (CI guard clean) | `template-neutrality-check` + `internal_content_leak_test.go` pass for the new path |
| AC-DCC-013 | REQ-DCC-016 | Clean-room — MECHANICAL half (absence invariant + house-convention license presence) | ALL THREE clauses (a)-(c) must return the stated result; clause (c) carries two sub-checks (c1, c2), both required. (a) vendor-token absence: `grep -ril 'skillstead' internal/template/templates/.claude/skills/moai-workflow-docs-claim-check/ .claude/skills/moai-workflow-docs-claim-check/` → **0 matches**; (b) no NOTICE at EITHER surface REQ-DCC-016 names — the two skill directories AND the repository root: `test ! -e internal/template/templates/.claude/skills/moai-workflow-docs-claim-check/NOTICE && test ! -e .claude/skills/moai-workflow-docs-claim-check/NOTICE && test ! -e NOTICE` → exit 0 (the repo root carries no NOTICE today — verified at plan time — so this check starts green and guards against one being introduced; matches AC-SVG-016(b)); (c) **house-convention license PRESENT, no skillstead-tied value** — BOTH sub-checks: (c1) `grep -c '^license: Apache-2.0$' internal/template/templates/.claude/skills/moai-workflow-docs-claim-check/SKILL.md` → **exactly 1**; (c2) the frontmatter carries no skillstead reference — `awk '/^---$/{c++;next} c==1 && tolower($0) ~ /skillstead/ {n++} END{print n+0}' internal/template/templates/.claude/skills/moai-workflow-docs-claim-check/SKILL.md` → prints **0** (a clause-local, narrower restatement of (a); written pipe-free so the command survives verbatim copy out of this table cell). Rationale: `license: Apache-2.0` is moai's OWN license declaration per house convention — 18 of the 28 existing template skills carry exactly this field, and the repository LICENSE is Apache-2.0 — so it is NOT skillstead attribution. A bare `grep -n '^license:' → 0 matches` assertion is **PROHIBITED** here: it contradicts REQ-DCC-016's affirmative clause and would false-FAIL a correct implementation. See the absence-invariant caveat below the matrix. |
| AC-DCC-014 | REQ-DCC-001 | Embedded FS rebuilt | `make build` succeeds; new skill embedded |
| AC-DCC-015 | REQ-DCC-005 | Preflight phase covers all 3 sub-steps | body Preflight step: (a) confirms the document is public-facing, (b) inventories the provided evidence with versions + timestamps, (c) flags secrets/sensitive data for redaction before proceeding |
| AC-DCC-016 | REQ-DCC-016 | Clean-room — JUDGMENT half (process attestation) | The implementing agent records a clean-room PROCESS attestation in `progress.md` §E.2 stating that the SKILL.md body, references, and any script comments were drafted **from the functional-capability description in this SPEC only**, and that skillstead source text was NOT consulted while drafting. A reviewer separately confirms the shipped structure follows moai skill conventions (`.claude/rules/moai/development/skill-authoring.md`). This is a PROCESS attestation, NOT a textual-diff check: this SPEC deliberately records no skillstead source text, so a "no verbatim/near-verbatim wording" comparison has no reference corpus and is not executable. |
| AC-DCC-017 | REQ-DCC-015 | At least one non-Go worked example | The shipped `references/` worked examples include ≥ 1 example whose subject documentation claim is NOT Go (e.g. a Python / JavaScript / Rust claim). Verify by reading the worked-example set and naming the non-Go example; a Go-only example set FAILS this AC. |

### Absence-invariant caveat (binds AC-DCC-013)

A passing `grep -ril 'skillstead' … → 0 matches` is an **absence-invariant guard only**. It proves that no attribution token leaked; it is **NOT** evidence that the skill was implemented, nor that the implementation is correct — an empty directory also returns 0 matches. Positive proof of implementation lives in the functional ACs (AC-DCC-001, AC-DCC-004 through AC-DCC-009) and must be evaluated independently.

Why this guard is needed at all: the existing template-neutrality CI guard does **not** cover it. `internal/template/internal_content_leak_test.go` `leakClasses` enforces internal SPEC-ID prefixes, REQ/AC token prefixes, audit citations, internal dates, and memory/archive paths — the vendor token `skillstead` is in **none** of those classes. A leaked attribution footer would therefore pass every existing repo check, which is precisely why this AC carries its own explicit grep.

## §D.1 Given-When-Then scenarios

### Scenario 1 — atomic decomposition + labeling (core path)
- **Given** a README claim "Works on macOS and Linux with Go 1.22+" and user-supplied evidence showing a macOS CI log and go.mod requiring Go 1.22,
- **When** the skill runs Claim Triage then Validation,
- **Then** it decomposes into 3 atomic claims (macOS support / Linux support / Go 1.22+), labels macOS `verified` (evidence anchor: CI log), labels Linux `unsupported` with reason `missing-evidence` (no Linux log), labels Go 1.22+ `verified` (go.mod), and executes zero commands.

### Scenario 2 — read-only boundary enforcement
- **Given** a request that says "verify this claim and fix the doc if wrong",
- **When** the skill processes the request,
- **Then** it verifies and labels the claim but declines the fix, naming the file/edit a human would make, and its Boundary Notes certify no commands were executed.

### Scenario 3 — stale-suspected temporal mismatch
- **Given** a release note claiming "latest version 2.9.0" and evidence (a tag list) showing 3.0.2,
- **When** the skill applies the ordered decision tree,
- **Then** it labels the claim `stale-suspected` (temporal mismatch, once-true-now-outdated) rather than `unsupported`.

## §D.2 Edge cases

- Composite claim with a subjective sub-clause ("fast and reliable on Linux"): the checkable sub-clause is atomized and labeled; the subjective sub-clause ("fast", "reliable") is excluded from labeling and noted in Input Scope Reviewed.
- Zero evidence supplied: all checkable claims label `unsupported`/`missing-evidence` or `needs-human`; the skill names the evidence it would need. It does NOT fabricate a pass.
- Secret/token detected in supplied evidence: Preflight flags it for redaction before continuing.

## §D.3 Definition of Done

- [ ] All 17 AC rows pass.
- [ ] 3 SPEC artifacts + progress.md consistent; status transitions owned downstream (draft set here).
- [ ] `make build` succeeds; skill discoverable in the catalog listing.
- [ ] Template-neutrality CI guard clean on the new path.
- [ ] `allowed-tools` exact-line match green (AC-DCC-002 a): `allowed-tools: Read, Grep, Glob` — no Write, Edit, Bash, or AskUserQuestion listed.
- [ ] Level-1 `description` length MEASURED and reported against the ~100-token budget (AC-DCC-002 Half 1).
- [ ] Clean-room MECHANICAL half green (AC-DCC-013): vendor-token grep → 0; no `NOTICE` in either skill directory nor at the repository root; house-convention `license: Apache-2.0` present exactly once; no skillstead reference anywhere in the frontmatter.
- [ ] Clean-room JUDGMENT half green (AC-DCC-016): process attestation recorded in progress.md §E.2.
- [ ] At least one non-Go worked example present in the shipped references (AC-DCC-017).
- [ ] Any residual clarification-gate marker in the plan-phase artifacts is resolved before Implementation Kickoff Approval (plan→run human gate). (Currently moot — the SPEC carries zero such markers; retained as clarification hygiene.)

## §D.4 Quality gate

- Progressive-disclosure budget respected (Level-1 description within the ~100-token Level-1 metadata budget; body ~5K tokens; heavy tables in Level-3).
- No duplication of `verification-claim-integrity.md` (pointer + framing only).
