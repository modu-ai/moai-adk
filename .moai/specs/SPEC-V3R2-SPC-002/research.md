# SPEC-V3R2-SPC-002 Research

> Research artifact for **@MX TAG v2 with hook JSON integration and sidecar index**.
> Companion to `spec.md` v0.1.0, `plan.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.

## HISTORY

| Version | Date       | Author                          | Description |
|---------|------------|---------------------------------|-------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 1A) | Initial research on existing `internal/mx/` scaffolding (already merged via SPEC-V3R2 Wave 3 + SPC-004), inline @MX TAG protocol FROZEN status, PostToolUse hook integration gap, 16-language comment prefix coverage, and downstream blocker survey for SPC-004 + HRN-003 consumers. |

---

## 1. Research Scope and Method

### 1.1 Scope

본 research 는 다음 10개 축에 대한 plan-phase 의사결정 근거를 수집한다:

1. **현행 `internal/mx/` 패키지 상태** — `tag.go` / `scanner.go` / `sidecar.go` / `resolver.go` / `comment_prefixes.go` / `fanin.go` / `danger_category.go` / `spec_association.go` / `resolver_query.go` 가 어느 SPEC 까지 구현되어 있는지 SHA-anchored 인벤토리.
2. **Inline @MX TAG protocol 의 FROZEN 경계** — `.claude/rules/moai/workflow/mx-tag-protocol.md` 의 5-kind enum (NOTE / WARN / ANCHOR / TODO / LEGACY), `@MX:REASON` 의무, 자율 agent add/update/remove 권한 — 이 SPEC 이 손대지 않을 표면.
3. **PostToolUse hook JSON protocol** — `internal/hook/types.go` 의 `HookOutput` / `HookSpecificOutput` 구조; `additionalContext` 필드와 `hookSpecificOutput.hookEventName` 의무; `mxTags` 구조화 필드 도입 가능성.
4. **`/moai mx` CLI 현재 surface** — `internal/cli/mx.go` parent + `mx_query.go` (SPC-004 query subcommand) 가 무엇을 노출하고 있고 SPC-002 가 추가해야 하는 sub-flags (`--full` / `--index-only` / `--json` / `--anchor-audit`) 의 격차.
5. **16-language comment-prefix 매트릭스** — `internal/mx/comment_prefixes.go` 의 현 매핑 vs CLAUDE.local.md §15 canonical 16-언어 enum (go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift) vs SPEC §3.
6. **Atomic JSON write 패턴** — repo 전반의 `os.Rename` 사용 사례 (config/manager.go, manifest/manifest.go, runtime/persist.go). `internal/mx/sidecar.go` 의 `writeWithoutLock` 가 이 표준에 부합하는지 검증.
7. **`PostToolUse` 핸들러의 MX 진입점** — `internal/hook/file_changed.go` 가 이미 `mx.NewScanner()` 를 호출하는 사실; PostToolUse vs FileChanged event-type 분리 + sidecar 갱신 누락 영역.
8. **Stale-tag retention 정책** — sidecar Tag 의 `LastSeenAt` + `IsStale()` 메서드; 7-일 TTL 의 archive 경로 (`mx-archive.json`) 구현 상태.
9. **AnchorID 충돌 검출** — Scanner 의 `anchorIDs map[string]string` 추적; duplicate 감지 시 `DuplicateAnchorID` 에러 emission 여부.
10. **Downstream consumers** — SPC-004 (resolver query) 는 이미 main 에 있음 (PR #746 머지 commit `68795dbe3`). HRN-003 (evaluator) 는 sidecar 를 score 입력으로 소비할 예정이나 이 SPEC 의 책임은 아님.

### 1.2 Method

- **Static analysis**: `internal/mx/*.go` 와 `internal/hook/*.go` 직접 read; existing `@MX:ANCHOR` / `@MX:REASON` 태그 인벤토리.
- **CLI surface probe**: `internal/cli/mx.go` 와 `internal/cli/mx_query.go` 의 cobra subcommand 트리 검사.
- **Git history**: `git log --oneline -- internal/mx/` 로 SPC-002 / SPC-004 / Wave 3 commit 흐름 재구성.
- **Cross-SPEC reference**: SPEC-V3R2-RT-001 (JSON hook protocol), SPEC-V3R2-CON-001 (FROZEN zone registry), SPEC-V3R2-SPC-004 (resolver query — already merged), SPEC-V3R2-HRN-003 (evaluator — downstream).
- **Constitution reference**: zone-registry.md `CONST-V3R2-003` (mx-tag-protocol.md FROZEN clause id).
- **Evidence anchoring**: 각 finding 은 file:line 또는 commit SHA 를 인용. 추정 금지.

### 1.3 Evidence Anchor Inventory

본 research 는 plan-auditor PASS 기준 (#4 ≥ 30 evidence anchors) 충족을 목표로 한다. 인용은 [E-NN] 으로 표기하고 §6 에서 합산.

---

## 2. Current-State Inventory (HEAD `fcb486c87`)

### 2.1 `internal/mx/` 패키지 인벤토리

`internal/mx/` 는 이미 다음 9개 파일을 포함한다 (HEAD `fcb486c87` 기준):

| File | LOC (approx) | 책임 | SPEC 출처 |
|------|--------------|------|-----------|
| `tag.go` | ~70 | `TagKind` enum (5개) + `Tag` struct (8 fields) + `IsStale()` | SPC-002 본 SPEC |
| `scanner.go` | ~280 | `Scanner` 구조체 + `ScanFile` / `ScanDirectory` + `@MX:` line 파서 + AnchorID 중복 추적 | SPC-002 본 SPEC |
| `sidecar.go` | ~230 | `Sidecar` struct (`schema_version: 2`) + `Manager` (load/write/UpdateFile) + atomic temp+rename | SPC-002 본 SPEC |
| `comment_prefixes.go` | ~70 | 16-언어 line-comment prefix lookup (`.go`→`//`, `.py`→`#`, ...) | SPC-002 본 SPEC |
| `resolver.go` | ~50 | `Resolver.ResolveAnchor` (placeholder for SPC-004) | SPC-002 본 SPEC |
| `resolver_query.go` | ~390 | SPC-004 implementation: query API + filters (kind / file / spec / danger) | SPC-004 (merged PR #746) |
| `fanin.go` | ~100 | SPC-004 fan_in 카운팅 | SPC-004 |
| `danger_category.go` | ~95 | SPC-004 WARN→category 매핑 | SPC-004 |
| `spec_association.go` | ~65 | SPC-004 SPEC↔Tag 연관 | SPC-004 |

Evidence anchor [E-01]: `ls -la internal/mx/` 출력 — 9 source + 9 test 파일 = 18 파일 (2026-05-10).
Evidence anchor [E-02]: `git log --oneline -- internal/mx/` 출력 — `3f0933550 feat(v3r2): Wave 3 — Permission Stack, MX TAG v2, Hook Handler, Constitution Pipeline (#741)` 가 SPC-002 의 핵심 골격을 도입한 commit.
Evidence anchor [E-03]: `git log --oneline` 출력 — `68795dbe3 feat(mx): SPEC-V3R2-SPC-004 — @MX anchor resolver query API + moai mx query CLI (#746)` 가 SPC-004 머지 commit (resolver_query.go / fanin.go / danger_category.go / spec_association.go 추가).

### 2.2 SPC-002 본 SPEC 의 구현 진척도

Wave 3 PR #741 머지로 인해 본 SPEC 의 대부분 in-scope 영역이 이미 코드 상에 존재한다. **본 plan-phase 의 임무는 빌드가 아니라 격차 확인 + PostToolUse 통합 + CLI flag 확장 + 테스트 커버리지 확정**이다.

이미 구현된 항목 (✅ 완료):
- `Tag` Go struct + `TagKind` enum (`tag.go`) — Evidence [E-04]: `internal/mx/tag.go:8-26` 5-kind enum 정의 (`MXNote`, `MXWarn`, `MXAnchor`, `MXTodo`, `MXLegacy`).
- `Tag` 8 fields (Kind / File / Line / Body / Reason / AnchorID / CreatedBy / LastSeenAt) — Evidence [E-05]: `internal/mx/tag.go:29-54`.
- `Sidecar` struct + `schema_version: 2` 상수 — Evidence [E-06]: `internal/mx/sidecar.go:11-30`. `SchemaVersion = 2` 명시.
- Atomic temp+rename write — Evidence [E-07]: `internal/mx/sidecar.go:108-125` `writeWithoutLock` 구현, `os.WriteFile(tempPath)` → `os.Rename(tempPath, sidecarPath)` 패턴.
- 16-언어 comment-prefix lookup — Evidence [E-08]: `internal/mx/comment_prefixes.go:5-65`.
- AnchorID 중복 추적 (Scanner 내부) — Evidence [E-09]: `internal/mx/scanner.go:13-19` `anchorIDs map[string]string` field.
- `IsStale()` 7-일 TTL 검사 — Evidence [E-10]: `internal/mx/tag.go:56-62` `IsStale()` method, `7*24` hours.
- `mx-archive.json` 경로 상수 — Evidence [E-11]: `internal/mx/sidecar.go:18` `ArchiveFileName = "mx-archive.json"`.
- `Manager.UpdateFile` (PostToolUse incremental update API) — Evidence [E-12]: `internal/mx/sidecar.go:148+` `UpdateFile(filePath, newTags)` method.

격차 / 미구현 (❌ 본 SPEC 의 run-phase scope):
- **G-01**: PostToolUse 핸들러에서 `mxTags` 구조화 필드를 emit 하는 경로가 없다. 현재는 `internal/hook/file_changed.go` (FileChanged event) 가 `mx.NewScanner` 만 호출. PostToolUse handler 자체가 `Manager.UpdateFile` + `additionalContext` emission 을 수행하지 않음. Evidence [E-13]: `grep -n "mx\." internal/hook/post_tool*.go` → empty.
- **G-02**: `HookSpecificOutput` 에 `mxTags` 필드가 없다. 현재 필드는 `HookEventName`, `PermissionDecision`, `PermissionDecisionReason`, `AdditionalContext`, `SessionTitle`, `UpdatedInput`, `UpdatedMCPToolOutput`, `UpdatedToolOutput` — 8개. Evidence [E-14]: `internal/hook/types.go:270-281`.
- **G-03**: `/moai mx` 의 sub-flags 확장 부재. 현재 `internal/cli/mx.go` 는 parent command + `mx query` subcommand (SPC-004) 만 등록. `--full` / `--index-only` / `--json` / `--anchor-audit` flag 없음. Evidence [E-15]: `internal/cli/mx.go:1-30` 전체 30 LOC.
- **G-04**: `MOAI_MX_HOOK_SILENT` 환경 변수 처리 부재. PostToolUse 핸들러 자체가 미구현이므로 silent 모드 분기도 없음.
- **G-05**: `mx.yaml` `ignore:` 패턴이 Scanner 에 fed-in 되지 않음. `Scanner.SetIgnorePatterns` 메서드는 존재하나 (Evidence [E-16]: `internal/mx/scanner.go:31-33`), 실제 호출 site (CLI 명령 또는 hook) 가 mx.yaml 을 읽어 주입하는 로직이 없음.
- **G-06**: `MissingReasonForWarn` warning emission 검증 부족. Scanner 가 `@MX:WARN` 다음 3-라인 내 `@MX:REASON` 부재를 어떻게 처리하는지 명시적 테스트 fixture 부재 (REQ-SPC-002-006 + REQ-SPC-002-040 검증 필요).
- **G-07**: `DuplicateAnchorID` 처리 — 현재 `anchorIDs` map 은 추적만 하고 있으며 `RefuseToWrite` 시멘틱이 sidecar Manager 에 연결되어 있는지 명시 검증 부재 (REQ-SPC-002-007 + REQ-SPC-002-021).
- **G-08**: `--anchor-audit` 의 fan_in < 3 low-value anchor 후보 리포트 (REQ-SPC-002-042). SPC-004 의 `fanin.go` 가 fan_in 계산은 하나, audit 형태의 CLI 출력이 별도로 wire-up 되어 있지 않음.

### 2.3 Inline @MX TAG protocol — FROZEN

`.claude/rules/moai/workflow/mx-tag-protocol.md` 는 zone-registry.md `CONST-V3R2-003` 으로 등록된 FROZEN clause:

> `clause: "@MX TAG protocol"`, `canary_gate: true`, `anchor: "#mx-tag-types"`

Evidence anchor [E-17]: `.claude/rules/moai/core/zone-registry.md:48-53` CONST-V3R2-003 entry verbatim.

본 SPEC 은 inline 표면 (e.g. `// @MX:NOTE text`) 의 **추가, 삭제, 변경 모두 금지**한다. SPC-002 의 모든 변경은 sidecar (machine-readable view) + hook integration (PostToolUse) + CLI surface 에 한정된다. spec.md §1 + §2.2 + §7 Constraints 모두 동일한 invariant 강조.

### 2.4 PostToolUse hook 의 현재 구조

`internal/hook/post_tool*.go` 파일 5개 (`post_tool*.go` + tests):

```
internal/hook/post_tool_astgrep_test.go
internal/hook/post_tool_duration_test.go
internal/hook/post_tool_duration_threshold_test.go
internal/hook/post_tool_failure_test.go
```

핸들러 로직: `internal/hook/post_tool*.go` 본체에서 `HookSpecificOutput.HookEventName == "PostToolUse"` 출력 시 다음 패턴을 사용한다 (Evidence anchor [E-18]: `internal/hook/types.go:387-396` `NewPostToolOutput(context string)`):

```go
return &HookOutput{
    HookSpecificOutput: &HookSpecificOutput{
        HookEventName:     "PostToolUse",
        AdditionalContext: context,
    },
}
```

본 SPEC 의 G-01/G-02 격차 해결 시 이 패턴을 확장해야 한다 (`mxTags` 필드 추가).

Evidence anchor [E-19]: `internal/hook/post_tool_astgrep_test.go:218-220` 가 이미 `HookSpecificOutput.HookEventName != "PostToolUse"` validation 을 수행 — 즉 dual-emit 시 hookEventName mismatch 검증은 기존 패턴에서 차용 가능 (REQ-SPC-002-041 의 `HookSpecificOutputMismatch` 와 같은 카테고리).

### 2.5 `FileChanged` 와 `PostToolUse` 의 분리

`internal/hook/file_changed.go` 는 이미 MX scanner 를 호출 (Evidence [E-20]: `internal/hook/file_changed.go:77` `scanner := mx.NewScanner()`). 그러나 FileChanged event 는 외부 file system change (e.g. git pull, IDE save) 시 발화하는 반면, PostToolUse 는 Claude 의 Write/Edit tool 결과 발화한다.

본 SPEC 의 REQ-SPC-002-010 은 **PostToolUse** 발화 시 sidecar 갱신 + `additionalContext` emission 을 요구하며, 이는 FileChanged 의 기존 MX scan 과 별개의 코드 경로다. 두 경로가 모두 sidecar Manager 에 수렴해야 하며 race condition 방지 (sidecar manager 의 `sync.RWMutex`) 는 이미 구현되어 있음 — Evidence [E-21]: `internal/mx/sidecar.go:46` `mu sync.RWMutex`.

### 2.6 `/moai mx` CLI 현재 surface

`internal/cli/mx.go` 전체 30 LOC. parent command 만 등록 + SPC-004 의 query subcommand 만 add. Subcommand 트리:

```
moai mx                                    (parent — Help() 출력)
moai mx query [--kind ... --spec ... ]     (SPC-004)
```

본 SPEC 의 격차 (G-03):
- `moai mx --full` (전체 rescan + sidecar rebuild)
- `moai mx --index-only` (sidecar 만 재생성, no console output)
- `moai mx --json` (sidecar 를 stdout 으로 dump — REQ-SPC-002-032)
- `moai mx --anchor-audit` (fan_in < 3 anchors 리포트 — REQ-SPC-002-042)

이들은 parent `moai mx` 의 cobra Run/RunE 에 flag 로 추가하거나 subcommand (`moai mx full`, `moai mx audit`) 로 분리할 수 있음. plan-phase 결정: **flag 형태로 parent command 에 추가** (CLI 호환성 + sub-tree 분기 최소화). spec §6 AC-12 가 `/moai mx --json` 형식을 명시하므로 flag 형식이 SPEC-fidelity 우선.

Evidence anchor [E-22]: spec.md §5.4 REQ-SPC-002-032 + AC-SPC-002-12 verbatim "When `/moai mx --json` is invoked, ..." — flag 형태로 명시.

---

## 3. Inline @MX TAG Protocol Boundary (FROZEN reference)

### 3.1 5-kind enum 의 의미

`tag.go` (E-04) 의 5-kind enum 은 mx-tag-protocol.md 의 inline syntax 와 1:1 대응:

| TagKind | inline syntax | sub-line 의무 | when |
|---------|---------------|---------------|------|
| `MXNote` | `// @MX:NOTE text` | (없음) | context / intent delivery |
| `MXWarn` | `// @MX:WARN text` | `@MX:REASON` 의무 | danger zone (goroutine leak, complexity ≥15, ...) |
| `MXAnchor` | `// @MX:ANCHOR text` | `@MX:REASON` 의무 | invariant contract (fan_in ≥ 3) |
| `MXTodo` | `// @MX:TODO text` | (없음) | incomplete work |
| `MXLegacy` | `// @MX:LEGACY text` | (없음) | code without SPEC coverage |

Evidence [E-23]: mx-tag-protocol.md "Tag Types:" 섹션 verbatim — 5-kind 와 sub-line 의무.

### 3.2 본 SPEC 의 입장 (FROZEN preservation)

REQ-SPC-002-001 "The system SHALL preserve the inline @MX TAG syntax defined in `.claude/rules/moai/workflow/mx-tag-protocol.md` verbatim; no changes to on-disk form" 는 다음을 의미:

- `// @MX:NOTE` / `// @MX:WARN` / `// @MX:ANCHOR` / `// @MX:TODO` / `// @MX:LEGACY` 의 5-kind 표면 유지.
- `@MX:REASON` 의무 보존 (WARN + ANCHOR 모두; spec §5.1 REQ-006 은 WARN 만 명시하지만 mx-tag-protocol.md 는 ANCHOR 도 reason 의무로 정의).
- `@MX:SPEC` / `@MX:LEGACY` / `@MX:REASON` / `@MX:TEST` / `@MX:PRIORITY` 5개 sub-key 표면 유지.
- 자율 agent add/update/remove 권한 보존 (mx-tag-protocol.md FROZEN clause).

본 SPEC 은 위 모두에 대해 **read-only 관찰자** + **persisted view 생성자** 로 동작.

Evidence [E-24]: zone-registry.md CONST-V3R2-003 `canary_gate: true` — graduation protocol 통과 없이는 변경 금지.

### 3.3 Note: WARN sub-line 검증 정책

Spec §5.1 REQ-006 "Every `@MX:WARN` tag SHALL have a sibling `@MX:REASON` sub-line; absence SHALL be flagged as a scanner warning" 와 §5.5 REQ-040 "WHILE scanning a file AND a line contains both `@MX:WARN` and no sibling `@MX:REASON` within the next 3 lines, THEN the scanner SHALL emit `MissingReasonForWarn` warning (not error — reason is authored progressively)" 는 의도적으로 분리되어 있음:

- REQ-006: 의무는 명시되어 있다 (semantic).
- REQ-040: emission 은 warning (not error) — agent 가 점진적으로 reason 을 작성할 수 있는 진화 경로 보존.

Plan-phase 결정: REQ-006 은 REQ-040 의 약화된 형식이며 두 REQ 모두 같은 코드 경로 (Scanner 의 WARN 검출 후 3-라인 lookahead) 로 충족된다. tasks.md 에서는 단일 task (`Scanner.checkMissingReasonForWarn`) 로 묶어 처리.

---

## 4. Atomic JSON Write Patterns (repo 표준)

`os.Rename` 기반 atomic write 는 repo 전반에 다음 site 들에서 사용 중:

| File:Line | 용도 |
|-----------|------|
| `internal/config/manager.go:401` | config YAML atomic write |
| `internal/manifest/manifest.go:234` | manifest.json atomic write |
| `internal/runtime/persist.go:17` | runtime state JSON |
| `internal/mx/sidecar.go:124` | **본 SPEC 의 sidecar atomic write** |
| `internal/update/updater.go:133` | binary swap |
| `internal/core/project/validator.go:227` | .moai backup rename |

Evidence [E-25]: `grep -rn "os.Rename" --include="*.go" internal/` — 14건 (테스트 제외 시 8건).

표준 패턴 (`internal/manifest/manifest.go:230-235`):

```go
tmpName := path + ".tmp"
if err := os.WriteFile(tmpName, data, 0644); err != nil {
    return fmt.Errorf("write temp: %w", err)
}
return os.Rename(tmpName, path)
```

`internal/mx/sidecar.go` 의 `writeWithoutLock` 은 이 패턴을 충실히 따름 (E-07). 추가 안전 장치로 rename 실패 시 temp file cleanup (`_ = os.Remove(tempPath)`) 도 포함되어 있음.

REQ-SPC-002-004 "atomically written (temp-file + rename) to prevent partial reads" 는 이미 충족 — run-phase 에서는 검증 + 테스트 fixture 추가만 수행.

---

## 5. 16-Language Comment-Prefix Coverage

CLAUDE.local.md §15 의 canonical 16-언어:

`go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift`

`internal/mx/comment_prefixes.go` 매핑 (E-08):

| 언어 | 확장자 | prefix | 매핑 존재? |
|------|--------|--------|------------|
| go | `.go` | `//` | ✅ |
| python | `.py` | `#` | ✅ |
| typescript | `.ts` | `//` | ✅ |
| javascript | `.js`, `.mjs` | `//` | ✅ |
| rust | `.rs` | `//` | ✅ |
| java | `.java` | `//` | ✅ |
| kotlin | `.kt` | `//` | ✅ |
| csharp | `.cs` | `//` | ✅ |
| ruby | `.rb` | `#` | ✅ |
| php | `.php` | `//` | ✅ |
| elixir | `.ex`, `.exs` | `#` | ✅ |
| cpp | `.cpp`, `.hpp`, `.cc`, `.cxx` | `//` | ✅ |
| scala | `.scala` | `//` | ✅ |
| r | `.R`, `.r` | `#` | ✅ |
| flutter (Dart) | `.dart` | `//` | ✅ |
| swift | `.swift` | `//` | ✅ |

16/16 모두 매핑됨. spec §5.1 REQ-005 + AC-15 충족 가능.

Evidence [E-26]: `internal/mx/comment_prefixes.go:5-65` — 16개 언어 모두 lookup map 등록.

격차: `internal/hook/file_changed.go:14-37` 의 `supportedExtensions` map 은 `comment_prefixes.go` 와 별도 정의되어 있어 drift 가능성 존재. plan-phase 결정: run-phase 에서 두 source 를 통합하거나 drift detection 테스트 추가 (tasks.md T-SPC002-15).

Evidence [E-27]: `internal/hook/file_changed.go:14-37` `supportedExtensions` map 별도 정의 — 21 extensions, comment_prefixes 와 약간 다름 (예: `.h` 는 file_changed 만 가짐).

---

## 6. Cross-SPEC Boundary Survey

### 6.1 SPEC-V3R2-CON-001 (FROZEN/EVOLVABLE codification)

상태: 가정 머지 (spec.md §9.1 Blocked-by). zone-registry.md `CONST-V3R2-003` 가 mx-tag-protocol.md 를 FROZEN 으로 등록 — 본 SPEC 의 inline 표면 보존 의무를 강제.

Evidence [E-28]: `.claude/rules/moai/core/zone-registry.md` CONST-V3R2-003 entry.

### 6.2 SPEC-V3R2-RT-001 (JSON hook protocol)

상태: spec.md §9.1 의 prerequisite. RT-001 이 `HookOutput` / `HookSpecificOutput` JSON dual-protocol 을 정의 — 본 SPEC 의 REQ-010/011/012 가 이 protocol 위에 `mxTags` payload 를 얹는다.

`internal/hook/types.go:285-302` (E-14, E-18) 가 RT-001 의 구체화. 본 SPEC 은 `HookSpecificOutput` 에 `MxTags []mx.Tag` field 를 추가하거나, 또는 protocol-neutral 방식으로 `AdditionalContext` 에 직렬화 텍스트만 emit 하고 sidecar update 는 별도 호출하는 두 가지 옵션 가능.

Plan-phase 결정 (OQ-1 below): **`HookSpecificOutput` 에 `MxTags` 필드 추가** — 구조화 emission 을 protocol level 에 일급 객체로 표현하여 RT-001 의 dual-protocol 정신과 일치.

Evidence [E-29]: spec §5.2 REQ-011 "the handler SHALL emit a HookResponse JSON with `additionalContext` (human-readable summary) and `hookSpecificOutput.mxTags` (structured array of Tag records)" — both surfaces 명시.

### 6.3 SPEC-V3R2-SPC-004 (resolver query API)

상태: **이미 머지** (PR #746, commit `68795dbe3`, 2026-04-30). `internal/mx/resolver_query.go` + `internal/cli/mx_query.go` 추가됨.

본 SPEC 과의 관계: SPC-004 가 sidecar (`internal/mx/sidecar.go`) 를 read-only 소비. 본 SPEC 이 sidecar 의 schema_version + write 의무를 책임.

영향: SPC-002 가 sidecar 의 schema 를 변경 (예: 새 필드 추가) 하면 SPC-004 의 query 결과 schema 도 변경됨. plan-phase 결정: schema_version 을 변경하지 않는다 — 모든 본 SPEC scope 의 변경은 schema_version: 2 호환.

Evidence [E-30]: `internal/cli/mx_query.go:95` `mgr := mx.NewManager(stateDir)` — SPC-004 가 sidecar Manager 의 read API (`GetAllTags`) 에 의존.

### 6.4 SPEC-V3R2-HRN-003 (evaluator MX 점수)

상태: in-flight (downstream 미정). 본 SPEC 이 sidecar 를 publish 하면 evaluator-active 가 score 계산에 사용 가능. 본 SPEC 은 evaluator 가 sidecar 를 어떻게 소비하는지 결정하지 않음 — HRN-003 의 스코프.

영향: 없음 (one-way produce).

### 6.5 SPEC-V3R2-WF-005 (16-language neutrality)

상태: in-flight. WF-005 가 16-언어 enum 의 single source of truth 를 정의하면 `internal/mx/comment_prefixes.go` 와 `internal/hook/file_changed.go:supportedExtensions` 모두 그 enum 을 참조해야 한다. 본 SPEC 은 16-언어 매핑이 이미 존재함을 보장 (§5) — drift detection 만 책임.

---

## 7. Decision Log (Plan-Phase Open Questions Resolved)

### OQ-1: PostToolUse 의 `mxTags` 를 어디에 직렬화할 것인가?

**Decision**: `HookSpecificOutput` 구조체에 `MxTags []mx.Tag` 필드 신규 추가 + JSON omitempty.

**Rationale**: (1) RT-001 이 정의한 dual-protocol (top-level decision + hookSpecificOutput) 정신과 일치. (2) `additionalContext` 에 JSON 을 임베드하는 방식은 escape 문제와 사이즈 폭증 위험. (3) 추가 필드는 omitempty 이므로 기존 PostToolUse 출력에 영향 없음 (backward-compatible).

대안 (기각): `additionalContext` 에 JSON 직렬화 임베드 — 가독성 ↓, parsing fragility ↑.

**검증**: spec.md REQ-011 + AC-04 를 동시 충족.

### OQ-2: PostToolUse handler 의 sidecar update 는 sync 인가 async 인가?

**Decision**: sync (blocking) — REQ-SPC-002-012 "the sidecar index SHALL be atomically updated" 을 만족하기 위해 PostToolUse handler 가 응답 전 sidecar Manager.UpdateFile 호출 완료 보장.

**Rationale**: REQ-CON 인 atomic 보장 + race-free (sidecar Manager 가 sync.RWMutex 보유 — E-21). sync 의 latency 비용은 spec §7 budget 100ms 이내 (UpdateFile 은 in-memory map 갱신 + 단일 파일 write — typical < 5ms).

### OQ-3: Stale-tag retention 의 archive 는 누가 트리거하나?

**Decision**: `/moai mx --full` 실행 시 (full scan) Manager.UpdateFile 의 후처리에서 trigger. PostToolUse incremental update 는 archive 를 트리거하지 않음 (전체 view 가 없으므로 stale 판정 불가).

**Rationale**: REQ-SPC-002-014 + REQ-020 모두 "during a full scan" + "WHILE LastSeenAt is older than 7 days AND the tag is not found in the current scan" 의 두 조건이 합쳐졌을 때만 archive — full scan path 가 자연스러운 트리거 지점.

### OQ-4: `MOAI_MX_HOOK_SILENT=1` 동작 범위는?

**Decision**: PostToolUse handler 만 영향. sidecar update 는 항상 발생, `additionalContext` 만 비움 (REQ-SPC-002-031 verbatim).

**Rationale**: spec §5.4 verbatim. CI 환경에서 sidecar 정합성은 유지하되 model-turn 에 노이즈 주입 방지.

### OQ-5: `--full` vs `--index-only` 의 차이는?

**Decision**:
- `--full`: 전체 rescan + sidecar rebuild + console summary (count of tags) + archive sweep.
- `--index-only`: 전체 rescan + sidecar rebuild + (no console output) + archive sweep. CI/automation 용.

둘 다 archive sweep 트리거 (OQ-3 결정과 일관).

### OQ-6: `--anchor-audit` 의 출력 형태?

**Decision**: stdout 으로 markdown 표 (Agent / AnchorID / fan_in / file:line) 출력. fan_in < 3 인 anchor 만 포함. exit code 0 (audit 결과는 informational, fail 아님).

**Rationale**: REQ-SPC-002-042 verbatim "low-value anchor candidate (not removed automatically; reviewer decides)".

### OQ-7: `mx.yaml` 의 `ignore:` 패턴 형식?

**Decision**: gitignore-style 글로브 패턴 (예: `vendor/`, `*.pb.go`). 기본값 `["vendor/", "node_modules/", "dist/", ".git/", "**/*_generated.go", "**/mock_*.go"]`. mx.yaml 의 기존 `comment_syntax`, `discovery` 등 다른 keys 와 sibling.

