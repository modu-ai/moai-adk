# SPEC Review Report: SPEC-FEEDBACK-AUTO-SUBMIT-001

Iteration: 3/3 (Tier L retry ceiling)
Verdict: **FAIL**
Aggregate Score: **0.863** — which *clears* the Tier L threshold of 0.85. **The FAIL is score-independent**: it rides on the Retry Loop Contract clause "unresolved defects from a prior iteration are automatically FAIL regardless of other scores." See §5 and §9 — the escalation view is where this report's actual value is.

Reasoning context ignored per M1 Context Isolation. Tree pinned: `git rev-parse --short HEAD` → `2d40bf634` — matches the pin.

Structural counts re-measured: `grep -c '^### REQ-' spec.md` → **13**; `grep -c '^### AC-F-' acceptance.md` → **24**; unique AC tokens → `AC-F-001…AC-F-024`, contiguous. **AC budget did not exceed 25** (24/25, unchanged from iteration 2 — the N1 fix added no AC, as it said it would not). `version: "0.4.0"`.

Scope: the two iteration-2 blocking items (N1, N2), the orchestrator-directed sibling `§E.1` alignment, a regression check over the four things iteration 2 recorded as *right*, and the three breakage risks the brief named.

---

## §1 Claim

1. **N2 is fully closed.** `AC-F-013` now sits in M3's exit list with a stated reason. Full AC↔milestone coverage re-derived independently: all 24 ACs appear in at least one milestone.
2. **N1 is HALF closed.** The requirement layer is fixed correctly (REQ-10 binds title **and** body and enumerates the four prose-surface obligations). The observation layer is not: **the decisive test still answers YES.** A delivered skill body that pipes only the body — and writes its `gh issue create` line with `--title`, which the current skill body's shape makes the natural form — satisfies all five of AC-F-019's greps and all 24 ACs, and the title reaches GitHub unscrubbed. This is the same reachable state N1 named, now reached by a differently-shaped body.
3. **Two of the five new greps are vacuous against the untouched tree** — measured, not predicted: `feedback-draft` → 1 and `conversation_language` → 8 in **both** copies today, before any implementation. Obligations ③(D4 branch) and ④(label language) therefore cannot be observed to be missing.
4. **AC-F-019 carries a new unexecutable instruction** — a supplementary grep for the Korean literals `종료 코드` and `60초` against a skill surface that ships **English-only** with a template mirror. The same author removed exactly this defect from the sibling SPEC **in this same commit**, with the reasoning written out; it survived here.
5. **Sibling `§E.1` alignment is substantially carried** — five clauses and the `depends_on` trade-off record both present, near-verbatim. Three residual asymmetries, all minor, and one self-contradiction in the section's own opening sentence (`6종` against a 9-row table).
6. **No regression** on any of the four items iteration 2 recorded as right. In particular the enforcement-honesty language is byte-unchanged — the revision widened the control's scope to the title **without** upgrading the claim to "masking is enforced". This was the brief's most serious named risk and it did not materialize.

---

## §2 Evidence — verbatim observations

### 2.1 N2 — AC↔milestone coverage, re-derived

`grep -n 'Exit' plan.md` (milestone exit lines only):

| Milestone | Exit AC set |
|---|---|
| M1 | F-001 |
| M2 | F-005 ~ F-011, F-014, F-024 |
| M3 | **F-008, F-012, F-013** |
| M4 | F-015 ~ F-018 |
| M5 | F-003, F-004 |
| M6 | F-002, F-019 ~ F-022 |
| M7 | F-023 (web half) |
| M8 | F-023 (template half) |

Union = F-001…F-024, **all 24, none orphaned**. `git diff 0375e6842 2d40bf634 -- plan.md` shows the entire plan.md delta is exactly this: M3's exit line plus a three-line justification (`F-013을 빠뜨리지 않는다`) naming the silent-miss failure mode. **N2: RESOLVED.**

