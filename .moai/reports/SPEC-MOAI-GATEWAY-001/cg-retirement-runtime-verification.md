# CG runtime 철거 로컬 검증

## Claim

SPEC-MOAI-CG-RETIRE-001의 runtime 철거를 로컬 seam으로 구현했다. `cg` root 등록을 제거했다. 직접 `moai cg` 토큰은 migration preview 안내 오류로 종료하고, unregistered `runCG`·`applyCGMode` 호환 경계와 `cg`/`claude_glm` unified launch도 동일하게 거부한다. `spawnLaunch(...,"cg",...)`도 tmux 사전조건·spawn 전에 거부한다. 지원 launcher로 자동 별칭 처리하지 않는다.

`TeamModeCG` runtime 상수를 `LegacyTeamModeCG` 데이터 식별자로 바꾸었다. `template.IsGLMBackend`는 legacy cg를 GLM으로 인식하지 않으며 독립 mode=glm과 cg가 함께 있어도 false다. 이 경우 launcher의 raw guard가 명시 이전을 요구한다. `kanban.Read`로 과거 backend=cg JSON을 읽는 fixture는 문자열과 파일 원문을 보존하며 읽은 값을 GLM backend로 활성화하지 않는다.

SessionStart의 예전 문자열 검색 `isCGMode`는 `hasLegacyCGConfiguration` raw guard로 교체했다. legacy cg나 모호한 YAML이 자동 credential 주입을 일으키지 않게 한다. 유효한 Claude 구성의 주석에 `team_mode: cg`가 적혀 있다는 이유만으로 CG 모드라고 판단하지 않는다. 일반 tmux helper와 GLM cleanup 기능은 삭제하지 않았다.

CLI root/help·GLM help/runtime 안내·작업트리 안내에서 CG 실행 권고를 제거했다. README/template 문서는 이번 범위에서 수정하지 않았다. 예전 CG 성공·쓰기 시험은 현재 SPEC이 요구하는 retired 오류·기존 settings 원문 보존 시험으로 대체했다. 일반 atomic settings·GLM/tmux credential transport 시험은 보존했다.

## Evidence

명령의 `<S>`는 같은 invocation 앞의 `unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY &&`이다. `<E>`는 `MOAI_HOME=/tmp/gateway-cli-verification-home GOCACHE=/tmp/gateway-foundation-cache`이다. hook 시험만 `MOAI_HOME`도 unset하여 기존 격리 fixture를 사용했다. 실제 Claude·Codex·provider는 호출하지 않았다.

### RED

`<S> GOCACHE=/tmp/gateway-foundation-cache go test ./internal/cli -run TestCGRetired -timeout 30s`

```text
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/cg_retirement_test.go:12:139: undefined: errCGRetired
internal/cli/cg_retirement_test.go:13:113: undefined: errCGRetired
internal/cli/cg_retirement_test.go:14:46: undefined: errCGRetired
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```

`<S> GOCACHE=/tmp/gateway-foundation-cache go test ./internal/template -run '^TestIsGLMBackend$' -timeout 30s`

```text
--- FAIL: TestIsGLMBackend (0.00s)
    --- FAIL: TestIsGLMBackend/team_mode=cg_is_retired (0.00s)
        glm_effort_overlay_test.go:35: IsGLMBackend(mode="", team_mode="cg") = true, want false
    --- FAIL: TestIsGLMBackend/retired_cg_conflicts_with_mode=glm (0.00s)
        glm_effort_overlay_test.go:35: IsGLMBackend(mode="glm", team_mode="cg") = true, want false
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template	0.420s
FAIL
```

hook raw guard 구현 전 `TestLegacyCGGuardIsRawDataNotLiveMode`는 `undefined: hasLegacyCGConfiguration`으로 build 실패했다. live help 수정 전:

```text
--- FAIL: TestCGRetirementLiveHelpHasNoLaunchRecommendation (0.00s)
    cg_retirement_test.go:165: root help recommends retired launcher
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.999s
FAIL
```

spawn 내부 경계 수정 전:

```text
--- FAIL: TestCGRetiredSpawnBoundaryDoesNotOpenWindow (0.00s)
    cg_retirement_test.go:173: retired spawn: <nil>
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.940s
FAIL
```

### GREEN

