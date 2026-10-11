---
id: SPEC-WORKFLOW-RULE-SPLIT-001
title: "spec-workflow.md 40k 예산 초과 수리 — core + detail companion 분할"
version: "0.1.0"
status: completed
created: 2026-10-08
updated: 2026-10-08
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".claude/rules/moai/workflow/spec-workflow.md, .claude/rules/moai/workflow/spec-workflow-detail.md, internal/template/templates/.claude/rules/moai/workflow/, internal/template/rule_template_mirror_test.go, internal/template/internal_content_leak_test.go"
lifecycle: spec-anchored
tags: "rule-budget, char-budget, stub-companion-split, template-mirror, zone-registry, tier-table"
tier: M
era: V3R6
---

# SPEC: spec-workflow.md 40k 예산 초과 수리 — core + detail companion 분할

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-10-08 | manager-spec | 최초 작성 — 카드 t1586 plan-phase. 측정 트리 `.moai/worktrees/spec-workflow-split` @ `81786284e` |

## 1. 문제 — 측정된 형태

카드 t1586: `.claude/rules/moai/workflow/spec-workflow.md` 가 규칙 파일 per-file 문자 예산을 초과했다.

**측정 (이 트리, `81786284e`, 2026-10-08):**

```
$ python3 -c "d=open('.claude/rules/moai/workflow/spec-workflow.md',encoding='utf-8').read(); print(len(d))"
40081
$ wc -c .claude/rules/moai/workflow/spec-workflow.md
40340
```

- 예산 상수: **40,000 runes** — `internal/hook/instructions_loaded.go:124` (`const charBudget = 40000`), `checkCharacterBudget` 가 `utf8.RuneCount` 로 잰다. 초과분은 **81 runes** (runes 기준).
- 카드가 적은 40,081 과 본 트리 실측 40,081 이 **일치**한다(바이트 수 40,340 은 멀티바이트 차이).
- 초과 시 효과: hook 이 `SystemMessage` 로 초과를 보고한다(advisory — 이벤트를 막지 않는다, `instructions_loaded.go:71-77`). 규칙 본문(`coding-standards.md` § File Size Limits)은 이 예산을 모든 instruction 파일의 CI-enforceable heuristic 로 선언한다.

### 1.1 예산 집행 지점 — 실측 지도

| 지점 | 파일:위치 | 대상 | 성격 |
|---|---|---|---|
| per-file 40k | `internal/hook/instructions_loaded.go:112-131` (`checkCharacterBudget`, 상수 L124) | InstructionsLoaded 이벤트의 `file_path`(paths:-scoped 규칙 로드 포함) + CWD 의 AGENTS.md | 런타임 advisory SystemMessage |
| hook 자기검증 | `internal/hook/instructions_loaded_test.go` (L444 "per-file 40000 check unchanged on same fixtures") | 픽스처만 | CI (go test) |
| 집계 210k | `internal/hook/instructions_loaded.go:98` (`sessionCharBudget = 210000`, L143) | AGENTS.md @-import 클로저 + **always-loaded** 규칙(`paths:` 없는 `.claude/rules/moai/**/*.md`) | 런타임 advisory |
| contract-mode 예산 | `internal/template/contract_mode_guided_test.go:586` `TestContractModeAlwaysLoadedBudget` | `grTargets` 중 `alwaysLoaded: true` 만 — **spec-workflow.md 미포함**(`contract_mode_blocks_test.go` grTargets 에 0회 등장, 실측) | CI (go test) |

**핵심 발견 — 배포 트리 규칙 크기를 재는 CI 테스트는 존재하지 않는다.** 카드가 후보로 적은 `TestDeployedAlwaysLoadedCharBudget` 라는 이름은 이 트리에 없다(전수 grep 실측, 0 적중). 따라서 본 SPEC 의 예산 판정 명령은 go test 가 아니라 **직접 rune 계수**다(acceptance.md AC-01). `spec-workflow.md` 는 `paths:` frontmatter(`"**/.moai/specs/**,**/.moai/config/sections/quality.yaml"`)를 가지므로 210k 집계 면에는 들어가지 않지만, per-file 40k 검사는 paths:-scoped 파일에도 적용된다(`coding-standards.md` § File Size Limits — "The budget applies to every instruction file the InstructionsLoaded hook measures — always-loaded and paths:-scoped alike").

## 2. 수리 방향 — 카드 지시대로 stub + lazy companion 분할

