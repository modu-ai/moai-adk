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

### M3 — 뮤턴트 오버레이 채득 (2026-09-08, 커밋 트리 `e365c2d30` 위 오버레이)

방법: t517 F1-D 오버레이 — 커밋 트리에 뮤턴트를 적용 → 가드 FAIL verbatim 채득 → 복원 → 같은 커맨드 PASS → `git diff --stat`로 복원 완전성 증명. 모든 채득 로그는 `.moai/reports/t544/` 아래 verbatim raw 출력이다.

**뮤턴트 A — 수리 되돌림** (absent 분기 제거, 원본 결함 형태 복원)

- 커맨드: `go test -count=1 -run 'TestPatchFileGreenfieldCreation' ./internal/settings/yamlpatch` → exit 1, 증거 `MUTANT-A-yamlpatch-fail.log` — `--- FAIL: TestPatchFileGreenfieldCreation` + `yamlpatch: stat …: no such file or directory`
- 커맨드: `go test -count=1 -run 'TestHandleSaveGreenfieldSectionCreation' ./internal/web` → exit 1, 증거 `MUTANT-A-web-fail.log` — `greenfield first-save status = 500, want 200`
- 복원 후 두 커맨드 PASS, `git diff --stat` 무출력 확인.

**뮤턴트 B — absent 브랜치 직접 비원자 쓰기** (`os.WriteFile` 교체 — REQ-4 공격)

- 커맨드: `go test -count=1 -run 'TestAtomicWriteAbsentModeUmaskIndependent' -v ./internal/settings/yamlpatch` → exit 1, 증거 `MUTANT-B-umask-fail.log` — `absent-target mode = 600 under umask 0077, want 644`. 직접 쓰기는 umask를 결과 모드에 누출하므로 M2에서 사전 설계한 판별 가드가 잡는다.
- 복원 후 PASS + `git diff --stat` 무출력.

**뮤턴트 C — 전 stat 오류 관용** (IsNotExist 체크 제거 — REQ-3 공격)

- 커맨드: `go test -count=1 -run 'TestAtomicWriteStatErrorNotWidened' -v ./internal/settings/yamlpatch` → exit 1, 증거 `MUTANT-C-enotdir-fail.log` — ENOTDIR stat 오류가 무시돼 CreateTemp 단계로 넘어가고 wrapping이 `yamlpatch: create temp: … not a directory`로 바뀜 → `yamlpatch: stat` wrapping 단정이 잡음.
- 복원 후 PASS + `git diff --stat` 무출력.

**뮤턴트 D — 기본 모드 오염** (defaultFilePerm 0644 → 0600 — REQ-2 공격)

- 커맨드: `go test -count=1 -run 'TestPatchFileGreenfieldCreation|TestYAMLPatchAtomicWriteErrors|TestAtomicWriteAbsentModeUmaskIndependent' ./internal/settings/yamlpatch` → exit 1, 증거 `MUTANT-D-mode-fail.log` — 세 가드가 모두 `mode = 600, want 644`로 잡음.
- 복원 후 PASS.

**뮤턴트 E — present 파일 모드-보존 파괴** (REQ-8 미검출 뮤턴트 기록)

- 뮤턴트: present 경로의 `mode`를 상수로 파괴 (`mode = 0o600`, `_ = info`로 컴파일 유지).
- 커맨드: `go test -count=1 ./internal/settings/yamlpatch` → **`ok` — 전체 스위트 통과**. 증거 `MUTANT-E-preserve-fail.log`. **미검출이다**: 기존 스위트에는 present 파일의 권한 모드를 단정하는 테스트가 하나도 없었다 — atomicWrite의 "원본 파일 모드를 보존한다" 계약이 무가드 상태였다는 것이 이 뮤턴트가 그린 가드의 경계다 (REQ-8).
- **경계 닫기**: `TestAtomicWritePreservesPresentFileMode` 신설 (yamlpatch_test.go, windows skip — chmod가 읽기전용 비트로 축약돼 단정 무의미). 파괴 값 0600은 기대값과 우연 일치해 가드가 통과하므로(뮤턴트 E의 값 선택이 가드와 같은 0600), 파괴 값 0777 뮤턴트 E'로 재시험: `go test -count=1 -run 'TestAtomicWritePreservesPresentFileMode' -v` → exit 1, 증거 `MUTANT-E-prime-guard-fail.log` — `present-file mode = 777, want 600 preserved`.
- 복원 후 전체 스위트 `ok` + `git diff --stat`이 `yamlpatch.go` 무변경을 증명 (남은 diff는 이 가드 추가 26줄뿐).

