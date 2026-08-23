# SPEC Review Report: SPEC-FEEDBACK-AUTO-SUBMIT-001

Iteration: 1/3
Verdict: **FAIL**
Overall Score: **0.75** — graded against the **Tier L PASS threshold 0.85**
(`.claude/rules/moai/workflow/spec-workflow.md:142`, the SSOT row: `L (Large) | > 1000 LOC or constitutional | > 15 files | 5 files | 0.85`).

Reasoning context ignored per M1 Context Isolation. The four lens reports under `.moai/reports/t170/` were read as *evidence artifacts* (the SPEC cites them as its measured ground truth), not as author reasoning; every claim below was re-verified against the working tree independently.

Tier L input contract satisfied: all 5 artifacts read (`spec.md` 249L, `plan.md` 173L, `acceptance.md` 322L, `design.md` 175L, `research.md` 68L).

---

## §1 Claim — what this audit asserts

1. The seven load-bearing claims the SPEC rests on are **true as stated**, with two enumeration imprecisions (D9, D10).
2. The SPEC's enforcement honesty is **genuinely good** — it does not overclaim "mechanized"; §E.3, `plan.md` AP-12 and `design.md` §1 bound the claim correctly.
3. Nevertheless the SPEC **fails**: it leaves one submission channel (the issue **title**) outside its only security control without naming it; it reuses a *detector's* pattern set as a *rewriter* without addressing the two marker-anchored patterns that are unsound as rewrite spans; it prescribes reuse of an **unexported** identifier; and two of its own AC selectors would print `ok` with nothing run — the exact failure mode `acceptance.md` opens by forbidding.
4. MP-2 fails on one requirement (REQ-12).

## §2 Evidence — verbatim observations

### Verification of the seven load-bearing claims

| # | Claim | Verdict | Observed evidence |
|---|---|---|---|
| 1 | `feedback.md` has no pre-submission confirmation; `gh issue create` at :118 is prose | **TRUE** | `grep -n 'AskUserQuestion\|gh issue create'` → `52`, `156`, `178` for AskUserQuestion; `118: The orchestrator executes directly: \`gh issue create --repo <resolved-target>\``. Nothing between 52 and 118. Prose, not Go. |
| 2 | `SensitiveContentPatterns` exported, case-insensitive, config-extensible, lacks `AIza` | **TRUE** | `internal/hook/pre_tool.go:262-273` — 9 patterns, no `AIza`; assigned to the exported field `SensitiveContentPatterns:` on the returned `*SecurityPolicy`; `compilePatterns` (`:101-112`) compiles `"(?i)" + p`; `MergeExtraPatterns` appends `extra.Security.ExtraSensitiveContentPatterns`. `AIza[0-9A-Za-z_-]{35}` present in `.moai/astgrep-rules/security/credentials.yml:23,35,47,59`. |
| 3 | Nothing rewrites arbitrary text to remove secrets | **TRUE** | `grep -rn 'func .*[Rr]edact\|func .*[Mm]ask' internal/ pkg/ cmd/ --include='*.go' \| grep -v _test` → exactly 3: `internal/cli/glm.go:454 maskAPIKey`, `internal/cli/glm_tools.go:992 maskPartial`, `internal/github/secret.go:144 MaskSecret`. All three take one value and return a display string. |
| 4 | No `~`-collapsing path helper | **TRUE** | `grep -rn '"~/"\|"~"' internal/ pkg/ cmd/ --include='*.go' \| grep -v _test` → 3 hits, all the reverse direction: `core/git/branch.go:179` (refname check), `shell/detect.go:135` (fallback literal), `shell/config.go:222` (`strings.HasPrefix(path, "~/")` — expansion). |
| 5 | `SECURITY.md` carries the routing rule REQ-7 cites | **TRUE** | `SECURITY.md:16` `1. **Do NOT** open a public GitHub issue for security vulnerabilities.` / `:17` … `[GitHub Security Advisories](https://github.com/modu-ai/moai-adk/security/advisories/new)`. |
| 6 | `kanban.BacklogStore` has the borrowed shape | **TRUE** | `internal/kanban/backlog_store.go:9-14` (sibling advisory lock, atomic rename, lock-free reads), `:30` imports `internal/atomicfile`, `:81-92` `Mutate` holds the sibling lock across the whole read-modify-write. |
| 7 | `feedback` is `RouteExcluded`, pinned by two tests | **TRUE** | `internal/settings/sectionroute_test.go:27` `"feedback": RouteExcluded,` under the comment `SPEC-WEBCONF-SIMPLIFY-001 M3`; `internal/web/scope_contract_test.go:79` lists `"feedback"` in the `excluded` slice. |

