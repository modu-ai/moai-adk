# SPEC-MEMORY-STORE-RECONCILE-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

**Iteration 3 (post plan-audit FAIL 0.78, up from 0.72).** N1-N8 addressed; optional N9-N11, N13
applied; N12 (AC ceiling) resolved without a merge. Per-defect mapping in the closing report.

- **Tier: M**, re-checked rather than inherited. Non-artifact files touched: **8** — four sources
  (`moai-memory.md`, `moai-constitution.md`, `moai.md`, `token_budget_guard.go`), **three** mirrors
  (`internal/config/` has none), and one test file. Inside Tier M's 5-15 band. The earlier
  "four sources plus four mirrors" phrasing was wrong and is corrected in plan.md §B.4, which
  carries the enumerated table this line must stay consistent with.
- Requirements: **15** (ceiling 16). Acceptance criteria: **16** (ceiling 16, zero headroom).
  N2 / N3 / N5 / N6 / N11 were absorbed as **strengthenings of existing criteria**, so no new slot
  was needed and no merge was forced. Any further criterion now requires a scope split or a tier
  change — stated so the ceiling is not quietly exceeded later.
- Evidence base: `.moai/reports/t383/measurements.md` §M5a (revision 2 — records **both** wrong
  relations) and §M5b, plus `.moai/reports/t383/measure-n1.sh`, an independent re-measurement run
  in this worktree at iteration 3. No figure in the SPEC is re-derived; each is cited, expressed in
  R4 re-measurement form, or explicitly withdrawn as unattributed.
- **A live-mutation observation that decides which figures may be pinned.** The index moved three
  times on 2026-08-31 — 26,280 → 26,290 → **26,577** bytes, 123 → **124** entry lines, 189 → **190**
  unique targets — with nobody here touching it. Across the same interval the defect figures held
  exactly (58 / 44 / 14 / 40). Size metrics are dated references; defect metrics may be asserted.
- SPEC ID regex check executed (Bash), verbatim output: `PASS`.
- Open decisions: none. G5 (the copy set) is decided in spec.md §A.2.2 — **all 58**.
- No `[NEEDS CLARIFICATION]` markers.
- **Two gaps recorded, not closed** (plan.md §I): G2 — `moai spec lint` is unmeasured for this SPEC
  at plan time (installed binary 20 commits behind for `internal/spec/`; the PATH run's 0 findings
  is recorded as weak evidence from a stale build, and AC-MSR-016 requires `./bin/moai` by path).
  G4 — doctrine-vs-code divergence on store derivation, now enumerated across **three** surfaces
  including `moai.md:165`, the line M1 itself edits.

## §E.2 Run-phase Evidence

Full evidence with verbatim output: `.moai/reports/t383/verdict.md`. Store readings:
`reconcile-before.json` / `reconcile-after.json`. Sampling gate: `m0-sample.md`.

`$D` = the active store resolved by `moai memory doctor` (path in `preflight.md`).

### AC matrix — 16 of 16 PASS

| AC | Status | Actual output |
|---|---|---|
| AC-MSR-001 | PASS | `0` / `0` / `0` (baselines at HEAD: `3` / `1` / `2`) |
| AC-MSR-002 | PASS | cond.1 `4` / `4` in local+mirror; cond.2 no numeric cap value in either copy |
| AC-MSR-003 | PASS | `131:#### Compressing the index means making entries shorter — never fewer` |
| AC-MSR-004 | PASS | 4 matches; § "Two stores, and only one of them is loaded" carries the `--dir` + `exists` rule |
| AC-MSR-005 | PASS | `head -1` → `# MoAI Constitution` (always-loaded); clause offers no drop branch |
| AC-MSR-006 | PASS | first grep exit `1`; `MEMORY.md` only at lines 109/114/127, all inside the TOMBSTONE |
| AC-MSR-007 | PASS | mutation → `--- FAIL: TestFixedSlotsExistInRepoTree`; revert → `--- PASS` |
| AC-MSR-008 | PASS | `18d17 < …/MEMORY.md`; same-tree A/B total `76009` both ways → slot contributed `0` |
| AC-MSR-009 | PASS | gate exit `0`; BEFORE `58`; AFTER `0` |
| AC-MSR-010 | PASS | entry `139 → 139`; targets `205 → 205`; bytes `31366 → 31366` |
| AC-MSR-011 | PASS | legacy `1098 → 1098`; recent-mtime `0 → 0`; `skipped_exists: 0` |
| AC-MSR-012 | PASS | `git status --short` shows no path under either store |
| AC-MSR-013 | PASS | `rc1=0 rc2=0 rc3=0` |
| AC-MSR-014 | PASS | existence gate ok ×3; scan `exit=1` exactly; planted red `exit=0` |
| AC-MSR-015 | PASS | `m0-sample.md`: 0 of 12 superseded vs threshold ≥4 → PROCEED |
| AC-MSR-016 | PASS | anchor `0`; `lint_exit=0`; liveness `0`/`0`; `0 error(s), 1096 warning(s)`; count `0` |

