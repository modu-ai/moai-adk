# SPEC-COMMIT-IDENTITY-GUARD-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- card t1289, Tier M. plan-phase 산출물 4 개(`spec.md`, `plan.md`, `acceptance.md`, `progress.md`),
  워크트리 `.claude/worktrees/t1289`, 브랜치 `WT-commit-identity-guard`, base `37dc766b9`.
- SPEC ID 사전 점검: Bash 정규식 검사 출력 `PASS`.
- 입력 근거: `.moai/reports/t1289/root-cause.md`. plan 페이즈의 새 측정은 픽스처 이메일 열거
  (51 개, `plan.md` §B.1)와 센티널 스윕(14 종)뿐이다.
- 요구사항 REQ-CIG-001..012, 인수조건 AC-CIG-001..014.
- `plan.md` §B.3 저장소 범위 판단: 레인 결정 (b) 로 해소(2026-09-28, spec 0.1.1). 미해소 `[NEEDS CLARIFICATION]` 0 건.
- status: `draft`.
- plan_status: audit-ready

## §E.2 Run-phase Evidence

측정 트리: 워크트리 `.claude/worktrees/t1289`, 브랜치 `WT-commit-identity-guard`.
아래 증거는 전임 레인 WIP 채택 + 수리가 끝난 디스크 상태에서 측정됐고, M1-M3 커밋
(`30947f28f`, `d5598bc4e`, `4dc3d06a1`)은 그 디스크 내용을 그대로 기록한 것이라
트리 내용 기준으로 동일하다. 최종 확정 재측정은 M4 이후 final HEAD (`git log -1`) 로
§E.3 완료 시점에 병기한다. (측정일 2026-09-28/29, 이 레인 실행.)

### E.2.1 — 픽스처 이메일 열거 재측정 (plan.md §B.1, run base)

명령: plan.md §B.1 의 열거 명령을 run base 에서 단일 호출로 재실행 →
`.moai/reports/t1289/fixture-emails-runbase.txt` 저장.

```
$ wc -l < .moai/reports/t1289/fixture-emails-runbase.txt
51
```

내장 목록 대조 (양방향 comm, 양쪽 모두 0):

```
missing_from_builtin_list_count=0
builtin_entries_not_in_enumeration: 0
```

→ 내장 목록 ⊇ 열거 결과, 51 = 51 로 일치. 목록 수정 불필요.

### E.2.2 — E1: 인수조건 테스트 스코프 실행 (AC-CIG-001..014 판정)

명령: `go test ./internal/hook -run 'TestCommitIdentityGuard_|TestAC_CIG_' -count=1 -v`

```
exit=0
--- PASS count (top-level): 19
(no --- FAIL lines)
ok  	github.com/modu-ai/moai-adk/internal/hook
```

AC ↔ 테스트 매핑 (전부 위 실행의 `--- PASS`):

| AC | 판정 테스트 | 상태 |
|----|------------|------|
| AC-CIG-001 | TestAC_CIG_001_FixtureIdentityCommitDenied | PASS |
| AC-CIG-002 | TestAC_CIG_002_AllEightTriggerVerbsDenied | PASS |
| AC-CIG-003 | TestAC_CIG_003_NonTriggerRunsNoProbes | PASS |
| AC-CIG-004 | TestAC_CIG_004_CommandLevelOverrides | PASS |
| AC-CIG-005 | TestAC_CIG_005_RealProbeRealIdentityAllowed | PASS |
| AC-CIG-006 | TestAC_CIG_006_ExactMatchOnly | PASS |
| AC-CIG-007 | TestAC_CIG_007_ResolutionFailureAllowsWithAudit + TestAC_CIG_007b_ProbeTimeoutRealPath | PASS |
| AC-CIG-008 | TestAC_CIG_008_DisabledGuardRunsNoProbes + TestCommitIdentityGuard_EnabledFlipsDecision + internal/config TestCommitIdentityGuard_DefaultOff (`go test ./internal/config -count=1` → `ok ... coverage: 82.8%`) | PASS |
| AC-CIG-009 | TestAC_CIG_009_ConfigListAddsToBuiltin + TestAC_CIG_009_WorkflowYamlCarriesKey | PASS |
| AC-CIG-010 | TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration | PASS |
| AC-CIG-011 | 변이 증거 — `.moai/reports/t1289/mutation-evidence.md` (E.2.6) | PASS |
| AC-CIG-012 | TestAC_CIG_012_WiringPreservesEarlierDeny + TestAC_CIG_012_PowerShellUnclassified + TestAC_CIG_012_SentinelDefinedOnce | PASS |
| AC-CIG-013 | TestAC_CIG_013_OtherRepositoryAllowed | PASS |
| AC-CIG-014 | TestAC_CIG_014_LinkedWorktreeDenied | PASS |

