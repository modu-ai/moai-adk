---
id: SPEC-HANDOFF-NEUTRAL-001
title: "핸드오프 하네스 중립화 — Claude Code↔Codex 교차 인계 무손실"
version: "0.1.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: worker-61 (lane session, card t1273)
priority: P1
phase: "v3.0.0"
module: "internal/cli, internal/codexadapter, internal/homestate"
lifecycle: spec-anchored
tags: "handoff, codex, harness-neutral, cross-session"
tier: L
---

# SPEC — 핸드오프 하네스 중립화

## §1 배경·문제

세션 핸드오프(`moai handoff save` → factory.db `resume_handoffs`(SQLite, 상태머신 pending→claimed→consumed) → SessionStart 자동 주입)는 Claude Code 하네스에만 소비 경로가 묶여 있다. Codex 세션은 저장된 인계를 발견·소비할 방법이 없어, 교차 인계(Claude→Codex)에서 6블록 텍스트가 사용자의 수동 복사에만 의존한다. 실측(research §C): 저장 매체·CLI는 이미 하네스 중립이나 **소비가 편향** — handoff 표면 Codex 참조 0건, `--harness codex` 어댑터는 존재하나 `additionalContextEvents`가 UserPromptSubmit만 담아 SessionStart 채널을 버리고, 프로젝트 `.codex/hooks.json` 배선은 untracked 런타임 생성물이라 **워크트리 Codex 세션에는 moai 훅이 로드되지도 않는다**.

## §2 용어

- **교차 인계**: 송신 세션과 수신 세션의 하네스가 다른 핸드오프 (Claude→Codex, Codex→Claude).
- **P3 (기본 경로)**: `moai handoff show` 재출력 + 사용자 paste — 하네스 무관.
- **P1 (조건부 경로)**: Codex SessionStart additionalContext 자동 주입 — 전제 2개(워크트리 배선 시딩, 관문 b 전달 측정) 통과 시만 채택.
- **materializer**: `moai worktree new`가 새 워크트리 트리를 만드는 공용 생성 경로.

## §3 요구사항 (GEARS)

### M1 — 이 카드의 run 범위

#### REQ-HN-001 (When) — handoff show 재출력
WHEN 사용자가 `moai handoff show [--project-dir <path>]`를 호출하면, 시스템은 factory.db의 미소개 pending row(`ReadPending`)를 우선 소스로, 없으면 소비 이력(`status='consumed'` 최신, `consumed_at` 내림차순)을 소스로 저장된 6블록 Body를 **바이트 그대로** stdout에 재출력한다. 둘 다 없으면 exit 1과 "저장된 핸드오프 없음" 안내를 낸다.

#### REQ-HN-002 (Ubiquitous) — show 무상태
`moai handoff show`는 어떤 경우에도 row의 상태 전이·수정을 일으키지 않는다(읽기 전용). 호출 반복은 항상 같은 출력을 낸다 (auto-inject 흐름과의 동시 실행 안전).

#### REQ-HN-003 (When) — --json 기계 판독
WHEN `--json` 플래그가 주어지면, 시스템은 PendingRecord 전체를 JSON으로 stdout에 낸다 (출처 pending/consumed 포함).

#### REQ-HN-004 (Where) — 4-로케일 헤더
사람 판독 출력의 헤더(출처·spec·phase·언어·저장시각·소비 여부)는 저장된 `ConversationLanguage`(ko/en/ja/zh)를 따라 현지화한다. 6블록 본문은 현지화하지 않는다.

#### REQ-HN-005 (When) — materializer 시딩
WHEN `moai worktree new`가 새 워크트리를 생성하면, 시스템은 그 트리에 `.codex/hooks.json`을 moai 소유 항목만 담아 시딩한다 (`internal/codexwiring` wire.go의 생성 로직 재사용). 파일이 이미 있으면 **건드리지 않는다** (사용자 소유 항목 보존).

