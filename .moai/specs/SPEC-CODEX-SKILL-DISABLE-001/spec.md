---
id: SPEC-CODEX-SKILL-DISABLE-001
title: "codex에서만 특정 스킬을 끄는 손 — 사용자가 이름으로 지목하고, moai가 파일 모양 게이트를 발행한다"
version: "0.1.0"
status: draft
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tier: M
tags: "codex, skills, skills-config, disable, opt-in, idempotent-merge, card-t502"
related_specs: [SPEC-CODEX-SKILLCONFIG-SHAPE-001, SPEC-CODEX-GHOST-SKILLS-PRUNE-001, SPEC-CODEX-SKILL-PATH-001, SPEC-CODEX-SKILLS-CANONICAL-001]
---

# SPEC-CODEX-SKILL-DISABLE-001 — codex 전용 스킬 비활성화 발행기

카드: **t502** (Class C, cycle_type=tdd)

## HISTORY

- 2026-09-07 (plan-phase, v0.1.0) — 카드 t502. 선행 실측 `.moai/reports/t504/skills-config-path-shape.md`(codex-cli 0.153.4, `CODEX_HOME` 격리 랩)의 채택된 행렬을 설계 제약으로 받아 작성. 쌍둥이 쓰기 경로 `internal/cli/codex_skills_prune.go`(t506)의 모양을 그대로 따른다.

## §A 배경

Claude Code와 codex는 같은 스킬 집합을 본다. 그런데 **한쪽에서만 끄고 싶을 때**가 있다 — Claude Code에서는 계속 쓰되 codex 세션에서는 노출을 빼고 싶은 스킬. 지금 이 저장소에는 그 수단이 **없다**. 사용자는 `~/.codex/config.toml`을 손으로 편집해야 하고, 그 파일은 codex가 스스로 쓰는 파일이라 손 편집은 위험하다.

t504가 그 수단이 실재함을 실측했다. `[[skills.config]]` 배열-테이블 엔트리는 **노출을 만들지는 못하지만**(V5/V7 마커 0) **기존에 로드되던 스킬을 끄기는 한다** — 단 조건이 둘이다.

| 실측 사실 | 셀 근거 | 설계에 미치는 구속 |
|---|---|---|
| `enabled` 는 파서의 **필수 필드**. 누락 엔트리 하나면 codex 시작이 rc=1로 죽는다 | V1/V2/V3 (`missing field \`enabled\``) | 발행하는 모든 엔트리는 `enabled` 를 **명시**해야 한다. 위반 시 사용자 codex **전면 장애** |
| 게이트는 **파일 모양**(`.../SKILL.md`)에만 묶인다. 디렉터 모양은 `enabled=false` 여도 무효 | D1f(마커 0) vs D1d(마커 1) | 발행 경로는 반드시 `SKILL.md` 파일 경로 |
| 이 키는 노출을 **생성하지 못한다** | V5·V7 마커 0 | 이 기능을 등록/프로비저닝 통로로 서술하거나 시험해서는 안 된다 |

## §B 이 SPEC이 만드는 것

사용자가 **스킬 이름**으로 지목하면, moai가 그 이름을 `SKILL.md` 절대 경로로 해석해 사용자 계층 `~/.codex/config.toml` 에 `enabled = false` 엔트리 하나를 발행한다. 발행은 **기본 비활성**이고, **명시 opt-in** 이며, **멱등**이고, **비파괴적**이다.

## §C 요구사항 (GEARS)

### C.1 기본 비활성 — 요청하지 않은 쓰기는 없다

- **REQ-CSD-001** (Ubiquitous) — moai는 사용자가 비활성화 동사를 **직접 호출한 경우가 아니면** `~/.codex/config.toml` 에 `[[skills.config]]` 를 쓰지 않아야 한다(shall not).
- **REQ-CSD-002** (Ubiquitous) — 이 기능은 프로젝트 설정 키로 구동되지 않아야 한다(shall not). 사용자의 HOME에 대한 쓰기는 그 순간 사용자가 요청한 것이어야 하며, 프로젝트 설정 파일이 대신 요청할 수 없다.
- **REQ-CSD-003** (Ubiquitous) — 이 동사는 사용자가 쓰기 대상 계층을 **호출문에서 명시**했을 때에만 실행되어야 한다.

### C.2 이름 해석 — 지목은 이름으로, 발행은 경로로

- **REQ-CSD-010** (Event-driven) — **When** 사용자가 스킬 이름을 지목하면, 해석기는 그 이름을 실존하는 `SKILL.md` 절대 경로 하나로 해석해야 한다.
- **REQ-CSD-011** (Event-driven) — **When** 해석이 실패하면(어떤 후보 루트에도 해당 `SKILL.md` 가 없으면), 동사는 거절하고 **아무것도 쓰지 않아야** 한다. 죽은 경로 엔트리를 발행하는 것은 t506이 청소 중인 유령 49건을 한 건 더 만드는 일이다.
- **REQ-CSD-012** (Event-driven) — **When** 한 이름이 둘 이상의 후보 루트에서 해석되면, 동사는 모호성을 보고하고 사용자가 경로를 확정할 때까지 쓰지 않아야 한다.
- **REQ-CSD-013** (Ubiquitous) — 해석된 경로는 **파일 모양**(`SKILL.md` 로 끝나는)이어야 한다. 디렉터 모양 경로를 발행해서는 안 된다(shall not) — 실측상 게이트가 묶이지 않아 조용히 무효인 엔트리가 된다.

### C.3 발행 내용 — 스키마 하드 에러를 만들지 않는다

