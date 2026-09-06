# sync-audit.md — SPEC-PREMERGE-SETTINGS-DRIFT-001 (카드 t488)

감사 트리: `.claude/worktrees/t488` · 브랜치 `WT-premerge-drift-assert` · HEAD `3c0abee98a217c2fc4dde98737a38bad7facf71d` · 감사 시작·종료 시점 모두 `git status --porcelain` 무출력.
감사자: sync-auditor (독립 판정). 아래 모든 수치는 **이 트리에서 이번에 실행한 명령의 출력**이며, 건네받은 값을 그대로 옮긴 것은 하나도 없다.

## 종합 판정

**PASS** — 종합 90.6 / Tier M 임계 0.80. must-pass 두 축(Functionality, Security) 모두 독립 통과.
blocking 결함 0건. optional 결함 7건(아래 §F). 그중 F3 하나는 병합 전 정정을 권고한다 — 비용이 없고, 리드가 요청한 "세 번째 CHANGELOG 부정확"이 정확히 그것이다.

### 차원 점수

| 차원 | 점수 | 판정 | 증거(이번 실행) |
|---|---|---|---|
| Functionality (40%) | 92/100 | PASS | `go test ./internal/cli/ -run 'SettingsDrift\|Acquire\|Preflight' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 23.372s` · `go test ./internal/kanban/ -run 'SettingsDrift\|Preserve\|Ledger' -count=1` → `ok … 2.173s` · `go test ./internal/config/ -run SettingsDriftGate -v` → `--- PASS` ×3 · M6 뮤턴트 직접 재발화(아래 §M) |
| Security (25%) | 95/100 | PASS | 보존 파일 `0o600`·원장 `0o600`(`settings_drift.go:402,433`), 출력·원장 어디에도 파일 내용 없음(`settings_drift.go:224-236`, `integration_settings_drift.go:100-111`), argv 전량 고정 리터럴(`settings_drift.go:148-157`), 라벨 sanitize로 경로 이탈 차단(`settings_drift.go:354-365` + `TestSettingsDriftLabelFallsBackWithoutBreakingThePath`), sha256(md5 아님), 자동 복원 경로 부재 |
| Craft (20%) | 84/100 | PASS | `go test ./internal/kanban/ -count=1 -coverprofile` → `ok … 137.055s coverage: 86.3%` · `go test ./internal/cli/ -count=1 -coverprofile` → `ok … 343.085s coverage: 80.6%` · `go vet` rc=0 무출력 · `gofmt -l` 무출력 · `golangci-lint run ./internal/cli/... ./internal/kanban/... ./internal/config/...` → `0 issues.` |
| Consistency (15%) | 88/100 | PASS | `diff -q` 미러 2쌍 전부 exit 0 · `NewDefaultWorkflowConfig`의 `SettingsDriftGate{Enabled:false}`가 형제 `IntegrationLock`/`AgentModelGuard`와 동형 · `configCacheSchemaVersion 2→3` 범프(과거 `AgentStopGuard` 사고의 교훈을 실제로 적용) |

가중합 = 92×0.40 + 95×0.25 + 84×0.20 + 88×0.15 = **90.55**.

## §M 뮤턴트 직접 재발화 — M6(게이트 경로 실행기 우회)

기록된 뮤턴트 로그를 읽는 것만으로는 이 카드가 막으려는 결함을 감사자가 그대로 재생산한다. 그래서 6종 중 **주장이 가장 강한 M6**을 골라 직접 뒤집었다. M6만이 "포착기가 단 하나"라고 주장하고((d-1)/(d-2)는 통과, (d-3)만 발화), 그 주장이 이 카드의 중심 논지이기 때문이다.

적용한 뮤턴트 — `internal/kanban/settings_drift.go`의 `AssessSettingsDrift`에서 원장 append 직전에, 기록 러너를 우회해 `os/exec`로 primary 루트에 직접 커밋:

```go
if preserved != "" {
    add := exec.Command("git", "add", "-A"); add.Dir = p.Root; _ = add.Run()
    com := exec.Command("git", "commit", "-m", "mutant: commit the preserved copy"); com.Dir = p.Root; _ = com.Run()
}
```

관측한 RED(`go test ./internal/kanban/ -run TestAssessSettingsDriftPreservesAndLedgers -count=1 -v`):

