# t802 — GLM 정리 간극 조사·수리 판정서

- 카드: t802 (Class B · 조사 + R1·R2 수리 + injectGLMEnv 삭제)
- 리드 판정(2026-09-18): R1+R2 를 이 카드에서 수리 · R3·R4 → 새 카드 t888 · R5 → 새 카드 t889 ·
  injectGLMEnv 삭제는 R1 과 **한 커밋** · 병합은 수리까지 한 번에
- 브랜치: `WT-glm-cleanup-gap` / base `881aa4bb8`
- 측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t802`
- 측정 일자: 2026-09-18
- 재현자(이력 보존, build tag `residue_probe`):
  - `internal/cli/glm_context_residue_probe_test.go`
  - `internal/hook/glm_context_residue_probe_test.go`

---

## Claim

1. **카드의 1차 주장은 참이다.** `ensureGLMCredentials`가 `.claude/settings.local.json`에
   `CLAUDE_CODE_MAX_CONTEXT_TOKENS`를 쓰지만, settings 축의 정리 함수 3종 어느 것도 그 키를
   지우지 않는다.
2. **카드보다 강한 성질이 하나 더 있다 — 잔여는 지연이 아니라 영구다.** `ANTHROPIC_BASE_URL`이
   GLM-활성 지표인데, `moai cc`(`removeGLMEnv`)가 지표를 먼저 지우고 컨텍스트 창 키를 남기므로
   이후 `cleanupGLMSettingsLocal`은 "GLM 모드 아님"으로 판단해 조기 반환한다. 하류의 어떤 경로도
   그 잔여를 다시 지울 수 없다.
3. **settings 축에서 그 키를 지우는 코드는 단 하나이고, 그것은 프로덕션에서 호출되지 않는다.**
   `delete(env, config.EnvClaudeCodeMaxContextTokens)`는 `internal/cli/glm.go:1032` 한 곳뿐이며,
   그 함수 `injectGLMEnv`는 비테스트 참조가 0건이다(t803 (1)번 항목과 동일 개체).
4. **카드의 「4종 목록 불일치」 주장은 참이지만, 불일치의 일부는 설계상 정상이다.** 목록은 서로 다른
   두 축(settings 파일 / tmux 세션 env)을 담당하고, 두 축이 나르는 키 집합이 애초에 다르다.
   `ANTHROPIC_REASONING_EFFORT` · `CLAUDE_CONFIG_DIR` · `DISABLE_PROMPT_CACHING`은 settings 파일에
   기록되는 경로가 없으므로, settings 축 정리 함수에 없는 것은 결함이 아니다.

---

## Evidence

### E1 — 재현자 실행 (태그 적용, 실패 = 결함)

```
$ go test -tags residue_probe ./internal/cli/ ./internal/hook/ -run 'TestProbe' -count=1
--- FAIL: TestProbeRemoveGLMEnvLeavesContextWindow (0.00s)
    glm_context_residue_probe_test.go:56: removeGLMEnv left CLAUDE_CODE_MAX_CONTEXT_TOKENS="1000000" in settings.local.json
--- FAIL: TestProbeStripGLMCredsLeavesContextWindow (0.00s)
    glm_context_residue_probe_test.go:75: stripGLMCredsAndSetTeammateMode left CLAUDE_CODE_MAX_CONTEXT_TOKENS="1000000" in settings.local.json
    glm_context_residue_probe_test.go:75: stripGLMCredsAndSetTeammateMode left CLAUDE_CODE_AUTO_COMPACT_WINDOW="1000000" in settings.local.json
    glm_context_residue_probe_test.go:75: stripGLMCredsAndSetTeammateMode left ANTHROPIC_DEFAULT_FABLE_MODEL="glm-5.3" in settings.local.json
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.885s
--- FAIL: TestProbeCleanupGLMSettingsLocalLeavesContextWindow (0.00s)
    glm_context_residue_probe_test.go:79: cleanupGLMSettingsLocal left CLAUDE_CODE_MAX_CONTEXT_TOKENS="1000000"
    glm_context_residue_probe_test.go:79: cleanupGLMSettingsLocal left CLAUDE_CODE_AUTO_COMPACT_WINDOW="1000000"
    glm_context_residue_probe_test.go:79: cleanupGLMSettingsLocal left MOAI_STATUSLINE_CONTEXT_SIZE="1000000"