**AC-005 통제군 재확인**

- 커맨드: `go test -count=1 -run 'TestPatchFileValueInvariantPreservesBytes|TestPatchFileScalarChangePreservesPresentation|TestPatchFileSpliceFallsBackForUpsert|TestPatchFileSpliceQuotedScalarChange|TestApplySchemaEditsSeamRoundTrip|TestApplySchemaEditsGateSeamRoundTrip|TestApplySchemaEditsAllFieldsRoundTrip|TestYAMLPatchScalarReplace_WorkflowFixture|TestYAMLPatchPreservesQuotedStyle|TestYAMLPatchPreservesTypedScalars|TestYAMLPatchMultiEditSingleWrite|TestYAMLPatchEmptyEditsIsNoop' -v ./internal/settings ./internal/settings/yamlpatch`
- 결과: `--- PASS` 12건, `ok` 2패키지. 증거 `M3-AC005-control-group.log`. `"read-only directory"` 서브테스트도 무수정 GREEN (M2/M3 실행에서 PASS 확인).

### M4 — 범위 한정 판정 (2026-09-08)

**M4-1 scoped 테스트**: `go test -count=1 ./internal/settings/... ./internal/web/...` → `ok` 4패키지(settings 2.156s / agentfm 1.270s / yamlpatch 1.322s / web 8.416s). 증거 `M4-scoped-tests.log` (측정 시점 M3 커밋 트리). **최종 커밋 트리 `ada4eda3b`에서 동일 커맨드 재측정** → `ok` 4패키지, 증거 `M4-final-scoped-tests.log` — M4의 테스트 파일 2건 추가·수리 이후의 트리에서도 GREEN 유지.

**M4-2 크로스플랫폼 빌드**: `GOOS=windows GOARCH=amd64 go build ./internal/...` → exit 0. 증거 `M4-crossbuild-windows.log`. absent-mode 코드는 GOOS-neutral 형태(defaultFilePerm 상수 + os.Chmod — POSIX 비트를 컴파일 타임 분기 없이 다룸).

**M4-3 vet**: `go vet ./internal/settings/... ./internal/web/...` → exit 0. 증거 `M4-vet.log`.

**M4-4 lint**: `golangci-lint run --timeout=2m ./internal/settings/... ./internal/web/...` → 최종 **0 issues, exit 0** (`M4-lint.log`). NEW 이슈 1건 발생·즉시 수리: `atomicwrite_mode_unix_test.go:22 SA4032` — `//go:build !windows` 파일 안의 `runtime.GOOS == "windows"` 죽은 분기. 빌드 태그가 이미 배제하므로 분기 제거로 수리.

**M4-5 커버리지** (`M4-coverage.log`, 이 트리): yamlpatch **82.5%** / settings **90.6%** / web **67.0%**.
- 수정 함수 `atomicWrite` 72.0%(함수 단위) — 수리가 만진 stat/mode 분기(absent·ENOTDIR·present 3경로)는 전부 커버. 패키지 85% 문턱 미달 gap은 (a) `tmp.Write`/`tmp.Close`/`os.Chmod` 실패 주입 경로 3분기 — 자연 주입이 비현실적인 I/O 실패 계열로 수정 전부터 존재, (b) 미수정 함수 `setScalar` 25.0% — 본 SPEC 범위 밖.
- M4에서 rename 실패 분기를 자연 주입 테스트로 커버(`TestAtomicWriteRenameFailure` — 대상이 디렉터리면 rename 실패 + temp 잔재 없음 단정): atomicWrite 64.0%→72.0%.
- web 67.0%는 소스 변경 0(테스트만 추가)이므로 baseline 대비 감소 불가 — 추가 테스트는 커버리지를 올리는 방향으로만 작동.

**M4-6 GREEN 가드 최종 재측정**: M4 커밋 트리에서 신설 가드 전체 `-v` 재실행 → `--- PASS` 17건, `ok` 2패키지. 증거 `M4-green-guards.log` (AC-001/AC-002/AC-003/AC-004 + umask·모드보존·rename 가드 + TestYAMLPatchErrors 재작성 서브테스트).

