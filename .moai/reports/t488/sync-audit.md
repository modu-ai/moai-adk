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

---

# §9 델타 확인 — 수리 이후 (`3c0abee98` → `b3d0091f3`)

**위 §1-§H 본문은 손대지 않았다.** 그것은 `3c0abee98`에서 발견된 것의 기록이고, 발견이 수리된 뒤에 판정서를 고쳐 쓰면 그것은 더 이상 증거가 아니다. 이 절은 그 위에 얹는 델타 확인이며, 판정을 재산정하지 않는다.

감사 트리: `.claude/worktrees/t488` · 브랜치 `WT-premerge-drift-assert` · HEAD `b3d0091f3c6da5d855244e5dc613845f3fcffa4f` · 시작·종료 시점 `git status --porcelain` 무출력.
델타 범위: 커밋 6개(`65146f0f6`, `1a34c51ef`, `25069d992`, `318d098c4`, `955e5190c`, `b3d0091f3`), Go 파일 5개 + 문서.

## §9.1 F1 — 닫혔다. 직접 재발화로 확인

수리는 세 줄이 아니었고, 그 편이 옳다. `preserveSettingsDriftCopy`가 **경로가 아니라 바이트를 받는다**(`settings_drift.go:412`). 경로 인자는 두 번째 읽기를 *허용*하지만 바이트 인자는 그것을 표현할 수 없다 — 방어를 주석이 아니라 시그니처에 넣은 것이다.

**두-읽기 모양을 직접 복원해 RED를 관측했다.** `AssessSettingsDrift`의 preserve 호출에 신선한 `os.ReadFile(result.Path)`를 끼워 넣었더니, 먼저 **컴파일이 거부했다**:

```
internal/kanban/settings_drift.go:296:2: declared and not used: data
FAIL	github.com/modu-ai/moai-adk/internal/kanban [build failed]
```

이것은 기록되지 않은 성질이고, 기록할 값이 있다. 두-읽기 회귀를 심으려면 **한 번 읽은 버퍼를 명시적으로 버려야** 한다(`_ = data`). 즉 테스트가 잡기 전에 타입 시스템이 먼저 막는다. `_ = data`로 뮤턴트를 완성한 뒤의 RED:

```
=== RUN   TestLedgerDigestDescribesThePreservedBytes
    settings_drift_test.go:780: the ledger digest does not describe the preserved bytes:
         ledger:    8d39bb027e9dc2f980612460729cf754be386c725ff0a300baa217faa1859028
         preserved: 92733cfa5e922781efae8739382ca015c9a34dc6b84a6783b8d0180e50eb5652
    settings_drift_test.go:784: ledger size_bytes 22 does not match the preserved copy's 38 bytes
    settings_drift_test.go:789: preserved copy is not the content that was measured
--- FAIL: TestLedgerDigestDescribesThePreservedBytes (1.17s)
```

`f1-double-read.txt`의 두 다이제스트와 **바이트 단위로 같다**. 기록된 증거가 재현 가능하다.
복원: `shasum -a 256 internal/kanban/settings_drift.go` = `990ef494cffa4b15ad63acb53b4bc8525bd2b088aae8b9c69f1434f1bb4190b2`(뮤턴트 이전 값과 동일), `git status --porcelain` 무출력.

**이음매(seam)는 건전하다.**

- **프로덕션 nil 보장 — 읽어서가 아니라 훑어서.** `grep -rn "settingsDriftPreserveTestHook" --include='*.go' .`의 대입은 두 곳뿐이고 **둘 다 `_test.go`**다(`settings_drift_test.go:743` 대입, `:749` `t.Cleanup`으로 nil 복원). 비-테스트 파일의 대입 0건. 호출 지점(`settings_drift.go:306`)은 nil 가드가 있고, Go에서 nil `func()` 호출은 패닉이므로 그 가드가 "nil이면 동작이 바이트 동일"을 의도가 아닌 사실로 만든다.
- **선례는 실재한다.** `integrationLockMutationTestHook`이 같은 패키지의 `integration_lock.go:83`(주석)·`:95`(선언)·`:254`(nil 가드 호출)에 있다. 이음매를 발명한 것이 아니라 옆집 것을 그대로 따랐다.
- **패키지 전역 변수인데 병렬 오염이 없는가 — 추론이 아니라 실측으로 답했다.** `go test ./internal/kanban/ -count=1 -race` → **`ok … 147.594s`**. race 보고 0건, 실패 0건. Go의 병렬 테스트 스케줄링을 근거로 "안전할 것"이라 논증할 수도 있었지만, 그것은 추론이다. 잰 결과가 깨끗하다.

