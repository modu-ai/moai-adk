# plan.md — SPEC-AUTONOMY-KICKOFF-CALIB-001 (v0.1.0)

이 카드는 측정 설계 카드다. M1–M2 는 플랜 페이즈에서 이미 끝났고, M3–M5 는 run-phase 에서 실행한다(지연 — REQ-CALIB-009 전제). 구현 순서가 곧 측정의 논리 순서다: 도구 → 모집단 → 대조 → 판사 → 판정.

## §A. 마일스톤

### M1 — 참조 정독과 실측 프루브 (플랜 페이즈, 완료)

- A1 0.5.2(`25283ebf8`) §C.8·REQ-CONTRACT-019 — 결정자 값 집합, `jev_min_confidence` 정의와 기본 0.50.
- A3 0.3.3 §A 표(A5 배정)·REQ-GR-010(R1–R5)·REQ-GR-013·025 — 이 값의 소비처와 원칙 개정 연동.
- `CLAUDE.local.md` §29 전문(684–745행)·§30 전문(746–850행) — 재측정 대상과 재실험 금지.
- `.moai/reports/t943/verdict.md` 서두 — 5섹션 판정서 형식과 상수 기준선 구조.
- 추출 술어 실측 프루브 실행(본 카드 귀속, 2026-09-26): `.moai/reports/t1244/probe/kickoff_extract_probe.py` → `probe_output_20260926.txt`. 결과: 후보 148·응답 짝 148/148·세션 87·오류 0. 산출물: research.md §3.

### M2 — 프로토콜 사전 등록 (플랜 페이즈, 완료)

- spec.md §C REQ-CALIB-001~010 저작 — 술어·라벨·한국어 원문·대조 두 팔·기준선·판정 기준·상한·전제·t943 규율.
- acceptance.md AC-CALIB-001~010 — [PLAN]/[RUN] 두 층 분리.

### M3 — 모집단 추출과 스냅샷 (run-phase)

- 프루브 스크립트를 `.moai/reports/t1244/extract/` 로 확장해 1단계·2단계 술어 전부 적용(REQ-CALIB-001).
- 스윕: 양쪽 코퍼스 루트, 스냅샷 일자·스윕 명령을 run-record.md 에 기록.
- 산출: 라운드 레코드 JSONL(필드는 REQ-CALIB-001), `excluded_unanswered`·`excluded_ambiguous` 집계와 원문 보존.

### M4 — 대조 두 팔과 판사 배치 (run-phase)

- 파이프라인 대조: 20라운드 손검증 표본 → 20/20(REQ-CALIB-004). 실패 시 수리 후 재시작, 이전 판사 출력 무효.
- 판사 대조: 응답-포함 항목 5건(REQ-CALIB-005).
- **스크럽 게이트 선행**: 스캐너의 더미 비밀 양성 대조(심기 → 발화 확인)를 run-record.md 에 기록하고(REQ-CALIB-011), 페이로드마다 `payload_id`+`scan: clean|hit` 행을 남긴다 — 적중은 전송 차단(fail-closed).
- 세 상한(턴·호출 수·벽시계)을 `declared_at` 시각과 함께 run-record.md 에 적고(REQ-CALIB-008·012), `criteria_commit:` 해시를 기록한 **그 다음에** 본 측정 배치 실행 — 순서 어김 없음. 기준 변경이 필요해지면 먼저 커밋하고(해시 갱신) 판사 호출한다.

### M5 — 기준선·밴드 분석과 판정서 (run-phase)

- always-approve·always-hold 산출(REQ-CALIB-006, 질문·카드 단위 각각).
- 밴드 분석과 REQ-CALIB-007 기준 적용 — 채택(최저 자격 t) 또는 기각.
- `.moai/reports/t1244/verdict.md` 작성 — 5섹션 형식, 판사 수치는 상시 기준선 옆에.
- develop 병합은 레인 창 관례(git-flow 통합 체인)를 따른다.

## §B. 판사 후보 상세

