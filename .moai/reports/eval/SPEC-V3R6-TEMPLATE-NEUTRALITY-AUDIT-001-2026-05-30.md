# Evaluation Report

SPEC: SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001
Phase: 5 (sync-audit, independent skeptical re-verification)
Evaluator: evaluator-active (fresh-judgment auditor)
Date: 2026-05-30
HEAD audited: c7448879f (status implemented, v0.2.0, pushed origin/main)
Overall Verdict: **PASS-WITH-DEBT**

## Headline

The 5 C-class sanitizations (C1/C2/C4/C5/C6/C8) and the new audit test + CI guard are genuinely delivered and pass independent re-verification. BUT two acceptance criteria — **AC-TNA-012 and the AC-TNA-013 checklist sub-check** — FAIL their own canonical grep commands: the required heading strings (`Acceptable Content Range` / `template-acceptable-content` / `contributor-checklist` / `Pre-PR Verification`) do not exist anywhere in `CLAUDE.local.md` or its delegated `coding-standards.md`. The prior 5-verifier workflow reportedly marked these PASS; they do not. The core sanitization work is sound; the documentation-guideline ACs are unmet by their SSOT commands.

## Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 80/100 | PASS (borderline) | 8 of 10 active grep/test-verifiable ACs PASS with EXACT canonical commands (C1=0, C2 6≤6, C4 7≤7, C5=0, C6=0, audit test ok, matrix=8, GOOS=3). **AC-TNA-012 FAILS** (exit 1) and **AC-TNA-013 checklist grep FAILS** — required headings absent repo-wide. The substantive sanitization (the SPEC's primary goal) is fully met; the two failing ACs are documentation-guideline discoverability, not content sanitization. |
| Security (25%) | 95/100 | PASS | No security-relevant doctrine citation broken; C8 `GOOS=` 3 files preserved (no build-critical env var removed); C6 PR-ref replacement meaning-preserving; sanitization removed only dev-trace metadata. |
| Craft (20%) | 86/100 | PASS | Go test well-formed; two-pass C2 exclusion correct and matches perl PCRE file set 1:1; disjoint from leak test; cross-platform (`filepath.WalkDir`/`relUnderRoot` ToSlash/`os.ReadFile`); C8 PRESERVE asserted as exactly 3. `go vet` clean. Gaps D-4..D-6. |
| Consistency (15%) | 88/100 | PASS | C3/C7 deferral coherent (disjoint sets, opt-in strict confirmed); CHANGELOG single accurate entry; run trailers manager-develop ×4 + orchestrator-direct ×1 (catalog-hash); sync trailer manager-docs with in-progress→implemented. acceptance.md frontmatter lag (D-2). |

Aggregation (harmonic-mean style; must-pass = Functionality + Security):
- Functionality is borderline (two doc-ACs fail) but the primary content-sanitization goal is met and 8/10 commands pass → above pass-threshold, no hard must-pass cap triggered.
- Harmonic mean of {80, 95, 86, 88} ≈ **86.7/100**.

## Methodology Note (fresh judgment — self-correction recorded)

My first verification pass used non-canonical grep patterns and produced misleading numbers (C2=146, C4=15). I discarded those and re-ran the EXACT verification commands from acceptance.md (the SSOT). This corrected pass surfaced what the prior 5-verifier workflow missed: AC-TNA-012 and the AC-TNA-013 checklist grep FAIL their literal commands. A fresh-judgment audit must run the SSOT command verbatim, not paraphrase it — that distinction is the entire value of this re-verification.

## Independent Re-Verification (claims re-run verbatim, not trusted)

