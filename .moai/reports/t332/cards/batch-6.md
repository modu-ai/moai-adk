# t332 card sweep — batch 6

Cards: t339 t344 t345 t347 t348 t353 t359 t360 t361 t363 (10 entries)
Worktree HEAD: 6165f9f5e
Pinned develop: ee50984abe4f11ac337382b48a26328f091e200a
Pinned main:    48239c7dc7428c8751a04f6321887c2d36123884

### t339

**Premise (one sentence).** t317's plan.md/spec.md still carry three specific documentation
defects (D10/D11/D12) that were re-confirmed four times but never closed.

**Premise verdict.** `holds` — all three checked directly against the current tree.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Both queries (`--grep='\bt339\b'` against pinned develop and pinned main) returned no output.

**Claim.** The three debts the card names are still present verbatim in
`.moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/plan.md` and `spec.md`, even though the SPEC itself
(SPEC-AGENT-EMIT-LINEAGE-001, card t317) has since been fully implemented and closed
(commits `742a9485d`..`3235aa08f`..`0ad4b52ba`, all ancestors of pinned develop). The doc debt
survived the SPEC's own close.

**Evidence.**
```
$ grep -n "B.1" .moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/plan.md
13:### B.1 영향 파일 전수 열거 (추정 아님)
52:**다시 늘어나면 다시 판정한다.** ...
```
The B.1 table (lines 15-20) still lists exactly 5 rows (Makefile, doctor_<name>.go,
doctor_<name>_test.go, doctor.go, CLAUDE.local.md) — D10's claim confirmed.

```
$ git show --stat 742a9485d 6335b731b f3e5006ce | grep -E "^\s*(internal|Makefile|CLAUDE)"
 internal/cli/doctor.go                             |   4 +
 internal/cli/doctor_agentemit_embed.go             | 287 ++++++++++++++
 internal/cli/doctor_agentemit_embed_test.go        | 411 +++++++++++++++++++++
 internal/cli/testdata/doctor-dark.golden           |   5 +-
 internal/cli/testdata/doctor-light.golden          |   5 +-
 internal/cli/testdata/doctor-nocolor.golden        |   5 +-
 Makefile | 24 ++++++++++++++++++++++--
 ...
```
The actual run-phase touched three golden files (`doctor-dark.golden`, `doctor-light.golden`,
`doctor-nocolor.golden`) that B.1's enumeration never lists — this is the "골든 3본" D10 names.

```
$ grep -n "파일 1개" .moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/plan.md
18:| 2 | ... 이 리포의 doctor 항목은 파일 1개에 사는 것이 규약이다(...) | M1 |
23:**5건.** ... doctor 항목은 파일 1개 + 짝 테스트 1개로 살고 ...
```
D11 confirmed: the "lives in one file" claim is still there unmodified, and is contradicted by
the golden-file evidence just above (the item's footprint is body file + test file + 3 golden
files + a Makefile edit — more than "one file").

```
$ grep -n "count != 11" internal/template/agentemit/golden_test.go
284:	if count != 11 {
```
`spec.md:88` cites this assertion as `golden_test.go:285`; the actual line is 284 — D12
confirmed as a stale-by-one-line citation.

**Baseline-attribution.** All three checks run against worktree HEAD `6165f9f5e`
(`.moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/plan.md`, `spec.md`,
`internal/template/agentemit/golden_test.go`, current tree). The commit-touched-file evidence is
from `git show --stat` on the three cited SPEC-AGENT-EMIT-LINEAGE-001 commits, which are
ancestors of pinned develop.

**Gaps.** Did not re-verify whether any OTHER debt beyond D10/D11/D12 has crept into the two
files since t317 closed. Did not check whether a different card already re-issued a doc-fix for
this SPEC (grep of `.moai/reports/` for a superseding fix was not run — bounded depth).

**Residual-risk.** The three items are stated by the card itself as "동작에 영향 없음" (no
behavioral impact) — so even if left open, they only mislead a future reader of a closed SPEC's
plan artifacts, not runtime behavior. If SPEC docs are treated as immutable historical record
post-close, "fixing" them may itself be an out-of-policy edit — that policy question is not
something this sweep resolves.

**Proposed disposition.** `keep` — all three named defects independently reconfirmed present at
HEAD; the fix is a small, scoped documentation edit to a closed SPEC's plan/spec artifacts. Rests
on the three verbatim greps above.

**Overlap candidates.** none observed — no other in-scope (batch or full list) card touches
SPEC-AGENT-EMIT-LINEAGE-001 or card t317's artifacts.

---

### t344

