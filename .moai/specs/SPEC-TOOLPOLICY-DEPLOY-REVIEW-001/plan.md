# Implementation Plan — SPEC-TOOLPOLICY-DEPLOY-REVIEW-001

> tool-policy.yaml 사용자 프로젝트 배포 게이팅 (Direction A). Tier S.

## §A Context

`internal/template/templates/.moai/config/sections/tool-policy.yaml`(48KB/181 entries)는 모든 사용자 프로젝트에 배포되지만 소비되지 않는 orphan 파일이다. 조사(spec.md §1.1)로 검증됨:
- settings.json은 `settings.json.tmpl`에서 독립 렌더링(tool-policy.yaml 무관).
- tool-policy.yaml 소비자는 `internal/cli/tool_policy.go`(수동 CLI) 1개뿐.
- dev SSOT는 repo-root 복사본(byte-different)이며, 템플릿 복사본은 dev 빌드에서 읽히지 않음.

목표: 템플릿에서 파일을 제거(A1) 또는 최소 stub로 치환(A2)하여 사용자 프로젝트 배포에서 게이팅.

## §B Known Issues / 제약

- `TestAuditLoaderCompleteness`(`internal/config/audit_loader_completeness_test.go`)는 템플릿 sections 디렉터리를 스캔한다. 파일 제거 시 uncovered 실패는 없으나(uncovered-only 검사), `acknowledgedDedicatedLoaders`의 `"tool-policy"` 주석을 dev-only 재배치로 갱신할 것(REQ-TDR-004).
- `make build`는 오케스트레이터가 수행(gen-catalog-hashes + go build). 본 에이전트는 빌드하지 않음.
- Template-First / §15 언어 중립성 / §25 내부 콘텐츠 격리 doctrine 준수.

## §C Pre-flight (구현 전 확인)

- [ ] repo-root `.moai/config/sections/tool-policy.yaml`(dev SSOT)이 온전한지 확인 (템플릿 제거가 dev codegen을 깨지 않음).
- [ ] `catalog.yaml`에 `tool-policy` 항목이 없음을 재확인 (catalog 수정 불필요).
- [ ] `grep -rln "config/toolpolicy" --include="*.go" . | grep -v _test.go` → `internal/cli/tool_policy.go` 1건 재확인.

## §D Constraints

- settings.json / settings.json.tmpl permissions 블록 **불변**(REQ-TDR-002).
- `moai tool-policy` 명령·codegen 패키지 **불변**(REQ-TDR-003).
- 기존 프로젝트 소급 정리 **범위 밖**.

## §E Self-Verification

- `go test ./...` 전체 통과 (특히 `internal/config`, `internal/template`, `internal/cli`).
- 신규 init smoke: 임시 디렉터리 `moai init` 후 `tool-policy.yaml` 부재 + `settings.json` permissions 존재 확인.
- `moai tool-policy list` (repo-root) 정상 동작 확인.

## §F Milestones (우선순위 순, 시간 추정 없음)

- **M1 — A1/A2 하위결정 확정 (오케스트레이터/사용자 확인)**: spec.md §5의 A1(완전 제거) vs A2(최소 stub 치환) 중 선택. `moai tool-policy list`의 신규-프로젝트 사용성(EC-2/R1)이 결정 요인. 확정 전 구현 착수 금지.
- **M2 — 배포 게이팅 적용**:
  - A1 선택 시: `internal/template/templates/.moai/config/sections/tool-policy.yaml` 삭제.
  - A2 선택 시: 동 파일을 최소 예시 stub(수 개 엔트리 + 주석, 내부 181-entry 제거)로 교체.
  - `audit_loader_completeness_test.go`의 `acknowledgedDedicatedLoaders` 주석 갱신(REQ-TDR-004).
- **M3 — 사용자 커스터마이징 경로 문서화 (REQ-TDR-005)**: 권한 커스터마이징 권장 경로를 한 곳(예: docs-site 또는 README)에 명시. A1 선택 시 `moai tool-policy` 사용성이 dev-repo 한정임을 함께 명시.
- **M4 — 검증**: §E Self-Verification 전 항목 실행 + `make build`(오케스트레이터).

## §G Anti-Patterns (회피)

- 템플릿과 repo-root tool-policy.yaml를 혼동해 dev SSOT를 삭제하는 것 — 절대 금지(REQ-TDR-003).
- settings.json.tmpl permissions 블록을 함께 건드리는 것 — 범위 밖(REQ-TDR-002).
- 기존 프로젝트 파일 pruning 로직을 이 SPEC에 끼워넣는 것 — 별도 SPEC(R2).
- `moai tool-policy` 명령/codegen 패키지 수정 — 범위 밖.

## §H Cross-References

- spec.md §1.1 (조사 증거 표), §5 (결정·잔여위험).
- `internal/cli/init.go` `runInit`, `internal/template/deployer.go` `Deploy`, `internal/template/slim_fs.go` (배포 경로).
- `internal/cli/tool_policy.go`, `internal/config/toolpolicy/{loader,codegen}.go` (유일 소비자).
- `internal/config/audit_loader_completeness_test.go` (allowlist 갱신 대상).
- CLAUDE.local.md §2 (Template-First / settings 분리), §15 (언어 중립성), §25 (내부 콘텐츠 격리), §21 (dev-only 경로 패턴).
- 원 SPEC-V3R6-TOOL-POLICY-SSOT-001 (tool-policy codegen SSOT 도입).
