# t586 — plan §C 사전 점검 재측정 (흡수 트리, 2026-09-12)

기준: 흡수 병합 `10ea337fe` 이후, 측정 시점 HEAD `cea33e5a9`.
`plan.md` §C 6항목을 순서대로 다시 돌린 기록.

---

## §C 1 — BASELINE_SHA

```
BASELINE_SHA=cea33e5a9721ebf59d2ce5f970dda94383d75352
absorb_merge=10ea337fe   (부모 4d985fe48 / 03a48b0df)
```

이전 BASELINE_SHA `cd7dc491c` 는 직전 흡수점이다. §C 1 이 요구하는 "t583 흡수 뒤 한 번 더"
를 여기서 잡는다.

## §C 2 — 선택자 기준선 (두 패키지 합산)

```
$ go test ./internal/cli/wizard/... ./internal/cli/ -count=1 -list 'Profile|Wizard|HuhTheme|UpdateVersion|TUI' -timeout 600s
→ 161 이름, exit 0

$ go test ./internal/cli/wizard/... ./internal/cli/ -run 'Profile|Wizard|HuhTheme|UpdateVersion|TUI' -count=1 -v -timeout 600s
→ ok internal/cli/wizard 0.479s · ok internal/cli 12.217s, exit 0
   RUN 256 / top-level PASS 161 / FAIL 0 / SKIP 0
```

**선택된 수 161 ≠ 0 을 먼저 셌고**(§C 2 의 요구), **161 == top-level PASS 161** 로 일치한다.

[HARD] 명령은 **순수 파이프** 형태로 적었다. progress.md 에 적혀 있던 `\|` 형태는 RE2 에서
리터럴 파이프라 0개를 쓸고 `ok` 를 찍는다 — 별도 기록은 `slot/README.md` § 계기 결함,
전수 스윕은 리드가 카드 t659 로 발행했다.

증거: `slot/c2-list.txt`, `slot/c2-combined.txt`.

## §C 3 — acceptance.md §D.3 RED 원장 4건 재실행

판정 규칙은 모두 "출력 0줄 = GREEN". 넷 다 출력이 있으므로 **여전히 RED** 이고, 이는
수리가 아직 안 됐다는 뜻으로 정상이다.

### L1 — `git grep -l '"github.com/charmbracelet/huh"' -- '*.go' ':!*_test.go'` (exit 0)

```
internal/cli/huh_theme.go
internal/cli/init.go
internal/cli/profile_setup.go
internal/cli/update.go
```

**plan 기준선과 다름: 5줄 → 4줄.** `internal/cli/update_version.go` 가 빠졌다. t583 이 아니라
**이 카드의 M2**(v2 downgrade confirm)가 그 파일의 v1 import 를 제거한 결과다. AC-ITI-004 는
이 목록이 0줄이 되어야 GREEN 이므로, 4줄은 M6 가 남긴 몫이다.

### L2 — `git grep -n 'No profile found' -- '*.go' ':!*_test.go'` (exit 0)

```
internal/cli/init.go:618:			Title("No profile found. Set up profile preferences now?").
internal/cli/update.go:180:				Title("No profile found. Set up profile preferences now?").
```

줄 수 2 로 동일. 좌표 이동: init.go 651 → 618, update.go 179 → 180.

### L3 — `git grep -n -E 'HelpSelect|HelpInput' -- '*.go'` (exit 0)

`translations.go` 10줄 + `wizard_test.go` 9줄. 좌표가 크게 이동했다: 표 항목이 542/543 →
320/321. `translations.go` 는 405줄로 줄었다(t583 이 11개 ID 항목을 4개 로케일에서 삭제).

### L4 — `git grep -n 'runProfileSetup(' -- internal/cli/init.go internal/cli/update.go` (exit 0)

```
internal/cli/init.go:626:			if err := runProfileSetup(cmd, nil); err != nil {
internal/cli/update.go:188:				if err := runProfileSetup(cmd, nil); err != nil {
```

줄 수 2 로 동일. 좌표 이동: 651 → 626, 179 → 188.

## §C 4 — pty 조건

```
$ command -v tmux
/opt/homebrew/bin/tmux
$ tmux -V
tmux 3.6a
```

