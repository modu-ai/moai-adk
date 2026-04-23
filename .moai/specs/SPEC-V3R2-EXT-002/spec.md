---
id: SPEC-V3R2-EXT-002
title: Output-Styles and Memdir Go Loader
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 2 SPEC writer (Layer 6/7/Cleanup)
priority: P2 Medium
phase: "v3.0.0 — Phase 7 — Extension"
module: "internal/config/types.go, internal/config/loader.go, internal/template/"
dependencies:
  - SPEC-V3R2-WF-006
  - SPEC-V3R2-EXT-001
related_gap:
  - r6-template-only-loaders
  - r3-cc-architecture-memdir
related_theme: "Theme 7 — Extension"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "go-loader, output-styles, memdir, memory, config, v3"
---

# SPEC-V3R2-EXT-002: Output-Styles and Memdir Go Loader

## HISTORY

| Version | Date       | Author | Description                                                           |
|---------|------------|--------|-----------------------------------------------------------------------|
| 0.1.0   | 2026-04-23 | Wave 2 | Initial SPEC — Go loader for output-styles + memdir                   |

---

## 1. Goal (목적)

R6 audit는 output-styles와 agent-memory가 현재 **template-only**이며 Go 런타임에서 파싱되지 않고 있음을 확인했다(R6 §3.4, R6 §5.2). 본 SPEC은 두 리소스에 대한 Go 로더를 `internal/config/` 레이어에 추가하여 (a) SPEC-V3R2-WF-006의 output-style 스키마를 Go 측에서 검증하고, (b) SPEC-V3R2-EXT-001의 4-type 메모리 taxonomy를 Go 측에서 강제하며, (c) SessionStart 훅이 staleness wrap을 적용할 때 Go에서 memory 파일을 읽을 수 있도록 한다. 이는 template drift 감지, CI 검증, runtime enforcement의 단일 토대가 된다.

### 1.1 배경

R6 §3.4: "Both files are identical between `.claude/output-styles/moai/` and `internal/template/templates/.claude/output-styles/moai/`. Clean sync." 하지만 Go 코드에서 output-style을 파싱하는 loader는 존재하지 않는다 — drift detection은 파일 byte-compare에 의존한다. R6 §5.2: "5 yaml sections have no schema in `internal/config/types.go`." 유사하게 `.claude/agent-memory/<agent-name>/*.md`도 Go 레이어에서 read-only 접근 외에는 구조적 validation이 없다. SessionStart hook(`internal/hook/session_start.go`)은 이미 memory evaluation 단계를 가지므로 loader 도입은 인접한 확장이다.

### 1.2 비목표 (Non-Goals)

- Output-style 또는 memdir에 대한 CRUD API 신설
- Memory LLM retrieval/embedding 구현
- Output-style 자동 switching 로직
- Memdir GC / archival 자동화
- CLI 서브커맨드 `moai memory` 신설 (v3.1+ 고려)
- Output-style editing UI
- Memory 파일의 binary content 처리

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns**: `internal/config/output_styles.go` (신규), `internal/config/memory_loader.go` (신규), `internal/config/types.go` 확장.
- Output-style loader: `.claude/output-styles/moai/*.md` 파싱, 스키마(name/description/keep-coding-instructions) 검증, Go struct `OutputStyle` 생성.
- Memdir loader: `.claude/agent-memory/<agent-name>/*.md` 파싱, 4-type enum 검증, Go struct `MemoryEntry` 생성.
- Staleness 판정 helper: `func (m MemoryEntry) IsStale(threshold time.Duration) bool`.
- Drift detection: loader는 template tree와 local tree를 모두 읽어 byte-level diff를 반환.
- SessionStart 훅에서 loader 호출로 staleness wrap 적용(SPEC-V3R2-EXT-001 REQ-EXT001-006 runtime path).

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Output-style editing API / command
- Memory DB / SQLite persistence
- Memory embedding / semantic search
- Output-style 선택 UI
- Memdir 자동 정리/아카이브
- CLI `moai style` / `moai memory` 서브커맨드
- Multi-project 메모리 공유

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.26+), `internal/config/`, `internal/hook/session_start.go`
- 영향 디렉터리:
  - 신설: `internal/config/output_styles.go`, `internal/config/memory_loader.go`
  - 수정: `internal/config/types.go`, `internal/config/loader.go`, `internal/hook/session_start.go`
  - 읽기 전용 참조: `.claude/output-styles/moai/`, `.claude/agent-memory/<agent>/`
