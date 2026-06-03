# SPEC-CC-DOCS-ALIGNMENT-001 — Implementation Plan

## §A — Context

본 plan은 5개 공식 Claude Code 문서 대비 33개 documentation/rules 결함을 도메인별 5개 마일스톤(M1~M5)으로 구현한다. 모든 작업은 규칙·문서 본문 in-place 정정이며, Go 소스 변경은 없다(Exclusions §E.1). 핵심 제약은 **template-first mirror 규율**과 **§25 neutrality**다.

## §B — Tier Rationale (M↔L borderline)

**최종 판정: Tier M (3-artifact set: spec.md + plan.md + acceptance.md).**

| 신호 | 측정값 | Tier 함의 |
|------|--------|-----------|
| LOC delta | 결함당 1-3줄 편집, 총 ~80-120줄 | **Tier S/M** (작음) |
| 고유 대상 파일 수 | ~18 (template-managed 14 + LOCAL-ONLY 2 + 일부 중복) | **Tier L 신호** (많음) |
| 실제 파일 touch 수 | template-managed 14 × (source+mirror) + LOCAL-ONLY 2 ≈ **24-28 touch** | **Tier L 신호** |
| inter-file dependency | 없음 (각 결함 독립) | **Tier S/M** (낮은 응집·blast radius) |
| 설계 결정 | 없음 (기계적 본문 정정) | **Tier M** (design.md 불필요) |
| Go code / 아키텍처 변경 | 0 | **Tier M** (run-phase 단순) |

**M↔L borderline 명시**: 파일 touch 수(24-28)만 보면 Tier L 구간에 든다. 그러나 (1) 각 편집이 독립적·기계적이고 inter-file dependency가 없으며, (2) 별도 design.md가 필요한 설계 결정이 전무하고, (3) mirror 동기화는 아키텍처가 아닌 `make build` 기계적 절차이므로 **Tier M으로 분류**하고 3-artifact set을 채택한다. file-count는 높지만 그 높이는 전적으로 "template-managed 파일 1개당 source+mirror 2회 편집"이라는 mirror 이중화에서 비롯되며, 이는 SPEC 복잡도가 아니라 배포 구조의 산물이다. design.md / tasks.md를 추가하면 ceremony만 늘고 가치가 없다.

> 만약 run-phase에서 neutrality split(C3)의 source↔mirror 표현 차이가 예상보다 복잡하다고 판명되면, 그때 design.md를 추가 산출(4-artifact)하는 것을 권고한다 — 현재 plan-phase 검증으로는 split이 단순 토큰 치환 수준이어서 3-artifact로 충분하다.

## §C — Pre-flight Checklist (run-phase 진입 전)

- [ ] `make build`가 로컬에서 동작하는지 확인(template→embedded 재생성 경로)
- [ ] `go test ./internal/template/... -run TestTemplateNeutralityAudit` baseline green 확인
- [ ] `internal/template/internal_content_leak_test.go` baseline green 확인
- [ ] C3 대상 3개 파일(orchestration-mode-selection.md / agent-authoring.md / CLAUDE.md)의 현재 source↔mirror diff를 캡처(편집 후 split 보존 검증 기준선)
- [ ] §C REQ별 grep 앵커(literal 문자열)로 라인 위치 재확인(line-number drift 대비, C5)

## §D — Constraints (요약, spec.md §D 참조)

- C1 template-first mirror / C2 §25 neutrality / C3 neutrality split 보존 / C4 no Go source / C5 grep-anchor 편집.

## §E — Self-Verification (각 마일스톤 공통 DoD)

각 마일스톤 완료 시 다음을 만족해야 한다:

1. **해당 REQ의 본문 정정 완료** (acceptance.md AC grep 통과)
2. **template-managed 파일은 source+mirror 양쪽 갱신** (B.1 부류)
3. **`make build` 실행** (template 변경 후)
4. **neutrality green**: `go test ./internal/template/... -run TestTemplateNeutralityAudit` + `internal_content_leak_test.go`
5. **C3 파일은 split 보존**: source는 내부 토큰 무함유, mirror는 dev-local 표현 유지(byte-identical 강제 안 함)

## §F — Milestones

