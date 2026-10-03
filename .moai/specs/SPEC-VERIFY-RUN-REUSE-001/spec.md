---
id: SPEC-VERIFY-RUN-REUSE-001
title: "moai verify run — 같은 작업 트리 상태에서 같은 테스트 명령 재실행 억제 (run-or-reuse 단일 동사)"
version: "0.2.1"
status: draft
created: 2026-10-03
updated: 2026-10-03
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/verify"
lifecycle: spec-anchored
tags: "verify, snapshot-reuse, card-t1452, test-rerun-suppression, lane-verification"
tier: S
---

# SPEC-VERIFY-RUN-REUSE-001 — 같은 트리·같은 명령 재실행 억제

- 카드: t1452 (범위 (c) 한정), 브랜치 `WT-merge-window-hold-time`
- 근거 판정서: `.moai/reports/t1452/verdict.md` (겹침 판정 + (c) 조사 결과)
- 범위 (a)(b)(병합 창 대기·보유 시간)는 SPEC-MERGE-WINDOW-QUEUE-001 / SPEC-CANDIDATE-CI-001 이 흡수했으므로 이 SPEC 의 범위가 아니다.

## HISTORY

| 버전 | 일자 | 내용 |
|---|---|---|
| 0.1.0 | 2026-10-03 | 최초 작성 (draft). (c) 한정 Tier S. REQ 8·AC 8 (Tier S 상한). |
| 0.2.0 | 2026-10-03 | plan-audit iter1 FAIL 0.76 반영: D1 AC 선택자 실행 개수 단언 + RED-now 칸 기록, D2 포터블 self-exec 헬퍼, D3 `--tool-version-cmd` 반복 플래그·직접 exec·자체 타임아웃·실패 시 unbound, D4 재사용=이전 관측 문구(REQ-008, AC-007), D6 프로세스 그룹 종료, D8 `--timeout` 기본값·`recorded_at` 의미, D12 SPEC id 만 인용, D13 잔여 위험 3건. |
| 0.2.1 | 2026-10-03 | plan-audit iter2 N1 반영: AC-007 grep 패턴을 `-e` 로 전달(`--env` 옵션 오인 수정) + RED-now 실측(exit 1)·양성 대조(exit 0), AC-005 타임아웃 테스트를 서브테스트로 분리(PASS 줄 4), REQ-008 에 "재사용 결과는 Claim 이 아니라 Gap" 명시. |

## §A 배경과 문제

레인이 같은 트리 상태에서 같은 테스트 명령을 반복 실행한다. 카드 본문 수치(재측정하지 않은 인용): go test 711회 중 동일 명령 반복 약 286회, 한 레인이 30분짜리 `./internal/cli/...` 를 2회 실행.

저장·키·신선도 부품은 이미 있다 (develop `2b9e4a4d0` 에서 읽음):

- `internal/verify/key.go` `Key()` = HEAD SHA + `:` + 해시(`status --porcelain=v2`, `diff HEAD`, 추적되지 않는 파일 내용) — 트리 상태 키.
- `internal/verify/receipt.go` — `CheckReceipt` 가 Head·TreeDigest·ConfigDigest·Command·ToolVersion 5필드 일치 + TTL 로 재사용을 판정하고, `ConfigDigest(map)` 헬퍼가 있다.
- `internal/verify/schema.go` `FindCommand` 는 바이트 단위 명령 일치, `internal/verify/store.go` `RecordCheck` 는 키별 잠금 + 원자적 rename.
- CLI 는 `moai verify record`(기록)와 `moai verify check --key-current`(신선도 질의)의 수동 두 단계뿐이고, `check` 는 명령을 받지 않는다. "실행하거나 재사용" 단일 동사가 없다. 문서화된 소비자는 sync Stop 훅·sync-audit-4dim·sync-auditor 뿐(`.claude/rules/moai/workflow/snapshot-consumer-contract.md`)이며 레인의 카드 내 반복 `go test` 는 이 경로를 쓰지 않는다.

위험(설계 대상): 트리 키는 환경·도구 입력(GOFLAGS, GOOS, 툴체인 버전)을 덮지 않는다 — 키가 같아도 결과가 다를 수 있다 (t1413 sync 게이트 캐시 키 교훈의 계열).

