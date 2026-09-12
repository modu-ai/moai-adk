# SPEC-INIT-TUX-I18N-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready (iteration 3)
plan_complete_at: 2026-09-11

Plan 단계 산출물(manager-spec, 카드 t586, Tier L): spec.md v0.2.2(REQ 18개), plan.md, acceptance.md(AC 22개), design.md, research.md, progress.md. 워크트리 `.claude/worktrees/t586`, 브랜치 `WT-init-tux-i18n`, 착수 HEAD `18144b7aca714ea8924363b1eab4640cf101c6d0`, v0.2.1 개정 착수 HEAD `d8ebb39298b206ec8cc4183e728f27f48994df86`, v0.2.2 개정 착수 HEAD `538b56f1923c7b72e8dcb8379d55d05e4fadb1c5`.

1회차 plan 감사(`.moai/reports/t586/plan-audit.md`, FAIL 0.67) 결함 D1~D17 을 반영했다. 리드 판정 Q1~Q4 로 확인 필요 표식 4건을 닫았고, 남은 표식은 없다.

v0.2.1 개정(리드 판정 Q5, 2회차 감사 전): REQ-ITI-017 제안(`agent_wiring`·`autonomy_tier` 를 `Agents & Autonomy` 한 그룹으로, init 위저드 3페이지 → 2페이지)을 확정했다. 페이지 수(AC-ITI-018)와 스테퍼 분모(AC-ITI-021)를 서로 독립인 AC 로 나눴고, 그룹 라벨 비렌더와 번역 키 없음을 AC-ITI-022 로 고정했다(측정 `research.md` §6.1). `research.md` §13 의 실행 확인 4건은 `plan.md` §F M1 착수 검증 V-a~V-d 로 옮겨 추적한다 — run 단계 증거 절(§E.2)은 run 단계 소유라 이 개정에서 자리를 만들지 않았다. 2회차 plan 감사는 FAIL 0.84 였다(`.moai/reports/t586/plan-audit-iter2.md`, 결함 N1~N10). 요구 18개가 Tier M 상한 16 을 넘어 Tier L 로 올렸다(리드 판정: 분할하지 않음). Tier L plan 감사 통과 기준은 0.85 다.

v0.2.2 개정(2회차 감사 결함 N1~N10 반영, 3회차 감사 전): 실제 HOME 무기록 판정을 트리 전체 매니페스트에서 코드로 도출한 감시 목록(W1~W6)으로 바꾸고 가짜 HOME 양성 대조군을 더했다(N1, `acceptance.md` §B P8·AC-ITI-020 (4), 도출은 `research.md` §14). AC-ITI-003 의 pty 판정에서 단계 표시 줄 조건을 뺐다(N2). AC-ITI-010 (4) 에 S2 양성 절·S3·S4·S6 성질 제거 뮤턴트를 더했다(N3). 부재 단정 세 곳에 같은 형태의 대조군을 붙였다(N4, `research.md` §16). tmux `-e` 로 넘기는 자식 환경 정리 목록과 실효 환경 관측을 넣었다(N5, `research.md` §15). N6~N10 은 서술을 고쳤다. REQ 18·AC 22, Tier L 유지. 3회차 plan 감사는 FAIL 0.90 이었다(`.moai/reports/t586/plan-audit-iter3.md`, blocking F1, optional F2~F5).

v0.2.3 개정(3회차 감사 결함 F1~F5 반영, 착수 HEAD `268cffe2cb47107c8fcf720301df974c04e383e5`): S2 양성 절 기준을 저장 함수 호출 줄(정의 줄 제외)로 바꿨다(F1, `design.md` §10, `acceptance.md` AC-ITI-010 (2)(4), 호출 줄 양성·정의 줄 음성 대조 측정은 `research.md` §17). F2~F5 는 서술을 고쳤다. REQ 18·AC 22, Tier L 유지. 이 개정의 재감사 결과는 아직 없다.

t583 선행 게이트: t583(SPEC-INIT-QUIET-WIZARD-001)은 plan 단계이고 미커밋이다. `questions.go`·`wizard.go`·`types.go`·`translations.go`·`init.go` 를 건드리는 마일스톤(M4~M8)은 t583 병합·흡수와 흡수 트리 재측정 뒤에만 시작한다(`spec.md` §A.7).

SPEC ID 자기 검사:

```
$ ID="SPEC-INIT-TUX-I18N-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
PASS
```

제품 코드 기준선 동일성(v0.2.1 착수 HEAD 에서 측정. 3회차 트리의 제품 코드 차이는 `research.md` §0):

```
$ git diff --stat e7a7d4bb3 HEAD -- internal cmd pkg go.mod go.sum; echo "diff-exit=$?"
diff-exit=0
```

(diff 출력 없음.)

## §E.2 Run-phase Evidence

### Phase A (absorption-gate front, `internal/cli/wizard` only)

Scope of this phase: M1 opening checks V-a~V-d, M1 wizard-side work, M3 wizard-side scaffolding. The `internal/cli` package compile/test slot was not granted, so no command in this phase compiled package `internal/cli`. Everything below that needs `internal/cli` (`schemaSelectOptions` shape change, M2, the downgrade-confirm pre-fix RED capture, the `internal/cli` part of the plan.md §C 2 baseline) is Phase B.

#### Baseline (plan.md §C 1)

```
$ git rev-parse HEAD
500a73d444a28c750c7ab085813c710a5bb96686
$ git branch --show-current
WT-init-tux-i18n
```