> 각 마일스톤은 manager-develop(cycle_type per quality.yaml) 1회 위임 단위. SSE stall 임계(30+ tasks) 미달이므로 Round 분할 불필요 — 5개 Milestone 순차 진행.

### M1 — workflows (REQ-CDA-001 ~ 006)

- **대상**: `dynamic-workflows.md`(전 6 REQ), 보조 `orchestration-mode-selection.md`(REQ-CDA-002 cross-ref)
- **부류**: 둘 다 template-managed. dynamic-workflows.md는 byte-IDENTICAL, orchestration-mode-selection.md는 **DIFFER(neutrality split — C3)**.
- **discipline**: dynamic-workflows.md → source 편집 → mirror 동기화 → `make build` → neutrality. orchestration-mode-selection.md를 건드릴 경우 source는 중립 표현, mirror는 dev-local 표현 별도 갱신(split 보존).
- **grep 앵커**: `the \`workflow\` keyword no longer triggers`(REQ-001), `## Disabling Workflows`, `## Saved workflows`, `## How a Workflow Runs`.
- **우선순위**: REQ-001/002 high → REQ-003/004 med → REQ-005/006 low.

### M2 — skills (REQ-CDA-007 ~ 014)

- **대상**: `skill-writing-craft.md`(007/008/010/014), `skill-authoring.md`(008/009/010/011/014), `coding-standards.md`(013), `CLAUDE.md`(014), `claude-code-skills-official.md`(012)
- **부류**: skill-writing-craft / skill-authoring / coding-standards / claude-code-skills-official = byte-IDENTICAL template-managed. **CLAUDE.md = DIFFER(split — C3)**.
- **discipline**: 각 파일 source→mirror→`make build`→neutrality. CLAUDE.md는 split 보존(REQ-CDA-014의 Level 2 문구는 dev-local 토큰 없는 영역이므로 source/mirror 동일 표현 가능하나, 편집 후 기존 split 라인은 건드리지 않는다).
- **grep 앵커**: `type: skill`(REQ-007, 편집 후 schema/example 영역에서 0이어야 함), `≤80 chars`(REQ-008), `1,536`(REQ-008 통일 후 존재), `skillOverrides`(REQ-009), `--add-dir`(REQ-011).
- **주의(REQ-CDA-007)**: `type: skill` 제거는 schema 테이블(라인 219) + 4 example 블록(203/278/291/304) 모두 적용 — "every example, not just the first"(literal instruction following).

### M3 — hooks DOC (REQ-CDA-015 ~ 019, NO Go source)

- **대상**: `hooks-system.md`(015/016/018), `agent-common-protocol.md`(019) = template-managed. `CLAUDE.local.md` + `internal/hook/CLAUDE.md`(017) = **LOCAL-ONLY(mirror 없음)**.
- **discipline**: hooks-system.md / agent-common-protocol.md는 source→mirror→`make build`→neutrality. CLAUDE.local.md / internal/hook/CLAUDE.md는 **직접 편집, mirror·neutrality 검사 없음**(B.2 부류).
- **grep 앵커**: `v2.1.84`(REQ-015, `if` 필드 영역에서 v2.1.85로 변경), `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`(REQ-016), `default: 5 seconds`(REQ-017 CLAUDE.local.md), `≤5s`(REQ-017 internal/hook/CLAUDE.md).
- **주의(REQ-CDA-015)**: hooks-system.md 내부 자기모순(라인 191 v2.1.84 vs 라인 170 v2.1.85) 해소 — `if`/permission-rule filter 버전을 v2.1.85+로 통일.
- **주의(REQ-CDA-019)**: Stop caveat은 `agent-common-protocol.md`에서 Stop hook(sync-phase-quality-gate)을 서술하는 영역에 추가 — "Stop fires every turn-end, not only task completion, and not on user interrupts".

### M4 — goal (REQ-CDA-020 ~ 025)

