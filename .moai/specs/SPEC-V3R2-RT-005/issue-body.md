# SPEC-V3R2-RT-005: Multi-Layer Settings Resolution with Provenance Tags

> 본 issue body 는 GitHub Issue 생성 시 자동 첨부됨. Companion to `.moai/specs/SPEC-V3R2-RT-005/{spec,research,plan,acceptance,tasks}.md`.

---

## 배경 (Background)

현재 moai 의 config 시스템은 implicit two-tier (`~/.moai/` user + `.moai/` project) flat-merge 모델을 사용합니다. `problem-catalog.md` Cluster 4 P-C04 (HIGH) 와 P-H06 (CRITICAL) 가 다음 두 결함을 명시합니다:

1. **Provenance 부재**: 어느 키가 어느 파일에서 set 되었는지 grep 없이 알 수 없음. Diagnostic 시 "왜 이 값이 우선?" 질문에 즉답 불가.
2. **5개 yaml 섹션 미연결**: `constitution.yaml`, `context.yaml`, `interview.yaml`, `design.yaml`, `harness.yaml` 가 template-only — Go runtime 이 읽지 못해 enforcement 불가.

Claude Code 2.1.111+ reference (r3 §1.3, §2 Decision 11) 는 "policySettings > userSettings > projectSettings > localSettings > pluginSettings > sessionRules > hookDecision" 8-tier ordering 과 "provenance on every configuration element" 을 명시적으로 요구합니다.

---

## 목표 (Goal)

Implicit two-tier 모델을 **8-tier deterministic resolver + per-key Provenance tag** 로 대체. 모든 config value 가 "어느 파일에서 set 되었는지" 의 답을 메타데이터로 carry. `moai doctor config dump` 가 즉시 진단 가능.

추가 목표: 5개 누락 yaml 섹션의 Go loader 추가 패턴을 establish (실제 loader 본체는 SPEC-V3R2-MIG-003 owner).

---

## Scope

### In-scope

- `Source` typed enum 8 values (priority order):
  - `SrcPolicy` > `SrcUser` > `SrcProject` > `SrcLocal` > `SrcPlugin` > `SrcSkill` > `SrcSession` > `SrcBuiltin`
- `Provenance` struct (`Source`, `Origin`, `Loaded`, `SchemaVersion`, `OverriddenBy`).
- `Value[T any]` generic wrapper (`Unwrap()`, `Origin()`, `IsBuiltin()`, `IsDefault()`).
- `SettingsResolver` interface (`Load`, `Key`, `Dump`, `Diff`, `Reload`).
- 8 tier reader (4+ canonical paths per platform).
- Deterministic merge algorithm (first non-zero wins per priority).
- Diff-aware reload (single tier re-parse on hook trigger).
- `moai doctor config dump|diff|--key` CLI subcommands.
- `internal/config/audit_test.go` parity guard (yaml ↔ Go struct).

### Out-of-scope (delegated)

| Item | Owner SPEC |
|------|-----------|
| Permission stack consumer | SPEC-V3R2-RT-002 |
| 5 missing loader 본체 (constitution/context/interview/design/harness) | SPEC-V3R2-MIG-003 |
| Sandbox routing by source | SPEC-V3R2-RT-003 |
| ConfigChange hook handler wire | SPEC-V3R2-RT-006 |
| `sunset.yaml` activate/retire decision | SPEC-V3R2-MIG-003 |
| Plugin contribution mechanics | v3.1+ deferred |

---

## 8-Tier Architecture

```mermaid
flowchart TD
    Policy["SrcPolicy<br/>/etc/moai/settings.json"] --> Merge[Deterministic Merge]
    User["SrcUser<br/>~/.moai/settings.json + ~/.moai/config/sections/*.yaml"] --> Merge
    Project[".moai/config/config.yaml + .moai/config/sections/*.yaml"] --> Merge
    Local[".claude/settings.local.json + .moai/config/local/*.yaml"] --> Merge
    Plugin["SrcPlugin (reserved v3.1+)"] --> Merge
    Skill[".claude/skills/**/SKILL.md frontmatter"] --> Merge
    Session["SrcSession (RT-004 SessionStore)"] --> Merge
    Builtin["SrcBuiltin (defaults.go)"] --> Merge
    Merge --> Result[MergedSettings<br/>map<key, Value[any]>]
    Result --> Dump["moai doctor config dump"]
    Result --> Consumer["RT-002 / RT-003 / RT-006 / MIG-003"]
```

