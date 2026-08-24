---
id: SPEC-CODEX-INIT-001
title: "Codex 미배선 프로젝트 초기화 — 생성기 호출 제안과 AGENTS.md 지시 계약"
version: "0.2.0"
status: draft
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: P1
phase: "v3.2 target"
module: internal/cli
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "codex, init, wiring, agents-md, instruction-contract, dual-harness"
depends_on: [SPEC-CODEX-LAUNCHER-001]
related_specs: [SPEC-CODEX-WIRING-001, SPEC-CODEX-DUAL-AGENTS-001]
---

# SPEC-CODEX-INIT-001 — Codex 미배선 프로젝트 초기화

## HISTORY

| 버전 | 날짜 | 변경 |
|---|---|---|
| 0.1.0 | 2026-08-24 | SPEC-CODEX-LAUNCHER-001 에서 분리 (운영자 판정, 카드 t197). 런처 SPEC 이 Tier M 예산 16/16 에 닿아 AC 를 통합해야 했고, 그 통합이 곧 커버리지 삭감이었다 — 4차 감사의 시험가능성 0.50 정체가 그 신호였다. 이동분: REQ-CL-015/016 + AC-CL-015/016. 이동하며 4차 감사 지적 F2(상태 × 동사 교차곱)·F3(로컬 파일 도달성·멱등)을 풀린 예산으로 흡수 |
| 0.2.0 | 2026-08-24 | 최초 감사 FAIL 0.75 지적 반영. 4건 중 3건이 한 형태 — "호출/존재는 시험하는데 **결과 사용·실패 경로**는 시험하지 않는다" — 라서 개별 수정 대신 그 형태를 AC 전반에 1회 적용했다 (런처가 3라운드에 걸쳐 sentinel 전파·음성 케이스·횟수 판정으로 해소한 패턴의 이식). REQ-CI-001 을 '호출' 에서 '반환값에 따라 행동' 으로 재기술, 상태 5종 → **6종**(`invalid` 포함), 실패 경로(REQ-CI-010)·프롬프트 무발행(REQ-CI-009)·경로 봉쇄(REQ-CI-011) 신설 |

## §A. 분리 내역 (무엇이 어디서 왔나)

| 원 위치 | 새 위치 | 상태 |
|---|---|---|
| `SPEC-CODEX-LAUNCHER-001` REQ-CL-015 (미배선 시 초기화 제안) | REQ-CI-001 ~ REQ-CI-004 | 확장 — 동사별·상태별로 쪼갬 |
| `SPEC-CODEX-LAUNCHER-001` REQ-CL-016 (지시 계약) | REQ-CI-005 ~ REQ-CI-008 | 확장 — 도달성·멱등·로컬 파일을 분리 |
| `SPEC-CODEX-LAUNCHER-001` AC-CL-015 | AC-CI-001 ~ AC-CI-004 | 확장 — 5상태 × {cli, app} 교차곱 |
| `SPEC-CODEX-LAUNCHER-001` AC-CL-016 | AC-CI-005 ~ AC-CI-009 | 확장 — import 형태 고정, sentinel 도달성, 2회 실행 바이트 비교 |
| (신규, 최초 감사 지적) | REQ-CI-009 ~ REQ-CI-011 + AC-CI-010 ~ AC-CI-011 | 프롬프트 무발행 · 실패 경로 · 경로 봉쇄 |

런처 SPEC 에 남는 것: **배선 상태를 읽어 보고하는 것**(REQ-CL-006, 리드아웃). 이 SPEC 이 맡는 것: **그 상태에서 무엇을 할 것인가**(제안·초기화·계약 확보). 판정 로직은 한 곳에서만 산다 — 이 SPEC 은 런처의 배선 상태 판정을 **소비** 하고 재구현하지 않는다.

## §B. 측정 전제

> 근거: `.moai/reports/t197/` — `probe.sh`(자기완결 측정 스크립트), `probe-output.txt`(무편집 전사본), `measurement.md`(전사본 해석). 아래 각 항목은 전사본의 줄 범위를 가리킨다.

### §B.1 생성기는 이미 있다

`moai init --agent claude|codex|both` 가 `.codex/hooks.json` · `.codex/config.toml` 을 만든다 (전사본 L234-236, L211-214). 에이전트 TOML 11종은 템플릿에 있다 (L216-232). 이 SPEC 은 그 생성기를 **부를** 뿐이다.

### §B.2 이 저장소 자체가 미배선이다

`ls -a .codex` → rc 1 (전사본 L246-248). 배선이 깔린 프로젝트가 예외이고 미배선이 기본 상태다.

### §B.3 지시 계약의 선례

이 저장소의 `CLAUDE.md` 는 `@AGENTS.md` 한 줄로 `AGENTS.md` 를 가져온다 (t82 에서 정본화). Codex 는 `AGENTS.md` 를, Claude 는 `CLAUDE.md` 를 읽으므로 원본을 하나로 두고 나머지가 가리키는 형태다.

