# Implementation Plan — SPEC-V3R6-RULES-VERSION-FORMAT-001

## §A. Context

Plan-phase implementation plan for the version-staleness + format/consistency lane of the Sprint 16 rules-improvement cohort. Each milestone groups one in-scope target (A-G). Every deployed-file edit is paired with its template-mirror edit and a grep-verifiable AC. Run phase is DDD/TDD-agnostic — these are documentation edits validated by grep + the Go-test trio, not new Go code.

## §B. Known Issues / Risk Register

| ID | Risk | Severity | Mitigation |
|----|------|----------|------------|
| R1 | **Zone-registry clause-sync entanglement** — editing a `clause:` field without the identical source edit breaks the clause-substring invariant. | HIGH | M1 edits source line AND registry clause (L291/L299) as an atomic pair (§F.1), then verifies via `grep -Fq` DIRECTLY (re-scoped per D1 — `moai constitution validate` is baseline-RED, see R9). AC-VFM-001a/d assert the substring holds. |
| R2 | **team-pattern-cookbook.md same-file conflict with SSOT-DEDUP-001** — both this SPEC (time-fix) and SSOT-DEDUP (structural edit) touch this file. | MED | This SPEC OWNS the time-estimation fix (option (i)). Sequencing dependency recorded in §F.6 — coordinate so one merges before the other rebases; do NOT hand the time-fix to SSOT-DEDUP. |
| R3 | **csharp/kotlin multi-line version refs** — version identity appears in title + features list + multiple body refs (not a single line). | MED | M2 treats csharp/kotlin as multi-location edits; AC greps assert the NEW version string appears AND the OLD string is absent (deployed + template). |
| R4 | **ruby version-identity ambiguity** — file currently reads `Ruby 3.3+`, not the `Ruby 3.x` the brief assumed. | LOW | M2 bumps the version-identity pin to `Ruby 4.0` (Dec 2025 release); the floor-style `3.3+` phrasing is replaced by the current `4.0` identity. Recorded as a judgment call in §F.2. |
| R5 | **Korean-block over-translation in session-handoff** — the file has 92 Korean lines, but ~half are canonical paste-ready/localization EXAMPLES that must stay Korean. | HIGH | M4 translates ONLY the instruction-prose blocks (§ Diet Constraints, § V0 Abort Gate); the localization-table renderings + paste-ready skeleton examples (`전제 검증:`/`실행:`/`머지 후:`) stay verbatim. AC asserts examples preserved. |
| R6 | **manager-develop L234-240 double-hit** — Korean prose AND wall-time targets converge in one block. | LOW | This block is fixed once under M4 (language) with the time-target removal folded in; M6 AC cross-checks no wall-time survives. |
| R7 | **Mirror-parity AC vacuity** — `TestRuleTemplateMirrorDrift` enrolls ONLY `workflowOptMirroredPaths` (`hooks-system.md` + `session-handoff.md` among my edits + 3 others). The other 14 files I edit are byte-identical-by-discipline but NOT enrolled, so a bare "TestRuleTemplateMirrorDrift PASS" AC is VACUOUSLY green for them (a forgotten mirror edit passes both Go tests silently). | HIGH | Per-file mechanism assigned from VERIFIED plan-time `diff -q` (acceptance.md §D.4): enrolled→test+`diff -q`; 12 non-enrolled byte-identical→explicit `diff -q`; `zone-registry.md` (only true divergence)→leak-test + scoped Opus-clause-line `diff`. Run-phase re-runs `diff -q` pre-commit (TOCTOU guard). Cohort finding confirmed against test source, not assumed. |
| R8 | **zone-registry whole-file `diff -q` false-fail** — zone-registry deployed carries a deployed-only `CONST-V3R6-001` block (L1006-1018) §25-stripped from the neutral mirror, so a whole-file `diff -q` FALSE-FAILS even when my Opus-clause edits mirror correctly. | MED | My 2 edit lines (L291/L299) are in the byte-identical region (verified: the divergence grep returned empty for those lines). Mechanism = leak-test + SCOPED `diff` of only the 2 edited Opus-clause lines, NEVER whole-file `diff -q` (acceptance.md §D.4 + AC-VFM-008c). |
| R9 | **Keystone validator baseline-RED (D1)** — `moai constitution validate` aborts at registry load TODAY (`CONST-V3R6-001` rejected by `ruleIDPattern = ^CONST-V3R[25]-\d{3,}$`, `rule.go:10`), so any AC depending on "validate → 0 SentinelDrift" is unreachable. INDEPENDENT of this SPEC. | HIGH | AC-VFM-001a re-scoped to assert the clause-substring invariant DIRECTLY via `grep -Fq` (verifiable today). Full validator green-gate (AC-VFM-001b) is SHOULD + BLOCKED on sibling `SPEC-V3R6-RULES-CONST-RULEID-001` (1-line `ruleIDPattern` Go fix; OUT OF SCOPE here — doc-only SPEC). Sequencing: CONST-RULEID-001 → VERSION-FORMAT-001. Documented in acceptance.md §D.6 + §C corrected baseline. |
| R10 | **D2 L845 non-substring** — CONST-V3R5-022 (L845) clause is ALREADY a non-substring of its source (`grep -cF '1M context (Opus 4.7)' context-window-management.md` == 0; source is a table). | MED | L845 EXCLUDED from edit + verbatim-substring AC. AC-VFM-001e carve-out: clause LEFT UNEDITED, count stays 0, no table-text rewrite (out of scope). §F.1 zone-registry list updated to drop L845. |
| R11 | **D3 ruby/Rails inconsistency** — bumping Ruby→4.0 while leaving Rails 7.2 yields an internally inconsistent file. | LOW | M2 bumps Rails 7.2 → 8.0 alongside Ruby 4.0 (paired); AC-VFM-002c delta-greps BOTH. |

