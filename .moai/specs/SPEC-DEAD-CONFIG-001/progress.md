# Progress — SPEC-DEAD-CONFIG-001

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID pre-write self-check: `decomposition: SPEC ✓ | DEAD ✓ | CONFIG ✓ | 001 ✓ → PASS`
  (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`; digit-only end anchor; no alpha suffix).
- **Scope narrowed to `runtime.yaml` ONLY (v0.2.0)** after plan-auditor PASS-WITH-DEBT (0.76).
  github-actions.yaml removed from scope — it is the live config of `SPEC-CI-MULTI-LLM-001`
  (status: implemented), referenced by docs-site ×4 locales, and self-removes via `DeprecatedPaths`
  at v3.0.0. Its allowlist row is retained.
- **Self-corrected verification defect**: the v0.1.0 "grep found none — Verified" risk claim was false —
  the grep used `--include="*.go"` only and never searched the docs surface (verification-claim-integrity
  §1.1 violation). Claim removed in v0.2.0. Reversal facts independently re-verified: docs-site ×4
  (`guides/multi-llm-ci.md` en/ja/zh/ko, 1 each), `SPEC-CI-MULTI-LLM-001` frontmatter `status: implemented`,
  config-audit-2026-05-22.md §2.2 v2 correction, DeprecatedPaths dirs.go:245 Category B v3.0.0. NOT asserted:
  "settings-management.md ×2" (my grep found 0 matches — not restated on the coordinator's word).
- Tier: **S** (3 files: 2 deletions + 1 allowlist edit + `make build`; no production Go change).
- Plan-phase artifacts revised: spec.md, plan.md, acceptance.md; progress.md updated.
- Frontmatter: 12 canonical fields present + explicit `tier: S`; canonical field names
  (`created` / `updated` / `tags`) used.
- Headline finding: LoadRuntime removal is **SAFE-TO-REMOVE (no test fix)** — `budget_test.go`
  LoadRuntime tests use self-contained `t.TempDir()` fixtures; no production caller/importer of
  `internal/runtime`.
- Out-of-scope boundary recorded: `internal/defs` `DeprecatedPaths` migration manifest + `dirs_test.go`
  left intact (REQ-DC-007); live github-actions.yaml + its allowlist row retained; parked/governance
  YAMLs and MIG-003 dead loaders deferred.
- REQ↔AC↔milestone traceability (post-renumber): REQ-DC-001↔AC-DC-001↔M1; REQ-DC-002↔AC-DC-002/006/007;
  REQ-DC-003↔AC-DC-003↔M2; REQ-DC-004↔AC-DC-004↔M2; REQ-DC-005↔AC-DC-005/Scenario4↔M1→M2 order;
  REQ-DC-006↔AC-DC-008↔M1; REQ-DC-007↔AC-DC-009↔M3.
- Baseline (this tree, plan-phase read-only observation): `internal/config` audit test `ok`,
  `internal/runtime` LoadRuntime tests `PASS`, `internal/defs` DeprecatedPaths test `ok`.
- Status: draft — awaiting plan-audit gate + Implementation Kickoff Approval before run-phase.

## §E.2 Run-phase Evidence

- 구현 커밋: `d3337cc8e` (`chore(SPEC-DEAD-CONFIG-001): 미사용 runtime.yaml + dead allowlist row 제거`).
- 변경 파일 (git show d3337cc8e --stat 실측):
  - `.moai/config/sections/runtime.yaml` 삭제 (로컬, 27줄).
  - `internal/template/templates/.moai/config/sections/runtime.yaml` 삭제 (템플릿, 34줄).
  - `internal/config/audit_loader_completeness_test.go`: `acknowledgedDedicatedLoaders`에서 `runtime` 행 제거 (템플릿 runtime.yaml 삭제 후 dead).
  - `settings-management.md`: runtime.yaml 로더 행 제거 (양 트리 동일).
- `LoadRuntime` Go surface는 보존 (REQ-DC 범위: 온디스크 YAML만 제거, 프로덕션 코드 무변경).
- Out-of-scope 경계 준수: `internal/defs` DeprecatedPaths + `dirs_test.go` 무변경 (REQ-DC-007); github-actions.yaml + 그 allowlist row 보존.

## §E.3 Run-phase Audit-Ready Signal

AC PASS/FAIL 매트릭스 (2026-07-02 실측):

| AC | 검증 | 결과 |
|----|------|------|
| runtime.yaml 로컬 부재 | `ls .moai/config/sections/runtime.yaml` | PASS (No such file) |
| runtime.yaml 템플릿 부재 | `ls internal/template/templates/.moai/config/sections/runtime.yaml` | PASS (No such file) |
| audit_loader dead row 제거 | `grep -c runtime internal/config/audit_loader_completeness_test.go` | PASS (runtime allowlist 행 없음) |
| audit_loader 완전성 테스트 | `go test ./internal/config/` | PASS (ok) |
| DeprecatedPaths 무변경 | `go test ./internal/defs/` | PASS (ok, 43-count 핀 유지) |
| 전체 스위트 회귀 없음 | `go test ./...` | PASS (green) |

- LoadRuntime 테스트(`budget_test.go`)는 자기완결 `t.TempDir()` 픽스처 사용 — 온디스크 runtime.yaml 삭제와 무관, 무회귀.

## §E.4 Sync-phase Audit-Ready Signal

- 3-phase 소급 close: impl(`d3337cc8e`)이 draft 상태에서 병렬 세션에 의해 이미 main에 merge됨. 본 close는 plan-phase 산출물 커밋 + progress §E 증거 소급 기록 + frontmatter 전이를 정직하게 landing.
- spec.md frontmatter: `status: draft → completed`, `era: V3R6` 추가, `updated: 2026-07-02`.
- CHANGELOG.md `[Unreleased]` 항목 추가.
- 10 AC 전부 GREEN (§E.3 실측). Tier S 통합 lifecycle close (small SPEC consolidated lifecycle 허용).

### sync_commit_sha

sync_commit_sha: d3337cc8e
