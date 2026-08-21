# Progress — SPEC-FEEDBACK-AUTO-SUBMIT-001

## §E.1 Plan-phase Audit-Ready Signal

- 2026-08-22 (1차) — plan-phase 산출물 3종 작성(Tier M). 근거는 `.moai/reports/t170/` 읽기 전용 렌즈 4종. 카드 전제 4건 반증 → §A 정정으로 기록, 운영자 결정 D1~D4 반영.
- 2026-08-22 (2차, 분리) — **AC 예산 초과로 SPEC을 둘로 분리**. 측정: `### AC-` 32개 > Tier M 상한 16, Tier L 상한 25. `spec-workflow.md` § SPEC Complexity Tier에 따라 완화가 아니라 분할로 대응.
  - 이 SPEC: **Tier L**(5종 — spec/plan/acceptance/design/research)로 재작성. 범위는 피드백 축(확인 게이트·스크러버·마스킹·취약점 분류·로그·큐·스킬 조항 + `feedback.auto_submit` 배선). **AC 23 / 상한 25**, REQ 13.
  - 신설 형제: `SPEC-TODO-ENABLE-FLAG-001`(**Tier M**, 3종). 범위는 todo 축. **AC 11 / 상한 16**, REQ 6.
  - 두 SPEC은 파일 9종을 공유하며, 병합 규율([HARD] 같은 파일의 다른 항목만 추가)은 양쪽 §E.1에 대칭으로 기록했다. `depends_on`은 기능 의존이 없어 미기재하고 그 근거를 §E.1에 남겼다.
  - 남은 결정: §D 결정 D5(웹 노출 경로 A/B) — 착수 승인 시점 확정.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
