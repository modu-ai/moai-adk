# CG 마이그레이션·초기 차단기 로컬 구현 검증

## Claim

SPEC-MOAI-CG-RETIRE-001 Step 2의 로컬 구현을 작성했다. `moai migrate cg`는 기본 preview이며, `claude-only` 적용에는 `--apply --accept-role-change`가 모두 필요하다. 실제 TEAMMATE capability가 없는 `claude-glm` 적용 및 저장된 hybrid 정책의 실행은 닫혀 있다. 사용자 YAML의 `verified: true`는 권한을 부여하지 않는다.

`internal/config/cg_migration.go`는 YAML 노드 판독·변경으로 주석, 미지 필드, GLM 모델, credential 참조를 보존한다. 중복 키, anchor/alias, 다중 문서, 비정상 자료형, 충돌 정책, 비어 있지 않은 독립 `llm.mode`는 해당 판정에서 거부한다. 빈 원문과 일반 구성은 launcher guard를 통과한다. 원문 읽기는 1 MiB로 제한하고 UTF-8을 검사한다.

`internal/cli/migrate_cg.go`는 원문 SHA-256 → O_EXCL 잠금 → 원문 재확인 → 정확한 0600 backup → 0600 임시 파일 write/Sync/검증 → 원문 재확인 → atomic Replace → 재읽기 순서를 따른다. backup 위치는 `.moai/backups/cg-migration/<원문 sha256>.yaml`이다. 같은 backup은 private regular file이고 정확히 같은 bytes일 때만 재사용한다. 기존 lock을 자동 회수하지 않으며, 종료 시 자신이 연 inode와 같은 lock만 제거한다. 실패 후 backup은 보존한다. 같은 target은 unchanged, 반대 target은 오류다.

cc/glm/gpt 초기 진입은 profile·spawn·worktree·factory 처리보다 먼저 raw guard 및 saved teammate policy를 읽는다. unified launcher에도 준비 경계 차단기를 둔다. `root.go`는 정확한 `migrate cg` 경로만 typed dependency 초기화를 건너뛴다. GLM setup/status/tools와 GPT login/logout/status는 launch가 아니므로 기존 닫힌 명령 처리를 유지한다.

## Evidence

아래 명령은 모두 이 worktree에서 실행했다. `<SCRUB>`는 별도 호출이 아니라 각 명령에 붙인 `unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY &&`이다. `<ENV>`는 `MOAI_HOME=/tmp/gateway-cli-verification-home GOCACHE=/tmp/gateway-foundation-cache`이다.

### RED

`<SCRUB> GOCACHE=/tmp/gateway-foundation-cache go test ./internal/config -run 'TestCGMigration|TestCGRaw' -timeout 30s`

```text
# github.com/modu-ai/moai-adk/internal/config [github.com/modu-ai/moai-adk/internal/config.test]
internal/config/cg_migration_test.go:18:10: undefined: GuardLegacyCG
internal/config/cg_migration_test.go:18:58: undefined: ErrLegacyCG
internal/config/cg_migration_test.go:19:12: undefined: PlanCGMigration
internal/config/cg_migration_test.go:21:10: undefined: GuardLegacyCG
internal/config/cg_migration_test.go:22:14: undefined: ReadGatewayTeammatePolicy
internal/config/cg_migration_test.go:23:14: undefined: PlanCGMigration
internal/config/cg_migration_test.go:24:12: undefined: PlanCGMigration
internal/config/cg_migration_test.go:25:14: undefined: PlanCGMigration
internal/config/cg_migration_test.go:25:99: undefined: ReadGatewayTeammatePolicy
internal/config/cg_migration_test.go:36:35: undefined: PlanCGMigration
internal/config/cg_migration_test.go:36:35: too many errors
FAIL github.com/modu-ai/moai-adk/internal/config [build failed]
FAIL
```

CLI 구현 전 동일 범위 명령은 `newMigrateCGCommand`, `applyCGMigration`, `cgMigrationIO`, `guardCGLaunchAt` undefined로 build 실패했다. 실제 entry wiring 전의 동작 RED:

`<SCRUB> <ENV> go test ./internal/cli -run TestCGEntryGuardRunsBeforeLaunchAndSpawn -timeout 30s`

