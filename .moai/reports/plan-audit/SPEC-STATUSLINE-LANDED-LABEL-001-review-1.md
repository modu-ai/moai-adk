# SPEC Review Report: SPEC-STATUSLINE-LANDED-LABEL-001
Iteration: 1/1 (Tier S ceiling = 1, final)
Verdict: PASS-WITH-DEBT (conditional — three must-fix edits before run kickoff)
Overall Score: 0.81 (Tier S PASS threshold 0.75; harmonic mean of the four dimensions)

Reasoning context ignored per M1 Context Isolation. The caller's summary of the lead decision was used only as a scope statement (what not to reopen). Every factual claim was checked against code. Audited tree: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1281`, branch `WT-statusline-landed-label`, HEAD `16bb6e19a`. Artifacts read: spec.md (119 lines), acceptance.md (49), plan.md (88), progress.md (19). Cross-model backends were not invoked because no `audit_model` key exists in `.moai/config/sections/`.

All seven must-pass rows pass and the score is above the threshold. Three blocking-class defects remain (D1–D3). Each is a one-to-three-line edit to acceptance.md or spec.md. D1 and D2 are verification-layer holes: as written, the named AC would pass against an implementation that does not satisfy the REQ. They must be fixed before M3/M5 runs, and a grep can confirm each fix.

## Code-claim verification (requested by the caller)

