---
id: SPEC-V3R2-WF-002
document: research
version: "0.1.0"
status: audit-ready
created: 2026-04-30
updated: 2026-04-30
author: manager-spec (plan-phase enrichment)
related_spec: SPEC-V3R2-WF-002
phase: plan
language: ko
---

# SPEC-V3R2-WF-002 — Research (코드베이스 분석)

> 본 문서는 plan 단계의 사실 확인용 리서치 결과이다. spec.md의 가정(§4)과 제약(§7)을 코드 기반 증거로 검증하고, run 단계에서 참조 가능한 anchor를 모아둔다. 새 요구사항은 추가하지 않는다.

---

## 1. 현재 상태 — Fat Command 2종 분석

### 1.1 LOC 실측 (worktree HEAD 기준)

| 파일 | LOC (실측) | spec.md §1 명시 | 차이 |
|------|------------|------------------|------|
| `.claude/commands/98-github.md` | 698 | 698 | 0 (일치) |
| `.claude/commands/99-release.md` | 933 | 890 | +43 (spec.md 작성 이후 본문 갱신, **§5.1 REQ-WF002-002 ≤ 20 LOC 목표는 동일 적용**) |

[ANCHOR] 본 차이는 spec.md를 변경하지 않으며, 추출 시점에 99-release.md 최신 본문 933 LOC 전체를 `moai-workflow-release` skill로 이관함을 의미한다. REQ-WF002-002의 수용 기준은 "리팩토링 후 body ≤ 20 LOC"이므로 시작 LOC 변동과 무관하다.

### 1.2 98-github.md 구조 요약 (logic 분류)

`.claude/commands/98-github.md` 본문은 다음 섹션으로 구성된다 (앵커는 §2):

1. **Frontmatter** (lines 1–6): `description`, `argument-hint`, `type: local`, `allowed-tools` (CSV), `version`.
2. **Configuration** (§ "GitHub Workflow Configuration"): repo 자동탐지, 기본 모드, 브랜치 prefix, git strategy 참조.
3. **Execution Directive** (§ "EXECUTION DIRECTIVE - START IMMEDIATELY"): `$ARGUMENTS` 파싱 — `issues` / `pr` 분기 + 보조 플래그 (`--all`, `--label`, `--solo`, `--merge`, NUMBER).
4. **Pre-execution Context**: `gh repo view`, `git branch`, `git status`, `printenv CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` + `@.moai/config/sections/{system,language,workflow}.yaml` import.
5. **Sub-command 1 — `issues`**: GitHub Issue 처리 워크플로 (Agent Teams 분기, fix/issue-{N} 브랜치 생성, --merge 자동 머지).
6. **Sub-command 2 — `pr`**: PR review 워크플로 (Code review, label 부착, merge 전략).
7. **AskUserQuestion fallback**: sub-command 미지정 시 Socratic 인터뷰.

`Skill()` 호출은 본문 어디에도 등장하지 않는다 (현재 fat-command anti-pattern). 모든 orchestration이 commands 레이어에서 하드코딩.

### 1.3 99-release.md 구조 요약 (logic 분류)

`.claude/commands/99-release.md` 본문은 다음 섹션으로 구성된다:

1. **Frontmatter** (lines 1–18): `description`, `argument-hint`, `type: local`, `allowed-tools` (CSV), `disable-model-invocation: true`, `version`, `metadata` (release_target, branch, tag_format, changelog_format, merge_strategy 등).
2. **Release Configuration (Enhanced GitHub Flow)**: CLAUDE.local.md §18 참조, release 브랜치 명명, hotfix 패턴, merge commit 강제, GoReleaser 트리거 정책.
3. **Tool Selection Guide (§18.8 이중 경로)**: `./scripts/release.sh` vs `/99-release` 사용 결정표.
4. **Execution Directive**: `$ARGUMENTS` 파싱 (VERSION, `--hotfix`).
5. **Phase 1 — 사전 검증**: 현재 브랜치 = main, 작업트리 clean, 최신 tag 조회.
6. **Phase 2 — 버전 결정**: VERSION 미제공 시 patch/minor/major 선택 AskUserQuestion.
7. **Phase 3 — Release 브랜치 생성**: `release/vX.Y.Z` 분기 (또는 `hotfix/vX.Y.Z-*`).
8. **Phase 4 — Bilingual CHANGELOG 생성**: 한국어/영어 섹션 작성, version bump 적용.
9. **Phase 5 — PR 생성/머지**: gh pr create → merge commit (`--merge`, NOT `--squash`).
10. **Phase 6 — Tag + GoReleaser**: `./scripts/release.sh` 또는 `make release V=vX.Y.Z` 실행, 6 binary asset + checksums 검증.
11. **Phase 7 — Post-release**: docs-site 동기화 안내, 후속 PR 권장.
12. **Quality escalation**: 실패 시 `expert-debug` 위임.