```text
--- FAIL: TestCGEntryGuardRunsBeforeLaunchAndSpawn (0.00s)
    migrate_cg_test.go:171: cc [--model opus]: expected legacy guard, got <nil> (launch=1 spawn=0)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.000s
FAIL
```

root dependency 예외 구현 전:

```text
--- FAIL: TestCGMigrationAvoidsTypedDependencyDecode (0.00s)
    migrate_cg_test.go:195: cg migration initializes typed dependency graph
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.935s
FAIL
```

### GREEN 및 변경 범위 회귀

`<SCRUB> <ENV> go test ./internal/cli ./internal/config -run 'TestCG' -timeout 30s -coverprofile=/tmp/cg-focused.cover`

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	1.218s	coverage: 7.9% of statements
ok  	github.com/modu-ai/moai-adk/internal/config	1.308s	coverage: 6.8% of statements
```

coverage profile의 해당 파일 statement 합산 결과:

```text
github.com/modu-ai/moai-adk/internal/cli/migrate_cg.go 190 211 90.0
github.com/modu-ai/moai-adk/internal/config/cg_migration.go 137 140 97.9
```

`<SCRUB> <ENV> go test ./internal/cli -run 'TestCG|TestGPT|TestGateway|TestUnifiedGateway|TestCCCmd|TestGLMCmd' -timeout 90s -coverprofile=/tmp/cg-cli.cover`

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	1.433s	coverage: 11.2% of statements
```

이 명령의 첫 sandbox 실행은 기존 `TestGatewaySessionRealSupervisorHandoffAndCleanup`의 `gateway child configuration invalid`에서 실패했다. 기존 프로세스 fingerprint 판독 제한 때문에 escalation 승인 후 동일 명령을 재실행한 위 결과를 회귀 증거로 사용한다. helper는 test binary이며 Claude나 provider를 실행하지 않았다.

`<SCRUB> <ENV> go test -race ./internal/cli ./internal/config -run TestCG -timeout 60s`

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	2.753s
ok  	github.com/modu-ai/moai-adk/internal/config	1.516s
```

`<SCRUB> GOCACHE=/tmp/gateway-foundation-cache go vet ./internal/cli ./internal/config`: exit 0, stdout/stderr 없음.

`<SCRUB> GOOS=windows GOARCH=amd64 GOCACHE=/tmp/gateway-foundation-cache go test -c ./internal/cli -o /tmp/cg-cli-windows.test.exe`: exit 0, stdout/stderr 없음.

`GOCACHE=/tmp/gateway-foundation-cache gopls check internal/config/cg_migration.go internal/cli/migrate_cg.go internal/cli/cc.go internal/cli/glm.go internal/cli/gpt.go internal/cli/launcher.go internal/cli/root.go`: exit 0, diagnostics 없음.

### 격리 mutant

실제 공유 파일을 고치지 않고 `/tmp/cg-migration-mutants/{entry,role,readback,policy}.json` Go overlay로 네 변형을 실행했다. 각 `go test -overlay=...`는 exit 1이었다.

- entry guard 제거 → `TestCGEntryGuardRunsBeforeLaunchAndSpawn`:

```text
--- FAIL: TestCGEntryGuardRunsBeforeLaunchAndSpawn (0.00s)
    migrate_cg_test.go:171: cc [--model opus]: expected legacy guard, got <nil> (launch=1 spawn=0)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.939s
