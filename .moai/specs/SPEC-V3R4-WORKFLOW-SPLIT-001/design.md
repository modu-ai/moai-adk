# Design — SPEC-V3R4-WORKFLOW-SPLIT-001

## Architectural Decisions

본 SPEC은 6개 핵심 architectural decision으로 구성된다. 각 결정은 사용자 AskUserQuestion으로 lock-in되었거나, Workflow Audit Bundle F finding으로부터 직접 유도된다.

---

### AD-001 — Sub-skill Boundary: Phase-Scoped (not Concern-Based, not Threshold-Only)

**Decision**: 분할 단위는 workflow 내 **자연적 phase 경계** (H2/H3 헤더).

**Alternatives 평가**:

| 옵션 | Pros | Cons | 채택 여부 |
|------|------|------|----------|
| **Phase-scoped** (채택) | 자연 경계, 콘텐츠 응집성 유지, agent context loading 최적 | 일부 sub-skill이 borderline LOC | YES |
| **Concern-based** (e.g. error-handling, validation, ui) | 횡단 관심사 분리 가능 | workflow가 phase별 sequential이므로 부자연스러움 | NO |
| **Threshold-only** (e.g. 500 LOC마다 강제 cut) | 단순, 자동화 가능 | phase 중간에 끊겨 가독성 파괴 | NO |

**Rationale**: workflow는 사용자가 `/moai plan` 등을 호출하면 Phase 0 → Phase 1 → ... 순차 진행. agent는 각 phase 진입 시 해당 sub-skill만 on-demand 로드하면 됨. concern-based 분할은 phase 진행 중 여러 sub-skill을 동시 로드해야 하므로 token 효율성 악화.

**User-decided**: AskUserQuestion 결정 (4-round Socratic interview에서 phase-scoped 채택).

---

### AD-002 — Entry Router Pattern

**Decision**: 4개 workflow entry skill (`run.md`, `sync.md`, `project.md`, `plan.md`)을 ≤200 LOC thin router로 축소.

**Router 구조**:

```markdown
---
name: moai-workflow-run
description: "Run phase orchestrator — routes to phase-scoped sub-skills"
user-invocable: true
# ... (기존 metadata 유지)
---

# Run Workflow — Entry Router

## Phase Sequence (Router Table)

| Phase | Sub-skill | Trigger |
|-------|-----------|---------|
| Phase 0 (Context Loading) | `Read workflows/run/context-loading.md` | Always |
| Phase 1-N (Execution) | `Read workflows/run/phase-execution.md` | After Phase 0 |
| Mode Orchestration | `Read workflows/run/mode-orchestration.md` | If --team or --solo flag |

## Invocation Pattern

When MoAI orchestrator handles `/moai run SPEC-XXX`:
1. Load this entry router
2. Phase 0: Read sub-skill `run/context-loading.md`
3. Execute Phase 0 per sub-skill instructions
4. Phase N: Read sub-skill `run/phase-execution.md`
5. ...
```

**Rationale**: Entry router는 (1) Intent Router 매칭 surface 보존, (2) phase 진행 단계 명시, (3) sub-skill 로딩 trigger 명확화. agent는 entry router를 첫 로드한 후, phase 진입 시점에 해당 sub-skill을 `Read`로 로드 → token-efficient on-demand loading.

---

### AD-003 — Sub-skill Frontmatter Discipline

**Decision**: 모든 sub-skill은 다음 frontmatter 규칙을 강제한다.

```yaml
---
name: moai-workflow-{name}-{sub}   # e.g. moai-workflow-run-context-loading
description: "{Concise description of sub-skill purpose}"
user-invocable: false              # CRITICAL — 사용자 직접 호출 차단
metadata:
  parent: moai-workflow-{name}     # e.g. moai-workflow-run
  phase: "{phase identifier}"      # e.g. "Phase 0: Context Loading"
---
```

**Rationale**:
- `user-invocable: false`: sub-skill은 entry router 경유로만 접근. SKILL.md Intent Router가 sub-skill 이름을 매칭하지 않도록 안전 차단
- `metadata.parent`: 추적 가능성 — 어느 entry router에 속하는지 명시
- naming convention `moai-workflow-{name}-{sub}`: name collision 방지, alphabetical sort 시 부모-자식 인접