- 1차: Jev — `ask.sh choice` 형태(`noul` 은 yes/no 라 4값 출력과 불일치 — 감사 D3). 지시문에 4 선택값을 열거하고, 신뢰도는 응답 JSON 의 선택 라벨 확률 필드에서 읽는다(부재 시 `confidence-미측정` → 밴드 표본 탈락, REQ-CALIB-008). 경로 `/Users/goos/MoAI/moai-adk-go/scripts/jev/ask.sh`(primary 체크아웃, git 미추적 — 2026-09-26 `git ls-files` 카운트 0 실측). 입력: 스크럽 투영(REQ-CALIB-007·011 — 질문 원문 + 과업 지시문, 스크럽 통과분), 출력: 원시 JSON을 `runs/` 에 보존.
- 과업 지시문은 한국어로 쓰고(REQ-CALIB-003), 판사 원시 응답을 `runs/`에 그대로 남긴다 — 가공 재구성은 분석 단계에서도 하지 않는다(가공본 재독은 독립 검증이 아니라는 교훈).
- 대체 판사: 같은 프로토콜이면 무엇이든 — REQ-CALIB-004~007 이 그대로 적용된다. 판사 교체는 판정 기준 변경이 아니다.

## §C. run-phase 전제 (지연 사유 포함)

1. **A1 병합 확인** — `git merge-base --is-ancestor e4ea8eb05 HEAD; echo $?` → 0. (2026-09-26 본 워크트리 관측: 이미 0. 리드 전달 기준 착지 헤드 `b1a62fb2b` 도 HEAD 도달 가능 확인 — exit 0. 실행 시점에 명령으로 다시 읽는다.)
2. **게이트 처분 — 자율 Kickoff(운영자 결정 2026-09-26, 리드 경유)**: 이 카드의 측정 실행은 plan-audit 뒤 별도 운영자 게이트 없이 진행하며, REQ-CALIB-012 의 세 묶음 조건이 구속한다 — (a) 판정 기준의 사전 등록 커밋(해시를 run 기록이 `criteria_commit:` 으로 기록 — 본 SPEC 커밋 `cc4055f9e`·`26d1004b5`·이 0.2.0 커밋이 그 대상), (b) 턴·호출 수·벽시계 세 중복 상한의 선언 시각 포함 기록, (c) 스크럽 게이트 통과. 종전 「운영자 레인 창 직접 승인」 전제는 이 카드에 한해 대체됐다 — 배차 범위 기록이지 독트린 일반화가 아니다.
3. A3(t1236) 병합은 전제가 **아니다** — A3 는 A5 를 기다리지 않고(A3 §A 표), 측정은 세션 전사본만 읽는다.
4. `scripts/jev/` 부재 또는 키 부재는 측정 불가 판정서로 닫는 경로다(REQ-CALIB-008 fail-open) — 전제 미충족으로 run 을 막는 것과도, 스크럽 게이트(REQ-CALIB-011 — 실패 시 전송 차단, fail-closed)와 다른 갈래다.

## §D. 리스크와 처분

| 리스크 | 처분 |
|---|---|
| 전사본 보존 기간 안에 모집단이 줄어든다 | 스냅샷 일자 기록이 의무(REQ-CALIB-001); 오래된 세션일수록 잘렸을 가능성을 판정서 Gaps 에 적는다 |
| 2단계 술어의 제외물(`excluded_rule`·`excluded_ambiguous`)이 커진다 | X1 제외물은 건수+세션 파일 목록을, X2 는 원문 통째를 run 기록에 남긴다 — 모집단 과소 계상은 어느 경로에서나 판정서에 드러난다(REQ-CALIB-001 X1·X2) |
| 판사 fail-open 전파 | 전 항목 미측정 → 측정 불가 판정서(REQ-CALIB-008) — 부분 점수로 이어가지 않는다 |
| 복합 응답의 카드 귀속 실패 | 질문 단위가 1차 기준(REQ-CALIB-002·NC-2) — 카드 단위는 보조 보고 |

## §E. 마일스톤-REQ 대응

| 마일스톤 | 묶는 REQ |
|---|---|
| M1·M2 (플랜) | REQ-CALIB-001(프루브)·002~010(본문 등록) |
| M3 | REQ-CALIB-001 |
| M4 | REQ-CALIB-003·004·005·008·011·012 |
| M5 | REQ-CALIB-002(보고)·006·007·009·010 |
