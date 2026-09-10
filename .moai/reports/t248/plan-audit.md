# SPEC Review Report: SPEC-AUDIT-BUILD-IDENTITY-001

Card: t248 · Worktree: `.claude/worktrees/t248` · Branch `WT-audit-binary-sha` · Measured tree `64bba61aa`
Iteration: 1/1 (Tier S ceiling = 1, `harness.yaml:76`)

**Verdict: FAIL**
**Overall Score: 0.85** (Tier S PASS threshold 0.75 — the score does NOT drive this verdict; the enumerated blocking defect-list does)

Reasoning context ignored per M1 Context Isolation. Read: `spec.md` (v0.1.1), `plan.md`, `acceptance.md`, `progress.md` — Tier S input contract is spec.md + plan.md; acceptance.md and progress.md were read additionally because the SPEC's ACs live there rather than inline in spec.md §3.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-ABI-001..008`, sequential, no gaps, no duplicates, uniform 3-digit padding. Measured: `grep -o "REQ-ABI-[0-9]*" spec.md | sort -u` → 8 distinct, 001→008. AC side `AC-ABI-001..009`, likewise clean.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-ABI-*` in `spec.md` §2), not the verification layer. All 8 match a GEARS pattern: 001 ubiquitous (`The … shall each carry`), 002 event-driven (`When a convergence result is persisted … shall`), 003 ubiquitous + unwanted (`shall not carry a version-string field`), 004 ubiquitous + unwanted (`A nested build-identity object shall not be introduced`), 005 unwanted + Where (`Where the binary carries no commit metadata …`), 006 event-driven + unwanted, 007 ubiquitous + unwanted, 008 unwanted. The Given-When-Then entries in `acceptance.md` are `AC-ABI-*` verification-layer artifacts and are correctly NOT graded here (M3 § Scope); they are graded under Group 4.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types: `id` (matches the implementation regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`, `internal/spec/lint.go:877` — multi-segment IDs are accepted; the narrower regex printed in `spec-frontmatter-schema.md` is the drifted artifact, not this SPEC), `title`, `version: "0.1.1"` (quoted semver), `status: draft` (valid, enum at `spec-frontmatter-schema.md:86`), `created`/`updated` ISO, `author`, `priority: High`, `phase: "v3.1.5 target"` (a release target, not a prohibited stage name), `module`, `lifecycle: spec-anchored`, `tags` (comma-separated string). Plus `tier: S`. No rejected snake_case alias present. Mechanical cross-check: `mcp__moai__spec_audit(project_root=<this worktree>)` returned one INFO finding only (`EraAutoDetected`, V3R5, H-3) — no `FrontmatterInvalid`.
- **[N/A] MP-4 Section 22 language neutrality** — single-language SPEC (Go only: `internal/cli`, `internal/binlag`, `pkg/version`). No multi-language tooling surface. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — 2 external SPEC references, both exist and neither is retired/superseded/archived: `.moai/specs/SPEC-BINARY-LAG-VISIBILITY-001/spec.md` → `status: completed`; `.moai/specs/SPEC-BINLAG-INVOCATION-001/spec.md` → `status: completed`. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c "syscall" spec.md` → `0`. Auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn "NEEDS CLARIFICATION" .moai/specs/SPEC-AUDIT-BUILD-IDENTITY-001/` → exit 1, zero matches. `progress.md` §E.1 independently states 미해결 0 and names the two operator decisions that closed D1/D2. No `research.md` (Tier S) — the plan.md half of the check executed and is clean.

**No must-pass failure.** The FAIL is driven entirely by the `## Defects Found` list below.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | Requirements are precise and each names its own falsifier. Three localized ambiguities a reasonable engineer could resolve differently: AC-ABI-007 does not name which entry point(s) it exercises (`acceptance.md` AC-ABI-007 "감사 결과를 낼 때"); AC-ABI-005 does not say how the "pre-change key set" is obtained inside a test that cannot read the pre-change tree; and neither `spec.md` §2 nor `acceptance.md` AC-ABI-006 says whether `build_commit` should be `"none"` or empty when `version.Commit == "none"`. |
| Completeness | 1.00 | 1.0 | HISTORY (2 rows), WHY (§1 + §1.1-§1.4), WHAT (§2 requirements, §3 criteria table), traceability (§2.1), Out of Scope with **five** `### Out of Scope — <topic>` H3 sub-headings each carrying specific `-` bullets (§4), Residual risk (§5). Frontmatter complete (MP-3). |
| Testability | 0.65 | 0.50-0.75 | Every one of the 9 ACs names an executable `go test -run <Name>` command — strong. Against that: AC-ABI-008 carries a grep expectation that is **false in this tree** and therefore requires a human judgment call to evaluate (D-1); AC-ABI-002 can be satisfied vacuously (D-2); AC-ABI-007 leaves its entry-point scope open, so the lag half of the feature is unpinned for two of three entry points (D-4). Three of nine criteria are not cleanly binary as written. |
| Traceability | 1.00 | 1.0 | `spec.md` §2.1 maps all 8 REQs to ≥1 AC; `acceptance.md` §D maps all 9 ACs back to a REQ that exists. No orphaned AC, no uncovered REQ. Verified by reading both tables against the `grep -o` ID inventories. |