```
=== RUN   TestAssessSettingsDriftPreservesAndLedgers
    settings_drift_test.go:537: primary-root HEAD moved: "5ff80a603423b320b12a0063e21a80f8c44b0fbd" -> "00b4c3b0e995c7e6d4277ba9c66b5dfe2a30f535" (a committed preserved copy lands here, not in the target tree)
--- FAIL: TestAssessSettingsDriftPreservesAndLedgers (1.72s)
FAIL	github.com/modu-ai/moai-adk/internal/kanban	2.190s
```

두 가지가 실측으로 확인됐다. **첫째**, RED는 진짜다 — 뮤턴트가 실제로 뒤집힌다. **둘째**, `progress.md` §E.2의 귀속 주장도 맞다: 실패 줄이 **하나뿐**이고, 러너 기록을 읽는 (d-1)/(d-2)는 조용히 통과했다. 기록을 읽지 않는 단정만이 이 우회를 본다. 두-루트를 재라는 `AC-PSD-007(d-3)`의 요구가 장식이 아니라 필요조건이라는 뜻이다.

복원: 사전 백업본을 되돌린 뒤 `shasum -a 256 internal/kanban/settings_drift.go` = `bf0f659fda9fdc0b220d4a419a1d6538e3d396d8f5459b441da46209a5d37935`(뮤턴트 적용 전 값과 동일), `git status --porcelain` 무출력, `git rev-parse HEAD` = `3c0abee98…`. 트리는 감사 전과 바이트 동일하다.

## §A AC-PSD-001..013 개별 판정

| AC | 관측 가능한 반증이 있는가 | 판정 | 이번 감사에서 본 것 |
|---|---|---|---|
| 001 적중 | 예 — `count != 1`이 이분 판정 | PASS | `settings_drift_test.go:158-159`. 기록 로그 `m1`이 `got 0, want 1`로 뒤집힘 |
| 002 통과 | 예 — 양성 대조(`assertPredicateRan`) 위의 0 | PASS | `settings_drift_test.go:178` 대조 → 0 단정. 대조 없는 0이 아니다 |
| 003 경로 오지정 | 예 | PASS | `settings_drift_test.go:200` 대조 후 0 |
| 004 경로 누락 | 예 — 적극 단정(1)이라 대조 불요 | PASS | `settings_drift_test.go:220-222` |
| 005 argv 고정 | 예 — 실행 경계 기록 문자열 | PASS | `settings_drift_test.go:239-277`. 빌더(`settingsDriftPredicateArgv`)는 unexported이고 단정은 `runner.Commands()`를 읽는다 — 요구된 형태 그대로 |
| 006 종료 코드 미의존 | 예 — 대조 있는 스윕 | PASS | 스윕 자체를 넘어 **카드가 손댄 Go 파일 12개 전량**을 독립 훑음: `ExitCode`/`ExitError`/`$?` 히트는 스윕 파일 자신의 3줄뿐(§F5 참조) |
| 007 (a) 보존 | 예 — 바이트 비교 + 원본 비공백 대조 | PASS | `settings_drift_test.go:487-491` |
| 007 (b) 원장 | 예 | PASS | `settings_drift_test.go:493-516`, 6개 필드 전부 단정 |
| 007 (c) 충돌 접미 | 예 | PASS | `settings_drift_test.go:557-583`. `O_EXCL` 루프(`settings_drift.go:396-417`)가 결정성을 만든다 |
| 007 (d-1)(d-2) | 예 | PASS | `settings_drift_test.go:518-527` |
| 007 (d-3) | 예 — 내용 대조 3종 후 동일성 | PASS | **직접 재발화로 확인**(§M) |
| 007 (d-4) | 부분 — 대조가 "실행됨"까지만 | PASS | `settings_drift_test.go:545-552`. 한계가 코드 주석과 acceptance.md 양쪽에 명시돼 있다. 좁은 예외의 정당한 사용 |
| 008 원본 불변 | 예 — 기록 무관 단정 2 + 기록 단정 1 | PASS | `settings_drift_test.go:586-627` |
| 009 거절이 창을 안 잡음 | 예 — lock 파일 부재 + 사전 lock 바이트 동일 | PASS | `integration_settings_drift_test.go:169-234`. 코드상 거절이 `resolveIntegrationTarget` **앞**에서 반환된다(`integration.go` acquire 블록) |
| 010 우회 기록 / --force 분리 | 예 | PASS | `integration_settings_drift_test.go:260-300`. `acquireSettingsDriftPrecondition`는 `force`를 읽지 않는다(`integration_settings_drift.go:246-271`) |
| 011 실패 ≠ 통과 | 예 — `status == "undetermined"` 적극 단정 + `match_count` 키 부재 | PASS | `integration_settings_drift_test.go:322-334` |
| 012 보존 실패가 판정을 안 뒤집음 | 예 | PASS | `integration_settings_drift_test.go:355-369`. 판정층(`status`)에 걸어 기본 자세와 독립시킨 선택이 옳다 |
| 013 (a)-(e) 기본 자세 | 예 | PASS | `integration_settings_drift_test.go:421-503`. workflow.yaml을 **아예 두지 않은** 루트로 "키 부재 ≠ false 명시"까지 구별 |

