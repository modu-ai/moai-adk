# SPEC-TEMPLATE-DATE-NEUTRALITY-002 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-27
plan_iteration: 3
spec_id: SPEC-TEMPLATE-DATE-NEUTRALITY-002
tier: L
baseline_head: 760f09f73
iteration_1_commit: f5d3a93bf
iteration_2_commit: e1f24264d
branch: spec/template-date-2025
```

### Artifacts

| Artifact | Status |
|---|---|
| `spec.md` | authored — 24 requirements, 9 out-of-scope sub-sections |
| `plan.md` | authored — 6 milestones, 10 known issues, PRESERVE list |
| `acceptance.md` | authored — 33 criteria + §G known limitations (6 items), all baselines executed |
| `design.md` | authored — coupled-ordering, masking hazard, taxonomy rationale |
| `research.md` | authored — full measurement record, 1 refuted hypothesis, §K partition derivation |
| `classify.sh` | committed — year-widened classifier, reproduces 74 / 48 / 34 |
| `progress.md` | this file |

Tier L artifact set (5 files) complete; `classify.sh` is the committed measurement instrument (REQ-TDN2-008) and `progress.md` is emitted at every Tier — neither counts toward the Tier artifact total.

### Counts

| Metric | Value |
|---|---:|
| Requirements (`REQ-TDN2-001` … `REQ-TDN2-024`) | 24 |
| Acceptance criteria (`AC-TDN2-001` … `AC-TDN2-033`) | 33 |
| Occurrence-class rows in scope | 74 |
| Distinct findings | 48 |
| Distinct files carrying a finding | 34 |
| REMOVE rows (`DC-2a`) | 28 across 22 files |
| PRESERVE rows (`DC-2b`) | 13 |
| Per-row adjudicated rows (`DC-5`) | 33 |
| Dual-category findings | 4 |
| Open questions carried to M2 | 3 |
| Deferred / blocking questions | 0 |

### Measurement provenance

Every count traces to a command recorded in `research.md`, executed from the worktree root at `760f09f73`. The measurement instrument — a year-widened replica of the predecessor's committed classifier — was validated by reproducing the predecessor's own post-remediation residual of 88 `202[6-9]` rows (`carved out = 100 − k`, `k = 12`).

Two of the task brief's stated figures were re-verified and confirmed exactly (74 occurrences / 48 findings / 34 files; 10 frontmatter-shaped `updated:` lines). One stated hypothesis was **refuted**: the 10 `updated:` lines do not belong to the frontmatter category and are not carved by its structural gate, because all 10 sit at column 0 while the gate requires indentation. See `research.md` §D.

### Open questions (non-blocking, resolved at M2)

| # | Question | Rows | Recorded default |
|---|---|---:|---|
| 1 | Are version-history records factual records worth keeping, or internal history that should not ship? The predecessor's "Known cosmetic residue" note (its `spec.md` §5) is the constraining precedent and is now cited. | 14 | none — per-row; splits into a 2-row bullet form carrying the precedent and a 12-row table form that does not |
| 2 | Does a `Created:` stamp follow the prose-stamp REMOVE rule, or is it distinguishable? | 3 | none — per-row; either outcome is covered (dual-category finding, routed through AC-016) |
| 3 | **Re-posed as an editing instruction**: what is the residual line after the mid-line stamp is excised from each of the 2 named `COMPOSITE` rows? | 2 | candidate residual `Version: 5.0.0 \| Enterprise Ready:` — M2 confirms or replaces |

None blocks plan-phase completion; each has its measurement attached in `research.md` §J and an owning milestone in `plan.md` M2.

Question 3's earlier phrasing ("does removal constitute a placeholder substitution?") was mis-posed — REQ-TDN2-009 already answers it. The real gap was that the two rows were never located, so the shape of the edit could not be reviewed. Both are now named with file and line in `research.md` §J item 3 and §K.

---

## §E.1.1 Iteration-2 audit resolution

Plan-audit iteration 1 returned **FAIL 0.67** against the Tier L 0.85 threshold. All seven must-pass criteria passed and all 17 executable baselines reproduced exactly; the failure was confined to the acceptance and traceability layers. Resolution:

| Finding | Severity | Resolution |
|---|---|---|
| D1 — AC-016 unsatisfiable under a permitted M2 outcome | critical | Target recomputed from the M2 adjudication (`actual == expect` per file). AC-012's adjudication branch preserved. |
| D2 — AC-011 half-dead pattern, no control | major | Replaced with a single-stage unanchored command (catches all 3 shapes incl. mid-line). **Baseline corrected `0` → `9`** after measurement found 9 pre-existing ISO-format doc lines. AC-008 named as control. |
| D3 — AC-012 line-number dependency | major | Re-anchored on the `fenced` rationale marker now required by REQ-TDN2-012. AC-031 added to enforce the §A.4 claim mechanically. |
| D4 — no known-limitations disclosure; edit-scope criterion dropped | major | §G added with 5 disclosed gaps; AC-029 restores the predecessor's edit-scope criterion with `catalog.yaml` pre-excluded. |
| D5 — AC-001 traced to wrong REQ | minor | REQ column `022` → `008`; command re-anchored on the `DATE_RE` line (target `1`, was mis-stated). |
| D6 — 4 requirements uncovered | major | AC-029 (REQ-023), AC-030 (REQ-017), AC-031 (REQ-020), AC-032 (REQ-024) added; REQ-022 disclosed as unverifiable in §G item 1. |
| D7 — HIST question missing its precedent | major | Predecessor's residue note cited in `research.md` §J item 1 and `plan.md` M2. |
| D8 — "path suffix" wrong | minor | Corrected to exact relative path (`entry.File == relPath`) in `plan.md` M4 and `design.md` §C. |
| D9 — fourth year-bearing site omitted | minor | `research.md` §F now tables 4 sites; REQ-TDN2-024 brings the cross-SPEC doc comment into scope; AC-032 verifies the owning test stays green. |
| D10 — three `Where` used as data conditions | minor | REQ-003 / 012 / 015 converted to `When`. |
| Unverified 1-2 — sub-shape partition, COMPOSITE rows | — | `research.md` §K enumerates all 33 `DC-5` rows per code; both COMPOSITE rows named with file, line, and exact text. |
| Unverified 3 — classifier not committed | — | `classify.sh` committed; re-run reproduces 74 rows / 48 findings / 34 files, DC-2a 28 / DC-2b 13 / DC-5 33. |
| Unverified 4 — phantom lint flag | — | No artifact cites `moai spec lint --path`; verified by grep. |

**Two defects were self-inflicted during the repair and caught before commit**: AC-001's target was `1` but the committed `classify.sh` contains the token twice (functional line + header comment), so the command was re-anchored on `^DATE_RE=`; and AC-031's whole-file grep was tripped by this SPEC's own explanatory prose quoting the offending shape, so the prose was reworded and the false-positive mode disclosed.

---

## §E.1.2 Iteration-3 audit resolution

Plan-audit iteration 2 returned **FAIL 0.83** against the 0.85 threshold — a 0.67 → 0.83 progression with no regression. All seven must-pass criteria passed; 31 of 32 criteria were verified sound by execution, both iteration-1 MUST-FIX repairs held across 4 and 3 scenarios, and `classify.sh` cross-checked identical against an independently-written classifier. One critical defect drove the shortfall.

| Finding | Severity | Resolution |
|---|---|---|
| D-NEW-1 — AC-030 vacuously prints `ORDERED` | critical | Both `git log` ranges bounded to `$(git merge-base origin/main HEAD)..HEAD`. Reproduced the broken form (`ORDERED` with zero remediation commits) and the repaired form (`NOT-EVALUABLE`) before publishing; both transcripts recorded in `acceptance.md` and `research.md` §I. §G item 2 rewritten to name the vacuity as the primary limitation, with the squash-merge window demoted to secondary. |
| D-NEW-2 — AC-012's `/fenced/` also matches `unfenced` | minor | Marker pinned to the **uppercase** literal `FENCED` in REQ-TDN2-012; AC-012 matches on non-alphabetic boundaries. Reproduced the failure (`2 2` vs target `1 1`) and verified the fix returns `1 1` against a fixture carrying both `unfenced` and `UNFENCED`. |
| D-NEW-3 — AC-011 lacks AC-007/008's bold groups | minor | **Fixed rather than disclosed.** `(\*\*)?` groups added; fixture detection went 3/4 → 4/4 and the live baseline is unchanged at `9` (zero bolded stamps in this tree), so the change is free. |
| D-NEW-4 — AC-029 blind to an out-of-scope edit inside the template tree | minor | **Closed rather than disclosed.** New **AC-033** requires every edited template file to carry a triage row, with the triage file-count (34) as its paired control. Narrowing AC-029's exclude to the 22 REMOVE-bearing files was rejected — that list is not final until M2. The residual (line-level containment) is disclosed as §G item 6. |
| D-NEW-5 — allowlist-key rejection under-argued | minor | `design.md` §C now leads with the decisive argument: narrowing the key would force the Go guard to classify line shapes, duplicating `classify.sh`'s rules in a second implementation that can drift. The two weak grounds are recorded as explicitly not relied on, and "achieves the same assurance" is downgraded to "sufficient detection for this SPEC's two measured cases". |
| Clarity residual — REQ-003 default vs REQ-011 absolute | minor | REQ-003's table gained an "Adjudicable at M2?" column: `EX-FM` + `EX-DATA` (13 rows) are **pinned** by REQ-011 and not adjudicable; the other four codes (20 rows) carry a starting position. This is also what makes the two `DC-2b`+`DC-5` dual-category findings benign, now stated explicitly. |
| Unverified — 14 `HIST` line numbers | — | `research.md` §K gained a one-command regeneration of the `HIST` set plus the `HIST + COMPOSITE = 16` subtraction an auditor can check without per-row reading. |

REQ 24 (unchanged), AC 32 → **33**.

---

## §E.2 Run-phase Evidence

Run-phase executed as 6 milestone commits (M1 `64850b0bc` → M6 `f0c71168c`) on branch `spec/template-date-2025` in worktree `.claude/worktrees/date2025`. All 33 acceptance criteria PASS. Evidence persisted to `.moai/state/verify/tdn2/`.

### AC PASS/FAIL matrix (33/33 PASS)

| AC | Command (abbreviated) | Observed | Status |
|---|---|---|---|
| AC-001 | `grep -c "^DATE_RE='202\[5-9\]" classify.sh` | `1` | PASS |
| AC-002 | row/col check on triage.tsv | `74 0` | PASS |
| AC-003 | distinct findings + category dist | `48`; `DC-2a 28 / DC-2b 13 / DC-5 33` | PASS |
| AC-004 | DC-5 rows lacking sub-shape code | `0` | PASS |
| AC-005 | rows lacking REMOVE/PRESERVE | `0` | PASS |
| AC-006 | dual-category findings | `4` | PASS |
| AC-007 | REMOVE-scoped prose stamps (non-mirror) | `0` | PASS |
| AC-008 | PRESERVE mirror stamps | `13` | PASS |
| AC-009 | frontmatter example `updated:` values | `10` | PASS |
| AC-010 | EX-DATA schema values | `3` | PASS |
| AC-011 | placeholder tokens (YYYY-MM-DD etc.) | `9` | PASS |
| AC-012 | fenced DC-2a row (n, ok) | `1 1` | PASS |
| AC-013 | predecessor `202[6-9]` set (invariant) | `89` | PASS |
| AC-014 | guard 2025 entries == triage PRESERVE findings | `24 24` | PASS |
| AC-015 | LineStart/End/No fields (invariant) | `28` | PASS |
| AC-016 | dual-category files: actual == expect | `0==0` both files | PASS |
| AC-017 | `202[5-9]` occurrences in guard | `2` | PASS |
| AC-018 | `202[6-9]` occurrences in guard | `1` | PASS |
| AC-019 | `archive-202[6-9]` (invariant) | `1` | PASS |
| AC-020 | `20[0-9]{2}` attribution (invariant) | `1` | PASS |
| AC-021 | strict tier with widened class | `--- PASS (0.51s)`, exit 0 | PASS |
| AC-022 | quoting-agnostic leak-guard steps | `2` | PASS |
| AC-023 | unquoted `-run` invocations | `0` | PASS |
| AC-024 | quoted `-run` invocations | `3` | PASS |
| AC-025 | `MOAI_TEMPLATE_LEAK_STRICT` in workflow | `1` | PASS |
| AC-026 | narrow tier (env unset) | `--- PASS (0.41s)`, exit 0 | PASS |
| AC-027 | neutrality audit | `--- PASS`, exit 0 | PASS |
| AC-028 | `go build ./...` | exit 0 | PASS |
| AC-029 | edit-scope pathspec diff | `0` | PASS |
| AC-030 | remediate-before-widen ordering | `ORDERED` | PASS |
| AC-031 | line-number awk assertion / AC control | `0 / 33` | PASS |
| AC-032 | `TestLeakClassNoDateShaInDefaultTier` | `--- PASS (0.00s)`, exit 0 | PASS |
| AC-033 | edited files not in triage / file count | `0 / 34` | PASS |

### REQ-TDN2-022 build-step exit code (time-critical debt)

```
$ make build
make-build-exit=0
  catalog.yaml updated successfully (11569 bytes)
  go build -ldflags "..." -o bin/moai ./cmd/moai
