# SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001 — Progress

> §E lifecycle 신호 skeleton. plan-phase는 §E.1만 채운다. §E.2/§E.3은 run-phase(manager-develop), §E.4는 sync-phase(manager-docs)가 채운다.

## §E.1 Plan-phase Audit-Ready Signal

- 산출물: spec.md + plan.md + acceptance.md + progress.md (4 plan-phase artifacts).
- SPEC ID Pre-Write Self-Check: `decomposition: SPEC ✓ | V3R6 ✓ | DEV ✓ | HARNESS ✓ | CONSOLIDATION ✓ | 001 ✓ → PASS`.
- Frontmatter 12-field schema 검증 완료 (status: draft, module: .claude/commands/harness, priority: P2, phase: v3.0.0, tags: harness/dev-only/consolidation).
- GEARS REQ: REQ-DHC-001 ~ REQ-DHC-007 (7 scope 항목 커버).
- Out of Scope: §J에 4개 `### Out of Scope —` H3 (Go 코드 / 사용자 템플릿 / capability 확장 / memory 직접 작성).
- 하네스 이름 결정: `devkit` (정당화: spec.md §A.2).
- 핵심 설계 결정: Runner/human-gate 정합 (plan.md §B.1) — Runner는 비-상호작용 fan-out만, 사람-게이트는 specialist 위임.
- plan-auditor iter-2 (PASS-WITH-DEBT 0.83) defect 8건 반영 (v0.2.0): D1(BLOCKING) CI-guard re-anchor (dev_only_skill_test.go=skills-only walker → embedded_namespace_test.go 패턴 embedded-tree-absence 단언; §B "유일 보호 패턴" 오류 정정), D2 tier:S frontmatter, D6 §F-sync에 CLAUDE.local.md §2 추가, D8 AC-007a/b negative-proof 강화, D3/D4/D5/D7 부수 정정.

## §E.2 Run-phase Evidence

- Run-phase 구현 commit: `96fad88ff` (origin/main, plan→run).
- GENERATE: `.claude/commands/harness/devkit.md` + `manifest.json`(3 specialist, primitive=sub-agent) + `.claude/workflows/harness-devkit-run.js`(비-상호작용 fan-out only) + `.claude/agents/harness/harness-devkit-{release-update,github,release}-specialist.md`.
- DELETE: `97-release-update.md`/`98-github.md`/`99-release.md` + `agents/local/{release-update,github}-specialist.md` + `skills/moai/workflows/release.md` (956 LOC).
- CI guard: `internal/template/devkit_namespace_test.go` (`TestDevkitNamespaceNoLeak` — embedded-tree-absence).
- doctrine: `dev-only-commands-isolation.md` + `CLAUDE.local.md` §2/§21 + `skill-authoring.md`/`INDEX.md`/`reference.md` 참조 정합.
- 회귀 수정: `TestRootLevelCommandsThinPattern` zero-guard 완화 (root 통합 후 공백 정상).
- 구현 경로: manager-develop L1 worktree(rate-limit 중단) → orchestrator 정밀 이식 + test guard fix + 8 AC 독립 재검증.

## §E.3 Run-phase Audit-Ready Signal

- AC 8/8 PASS (AC-DHC-001~006 + 007a/007b) — orchestrator 독립 검증.
- cross-platform build: host exit 0 + `GOOS=windows GOARCH=amd64` exit 0.
- `go test ./internal/template/...` : ok (GREEN). `go vet` exit 0.
- AC-007a: Runner 주석 제외 `AskUserQuestion`/`gh` 매치 0. AC-007b: manifest 3 specialist 전부 `sub-agent` primitive.
- 템플릿 누출 0 (`find internal/template/templates *harness-devkit*`/`*commands/harness*` empty).
- run_commit_sha: 96fad88ff

## §E.4 Sync-phase Audit-Ready Signal

- 3-phase close: plan(`60f9b3401`) → run(`96fad88ff`) → sync(이 커밋).
- 독립 품질 검증 (sync-auditor rate-limit → orchestrator-direct 대체): 8/8 AC PASS 재확인 + 포팅 충실도(release 956L→169L = 8-phase Enhanced GitHub Flow 전 구조 보존, verbose 예시 제거이지 단계 손실 아님) + doctrine staleness 0(skill-authoring.md migration table만) + Runner 정직 설계(release/github=specialist-held, Runner 미모델링).
- 사용자-facing 문서 영향 없음 (dev-only 내부 도구 — CHANGELOG/README/docs-site 변경 0).
- 상태 전환: in-progress → completed (이 sync 커밋이 close 운반).
- sync_commit_sha: <pending-backfill>
- era: V3R6 (H-4: §E.2 + §E.4 + sync_commit_sha).
