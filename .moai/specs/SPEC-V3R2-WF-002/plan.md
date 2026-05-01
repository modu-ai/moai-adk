---
id: SPEC-V3R2-WF-002
document: plan
version: "0.1.0"
status: audit-ready
created: 2026-04-30
updated: 2026-04-30
author: manager-spec (plan-phase enrichment)
related_spec: SPEC-V3R2-WF-002
phase: plan
language: ko
---

# SPEC-V3R2-WF-002 — Implementation Plan (구현 계획)

> 본 문서는 spec.md의 15개 REQ × 12개 AC를 어떤 마일스톤 순서로 구현할지 정의한다. 새로운 요구사항은 추가하지 않으며, 기존 명세를 어떻게 코드와 테스트로 옮길지에만 집중한다.

---

## 1. 전체 전략 (Strategy)

본 SPEC은 **behavior-preserving extraction** 작업이다. 두 fat command(`98-github.md` 698 LOC, `99-release.md` 933 LOC)의 orchestration 로직을 새로 만들 두 dev-only skill로 옮기고, 원본 command는 ≤20 LOC thin-wrapper로 줄인다. 핵심 원칙:

1. **추출 우선, 압축 나중**: skill 골격을 먼저 만들고 원본 본문 전체를 그대로 옮긴 후, command를 thin-wrapper로 전환. 본문 의미론은 변경하지 않는다.
2. **테스트 게이트로 회귀 차단**: REQ-WF002-005 / 011 / 014 / 015를 `commands_audit_test.go` 및 신규 `dev_only_skill_test.go`에 넣어 미래 회귀를 자동 차단한다.
3. **의존 그래프 직선화**: M1 (skill 골격) → M2 (logic 추출) → M3 (thin-wrapper 변환) → M4 (CI 게이트 강화) → M5 (검증). 분기 없음.
4. **dev-only 격리 유지**: 두 skill은 `.claude/skills/`에만 존재하고 `internal/template/templates/.claude/skills/`로 누출되지 않도록 CI 검사를 둔다 (REQ-WF002-014).

---

## 2. 영향 범위 (Files Touched)

| 종류 | 파일 | 변경 형태 |
|------|------|-----------|
| 신설 | `.claude/skills/moai-workflow-github/SKILL.md` | 새 파일 (frontmatter + 추출 로직) |
| 신설 | `.claude/skills/moai-workflow-release/SKILL.md` | 새 파일 (frontmatter + 추출 로직) |
| 수정 | `.claude/commands/98-github.md` | 698 → ≤20 LOC thin-wrapper |
| 수정 | `.claude/commands/99-release.md` | 933 → ≤20 LOC thin-wrapper |
| 수정 | `internal/template/commands_audit_test.go` | root-level 검증 추가, rootCommands 화이트리스트 검토 |
| 신설 | `internal/template/dev_only_skill_test.go` | `DEV_ONLY_SKILL_LEAK` 검사 (REQ-WF002-014) |
| 변경 없음 | `internal/template/templates/.claude/skills/` | dev-only이므로 등록 금지 |
| 변경 없음 | `.moai/config/sections/quality.yaml` | 수정 금지 (사용자 제약) |

---

## 3. 마일스톤 (Milestones)

총 **5개 마일스톤** (M1–M5). 각 마일스톤은 단일 책임을 가지며 다음 마일스톤의 사전 조건이다.

### M1 — Skill 골격 신설

**목표**: 두 신규 skill 디렉터리 생성 + frontmatter만 포함된 빈 SKILL.md 작성.

**산출물**:

- `.claude/skills/moai-workflow-github/SKILL.md` (frontmatter only)
- `.claude/skills/moai-workflow-release/SKILL.md` (frontmatter only)

**작업 내역**:

1. `mkdir -p .claude/skills/moai-workflow-github` / `moai-workflow-release`
2. 각 SKILL.md frontmatter 작성 — research.md §4 패턴 준수:
   - `name`, `description` ("(dev-only)" prefix 필수, risk 완화책)
   - `license: Apache-2.0`
   - `compatibility: Designed for Claude Code`
   - `allowed-tools` (CSV) — 원본 command의 `allowed-tools` 동등
   - `user-invocable: false` ← **REQ-WF002-006**
   - `metadata.category: workflow`, `metadata.status: active`, `metadata.tags`
   - `progressive_disclosure.enabled: true`
   - `triggers.keywords` — dev-only 키워드 (e.g., `["github-dev-workflow"]`) — 일반 사용자 trigger 회피
