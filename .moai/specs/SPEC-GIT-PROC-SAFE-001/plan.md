# SPEC-GIT-PROC-SAFE-001 — 구현 계획

## §A Context

- 카드 t782, Tier M, harness standard. 문서 절차 교정 SPEC — 런타임 코드 변경 없음.
- 감사 baseline main `2213871af` → 발견 건 재검증 트리 develop `c9ceff175`(본 워크트리). 모든 file:line은 `c9ceff175` 기준 측정치이며 research.md에 귀속 기록.
- 개발 모드: 문서 변경이므로 TDD/DDD 사이클은 적용하지 않고, 문서 전용 RED/GREEN 2셀 검증(acceptance.md 참조)을 사용한다.

## §B Known Issues

| ID | 심각도 | 상태 | 요지 |
|----|--------|------|------|
| AC-01 | P1 | LIVE | primary 체크아웃 브랜치 변이 절차 교리 (manager-git.md + spec-workflow.md SSOT 궤적) |
| AC-11 | P2 | 수리됨(develop) | fetch/rev-list 배치 순서 — 검증 전용 폐쇄 |
| SX-R04 | P1 | 부분 수리 | delivery.md Step 3.4 중복 레시피 (SSOT = manager-git.md § PR Auto-Merge) |

## §C Pre-flight

- [x] 대상 6개 파일 존재 + 위반 텍스트 직접 판독 완료(본 워크트리, c9ceff175) — research.md 기록
- [x] SPEC ID 중복 없음(`.moai/specs/` 목록 확인), ID 정규식 PASS
- [ ] run-phase 진입 전: Implementation Kickoff Approval (plan→run 인간 게이트)
- [ ] 편집 대상 파일에 대한 병행 세션 probe (워크트리 내 작업이므로 격리 확보 상태)

## §D 제약

1. **Template-First, 양 사본**: 모든 편집 템플릿 먼저 → 로컬 둘째, 발견 건 범위 편집. 전체 파일 cp 금지, 범위 밖 분기(§B 이외 영역) 비접촉.
2. **대체 텍스트 자기규율**: REQ-GP-002 허용 형태 외의 primary-checkout 명령을 대체 텍스트가 담지 않는지 각 마일스톤 편집 후 grep 자체검증.
3. **템플릿 중립성**: 템플릿 사본에 내부 SPEC ID·dev-only 로컬 경로(`.claude/rules/local/*`) 미인용.
4. **언어**: 대상 파일(manager-git.md, spec-workflow.md, delivery.md)은 기존 문서 언어(영어) 규약을 따른다 — 파일 내 기존 스타일 일관성 우선. SPEC 아티팩트만 한국어.
5. `make build` 실행 금지 — 오케스트레이터가 git·빌드를 담당한다. agents-emit 필요성(run-phase 종료 시 오케스트레이터에게 보고): manager-git.md의 `.md` 편집은 `make agents-emit` 재생성 대상이 아닌가 확인 — emit 대상은 `.claude/agents/moai/*.md` → `.codex/agents/moai/*.toml` 이므로 **manager-git.md 편집 시 필요**. 보고 필수.

## §E 자체 검증

- [ ] REQ-GP-002 grep 자체검증: 편집된 manager-git.md 양 사본에서 primary-checkout 문맥의 금지 명령(`git checkout main`, `git switch -c`, `git reset --hard`) 잔존 0건 — 문맥 근거(워크트리 내부 명령인지)와 함께 판독
- [ ] REQ-SX-01: delivery.md Step 3.4 내 "Mode conditions (same as" 중복 재인용 잔존 0건, "Auto-Merge Execution" 5단계 레시피 잔존 0건, Post-Merge Automatic Cleanup 절 존재 유지
- [ ] AC-AC11-01: 관련 파일 diff에서 AC-11 대상 행(158행/156행) 불변 확인
- [ ] REQ-TF-001: 양 사본 diff가 발견 건 범위 행만 포함하는지 `git diff` 판독

## §F Milestones

우선순위: 변경 가능성이 높은 결정(절차 모델 교체)을 앞세우고 기계적 정리를 뒤로.

### M1 (Priority High) — manager-git.md Late-Branch 절차 워크트리 전환 (양 사본)