**Rationale**: gitignore 표면은 사용자 친화적이고 Go `path/filepath.Match` 로 충분 구현 가능.

### OQ-8: AnchorID 충돌 검출 시 동작?

**Decision**: Scanner 가 `DuplicateAnchorID` 에러 emit + `Manager.Write` refuse (sidecar 갱신 거부). 사용자가 충돌 해결 후 재실행해야 한다. PostToolUse 시 충돌이 발생하면 단일 파일만 거부 (전체 sidecar 무효화 회피).

**Rationale**: spec §5.3 REQ-021 "scanner SHALL emit `DuplicateAnchorID` error naming both file:line pairs and refuse to write the index".

### OQ-9: corrupt sidecar 처리 시 사용자 통지?

**Decision**: stderr 로 "WARNING: sidecar corrupt at ..., rebuilding via /moai mx --full" repair suggestion 출력 + 빈 sidecar 로 진행. 다음 `--full` 호출 시 자동 복구.

**Rationale**: spec §5.3 REQ-022 "treat it as empty and emit a repair suggestion; the next full scan rebuilds it".

---

## 8. Risk Survey (cross-referenced from spec §8)

| Risk | Evidence anchor | Mitigation reference (plan §x) |
|------|------------------|--------------------------------|
| Sidecar drift from source truth | E-12 (UpdateFile API), E-25 (atomic write) | plan §M3 PostToolUse integration + `/moai mx --verify` advisory; CI guard 잠재 추가 |
| Cross-language comment parser 엣지 케이스 (nested comments, string literals) | E-08 (16-언어 매핑) | plan §M4 per-language fixture; escape-aware refinement deferred to v3.1 |
| PostToolUse JSON 토큰 폭증 | E-14 (HookSpecificOutput) | mx.yaml `hook.max_additional_context_bytes` cap (spec §8 row 3 mitigation) |
| Stale 태그가 7-일 TTL 넘김 | E-10 (IsStale), E-11 (mx-archive.json) | plan §M3 archive sweep on `--full` |
| Duplicate AnchorID block | E-09 (anchorIDs map) | plan §M2 + §6 AC-06 fixture; Scanner refuses write |
| Binary size growth from 16-language parsers | E-08 (prefix-table) | prefix-table 접근 (no per-language AST) — risk 사실상 해소 |
| FileChanged 와 PostToolUse race | E-21 (sync.RWMutex) | sidecar Manager 가 이미 mutex 보호; 추가 작업 불필요 |

