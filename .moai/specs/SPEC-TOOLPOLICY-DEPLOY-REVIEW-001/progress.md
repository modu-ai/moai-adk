# Progress — SPEC-TOOLPOLICY-DEPLOY-REVIEW-001

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase 산출물: spec.md + plan.md + progress.md (Tier S, AC는 spec.md §3 inline).
- SPEC ID pre-write self-check: `SPEC ✓ | TOOLPOLICY ✓ | DEPLOY ✓ | REVIEW ✓ | 001 ✓ → PASS`.
- Frontmatter: 12 canonical fields + `tier: S` + `era: V3R6`.
- 조사 완료(spec.md §1.1 7-row 증거표): settings.json은 .tmpl에서 독립 렌더링, tool-policy.yaml 소비자는 `internal/cli/tool_policy.go` 단 1개, 템플릿 복사본은 orphan SSOT, 배포 단언 테스트 0건.
- 권고 방향: **Direction A** (dev-only 게이팅); A1(완전제거)/A2(최소 stub) 하위결정은 run-phase M1에서 확정.
- Out of Scope 3개 H3 섹션 명시(settings.json 재설계 / CLI 제거 / 기존 프로젝트 소급 정리).
- status: draft. plan-auditor 게이트 및 구현 착수 승인 대기.

## §E.2 Run-phase Evidence

- **A1/A2 하위결정 → A1 (완전 제거) 확정**: orphan 파일(사용자 런타임 소비자 0), A2 stub은 혼란스러운 부분 콘텐츠라 배제. 오케스트레이터가 독립 검증한 5개 근거(소비자 1개 `internal/cli/tool_policy.go` / catalog 0건 / settings.json.tmpl 독립 permissions 렌더 / repo-root dev SSOT 별도 보존 48,420 bytes)로 A1이 최소 올바른 선택임 확인.
- **구현**: `internal/template/templates/.moai/config/sections/tool-policy.yaml` (48,315 bytes / 181 entries) 삭제.
- `internal/config/audit_loader_completeness_test.go` `acknowledgedDedicatedLoaders`의 `tool-policy` 주석을 dev-only 재배치로 갱신(엔트리 유지 — 테스트가 스캔된 파일과만 대조하고 역방향 검사 없어 무해; REQ-TDR-004).
- settings.json.tmpl permissions 블록 / `moai tool-policy` 명령 / codegen 패키지 불변(REQ-TDR-002/003).

## §E.3 Run-phase Audit-Ready Signal

- **AC-TDR-001a PASS**: `go run ./cmd/moai init --all` (fresh embed) → 신규 프로젝트에 `tool-policy.yaml` 부재 실측(`No such file`).
- **AC-TDR-001b PASS**: 신규 init `.claude/settings.json`에 `permissions` 블록 존재; settings.json.tmpl 불변이므로 byte-identical 보장(구조적).
- **AC-TDR-002 PASS**: `go test ./internal/config/... ./internal/template/...` GREEN (TestAuditLoaderCompleteness 포함).
- **AC-TDR-003 PASS**: `go run ./cmd/moai tool-policy list` (repo-root dev SSOT) → 182줄 = 헤더 1 + **181 엔트리** (dev codegen 능력 보존).
- **AC-TDR-004 PASS**: `grep -rln config/toolpolicy --include=*.go` → 여전히 `internal/cli/tool_policy.go` 1건 (신규 런타임 소비자 미추가).
- **REQ-TDR-005 (SHOULD, non-AC)**: 권한 커스터마이징 권장 경로 = 표준 Claude Code `/permissions` + `.claude/settings.json` `permissions` 블록 직접 편집(inherent CC 기능 — 신규 doc surface 불요; settings.json.tmpl 불변 제약상 주석 추가도 배제). A1하에서 `moai tool-policy`는 dev-repo 한정.

## §E.4 Sync-phase Audit-Ready Signal

- 3-phase close (consolidated Tier S, orchestrator-direct sync): status `draft → completed`, era `V3R6`, updated `2026-07-02`.
- CHANGELOG `### Removed` 엔트리 추가.
- sync_commit_sha: 7d9bc4147c45ed8344c19535bd9a0bba5150c5ad
- MX Tag: Tier S consolidated — MX는 sync sub-step, 별도 Mx 커밋 없음(3-phase close).
