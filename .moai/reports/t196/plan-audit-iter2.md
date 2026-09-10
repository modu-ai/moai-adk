# SPEC Review Report: SPEC-CODEX-SKILL-NEUTRAL-001

Iteration: 2/2 (Tier M ceiling per `harness.plan_audit_tier_ceilings`) — **final iteration**
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.825** (Tier M PASS threshold 0.80)

Measurement tree: `.claude/worktrees/t196`, HEAD `297a21ea7`, branch `WT-codex-skill-neutral`.
Reasoning context ignored per M1 Context Isolation — the author's iter-2 rationale, as relayed in
the delegation, was treated as a set of claims to test, not as authority. The four dimensions were
scored afresh against the current artifacts; the iter-1 numbers were not carried forward.

Scope: delta audit over the enumerated iter-1 defect list (D1-D10) plus a new-defect sweep, per the
iteration-2 contract. Parts iter-1 passed were re-checked only where an iter-2 edit touched them —
except the must-pass sweeps, which were re-run in full because the artifacts changed substantially
(REQ 14→15, AC 10→13).

No score regression (0.75 → 0.825), so the STOP escalation clause does not fire.

---

## Must-Pass Results (all re-run at this HEAD)

- **[PASS] MP-1 REQ number consistency** — `grep -o 'REQ-CSN-[0-9]*' spec.md | sort -u` → 001..015
  contiguous, no gaps, uniform padding. `grep -o '^- \*\*REQ-CSN-[0-9]*' | sort | uniq -c` → every
  id exactly `1`, and `grep -c '\*\*REQ-CSN-'` → 15 equals the definition count, so **no residue of
  the near-duplicate definition the author reports reverting** survives anywhere in the file.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-CSN-*` in
  `spec.md`). All 15 match a GEARS shape. The two new/changed ones: REQ-CSN-001 now opens `While`
  (iter-1 D8's note adopted — a transient unobserved state is `While`, not `Where`); REQ-CSN-015 is
  ubiquitous shall/shall-not. REQ-CSN-015 chains two obligations in one requirement (remove-or-change
  the norm **and** retain the fact) — the same soft note as iter-1 D8, still not reaching FAIL.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (spec.md:2-15); `version` bumped to `"0.2.0"` quoted, `updated: 2026-09-01` ISO. No rejected
  snake_case alias. Optional `tier: M` present. `plan.md` / `acceptance.md` carry no frontmatter.
- **[N/A] MP-4 language neutrality** — unchanged from iter-1; the SPEC names no language-specific
  toolchain. Its template-neutrality axis is REQ-CSN-013, audited below as D11.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — two referenced SPECs, both resolve, both
  `status: completed` (re-read at this HEAD). No retired/superseded/archived reference, so no
  reconciliation clause is owed. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → `0`. Auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .` over the SPEC directory
  → `0`.

Additional cheap sweep: time-estimate scan (`[0-9]+ *(일|주|시간|days?|weeks?|hours?)`) across all
three artifacts → no matches. Five `### Out of Scope — <topic>` H3 sub-headings, each with specific
`-` bullets (spec.md:264, 269, 274, 279, 284) — one more than iter-1, the new one being the
fact-description carve-out.

---

## Category Scores