Aggregate (unweighted mean of the four dimensions): **0.85**.

---

## Answers to the six named lines of attack

**1. Is the anti-mutant criterion actually anti-mutant? — YES, and the `[HARD]` fixture rule is genuinely load-bearing, but for a *different* mutant than the one the SPEC names.**

Computing the representative mutant (records `build_version: "v3.1.3"`, leaves `build_commit` empty) against AC-ABI-004's three assertions:
- (a) "`build_commit`이 빈 문자열이면 테스트는 실패한다" → **kills it**.
- (b) "두 결과의 `build_commit`은 서로 다르고" → kills it (both empty ⇒ equal).
- (c) "버전 문자열을 담은 키가 존재하면 테스트는 실패한다" → **kills it independently**.

So the representative mutant dies three times over, and dies under assertion (a) alone **without** the identical-version fixture pair. The `[HARD]` fixture rule (identical version, differing commit) is therefore NOT what kills the *stated* representative mutant. It IS load-bearing against a distinct and more subtle mutant: an implementation that populates the field *named* `build_commit` with the **version string**. Under a differing-version fixture pair that mutant satisfies assertion (b) (two different values); only under the identical-version pair does it collapse to two equal values and die. `acceptance.md`'s own rationale line ("버전 비교만으로도 두 케이스가 갈리기 때문이다") states exactly this mechanism. **Conclusion: the rule earns its `[HARD]`; the SPEC's §1.4 sentence pointing AC-ABI-004 at the version-only mutant is correct but under-states what the fixture rule is actually for.** Recorded as D-6 (optional, accuracy).

No other criterion silently lets the version-only mutant through: AC-ABI-001 pins `build_commit == "abc123def"` for all three entry points, which is fatal to it. AC-ABI-002 does not kill it (see D-2) but does not admit it either — it is simply silent.

**2. Empty-operand / vacuous assertions? — YES, one confirmed (D-2), one benign-by-design.**

AC-ABI-002 compares key names and asserts "`build_commit` 값이 서로 같다" across three results. With `omitempty` (mandated by REQ-ABI-004 and D1), an implementation emitting an empty `build_commit` everywhere produces three results in which the key is **absent** — the name comparison then compares nothing and the value comparison compares `"" == "" == ""`. The criterion holds vacuously. It is backstopped by AC-ABI-001, but AC-ABI-002 *alone* asserts nothing about a degenerate implementation. → D-2.

AC-ABI-005 is deliberately an assertion over absence (the `omitempty` round-trip) and is correct as an intentional negative control, not a vacuity defect. AC-ABI-009's selector was measured, not assumed: `grep -c "^func Test(Converge|RunMultiAudit|AuditMulti)"` over `internal/cli/*_test.go` → **29 matching tests**. Not a zero-match selector; that criterion is real.

**3. Can fail-open swallow the feature? — NO for the commit half; YES for the lag half.**

Commit half: an always-absent `build_commit` is killed by AC-ABI-001, which pins the value to `"abc123def"` for each of the three handlers given `version.Commit` set (settable in-test — `version.Commit` is an exported package var, `pkg/version/version.go:9`). Verified negative.

Lag half: `build_lag` empty is a **legitimate** outcome under REQ-ABI-005 (fail-open) and under `binlag.Advisory`, which returns `""` for every status except `StatusBehind` (`internal/binlag/binlag.go:156-158`). Nothing in the SPEC lets a reader distinguish "compared, fresh" from "never compared". Combined with D-3 and D-4 below, an implementation in which `build_lag` is **never** non-empty for `glm_audit` and `audit_multi` in real use passes every stated criterion. → D-3, D-4.

