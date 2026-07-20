# moai-adk 배포 형식 비교 — 바이너리 vs 플러그인 vs 글로벌 심링크 하이브리드

Date: 2026-07-15 · Mode: explainer · Sources: code.claude.com/docs/en/{skills,sub-agents,memory}

## 결론

- **채택 권고**: `~/.moai/assets/<version>` 글로벌 저장소 + 프로젝트별 심링크 옵트인 (D안)
- **기각**: 전면 `~/.claude` 글로벌화 (C) — skills 우선순위 역전(personal > project) + 전 프로젝트 스코프 오염 + 팀 공유 소멸
- **기각**: 전면 플러그인화 (B) — 배포 채널 2개 병행 + 버전 스큐 신규 결함 클래스 + read-only로 user-owned 커스터마이징 충돌
- **보완**: 발견성용 게이트웨이 플러그인(skills 소수 + 설치 안내)은 별도 등록 가치 있음

## 전제: 기능 층 분해

- 플러그인/글로벌 모두 **바이너리 층 대체 불가**: `moai` CLI(session/handoff/statusline/worktree/glm), settings.json 런타임 관리, `.moai/` 스캐폴딩
- 이전 가능한 것은 자산 층(skills/agents/rules/commands/output-style)뿐

## 공식 우선순위 (검증 완료)

| 자산 | 글로벌 위치 | 이름 충돌 우선순위 | 심링크 |
|---|---|---|---|
| Skills | ~/.claude/skills/ | **enterprise > personal > project** (글로벌이 이김) | 지원 |
| Agents | ~/.claude/agents/ | managed > CLI > **project > user** > plugin | — |
| Rules | ~/.claude/rules/ | user 먼저 로드, project 나중(우선) | 지원 (문서 직접 예시) |
| CLAUDE.md | ~/.claude/CLAUDE.md | user → project concat | — |

핵심: skills만 방향이 반대 — 글로벌 skill은 동명 프로젝트 skill을 영구히 가림 → C안 기각의 결정 근거.

## 종합 매트릭스

| 축 | A 현행 | B 플러그인 | C 전면 글로벌 | D 심링크 |
|---|---|---|---|---|
| 설치 마찰 | 낮음 | 2-스텝 | 중간 | 낮음 |
| 업데이트 UX | 프로젝트마다 | 일괄 | 일괄 | 일괄 |
| 버전 스큐 | 없음 | 신규 발생 | 핀 불가 | 없음(버전 디렉터리 핀) |
| 프로젝트 커스터마이징 | 완전 | read-only | skills 역전으로 불가 | 완전 |
| 비-moai 프로젝트 영향 | 없음 | enable 범위 | 전 프로젝트 오염 | 없음 |
| 팀 공유 | clone만 | 플러그인 설치 필요 | 머신별 셋업 | clone + init 1회 |
| git 노이즈 | 수백 파일 | 최소 | 최소 | 심링크만 |
| 발견성 | 없음 | 있음 | 없음 | 없음(게이트웨이 플러그인 보완) |
| Windows | 동일 | 동일 | 동일 | 복사 폴백 필요 |

## D안 구조

```
~/.moai/assets/v3.1/{skills,agents,rules}   # 실체 1벌, moai update가 여기만 갱신
project/.claude/skills/moai -> ~/.moai/assets/v3.1/skills   # 심링크 (.gitignore)
project/.claude/rules/moai  -> ~/.moai/assets/v3.1/rules
project/.claude/skills/hns-*/                # user-owned 실파일 공존
```

- 중복 제거 + overwrite 사고 클래스 소멸 + 스코프 오염 없음 + 프로젝트별 버전 핀
- 약점: Windows 심링크(복사 폴백), 팀원은 clone 후 `moai init` 1회, 기존 프로젝트 마이그레이션

## 오염 근거 (C안)

- ~/.claude 자산은 머신 전 프로젝트 로드: skills 메타 ~100tok × 수십 개 + always-loaded rules 수만 토큰이 비-moai 프로젝트에도 매 세션 주입
- `paths:` 스코핑은 "moai 프로젝트인가" 조건 표현 불가; `claudeMdExcludes`는 비-moai 프로젝트마다 설정 필요 → 목표 역전
