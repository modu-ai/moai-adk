---
id: "SPEC-CODEX-LOCALMD-001"
title: "Codex launcher loads CLAUDE.local.md alongside AGENTS.local.md as developer_instructions"
version: "1.0.0"
status: "draft"
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
| 2026-09-22 | 1.0.0 | 초판 작성 (card t1078, plan phase, Tier M) | manager-spec |

### 배경

카드 t1078의 운영자 요구: `moai codex`가 프로젝트 루트의 `CLAUDE.local.md`를 `AGENTS.local.md`와 함께 Codex 세션 지침으로 적재해야 하며, 모든 런치 경로(direct, `cli`/`app` verb, `--spawn`, `-w`, `-f` lead/agent/lane)에서 동일하게 동작해야 한다.

공식 제약(WebFetch 검증 완료, research.md 참조): Codex는 디렉터리당 최대 1개의 지침 파일만 선택하며(`AGENTS.override.md` > `AGENTS.md` > `project_doc_fallback_filenames` 순), 루트에 `AGENTS.md`가 존재하면 fallback 이름 추가 방식은 동작하지 않는다. 따라서 `-c developer_instructions=` 주입이 유일한 수단이다.

이 변경은 2026-09-10 커밋 `6c647bbe2` 이후 기록된 "CLAUDE-local 상태는 Claude 전용" 계약(`CHANGELOG.md:77`, 리포 `AGENTS.md` §8, `AGENTS.md.tmpl` §8, docs-site 4-locale 페이지, 헬프 카피의 "Codex-only" 문언)을 되돌린다. Gen1(SPEC-CODEX-INIT-001 REQ-CI-008)이 원래 의도했던 CLAUDE.local.md의 Codex 도달성을 `-c` 채널로 회복하는 성격이며, 계약 역전이므로 기록 표면들의 정합 갱신이 요구 사항이다(decision-index R1 참조).

### 현재 baseline (이번 실행 직접 판독)

- `codexLocalDeveloperInstructionArgs` (`internal/cli/codex_launcher.go:114-142`)는 `AGENTS.local.md` 1개만 적재한다. 유일 호출 지점 `codex_launcher.go:495` — 모든 런치 경로가 이 funnel을 지난다(구조적 균일성).
- 리포 루트 `CLAUDE.local.md` = 61,360 bytes; `AGENTS.local.md`는 현재 부재.
- `codexLocalInstructionName = "AGENTS.local.md"` 상수(`codex_contract.go:33`), exactly-3 path table 핀(`codex_contract.go:225-227`), 헬프 핀 테스트(`codex_local_instructions_test.go:141-147`) 존재.

## 1. 요구 사항 (GEARS)

### REQ-LMD-001 — 이중 파일 적재 (Ubiquitous)

The `moai codex` launcher shall load BOTH project-root `CLAUDE.local.md` and project-root `AGENTS.local.md` as Codex developer-instruction inputs, reading each only through its own name constant in `internal/cli/codex_contract.go`.

### REQ-LMD-002 — 경로별 구조 가드 (Event-driven)

**When** the launcher probes either local-instruction path, the launcher shall apply the per-file regular-file guard (closed-set `IsRegular` leaf judgement via `os.Lstat`, refusing symlink / FIFO / directory / socket / device) and shall reuse the `codexPathGuardError` / `codexModeName` diagnostic vocabulary from `internal/cli/codex_contract.go` — a per-file refusal shall abort the launch before any child process starts.

### REQ-LMD-003 — 단일 오버라이드 합성과 출처 순서 (Ubiquitous)

