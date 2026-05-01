---
id: SPEC-V3R2-WF-002
document: progress
version: "0.1.0"
status: audit-ready
created: 2026-04-30
updated: 2026-04-30
author: manager-spec (plan-phase enrichment)
related_spec: SPEC-V3R2-WF-002
phase: plan
language: ko
plan_phase_started: 2026-04-30
plan_phase_completed: 2026-04-30
run_phase_started: null
sync_phase_started: null
---

# SPEC-V3R2-WF-002 — Progress Tracker (진행 상황 추적)

> 본 문서는 마일스톤별 진행 상황과 다음 단계를 기록한다. plan/run/sync 사이의 핸드오프 정보, 컨텍스트 한계 시 `/clear` 후 paste-back용 resume 메시지를 포함한다.

---

## 1. 현재 단계 (Current Phase)

| 항목 | 값 |
|------|-----|
| 현재 단계 | **Plan 완료, Run 대기** |
| 활성 worktree | `/Users/goos/.moai/worktrees/moai-adk/feat/SPEC-V3R2-WF-002-thin-wrapper` |
| 활성 브랜치 | `feat/SPEC-V3R2-WF-002-thin-wrapper` |
| 베이스 브랜치 | `main` (HEAD: `3f0933550` Wave 3 머지 완료) |
| 의존 SPEC | SPEC-V3R2-WF-001 (이미 main 머지됨) |
| 다음 명령 | `/moai run SPEC-V3R2-WF-002` |

---

## 2. Plan 단계 산출물 (완료)

| 문서 | 절대 경로 | 상태 |
|------|-----------|------|
| spec.md | `/Users/goos/.moai/worktrees/moai-adk/feat/SPEC-V3R2-WF-002-thin-wrapper/.moai/specs/SPEC-V3R2-WF-002/spec.md` | 기존 보존 (수정 금지) |
| research.md | `.moai/specs/SPEC-V3R2-WF-002/research.md` | **작성 완료** |
| plan.md | `.moai/specs/SPEC-V3R2-WF-002/plan.md` | **작성 완료** |
| acceptance.md | `.moai/specs/SPEC-V3R2-WF-002/acceptance.md` | **작성 완료** |
| progress.md | `.moai/specs/SPEC-V3R2-WF-002/progress.md` | **본 문서 작성 완료** |

---

## 3. 마일스톤 체크리스트 (Run 단계)

전체 5개 마일스톤. 모두 미완료 상태로 시작. 각 마일스톤은 plan.md §3 상세 참조.

### M1 — Skill 골격 신설

- [ ] `.claude/skills/moai-workflow-github/` 디렉터리 생성
- [ ] `.claude/skills/moai-workflow-github/SKILL.md` 작성 (frontmatter only)
  - [ ] `name: moai-workflow-github`
  - [ ] `description` "(dev-only)" prefix 포함
  - [ ] `license: Apache-2.0`
  - [ ] `compatibility: Designed for Claude Code`
  - [ ] `allowed-tools` (CSV) — 원본 98-github.md `allowed-tools` 동등
  - [ ] `user-invocable: false` ← REQ-WF002-006
  - [ ] `metadata.{category, status, version, tags}`
  - [ ] `progressive_disclosure.enabled: true`
  - [ ] `triggers.keywords` (dev-only 키워드)
- [ ] `.claude/skills/moai-workflow-release/` 디렉터리 생성
- [ ] `.claude/skills/moai-workflow-release/SKILL.md` 작성 (frontmatter only)
  - [ ] 동일 frontmatter 구조 + release 도메인 키워드
- [ ] M1 검증: 두 SKILL.md frontmatter parse 성공 + `user-invocable: false` 확인

**충족 REQ**: REQ-WF002-006

---

### M2 — Fat Command Logic Extraction

- [ ] **추출 전 parity baseline 생성**:
  ```bash
  grep '^## Phase' .claude/commands/99-release.md > /tmp/before-phases.txt
  grep -E '^- \*\*(issues|pr|--|NUMBER)' .claude/commands/98-github.md > /tmp/before-flags.txt
  ```
