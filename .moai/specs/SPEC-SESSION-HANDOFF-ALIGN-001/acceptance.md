# Acceptance — SPEC-SESSION-HANDOFF-ALIGN-001

> Given-When-Then acceptance criteria. MUST-PASS criteria MUST all pass for the SPEC to close.
> Tier M — every REQ has at least one AC; MUST ACs are gating, SHOULD ACs are non-gating quality improvements.

## §D. AC Matrix

### Axis 1 — Template↔Local drift closure

**AC-SHA-001** (MUST, REQ-SHA-001) — Diet generic core ships to users
- **Given** the template tree at `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md`
- **When** a maintainer runs `grep -c 'AP-D-001\|AP-D-002\|AP-D-003\|AP-D-004\|AP-D-005'` on the template copy
- **Then** the count is ≥5 (the full AP-D catalogue is present)
- **And** `grep -c 'Block 2 ≤ 4 references\|Block 4 각 precondition ≤ 200 chars\|Block 5 single primary action\|Block 6 ≤ 2 lines'` returns ≥4 distinct budget rules
- **And** `grep -c 'Pre-emit self-check'` returns ≥1 with the 8-item checklist present in the section.

**AC-SHA-002** (MUST, REQ-SHA-002) — Diet port is neutrality-clean
- **Given** the template copy post-M2
- **When** a maintainer runs `grep -nE 'SPEC-V3R[0-9]-[A-Z]|LIFECYCLE-SYNC-GATE|HARNESS-NAMESPACE|SESSION-AUTO-RESUME'` on the Diet section
- **Then** zero matches (the L129 parenthetical and L183 scope bullet are neutralized to generic phrasing)
- **And** the template passes `go test ./internal/template/... -run TestTemplateNoInternalContentLeak` (the leak test).

**AC-SHA-003** (MUST, REQ-SHA-003) — V0 generic core ships to users
- **Given** the template copy post-M3
- **When** a maintainer runs `grep -c 'lsof -a -c claude'` on the V0 section
- **Then** the count is ≥2 (V0-b and V0-c canonical commands present)
- **And** `grep -c 'STRICT 0\|STRICT ≤2'` returns ≥2 (both thresholds documented)
- **And** `grep -c 'AP-V-001\|AP-V-002\|AP-V-003\|AP-V-004'` returns ≥4 (full AP-V catalogue present, AP-V-004 carrying its generic COMMAND-column-filter lesson).

