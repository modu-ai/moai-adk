# SPEC-SYNC-GATE-VERDICT-001 — Implementation Plan

frontmatter-less sibling artifact (per schema SSOT § Artifact Statelessness).
SPEC: `.moai/specs/SPEC-SYNC-GATE-VERDICT-001/spec.md` · card t783 · branch `WT-syncgate-hook` · base `a404132e7`

## §A Context

- Defect surface (all paths work-tree-relative):
  - Hook pair (byte-identical, no source-neutrality constraint but no NEW internal card IDs on
    edited lines): `.claude/hooks/moai/sync-phase-quality-gate.sh` and
    `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh`.
  - Doc pair (diverged):
    `.claude/skills/moai/workflows/sync/quality-gates-quality.md` (local, 27018 bytes) and
    `internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md`
    (template, 22595 bytes).
- Provenance: card t783 (issued 2026-09-10, audit baseline `main` `2213871af`); the H01
  machinery was already delivered by SPEC-SYNC-GATE-FAILSTATE-001 (card t624, completed) and
  subsequent hook commits (t603/t604/t663/t664) touched the same file after that close. The
  line numbers in the audit record are main-based — locate every target by symbol/phrase.
- Plan-audit provenance: PASS 0.88 (threshold 0.80), 6 findings, report
  `.moai/reports/t783/plan-audit.md`; v0.2.0 applies F1-F6. F1 (High) chose repair option A:
  the Phase 8 relationship paragraph's two stale clauses are aligned OUT of both copies
  (measured byte-identical in both at template line 148 / local line 164), because the
  clauses must actually leave both copies — a stale-clause grep guarding a surviving false
  sentence is the vacuous-green shape (verification-completeness §2) — and because the local
  copy already carries the contradiction today, so option A repairs a live defect where
  option B would enshrine it. The amendment changes the artifact hash, so the run-phase
  Plan Audit Gate re-executes (skip-eligibility invalidated by design).
- Key measured deltas on `a404132e7` (verbatim phrase counts):
  - template doc: "Only CRITICAL findings block" = 1; "Continue with warning" = 1;
    "Continue by approved exception" = 0; "sync-auditor FAIL" = 1 (t624 canonical sentence,
    carrying the two stale clauses); "security-decision-contract" = 0.
  - local doc: mirrored counts (0 / 0 / 1 / 1 / 2 — the local-only rule is referenced twice).
  - `security-decision-contract.md` exists at `.claude/rules/moai/core/` and is ABSENT from
    `internal/template/templates/.claude/rules/` — the template doc's integrated text must
    state the contract inline.
- Development mode: verification-and-docs SPEC; TDD/DDD selection does not apply to the doc
  edit; the control-arm harness in M1/M2 is the RED/GREEN instrument for the H01 surface.

## §B Known Issues

| # | Issue | Evidence |
|---|-------|----------|
| B1 | Template doc copy internally contradicts itself: old severity trio (lines 136-138) + old CRITICAL-only gate (150-156) sit under the t624 canonical sentence (line 148) | phrase counts §A; template doc read |
| B2 | The unified decision text lives ONLY in the local copy (later local-only edit, Template-First violated) | local doc lines 154/166-177 |
| B3 | The model rule `.claude/rules/moai/core/security-decision-contract.md` is not template-mirrored — a verbatim cp of the local text into the template would create a dangling reference (the same doc-promises-more-than-exists class H03 names) | `ls` both trees |
| B4 | Hook pair byte-identical today (`diff -q` exit 0); any M2 repair must preserve that byte-identity | measured this run |
| B5 | H01 execution proof does not exist against the current tree; hook file changed after FAILSTATE-001's close (t603/t604/t663/t664) | git log on the hook path |
| B6 | The Stop-hook block path requires a sync-phase commit subject (`docs(...): sync-phase` family), a detected language, and a non-zero code-file delta — the harness fixture must satisfy all three gates or the arms measure nothing | hook lines 208-253 |
| B7 | `worktree_content_id` excludes `.moai/state` and `.moai/logs` — the harness's own state writes must not invalidate the same-HEAD arms; conversely a worktree edit invalidates them (that IS arm B's mechanism) | hook lines 371-384 |
| B8 | `moai update` deletes and redeploys `.claude/skills/moai*`; until the template copy is corrected, the local copy's unified text is one update away from being reverted — this is why the template edit lands first | CLAUDE.local.md §2.3 |
| B9 | BOTH copies carry the Phase 8 relationship paragraph's two stale clauses ("its CRITICAL-only stop gate below"; "a HIGH finding that Phase 8 reports only as a warning") — false once the unified Critical/High gate lands, and already false in the local copy today | template line 148 / local line 164, byte-identical |
| B10 | The `2213871af` baseline hook has a distinct gate layout (single sentinel file, SHA recorded before checks) — the fixture recipe tuned to the current hook's gates may need adaptation for the positive control | baseline hook extract; M1 self-catches the difference and records the adaptation as an evidence row |

## §C Pre-flight (run-phase entry)

