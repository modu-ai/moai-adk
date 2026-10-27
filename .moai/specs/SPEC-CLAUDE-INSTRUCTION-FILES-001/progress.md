# progress.md — SPEC-CLAUDE-INSTRUCTION-FILES-001

> 카드 t1985, lane-16. §E 섹션 골격 — §E.1만 plan-phase에서 작성, §E.2~§E.4는 placeholder.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: pending-audit
plan_complete_at: 2026-10-27
artifacts: spec.md (16 REQ, GEARS), plan.md (M0~M5), acceptance.md (AC-CIF-001~016, GUIDE-01..12 조정표, RED-now 기준선 @cb30e05b3e)
run_gate: iter2+ 독립 감사 PASS(최신 plan HEAD에 결속) + 리더 킥오프 판정까지 run 금지; t1984는 본 카드 merge-ready 뒤 임대·plan, 본 카드의 PR 병합·main 흡수가 t1984 RUN에 선행 (plan.md §C, 리더 §15.124 확정 순서)

## §E.2 Run-phase Evidence

### M0 — run 진입 게이트 (측정 트리 `746b8af14e`, 커밋 전, 2026-10-27)

- Pre-flight: `go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 · golangci-lint exit 1 = 사전 존재 기준선 1건(`internal/discovery/leader_readers_darwin.go:93` SA1019, 본 카드 비관련) · t1984 소유 3파일 존재·비편집 확인.
- RED-now 재측정: acceptance.md §D.1 10행 + 행동 행 전부 verbatim 부합(행 1 `--- PASS: TestCodexOnlyForceUpdateVariant` = AGENTS 배치 단언 → CLAUDE-정규 기준 red; 행 5/9/10 exit 1 표면 부재). 리더 파일 목록 승인(2026-10-27T17:1x) 후 M1 착수.

### M1 — 정규 이름·생성 경로 전환 (RED→GREEN, 2026-10-27)

- **E8 RED (구현 전환 전, 트리 `746b8af14e`+테스트 재타겟)**: `TestCodexOnlyForceUpdateVariant` FAIL(`CLAUDE.md missing from force-update codex deploy` + `AGENTS.md deployed … must not be created`) · `TestCodexOnlyDeployerWalkIntegrity` FAIL(`retired AGENTS instruction file "AGENTS.md" visible` + `"CLAUDE.md" missing from listing`) · `TestHarnessProfilesResolveSharedReferences` FAIL(3 프로파일 전부 CLAUDE.md 부재/AGENTS.md 존재) · `TestValidate/existing_CLAUDE.md_triggers_warning` FAIL(`expected warning containing "CLAUDE.md already exists", got […]`) · `TestValidateMoAI_MissingClaudeMD` FAIL(`expected warning about missing CLAUDE.md, got [.claude/ … AGENTS.md not found…]`) · `TestDeployedWiringIsGPTReadsTheClaudeSurface/codex-only_layout` FAIL(`deployedWiringIsGPT = false, want true`) · `TestInitCmd_NonInteractiveExecution` FAIL(`expected CLAUDE.md to be created` + `AGENTS.md must not be created by init`). 대칭적 2건(`TestInitResumeRefusesGPTLayoutWithoutRecord`, `TestInitResumeDoesNotInferBothFromInstructionFile`)은 현행 green — 회귀 가드.
- **전환**: `defs/files.go` 상수 역할 반전(둘 다 유지) · `templates/AGENTS.md.tmpl`→`CLAUDE.md.tmpl` git mv · `embed_manifest_gen.go` :647 재생성(make embed-manifest, diff 1행) · `harness_fs.go` codexHiddenPaths에서 CLAUDE.md 제외 + normalizeHarnessReferences AGENTS→CLAUDE 양쌍 반전 · `validator.go` CLAUDE.md 존재 검사/AGENTS.md 불활성 · `init_resume.go` deployedWiringIsGPT 프로브 CLAUDE.md 축 · `initializer.go` 주석.
- **따라옴(이름 참조)**: template 패키지 테스트 11파일 + agentemit 테스트 2파일 + `obligations.yaml` claude_path 2행 + config/constitution 테스트 이름 참조. `token_budget_guard.go:141` 미러 경로만 CLAUDE.md.tmpl로(루트 축 :131은 M3/M4와 함께 — 조기 전환 시 이 저장소 측정 소멸). zone-registry `file: AGENTS.md` 14행 전환은 시도 후 **보류 복원** — registry 전환은 저장소 루트 CLAUDE.md(=M4 verb 전환)를 요구하므로, M1 중간 상태에서는 `registry_sync_test.go`의 배포 미러 매핑(rel AGENTS.md → source CLAUDE.md.tmpl)으로 브리지. M4에서 registry 플립+브리지 제거로 종결.
- **GREEN**: `go test ./internal/template/ ./internal/template/agentemit/` ok(369.9s/0.6s) · `./internal/core/project/` ok · `./internal/config/ ./internal/constitution/ ./internal/cli/harness/` ok · cli 대상(init/resume/headroom/import-resolution) ok. `go build ./...` exit 0 · windows build exit 0 · lint = 기준선 1건 동일(NEW 0).
- **Gaps**: internal/cli 전체 스위트 실행은 M5 총괄로 연기(리소스 규율) — 대상 실행으로 방향 민감 표면 커버.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
