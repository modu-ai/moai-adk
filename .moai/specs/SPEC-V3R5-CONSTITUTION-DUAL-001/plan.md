---
id: SPEC-V3R5-CONSTITUTION-DUAL-001
title: "Constitution Dual-Zone Formalization with Validate CLI — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P1
phase: "v3.5.0"
module: ".claude/rules/moai + internal/constitution + internal/cli"
lifecycle: spec-anchored
tags: "constitution, dual-zone, frozen, evolvable, zone-registry, mega-sprint, w1, plan"
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Initial implementation plan — 4 phases (A→B→C→D) sequential |
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Iteration 2 revision — addressed plan-auditor BLOCKING defects (zone-registry 75→72, HARD rules 102→111 empirical, AC methodology unified) + SHOULD defects (5 new ACs for traceability, V3R5-001 namespace, plan.md/acceptance.md Out of Scope sections, 3 sentinel REQs added, AC-CDL-005 split). |

---

## 1. 구현 전략 (Strategy)

### 1.1 단계 분리 원칙

본 SPEC은 **4단계 sequential pipeline** 로 진행한다. 단계 간 의존성:

```
Phase A (annotation)
  ↓ provides ground truth for
Phase B (registry extension)
  ↓ provides registry data for
Phase C (validate CLI)
  ↓ enables enforcement via
Phase D (CI integration)
```

각 phase는 독립 commit으로 분리하여 검증 용이성 + rollback 단순성을 확보한다.

### 1.2 TDD 적용 (Phase C)

`quality.yaml development_mode = tdd` 기본값에 따라 Phase C (Go 코드)는 RED → GREEN → REFACTOR cycle을 적용한다:

- RED: `validator_test.go` 작성, 실패 확인
- GREEN: `validator.go` 최소 구현, 테스트 통과
- REFACTOR: 코드 정리 + coverage 보강 (목표 85%+)

Phase A/B/D는 documentation/YAML 편집이므로 TDD 적용 외 (TRUST 5 Tested pillar는 Phase C에 집중).

### 1.3 Brownfield 안전성

기존 `zone-registry.md` **72 entries** (CONST-V3R2-001..046 + 049 + 051..072 + 150..152; 3 internal gaps at 047/048/050) 및 `internal/cli/constitution.go` `list`/`guard`/`amend` verbs 는 변경하지 않는다 (REQ-CDL-001 Constraint C-CDL-001, C-CDL-006).

**ID 할당 정책 (iteration 2 revision)**: 신규 entries 는 `CONST-V3R5-001` 부터 zero-padded 3-digit 으로 부여하며, V3R2 시퀀스와 **parallel namespace** 로 운영한다. V3R2 의 3 internal gaps (047/048/050) 은 fill 하지 않고 역사적 기록으로 보존한다.

---

## 2. Phase A — Annotation (D1: ZONE 마커 부착)

### 2.1 목표

15 헌법 source files (spec.md §2.2 enumerated) 의 모든 `[HARD]` 규칙에 `[ZONE:Frozen]` 또는 `[ZONE:Evolvable]` 마커를 부착하여 zone classification을 inline으로 명시한다. Empirical baseline N=111 [HARD] occurrences at main HEAD `3bd2aa291`.

### 2.2 Task A-1: HARD 규칙 inventory 수집

- 15 source files (spec.md §2.2) 에 대해 `grep -nE '\[HARD\]' <files>` 실행하여 모든 [HARD] 규칙의 (file, line, text) tuple 수집
- 출력: `inventory-phase-a.md` (임시 작업 파일, 결과는 commit하지 않음)
- **Empirical ground truth** (verified at main HEAD `3bd2aa291`):

| File | HARD count |
|------|-----------|
| `CLAUDE.md` | 14 |
| `.claude/rules/moai/core/agent-common-protocol.md` | 11 |
| `.claude/rules/moai/core/askuser-protocol.md` | 3 |
| `.claude/rules/moai/core/moai-constitution.md` | 11 |
| `.claude/rules/moai/design/constitution.md` | 19 |
| `.claude/rules/moai/workflow/ci-autofix-protocol.md` | 10 |
| `.claude/rules/moai/workflow/ci-watch-protocol.md` | 8 |
| `.claude/rules/moai/workflow/context-window-management.md` | 5 |
| `.claude/rules/moai/workflow/session-handoff.md` | 5 |
| `.claude/rules/moai/workflow/spec-workflow.md` | 3 |
| `.claude/rules/moai/workflow/worktree-integration.md` | 12 |
| `.claude/rules/moai/workflow/worktree-state-guard.md` | 1 |
| `.claude/rules/moai/development/agent-authoring.md` | 1 |
| `.claude/rules/moai/development/branch-origin-protocol.md` | 7 |
| `.claude/rules/moai/development/skill-authoring.md` | 1 |
| **TOTAL** | **111** |

