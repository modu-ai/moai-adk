# t900 판정서 — `buildEnvForLaunch` 처분

- 카드: t900 (Tier S, Class B — plan 생략, run→sync)
- 브랜치: `WT-dead-launch-env` (워크트리 `.claude/worktrees/t900`)
- 베이스: 로컬 `develop` 9dcbc3dbe 를 카드 브랜치로 병합 (`5ef4a052d`)
- 판정: **제거** (호출 시 실패하는 가드는 기각)

---

## 1. Claim

1. `buildEnvForLaunch` 는 `internal/cli` 에 프로덕션 호출자가 0이다.
2. 현행 effort 주입 경로(`--settings` 임시 파일의 `effortLevel`)는 살아 있고, Claude 를 호스팅하는 모든 launcher 를 실제로 덮는다.
3. 따라서 `buildEnvForLaunch` 를 제거해도 잃는 기능이 없다. "재배선 금지" 불변식의 기계적 담지자는 **다른 seam**(`buildEnvForClaudeLaunch` + `TestClaudeLaunchEnvPreservesInheritedEffort`)이며, 카드 지시대로 그쪽은 건드리지 않았다.
4. 제거로 함께 죽는 것은 이 함수와 그 전용 테스트 7건뿐이다.

## 2. Evidence

### E1 — 호출자 0 (전제 재측정)

`grep -rn 'buildEnvForLaunch' --include='*.go' .` 의 비테스트 적중 3행이 전부 정의와 주석이다:

```
internal/cli/launcher.go:1159    // buildEnvForLaunch returns an environment slice with …   (주석)
internal/cli/launcher.go:1171    func buildEnvForLaunch(effortLevel string, base []string)  (정의)
internal/cli/launch_chain.go:105 // for key (the buildEnvForLaunch pattern).                (주석)
```

행번호는 배차문에서 옮겨 온 값이 아니라 이번 측정에서 얻은 값이다(카드 [HARD] 좌표 재측정 준수 — 심볼로 찾았다). 나머지 적중은 전부 `_test.go`.

### E2 — 대체 경로가 실재한다 (주석이 아니라 코드로)

호출 사슬을 코드로 따라갔다:

```
internal/cli/launcher.go:219   extraArgs = appendCrossSessionSettings(root, profileName, extraArgs)
  └ internal/cli/crosssession_settings.go:88  payload := applyLaunchEffort(crossSessionSettingsPayload(root), profileName)
      └ internal/cli/launch_effort_settings.go:52  effort := resolveLaunchEffort(prefs.EffortLevel, prefs.ModelPolicy)
      └ internal/cli/launch_effort_settings.go:57  payload[effortSettingsKey] = effort      // "effortLevel"
  └ internal/cli/crosssession_settings.go:96  return append(args, settingsFlagLong, path)  // "--settings" <file>
```

`appendCrossSessionSettings` 는 `unifiedLaunchDefault` 본문 안에 있다 — cc / glm / gpt 세 launcher 가 공통으로 통과하는 깔때기임을 **주석이 아니라 호출 위치 자체**가 보인다. Claude 분기(`launcher.go:796`)는 `buildEnvForClaudeLaunch(os.Environ())` 로 env 를 통과시키기만 한다. effort 는 env 를 타지 않는다.

기존 테스트가 이 경로를 실제로 집는다(제거 **전** 기준선, 전부 PASS):

```
--- PASS: TestApplyLaunchEffort (0.00s)
      explicit_effort_level_lands_under_effortLevel
      model_policy_supplies_the_fallback
      both_empty_injects_nothing
      unreadable_profile_fails_open
      nil_payload_is_materialized
--- PASS: TestResolveLaunchEffort (0.00s)
--- PASS: TestClaudeLaunchEnvPreservesInheritedEffort (0.00s)
--- PASS: TestCrossSessionSettingsPayloadDefault / Full / IsolateOptIn / FiltersInvalidValues
ok  	github.com/modu-ai/moai-adk/internal/cli	0.801s
```

