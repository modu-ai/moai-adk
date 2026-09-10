# SPEC Review Report: SPEC-VERSION-STAMP-GUARD-001
Iteration: 3/3
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.90** (iter-1 0.77 → iter-2 0.80 → iter-3 0.90 — monotonic, no STOP)
Tier: S (PASS threshold 0.75, `spec-workflow.md` § SPEC Complexity Tier)

Audited tree: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t388`, branch
`WT-version-sync-list`, HEAD `9328a52422baa13fdb7d7fd0c8409151da3ba3c1`
(`git rev-parse HEAD`). Artifacts read: `spec.md` (396 L), `plan.md` (207 L),
`acceptance.md` (286 L), `progress.md` (115 L). `design.md` / `research.md` absent —
correct for Tier S.

Reasoning context ignored per M1 Context Isolation. The dispatch's own verification of
D1/D3/D8 was re-run independently; none of it was consumed as evidence.

---

## Regression Check — iter-2 defects (D1-D9)

Every closure was re-measured in this tree. Nothing was accepted on the HISTORY row's claim.

| # | iter-2 severity | State | Evidence |
|---|---|---|---|
| D1 | major / blocking | **CLOSED (redesigned)** | see § D1 below |
| D2 | minor / blocking | **CLOSED** | `grep -n '단위 고정' spec.md` → the only hit is `spec.md:24`, the HISTORY row recording the removal. The body clause is gone. |
| D3 | minor / blocking | **CLOSED, and the replacement reproduces exactly** | see § D3 below |
| D4 | minor / blocking | **CLOSED** | `acceptance.md:215-239` replaces the judgement with two commands: (a) positive literals `partial` / `does not detect`, (b) a 5-literal deny list, both grep-judged. Decidable. |
| D5 | minor / blocking | **CLOSED** | `spec.md:209` now reads "shall fail when that count differs from **an expected entry count the check itself holds, independent of the parse** (7 …)". The comparison is no longer parse-vs-parse. `acceptance.md:177-180` states the same and forbids deriving 7 from the parse. |
| D6 | minor / optional | **CLOSED** | `spec.md:290` `[HARD]` 존재 보고는 **비치명**(`t.Errorf`), restated at `plan.md:58-60` and as an anti-pattern at `plan.md:185-186`. |
| D7 | minor / optional | **CLOSED** | `acceptance.md:205` Given is scoped to the document only; `acceptance.md:246-253` records the SPEC half as already-satisfied with a presence grep, so REQ-VSG-006 keeps coverage without a green-before-start Then. |
| D8 | minor / optional | **CLOSED in `spec.md`, SURVIVES in `plan.md`** | see F2 |
| D9 | minor / optional | **CLOSED** | `grep -n 'Where' spec.md` → only `spec.md:23,24` (HISTORY). REQ-VSG-004 (`spec.md:207`) folds the restriction into the main clause: "The check shall judge only that heading's paths; paths named under the release-artifact heading shall not be judged for existence…". |

No iter-2 defect appears unchanged in all three iterations — no stagnation flag.

### D1 — verified closed by redesign, not by rewording

The scheme now attributes each RED to one assertion, and I could not find a tree where the
attribution collapses:

- **E3-a (`plan.md:83-85`, `acceptance.md:188-196`)** — M1 landing tree. The anchor points at a
  stamp subheading that does not exist. I confirmed it does not: `sed -n '66,79p'
  .moai/docs/version-management.md` shows `### Files Requiring Version Sync` (66) with only
  `**Documentation Files:**` (70) and `**Configuration Files:**` (76) beneath it — the axis is
  doc-vs-config, not stamp-vs-artifact. The parse yields 0, the count assertion fires
  `version-stamp entries: parsed=0 expected=7`, and the existence assertion sees no path, so it
  asserts nothing. Single cause, correctly attributed. `plan.md:119-122` and `plan.md:180-182`
  name the old trap as an anti-pattern.