## §C. Pre-flight checklist (run-phase entry)

- [ ] `git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD` clean (no parallel-session race).
- [ ] Confirm `internal/template/templates/.claude/rules/moai/**` mirrors exist for all in-scope files (verified at plan: all present).
- [ ] Baseline: `go test ./internal/template/... -run 'TestRuleTemplateMirrorDrift|TestTemplateNoInternalContentLeak|TestTemplateNeutralityAudit'` GREEN before edits.
- [ ] Baseline observed (CORRECTED — D1): `moai constitution validate` is **RED today** — `load registry: entry 114 validation error: rule ID "CONST-V3R6-001" does not match pattern "^CONST-V3R[25]-\d{3,}$"`. The registry load aborts before the drift loop (`rule.go:10` excludes V3R6). This is a baseline-RED keystone gate INDEPENDENT of this SPEC. The full validator green-gate is BLOCKED on `SPEC-V3R6-RULES-CONST-RULEID-001` (1-line `ruleIDPattern` Go fix) landing first. This SPEC verifies the clause-substring invariant DIRECTLY via `grep -Fq` instead (acceptance.md AC-VFM-001a + §D.6). The earlier "baseline clean" claim was FALSE and is corrected here per verification-claim-integrity §2 (attribution: command run, output observed).

## §D. Constraints (inherited from spec.md §C)

Template-First (Go-test trio, no `embedded.go`, no naive diff); §25 neutrality; zone-registry clause-substring invariant verified via `grep -Fq` (NOT `moai constitution validate`, which is baseline-RED per §C/§D.6 until CONST-RULEID-001 lands); each fix grep-verifiable in deployed + template; stay in lane.

## §E. Self-Verification

Plan-phase audit-ready signal lives in `progress.md` §E.1. Run-phase populates §E.2/§E.3.

## §F. Milestones (A-G grouped)

> Milestones are priority-ordered, not time-ordered (per agent-common-protocol § Time Estimation). Each milestone: deployed edit + template mirror edit + grep AC(s). M1 first (highest risk, validator coupling); M7 last (precision polish).

### §F.1 — M1 (target A): Opus model identity correction

