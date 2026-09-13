# t605 — develop 흡수 후 병합 트리 재측정 (통합 창, lane-5)

- 카드: t605 · ConfigChange 단발성 프로세스 비동기 결과 수리
- 흡수 대상: 로컬 `develop` = `fb8bfff95` (원격 `1d150a27d`보다 9커밋 앞선 상태, 리드 지시대로 로컬을 흡수)
- 흡수 병합: `e4e3c3005` (트리 `c892c0967364c002987cd7cb5a238dfe99a74039`)
- 흡수 결과: ort 전략, 충돌 0. 흡수분은 `internal/hook/session_end.go`(+135), `worktree_create.go`(+120), template 테스트 1본, 셸 게이트 1본 — 카드 변경 4본(`config_change.go`, `registry.go`, `defaults.go`, 신규 테스트)과 파일 교집합 없음

## Claim

develop 9커밋을 흡수한 병합 트리에서 카드의 수리가 그대로 성립한다 — 단위 군 3실행 전부 통과, 실제 CLI 프로세스 10/10에서 비동기 결과가 종료 후 감사 로그에 남는다, lint 0 issues.

## Evidence

단위 군 (모두 병합 트리 `e4e3c3005`, `unset MOAI_KANBAN* &&` 형태로 직렬 실행):

```
go test ./internal/hook/ -run 'TestRegistryShutdownJoinsConfigChangeAsyncWork|TestConfigChangeAuditRecordsBothOutcomes|TestConfigChangeJoinAsync|TestRegistryShutdownTolerantOfPlainHandlers' -count=1 -v
→ --- PASS 4식별자 전부 (하위 테스트 포함), ok ... 0.671s

go test ./internal/config/ -count=1
→ ok  github.com/modu-ai/moai-adk/internal/config  2.392s

go test ./internal/hook/ -count=1
→ ok  github.com/modu-ai/moai-adk/internal/hook  166.310s
```

CLI 프로세스 군 (바이너리 `bin/moai`, `make build` — ldflags `Commit=e4e3c3005` 로 병합 트리임을 확인. 각 실행은 별도 프로세스, 5회씩 `&&` 체인으로 exit 0 전부 확인):

```
bin/moai hook config-change < /tmp/t605-verify/payload-absorb-bad.json   ×5 → stdout {} ×5
bin/moai hook config-change < /tmp/t605-verify/payload-absorb-good.json  ×5 → stdout {} ×5

grep -c 'session=absorb-fail.*result=rejected' → 5
grep -c 'session=absorb-ctrl.*result=reloaded' → 5
(전체 10줄은 본 문서 상단 grep 출력과 동일 — session id 를 새로 내어 기존 원판 증거 10줄과 분리)
```

lint:

```
golangci-lint run ./internal/hook/... ./internal/config/...
0 issues.
```

## Baseline-attribution

모든 측정은 이 런(2026-09-12 18:16~18:21 KST), 흡수 병합 트리 `e4e3c3005`에서 직접 수행했다. 부하: 빌드 전 7.06, 패키지 전체 수트 실행 시 7.91, 전 구간 종료 후 15.52(타 레인 영향 — 측정 자체는 저부하 구간에서 끝났다). `TestSessionStart_DeferredScanDoesNotBlockReturn` 은 패키지 전체 수트에서 **초록**(부하 ~8) — 부하 민감성 가설과 일치하는 관측이며, 판별과 귀속은 신규 카드 t662 소관으로 이 카드에서 다루지 않는다.

## Gaps

- `-race` 는 원판과 같이 돌리지 않았다 — CI 판정에 맡긴다(원판 Gaps 동일).
- `detail=rt005-manager` 경로(mgr 주입 가지)는 원판과 같이 미도달 — 이 카드 범위 밖.
- develop 흡수분이 가져온 `session_end`/`worktree_create` 변경은 본 카드 재측정에서 패키지 전체 수트로 겹쳐 봤을 뿐, 그 변경 자체의 정합 판정은 각 카드(t597, t600)의 증거와 CI 소관이다.

## Residual-risk

- 병합 트리 재측정은 단일 실행이다. CLI 군 10/10 은 원판과 동일 수이지만, 비동기 조인은 타이밍 의존이라 저부하에서 우연 통과할 이론적 여지는 원판과 같이 남는다.
- lint 를 패키지 2개로 국한했다 — 흡수분이 건드린 `internal/template` 테스트 파일의 lint 는 미포함(컴파일은 `make build` 성공으로 확인).
