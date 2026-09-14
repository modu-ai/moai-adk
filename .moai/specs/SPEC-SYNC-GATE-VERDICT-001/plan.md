# SPEC-SYNC-GATE-VERDICT-001 — Implementation Plan

frontmatter-less sibling artifact (per schema SSOT § Artifact Statelessness).
SPEC: `.moai/specs/SPEC-SYNC-GATE-VERDICT-001/spec.md` · card t783 · branch `WT-syncgate-hook` · base `a404132e7`

## §A Context

- Defect surface (all paths work-tree-relative):
  - Hook pair (byte-identical, no neutrality constraint):
    `.claude/hooks/moai/sync-phase-quality-gate.sh` and
    `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh`.
  - Doc pair (diverged):
    `.claude/skills/moai/workflows/sync/quality-gates-quality.md` (local, 27018 bytes) and
    `internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md`
    (template, 22595 bytes).
- Provenance: card t783 (issued 2026-09-10, audit baseline `main` `2213871af`); the H01
  machinery was already delivered by SPEC-SYNC-GATE-FAILSTATE-001 (card t624, completed) and
  subsequent hook commits (t603/t604/t663/t664) touched the same file after that close. The
  line numbers in the audit record are main-based — locate every target by symbol/phrase.
- Key measured deltas on `a404132e7` (this run, verbatim phrase counts):
  - template doc: "Only CRITICAL findings block" = 1; "Continue with warning" = 1;
    "Continue by approved exception" = 0; "sync-auditor FAIL" = 1 (t624 canonical sentence);
    "security-decision-contract" = 0.
  - local doc: all mirrored (0 / 0 / 1 / 1 / 2 — the local-only rule is referenced twice).
  - `security-decision-contract.md` exists at `.claude/rules/moai/core/` and is ABSENT from
    `internal/template/templates/.claude/rules/` — the template doc's integrated text must
    state the contract inline.
- Development mode: verification-and-docs SPEC; TDD/DDD selection does not apply to the doc
  edit; the control-arm harness in M1/M2 is the RED/GREEN instrument for the H01 surface.

## §B Known Issues

| # | Issue | Evidence |
|---|-------|----------|
| B1 | Template doc copy internally contradicts itself: old severity trio (lines 136-138) + old CRITICAL-only gate (150-156) sit under the t624 canonical sentence (line 148) | phrase counts §A; template doc read |
| B2 | The unified decision text lives ONLY in the local copy (later local-only edit, Template-First violated) | local doc lines 154/164/166-177 |
| B3 | The model rule `.claude/rules/moai/core/security-decision-contract.md` is not template-mirrored — a verbatim cp of the local text into the template would create a dangling reference (the same doc-promises-more-than-exists class H03 names) | `ls` both trees |
| B4 | Hook pair byte-identical today (`diff -q` exit 0); any M2 repair must preserve that byte-identity | measured this run |
| B5 | H01 execution proof does not exist against the current tree; hook file changed after FAILSTATE-001's close (t603/t604/t663/t664) | git log on the hook path |
| B6 | The Stop-hook block path requires a sync-phase commit subject (`docs(...): sync-phase` family), a detected language, and a non-zero code-file delta — the harness fixture must satisfy all three gates or the arms measure nothing | hook lines 208-253 |
| B7 | `worktree_content_id` excludes `.moai/state` and `.moai/logs` — the harness's own state writes must not invalidate the same-HEAD arms; conversely a worktree edit invalidates them (that IS arm B's mechanism) | hook lines 371-384 |
| B8 | `moai update` deletes and redeploys `.claude/skills/moai*`; until the template copy is corrected, the local copy's unified text is one update away from being reverted — this is why the template edit lands first | CLAUDE.local.md §2.3 |

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
      mode default (no `MOAI_SYNC_GATE_BLOCKING`), stdin `echo '{}'`.
- [ ] Confirm the hook copies are byte-identical at pre-flight (`diff -q`, exit 0) — the
      pairing precondition for any M2 repair.
- [ ] Confirm `security-decision-contract.md` is still local-only (B3) — if a parallel card
      has mirrored it meanwhile, the template text MAY then reference it; otherwise inline.

## §C.2 Design decisions (committed direction)