**Files (deployed + mirror each):**
- `core/moai-constitution.md` L56-57 — Principle 4/5 lines: generalize "Opus 4.7 does not auto-spawn" / "Opus 4.7 prefers reasoning" to "Opus 4.7+ / 4.8" framing. Header L47 already correct ("Opus 4.7+ / 4.8") — leave.
- `core/zone-registry.md` L291 (CONST-V3R2-028), L299 (CONST-V3R2-029) — clauses MUST be edited to the IDENTICAL new substring as the constitution source lines (atomic pair). Verify each via `grep -Fq "<clause text>" <source file>` — NOT `moai constitution validate` (baseline-RED, §C/§D.6).
- `core/zone-registry.md` L845 (CONST-V3R5-022) — **LEFT UNEDITED (D2 carve-out, AC-VFM-001e).** Verified: this clause is ALREADY a non-substring of its source — `grep -cF '1M context (Opus 4.7)' context-window-management.md` == 0, because the source uses a TABLE (`| Opus 4.8 (1M) | ... | 50% |`), not the clause's prose phrasing. Editing the clause cannot produce a clean substring without a table-text rewrite (out of scope). L845 stays as-is; AC-VFM-001e asserts the count stays 0 and the line is untouched. Do NOT touch `context-window-management.md` for this.
- `development/agent-authoring.md` L61, L83 — "require Opus 4.7+" effort caveats: keep the genuine 4.6 floor caveat ("On Opus 4.6, highest is `high`") verbatim; state 4.8 as current where the line reads as current substrate. L216-217/L220 already say "Opus 4.8" — leave.
- `development/skill-authoring.md` L25 — "(xhigh/max require Opus 4.7+)" — acceptable as-is (it is a back-compat floor); annotate only if it reads as current-substrate identity. Treat as SHOULD.
- `workflow/worktree-integration.md` L351, L355 — CC-version compatibility table referencing "Opus 4.7 support". These are genuine version-availability facts (effortLevel landed in 2.1.110 for Opus 4.7); they are back-compat history, NOT stale current-substrate claims. Treat as SHOULD / annotate; do not rewrite version history.

**ACs:** AC-VFM-001a (clause-substring via `grep -Fq`, validator-independent), AC-VFM-001b (full validator green — SHOULD, BLOCKED on CONST-RULEID-001), AC-VFM-001c (constitution Principle 4/5 reads 4.7+/4.8, delta-grep), AC-VFM-001d (zone-registry clauses are verbatim substrings — L291/L299 only), AC-VFM-001e (L845 carve-out — count stays 0, untouched).

**Zone-registry-clause-sync (R1, re-scoped per D1):** The run-phase agent MUST edit each editable `clause:` (L291/L299) and its coupled constitution source line as a single atomic change, then verify the substring DIRECTLY via `grep -Fq` (NOT `moai constitution validate`, which is dead at baseline). A clause edited without its source = a broken substring invariant (the same hazard `SentinelDrift` would catch if the validator ran).

### §F.2 — M2 (target B): Language toolchain version corrections

**HARD (genuinely stale):**
- `languages/csharp.md` — `.NET 8 / C# 12` → `.NET 10 (LTS) / C# 14` at title (L11), features (L18-20), body refs (L27, L41-42, L53). Multi-location (R3).
- `languages/kotlin.md` — `Kotlin 2.0 / Ktor 3.0` → `Kotlin 2.2+ (current 2.4) / Ktor 3.x (current 3.5)` at title (L11), features (L16-17), section headers (L26, L32), Context7 refs (L76, L78), body (L129). Multi-location (R3).
- `languages/ruby.md` — TWO paired bumps (D3 — consistency): (a) `Ruby 3.3+` version-identity → `Ruby 4.0` at title (L11) + features (L16, L42, L46); (b) `Rails 7.2` → `Rails 8.0` at title (L11) + features (L17) + section header (L60). Both MUST move together — a Ruby-4.0/Rails-7.2 pairing is internally inconsistent (Rails 8.x is current). Judgment call (R4): brief said "any Ruby 3.x pin"; file actually has `3.3+` and the paired Rails 7.2 the brief did not flag. AC-VFM-002c delta-greps BOTH (Ruby 4.0 + Rails 8.0 present, Rails 7.2 → 0).
- `languages/go.md` L101 — Docker tag `golang:1.23-alpine` → `1.26-alpine`. Keep `Version: Go 1.23+` (L7) floor as acceptable.

