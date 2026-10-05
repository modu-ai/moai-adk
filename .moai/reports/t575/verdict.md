# Card t575 Verdict — docs-site 4로케일 moai-todo 임시 루트 가드 문장 수리

- Date: 2026-09-13 · Lane: lane-4 (Factory) · Branch: `WT-docs-todo-locales` (unpushed)
- SPEC: `SPEC-DOCS-TODO-TEMP-GUARD-001` (draft → completed, 3-phase close, v1.1.0)
- Base: e83b4ec20 계열 (리드 지명 9bfe424ee 포함, origin/develop lineage)

## Claim

docs-site 4로케일 `utility-commands/moai-todo.md`의 무조건 홈 경로 문장(59·248행)이 임시 루트 가드(SPEC-TODO-HOME-TEMP-GUARD-001)로 거짓이 된 것을, 코드 검증 진리표 기반의 조건부 문장으로 4로케일 동시 수리했고, 동종 결함 census 21건을 기록했으며, sync-auditor 독립 감사 94/100 PASS.

## Evidence

| Phase | Result | Evidence |
|---|---|---|
| plan | 진입 PASS 1.00 (0.94 → D1 lifecycle 열거값 수리 → 1.00) | `.moai/reports/t575/plan-audit.md` (리드 요청으로 영속화) |
| run | AC 7종 실측 PASS, 커밋 `55adec04f`·`7ea3fb026` | AC-002 carve-out 2/로케일 ×4, AC-003 구248형 0건 ×4, AC-004 행시작 앵커 0건 ×4(이중 프로브 기록 §E.2), AC-005 헤딩 29/29/29/29, AC-006 census 21건, AC-007 범위 격리(diff에 Go/템플릿/README 없음) |
| sync | 커밋 `03d72705b`(3단계 close)·`acd8b4b40`(sha 백필)·`d248fb45b`(F1) | CHANGELOG Unreleased→Fixed 1항목, frontmatter 전환 3회 모두 정합 |
| sync-audit | **PASS 94/100** (기능 95/보안 100/품질 90/일관성 92) | 7 AC 전수 재측정 + **새 문장 8건 전부 코드 대응 진위 확인**(state_dir.go:26-41, temp_origin.go:60-62 트리플렛 일치, todo_root.go:135-144) + §E.2 기록의 git show 바이트 대조 |

- 감사 보고서: `.moai/reports/t575/sync-audit.md`, census: `.moai/reports/t575/docs-census.md` (브랜치 커밋)
- 재측정 범위: `e83b4ec20..HEAD` (docs 4파일 + census + CHANGELOG + SPEC 디렉터리만)

## Baseline-attribution

모든 수치는 이번 레인 세션에서 이 워크트리에서 실행한 명령의 관측값. hugo 빌드는 미실행(CI/Vercel 몫 — 기록된 gap). push는 리드 일괄.

## Gaps

- hugo 사이트 빌드 미실시(AC 범위 밖, CI 소관)
- census 9-12행 TRUE 분류는 존재 확인 수준(탐색 순서 의미 심층 재유도 안 함)
- factory 경로 진위 — 설계상 미검증(후속 t706)

## Residual-risk

- t705(StateDirForRoot 임시-git 변칙 수리)가 착지하면 이 페이지 문구 1회 추가 정정 필요할 수 있음(감사 잔여 위험)
- README 4파일의 factory `.db` 문장(80행)은 t706 소관으로 미수정

## 후속 카드 (리드 발행 완료)

t704(템플릿 todo-queue-storage.md 동종 문장)·t705(StateDirForRoot 변칙)·t706(factory 경로 검증) — 이 카드에 미포함
