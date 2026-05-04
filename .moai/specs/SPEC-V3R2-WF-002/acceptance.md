---
id: SPEC-V3R2-WF-002
document: acceptance
version: "0.1.0"
status: audit-ready
created: 2026-04-30
updated: 2026-04-30
author: manager-spec (plan-phase enrichment)
related_spec: SPEC-V3R2-WF-002
phase: plan
language: ko
total_acceptance_criteria: 12
---

# SPEC-V3R2-WF-002 — Acceptance Criteria (수용 기준)

> spec.md §6의 12개 AC를 Given-When-Then 형식으로 풀어쓴다. 각 AC는 자동 또는 수동 검증 단계를 명시한다. 12/12 모두 GWT 변환 완료.

---

## AC-WF002-01 — 98-github.md body LOC ≤ 20

**REQ 매핑**: REQ-WF002-001 (Ubiquitous)

### Given

- M3 마일스톤이 완료되어 `.claude/commands/98-github.md`가 thin-wrapper로 변환됨.
- frontmatter는 `description`, `argument-hint`, `type`, `allowed-tools`, `version` 5개 필드 보존.

### When

- 다음 명령으로 frontmatter를 제외한 본문 non-empty LOC를 측정:
  ```bash
  awk '/^---$/{c++; next} c==2' .claude/commands/98-github.md | grep -cv '^$'
  ```

### Then

- 결과값이 **≤ 20**이어야 한다 (실제 목표: 1).
- `internal/template/commands_audit_test.go`의 `TestCommandsThinPattern`이 `98-github.md`에 대해 PASS.

### 검증 방식

- **자동**: `go test -run TestCommandsThinPattern ./internal/template/...`

---

## AC-WF002-02 — 99-release.md body LOC ≤ 20

**REQ 매핑**: REQ-WF002-002 (Ubiquitous)

### Given

- M3 마일스톤이 완료되어 `.claude/commands/99-release.md`가 thin-wrapper로 변환됨.
- frontmatter는 `description`, `argument-hint`, `type`, `allowed-tools`, `disable-model-invocation`, `version`, `metadata.*` 보존.

### When

- 다음 명령으로 본문 non-empty LOC를 측정:
  ```bash
  awk '/^---$/{c++; next} c==2' .claude/commands/99-release.md | grep -cv '^$'
  ```

### Then

- 결과값이 **≤ 20**이어야 한다 (실제 목표: 1).
- `TestCommandsThinPattern`이 `99-release.md`에 대해 PASS.

### 검증 방식

- **자동**: `go test -run TestCommandsThinPattern ./internal/template/...`

---

## AC-WF002-03 — moai-workflow-github skill에 GitHub orchestration 보존

**REQ 매핑**: REQ-WF002-003 (Ubiquitous)

### Given

- M2 마일스톤이 완료되어 `.claude/skills/moai-workflow-github/SKILL.md`가 신설됨.
- 추출 전 `98-github.md` 본문 (698 LOC)이 모두 SKILL.md 본문으로 이관됨.

### When

- 다음 키워드/sub-command가 SKILL.md에 존재하는지 검사:
  ```bash
  grep -E '(issues|^- \*\*pr|--all|--label|--solo|--merge|NUMBER|fix/issue-|gh repo view|Agent Teams)' .claude/skills/moai-workflow-github/SKILL.md | wc -l
  ```
- 동일 grep을 추출 전 `98-github.md` (git history)와 비교:
  ```bash
  git show HEAD:.claude/commands/98-github.md | grep -E '...' | wc -l
  ```

### Then

- 두 카운트가 **동일**해야 한다 (parity check).
- 추출 후 SKILL.md에 다음 절이 모두 등장:
  - "GitHub Workflow Configuration"
  - "Argument Parsing"
  - "Pre-execution Context"
  - sub-command `issues` 워크플로 전체
  - sub-command `pr` 워크플로 전체
  - AskUserQuestion fallback 흐름

### 검증 방식

- **자동**: grep 카운트 비교 (M2 종료 시 manual diff 실행).
- **수동**: PR 리뷰어가 SKILL.md를 읽고 원본 98-github.md와 의미론 동등 확인.

---

## AC-WF002-04 — moai-workflow-release skill에 Release orchestration 보존

**REQ 매핑**: REQ-WF002-004 (Ubiquitous)

### Given