- [ ] `98-github.md` body → `moai-workflow-github/SKILL.md` 본문 이관
  - [ ] Section: GitHub Workflow Configuration
  - [ ] Section: Argument Parsing (issues / pr 분기)
  - [ ] Section: Pre-execution Context (gh repo view, git, env, config)
  - [ ] Section: sub-command `issues` (Agent Teams, fix/issue-{N}, --merge)
  - [ ] Section: sub-command `pr` (review, label, merge)
  - [ ] Section: AskUserQuestion fallback
- [ ] `99-release.md` body → `moai-workflow-release/SKILL.md` 본문 이관
  - [ ] Section: Release Configuration (Enhanced GitHub Flow §18)
  - [ ] Section: Tool Selection Guide
  - [ ] Section: Execution Directive (VERSION, --hotfix)
  - [ ] Phase 1: 사전 검증
  - [ ] Phase 2: 버전 결정 (AskUserQuestion)
  - [ ] Phase 3: Release 브랜치 생성
  - [ ] Phase 4: Bilingual CHANGELOG 생성
  - [ ] Phase 5: PR 생성/머지 (merge commit)
  - [ ] Phase 6: Tag + GoReleaser
  - [ ] Phase 7: Post-release 안내
  - [ ] Quality escalation (expert-debug)
- [ ] **추출 후 parity 검증**:
  ```bash
  grep '^## Phase' .claude/skills/moai-workflow-release/SKILL.md > /tmp/after-phases.txt
  diff /tmp/before-phases.txt /tmp/after-phases.txt   # 빈 출력이어야 함
  ```
- [ ] PR description에 diff 결과 첨부 (manual review용)

**충족 REQ**: REQ-WF002-003, REQ-WF002-004, REQ-WF002-013

---

### M3 — Thin-Wrapper 변환

- [ ] `98-github.md` thin-wrapper 변환
  - [ ] frontmatter 보존 (`description`, `argument-hint`, `type: local`, `version` bump 2.0.0 → 3.0.0)
  - [ ] `allowed-tools: Skill` (CSV 단일 값)
  - [ ] body = `Use Skill("moai-workflow-github") with arguments: $ARGUMENTS`
  - [ ] LOC 검증:
    ```bash
    awk '/^---$/{c++; next} c==2' .claude/commands/98-github.md | grep -cv '^$'
    ```
    결과 ≤ 20
- [ ] `99-release.md` thin-wrapper 변환
  - [ ] frontmatter 보존 (`description`, `argument-hint`, `type: local`, `disable-model-invocation: true`, `metadata.*`, `version` bump 5.0.0 → 6.0.0)
  - [ ] `allowed-tools: Skill`
  - [ ] body = `Use Skill("moai-workflow-release") with arguments: $ARGUMENTS`
  - [ ] LOC 검증 ≤ 20
- [ ] M3 1차 자동 검증:
  ```bash
  go test -count=1 -run TestCommandsThinPattern ./internal/template/...
  ```

**충족 REQ**: REQ-WF002-001, REQ-WF002-002, REQ-WF002-007, REQ-WF002-008

---

### M4 — Audit 테스트 확장 + DEV_ONLY_SKILL_LEAK

- [ ] `internal/template/commands_audit_test.go` 수정
  - [ ] root-level command 명시적 검증 sub-test 추가
  - [ ] `rootCommands` 화이트리스트 검토 (98-github.md, 99-release.md가 묵시적으로 포함되지 않았는지 확인)
  - [ ] 정책 주석 추가 (REQ-WF002-005 / 009 참조)
  - [ ] partial migration 게이트 (REQ-WF002-015): `Skill("<name>")` 호출 시 `<name>` skill 디렉터리 존재 검증