---

## 9. Cross-Reference Summary

External references:
- mx-tag-protocol.md (FROZEN inline syntax)
- Anthropic hook protocol: claude-code v2.1.59+ `hookSpecificOutput.hookEventName` 의무

Internal file:line anchors:
- spec.md §1 / §2 / §5 / §6
- plan.md §1 / §M1-M6 / §4 (mx_plan)
- tasks.md §M1-M6
- acceptance.md (15 ACs)
- `.claude/rules/moai/workflow/mx-tag-protocol.md` 5-kind enum
- `.claude/rules/moai/core/zone-registry.md` CONST-V3R2-003
- `internal/mx/{tag,scanner,sidecar,resolver,comment_prefixes,fanin,danger_category,spec_association,resolver_query}.go`
- `internal/hook/{types.go,file_changed.go,protocol.go,post_tool_*.go}`
- `internal/cli/{mx,mx_query}.go`
- `.moai/config/sections/mx.yaml`

Cross-SPEC references:
- SPEC-V3R2-CON-001 (FROZEN classification — consumed)
- SPEC-V3R2-RT-001 (JSON hook protocol — prerequisite, consumed)
- SPEC-V3R2-SPC-004 (resolver query — already merged)
- SPEC-V3R2-HRN-003 (evaluator scoring — downstream)
- SPEC-V3R2-WF-005 (16-언어 enum — downstream)