**SHOULD (acceptable floors, optionally annotate):**
- `languages/java.md` — Java 21 LTS floor accurate; MAY annotate "current 25".
- `languages/cpp.md` — C++23/C++20 accurate; MAY annotate "C++26 ratified Mar 2026".
- `languages/typescript.md` — TS 5.9 floor; MAY annotate "current 6.0". React 19 / Next.js 16 ACCURATE — leave.

**LEAVE AS-IS (verified correct):** python (3.13+/Django 5.2 LTS), swift (6+), php (8.4), elixir (1.18), rust (1.92).

**ACs:** one grep AC per HARD fix asserting NEW string present + OLD string absent, in deployed + template (8 greps: csharp/kotlin/ruby/go × {deployed,template}).

### §F.3 — M3 (target C): Footer consistency policy

**DECISION: option (b) — consistent-by-absence + targeted application.**

Rationale: 36 of 61 rule files lack a footer. Many are short path-scoped rules whose footer would be pure ceremony. Adding a uniform footer to all 36 would dominate the SPEC's diff for near-zero value and risk neutrality-test churn. Instead:
- Document the policy: short path-scoped rules MAY omit a `Version/Status` footer (consistent-by-absence); SSOT-owning canonical-reference rules SHOULD carry a footer.
- Apply the footer only to the highest-value SSOT-owning files that currently lack one (small, bounded set — identified at run-phase by intersecting "missing footer" with "Classification: Canonical Reference" or SSOT-declaring files).

The policy STATEMENT is the deliverable, not bulk footer-adding. Where the policy statement itself lands (a short note in `coding-standards.md` § content/format conventions OR a new `## Footer Convention` micro-section) is a run-phase decision; pick the lowest-churn location.

**ACs:** AC-VFM-003 — the footer policy statement exists (grep for the policy sentence) in deployed + template; the highest-value SSOT files named in run-phase carry a footer.

### §F.4 — M4 (target D): Korean → English instruction prose

> Own milestone (large translation effort, per brief). SSOT-DEDUP-001 does NOT touch session-handoff — no conflict.

**Files:**
- `core/settings-management.md` L40-44 — translate `alwaysLoad` Korean notes to English (small, 4 lines).
- `workflow/session-handoff.md` § Diet Constraints (~L243-300) + § V0 Abort Gate Doctrine (~L243-340) — translate instruction-prose Korean to English. **PRESERVE (R5):** the localization-table renderings + paste-ready skeleton examples (`✂──── 여기부터 복사 ────✂`, `전제 검증:`, `실행:`, `머지 후:`, the en/ko/ja/zh table) — these are intentional locale-OUTPUT examples, NOT instruction prose. Anti-patterns AP-D-001..005 / AP-V-001..004 prose translates; the example renderings stay.
- `development/manager-develop-prompt-template.md` L228-245 (§ 검증 block, Korean) — translate to English (folds into M6 time-removal, R6).

**ACs:** AC-VFM-004a (settings-management: no Korean instruction prose; alwaysLoad note reads English — deployed + template), AC-VFM-004b (session-handoff Diet/V0 blocks English; localization examples `전제 검증:`/`실행:` PRESERVED — deployed + template), AC-VFM-004c (manager-develop §검증 block English — deployed + template).

### §F.5 — M5 (target E): Emoji removal

**Files:**
- `development/skill-writing-craft.md` L217 — `✅` → "Required" / "Yes".
- `development/skill-ab-testing.md` L210-217 — `✅` → "PASS" / "Met".

**ACs:** AC-VFM-005 — `grep -c '✅\|❌' <file>` == 0 in deployed + template for both files.

### §F.6 — M6 (target F): Time-estimation removal

**Files:**
- `development/orchestrator-templates.md` L199 ("Parallel development for 2 days") → phase ordering ("Parallel development across two phases" / priority label). L56 ("waits 10s") is a literal hook-timeout value, NOT a planning estimate — leave (verify framing).
- `development/manager-develop-prompt-template.md` L234-240 wall-time targets ("목표: ≤30분, W3에서 91분", "≤1") → phase/priority ordering (folds with M4 translation). L174 `0.5s` is example test-output, not a target — leave.
- `workflow/team-pattern-cookbook.md` L93-106 ("Day 1/2/3") → phase ordering ("Phase 1/2/3").