The launcher shall synthesize the loaded bodies into exactly one `developer_instructions` override — the two-element argv pair `-c` and `developer_instructions=<value>` (one key=value override; never two overrides of the key), concatenating the bodies BEFORE JSON-marshal. Each contributing body shall be preceded by an observable provenance header line naming its source file (shape pinned: `<!-- source: <filename> -->` on its own line, immediately before that file's body), and the fixed order is: `CLAUDE.local.md` body first, `AGENTS.local.md` body later, so the more Codex-specific file is positioned to win on conflict.

### REQ-LMD-004 — 부재/빈 파일 매트릭스 (State-driven)

**While** each input file is evaluated independently, the launcher shall treat an EMPTY file (0 bytes after read) exactly as an ABSENT file on the per-file axis, shall inject the concatenation (per REQ-LMD-003 order and provenance shape) of whatever non-empty content exists across the two files, and shall leave the child argv unchanged when neither file yields a non-empty body — including every cell of the matrix: one file present non-empty → that file only; `CLAUDE.local.md` empty + `AGENTS.local.md` non-empty → AGENTS content injected; either file empty-or-absent while the other is non-empty → the non-empty one injected; both absent or both empty → no `-c` override at all.

### REQ-LMD-005 — 연산자 `-c` 충돌 거부 (Event-detected)

**When** an operator-supplied `--` tail contains any token that would override the `developer_instructions` key — pinned scope: any `-c` token in the tail whose value begins with `developer_instructions=`, which is the only argv spelling that can reach the child as an override of this key — and the launcher has a synthesized value for the same key, the launcher shall fail the launch with an explicit, named error directing the operator to remove the duplicate — never silently overwrite either value, and never emit two overrides of the key. A tail token of this shape with NO synthesized launcher value (both files absent or empty) is not a collision and launches normally.

### REQ-LMD-006 — 무절단·바이트 보존·신선 판독 (Ubiquitous)

The launcher shall preserve the full UTF-8 byte content of each loaded file with no truncation and no size clipping, such that a body of 61,360 bytes (the current repo-root `CLAUDE.local.md` size) round-trips byte-identically through JSON encoding into the single override value; and shall read both files fresh at EVERY launch — each launch reads the files' current bytes from disk, with no caching of file contents across launches.

### REQ-LMD-007 — argv 한도 초과 fail-closed (Event-detected)

**When** the final encoded length of the synthesized override — pinned measure: the byte length of the complete JSON-escaped `developer_instructions=<value>` token exactly as handed to exec, after all escaping — would exceed the launcher's declared per-token ceiling, the launcher shall diagnose the overflow and fail closed BEFORE launching any child process — including before the `--spawn` path shell-quotes the whole argv into one tmux command string (`buildCodexSpawnCommand`, `codex_launcher.go:193-204`). The ceiling shall be a named constant defined once in `internal/config/defaults.go`, whose default value is derived from Linux `MAX_ARG_STRLEN` (131,072 bytes) minus a declared safety margin (margin policy pinned in plan.md §M2: 4,096 bytes, reserving room for the `-c` prefix, the shell-quote expansion on the spawn path, and kernel accounting variance).

### REQ-LMD-008 — 구조적 경로 균일성 (Ubiquitous)

The dual-file load shall live inside the existing single funnel (`codexLocalDeveloperInstructionArgs` invoked once at `codex_launcher.go:495` in `runCodexLaunch`), so that the bare form, `cli`, `app`, `--spawn`, `-w`, and the `-f` factory roles (lead, `agent`/`agent-N`, `lane-N`) all receive the identical synthesized payload without per-path branch logic.

### REQ-LMD-009 — 문서·헬프 표면 정합 (Ubiquitous)

The launcher help copy (`codexCmd.Long`) shall accurately describe the dual-file injection — naming both local-instruction files and no longer carrying the inaccurate "Codex-only" phrasing — such that its pinning test (`TestCodexLocalInstructions_DocumentedInLauncherHelp`, updated to the new expected tokens) passes against the help text alone.

### REQ-LMD-010 — 입력 비가공·계약 테이블 불변 (Unwanted — shall not)

The launcher shall not modify `AGENTS.md` or `CLAUDE.md`, shall not create symlinks, shall not write/rename/link either local-instruction file, and shall not add `CLAUDE.local.md` to `defaultCodexInstructionRelPaths` (the exactly-3-path contract table pinned at `codex_contract.go:225-227` stays exactly 3).

## 2. 제약 사항

- 입력 파일은 읽기 전용이다. 기록·개명·링크 금지(REQ-LMD-010; 선행 SPEC-CODEX-LAUNCHER-001 REQ-CL-013 전통 계승).
- JSON 문자열 인코딩은 유효한 TOML basic string이다 — 인코딩 방식 변경 금지(단일 토큰 유지의 전제).
- 신규 argv-overflow 가드는 리포 최초의 E2BIG 계열 방어다(`E2BIG|ErrArgListTooLong` grep 0건 — research.md §6). 상한 상수는 `internal/config/defaults.go` 단일 원천 규율을 따른다.
- `internal/template/templates/AGENTS.md.tmpl` §8 문구 갱신은 SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001 M4(템플릿 공개 콘텐츠에서 CLAUDE.local.md 참조 17건 제거)와 긴장한다 — 본 SPEC은 run phase에서 템플릿을 함부로 고치지 않고 decision-index R1으로 상신한다.
- Sync-phase 문서(CHANGELOG, docs-site 4-locale, 리포 AGENTS.md §8)는 run phase에서 편집하지 않고 manager-docs에 인계한다(plan.md §F 참조).

## 3. 수용 기준 요약 (Tier M — 상세는 acceptance.md)

12건의 AC-LMD-XXX Given/When/Then 시나리오로 검증한다. 6-런치-경로 동일 페이로드, precedence 충돌 fixture, 부재/빈 매트릭스, 경로별 거부, 61KiB 바이트 보존, `-c` 충돌 거부, argv 초과 fail-closed, 재시작 재판독, 헬프 정합, LIVE 수용(실제 별도 codex 세션에서 nonce 확인 — NOT_RUN이면 SPEC 전체 판정은 PASS 불가)을 포함한다.

## 4. Out of Scope

### Out of Scope — factory 브로커·idle wake

- Factory broker scope과 idle-wake는 t1074/t1075가 소관이다. 본 SPEC은 건드리지 않는다.

### Out of Scope — 공유 지침 파일·링크 체인 변경

- `AGENTS.md` / `CLAUDE.md` 본문 수정, symlink 생성, `@AGENTS.md` import 체인 변경 — 금지(REQ-LMD-010).
- `defaultCodexInstructionRelPaths` 3-path 테이블 확장 — 금지.

### Out of Scope — Codex CLI 외부 동작 변경

- Codex CLI의 `developer_instructions` vs 파일 discovery 상호작용 자체는 외부 소프트웨어 동작이며, 본 SPEC은 주입 콘텐츠의 정확성만 보장한다(조합 관계는 LIVE acceptance에서 실측).

### Out of Scope — sync-phase 문서 편집의 run-phase 선실행

- CHANGELOG, docs-site 4-locale `advanced/codex-dual-harness.md`, 리포 `AGENTS.md` §8 갱신은 sync phase(manager-docs) 소관이다. run phase는 실행하지 않는다.