- **REQ-CSD-020** (Ubiquitous) — 발행하는 모든 엔트리는 `path` 와 `enabled` 두 키를 **모두** 가져야 한다.
- **REQ-CSD-021** (Unwanted) — 동사는 `enabled` 키가 없는 엔트리를 쓰지 않아야 한다(shall not). 하나라도 쓰이면 그 사용자의 codex 전체가 시작 실패한다.
- **REQ-CSD-022** (Ubiquitous) — 비활성화 발행의 `enabled` 값은 `false` 여야 한다.

### C.4 병합 — 멱등하고 비파괴적

- **REQ-CSD-030** (Event-driven) — **When** 대상 경로에 해당하는 엔트리가 이미 존재하면, 병합기는 그 엔트리의 `enabled` 값만 갱신하고 새 엔트리를 덧붙이지 않아야 한다.
- **REQ-CSD-031** (Event-driven) — **When** 이미 `enabled = false` 로 같은 경로가 등록돼 있으면, 결과 바이트는 입력과 **동일**해야 하며 쓰기는 일어나지 않아야 한다.
- **REQ-CSD-032** (Ubiquitous) — 병합기는 대상 엔트리 이외의 기존 엔트리·주석·공백·줄끝 형식을 보존해야 한다.
- **REQ-CSD-033** (State-driven) — **While** 대상 파일의 어떤 엔트리 구간에 파서가 인식하지 못한 줄이 있는 동안, 병합기는 그 엔트리를 건드리지 않아야 한다 — 쌍둥이 prune이 `FirstUnrecognizedLine` 으로 세운 것과 같은 보수 규율이다.

### C.5 실패 시 무쓰기 · 백업 · dry-run

- **REQ-CSD-040** (Ubiquitous) — 기본 실행은 **dry-run** 이어야 한다. 실제 쓰기는 명시적 `--force` 에서만 일어나야 한다.
- **REQ-CSD-041** (Event-driven) — **When** 쓰기가 실행되면, 쓰기 **이전에** 원본 사본이 `<cfg>.bak-<UTC RFC3339-compact>` (mode 0600)으로 남아야 하고, 원본의 sha256이 출력되어야 한다.
- **REQ-CSD-042** (Event-driven) — **When** 해석·병합·백업 중 어느 단계라도 실패하면, 대상 파일은 **바이트 불변**이어야 한다.
- **REQ-CSD-043** (Event-driven) — **When** codex home이 해석되지 않거나 config가 없거나 읽히지 않으면, 동사는 그 사실을 말하고 fail-open으로 종료해야 한다(에러가 아니다).
- **REQ-CSD-044** (Ubiquitous) — 동사는 대상이 되지 않은 이유·건너뛴 이유를 항목별로 보고해야 한다. 침묵은 사용자에게 아무것도 가르치지 않는다.

### C.6 경계

- **REQ-CSD-050** (Ubiquitous) — `internal/codexwiring/skills.go` 는 읽기 전용으로 유지되어야 한다. 발행기는 새 파일에 놓이고, 기존 파서의 `ParseSkillEntries` / `SplitConfigLines` / `JoinConfigLines` 를 **소비만** 해야 한다.
- **REQ-CSD-051** (Unwanted) — 이 SPEC의 구현은 `internal/cli/codex_skills_prune.go` 의 동작을 바꾸지 않아야 한다(shall not).
- **REQ-CSD-052** (Ubiquitous) — 순수 함수(내용 in → 내용 out + 항목별 판정)와 I/O 러너는 분리되어야 한다.

## §D 범위 밖

### Out of Scope — 노출 생성 / 프로비저닝

- 이 키로 스킬을 **등록**하거나 노출을 만드는 일. t504 V5/V7이 무력을 실측했다. 노출 통로는 `$CODEX_HOME/skills/` 디렉터 컨벤션, `.agents/skills/` 미러, 플러그인 루트다.
- 미러 생성·수정(`internal/template/skill_mirror.go`)에 대한 어떤 변경도 하지 않는다.

### Out of Scope — 기존 유령 엔트리 청소

- 이 머신의 죽은 경로 엔트리 49건 정리는 카드 t506(`SPEC-CODEX-GHOST-SKILLS-PRUNE-001`) 소관이다. 인접하되 공유하지 않는다.
- `moai clean --codex-skills` 의 판정·출력·플래그를 바꾸지 않는다.

### Out of Scope — 재활성화 / 일괄 조작

- `enabled = true` 로 되돌리는 동사, 여러 스킬 일괄 비활성화, 프로필/프리셋은 이 카드 범위 밖이다. 후속 카드에서 다룬다.

### Out of Scope — Claude Code 쪽 비활성화

- `.claude/settings.json` 이나 Claude Code의 스킬 노출은 건드리지 않는다. 이 기능은 **codex 계층 한정**이다.

### Out of Scope — 데스크톱 앱 의미론

- Codex 데스크톱 앱이 이 키를 어떻게 쓰는지는 t504에서 미측정(Gaps)이다. 이 SPEC은 CLI 관측만을 근거로 하며 앱 동작을 주장하지 않는다.

## §E 근거 문서

- `.moai/reports/t504/skills-config-path-shape.md` — 채택된 셀 행렬(IV/N/V4-V7/D1f·D2f·D1d·D2d), F1(`enabled` 필수), 노터치 해시 쌍
- `internal/cli/codex_skills_prune.go` — 쌍둥이 쓰기 경로의 모양(dry-run 기본, `--force`, 선백업, 항목별 판정 보고)
- `internal/codexwiring/skills.go:9` — 읽기 전용 선언(docstring)