| Claim | Evidence | Result |
|---|---|---|
| `kanban.LandedScanArgs(ref) []string` | `internal/kanban/autodone_scan.go:222` returns `{"log", ref, "--format=%H%x00%ct%x00%s"}` | EXISTS, matches plan |
| `kanban.ScanLandedSubjects(run CommandRunner, ref string) ([]LandedCommit, error)` | autodone_scan.go:237. Calls `run("git", LandedScanArgs(ref)...)` once. Returns an error on runner failure and on each of 3 malformed-line shapes (L253/L257/L261) | EXISTS; one call per scan |
| `kanban.CommandRunner func(name string, args ...string) (string, error)` (no ctx, no dir) | `internal/kanban/prlink_landed.go:88` | Matches plan §B L17. The closure adapter in plan M3 L52 is sound |
| `kanban.LandedAttributions(commits, landedBranch) map[string]LandedCommit` keeps the newest commit | autodone_scan.go:273–285 (`if _, seen := out[id]; !seen`) | EXISTS; routes through `subjectAttribution` (no second matcher) |
| `kanban.LandedBranchFromRef(ref) string` | prlink_landed.go:270 | EXISTS |
| `kanban.AutoDoneSubjectFresh(hit LandedCommit, addedAt string) bool`, fails closed on unparsable added_at | autodone_scan.go:300–306 | EXISTS; supports AC-SLL-003b |
| Reference composition `ScanLandedSubjects → LandedAttributions(…, LandedBranchFromRef(ref)) → AutoDoneSubjectFresh(hit, it.AddedAt)` | `internal/cli/todo_autodone.go:213/217/315` | CONFIRMED |
| One git call per refresh | The scan itself is one call. But `RefreshLandedCounts` also calls `kanban.LandedRefFor(boardRoot)` (landed.go:179). When `worktree_base_branch` is unconfigured, that function runs `git -C <root> symbolic-ref refs/remotes/origin/HEAD` through a separate seam (`prlink_landedref.go:69`, `landedRefGitRun`) | ACHIEVABLE for the scan. The REQ-SLL-004 wording over-claims (see D3) |
| `⚑` U+2691 is one cell wide | Measured with go-runewidth v0.0.29 (the repo's pinned version): `⚑ U+2691 default=1 nonEA=1 EA=1 ambiguous=false` | SATISFIABLE. Caveat: `⚠ U+26A0` also measures 1 here, so AC-SLL-010's runewidth check is necessary but does not by itself separate emoji-risk glyphs. Pinning the exact code point closes that gap |
| `✓` occurs in no template file | `grep -rln "…✓N"` over `internal/template/templates`, `.claude/rules`, `.claude/skills`, and READMEs matched only the 4 docs-site `advanced/statusline.md` files | CONFIRMED (plan M5 L66) |
| spec lint | `moai spec lint`: `0 error(s), 0 warning(s)`, plus 1 INFO (no Authored-By-Agent trailer). `spec_audit`: modern_era_clean=1 | CLEAN |

## Must-Pass Results
- [PASS] MP-1 REQ number consistency: REQ-SLL-001…012 are contiguous and unique, with 3-digit padding (spec.md L49, L53, L57, L61, L65, L69, L73, L77, L81, L85, L89, L93).
- [PASS] MP-2 GEARS (judged on the requirement layer only): the Ubiquitous "The landed refresh shall…" form appears at L51, L63, L71, L83, L87, L91. L55 is the shall-not form. L59, L67, L75, L79 are When forms. L95 is a Where form. The Given/When/Then rows in acceptance.md are the verification layer and are graded under Group 4.
- [PASS] MP-3 frontmatter: all 12 canonical fields are present with correct types (spec.md L2–L13). `version: "0.1.0"` is quoted, `status: draft`, `priority: P2`, `lifecycle: spec-anchored`, `tags` is a CSV string. No rejected aliases. `tier: S` is an extra field.
- [N/A] MP-4 language neutrality: the SPEC targets a single Go package plus docs. The template tree is untouched (plan L66, verified above).
- [PASS] MP-5 D7: the only `SPEC-…` token in the SPEC body is its own id (spec.md L108 names "SPEC-status", which is not an id). No referenced SPEC exists that could be retired or superseded.
- [PASS] MP-6 D8: `grep syscall` over all four artifacts: 0 matches.
- [PASS] MP-7 clarification gate: `grep '\[NEEDS CLARIFICATION'` over plan.md/progress.md: 0 matches. research.md does not exist (Tier S).

## Category Scores (0.0-1.0, rubric-anchored)
| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | REQ-SLL-004 (L63) says "exactly one git invocation" but the ref-resolution path can add a second one (D3). REQ-SLL-005 (L67) and REQ-SLL-008 (L79) both fire on a failed refresh over an old-criterion cache and prescribe different outcomes (D4). §A L27 says "Two defects" and then lists three (D8). A reasonable engineer resolves each of these the same way (plan M1 L41 already does for D4). |
| Completeness | 0.90 | 1.0−0.75 | Sections present: HISTORY L17, context/WHY §A L23, REQs §B L47, constraints §C L97, three `### Out of Scope — …` H3s with bullets (L106, L111, L117). Missing from Out of Scope: auto-done's evidence form 1 (recorded delivering SHA) is silently dropped (D5). |
| Testability | 0.75 | 0.75 | Most ACs discriminate old from new behavior: AC-SLL-003, 004, 005, 005b, 006, 008, 012 each fail against the current code. AC-SLL-001 does not (D1). AC-SLL-011 leaves the `✓2` example unchecked (D2). AC-SLL-004 counts only one of the two git seams (D3). |
| Traceability | 0.85 | 1.0−0.75 | All 12 REQs appear in the AC matrix REQ column. Three REQ clauses have no AC and are covered only by DoD prose or an existing test not cited: REQ-004's "render path shall issue none", REQ-011's rationale-comment clause, REQ-010's "every conversation locale" clause (D6). |

Harmonic mean = 4 / (1/0.75 + 1/0.90 + 1/0.75 + 1/0.85) = 0.807.

## Defects Found (structured defect-list)

D1. AC-VACUOUS — acceptance.md:L14 (AC-SLL-001) — The fixture subject `docs: report (t999)` never contains the picked id `t101`. The CURRENT `countNamed` body-mention implementation (landed.go:227–250) also returns 0 on this input, so the AC passes whether or not REQ-SLL-002 is implemented. It does not discriminate. Its Then clause also claims coverage of the body-only case, which a fake runner cannot express. That case is actually pinned by AC-SLL-004 (argv equality) and AC-SLL-012 (the `%B` grep). — Severity: major — Class: blocking — Required fix: use a subject that NAMES `t101` in a non-attributing position. Examples: `chore: follow-up review notes for t101` (no positional form matches, so it attributes nothing), or `docs(t999): report mentions t101` (Form 1 scope attributes t999). Then `landed == 0` under the new criterion and `1` under the old one. Change the Then text to reference AC-SLL-004/AC-SLL-012 for the body-only clause instead of claiming it.

D2. AC-INCOMPLETE — acceptance.md:L30 (AC-SLL-011) — `grep -c '✓N'` does not catch the literal example `✓2`. That example appears on line 28 of all four `docs-site/content/{ko,en,ja,zh}/advanced/statusline.md` files. On en L28, `✓2` is the ONLY check mark on the line. An implementer who replaces only `✓N` passes the AC with four stale `✓2` examples still published, which violates REQ-SLL-012 ("shall no longer present ✓N"). The measured occurrences of `✓` in those files are limited to lines 28, 33, and 200, all of them the annotation. — Severity: major — Class: blocking — Required fix: change the check to `grep -c '✓' docs-site/content/{ko,en,ja,zh}/advanced/statusline.md` → 0 in every locale. Keep the `⚑N` ≥ 1 check, and add `grep -c '⚑2'` ≥ 1 (or the example's replacement) so the rendered example is also migrated.

D3. REQ-OVERCLAIM — spec.md:L63 (REQ-SLL-004) and acceptance.md:L18 (AC-SLL-004) — "exactly one git invocation per refresh" is not what the code will do. `RefreshLandedCounts` resolves the ref through `kanban.LandedRefFor`, which runs `git -C <root> symbolic-ref refs/remotes/origin/HEAD` through `landedRefGitRun` whenever `worktree_base_branch` is unconfigured (prlink_landedref.go:63–75). That includes every AC test root created with `t.TempDir()`. AC-SLL-004 counts only `landedGitRunner`, so it passes while the literal requirement is violated. The property that actually matters, and that the code can meet, is that the number of scan queries does not grow with the card count. — Severity: major — Class: blocking — Required fix: reword REQ-SLL-004 to "exactly one `git log` subject-stream query per refresh regardless of the number of picked cards; ref resolution is unchanged (at most one read-only `symbolic-ref` when the base branch is unconfigured)". Alternatively, keep the wording and have AC-SLL-004 also stub and count `kanban`'s ref seam. That route is not reachable from `internal/statusline` without a kanban change, which conflicts with §C L99, so the rewording is recommended.

D4. REQ-CONFLICT — spec.md:L67 vs L79 — REQ-SLL-005 says a failed query "shall keep the previously stored measurement unchanged apart from its timestamp". REQ-SLL-008 says that when the prior cache is old-criterion, the prior number shall not be carried forward. On a failed refresh over an old-criterion cache, both requirements fire and prescribe different outcomes. — Severity: minor — Class: blocking (internal consistency) — Required fix: scope REQ-SLL-005's first clause with "…keep the previously stored current-criterion measurement…", or add "except as REQ-SLL-008 requires".

D5. SCOPE-GAP — spec.md:L106–L109 — The SPEC says the count uses "the same predicate `moai todo auto-done` evaluates" (L51). But auto-done has two evidence forms, and the statusline adopts only form 2 (subject). Form 1 (a recorded delivering SHA reachable from the ref, `AutoDoneFormSHA`) needs a per-card reachability query, which the one-query constraint forbids. So a card closed by auto-done via form 1 alone will not count. This is not stated anywhere. The statusline number also differs from `auto-done --dry-run closed=N` by the M1/M2 skips, so `closed=4` (§A L29) is not an oracle for the new count. — Severity: minor — Class: optional — Required fix: add one bullet under "Out of Scope — close-policy guards": "Evidence form 1 (recorded delivering SHA) is not consulted: it needs a per-card reachability query; the count is subject-attribution only and is not expected to equal auto-done's `closed=` figure."

D6. TRACE-GAP — acceptance.md §D — Three REQ clauses have no AC row: (a) REQ-SLL-004 "the render path shall issue none" is already pinned by the existing `TestLandedRenderPath_SpawnsNoGit` (landed_test.go:206) but is not cited; (b) REQ-SLL-011's rationale-comment clause appears only in DoD L48; (c) REQ-SLL-010's "every conversation locale" clause is implicit (the renderer has no locale branch). — Severity: minor — Class: optional — Required fix: add an AC row citing `TestLandedRenderPath_SpawnsNoGit` for (a). For (b), add a grep row such as `grep -c 'auto-done' internal/statusline/renderer.go` ≥ 1 near the segment, or accept the DoD bullet.

D7. AC-PRECISION — acceptance.md:L15/L29 — AC-SLL-002 says "committer time after `added_at`" without a concrete value, and the same-second boundary (`>=` counts as fresh, autodone_scan.go:305) is untested. AC-SLL-010 relies on runewidth's default condition, which reads the environment locale. — Severity: minor — Class: optional — Required fix: give a concrete ct (e.g. `added_at + 60`), and optionally add an equal-second case. In AC-SLL-010, assert width 1 under both `EastAsianWidth=false` and `=true` (both measured as 1 on v0.0.29).

D8. WORDING — spec.md:L27 — "Two defects, one visible reading:" is followed by three numbered items. — Severity: minor — Class: optional — Required fix: change it to "Three defects".

## M1/M2 exclusion safety (requested)

Safe, with one residual risk that should be documented.

- **M2 (SPEC-status sync gate)**: Excluding it is exactly what the glyph semantics require. The count answers "a landing commit exists", and plan-only landings are meant to show (spec.md L108). REQ-SLL-009/010 carry the verify-before-done meaning, and REQ-SLL-011 keeps picked unchanged, so no reading of the number implies closability.
- **M1 (reissued-id collision)**: The failure M1 guards against (a predecessor's landing credited to a reissued id) is mostly handled by the clock. REQ-SLL-003 applies `AutoDoneSubjectFresh`, and that function's own comment (autodone_scan.go:290–294) names it as the cross-generation guard M1 cannot provide. `LandedAttributions` keeps the NEWEST attributing commit, so the freshness test runs on the most favorable commit and cannot wrongly exclude a genuine new-generation landing. Residual: if the predecessor's commit lands AFTER the reissued card's `added_at` (concurrent reuse), the count includes it. This is rare, and the verify-before-done glyph absorbs it, but it is an overcount the SPEC does not state. It could be added as a sub-bullet next to D5.
- Plan §H L81 already accepts that `subjectAttribution` recognizes only `t[0-9]+` (prlink_landed.go `subjectCardToken`).

## Recommendation

Verdict PASS-WITH-DEBT. Tier S allows no second iteration, so the lead or manager-spec should apply these edits before Implementation Kickoff:

1. D1: replace the AC-SLL-001 fixture with a subject that names `t101` in a non-attributing position. Confirm that the old `countNamed` would count it (1) and the new criterion does not (0).
2. D2: change AC-SLL-011 to `grep -c '✓'` = 0 across the 4 locale files (catches `✓2` on L28).
3. D3: reword REQ-SLL-004 to one `git log` subject query per refresh, independent of card count, with ref resolution unchanged.
4. D4: scope REQ-SLL-005's keep-unchanged clause to current-criterion caches.

D5–D8 are optional. D5 is recommended because it prevents a run-phase reader from using auto-done's `closed=` number as the expected statusline value.

Must-pass rationale: MP-1 L49–L93 contiguous; MP-2 every REQ matches a GEARS form (layer = requirements); MP-3 12/12 fields (L2–L13); MP-5/6/7 grep-verified with 0 matches; MP-4 N/A (single-package scope, template tree untouched).
