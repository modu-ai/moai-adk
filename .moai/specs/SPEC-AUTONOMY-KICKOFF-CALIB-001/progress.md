# progress.md — SPEC-AUTONOMY-KICKOFF-CALIB-001

> 본문 판정 기록(2026-09-26, 저작 당시): [PLAN] AC 10건 전부 재실행 통과 — AC-CALIB-001..008·010 grep 카운트 양성, AC-CALIB-009 `git merge-base --is-ancestor e4ea8eb05 HEAD` exit 0.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-26
- spec_version: 0.2.1 (plan-audit iter-1 수리 D1–D4 + 부수 MINOR D7–D13; 0.2.0 리드 중간 지시 2건 접기 — REQ-CALIB-011 아웃바운드 스크럽·REQ-CALIB-012 자율 Kickoff 세 조건; 0.1.1 lint 수리; 0.1.0 초안)
- plan_audit: iter-1 **PASS-WITH-DEBT 0.83** (Tier M 문턱 0.80, BLOCKING 0·SHOULD-FIX 4·MINOR 9) — 보고서 `.moai/reports/plan-audit/SPEC-AUTONOMY-KICKOFF-CALIB-001-review-1.md`, 감사 대상 `7502b84b4`. D1–D4·D5–D13 처분: HISTORY 0.2.1. artifact hash 는 이 수리로 이동 — 재심사 시 재-pin
- lint_status: 0.2.0 트리에서 `✓ No findings`(감사 Claim 5 독립 재실행과 저자 실행 일치); 0.2.1 수리 커밋 직후 재실행 결과를 HISTORY 0.2.1 다음 기록으로 남긴다 — 0.2.0 때 예고만 남긴 것(D10)의 이행
- tier: M (spec.md, plan.md, acceptance.md, research.md — 측정 설계 카드, 코드 산출물 없음)
- requirements: 12 (REQ-CALIB-001..012, 연속) / acceptance criteria: 14 (AC-CALIB-001..014; AC-013·014 는 REQ-CALIB-012 를 함께 커버 — 감사 D5 가 짚은 0.2.0 오기 「14 REQ」의 수리; [PLAN] 11건, [RUN] 3건 AC-CALIB-012·013·014는 run 기록 대상 `--- PENDING-RUN`)
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

실행 주체: 레인 오케스트레이터 직접(카드 t1244, 세션 359abee0 — 측정 프로토콜 이행, 제품 코드 0).
전 증거는 `.moai/reports/t1244/`(미추적) — run-record.md(사전 등록·스캔 행·수리 경위)·extract/(모집단)·runs/(판사 원시)·judge_control.jsonl·judge_main.jsonl·scrub_scan.py·judge_batch.py·verdict.md(판정서 5섹션).

- **게이트 진입(REQ-CALIB-009·012)**: A1 앵커 `git merge-base --is-ancestor e4ea8eb05 HEAD` exit 0·`b1a62fb2b` exit 0(실행 시 재확인). 자율 Kickoff 세 조건 이행 — criteria_commit `2e0ef3c76` 선언 2026-09-26T17:37+09:00(첫 판사 호출 선행), 세 상한 turn=2/호출·call=2×(모집단+5+재결)·벽시계 8h declared_at 2026-09-26T17:37:30+09:00, 스크럽 게이트 통과.
- **스크럽(REQ-CALIB-011, AC-CALIB-012)**: 스캐너 양성 대조 8/8 발화(sk-/ghp_/AKIA/PEM/절대경로/이메일/전화/주민번호)+클린 대조 OK — 첫 전송에 선행. 전 페이로드(110건) 스캔 `returncode: 0`·적중 상태 전송 0건(scrub_log.jsonl).
- **M3 모집단(REQ-CALIB-001)**: 스냅숏 2026-09-26·양쪽 코퍼스 루트 전체 스윕. 1단계 후보 173(응답 짝 173/173) → 게이트 라운드 **105**(세션 60, 카드 토큰 155종). X1 50·X2 27(전부 "No response after 60s" 폴백 — 운영자 미응답, 원문 보존). 추출 중 계측기 결함 2건 수리(응답 파서 형태 오가정·finditer 언팩킹) — 술어 불변.
- **라벨(REQ-CALIB-002)**: approve 104 / modify 1 / hold 0 — hold 어휘 답 실측 0건으로 매핑 누락 아님 검증(AC-CALIB-004 파이프라인 팔).
- **M4 판사(REQ-CALIB-005·008)**: 계측기 수리 1건 — `ask.sh` choice 가 필수 `criteria` 누락으로 HTTP 422(공식 문서 확인); 동일 엔드포인트·키·모델로 criteria 포함 직접 POST 로 수리, 무효분 보존(judge_control.void-422.jsonl). 대조 5/5 재현 통과(modify 1·approve 4, conf 0.97~1.0). 본 배치 105/105 ok(상한 내 — 실제 110회 ≤ 220회, 벽시계 ~35분 ≤ 8h, 재시도 0).
- **M5 판정(REQ-CALIB-006·007)**: 판사 일치 96/105 = 91.43%(approve 95/modify 5/hold 4/other 1 — 상수 아님). **상수 기준선 병기: always-approve 99.05%·always-hold 0%**. wrong-automation 0(모집단 hold 0). 밴드 t=0.30~0.80 전 구간 자격 없음 — (b) 109.05% 불가능; t=0.80 의 100%(12/12)는 n<20 으로 근거 불인정. 결론(운영자 지정 문구 2026-09-26): **이 모집단으로는 판별 불가 — 판사 품질이 아니라 모집단 검정력 문제**. llm 팔 미측정 — 새 카드 t1261 이관.

