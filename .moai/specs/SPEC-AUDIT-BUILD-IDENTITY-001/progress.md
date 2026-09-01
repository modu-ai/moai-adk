# SPEC-AUDIT-BUILD-IDENTITY-001 — 진행 기록

카드 t248 · 워크트리 `WT-audit-binary-sha` (base `64bba61aa`)

## §E.1 Plan-phase Audit-Ready Signal

- 산출물: `spec.md`(요구 8, v0.1.2), `acceptance.md`(수락 8), `plan.md`, 이 파일
- status: `draft`
- 미해결: 없음. plan.md §B의 열린 질문 2건은 운영자 결정으로 D1(평탄 형제 필드 `build_commit`/`build_lag`, 중첩·버전 필드 없음)·D2(지연 권고도 형제 필드)로 닫혔다 — spec.md v0.1.1
- 구현 0줄. 커밋 없음

### plan-audit 결함 수리 라운드 (v0.1.2, 2026-09-01)

plan-audit 판정 FAIL(점수 0.85, 반복 1/1, `.moai/reports/t248/plan-audit.md`) → 차기 세션이 결함별로 디스크 상태를 먼저 검증하고 미착지분만 적용했다. 선행 세션이 적용한 분은 재적용하지 않았다.

| 결함 | 처분 |
|---|---|
| D-1 (소스 스윕 기대가 거짓) | 선행 세션이 착지. 본 세션이 같은 grep을 본 트리(`64bba61aa`)에서 재실행해 히트 3건(`graph_stamp.go:68`/`:131`, `mcp_review_material.go:95`)으로 기준선 재확인 — exit 0. 기대 절은 기준선 3좌표 **정확 집합** 술어 + `resolveReviewMergeBase`는 diff 기준점 해석임을 명시하는 문장을 갖는다 |
| D-2 + D-5 (공허화 + Tier S 상한 9>8) | 선행 세션이 착지. 모양 기준을 AC-ABI-001에 병합(`build_commit` 비어있지 않음 선행 전제) — 수락 9→8, 생존 AC 재번호 001..008, §2.1/§3/plan.md 상호 참조 정합 확인 |
| D-3 (빈 `projectRoot`에서 REQ-ABI-006 도달 불가) | 선행 세션이 착지. option (a) — `os.Getwd()` 폴백(`doctor.go:521` 선례). REQ-ABI-006 본문 + plan.md §B D3 + AC-ABI-006 빈-root 경로 절(`StatusBehind` 스텁에서 `build_lag` 비어있지 않음) 모두 존재 |
| D-4 (지연 AC가 진입점 범위 미구속) | 선행 세션이 착지. AC-ABI-006이 세 핸들러 전수 + 테이블 구동 [HARD] + `StatusFresh` 대조군 |
| D-6 (§1.4가 [HARD] 픽스처 규율의 실제 사냥감 미명명) | 선행 세션이 착지. §1.4 말미 문단이 M2(필드 이름은 `build_commit`, 값은 버전)를 명명 |
| D-7 (좌표 3건 드리프트) | 선행 세션이 착지. 본 세션이 각 정정 좌표를 본 트리에서 개 줄 검증 — `mcp_codex.go:1493`, `mcp_glm.go:245`(둘 다 `resolveToolProjectRoot(req)`), `pkg/version/version.go:32`(인용)/`:37`(`func GetBuildID()`) 전부 일치 |
| D-8 (`commit=="none"`에서 `build_commit` 값 미특정) | **본 세션이 보강.** AC-ABI-005 절 4를 `version.Commit` ∈ {`""`, `"none"`, `"unknown"`} 전 집합으로 확대(`binlag.go:108-110` 선례) |
| (부수) plan.md 낡은 참조 | **본 세션이 수리.** §B D5의 기준 수(9→8)와 반뮤턴트 기준 번호(004→003), §F의 회귀 가드 기준 번호(009→008)를 병합 후 번호로 정렬 — 병합 재번호 누락분 |
| (부수) spec.md frontmatter/HISTORY | **본 세션이 착지.** version 0.1.1→**0.1.2**, HISTORY v0.1.2 행 추가, updated 2026-09-01 유지 |

불변 확인(본 세션 관측): 평탄 형제 필드 유지(중첩 0, 버전 필드 0), `internal/binlag.Evaluate` 재사용 요구 유지, D-1 기준선 3좌표 유지. 미해결 마커(`[NEEDS` 로 시작하는 주석 토큰)와 낡은 AC 번호 잔존 참조 모두 grep 매치 0 — 증거는 아래 검증 배치에.

### plan-audit 판정 경로 — 리드 판독 갈음 (2026-09-01)

