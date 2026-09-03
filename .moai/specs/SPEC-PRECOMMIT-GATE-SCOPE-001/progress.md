# progress.md — SPEC-PRECOMMIT-GATE-SCOPE-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-PRECOMMIT-GATE-SCOPE-001
card: t461
branch: WT-precommit-gate-scope
base: a239cf050
status: draft
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
verified_facts:
  install_call_site: "installPreCommitHookOptional(...) — internal/cli/update_template_sync.go (심볼 인용; 카드의 :574는 실측 :575로 drift)"
  twin_pair: "preCommitHookContent 상수 (internal/cli/hook_install_precommit.go) + internal/template/templates/.git_hooks/pre-commit (3,245 bytes)"
  twin_guard: "TestPreCommitTemplateMatchesConstant (internal/cli/hook_install_precommit_test.go) — AC-PGM-010 교차참조 (internal/cli/precommit_relocation_test.go)"
  defect_entry: "883d53852 (2026-07-28, SPEC-PRETOOL-GATE-MOVE-001, #1189) — 사전 커밋 52b5e4bf5 (2026-07-05, SPEC-PRECOMMIT-001)는 무해 fast subset"
  gate_defaults: "NewDefaultGateConfig(): Enabled=true, SkipTests=false (internal/config/defaults.go)"
  gate_keys: "GateConfig.enabled / skip_tests / disabled_steps (internal/config/types.go) — 카드 지목 3키 모두 존재 확인"
  gate_yaml_path: ".moai/config/sections/gate.yaml (loadGateSection, internal/config/loader_gate.go)"
  config_wipe: "CleanMoaiManagedPaths가 .moai/config/ 통째 삭제 후 템플릿 재배포 (internal/cli/update/deploy/deploy.go) — REQ-006의 근거"
  enabled_shared_switch: "QualityGate.Run이 Enabled=false에서 즉시 통과 (internal/hook/quality/gate.go) — 단독 moai gate와 공유 스위치"
t237_collision: "t237 / issue #1641 — 동일 twin 파일 편집 카드. 본 SPEC 소관 아님(REQ-008/AC-005). 병합 순서 메모는 run-phase 시 본 파일에 기록"
decisions_recorded:
  - "D1 해소 — 축 (b) 확정: pre-commit 맥락 heavy gate 기본 OFF, gate.yaml opt-in (Implementation Kickoff Approval, 운영자 결정 2026-09-03)"
  - "D1 해소 — 메커니즘 1 확정: 훅이 MOAI_PRECOMMIT=1 마커와 함께 moai gate 호출, 러너는 마커 하에서만 gate.pre_commit.enabled 존중. 단독 moai gate 불변, 새 서브커맨드 없음, 러너 분기점 1개"
  - "D1 해소 — 신설 REQ-009: gate.pre_commit.enabled를 moai web 설정 화면에서 편집 가능 (운영자 지시). 저장은 정확히 .moai/config/sections/gate.yaml"
web_axis_anchors:
  schema_gap: "settings 스키마에 gate 섹션 없음 (internal/settings/schema.go SectionID 목록) — SectionGate 신설 + FieldDef(PersistSeam, Section: gate) 등록 필요"
  render_wiring: "internal/web/schemaform.go schemaSectionMetas() 패널 배선 + fieldsets.templ fieldsetSchemaSection 스키마 주도 렌더"
  naming_convention: "폼 컨트롤 name=\"gate.pre_commit.enabled\" + bool hidden companion __present (parseSchemaForm EC-1)"
  hide_precedent: "workflow_agents 은닉은 레지스트리 수준 부재 (agent_settings_test.go:97 폼 컨트롤 미렌더, :230 TestWorkflowAgentsWebSubmissionIgnored) — 노출은 레지스트리 등록으로 충분, 별도 플래그 불필요"
  save_path: "ApplySchemaEdits → WriteSectionViaSeam (internal/settings/sectionwrite.go) — .moai/config/sections/<section>.yaml yamlpatch 기록, 주석 보존. gate.yaml 정확히 기록 확인"
  i18n: "internal/web/assets/i18n.js sec.*.title 4-locale 패턴 (en/ko 확인, ja/zh는 run-phase 확인)"
  naming_overlap: "git_strategy.<mode>.hooks.pre_commit (validation.go checkStringField 3곳)은 다른 subtree — plan.md 제약 8에 기록"
d5_adjudication: "기각(변경 없음) — Event-detected는 GEARS 정캐논: .claude/skills/moai-workflow-spec/SKILL.md:59 (GEARS Five Patterns 표 행 'Event-detected (replaces IF/THEN)'), .claude/skills/moai-foundation-core/SKILL.md:113 (5패턴 서술에 'Event-detected (replaces the deprecated conditional modality)' 명시)"
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
