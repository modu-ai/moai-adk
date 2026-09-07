# SPEC-SEAM-GREENFIELD-001 Progress

카드 t544 — seam greenfield 첫 저장 500 결함 (absent 섹션 파일 원자적 기록의 stat 부재-불내성).

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-08
plan_artifacts:
  - .moai/specs/SPEC-SEAM-GREENFIELD-001/spec.md
  - .moai/specs/SPEC-SEAM-GREENFIELD-001/plan.md
plan_baseline_tree: "52f863f36"   # plan-phase 산출물 저작 기준 트리 (WT-save-absent-file)
tier: S
notes: "근거 앵커 8곳 spec.md §1.2 직접 확인(트리 52f863f36). 카드 전제 정정 1건 — 결함을 인코딩한 기존 서브테스트(plan.md §B-a, AC-003). 통제군 존재 오판정 수리 1건 — TestPatchFileValueInvariantPreservesBytes는 write_safety_test.go:29에 실재(plan.md §B-b, AC-005, HISTORY). run-phase 착수 시 §C content-token 재검증 선행."
```

## §F Phase 4 Mode Selection

**Implementation Kickoff Approval**: 통과 — 리드 경유 운영자 결재 접수(2026-09-08, 리드 세션 교체 직후). 레인 세션의 직접 질문은 거절됐었다가 리드 채널로 승인이 접수된 경위를 함께 기록한다.

**Plan Audit Gate skip 결정**: 스킵한다. 3조건 전부 충족 — (1) verdict PASS, (2) score 0.94 ≥ Tier S 문턱 0.75, (3) 산출물 해시 불변(감사 트리 = 현재 HEAD `5d8411927`, 감사 직후 plan-artifact 편집 0).

```yaml
tier: S
scope_files: 4   # yamlpatch.go + yamlpatch_test.go + write_safety_test.go(옵션) + internal/web 가드 테스트
domain_count: 2  # internal/settings/yamlpatch + internal/web
language_mix: Go 100%
concurrency_benefit: low   # coding-heavy — Anthropic 코딩 병렬화 주의사항 적용
```

| Mode | 선택 | 근거 |
|------|------|------|
| direct | 아니오 | 코드 변경 수반 — 위임 소관 |
| **serial** | **예** | coding-heavy 단일 도메인 수리 — 마일스톤당 1 스폰 기본값 |
| fanout | 아니오 | 다중 도메인 리서치 아님 |
| sweep | 아니오 | 기계적 대량 변형 아님 |

**Decision: serial** — manager-develop 단일 스폰, M1→M4 직렬. 코딩 과업의 병렬화 주의사항(Anthropic)에 따라 기본 폴백 선택. 진행 모드: semi-autonomous(리드 지정 — 마일스톤마다 보고, goal 미무장).

## §E.2 Run-phase Evidence

### M1 — RED 가드 채득 (2026-09-08, 트리 `b91372794` + M1 테스트 파일 미커밋 상태)

**M1-1 — AC-001 yamlpatch 단위 greenfield (RED-now cell)**

- 커맨드: `go test -v -count=1 -run 'TestPatchFileGreenfieldCreation|TestAtomicWriteStatErrorNotWidened' ./internal/settings/yamlpatch`
- exit code: `1`
- 증거: `.moai/reports/t544/RED-m1-yamlpatch.log` (verbatim raw 출력)
- RED 이유(옳은-이유 RED): `--- FAIL: TestPatchFileGreenfieldCreation` — `yamlpatch: stat …mcp.yaml: no such file or directory`. 결함 그 자체(C3 stat 부재-불내성 경로)다.
- 동반 가드 `TestAtomicWriteStatErrorNotWidened`(AC-002, REQ-3 회귀 가드)는 수리 전 트리에서 PASS — absent 외 stat 오류(ENOTDIR)는 현행에서도 오류로 반환되므로 RED가 아니라 유지돼야 하는 가드다.

**M1-2 — AC-004 웹 레벨 greenfield 첫 저장 (RED-now cell)**

- 커맨드: `go test -v -count=1 -run 'TestHandleSaveGreenfieldSectionCreation' ./internal/web`
- exit code: `1`
- 증거: `.moai/reports/t544/RED-m1-web-save.log` (verbatim raw 출력)
- RED 이유(옳은-이유 RED): `greenfield first-save status = 500, want 200` — 배너는 `section config write failed: yamlpatch: stat …mcp.yaml: … no such file or directory`로 stat 오류를 직접 가리킨다(AC-004 vacuous-green 방지 요건 — 제출이 C6 게이트를 통과해 PatchFile에 도달했음의 증거).

**M1 커밋 범위**: `internal/settings/yamlpatch/yamlpatch_test.go`(AC-001 + AC-002 가드), `internal/web/write_safety_test.go`(AC-004 가드), spec.md `draft → in-progress` 전환.

### M2 — 수리 (2026-09-08, 커밋 트리 기준 기술)

**M2-1 — 수리 본체** (`internal/settings/yamlpatch/yamlpatch.go`)

- `atomicWrite`의 무조건 stat-fail을 분기: stat 오류가 `os.IsNotExist`를 만족하면 패키지 단일 정의점 `defaultFilePerm = 0644`(spec.md §4)로 진행, 그 외 stat 오류는 기존 `yamlpatch: stat %s: %w` 래핑 유지 (REQ-1/REQ-2/REQ-3). Chmod는 `mode` 변수(원본 보존 또는 absent 기본)를 쓴다 — temp+rename 구조 무변경 (REQ-4).
- `@MX:NOTE` + `@MX:SPEC: SPEC-SEAM-GREENFIELD-001` 태그로 absent 계약 문서화.

**M2-2 — AC-003 기대 전환**: `TestYAMLPatchAtomicWriteErrors/"stat missing target"`을 absent greenfield 생성 성공 + 0644 + 내용 반영 기대로 재작성. 형제 `"read-only directory"` 서브테스트는 무수정 GREEN 유지 (수리 후 실행에서 PASS 확인).

**M2-3 — plan-phase 식별 누락 발견·수리**: M2 회귀 실행(`go test ./internal/settings/... ./internal/web/...`)에서 `TestYAMLPatchErrors/"missing file"`(`yamlpatch_test.go:243`)이 absent PatchFile 오류를 기대하는 **결함-인코딩 테스트 2번째 사례**로 발견됐다 — plan-phase C8이 이 곳을 놓쳤다(수리 후 RED로 남아 수리를 가렸을 것이다). AC-003과 동일 클래스로 수리와 함께 absent greenfield 생성 기대로 재작성했다. SPEC body 수정 불요 — 재작성 파일은 delegation 명시 범위 안(`yamlpatch_test.go`)이고 방향은 SPEC 수리 방향과 동일하다.

**M2-4 — 뮤턴트 B 판별 가드 신설**: `internal/settings/yamlpatch/atomicwrite_mode_unix_test.go`(`//go:build !windows`) — `TestAtomicWriteAbsentModeUmaskIndependent`: umask 0077에서도 absent 생성 모드가 0644임을 단정. temp+rename+chmod는 umask가 결과에 스며들지 않지만 absent 브랜치가 직접 `os.WriteFile`로 대체되는 뮤턴트는 0600을 남겨 이 가드에 잡힌다 (M3 뮤턴트 B의 사전 설계된 판별 증거; REQ-4의 관측 가능한 형태). umask는 프로세스 전역이므로 비병렬 테스트로 작성 — 병렬 테스트들은 비병렬 완료 후 재개되므로 창이 겹치지 않는다.

**M2 GREEN 확인**: `go test ./internal/settings/... ./internal/web/...` → `ok` 4패키지 (이 트리, M2 적용 후). M1의 두 RED 가드(TestPatchFileGreenfieldCreation / TestHandleSaveGreenfieldSectionCreation) 모두 PASS로 뒤집힘 — verbatim은 M4 §E.1 최종 판정에서 재측정해 귀속한다.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop 소관.>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관. sync_commit_sha는 sync 커밋 시 채워진다(pending-backfill 규약).>_
