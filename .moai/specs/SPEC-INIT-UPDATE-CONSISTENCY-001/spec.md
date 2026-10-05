---
id: SPEC-INIT-UPDATE-CONSISTENCY-001
title: "init/update 정합성·문서 정리 (F8·F9·F12-F15·F17, F16 소멸 판정 포함)"
version: "0.1.0"
status: completed
created: 2026-09-13
updated: 2026-09-14
author: manager-spec
priority: P2
phase: "v3.2.0"
module: "internal/cli, internal/config, internal/core/project, internal/template/templates, internal/web"
lifecycle: spec-anchored
tags: "init, update, consistency, config-defaults, web-console, update-display"
tier: M
---

# SPEC-INIT-UPDATE-CONSISTENCY-001 — init/update 정합성·문서 정리

## HISTORY

| 버전 | 날짜 | 변경 |
|------|------|------|
| 0.1.0 | 2026-09-13 | 초판 — 카드 t588 (init/update 전수 조사 2026-09-09 감사 보고서 C6 축). F16 소멸 판정 + F15·F17 수용 판정 반영 |

## §1 개요

init/update 전수 조사(2026-09-09, `.moai/reports/init-tui-audit-20260909.md`)에서 보고된 정합성·문서 결함군 중 카드 t588 소관 축을 정리한다. 본 SPEC은 **워크트리 `.claude/worktrees/t588` (HEAD fac132d38, develop 동기)에서 재측정한 결과**를 기준으로 한다 — 감사 보고서의 file:line 앵커는 t583(질문 16→4 정온화), t586(위저드 재구조화), t694(update TUI 표시) 이후 이 트리에서 이동했거나 소멸한 것이 있으며, 본 문서의 모든 근거는 이 트리 실측값이다.

### 축별 처분 요약 (관측 → 처분 → 검증)

| 축 | 관측된 결함 (이 트리 기준) | 처분 | 검증 형태 |
|----|---------------------------|------|-----------|
| F8 | `project.mode` 유령 설정 — 템플릿 `project.yaml.tmpl:15` 가 `mode: personal` 내장, `writeProjectModeYAML` (`internal/core/project/initializer_expansion.go:52-70`) 이 패치, Go 리더 0건 (`grep` 전수 — `internal/config`, `internal/web` 어디에도 project.yaml `mode` 키 리더 없음) | **fix** — 쓰기 경로+플래그+템플릿 키 제거 | 리더 0건 grep 재실행 + init 테스트 갱신 |
| F9 | `workflow.execution_mode` 기본값 갈라짐 — 컴파일 기본 `internal/config/defaults.go:882` `"team"` vs 템플릿 `workflow.yaml:20` `auto` | **fix** — 컴파일 기본을 `auto`로 정렬 | default↔template parity 테스트 |
| F12 | "Updated N" 과대 — `.sh`/`.sh.tmpl` 4페어(hook 래퍼)가 `ListTemplates` (`internal/template/deployer.go:314`) 에서 stripped-target 기준 2회 계수 → `managedRedeployed` 4 과대, `update_tux.go:165` `fileCount+detail.ManagedRedeployed` 로 합산 | **fix** — stripped-target 기준 dedupe | 페어 포함 목록 카운트 단위 테스트 |
| F13 | 백업 뿌리 3곳 분산 + 결과 요약이 config 백업 1곳만 인쇄 (`update_tux.go:185-186`) — 단, 3-뿌리 분산 자체는 `update_namespace_protect.go` 패키지 문서가 "No consolidation" 으로 기록한 의사결정 | **분산=record-only / 요약 누락=fix** | 렌더 테스트 (존재하는 뿌리 전부 표기) |
| F14 | (a) 탭 수 기록 14 vs 실측 12 드리프트 → **이 트리에서 소멸**: `consoleTabs()` (`internal/web/schemaform.go:34-86`) 실측 14탭, `tab_layout_test.go:11` · `primary_surface_test.go:67` 가 14탭 계약 고정. (b) model_policy 숨은 키 → **소멸**: `handlers.go:380-385` G3-5 주석이 의도적 carry-forward 로 문서화. (c) FieldDefs 파싱-렌더 갭 잔존 가능 — `parseSchemaForm` (`schemaform.go:324`) 이 `AllFields()` 전체를 파싱하는데 `SectionHarness` 필드(`internal/settings/schema_sections.go:310, 433-443` — 주의: `internal/web/schema_sections_test.go` 는 동명의 테스트 파일이며, 렌더 측은 `internal/web/schemaform.go:213`)의 렌더 패널이 `schemaSectionMetas()` 에 없음 | **fix** — parity 가드 테스트로 잔존 간극 고정 | AllFields-editable ⊆ 렌더됨 ∪ 면서(exempt) 테스트 |
| F15 | update 사전 청소(`.moai/config` 통째 삭제) 후 복원이 `sections/*.yaml` 한정 (`internal/cli/update/backup/restore.go:92-100` — sections 디렉터리 walk + 비-YAML skip). `evaluator-profiles/*.md`, `astgrep-rules/**` 사용자 수정은 병합 복원 대상 아님 | **accept (수동 복구) + 안내 문구** | 요약 안내 문구 존재 grep/렌더 테스트 |
| F16 | "update -c 가 감사 선택 재묻기 없음" | **소멸 (dissolved)** — 아래 §3 실측 근거 | 소멸 근거 = 아래 4건 코드 판독 |
| F17 | `ConfigManager.Save` 6섹션 한정(user/language/quality/git-convention/git-strategy/llm, `internal/config/manager.go:186-262`) + init/web 쓰기가 yamlpatch·전용 writer 로 우회 — 구조적 사실이며 회귀 방지 설계(REQ-GSI-001/002 byte 보존)와 일치 | **record-only** — @MX:DEBT 어노테이션으로 기록 내구화 | `@MX:DEBT` 존재 grep |
| — | (F13 분산 본체) | **record-only** — 기존 패키지 문서 기록 존중, 신규 산출물 없음 | 기록 위치 = `update_namespace_protect.go` 패키지 문서 (기존) |

