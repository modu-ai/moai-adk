---
id: SPEC-V3R6-RULES-HOTFIX-001
title: "rules/ low-risk hotfix sweep — inverted tool names, inverted-logic clause, typo, broken cross-references"
version: "0.2.0"
status: in-progress
created: 2026-06-19
updated: 2026-06-19
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai + internal/template/templates/.claude/rules/moai"
lifecycle: spec-anchored
tags: "rules, hotfix, cross-reference, template-mirror, glm, lint"
tier: S
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-19 | manager-spec | Initial plan-phase draft. Tier S LEAN (inline acceptance criteria). SPEC 1 of 4 in the Sprint 16 rules-improvement cohort. |
| 0.2.0 | 2026-06-19 | manager-spec | plan-auditor 0.82 PASS-WITH-DEBT revision. BLOCKING: purged the `embedded.go` premise (verified absent; real mechanism is `//go:embed all:templates` + `TestRuleTemplateMirrorDrift`). D1 cite live-runtime verification + tool-policy.yaml:776 forward-note. D2-file intentional neutrality divergence (no byte-identity reconcile). D4/D5 prose rewrites (`skill:` label removed; `**Handoff schema on exit 2**` bold-marker label). AC-RHF-006 `:0` predicate tightened. |

---

## §A. Context

This SPEC is **1 of 4** in the Sprint 16 "rules-improvement" cohort derived from a full audit of `.claude/rules/`. It covers ONLY the low-risk, single-token / broken-cross-reference fixes that carry no behavioral ambiguity and no new content authoring. The remaining 3 cohort SPECs (CATALOG-SCRUB, SSOT-DEDUP, VERSION-FORMAT) own archived-agent scrub, SSOT de-duplication, and version/format normalization respectively — their scope is fenced out of this SPEC (see Exclusions).

Each defect below was verified during the audit against the live tree (deployed file + template mirror). Every affected file under `.claude/rules/**` has a mirror under `internal/template/templates/.claude/rules/**`; per CLAUDE.local.md §2 (Template-First Rule), the corrected token/line MUST be applied to BOTH copies.

**Embed mechanism (verified, NOT `embedded.go`):** there is NO generated `internal/template/embedded.go` — `find internal/template -name embedded.go` returns nothing. The template tree is embedded directly at compile time via `//go:embed all:templates` (`internal/template/embed.go:28`), so the corrected template SOURCE files are picked up automatically by the next `go build ./...`. `make build` runs `templ-generate` + `go run ./internal/template/scripts/gen-catalog-hashes.go --all` + `go build` — it regenerates NO `embedded.go`. The post-edit verification is therefore: (a) `go build ./...` succeeds (the `//go:embed` re-embeds the corrected template tree), and (b) deployed↔template mirror parity is checked by the existing Go test `TestRuleTemplateMirrorDrift` in `internal/template/rule_template_mirror_test.go`.

**Intentional deployed↔template divergence (D2 file):** `verification-claim-integrity.md` deployed and template copies are NON-identical by design — the template scrubs internal SPEC-IDs and coverage numbers per CLAUDE.local.md §25 neutrality (e.g. deployed `SPEC-EVIDENCE-CLAIM-INVARIANT-001` / `coverage 87%` → template generic). The D2 line-9 correction happens to be byte-identical in both copies, so applying it to both is safe. The run-phase MUST apply ONLY the corrected token/line to each mirror and MUST NOT reconcile the pre-existing intentional neutrality divergence to byte-identity.

This SPEC is **documentation/rule-content only**. No Go logic changes are expected beyond the `//go:embed`-driven re-embed at `go build` time (no codegen, no `embedded.go`).

---

## §B. In-Scope Defects (verified, file:line cited)

### D1 — glm-web-tooling.md inverted z.ai vision tool names

