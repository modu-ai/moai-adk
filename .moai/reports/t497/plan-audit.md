# SPEC Review Report: SPEC-CODEX-BODY-NEUTRALITY-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: 0.63
Auditor: plan-auditor (adversarial, M1–M6 bias-prevention active)
Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t497` · branch `WT-codex-neutrality`
HEAD at audit start: `46ea34137` · **HEAD at audit end: `53886ba40`** (see D0)

Reasoning context ignored per M1 Context Isolation. The audit read only the four SPEC
artifacts plus `.moai/reports/t497/measurement.md` and the in-tree evidence it cites; every
load-bearing number below was re-run in this tree in this session.

---

## D0 — Process defect: the audited tree gained a writer mid-audit

- **Claim.** A commit landed on the audited worktree while this audit was in progress, so the
  audited input set changed under the audit.
- **Evidence.**
  ```
  $ git log --oneline -3
  53886ba40 docs(t497): resolve the 3-vs-4 binding-row gap from in-tree evidence
  46ea34137 docs(t497): SPEC-CODEX-BODY-NEUTRALITY-001 plan artifacts + baseline corrections
  ada8dcf88 docs(t497): addendum — mirror-skill match distribution rules out a blanket sweep
  $ git show --stat --oneline 53886ba40
   .moai/reports/t497/measurement.md | 45 +++++++++++++++++++++++++++++++++++++++
  ```
  The dispatch named HEAD `46ea34137`. `53886ba40` is a foreign commit relative to this audit.
- **Baseline-attribution.** This tree, this run; `git log` / `git show` above.
- **Gaps.** I did not determine which session authored `53886ba40`.
- **Residual-risk.** `agent-common-protocol.md` § Background Agent Execution binds an actively
  audited worktree to exactly one writer. Reported, not repaired. The audit below is pinned to
  the artifacts as they stand at `53886ba40`; the four SPEC files are byte-unchanged between the
  two commits (`git show --stat` shows only `measurement.md`), so no verdict is invalidated —
  but D1 exists *because* of this commit.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-CBN-001 … REQ-CBN-015`, each defined exactly
  once, no gaps, uniform 3-digit padding.
  `grep -o 'REQ-CBN-[0-9]\{3\}' spec.md | sort | uniq -c` → 15 rows, all count 1, 001→015.
  AC ids `AC-CBN-001 … AC-CBN-012` likewise contiguous.
- **[PASS] MP-2 GEARS format compliance (requirement layer only)** — judged against the
  `REQ-XXX` layer in `spec.md` §C, **not** against `acceptance.md`, whose Given-When-Then form is
  the correct verification-layer format. All 15 REQs carry a GEARS shape: ubiquitous
  (`…해야 한다`), unwanted (`…해서는 안 된다`), `When`(`…한다,`), `While`, `Where`. No
  Given-When-Then appears as a REQ. Two soft deductions, not failures: `REQ-CBN-002`
  (spec.md:111) uses `Where` for a per-line property judgement, which is a `When`/`While`
  condition rather than a capability gate; `REQ-CBN-011` (spec.md:126) is a subjectless passive
  ("41줄의 문면은 유지되어야 한다").
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct
  types (spec.md:2-15): `id`, `title`, `version: "0.1.0"` (quoted semver), `status: draft`,
  `created`/`updated: 2026-09-07` (ISO), `author`, `priority: P2`, `phase`, `module`,
  `lifecycle: spec-anchored`, `tags` (comma-separated string). No rejected snake_case alias
  (`created_at` / `updated_at` / `labels` / `spec_id`) present. Extra keys `tier: M` and
  `related_specs` are additive.
- **[N/A] MP-4 language neutrality** — the SPEC is scoped to the Claude↔Codex harness axis and
  names no programming-language toolchain; the 16-language enumeration obligation does not
  engage. N/A auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — three referenced SPECs, all present, none in
  {retired, superseded, archived}:
  ```
  SPEC-CODEX-DUAL-AGENTS-001      status: completed
  SPEC-CODEX-SKILL-NEUTRAL-001    status: completed
  SPEC-CODEX-SKILLS-CANONICAL-001 status: completed
  ```
  No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → `0`. Auto-PASS.
- **[FAIL] MP-7 clarification gate** — two unresolved markers. **Critical, score-independent.**
  ```
  $ grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/
  plan.md:45: [NEEDS CLARIFICATION: 코덱스 능력 부재 측정 가능 여부]
  plan.md:46: [NEEDS CLARIFICATION: M5 미러 스킬 77파일 착수 여부]
  ```
  (`research.md` absent — Tier M, not required.) See D1 and D8 for the per-marker judgement the
  dispatch asked for.

**MP-7 alone forces `Verdict: FAIL`.** D1, D2, D3 and D5 are independently blocking.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.50 | 0.50 — multiple requirements require interpretation | Three mutually inconsistent expected values for the same M1 outcome (acceptance.md:63 = 4, plan.md:62 = 3, acceptance.md:105 = 5); `spec.md:123` names an ambiguous source path; `acceptance.md:93` "M2·M3 이 명시한 대상 수" is never quantified |
| Completeness | 0.75 | 0.75 — one area sparse, frontmatter complete | All sections present; four `### Out of Scope — <topic>` H3s with concrete bullets (spec.md:139/145/150/154); §E delegates traceability to `acceptance.md` by design. Sparse: no AC asserts the existence of the two M1/M2 report artifacts |
| Testability | 0.50 | 0.50 — several ACs need judgement or are already-green | AC-CBN-006/007/009 all pass on the untouched tree (D2); AC-CBN-002 pins a coordinate that carries no such token (D3); AC-CBN-011 is unquantified (D5) |
| Traceability | 0.75 | 0.75 — one requirement indirectly covered | All 15 REQs appear in a `maps REQ-…` line; all 12 ACs map to REQs that exist. But REQ-CBN-009 ("every directive line is rewritten") is covered only for `Task*` (D6) |

Aggregate: (0.50 + 0.75 + 0.50 + 0.75) / 4 = **0.625**.

---

## Verified-correct findings (adversarial checks that the SPEC survived)

Recorded so a re-audit does not re-litigate them.

- **The 84 sum is right — independently reproduced.**
  ```
  $ grep -rhoE 'AskUserQuestion|TaskCreate|TaskUpdate|TaskList|TaskGet|DesignSync|Skill\(|Agent\(' \
      internal/template/templates/.codex/agents/moai/*.toml | wc -l
        84
  $ (same pattern, grep -rhE ... | wc -l)          # distinct lines
        81
  per-file: 14 manager-develop · 13 manager-spec · 13 manager-lead · 10 manager-design ·
            8 manager-docs · 5 sync-auditor · 5 super-advisor · 5 manager-git ·
            4 plan-auditor · 4 e2e-tester · 3 builder-harness
  ```
  Both the per-file decomposition and the per-pattern table (`41/45/25/6/4/4`) reproduce exactly.
  The lane's self-correction 74 → 84 is correct, and the 84-occurrence / 81-line unit split is
  correct. **No further correction needed.**
- **Correction 3 (budget headroom) is right.**
  `go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v` →
  `always-loaded surface = 74535 tokens (budget 77600, headroom 3065, 17 entries)`, `PASS`.
  The anchored selector is non-vacuous — it matches exactly one test and that test emits the log
  line the AC reads.