- 템플릿 `internal/template/templates/.claude/agents/moai/manager-git.md` 88-129행 영역: § Late-Branch Invocation Pattern을 워크트리 모델로 재작성. 새 모델: (1) 절차 진입 전 런처 경유 워크트리 진입(`moai cc -w <name>` / `EnterWorktree(<path>)`), (2) plan/run 커밋은 워크트리 자기 브랜치에 적립(push 없음), (3) 브랜치 승격은 구성된 workflow에 조건부 — PR 통합 모드면 PR 시점에 해당 브랜치 push + `gh pr create` → 해석된 merge_method로 병합, git-flow 모드면 통합 창(integration window) 규율에 따라 develop 통합 워크트리로 병합(템플릿 중립 표현 유지 — dev-only 로컬 규칙 경로 미인용), (4) Phase D 삭제 — main이 커밋을 받지 않으므로 main 정렬 수순 소멸, 실패 복구도 `reset --hard` 없이 재정의. 139행 `main_late_branch` 옵션 설명도 동기화. 131행 교차참조 텍스트는 REQ-GP-004에 따라 유지/갱신.
- 로컬 `.claude/agents/moai/manager-git.md` 동일 영역(90-131행 + 139행) 동일 적용.
- 소유: manager-develop. 검증: §E 자체검증 1번 + AC-GP-01a GREEN 셀.

### M2 (Priority High) — spec-workflow.md SSOT 궤적 수리 (양 사본)

- 템플릿/로컬 `.claude/rules/moai/workflow/spec-workflow.md`: 50행 Route B Late-branch 사전조건과 53-62행 Step 4 closure 블록·post-condition을 워크트리 모델로 재작성. closure 블록(`git checkout main`... `reset --hard`)을 워크트리 브랜치 모델의 검증 수순으로 대체. 62행 교차참조(→ manager-git.md § Late-Branch Invocation Pattern) 섹션명 정합 유지.
- M1 완료 후 착수(교차참조 정합 의존). 소유: manager-develop. 검증: AC-GP-01b GREEN 셀 + REQ-GP-004.

### M3 (Priority Medium) — delivery.md Step 3.4 SSOT 위임 축소 (양 사본)

- 템플릿/로컬 `.claude/skills/moai/workflows/sync/delivery.md`: Step 3.4에서 중복을 제거 — 모드 조건 재인용("Mode conditions (same as ...)" 뒤 2불릿), Auto-Merge Execution 5단계 레시피, Auto-Merge Failures 3불릿. 남기는 것: "Step 3.2에서 PR 생성 시에만 적용" 프레이밍, 트리거 단일 기준 문장(이미 있음), `merge_method` 설정 해석 1행(이미 있음), `manager-git.md` § PR Auto-Merge 포인터, Post-Merge Automatic Cleanup 절(delivery 고유, `workflow.worktree.auto_cleanup` 키).
- 소유: manager-develop. 검증: AC-SX-01 GREEN 셀.

### M4 (Priority Medium) — AC-11 검증 전용 폐쇄 + 전체 재측정

- 파일 편집 0건. 양 사본 158행/156행 텍스트를 acceptance.md AC-AC11-01 증거로 판독·기록.
- 6개 파일 전체 `git diff`로 범위 밖 행 불변 확인(AC-MIRROR-01). late-branch 문맥 2차 grep 스윕 — 적중 시 blocker 보고.
- `make agents-emit` 필요성을 오케스트레이터 완료 보고에 명기(§D 5번).
- 소유: manager-develop. 검증: §E 전 항목.

## §G Anti-Patterns

- `cp` 전체 파일 미러링 — 범위 밖 분기 소멸 위험 (REQ-TF-001 위반)
- 대체 텍스트에 새 primary-checkout 명령 심기 (REQ-GP-002 위반)
- spec-workflow.md를 고치지 않고 manager-git.md만 고치기 (모순 잔존 — REQ-GP-003 위반)
- AC-11에 "한 번 더 수리" 시도 (편집 금지 위반)
- 템플릿 사본에 본 SPEC ID나 dev-only 경로 인용 (REQ-TN-001 위반)

## §H Cross-References

- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline (50, 53-62행)
- `.claude/agents/moai/manager-git.md` § Late-Branch Invocation Pattern / § PR Auto-Merge / § Synchronization
- `.claude/skills/moai/workflows/sync/delivery.md` Step 3.4
- `AGENTS.md` §2 (primary 체크아웃 금지 명령 목록 — 대체 텍스트의 규율 근거)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` (frontmatter SSOT)

미해결 명확화 항목이 없다 — 설계 방향은 발신 카드에서 사전 결정됨(웨지 모델 전환·SSOT 귀속·AC-11 검증 전용). 클리어런스 게이트 대상 표식은 플랜 아티팩트 어디에도 존재하지 않는다(plan-auditor 1차 판정 F6 반영).
