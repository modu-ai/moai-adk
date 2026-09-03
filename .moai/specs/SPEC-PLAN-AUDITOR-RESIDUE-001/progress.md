# SPEC-PLAN-AUDITOR-RESIDUE-001 — 진행 기록

카드: t450 · 브랜치: WT-plan-auditor-residue · Tier: M

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec_id: SPEC-PLAN-AUDITOR-RESIDUE-001
phase: plan
status: draft
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
author: manager-spec
tree_sha_at_authoring: "7835148d3"
precondition_verified:
  t367_ancestor_of_develop: true   # merge-base 측정 — 18fc2c9ef ∈ origin/develop 7835148d3
needs_clarification_markers: 0      # t443 우회는 t367 선례(2549f775f)로 확정 — 질의 불필요
hazard_recorded:
  t443_agentemit_drift: sync-auditor.toml sha256 mismatch — go test ./internal/template/agentemit/... FAIL 관측
                                    # t443 소관 — 본 SPEC은 record-and-not-repair (REQ-008)
evidence_path_planned: .moai/reports/t450/verdict.md   # lane이 작성 — plan-phase에서 작성 안 함
```

plan-phase 측정 근거 (모두 이번 실행, 트리 7835148d3):
- 지연 조항 출처: f47d7f5a9 본문 + `.moai/reports/t387/verdict.md` Gaps / 4244c4a06 본문
- 금지 경로 명령: plan-auditor.md:395 (gitignored — `.gitignore:230`, check-ignore exit=0)
- 쌍둥이 기존 드리프트 2 hunk: D7-1 예시 식별자 + Tier-resolved ceiling 문단 (수리 안 함, 기록만)
- agentemit 골든 FAIL: sync-auditor.toml sha256 mismatch (t443 소관)

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