## §C. 문제 진술

t88(M4)이 배선 생성기를 만들었지만, 미배선 프로젝트에서 Codex 를 띄우려는 사람에게 그 생성기로 가는 길이 없다. 그대로 기동하면 훅이 하나도 붙지 않은 채 세션이 열리고 — 조용히 잘못된 상태다. 그리고 배선이 깔려도 두 하네스가 **서로 다른 지시 파일** 을 읽으면 Codex 에서의 개발이 Claude 에서의 개발과 달라진다.

## §D. 범위

### 포함

- 미배선 판정 시 초기화 제안 (기동 전)
- 수락 시 기존 `moai init --agent codex` 생성기 호출
- `AGENTS.md` ↔ `CLAUDE.md` 지시 계약 확보 (기존 내용 보존, 멱등)
- `CLAUDE.local.md` 의 두 하네스 도달성

### Out of Scope (제외)

- 배선 **내용** 의 결정 — 무엇을 어떤 형태로 까는지는 SPEC-CODEX-WIRING-001 소관
- 배선 **상태 판정 로직** — REQ-CL-006(런처)이 정의하고 이 SPEC 은 소비만 한다
- 이미 깔린 배선의 표류 수리 — `moai doctor` 몫
- 지시 파일 **내용** 의 작성 — 계약(연결)만 확보하고 본문은 생성기·템플릿 소관

## §E. 요구사항 (GEARS)

### E.1 제안과 기동 차단

- **REQ-CI-001** — Where a launch verb runs in a project whose Codex wiring is incomplete, the system shall offer to initialize before launching. It shall obtain that judgement by calling the readout's wiring classifier and acting on the state that call returns — calling it and then deciding by other means is the same defect as defining a second test.
- **REQ-CI-002** — The offer shall bind both launch verbs identically; a verb that launches into an incomplete project while its sibling gates is a defect, not a variation.
- **REQ-CI-003** — Where the operator declines, the system shall write nothing and shall not launch — entering an unwired project is the failure the offer exists to prevent, and proceeding anyway would make the offer decorative.
- **REQ-CI-004** — Where the operator accepts, the system shall invoke the existing generator with the codex agent selection exactly once, and shall contain no path that writes a wiring file itself.
- **REQ-CI-009** — Where the environment cannot prompt, the system shall report the incomplete state and its remedy and exit without launching, issuing no prompt at all — a prompt that no one can answer hangs the automation the check was meant to protect.
- **REQ-CI-010** — Where the generator fails, or the instruction contract cannot be written, the system shall not launch: it shall report the failure with its remedy and exit non-successfully, and shall not proceed to the contract step after a failed generator. Launching on a failed initialization delivers exactly the unwired session this SPEC exists to prevent.

### E.2 지시 계약

- **REQ-CI-005** — Initialization shall ensure the project carries an `AGENTS.md` and a `CLAUDE.md` linked by a single import directive, so both harnesses resolve one instruction source.
- **REQ-CI-006** — Where either file already exists, its content shall be preserved byte-for-byte and only the missing link added; initialization shall never rewrite instruction content a person authored.
- **REQ-CI-007** — Initialization shall be idempotent: running it twice shall leave every instruction file byte-identical to its state after the first run.
- **REQ-CI-008** — Where the project carries a local, uncommitted instruction file, initialization shall make its content reachable from both harnesses through the import chain, referenced from exactly one place so it is not loaded twice; where no such file exists, nothing shall reference it.
- **REQ-CI-011** — Initialization shall treat an instruction path that is not a regular file inside the project as out of bounds: it shall neither read, write, nor import through such a path, and shall report the condition instead. A repository can make any instruction filename a symlink pointing outside the project, and following one would let a checked-out repository publish arbitrary host files into both harnesses, or write through the link to a location the project does not own.

## §F. 비기능 요구

- 초기화는 프로젝트 트리 밖(사용자 홈, `CODEX_HOME`)에 쓰지 않는다. 경로 봉쇄는 스냅샷이 아니라 **구조** 로 건다 — 심링크는 스냅샷 판정을 통과하면서 밖에 쓰기 때문이다 (REQ-CI-011).
- 요구 번호는 관심사별로 묶여 있어 항번이 연속하지 않는다 (E.1 에 001-004·009·010, E.2 에 005-008·011). 나중에 늘어난 요구를 성격이 같은 절에 붙인 결과이며, 참조 안정성을 위해 재번호하지 않는다.
- 크로스 플랫폼: 경로 조합은 `filepath` 기반이며 구분자를 하드코딩하지 않는다.

## §G. 성공 판정

미배선 프로젝트에서 `moai codex cli` 와 `moai codex app` 이 각각 초기화를 제안하고, 거절하면 아무것도 쓰이지 않고 기동도 없으며, 수락하면 생성기가 한 번 돌아 배선과 지시 계약이 함께 생긴다. 같은 명령을 다시 돌려도 파일은 바이트 동일하다.
