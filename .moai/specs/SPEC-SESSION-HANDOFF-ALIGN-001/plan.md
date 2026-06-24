# Plan — SPEC-SESSION-HANDOFF-ALIGN-001

> Tier M implementation plan. Era V3R6 4-phase lifecycle (plan → run → sync → Mx).
> Primary artifacts: `.md` edits in two trees + 1 Go test allowlist append. Near-zero production code.

## §A. Context

See `spec.md` §A/§B for the full problem statement and scope. Condensed: a 7-agent structural analysis identified a 105-line net template↔local content delta on `session-handoff.md` (111 change-line diff / 117 raw-diff lines), a systemic mirror-coverage gap (1 of 17 mirrored workflow/ files enrolled in the CI parity test; 1 LOCAL-ONLY template-missing file `lifecycle-sync-gate.md` identified at iter-2), and an internal-duplication + i18n debt cluster. The canonical FORMAT (L1-126) is healthy; the maintenance burden is in the three appended local-only content blocks (Diet, V0, `/cd` cache-preserving) and the duplication they introduced.

## §B. Known Issues (pre-run discovery)

- **K1**: The V0 "Cross-pollination 이력" block (L213-219) embeds internal SPEC-IDs (LIFECYCLE-SYNC-GATE-001, SESSION-AUTO-RESUME-001, HARNESS-NAMESPACE) AND is the exact AP-D-002 shape the file forbids. Both must be resolved: the SPEC-IDs by stripping (EXCL-003), the meta-irony by collapsing to a 1-line lesson reference (REQ-SHA-014). These are two distinct fixes on the same block.
- **K2**: The 3 stale local SPEC-ID lines (L68, L69, L122) reference `SPEC-V3R6-MULTI-SESSION-COORD-001`, which IS shipped (the feature is in the template). The template's neutralization is correct; local is the laggard. Direction: local → template, never the reverse.
- **K3**: The cut-line marker literal appears in 4 sections (L27, L47-54, L113, L124-125). De-duplication MUST preserve the SSOT body verbatim and only replace the redundant re-spellings with pointers — the `✂` / `─` Unicode characters are load-bearing and any byte drift fails the mirror test post-enrollment.
- **K4**: The mirror test's allowlist (L42-64) documents a §25 carve-out pattern — 6 files were REMOVED from byte-parity because they retain internal-content tokens in the working copy. session-handoff.md must NOT join that carve-out; it must achieve true byte-parity via neutralization.

## §C. Pre-flight (before M1)

- [ ] Confirm both target files match the iter-2 reconciled line counts (local=314, template=209, change-only diff=111, raw diff=117 — per research.md §A.0 verbatim output).
- [ ] Confirm `rule_template_mirror_test.go` line 46 is the only workflow/ entry (iter-2 confirmed: the only enrolled workflow/ file is `spec-workoff.md`; the comment block at L59-62 lists 3 §25-carve-out files but they are NOT in the byte-parity allowlist).
- [ ] Confirm the 3 adjacent SPECs remain completed and orthogonal (no scope collision).
- [ ] Read `internal/template/internal_content_leak_test.go` to enumerate the exact forbidden token patterns the neutrality guard rejects (SPEC-ID regex, REQ/AC tokens, internal dates, macOS-bias paths).

## §D. Constraints (carry-forward from spec.md §D)

1. Template-First: template edits land FIRST in each milestone, local mirrors in the SAME commit.
2. §25 neutrality: zero internal tokens in the template tree post-port.
3. No behavioral change to FORMAT (L1-126) beyond the 4 explicit realignments.
4. Era V3R6 lifecycle: progress.md §E.1 at plan-phase, §E.2/§E.3 at run, §E.4 at sync, §E.5 at Mx.
5. Ownership Matrix: manager-spec authors plan-phase; manager-develop owns M1-M6; manager-docs owns sync; orchestrator-direct OR manager-docs owns Mx.
6. Mirror-test enrollment (M4) MUST follow parity (M1-M3).

## §E. Self-Verification (manager-develop §E.1 placeholder — populated at run-phase)

This section is a placeholder. The run-phase manager-develop populates §E.1 audit-ready signal, §E.2 run-phase evidence, and §E.3 run-phase audit-ready signal in `progress.md`. Per the Status Transition Ownership Matrix, this plan.md body is owned by manager-spec and is NOT modified mid-run by manager-develop.

## §F. Milestones (Tier M, M1-M6)

### M1 — Coverage audit + SPEC-ID realignment (priority High)

