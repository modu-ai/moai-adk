# 카드 t293 진행 기록

- **카드**: t293 · 이슈 modu-ai/moai-adk#1675
- **레인**: factory lane-2 (GLM 백엔드)
- **워크트리**: `.claude/worktrees/t293` (origin/main 팁 `3abde7053` 기점)
- **브랜치**: `WT-statusline-profile` (생성 직후 개명 완료)

## 조사 결과 (plan-phase 근거, 실측 2026-08-27)

### 결함 A — 운영자 옵트아웃이 다른 프로젝트에 적용됨
2026-08-17 지시분(`segments.github: false` + `forge: none`, 한국어 주석 포함)은
`/Users/goos/MoAI/mo.ai.kr/.moai/config/sections/statusline.yaml` 43~57행에 존재.
moai-adk-go의 같은 경로 파일(33줄)에는 어느 키도 없음.

### 결함 B — 워크트리 프로필 매핑 누락
`lookupProjectKey`(`internal/profile/profile.go:297`)는 exact-match + SameFile 별칭만
확인하고 상위 프로젝트로 거슬러 올라가지 않음. launch.yaml에는 과거 워크트리(t267·t289·
release-v313)만 개별 등록돼 있어 새 워크트리는 무명 프로필로 하락.
관측: 이 레인 세션 `CLAUDE_CONFIG_DIR=~/.moai/claude-profiles/__no_such_profile__`.

### 결함 C — rate-limit 회귀 경로 부활 구조
`forge` 미설정 시 origin 호스트 자동 판별(github.com → `gh` 폴링, TTL 10분,
`internal/statusline/github.go`). 명시적 off 수단이 유실되면 전 세션에서 폴링 재개.

## 이슈 표제와 실체의 차이
"폴백 프로필이 statusline 설정을 무시"는 결과 증상 — 렌더러가 읽는 프로젝트 파일에
off 키 자체가 없어 어느 프로필에서든 github 세그먼트가 켜진다.

## Gaps (미검증, SPEC에 Gap으로 남김)
- 리터럴 디렉터리명 `__no_such_profile__`의 주입 원천 — 레포·깃 이력·설치 바이너리·
  로컬 설정·tmux 전역 환경 모두 부재 확인. 상위 프로세스 체인 상속으로 추정하나
  본 레인에서 관측 불가. 코드측 수정(조상-walk 매핑)은 이 미스터리와 무관하게 동작.
- "0/0" 리터럴 렌더 경로 존재 여부(Available=false 설계상 "-/-") — run-phase 재확인.

## plan-auditor iter-1 결과 (2026-08-27)
- VERDICT **FAIL** 0.88 — Tier M 임계 0.80 상위이나 MP-7(미확정 마커) 필수통과 실패
- D1 BLOCKING: plan.md:130 `[NEEDS CLARIFICATION]` / D2 SHOULD-FIX 이하 4건
- REQ-001 전제(builder.go:262 무조건 spawn) 코드 대조로 실측 참 확인

## 승인 게이트 결정 (운영자, 2026-08-27)
1. 조상-walk = 읽기 경로 전용, 쓰기 정규화(REQ-009)는 후속 카드 이연
2. 익명 세션 github 세그먼트 기본값 변경 없음(명시적 설정으로만)
3. 본 저장소 로컬 statusline.yaml `forge: none` 편집을 run 산출에 포함
4. SPEC 수정 → iter-2 델타 재감사 → manager-develop run 착수 승인

## 현재 단계
✅ run 완료(HEAD `0a92b22f9`, 4커밋) — lane 독립 검증 9/9 PASS, 판정은 `verdict.md`.
병합·push 없음(리드 지시). 리드 완료 보고 송신, 지시 대기.