---

### AD-004 — Cross-Reference Syntax

**Decision**: Sub-skill 간 reference는 다음 우선순위로 사용한다.

| 우선순위 | Syntax | Use case |
|---------|--------|----------|
| 1 (preferred) | `Read workflows/{name}/{sub}.md` (relative) | Entry router 내부, 같은 workflow 내 |
| 2 | `Read .claude/skills/moai/workflows/{name}/{sub}.md` (absolute from project root) | 다른 workflow에서 cross-ref |
| 3 (fallback) | `${CLAUDE_SKILL_DIR}/workflows/{name}/{sub}.md` | 환경 변수 활용 (현재 미사용) |

**Validation**: T0.1 audit script가 모든 `Read .../...` 패턴을 grep 후 실제 파일 존재 검증. 깨진 link 0건이어야 PASS.

**Rationale**: Markdown 표준 link `[text](path)` 대신 명시적 `Read ...` 패턴 채택 — agent가 실제 sub-skill을 로드해야 함을 분명히 신호.

---

### AD-005 — Template Mirror Strategy

**Decision**: Wave별 incremental sync. 각 Wave PR이 local + template 양쪽을 동일 commit에 포함.

**Mirror 구조**:

```
.claude/skills/moai/workflows/         (local — user-facing)
   ├─ run.md                          (entry router, ≤200 LOC)
   ├─ run/                            (sub-skill directory)
   │   ├─ context-loading.md
   │   ├─ phase-execution.md
   │   └─ mode-orchestration.md
   ├─ sync.md, sync/
   ├─ project.md, project/
   └─ plan.md, plan/

internal/template/templates/.claude/skills/moai/workflows/   (template source)
   └─ (1:1 mirror of above)
```

**Build flow**:
1. Edit local + template 양쪽 in same commit
2. `make build` → `internal/template/embedded.go` regenerates (embed FS update)
3. `moai init <new-project>` → 새 프로젝트에 sub-skill 디렉토리 배포

**Rationale**:
- Template-First rule (CLAUDE.local.md §2) 준수: 모든 변경은 template에 먼저 → make build → local 동기화. 단 본 SPEC은 local과 template이 같은 commit에 들어가므로 순서 무관
- Per-Wave atomicity: 각 Wave PR은 4 (Wave 1/2/4) 또는 5 (Wave 3) sub-skill + entry router + template mirror + embedded.go 재생성 = atomic unit
- `moai init` 사용자 프로젝트 회귀 방지

---

### AD-006 — LOC Ceiling Rationale

**Decision**:
- Sub-skill ceiling: **500 LOC**
- Entry router ceiling: **200 LOC**

**Rationale**:

| 지표 | 값 | 근거 |
|------|---|------|
| Progressive disclosure level-2 budget | ~5000 tokens | `.claude/rules/moai/development/skill-authoring.md` § Progressive Disclosure |
| Token-to-LOC ratio (markdown narrative) | ~10 tokens/line | empirical (frontmatter + headers + prose) |
| Level-2 LOC budget | ~500 LOC | 5000 tokens / 10 tokens per line |
| Safety margin | 0% | sub-skill은 phase-scoped self-contained이므로 마진 최소화 가능 |
| Entry router budget | 200 LOC | router는 phase map + invocation pattern만 필요, 짧을수록 좋음 |

**Enforcement**: T0.2 Go test (`internal/skills/workflow_split_test.go`) 가 CI에서 영구 enforce. 미래 PR이 위반하면 자동 block.

**Exception protocol**: 500 LOC 초과가 불가피한 경우 추가 sub-split. exception 허용 안 함 (안 그러면 1년 후 다시 1000 LOC monolithic 회귀).

---

## Cross-Reference Inventory (Plan-Phase 사전 조사)

### Internal references (split 후 sub-skill 간)

각 workflow 내 Phase 간 cross-ref는 source 파일 grep으로 사전 식별:

- **run.md**: "Phase 0 → Phase 1" 표현, "Worktree path rules (see §Context Loading)" — split 후 entry router에서 명시
- **sync.md**: HUMAN GATE 1/2/3 references (모두 quality-gates.md 내부, 안전), "Phase 1 docs scope (Phase 0.5 outcome)" — quality-gates.md → doc-execution.md cross-ref
- **project.md**: Phase 1.5 round-N references (codebase-analysis.md 내부), "Phase 6 meta-harness uses Phase 3 doc-generation output" — codebase-analysis.md → doc-generation.md → meta-harness.md sequential
- **plan.md**: Decision Point 1/2/3.5 references (clarity-interview.md → spec-assembly.md), Annotation Cycle reference (clarity-interview.md 내부)

### External references (다른 파일 → workflow)

Pre-write grep 결과:

```bash
$ grep -rln "workflows/run\|workflows/sync\|workflows/project\|workflows/plan" \
    .claude/ .moai/ docs-site/ --include='*.md' --include='*.yaml' \
    --exclude-dir=node_modules
```

검증 대상:
- `.claude/skills/moai/SKILL.md` Intent Router (entry router 경로 참조 → split 후 변경 불필요, entry skill survives)
- `.claude/commands/*.md` (slash command entry, → entry skill 호출 패턴 무변경)
- `.moai/research/*.md`, MEMORY entries, lessons (text-only mention)
- `docs-site/content/` (사전 grep 결과 0건)

**Action**: Wave 1 진입 직전에 위 grep 재실행하여 신규 reference가 추가되지 않았는지 확인.

---

## Template Synchronization Protocol

### Per-Wave incremental sync

각 Wave PR이 다음 구조를 atomic하게 포함:

```
PR commits structure:
  1. Add sub-skill files (local + template)
  2. Refactor entry router (local + template)
  3. Run `make build` (auto-regenerates embedded.go)
  4. Add/update audit test fixtures
  5. PR body 명시: "Wave N — splits {name}.md into N sub-skills + thin router"
```

**Why per-Wave (not big-bang)**:
- 각 Wave가 독립 PR → review burden 감소
- Wave 1 → Wave 2 사이에 main 머지로 다른 작업과 isolation
- Wave 1 lesson learned (LOC borderline 판정 등)를 Wave 2-4에 반영 가능

### `make build` 자동화 검증

- Wave별 PR에서 `internal/template/embedded.go` diff가 반드시 포함되어야 함
- 누락 시 audit test (T0.2 `TestTemplateMirrorParity`) FAIL → CI block
- `make build`는 deterministic이므로 동일 input에 동일 output (CI에서 재실행하여 diff verify 가능)

---

## docs-site Impact Analysis (AC-WFSP-007 Deliverable)

### Pre-write 검증

```bash
$ grep -rln "workflows/run\|workflows/sync\|workflows/project\|workflows/plan" docs-site/
# Output: (empty — 0 matches)
```

**결론**: docs-site `content/{ko,en,ja,zh}/` 어디에도 `.claude/skills/moai/workflows/*.md` 직접 reference 없음. 따라서 본 SPEC은 §17 docs-site 4-locale sync 의무에서 면제된다.

### Run-phase 재검증 의무

각 Wave PR 종료 직전에 동일 grep 재실행:

- 0건 유지 → 정상, follow-up SPEC 불필요
- ≥1건 발견 → `SPEC-V3R4-WORKFLOW-SPLIT-001-DOCS-FOLLOWUP` 별도 SPEC 생성 (scenarios.md "Out of Scope" 참조)

### Why 0건이 자연스러운가

`.claude/skills/moai/workflows/*.md`는 **agent용 internal skill**. 사용자 문서 (docs-site)는 사용자 관점의 슬래시 커맨드 (`/moai plan`) + 워크플로우 개념을 설명하지, agent skill 파일의 내부 구조를 노출하지 않음. 따라서 reference 0건이 design intent에 부합.

---

## Router Design Pattern (Illustrative Example, Not Implementation)

### Before (current `workflows/run.md` 1073 LOC)

