# Progress — SPEC-HUMANIZE-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-07-09
- plan_revision: v0.2.0 (iter-2 plan-audit fix — prior verdict FAIL 0.78 < 0.85; revised in place, not re-authored)
- plan_revision_iter3: iter-2 re-audit FAIL 0.84 (missed 0.85 by 0.01) — a self-introduced plan↔spec inconsistency (D2/D5 scope expansion applied to spec/acceptance/design/progress but NOT plan.md). iter-3 fix: plan.md §B6 (out-of-scope → in-scope NOTICE contract), §D constraints (name deliberate frontmatter changes), §F M5 (REQ-HUM-015/016 given a milestone owner + AC refs), §G anti-patterns (removed "create NOTICE.md" prohibition, replaced with the dangling-pointer inverse); acceptance.md AC-006b negative grep alternation gained `verbless` (research.md §2.3 synonym; prevents false PASS). Now every REQ 001-016 has a milestone owner.
- plan_revision_iter4: iter-3 re-audit PASS-WITH-DEBT 0.87. Two debt items closed: (D1 MAJOR) the `verbless` token added in iter-3 created a false-FAIL — ENC-7's own definition "bland verbless: Get Started, Submit" (research.md §2) would trip the AC-006b negative grep, rejecting a correctly-authored module (same class as the iter-1 date collision). Fix: scoped the fragment-family negative so a token must be ADJACENT to a headline-shaped noun (headline|title|slide|fragment), added an AC-006a-style authoring-contract note recording the ENC-7 collision; positive grep untouched. Verified: benign ENC-7 row → NEG=0, genuinely-bad removable-verbless-headline row → NEG=1, ENC-3 "parallel fragments" → NEG=0. (D2 MINOR) renamed spec.md §F self-contradicting heading `### Out of Scope — … is now IN scope` → `### Reversal note (v0.2.0) — NOTICE.md creation moved IN scope`.
- plan_revision_iter5: v0.3.0 — USER DECISION at Implementation Kickoff gate: Korean module RE-AUTHORED as original work from the maintainer's own taxonomy; MIT dependency dissolved. Verified fact (grep, this session): the reference source (`general-humanize-korean` at claude.mo.ai.kr) carries ZERO `im-not-ai|epoko` references anywhere in the source skill dir — the MIT encumbrance was an inaccurate v1.0.0 self-description in adk's korean.md:133, not a content lineage. Revisions: REQ-HUM-001 rewritten (full korean.md rewrite, prose A–J + copy; no port claim; AC-HUM-001 gained a no-port-claim negative); REQ-HUM-015 rewritten in place (attribution cleanup: 5 `See NOTICE.md` pointers removed, courtesy credit "structure inspired by the im-not-ai (Humanize KR) project" with no license claim; AC-HUM-016 = `grep -rn NOTICE.md` → 0); REQ-HUM-016 simplified in place (`license: Apache-2.0` unchanged, no compound; AC-HUM-017 = `grep -rn 'MIT License'` → 0 + license-field verbatim check); REQ-HUM-011 re-amended (license moved from deliberate-change to PRESERVE; attribution blocks moved from preserve to REWRITE); spec.md §F second reversal recorded (NOTICE.md: out → in → MOOT) + Korean prose-layer exclusion carved out; plan.md header/§A/§B6/§D/M1/M5/§G aligned; design.md §G rewritten as the resolved decision record (v0.2.0 mixed-license design superseded — premise falsified). REQ IDs stable (015/016 rewritten in place, not deleted). research.md untouched (zero NOTICE/MIT refs — checked).
- tier: L
- artifacts: spec.md, plan.md, acceptance.md, research.md, design.md, progress.md (6 plan artifacts; NO NOTICE.md — dissolved per v0.3.0)
- req_count: 16 (REQ-HUM-001 … REQ-HUM-016; 015 = attribution cleanup, 016 = license unchanged)
- ac_count: 17 AC IDs / 19 checks (incl. AC-HUM-006a/b/c; 016 = 0 NOTICE.md refs + courtesy credit, 017 = 0 MIT-license tokens + Apache-2.0 verbatim). Method mix: 13 mechanical, 2 hybrid, 4 manual (declared in matrix)
- non_transfer_constraints: 3 (KR M-2→EN, KR M-2→JA, ZH 对偶/排比 count→content-first)
- run_phase_nature: documentation authoring (full korean.md re-authoring + 3 module copy-layer appends + SKILL.md shared/attribution edits + catalog.yaml + make build + local sync); NO Go code; NO NOTICE.md
- plan_audit_fixes: D1 (AC-HUM-008 reachable-to-0 rewrite: date body-scoped + commit-SHA class added; scoped to humanize dir, not whole-repo green); D2 (MIT NOTICE.md now in scope); D3 (mechanical positive+negative for AC-006a/b/c); D4 (Guard 2 concrete JA sample); D5 (license reconciliation + REQ-011 amended); D7 (ENC-7 evidence footnote + §6 two-tier header)
- notes: SPEC ID pre-write self-check PASS; ID unique; byte-identity baseline IDENTICAL; catalog entry version 1.0.0 → target 1.1.0; real leak-test TestTemplateNoInternalContentLeak currently RED for pre-existing unrelated agent-common-protocol.md leak (out of scope)