3. 본문은 placeholder 한 줄 (`<!-- M2에서 logic 추출 예정 -->`)로 둔다.

**검증**:

- 두 SKILL.md 모두 YAML frontmatter parse 성공
- `user-invocable: false` 명시 확인 → AC-WF002-06 1차 통과 가능 상태

**의존성**: SPEC-V3R2-WF-001 머지 완료 (이미 충족).

**충족 REQ**: REQ-WF002-006 (frontmatter `user-invocable: false`).

---

### M2 — Fat Command Logic Extraction

**목표**: `98-github.md` / `99-release.md`의 본문 전체를 각 SKILL.md로 의미론 보존하며 이관.

**산출물**:

- `moai-workflow-github/SKILL.md` 본문에 98-github.md §"GitHub Workflow Configuration" 이하 전체 섹션 이관.
- `moai-workflow-release/SKILL.md` 본문에 99-release.md §"Release Configuration (Enhanced GitHub Flow)" 이하 전체 섹션 이관.

**작업 내역**:

1. **98-github.md → moai-workflow-github/SKILL.md**:
   - Section 1: GitHub Workflow Configuration (repo 자동탐지, 기본 모드, 브랜치 prefix)
   - Section 2: Execution Directive — `$ARGUMENTS` 파싱 (issues / pr 분기)
   - Section 3: Pre-execution Context (gh repo view, git branch/status, env check, config import)
   - Section 4: Sub-command `issues` (Agent Teams 분기, fix/issue-{N}, --merge)
   - Section 5: Sub-command `pr` (Code review, label, merge)
   - Section 6: AskUserQuestion fallback (sub-command 미지정 시)
2. **99-release.md → moai-workflow-release/SKILL.md**:
   - Section 1: Release Configuration (Enhanced GitHub Flow §18 참조)
   - Section 2: Tool Selection Guide
   - Section 3: Execution Directive (VERSION, --hotfix 파싱)
   - Section 4: Phase 1–7 (사전 검증 → 버전 결정 → release 브랜치 → CHANGELOG → PR → tag/GoReleaser → post-release)
   - Section 5: Quality escalation (expert-debug 위임)
3. 추출 시 **diff-parity 절차** (research.md §7) 적용:
   - 추출 전: `grep '^## Phase' 99-release.md > before-phases.txt`
   - 추출 후: `grep '^## Phase' .claude/skills/moai-workflow-release/SKILL.md > after-phases.txt`
   - `diff before-phases.txt after-phases.txt` empty 확인 → AC-WF002-11 충족
4. 본문 길이 제한 없음 (skill body는 thin-wrapper 규제 대상 아님).

**검증**:

- 98-github.md 원본의 모든 sub-command (issues / pr) + 모든 플래그 (--all, --label, --solo, --merge, NUMBER)가 SKILL.md에 등장 — `grep -E` 기반 manual diff
- 99-release.md 원본의 Phase 1–7 순서 보존 — diff-parity check 통과
- AskUserQuestion 호출 횟수 일치
- `manager-git`, `expert-debug` 등 위임 keyword 출현 횟수 일치

**의존성**: M1 완료.

**충족 REQ**: REQ-WF002-003 (GitHub orchestration 보존), REQ-WF002-004 (Release orchestration 보존), REQ-WF002-013 (ordering 손실 방지).

---

### M3 — Thin-Wrapper 변환

**목표**: `98-github.md` / `99-release.md`를 ≤20 LOC body + 기존 frontmatter 구조로 압축.

**산출물**:

- `.claude/commands/98-github.md` (≤20 LOC body, `Skill("moai-workflow-github")` 호출 포함)
- `.claude/commands/99-release.md` (≤20 LOC body, `Skill("moai-workflow-release")` 호출 포함)

**작업 내역**:

1. **98-github.md 변환** (research.md §2.1 표준 패턴 준수):

   ```
   ---
   description: "GitHub Workflow - Manage issues and review PRs with Agent Teams"
   argument-hint: "issues [--all | --label LABEL | NUMBER | --merge | --solo | --tmux] | pr [--all | NUMBER | --merge | --solo]"
   type: local
   allowed-tools: Skill
   version: 3.0.0
   ---

   Use Skill("moai-workflow-github") with arguments: $ARGUMENTS
   ```

   - 기존 frontmatter의 `description`, `argument-hint`, `type: local` 보존.
   - `allowed-tools`는 thin-wrapper에 맞게 `Skill` 단일 값으로 축소 (실제 도구는 skill에서 호출).
   - `version: 2.0.0` → `3.0.0` bump (extraction 반영).
   - Body = 1 LOC (Skill 호출).