## §2 요구사항 (GEARS)

### REQ-ICU-001 — project.mode 유령 설정 제거 (F8)

**Where** 프로젝트가 초기화되는 상황에서, the init pipeline shall not persist `project.mode` configuration key that has no Go consumer. The init pipeline shall not register a `--project-mode` flag whose value no component reads.

- 대상: `internal/core/project/initializer_expansion.go` `writeProjectModeYAML` (+ `WritePhase1Configs` 내 호출), `internal/core/project/initializer.go:50` `ProjectMode` 필드, `internal/cli/init.go:91,369-375,581` 플래그 등록·검증·할당, `internal/template/templates/.moai/config/sections/project.yaml.tmpl` `mode:` 키와 주석.
- 기존 프로젝트의 잔존 키는 다음 update 사이클의 config 재배포로 자연 소멸한다 — 별도 마이그레이션 없음.

### REQ-ICU-002 — execution_mode 기본값 정렬 (F9)

The compiled default for `workflow.execution_mode` shall equal the value shipped in the template workflow.yaml. **When** the loader seeds a workflow.yaml whose `execution_mode` key is absent, the effective value shall not invert the template-declared meaning (`auto` — 하네스 자동 선택 위임, `internal/config/closed_sets.go:20-22`).

- 방향 판정: `auto` 가 정답이다 — (1) 템플릿이 `auto` 를 배포하고, (2) `closed_sets.go` 가 `ExecutionModeAuto` 를 "defers the choice to harness auto-selection" 으로 정의하며, (3) `execution_modes_test.go:55` 가 `auto` 를 "the defer-to-harness default" 로 부른다. `"team"` 은 `auto` 값 도입 이전의 잔재다.
- 현재 Go 리더 0건(`closed_sets.go:35` 주석 "Nothing in the Go tree reads ExecutionMode" — 이 트리에서 grep으로 재확인)이지만, 로더 부분-재정의 계약(키 없으면 컴파일 기본 시딩)상 파일 키 삭제가 의미 반전을 일으키므로 정렬한다.

### REQ-ICU-003 — "Updated N" 페어 이중 계수 제거 (F12)

**When** the update outcome summary renders the updated-file count, the count shall not charge both members of a `.sh`/`.sh.tmpl` deployment pair as two files — each rendered deployment target (stripped of `.tmpl`) counts once.

