# SPEC Review Report: SPEC-CODEX-SKILL-NEUTRAL-001

Iteration: 1/2 (Tier M ceiling per `harness.plan_audit_tier_ceilings`)
Verdict: **FAIL**
Overall Score: **0.775** (Tier M PASS threshold 0.80)

Measurement tree: `.claude/worktrees/t196`, HEAD `297a21ea7`, branch `WT-codex-skill-neutral`.
Reasoning context ignored per M1 Context Isolation — the SPEC author's rationale was not read as
authority; only the artifact files and the tree were.

Input contract: Tier M → `spec.md` + `plan.md` + `acceptance.md` read (all three), plus
`progress.md` and `.moai/reports/t196/premise-remeasure.md` as cited evidence.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -o 'REQ-CSN-[0-9]*' spec.md | sort -u` → 001..014
  contiguous, no gaps, no duplicates, uniform 3-digit padding; `grep -c '^- \*\*REQ-CSN-'` → 14
  (definition count equals distinct-id count, so no id is defined twice).
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-CSN-*` in
  `spec.md`), not the verification layer. All 14 match a GEARS shape: ubiquitous shall/shall-not
  (002, 003, 005, 006, 007, 008, 010, 012, 013, 014), When-trigger (004, 011), Where-gate (001, 009).
  Two soft notes below (D8) do not reach FAIL.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`id`/`title`/`version`/`status`/`created`/`updated`/`author`/`priority`/`phase`/`module`/
  `lifecycle`/`tags`, spec.md:2-14). No rejected snake_case alias. `phase: "v3.2.0 target"` is a
  release label, not a prohibited lifecycle token. Optional `tier: M` present. `plan.md` and
  `acceptance.md` carry no frontmatter at all, satisfying artifact statelessness.
- **[N/A] MP-4 language neutrality** — the SPEC names no language-specific toolchain and does not
  cover multi-language tooling. Its template-neutrality concern (REQ-CSN-013) is a different axis
  (internal-content isolation), audited under D1 below.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — two referenced SPECs, both resolve and both are
  `status: completed`: `.moai/specs/SPEC-CODEX-SKILLS-CANONICAL-001/spec.md:5` → `completed`,
  `.moai/specs/SPEC-CODEX-DUAL-AGENTS-001/spec.md` → `completed`. Neither is retired/superseded/
  archived, so no reconciliation clause is owed. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → `0`. Auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-SKILL-NEUTRAL-001/`
  → no matches.

All seven must-pass criteria hold. **The FAIL is a score verdict, not a firewall trip** — the
defects below are localized to the acceptance layer and two plan premises, so iter-2 should be a
bounded delta fix rather than a rewrite.

---

## Category Scores

| Dimension | Score | Rubric band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75–1.0 | Requirements single-interpretation throughout; §B design record states each rejection's ground. Deduction: §A.7 / spec.md:93-99 attributes the relative-path property to the wrong cause (D2). |
| Completeness | 0.85 | 0.75–1.0 | All sections present; four `### Out of Scope — <topic>` H3 sub-headings each with specific `-` bullets (spec.md:178-195). Deduction: the load-bearing cwd precondition is nowhere recorded (D2). |
| Testability | 0.70 | 0.50–0.75 | Most ACs carry an executable command with an explicit pre/post value pair — genuinely strong, and AC-CSN-009 is a correctly-specified RED observation. Deduction: AC-CSN-010 is not executable as written (D4); AC-CSN-005 is mutant-passable against the REQ it claims to judge (D1). |
| Traceability | 0.70 | 0.50–0.75 | §D.2 maps every AC to a REQ and §D.3 declares two uncovered REQs as debt rather than faking coverage — the honest shape. Deduction: one mapping is false (D1), and plan.md milestone closing conditions contradict §D.2 (D3). |

Aggregate = mean(0.85, 0.85, 0.70, 0.70) = **0.775** < 0.80.

---

## Defects Found

### D1 — `REQ-CSN-013` is unguarded in the exact file M2 edits, and the AC mapped to it tests nothing about it — `acceptance.md`:L124, `plan.md`:L21, L113 — Severity: **critical** — Class: **blocking**

**This defect was verified by running, not reasoned from reading.**

`plan.md` §B.3 states the template-neutrality CI guard watches `internal/template/templates/**` and
therefore the binding table "cannot" carry a SPEC ID or REQ token; `plan.md` AP-7 repeats it as
"the neutrality CI guard catches it". Both are false for `internal/template/templates/AGENTS.md`,
which is precisely the file M2 writes.