plan-audit FAIL(반복 1/1, Tier S 상한 소진) → 수리 v0.1.2 → 리드가 **재감사 없이 판독으로 갈음** 판정(수리 완료 인정). 반복 상한 1은 소진 상태 그대로 두고 판정 경로만 갈음했다.

리드 실측 근거(트리 `311b1936b`):

- 구조: AC 집합 = `AC-ABI-001..008` 정확히 8개(`009`·"9개 기준" 참조 grep 0건, exit 1) — D-5 해소
- D-2: AC-ABI-001 `[HARD]` 비어있지 않음 선행 전제 존재(`acceptance.md:31`) + 공허화 메커니즘 설명(`:39`)
- D-3: `os.Getwd()` 폴백 REQ-ABI-006 명시(`doctor.go:521` 선례) + 비전달 계약이 §D.1 체크리스트 항목으로 기계화(`performCodexAudit`/`performGLMAudit` 인자 `git diff`)
- D-1: AC-ABI-008(구 009 회귀 건이 아니라 소스 스윕 건은 AC-ABI-007)이 거짓 기대에서 기준선 3좌표 정확 집합 술어로 교체
- D-6: M2 뮤턴트 명명 확인(`acceptance.md:72`)

lane이 수정 에이전트의 "이미 착지" 처분표 라벨을 불신했던 단서와 리드의 독자 실측이 서로 합치한다 — 판정 근거는 라벨이 아니라 실측이다. lane 독립 스팟체크 9항목(D-1 기준선 본 트리 재측정 포함)도 전부 합치.

산출물 커밋: `311b1936b` (SPEC 4종 + 감사 보고서, 5파일 682줄). run 진입은 Implementation Kickoff 승인 후.

## §F Phase 4 Mode Selection

- 입력: tier S / scope 4-6파일(`mcp_codex.go`·`mcp_convergence.go`·신규 테스트) / 도메인 1(Go backend) / 병렬 이득 낮음(코딩 중심) / Agent Teams 미요청
- 평가: direct 미해당(구현 존재) · fanout 미해당(단일 도메인 코딩 작업) · sweep 미해당(기계적 대량 변형 아님) · serial 해당(단일 스콥 구현)
- Decision: serial
- 근거: 코딩 중심 단일 도메인 구현 — Anthropic coding-task parallelism caveat상 serial이 기본이며, Tier S 단일 스콥이라 마일스톤 병렬화 이득이 없다.
- 실행 모델: manager-develop GLM 상속(opus 주간 한도 2026-09-04 19:00 리셋 대기 — 프로필 이격, 완료 보고에 기록)
- Kickoff: 운영자 승인(자율 모드), 리드 경유 전달. 판정 경로는 리드 판독 갈음(§E.1 기록). run 진입은 이 기록 이후 첫 spawn.

## §E.2 Run-phase Evidence

run 커밋: `1c3adc4d5` (branch `WT-audit-binary-sha`, 미푸시 — 통합은 리드 창). 아래 모든 판정은 **그 커밋 트리에서 재측정**한 값이다(커밋 시점 재측정 — 커밋 전 예비 측정과 구분).

### E1 AC 매트릭스 (8/8 PASS)

| AC | 판정 | 명령 (공통 접두: `go test ./internal/cli/ -run <TEST> -count=1`) | 실제 출력 (커밋 `1c3adc4d5`) |
|---|---|---|---|
| AC-ABI-001 | PASS | `-run TestAuditVerdictCarriesBuildCommit` | `ok github.com/modu-ai/moai-adk/internal/cli 1.832s` |
| AC-ABI-002 | PASS | `-run TestPersistedConvergenceCarriesBuildCommit` | `ok github.com/modu-ai/moai-adk/internal/cli 0.752s` |
| AC-ABI-003 | PASS | `-run TestBuildIdentityVersionAloneIsRejected` | `ok github.com/modu-ai/moai-adk/internal/cli 1.676s` |
| AC-ABI-004 | PASS | `-run TestBuildIdentityOmittedWhenAbsent` | `ok github.com/modu-ai/moai-adk/internal/cli 0.738s` |
| AC-ABI-005 | PASS | `-run TestAuditCompletesWithoutBuildIdentity` | `ok github.com/modu-ai/moai-adk/internal/cli 2.600s` |
| AC-ABI-006 | PASS | `-run TestAuditLagAdvisoryNamesBothCommits` | `ok github.com/modu-ai/moai-adk/internal/cli 2.497s` |
| AC-ABI-007 | PASS | `-run TestAuditLagUsesBinlagSeam` + 소스 스윕 | `ok ... 1.347s` + 스윕 히트 = 기준선 3좌표 정확 집합(`graph_stamp.go:68`/`:131`, `mcp_review_material.go:95`) — 아래 E-스윕 |
| AC-ABI-008 | PASS | `-run 'TestConverge\|TestRunMultiAudit\|TestAuditMulti' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 0.753s` — 기존 테스트 무수정 통과 |