### Invariants

| Invariant | Status | Evidence |
|---|---|---|
| No store file is a commit target (REQ-MSR-015) | HOLDS | `git status --short` — no store path, staged per-file by explicit pathspec |
| Copy-only, never-overwrite (REQ-MSR-010) | HOLDS | `cp -n`; log `copied 58 / skipped_exists 0 / no_source 0`; legacy count and mtimes unmoved |
| Neither reachability metric decreased (REQ-MSR-011) | HOLDS | `139 → 139`, `205 → 205` |
| Index not edited by this card (REQ-MSR-012/013) | HOLDS | bytes `31366 → 31366` |
| Mirror parity (REQ-MSR-014) | HOLDS | three `diff -q` rc=0 after `make build` |
| Defect figures stable across 5 readings | HOLDS | dangling `58` at every pre-M3 reading while every size figure moved |

### Debt items discharged

| Debt | How |
|---|---|
| 1 — stale `123` denominator | Removed from spec.md:92 (the [HARD] sentence), :102, :306, :312 **before** M1 wrote doctrine; cascade-corrected REQ-MSR-011 and AC-MSR-010's stale `124/135/190`. Verified: **zero** stale figures reached any of the six doctrine files. Vindicated in-flight — the denominator moved 123 → 139 during this card |
| 2 — AC-MSR-016 misfires | Anchor + liveness assertion added; both misfires reproduced first (`6` from the SPEC dir; `0` from an empty file). The first anchor form was itself un-runnable and was corrected (see below) |
| 3 — prefix pathspec sweeping runtime state | AC-MSR-012 narrowed to the four named artifacts; staged per-file; `git status --short` re-read in the staging call |
| 4 — commit `measure-n1.sh` | Committed with `derive-missing.sh` and `reconcile.sh`; two caveats recorded (needs bash, word-splits on filenames) |
| 5 — re-run what you rewrote | Rule written into plan.md §G with the base-rate table, now **five for five** including this round's two |
| 6 / G8 — orphan count | Recorded before **and** after: `46 → 46`, unmoved as predicted but now measured rather than assumed |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-01
run_commit_sha: <backfill — commits created at close of this run-phase>
run_status: complete
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: performed (origin/develop 26 ahead; NOT absorbed per lead instruction)
l44_post_push_fetch: n/a — push is out of scope for this run-phase
new_warnings_or_lints_introduced: 0
  spec_lint: "0 error(s), 1096 warning(s)" via ./bin/moai by path; 0 findings name this SPEC
  go_vet: clean
  gofmt: clean
cross_platform_build:
  darwin_arm64: pass (make build)
  other: not attempted locally — CI owns the matrix
total_run_phase_files: 8 tracked modified + 2 untracked trees (SPEC artifacts, evidence reports)
m1_to_mN_commit_strategy: three commits (doctrine+mirrors / guard removal / SPEC+evidence)
budget_change:
  always_loaded_token_budget: 76000 -> 76210
  reason: REQ-MSR-004 clause must be always-loaded (C5); headroom was 201 tokens
  measured_before: 75799
  measured_after: 76009
  note: raised by exactly this card's 210-token addition so prior headroom is preserved, not inflated