- **The M1 premise holds.** I read the `agents-codex.yaml` rationales myself. `skill-loader`
  (:133-) records "Session skill loading itself **IS confirmed** on this version (the roots table
  lists the project `.agents/skills`) … this drop is about the per-agent grant, not about skills
  reaching a session"; `subagent-spawn` (:143-) records "Codex delegation **exists** (internal
  collaboration\* tools) but a per-agent spawn grant is not expressible in agent TOML". manager-spec's
  reading — *what is absent is the agent-TOML field, not the capability* — is the correct one, and
  it is corroborated by a sweep I ran independently: the three landed rows
  (`question-channel`, `task-list`, `design-sync`) are exactly the three documented-drop classes
  whose rationale asserts capability absence ("codex reported it unavailable"; "no known Codex
  equivalent"; "no Codex equivalent"), while every class whose rationale asserts a surviving
  surface (`file-write`, `web`, `skill-loader`, `subagent-spawn`, `cross-session-messaging`) has
  no row. **The derivation criterion is applied consistently in the landed table. M1's refusal
  to add the two rows is right.** What is wrong is the milestone built on top of it — D1, D4.
- **`grep -c '^| ' AGENTS.md` → 4** and the two copies' table regions are byte-identical
  (`diff <(sed -n …) <(sed -n …)`; rc=0, no output). The pre-work baselines in AC-CBN-006/007
  are accurate as measurements. Their *use* is the problem — D2.
- **AC-CBN-005's second condition is non-vacuous.**
  `grep -rn 'task-list' internal/template/templates/.codex/agents/moai/*.toml | wc -l` → `0`
  today, so "≥3 coordinates after the fix" cannot pass by accident. Good AC design: it closes the
  "just delete the lines" route that the `Task* → 0` check alone would leave open.

---

## Defects Found (structured defect-list)

### D1. Stale-premise — the SPEC defers to M1 a question its own baseline has since answered · `spec.md:51`, `plan.md:18,45,54` · Severity: **critical** · Class: **blocking**

- **Claim.** `spec.md:51` (정정 2) states the identity of the missing 4th binding row "**이 SPEC 이
  관측하지 않았다 — M1 의 산출 대상이다**", and `plan.md` §F M1 is built around a Codex probe to
  settle it. Both are false as of `53886ba40`: the question is answered, in the SPEC's own cited
  baseline, and it was answerable from committed evidence before that.
- **Evidence.** Two independent routes, both run in this tree in this session.
  1. The 4-row figure in `REQ-CSN-003` comes from a candidate table sized during the D7 budget
     exercise, and that table is committed and readable:
     ```
     $ cat .moai/reports/t196/csn003-table-4row.txt
     | question-channel | AskUserQuestion tool | return a blocker report |
     | task-list | TaskCreate, TaskUpdate, TaskList, TaskGet | report progress in prose |
     | design-sync | DesignSync tool | skip design sync |
     | cross-session-messaging | SendMessage, ListAgents tools | use the moai MCP broker |
     ```
     The 4th row is `cross-session-messaging`. `SPEC-CODEX-SKILL-NEUTRAL-001` §B.D7 names this
     exact file (`csn003-table-{11row,11row-honest,4row}.txt`) in the same paragraph the SPEC
     quotes for its doctrine sentence — so the evidence was one `cat` away from the citation the
     SPEC already made.
  2. Applying the derivation criterion to that row: `agents-codex.yaml` `cross-session-messaging`
     rationale — "The Codex **counterpart rides the moai MCP broker**
     (session_msg_register/list/send/poll)". A counterpart exists ⇒ capability present ⇒ no row.
     The landed 3 rows are the corrected derivation; **`REQ-CSN-003`'s "현재 측정값 4행" is the
     stale figure.** `53886ba40` reached the same conclusion by a third route (a 9-class rationale
     sweep) and appended it to `measurement.md:198-`.
- **Baseline-attribution.** This tree at `53886ba40`; `cat`, `grep`, and the `classes:` block of
  `internal/template/agentemit/agents-codex.yaml` read in this run.
- **Gaps.** I did not re-probe Codex runtime behaviour for any class; like the lane's sweep, my
  route 2 is a rationale-text judgement.
- **Residual-risk.** The four SPEC artifacts were **not** updated by `53886ba40`. A reader of
  `spec.md` §A.2 and `plan.md` §E/§F today is told to run a probe that the SPEC's own baseline
  now says is unnecessary, and NEEDS-CLARIFICATION ① asks the operator to approve a fallback for
  a measurement that is no longer on the critical path. The plan artifacts are stale against
  their cited baseline.
- **Required fix.** Fold the `measurement.md:198-` resolution into `spec.md` §A.2 정정 2 and
  `plan.md` §B-2. Re-scope M1 from "probe for capability absence" to what the evidence actually
  leaves open — "record the derivation and correct `REQ-CSN-003`'s stale 4행 figure" — and
  re-derive M1's verification accordingly (see D2). **Marker ① must be withdrawn or rewritten**:
  as posed it asks the operator to decide something the tree has already decided.

### D2. Vacuous milestone gate — every M1 verification passes on the untouched tree · `plan.md:59-62`, `acceptance.md:62-63,68-69,80-81` · Severity: **critical** · Class: **blocking**

- **Claim.** All three of M1's pre-fixed checks — and therefore must-pass AC-CBN-006, AC-CBN-007
  and AC-CBN-009 — print their expected output **right now, with M1 not started**. On the
  zero-absence path (which §B.1 already concludes is today's outcome) the milestone has no
  discriminating check at all.
- **Evidence.** Run against the untouched tree, before any M1 work exists:
  ```
  $ grep -c '^| ' AGENTS.md
  4                                        # AC-CBN-006 expects "4 + absent"; absent=0 ⇒ 4. PASSES.
  $ diff <(sed -n '/^\*\*Capability bindings/,/^---$/p' AGENTS.md) \
         <(sed -n '/^\*\*Capability bindings/,/^---$/p' internal/template/templates/AGENTS.md)
  ; echo rc=$?
  rc=0                                     # AC-CBN-007 expects no output, rc 0.  PASSES.
  $ go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v
  always-loaded surface = 74535 tokens (budget 77600, headroom 3065, 17 entries)
  PASS                                     # AC-CBN-009 expects PASS + positive headroom. PASSES.
  ```
- **Baseline-attribution.** This tree at `53886ba40`, commands above, no artifact created.
- **Gaps.** I did not construct the mutant that would show a non-vacuous variant green — the
  three commands' pre-work output *is* the demonstration.
- **Residual-risk.** M1 is the milestone the plan calls "되돌리기 가장 어려움". Its acceptance can
  be discharged by doing nothing. AC-CBN-006's `Given` names `capability-absence.md`, but its
  `Then` never reads that file, so the artifact's existence and content are unasserted; §F 완료
  정의 mentions it in prose, which is not a check.
- **Required fix.** Add a discriminating check to AC-CBN-006 that binds the artifact, not only
  the table: assert `.moai/reports/t497/capability-absence.md` exists **and** that the count of
  its `absent` verdict rows equals `(grep -c '^| ' AGENTS.md) - 4`, measured in the same run.
  Additionally scope the selector: `grep -c '^| '` reads the whole of `AGENTS.md`, so a `| `-line
  added anywhere in that file satisfies the count. Use the `sed -n '/^\*\*Capability
  bindings/,/^---$/p' AGENTS.md | grep -c '^| '` region form the SPEC already uses in AC-CBN-007.

### D3. False coordinate in a must-pass AC — `manager-develop.toml:64` carries no `AskUserQuestion` · `acceptance.md:38`, `spec.md:90` · Severity: **critical** · Class: **blocking**

- **Claim.** AC-CBN-002 instructs the judge to read four coordinates and assert all four are
  `verdict=prose` `AskUserQuestion` lines. One of the four does not exist as described, so the AC
  cannot be satisfied as written; and the "prose `AskUserQuestion` 4건" framing that produced it
  is a unit conflation of occurrences with coordinates.
- **Evidence.**
  ```
  $ grep -rn 'AskUserQuestion' internal/template/templates/.codex/agents/moai/*.toml
  sync-auditor.toml:131:  … no `sync-auditor` path invokes `AskUserQuestion` …
  plan-auditor.toml:146:  … MUST resolve each marked topic via `AskUserQuestion` … select:AskUserQuestion …
  super-advisor.toml:62:  … user via `AskUserQuestion`. The prescription is advisory …
  $ for f in …/*.toml; do n=$(grep -o 'AskUserQuestion' $f | wc -l); …   # per-file
  2 plan-auditor.toml
  1 super-advisor.toml
  1 sync-auditor.toml
  $ sed -n '62,68p' …/manager-develop.toml
  | SPEC creation, or an unclear SPEC | manager-spec |
  | Security audits … | per-spawn `Agent(general-purpose)` security reviewer … |
  …
  ```
  The population is **3 distinct lines / 4 occurrences** (plan-auditor:146 carries two).
  `manager-develop.toml:64-66` is a delegation routing table whose cells carry `Agent(`, not
  `AskUserQuestion`. `manager-develop.toml` has **zero** `AskUserQuestion` occurrences.
- **Baseline-attribution.** This tree at `53886ba40`, `grep -rn` / per-file `grep -o` / `sed -n`
  above.
- **Gaps.** I did not audit whether `manager-develop.toml:64-66` is correctly classified as prose
  under the `Agent(` axis — it plainly is; the defect is that it is filed under the
  `AskUserQuestion` heading and then pinned into an `AskUserQuestion` AC.
- **Residual-risk.** This is the exact hazard `plan.md` §G AP-3 names ("발생 84 와 줄 81 을 섞어
  쓰면 판정이 어긋난다"). §A.2's `AskUserQuestion 4` is an occurrence count; §B.4 spends it as
  four coordinates and AC-CBN-002 inherits the error. Left unfixed, the judge either fails a
  correct implementation or, worse, waves the AC through and normalizes ignoring a pinned
  coordinate.
- **Required fix.** Correct §B.4 to three `AskUserQuestion` coordinates
  (`sync-auditor:131`, `plan-auditor:146`, `super-advisor:62`) and state the unit
  (3 lines / 4 occurrences). Move `manager-develop.toml:64-66` to the `Agent(` axis. Rewrite
  AC-CBN-002 over the corrected coordinate set, and say which of them carry
  `subject=orchestrator` (plan-auditor:146 and super-advisor:62 do; sync-auditor:131 is a
  prohibition statement, a third category the AC's two-value `subject` column does not admit).

### D4. Internal contradiction — three different expected values for the same M1 outcome · `acceptance.md:63` (4) vs `plan.md:62` (3) vs `acceptance.md:105` (5) · Severity: **major** · Class: **blocking**

- **Claim.** For the identical scenario "M1 measures zero absent classes", the artifacts state
  three mutually exclusive expected outputs of `grep -c '^| ' AGENTS.md`.
- **Evidence.** Verbatim:
  - `acceptance.md:63` — "`absent` 가 0건이면 출력은 **4** 이고, 이것도 통과다".
  - `plan.md:62` — "행이 0개 추가되는 결과도 정당한 M1 완료다. 그때 검증은 **`N = 3`**".
  - `acceptance.md:105` — "부재가 하나도 실측되지 않는 경우. AC-CBN-006 은 출력 **5** 로 통과하고".
  Measured truth: `grep -c '^| ' AGENTS.md` → `4` (1 header + 3 body rows; the `|---|---|---|`
  separator is not matched by the `^| ` selector). So `acceptance.md:63` is right, `plan.md:62`
  is off by the header row, `acceptance.md:105` is off by a phantom added row.
- **Baseline-attribution.** This tree at `53886ba40`; `grep -c '^| ' AGENTS.md` → 4;
  `grep -n '^| ' AGENTS.md` → lines 19,21,22,23.
- **Gaps.** None — all three literals are quoted and the measurement is direct.
- **Residual-risk.** `acceptance.md:105` sits in §D 경계 사례, the section a judge reads precisely
  when the zero-absence path materialises — the most likely outcome per §B.1. A judge following
  §D fails a correct M1.
- **Required fix.** Make `4` the single stated value in all three places, or state the value once
  and cross-reference it from the other two.

### D5. Scope containment is not pinned — an unrelated `.md` edit rides along undetected · `plan.md:95-96`, `acceptance.md:93` · Severity: **major** · Class: **blocking**

- **Claim.** The dispatch asks whether the M4 radius is pinned concretely enough to catch an
  unrelated `.md` edit. It is not. The allowlist is directory-glob-shaped and the file-count
  cross-check is unquantified, so an edit to any of the 11 agent `.md` files — including one from
  another card — lands inside the permitted radius and passes AC-CBN-011.
- **Evidence.** `plan.md:95` permits `{.claude/agents/moai/*.md,
  internal/template/templates/.claude/agents/moai/*.md, …/.codex/agents/moai/*.toml, AGENTS.md,
  internal/template/templates/AGENTS.md, .moai/specs/<this>/*, .moai/reports/t497/*}`. That glob
  covers all 11 agents on both sides; M3 names only 4 target files
  (`manager-develop`, `e2e-tester`, `manager-design`, `manager-lead`), one of them conditional on
  M1. `acceptance.md:93`'s second clause — "변경 파일 수가 M2·M3 이 명시한 대상 수와 일치" —
  has no number to compare against: `plan.md` §F M3's table gives 4 rows of which the
  `manager-lead` row is gated on M1 having created a row, and §B.1 concludes it will not.
- **Baseline-attribution.** This tree at `53886ba40`; verbatim quotation of `plan.md:95-96` and
  `acceptance.md:93`; M3 target table at `plan.md:77-83`.
- **Gaps.** I did not run a golden regeneration to observe an actual radius.
- **Residual-risk.** The plan's own §B-4 names golden-radius pollution as a known hazard, then
  writes the check that would catch it in a form that cannot. The `.moai/reports/t497/*` and
  spec-dir globs additionally make the count non-deterministic.
- **Required fix.** Pin the radius by file, not by glob: enumerate the exact source `.md` files
  M3 will touch and derive the expected regenerated `.toml` set from them; state the expected
  `git diff --name-only` set as a literal list and assert set equality, not a count.

### D6. REQ-CBN-009 is under-covered — only the `Task*` directives have an AC · `spec.md:124`, `acceptance.md:53-57` · Severity: **major** · Class: **blocking**

- **Claim.** REQ-CBN-009 binds *every* line judged `directive`. The only AC that maps to it
  (AC-CBN-005) tests `Task*` alone. The two other directive classes the SPEC itself identifies —
  `manager-lead`'s self-spawn lines and the `manager-design` design-sync ladder — have no
  acceptance criterion.
- **Evidence.** `acceptance.md:53` maps `AC-CBN-005 · maps REQ-CBN-009` and its `When` is
  `grep -rhoE 'Task(Create|Update|List|Get)'`. `spec.md:94` classifies
  `manager-lead.toml:37,57,59,193` as "지시, 범위 안"; those coordinates exist and are self-spawn
  directives:
  ```
  $ grep -n 'Agent(' …/manager-lead.toml
  37: … spawning and orchestrating write-capable leaf workers (per-spawn `Agent(gene…
  57: … manager-lead spawns a se…
  59: … Background parallel dispatch (lead posture) …
  193: … delegate coordination duties to a deputy — a manager-lead instance spawned as an UNNAMED ba…
  ```
  `spec.md:98` requires a paragraph near `manager-design.toml:109`. Neither has an AC; both are
  additionally gated in `plan.md` §F M3 on M1 creating a `subagent-spawn` row that §B.1 says
  cannot be created today.
- **Baseline-attribution.** This tree at `53886ba40`; `grep -n 'Agent(' …/manager-lead.toml`;
  `acceptance.md` §C read in full.
- **Gaps.** Whether the 6 remaining `Agent(` lines in `manager-lead.toml` are directives is M2's
  output; not judged here.
- **Residual-risk.** The card can land with AC-CBN-001..012 all green while the `manager-lead`
  and `manager-design` directive lines — two of the three in-scope directive classes — are
  untouched, because `acceptance.md` §D:108 explicitly permits skipping the `manager-lead`
  reference when M1 adds no row, and nothing then requires the alternative ("본문에 대체 행동을
  직접 적는다") to have happened.
- **Required fix.** Add an AC per directive class with a positive, measurable post-state:
  for `manager-design`, that the fallback paragraph exists near the ladder; for `manager-lead`,
  that each of the four coordinates carries either a `subagent-spawn` binding reference or an
  in-body fallback sentence. Both are greppable.

### D7. `REQ-CBN-008` names a source path that is not the emitter's input and is prohibited by a landed SPEC · `spec.md:123` · Severity: **major** · Class: **blocking**

- **Claim.** REQ-CBN-008 says edits go into "중립 소스 `.claude/agents/moai/*.md`". Read
  literally that is the repository-root copy, which the emitter does not read and which
  `SPEC-CODEX-SKILL-NEUTRAL-001` REQ-CSN-012 forbids using as an origin.
- **Evidence.**
  ```
  $ sed -n '30,34p' internal/template/agentemit/golden_test.go
  // templatesDir is the template tree root relative to this package's dir.
  const templatesDir = "../templates"
  // agentMDRoot is the neutral-layer source root inside the template tree.
  const agentMDRoot = ".claude/agents/moai"
  ```
  → the emitter reads `internal/template/templates/.claude/agents/moai/`, i.e. the **template**
  copy. And `REQ-CSN-012` (completed, spec-anchored): "모든 편집은
  `internal/template/templates/**` 를 원본으로 수행해야 하며, 로컬 `.claude/**` 사본을 원본으로
  편집해서는 안 된다."
- **Baseline-attribution.** This tree at `53886ba40`; `golden_test.go:30-34`;
  `.moai/specs/SPEC-CODEX-SKILL-NEUTRAL-001/spec.md` REQ-CSN-012.
- **Gaps.** I did not diff the root and template copies of the 11 agent `.md` files.
- **Residual-risk.** Compounds D5: `plan.md:95`'s allowlist admits **both** copies, so an
  implementer who edits the wrong one stays inside the permitted radius. The error is caught
  downstream (AC-CBN-005 would fail because the TOML would not change), but only after a
  regeneration, and only if the wrong-copy edit is not accompanied by a right-copy one.
- **Required fix.** Write the path in full in REQ-CBN-008
  (`internal/template/templates/.claude/agents/moai/*.md`) and drop the bare
  `.claude/agents/moai/*.md` entry from the M4 allowlist.

### D8. Marker ② (M5 mirror-skill scope) is a genuine operator decision, but the SPEC's disposition rests on an unverified premise it does not label as one · `plan.md:46`, `spec.md:102,156`, `measurement.md:133-145` · Severity: **minor** · Class: **optional**

- **Claim.** The dispatch asks whether M5's disposition hides an unverified assumption. It does
  not hide it — `measurement.md:143-145` states the Gap explicitly — but that Gap does not travel
  into `spec.md` §B.6 or §D, which present the deferral as settled scope reasoning.
- **Evidence.** `measurement.md:133` — "719 중 639(89%) 가 상위 3개 디렉터리에 몰려 있고, 그
  셋은 주제 자체가 Claude Code 인 문서다"; and `measurement.md:143` — "위 판단은 분포와 세
  문서의 주제에 근거한 scope 판단이지, **639건이 전부 산문임을 확인한 것이 아니다**. 상위 3본을
  범위에서 빼려면 표본 확인이 별도로 필요하다." `spec.md:102` (§B.6) and `spec.md:156` (§D)
  carry the deferral without that qualification.
- **Baseline-attribution.** `measurement.md` at `53886ba40`; `spec.md` §B.6/§D read in full. I
  did not re-run the 719/639 counts — they are not load-bearing for the verdict, since the
  disposition is "defer", the conservative direction.
- **Gaps.** The 719 and 639 figures are cited, not re-measured by me. Sampling of the three large
  directories was not performed by anyone.
- **Residual-risk.** Low as long as the disposition stays "defer to a follow-up card" — deferral
  cannot be wrong in the damaging direction. It becomes load-bearing the moment the follow-up
  card scopes itself to "the other 12 directories, 71 matches" on the strength of this
  unsampled premise.
- **Required fix (optional).** Carry the `measurement.md:143` Gap sentence into `spec.md` §B.6
  so the follow-up card inherits the qualification rather than the conclusion.
- **Marker judgement (as the dispatch requested).** Marker ② **is** a genuine operator decision:
  it is a scope/batch-size question with no right answer derivable from the tree, the SPEC states
  a default (defer), and both branches are coherent. Keep it; it needs an operator answer, not a
  measurement.

### D9. AC-CBN-001's row count does not establish correspondence · `acceptance.md:29-33`, `plan.md:69` · Severity: **minor** · Class: **optional**

- **Claim.** AC-CBN-001 asserts "classification rows = 84 and population = 84, measured together".
  Pairing the two measurements is good design and closes the population-drift hole. It does not
  close the correspondence hole: 84 rows all citing the same coordinate satisfy it.
- **Evidence.** `plan.md:69` — `grep -c '^| .*\.toml | [0-9]' body-classification.md` → 84;
  `plan.md:70` re-measures the population → 84. Neither joins the two sets.
- **Baseline-attribution.** `plan.md` §F M2 and `acceptance.md` AC-CBN-001, read at `53886ba40`.
- **Gaps.** No artifact exists yet to test the selector against.
- **Residual-risk.** Low — the classification table is human-reviewed downstream — but the cheap
  strengthening exists.
- **Required fix (optional).** Sort the table's `file:line` column and diff it against
  `grep -noE '<union pattern>' …/*.toml | cut -d: -f1-2 | sort` (81 distinct lines; the 3
  double-token lines are the documented delta).

### D10. §D's 41-line cover check is not a runnable command · `acceptance.md:105` · Severity: **minor** · Class: **optional**

- **Claim.** The boundary case prescribes verifying the cover sentence with
  `grep -c '.agents/skills'` — no path argument, and `.` is an unanchored wildcard.
- **Evidence.** `acceptance.md:105` verbatim: "그 문장의 존재를 `grep -c '.agents/skills'` 로
  확인한다". Baseline for the intended check, measured now:
  ```
  $ grep -rc 'agents/skills' internal/template/templates/.codex/agents/moai/*.toml   # all zero
  $ grep -rl 'agents/skills' internal/template/templates/.claude/agents/moai/ | wc -l
        0
  ```
  So the check would be non-vacuous if written runnably (baseline 0), and it is not promoted to
  an AC at all.
- **Baseline-attribution.** This tree at `53886ba40`, commands above.
- **Gaps.** None.
- **Residual-risk.** The zero-absence path is the likely path (§B.1); on that path the cover
  sentence is the *only* deliverable for the 41 `invoke Skill(` lines, and nothing must-pass
  asserts it exists.
- **Required fix (optional).** Promote it to an AC with a runnable form, e.g.
  `grep -rl '\.agents/skills' internal/template/templates/.codex/agents/moai/*.toml | wc -l` → ≥1,
  against the measured baseline of 0.

---

## Marker judgements requested by the dispatch

| Marker | Genuine operator decision? | Basis |
|---|---|---|
| ① 코덱스 능력 부재 측정 가능 여부 (`plan.md:45`) | **No — answerable from the tree, and now answered** | The absence/presence judgement for all 9 documented-drop classes is recorded in `agents-codex.yaml` rationales; the landed 3 rows are the correct derivation; the "4행" figure is a stale candidate-table number whose 4th row (`cross-session-messaging`) is committed at `.moai/reports/t196/csn003-table-4row.txt`. `measurement.md:198-` reaches the same conclusion and states the marker becomes unnecessary. **The SPEC artifacts do not reflect this.** Withdraw or re-pose the marker; do not spend an operator round on it. See D1. |
| ② M5 미러 스킬 77파일 착수 여부 (`plan.md:46`) | **Yes** | A scope/batch-size decision with a stated default (defer) and two coherent branches; no tree evidence decides it. See D8. |

Both markers are nevertheless **open at audit time**, which is what MP-7 measures. MP-7 fails on
the presence of the markers, not on their quality.

---

## Answers to the dispatch's remaining direct questions

- **Is 84 still wrong?** No. 84 is correct, independently reproduced by both routes (per-file sum
  and per-pattern sum), together with the 84-occurrence / 81-line unit split. The lane's
  self-correction stands.
- **Does the M1 premise hold?** Yes — and it is stronger than the SPEC argues, because the
  landed 3-row table is exactly the derivation the premise predicts across all 9 documented-drop
  classes. M1's *conclusion* is right; M1 as a **milestone** is wrong (D1, D2).
- **Could the 3-vs-4 discrepancy have been resolved from evidence?** Yes, twice over (D1). And
  deferring it left M1 with an ambiguous entry state, now visible as three contradictory expected
  row counts (D4).
- **Does each milestone's verification discriminate?** M1: **no** — all three checks pass
  pre-work (D2). M2: partially — population pairing is good, correspondence is not asserted (D9).
  M3: **yes** — `Task* → 0` paired with `task-list ≥ 3` (measured baseline 0) closes the
  delete-instead-of-rewrite route, and both invariance criteria are measurable over non-empty
  populations (`invoke Skill(` = 41, `AskUserQuestion` = 4, both re-measured here). M4: **no** —
  the radius clause is unquantified (D5).
- **Would a blanket find-and-replace satisfy any AC?** No. AC-CBN-003 (41 invariant), AC-CBN-004
  (4 invariant) and AC-CBN-005's second condition each break under a blanket substitution. This
  is the SPEC's strongest design work and should be preserved verbatim through any revision.

---

## Recommendation

FAIL. Route the fixes in this order; the first three are what make a re-audit worth running.

1. **Resolve or withdraw the two `[NEEDS CLARIFICATION]` markers before Implementation Kickoff
   Approval** (MP-7). Marker ① should be withdrawn on the tree evidence in D1, not put to the
   operator; marker ② needs a genuine operator answer.
2. **Fold `measurement.md:198-` into `spec.md` §A.2 정정 2 and `plan.md` §B-2, and re-scope M1**
   from a Codex probe to the derivation-correction it has become (D1).
3. **Give M1 a check that can fail** — bind `capability-absence.md` and scope the `grep -c '^| '`
   selector to the table region (D2).
4. **Correct the `AskUserQuestion` coordinate set** in `spec.md` §B.4 and AC-CBN-002 to the three
   measured lines, and state the unit (D3).
5. **Unify the M1 expected row count to 4** at `plan.md:62` and `acceptance.md:105` (D4).
6. **Pin the M4 radius as a literal file set** and drop the root `.claude/agents/moai/*.md` entry
   (D5, D7); write REQ-CBN-008's path in full.
7. **Add an AC per remaining directive class** — `manager-design` fallback paragraph,
   `manager-lead` four coordinates (D6).
8. Optional, operator's discretion: D8 (carry the sampling Gap into §B.6), D9 (coordinate-set
   diff), D10 (make the cover-sentence check runnable and promote it to an AC).

Separately, and outside the SPEC: **report the mid-audit commit `53886ba40` to the lead** (D0).
An actively audited worktree has exactly one writer; the audit was reading an input that moved.

---

# Iteration 2

# SPEC Review Report: SPEC-CODEX-BODY-NEUTRALITY-001
Iteration: 2/2 (Tier M ceiling — final round, no third iteration available)
Verdict: **FAIL**
Overall Score: **0.85** (Tier M PASS threshold 0.80 — the score is above threshold; the FAIL is carried by two blocking correctness defects, not by the score)
Score trajectory: iter1 0.63 → iter2 0.85. **No regression — no STOP signal, no scope-reduction recommendation.**

Reasoning context ignored per M1 Context Isolation. Audited artifacts: `spec.md` v0.2.0, `plan.md`,
`acceptance.md` (Tier M input contract), plus the cited evidence tree.

## D0 follow-up — the audit window held

- **Claim.** No writer entered the tree during this audit; iteration-1 process defect PD-1 did not recur.
- **Evidence.**
  ```
  $ git rev-parse HEAD       (window open)   cb826d42ba41720433d82892830c30c4448dc6f7
  $ git rev-parse HEAD       (window close)  cb826d42ba41720433d82892830c30c4448dc6f7
  $ git status --short       (both)          <empty>
  $ git branch --show-current                WT-codex-neutrality
  ```
- **Baseline-attribution.** Both reads in this run, this tree.
- **Gaps.** None for this claim.
- **Residual-risk.** None.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — 16 definition lines, 16 unique ids, `REQ-CBN-001`..`REQ-CBN-016`, no gaps, no duplicates, uniform 3-digit padding. `grep -cE '^- \*\*REQ-CBN-[0-9]{3}\*\*' spec.md` → 16; the same piped through `sort -u | wc -l` → 16. Document order places `REQ-CBN-016` after `-005` (§C.2 grouping) — a presentation choice, not a numbering gap.
- **[PASS] MP-2 GEARS format compliance (requirement layer)** — judged against the 16 `REQ-CBN-XXX` entries in `spec.md` §C, **not** against the `AC-CBN-XXX` Given-When-Then entries in `acceptance.md` (verification layer, graded under Group 4). All 16 carry a valid pattern: ubiquitous (001, 003, 005, 008, 011, 014, 015, and 016's first limb), `When` (004, 006, 007, 009, 012, 013, and 016's second limb), `While` (010), `Where` (002). Advisory nit A6 below on `REQ-CBN-002`'s `Where` semantics.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types: `id`, `title` (quoted), `version: "0.2.0"` (quoted semver), `status: draft`, `created: 2026-09-07`, `updated: 2026-09-07`, `author`, `priority: P2`, `phase` (quoted), `module`, `lifecycle: spec-anchored`, `tags` (comma-separated string). No rejected snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id`). Optional `tier: M` and `related_specs` additionally carried. `moai spec lint .../spec.md` → `✓ No findings`, rc 0.
- **[N/A] MP-4 language neutrality** — the SPEC is scoped to one repository's agent-body / emitter tree and names no per-language tooling. Criterion does not apply; auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — 3 referenced SPECs, all present, all `status: completed`; none in {retired, superseded, archived}. No BLOCKING finding.
  ```
  SPEC-CODEX-DUAL-AGENTS-001      -> status: completed
  SPEC-CODEX-SKILL-NEUTRAL-001    -> status: completed
  SPEC-CODEX-SKILLS-CANONICAL-001 -> status: completed
  ```
  In the SPEC's favour: M1 amends a `completed` SPEC and plans an explicit Amendments HISTORY row for it (`plan.md` §F M1 산출 2) — that is the reconciliation practice D7 exists to require.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -rc 'syscall' .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/` → `0` on all four files. D8 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/` → no output, rc 1. Both iteration-1 markers are discharged: ① withdrawn (§A.2 correction 2), ② converted to an operator-decision record (`spec.md` §D last block, `plan.md` §E ②). **The canonical selector is clean — but see N1: the SPEC's own restatement of this check uses a different, self-falsifying selector.**

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 — minor ambiguity in one or two requirements | `REQ-CBN-009` (`spec.md:154`) names a closed set "manager-lead 자기 스폰 4줄" that `REQ-CBN-002`'s own subject rule contradicts (N2). Everything else — units, selectors, populations — is pinned to a single value with a command. |
| Completeness | 0.90 | 1.0 band, docked for N2's under-inclusive enumeration | All required sections present. `spec.md` §D carries **five** `### Out of Scope — <topic>` H3 sub-headings, each with specific `-` bullets (`spec.md:198,204,209,213,217`). Frontmatter complete. §C.5 supplies the literal 11-file radius the ACs compare against. |
| Testability | 0.75 | 0.75 / 0.50 boundary | Most ACs are binary with command + unit + measured RED baseline. Docked for N1 (a Definition-of-Done check that cannot pass as written), A2 (`[^/]design-sync` cannot match a line-initial occurrence), and `AC-CBN-013` being structurally unable to detect N2's shortfall. |
| Traceability | 1.00 | 1.0 | Every REQ has ≥1 AC and every AC maps to an existing REQ. Union of the 14 `maps REQ-…` declarations = `REQ-CBN-001..016`, exactly 16, no orphans, no uncovered REQ. |

Aggregate = (0.75 + 0.90 + 0.75 + 1.00) / 4 = **0.85**.

---

## Regression Check — iteration-1 defects D1-D10

Every finding re-verified mechanically against this tree. **All ten are RESOLVED.** No defect appears
unchanged across both iterations, so no stagnation flag.

| # | iter-1 finding | Status | Verifying measurement (this run, `cb826d42b`) |
|---|---|---|---|
| D1 | M1 deferred a question the baseline already answered | **RESOLVED** | Three strands independently reproduced — V1. M1 is now a documentation correction (`plan.md` §F M1: "프로브가 아니다"). |
| D2 | Vacuous M1 gate — passed on the untouched tree | **RESOLVED** | Three RED states confirmed — V2. |
| D3 | False coordinate `manager-develop.toml:64` | **RESOLVED** | `manager-develop.toml` `AskUserQuestion` occurrences = **0**; the three real coordinates confirmed — V3. |
| D4 | Three expected values for one outcome | **RESOLVED** | One selector, one value: **3** on both `AGENTS.md` copies; the banned selector yields 4 and appears only inside explicit prohibitions — V4. |
| D5 | Scope containment not pinned | **RESOLVED** | §C.5 is a literal 11-path set compared by `diff` (set equality); all 11 paths exist; evidence radius prefix-bounded — V5. |
| D6 | `REQ-CBN-009` under-covered (only `Task*` had an AC) | **RESOLVED** (with N2 as a new, narrower defect) | `AC-CBN-013` (`subagent-spawn`) and `AC-CBN-014` (`design-sync`) added, each positive against a measured-zero baseline — V6. |
| D7 | `REQ-CBN-008` named the wrong source path | **RESOLVED** | Path is now `internal/template/templates/.claude/agents/moai/*.md`; the citation is exact; the allowlist admits only one copy — V7. |
| D8 | Unverified premise not labelled | **RESOLVED** | `spec.md:129-131` now carries a `[HARD]` block naming the premise as unverified and making sample-confirmation a **precondition of the follow-up card**. |
| D9 | `AC-CBN-001` row count ≠ correspondence | **RESOLVED** | `AC-CBN-001` (c) now `diff`s the classification's `file:line` column against the coordinate set; the set is **81** distinct lines (measured). |
| D10 | Cover-sentence check was not a runnable command | **RESOLVED** | `grep -c '\.agents/skills' AGENTS.md` runs and returns **0** on both copies (measured), and is promoted into `plan.md` M3 ⑥⑦ as a paired check. |

---

## Verified-correct findings (adversarial checks the repair survived)

### V1 — §A.2 correction 2's three strands are all independently reproducible

- **Claim.** The "4th binding row" is `cross-session-messaging`, its rationale asserts capability **presence**, and an exhaustive pass over `tool_classes` yields exactly the 3 rows already shipped.
- **Evidence.**
  ```
  $ ls -l .moai/reports/t196/csn003-table-4row.txt
  -rw-r--r-- 373 .moai/reports/t196/csn003-table-4row.txt   # 373 B — the exact figure REQ-CSN-003 :266 cites
  $ tail -1 .moai/reports/t196/csn003-table-4row.txt
  | cross-session-messaging | SendMessage, ListAgents tools | use the moai MCP broker |

  $ sed -n '148,162p' internal/template/agentemit/agents-codex.yaml
    - class: cross-session-messaging
      rationale: >- ... The Codex counterpart rides the moai MCP
        broker (session_msg_register/list/send/poll) under the existing
        server-level moai-mcp grant ...

  $ sed -n '/^tool_classes:/,/^$/p' internal/template/agentemit/agents-codex.yaml \
      | grep -oE ': [a-z-]+$' | sed 's/^: //' | sort -u | wc -l
  11
  ```
  Applying the SPEC's discriminant (a rationale describing a counterpart ⇒ capability present) to all
  11 values yields absent = {`task-list` "no known Codex equivalent", `design-sync` "no Codex
  equivalent", `question-channel` "codex reported it unavailable"} — **exactly the 3 rows in
  `AGENTS.md`**. `file-read` / `file-write` / `shell` / `web` / `skill-loader` / `subagent-spawn` /
  `cross-session-messaging` / `moai-mcp` all read as present.
- **Baseline-attribution.** Every command above run in this tree, this run.
- **Gaps.** I did not probe Codex runtime behaviour either; like the SPEC, this is a rationale-text judgement.
- **Residual-risk.** A stale rationale would err toward **fewer** rows — conservative, as the SPEC states.

### V1a — the M1 enumeration command actually works

The `plan.md` §F M1 command that extracts the population is executable as written and returns the
full 11-value set (output above). A broken extraction command would have made `AC-CBN-006 (a) = 11`
unreachable; it is reachable.

### V2 — M1's checks genuinely fail on the untouched tree, and the fourth condition is a real equality

```
$ ls .moai/reports/t497/capability-absence.md
No such file or directory                                    # (a)(b) RED
$ grep -c '현재 측정값 4행' .moai/specs/SPEC-CODEX-SKILL-NEUTRAL-001/spec.md
1                                                            # (d) RED — expected 0
$ grep -c '현재 측정값 3행' .moai/specs/SPEC-CODEX-SKILL-NEUTRAL-001/spec.md
0                                                            # plan.md ⑤ RED — expected 1
```
All three lane-reported RED states reproduce. The dispatch's specific question — whether the
absent-count / table-row-count relation is asserted as an **equality** or as two loose constants —
resolves in the SPEC's favour: `acceptance.md:81` reads "(a) = **11**, (b) = **3**, (c) = **3**,
**(b) == (c)**, (d) = **0**", and `plan.md` ③ reads "**3**, 그리고 ② 와 같은 값", with `plan.md:89`
requiring both be measured in the same judgement. The equality is explicit and **additional** to the
constants, so a coordinated drift of both sides cannot slip through.

### V3 — the rebuilt coordinate table is correct in every cell

```
$ grep -rn 'AskUserQuestion' .../*.toml | wc -l    → 3   (distinct lines)
$ grep -rho 'AskUserQuestion' .../*.toml | wc -l   → 4   (occurrences)
$ grep -rco 'AskUserQuestion' .../*.toml | grep -v ':0'
  plan-auditor.toml:2   super-advisor.toml:1   sync-auditor.toml:1
```
`manager-develop.toml` carries **0** — the iteration-1 false coordinate is gone. Reading the three
lines verbatim confirms each `subject` value:

- `sync-auditor.toml:131` — "no `sync-auditor` path invokes `AskUserQuestion` or `mcp__askuser`" → `prohibition`. Correct; neither `this-agent` nor `orchestrator` fits.
- `plan-auditor.toml:146` — "**The orchestrator MUST** resolve each marked topic via `AskUserQuestion` (preload `ToolSearch(query: "select:AskUserQuestion")`)" → `orchestrator`, **2 occurrences on one line**. Correct.
- `super-advisor.toml:62` — read with `:59-62`: "the orchestrator spawns … then either re-seeds the executor … or escalates to the user via `AskUserQuestion`" → `orchestrator`. Correct.
- `manager-develop.toml:64,65,66` — three `Agent(general-purpose)` destination cells in the delegation routing table. Correct as `Agent(` axis / `orchestrator`.

The introduction of the third `subject` value `prohibition` is justified: forcing `sync-auditor:131`
into either of the other two would make the row false.

**New `.md` coordinates (never seen by iteration 1) all verified exactly:**
```
manager-develop.md:103,128   → the two Task* directive lines                     ✓
e2e-tester.md:146            → "tracked via TaskCreate/TaskUpdate"                ✓
manager-design.md:115        → "(1) default = DesignSync tool push"               ✓
manager-lead.md:44,64,66,200 → the four Agent( lines, +7 offset from toml 37,57,59,193 ✓
```
The `.md`↔`.toml` offset is uniform per file (+7 for `manager-lead`, `e2e-tester`, `manager-design`;
+13 for `manager-develop`), consistent with frontmatter stripping — the coordinate pairs are not guesses.

### V4 — one selector, one value; no stale 4 or 5 survives

```
$ sed -n '/^\*\*Capability bindings/,/^---$/p' AGENTS.md | grep -c '^| [a-z]'                             → 3
$ sed -n '/^\*\*Capability bindings/,/^---$/p' internal/template/templates/AGENTS.md | grep -c '^| [a-z]' → 3
$ grep -c '^| ' AGENTS.md                                                                                  → 4  (the banned selector)
```
The banned selector appears in the SPEC only inside explicit prohibitions (`acceptance.md:17`,
`plan.md:62`) and in the RED-baseline narration at `acceptance.md:82`. No occurrence of `4` or `5`
survives as an *expected value* anywhere in the four artifacts. Advisory A5 concerns the *reason*
given for the ban, not the ban itself.

### V5 — the radius is a real 11-file set, template-side, and prefix-bounded

All 11 paths in `spec.md` §C.5 exist (checked individually). The set names the **template-side**
copies for all four agent bodies and both `AGENTS.md` copies; the repo-root `.claude/agents/moai/*.md`
is absent from it, and `acceptance.md:126` makes its appearance in `git status` an explicit failure.
The evidence radius is bounded by **prefix**, not count — the right shape for a directory whose file
count is not determinable in advance.

Structurally confirmed complete for the emit path: the goldens **are** the committed
`templates/.codex/agents/moai/*.toml`, not a separate fixture tree — `golden_test.go:1-12` ("pin the
committed artifacts under templates/.codex/agents/moai/"). So `AGENTEMIT_UPDATE=1` cannot dirty a
path outside the enumerated set.

### V6 — the new per-class ACs are all positive against a genuinely-zero baseline

```
$ grep -rn 'task-list'       .../*.toml                     | wc -l  → 0   (AC-CBN-005 second limb)
$ grep -c  'subagent-spawn'  .../manager-lead.toml                   → 0   (AC-CBN-013)
$ grep -oE '[^/]design-sync' .../manager-design.toml | wc -l         → 0   (AC-CBN-014)
$ grep -o  'design-sync'     .../manager-design.toml | wc -l         → 4
$ grep -c  'default = DesignSync tool push' .../manager-design.toml  → 1
$ grep -c  '\.agents/skills' AGENTS.md ; ... templates/AGENTS.md     → 0 ; 0
```
The dispatch's specific question on `AC-CBN-014` resolves in the SPEC's favour: all four existing
`design-sync` occurrences are indeed `/design-sync` slash commands (lines 25, 62, 77, 184 of
`manager-design.toml`, each preceded by `/`), so a bare `grep -c 'design-sync'` would return 4 and be
vacuous, and the `[^/]` guard correctly reads **0** today. See A2 for the guard's one hole.

### V7 — `REQ-CBN-008`'s path and its citation are both exact

```
$ grep -n 'templatesDir\|agentMDRoot' internal/template/agentemit/golden_test.go
31: const templatesDir = "../templates"
34: const agentMDRoot  = ".claude/agents/moai"
```
The cited range `golden_test.go:31-34` is correct to the line. The M4 allowlist no longer admits both
copies: `spec.md` §C.5 lists only the template-side bodies, `spec.md` §D moves the repo-root copy to
Out of Scope with the single justified exception (root `AGENTS.md`, edited **together** with the
template copy under `REQ-CBN-006`), and `acceptance.md:126` turns a root-copy appearance into a
failure. D7 is fully discharged.

### V8 — the carry-forward invariants are intact; nothing was weakened to make a check pass

Every iteration-1 figure independently re-measured in this run:
```
$ find internal/template/templates/.codex/agents -name '*.toml' | wc -l            → 11
$ grep -rhoE '<union>' .../*.toml | wc -l                                           → 84   (occurrences)
$ grep -rnoE '<union>' .../*.toml | cut -d: -f1-2 | sort -u | wc -l                 → 81   (distinct lines)
$ grep -rhoE 'invoke Skill\(' .../*.toml | wc -l → 41    $ grep -rhoE 'Skill\('  ... → 45
$ grep -rhoE 'Agent\('        .../*.toml | wc -l → 25    $ grep -rhoE 'DesignSync' ... → 6
$ grep -rhoE 'Task(Create|Update|List|Get)' .../*.toml | wc -l                      → 4
$ grep -rlE '<union>' internal/template/templates/.claude/skills/ | wc -l           → 77
```
`45 + 25 + 4 + 6 + 4 = 84` — the pattern table and the file-wise total agree, and the 84-occurrence /
81-line split is exactly as stated.

The **field-vs-capability premise** is intact and, importantly, applied **consistently**:
`agents-codex.yaml` uses the same "no agent-TOML carrier, counterpart exists" shape for
`cross-session-messaging`, `skill-loader` and `subagent-spawn`, and the SPEC classifies all three as
present. The **anti-blanket-replace design** is intact and unweakened: `AC-CBN-003` (`41` invariant),
`AC-CBN-004` (`4` invariant, explicitly written "shrink ⇒ fail"), `AC-CBN-005` (`Task*→0` **paired**
with `task-list ≥ 3` coordinates). Nothing was relaxed to make a check pass.

### V9 — the `ParseFailure` claim is a schema property, confirmed against a sibling completed SPEC

```
$ moai spec lint .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/spec.md
✓ No findings — all SPEC documents are valid                                   rc 0

$ moai spec lint .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/plan.md
ERROR ParseFailure ... YAML frontmatter missing or does not start with '---'
$ moai spec lint .moai/specs/SPEC-CODEX-SKILL-NEUTRAL-001/plan.md         # completed sibling
ERROR ParseFailure ... YAML frontmatter missing or does not start with '---'
$ moai spec lint .moai/specs/SPEC-CODEX-SKILL-NEUTRAL-001/acceptance.md   # completed sibling
ERROR ParseFailure ... YAML frontmatter missing or does not start with '---'
```
A `completed` sibling produces the byte-identical error on both files. The linter expects
frontmatter; `plan.md` / `acceptance.md` carry none by convention. **Claim discharged — not a defect.**

### V10 — M4's `awk '{print $NF}'` filter: reported as reasoned, now actually observed

The repair did not execute this. I did, against synthetic `git status --short` entries covering the
shapes that will occur:
```
$ printf ' M AGENTS.md\n?? .moai/reports/t497/x.md\nR  old/a.md -> internal/template/templates/AGENTS.md\nMM .../manager-lead.md\n?? internal/newpkg/\n?? "\353\260\224\355\214\214.md"\n?? my file.md\n' | awk '{print $NF}'
AGENTS.md
.moai/reports/t497/x.md
internal/template/templates/AGENTS.md      ← rename: yields the NEW path (correct for radius comparison)
internal/template/templates/.../manager-lead.md
internal/newpkg/                           ← untracked dir: one entry, trailing slash
"\353\260\224\355\214\214.md"              ← quoted non-ASCII: survives as one mangled token
file.md                                    ← space-bearing path: MANGLED
```
**Assessment: correct for the shapes that will actually occur, and fail-safe where it is not.**
Renames resolve to the new path, which is what set-equality needs. The two failure modes
(space-bearing and non-ASCII paths) both produce an *extra unexpected line* in the `diff` — a false
FAIL, never a false PASS — and the enumerated radius is entirely ASCII and space-free. The one shape
that could hide a path is an untracked **directory** collapsing its contents into a single entry, but
that only matters inside the two prefix-filtered evidence directories, where an unbounded file count
is the intended design. No defect; recorded so the reasoning is now attributed to a measurement.

### V11 — the `AGENTS.md` cover is nowhere near either ceiling

The dispatch asks whether the added text is bounded ex ante or only checked post hoc. It is checked
post hoc (`AC-CBN-009`) — and both ceilings were measured to see whether that matters:
```
$ go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v
    always-loaded surface = 74535 tokens (budget 77600, headroom 3065, 17 entries)
--- PASS
$ wc -c AGENTS.md internal/template/templates/AGENTS.md
14774 AGENTS.md      14774 internal/template/templates/AGENTS.md
$ grep -n 'CodexContractByteCeiling =' internal/config/token_budget_guard.go
95: const CodexContractByteCeiling = 24576
```
Headroom is **3,065 tokens (~12 KB)** on the always-loaded budget and **9,802 B** on the Codex
contract byte ceiling, against an addition the SPEC scopes to **one sentence × two copies**
(`spec.md:95`, `REQ-CBN-011`). A post-hoc invariant is proportionate at this margin. See A4 for the
un-named second guard.

### V12 — the cover does not smuggle back a forbidden row

`REQ-CSN-003`'s doctrine constrains the binding table's **row set**; the cover is one sentence in the
surrounding paragraph, adds no first-column value, and `AC-CBN-008` checks first-column values
against `tool_classes` — so `AC-CBN-006 (b)==(c)` and `AC-CBN-008` both remain honest. The cover
records an **invocation-form mapping** (`Skill("<name>")` ↔ `.agents/skills/<name>/SKILL.md`), not an
absence fallback, which is consistent with §B.1's own field-vs-capability discriminant. **Not a
smuggled row.** The residual tension is recorded as advisory A7, not as a defect.

### V13 — `REQ-CBN-016`'s merge is constraint-driven and its coverage did not thin

The Tier M ceiling the merge was made to respect is real:
`.claude/rules/moai/workflow/spec-workflow.md:146-150` — Tier M requirement ceiling **16**,
acceptance-criterion ceiling **16**, applied independently. The SPEC carries 16 REQs and 14 ACs —
both within budget. `REQ-CBN-016` is a compound (an exhaustive-record obligation plus a
`When`-guarded correction obligation), so it is not perfectly atomic — **but both limbs are covered**:
`AC-CBN-006` tests `(a) = 11` for the record and `(d) = 0` for the correction, in the same judgement.
Coverage did not thin. Recorded as advisory A3 for the atomicity nit and the at-ceiling signal only.

---

## Defects Found (structured defect-list)

### N1. The SPEC's own marker-verification command is self-falsifying, and it is a Definition-of-Done item · `spec.md:235`, `plan.md:50`, `acceptance.md:173` · Severity: **major** · Class: **blocking**

- **Claim.** All three artifacts state that `grep -rn 'NEEDS' .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/` returns **무출력** (no output). On the tree the repair was written against it returns **3 lines** — and every one of them is the SPEC's own text making that claim. `acceptance.md:173` promotes this to a **Definition of Done** item, so the card's DoD is unsatisfiable as written.
- **Evidence.** Run verbatim, this tree, this run:
  ```
  $ grep -rn 'NEEDS' .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/
  .moai/specs/.../acceptance.md:173:- 미해결 마커 **0건** — `grep -rn 'NEEDS' .moai/specs/...
  .moai/specs/.../plan.md:50:검증: `grep -rn 'NEEDS' .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/` → **무출력**...
  .moai/specs/.../spec.md:235:- 미해결 마커: **0건.** ... 검증: `grep -rn 'N...
  rc=0    lines=3
  ```
  The canonical marker check is clean, which is why MP-7 passes:
  ```
  $ grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/
  (no output)   rc=1
  ```
- **Baseline-attribution.** Both commands run in this tree at `cb826d42b` in this run.
- **Gaps.** I did not check whether the same selector appears in `progress.md` or in the t497 reports; the three occurrences above are the ones inside the SPEC artifact set the DoD names.
- **Residual-risk.** Beyond the unsatisfiable DoD, the failure is self-reinforcing: the only way to make the stated command return no output is to delete the sentences that assert it. A run-phase actor running the DoD check verbatim sees a RED that looks like surviving markers and may hunt for markers that do not exist.
- **Why blocking, not advisory.** This is the anti-pattern the SPEC itself names — `plan.md` §G **AP-6 자기참조 수치** ("이 SPEC 자신의 산출물을 주어로 삼는 계수는 문서를 편집하는 행위가 곧 무효화한다") — applied to the SPEC's own self-verification, and it is an unobserved-verification claim asserted three times (`verification-claim-integrity.md` §1). It is **new in the repair**: at v0.1.0 the markers existed, so the command legitimately returned hits and no "무출력" claim was made.
- **Required fix.** Replace the selector in all three places with the marker-shaped one and restate the observed result: `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/` → no output, rc 1. Three single-line edits; no design change.

### N2. `REQ-CBN-009` fixes the `manager-lead` directive class at a hand-enumerated 4 lines that its own subject rule contradicts · `spec.md:113,154`, `acceptance.md:138`, `plan.md:115` · Severity: **major** · Class: **blocking**

- **Claim.** `REQ-CBN-009` binds "`manager-lead` 자기 스폰 **4줄** (`subagent-spawn`)" as a closed set, and `AC-CBN-013` pins exactly `manager-lead.toml:37,57,59,193`. `manager-lead.toml` carries **10** `Agent(` lines, and at least one of the six unlisted lines satisfies `REQ-CBN-002`'s own discriminant for `verdict=directive` / `subject=this-agent`. The enumeration is under-inclusive, and no AC can detect the shortfall.
- **Evidence.**
  ```
  $ grep -c 'Agent(' internal/template/templates/.codex/agents/moai/manager-lead.toml
  10
  $ grep -n 'Agent(' ... | cut -d: -f1 | tr '\n' ' '
  7 23 29 37 57 59 130 172 193 261
  ```
  Compare the listed `:57` with the unlisted `:172`:
  ```
  :57  - **Peer cross-validation orchestration** — when a leaf worker marks an AC PASS at Tier M/L,
         manager-lead spawns a second read-only `Agent(general-purpose)` ...      [SPEC: directive]
  :172 At Tier M/L milestones, every AC the author leaf worker marks PASS is re-run by a second
         read-only `Agent(general-purpose)`:                                      [SPEC: unlisted]
  ```
  These are the same behaviour — `:57` states it in the capability list, `:172` states it as the
  procedure this agent performs. Under `REQ-CBN-002` ("Where a line's directive subject is the
  orchestrator rather than this agent → prose"), `:172`'s subject is this agent, so it is a
  directive. `:261` ("Domain consultation … → leaf worker as `Agent(general-purpose)` with domain
  whitelist") is a routing instruction to this agent and is arguably a second such line.
- **Baseline-attribution.** All coordinates read from `manager-lead.toml` in this tree, this run.
- **Gaps.** I did not adjudicate all six unlisted lines; `:7`, `:23`, `:29`, `:130` read to me as descriptive prose and I make no claim about them. My finding rests on `:172` alone, with `:261` flagged as probable.
- **Residual-risk.** `AC-CBN-013` passes at `grep -c 'subagent-spawn' manager-lead.toml ≥ 4`, so revising exactly the enumerated four turns the AC green while a genuine directive line keeps instructing a Codex-driven harness to spawn a Claude subagent — precisely the failure this SPEC exists to prevent. `M2`'s 84-row classification, applying `REQ-CBN-002`, would then contradict `REQ-CBN-009` inside the same deliverable.
- **Why blocking.** It is an internal inconsistency between two of the SPEC's own requirements, and it is the same defect *class* as iteration-1's D3 — a coordinate set asserted as complete without stating the discriminant that closes it. `§B.4` is honestly labelled "경계 판정 4건" (a boundary sample), but `REQ-CBN-009` and `AC-CBN-013` then treat that sample as the population.
- **Required fix (either is sufficient).** (a) Restate `REQ-CBN-009`'s second class **by property** rather than by coordinate — "every `Agent(` line in `manager-lead` that M2 classifies `directive` / `this-agent`" — and rebind `AC-CBN-013` to compare `subagent-spawn` coordinates against M2's directive rows for that file rather than against the constant 4. Or (b) keep the closed set but adjudicate `:172` and `:261` explicitly in `§B.4` with the reason they are prose, citing the measured `grep -c 'Agent(' → 10` so the reader sees the 10→4 narrowing was performed rather than assumed.

### A1. `moai-mcp` is an unflagged trap in M1's derivation — its rationale contains "unavailable" while the capability is present · `spec.md:145` (`REQ-CBN-004`), `plan.md:68` · Severity: minor · Class: **optional**

- **Claim.** M1 must classify 11 classes; exactly one, `moai-mcp`, carries the word `unavailable` in a rationale that nonetheless describes a **present** capability. A derivation keying on the word rather than on the capability/field discriminant yields `(b) = 4`, and `AC-CBN-006`'s `(b) == (c)` then fails.
- **Evidence.**
  ```
  - class: moai-mcp
      Agents carrying any mcp__moai__* token declare the server-level grant ...
      Per-tool filtering inside one MCP server is unavailable — documented drop; ...
  ```
  What is unavailable is per-tool filtering; the MCP capability itself is granted at server level. Under `REQ-CBN-004`'s discriminant this is **present** — consistent with `cross-session-messaging`, `skill-loader`, `subagent-spawn`.
- **Baseline-attribution.** `agents-codex.yaml`, this tree, this run.
- **Gaps.** None — this is the only class where the two readings diverge.
- **Residual-risk.** A run-phase actor producing `capability-absence.md` hits one genuinely hard call and the SPEC names none. This SPEC names traps everywhere else; this one is missing.
- **Suggested fix.** One line in `spec.md` §A.2 correction 2 or `plan.md` M1: `moai-mcp` reads `present` — "unavailable" there qualifies per-tool filtering, not the capability.

### A2. `[^/]design-sync` cannot match a line-initial occurrence · `acceptance.md:145`, `plan.md:126`, `spec.md:123` · Severity: minor · Class: **optional**

- **Claim.** `grep -oE '[^/]design-sync'` requires one non-`/` character **before** the token. A paragraph opening a line with a bare `design-sync` produces no match, so `AC-CBN-014` reads 0 and fails although the required paragraph exists.
- **Evidence.** The regex carries no start-of-line alternation; the four existing occurrences all sit mid-line preceded by `/` (measured: bare 4, guarded 0).
- **Baseline-attribution.** `manager-design.toml`, this tree, this run.
- **Gaps.** I did not write a synthetic line-initial fixture into the file (the tree is read-only during the audit).
- **Residual-risk.** Fail-safe direction — the hole produces a **false FAIL**, never a false PASS, so it costs a run-phase round trip rather than admitting a defect.
- **Suggested fix.** `grep -oE '(^|[^/])design-sync'`, keeping the same expected value.

### A3. `REQ-CBN-016` is compound, and the SPEC sits exactly at the Tier M requirement ceiling · `spec.md:147` · Severity: minor · Class: **optional**

Two obligations in one requirement (produce an exhaustive derivation record; correct the mismatched
text). Both limbs are covered by `AC-CBN-006` `(a)` and `(d)`, so testability holds and coverage did
not thin — but 16/16 against a ceiling whose own rule says exceeding it "is a signal to tier up or to
split the SPEC" means the next requirement has nowhere to go. Worth stating in the SPEC that the
merge was budget-driven, so a future editor does not split it back and silently breach the ceiling.

### A4. `TestCodexContractByteCeiling` is never named, and the pinned verification scope would not run it · `acceptance.md:99-103,132`, `spec.md:149` (`REQ-CBN-007`) · Severity: minor · Class: **optional**

`AC-CBN-009` covers only `TestAlwaysLoadedTokenBudget`; `AC-CBN-012` restricts run-phase verification
to `./internal/template/agentemit/...` and `./internal/config/`, and `AC-CBN-009`'s
`-run 'TestAlwaysLoadedTokenBudget$'` filter excludes the byte guard. Both `AGENTS.md` copies are
covered by that guard (`token_budget_guard_test.go:140-143` exercises live and mirror), and it
**fails the build** rather than warning. Measured headroom **9,802 B** against a one-sentence
addition, so this is not a live hazard — but the guard belongs in `REQ-CBN-007`'s scope by name.

### A5. `acceptance.md:17`'s reason (a) for banning `grep -c '^| '` is false on this tree · `acceptance.md:17` · Severity: minor · Class: **optional**

The ban is right; one of its two reasons is not. `acceptance.md:17` says the banned selector
"(a) counts `| ` lines outside the binding table and (b) includes the header". Measured: all four
matches are inside the table.
```
$ grep -c '^| ' AGENTS.md → 4
$ sed -n '/^\*\*Capability bindings/,/^---$/p' AGENTS.md | grep -c '^| ' → 4
$ grep -n '^| ' AGENTS.md → 19 (header), 21, 22, 23 (data)
```
The whole discrepancy is (b). Reason (a) is a hypothetical presented as an observation — a right
verdict with a wrong argument, in a document whose entire method is attributing claims to measurements.

### A6. `REQ-CBN-002` uses `Where` for a state condition, not a capability gate · `spec.md:140` · Severity: minor · Class: **optional**

GEARS `Where` denotes a capability gate / feature flag / static config; legacy EARS `Where` denotes an
optional feature. `REQ-CBN-002`'s condition ("a line's directive subject is the orchestrator rather
than this agent") is a per-item state, which `When` or `While` carries. The pattern is syntactically
valid so MP-2 passes; the modality is semantically loose.

### A7. Two "field-absent, capability-present" classes get two different dispositions with no stated discriminant · `spec.md:85,89,95` vs `spec.md:154`, `acceptance.md:139` · Severity: minor · Class: **optional**

`skill-loader` and `subagent-spawn` are diagnosed identically in §B.1 (the missing thing is the
agent-TOML field, not the capability). `skill-loader` is then disposed of by leaving 41 body lines
untouched and adding one `AGENTS.md` cover sentence; `subagent-spawn` by rewriting four body lines in
place. §B.2 gives a reason for the first (progressive-disclosure savings) but the SPEC never states
the rule that decides "cover sentence" vs "body rewrite", so a follow-up card has no principle to
apply to the remaining classes.

### A8. A third stale "4행" claim in the sibling SPEC is neither corrected nor exempted · `SPEC-CODEX-SKILL-NEUTRAL-001/spec.md:28` · Severity: minor · Class: **optional**

M1 names one line to correct (`:280`) and one to leave alone (`:266`, a byte-size record). **The
repair is right not to touch `:266`** — `csn003-table-4row.txt` measures 373 B, exactly the figure
cited there, so the two are correctly distinguished and **no conflation occurred**. A third
occurrence at `:28` ("오늘의 4행은 결과이지 기준이 아니다") sits in that SPEC's HISTORY. Leaving a
HISTORY entry alone is defensible — history records past state — but the SPEC enumerates two
occurrences as though they were all of them. One sentence saying `:28` is HISTORY and stays would
close the enumeration.

---

## Recommendation

**FAIL**, at score 0.85 against a Tier M threshold of 0.80. The score is above threshold and **all
seven must-pass criteria are PASS or N/A**; the FAIL is carried entirely by the two blocking
findings, and one of them (N1) makes the card's own Definition of Done unsatisfiable as written.

**This is a narrow, mechanical fix delta — not a scope or design problem.** The repair is otherwise
of high quality: all ten iteration-1 defects are discharged, each verified by re-running the
measurement rather than reading the assertion, and none of the carry-forward invariants (84/81, the
field-vs-capability premise, the anti-blanket-replace AC pairs) was weakened to make a check pass.
Score moved 0.63 → 0.85, so no STOP signal and no scope-reduction recommendation.

Since no third iteration is available:

**Blocking — fix before the card leaves plan-phase (both are edit-in-place, no redesign):**

1. **N1** — replace the marker selector with `grep -rn '\[NEEDS CLARIFICATION' …` and restate the observed result (no output, rc 1) in `spec.md:235`, `plan.md:50`, `acceptance.md:173`. Three single-line edits.
2. **N2** — either restate `REQ-CBN-009`'s `manager-lead` class **by property** and rebind `AC-CBN-013` to M2's directive rows for that file, or adjudicate `manager-lead.toml:172` and `:261` explicitly in `§B.4` and cite the measured `grep -c 'Agent(' → 10`.

After those two land, the SPEC meets the bar: I found no third blocking defect, and the must-pass
firewall is already clean. A confirming re-read scoped to exactly these two edits is sufficient — a
full third audit is not warranted by the evidence.

**Advisory — operator's discretion, safe to carry as documented debt (A1-A8).** A1 (the `moai-mcp`
trap) and A2 (the `[^/]` line-initial hole) are the two most likely to cost a run-phase round trip
and are each a one-line change; A3-A8 are documentation-quality items with no effect on whether the
card can land. Per M6 this list does not by itself justify the FAIL and must not be routed as if it did.

**Process.** No PD-1 recurrence: `HEAD` was `cb826d42ba41720433d82892830c30c4448dc6f7` with a clean
tree at both the opening and closing measurement of this audit. The window held.

---

## Iteration 3 (delta — N1/N2 only)

**SPEC**: SPEC-CODEX-BODY-NEUTRALITY-001 · v0.2.1
**Tree**: `.claude/worktrees/t497` · branch `WT-codex-neutrality` · HEAD `b6442e848` (opening and closing read identical; tree clean both times — the audit window held)
**Scope**: N1 and N2 only, per the lead's grant. D1-D10 are RESOLVED by the iteration-2 verdict and are not re-opened. The carried-forward invariants (84-occurrence / 81-distinct-line unit split, field-vs-capability premise, `Task*→0` paired with `task-list ≥3`) are not re-litigated. A3-A8 remain recorded advisories.
**Verdict (scoped to N1 + N2)**: **PASS**. Both blocking findings are closed. Two minor residuals are recorded below as advisories and carry no FAIL.

Reasoning context ignored per M1 Context Isolation. The dispatch's lane-measured figures were treated as claims to reproduce, never as evidence — every number below was re-measured in this run.

### N1 — CLOSED

- **Claim.** The three marker-verification sites (`spec.md` §F, `plan.md` §E, `acceptance.md` §F) no longer assert a result their own command contradicts, and the replacement selector is not vacuous.
- **Evidence.** All run in this tree, this run:

  ```
  $ grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/
  (no output)   rc=1
  $ grep -rnE '\[NEEDS[[:space:]]CLARIFICATION' .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/
  (no output)   rc=1
  $ grep -rn 'NEEDS' .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/
  plan.md:50 · spec.md:22 · spec.md:245 · acceptance.md:185     rc=0   lines=4
  ```

  The three verification sites now read the ERE marker selector → **무출력, rc 1** (`plan.md:50`, `spec.md:245`, `acceptance.md:185`), and each states in the same sentence that the broad `'NEEDS'` form matches itself. **No artifact asserts that the broad form returns no output any more** — the fourth `NEEDS` line, `spec.md:22`, is the v0.2.1 HISTORY entry recording the repair. The DoD item (`acceptance.md:185`, inside §F 완료 정의) is satisfiable as written.
- **Non-vacuity — my own positive control**, independent of the lane's. Two fixtures written to the session scratchpad (outside the audited tree): one carrying `[NEEDS CLARIFICATION: which emitter owns the mirror?]`, one carrying a tab-separated `[NEEDS<TAB>CLARIFICATION: tabbed]`.

  ```
  ERE selector        → 2 hits (both fixtures), rc=0
  canonical selector  → 1 hit  (space fixture only), rc=0
  ```

  The ERE matches a genuine space-separated marker **and** a tab-separated one; the canonical BRE matches only the space form. The repair's selector is a strict **superset** of the canonical one for this pattern — the substitution cannot introduce a false negative. The green is not vacuous.
- **AP-6 sweep (dispatch item 4) — verified, with one omission.** Self-targeting commands naming this SPEC's own directory, measured this run:

  | site | form | self-falsifying? |
  |---|---|---|
  | `plan.md:50` · `spec.md:245` · `acceptance.md:185` | ERE marker selector | no — written form cannot match itself |
  | `plan.md:144` · `acceptance.md:117` | changed-path listing filtered by two exclusions, on the SPEC dir and on `.moai/reports/t497/` | no — excludes both evidence dirs |
  | `spec.md:247` | `ls -d` on this SPEC's own directory, asserting `No such file or directory` | **stale by construction** — see A9 |

  The coordinate the dispatch reported as `plan.md:142` measures at **`plan.md:144`** in this tree (`acceptance.md:117` is exact). The two exclusion-filtered commands do what the repair claims. `spec.md:247` was not in the repair's enumeration.
- **Baseline-attribution.** All commands run at HEAD `b6442e848`, this tree, this run. Positive-control fixtures written outside the audited tree, so the SPEC directory was not mutated.
- **Gaps.** I did not sweep `progress.md` or the other `.moai/reports/t497/` files for the same selector class — the DoD names the SPEC artifact set, which is what I swept. I did not verify that `[NEEDS CLARIFICATION` is the only marker spelling the project convention permits; I took the convention as given.
- **Residual-risk.** A marker wrapped **across** two lines would evade both selectors equally. That risk predates the repair and is not introduced by it.

### N2 — CLOSED

- **Claim.** `REQ-CBN-009`'s `manager-lead` class is now defined by property rather than by coordinate, `AC-CBN-013` is bound to a measured population `N` rather than to the constant 4, and the §D skip-permission is removed.
- **Evidence — property restatement.** `spec.md:164` (`REQ-CBN-009`) now binds "**M2 분류표가 `manager-lead.toml` 의 `Agent(` 줄 중 `verdict = directive` · `subject = this-agent` 로 판정한 줄 전부**". `spec.md:121-125` (§B.4) states the population, labels the four-coordinate table a **경계 표본** (boundary sample) rather than the population, names the closing discriminant as `REQ-CBN-002` applied at M2, and cites the measurement. Re-measured this run:

  ```
  $ grep -c 'Agent(' internal/template/templates/.codex/agents/moai/manager-lead.toml
  10
  $ grep -n 'Agent(' … | cut -d: -f1
  7 23 29 37 57 59 130 172 193 261
  $ grep -c 'subagent-spawn' …/manager-lead.toml
  0
  ```

  Population 10 and the coordinate list reproduce exactly; the pre-work baseline `M = 0` reproduces, so condition (c) is a real RED.
- **Evidence — the N selector is non-vacuous and fails loudly.** `AC-CBN-013` (`acceptance.md:134-151`) defines
  `N = grep -cE '^\| [^|]*manager-lead\.toml \| [0-9]+ \| Agent\( \| directive \|' …/body-classification.md`.
  The M2 column order pinned at `plan.md:93` is `file · line · token · verdict · subject · rationale`, which the selector matches positionally. Against a synthetic M2 table written this run with 5 `manager-lead.toml` / `Agent(` / `directive` rows plus 2 decoys (one `prose` row, one `manager-develop.toml` row):

  ```
  N = 5     both decoys correctly excluded
  N = 0     same table with the token cell backtick-quoted instead of bare
  ```

  The notation dependency the AC discloses is real, and it **fails in the safe direction**: a notation mismatch drives N to 0, breaking condition (a) `N ≥ 4` loudly rather than producing a false green. The `[HARD]` token-column pin at `plan.md:95` is what makes it dependable, and the AC states the failure direction itself.
- **Evidence — source↔TOML correspondence (dispatch item 3), re-verified independently.**

  ```
  $ grep -n 'Agent(' .claude/agents/moai/manager-lead.md | cut -d: -f1
  6 30 36 44 64 66 137 179 200 268
  ```

  Positional pairing against the TOML list gives `44↔37 · 64↔57 · 66↔59 · 179↔172 · 200↔193 · 268↔261`, exactly as `spec.md:125` claims. Content is byte-identical at both flagged pairs:

  ```
  toml:172 == md:179   At Tier M/L milestones, every AC the author leaf worker marks PASS is re-run by a
                       second read-only `Agent(general-purpose)`:
  toml:261 == md:268   - Domain consultation (backend / frontend / devops) → leaf worker as
                       `Agent(general-purpose)` with domain whitelist per `archived-agent-rejection.md` §C rows 7-10.
  ```

  And `toml:57` ("manager-lead **spawns** a second read-only `Agent(general-purpose)`") is the same behaviour `:172` restates as procedure — iteration 2's observation still holds on the tree, which is why flagging it is the correct disposition.
- **Evidence — §D permission removed.** A search for the permitting phrase returns exactly one hit, `acceptance.md:167`, and there it appears **inside a quotation of its own removal**: "앞 라운드의 이 항목은 「M1 이 행을 만들지 않았으면 건너뛴다」로 읽혀 … **그 허용을 제거한다**". The permitting form no longer exists as an instruction. Dispatch item 4 confirmed.
- **Assessment of the deferral (dispatch item 2) — sound, not a hole.** The literal answer to "can `AC-CBN-013` pass while `:172` is misclassified as prose?" is **yes**: N would be 4, (a) passes, (b) `M ≥ 4` passes. But that is a *judgment* risk, not a *coverage* gap, and the two are not interchangeable:
  - Coverage is closed mechanically. M2 ③ (`plan.md:102`) diffs the classification table's `file:line` column against the measured 81-line coordinate set, rc 0 required. No `Agent(` line in this file can remain unclassified.
  - The discriminant is stated (`REQ-CBN-002`), the population is measured (10), and the closing site is named (M2).
  - `:172` and `:261` are explicitly named at `spec.md:125` and `plan.md:117` as lines M2 **must** adjudicate, with the reason each likely reads `this-agent`. They cannot pass through unexamined; a wrong verdict must be written down with a rationale, on a visible row, next to the boundary-sample rows it contradicts.
  - Pre-adjudicating them here would be hand-enumeration a second time — the exact defect class of iteration-1 D3 that the SPEC's own §G AP-6 names. The repair declining to do it is the correct call, and it says so.

  N2's blocking property was "a coordinate set asserted as complete without stating the discriminant that closes it." That property no longer holds.
- **Baseline-attribution.** All coordinates and counts measured at HEAD `b6442e848`, this tree, this run. Synthetic M2 fixture written to the session scratchpad, outside the audited tree.
- **Gaps.** I did not adjudicate `:7`, `:23`, `:29`, `:130` — no claim about them, same as iteration 2. I did not verify that M2's future output will in fact carry the pinned token notation (it does not exist yet); I verified only that a mismatch fails safe.
- **Residual-risk.** If M2 classifies `:172` as prose **and** the rationale is plausible on its face, no mechanical check catches it — the guard is the reviewer reading the rationale row. This is a run-phase execution risk with a named review surface, not a plan-phase specification defect, and it is the irreducible remainder of any property-based binding.

### Advisories found this round (recorded, NOT carrying the verdict)

Both lie outside the N1/N2 blocking property. Per the dispatch's scope clause they are stated as advisory and do not carry a FAIL.

**A9. `spec.md:247`'s `ls -d` self-check is stale by construction · `spec.md:247` · Severity: minor · Class: optional.**
The line records an `ls -d` on this SPEC's own directory returning `No such file or directory` (작성 전 실측). Re-run today the command returns rc 0 and lists the directory. It is the same *class* as N1 — a self-targeting command in a self-verification block whose stated output the SPEC's own existence falsifies — but it differs on both counts that made N1 blocking: it carries an explicit temporal qualifier (작성 전 실측, "measured before authoring"), so it is an honest historical baseline rather than a present-tense claim; and it is **not** a Definition-of-Done item (`acceptance.md` §F carries no `ls -d` line — I read the whole block). It is also the one site the repair's AP-6 enumeration omitted. Optional fix: re-word to name the measurement instant, or drop the line.

**A10. Two residual "4줄 / 네 줄" phrasings survive the coordinate→property move · `acceptance.md:150`, `acceptance.md:167` · Severity: minor · Class: optional.**
`acceptance.md:150` ("**네 줄은** 능력 이름을 부르고…") and the §D label at `acceptance.md:167` ("**`manager-lead` 자기 스폰 4줄**") still describe the class by the old constant. Neither is a **Then** condition — `AC-CBN-013`'s binding text is property-bound to `N` — so neither can produce a false green: if M2 finds N=5 and only four lines are revised, condition (b) `M ≥ N` breaks. The defect is readability drift, not a weakened gate. Optional fix: two phrase edits ("네 줄" → "그 줄들", "4줄" → "자기 스폰 지시 줄").

### Confirmed in passing (A1, A2 — not blocking, dispatch permitted)

- **A1** — repaired. `plan.md:73` now carries a `[HARD]` trap notice stating `moai-mcp` classifies as `present`, and `spec.md:60` states the discriminant is the capability, not the word `unavailable`.
- **A2** — repaired. Both sites now use the leading-alternation form `(^|[^/])design-sync` (`plan.md:128`, `acceptance.md:155`), and `acceptance.md:157` states why that branch is required.

### Regression check (iteration 2 → 3)

| Iteration-2 finding | Status | Evidence |
|---|---|---|
| N1 marker selector self-falsifying | **RESOLVED** | canonical + ERE both rc 1; broad form 4 lines, none asserting 무출력; independent positive control matches a real marker |
| N2 hand-enumerated 4-line class | **RESOLVED** | `REQ-CBN-009` property-bound (`spec.md:164`); `AC-CBN-013` bound to N (`acceptance.md:143`); population 10 cited (`spec.md:121`); §D permission removed (`acceptance.md:167`) |
| A1 `moai-mcp` word trap | RESOLVED (advisory) | `plan.md:73`, `spec.md:60` |
| A2 leading-alternation `design-sync` | RESOLVED (advisory) | `plan.md:128`, `acceptance.md:155` |
| D1-D10 | not re-opened (scope) | iteration-2 verdict stands |

No regression: nothing that passed in iteration 2 was weakened to make N1 or N2 pass. The v0.2.1 HISTORY entry (`spec.md:22`) states the repair touched nothing else, and the two invariants most at risk from a scope-limited repair — the 84/81 unit split and the `Task*→0` / `task-list ≥3` pairing — were spot-checked and are intact (`plan.md:100-102`, `plan.md:127`).

### Recommendation

**N1: closed. N2: closed.** Nothing blocking remains within this round's scope. A9 and A10 are a one-line and a two-phrase edit respectively, both repairable in place, and neither is a precondition for run-phase entry.

The iteration-2 aggregate FAIL was carried by N1 and N2 alone; with both closed and no regression, the plan-phase blocking set for this SPEC is empty. The remaining recorded debt is A3-A10, all optional, all surfaced for the orchestrator's discretion per M6 — routing them into a further revision round would be the over-engineering the Enforce Simplicity core behavior forbids, and iteration 3 is the hard cap regardless.

Process note: HEAD read `b6442e848` at the opening and closing of this audit, with a clean tree at both reads. No recurrence of the defect recorded at `.moai/reports/t497/process-defects.md`.
