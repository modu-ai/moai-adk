---
id: SPEC-STEERING-ALIGN-LOCAL-DIET-001
title: "Steering-Align: CLAUDE.local.md always-loaded body conservative diet (verified external-SSOT pointer-ization + stale cross-ref correction, §19.1 HUMAN GATE preserved)"
version: "1.0.0"
status: completed
created: 2026-06-23
updated: 2026-06-23
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "CLAUDE.local.md"
lifecycle: spec-anchored
tags: "steering, claude-local, diet, always-loaded, context-budget, pointer-ization, conservative, stale-cross-ref"
tier: S
era: V3R6
---

## HISTORY

- 2026-06-23 — v0.1.0 — manager-spec — Plan-phase artifacts authored (Tier S, Section A-H, 4 artifacts: spec.md + plan.md + acceptance.md + progress.md skeleton). FINAL SPEC (P6, 5 of 5) in Epic Steering-Align. Applies the same per-line-test diet doctrine validated by P2 CLAUDEMD-DIET-001 (COMPLETED) and P5 OUTPUT-STYLE-SLIM-001 (COMPLETED, origin ed8172482) to the maintainer-local always-loaded `CLAUDE.local.md` body. Diet bound is **CONSERVATIVE** (user-confirmed): pointer-ize ONLY verified external-SSOT duplication; preserve ALL dev-local-unique operational knowledge; the §19.1 구현 착수 승인 [HARD] HUMAN GATE body is KEPT (user decision). `status: draft`.
- 2026-06-23 — v1.0.0 — orchestrator-direct — Sync-phase 3-phase close. Run-phase diet applied (51b4c495e: `CLAUDE.local.md` 806→796L net −10L, 7/7 AC PASS, §19.1 [HARD] HUMAN GATE body 보존, dev-local 고유지식 전량 보존). plan-auditor iter-2 PASS 0.89 (D1-D4 전부 해소). `status: in-progress → completed`. **Epic Steering-Align 5/5 (P6, FINAL) — Epic 완결.** era V3R6 (H-override). sync orchestrator-direct (P5 OUTPUT-STYLE-SLIM 동일 패턴, Tier S 경량).

---

## A. Context / Background

### A.1 The doctrine being applied

This SPEC is the FINAL entry in Epic Steering-Align — the roadmap aligning moai-adk to Anthropic's official Claude Code "steering" guidance. The two canonical statements that anchored every prior Epic SPEC anchor this one too:

1. **best-practices — "Write an effective always-loaded instruction surface"**: the per-line test — *"Would removing this cause Claude to make mistakes? If not, cut it."* — and the bloat warning — *"Bloated always-loaded instructions cause Claude to ignore your actual instructions."*
2. **blog — "Steering Claude Code: skills, hooks, rules, subagents and more"**: distinguishes ALWAYS-LOADED steering surfaces (CLAUDE.md + unscoped rules + output-style + the project-local instruction file) from CONDITIONALLY-loaded surfaces.

`CLAUDE.local.md` is the maintainer's **private project instructions** file (loaded as project instructions at every session launch — it appears in the session-launch context alongside `CLAUDE.md`). It is therefore an ALWAYS-LOADED surface and a legitimate per-line-test diet target, exactly like the `CLAUDE.md` body P2 dieted and the output-style `moai.md` body P5 dieted.

### A.2 Observed ground-truth (re-verified live; commands + output in §F.1)

| Metric | Value |
|--------|-------|
| Line count | **806** |
| Byte count | **33939** |
| Section structure | `## 1.` .. `## 25.` (25 top-level sections) + Status/Version footer |
| Git-tracked? | **YES** — `CLAUDE.local.md` IS git-tracked (verified: `git ls-files CLAUDE.local.md` → exit 0; committed history `8e78530bb` / `de13ecc4c` / `96fad88ff`; `git check-ignore -v CLAUDE.local.md` → exit 1 = NOT ignored; `.gitignore` carries no `CLAUDE.local` entry). The file's own header (L1-5) says only "Audience: GOOS (local developer only)" — it makes NO "not checked in" claim. See §A.5 lifecycle note. |