--- FAIL: TestProbeResidueIsPermanentOnceIndicatorIsGone (0.00s)
    glm_context_residue_probe_test.go:99: residue is permanent: CLAUDE_CODE_MAX_CONTEXT_TOKENS="1000000" survives with no indicator present
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.637s
FAIL
```

`TestProbeTmuxClearVarsIsTheCompleteList`는 **통과**했다 — 양성 대조다. tmux 축 목록
(`buildTmuxClearVars`)은 세 잔여 키를 모두 지우므로, settings 축의 결손은 "미지의 키 집합"이
아니라 **두 목록 사이의 갈라짐**이다. 이 대조가 통과하지 않으면 위 실패는 공허할 수 있었다.

### E2 — 기본 스위트에서 재현자가 제외됨 (공허 초록 방지)

```
$ go list -f '{{range .TestGoFiles}}{{.}}{{"\n"}}{{end}}' ./internal/cli/ | grep -c residue_probe
0
$ go list -f '{{range .TestGoFiles}}{{.}}{{"\n"}}{{end}}' ./internal/hook/ | grep -c residue_probe
0
$ go test ./internal/cli/ ./internal/hook/ -run 'TestProbe' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	1.465s
ok  	github.com/modu-ai/moai-adk/internal/hook	1.085s [no tests to run]
```

`internal/cli`의 `ok`에 `[no tests to run]` 표기가 없는 것은 기존 `TestProbeVersionSignal_*`가
이름 접두사만 겹쳐 실행된 결과이며, 재현자 유입이 아니다(위 `go list` 0건이 근거).

### E3 — settings 축의 유일한 삭제 지점과 그 도달 불가

```
$ grep -rn 'EnvClaudeCodeMaxContextTokens' internal/ | grep -v '_test.go'
internal/config/envkeys.go:397:	EnvClaudeCodeMaxContextTokens = "CLAUDE_CODE_MAX_CONTEXT_TOKENS"
internal/cli/glm.go:383:		_ = os.Setenv(config.EnvClaudeCodeMaxContextTokens, tokens)
internal/cli/glm.go:539:		vars[config.EnvClaudeCodeMaxContextTokens] = tokens
internal/cli/glm.go:623:		config.EnvClaudeCodeMaxContextTokens,
internal/cli/glm.go:1030:			env[config.EnvClaudeCodeMaxContextTokens] = tokens
internal/cli/glm.go:1032:			delete(env, config.EnvClaudeCodeMaxContextTokens)
internal/cli/mcp_claude.go:235:	config.EnvClaudeCodeMaxContextTokens:  {},
internal/hook/session_start.go:886,889,993,997  (ensureGLMCredentials / maybeDeclareGLMContextWindow)
```

`:623`은 tmux 축(`buildTmuxClearVars`), `:1032`는 settings 축이지만 `injectGLMEnv` 안이다.
`injectGLMEnv`의 비테스트 참조는 0건 — 정의·주석·테스트뿐:

```
$ grep -rn 'injectGLMEnv' internal/ cmd/ pkg/
internal/cli/glm.go:984   (주석)
internal/cli/glm.go:990   (정의)
internal/cli/launcher_test.go / glm_new_test.go / coverage_improvement_test.go / glm_test.go / oauth_token_preservation_test.go
```

### E4 — 축별 목록 대조표

`L` = 라이브 프로덕션 기록 경로(`ensureGLMCredentials`, SessionStart 훅)가 실제로 쓰는 키.
`D` = 도달 불가한 `injectGLMEnv`만 쓰는 키(구 바이너리가 남긴 잔재로는 실재 가능).

| 키 | 기록 | `removeGLMEnv` (settings) | `stripGLMCreds…` (settings) | `cleanupGLMSettingsLocal` (settings) | `buildTmuxClearVars` (tmux) | `glmEnvVarsToClean` (tmux) |
|---|---|---|---|---|---|---|
| `ANTHROPIC_AUTH_TOKEN` | L | 복원/삭제 | 복원/삭제 | 복원/삭제 | 의도적 제외(문서화) | 삭제 |
| `MOAI_BACKUP_AUTH_TOKEN` | L | 삭제 | 삭제 | 삭제 | — | — |
| `ANTHROPIC_BASE_URL` | L | 삭제 | 삭제 | 삭제 | 삭제 | 삭제 |
| `ANTHROPIC_DEFAULT_OPUS_MODEL` | L | 삭제 | 삭제 | 삭제 | 삭제 | 삭제 |
| `ANTHROPIC_DEFAULT_SONNET_MODEL` | L | 삭제 | 삭제 | 삭제 | 삭제 | 삭제 |
| `ANTHROPIC_DEFAULT_HAIKU_MODEL` | L | 삭제 | 삭제 | 삭제 | 삭제 | 삭제 |
| `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` | L | 삭제 | 삭제 | **없음** | 삭제 | **없음** |
| `CLAUDE_CODE_AUTO_COMPACT_WINDOW` | L | 삭제 | **없음** | **없음** | 삭제 | **없음** |
| `CLAUDE_CODE_MAX_CONTEXT_TOKENS` | L | **없음** | **없음** | **없음** | 삭제 | **없음** |
| `ANTHROPIC_DEFAULT_FABLE_MODEL` | D | 삭제 | **없음** | **없음** | 삭제 | **없음** |
| `API_TIMEOUT_MS` | D | 삭제 | 삭제 | **없음** | 삭제 | **없음** |
| `MOAI_STATUSLINE_CONTEXT_SIZE` | tmux 전용 | 삭제 | 삭제 | **없음** | 삭제 | **없음** |
| `CLAUDE_CODE_TEAMMATE_DISPLAY` | 런처 | 삭제 | 삭제 | **없음** | — | — |
| `ANTHROPIC_REASONING_EFFORT` | tmux/proc 전용 | 해당 없음 | 해당 없음 | 해당 없음 | 삭제 | — |
| `CLAUDE_CONFIG_DIR` | tmux 전용 | 해당 없음 | 해당 없음 | 해당 없음 | 삭제 | — |
| `DISABLE_PROMPT_CACHING` | 레거시 | 해당 없음 | 해당 없음 | 해당 없음 | 삭제 | — |

라이브 기록 키(`L`) 9개 기준 결손: `removeGLMEnv` 1건 · `stripGLMCredsAndSetTeammateMode` 2건 ·
`cleanupGLMSettingsLocal` 3건.

### E5 — 비범위 건전성

```
$ go vet ./internal/cli/ ./internal/hook/
(출력 없음, exit 0)