`<S> <E> go test ./internal/cli -run 'CG|Cg|SpawnLaunch|HelpGroup|Fang|FactoryEntry|FactoryWorkerEntry|ExistingBranch' -timeout 60s`

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	5.966s
```

이 범위에는 과거 record Read fixture, retired root registration/내부 entry, 기존 CG settings 원문 보존, 완전한 정상/legacy entry 대조군이 포함된다. 완전한 입력은 각 cc/glm/gpt에 대해 기본, `--model opus`, `--continue`, `--resume prior-session`, `--spawn`, `-w owned-feature --branch existing`, `-p profile`, `-k 2`, `-f 2`의 9개다. legacy 27개에서 launch·spawn·worktree materialize·credential-home 계수는 모두 0이다. 명시 claude-only migration 이후 같은 27개는 준비 launch 또는 mock spawn에 정확히 한 번씩 도달하며 worktree materialize 계수는 3이다. factory/kanban 메타데이터는 테스트 임시 프로젝트와 격리 MOAI_HOME에만 기록했다.

`<S> <E> go test -race ./internal/cli -run 'CG|Cg|HelpGroup|Fang|FactoryEntry|FactoryWorkerEntry|ExistingBranch' -timeout 90s`

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	8.161s
```

`<S> GOCACHE=/tmp/gateway-foundation-cache go test ./internal/template ./internal/config -run 'GLM|Effort|CG' -timeout 60s`

```text
ok  	github.com/modu-ai/moai-adk/internal/template	0.539s
ok  	github.com/modu-ai/moai-adk/internal/config	0.420s
```

`unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_HOME && GOCACHE=/tmp/gateway-foundation-cache go test -race ./internal/hook -run 'TestLegacyCG|TestEnsureGLMCredentials|TestGateway' -timeout 60s`

```text
ok  	github.com/modu-ai/moai-adk/internal/hook	1.805s
```

`<S> GOCACHE=/tmp/gateway-foundation-cache go vet ./internal/cli ./internal/config ./internal/hook ./internal/template`: exit 0, stdout/stderr 없음.

`<S> GOOS=windows GOARCH=amd64 GOCACHE=/tmp/gateway-foundation-cache go test -c ./internal/cli -o /tmp/cg-retirement-cli-windows.test.exe`: exit 0, stdout/stderr 없음. 마지막 spawn 경계 보강 전 컴파일이며 Windows 실행 증거가 아니다.

`GOCACHE=/tmp/gateway-foundation-cache gopls check internal/cli/cg.go internal/cli/launcher.go internal/cli/root.go internal/cli/glm.go internal/cli/help.go internal/cli/help_order.go internal/cli/worktree_advisory.go internal/config/team_mode.go internal/config/cg_migration.go internal/template/glm_effort_overlay.go internal/hook/session_start.go`: exit 0, 아래 경고 1개.

```text
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/internal/hook/session_start.go:1361:13-35: bufio.Scanner "scanner" is used in Scan loop at line 1362 without final check of scanner.Err()
```

`git show HEAD:internal/hook/session_start.go`와 현재 파일에서 `func loadGLMKeyFromEnvFile` 전체를 추출한 문자열 비교 결과:

```text
loadGLMKeyFromEnvFile identical to HEAD: True
```

해당 함수는 이번에 수정하지 않았다. 경고를 숨기거나 범위 밖 수정을 하지 않았다.

마지막 spawn 경계 변경 후 `<S> <E> go test -race ./internal/cli -run 'TestCGRetiredSpawn|TestApplyCGModeRetired' -timeout 30s`:

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	2.416s
```

실제 top-level `Execute()`에 cg 기본/help/spawn+factory 인수를 넣은 `<S> GOCACHE=/tmp/gateway-foundation-cache go test ./internal/cli -run TestCGRetiredTopLevelDiagnostic -timeout 30s`:

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	1.091s
```

### 격리 mutant

공유 소스는 유지하고 `/tmp/cg-retirement-mutants/*.json` Go overlay로 실제 시험했다.

`<S> GOCACHE=/tmp/gateway-foundation-cache go test -overlay=/tmp/cg-retirement-mutants/alias.json ./internal/cli -run TestCGRetiredEntryAndModeHaveZeroEffects -timeout 30s`: CG 오류를 Claude launch 별칭으로 바꾼 변형, exit 1.

```text
--- FAIL: TestCGRetiredEntryAndModeHaveZeroEffects (0.00s)
    cg_retirement_test.go:33: retired []: <nil>
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.954s
FAIL
```

`<S> GOCACHE=/tmp/gateway-foundation-cache go test -overlay=/tmp/cg-retirement-mutants/backend.json ./internal/template -run '^TestIsGLMBackend$' -timeout 30s`: legacy CG를 GLM backend로 되살린 변형, exit 1.

```text
--- FAIL: TestIsGLMBackend (0.00s)
    --- FAIL: TestIsGLMBackend/team_mode=cg_is_retired (0.00s)
        glm_effort_overlay_test.go:35: IsGLMBackend(mode="", team_mode="cg") = true, want false
    --- FAIL: TestIsGLMBackend/retired_cg_conflicts_with_mode=glm (0.00s)
        glm_effort_overlay_test.go:35: IsGLMBackend(mode="glm", team_mode="cg") = true, want false
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template	0.418s
FAIL
```

## Baseline-attribution

