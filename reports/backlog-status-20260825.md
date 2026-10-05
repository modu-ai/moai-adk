# 백로그 남은 카드 현황 — 2026-08-25

modu-ai/moai-adk · 팩토리 tkalrl 리드 세션 · expert tier

## Metrics

- 남은 카드 53 (44 queued + 9 picked)
- 오늘 머지 PR 7: #1600(t184) #1613(t202) #1614(t203) #1651(t261, 오버라이드·사유 기록) #1601(t188) #1630(dependabot) #1617(t205)
- 활성 레인 4: lane-4 · lane-10 · lane-12 · lane-13
- 유일 사본 브랜치 39 → bundle 백업 완료(~/.moai/backups/moai-adk-all-local-branches-20260825.bundle, 178MB, refs 323, verify 통과)

## Picked 9

| 카드 | 주제 | 상태 |
|---|---|---|
| t272 | SkillStead 벤치마크 | PR #1653 리드 검증 통과, Race Test 재실행 중 (lane-12) |
| t273 | 워크플로 4로케일 문서화 | plan-audit 진행, 네비 변경 승인(신규 페이지 2종) (lane-4) |
| t250 | Graft 도입 | #1648 CI 3차 초록(26 pass), CR 쿼터 대기 (lane-10) |
| t259 | LSEL 드레인 3주 정지 | 착수 (lane-13) |
| t215 / t228 / t229 / t235 / t251 | statusline·astgrep·audit_multi·gate·보안게이트 | 이전 레인 진행 |

## Queued 44 — 그룹 분포 (검산 44 ✓)

| 그룹 | 수 | 카드 |
|---|---|---|
| 프로세스·도구 결함 | 11 | t224 t241 t242 t243 t244 t252 t253 t254 t255 t258 t260 |
| 릴리즈·버전 | 6 | t204(배포게이트·운영자 선행) t274 t90 t196 t197 t191 |
| 감사(audit_multi/codex) | 6 | t225 t234 t240 t246 t248 t249 |
| GitHub 이슈 연동 | 5 | t200(#1593/94) t201(#1595) t233(#1631) t236(#1640) t237(#1641) |
| 워크트리·세션 | 5 | t223 t231 t247 t264 t267 |
| CI·테스트 품질 | 4 | t256 t270 t271 t278 (flake 3연속 계열 — 상위 조사 t278에 포함) |
| 문서·기타 | 4 | t154(보류) t125 t239 t269 |
| 훅·성능 | 2 | t216 t263 |
| 보안 | 1 | t262 |

## 기타 오늘 확정 사항

- flake 계열: t270(sessionmsg poll)·t271(timing paired)·t278(hook RT005) — 공통 양상 "로컬 darwin 초록 + CI ubuntu 1회 붉음 + 수 초 내 실패"
- CC 업스트림 2.1.239→2.1.241 갱신(null delta, 242/243 창구 개방)
- 카드 done 4장: t202 t203 t205 t261 · 발행 6장: t270–t274, t278

Sources: .moai/state/kanban/backlog.json (44+9+10 실측) · gh pr view/checks · 리드 독립 재검증
