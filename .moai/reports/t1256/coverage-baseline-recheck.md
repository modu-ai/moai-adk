# t1256 커버리지 기준 성공 재측정

## Claim

원래 merge-base `b59a5d69c`를 `/tmp` 밖의 격리 worktree에서 다시 실행해 `internal/kanban`과 `internal/homestate`의 기준 테스트가 모두 통과함을 확인했다. `kanban` 기준 86.5%는 t1256 보완 후 86.6%보다 낮고, `homestate` 기준 69.0%는 현재 로컬 `develop`의 85.0%보다 낮다. 앞선 두 판정서의 `/tmp` 기준 테스트 실패라는 측정 간극은 이 두 패키지에 한해 해소됐다.

## Evidence

`moai worktree new t1256-baseline-probe`로 만든 별도 worktree를 고정 커밋 `b59a5d69c`에 detached checkout한 뒤 실행:

```text
$ go test -count=1 -cover ./internal/homestate
ok  github.com/modu-ai/moai-adk/internal/homestate  20.122s  coverage: 69.0% of statements
$ go test -count=1 -coverprofile=/tmp/t1256-kanban-original-base.cover ./internal/kanban
ok  github.com/modu-ai/moai-adk/internal/kanban  173.349s  coverage: 86.5% of statements
covered=3396 total=3927 percent=86.4782
```

보완 후 직접 측정한 커버리지는 `.moai/reports/t1256/coverage-gate-verdict.md`의 `internal/kanban` 86.6%(3419/3948, 테스트 통과)와 `.moai/reports/t1256/homestate-coverage-baseline.md`의 현행 `internal/homestate` 85.0%(테스트 통과)다. `git diff --name-only 543d102f5 456a65d70`은 두 측정 사이에 추가된 파일로 `homestate-coverage-baseline.md` 하나만 출력했다.

## Baseline-attribution

두 기준 명령은 이번 작업에서 같은 머신의 MoAI 생성 worktree `t1256-baseline-probe` @ `b59a5d69c`에서 직접 실행했다. 기준 검증 후 그 worktree는 원래 `WT-role-naming-baseline-probe` 브랜치로 복귀했다. 보완 후 `kanban` 수치는 `WT-role-naming-coverage-gate`의 코드 트리(로컬 병합 `543d102f5`와 동일한 tree SHA)에서, `homestate` 수치는 로컬 `develop` `543d102f5`에서 직접 측정했다.

## Gaps

- 이 재측정은 두 패키지의 커버리지 기준에 한한다. t1256의 나머지 품질 게이트를 전부 새로 실행한 것은 아니다.
- 현재 `develop`에는 t1256 이후 다른 카드의 변경도 포함된다. 현재 커버리지 증가를 t1256 단독 효과로 귀속하지 않는다.
- 원격 `develop`에는 이번 로컬 병합이 아직 반영되지 않았으므로 새 통합 트리의 CI 판정이 없다.
- 현재 goal 계약에 카드 `done` 권한이 없어 t1256은 여전히 `picked`다.

## Residual-risk

다른 카드가 로컬 `develop`을 더 변경하면 패키지 수치가 달라질 수 있다. 이후 병합은 해당 병합 트리에서 필요한 범위를 다시 확인해야 한다.