$ gofmt -l internal/cli/glm_context_residue_probe_test.go internal/hook/glm_context_residue_probe_test.go
(출력 없음)

$ golangci-lint run ./internal/cli/... ./internal/hook/...
internal/cli/gtd_answer.go:84:15: S1038 (staticcheck)
internal/cli/launcher.go:811:3: S1021 (staticcheck)
2 issues
```

린트 2건은 이 카드가 건드리지 않은 파일의 **선재 결함**이다(이 카드의 변경 표면은 신규 테스트
파일 2개뿐 — `git status --porcelain`이 `?? ` 2줄). 이 카드에서 수리하지 않는다(범위 규율).

---

## Baseline-attribution

- 트리: `WT-glm-cleanup-gap` @ `881aa4bb8` (이 실행에서 `git rev-parse --short HEAD`로 재측정)
- 이 판정서의 모든 수치·출력은 위 트리에서 이번 실행에 직접 측정한 것이다. 카드 본문의
  라인 번호는 발견 당시(미병합 `moai-proxy-unified` 워크트리) 값이며, 이 트리에서 재확인한
  좌표는 본문에 다시 적었다.
- 도구: 트리의 `go` 툴체인과 `golangci-lint`를 직접 호출했다. `moai` 바이너리로 만든 측정은
  이 판정서에 없으므로 바이너리 지연(§2.2) 좌표는 해당 없음.

---

## Gaps (명시적 미검증)

- **`internal/cli` 전량은 base 기준으로 1건 실패했다** — `TestGTDAllTodoVerbsParity`. 이 카드의
  변경이 원인이 아님은 A/B 로 확정했고, t867 이 `develop` 에서 이미 해소했다(위 § 참조).
  그러므로 이 카드가 측정한 것은 「`internal/cli` 전량 초록」이 아니라 「이 카드가 새로 깨뜨린
  것이 없음」이다. 두 주장은 다르고, 후자만 측정됐다. 전자는 **develop 흡수 후** 성립하며,
  흡수 시점에 그 테스트를 지목 재실행해 확인한다 — 흡수 전 이 판정서로 전량 초록을 주장하지
  않는다.
- **런타임 미재현.** 잔여 `CLAUDE_CODE_MAX_CONTEXT_TOKENS`가 후속 `moai cc` 세션에서 실제로
  잘못된 컨텍스트 창을 만드는지는 재현하지 않았다. 이 저장소에서 GLM 통합 테스트 실행이
  금지되어 있고(리드 가드), 실제 settings 파일을 건드리기 때문이다. 잔여의 **존재**는 기계적으로
  증명됐고, 그 **해악**은 아래 Residual-risk의 추론이다.
- **전체 스위트 미실행.** 변경이 build tag로 격리된 테스트 파일 2개뿐이므로 기본 스위트는
  구성상 영향받을 수 없고(E2가 근거), 머신 load가 19.98이어서 대형 패키지 전량 실행을
  일부러 하지 않았다. 전 패키지 판정은 CI 몫이다.
- **CG 티메이트 주입 비대칭 미판정.** `glmTmuxKeys`(주입, 9키)는 `ACW`/`MCT`/`FABLE`을 넣지
  않는데 `buildTmuxClearVars`(정리, 14키)는 지운다. 정리가 주입의 상위집합인 방향은 안전하지만,
  CG 티메이트 페인이 리드와 다른 창 선언을 갖는지는 재지 않았다. 별개 결함 계열 후보.
- **구 바이너리 잔재 실측 없음.** `D` 등급 키(`FABLE`, `API_TIMEOUT_MS`)가 실사용자 파일에
  남아 있는지는 세지 않았다.

---

## Residual-risk

- **잔여 창 선언의 방향은 양쪽 모두 해롭다(추론).** 코드베이스 자신의 근거 주석
  (`internal/config/envkeys.go:390-397`)이 적듯 Claude Code는 미인식 모델 id에 200K를 가정하고
  `CLAUDE_CODE_AUTO_COMPACT_WINDOW`를 그 가정치로 깎는다. GLM이 1M 모델이었다면 잔여
  `MCT=1000000`이 200K 슬롯 세션의 창을 5배 과대 선언하고(실제 천장 근처에서 스트림 정지),
  GLM이 200K 모델이었다면 1M 슬롯 세션의 창을 과소 선언한다(조기 압축·문맥 낭비). 어느 쪽도
  경고를 내지 않는다 — 이것이 이 결함이 조용한 이유다.
- **`injectGLMEnv`가 되살아나면 결함이 가려진다.** 그 함수만이 settings 축에서 스테일 값을
  정리하므로, 수리 없이 배선이 복구되면 증상이 간헐적으로 사라져 재현이 더 어려워진다.
  t803과 같은 개체이므로 두 카드의 처분은 함께 정해야 한다.
- **정리 목록이 다섯 벌인 구조 자체가 재발 원인이다.** `GLMEnvVarSet()`(envkeys.go:337)이
  3키에 대해서만 SSOT 역할을 하고, 나머지 키는 목록마다 손으로 열거된다. 키를 하나 더
  주입하는 다음 변경도 같은 방식으로 갈라질 수 있다.

---

## 수리 (R1 + R2 + injectGLMEnv 삭제) — 실행 완료

### 무엇을 고쳤는가

| 함수 | 파일 | 추가한 삭제 키 |
|---|---|---|
| `removeGLMEnv` | `internal/cli/launcher.go` | `CLAUDE_CODE_MAX_CONTEXT_TOKENS` |
| `stripGLMCredsAndSetTeammateMode` | `internal/cli/settings.go` | `CLAUDE_CODE_AUTO_COMPACT_WINDOW`, `CLAUDE_CODE_MAX_CONTEXT_TOKENS` |
| `cleanupGLMSettingsLocal` | `internal/hook/session_end.go` | `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`, `CLAUDE_CODE_AUTO_COMPACT_WINDOW`, `CLAUDE_CODE_MAX_CONTEXT_TOKENS` |

라이브 기록 키 9개 기준 결손이 세 함수 모두 **0** 이 됐다. `cleanupGLMSettingsLocal` 의 doc 주석이
설명하던 삭제 목록도 함께 정정했다(목록을 서술하는 문장이 목록과 갈라지면 그게 t803 계열의 결함이
된다).

`injectGLMEnv`(`internal/cli/glm.go`)와 그 테스트 8개를 같은 커밋에서 제거했다. 순서가 규율이다 —
**R1 이 settings 축으로 정리 경로를 옮긴 뒤 죽은 원본을 제거**한다. 두 커밋으로 나누면 그 사이
트리에 「그 키의 정리 경로가 0」인 지점이 실제로 생기고, 누가 그 지점을 base 로 잡으면 결함을
물려받는다. 새로 죽은 헬퍼는 없다(`getGLMAPIKey` · `glmAutoCompactWindow` · `glmMaxContextTokens` ·
`settingsEnvMap` · `mutateSettingsLocal` 전부 생존 확인).

### 상시 가드 신설 (기본 스위트에서 돈다)

`internal/cli/glm_settings_cleanup_test.go` · `internal/hook/glm_settings_cleanup_test.go` — build
tag 없이 CI 에서 돈다. **수리를 지키는 검사가 돌지 않으면 수리는 보호되지 않는다.** 네 축:

1. 라이브 9키 결손 0 (세 정리 함수 각각)
2. OAuth 백업 복원 경로 — `MOAI_BACKUP_AUTH_TOKEN` 이 있으면 `ANTHROPIC_AUTH_TOKEN` 으로 **복원**되고
   백업 키는 소비된다. 이 짝이 없으면 1번은 「사용자 토큰을 파괴하는 정리」로도 만족된다
3. tmux 축 대조 — 두 축이 같은 키 집합을 지우는지
4. **음성 대조** — 비-GLM 파일은 건드리지 않는다(`ANTHROPIC_BASE_URL` 게이트의 존재 이유이고,
   게이트 제거가 별도 설계 판단인 근거)

빈 결과집합 통과 방지로 픽스처 오염 사전 단언(`assertFixtureIsDirty`)을 넣었다 — 픽스처가 키를
심지 않게 되면 모든 단언이 공허하게 통과한다.

### RED-now (가드가 수리 없이 빨개짐을 실측)

base 파일 3개를 끼워 넣고 동일 가드를 실행했다. 틀린 이유가 아니라 **정확히 그 키들** 때문에
빨개진다:

```
$ (base 파일 교체 후) go test -timeout 10m ./internal/cli/ -run 'TestRemoveGLMEnvClearsEveryLiveKey|TestStripGLMCredsClearsEveryLiveKey|TestGLMCleanupRestoresBackedUpAuthToken' -count=1
--- FAIL: TestRemoveGLMEnvClearsEveryLiveKey
    removeGLMEnv left live key CLAUDE_CODE_MAX_CONTEXT_TOKENS="seeded"