### Mechanical lint (domain tool, per verification-claim-integrity §1.1 surface 3)

```
$ ~/go/bin/moai spec lint .moai/specs/SPEC-FEEDBACK-AUTO-SUBMIT-001/spec.md
✓ No findings — all SPEC documents are valid
```

### Structural counts (independently confirmed, as instructed)

```
$ grep -c '^### REQ-' spec.md                 → 13     (REQ-1 … REQ-13, sequential, no gaps, no duplicates)
$ grep -c '^| AC-F-' acceptance.md            → 23
$ grep -c '^### AC-F-' acceptance.md          → 23
$ grep -o 'AC-F-0[0-9][0-9]' acceptance.md | sort -u | wc -l → 23
```
REQ 13 / AC 23 — the lead's measurement is confirmed. Both inside the Tier L ceilings (25 / 25).

---

## §3 Baseline-attribution

All observations were made **in this run, against this tree**: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t170`, `pwd` confirmed before the first command. Commands: `grep`, `sed`, `ls`, and one `~/go/bin/moai spec lint` invocation. No test was executed — see §7 Gaps.

---

## §4 Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — 13 entries, `REQ-1` … `REQ-13` at `spec.md:81,87,93,101,109,115,121,129,137,143,149,158,162`. Sequential, no gaps, no duplicates, consistent padding style. Corroborated by `moai spec lint` clean.
- **[FAIL] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in `spec.md`); the `Given/When/Then` entries in `acceptance.md` are the verification layer and were graded under Group 4, not here. 12 of 13 REQs match a GEARS pattern (event-driven `…되면/…경우, <subject>는 …해야 한다`, ubiquitous `<subject>는 …해야 한다`, unwanted `…해서는 안 된다(shall not)`), each naming a subject (워크플로 / 스크러버 / 명령 / 마법사 / 스킬 본문). **REQ-12 (`spec.md:160`) does not**: `웹 설정 화면에서 feedback.auto_submit을 토글할 수 있어야 한다` is a subject-less passive capability statement with no actor performing a `shall`, and its second sentence defers the behavior itself (`노출 방식은 … 착수 승인 시점에 확정한다`). One informal requirement = MP-2 FAIL.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types: `id`, `title`, `version`, `status: draft`, `created: 2026-08-22`, `updated: 2026-08-22`, `author`, `priority: High`, `phase: "v3.1.3"`, `module: "internal/feedback"`, `lifecycle: spec-anchored`, `tags: "…"` (`spec.md:2-15`). No rejected snake_case alias. MINOR deviation only: `version: 0.2.0` is unquoted where the schema shows `version: "X.Y.Z"` quoted (`spec-frontmatter-schema.md:40`); YAML still decodes it as a string and the lint is clean, so this is a style finding (D12), not a type failure.
- **[N/A] MP-4 language neutrality** — the SPEC is scoped to this Go project, not to multi-language tooling; it privileges none of the 16 supported programming languages, and REQ-13 (`spec.md:166`) carries the template-neutrality constraint explicitly. N/A auto-passes. The SPEC correctly separates the two language sets (16 programming vs 4 conversation) at §A.1 P3.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — 3 external refs, all resolve: `SPEC-TODO-ENABLE-FLAG-001` (`status: draft`), `SPEC-INVOCATION-MODEL-001` (`status: completed`), `SPEC-WEBCONF-SIMPLIFY-001` (`status: completed`). None in {retired, superseded, archived} → no BLOCKING. Notably the SPEC *does* reconcile the completed-SPEC interaction it creates: §D option A is described as `SPEC-WEBCONF-SIMPLIFY-001 M3의 기록된 결정을 되돌린다` with the two pinning tests named and a commit-message obligation.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → `0`. Auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-FEEDBACK-AUTO-SUBMIT-001/` → no output. No unresolved markers. See D8 below: an open decision is carried in prose instead of the marker convention, which is a finding but not an MP-7 failure.

