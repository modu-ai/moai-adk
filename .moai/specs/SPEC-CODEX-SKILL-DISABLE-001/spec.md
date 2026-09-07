---
id: SPEC-CODEX-SKILL-DISABLE-001
title: "codex에서만 특정 스킬을 끄는 손 — 사용자가 이름으로 지목하고, moai가 파일 모양 게이트를 발행한다"
version: "0.2.0"
status: in-progress
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
- 2026-09-07 (plan-phase, v0.2.0) — 두 번째 선행 실측 `.moai/reports/t502/gate-path-shape.md`(19셀, 같은 격리 방식) 반영. 게이트가 **realpath 정규화 파일 비교**임이 확정돼 발행 표기가 하나로 고정됐고(§A), 「`MirrorModeSkipped` 는 해소 실패」라는 이전 판(v0.1.0)의 주장이 **반증돼 철회**됐다(§A 각주). 측정 게이트였던 조항은 「측정하라」에서 「측정된 표기를 발행하라」로 바뀌었다.

## §A 배경

Claude Code와 codex는 같은 스킬 집합을 본다. 그런데 **한쪽에서만 끄고 싶을 때**가 있다 — Claude Code에서는 계속 쓰되 codex 세션에서는 노출을 빼고 싶은 스킬. 지금 이 저장소에는 그 수단이 **없다**. 사용자는 `~/.codex/config.toml`을 손으로 편집해야 하고, 그 파일은 codex가 스스로 쓰는 파일이라 손 편집은 위험하다.

t504가 그 수단이 실재함을 실측했고, t502의 후속 측정이 **발행해야 할 경로 표기를 하나로 확정**했다.

| 실측 사실 | 셀 근거 | 설계에 미치는 구속 |
|---|---|---|
| `enabled` 는 파서의 **필수 필드**. 누락 엔트리 하나면 codex 시작이 rc=1로 죽는다 | t504 V1/V2/V3 (`missing field \`enabled\``) | 발행하는 모든 엔트리는 `enabled` 를 **명시**해야 한다. 위반 시 사용자 codex **전면 장애** |
| 게이트는 **문자열 비교가 아니라 realpath 정규화 파일 비교**다 | t502 E2 — 미러와 경로 성분을 하나도 공유하지 않는 별개 심링크로 쓴 엔트리가 실제로 스킬을 껐다 | 「리터럴이 통한다」는 문자열 일치 때문이 아니라 **같은 실파일로 해소되기 때문**이다. 표기 선택은 어느 실파일로 해소되느냐로 판단한다 |
| **두 미러 모양 모두에서 묶이는 표기는 하나뿐** — 절대 리터럴 미러 경로 | t502 Slit·Clit(차단) vs **Cres**(복사 미러 + `.claude/…` 표기 → 차단 안 됨) | 발행 표기는 `<projectRoot>/.agents/skills/<skill>/SKILL.md` 로 **고정**된다 |
| 금지 표기 3종은 **조용히** 무효다 | `.claude/…` 해소 경로(Cres), 디렉터 모양(E4·E5), 상대경로(E6) | 사용자는 껐다고 믿고 스킬은 계속 노출된다 — 신호가 전혀 없다 |
| 이 키는 노출을 **생성하지 못한다** | t504 V5·V7 마커 0 | 이 기능을 등록/프로비저닝 통로로 서술하거나 시험해서는 안 된다 |

`.claude/skills/…` 표기를 골랐다면 심링크 미러를 쓰는 저자의 기계에서는 동작하고 복사 폴백 사용자에게서는 조용히 죽었을 것이다 — t504가 막으려고 존재하는 실패 등급 그 자체다.

> **철회 (v0.1.0 → v0.2.0)**: 이전 판은 `MirrorModeSkipped` 가 이름 해소 실패로 이어져 거절 경로를 탄다고 적었다. **측정이 반증했다** — skipped 가 남기는 파일시스템 모양(실디렉터리 + 실 `SKILL.md`)은 복사 모양과 구별되지 않아 **정상 로드되고 정상 차단된다**(t502 C0/Clit). 해소는 성공한다. 다만 사용자 실엔트리가 **`SKILL.md` 를 품지 않은** 하위 사례(빈 디렉터리, 동명의 일반 파일)는 미측정이며, 그 경우에만 해소 실패가 성립할 수 있다 — Gaps 로 남긴다.

## §B 이 SPEC이 만드는 것