Tier S 근거: 신규 코드는 결정 함수·동사·문장 한 개로 얇고 기존 부품(`Key`/`CheckReceipt`/`RecordCheck`)을 재사용한다. 파일 수(7)와 acceptance.md 보유는 Tier S 지침(< 5 파일, 2 산출물)에서 벗어나며, 리더 지시(Tier S 유지, acceptance.md 포함)에 따른 선언된 이탈이다. 이탈이 문제면 Tier M(통과선 0.80)으로 올린다.

최소 해법: 기존 부품 위에 단일 동사 `moai verify run` 을 얹고, 레인 검증 규율 문장 하나를 더한다. 새 저장소는 만들지 않는다.

## §B 요구사항 (GEARS)

- **REQ-VRR-001** (Event) — **When** `moai verify run [--check-id <id>] [--ttl <d>] [--env NAME,...] [--tool-version-cmd <elem>]... [--tool-version-timeout <d>] [--timeout <d>] -- <command...>` is invoked and a recorded entry for the current tree key satisfies every reuse condition of REQ-VRR-002, the verb shall not execute the command, shall print a reuse notice on stderr naming the recorded key, `recorded_at`, and `duration_ms`, and shall exit 0.
- **REQ-VRR-002** (Ubiquitous) — The verb shall reuse a recorded entry only when ALL hold: (1) the stored key equals the current key (`Key()`); (2) `FindCommand` matches the canonical command string of REQ-VRR-005 byte for byte; (3) the entry's exit code is 0 and its verdict is `pass` — an entry with a non-zero exit code is never reused; (4) the entry's `config_digest` equals the digest of REQ-VRR-006; (5) the entry's `tool_version` equals the tool identity of REQ-VRR-006; (6) `recorded_at` is within the TTL (`--ttl`, default `verify.DefaultTTL`). The comparison shall be performed by `verify.CheckReceipt`.
- **REQ-VRR-003** (Event) — **When** any reuse condition of REQ-VRR-002 fails, the verb shall execute the command with inherited stdin/stdout/stderr in the project root, bounded by `--timeout` (default 60m), and shall exit with the command's own exit code; **when** `--timeout` elapses it shall terminate the command's whole process group on platforms that support one (Unix) and only the direct child elsewhere (Windows, residual risk §D) and exit 124, and **when** the command cannot be started it shall exit 127, in both cases without recording.
- **REQ-VRR-004** (Event) — **When** an executed command finishes and the tree key computed before the execution equals the key computed after it, the verb shall record one entry through `verify.RecordCheck` under the pre-execution key, carrying check id, canonical command, exit code, duration, config digest, tool version, `recorded_at` (the completion time of the command, the instant the TTL is measured from), and verdict (`pass` for exit 0, else `fail`); **when** the two keys differ, the verb shall not record, shall print a one-line notice on stderr, and shall still exit with the command's exit code.
- **REQ-VRR-005** (Ubiquitous) — The canonical command string shall be the argv elements joined by a single space, an element that is empty or contains whitespace or a quote being rendered with `strconv.Quote`; the verb shall exec argv directly, without a shell. A caller needing shell syntax passes it as one element (`-- sh -c '<script>'`).
- **REQ-VRR-006** (Ubiquitous) — The config digest shall be `verify.ConfigDigest` over one input `env:NAME` per `--env` name, whose value distinguishes an unset variable from a variable set to the empty string (no `--env` yields the digest of the empty input set). **Where** `--tool-version-cmd` is given, it is a repeatable flag whose occurrences, in order, are the argv elements of the tool-identity command (first element the program; no whitespace splitting, no shell — the same direct exec as REQ-VRR-005); the verb shall run it once in the project root, bounded by `--tool-version-timeout` (default 30s), and the tool identity shall be its stdout alone, whitespace-trimmed (stderr is discarded); **where** the flag is absent the identity shall be the fixed marker `unversioned`. **When** the tool-version command cannot start, exits non-zero, times out, or prints nothing, the tool identity is unbound: the verb shall neither reuse nor record, shall print the cause on stderr, and shall execute the command, exiting with the command's exit code.
- **REQ-VRR-007** (Ubiquitous) — The verb shall use the existing snapshot store (`.moai/state/verify/snapshots/`, resolved through `verifyStoreRoot`) and shall create no new store, file format, or schema field. Entries written by `moai verify record` carry no `config_digest`/`tool_version`, so `CheckReceipt` reads them as unbound and the verb never reuses them. **When** key computation or snapshot I/O fails, the verb shall execute the command without reuse and without record, print the failure on stderr, and exit with the command's exit code (fail-open: the failure mode is a re-execution, never a fabricated hit).
- **REQ-VRR-008** (Ubiquitous) — The lane verification doctrine (`AGENTS.md` §4 and its template mirror `internal/template/templates/AGENTS.md.tmpl`) shall carry one rule sentence, each of whose key phrases sits on a single line, telling lanes to run repeated verification through `moai verify run`, stating that a reuse is bounded by the TTL and by the bound `--env` and tool identity, and stating that a reuse is a prior observation (not one made in this run): a verdict citing it names the key and `recorded_at` from the reuse notice and lists "output not re-observed" under Gaps — a reused result is reported as a Gap, not a Claim — and a claim needing verbatim output runs the command directly. This reconciles the verb with the observed-in-this-run rule of `AGENTS.md` §1.