```

`make build` ran twice during M3 (after the initial 27 DC-2a deletions, and again after the fenced-row correction); both exit 0. Per acceptance.md §G item 1, REQ-TDN2-022 is structurally unverifiable by a test (`//go:embed` re-embeds on every `go build`); the exit code is recorded here as run-phase evidence.

### AC-030 ordering — both transcripts (time-critical debt)

**Repaired (bounded) form** — the live criterion:
```
BASE=3e6c92ef7 (merge-base origin/main HEAD)
WIDEN=27aef363e (M4 — first commit to introduce literal 202[5-9] into the guard)
REMED=2fb84ce4b (M3 — remediation)
git merge-base --is-ancestor 2fb84ce4b 27aef363e → true
RESULT: ORDERED
```

**Broken (unbounded) form** — falsification reproducing the false pass (acceptance.md §G item 2):
```
WIDEN_BROKEN=f0c71168c (HEAD)
REMED_BROKEN=ccd6be1f6 (2026-02-03, "feat(templates): add embedded template system" — predates this branch by 5 months)
git merge-base --is-ancestor ccd6be1f6 f0c71168c → true
RESULT: ORDERED  ← FALSE PASS (18 history matches; REMED_BROKEN never empty, just wrong)
```

### Per-file removal verification — dual-category files (B2 masking hazard)