| Dimension | iter-1 | iter-2 | Rubric band | Evidence |
|---|---|---|---|---|
| Clarity | 0.85 | **0.85** | 0.75–1.0 | D2's re-grounding is correct and its line pin verified accurate; §A.4's inference column now labelled; §B.D4 split into an inference-independent primary ground and an inference-dependent secondary one. Deduction: `plan.md`:L13 still frames the silent-failure claim as `측정 결과`, contradicting spec.md §A.4 and plan.md's own AP-1 (**D12**). |
| Completeness | 0.75 | **0.90** | 0.75–1.0 | D10 fully absorbed: REQ-CSN-015, §A.8, §B.D6, a dedicated Out-of-Scope entry, plan M3 step, AP-12/13, §E.1 file count 12→14, §E.2 completion state rewritten. Token footprint now closed at 50 = 46 in-scope-to-zero + 4 in-scope-by-content. My own sweeps found no further unseen surface. |
| Testability | 0.70 | **0.75** | 0.50–0.75 | AC-010's base pinned to a tree SHA with an explicit re-measure-and-record instruction; AC-011 written as a genuine set assertion; AC-012 added with a positive control that I verified fires on all four arms; AC-006's subsumption disclosed. Deduction: **D11** (AC-012 under-scopes the REQ it is mapped to — verified by probe), plus D13/D14/D15. |
| Traceability | 0.70 | **0.80** | 0.75–1.0 | D1's false mapping corrected; D3's milestone/debt contradiction swept across all four milestones (verified exhaustively below); §D.2 covers 14/15 REQs with REQ-CSN-012 declared debt in **both** acceptance.md §D.3(3) and plan.md §F; plan §E now carries the rule as a general invariant. Deduction: the AC-012 → REQ-CSN-013 mapping asserts more coverage than the AC delivers (**D11**) — the iter-1 D1 shape at reduced scope. |

Aggregate = mean(0.85, 0.90, 0.75, 0.80) = **0.825** ≥ 0.80.

---

## Per-defect closure judgment

### Blocking defects from iter-1

| iter-1 defect | Judgment | Evidence |
|---|---|---|
| **D1** — REQ-CSN-013 unguarded + false plan.md claim + bogus AC mapping | **Partially closed** | Two of three parts closed and verified; the third re-opens narrower as **D11**. |
| **D10** — rules-tree remainder outside all scope | **Closed** | Verified. |
| **D3** — milestone closing conditions contradicting declared debt | **Closed** | Verified exhaustively. |
| **D4** — unresolved `<base>` placeholder | **Closed** | Verified. |
| **D2** — §A.7 wrong ground; cwd precondition unrecorded; REQ-CSN-009 mis-aimed | **Closed on ground and citation; the re-aim is disclosure, not binding** | See below. |
| **D5** — §A.4 inference under a "measured" heading | **Closed in spec.md; leaked in plan.md** | New minor **D12**. |
| **D6** — M1 gates edits but not the reasoning | **Closed** | Verified. |

**D1 — partially closed.** Two parts hold. (a) The false claim is corrected and correctly bounded:
`plan.md`:L21-23 and AP-7:L134 now state that the neutrality guard does not watch
`templates/AGENTS.md` for this SPEC's token forms, cite the class table as the ground, and
**explicitly decline to claim anything about the date/SHA classes, which were never probed**. That
restraint held — I checked for an overclaim in either direction and found none. (b) The bogus
mapping is gone: `acceptance.md` §D.2 now reads `AC-CSN-005 → REQ-CSN-005` and
`AC-CSN-012 → REQ-CSN-013`, with the correction narrated in-place. What does not hold is (c): the
replacement AC does not cover the whole of the REQ it is mapped to — see **D11**.

Side note on the restraint: reading `internal_content_leak_test.go` shows `S1-internal-date` and
`S2-short-sha-sentence-final` are **not** `skillBodyScoped`, and
`.github/workflows/template-neutrality-check.yaml`:85 sets `MOAI_TEMPLATE_LEAK_STRICT: '1'`. So
dates and short SHAs **are** guarded tree-wide in CI. The plan's "no claim" stance is therefore
conservative rather than wrong — it under-claims a guard that exists. Not a defect.

**D2 — closed on the axis it was raised on; one residue named rather than scored as a fix failure.**
The line pin is accurate, which mattered: an inaccurate pin inside the fix for a wrong-ground defect
would have been the same failure twice. Read directly, `internal/cli/codex_launcher.go`:242-244 is
the comment (`The launch cwd is the PROJECT ROOT, not the process cwd … An unresolvable root
degrades to the process cwd rather than refusing to launch`), :245-250 is the resolution block
verbatim as §A.7 quotes it, and :248-250 is the degradation branch. §A.7 cites all three correctly
and §F's cross-reference `codex_launcher.go:242-250` matches.