--- FAIL: TestStripGLMCredsClearsEveryLiveKey
    stripGLMCredsAndSetTeammateMode left live key CLAUDE_CODE_AUTO_COMPACT_WINDOW="seeded"
    stripGLMCredsAndSetTeammateMode left live key CLAUDE_CODE_MAX_CONTEXT_TOKENS="seeded"
--- FAIL: TestGLMCleanupRestoresBackedUpAuthToken
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.760s

$ (base 파일 교체 후) go test -timeout 10m ./internal/hook/ -run 'TestCleanupGLMSettingsLocal...' -count=1
--- FAIL: TestCleanupGLMSettingsLocalClearsEveryLiveKey
    cleanupGLMSettingsLocal left live key CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS="seeded"
    cleanupGLMSettingsLocal left live key CLAUDE_CODE_AUTO_COMPACT_WINDOW="seeded"
    cleanupGLMSettingsLocal left live key CLAUDE_CODE_MAX_CONTEXT_TOKENS="seeded"
--- FAIL: TestCleanupGLMSettingsLocalRestoresBackedUpAuthToken
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.548s
```

음성 대조 `TestCleanupGLMSettingsLocalLeavesNonGLMFileAlone` 은 base 에서도 통과했다 — 그 검사는
수리와 독립인 성질을 재므로 이것이 맞는 거동이다. 수리본 복원은 sha256 재확인
(`session_end.go` = `64f2d21f71dbb0daf6252dc8e6a9a85789056963e6fb2615b0b7d0b9816545e5`).

### 수리 후 GREEN

```
$ go test -timeout 10m ./internal/cli/ -run 'TestRemoveGLMEnvClearsEveryLiveKey|TestStripGLMCredsClearsEveryLiveKey|TestGLMCleanupRestoresBackedUpAuthToken|TestTmuxClearVarsCoversEveryLiveKeyItOwns' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	0.604s

