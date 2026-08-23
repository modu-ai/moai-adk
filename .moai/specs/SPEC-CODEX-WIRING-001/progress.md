# Progress — SPEC-CODEX-WIRING-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-08-23
- artifacts: spec.md / plan.md / acceptance.md / progress.md (Tier M — 3 + progress)
- lint: `moai spec lint .moai/specs/SPEC-CODEX-WIRING-001/spec.md` → exit 0 (2026-08-23 관측, v0.2.0 재확인)
- REQ 14 / AC 14 (MUST 13 + SHOULD 1 — Tier M 상한 16/16 이내)
- v0.3.0 (2026-08-24): 운영자 지시 statusline 범위 추가(카드 t88 본문 개정). 리드 배차
  메시지(main rebase 915c310de 후 착수)로 진행. §A.6 사실 기준 신설 — /statusline 문서·
  config reference·sample·소스 StatusLineItem enum(정식 29종+별칭 6종) 확정, 카드 제안 5종을
  정식 토큰(model-with-reasoning·context-remaining·git-branch·current-dir·thread-id)으로
  확정. REQ-CW-013·014 신설, AC 14건. 신규 토큰 base-0 실측: statusLineAllowlist 0·17827 0
  (status_line 단독은 46hit로 채택 제외 — 산출물 grep으로만 사용)
- grep-AC 토큰 base-0 실측: acceptance.md §A 서두에 기록
- plan-audit review-1 (2026-08-23): FAIL 0.86 — D1 차단(annotation 실태 오측 정정:
  명시 15/21·미선언 6·유효 불일치 4·실효 승인 집합 10) + D2-D6 경미 → v0.2.0으로 전량 적용.
  D1의 4도구 annotation 수정은 plan M2 + PRESERVE 협정 예외로 확정

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
