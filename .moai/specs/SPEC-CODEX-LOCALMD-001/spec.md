---
id: "SPEC-CODEX-LOCALMD-001"
title: "Codex launcher loads CLAUDE.local.md alongside AGENTS.local.md as developer_instructions"
version: "1.0.1"
status: "completed"
created: "2026-09-22"
updated: "2026-09-22"
author: "manager-spec"
priority: "P1"
phase: "v3.1.3 target"
module: "internal/cli"
lifecycle: "spec-anchored"
tags: "codex,launcher,local-instructions,cli"
tier: "M"
---

## HISTORY

| 날짜 | 버전 | 내용 | 작성자 |
|---|---|---|---|
| 2026-09-22 | 1.0.1 | 구현 전 재검증 반영: 긴 옵션 충돌, spawn 인용 팽창, 안전 판독, 독립 LIVE fixture (card t1078) | manager-spec |
| 2026-09-22 | 1.0.0 | 초판 작성 (card t1078, plan phase, Tier M) | manager-spec |

### 배경

카드 t1078의 운영자 요구: `moai codex`가 프로젝트 루트의 `CLAUDE.local.md`를 `AGENTS.local.md`와 함께 Codex 세션 지침으로 적재해야 하며, 모든 런치 경로(direct, `cli`/`app` verb, `--spawn`, `-w`, `-f` lead/agent/lane)에서 동일하게 동작해야 한다.

공식 제약(WebFetch 검증 완료, research.md 참조): Codex는 디렉터리당 최대 1개의 지침 파일만 선택하며(`AGENTS.override.md` > `AGENTS.md` > `project_doc_fallback_filenames` 순), 루트에 `AGENTS.md`가 존재하면 fallback 이름 추가 방식은 동작하지 않는다. 따라서 Codex CLI의 config override(`-c` 또는 동등한 긴 옵션 `--config`)로 `developer_instructions`를 주입한다.

이 변경은 2026-09-10 커밋 `6c647bbe2` 이후 기록된 "CLAUDE-local 상태는 Claude 전용" 계약(`CHANGELOG.md:77`, 리포 `AGENTS.md` §8, `AGENTS.md.tmpl` §8, docs-site 4-locale 페이지, 헬프 카피의 "Codex-only" 문언)을 되돌린다. Gen1(SPEC-CODEX-INIT-001 REQ-CI-008)이 원래 의도했던 CLAUDE.local.md의 Codex 도달성을 config override 채널로 회복하는 성격이며, 계약 역전이므로 기록 표면들의 정합 갱신이 요구 사항이다.

### 현재 baseline (이번 실행 직접 판독)

- `codexLocalDeveloperInstructionArgs` (`internal/cli/codex_launcher.go:114-142`)는 `AGENTS.local.md` 1개만 적재한다. 유일 호출 지점 `codex_launcher.go:495` — 모든 런치 경로가 이 funnel을 지난다(구조적 균일성).
- 리포 루트 `CLAUDE.local.md` = 61,360 bytes; `AGENTS.local.md`는 현재 부재.
- `codexLocalInstructionName = "AGENTS.local.md"` 상수(`codex_contract.go:33`), exactly-3 path table 핀(`codex_contract.go:225-227`), 헬프 핀 테스트(`codex_local_instructions_test.go:141-147`) 존재.

## 1. 요구 사항 (GEARS)

### REQ-LMD-001 — 이중 파일 적재 (Ubiquitous)

The `moai codex` launcher shall load BOTH project-root `CLAUDE.local.md` and project-root `AGENTS.local.md` as Codex developer-instruction inputs, reading each only through its own name constant in `internal/cli/codex_contract.go`.

### REQ-LMD-002 — 경로별 안전 판독 (Event-driven)

**When** the launcher reads either local-instruction path, the launcher shall obtain and validate the same opened file descriptor without following a symbolic link, shall accept the descriptor only when `fstat` identifies a regular file, and shall read the body from that validated descriptor; this rule applies independently and identically to BOTH `CLAUDE.local.md` and `AGENTS.local.md`. A symlink, FIFO, directory, socket, device, or validation/open failure shall reuse the `codexPathGuardError` / `codexModeName` diagnostic vocabulary and abort before child launch; an unsupported platform shall provide an equivalently race-resistant, fail-closed result rather than a check-then-reopen sequence.