`Skill()` 호출 없음. manager-git, expert-debug 등 위임 흐름이 commands 레이어에 직접 기술됨.

---

## 2. 비교 기준 — 이미 thin-wrapper인 `/moai/*` 15개 패턴

### 2.1 표준 thin-wrapper body 형태 (실측)

`.claude/commands/moai/*.md` 15개 파일은 모두 **6–8 LOC body** + frontmatter로 구성. 대표 예시 `moai/plan.md`:

```yaml
---
description: Create SPEC document with EARS format requirements and acceptance criteria
argument-hint: "\"description\" [--team] [--worktree] [--branch] [--resume SPEC-XXX]"
allowed-tools: Skill
---

Use Skill("moai") with arguments: plan $ARGUMENTS
```

- Body LOC = **1줄** (non-empty), frontmatter 제외 시 thin-wrapper 규격(≤20 LOC) 압도적 만족.
- Frontmatter 필수 필드: `description`, `argument-hint`, `allowed-tools` (CSV).
- Body 필수 패턴: `Skill("<skill-name>")` 호출 + `$ARGUMENTS` 전달.

### 2.2 다른 15개 subcommand 카탈로그 (참조용)

| 파일 | Body LOC (실측) | 위임 Skill |
|------|-----------------|-------------|
| `clean.md` | 1 | `moai` (clean) |
| `codemaps.md` | 1 | `moai` (codemaps) |
| `coverage.md` | 1 | `moai` (coverage) |
| `db.md` | 1 | `moai` (db) |
| `design.md` | 1 | `moai` (design) |
| `e2e.md` | 1 | `moai` (e2e) |
| `feedback.md` | 1 | `moai` (feedback) |
| `fix.md` | 1 | `moai` (fix) |
| `gate.md` | 1 | `moai` (gate) |
| `loop.md` | 1 | `moai` (loop) |
| `mx.md` | 1 | `moai` (mx) |
| `plan.md` | 1 | `moai` (plan) |
| `project.md` | 1 | `moai` (project) |
| `review.md` | 1 | `moai` (review) |
| `run.md` | 1 | `moai` (run) |
| `sync.md` | 1 | `moai` (sync) |

→ `moai-workflow-github` / `moai-workflow-release` thin-wrapper도 동일 패턴 채택. 단, `Use Skill("moai-workflow-github") with arguments: $ARGUMENTS` 처럼 unified `/moai` skill을 거치지 않고 **신규 dev-only skill을 직접 호출** (REQ-WF002-007 / 008).

---

## 3. `commands_audit_test.go` 현재 검증 로직

### 3.1 핵심 함수 — `TestCommandsThinPattern`

`internal/template/commands_audit_test.go:14–104` 테스트 (worktree HEAD):

```go
// 1. .claude/commands 하위 모든 .md / .md.tmpl 수집 (재귀)
fs.WalkDir(fsys, ".claude/commands", ...)

// 2. rootCommands 화이트리스트 (router 로직 허용)
rootCommands := map[string]bool{
    ".claude/commands/agency/agency.md": true,
}

// 3. 각 파일에 대한 검증 (R1-R4):
//    R1: frontmatter 필수 필드 (description, allowed-tools)
//    R2: allowed-tools는 CSV string (YAML array 금지)
//    R3 (서브커맨드 한정): body non-empty LOC < 20
//    R4 (서브커맨드 한정): body에 Skill() 호출 포함
```

### 3.2 현재 동작 — Root-Level 2종이 검증되지 않는 이유

`fs.WalkDir(".claude/commands", ...)`은 root-level `98-github.md` / `99-release.md`를 **포함**하여 순회한다. 그러나 두 파일은 `rootCommands` 화이트리스트에 등재되지 않았으므로 **현재 R3 / R4 검증이 적용된다**.

[ANCHOR-RESEARCH-001] 실제 테스트 실행 시 두 파일은 이미 R3 위반(body LOC ≥ 20) + R4 위반(`Skill(` 미포함)으로 **fail해야 한다**. 현재 main에서 본 테스트가 통과하고 있다면, 두 파일이 `agency/agency.md`와 동일한 router 예외로 묵시적 처리되고 있을 가능성이 있다.