AC-CIG-012 센티널 단일 정의 증거:

```
$ grep -rn '"TEST_IDENTITY_VIOLATION' internal/hook --include='*.go' | grep -v _test
internal/hook/commit_identity_guard.go:41:const testIdentityViolationPrefix = "TEST_IDENTITY_VIOLATION:"
```

(비테스트 파일 정확히 1 건. 센티널 패밀리 스윕 `grep -rhoE '"[A-Z_]+_VIOLATION"' … | sort -u | wc -l` → 15 — plan 시점 14 + TEST_IDENTITY 1.)

### E.2.3 — E1(확장): `internal/hook` 전체 스위트 (slot lease 하 3회: 2회 실패 원인 규명 + 1회 최종)

명령: `moai slot acquire --resource hook-suite --max-duration 30m` →
`go test -timeout 30m ./internal/hook/...` → `moai slot release --resource hook-suite`

최종 실행 (마지막):

```
E1-final-exit=0
ok  	github.com/modu-ai/moai-adk/internal/hook	474.963s
ok  	github.com/modu-ai/moai-adk/internal/hook/handoff	(cached)
ok  	github.com/modu-ai/moai-adk/internal/hook/memo	(cached)
… (전 11 패키지 ok)
```

중간 2회 실패는 **카드 귀속 결함이었다가 수리된 케이스**: 1차는 전임 WIP `TestAC_CIG_012_…/push_readiness_wins` 가
contract 모드 풀 Handle 경로에서 escalation 관측을 **공유 패키지 테스트 홈**(`moai-hook-test-home-*`)에 기록했고,
기존 `TestEscalationGuidedGolden` 이 홈 전체를 걷는 `assertNoEscalationDir(t, os.Getenv(HOME))` 에서 이를 적중했다.
귀속 판별 실험: golden + 본 카드 스코프 테스트만 실행 → FAIL 재현(4건), golden + 기존 `TestClosurePush` 만 실행 → ok.
수리: 해당 서브테스트에 `t.Setenv("MOAI_HOME", t.TempDir())` 로 MoaiHome 을 테스트 자체 임시디렉터로 격리 →
golden + 배선 테스트 조합 ok, 최종 풀 스위트 ok. (HOME 교체 금지와 무관 — MOAI_HOME 은 paths.MoaiHome() 이
전면 대체하는 별도 변수이고 해당 테스트는 비병렬이다.)

`internal/gitenv` (plan §E E2): `go test ./internal/gitenv/... -count=1` → `ok ... 1.145s`.

### E.2.4 — E2: 빌드 (E6: make build 포함)

```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./internal/hook ./internal/config → exit 0
$ make build                              → exit 0
```

### E.2.5 — E3: 커버리지

```
$ go test -cover ./internal/config
ok  	github.com/modu-ai/moai-adk/internal/config	3.016s	coverage: 82.8% of statements
```

가드 파일 스코프 커버리지 (스코프 실행 `-coverprofile`, 문장 가중):

```
guard-file coverage: 311/358 = 86.9%   (internal/hook/commit_identity_guard.go)
```

초기 스코프 측정 75.7% → 헬퍼 단위 테스트(`TestCommitIdentityGuard_HelperUnits`:
shellFields / splitShellSegments / gitGlobalOptionSpan / emailFromAuthorValue /
resolveAgainst / sameCommonDir / gitVarIdent 실패경로 / applySegmentFacts attached-form) 추가로 85% 게이트 충족.
훅 패키지 전체 기준치: 패키지 스위트 대상 총 문장 대비 스코프 집계라 의미가 낮아
파일 단위 수치를 정본으로 쓴다.

### E.2.6 — E7: 변이 증거 (AC-CIG-011)

증거 파일: `.moai/reports/t1289/mutation-evidence.md`. 12 변이(비교 항상-불일치, 8 동사 각각 제거,
emailFromIdent 파손, sameCommonDir → true/false) 각각 exit 1 + 이름 붙은 테스트 FAIL.
(d1) → `TestAC_CIG_013_OtherRepositoryAllowed`, (d2) → `TestAC_CIG_014_LinkedWorktreeDenied` 가 잡음.
변이는 커밋하지 않았고 각 실행 후 원본 복원을 `cmp` 로 검증(`TREE_RESTORED_IDENTICAL`).

