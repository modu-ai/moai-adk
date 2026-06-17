# progress.md — SPEC-V3R6-DOCS-V3-README-001

> Run-phase / sync-phase / Mx-phase evidence skeleton. Plan-phase에서는 §E.1만 채우고 §E.2-§E.5는 placeholder heading만 둔다.

---

## §A. SPEC Status

- **ID**: SPEC-V3R6-DOCS-V3-README-001
- **Tier**: M (standard)
- **Status**: draft (plan-phase)
- **Created**: 2026-06-17
- **Plan-phase commit**: (pending run-phase)

---

## §B. Milestone Tracker

| Milestone | Scope | Status | Commit |
|-----------|-------|--------|--------|
| M1 | Agent catalog rewrite (en) | pending | — |
| M2 | GLM tier-model 정정 (en) | pending | — |
| M3 | `/moai` 17-command + "47 Skills" 헤더 제거 (en) | pending | — |
| M4 | README.ko.md 동기화 (ko) | pending | — |
| M5 | statusline 보존 + scope boundary 확인 | pending | — |
| M6 | en/ko cross-check + 최종 AC 검증 | pending | — |

---

## §C. AC Tracker

| AC | Severity | Status | Evidence |
|----|----------|--------|----------|
| AC-1 (agent catalog en) | MUST | pending | — |
| AC-2 (agent catalog ko) | MUST | pending | — |
| AC-3 (command set en) | MUST | pending | — |
| AC-4 (GLM tier en) | MUST | pending | — |
| AC-5 (GLM tier ko) | MUST | pending | — |
| AC-6 (en/ko sync) | MUST | pending | — |
| AC-7 (statusline 보존) | MUST | pending | — |
| AC-8 (scope boundary) | MUST | pending | — |

---

## §D. Pre-flight Verification (plan-phase)

- [x] docs-truth.md 존재 (122L, commit 4a6f4b4d3)
- [x] §1 1차 소스 재검증: `ls .claude/agents/moai/*.md` → 7 files ✓
- [x] §2 1차 소스 재검증: `internal/spec/status.go` ValidStatuses → 8 values ✓
- [x] §3 1차 소스 재검증: `internal/spec/lint.go` required slice → 12 entries ✓
- [x] §4.2 1차 소스 재검증: `ls .claude/commands/moai/*.md` → 17 files ✓
- [x] §5 1차 소스 재검증: `internal/config/defaults.go` DefaultGLMHigh → `glm-5.2[1m]` ✓
- [x] README.md (1370L) + README.ko.md (1418L) 존재 ✓
- [x] drift inventory 17항 완료 (rewrite 13 + info-only 4축)

---

## §E.1 Plan-phase Audit-Ready Signal

- **Plan-phase artifacts**: 4 files authored (spec.md + plan.md + acceptance.md + progress.md)
- **Tier**: M (standard)
- **Drift inventory**: 17 items across 7 axes (§1 agent catalog, §4.2 command set, §5 GLM tier = rewrite; §2/§3/§6/§7 = info-only/preservation)
- **1차 소스 재검증**: 전수 PASS at commit 4a6f4b4d3 — blocker 없음
- **Milestones**: M1..M6 (6 milestones, 파일-단위)
- **AC**: 8 AC (AC-1..AC-8), 전수 grep/diff 기반 기계적 검증 가능
- **Scope boundary**: README.md + README.ko.md 2개 파일만; docs-site / Go / CLAUDE.md / template EXCLUDE
- **Anti-overengineering**: 사실 정정(reconciliation)만; 추상화/설정 시스템/미래 확장 hook 금지
- **Plan-phase readiness**: audit-ready (plan-auditor verdict 대기)

---

## §E.2 Run-phase Evidence

_<pending run-phase>_

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

---

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase>_

---

## HISTORY

- 2026-06-17: plan-phase artifacts authored (4 files). §E.1 채움, §E.2-§E.5 placeholder heading만. drift inventory 17항, 1차 소스 전수 재검증 PASS.
