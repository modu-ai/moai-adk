# SPEC Review Report: SPEC-DB-RETIRE-001

Iteration: 1 (independent adversarial plan-audit)
Tier: M (PASS threshold 0.80)
Verdict: **FAIL**
Overall Quality Score (harmonic-mean, pre-firewall): ~0.86
Verdict driver: MP-7 must-pass firewall (unresolved [NEEDS CLARIFICATION]) + 2 substantive scope defects (D1, D2)

> M1 Context Isolation: reasoning context from manager-spec (the "claimed locations" and pushback claims passed in the prompt) was treated as claims-to-verify, not truth. Every claim below was independently re-grepped against the live tree at HEAD.

---

## Must-Pass Results

- [PASS] **MP-1 REQ number consistency** — REQ-DBR-001 … REQ-DBR-021, strictly sequential, no gaps, no duplicates, consistent 3-digit zero-padding (spec.md L44-L82).
- [PASS] **MP-2 GEARS compliance** — all 21 REQs match a valid GEARS pattern: Unwanted (`shall not`, REQ-001/002/006/007/009/010/011/016), Ubiquitous (`shall`, REQ-003/004/005/008/013/014/015/018/019/021), Event-driven (`When … shall`, REQ-012/020), State-driven (`While … shall not … shall`, REQ-017). No IF/THEN, no informal "should/must-try". (spec.md §B)
- [PASS] **MP-3 YAML frontmatter validity** — spec.md L2-L15 carries all 12 canonical fields (id/title/version/status/created/updated/author/priority/phase/module/lifecycle/tags) with correct types + valid optionals `tier: M`, `related_specs`. No rejected snake_case alias. id = `SPEC-DB-RETIRE-001`, status = draft.
- [N/A→PASS] **MP-4 language neutrality** — single-language (Go) removal SPEC; template-directed edits (guidance-file DB-section removal) are language-neutral, and REQ-DBR-015/AC-DBR-015 explicitly enforce §25 (no SPEC-ID leak). No 16-language tooling surface. Auto-passes.
- [PASS] **MP-5 D7 cross-SPEC reconciliation** — verified statuses of all referenced SPECs: SPEC-DB-SYNC-001 = implemented, SPEC-DB-SYNC-RELOC-001 = implemented, SPEC-DEPRECATEDPATHS-RECONCILE-001 = completed, SPEC-UPDATE-REINSTALL-LOOP-001 = completed. None ∈ {retired, superseded, archived} → **no D7 BLOCKING**. §H additionally carries reconciliation prose.
- [N/A→PASS] **MP-6 D8 cross-platform discipline** — `grep syscall .moai/specs/SPEC-DB-RETIRE-001/` = 0 matches → D8 auto-PASS (no cross-platform build-tag concern).
- [**FAIL**] **MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md` returns **1 unresolved marker** at plan.md:160 (`[NEEDS CLARIFICATION: db.yaml DeprecatedPaths 능동 정리 여부]`). Per MP-7 this is a score-independent must-pass failure that forces Verdict = FAIL. The marker is well-formed and correctly deferred to a pre-Kickoff user decision (acceptance.md §D.5 DoD commits to resolving it), and its recommended default ("do not register") is adoptable with zero REQ/AC edits — but the orchestrator MUST resolve it via `AskUserQuestion` before Implementation Kickoff Approval; presence at audit time is the gate failure.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.90 | 0.75-1.0 | REQs content-anchored, unambiguous; single minor looseness (AC-017 grep). |
| Completeness | 0.80 | 0.75 band | All sections present; OOS §C has 5 `### Out of Scope — <topic>` H3s each with `-` bullets. Knocked by D2 (missed target). |
| Testability | 0.78 | 0.50-0.75 | Most ACs are grep/file-exists/build-exit (binary). Knocked by D1 (AC-DBR-022 false-fails) + AC-017 weasel grep. |
| Traceability | 1.0 | 1.0 | 21 REQ → 22 AC; every REQ has ≥1 AC; every AC cites a valid REQ; no orphans (acceptance.md §D). |

---

## Confirmed / Refuted — manager-spec pushback claims

