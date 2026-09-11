# SPEC Review Report: SPEC-GIT-DELIVERY-PROCEDURE-001 (reduced, version 0.2.0)

Iteration: 1/2 (Tier M ceiling; count restarted by the operator's split decision)
Verdict: **FAIL**
Overall Score: **0.71** (harmonic mean; Tier M PASS threshold 0.80)

Reasoning context ignored per M1 Context Isolation. This is a full audit of the reduced plan, not a regression check of the earlier full-scope reports. The operator decisions (scope split, OD-2 = B, moves to t658/t659) were treated as binding inputs and were not re-opened. Every figure below was measured in this run.

## Tree attribution

| Coordinate | Command | Output |
|---|---|---|
| Root | `git rev-parse --show-toplevel` | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t622` |
| Branch | `git branch --show-current` | `WT-git-procedure-fixes` |
| Log | `git log --format='%h %s' -3` | `24e20ec50 feat(SPEC-GIT-DELIVERY-PROCEDURE-001): reduce to mechanical repairs after split` / `87988e946 … apply operator decisions B1-B3` / `880c0c702 …` |
| Status | `git status --porcelain` | (empty) |
| Change set | `git diff --stat 87988e946 HEAD` | 4 files, all under `.moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/` (385 insertions, 1005 deletions) |
| Scope vs base | `git diff --stat b412f8a33 HEAD -- .claude internal AGENTS.md` | (empty) — every cited base line resolves against the working tree |
| Lifecycle audit | `mcp__moai__spec_audit(project_root=<worktree>, filter_spec=…)` | `{"total_specs":1,"grandfathered":0,"modern_era_clean":1,"drift_findings":[]}` |
| Lint | `moai spec lint .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001` | `0 error(s), 1 warning(s)` — `REQTableRowsRejected … LINE 237 … rejected_rows=14` (body-relative offset; the §G.1 tracking table at spec.md:253). Judging build: `v3.2.0-rc.5 list-974-g84fa4ece4`; `git merge-base --is-ancestor 84fa4ece4 HEAD` → exit 0, so the installed build lags the tree and lint rules added after `84fa4ece4` did not run |

Mutant fixtures were built in the session scratchpad (`/private/tmp/claude-501/…/scratchpad/r-*`) and not exported. Each one is reproducible from the command quoted with it.

## Must-Pass Results

- [PASS] **MP-1 REQ number consistency.**

  ```
  /usr/bin/grep -o -E '^\*\*REQ-GDP-[0-9]{3}\*\*' spec.md > r-reqs.txt
  seq -f '**REQ-GDP-%03g**' 1 24 > r-reqs-expected.txt
  diff r-reqs-expected.txt r-reqs.txt   → diff-exit=0
  uniq -d (sorted)                      → (no duplicates)
  total 24 · pattern-labelled (judged) 10 · MOVED/WITHDRAWN placeholders 14
  ```

  **Judgement on the placeholder approach.** MP-1 as written tests the *numbering*: "sequential (REQ-001 … REQ-N) with no gaps, no duplicates, and consistent zero-padding". Every number 001-024 appears exactly once as a bold `**REQ-GDP-NNN**` entry, in order, zero-padded; that is literally sequential with no gap. A one-line tombstone for a number that moved to another card is not a gap, and renumbering would give one token two meanings between this document and `87988e946` (spec.md §C.2) — the traceability hazard MP-1 exists to prevent. This is not a defect.

- [PASS] **MP-2 GEARS** (requirement layer, spec.md only). All 10 judged entries carry and match a pattern label:
  - Ubiquitous: 001, 002, 004, 013, 014
  - Unwanted: 003, 005, 015, 024
  - Event-driven: 006 ("When sync 단계가 PR 병합 여부를 판단할 때", L117)

  The 14 placeholders are explicitly labelled non-requirements (§B.4 L121 "요구사항으로 판정하지 않는다") and carry no normative text, so the GEARS test does not apply to them. Lint reports 0 errors on them; that lint came from a lagging build (see Tree attribution).

- [PASS] **MP-3 Frontmatter** — spec.md:L2-L13 carry all 12 fields.
  - `version: "0.2.0"` is quoted; `status: draft`; `created` 2026-09-10 and `updated` 2026-09-11 are ISO dates; `priority: P1`; `phase: "v3.2.0 target"`; `lifecycle: spec-anchored`; `tags` is a string.
  - `era` / `tier` / `related_specs` are extras, with precedent in existing SPECs.
- [N/A] **MP-4** — not multi-language tooling.
- [PASS] **MP-5 D7** — references found: SPEC-MERGE-METHOD-CONFIG-001 (completed), SPEC-WORKTREE-BRANCH-GUARD-001 (completed), SPEC-AUTH-001. `ls -d .moai/specs/SPEC-AUTH-001` → `No such file or directory`. That one is SHOULD severity only (see N7): the ID is quoted as the leak-test allowlisted example at spec.md:L78, not referenced as a dependency. No retired, superseded, or archived reference.
- [PASS] **MP-6 D8** — `/usr/bin/grep -rn -e '\[NEEDS CLARIFICATION' -e 'syscall' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/` → `exit=1`.
- [PASS] **MP-7** — same command, no markers; no research.md (Tier M).

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.50 | material ambiguity in a core requirement | REQ-GDP-006 names an opt-in flag no command exposes and leaves `--merge`, `--no-merge`, and non-team modes undefined; two engineers would implement M1 differently (B1) |
| Completeness | 0.75 | one area sparse | all sections present. The four sibling documents that advertise `--merge` as the auto-merge flag are neither in scope nor in §D, and §E lacks the corresponding gap (B1) |
| Testability | 0.75 | detectors narrow but reading steps explicit | AC-GDP-001 and AC-GDP-006 automated checks pass reworded mutants and rely on their reading steps (N1). AC-GDP-025's registry check lacks a runnable positive control (N2). The remaining ACs are RED-now with working controls (table below) |
| Traceability | 1.00 | full | the 10 judged REQs each map to an AC (001-006, 013-015, 024→025); AC-GDP-016 is declared a process check outside the trace (acceptance.md:L41, L45); moved IDs are tracked in §G.1 |

Harmonic mean: 4 / (1/0.5 + 1/0.75 + 1/0.75 + 1/1) = 4 / 5.667 = **0.71**.

## Retained AC verification (RED-now, control, mutants, swept set)

| AC | RED-now on `b412f8a33` for the right reason | Positive control run | Mutant result | Swept set |
|---|---|---|---|---|
| 001 | Base Synchronization section: paragraphs with fetch and rev-list `1`; autofail paragraph printed (the 156 paragraph) | yes | "batch after the checkpoint" → autofail `0`; table form → paras `1`, autofail `0`, listgroup `0`; list form → listgroup non-empty. Only the reading step (THEN 1-2) catches the first two — see N1 | non-empty (1 paragraph) |
| 002 | base block: standalone fetch 1, standalone rev-list at block line 5, joined `0`, rev-list count `1` | yes | iteration-2 `git -C . rev-list` leftover → `rev-list` count `2` → FAIL (caught); earlier comment and reversed-order mutants caught (earlier runs) | non-empty (9-line block) |
| 003 | invariance check; its detector-works control is the Pre-Spawn section diff | yes (post-edit) | — | 328 → sweep-prohibition section |
| 004 | `gh pr merge --squash --delete-branch` count `2`; `merge_method` in delivery.md → `exit=1` | yes | double-space evasion of the literal is caught by AC-005's `[^|]*` form | 2 lines |
| 005 | `gh pr merge[^|]*--squash`: manager-git 32·114, delivery 343·355, toml 26·108; acp and doc-execution `0` | yes | literal-variant mutants from the earlier reports are caught by the widened pattern | 6 lines |
| 006 | delivery Step 3.4 section `52` lines, default hits at section lines 8 and 20; doc-execution subsection hit at line 8; `manager-git.md` mentions `0`/`0`; opt-in sentences `1`/`1` | yes | reworded default ("runs inside a worktree … merge once checks pass" / "Worktree sessions merge automatically") plus a `manager-git.md` mention → default detector `exit=1`, source `1` → automated parts PASS; only the reading step catches it (N1) | non-empty (52 and 9 lines) |
| 013 | invariance; base copy diffs `manager-git`/`acp` exit 0, `delivery` 3 hunks, `doc-execution` `138,143d137` | yes | selector `^(TestSanitizedPairParity\|TestTemplateNoInternalContentLeak)$` — both functions exist; `t.Run(filepath.Base(rel))` names the subtest `agent-common-protocol.md`, so the `-v` trace check is satisfiable. Tolerance `structuralDriftToleranceLines = 4` with reword pairs netting zero means the test alone will not catch a one-line unpropagated change; the `diff` carries parity, as §E.1 already declares | 4 file pairs |
| 014 | toml `gh pr merge <PR> --<merge_method> --delete-branch` count `0` | yes | — | 1 file |
| 015 | SPEC-ID/REQ/date controls ≥ 1; hex tokens on spec.md `49`, `:980ccdc56` hit `1` | yes | fixture incl. `538684c47`, `7374b183e 2213871af 980ccdc56 b412f8a33` on one line, `adjacent_7374b183e`, `x-2213871af` → all caught; digits `647460835`, `1000000` routed to reading; uppercase `7374B183E` not caught (documented limit L420) | added template lines |
| 016 | process check at run time | control command present | — | run-phase commits |
| 025 | Frozen tags: acp `1` (L17) in both copies, others `0`; four clause strings `1` each in both copies; registry "others" regex `0` | **partially** (see N2) | — | acp L/T, registry |

AC-GDP-006 under OD-2 = B: the RED cells, controls, and swept sets are correct. The requirement it verifies is under-specified (B1).

## Findings

### Blocking

**B1. REQ-GDP-006 points the sync workflow at an opt-in flag that no command carries, and leaves the fate of `--merge`, `--no-merge`, and non-team modes undefined — spec.md:L117, plan.md:L66-L68, acceptance.md:L237-L288 — Severity: major — Class: blocking (implementability and internal consistency; not a re-opening of OD-2)**

```
$ /usr/bin/grep -rn -e '--auto-merge' internal/template/templates/.claude internal/template/templates/CLAUDE.md internal/template/templates/AGENTS.md
internal/template/templates/.claude/agents/moai/manager-git.md:148:- Auto-merge: only with the `--auto-merge` flag, per § PR Auto-Merge (Team Mode)
internal/template/templates/.claude/agents/moai/manager-git.md:166:Execute only with `--auto-merge` flag AND all approvals obtained:

$ sed -n '85p;104p' internal/template/templates/.claude/skills/moai/workflows/sync.md
/moai sync [mode] [--pr] [--merge] [--skip-mx]
**Flags**: `--pr` (PR 생성) | `--merge` (deprecated, auto-merge) | `--skip-mx` (MX 검증 스킵)

$ (quality-gates-context.md) 30:  - Flag: --merge
                           101:- --merge: After sync, auto-merge PR and clean up branch. Worktree/branch environment is auto-detected from git context.
$ (moai/SKILL.md) 140: Modes: auto, force, status, project. Flags: --merge, --skip-mx
$ (moai/references/reference.md) 161: - --merge: Auto-merge PR and clean up branch after sync

$ sed -n '336-338 / 346-349' delivery.md
- `is_worktree_context == true` AND `--no-merge` flag NOT set
- OR `--merge` flag explicitly set (deprecated, logged as warning)
- `--no-merge`: Skip auto-merge even in worktree context. PR is created but not merged.
- `--merge`: Deprecated. Logs warning: "The --merge flag is deprecated. Auto-merge is now the default for worktree contexts."
```

REQ-GDP-006 requires delivery.md and doc-execution.md to follow "manager-git.md 의 옵트인(`--auto-merge` 플래그와 전원 승인)을 유일한 기준으로". The evidence shows five gaps:

1. The `/moai sync` surface has no `--auto-merge` flag. Its only merge flag is `--merge`, which delivery.md itself marks deprecated.
2. `manager-git.md`'s opt-in lives in a section titled "PR Auto-Merge (**Team Mode**)", and REQ-GDP-006 does not say what personal or manual mode does.
3. The spec does not say whether the delivery trigger line "OR `--merge` flag explicitly set" survives, becomes the opt-in, or is removed.
4. `--no-merge` has no meaning once nothing merges by default, and the spec does not say whether it is kept.
5. plan.md M1 says only "`--no-merge`·`--merge` 설명과 폐기 경고 문구를 새 기본값에 맞춘다". §C.3's option B row itself records "`--merge` 폐기 경고의 논리가 뒤집힌다" without a requirement resolving it.

Three different, reasonable M1 implementations are possible: (a) un-deprecate `--merge` as the opt-in, (b) keep `--merge` deprecated and add `--auto-merge` to the sync surface, (c) remove both flags. AC-GDP-006 passes all three, because it checks only for absent default phrases and a `manager-git.md` mention. Readings (a) and (b) also change the four sibling documents above, which are neither in scope nor listed in §D.

Required fix:
- Add the post-B flag semantics as one sentence to REQ-GDP-006, or as a new REQ (the judged count is 10, well under 16): what `--merge` and `--no-merge` do and what non-team modes do.
- If that wording is itself a decision the operator has not made, route it through the lead as a sub-decision of OD-2 = B rather than letting run-phase choose.
- Extend AC-GDP-006 with a check for the chosen semantics.
- Either bring `sync.md:104`, `quality-gates-context.md:30/101`, `moai/SKILL.md:140`, `moai/references/reference.md:161` into scope or list them in §D with the reason they stay consistent.
- Add the corresponding gap to §E.1.

### Optional (should-fix)

**N1. AC-GDP-001 and AC-GDP-006 automated checks pass reworded mutants; the reading step is the real judge — acceptance.md:L69-L88, L258-L288 — Severity: minor — Class: optional**

```
AC-001 mutant "Issue these as ONE single-turn multi-Bash batch after the checkpoint: `fetch`, `rev-list …`, …"
  → autofail wc = 0
AC-001 table mutant (| remote | `fetch` | / | divergence | `rev-list …` |)
  → paras 1, autofail 0, listgroup 0
AC-006 mutant ("Merge policy follows manager-git.md." / "When the session runs inside a worktree and the no-merge flag is absent, merge once checks pass" / "Worktree sessions merge automatically.")
  → default-phrase detector exit=1, source count 1
```

Both reading steps are written as binary questions and are recorded to files (`ac001-reading.md`, `ac006-reading.md`), so this is not a missing verification. Two suggestions: state in §D.3 that an AC with a reading step is not PASS until the reading file exists and answers every paragraph or section; and add `(for|in) (a )?worktree` plus `merge (automatically|by default)` to the AC-006 detector.

**N2. AC-GDP-025 registry check has no runnable positive control — acceptance.md:L481-L482 — Severity: minor — Class: optional**

```
$ /usr/bin/grep -c -E 'file: \.claude/rules/moai/core/agent-common-protocol\.md' .claude/rules/moai/core/zone-registry.md   → 13
$ /usr/bin/grep -c -E 'file: \.claude/(agents/moai/manager-git\.md|skills/moai/workflows/sync/(delivery|doc-execution)\.md)' …   → 0
```

The expected value is 0, so a registry format change (for example quoted paths) would pass vacuously. Add the acp-path count, expected 13 at base, as the control. The Frozen-line judge `^[-+][^-+].*\[ZONE:Frozen\]` would also skip a diff line whose content begins with `-` or `+`. The only Frozen tag in scope begins with `[` (acp:17), so this is not exploitable today; note it in §D.2.

**N3. The t658 ordering dependency is recorded only in this SPEC — spec.md §E.2 L232, §G.2 L292, plan.md §B.7 — Severity: minor — Class: optional**

```
$ moai todo list --limit 0 … t658
t658	queued	[t622 분할 · late-branch 재설계] … 의존: amend 적용 함수 구현 카드 선행(B4). 입력: t622 워크트리 SPEC 0.1.3 · .moai/reports/t622/. …
```

The card names the t659 dependency but not "land after t622" or the `manager-git.md:114` `--<merge_method>` line. t658 inherits no AC that guards that line: AC-GDP-005 and AC-GDP-014 stay here. If t658 rewrites the Late-Branch block from its 0.1.3 input first, nothing in t658's criteria stops `--squash` from returning. The in-SPEC note is sufficient for this SPEC's own correctness, but the carrier that will actually be read at t658 dispatch is the card. Required action (lead, not manager-spec): append the ordering dependency to t658's card text, or give t658 an AC that `gh pr merge[^|]*--squash` in manager-git.md stays at the single default sentence.

**N4. AC-GDP-014 THEN claims the regenerated `.toml` carries the same synchronization sentence, but no command checks it — acceptance.md:L352 vs L364 — Severity: minor — Class: optional.** Add a fixed-string count of the new line-156 sentence in the toml (expect 1), or drop the clause.

**N5. AC-GDP-001's toml check is prose only — acceptance.md:L88 — Severity: minor — Class: optional.** Give it the same `sed`/`awk` command pair against `$T/.codex/agents/moai/manager-git.toml`. Its `## Synchronization` (148) and `## PR Auto-Merge (Team Mode)` (158) markers exist, so the range extraction works.

**N6. §C.3 option C wording — spec.md:L184 — Severity: note.** "폐기 예정 키를 되살림" is an evaluative verb in a table the SPEC calls "결정 당시 자료"; "폐기 예정으로 분류된 키에 첫 소비자를 만든다" states the same fact neutrally. The OD-2 table is otherwise accurate: option A/B/C file lists, the B-row admission that the `--merge` warning's logic inverts, and C's `class: D` / `deprecate_after: "v3.1.0"` all match the files. The decision record ("C first, then B after the key's classification was presented") is consistent with the earlier iteration record.

**N7. D7 SHOULD — SPEC-AUTH-001 not found — spec.md:L78.** Illustrative ID quoted from the leak-test allowlist, not a dependency; no action required beyond awareness.

**N8. Lint advisory — spec.md:L253 (§G.1 table).** `REQTableRowsRejected` is expected for a tracking table that deliberately carries REQ IDs without modality; it is a warning under a lagging build (see Tree attribution).

## Scope leakage check

- **No retained REQ, AC, milestone, or risk depends on moved work,** except the deliberate `manager-git.md:114` substitution inside the Late-Branch block (REQ-GDP-005, AC-GDP-005/014, plan M2, §B.7, §G anti-pattern, §E.2).
- **The M3 edit at 156 is separable from the moved 160.** The Synchronization section edit (156) sits next to the moved `git pull` line (160); plan §D lists 160 as untouched and AC-GDP-001's extraction does not depend on 160.
- **The M1 Step 3.4 edit leaves the moved line alone.** It touches 335-338/343/348-349/355 and leaves 356 ("Checkout target branch") to t658; AC-GDP-006's section extraction does not require 356 to change.
- **Nothing needed for the retained work was moved out, with one exception:** the `--merge`/`--no-merge` semantics (B1) were never assigned to either card.
- **Ordering note:** sufficient inside this SPEC, insufficient on the carrier t658 will be dispatched from (N3).

## Tier M justification

`spec-workflow.md:140-150`: Tier S is `< 5 files`, REQ ceiling 8 and AC ceiling 8; Tier M is `5 - 15 files`, ceilings 16/16. This SPEC has 10 judged REQs (> 8), 11 judged ACs (> 8), and 9 affected files (≥ 5). Only the LOC guidance points to S, and it is explicitly "guidance, not enforcement". **Tier M holds.** Excluding placeholders from the ceiling count is correct: they carry no obligations.

## §E gaps and risks (retained scope)

- **E.1:** the gaps are scoped to retained work and are true (TestSanitizedPairParity tolerance confirmed in code: `structuralDriftToleranceLines = 4`, reword pairs net zero). Missing: the flag-semantics gap from B1.
- **E.2:** the risks are scoped (t658 overlap, the OD-2 = B user-visible default change, line-level detectors, base-identity premise). The t658 overlap risk is real; its mitigation belongs on the t658 card (N3).

## Recommendation

1. **B1** — specify, in REQ-GDP-006 or one new REQ, what `--merge`, `--no-merge`, and non-team modes do under OD-2 = B.
   - If that choice was not part of the operator's B decision, surface it through the lead first.
   - Add the matching AC-GDP-006 check.
   - Place the four sibling flag documents in scope or in §D, and add the §E.1 gap.
2. **Optional, in the same pass:** N1 (§D.3 reading-file gate + two detector alternations), N2 (registry positive control), N4, N5.
3. **Lead-side:** N3 — record "land after t622; keep `manager-git.md:114` on `--<merge_method>`" on card t658.

The rest of the plan is sound. Numbering (MP-1), GEARS, frontmatter, Frozen non-touch design, parity/test selection, `.toml` regeneration, the SHA detector (now word-level and verified against late-letter SHAs), tier, and traceability all check out.

## Gaps (not observed)

- No `go test`, `make agents-emit`, or `make agents-emit-check` was run (dispatch prohibition). AC-GDP-013's `-v` subtest name and top-level PASS format are inferred from `t.Run(filepath.Base(rel))` in `sanitized_pair_parity_test.go` and Go's standard output, not observed.
- The `moai spec lint` result came from build `84fa4ece4`, an ancestor of HEAD; rules landed after that build were not exercised.
- Whether the operator's OD-2 = B decision already implied a specific `--merge` semantics was not observable from the tree; only the SPEC text and the card queue were read.
- No cross-model audit (no `audit_model` key in the project config; checked in the first full-scope audit, not re-run).

## Residual risk

- Even with B1 fixed, AC-GDP-001 and AC-GDP-006 conclude on human reading; a misreading remains possible.
- OD-2 = B changes a user-visible default. Users relying on worktree auto-merge will see unmerged PRs until the flag semantics are documented where they look (the sibling flag documents in B1).
- Base line citations rely on the scope files staying byte-identical to `b412f8a33`. Absorbing develop before run-phase requires re-running `git diff --stat $BASE HEAD -- <scope>` (plan §C step 1 already requires this).