$ go test -timeout 10m ./internal/hook/ -run 'TestCleanupGLMSettingsLocal...' -count=1 -v
--- PASS: TestCleanupGLMSettingsLocalClearsEveryLiveKey
--- PASS: TestCleanupGLMSettingsLocalRestoresBackedUpAuthToken
--- PASS: TestCleanupGLMSettingsLocalLeavesNonGLMFileAlone
ok  	github.com/modu-ai/moai-adk/internal/hook	0.541s
```

### 남긴 미수리분 (태그 프로브가 재현자로 보존 — t888)

```
$ go test -tags residue_probe -timeout 10m ./internal/cli/ ./internal/hook/ -run 'TestProbe' -count=1
--- FAIL: TestProbeStripGLMCredsLeavesLegacyKeys
    stripGLMCredsAndSetTeammateMode left legacy key ANTHROPIC_DEFAULT_FABLE_MODEL="legacy"
--- FAIL: TestProbeSettingsAxisListsAgree
    stripGLMCredsAndSetTeammateMode keeps "ANTHROPIC_DEFAULT_FABLE_MODEL" that removeGLMEnv clears
--- FAIL: TestProbeCleanupLeavesLegacyAndOtherAxisKeys
    cleanupGLMSettingsLocal left ANTHROPIC_DEFAULT_FABLE_MODEL / MOAI_STATUSLINE_CONTEXT_SIZE / API_TIMEOUT_MS