```markdown
---
name: moai-workflow-run
user-invocable: true
---

# Run Workflow

## Purpose
(... 50 LOC ...)

## Scope
(... 30 LOC ...)

## Phase 0: Context Loading
(... 130 LOC ...)

## Phase 1: Execution
(... 700 LOC ...)

## Mode Orchestration
(... 120 LOC ...)

## Test Scenarios
(... 50 LOC ...)
```

### After (entry router ≤200 LOC + 3 sub-skills)

**`workflows/run.md`** (entry router, ~150 LOC):

```markdown
---
name: moai-workflow-run
description: "Run phase orchestrator — routes to phase-scoped sub-skills"
user-invocable: true
---

# Run Workflow — Entry Router

## Purpose (compact)
(... 20 LOC summary ...)

## Phase Sequence

| Phase | Sub-skill | Trigger |
|-------|-----------|---------|
| Phase 0 — Context Loading | `Read workflows/run/context-loading.md` | Always (entry) |
| Phase 1-N — Execution | `Read workflows/run/phase-execution.md` | After Phase 0 complete |
| Mode Orchestration | `Read workflows/run/mode-orchestration.md` | If `--team` or `--solo` flag |

## Invocation Flow

1. MoAI orchestrator receives `/moai run SPEC-XXX`
2. Loads this entry router → understands phase map
3. Phase 0: `Read workflows/run/context-loading.md` → execute its instructions
4. Phase 1: `Read workflows/run/phase-execution.md` → execute
5. Mode: `Read workflows/run/mode-orchestration.md` → conditional execution
6. Completion check → return to orchestrator

## Test Scenarios (router-level)
(... 30 LOC for router invocation correctness only ...)
```

**`workflows/run/context-loading.md`** (~210 LOC, frontmatter user-invocable: false):

```markdown
---
name: moai-workflow-run-context-loading
description: "Run Phase 0 — Context loading, mode dispatch, harness level resolution"
user-invocable: false
metadata:
  parent: moai-workflow-run
  phase: "Phase 0: Context Loading"
---

# Run Phase 0: Context Loading

(... 200 LOC of original lines 1-214 from workflows/run.md, with internal headers promoted H2→H1 ...)
```

**Token efficiency gain**:
- Before: agent loads 1073 LOC = ~10K tokens for entire workflow context
- After: agent loads entry router (~150 LOC = ~1.5K tokens) + on-demand sub-skill (~210 LOC = ~2.1K tokens) = ~3.6K tokens for Phase 0 only
- **Reduction**: ~64% for Phase 0, scales similarly for other phases

---

## Failure Modes & Recovery

### FM1: Sub-skill loading fails (file not found)

- Detection: `Read workflows/run/context-loading.md` returns error
- Recovery: Entry router has fallback — if sub-skill missing, log warning + abort Phase
- Prevention: T0.1 audit script in CI catches broken refs before merge

### FM2: Template/local divergence (Wave PR forgets template mirror)

- Detection: T0.2 `TestTemplateMirrorParity` FAILS in CI
- Recovery: PR blocked, force template mirror commit
- Prevention: per-Wave commit checklist (in plan.md T*.4)

### FM3: SKILL.md Intent Router accidentally modified

- Detection: T0.4 baseline hash 비교 FAILS in CI
- Recovery: revert SKILL.md change
- Prevention: T0.4 hash check is mandatory in every Wave PR

### FM4: `moai init` user project regression

- Detection: post-merge `moai init /tmp/test-X` 실행 → sub-skill 디렉토리 누락 확인
- Recovery: Wave PR revert + template mirror fix + re-merge
- Prevention: Wave별 T*.5 `make build` + embedded.go diff 검증

---

## References

- spec.md § Affected Files (file count: 37)
- plan.md § Wave-by-Wave Tasks
- acceptance.md § Quality Gates (8 AC criteria)
- scenarios.md § Test Plans
- `.moai/research/workflow-audit-2026-05-16.md` Bundle F
- `.claude/rules/moai/development/skill-authoring.md` § Progressive Disclosure
- CLAUDE.local.md §2 Template-First Rule