Measured (mutant probe, both tiers, this tree, HEAD `297a21ea7`):

```
baseline, no probe:
  go test ./internal/template/ -run TestTemplateNoInternalContentLeak            → ok (1.055s)
  MOAI_TEMPLATE_LEAK_STRICT=1 go test ... -run TestTemplateNoInternalContentLeak → ok (0.957s)

probe planted at templates ROOT (internal/template/templates/ZZ_audit_probe.md),
body containing the literals SPEC-CODEX-SKILL-NEUTRAL-001, REQ-CSN-003, AC-CSN-004:
  narrow tier → ok (0.919s)          <- NOT CAUGHT
  strict tier → ok (0.867s)          <- NOT CAUGHT

control, identical literals inside a skill body
(internal/template/templates/.claude/skills/zz-audit-probe/SKILL.md):
  narrow tier → FAIL, 2 occurrences, class=S3-req-ac-token-any-prefix
                 match=REQ-CSN-003, match=AC-CSN-004
```

Both probes removed; `git status --short internal/template/templates/` empty before and after.

Cause, from the guard source: the classes that would match this SPEC's own tokens are
`skillBodyScoped: true` — `C1b-spec-id-skill-v3r` and `S3-req-ac-token-any-prefix`
(`internal/template/internal_content_leak_test.go`, leakClasses/strictLeakClasses scope flags), so
they fire only under `.claude/skills/`. The whole-tree classes are prefix-limited:
`C1-spec-id-prefix` matches `SPEC-(V3R[2-6]|AGENCY|WORKTREE)-` and `C2-req-ac-internal-prefix`
matches `(REQ|AC)-(ATR|WO|COORD|UNP|LNC|TII|HRN|ORC)-[0-9]{3}`. `SPEC-CODEX-…` and `REQ-CSN-…`
match neither.

Compounding it, `acceptance.md` §D.2:L124 maps `AC-CSN-005 → REQ-CSN-005, REQ-CSN-013`, but
AC-CSN-005's Then clause (acceptance.md:L47) tests only `cmp` equality of the two copies and the
byte-ceiling test. A binding table citing `SPEC-CODEX-SKILL-NEUTRAL-001` as its ground would pass
AC-CSN-005 in full while violating REQ-CSN-013 — a writable mutant, which is the shallow-criterion
signature. So REQ-CSN-013 has **no mechanical guard and no real AC**, while two documents assert it
is covered.

Required fix: (a) correct `plan.md` §B.3 and AP-7 to state that the guard does **not** cover
`templates/AGENTS.md` for the SPEC-ID and REQ-token classes (dates/SHAs were not probed — do not
claim about them without measuring); (b) either give REQ-CSN-013 its own AC with an executable
command scoped to the two AGENTS.md copies, or move it to §D.3 as declared debt. Do not leave it
mapped to AC-CSN-005.

### D2 — `§A.7` attributes the relative-path property to the wrong cause; the real precondition (cwd = project root) is unrecorded, and `REQ-CSN-009` guards the arm that cannot fail — `spec.md`:L93-99, L164 — Severity: **major** — Class: **blocking**

Reasoned from reading plus one code read; not runtime-observed.

§A.7 argues: *because* the mirror is a relative symlink, a project-root-relative path
`.claude/skills/<name>/…` resolves identically under both harnesses. The premise does not carry the
conclusion. A project-root-relative path names the canonical file directly; whether
`.agents/skills/<name>` is a symlink, a real copy, or absent is irrelevant to its resolution. The
property the claim actually depends on is that the reading process's working directory is the
project root.

That premise is supported on one path and undefended on others. `internal/cli/codex_launcher.go:243-250`
sets the launch cwd to the project root — and explicitly degrades to the process cwd when the root
is unresolvable. A codex session started directly (not through `moai codex`) inherits whatever cwd
the user had. None of this is recorded in the SPEC, and no requirement binds it.

The consequence is that REQ-CSN-009 (spec.md:L164) obliges run-phase to observe the copy-fallback
arm — an arm no mutation of the mirror mode can break, since the path never reaches the mirror —
while the arm that can break is unbound. This inverts the SPEC's own stated discipline of
observing rather than arguing.

Worth recording as an observation, not a defect: under copy-fallback the mirror body is frozen at
first-run content while the paths it names resolve to the fresh canonical. That mixed freshness is
strictly better than the status quo, but it is a consequence of B.D5 that the SPEC does not state.