→ run 단계 시작 시 다음 검증을 우선 수행:

```bash
go test -run TestCommandsThinPattern ./internal/template/... -v
```

- 만약 이미 fail → REQ-WF002-005 / 009는 "기존 fail을 유지"가 아니라 "추출 후 PASS"로 명세됨. M3 종료 시점에 두 파일이 thin-wrapper로 변환되어 자동 PASS 예정.
- 만약 PASS (= 묵시적 화이트리스트 존재) → `rootCommands` 화이트리스트에서 두 파일을 명시적으로 제외하는 추가 변경 필요 (REQ-WF002-005에 자연 포함).

### 3.3 보조 테스트 — `TestCommandsFrontmatterConsistency`

`commands_audit_test.go:105–150` (deprecated field 검사):

- 금지 필드: `tools`, `disallowed-tools`, `model`.
- `99-release.md`에는 현재 `disable-model-invocation: true` + `metadata.*` 다수 필드가 존재 — 이는 deprecated 목록에 없으므로 thin-wrapper 전환 시 **보존 가능**.

---

## 4. Skill 프론트매터 패턴 (신규 skill 2종 적용)

### 4.1 dev-only skill 식별자

| 필드 | 값 | 출처 |
|------|----|------|
| `name` | `moai-workflow-github` / `moai-workflow-release` | spec.md §2.1 |
| `user-invocable` | `false` | REQ-WF002-006 |
| `description` | "(dev-only) GitHub workflow ..." / "(dev-only) Release workflow ..." | spec.md §1.1 dev-local only 명시 |
| `allowed-tools` (CSV) | 추출 대상 command의 `allowed-tools` 동등 + `Skill` 추가 가능 | 98-github.md / 99-release.md 본문 frontmatter |
| `metadata.category` | `"workflow"` | 기존 `moai-workflow-spec/SKILL.md` 관례 |
| `metadata.status` | `"active"` | 동일 |
| `metadata.tags` | `"workflow, github, dev-only, ..."` / `"workflow, release, dev-only, ..."` | 명시적 dev-only 태그 |

### 4.2 참조 구현 — `moai-workflow-spec/SKILL.md` 스니펫

```yaml
---
name: moai-workflow-spec
description: >
  SPEC workflow orchestration with EARS format ...
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Bash(git:*), ...
user-invocable: false
metadata:
  version: "1.2.0"
  category: "workflow"
  status: "active"
  ...
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000
triggers:
  keywords: ["SPEC", "requirement", "EARS", ...]
---
```

→ 두 신규 skill도 동일 골격 채택. 단, `triggers.keywords`는 **dev-only이므로 일반 사용자 trigger 회피** (e.g., `["github-dev-workflow", "release-dev-workflow"]` 처럼 일반 키워드와 충돌하지 않게 설계 — REQ-WF002-006의 "user-invocable: false" + risk §8 "trigger keyword가 일반 user에게 활성화됨" 완화책).

### 4.3 참조 사례 — 기존에 추출된 dev-only skill 부재

- worktree 내 `.claude/skills/` 33개 디렉터리 중 `user-invocable: false` 이며 commands에서 직접 호출되는 dev-only skill **선례 없음** (대부분 `moai` skill 경유).
- 본 SPEC이 첫 dev-only direct-call skill 사례 → CI에서 `DEV_ONLY_SKILL_LEAK` 검사 신설 필요 (REQ-WF002-014).

---

## 5. CI / Build 게이트 — `DEV_ONLY_SKILL_LEAK`

### 5.1 검사 위치

REQ-WF002-014: "If `moai-workflow-github` or `moai-workflow-release` is accidentally added to `internal/template/templates/.claude/skills/`, then `make build` shall fail with `DEV_ONLY_SKILL_LEAK`."

→ 검증 대상 경로:

- 허용: `.claude/skills/moai-workflow-github/` (worktree dev tree)
- 금지: `internal/template/templates/.claude/skills/moai-workflow-github/` (사용자 프로젝트 배포 경로)

### 5.2 구현 후보 위치

1. **신규 테스트** (선호): `internal/template/dev_only_skill_test.go` 신설 — `EmbeddedTemplates()` fs walk + 두 skill 이름이 등장하면 `t.Fatal("DEV_ONLY_SKILL_LEAK: ...")`.
2. Makefile target (보조): `make build` 시 `find internal/template/templates/.claude/skills -name 'moai-workflow-github' -o -name 'moai-workflow-release'` 결과가 비어있는지 확인.