| # | Claim | Verdict | Evidence |
|---|-------|---------|----------|
| 1 | `internal/hook/dbsync/` = 3 files | **CONFIRMED** | `db_schema_sync.go`, `db_schema_sync_test.go`, `db_schema_sync_internal_test.go` (ls). |
| 2 | hook.go: db-schema-sync reg + runDBSchemaSync + loadMigrationPatterns + defaultMigrationPatterns | **CONFIRMED** | hook.go:136-144 (reg), 409-445 (handler), 501-509 (defaultMigrationPatterns), 511-548 (loader). |
| 3 | splitLines/trimSpace dead — sole caller = loadMigrationPatterns | **CONFIRMED** | Within `internal/cli`, both used ONLY at hook.go:525/526/533 (inside loadMigrationPatterns 511-548); defined 550/565. `grep -rn splitLines\|trimSpace internal/cli/` shows only hook.go. Removing loadMigrationPatterns orphans them → golangci-lint `unused`. Correct. NOTE: identically-named `splitLines`/`trimSpace` in `internal/merge/*` + `internal/core/project/initializer_expansion.go` are SEPARATE packages (no collision) — hook.go's copies are genuinely local; deletion scoped to hook.go is safe from over-deletion. |
| 4 | audit_registry.go `yamlAuditExceptions["db"]` real entry | **CONFIRMED** | audit_registry.go:74 (`"db": "consumed via hook line-scan…"`). |
| 5 | dirs.go + v2_detection.go db.yaml comment refs | **CONFIRMED** | dirs.go:44/46/238/241; v2_detection.go:33-34. Comment-only (no DeprecatedPaths slice entry) → count-invariant claim CORRECT. |
| 6 | internal_content_leak_test.go C7 path + settings_test.go comment | **CONFIRMED** | leak_test:980 (`"internal/hook/dbsync/db_schema_sync.go"` as C7 "real restricted-package path"); settings_test:1054 (stale "remain intact" comment). |
| 7 | dbsync real importer = hook.go:27 only | **CONFIRMED** | `grep -rn hook/dbsync`: hook.go:27 (real import), leak_test:980 (string), settings_test:1054 (comment), dbsync's own test:17. No MISSED production importer. |

The audit-parity ordering worry (AP-6) is over-stated but harmless: `TestAuditRegistry_NoUnexpectedYAMLOrphans` scans a synthetic `t.TempDir()` (no db.yaml), and `TestAuditParity_ExceptionsRespected` only asserts 5 expected exceptions {constitution, context, interview, design, harness} — NOT "db". Removing `yamlAuditExceptions["db"]` is build-safe regardless of db.yaml deletion order. Also verified `TestRender_DbSchemaChangeHook_Removed` body references only `handle-db-schema-change.sh` (SPEC-DB-SYNC-RELOC-001; does NOT match `db-schema-sync` pattern) — the test passes after CLI-subcommand removal; only its comment is stale (REQ-DBR-009). Settings-package `db`-section handling (sectionroute/sectionwrite/yamlpatch/web scope) is correctly OOS (§C).

---

## Defects Found

**D1. AC-DBR-022 / §D closure grep FALSE-FAILS on a correct removal — internal/settings/schema_sections_test.go:292 — Severity: major (verification-integrity).**
The closure command `grep -rn "db-schema-sync\|dbsync\|db\.enabled" internal/ cmd/ | grep -v "internal/settings/testdata/"` will return a residual match at `internal/settings/schema_sections_test.go:292` (`"db.enabled", "db.orm", ...`), a settings edit-rejection test key-path that is explicitly PRESERVED (OOS per §C settings machinery). The §D.2 exclusion list only excludes `internal/settings/testdata/`, not this test file; spec.md §D's success grep has the identical hole. Consequence: a complete, correct removal can never make AC-DBR-022 pass, and the pressure is to wrongly delete the settings-test line (settings regression). Verified: this is the ONLY residual match post-removal (all template matches are deleted by REQ-010/013/014).
**Required fix:** add `internal/settings/` (or the specific `schema_sections_test.go`) to the AC-DBR-022 grep filter AND the §D.2 exclusion list AND the spec.md §D success-grep exclusion; OR narrow the `db\.enabled` pattern so it does not match settings key-path test inputs.

