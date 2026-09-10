# SPEC Review Report: SPEC-CODEX-SKILL-LOADER-001 (card t452)

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.79** — Tier M PASS threshold 0.80 (spec-workflow.md § SPEC Complexity Tier)

Reasoning context ignored per M1 Context Isolation. The audit reads only the four artifacts.

## Pinned state (artifacts were being edited during this audit)

The artifacts are untracked and were rewritten twice while I read them (two mid-read change
notices). This verdict is pinned to these content hashes:

| artifact | sha256 (12) |
|---|---|
| spec.md | `7554f03441ff` |
| plan.md | `8c0f256c547d` |
| acceptance.md | `2a5ffe5dda4b` |
| progress.md | `f8b7844e27cb` |

A later edit may already have addressed a finding below. Re-audit against a quiesced tree.

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -oE 'REQ-CSL-[0-9]{3}' spec.md | sort -u` → 001..013,
  contiguous, no duplicates, uniform 3-digit padding; `grep -cE '^- \*\*REQ-CSL-'` → 13 definitions.
- **[PASS] MP-2 GEARS compliance (requirement layer)** — judged against the 13 `REQ-CSL-*` entries in
  `spec.md` §C only. 001/004 `While`, 002/003/006/007/008 `When`, 006/012 `Where`, 005/010/011
  unwanted (`…해서는 안 된다`), 009/013 ubiquitous (`…해야 한다`). No Given-When-Then entry sits in the
  requirement layer; the Given-When-Then in `acceptance.md` is the correct verification-layer form
  and is graded under Group 4.
- **[PASS] MP-3 frontmatter validity** — all 12 canonical fields present plus optional `tier: M`
  (`grep -nE '^(id|title|version|status|created|updated|author|priority|phase|module|lifecycle|tags|tier):' spec.md | wc -l` → 13);
  `grep -nE '^(created_at|updated_at|labels|spec_id):' spec.md` → no output. `phase: "v3.2.0 target"`
  is a release target, not a prohibited lifecycle token.
- **[N/A] MP-4 language neutrality** — single-domain SPEC (Go emitter + codex CLI). It carries no
  multi-language tooling surface, so the 16-programming-language enumeration does not apply.
  (`REQ-CSL-012` is the template *internal-content* axis, a different concern.)
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — every referenced SPEC resolves and none is
  retired/superseded/archived: SPEC-CODEX-SKILLS-CANONICAL-001 `completed`,
  SPEC-CODEX-SKILL-NEUTRAL-001 `completed`, SPEC-CODEX-WIRING-001 `completed`. No BLOCKING.
- **[PASS] MP-6 D8 cross-platform** — `grep -c 'syscall'` → 0 across all three artifacts. Auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-SKILL-LOADER-001/`
  → no match. `progress.md:10` states none open.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.80 | 0.75 | Inference is marked as inference (spec.md:92, :94); the A/B branch split reads unambiguously. Deductions: `acceptance.md:42` "확인됨" has two readings; plan.md:55 defers the observable load signal entirely to run-phase. |
| Completeness | 0.75 | 0.75 | All required sections present; 4 `### Out of Scope — <topic>` sub-headings with specific bullets (spec.md:153,158,163,168); `moai spec lint` → 0 error. Deductions: D3, D6 below. |
| Testability | 0.60 | 0.50 | One criterion unsatisfiable as written (D1), two vacuous on the branch the SPEC itself calls likelier (D2), no run-provenance on M0 evidence (D4). Offset by a genuine control group at `acceptance.md:79`. |
| Traceability | 1.00 | 1.0 | 13 REQ ↔ 13 AC bijection: `grep -oE 'AC-CSL-[0-9]{3}' acceptance.md \| sort -u` → 001..013, `grep -cE '^### AC-CSL-'` → 13, and spec.md §E maps each REQ to exactly one AC. See "lint control group" below. |

### Lint control group — independently verified, not accepted on report

`moai spec lint .../SPEC-CODEX-SKILL-LOADER-001/spec.md` → `0 error(s), 13 warning(s)`, all
`CoverageIncomplete`, one per REQ — including REQ-CSL-001, which spec.md §E maps to AC-CSL-001.
Control: `moai spec lint .../SPEC-CODEX-SKILL-NEUTRAL-001/spec.md` (also `tier: M`, `completed`) →
`0 error(s), 16 warning(s)`, 15 of them the same `CoverageIncomplete` class over its own 15 REQs.
The rule reads `spec.md` alone and sees neither `acceptance.md` nor the §E table, so it emits the
finding for **every** REQ of **every** Tier M SPEC. Linter limitation, not a traceability gap.