**대조 세 개는 옳은 세 개다.** ① 두 내용이 실제로 다름(`:738`) — 없으면 단정이 어느 쪽이든 참이 된다. ② 훅이 **정확히 1회** 발화(`:759`) — `!= 1`이라 미발화와 중복발화를 **둘 다** 잡는다(`> 0`이었다면 후자를 놓쳤다). ③ 끼어든 쓰기가 실제로 착지(`:762-766`) — 훅이 돌았어도 쓰기가 실패했으면 단정은 공허하다. 세 개가 서로 다른 공허 경로를 막으며 겹치지 않는다.

**판정: 결함을 닫았다. 이전(relocate)이 아니다.**

## §9.2 F4 — 닫혔다. 순서도 성립한다

두 표면, 두 sentinel(`integration_settings_drift.go:55-63`). `preflight`은 `errSettingsDriftDetected`를 반환한다.

**순서 주장은 실측으로 성립한다.** `integration_settings_drift_report_test.go`에서 정체성 단정이 `:174`(자기 sentinel)·`:177`(acquire sentinel 아님), 문구 단정이 `:186`·`:189` — 구조가 먼저, 문구가 나중이다. 두 sentinel은 서로 다른 `errors.New` 값이므로 `errors.Is`가 동일성 비교로 갈라낸다. 양방향(`preflight`은 거절을 주장하지 않고 / `acquire`는 거절을 주장한다)을 한 테스트에서 함께 잰 것도 옳다 — 한쪽만 쟀다면 acquire가 조용해지는 다른 결함이 이 테스트를 통과했을 것이다.

**문구 단정은 남길 값이 있다 — 축이 다르기 때문이다.** sentinel 단정은 **정체성**을 고정하고, 문구 단정은 **내용**을 고정한다. `errSettingsDriftDetected`의 메시지를 나중에 누군가 "refused"를 담은 문장으로 고쳐도 정체성 단정은 여전히 통과하고, 그때 잡는 것은 문구 단정뿐이다. 중복이 아니라 다른 회귀를 막는다. 게다가 테스트가 자기 한계를 주석에 적어 두었다("다르게 표현된 거짓 주장은 여전히 통과한다") — 근사임을 알고 쓰는 근사는 과잉 주장이 아니다. **남기기를 권고한다.**

명사가 아니라 주장을 금지하도록 좁힌 것도 옳다. `no integration window was taken`은 정직한 설명이지 거짓 주장이 아니며, 명사 `window`를 금지했다면 정직한 문장이 테스트에 막혔을 것이다.

## §9.3 F5 / F2 / F3

- **F5** — `integration.go`가 스윕 목록에 들어갔다(`integration_settings_drift_exitcode_test.go:30-33`). 이번 실행에서 스윕 테스트 통과. §1에서 이미 살아 있는 결함 0을 세워 두었으므로 커버리지 간극만 닫혔다.
- **F2** — 리드 목록에 없었으나 **가장 철저히** 닫혔다. 재생성한 인용이 현재 트리에서 전부 해석된다(실측: `settings_drift_test.go:537/780/784/789`, `integration_settings_drift_report_test.go:174/177/186/189` 모두 해당 단정 줄). 재생성된 `m6-executor-bypassed.txt`가 **537행을 인용**하는데, 이는 §M에서 내가 독립적으로 재발화했을 때 발화한 바로 그 행이다. 여기에 `mutants/README.md`가 각 항목에 **줄 번호와 별개로 안정 앵커**(테스트 함수명 + 단정 문구)를 주고, 이전 세대 로그가 왜 틀렸는지(+14/+16)를 지우지 않고 기록했다. 요구한 것보다 나은 처분이다.
- **F3** — `CHANGELOG.md`의 `A new read-only verb` → `A new verb`. 닫혔다.

## §9.4 수리가 새로 만든 것

코드 쪽은 없다. 문서 쪽에 하나 있고, 하나는 자기충족성 간극이다.

