# t802 — GLM 정리 간극 조사 판정서

- 카드: t802 (Class B · 조사 · 수리 미실행)
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

## 수리 범위 제안 (리드 판정 대기 — 이 카드에서 실행하지 않음)

| # | 범위 | 성격 | 비고 |
|---|---|---|---|
| R1 | settings 축 3종에 `CLAUDE_CODE_MAX_CONTEXT_TOKENS` 삭제 추가 | 최소 수리 | 카드 1차 주장 직결. 재현자 4개 중 3개가 초록으로 바뀜 |
| R2 | `stripGLMCreds…`에 `ACW`, `cleanupGLMSettingsLocal`에 `EB`+`ACW` 추가 | 라이브 결손 완결 | `L` 9키 기준 결손 0 |
| R3 | settings 축 정리 키를 `GLMEnvVarSet()` 식 SSOT 하나로 통합 | 재발 방지(구조) | Residual-risk 3항. 범위가 가장 큼 |
| R4 | 지표 의존 제거 — `cleanupGLMSettingsLocal`의 `BASE_URL` 조기 반환 재설계 | 영구화 성질 제거 | Claim 2 직결. 조기 반환은 의도된 설계라 별도 판단 필요 |
| R5 | `glmTmuxKeys` 주입 비대칭 | 별개 계열 | 먼저 재현 필요. 별도 카드 권고 |

R1·R2는 재현자가 이미 있어 RED→GREEN이 바로 성립한다. R3·R4는 설계 변경이라 Class C
(plan 경유)로 보는 것이 타당하다. R5는 이 카드 범위 밖.