### RED 증거 (E8 — GREEN 이전 관측, 구현 전 트리 `fd26c6cf2`)

```
$ go test ./internal/cli/ -run 'TestAuditVerdictCarriesBuildCommit|...|TestAuditLagUsesBinlagSeam' -count=1
--- FAIL: TestAuditVerdictCarriesBuildCommit (0.29s)
    mcp_build_identity_test.go:281: codex_audit: "build_commit" key ABSENT — the verdict carries no build identity
--- FAIL: TestPersistedConvergenceCarriesBuildCommit (77.27s)
    mcp_build_identity_test.go:323: returned ConvergenceResult: "build_commit" key ABSENT — the verdict carries no build identity
--- FAIL: TestBuildIdentityVersionAloneIsRejected (0.29s)
    mcp_build_identity_test.go:364: codex_audit build A: "build_commit" key ABSENT — the verdict carries no build identity
--- FAIL: TestBuildIdentityOmittedWhenAbsent (0.00s)
    (테스트 버그 — 정렬되지 않은 기대 집합 비교. 정렬 후 재작성, 변경 전 상태 GREEN 확인 = 회귀 가드 성격 부합)
--- FAIL: TestAuditLagAdvisoryNamesBothCommits (0.30s)
    mcp_build_identity_test.go:511: codex_audit: "build_lag" empty/absent on a StatusBehind build — the lag advisory never fired
--- FAIL: TestAuditLagUsesBinlagSeam (0.62s)
    mcp_build_identity_test.go:573: codex_audit/glm_audit/audit_multi: counter 0 → 0 — the comparison never ran through the seam
FAIL	github.com/modu-ai/moai-adk/internal/cli	81.752s
```

AC-ABI-004는 회귀 가드(변경 전 상태를 단언)라 RED가 구조적으로 불가능하다 — 최초 1회는 테스트 결함(기대 미정렬)으로 적색이었고 수정 후 변경 전 트리에서 GREEN을 관측했다.

### E-스윕 (AC-ABI-007 관측 2)

```
$ grep -rn "merge-base\|is-ancestor" internal/cli --include='*.go' | grep -v _test.go
internal/cli/graph_stamp.go:68:  moai graph stamp codemaps --commit "$(git merge-base HEAD origin/main)"
internal/cli/graph_stamp.go:131:		`explicit commit anchor ... Use "$(git merge-base HEAD origin/main)" ...`)
internal/cli/mcp_review_material.go:95:		out, err := runReviewGit(root, "merge-base", ref, "HEAD")
```

히트 집합 = 기준선(64bba61aa 실측) 3좌표와 정확히 동일. `mcp_review_material.go:95`는 리뷰 diff 기준점 해석(`resolveReviewMergeBase`)으로 D-1 판정대로 기준선에 남는다.

### E2 크로스 플랫폼 빌드 (커밋 `1c3adc4d5`)

```
$ go build ./...                           → native_exit=0
$ GOOS=windows GOARCH=amd64 go build ./... → windows_exit=0
$ go vet ./internal/cli/... ./internal/binlag/... → vet_exit=0
```

### E3 커버리지 (커밋 `1c3adc4d5`)

```
$ go test -cover ./internal/cli/ ./internal/binlag/ -count=1
ok  github.com/modu-ai/moai-adk/internal/cli   285.720s  coverage: 80.0% of statements
ok  github.com/modu-ai/moai-adk/internal/binlag  3.071s  coverage: 90.9% of statements
```

변경 대상 함수 커버리지 (`go tool cover -func`): `auditBuildIdentity` 100%, `normalizeBuildCommit` 100%, `handleCodexAudit` 100%, `handleGLMAudit` 96%, `runMultiAudit` 100%. 패키지 전체 internal/cli 80.0%는 기존 수준이며(대형 패키지, 변경 전 baseline 미측정 — Gap 기록) 이 카드가 만진 함수는 전부 96% 이상.

### E4 서브에이전트 경계 grep

```
$ grep -rn 'AskUserQuestion' internal/cli internal/binlag --include='*.go' | grep -v "_test.go" | grep -v "// "  → 18 matches
```

18건 전부 **기존 baseline**(문서 주석 연속 줄과 출력 메시지 문자열 리터럴 — 예: `harness.go:291`의 `Fprintln(..., "... skill calls AskUserQuestion ...")`). 본 변경이 추가한 AskUserQuestion 발생 0건: 3개 수정 파일 diff에서 토큰 2건은 모두 컨텍스트 줄(선행 공백, 기존 주석), 신규 2파일에서 0건.

### E5 린트

```
$ golangci-lint run --timeout=5m ./internal/cli/... ./internal/binlag/...
0 issues.   (변경 전 baseline도 0 issues — 신규 지적 0)
```