- M2 마일스톤이 완료되어 `.claude/skills/moai-workflow-release/SKILL.md`가 신설됨.
- 추출 전 `99-release.md` 본문 (933 LOC)이 모두 SKILL.md 본문으로 이관됨.

### When

- Phase 순서 검사:
  ```bash
  grep '^## Phase' .claude/skills/moai-workflow-release/SKILL.md > after-phases.txt
  git show HEAD:.claude/commands/99-release.md | grep '^## Phase' > before-phases.txt
  diff before-phases.txt after-phases.txt
  ```
- 핵심 키워드 검사:
  ```bash
  grep -E '(release/v|hotfix/v|merge-commit|expert-debug|manager-git|GoReleaser|CHANGELOG)' .claude/skills/moai-workflow-release/SKILL.md | wc -l
  ```

### Then

- `diff` 결과가 **빈 출력**이어야 한다 (Phase 1–7 순서 동일).
- 키워드 카운트가 추출 전과 동일.
- Enhanced GitHub Flow 참조 (`CLAUDE.local.md §18`) 보존.

### 검증 방식

- **자동**: diff 명령으로 Phase 순서 검증.
- **수동**: PR 리뷰어가 SKILL.md Phase 1–7 본문 의미론 검토.

---

## AC-WF002-05 — go test가 root-level 2개 command를 thin-wrapper로 검증

**REQ 매핑**: REQ-WF002-005, REQ-WF002-009 (Ubiquitous + Event-Driven)

### Given

- M4 마일스톤이 완료되어 `commands_audit_test.go`가 root-level 검증을 명시적으로 포함하도록 수정됨.
- M3 마일스톤이 완료되어 두 command가 thin-wrapper로 변환됨.

### When

- 다음 명령 실행:
  ```bash
  go test -count=1 -v -run TestCommandsThinPattern ./internal/template/...
  ```

### Then

- 출력에 다음 두 sub-test가 등장하고 모두 PASS:
  - `TestCommandsThinPattern/.claude/commands/98-github.md`
  - `TestCommandsThinPattern/.claude/commands/99-release.md`
- 전체 테스트 PASS, exit code 0.
- `/moai/*` 15개 sub-command도 회귀 없음 (모두 PASS).

### 검증 방식

- **자동**: `go test ./internal/template/...` 출력의 sub-test 라인 검증.

---

## AC-WF002-06 — 두 skill frontmatter `user-invocable: false`

**REQ 매핑**: REQ-WF002-006 (Ubiquitous)

### Given

- M1 마일스톤이 완료되어 두 skill frontmatter가 작성됨.

### When

- 다음 명령으로 frontmatter 필드 확인:
  ```bash
  grep -E '^user-invocable:' .claude/skills/moai-workflow-github/SKILL.md
  grep -E '^user-invocable:' .claude/skills/moai-workflow-release/SKILL.md
  ```

### Then

- 두 출력 모두 `user-invocable: false`.
- 추가로 `description` 필드 값이 `(dev-only)` prefix 또는 동등한 식별자를 포함 (risk §8 trigger keyword 충돌 완화).

### 검증 방식

- **자동**: grep 결과 정확 매칭.
- **수동**: description 문구가 dev-only 의미를 명확히 전달하는지 리뷰.

---

## AC-WF002-07 — `/98-github` 호출 시 moai-workflow-github skill 위임

**REQ 매핑**: REQ-WF002-007 (Event-Driven)

### Given

- M3 마일스톤이 완료되어 `.claude/commands/98-github.md`가 thin-wrapper.
- maintainer가 Claude Code 환경에서 `/98-github issues --all` 등의 명령을 호출.

### When

- 명령 호출 시 wrapper body가 실행됨.

### Then

- wrapper body가 **단일 `Skill("moai-workflow-github")` 호출**을 포함하고 `$ARGUMENTS` 변수를 그대로 전달:
  ```
  Use Skill("moai-workflow-github") with arguments: $ARGUMENTS
  ```
- Static check (grep): 본문에 `Skill("moai-workflow-github")` 정확 일치 1회 등장, `$ARGUMENTS` 1회 등장.

### 검증 방식

- **자동 (정적)**:
  ```bash
  grep -c 'Skill("moai-workflow-github")' .claude/commands/98-github.md   # 1
  grep -c '\$ARGUMENTS' .claude/commands/98-github.md                      # 1
  ```
