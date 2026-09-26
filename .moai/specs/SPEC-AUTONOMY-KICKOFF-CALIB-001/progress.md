# progress.md — SPEC-AUTONOMY-KICKOFF-CALIB-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-26
- spec_version: 0.1.0 (initial plan-phase draft)
- tier: M (spec.md, plan.md, acceptance.md, research.md — 측정 설계 카드, 코드 산출물 없음)
- requirements: 10 (REQ-CALIB-001..010, 연속) / acceptance criteria: 10 (AC-CALIB-001..010, 1:1; [PLAN] 10건 중 AC-CALIB-009는 지금 실행 가능, [RUN] 이행 검사는 run 기록 대상)
- card: t1244 (AUTONOMY-A5 — Kickoff 판단 모드 보정)
- measurement_target: `workflow.autonomy.kickoff.jev_min_confidence` (A1 0.5.2 정의, 기본 0.50; 소비처 A3 R2 `jev_low_confidence`)
- population: 과거 운영자 Implementation Kickoff Approval 결정 라운드 — 세션 전사본 코퍼스(루트 A 335디렉터리 + 루트 B 4디렉터리, 2026-09-26 실측)
- extraction_predicate_feasibility: 후보 148건·응답 짝 148/148(100%)·세션 87파일·오류 0건 — 프루브 `.moai/reports/t1244/probe/`(스크립트+출력, 2026-09-26 본 카드 실행). 같은 날 두 실행 사이 +1 드리프트(906→907) — 코퍼스 생존 확인
- preregistration: 판정 기준(밴드 n≥20·기준선+10%p·wrong-automation≤10%·최저 자격 밴드 채택·무자격 시 기각), LIVE 상한(항목당 2회·총 상한 산식·벽시계 8시간/배치·배치 1회), 대조 두 팔(파이프라인 20/20 + 판사 5건, 양팔 무효 규칙), 라벨 매핑(4값+compound) — 전부 측정 전 본문 고정
- t943_discipline: §30 전문 정독 완료. 비재실행 문장(REQ-CALIB-010), §30 인용 시 상시 상수 기준선 행 동반, 기각은 「판정으로 쓰는 것」 한정이며 A5 긍정 결과의 적용처는 REQ-GR-013 한 곳으로 좁음
- a1_reference_baseline: 25283ebf8 (A1 0.5.2); 관측 2026-09-26 — A1 SPEC 마감 `e4ea8eb05`·코드 `internal/contract/`(최상위 `fbd277ab7`)가 본 브랜치에서 도달 가능. 리드 디스패치의 「develop 미병합」 서술과 갈림 — 전제는 명령으로 판정(`git merge-base --is-ancestor e4ea8eb05 HEAD`, 현재 0)
- a3_reference: 0.3.3 draft (t1236 워크트리) — A5 비차단 배정 문구·R1–R5·REQ-GR-013/025 인용
- linked_pr: 없음 (리드 확인 2026-09-26 — 대조할 PR 없음)
- judge_candidate: Jev `ask.sh` — `/Users/goos/MoAI/moai-adk-go/scripts/jev/ask.sh`(primary 체크아웃, **git 미추적** — `git ls-files` 카운트 0 실측). 프로토콜은 판사 무관
- run_preconditions: (a) A1 병합 확인 명령 exit 0 — 현재 성립, 실행 시 재확인. (b) 운영자 Implementation Kickoff Approval(레인 창). A3 병합은 전제 아님. `scripts/jev`·키 부재는 전제 미충족이 아니라 측정 불가 판정서 갈래
- open_clarifications: NC-1(밴드 기준 수치는 측정 전 판단값 — 기본값 유지), NC-2(카드 단위 보고는 보조 — 기본값 유지). 둘 다 본문 결정이며 운영자가 측정 전에만 바꿀 수 있다
- scope_guard: 제품 코드·`internal/`·템플릿 미러 변경 없음 — 측정 도구는 전부 `.moai/reports/t1244/`(미추적) 아래 일회용. AC-CALIB-010 이 변경 집합을 검사한다
- plan_artifact_hash: (plan-audit 진입 시 기록)
- Implementation Kickoff Approval: not requested at plan phase

## §E.2 Run-phase Evidence

_<pending run-phase — A1 병합 확인 재실행 결과, 운영자 Kickoff 승인 일시, 스냅샷 일자·스윕 명령, 모집단 크기, 대조 두 팔 결과, 판사 배치 상한 선언문이 여기 기록된다>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