--- FAIL: TestProbeStrandedResidueSurvivesIndicatorLoss
    stranded residue: CLAUDE_CODE_MAX_CONTEXT_TOKENS="1000000" survives with no indicator present
```

stranding 의 **라이브 진입 경로는 이 카드에서 닫혔다** — `removeGLMEnv` 가 지표와 창 키를 함께
지운다. 남은 것은 구 바이너리가 남긴 파일이며, 게이트 자체를 여는 것은 R4(t888)다.

---

## 측정 실패와 코드 판정의 구분 (리드 지시로 별도 기록)

**처음 실행한 `internal/hook` FAIL 은 코드 판정이 아니라 측정 실패였다.** 레인 세션에 주입된 환경이
원인이고, 두 계열이다:

1. `MOAI_KANBAN_BACKEND=claude` → `session_start_record_test.go:113 backend = "claude", want "glm"`
2. **더 큰 것** — `isGatewaySession()`(`internal/hook/gateway_guard.go:9-15`)은
   `MOAI_LAUNCH_PROVIDER ∈ {claude, glm}` 이면 true 이고, `cleanupGLMSettingsLocal` 의 첫 줄이 그
   게이트다. 제 세션은 `MOAI_LAUNCH_PROVIDER=claude` 라서 **정리 함수 전체가 no-op** 이었다. 그래서
   초기 가드가 「아무것도 지워지지 않음」으로 빨개졌는데, 그것은 코드가 아니라 환경이었다.

**귀속 A/B (이번 실행에서 직접 측정):** `git show 881aa4bb8:internal/hook/session_end.go` 로 base
파일을 끼워 동일 테스트를 재실행해 **동일하게 실패**함을 확인했다 → 그 실패는 이 카드의 변경과
무관하다. 그러지 않고 「제 변경이 아닐 것이다」로 넘어갔다면 그것이 미검증 주장이었다.

**세정 후 판정 (독트린의 단일 호출 형식):**

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR \
        MOAI_KANBAN_SETTINGS_INJECTED MOAI_LAUNCH_PROVIDER MOAI_FACTORY_WORKER \
        MOAI_FACTORY_WORKERS MOAI_KANBAN_BACKEND && go test -timeout 40m ./internal/hook/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/hook	196.521s
```

