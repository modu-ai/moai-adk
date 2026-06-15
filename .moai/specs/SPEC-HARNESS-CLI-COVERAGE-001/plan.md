# SPEC-HARNESS-CLI-COVERAGE-001 — 구현 계획 (plan.md)

## §A — 컨텍스트

`internal/cli/harness` 패키지(`execute.go` / `install.go` / `propose.go`)의 statement coverage를
**77.9% → ≥90%**로 상향한다. test-only(프로덕션 non-test `.go` 무수정 원칙). 측정 baseline은
spec.md §B(ground-truth, 재도출 금지)를 SSOT로 삼는다.

cycle_type: **tdd**. 다만 본 SPEC은 "기능 추가"가 아니라 "기존 동작을 테스트로 고정"이므로, TDD의
RED 의미는 "신규 테스트가 미커버 statement를 도달시키는지 coverprofile로 확인 → 추가 → 커버리지
증가 검증"의 순서로 적용한다. 각 milestone 종료 시 `go test -coverprofile` 재측정으로 진척을 고정한다.

## §B — Known Issues / 사전 관찰

- **`NewExecuteCmd` os.Exit 분기 (execute.go:329-334)**: 표준 Go 테스트는 `os.Exit`를 가로지를 수
  없다. 이 분기는 §D(os.Exit 접근 결정)에서 처리한다.
- **`MarkFlagRequired` panic 분기 2개 (execute.go:344-345, install.go:168-169)**: flag 이름이
  존재하지 않을 때만 실패하는 방어적 도달-불가 코드. documented residual로 유지(테스트 왜곡 금지, EX-2).
- **`runExecuteWith` 90.0%**: 이미 목표선. 잔존 1 statement는 empty-root error 분기일 가능성 —
  M2에서 부수적으로 도달 가능하면 도달, 아니면 residual 후보. 강제하지 않는다.

## §C — Pre-flight (run-phase 진입 전 확인)

1. `go test -coverprofile=/tmp/hcc-base.out ./internal/cli/harness/...` → total 77.9% 재확인
   (baseline 고정; manager-develop가 RED 전 baseline을 캡처).
2. `go tool cover -func=/tmp/hcc-base.out | grep -vE '100.0%'` → spec.md §B.1 표와 일치 확인.
3. FROZEN-diff 기준 HEAD 기록(프로덕션 `.go` 무수정 검증용).

## §D — 핵심 결정: os.Exit 분기 접근 (NewExecuteCmd execute.go:329-334)

[HARD] **선택: (a) documented acceptable residual** — os.Exit 분기를 미커버 잔존으로 명시하고,
서브프로세스 re-exec 테스트는 도입하지 않는다.

**근거**:
- 로직은 이미 `runExecuteCommand`(테스트 대상, REQ-HCC-003/009/010)에 추출되어 있다. RunE 클로저는
  `runExecuteCommand` 호출 + error 시 `os.Exit(ExitCodeForError(err))`의 얇은 wrapper이며, exit
  코드 분류 로직(`ExitCodeForError`)은 `TestExecute_*ExitsUserError` 계열로 이미 완전 커버된다.
- 서브프로세스 re-exec 테스트(`BE_CRASHER` env + `exec.Command(os.Args[0], "-test.run=...")`)는
  cross-platform 취약성(Windows 셸 차이) + 테스트 wall-time 증가 + 본질적으로 "os.Exit가 호출됨"만
  확인하는 낮은 가치 대비 비용이 크다. Karpathy Simplicity First / EX-2와 충돌.
- ≥90% 목표는 os.Exit 분기 없이 도달 가능하다(`NewInstallCmd` 클로저 전체 + `runExecuteCommand`
  success/inherit 경로 + `runPropose` write/default 경로가 gap의 대부분을 닫는다).

**조건부 fallback**: 만약 run-phase 실측에서 os.Exit 분기 제외 후에도 `NewExecuteCmd`가 90% 미만이고
다른 도달 가능 statement가 없다면, 그때 한해 REQ-HCC-022 seam 예외(injectable exitFunc)를 발동하되
**plan-auditor 재검토 + 사용자 승인** 경로를 거친다. 기본 경로는 (a) residual이다.

## §E — Self-Verification (각 milestone 종료 시)

- `go test -coverprofile` 재측정 → 해당 milestone 목표 함수 커버리지 증가 확인
- `go test ./...` 전체 green (회귀 catch)
- `git diff --name-only internal/cli/harness/` → `*_test.go`만 변경(프로덕션 무수정, EX-1/REQ-HCC-021)

## §F — Milestones (6개 테스트 그룹 → M1..M6 매핑)

> 각 milestone은 spec.md §B.2 미커버 블록의 특정 부분집합을 도달시킨다. 우선순위는 "단일 gain
> 최대" 순(M1이 최대 gap `NewInstallCmd` 33.3%).