- 외부 레퍼런스: R6 §3.4, R6 §5.2, SPEC-V3R2-WF-006, SPEC-V3R2-EXT-001

---

## 4. Assumptions (가정)

- Output-style과 memory 파일은 UTF-8 markdown with YAML frontmatter 형식이다.
- `gopkg.in/yaml.v3`는 이미 `internal/config/`에서 사용 중이므로 신규 의존성 필요 없다.
- SessionStart hook은 동기 실행되며 loader 호출 레이턴시(< 50ms 기대)를 허용한다.
- 각 agent의 memory 파일 수는 50개 이하로 가정(큰 규모 벤치는 후속 SPEC).
- Staleness threshold 기본값은 24시간 (SPEC-V3R2-EXT-001 §7 Constraints와 일치).

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-EXT002-001**
The system **shall** provide a Go struct `OutputStyle` in `internal/config/types.go` with fields: `Name` (string), `Description` (string), `KeepCodingInstructions` (bool), `Body` (string), `SourcePath` (string).

**REQ-EXT002-002**
The system **shall** provide a Go struct `MemoryEntry` with fields: `Name` (string), `Description` (string), `Type` (enum), `Body` (string), `SourcePath` (string), `ModTime` (time.Time).

**REQ-EXT002-003**
The system **shall** provide `LoadOutputStyles(root string) ([]OutputStyle, error)` that reads all style files and validates the frontmatter schema.

**REQ-EXT002-004**
The system **shall** provide `LoadAgentMemory(agentDir string) ([]MemoryEntry, error)` that reads all memory files except `MEMORY.md` index.

**REQ-EXT002-005**
The `Type` enum for `MemoryEntry` **shall** accept exactly 4 values: `user`, `feedback`, `project`, `reference` (SPEC-V3R2-EXT-001 REQ-EXT001-004).

**REQ-EXT002-006**
The loader **shall** reject files with invalid frontmatter and return an aggregated error listing all offending files and keys.

### 5.2 Event-Driven Requirements

**REQ-EXT002-007**
**When** `LoadOutputStyles` encounters a file missing `keep-coding-instructions`, the loader **shall** return `OUTPUT_STYLE_SCHEMA_ERROR` (REQ-WF006-013).

**REQ-EXT002-008**
**When** `LoadAgentMemory` encounters a file with invalid `type`, the loader **shall** return `MEMORY_INVALID_TYPE`.

**REQ-EXT002-009**
**When** SessionStart hook calls `LoadAgentMemory` and any entry has `mtime > threshold`, the hook **shall** wrap the entry's body in `<system-reminder>` per SPEC-V3R2-EXT-001 REQ-EXT001-006.

### 5.3 State-Driven Requirements

**REQ-EXT002-010**
**While** `MemoryEntry.IsStale(threshold)` returns `true`, callers **shall** use `WrapStale(entry)` to format the caveat for agent context.

**REQ-EXT002-011**
**While** the template and local output-style trees are compared via loader, the loader **shall** return a `DriftReport` with per-file SHA256 comparison.

### 5.4 Optional Requirements

**REQ-EXT002-012**
**Where** a consumer needs output-style lookup by `Name`, the loader **shall** provide `FindOutputStyle(styles []OutputStyle, name string) (OutputStyle, bool)`.

**REQ-EXT002-013**
**Where** memory size exceeds 50 entries per agent, the loader **shall** emit metrics `memory.load.entries_per_agent` for observability (no hard fail).

### 5.5 Complex Requirements (Unwanted Behavior / Composite)

**REQ-EXT002-014 (Unwanted Behavior)**
**If** the loader is asked to parse a binary file, **then** it **shall** skip the file and record a warning without failing the batch.

**REQ-EXT002-015 (Unwanted Behavior)**
**If** the frontmatter YAML is malformed (syntax error), **then** the loader **shall** include the raw parse error in the aggregated error and continue to the next file.

