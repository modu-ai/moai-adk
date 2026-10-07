# SPEC-GITHUB-FLOW-CI-RESIDUE-001 — Decision Index

`interview.decision_gate: on` (.moai/config/sections/interview.yaml:6)에 따라 작성됐다. 상태 축은 stateless — SPEC 생명주기는 spec.md frontmatter만이 운반한다.

### Q1: `moai worktree sweep`·`moai worktree done`의 착지 판정 기준 브랜치는 무엇이 답하는가

- Label: POLICY-COVERED
- Authority anchor: `.moai/config/sections/git-strategy.yaml` — `git_strategy.worktree_base_branch: main` (HEAD 커밋본 5행; `git show HEAD:.moai/config/sections/git-strategy.yaml`으로 검증 — 본 회차 실행)
- Why unresolved: 카드가 "기준 브랜치 main 전환 필수"와 "설정 키 우선"을 함께 지시해 해석 여지가 있었으나, 운영자 설정이 이미 main을 이름붙이고 있어 문면 그대로 덮인다 — M1은 그 설정을 찾지 못했던 sweep·done 해석을 사슬에 연결하는 수리일 뿐이다.
- Operator verdict:

### Q2: 새 다중 OS 집계 게이트를 브랜치 보호의 필수 체크로 등록하는 행위자는 누구인가

- Label: FOUNDER
- Class: implementation-level
- Authority anchor: 해당 없음 — 브랜치 보호 등록은 GitHub 저장소 설정의 운영자 콘솔 활이며 본 레지스트리의 커밋된 산출물이 정의하지 않는다 (cutover §I D-22의 선례가 있으나 그 SPEC은 아직 completed가 아니다 — 인용 불가).
- Why unresolved: REQ-GFC-010이 "등록 가능한 이름 게시"까지만 SPEC의 경계로 삼았고, 실제 필수 체크 등록은 저장소 설정 변경이라 코드로 증명할 수 없다.
- Default: 집계 게이트를 워크플로로 착지하고 필수 체크 등록은 운영자 활으로 운영 기록에 남긴다 (rule: 단일 revert — 워크플로 파일 삭제로 완전 되돌림)
- Alternate: advisory(비필수) 게이트로 착지 — 등록 없이 관측만
- Operator verdict: DEFAULT-APPLIED 2026-10-06T18:55:40Z manager-spec (card t1535)

### Q3: SPEC 상태 자동 동기화의 전달 정책은 무엇인가 (스코프 4 — v0.5.0 분할, hold)

- Label: FOUNDER
- Class: product-level (전달 정책 선택이 사용자 가시 워크플로 동작과 저장소 보호 상호작용을 바꾼다)
- Authority anchor: 해당 없음 — chore 성격 자동 동기화 PR의 병합 정책을 정하는 커밋된 규칙이 없다.
- Why unresolved: **상태 = hold — 본 SPEC에서 스코프 4를 분할한 운영 정책 선택 대기.** iter1 D2(무효 플래그)→iter2 D10(인증 배선)→iter3 D16(CI 미발화·auto-merge 정체)까지 세 계층의 수리가 매번 다음 계층을 드러냈고, iter3 판정(§Final Escalation)은 "무인 자동 병합 계약 자체가 이 저장소의 보호 설정과 양립하지 않는 구조 문제"로 진단했다 — 잔여 처분은 문안 수리가 아니라 아래 3택 중 하나의 운영 정책 선택이다. 스코프 4는 v0.5.0에서 본 SPEC에서 제외됐고(spec.md §C Out of Scope), 결정 후 별도 카드로 재계획한다.
- Operator verdict: (a) 라벨 부착+수동 병합 — 운영자 결정 2026-10-07, 후속 카드 t1557로 이관. 상태는 hold 유지(본 SPEC 밖 항목 — 결정 기록만 남긴다).
  - (a) 라벨+수동 병합 — auto-merge를 포기하고 동기화 PR에 라벨을 붙여 운영자가 수동 병합. 보호 설정과 무충돌, 무인 완료 포기.
  - (b) GitHub App 토큰 — App 발급 토큰으로 봇 PR의 병합 권한을 확보. 무인 완료 유지, App 운영 부담.
  - (c) auto-merge 유지+운영자 승인 명시 — auto-merge를 유지하되 운영자 승인 단계를 계약에 명시. 무인 완료는 절충.