### M1 — `NewInstallCmd` RunE 클로저 (install.go:125-155) — 최대 gain

대상 블록: `125.52-155.14` (RunE 전체).

- `TestNewInstallCmd_Execute_Success`: `cmd.SetArgs([]string{"--spec-id", "X", "--domain", "Y",
  "--project-root", t.TempDir()})` + CLAUDE.md 픽스처(`writeHarnessFile`) → `cmd.Execute()` 호출.
  `.moai/harness/main.md` + CLAUDE.md marker 생성 + success stdout(`SetOut` 버퍼) emit 검증.
  (REQ-HCC-007)
- `TestNewInstallCmd_Execute_DefaultCwd`(**non-parallel**, `t.Setenv` 미사용): `os.Chdir(tmp)` 후
  `--project-root` 생략 → getwd 기본-cwd 분기(126-132) 도달. `t.Cleanup`으로 원래 cwd 복원.
  (REQ-HCC-008 / REQ-HCC-018)
- `TestNewInstallCmd_Execute_RunInstallError`: CLAUDE.md 부재 root로 `Execute()` → RunInstall 전파
  error 분기(149-150) 도달, exit-error 반환 확인.

→ 목표: `NewInstallCmd` ≥90% (REQ-HCC-002).

### M2 — `RunInstall` error 분기 (install.go:61-63, 74-76) — 100% 목표

대상 블록: `61.28-63.3`(empty root), `74.77-76.3`(ScaffoldHarnessDir 실패).

- `TestRunInstall_EmptyProjectRoot`: `RunInstall(InstallOptions{ProjectRoot: ""})` → empty-root
  error(61-63) 도달, `"empty project root"` 메시지 검증. (REQ-HCC-013)
- `TestRunInstall_ScaffoldFails_PreexistingFile`: `<root>/.moai/harness` 경로에 regular file을
  미리 생성하여 `ScaffoldHarnessDir`의 MkdirAll 충돌 유발 → 실패 분기(74-76) 도달, wrapped error
  `"scaffold .moai/harness"` 검증. (REQ-HCC-014)

→ 목표: `RunInstall` 100% (REQ-HCC-004). 기존 `TestRunInstall_MissingClaudeMd`가 InjectMarker
error 분기(85-88)를 이미 커버하므로 happy + 2 error = 전 분기 도달.

### M3 — `runExecuteCommand` success + InheritedFlags (execute.go:358-360, 366-369)

대상 블록: `358.65-360.4`(InheritedFlags lookup), `366.2-369.12`(success stdout emit).

- `TestRunExecuteCommand_Success_EmitsTelemetryNotice`: `writeE2EFixtureModule`(real Go 모듈) +
  `writeProposalFixture`로 valid proposal 구성 → `runExecuteCommand(cmd, id, root)` 성공 →
  success stdout(`apply-outcome telemetry recorded`) emit 블록(366-369) 도달. (REQ-HCC-009)
  - 주: 기존 `execute_e2e_test.go`가 `RunExecute` 성공 경로를 검증하나 `runExecuteCommand`의 stdout
    emit 블록은 미커버 — 이 테스트가 그 래퍼 라인을 도달시킨다.
- `TestRunExecuteCommand_InheritsProjectRootFromParent_Success`: 부모 cobra command에 persistent
  `--project-root`를 t.TempDir() fixture로 세팅 + 자식 execute에 own flag 없이 호출 → InheritedFlags
  lookup 분기(358-360)를 **success** 경로로 도달(기존 `TestRunExecuteCommand_EmptyProjectRoot_InheritsFromParent`는
  missing-proposal error 경로만 커버). (REQ-HCC-010)

→ 목표: `runExecuteCommand` ≥90% (REQ-HCC-003).

### M4 — `runPropose` non-dry-run write + default-flag fallback (propose.go:80-95, 125-137)

대상 블록: `80-95`(default fallback), `125.78-137.3`(WriteProposals 경로).

- `TestRunPropose_WriteMode_PersistsProposals`: actionable 패턴 1건을 담은 `tier-promotions.jsonl`
  픽스처(`code_change:func_extract:auth_module`, confidence 0.85)를 t.TempDir()에 작성 →
  `runPropose(cmd, OutputFlags{InputPath: fixture, OutputDir: outDir, DryRun: false})` →
  WriteProposals 경로(125-137) 도달, `<outDir>/<draft>/spec.md` + `proposal.json` 생성 검증.
  (REQ-HCC-011)
  - 주: 기존 `TestPropose_WriteMode_CreatesFiles`는 `cmd.Execute()` 경로. 본 테스트는 `runPropose`
    직접 호출로 default-flag 처리 라인을 함께 exercise.
