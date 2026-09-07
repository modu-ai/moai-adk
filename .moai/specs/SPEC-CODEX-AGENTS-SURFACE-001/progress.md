# progress.md — SPEC-CODEX-AGENTS-SURFACE-001

Card: t505 (factory) · worktree `.claude/worktrees/t505` · branch `WT-codex-agents-table` · base `0b1e27877` (origin/develop)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-07
artifacts: spec.md + plan.md (Tier S, 2 artifacts) + spec-compact.md (auto-generated compact view)
authoring note: Phase 8 produced an evidence-driven judgment proposal (three dispositions: model omit retained / skills.config dropped / [agents] not wired with per-key type map + explicit t494 A1 overturn); the operator approved it at the DP1 gate ("진행 — SPEC 생성", Tier S confirmed) and manager-spec formalized it into REQ-CAS-001..005 / AC-CAS-001..006 with two-cell RED-now greps measured on tree 0b1e27877. DP1 correction applied: AC-CAS-002 uses the corrected `grep -cE` alternation form (unescaped pipes).
plan-audit note: plan-auditor runs at the lead's task #11 before run-phase entry (Phase 1 Plan Audit Gate; Tier S PASS threshold 0.75). This file's audit-ready signal records artifact completeness, not an audit verdict.

## §E.2 Run-phase Evidence

Milestones M1-M4 complete. Runs in worktree `.claude/worktrees/t505`, branch `WT-codex-agents-table`, base `0b1e27877`, run commit `d3d529631` (`docs(t505): codify 0.153.4 agents-surface judgments in agents-codex.yaml`; 1 file changed, 94 insertions(+), 7 deletions(-)).