Two ACs appear in more than one milestone, both defensibly:
- **F-008** in M2 (inside `F-005 ~ F-011`) and M3. Its three assertions straddle the boundary — `Body` byte-identity and empty `Findings` are masking (M2), `Verdict == "ok"` is classification (M3). Overlap by construction, not by error. Reported per instruction; **not a defect**.
- **F-023** split M7/M8, declared as halves in both exit lines. Intentional.

### 2.2 N1 — requirement layer: CLOSED

`spec.md:182` now reads `**이슈 제목과 본문을** moai feedback scrub에 통과시킨 뒤`, and `spec.md:186-191` adds a `[HARD]` block enumerating the four obligations, each with its consequence. Obligation 1 is stated precisely: `스크러버 호출에 --title 이 실려 있을 것` — **the call**, not the file. Hold that phrasing; §2.3 is about whether anything observes it.

### 2.3 N1 — observation layer: NOT CLOSED

AC-F-019's command block (`acceptance.md:264-269`), run per copy:

```
grep -c 'moai feedback scrub' "$SKILL"    # ①
grep -c -- '--title' "$SKILL"             # ②
grep -c 'verdict' "$SKILL"                # ③
grep -c 'feedback-draft' "$SKILL"         # ④
grep -c 'conversation_language' "$SKILL"  # ⑤
```

Measured against the **untouched** tree, both copies (`.claude/skills/moai/workflows/feedback.md`, `internal/template/templates/.claude/skills/moai/workflows/feedback.md`):

| grep | source copy | template mirror | discriminating? |
|---|---|---|---|
| ① `moai feedback scrub` | 0 | 0 | **yes** |
| ② `--title` | 0 | 0 | **no — see below** |
| ③ `verdict` | 0 | 0 | yes (existence only) |
| ④ `feedback-draft` | **1** | **1** | **NO — already passes** |
| ⑤ `conversation_language` | **8** | **8** | **NO — already passes** |

**④ and ⑤ are vacuous.** The pre-existing matches are unrelated to the obligations they claim to observe:
- `feedback.md:40` — `Offer to save the drafted issue body locally (e.g. under .moai/state/feedback-draft-<timestamp>.md)`. That is the *existing* draft path, not REQ-9's failure-class **branch**. A skill body that never mentions the queue at all still returns 1.
- `feedback.md:38,100,102,103,104,109,120,146` — eight pre-existing `conversation_language` mentions covering issue title language and template headers. The D11 obligation is about the **confirmation gate's option labels and findings summary**, which need not exist for the grep to pass.

**② is unanchored.** The grep matches the bare token anywhere in the file. The skill body already documents the submission command (`feedback.md:118` — `gh issue create --repo <resolved-target>`) and lists `Title: User-provided title` as an input (`:87`). The natural way to write the masked-submission step is `gh issue create --repo <x> --title "<masked title>"` — which returns ≥1 for ② while the scrubber never receives the title. AC-F-019 imposes no co-location requirement between `--title` and the scrub invocation (the one co-location check it does carry, `grep -n`, is for the verbatim-exception sentence against ①). So the AC's command **cannot distinguish** the requirement REQ-10 clause 1 actually states from its violation.

**The brief's decisive test, answered:** *could a delivered skill body that pipes only the body still satisfy every AC?* — **Yes.** Body-only scrub + a `gh issue create --title` line ⇒ ①✓ ②✓ ③✓ ④✓(vacuously) ⑤✓(vacuously), and every other AC is untouched by skill-body content. This is precisely the state AC-F-019's own rationale paragraph claims to have eliminated (`acceptance.md:271` — `②가 없으면 D1은 문서상으로만 닫힌다`). The rationale is right about the mechanism and wrong about whether its grep implements it.

Note the direction of the finding: the requirement text is now correct in three places, so a careful implementer will very likely do the right thing. What is missing is the *detector*, and the SPEC states its own criterion for needing one (`spec.md:186` — `본문이 그중 하나를 빠뜨려도 나머지 AC가 전부 통과하는 상태를 만들지 않기 위해`). It is graded against that stated criterion, not against an imported standard.

### 2.4 The Korean-literal supplement (brief risk #1, materialized)