`.claude/rules/moai/core/glm-web-tooling.md` lines 33, 35, 50, 65, 72 use `image_analysis` and line 72 uses `video_analysis`. These names are **inverted** vs the actual runtime tool names. **Verified against the live runtime deferred-tool list (this session):** the runtime exposes `mcp__zai-mcp-server__analyze_image` and `mcp__zai-mcp-server__analyze_video`; `image_analysis` / `video_analysis` do NOT exist in the runtime. The correction direction is therefore confirmed (not a bare assertion). Impact: the documented `ToolSearch(query: "select:mcp__zai-mcp-server__image_analysis")` preload (line 50) and the subsequent MCP call hard-fail under a GLM backend, breaking the documented image-read fallback path.

Correction:
- `image_analysis` → `analyze_image` (lines 33, 35, 50, 65 — the "default" tool)
- `video_analysis` → `analyze_video` (line 72)

The other 6 vision tool names are **correct and MUST NOT be changed**: `analyze_data_visualization`, `diagnose_error_screenshot`, `extract_text_from_screenshot`, `ui_diff_check`, `ui_to_artifact`, `understand_technical_diagram`.

**Forward-note (NOT fixed here — out of scope):** `internal/template/templates/.moai/config/sections/tool-policy.yaml:776` carries the SAME `image_analysis` error (in the `audit:` field describing the GLM Read-on-image replacement). This is a config artifact, outside this SPEC's `.claude/rules/**` charter — recorded as a follow-up for VERSION-FORMAT-001 or a dedicated config SPEC. See §F Out of Scope.

### D2 — verification-claim-integrity.md inverted-logic HARD clause

`.claude/rules/moai/core/verification-claim-integrity.md` line 9:

> `[ZONE:Evolvable] [HARD] No actor MUST assert a verification, a completion, **OR a defect / debt / drift** it did not actually verify with the domain's mechanical tooling.`

"No actor MUST assert …" literally reads as "no actor is *obligated* to assert", inverting the intended prohibition. The intent is the opposite — actors are *forbidden* from asserting unobserved claims.

Correction: rewrite the modal to `An actor MUST NOT assert …`, preserving the rest of the clause verbatim (the `**OR a defect / debt / drift**` emphasis and the "did not actually verify with the domain's mechanical tooling" tail).

**Mirror divergence caveat (verified):** the deployed and template copies of this file are **intentionally non-identical** (template scrubs internal SPEC-IDs / coverage numbers per CLAUDE.local.md §25 — confirmed via `diff`). Line 9 (the clause being corrected) IS byte-identical across both copies, so the same edit applies to both. The run-phase MUST apply ONLY the line-9 token correction to each mirror and MUST NOT reconcile any other pre-existing divergence — that divergence is by design.

### D3 — cpp.md compiler-flag typo

`.claude/rules/moai/languages/cpp.md` line 110: `-fsanitize=adddess,undefined` → the sanitizer name is misspelled. Correction: `-fsanitize=address,undefined`.

### D4 — karpathy-quickref.md broken skill cross-reference

`.claude/rules/moai/development/karpathy-quickref.md` line 47 reads `For concrete wrong/right code examples, see skill: \`moai-reference-anti-patterns\``. The named skill does not exist; the actual artifact is a reference file at `.claude/skills/moai/references/anti-patterns.md` (NOT a skill). Correction: a **prose rewrite**, not a token-only swap — the word `skill:` must change too, e.g. `For concrete wrong/right code examples, see \`.claude/skills/moai/references/anti-patterns.md\``. The post-edit text MUST NOT retain the `skill:` label (the target is a reference file, not a skill).

### D5 — ci-watch-protocol.md broken handoff-schema cross-reference