- **G1** [Low] [**blocking — 창 진입 전 정정 권고**] `CHANGELOG.md:12` 말미가 여전히 `backfilled to 199d2777be085033a96d89cc45466d11ef4a37b9`라고 적는다. 그런데 `progress.md` §E.4는 그 값을 **superseded, invalidated**로 명시하고 `955e5190c86c88605efcaaf41bd002005516ed51`을 sync 커밋으로 기록한다. 실측: `grep -c '199d2777b' CHANGELOG.md` → 1, `grep -c '955e5190c' CHANGELOG.md` → 0. **같은 카드의 두 커밋된 기록이 "무엇이 이 카드를 닫았는가"에 대해 서로 다른 답을 하고, 한쪽이 다른 쪽의 값을 무효라고 부른다.** 재마감이 §E.4는 갱신했으나 그것을 반사하는 CHANGELOG 문장은 갱신하지 않았다. 이 항목이 이미 세 번 겪은 계열("이미 움직인 상태를 기술한 문장")의 **네 번째**다. SHA 하나 바꾸면 닫힌다.
- **G2** [Low] [optional] §E.4의 재측정이 `HEAD 318d098c4` 기준이라고 적혀 있으나 현재 HEAD는 `b3d0091f3`이고, 그 사이 두 커밋이 코드를 건드리지 않았다는 **증거를 아티팩트가 스스로 담지 않는다**. 내가 재서 확인했다 — `git diff --name-only 318d098c4..HEAD -- '*.go'` 무출력, `git diff --stat 318d098c4..HEAD`는 `progress.md` 한 파일뿐. 거짓 주장이 아니라 독자 자기충족성의 간극이다(리드 질문 3의 답이 여기에 걸린다). 한 줄 추가로 닫힌다.
- **G3** [Very Low] [optional] §E.4 「Template-First check」가 `git diff --stat 256b30fa5..HEAD`를 인용하는데 `HEAD`는 움직이는 앵커다. 원래도 그랬으나 HEAD가 6커밋 더 가면서 오해 소지가 커졌다. 미러 동일성 자체는 이번에도 재확인했다(`diff -q` 두 쌍 exit 0).

**회귀 없음 — 잰 결과.** `go build ./...` rc=0 · `go test ./internal/kanban/ -count=1 -race` → `ok … 147.594s` · `go test ./internal/cli/ -run 'SettingsDrift|Acquire|Preflight|ExitCode' -count=1` → `ok … 25.298s` · `go vet` 무출력 rc=0 · `gofmt -l internal/{kanban,cli,config}/` 무출력 · `golangci-lint run --timeout=3m` → `0 issues.` · 감시 대상 여전히 단일(`grep -c 'settings.local.json'` → 두 파일 모두 0) · 미러 두 쌍 바이트 동일.

## §9.5 §E.4 재측정의 건전성 (리드 질문 3)

**건전하다.** 세 가지가 옳게 되어 있다.

1. **소비가 아니라 재실행을 골랐다.** 두 수리 커밋 어느 쪽도 전 패키지 실행을 인용하지 않았으므로, 소비할 귀속 가능한 증거가 애초에 없었다. `agent-common-protocol.md`의 attributable diff-check는 세 조건이 모두 맞을 때만 소비를 허용하고 어긋나면 재실행하라고 말한다 — 어긋났고, 재실행했다. 규칙대로다.
2. **폐기된 값을 지우지 않고 남겼다.** `sync_commit_sha` 줄에 `Superseded prior value: 199d2777b…`가 인라인으로 붙어 있다. 조용히 덮어쓰면 재마감이 있었다는 사실 자체가 사라진다.
3. **순서 결함을 사실로 진술했다.** "두 write-capable 에이전트를 한 트리에 배차했다"를 명시하고, 구조적 수정("sync 마감은 마지막에 간다")까지 적었다. 비난이 아니라 사실 진술이며, 재현을 막는 형태다.

독자 자기충족성은 **G2 하나를 제외하면** 성립한다 — 인용된 명령과 그 출력이 그대로 있어 누구도 물어볼 필요가 없다.

## §9.6 원래 F4 판정에 붙이는 정정 (리드 질문 4)