On whether REQ-CSN-009 now binds anything: it binds a **documentation state**, and AC-CSN-013 judges
that state by reading prose. The honest reading is that both were **already satisfied by spec.md as
written at plan time** — §A.7 carries (a) and (b), and §D.3(2) carries (c). So M1's REQ-CSN-009
closing condition is met before run-phase begins, and the criterion can only fail by someone
deleting the paragraph. I judge this **honest scoping rather than a criterion that cannot fail**,
because AC-CSN-013 says outright what it is for — *"이 SPEC 은 cwd 팔을 닫지 않는다. 닫히지 않은 것을
닫혔다고 읽히게 두지 않는 것이 이 AC 의 전부다"* — and the alternative (a requirement binding the cwd
arm) would require launcher work that REQ-CSN-014 forbids and card t391 owns. It is disclosure, and
it is labelled as disclosure. The residue worth naming: a reader of `plan.md` alone sees M1 close
REQ-CSN-009 and may take that as evidence *about* cwd rather than evidence that cwd was *written
down*. Minor, recorded as **D15**.

**D3 — closed.** I re-derived the full milestone/AC/REQ closure rather than spot-checking:

| Milestone | Closing REQs | Judging ACs | §D.2 mapping agrees? |
|---|---|---|---|
| M1 | 001, 009 | AC-001, AC-013 | yes |
| M2 | 002, 003, 004, 005, 013 | AC-002/003/004/005/012 | yes |
| M3 | 006, 007, 008, 015 | AC-006/007/008/011 | yes |
| M4 | 010, 011, 014 | AC-009, AC-010 | yes (AC-009 judges 010+011) |
| carried debt | 012 | — | declared in acceptance §D.3(3) **and** plan.md:L121 |

14 REQs in closing conditions + REQ-CSN-012 as debt = 15. No milestone now names a condition the
judgment layer cannot close, and no REQ is orphaned. The §E rule is stated as a general invariant —
*"판정 계층이 닫을 수 없는 요구사항은 닫힘 조건이 아니다"* (plan.md:L53) — with the current carried
item named beneath it, which is the right shape: the invariant survives a future change to which
item is carried.

**D4 — closed.** `acceptance.md`:L105-116. The base is pinned to the full 40-char SHA
`297a21ea73b24e6605280625e576555e4316263e` in the command, restated in the [HARD] clause, with the
moving-ref rationale stated and an explicit instruction to re-measure and record in `progress.md` if
run-phase starts from a different HEAD rather than swapping it silently. The `--stat` choice is
justified against the prefix-filter misread. Two occurrences of the pinned SHA, as expected.

