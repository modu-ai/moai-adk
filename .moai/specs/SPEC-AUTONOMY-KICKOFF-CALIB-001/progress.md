# progress.md — SPEC-AUTONOMY-KICKOFF-CALIB-001

> 본문 판정 기록(2026-09-26, 저작 당시): [PLAN] AC 10건 전부 재실행 통과 — AC-CALIB-001..008·010 grep 카운트 양성, AC-CALIB-009 `git merge-base --is-ancestor e4ea8eb05 HEAD` exit 0.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-26
- spec_version: 0.2.0 (리드 중간 지시 2건 접어넣음: REQ-CALIB-011 아웃바운드 스크럽·REQ-CALIB-012 자율 Kickoff 세 조건 — REQ-CALIB-009 (b) 대체. 0.1.1 lint 수리, 0.1.0 초안)
- lint_status: `moai spec lint SPEC-AUTONOMY-KICKOFF-CALIB-001` → `✓ No findings` (2026-09-26, 설치본 `~/go/bin/moai` — 0.1.1 시점 기준, 0.2.0 커밋 전 재실행 결과를 HISTORY 와 함께 기록)
- tier: M (spec.md, plan.md, acceptance.md, research.md — 측정 설계 카드, 코드 산출물 없음)
- requirements: 14 (REQ-CALIB-001..014, 연속) / acceptance criteria: 14 (AC-CALIB-001..014, 1:1; [PLAN] 11건 — AC-CALIB-009는 지금 실행 가능, [RUN] 3건 AC-CALIB-012·013·014는 run 기록 대상 `--- PENDING-RUN`)
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
- run_preconditions: (a) A1 병합 확인 명령 exit 0 — 현재 성립(앵커 `e4ea8eb05`, 리드 전달 착지 헤드 `b1a62fb2b` 모두 HEAD 도달 가능 확인), 실행 시 재확인. (b) 게이트 처분 — 운영자 결정 2026-09-26(리드 경유) 「모든 킥오프 승인은 자율적으로 진행한다」: 이 카드 측정 실행은 plan-audit 뒤 자율 진행, 세 묶음 조건(사전 등록 커밋·세 중복 상한·스크럽 게이트) 구속 — 배차 범위 기록, 독트린 일반화 아님. A3 병합은 전제 아님. `scripts/jev`·키 부재는 측정 불가 판정서 갈래, 스크럽 실패는 전송 차단(fail-closed) 갈래
- outbound_scrub: REQ-CALIB-011 — 스크럽 4부류(키·settings.local·절대 경로·고객 데이터)·데이터 최소화(세션 덤프 금지)·전송 전 기계 스캔·더미 비밀 양성 대조 선행·적중 차단 fail-closed. §29 「키와 경계」의 코퍼스 확장
- autonomous_kickoff_provenance: 운영자 결정 2026-09-26 (리드 경유) — REQ-CALIB-012·spec.md §A.2·plan.md §C 기록. 세 조건의 기계 판정 AC: AC-CALIB-013(criteria_commit이 첫 판사 호출에 선행)·AC-CALIB-014(세 상한 declared_at 선행·상한 내 집행)
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