- [ ] Re-locate all four target files by symbol/phrase on the then-current tree; if any cited
      phrase is already gone, record that fact as a finding in `.moai/reports/t783/` before
      proceeding (baseline discipline; audit line numbers are main-based and not load-bearing).
- [ ] Confirm work-tree identity: branch `WT-syncgate-hook`, base `a404132e7` reachable
      (`git merge-base --is-ancestor a404132e7 HEAD`), `git status --short` clean of foreign
      entries before the first write.
- [ ] Extract the audit-baseline artifacts read-only:
      `git show 2213871af:.claude/hooks/moai/sync-phase-quality-gate.sh > <report>/baseline-hook.sh`
      and the baseline doc copy — these are the positive-control inputs, never edited.
- [ ] Confirm the harness fixture recipe satisfies B6/B7: temp git repo under `/tmp`
      (auto-cleaned), `go.mod` + one Go file with a deliberate fast-check failure, HEAD commit
      subject in the sync-phase family, `CLAUDE_PROJECT_DIR` pointed at the fixture, blocking
      mode default (no `MOAI_SYNC_GATE_BLOCKING`), stdin `echo '{}'`. The baseline hook's
      gate layout differs (B10): if the current-hook recipe does not drive the baseline copy
      to its checks, adapt the fixture for the positive control and record the adaptation as
      an M1 evidence row — the control observes the defect shape, not a shared recipe.
- [ ] Confirm the hook copies are byte-identical at pre-flight (`diff -q`, exit 0) — the
      pairing precondition for any M2 repair.
- [ ] Confirm `security-decision-contract.md` is still local-only (B3) — if a parallel card
      has mirrored it meanwhile, the template text MAY then reference it; otherwise inline.

## §C.2 Design decisions (committed direction)

**(i) Severity unification = decision block AND relationship-paragraph clauses, both copies
(option A).** The template doc's Step 0.55.1/0.55.2 replacement text states the verdict
mapping verbatim (Critical and High block; Medium and Low advisory; exception-record fields)
with the contract INLINE (B3). The Phase 8 relationship paragraph keeps its canonical claim
(Step 0.5.4 rubric canonical; Phase 8 additional lens; its stop gate never clears an earlier
sync-auditor FAIL) but LOSES the two clauses the unified gate falsifies — "its CRITICAL-only
stop gate below" and "a HIGH finding that Phase 8 reports only as a warning" (B9). The
alignment lands in BOTH copies, template first; exact replacement wording for the
never-clears clause is the run phase's, subject to the AC-SGV-009 greps. Option B (keep the
freeze, grep-watch the false clauses) was rejected: it would guard a known-false sentence
surviving — the vacuous-green shape — and would leave the local copy's live contradiction in
place. "Continue with warning" becomes "Continue by approved exception".

**(ii) Local copy: body change limited to the two-clause alignment.** Its decision block
already carries the unified text (verify, align only on drift), but its relationship
paragraph carries the same two stale clauses as the template (B9) — that paragraph is inside
the scoped surface (AC-SGV-007(b)) and is aligned in M3. The one deliberate pair delta stays
documented (local names the local-only rule path; template states inline) — REQ-SGV-007.

**(iii) H01 arms measure behavior, never source.** The three arms run the deployed hook
script against `/tmp` fixtures; source reading may explain a failure but never substitutes
for the executed evidence. Positive control first: the `2213871af` hook copy must reproduce
the defect (call1 block JSON, call2 empty stdout) before any current-tree arm result is
interpreted (B10 governs fixture adaptation for the older gate layout).

**(iv) Conditional repair fence.** All-arms-green ⇒ zero hook code change, proof is the
deliverable (REQ-SGV-003). Any red arm ⇒ minimal repair in BOTH hook copies, template first,
same commit, byte-identity re-proven, arms re-run. No drive-by hook cleanup.

**(v) Write-ordering provenance.** The payload-before-fail-record ordering is
FAILSTATE-001's verified surface (its torn-write shims and mutants M23/M24); this SPEC
consumes it (REQ-SGV-002) and verifies only the post-hoc state shape (AC-SGV-003), so no
mutant writing the record first can pass HERE unobserved-and-unowned — it is owned upstream.

## §D Constraints

- No `make build` (lead builds at batch end); no settings.json edits; no commandemit surface
  (verified unaffected).
- Template DOC-copy edited passages carry NO internal SPEC IDs, no audit citations, no
  internal dates, no commit SHAs, no non-shipped rule paths (template neutrality). The HOOK
  copies are exempt from source neutrality but carry a tighter rule: NO NEW internal card IDs
  on edited/added lines — pre-existing citations (t604/t663/t664) on untouched lines are
  preserved verbatim (plan-audit F5).
- Stage by explicit pathspec; commit subject per the ownership matrix
  (`feat(SPEC-SYNC-GATE-VERDICT-001): ...` for plan; run-phase commits per convention).
