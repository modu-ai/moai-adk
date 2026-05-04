<!-- Verifies REQ-BRAIN-001: 7 phases execute sequentially -->
<!-- Verifies REQ-BRAIN-002: Discovery rounds capped at 5 -->
<!-- Verifies REQ-BRAIN-009: Phase 7 exit AskUserQuestion with 3 options -->
<!-- Verifies REQ-BRAIN-010: NO auto-execution of /moai project -->
<!-- Verifies REQ-BRAIN-012: NO prose questions (AskUserQuestion only) -->
---
name: moai-workflow-brain
description: >
  Brain workflow orchestration: 7-phase idea-to-proposal pipeline with Claude Design
  handoff package. Use for /moai brain invocations — converts vague ideas into validated
  product proposals with SPEC decomposition candidates.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-05-04"
  tags: "brain, ideation, workflow, handoff, claude-design, proposal, spec-decomposition"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["brain", "idea", "ideation", "brain workflow"]
  agents: ["manager-brain"]
  phases: ["brain"]
---

<!-- @MX:NOTE: [AUTO] 7-phase brain workflow — derived from moai-foundation-thinking composition -->
<!-- The 7 phases map to foundation-thinking modules as follows:
  Phase 1 Discovery → modules/deep-questioning.md (Socratic interview)
  Phase 2 Diverge   → modules/diverge-converge.md (Diverge step)
  Phase 3 Research  → moai-domain-research (parallel WebSearch + Context7)
  Phase 4 Converge  → modules/diverge-converge.md (Converge step)
  Phase 5 Critical  → modules/critical-evaluation.md + modules/first-principles.md
  Phase 6 Proposal  → moai-domain-ideation (SPEC decomposition assembly)
  Phase 7 Handoff   → moai-domain-design-handoff (5-file package)
  New logic in this workflow: IDEA-NNN auto-increment, brand context detection, phase gating
-->

# Brain Workflow Orchestration

The `/moai brain` workflow converts vague ideas into validated product proposals with Claude Design handoff packages. It is a pre-spec ideation workflow — it runs BEFORE `/moai project` and `/moai plan`.

## Workflow Position

```
brain (once) → [user external claude.com Design] → design --path A → project (once) → plan (per SPEC) → run → sync
```

`brain` and `project` are run-once artifacts. `plan/run/sync` repeat per SPEC.

## Input

`$ARGUMENTS` — the user's idea, in any language, any form, any level of vagueness.

## Pre-Execution Setup

### IDEA-NNN Auto-Increment

Before Phase 1, determine the next IDEA number:

1. List `.moai/brain/IDEA-*/` directories (Glob pattern: `.moai/brain/IDEA-[0-9]*/`)
2. Extract the numeric suffix from each directory name
3. Take max + 1; if none exist, start at 001
4. Zero-pad to 3 digits: IDEA-001, IDEA-002, ..., IDEA-999
5. Create the directory: `.moai/brain/IDEA-NNN/`

Resume detection: If `.moai/brain/IDEA-NNN/` already exists with partial files, offer resume via AskUserQuestion (see Edge Case: Mid-Workflow Interrupt below).

### Brand Context Detection

Check: does `.moai/project/brand/brand-voice.md` exist and is it non-empty?

Set `brand_present` flag accordingly. Pass to Phase 7 (moai-domain-design-handoff).

---

## Phase 1: Discovery

**Purpose**: Clarify the idea to a clarity score of 4+ out of 5.
**Foundation reuse**: `moai-foundation-thinking` modules/deep-questioning.md
**Key constraint**: Max 5 rounds (REQ-BRAIN-002). In practice, 1-2 rounds typically suffice.

### Clarity Scoring

After each AskUserQuestion round, internally score the idea clarity on 5 dimensions:
1. Target user identified (0/1)
2. Core problem defined (0/1)
3. Success metric defined (0/1)
4. Scope bounded (0/1)
5. Competitive context known (0/1)

Proceed to Phase 2 when score >= 4. After 5 rounds regardless of score, surface a summary and proceed.

### Discovery Question Design

Use `moai-foundation-thinking` deep-questioning framework. Focus on:
- WHO has the problem most acutely? (persona clarity)
- WHAT is the user doing today instead of using your product? (problem depth)
- HOW will you know the product succeeded? (success metric)
- WHAT is explicitly out of scope? (boundary setting)

[HARD] AskUserQuestion Protocol for Discovery:
1. Call `ToolSearch(query: "select:AskUserQuestion")` BEFORE each AskUserQuestion call
2. Maximum 4 questions per AskUserQuestion call
3. Maximum 4 options per question
4. First option must have `(권장)` suffix (Korean) or `(Recommended)` suffix (other languages)
5. NEVER ask questions via free-form prose — all questions via AskUserQuestion only