## §3 인수 기준 (Tier S — 색인; 시나리오 본문은 acceptance.md)

| AC | 검증 대상 REQ | 요지 |
|---|---|---|
| AC-VRR-001 | REQ-VRR-001, REQ-VRR-002 | 적중: 같은 명령 N회 반복 시 실행 1회 (이후 실행 0회) |
| AC-VRR-002 | REQ-VRR-002, REQ-VRR-005 | 트리 변경·명령 바이트 차이 시 미스 |
| AC-VRR-003 | REQ-VRR-002 | 이전 비정상 종료 기록·TTL 초과 시 미스 |
| AC-VRR-004 | REQ-VRR-002, REQ-VRR-006 | env 다이제스트·도구 식별 변경 시 미스 (unset ≠ 빈 값) |
| AC-VRR-005 | REQ-VRR-003 | 종료 코드 패스스루, 124/127 |
| AC-VRR-006 | REQ-VRR-004, REQ-VRR-007 | 기록 규칙: 실행 중 트리 이동·수동 기록 항목·저장소 오류 fail-open |
| AC-VRR-007 | REQ-VRR-008 | 교리 문장이 AGENTS.md 와 템플릿 미러에 존재 |
| AC-VRR-008 | §C 제약 | 파일 범위: t1479 표면 파일 무변경 |

## §C 제약

- 파일 범위는 plan.md §B 목록으로 한정한다. 아래 파일은 다른 진행 카드(t1479) 소유이므로 이 SPEC 의 파일 목록에 들어올 수 없다: `internal/kanban/**`, `internal/cli` 의 integration*/factory_complete*/factory_merge*/factory_card*, `internal/homestate/**`, `internal/factorylane/**`, `AGENTS.local.md`, `.claude/rules/local/gitflow-lane-protocol.md`, `kanban-dispatch-mechanics.md`.
- CLI 표면은 AskUserQuestion 을 호출하지 않는다 (서브에이전트 경계). 종료 코드와 stderr/stdout 만 사용한다.
- 코드 주석은 영어 (`language.yaml` code_comments).

## §D 잔여 위험 (명시)

키가 덮지 않는 입력은 `moai verify run` 이 막을 수 없다. 사용자가 `--env`·`--tool-version-cmd` 를 주지 않은 입력이 바뀌면 같은 키·같은 명령에서 오적중이 가능하다.