- 대상: `internal/cli/update_template_sync.go` 의 `managedRedeployed` 계수 루프(t40 주석 블록, stripped-target 적재 뒤 `IsMoaiManaged` 카운트) — `restoredSet` 이 이미 stripped-path 집합이므로 기집재 여부로 skip 하는 dedupe 가 최소 수정이다.
- 소비자: `internal/cli/update_tux.go:165` `renderUpdateOutcome(report.OutcomeUpdatedFiles, fileCount+detail.ManagedRedeployed, ...)` — t694 이후 표시 경로. 표시 측 수정 불요, 계수 원천만 수리.

### REQ-ICU-004 — 백업 뿌리 요약 완전 표기 (F13)

**When** the update outcome summary renders after a run that created more than one backup root, the summary shall name every backup root it created (config backup, user-owned namespace backup, archive-drift) with its recoverable path. The three-root structure itself is a recorded deliberate decision and shall not be consolidated.

- 분산 본체 처분: record-only — `update_namespace_protect.go` 패키지 문서의 "No consolidation; the three concerns remain separately tracked" 기록이 SSOT.
- 요약 누락 처분: fix — `update_tux.go:164-186` `renderUpdateOutcome` 이 config 백업 `backupPath` 만 받는 시그니처를 확장해 존재하는 뿌리를 전부 표기.

### REQ-ICU-005 — 스키마 필드 파싱-렌더 parity 가드 (F14 잔존분)

The web console shall not parse a schema field that no panel renders, except fields explicitly declared read-only or render-exempt. **Where** the schema-driven form parser consumes `AllFields()`, a parity guard test shall pin that every editable FieldDef has a render surface or an explicit exemption entry.

- 감사가 지목한 탭 수 드리프트(14 기록 vs 12 실측)와 model_policy 숨은 키는 이 트리에서 소멸했다(§3) — 본 REQ는 그 잔존 간극(현 시점 관측: `SectionHarness` FieldDefs 의 렌더 홈 부재 정황)만 다룬다. 면서(exempt) 목록 확정과 harness 필드의 렌더 홈 부여/면서 선언은 run-phase census 의 산출이다.

### REQ-ICU-006 — 비-섹션 config 사용자 수정 수동 복구 안내 (F15, 수용 판정)

**When** the update pre-clean removes `.moai/config` and the restore step re-lays only `sections/*.yaml`, the outcome summary shall state that customizations outside `sections/` (evaluator-profiles/, astgrep-rules/) are not merge-restored and shall name the backup path for manual recovery.

- 수용 판정 근거: 두 디렉터리의 병합 복원은 과잉 설계다 — `evaluator-profiles/*.md` 는 배포 기본값 문서 4종, `astgrep-rules/` 는 실험적 룰셋(dogfood, CLAUDE.local.md §2.3 이 local-only 이동 완료)이며, 백업 자체는 존재하므로 수동 복구 경로가 유효하다. 병합 확장 대신 안내로 닫는다.

### REQ-ICU-007 — ConfigManager.Save 범위 기록 내구화 (F17, 기록 판정)

The `ConfigManager.Save` godoc shall declare its six-section persistence scope, and the function shall carry an `@MX:DEBT` annotation naming the ceiling (나머지 섹션은 yamlpatch/typed 경로로 영속됨) and the upgrade trigger (신규 섹션이 SetSection 흐름에 합류할 때).

- 기록 판정 근거: Save 의 6섹션 한정은 버그가 아니라 구조다 — git-convention/git-strategy 는 dirty-gate 로 byte 보존하며(SPEC-GITSTRATEGY-SAVE-ISOLATION-001), 나머지 26개 섹션의 대화형 쓰기는 yamlpatch seam 이 담당한다(주석 보존 계약). Save 를 32섹션으로 확장하면 hand-edit 파괴 회귀가 재발한다. 기록만으로 충분하다.

## §3 소멸 판정 기록 (dissolved-with-evidence)

### F16 — "update -c 재묻기 없음" — 소멸

**판정: 소멸.** 측정 대상 트리: `.claude/worktrees/t588` HEAD `fac132d38`. 근거 4건(전부 본 트리 코드 판독):