81 runes 만 깎는 최소 수리는 앞으로의 성장 여유를 주지 못한다(이 파일은 SPEC 워크플로의 중심 규칙으로 성장 압력이 가장 큰 파일이다). 카드가 지시한 **core stub + lazy companion 분할**을 따른다. 선행: `factory-dispatch.md` (stub 29,309 B) ↔ `factory-dispatch-detail.md` 등 4 companion 분할 (카드 t1483 계열). companion 명명은 workflow/ 의 지배 패턴 `*-detail.md` 를 따른다(`context-window-management-detail.md`, `factory-dispatch-detail.md`, `main-checkout-branch-guard-detail.md` 등 7개 실측).

### 2.1 분할 경계 (실측 바이트 기준)

**CORE — `.claude/rules/moai/workflow/spec-workflow.md` (제자리 재작성, 프로젝션 ≈ 28.8KB ≈ 28% 여유):**

| 블록 | 실측 B | core 잔류 근거 |
|---|---|---|
| frontmatter(`paths:` 불변) + 제목 + intro + companion 안내 | ~610 | 로딩 스코프 계약 |
| Phase Overview | 1,139 | `#phase-overview` 앵커 인용 1건 + 카탈로그 정책 |
| SPEC Phase Discipline — 2-라우트 정의 [ZONE:Frozen][HARD] | 1,429 | 바인딩 룰 |
| SPEC Phase Discipline — step ordering [HARD] bullets | 3,255 | **zone-registry 등록 clause 2건(CONST-V3R5-027/-028)의 원문이 이 안에 있다** — 앵커 `#spec-phase-discipline` 도 유지 |
| Conditional Design Route | 1,061 | 라우팅 룰 (짧음) |
| Subcommand Classification 전체 | 5,530 | 분류 매트릭스 — 앵커 `#subcommand-classification*` 인용 **50건**, `#mode-dispatch-cross-reference` 6건 |
| SPEC Complexity Tier 전체 | 3,450 | **Go 파서가 이 파일에서 직접 읽는다**: `internal/spec/lint_tier_artifacts.go:36` (`tierArtifactRuleRelPath`) + `internal/spec/lint.go:225` — "Artifact set" 열 헤더 + `^([SML]) \(` 행이 이 파일에 있어야 `TierArtifactMissing` 린트가 산다 |
| Plan Phase 전체 | 1,152 | **등록 clause CONST-V3R2-001** ("Create comprehensive specification using EARS format.") + 앵커 `#plan-phase` |
| Run Phase core (진입+토큰 558 / 자동감지 410 / MX 표+성공기준 565 / 재계획 게이트 1,403) | 2,936 | 게이트 트리거·성공기준은 바인딩; run.md Phase 14·loop.md 인용 |
| Sync Phase | 522 | — |
| Context Management | 302 | — |
| Phase Transitions 전체 | 3,996 | **skip 계약 SSOT** ("the ONE authoritative skip contract") — 다른 표면은 인용만 허용되므로 이동 불가 |
| Phase 1 Plan Audit Gate core (intro+entry 498 / verdicts 498 / grace window 292) | 1,288 | `§ Phase 1 Plan Audit Gate` 인용(orchestration-mode-selection.md 등) |
| Agent Teams Variant | 1,407 | — |
| 이동 섹션 포인터 (~5건) | ~750 | stub 계약: "companion 소유 § 나열" |

**COMPANION — `.claude/rules/moai/workflow/spec-workflow-detail.md` (신규, 프로젝션 ≈ 11.3KB):**

| 블록 | 실측 B |
|---|---|
| 자체 frontmatter (`description:` + `paths:`) | ~350 |
| 제목 + 헤더 blockquote (detail-companion 계약문) | ~500 |
| Route A/B step 표 + merge_method 각주 (L30-48) | 2,287 |
| SPEC Phase Discipline anti-patterns + xref (L62-68, [SHOULD]) | 736 |
| Run 방법론 본문: DDD 모드 492 + TDD 모드 562 + self-review/drift/delegation 965 | 2,019 |
| Phase 1: Depends_on Pre-flight (L372-389) | 1,802 |
| Phase 1: Report Persistence (L399-413) | 3,484 |
| footer (Version/Classification) | ~150 |

28.8K + 11.3K ≈ 40.1K ≈ 원본 40.34K + 스캐폴딩 — **이동은 verbatim, 의미 변경 0**(REQ-008).

### 2.2 기계적 구속 (이동 불가 목록 — plan-auditor 가 이 표로 판정한다)