---

## §5 Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 — minor ambiguity in one or two requirements | 11 of 13 REQs are unambiguous with file:line anchors. REQ-12 (`spec.md:160`) is passive + deferred; REQ-6 (`spec.md:117`) says the vocabulary `sandbox.defaultDenyList`에서 가져온다 while that identifier is unexported (D3), so "가져온다" hides a decision the implementer must make. |
| Completeness | 0.70 | between 0.50 and 0.75 | All required sections present (`§G HISTORY`, `§A` WHY, `§B` WHAT with 3 `### Out of Scope — <topic>` H3s each carrying specific `-` bullets, `§C` REQUIREMENTS, `acceptance.md` AC, `§E` constraints). But three substantive holes: the title channel is never scoped in or out (D1), the scrubber's project-root resolution and timeout behavior are unspecified (D5, D6), and the queue is introduced without reconciling the **already-existing** local-draft recovery artifact at `feedback.md:36-44` (D4). |
| Testability | 0.70 | between 0.50 and 0.75 | The AC discipline is above average — bipolar pairs, an explicit pre-edit FAIL baseline (AC-F-002), degenerate-implementation guards (AC-F-008 three-axis, AC-F-013 second assertion), a no-hardcoded-shape rule (AC-F-009), no weasel words anywhere. Degraded by AC-F-023 naming two selectors that match no test (D2) and by no AC covering the marker-anchored rewrite hazard (D1b). |
| Traceability | 0.80 | 0.75 — one indirect mapping | REQ→AC coverage is **complete**: 1→001, 2→002, 3→003/004/005/014, 4→006-009, 5→010, 6→011, 7→012/013, 8→015/016, 9→017/018, 10→019, 11→020/021/022, 12→023, 13→023. No orphan AC, no uncovered REQ. Degraded only by `plan.md` milestone exit criteria misreferencing AC ranges in 3 of 9 milestones (D7). |

Aggregate: **0.75** < Tier L threshold **0.85**.

---

## §6 Defects Found

**D1 — `internal/feedback` scrubs the BODY only; the issue TITLE reaches `gh issue create` unscrubbed, and the SPEC never names this path.** — `spec.md:95` (REQ-3 `표준 입력으로 받은 본문`), `design.md` §1 diagram (`stdin: 본문`), `design.md` §7 (`마스킹된 본문 전문`), vs `.claude/skills/moai/workflows/feedback.md:84` (`Inputs for the gh issue create invocation`) and `:102` (`**Title**: Written in conversation_language`) — the title is a **separately composed input** to the same command. The audit brief asked whether any path reaches `gh issue create` without passing the scrubber and whether the SPEC names it: this one exists and the SPEC does not name it. §E.3's second residual risk mentions the title only in the *duplicate-search* context (a pre-gate leak via `feedback.md:71`), never that the title is itself **submitted** unmasked. The confirmation gate compounds it: `design.md` §7 shows the user the masked body and the findings summary, so a secret in the title is invisible at the only human checkpoint too. — Severity: **critical** — Class: **blocking** — Required fix: either extend REQ-3 so `scrub` accepts and returns the title alongside the body (and AC-F-003's contract + `design.md` §1/§4/§7 with it), or add an explicit `### Out of Scope — 제목 스크럽` entry plus a fourth §E.3 residual risk stating in one sentence that titles are submitted unmasked. Silence is not an option — this is the SPEC's only security control and its declared scope leaks.