- **수동 (동적)**: dev 환경에서 실제 `/98-github` 호출 시 위임 정상 동작 확인 (run 단계).

---

## AC-WF002-08 — `/99-release` 호출 시 moai-workflow-release skill 위임

**REQ 매핑**: REQ-WF002-008 (Event-Driven)

### Given

- M3 마일스톤이 완료되어 `.claude/commands/99-release.md`가 thin-wrapper.
- maintainer가 Claude Code 환경에서 `/99-release v2.18.0` 또는 `/99-release v2.18.1 --hotfix` 호출.

### When

- 명령 호출 시 wrapper body가 실행됨.

### Then

- wrapper body가 **단일 `Skill("moai-workflow-release")` 호출**을 포함하고 `$ARGUMENTS` 전달:
  ```
  Use Skill("moai-workflow-release") with arguments: $ARGUMENTS
  ```

### 검증 방식

- **자동 (정적)**:
  ```bash
  grep -c 'Skill("moai-workflow-release")' .claude/commands/99-release.md   # 1
  grep -c '\$ARGUMENTS' .claude/commands/99-release.md                       # 1
  ```
- **수동 (동적)**: dev 환경에서 `/99-release` 호출 시 위임 정상 확인.

---

## AC-WF002-09 — DEV_ONLY_SKILL_LEAK CI fail

**REQ 매핑**: REQ-WF002-014 (Unwanted Behavior)

### Given

- M4 마일스톤이 완료되어 `internal/template/dev_only_skill_test.go`가 신설됨.
- 누군가 실수로 `internal/template/templates/.claude/skills/moai-workflow-github/SKILL.md` 또는 `moai-workflow-release/SKILL.md`를 추가했다고 가정.

### When

- 다음 명령 실행:
  ```bash
  go test -count=1 -run TestDevOnlySkillLeak ./internal/template/...
  ```

### Then

- 테스트가 **fail**하고, 오류 메시지에 `DEV_ONLY_SKILL_LEAK` 문자열이 포함:
  ```
  FAIL: DEV_ONLY_SKILL_LEAK: skill "moai-workflow-github" found at "internal/template/templates/.claude/skills/moai-workflow-github". This skill is dev-only (REQ-WF002-014, SPEC-V3R2-WF-002).
  ```
- exit code != 0.
- 정상 상태 (template tree에 두 skill 부재) 시 PASS.

### 검증 방식

- **자동**:
  - 정상 케이스: `go test ./internal/template/...` PASS.
  - 부정 케이스 (의도적 leak 생성 후 즉시 rollback): `go test` fail 확인.

---

## AC-WF002-10 — 99-release.md 21+ LOC 회귀 차단

**REQ 매핑**: REQ-WF002-011 (State-Driven)

### Given

- M3 / M4 마일스톤이 완료된 후 누군가 `99-release.md` body에 추가 로직을 넣어 LOC를 21줄 이상으로 늘리는 PR을 생성했다고 가정.

### When

- CI 실행 시 `go test -run TestCommandsThinPattern ./internal/template/...`이 자동 실행됨.

### Then

- 테스트가 **fail**하고, 오류 메시지가 다음 형식:
  ```
  body has 21 non-empty lines (max 19 for thin commands)
  ```
- exit code != 0, PR merge 차단.
- 동일 검증이 `98-github.md`에도 적용 (REQ-WF002-001 보호).

### 검증 방식

- **자동**: `commands_audit_test.go:R3` 분기. (이미 구현됨, M4에서 root-level 적용 확장).
- **부정 케이스 검증**: M5에서 임시로 `99-release.md` body에 빈 줄 21개 추가 → 테스트 fail 확인 → rollback.

---

## AC-WF002-11 — Release-step ordering 손실 시 PR review 차단

**REQ 매핑**: REQ-WF002-013 (Unwanted Behavior — Complex)

### Given

- M2 마일스톤 (logic 추출) 도중 `99-release.md`의 Phase 1–7 중 Phase 4와 Phase 5가 swap되어 SKILL.md에 잘못 이관되었다고 가정.

### When

- M2 종료 시 diff-parity 절차 (research.md §7.1) 실행:
  ```bash
  diff before-phases.txt after-phases.txt
  ```

### Then