Required fix: restate §A.7's ground as the cwd property, record the launcher's project-root cwd
(`codex_launcher.go:243-250`) with its degradation branch as the supporting measurement, and
re-point REQ-CSN-009 at the cwd arm — or retire REQ-CSN-009 and state plainly that the mirror mode
is irrelevant to a root-relative path.

### D3 — `plan.md` milestone closing conditions claim coverage that `acceptance.md` §D.2 denies — `plan.md`:L90, L101 vs `acceptance.md`:L131 — Severity: **major** — Class: **blocking**

`plan.md` §E declares each milestone closes its own conditions within itself. But:

- M3 (plan.md:L90) lists closing conditions REQ-CSN-006/007/008/**009**, judged by AC-CSN-006/007/008
  — three ACs for four REQs. `acceptance.md`:L131 states REQ-CSN-009 has **no** dedicated AC.
- M4 (plan.md:L101) lists closing conditions REQ-CSN-010/011/**012**/014, judged by AC-CSN-009/010.
  AC-CSN-009 judges 010+011, AC-CSN-010 judges 014. `acceptance.md`:L131 states REQ-CSN-012 has
  **no** dedicated AC.

So two milestones name a closing condition that nothing in the judgment layer can close. Under the
plan's own §E rule, neither M3 nor M4 can close as written. This is the cross-layer shape: the
debts were declared honestly in `acceptance.md` but the plan was not swept to match.

Required fix: in plan.md M3 and M4, move REQ-CSN-009 and REQ-CSN-012 out of the closing-condition
lists into an explicit "carried as declared debt (acceptance.md §D.3)" line, so a reader of plan.md
alone cannot conclude they are judged.

### D4 — `AC-CSN-010` is not executable as written — `acceptance.md`:L105 — Severity: **major** — Class: **blocking**

The command is `git diff --stat <base>..HEAD -- …`. `<base>` is an unresolved placeholder; the AC
names no commit to diff against. An out-of-scope-surface assertion is an invariant claim, so it
must pin a tree SHA rather than a moving ref — a base read as a branch name advances under the
assertion and can report changes that were never made, or hide changes that were.

Required fix: pin the base to the plan-phase commit SHA (the SPEC's own baseline `297a21ea7`, or
whatever HEAD the run-phase starts from, recorded at start), and state the SHA in the AC.

### D5 — `§A.4`'s failure-mode column is inference presented under a heading that declares the section measured — `spec.md`:L32, L70-79 — Severity: **major** — Class: **blocking**

§A is titled "검증된 기준선 (실측)" — verified baseline, measured. The §A.4 table's line counts (6
HARD / 40 SOFT) are measured. Its "실패 방식" column is not: "실행이 실패하지 않는다. 모델이
추론으로 메운다" for the SOFT grade is an unobserved behavioral claim, and the premise-remeasure
report's own Gaps section says so. The SPEC flags exactly one such inference (the axis-A tool-name
substitution, bound by REQ-CSN-001) and leaves the structurally identical axis-B SOFT claim
unflagged and unbound.

This matters because §B.D4 — the rejection of the HARD-only scope — rests on it: "SOFT/조용한
계열이 카드가 말하는 신뢰 격차를 실제로 만든다 … 남겨진 40줄 + 축 A 전체는 아무도 배우지 못하는
자리다." A rejection resting on an unverified premise is the failure mode this audit was asked to
hunt.

Mitigating, and the reason this is major rather than critical: the **prescription** does not depend
on it. REQ-CSN-006/007 remove the token, and an empty expansion is wrong whether it fails loudly or
quietly. Only the priority ordering and the B.D4 rationale lean on the inference.

Required fix: mark the §A.4 "실패 방식" column as inference in the table itself (the line counts
stay measured), and either extend REQ-CSN-001's observation obligation to cover one SOFT-grade
site or state in §B.D4 that the rejection stands on the token being wrong under any failure mode.

### D6 — `REQ-CSN-001`'s placement protects axis A only, and does not protect the design decisions that actually consume the inference — `spec.md`:L153, `plan.md`:L53-64 — Severity: **minor** — Class: **blocking**

Answering the lead's probe directly: the M1 placement **does** genuinely protect the axis-A
requirements from proceeding on an unconfirmed inference — REQ-CSN-002..005 are downstream of it,
M1's blocker instruction stops M2 before any body edit, and AC-CSN-001 correctly passes on a
falsifying observation (acceptance.md:L13), which is the right shape for an observation-not-outcome
criterion.