**D2. Missed removal target — internal/config/audit_loader_completeness_test.go:16 — Severity: major (incomplete removal; invisible to SPEC's own closure grep).**
`acknowledgedUnloadedSections` carries `"db"` (line 16) with the comment "consumed via hook line-scan (internal/cli/hook.go migration_patterns) … owned by the DB subsystem track". After this SPEC removes hook.go's migration_patterns line-scan and the whole DB subsystem, this allowlist entry + comment become a STALE reference to deleted code. It is NOT captured by any REQ/AC, and does NOT match the closure patterns (`db-schema-sync|dbsync|db.enabled`), so §D verification cannot catch it. It is the exact sibling of the captured `yamlAuditExceptions["db"]` (REQ-DBR-007); leaving one and removing the other is an inconsistent half-removal. Non-build-breaking (the test iterates the template dir, which no longer has db.yaml, so the dead entry is harmless to pass/fail), but it contradicts the SPEC's own "full removal" scope and the §E.3 claim of having found "4 extra references" (this is the uncaptured 5th).
**Required fix:** add a REQ/AC to remove the `"db"` entry (+ its comment) from `acknowledgedUnloadedSections` in audit_loader_completeness_test.go, coordinated with REQ-DBR-007; OR explicitly list it in §C Out of Scope with a rationale for retention.

**D3. Stale historical comments — internal/defs/dirs_test.go:7,24,47,98,127 — Severity: minor.**
dirs_test.go carries db.yaml historical comments ("design.yaml + db.yaml un-deprecated", count-narrative). REQ-DBR-005 corrects dirs.go comments but leaves dirs_test.go describing db.yaml as un-deprecated live config. Non-build-breaking, does not match closure greps. Consistency gap only.
**Required fix (optional):** add dirs_test.go comment correction to REQ-DBR-005's scope, or note as intentionally-retained historical narrative.

**D4. AC-DBR-017 grep is too broad — acceptance.md:35 — Severity: minor (testability).**
`grep -in "db" CHANGELOG.md` matches any line containing the substring "db" (e.g. "added", "feedback"), so it does not actually verify the manual-deletion guidance for `.moai/project/db/` + `.moai/config/sections/db.yaml`.
**Required fix:** grep for the specific guidance tokens (e.g. `\.moai/project/db/` and `sections/db\.yaml`) instead of bare `db`.

**D5 (INFO, non-blocking). session/phase.go:14 `PhaseDB Phase = "db"`** — a workflow-phase enum value (likely a vestige of the retired `/moai db` command), used only in phase.go's own `Valid()` switch + phase_test.go. Not captured, doesn't match closure greps, plausibly OOS (generic session enum, distinct subsystem). Recommend the SPEC either confirm-and-exclude it in §C or add it, so the "full removal" scope is unambiguous.

**D6 (INFO, non-blocking). Lifecycle hygiene** — SPEC-DB-SYNC-001 / SPEC-DB-SYNC-RELOC-001 are `implemented` (not `superseded`) yet this SPEC removes the subsystem they built. Not a D7 blocker; §H has reconciliation prose. Optional follow-up: transition those two to `superseded`.

---

## Dimension checks explicitly requested

- **Race-safety of the PLAN (dim 6): PASS.** Plan edit scope = {hook.go, dbsync/, hook_e2e_test.go, dirs.go, v2_detection.go, audit_registry.go, internal_content_leak_test.go, settings_test.go, template db/, template db.yaml, local db/, local db.yaml, doc-generation.md ×2, quality-gates-context.md ×2, CHANGELOG.md}. NONE overlap the race-owned set {internal/cli/cc.go, cg.go, glm.go, launcher.go, launcher_test.go, CLAUDE.local.md}. The two fix targets I added (schema_sections_test.go exclusion is grep-only; audit_loader_completeness_test.go) are also not race-owned.
- **Template + mirror parity (dim 4): PASS as specified.** Both guidance files exist in local + template and are currently byte-IDENTICAL (`diff -q` → IDENTICAL, verified). REQ-DBR-013/014/015 mandate editing BOTH copies in the same run-phase commit + re-verifying `diff -q` + `grep SPEC-DB-RETIRE-001 template = 0`.
- **Template neutrality (dim 7): PASS.** REQ-DBR-015 + AC-DBR-015 + plan §B.4 [HARD] + AP-5 enforce §25 (no SPEC-ID / internal token in template edits), with AC-DBR-015 verifying `grep SPEC-DB-RETIRE-001 internal/template/.../workflows/ = 0`.
- **[NEEDS CLARIFICATION] handling (dim 5): CONFIRMED present + correctly deferred, but MP-7 gate is OPEN.** The single marker (plan.md:160, deployed-user db.yaml active-cleanup via DeprecatedPaths) is well-structured and DoD-committed to pre-Kickoff resolution (acceptance.md §D.5), default adoptable with zero SPEC edits. Per MP-7 it nonetheless forces FAIL until the orchestrator runs the `AskUserQuestion` round before Implementation Kickoff Approval.

---

## Chain-of-Verification Pass

Second-look re-checks performed (not skimmed):
- Re-read every REQ-DBR-001…021 for GEARS shape end-to-end (not spot-checked) — all valid.
- Re-verified REQ→AC coverage by walking all 22 ACs against the 21 REQs — complete, no orphan/uncovered.
- Ran a FULL independent Go-tree sweep (`db-schema-sync|dbsync|db.enabled`, `migration_patterns|detected_db|db_schema`, `"db"|db.yaml`) beyond the claimed locations — this surfaced D1 (schema_sections_test.go:292) and D2 (audit_loader_completeness_test.go:16) and D5 (session/phase.go) that spot-checking the 8 claimed locations would have missed.
- Verified the dead-helper claim across the whole repo, catching the same-named-but-separate-package `splitLines`/`trimSpace` in internal/merge + internal/core/project (no over-delete hazard).
- Confirmed OOS section specificity: 5 H3 `### Out of Scope — <topic>` sub-headings, each with concrete `-` bullets (not vague).
- Cross-checked audit-parity test wiring (orphan-scan uses synthetic temp dir; expectedExceptions omits "db") to confirm REQ-DBR-007 is build-safe.

New defects found in second pass: D1, D2, D5 (all from the full sweep vs. the 8 claimed spots).

---

## Recommendation (fixes required before Implementation Kickoff Approval)

1. **[MP-7 gate]** Orchestrator runs an `AskUserQuestion` round on the db.yaml-DeprecatedPaths clarification (plan.md:160) BEFORE Implementation Kickoff Approval. If the user adopts the recommended default ("do not register"), no SPEC edit is needed — the marker is then resolved and this gate clears.
2. **[D1 — major]** Fix AC-DBR-022 + §D.2 + spec.md §D closure grep: exclude `internal/settings/` (schema_sections_test.go's `db.enabled` key-path is preserved OOS), or narrow the `db.enabled` pattern. As written the closure AC cannot pass on a correct removal.
3. **[D2 — major]** Capture `internal/config/audit_loader_completeness_test.go:16` (`acknowledgedUnloadedSections["db"]` + comment) as a removal target coordinated with REQ-DBR-007, OR add it to §C Out of Scope with rationale. This is the missed 5th sibling reference.
4. **[D3/D4/D5 — minor]** Optionally: fold dirs_test.go stale comments into REQ-DBR-005; tighten AC-DBR-017 to grep the specific db paths; confirm session `PhaseDB` OOS in §C.

The SPEC is otherwise strong — clean GEARS, full traceability, correct dead-helper analysis, precise reference-vs-bulk-delete discipline, byte-parity mirror obligation, verified race-safe plan scope, and a correct out-of-scope firewall for the settings/MCP/deployment-migration adjacencies. The FAIL is driven by the MP-7 open gate plus the two verification/scope defects (D1, D2), all of which are cheap to resolve.

---

# Re-Audit — Iteration 2 (DELTA re-verification of the amendment)

Verdict: **FAIL (narrow)** — all 5 prior defects RESOLVED, but the amendment introduced ONE new **major** contradiction (RD-1) + one **minor** hygiene issue (RD-2).
Must-pass firewall: **all 7 PASS**. Aggregate quality ~0.88 (above Tier M 0.80). The FAIL is auditor judgment on a build-relevant requirement contradiction (CN-1), not a must-pass or below-threshold failure — it is one reconciliation edit from PASS.

## Regression check (iter1 defects D1-D5)

| Prior | Status | Evidence |
|-------|--------|----------|
| D1 (closure grep false-fails on settings) | **RESOLVED** | AC-DBR-022 now `... \| grep -v "internal/settings/"` (acceptance.md:40); §D.2 excludes whole `internal/settings/` (acceptance.md:48-50). Independently confirmed the whole-dir widening hides NO real removal target: `grep -rn "dbsync\|db-schema-sync\|runDBSchemaSync" internal/settings/` = 0 matches; the only db-pattern hit under `internal/settings/` is `schema_sections_test.go:292 "db.enabled"` (OOS key-path test, correctly preserved). |
| D2 (audit_loader_completeness_test.go allowlist) | **RESOLVED** | REQ-DBR-007 now binds BOTH maps: `yamlAuditExceptions` (real-tree) AND `acknowledgedUnloadedSections` (template-tree) (spec.md:53). AC-DBR-007 greps both: `'"db":' audit_registry.go = 0 AND '"db",' audit_loader_completeness_test.go = 0 AND go test -run 'TestAuditParity\|TestAuditLoaderCompleteness'` (acceptance.md:25). Both live entries confirmed present (audit_registry.go:74, audit_loader_completeness_test.go:16) and each is the sole `"db"` occurrence in its file. Template-first ordering correctly required (plan §C.5(a), AP-10). |
| D3 (dirs_test.go stale comments) | **RESOLVED (folded)** | Absorbed into REQ-DBR-022 atomic dirs_test.go update ("파생 주석 정정 — D3 = 여기 흡수", plan §B.8 count-row + M2.5). |
| D4 (AC-017 grep too broad) | **RESOLVED** | AC-DBR-017 narrowed to `grep -inE "db\.yaml\|\.moai/project/db\|db-schema-sync"` with a two-clause content check (acceptance.md:35). |
| D5 (PhaseDB uncaptured) | **RESOLVED** | spec.md §C adds `### Out of Scope — session.PhaseDB` (spec.md:111-113), correctly classifying `session/phase.go:14 PhaseDB="db"` as a `/moai db` retired-command vestige in the session-phase enum, distinct from the DB-doc subsystem. |

All 5 iter1 defects resolved. No stagnation.

## Coordinator-requested item verification

1. **MP-7 cleared** — PARTIAL. The REAL marker is gone: `grep -rn '\[NEEDS CLARIFICATION:' .moai/specs/SPEC-DB-RETIRE-001/*.md` (with colon) returns 0 in the 4 artifacts (only my own plan-audit-report.md quotes the old marker). BUT the exact coordinator grep `grep -rn '\[NEEDS CLARIFICATION' <4 artifacts>` returns **4 matches**, all benign meta-references stating "0 markers remain" (progress.md:17, acceptance.md:96, plan.md:127, plan.md:194). Substantively MP-7 is cleared (no open `[NEEDS CLARIFICATION: <topic>]` question); mechanically the literal `[NEEDS CLARIFICATION]` token in the resolution prose collides with the detector → see RD-2.
2. **D1 fixed** — CONFIRMED both (a) and (b). (a) `internal/settings/` whole-dir exclusion no longer false-matches `schema_sections_test.go:292 "db.enabled"`. (b) Independently grepped `internal/settings/` for `dbsync\|db-schema-sync\|runDBSchemaSync` → 0 matches, so the widened exclusion hides NO genuine removal target (verified, not assumed).
3. **D2 fixed** — CONFIRMED. AC-DBR-007 covers both audit maps (real-tree `yamlAuditExceptions` + template-tree `acknowledgedUnloadedSections`), each with its own template-first/local-first ordering dependency (plan §C.5).
4. **REQ-DBR-022/023/024 independent verification:**
   - (a) `internal/defs/dirs.go:61` — `var DeprecatedPaths = []DeprecatedPathEntry{` CONFIRMED.
   - (b) `internal/cli/update_cleanup.go:126` — `func scanDeprecatedPaths(projectRoot string)` CONFIRMED.
   - (c) `internal/cli/deprecated_paths_collision_test.go:53` — `func TestDeprecatedPaths_NoTemplateCollision` CONFIRMED.
   - (d) New ACs mechanically verifiable: AC-023(a) grep dirs.go for the entry ✓; AC-023(b) clean-reinstall removal test (TDD-add — the run-phase must author the seed-and-remove case, softer but acceptable) ✓; AC-023(c) `go test ./internal/defs/... -run TestDeprecatedPaths` (count 39→40) ✓; AC-024 `go test -run TestDeprecatedPaths_NoTemplateCollision` ✓.
   - (e) `TestDeprecatedPathsCategorySplit` caveat CORRECTLY flagged as a run-phase decision (plan §B.8 bucket-caveat + AP-11). Verified the test (dirs_test.go:51) has 3 `DeprecatedSince` buckets AND a `default: t.Errorf("...unexpected DeprecatedSince value...")` (dirs_test.go:208-222) — so a new 4th bucket `SPEC-DB-RETIRE-001` WOULD build-break unless the split test is reconciled. The plan surfaces exactly this ("(a) reuse an existing bucket vs (b) add a 4th bucket case, then reconcile dirs_test.go") — not a silent break. Correct.
5. **Ordering dependencies (plan §C.5)** — sound. (a) template db.yaml delete (REQ-011) BEFORE DeprecatedPaths register (REQ-022, collision guard) AND BEFORE `acknowledgedUnloadedSections["db"]` delete (REQ-007, loader-completeness scans template dir) — both correctly template-first. (b) local db.yaml delete (REQ-016) ⟷ `yamlAuditExceptions["db"]` delete (REQ-007, TestAuditParity scans real tree) — correctly local-coordinated. (c) DeprecatedPaths slice ⟷ dirs_test count atomic (@MX:ANCHOR). AP-9/AP-10/AP-11 codify each. All three ordering rules are correct and self-consistent — with the RD-1 count-value exception below.
6. **D3-D5 confirmed resolved/OOS** — see regression table (D5 `PhaseDB="db"` correctly OOS).

## New defects introduced by the amendment

**RD-1. DeprecatedPaths count contradiction (baseline 40 vs live 39) — spec.md:51 / acceptance.md:23 / plan.md:41 — Severity: major (CN-1 contradiction, build-relevant).**
The live tree is authoritative: `internal/defs/dirs.go` slice has exactly **39** entries; `internal/defs/dirs_test.go:32 const want = 39` (9 Cat-A + 27 Cat-B + 3 Cat-C = 39). The `40→39` reduction landed with SPEC-CONFIG-AUDIT-REPAIR-001 (dirs_test.go:29 comment). The amendment's NEW parts state the baseline correctly — REQ-DBR-022 / §B.8 / M2.5 / AC-DBR-023 all say "total **39→40**" — but the pre-existing REQ-DBR-005 / AC-DBR-005 / plan §B.2 were NOT reconciled and still assert the STALE "**40**-entry count 불변" + "db.yaml was never a slice entry — comment-only reference." Two problems compound:
  - **Factual error**: the baseline is 39, not 40 (stale premise predating CONFIG-AUDIT-REPAIR-001).
  - **Contradiction**: REQ-DBR-005 asserts the slice count "shall remain unchanged" and "db.yaml was never a slice entry", but REQ-DBR-022 ADDS db.yaml as a slice entry and changes the count 39→40 — the two requirements cannot both hold in the final state (CN-1).
  - **Build-relevance**: plan §B.2 tells the implementer "40-entry count 불변 → dirs_test.go count 수정 불필요" (no count edit). An implementer trusting that either (i) skips the dirs_test.go update entirely (leaving `want=39` against a 40-element slice → `TestDeprecatedPaths` FAIL), or (ii) computes 40+1=41 (→ FAIL). Only cross-referencing REQ-022/§B.8's correct "39→40" + the live `const want=39` avoids the break.
**Required fix:** (1) reword REQ-DBR-005 so its "count unchanged" claim is scoped to REQ-005's own comment-only edits and drop "(db.yaml was never a slice entry — comment-only reference)" — REQ-022 makes it a slice entry; (2) correct AC-DBR-005's parenthetical from "(40-entry count 불변)" to "(REQ-005 is comment-only; DeprecatedPaths baseline is 39; REQ-022 separately takes it 39→40)"; (3) fix plan §B.2 "40-entry count 불변 → dirs_test.go count 수정 불필요" to "39-entry baseline; comment-only here, but REQ-022 does edit the count 39→40".

**RD-2. MP-7 meta-reference grep collision — plan.md:127, plan.md:194, acceptance.md:96, progress.md:17 — Severity: minor (verification hygiene).**
The resolution prose embeds the literal token `[NEEDS CLARIFICATION]` in "마커 0건" statements, so the mechanical MP-7 detector `grep -rn '\[NEEDS CLARIFICATION' <artifacts>` returns 4 matches instead of the expected 0 — a superstring collision that would false-trip an automated clarification gate at the run-gate. There is NO actual open clarification (the `[NEEDS CLARIFICATION: db.yaml DeprecatedPaths]` topic marker is genuinely resolved + removed).
**Required fix:** reword the four "0 markers remain" statements to avoid the literal bracketed token (e.g. "clarification markers: none remaining" / "미해결 clarification 없음"), so the mechanical `grep '\[NEEDS CLARIFICATION'` returns a clean 0.

## Must-Pass firewall (iter2)

- [PASS] MP-1 REQ consistency — REQ-DBR-001..022, complete, no gap/dup (note: §B.8 REQ-022 is authored before §B.7 REQ-018..021 in section order, but the REQ *numbers* 001-022 are contiguous — cosmetic ordering only).
- [PASS] MP-2 GEARS — REQ-022 is a valid Event-driven + State-driven compound ("When moai update … shall remove …; While the template still ships db.yaml … shall not be added").
- [PASS] MP-3 frontmatter — unchanged 12-field schema.
- [N/A→PASS] MP-4 — single-language.
- [PASS] MP-5 D7 — referenced SPECs all implemented/completed; no BLOCKING.
- [N/A→PASS] MP-6 D8 — no `syscall` in artifacts.
- [PASS] MP-7 clarification gate — substantively cleared (no open `[NEEDS CLARIFICATION: <topic>]`); the mechanical grep-collision is RD-2 (minor reword), not an open question.

## Chain-of-Verification Pass (iter2)

Second-look re-checks: re-counted the live `DeprecatedPaths` slice (`awk` bounded extraction = 39 `Path:` entries) and cross-checked against `const want = 39` + the Cat A/B/C subtotals (9+27+3) — this is what surfaced RD-1 (I under-verified the count in iter1 because REQ-005 was comment-only and "count unchanged" read as low-risk; REQ-022 now makes the exact baseline load-bearing). Re-ran the colon-qualified marker grep to distinguish a real marker from the meta-references (RD-2). Verified the CategorySplit `default` branch (`t.Errorf` on unexpected `DeprecatedSince`) to confirm the plan's 4th-bucket caveat is a genuine, correctly-flagged hazard rather than a phantom.

## Recommendation (iter2 → before Implementation Kickoff Approval)

1. **[RD-1, major]** Reconcile the DeprecatedPaths count: baseline is **39** (live `const want=39`), not 40. Fix REQ-DBR-005 (scope its count claim + drop "never a slice entry"), AC-DBR-005 ("40-entry count 불변" → 39 baseline + note REQ-022 does 39→40), and plan §B.2 ("count 수정 불필요" → REQ-022 edits it). REQ-022/§B.8/AC-023 are already correct ("39→40") — align the stale REQ-005 side to them.
2. **[RD-2, minor]** Reword the four "0 markers remain" statements to drop the literal `[NEEDS CLARIFICATION]` token so the mechanical MP-7 grep returns clean 0.

Both fixes are localized text edits (no structural change). This is a narrow FAIL: all 5 prior defects are resolved, the REQ-022 active-deletion mechanism is fully verified against the live tree, the ordering dependencies are sound, and the must-pass firewall is clean — RD-1 (a genuine build-relevant contradiction) is the sole blocker, RD-2 a hygiene follow-on. An orchestrator that judges RD-1 self-correcting (authoritative live `const want=39` + correct REQ-022 statements) could alternatively route this as PASS-with-debt; the adversarial call here is FAIL to force the count reconciliation before run-phase.