gaps_open: [G4, G6, G7]
gaps_closed: [G2]
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-01
sync_commit_sha: pending-backfill-sync
sync_commit_sha_note: |
    A commit cannot cite its own hash, and this card's dispatch bounds the sync
    phase to a SINGLE commit (no push, no PR, no merge, no integration lock), so
    no backfill commit is created here. The value is reported to the lead in the
    sync-phase completion report instead. Run-phase commits, which §E.3 left as
    `<backfill>` and which are NOT edited here (§E.2/§E.3 are manager-develop's):
    36aa5bf4c (guard removal + doctrine) and 4064c7857 (run-phase close).
sync_status: complete
changelog_entry_position: "CHANGELOG.md [Unreleased] > ### Added, first entry (inserted above SPEC-EVIDENCE-CITATION-CANON-001)"
b12_self_test_a: "pre-emission grep — `grep -c 'SPEC-MEMORY-STORE-RECONCILE-001' CHANGELOG.md` = 0 before writing (no duplicate from a parallel session)"
b12_self_test_b: "AC count — `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l` = 16, non-zero and matching §E.2's 16-row PASS matrix; the CHANGELOG entry states 16/16"
b12_self_test_c: "file-path verification — every path named in the CHANGELOG entry resolves in this tree (`ls` batch, all present); `TestFixedSlotsExistInRepoTree` confirmed present at token_budget_guard_test.go:357; `moai mx query --kind DEBT` confirmed as a real CLI contract (internal/cli/mx_query_debt_test.go)"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (single sync commit, 3-phase close)"
  plan_md: "none — no `status:` field (ArtifactStatusFieldForbidden, card t357)"
  acceptance_md: "none — no `status:` field"
  progress_md: "none — no frontmatter"
  updated_field: "spec.md `updated:` already 2026-09-01 (sync-commit date); unchanged, not re-stamped"
mx_tag_changes:
  added_count: 1
  detail: |
    `@MX:DEBT` + `@MX:CEILING` + `@MX:UPGRADE` + `@MX:SPEC` on
    `AlwaysLoadedTokenBudget` (internal/config/token_budget_guard.go). Judged
    warranted: the raise is a deliberate, working simplification standing in for
    the large always-loaded rule diet, correct within a named limit (0.26%
    headroom) and carrying a known revisit trigger — the exact shape
    mx-tag-protocol.md § When to Add Tags gives for DEBT. The marker also gives
    G7 a machine-harvestable home (`moai mx query --kind DEBT`) rather than a
    prose-only warning a future author must happen to read.
  not_added: |
    No tag on the head-cap TOMBSTONE. The @MX taxonomy annotates live
    constructs (a constant, a function, an API boundary); the tombstone annotates
    an ABSENCE, and there is no construct for a tag to hang on. Its re-addition
    hazard is already mechanically guarded by TestFixedSlotsExistInRepoTree, which
    is stronger than a comment marker. No @MX:WARN either: the removal eliminated
    danger rather than introducing it, and WARN's stated triggers (goroutines,
    complexity >= 15, global mutation, branch count) are all absent.
canary_compliance_check: n/a — this SPEC defines no forward-looking policy that its own sync tests
sync_phase_files_touched_count: 4
sync_phase_files_touched:
  changelog: "CHANGELOG.md — [Unreleased] > Added entry (repo convention: one prose entry per sync-phase close)"
  spec_frontmatter: ".moai/specs/SPEC-MEMORY-STORE-RECONCILE-001/spec.md — `status:` only, zero body edits"
  progress: ".moai/specs/SPEC-MEMORY-STORE-RECONCILE-001/progress.md — this §E.4 block only"
  mx: "internal/config/token_budget_guard.go — @MX:DEBT marker only, no behavior change"
docs_surfaces_judged_unwarranted:
  - "README (4 locales) — carries no auto-memory / index-budget claim: `grep -c 'MEMORY\\.md' README{,.ko,.ja,.zh}.md` = 0 on all four, `grep -c '25KB'` = 0 on all four. Nothing to correct."
  - "`.moai/project/{product,structure,tech}.md` — `grep -rn 'MEMORY.md' .moai/project/*.md` = no matches; no store-derivation or index-budget assertion"
docs_surface_finding_not_fixed:
  where: "docs-site/content/{en,ko,ja,zh}/claude-code/context-memory/{memory,context-window}.md"
  measured: |
    8 files, 16 lines assert the SAME unconfirmed cut this SPEC withdrew from the
    in-repo surfaces — "first 200 lines or 25KB, whichever comes first" and its
    ko/ja/zh renderings, including a mermaid node and a paragraph stating the limit
    applies to MEMORY.md only. Command:
    `grep -rn '200 lines or 25KB|앞 200줄|冒頭 200 行|前 200 行|first 200 lines' docs-site/content | wc -l` -> 16;
    `grep -rl ... | sort` -> the 8 files above.
  why_not_fixed: |
    Same reasoning as G6 and one step stronger. (1) It is outside spec.md §A.5's
    enumerated surface, so editing it is silent scope expansion — the exact move
    the run phase declined. (2) This SPEC does not ASSERT the opposite claim; it
    records the cut as UNCONFIRMED, so replacing 16 user-facing lines would trade
    an unverified claim for another unverified claim on the surface with the widest
    audience. (3) docs-site is a 4-locale surface under a same-PR i18n obligation
    and a Vercel production binding; correcting it is a card, not a sync-phase
    side-effect. Reported to the lead instead.
  discovered_by: "sync-phase docs-surface sweep — NOT named by the run phase, which enumerated only the skills-reference pair (G6)"
