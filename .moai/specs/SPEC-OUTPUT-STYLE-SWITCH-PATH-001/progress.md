# Progress — SPEC-OUTPUT-STYLE-SWITCH-PATH-001

Card t906. Tier M. Plan-phase base: `9dcbc3dbe` (branch `WT-output-style-switch-path`).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-18
tier: M
artifacts: spec.md, plan.md, acceptance.md (+ this progress.md)
requirements: 11 (REQ-OSP-001..011) — Tier M ceiling 16
acceptance_criteria: 14 (AC-OSP-001..011 + AC-OSP-GATE-001..003) — Tier M ceiling 16
open_clarifications: 0
clarification_resolved: `plan.md` §B.2 — whether `moai.md` gains switch guidance it does not currently carry. Ruled by the lead 2026-09-18, before Implementation Kickoff Approval: **option (가) — four files changed; both `moai.md` copies stay byte-unchanged** (asserted by AC-OSP-006). Deciding predicate: a file is not opened to fix a defect it does not have — the same predicate that cut docs-site out of scope, and `moai.md` carries zero switch-guidance sites, so adding guidance would be a new surface rather than a correction. The sibling-asymmetry finding (of the three personas, only `moai.md` has no exit route) was NOT dismissed — it is split to card **t936**, carrying this card's measurements as evidence (sites 4 / 1 / 0; `moai-easy` / `moai-learn` grep 0 hits inside `moai.md`; contrasted against the `moai-easy.md` §14 switch table and the `moai-learn.md` line 35 pointer). t936 carries a [HARD] precondition of its own: decide first whether the silence is a defect or a deliberate design choice, with the three files' `git log` as evidence — if deliberate, recording that fact is t936's deliverable. Scope impact on this SPEC: none; the ruling confirmed the authored four-file scope rather than re-scoping it.
scope_note: docs-site excluded by lead decision and recorded as a proposed follow-up card. The embed refresh (`make build` + `moai doctor --check "Agent Emit Embed"`) is owed by the change but assigned to the batch-close step, per the lead's ruling; no acceptance criterion depends on a built binary.

## §E.2 Run-phase Evidence

Run-phase base: `9dcbc3dbe`. Every figure below was measured in this worktree, in
this run. The pre-flight baselines (`plan.md` §C) were captured **before** the
first edit, in their own step; the after-values were captured after the five
edits and the two mirror copies.

### E.2.0 Pre-flight baselines (measured before any edit)

| Measurement | Command | Observed |
|---|---|---|
| anchored `/output-style`, working copies | `grep -cE '/output-style([^-a-zA-Z]\|$)'` per file | `moai.md 0`, `moai-easy.md 0`, `moai-learn.md 0` |
| anchored `/output-style`, template mirrors | same, against the mirror tree | `moai.md 0`, `moai-easy.md 0`, `moai-learn.md 0` |
| `/config` per file | `grep -c '/config'` | `7`, `7`, `7` |
| `language.yaml` per file | `grep -c '\.moai/config/sections/language\.yaml'` | `moai.md 7`, `moai-easy.md 3`, `moai-learn.md 6` |
| pair identity | `cmp` per pair | `identical` × 3 |
| `moai.md` unanchored / anchored | `grep -c` then `grep -cE` | `2` / `0` |
| package test state | `unset <9 vars> && go test -count=1 ./internal/template/...` | exit 0; `ok internal/template 58.375s`, `ok agentemit 0.365s`, `ok commandemit 0.279s`, `? scripts [no test files]` |

The package baseline is **fully green**. `plan.md` §C anticipated possible
pre-existing failures; none were present in this tree at `9dcbc3dbe`. The
comparison in AC-OSP-GATE-001 is therefore green-against-green, and any failure
after the edit would have been new by construction.

### E.2.1 AC matrix

| AC | Command | Actual output | Status |
|---|---|---|---|
| AC-OSP-001 | anchored loop over `<W>` | `moai.md 0`, `moai-easy.md 4`, `moai-learn.md 1` | PASS |
| AC-OSP-002 | anchored loop over `<T>` | `moai.md 0`, `moai-easy.md 4`, `moai-learn.md 1` | PASS |
| AC-OSP-003 | `cmp` loop over the three pairs | `identical moai.md`, `identical moai-easy.md`, `identical moai-learn.md` | PASS |
| AC-OSP-004 | `/config` loop over `<W>` | `moai.md 7`, `moai-easy.md 7`, `moai-learn.md 7` | PASS |
| AC-OSP-005 | `language.yaml` loop over `<W>` | `moai.md 7`, `moai-easy.md 3`, `moai-learn.md 6` | PASS |
| AC-OSP-006 | paired `git diff 9dcbc3dbe..HEAD --stat` (see E.2.3) | empty on both `moai.md` copies; non-empty naming both `moai-easy.md` copies | PASS |
| AC-OSP-007 | two isolated `-run` guard runs with `-v` | both exit 0; `--- PASS: TestTemplateNeutralityAudit`, `--- PASS: TestTemplateNoInternalContentLeak` named in the output | PASS |
| AC-OSP-008 | tripwire grep over the added template lines only | 5 added lines; grep matched nothing (exit 1) | PASS |
| AC-OSP-009 | `git diff 9dcbc3dbe..HEAD --name-only` (see E.2.3) | the four in-scope files plus this SPEC directory; nothing under `internal/` source, `pkg/`, `cmd/`, `docs-site/`, `.claude/agents/`, `.claude/skills/`, `.claude/commands/`, `.claude/rules/`, `.claude/hooks/` | PASS |
| AC-OSP-010 | unanchored then anchored `grep -c` on `<W>/moai.md` | `2` then `0` | PASS |
| AC-OSP-011 | byte comparison of the five added sentences (quoted in E.2.2) | the `moai-learn.md` sentence is not byte-identical to any of the four `moai-easy.md` sentences; all five differ from one another | PASS |
| AC-OSP-GATE-001 | `unset <9 vars> && go test -count=1 ./internal/template/...` after the edit | exit 0; `ok internal/template 56.508s`, `ok agentemit 0.171s`, `ok commandemit 0.084s`, `? scripts [no test files]` — no failure new against the E.2.0 baseline | PASS |
| AC-OSP-GATE-002 | no `make build`, `go build`, or `moai` invocation in this run | none was issued; only `go test`, `grep`, `cmp`, `sed`, `cp`, and `git` commands appear in the run-phase record | PASS |
| AC-OSP-GATE-003 | the two self-avoiding clarification sweeps | first: `0` for all four SPEC files; second: `1` | PASS |