1. **등록 clause 3건의 원문이 core 에 남는다**: `zone-registry.md` 의 CONST-V3R2-001 / CONST-V3R5-027 / CONST-V3R5-028 은 `internal/constitution` 로더가 `strings.Contains` 로 `file:` 이 가리키는 파일에서 검증한다. §2.1 경계는 세 문장이 모두 core 의 보존 블록 안에 있게 설계됐다(앵커 `#plan-phase`, `#spec-phase-discipline` 헤딩도 core 유지).
2. **Tier 표("Artifact set" 열 + S/M/L 행)가 core 에 남는다** — `internal/spec/lint_tier_artifacts.go` 파서 대상.
3. **skip 계약 본문이 core 에 남는다** — 인용 전용 표면 계약상.
4. **분류 매트릭스 + Mode Dispatch Cross-Reference 가 core 에 남는다** — 앵커 인용 50+6건.
5. **`paths:` frontmatter는 core 에서 불변** — 로딩 스코프 변경 금지.

## 3. 이 카드의 대표 mutant

"예산 아래"만 재는 AC는 **내용 유실 분할**로 만족된다: 등록 clause 를 companion 으로 옮기고 registry `file:` 만 고치기, Tier 표 이동 뒤 파서 fail-open 방치, 포인터 없이 섹션 삭제, 한쪽 트리만 수정. 이를 막으려면 AC 가 (a) 예산, (b) mirror byte-parity, (c) 등록 clause/Tier 표/skip 계약의 core 잔류, (d) 이동 원문 보존(verbatim), (e) 크로스레퍼런스 해소를 **각각 독립 명령으로** 판정해야 한다. 상세는 acceptance.md, mutant 대응은 plan.md §G.

## 4. 요구사항 (GEARS)

| REQ | 요구사항 (The system shall …) |
|-----|-------------------------------|
| REQ-001 | core stub 은 §2.1 경계대로 모든 바인딩 룰([ZONE] 태그 라인, 등록 clause 원문, Tier 표, skip 계약, 분류 매트릭스)과 이동 섹션 각각의 포인터를 보존한다. |
| REQ-002 | companion 은 `*-detail.md` 관례(frontmatter `description:`+`paths:`, 제목, 헤더 blockquote, footer)를 갖추고 §2.1 의 이동 블록을 **원문 그대로** 소유한다. |
| REQ-003 | 분할 뒤 core 는 양쪽 트리에서 40,000 runes 미만이다. |
| REQ-004 | core·companion 쌍은 `internal/template/templates/` 미러와 byte-identical 이다. companion 은 `workflowOptMirroredPaths` 에 등재되고 leak-test allowance 에 (파일×SPEC-ID) 항목을 얻는다. |
| REQ-005 | embed manifest 가 재생성되어 새 템플릿 파일이 `go:embed` 발행 집합에 들어간다(`make embed-manifest` → `embed-manifest-check` 클린). |
| REQ-006 | `spec-workflow.md` 를 향한 모든 크로스레퍼런스가 해소된다 — 앵커 링크(#…)는 core 헤딩으로, 이동 섹션 인용(§ Report Persistence 2처×2트리, Run 방법론 안내 1처×2트리)은 companion 경로로 갱신한다. |
| REQ-007 | 기계 소비자가 영향 없다 — `internal/spec` Tier 파서, `internal/constitution` 레지스트리 검증, `internal/hook` 예산 셀프테스트, `internal/template` mirror/leak 테스트가 전부 통과한다. |
| REQ-008 | 이동은 verbatim 이다 — companion 본문은 원본 해당 블록과 바이트 동일(헤더/포인터 문장만 새 텍스트), core 의 보존 블록도 바이트 동일한다. |

## 5. 범위 밖

- `TestDeployedAlwaysLoadedCharBudget` 류 신규 CI 예산 테스트 도입 (후속 카드 후보 — §1.1 의 미측정 공백은 이 SPEC 이 채우지 않는다).
- `spec-workflow.md` 내용의 재편집·요약 (분할은 이동이지 개작이 아니다).
- zone-registry 엔트리 변경 (clause 가 core 에 남으므로 불필요).
- AGENTS.md / output-style 등 타 always-loaded 슬롯 정리.

## 6. 참조

- 선행 분할: `factory-dispatch.md` ↔ `factory-dispatch-{detail,cards,gates,mechanics}.md` (카드 t1483 계열 — byte-parity 등재와 stub 포인터 문장의 규범 예)
- 예산 도구: `internal/hook/instructions_loaded.go` (`charBudget` L124, `sessionCharBudget` L143)
- Go 소비자: `internal/spec/lint_tier_artifacts.go:36,53-99`, `internal/spec/lint.go:225`, `internal/constitution/validator.go`, `internal/runtime/audit_snapshot.go:5-27`(주석 인용)
- mirror/leak 등재: `internal/template/rule_template_mirror_test.go:193,181-289`, `internal/template/internal_content_leak_test.go:514-541,580-592`