### A.3 The CONSERVATIVE diet bound (user-confirmed this session)

[HARD] The diet is bounded **CONSERVATIVE** (user-confirmed via AskUserQuestion). This is STRONGER preservation than P5's MODERATE bound. The two binding user decisions:

1. **Pointer-ize ONLY verified external-SSOT duplication.** Preserve ALL dev-local-unique operational knowledge. The estimated reduction is **~25-35 lines** — a GUIDANCE figure, NOT a hard requirement. Do NOT set an aggressive numeric target.
2. **§19.1 (구현 착수 승인 / Implementation Kickoff Approval, plan-to-implement HUMAN GATE [HARD]) = KEEP BODY.** Only minor header compression is allowed. The `[HARD]` HUMAN GATE policy MUST stay visible in always-loaded context every session. §19.1 MUST NOT be pointer-ized.

This SPEC inherits the P5 over-cut lesson explicitly: **behavioral-PASS over numeric-proxy.** P5's original estimate was −150 to −250 lines but the actual honest reduction was only −26 lines, because preservation invariants forced fewer cuts — and over-cut avoidance was the correct answer. The same discipline binds here: an honest CONSERVATIVE diet that removes only verified duplication is the goal; hitting a line number is NOT.

### A.4 Verified pointer-ize / correction candidates (ground-truth, §F.1)

Only TWO edits qualify under the CONSERVATIVE bound. Both were tool-verified before this SPEC was authored (the "this is duplicated / this cross-ref is stale" claim is a defect claim that MUST be tool-verified per verification-claim-integrity.md §1.1):