13개 전부가 실행된 명령의 출력으로 이분 판정되며, **요약을 신뢰할 것을 요구하는 항목은 없다**. 종료 코드에 걸린 AC는 하나도 없다(독립 스윕으로 확인).

## §B 리드가 지목한 항목별 판정

**부재 단정 규율(`acceptance.md:7`)을 전달된 테스트가 실제로 지키는가 — 지킨다.** 절대적 부재/동일성 단정 11곳을 전수 대조했고, 전부 앞에 내용 기반 양성 대조가 서 있다: `assertPredicateRan`(002·003·007d-1·008·`CleanTree`), 40자 SHA + 두 루트 상이 + 비공백 `ls-remote`(007 d-3), 64자 hex + 비공백 파일 목록(008), 비공백 refusal 출력 + sha256 포함(009), 비공백 사전 lock + 보유자 이름 포함(009 변형), 비공백 원본(007a·013b), 파일 존재+비공백+구현 심볼(006). 내용 대조가 불가능한 곳은 `(d-4)` 하나뿐이고, 그 한계가 단정 옆에 적혀 있다 — 규율이 허용한 좁은 예외의 정확한 사용이다.

**알려진 red의 귀속 — 산술과 귀속 모두 성립한다.** 이번 실행에서 `go test ./internal/config/ -run TestAlwaysLoadedTokenBudget` → `surface = 77801 tokens (budget 77600, headroom -201)`. 기저 커밋 `256b30fa5`의 77723은 `budget-baseline-before-m6.txt`에 남아 있고, 77801 − 77723 = **78**, 잔여 **123은 상속분**이다. `git diff --numstat 256b30fa5..HEAD -- .claude/ CLAUDE.md AGENTS.md`는 두 파일만 낸다 — `kanban-dispatch.md` +2, `kanban-dispatch-detail.md` +27/−1. 후자는 프런트매터에 `paths: "**/kanban-dispatch*.md,…"`가 있어 always-loaded 표면이 아니다(직접 확인). 즉 표면 증가분의 유일한 출처는 stub의 [HARD] 두 줄이다. **`AlwaysLoadedTokenBudget`을 올리지 않은 판단에 동의한다** — 만들지 않은 123을 흡수하면 그 초과가 이 카드 안에서 사라져 보인다. 상수 인상을 권고하지 않는다.

**소비된(재실행하지 않은) run-phase 증거의 귀속 — 건전하고, 독자가 혼자 검증할 수 있다.** `git diff --name-only 63e8e900d..HEAD`는 `progress.md`·`spec.md`·`CHANGELOG.md` 셋뿐이며 `.go` 0개다(이번 실행으로 독립 확인). §E.4가 그 `--stat` 출력을 인용해 두었으므로, 뒤에 읽는 사람은 커밋 두 개와 명령 하나로 같은 결론에 도달한다 — 물어볼 필요가 없다. 다만 소비 판단이 옳다는 것과 **증거가 그 커밋에 귀속된다는 것은 별개**이고, 후자에 한 건의 결함이 있다(§F2).

**잔여 위험 ⑤(`--allow-settings-drift` + `--force` 동시 지정)를 이름으로 남긴 처분 — 옳다.** 두 값은 코드상 독립이다: `acquireSettingsDriftPrecondition`(`integration_settings_drift.go:246-271`)은 `force`를 인자로 받지도 않고, `force`는 `AcquireIntegrationLock`에만 전달된다(`integration.go`). §E.2가 "읽어서 안 것이지 재서 안 것이 아니다"라고 적은 것은 정확한 자기 기술이며, 증거 제출 뒤에 테스트를 덧붙여 목록을 짧게 만드는 것보다 정직하다. **막을 사유가 아니다.** 다만 후속에서 12줄짜리 테스트로 닫기를 권고한다 — REQ-PSD-010이 고정하는 두-축 분리의 자연스러운 마지막 칸이다.