#### REQ-HN-006 (When) — 런처 진입 보완
WHEN `moai cc -w`/`moai codex -w`가 기존 워크트리에 진입할 때 `.codex/hooks.json`이 부재하면, 시스템은 REQ-HN-005와 같은 멱등 시딩으로 채운다.

#### REQ-HN-007 (Ubiquitous) — 시딩 fail-open
시딩 실패는 워크트리 생성·진입을 막지 않는다 (stderr 진단 + 진행). 시딩은 부가 기능이지 관문이 아니다.

#### REQ-HN-008 (When, 조건부) — SessionStart 채널 매핑
WHEN LIVE 관문 (b) — 격리 CODEX_HOME·스크래치 프로젝트에서 Codex SessionStart의 additionalContext가 세션에 도달함이 관측되면, 시스템은 `internal/codexadapter/output.go`의 `additionalContextEvents`에 `EventSessionStart`를 추가한다. 관문이 실패하면 이 REQ는 기각으로 종결하고 P3 단독으로 M1을 닫는다 (기각 사유 기록).

#### REQ-HN-009 (Ubiquitous) — LIVE 검증 격리·상한
LIVE 검증은 임시 CODEX_HOME과 스크래치 프로젝트에서만 실행한다 (운영자 `~/.codex/config.toml`·primary `.codex/hooks.json` 무접촉). 상한: 관문당 실행 3회·벽시계 30분. 결과(관문별 판정·관측 출력)는 카드 증거 경로에 기록한다.

#### REQ-HN-010 (Ubiquitous) — M1 렌더 형식 불변
M1은 6블록 본문·Directives 렌더 형식을 변경하지 않는다 (ultrathink 등 하네스 고유 키워드 제거는 M2 — SSOT 문서와 렌더 코드 동시 변경).

#### REQ-HN-011 (Ubiquitous) — 방향 중립 문서화
Codex→Claude 방향( Codex 세션이 `moai handoff save --stdin`으로 남긴 인계를 Claude 인젝터가 소비)이 코드 변경 없이 성립함을 design에 문서화한다.

### 후속 마일스톤 (이 카드 run 범위 밖 — 병합 의존)

- **M2 (t1175 병합 뒤)**: 6블록 공통 본문에서 ultrathink 등 하네스 고유 키워드 제거 — SSOT 문서(session-handoff.md·-examples.md·output-styles §8)와 렌더 코드(handoff.go·pending.go·handoff_inject_render.go)를 **같은 마일스톤에서** 변경.
- **M3 (t1243 병합 뒤)**: AGENTS.md 소비 계약 문구 배치(§D5 초안 제출) 및 옵션 C(hooks.json 템플릿화 분리 규칙) 후속 카드 여부 판정.

## §4 제약

- Template-First: M1은 템플릿 배포 파일을 만들지 않는다 (show는 CLI 동사, 시딩은 런타임 생성물 — 둘 다 `internal/template/templates/` 무접촉).
- 16-프로그래밍-언어 중립·4-로케일(en/ko/ja/zh) 표 유지 — show 헤더 현지화는 기존 `handoffLocaleStrings` 관례 준수.
- 검증은 레인-로컬 대상 테스트만 (`go test ./internal/cli/ -run …`, `./internal/codexadapter/`, `./internal/homestate/` 관련 그룹). 전체 스위트는 CI.
- 커밋 메시지에 카드 id(t1273) 포함, push 금지(리드 일괄).

### §5.1 Out of Scope

- AGENTS.md.tmpl 수정 (M3)
- 6블록 렌더 형식·키워드 변경 (M2)
- hooks.json 템플릿화(옵션 C)·Codex 측 사용자/moai 소유 분리 규칙 설계 (후속 카드 후보)
- Codex↔Claude 실시간 메시징(브로커) — 이미 존재, 별개 계층(research §C.5)
- 세션-하네스 통합 resume(Codex가 Claude 세션을 재개하는 것) — 불가 확인(research §A.3), 대상 아님

## HISTORY

- 2026-09-26 최초 작성 (card t1273, worker-61). 리드 B안+조정 승인 구조 반영: M1(본 run) / M2·M3(병합 의존 후속).