`.claude/rules/moai/workflow/ci-watch-protocol.md` line 100 reads `On exit 2, stdout contains JSON (see \`modules/trigger-handoff.md\` in the skill):`. The ci-loop skill (`.claude/skills/moai-workflow-ci-loop/`) has NO `modules/` directory; the file does not exist. The real handoff schema lives inside `.claude/skills/moai-workflow-ci-loop/SKILL.md` at the **bold inline marker** `**Handoff schema on exit 2**` (line 97 — it is a `**bold**` inline marker, NOT a `§` / `##` heading) (Go source: `internal/ciwatch/handoff.go::Handoff`). Correction: a **prose rewrite**, not a token-only swap — repoint to the SKILL.md and reference the bold marker by its correct label, e.g. `... see \`.claude/skills/moai-workflow-ci-loop/SKILL.md\` (the **Handoff schema on exit 2** marker)`. Do NOT label it a `§` section.

### D6 — two `workflow/` cross-reference path errors

Two rules reference files under a `workflow/` path that does not hold them:
- `.claude/rules/moai/workflow/branch-origin-protocol.md` — referrer: `worktree-integration.md:320`. Actual location: `.claude/rules/moai/development/branch-origin-protocol.md`.
- `.claude/rules/moai/workflow/git-workflow-doctrine.md` — referrer: `manager-develop-prompt-template.md:108`. Actual location: `.moai/docs/git-workflow-doctrine.md`.

Correction: fix the path in each referrer to point at the real file.

---

## §C. Requirements (GEARS)

- **REQ-RHF-001** (Ubiquitous) — The glm-web-tooling rule shall name the live z.ai vision tools `analyze_image` and `analyze_video` (replacing the inverted `image_analysis` / `video_analysis`), and shall leave the 6 correct sibling tool names unchanged.
- **REQ-RHF-002** (Ubiquitous) — The verification-claim-integrity rule shall state the invariant as a prohibition (`An actor MUST NOT assert …`), preserving the remainder of the clause's meaning.
- **REQ-RHF-003** (Ubiquitous) — The cpp language rule shall spell the sanitizer flag `-fsanitize=address`.
- **REQ-RHF-004** (Ubiquitous) — The karpathy-quickref rule shall reference the anti-patterns examples at `.claude/skills/moai/references/anti-patterns.md` as a reference file (prose rewrite — the `skill:` label removed, since the target is a reference file, not a skill).
- **REQ-RHF-005** (Ubiquitous) — The ci-watch-protocol rule shall reference the handoff schema at the existing `.claude/skills/moai-workflow-ci-loop/SKILL.md` **`**Handoff schema on exit 2**`** bold inline marker (NOT labelled a `§` section; prose rewrite).
- **REQ-RHF-006** (Ubiquitous) — Every rule that references `branch-origin-protocol.md` shall use the path `.claude/rules/moai/development/branch-origin-protocol.md`; every rule that references `git-workflow-doctrine.md` shall use the path `.moai/docs/git-workflow-doctrine.md`.
- **REQ-RHF-007** (Ubiquitous, Template-First) — For every edited file under `.claude/rules/**`, the rule shall apply the corrected token/line to its mirror under `internal/template/templates/.claude/rules/**`, and shall NOT reconcile any pre-existing intentional neutrality divergence to byte-identity (the divergence in `verification-claim-integrity.md` is by design per CLAUDE.local.md §25).
- **REQ-RHF-008** (Event-driven) — **When** the template source files have been edited, the build shall re-embed the corrected template tree at `go build ./...` time via the `//go:embed all:templates` directive (`internal/template/embed.go:28`); there is NO `embedded.go` to regenerate. Mirror parity shall be verified by the existing Go test `TestRuleTemplateMirrorDrift`.
- **REQ-RHF-009** (Ubiquitous, neutrality) — The rule shall introduce no internal SPEC IDs, internal work dates, or commit SHAs into any `internal/template/templates/**` content (CLAUDE.local.md §25).

---

## §D. Acceptance Criteria (inline — Tier S LEAN)

Each AC is grep-verifiable in BOTH the deployed path and the template mirror unless noted. Run from repo root `/Users/goos/MoAI/moai-adk-go`.