### E.2.7 — E4: 서브에이전트 경계 (dispatch B3 서술 정정)

dispatch 가 기대한 `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook --include='*.go' | grep -v _test.go` →
0 매치는 이 트리에서 (base `bdcb90b34` 에서도) 성립하지 않는다 — 실측 22 매치, 전부 주석 줄이거나
`askuser_observer.go`·`pre_tool.go:821` 의 **AskUserQuestion 이벤트 관측자**(도구 호출이 아니라 도구 이름 문자열 비교)다.
규율의 실제 불변식은 "훅이 AskUserQuestion 을 호출하지 않는다"이고, WIP 가 추가한 신규 매치는
`git diff internal/hook/pre_tool.go | grep -c AskUserQuestion` → **0** 으로 0 건이다.
기계 가드 `internal/hook/subagent_boundary_test.go` 가 풀 스위트에서 통과한다 (E.2.3).

### E.2.8 — E5: 린트 (v2.1.6, CI 판 확인: `golangci-lint version` → `v2.1.6`)

```
$ golangci-lint run ./internal/hook/... ./internal/config/...
0 issues.
```

과정에서 전임 WIP 의 신규 이슈 1 건(SA9003 빈 분기, commit_identity_guard_test.go)을 발견·수리했다 —
최종 상태 신규 이슈 0. 기존 baseline 이슈 없음.

### E.2.9 — E8: RED 증거

세션 내 관측 가능한 유일한 RED 는 인계 시점에 리드가 측정하고 본 레인이 재현한
AC-CIG-007 픽스처 버그다 — 전임 WIP 는 본 세션 이전에 test-first 로 작성됐다.

```
$ go test ./internal/hook -run 'TestAC_CIG_007_ResolutionFailureAllowsWithAudit' -count=1
--- FAIL: TestAC_CIG_007_ResolutionFailureAllowsWithAudit (0.05s)
    --- FAIL: TestAC_CIG_007_ResolutionFailureAllowsWithAudit/project_scope_failure (0.00s)
        commit_identity_guard_test.go:488: audit lines = 6, want exactly 1
FAIL
```

(실행마다 1줄씩 누적 — 고정 상대경로 projectDir `ci-project-side` 의 결함 자체가 증거. 원문 저장:
`.moai/reports/t1289/red-evidence-ac-cig-007.txt`.) 수리 후 동일 스코프 실행은 E.2.2 의 초록이다.
그 밖의 세션 내 RED 는 없으며, acceptance.md 채택 셀의 plan-time 부재 증거(측정 트리 `37dc766b9`)
가 나머지 AC 의 RED-now 셀을 대신한다.

### E.2.10 — WIP 수리 내역 (채택된 전임 WIP 에 대한 본 레인의 변경)

1. AC-CIG-007 `project_scope_failure` 픽스처: 고정 상대경로 `ciProjectSideMarker` → `t.TempDir()` 격리
   (누적 감사 로그 결함 수리). `target_scope_failure` 클로저도 무조건 실패형으로 단순화, 미사용 상수 제거.
2. AC-CIG-005 양 서브테스트: `input.CWD = repo` 누락 수리 — Given(임시 저장소 = 프로젝트 디렉터이자 명령 cwd)을
   지키지 않아 범위 단계에서 조용히 allow 됐고 `fixture_pair_denies` 는 거부 경로를 전혀 검증하지 못했다
   (subtest 1 의 counts 단언은 직접 probe 호출이 메꿔주는 공허 초록이었다).
3. `push_readiness_wins` MoaiHome 격리(E.2.3 참조).
4. SA9003 빈 분기 제거, gofmt 정렬.
5. `internal/hook/zz_debug_test.go` 파기, 잔해 `internal/hook/ci-project-side/` 삭제(리드 지시).
6. `internal/config/testdata/shipped_key_inventory.yaml` 에 신규 키 2건 등록(class W, reader) —
   REQ-CKH-008 anti-rot 게이트 통과. 총 엔트리 977 → 979.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-29