### E3 — 죽는 것과 사는 것

**죽는 것 (이번 제거로 사라짐)**

| 대상 | 위치 |
|---|---|
| `buildEnvForLaunch` 함수 + 독 코멘트 + `@MX:NOTE` | `internal/cli/launcher.go` −34행 |
| `TestBuildEnvForLaunch` (서브테스트 3) | `internal/cli/launcher_test.go` −54행 |
| `TestBuildEnvForLaunch_EmptyEffort` / `_AddsNewEntry` / `_ReplacesExisting` / `_PreservesOtherVars` | `internal/cli/mcp_doctor_coverage_test.go` −77행 |

**사는 것 (확인함 — 같이 죽지 않는다)**

| 대상 | 살아 있는 이유 |
|---|---|
| `config.EnvClaudeCodeEffortLevel` | `buildEnvForGLMLaunch` 가 inert 값 스트립에 쓴다 |
| `resolveLaunchEffort` | `applyLaunchEffort`(현행 경로) + `launcher.go:763` GLM 분기가 쓴다 |
| `buildEnvForClaudeLaunch` + `TestClaudeLaunchEnvPreservesInheritedEffort` | **카드 지시대로 미변경.** 불변식의 실제 담지자 |
| `replaceEnvValue` (`launch_chain.go`) | 별개 헬퍼. 주석의 dangling 참조만 수리 |

### E4 — 제거 후 재측정

```
go build ./...                                   exit=0
GOOS=windows GOARCH=amd64 go build ./...         exit=0
go vet ./internal/cli/                           exit=0
golangci-lint run --timeout=5m ./internal/cli/   exit=0   "0 issues."
go test -count=1 -run '<생존 가드 7건>' ./internal/cli/   exit=0
  --- PASS: TestCrossSessionSettingsPayloadDefault / Full / IsolateOptIn / FiltersInvalidValues
  --- PASS: TestApplyLaunchEffort
  --- PASS: TestResolveLaunchEffort
  --- PASS: TestClaudeLaunchEnvPreservesInheritedEffort
  ok  	github.com/modu-ai/moai-adk/internal/cli	0.828s
```

전 패키지 결과는 §5.

Go 소스에서 `buildEnvForLaunch` 잔존 적중 0. 남은 적중은 `CHANGELOG.md` 1행과 닫힌 SPEC 인수기록 8행뿐이며, 인수기록은 작성 시점의 트리를 서술하는 문서이므로 정책상 미수정으로 남겼다.

변경 규모:

```
 internal/cli/launch_chain.go             |  2 +-
 internal/cli/launcher.go                 | 34 --------
 internal/cli/launcher_test.go            | 54 ---------------
 internal/cli/mcp_doctor_coverage_test.go | 77 ----------------------
 4 files changed, 1 insertion(+), 166 deletions(-)
```

## 3. Baseline-attribution

모든 수치는 이 워크트리(`.claude/worktrees/t900`, 브랜치 `WT-dead-launch-env`, 로컬 develop 9dcbc3dbe 흡수 후)에서 이번 실행으로 측정했다. 다른 트리·다른 시점의 값을 옮겨 오지 않았다. 캐시 회피를 위해 모든 `go test` 는 `-count=1`. 환경 오염 회피를 위해 모든 검증은 단일 복합 호출 `unset <9개 레인 변수> && <명령>` 형태로 돌렸다.

## 4. 판정 근거 — 왜 가드가 아니라 제거인가

카드가 요구한 선택지는 「제거」 대 「호출 시 실패하는 가드」였다.

**가드를 기각한 이유 세 가지.**