## §E.2 Run-phase Evidence

Run executed in the dedicated agent worktree (branch `worktree-agent-a205e7a01ec2e0f27`, base `92d955daa`), commits local-only per delegation (push owned by orchestrator). All mechanical AC commands below were run verbatim from `acceptance.md` §D.1 against the templates tree at the final run-phase commit.

### AC PASS/FAIL Matrix (19 checks / 17 AC IDs)

| AC | Method | Status | Verification Command (verbatim from acceptance.md) | Actual Output |
|----|--------|--------|-----------------------------------------------------|---------------|
| AC-HUM-001 | mechanical | PASS | §D.1 AC-HUM-001 block (category greps + no-port-claim negative) on korean.md | `CATEGORIES PASS` + `NO PORT CLAIM PASS` |
| AC-HUM-002 | mechanical | PASS | grep ENC-1 + ENC-9 english.md | `PASS` |
| AC-HUM-003 | mechanical | PASS | grep JA-10 + JA-14 japanese.md | `PASS` |
| AC-HUM-004 | mechanical | PASS | grep CN-L + CN-Q chinese.md | `PASS` |
| AC-HUM-005 | hybrid | PASS | grep fact.anchor + copy mode + prose SKILL.md; manual: two grading tables | `PASS`; manual confirm: "### Prose-Mode Grade Table" + "### Copy-Mode Grade Table" both present in the shared section |
| AC-HUM-006a | mechanical | PASS | §D.1 AC-HUM-006a block (JA-10 row S2 + frequency framing, row not S1) | `PASS` (POS=1, NEG=0) |
| AC-HUM-006b | mechanical | PASS | §D.1 AC-HUM-006b block (NEGPAT on ENC table rows + false-positive positive) | `PASS` (NEG=0, POS=1; ENC-7 "bland verbless: Get Started, Submit" phrasing kept intact per the authoring contract) |
| AC-HUM-006c | hybrid | PASS | §D.1 AC-HUM-006c positive grep | `POS PASS`; manual negative confirm: the 对偶/排比 boundary subsection makes content-first vs template-first decisive, with "count is a weak signal on copy" demoting count explicitly |
| AC-HUM-007 | mechanical | PASS | `diff -rq` templates vs local `.claude/` copy after make build + sync | `PASS IDENTICAL` |
| AC-HUM-008 | mechanical | PASS | §D.1 6-class neutrality block (date body-scoped) + advisory leak-test tie | `NEUTRALITY PASS`; advisory: `humanize dir clean (unrelated pre-existing failures ignored)` |
| AC-HUM-009 | manual | PASS | Inspection of all four modules' copy layers | Instruction prose is English in all four; before/after examples are Hangul (korean.md), kana/kanji (japanese.md), Simplified Han (chinese.md), English (english.md) |
| AC-HUM-010 | mechanical | PASS | §D.1 preservation + version-bump greps on SKILL.md | `PRESERVE PASS` + `VERSION BUMP PASS` (im-not-ai token now matched by the courtesy credit) |
| AC-HUM-011 | manual | PASS | Advisory ID listing + reviewer cross-check against research.md §6 / §5 | ID grep returned exactly ENC-1..9, JA-10..14, CN-L..Q (20 IDs); each traces to the research.md §2/§3/§4 catalogue tables (§6 verified sources; ENC-7 backstopped by fetched ENC-1 per §2 [†]); no §5 quarantined hypothesis promoted — the one borderline (childhood-origin beat in a JA-14 example) was removed in commit 6c2fa8aa6 |
| AC-HUM-012 | mechanical | PASS | catalog version grep + hash recompute freshness | `VERSION PASS`; re-running `gen-catalog-hashes.go --all` after commit → `HASH FRESH (no dirty diff)` (skill hash covers root SKILL.md by generator design) |
| AC-HUM-013 | manual (scenario) | PASS | 8-sample matrix per §D.2 (results below) | 8/8 graded, 0 meaning drift |
| AC-HUM-014 | manual (scenario) | PASS | 3 false-positive guards per §D.3 (results below) | 3/3 NOT flagged |
| AC-HUM-015 | mechanical | PASS | grep conservativ + 30%/50%/threshold SKILL.md | `PASS` (conservative-judgment-near-thresholds instruction + no-quantitative-layer limitation note in Over-Editing Guardrails) |
| AC-HUM-016 | mechanical | PASS | §D.1 NOTICE.md + courtesy-credit block | `NOTICE REF PASS (0 matches)` + `NOTICE FILE PASS (absent)` + `CREDIT PASS` |
| AC-HUM-017 | mechanical | PASS | §D.1 MIT-token + license-field block | `MIT TOKEN PASS (0 matches)` + `LICENSE UNCHANGED PASS` |