- [ ] `internal/template/dev_only_skill_test.go` 신설
  - [ ] `TestDevOnlySkillLeak` 함수 작성
  - [ ] 검사 대상: `moai-workflow-github`, `moai-workflow-release`
  - [ ] 검사 위치: `internal/template/templates/.claude/skills/`
  - [ ] fail 메시지 형식: `DEV_ONLY_SKILL_LEAK: skill %q found at %q. ...` (research.md §5.3)
- [ ] 부정 케이스 검증
  - [ ] 임시로 `internal/template/templates/.claude/skills/moai-workflow-github/SKILL.md` 생성
  - [ ] `go test -run TestDevOnlySkillLeak ./internal/template/...` fail 확인
  - [ ] 임시 파일 rollback
- [ ] M4 자동 검증:
  ```bash
  go test -count=1 ./internal/template/...
  ```

**충족 REQ**: REQ-WF002-005, REQ-WF002-009, REQ-WF002-011, REQ-WF002-014, REQ-WF002-015

---

### M5 — 빌드 / 통합 검증

- [ ] 전체 테스트 실행:
  ```bash
  go test -count=1 ./...
  go test -race -run TestCommandsThinPattern ./internal/template/...
  go test -run TestDevOnlySkillLeak ./internal/template/...
  ```
- [ ] 빌드 검증:
  ```bash
  make build
  make install
  moai version
  ```
- [ ] diff-parity 최종 확인 (M2 산출물 재실행)
- [ ] 의존성 SPEC 회귀 없음 확인 (`git status`, `git diff` 의도하지 않은 변경 없음)
- [ ] 12/12 AC 모두 PASS 확인 (acceptance.md Definition of Done 체크리스트)

**충족 REQ**: REQ-WF002-009, REQ-WF002-010

---

## 4. 다음 단계 (Next Action)

```
/moai run SPEC-V3R2-WF-002
```

**시작 조건 충족 여부**:

- [x] spec.md 변경 없음 (작성 제약 준수)
- [x] research.md 작성 완료
- [x] plan.md 작성 완료
- [x] acceptance.md 작성 완료
- [x] progress.md 작성 완료
- [x] 의존 SPEC-V3R2-WF-001 main 머지 확인
- [x] worktree 활성

→ **run 단계 진입 가능 상태**.

---

## 5. Resume Message (컨텍스트 클리어 후 paste-back용)

만약 run 단계 작업 중 컨텍스트 사용량이 75%를 초과하여 `/clear`가 필요하면, 다음 메시지를 그대로 paste:

```
ultrathink. SPEC-V3R2-WF-002 Run 단계 이어서 진행. 두 fat command 추출 작업.
worktree: /Users/goos/.moai/worktrees/moai-adk/feat/SPEC-V3R2-WF-002-thin-wrapper
progress.md: .moai/specs/SPEC-V3R2-WF-002/progress.md
직전까지 완료한 마일스톤: <Mn 기록>.
다음 단계: <plan.md §3 의 다음 마일스톤 작업 시작>.
완료 후: 12/12 AC PASS 확인 → /moai sync SPEC-V3R2-WF-002.
```

---

## 6. 변경 이력 (Change Log)

| 시점 | 작성자 | 내용 |
|------|--------|------|
| 2026-04-30 | manager-spec | Plan 단계 4종 문서 (research/plan/acceptance/progress) 신설 |

---

## 7. 차단 요인 (Blockers)

현재 차단 요인 없음. run 단계 진입 가능.

향후 등장 가능한 차단 요인 (사전 인지):

- M2 단계에서 99-release.md `metadata.*` 핵심 필드 누락 시 → release 자동화 회귀 → plan.md §6 R7 완화책 적용
- M4 단계에서 commands_audit_test.go 묵시적 화이트리스트 발견 시 → research.md §3.2 ANCHOR-RESEARCH-001 참조하여 대응
- 컨텍스트 75% 도달 시 → 본 progress.md update 후 `/clear` + §5 resume message 사용

End of progress.md.