| File | actual prose stamps | expect (PRESERVE LS-PROSE-STAMP) | Status |
|---|---|---|---|
| `moai-workflow-spec/references/examples.md` | 0 | 0 | PASS |
| `moai-workflow-project/references/examples.md` | 0 | 0 | PASS |

The allowlist entries mask the deleted DC-2a rows, but AC-016 confirms deletions via direct grep independent of the guard.

### Rows whose adjudication differs from the sub-shape default

| Row(s) | Sub-shape default | Actual | Reason |
|---|---|---|---|
| Fenced DC-2a (1 row) | DC-2a default REMOVE | REMOVE (re-adjudicated at M3; initially PRESERVE at M2) | AC-007 requires zero non-mirror prose stamps; no structural carve-out for fenced stamps exists. REQ-TDN2-012 requires explicit adjudication — REMOVE is valid. |
| HIST bullet (2 rows) | per-row | REMOVE | Predecessor (SPEC-001 spec.md §5) removed the 2026 halves of this same list; completing the remediation resolves the flagged residue. |
| HIST table (12 rows) | per-row | PRESERVE | Legitimate release-note documentation; table form does not permit clean inline excision. |
| CREATED (3 rows) | per-row | PRESERVE | Documentation-example Created: stamps. |
| COMPOSITE (2 rows) | per-row | REMOVE | Mid-line excision; residual `Version: 5.0.0 \| Enterprise Ready:` (no placeholder, REQ-TDN2-009). |