### 8-Sample Verification Matrix (AC-HUM-013 — detect → rewrite → grade)

| # | Sample | Tells detected | Rewrite outcome | Grade | Meaning drift |
|---|--------|----------------|-----------------|-------|----------------|
| 1 | KR prose (`~을 통해` + `~할 수 있을 것으로 보인다`) | A-2 (S1), G-1 (S2) | `고객 설문을 분석해 개선점을 찾는다` — calque and hedge removed, claim intact | A (prose: 0 S1, 0 S2 residual) | none |
| 2 | KR copy `자동화는 24시간 굴러갑니다 — 복붙에서 위임으로` | A-20 (S1), M-1 (S1), M-3 (S1) | `복붙하던 일을 맡기면, 자동화가 24시간 알아서 돌아갑니다` — fact anchor `24시간` intact, promise preserved | A (copy: 0 S1, 0 anchor loss) | none |
| 3 | EN prose (delve + tapestry + negative parallelism) | EN-A (S1 singletons ×2), EN-B (S1) | `This report examines how team workflows interact. The change saves time and reshapes how the team works.` — change rate estimated >30%, WARN discharged because every edit maps to a detected tell | A (0 S1, 0 S2 residual) | none |
| 4 | EN copy hero (`In today's fast-paced digital world…` + `not just a tool — it's a movement` + `Fast. Simple. Scalable.`) | ENC-4 (S1), ENC-1 whole-phrase (S1-behaving), ENC-2 (S1), ENC-3 (S2) | S1s resolved first, then the tricolon: `One tool that replaces your busywork. Set up in minutes, and it scales with your team.` — no fact anchors present, none lost | A (copy) | none |
| 5 | JA prose (`〜することができます` + `これにより`) | JA-01 (S2), JA-03 (S1) | `このツールが使えます。そのぶん、生産性が上がります。` — no invented specifics | A (0 S1, 0 S2 residual) | none |
| 6 | JA copy (3 consecutive 体言止め + `見出し：` colon) | JA-10 fires on the ≥3-consecutive gate; JA-11 (S1) | endings varied (`…減り、…下がります。導入はかんたん。`), colon heading restructured natively; offer preserved | A (copy) | none |
| 7 | ZH prose (首先…其次…综上所述 + 赋能/闭环) | CN-A (S1), CN-C (S1) | `先把目标定清楚，再帮团队把整个流程跑通——说到底，关键在执行。` — every edit tell-anchored (the —— is legitimate 解释说明) | A (0 S1 residual) | none |
| 8 | ZH copy (`这不仅仅是一双跑鞋，而是…承诺` + `让我们携手共创美好未来`) | CN-L (S2, headline slot), CN-O (S1) | `这双跑鞋，陪你把晨跑坚持下去。` + elevation cut, concrete close; no spec numbers present, none invented or lost | A (copy) | none |