1. **가드는 이미 다른 자리에 있고, 이 함수와는 무관하다.** 재배선 금지 불변식을 기계적으로 집는 것은 `buildEnvForClaudeLaunch` + `TestClaudeLaunchEnvPreservesInheritedEffort` 다. 그 seam 은 "env 를 손대지 않는다"를 단언하고, 누가 env 를 오염시키면 적색이 된다. `buildEnvForLaunch` 에 가드를 덧대도 같은 불변식을 두 자리에서 지키게 될 뿐, 새로 잡히는 것이 없다.
2. **panic 가드는 launcher 에 런타임 실패 경로를 새로 만든다.** launcher 는 `syscall.Exec` 로 프로세스를 갈아치우는 경로다. 여기에 "잘못 불리면 죽는다"를 심는 것은 컴파일 타임에 닫을 수 있는 문제를 런타임으로 미루는 거래다.
3. **제거가 더 강한 기계적 금지다.** 함수가 없으면 재배선은 "한 줄 호출"이 아니라 "함수를 새로 쓰는 일"이 된다 — 카드가 지적한 "주석과 `@MX:NOTE` 는 사람에게만 말을 건다"를 실제로 닫는 유일한 방법이다.

**속도는 근거가 아니다.** 제거가 빠르다는 점은 결정에 넣지 않았다.

## 5. 전 패키지 재측정 (`internal/cli`)

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR \
        MOAI_KANBAN_SETTINGS_INJECTED MOAI_LAUNCH_PROVIDER MOAI_PROFILE_LEASE_TOKEN \
        MOAI_PROJECT_DIR CLAUDE_PROJECT_DIR \
  && go test -count=1 -timeout 30m ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	1226.404s
exit=0
```

`(cached)` 가 아니라 1226.404s 의 실제 실행이다. 카드의 부하 지시대로 `internal/cli` 한 패키지로 좁혔고 전 패키지(`./...`)는 돌리지 않았다.

## 6. Gaps — 관측하지 않은 것

- `moai cc` / `moai glm` 실행 검증은 **하지 않았다.** CLAUDE.local.md §13 이 dev 프로젝트에서의 실행을 금지하므로, 대체 경로 확인은 소스 판독 + 단위 테스트에 한정했다. 즉 "실제 자식 프로세스가 `--settings` 파일을 읽어 `effortLevel` 을 적용한다"는 런타임 관측은 이 판정서에 없다 — 코드 경로와 단위 테스트까지만이 근거다.
- `internal/cli` 외 패키지는 돌리지 않았다. 전 패키지 판정은 CI 몫(CLAUDE.local.md §4).
- `-race` 는 돌리지 않았다. 이번 변경은 삭제뿐이고 동시성 코드를 건드리지 않는다.

## 7. Residual-risk

- **낮음.** 순수 삭제이며 빌드(darwin/windows)·vet·린트·생존 가드가 전부 초록이다. 되돌리기는 revert 한 번.
- 닫힌 SPEC 인수기록(`SPEC-INFINITE-GOAL-001`, `SPEC-MOAI-GATEWAY-001` 등)이 이제 존재하지 않는 심볼과 행번호를 가리킨다. 의도된 상태다.

---

## 8. [별건] 조사 중 발견한 미검증 전제 — 리드 보고용

이 카드의 범위 밖이라 **수리하지 않았다.** 카드 발행 여부는 리드 판단.

`internal/cli/launcher.go` 의 Claude/게이트웨이 분기 주석(삭제 전 파일 `:794-795`)이 이렇게 단언한다:

```
// TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode
// and TestGatewayLaunchEnvPreservesInheritedEffort pin both halves.
```

**두 테스트 모두 존재하지 않는다.**

```
$ grep -rn 'func TestGateway.*Effort' --include='*_test.go' internal/
(무출력)

$ grep -rn 'TestGatewayLaunchEnvPreservesInheritedEffort\|TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode' .
internal/cli/launcher.go:794   (주석 자신)
internal/cli/launcher.go:795   (주석 자신)