`acceptance.md:273`:
> ③은 … `grep -n '종료 코드\|verdict\|60초' "$SKILL"` 이 세 조건을 각각 담은 줄을 보여야 한다.

The target files are the skill body and its **distributed template mirror**, both English-only (CLAUDE.md §9: Skills instructions always English; §15 / CLAUDE.local.md §2.1: template neutrality). Both branches fail:
- write the clause correctly in English (`exit code`, `60 seconds`) → the check finds nothing and this MUST-PASS AC reports a false failure;
- satisfy the check literally → Korean prose lands in a template shipped to all users, breaking the language rule.

This is not a novel diagnosis on my part — it is the author's own, applied to the sibling and not to this SPEC. `git diff 0375e6842 2d40bf634 -- ../SPEC-TODO-ENABLE-FLAG-001/acceptance.md` removes `grep -c '명시적' …` and replaces it with a paragraph reading `한국어 리터럴을 통과 조건에 넣었는데, 대상은 영어 전용 표면이다 … 스킬을 규칙대로 영어로 쓰면 이 MUST-PASS AC가 거짓 실패하고, AC를 통과시키려 한국어를 넣으면 배포 템플릿의 언어 규칙을 어긴다`, delegating to a behavioural round-trip instead. The same commit that wrote that paragraph left the same shape at `acceptance.md:273` here.

### 2.5 Sibling `§E.1` alignment

Read both sections in full. Carried correctly: the `[HARD]` text-conflict framing, all **five** resolution clauses (second-lander owns resolution / both items preserved / no relocation / re-run is the only evidence / revert-and-report on failure), and the **`depends_on` trade-off record** — the full "this is a choice, not an observation; serialization would have erased the 9-file risk; concurrency was chosen; reversing means adding `depends_on` to both SPECs" paragraph, near-verbatim. The brief asked specifically whether the trade-off record came over: **it did.**

Three asymmetries remain, all minor:

| # | This SPEC | Sibling | Effect |
|---|---|---|---|
| a | Clause 4 adds a `-v` / `=== RUN` obligation naming `TestFeedbackAutoSubmitQuestion` **and** `TestTodoEnabledQuestion` | Clause 4 delegates to `AC-T-011`, no `=== RUN` obligation | The resolution discipline differs by which document the second-lander reads — the exact hazard the alignment sentence exists to prevent, in miniature. This SPEC is the stricter one. |
| b | Inventory row: sibling adds `항목 1개` (unconditional) | Its own row says `(싣는 경우) 항목 1개` (conditional) | Each document states the other's obligation differently from how that document states it. |
| c | Names the rule with no D-number | Names it `(D4)` | Cosmetic. |

And one self-contradiction inside the section: `spec.md:247` opens `같은 파일 6종을 동시에 건드린다` above a **9-row** table, while the same section says `같은 9개 파일` (`:263`) and `공유 파일 위험 9종` (`:273`). Pre-existing (present at `0375e6842:240`), untouched by the alignment rewrite, and in scope because that rewrite passed directly over it.

The section's closing claim — `이 규칙은 형제 SPEC과 **같은 내용이다**` — is an overstatement given (a) and (b), though it is true of the five clauses and the trade-off record, which is what matters operationally.

### 2.6 Regression check — the four items iteration 2 recorded as right