- `TestRunPropose_DefaultFlags_UsesDefaults`: `runPropose(cmd, OutputFlags{})` (모든 override 생략)
  → default fallback 블록(80-95: InputPath/OutputDir/Limit 기본값 대입) 도달. 기본 InputPath
  (`proposalgen.DefaultInputPath`)는 cwd 상대로 부재 → no-op JSON(`absent or empty`) 반환,
  WriteProposals 미실행. (REQ-HCC-012)

→ 목표: `runPropose` ≥90% (REQ-HCC-005).

### M5 — `RunExecute` / `runExecuteWith` 잔존 error 분기 (execute.go:116/124/159)

대상 블록: `116.17-118.4`, `124.17-126.4`(RunExecute error), `159.17-161.4`(runExecuteWith error).

- 분석: `116/124`는 `os.Getwd()` 실패 / `filepath.Abs` 실패 분기로, 결정론적으로 유발하기 어렵다
  (정상 환경에서 둘 다 거의 실패하지 않음). 표준 도달 불가 가능성이 높음 → §D.4 residual 후보.
- `159`는 `runExecuteWith`의 empty-root `os.Getwd()` 실패 분기 — 동일하게 환경 의존.
- **접근**: M5는 "도달 가능한 것만 도달"한다. `runExecuteWith` empty-root **성공** fallback은 기존
  `TestExecute_RunExecuteWith_EmptyRoot`가 이미 커버. getwd/Abs **실패** 분기는 결정론적 유발이
  test-only로 불가하면 §D.4 residual로 명시하고 강제하지 않는다(EX-2/EX-4). (REQ-HCC-015)
- 만약 total이 이미 ≥90%를 달성했다면 M5는 잔존 분기를 residual로 문서화만 하고 종료.

→ 목표: total ≥90% 달성에 기여(잔존은 residual 명시).

### M6 — 통합 검증 + residual 문서화 + traceability

- `go test -coverprofile` 최종 측정 → total ≥90.0% 확인(REQ-HCC-001) + 함수별 목표 충족 확인.
- documented residual 최종 확정(os.Exit 분기 + 2 panic 분기 + 도달 불가 시 getwd/Abs 실패 분기)을
  acceptance.md AC-HCC-009 / progress.md에 열거.
- `go test ./...` 전체 green / `go vet ./...` clean / `golangci-lint run` baseline (REQ-HCC-020).
- `git diff --name-only` → 패키지 하위 `*_test.go`만(프로덕션 무수정, REQ-HCC-021).
- `execute_e2e_test.go` green(REQ-HCC-019) + `TestPropose_NoAskUserQuestion` green(REQ-HCC-017).
- AC ↔ REQ ↔ milestone traceability 표 최종 정합.

## §G — Anti-Patterns (회피)

- **AP-1 — os.Exit 분기 강제 커버**: 서브프로세스 re-exec로 os.Exit 분기를 억지로 커버 → §D 결정 (a) 위반.
- **AP-2 — getwd/Abs 실패 분기 강제**: 환경 의존 분기를 도달시키려 chdir 후 권한 조작 등 fragile
  hack 도입 → cross-platform 취약 + EX-2 위반.
- **AP-3 — 프로덕션 로직 변경으로 커버리지 확보**: seam 도입을 기본 경로로 사용 → EX-1/REQ-HCC-022
  조건부 예외만 허용(plan-auditor 재검토 필수).
- **AP-4 — t.Parallel + os.Chdir 혼용**: cwd는 프로세스 전역 → parallel chdir은 데이터 레이스.
  REQ-HCC-018 위반.
- **AP-5 — 100% 추구**: residual을 0으로 만들려 테스트 왜곡 → EX-4 위반.

## §H — Cross-References

- spec.md §B (ground-truth baseline, SSOT) / §C (REQ-HCC-*) / §E (Exclusions)
- acceptance.md (AC-HCC-* 측정 기준)
- `internal/cli/CLAUDE.md`(절대경로 / exit code / 경계) / `CLAUDE.local.md` §6(t.TempDir)
- `.claude/rules/moai/languages/go.md`(table-driven / t.Parallel / coverage)

## §I — Tier 정당화

**Tier S (minimal)**. 단일 패키지(`internal/cli/harness`), test-only 추가, 결정론적, 저위험.
프로덕션 동작 변경 없음(seam 예외는 조건부·미발동 기본). milestone을 6개로 나눈 것은 SSE-stall이나
복잡도 때문이 아니라 **6개 독립 테스트 그룹의 논리적 단위**를 명확히 하기 위함이며, 각 그룹은 동일
세션에서 순차 처리 가능하다. Tier M로 격상할 필요 없음 — 파일 1개(신규 `*_test.go`) 추가가 주된
변경이고, 인접 파일 의존/cascade 위험이 없다.