What it does not protect: §C's priority ordering and §B.D4's rejection were written and fixed
**before** the observation, and nothing in M1 requires them to be revisited if the observation
falsifies the inference. plan.md:L62 says only "본문을 편집하기 전에 blocker 로 올린다" — it does
not say the design record is reopened. And REQ-CSN-003/004 (the binding table and the blocker-return
rule) are inference-independent: they are the right prescription whether codex substitutes silently
or refuses loudly. So the observation gates the edits but not the reasoning that ordered them.

Required fix: add to M1's blocker clause that a falsifying observation reopens §B.D4 and the §C
ordering, not only the M2 edits.

### D7 — the stated upper bound of 14 is used as an exact addend in the Tier rationale — `spec.md`:L123, L212; `plan.md`:L22 — Severity: **minor** — Class: **optional**

Answering the lead's probe: no *requirement* and no *AC* treats 14 as exact — §D:L181 states the
upper-bound caveat plainly and the out-of-scope entry is correctly scoped. But §B.D2:L123 writes
"조건절이 25개 파일(스킬 14 + 에이전트 11)에 복제된다" and §E.1:L212 writes "Tier L 로 넘겼을
항목(25파일 산문 재작성)" — both consume 14 as a count, without the caveat, inside the argument that
fixes the Tier at M. If the true figure is lower, B.D2's duplication-cost argument weakens
proportionally. It does not collapse: B.D2 explicitly disclaims budget as its ground and rests on
(a) no single audit point, (b) author-memory burden, (c) always-loaded body inflation — all
count-independent. Hence optional.

Secondary, same family: `plan.md` §B.4:L22 instructs "M2 는 파일 수를 근거로 판정하지 말고 자리별로
읽어야 한다" — but M2 does not touch those 14 files at all; they are out of scope. The warning is
aimed at a judgment M2 never makes.

Required fix (if taken): carry the "≤" qualifier into §B.D2 and §E.1, and re-point or drop plan.md
§B.4.

### D8 — two `Where`-clause misuses and one compound requirement — `spec.md`:L153, L164 — Severity: **minor** — Class: **optional**

