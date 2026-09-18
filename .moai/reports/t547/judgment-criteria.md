# t547 판정 기준 — AC-JFM-018 pull 창 (확정·커밋본)

2026-09-13 확정. `acceptance.md` AC-JFM-018 세 절반 + `design.md` §6.3 의 분모 규칙을 이 카드의
실행 가능한 기준으로 고정한다. 근거 원문은 SPEC 산출물이 소유하고 이 문서는 복제하지 않는다 —
충돌 시 SPEC 이 이긴다.

## 1. 분모 정의

- 분모 = **기록된 모든 행 중 `mode == "pull"`**. `question_type` 필터는 그 어떤 갈래에도 붙이지
  않는다(0.1.0 형식의 영구 공백 결함 — acceptance.md AC-JFM-018 본문).
- 창 시작점 앵커: **2026-09-13T23:07:40+0900** — primary `interview.yaml` 에
  `recommendation_mode: pull` 이 적용된 시각(파일 mtime, §4 참조).

## 2. 기존 push 76행 제외 — 구조적 근거

관측기(`internal/hook/askuser_observer.go` `resolveRecommendationMode`)는 질문 시점의 질문 세션
자기 트리 구성으로 각 행의 `mode` 를 찍는다. 따라서 적용 시각 이전에 기록된 행은 전부
`mode: "push"` 이고, `mode=="pull"` 필터가 그것들을 구조적으로 제외한다 — 시각 컷을 수동으로
대입해 행을 고를 필요가 없고, 그래서 경계 사고가 없다. 이 제외는 필터의 부산물이지 별도 규칙이
아니다. (현재 primary 로그 76행 전부 push — 2026-09-13 23:10 실측, 판정서 §6.)

## 3. 최소 분모 N

- **N = 20** (`acceptance.md` AC-JFM-018 본문 "at least 20"). `n < 20` 은 **gap** 이지 pass 가
  아니다. 20 은 준수 주장의 하한이지 통계 검정력 목표가 아니다.
- `n = 0` 은 "측정 불가" 상태다 — `violations == 0` 으로 읽는 것을 금지한다(공백 결함 방지,
  본 카드의 존재 이유).

## 4. 전제 충족 기록 (2026-09-13)

- 운영자 결정: 리드가 primary `.moai/config/sections/interview.yaml` 에
  `recommendation_mode: pull` 을 **수동 추가**(미커밋 로컬 설정; `moai update` 후 재적용 대상).
- 레인 독립 실측(2026-09-13 23:09~23:11 KST): `grep -n recommendation_mode` → 6행
  `recommendation_mode: pull` 적중; `git status --porcelain` → 해당 파일 `M`(미커밋) 확인.
  적용 방식은 리드 보고를 읽은 것이 아니라 이번 실행에서 파일을 직접 읽어 확인했다.

## 5. 판독 절차 (수집·판정 시)

1. `collect-pull-window.sh --selftest` — 필터 양성 대조(고장난 필터의 0 은 아무것도 증명하지
   않는다). FAIL 이면 수집 도구 수리 전까지 판정 금지.
2. `collect-pull-window.sh` — 전 트리 스캔 → export + 측정. 스크립트가 찍어주는 READING 줄이
   1차 판독이다.
3. `rows_recorded >= 20 && violations == 0` 이면: provenance md 작성(원본 절대경로, 세션별
   `session_id`, 구간 최초/최종 timestamp, 행 수, `label_present:true` 수, export 명령) +
   세션별 `calls_issued` 를 물어 세션에서 받아 4-way 대조:
   - `rows_recorded == calls_issued` → 표본 성립
   - `0 rows, calls_issued > 0` → 관측기 미배선 (2026-09-02 음성 관측과 동형)
   - `0 < rows < calls_issued` → 부분 손실 — **Gap**, 표본으로 읽지 않는다
   - `rows > calls_issued` → 타 세션 혼입 — `session_id` 분할 후에만 판독
4. provenance 의 행 수 == 실제 행 수 일치 확인(`wc -l`).

## 6. 선행 조건 플래그 — AC-JFM-023

AC-JFM-018 의 녹색 경로는 **AC-JFM-023(탐지기 양성 대조)이 먼저 녹색**이어야 연다
(`acceptance.md` "gated on AC-JFM-023 being green first"). 023 은 현재 RED —
`baseline-push-window.jsonl` export + provenance + `calls_issued` 4-way 대조가 미수행
(`progress.md` §E.4, sync-audit F1 정정). 기존 push 행(현재 76)이 023 기준선의 후보 재료지만
"관례 착지 전 기준선" 요건 적합 판정은 023 소관이고 이 카드가 대신 판정하지 않는다.
→ **pull 행이 20을 넘어도 023 이 녹색이 아니면 018 판정은 서지 않는다.** 023 처분은 리드·운영자
몫으로 본 판정서에 플래그만 남긴다.

## 7. 커밋·보고 규율

- 수집·판정 산출물은 `.moai/reports/t401/pull-window.jsonl` + `.provenance.md` 에 착지시키고
  판정 기록은 `.moai/reports/t547/verdict.md` 에 추가한다.
- push 금지(카드 지시). 판정이 착지하면 리드가 읽고 release 게이트(t204)가 참조한다.