- **E3-c (`plan.md:90-92`, `acceptance.md:108-165`)** — the substitution. **The trick is sound.**
  Replacing `docs-site/hugo.toml` with `docs-site/nonexistent-stamp.toml` holds the entry count
  at 7, so the count assertion stays silent and the red has exactly one cause. `acceptance.md:131-132`
  adds the discriminator that makes this checkable rather than asserted: if a `parsed=` other
  than 7 appears in the same run, the substitution was an addition and the observation is void.
  Both properties of the planted path verified here: `test -e docs-site/nonexistent-stamp.toml`
  → exit 1, and its directory and extension both occur in the real stamp list (the substitution
  target is itself the example), closing the "unknown extension, skipped" escape.
- **Pinned literals.** `plan.md:61-68` fixes both messages before measurement:
  `version-stamp entries: parsed=<N> expected=7` and
  `version-sync list names a path that does not exist: <경로>`; `acceptance.md:129` fills the
  second with the actual planted path. Precise enough that "which assertion cried, and why" is
  a grep rather than a narration. One weakness in the *mechanism* rather than in the literals:
  see F3.

### D3 — the corrected enumeration reproduces to the digit

I re-ran the whole chain in this tree rather than checking the count:

    git grep -nE 'v[0-9]+\.[0-9]+\.[0-9]+' -- . ':!.moai/reports' ':!.moai/specs' \
      ':!.moai/release-notes' ':!CHANGELOG.md' ':!*_test.go' ':!docs-site/content/*/changelog*'
    → 2225 lines / 592 files                                    (spec.md:134 — match)

    awk -F: '{print $1}' <that output> | grep -E 'v[0-9]+\.[0-9]+\.[0-9]+' | sort | uniq -c | sort -rn
    →  40 docs/design/v2.14.0-release-plan.md
       24 .moai/release/RELEASE-NOTES-v2.17.0.md
       16 .moai/release/MIGRATION-v2.17.0.md
       12 .moai/release/RELEASE-NOTES-v2.16.0.md
        7 .moai/marketing/awesome-lists/github-release-v2.12.0-enhanced.md
        6 .moai/release/RELEASE-NOTES-v2.15.0.md
        4 .moai/release/v2.15.0-draft.md
        4 .moai/release/RELEASE-NOTES-v2.20.0.md
      113 total                                                  (spec.md:156-165 — exact match)
    ... | grep -c '^\.moai/release/'  → 66                       (spec.md:180 — match)

    -h histogram total 2494; -n-derived token total 2607; delta 113.               (spec.md:150,170)
    v2.14.0 72→112 (+40) · v2.12.0 83→90 (+7) · v2.17.0 25→65 (+40).               (spec.md:171-172)

The `.moai/release/` vs `.moai/release-notes/` warning at `spec.md:176-181` is measured and
correct — the deny list excludes the latter and not the former, and 66 of the 113 inflation
lines come from the included one.

---

## Independent verification of the other claims (not requested, run anyway)

Every RED-now cell, every cited line, every commit:

| Claim | Command | Result |
|---|---|---|
| ghost still in list (AC-VSG-001 RED-now) | `grep -c 'internal/template/templates/.moai/config/config.yaml' .moai/docs/version-management.md` | `1`, exit 0 — match |
| ghost path absent | `test -e internal/template/templates/.moai/config/config.yaml` | exit 1 — match |
| check file absent (AC-004/005 RED-now) | `test -e internal/cli/version_sync_list_test.go` | exit 1 — match |
| placeholder absent | `test -e .moai/release-notes/vX.Y.Z.ko.md` | exit 1; real files `v3.1.0.ko.md`, `v3.1.3.ko.md` — match |
| derived-value assertions (AC-003 RED-now) | `sed -n '1,14p' .moai/docs/version-management.md` | `:8` `reads from git tags at build time`, `:12` `via git describe` — verbatim match |
| doc item lines (D8) | `sed -n '66,79p'` | items 71-74, 77-78; 75 blank, 76 label — match |
| AC-006 RED-now | `grep -nF -e partial -e 'does not detect' .moai/docs/version-management.md` | rc=1 (0 hits, whole file) — match, stronger than claimed |
| AC-006 deny list | `grep -nF -e 'no longer rot' -e 'can no longer' -e 'fully prevent' -e 'guarantees that the list' -e 'ensures the list'` | rc=1 — 0 hits |
| §4 SPEC-half literal | `grep -cF '이 카드가 착지해도 목록은 여전히 썩을 수 있다' spec.md` | `1` — match |
| R-1: `61921f1ba` unreachable | `git merge-base --is-ancestor 61921f1ba HEAD` | rc=1 — match |
| §1: `eba919e44` / `175d63f3f` reachable | same | rc=0 both — the SPEC's asymmetric treatment is correct, not an oversight |
| §1 hugo.toml incident | `git show --stat eba919e44`; `git show eba919e44:pkg/version/version.go`; `…:docs-site/hugo.toml` | 6 files, no hugo.toml; `Version = "v3.1.3"` at line 8; `version = "v3.1.2"` at line 55 — verbatim match |
| §1.1 injection paths | `sed -n '20p;36p;72p' Makefile`; `sed -n '22p' .goreleaser.yml` | all four exact; `Version` sits inside a `var (` block, so "package-level `var`, not a constant" is right |
| R-1 fetch-depth refutation | `grep -c 'actions/checkout' .github/workflows/ci.yml` → 7; `grep -c 'fetch-depth: 0'` → 6; job `test:` starts 114, its checkout 127, `fetch-depth: 0` 129, `go test … ./...` 208 | same job — the refutation holds |
| §5 precedent | `internal/cli/deprecated_paths_text_reference_test.go` exists; `repoRootFromCLITest` at `hook_flush_test.go:22` | both — match |
| Template-First not triggered | `find internal/template/templates -name 'version-management*'` | no hits — `spec.md:294-296` correct |
| SPEC lint | `go build -o /tmp/t388-audit3-moai ./cmd/moai` (rc=0, `git status --porcelain -- cmd internal pkg` empty), then `/tmp/t388-audit3-moai spec lint .moai/specs/SPEC-VERSION-STAMP-GUARD-001/spec.md` | `✓ No findings`, exit 0. Built from this tree — the PATH binary is v3.1.2 / `343399d2f` and its green would be unattributable |
| `mcp__moai__spec_audit` (project_root = this worktree) | — | 1 INFO only (`EraAutoDetected` V3R5, H-3); no drift finding |

