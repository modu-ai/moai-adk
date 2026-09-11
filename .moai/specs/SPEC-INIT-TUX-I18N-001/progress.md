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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
