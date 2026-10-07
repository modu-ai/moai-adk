# 백로그 남은 카드 현황 — 2026-08-26

modu-ai/moai-adk · 팩토리 tkalrl 리드 세션 · expert tier

## Metrics

- 남은 카드 50 (42 queued + 8 picked)
- 누적 머지 PR 11 · 열린 PR 0
- 레인: lane-4·lane-10 활성 / lane-12·lane-13 가용
- moai codex 런처: 미구현 — 실행 실측 `Unknown command "codex"`. 구현 카드 t197 [v3.2] 후순위. 배포된 것은 Codex 배선(init --agent codex·hook --harness codex·doctor)이며 런처와 별개

## Picked 8

| 카드 | 상태 |
|---|---|
| t274 | v3.1.3 문서 반영 + version-sync 선결함 (lane-4 배차) |
| t279 | t250 후속 — 연기 31·SPEC 재위임·Minor 2·backfill (lane-10 배차) |
| t215·t228·t229·t235·t251·t267 | 이전 세션/레인 진행 |

## Queued 42 — 그룹 분포 (검산 42 ✓)

| 그룹 | 수 | 카드 |
|---|---|---|
| 프로세스·도구 결함 | 12 | t224 t241 t242 t243 t244 t252 t253 t254 t255 t258 t260 t280 |
| 감사(audit_multi/codex) | 6 | t225 t234 t240 t246 t248 t249 |
| 릴리즈·버전 | 5 | t204(배포게이트·운영자 선행) t90 t196 t197(codex 런처) t191 |
| GitHub 이슈 연동 | 5 | t200 t201 t233 t236 t237 |
| CI·테스트 품질 | 4 | t256 t270 t271 t278 (flake 3연속 계열) |
| 워크트리·세션 | 4 | t223 t231 t247 t264 |
| 문서·기타 | 3 | t154(보류) t125 t239 |
| 훅·성능 | 2 | t216 t263 |
| 보안 | 1 | t262 |

## Shipped 11 (세션 누적)

#1600(t184) · #1613(t202) · #1614(t203) · #1651(t261, 오버라이드) · #1601(t188) · #1630(dependabot) · #1617(t205) · #1653(t272) · #1655(t259) · #1656(t273) · #1648(t250, 3라운드 수렴)

## 기타

- CI flake 3연속: t270·t271·t278 — 공통 양상 "로컬 darwin 초록 + CI ubuntu 1회 붉음"
- 유일 사본 39개 전역 bundle 백업 유지 — 커밋 유실 경로 없음
- t269는 큐에서 소멸(다른 세션 처리 추정)

Sources: .moai/state/kanban/backlog.json (42+8+10 실측 2026-08-26) · ~/go/bin/moai codex 실행 실측 · gh pr view/checks