즉 hook 쪽에는 선재 결함도 없었다 — 전부 측정 실패였다. 가드 테스트 자신도 이 의존을 통제하도록
`scrubGatewayEnv(t)`(`t.Setenv(config.EnvMoaiLaunchProvider, "")`)를 넣었다. 그러지 않으면 그
가드는 코드가 아니라 주변 환경을 재게 되고, 그것이 없는 결함을 보고하고 있는 결함을 가린다.

**부수 관측 — 타임아웃 축:** 기본 `-test.timeout=10m` 으로는 부하 상태의 `internal/cli` 가 완주하지
못한다(리드 실측: 무부하 601s, 부하 931s). 이 판정서의 모든 실행은 `-timeout` 을 명시했다. 이번
`internal/cli` 전량은 **966.368s** 로 완주했다 — 기본 10분이면 측정 실패였을 값이다.

### `internal/cli` 전량 — 1건 실패, 선재 결함으로 귀속

```
$ unset (위 9변수) && go test -timeout 40m ./internal/cli/ -count=1
--- FAIL: TestGTDAllTodoVerbsParity (0.00s)
    gtd_compat_test.go:61: gtd verbs = [add analyze answer auto-done capture clarify done drop edit
    engage export-json history landed list move next organize pr reflect relate undone undrop unpick
    unrelate why], want [add analyze auto-done capture clarify done drop edit engage export-json
    history landed list move next organize pr reflect relate undone undrop unpick unrelate why]
FAIL	github.com/modu-ai/moai-adk/internal/cli	966.368s
```