### E6 커밋

- `1c3adc4d5` feat(SPEC-AUDIT-BUILD-IDENTITY-001): audit verdicts carry the producing binary's build commit (t248)
- **푸시하지 않았다** — 통합은 리드의 develop 병합 창 소관. 브랜치 `WT-audit-binary-sha` 로컬에만 존재.

### §D.1 체크리스트 대조

- [x] AC 8건 PASS (E1)
- [x] `go test ./internal/cli/ ./internal/binlag/ -count=1` 통과 (304.4s / 4.1s, 커밋 전 예비 측정 + 커밋 후 AC 전수 재측정)
- [x] `go vet` 통과
- [x] golangci-lint 신규 지적 0
- [x] 새 MCP 도구 0 / 새 verdict 열거값 0 / 새 Go 패키지 0 / 중첩 신원 객체 0 / 버전 필드 0 (diff 상단 근거)
- [x] `PerBackendVerdict` 무변경 — `git diff mcp_convergence.go`에서 해당 타입 본문 +/- 0건 (주석 언급 1건은 ConvergenceResult 문서)
- [x] 폴백 cwd 미전달 — `git diff`에서 `performCodexAudit`/`performGLMAudit` 호출부 토큰 0건 변경; 폴백 디렉터리는 `auditBuildIdentity` 내부에서만 소비
- [x] 모든 판정 인용에 커밋 SHA `1c3adc4d5` 부착

### 실행 중 발견·수리 2건

1. **테스트가 실제 codex를 켠 사고 (수리 완료)**: `TestPersistedConvergenceCarriesBuildCommit` 최초판이 `backendCall`을 스텁하지 않아 `defaultBackendCaller → performCodexAudit → 실제 codex 세션`이 뜨고 fail-open 타임아웃(~60-77s)까지 대기했다. §D.2(트리 빌드 바이너리만) + CLAUDE.local.md §13(dev 프로젝트 라이브 백엔드 금지) 위반 — 타임아웃 고루틴 덤프로 원인 확정 후 `withBackendCallStub` 추가. 수리 후 0.00s.
2. **ast-grep 훅 룰 오탐 우회**: 로컬 룰 `go-error-ignored-blank`(`$_, $ERR = $FUNC(...)` 패턴)이 `:=`가 아닌 **plain `=`** 2값 대입 전부를 적중한다(공백 무관 — `res, err = f()` 도 차단). 테스트 코드를 `:=` 형태로 재구성해 우회했다. 룰 자체의 과다 매칭은 별도 소관(개선 필요 시 `/moai:feedback` 재료).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-01
run_commit_sha: "1c3adc4d5"
run_status: complete
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run (worktree-isolated card branch; no shared-checkout commit)
l44_post_push_fetch: not-run (NO PUSH — integration is the lead's window)
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin_arm64: pass
cross_platform_build.windows_amd64: pass
total_run_phase_files: 6
m1_to_mN_commit_strategy: single implementation commit (M1-M3 combined; Tier S, milestones implemented+verified as one unit)
```


## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-01
sync_commit_sha: "b60ca5583"   # backfilled — the sync commit cannot cite its own hash (canonical D3 exemption)
sync_status: complete
b12_self_test_a: pre_emission_grep_count_0 (grep -c 'SPEC-AUDIT-BUILD-IDENTITY-001' CHANGELOG.md → 0)
b12_self_test_b: ac_count_match (acceptance.md distinct ACs = 8 = AC-ABI-001..008; entry cites 8/8; no ambiguity → count emitted)
b12_self_test_c: file_paths_verified (spec.md link target, internal/cli/mcp_build_identity.go, internal/cli/mcp_build_identity_test.go — all ls-verified in this run)
changelog_entry_position: "[Unreleased] → Added, first entry (above SPEC-MEMORY-STORE-RECONCILE-001)"
frontmatter_status_transitions:
  spec_md: "in-progress → implemented → completed (merged into single sync commit; updated: 2026-09-01)"
  plan_md: none (no status field — ArtifactStatusFieldForbidden, card t357)
  acceptance_md: none (no status field)
  progress_md: none (no status field)
canary_compliance_check:
  mx_tag_pass: pass — no tag changes required; auditBuildIdentity carries @MX:NOTE (fail-open short-circuit) + @MX:SPEC; helper normalizeBuildCommit is unexported with dedicated tests; wiring files carry pre-existing @MX:SPEC anchors
  docs_scope_decision: no-change — README.md/README.ko.md/README.ja.md/README.zh.md grep for build_commit/buildCommit/audit verdict = 0 hits; .moai/docs/ = 0 hits; additive internal MCP JSON fields need no user-facing doc edit
  push: not-performed (integration is the lead's develop merge window)
```