2. **99-release.md 변환**:

   ```
   ---
   description: "MoAI-ADK production release via Enhanced GitHub Flow (CLAUDE.local.md §18). All git operations delegated to manager-git. Quality failures escalate to expert-debug."
   argument-hint: "[VERSION] [--hotfix]"
   type: local
   allowed-tools: Skill
   disable-model-invocation: true
   version: 6.0.0
   metadata:
     release_target: "production"
     branch: "main"
     tag_format: "vX.Y.Z"
     merge_strategy: "merge-commit"
     reference_policy: "CLAUDE.local.md §18"
   ---

   Use Skill("moai-workflow-release") with arguments: $ARGUMENTS
   ```

   - 기존 frontmatter 핵심 메타(`disable-model-invocation`, `metadata.*`) 보존.
   - `version: 5.0.0` → `6.0.0` bump.
   - Body = 1 LOC.

3. 변환 후 LOC 검증:

   ```bash
   # body LOC만 카운트 (frontmatter 제외)
   awk '/^---$/{c++; next} c==2' .claude/commands/98-github.md | grep -cv '^$'
   awk '/^---$/{c++; next} c==2' .claude/commands/99-release.md | grep -cv '^$'
   ```

   둘 다 ≤20 (실제 1) 확인.

**검증**:

- AC-WF002-01 / 02 (body LOC ≤ 20) 자동 통과
- AC-WF002-07 / 08 (Skill 호출 + $ARGUMENTS 전달) 본문 grep 통과
- Frontmatter `description`, `argument-hint` 보존 (기존 사용자가 `--help` 시 동일 메시지)

**의존성**: M2 완료 (skill에 logic 이관 후에만 thin-wrapper 안전).

**충족 REQ**: REQ-WF002-001, REQ-WF002-002, REQ-WF002-007, REQ-WF002-008.

---

### M4 — Audit 테스트 확장 + DEV_ONLY_SKILL_LEAK 게이트

**목표**: `commands_audit_test.go`에서 root-level 2개 command를 명시적으로 검증 대상에 포함하고, 신규 `dev_only_skill_test.go`로 template 누출을 차단한다.

**산출물**:

- `internal/template/commands_audit_test.go` 수정
- `internal/template/dev_only_skill_test.go` 신설

**작업 내역**:

1. **commands_audit_test.go 수정** (research.md §3.2 분석 기반):
   - 현재 `rootCommands` 화이트리스트는 `agency/agency.md` 만 포함.
   - 만약 `98-github.md` / `99-release.md`가 묵시적 화이트리스트로 fail-skip 되고 있다면 해당 로직 제거.
   - 명시적 정책 주석 추가:
     ```go
     // Root-level dev-only commands MUST also follow thin-wrapper pattern.
     // SPEC-V3R2-WF-002 REQ-WF002-005 / 009: 98-github.md and 99-release.md
     // were extracted into moai-workflow-github / moai-workflow-release skills.
     ```
   - 검증 대상 명시화: 두 파일이 fs.WalkDir 결과에 포함되는지 sub-test 형태로 확인.

2. **dev_only_skill_test.go 신설** (REQ-WF002-014 구현):
   ```go
   package template

   import (
       "io/fs"
       "strings"
       "testing"
   )

   // TestDevOnlySkillLeak ensures dev-only skills moai-workflow-github and
   // moai-workflow-release are NOT registered in the user-facing template tree.
   //
   // Source: SPEC-V3R2-WF-002 REQ-WF002-014
   func TestDevOnlySkillLeak(t *testing.T) {
       t.Parallel()

       fsys, err := EmbeddedTemplates()
       if err != nil {
           t.Fatalf("EmbeddedTemplates() error: %v", err)
       }

       devOnlySkills := []string{
           "moai-workflow-github",
           "moai-workflow-release",
       }

       _ = fs.WalkDir(fsys, ".claude/skills", func(path string, d fs.DirEntry, err error) error {
           if err != nil || !d.IsDir() {
               return nil
           }
           for _, leak := range devOnlySkills {
               if strings.HasSuffix(path, "/"+leak) {
                   t.Fatalf("DEV_ONLY_SKILL_LEAK: skill %q found at %q. "+
                       "This skill is dev-only (REQ-WF002-014, SPEC-V3R2-WF-002).",
                       leak, path)
               }
           }
           return nil
       })
   }
   ```