### AC-RHF-001 — D1 vision tool names corrected (deployed + template)

```bash
# Inverted names fully removed in both copies:
grep -c 'image_analysis\|video_analysis' \
  .claude/rules/moai/core/glm-web-tooling.md \
  internal/template/templates/.claude/rules/moai/core/glm-web-tooling.md
# Expected: 0 for BOTH files

# Live names present in both copies:
grep -c 'analyze_image' .claude/rules/moai/core/glm-web-tooling.md         # Expected: >=4
grep -c 'analyze_video' .claude/rules/moai/core/glm-web-tooling.md         # Expected: >=1
grep -c 'analyze_image' internal/template/templates/.claude/rules/moai/core/glm-web-tooling.md  # Expected: >=4
grep -c 'analyze_video' internal/template/templates/.claude/rules/moai/core/glm-web-tooling.md  # Expected: >=1

# The 6 correct sibling tools UNCHANGED (regression guard):
grep -c 'analyze_data_visualization\|diagnose_error_screenshot\|extract_text_from_screenshot\|ui_diff_check\|ui_to_artifact\|understand_technical_diagram' \
  .claude/rules/moai/core/glm-web-tooling.md
# Expected: unchanged from pre-edit baseline (6 distinct tool rows intact)
```

### AC-RHF-002 — D2 inverted-logic clause rewritten (deployed + template)

```bash
# Prohibition phrasing present:
grep -c 'An actor MUST NOT assert' \
  .claude/rules/moai/core/verification-claim-integrity.md \
  internal/template/templates/.claude/rules/moai/core/verification-claim-integrity.md
# Expected: 1 for BOTH files

# Inverted phrasing removed:
grep -c 'No actor MUST assert' \
  .claude/rules/moai/core/verification-claim-integrity.md \
  internal/template/templates/.claude/rules/moai/core/verification-claim-integrity.md
# Expected: 0 for BOTH files

# Tail meaning preserved:
grep -c "did not actually verify with the domain's mechanical tooling" \
  .claude/rules/moai/core/verification-claim-integrity.md
# Expected: 1
```

### AC-RHF-003 — D3 sanitizer typo fixed (deployed + template)

```bash
grep -c 'adddess' \
  .claude/rules/moai/languages/cpp.md \
  internal/template/templates/.claude/rules/moai/languages/cpp.md
# Expected: 0 for BOTH files

grep -c 'fsanitize=address' \
  .claude/rules/moai/languages/cpp.md \
  internal/template/templates/.claude/rules/moai/languages/cpp.md
# Expected: 1 for BOTH files
```

### AC-RHF-004 — D4 anti-patterns cross-reference repointed (deployed + template)

```bash
grep -c 'moai-reference-anti-patterns' \
  .claude/rules/moai/development/karpathy-quickref.md \
  internal/template/templates/.claude/rules/moai/development/karpathy-quickref.md
# Expected: 0 for BOTH files

grep -c 'skills/moai/references/anti-patterns.md' \
  .claude/rules/moai/development/karpathy-quickref.md \
  internal/template/templates/.claude/rules/moai/development/karpathy-quickref.md
# Expected: 1 for BOTH files

# Prose-rewrite guard: the line must NOT retain the 'see skill:' label (target is a reference file):
grep -c 'see skill:' .claude/rules/moai/development/karpathy-quickref.md
# Expected: 0 (the 'skill:' label was removed in the prose rewrite)
# (note: 'see skill:' may legitimately appear elsewhere in this file for real skills;
#  the run-phase MUST scope this check to L47 — confirm the edited line no longer says 'skill:')

# Target existence guard:
test -f .claude/skills/moai/references/anti-patterns.md && echo OK   # Expected: OK
```

### AC-RHF-005 — D5 handoff-schema cross-reference repointed (deployed + template)