## Defects Found

**D1 — `acceptance.md:38-42` (AC-CSL-005) — unsatisfiable as written (impossible-at-arrival) — Severity: critical — Class: blocking.**
The `When` sweeps *every* key of the emitted TOMLs; the `Then` forbids any key that **AC-CSL-004**
has not judged `확인됨`. AC-CSL-004 (`acceptance.md:32-36`) judges exactly one key — `skills`.
The emitted set today is 7 other keys (`args`, `command`, `description`, `developer_instructions`,
`model_reasoning_effort`, `name`, `sandbox_mode`), each confirmed by a **prior** SPEC and none of
them judged by AC-CSL-004. Read literally the criterion is red at arrival and red after any correct
work, which is the "impossible" direction `verification-completeness.md` §2 names. It also does not
measure REQ-CSL-005 (`spec.md:129`), which constrains only fields *this* SPEC would newly emit.
**Required fix:** scope the `When` selector to the field(s) this SPEC introduces (`grep '^skills'`),
or state explicitly that a prior-SPEC measurement counts as `확인됨` and name where that record lives.

**D2 — `acceptance.md:99` (완료 정의, branch-B row) — mandates recording a vacuous green as PASS — Severity: major — Class: blocking.**
The line requires AC-CSL-010~013 to be `PASS` on the branch-B path, parenthesised
"(프로브만으로도 이 넷은 적용된다)". That is true for 010 and 011 and false for the other two:
AC-CSL-012 (`acceptance.md:83-85`) greps *changed files under* `internal/template/templates/`, and
branch B changes none — the swept set is empty and the criterion asserts nothing
(`verification-completeness.md` §1.1). AC-CSL-013 (`:89-91`) is half-empty the same way: its
`os.Stat` clause has no subject when branch B adds no Go code, and its hardcoding grep has no subject
unless a probe script is deliberately committed, which no REQ requires. AC-CSL-011 is the
counter-example that shows the SPEC already knows the shape — its control-group clause (`:79`) would
expose an empty operand, and 012/013 have no equivalent.
**Required fix:** on the branch-B path record 012/013 as 해당 없음, or give each a swept-count that
must be non-zero plus a control expression, mirroring `:79`.

**D3 — `plan.md:46-57` (M0 probe table) — the four candidate roots do not discriminate — Severity: major — Class: blocking.**
The whole SPEC branches on M0's verdict, so an M0 that cannot attribute a load to a specific root
decides nothing. Three concrete holes: (a) the fixture the SPEC pre-built (`spec.md:101`) reaches the
same `SKILL.md` through **two** paths — `.agents/skills/moai-probe-x` is a symlink into
`.claude/skills/moai-probe-x` — so a load could be produced by either, and `.claude/skills/` is not
in the candidate list at all even though the codex binary's own migration-detection blob names
`.claude.json` alongside it (`spec.md:85`); (b) nothing requires per-root isolation or a distinct
marker name per root, so two roots can yield identical observations; (c) no negative control (a run
with no marker planted, showing the signal absent) is required, and the observable signal itself is
left undefined (`plan.md:55` says only "정해 놓고"), which under P4's finding that `codex doctor`
enumerates no roots (`spec.md:88`) is the hardest open question in the milestone.
**Required fix:** name the observable signal in `plan.md` (or make its selection an AC-judged step);
require one root populated per probe run with a per-root distinct marker name; add `.claude/skills/`
as a fifth candidate; require a no-marker negative control run.

**D4 — `acceptance.md:11-17` (AC-CSL-001) — M0 evidence carries no run-provenance — Severity: major — Class: optional.**
The `Then` requires a file holding a command, output and verdict; it does not require the recorded
output to have been produced **in this run, at this codex version, with a recorded exit code**. That
is the specific gap the lead asked me to test, and the structural resolution is *nearly* complete
rather than complete: the two new 판정 표면 제한 clauses (`:16`, `:17`) do exclude the two surfaces a
shortcut would reach for (`codex doctor`, `strings`), and AC-CSL-010 (`:73`) forces the probe command
text to exhibit an isolated `CODEX_HOME`, which raises the cost of citing a carried-over claim
considerably. What remains reachable: a run could record output that never came from those exhibited
commands, and nothing would catch it. The exposure is worst on branch B, where AC-CSL-009 — the only
criterion that pins the codex version anywhere — is recorded 해당 없음 (`:99`), so the branch whose
entire value is "the inherited premise is false **at 0.152.1**" pins no version mechanically.
`verification-claim-integrity.md` §2 requires command + observed output in this run against this
tree; `verification-completeness.md` §2.1 adds exit code and a pinned SHA.
**Required fix:** add to AC-CSL-001's `Then` — each per-root record carries the exit code, the
`codex --version` string observed in the same run, and the tree SHA; and make the codex version a
required element of the branch-B blocker report in AC-CSL-003 (`:29`), so version attribution
survives on both paths.

