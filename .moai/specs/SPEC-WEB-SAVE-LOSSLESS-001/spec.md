---
id: SPEC-WEB-SAVE-LOSSLESS-001
title: "moai web 설정 Save 무손실 — 편집 없으면 무기록, 편집은 해당 필드만, 미모델링 키·주석 영생존"
version: "0.1.2"
status: in-progress
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/settings, internal/config, internal/profile, internal/web"
lifecycle: spec-anchored
tags: "web, config, lossless, yaml-patch, dirty-gate, save, defect, issue-1731"
tier: M
issue_number: 1731
related_specs: [SPEC-WEB-WRITE-SAFETY-001, SPEC-GITSTRATEGY-SAVE-ISOLATION-001, SPEC-SEAM-GREENFIELD-001, SPEC-SEAM-GREENFIELD-002, SPEC-WEB-CONSOLE-011]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-29 | manager-spec | 최초 draft. 카드 t1314 (GitHub issue #1731 — 외부 사용자 michaelleone 재현 보고). 워크트리 `WT-web-save-lossless` (develop `2b1233b13` 기점)에서 plan-phase 산출 작성. 코드 근거 12곳 plan-phase 직접 확인. |
| 0.1.1 | 2026-09-29 | manager-spec | plan-audit iter-1 (FAIL 0.85) 정정. **F1(차단)** — D2 수리 재설계: UserConfig가 `name`만 모델링(`pkg/models/config.go:32-37`)하므로 구조체 복사 수리는 공허 — user.yaml `name:` 행의 seam 라인-스플라이스로 전환(REQ-WSL-005/AC-WSL-005 개정). **F2(차단)** — quality_extras save-time 강제 폐기 결정 확정(plan §A.4, `schema_sections_test.go:285-290` M4 전환 목록 추가). **F3(차단)** — AC-WSL-002 술어를 line-splice/upsert-폴백 변이로 분리(C3 정합). **F4** — M1에 seam 허용 확장(`sectionwrite.go:56-64`) 명시. Q1-Q3 감사 판정 기록(plan §A.4). |
| 0.1.2 | 2026-09-29 | manager-spec | plan-audit iter-2 (CONCERNS 0.90 — PASS-with-debt, F1-F4 해소 확인) 처분 반영. **F5(major, 처분 (a))** — 잔여 재마샬 3경로(devMode/convention 스칼라 쓰기 seam, 중첩 쓰기 seam)를 M1 공동 범위로 편입: REQ-WSL-002/003 범위 문구에 명시적 포괄, 신규 AC-WSL-009로 이슈 지목 키(`constitution.session_effort_default`) 생존 고정. **F6(minor)** — AC-WSL-005 name-부재 최초 설정 변이에 C3 허용 1행 추가. |

---

## §1 Context & Motivation

### §1.1 외부 재현 보고 (GitHub #1731 — ground truth로 취급하되 본 SPEC이 정리한 현황과 대조)

`moai web` 설정 페이지에서 **아무것도 편집하지 않고 Save를 누르면** 사용자가 손으로 추가한 키와 주석이 설정 파일에서 사라진다는 보고(GitHub issue #1731, 2026-09). 보고자는 typed 섹션이 `ConfigManager.LoadRaw` → `Save`로 재직렬화되는 경로(`internal/settings/sectionapply.go:122-199`, `internal/config/manager.go:199`)를 결함 기제로 지목했다. 본 카드는 이 보고를 현재 트리(develop `2b1233b13`)에서 재검증해, **이미 착지된 수리**와 **본 SPEC이 수리할 잔여 결함**을 분리했다.

### §1.2 이미 착지된 수리 (본 SPEC의 재수리 대상 아님 — 착지 맥락으로 인용)

| 보고된 결함 | 현황 | 근거 |
|---|---|---|
| v3.1.2에서 `git-strategy.yaml` 무조건 재기록 | 착지 | `ConfigManager.Save`의 git-strategy/git-convention dirty-or-absent 게이트 (`internal/config/manager.go:236-260`, SPEC-GITSTRATEGY-SAVE-ISOLATION-001) |
| `mcp.yaml` 부재 시 HTTP 500 + 부분 저장 | 착지 (`e365c2d30`, SPEC-SEAM-GREENFIELD-001/002) | `PatchFile`의 부재 파일 greenfield 시딩 (`internal/settings/yamlpatch/yamlpatch.go:75-95`) |
| 무편집 Save의 typed 섹션 재기록(보고 당시 `c000a1fcb`) | 대부분 착지 (SPEC-WEB-WRITE-SAFETY-001, REQ-WWS-003) | typed `reflect.DeepEqual` 게이트 (`internal/settings/sectionapply.go:186`), `changed==0`이면 `Save()` 스킵 (`sectionapply.go:194-198`), nested 동일 게이트 (`internal/settings/nested.go:127,150`) |

### §1.3 잔여 결함 (본 SPEC의 수리 대상 — 현재 트리 기준 코드 근거)

**D1 — 한 필드라도 실제로 바뀌면 Save()가 편집하지 않은 섹션 4개를 재기록한다.**
`ConfigManager.Save`는 user/language/quality/llm 4개 파일을 dirty 게이트 없이 무조건 재마샬한다 (`internal/config/manager.go:215-225,263`). 예: `llm.glm.models.high` 하나만 바꿔도 user.yaml·language.yaml·quality.yaml가 함께 재직렬화되어 (a) 구조체가 모델링하지 않는 키(user `github_username` — `pkg/models/config.go:32`의 UserConfig는 `name` 하나뿐, quality `constitution.session_effort_default`)가 소실되고, (b) 주석이 제거되고, (c) 파일에 없던 기본값 키가 써진다. 보고서의 `user.yaml`/`quality.yaml`/`llm.yaml` 손실이 이 기제다.

**D2 — 사용자 이름 편집 시 user 섹션이 전체-교체된다.**
`SyncToProjectConfig`는 이름이 다르면 로드된 구조체를 통째로 `models.UserConfig{Name: ...}`로 교체한다 (`internal/profile/sync.go:28`) — 모델링 여부와 무관하게 로드된 모든 비(非)Name 필드가 폐기된 뒤 `Save()`(`sync.go:65`)가 재마샬한다.

**D3 — 부재 키에 대한 빈 제출(`EmptySubmits`)이 빈 키를 새로 기록한다.**
workflow.yaml 감사 pin 필드(`workflow.audit.{claude,codex,glm}.model/effort`, `internal/settings/schema_sections.go:385-394`)는 `EmptySubmits` 옵트인 — 폼이 미설정 상태를 `""`로 제출하고 `parseSchemaForm`은 이를 실제 제출값으로 통과시킨다(`internal/web/schemaform.go:388-391`). seam no-op 게이트는 키가 **존재**할 때만 동치를 검사하고(`internal/settings/sectionapply.go:68-70`), 부재 키 + `""` 제출은 편집으로 남아 upsert되어 `model: ""` 빈 키가 파일에 추가된다. 의미상 이미 "unset"인 상태가 바이트 수준에서 변형되는 것으로, 보고서가 관측한 workflow.yaml 빈 키 추가다.

**D4 — 부분 저장에 대한 명시적 보증 부재.**
`handleSave`는 7개 writer를 순차 호출한다(`internal/web/handlers.go:455-541`). 다중 파일 트랜잭션은 아무 곳에도 보증돼 있지 않고, 뒤 단계 실패 시 앞서 기록된 파일이 남는다. 성취 가능한 보증(검증-선결·per-file 원자성·사전-플라이트)을 계약으로 명문화할 필요가 있다.

### §1.4 설계 결정 (plan.md §A에 1급 항목으로 상세)

채택: **A+B 결합** — (A) typed 섹션 스키마 편집을 yamlpatch seam 라인-스플라이스로 라우팅(구조 보존), (B) 차이-게이트 확장(무차이 무기록, 부재-키 빈 제출 no-op), 그리고 잔여 `Save()` 호출 경로를 막는 **Save 더티-게이트 백스톱**(git-strategy 선례의 4섹션 확장) + **user.yaml `name:` 행의 seam 라인-스플라이스**(D2 — plan-audit F1 재설계). 기각: 순수 B(편집된 섹션 자체의 손실이 남음), 순수 A(동치 재기록·user/language 미커버), 구조체-복사 기반 user 수리(UserConfig가 `name`만 모델링해 재마샬 시점에 손실 — plan.md §A.1). 근거와 기각 사유는 plan.md §A.1, 결정·판정 기록은 plan.md §A.4.

## §2 Requirements (GEARS)

### §2.1 무손실 계약 (lossless contract)

- REQ-WSL-001 (Ubiquitous): 제출된 값이 영속된 값과 하나도 다르지 않으면, The save flow shall 어떤 설정 파일도 기록하지 않는다 (mtime 변경 포함).
- REQ-WSL-002 (Event): 제출이 정확히 한 필드의 값을 변경할 때, The save flow shall 그 필드가 영속되는 행(또는 노드)만 디스크에서 변경한다 — 동일 파일의 다른 행과 다른 파일의 내용은 불변이어야 한다. 이 계약의 대상에는 development_mode·git_convention 스칼라 쓰기 seam과 중첩(nested) 프로젝트-설정 쓰기 seam의 잔여 전체-재마샬 경로가 명시적으로 포함된다 — 해당 경로의 편집도 본 요구의 무손실 술어에서 예외가 아니다 (plan-audit iter-2 F5).
- REQ-WSL-003 (Ubiquitous): The save flow shall 구조체가 모델링하지 않는 키와 주석을, 편집 대상 파일을 포함해 항상 보존한다. 이 보존 의무는 schema-field 경로뿐 아니라 development_mode·git_convention 스칼라 쓰기 seam과 중첩(nested) 쓰기 seam을 포함한 Save 흐름의 모든 writer에 적용된다 (plan-audit iter-2 F5).
- REQ-WSL-004 (Event-detected): 편집 대상 키가 파일에 부재하고 제출값이 빈 문자열일 때(빈 제출이 허용되는 필드 포함), The save flow shall 키를 새로 기록하지 않는다 — 부재는 이미 해당 필드의 미설정 상태로 해석된다.
- REQ-WSL-005 (Event): 사용자 이름 편집이 user 섹션을 변경할 때, The save flow shall user.yaml의 `name:` 행만 변경하고 나머지 모든 바이트(미모델링 키·주석·빈 줄 포함)를 원문으로 보존한다 — 구조체 재마샬이 아닌 행 단위 치환으로.

### §2.2 부분 저장 보증

- REQ-WSL-006 (Event-detected): 어떤 writer가 파일 기록에 실패로 끝나면, The save flow shall 이미 성공한 기록을 되돌리려 시도하지 않고 — 대신 — 실패 이전에 검증 가능한 단계(폼 검증, 대상 파일 판독 가능성)를 모두 통과한 경우에만 기록을 시작한다. 즉, 사전-플라이트에서 검증 가능한 실패는 어떤 파일도 기록하지 않은 상태로 실패해야 한다.
- REQ-WSL-007 (Ubiquitous): The save flow shall 모든 파일 기록을 per-file 원자 쓰기(temp + rename)로 수행한다.
- REQ-WSL-008 (Event-detected): 다중 파일 트랜잭션 원자성(전체 성공 또는 전체 무효)은 보증 대상이 아니다 — 부분 저장이 발생하면 The response shall 어느 파일이 기록됐고 어느 파일이 남았는지 사용자가 판독 가능한 오류 메시지를 낸다 (현행 `renderErrorPage` 관행 계승).

### §2.3 게이트 위계 (기존 수리의 회귀 금지)

- REQ-WSL-009 (Ubiquitous): The regression suite shall SPEC-WEB-WRITE-SAFETY-001이 정립한 값-불변 게이트들(typed/nested `DeepEqual`, seam 스칼라 동치, 부재 bool 유효 기본값)을 본 SPEC의 변경 후에도 동일하게 유지한다.
- REQ-WSL-010 (Ubiquitous): The regression suite shall 손실 행위(무편집 재기록, 미모델링 키 소실, 주석 제거, 빈 키 추가)를 단언하는 테스트를 포함하지 않는다 — 그런 단언은 RED여야 하며, 해당 행위의 부정(무손실)을 단언하는 테스트로 교체된다. 본 SPEC이 전환하는 기존 동작의 단언 테스트(save-time `quality_extras_enabled` 강제 포함 — plan.md §A.4)도 전환 목록에 명시적으로 포함한다.

## §3 Acceptance Criteria

Tier M — AC 열거와 Given-When-Then 시나리오는 `acceptance.md` §D AC 매트릭스에 위임한다 (AC-WSL-001..008).

## §4 Constraints

- C1: 영속화는 반드시 기존 규약 경로를 통해서만 수행한다 — seam은 `settings.WriteSectionViaSeam`/yamlpatch, 프로필은 `profile` 패키지, 나머지는 `config.NewConfigManager` API. 웹 레이어의 직접 `yaml.Marshal`/`os.WriteFile`은 금지 (REQ-WC3-008 계승).
- C2: `FieldDef` 스키마(`internal/settings/schema*.go`)가 편집 가능 표면의 SSOT다 — 스키마에 없는 키는 어떤 경로로도 기록되지 않는다 (REQ-WC11-013/018 계승).
- C3: yamlpatch의 문서화된 한계를 계약에 반영한다 — 기존 스칼라 교체는 라인-스플라이스(바이트 보존), upsert(부재 키 신설)는 재직렬화 폴백으로 빈 줄이 정규화될 수 있다 (`internal/settings/yamlpatch/yamlpatch.go:1-60` 패키지 헤더). REQ-WSL-002의 "해당 필드만"은 폴백 시 동일 파일 내 표현 정규화를 허용하되 데이터(키·값·주석) 손실은 허용하지 않는다.
- C4: 본 SPEC은 `moai web` Save 흐름에 적용된다. 동일 writer를 공유하는 TUI/CLI 경로(`settings.WriteProjectNestedConfig` 등)는 공유 seam 특성상 자동 수혜이며, 별도 회귀 검증만 요구한다.

## §5 Out of Scope

### Out of Scope — 이미 착지된 수리의 재구현

- git-strategy/git-convention dirty 게이트(SPEC-GITSTRATEGY-SAVE-ISOLATION-001), `mcp.yaml` 부재 500 수리(SPEC-SEAM-GREENFIELD-001/002, `e365c2d30`), 무편집 typed/nested no-op 게이트(SPEC-WEB-WRITE-SAFETY-001 REQ-WWS-003)의 재구현. 본 SPEC은 이들의 회귀 가드만 요구한다.

### Out of Scope — Save 흐름 밖의 쓰기 경로

- `moai init`/`moai update`의 템플릿 재배포 경로, 프로필 스토어(`preferences.yaml`) 자체의 포맷, `statusline.yaml`의 TUI/CLI 전용 쓰기 경로(`syncStatusline`은 웹 폼이 statusline 필드를 더 이상 제출하지 않아 `moai web` Save 흐름에서 도달 불가 — `internal/web/handlers.go:36-37,626-627`).
- 다중 파일 트랜잭션 원자성(staging 저널, 2-phase commit) — REQ-WSL-008이 명시적으로 비(非)목표로 선언했다.

### Out of Scope — 스키마 표면 확장

- 새 설정 키·섹션의 편집 가능화, read-only 해제(llm.mode/team_mode 등), `db`/`research` 등 폐선·제외군 섹션의 쓰기 허용 (REQ-WC11-018 유지).

## §6 Cross-References

- GitHub issue #1731 — 외부 재현 보고 원문.
- SPEC-WEB-WRITE-SAFETY-001 — 값-불변 게이트의 선행 SPEC (REQ-WWS-003/005/006).
- SPEC-GITSTRATEGY-SAVE-ISOLATION-001 — Save dirty 게이트의 선례 (백스톱 설계의 원형).
- SPEC-SEAM-GREENFIELD-001/002 — seam 부재-파일 greenfield 시딩.
- SPEC-WEB-CONSOLE-011 — 섹션 라우팅 SSOT(`FieldDef.Persist.Kind`)과 seam/typed 구분의 원류.