### 2.3 Task A-2: Classification 결정 (rule-by-rule)

각 규칙에 대해 다음 휴리스틱으로 zone 결정:

| 규칙 카테고리 | Zone | 근거 |
|---------------|------|------|
| AskUserQuestion 채널 독점 (CONST-V3R2-012/025/038) | Frozen | 사용자-에이전트 안전 경계, 변경 시 silent failure |
| AAQ 사전 preload (deferred tool 강제) | Frozen | InputValidationError 방지 메커니즘 |
| 보안 (Background agent write 금지, secrets 금지) | Frozen | OWASP 준수, 변경 시 보안 침해 |
| 헌법 자체 (constitution files) | Frozen | meta-rule, 자기 참조 무한루프 방지 |
| TRUST 5 framework | Frozen | 품질 게이트 SSOT |
| CI watch + autofix protocols (verbatim merge gates) | Frozen | CI 인프라 정합성, 변경 시 silent merge incidents |
| Worktree-integration (L1/L2/L3 layer definitions) | Frozen | 사용자 cwd 안전 (worktree-state-guard 의존) |
| Branch-Origin Protocol (BODP) | Frozen | 신규 SPEC 안전 분기 가정 (3-signal matrix) |
| Opus 4.7 prompt philosophy (Principle 4/5) | Evolvable | 모델 버전 의존, 향후 변경 가능 |
| Agent Core Behaviors (6 behaviors) | Evolvable | 휴리스틱 튜닝 가능 |
| Worktree isolation 규칙 (L1 use vs not) | Evolvable | 사용자 정책 (2026-05-17 user policy) 기반 |
| Markdown output, language handling | Evolvable | 운영 정책, 향후 i18n 확장 가능 |
| Context window thresholds (1M=75%, 200K=90%) | Evolvable | 모델 의존, 미래 모델 도입 시 갱신 |
| Session-handoff 6-block format | Evolvable | 운영 컨벤션, 효과적 시 갱신 가능 |

기존 `design/constitution.md` §2 의 FROZEN list와 정합성 검증:

- [FROZEN] constitution file 자체 → Frozen
- [FROZEN] §3.1/§3.2/§3.3 → Frozen
- [FROZEN] Safety architecture (Section 5) → Frozen
- [FROZEN] GAN Loop contract (Section 11) → Frozen
- [FROZEN] Evaluator leniency (Section 12) → Frozen
- [FROZEN] Pipeline phase ordering → Frozen
- [FROZEN] Pass threshold floor → Frozen
- [FROZEN] Human approval requirement → Frozen
- [EVOLVABLE] skill body content → Evolvable
- [EVOLVABLE] Pipeline adaptation weights → Evolvable
- [EVOLVABLE] Evaluation rubric criteria → Evolvable
- [EVOLVABLE] Design tokens, brand heuristics → Evolvable
- [EVOLVABLE] Iteration limits → Evolvable

### 2.4 Task A-3: 마커 부착 (Edit, file-by-file)

순서 (file by file):

1. **`CLAUDE.md`** — §1 Hard Rules (8개), §7 Safe Development Protocol (5개), §8 User Interaction Architecture, §14 Parallel Execution Safeguards (worktree 4개), §19 AskUserQuestion Enforcement, 기타 §11/§16
2. **`.claude/rules/moai/core/moai-constitution.md`** — MoAI Orchestrator (3개 HARD), Opus 4.7 (Principles 4-5), Agent Core Behaviors (6개)
3. **`.claude/rules/moai/core/agent-common-protocol.md`** — User Interaction Boundary (3개), Background Agent Execution (1개), Tool Usage Guidelines (1개), Language Handling, Output Format
4. **`.claude/rules/moai/core/askuser-protocol.md`** — Channel Monopoly, ToolSearch Preload, Free-form Circumvention Prohibition (3 HARD)
5. **`.claude/rules/moai/design/constitution.md`** — 기존 [FROZEN] / [HARD] 규칙에 inline `[ZONE:Frozen]` 추가, 정합성 검증
6. **`.claude/rules/moai/workflow/spec-workflow.md`** — SPEC Phase Discipline Step 1~4 (3 HARD)
7. **`.claude/rules/moai/workflow/worktree-integration.md`** — 12 HARD (terminology glossary + decision tree)
8. **`.claude/rules/moai/workflow/worktree-state-guard.md`** — AskUserQuestion HARD (1)
9. **`.claude/rules/moai/workflow/ci-autofix-protocol.md`** — 10 HARD
10. **`.claude/rules/moai/workflow/ci-watch-protocol.md`** — 8 HARD
11. **`.claude/rules/moai/workflow/context-window-management.md`** — 5 HARD (thresholds + handoff)
12. **`.claude/rules/moai/workflow/session-handoff.md`** — 5 HARD (5 triggers + 6-block format)
13. **`.claude/rules/moai/development/branch-origin-protocol.md`** — 7 HARD (3-signal + matrix)
14. **`.claude/rules/moai/development/agent-authoring.md`** — 1 HARD
15. **`.claude/rules/moai/development/skill-authoring.md`** — 1 HARD

