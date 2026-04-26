---
id: SPEC-V3R3-PATTERNS-001
version: "0.1.0"
status: draft
created_at: 2026-04-26
updated_at: 2026-04-26
author: manager-spec
priority: High
labels: [patterns, harness, tasks, v3r3]
issue_number: null
---

# SPEC-V3R3-PATTERNS-001 — Task Breakdown

> Wave-based execution. 각 wave는 sequential. 동일 wave 내 task는 단일 sub-agent에게 한 번에 위임.

---

## Wave 1 — Prior Art Analysis (Read-only)

**Owner**: Explore agent (general-purpose, mode: plan)
**Isolation**: None (read-only, 임시 디렉토리 사용)
**Token estimate**: ~15K (clone metadata + 6 reference 분석)

| ID | Task | Tool/Action | Output |
|----|------|-------------|--------|
| T1.1 | harness 저장소 clone (depth=1) | `git clone --depth=1 https://github.com/revfactory/harness.git /tmp/harness-analysis/harness` | `/tmp/harness-analysis/harness/` |
| T1.2 | LICENSE 검증 | `head -10 /tmp/harness-analysis/harness/LICENSE` | Apache 2.0 확인 또는 EC-001 트리거 |
| T1.3 | 6 reference 파일 존재 확인 | `ls /tmp/harness-analysis/harness/skills/harness/references/` | 6 파일 모두 확인 또는 EC-002 트리거 |
| T1.4 | 각 reference 핵심 섹션 추출 | Read each file, extract structure | progress.md에 import 매핑 표 작성 |
| T1.5 | NOTICE 파일 사본 확보 | `cat /tmp/harness-analysis/harness/NOTICE` (있다면) | NOTICE 텍스트 또는 "NOTICE absent" 기록 |

**Wave 1 Exit**: progress.md에 다음 모두 기록:
- LICENSE: Apache 2.0 (확정)
- 6 reference 파일 path 일람
- import 매핑 표 (source section → target section)
- NOTICE 텍스트 또는 부재 사실

---

## Wave 2 — Cookbook Authoring (Write)

**Owner**: manager-tdd (sub-agent, acceptEdits)
**Isolation**: None (직접 write)
**Token estimate**: ~80K (6 파일 × 평균 13K + NOTICE)

### T2.1 — 디렉토리 신규 생성

```bash
mkdir -p .claude/rules/moai/quality
mkdir -p internal/template/templates/.claude/rules/moai/quality
```

### T2.2 — `.claude/rules/moai/NOTICE.md` 작성

내용:
- 헤더: "MoAI-ADK Third-Party Notices"
- Apache 2.0 LICENSE 텍스트 (전체 사본 또는 URL + summary)
- 흡수된 source 일람 (revfactory/harness, 6 파일, import 날짜 2026-04-26)
- attribution clause: "This product includes software developed by revfactory/harness contributors."
- harness NOTICE 파일 사본 (있다면)

### T2.3 — `agent-patterns.md` 작성

경로: `.claude/rules/moai/development/agent-patterns.md`
- frontmatter: `paths: ".claude/agents/**/*.md,.claude/rules/moai/development/agent-authoring.md"`
- 상단 attribution 주석
- 6 patterns 섹션: Team / Sub-agent / Hybrid / Orchestrator / Specialist / Pipeline
- 각 pattern: definition / when to use / example / anti-pattern
- harness reference로부터 import + 16-language neutralization

### T2.4 — `boundary-verification.md` 작성

경로: `.claude/rules/moai/quality/boundary-verification.md`
- frontmatter: `paths: "**/*_test.go,**/*test*.py,**/*test*.ts,**/*test*.rs,**/test/**,**/tests/**"`
- 상단 attribution 주석
- 7 bug case: case header 7개
- 각 case: symptom / root cause / boundary condition / verification strategy
- 16-language neutralization

### T2.5 — `skill-ab-testing.md` 작성

경로: `.claude/rules/moai/development/skill-ab-testing.md`
- frontmatter: `paths: ".claude/skills/**/SKILL.md,.claude/skills/**/*.md"`
- 상단 attribution 주석
- A/B methodology 섹션: with-skill vs baseline / evaluation criteria / sample size / statistical significance

### T2.6 — `team-pattern-cookbook.md` 작성

경로: `.claude/rules/moai/workflow/team-pattern-cookbook.md`
- frontmatter: `paths: ".moai/config/sections/workflow.yaml,.claude/rules/moai/workflow/team-protocol.md"`
- 상단 attribution 주석
- 5 팀 예시: Research / Implementation / Review / Design / Debug
- 각 팀: role profile / file ownership / communication / shutdown sequence

### T2.7 — `orchestrator-templates.md` 작성

경로: `.claude/rules/moai/development/orchestrator-templates.md`
- frontmatter: `paths: ".claude/rules/moai/core/moai-constitution.md,CLAUDE.md"`
- 상단 attribution 주석
- 3 templates: Team-orchestrator / Sub-orchestrator / Hybrid-orchestrator
- 각 template: when to use / how to spawn / error recovery / escalation