**AC-SHA-004** (MUST, REQ-SHA-004) — V0 port drops dev-incident provenance
- **Given** the template copy post-M3
- **When** a maintainer greps the V0 section for `'Cross-pollination\|Hugo docs\|claude-md-guide\|M4 1·2차\|LIFECYCLE-SYNC-GATE-001 M4'`
- **Then** zero matches (the 5-line history block and AP-V-004's internal-file trailing provenance are gone)
- **And** the V0 section's "Cross-pollination 이력" block (if retained at all) is a single 1-line lesson reference, not a 5-bullet narrative.

**AC-SHA-005** (MUST, REQ-SHA-005) — 3 stale local SPEC-ID lines realigned (whole-file grep, no windowed diff)
- **Given** the local copy at `.claude/rules/moai/workflow/session-handoff.md` post-M1 AND the template copy at `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` (which already has zero SPEC-V3R6-MULTI-SESSION-COORD-001 / REQ-COORD-009 matches — the template's neutralization is the ground truth LOCAL converges toward)
- **When** a maintainer runs `grep -nE 'SPEC-V3R6-MULTI-SESSION-COORD-001|REQ-COORD-009' .claude/rules/moai/workflow/session-handoff.md` (WHOLE-FILE grep, no line window)
- **Then** zero matches (all 3 stale lines — previously at LOCAL L68, L69, L122 — are realigned to the template's generic "the canonical multi-session coordination policy" phrasing)
- **And** `grep -cE 'SPEC-V3R6-MULTI-SESSION-COORD-001|REQ-COORD-009' internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` returns 0 (the template remains the neutral ground truth; this clause guards against accidental re-introduction of the SPEC-ID into the template during the same-commit mirror-back)
- **Note**: the iter-1 AC used a `lines 60-130` window and a `sed -n '60,130p'` diff clause. That window was both **non-discriminating** (it missed the third stale line at L122, which sits OUTSIDE the 60-130 window) and **unsatisfiable as written** (the window covers the Diet/Worktree-Anchored heading-order delta, which is non-zero pre-M4 regardless of REQ-SHA-005 satisfaction). iter-2 rewrites the AC to a whole-file grep that fails-loud in the correct direction: any residual SPEC-ID-bearing line anywhere in the LOCAL file fails the AC.

**AC-SHA-006a** (MUST, REQ-SHA-006) — Diet + V0 + `/cd` section-body content parity (post-M2/M3)
- **Given** both copies post-M2 (Diet port) and post-M3 (V0 port) and post the `/cd` cache-preserving port (REQ-SHA-006 scope, iter-2 added per research.md §B.3)
- **When** a maintainer extracts the Diet Constraints section body, the V0 Abort Gate Doctrine section body, and the `### /cd cache-preserving alternative (CC 2.1.169+)` subsection body from each tree and diffs them pairwise (ignoring heading-order context — only the section BODY bytes are compared)
- **Then** each of the 3 section bodies is byte-identical between local and template (the local SPEC-ID openings have been replaced by the neutralized template versions; the `/cd` block — which had no internal tokens — is identical by the direct port)
- **Note**: this AC verifies CONTENT parity of the 3 neutralized appendix blocks. The SECTION-ORDER parity (the structural relocation of Diet + V0 from mid-file to after-Worktree-Anchored) is a separate property verified by AC-SHA-006b post-M4. The split is required because content parity is achievable at M2/M3 while structural parity is only achievable at M4 (the relocation milestone) — binding both to M2/M3 made the iter-1 AC unsatisfiable at its declared milestone.

**AC-SHA-006b** (MUST, REQ-SHA-006 + REQ-SHA-016) — Post-restructure whole-file byte-parity (post-M4)
- **Given** both copies post-M4 (the structural relocation of Diet + V0 to after-Worktree-Anchored has landed in both trees)
- **When** a maintainer runs `diff .claude/rules/moai/workflow/session-handoff.md internal/template/templates/.claude/rules/moai/workflow/session-handoff.md`
- **Then** the diff is empty (the two files are byte-identical: content parity from AC-SHA-006a + structural parity from REQ-SHA-016 + the `/cd` block parity + the 3 stale-SPEC-ID realignment from REQ-SHA-005)
- **And** the only residual LOCAL↔TEMPLATE delta is zero (not "structural" — there is no structural delta post-M4 either, because both trees moved in lockstep in the M4 commit)
- **Note**: iter-1's AC-SHA-006 claimed "the only remaining delta is structural" while bound to M2/M3; that was internally inconsistent (the structural delta is what M4 resolves, not M2/M3). iter-2 splits the "structural" clause out to this M4-bound AC so each clause maps to a milestone where it is individually satisfiable.

### Axis 2 — Mirror-coverage gap

**AC-SHA-007** (MUST, REQ-SHA-007) — session-handoff.md enrolled in mirror test
- **Given** `internal/template/rule_template_mirror_test.go` post-M4
- **When** a maintainer runs `grep -n 'session-handoff.md' internal/template/rule_template_mirror_test.go`
- **Then** at least one match exists in either `workflowOptMirroredPaths` or a dedicated allowlist slice
- **And** `go test ./internal/template/... -run TestRuleTemplateMirrorDrift -v` passes GREEN on the session-handoff.md subtest
- **And** artificially introducing a 1-byte drift in EITHER copy causes the test to FAIL with the `RULE_TEMPLATE_MIRROR_DRIFT` sentinel (verified by a manual probe, then reverted).

**AC-SHA-008** (SHOULD, REQ-SHA-008) — 18-file coverage audit table produced (iter-2 reconciled)
- **Given** research.md §A (with the iter-2 §A.0 verbatim audit-command output pasted, NOT a hand-transcribed table)
- **When** a reviewer reads the audit table
- **Then** all 18 LOCAL workflow/ files appear with `local_lines`, `template_lines` (where mirrored), `diff_lines` (both the change-only `^[<>]` metric AND the raw `wc -l` metric), and a severity classification (in-sync / minor-drift / major-drift / local-only-sections / template-missing)
- **And** the table identifies session-handoff.md as the sole major-drift file pre-SPEC, with all 16 in-sync mirrored siblings at diff=0, AND identifies lifecycle-sync-gate.md as a `template-missing` LOCAL-ONLY case (iter-2 NEW — iter-1 missed this row entirely)
- **And** the verbatim audit-command output (§A.0) is present so the evidence base is reproducible, not hand-transcribed.

### Axis 3 — Internal duplication + i18n debt

**AC-SHA-009** (MUST, REQ-SHA-009) — Cut-line marker spec de-duplicated
- **Given** both copies post-M5
- **When** a maintainer greps for the literal `✂──── 여기부터 복사 ────✂` outside fenced example blocks
- **Then** the full marker spec appears ONLY in § Cut-line Marker Specification (L47-54 region) and the fenced Example blocks (L29-45, L82-98, L277-296) — the § Canonical Format intro, § Output Surface, and § Anti-Patterns sections reference it via "see § Cut-line Marker Specification" pointers instead of re-spelling the literal.

**AC-SHA-010** (MUST, REQ-SHA-010) — 4-locale Header translation table replicated
- **Given** both copies post-M5
- **When** a maintainer reads the Localization Table section
- **Then** the table contains Block 1 entering verb, Block 3 Preconditions header, Block 5 Run header, and Block 6 workflow-context headers — each with all 4 locale renderings (en / ko / ja / zh)
- **And** the table content is consistent with `moai.md §8`'s header renderings (cross-verification, not byte-identity — moai.md §8 is the render surface, session-handoff.md is the SSOT).

**AC-SHA-011** (MUST, REQ-SHA-011) — Skeleton verb placeholder replaces Korean lock-in
- **Given** both copies post-M5
- **When** a maintainer greps the 6-block skeleton (the canonical Format block at L25-45, not the Example) for `진입`
- **Then** zero matches in the skeleton; the verb is `<entering verb>` placeholder
- **And** the Example blocks render `<entering verb>` OR explicitly note the verb translates per the header table (Korean example may retain `진입` if clearly marked as an illustrative ko-rendering, matching moai.md §8 discipline).

**AC-SHA-012** (MUST, REQ-SHA-012) — Trigger #1 model-label drift eliminated
- **Given** both copies post-M5
- **When** a maintainer reads the Trigger #1 row (L17 region)
- **Then** the row references `.claude/rules/moai/workflow/context-window-management.md § Context Window Targets` for the threshold instead of inlining model-class numbers
- **And** `grep -E 'Opus 4\.7|Opus 4\.8'` on the session-handoff.md file returns zero matches in the Trigger #1 row (the label drift vector is killed at the source).

**AC-SHA-013** (SHOULD, REQ-SHA-013) — Anti-pattern catalogues consolidated
- **Given** both copies post-M5
- **When** a maintainer reads any one of the three anti-pattern surfaces (general prose L115-125, AP-D catalogue, AP-V catalogue)
- **Then** the section references the other two via cross-link pointers (e.g., "see also § Diet Constraints Anti-pattern catalogue for paste-ready-specific patterns")
- **And** a reader encountering one catalogue can discover the other two without ctrl-F'ing.

**AC-SHA-014** (MUST, REQ-SHA-014) — Cross-pollination meta-irony collapsed
- **Given** the local copy post-M3
- **When** a maintainer greps for the 5-line "Cross-pollination 이력" narrative
- **Then** zero matches on the 5-bullet block (`Line C.*9차`, `Line C.*10차`, `Line A.*13차`, `Line B.*14차`)
- **And** a single 1-line lesson reference replaces it (e.g., "Cross-line provenance: retained in lesson memory; this section codifies the doctrine.")
- **And** the file no longer self-violates AP-D-002 (the file's own anti-pattern forbidding history/lesson narrative in paste-ready-adjacent prose).

**AC-SHA-015** (SHOULD, REQ-SHA-015) — moai.md §8 forward-link added
- **Given** both copies post-M5
- **When** a maintainer reads the Cross-references section
- **Then** an entry exists pointing to `.claude/output-styles/moai/moai.md §8 (Response Templates → Session Handoff)` with a note that session-handoff.md is the SSOT and moai.md §8 is the canonical render surface
- **And** the previously one-sided bidirectional link is now bidirectional.

**AC-SHA-016** (SHOULD, REQ-SHA-016) — Local file reader-flow restored
- **Given** the local copy post-M4
- **When** a maintainer reads the file top-to-bottom
- **Then** the heading order is: ... Anti-Patterns → Worktree-Anchored Resume Pattern → Diet Constraints → V0 Abort Gate Doctrine → Cross-references (canonical content first, appendices grouped after)
- **And** the Diet + V0 sections no longer sit between Anti-Patterns and Worktree-Anchored (the mid-file doctrine insertion is resolved).

## §D.1 Severity classification

| AC | Severity | Gating? | Milestone |
|----|----------|---------|-----------|
| AC-SHA-001 | MUST | YES | M2 |
| AC-SHA-002 | MUST | YES (neutrality CI guard) | M2 / M6 verify |
| AC-SHA-003 | MUST | YES | M3 |
| AC-SHA-004 | MUST | YES (neutrality CI guard) | M3 / M6 verify |
| AC-SHA-005 | MUST | YES | M1 |
| AC-SHA-006a | MUST | YES (content parity) | M2 / M3 |
| AC-SHA-006b | MUST | YES (structural + whole-file byte-parity) | M4 |
| AC-SHA-007 | MUST | YES (mirror parity) | M4 |
| AC-SHA-008 | SHOULD | no | M1 (research.md) |
| AC-SHA-009 | MUST | YES | M5 |
| AC-SHA-010 | MUST | YES | M5 |
| AC-SHA-011 | MUST | YES | M5 |
| AC-SHA-012 | MUST | YES | M5 |
| AC-SHA-013 | SHOULD | no | M5 |
| AC-SHA-014 | MUST | YES (meta-irony) | M3 |
| AC-SHA-015 | SHOULD | no | M5 |
| AC-SHA-016 | SHOULD | no | M4 |

**MUST count**: 13. **SHOULD count**: 4. **Total**: 17 AC rows (16 logical AC IDs — AC-SHA-006a and AC-SHA-006b share the REQ-SHA-006 logical binding but count as 2 gating rows). iter-2 split AC-SHA-006 into 006a + 006b per D5, adding 1 MUST row and 1 total row vs iter-1. (iter-1 footer undercounted its own table by 1 — claimed "MUST count: 11" while listing 12 MUST rows; iter-2 reconciles.)

## §D.2 Traceability

| REQ | AC | Milestone |
|-----|----|-----------|
| REQ-SHA-001 | AC-SHA-001 | M2 |
| REQ-SHA-002 | AC-SHA-002 | M2 |
| REQ-SHA-003 | AC-SHA-003 | M3 |
| REQ-SHA-004 | AC-SHA-004 | M3 |
| REQ-SHA-005 | AC-SHA-005 | M1 |
| REQ-SHA-006 | AC-SHA-006a (content) + AC-SHA-006b (structural) | M2/M3 (006a) + M4 (006b) |
| REQ-SHA-007 | AC-SHA-007 | M4 |
| REQ-SHA-008 | AC-SHA-008 | M1 (research) |
| REQ-SHA-009 | AC-SHA-009 | M5 |
| REQ-SHA-010 | AC-SHA-010 | M5 |
| REQ-SHA-011 | AC-SHA-011 | M5 |
| REQ-SHA-012 | AC-SHA-012 | M5 |
| REQ-SHA-013 | AC-SHA-013 | M5 |
| REQ-SHA-014 | AC-SHA-014 | M3 |
| REQ-SHA-015 | AC-SHA-015 | M5 |
| REQ-SHA-016 | AC-SHA-016 (+ AC-SHA-006b carries the structural-parity half) | M4 |

Every REQ (16) traces to at least one AC. Every AC traces to exactly one REQ (AC-SHA-006a and AC-SHA-006b both trace to REQ-SHA-006, with AC-SHA-006b additionally co-verifying REQ-SHA-016's structural relocation — the split is documented per D5).

## §D.3 Indirect verification (quality gates)

- **Neutrality CI guard**: `template-neutrality-check.yaml` + `internal/template/internal_content_leak_test.go` MUST pass on the final template copy. This indirectly verifies AC-SHA-002 and AC-SHA-004 (no residual internal tokens).
- **Mirror parity test**: `TestRuleTemplateMirrorDrift` MUST pass green on the session-handoff.md subtest post-enrollment. This indirectly verifies AC-SHA-006b (whole-file byte-parity post-M4) and AC-SHA-007 (enrollment is active).
- **Full template suite**: `go test ./internal/template/...` MUST pass — covers both the mirror test and the leak test in one invocation.

## §D.4 Closure gates

The SPEC cannot close until ALL of the following are true:
1. All 13 MUST AC rows pass with verifiable evidence (iter-2: 12 → 13 after splitting AC-SHA-006 into 006a + 006b per D5; iter-1 undercounted its own table by 1).
2. `go test ./internal/template/...` passes GREEN (mirror + leak).
3. `moai spec lint` on this SPEC directory passes clean (no FrontmatterInvalid, no OwnershipTransitionInvalid).
4. progress.md carries §E.2 (run evidence), §E.3 (run audit-ready), §E.4 (sync audit-ready + sync_commit_sha), §E.5 (Mx audit-ready + mx_commit_sha).
5. The close commit subject names the full SPEC-ID: `chore(SPEC-SESSION-HANDOFF-ALIGN-001): Mx-phase audit-ready signal + 4-phase close`.

## §D.5 Forward-looking checks (non-blocking)

- **FL-1**: After this SPEC lands, a follow-up audit SHOULD evaluate whether the remaining 16 in-sync workflow/ files should be bulk-enrolled in the mirror test for defense-in-depth. The research.md §A audit table (iter-2 reconciled: 16 in-sync mirrored files + 1 template-missing lifecycle-sync-gate.md) informs this decision; bulk enrollment is out of scope here.
- **FL-2**: The moai.md §8 self-check list consolidation (Finding #10 from the analysis report) is explicitly deferred — `moai.md` is an output-style file and its self-check is scoped to render-surface localization, not paste-ready budget enforcement.
- **FL-3**: If the V0 section's thematic orphan status (spawn-gating doctrine in a message-format file) becomes a maintenance burden, a future SPEC may extract it to a dedicated runtime-rule file. This is out of scope; the current co-location is doctrinally justified by Block 4 emission.

## §D.6 Edge cases

- **EC-1**: If the neutrality CI guard rejects the ported Diet/V0 content despite M2/M3 stripping, the blocker is a residual token the design.md §B enumeration missed. Remediation: grep the offending token, strip it, re-run. Do NOT add session-handoff.md to the §25 carve-out (EXCL-004).
- **EC-2**: If the mirror test fails post-enrollment due to a Unicode byte difference in the `✂` / `─` characters (e.g., one tree has a combining-character form the other lacks), the remediation is `cp` from template to local (template is authoritative post-neutralization) — never the reverse.
- **EC-3**: If M5's cut-line marker de-duplication accidentally strips the marker from a fenced Example block, the paste-ready format breaks. The AC-SHA-009 "outside fenced example blocks" qualifier guards this; the Examples retain their markers verbatim.

## §D.7 Definition of Done

The SPEC is Done when:
- All gating ACs pass.
- The orchestrator's verification batch (`go test ./internal/template/...`, neutrality guard, `moai spec lint`) is GREEN.
- progress.md §E.1-§E.5 are all populated with commit SHAs.
- The close commit has landed on main with the full-SPEC-ID subject.
- The 105-line net template↔local content delta (111 change-line diff / 117 raw-diff lines) is collapsed to zero on canonical + neutralized-appendix content.
- A user running `moai init` or `moai update` receives the Diet and V0 generic doctrine that was previously trapped local-only.