```bash
grep -c 'modules/trigger-handoff.md' \
  .claude/rules/moai/workflow/ci-watch-protocol.md \
  internal/template/templates/.claude/rules/moai/workflow/ci-watch-protocol.md
# Expected: 0 for BOTH files

# New reference points at the real SKILL.md handoff marker (both copies):
grep -c 'moai-workflow-ci-loop/SKILL.md' \
  .claude/rules/moai/workflow/ci-watch-protocol.md \
  internal/template/templates/.claude/rules/moai/workflow/ci-watch-protocol.md
# Expected: >=1 for BOTH files

# Prose-rewrite guard: the new text references the bold marker by its correct label,
# and does NOT call it a '§' section (it is a **bold** inline marker, not a heading):
grep -c 'Handoff schema on exit 2' .claude/rules/moai/workflow/ci-watch-protocol.md   # Expected: >=1

# Target existence guard (the bold marker literal exists in the skill at L97):
grep -c '\*\*Handoff schema on exit 2\*\*' .claude/skills/moai-workflow-ci-loop/SKILL.md   # Expected: >=1
```

### AC-RHF-006 — D6 two path errors corrected (deployed + template)

```bash
# Wrong paths fully eliminated across deployed + template rules trees.
# `grep -rc` prints one `<file>:<count>` line per file; piping through `grep -v ':0$'`
# drops every file whose count is 0. PASS predicate: the piped command prints NOTHING
# (i.e. every file's count is exactly 0 → every line ends in ':0' → all filtered out).
grep -rc 'workflow/branch-origin-protocol.md' \
  .claude/rules/ internal/template/templates/.claude/rules/ 2>/dev/null | grep -v ':0$'
# PASS = zero output lines (every matched file printed `:0`, all filtered away)

grep -rc 'workflow/git-workflow-doctrine.md' \
  .claude/rules/ internal/template/templates/.claude/rules/ 2>/dev/null | grep -v ':0$'
# PASS = zero output lines (every matched file printed `:0`, all filtered away)

# Correct paths present at the two referrer sites (deployed):
grep -c 'development/branch-origin-protocol.md' .claude/rules/moai/workflow/worktree-integration.md   # Expected: >=1
grep -c '.moai/docs/git-workflow-doctrine.md' .claude/rules/moai/development/manager-develop-prompt-template.md  # Expected: >=1

# Target existence guards:
test -f .claude/rules/moai/development/branch-origin-protocol.md && echo OK   # Expected: OK
test -f .moai/docs/git-workflow-doctrine.md && echo OK                        # Expected: OK
```

### AC-RHF-007 — template SOURCE corrected + re-embed via `go build` (NO embedded.go)

There is NO `internal/template/embedded.go` — the template tree is embedded at compile
time via `//go:embed all:templates` (`internal/template/embed.go:28`). Verification is
(a) the template SOURCE file carries the corrected token, and (b) `go build` succeeds
(re-embedding the corrected tree).

```bash
# Premise guard — confirm embedded.go genuinely does not exist (the corrected premise):
find internal/template -name embedded.go    # Expected: NO output (file does not exist)

# (a) Corrected token present in the TEMPLATE SOURCE (not a non-existent embedded.go):
grep -c 'analyze_image' internal/template/templates/.claude/rules/moai/core/glm-web-tooling.md   # Expected: >=4
grep -c 'image_analysis' internal/template/templates/.claude/rules/moai/core/glm-web-tooling.md  # Expected: 0

# (b) go build succeeds → the //go:embed directive re-embeds the corrected template tree:
go build ./...    # Expected: exit 0

# (c) deployed↔template mirror parity verified by the existing Go test:
go test ./internal/template/... -run TestRuleTemplateMirrorDrift   # Expected: PASS
```

### AC-RHF-008 — template neutrality preserved (no internal-content leak)

