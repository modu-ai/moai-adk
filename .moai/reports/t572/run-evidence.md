# t572 run-evidence — SPEC-OWNERSHIP-SILENCE-001 (§E 자가 검증 원장)

- 트리: `.claude/worktrees/t572`, branch `WT-ownership-lint-silent`
- 시작 커밋(plan-phase): `b642479ec`
- 이 파일은 AC-OWN-001..008의 4요소 기록 원장이다. 각 항목: 커맨드 + verbatim 출력 + 종료 코드(별도 필드) + 트리 SHA.

---

## AC-OWN-001 — RED-now (4요소, right-reason)

- **RED 사유**: 무음 nil 분기(`lint_ownership.go:414-416`)가 발견을 반환하지 않기 때문이다. 컴파일 오류도, 무관 테스트 충돌도 아니다 — 두 서브테스트 모두 `expected exactly 1 OwnershipTransitionUnmeasured finding, got 0: []` 로 실패 = 분기가 조용한 nil로 통과하는 것이 관측된 원인이다.
- **(a) 커맨드** (단일 호출 형태):

```
go test ./internal/spec/ -run TestOwnershipTransitionUnmeasured -count=1
```

- **(b) verbatim stdout**:

```
--- FAIL: TestOwnershipTransitionUnmeasured (0.00s)
    --- FAIL: TestOwnershipTransitionUnmeasured/trailer_absent_emits_unmeasured (0.00s)
        lint_ownership_test.go:647: expected exactly 1 OwnershipTransitionUnmeasured finding, got 0: []
    --- FAIL: TestOwnershipTransitionUnmeasured/none_to_draft_emits_unmeasured_with_none_literal (0.00s)
        lint_ownership_test.go:701: expected exactly 1 OwnershipTransitionUnmeasured finding, got 0: []
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/spec	0.419s
FAIL
```

- **(c) 종료 코드**: `1`
- **(d) 트리 SHA**: `b642479ec` (plan-phase 커밋 — 테스트 파일 추가분은 미커밋 워킹 트리, 피시험 코드는 정확히 이 커밋의 pre-fix 상태)

---

## AC-OWN-002..005, 007 — (M2에서 아래에 추가)

## AC-OWN-006 — (M3에서 아래에 추가)

## AC-OWN-008 — 증거 경로 색인

- 본 파일: `.moai/reports/t572/run-evidence.md` (AC-OWN-001/004-RED/005/007 원장)
- `.moai/reports/t572/run-preflight.md` (Section C 재측정 — C.1 좌표/C.3 쌍둥이/C.4 baseline GREEN)
- `.moai/reports/t572/measurement-baseline.md` (plan-phase 실측 — 본 카드가 인용하는 1차 근거)
- 최종 AC 매핑 표: progress.md §E.2