**감시 대상 범위 — `.claude/settings.json` 하나 그대로다.** `SettingsDriftWatchedPath` 단일 상수가 술어 argv와 보고 경로 양쪽의 유일 원천이고(`settings_drift.go:48,149,270`), `.claude/settings.local.json`이나 `.moai/config/**`를 참조하는 코드 경로는 없다(grep 무히트). 범위 확대 0.

**Template-First — 재확인했다.** `diff -q` 두 쌍 모두 exit 0(로컬 ↔ `internal/template/templates/` 미러). 템플릿 쪽에 `enabled: true`가 새로 들어간 곳은 없고, 새 설정 키는 doctrine 산문 1곳에만 등장한다(중립성 유지). `CLAUDE.local.md`는 문서화된 로컬 전용이라 미러 의무가 없다 — 정확한 판단.

**docs-site 미확장 판단 — 근거가 실측으로 성립한다.** `grep -rl "moai integration" docs-site/content` rc=1 무출력. 기존 세 동사 어느 것도 문서화돼 있지 않으므로 새 동사만 문서화하면 오히려 비대칭이 생긴다. 동의한다.

## §F 결함 목록

전부 confidence·severity를 함께 적었고, **blocking 판정은 0건**이다(어느 것도 AC를 깨거나 SPEC이 명시한 요구를 위반하지 않는다).

- **F1** [Medium] [optional] `internal/kanban/settings_drift.go:286` + `:382` — **원장의 sha256/size와 보존 사본이 서로 다른 바이트를 가리킬 수 있다.** `hashSettingsDriftSource`가 `os.ReadFile`로 한 번 읽어 해시를 내고, `preserveSettingsDriftCopy`가 **같은 파일을 다시 읽어** 복사한다. 두 읽기 사이에 런타임(`moai glm`/`moai cc`/SessionStart 훅)이 그 파일을 쓰면, 원장 행은 보존 사본이 아닌 바이트를 기술한다. 이 상태는 도달 불가능한 가정이 아니라 **이 카드의 전제 그 자체**다 — REQ-PSD-006이 자동 복원을 금지한 이유가 바로 "런타임이 이 파일을 쓴다"이다. 픽스처는 정적이라 어떤 테스트도 이 창을 보지 못한다. 필요한 수정: 한 번 읽은 바이트를 `preserveSettingsDriftCopy`에 넘겨 해시·크기·사본이 같은 관측에서 나오게 한다(대략 3줄).
- **F2** [Low-Medium] [optional] `.moai/reports/t488/mutants/m1,m3,m4,m5,m6*.txt` — **인용한 줄 번호가 전달된 트리의 어느 커밋에서도 해석되지 않는다.** 실측: `m1`이 인용한 `settings_drift_test.go:145`는 현재 트리에서 빈 줄이고, 해당 단정은 159행에 있다(오프셋 +14). `m6`이 인용한 521행은 현재 `for _, c := range runner.Commands()`이며, 내가 같은 뮤턴트를 재발화했을 때 발화 지점은 **537행**이었다(오프셋 +16). `internal/kanban/settings_drift_test.go`는 `63e8e900d` 이후 변경 0이므로(§B), 이 로그들은 **커밋되지 않은 중간 버전**에 대해 캡처된 뒤 갱신되지 않았다. RED 자체는 진짜다(직접 확인) — 결함은 관측이 아니라 **앵커**에 있다. `internal/cli` 쪽 인용(`:179`, `:78`)은 정확히 해석된다. 권고: `progress.md` §E.2의 뮤턴트 표에 "인용 줄 번호는 캡처 시점 기준이며 단정 문구가 안정 식별자다"를 한 줄 남기거나, 로그를 재캡처한다.
- **F3** [Low] [optional — 단, 병합 전 정정 권고] `CHANGELOG.md:12` — **`moai integration preflight`을 "a new read-only verb"라고 적었으나 읽기 전용이 아니다.** 적중 시 보존 사본을 쓰고 원장 행을 append한다. 구현 자신의 doc comment가 정반대로 말한다: `integration_settings_drift.go:160` — *"It is not purely read-only: a hit preserves and appends a ledger row"*. 리드가 요청한 "세 번째 부정확"이 이것이다. 앞의 두 건과 같은 계열 — 이미 움직인 상태를 기술한 문장. 두 단어 수정으로 닫힌다.
- **F4** [Low] [optional] `internal/cli/integration_settings_drift.go:211` — **`preflight`이 창을 잡지 않는데 "release-integration window refused"를 낸다.** `errSettingsDriftRefused`(`:48`)를 `acquire`와 공유하기 때문이다. 사용자가 보는 문장이 일어나지 않은 일을 기술한다 — 이 카드가 lock 레코드에 대해 [HARD]로 금지한 모양("우회할 거절이 없는데 우회로 적으면 기록이 거짓말을 한다")과 같은 형태가 한 층 위에서 재현된 것이다. 보고 본문(`settings drift: DRIFT …`)은 정확하므로 피해는 제한적이다. 권고: `preflight` 전용 sentinel 분리.
- **F5** [Low] [optional] `internal/cli/integration_settings_drift_exitcode_test.go:26-35` — **스윕 대상 목록이 `internal/cli/integration.go`를 담지 않는다.** 거절을 호출자에게 전파하는 acquire 배선이 그 파일에 있다. 이번 감사에서 카드가 손댄 Go 파일 12개 전량에 대해 독립 스윕을 돌린 결과 히트는 스윕 파일 자신의 3줄뿐이었으므로 **현재 살아 있는 결함은 없다** — 순수한 커버리지 간극이다. 권고: 목록에 `integration.go` 추가(1줄).
- **F6** [Very Low] [optional] `internal/cli/integration_settings_drift.go:135` — `if r.SizeBytes > 0`이라 **0바이트 dirty 파일에서 `size_bytes` 키가 사라진다.** 같은 상황에서 `sha256`은 (빈 입력의 해시라) 남으므로 JSON이 비대칭해진다. 소비자가 아직 없어 유계다.
- **F7** [Low] [optional] `--allow-settings-drift` + `--force` 동시 경로 무검증(§B에서 처분에 동의). 후속 카드 권고.