3. **부분 마이그레이션 게이트** (REQ-WF002-015 — Complex):
   - `commands_audit_test.go`에 추가 검사: 각 root-level command body가 `Skill("<name>")` 호출 포함 시, `<name>` skill 디렉터리가 `.claude/skills/`에 존재해야 함.
   - 부분 추출 상태(wrapper만 줄이고 skill 미생성) 차단.

**검증**:

- `go test ./internal/template/...` 두 신규/확장 테스트 통과.
- 의도적 leak (e.g., `internal/template/templates/.claude/skills/moai-workflow-github/SKILL.md` 임시 생성) 시 `TestDevOnlySkillLeak` fail 확인 후 rollback.

**의존성**: M3 완료 (thin-wrapper 변환 후에만 audit 테스트 자연 통과).

**충족 REQ**: REQ-WF002-005, REQ-WF002-009, REQ-WF002-011, REQ-WF002-014, REQ-WF002-015.

---

### M5 — 빌드 / 통합 검증

**목표**: 전체 회귀 없음 확인 + go test all-pass + make build 성공.

**산출물**: 검증 로그 (commit 노트 또는 PR description에 첨부).

**작업 내역**:

1. **테스트 전수 실행**:
   ```bash
   go test -count=1 ./...
   go test -race -run TestCommandsThinPattern ./internal/template/...
   go test -run TestDevOnlySkillLeak ./internal/template/...
   ```
2. **빌드 검증**:
   ```bash
   make build
   make install
   moai version
   ```
3. **수동 검증** (Claude Code 환경):
   - `/98-github issues --all` 호출 시 wrapper가 skill로 위임되는지 (실제 호출은 dev 환경에서만 가능, 본 SPEC은 배포 대상 아님)
   - `/99-release` 호출 시 동일 위임 확인
4. **diff-parity 최종 확인** (M2의 grep 산출물 비교):
   ```bash
   diff before-phases.txt after-phases.txt    # 빈 출력
   diff before-flags.txt after-flags.txt       # 빈 출력
   ```

**검증**:

- 모든 테스트 PASS
- `make build` zero-warning
- 의존성 SPEC (V3R2-MIG-002, V3R2-WF-003)에 영향 없음 — git status에서 의도하지 않은 파일 변경 없음

**의존성**: M4 완료.

**충족 REQ**: REQ-WF002-009 (post-refactor go test 통과), REQ-WF002-010 (dev-only exception 적용).

---

## 4. 마일스톤 의존성 그래프

```
        M1 (skill 골격)
           |
           v
        M2 (logic 추출)  ← parity check (REQ-013 완화)
           |
           v
        M3 (thin-wrapper 변환)
           |
           v
        M4 (audit 테스트 확장 + DEV_ONLY_SKILL_LEAK)
           |
           v
        M5 (전체 검증)
```

분기 없음. 각 단계의 출력이 다음 단계의 입력이다. M2와 M3을 동시에 진행하면 logic 손실 위험 (parity check 불가)하므로 순차 실행이 강제된다.

---

## 5. REQ → Milestone 매핑 (15개)

| REQ ID | 분류 | 충족 마일스톤 | 비고 |
|--------|------|----------------|------|
| REQ-WF002-001 | Ubiquitous (98 ≤20 LOC) | M3 | thin-wrapper 변환 |
| REQ-WF002-002 | Ubiquitous (99 ≤20 LOC) | M3 | thin-wrapper 변환 |
| REQ-WF002-003 | Ubiquitous (GitHub logic 보존) | M2 | logic 추출 + parity |
| REQ-WF002-004 | Ubiquitous (Release logic 보존) | M2 | logic 추출 + parity |
| REQ-WF002-005 | Ubiquitous (audit test 확장) | M4 | commands_audit_test.go 수정 |
| REQ-WF002-006 | Ubiquitous (user-invocable false) | M1 | frontmatter |
| REQ-WF002-007 | Event-Driven (Skill 호출 - github) | M3 | wrapper body |
| REQ-WF002-008 | Event-Driven (Skill 호출 - release) | M3 | wrapper body |
| REQ-WF002-009 | Event-Driven (go test 통과) | M5 | 통합 검증 |
| REQ-WF002-010 | State-Driven (dev-only exception) | M1+M4 | frontmatter + leak test |
| REQ-WF002-011 | State-Driven (>20 LOC fail) | M4 | audit test 강화 |
| REQ-WF002-012 | Optional (future user-invocable flip) | M1 | 설계만, 구현 불필요 |
| REQ-WF002-013 | Unwanted (ordering 손실) | M2 | diff-parity 절차 |
| REQ-WF002-014 | Unwanted (template leak) | M4 | dev_only_skill_test.go |
| REQ-WF002-015 | Complex (partial migration gate) | M4 | audit test 추가 검사 |