**(i) Template severity text = inline contract, not a path reference.** The template doc's
Step 0.55.1/0.55.2 replacement text states the verdict mapping verbatim (Critical and High
block; Medium and Low advisory; exception-record fields) and keeps the existing t624
canonical sentence ("sync-auditor rubric is canonical; Phase 8's gate never clears an earlier
sync-auditor FAIL") and the line-71 HARD THRESHOLD untouched. Model text: the local copy's
lines 154/164/166-177, with the `security-decision-contract.md` path reference replaced by
the inline statement (B3). "Continue with warning" becomes "Continue by approved exception".

**(ii) Local copy: no body change required on the scoped passages.** It already carries the
unified text. The plan still VERIFIES it (AC-SGV-006/007) and aligns it only if pre-flight
finds drift. The one deliberate pair delta is documented (local names the local-only rule
path; template states inline) — REQ-SGV-007.

**(iii) H01 arms measure behavior, never source.** The three arms run the deployed hook
script against `/tmp` fixtures; source reading may explain a failure but never substitutes
for the executed evidence. Positive control first: the `2213871af` hook copy must reproduce
the defect (call1 block JSON, call2 empty stdout) before any current-tree arm result is
interpreted.

**(iv) Conditional repair fence.** All-arms-green ⇒ zero hook code change, proof is the
deliverable (REQ-SGV-003). Any red arm ⇒ minimal repair in BOTH hook copies, template first,
same commit, byte-identity re-proven, arms re-run. No drive-by hook cleanup.

## §D Constraints

- No `make build` (lead builds at batch end); no settings.json edits; no commandemit surface
  (verified unaffected).
- Template-copy edited passages carry NO internal SPEC IDs, no audit citations, no internal
  dates, no commit SHAs, no non-shipped rule paths (template neutrality; the hook file is
  exempt).
- Stage by explicit pathspec; commit subject per the ownership matrix
  (`feat(SPEC-SYNC-GATE-VERDICT-001): ...` for plan; run-phase commits per convention).
- Verdict-bearing commands: file redirect + exit code, no pipes (card measurement contract).
- Do not touch the fenced doc passages (§B of spec.md); do not edit anything under
  `.moai/reports/t783/` expectations — those files stay UNTRACKED in the primary checkout.
- Existing gates preserved verbatim (REQ-SGV-005): the `deps_modified` block, the
  per-language fast-check `case`, and every early-exit path.

## §E Self-Verification

Run-phase evidence lands in `progress.md` §E.2/§E.3 (manager-develop) plus untracked
`.moai/reports/t783/` artifacts. Plan-phase commitment: every AC in acceptance.md names its
verification command form and expected output shape before implementation starts.

## §F Milestones

| ID | Priority | Milestone | Deliverable |
|----|----------|-----------|-------------|
| M1 | High | Reproduction + positive control | Baseline hook + baseline doc extracted from `2213871af` (read-only); positive-control two-call sequence against the baseline hook copy reproduces the H01 defect shape (call1 block, call2 empty); baseline-doc phrase greps prove the grep patterns catch the existing target (H03/SX-R05 patterns green on the baseline before any absence is claimed); fixture recipe frozen under `.moai/reports/t783/` |
| M2 | High | H01 three-arm execution on the current tree | Arms A/B/C executed against the deployed hook via `/tmp` fixtures with file-redirect + exit-code evidence: (A) same-HEAD failing check twice → call2 re-delivers the block byte-identically, no check re-run; (B) worktree input change under unchanged HEAD → checks re-execute; (C) pass record → silent empty stdout. Conditional minimal repair per REQ-SGV-003 if any arm is red (both copies, template first, byte-identity re-proven, arms re-run); §E.2 evidence written |
| M3 | High | SX-R05 template-first doc alignment + H03 verification | Template doc Step 0.55.1/0.55.2 replaced with the unified inline severity contract (design §C.2(i)), old trio + "Continue with warning" removed, canonical sentence + HARD THRESHOLD preserved; H03 wording verified present in both copies + hook comment; local copy verified aligned on scoped passages (edit only on drift); neutrality scan of the template edits |
| M4 | Medium | Source-axis verification + close | Pairing evidence (hook pair byte-identical full-file; doc scoped-passage semantic parity with the one documented delta), scope-fence diff audit (no gate deletion, no scanner, fenced passages untouched), all evidence under `.moai/reports/t783/` (untracked, primary checkout), progress.md §E.2/§E.3 populated |

## §G Anti-Patterns

- Do NOT verbatim-cp the local doc over the template doc (or vice versa) — the deliberate
  integration is the deliverable; a whole-file cp would (a) drag in the fenced divergence
  passages and (b) reference a non-shipped rule path.
- Do NOT add the security-decision-contract path reference to the template copy while the
  rule is local-only (B3) — that re-creates the H03 defect class in new form.
- Do NOT accept arm evidence from source reading, a stale run, or another tree; each arm is
  an executed command on the current tree in this run.
- Do NOT run the positive control against a re-written "old hook" — extract it from git
  (`git show 2213871af:...`), read-only.
- Do NOT delete or weaken any gate check while repairing (REQ-SGV-005); a repair that passes
  arms by loosening a check is a FAIL.
- Do NOT leave the doc pair "fixed" only in the local copy (B8) — template first, always.

## §H Cross-References

- Owning predecessor: SPEC-SYNC-GATE-FAILSTATE-001 (card t624, completed) — outcome-record
  machinery, payload/retry design, AC-011 (H03), AC-012 (SX-R05 partial, regression-guard
  preserved the CRITICAL-only gate).
- Card: t783 (H01 · H03 · SX-R05; audit baseline `main` `2213871af`).
- Evidence: `.moai/reports/t783/` (untracked, primary checkout).
- Governing rules: verification-claim-integrity.md (§1.1 surface 3 — the develop-side
  repair-status claims here are execution-verified, not text-inferred);
  CLAUDE.local.md §2.3 (update deletes `.claude/skills/moai*` — why the template edit leads).