# 양성 대조 — 실재하는 이름은 잡힌다
$ grep -rln 'TestClaudeLaunchEnvPreservesInheritedEffort' .
internal/cli/launch_effort_settings_test.go
internal/cli/launch_effort_settings.go
```

무출력이 "없음"이 아니라 "내 명령이 못 봄"일 가능성을 배제하려고 양성 대조를 같은 세션에서 돌렸다.

즉 게이트웨이 경로의 effort 불변식은 **기계적 담지자 없이 주석의 주장으로만 존재한다.** Claude 경로 절반은 실재하는 seam 이 집지만, 게이트웨이 절반은 아무것도 집지 않는다. 주석을 읽은 다음 사람은 덮여 있다고 믿게 된다 — 가드 주석이 인접 층을 "이미 덮는다"고 단언하는 바로 그 형태다.

(행번호는 삭제 **전** 파일 기준. 삭제 후에는 34행 앞당겨졌다.)

🗿 MoAI

---

## 9. 리드 판정 — run 종료

- **승인**: 제거가 맞다는 판정을 리드가 승인. sync 산출물 불요(SPEC 없는 Class B, 사용자 표면 아닌 내부 죽은 코드 제거) — CHANGELOG 도 쓰지 않는다.
- **§8 별건**: 리드가 카드 `t938` 로 발행. 리드가 같은 관측을 독립 재현했다(정의 0 + 양성 대조 1적중).
  [HARD] t938 은 두 테스트가 삭제·개명·부재 중 어느 쪽인지 `git log -S` 로 먼저 가른 뒤 조치를 정하며,
  주석 수정이 답이 아닐 수 있다(불변식이 기계적 담지자를 가져야 하는지부터 판정).
- **창**: 리드 지명 대기. 지명 시 흡수 시점 이후 전진한 develop 을 **재흡수**하고 **병합 트리에서 재측정**한 뒤 병합한다.

- **재측정 범위는 흡수 후에 정한다.** 흡수 전에 정한 범위를 흡수 후에 그대로 쓰면 그것도 낡은
  근거의 재사용이다. 흡수 직후 `git diff --stat` 으로 무엇이 들어왔는지 읽고 `internal/cli` 만으로
  충분한지 다시 판정한다(직전 실측 1226.404s — 창 점유가 길어지므로 범위 확대는 근거를 갖춰서).
- **반출 상태(측정)**: primary `.moai/reports/t900/verdict.md` 와 워크트리 사본 sha256 일치
  `899c1020116611489e132eceadb5c779e84db1bc09ed13d71977f7bd43e9d3de` (11,291 B 시점). 리드가 primary 에서
  `git status --porcelain` → `?? .moai/reports/t900/`, `git ls-files --error-unmatch` 실패로 **미추적 확정**.
  반출 전 대상 자리에 §9 이전 구판이 있었고 `diff` 가 `183a184,195`(순수 append)로 분기 부재를 보여 덮었다.
  이 트리에서 primary 의 무시 규칙을 물으면 다른 질문이 되므로, 미추적 판정은 리드 측정이 근거다.

🗿 MoAI

## 반출 증거의 한계 (2026-09-19 추가, 리드 지시)

1. `logs/t900-build.log` 와 `logs/t900-vet.log` 는 0바이트다. 이는 성공 시 아무것도 출력하지 않는 명령(`go build` / `go vet`)의 정상적인 결과다.
2. 그러나 **빈 파일은 통과의 증거가 아니다** — 명령이 아예 실행되지 않았어도 결과는 같은 0바이트다. 두 파일은 그 자체로는 어느 쪽인지 구분하지 못한다.
3. 실제 관측은 같은 Bash 호출에서 echo 한 `BUILD_EXIT:0` / `VET_EXIT:0` 이었고, 그 값은 세션 전사에만 있으며 반출된 파일에는 들어 있지 않다.

따라서 이 디렉터리의 반출물 중 **자립적으로 해석되는 증거는 `logs/t900-test.log` 하나뿐**이다. 재실행으로 메우지 않는다 — 병합 이후 develop 이 이미 움직였으므로 지금의 재실행은 그때 그 트리의 측정이 아니다.