**D5 — `spec.md:36` (§A.1) — the out-of-scope decision rests on an absence claim wider than its selector — Severity: minor — Class: optional.**
`grep -rn 'skills.config' internal | grep -v _test` is cited to conclude "현재 Go 코드베이스는 이런
항목을 쓴 적이 없다". The selector covers `internal/` only and matches one literal, so it cannot see
a writer in `pkg/`/`cmd/` or one that marshals a struct rather than writing the header text. I widened
it (`grep -rn 'skills\.config\|SkillConfig\|skills_config' internal pkg cmd | grep -v _test`) and the
conclusion holds — only the read-only parser and the advisory doctor appear — but the SPEC's own
sentence claims more than the SPEC's own command shows, and the §D out-of-scope boundary for the 49
ghost entries rests on it. The boundary itself is drawn defensibly: the entries point at paths that do
not exist, t451's doctor already reports them, and REQ-CSL-011 forbids the user-layer write that
cleaning them would require. Only the evidence sentence overreaches.
**Required fix:** narrow the sentence to the scanned scope, or cite the widened command.

**D6 — `acceptance.md:48` (AC-CSL-006) + `spec.md:135` (REQ-CSL-008) — the literal `11` — Severity: minor — Class: optional.**
`ls internal/template/templates/.codex/agents/moai/*.toml | wc -l` → 11 today, so the number is
correct now and rots the moment an agent is added. Worse, "재생성된 11 개 TOML **각각**에 …
스킬 필드가 존재한다" pre-decides that the emission rule applies uniformly to all 11 classes — a
design constraint the M0/M2 measurement has not earned.
**Required fix:** derive the count (`ls … | wc -l`) rather than literalise it, and phrase the field
assertion over the agents the emission rule selects rather than over all of them.

**D7 — `spec.md:13` / `progress.md:7` — Tier M rests on the branch the SPEC calls less likely — Severity: minor — Class: optional.**
The file-axis justification (manifest + emitter source + 11 emitted TOMLs + golden tests) is sound and
lands inside the Tier M 5-15 file band — but only on **branch A**. On branch B, which `spec.md:96`
now calls "더 유력한 쪽", the change set is `progress.md` plus evidence files: zero code files, Tier S
shape. The REQ/AC budget (13/13 against the Tier M ceiling of 16/16) is comfortable either way.
Tier is fixed at plan time and is not worth re-litigating; recording the asymmetry is enough.
**Required fix:** none required. Optionally note in `plan.md` §B that branch B collapses the file axis.

## Recommendation

FAIL at 0.79 against the Tier M threshold of 0.80, on three blocking findings — but the gap is small
and the fixes are all wording/scope edits inside `acceptance.md` and `plan.md`, not a redesign. The
SPEC's central move (measurement before emission, branch B priced as a deliverable rather than a
failure) is sound and survives adversarial reading: D3 in `spec.md` §B (`:111`) states it,
`plan.md:59-65` gates it, `acceptance.md:99` gives it its own completion definition, and `plan.md:98`
names the pressure toward branch A as an anti-pattern. Fix in this order:

1. **D1** — rescope AC-CSL-005's selector to the field this SPEC introduces (`acceptance.md:41-42`).
2. **D2** — on the branch-B path record AC-CSL-012/013 as 해당 없음, or give each a non-zero swept
   count plus a control expression modelled on `acceptance.md:79` (`acceptance.md:99`).
3. **D3** — name the observable load signal, require per-root isolation with distinct marker names,
   add `.claude/skills/` as a fifth candidate, and require a no-marker negative control
   (`plan.md:46-57`).
4. Optional, cheap, and worth taking while iterating: **D4** (exit code + `codex --version` + tree SHA
   on every M0 record; codex version added to AC-CSL-003's required blocker contents) and **D5**.

Iteration 2 should be scoped to this enumerated delta plus a regression check over D1-D3 — not a
from-scratch re-audit. Re-audit against a quiesced tree; the hashes above are what iteration 2's
regression check compares to.