**Base drift, and why it does not invalidate the above.** `origin/develop` is now
`5928095ea`, **31 commits ahead** of the pinned tree, and `9328a5242` is an ancestor of it
(rc=0). I checked whether the drift reaches this SPEC's subjects:
`git log --oneline 9328a5242..origin/develop -- .moai/docs/version-management.md
pkg/version/version.go .moai/config/sections/system.yaml docs-site/hugo.toml README.md` → empty,
and `git diff --stat 9328a5242 origin/develop -- .moai/docs/version-management.md` → empty.
Every RED-now cell above survives the drift unchanged.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -oE '\*\*REQ-VSG-[0-9]+\*\*' spec.md | sort -u`
  → 001, 002, 003, 004, 005, 006. Contiguous, zero-padded, no duplicate. `grep -n '^### AC-VSG-'
  acceptance.md` → AC-VSG-001..006 at lines 35, 69, 87, 108, 169, 203.
- **[PASS] MP-2 GEARS format compliance (requirement layer)** — judged against the six
  `REQ-VSG-*` entries at `spec.md:201-211`, **not** against the ACs. All six are Ubiquitous
  (`The <subject> shall …`), several with a compound response; REQ-VSG-004 (`spec.md:207`) and
  REQ-VSG-006 (`spec.md:211`) additionally carry the GEARS Unwanted form (`shall not be judged`,
  `Neither shall assert`). No Given-When-Then appears in the requirement layer. The
  Given-When-Then in `acceptance.md` is the verification layer and is graded under Group 4.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present at `spec.md:1-15`
  with correct types: `id`, `title`, `version: "0.4.0"` (quoted), `status: draft` (in the 8-value
  enum, `spec-frontmatter-schema.md:86`), `created: 2026-08-31`, `updated: 2026-09-01`, `author`,
  `priority: Medium`, `phase`, `module`, `lifecycle: spec-anchored`, `tags` (comma-separated
  string). Plus `tier: S`. No rejected snake_case alias.
- **[N/A] MP-4 language neutrality** — single-language scoped: the deliverable is one Go test in
  `internal/cli/` guarding this repository's own `.moai/docs/version-management.md`. No
  multi-language tooling claim is made. Auto-passes per the MP-4 N/A rule.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `grep -Eoh 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'
  spec.md plan.md acceptance.md | sort -u` → `SPEC-VERSION-STAMP-GUARD-001` only. No external
  SPEC reference, so no retired/superseded reconciliation is owed. Card t392 is referenced as a
  **card**, not a SPEC, and `spec.md:243` / `:396` explicitly record it as "SPEC 미발행".
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -n 'syscall' spec.md plan.md acceptance.md`
  → rc=1. Auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-VERSION-STAMP-GUARD-001/`
  → rc=1, no match. `research.md` absent (Tier S); `plan.md` present and clean.

No must-pass failure. No BLOCKING D7 or D8 finding.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.85 | 0.75-1.0 | Every requirement has one reading. The per-assertion RED scheme is stated three times consistently (`spec.md` §5.1, `plan.md` §E/§F, `acceptance.md` §D header) and each names which assertion owns which red. One genuine ambiguity survives, at F1 — two clauses of `plan.md` M2 imply different entry counts for the same tree. |
| Completeness | 0.85 | 0.75-1.0 | HISTORY, problem, premise, GEARS requirements, traceability table, partial-guarantee section, check siting, five `### Out of Scope — <topic>` sub-headings each with specific bullets, five residual risks, references. Frontmatter complete. Deductions: `plan.md:29` carries a citation the previous iteration corrected elsewhere (F2), and the plan-phase artifacts are not committed (F3), which the `(none) → draft` row of the ownership matrix expects. |
| Testability | 0.90 | 0.75-1.0 | Every AC is decided by a named command. AC-VSG-006 item 3 became two greps (`acceptance.md:215-239`) — the D4 fix. Expected RED strings are fixed before measurement (`plan.md:61-68`). Mutant probes at `acceptance.md:162-165` and `:198-199` name three mutants and which clause kills each. Six edge cases at §E. Deduction: the grep scope for AC-VSG-006 is "문서의 해당 절" with no line range or heading name (F4). |
| Traceability | 1.00 | 1.0 | `spec.md:215-222` maps REQ-VSG-001..006 → AC-VSG-001..006 one-to-one. Verified in both directions by extraction: six REQ, six AC, no orphan AC, no uncovered REQ. Every internal `§n` reference resolves (`grep -oE '§[0-9]+(\.[0-9]+)?' spec.md` against the heading list; `§11` is an external `CLAUDE.local.md` reference). Every `plan.md` §D / E3-a/b/c and `spec.md` §1.1/§1.2/§3/§4/§5.1/§7 R-1/R-3/R-5 pointer cited from a sibling artifact exists. |

Aggregate: (0.85 + 0.85 + 0.90 + 1.00) / 4 = **0.90**.

---

## Defects Found

