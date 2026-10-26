# t1968 — jevfetch 이식 (lane-7)

Card: t1968 · Class C (plan→run→sync) · P1 · parallelizable · run tnbn7e
Dispatch: 2026-10-26 리더(u cust socket 50218) — cmd: plan → run → sync; wt: factory next 생성(선-생성 금지); evidence: 본 파일 + card-review.md + sync-audit.md + verdict.md
Worktree: `.moai/worktrees/t1968` · branch `WT-jevfetch-a2sys-ai`

## 1. Intake (2026-10-26 20:4x KST)

- 리스: `moai factory next --card t1968` — 첫 시도 AMBIGUOUS_FACTORY(런 5개), live owner 유일 런 tnbn7e 명시 후 성공(`t1968 stage=- worktree=t1968`). 진입 검증: `git rev-parse --show-toplevel` 일치, 트리 clean — 직접 관측.
- Base 흡수: 트리가 cb30e05b3e(t1964) 기준 생성 — 배차 전제(origin/main 2a1898ea22)보다 1 landing 뒤. `git fetch origin main && git merge --no-edit origin/main` → **fast-forward, 충돌 0**(8 files changed; t1921 착지 내용 포함). 관측 출력: `Updating cb30e05b3e..2a1898ea22 / Fast-forward`. HEAD=2a1898ea22.
- 설계 원문: 1차 체크아웃의 리더 미추적 파일 → 트리 로컬 복사 `.moai/reports/leader-plan-20261023/jevfetch-analysis-20261026.md`(3,664B). 전문 확인: 적용안 1~4, 미가져올 것(Python 의존·헤드리스 렌더 — t1965 연동 후속), Gaps(벤치 미재현·비용 기본 off·토큰 계수 근사).
- 재점검 크론: 3347961f(`*/5 * * * *`) 가동 중.
- 팩토리 기록 직독: t1921 state=merged-pr(전 카드 종결 확인), t1968 state=assigned → 리스.

## 2. Plan stage

- Spawn: manager-spec(depth-1 leaf, frontmatter 기본 모델/effort, 무격리 — 레인 트리 착지 예정, 귀환 후 착지 검증 예정). 프롬프트 축약: 적용안 1~4 전문 + 제약(새 키·비용 금지 기본 off / Python 의존 금지·토큰 근사 / 템플릿 중립성 / internal/jev는 클라이언트 캐시·재시도만, t1762=장부·consumers 경계) + 완료 판정(블록 분할 골든·채점 선택 규칙·출력 형식·종료 코드 단위 테스트, 페이크 Jev 서버 E2E 1건, jev off 네트워크 0건 테스트) + GEARS 2-cell 기준.
- 2026-10-26 20:4x KST 시작 — 진행 중.