tmux 가 있으므로 `MOAI_PTY_CAPTURE=1` pty 판정은 **FAIL 조건에 걸리지 않는다**
(`acceptance.md` §B P2). 없었다면 E1 표에 FAIL 로 적고 SKIP/Gap 으로 바꿔 적지 않는 규칙이
발동했을 자리다.

## §C 5 — t583 병합 확인

```
$ git merge-base --is-ancestor c2a9dbcc8 HEAD
→ exit 0 (c2a9dbcc8 IS ancestor)

$ ls -d .moai/specs/SPEC-INIT-QUIET-WIZARD-001
.moai/specs/SPEC-INIT-QUIET-WIZARD-001      ← 흡수로 생겼다

$ ls -d .moai/specs/SPEC-INIT-TUX-I18N-001   (대조군)
.moai/specs/SPEC-INIT-TUX-I18N-001
```

plan 단계 트리에서 종료 1 이던 경로가 흡수 뒤 존재한다 — `spec.md` §A.7 이 예상한 대로다.
게이트를 열지 않고 리드에게 올려야 하는 조건(흡수 뒤에도 부재)에 해당하지 않는다.

## §C 6 — 게이트 뒤 좌표 재측정

### §A.7 함수 시작 줄

| 심볼 | plan 기준 | 흡수 트리 | 이동 |
|---|---|---|---|
| `DefaultQuestions` | 48 | 50 | +2 |
| `ReconfigureQuestions` | 268 | 270 | +2 |
| `InitQuestions` | 296 | 309 | +13 |
| `Page3Questions` | 349 | 369 | +20 |
| `RunWithDefaults` | 35 | 35 | 0 |
| `buildFormGroups` | 158 | 158 | 0 |
| `stepperDenominator` | 236 | 236 | 0 |
| `saveAnswer` | 397 | 397 | 0 |
| `saveBoolAnswer` | 467 | 456 | −11 |
| `buildConfirmField` | 487 | 459 | −28 |
| `buildField` | — | 218 | — |
| `var uiStrings` (translations.go) | 540 | 318 | −222 |
| `applyWizardPage3ToOpts` (init.go) | 277 | 287 | +10 |
| 프로필 확인창 블록 (init.go) | 647 | 618 | −29 |

**V-a 종결**: `var uiStrings` 는 이 트리에서 318 이다. plan 이 기록한 두 값(t583 lane-1
보고 548, 흡수 전 트리 540)은 모두 흡수 전 좌표이고, 흡수로 무효가 됐다. M7 의
`translations.go` 편집 범위는 318 이후를 쓴다.

### §A.7 예측 검증 — 셋 다 참

1. **Page3Questions 가 2문항만 남는다.** 참. 현재 본문 주석이 명시한다 — `agent_wiring`
   ("Quality & Workflow") 와 `autonomy_tier` ("Autonomy") 둘뿐이고, 나머지 11문항은
   SPEC-INIT-QUIET-WIZARD-001 REQ-IQW-002 로 제거됐다.
2. **확인형 질문이 남지 않는다.** 참. `QuestionTypeConfirm` 의 비테스트 출현은
   `types.go:73-74`(열거값 정의)와 `wizard.go:224`(`buildField` switch 분기) 둘뿐이며,
   질문 리터럴에서 쓰는 곳이 0 이다. 열거값과 `buildConfirmField` 는 남았다 — 이것도 §A.7 이
   말한 그대로다.
3. **t586 이 새로 맡는 일(REQ-ITI-017)이 유효하다.** 참. `agent_wiring` 하나만 든
   "Quality & Workflow" 그룹이 실제로 남아 있어, Q5 가 확정한 `Agents & Autonomy` 한 그룹
   통합안이 그대로 적용된다.

**부수 관측(범위 밖, 기록만)**: `saveBoolAnswer` 는 이제 본문이 빈 함수다 —
`func saveBoolAnswer(id string, value bool, result *WizardResult) {}`. t583 이 모든 분기를
지우고 껍데기를 남겼다. 이 카드의 수리 대상이 아니므로 손대지 않는다.

### research.md §15 — 자식 환경 정리 목록: **불변**