**4. Is the `internal/binlag` reuse mechanically verifiable? — HALF yes.**

Observation 1 of AC-ABI-008 (call-counter stub on `binlag.Comparer`) is genuinely mechanical: `Comparer` is an exported package-level `var` seam — `var Comparer = gitCompare` at `internal/binlag/binlag.go:81`, with `Evaluate` routing through it at `:90-91`. A test in `internal/cli` can replace it and count. That is not prose; that is a real seam.

Observation 2 (the source sweep) is **not** mechanically decidable as written. → D-1.

**5. Scope discipline — CLEAN on the card's three items, but the Tier S AC budget is exceeded.**

The eight REQs decompose to the card's three scope items (001/002 = record + persist, 006 = lag warning) plus five constraint requirements (003 field-shape prohibition, 004 additive/flat shape, 005 fail-open, 007 single constructor, 008 no new surface). None of the five adds scope; each narrows the implementation space of the three. `§4 Out of Scope` explicitly refuses the adjacent temptations (binary-lag repair, other MCP tools' identity, `PerBackendVerdict`, nested objects, new packages). **No over-engineering found on the scope axis.** Separately, the SPEC exceeds the Tier S acceptance-criterion ceiling → D-5.

**6. Coordinate accuracy — 8 of 11 exact, 3 drifted.** Full table below; two are 2-line drifts on the same symbol class (`resolveToolProjectRoot` call sites) and one is a line/quote mismatch. → D-7 (optional).

| Cited | Measured in `64bba61aa` | Status |
|---|---|---|
| `internal/cli/mcp_codex.go:262` `type ReviewOutput struct` | line 262 | EXACT |
| `internal/cli/mcp_convergence.go:89` `type PerBackendVerdict struct` | line 89 | EXACT |
| `internal/cli/mcp_convergence.go:106` `type ConvergenceResult struct` | line 106 | EXACT |
| `internal/cli/mcp_convergence.go:646` `persistConvergenceResult` | line 646 | EXACT |
| `internal/cli/mcp_audit_multi.go:145` text fallback | line 145 (`mcp.NewToolResultText(... overall=%s ...)`) | EXACT |
| `internal/cli/mcp_audit_multi.go:71` `resolveOptionalToolProjectRoot` | line 71 | EXACT |
| `internal/cli/doctor.go:518` consumer precedent | line 518 = `func checkBinaryFreshness` | EXACT |
| `internal/cli/doctor.go:530` `GetCommit()` consumer | line 530 | EXACT |
| `pkg/version/version.go:8` `Version = "v3.1.3"` | line 8 | EXACT |
| `internal/cli/mcp_codex.go:1495` `resolveToolProjectRoot(req)` | **line 1493** | DRIFT −2 |
| `internal/cli/mcp_glm.go:247` `resolveToolProjectRoot(req)` | **line 245** | DRIFT −2 |
| `pkg/version/version.go:37` quoting "Version is the string that cannot order two builds" | line 37 = `func GetBuildID()`; the quoted sentence is at **line 32** | MIXED (symbol line correct, quote line off by 5) |

The `internal/binlag` API described in §1.3 was verified whole and is accurate: `Evaluate(ctx, Request{Dir, BinaryCommit, BinaryVersion}) Verdict` (`:90`), the four statuses (`:30-45`), `Advisory` (`:156`), `Short` (`:169`), `RemedyCommand` (`:144`), and the documented "`BinaryVersion` takes NO part in the verdict" contract (`:56-59`) — which is exactly the grounding §1.4's anti-mutant argument claims it is.

---

## Defects Found

