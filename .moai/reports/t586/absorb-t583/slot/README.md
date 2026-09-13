# t586 — internal/cli 슬롯 재측정 (2026-09-12)

흡수 병합 `10ea337fe` 위에서, stage B 와 같은 선택자 범위로 `internal/cli` 테스트 층을
재측정한 기록. 리드가 lane-6 반납 직후 슬롯을 지명해 수행했다.

시작 전 부하: `uptime` load average 13.65 / 16.85 / 18.82 — 다른 레인 작업 중.
전체 스위트는 돌리지 않았고, 아래 선택자 범위만 돌렸다.

## 계기 결함 — 기록된 명령은 아무것도 쓸지 않는다

stage B 가 progress.md 에 기록한 명령은 다음 형태다:

```
go test ./internal/cli/ -count=1 -list 'Profile\|Wizard\|HuhTheme\|UpdateVersion\|TUI' -timeout 600s
```

이 명령을 **적힌 그대로** 실행하면 이름을 **0개** 반환하고 exit 0 으로 `ok` 를 찍는다.
Go 의 `-list` / `-run` 은 RE2 정규식을 받는데, RE2 에서 `\|` 는 교대(alternation)가 아니라
**리터럴 파이프 문자**다. 어느 테스트 이름에도 `|` 가 없으므로 매치가 0 이 된다.

대조 실험 (같은 트리, 같은 순간):

| 형태 | 반환 이름 수 | exit |
|---|---|---|
| `'Profile\|Wizard\|HuhTheme\|UpdateVersion\|TUI'` (기록된 형태) | **0** | 0 (`ok`) |
| `'Profile\|Wizard\|HuhTheme\|UpdateVersion\|TUI'` 의 백슬래시를 뺀 형태 — 즉 `Profile`, `Wizard`, `HuhTheme`, `UpdateVersion`, `TUI` 를 `\|` 가 아닌 순수 파이프로 이은 형태 | **134** | 0 |

즉 stage B 의 evidence 파일(129/132 이름, 217/224 PASS)을 만든 실제 실행은 기록된 명령이
**아니다** — 어딘가에서 백슬래시가 벗겨진 형태로 돌았고, progress.md 에는 벗겨지기 전
형태가 적혔다. 기록과 실행이 갈라져 있다.

위험의 모양: 이 명령을 신뢰해 재실행하는 다음 사람은 0개를 쓸고 `ok` 를 받는다. 빈 집합
위의 초록은 아무것도 주장하지 않는데(`verification-completeness.md` §1.1), 출력만 보면
전수 통과와 구분되지 않는다. 이 카드의 재측정은 이름 수를 먼저 세었기 때문에 걸렸다.

아래 모든 측정은 **순수 파이프 형태**로 수행했고, 이 디렉터리의 증거 파일도 그 출력이다.

## 측정 결과

| 항목 | 값 | 대조(stage B 직후) | 델타 |
|---|---|---|---|
| 선택된 top-level 이름 | 134 | 132 | +2 |
| `=== RUN` | 227 | 224 | +3 |
| `--- PASS` (하위 포함) | 227 | 224 | +3 |
| top-level `--- PASS` | 134 | — | — |
| `--- FAIL` | 0 | 0 | 0 |
| `--- SKIP` | 0 | 0 | 0 |
| exit | 0 | 0 | — |

**선택자 이름 수 134 == top-level PASS 수 134.** `-run` 은 존재하지 않는 이름을 조용히
버리므로 이 두 수를 맞춰 보지 않으면 줄어든 실행을 초록으로 읽는다. 여기서는 일치한다.

## 이름 델타 귀속 — 잔여 0

`comm` 으로 stage B 직후 목록과 대조한 결과:

**추가 5건** (전부 t583):
`TestRunInit_QuietWizardFlagsStillPersist`, `TestRunInit_QuietWizardObserverDetectsNonDefault`,
`TestRunInit_QuietWizardProvisionsMCPByDefault`,
`TestRunInit_QuietWizardSectionFilesMatchNonInteractive`,
`TestRunInit_QuietWizardUnsetResolvesToDefaults`
→ `git log -S` 로 `82fc81b69` (t583 M2 RED) 에서 도입된 것을 확인.

**삭제 3건** (전부 t583):
`TestApplyWizardPage3ToOpts_AuditConfigSetFalseByDefault`,
`TestApplyWizardPage3ToOpts_AuditSelection`, `TestRunInit_WizardAuditSelectionPersists`
→ `git log -S` 로 `17b52894d` (t583 M3+M4, "cut the init wizard to four questions") 에서
제거된 것을 확인. 위저드 질문 16→4 축소로 audit 질문이 사라진 결과다.

순증 +2 이며, **귀속되지 않은 잔여는 0** 이다.

## 게이트가 걱정한 지점 — golden 4건

`TestUpdateVersionDowngradeConfirm_Localized` 의 네 하위 케이스가 모두 PASS:
`a-project-ja-profile-ko`, `b-no-project-profile-ko`, `c-no-project-no-profile`,
`d-project-without-lang-profile-zh`. 옵션 모양 쪽도 `TestProfileOptions_ValueSetsEqualSchema`,
`TestProfileOptions_Labels` PASS.

즉 t583 의 `WizardResult` 필드 삭제와 질문 축소는 이 카드가 고정한 golden·옵션 모양을
깨지 않았다.

## 아직 세우지 않은 것

- `internal/cli` 전체가 아니라 이 선택자 범위만 돌렸다. 전수 판정은 CI 몫이다.
- t621(`origin/develop` `8ba7ab4bb`)은 **흡수하지 않은 상태**에서 측정했다. t621 의 변경
  9 파일 중 `internal/cli` 는 `todo_axisa_guard_test.go`, `todo_queue_root_test.go` 둘이고
  이 카드 파일 집합과 교집합 0, 선택자(`Profile|Wizard|HuhTheme|UpdateVersion|TUI`)에도
  걸리지 않는다 — 측정에 영향을 줄 수 없어 변수를 늘리지 않으려 미뤘다. 병합 창 직전에
  흡수한다.
- 부하 13.65 환경에서 잰 값이라 소요 시간(12.457s)은 머신 상태를 반영한다. 판정은
  통과/실패이지 시간이 아니다.

🗿 MoAI