- Verdict-bearing commands: file redirect + exit code, no pipes (card measurement contract).
- Do not touch the fenced doc passages (§B of spec.md); `.moai/reports/t783/` files stay
  UNTRACKED in the primary checkout.
- Existing gates preserved verbatim (REQ-SGV-005): the `deps_modified` block, the
  per-language fast-check `case`, and every early-exit path.

## §E Self-Verification

Run-phase evidence lands in `progress.md` §E.2/§E.3 (manager-develop) plus untracked
`.moai/reports/t783/` artifacts. Plan-phase commitment: every AC in acceptance.md names its
verification command form and expected output shape before implementation starts.

## §F Milestones

| ID | Priority | Milestone | Deliverable |
|----|----------|-----------|-------------|
| M1 | High | Reproduction + positive control | Baseline hook + baseline doc extracted from `2213871af` (read-only); positive-control two-call sequence against the baseline hook copy reproduces the H01 defect shape (call1 block, call2 empty; fixture adapted to the older gate layout per B10, adaptation recorded); pattern-proof per family (F7): the H03 absence pattern proves green against the `2213871af` baseline doc (measured 1 hit there), while the stale-clause removal patterns prove against the CURRENT pre-M3 tree (1 hit each in both current copies — the `2213871af` baseline doc has 0, the clauses postdate it via t624's M3); the `2213871af` baseline positive control otherwise applies to the H01 hook-state arms only; fixture recipe frozen under `.moai/reports/t783/` |
| M2 | High | H01 three-arm execution on the current tree | Arms A/B/C executed against the deployed hook via `/tmp` fixtures with file-redirect + exit-code evidence: (A) same-HEAD failing check twice → call2 re-delivers the block byte-identically, no check re-run; (B) worktree input change under unchanged HEAD → checks re-execute; (C) pass record → silent empty stdout. Conditional minimal repair per REQ-SGV-003 if any arm is red (both copies, template first, byte-identity re-proven, arms re-run); §E.2 evidence written |
| M3 | High | SX-R05 template-first doc alignment + H03 verification | Template doc Step 0.55.1/0.55.2 replaced with the unified inline severity contract (design §C.2(i)); old trio + "Continue with warning" removed; Phase 8 relationship paragraph's two stale clauses aligned out of BOTH copies (option A, B9) while the canonical-direction claim and never-clears semantics are retained; H03 wording verified present in both copies + hook comment; local copy verified aligned on the decision block (edit only on drift) and aligned on the relationship paragraph; neutrality scan of the template edits |
| M4 | Medium | Source-axis verification + close | Pairing evidence (hook pair byte-identical full-file; doc scoped-passage parity via the AC-SGV-007(b) mechanical proxy with the one documented delta), scope-fence diff audit (no gate deletion, no scanner, fenced passages untouched, no NEW card IDs on hook edited lines), all evidence under `.moai/reports/t783/` (untracked, primary checkout), progress.md §E.2/§E.3 populated |

## §G Anti-Patterns

- Do NOT verbatim-cp the local doc over the template doc (or vice versa) — the deliberate
  integration is the deliverable; a whole-file cp would (a) drag in the fenced divergence
  passages and (b) reference a non-shipped rule path.
- Do NOT add the security-decision-contract path reference to the template copy while the
  rule is local-only (B3) — that re-creates the H03 defect class in new form.
- Do NOT freeze a sentence whose neighbors you falsify: after unifying the gate, the
  relationship paragraph's "CRITICAL-only" and "HIGH-as-warning" clauses are false text, and
  a grep that watches them survive is a vacuous green, not a guard (F1 option-A rationale).
- Do NOT accept arm evidence from source reading, a stale run, or another tree; each arm is
  an executed command on the current tree in this run.
- Do NOT run the positive control against a re-written "old hook" — extract it from git
  (`git show 2213871af:...`), read-only, and adapt the fixture to ITS gate layout (B10).
- Do NOT delete or weaken any gate check while repairing (REQ-SGV-005); a repair that passes
  arms by loosening a check is a FAIL.
- Do NOT leave the doc pair "fixed" only in the local copy (B8) — template first, always.

## §H Cross-References

- Owning predecessor: SPEC-SYNC-GATE-FAILSTATE-001 (card t624, completed) — outcome-record
  machinery, payload/retry design, write-ordering shims (consumed by REQ-SGV-002), AC-011
  (H03), AC-012 (SX-R05 partial, regression-guard preserved the CRITICAL-only gate).
- Card: t783 (H01 · H03 · SX-R05; audit baseline `main` `2213871af`).
- Plan-audit: `.moai/reports/t783/plan-audit.md` (PASS 0.88; F1-F6 applied in v0.2.0).
- Evidence: `.moai/reports/t783/` (untracked, primary checkout).
- Governing rules: verification-claim-integrity.md (§1.1 surface 3 — the develop-side
  repair-status claims here are execution-verified, not text-inferred);
  verification-completeness.md §2 (the option-B vacuous-green rejection);
  CLAUDE.local.md §2.3 (update deletes `.claude/skills/moai*` — why the template edit leads).