### Round Example (Phase 1)

```
ToolSearch(query: "select:AskUserQuestion")
AskUserQuestion({
  questions: [
    {
      question: "아이디어를 더 잘 이해하기 위해 몇 가지 확인이 필요합니다.",
      header: "아이디어 명확화 (1/3)",
      options: [
        { label: "주요 대상 사용자: 개인 소비자 (B2C) (권장)", description: "일반 사용자, 소비자, 개인 취미 등" },
        { label: "주요 대상 사용자: 기업/팀 (B2B)", description: "기업 직원, 팀, 비즈니스 사용자" },
        { label: "주요 대상 사용자: 개발자/기술자", description: "소프트웨어 개발자, 엔지니어, 기술 전문가" },
        { label: "주요 대상 사용자: 모두에게 적합함", description: "특정 세그먼트 없이 범용" }
      ]
    }
  ]
})
```

---

## Phase 2: Diverge

**Purpose**: Generate 5-15 divergent angles to prevent premature convergence.
**Foundation reuse**: `moai-domain-ideation` Phase 2 (which delegates to diverge-converge.md)
**Output**: In-memory concept map — NOT written to disk

Execution: Invoke `moai-domain-ideation` Phase 2 Diverge with the clarity-scored idea.

---

## Phase 3: Research

**Purpose**: Validate the idea against existing market, user research, and technical ecosystem.
**Domain skill**: `moai-domain-research` (parallel WebSearch + Context7)
**Output**: `.moai/brain/IDEA-NNN/research.md`

Execution: Invoke `moai-domain-research` Phase 3 execution with:
- The clarity-scored idea
- The top 3 diverged angles from Phase 2 (context for query design)

[HARD] Parallel tool calls: research MUST issue WebSearch and Context7 in a single message (parallel tool call pattern).

---

## Phase 4: Converge

**Purpose**: Reduce diverged angles to single strongest product concept; produce Lean Canvas.
**Foundation reuse**: `moai-domain-ideation` Phase 4 (which delegates to diverge-converge.md)
**Output**: `.moai/brain/IDEA-NNN/ideation.md`

Execution: Invoke `moai-domain-ideation` Phase 4 Converge with:
- Phase 2 diverged concept map
- Phase 3 research.md findings (as grounding evidence for Converge decisions)

The Lean Canvas section in ideation.md MUST have all 9 blocks present.

---

## Phase 5: Critical Evaluation

**Purpose**: Challenge the converged concept with adversarial evaluation and first-principles analysis.
**Foundation reuse**: `moai-foundation-thinking` modules/critical-evaluation.md + modules/first-principles.md
**Output**: Appended "Evaluation Report" section in `.moai/brain/IDEA-NNN/ideation.md`

Execution: Invoke `moai-foundation-thinking` Critical Evaluation on the Lean Canvas from Phase 4.
Then invoke First Principles decomposition on the core solution concept.

Append the combined evaluation to ideation.md (do not create a new file).

---

## Phase 6: Proposal

**Purpose**: Translate converged + evaluated concept into actionable SPEC decomposition candidates.
**Domain skill**: `moai-domain-ideation` Phase 6
**Output**: `.moai/brain/IDEA-NNN/proposal.md`

Execution: Invoke `moai-domain-ideation` Phase 6 Proposal with:
- ideation.md as primary input
- research.md for market context

The proposal.md MUST contain `### SPEC Decomposition Candidates` section with 2-10 entries matching grammar `- SPEC-{DOMAIN}-{NUM}: {scope}`.

[HARD] No tech-stack assumptions: proposal.md solution sections describe capabilities, not implementations (REQ-BRAIN-011).

After proposal.md is written, it becomes the input for downstream workflows:
- `REQ-BRAIN-007`: `/moai project --from-brain IDEA-NNN` reads proposal.md as primary product scope input (Phase A8.1 downstream patch)
- `/moai plan` detects proposal.md and surfaces SPEC Decomposition Candidates via AskUserQuestion suggestion (Phase A8.2 downstream patch)

---

## Phase 7: Handoff Package

<!-- @MX:WARN: [AUTO] Phase 7 prompt template generation -->
<!-- @MX:REASON: Output (prompt.md) is paste-ready into external claude.com Design session. Changes to the template affect user trust in the handoff quality. Any structural changes to the 5-section template must be validated against current claude.com Design prompt best practices before committing. -->