**귀속 A/B (이번 실행에서 직접 측정):** 이 카드가 `internal/cli` 에서 고친 프로덕션 파일 3개
(`launcher.go` · `settings.go` · `glm.go`)를 모두 base(`881aa4bb8`) 판으로 되돌리고 같은 테스트를
재실행 → **바이트 동일한 실패 메시지**. 따라서 이 실패는 이 카드의 변경과 무관하다. 복원은
`git status --porcelain` 이 판정서 외 변경 0을 보여 커밋본과 바이트 동일함으로 확인했다.

**원인 커밋:** `1b644372d` — "feat(cli): gtd answer verb + factory role tokens, N removed (cards
t863+t864)". `gtd answer` 동사를 추가했으나 `todo` 쪽 짝을 맞추지 않았고, base 의 기대 목록에
`answer` 가 없다(`git show 881aa4bb8:internal/cli/gtd_compat_test.go | grep answer` → 0건).

**그리고 이 실패는 살아 있는 결함이 아니라 스테일한 측정이었다 — t867 이 이미 해소했다.**
리드 실측(2026-09-18): `develop`(`ba09526be`)의 기대 목록에 `answer` 가 들어 있다.

```
$ git -C .claude/worktrees/develop show HEAD:internal/cli/gtd_compat_test.go | grep -n answer
52:	gtdWant := append(slices.Clone(todoWant), "capture", "clarify", "organize", "reflect", "engage", "answer")
```

수리 주체는 **t867**(병합 `90b0a32bc`)이고 그 레인이 `TestGTDAllTodoVerbsParity --- PASS` 를
관측했다. 이 판정서의 실패는 base `881aa4bb8` 에서 잰 것이라 **측정 시점이 수리 이전**이었을
뿐이다. 내가 던진 「기대 목록이 문제냐 `moai todo answer` 부재가 문제냐」는 t867 이 전자로
답했다(기대 목록에 `answer` 추가) — 별도 카드는 불필요하다.

**남는 교훈은 귀속과 결론이 별개라는 것이다.** A/B 귀속("이 카드 변경 무관")은 옳았지만, 그것이
"살아 있는 결함"을 뜻하지는 않았다. base 에서 잰 실패는 base 에 대한 참이고 현행 트리에 대한 참이
아니다 — 창에서 `develop` 을 흡수한 뒤 그 테스트를 지목 재실행해 초록을 확인하고 병합한다.

---

## 수리 범위 제안 (리드 판정 완료 — 처분 기록)

| # | 범위 | 처분 | 상태 |
|---|---|---|---|
| R1 | settings 축 3종에 `CLAUDE_CODE_MAX_CONTEXT_TOKENS` 삭제 추가 | 이 카드 | **완료** |
| R2 | `stripGLMCreds…`에 `ACW`, `cleanupGLMSettingsLocal`에 `EB`+`ACW` 추가 | 이 카드 | **완료** — 라이브 9키 결손 0 |
| — | `injectGLMEnv` 삭제 (t803 (1)번과 동일 개체) | 이 카드, R1 과 **한 커밋** | **완료** |
| R3 | settings 축 정리 키를 `GLMEnvVarSet()` 식 SSOT 하나로 통합 | 새 카드 **t888** | 이관 — 태그 프로브가 재현자 |
| R4 | 지표 의존 제거 — `cleanupGLMSettingsLocal`의 `BASE_URL` 조기 반환 재설계 | 새 카드 **t888** | 이관 — 음성 대조가 게이트의 존재 이유를 기록 |
| R5 | `glmTmuxKeys` 주입 비대칭 (주입 9키 / 정리 14키) | 새 카드 **t889** | 이관 — 미재현 |

`injectGLMEnv` 삭제를 R1 과 한 커밋으로 묶은 것은 리드 판정이며 근거는 위 § 수리에 적었다 —
두 커밋 사이에 「그 키의 정리 경로가 0」인 트리 지점을 만들지 않기 위해서다.
