# spec-compact — SPEC-CODEX-AGENTS-SURFACE-001

Auto-generated compact view. Canonical sources: spec.md (REQ/AC bodies), plan.md (milestones/risks).

## REQ (5)

- **REQ-CAS-001** (Ubiquitous) — retain `fields.model.emit: false` + class-`model` `omit` + `model-pin-manager-git` drop; rationales carry the 0.153.4 evidence set (precedence-winning model field, #32587, id churn, P-03 accepted-but-wrong).
- **REQ-CAS-002** (Ubiquitous) — retain the skill-loader documented-drop; rationale carries 0.153.4 `SkillConfig` override-not-grant + P4–P7 measured silence + retained 0.152.1 non-read + the re-probe clause.
- **REQ-CAS-003** (Ubiquitous) — add the `[agents]` judgment record: `documented_drops` entry `codex-global-agents-table` + adjacent type-map comment block (0.153.4), stating no-wire, the dual-source hazard, the rc=1 strict-parse hazard, and the explicit t494 A1 overturn. No Go change.
- **REQ-CAS-004** (Event-driven) — when manifest rationales change, regeneration proves zero emission delta: `make agents-emit` → byte-identical TOMLs → `make agents-emit-check` rc 0 (REQ-CSL-008; `TestCodexAgentsDeployFixture`).
- **REQ-CAS-005** (State-driven) — while `codex_measured_version` stays `"0.147.0"`, record axis-wise stamps in rationales and never raise the top-level stamp (AC-CSL-009 decision).

## AC (6, inline in spec.md §D)

- **AC-CAS-001** — judgment-record id grep: RED 0/rc1 @ 0b1e27877 → ≥1 after M1.
- **AC-CAS-002** — type-map key grep (`-cE`, 3 keys): RED 0/rc1 @ 0b1e27877 → ≥3 distinct lines after M1 (mutant depth guard).
- **AC-CAS-003** — `0\.153\.4` cite grep: RED 0/rc1 @ 0b1e27877 → ≥3 (three required sites) after M1.
- **AC-CAS-004** — version-stamp preservation: 1 before, 1 after; non-vacuous co-observed with AC-CAS-001/003.
- **AC-CAS-005** — zero emission delta (M2): regen → `git diff --stat templates/.codex` empty + control non-empty; mutants C (TOML touch) / D (parsed-value edit) define the red direction.
- **AC-CAS-006** — `go test ./internal/template/... -count=1` green + `make agents-emit-check` rc 0 (M3); inherited guards keep their meaning.

## Files to modify

- `[MODIFY] internal/template/agentemit/agents-codex.yaml` — the ONLY source edit (rationale/comment refresh + one judgment record).
- `[NEW] .moai/reports/t505/verdict.md` — evidence record (M4).
- Untouched by design: `internal/template/templates/.codex/**`, all `*.go`, `internal/codexwiring/**`.

## Exclusions (Out of Scope)

- No emitter field additions (no model flip, no skills.config emission, no `[agents]` emission; zero Go changes).
- No codexwiring changes (no new `.codex/config.toml` merge surface).
- No full-manifest 0.153.4 re-measurement of other axes (sandbox_mode / mcp_servers / layout / effort map — separate candidate card).
- No doctor enhancements (t504 lineage, separate).
- No skills provisioning on `skills.config` (t504 directive stands).
- No direct TOML edits (regenerate-not-edit; byte-guarded).
