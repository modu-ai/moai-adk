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
open_decisions:
  - "축 (a) vs (b) 최종 확정 — Implementation Kickoff Approval에서 운영자 판정 (plan.md §C, 권고: (b))"
  - "축 (b) 확정 시 opt-in 메커니즘 1~3 중 선택 (plan.md §C 축 (b))"
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