**D-1 — AC-ABI-008 source-sweep expectation is FALSE in this tree** — `.moai/specs/SPEC-AUDIT-BUILD-IDENTITY-001/acceptance.md` § AC-ABI-008 — **Severity: major — Class: blocking**
The criterion states `기대: 히트 0(또는 기존 히트만, internal/binlag 밖의 새 히트 0)`. Measured now, before any implementation:
```
$ grep -rn "merge-base\|is-ancestor" internal/cli --include='*.go' | grep -v _test.go
internal/cli/graph_stamp.go:68
internal/cli/graph_stamp.go:131
internal/cli/mcp_review_material.go:95
hits=3
```
Three defects compound here. (i) The primary stated expectation ("hits 0") is false today, so a run-phase actor executing the criterion verbatim reports a spurious failure. (ii) The escape clause "기존 히트만" records **no baseline count**, so distinguishing existing from new is an unrecorded human judgment, not a measurement. (iii) The qualifier "`internal/binlag` 밖의 새 히트" is incoherent for a grep scoped to `internal/cli`, which can never contain `internal/binlag`. Worse, hit (3) is **on the audit path itself**: `handleGLMAudit` → `collectReviewDiff` (`mcp_glm.go:257`) → `reviewDiffArgs` → `resolveReviewMergeBase` (`mcp_review_material.go:95`, `runReviewGit(root, "merge-base", ref, "HEAD")`). It is a legitimate diff-base resolution, not a binary-ancestry comparison — but the criterion as written cannot say so.
**Required fix**: replace the expectation with a recorded baseline and a delta predicate, e.g. "baseline in `64bba61aa` = 3 hits (`graph_stamp.go:68`, `graph_stamp.go:131`, `mcp_review_material.go:95`); PASS iff the post-change hit set is exactly this set", and add one sentence stating that `resolveReviewMergeBase` is a diff-base resolution and is explicitly not the ancestry comparison REQ-ABI-006 forbids duplicating.

**D-2 — AC-ABI-002 can be satisfied vacuously (empty operand)** — `acceptance.md` § AC-ABI-002 — **Severity: major — Class: blocking**
The criterion compares key names and `build_commit` values across three results but states no non-empty precondition. Under the mandated `omitempty` (REQ-ABI-004), an implementation emitting empty `build_commit` everywhere produces three results with the key **absent**; the name comparison then has an empty operand on all three sides and the value comparison reduces to `"" == "" == ""`. The criterion holds while asserting nothing. It is backstopped by AC-ABI-001, but a criterion whose own truth is compatible with the feature being entirely absent is a defect in the verification layer regardless of backstop.
**Required fix**: add a Given clause pinning `version.Commit` to a non-empty fixture value and a Then clause asserting `build_commit` is non-empty in all three results before the equality comparison runs.

**D-3 — REQ-ABI-006 is unreachable for `audit_multi` under its normal invocation, and `plan.md` D3 prescribes the unreachability** — `spec.md` §2 REQ-ABI-006 vs `plan.md` §B D3 / §D4 — **Severity: major — Class: blocking**
REQ-ABI-006 promises a lag advisory whenever the build commit is a strict ancestor of the reviewed tree's HEAD. `plan.md` D3 prescribes: "`projectRoot`가 비면 비교를 건너뛰고 `buildCommit`만 채운다". `audit_multi` obtains its root from `resolveOptionalToolProjectRoot`, which returns `("" , nil)` whenever the caller omits `project_root` (`internal/cli/mcp_project_root.go:101-107`) — and omitting it is the **documented normal case** for a session in the primary checkout (`.claude/rules/moai/core/moai-mcp-tools.md` § The `project_root` input: "Session in the primary checkout → pass nothing"). Under the SPEC's own prescribed design, therefore, the most common `audit_multi` invocation never runs the comparison and never emits a lag advisory, while REQ-ABI-006 reads as a guarantee. No acceptance criterion detects this: `build_lag` empty is indistinguishable from `StatusFresh`. The existing precedent the SPEC cites approvingly does the opposite — `checkBinaryFreshness` supplies `os.Getwd()` rather than requiring a caller-named root (`internal/cli/doctor.go:522-531`).
**Required fix**: either (a) add a requirement that the comparison falls back to the process working directory when `projectRoot` is empty (mirroring `doctor.go:522`), with an AC exercising the empty-`projectRoot` path and asserting a non-empty `build_lag` under a `StatusBehind` stub; or (b) narrow REQ-ABI-006 explicitly to "when a reviewed tree is named" and add a residual-risk entry stating that the default `audit_multi` call carries no lag signal — so the limitation is disclosed rather than silent.

