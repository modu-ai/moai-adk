# Progress: SPEC-DOCS-CODEX-WIRING-CALLOUT-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-08
plan_phase_artifacts: spec.md, plan.md, acceptance.md, progress.md (Tier M — 3+1)
baseline_tree: a849d99d2
baseline_branch: WT-docs-codex-callout
plan_phase_measurements: spec.md §1 전체(부재·대조군·콜아웃 관례·doctor 동작·버전 귀속·범위 판정)는 이 워크트리에서 이번 실행으로 재측정됨 (VCI §2)

## §E.2 Run-phase Evidence

실행: 2026-09-08, 워크트리 t535 @ `4bb6503a4`. 실행 수단: hns-oss-docs 하네스 스페셜리스트(ko content-author → en translator → ja·zh translator 병렬) — `hns-oss-docs-run` 워크플로우 러너는 저장 스크립트 문법 오류(Unexpected keyword 'export')로 기동 실패하여 스페셜리스트 직접 스폰으로 동일 페이즈(scope→author→translate→verify) 수행. 모든 검증은 오케스트레이터가 하네스 보고와 무관하게 직접 재측정(리드 지시).

| AC | 판정 | 명령 → 관측 (이번 실행, 이 트리) |
|----|------|--------------------------------|
| 001 | PASS | `/usr/bin/grep -rn 'Codex Wiring' docs-site/content/` → 8히트(로케일당 H2+본문 굵은 2), exit 0 — RED-now(0히트 exit 1) 역전 |
| 002 | PASS | `/usr/bin/grep -c 'Home Disk Usage'` ×4 → 2/2/2/2 유지 (기존 절 무훼손) |
| 003 | PASS | `/usr/bin/grep -c '^## '` ×4 → 8/8/8/8 (7→8) |
| 004 | PASS | 헤딩 4종 확인(ko 진단/en check/ja 診断/zh 诊断, 배지 verbatim) + `/usr/bin/grep -rn 'v3\.1\.3'` ×4파일 → 0히트 exit 1 |
| 005 | PASS | 절 순서 4로케일: Hook Delivery(53/55) < Codex Wiring(65/67) < 종료코드(93/95) |
| 006 | PASS | `/usr/bin/grep -n 'advanced/codex-dual-harness'` ×4 → 로케일 접두 링크 각 1 + 대상 파일 4개 `ls` 실재 |
| 007 | PASS | 지시문 4종 토큰 ×4: `moai init --agent codex` 1×4 · `codex /hooks` 1×4 · `moai update --templates-only --force --yes` 1×4 · 스테일 조언 문장 존재 / `moai clean --codex-skills` 1×4 — 전부 별도 동사 서술로만 |
| 008 | PASS | `enabled = ` 1×4 · `0.153.4` 1×4 · "유일한 fatal"+나머지 조언형 서술 4로케일 정독 확인 |
| 009 | PASS | un-nagging 문장 4로케일 정독 확인 (조용한 정보성 스킵) |
| 010 | PASS | 토큰 로케일 간 동일(위 007·008 카운트) |
| 011 | PASS | 이모지 perl 스캔 0히트 · URL 블랙리스트 0히트 exit 1 · Mermaid LR/RL 0히트 exit 1 · 강조 간격: 신규 절 위반 0 (스캔 10히트 전수 확인 결과 기존 절 위반 + 패턴 허위양성뿐) |
| 012 | PASS | `hugo --minify` exit 0, WARN/ERROR 0행 (`/usr/bin/grep -c -e WARN -e ERROR` → 0) |
| 013 | PASS | `git diff --stat` → doctor.md ×4 (+28×4) + progress.md(라이프사이클 기록 — REQ-DWC-010 개정 예외) |
| 014 | PASS | 4파일이 같은 커밋 후보 변경 집합에 존재 |

정직 기록: ① 강조 간격 스캔 1차 실행은 `cd docs-site` 지속으로 상대경로 실패(빈 출력) — 절대경로 재실행으로 성립. ② 신규 절 밖 기존 위반 3곳(ko:41 `**권고(advisory)**`, ja:39, zh Home Disk Usage 절)·zh 예시 절의 `doctor permission`/`sandbox` 누락은 범위 밖 관측으로 기록 — t538 스윕 후보. ③ `hns-oss-docs-run` 러너 결함은 별도 카드 후보로 리드 보고.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-09-08
ac_matrix: 14 PASS / 0 FAIL / 0 PASS-WITH-DEBT (§E.2)
write_surface: docs-site/content/{ko,en,ja,zh}/cli-reference/doctor.md (+28×4)

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관>_

## §F Phase 4 Mode Selection

- 입력 파라미터: tier M · scope 4파일(doctor.md ×4) · 도메인 1(docs-site 콘텐츠) · 언어 믹스 markdown 100% · 동시성 이익 LOW(코딩/문서 파생 체인) · 킥오프 승인: 운영자 승인 접수(리드 경유, 2026-09-08)
- 모드 평가: direct(아님 — 산출물이 사용자 문서 4파일, 검토 필요) / fanout(아님 — ko→en→ja·zh 체인 의존성 있음) / sweep(아님 — 기계적 균일 변환이 아님) / **serial(선택)** — 마일스톤당 1 스폰 체인
- Decision: **serial** (hns-oss-docs-run 하네스가 translate 페이즈에서 로케일별 파생 워커를 내부 fan-out하는 것은 하네스의 위임 구조이며, 본 세션의 직접 스폰은 순차 유지)
- 근거: 파생 체인(ko 정본 → en → ja/zh)에 선행 의존이 있어 coding/docs 병렬화 경고가 적용된다. M4 검증 스윕은 오케스트레이터가 하네스 보고와 무관하게 직접 재측정한다(리드 지시).
- kick-off 전제 재측정(2026-09-08, `4bb6503a4`): 부재 0히트 exit 1 유지 · H2 7×4 · 태그 `v3.1.2` 까지 · CHANGELOG `[Unreleased]` 8행·WIRING-001 239행 — 배지 `v3.1.4` 전제 유효.