### Functionality — active ACs with EXACT canonical commands
- **AC-TNA-001 (C1 `/Users/`)**: `grep -rln '/Users/' ... | wc -l` → **0**. PASS.
- **AC-TNA-002 (C2 bare-narrative)**: `grep -rlP '(?<![A-Za-z0-9-])V3R[0-9]'` → **6**; awk allow-list (`### C2`..`### C3`) → **6**; `6 ≤ 6` PASS at exact boundary. The 6 PRESERVE files are genuine decision-record/doctrine citations; Go test allow-list matches the perl file set 1:1.
- **AC-TNA-004 (C4 feedback_/memory)**: → **7**; awk allow-list → **7**; `7 ≤ 7` PASS.
- **AC-TNA-005 (C5 CLAUDE.local.md)**: → **0**. PASS.
- **AC-TNA-006 (C6 PR #N)**: → **0**. PASS.
- **AC-TNA-008 (audit test isolated)**: `go test ... -run TestTemplateNeutralityAudit` → `ok`, exit 0. PASS.
- **AC-TNA-009 (CI workflow)**: file exists + `python3 yaml.safe_load` → YAML OK. PASS (pre-merge form).
- **AC-TNA-010 (matrix 8 sections)**: `grep -cE '^### C[1-8] '` → **8**. PASS.
- **AC-TNA-011 (GOOS= preserve)**: → **3**. PASS.
- **AC-TNA-012 (guideline)**: `grep -q 'Acceptable Content Range\|template-acceptable-content' CLAUDE.local.md` → **exit 1 = FAIL**. See D-1.
- **AC-TNA-013 (checklist sub-check)**: `grep -q 'Acceptable Content Range\|template-acceptable-content\|contributor-checklist\|Pre-PR Verification' CLAUDE.local.md` → **exit 1 = FAIL**. See D-1. (The `moai init` clean-run portion of AC-013 was not executed in this read-only audit; the checklist-discoverability grep portion FAILS.)

Repo-wide confirmation: none of `Acceptable Content Range`, `template-acceptable-content`, `contributor-checklist`, `Pre-PR Verification` appear in `CLAUDE.local.md` OR in `.claude/rules/moai/development/coding-standards.md` (the file CLAUDE.local.md §2.1 delegates to). The guideline content exists under the heading `**§2.1 Template Content Neutrality**` (CLAUDE.local.md line 106) — substantively correct, but matching NONE of the AC's grep alternatives.

### Package-RED baseline — central judgment call, rigorously proven
- HEAD c7448879f: **13 unique failing test functions**.
- Detached worktree at pre-SPEC parent **a9757f484** (parent of first SPEC commit 1046c6a3c): **13 unique failing test functions**.
- `comm -3` of the two unique failing-test-name sets → **IDENTICAL**. Zero regressions, zero new failures attributable to this SPEC.
- Neutrality test file confirmed ABSENT at baseline → net-new deliverable passing in isolation + targeted CI.
- **Conclusion: the 13 package-RED failures are genuinely pre-existing. The deferral is sound scope discipline, NOT defect-masking.**

### Disjointness (no double-enforcement)
- Leak test (ISOLATION-001) owns spec-id / REQ-AC / audit-citation / date (C3) / commit-sha (C7); strict classes opt-in via `MOAI_TEMPLATE_LEAK_STRICT` (5 source references confirmed).
- Neutrality test owns C1, C2 (bare-narrative two-pass), C4, C5, C6, C8-preserve.
- Both run together → `ok` (no double-FAIL). C2 exclusion regex `(SPEC|CONST|REQ|AC|BC|HARNESS|PROJECT|HL)-V3R[0-9]` + preceding-rune `[A-Za-z0-9-]` cleanly cedes ID-embedded matches to the leak test. No class enforced by both files.

### Security
- C8 `GOOS=(linux|windows|darwin|freebsd|openbsd|netbsd)` preserved across 3 files; no Go build env var removed.
- C6 PR-ref sanitization replaced specific PR numbers with meaning-preserving generic prose; no broken sentences, no removed OWASP/auth/security doctrine.
- Ownership trailers: 4 run-phase commits `manager-develop`; catalog-hash fix `eeab1d0e6` = `orchestrator-direct` (acceptable — mechanical hash regen, not a status transition); sync commit `manager-docs` with `in-progress → implemented` — matches canonical Status Transition Ownership Matrix.

## Findings (coverage-first — includes low-severity/uncertain)

- **[MEDIUM, confidence HIGH] D-1 — AC-TNA-012 AND AC-TNA-013 (checklist sub-check) FAIL their canonical commands.** `grep -q 'Acceptable Content Range\|template-acceptable-content' CLAUDE.local.md` → exit 1; the broader AC-013 grep adding `contributor-checklist|Pre-PR Verification` → also exit 1. None of these strings exist in CLAUDE.local.md or in the delegated `coding-standards.md`. The guideline CONTENT exists under `**§2.1 Template Content Neutrality**` and substantively satisfies REQ-TNA-012/013 intent (acceptable kept-classes + FORBIDDEN classes + CI-guard reference), but two acceptance criteria are unmet by their own verification commands. This is a genuine AC miss — the prior workflow's PASS for AC-012 is incorrect. Remediation (preferred): add a heading containing the literal string `Acceptable Content Range for Templates` (and ideally a `Pre-PR Verification` checklist) to CLAUDE.local.md §2.1. Alternative: amend AC-012/013 grep alternations to include `Template Content Neutrality`. This is the primary debt driving PASS-WITH-DEBT.

- **[INFO, confidence HIGH] D-2 — acceptance.md frontmatter status lag.** `acceptance.md` reads `status: in-progress` / `version: "0.1.2"` while `spec.md` reads `status: implemented` / `version: 0.2.0`. Sibling artifacts out of sync. Not itself an AC failure (acceptance.md body is criteria SSOT; ownership matrix mandates only spec.md transition), but a one-line fix at Mx-close.

- **[INFO, confidence HIGH] D-3 — C3/C7 deferral target test is itself RED.** `TestTemplateNoInternalContentLeak` (owns C3/C7 per the deferral) is one of the 13 pre-existing failures. The deferred classes therefore currently point at a failing gate. Acknowledged in spec.md §3.4 and routed to the successor cleanup SPEC; strict classes are opt-in, so default-mode leak failure is from a non-strict class — but a reader should know the deferral target is RED right now.

- **[LOW, confidence HIGH] D-4 — C2/C4 WARN classes do not enforce the count ceiling in the Go/CI gate.** The `C2`/`C4` subtests treat over-allow-list hits as `t.Logf` WARN only (never `t.Errorf`). The `actual ≤ allowlist` bound is enforced ONLY by the shell AC commands (AC-002/004), NOT by the CI Go gate. Consistent with the documented "WARN-level advisory" design, but the CI guard does not regression-protect C2/C4 counts the way the AC pass implies. Binary classes C1/C5/C6 ARE enforced (t.Errorf).

- **[LOW, confidence HIGH] D-5 — C2 boundary has zero headroom.** AC-002 sits at exactly 6/6. Any future bare-narrative `V3R<n>` in a non-allowlisted template file breaks the shell AC immediately (though not the Go test, per D-4). The 6 residuals are documented PRESERVE, but no maintainer slack.

- **[LOW, confidence MEDIUM] D-6 — audit walks on-disk source tree, not `go:embed` snapshot.** `neutralityTemplatesRoot = "templates"` audits source files, not `embedded.go`. Correct for this SPEC's purpose; `embedded_test.go` separately guards regeneration parity. Awareness only.

- **[INFO, confidence MEDIUM] D-7 — CI workflow self-trigger gap.** A PR editing only the workflow file may not re-run the guard if the `pull_request` paths filter omits the workflow path. Cosmetic; template/test edits still trigger it.

## Judgment Call Assessment (skeptical)

**Deferring the 13-test package-RED — sound or masking?** Sound. Proven via worktree baseline: byte-identical 13-failing-test sets at the pre-SPEC parent and HEAD. This SPEC's scope is template-tree content sanitization, not the deployment/render/metadata test infrastructure where the 13 failures live. Deferral is documented (spec.md §3.4, acceptance.md §1, CHANGELOG) with a named successor SPEC. The new audit test passes in isolation and is wired into CI. Not masking.

**Is the SPEC genuinely "implemented" by its own acceptance criteria?** Partially. The PRIMARY goal — sanitizing the template tree of dev-local/dev-incident traces — is fully achieved (C1/C2/C4/C5/C6 all pass, C8 preserved, audit test + CI guard delivered, all on independent re-verification). But two acceptance criteria (AC-TNA-012, AC-TNA-013 checklist) FAIL their literal SSOT commands because the required guideline headings were never added under the strings the ACs grep for. By a strict SSOT reading the SPEC is NOT fully green. By an intent reading the guideline exists. PASS-WITH-DEBT captures this precisely: the content work is done and regression-guarded, but the documentation-discoverability ACs are unmet as written.

## Recommendation

**Remediate D-1 before Mx-close — do NOT close as clean PASS.** Add a heading containing `Acceptable Content Range for Templates` plus a `Pre-PR Verification` contributor checklist to `CLAUDE.local.md` §2.1 (or, less preferred, amend the AC-012/013 grep alternations to match the existing `Template Content Neutrality` heading). That single edit makes AC-TNA-012 and the AC-TNA-013 checklist sub-check pass their own commands. While editing, align acceptance.md frontmatter (D-2) and confirm the successor SPEC (`SPEC-V3R6-TEMPLATE-PACKAGE-RED-CLEANUP-001`) is filed since the C3/C7 deferral target is currently RED (D-3). After D-1 is fixed, this SPEC is a clean PASS. As-shipped (HEAD c7448879f) the accurate verdict is **PASS-WITH-DEBT**: the core sanitization passes with zero regressions, but two documentation ACs fail their SSOT verification commands and the surrounding package + deferral target remain RED (pre-existing, out of scope).