**REQ-EXT002-016 (Complex: State + Event)**
**While** SessionStart is executing, **when** `LoadAgentMemory` takes longer than 500ms, the hook **shall** log `memory.load.slow` warning but not block session initialization.

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-EXT002-01**: Given `.claude/output-styles/moai/moai.md` with valid frontmatter When `LoadOutputStyles` is called Then an `OutputStyle` with `Name="MoAI"` is returned (maps REQ-EXT002-003).
- **AC-EXT002-02**: Given a memory file with `type: project` When `LoadAgentMemory` is called Then a `MemoryEntry` with `Type=project` is returned (maps REQ-EXT002-004, REQ-EXT002-005).
- **AC-EXT002-03**: Given an output-style file missing `keep-coding-instructions` When the loader runs Then `OUTPUT_STYLE_SCHEMA_ERROR` with the filename is in the aggregated error (maps REQ-EXT002-007).
- **AC-EXT002-04**: Given a memory file with `type: unknown` When the loader runs Then `MEMORY_INVALID_TYPE` error is returned (maps REQ-EXT002-008).
- **AC-EXT002-05**: Given a memory file with mtime > 24h When SessionStart runs Then body is wrapped in `<system-reminder>` (maps REQ-EXT002-009, REQ-EXT002-010).
- **AC-EXT002-06**: Given both template and local output-style trees When `LoadOutputStyles` diffs them Then a `DriftReport` with SHA256 per file is returned (maps REQ-EXT002-011).
- **AC-EXT002-07**: Given 2 output-styles loaded When `FindOutputStyle(styles, "Einstein")` is called Then returns the Einstein entry with `found=true` (maps REQ-EXT002-012).
- **AC-EXT002-08**: Given a binary file in output-styles directory When loader runs Then the file is skipped with warning (maps REQ-EXT002-014).
- **AC-EXT002-09**: Given a malformed YAML frontmatter When loader runs Then the parse error is included in aggregated error and other files are still loaded (maps REQ-EXT002-015).
- **AC-EXT002-10**: Given 60 memory entries for a single agent When loader runs Then metric `memory.load.entries_per_agent` > 50 is emitted (maps REQ-EXT002-013).
- **AC-EXT002-11**: Given `LoadAgentMemory` takes 700ms When SessionStart completes Then `memory.load.slow` warning is logged and session is not blocked (maps REQ-EXT002-016).
- **AC-EXT002-12**: Given the Go loader is implemented When `go test ./internal/config/...` runs Then all unit tests pass (maps REQ-EXT002-001, REQ-EXT002-002).

---

## 7. Constraints (제약)

- 신규 외부 의존성 도입 금지 (9-direct-dep 정책; `yaml.v3` 기존 의존성 재사용).
- Loader는 read-only — output-style이나 memory 파일을 수정하지 않는다.
- Go 표준 라이브러리 + 기존 `internal/config/` 유틸리티만 사용.
- 에러는 aggregate로 반환 (fail-slow, partial-success 허용).
- 성능: SessionStart 총 레이턴시 증가 < 100ms (p95).

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 완화 |
|---|---|---|
| Memory 파일이 큰 규모일 때 SessionStart 지연 | 세션 시작 느림 | REQ-EXT002-016의 500ms warn threshold + 비블록 처리 |
| YAML frontmatter 스펙 변화 | loader 회귀 | 스키마는 SPEC-V3R2-WF-006/EXT-001에서 고정 |
| Template/local drift 감지 오탐 | false positive | SHA256 기반 byte-compare (오탐 없음) |
| Binary 파일 혼입 시 loader 크래시 | 세션 시작 실패 | REQ-EXT002-014의 skip-on-binary |
| Agent memory의 무단 삭제가 loader에 감지 안됨 | 다운스트림 누락 | SessionStart에서 개수 변화를 메트릭으로 로깅 |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3R2-WF-006: output-style schema 확정.
- SPEC-V3R2-EXT-001: 4-type enum 확정.

### 9.2 Blocks

- SPEC-V3R2-MIG-003: config loader completeness에서 본 SPEC 출력 구조 활용 가능.

### 9.3 Related

- R6 §3, §5.2.

---

## 10. Traceability (추적성)

- REQ 총 16개: Ubiquitous 6, Event-Driven 3, State-Driven 2, Optional 2, Complex 3.
- AC 총 12개, 모든 REQ에 최소 1개 AC 매핑 (100% 커버리지).
- Wave 2 소스 앵커: R6 §3.4 template/local sync; R6 §5.2 5 unloaded sections; SPEC-V3R2-WF-006, SPEC-V3R2-EXT-001.
- BC 영향: 없음 (신규 loader, 기존 파일 수정 없음).
- 구현 경로 예상:
  - `internal/config/output_styles.go`
  - `internal/config/memory_loader.go`
  - `internal/config/types.go` (확장)
  - `internal/hook/session_start.go` (staleness wrap 통합)
  - `internal/config/output_styles_test.go`, `memory_loader_test.go`

---

End of SPEC.