- diff 결과가 **비어있지 않음** → ordering 위반 감지.
- PR review 체크리스트가 fail 표시.
- run 단계 작업자는 SKILL.md를 수정하여 Phase 순서 복원 후 재검증 필요.

### 검증 방식

- **수동 (필수)**: M2 작업 종료 시 diff 명령 실행 + PR description에 결과 첨부.
- **자동 (선택)**: 향후 CI에 grep + diff 자동화 가능 (본 SPEC 범위 외).

---

## AC-WF002-12 — 부분 마이그레이션 commit 차단

**REQ 매핑**: REQ-WF002-015 (Complex: State + Event)

### Given

- M3 도중 누군가 `98-github.md`의 body만 ≤20 LOC로 줄이고 `Skill("moai-workflow-github")` 호출을 추가했으나, 대응되는 `.claude/skills/moai-workflow-github/SKILL.md`를 아직 생성하지 않은 상태에서 commit을 시도.

### When

- pre-commit / CI 실행 시 `go test -run TestCommandsThinPattern ./internal/template/...` 자동 실행.

### Then

- M4에서 추가한 검증 로직이 동작:
  - wrapper body가 `Skill("<name>")` 패턴 포함 시
  - `<name>` skill 디렉터리가 `.claude/skills/`에 존재해야 함
- 두 조건 중 하나라도 위반 시 테스트 **fail**:
  ```
  THIN_WRAPPER_PARTIAL_MIGRATION: 98-github.md references Skill("moai-workflow-github")
  but .claude/skills/moai-workflow-github/SKILL.md does not exist (REQ-WF002-015).
  ```
- commit 차단.

### 검증 방식

- **자동**: M4의 추가 검증 로직.
- **부정 케이스 검증**: M5 도중 임시로 SKILL.md를 한 측에서 삭제 → 테스트 fail 확인 → rollback.

---

## AC 커버리지 요약

| AC ID | 매핑 REQ | 검증 방식 | 마일스톤 |
|-------|----------|------------|----------|
| AC-WF002-01 | REQ-001 | 자동 | M3 |
| AC-WF002-02 | REQ-002 | 자동 | M3 |
| AC-WF002-03 | REQ-003 | 자동(parity) + 수동 | M2 |
| AC-WF002-04 | REQ-004 | 자동(diff) + 수동 | M2 |
| AC-WF002-05 | REQ-005, REQ-009 | 자동 | M4, M5 |
| AC-WF002-06 | REQ-006 | 자동 + 수동 | M1 |
| AC-WF002-07 | REQ-007 | 자동(정적) + 수동(동적) | M3 |
| AC-WF002-08 | REQ-008 | 자동(정적) + 수동(동적) | M3 |
| AC-WF002-09 | REQ-014 | 자동 + 부정케이스 | M4 |
| AC-WF002-10 | REQ-011 | 자동 + 부정케이스 | M4 |
| AC-WF002-11 | REQ-013 | 수동(필수) | M2 |
| AC-WF002-12 | REQ-015 | 자동 + 부정케이스 | M4 |

→ **12/12 AC 모두 GWT 변환 완료**, 15 REQ 중 14개를 직접 매핑 (REQ-WF002-010 / 012는 spec.md §6에서 별도 AC 미할당; REQ-010은 AC-09에 간접 포함, REQ-012는 Optional 미래용 설계).

### REQ-WF002-010 / 012 보충 매핑

- **REQ-WF002-010** (dev-only exception, State-Driven): AC-WF002-09에 간접 포함 — `internal/template/templates/.claude/skills/`에 두 skill이 부재하면 `make build` template drift check 미적용.
- **REQ-WF002-012** (future user-invocable flip, Optional): 별도 AC 불필요 — frontmatter `user-invocable: false`를 `true`로 변경하는 것만으로 활성화 (M1에서 설계 보장, AC-WF002-06으로 검증 가능).

---

## Definition of Done (전체 SPEC 완료 조건)

- [ ] 12/12 AC 모두 PASS (위 표 검증 결과)
- [ ] `go test -count=1 ./...` PASS (회귀 없음)
- [ ] `make build` 성공 (zero warnings)
- [ ] `make install` 성공
- [ ] `moai version` 출력 정상
- [ ] M5 diff-parity 최종 검증 (Phase + flag 카운트 동일)
- [ ] BC-V3R2-012 sync 단계로 인계 (CHANGELOG entry 작성은 sync 책임)

End of acceptance.md.