FAIL
```

- role-change 동의 검사 제거 → `TestCGMigrationPreviewApplyAndGuards`:

```text
--- FAIL: TestCGMigrationPreviewApplyAndGuards (0.02s)
    migrate_cg_test.go:47: accepted [--apply --target claude-only]
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.042s
FAIL
```

- readback 실패를 성공으로 변경 → `TestCGMigrationTransactionFailures`:

```text
--- FAIL: TestCGMigrationTransactionFailures (0.05s)
    --- FAIL: TestCGMigrationTransactionFailures/readback (0.01s)
        migrate_cg_test.go:104: failure claimed success
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.995s
FAIL
```

- team_mode만 바꾸고 saved policy 두 키 누락 → `TestCGMigrationPlanPreservesNodesAndBecomesIdempotent`:

```text
--- FAIL: TestCGMigrationPlanPreservesNodesAndBecomesIdempotent (0.00s)
    cg_migration_test.go:27: lost "teammate_mode: in-process": llm:
          team_mode: claude # original mixed roles
          glm_env_var: MY_GLM_KEY
          glm:
            models: {high: glm-5.2, medium: glm-4.7, low: glm-4.5-air, fable: glm-4.7}
          unknown_keep: yes # user comment
          gateway: {}
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/config	0.384s
FAIL
```

## Baseline-attribution

2026-09-11, `.claude/worktrees/moai-proxy-unified`, `WT-unified-gateway`, `81c1d58f9`. 이 작업 시작 전 fetch 결과 `From https://github.com/modu-ai/moai-adk` / `* branch main -> FETCH_HEAD`, 비교 결과 `0 2879`였다. 마지막 HEAD·branch 직접 재확인:

```text
81c1d58f9
WT-unified-gateway
```

기준 문서는 CG-RETIRE 0.1.0의 여섯 문서와 `sibling-plan-audit-iter2.md` PASS 1.00이다. 기존 dirty 구현은 별도 선행 milestone이며 이번 신규 구현과 합산해 완료라고 주장하지 않는다. 위 coverage는 이 run에서 생성한 profile이며 repository 전체 coverage가 아니다.

## Gaps

- CG 명령의 wholesale 제거, 역사 backend 판독 호환, template/4locale 문서 및 문서 빌드는 이번 Step 2 범위가 아니다. AC-CR-001/004/006/008 및 전체 CG 폐기는 미완료다.
- 실제 TEAMMATE pane의 모델·provider·인증·lifetime 검증, 실제 Claude/GPT/GLM 대화는 미실행이다. hybrid 적용·launch는 계속 불가하다. production gateway 활성화도 하지 않았다.
- AC-CR-002/005의 로컬 entry 시험은 cc/glm/gpt 각각 model/continue/resume/spawn/작업트리·kanban·factory 입력에서 정확한 legacy 오류와 launch/spawn 계수 0을 검사한다. worktree는 값 없는 `-w`, kanban/factory는 후속 잘못된 profile flag를 포함하여 guard보다 뒤의 실제 프로세스 생성을 막았다. typed backend·credential·worktree 개별 내부 호출 계수와 완전한 정상 입력 행렬을 모두 계측한 것은 아니다. 그 전체 AC를 PASS로 표시하지 않는다.
- 저장된 hybrid 정책은 실제 capability가 없으므로 정상 hybrid positive control은 없다. claude-only 및 일반 명시 구성의 준비 seam 대조군만 통과했다.
- Windows compile만 관측했으며 Windows 파일 잠금/교체 런타임을 실행하지 않았다. 기존 AUTH Windows 미지원 여부를 이 결과로 덮지 않는다.
- 첫 구현 전 LSP baseline은 별도로 저장하지 못했다. 최종 diagnostics 0만 관측했다.
- 실제 전원 손실·디스크 full·비협조적 외부 writer와 OS 수준 race를 포괄하지 않는다. 실패 주입은 동시 source 변경, 기존 lock, backup 변조/권한/symlink, write/replace/readback 경계이다.
- 저장소 전체 test verdict는 통합 브랜치 CI가 소유하며 현재 PENDING이다. commit/push/PR/merge는 수행하지 않았다.

## Residual-risk

YAML 노드 재직렬화는 의미·주석을 보존하지만 원본 formatting의 byte 동일성을 보장하지 않는다. 정확한 원본은 backup에 있다. 마이그레이션 lock은 협조하는 writer를 직렬화하고 전후 hash로 비협조적 변경을 탐지하지만, 마지막 hash 검사와 rename 사이의 비협조적 수정까지 OS compare-and-swap으로 잠그지는 못한다. 기존 lock은 자동 탈취하지 않으므로 비정상 종료 후 운영자가 소유 프로세스와 상태를 확인해야 한다. 상위 경로를 악의적으로 동시 교체하는 공격 및 모든 플랫폼의 crash durability는 독립 감사·추가 runtime 검증 대상으로 남긴다. 현재 source도 private0600으로 교체되며 원래보다 권한을 넓히지 않는다.