**D1b — reusing a detector's pattern set as a rewriter is unsound for the two marker-anchored patterns, and no AC probes it.** — `spec.md:103` (REQ-4 `재사용해야 한다`) against `internal/hook/pre_tool.go:263-264`: `-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----` and `-----BEGIN\s+CERTIFICATE-----` match only the **header marker**, not the key material. For a *detector* that is sufficient (the marker proves sensitivity); for a *rewriter* replacing the matched span it is an under-mask that removes the header and ships the entire private key body into a public issue — the worst possible outcome of this SPEC. Symmetrically, `compilePatterns` prepends `(?i)` (`pre_tool.go:104`), so `AKIA[0-9A-Z]{16}` becomes case-insensitive: harmless for a permission prompt, but in a rewriter it can eat an ordinary lowercase word-run inside the user's prose. The SPEC cites case-insensitivity as a virtue (`spec.md:103`) without noting the detector→rewriter asymmetry. AC-F-008 checks over-masking on one benign string only; nothing checks either failure. — Severity: **critical** — Class: **blocking** — Required fix: add a REQ-4 sub-clause stating that a matched span is replaced only when the pattern anchors the secret itself, and that marker-anchored patterns (`-----BEGIN …`) mask **through the block terminator**, not the marker; add two ACs — one asserting a full `PRIVATE KEY` block is absent from `Result.Body`, one asserting a lowercase prose run resembling `akia…` is not masked.

**D2 — AC-F-023 names two test selectors that match no existing test; the command exits 0 with nothing run.** — `acceptance.md` §AC-F-023, `go test ./internal/settings/ ./internal/web/ -run 'TestSchemaCurrentValuesReadsAllSections|TestSchemaLabel|TestSectionRoute|TestScopeContract'`. Observed function names: `internal/settings/schema_sections_test.go:441 TestSchemaCurrentValuesReadsAllSections` ✓; `internal/web/scope_contract_test.go:22,67 TestScopeContractEditableSections/Exclusions` ✓ (prefix matches); but `internal/web/schema_label_test.go` defines only `TestSchemaEmptyLabelParity:16`, `TestI18nKeySetParity:74`, `TestI18nSegmentKeysRemovedFromWebDictionary:133` — **no `TestSchemaLabel`**; and `internal/settings/sectionroute_test.go` defines only `TestRouteForSectionTable:8`, `TestSeamSectionsMatchesRoutes:51`, `TestExcludedSectionsAllRejected:74` — **no `TestSectionRoute`**. Half the selectors in the AC that guards the web/i18n/route-reversal surface are vacuous. `acceptance.md`'s own preamble forbids exactly this: `존재하지 않는 테스트를 가리키는 -run 패턴으로 통과를 주장하는 것, 아무것도 실행하지 않고 통과하는 AC는 둘 다 … 금지한다`. — Severity: **major** — Class: **blocking** — Required fix: replace `TestSchemaLabel` → `TestI18nKeySetParity`, `TestSectionRoute` → `TestRouteForSectionTable|TestExcludedSectionsAllRejected`, and require `-v` output-line inspection so a zero-test run is visible.