sync_verification:
  go_test: |
    `go test ./internal/config/...` after the @MX edit ->
    ok internal/config 3.549s / ok internal/config/atomicfile 0.719s /
    ok internal/config/toolpolicy (cached). The edit is comment-only.
  go_vet: "`go vet ./internal/config/...` -> exit 0, no output"
  gofmt: |
    SCOPED, not a whole-package pass. `gofmt -l internal/config/token_budget_guard.go`
    -> empty (the file this sync-phase edited is clean). `gofmt -l internal/config/`
    lists SEVEN pre-existing unformatted files — audit_struct_yaml_symmetry_test.go,
    model_routing_test.go, profile_test.go, slice.go, template_removed_key_test.go,
    toolpolicy/codegen.go, toolpolicy/settings_region.go — none of which this card
    touches (none appear in `git diff --name-only 297a21ea7..HEAD`). Recorded rather
    than fixed: reformatting them is a drive-by outside this card's scope. An
    unqualified "gofmt clean" here would have been false.
  mirror_parity: "re-verified this phase, not inherited: `diff -q` x3 on moai-memory.md / moai-constitution.md / moai.md against their internal/template/templates/ mirrors -> rc1=0 rc2=0 rc3=0. Sync-phase touched none of the six."
  spec_lint: |
    Targeted: `./bin/moai spec lint <this spec.md>` (tree-built binary, not the
    PATH one) -> exit 0, "No findings". Whole-repo: `./bin/moai spec lint` ->
    "0 error(s), 1096 warning(s)", exit 0 — byte-identical to §E.3's run-phase
    figure, so the status transition introduced no finding (delta 0).
  store_files_committed: 0
  evidence_exported:
    - ".moai/reports/t383/sync-spec-lint.txt (verbatim targeted-lint output)"
    - ".moai/reports/t383/sync-verification.txt (every sync-phase command with its observed output)"
gaps_open: [G4, G6, G7, M0-sampling, G8-docs-site]
gaps_detail:
  G4: |
    Doctrine-vs-code store-derivation divergence across three surfaces; neither
    side was asserted, deliberately. A follow-up card should be scoped to the
    DERIVATION, not to `memory doctor` alone — the same `os.Getwd()` path drives
    `spec lint` too, so fixing one consumer leaves the other on the old rule.
  G6: |
    `claude-code-memory-official.md:125` still asserts the unconfirmed cut ("the
    first 200 lines ... or the first 25KB, whichever comes first"). Sync-phase
    measurement REFINES the run-phase statement of this gap: it is a PAIR, not one
    file — `.claude/skills/moai-foundation-cc/reference/` carries it as well as the
    `internal/template/templates/` mirror the run phase named, both at line 125.
    Deliberately NOT edited: it sits outside spec.md §A.5's enumerated surface, and
    editing it here would be silent scope expansion. Named as known residue; a
    follow-up must move BOTH copies together (Template-First parity).
  G7: |
    The always-loaded surface is saturated at 0.26% headroom (201 tokens of
    76,210). The next card that grows an always-loaded file hits this guard. The
    @MX:DEBT marker above is where a future author meets this fact; the root fix is
    the large always-loaded rule diet, out of this card's scope.
  G8-docs-site: |
    NEW, found in sync-phase: `docs-site/content/{en,ko,ja,zh}/claude-code/
    context-memory/{memory,context-window}.md` — 8 files, 16 lines — assert the
    same unconfirmed cut on the user-facing surface. Deliberately not fixed; see
    `docs_surface_finding_not_fixed` above for the measurement and the three
    reasons. Belongs in the same follow-up as G6, which should then cover THREE
    surfaces: the local skill reference, its template mirror, and docs-site x4
    locales.
  M0-sampling: |
    Sampling reached 12 of 58 (21%), so a superseded file among the 46 unsampled
    copied targets is possible — reversible by deletion, since the copy was
    `cp -n` and left the legacy store untouched. Concretely: sample entries 16 and
    26 give OPPOSITE instructions about running the full test suite locally,
    neither marked superseded, so BOTH were copied. This card's copy step is what
    makes that contradiction reachable. Recorded as a finding, plausibly its own
    follow-up card; not resolved here, because resolving it means adjudicating two
    lessons' content, which is neither sync-phase work nor this card's scope.
operator_decision_recorded:
  what: "AlwaysLoadedTokenBudget 76,000 -> 76,210"
  shape_disclosed: "conflict of interest — the card that trips a guard is the card that raises it"
  operator_ruling: keep
  rationale: |
    The clause was cut by roughly 1,000 bytes FIRST and the constant raised by
    only the residual 210, so prior headroom (201 tokens) is preserved rather than
    inflated. The raise is documented in the constant's own comment and in
    `.moai/reports/t383/verdict.md`, where a reviewer can object to it. Recorded,
    not re-litigated.
```