| Item | Status | Evidence |
|---|---|---|
| **Enforcement honesty** (the brief's named serious risk) | **INTACT** | `git diff 0375e6842 2d40bf634` touches neither `design.md` nor `§E.3` nor `plan.md` AP-12. `§E.3` still reads `강제는 규약 수준이다 … "이제 마스킹이 강제된다"고 적는 것은 미검증 주장이다`; `plan.md:171` AP-12 unchanged. The scope widened to the title and the claim did **not** move. |
| **D2's generalized repair** (AC-F-023, `-v` / `=== RUN`, §D.3 precondition, closure gate) | **INTACT** | `acceptance.md` diff is confined to the AC-F-019 block and its matrix row. AC-F-023 byte-unchanged. |
| **D1b `shall` / `shall not` precision** (REQ-4 sub-clause) | **INTACT** | `spec.md` diff touches only frontmatter, REQ-10, §E.1, and HISTORY. REQ-4 unchanged. |
| **D4's mechanism-not-sentence answer** (REQ-9 failure-class table) | **INTACT** | Unchanged; `spec.md:175` still carries `즉 gh issue create 실패는 큐가 쓰인다`. |

**Nothing was folded to make budget room** (brief risk #2): AC count 24 before and after; the `acceptance.md` delta is +27/−12, entirely inside one AC block and its matrix row, and is pure addition in substance. No AC lost an assertion.

### 2.7 Mechanical lint (domain tool)

```
$ ~/go/bin/moai spec lint .moai/specs/SPEC-FEEDBACK-AUTO-SUBMIT-001/spec.md
✓ No findings — all SPEC documents are valid
rc=0
```

---

## §3 Baseline-attribution

Every observation was made **in this run, against this tree**: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t170`, HEAD `2d40bf634` (verified before the first read). Commands: `grep`, `sed`, `awk`, `wc`, `git diff 0375e6842 2d40bf634`, `git show 0375e6842:…`, `git log`, and one `~/go/bin/moai spec lint`. The grep counts in §2.3 are literal `grep -c` outputs against the two live skill files, not estimates. No test was executed.

Tier L input contract satisfied: all 5 artifacts read (`spec.md`, `plan.md`, `acceptance.md` at v0.4.0; `design.md`, `research.md` unchanged since v0.3.0 and re-read for the regression check). Sibling `SPEC-TODO-ENABLE-FLAG-001` `§E.1` and its iteration-3 `acceptance.md` delta read for the alignment comparison — closing an iteration-2 gap.

Per the delta-scoped re-audit contract, the load-bearing claims verified TRUE in iterations 1 and 2 were **not** re-derived.

---

## §4 Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — 13 entries, `REQ-1`…`REQ-13`, sequential, no gaps, no duplicates (`spec.md:82,88,94,108,131,137,147,155,163,180,193,202,210`). Corroborated by clean `moai spec lint`.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in `spec.md`); `Given/When/Then` entries in `acceptance.md` are the verification layer, graded under Group 4. The only requirement text changed this iteration is REQ-10, whose lead sentence stays ubiquitous with a named subject (`… 템플릿 미러는 … [HARD] 조항을 담아야 한다`) and whose four added clauses are enumerated obligations under it, not free-standing requirements. No regression.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types (`spec.md:2-16`); `version: "0.4.0"` quoted. `era` / `tier` / `related_specs` are additive. No rejected snake_case alias.
- **[N/A] MP-4 language neutrality** — single-project Go scope. N/A auto-passes. (Note: the §2.4 Korean-literal finding is a *template-neutrality* hazard, not an MP-4 multi-language-enumeration failure; it is graded under Testability.)
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — 3 external refs, all resolve, none in {retired, superseded, archived}: `SPEC-INVOCATION-MODEL-001` `completed`, `SPEC-WEBCONF-SIMPLIFY-001` `completed`, `SPEC-TODO-ENABLE-FLAG-001` `draft`. The completed-SPEC decision reversal stays reconciled (REQ-12 + §D D5 + `plan.md` M7 commit-message obligation).
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → `0`. Auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → no output.

**All seven must-pass criteria pass, for the third iteration running.**

---

## §5 Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.86 | 0.75→1.0 | REQ-10 is now unambiguous where iteration 2 found it split. §E.1's five clauses are procedurally precise. Held back by four wording defects, deliberately carried or newly noticed: `6종` against a 9-row table (N7), `같은 내용이다` overstating an asymmetric alignment (N8), plus the unfixed N3 (four edit points, three files) and N5 (`그대로 반환` ↔ `사본을 반환`). All are wording; none change what an implementer builds. |
| Completeness | 0.90 | 0.75→1.0 | §E.1 gained the resolution procedure and the trade-off record it lacked — the largest completeness gain this iteration. Every REQ still has a home; the four prose obligations are now enumerated in the requirement that governs them. Residual: N6 (the end-of-input over-mask cost is still not a fourth §E.3 residual risk) and N4 (`design.md §6`→`§10`), both consciously deferred. |
| Testability | 0.80 | 0.75 band | Genuine improvement over iteration 2's single grep: the four obligations are now named and observed **per copy**, and ① and ③ discriminate. But of five observations, **two are vacuous against today's tree** (④ ⑤), **one is unanchored** (②, the one the AC calls its core), and the ③ supplement is **unexecutable in the target's language**. The AC reports PASS on a non-compliant delivery of three of the four obligations it exists to observe. Placed at 0.80, not lower: 23 of 24 ACs remain exemplary and mechanically sound, and this AC did improve. Not higher: it fails the criterion the SPEC itself sets for it. |
| Traceability | 0.90 | 0.75→1.0 | N2 closed and closed well — F-013 restored **with the reason it must not be dropped again**, which repairs the failure class rather than the instance. Coverage re-derived from scratch: 24/24 ACs in a milestone, no orphan AC, no uncovered REQ, F-008's dual listing structurally justified. Not 1.0 only because the AC-F-019 ↔ REQ-10 link, while present in the matrix, does not carry the obligations it claims (§2.3). |

Aggregate (harmonic mean, per `agent-common-protocol.md` § Skeptical Evaluation Stance):
4 / (1/0.86 + 1/0.90 + 1/0.80 + 1/0.90) = 4 / 4.63501 = **0.863**

**0.863 ≥ Tier L threshold 0.85 — the aggregate clears.**

Score trajectory: 0.75 → 0.837 → 0.863. **No regression**; the LEAN STOP-escalation clause does not fire.

### Why the verdict is still FAIL

The Retry Loop Contract is explicit and score-independent: *"Unresolved defects from a prior iteration are automatically FAIL regardless of other scores."* N1 was iteration 2's live blocking defect, and it was defined by a reachable state, not by a wording. That state is still reachable (§2.3). The requirement half is closed; the observation half — the half N1 was actually about — is not.

I am recording the aggregate honestly at 0.863 rather than deflating a dimension to manufacture agreement between the score and the verdict. The two genuinely disagree, and the disagreement is the finding: **this SPEC is now good enough on every axis a rubric measures, and its one security control still cannot be observed to have been delivered.**

---

## §6 Defects Found (structured defect-list)

**N1-a — AC-F-019's ② is unanchored: `grep -c -- '--title'` passes on the submission command line, so it cannot distinguish a scrubbed title from an unscrubbed one.** — `acceptance.md:265` against `spec.md:188` (`스크러버 호출에 --title 이 실려 있을 것`) and the live `feedback.md:118` (`gh issue create --repo <resolved-target>`) + `:87` (`Title: User-provided title`). A skill body that pipes only the body and writes `gh issue create --repo <x> --title "<title>"` returns ≥1 and passes. — Severity: **major** — Class: **blocking** — Required fix: anchor to the call, e.g. `grep -cE 'moai feedback scrub[^\n]*--title' "$SKILL"  # >= 1`, or require ② and ① on the same line via `grep -n` and assert equal line numbers.

**N1-b — AC-F-019's ④ and ⑤ are vacuous: both already return ≥1 on the untouched tree, in both copies.** — Measured: `grep -c 'feedback-draft'` → 1 / 1 (matches the pre-existing draft-save line, `feedback.md:40`, not REQ-9's failure-class branch); `grep -c 'conversation_language'` → 8 / 8 (pre-existing title/header language rules, `:38,100,102,103,104,109,120,146`, not D11's gate-label obligation). Neither obligation can be observed to be **missing**. — Severity: **major** — Class: **blocking** — Required fix: grep for tokens that do not exist today, e.g. ④ `grep -c 'queue.json'` (or a co-location of `gh issue create` with the queue path) and ⑤ a check tied to the confirmation gate rather than to the word alone. Before adopting any replacement token, **run it against the untouched tree and require 0** — that check is what was missing here.

**N1-c — AC-F-019's ③ supplement gates a MUST-PASS AC on Korean literals in an English-only, template-mirrored surface.** — `acceptance.md:273` (`grep -n '종료 코드\|verdict\|60초' "$SKILL"`). Writing the clause correctly in English makes this MUST-PASS AC fail; satisfying it literally puts Korean into a distributed template (CLAUDE.md §9; CLAUDE.local.md §2.1 / §15). The same defect was diagnosed and removed from the sibling in this same commit (`SPEC-TODO-ENABLE-FLAG-001` AC-T-004 → delegated to AC-T-005's behavioural round-trip). — Severity: **major** — Class: **blocking** — Required fix: restate the three-condition check with English tokens the clause will actually contain (`exit code` / `verdict` / `60`), or follow the sibling's precedent and delegate the sentence-count check to a behavioural observation.

**N7 — `§E.1` opens with `같은 파일 6종` above a 9-row table, contradicting the same section twice (`같은 9개 파일`, `공유 파일 위험 9종`).** — `spec.md:247` vs the table at `:249-259` and `:263`, `:273`. Pre-existing (`0375e6842:240`); the alignment rewrite passed over it. — Severity: **minor** — Class: **optional** — Required fix: `6종` → `9종`.

**N8 — the sibling-alignment claim `이 규칙은 형제 SPEC과 같은 내용이다` overstates: clause 4 and the inventory row differ between the two documents.** — `spec.md:271` vs `SPEC-TODO-ENABLE-FLAG-001` §E.1. Clause 4 here adds a `-v` / `=== RUN` obligation the sibling lacks; the inventory row is unconditional here and conditional there. The second-lander's discipline therefore depends on which document they read — a small instance of exactly what the sentence forbids. — Severity: **minor** — Class: **optional** — Required fix: either mirror clause 4 into the sibling (preferred — this SPEC's version is the stronger one) or soften the claim to name the five clauses and the trade-off record as the shared part.