Final tally: **32 REMOVE** (28 DC-2a + 2 HIST bullet + 2 COMPOSITE), **42 PRESERVE** (13 DC-2b + 10 EX-FM + 3 EX-DATA + 12 HIST table + 3 CREATED + 1 DEADLINE).

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-28
run_commit_sha: f0c71168c
run_status: audit-ready
ac_pass_count: 33
ac_fail_count: 0
preserve_list_post_run_count: 6
  # PRESERVE list (plan.md §A.2) verified intact at run close:
  # 1. SPEC-TEMPLATE-DATE-NEUTRALITY-001 artifacts — untouched (read-only reference)
  # 2. predecessor date allowlist entries — additive-only (24 new, 0 rewritten)
  # 3. narrow-tier class set + neutrality-audit test — untouched (AC-026/027 green)
  # 4. 202[6-9] date literals — AC-013=89 invariant held
  # 5. catalog.yaml — build-regenerated only (not a remediation target)
  # 6. main checkout — untouched (all work confined to .claude/worktrees/date2025)
l44_pre_commit_fetch: "0 5 (5 commits ahead of origin/main at pre-flight; not pushed)"
l44_post_push_fetch: "N/A — did not push (PR-mandatory repo; push is orchestrator/manager-git decision)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  go_build_darwin_arm64: exit 0
  note: "No syscall/windows-specific code changed; cross-platform N/A for this SPEC scope"