```
$ grep -c -E '"MOAI_KANBAN' internal/config/envkeys.go
9
$ grep -rn -E 'Getenv\((config\.)?EnvMoaiKanban[A-Za-z]*\)' internal --include='*.go' | grep -v -c '_test\.go:'
20
```

둘 다 기준선과 같다(9, 20). `acceptance.md` §B 의 자식 환경 정리 목록은 고칠 것이 없고,
그 상수를 쓰는 **AC-ITI-020 은 재판정 사유가 없다**. (`EnvClaudeConfigDir` 좌표만 365 → 372
로 밀렸는데, 이 값은 목록에 쓰이지 않는다.)

### research.md §14 — 감시 목록 도출: 출력이 **11줄 → 10줄**로 줄었다

빠진 줄은 `internal/cli/init.go:890  if tierErr := project.ApplyAutonomyTierBundle(` 이다.

**원인은 t583 이 아니라 t656 이다.** `git log -S 'ApplyAutonomyTierBundle' -- internal/cli/init.go`
가 `b5b5883e9` (t656, "wire settings.json snapshot into init, update and clean reinstall") 를
가리킨다. t656 이 테스트 시접 변수를 도입해 호출 형태를 바꿨다:

```go
// internal/cli/update_settings_snapshot.go:32
var applyAutonomyTierBundleFn = project.ApplyAutonomyTierBundle
// internal/cli/init.go:873
if tierErr := applyAutonomyTierBundleFn(
```

**호출은 사라지지 않았다 — `init.go:873` 에 그대로 있다.** §14 패턴이 대문자 `A` 로 시작하는
`ApplyAutonomyTierBundle\(` 를 찾는데, 시접 호출은 소문자 `a` 로 시작하고 `Bundle` 뒤에 `Fn`
이 끼어 매치되지 않을 뿐이다.

대조 실험:

| 패턴 | init.go 자율성 호출 적중 |
|---|---|
| `ApplyAutonomyTierBundle\(` (기록된 형태) | **0** |
| `[Aa]pplyAutonomyTierBundle(Fn)?\(` (정정 형태) | **1** — `internal/cli/init.go:873` |

**판정 — 감시 목록 W1~W6 은 고치지 않는다.** 생산자도 대상 파일도 그대로이기 때문이다:
같은 호출이 같은 `~/.claude/settings.json` 을 쓴다. 낡은 것은 **도출 명령**이지 그 결과가
아니다. 따라서 `acceptance.md` §B 의 감시 목록은 불변이고, **AC-ITI-019 도 재판정 사유가
없다**.

다만 도출 명령을 이대로 두면 다음 재도출에서 이 생산자가 조용히 빠진다 — 그래서 정정 형태를
여기 남긴다. `research.md` §14 본문 갱신은 sync 단계 문서 작업에서 함께 한다(지금 고치면
plan 산출물을 run 단계가 고치는 소관 위반이다).

**부수 관측(기록만)**: `internal/cli/ptycaptest/homewatch.go:71` 의 Producer 문자열이
`"project.ApplyAutonomyTierBundle (init), ensureGlobalSettingsEnv (init, update)"` 다. 시접을
거칠 뿐 결국 같은 함수를 부르므로 의미는 여전히 참이고, 문자열 정정은 필요하지 않다.

---

## §C 종합

| 항목 | 결과 |
|---|---|
| 1 BASELINE_SHA | 잡음 — `cea33e5a9` |
| 2 선택자 기준선 | 161 ≠ 0, 161 == PASS 161, FAIL 0 |
| 3 RED 원장 4건 | 넷 다 재실행, 전부 여전히 RED(정상). L1 만 5→4줄(이 카드 M2 기여) |
| 4 pty 조건 | tmux 3.6a 존재 — FAIL 조건 아님 |
| 5 t583 병합 | 조상 확인 + SPEC 경로 존재(예상대로) |
| 6 좌표 재측정 | §A.7 좌표 갱신, 예측 3건 전부 참, §15 불변, §14 도출 명령만 낡음(목록·AC 영향 없음) |

**게이트를 막을 사유 없음.** AC-ITI-019 / AC-ITI-020 재판정 사유도 없다.

🗿 MoAI