### False-Positive Guard Results (AC-HUM-014 — must NOT flag)

| Guard | Sample | Outcome |
|-------|--------|---------|
| 1 — EN human slide fragment | "Q1 Revenue" / "Our Approach" / "Why It Matters" | NOT flagged — English copy layer has no removable fragment category; terse headlines named a high-false-positive natural register; none is an ENC-1 buzzword noun phrase |
| 2 — JA one strategic 体言止め | §D.3 concrete block (endings か / ます / ません / noun 近道) | NOT flagged — JA-10 frequency gate not tripped (1 non-consecutive 体言止め amid varied endings); `40件が10分で終わります` is a concrete claim, so JA-13 does not fire; no colon/dash artifacts |
| 3 — ZH one crafted 排比 | 万科 「感谢冰峰，感谢风暴，感谢悬崖，感谢缺氧。」 | NOT flagged — content-first test passes (each member a distinct concrete referent, information concentrated, 不落窠臼); count is a weak signal and never decides alone |

### Invariants

| Invariant | Status | Evidence |
|-----------|--------|----------|
| Template-First order (templates → make build → local sync) | PASS | Edit sequence in commits d5df06b5f..89fbac7c6; local mirror synced only after `make build` (exit=0) |
| EN/JA/ZH prose layers additive-only | PASS | `git diff 92d955daa..HEAD` removed-line count per module = 2 (the old `## Source & License` block only); prose category sections byte-untouched |
| SKILL.md frontmatter preservation (user-invocable / allowed-tools / license) | PASS | AC-HUM-010 + AC-HUM-017 outputs above; only `metadata.version` → "1.1.0", `metadata.updated` → run date, `metadata.tags` +copy changed |
| No Go source modified | PASS | `git diff --stat 92d955daa..HEAD` — 17 files: 5 template skill md + 5 mirror md + catalog.yaml + 6 SPEC artifacts; zero `.go` files |
| No new test failures (NEW vs baseline) | PASS | `go test ./internal/template/` → 5 FAILs (TestTemplateNoInternalContentLeak, TestRuleProvenanceAudit, TestOutputStylesTemplateLiveParity, TestAllAgentsInCatalog, TestRuleTemplateMirrorDrift); identical 5 FAILs reproduced at base HEAD 92d955daa in a detached verification worktree — all pre-existing, none reference humanize files |
| Cross-platform build | PASS | `go build ./...` exit=0; `GOOS=windows GOARCH=amd64 go build ./...` exit=0 |

### Run-phase commits (worktree-local, not pushed)