total_run_phase_files: 31
m1_to_mN_commit_strategy: "one commit per milestone (M1-M6), explicit-pathspec staging, no --amend, no force-push, no --no-verify; progress.md §E.2/§E.3 populated in a follow-up evidence commit"
```

---

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-07-28
spec_id: SPEC-TEMPLATE-DATE-NEUTRALITY-002
sync_commit_sha: pending-backfill
close_type: 3-phase close (plan→run→sync)
ac_total: 33
ac_pass: 33
ac_fail: 0
run_commit_sha: f0c71168c
preserve_verified: true
```

The single sync commit carries the `in-progress → implemented → completed` frontmatter transition (3-phase close; MX Tag is a cross-cutting sync concern, not a separate phase). `sync_commit_sha` is `pending-backfill` — a commit cannot reference its own SHA (D3 self-referential-hazard backfill exemption, per spec-frontmatter-schema.md); the orchestrator backfills the real SHA after the PR's squash merge lands.

Sync produced via orchestrator-direct recovery after manager-docs spawn failed with `Prompt is too long` (PTL) — rung-1 in-turn self-correction per runtime-recovery-doctrine §2; no work was lost (the spawn terminated before any file edit).

---

## §F Phase 4 Mode Selection

### Input parameters

- **tier**: L
- **scope (file count)**: ≥24 (22 REMOVE-bearing template files + 1 Go guard + 1 CI workflow, before any DC-5 adjudicated REMOVE adds more)
- **domain count**: 4 (template tree, Go internal-content-leak guard, CI workflow, SPEC triage artifacts)
- **file language mix**: markdown-heavy (template tree) + Go (guard test) + YAML (CI workflow) + TSV (triage)
- **concurrency benefit**: LOW — coupled-ordering hazard (design.md §A); M5 widening is gated on M3+M4 verification, so milestones are strictly sequential, not independent
- **Agent Teams prereqs**: n/a (static layer retired)

### Mode evaluation

| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | 24+ files, semantic changes — not trivial |
| 2 background | no | write-capable implementation, not read-only async |
| 3 agent-team | no | RETIRED (static layer tombstone) |
| 4 parallel | no | LOW concurrency benefit; coupled-ordering makes milestones dependent, not independent; coding-heavy (Anthropic coding-task parallelism caveat) |
| 5 sub-agent | **yes** | sequential per-milestone delegation matches the M1→M6 dependency chain; default fallback for coding-heavy Tier L |
| 6 workflow | no | not a single uniform mechanical transform — M2 adjudication is judgment work, M3/M4 are content-anchored edits, M5 is a coupled one-line flip |

### Decision

Mode: sub-agent

### Justification

The milestone chain has a hard ordering constraint (plan.md §F, design.md §A): the year-class widening in M5 turns CI red on 48 findings unless every REMOVE row (M3) and every PRESERVE carve-out (M4) has already landed. That dependency is strictly sequential — no two milestones are safely parallelizable. Combined with the judgment-heavy M2 adjudication (3 open questions on HIST/CREATED/COMPOSITE rows) and the content-anchored (not line-numbered) edits in M3/M4, this is the canonical Mode 5 case per Anthropic's coding-task parallelism caveat. Mode 6 is rejected because the transform is not uniform-mechanical: each REMOVE row carries its own disposition and each PRESERVE row its own literal-date anchor.

### Kickoff confirmation

Implementation Kickoff Approval: PASSED (prior session 3053725a, plan-audit PASS 0.90, score ≥ 0.90 with 4-condition skip-eligibility satisfied for Phase 1 verdict re-execution — distinct from this kickoff gate). All preferences collected at kickoff; no mid-run user input required.