**Purpose**: Produce paste-ready 5-file Claude Design handoff bundle.
**Domain skill**: `moai-domain-design-handoff`
**Output**: `.moai/brain/IDEA-NNN/claude-design-handoff/{prompt,context,references,acceptance,checklist}.md`

### Pre-Phase 7: Brand Check AskUserQuestion

If `brand_present = false`, BEFORE executing Phase 7, offer a brand interview option:

```
ToolSearch(query: "select:AskUserQuestion")
AskUserQuestion({
  questions: [{
    question: "브랜드 컨텍스트가 정의되지 않았습니다. Phase 7(핸드오프) 전에 어떻게 진행하시겠습니까?",
    header: "브랜드 컨텍스트 없음",
    options: [
      { label: "기본 브랜드 보이스로 계속 진행 (권장)", description: "prompt.md에 커스터마이징 안내가 포함된 기본 브랜드 섹션이 생성됩니다. Claude Design 사용 전 직접 편집하실 수 있습니다." },
      { label: "브랜드 인터뷰 먼저 진행", description: "잠시 멈추고 .moai/project/brand/brand-voice.md를 작성합니다. 완료 후 Phase 7을 재개합니다." }
    ]
  }]
})
```

If user chooses brand interview, pause Phase 7, guide brand context creation, then resume.

Execution: Invoke `moai-domain-design-handoff` with:
- ideation.md, proposal.md as inputs
- research.md Sources section for reference URLs
- `brand_present` flag

### Phase 7 Exit: AskUserQuestion (REQ-BRAIN-009)

After all 5 files are written, present next-action options. See `moai-domain-design-handoff` for the exact AskUserQuestion payload.

[HARD] NO auto-execution of `/moai project` (REQ-BRAIN-010). User MUST explicitly select the option to proceed.

---

## Edge Cases

### Mid-Workflow Interrupt Resume

If `.moai/brain/IDEA-NNN/` exists with partial files when `/moai brain` is invoked again:

```
ToolSearch(query: "select:AskUserQuestion")
AskUserQuestion({
  questions: [{
    question: "IDEA-NNN이 이미 존재합니다. 어떻게 진행하시겠습니까?",
    header: "기존 IDEA 발견",
    options: [
      { label: "마지막 완료 단계부터 재개 (권장)", description: "존재하는 파일을 유지하고 누락된 단계부터 계속합니다." },
      { label: "처음부터 다시 시작 (새 IDEA-NNN+1 생성)", description: "현재 IDEA를 유지하고 새 IDEA 번호로 처음부터 시작합니다." }
    ]
  }]
})
```

### Empty SPEC Decomposition

If Phase 6 produces 0 SPEC candidates (idea too small or abstract), the workflow does NOT fail. It produces a proposal.md with a placeholder decomposition section noting the scope is atomic.

---

## Output Summary

After successful completion, confirm all deliverables:

```
.moai/brain/IDEA-NNN/
├── research.md        (Phase 3)
├── ideation.md        (Phase 4 + Phase 5 append)
├── proposal.md        (Phase 6)
└── claude-design-handoff/
    ├── prompt.md      (Phase 7 — paste-ready)
    ├── context.md     (Phase 7)
    ├── references.md  (Phase 7)
    ├── acceptance.md  (Phase 7)
    └── checklist.md   (Phase 7)
```

---

## Works Well With

- `moai-foundation-thinking`: Primary framework library (Deep Questioning, Diverge-Converge, Critical Evaluation, First Principles)
- `moai-domain-ideation`: Lean Canvas assembly, SPEC decomposition (Phases 2, 4, 6)
- `moai-domain-research`: Parallel research execution (Phase 3)
- `moai-domain-design-handoff`: 5-file handoff package (Phase 7)
- `moai-workflow-project`: Downstream consumer via `--from-brain IDEA-NNN` flag
- `moai-workflow-plan`: Downstream consumer parsing SPEC Decomposition Candidates
- `moai-workflow-design-import`: Downstream consumer of `claude-design-handoff/` directory (path A)

---

## Verification

- [ ] 7 phases execute in order (1-Discovery, 2-Diverge, 3-Research, 4-Converge, 5-Critical, 6-Proposal, 7-Handoff)
- [ ] Phase 1 uses AskUserQuestion with ToolSearch preload (no prose questions)
- [ ] Phase 3 issues parallel tool calls (WebSearch + Context7 in single message)
- [ ] Phase 6 proposal.md has `### SPEC Decomposition Candidates` section
- [ ] Phase 7 produces all 5 handoff files
- [ ] Phase 7 exit invokes AskUserQuestion with 3 options (proceed / review / regenerate)
- [ ] No auto-execution of /moai project (user choice only)
- [ ] IDEA-NNN auto-incremented from existing directories