2026-09-11, WT `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, branch `WT-unified-gateway`, HEAD `81c1d58f9`를 마지막에 다시 읽었다. 같은 WT의 앞선 CG 마이그레이션 Step 2 위에 이번 변경을 했다. 최신 CG-RETIRE 0.1.0 설계·AC와 sibling-plan-audit-iter2 PASS1.00이 입력이다. 다른 writer의 auth/supervisor/translate 파일은 수정하지 않았다. hook/template 공유 경계 수정은 부모에게 사전 통지했다. commit/push/PR/merge 없음.

## Gaps 및 정확한 AC 상태

| AC | 이번 관측 | 전체 판정 |
|---|---|---|
| CR-001 | root 등록 제거, retired 내부 mode/spawn 거절, mock 부작용 계수 0, help 제거 | 로컬 구현 통과; 설치 바이너리 전체 invocation 미실행 |
| CR-002 | 27 legacy 입력 guard와 27 migrated 정상 대조군, source 불변 | 로컬 부분 통과; typed backend의 모든 내부 함수별 계수·실제 egress 계측은 아님 |
| CR-003 | 앞선 cg-migration-verification.md의 저장·잠금·실패·재실행 증거 유지 | 로컬 부분 통과; OS crash/race 포괄 아님 |
| CR-004 | CG execution 연결 제거, 실제 kanban.Read 역사 문자열/원문 보존 | 로컬 부분 통과; 모든 역사 event 종류·UI display 포괄 아님 |
| CR-005 | 완전한 profile/worktree/factory/kanban/spawn 모의 정상 행렬 | 로컬 부분 통과; 실제 tmux/Claude argv/permission 런타임 미실행 |
| CR-006 | CLI help만 정리 | PENDING — README/template/4locale 문서·렌더·빌드·역사 해시 전체 미실행 |
| CR-007 | 관련 기존 시험과 두 runtime mutant 통과/실패 대조 | 로컬 부분 통과; 독립 전체 감사·CI PENDING |
| CR-008 | hybrid capability 미검증으로 apply/launch 닫힘 | PENDING — 실제 TEAMMATE 통합 필수 |

실제 Claude/GPT/GLM 요청·대화·도구·stream 시험과 production gateway 활성화는 하지 않았다. Windows는 앞서 적은 compile 범위만 관측했다. 변경 전 LSP 전체 baseline이 없으며 최종 경고 하나는 HEAD 동일 함수에 귀속했다. repository-wide test verdict는 통합 브랜치 CI가 소유하며 PENDING이다.

## Residual-risk 및 후속 참조 목록

`internal/tmux/cg_detect.go:IsCGMode`와 `sessionEnvHasGLM`/`hasGLMEnv`는 이번에 지우지 않았다. `rg`로 production Go 호출 위치를 찾았을 때 IsCGMode 정의만 반환되었지만, 이것만으로 모든 외부/반사 호출의 부재나 삭제 안전성을 주장하지 않는다. `internal/cli/factory.go:rejectFactoryOnCG`, `kanban.go:rejectKanbanOnCG`는 옛 입력 파서 오류 경계로 남고 root CG에서 실행되지 않는다. 일반 tmux·GLM credential transport를 단순 검색만으로 철거하지 않았다.

문서 담당 후속 후보를 실제 읽기/검색으로 확인했다. 이 목록은 렌더 도달성 검사나 4locale 빌드 결과가 아니다.

- README.md / README.ko.md / README.ja.md / README.zh.md: 571의 CG 실행 예시 및 685 부근 모드 표 등.
- internal/template/templates/CLAUDE.md, AGENTS.md: CG 실행/역할 안내.
- internal/template/templates/.claude/rules/moai/core/glm-web-tooling.md, settings-management.md, agent-common-protocol.md, moai-constitution.md.
- internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md, cross-session-messaging.md, cross-session-messaging-detail.md, session-handoff-examples.md.
- internal/template/templates/.claude/skills/moai/SKILL.md 및 workflows/{goal,factory,moai}.md, workflows/plan/spec-assembly.md, workflows/run/mode-orchestration.md.
- internal/template/templates/.claude/skills/moai-workflow-worktree/SKILL.md 및 modules/moai-adk-integration.md.
- internal/template/templates/.claude/agents/moai/{manager-lead,super-advisor}.md, .codex/agents/moai/{manager-lead,super-advisor}.toml.
- internal/template/templates/.claude/output-styles/moai/moai-learn.md, .moai/docs/generic-patterns-guide.md, .claude/skills/moai-foundation-quality/modules/integration-patterns.md.

역사 SPEC/report/release note와 사용자 설정은 문서 후속 시 보존 목록으로 구분해야 한다. 현재 출력용 문서 후보와 역사 파일을 같은 문자열 치환으로 처리하지 않았다.