| Commit | Milestone |
|--------|-----------|
| d5df06b5f | M1 korean.md re-author + SPEC artifacts + draft→in-progress |
| 0628d50f3 | M2 english copy layer |
| 3dde735e4 | M3 japanese copy layer |
| cd0d58793 | M4 chinese copy layer |
| c7dec5517 | M5 SKILL.md shared sections + attribution cleanup |
| 89fbac7c6 | M6 catalog 1.1.0 + hash regen + mirror sync (also carries 4 pre-existing stale agent-hash corrections from the mandated `gen-catalog-hashes --all` regen — mechanical cascade, no agent content change) |
| 6c2fa8aa6 | M3 follow-up: quarantined childhood-origin beat removed from JA-14 example |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-10
run_commit_sha: 6c2fa8aa6  # last implementation commit; the §E evidence commit (M7) follows and cannot self-reference — orchestrator may backfill
run_status: complete
ac_pass_count: 19   # of 19 checks (17 AC IDs incl. 006a/b/c)
ac_fail_count: 0
preserve_list_post_run_count: 6  # 3 EN/JA/ZH prose layers verified additive-only + 3 preserved SKILL.md frontmatter fields (user-invocable, allowed-tools, license)
l44_pre_commit_fetch: not-run  # isolated agent worktree, commits local-only per delegation; orchestrator owns the pre-merge fetch (main observed advanced to 911c338fd during the run — merge-time reconciliation is the orchestrator's step)
l44_post_push_fetch: n/a  # no push performed (delegation: commit locally only)
new_warnings_or_lints_introduced: none  # make build clean x3; internal/template test FAIL set identical to base HEAD (5 pre-existing, verified in detached worktree at 92d955daa)
cross_platform_build:
  darwin: exit 0 (go build ./...)
  windows: exit 0 (GOOS=windows GOARCH=amd64 go build ./...)
total_run_phase_files: 17  # 5 template skill md + 5 local mirror md + catalog.yaml + 6 SPEC artifacts
m1_to_mN_commit_strategy: per-milestone commits M1..M6 + one M3 follow-up fix + M7 evidence commit; worktree-local (no push); status transition draft→in-progress rode M1 (d5df06b5f)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-09T15:43:52Z
sync_commit_sha: "pending-backfill-sync"  # self-referential hazard — backfilled by follow-up chore commit per spec-frontmatter-schema.md D3 exemption
sync_status: complete
frontmatter_status_transitions:
  spec_md: "in-progress → completed (merged transition on single sync commit, 3-phase close); updated: 2026-07-10 already at sync date — no bump needed"
  plan_md: "n/a — header-only artifact, no YAML frontmatter"
  acceptance_md: "n/a — header-only artifact, no YAML frontmatter"
  progress_md: "n/a — header-only artifact, no YAML frontmatter"
ac_pass_count: 19   # of 19 checks (17 AC IDs incl. AC-HUM-006a/b/c) — acceptance.md SSOT
ac_fail_count: 0
changelog_entry_position: "[Unreleased] > Added (first entry)"
b12_self_test_a: "pre-emission grep -c 'SPEC-HUMANIZE-001' CHANGELOG.md == 0 (no duplicate; post-emission == 1)"
b12_self_test_b: "AC count 19 checks / 16 REQ verified against acceptance.md SSOT (grep -cE '^\\| AC-HUM-[0-9]+[a-c]? \\|' acceptance.md == 19)"
b12_self_test_c: "CHANGELOG file paths verified via ls: template skill dir (SKILL.md + modules/{korean,english,japanese,chinese}.md) + local mirror + catalog.yaml humanize entry version 1.1.0 + hash 07c6509a4"
mx_tag_validation: "n/a — markdown-only skill content in scope; MX tags bind source code, no .go/.sh files touched by this SPEC"
readme_docs_sync: "no change — skill-internal change; no user-facing doc surface beyond CHANGELOG (README + adk.mo.ai.kr docs-site untouched per delegation)"
```

## §F Phase 0.95 Mode Selection

- Inputs: tier=L, scope=7 files (5 skill md + NOTICE 제거 대상 없음 + catalog.yaml), domains=1 (skill documentation), language mix=100% markdown, concurrency benefit=LOW (single-skill coherent authoring, shared severity model)
- Mode evaluation: trivial=no (multi-file semantic authoring) / background=no (writes) / agent-team=no (prereqs unmet, domains<3) / parallel=no (modules share SKILL.md contract — coherence over parallelism) / workflow=no (<30 files, non-mechanical) / sub-agent=SELECTED
- Decision: sub-agent
- Justification: documentation-authoring analogous to coding-heavy work (Anthropic coding-task parallelism caveat); 4 modules must stay consistent with one shared severity/grading contract, so a single sequential manager-develop preserves cross-module coherence better than parallel spawns.
- Implementation Kickoff Approval: PASSED (user selected "run-phase 진입" at the gate); all preferences collected (license decision resolved v0.3.0).
