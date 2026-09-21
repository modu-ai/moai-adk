# t274 verdict — v3.1.3 전체 업데이트/추가 기능 문서 반영 (SPEC-DOCS-V313-CATCHUP-001)

- **Card**: t274 (Class C · Tier M) · **SPEC**: SPEC-DOCS-V313-CATCHUP-001 (status: completed)
- **Branch**: `WT-v313-docs` · **PR**: #1662 `docs(t274): v3.1.3 documentation catch-up — SPEC-DOCS-V313-CATCHUP-001 (4-locale)` (base main, MERGEABLE)
- **Date**: 2026-08-26 · **Lane**: lane-4 (후속 세션 — /clear 이전 세션과의 인계·흡수 경위는 아래 기록)

## Claim

카드 지시 4요건 전부 이행: (1) CHANGELOG [3.1.3] 26항목 전수 추출 (2) 격차 표 작성 — D 3 / U 11 / N 4 / NA 8 + version SSOT 갭 V1–V8 (3) 미반영 항목 문서화 — U 11 기존 페이지 갱신 + N 4 신규 페이지 `advanced/codex-dual-harness.md` 4로케일(운영자 승인) + main.yaml 내비 항목 (4) 4로케일 동시 반영 + oss-docs verify 종료 게이트 통과. **선결함 version-sync는 격차 표 §1.4 V1–V8로 첫 항목화해 해소** — hugo.toml v3.1.2→v3.1.3(+releaseDate) 정렬, 전 버전 표시 동조.

## Evidence (요약 — 전문은 progress.md §E.2)

- **AC**: AC-DVC-001..009 = 9 PASS / 0 FAIL (§E.2 AC 매트릭스, 관측 명령+출력 원문)
- **종료 게이트**: hns-oss-docs-verify 8축 전부 green — 병합 트리 `0044c7a83`에서 재검증 포함 (§E.2 병합 재검증 표)
- **sync-audit**: PASS 0.937 (F97·S95·C90·Cons93, 차단 0) — `.moai/reports/t274/sync-audit-verdict.md` (측정은 `bed33bbde` 핀)
- **CI (PR #1662)**: Lint·Build×5·Analyze·4-locale parity pass; graph-freshness fail — **main 선재 레드 상속, required 아님** (아래 Residual)
- **diff 범위**: 64파일 +1370/−264 — docs-site 54 · README×4 · CHANGELOG 1 · `.moai` 5 (SPEC 산출물+감사verdict). Go/템플릿/훅 0파일 (AC-DVC-006)

## Baseline-attribution

- 격차 표 관측: baseline `e07a6d0f4` (plan-phase) → run pre-flight `311d5498a` 재관측 (AC-DVC-001)
- M5 종료 게이트: `5d68cdac9` → 병합 후 `0044c7a83` 전 축 재측정
- sync-audit: `bed33bbde` 핀 측정 (레이스 방어 — 감사 중 병합 감지, Race note 참조)
- sync close: `bc87bc9ca` (3-phase close) + backfill `07a5c6185` (sync_commit_sha) + 감사verdict 커밋 `d99f4f24e` (PR head)

## 세션 경위 (레이스 기록 — 프로세스 부채)

리드 디스패치는 "신규 워크트리"로 알렸으나 실제는 **/clear 이전 lane-4 세션이 plan(감사 2회)~run M1–M5까지 진행 중이었음**. 후속 세션(본 세션)이 조사 중 선행 체인의 M5 커밋(`bed33bbde`)이 착지 — 운영자 승인으로 흡수 후 origin/main 4커밋(t269·t250·t259·t273) 병합·8축 재검증·sync close 수행. 이후 선행 체인의 잔여 sync-audit→PR 단계가 병렬 재개되어 sync-audit verdict 커밋(`d99f4f24e`)·push·PR #1662 생성까지 완료 (두 작성자는 전 구간 순차 스태킹, divergence 0; 후속 세션의 독립 감사 에이전트는 스트림 스톨로 부분 진행 후 중단 — 소견은 선행 감사와 일치). **근본 원인: 카드 재배차 시 선행 세션의 생사·진행 상태 확인 절차 부재** — 리드 보고 완료.

## Gaps

- PR #1662 머지 미완료 (required 체크 잔여 대기 중 — 머지 SHA는 머지 후 확정)
- 좀비 체인의 잔여 활동 가능성은 배제 못함 (마지막 관측: PR 생성. 이후 추가 쓰기 없음 확인 시점 기준)
- sync-audit 측정 트리가 병합 전 `bed33bbde` — 병합 트리 커버는 후속 세션의 8축 재검증이 담당 (감사 Gaps에도 명시됨)

## Residual-risk

- **graph-freshness 게이트 main 레드 (선재, 본 카드 비관련)**: `6786c3fa4`(t250 #1648) squash 머지가 `.moai/project/codemaps/provenance.json` 스탬프의 원본 커밋 `0d15864ae90b`(origin/WT-graph-freshness에만 존재)을 고아화 — main 최근 2커밋 연속 FAIL, 모든 PR 상속. required 목록(5종)에는 없어 머지는 불가 아님. **신규 카드 권장** (t279 t250-후속에 접목 가능)
- 감사 F1(ko 신규 페이지 강조표기 4곳 마커 내부 괄호 — cosmetic)·F2(sitemap 바이트 수 형식 차이) 미고침 — optional 판정 그대로
- 백로그: version-sync 프로세스 근본 원인(릴리즈 시 표시 일괄 갱신 누락)은 별도 카드 권장 (spec.md §E.1 notes)