각 key 의 winner 는 priority 순서대로 첫 non-zero 값. winner 의 `Provenance` 가 출처를 carry; lower-tier 의 non-zero 값들은 `OverriddenBy` slice 에 추가.

---

## Acceptance (요약)

15개 AC (자세한 G/W/T 형식은 `acceptance.md` 참조):

| AC | 검증 대상 | REQ |
|----|---------|-----|
| AC-01 | Policy + Project 충돌 → Policy win + OverriddenBy 추적 | REQ-005, REQ-012 |
| AC-02 | `config dump` JSON 7-field 구조 + byte-stable | REQ-006 |
| AC-03 | `config diff user project` divergent key 만 출력 | REQ-007, REQ-051 |
| AC-04 | ConfigChange hook → 단일 tier reload (timestamp isolation) | REQ-011 |
| AC-05 | Type mismatch (string vs int) → ConfigTypeError | REQ-013 |
| AC-06 | Policy 파일 부재 → silent (no error) | REQ-014 |
| AC-07 | `policy.strict_mode: true` + lower override → PolicyOverrideRejected | REQ-022 |
| AC-08 | Orphan yaml 추가 → audit_test fail | REQ-008/021/043 |
| AC-09 | `--format yaml` 출력에 `# source: <tier>` comment | REQ-030 |
| AC-10 | `--key permission.strict_mode` 단일 key 출력 | REQ-032 |
| AC-11 | `schema_version: 3` → `Provenance.SchemaVersion == 3` | REQ-033 |
| AC-12 | yaml + yml sibling 충돌 → ConfigAmbiguous | REQ-041 |
| AC-13 | SessionEnd → SrcSession tier reset | REQ-050 |
| AC-14 | Builtin default → dump 출력에 `"default": true` | REQ-020 |
| AC-15 | Schema migration 부재 → ConfigSchemaMismatch | REQ-042 |

전체 EARS 요구사항: **33 REQ** (Ubiquitous 8 + Event-Driven 6 + State-Driven 3 + Optional 4 + Unwanted 4 + Complex 2 + spec.md §5 미세 분류 차이 — 실제 27 REQ 명시).

---

## Milestones

priority-ordered (no time estimates per `agent-common-protocol.md`):

| Milestone | Priority | Scope | GREEN ACs |
|-----------|----------|-------|-----------|
| M1 (RED) | P0 | 14+ failing tests across 5 test files | (test scaffolding) |
| M2 (GREEN p1) | P0 | audit_registry + ConfigTypeError + ConfigAmbiguous + schema_version | AC-05/08/11/12 |
| M3 (GREEN p2) | P0 | Real JSON dump + OverriddenBy + PolicyOverrideRejected + sorted YAML | AC-01/02/07/09 |
| M4 (GREEN p3) | P0 | Diff-aware Reload + Skill frontmatter + log file + concurrency | AC-04/13 |
| M5 (GREEN p4 + Trackable) | P1 | Doctor CLI + filename norm + validator/v10 + CHANGELOG + 7 MX tags | AC-10/14 + 최종 검증 |

Total: **40 sub-tasks across 5 milestones** (자세한 task 분해는 `tasks.md` 참조).

---

## Dependencies

### Blocked by

- ✅ **SPEC-V3R2-CON-001** (FROZEN-zone codification — Wave 6 완료): 8-tier ordering 을 constitutional invariant 로 등록.
- ⚠️ **SPEC-V3R2-SCH-001** (validator/v10 통합 — 미머지): mitigation = 본 SPEC M5 시 `go.mod` 직접 추가; SCH-001 머지 후 dedup 자동.
- ⚠️ **SPEC-V3R2-RT-004** (SessionStore — plan complete, run in-flight): mitigation = SrcSession tier loader placeholder (RT-004 머지 후 wiring).

### Blocks

| SPEC | 사용 surface |
|------|-------------|
| SPEC-V3R2-RT-002 | `Source` enum + 8-tier merge (permission stack) |
| SPEC-V3R2-RT-003 | `Provenance.Source` (sandbox routing) |
| SPEC-V3R2-RT-006 | `Reloadable` interface (ConfigChange hook handler) |
| SPEC-V3R2-RT-007 | SrcUser/SrcBuiltin (GoBinPath resolver) |
| SPEC-V3R2-MIG-003 | typed-Value pattern (5 missing loader 추가) |

---