**N3, N4, N5, N6 — carried unchanged from iteration 2, deliberately.** — `spec.md:204` (four edit points / three files), `spec.md:270` (`design.md §6` → `§10`), `spec.md:145` (`그대로 반환` ↔ `사본을 반환`), `§E.3` (over-mask cost not listed). The HISTORY entry states the reason: not making non-blocking edits at the ceiling iteration. **That reasoning is sound and I endorse it** — each is a wording fix with no implementation consequence. — Severity: **minor** — Class: **optional** — Required fix: as recorded in iteration 2; fold into the run phase's first documentation touch.

---

## §7 Gaps (what this audit did NOT observe)

- **No test was executed.** Every "this test would pass/fail" statement in the SPEC remains a prediction.
- **§2.3's `gh issue create --title` case is an argument about what AC-F-019's command can and cannot distinguish, verified against the AC text and the current skill file — not an observation of a future skill body.** The load-bearing half (the grep does not require co-location with the scrub call) is textual and verified; the likelihood half (that an implementer writes `--title` on the gh line) is a judgment grounded in `feedback.md:118` and `:87`.
- I did not re-derive the load-bearing claims from iterations 1 and 2 — instructed delta scope.
- I did not re-audit AC bodies untouched by this delta (F-001…F-018, F-020…F-024) beyond the fold check.
- I did not audit the sibling SPEC as a whole — only its `§E.1` and its iteration-3 `acceptance.md` delta, for the two comparisons this brief required.
- I did not run `make build`, `golangci-lint`, `go test`, or `GOOS=windows go vet`.
- I did not verify that `design.md` / `research.md` remained appropriate under REQ-10's widened text beyond confirming they were unmodified and re-reading the affected sections.

