# F01 카드 판정

## Claim

최종 plan-phase `review-N` 보고서만 감사 캐시 원천으로 사용하고 날짜별 run-gate 보고서는 이력으로만 남긴다. 동일한 SPEC 산출물 해시에서 날짜 기록이 바뀌어도 plan-auditor를 다시 호출하지 않는다.

## Evidence

- `go test ./internal/runtime -count=1`
- 관찰된 출력: `ok   github.com/modu-ai/moai-adk/internal/runtime 1.523s`
- `TestFileAuditCacheIgnoresDateHistory`에서 2099-12-31 날짜 이력 파일을 추가하고도 `CacheHit=true`, auditor 호출 수 `0`, 반환 보고서가 `SPEC-CACHE-001-review-1.md`임을 확인하였다.
- `ResolveLatestPlanAudit`가 `review-2`를 선택하고 hash·score·auditor version을 같은 파일에서 읽으며, malformed score는 오류로 거부한다.

## Baseline-attribution

위 출력은 `WT-workflow-audit-f01`에서 F01 수정 후 실행한 `internal/runtime` 패키지 기준이다.

## Gaps

실제 Claude 런타임의 plan-auditor 호출 계측과 운영 보고서 포맷 전체 변형은 이 Go 단위 fixture에서 관찰하지 않았다.

## Residual-risk

기존 `review-N` 파일에 hash metadata가 없으면 안전하게 cache miss가 되어 재감사한다. 이후 auditor 출력은 `Plan Artifact Hash`와 `Auditor Version`을 포함해야 한다.