형식 예시:

```markdown
- [ZONE:Frozen] [HARD] All user-facing questions MUST go through AskUserQuestion ...
- [ZONE:Evolvable] [HARD] Background subagents (run_in_background: true) MUST NOT perform Write/Edit ...
```

### 2.5 Task A-4: 정합성 검증

- `grep -hcE '\[HARD\]' <15 files>` 합산 결과 X (target X=111)
- `grep -hcE '\[ZONE:(Frozen|Evolvable)\]' <15 files>` 합산 결과 Y
- Y >= X 확인 (AC-CDL-001)
- 중복 마커 확인 (`grep -cE '\[ZONE:.*\].*\[ZONE:.*\]'` 결과 0이어야 함, 또는 REQ-CDL-018 DUPLICATE_ZONE_MARKER warning 으로 보고)

### 2.6 Phase A 산출물

- Commit 1: `docs(constitution): annotate ZONE markers on 15 canonical files [Phase A]`
- 변경 파일: 15 (.md only, no Go code)
- LOC 변경: ~230 insertions (each marker ~12 chars, 111 rules)

---

## 3. Phase B — Registry Extension (D2: 100% coverage)

### 3.1 목표

`.claude/rules/moai/core/zone-registry.md` 의 entry 수를 **72 → ≥ 111** 로 확장하여 모든 [HARD] 규칙을 매핑. 신규 `zone_class` 필드 도입. **39 new entries** 추가 (111 − 72 = 39).

### 3.2 Task B-1: 미매핑 [HARD] 규칙 식별

Phase A의 inventory와 현재 registry entry의 차집합 계산:

- **현재 registry**: 72 entries — CONST-V3R2-001..046 (46) + 049 (1) + 051..072 (22) + 150..152 (3) = 72 confirmed
- **V3R2 sequence gaps**: 3 internal gaps at IDs 047, 048, 050 (역사적 미할당 또는 제거; W1 은 fill 하지 않음 — namespace separation 정책에 따라 V3R5 신규 namespace 사용)
- **Phase A inventory**: 111 (empirical at main HEAD `3bd2aa291`)
- **미매핑 (target for Phase B)**: 39 entries

미매핑 후보 (architecture-audit F-006 PENDING set 갱신):

- CLAUDE.md §19 AskUserQuestion Enforcement Protocol (§19.2~§19.6 신규)
- CLAUDE.md §16 Context Search Protocol HARD 규칙
- CLAUDE.md §17 Troubleshooting 일부
- moai-constitution.md Agent Core Behaviors (Behaviors 1-6 중 zone-registry 미매핑 분)
- agent-common-protocol.md Tool Usage Guidelines, Time Estimation HARD
- askuser-protocol.md Channel Monopoly + Preload + Circumvention (3 HARD 신규)
- design/constitution.md 일부 §4 Pipeline Architecture 규칙
- workflow/ci-autofix-protocol.md 전체 (10 HARD)
- workflow/ci-watch-protocol.md 전체 (8 HARD)
- workflow/context-window-management.md (5 HARD)
- workflow/session-handoff.md (5 HARD)
- workflow/worktree-integration.md 일부
- development/branch-origin-protocol.md (7 HARD)
- development/agent-authoring.md (1 HARD)
- development/skill-authoring.md (1 HARD)

### 3.3 Task B-2: zone_class 4-classification 부여

각 entry에 `zone_class` 필드 추가:

| zone_class | 정의 | 예시 |
|------------|------|------|
| `frozen-canonical` | 메커니즘 그 자체 (없으면 시스템 작동 불가) | AskUserQuestion 채널 독점, deferred preload |
| `frozen-safety` | 안전 보장 (제거 시 보안/정합성 침해) | Background agent write 금지, secrets 금지 |
| `evolvable-tuning` | 운영 튜닝 (학습자 안전 변경 가능) | effort level 선택, agent behavior weights |
| `evolvable-experimental` | 실험적 (불안정, 사용자 opt-in) | L3 worktree 모드, CG mode |

기존 72 entries에 대해서도 zone_class를 retroactive하게 부여 (AC-CDL-006 에 따라 모든 entries 가 4-enum 중 하나를 가져야 함; canary_gate 값을 참조: `canary_gate: true && zone: Frozen` → `frozen-canonical` 또는 `frozen-safety` 중 case-by-case 판단).