---

## §8 Residual-risk

- **The gap between what this SPEC says and what it can detect is now its defining characteristic.** Three documents state the title obligation; four obligations are enumerated in the requirement that governs them; and the single AC that observes them reports PASS on a delivery that honours one of the four. The risk is not that the run phase gets it wrong — the instructions are clear enough that it probably will not. The risk is that **nobody would know if it did**, and the failure is silent, public, and irreversible: a credential in a GitHub issue title.
- **The vacuous-grep failure mode will recur unless the lesson is written down.** Both N1-b greps were chosen by reading the obligation and picking a plausible token, without running the token against the tree first. That is a one-command check (`grep -c <token> <file>` must return 0 before the work starts) and its absence produced two false observations in the same block.
- **Enforcement remains convention-level, by the SPEC's own honest admission**, and widening the scrubber to the title enlarged what that convention carries from one obligation to four. §E.3 and AP-12 still say this plainly; that honesty is the SPEC's strongest property and it survived three iterations of scope growth.
- **The classifier (REQ-7 / `design.md` §6) still has no precedent and no calibration.** Unchanged across all three iterations; correctly deferred to post-release reports.
- **The two-SPEC shared-file discipline is enforced by nothing mechanical**, and after this iteration the two documents' versions of it are near-identical but not identical (N8).

