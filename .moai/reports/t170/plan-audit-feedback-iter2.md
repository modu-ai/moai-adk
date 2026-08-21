# SPEC Review Report: SPEC-FEEDBACK-AUTO-SUBMIT-001

Iteration: 2/3
Verdict: **FAIL**
Overall Score: **0.84** — graded against the **Tier L PASS threshold 0.85**
(`.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier, row `L`).

Reasoning context ignored per M1 Context Isolation. Tree pinned: `git rev-parse --short HEAD` → `0375e6842` — matches the pin; nothing was audited against a moved HEAD.

> **Note on notation in this report**: PEM block markers are written descriptively (`the PEM begin-header`, `the PEM end-terminator`) rather than as literals. Writing the literals into this file was refused by the PreToolUse hook with `Content contains sensitive data` — the same live-tree behavior `acceptance.md` records for AC-F-008. No claim below depends on the literal form.

Scope: the enumerated defect delta from `.moai/reports/t170/plan-audit-feedback.md` (iteration 1) plus a regression check over it, plus a probe for what the revision broke. The seven load-bearing claims verified TRUE in iteration 1 were **not** re-derived — nothing in the revision depends on a different reading of them.

Structural counts re-measured independently: **REQ 13 / AC 24** — the lead's numbers confirmed. `grep -c '^### REQ-' spec.md` → 13; `grep -c '^### AC-F-' acceptance.md` → 24; the unique-token scan yields 24 contiguous IDs, AC-F-001…AC-F-024. One AC slot remains under the Tier L ceiling of 25.

---

## §1 Claim

1. **Ten of the eleven iteration-1 defects are genuinely resolved**, verified against the tree rather than accepted from the HISTORY narration. D2 and D3 in particular were re-measured mechanically, not read.
2. **MP-2 now passes.** REQ-12 is event-driven GEARS with a named subject, and decision D5 was resolved (option A confirmed, option B deleted) rather than deferred — so nothing needed a `[NEEDS CLARIFICATION]` marker. All seven must-pass criteria pass.
3. **D1 is closed at the Go contract layer but is only partly carried at the enforcement layer**, and nothing in the AC set observes the enforcement side. REQ-3's [HARD] binds masked-title-only submission; REQ-10 — the requirement that dictates what the skill clause must say — still enumerates the **body only**, and AC-F-019, the sole observation of the skill body, is a single `grep -c 'moai feedback scrub'`. A run phase can satisfy every one of the 24 ACs while shipping a skill body that never passes `--title`.
4. **The D7 fix introduced a new gap**: correcting the three milestone AC ranges dropped **AC-F-013** out of every milestone exit list. It now belongs to no milestone.
5. The revision did **not** trade observation resolution for AC budget in the way the lead suspected — the AC-F-008 fold is sound. And the end-of-input masking fallback is a deliberate fail-safe, not a leak-direction hazard, though its over-mask cost is unenumerated.

---

## §2 Evidence — verbatim observations

### Regression check over the iteration-1 defect delta

| ID | iter1 required fix | Status | Observed evidence |
|---|---|---|---|
| **D1** | scrub the title, or scope it out | **PARTIAL** | Extend-the-scrubber was taken, and carried consistently through the code-facing surfaces: `spec.md` §B In Scope (`스크럽 대상은 제목과 본문 둘 다`), REQ-3 (`--title` input surface, `title` output field, `findings` carries `where`), `design.md` §1 diagram (`stdin: 본문 · --title: 제목`) plus a [HARD] paragraph, §4 (`Finding.Where`, `Result.Title`), §7 (the gate shows **마스킹된 제목 + 본문 전문 + findings 요약(위치 포함)**, with an explicit `제목을 빼면 D1이 닫히지 않는다`), AC matrix rows F-003 / F-006, and `plan.md` M5. **The gate's display in §7 is NOT the half-fix** the brief anticipated — it carries the title. The half-fix is elsewhere: see N1. |
| **D1b** | state the rewrite-span rule as a requirement; add the two observations | **RESOLVED** | `spec.md` REQ-4 carries a `[HARD] REQ-4 하위 조항` with three `shall`/`shall not` clauses: replace only when the pattern anchors the secret; marker-anchored patterns mask **through the PEM end-terminator**, and to end-of-input when no terminator is found (`잘린 키 블록을 통과시켜서는 안 된다`); case-sensitive recompilation for AWS-key-shaped patterns, with `(?i)` reuse in the rewrite path explicitly prohibited. It is a requirement, not prose. AC-F-024 asserts `Result.Body` retains neither the begin-header, the end-terminator, nor any intervening line (case 1) and end-of-input masking (case 2). AC-F-008's fourth case asserts byte-identity of `Result.Body` for a lowercase AWS-prefix-shaped run — the over-masking direction. |
| **D2** | replace the two vacuous `-run` selectors; require `-v` | **RESOLVED** | All five selectors in AC-F-023 resolve against the tree: `grep -c 'func TestI18nKeySetParity' internal/web/schema_label_test.go` → `1`; `grep -c 'func TestRouteForSectionTable' internal/settings/sectionroute_test.go` → `1`; `grep -c 'func TestExcludedSectionsAllRejected' …` → `1`; `grep -c 'func TestSchemaCurrentValuesReadsAllSections' internal/settings/schema_sections_test.go` → `1`; `grep -n 'func TestScopeContract' internal/web/scope_contract_test.go` → `22:…EditableSections`, `67:…Exclusions` (prefix match, non-vacuous). AC-F-023 carries `[HARD] -v 출력에서 위 5개 테스트 이름이 각각 === RUN 으로 찍히는지 확인한다` plus three pre-judgment `grep -c 'func …'` checks, and the closure gate (§D.4) repeats it. I additionally re-checked the other selectors naming **pre-existing** tests: `TestQuestionOrder` (`questions_test.go:87`), `TestReconfigureQuestions` → `TestReconfigureQuestionsOrder:184` (prefix), `TestWizardQuestionTranslationCompleteness` (`translations_completeness_test.go:95`) — all resolve. |
| **D3** | name the accessor mechanism; put the file in scope | **RESOLVED** | REQ-6 `[HARD]` names `func DefaultEnvDenyList() []string` in `internal/sandbox`, forbids copying the names into `internal/feedback` (`shall not`), and states the reason (drift — the same failure AP-4 forbids for pattern sets). Present in `spec.md` §B In Scope line 3, `plan.md` §A module list, `plan.md` M2 (`internal/sandbox/env.go — func DefaultEnvDenyList() []string 신설`), M2 Exit (`go test ./internal/sandbox/...`), and the M9 sweep + `GOOS=windows go vet` list. Tree confirms the accessor does not yet exist: `grep -rn 'DefaultEnvDenyList' internal/ --include='*.go'` → no output; `internal/sandbox/env.go:32 var defaultDenyList = []string{` still unexported. |
| **D4** | say whether the queue replaces/wraps/coexists, and which one a `gh` failure writes | **RESOLVED** | REQ-9 `[HARD]` splits by **failure class** with a two-row table: pre-submission failure (`gh auth status`, rate limit) → the existing `feedback-draft-<ts>.md` (pre-scrub original); post-attempt failure (`gh issue create`) → the new `queue.json` (masked). Followed by the explicit answer — `즉 gh issue create 실패는 큐가 쓰인다` — and a `shall not` forbidding the resend path from mistaking a draft file for a queue entry. `plan.md` M4 restates the warning with the leak consequence spelled out. |
| **MP-2 / D8** | GEARS-rewrite REQ-12; resolve or mark D5 | **RESOLVED** | REQ-12: `사용자가 웹 설정 화면을 열면, 웹 콘솔은 feedback.auto_submit을 불리언 토글로 렌더하고 그 변경을 .moai/config/sections/feedback.yaml에 영속해야 한다` — event-driven, subject `웹 콘솔`, deferral clause gone. §D 결정 D5 is retitled `해소됨: 선택지 A`, option B deleted, with the reversal cost stated (two pinning tests updated + commit-message record). `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → no output; the only occurrence in the directory is `spec.md:292`, a HISTORY sentence stating resolution was chosen over the marker. |
| **D7** | correct the three milestone AC ranges | **RESOLVED, but see N2** | M4 Exit → `F-015 ~ F-018` (log content/perm, log fail-open, enqueue, resolve) ✓; M6 Exit → `F-002, F-019 ~ F-022` ✓; M7 Exit → `F-023의 웹 절반`, M8 Exit → `F-023의 템플릿 절반` ✓. Re-derived from `acceptance.md` §D, not taken from the new numbers. |
| **D5** | name the root-resolution rule | **RESOLVED** | REQ-3 `프로젝트 루트 해석(D5)`: upward walk to the `.moai/` marker, `--root <path>` supersedes it (named as the `t.TempDir()` injection path the ACs assume), and a `shall not` — an unresolvable root skips log/queue writes (fail-open) but must not block scrubbing. `design.md` §9 gains the matching row; `plan.md` M5 declares the flag. |
| **D6** | two failure rows with a stated bound; correct the `paths.Home()` row | **RESOLVED** | `design.md` §9 gains `스크러버 무응답(행) | 호출자 타임아웃 **60초** | 제출 금지` — a numeric bound, not "a timeout" — and `stdout 파싱 불가(절단·비JSON, 종료 코드 0) | jq 파싱 실패 또는 verdict 필드 부재 | 제출 금지`, closing with **부재하는 `verdict`는 `ok`가 아니다**. The `paths.Home()` row now reads `HOME 미설정 시 os.UserHomeDir()로 폴백해 대개 성공한다(internal/paths/paths.go:55-60)`. The skill clause is now three sentences (`design.md` §9 tail, mirrored in `plan.md` M6 (c)). |
| **D11** | one sentence requiring `conversation_language` | **RESOLVED** | `design.md` §7 `[HARD] 라벨과 요약은 conversation_language로 낸다(D11)` — states the Korean table is meaning-not-literal, cites `askuser-protocol.md`, binds both the source skill and the template mirror, and fixes the mirror's example label set as English. Repeated in `plan.md` M6 (b). |
| **D9 / D10 / D12** | optional | **RESOLVED** | `research.md:15` records `validator.go:155` as `조사됨 — 미채택` with the reason; REQ-4 narrowed to `hook의 목록에는 없다`. `research.md:19` corrects the `~` grep count to **3건** and names `shell/config.go:222`. `spec.md:4` → `version: "0.3.0"` (quoted). |

### Mechanical lint (domain tool)

```
$ ~/go/bin/moai spec lint .moai/specs/SPEC-FEEDBACK-AUTO-SUBMIT-001/spec.md
✓ No findings — all SPEC documents are valid
rc=0
```

### Probe 1 — did an AC fold reduce observation resolution? (lead's risk #1)

**No.** The only fold is AC-F-008, which now carries two `Given` blocks and four assertions under one test function. The failure directions stay separable: the first block's three assertions are on different axes (`Body` byte-identity / `Findings` empty / `Verdict == "ok"`), and the fourth case's assertion is a byte-identity check on a distinct input. `acceptance.md` states the reasoning explicitly (`두 단언이 분리돼 있어 실패 시 어느 방향인지 구별된다`), and the two D1b directions live in **different** ACs (under-mask → AC-F-024, over-mask → AC-F-008 case 4), so a failure always names its direction. No other AC absorbed a second failure direction.

**The hook-block indirection does not make AC-F-008 case 4 untestable.** The shape is fully determined — the AWS key prefix written in lowercase followed by 16 alphanumerics — which is exactly the input class the case-insensitively compiled AWS pattern matches and a case-sensitive recompile does not. The assertion (`Result.Body` byte-identical to input) is binary. A naive implementation that reuses `(?i)` fails it; a correct one passes. The AC is falsifiable as written. The `이 케이스를 문서에 리터럴로 적지 않는 이유(실측)` note additionally records the PreToolUse block as a live-tree observation of the very case-insensitivity asymmetry D1b describes — evidence, not an excuse. (This report hit the same hook on the PEM literals; see the note at the top.)

### Probe 2 — can the end-of-input fallback swallow the issue body? (lead's risk #2)

**In the leak direction, no.** The fallback fires only after a marker-anchored pattern matches, and those two patterns (`internal/hook/pre_tool.go:263-264`) require a literal PEM begin-header for a private key or a certificate. A false positive therefore requires the user to have typed a PEM header. The failure mode is over-masking (fail-safe), never under-masking.

**But the over-mask cost is unobserved and unenumerated.** A user writing "the parser choked on a PEM private-key begin-header" with the literal header and no terminator has everything after that point masked to end-of-input — a mangled issue body. AC-F-024's second case **pins that as the intended behavior**, so no AC distinguishes "correct end-of-input fallback" from "a degenerate implementation that masks everything after the first begin-header-like token". §E.3 carries three residual risks and this is not one of them. Recorded as N6 (optional): the tradeoff is deliberately chosen and correctly directed, so it is a disclosure gap, not a design defect.

---

## §3 Baseline-attribution

All observations were made **in this run, against this tree**: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t170`, HEAD `0375e6842` (verified before the first read, matching the pin). Commands used: `grep`, `sed`, `git show e7fb0e1d2:…` (to attribute N2 to the revision rather than to the original), `ls`, and one `~/go/bin/moai spec lint`. No test was executed — see §7.

Tier L input contract satisfied: all 5 artifacts read at v0.3.0 (`spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`).

---

## §4 Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — 13 entries, `REQ-1`…`REQ-13`, sequential, no gaps, no duplicates (`spec.md:82,88,94,108,131,137,147,155,163,180,186,195,203`). Corroborated by clean `moai spec lint`.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in `spec.md`); the `Given/When/Then` entries in `acceptance.md` are the verification layer and were graded under Group 4, not here. The iteration-1 failure was REQ-12 alone; it is now event-driven with the subject `웹 콘솔` (`spec.md:197`). The other twelve are unchanged from the iteration-1 PASS judgment except REQ-3/4/6/9, whose added clauses are all `shall` / `shall not` with named subjects (`스크러버`, `워크플로`).
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types (`spec.md:2-16`); `version: "0.3.0"` now quoted, closing D12. Two extra fields (`era: V3R6`, `tier: L`) are additive, not rejected aliases. No snake_case alias.
- **[N/A] MP-4 language neutrality** — single-project Go scope; REQ-13 carries the template-neutrality constraint and AC-F-023 mechanically greps for SPEC-ID / REQ-token leakage in the two template files. N/A auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — 3 external refs, all resolve, none in {retired, superseded, archived}: `SPEC-INVOCATION-MODEL-001` `status: completed`, `SPEC-WEBCONF-SIMPLIFY-001` `status: completed`, `SPEC-TODO-ENABLE-FLAG-001` `status: draft`. The completed-SPEC decision reversal is reconciled explicitly (REQ-12 + §D D5 + `plan.md` M7 AP-9 commit-message obligation).
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → `0`. Auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → no output. No open marker.

**No must-pass failure. The verdict below rides entirely on the aggregate score.**

---

## §5 Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.88 | 0.75→1.0 | Both iteration-1 ambiguities closed: REQ-12 is no longer passive/deferred, REQ-6 no longer hides the reuse decision. Residual: REQ-10 says `이슈 본문을` where REQ-3 binds title+body (N1); REQ-12 says `편집 지점은 넷이다` then lists three paths (N3); REQ-6 says the accessor `defaultDenyList를 그대로 반환한다` and then parenthetically that it returns a copy (N5). All three are wording, not interpretation-changing. |
| Completeness | 0.90 | 0.75→1.0 | All three iteration-1 holes filled with substance, not gestures: the title channel (REQ-3 [HARD] + §B + design §1/§4/§7), root resolution + timeout + unparseable-stdout (REQ-3 + design §9, with a **60초** numeric bound), and the queue↔draft reconciliation (REQ-9's failure-class table). Residual: §E.3 does not carry the end-of-input over-mask cost as a fourth residual risk (N6). |
| Testability | 0.78 | 0.75 band | The vacuous-selector defect is fully repaired and generalized into a `-v` / `=== RUN` discipline in AC-F-023, §D.3, and the closure gate — the strongest single improvement in this revision. AC-F-024 and AC-F-008 case 4 are both binary and falsifiable despite the literal-avoidance indirection. Held down by N1: **four** skill-body obligations added or strengthened this iteration (title pass-through, the 3-sentence [HARD] clause, the D4 failure-class branch, `conversation_language` labels) are observed by exactly one AC, AC-F-019, whose command is a single `grep -c 'moai feedback scrub'` per file. The AC set grew by one while the prose-surface obligations grew by four. |
| Traceability | 0.80 | 0.75 band | REQ→AC coverage is complete: every REQ-1…REQ-13 appears in the §D matrix, no orphan AC, no uncovered REQ (24 rows, contiguous). But the D7 repair dropped **AC-F-013** out of every milestone exit list (N2) — the unique-token scan of `plan.md` yields 15 tokens covering 23 of the 24 ACs; F-013 is absent. Attributed to this revision: `git show e7fb0e1d2:…/plan.md` shows F-013 was present (inside M4's then-incorrect `F-013~F-017` range). |

Aggregate (harmonic mean, per `agent-common-protocol.md` § Skeptical Evaluation Stance): 4 / (1/0.88 + 1/0.90 + 1/0.78 + 1/0.80) = **0.84**.

0.84 < Tier L threshold **0.85**.

**No score regression** (0.75 → 0.84), so the LEAN STOP-escalation clause does not fire.

---

## §6 Defects Found (structured defect-list)

**N1 — REQ-10 binds the skill clause to the BODY only, and AC-F-019 cannot observe the title pass-through; the enforcement surface of the SPEC's only security control is untested.** — `spec.md:182` (`이슈 본문을 moai feedback scrub에 통과시킨 뒤 그 출력만을 제출하도록`) against `spec.md:98` (REQ-3 [HARD]: `제목은 스크럽 대상이며, 마스킹된 제목만 제출·표시된다`) and `plan.md:113` M6 (a) (`**제목과 본문을 함께 통과시킬 것**(--title)`). The obligation exists at the requirement layer — REQ-3 carries it — so this is not an unnamed leak. But the requirement that dictates the clause's *content* enumerates the body, and `acceptance.md` AC-F-019 observes the skill body with `grep -c 'moai feedback scrub'` ≥ 1 in each of the two copies plus a line-number check for the verbatim-exception sentence. `grep -n -- '--title' acceptance.md` → only `:72` and `:79`, both inside AC-F-003, which tests the **CLI**, not the skill body. Consequence: a delivered skill body that pipes only the body satisfies all 24 ACs, and the title reaches `gh issue create` unscrubbed — the exact path D1 named. The same coarseness leaves three other iteration-2 additions unobserved: the 3-sentence [HARD] clause (`design.md` §9), the D4 failure-class branch (REQ-9 says `스킬 본문(REQ-10)은 이 분기를 명시해야 한다(shall)` — REQ-10 itself never mentions it), and the `conversation_language` label obligation (D11). — Severity: **major** — Class: **blocking** — Required fix: amend REQ-10 to bind `이슈 제목과 본문을`, and extend AC-F-019's command block so each of the four obligations is greppable in **both** copies (e.g. `grep -c -- '--title'` ≥ 1; a grep for the 60-second / `verdict` sentences; a grep for `feedback-draft`; a grep for `conversation_language`). This needs **no new AC** — the budget stays at 24/25.

**N2 — AC-F-013 belongs to no milestone; the D7 repair dropped it.** — `plan.md` M3 Exit (`:88`) reads `AC-F-008, AC-F-012`; M2 Exit (`:80`) reads `F-005 ~ F-011, F-014, F-024`; M4 Exit (`:97`) reads `F-015 ~ F-018`. The unique-token scan of `plan.md` yields `F-001 F-002 F-003 F-004 F-005 F-008 F-011 F-012 F-014 F-015 F-018 F-019 F-022 F-023 F-024` — the union of the ranges covers 23 of 24 ACs, omitting **F-013**. Introduced by this revision, not inherited: `git show e7fb0e1d2:.moai/specs/SPEC-FEEDBACK-AUTO-SUBMIT-001/plan.md` shows M4 previously read `AC-F-013~F-017` (wrong range, but F-013 was inside it). AC-F-013 is the pre-mask classification order guard — `design.md` §3 calls the inverted order `조용한 미탐이며, 테스트로 잡기 어렵다`, `acceptance.md` §D.3 names it as one of the two degenerate-implementation defenses, and §D.6 cites it under Secured. A run-phase implementer working milestone by milestone never runs it. — Severity: **major** — Class: **blocking** — Required fix: append `AC-F-013` to M3's Exit list (`AC-F-008, AC-F-012, AC-F-013`), M3 being the classifier milestone.

**N3 — REQ-12 announces four edit points and lists three.** — `spec.md:199` (`편집 지점은 넷이다: internal/settings/schema_sections.go(필드), internal/settings/sectionroute.go(라우트 + ExcludedSections() 제거), internal/web/schemaform.go(탭 + 패널)`). §D D5 (`:180`) counts four *edits* across those three files (schema field + `sectionRoutes` + `consoleTabs` + `schemaSectionMetas`), while `plan.md` M7 adds `internal/web/assets/i18n.js` as a fourth **file**. The count is defensible under one reading and wrong under the other. — Severity: **minor** — Class: **optional** — Required fix: say `편집은 네 곳, 파일은 셋이다`, or enumerate the four edits.

**N4 — stale cross-reference: `design.md §6` should be `§10`.** — `spec.md:270` (`gh issue create 자체를 Go 명령으로 감싸야 하며, 그 경로는 후속 카드 후보다(design.md §6)`). `design.md` §6 is `취약점 분류기 설계`; the follow-up-card list is §10, which the very next residual-risk paragraph (`spec.md:272`) cites correctly. — Severity: **minor** — Class: **optional** — Required fix: `§6` → `§10`.

**N5 — REQ-6's accessor contract contradicts itself in one sentence.** — `spec.md:145` (`접근자는 defaultDenyList를 그대로 반환한다(사본을 반환해 호출자가 원본을 변경하지 못하게 한다)`). `그대로 반환` and `사본을 반환` are opposite contracts; `plan.md:77` resolves it as a copy. — Severity: **minor** — Class: **optional** — Required fix: drop `그대로` — `접근자는 defaultDenyList의 사본을 반환한다`.

**N6 — the end-of-input masking fallback's over-mask cost is neither enumerated as a residual risk nor observable.** — REQ-4's sub-clause (`종료자를 찾지 못하면 입력 끝까지 마스킹한다`) + AC-F-024 case 2, which pins that behavior as expected. A benign PEM-header mention with no terminator masks the remainder of the issue body. The direction is fail-safe (over-mask, never leak) and deliberately chosen, so this is a disclosure gap: §E.3 lists three residual risks and this is not one, and no AC separates a correct fallback from a degenerate mask-everything-after-the-header implementation. — Severity: **minor** — Class: **optional** — Required fix: one sentence in §E.3 stating the tradeoff (truncated key blocks are never emitted, at the cost of masking the tail of a body that mentions a PEM header without its terminator).

### What this revision got right (recorded so iteration 3 does not regress it)

- **D2's repair was generalized, not patched.** The fix did not merely swap two selector names; it added a `[HARD] === RUN` inspection to AC-F-023, a pre-judgment `grep -c 'func <TestName>'` triple, a §D.3 rule making selector-existence a precondition of *every* AC's judgment, and a closure-gate checkbox. The failure class was closed, not the instance.
- **D1b's requirement text is precise where it needed to be.** Masking through the end-terminator and the `(?i)` prohibition are stated as `shall` / `shall not`, not as guidance, and the two failure directions land in two different ACs.
- **D4's answer is a mechanism, not a sentence.** The failure-class table answers the question the audit asked (*which one does a `gh` failure write?*) explicitly, and `plan.md` M4 restates the leak consequence of confusing the two files.
- **D5 was resolved rather than marked.** Choosing resolution over a `[NEEDS CLARIFICATION]` marker is the stronger outcome, and the reversal cost of option A is paid explicitly (two pinning tests + commit-message record + a stated path back if the operator rejects the reversal).
- **The enforcement-honesty language survived intact.** §E.3, `design.md` §1's `설계상 가장 중요한 한 줄`, and `plan.md` AP-12 are unchanged; the revision did not quietly upgrade the claim to "masking is enforced" while widening the control's scope to the title.

---

## §7 Gaps (what this audit did NOT observe)

- **No test was executed.** Every "this test would fail / pass" statement in the SPEC remains a prediction, as `research.md` §6 itself declares. I verified test **names, file anchors, and selector resolution** — never behavior.
- I did **not** re-derive the seven load-bearing claims from iteration 1 — an instructed scope exclusion; nothing in the revision reads them differently.
- I did not re-audit `acceptance.md` §D.2 bodies for ACs untouched by the delta (F-001, F-002, F-010, F-011, F-012, F-014…F-018, F-020…F-022) beyond one read for the fold probe; their iteration-1 grades stand.
- I did not verify AC-F-020's `ReconfigureQuestions는 12개` count against the tree — outside the delta.
- I did not audit the sibling `SPEC-TODO-ENABLE-FLAG-001`, so §E.1's shared-file table was read but not cross-checked against that SPEC's current text.
- I did not run `make build`, `golangci-lint`, or `GOOS=windows go vet`.

## §8 Residual-risk

- **N1 is the live one.** If iteration 3 amends only REQ-10's wording without extending AC-F-019, the SPEC will *state* the title obligation in three places and still have no mechanical way to detect its omission at the surface where the omission actually happens.
- The control remains enforcement-by-convention, by the SPEC's own admission. Widening the scrubber's scope to the title enlarges what the convention must carry — the prose surface now holds four [HARD] obligations where iteration 1 had one. That is the structural reason N1 matters more than its one-line fix suggests.
- The classifier (REQ-7 / `design.md` §6) still has no precedent and no calibration; the asymmetric-threshold instruction remains unquantified, and §D.5 correctly defers it to post-release reports.
- The two-SPEC shared-file merge discipline (§E.1) is a [HARD] rule enforced by nothing mechanical.

---

## §9 Recommendation

FAIL at **0.84** against the Tier L threshold **0.85**, with **all seven must-pass criteria passing**. This is a near-miss on a revision that closed ten of eleven defects and repaired the worst two — the title channel and the detector→rewriter asymmetry — properly. The gap is two blocking findings, each a one-line-to-one-block edit, and neither requiring a new AC.

MUST-FIX before re-audit:

1. **N1** — amend REQ-10 (`spec.md:182`) to bind title **and** body, and extend AC-F-019's command block so the four skill-body obligations (title pass-through, the 3-sentence clause, the D4 failure-class branch, `conversation_language` labels) are each observed by a grep in **both** copies. AC budget stays at 24/25.
2. **N2** — append `AC-F-013` to `plan.md` M3's Exit list.

Optional (orchestrator's discretion, per M6): N3, N4, N5, N6.

Iteration 3, if run, should be scoped to these two items plus a regression check — not a full re-audit. Should the orchestrator instead choose **PASS-with-debt**, the debt is precisely: *the skill body — the surface that actually decides whether a secret reaches a public issue — is verified by a single grep, and AC-F-013 is in no milestone.* Both are cheap to close now and expensive to discover in the run phase.

**VERDICT: FAIL 0.84** (Tier L threshold 0.85)