**D3 — REQ-6 prescribes reuse of an unexported identifier; `internal/feedback` cannot import it.** — `spec.md:119` (`민감 변수 이름 어휘는 sandbox.defaultDenyList(internal/sandbox/env.go:31-37)와 …에서 가져온다`) and `research.md` §2 (`재사용 | sandbox.defaultDenyList`). Observed: `internal/sandbox/env.go:32` `var defaultDenyList = []string{…}` — **lowercase, unexported**; `grep -n '^func [A-Z]\|^var [A-Z]' internal/sandbox/env.go` shows the only exported symbol on that surface is `ScrubEnv:51`, which the SPEC already rules out. There is no exported accessor. So the run-phase implementer must either export it (a change to another package that appears in no §B In-Scope line and no plan milestone) or copy the five names (contradicting `research.md`'s `재사용` verdict and the spirit of AP-4). The asymmetry is telling: REQ-4 explicitly verified exportedness (`이미 export돼 있고`); REQ-6 did not. — Severity: **major** — Class: **blocking** — Required fix: state the mechanism in REQ-6 — add `func DefaultEnvDenyList() []string` to `internal/sandbox` (and put that file in §B In Scope and a plan milestone), or explicitly accept a copied vocabulary with a parity test against `defaultDenyList`.

**D4 — REQ-9's retry queue is introduced without reconciling the already-existing failure-recovery artifact.** — `spec.md:139` (`.moai/state/feedback/queue.json`) vs `.claude/skills/moai/workflows/feedback.md:36-44`, an existing [HARD] block that already handles `gh` failure by writing `the full drafted title + body` to `.moai/state/feedback-draft-<timestamp>.md` and closes with `No drafted feedback is discarded on a gh failure; the local draft is the recovery artifact.` Two overlapping recovery surfaces under the same directory for the same failure class, one storing the **unmasked** title+body and one the masked body. `design.md` §7 references this convention for the gate's decline path, so the SPEC is aware of the file — but neither REQ-9 nor `design.md` §5 (the on-disk-artifact comparison table) mentions it, and `research.md` §2 records `실패한 외부 전송 스풀 0건` without qualifying that a draft path exists. — Severity: **major** — Class: **blocking** — Required fix: state in REQ-9 whether the queue replaces, wraps, or coexists with the `:36-44` draft path, and if it coexists, say which one a `gh issue create` failure writes.

**D5 — the scrubber's project-root resolution is unspecified, so REQ-8/REQ-9's paths are undefined under a different cwd.** — `spec.md:131,139` name project-relative paths (`.moai/logs/feedback-mask.log`, `.moai/state/feedback/queue.json`); `plan.md` M5 specifies the CLI as `stdin→stdout JSON` with no `--root`/`--cwd` flag; `design.md` §5 lists the paths but not how the root is found. The ACs (F-015, F-017) test with `t.TempDir()` as the project root, implying an injectable root the CLI surface does not declare. On the different-cwd path the two fail-open surfaces degrade (log and queue land elsewhere), so the control does not flip to fail-open — but the transparency guarantee REQ-8 sells silently evaporates. — Severity: **minor** — Class: **blocking** — Required fix: name the root-resolution rule in REQ-3 (walk-up to the `.moai/` marker, or an explicit flag) and add it to `design.md` §9.

**D6 — the fail-closed enumeration omits the hang and the malformed-stdout cases.** — `design.md` §9 covers binary absent / policy-load failure / regex-compile failure / mask-log failure / `HOME` unset / `gh` failure / queue-write failure. It does **not** cover (a) a scrubber that hangs — no timeout is specified anywhere in the SPEC, and a hung command yields neither a non-zero exit nor JSON, leaving the prose caller in an undefined wait; (b) stdout that is truncated or malformed at exit 0. Case (b) is arguably rescued by the two-sentence [HARD] rule (`verdict != ok` → don't submit, and an absent `verdict` is not `ok`), but the SPEC never says so, and a `jq` failure on malformed input is a third outcome the skill body has no instruction for. Also `design.md` §9's `HOME 미설정 → paths.Home() 에러` row is imprecise: `internal/paths/paths.go:55-60` falls back to `os.UserHomeDir()`, which commonly succeeds with `HOME` unset. — Severity: **minor** — Class: **blocking** — Required fix: add two rows to `design.md` §9 (timeout with a stated bound → fail-closed; unparseable stdout → fail-closed), make the skill's [HARD] clause a three-sentence rule, and correct the `paths.Home()` row.

**D7 — `plan.md` milestone exit criteria misreference AC ranges in 3 of 9 milestones.** — `plan.md` M4 Exit cites `AC-F-013~F-017` for the mask-log + queue milestone, but F-013 is the pre-mask classification AC and F-014 is idempotency; the milestone's ACs are F-015~F-018. M6 Exit cites `AC-F-002, F-018~F-021` for the skill + wizard milestone; F-018 is queue-resolve — the range should be F-019~F-022. M7 Exit cites `AC-F-022` for the web milestone; the web AC is F-023. A run-phase implementer following M4's exit list would verify the classifier ACs and never run the log/queue ACs. — Severity: **major** — Class: **blocking** — Required fix: renumber the three exit lists.

**D8 — an open decision that can change REQ-1 is carried in prose rather than the `[NEEDS CLARIFICATION]` convention.** — `spec.md:158-160` (REQ-12) + §D 결정 D5 + `plan.md` §A `남은 결정 1건`. Option B is stated to move the key to `workflow.feedback.auto_submit`, which changes REQ-1's key home and, per the SPEC itself, `SPEC 개정이 동반된다`. A decision whose resolution rewrites another requirement is exactly what the marker convention exists to gate; carrying it as prose means the mechanical clarification gate (MP-7) sees nothing. — Severity: **minor** — Class: **blocking** — Required fix: either mark it `[NEEDS CLARIFICATION: feedback.auto_submit 웹 노출 경로 — 선택지 A/B]` in `plan.md` so the gate fires, or resolve it now and delete option B.

**D9 — REQ-4's `Go 목록에는 없다` is repo-wide false.** — `spec.md:105`. Observed: `internal/github/workflow/validator.go:155` `` `AIza[0-9A-Za-z\\-_]{35}`, // Google API key `` — a **third** Go secret-pattern list that neither the SPEC nor `research.md` §2 surveyed. The statement is true of `hook`'s list only. The union direction is still correct (that list is workflow-file validation, not a reusable policy object), but the reuse/new table claims completeness it did not measure. — Severity: **minor** — Class: **optional** — Required fix: qualify to `hook의 목록에는 없다` and add one row to `research.md` §2 recording `validator.go:155` as surveyed-and-not-adopted with the reason.

**D10 — `research.md` §2's `~` grep enumeration is incomplete.** — `research.md` §2 says `grep 2건은 모두 반대 방향(branch.go:179, detect.go:135)`. Observed 3 hits; the unnamed third is `internal/shell/config.go:222` `strings.HasPrefix(path, "~/")`. Same direction (expansion), so the conclusion holds — the count does not. — Severity: **minor** — Class: **optional** — Required fix: correct the count to 3 and name the third.

**D11 — the confirmation gate's option labels are specified in Korean literals with no `conversation_language` obligation.** — `design.md` §7 table (`제출하지 않음 (권장)` / `이대로 제출` / `본문 수정 후 제출`). The gate lands in a skill body that ships to a 4-locale template; `askuser-protocol.md` requires option text in `conversation_language`. The SPEC gets the `(권장)`-on-first-option rule right but never states the language obligation, so a literal implementation ships Korean labels to every locale. The decline path itself **is** specified (option 1 → the `feedback.md:40` local-draft convention), so this is a language gap, not a missing-path gap. — Severity: **minor** — Class: **blocking** — Required fix: one sentence in `design.md` §7 requiring the three labels and the findings summary in `conversation_language`, with the English label set in the template mirror.

**D12 — `version: 0.2.0` unquoted.** — `spec.md:4` vs `spec-frontmatter-schema.md:40` (`semver X.Y.Z, quoted`). Decodes as a string; lint clean. — Severity: **minor** — Class: **optional** — Required fix: quote it.

### What the SPEC gets right (recorded so a revision does not regress it)

- **The enforcement claim is honest.** §E.3 (`변환은 테스트 가능한 Go지만, 강제는 규약 수준이다` … `자기 지시문을 무시하는 세션은 스크러버를 우회할 수 있다`), `design.md` §1 (`설계상 가장 중요한 한 줄` — code owns only the ②→③ transform), and `plan.md` AP-12 (reporting it as "masking is now enforced" is itself an anti-pattern) bound the claim correctly. This was the sharpest thing to attack and it survives — the SPEC does **not** overclaim mechanization. D1 is a *scope* leak inside an honestly-bounded control, not an overclaim.
- **fail-open / fail-closed axis separation is clean** and correctly justified: `design.md` §4's exit-code table keeps policy-block (`verdict`, stdout) off the tool-failure channel (exit code), which is why the [HARD] clause fits in two sentences. The gaps are the two unenumerated cases (D6), not the design.
- **Degenerate-implementation defenses are real**, not decorative: AC-F-008's three axes exclude both "mask everything" and "block everything"; AC-F-013's second assertion catches the silent-miss order inversion that `design.md` §3 nails down; AC-F-009 forbids hardcoding the mask shape by requiring the expectation to be produced by calling the adopted function; AC-F-002 requires observing a **pre-edit FAIL**.
- **design.md and research.md are load-bearing, not padding.** `design.md` §3 (pipeline order and why reversing it is a silent miss), §4 (the two-channel argument), §5 (why the log and the queue have different shapes), §6 (classifier signals + why the false-positive test is the only defense against a degenerate classifier) each carry decisions nothing else in the artifact set records. `research.md` §6 explicitly declares its own gap (`테스트를 하나도 실행하지 않았다 … 전부 소스 읽기 기반 예측이다`) — that self-limitation is what a research artifact is for.
- **The four card-premise corrections match the tree.** All seven load-bearing claims re-verified TRUE (§2). No drift from the lens reports was found.

---

## §7 Gaps (what this audit did NOT observe)

- **No test was executed.** Every "this test would fail / pass" statement in the SPEC remains a prediction, as `research.md` §6 itself declares. I verified test **names and line anchors** exist, not their behavior.
- I did not read `internal/cli/init_test.go`, `init_coverage_test.go`, or `wizard/wizard_test.go` in full, so additional wizard count/order assertions beyond `TestQuestionOrder` / `TestReconfigureQuestionsOrder` may exist (the SPEC flags this same gap).
- I did not audit the sibling `SPEC-TODO-ENABLE-FLAG-001`; the §E.1 shared-file table was read but its counterpart entries were not cross-checked against that SPEC's text.
- I did not verify the `internal/web` render/persist surface beyond the schema/i18n registration points, so REQ-12's true edit cost under either option is unconfirmed.
- I did not run `make build`, `golangci-lint`, or any `GOOS=windows go vet`.

## §8 Residual-risk

- The classifier (REQ-7 / `design.md` §6) has no precedent in this repo and no external calibration. Its false-negative cost is a public-channel leak. The SPEC's asymmetric-threshold instruction is right in direction but unquantified — nothing in the AC set can distinguish "conservatively tuned" from "arbitrarily tuned", and §D.5 correctly defers that to post-release reports.
- Even with D1 and D1b fixed, the control remains enforcement-by-convention (the SPEC says so). A session that ignores its own [HARD] clause bypasses everything. The follow-up card in `design.md` §10 (wrap `gh issue create` in Go) is the only structural fix and is out of scope here.
- The two-SPEC shared-file merge discipline (§E.1) is stated as a [HARD] rule but is enforced by nothing mechanical; the second SPEC to land will discover any violation as a merge conflict at best, a silently reordered question list at worst.

---

## §9 Recommendation

FAIL at 0.75 against the Tier L threshold 0.85, with MP-2 also failing. This is a **good SPEC with a leaking scope boundary**, not a weak one — the corrections are surgical.

MUST-FIX before re-audit, in priority order:

1. **D1** — scrub the title, or name it out of scope with a residual risk. (`spec.md` REQ-3, `design.md` §1/§4/§7, `acceptance.md` AC-F-003)
2. **D1b** — state the rewrite-span rule for marker-anchored patterns; add the private-key-block AC and the lowercase false-positive AC. (`spec.md` REQ-4, `acceptance.md`)
3. **D2** — fix AC-F-023's two vacuous `-run` selectors (`TestSchemaLabel` → `TestI18nKeySetParity`, `TestSectionRoute` → `TestRouteForSectionTable|TestExcludedSectionsAllRejected`).
4. **D3** — name the mechanism by which `internal/feedback` obtains the env-name vocabulary from an unexported `defaultDenyList`.
5. **MP-2 / D8** — rewrite REQ-12 in GEARS with a named subject, and either resolve decision D5 or carry it as a `[NEEDS CLARIFICATION]` marker in `plan.md`.
6. **D4** — reconcile the retry queue with the existing `feedback.md:36-44` draft artifact.
7. **D7** — correct the three milestone AC ranges in `plan.md` (M4, M6, M7).
8. **D5, D6, D11** — root resolution, the two missing failure rows, and the gate's `conversation_language` obligation.

Optional (orchestrator's discretion, per M6): D9, D10, D12.

Iteration 2 will be scoped to this enumerated defect delta plus a regression check over it.

**VERDICT: FAIL 0.75** (Tier L threshold 0.85)