---

## §9 Recommendation — escalation view (Tier L ceiling reached)

**FAIL at aggregate 0.863 against the Tier L threshold 0.85**, on the score-independent unresolved-prior-defect clause, with all seven must-pass criteria passing. This is the third iteration; the orchestrator escalates rather than iterating again.

### Defect history across three iterations

| Iteration | Verdict | Score | Blocking defects | Outcome |
|---|---|---|---|---|
| 1 | FAIL | 0.75 | 11 (incl. MP-2 FAIL, D1 title channel, D1b detector asymmetry) | 10 of 11 closed |
| 2 | FAIL | 0.837 | 2 (N1 observation gap, N2 orphaned AC) | N2 closed; N1 half closed |
| 3 | FAIL | 0.863 | 3 (N1-a, N1-b, N1-c — all inside one AC block) | — |

**No stagnation.** N1 is not the same defect appearing unchanged three times: iteration 2's N1 was "one grep observes four obligations"; iteration 3's is "five greps, three of which do not discriminate". The revision moved, and moved in the right direction — it simply did not land. Progress is real and monotonic on every axis.

### The two choices, and the debt each inherits

**Choice A — PASS-with-debt (recommended).** All three remaining defects live inside **one AC block** (`acceptance.md:264-273`) and are three grep-expression edits. None touches a requirement, a milestone, a design decision, or the AC count. The SPEC is structurally sound: 7/7 must-pass, complete REQ→AC→milestone traceability, an honest enforcement claim, and a repaired detector-vs-rewriter asymmetry.

The debt the run phase inherits under A, stated precisely so it can be written into the SPEC rather than remembered:

> **AC-F-019 as written passes on a skill body that (i) never passes `--title` to the scrubber, (ii) never mentions the queue branch, and (iii) never binds gate labels to `conversation_language` — and its ③ supplement cannot be satisfied without breaking the template language rule.** Before M6 is judged complete, the four greps must be re-authored, and each replacement token must be shown to return **0** against the pre-implementation tree (`git show <base>:<skill path>`). Until then, M6's exit is a claim, not an observation.

This debt is discharged in one edit at the start of M6 and is cheap **only if it is written down**. Verbally noted, it is exactly the kind of item that evaporates between plan and run — and its failure mode is a credential in a public issue title.

**Choice B — hold the SPEC and fix before run entry.** Costs one revision round beyond the ceiling (an explicit user override per the Retry Loop Contract) to change three grep lines. Justified only if the operator judges that a security control's detector must be correct *in the document* before implementation begins — a defensible position for this specific SPEC, since the control in question is its entire security rationale. Under B the run phase inherits **no** debt on this axis; the residual debt is then only the optional wording items (N3, N4, N5, N6, N7, N8).

**My recommendation is A, with the debt paragraph pasted verbatim into `acceptance.md` above AC-F-019 and into `plan.md` M6**, so the obligation travels with the milestone that discharges it. The SPEC has earned run entry on every axis except the one that is cheapest to fix — and the fix is better made by the person about to write the skill body than by another planning round.

**VERDICT: FAIL 0.863** (Tier L threshold 0.85; aggregate clears, contract clause binds — see §5)