### REQ-LMD-003 — 단일 오버라이드 합성과 출처 순서 (Ubiquitous)

The launcher shall synthesize the loaded bodies into exactly one `developer_instructions` override — the two-element argv pair `-c` and `developer_instructions=<value>` (one key=value override; never two overrides of the key), concatenating the bodies BEFORE JSON-marshal. Each contributing body shall be preceded by an observable provenance header line naming its source file (shape pinned: `<!-- source: <filename> -->` on its own line, immediately before that file's body), and the fixed order is: `CLAUDE.local.md` body first, `AGENTS.local.md` body later, so the more Codex-specific file is positioned to win on conflict.

### REQ-LMD-004 — 부재/빈 파일 매트릭스 (State-driven)

**While** each input file is evaluated independently, the launcher shall treat an EMPTY file (0 bytes after read) exactly as an ABSENT file on the per-file axis, shall inject the concatenation (per REQ-LMD-003 order and provenance shape) of whatever non-empty content exists across the two files, and shall leave the child argv unchanged when neither file yields a non-empty body — including every cell of the matrix: one file present non-empty → that file only; `CLAUDE.local.md` empty + `AGENTS.local.md` non-empty → AGENTS content injected; either file empty-or-absent while the other is non-empty → the non-empty one injected; both absent or both empty → no `-c` override at all.

### REQ-LMD-005 — 연산자 config override 충돌 거부 (Event-detected)

**When** an operator-supplied `--` tail contains any of the five Codex 0.155.1 parser-confirmed spellings targeting `developer_instructions` — `--config developer_instructions=...`, `--config=developer_instructions=...`, `-c developer_instructions=...`, `-c=developer_instructions=...`, or `-cdeveloper_instructions=...` — and the launcher has a synthesized value for the same key, the launcher shall fail with an explicit, named duplicate-override error before child launch. It shall never silently overwrite either value or emit two overrides. The same operator token with NO synthesized value is not a collision, and unrelated config keys are not collisions.

### REQ-LMD-006 — 무절단·바이트 보존·신선 판독 (Ubiquitous)

The launcher shall preserve each loaded file's full UTF-8 bytes without truncation or clipping. After JSON decoding and provenance-aware extraction, each body slice shall be byte-identical to its source fixture; provenance headers are framing and are not part of either source-body hash. Both files shall be read fresh at EVERY launch without a cross-launch content cache.

### REQ-LMD-007 — direct·spawn 실행 인자 한도 초과 fail-closed (Event-detected)

**When** either the complete JSON-escaped `developer_instructions=<value>` token for a direct launch exceeds the direct ceiling of 126,976 bytes, or the fully assembled and shell-quoted command string actually handed off by `--spawn` exceeds the independently applied spawn ceiling of 126,976 bytes, the launcher shall report the measured byte length and applicable ceiling and fail closed before any child or tmux window starts. The shared numeric default is Linux `MAX_ARG_STRLEN` 131,072 minus 4,096 bytes, but each final representation shall be measured independently after its own escaping; the margin on the pre-quoted token shall not be treated as a bound on shell-quote expansion. The named single source shall live in `internal/config/defaults.go`, and an OS execution error that still occurs within bounds shall preserve and identify the underlying error.

### REQ-LMD-008 — 구조적 경로 균일성 (Ubiquitous)

The dual-file load shall live inside the existing single funnel (`codexLocalDeveloperInstructionArgs` invoked once at `codex_launcher.go:495` in `runCodexLaunch`), so that the bare form, `cli`, `app`, `--spawn`, `-w`, and the `-f` factory roles (lead, `agent`/`agent-N`, `lane-N`) all receive the identical synthesized payload without per-path branch logic.

### REQ-LMD-009 — 문서·헬프 표면 정합 (Ubiquitous)

The launcher help copy (`codexCmd.Long`) shall distinguish `CLAUDE.local.md` as a common local instruction source shared with Claude workflows and `AGENTS.local.md` as the Codex-specific source, and shall state that `moai codex` injects both when non-empty. The pinning test shall reject inaccurate exclusivity claims; "Codex-only" remains valid when it refers specifically to `AGENTS.local.md`.

### REQ-LMD-010 — 입력 비가공·계약 테이블 불변 (Unwanted — shall not)

During launcher execution and LIVE verification, the system shall not modify `AGENTS.md` or `CLAUDE.md`, shall not create symlinks, and shall not write/rename/link either local-instruction file. User-local inputs `CLAUDE.local.md` and `AGENTS.local.md` shall remain unmodified in every phase. This runtime-input rule does not prohibit manager-docs from updating tracked descriptive text in repository `AGENTS.md` during sync. `defaultCodexInstructionRelPaths` shall remain exactly 3.

## 2. 제약 사항

- 입력 파일은 읽기 전용이다. 기록·개명·링크 금지(REQ-LMD-010; 선행 SPEC-CODEX-LAUNCHER-001 REQ-CL-013 전통 계승).
- JSON 문자열 인코딩은 유효한 TOML basic string이다 — 인코딩 방식 변경 금지(단일 토큰 유지의 전제).
- 신규 argv-overflow 가드는 리포 최초의 E2BIG 계열 방어다(`E2BIG|ErrArgListTooLong` grep 0건 — research.md §6). 상한 상수는 `internal/config/defaults.go` 단일 원천 규율을 따른다.
- `internal/template/templates/AGENTS.md.tmpl` §8은 특정 로컬 파일명을 재열거하지 않는 정확한 일반 문구로 교정한다. `moai codex status`에는 local-instruction 행을 추가하지 않는다.
- Sync-phase 문서(CHANGELOG, docs-site 4-locale, 리포 AGENTS.md §8)는 run phase에서 편집하지 않고 manager-docs에 인계한다(plan.md §F 참조).

## 3. 수용 기준 요약 (Tier M — 상세는 acceptance.md)

12건의 AC-LMD-XXX Given/When/Then 시나리오로 검증한다. 경로 동일성, 부재/빈 매트릭스, 양쪽 파일의 안전 open/fstat, `-c`/`--config` 충돌, direct·spawn 최종 표현별 초과 차단, 재시작 재판독, 헬프 정합, 독립 fixture 프로젝트의 Codex 레인 LIVE 수용을 포함한다. LIVE가 NOT_RUN이면 전체 판정은 PASS가 아니다.

## 4. Out of Scope

### Out of Scope — factory 브로커·idle wake

- Factory broker scope과 idle-wake는 t1074/t1075가 소관이다. 본 SPEC은 건드리지 않는다.

### Out of Scope — 공유 지침 파일·링크 체인 변경

- 런처 실행·LIVE 검증·run phase에서 `AGENTS.md` / `CLAUDE.md` 본문을 수정하거나 symlink 및 `@AGENTS.md` import 체인을 생성·변경하는 작업 — 금지(REQ-LMD-010). 단, sync phase의 manager-docs가 추적 문서인 리포 루트 `AGENTS.md`의 설명 문구를 실제 런처 계약에 맞게 갱신하는 작업은 이 제외 범위에 들지 않는다. 사용자 로컬 입력인 `CLAUDE.local.md`와 `AGENTS.local.md`는 모든 단계에서 수정 금지다.
- `defaultCodexInstructionRelPaths` 3-path 테이블 확장 — 금지.

### Out of Scope — Codex CLI 외부 동작 변경

- Codex CLI의 `developer_instructions` vs 파일 discovery 상호작용 자체는 외부 소프트웨어 동작이며, 본 SPEC은 주입 콘텐츠의 정확성만 보장한다(조합 관계는 LIVE acceptance에서 실측).

### Out of Scope — sync-phase 문서 편집의 run-phase 선실행

- CHANGELOG, docs-site 4-locale `advanced/codex-dual-harness.md`, 리포 `AGENTS.md` §8 갱신은 sync phase(manager-docs) 소관이다. run phase는 실행하지 않는다.

### Out of Scope — status 출력 확장

- `moai codex status`에 local-instruction 상태 행을 추가하거나 status 출력 계약을 바꾸는 작업은 포함하지 않는다.