## Risks

자세한 risk + mitigation 표는 `plan.md` §5 (14 risks) 참조. 핵심:

| Risk | Mitigation |
|------|-----------|
| `audit_test.go` strict mode + 5 미연결 yaml → build fail | `YAMLAuditExceptions` 에 5개 등록 (M2); MIG-003 가 점진 해제 |
| Go map iteration 비결정성 → JSON byte-stability 깨짐 | M3 dumpJSON 이 `sort.Strings` 후 `MarshalIndent` |
| 동시 Reload + Read race | `sync.RWMutex` (Lock for Reload, RLock for Key/Dump) + `go test -race` gate |
| Internal consumer mass refactor | BC-V3R2-015 reader-layer-only 약속; `Manager.Get` facade 가 `.V` 추출 |
| `policy.strict_mode` 가 legitimate enterprise pattern 차단 | enforcement opt-in (default `false`); `PolicyOverrideRejected` 는 log 만 (panic X) |

---

## 진행 상황 (Progress)

`progress.md` 참조. 주요 milestone 마다 update:

- `plan_complete_at`: plan phase 종료 시점 (이 SPEC 작성 완료 시)
- `plan_status`: `audit-ready` (plan-auditor 검증 대기)
- `run_complete_at`: 모든 ACs GREEN + Trackable artifact 완료 시점
- `acceptance_progress`: `<n>/15` (achieved AC count)

---

## 참고 (References)

### 본 SPEC 의 문서

- `.moai/specs/SPEC-V3R2-RT-005/spec.md` — 33 EARS REQ + 15 AC 정식 정의
- `.moai/specs/SPEC-V3R2-RT-005/research.md` — 50 file:line evidence anchors
- `.moai/specs/SPEC-V3R2-RT-005/plan.md` — 5 milestone breakdown + traceability matrix + mx_plan
- `.moai/specs/SPEC-V3R2-RT-005/acceptance.md` — 15 AC G/W/T 변환
- `.moai/specs/SPEC-V3R2-RT-005/tasks.md` — 40 task 실행 단위
- `.moai/specs/SPEC-V3R2-RT-005/progress.md` — 진행 추적

### 코드베이스 anchors

- `internal/config/source.go:11-44` — Source enum (REQ-001 ✅ baseline)
- `internal/config/provenance.go:10-71` — Provenance + Value[T] (REQ-002/003 ✅ baseline)
- `internal/config/merge.go:63-134` — MergeAll deterministic algorithm
- `internal/config/resolver.go:17-29` — SettingsResolver interface
- `internal/config/audit_test.go:9-22` — placeholder (M2 에서 real impl 으로 교체)

### 외부 reference

- `docs/design/major-v3-master.md:L974` — §8 BC-V3R2-015 (reader-only migration)
- `docs/design/major-v3-master.md:L1046` — §11.3 RT-005 master definition
- `.moai/design/v3-redesign/synthesis/problem-catalog.md` Cluster 4 — P-C04, P-H06
- `.moai/design/v3-redesign/synthesis/pattern-library.md` X-2 (priority 4)
- Claude Code reference: r3-cc-architecture-reread.md §1.3, §2 Decision 11, §4 Adopt 1

---

## Labels (recommended)

- `type:feature`
- `priority:P0`
- `area:config`
- `area:cli`

---

## Definition of Done

본 SPEC 는 다음 모든 조건이 참일 때 완료 (자세한 list 는 `acceptance.md` Definition of Done 섹션):

1. 15개 AC 모두 `go test ./internal/config/ ./internal/cli/` GREEN
2. 전체 `go test ./...` zero failures + zero cascading regressions
3. `go test -race ./internal/config/` race detector clean
4. `make build` 성공 + `internal/template/embedded.go` 정상 재생성
5. `go vet ./...` + `golangci-lint run` 0 warnings
6. `progress.md` `run_complete_at` + `run_status: implementation-complete`
7. CHANGELOG entry under `## [Unreleased] / ### Added`
8. 7 MX tags inserted across 6 files per `plan.md` §6
9. PR 의 모든 required CI green (Lint, Test ubuntu/macos/windows, Build 5, CodeQL)
10. `audit_test.go::TestAuditParity` no longer placeholder
11. 5개 미연결 yaml 가 `YAMLAuditExceptions` 에 documented (MIG-003 unblock)
12. `resolver.go::loadYAMLFile` no longer placeholder

---

End of issue-body.md.