## §E.3 Run-phase Audit-Ready Signal

- run_status: audit-ready
- run_head: (본 파일 커밋 시점 HEAD — 브랜치 WT-kickoff-decider-eval, 미푸시·미병합)
- ACs: [PLAN] 11건 plan-audit 통과 유지 · [RUN] 3건 — AC-CALIB-012 PASS(양성 대조 선행·페이로드별 스캔·적중 전송 0)·AC-CALIB-013 PASS(criteria_commit 존재+첫 호출 선행)·AC-CALIB-014 PASS(세 상한 declared_at 선행·실제 110회/벽시계 ~35분 안 집행)
- run-phase 판정 대상(acceptance.md §이행 노트): REQ-001 스냅숏 ✓·004 대조 ✓·005 판사 대조 5/5 ✓·006 기준선 산출 ✓·007 판정 적용(자격 밴드 0 — 운영자 지정 문구) ✓·008 상한 ✓·011 스크럽 ✓·012 자율 3조건 ✓
- Deviations: 판사 계측기 수리 1건(ask.sh criteria 누락 → 직접 POST, 무효분 보존); X2 27건이 no-response 폴백으로 판명 — 모집단 제외(운영자 결정 아님, 원문 보존); SPEC status 전환(draft→in-progress)을 run 시작 커밋에 실으못하고 sync 커밋에 태운다(오케스트레이터 직접 run 경로 — 경위 본 행)
- Gaps: llm 팔 미측정(t1261)·카드 단위 보조 보고 생략(병합 기준 미등록, NC-2)·전사본 보존 기간·PII 하한 패턴 한계 — 판정서 Gaps 절 참조
- Next: sync(manager-docs — 코드 변경 0: CHANGELOG·배포 문서 영향 없음 판정부터) → sync-audit → 리드 병합 창

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: audit-ready
- sync_complete_at: 2026-09-26T17:56:42+09:00
- sync_commit_sha: `2e871e222` (커밋 1 — 본 필드가 담기는 커밋 자신의 SHA 는 커밋 전에 알 수 없어 플레이스홀더로 착지한 뒤 직후 커밋에서 backfill한 것; D3 자기참조 면제)
- 코드 제로 판정: 본 SPEC 은 제품 코드를 변경하지 않았다. `git diff --name-only 38148d891..HEAD` = 5파일 전부 `.moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/` 하위(acceptance·plan·progress·research·spec.md)이고, 비-SPEC 파일 수(`| grep -vcE '^\.moai/specs/'`) = 0 — `internal/`·템플릿·README·docs-site 변경 없음 (sync 페이즈 직접 재측정, 본 트리 HEAD `7c26395e1`, 2026-09-26)
- CHANGELOG 결정: NO-ENTRY. `grep -c 'SPEC-AUTONOMY-KICKOFF-CALIB-001' CHANGELOG.md` = 0(기존 항목 부재를 직접 확인)이고, 본 카드는 내부 도그푸드 측정(사전 등록 설계·실행·판정)으로 사용자 대면 동작 변화 0건 — 항목을 새로 만들지 않는다 (B12 배출 전 0건 확인 원칙)
- AC 정합: acceptance.md AC 14건(AC-CALIB-001..014, sync 재측정) — CHANGELOG 항목이 없으므로 정산(reconcile) 대상 없음
- frontmatter: status draft → completed — terminal 전환을 본 sync 커밋이 담는다 (3페이즈 close 관례상 최종 전환은 sync 커밋 몫. draft→in-progress 가 run 커밋에 실리지 않은 경위는 §E.3 Deviations 행 — 결과 상태는 관례와 동일). `updated:` 는 당일이라 값 불변
- b12 self-test: ① diff 비-SPEC 파일 수 = 0 (직접 재측정) ② CHANGELOG 매치 수 = 0 (직접 재측정) ③ acceptance.md AC 유니크 수 = 14 (직접 재측정) ④ spec lint — 아래 행
- spec lint: `moai spec lint SPEC-AUTONOMY-KICKOFF-CALIB-001` → "✓ No findings — all SPEC documents are valid", exit 0 (frontmatter 전환 직후 본 트리에서 실행)
- 산출물: SPEC 아티팩트 5파일 + 런 증거 `.moai/reports/t1244/`(untracked — verdict.md 5-섹션·run-record.md·extract/·runs/·judge_control.jsonl·judge_main.jsonl·scrub_scan.py·judge_batch.py). 트리 변경 없음 — 커밋 대상 아님
- Gaps: 런 증거 파일의 내용 정합은 §E.2/§E.3 가 담당하며 sync 는 존재 확인만 수행 — 내용 재감사는 sync-audit 몫. 그 외 sync 관측면(diff·CHANGELOG·AC 수·lint·frontmatter)은 전부 관측됨
- Next: sync-audit → 리드 develop 병합 창 (병합 후 push 는 리드 일괄)
