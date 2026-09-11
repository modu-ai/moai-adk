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

_Phase A milestone work follows in the next entries._

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