**F1** — `plan.md:131-136` vs `plan.md:149-151` — **two clauses imply different entry counts for
the M2.1 tree, and the plan does not say where the ghost lives at that moment.** M2.1 instructs
"아래 1·3·4·5·6항을 수행하되 2항(유령 제거)은 하지 않는다" and pins the expected output as
`parsed=8 expected=7` plus the existence assertion naming the ghost. For the existence assertion
to see the ghost, the ghost must sit **under the new stamp subheading**. But item 1 as written
constructs that subheading as "**버전 스탬프**(7경로)" and states it *replaces* the current
doc/config axis (`plan.md:150-151`) — which is where the ghost currently lives
(`.moai/docs/version-management.md:78`, under `**Configuration Files:**`, measured). An
implementer executing item 1 literally lands at `parsed=7`, the check goes green, and E3-b is
unobservable; the pinned expectation then fails to reproduce. Only the reading "at M2.1 the ghost
is carried into the stamp subheading, giving 8 entries" makes the pin true, and that reading is
inferred rather than written. The card's whole thesis is that inferred expectations rot silently.
Scope of the damage is bounded: E3-b is labelled auxiliary at `plan.md:89` and
`acceptance.md:157-160`, and neither AC-VSG-004 nor AC-VSG-005 draws its evidence from it —
their chains (E3-a, E3-c) are unaffected. — Severity: **minor** — Class: **blocking** —
Required fix: one clause in M2.1, e.g. "이 단계에서 유령 항목은 스탬프 소제목 아래에 그대로
둔다(따라서 항목 8건)".

**F2** — `plan.md:29` — **D8's correction landed in one artifact and not the other.**
`spec.md:30` and `spec.md:388` now read `항목 71-74·77-78행`; `plan.md:29` still reads
`항목 71-78행`. Measured: `grep -rn '71-78\|71-74' .moai/specs/SPEC-VERSION-STAMP-GUARD-001/`
→ `plan.md:29` (stale), `spec.md:30`/`:388` (corrected), `progress.md:56`/`:64` (records the
correction). iter-2 cited `spec.md:29` specifically, so D8-as-cited is closed; but the HISTORY row
at `spec.md:24` says the citation was corrected, and one of the two artifacts still carries the
uncorrected form. A reader of `plan.md` §B gets the wrong item enumeration. — Severity: **minor**
— Class: **blocking** — Required fix: `plan.md:29` → `항목 71-74·77-78행`.

**F3** — the plan-phase artifacts are **untracked**, so the anti-retrofit pin has no durable
record. `git status --porcelain -- .moai/specs/SPEC-VERSION-STAMP-GUARD-001 .moai/reports/t388`
→ `?? .moai/reports/t388/`, `?? .moai/specs/SPEC-VERSION-STAMP-GUARD-001/`. The design's central
device against a retrofitted expectation is that `plan.md` §D fixes the RED literals *before*
measurement; that property is only enforceable if editing them afterwards is visible in a diff.
With nothing committed, an implementer who sees `parsed=0 expected=6` can silently make §D say
6. Independently, the ownership matrix expects a `feat(SPEC-{ID}): plan-phase artifacts (S, 2
artifacts)` commit on the `(none) → draft` transition, and `progress.md` §E.1 (line 5) carries
neither `plan_status: audit-ready` nor `plan_complete_at`. This is an artifact-state defect, not
a SPEC-content defect — but it is the one that undermines the SPEC's own evidence discipline.
— Severity: **major** — Class: **blocking** — Required fix: commit the four plan-phase artifacts
(and this report set) before the first run-phase edit, and record the plan-phase completion signal
in `progress.md` §E.1.

**F4** — `acceptance.md:213`, `:227`, `:235` — AC-VSG-006 item 3's grep scope is unbound.
(a) says "문단이", (b) says "이 절 안에", the judging line says "`<문서의 해당 절>`"; the RED-now
line (`:242`) uses "「Files Requiring Version Sync」 절에 대고". Four phrasings, no line range and
no heading name fixed. Practical risk is low — measured, all seven literals are 0 across the
**whole** file (`grep -nF … .moai/docs/version-management.md` → rc=1 for both sets), so the safe
superset is available and gives the same verdict today. — Severity: **minor** — Class: **optional**
— Required fix: name the scope once, e.g. "「Files Requiring Version Sync」 절 전체(또는 문서 전체)".

