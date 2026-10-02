# decision-index — SPEC-WEB-SETTINGS-SAVE-001

> `interview.decision_gate: on`(.moai/config/sections/interview.yaml)에 따라 작성됐다. 각 행은 Detect → Explain → Ask를 담고, 포함된 권고를 갖지 않는다. `Operator verdict:`는 작성 시점에 비어 있다.

### Q1: 스코프 ③의 원격병합 확인 술어는 무엇인가 — 종료 경로의 신선한 fetch인가, 원격추적 참조 도달 검사인가?

Label: DECIDED

Authority anchor: 카드 t1393 리드 카드자율 처분(plan-audit iter 1 수리 재위임 지시 R5/D5, 2026-10-01 — 이 행이 그 기록 원장). 참조 정책 근거: AGENTS.md §3(폐기 규율).

Why unresolved: 해당 없음 — 리드가 구현-설계 결정으로 채택했다: **fetch 없는 원격추적 도달성**. 술어: (i) `git merge-base --is-ancestor <branch-tip> refs/remotes/origin/develop` 또는 (ii) `git cherry refs/remotes/origin/develop <branch>` 공집합(patch-id 등가 — 스쿼시 병합 커버, SPEC-WORKTREE-SQUASH-MERGE-001 교훈). 채택 근거: 착지는 단조적 — 스테일 원격추적 참조가 만들 수 있는 오판은 "아직 못 착지"→보존(안전 방향)뿐이다. 반면 종료 경로마다의 fresh fetch는 네트워크 비용과 새 실패 모드를 모든 종료에 얹어 REQ-WSS-304(저렴-공통-경로)와 충돌한다. 참조 신선도·이력 재작성·스쿼시 축의 잔여 위험은 REQ-WSS-302 ③(확인 결과의 이상 시 보존)과 fail-open이 흡수한다. <!-- moving-ref-ok: 술어가 참조하는 origin/develop은 종료 시점 판정의 대상(SUBJECT)이지 측정 앵커가 아니다 — 스테일 오판은 보존 방향으로만 실패한다 -->

Operator verdict: 리드 처분으로 채택(2026-10-01) — REQ-WSS-302에 술어 명시 반영 완료.

### Q2: 운영자 스크린샷의 "미저장" 배지는 어느 상태였는가 — 실제 dirty 상태인가, 오류 배지인가?

Label: EVIDENCE-NEEDED

Authority anchor: —

Why unresolved: 본 트리(develop f130aa041)에서 `data-save-state="dirty"`의 런타임 생산자가 관측되지 않는다(app.js 폼 변경 리스너 0건; `settingsSaveState`는 error|saved|clean만 반환 — internal/web/settings_shell.go:61). 운영자 관측이 실제 dirty였다면 실행 바이너리가 본 트리와 다르거나 미관측 생산자가 존재하고, 오류 배지였다면 가설 (a)(검증 거절)이 올라간다. plan-audit iter 1의 codex overlay 보고("최소 제출은 정상 저장" — 중간 신뢰도, 미검증 Gap)가 "운영자 바이너리 ≠ 본 트리" 대립 가설을 강화했다 — AC-WSS-001의 미재현 분기(acceptance.md §D.1)가 이 질의를 M1 진입 조건으로 흡수한다. 배지 정체가 가설 판별의 분기점이므로 M1 진단 전에 운영자 바이너리 버전 또는 스크린샷 재판독이 필요하다.

Operator verdict:

### Q3: 감사 핀 기본값은 어느 표면에 반영하는가?

Label: DECIDED

Authority anchor: SPEC-MODEL-MATRIX-UPDATE-001(HISTORY — M1/M2 codex 핀을 Go 기본값·배포 템플릿·resolver 터미널 폴백 3표면에 반영) + `internal/config/audit_models.go` AuditConfig.Codex doc comment(REQ-MMU-001 SUPERSEDES 기록).

Why unresolved: 해당 없음 — 동일 조건의 동일 질문을 선행 완료 SPEC의 이력이 이미 결정했다. claude·glm 핀도 같은 3중 정합 규율을 따른다(REQ-WSS-201).

Operator verdict:

### Q4: 푸시는 됐지만 원격 병합이 착지 전인 워크트리의 처분을 가드가 거부하는 것은 의도된 동작인가?

Label: POLICY-COVERED

Authority anchor: AGENTS.md §3("Dispose of no worktree — L1 or L2 — until the branch is integrated and the remote merge has landed") + 카드 t1393 발행문(향후 auto_cleanup 전환의 선행 조건 명시).

Why unresolved: 해당 없음 — 커밋된 정책 문서가 동일 질문을 문면 그대로 덮는다. 자동머지 성공 직후(로컬 병합·원격 미착지) 거부도 이 정책의 직접 귀결이다(REQ-WSS-301/306).

Operator verdict:
