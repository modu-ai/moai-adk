# t1256 커버리지 게이트 보완 판정

## Claim

`internal/kanban`의 옛 팩토리 역할값과 라벨 감지 경계를 검증하는 테스트를 추가했다. 같은 로컬 `develop` 기반의 변경 전후 측정에서 패키지 테스트는 통과했고 커버리지는 86.4%에서 86.6%로 올랐다. t1256의 기존 기준 보고서가 기록한 86.5%보다 높다. 이 판정은 카드 종료 또는 전체 SPEC 재감사를 뜻하지 않는다.

## Evidence

변경 전, 로컬 `develop` `4f6f46be5`에서 실행:

```text
$ go test -count=1 -coverprofile=/tmp/t1256-kanban-current.cover ./internal/kanban
ok  github.com/modu-ai/moai-adk/internal/kanban  176.720s  coverage: 86.4% of statements
covered=3411 total=3948 percent=86.3982
```

변경 후, `WT-role-naming-coverage-gate`에서 실행:

```text
$ go test -count=1 ./internal/kanban -run '^TestLegacyFactoryValuesAreDetectedWithoutBecomingLanes$' -v
--- PASS: TestLegacyFactoryValuesAreDetectedWithoutBecomingLanes (0.00s)
PASS
ok  github.com/modu-ai/moai-adk/internal/kanban  0.401s
$ go test -count=1 -coverprofile=/tmp/t1256-kanban-gate.cover ./internal/kanban
ok  github.com/modu-ai/moai-adk/internal/kanban  178.022s  coverage: 86.6% of statements
covered=3419 total=3948 percent=86.6008
```

`go tool cover -func=/tmp/t1256-kanban-gate.cover`는 `SplitFactoryLegacyAgentLabel`, `SplitFactoryLegacyLabel`, `IsLegacyFactoryRoleValue` 각각 100.0%를 출력했다. `git diff --check`는 출력 없이 exit 0이었다.

## Baseline-attribution

위 두 커버리지 프로필은 이번 작업에서 동일한 머신의 Go 도구로 각각 직접 측정했다. 변경 전 HEAD는 `4f6f46be5`이고, 변경 후는 그 커밋에서 분기한 작업 트리의 테스트 추가분이다. SPEC의 원래 merge-base `b59a5d69c` 기준 86.5%는 기존 `.moai/reports/t1256/raw/cover-base-light.txt`에 기록된 과거 측정치이며, 이번 작업에서 재측정한 값은 아니다.

## Gaps

- 원래 merge-base의 커버리지를 이번 작업에서 새로 재측정하지 않았다. 과거 기준 테스트는 `/tmp` 추출 트리에서 `TestTempOrigin_FailsOpenOnUnresolvable`이 실패했지만 커버리지 수치 86.5%는 출력했다.
- 전체 CI와 다른 패키지의 커버리지는 이 보완 작업에서 실행하지 않았다. 원격 CI는 리드가 `develop` 병합분을 일괄 push한 뒤 판정한다.
- t1256의 큐 상태는 아직 `picked`이며, 현재 goal 계약에는 `done` 권한이 없다.

## Residual-risk

다른 브랜치를 흡수하거나 `develop`에 병합하면 측정 트리가 달라질 수 있다. 병합 전후 트리의 동일성을 확인하고, 차이가 생기면 영향 범위를 다시 측정해야 한다.