C7 controls, each in the same invocation as the zero it witnesses: AC-OSP-001's
`moai.md 0` sits beside a `4` and a `1` from the same loop; AC-OSP-006's empty
diff is paired with a non-empty one; AC-OSP-010's anchored `0` is witnessed by
the unanchored `2` from the same file.

### E.2.2 The five added sentences, quoted for register review

`moai-easy.md` — colloquial, second-person, one phrasing per site:

1. line 21 (redirect to the professional persona): `(via `/config` → Output style → MoAI, or just type `/output-style MoAI` right here in the chat)`
2. line 493 (FAQ — switching to MoAI): `If you'd rather not click through menus, typing `/output-style MoAI` does exactly the same thing. And you can switch back either way whenever you like.`
3. line 496 (FAQ — redirect to the tutor persona): `(`/config` → Output style → MoAI-Learn, or `/output-style MoAI-Learn` if typing is quicker for you)`
4. line 533 (the standing switch line): `— or type `/output-style` followed by the name you want, like `/output-style MoAI-Learn`. Both routes land in the same place, so use whichever feels easier.`

`moai-learn.md` — inside the `[HARD]` bullet's quoted instruction, tutor voice:

5. line 35: `"Switch to MoAI via /config → Output style → MoAI — or, if you already know which persona you want, name it directly with /output-style MoAI"`

The four `moai-easy.md` additions are phrased differently from one another
rather than being one parenthetical repeated four times — the failure mode
`plan.md` §B.1 named and AC-OSP-011 cannot mechanically see. Site 1 offers the
command as an alternative to the menu, site 2 frames it as skipping the clicks,
site 3 as the quicker option for a typist, site 4 teaches the command's shape
(`/output-style <name>`). The `moai-learn.md` addition keeps the tutor's
conditional framing ("if you already know which persona you want") and stays
inside the quoted instruction and its unbackticked style.

### E.2.3 Diff evidence

`git diff 9dcbc3dbe..HEAD --stat` on both `moai.md` copies printed nothing and
exited 0; the same command on both `moai-easy.md` copies printed a non-empty
stat naming both files. The changed-file set is the four files named in
`spec.md` §C plus this SPEC directory (`spec.md` frontmatter `status`, and this
`progress.md`).

### E.2.4 Gaps and residual risk

- **Gap** — the M2 evidence report under `.moai/reports/t906/` was not authored or exported to the primary checkout by this run phase. `plan.md` §D M2 and `acceptance.md` §F require it; this run phase recorded the full evidence here in `progress.md` (a tracked path) instead. The export remains outstanding and is reported as a gap rather than worked around.
- **Residual risk** — AC-OSP-011 compares across files only. That the four `moai-easy.md` sentences differ from one another is asserted by review of the quotes in E.2.2, not by a mechanical check; `acceptance.md` names this limit explicitly.
- **Residual risk** — the embed refresh owed by the template-side edit (`make build` + the emit-embed doctor check) is deferred to batch close per REQ-OSP-010. Until it runs, an installed binary embeds the pre-edit output-style bytes.
- **Not claimed** — that `/output-style <style>` completes a switch. `spec.md` §A.4 observed the command's usage output only, and this run phase probed nothing further.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-09-18
run_commit_sha: `71a4d3b4e` — the M1 content commit. This line is recorded by a
  small follow-up commit on the same branch, since a commit cannot name its own
  SHA; the follow-up touches this file only.
run_status: complete
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 2 (`moai.md` working copy and template mirror — both asserted byte-unchanged by AC-OSP-006)
l44_pre_commit_fetch: not performed — the card commits to its own worktree branch `WT-output-style-switch-path` and does not push; integration is the lead's batch-close step
l44_post_push_fetch: not applicable — no push performed by this run phase
new_warnings_or_lints_introduced: none. `go test ./internal/template/...` exit 0 before and after; both isolated neutrality guards exit 0
cross_platform_build: not performed, and not owed — no criterion of this SPEC is decided by a built binary (REQ-OSP-010); `make build` / `go build` / any `moai` invocation would have failed AC-OSP-GATE-002
total_run_phase_files: 6 (4 in-scope content files + `spec.md` frontmatter status + this `progress.md`)
m1_to_mN_commit_strategy: one content commit (`71a4d3b4e`) on `WT-output-style-switch-path`, plus one follow-up commit recording that SHA in this file; M2 evidence export outstanding (see §E.2.4)

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