**SSOT-DEDUP dependency note (R2):** `team-pattern-cookbook.md` is ALSO edited structurally by SPEC-V3R6-RULES-SSOT-DEDUP-001. This SPEC OWNS the time-estimation fix (brief option (i)). Coordination: whichever of the two SPECs merges first, the second rebases onto it. Recommended sequencing — land this SPEC's `team-pattern-cookbook.md` time-fix as an isolated edit so SSOT-DEDUP's structural edit rebases cleanly (the two edits are in different sections: time-fix in the Day-labeled example block, SSOT-DEDUP in the consolidation targets). The orchestrator MUST sequence these two SPECs (not run them as a blind parallel pair) to avoid a same-file merge conflict.

**ACs:** AC-VFM-006a (orchestrator-templates: no "N days/hours" planning estimate — deployed + template), AC-VFM-006b (manager-develop: no wall-time target survives — deployed + template), AC-VFM-006c (team-pattern-cookbook: no "Day N" label; phase ordering present — deployed + template).

### §F.7 — M7 (target G): Precision framing fixes

**Files:**
- `workflow/session-handoff.md` Block 1 spec (~L77 region, "`ultrathink.` triggers Adaptive Thinking xhigh effort") → reframe: `ultrathink` sets `effort: xhigh`; Adaptive Thinking is a DISTINCT axis (the thinking mode, explicitly enabled via `thinking:{type:"adaptive"}`) — not "ultrathink triggers Adaptive Thinking" as if one toggles the other. (Consistent with CLAUDE.md §12 verified framing; CLAUDE.md itself is out-of-tree, informational only.)
- `core/hooks-system.md` L62 / L11-12 / L66 / L95 — disambiguate "Setup REMOVED": **VERIFIED via official CC docs** that upstream `Setup` IS a current hook event (triggered by `--init-only`/`--init`/`--maintenance`). Reframe: the moai-adk Go `EventSetup` constant + `moai hook setup` subcommand were retired (INTERNAL); the upstream CC `Setup` event still exists. Do NOT imply the upstream event is gone.

**ACs:** AC-VFM-007a (session-handoff: ultrathink/effort/Adaptive-Thinking framing distinguishes the two axes — deployed + template), AC-VFM-007b (hooks-system: Setup framing names the internal-constant retirement AND states upstream Setup event is current — deployed + template).

## §G. Anti-Patterns to avoid (run-phase)

- AP-1: Editing a zone-registry `clause:` without the coupled source edit → `SentinelDrift`. Always atomic-pair.
- AP-2: Translating the Korean localization-table / paste-ready EXAMPLES in session-handoff (they are intentional locale output, not instruction prose).
- AP-3: Bulk-adding footers to all 36 files (option (a)) — scope bloat; policy is option (b).
- AP-4: Restructuring language files onto a MUST/MUST NOT skeleton — DEFERRED to SPEC-V3R6-RULES-LANG-SKELETON-001.
- AP-5: Handing `team-pattern-cookbook.md` time-fix to SSOT-DEDUP — this SPEC owns it (option (i)).
- AP-6: Asserting mirror parity via naive `diff` instead of the Go-test trio.
- AP-7: Touching archived-agent refs or SSOT de-dup content (other lanes).

## §H. Cross-References

- spec.md §A.2 — zone-registry clause-as-substring coupling table.
- `internal/constitution/validator.go:261-271` — `SentinelDrift` substring check.
- `internal/template/rule_template_mirror_test.go` + `internal_content_leak_test.go` + `template_neutrality_audit_test.go` — Go-test trio.
- `coding-standards.md` § Language Policy / § Content Restrictions — instruction-language + emoji + time-estimate prohibitions.
- `agent-common-protocol.md` § Time Estimation — phase/priority replacement rule.
- Sibling SPECs: HOTFIX-001 (tokens), CATALOG-SCRUB-001 (archived agents), SSOT-DEDUP-001 (de-dup; `team-pattern-cookbook.md` sequencing dependency).
- Forward-link: `SPEC-V3R6-RULES-LANG-SKELETON-001` (deferred language-skeleton unification).