**D5 — closed where it was raised, leaked one document over.** In spec.md the column header now reads
`실패 방식 (**추론 — 미관측**)` and a [HARD] paragraph beneath the table says the last column is not
measured and explains the asymmetry that motivated the flag. §B.D4 is genuinely split, not
cosmetically: the primary ground (*"빈 전개면 그 경로는 틀렸다 … 이 근거는 §A.4 의 미관측 열에 전혀
기대지 않는다"*) is inference-independent, and the secondary (priority ordering) is labelled
inference-dependent and named as reopened by M1.

I checked what still consumes the secondary ground as if it were primary. Downstream of §B.D4, the
consumers are plan.md M3:L101 (*"시끄러운 쪽만 먼저 닫는 분할은 §B.D4 가 기각한 형태"*) — this
consumes the **primary** ground, correctly — and plan.md §G AP-2 (*"배우게 되는 쪽을 고치고 못 배우는
쪽을 남긴다"*), whose gloss leans on the inference while citing §B.D4. That is small. The real leak is
**plan.md §A:L13** — see **D12**.

**D6 — closed.** `plan.md`:L70-75 adds a [HARD] clause naming three reopen targets on a falsifying
observation (§B.D4's secondary ground, §C's ordering, §A.4's failure-mode column), plus the reason
the edits-only stop is insufficient, and mirrors it as AP-1b at L128. This is exactly the fix asked
for.

### Optional defects from iter-1

| iter-1 defect | Judgment |
|---|---|
| **D7** (14 used as an exact addend) | **Adopted.** §B.D2 now writes `최대 25개 파일(스킬 ≤14 + 에이전트 11)` and adds that (a)(b)(c) are count-independent so the rejection survives a lower figure; §E.1 writes `최대 25파일`; plan.md §B.4 is re-pointed to say the 14 is used only in the §B.D2 argument and that M2 does not touch those files. |
| **D8** (`Where` misuse, compound requirement) | **Adopted for REQ-CSN-001** (`Where` → `While`). REQ-CSN-009's compound is resolved by the rewrite. REQ-CSN-015 introduces a new mild compound — noted, not scored. |
| **D9** (AC-006 subsumed by AC-008) | **Adopted.** A [HARD] note at acceptance.md:L68 states the post-state assertion is subsumed, names the pre/post pair as the AC's real content, and says a later reader must not read its green as independent evidence. |
| **(1c)** renderer passthrough | **Adopted.** plan.md §B.8:L28 and AP-11:L135, both citing `renderer_test.go:429-446`. |
| M3 dogfood census | **Adopted.** plan.md M3:L104-105, as a pre/post pair, with the [HARD] rationale for *not* widening the guard to the local tree (binary-lag false red). |

---

## New defects

### D11 — `AC-CSN-012` judges two files, but the REQ it is mapped to is scoped to the whole template tree — and the gap is exactly where iter-2 newly put edits — `acceptance.md`:L127-140, §D.2:L156 — Severity: **major** — Class: **blocking**

**Verified by running.**

REQ-CSN-013 (spec.md:L257) binds *"`internal/template/templates/**` 에 실리는 어떤 내용도"* — the whole
template tree. `acceptance.md` §D.2 maps `AC-CSN-012 → REQ-CSN-013`, but AC-CSN-012's command reads
exactly two paths: `AGENTS.md` and `internal/template/templates/AGENTS.md`.

Before iter-2 that gap was mostly covered by the CI guard, because the SPEC's only other template
edits were under `templates/.claude/skills/**`, where the `skillBodyScoped` classes do fire.
**Iter-2's D10 fix moved two files into scope that are in neither set**:
`templates/.claude/rules/moai/development/skill-authoring.md` and
`.../workflow/worktree-integration.md` (plan.md M3:L102).

Probe, this tree, HEAD `297a21ea7` — a new file under the rules tree carrying this SPEC's own
tokens:

```
probe: templates/.claude/rules/moai/development/zz_audit_probe.md
body:  "Ground: SPEC-CODEX-SKILL-NEUTRAL-001 REQ-CSN-003 AC-CSN-012 requires this."
  narrow tier → ok (0.912s)   <- NOT CAUGHT
  strict tier → ok (0.871s)   <- NOT CAUGHT

positive control, same path, body "Landed 2026-08-31 in commit 297a21ea.":
  narrow tier → ok (0.912s)
  strict tier → FAIL, 2 occurrences
    [1] …/zz_audit_probe.md | class=S1-internal-date          | match=2026-08-31
    [2] …/zz_audit_probe.md | class=S2-short-sha-sentence-final | match=297a21ea.
```

The control is what makes the first result interpretable: the guard **does** reach that path, so the
non-catch is a genuine class-scope gap and not an unreached code path. Probe removed;
`git status --short internal/template/` empty before and after.

The reachable mutant: M3 edits `skill-authoring.md`, and the single most natural sentence to write
when inverting a normative rule is *why* it was inverted — a SPEC-ID citation. That violates
REQ-CSN-013, and it is caught by neither the guard (class scope) nor AC-CSN-012 (file scope). It is
the iter-1 D1 shape recurring at reduced scope, created by the D10 fix.

Two things keep this **major** rather than critical, and keep it run-phase-survivable. First, unlike
iter-1's D1, no document now makes a false coverage claim — plan.md §B.3 and AP-7 are accurate and
appropriately narrow. Second, the fix is one line and needs no new mechanism: extend AC-CSN-012's
file list to the two rules files, recorded with the same pre/post pair the SPEC already mandates
everywhere else. The baseline is already measured and non-blocking:

```
$ grep -rEc '<AC-012 regex>' …/skill-authoring.md …/worktree-integration.md
  skill-authoring.md:      2      (:45 and :89 — pre-existing ISO-date examples in frontmatter samples)
  worktree-integration.md: 0
```

So the pre/post form is required rather than a bare `0` assertion — which is exactly the discipline
AC-CSN-006/008/011 already use.

Required fix: add the two rules files to AC-CSN-012's command with their measured pre-values, or
narrow REQ-CSN-013's mapping and declare the remainder as debt. Do not leave §D.2 asserting that
AC-CSN-012 judges REQ-CSN-013 in full.

### D12 — `plan.md` §A states the silent-failure inference as a measurement, contradicting both spec.md §A.4 and plan.md's own AP-1 — `plan.md`:L13 — Severity: **minor** — Class: **blocking (one line)**

Reasoned from reading.

`plan.md`:L13: *"**측정 결과** 그 신뢰를 깨는 것은 두 가지 결합이며, 둘 다 **조용히** 깨진다(spec.md
§A.4). 조용한 실패가 이 카드의 표적이다."*

"둘 다 조용히 깨진다" is precisely the claim spec.md §A.4 now labels `추론 — 미관측`, and plan.md §G
AP-1 itself calls 추론 (*"'조용한 실패가 더 위험하다'는 추론이 검증 없이 설계를 굳힌다"*). Attributing
it to `측정 결과` in the plan's opening context section is the same wrong-heading shape D5 named,
surviving one document over, and it is the first framing a run-phase reader meets.

Required fix: one clause — `측정 결과` → the measured part (the couplings and their line counts) with
the failure mode marked as the inference M1 will observe.

### D13 — `AC-CSN-012`'s cited positive-control figure does not reproduce — `acceptance.md`:L140 — Severity: **minor** — Class: **optional**

**Verified by running.**

The [HARD] clause cites `계획 단계 측정: 34` for the control. At this HEAD:

```
$ grep -cE '<AC-012 regex>' spec.md                      → 45
```

Per-arm decomposition of that 45: SPEC-ID `7` / REQ-AC `34` / hex `2` / date `4`. The pass condition
is written as **non-zero**, not as the specific value, so the criterion itself is unaffected and the
control is not false — but a figure attributed as a plan-phase **measurement**, inside the document
whose own iter-2 edits moved it, is an unreproducible attribution. Note the coincidence worth
recording rather than resolving: `34` is exactly the REQ/AC arm alone, so the figure may have been
taken with a partial regex rather than gone stale. Either reading makes it an incorrect attribution.

Required fix (if taken): re-measure and restate as `45` with its per-arm decomposition, or drop the
parenthetical and keep the non-zero condition — the condition is what carries the criterion.

### D14 — `AC-CSN-012`'s hex arm fires on ordinary English, and the command's `-c` form discards the match text — `acceptance.md`:L134-138 — Severity: **minor** — Class: **optional**

**Verified by running.** The lead's question was whether `\b[0-9a-f]{7,40}\b` can fire on ordinary
words — it can:

```
$ printf 'the table was defaced by hand\naccessed the deadbeef cafe\n' | grep -nE '\b[0-9a-f]{7,40}\b'
1:the table was defaced by hand
2:accessed the deadbeef cafe
```

(`defaced`, `deadbeef`. The system word list has one such word under a case-sensitive match, so the
population is small but non-empty, and technical prose adds more.)

This is a **false-red** hazard, not a vacuous-green one — the AC's subject files measure 0 today and
the control is live, so the criterion works. But the command uses `grep -cnE`, which yields a count
with no match text, so a run-phase FAIL on the word "defaced" in a rewritten `AGENTS.md` would be
unattributable from the AC's own output. The control in spec.md is unaffected: its two hex matches
are the two real SHAs, so the control passes for the right reason on all four arms (7/34/2/4 — every
arm non-zero, which is stronger than the AC asks for).

Required fix (if taken): on the failure path, re-run without `-c` and record the matched strings, so
a false red is separable from a real one in one step.

### D15 — `AC-CSN-011` permits `worktree-integration.md:386` either way while plan M3 mandates its edit; and M1 closes a requirement already satisfied at authoring time — `acceptance.md`:L124, `plan.md`:L102, L79 — Severity: **minor** — Class: **optional**

Reasoned from reading.

Two small permissive gaps, grouped because both are of the "the plan requires more than the AC
judges" shape:

1. `AC-CSN-011`'s expected-residual note says `worktree-integration.md:386` *"수정 결과에 따라
   잔존하거나 사라질 수 있으며, 어느 쪽이든 분류 결과를 적는다."* But plan M3:L102 instructs the edit
   (*"`worktree-integration.md:386` 의 예시를 채택된 설계에 맞게 고치고"*). A run-phase that leaves
   :386 untouched and classifies it `예시` passes AC-CSN-011 while skipping a plan instruction.
2. REQ-CSN-009 / AC-CSN-013 are satisfied by spec.md §A.7 + §D.3(2) as already written, so M1's
   REQ-CSN-009 closure carries no run-phase work. Correct as disclosure (see D2 above), but a
   plan.md-only reader may read the closure as evidence about cwd.

Required fix (if taken): give :386 a stated expected classification, and annotate plan M1's closing
line to say REQ-CSN-009 closes by record, not by observation.

---

## On the two judgments the lead asked me to make hard

**Is `skill-authoring.md:219` really fact rather than norm, in context?** Yes. Read at the source,
:219 is one row of a six-row table under the heading `## Built-in Variables — Variables available
inside skill SKILL.md content`, sitting alongside `${CLAUDE_SESSION_ID}`, `${CLAUDE_PLUGIN_ROOT}`,
`$ARGUMENTS`, `$ARGUMENTS[N]`, `$N`, each with an "Available Since" version column. It states what
Claude Code provides, in a list of what Claude Code provides. Deleting it would make the table
**wrong** — it would assert by omission that the variable does not exist. The author's split is
correct, and the consequence they draw from it is correct too: widening REQ-CSN-010's guard to the
whole template tree would fire on a true sentence, so keeping the guard on the skills tree and
judging the rules tree by an enumerated line set is the right division of mechanism.

**Is AC-CSN-011 actually a set assertion, or a disguised count?** A set assertion. Its Then clause is
*"남은 줄이 사실 기술 집합과 정확히 일치"*, its judgment instruction is to enumerate every remaining
line and classify each in one word, and it carries a [HARD] clause naming the exact substitution a
count would permit (*"규범 문장 하나를 남기고 사실 기술 둘을 지워도 같은 수가 나온다"*), mirrored as
AP-12. Nothing in the criterion reduces to a number. Its one soft spot is the :386 permissiveness
above, which is a coverage gap in the expected set, not a reversion to counting.

**Does keeping a true "this variable exists" row beside a removed "prefer it" sentence still teach
the wrong thing?** Weakly, and less than the alternative. After the fix the shipped rules teach
nothing about *whether* to use the token — the table says it exists, and the normative preference is
gone. A skill author who reaches for it anyway meets REQ-CSN-010's guard, which is the mechanism the
SPEC put there. That is a materially better state than D10's: the author no longer has a documented
rule to follow into the guard. It is not the best available state — a one-clause note on the table
row ("Claude-harness only; empty under other harnesses — prefer a project-root-relative path") would
close the residue entirely, and neither REQ-CSN-015 nor AC-CSN-011 requires it, since REQ-CSN-015
permits pure deletion. I record this as an **optional** improvement rather than a defect, because
silence is not instruction and the mechanical guard covers the tree where the breakage lives.

---

## Verified by running vs reasoned from reading

**Verified by running** (this tree, HEAD `297a21ea7`):

- MP-1 id sets and per-id definition counts, REQ and AC, including the no-duplicate check.
- MP-3 frontmatter field list; MP-5 both referenced SPEC statuses; MP-6 `syscall` count; MP-7
  clarification sweep; time-estimate sweep; Out-of-Scope H3 enumeration.
- **D11's mutant probe and its positive control** — rules-tree SPEC/REQ tokens not caught in either
  tier; date/SHA in the same path caught in strict tier with named classes. Probe removed, tree
  verified clean before and after.
- CI strict-tier setting (`template-neutrality-check.yaml`:85).
- AC-CSN-012's subject measurement (both `0`) and control measurement (`45`), plus the per-arm
  decomposition 7/34/2/4.
- **D14's hex-arm false-positive probe** on ordinary English, plus the case-sensitive word-list count.
- Token census: template tree `50` / skills tree `46` / files `11`, and the four rules-tree lines
  verbatim — §A.8 reproduces exactly.
- §A.6's byte figures: both `AGENTS.md` copies `14229`, `CodexContractByteCeiling = 24576` →
  free `10347`. §A.4's HARD site count `6`.
- AC-CSN-012 regex baseline on the two rules files (`2` / `0`) and the two matching lines.

**Reasoned from reading**:

- MP-2 GEARS shape on all 15 requirements.
- D2's line-pin verification — `codex_launcher.go`:236-256 was read, not executed; no codex session
  was launched, and the cwd property remains unobserved at runtime exactly as the SPEC says.
- The D3 milestone/AC/REQ closure table — cross-document reading.
- The leak-class scope analysis (`skillBodyScoped` gate at `internal_content_leak_test.go`:1105) —
  read as source, then confirmed by the probe above.
- D12, D13's severity classification, D15, and the fact-vs-norm judgment on `skill-authoring.md`:219
  (read in its section context).

**Not observed at all**, carried forward as the SPEC itself states: codex runtime behavior on an
unknown tool name (REQ-CSN-001 binds it as a run-phase obligation); the cwd of a directly-launched
codex session; the failure mode of an empty `${CLAUDE_SKILL_DIR}` expansion under codex.

---

## Recommendation

**PASS-WITH-DEBT at 0.825** against the Tier M threshold of 0.80. The must-pass firewall is clean on
all seven criteria, six of the seven iter-1 blocking defects are fully closed with the closures
verified rather than accepted on report, and the seventh (D1) is closed in two of three parts. The
delta was genuinely a delta: the parts iter-1 passed were not disturbed, and the two design
judgments the fix rested on — the fact/norm split at `:219` and the set-not-count form of
AC-CSN-011 — hold up under direct examination.

This being the final iteration, the residue and its disposition:

**Debt run-phase inherits — must be closed before the milestone that creates the exposure:**

1. **D11 (blocking, major)** — extend `AC-CSN-012`'s command to
   `templates/.claude/rules/moai/development/skill-authoring.md` and
   `.../workflow/worktree-integration.md`, with their measured pre-values (`2` / `0`), **before M3
   lands**. Alternatively narrow REQ-CSN-013's §D.2 mapping and declare the remainder as debt. This
   is run-phase-survivable: it does not block entry, it blocks M3's close, and it is a one-line edit
   to an existing criterion using discipline the SPEC already applies elsewhere.
2. **D12 (blocking, minor)** — one clause in `plan.md`:L13. Cheapest possible fix, and it is the
   first framing a run-phase reader meets, so it should be taken at entry rather than deferred.

**Optional, orchestrator's discretion:** D13 (restate the control figure as `45` or drop the
parenthetical), D14 (record match text on AC-CSN-012's failure path), D15 (state :386's expected
classification; annotate M1's REQ-CSN-009 closure as closing-by-record), and the
`skill-authoring.md`:219 counter-note.

**Neither residue blocks run-phase entry.** D11 blocks M3's close, D12 is a one-clause correction,
and everything else is discretionary. No stagnation applies — no defect appeared unchanged across
both iterations, and the score moved 0.75 → 0.825, so the STOP-on-regression clause does not fire.

## Probe hygiene

One throwaway artifact was created and removed this round:
`internal/template/templates/.claude/rules/moai/development/zz_audit_probe.md` (D11's mutant probe
and its positive control, written twice, removed once). `git status --short internal/template/` was
empty before and after; the only untracked paths in the tree at any point were the two this card
owns (`.moai/reports/t196/`, `.moai/specs/SPEC-CODEX-SKILL-NEUTRAL-001/`). No tracked file was
modified at any point during this audit.