**M4-7 E8 RED 귀속 요약**: AC-001 RED = `RED-m1-yamlpatch.log`(트리 b91372794+M1 테스트, exit 1, `yamlpatch: stat …: no such file or directory`) · AC-004 RED = `RED-m1-web-save.log`(동일 트리, exit 1, 500 + `section config write failed: yamlpatch: stat …` 배너). RED는 결함 경로 그 자체를 가리키는 옳은-이유 RED다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-08
run_commit_sha: "5fabf55d6"   # M4 verdict 커밋 — placeholder에서 backfill (D3 규약, manager-develop 소관)
run_status: complete
ac_pass_count: 6
ac_fail_count: 0
preserve_list_post_run_count: 0   # PRESERVE 표면(sectionapply.go, sectionwrite.go, PatchFile 읽기 경로, renderer, C6 게이트, 13곳 타 atomicWrite) 접촉 0 — 변경 파일은 delegation 명시 범위 4개뿐
l44_pre_commit_fetch: n/a         # 레인은 push 금지(git-flow lane protocol §4) — fetch/push 사이클 없음
l44_post_push_fetch: n/a
new_warnings_or_lints_introduced: 0   # 발생 1건(SA4032) 즉시 수리, 최종 0
cross_platform_build:
  windows_amd64: pass
  note: "compile-only — permission 비트가 windows에서 무의미한 두 테스트(umask·모드보존)는 파일 태그/in-test skip으로 GOOS-neutral 유지"
total_run_phase_files: 4   # yamlpatch.go, yamlpatch_test.go, atomicwrite_mode_unix_test.go, internal/web/write_safety_test.go
m1_to_mN_commit_strategy: per-milestone commits (M1 RED b6bd0d011 / M2 fix e365c2d30 / M3 mutants 1679c23b9 / M4 verdict)
```

## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-09-08
sync_commit_sha: "5a3db057b"   # D3 백필 확정값 — sync 본체 커밋 `docs(SPEC-SEAM-GREENFIELD-001): sync-phase artifacts — 3-phase close [t544]`
sync_status: complete
changelog_entry: CHANGELOG.md [Unreleased] `### Fixed` 첫 항목 — t517 항목에 out-of-scope로 기록됐던 관측("a missing mcp.yaml still makes a full-form save return 500")의 수리로서 서술
b12_self_test_a: pass (pre-emission grep — `grep -c 'SPEC-SEAM-GREENFIELD-001' CHANGELOG.md` = 0, 병렬 BATCH-SYNC 중복 없음)
b12_self_test_b: pass (AC count match — Tier S, AC SSOT = spec.md §3 inline이며 acceptance.md는 존재하지 않음; `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' spec.md | sort -u | wc -l` = 6, 엔트리 기재 "6 acceptance criteria (AC-001..006)"와 동일)
b12_self_test_c: pass (file path verification — 엔트리 인용 경로 전부 `ls` 실존 확인: internal/settings/yamlpatch/yamlpatch.go, internal/settings/yamlpatch/yamlpatch_test.go, internal/settings/yamlpatch/atomicwrite_mode_unix_test.go, internal/web/write_safety_test.go, .moai/reports/t544/)
frontmatter_status_transitions:
  spec.md: "in-progress → completed" (본 sync 커밋에 병합 — 3-phase close; status + updated만 변경. updated는 2026-09-08 기유)
  plan.md / acceptance.md: 해당 없음 — plan.md는 status 필드 미보유(Artifact Statelessness), acceptance.md는 미존재(Tier S)
mx_tag_pass:
  result: no-additions — run-phase M2가 이미 `@MX:NOTE: [AUTO]`(absent→create+0644 계약, REQ-1..4 문서화) + `@MX:SPEC: SPEC-SEAM-GREENFIELD-001` 서브라인을 atomicWrite에 착지했고(yamlpatch.go:368-374), defaultFilePerm 상수 주석(359-363)이 §4 기각 근거를 문서화한다. [AUTO] 접두사·code_comments(en) 준수 확인 — 중복 추가는 태그 스팸이므로 생략
sync_phase_verification:   # 측정 트리 df50ab04e — 본 sync 커밋의 변경은 .md 3파일뿐이므로 Go 소스 트리는 커밋 전후 동일하다
  go_test: pass — `go test -count=1 ./internal/settings/yamlpatch/` → `ok  github.com/modu-ai/moai-adk/internal/settings/yamlpatch 0.359s`, exit 0
  go_vet: pass — `go vet ./internal/settings/yamlpatch/` → 무출력, exit 0