`BASELINE_SHA=500a73d444a28c750c7ab085813c710a5bb96686` (local develop `f1f034bb4` absorbed; t583 not yet absorbed — the post-absorption re-capture is plan.md §C 1's second capture).

#### plan.md §C 2 — baseline selection (wizard package part only)

```
$ go test ./internal/cli/wizard/... -run 'Profile|Wizard|HuhTheme|UpdateVersion|TUI' -count=1 -v -timeout 300s
exit=0 · `=== RUN` lines 20 (18 top-level + 2 subtests) · PASS lines 20 · FAIL/SKIP 0
```

Full output: `.moai/reports/t586/phase-a/baseline-wizard-tests.txt`. The `./internal/cli/` half of this command is a Gap until the slot is granted.

#### plan.md §C 3 — RED ledger L1~L4 re-run (this tree)

All four exit 0 with output identical to `acceptance.md` §D.3 (L1: 5 files `huh_theme.go`, `init.go`, `profile_setup.go`, `update.go`, `update_version.go`; L2: `init.go:651`, `update.go:179`; L3: `translations.go:18,19,542,543,549,550,556,557,563,564` + `wizard_test.go:283,284,289,290,298,1256,1257,1259,1260`; L4: `init.go:659`, `update.go:187`). Verbatim: `.moai/reports/t586/phase-a/red-ledger-l1-l4.txt`.

#### plan.md §C 4 — pty preconditions

```
$ command -v timeout gtimeout tmux; tmux -V
/opt/homebrew/bin/timeout
/opt/homebrew/bin/gtimeout
/opt/homebrew/bin/tmux
tmux 3.6a
```

#### M1 opening checks V-a~V-d

| Item | Closed | Command | Observed | Evidence |
|---|---|---|---|---|
| V-a | recorded (judgement at gate §C 6) | `grep -n 'var uiStrings' internal/cli/wizard/translations.go` | `540:var uiStrings = map[string]UIStrings{` exit 0 | `.moai/reports/t586/phase-a/va-uistrings.txt` |
| V-b | **true** | `go test ./internal/cli/wizard/ -run 'TestProbeT586V[bc]' -count=1 -v` (throwaway probe) | ko key map changes the confirm help line: control `←/→ toggle • enter submit • y 예 • n 아니오` → with key map `←/→ 전환 • enter 제출 • y 예 • n 아니오`, frame 1 and frame 2 identical. The `y`/`n` entries read the button labels (`ConfirmYes`/`ConfirmNo`); key-map-only strings set on `Accept`/`Reject` (`KM-ACCEPT`/`KM-REJECT`) never render — huh v2 `Confirm.View` resets those two helps from `Affirmative`/`Negative` (`field_confirm.go:276`, `:282`). So the `design.md` §7 last row holds by construction | `.moai/reports/t586/phase-a/vb-vc-probe.txt`, probe source `probe-src-vb-vc.go.txt` |
| V-c | **true** | same run | note `TitleFunc` bound to a non-`*WizardResult` struct, visibility as a closure: before `● ○ 1 / 2`, after answering the select (`res.First="yes"`, which reveals a conditional id) the same group re-renders `● ○ ○ 1 / 3` | same files |
| V-d | **true** | `MOAI_T586_VD=1 go test ./internal/cli/wizard/ -run '^TestProbeT586Vd(Parent\|Child)$' -count=1 -v -timeout 300s` (throwaway probe) | parent built the child with `go test -c` into `t.TempDir()`, opened sentinel `moai-ptycap-sentinel-702d48a1`, ran the child under `MOAI_PTY_CAPTURE_SELFTEST=fail` with a 60 s context and `-test.timeout 30s`: child `--- FAIL: TestProbeT586VdChild`, `exit status 1`, child session `moai-ptycap-vdchild-b33421a3` absent afterwards, sentinel alive, before/after `moai-ptycap-` sets equal. `moai-ptycap-` sessions on the host before and after the run: 0 (`grep` exit 1 both times) | `.moai/reports/t586/phase-a/vd-probe.txt`, probe source `probe-src-vd.go.txt` |

The throwaway probe files were deleted after export and are not committed. No V item closed false, so no dependent AC is held.

#### M1 wizard-side work (profile question set, 4-locale form strings)

New files only, not wired: `internal/cli/wizard/profile_questions.go` (`ProfileOptions`, `ProfileResult`, `ProfileQuestionIDs`, `ProfileQuestions`), `internal/cli/wizard/profile_translations.go` (`profileQuestionTexts`, `LocalizeProfileQuestion`), tests `profile_questions_test.go`, `profile_translations_test.go`. The form strings are the v1 profile strings (`profileSetupText` in `internal/cli/profile_setup_translations.go`, en/ko/ja/zh) copied unchanged; the `profileSetupText` originals are untouched.

| Step | Command | Observed | Evidence |
|---|---|---|---|
| RED (own commit `ab0848d05`) | `go test ./internal/cli/wizard/ -run 'TestProfileQuestions_\|TestProfileTranslations_\|TestLocalizeProfileQuestion' -count=1 -v` against empty stubs | exit 1 · `=== RUN` 10 · FAIL 10 (e.g. `ProfileQuestions returned no questions`, `profile translation table carries 0 locales, want 4`) | `.moai/reports/t586/phase-a/m1-red.txt` |
| GREEN | same command | exit 0 · `=== RUN` 10 · PASS 10 | `.moai/reports/t586/phase-a/m1-green.txt` |
| Mutants | drop the ja `doc_lang` entry; wire `git_commit_lang` to `ProfileOptions.Model` | `--- FAIL: TestProfileTranslations_FourLocaleKeyParity` (`locale "ja" key set`); `--- FAIL: TestProfileQuestions_TypesAndOptionsPassThrough` (`git_commit_lang: options`); both restored | `.moai/reports/t586/phase-a/m1-mutants.txt` |

What these tests cover: question id set and order equal the `design.md` §2.2 ten ids; group labels and `conversation_language` alone on the first page; no conditional question (N = 10); select/input types; option lists passed through by value (copies, not shared slices); initial values become defaults; 4-locale key parity with non-empty title/description; ko/ja/zh actually translated; English question text sourced from the table; `LocalizeProfileQuestion` changes only title/description.

Pending for Phase B (needs `internal/cli`): the `schemaSelectOptions` `{Label, Value}` shape change and its three test files, and building `ProfileOptions` from the settings schema in `cli`. The wizard side needs no new option type — the existing `wizard.Option{Label, Value, Desc}` is the argument type (`design.md` §3). AC-ITI-005 (4) "option value sets equal the schema" is therefore judged in Phase B, where `cli` builds the lists; Phase A only proves the lists pass through unchanged.

Found while wiring (for M5/M7, not acted on): four profile ids (`conversation_language`, `user_name`, `model_policy`, `development_mode`) also exist in the init translation table (`translations.go:33,38,46,99` for ko, repeated for ja/zh). The shared `buildSelectField`/`buildInputField` localize through `GetLocalizedQuestion`, which is keyed by question id alone, so a profile form routed through them unchanged would render the init wording for those four ids. That is why the profile strings sit in their own table; M5 has to give the profile form a localizer that reads `profileQuestionTexts` (a `wizard.go` edit, after the gate).

#### M3 wizard-side scaffolding (pty harness, HOME watch list, render helpers)

New test files: `ptycap_harness_test.go` (gate, single name function, per-variable `-e` scrub, exact-name cleanup, anchor wait, key send, export, effective-env record and check, child build, subprocess runner), `ptycap_homewatch_test.go` (W1–W6 constant with producing function per row, snapshot, diff, real-HOME check), `ptycap_render_test.go` (`stripANSI`, `displayColumn` via go-runewidth with a fixed East-Asian condition, `compareGolden`, `requireLines`), `ptycap_selfcheck_test.go` (the self-checks and the two child helpers).

| Step | Command | Observed | Evidence |
|---|---|---|---|
| RED (own commit `8b51536a9`) | `go test ./internal/cli/wizard/ -run 'Ptycap\|PtyCapture\|HomeWatch' -count=1 -v` without and with `MOAI_PTY_CAPTURE=1`, stubbed harness | exit 1 both · `=== RUN` 29 · FAIL 15, SKIP 2 (child helpers), PASS 1 (`TestPtycapTopLevelResults` — tests a parser defined in the same file) | `.moai/reports/t586/phase-a/m3-red-ungated.txt`, `m3-red-gated.txt` |
| First gated GREEN attempt | same, `MOAI_PTY_CAPTURE=1` | `TestPtyCapture_NormalRun` FAIL: `child effective environment: [TERM "screen-256color", want "xterm-256color"]` — tmux sets a pane's `TERM` from `default-terminal`, overriding `-e TERM=…` (`man tmux`: "default-terminal … the default value of the TERM environment variable"). The observation layer caught what the `-e` reasoning assumed. Fix: the child command is `TERM=xterm-256color exec <bin> …`; `-e TERM` stays but is not relied on | `.moai/reports/t586/phase-a/m3-gated-run1-term-finding.txt` |
| GREEN, ungated | `go test ./internal/cli/wizard/... -count=1 -cover -v -timeout 300s` | exit 0 · `=== RUN` 231 · PASS 224 · FAIL 0 · SKIP 7 (the 7 capture tests) · `coverage: 92.9% of statements` | `.moai/reports/t586/phase-a/wizard-full-ungated.txt` |
| GREEN, gated | `MOAI_PTY_CAPTURE=1 MOAI_PTY_CAPTURE_OUT=<abs>/.moai/reports/t586/phase-a/pty go test ./internal/cli/wizard/ -run 'Ptycap\|PtyCapture\|HomeWatch' -count=1 -v -timeout 600s` | exit 0 · `=== RUN` 29 · PASS 15 · SKIP 2 (child helpers, which run only as children) · FAIL 0 | `.moai/reports/t586/phase-a/m3-green-gated.txt`, captures under `.moai/reports/t586/phase-a/pty/` |
| Mutant H-a | remove the `t.Cleanup(ptycapKill)` in `ptycapStart`, run `TestPtyCapture_ForcedFailure` | `session moai-ptycap-TestPtyCaptureSelfTestChild-cbf3f197 opened by the failing child survived` · `--- FAIL`; the leaked session was then killed by that exact name; harness restored | `.moai/reports/t586/phase-a/m3-mutant-ha-no-cleanup.txt` |
| Mutant H-c | `WaitFor` returns the last capture on timeout instead of failing, run `TestPtyCapture_ForcedTimeout` | `self-test child output lacks "not visible within"` · `--- FAIL`; harness restored | `.moai/reports/t586/phase-a/m3-mutant-hc-waitfor-returns.txt` |
| Lint | `go vet ./internal/cli/wizard/...`; `golangci-lint run ./internal/cli/wizard/...` | vet exit 0; `0 issues.` (first run flagged an unused golden wrapper + flag; removed — the `testdata/golden` wrapper and its update flag land with the first committed golden) | `.moai/reports/t586/phase-a/golangci-wizard.txt` |

AC-ITI-019 / AC-ITI-020 self-check status (harness level, this tree):

| Clause | Test | Observed |
|---|---|---|
| 019 (a) ungated → all capture tests SKIP; `moai-ptycap-` set unchanged, sentinel present first | `TestPtyCapture_SkipWithoutGate` | child `-test.run ^TestPtyCapture`: 7 top-level results, all `--- SKIP`; sets equal | 
| 019 (b) gate on, no tmux on PATH → FAIL, no PASS/SKIP | `TestPtyCapture_FailWithoutTmux` | child `-test.run ^TestPtyCapture_` with `PATH=<empty dir>`: 5 of 5 `--- FAIL` with `MOAI_PTY_CAPTURE=1 but tmux is not on PATH` |
| 020 (1) normal run: capture file carries the anchor; only the sentinel remains; effective env four checks | `TestPtyCapture_NormalRun` | capture `pty/normal-run-init-first-page.txt` has `Select conversation language` + the four language option lines; `before=[moai-ptycap-sentinel-…] after=[same]`; `child effective environment verified (14 vars)`; child cwd in `/var/folders/…`, repository `/Users/goos/…/t586` |
| 020 (2) forced failure | `TestPtyCapture_ForcedFailure` | child `--- FAIL: TestPtyCaptureSelfTestChild`, opened session absent afterwards, sentinel alive, sets equal |
| 020 (3) forced timeout | `TestPtyCapture_ForcedTimeout` | child FAIL with `anchor "ZZ-PTYCAP-ANCHOR-NEVER-RENDERED" not visible within 2s`, session absent afterwards |
| 020 (4) watch-list positive control on a fake HOME | `TestHomeWatch_PositiveControl` | (i) PASS `changed=[]`; (ii) FAIL names `.claude/settings.json`; (iii) FAIL names `.moai/db/<key>`; (iv) FAIL names `.moai/claude-profiles/p2/preferences.yaml` and the glob; (v) PASS `changed=[]` (runs in every `go test`, never touches the real HOME) |
| 020 "And" — counts printed, real-HOME comparison PASS | the three pty runs | each prints `real HOME watch list: 17 entries, 11 present on the real HOME` before and after, no change reported |

Host tmux state: `moai-ptycap-` sessions before the gated runs 0, after 0 (`tmux list-sessions -F '#{session_name}' | grep -c '^moai-ptycap-'`). No other session was read or touched; nothing ran `kill-server`.

Harness contract notes for Phase B:
- Captures are exported only to `MOAI_PTY_CAPTURE_OUT` (absolute path) when set; otherwise to the test's temp dir. The harness never writes into the repository on its own, so the P9 export is an explicit choice of the verdict run's environment.
- Location decided by the lead (see "Harness move (lead decision)" below): the harness is the test-support package `internal/cli/ptycaptest`, one copy. An `internal/cli` test imports it directly. The AC-ITI-003 child is a `TestPtyCaptureChild` (the name in `ptycaptest.ChildTestName`) in package `cli`, built with `ptycaptest.BuildChild(t, ".")` — `go test` runs a test binary in its package directory, so `"."` is `./internal/cli`, which is what `design.md` §11 builds.
- D4 observation only (not a verdict, pre-t583 tree): in the 80×30 capture of the current init first page the stepper line is absent — the first captured line is `┃ Select conversation language` and the capture holds 0 `●`/`○` characters. Recorded for the §G deferred item.

Plan-audit iter4 notes: O1 (S2 anchor fooled by a trailing comment) and O2 (design.md §10 wording) concern the S2 guard retarget in M5 inside `internal/cli`; neither applies to Phase A. O2 is a `design.md` body edit, which this agent does not own.

#### Harness move (lead decision)

Decision, as given by the lead: "Move the harness into a small importable test-support package — precedent internal/hook/testutil. Condition: add one check (go list -deps or a test) proving no production (non-_test.go) code imports this package. Do not create two copies."

Package: `internal/cli/ptycaptest` (package `ptycaptest`), path as suggested. It imports `internal/config`, `internal/homestate`, and `go-runewidth` — neither `internal/cli` nor `internal/cli/wizard` (`go list -deps ./internal/cli/ptycaptest/ | grep moai-adk/internal/cli` lists only the package itself). Package doc follows the `internal/hook/testutil` style (test-only, production code MUST NOT import it) and names the guard.

What moved (one copy; the four `internal/cli/wizard/ptycap_{harness,homewatch,render,selfcheck}_test.go` files are deleted):

| File | Exported API |
|---|---|
| `harness.go` | gate constants once (`GateEnv`, `ChildEnv`, `SelfTestEnv`, `CaseRootEnv`, `OutDirEnv`, `CanaryEnv`, `EnvOutEnv`), `ChildTestName`, `SessionPrefix`, `Width`/`Height`/`Term`, `AnchorTimeout`; `Gate`, `SessionName`, `ListSessions`, `OpenSentinel`, `Case`/`NewCase`/`(*Case).ChildEnv`, `RecordEnv`, `VerifyChildEnv`, `BuildChild(tb, pkg)`, `Start`, `(*Session).Capture/WaitFor/SendKeys/Close`, `Export`, `SessionsNamedIn`, `SubprocessEnv`, `RunChild`, `AssertSkipWithoutGate`, `AssertFailWithoutTmux` |
| `homewatch.go` | W1–W6 list (unexported, single constant), `SnapshotHome(root, keys)`, `DiffSnapshots`, `HomeSnapshot.Entries/Existing`, `WatchRealHome` |
| `render.go` | `StripANSI`, `DisplayColumn` (pinned `runewidth.Condition{EastAsianWidth: false, StrictEmojiNeutral: true}`), `CompareGolden`, `RequireLines` |

All signatures take `testing.TB`. `BuildChild` takes the package argument for `go test -c` and resolves it against the working directory; callers pass `"."`, which is the calling package because `go test` runs a test binary in its package directory. The only behavioral edits are mechanical: `*testing.T` → `testing.TB`, the child test name read from `ChildTestName`, `containsString` → `slices.Contains`, the child binary named `child.test`, and the two AC-ITI-019 test bodies lifted into the shared `Assert…` functions with the capture-test count as a parameter.

Self-check placement:

| Clause | Where | Why there |
|---|---|---|
| 020 (1) normal run | `ptycaptest.TestPtyCapture_NormalRun` on a fixture child (prints `PTYCAP FIXTURE READY` + the four language lines, blocks on stdin) **and** `wizard.TestPtyCapture_NormalRun` on the product screen | the package run proves the harness without product coupling; the wizard run proves the product-screen path still works through the moved harness |
| 020 (2) forced failure, (3) forced timeout | `ptycaptest` only (`TestPtyCaptureSelfTestChild` opens a fixture-child session) | they test harness cleanup and the timeout path, which no product screen changes |
| 020 (4) watch-list positive control | `ptycaptest.TestHomeWatch_PositiveControl` (pure, fake HOME) | the list and the comparison function live in the package |
| 019 (a) ungated → all SKIP, (b) no tmux → all FAIL | both packages, bodies shared via `AssertSkipWithoutGate` / `AssertFailWithoutTmux` | the clause is about "all capture tests"; each package asserts over its own set (ptycaptest 7 / 5, wizard 4 / 3) |

Moved tests keep their names; only the package changes. `wizard` keeps `TestPtyCaptureChild` (product child), `TestPtyCapture_NormalRun`, `TestPtyCapture_SkipWithoutGate`, `TestPtyCapture_FailWithoutTmux` in `ptycap_test.go`. Full old → new mapping: `.moai/reports/t586/phase-a/move-before-after-summary.txt`.

Import guard: `ptycaptest.TestNoProductionImport` runs `go list -f '<P line: .Imports> <T line: .TestImports .XTestImports>' ./...` from the module root (`go list -m -f {{.Dir}}`), no compilation. Positive existence first (142 packages scanned, self and wizard listed), positive control (the same scan sees the wizard `_test.go` import of ptycaptest), then the verdict (no `P` line lists ptycaptest). `TestCheckNoProductionImport_Synthetic` pins the matcher (test-only import allowed, production import reported, prefix path not a match).

| Step | Command | Observed | Evidence |
|---|---|---|---|
| Before, ungated (own commit `4061b6bda`, tree c7646c48a) | `go test ./internal/cli/wizard/... -count=1 -cover -v -timeout 300s` | exit 0 · RUN 231 · PASS 224 · FAIL 0 · SKIP 7 · cov 92.9% | `.moai/reports/t586/phase-a/move-before-ungated.txt` |
| Before, gated | `MOAI_PTY_CAPTURE=1 go test ./internal/cli/wizard/ -run 'Ptycap\|PtyCapture\|HomeWatch' -count=1 -v -timeout 600s` | exit 0 · RUN 29 · top-level PASS 15 · SKIP 2 · FAIL 0 | `move-before-gated.txt` |
| Guard mutant | scratch production file `internal/cli/ptycaptest/zzmutant/mutant.go` importing ptycaptest; `go test ./internal/cli/ptycaptest/ -run '^TestNoProductionImport$' -count=1 -v` | exit 1: `production (non-_test.go) code imports the test-only package …: [github.com/modu-ai/moai-adk/internal/cli/ptycaptest/zzmutant]` (143 packages scanned); scratch removed → exit 0, 142 packages | `move-guard-mutant.txt` |
| Harness mutants on the moved code | H-a: drop `tb.Cleanup(killSession)` in `Start`; H-c: `WaitFor` returns on timeout | H-a `--- FAIL: TestPtyCapture_ForcedFailure` (`session moai-ptycap-TestPtyCaptureSelfTestChild-61e7d4aa … survived`; killed by that exact name); H-c `--- FAIL: TestPtyCapture_ForcedTimeout` (`lacks "not visible within"`); both restored | `move-after-harness-mutants.txt` |
| After, wizard ungated (tree = commit `18e97065b`, `git diff HEAD -- internal/` empty at measurement) | same as before | exit 0 · RUN 206 · PASS 202 · FAIL 0 · SKIP 4 · cov 92.9% (231 − 29 moved + 4 kept) | `move-after-ungated.txt` |
| After, wizard gated | same as before | exit 0 · 4 top-level: NormalRun / SkipWithoutGate / FailWithoutTmux PASS, child SKIP · env verified (14 vars) · sentinel-only sets equal · `real HOME watch list: 17 entries, 11 present` before and after, no change | `move-after-gated.txt` |
| After, ptycaptest ungated | `go test ./internal/cli/ptycaptest/... -count=1 -cover -v -timeout 300s` | exit 0 · RUN 31 · top-level PASS 12 · SKIP 7 · FAIL 0 · cov 38.3% | `move-after-pkg-ungated.txt` |
| After, ptycaptest gated | same with `MOAI_PTY_CAPTURE=1`, `-timeout 600s` | exit 0 · top-level PASS 17 · SKIP 2 (child helpers) · FAIL 0 · cov 82.2%; all 17 pre-move top-level names present, plus the 2 guard tests | `move-after-pkg-gated.txt` |
| Lint / vet | `go vet ./internal/cli/wizard/... ./internal/cli/ptycaptest/...`; `golangci-lint run` (same); `GOOS=windows GOARCH=amd64 go vet ./internal/cli/ptycaptest/ ./internal/cli/wizard/` | exit 0 · `0 issues.` · exit 0 | `move-after-vet.txt`, `move-after-golangci.txt`, `move-after-vet-windows.txt` |
| Dependency check | `go list -deps -test ./internal/cli/wizard/ \| grep -c 'moai-adk/internal/cli$'`; control `grep -c 'internal/cli/ptycaptest$'` | `0`; control `1` | `move-after-deps.txt` |

Host tmux: `moai-ptycap-` sessions 0 before the before-runs, 0 before and after the after-runs ("no server running" counted as 0). Summary: `move-before-after-summary.txt`.

Coverage note: ptycaptest is 38.3% ungated and 82.2% gated, below the 85% package target. The ungated gap is the tmux-driven code, which runs only under the gate; the gated remainder is error branches (tmux `list-sessions`/`new-session` failure, export directory creation failure, the Windows skip). Recorded as residual, not raised here.

#### Lead dispositions

- (a) `acceptance.md` §B P4 TERM row is a plan-side correction: tmux's `default-terminal` overrides `-e TERM`, and the harness sets `TERM` in the child command (`TERM=xterm-256color exec <bin> …`). Pending — applied at sync, not edited now (this agent does not edit `acceptance.md`).
- (b) Real-HOME W6: 58 new `~/.moai/run/<key>/home-state-migration.json.lock` directories were created 19:26:13–19:29:44 on 2026-09-11. Not reachable from the wizard package (0 `homestate` deps at that tree); the writer is unidentified → Gap, accepted by the lead as unrelated to this card. W1–W5 before/after identical; W4 absent both times.
- (c) Process incident: after `/clear` a second manager-develop was spawned while the previous Phase A agent was still running. The second agent detected the HEAD move and wrote nothing, and self-reported to the lead.

### Stage B (internal/cli slot)

Slot granted by the lead (confirmed "유지"). Opening state: HEAD `0e3718b7d`, branch `WT-init-tux-i18n`, no `.moai/reports/t586/stage-b/` directory (a previous attempt died on a rate limit before running anything). `plan.md` §F places M2 under "흡수 게이트 앞 — t583 과 겹치지 않음", i.e. before the absorption gate, so the batch order holds. Host tmux before the batch: `no server running on /private/tmp/tmux-501/default` (0 `moai-ptycap-` sessions).

#### Step 1 — plan.md §C 2, `internal/cli` half — STOPPED (baseline FAIL)

| Command | Exit | Observed | Evidence |
|---|---|---|---|
| `go test ./internal/cli/ -run 'Profile\|Wizard\|HuhTheme\|UpdateVersion\|TUI' -count=1 -list '.*' -timeout 600s` (as dispatched) | 0 | 3602 names — `-list` takes its own regexp and ignores `-run`, so this counts every test/benchmark in the package, not the selection | `.moai/reports/t586/stage-b/baseline-cli-list.txt` |
| `go test ./internal/cli/ -count=1 -list 'Profile\|Wizard\|HuhTheme\|UpdateVersion\|TUI' -timeout 600s` (selection count) | 0 | 129 top-level `Test…` names (non-zero) | `.moai/reports/t586/stage-b/baseline-cli-list-selection.txt` |
| `go test ./internal/cli/ -run 'Profile\|Wizard\|HuhTheme\|UpdateVersion\|TUI' -count=1 -v -timeout 600s` | **1** | `=== RUN` 217 · `--- PASS` 216 · `--- FAIL` 1 · `--- SKIP` 0. The failure: `--- FAIL: TestHomeStateChangedSurfaceCoverageConsumesFreshProfile (6.04s)` with `home_state_coverage_test.go:420: audited production file changed after coverage tip: internal/cli/launcher.go` | `.moai/reports/t586/stage-b/baseline-cli-tests.txt` |

Attribution (read-only, no fix attempted — the dispatch says stop on any baseline FAIL):
- The test is selected only because its name contains `Profile`; it reads git history (`resolveHomeStateCoverageChangeSet`, `internal/cli/home_state_coverage.go:196-300`) and fails when a file audited by the t592 home-state coverage marker commits has a different blob at HEAD.
- `internal/cli/launcher.go` was last changed by `6010d5d82 feat(launcher): support pinning the Claude Code binary (#1697)` (2026-09-10), after the t592 marker `0c86e61d0` (`git merge-base --is-ancestor 0c86e61d0 6010d5d82` → 0). `6010d5d82` is an ancestor of this branch's base `500a73d44` (exit 0), and this branch changed none of `launcher.go`, `home_state_coverage.go`, `home_state_coverage_test.go` since the base (`git diff --stat 500a73d44 HEAD -- …` empty, exit 0). The failure is inherited from local develop, not produced by this card.
- Consequence: steps 2–4 were not started. No product or test file was changed in Stage B.

#### Develop absorption (lead disposition (b)) and baseline re-capture

Disposition: option (b), carried out by the orchestrator. Local develop `eb50af5a8ec51862e3df76c2e3b08377ce01c4d8` — which carries the t600 `5b7927b15` and t606 `92494400f` fixes to `internal/cli/home_state_coverage{,_test}.go`, the file pair behind the step-1 FAIL — was merged into `WT-init-tux-i18n` as `cd7dc491c66eedd490a5422e643a579486746b5f` (parents `7b63ea61d` and `eb50af5a8`, clean merge). Checked in this run: `git diff --stat 7b63ea61d cd7dc491c -- internal/cli/wizard internal/cli/init.go internal/cli/update_version.go internal/cli/profile_setup.go internal/cli/update.go` prints nothing (exit 0); the merge brings 63 files, none of them in the t583-overlap set or this card's M1/M2 targets.

This is a **non-t583 absorption**. plan.md §C 1 capture at the absorption commit:

```
$ git rev-parse HEAD
cd7dc491c66eedd490a5422e643a579486746b5f
$ git branch --show-current
WT-init-tux-i18n
```

`BASELINE_SHA=cd7dc491c66eedd490a5422e643a579486746b5f` for Stage B. The post-t583 §C 1 capture (the absorption-gate one) is still pending.

#### Step 1 re-run on `cd7dc491c` — GREEN

| Command | Exit | Observed | Evidence |
|---|---|---|---|
| `go test ./internal/cli/ -count=1 -list 'Profile\|Wizard\|HuhTheme\|UpdateVersion\|TUI' -timeout 600s` | 0 | 129 top-level `Test…` names | `.moai/reports/t586/stage-b/baseline2-cli-list-selection.txt` |
| `go test ./internal/cli/ -run 'Profile\|Wizard\|HuhTheme\|UpdateVersion\|TUI' -count=1 -v -timeout 600s` | 0 | `=== RUN` 217 (129 top-level, matching the list count) · `--- PASS` 217 · FAIL 0 · SKIP 0; `--- PASS: TestHomeStateChangedSurfaceCoverageConsumesFreshProfile (4.30s)` | `.moai/reports/t586/stage-b/baseline2-cli-tests.txt` |

The failing run's evidence (`baseline-cli-*.txt`) is kept alongside, not overwritten. Selection count note stands: use `-list '<pattern>'`; `-run … -list '.*'` lists the whole package.

#### Step 2 — M3 cli part: downgrade confirm pre-fix pty RED (before M2)

New test file only: `internal/cli/ptycap_child_test.go`. One child dispatcher `TestPtyCaptureChild` (switch on `ptycaptest.ChildEnv`; later cases such as the AC-ITI-003 init case add a `case`), which records the effective environment, then blocks the network seams without editing their file (`deferredUpdateEnabled`/`deferredUpdateCheck` stubbed, `versionInstallHTTPClient` a transport that refuses every request, `MOAI_UPDATE_URL` set empty), and runs `runVersionBranch` with `version.Version = "v9.9.9"`, tag `v1.0.0`, `--binary`, and `versionInstallBinaryPath` under the case temp dir. The case directory carries `.moai/config/sections/language.yaml` with `conversation_language: ko`. Capture tests: `TestPtyCapture_DowngradeConfirmLocalized` (ko title/description/buttons/help labels, no English strings; also AC-ITI-015 (a) reachability) and `TestPtyCapture_DowngradeConfirmButtonAlignment` (AC-ITI-015 (a): first button label column = description first column), each with the positive-existence gate on title/description/button lines; plus `TestPtyCapture_SkipWithoutGate` / `TestPtyCapture_FailWithoutTmux` calling the shared `AssertSkipWithoutGate` (5 results) / `AssertFailWithoutTmux` (4 results) over cli's capture tests (AC-ITI-019 "all capture tests").

| Step | Command | Observed | Evidence |
|---|---|---|---|
| Ungated | `go test ./internal/cli/ -run '^TestPtyCapture' -count=1 -v -timeout 600s` | exit 0 · `=== RUN` 5 · `--- SKIP` 5 | `.moai/reports/t586/stage-b/step2-red-ungated.txt` |
| Gated pre-fix RED | `MOAI_PTY_CAPTURE=1 MOAI_PTY_CAPTURE_OUT=<abs>/.moai/reports/t586/stage-b/pty-prefix go test ./internal/cli/ -run '^TestPtyCapture' -count=1 -v -timeout 600s` | exit 1 · `=== RUN` 5 · PASS 2 (SkipWithoutGate, FailWithoutTmux) · FAIL 2 · SKIP 1 (child). Localized FAIL for the documented reason — v1 English render in a ko project: `title line "┃ Downgrade v9.9.9 → v1.0.0?" lacks the ko title "다운그레이드할까요? v9.9.9 → v1.0.0"`, description not ko, `button line "┃                     Yes     No" lacks the ko labels 예 / 아니오`, help lacks `전환`/`제출`, frame carries `toggle`, `submit`, `Downgrade`, `older than the running version`. Alignment FAIL: `first button label "Yes" at column 22, description at column 2` — the documented v1 22-column indent. Child env verified (14 vars) in both sessions; real HOME watch list `17 entries, 11 present` before and after, no change | `.moai/reports/t586/stage-b/step2-red-gated.txt`; captures `pty-prefix/downgrade-confirm-localized.txt`, `pty-prefix/downgrade-confirm-alignment.txt`; AC-ITI-019 child outputs `pty-prefix/ac019a-ungated-child-output.txt`, `pty-prefix/ac019b-no-tmux-child-output.txt` |

Host tmux: `no server running` before and after the gated run (0 `moai-ptycap-` sessions).

Divergence — pre-fix **golden** RED: the v1 confirm is built inline inside `runVersionBranch` (`update_version.go:327-331`) and calls `form.Run()` directly; no builder returns the form, so its `View()` cannot be drawn from a test without a product edit, and a test-only copy of those lines would measure a copy, not the product. The pre-fix evidence for this surface is therefore the pty capture above (its ANSI-free screen is the exported file). The golden RED for the confirm surface is AC-ITI-012's four-case golden, established in step 4 as its own commit against a stub helper, before the M2 fix.

#### Step 3 — M1 cli part (option shape, ProfileOptions, AC-ITI-005 (4))

Changes: `schemaSelectOptions` (`internal/cli/profile_setup.go`) now returns `[]wizard.Option` (`{Label, Value}`, the existing wizard type — design.md §3); the four v1 call sites in the same file wrap it in the new `huhV1Options` adapter, so the v1 form renders the same options until M5 retires it (no other `profile_setup.go` logic touched). New `internal/cli/profile_options.go`: `buildProfileOptions(t profileSetupText) wizard.ProfileOptions` — language values from `settings.FieldOptionDefs("conversation_lang")` with the init wizard's native-name labels and descriptions, `model`/`effort_level`/`development_mode` from `schemaSelectOptions(…, true)`, `permission_mode` from `schemaSelectOptions(…, false)`, `model_policy` = `settings.EmptyLabelFor("model_policy")` empty option + `template.ValidModelPolicies()` labelled from `t.ModelPolicy{High,Medium,Low}`. Not wired into the flow (M5). Test-file `.Key` → `.Label`: 7 lines in 2 files (`profile_setup_schema_options_test.go` 72, 81, 82, 109, 112; `profile_setup_nested_test.go` 112, 113); the third file named by plan.md, `profile_setup_projectconfig_test.go`, reads only `.Value` and needed no edit — plan.md says "3 files, 6 sites", measured 2 files, 7 lines.

| Step | Command | Observed | Evidence |
|---|---|---|---|
| RED (own commit `7c00ff9ba`) | `go test ./internal/cli/ -run '^TestProfileOptions_' -count=1 -v -timeout 600s` against an empty `buildProfileOptions` stub | exit 1 · `=== RUN` 2 · FAIL 2 (`locale en: select "model" has no options`, … every select, every locale) | `.moai/reports/t586/stage-b/m1-cli-red.txt` |
| GREEN | `go test ./internal/cli/ -run '^(TestProfileOptions_\|TestSchemaSelectOptions_\|TestModelPolicyLabels_\|TestTUINestedConfigNoParallelWriter\|TestPermissionModeNormalizeAcceptEdits\|TestTUIEmptyLabelsSchemaSourced\|TestPersistProjectConfig_\|TestProfileSetupConstructsProjectSelects)' -count=1 -v -timeout 600s` (new tests + every test in the three named files) | exit 0 · `=== RUN` 22 (15 top-level) · PASS 22 · FAIL 0 · SKIP 0 | `.moai/reports/t586/stage-b/m1-cli-green.txt` |
| Mutant A | drop the empty `model_policy` option | `--- FAIL: TestProfileOptions_ValueSetsEqualSchema` (`select "model_policy" values ["high" "medium" "low"], want the set ["" "high" "medium" "low"]`, 4 locales); restored | `m1-cli-mutant-a-no-empty-policy.txt` |
| Mutant B | wire `Model` to the `effort_level` list | `--- FAIL: TestProfileOptions_ValueSetsEqualSchema` (`select "model" values ["" "low" …], want the set ["" "fable[1m]" "opus[1m]" "sonnet[1m]" "haiku"]`) and `--- FAIL: TestProfileOptions_Labels`; restored | `m1-cli-mutant-b-model-wired-to-effort.txt` |

What AC-ITI-005 (4) now covers: per locale, the nine select questions of `wizard.ProfileQuestions(buildProfileOptions(t), …)` carry exactly the schema value sets (empty value where the wizard offers one; none for `permission_mode` and the four language fields), with no repeated value, after a positive-existence check that the schema supplies every list. Mutant A also shows a limit of the label test: `settings.EmptyLabelFor("model_policy")` is `""` (the pre-existing blank-label defect named in `profile_setup_nested_test.go:121-125`), so a missing empty option is caught by the value test, not the label test.

#### Step 4 — M2 downgrade confirm v2 move and language resolution

Changes: new `internal/cli/wizard/downgrade_confirm.go` — `NewDowngradeConfirmForm(locale, current, target string, value *bool) *huh.Form` (design.md §6 arguments), title/description from a four-locale `downgradeConfirmTexts` table (moves into `translations.go` in M7), buttons from `ConfirmYes`/`ConfirmNo`, theme `newMoAIWizardTheme()`, help from `localizedKeyMap(locale)`. New `internal/cli/wizard/help_keymap.go` — `helpActionLabels` (the design.md §7 table, en identity + ko/ja/zh) and `localizedKeyMap`, which relabels the help action of every default binding whose English action is in the table; the confirm `y`/`n` entries need no row because huh v2 rebuilds them from the button labels (V-b). New `internal/cli/downgrade_locale.go` — `resolveDowngradeLocale(cwd)`: project `language.yaml` (only when `cwd/.moai` is a directory) → `profile.ReadPreferences(profile.GetCurrentName()).ConversationLang` → `"en"`; only the resolved string reaches the wizard. `internal/cli/update_version.go`: the inline v1 confirm (former lines 326-331) is replaced by the helper with `resolveDowngradeLocale(os.Getwd())`; the `downgrade confirmation: %w` wrapping and the `Downgrade aborted` pill are unchanged; the file no longer imports `github.com/charmbracelet/huh`. None of the t583-overlap files, `init.go`, `update.go` was touched.

| Step | Command | Observed | Evidence |
|---|---|---|---|
| RED (own commit `980bc5881`), wizard | `go test ./internal/cli/wizard/ -run '^(TestHelpActionLabels_\|TestDowngradeConfirmTexts_\|TestNewDowngradeConfirmForm_)' -count=1 -v -timeout 600s` against stubs (empty tables, v1 English text on a v2 confirm) | exit 1 · RUN 3 · FAIL 3 (`help label table has no en entries`, `locale ko: no downgrade confirm text`, `ko confirm lacks "예"`) | `.moai/reports/t586/stage-b/m2-red-wizard.txt` |
| RED, cli (AC-ITI-012 four cases) | `go test ./internal/cli/ -run '^TestUpdateVersionDowngradeConfirm_' -count=1 -v -timeout 600s` against a stub resolving `en` | exit 1 · RUN 5 (1 + 4 subtests) · FAIL 5: `resolved locale "en", want "ja"` / `"ko"` / `"zh"`, title/description/buttons/help not rendered in the expected locale, `ja view still carries the English "Downgrade"`, and each golden missing (case (c) fails only on the missing golden — its English render is already the expected one) | `.moai/reports/t586/stage-b/m2-red-cli.txt` |
| Goldens written | same command with `-args -update-golden` | exit 0; four files `internal/cli/testdata/downgrade-confirm/{a-project-ja-profile-ko,b-no-project-profile-ko,c-no-project-no-profile,d-project-without-lang-profile-zh}.golden` (ANSI-stripped `View()` at 80×40) | `m2-golden-write.txt` |
| GREEN, wizard | the wizard RED command | exit 0 · PASS 3 | `m2-green-wizard.txt` |
| GREEN, cli — AC-ITI-012 + every test in `update_version_test.go` | `go test ./internal/cli/ -run '^(TestUpdateVersionDowngradeConfirm_\|TestRunVersionBranch_\|TestIsVersionDowngrade\|TestCompareVersionLoose\|TestInstallVersionTag_\|TestNormalizeVersionTag\|TestTagReleaseURL\|TestValidateUpdateVersionConflicts\|TestVersionFlagRegistered\|TestUpdateFlagsNoVersionDefaultIsNoop\|TestEnsureUpdate_DevBranchPreserved)' -count=1 -v -timeout 600s` (no update flag — goldens compared) | exit 0 · RUN 56 (20 top-level) · PASS 55 · FAIL 0 · SKIP 1 — the skip is the pre-existing `TestRunVersionBranch_NonTTYProceeds` (`update_version_test.go:645`, skips when stdin is a character device; the `go test` binary gets one here, also with `true \|` and `< /dev/null` — `m2-green-cli-runversionbranch-{pipe,devnull}.txt`) | `m2-green-cli.txt` |
| Mutant A — profile before project | swap the two lookups in `resolveDowngradeLocale` | `--- FAIL: …/a-project-ja-profile-ko` (`resolved locale "ko", want "ja"`); b, c, d PASS; restored | `m2-mutant-a-profile-first.txt` |
| Mutant B — English fallback removed | return `""` instead of `"en"` | `--- FAIL: …/c-no-project-no-profile` (`resolved locale "", want "en"`); restored | `m2-mutant-b-no-en-fallback.txt` |
| Mutant C — key map dropped | remove `.WithKeyMap(localizedKeyMap(locale))` | a, b, d FAIL (`toggle help "切替" not rendered`, `ja view still carries the English "toggle"`); c PASS; restored | `m2-mutant-c-no-keymap.txt` |
| Post-fix pty capture | step 2's gated command with `MOAI_PTY_CAPTURE_OUT=<abs>/.moai/reports/t586/stage-b/pty-postfix` | exit 1 · PASS 3 (`DowngradeConfirmLocalized`, `SkipWithoutGate`, `FailWithoutTmux`) · FAIL 1 · SKIP 1 (child). Localized render on the real TTY: `┃ 다운그레이드할까요? v9.9.9 → v1.0.0` / `┃ 요청한 태그가 지금 실행 중인 버전보다 오래되었습니다.` / `┃                    예     아니오` / `←/→ 전환 • enter 제출 • y 예 • n 아니오`. The FAIL is `DowngradeConfirmButtonAlignment` (`first button label "예" at column 21, description at column 2`) — AC-ITI-015 (a), whose fix is REQ-ITI-014 in M7; it stays red by design until then. Child env verified (14 vars) in both sessions; real HOME watch list `17 entries, 11 present` before and after, no change | `m2-postfix-gated.txt`, `pty-postfix/downgrade-confirm-localized.txt`, `pty-postfix/downgrade-confirm-alignment.txt` |

Host tmux: `no server running` before and after the post-fix run.

Doc divergence (measured, not edited): acceptance.md AC-ITI-015 gives the pre-fix v2 button indent as 7 columns. The v2 downgrade confirm measures 21 columns on the 80-column pty (`예` at column 21) and 20–22 in the 80×40 goldens depending on label width (huh centres the button row across the field width). The 7-column figure evidently belongs to another field width; AC-ITI-015's equality criterion is unaffected.

#### After the batch

Two coverage tests were added after M2 GREEN (`TestHuhV1Options_PreservesLabelAndValue` in cli, `TestNewDowngradeConfirmForm_UnknownLocaleRendersEnglish` in wizard) because the first coverage pass showed `huhV1Options` 0.0% (only the interactive v1 flow reached it) and the unknown-locale branch of `NewDowngradeConfirmForm` uncovered (71.4%).

| Check | Command | Exit | Observed | Evidence |
|---|---|---|---|---|
| vet | `go vet ./internal/cli/ ./internal/cli/wizard/... ./internal/cli/ptycaptest/...` | 0 | no output | `.moai/reports/t586/stage-b/after-vet.txt` |
| lint | `golangci-lint run ./internal/cli/ ./internal/cli/wizard/... ./internal/cli/ptycaptest/...` | 0 | `0 issues.` | `after-golangci.txt` |
| build | `go build -o <scratchpad>/moai ./cmd/moai/` (output redirected out of the tree; the binary was not run) | 0 | no output | `after-build.txt` |
| cross build | `GOOS=windows GOARCH=amd64 go build -o <scratchpad>/moai.exe ./cmd/moai/` | 0 | no output | `after-build-windows.txt` |
| wizard cover | `go test ./internal/cli/wizard/... -count=1 -cover -v -timeout 600s` | 0 | RUN 210 · PASS 206 · FAIL 0 · SKIP 4 (the four pty capture tests, ungated) · `coverage: 93.2% of statements`; `NewDowngradeConfirmForm`, `localizedKeyMap`, `keyMapBindings` 100.0% | `after-wizard-cover.txt`, `after-wizard-changed-func-cover.txt` |
| ptycaptest cover | `go test ./internal/cli/ptycaptest/... -count=1 -cover -v -timeout 600s` | 0 | RUN 31 · PASS 24 · FAIL 0 · SKIP 7 · `coverage: 38.3% of statements` (unchanged by Stage B; residual already recorded in Phase A) | `after-ptycaptest-cover.txt` |
| cli changed-path cover | `go test ./internal/cli/ -run '^(TestProfileOptions_\|TestHuhV1Options_\|TestSchemaSelectOptions_\|TestUpdateVersionDowngradeConfirm_\|TestTUIEmptyLabelsSchemaSourced)' -count=1 -v -timeout 600s -coverprofile=<scratchpad>` + `go tool cover -func` | 0 | PASS 9 top-level; `resolveDowngradeLocale` 100.0%, `buildProfileOptions` 100.0%, `schemaSelectOptions` 100.0%, `huhV1Options` 100.0%, `profileLanguageOptions` 88.9% (unknown-value fallback), `profileModelPolicyOptions` 87.5% (empty-label fallback) | `after-cli-new-tests.txt`, `after-cli-changed-func-cover.txt` |
| step-1 selection, again | `go test ./internal/cli/ -count=1 -list 'Profile\|Wizard\|HuhTheme\|UpdateVersion\|TUI' -timeout 600s`, then the `-run` form with `-v` | 0 / 0 | 132 selected (129 + `TestProfileOptions_ValueSetsEqualSchema`, `TestProfileOptions_Labels`, `TestUpdateVersionDowngradeConfirm_Localized`) · RUN 224 · PASS 224 · FAIL 0 · SKIP 0 | `after-baseline-cli-list-selection.txt`, `after-baseline-cli-tests.txt` |

Host tmux at the end: `no server running on /private/tmp/tmux-501/default`. No binary was run against the real HOME; every pty child ran with the scrubbed environment (verified per session). No `go test` ran in parallel with another.

AC status after Stage B (this tree): AC-ITI-005 (4) GREEN (mech). AC-ITI-012 GREEN (golden, four cases, three mutants). AC-ITI-015 (a) RED captured pre-fix (v1, column 22) and still RED post-M2 (v2, column 21) — fix is M7. AC-ITI-019 now also asserted over cli's capture tests (SKIP ×5 ungated, FAIL ×4 without tmux). AC-ITI-013 remains M7 (the confirm surface already renders table labels only: `←/→ 전환 • enter 제출 • y 예 • n 아니오`).

#### t583 absorb gate (lead-opened, 2026-09-12)

The lead opened the t583 absorb gate after `origin/develop` advanced to `03a48b0df`. Verified before merging that `c2a9dbcc8` (the t583 tip) is an ancestor of `origin/develop` (`git merge-base --is-ancestor`, exit 0), and that local `develop` and `origin/develop` name the same commit (`git show-ref`).

Merge commit `10ea337fe`, parents `4d985fe48` (this branch) and `03a48b0df` (develop). No conflicts; the merge committed cleanly.

The concern the gate was opened for — t583 cut the wizard questions 16→4 and deleted `WizardResult` fields, which could collide with the option shapes and translation keys Stage B fixed — did not materialize at either the textual or the compile layer.

| Check | Command | Exit | Observed | Evidence |
|---|---|---|---|---|
| file overlap | `git merge-base develop HEAD` → `CARD_BASE`, then `comm -12` of `git diff --name-only $CARD_BASE..HEAD` (report paths dropped) against `git diff --name-only eb50af5a8..03a48b0df` | 0 | card contributes 138 files (33 outside `.moai/reports/`); develop delta 2322 files; **intersection 0**. Control: the card-scope side is 33, not 0, so the empty intersection is a measurement rather than an empty operand | `.moai/reports/t586/absorb-t583/file-overlap.txt` (empty), `README.md` |
| build, before | `go build ./...` (at `4d985fe48`, pre-merge) | 0 | no output | observed in-session; re-derivable at `4d985fe48` |
| wizard, before | `go test ./internal/cli/wizard/` (at `4d985fe48`, pre-merge) | 0 | `ok … 3.399s` | observed in-session; re-derivable at `4d985fe48` |
| build, after | `go build ./...` | 0 | no output | `.moai/reports/t586/absorb-t583/post-build.txt` |
| vet, after | `go vet ./internal/cli/...` | 0 | no output | `post-vet.txt` |
| wizard, after | `go test ./internal/cli/wizard/` | 0 | `ok … 3.130s` | `post-wizard.txt` |

The before/after pair is what attributes the result: the absorb broke nothing. A deleted `WizardResult` field still referenced by Stage B code would have failed the build, and it did not.

The overlap measurement takes its left edge from the merge-base with the absorbed ref, never a literal base SHA (`gitflow-lane-protocol.md` §8). Measured against the previous absorb pin `cd7dc491c` instead, the card reads as 56 files and the Phase A output (`internal/cli/ptycaptest/**`, `profile_questions.go`, `profile_translations.go`) falls outside the range — the same false narrowing that rule exists to prevent.

Not established yet: a green build says the tree compiles, not that the goldens still match. The test-layer verdict for `internal/cli` — including the four `internal/cli/testdata/downgrade-confirm/*.golden` files — waits on the `internal/cli` slot. It will run the same selector range Stage B used (`'Profile\|Wizard\|HuhTheme\|UpdateVersion\|TUI'`), compared against the after-batch control (132 selected, RUN 224, PASS 224, FAIL 0, SKIP 0) rather than `baseline2` (129/217), which predates Stage B. The selected-name count and the top-level PASS count are compared to each other: a `-run` selector drops names that no longer exist without saying so, and t583 deleted wizard tests, so a count below 132 may be correct — it has to be attributed to t583's deletions, and a residue that is not is a defect.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

### Carried sync obligations

Work the run phase deliberately did not do, because doing it would have crossed an ownership boundary. Each line names the artifact, the edit, and why it waits.

- **`research.md` §14 — correct the derivation command to `[Aa]pplyAutonomyTierBundle(Fn)?\(`.** t656 (`b5b5883e9`) put the call behind the test seam `applyAutonomyTierBundleFn`, so the recorded pattern's capital-`A` literal no longer matches it and the step-1 output dropped from 11 lines to 10. The producer and the watch list are unchanged — the call still stands at `internal/cli/init.go:873` and still writes the same `~/.claude/settings.json` — so W1-W6 and AC-ITI-019 need no revision. What is stale is the command, and only the command. Left to sync because run may not edit plan artifacts' body content. Without this line the next re-derivation drops the producer silently, which is the whole reason it is written down. Control pair and measurement: `.moai/reports/t586/absorb-t583/slot/c-recheck.md` § research.md §14.

- **`spec.md` §D (D4) — re-examine the exclusion wording against a fresh measurement.** D4 excludes the step-indicator line on the ground that it is not visible at the same pty size. The AC-ITI-003 M4-mutant capture taken during this run shows `● ○ ○ ○ 1 / 4` **visible** on an 80×30 pty (`.moai/reports/t586/ac003-init-first-screen-m4mutant-green.txt`), which contradicts the observation D4 rests on. This card does not act on it: the step indicator belongs to AC-ITI-021, and D4 is `spec.md` body content that run may not edit. What sync owes is a re-measurement and, if it confirms the line is visible, a correction of D4's stated ground — not necessarily of its conclusion, since an exclusion can survive a wrong reason. Recorded because an exclusion resting on a falsified observation is invisible to every later reader who does not re-measure.

## AC-ITI-003 — pty 사례 착지 (수리 전 RED)

`internal/cli/ptycap_child_test.go` 에 init 첫 화면 사례를 더했다. 자식 사례 `init-first-screen` 은
`TestPtyCaptureChild` 의 switch 에 case 하나로 들어가고(`ptycaptest.ChildTestName` 이 패키지당 함수
하나만 지목하므로 두 번째 자식 함수를 만들지 않는다), 실제 명령 경로를 cobra 순서 그대로
(`validateInitFlags(initCmd, nil)` → `runInit(initCmd, nil)`) 돈다. 부모 테스트는
`TestPtyCapture_InitFirstScreen` 이다.

판정 순서는 §B P6 대로다. 기준 문자열 `Select conversation language` 가 뜬 화면을 잡고 →
실효 환경 관측(`VerifyChildEnv`) → 반출 → 기준 문자열 줄과 옵션 줄 4개를 **먼저** 단정 →
그 뒤에야 `No profile found` 부재와 `confirmButtonPairs` 버튼 줄 부재를 본다 → `Ctrl+C` →
`Initialization cancelled.` 대기 → 세션 집합 비교. 실제 HOME 감시 비교는 `t.Cleanup` 이다.
단계 표시 줄은 이 AC 에서 판정하지 않는다(`spec.md` §D, D4).

**판정: FAIL(수리 전 RED).** REQ-ITI-001 이 `runInit` 의 프로필 확인창을 지우기 전까지(plan.md M4)
확인창이 화면을 잡고 있어 첫 기준 문자열이 기한 안에 나타나지 않는다. 이 RED 는 AC 가 겨누는
결함 그 자체를 찍었다 — 마지막 캡처에 `No profile found. Set up profile preferences now?` 와
`Yes     No` 버튼 줄이 함께 있다. 증거 `.moai/reports/t586/ac003-init-first-screen-red.txt`.

**도달성(공허한 초록 아님).** 부재 단정이 "코드가 아예 안 돌아서" 통과하는 것이 아님을 M4 뮤턴트로
보였다. `init.go` 의 프로필 블록 조건을 `false` 로 만든 트리에서 같은 테스트가 PASS 하고, 캡처에
기준 문자열·옵션 줄 4개·`Initialization cancelled.` 가 모두 있다. 뮤턴트는 되돌렸다(`git status`
에서 `internal/cli/init.go` 는 미수정). 증거
`.moai/reports/t586/ac003-init-first-screen-m4mutant-green{,-cancelled,-log}.txt`.
그 캡처에는 `● ○ ○ ○ 1 / 4` 스테퍼 줄이 80×30 에서 보인다 — D4 관측 기록일 뿐 이 AC 의 판정 근거가
아니다.

**테스트 수 결합.** 새 `TestPtyCapture_*` 하나가 늘어 `AssertSkipWithoutGate` 최소치 5→6,
`AssertFailWithoutTmux` 최소치 4→5 로 같은 커밋에서 고쳤다. 선택자 수 대조:
`-list '^TestPtyCapture'` 6개 ↔ 게이트 실행 결과 6개(SKIP 1 · PASS 3 · FAIL 2).
FAIL 2 는 이 사례와 M7 대기 중인 `TestPtyCapture_DowngradeConfirmButtonAlignment` 다.

---

## M4 — init·update 흐름에서 프로필 경로 제거 (2026-09-12, `WT-init-tux-i18n`)

기준선 HEAD `e5f5692c6`. 커밋 둘: RED `f3fc13fae`, 수리 `312f30825`.

### plan.md 좌표는 낡아 있었다 — 재측정

`plan.md` §F M4 은 `init.go:647-663`, `update.go:174-192` 를 가리키지만 t583 흡수 전 좌표다.
흡수 트리 `e5f5692c6` 에서 다시 쟀다(`git grep -n`):

| 대상 | plan.md | 실측 `e5f5692c6` |
|---|---|---|
| init 확인창 제목 | 647-663 | `internal/cli/init.go:618` (호출 `:626`) |
| update 확인창 제목 | 174-192 | `internal/cli/update.go:180` (호출 `:188`) |

### 이음새

`profile.go` 에 `runProfileSetupFn` 을 뒀다(`runWizardFn` 관용구와 같은 주입형 패키지 변수).
두 **명시** 진입이 이 이음새를 지난다 — `profileSetupCmd.RunE` 와 `runProfileCmd` 의 `--setup`
분기. `runProfileSetup` 자체는 건드리지 않았다(M5 의 대상).

### 판정 (모두 이 실행, 이 트리, HEAD `312f30825`)

| 항목 | 명령 | 결과 |
|---|---|---|
| AC-ITI-001 행동 | `go test -run 'TestInitUpdateEntry_NeverRunsProfileWizard'` | PASS (8 조합: stdin tty/pipe × init·init --non-interactive·update·update --yes) |
| AC-ITI-001 소스 | `go test -run 'TestInitUpdateSource_CarriesNoProfileEntry'` | PASS (대조군 2개 포함) |
| AC-ITI-002 | `go test -run 'TestInitInteractive_AsksConversationLanguageExactlyOnce'` | PASS (대화 언어 질문 1회·인덱스 0) |
| AC-ITI-002 꼬리 | `go test -run 'TestProfileExplicitEntries_RunProfileWizardOnce'` | PASS (두 명시 진입 각 1회) |
| AC-ITI-003 | `MOAI_PTY_CAPTURE=1 go test -run 'TestPtyCapture_InitFirstScreen'` | **RED→PASS** |
| RED 원장 L2 | `git grep -n 'No profile found' -- '*.go' ':!*_test.go'` | 0행(exit 1) · 대조군 `Initialization cancelled` 1행(exit 0) |
| RED 원장 L4 | `git grep -n 'runProfileSetup(' -- init.go update.go` | 0행(exit 1) · 대조군 profile.go 1행(exit 0) |
| 패키지 전량 | `go test ./internal/cli/ -count=1 -timeout 1800s` | `ok … 1582.680s`, `--- FAIL` 0건 |

선택자 수 대조: `-list` 4개 ↔ 최상위 PASS 4개. (600s 로는 이 패키지가 완주하지 못한다 —
`panic: test timed out after 10m0s`. 1800s 로 다시 재서 통과.)

### RED 는 겨눈 이유로 실패했다

수리 전 `TestInitUpdateSource_CarriesNoProfileEntry` 가 4행 모두 "1건 잔존" 으로 FAIL 했고,
대조군 2개는 통과했다 — 0 기대가 빈 읽기가 아님을 그 자리에서 세웠다.
증거 `.moai/reports/t586/m4-ac001-ac002-red.txt`, `m4-red-ledger-l2-l4.txt`.

행동 쪽 판정은 수리 전에는 **공허한 초록**이었다. 수리 전 프로필 블록은 `isatty` 로 막혀
`go test` 에서 도달 불가이고, 그 호출은 이음새를 지나지도 않았다. 비공허성은 AC-ITI-002 가
규정한 뮤턴트로 세웠다 — init 위저드 호출 **직전**에 `runProfileSetupFn(cmd, nil)` 을 되살린
트리에서 `conversation_language questions issued in the whole run = 2, want 1` 로 FAIL 하고,
이음새 카운터도 `calls = 1, want 0` 으로 FAIL 한다. 뮤턴트는 되돌렸다(sha256 대조로 확인).
증거 `.moai/reports/t586/m4-ac002-mutant-red.txt`.

두 판정이 서로 다른 모양을 잡는다: 이음새를 **지나는** 되살림은 카운터가, 이음새를 **우회한**
직접 호출은 소스 스캔이 잡는다. 어느 하나만으로는 두 모양을 다 못 본다.

### AC-ITI-003 첫 화면

수리 뒤 캡처의 첫 화면이 `Select conversation language` + 옵션 4줄
(`English`, `Korean (한국어)`, `Japanese (日本語)`, `Chinese (中文)`) 이고, `No profile found` 도
`Yes`/`No` 버튼 줄도 없다. `Ctrl+C` 뒤 `Initialization cancelled.` 가 나온다. 즉 M4 뮤턴트
(`false &&`)로 미리 본 초록을 실제 수리가 그대로 재현했다.
캡처 `.moai/reports/t586/ac003-green-capture/init-first-screen{,-cancelled}.txt`.

### 테스트 수 결합

새 `TestPtyCapture_*` 를 **더하지 않았다**. `AssertSkipWithoutGate` 의 `^TestPtyCapture` 6,
`AssertFailWithoutTmux` 의 `^TestPtyCapture_` 5 는 그대로 맞다(파일 스캔으로 확인:
`TestPtyCaptureChild` + `TestPtyCapture_*` 5개 = 6).

### 빌드·정적검사

`go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 ·
`go vet ./internal/cli/...` exit 0 · `golangci-lint run ./internal/cli/...` `0 issues.`

`huh`·`isatty` import 가 `init.go`·`update.go` 에서 쓰이지 않게 돼 함께 지웠다. `moaiHuhTheme`
은 `profile_setup.go` 가 아직 쓰므로 남는다(M6 소관).

## M5 — 흡수 배선과 스테퍼 (2026-09-12, `WT-init-tux-i18n`)

기준선 HEAD `7fc66fc72`. 커밋 넷: RED `293f1373f`, 수리 `244f3db38`, 골든 `f39fd1844`,
가드 재조준 `fa07e3352`. plan.md 좌표는 재측정 없이 design.md §2.2/§4/§5/§10 과
acceptance.md 본문을 직접 판든했다(M4 교훈).

### 이음새와 흡수 구조

`profile.go` 에 시음새 셋을 뒀다 — `profileWizardRunner`(design.md §2.2, 초기값·옵션·로케일 →
응답), `enterSessionWorktreeFn`/`cleanupSessionWorktreeFn`(§5, runWizardFn 관용구).
`runProfileSetup` 은 v2 폼을 `wizard.RunProfile` 로 실행하고, `errors.Is(err,
wizard.ErrCancelled)` 로 취소를 판별해 `선택된 로케일`의 `SetupCancelled` 문구를 stdout 에 쓰고
nil 을 돌려준다(취소 로케일 = 답한 언어 → 저장값 → en). 저장 이하는 v1 과 동일한 순서
(acceptEdits 정규화 → WritePreferences → SyncToProjectConfig(nil 세그먼트) →
persistProjectConfig(devMode, "") → 저장 문구·요약)이며 `@MX:NOTE`/`@MX:REASON` 은
normalizeModel 에서 새 바인딩 지점 `initialProfileResult` 로 옮겼다.

스테퍼는 design.md §4 대로 `questionVisibility`(가시성 판정 클로저) + 바인딩 대상으로
일반화했고 init 은 `wizardResultVisibility` 얇은 감싸개로 문자열을 유지한다(AC-ITI-009
init 대조군 통과). 프로필 폼의 표시기는 **그룹 타이틀**으로 넣었다 — huh v2 가 그룹
뷰포트를 focused 필드 위치로 스크롤해 콘텐츠 내 note 행을 잘라버리는 것을 formDriver 로
측정했고(plan.md §G D4 와 같은 현상군), 헤더는 뷰포트 밖에서 렌더링돼 잘리지 않는다.
문자열은 두 쪽 모두 `tui.Stepper(k, N, nil)`.

### AC-ITI-006 — 보존 표 테스트 (9사례 전부 PASS)

`go test -run 'TestProfileSetupAbsorbed_PreservationTable' -count=1` → `ok`. RED 는
`293f1373f` 에서 라우팅 게이트로 세웠다(19 하위 테스트 전부 "does not route through
profileWizardRunner(" 로 실패, `.moai/reports/t586/m5-ac006-ac007-red.txt`). 수리 전
(6)(7) 사례는 실제로 FAIL 했다 — config manager 의 `Save()` 가 무관한 섹션 쓰기마다
git-convention.yaml 을 partial-override 확장으로 재작성하기 때문(2줄 시드가 11줄로 확장되는
것을 프로브로 측정). 수리는 ConfigManager.Save 에 git-strategy 선례(SPEC-GITSTRATEGY-SAVE-
ISOLATION-001)와 같은 git_convention dirty/absent 격리를 더한 것 — **범위 주의**:
plan.md M5 의 파일 목록(profile_setup.go, wizard.go)을 넘어 internal/config/manager.go 를
만졌고, fan_in 12 함수다. 리드 재판정 대상.

### AC-ITI-007 — 명령 수준 이음새 계약 (8조합 전부 PASS)

setup·--setup 두 진입 × (a)취소 (b)오류 (c)성공(이름 없음/work). (a) nil 오류 + ko
`설정이 취소되었습니다.` 출력 + 프로필 파일 부재, (b) non-nil 오류 + 파일 부재, (c) 저장
문구·요약 출력 + preferences.yaml 생성. 정리 이음새는 실행당 정확히 1번, clean-exit 인자는
(a)(c) 참·(b) 거짓. 진입-선행-읽기는 enter 이음새가 쓴 센티넬 preferences 파일이 캡처된
초깃값으로 흘러드는 것으로 증명(c 성공 사례).

### AC-ITI-008 — 로케일 골든 (ko/ja/zh PASS) + 집합 E 판정

`TestProfileWizardGolden_LocaleFrames` PASS(골든 `internal/cli/testdata/profilewizard/
groups-{ko,ja,zh}.golden`, RED 원문 `.moai/reports/t586/m5-ac008-red.txt`),
`TestProfileWizardGolden_NoEnglishLeak` PASS. 폼은 실행 초기 로케일로 구성한다 — huh v2 의
키맵과 옵션 목록은 폼당 정적이라 언어를 폼 안에서 바꾸는 구성은 도달 불가다(제목·설명은
로케일 포인터로 폼 안에서 재렌더링된다). **집합 E 해석(판정 기록)**: E 는 "대상 로케일
렌더링이 en과 다른" 텍스트의 en값 — 실제 번역된 문자열만 누설 판정 대상으로 삼는다.
로케일 불변 렌더링(스키마의 빈 옵션 리터럴 "(runtime default)"/"(project default)",
번역표가 en 텍스트를 그대로 지닌 model_policy 3개 라벨)은 모든 로케일에서 자기 자신을
그릴 뿐 English 누설이 아니며, 이를 X에 하나씩 적는 것은 SPEC이 번역을 제공하지 않은
텍스트를 위해 닫힌 목록을 넓히는 일이다. 리드가 (A)엄격 독해(전부 결함→번역 추가)를
택하려면 model_policy 라벨 번역 + settings 빈 라벨 현지화가 후속 작업이다.

### AC-ITI-009 — 스테퍼 형식 (프로필 + init 대조군 PASS)

`TestProfileWizardStepper_SameFormatAsInit` PASS. 프로필: N=10, 페이지 첫 줄이
`<k> / 10`(k=1,2,3,6,10)로 끝나고 ●/○ 합이 10. init 대조군: `InitQuestions` 기준 N=4,
페이지 첫 줄 k=1,3,4 — plan.md의 "N은 보이는 init 질문 수"대로 실제 init 세트(퇴고한
4문항)를 썼다. 행 우측 패딩은 TrimRight 후 접미 판정.

### AC-ITI-010 — 가드 재조준 9건 + 뮤턴트 10건 전부 BITES

재조준된 9개 가드 셀렉터 실행: `=== RUN` 최상위 9개, 전부 `--- PASS`. 뮤턴트(양성 7:
S1 모델정책 질문 삭제, S2 호출줄 삭제, S3 정규화 비교 삭제, S4 인라인 리터럴, S5
development_mode 삭제, S6 effort_level 삭제, S9 nil 세그먼트 대입 삭제 / 음성 3: S7
`ID: "statusline_theme"` v2형 추가, S2 yaml.Marshal 호출 추가, S8 StatuslineTheme 대입
추가) 10건 전부 해당 가드가 실패했다. 원문 `.moai/reports/t586/m5-ac010-mutants.txt`
(BITES: yes ×10). 뮤턴트 적용·복원은 cp 백업 + cmp 검증으로 했고 git 명령은 일절 없다.
복원 후 제품 파일은 HEAD와 byte-identical(`git diff --stat` 공집합).

### pty 스위트 (게이트 켬)

`MOAI_PTY_CAPTURE=1 go test -run 'TestPtyCapture'` → Localized·InitFirstScreen·
SkipWithoutGate·FailWithoutTmux PASS, **잔여 FAIL은 정확히 1개** —
`TestPtyCapture_DowngradeConfirmButtonAlignment`(M7 소관, RED 유지가 M5의 의무였다).
새 `TestPtyCapture_*` 를 더하지 않았다: `AssertSkipWithoutGate` 6 / `AssertFailWithoutTmux`
5 그대로.

### 빌드·정적검사·커버리지

`go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 ·
`golangci-lint run ./internal/cli/...` `0 issues.` · `go vet` 문제 없음 ·
`go test -cover ./internal/cli/wizard/` `coverage: 93.0% of statements`(하한 85% 이상,
M4 종료 시 93.2% 대비 -0.2p). 전량 스위트 `go test ./internal/cli/ -count=1
-timeout 1800s` 결과는 `.moai/reports/t586/m5-internal-cli-suite.txt`.

### 남긴 관찰 (M7/M8 소관 이하 — 수리하지 않음)

1. huh v2의 Eval(제목 비동기 적용)은 프로그램 드라이버에서 적용이 한 번 유실되면 해당
   페이지 표시기가 빈 채로 남는다 — 프로필 표시기를 그룹 타이틀로 옮겨 회피했고, init 쪽은
   기존 반응형 노트가 그대로다(기존 테스트 전부 통과).
2. 프로필 옵션 라벨은 실행 초기 로케일로 고정된다(M1 설계대로). 새 프로필(en 초기)이 폼
   안에서 ko를 골랐을 때 이후 그룹의 제목·설명·도움말… 중 제목·설명만 재현역화되고 옵션
   라벨은 초기 로케일에 머문다. REQ-ITI-007 전항을 충족하려면 옵션 데이터의 로케일별
   전달이 필요하다 — design.md §3 기각 사항과 맞물려 있어 설계 변경 몫이다.
3. model_policy 3개 라벨과 스키마 빈 옵션 리터럴은 번역표가 en 텍스트를 지닌다(현지화
   없음). AC-ITI-008 판독 (B)에서는 허용, (A)에서는 번역 추가가 필요 — 위 판정 기록 참조.

### 슬롯 임대 경위 (리드 지시 접수, 2026-09-12 야간)

M5의 중실행 4건(전량 스위트 1800s 타임아웃 실행·2700s 재실행, internal/cli·wizard 게이트 켠
pty 스위트)은 모두 슬롯 임대 지시 접수 **이전**에 실행됐고 임대 없이 돌았다. 이 중 첫 전량
실행이 lane-8의 타이밍 재측정을 오염시켰다(부하 붕괴 3-9배, 리드 관측) — 타임아웃 원문은
m5-internal-cli-suite-timeout1800.txt로 보존돼 있다. 지시 접수 이후 이 레인의 남은 중실행은
없으며, 재실행이 필요해지면 lane-local 바이너리로 `slot acquire --resource heavy-test`를
선취한 뒤에 돌린다.

## M6 — v1 퇴역과 모듈 정리 (2026-09-12, `WT-init-tux-i18n`)

기준선 HEAD `c65e1d856`. 커밋 셋: RED `f95d7e931`, 수리 `bdd6b952e`, 마감(이 커밋).
plan.md 좌표 `launcher.go:1094-1110` 은 내용으로 재탐색해 판독했다.

### 삭제와 정리

`huh_theme.go`·`huh_theme_test.go` 삭제. `profile_setup.go` 에서 `huhV1Options` 와
huh v1 임포트 제거(어댑터 테스트 `TestHuhV1Options_PreservesLabelAndValue` 도 함께 은퇴).
RED 단계에서 발견: `update/preview_tui.go:85` 주석이 삭제될 `cli.huhThemeIsDark` 를
참조해 AC-ITI-011 (2) 를 계속 실패시키게 되므로 본문 수정 대상에 추가해 `wizard.wizardIsDark`
만 남겼다. `go mod tidy` 는 huh v1 과 그 전이 의존성(bubbles/bubbletea v1, coninput,
localereader, muesli/ansi, golang.org/x/sys 구행)을 떼어냈다.

`launcher.go` 주석(내용으로 재탐색 — `runProfileSetup` 확인 문구 참조 블록, ~1094-1110)
재독 판정: acceptEdits 확인 문구가 M5 의 v2 본문에서 그대로 나오므로 거짓이 된 부분 없음 —
변경 없음. (블록 내 단락 중복은 기존 상태로, 이 SPEC 의 판정 대상이 아니다.)

### AC-ITI-004 판정 (전 조항 PASS, 커밋 `bdd6b952e` 이후 측정)

(1) huh v1 임포트 비테스트 0줄·exit 1(대조군 huh v2 다수) ✅ (2) 테스트 0줄·exit 1 ✅
(3) go.mod huh v1 0, 대조군 huh v2 1 ✅ (4) `go mod tidy && git diff --exit-code go.mod
go.sum` exit 0 ✅ (5) `GOOS=windows GOARCH=amd64 go build ./...` exit 0 ✅ (6) 두 명시
진입의 동일 이음새 도달 = `TestProfileExplicitEntries_RunProfileWizardOnce` PASS(M5의
AC-ITI-007 테이블이 같은 사실을 행동으로도 재확인). 원문
`.moai/reports/t586/m6-ac004-ac011-green.txt`.

### AC-ITI-011 판정 (전 조항 PASS)

(1) `git ls-files internal/cli/huh_theme.go internal/cli/huh_theme_test.go
internal/cli/wizard/wizard.go` → 정확히 `internal/cli/wizard/wizard.go` 1줄 ✅
(2) 테마 심볼 정규식 0줄·exit 1(대조군 `var wizardIsDark` 1줄) ✅
(3) `wizardIsDark` 참/거짓 강제(패키지 변수만, 환경 변수 불볕)로 그린 downgrade confirm·
프로필 첫 그룹의 ANSI 포함 골든 4개(`internal/cli/wizard/testdata/axis/`) — 두 축 문자열이
서로 다름을 단정하고 각각 저장 골든과 일치. **잔여 위험**: 골든은 truecolor ANSI 시퀀스를
포함하므로 색 프로파일 감지 환경(TERM 등)이 다른 곳에서 재실행하면 불일치할 수 있다 —
절 (3)이 ANSI 포함 비교를 요구하는 귀결이다.

### 슬롯 임대 중측정 (리드 프로토콜 준수)

`/tmp/moai-lane2 slot acquire --resource heavy-test --max-duration 30m --name lane-2`
→ 취득(da011b34-…, 2026-09-12T12:58:42Z까지). 임대 창 안: `go test ./internal/cli/
-count=1 -timeout 2700s` → `ok 1077.913s`, `go test ./internal/cli/wizard/ -count=1
-timeout 600s` → `ok 3.538s`. 완료 후 `slot release` → "released" 확인,
`slot status` → "no slot leases recorded". 원문 m6-cli-remeasure.txt·
m6-wizard-remeasure.txt·m6-slot-release.txt.

### 빌드·정적검사 (tidy 이후)

`go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 ·
`golangci-lint run ./internal/cli/...` `0 issues.` · AC-ITI-010 의 9가드 셀렉터 PASS.

## M7 — 레이아웃·도움말·그룹 (2026-09-12, `WT-init-tux-i18n`)

기준선 HEAD `5403cdeb8`. 커밋: RED/가드 `b7b9d2ab6`, 수리 `5e716116b`(REQ-ITI-017),
구현 `7b5d056ca`(REQ-ITI-012~016 + 문자열 통합), 마감(이 커밋). 모든 좌표는 내용으로
재탐색했다(wizard.go:290 단순 연결, launcher.go는 M6에서 판정 완료).

### REQ-ITI-014 — 확인 버튼 좌측 정렬 (AC-ITI-015 PASS)

`buildConfirmField` 과 다운그레이드 헬퍼에 `WithButtonAlignment(lipgloss.Left)`.
추가 측정: huh 기본 버튼 스타일의 좌측 패딩 2칸이 라벨을 설명 열(2)보다 4로 밀었으므로
테마에서 `PaddingLeft(0)` 으로 덮어 라벨 열을 설명 열과 맞췄다. 두 표면(다운그레이드
확인창, 위저드 확인형 픽스처) 모두 버튼 라벨 시작 열 == 설명 첫 글자 열. 골든
`testdata/axis/confirm-alignment-downgrade.golden`. **게이트 켠 pty:
`TestPtyCapture_DowngradeConfirmButtonAlignment` 이 GREEN 으로 전환 — 잔여 FAIL 0건.**

### REQ-ITI-015 — 필드 구분자·선택 높이 (AC-ITI-016 PASS)

테마 `FieldSeparator` 를 개행 1개로 축소(필드 사이 빈 줄 0), `NoteTitle` 마진 제거(콘텐츠
영역 갭 프리), 선택 높이 산정 — huh v2 의 `Height(n)` 은 필드 총높이(제목+설명+옵션)라서
옵션 수만 넣으면 옵션이 잘린다(측정: 언어 4옵션이 1개로 렌더). OptionsFunc 경로(init)는
`selectHeight(q)` 로 제목 1+설명 줄 수+옵션 수를 계산하고, 비동적 경로(프로필)는 Height
미호출로 huh 자체 정확 산정을 쓴다. 판정: init 첫 페이지·프로필 전 그룹의 콘텐츠 접두사에
모든 필드 앵커가 연속 존재하고 빈 카드 행 0.

### REQ-ITI-016 — 설명 열 정렬 (AC-ITI-017 PASS + 뮤턴트 2건 BITES)

`wizard.go:290` 의 단순 연결을 `alignOptionLabels` 로 대체 — 라벨을 표시 폭(runewidth,
전각 2칸)으로 재폭측정해 최대 폭(80 컷오프)까지 패딩한 뒤 설명을 붙인다. init·프로필
conversation_language 옵션 4줄의 설명 시작 표시 열이 동일함을 판정. 뮤턴트 2건 — 룬 수
패딩, 바이트 수 패딩 — 모두 이 검사가 FAIL 했다(m7-ac017-mutants.txt, BITES yes ×2).

### REQ-ITI-012 — 키별 도움말 라벨 (AC-ITI-013 M7 몫 PASS)

`buildUnifiedForm` 에 `WithKeyMap(localizedKeyMap(locale))` 을 걸어 init·reconfigure
표면의 도움말 줄을 로케일화했다(프로필·다운그레이드는 기존 적용). 4표면 × 4로케일:
위저드 패키지 12골든(help-surfaces/, ko·ja·zh 도움말 줄에 영어 동작 라벨 부재 단정) +
프로필 표면 4골든은 cli 의 AC-ITI-008 몫과 NoEnglishLeak 이 판정. HelpSelect/HelpInput
삭제는 M8 몫이라 손대지 않았다.

### REQ-ITI-017 — 그룹 재구성 (AC-ITI-018/021/022 PASS + 독립성 뮤턴트 2건)

`questions.go` 두 리터럴의 `Group` 을 `Agents & Autonomy` 로 — init 페이지 3→2, 분모 4
불변. 영향 갱신: agent_wiring_question_test 단정, 패키지 주석, expansion/
question_removal/restructure 테스트의 구조 단정, M5 AC-ITI-009 init 대조군.
독립성 뮤턴트: (a) 그룹 분리 → 018 FAIL(3 groups)·021 PASS, (b) 무조건 질문 추가 → 021
FAIL(분모 5)·018 PASS. 두 번째 그룹 회귀 골든(testdata/axis/init-regroup-second-group,
끝이 3 / 4) 커밋 — 두 뮤턴트 모두에서 변하므로 어느 쪽 증거로도 세지 않는다.
원문 m7-ac018-ac021-mutants.txt.

### AC-ITI-022 판독 주의

`.Group` 읽기 허용 범위는 design.md §9 측정 시점(읽는 줄이 buildFormGroups 2줄뿐) 이후
M5 흡수로 buildProfileForm 의 묶기 비교·대입 2줄이 같은 성질로 추가됐다. 가드는 이 두
분할 지점의 묶기 비교/대입만 허용한다(라벨 렌더 금지라는 성질은 센티널+번역 키 부재
단정이 그대로 수행).

### 문자열 통합

`helpActionLabels`·`downgradeConfirmTexts`·`profileQuestionTexts` 를 translations.go 로
이동(profile_translations.go 삭제, LocalizeProfileQuestion 은 profile_questions.go 로).
이 이동이 만든 추적-작업트리 불일치 1건(삭제 미스테이지)을 TestVersionStampRegistry 의
judged/handed 갭이 정확히 잡았다 — 레지스트리 설계대로의 동작이며 커밋으로 해소.

### 슬롯 임대 중측정 (리드 프로토콜 준수)

acquire(da011b34-…, 14:24:59Z까지) → wizard ok 3.8s·internal/cli ok 990.931s → release
확인. 이전 600s 무-flarge 실행은 -timeout 누락 아티팩트였다. pty 스위트도 임대 창 안에서
실행(ok 28.311s, 잔여 FAIL 0). 원문 m7-remeasure-slot-leased.txt(1차: 축 골든 스테일
FAIL로 중단 — 골든 재생성 후 재실행)·m7-cli-remeasure.txt(3건 FAIL — 골든 2세트 스테일+
레지스트리 갭, 모두 해소)·m7-pty-capture-suite.txt·m7-final-cli-suite.txt
(ok 990.931s)·m7-slot-release2.txt.

### 빌드·정적검사

`go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 ·
`golangci-lint run ./internal/cli/...` `0 issues.`