```bash
# No internal SPEC ID / commit SHA / internal date introduced into edited template files:
grep -rEc 'SPEC-V3R6-RULES-HOTFIX-001|SPEC-V3R6-[A-Z]' \
  internal/template/templates/.claude/rules/moai/core/glm-web-tooling.md \
  internal/template/templates/.claude/rules/moai/core/verification-claim-integrity.md \
  internal/template/templates/.claude/rules/moai/languages/cpp.md \
  internal/template/templates/.claude/rules/moai/development/karpathy-quickref.md \
  internal/template/templates/.claude/rules/moai/workflow/ci-watch-protocol.md \
  internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md \
  internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md 2>/dev/null \
  | grep -v ':0$'
# Expected: NO output (no introduced SPEC-ID tokens)

# CI neutrality gate passes (the canonical guard):
go test ./internal/template/... -run TestTemplateNeutralityAudit
# Expected: PASS (or the test name in the suite — confirm exact name during run-phase)
```

### AC-RHF-009 — no unintended behavior change (Go build + vet)

```bash
go build ./...        # Expected: clean (no Go source changed; //go:embed re-embeds corrected template tree)
go vet ./...          # Expected: clean
go test ./internal/template/...   # Expected: PASS (includes TestRuleTemplateMirrorDrift mirror-parity)
```

---

## §E. Self-Verification (plan-phase)

- [x] SPEC ID `SPEC-V3R6-RULES-HOTFIX-001` passes the canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (decomposition printed in agent turn: `SPEC ✓ | V3R6 ✓ | RULES ✓ | HOTFIX ✓ | 001 ✓ → PASS`).
- [x] Directory format `.moai/specs/SPEC-V3R6-RULES-HOTFIX-001/` (not flat file).
- [x] ID uniqueness verified (no prior `.moai/specs/SPEC-V3R6-RULES-HOTFIX-001/`; no grep match in `.moai/specs/`).
- [x] All 6 defects verified against deployed + template mirror with cited file:line.
- [x] `embedded.go` premise purged (verified absent via `find`; real mechanism `//go:embed all:templates` at `embed.go:28`; parity via `TestRuleTemplateMirrorDrift`).
- [x] D2-file intentional neutrality divergence documented (no byte-identity reconcile mandate).
- [x] D1 cited as verified against live-runtime deferred-tool list; tool-policy.yaml:776 sibling error recorded as forward-note.
- [x] 12 canonical frontmatter fields present; `created`/`updated` (not `created_at`/`updated_at`); `tags` comma-separated string (not `labels` array); `version` quoted string.
- [x] GEARS format used (Ubiquitous + one Event-driven REQ); no legacy `IF/THEN`.
- [x] Exclusions section present with ≥1 entry (glossary deferral + 3 sibling-SPEC fences + 6 correct vision tools).
- [x] No implementation HOW in requirements (function names / schemas) — observable behavior only.

---

## §F. Exclusions (What NOT to Build)

This SPEC is deliberately narrow.

### Out of Scope