→ 15/15 REQ 모두 마일스톤 매핑 완료.

---

## 6. 위험 요소 및 완화 (Risk Register)

| ID | 위험 | 영향 | 발생 가능성 | 완화 |
|----|------|------|-------------|------|
| R1 | 추출 시 Phase 순서 손실 | 릴리스 회귀 | 중 | M2 diff-parity check 강제 (REQ-013) |
| R2 | dev-only skill의 template 누출 | 사용자 혼란 | 저 | M4 dev_only_skill_test.go (REQ-014) |
| R3 | trigger keyword 충돌로 일반 사용자 활성 | UX 오염 | 저 | M1 frontmatter `user-invocable: false` + dev-only 키워드 |
| R4 | thin-wrapper에서 `$ARGUMENTS` 전달 실수 | 실행 실패 | 저 | M3 표준 패턴 그대로 차용 (`/moai/*` 15개 reference) |
| R5 | commands_audit_test.go의 묵시적 화이트리스트가 본 SPEC 검증 우회 | spec violation 미감지 | 중 | M4 명시적 검증 + sub-test 추가 |
| R6 | BC-V3R2-012 (breaking) 사용자 통보 누락 | 마이그레이션 confusion | 저 | sync 단계에서 CHANGELOG 명시 (본 plan 범위 외, sync 책임) |
| R7 | 99-release.md `metadata.*` 필드 누락 | release 자동화 회귀 | 중 | M3에서 핵심 metadata (`merge_strategy`, `reference_policy`) 보존 |
| R8 | wave 분할 없이 한 PR에 5 마일스톤 합쳐 stream stall | 작업 중단 | 저 | M1+M2를 1 commit, M3 1 commit, M4+M5 1 commit으로 분할 권장 |

---

## 7. 추가 도구 / 패키지

- **신규 외부 의존성**: 없음. Go 표준 라이브러리(`io/fs`, `testing`, `strings`)만 사용.
- **신규 CLI 도구**: 없음. `awk`, `grep`, `diff`는 macOS/Linux 표준.
- **빌드 도구**: 기존 `make build`, `go test` 그대로 사용.

---

## 8. Breaking Change 처리 (BC-V3R2-012)

spec.md frontmatter `breaking: true`, `bc_id: [BC-V3R2-012]`. 이는 **command 파일 frontmatter `version` bump**로 표현되며, 사용자 영향은 다음 측면에서 평가된다:

- **dev-only commands** (`98-github.md`, `99-release.md`)는 `internal/template/templates/`에 등록되지 않으므로 사용자 프로젝트에 배포되지 않음 → **사용자 측 breaking change 없음**.
- moai-adk-go maintainer 입장에서는 `/98-github`, `/99-release` 호출 시 내부 동작 (thin-wrapper → skill 위임)이 바뀌나, **외부 동작 (issues 처리, release 워크플로)은 보존** (REQ-WF002-003 / 004).
- BC-V3R2-012 통보는 sync 단계의 CHANGELOG 책임이며 본 plan에서는 다루지 않음.

---

## 9. Sync 단계 인계 사항 (Plan→Run→Sync 흐름)

본 plan은 run 단계의 입력이다. sync 단계에서 다뤄야 할 사항은 다음과 같다 (본 plan 범위 외 — 메모용):

- BC-V3R2-012 CHANGELOG entry (한국어/영어)
- 신규 두 skill 파일을 `.claude/rules/moai/`에 등록할 필요는 없음 (skill catalog 자동 발견)
- 24-skill 카탈로그 변경 없음 (dev-only 별도 취급)
- docs-site 4개국어 동기화 불필요 (dev-only 도구)

---

## 10. 시작 조건 체크리스트 (run 단계 진입 전)

- [x] spec.md 변경 없음 확인 (본 작업 제약)
- [x] SPEC-V3R2-WF-001 main 머지 확인 (Wave 3까지 머지됨)
- [x] worktree `feat/SPEC-V3R2-WF-002-thin-wrapper` 활성
- [x] research.md 작성 완료
- [x] plan.md (본 문서) 작성 완료
- [ ] acceptance.md 작성 완료 (다음 단계)
- [ ] progress.md 작성 완료 (다음 단계)
- [ ] `/moai run SPEC-V3R2-WF-002` 진입 가능

End of plan.md.