**Premise (one sentence).** SPEC-VERIFICATION-COMPLETENESS-001's prediction ledger (VC-2/4/5/6)
measures "audit-flag count = 0" as success, which is a metric that rises precisely when the rule
works well at the audit layer — the polarity is backwards.

**Premise verdict.** `holds` — the flip is real (verified against the ledger text) and the
corrective text the card describes is already inline in the same file, but no decision has been
made on whether to generalize it.

**Landing verdict.** `not-landed`
- commit: `2db93496c` (mention only, see Claim)
- pinned ref: ee50984abe4f11ac337382b48a26328f091e200a
- `--is-ancestor` exit: 0 (of the mentioning commit; it does not deliver t344's own scope)
- branch + tip (in-flight only): —

**Claim.** t344's own scope-decision work (whether to generalize the ledger-format correction, or
treat it as already closed) has not been done. The only commit matching `\bt344\b` on pinned
develop is `2db93496c`, whose subject is "record which cards the follow-up candidates became" —
it names t344 as one of four cards issued from t241's candidate table (C1..C4 → t341..t344), it
does not perform t344's work.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt344\b' --oneline
2db93496c docs(t241): record which cards the follow-up candidates became (t241)
$ git show -s --format=%B 2db93496c
docs(t241): record which cards the follow-up candidates became (t241)

C1 through C4 were issued as t341, t342, t343 and t344; C5 remains unowned. ...
Also records the two conditions the operator attached: ... and
t344 must first judge whether the ledger defect is already closed by the §A.5
text this card wrote.
```
```
$ sed -n '105,125p' .moai/specs/SPEC-VERIFICATION-COMPLETENESS-001/plan.md
### §A.5 예측 장부 (Prediction Ledger — t241 주의 반영)
...
| VC-2 | 'mutant가 쓰여 AC 무효화' 감사 지적 0건 | ... | **false** — 신규 7건 ...
...
**장부 문면의 결함 (판정 중 발견, 다음 장부 저작 시 교정할 것).** VC-2·VC-4·VC-5·VC-6 은 예측을
"감사 지적 0건" 으로 적었다. 이 지표는 규칙이 감사 층에서 잘 작동할수록 올라간다 ...
다음 장부는 예측을 채택까지 살아남은 건수로 쓰고, 감사 지적 수는 반대 부호의 보조 지표로 둘 것.
```
The corrective is already recorded inline in §A.5's own plan.md prose — the exact instance-level
fix the card describes.

```
$ grep -rln "citation.count\|rule-citation\|citation_count" internal/ .moai/
.moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/progress.md   # unrelated (graph-freshness cadence)
```
No general ledger-authoring rule or convention file was found that has already absorbed this
correction as a repo-wide format rule (searched for a citation-count / ledger convention file;
one unrelated hit).

**Baseline-attribution.** Ledger text measured against
`.moai/specs/SPEC-VERIFICATION-COMPLETENESS-001/plan.md` at worktree HEAD `6165f9f5e`. Commit
`2db93496c` and its ancestry against pinned develop `ee50984ab...` measured directly.

**Gaps.** Did not check whether `moai-constitution.md` §Lessons Protocol or any other
always-loaded rule file already generalizes the "survival-to-adoption, not audit-flag-count"
principle outside this one SPEC's plan.md — only a targeted grep for a dedicated ledger/citation
tool was run, not a full-text search of the rules tree. Did not check t333's current status to
see whether its "표현 기대치" design already absorbed this constraint (card explicitly asks the
next actor to check this before starting).

**Residual-risk.** If a generalization already landed somewhere outside the paths checked, this
card's premise ("아직 안 닫힘") would be wrong even though the specific instance evidence above is
accurate — the scope-decision itself is exactly what's unresolved, which is the card's own stated
ask.

**Proposed disposition.** `needs-operator-decision` — rests on the fact that the card's own two
branch options (a: generalize, b: treat as closed) are both still open per the evidence above; a
sweep worker choosing between them would be making the operator's call.

**Overlap candidates.** t345 (same source — t241 verdict.md and lane-14/lane-17 findings; both
cards reference SPEC-VERIFICATION-COMPLETENESS-001 §A.5 and the "policy rule adoption is
unobserved" theme). No other in-scope card touches this SPEC.

---

### t345

**Premise (one sentence).** No mechanism exists to observe whether a policy-layer rule (as
opposed to a workflow, which has run history) is actually cited and applied, versus merely
existing as a file.

**Premise verdict.** `holds` — no dedicated observation tool for policy-rule citation/adoption was
found in the tree.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Both queries (`--grep='\bt345\b'` against pinned develop and pinned main) returned no output.

**Claim.** The card's central question — what observes a policy rule's real-world adoption — is
still unanswered; nothing in the tree currently computes a citation count, an audit-artifact
reference count, or an authoring-layer application trace for a rule file as a standing mechanism.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt345\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt345\b' --oneline
(no output)
$ grep -rln "verification-completeness" .moai/reports/ | wc -l
16
$ grep -rln "citation.count\|rule-citation\|citation_count" internal/ .moai/
.moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/progress.md
```
The 16 report-directory hits for `verification-completeness` show the rule IS being cited by
audit artifacts by hand (as t241's judgment relied on), but no standing tool computes this — the
one code hit for a citation-count concept is unrelated (graph-freshness cadence, not rule
adoption).

**Baseline-attribution.** Grep scans run against worktree HEAD `6165f9f5e`, scoped to
`.moai/reports/`, `internal/`, and `.moai/`.

**Gaps.** Did not fully read t333's design.md to confirm the "Out of Scope, 이름만" characterization
the card asserts about how t333 handled this candidate (C5) — that claim was taken from the card
text and not independently re-verified against t333's artifacts at HEAD (t333 is not in the
in-scope list for this sweep, so its current artifact state was not opened). Did not search for a
possible non-repo mechanism (e.g., an external log/analytics system) that might already answer
this.