### T2.8 — `skill-writing-craft.md` 작성

경로: `.claude/rules/moai/development/skill-writing-craft.md`
- frontmatter: `paths: ".claude/skills/**/SKILL.md"`
- 상단 attribution 주석
- 3 영역: description craft / 3-level progressive disclosure / schema validation
- 16-language neutralization

**Wave 2 Exit**: 7 파일 모두 작성 완료. 각 파일 head 30줄 review로 frontmatter + attribution 검증.

---

## Wave 3 — Template-First Sync + Build (Write)

**Owner**: manager-git (sub-agent, acceptEdits)
**Isolation**: None
**Token estimate**: ~10K

### T3.1 — Template mirror 동기화

```bash
# Local → Template byte-identical mirror
for sub in development/agent-patterns.md \
           development/skill-ab-testing.md \
           development/orchestrator-templates.md \
           development/skill-writing-craft.md \
           quality/boundary-verification.md \
           workflow/team-pattern-cookbook.md \
           NOTICE.md; do
  cp -f ".claude/rules/moai/$sub" \
        "internal/template/templates/.claude/rules/moai/$sub"
done
```

### T3.2 — `make build` 실행

```bash
make build
```

성공 시 `internal/template/embedded.go` 재생성 확인.

### T3.3 — Go test 회귀 검증

```bash
go test ./internal/template/...
```

특히 `commands_audit_test.go` 통과 확인.

### T3.4 — `diff` 검증

```bash
for sub in development/agent-patterns.md ...; do
  diff ".claude/rules/moai/$sub" \
       "internal/template/templates/.claude/rules/moai/$sub" || \
    echo "MIRROR DRIFT: $sub"
done
# Expected: no output
```

**Wave 3 Exit**: build 성공 + test green + mirror diff 0.

---

## Wave 4 — Validation (Test)

**Owner**: manager-quality (sub-agent, mode: plan)
**Isolation**: None (read-only)
**Token estimate**: ~10K

### T4.1 — AC-001 ~ AC-005 검증

`acceptance.md`의 각 AC 검증 명령 실행 + 결과 progress.md 기록.

### T4.2 — 16-language neutrality grep

```bash
files=( "agent-patterns.md" "skill-ab-testing.md" "orchestrator-templates.md" \
        "skill-writing-craft.md" "boundary-verification.md" "team-pattern-cookbook.md" )
banned=("pip install" "go get " "npm install" "cargo install")
# 단독 표기 검사 → flagged 파일 manual review
```

### T4.3 — frontmatter `paths` 글로브 매칭 확인

각 파일 frontmatter `paths`가 의도된 파일 패턴과 일치하는지 manual review.

### T4.4 — 최종 리포트 작성

progress.md에 Wave 1-4 전체 결과 + 미해결 OQ 목록 + AC trace matrix 기록.

**Wave 4 Exit**: 모든 AC PASS, plan-auditor (있다면) PASS, status를 `draft` → `audit-ready`로 변경 가능 상태.

---

## AC ↔ REQ ↔ Task Trace Matrix

| AC | REQs | Tasks |
|----|------|-------|
| AC-001 | REQ-PAT-001 | T2.3, T3.1, T3.4, T4.1 |
| AC-002 | REQ-PAT-002, REQ-PAT-003 | T2.4, T2.5, T3.1, T3.4, T4.1 |
| AC-003 | REQ-PAT-004, REQ-PAT-005 | T2.6, T2.7, T3.1, T3.4, T4.1 |
| AC-004 | REQ-PAT-006 | T2.8, T3.1, T3.4, T4.1 |
| AC-005 | REQ-PAT-007, REQ-PAT-008 | T2.1, T2.2, T3.1, T3.2, T3.3, T3.4, T4.1, T4.2 |

| REQ | ACs | Tasks |
|-----|-----|-------|
| REQ-PAT-001 | AC-001 | T2.3 |
| REQ-PAT-002 | AC-002 | T2.4 |
| REQ-PAT-003 | AC-002 | T2.5 |
| REQ-PAT-004 | AC-003 | T2.6 |
| REQ-PAT-005 | AC-003 | T2.7 |
| REQ-PAT-006 | AC-004 | T2.8 |
| REQ-PAT-007 | AC-005 | T2.2, T4.1 |
| REQ-PAT-008 | AC-005 | T3.1, T3.2, T3.4 |

Coverage: 100% (모든 REQ가 적어도 1 AC + 1 Task에 매핑, 모든 AC가 적어도 1 REQ + 1 Task에 매핑).

---

## Risks recap (from plan.md §4)

- LICENSE 불일치 (R-001) → Wave 1에서 abort path
- reference 부재 (R-002) → Wave 1에서 사용자 협의
- 16-language 위반 (R-003) → Wave 2 자가검증 + Wave 4 grep
- mirror drift (R-004) → Wave 3 diff 검증
- frontmatter glob 오류 (R-005) → Wave 4 manual review
- NOTICE 누락 (R-006) → T2.2에서 별도 작성 + AC-005 grep