**F5** — `plan.md:104-105` — the anchor string itself is the one unpinned literal in a design
built on pinning. M1.1 says to hold the stamp-subheading string as a constant "M2의 문서 수정과
짝을 맞춘다", but neither `plan.md` nor `acceptance.md` states what that string is, while both
pin the two assertion messages verbatim. M2 must then reproduce a string chosen in M1 from memory,
and `spec.md` §7 R-5 already names heading rename as the check's silent-stop vector. Cost of
closing it is one line. — Severity: **minor** — Class: **optional** — Required fix: fix the
subheading literal in `plan.md` §D alongside the two message contracts.

**F6** — `acceptance.md:7` — "측정 시점 `origin/develop`과 동일" is now false and unverifiable.
`git rev-parse origin/develop` → `5928095ea`, 31 commits ahead of the pinned `9328a5242` (which
is an ancestor, rc=0). The tense scopes the claim to the moment of measurement, so it is not an
error so much as a claim no later reader can check. Substantively harmless: I verified no commit
in that range touches any measured subject and the target document is byte-identical across it.
— Severity: **minor** — Class: **optional** — Required fix: drop the parenthetical, or restate as
"측정 시점 `origin/develop` = `<sha>`".

**F7** — carried residual, no fix required. Two AC conditions are green before run-phase begins:
AC-VSG-003 item 3 (`constant` / `상수` = 0 — measured `grep -niF -e constant -e 상수
.moai/docs/version-management.md` → rc=1) and AC-VSG-006 (b) (the five-literal deny list, already
0). Both are regression guards rather than failing inputs, and each sits inside an AC whose
overall RED is established by its other items (AC-003 items 1-2, AC-006 item (a)). This is the
shape iter-2 raised as D7 and it is correctly handled here; recorded so a later reader does not
mistake it for an unnoticed gap. — Severity: **minor** — Class: **optional** — Required fix: none.

---

## §4 partial-guarantee section — not weakened

Checked directly against the dispatch's item 5. `spec.md:229-247` is intact and, if anything,
sharper than in 0.3.0: the `[HARD]` opener ("이 카드가 착지해도 목록은 여전히 썩을 수 있다"),
the two-direction table naming which one closes and why the other cannot, the concession that the
uncaught direction *is* the incident that produced the card, and the closing prohibition on
writing an over-claim into either the document or the SPEC.

I also swept the four artifacts and the target document for any statement reading as "the list can
no longer rot": `grep -nF` over the five deny literals returns rc=1 on the document, and the
Korean-side equivalents appear only inside the prohibition sentences themselves
(`spec.md:245-247`, `plan.md:69`, `plan.md:191`). Nothing anywhere claims the guarantee is total.

---

## Run-phase readiness

**The SPEC is fit to enter run-phase.** The remaining work is wording and artifact hygiene, not
design: no requirement is ambiguous, no AC is undecidable, and the two evidence chains the card
turns on (E3-a for the count assertion, E3-c for the existence assertion) are each attributable
to one cause and pinned to a literal written before measurement.

What the implementer does first, and whether the SPEC says so: **yes, unambiguously.** `plan.md`
§C requires re-reading HEAD against `9328a5242`, confirming `internal/cli/version_sync_list_test.go`
does not exist, and re-reading the current list formatting — all three verified reproducible in
this tree. M1 then lands the check *before* any documentation edit (`plan.md:100-124`), holding
the expected count 7 as a constant, and the first observation is the count assertion crying
`version-stamp entries: parsed=0 expected=7`. `plan.md` §G names the trap of reversing that order.

Three things to do before the first run-phase edit:

1. **F3 first** — commit the plan-phase artifacts. Until they are in git, the pinned literals
   are not pinned to anything, and the card's own anti-retrofit device is unenforceable.
2. **F1** — one clause in M2.1 saying the ghost stays under the stamp subheading, so the
   `parsed=8` pin is reproducible.
3. **F2** — the one-token citation fix in `plan.md:29`.

F4-F6 can ride the same commit; F7 needs nothing.

The debt label is deliberate rather than generous: the score (0.90) and the must-pass firewall
both say PASS, but F1 and F3 touch the SPEC's own evidence discipline, and calling the artifact
defect-free would be the rationalization this audit exists to refuse.