### 3.4 Task B-3: 신규 entry 작성 (CONST-V3R5-001 ~ )

**ID Namespace Policy (iteration 2 revision)**:

신규 entries 는 **CONST-V3R5-001** 부터 시작하는 parallel namespace 를 사용한다. V3R2 와 V3R5 는 별개 namespace 로 운영되며 V3R2 의 3 internal gaps (047/048/050) 은 fill 하지 않는다.

YAML 형식 (기존과 동일 + `zone_class` 추가):

```yaml
- id: CONST-V3R5-001
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#19-askuserquestion-enforcement-protocol"
  clause: "[HARD] 모든 MoAI 세션은 첫 사용자 입력 수신 직후 ToolSearch 호출을 실행해야 한다"
  canary_gate: true

- id: CONST-V3R5-002
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#19-askuserquestion-enforcement-protocol"
  clause: "[HARD] 모든 사용자 응답 전송 전, 4항목 자가 점검 (question mark, option list, schema load, silent wait)"
  canary_gate: true
```

목표: CONST-V3R5-001..039 (또는 그 이상) 까지 신규 39 entries 추가하여 총 entry 수 ≥ 111 달성.

### 3.5 Task B-4: zone-registry.md HISTORY 갱신

기존 HISTORY 테이블에 새 행 추가:

```markdown
| 1.1.0 | 2026-05-20 | SPEC-V3R5-CONSTITUTION-DUAL-001 | F-006 coverage gap 해소 — CONST-V3R5-001..039 추가 (parallel namespace), zone_class 4-classification 도입 (retroactive on all 111 entries) |
```

zone-registry.md "ID Allocation Policy" section 도 갱신하여 V3R5 namespace 정책 명시.

### 3.6 Task B-5: list CLI compatibility 검증

`moai constitution list --format json` 실행 후:

- 신규 entry 정상 출력
- 신규 `zone_class` 필드가 JSON 에 포함
- `--zone frozen` / `--zone evolvable` 필터 정상 작동
- 기존 72 entries 동작 변경 없음 (Backward compatibility)

### 3.7 Phase B 산출물

- Commit 2: `docs(constitution): extend zone-registry to 100% coverage with zone_class (parallel V3R5 namespace) [Phase B]`
- 변경 파일: 1 (zone-registry.md)
- LOC 변경: ~350 insertions (39 entries × ~9 LOC each + zone_class on existing 72 entries)

---

## 4. Phase C — Validate CLI (D3: Go 구현, TDD)

### 4.1 목표

`moai constitution validate` CLI verb 신설. drift 탐지, missing source file 탐지, unregistered HARD rule 탐지를 지원. 기존 `guard`/`list`/`amend` verbs 와 직교 (C-CDL-006).

### 4.2 Task C-1: 디렉토리 구조 및 skeleton

신규 파일:

- `internal/constitution/registry.go` — YAML registry parser (struct definitions, Load function)
- `internal/constitution/validator.go` — drift detection logic
- `internal/constitution/validator_test.go` — unit tests
- `internal/cli/constitution_validate.go` — Cobra subcommand wiring

기존 파일 수정:

- `internal/cli/constitution.go` — `validate` subcommand 등록 (기존 `guard`/`list`/`amend` 와 sibling 으로 추가)

### 4.3 Task C-2: TDD Cycle 1 — Happy path (REQ-CDL-003 + AC-CDL-003)

**RED**:

```go
// validator_test.go
func TestValidate_HappyPath(t *testing.T) {
    // Given: registry in sync with sources
    // When: Validate() called
    // Then: status=ok, drift_count=0
}
```

실패 확인 → validator.go 미존재.

**GREEN**:

- `Registry` struct + `Entry` struct 정의
- `LoadRegistry(path string) (*Registry, error)` 구현
- `Validate(reg *Registry) (*ValidationResult, error)` 구현 (happy path만)
- `ValidationResult` struct: `Status`, `DriftCount`, `MissingCount`, `UnregisteredCount`, `Entries []EntryResult`
- 테스트 통과 확인

**REFACTOR**:

- file read caching (REQ-CDL-003 5-second target)
- error wrapping 패턴 (`fmt.Errorf("operation: %w", err)`)

### 4.4 Task C-3: TDD Cycle 2 — Drift detection (REQ-CDL-006 + AC-CDL-004)

**RED**:

```go
func TestValidate_DriftDetection(t *testing.T) {
    // Given: registry entry's clause no longer matches source
    // When: Validate() called
    // Then: entry has status=DRIFT, exit code 1
}
```

**GREEN**:

- `checkDrift(entry Entry, sourceContent string) EntryResult` 구현
- substring matching (registry clause 가 source 에 존재하는지)
- anchor-aware: anchor section 내에서만 검색 (false positive 방지)

**REFACTOR**:

- diff summary 생성 (human-readable output)
- sentinel error key 분리 (`ErrDrift`, `ErrSourceMissing`, etc.)

### 4.5 Task C-4: TDD Cycle 3 — Missing source file (REQ-CDL-013)

**RED**: source file 삭제된 상황에서 `status: SOURCE_FILE_MISSING` 확인.

**GREEN**: os.Stat 으로 file existence 확인 → 없으면 `SOURCE_FILE_MISSING` 마킹, exit code 2.

### 4.6 Task C-5: TDD Cycle 4 — Unregistered HARD rule (REQ-CDL-008)

**RED**: 15 헌법 source files 에 [HARD] 규칙이 있는데 registry 에 매핑되지 않은 경우 `ZONE_UNREGISTERED` 확인.

**GREEN**:

- `checkUnregistered(files []string, registry *Registry) []EntryResult` 구현
- 각 source file 에서 `[HARD]` 규칙 추출 → registry entry 의 clause 와 매칭 시도
- 미매칭 시 unregistered 로 보고

### 4.7 Task C-6: TDD Cycle 5 — FROZEN_WITHOUT_CANARY + ANCHOR_NOT_FOUND + DUPLICATE_ID + INVALID_ZONE_CLASS (REQ-CDL-015, 016, 017, AC-CDL-006)

**RED**:

- `zone: Frozen` + `canary_gate: false` entry → reject (REQ-CDL-015)
- anchor가 source file 에 존재하지 않는 entry → reject (REQ-CDL-016)
- 두 entries 가 동일 `id` → reject (REQ-CDL-017, exit 1 always)
- `zone_class` 가 4-enum 외 → reject `INVALID_ZONE_CLASS` (AC-CDL-006)

**GREEN**:

- `checkEntryIntegrity(entry Entry, sourceContent string) []EntryResult` 구현
- anchor heading 존재 확인 (markdown heading 또는 section marker)
- id duplicate detection at parse time
- zone_class enum validation

### 4.8 Task C-7: TDD Cycle 6 — DUPLICATE_ZONE_MARKER + STALE_ENTRY warnings (REQ-CDL-018, 019)

**RED**: 단일 [HARD] 행에 2개 [ZONE:*] 마커 → warning, 90일 이상 오래된 entry → warning.

**GREEN**: warning emission, exit code unaffected unless `--strict --fail-on-warning`.

### 4.9 Task C-8: TDD Cycle 7 — Live reload + read-only assertion (AC-CDL-008, AC-CDL-010)

**RED**:

- Modify zone-registry.md → re-invoke list → updated entries reflected (TestList_ReflectsUpdatesWithoutRestart)
- Read-only filesystem → validator runs without write attempt (TestValidate_ReadOnlyAssertion)

**GREEN**:

- Registry loading per-invocation (no in-memory cache between invocations)
- Validator code review: no `os.WriteFile`, `os.Create`, etc. calls

### 4.10 Task C-9: CLI wiring (`constitution_validate.go`)

```go
// Pseudo-code outline (not literal implementation)
// newConstitutionValidateCmd() *cobra.Command
//   Flags: --strict (bool), --format (string), --allow-warnings (bool), --fail-on-warning (bool)
//   Run: load registry → call Validate() → format output → exit code
```

Flags:

- `--strict` : fail on any drift/missing/unregistered (default false)
- `--format <text|json>` : output format (default text)
- `--allow-warnings` : exit 0 even on warnings (overrides --strict for warnings)
- `--fail-on-warning` : promote warnings to errors (combined with --strict)

환경 변수 처리 (REQ-CDL-010, REQ-CDL-011):

- `CI=true` → `--strict` 강제
- `MOAI_CONSTITUTION_SKIP_VALIDATE=1` → warning + exit 0

### 4.11 Task C-10: JSON Schema 정의 (REQ-CDL-012)

```json
{
  "status": "ok|drift|missing",
  "drift_count": 0,
  "missing_count": 0,
  "unregistered_count": 0,
  "entries": [
    {
      "id": "CONST-V3R2-001",
      "file": ".claude/rules/moai/workflow/spec-workflow.md",
      "anchor": "#phase-overview",
      "status": "OK|DRIFT|SOURCE_FILE_MISSING|ZONE_UNREGISTERED|FROZEN_WITHOUT_CANARY|ANCHOR_NOT_FOUND|DUPLICATE_ID|DUPLICATE_ZONE_MARKER|STALE_ENTRY|INVALID_ZONE_CLASS",
      "detail": "(optional human-readable detail)"
    }
  ]
}
```

jq compatibility 검증:

```bash
moai constitution validate --format json | jq '.entries[] | select(.status == "DRIFT")'
```

### 4.12 Task C-11: Coverage 검증

`go test -cover ./internal/constitution/` 결과 ≥ 85% 확인.

