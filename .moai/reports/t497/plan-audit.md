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