## §G 이번 감사에서 **보지 않은 것**

- **전체 테스트 스위트와 CI 판정.** 카드 제약(§4)과 `CLAUDE.local.md` §4/§6에 따라 손댄 3개 패키지로 범위를 한정했다. 전 패키지 판정은 CI 몫이며, 이 보고서는 그것을 주장하지 않는다.
- **크로스 플랫폼 빌드.** §E.3의 `GOOS=windows`/`GOOS=linux` 통과 주장은 **재실행하지 않았다**. 미검증으로 남긴다.
- **뮤턴트 M1·M2·M3·M4·M5·M7.** 기록된 출력을 읽었을 뿐 직접 뒤집지 않았다. 재발화한 것은 M6 하나다. 다섯 건의 RED는 여전히 기록에 대한 신뢰 위에 있다(다만 F2의 앵커 결함이 그 기록에 걸려 있다).
- **plan.md 본문.** `spec.md`·`acceptance.md`·`progress.md`는 전문을 읽었으나 `plan.md`(215줄)와 4건의 plan-audit 판정서는 교차 참조 확인 수준에서만 열었다. plan 단계 자체의 재감사는 이 감사의 범위가 아니다.
- **워크트리 `t334`.** 운영자 판정에 따라 읽지도, 손대지도 않았다.
- **F1이 기술한 경합의 실증.** 코드 두 지점을 읽어 도출한 것이며, 실제로 창 안에서 파일을 바꿔 재현하지는 않았다. 추론이지 측정이 아니다 — 그 구별을 여기 남긴다.

## §H 잔여 위험

- 게이트는 **기본값에서 아무것도 막지 않는다**(`enabled: false`). 검출·보존·원장·보고만 돈다. 이 카드가 근거로 든 실패("9일 동안 아무도 안 봤다")에 대해서는 그것으로 충분하지만, 병합에 실리는 것을 **막는** 성질은 운영자가 플래그를 켜기 전까지 존재하지 않는다. 이 감사는 "켜야 한다"를 권고하지 않는다 — 계열 (가)는 운영자 결정이다.
- `undetermined`는 거절하지 않는다(fail-open). git이 없는 환경에서 이 게이트는 출력으로만 알린다. §E.2가 "조용한 실패가 아니라 시끄러운 무력화"라고 정확히 규정했다.
- 보존 사본에 정리 정책이 없다. 적중이 드물다는 전제가 깨지면(오탐 발생 시) 무한 누적된다.
- F1의 경합은 좁지만, 하필 **이 카드가 지키려는 기록**(원장) 위에 있다.

---

판정 소유권: 이 PASS는 sync-auditor의 것이며 위임되지 않았다. 위 모든 수치는 HEAD `3c0abee98`의 트리에 대해 이번 세션에서 실행한 명령의 출력이다.