- M1: `internal/template/agentemit/agents-codex.yaml` 만 수정 — ① `fields.model` / class `model` / `model-pin-manager-git` 근거에 0.153.4 증거 세트(에이전트 파일 model이 부모·[agents] 기본값 양쪽에 선행하는 precedence, id churn #42874, open-bug #32587, flip 불활성 근거 writer.go:87-155·manifest.go:126-137), ② skill-loader 근거에 0.153.4 `SkillConfig = {path?, name?, enabled}` override-not-grant + `enabled` 필수(t504) + P4-P7 무관측 실측, 기존 0.152.1 소견과 재프로브 조항 보존, ③ 신규 `documented_drops` 엔트리 `codex-global-agents-table` + 인접 주석 블록에 0.153.4 측정 키별 타입 지도와 t494 A1 명시적 번복 기록. `codex_measured_version: "0.147.0"` 불변(AC-CSL-009).
- M2: `make agents-emit` rc=0 → `git diff --stat -- internal/template/templates/.codex` 빈 출력(rc 0) + 대조군(매니페스트 자체 diff) 비어있지 않음(+94/−7) + `make agents-emit-check` rc=0. 커밋된 11개 TOML 바이트 동일 — 직접 편집 0.
- M3: `go test ./internal/template/agentemit/... -count=1` ok/rc 0(매니페스트 파스 AC-013, 골든 AC-009, `TestEmitAllOmitsModel` 포함), `make agents-emit-check` rc 0. `go test ./internal/template/... -count=1`의 유일 실패 `TestManifestHashFormat`은 origin/develop 상속 결함으로 귀속 입증(catalog.yaml·.md·테스트 코드는 base..HEAD 델타 0; catalog 재생성 b4b34805d 이후 본문 개정 321111fe5가 해시를 낡게 함). 상세: `.moai/reports/t505/verdict.md`.
- 변이체 A-D(커밋 `d3d529631` 기준 복원): A 심도 가드 작동(001 green 유지 + 002 0/rc1 red) / B 복합 삭제 双 red / C TOML 손편집 → templates diff 비어있지 않음 + 골든 가드 red(sha256 mismatch) + 배포 픽스처는 구조적으로 무red(배출-대-커밋소스 등가 검사라서; SPEC 가드 지명 정정 후보) / D 매니페스트 값 오염 → 파서 fail-closed("outside the measured value set"). 전부 단일 파일 `git restore` 복원 후 green 재확인.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-09-07
run_commit_sha: d3d529631 (M1+M2; M4 evidence commit follows)
run_status: complete
ac_pass_count: 5 (AC-CAS-001..005) + AC-CAS-006 PASS-WITH-DEBT (상속 red 1건 귀속 입증 — verdict.md 발견 3)
ac_fail_count: 0
preserve_list_post_run_count: 0 (`templates/.codex/**`·전체 `*.go`·codexwiring 무변경; `git status --short` tracked 변경은 SPEC 아티팩트 2개뿐)
l44_pre_commit_fetch: n/a (레인은 push하지 않음 — develop push는 리드 일괄)
l44_post_push_fetch: n/a (동일)
new_warnings_or_lints_introduced: 0 (Go 코드 0변경 — lint/vet 대상 없음, plan.md M3 범위 명시)
cross_platform_build: n/a (Go 코드 0변경 — 크로스빌드 재측정 대상 없음)
total_run_phase_files: 1 (agents-codex.yaml) + M4 기록(verdict.md 신규, progress.md 본 절, spec.md frontmatter 전이)
m1_to_mn_commit_strategy: 단일 run 커밋 d3d529631 (M1+M2) + M4 evidence 커밋(본 커밋)

## §F Phase 4 Mode Selection

Input parameters: tier S; scope = 1 source file (manifest YAML comments/rationales, ~40 LOC); domains = 1 (agentemit manifest); language mix = YAML + markdown; concurrency benefit = LOW (serial milestones — M2 depends on M1, M3 on M2); agent teams prereqs = n/a.

Mode evaluation (pre-assessment only — the Decision line is recorded by the orchestrator before the first run-phase spawn per orchestration-mode-selection.md §D):
- direct — candidate: the work is one-file YAML comment editing + regeneration + inherited-test verification; verification-shaped.
- serial — candidate: canonical owner for run-phase implementation (manager-develop) on a real source file.
- fanout — not indicated: single domain, strict milestone dependencies.
- sweep — not indicated: not a mechanical bulk transform.

Decision: serial (manager-develop `dev-t505`, spawned opus/medium per profile)

Recording-latency note: this Decision line was recorded by the orchestrator (lane-10) at run completion rather than strictly before the first run-phase spawn — the run executor was spawned directly after the Implementation Kickoff Approval gate (operator-selected autonomous progression), and the pre-assessment block above was in place before that spawn. The chosen mode matches the spawn that executed: one sequential manager-develop over M1→M4, no concurrent spawns.

Justification note: measurement/judgment cards move the risk from code correctness to discipline (byte-identity proof, REQ-CSL-008 regeneration obligation, no-stamp-raise); t504 precedent logged `direct` for a zero-source-file measurement, while this SPEC edits one real source file, which weighs toward the canonical manager-develop owner. The orchestrator owns the call.

## §E.4 Sync-phase Audit-Ready Signal

sync_status: complete
sync_commit_sha: f89b922e0
status_transition: in-progress → completed (frontmatter, on the sync commit)
changelog_decision: NO user-facing CHANGELOG entry (manager-docs B12 assessment). Reasoning: the run commit's change surface is `internal/template/agentemit/agents-codex.yaml` alone, whose own header (lines 4-5) states "Build input only: this file lives in the emitter package, NOT under templates/, and is never distributed to user projects" — it is not a template output. The 11 committed TOMLs are byte-identical (run-phase M2 proof: `git diff --stat -- internal/template/templates/.codex` empty + `make agents-emit-check` rc=0), and the Go diff since base is empty (verified this session: `git diff --stat 0b1e27877..HEAD -- '*.go'` → empty). No user-visible behavior, CLI surface, or distributed artifact changed, so no CHANGELOG entry is warranted. README / docs-site: same basis — no user-facing surface changed; expected no-op, not re-measured.
mx_tag_validation: 0 added / 0 removed / 0 updated — validated against plan.md §D's zero-tag record. Evidence (this session, tree @ 84fa5fde1): `git diff --stat 0b1e27877..HEAD -- '*.go'` → empty output (rc 0); the full branch diffstat is 4 SPEC artifacts + the manifest YAML only. No Go file touched → no @MX-tagged surface exists to annotate, matching plan.md §D's rationale (edit surface is a YAML build input; the deliverable IS rationale comments).
documentation_sync: spec.md frontmatter status + progress.md §E.4 are the entire sync-phase write surface. No body-content problems found (no blocker).