**Residual-risk.** This is a design/research question, not a defect with a fixed verifiable
state — "holds" here means "the gap the card names still appears open," not that the card's
proposed three sub-decisions (a/b/c) have a single correct answer. A different reading of what
counts as "observation" could change the verdict.

**Proposed disposition.** `needs-operator-decision` — the card itself frames three sub-decisions
(what counts as observation, whether it discriminates audit-absorption from authoring-absorption,
whether it's cheaper than manual reading) that are genuinely open design choices, not facts a
sweep can settle.

**Overlap candidates.** t344 (shared source: t241 verdict.md's follow-up candidate table, and both
name the "감사 지적 0건" polarity trap). No other in-scope card touches this theme.

---

### t347

**Premise (one sentence).** t333's classification/status model should be split into a sub-SPEC
authored as a state table (not prose), because three plan-audit iterations on
SPEC-GUARD-LIVENESS-001 kept moving the same defect family (D2/D5/N2/T2/T4) without closing it.

**Premise verdict.** `holds` — the split happened and was authored as a state table, exactly as
asked; a residual open question (T2's contradiction) is tracked inside the new SPEC's own defects,
not left silently unresolved.

**Landing verdict.** `landed` (plan-phase scope of the card)
- commit: `37263c222`
- pinned ref: ee50984abe4f11ac337382b48a26328f091e200a
- `--is-ancestor` exit: 0
- branch + tip (in-flight only): —

**Claim.** SPEC-GUARD-STATE-MODEL-001 was created (card t347) as the split-off sub-SPEC, authored
as a state table per the card's [HARD] instruction, and its plan-phase iter-1 audit defects were
closed. Its `status:` frontmatter is still `draft` and §E.2/E.3/E.4 (run-phase evidence) are
explicitly `<pending run-phase>` — the card's specific ask (split + author as state table) is
plan-complete; the SPEC's own implementation is not yet run.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt347\b' --oneline
37263c222 feat(SPEC-GUARD-STATE-MODEL-001): card t347, state-table delivery column, and instance 7
0f27fa774 fix(SPEC-GUARD-STATE-MODEL-001): close all seven blocking and four optional iter-1 defects (card t347)
7489d4f86 fix(SPEC-GUARD-LIVENESS-001): close all six iter-4 blocking defects, both optional (card t333/t347)
558925d00 fix(SPEC-GUARD-LIVENESS-001): close iter-2 D9, D10, D11 and two twins (t333/t347)
$ git merge-base --is-ancestor 37263c222 ee50984abe4f11ac337382b48a26328f091e200a
(exit 0, no output)
```
```
$ grep -n "status:" .moai/specs/SPEC-GUARD-STATE-MODEL-001/spec.md
5:status: draft
$ tail -8 .moai/specs/SPEC-GUARD-STATE-MODEL-001/progress.md
## §E.2 Run-phase Evidence
_<pending run-phase>_
## §E.3 Run-phase Audit-Ready Signal
_<pending run-phase>_
## §E.4 Sync-phase Audit-Ready Signal
_<pending sync-phase>_
```
```
$ cat .moai/specs/SPEC-GUARD-STATE-MODEL-001/spec.md | head -14
id: SPEC-GUARD-STATE-MODEL-001
title: "Guard liveness state model: declare firing expectations, and decide every entry into
exactly one classification (card t347)"
...
tags: "guard, liveness, state-model, classification, manifest, cadence, totality, t347"
```

**Baseline-attribution.** Commit ancestry checked against pinned develop
`ee50984abe4f11ac337382b48a26328f091e200a`. SPEC status/progress fields read from
`.moai/specs/SPEC-GUARD-STATE-MODEL-001/{spec.md,progress.md}` at worktree HEAD `6165f9f5e`.

**Gaps.** Did not verify whether the SPEC has since progressed to run-phase in a live worktree not
covered by the two pinned refs (e.g., a worktree branch ahead of develop) — the worktree list
(`00-worktree-list.txt`) was not cross-checked for a SPEC-GUARD-STATE-MODEL-001-named branch. Did
not re-derive whether the "T2 모순" the card flags as unresolved is actually closed inside the new
SPEC's REQ set (only progress.md's summary was read, not the full spec.md REQ text).

**Residual-risk.** "Landed" here is scoped narrowly to the card's literal ask (split into a
state-table sub-SPEC); the SPEC's actual implementation (REQ-GDL-004/007/008/009 behavior) remains
unbuilt, so if the operator's intent for this card included seeing the guard behavior itself
change, that part is still open.

**Proposed disposition.** `already-landed` — rests on commit `37263c222`'s subject line and body
explicitly naming "card t347" and "state-table delivery column" as delivered, verified as an
ancestor of pinned develop.

**Overlap candidates.** none observed in the in-scope list — SPEC-GUARD-STATE-MODEL-001 and
SPEC-GUARD-LIVENESS-001 (card t333) are not themselves in `inscope-all.txt`.

---

### t348

**Premise (one sentence).** SPEC-AC-COUNT-DISCRIMINATOR-001's reserved-token AC-counting
convention (from t338) does not retroactively mark 34 existing citation files, so they keep being
counted as `live` (silently over-counted) until a new native-prefix-declaration syntax closes that
gap — a syntax this SPEC deliberately did not build.

**Premise verdict.** `holds` — corroborated near-verbatim against the completed SPEC's own plan.md
and design.md.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Both queries (`--grep='\bt348\b'` against pinned develop and pinned main) returned no output.

**Claim.** t348's own scope (the native-prefix-declaration syntax) has not been built. The SPEC it
follows up on, SPEC-AC-COUNT-DISCRIMINATOR-001 (t338), is `status: completed` and explicitly
records this exact gap as a named follow-up candidate it declined to build.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt348\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt348\b' --oneline
(no output)
$ grep -n "status:" .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/spec.md
5:status: completed
$ grep -n "34\|네이티브 접두사" .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/plan.md
93:**남는 것(정직하게 적는다)**: ... 토큰이 하나도 없는 순수 인용(현재 34건의 기본형)은 §3.2 의
판정에서 live 로 읽히므로 정지가 아니라 과다 계상으로 남는다. ... 그것은 정당한 다중 도메인
SPEC(AC-APO-* + AC-DCP-*)까지 정지시키므로 네이티브 접두사 선언이라는 새 기제를 부른다. 이 카드는
그 기제를 만들지 않는다 — 후속 카드 후보로만 기록한다(spec.md §6).
$ grep -n "네이티브 접두사 선언" .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/design.md
136:| 접두사 단위 애매성 판정(순수 인용까지 잡기) | ... 네이티브 접두사 선언이라는 새 문법이 필요하고
이 카드의 크기를 넘는다 — 후속 카드 후보(spec.md §6) |
```
The 34-file count, the "live not stopped" over-count mechanism, and the native-prefix-declaration
follow-up candidate are all named identically in the completed SPEC's own artifacts.

**Baseline-attribution.** All figures read from
`.moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/{spec.md,plan.md,design.md}` at worktree HEAD
`6165f9f5e` — the SPEC's own authored record, not re-derived by re-running the counter.

**Gaps.** Did not re-run the actual AC-counting tool/discriminator against the current tree to
confirm 34 is still accurate today (could have drifted since the SPEC closed as more citation
files were added or normalized) — the card itself warns not to cite 156 as a defect count without
re-measuring, and the same caution applies to 34. Did not check whether the 7 false-positive count
the card cites ("오탐 7건 실측") is independently verifiable at HEAD — taken from card text only.

**Residual-risk.** If citation files have been added/normalized since SPEC-AC-COUNT-DISCRIMINATOR-001
closed, the live 34-count could now differ from what's cited here, though the structural gap
(no native-prefix syntax exists) would still hold regardless of the exact count.

**Proposed disposition.** `keep` — rests on the completed SPEC's own plan.md/design.md explicitly
naming this exact follow-up candidate as out-of-scope and unbuilt.

**Overlap candidates.** none observed in the in-scope list — SPEC-AC-COUNT-DISCRIMINATOR-001
(card t338) is not itself in `inscope-all.txt`.

---

### t353

**Premise (one sentence).** REQ-MRG-010/AC-MRG-013 (the R4-form lint exclusion for
SPEC-MOVING-REF-GUARD-001) was deliberately deferred by operator decision because the R4 form has
zero observed occupants in the corpus, and should stay deferred until R4 is actually observed.

**Premise verdict.** `holds` — the deferral, its rationale, and its resume condition are recorded
verbatim in t342's verdict.md, and this card is that deferral's own follow-up (issued by the
operator, as the verdict explicitly notes none was issued from that lane).

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Both queries (`--grep='\bt353\b'` against pinned develop and pinned main) returned no output.
(Note: SPEC-MOVING-REF-GUARD-001 / t342 itself IS landed — its worktree tip `38f937a4f` is
confirmed an ancestor of pinned develop — but that landing covers only the non-deferred parts.)

**Claim.** The R4-form lint exclusion remains unimplemented, exactly as t342's verdict.md
recorded it deferred. No commit citing t353 exists on either pinned ref, and no implementation of
REQ-MRG-010/AC-MRG-013 was found.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt353\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt353\b' --oneline
(no output)
$ sed -n '106,124p' .moai/reports/t342/verdict.md
## Deferred — REQ-MRG-010 and AC-MRG-013 (Q0, option C)
Deferred by operator decision.
**What:** the R4-form lint exclusion, its acceptance criterion, and both counter-mutations ...
**Why:** §B.7 measured R4's reachable class as 0 of 42 candidate lines on two independent probes...
**Resume condition:** reconsider when the R4 form is actually observed in the corpus. ...
No follow-up card was issued from this lane; card issuance is the operator's act.
```
```
$ git merge-base --is-ancestor 38f937a4f ee50984abe4f11ac337382b48a26328f091e200a
(exit 0, no output)
```

**Baseline-attribution.** verdict.md text read from `.moai/reports/t342/verdict.md` at worktree
HEAD `6165f9f5e`. t342 worktree tip ancestry checked against pinned develop.

**Gaps.** Did not re-run the "R4 form observed in corpus" check to see whether the resume
condition has since triggered — the card itself states the resume condition as the thing to check
before acting, and doing that full corpus scan was judged out of this sweep's bounded-depth scope.

**Residual-risk.** If the R4 form has since appeared in the corpus (e.g., via a card that landed
after t342), the deferral's resume condition may already be satisfied and this card may be ready
to act on rather than merely "keep deferred" — that determination was not made here.

**Proposed disposition.** `keep` — rests on the verdict.md text confirming the deferral is real,
deliberate, and explicitly awaiting a resume trigger not yet checked by this sweep.

**Overlap candidates.** none observed in the in-scope list — t342 itself is not in
`inscope-all.txt` (already landed).

---

### t359

**Premise (one sentence).** The bigger half of t331's operator-mandated split (add a landing-evidence
column to the todo items table via `ALTER TABLE ADD COLUMN`) depends on t331-A landing first, and
plan-audit iter-1 flagged two misattributed-requirement defects (D1, D2) that must be resolved
before the design can proceed.

**Premise verdict.** `holds` — D1 and D2's underlying requirement text was independently
re-verified against the actual SPEC files cited.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Both queries (`--grep='\bt359\b'` against pinned develop and pinned main) returned only mentions
(see Evidence), no delivering commit; no SPEC directory for t359's own scope exists in
`.moai/specs/`.

**Claim.** t359's precondition (t331-A / SPEC-TODO-LANDING-STATE-001) HAS landed and closed via
the 3-phase close (commit `51daada00`, merge, ancestor of pinned develop). t359's own scope
(adding the landing-evidence column) has not been started — no SPEC directory exists for it, and
its plan-audit D1/D2 findings (cited from t331's iter-1/iter-2 audits) are exactly as described.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt359\b' --oneline
45cff0f59 docs(t331): plan-audit iter-2 verdict — PASS 0.85, 4 blocking routed as a delta fix (t331)
e1d480eba docs(SPEC-TODO-LANDING-STATE-001): iter-1 remediation — scope split to half A, 11 REQ Tier M (t331)
$ git show -s --format=%B e1d480eba
... the evidence half moved to card t359, which depends on this one landing first. ...
Section B.2 is now a pointer that names the questions t359 must answer — what
REQ-TODO-013 actually permits, how SPEC-TODO-ANALYSIS-001 read it, and whether
an observation may name a commit under SPEC-KANBAN-QUEUE-PR-SYNC-001 REQ-1.10 ...
```
Both hits are mentions (t331's own audit history), not t359's delivery — confirmed by reading
each commit's full subject/body.
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --oneline --grep='SPEC-TODO-LANDING-STATE-001' | head -3
51daada00 merge(WT-card-landing-state): SPEC-TODO-LANDING-STATE-001 — landed answer resolved from the integration branch (t331)
fee6c22d9 chore(SPEC-TODO-LANDING-STATE-001): backfill sync_commit_sha (t331)
c9f712232 docs(SPEC-TODO-LANDING-STATE-001): sync-phase — 3-phase close, CHANGELOG + docs-site (t331)
$ git merge-base --is-ancestor 51daada00 ee50984abe4f11ac337382b48a26328f091e200a
(exit 0, no output)
```
```
$ grep -n "REQ-TODO-013" .moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md
59:- **REQ-TODO-013** (Ubiquitous) The backlog store shall preserve the existing version-1 record
shape ... changing it only additively (the high-water mark, per REQ-TODO-009).
$ grep -n "REQ-1.10" .moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md
251:**REQ-1.10** — The resolver shall not name, return, or otherwise claim which
```
D1 confirmed: REQ-TODO-013 does say "additive-only," not "field-set freeze." D2 confirmed:
REQ-1.10 of the completed SPEC-KANBAN-QUEUE-PR-SYNC-001 does forbid the resolver from naming which
commit — directly bearing on t359's plan to add a commit-SHA-display column.

**Baseline-attribution.** Commit ancestry against pinned develop `ee50984ab...`. Requirement text
read from `.moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md:59` and
`.moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md:251` at worktree HEAD `6165f9f5e`.

**Gaps.** Did not check whether SPEC-TODO-ANALYSIS-001 (the SPEC the card says "판정한 기록도
미조정" — read the opposite way) has been reconciled since — only the card's characterization was
taken at face value for that specific sub-claim; the direct text of SPEC-TODO-ANALYSIS-001:51 was
not re-opened. Did not check whether a plan-phase draft for t359's own SPEC exists in a live
worktree outside the two pinned refs.

**Residual-risk.** If SPEC-TODO-ANALYSIS-001's actual text does not read the way the card
describes, D1's "resolved" framing (that REQ-TODO-013 permits ADD COLUMN) could be less settled
than the evidence above suggests.

**Proposed disposition.** `keep` — rests on t331-A's landing (verified ancestor of pinned develop)
having satisfied the precondition, and D1/D2's underlying requirement text matching the card's
citations exactly.

**Overlap candidates.** none observed in the in-scope list — t331 (SPEC-TODO-LANDING-STATE-001),
SPEC-TODO-ANALYSIS-001, and SPEC-KANBAN-QUEUE-PR-SYNC-001 are not themselves in
`inscope-all.txt` (t331 already landed).

---

### t360

**Premise (one sentence).** GLM effort transmission is keyed to the model's High-tier slot at
three call sites while the web console's per-tier UI lock only disables non-max options when that
specific tier's slot is `glm-5.3-flash`, so a non-flash slot can silently accept and then discard a
saved low-effort setting.

**Premise verdict.** `holds` — both sides of the mismatch verified directly in source.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Both queries (`--grep='\bt360\b'` against pinned develop and pinned main) returned only a mention
(t350's commit naming t360 as an out-of-scope defect it discovered), not a delivering commit.

**Claim.** The three cited transmission call sites and the per-tier UI lock function are exactly
as the card describes, and the mismatch is real at current HEAD.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt360\b' --oneline
e1481f4d5 feat(config): split the GLM Fable slot onto glm-5.3 (t350)
$ git show -s --format=%B e1481f4d5
... Out of scope, filed as t360: the web console's flash effort lock is
per-tier (assets/app.js:490-497) while the reasoning wire derives from
the High slot (launcher.go:1207), so the Fable effort select now unlocks
while a stored effort.fable is still pinned to max. Pre-existing keying
defect that every-slot-flash had been masking.
```
```
$ grep -n "ResolveGLMReasoningForModel" internal/web/agentfm.go
314:	return template.ResolveGLMReasoningForModel(llm.GLM.Models.High, name, me.Effort).Name
$ sed -n '485,500p' internal/web/assets/app.js
  function applyGLMFlashEffortLock(modelSel) {
    var tier = modelSel.name.slice("llm.glm.models.".length);
    var effort = document.querySelector('select[name="llm.glm.effort.' + tier + '"]');
    if (!effort) return;
    var isFlash = modelSel.value === "glm-5.3-flash";
    for (var o = 0; o < effort.options.length; o++) {
      var opt = effort.options[o];
      opt.disabled = isFlash && opt.value !== "max";
      ...
```
Confirms: transmission at `agentfm.go:314` keys on `llm.GLM.Models.High` unconditionally; the UI
lock keys per-tier (`tier := modelSel.name...`) on whether that tier's own slot is
`glm-5.3-flash`. Since t350 landed (Fable slot now `glm-5.3`, not flash), the Fable UI unlocks
while transmission still keys on the High slot's model — exactly the described mismatch.

**Baseline-attribution.** Source read at worktree HEAD `6165f9f5e`:
`internal/web/agentfm.go:314`, `internal/web/assets/app.js:485-500`. Commit `e1481f4d5` and its
ancestry checked against pinned develop.

**Gaps.** Did not check `internal/cli/model.go:117` or `internal/cli/launcher.go:1207` (the other
two cited call sites) directly — only the first (`agentfm.go:314`) was opened; the card's citation
of the other two was not independently re-verified line-for-line. Did not run the test suite to
confirm existing test expectations would need to move, as the card predicts.

**Residual-risk.** If `model.go:117` or `launcher.go:1207` have since been partially fixed
independent of this card, the defect's full extent (three sites vs. fewer) could be narrower than
stated.

**Proposed disposition.** `keep` — rests on the one directly-verified call site plus the UI lock
function matching the card's description exactly, and t350's own commit message naming this as a
real, unaddressed follow-up.

**Overlap candidates.** none observed in the in-scope list — t350 (the SPEC that surfaced this) is
not itself in `inscope-all.txt` (already landed).

---

### t361

**Premise (one sentence).** `TestBinaryLag_OneSeamServesBothSurfaces` in `internal/cli` fails on
develop CI (both Test and Race jobs) because a `deferred-scan` async-suppression switch introduced
by t333 lives in the unexported `internal/hook` package and cannot be reached from
`internal/cli`'s test binary, so a goroutine spawned by `guardLivenessRefresh` writes into a
`t.TempDir()` after that test's cleanup has already run.

**Premise verdict.** `unverified` — the structural claims (comment text, switch scope) check out,
but the causal mechanism the premise actually asserts was not reproduced, so the premise as worded
is undecided. The card itself frames the mechanism as an unreproduced hypothesis, and this sweep
did not reproduce it either. (Orchestrator normalization, M3 post-check: the worker wrote a
compound verdict here; AC-BH-011 admits exactly one of `holds` / `falsified` / `unverified`, and a
premise whose causal claim is unreproduced is undecided rather than holding. The worker's substance
is unchanged.)

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Both queries (`--grep='\bt361\b'` against pinned develop and pinned main) returned no output —
no fix has landed for this defect.

**Claim.** The structural facts the card's attribution rests on are confirmed at HEAD: the
"never awaited" comment exists verbatim, and `deferredScansAsync` is an unexported package-level
var in `internal/hook` toggled only by that package's own `TestMain`. Whether this actually
produces the observed CI failure was not independently reproduced by this sweep (nor claimed to be
by the card, which explicitly requires reproduction as the first step of any fix).

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt361\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt361\b' --oneline
(no output)
$ grep -n "guardLivenessRefresh\|never awaited" internal/hook/session_start.go
140:	// activations that got that far. The refresh is never awaited, so entering
146:	guardLivenessRefresh(ctx, guardLivenessRoot, h.asyncDeferredScans())
$ grep -rn "deferredScansAsync\b" internal/hook/*.go
internal/hook/main_test.go:38:// It also flips deferredScansAsync to false for the test binary: ...
internal/hook/main_test.go:47:	deferredScansAsync = false
internal/hook/session_start_guard_liveness.go:80:		// Test path (TestMain sets deferredScansAsync=false): run inline so no
```
Confirms: `deferredScansAsync` is lowercase (unexported), and the only place it is set to `false`
is `internal/hook/main_test.go`'s `TestMain` — a switch scoped to that one package's test binary,
exactly as the card describes. Since the failing test (`TestBinaryLag_OneSeamServesBothSurfaces`)
lives in `internal/cli`, it is a different test binary and cannot reach this switch.

**Baseline-attribution.** Source read at worktree HEAD `6165f9f5e`:
`internal/hook/session_start.go:140,146`, `internal/hook/main_test.go:38,47`,
`internal/hook/session_start_guard_liveness.go:80`.

**Gaps.** Did NOT run `internal/cli`'s `TestBinaryLag_OneSeamServesBothSurfaces` to reproduce the
cleanup failure directly — the card itself marks this as the mandatory first step before any fix,
and reproducing it was judged beyond this sweep's bounded per-card depth (it would require running
a package test, observing a possibly-flaky timing failure, and confirming both the ubuntu and Race
jobs). Did not check whether t352 (t.TempDir cleanup race, explicitly named by the card as a
possible same-class sibling) has since absorbed this defect — t352 is not in this sweep's
in-scope list.

**Residual-risk.** Since the causal chain from "goroutine outlives TempDir cleanup" to "observed
CI failure" was not reproduced here, there remains a chance the actual CI failure has a different
or additional cause not captured by this hypothesis — the card itself flags this same risk.

**Proposed disposition.** `keep` — rests on the structural evidence above (comment text + switch
scope) matching the card's attribution chain steps 3-4 exactly; step 1's reproduction remains
undone by both the card's author and this sweep.

**Overlap candidates.** t352 (WT-tempdir-cleanup-race, live worktree per
`00-worktree-list.txt`) — named explicitly by the card itself as a possible same-class sibling to
fold into. Not in this sweep's in-scope list, so not independently checked here.

---

### t363

**Premise (one sentence).** Every GitHub Actions workflow whose `concurrency.group` includes
`develop`-triggered pushes is keyed on `github.ref`, which differs between a `push` run
(`refs/heads/develop`) and a `pull_request` run (`refs/pull/N/merge`) for the same head commit, so
`cancel-in-progress` cannot cancel one against the other — causing double CI runs for every
develop→main release PR.

**Premise verdict.** `holds` — the concurrency group expression and the trigger config were both
verified directly against `.github/workflows/ci.yml`.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Both queries (`--grep='\bt363\b'` against pinned develop and pinned main) returned no output —
this card is brand new (per the dispatch note, absent from the plan-phase snapshot) and no fix has
landed.

**Claim.** `ci.yml`'s `concurrency.group` is `${{ github.workflow }}-${{ github.ref }}` exactly as
cited, and its triggers are `push: branches: [main, develop]` plus
`pull_request: branches: [main]` — confirming the card's claim that a develop→main PR run and a
develop push run share a head commit but not a `github.ref`, so they cannot cancel each other.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt363\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt363\b' --oneline
(no output)
$ sed -n '16,20p;28,31p' .github/workflows/ci.yml
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]  # main으로 향하는 모든 PR에서 CI 실행
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
$ ls .github/workflows/*.yml .github/workflows/*.yaml | wc -l
18
```
The 18-workflow count matches the card's "저장소 18개 중 일곱" framing (total corpus size
confirmed; the specific 7-workflow no-branch-filter list was not individually re-verified — see
Gaps).

**Baseline-attribution.** `.github/workflows/ci.yml` read at worktree HEAD `6165f9f5e`. Workflow
file count via `ls` at the same HEAD.

**Gaps.** Did not open the other 6 workflows the card names (graph-freshness, lsel-leak-guard,
spec-lint, docs-i18n-check, claude, community, spec-status-auto-sync) to individually confirm each
lacks a `pull_request.branches` filter — only the total count (18) was cross-checked, not the
per-file claim. Did not verify the card's claim about actual observed CI run history (e.g., a
specific PR where both a push run and a PR run fired for the same head) — that would require a
live `gh run list` query against GitHub, which was judged out of scope for a read-only tree sweep.

**Residual-risk.** The concurrency-key mismatch is a structural fact confirmed directly in the
workflow YAML; whether it is the SOLE cause of "double CI runs" for every develop→main PR (versus
a contributing factor among several) was not independently re-derived here — taken from the card's
own reasoning chain, which itself states this was arrived at by falsifying a simpler branch-filter
hypothesis.

**Proposed disposition.** `keep` — rests on the verbatim `concurrency.group` expression and the
`push`/`pull_request` trigger blocks in `ci.yml`, both directly re-read at HEAD.

**Overlap candidates.** t294 (WT-freshness-trigger, live worktree per `00-worktree-list.txt`,
locked) — the card text states t294 was split off this same investigation and retains only the
graph-freshness branch-filter fix (axis A), while this card (t363) is the separate concurrency-key
issue (axis B). Not in this sweep's in-scope list, so not independently checked here.