미달 시 fixture 추가하여 edge case coverage 보강.

### 4.13 Phase C 산출물

- Commit 3-9: TDD cycle 별로 분리 commit (RED/GREEN/REFACTOR 묶음 cycle별 1 commit)
- 또는 Commit 3 통합: `feat(constitution): add validate CLI with drift detection [Phase C, TDD]`
- 변경 파일: 4 new + 1 modified
- LOC 변경: ~350 LOC new code + ~300 LOC test = ~650 total

---

## 5. Phase D — CI Integration (검증 강제)

### 5.1 목표

`.github/workflows/ci.yml` 에 `moai constitution validate --strict` 단계 추가. `main` 브랜치 보호 규칙의 required check 로 등록.

### 5.2 Task D-1: ci.yml step 추가 (AC-CDL-005a)

Tier 1 fast-track 정책 (CLAUDE.local.md §18.7) 준수:

- Ubuntu only (macOS/Windows는 Tier 2 release PR 로 이관)
- Lint job 이후 실행 (Test job 이전)

YAML 예시 (~20 LOC):

```yaml
  constitution-validate:
    name: Constitution Validate
    runs-on: ubuntu-latest
    needs: [lint]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - name: Build moai binary
        run: make build
      - name: Validate constitution
        run: ./moai constitution validate --strict --format json
        env:
          CI: 'true'
          MOAI_CONSTITUTION_SKIP_VALIDATE: ''  # explicitly clear to prevent CI bypass
```

### 5.3 Task D-2: Branch protection 업데이트 (AC-CDL-005b — manual maintainer verification)

CLAUDE.local.md §18.7 의 4 required checks (`Lint`, `Test (ubuntu-latest)`, `Build (linux/amd64)`, `CodeQL`) 에 `Constitution Validate` 추가하여 **5 required checks** 로 확장.

```bash
gh api -X PUT /repos/modu-ai/moai-adk/branches/main/protection \
  --input - <<'EOF'
{
  "required_status_checks": {
    "strict": true,
    "contexts": ["Lint", "Test (ubuntu-latest)", "Build (linux/amd64)", "CodeQL", "Constitution Validate"]
  },
  ...
}
EOF
```

본 작업은 admin 권한 필요 → SPEC 머지 후 메인테이너가 수동 적용. CLAUDE.local.md §18.7 의 baseline 도 4→5 로 업데이트. **AC-CDL-005b 는 manual verification 으로 분류** (CI 자동 검증 불가).

### 5.4 Task D-3: 로컬 개발 검증 경로

`lefthook.yml` `pre-push` hook (CLAUDE.local.md §18.7) 에 validate 추가 검토. 단, 5초 budget 내에서 수행 가능해야 함 (현재 `make preflight` 1-2분).

옵션 A (권장): `make preflight` 에 validate 단계 추가 → 로컬 push 전 검증.
옵션 B: `lefthook.yml` 신규 step 으로 추가 → 더 명시적, 사용자 친화.

본 SPEC 에서는 옵션 A 채택. `Makefile` `preflight` target 에 `./moai constitution validate --strict` 한 줄 추가.

### 5.5 Phase D 산출물

- Commit 10: `ci(constitution): require validate step in main branch protection [Phase D]`
- 변경 파일: 2 (`.github/workflows/ci.yml`, `Makefile`)
- LOC 변경: ~30 lines (YAML + Makefile target update)

---

## 6. Milestone 분해 (Priority 기반, no time estimates)

| Milestone | 우선순위 | 의존성 | 산출물 |
|-----------|---------|--------|--------|
| M1 — Phase A 완료 | P1 | 없음 | 15 .md 파일 ZONE 마커 부착, AC-CDL-001 PASS |
| M2 — Phase B 완료 | P1 | M1 (ground truth) | zone-registry.md ≥ 111 entries (CONST-V3R5-001..039 추가), zone_class on all entries, AC-CDL-002/006/007 PASS |
| M3 — Phase C 완료 (TDD) | P1 | M2 (registry data) | validator CLI 동작, AC-CDL-003/004/008/009/010 PASS, coverage ≥ 85% |
| M4 — Phase D 완료 (CI) | P1 | M3 (CLI 존재) | CI step 동작, AC-CDL-005a PASS (CI auto); AC-CDL-005b PASS (manual maintainer admin merge 후) |

진행 순서: M1 → M2 → M3 → M4 sequential.

---

## 7. 위험 분석 (Risk Matrix)