**Scope**: Establish the evidence base and close the cleanest drift.
- Produce the 18-file workflow/ coverage audit table (research.md §A, iter-2 reconciled). The audit is observational — it does NOT authorize cleanup of the 16 in-sync mirrored siblings, AND it does NOT authorize porting the 1 template-missing file (`lifecycle-sync-gate.md`, per EXCL-006).
- Realign the 3 stale local SPEC-ID lines (L68, L69, L122) to the template's already-generic phrasing. Single-tree local edit; template is already correct (REQ-SHA-005). Verify via AC-SHA-005's WHOLE-FILE grep (iter-2: the iter-1 `lines 60-130` window is dropped because it missed L122 and was unsatisfiable due to the heading-order delta inside the window).
- Verify the canonical FORMAT section (L1-126) is now byte-identical between the two trees on the edited lines.

**AC binding**: AC-SHA-005 (whole-file grep, iter-2 rewritten), AC-SHA-008 (audit, iter-2 reconciled to 18 files + verbatim output), AC-SHA-012 (trigger #1 pointer is deferred to M5 but the audit informs it).
**Ownership**: manager-develop (M1 commit). `draft → in-progress` frontmatter transition on M1 commit start.

### M2 — Diet Constraints neutralized port (priority High)

**Scope**: Port the Diet generic core to template, strip internal tokens, mirror back to local.
- Template: insert the Diet Constraints section AFTER Worktree-Anchored Resume Pattern (NOT at local's mid-file position — the template's natural order is correct). Strip the L129 SPEC-ID parenthetical and the L183 scope bullet naming 3 internal SPEC lines (REQ-SHA-001, REQ-SHA-002).
- Local: replace the SPEC-ID-bearing Diet opening with the neutralized version from template; the per-block budgets, AP-D-001..005, and 8-item Pre-emit self-check stay verbatim (REQ-SHA-006).
- The local file's Diet section physically moves to AFTER Worktree-Anchored in M4 (structural reorganization); M2 is content-neutralization only, not section relocation.

**AC binding**: AC-SHA-001, AC-SHA-002, AC-SHA-006a (content-parity half — iter-2 split per D5).
**Ownership**: manager-develop.

### M3 — V0 Abort Gate neutralized port (priority High)

**Scope**: Port the V0 generic core to template, drop the dev-incident block, mirror back to local.
- Template: insert the V0 Abort Gate section AFTER the (newly ported) Diet section. Opening minus SPEC-ID parenthetical; the 3 canonical V0-a/b/c commands verbatim; abort obligation; AP-V-001..003 verbatim; AP-V-004 generic lesson retained, internal-file provenance stripped (REQ-SHA-003, REQ-SHA-004).
- Template: DROP entirely the "Cross-pollination 이력" 5-line block — it is 100% internal incident log.
- Local: replace the V0 opening with the neutralized version; collapse the Cross-pollination 이력 block to a 1-line lesson reference (REQ-SHA-014 — the meta-irony fix). Mirror the neutralized AP-V-004 back so both trees agree.

**AC binding**: AC-SHA-003, AC-SHA-004, AC-SHA-006a (content-parity half — iter-2 split per D5), AC-SHA-014.
**Ownership**: manager-develop.

### M4 — Mirror-test enrollment + local restructure (priority High)

**Scope**: Lock the parity invariant; finish the reader-flow fix; achieve whole-file byte-parity.
- Port the `### /cd cache-preserving alternative (CC 2.1.169+)` subsection (LOCAL L249) to the TEMPLATE (REQ-SHA-006 scope, iter-2 added per research.md §B.3 — the block is generic CC-2.1.169+ platform documentation with no internal tokens, so it ports verbatim with no stripping). Mirror-back to local (identity operation since local is the source).
- Enroll `.claude/rules/moai/workflow/session-handoff.md` in `rule_template_mirror_test.go`'s `workflowOptMirroredPaths` allowlist (REQ-SHA-007). Add a comment explaining the enrollment origin and the §25 neutrality contract.
- Run `go test ./internal/template/... -run TestRuleTemplateMirrorDrift` and confirm GREEN. If red, the M2/M3/M4 neutralization missed a token — halt and fix before proceeding.
- Restructure the LOCAL file: move Diet + V0 sections from their mid-file position (currently between Anti-Patterns and Worktree-Anchored) to AFTER Worktree-Anchored and BEFORE Cross-references (REQ-SHA-016). The template already has the correct order post-M2/M3; local converges.
- After restructure, re-run the mirror test — must remain GREEN (restructure is byte-preserving on content, only section order changes; both trees move in lockstep).

**AC binding**: AC-SHA-007, AC-SHA-016, AC-SHA-006b (whole-file byte-parity post-restructure — iter-2 split per D5).
**Ownership**: manager-develop.

### M5 — i18n + dedup consolidation (priority Medium)

**Scope**: Pay down the duplication and i18n debt.
- Replicate the 4-locale Header translation table from `moai.md §8` into the Localization Table section of BOTH copies (REQ-SHA-010). Add Block 1 entering verb row (entering / 진입 / 開始 / 进入).
- Replace the Korean-locked skeleton verb `진입` (L32, L85, L284) with `<entering verb>` placeholder in BOTH copies (REQ-SHA-011). The Example blocks render the placeholder + a note "translates per header table".
- De-duplicate cut-line marker spec: keep full literal ONLY in § Cut-line Marker Specification; replace re-spellings at § Canonical Format intro, § Output Surface, § Anti-Patterns with pointers (REQ-SHA-009).
- Replace inline Trigger #1 model-class numbers (L17) with pointer to `context-window-management.md § Context Window Targets` in BOTH copies (REQ-SHA-012).
- Consolidate the 3 anti-pattern catalogues with cross-link pointers (REQ-SHA-013). Add the `moai.md §8` forward-link to Cross-references in both copies (REQ-SHA-015).
- After all edits, re-run mirror test — must remain GREEN.

**AC binding**: AC-SHA-009, AC-SHA-010, AC-SHA-011, AC-SHA-012, AC-SHA-013, AC-SHA-015.
**Ownership**: manager-develop.

### M6 — Verification + neutral CI guard (priority High)

**Scope**: Verify all three axes closed; confirm the neutrality guard passes.
- Run `go test ./internal/template/...` (full template test suite — mirror drift + leak detection).
- Run the neutrality CI guard locally (`.github/workflows/template-neutrality-check.yaml` equivalent).
- Grep the template copy for forbidden token patterns: internal SPEC-IDs (`SPEC-V3R6-*`, `SPEC-V3R5-*`, `SPEC-V3R2-*` outside historical context), REQ/AC tokens, commit SHAs, internal dates, macOS-bias paths (`/Users/`). Expected: zero findings on the ported content.
- Grep both copies for the `Cross-pollination 이력` block — expected: present only as the 1-line lesson reference, the 5-line narrative gone.
- Confirm the mirror test passes green on the final commit.
- Produce §E.2 run-phase evidence (commands + verbatim output) for the progress.md.

**AC binding**: AC-SHA-002 (neutrality grep), AC-SHA-014 (cross-pollination collapse), all AC final verification.
**Ownership**: manager-develop. On M6 completion, manager-develop reports run-phase audit-ready signal; orchestrator hands off to manager-docs for sync.

### Sync phase (manager-docs)

- Update CHANGELOG.md (under the appropriate release section).
- Update README.md / docs-site if session-handoff.md is surfaced there (check docs-site i18n rules — likely no docs-site change needed for an internal rule file).
- frontmatter `in-progress → implemented` transition on sync commit.

### Mx phase (manager-docs OR orchestrator-direct)

- §E.4 sync-phase audit-ready signal backfill (`sync_commit_sha`).
- §E.5 Mx-phase audit-ready signal (`mx_commit_sha`).
- `implemented → completed` transition. Close commit subject names the full SPEC-ID: `chore(SPEC-SESSION-HANDOFF-ALIGN-001): Mx-phase audit-ready signal + 4-phase close`.

## §G. Anti-Patterns (specific to this SPEC)

- **AP-1**: Porting Diet/V0 to the template without stripping internal tokens, then enrolling in the mirror test — produces a red test OR a leak-test failure. Neutralization FIRST, enrollment LAST.
- **AP-2**: Editing only one tree and relying on a later commit to sync the other — violates the Template-First same-commit invariant and risks a transient red mirror test.
- **AP-3**: "Cleaning up" the canonical FORMAT section (L1-126) beyond the 4 explicit realignments — violates constraint #3 and risks unintended behavioral drift in the paste-ready emission contract.
- **AP-4**: Enrolling the file in the mirror test, THEN discovering the §25 leak test fails on a residual token, THEN adding session-handoff.md to the §25 carve-out (L50-63) as a workaround — the carve-out is for files that CANNOT be neutralized; session-handoff.md CAN and MUST achieve true byte-parity.
- **AP-5**: Restructuring local (M4) and template independently — the restructure must be byte-preserving on content and land in the same commit on both trees.

## §H. Cross-References

- spec.md §B (scope), §C (REQ-SHA-001..016), §D (constraints)
- acceptance.md §D (AC-SHA-001..016 matrix)
- research.md §A (18-file coverage audit table, iter-2 reconciled with verbatim output + lifecycle-sync-gate.md template-missing row), §B (Diet/V0/`/cd` neutral-vs-dev content classification, iter-2 added the `/cd` block per §B.3)
- design.md §A (mixed-content split strategy), §B (token-stripping enumeration), §C (i18n placeholder reconciliation), §D (anti-pattern catalogue consolidation approach)
- progress.md §E.1..§E.5 (lifecycle placeholders)
- `.claude/rules/moai/development/spec-frontmatter-schema.md § Status Transition Ownership Matrix` (canonical owner per transition)