run 단계에서 (1)을 우선 채택 — 기존 `commands_audit_test.go`와 동일 파일 옆에 위치하므로 일관성 확보.

### 5.3 진단 메시지 형식

```
FAIL: DEV_ONLY_SKILL_LEAK: skill `moai-workflow-github` found at `internal/template/templates/.claude/skills/moai-workflow-github/SKILL.md`. This skill is dev-only (REQ-WF002-014, SPEC-V3R2-WF-002). Move it back to `.claude/skills/` or remove it.
```

---

## 6. SPEC-V3R2-WF-001 (24-skill 카탈로그)와의 관계

### 6.1 의존성 검증

spec.md §9.1: "SPEC-V3R2-WF-001: 24-skill 카탈로그가 먼저 확정되어야 dev-only skill 2종의 별도 취급이 가능."

[ANCHOR-RESEARCH-002] worktree git log 확인:

```
3f0933550 feat(v3r2): Wave 3 — Permission Stack, MX TAG v2, Hook Handler, Constitution Pipeline (#741)
```

→ Wave 3까지 main에 머지됨. SPEC-V3R2-WF-001은 사용자 메시지에서 "이미 main에 머지됨"으로 명시 → **dependency 충족**.

### 6.2 24-skill 카탈로그 vs dev-only 2 skill

- spec.md §6.2 판정표 (Constraints): "본 SPEC의 2개 skill은 dev-only로 별도 취급" → 24-skill 카탈로그 수치에 영향 없음.
- 신규 skill 2종은 `.claude/skills/`에만 존재, 사용자 프로젝트로 배포되지 않으므로 카탈로그 검증 대상 외.
- WF-001이 정의한 "user-facing skill 24종" 집합과 비교 시: WF-002 두 skill은 **카탈로그 외부**.

---

## 7. Behavior Preservation 검증 전략 (REQ-WF002-013 대비)

### 7.1 risk: release-step ordering 손실

`99-release.md`의 Phase 1–7 순서가 추출 후 `moai-workflow-release/SKILL.md`에서 동일 순서로 보존되어야 한다. 자동 검증이 어려우므로 **diff-parity 체크**를 manual 절차로 정의:

1. 추출 전: `99-release.md` body의 Phase heading만 `grep '^## Phase'` 추출 → `before-phases.txt`
2. 추출 후: `moai-workflow-release/SKILL.md`에서 동일 grep → `after-phases.txt`
3. `diff before-phases.txt after-phases.txt`가 빈 결과여야 PASS.

### 7.2 risk: 98-github.md sub-command 분기 손실

`issues` / `pr` 분기 로직, `--all` / `--label` / `--solo` / `--merge` / NUMBER 플래그가 추출 후에도 동일하게 동작해야 한다. 동일하게 grep 기반 parity check:

1. 추출 전: `grep -E '^- \*\*(issues|pr|--)' 98-github.md` → `before-flags.txt`
2. 추출 후: 동일 grep 대상 `moai-workflow-github/SKILL.md` → `after-flags.txt`
3. diff empty여야 PASS.

→ 이 절차는 acceptance.md AC-WF002-11에 GWT 형태로 명시.

---

## 8. 외부 anchor 인덱스 (run 단계 참조용)

| anchor | 경로 |
|--------|------|
| Thin Command Pattern 정책 | `.claude/rules/moai/development/coding-standards.md#thin-command-pattern` |
| R6 audit §1.1 (15개 thin-wrapper 현황) | `docs/design/major-v3-master.md` (history 검증, 본 worktree에는 미존재 가능) |
| R6 audit §1.2 (fat command 권장사항) | 동일 |
| Problem catalog | `.moai/design/v3-redesign/synthesis/problem-catalog.md` (P-H08, P-H09, P-H18) |
| 기존 thin-wrapper 표준 | `.claude/commands/moai/plan.md` (모범 사례) |
| Skill frontmatter 표준 | `.claude/skills/moai-workflow-spec/SKILL.md:1-30` |
| commands_audit_test.go 검증 로직 | `internal/template/commands_audit_test.go:14–150` |

---

## 9. 결론 & run 단계 시작 조건

- 의존성 SPEC-V3R2-WF-001 머지 완료 → **시작 가능**.
- spec.md 가정 5종(§4) 모두 코드 기반 증거로 확인됨.
- 추출 대상 LOC: 98-github.md 698 / 99-release.md 933 (총 1631 LOC → 2개 thin-wrapper ≤40 LOC + 2개 SKILL.md).
- run 단계 첫 작업: M1 (skill 골격 생성) — plan.md §4 milestones 참조.

End of research.md.