| Risk ID | 설명 | 영향 | 가능성 | 완화 |
|---------|------|------|--------|------|
| R-CDL-01 | Phase A에서 [HARD] 규칙의 zone classification 판단이 주관적이라 검토자 의견 분기 | Medium | High | plan-auditor가 zone classification heuristic table (§2.3) 을 명시적 PASS 기준으로 검증. 의견 분기 시 보수적으로 Frozen 선택 (안전 우선) |
| R-CDL-02 | zone-registry.md entry 수 추정 (111)이 실제 audit 결과와 다름 | Low | Medium | iteration 2 에서 main HEAD `3bd2aa291` 에 대해 15 source files grep sum = 111 verified. 본 SPEC 은 "≥ 111 entries" 로 lower bound 명시. 미래 source files 추가 시 §2.2 갱신 필요 |
| R-CDL-03 | validator의 substring matching이 false positive (긴 [HARD] 규칙이 부분 match) | Medium | Medium | anchor-aware matching 으로 false positive 감소. 추가로 minimum match length (≥ 30 chars) heuristic 도입 |
| R-CDL-04 | CI step 추가 후 기존 PR 모두 실패 (registry stale) | High | Low | Phase B 완료 직후 main merge 하여 baseline 확보. CI step 도입 (Phase D) 은 Phase B 머지 후 follow-up commit 으로 분리 가능 |
| R-CDL-05 | Branch protection 5→6 required checks 추가가 메인테이너 admin 권한 의존 | Medium | Low | Task D-2 는 메인테이너 수동 작업으로 명시. AC-CDL-005b (manual verification) 로 분리. SPEC sync 단계에서 사용자에게 별도 안내 |
| R-CDL-06 | design/constitution.md 의 기존 §2 zone list 와 inline marker 정합성 충돌 | Low | Medium | Phase A Task A-3 step 5 에서 명시적 정합성 검증 단계 포함 |
| R-CDL-07 | `MOAI_CONSTITUTION_SKIP_VALIDATE=1` override 가 CI 에서 오용 | Low | Low | CI workflow YAML 에서 해당 env 차단 (`env: { MOAI_CONSTITUTION_SKIP_VALIDATE: '' }`). docs 에 "development-only" 강조 |
| R-CDL-08 | V3R5 namespace 와 V3R2 namespace 의 ID 충돌 (예: 두 시리즈 모두 001 시작) | Low | Low | namespace prefix (`CONST-V3R2-` vs `CONST-V3R5-`) 가 다르므로 ID 글로벌 unique. AC-CDL-006 + REQ-CDL-017 (DUPLICATE_ID) 가 충돌 시 reject |

---

## 8. 기술 접근 (Technical Approach)

### 8.1 Go package layout

```
internal/constitution/
├── registry.go          # YAML parser, Registry/Entry struct
├── registry_test.go     # parser unit tests
├── validator.go         # drift detection logic
├── validator_test.go    # validator unit tests
└── testdata/
    ├── valid-registry.yaml
    ├── drift-fixture.md
    └── missing-source.yaml

internal/cli/
├── constitution.go              # (modify) register validate subcommand
└── constitution_validate.go     # (new) Cobra command wiring
```

### 8.2 Key struct definitions

(Pseudo-code for design clarity; literal implementation in Phase C)

`Entry` struct: id, zone (enum), zone_class (enum), file, anchor, clause, canary_gate fields parsed from YAML registry.

`Registry` struct: collection of Entry.

`ValidationResult` struct: status (enum: ok/drift/missing), counts (drift/missing/unregistered), entries (slice of EntryResult).

`EntryResult` struct: id, status (enum: OK/DRIFT/SOURCE_FILE_MISSING/ZONE_UNREGISTERED/FROZEN_WITHOUT_CANARY/ANCHOR_NOT_FOUND/DUPLICATE_ID/DUPLICATE_ZONE_MARKER/STALE_ENTRY/INVALID_ZONE_CLASS), detail.

### 8.3 Anchor-aware substring matching

source file 을 anchor heading 별로 split → registry entry 의 anchor 에 해당하는 section 내에서만 clause substring 검색. false positive 방지.

알고리즘:

1. source file을 markdown heading (`^#+\s+`) 기준으로 split
2. anchor (e.g., `#user-interaction-boundary`) 를 heading slug 와 매칭
3. 해당 section 내에서 clause substring 검색
4. found → OK, not found → DRIFT

### 8.4 Performance optimization

- File read caching: 같은 source file 을 여러 entry 가 참조 → `map[string]string` cache
- Lazy loading: registry 전체 load 후 file open 은 실제 검증 시점에 1회

목표 < 5초 (REQ-CDL-003 / AC-CDL-003 기준).

---

## 9. 검증 전략 (Verification Strategy)

### 9.1 Acceptance Criteria 매핑