- **Terminology Glossary authoring (DEFERRED — forward-link).** `worktree-state-guard.md:19`, `spec-workflow.md:23`, and `CLAUDE.md §14` all reference a `worktree-integration.md` § "Terminology Glossary" (L1/L2/L3 layer definitions) that does NOT exist — `worktree-integration.md` has only a `## Sentinel Key Glossary`. **Decision: DEFER.** Creating the L1/L2/L3 glossary is *new content authoring* (multiple definitions describing the worktree isolation layers), not a single-token pointer fix, so it falls outside this hotfix SPEC's "low-risk single-token / broken-cross-reference" charter. It is recorded here as a dependency/forward-link for a follow-up structure SPEC (candidate owner: CATALOG-SCRUB or a dedicated worktree-doc SPEC). This SPEC does NOT repoint the 3 referrers either, because the correct target section does not yet exist — repointing now would just move the broken reference.
- **Archived-agent scrub** — owned by the sibling SPEC-V3R6-CATALOG-SCRUB. Do not touch archived-agent references here.
- **SSOT de-duplication** — owned by the sibling SPEC-V3R6-SSOT-DEDUP. Do not consolidate duplicated rule content here.
- **Version/format normalization** — owned by the sibling SPEC-V3R6-VERSION-FORMAT. Do not normalize `Version:` footers or frontmatter format here.
- **The 6 correct z.ai vision tool names** (`analyze_data_visualization`, `diagnose_error_screenshot`, `extract_text_from_screenshot`, `ui_diff_check`, `ui_to_artifact`, `understand_technical_diagram`) — these are correct and MUST NOT be modified.
- **`tool-policy.yaml:776` `image_analysis` error (DEFERRED — forward-note).** `internal/template/templates/.moai/config/sections/tool-policy.yaml:776` carries the same inverted `image_analysis` token in its `audit:` field. This is a CONFIG artifact (`.moai/config/sections/**`), outside this SPEC's `.claude/rules/**` charter. Recorded as a follow-up for the sibling SPEC-V3R6-VERSION-FORMAT or a dedicated config-hotfix SPEC — NOT fixed here.
- **Any Go logic change** beyond the `//go:embed`-driven re-embed at `go build` time — this is a rule-content-only SPEC. There is NO `embedded.go` codegen step.

---

## §G. HARD Constraints

- **[HARD] Template-First (CLAUDE.local.md §2):** every edited `.claude/rules/**` file MUST have the corrected token/line applied to its `internal/template/templates/.claude/rules/**` mirror, then `go build ./...` re-embeds the corrected tree via `//go:embed all:templates`. Acceptance criteria verify BOTH copies + `TestRuleTemplateMirrorDrift`.
- **[HARD] No byte-identity reconciliation:** the `verification-claim-integrity.md` deployed↔template copies are intentionally divergent (neutrality scrub per CLAUDE.local.md §25). Apply ONLY the corrected line-9 token to each; do NOT reconcile the rest to byte-identity.
- **[HARD] Template neutrality (CLAUDE.local.md §25):** no internal SPEC IDs / internal dates / commit SHAs introduced into template content. The corrections are all generic tool-name / typo / path / prose-reference fixes — neutral by construction.
- **[HARD] Rule-content only:** no Go logic change. There is NO `embedded.go` to regenerate; `go build` re-embeds the corrected template source via `//go:embed all:templates` (`internal/template/embed.go:28`).

---

## §H. Cross-References

- CLAUDE.local.md §2 (Template-First Rule), §25 (Template Internal-Content Isolation)
- `.claude/rules/moai/core/glm-web-tooling.md` (D1 target)
- `.claude/rules/moai/core/verification-claim-integrity.md` (D2 target)
- `.claude/rules/moai/languages/cpp.md` (D3 target)
- `.claude/rules/moai/development/karpathy-quickref.md` (D4 target)
- `.claude/rules/moai/workflow/ci-watch-protocol.md` (D5 target) → `.claude/skills/moai-workflow-ci-loop/SKILL.md` `**Handoff schema on exit 2**` bold marker (L97)
- `.claude/rules/moai/workflow/worktree-integration.md` (D6 branch-origin referrer)
- `.claude/rules/moai/development/manager-develop-prompt-template.md` (D6 git-workflow-doctrine referrer)
- `internal/template/embed.go:28` (`//go:embed all:templates` — the real embed mechanism; NO `embedded.go`)
- `internal/template/rule_template_mirror_test.go` (`TestRuleTemplateMirrorDrift` — deployed↔template parity gate)
- `internal/template/templates/.moai/config/sections/tool-policy.yaml:776` (D1 sibling `image_analysis` error — forward-note, out of scope)
- Sibling cohort SPECs: CATALOG-SCRUB, SSOT-DEDUP, VERSION-FORMAT (scope fences)
