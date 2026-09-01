# SPEC Review Report: SPEC-VERSION-STAMP-GUARD-001

- Iteration: 1/1 (Tier S ceiling — `harness.plan_audit_tier_ceilings` S=1)
- Verdict: **FAIL**
- Overall Score: **0.77** (Tier S PASS threshold 0.75 — the aggregate alone would have passed; the FAIL is driven by D2, a critical blocking defect under M6)
- Audited tree: `9328a52422baa13fdb7d7fd0c8409151da3ba3c1`, worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t388`, branch `WT-version-sync-list`
- Reasoning context ignored per M1 Context Isolation. The dispatch carried the lead's own measurements; every one of them was re-measured in this tree before being used, and the re-measurements are quoted below.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -n 'REQ-VSG-[0-9]*' spec.md`: REQ-VSG-001..007 at spec.md:108,110,112,114,116,118,120. Sequential, no gap, no duplicate, uniform 3-digit padding.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-VSG-*` in spec.md §2); the `AC-VSG-*` Given-When-Then entries in acceptance.md are the verification layer and are graded under Group 4, not here. 001/002/003/004/006/007 are Ubiquitous (`The <subject> shall …`); 005 is a compound `**When** [trigger] … it shall fail …` plus a `**Where** …` clause. All seven match a GEARS pattern. See D8 for a MINOR semantic note on the `Where` clause.
- **[PASS] MP-3 YAML frontmatter validity** — spec.md:1-15. All 12 canonical fields present with correct types: `id`, `title`, `version: "0.1.0"` (quoted), `status: draft`, `created: 2026-08-31`, `updated: 2026-08-31`, `author`, `priority: Medium`, `phase: "v3.1.5 target"`, `module`, `lifecycle: spec-anchored`, `tags` (comma string). No rejected snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id`) present. `tier: S` additionally present.
- **[N/A] MP-4 language neutrality** — single-language (Go) repo-internal SPEC. Verified nothing lands in template-bound content: `find internal/template/templates -name 'version-management*'` returned empty output. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — the SPEC-ID extraction over all four artifacts returns only the self-reference `SPEC-VERSION-STAMP-GUARD-001`. No external SPEC reference, therefore no retired/superseded reconciliation obligation. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` returned `0`. Auto-PASS.
- **[PASS] MP-7 clarification gate** — the `[NEEDS CLARIFICATION` sweep over the SPEC directory returned rc=1 with no matches. `research.md` absent (Tier S), `plan.md` present and clean.

Corroboration: `moai spec lint` (binary built from this tree, per `.moai/reports/t388/spec-lint.log`) reported `✓ No findings`, exit 0. `mcp__moai__spec_audit` with `project_root` set to this worktree returned one `INFO / EraAutoDetected` finding only, no drift.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.60 | between 0.50 and 0.75 | REQ statements are unambiguous and every external citation I checked is exact (Makefile:20/36/72, ci.yml:208, doc:71-78, hugo.toml:55, version.go:8). But §3 — the load-bearing scope section — contradicts itself on its own survivor set (D1) and never fixes its counting unit (D6), so a reasonable engineer implementing spec.md:151-166 would build a different-sized thing than spec.md:275 (R-4) expects. |
| Completeness | 0.70 | between 0.50 and 0.75 | All required sections present; five `### Out of Scope — <topic>` sub-headings with specific bullets (spec.md:223,229,235,241,246). The gap is evidential, not structural: the false-positive area that justifies the whole guard shape was measured against ONE literal string while the predicate the SPEC then writes covers all of them (D2). |
| Testability | 0.78 | 0.75 | Zero weasel words (the `적절 / 합리적 / 충분히 / 적당 / reasonable / appropriate / adequate` sweep over spec.md + acceptance.md returned rc=1, no matches). AC-VSG-005 is exemplary — its bidirectional requirement provably defeats a constant-return implementation. Deductions: AC-VSG-001's stated method needs an object this tree does not contain (D4); AC-VSG-006 does not pin its synthetic input (D7); AC-VSG-002's judging command is illustrative (`판정 명령(예)`, acceptance.md:58) rather than pinned. |
| Traceability | 1.00 | 1.0 | spec.md §2.1 (lines 124-132) and acceptance.md §D (lines 13-21) agree in both directions. Every REQ-VSG-00N has at least one AC; every AC-VSG-00N names an existing REQ. No orphan AC, no uncovered REQ. |

Aggregate: arithmetic mean (0.60 + 0.70 + 0.78 + 1.00) / 4 = **0.77**.

---

## Defects Found

**D1** — `spec.md:144-149` vs `spec.md:153-157` — §3 enumerates **10** files as what survives after excluding the record series, and that list includes `internal/binlag/binlag_test.go` and `internal/cli/hook_install_precommit_attribution_test.go`. The scope bullet immediately below then puts `*_test.go` in the deny list, so those two cannot survive the SPEC's own scope. Only **8** do. Re-measured in this tree with the deny-list exactly as §3 states it (record series plus `*_test.go` plus `docs-site/content/*/changelog*`): **8 files**, **12 lines**. The 8 are the 4 READMEs (line 781 each) and the 4 `docs-site/content/*/getting-started/faq.md` (2 lines each); all 12 lines are the same class — a statusline upgrade example `🗿 v3.1.2 -> 🗿 v3.1.3`. There is no test fixture among them. The two named `_test.go` files do exist with exactly 1 hit each, so the "10" is a real count — just taken *before* the deny-list it is then paired with. The wrong figure propagates into three downstream consumers: the exemption bullet (`spec.md:160`, "위 10건 중"), R-4 (`spec.md:275`, "§3의 10건 중"), and plan.md M3.2 ("예상: `spec.md` §3의 10건 언저리"). — Severity: **major** — Class: **blocking** — Required fix: restate §3's survivor set as 8 files / 12 lines measured under the full deny-list, drop the two `_test.go` entries from the enumeration (or move them to an explicit "deny-listed, shown for completeness" note), and re-derive `spec.md:160`, `spec.md:275`, and plan.md M3.2 from the corrected figure.

**D2** — `spec.md:141,144` vs `spec.md:158-159` and `spec.md:116` (REQ-VSG-005) — **the measured false-positive area is not the area the stated predicate produces.** §3 justifies the guard's shape by measuring occurrences of the single literal string `v3.1.2` (87 files tracked; 10 after excluding records). But §3's 판정 is *"남은 파일에서 `vX.Y.Z` 꼴 토큰을 찾아 권위 버전과 대조한다. 다른 값이면 실패"*, and REQ-VSG-005 states it normatively: *any* in-scope version token differing from the authoritative version fails unless exempted. §7 R-1 (`spec.md:259-262`) confirms this reading explicitly.

Under that predicate, measured in this tree with the SPEC's own full deny-list: **2,225 matching lines across 592 distinct files**. Token histogram head: `270 v3.0.0`, `112 v2.14.0`, `90 v2.12.0`, `80 v3.1.1`, `80 v2.1.219`, `68 v2.1.198`, `62 v1.0.0`. Top directories: `docs-site/content` 535, `go.sum` 243, `internal/template` 233, `.claude/rules` 145. Restricting even to this product's own `v3.*` family does not rescue the figure: **488 occurrences across 226 files** are `v3.x.y` and not `v3.1.3` — and `v3.*` is not a safe narrowing either, since `v3.13.1` / `v3.5.0` in that set are third-party versions.

So the first run of the repo guard yields roughly 2,225 findings across 592 files, not "10건 언저리". This collides head-on with three [HARD] obligations that all assume the small number: `spec.md:275-277` (R-4, "각 건을 개별로 분류해야 한다 … 일괄 면제로 초록을 만들면 이 카드가 막으려던 공허한 통과를 스스로 만들어 낸다"), `acceptance.md:90-92`, and `plan.md:45` plus M3.3. Per-item classification of 592 files is not Tier S work, and the only cheap way out is the blanket exemption the SPEC itself names as its first anti-pattern (`plan.md:124`). The SPEC's stated reason for choosing this predicate — *"「스탬프란 무엇인가」라는 내구성 있는 정의를 발명하지 않는다 … 정의가 필요 없는 술어를 쓴다"* (`spec.md:164-166`) — is precisely what produces the explosion: without a discriminator between "this repo's own version stamp" and "some other software's version", every dependency pin, every Claude Code version reference, and every historical version mention in `.claude/rules` is a finding. — Severity: **critical** — Class: **blocking** — Required fix: decide and state the discriminator in §3 before run-phase, then re-measure the first-run surface under the decided predicate and re-derive R-4 and plan M3 from the measured figure. Note that any narrowing widens R-3's blind area, so R-3 must be restated against the narrowed predicate rather than left as written.

**D3** — `spec.md:185-187` / `plan.md:82-83` vs `spec.md:153-155` — the mandated RED fixture is itself an unexempted finding for the repo guard. §4 requires [HARD] that the `eba919e44` fixture be a **committed file** under `internal/versionstamp/testdata/eba919e44/`, and that fixture by construction contains `version = "v3.1.2"` (verified at the pin: `eba919e44:docs-site/hugo.toml` line 55 reads `version = "v3.1.2"`). That path matches no entry in §3's deny list — it is not `*_test.go`, not `.moai/reports/`, `.moai/specs/`, `.moai/release-notes/`, `CHANGELOG.md`, or `docs-site/content/*/changelog*`. So AC-VSG-004 (repo guard green at HEAD) fails on the artifact AC-VSG-005 mandates, and §3's enumerated deny-list never anticipates a `testdata/` class. — Severity: **major** — Class: **blocking** — Required fix: add `testdata/` (or the specific fixture directory) to §3's deny list with its reason, or state in §4 that the fixture requires a reasoned exemption entry — either way, decide it in the SPEC rather than leaving run-phase to discover it and reach for a blanket exemption.

**D4** — `acceptance.md:28` — AC-VSG-001's stated judging method requires a repository object this tree does not contain, which is the exact pattern §4 and plan §D forbid [HARD] elsewhere. The AC says to compare the doc's list against *"v3.1.4 범프 커밋 `61921f1ba`가 바꾼 경로 집합"*. Measured: the ancestry check of `61921f1ba` against HEAD returns rc=1 (not an ancestor), and the containing-branch query returns only `release/v3.1.4` and `remotes/origin/release/v3.1.4`. `61921f1ba` lives only on the unmerged release branch. `spec.md:185-187` and `plan.md:43` both state [HARD] that verification must not query repository objects, because a CI shallow checkout will not have them — and this AC does exactly that, against an object that is not even on the card's own branch. The fix is cheap and already half-written: `acceptance.md:31-39` states the expected 7-path set literally. (For the record, the underlying set is correct: the numstat of `61921f1ba` shows 7 paths and 9 changed lines, matching `spec.md:31-36` and `baseline.md` Claim 2 exactly.) — Severity: **major** — Class: **blocking** — Required fix: make the literal 7-path set in `acceptance.md:33-39` the judge, and demote `61921f1ba` to a provenance citation. Note in §7 that the authoritative stamp set was derived from a commit not reachable from the card's branch.

**D5** — `spec.md:112` (REQ-VSG-003) and `acceptance.md:71` — both call `Version` a "constant" / "상수", contradicting §1.2's own explicit instruction two paragraphs earlier: *"(`Version`은 `const`가 아니라 패키지 수준 `var`다 … 아래에서 「상수」로 부르지 않고 「기본값」 또는 「폴백」으로 부른다.)"* (`spec.md:78-79`). Verified against source — `pkg/version/version.go:7-8` is `var (` followed by `Version = "v3.1.3"`. The terminological error appears inside the very requirement that mandates correcting a terminological error in the target document. — Severity: **minor** — Class: **blocking** — Required fix: replace "constant" / "상수" with "default" / "fallback" in REQ-VSG-003 and AC-VSG-003 item 2.

**D6** — `spec.md:141,144,146` — unit ambiguity in the section whose figures drive the classification obligation. "87건" and "10건" are file counts (I re-measured 87 as files), but "각 1건" attached to the READMEs reads as a line count, and the four `faq.md` files carry 2 lines each — so the survivor set is 8 files / 12 lines and file-count and line-count diverge. Neither §3, nor R-4, nor plan M3.3 states which unit the "개별 분류" obligation is denominated in. — Severity: **minor** — Class: **blocking** — Required fix: fix the unit explicitly in §3 (findings are (path, line, observed) tuples per plan M1.1, so lines is the natural unit) and use that unit consistently in R-4 and plan M3.

**D7** — `acceptance.md:127-132` — AC-VSG-006 does not pin its synthetic input. "임의의 감시 대상 파일" leaves an implementation carrying a hardcoded extension or directory allowlist able to pass, which weakens the AC that exists specifically to prove the guard is not list-shaped. — Severity: **minor** — Class: **optional** — Required fix: pin the synthetic input to a path whose extension and directory appear nowhere in the doc's stamp list.

**D8** — `spec.md:116` (REQ-VSG-005) — the `**Where** an occurrence is neither authoritative nor exempted …` clause states a data condition, not a capability gate / feature flag / static config, which is what GEARS `Where` denotes; `While` fits the intent. The clause is also largely redundant with the preceding `When`. Syntactically it matches a GEARS pattern, so MP-2 passes. — Severity: **minor** — Class: **optional** — Required fix: recast as `While` or fold into the `When` clause.

**D9** — `spec.md:120` (REQ-VSG-007) — "This check **may** read the list" places a modal in normative text (RQ-5). It reads as a permission grant rather than a weasel obligation, so it is defensible as written. — Severity: **minor** — Class: **optional** — Required fix: optional; "is permitted to read the list" removes the modal without changing meaning.

**D10** — `spec.md:73-76` — §1.2 concludes that without LDFLAGS *"기본값이 그대로 사용자에게 보인다"*, citing only the Makefile. Released binaries do not take that path: `.goreleaser.yml:20-24` injects `-X github.com/modu-ai/moai-adk/pkg/version.Version={{.Version}}` as well. The conclusion still holds — the default is a hand-maintained fallback and every bump edits it, and `CLAUDE.local.md` §11 independently records the bare `go install ./cmd/moai` path as one that actually occurs — but the user-visible surface is narrower than the phrasing implies. — Severity: **minor** — Class: **optional** — Required fix: qualify the sentence to the hand-built path and cite `.goreleaser.yml:20-24` alongside the Makefile.

---

## Rulings on the questions raised in the dispatch

**The §3 inconsistency (independently confirmed).** My measurement reproduces the lead's exactly — 8 files / 12 lines under the SPEC's stated deny-list, against §3's stated 10. On the lead's question — documentation slip, or under-specified scope? — **it is both, and the under-specification is the larger of the two.** D1 is the slip. D2 is the under-specification, and it is an order of magnitude worse than the `*_test.go` question: the SPEC measures its false-positive area against one literal string and then writes a predicate over all version tokens. The `*_test.go` width the SPEC worries about in R-3 is a rounding error next to the 592-file surface the predicate actually produces.

**Non-vacuity of the guard (dispatch item 1) — the strongest part of the SPEC.** The RED pin is real and I verified both objects at the pin myself: `eba919e44:pkg/version/version.go` line 8 reads `Version = "v3.1.3"`; `eba919e44:docs-site/hugo.toml` line 55 reads `version = "v3.1.2"`. The numstat of `eba919e44` confirms 6 files with `docs-site/hugo.toml` absent — the omission the SPEC describes. On whether AC-VSG-005 as written forces a non-constant implementation: **yes.** The bidirectional requirement (`acceptance.md:114-115`) defeats a constant return — an implementation always answering `hugo.toml:55 / v3.1.2` passes the forward direction and fails the reverse. AC-VSG-006 then defeats the next-cheapest cheat, a hardcoded two-path scanner, though it is under-pinned (D7). On the committed-fixture requirement surviving a CI shallow checkout: it is stated firmly and in three places (`spec.md:185-187`, `acceptance.md:117-118`, `plan.md:43`), plus an anti-pattern entry (`plan.md:128`). That is firm enough. The failure is that the same discipline was not applied to AC-VSG-001 (D4) — the rule is stated well and then not swept for.

**AC mechanical checkability (dispatch item 2).** All seven are decidable by command and none rests on a judgement call dressed as a check — zero weasel words. Two deductions only: AC-VSG-001's method needs an unreachable object (D4), and AC-VSG-006's input is unpinned (D7). AC-VSG-002's command is explicitly illustrative (`판정 명령(예)`), which is acceptable for a doc-structure AC whose Then-clause already states a set-intersection-is-empty condition.

**The version.go ruling (dispatch item 3) — holds.** I verified the injection path independently: `Makefile:20` carries `-X $(MODULE)/pkg/version.Version=$(VERSION)`, applied at `Makefile:36` (`go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/moai`) and `Makefile:72` (`go install $(LDFLAGS) ./cmd/moai`) — all three citations exact. `pkg/version/version.go:7-8` is a package-level `var`, not a `const`. The two doc assertions the SPEC calls wrong are verbatim at `.moai/docs/version-management.md:8` and `:12`, and the stamp list is exactly at `:71-78` as cited. The reasoning is sound and the prescribed correction is precise, with two blemishes: the SPEC calls the var a "constant" in the requirement itself (D5), and it names only the Makefile when GoReleaser also injects (D10).

**Scope discipline (dispatch item 4) — currently held, but D2's honest fix threatens it.** Seven requirements for Tier S is not creep: 001/002/003 are the doc fix, 004/005/006 are the one guard, and 007 is a small existence check that guards exactly the ghost class that motivated the card. §6 Out of Scope is unusually strong — five specific sub-headings including an explicit refusal to reopen the operator-rejected Option B. The warning is forward-looking: resolving D2 requires deciding a discriminator, which is close to the "durable definition of a stamp" the SPEC deliberately declined to invent. Whichever way that decision goes, re-check the Tier S classification afterwards.

**R-2 / R-3 honesty (dispatch item 5) — both defensible, both verified true.** R-2: `docs-site/hugo.toml:56` reads `releaseDate = "2026-08-24"` at HEAD and `"2026-08-21"` at `eba919e44`, so it does move every bump and is not a `vX.Y.Z` token. Closing it needs a second, date-shaped predicate, which is genuinely a different mechanism — leaving it open with the stated mitigation (name it in the doc list) is proportionate, not an evasion. Notably `hugo.toml:54` already carries an in-file SSOT comment covering both lines. R-3: the claim *"실측상 지금 그런 사이트는 없지만(§3의 2건은 픽스처)"* is **true** — the only two `*_test.go` hits are `internal/cli/hook_install_precommit_attribution_test.go:19` (`const pinnedReleasedHookTag = "v3.1.2"`) and `internal/binlag/binlag_test.go:147` (`BinaryVersion: "v3.1.2"`), both genuine fixtures. Leaving R-3 open is defensible **as written**; it stops being defensible once D2 is fixed by narrowing the predicate, because narrowing widens the blind area R-3 describes. That coupling must be stated when D2 is resolved.

**Template-First (dispatch item 6) — the SPEC's claim is correct.** Verified: the guard lands in `internal/versionstamp/`, which does not yet exist (the directory listing errors, and the tracked-file grep for `versionstamp` returns nothing) — so plan §C.2's pre-check passes and there is no name collision. The template search for `version-management*` under `internal/template/templates` returns nothing, so the target doc has no template mirror; `internal/template/templates/.moai/docs/` exists but does not carry it. Nothing lands under `.claude/` or `internal/template/templates/`. §5's claim holds as stated.

---

## Recommendation

The score would pass a Tier S threshold and the must-pass firewall is clean; the FAIL is D2's doing, and I want that to be legible rather than hidden inside an aggregate. This is a well-built SPEC — the RED pin is real, the bidirectional non-vacuity test is the right instrument and is correctly specified, the citations are exact, and Out of Scope is disciplined. It fails on one thing: the section that decides how much work the guard creates measured the wrong quantity, and three [HARD] obligations downstream are written against that wrong quantity. Entering run-phase in this state leads to one of two places — a discovery of 592 files that stops the milestone, or the blanket exemption that manufactures exactly the vacuous green this card exists to prevent.

Fix route, in order:

1. **D2 (critical)** — decide the discriminator between this repo's own version stamps and every other software's version tokens, state it in §3, and re-measure the first-run surface under the decided predicate. Re-derive R-4 and plan M3 from the measured figure. Restate R-3 against the narrowed predicate. Re-check the Tier S classification once the decision is made.
2. **D4 (major)** — make `acceptance.md:33-39`'s literal 7-path set the judge for AC-VSG-001; demote `61921f1ba` to provenance. Record in §7 that the authoritative set came from a commit not reachable from the card's branch.
3. **D3 (major)** — resolve the fixture-vs-guard collision in the SPEC: add `testdata/` to §3's deny list with its reason, or mandate a reasoned exemption entry in §4.
4. **D1 (major)** — restate §3's survivor set as 8 files / 12 lines under the full deny-list; re-derive `spec.md:160`, `spec.md:275`, plan M3.2.
5. **D6, D5 (minor, blocking)** — fix the counting unit; stop calling the `var` a constant.
6. **D7, D8, D9, D10 (minor, optional)** — orchestrator's discretion. D7 is the one worth taking: it costs a line and hardens the AC that proves the guard is not list-shaped.

---

## Appendix — verification commands

All executed in `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t388` at HEAD `9328a52422baa13fdb7d7fd0c8409151da3ba3c1`. Listed by purpose; the shell forms are reproduced in the session transcript.

| Purpose | Command (abbreviated) | Observed |
|---|---|---|
| Tree identity | `git rev-parse --show-toplevel` / `HEAD` / `--show-current` | worktree t388, `9328a5242`, `WT-version-sync-list` |
| REQ enumeration | `grep -n 'REQ-VSG-[0-9]*' spec.md` | 7 REQ at :108,110,112,114,116,118,120 |
| AC enumeration | `grep -n '^### AC-' acceptance.md` | 7 AC at :25,47,63,79,96,125,136 |
| Out of Scope | `grep -n '^### Out of Scope' spec.md` | 5 sub-headings |
| D7 refs | SPEC-ID extraction over the 4 artifacts | self-reference only |
| D8 | `grep -c 'syscall' spec.md` | `0` |
| MP-7 | `[NEEDS CLARIFICATION` sweep over the SPEC dir | rc=1, no match |
| D1 survivor set | `git grep -l 'v3\.1\.2'` with the §3 deny-list pathspecs | 8 files |
| D1 line count | same with `-n` | 12 lines |
| §3 "87건" | `git grep -l 'v3\.1\.2' -- .` | 87 files |
| §3 test-file entries | `git grep -c 'v3\.1\.2'` on the two named test files | 1 hit each |
| **D2 surface** | `git grep -nE 'v[0-9]+\.[0-9]+\.[0-9]+'` with the §3 deny-list pathspecs | **2225 lines / 592 files** |
| D2 own-family | same, restricted to `v3.*` excluding `v3.1.3` | 488 occurrences / 226 files |
| Authoritative version | `sed -n '1,15p' pkg/version/version.go` | `var ( Version = "v3.1.3"` |
| LDFLAGS | `grep -n 'LDFLAGS :='` + `grep -nF 'LDFLAGS)'` on Makefile | :20, :36, :72 |
| CI runner | `grep -n 'go test -json' .github/workflows/ci.yml` | :208 |
| Doc list span | `grep -n 'Files Requiring Version Sync'` + `sed -n '66,80p'` | heading :66, bullets :71-78 |
| Ghost path | `git cat-file -e HEAD:internal/template/templates/.moai/config/config.yaml` | rc=128, does not exist |
| v3.1.4 bump | `git show --numstat --format= 61921f1ba` | 7 paths, 9 lines |
| v3.1.3 bump | `git show --numstat --format= eba919e44` | 6 paths, hugo.toml absent |
| RED pin (a) | `git cat-file -p eba919e44:pkg/version/version.go` line 8 | `Version = "v3.1.3"` |
| RED pin (b) | `git cat-file -p eba919e44:docs-site/hugo.toml` line 55 | `version = "v3.1.2"` |
| D4 ancestry | `git merge-base --is-ancestor 61921f1ba HEAD` | rc=1 (not an ancestor) |
| D4 ancestry | `git merge-base --is-ancestor eba919e44 HEAD` | rc=0 (ancestor) |
| D4 location | `git branch -a --contains 61921f1ba` | `release/v3.1.4` only |
| Deny-list not a no-op | `git ls-files 'docs-site/content/*/changelog*'` | 4 real files |
| R-2 | `sed -n '54,57p' docs-site/hugo.toml` | `releaseDate = "2026-08-24"` |
| §1.3 render claim | `sed -n '1,12p' …/system.yaml.tmpl` | `{{.Version}}` at :6, :9 |
| Template-First | `find internal/template/templates -name 'version-management*'` | empty |
| Path collision | `ls -d internal/versionstamp` / tracked-file grep | absent / empty |
| D10 | `grep -n -A8 'ldflags' .goreleaser.yml` | injects Version at :20-24 |
| Weasel scan | `적절\|합리적\|충분히\|적당\|reasonable\|appropriate\|adequate` over spec+acceptance | rc=1, no match |
| SPEC lifecycle | `mcp__moai__spec_audit` (project_root = this worktree) | 1 INFO finding, no drift |