| # | Candidate | Verified finding | Mechanism |
|---|-----------|------------------|-----------|
| 1 | **Pre-PR Verification checklist** (CLAUDE.local.md L108-118) | A 7-item template-contributor checklist that DUPLICATES the canonical 5-item self-check in `.moai/docs/template-internal-isolation-doctrine.md §25.3`. The doctrine §25.3 ("Pre-commit Self-Check (5-item Mandatory Checklist)") + §25.1 (Allowed/Forbidden content-class catalogue) are VERIFIED to exist (§F.1). Line 118 ALREADY carries the pointer (`See .moai/docs/template-internal-isolation-doctrine.md §25.3 ...`). | **M-POINTER** — collapse the duplicated 7-item checklist body to the one-line pointer that already exists at L118 (keep the pointer, remove the duplicated bullet list). |
| 2 | **§2.1 Template Content Neutrality cross-ref** (CLAUDE.local.md L106) | The prose cites `.claude/rules/moai/development/coding-standards.md § MUST` and "C1/C2/C4/C5/C6/C8 per coding-standards.md MUST constraints". VERIFIED: coding-standards.md has NO `§ MUST` section (its headers are Language Policy / File Size Limits / Content Restrictions / Footer Convention / Duplicate Prevention / Thin Command Pattern / CC Version Compatibility / Paths Frontmatter / Bash Risk-Amplifier) and NO C1-C8 content-class definitions (grep returned 0 matches, §F.1). The actual C1-C8 canonical lives in `template-internal-isolation-doctrine.md §25.1`. | **STALE-CROSS-REF CORRECTION + M-POINTER** — fix the broken cross-ref to point at `template-internal-isolation-doctrine.md §25.1 / §25` (which the file's §25 stub already references) AND compress. This is a correction-plus-pointer, NOT a blind delete. |

No third candidate qualifies. Sections §1, §3-§16, §17/§18/§21/§23/§24/§25, §19 Quick Pointer + §19.1 body, §20, §22 are all dev-local-unique operational knowledge or already-stubbed pointers (§A.6 preserve map) and are out of the CONSERVATIVE diet scope.

### A.5 Lifecycle note — CLAUDE.local.md is git-tracked (close mechanics identical to P2/P5)

`CLAUDE.local.md` IS git-tracked (verified §A.2 / §F.1). The 3-phase close (plan → run → sync) is therefore the SIMPLE, standard case — identical to how P2 (CLAUDEMD-DIET-001) closed with `CLAUDE.md` in-diff and P5 (OUTPUT-STYLE-SLIM-001) closed with `.claude/output-styles/moai/moai.md` in-diff. Both of those primary-edited files are likewise git-tracked (`git ls-files CLAUDE.md` and `git ls-files .claude/output-styles/moai/moai.md` both return the path). There is NO special untracked-file handling:

- The SPEC artifacts under `.moai/specs/SPEC-STEERING-ALIGN-LOCAL-DIET-001/` (spec.md + plan.md + acceptance.md + progress.md) ARE git-tracked — prior Epic SPECs committed their artifacts identically.
- The run-phase commit + sync-phase close commit carry the SPEC artifacts + progress.md **AND** the `CLAUDE.local.md` diet edit — `CLAUDE.local.md` appears in the commit diff alongside the SPEC artifacts (it is a tracked file).
- The `sync_commit_sha` recorded in progress.md §E.4 DOES point at a commit whose diff contains `CLAUDE.local.md` (plus the SPEC artifacts) — exactly the P2/P5 pattern.
- AC verification (acceptance.md) reads the live `CLAUDE.local.md` on disk, which is the authoritative state; for a tracked file the live state equals the committed state after the edit lands.

### A.6 Preserve map (dev-local-unique — out of CONSERVATIVE diet scope)

The following are explicitly PRESERVED (verbatim, except where §19.1 allows minor header compression):

| Section | Why preserved |
|---------|---------------|
| §1 (moai CLI vs /moai slash command) | dev-local-unique disambiguation — no external SSOT |
| §2 core (Protected Dirs, Template Source, Template-First, Local-Only Files, settings.local.json Separation, OTEL-in-tests, Embedded Template System) | dev-local operational rules |
| §2.1 (Template Content Neutrality) | KEEP the section; only CORRECT its stale cross-ref + compress (candidate #2) |
| §2 Pre-PR checklist (L108-118) | POINTER-ize the duplicated body (candidate #1); the section itself stays |
| §3–§16 | operational rules: Go test isolation, version mgmt, hook dev, template variable strategy, build commands, frequent issues, YAML frontmatter, GLM testing, hardcoding prohibition, language neutrality, orchestrator self-check — all dev-local-unique |
| §17 / §18 / §21 / §23 / §24 / §25 | already stub+pointer pattern — leave verbatim |
| §19 Quick Pointer table + Local Notes | cross-ref index — keep |
| **§19.1 body** | **KEEP per user decision** — [HARD] HUMAN GATE policy stays in always-loaded context; minor header compression only |
| §20 (Vercel guard), §22 (Dev Settings Intent) | settings.json intent — explicit preserve |

### A.7 Epic Steering-Align context

This is **SPEC 5 of 5** (P6) in Epic Steering-Align. The roadmap (named for Epic context ONLY — this SPEC authors artifacts for the CLAUDE.local.md diet alone):

1. **SPEC-STEERING-ALIGN-RULE-SCOPING-001** (COMPLETED, origin ab81e7f42) — path-scope file-touch-triggered always-loaded rules.
2. **SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001** (COMPLETED, P2, origin d0ca1f214) — per-line-test diet of CLAUDE.md's always-loaded body (M-DELETE + M-POINTER + AC-009 over-cut gate).
3. **SPEC-STEERING-ALIGN-GUARDRAIL-HOOK-001** (COMPLETED, P3, origin c463257b8) — deterministic SessionStart guardrail hook + always-load removal.
4. **SPEC-STEERING-ALIGN-OUTPUT-STYLE-SLIM-001** (COMPLETED, P5, origin ed8172482) — output-style `moai.md` body diet (782→756L, −26L; over-cut avoidance lesson).
5. **SPEC-STEERING-ALIGN-LOCAL-DIET-001** (THIS SPEC, P6, FINAL) — CONSERVATIVE diet of the maintainer-local `CLAUDE.local.md` always-loaded body.

---

## B. Requirements (GEARS notation)

### B.1 Verified external-SSOT pointer-ization (M-POINTER)

- **REQ-LD-001 (Ubiquitous)**: The `CLAUDE.local.md` §2 Pre-PR Verification checklist (L108-118) — a 7-item template-contributor checklist that duplicates the canonical 5-item self-check in `.moai/docs/template-internal-isolation-doctrine.md §25.3` — SHALL be condensed to the one-line pointer that already exists at L118 (`See .moai/docs/template-internal-isolation-doctrine.md §25.3 ...`). The duplicated bullet-list body SHALL be removed; the pointer SHALL be preserved (and MAY be lightly expanded to name §25.1 as the content-class catalogue). The §2 section heading and the §2.1 prose framing SHALL remain.

- **REQ-LD-002 (Event-driven)**: **When** the run-phase is about to apply the REQ-LD-001 POINTER edit, the run-phase SHALL FIRST re-grep `.moai/docs/template-internal-isolation-doctrine.md` for the §25.3 5-item self-check heading AND the §25.1 content-class catalogue heading; **When** the grep returns 0 hits of either canonical anchor, the run-phase SHALL block the POINTER edit and reclassify the passage as KEEP (return a blocker report). This is the P2 D1 over-cut defense (AC-CMD-009) / P5 AC-OSS-006 pattern applied to the local-diet file — a "this is duplicated elsewhere" claim is a defect claim that MUST be tool-verified before deletion (verification-claim-integrity.md §1.1).

### B.2 Stale cross-reference correction (CORRECTION + M-POINTER)

- **REQ-LD-003 (Ubiquitous)**: The `CLAUDE.local.md` §2.1 Template Content Neutrality prose (L106) SHALL be corrected so that it no longer cites the non-existent `.claude/rules/moai/development/coding-standards.md § MUST` section nor the non-existent "C1/C2/C4/C5/C6/C8 per coding-standards.md MUST constraints". The corrected prose SHALL point at the ACTUAL canonical owner of the C1-C8 content classes — `.moai/docs/template-internal-isolation-doctrine.md §25.1 / §25` — which the file's own §25 stub already references. The correction MAY compress the §2.1 prose but SHALL preserve the section's operational meaning (template content neutrality obligation + the CI guard reference `.github/workflows/template-neutrality-check.yaml`).

- **REQ-LD-004 (Unwanted behavior)**: The corrected §2.1 cross-ref SHALL NOT introduce any NEW broken cross-reference. **When** the run-phase rewrites the §2.1 cross-ref, it SHALL verify the new target anchor (`template-internal-isolation-doctrine.md §25.1`) exists on disk before committing the edit (grep, per verification-claim-integrity.md §1.1).

### B.3 §19.1 HUMAN GATE body preservation (the load-bearing user decision)

- **REQ-LD-005 (Unwanted behavior)**: The diet SHALL NOT pointer-ize, summarize away, or remove the §19.1 구현 착수 승인 (Implementation Kickoff Approval, plan-to-implement HUMAN GATE) body. The `[HARD]` directive line, the skip-eligible-scope clarification, the 4-step orchestrator obligation list, and the violation anti-pattern SHALL all remain present and renderable in always-loaded context. Only minor header compression (e.g., trimming the "(renamed from GATE-2)" parenthetical or the trailing SPEC-reference footnote) is permitted. The `[HARD]` token on the §19.1 directive line SHALL be preserved.

### B.4 Dev-local-unique knowledge preservation

- **REQ-LD-006 (Unwanted behavior)**: The diet SHALL NOT remove or pointer-ize any dev-local-unique operational knowledge (§A.6 preserve map): §1, §3–§16, §17/§18/§21/§23/§24/§25, §19 Quick Pointer + Local Notes, §20, §22. These sections carry maintainer-machine-specific or moai-adk-development-specific knowledge with no external SSOT; they are out of the CONSERVATIVE diet scope. The two §2 candidate edits (REQ-LD-001, REQ-LD-003) are the ONLY content changes.

### B.5 Derived target (range, not a hard number)

- **REQ-LD-007 (Ubiquitous)**: The final `CLAUDE.local.md` line count SHALL be a RANGE derived from the two qualifying edits (plan.md §C), NOT an arbitrary fixed number. The "~25-35 lines reduction (806 → ~771-781L)" figure is GUIDANCE only. No dev-local-unique content SHALL be cut to hit a line number; an honest CONSERVATIVE diet that lands above ~781L is acceptable if only the two verified candidates qualify. The derived target + its justification SHALL be stated in plan.md §C. This is the P5 over-cut lesson applied (behavioral-PASS over numeric-proxy).

---

## C. Constraints

- **C-1** `CLAUDE.local.md` is git-tracked (§A.5 / §F.1). The run/sync commits carry the SPEC artifacts + progress.md AND the `CLAUDE.local.md` diet edit (it appears in the commit diff, exactly like P2's CLAUDE.md and P5's moai.md). `sync_commit_sha` points at a commit whose diff contains `CLAUDE.local.md`. AC verification reads the live on-disk file (for a tracked file, live == committed state after the edit lands). Standard close mechanics — no special handling.
- **C-2** [HARD] CONSERVATIVE bound (user-confirmed): pointer-ize ONLY the two verified external-SSOT-duplication candidates (REQ-LD-001, REQ-LD-003). Preserve ALL dev-local-unique knowledge. No aggressive numeric target.
- **C-3** [HARD] §19.1 body KEEP (user decision, REQ-LD-005). Only minor header compression; the `[HARD]` HUMAN GATE policy stays visible.
- **C-4** [HARD] A candidate edit is applied ONLY after its SSOT target is tool-verified to EXIST and CARRY the content (grep, per verification-claim-integrity.md §1.1). The §2.1 correction (REQ-LD-003) FIRST verifies the broken target is genuinely absent AND the new target genuinely exists. plan.md §C records the verification command + observed output per candidate.
- **C-5** No Go code change, no new lint rule, no template edits. This SPEC is a markdown body edit of a SINGLE git-tracked maintainer-local file (`CLAUDE.local.md`) + the SPEC artifacts.
- **C-6** [HARD] `CLAUDE.local.md` is a maintainer-LOCAL file (NOT a distributed `internal/template/templates/` asset). The template neutrality rules (CLAUDE.local.md §15 / §25) do NOT apply to `CLAUDE.local.md` itself — internal SPEC IDs / dates / SHAs ARE legitimate inside it (e.g., the §19.1 body already names `SPEC-V3R6-AGENT-TEAM-REBUILD-001` / `REQ-ATR-015` legitimately). The diet MUST NOT "neutralize" these — that would be out-of-scope scope-creep and would corrupt dev-local operational knowledge.
- **C-7** Untouched paths PRESERVE (scope discipline): the diet touches ONLY `CLAUDE.local.md` §2 (two edits) + the SPEC artifacts. No other file (including the SSOT doctrine files `template-internal-isolation-doctrine.md`, `coding-standards.md`) is modified. If a candidate reveals a genuine SSOT gap, run-phase returns a blocker — it does not silently edit the SSOT under this SPEC.

---

## D. Out of Scope / Exclusions

The SPEC scope is exactly: two verified `CLAUDE.local.md` §2 edits (Pre-PR checklist pointer-ization + §2.1 stale-cross-ref correction), bounded CONSERVATIVE, with the §19.1 HUMAN GATE body preserved. The following are explicitly excluded.

### Out of Scope — §19.1 HUMAN GATE body pointer-ization or removal

- Pointer-izing, summarizing, or removing the §19.1 구현 착수 승인 (Implementation Kickoff Approval) body. The `[HARD]` HUMAN GATE policy MUST stay visible in always-loaded context every session (user decision, REQ-LD-005). Only minor header compression is permitted. This is the explicit user-confirmed preservation that distinguishes this SPEC from a blind line-count diet.

### Out of Scope — aggressive diet / dev-local-unique knowledge removal

- Cutting any dev-local-unique operational knowledge (§1, §3–§16, §17/§18/§21/§23/§24/§25, §19 Quick Pointer + Local Notes, §20, §22) to reach a line-count target. The "~25-35 lines" figure is guidance; the real target is a range derived from the two verified candidates (REQ-LD-007). Over-cutting dev-local knowledge to hit a number is the P5 over-cut failure mode and is forbidden (REQ-LD-002 gate). An aggressive multi-section condense was NOT chosen this session — the user confirmed CONSERVATIVE.

### Out of Scope — neutrality sweep of CLAUDE.local.md internal references

- "Neutralizing" the legitimate internal SPEC IDs / REQ tokens / dates inside `CLAUDE.local.md` (e.g., the §19.1 `SPEC-V3R6-AGENT-TEAM-REBUILD-001` reference). `CLAUDE.local.md` is a maintainer-LOCAL file, NOT a distributed template; the template neutrality rules (§15 / §25) bind `internal/template/templates/**`, NOT this file (C-6). Treating internal refs as forbidden here would corrupt dev-local operational knowledge and is out of scope.

### Out of Scope — SSOT doctrine / Go code / lint / template edits

- Edits to the `.moai/docs/template-internal-isolation-doctrine.md` / `.claude/rules/...` SSOT files themselves, Go code changes, new lint rules, and `internal/template/templates/` edits. Pointer-ization (M-POINTER) only edits `CLAUDE.local.md` to point AT an existing SSOT; it does NOT modify the SSOT. (If a pointer-ization reveals a genuine SSOT gap, run-phase returns a blocker — it does not silently edit the SSOT under this SPEC.)

### Out of Scope — the other Epic Steering-Align SPECs and parallel-session SPECs

- RULE-SCOPING (COMPLETED), CLAUDEMD-DIET (COMPLETED P2), GUARDRAIL-HOOK (COMPLETED P3), OUTPUT-STYLE-SLIM (COMPLETED P5) — named in §A.7 for Epic context only. This SPEC authors artifacts ONLY for the CLAUDE.local.md diet. Any parallel-session SPEC directory (e.g., `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001`) MUST NOT be touched (scope discipline, B10 untouched-paths PRESERVE).

---

## E. Acceptance Criteria Reference

Concrete GEARS-format acceptance criteria with re-runnable verification commands live in `acceptance.md`. Each AC is **behavioral** (verify the right content moved/stayed) NOT numeric-proxy (no "must remove exactly N lines" pass condition — P5 lesson). Summary of AC themes:

- **AC-LD-001** — §19.1 [HARD] HUMAN GATE body present post-diet (grep-verifiable: `[HARD]`, 구현 착수 승인, the 4-step orchestrator obligation, the violation anti-pattern all survive).
- **AC-LD-002** — Pre-PR checklist duplicated 7-item body replaced by the one-line pointer to §25.3 (grep: the §25.3 pointer survives; the duplicated `[ ] No /Users/` … bullet list is gone).
- **AC-LD-003** — §2.1 cross-ref no longer cites the non-existent `coding-standards.md § MUST` / "C1/C2/C4/C5/C6/C8 ... MUST constraints"; the corrected ref points at `template-internal-isolation-doctrine.md §25.1 / §25` which is verified to exist.
- **AC-LD-004** — dev-local-unique sections byte-preserved (§1, §5, §22 sentinel anchors + §3-§16 + §17/§18/§21/§23/§24/§25 headings all present; the §A.6 preserve map holds).
- **AC-LD-005** — line count lands in the derived range (806 → soft band ~771-781L); SOFT band — behavioral-PASS allowed if only the two candidates qualify (REQ-LD-007).
- **AC-LD-006** — no NEW broken cross-ref introduced; the new §2.1 target anchor exists on disk (grep).
- **AC-LD-007** — scope discipline: ONLY `CLAUDE.local.md` (+ SPEC artifacts) changed; the SSOT doctrine files (`template-internal-isolation-doctrine.md`, `coding-standards.md`) are byte-unchanged.

Each AC is mechanically verifiable. Severity / blocking classification and REQ→AC traceability are in acceptance.md §D.1 / §D.2.

---

## F. Evidence (verification-claim-integrity)

### F.1 Re-verified ground-truth (command → observed output, this tree, 2026-06-23)

| Claim | Command | Observed |
|-------|---------|----------|
| 806 lines | `wc -l < CLAUDE.local.md` | `806` |
| 33939 bytes | `wc -c < CLAUDE.local.md` | `33939` |
| 25 top-level sections | `grep -cE '^## [0-9]+\.' CLAUDE.local.md` | `25` (§1..§25) |
| Pre-PR checklist is a 7-item bullet list (L108-118) | `sed -n '108,118p' CLAUDE.local.md` (read via Read tool) | 7 `- [ ]` bullets + the §25.3 pointer at L118 (`See .moai/docs/template-internal-isolation-doctrine.md §25.3 ...`) → duplicated body confirmed |
| §25.3 5-item self-check is the canonical owner | `grep -n '§25.3 Pre-commit Self-Check (5-item' .moai/docs/template-internal-isolation-doctrine.md` | `44:### §25.3 Pre-commit Self-Check (5-item Mandatory Checklist)` → SSOT EXISTS (REQ-LD-001 target valid) |
| §25.1 content-class catalogue is the canonical C1-C8 owner | `grep -n '§25.1 정의 — Allowed vs Forbidden Content Classes' .moai/docs/template-internal-isolation-doctrine.md` | `9:### §25.1 정의 — Allowed vs Forbidden Content Classes` → SSOT EXISTS (REQ-LD-003 target valid) |
| §2.1 cites a NON-EXISTENT coding-standards.md §MUST | `grep -nE 'C1/C2\|MUST constraints\|## MUST\|### MUST' .claude/rules/moai/development/coding-standards.md` | `(no match)` → coding-standards.md has NO §MUST section and NO C1-C8 → **§2.1 cross-ref is STALE** (REQ-LD-003 confirmed) |
| coding-standards.md actual headers (proof of stale ref) | `grep -nE '^#{1,4} ' .claude/rules/moai/development/coding-standards.md` | Language Policy / File Size Limits / Content Restrictions / Footer Convention / Duplicate Prevention / Thin Command Pattern / CC Version Compatibility / Paths Frontmatter / Bash Risk-Amplifier — NO `MUST` heading |
| §19.1 HUMAN GATE body present (KEEP target) | `grep -n '§19.1 구현 착수 승인\|plan-to-implement HUMAN GATE' CLAUDE.local.md` | `702:### §19.1 구현 착수 승인 ...`, `704:[HARD] **구현 착수 승인 (plan-to-implement HUMAN GATE)...` → body present (REQ-LD-005 target) |
| CLAUDE.local.md is git-TRACKED | `git ls-files CLAUDE.local.md; echo exit=$?` ; `git check-ignore -v CLAUDE.local.md; echo exit=$?` ; `git log --oneline -3 -- CLAUDE.local.md` | `CLAUDE.local.md` + `exit=0` (tracked in index); `exit=1` (NOT ignored, no `.gitignore` entry); committed history `8e78530bb` / `de13ecc4c` / `96fad88ff`. File header L1-5 says only "Audience: GOOS (local developer only)" — makes NO "not checked in" claim. The "not checked in" phrase is the system-reminder's file-DESCRIPTION, not the file's content; it was mis-sourced and is corrected here (§A.5 / C-1). |

### F.2 Gaps (explicitly NOT observed at plan-phase)

- The exact final line-count is NOT fixed at plan-phase — it is a RANGE derived from the two qualifying edits (REQ-LD-007); the precise number is a run-phase outcome.
- The PRECISE condensed wording of the §2.1 corrected cross-ref and the Pre-PR pointer line is NOT authored at plan-phase — plan.md §C identifies WHICH prose is duplicated/stale; the run-phase authors the corrected one-line pointer.
- Whether any OTHER §2 passage is duplicated-elsewhere was assessed and found NEGATIVE (only the two candidates qualify under the CONSERVATIVE bound); the plan-phase default for every non-candidate passage is KEEP.
- **(D1-equivalent)** The duplication/staleness re-audit used `grep` of the distinctive content. This is a plan-phase grep snapshot — the run-phase MUST re-run the duplication/existence grep (AC-LD-002 / AC-LD-006) before each edit, since the SSOT files could change between plan and run.

### F.3 Residual risk

- Over-aggressive interpretation of "CONSERVATIVE" could tempt the run-phase to also condense a dev-local section that LOOKS verbose but is dev-local-unique. Mitigated by the §A.6 preserve map + REQ-LD-006 (only the two §2 candidates are content changes) + AC-LD-004 (dev-local section survival guard).
- The §19.1 body could be accidentally caught by a "§19 is a pointer-index section, so condense it" heuristic. Mitigated by REQ-LD-005 + C-3 + AC-LD-001 (the §19.1 `[HARD]` body survival is a blocking AC).
- The §2.1 correction could introduce a new broken cross-ref if it points at a wrong anchor. Mitigated by REQ-LD-004 + AC-LD-006 (verify the new target exists on disk before committing).
- `CLAUDE.local.md` is git-tracked, so the diet edit lands in the run/sync commit diff (standard P2/P5 close mechanics, §A.5). Low residual risk: the close is the ordinary tracked-file path — the SPEC artifacts + the `CLAUDE.local.md` edit are in the same commit, and `sync_commit_sha` points at a commit whose diff contains `CLAUDE.local.md`. AC verification reads the live on-disk file (== committed state after the edit lands).
- A `make build` re-embed is NOT triggered by this SPEC (no template edit, C-5) — the output-styles/template parity guards are irrelevant here. Risk only if the run-phase accidentally edits a template file; guarded by C-7 (untouched paths PRESERVE).

---

## G. SPEC ID Pre-Write Self-Check (recorded per protocol)

decomposition: SPEC ✓ | STEERING ✓ | ALIGN ✓ | LOCAL ✓ | DIET ✓ | 001 ✓ → PASS

Canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`: first segment literal `SPEC`; middle segments STEERING / ALIGN / LOCAL / DIET each match `[A-Z][A-Z0-9]*` (first char uppercase letter, rest uppercase alphanumerics, length ≥ 1, valid); last segment `001` matches `\d{3}` (digit-only, no alpha suffix). All segments PASS.

---

## H. Cross-References

- Anthropic best-practices "Write an effective always-loaded instruction surface" (per-line test, bloat warning) — the doctrine applied.
- Anthropic blog "Steering Claude Code: skills, hooks, rules, subagents and more" (always-loaded vs conditional steering surfaces; the project-local instruction file is always-loaded).
- `CLAUDE.local.md` — the diet target (§2 Pre-PR checklist L108-118; §2.1 Template Content Neutrality L106; §19.1 HUMAN GATE body L702-718 = KEEP).
- `.moai/docs/template-internal-isolation-doctrine.md` §25.3 (5-item Pre-commit Self-Check — the M-POINTER target for the Pre-PR checklist) + §25.1 (Allowed/Forbidden content-class catalogue — the corrected target for the §2.1 cross-ref).
- `.claude/rules/moai/development/coding-standards.md` — the file the §2.1 stale cross-ref WRONGLY cited (it has NO §MUST section and NO C1-C8); the correction removes this broken citation.
- CLAUDE.local.md §15 (language-neutrality) + §25 (internal-content isolation) — these bind `internal/template/templates/**`, NOT `CLAUDE.local.md` itself (C-6); the file's legitimate internal refs are NOT neutralized.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — `(none) → draft` owned by manager-spec.
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` — era classification (era: V3R6 set explicitly; H-override).
- SPEC-STEERING-ALIGN-OUTPUT-STYLE-SLIM-001 (sibling P5, COMPLETED, origin ed8172482) — the over-cut-avoidance lesson (−26L actual vs −150-250L estimate) this SPEC inherits via the CONSERVATIVE bound + REQ-LD-007 behavioral-PASS escape.
- SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001 (sibling P2, COMPLETED, origin d0ca1f214) — the M-DELETE + M-POINTER diet technique + the AC-009 over-cut-gate pattern this SPEC mirrors via REQ-LD-002.