사용자가 **스킬 이름**으로 지목하면, moai가 그 이름을 절대 리터럴 미러 경로 `<projectRoot>/.agents/skills/<skill>/SKILL.md` 로 해석해 사용자 계층 `~/.codex/config.toml` 에 `enabled = false` 엔트리 하나를 발행한다. 발행은 **기본 비활성**이고, **명시 opt-in** 이며, **멱등**이고, **비파괴적**이며, 그 결과로 해당 스킬이 codex 노출에서 **실제로 사라져야** 한다.

## §C 요구사항 (GEARS)

> Tier M 예산: 요구사항 16개 이하. 아래는 정확히 16개다(`grep -o … | sort -u | wc -l` 로 실측). 조항을 넓혀서 수를 맞추지 않았다 — 한 조항이 두 성질을 지면 그 조항을 가르는 셀도 둘이어야 하고, 그러지 못하면 한 성질이 무방비가 된다.

### C.1 기본 비활성 — 요청하지 않은 쓰기는 없다

- **REQ-CSD-001** (Ubiquitous) — moai는 사용자가 비활성화 동사를 **직접 호출하고 쓰기 대상 계층을 호출문에서 명시한 경우가 아니면** `~/.codex/config.toml` 에 `[[skills.config]]` 를 쓰지 않아야 한다(shall not). 이 기능은 프로젝트 설정 키로 구동되지 않아야 한다(shall not) — 사용자 HOME에 대한 쓰기는 그 순간 사용자가 요청한 것이어야 하며, 프로젝트 설정 파일이 대신 요청할 수 없다.

### C.2 이름 해석과 발행 표기

- **REQ-CSD-010** (Event-driven) — **When** 사용자가 스킬 이름을 지목하면, 해석기는 그 이름을 실존하는 `SKILL.md` 절대 경로 **하나**로 해석해야 한다.
- **REQ-CSD-011** (Event-driven) — **When** 해석이 실패하거나(후보 루트 어디에도 없음, 또는 미러 자체가 없음) 둘 이상에서 해석되면(모호), 동사는 거절하고 **아무것도 쓰지 않아야** 하며, 세 경우를 서로 구별되는 사유로 보고해야 한다. **종료 코드는 사례마다 다르다**: 해석 불가와 모호는 **0이 아닌 값**으로 끝나야 하고(오타난 이름이 스크립트·CI에서 검출 가능해야 한다), 미러 부재만 **0** 으로 끝나야 한다(미러는 배포 실행이 만드는 산물이지 체크아웃이 만드는 것이 아니므로 부재는 평범한 상태다 — 이 저장소에서도 실측상 부재, `ls .agents/skills` exit 1).
- **REQ-CSD-012** (Ubiquitous) — 발행하는 `path` 값은 **프로젝트 루트 기준 절대 리터럴 미러 경로** `<projectRoot>/.agents/skills/<skill>/SKILL.md` 여야 한다. 심링크 해소 표기(`.claude/skills/…`), 디렉터 모양, 상대경로 표기를 발행해서는 안 된다(shall not) — 실측상 각각 복사 미러에서 무효(Cres) · 두 모양 모두에서 무효(E4·E5) · 무효(E6)이며, 셋 다 사용자에게 아무 신호 없이 조용히 무효다.
- **REQ-CSD-013** (Event-driven) — **When** 발행이 완료되면, **해석기가 고려하는 루트에서 비롯된** 그 스킬의 노출은 codex의 스킬 목록에서 **실제로 사라져야** 한다. 엔트리가 문법적으로 올바르게 쓰였다는 것만으로는 이 조항을 만족하지 않는다 — 게이트가 realpath 비교이므로, 올바른 모양의 엔트리도 다른 실파일을 가리키면 아무것도 하지 않는다.

  *범위 한정의 근거(요구사항이 아님):* 게이트는 **파일 단위**다. 같은 이름의 스킬이 해석기 범위 밖의 다른 루트(플러그인 루트 등)에서 **다른 실파일로** 동시에 올라오면, 미러 경로 엔트리 하나로는 그 노출이 사라지지 않는다 — 측정이 남긴 유보다(`gate-path-shape.md` Residual-risk). 조항을 "모든 노출이 사라진다"로 두면 도달 가능한 환경에서 **만족 불가능한 요구사항**이 되므로, 약화가 아니라 범위 한정으로 진술한다. 이 겹침은 REQ-CSD-011의 모호 거절로도 잡히지 않는다(그 판정은 해석기가 보는 루트만 본다). 사용자에게 보이는 대응은 M5 도움말이 진다.

### C.3 발행 내용 — 스키마 하드 에러를 만들지 않는다