**D-4 — AC-ABI-007 does not bind its entry-point scope, leaving the lag half unpinned for two of three entry points** — `acceptance.md` § AC-ABI-007 — **Severity: major — Class: blocking**
The criterion says only "**When** 감사 결과를 낼 때" — it never says *which* handler. An implementation that wires `binlag` into `codex_audit` alone satisfies AC-ABI-007 (one entry point exercised), AC-ABI-008 observation 1 ("감사 1회당 카운터가 1 이상" — one audit, one increment), and AC-ABI-002 (all three share the key *shape*; `build_lag` may be empty in all three). Contrast AC-ABI-001, which explicitly enumerates "codex_audit, glm_audit, audit_multi 각각의 핸들러" for the commit field. The commit half is pinned per-entry-point; the lag half is not.
**Required fix**: mirror AC-ABI-001's enumeration into AC-ABI-007 — require the `StatusBehind` stub observation and the `StatusFresh` control to be run against **each of the three handlers**, table-driven from the same two key constants AC-ABI-002 already mandates.

**D-5 — Tier S acceptance-criterion ceiling exceeded (9 > 8)** — `spec.md` frontmatter `tier: S` vs `acceptance.md` §D (9 rows) — **Severity: minor — Class: blocking**
`.claude/rules/moai/workflow/spec-workflow.md:146-150` sets the Tier S ceilings at 8 requirements **and** 8 acceptance criteria, applied independently, and states that exceeding either "is a signal to tier up or to split the SPEC, not to relax the budget". Measured: `AC-ABI-001..009` = 9 (requirements = 8, exactly at ceiling). Separately, the Tier S artifact set is defined as 2 files with AC inline in `spec.md` §3; this SPEC ships a full `acceptance.md` and reduces §3 to a pointer table. The artifact deviation is additive and I do not treat it as a defect on its own, but combined with the AC overage it is a signal that this SPEC sits at the S/M boundary.
**Required fix**: either merge two criteria (AC-ABI-001 and AC-ABI-002 are natural candidates — the second is a shape generalization of the first, and merging also resolves D-2), or reclassify to `tier: M`, which raises the ceiling to 16 and matches the 3-artifact set already shipped.

**D-6 — `spec.md` §1.4 under-states what the `[HARD]` fixture rule actually kills** — `spec.md` §1.4, last paragraph — **Severity: minor — Class: optional**
§1.4 says "§3의 AC-ABI-004가 이 뮤턴트에 대해 실패하는 기준이다", pointing the fixture rule at the version-only mutant. As computed in attack line 1 above, the version-only mutant is killed by the bare non-empty assertion regardless of fixture choice; the identical-version/differing-commit pair is load-bearing against a *different* mutant (a `build_commit` field populated with the version string). `acceptance.md`'s own bullet states the correct mechanism, so this is a narrative imprecision in `spec.md`, not a hole in the criterion.
**Required fix**: one sentence in §1.4 naming the second mutant — "필드 이름은 `build_commit`인데 값이 버전인 구현" — as the thing the `[HARD]` fixture rule kills.