1. **감사 선택 질문 자체가 어느 표면에도 존재하지 않는다** — 감사 모델·게이트 3종·codex 훅 질문(구 #10-14)은 t583 질문 축소로 위저드에서 완전 제거됐다. `internal/cli/wizard/questions.go:293` `initSharedQuestionIDs = ["conversation_language", "user_name"]` + `Page3Questions`(agent_wiring, autonomy_tier) = init 4질문. 제거된 설정은 배포 기본값으로 해소되고 웹 콘솔에서 변경한다(`questions.go:30-32` 계약 주석).
2. **재구성 경로가 의도적으로 같은 집합을 쓴다** — `ReconfigureQuestions` (`questions.go:269-288`) = `DefaultQuestions`(5) + `GitQuestions`(7)이며, page-3 배제는 `AC-WIZ-012a` 로 고정돼 있다(`questions.go:301-302`).
3. **init 과 update -c 사이 불일치가 없다** — F16 의 결함 형태는 "init 은 묻는데 update -c 는 안 묻는다"였다. 이제 어느 쪽도 묻지 않으므로 불일치가 정의상 성립하지 않는다. 배포 기본값 + 웹 콘솔이라는 동일한 해소 경로를 양쪽이 공유한다.
4. **재구성 후속 단계도 감사 설정을 만지지 않는다** — `runWorkflowConfigStep` (`internal/cli/init_workflow_flags.go:66-101`) 이 토글하는 것은 branch_guard + worktree auto_* 4개뿐이다.

카드 전제("t583 질문 축소로 소멸") — **확인**. 웹 안내 문구 손보는 대안도 불요: 재구성 위저드가 감사 선택을 약속하는 문구를 이 트리에서 발견하지 못했다(§Gaps 참조).

### F14 (a)(b) — 탭 수 드리프트·model_policy 숨은 키 — 소멸

- 탭 수: `consoleTabs()` 실측 14개(identity, language, launch, llm, workflow, git-worktree, audit, codex, agentfm, report, mcp, crosssession, feedback, gate) = 테스트 계약 14. 드리프트 없음.
- model_policy: `handlers.go:380-385` G3-5 — "no longer a UI field… forged model_policy form value is therefore ignored, not persisted" 로 의도적 carry-forward 가 코드에서 문서화됨. 결함 아님.

## §4 제약

- Template-First: `internal/template/templates/**` 수정 시 `make build` 필수, `agents-emit`/`commands-emit` 영향 검토(project.yaml.tmpl은 에미터 대상 아님 — 확인 후 진행).
- 훅 래퍼 `.sh`/`.sh.tmpl` 쌍은 함께 수정(CLAUDE.local.md §2.3) — 본 SPEC은 쌍을 삭제하지 않는다(계수만 수리).
- F12 수리는 표시 경로(update_tux.go)와 계수 원천(update_template_sync.go) 중 **계수 원천**만 고친다 — t694가 재작업한 표시 렌더를 무접촉한다.
- question 재확장 금지 — t583 4질문 baseline 유지(카드 의존 조건).

## §5 수용 기준 요약

세부 Given-When-Then 은 `acceptance.md` §D AC 행렬 참조. 전체 축의 검증은 단위 테스트 + grep 계열로 기계 판정 가능하며, 실행 재현(F12 실측 카운트)은 run-phase 첫 산출물로 한다.

### Out of Scope — 질문 집합 변경

- init/reconfigure 질문의 재확장·재편성 — t583 이 확정한 4질문 baseline 과 `AC-WIZ-012a` 배제 계약은 본 SPEC 이 건드리지 않는다.

### Out of Scope — 백업 뿌리 통합

- 3개 백업 뿌리(`.moai-backups/`, `.moai/archive/skills/…`, `.moai/backups/update-<ISO>/`)의 단일 뿌리 통합 — SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 이 기록한 "No consolidation" 의사결정을 존중한다.

### Out of Scope — 타 카드 소관 결함 축

- F1-F7, F10, F11(카드 t583-t587 소관), t589(update --add-codex·하네스 3-way 배포), F2/F3 등 update 수리군 — 본 SPEC 은 F8·F9·F12-F15·F17 만 담는다.

## §Gaps (plan-phase 미관측)

- F16 웹 안내 문구: 재구성 위저드가 감사 선택을 재묻는다고 약속하는 웹 콘솔 문구의 존재 여부를 본 트리 전수에서 확인하지 않았다 — 발견되면 run-phase 에서 문구 1건 수정으로 흡수한다(REQ-ICU-006 과 동일한 규모).
- F14 census: `AllFields()` editable 전체의 렌더 표면 대응표는 run-phase 산출물이다 — 현 시점 관측은 `SectionHarness` 의 패널 부재 정황까지만이다.
