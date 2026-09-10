# 카드 t291 판정 — codemaps 스탬프 도달성 가드 + 명시적 --commit 모드

> Evidence-bearing report (AGENTS.md §1). 작성: factory lane-4, 2026-08-27.
> 브랜치 `WT-stamp-orphan-guard`(워크트리 `.claude/worktrees/t291`) · develop 기점(da791eb0a 계열, 기존 진행 이월) · 병합 전 커밋 완료 상태.

## Claim (주장)

SPEC-STAMP-REACHABILITY-001의 세 마일스톤이 TDD로 구현·검증을 마쳤다: M1 `--commit` 명시적 스탬프 모드, M2 CI 머지-전 스탬프 도달성 가드, M3 회귀 잠금 + 운영자 문서(4개국어). 싱크 감사는 4차원 바인딩 PASS(조화평균 0.912 ≥ 0.85)이며, 감사 지적 신규코드의 오류 접두사 중복은 마감 전 수리했다. 추적 중인 provenance 스탬프는 한 번도 다시 찍히지 않았다(REQ-GFR-014 유지, 인스턴스 수리는 본 카드 범위 밖).

## Evidence (증거)

독립 검증 배치(오케스트레이터 직접 재측정, 본 트리에서 관측):

| # | 항목 | 명령 | 관측 |
|---|---|---|---|
| V1 | 커밋 이력·트리 | `git log --oneline -6 ; git status --porcelain` | 4개 실행 커밋 전부 t291 태그, 트리 클린 |
| V2 | 영향 패키지 테스트 | `go test ./internal/cli`(무파이프 재측정) | `ok … internal/cli 181.723s`, exit 0 |
| V3 | 가드 스텝 해부학 | `sed`/`grep` on shipped YAML | checkout L26→가드 L29→setup-go L66, GITHUB_BASE_REF L43, 빈값 분기 L57, `::error::` 조상 실패 L61-62, continue-on-error 부재, 트리거 블록 미변경 |
| V4 | --commit 플래그 표면 | grep | Long text 레시피+예제 존재, dirty 사전거부 양단 구현 |
| V5 | SPEC 전이·증거 | frontmatter+ls | version 0.2.1, status in-progress(본 커밋에서 최종 전이), 증거 실물 존재 |
| V6 | 4개국어 패리티 | `grep -c -- '--commit' ×4 locale` | en/ja/ko/zh 각 3건 일치 |

싱크 감사(FO-SYNC-1 4차원, 읽기전용 판사 4+컨텍스트 1): PASS · harmonic_mean 0.912 · threshold 0.85 · 발견 10건 전부 minor · critical 0건 → SPEC-AUDIT-SNAPSHOT-001 A3 바인딩 승격 적용, 콜드 sync-auditor 미소생. 부록으로 존재하는 명시적 Gap 16건은 감사 산출물에 기록됨(읽기전용 제약에 따른 픽스처 대체 실험 방식).

접두사 수리(ad33cc9f4): `graph stamp:` 균일 접미 래퍼로 통일, mx 자기참조 접두사 제거. 국소 검증 `go test -count=1 -run 'TestGraphStampCmd_|TestStampCodemaps_|TestResolveCommit' ./internal/cli/ ./internal/mx/` → 양 패키지 ok(4.7s/3.3s).

마일스톤 커밋: cd71c701e(M1) · e4eb15ea4(M2) · 2378bc14c(M3) · e504c8c08(증거 백필) · cbd33980a(모드 선택 기록) · ad33cc9f4(접두사 수리).

## Baseline-attribution (baseline 귀속)

모든 관측값은 2026-08-27 본 워크트리(`.claude/worktrees/t291`)의 위 커밋 목록 상태에서 이번 세션에 실행한 명령의 출력이다. 싱크 감사 수치는 동일 트리에서 돈 동적 워크플로(run wf_b8af067e-a0b) 결과 JSON에서 인용했다.

## Gaps (미검증)

- 라이브 GitHub Actions 실행은 어디서도 아직 관측되지 않았다(YAML 문법 검증+셸 드라이런만). 통합 후 origin/develop 첫 실행이 증인이다.
- **예측(주장 아님)**: 알려진 라이브 오온 인스턴스(`a995e58fa69b…`, main 비조상 — 리드 트리아지 소관)가 수리되기 전이라면 그 첫 실행에서 가드 RED(codemaps 행)가 올바르게 발화한다. 이는 결함이 아니라 새 계측기의 자기입증이며, merge-base 리스탬프(새 `--commit` 레시피)로 해소된다.
- darwin/windows 크로스빌드는 로컬에서 생략(레인 로컬 검증 규율 — CI 전담).

## Residual-risk (잔여 위험)

- 오온 인스턴스 미수리 상태로 릴리스 흐름이 진행되면 codemaps red가 상속된다 — 리드 트리아지를 통한 post-merge 리스탬프가 해소 경로다.
- described-source churn 현재 30/40(임계 미달)이나, 머지 직전 대량 편집이 임계를 넘기면 스펙 §D대로 red 수용 후 POST-MERGE main 리스탬프로 넘긴다.