**D-7 — three drifted coordinates** — `spec.md` §1.3, §1.2; `plan.md` §D4 — **Severity: minor — Class: optional**
`mcp_codex.go:1495` → actual **1493**; `mcp_glm.go:247` → actual **245** (both cite the `resolveToolProjectRoot(req)` call site; the SPEC's `mcp_glm.go:247` appears to have counted from the comment at `:238` rather than the statement at `:245`). `pkg/version/version.go:37` is correct as a symbol coordinate (`func GetBuildID()`) but the sentence quoted beside it lives at line 32. All three are 2-5 line drifts against a base the SPEC itself pins (`64bba61aa`), so they are cheap to correct and cheap to leave; none misidentifies a symbol.
**Required fix**: correct the three line numbers, or drop the line numbers and cite the symbol names (`handleCodexAudit`'s `resolveToolProjectRoot` call, `handleGLMAudit`'s), which do not drift.

**D-8 — `build_commit` value is unspecified for the `commit == "none"` case** — `spec.md` §2 REQ-ABI-005 / `acceptance.md` § AC-ABI-006 — **Severity: minor — Class: optional**
AC-ABI-006 pins the no-error outcome, the unchanged verdict, and `build_lag` empty for a `commit == "none"` dev build, but says nothing about `build_commit`. An implementation may emit `build_commit: "none"` — a string that *looks* like an identity to a downstream reader and defeats the card's whole purpose — or omit it, and both satisfy the criterion. `binlag.gitCompare` already treats `""`, `"none"`, and `"unknown"` identically as not-applicable (`internal/binlag/binlag.go:108-110`), which is the natural precedent for normalizing all three to an omitted field.
**Required fix**: add one Then clause to AC-ABI-006: `build_commit` is empty (and therefore omitted from the JSON) when `version.Commit` ∈ {`""`, `"none"`, `"unknown"`}.

**Blocking: D-1, D-2, D-3, D-4, D-5. Optional: D-6, D-7, D-8.**

---

## Recommendation

The SPEC is well above the Tier S score threshold and clears all seven must-pass criteria. It is a genuinely careful document: the anti-mutant criterion works, the `binlag` seam it relies on is a real substitutable `var` and not a prose assertion, the traceability tables are complete, and the Out-of-Scope section refuses five specific temptations by name rather than gesturing at them. The FAIL is narrow and its remedy is a single plan-phase revision touching `acceptance.md` and two sentences of `spec.md` — no re-planning.

Ordered fix instructions for manager-spec:

1. **D-3 first** (it is the only defect that changes what gets built). Decide between the `os.Getwd()` fallback and the narrowed-guarantee-plus-disclosure. If (a): add the fallback to REQ-ABI-006 and to `plan.md` §B D3, and extend AC-ABI-007 with an empty-`projectRoot` case. If (b): reword REQ-ABI-006 to bind only a named tree and add a §5 residual-risk bullet stating that a default `audit_multi` call carries no lag signal.
2. **D-4**: rewrite AC-ABI-007 to enumerate all three handlers, table-driven, mirroring AC-ABI-001's phrasing.
3. **D-1**: replace AC-ABI-008's grep expectation with the measured baseline (3 hits: `internal/cli/graph_stamp.go:68`, `:131`, `internal/cli/mcp_review_material.go:95`) and an exact-set predicate; add the one-sentence note that `resolveReviewMergeBase` is diff-base resolution, not binary ancestry.
4. **D-2 + D-5 together**: merge AC-ABI-002 into AC-ABI-001 as a shape clause with a non-empty `build_commit` precondition. This closes the vacuity hole and brings the criterion count to 8, restoring the Tier S budget in one edit. If the two criteria are judged worth keeping separate, reclassify to `tier: M` instead.
5. **D-6, D-7, D-8** at the orchestrator's discretion — each is a one-line edit and none blocks run-phase.

**Tier S ceiling note**: `harness.yaml:76` sets the Tier S plan-audit ceiling to 1, so iteration 2 is not available to this agent. After manager-spec applies the fixes, the operator decides whether to spend a re-audit (scoped to the D-1..D-5 delta) or to proceed on the corrected artifacts. The ceiling bounds iteration count, not verdict authority: this FAIL stands until an auditor re-reads the corrected artifacts, and an orchestrator self-assessment does not substitute for it (AP-SPD-004).

---

## Evidence and attribution

- **Baseline**: every measurement in this report was taken in `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t248` at `git rev-parse --short HEAD` = `64bba61aa`, branch `WT-audit-binary-sha`, via the `Bash`/`Grep` tools reading the tree directly. No value was carried over from another tree or another run.
- **Tool-served measurement, disclosed**: exactly one finding used a served tool — `mcp__moai__spec_audit(project_root=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t248)`, cited under MP-3 as a cross-check only. That MCP server self-reports build `v3.1.2 (commit 343399d2f)`, which is **not** this tree's HEAD. Per this very card's subject matter I record it rather than assume it: the frontmatter verdict does not rest on that tool — it rests on the field-by-field read of `spec.md`'s frontmatter block against `spec-frontmatter-schema.md:50-64` and on `internal/spec/lint.go:877`, both read from this tree.
- **Gaps (explicitly not observed)**: I did not compile or run any Go test — no implementation exists yet (`progress.md` §E.1: "구현 0줄"), so the nine AC commands are unexecutable by construction at plan-phase and I make no claim about whether they pass. I did not verify that `version.Commit` is settable from an `internal/cli` test by actually writing one; I verified only that it is an exported package-level `var` (`pkg/version/version.go:9`), which makes it settable in principle. I did not audit `internal/binlag`'s own correctness (out of scope per `spec.md` §4, closed by card t326).
- **Residual risk**: D-3's severity rests on the claim that omitting `project_root` is the *normal* `audit_multi` invocation. That is grounded in the documented contract (`moai-mcp-tools.md` § The `project_root` input) and in `resolveOptionalToolProjectRoot`'s empty-string return, not in call-frequency telemetry. If, in practice, every `audit_multi` caller passes `project_root`, D-3 degrades from a shipped-broken-feature defect to a disclosure gap — but the SPEC would still need the disclosure, so the required fix is unchanged in either case.