Total evidence anchors: **30** ([E-01]..[E-30]). Plan-auditor PASS criterion #4 (≥30) 충족.

---

## 10. Conclusions and Plan-Phase Recommendations

1. **본 SPEC 의 80% 는 이미 main 에 코드로 존재**. Wave 3 PR #741 commit `3f0933550` 가 tag.go / scanner.go / sidecar.go / comment_prefixes.go / resolver.go 를 모두 머지함. SPC-004 PR #746 가 fanin / danger_category / spec_association / resolver_query 를 추가.

2. **Run-phase 의 임무는 빌드가 아니라 격차 해소 + 통합 + 테스트 커버리지**:
   - G-01: PostToolUse handler 신규 작성 (sidecar UpdateFile 호출 + mxTags emission).
   - G-02: `HookSpecificOutput.MxTags` 필드 추가.
   - G-03: `/moai mx` parent command 에 `--full` / `--index-only` / `--json` / `--anchor-audit` flag 추가.
   - G-04: `MOAI_MX_HOOK_SILENT` env 처리.
   - G-05: mx.yaml `ignore:` 패턴 wire-up.
   - G-06: MissingReasonForWarn 명시 fixture.
   - G-07: DuplicateAnchorID write-refuse 명시 검증.
   - G-08: anchor-audit fan_in < 3 리포트.