- **대상**: `goal-directive.md`(전 6 REQ) = template-managed, byte-IDENTICAL.
- **discipline**: source→mirror→`make build`→neutrality.
- **grep 앵커**: `/moai loop` (REQ-020 비교 테이블), `auto mode`(REQ-021), `Remote Control`(REQ-022), `◎ /goal active`(REQ-023), `disableAllHooks`(REQ-024), evaluator cost 노트(REQ-025).
- **주의(REQ-CDA-020)**: native `/loop`(time-interval)을 **별도 행/노트**로 추가하되 기존 `/moai loop`(diagnostic) 행은 유지 — 둘은 다른 명령이므로 병존.

### M5 — sub-agents (REQ-CDA-026 ~ 033)

- **대상**: `agent-authoring.md`(026/027/028/031/032/033), `agent-hooks.md`(027), `archived-agent-rejection.md`(029), `CLAUDE.md`(029), `model-policy.md`(030), 보조 `agent-patterns.md`(028 대안)
- **부류**: agent-hooks / archived-agent-rejection / model-policy = byte-IDENTICAL template-managed. **agent-authoring.md = DIFFER(split — C3)**, **CLAUDE.md = DIFFER(split)**.
- **discipline**: 각 파일 source→mirror→`make build`→neutrality. agent-authoring.md / CLAUDE.md는 split 보존.
- **grep 앵커**: `maxContextSize`(REQ-026, 편집 후 0이어야 함), `maxTurns`(REQ-026 current로 복원), `SubagentStart`(REQ-027), `CLAUDE_CODE_FORK_SUBAGENT`(REQ-028), `claude-code-guide`(REQ-029, built-in 구분 노트 존재), full-ID(REQ-030 `claude-opus-4-8`), `delegate`(REQ-031 enum 밖으로), `agent_type`(REQ-032).
- **우선순위**: REQ-026 high → REQ-027/028/029 med → REQ-030/031/032/033 low.
- **주의(REQ-CDA-026)**: `maxContextSize` 토큰을 agent-authoring.md(source+mirror)에서 완전 제거하고 `maxTurns` deprecation 노트를 제거 — 존재하지 않는 필드 인용 제거가 핵심.
- **주의(REQ-CDA-029)**: `claude-code-guide`를 archived-agent-rejection.md §C 라인 84/131 + CLAUDE.md 라인 127에서 제거하지 말 것(MoAI-archived custom 파일 이력은 유효). 대신 동명 built-in이 별개 유효 agent임을 disambiguation 노트로 추가 — 두 동명 엔터티 구분이 목적.

## §G — Anti-Patterns (이번 작업에서 회피)

- **AP-1 byte-parity 강제**: C3 split 파일을 byte-identical로 만들려다 mirror에 dev-local 토큰을 source로 역류시키면 neutrality 위반. source는 중립·mirror는 dev-local 각각 갱신.
- **AP-2 첫 항목만 정정**: REQ-CDA-007(`type: skill` 5곳), REQ-CDA-026(maxContextSize) 등은 모든 출현 위치를 정정해야 함. "first only" 금지(Opus 4.8 literal instruction following에 맞춰 scope 명시).
- **AP-3 LOCAL-ONLY를 mirror 동기화**: CLAUDE.local.md / internal/hook/CLAUDE.md(REQ-017)를 template에 동기화 시도 금지(template 트리에 부재).
- **AP-4 claude-code-guide 이력 삭제**: REQ-029에서 archived 이력 자체를 지우면 회귀. 구분 노트만 추가.
- **AP-5 make build 누락**: template-managed 편집 후 `make build` 미실행 시 embedded 산출물 drift → neutrality/test 실패.
- **AP-6 규칙 본문에 SPEC ID 임베드**: 어떤 template 규칙 본문에도 "per SPEC-CC-DOCS-ALIGNMENT-001" 등 내부 토큰 작성 금지(§25/C2).

## §H — Cross-References

- spec.md §B File Ownership Map (template-managed vs LOCAL-ONLY 검증표)
- spec.md §E Exclusions (Go hook 결함은 SPEC-HOOK-EVENT-REGISTRY-001)
- CLAUDE.local.md §2 (settings/template separation) + §25 (template internal-content isolation)
- `.moai/docs/template-internal-isolation-doctrine.md` (§25 forbidden/allowed content-class)
- `internal/template/internal_content_leak_test.go` + `.github/workflows/template-neutrality-check.yaml` (CI guard)