**재산정이 아니라, 논거 하나를 정정한다.** §F의 F4 항목에 나는 severity Low를 주면서 그 근거로 "이것은 기록이 아니라 에러 문자열"이라는 구별을 썼다. **그 구별은 성립하지 않는다.** REQ-PSD-009가 금지하는 것은 *일어나지 않은 행위를 주장하는 것*이고, 그 금지는 주장을 나르는 매체가 JSON 필드인지 사람이 읽는 문장인지를 구별하지 않는다. 오히려 사람이 읽는 표면 쪽이 스키마의 제약을 받지 않아 더 쉽게 거짓말한다.

내 원문은 같은 문단에서 "이 카드가 lock 레코드에 대해 [HARD]로 금지한 모양이 한 층 위에서 재현된 것"이라고 **정확히 관측해 놓고도**, 그 관측이 severity를 끌어올리게 두지 않았다. 리드의 승격이 옳았다.

§7의 종합 판정과 차원 점수는 그대로 둔다 — 그것은 `3c0abee98`에 대한 판정이고, 이 정정은 그 옆에 선다.

## §9.7 알려진 red (리드 질문 5)

**아무것도 움직이지 않았다.** `b3d0091f3`에서 이번에 실행:

```
token_budget_guard_test.go:69: always-loaded surface = 77801 tokens (budget 77600, headroom -201, 17 entries)
```

§1에서 `3c0abee98`에 대해 잰 값과 **표면·예산·초과분 세 수치가 전부 동일**하다. 델타 6커밋 중 `.claude/` 아래를 건드린 것이 없으므로 예상과 일치한다. `AlwaysLoadedTokenBudget`은 여전히 손대지 않았고, 인상을 권고하지 않는다는 §B의 입장도 그대로다.

## §9.8 창 진입 준비 상태

**G1을 정정하면 준비됐다.** 그 외 모든 축이 정리됐다 — 코드 회귀 0(race 포함), 세 수리 모두 제기한 결함을 닫았고 이전시키지 않았으며, F1은 타입 시그니처가 회귀를 컴파일 단계에서 막는 데까지 갔고, F2는 요구 이상으로 닫혔고, §E.4는 재측정이 건전하며 순서 결함을 숨기지 않았다.

G1은 SHA 하나짜리 편집이지만 **병합 전에 하는 편이 옳다**: 같은 카드의 두 커밋된 기록이 서로 모순하며 한쪽이 다른 쪽을 무효라고 부르는 상태이고, `done` 이후에는 그것을 고칠 계기가 생기지 않는다. 이 판정은 그 정정을 **권고**하며, 창을 잡을지는 리드의 결정이다.

## §9.9 이번 델타 라운드에서 보지 않은 것

- **커버리지 재측정.** `internal/kanban` 86.3% / `internal/cli` 80.6% / 신규 파일 83.2% · 90.8%는 §1에서 `3c0abee98`에 대해 실측했고, 델타 이후 **다시 재지 않았다**. 훅 분기와 새 테스트가 더해졌으므로 수치는 소폭 움직였을 수 있다. 미검증으로 남긴다.
- **`internal/cli` 전 하위 패키지 실행.** 이번에는 drift·acquire·preflight·exitcode 필터로만 돌렸다(25.3s). §E.4가 인용한 전체 실행(389s)은 재현하지 않았고, 그 인용을 소비하지도 않았다 — 내가 잰 것은 필터된 범위뿐이다.
- **크로스 플랫폼 빌드**(`GOOS=windows`/`linux`) — 델타에서도 재실행하지 않았다.
- **뮤턴트 M1-M5·M7의 재발화.** 재생성된 로그의 인용 줄이 해석되는 것은 확인했으나, 뒤집어 본 것은 이번에도 M6(§M)과 F1(§9.1) 둘뿐이다.
- **`955e5190c` 재마감 커밋 자체의 문서 diff 전문.** §E.4 최종 상태를 읽었을 뿐, 재마감이 §E.2·§E.3에 가한 편집을 줄 단위로 대조하지는 않았다.
- **G1의 파급 범위.** CHANGELOG의 그 문장 하나만 확인했고, 다른 아티팩트가 `199d2777b`를 sync 커밋으로 인용하는지는 전수 훑지 않았다.

---

델타 확인 소유권: 이 §9는 sync-auditor의 것이다. 위 모든 수치는 HEAD `b3d0091f3`의 트리에 대해 이번 세션에서 실행한 명령의 출력이며, 건네받은 값을 옮긴 것은 없다.