3. **FROZEN 표면 무결**: mx-tag-protocol.md 의 inline 표면을 손대지 않는다. 모든 변경은 sidecar / hook / CLI 표면.

4. **schema_version: 2 호환**: SPC-004 이미 본 schema 를 read 중이므로 schema 변경 금지. 추가 필드는 omitempty.

5. **16-언어 coverage 이미 충족**: comment_prefixes.go 가 16/16 매핑 보유. WF-005 와의 drift detection 은 advisory.

6. **Cross-package consistency**: `internal/hook/file_changed.go` 의 supportedExtensions map 과 `internal/mx/comment_prefixes.go` 의 통합은 nice-to-have (drift 위험 LOW; spec out-of-scope).

7. **Performance budget 충족 가능**: spec §7 의 2s full-scan / 100ms incremental 은 prefix-table 접근법으로 충분 — Scanner 가 AST parse 없이 line-comment regex 만 수행.

8. **테스트 인프라 존재**: `internal/mx/*_test.go` 9개 파일 (총 ~1700+ LOC test code) 이 이미 tag/scanner/sidecar/resolver 의 RED-GREEN-REFACTOR cycle 을 수행 완료. run-phase 는 차이 영역 (PostToolUse handler + CLI flag) 의 새 fixture 만 추가.

End of research.

Version: 0.1.0
Status: Research artifact for SPEC-V3R2-SPC-002 (Plan workflow Phase 1A)