- **REQ-CSD-020** (Ubiquitous) — 발행하는 모든 엔트리는 `path` 와 `enabled` 를 **둘 다** 가져야 하며, 비활성화 발행의 `enabled` 값은 `false` 여야 한다.
- **REQ-CSD-021** (Unwanted) — 동사는 `enabled` 키가 없는 엔트리를 쓰지 않아야 한다(shall not). 하나라도 쓰이면 그 사용자의 codex 전체가 시작 실패한다.

### C.4 병합 — 멱등하고 비파괴적

- **REQ-CSD-030** (Event-driven) — **When** 대상 경로 엔트리가 없으면 병합기는 그것을 **하나** 덧붙여야 하고, 이미 있으면 그 엔트리의 `enabled` 값만 갱신하고 새 엔트리를 덧붙이지 않아야 한다. 어느 경로로 끝나든 그 경로를 가진 엔트리 수는 정확히 1이어야 한다.
- **REQ-CSD-031** (Event-driven) — **When** 이미 `enabled = false` 로 같은 경로가 등록돼 있으면, 결과 바이트는 입력과 **동일**해야 하며 쓰기는 일어나지 않아야 한다.
- **REQ-CSD-032** (Ubiquitous) — 병합·쓰기는 대상 엔트리 이외의 기존 엔트리·주석·공백·줄끝 형식, 그리고 **원본 파일의 퍼미션 모드**를 보존해야 한다. 파일이 이미 존재하면 그 모드를 유지하고, 새로 만들 때만 `0600` 을 쓴다 — 표제 성질이 「비파괴」인 동작이 사용자의 `0644` 설정을 말없이 바꾸는 것은 그 성질과 어긋난다.
- **REQ-CSD-033** (State-driven) — **While** 대상 파일의 어떤 엔트리 구간에 파서가 인식하지 못한 줄이 있는 동안, 병합기는 그 엔트리를 건드리지 않아야 한다 — 쌍둥이 prune이 `FirstUnrecognizedLine` 으로 세운 것과 같은 보수 규율이다.

### C.5 실패 시 무쓰기 · 백업 · dry-run

- **REQ-CSD-040** (Ubiquitous) — 기본 실행은 **dry-run** 이어야 한다. 실제 쓰기는 명시적 `--force` 에서만, 그리고 쓰기 **이전에** 원본 사본이 `<cfg>.bak-<UTC RFC3339-compact>` (mode 0600)으로 남고 원본 sha256이 출력된 뒤에만 일어나야 한다.
- **REQ-CSD-041** (Event-driven) — **When** 어느 단계라도 실패하면 대상 파일은 **바이트 불변**이어야 한다. codex home 미해석·config 부재·읽기 실패는 에러가 아니라 사유를 말하고 fail-open으로 끝나야 한다.
- **REQ-CSD-042** (Ubiquitous) — 동사는 대상이 되지 않은 이유를 항목별로, **서로 구별되게** 보고해야 한다. 이 의무는 해석 실패 3종뿐 아니라 **병합 단계의 건너뛰기 사유**(미인식 줄, 이미 false 등)까지 포함한다. 침묵도, 한 문구로 뭉뚱그린 사유도 사용자에게 무엇을 고쳐야 하는지 가르치지 않는다.

### C.6 경계

- **REQ-CSD-050** (Ubiquitous) — `internal/codexwiring/skills.go` 는 읽기 전용으로 유지되어야 한다. 발행기는 새 파일에 놓여 기존 파서의 `ParseSkillEntries` / `SplitConfigLines` / `JoinConfigLines` 를 **소비만** 해야 하며, 순수 함수(내용 in → 내용 out + 판정)와 I/O 러너는 분리되어야 한다.
- **REQ-CSD-051** (Unwanted) — 이 SPEC의 구현은 `internal/cli/codex_skills_prune.go` 의 동작을 바꾸지 않아야 한다(shall not).

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
- `.moai/reports/t502/gate-path-shape.md` — **두 번째 선행 실측**(19셀). 게이트가 realpath 정규화 파일 비교임(E2), 두 미러 모양 공통으로 묶이는 표기가 리터럴 하나뿐임(Slit·Clit vs Cres), 디렉터·상대 표기 무효(E4·E5·E6), `MirrorModeSkipped` 해소 성공. REQ-CSD-012·REQ-CSD-013의 출처
- `.moai/reports/t502/probe.sh` — 파라미터화된 계측기(`--codex-home` / `--entry-path` / `--enabled` / `--skill` / `--expect`). AC-CSD-003의 E2E 기구이자 AC-CSD-050 재측정 기구
- `internal/template/skill_mirror.go:198-245` (`mirrorOneSkill`) — 미러 결과가 한 모양이 아니라는 근거(symlink / copy / skipped / failed). REQ-CSD-011의 출처