run_status: complete
run_commit_sha: 3474bcea8   # M4 커밋 — D3 백필 창 완료(후속 커밋에서 기입); §E.2 서두의 M1-M3 SHA와 git log로 검증 가능
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-performed    # 레인은 push 하지 않는다 — fetch 기반 race 점검은 리드 통합 창의 몫
l44_post_push_fetch: not-applicable    # 상동 — 레인 push 없음
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin_amd64: pass
cross_platform_build.windows_amd64: pass
total_run_phase_files: 12
m1_to_mN_commit_strategy: milestone-shaped M1-M4 (M1 config+transition 30947f28f, M2 guard core d5598bc4e, M3 wiring 4dc3d06a1, M4 evidence+progress)
```


## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-29
sync_status: complete
sync_commit_sha: 8d0cc46f1   # sync commit — D3 backfill window completed (follow-up commit); verifiable via git log
b12_self_test_a: pass   # grep -c 'SPEC-COMMIT-IDENTITY-GUARD-001' CHANGELOG.md → 0 (pre-emission; emission-safe)
b12_self_test_b: pass   # AC count 14 (AC-CIG-001..014, all live — no [RETIRED]/[REF] markers) matches CHANGELOG entry
b12_self_test_c: pass   # all 12 cited implementation file paths verified via ls before drafting
changelog_entry_position: "[Unreleased] → ### Added (first bullet)"
frontmatter_status_transitions.spec: in-progress → completed (merged 3-phase close, single sync commit)
frontmatter_status_transitions.updated: 2026-09-29 (unchanged — already current)
mx_tag_changes.added: 1   # @MX:WARN on checkCommitIdentity (complexity ≥ 15), [AUTO] + @MX:REASON + @MX:SPEC
mx_tag_changes.updated: 0
mx_tag_changes.removed: 0
mx_tag_anchor_check: checkCommitIdentity fan_in = 1 (pre_tool.go:689 sole caller, measured by grep over internal/ non-test) — no ANCHOR; no exported symbols; no goroutines
readme_docs_site_assessment: no-edit — sibling guards (branch_guard, integration_lock) have no README/docs-site sections; no doc surface enumerating guard keys found (workflow.yaml comments are the only config surface)
sanity_build: go build ./internal/hook ./internal/config → exit 0 (post-MX-edit)
```

## 진행 기록

### 2026-09-28 — plan-audit iter2 D14/D15 수리 (감사 한도 2/2 소진, 리드 채널로 검증)

- **D14(blocking, MP-2 GEARS)** — REQ-CIG-001·008 이 상시 의무와 조건부 의무를 한 항목에 묶은 것을 **분할, 압축 금지** 방향으로 고쳤다. REQ-CIG-001 → 분류·비트리거 통과(Ubiquitous, REQ-CIG-001) + 트리거 동사 탐침(Event-driven, REQ-CIG-002)으로 분리. REQ-CIG-008 → 배선 순서·선행 거부 보존(Ubiquitous, REQ-CIG-009) + 분류 불가 PowerShell 간접 구문 통과·감사 1줄(Event-driven, REQ-CIG-010)으로 분리. 나머지는 001..012 연속 번호로 재배치했고(구 002→003 … 구 009→011, 구 010→012), AC 수는 14 개 그대로다 — 분할 REQ 모두 기존 AC 가 덮어서 신규 AC 는 불필요했고, AC-CIG-001..014 의 Covers 만 새 번호로 재귀속했다.
- **구 REQ-CIG-010 분할 여부**: 세미콜론 뒤 별도 의무가 없다 — "While <범위 불일치> … shall allow … and without an identity probe" 는 단일 State-driven 응답 안의 한 동작이라 판정했고, 분할 없이 REQ-CIG-012 로 번호만 옮겼다.
- **D15(optional)** — AC-CIG-012(c) 의 Given 은 선행된 수리 커밋(`5e8599426`)에서 이미 구체화돼 있었다(autonomy `contract` 모드 + `checkClosurePush` 가 `second_review_not_performed` 로 거부하는 push-readiness fixture). 근거 재확인: 모드 게이트는 `internal/hook/closure_push.go` L83(`s.Mode != "contract"` → 빈 결정), 거부 코드 `second_review_not_performed` 는 `internal/closure/readiness.go:22`. 이번 커밋에서 acceptance.md 는 Covers 재귀속만 했다.
- **교차 참조 갱신**: spec.md §6 제약(구 REQ-CIG-006→REQ-CIG-007), plan.md §B.1(007→008)·§B.3(010→012)·§F M2(006→007)·§H(001..010→001..012), progress.md §E.1 범위 표기.
- **검증 한계**: 정식 plan-auditor 재감사는 하지 않는다 — 감사 한도 2/2 소진. 레인이 `moai spec lint` + REQ↔AC 매핑 재검증 + 연속 번호 검사를 이 트리에서 실행했고, MP-2 해소 판정은 리드 채널이 읽는다.