1. `--env` 에 나열하지 않은 환경변수 (GOFLAGS, GOOS, GOARCH, CGO_ENABLED, GOENV 등). 호출자가 나열해야 한다.
2. 툴체인 버전 — `--tool-version-cmd` 를 주지 않으면 `unversioned` 로 묶여 컴파일러 교체가 보이지 않는다 (예: `--tool-version-cmd go --tool-version-cmd version`).
3. 작업 트리 밖 또는 gitignore 된 입력 — 키는 gitignore 된 파일 변경을 보지 않는다. 캐시·테스트 데이터·외부 서비스·시각 의존 테스트의 변화는 보이지 않는다.
4. 비결정(flaky) 테스트 — 한 번의 통과가 TTL 안에서 재사용된다.
5. 타임아웃 종료 범위 — Unix 는 프로세스 그룹 전체를 종료하지만 Windows 는 직접 자식만 종료하므로, `-- sh -c` 류 래퍼 아래의 손자 프로세스(예: `go test` 바이너리)가 남아 AGENTS.md §4 가 금지하는 배경 부하가 될 수 있다 (작업 객체 도입은 범위 밖).
6. 저장소 위치 — 스토어는 기본 체크아웃으로 해석되고(`auditreceipt.StoreRoot`) 키에 작업 디렉터리 경로가 없으므로, 레인 A 트리에서 기록된 결과가 같은 키의 다른 경로 레인 B 트리에서 재사용될 수 있다. 경로 의존 테스트는 이 동사로 재사용하지 않는다.
7. 다른 소비자 — `verify check --key-current` 소비자(sync Stop 훅·sync-audit-4dim·sync-auditor)는 `config_digest` 와 종료 코드를 보지 않으므로 `verify run` 항목(`fail` 포함)을 env 바인딩 없이 볼 수 있다. 소비자 계약은 이 SPEC 이 바꾸지 않는다.
8. 키를 움직이는 명령 — 무시되지 않는 untracked 파일을 쓰는 명령은 키를 바꿔 REQ-VRR-004 에 따라 조용히 기록되지 않으므로 절대 재사용되지 않는다 (안전한 방향의 미적중).
9. 재사용 = 이전 관측 — 적중 시 출력이 다시 관측되지 않는다 (AGENTS.md §1). REQ-VRR-008 문장이 인용 방식을 고정한다.
10. 키 다이제스트는 16 hex 접두라 이론상 충돌 가능성이 있다 (기존 `Key()` 특성, 이 SPEC 이 바꾸지 않음).

완화: TTL 기본 10분, 적중 시 stderr 에 키·기록 시각을 항상 표기, 의심 시 `--ttl 1ns` 또는 직접 실행으로 우회 가능. 레인 규율 문장(REQ-VRR-008)이 이 경계를 명시한다.

## §F 제외 사항

### Out of Scope — t1479 표면과의 접점

- 병합 창 게이트(`moai integration`, `moai factory merge/complete/card`)가 이 동사로 재측정 재사용을 소비하는 배선. t1479 (SPEC-MERGE-WINDOW-QUEUE-001) 착지 후 별도 변경으로 다룬다.
- `.claude/rules/local/gitflow-lane-protocol.md`, `kanban-dispatch-mechanics.md` 검증 절의 문장 추가.
- `internal/kanban/**`, `internal/homestate/**`, `internal/factorylane/**`, `AGENTS.local.md` 변경.

### Out of Scope — 이 동사가 하지 않는 일

- 결과 출력(stdout/stderr 본문)의 저장·재생 — 적중 시 명령 출력은 다시 나오지 않는다.
- 새 저장소, 새 스키마 필드, MCP 도구 추가 (`verify_snapshot` 변경 없음).
- `go test` 인자 자동 분석(패키지별 부분 재사용, 테스트 캐시 `-count` 판정) — 명령은 불투명 바이트열로 취급한다.
- 환경변수·툴체인 자동 탐지 — 호출자가 `--env`/`--tool-version-cmd` 로 명시한다.
- `moai verify check` 의 동작 변경, 기존 `record` 항목의 재사용.
- 범위 (a)(b) 병합 창 대기·보유 시간 단축 (SPEC-MERGE-WINDOW-QUEUE-001 / SPEC-CANDIDATE-CI-001 소관).

## §H 참조

- `.moai/reports/t1452/verdict.md`
- `internal/verify/{key,receipt,schema,store,freshness}.go`, `internal/cli/verify.go`, `internal/cli/verify_receipts.go`
- `.claude/rules/moai/workflow/snapshot-consumer-contract.md`
- 교훈: t1413 sync 게이트 캐시 키 (env/tool 축 누락)