| AC | Phase | 검증 방법 |
|-----|-------|-----------|
| AC-CDL-001 | A | `grep -hcE` 합산 비교 (X vs Y across 15 files) |
| AC-CDL-002 | B | `moai constitution list --format json \| jq` 카운트 (N≥111) |
| AC-CDL-003 | C | `validator_test.go TestValidate_HappyPath` |
| AC-CDL-004 | C | `validator_test.go TestValidate_DriftDetection` 외 |
| AC-CDL-005a | D | CI workflow run (PR 시뮬레이션) |
| AC-CDL-005b | D | `gh api .../branches/main/protection` (manual maintainer 검증) |
| AC-CDL-006 | B | zone_class enum validation script |
| AC-CDL-007 | B | grep regex `^CONST-V3R5-[0-9]{3}$` per new entry |
| AC-CDL-008 | C | `TestList_ReflectsUpdatesWithoutRestart` |
| AC-CDL-009 | C | `TestValidate_SkipEnvOverride` |
| AC-CDL-010 | C | `TestValidate_ReadOnlyAssertion` |

### 9.2 TRUST 5 quality gate

- Tested: `internal/constitution/` coverage ≥ 85% (TDD cycle 결과)
- Readable: `golangci-lint run` zero warnings
- Unified: `gofmt -s -w` 적용
- Secured: validator 는 read-only (REQ-CDL-014), 외부 입력 없음
- Trackable: commit message 는 Conventional Commits, 각 Phase 별 commit 분리

### 9.3 plan-auditor 검증 포인트

본 plan 은 plan-auditor 의 다음 dimension 을 통과해야 함:

- D1 (Brief Quality): EARS 형식 19 reqs, 10 ACs, exclusions 명시, REQ↔AC traceability 100%
- D2 (Phase Decomposition): 4 phases sequential, task atomic
- D3 (Risk Management): 8 risks matrix
- D4 (Frontmatter Compliance): 12-field canonical schema 준수
- D5 (Exclusion Discipline): 6 exclusions 명시 (EXCL-001~006)
- D6 (Lint Baseline): 본 SPEC 자체는 lint warnings 0개 목표

---

## 10. Scope Boundaries

### 10.1 Out of Scope

See `spec.md` §5.2 for the canonical exclusion list (EXCL-001 through EXCL-006). Brief summary for plan.md siblings:

- **EXCL-001**: PreToolUse Frozen Guard hook implementation deferred to W3 HARNESS-AUTONOMY-001
- **EXCL-002**: agent/skill frontmatter `zone:` field deferred to T3 Full / follow-up SPEC-V3R5-AGENT-ZONE-001
- **EXCL-003**: expert-backend / expert-frontend / expert-mobile retirement deferred to W2 CORE-SLIM-001
- **EXCL-004**: V3R3/V3R4 workflow rules beyond §2.2 enumerated 15 files — retroactive classification deferred
- **EXCL-005**: design/constitution.md structural change forbidden — inline `[ZONE:*]` markers only
- **EXCL-006**: precommit/generation-time check deferred — CI-time validation only

본 plan.md scope 외부의 implementation decisions (예: validator's substring matching algorithm refinement, future i18n of error messages) 은 후속 SPECs (TBD) 로 이관한다.

---

## 11. 후속 SPEC 연결 (Dependencies)

- **Unblocks**:
  - SPEC-V3R5-CORE-SLIM-001 (W2): canonical 17 agent FROZEN 명시 → expert-backend/frontend retirement 헌법적 근거 제공
  - SPEC-V3R5-HARNESS-AUTONOMY-001 (W3): PreToolUse Frozen Guard hook 이 본 SPEC 의 zone-registry 를 참조 데이터로 사용

- **Depends on**:
  - SPEC-V3R5-CLAUDE-REFRESH-001 (W0, COMPLETE): CLAUDE.md template baseline v14.0.0 확정
  - SPEC-V3R5-LINT-CLEAN-001 (LCLN, COMPLETE): lint baseline 안정화 (본 SPEC plan-auditor PASS 기준)

- **Follow-up SPECs (v3.5.0 후)**:
  - SPEC-V3R5-AGENT-ZONE-001 (가칭, T3 Full): agent/skill frontmatter `zone:` 필드 추가
  - SPEC-V3R5-WORKFLOW-ZONE-001 (가칭): §2.2 외 추가 workflow rule retroactive classification

---

## 12. 산출물 요약 (Summary of Deliverables)

| Phase | Commit 수 | 변경 파일 | LOC 변경 (예상) |
|-------|-----------|-----------|-----------------|
| A | 1 | 15 (.md) | +230 (markers) |
| B | 1 | 1 (zone-registry.md) | +350 (39 new entries + zone_class on existing 72) |
| C | 7 (TDD cycles) 또는 1 통합 | 5 (4 new + 1 modified) | +650 (code + tests) |
| D | 1 | 2 (ci.yml + Makefile) | +30 (YAML + target) |
| **총합** | **3-10 commits** | **23 files** | **~1260 LOC** |