REQ-CSN-001 opens `Where 코덱스가 … 거동이 아직 관측되지 않았다` — GEARS reframes `Where` as a
capability gate / feature flag / static config; "a fact has not yet been observed" is a transient
state, closer to `While`. REQ-CSN-009's `Where 미러가 … 복사로 만들어졌다` is a genuine static-config
gate and is fine, but it then chains two obligations in one requirement ("확인해야 하며, 논증으로
갈음해서는 안 된다"). Neither reaches MP-2 FAIL.

### D9 — `AC-CSN-006` is strictly subsumed by `AC-CSN-008` — `acceptance.md`:L63, L85 — Severity: **minor** — Class: **optional**

AC-CSN-006 pipes `grep -rn 'CLAUDE_SKILL_DIR' … | grep -E ':(bash|node) ' | wc -l`. Once AC-CSN-008
holds (zero total occurrences), the first stage of AC-CSN-006's pipeline is empty and its result is
0 unconditionally — it cannot independently fail. It is not vacuous in the harmful sense, because
its real content is the **pre-value pairing** the AC mandates (acceptance.md:L66), and that pairing
is exactly right. Worth stating that the post-state assertion is redundant so a later reader does
not mistake its green for independent evidence.

---

## Answers to the six directed probes

1. **REQ-CSN-001 placement** — genuinely protective for the axis-A edits; not protective for the
   §C ordering and the §B.D4 rejection that consume the same inference. See **D6**, and **D5** for
   the unflagged sibling inference on axis B.
2. **The (c) rejection (B.D3)** — **the cited defect says what the rejection claims.** Read directly:
   `SPEC-CODEX-SKILLS-CANONICAL-001` §D, "폴백 플랫폼 미러 고착" — on symlink-incapable platforms the
   mirror becomes a real directory copy (REQ-CSC-004) and from the second deploy onward is caught by
   REQ-CSC-014's real-entry branch and skipped, freezing at first-run content. REQ-CSC-014 verified
   verbatim at that SPEC's spec.md:208, and the branch verified in code —
   `internal/template/skill_mirror.go:198-205` returns `MirrorModeSkipped` on a non-symlink entry
   and leaves it untouched. B.D3's inference that forcing the copy path on all platforms promotes
   this from exception to default follows correctly. **The rejection stands.**
3. **The relative-path design finding** — the conclusion holds but for a different reason than the
   one given, and the copy-fallback arm REQ-CSN-009 guards is not where the risk is. See **D2**.
4. **The §D.3 undecidable debt** — this is **honest scoping, not evasion**, and it is the strongest
   part of the artifact set. §D.3(1) states outright that AC-CSN-002~004 judge only that the table
   exists and has the right shape, and that the effectiveness claim is not made by this SPEC. Naming
   the central promise as unjudged, rather than dressing a grep as proof of it, is the correct
   disposition — an effectiveness verdict needs a two-harness differential run whose baseline this
   SPEC does not have. The debt in §D.3(2) and (3) is likewise real and correctly typed. The defect
   is not in §D.3; it is that §D.2 *contradicts* §D.3 by mapping AC-CSN-005 to REQ-CSN-013 (**D1**),
   and that plan.md contradicts it again at the milestone layer (**D3**).
5. **AC-CSN-009's RED** — **the AC does specify how the RED is produced and observed**, not merely
   assert it: plant one `CLAUDE_SKILL_DIR` token in the watched tree, run the guard, observe FAIL,
   revert, re-run, observe PASS, and record the census on both sides (acceptance.md:L93-96). The
   pre-plant census requirement is what makes the failure attributable to the plant. This criterion
   is correctly constructed and needs no change.
6. **The 14 upper bound** — no requirement or AC treats it as exact; two argument passages do. See
   **D7**.

---

## Verified-by-running vs reasoned-from-reading

**Verified by running** (this tree, HEAD `297a21ea7`):

- REQ id set and definition count (`grep -o` + `grep -c`) — MP-1.
- AC id set and heading count — 10/10.
- `[NEEDS CLARIFICATION]` sweep — 0 matches (MP-7).
- `syscall` count in spec.md — 0 (MP-6).
- Referenced-SPEC existence and status — both `completed` (MP-5).
- Time-estimate sweep across all three artifacts — no matches.
- **D1's mutant probe** — baseline green both tiers; probe at templates root green both tiers;
  identical tokens in a skill body FAIL with `class=S3-req-ac-token-any-prefix`. Probes removed,
  tree clean before and after.
- `internal/template/templates/AGENTS.md` is in the byte-ceiling guard's subject set
  (`internal/config/token_budget_guard.go` `contractDocuments`) — the SPEC's §A.6 claim that the
  guard watches both copies is correct.

**Reasoned from reading** (code and artifacts, not executed):

- MP-2 GEARS shape judgment on all 14 requirements, and D8.
- MP-3 frontmatter field-by-field check.
- D2's cwd analysis — `codex_launcher.go:243-250` was read, not run; no codex session was launched.
- D3, D4, D6, D7, D9 — internal-consistency reading across the three artifacts.
- D5 — classification of the §A.4 column as inference, corroborated by the premise-remeasure
  report's own Gaps section.
- Probe 2's chain — `SPEC-CODEX-SKILLS-CANONICAL-001` §D and REQ-CSC-014 were read; the skip branch
  in `skill_mirror.go` was read, not exercised.

**Not observed at all** (carried forward as the SPEC itself states): codex runtime behavior on an
unknown tool name; copy-fallback mirror materialization; the `${CLAUDE_SKILL_DIR}` line counts
(6/40/46), which were taken as given from the orchestrator's own re-measurement rather than
re-derived here.

---

## Recommendation

FAIL at 0.775 against the Tier M threshold of 0.80. The must-pass firewall is clean and the design
record is unusually strong — the deficit is concentrated in the judgment layer. iter-2 should be a
bounded delta over the enumerated defects, not a re-draft:

1. **D1** — correct `plan.md` §B.3 and AP-7 to state the guard's real scope; remove REQ-CSN-013
   from AC-CSN-005's mapping and either give it an executable AC over the two AGENTS.md copies or
   move it to §D.3 as debt. (blocking)
2. **D3** — remove REQ-CSN-009 from M3's closing conditions and REQ-CSN-012 from M4's, marking both
   as carried debt. (blocking)
3. **D4** — pin AC-CSN-010's `<base>` to a stated commit SHA. (blocking)
4. **D2** — restate §A.7's ground as the cwd property, cite `codex_launcher.go:243-250` including
   its degradation branch, and re-point or retire REQ-CSN-009. (blocking)
5. **D5** — mark §A.4's failure-mode column as inference; either extend REQ-CSN-001 to one
   SOFT-grade observation or state that B.D4 holds under any failure mode. (blocking)
6. **D6** — add to M1 that a falsifying observation reopens §B.D4 and §C ordering. (blocking, small)
7. **D7 / D8 / D9** — optional; orchestrator's discretion.

No stagnation applies (first iteration). Tier M permits one further iteration.
