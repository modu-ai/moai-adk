# SPEC-V3R2-RT-003 Progress Tracker

> Live progress and session-handoff state for **Sandbox Execution Layer (Bubblewrap / Seatbelt / Docker)**.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `plan.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.

## HISTORY

| Version | Date       | Author                            | Description                                                            |
|---------|------------|-----------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)      | Initial progress tracker — plan documents written; ready for plan-auditor |

---

## Current Status

| Field | Value |
|-------|-------|
| Phase | `plan` |
| Status | `plan-complete-pending-audit` |
| Branch | `plan/SPEC-V3R2-RT-003` |
| Worktree | `(none — plan executed in main checkout per doctrine #822 plan-in-main)` |
| Base | `origin/main` (`c810b11b7`) |
| Plan-auditor | iteration 1 pending (target: PASS 0.85+) |
| Run-phase entry | pending plan-auditor PASS + RT-002 merge for full wiring (mock-based M4 unit test 가능 prior to RT-002) |

---

## Plan Phase Deliverables (this session)

- [x] `spec.md` (24KB; pre-existing v0.1.0 — body §1-§10 unchanged from 2026-04-23 author GOOS draft; 33 EARS REQs + 16 ACs)
- [x] `plan.md` v0.1.0 (~26KB; 6 milestones M1-M6, traceability matrix 33 REQ → 20 AC → 52 tasks, 10 risks PR-01..PR-10, 10 MX tags, plan-audit-ready checklist 15/15 PASS, AUTO migration BC-V3R2-003 contract documented)
- [x] `research.md` v0.1.0 (~30KB; 14 sections, OS sandbox tooling landscape, OWASP Agentic Top 10 2025 mandate, 2026 incidents — Cline npm-token + Claude Code rm -rf, network allowlist 8-host evidence, env scrub 6-pattern OWASP/SLSA/SOC2 cover, 16-language neutrality verification, 0 new external dependency)
- [x] `acceptance.md` v0.1.0 (~22KB; 20 ACs in G/W/T format with happy-path + edge cases + test mapping; per-OS gating policy; AC-21 LSP carve-out baseline + alpha.2 deferred runtime validation per master §12 Q7)
- [x] `tasks.md` v0.1.0 (~21KB; 52 tasks T-RT003-01..52 across M1-M6; greenfield package; per-OS test gating; LOC est ~3240; performance budget tasks T-43/44; doc-only AskUser stub T-50 for REQ-050)
- [x] `progress.md` v0.1.0 (this file)
- [x] `issue-body.md` (GitHub PR body for tracking)

---

## Run Phase Plan (next session, after plan PR merged)

Per `plan.md` §8 Implementation Order Summary:

1. **M1 (P0, RED)**: ~32 신규 test functions (T-RT003-01..14). 모두 RED. baseline 100% GREEN. Per-OS skip 정상.
2. **M2 (P0, GREEN type-level)**: enum + interface + sentinel errors (T-RT003-15..17). AC-13 type 부분 + sentinel matching GREEN.
3. **M3 (P0, GREEN per-OS)**: bubblewrap + seatbelt + profile generator + LSP carve-out clause (T-RT003-18..22). AC-01/02/09/12/14/21 GREEN.
4. **M4 (P0, GREEN CI)**: docker + dispatcher + permission-divergence handling (T-RT003-23..26). AC-11 + AC-16 GREEN.
5. **M5 (P0, GREEN wiring)**: env scrubbing + workflow.yaml + security.yaml + frontmatter parser + agent_lint rule (T-RT003-27..39). AC-03/06/07/08/10/15 GREEN.
6. **M6 (P1, REFACTOR + final)**: dispatcher 추출, profile strategy pattern, env scrub regex cache, benchmark, `moai doctor sandbox`, MX tags, CHANGELOG, quality gate (T-RT003-40..52). AC-04/17/18/19 GREEN.

Total: **52 tasks** across 6 milestones. Estimated LOC: ~1300 source + ~1190 test + ~80 bench + ~95 config + ~140 CLI + ~30 lint + ~195 refactor delta + ~210 doc = **~3240 LOC** new (≤ 500 KiB binary delta budget per spec §7).

> **NOTE**: This SPEC is **greenfield** — `internal/sandbox/` 는 현재 0 file. M2 부터 모든 source 가 신규. RT-005 (`Source` enum / multi-tier merge) 는 이미 머지됨 (PR #826 2026-05-10) → config 통합 가능. RT-002 (permission resolver) 미머지 → M4 의 permission-divergence handling 은 mock-based 우선 (T-RT003-25), real wiring 은 RT-002 머지 후 follow-up. ORC-004 (worktree MUST) 미머지 → `WritableScope` 는 caller 가 채우는 contract 그대로 stub 가능.

---

## Out-of-scope notes (orchestrator-side responsibility)

다음은 본 SPEC plan/tasks 범위 밖 — orchestrator (`/moai run` / `/moai project` skill) 또는 별도 SPEC owner:

- **REQ-V3R2-RT-003-050 (AskUser flow when backend missing)** — task T-RT003-50 doc-only stub. 실제 AskUserQuestion 호출은 orchestrator 가 owner (subagent 는 user 와 직접 대화 불가 per `agent-common-protocol.md` § User Interaction Boundary). RT-003 launcher 는 `*SandboxBackendUnavailable` error + 정확한 missing binary name 만 제공; orchestrator 가 이를 받아 AskUser round 실행.
- **frontmatter 자동 채움 (BC-V3R2-003 AUTO migration)** — SPEC-V3R2-MIG-001 owner. 본 SPEC plan §3 (AUTO Migration Path) 가 contract만 정의. MIG-001 plan 작성 시 본 SPEC 을 dependency 로 reference.
- **Worktree 자동 생성** — SPEC-V3R2-ORC-004 owner. `WritableScope` 는 caller 가 자동 worktree path 를 주입; RT-003 은 자체적으로 worktree create 하지 않음.
- **Hook JSON for sandbox SystemMessage** — SPEC-V3R2-RT-001 owner. RT-003 은 `SystemMessage` 텍스트만 emit; 실제 JSON 포맷 + Claude Code stream wiring 은 RT-001.
- **Container image 빌드 (`moai/sandbox:latest`)** — SPEC-V3R2-EXT-004 (migration framework) deferred. 본 SPEC 은 default `alpine:latest` fallback (M3에서 결정).
- **Windows native sandbox** — v3.1+ deferred. master §10 R3.
- **LSP runtime 호환성 alpha.2 validation** — master §12 Q7. 본 SPEC AC-21 은 baseline clause-only.

---

## Downstream Consumer Pipeline (post-merge)

본 SPEC merge 후 unblock 되는 downstream:

| SPEC | Consumes | Impact |
|------|----------|--------|
| SPEC-V3R2-MIG-001 | frontmatter contract (plan §3.1) + decision tree (§3.2) | v2 → v3 마이그레이션 시 agent frontmatter `sandbox:` 필드 자동 채움 |
| SPEC-V3R2-HRN-003 | `Launcher.Exec` API + sandbox 결정성 | harness thorough mode 가 sandboxed agent 호출 신뢰 가능 |
| SPEC-V3R2-WF-003 | Sandbox layer | `--mode loop` Ralph fresh-context iteration 시 sandboxed implementer |
| SPEC-V3R2-ORC-001 | per-role default sandbox | agent roster reduction 시 manager-cycle/builder-platform 의 sandbox default 자동 적용 |

---

## Dependency check (pre-merge sanity)

| Dependency SPEC | Merge state (2026-05-10) | RT-003 plan impact |
|-----------------|---------------------------|---------------------|
| SPEC-V3R2-CON-001 (P7 Frozen zone) | ✅ MERGED — Constitution 에서 FROZEN | RT-003 plan 의 Section 3.1 Brand Context 와 §3.2 Design Brief contract 와 무관 (CON-001 는 design subsystem 이며 RT-003 는 runtime — 별 레이어). spec §10 Traceability 의 P7 reference 만 inherit. |
| SPEC-V3R2-RT-002 (Permission resolver) | NOT MERGED | M4 의 T-RT003-25 (permission-divergence) 는 mock-based 로 unit test GREEN; real wiring 은 RT-002 merge 후 follow-up commit. plan §1.5 + §5 PR-09 명시. |
| SPEC-V3R2-RT-005 (Settings resolver) | ✅ MERGED (PR #826, 2026-05-10) | M5 의 `SecurityConfig.Sandbox` substruct 가 RT-005 `Source` enum + provenance 활용 가능. config integration smooth. |
| SPEC-V3R2-ORC-004 (Worktree MUST for implementer) | NOT MERGED | `WritableScope` 는 caller 가 채우는 contract 이므로 RT-003 자체 unit test 는 `t.TempDir()` 사용 가능; ORC-004 머지 후 e2e test 가능. |

→ run-phase 진입 권장 시점: **plan PR merged + RT-002 merged**. 그 이전 시점에는 M1-M3 완료 후 M4 mock-based GREEN 까지만 진행 가능 (M5 full wiring 은 RT-002 의존 X — config + frontmatter 단독으로 wiring 가능, RT-002 는 launcher.go 안 단일 함수만 영향).

---

## Plan-auditor entry checklist

`.claude/agents/moai/plan-auditor.md` v1 PASS criteria 자체 점검 (plan.md §7 reproduce):

- [x] Criterion 1 (spec ↔ plan goal restatement)
- [x] Criterion 2 (REQ → AC → Task 100% coverage; 33 × 20 × 52 매트릭스)
- [x] Criterion 3 (AC G/W/T testable + evidence)
- [x] Criterion 4 (TDD methodology aligned with quality.yaml)
- [x] Criterion 5 (file-touch list aligned with spec affected modules)
- [x] Criterion 6 (validation gates per milestone — M1-M6)
- [x] Criterion 7 (risk register w/ mitigations — 10 entries PR-01..PR-10)
- [x] Criterion 8 (perf budget benchmarks — AC-17/18/19 + tasks T-43/44)
- [x] Criterion 9 (mx_plan with file:line + reason — 10 tags)
- [x] Criterion 10 (out-of-scope clear vs spec §2)
- [x] Criterion 11 (dependency SPEC contract reference — RT-002/RT-005/ORC-004/MIG-001)
- [x] Criterion 12 (greenfield vs brownfield 명확 — greenfield package)
- [x] Criterion 13 (embedded-template parity via `make build` — M5 gate)
- [x] Criterion 14 (breaking: true SPEC AUTO migration path 상세 — plan §3)
- [x] Criterion 15 (16-language neutrality 보존 — research §10)

자체 score: 15/15 expected PASS at first audit pass (target 0.85+).

---

## M1-M6 milestone progress (pending run phase)

| Milestone | Tasks | Status | Gate |
|-----------|-------|--------|------|
| M1 (RED scaffolding) | 14 (T-01..14) | ⏳ PENDING (run-phase entry) | go test → all new RED, baseline GREEN |
| M2 (Type-level GREEN) | 3 (T-15..17) | ⏳ PENDING | enum + sentinel matching tests GREEN |
| M3 (Per-OS GREEN) | 5 (T-18..22) | ⏳ PENDING | per-OS smoke test GREEN; AC-01/02/09/12/14/21 |
| M4 (Docker + dispatcher) | 4 (T-23..26) | ⏳ PENDING (mock-based; full RT-002 merge optional) | TestDocker GREEN (CI), dispatcher 4-scenario GREEN |
| M5 (Config + lint wiring) | 13 (T-27..39) | ⏳ PENDING | embedded.go regenerated; AC-03/06/07/08/10/15 |
| M6 (REFACTOR + final) | 13 (T-40..52) | ⏳ PENDING | bench p99 ≤ budget; quality gate PASS; AC-04/17/18/19 |

---

## 다음 세션 시작점 (paste-ready resume message)

> Per `.claude/rules/moai/workflow/session-handoff.md` canonical 6-block format. 본 SPEC plan PR merged 후 다음 세션 (run-phase 진입) 시 사용.

```text
ultrathink. SPEC-V3R2-RT-003 run 진입.
applied lessons: project_v3_master_plan_post_v214 (RT-003 plan PR open / merged), lessons #11 retired-agent stub chain, lessons #14 worktree paste-ready Block 0.

전제 검증:
0) git rev-parse --show-toplevel → /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-003 (★ critical, after `moai worktree new SPEC-V3R2-RT-003 --base origin/main`)
1) git branch --show-current → feat/SPEC-V3R2-RT-003
2) gh pr view <RT-003-plan-PR> → MERGED
3) ls .moai/specs/SPEC-V3R2-RT-003/ → 6 files (spec/plan/research/acceptance/tasks/progress)

실행: /moai run SPEC-V3R2-RT-003

머지 후: SPEC-V3R2-RT-002 (permission resolver) → SPEC-V3R2-MIG-001 (migrator wires sandbox: <backend>) → master §12 Q7 alpha.2 LSP carve-out validation
```

> Note: `--worktree` paste-ready Block 0 강제 (lessons #14). RT-003 run 은 SPEC worktree 안에서 실행 — main checkout 에서 `/moai run` 시 `Agent(isolation: "worktree")` base mismatch 회피.

---

End of progress.md v0.1.0.
