# SPEC-CODEX-LAUNCH-VERB-001 — 진행 기록

카드: t391 · 워크트리 `.claude/worktrees/t391` · 브랜치 `WT-codex-launch-verb`

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (3-artifact set: spec.md + plan.md + acceptance.md). 파일·LOC 축은 S 로 읽히나 REQ 12건이 Tier S 상한 8을 넘고, 완결 SPEC 의 요구 하나를 대체하는 승계 검증 부담이 별도 검증 층을 요구한다 — 높은 쪽을 택했다.
- REQ: REQ-CLV-001..012 (12건 / 상한 16)
- AC: AC-CLV-001..015 (15건 / 상한 16), REQ 전수 커버
- 승계: SPEC-CODEX-LAUNCHER-001 REQ-CL-002 를 대체. `depends_on: [SPEC-CODEX-LAUNCHER-001]`
- 게이트: 미해소 명확화 마커 **0건 — 2026-08-31 운영자 판정으로 2건 모두 해소**
  - §B.3 `-w` → (가) strip-and-set-Dir. resolve 이며 create 아님. 비대칭과 그 이유를 REQ-CLV-007/008 에 [HARD] 로 고정
  - §B.4 argv → 합성 동사는 어느 경로로도 자식에 닿지 않음. 라우팅 하류에 별도 번역 표(REQ-CLV-004 정규 5행)
- 이 트리에서 빌드·테스트 0건 (`plan.md` P2)
- 기계 검증 범위: `spec.md` 만 `moai spec lint --strict` 통과. `plan.md`·`acceptance.md` 는 린터 판정 없음 (`plan.md` P6)

## §E.2 Run-phase Evidence

> 측정 트리: 워크트리 `.claude/worktrees/t391`, 브랜치 `WT-codex-launch-verb`, 진입 HEAD `e445e3276`(= `origin/develop` 대비 `0 4`). M1 커밋 `e33eeb93c`.
> 아래 모든 셀은 **이 트리에서 실제로 돌린 명령의 출력**이며, `plan.md` P2("이 트리에서 빌드·테스트 0건")를 해소한다. 증거 파일은 `.moai/reports/t391/run/` 아래에 있다.

### 검증 명령과 종료 코드

| # | 명령 | exit | 증거 |
|---|---|---|---|
| V1 | `go build ./cmd/moai` | **0** | `.moai/reports/t391/run/v1-build.txt` (0바이트 — 출력 없음) |
| V2 | `go test ./internal/cli/...` | **0** | `.moai/reports/t391/run/v2-test.txt` — `ok github.com/modu-ai/moai-adk/internal/cli 306.710s` 외 16개 하위 패키지 전부 `ok` |
| V3 | `GOOS=windows go vet ./internal/cli/...` | **0** | `.moai/reports/t391/run/v3-vet-windows.txt` (0바이트) |
| V4 | `go vet ./internal/cli/...` | **0** | `.moai/reports/t391/run/v4-vet.txt` (0바이트) |
| V5 | `golangci-lint run ./internal/cli/...` | **0** | `.moai/reports/t391/run/v5-lint.txt` — `0 issues.` (`plan.md` §E E5) |

RED 근거: `.moai/reports/t391/run/red-01-new-tests.txt` — 구현 **전** 새 시험 30셀이 FAIL(exit 1). GREEN 근거: `.moai/reports/t391/run/green-01-new-tests.txt`(exit 0), 전 패키지 `.moai/reports/t391/run/green-02-cli-pkg.txt`(exit 0). AC별 상세 PASS 목록: `.moai/reports/t391/run/v2b-ac-evidence.txt` — 상위 41건 · 하위 포함 120건 PASS, FAIL 0.

### AC 판정표

| AC | 판정 | Actual Output (판정 근거) |
|---|---|---|
| AC-CLV-001 | **PASS** | `TestCodexLaunchVerb_BareInvocationLaunches` PASS. 기동 seam 1회(direct 1 / spawn 0), stdout·stderr 모두 0바이트 |
| AC-CLV-002 | **PASS** | `TestCodexLaunchVerb_StatusStaysTheReadout` PASS. 리드아웃 6행, 기동 0, rc 0 |
| AC-CLV-003 | **PASS** | `TestCodexLaunchVerb_BareAndCliBuildTheSameRequest` PASS. (Program, Argv, Dir) 세 값 **값 비교** 일치 |
| AC-CLV-004 | **PASS** | `TestCodexLaunchVerb_RoutingSetsAfterReversal` PASS. 심볼 도출 집합 = 기동 {"", cli, app} · 리드아웃 {status}. 미등록 토큰 6종(`launch` `run` `bogus` `cl` `CLI` `start`) 전부 rc 1 + `codexUsageDiag` 정확 일치 + 기동 0 |
| AC-CLV-005 | **PASS** | (1) `TestCodexLaunchVerb_SynthesizedVerbNeverReachesChild` PASS — 맨몸·`cli` 두 경로 모두 자식 argv tail 길이 0, `cli`·빈 문자열 부재. (2) 기대값 반전 근거가 커밋 `e33eeb93c` 본문에 있다 — 왜 바뀌는지 + `codex --help` Commands 목록 축자 인용(`cli` 없음, `app` 있음). 원본 출력 `.moai/reports/t391/run/codex-help.txt` |
| AC-CLV-006 | **PASS** | `TestCodexLaunchVerb_AppTokenStillForwarded` PASS + 기존 `TestCodexVerbRouting_AppArgvExact` PASS. 자식 argv tail = `["app"]` — 과잉 적용 없음 |
| AC-CLV-015 | **PASS** | `TestCodexLaunchVerb_TailPassesThroughVerbatim` PASS(맨몸/`cli` 2셀) — 두 경우 모두 `["--model","o3"]`. 기존 `TestCodexVerbRouting_PassthroughTailExact` PASS(공백·`$`·`=` 포함 5토큰) |
| AC-CLV-007 | **PASS** | `TestCodexLaunchVerb_CodexHomeExplicitToChild` PASS 2셀. env 제공 시 자식 Env 마지막 `CODEX_HOME`이 그 값과 일치; **부모에서 `os.Unsetenv` 한 상태**에서도 자식 Env에 `resolveCodexHomeDir()` 해석값이 존재 — 주변 상속 구현은 이 셀에서 떨어진다 |
| AC-CLV-008 | **PASS** | `TestCodexLaunchVerb_ParentEnvPreserved` PASS(표식 변수 보존) + `TestCodexLaunchVerb_SpawnCarriesTheSameCodexHome` PASS(spawn 명령 문자열 접두사가 direct 경로의 값과 동일) |
| AC-CLV-009 | **PASS** | 기존 `TestCodexLauncher_NoWriteSnapshot` PASS. 격리 홈 아래 맨몸·`status`·`cli`·`app` × (존재/부재 CODEX_HOME) 8셀, 실행 전후 트리 스냅숏 `reflect.DeepEqual` 일치. 부재 CODEX_HOME은 실행 후에도 `IsNotExist` |
| AC-CLV-010 | **PASS** | `TestCodexLaunchVerb_WorktreeSetsDirAndIsNotForwarded` PASS 5셀(`-w NAME` / `-w=NAME` / `--worktree NAME` / `--worktree=NAME` / 동사 뒤 / tail 동반). 자식 Dir = 워크트리 루트, argv 에 `-w`·`--worktree`·이름 토큰 모두 부재 |
| AC-CLV-011 | **PASS** | `TestCodexLaunchVerb_WorktreeResolvesNeverCreates` PASS 3셀. (i) 접두사 밖 절대경로 → 기동 0 + 진단이 `resolveWorktreeL2Path` 출력과 **바이트 일치**(cc 와 동일 진단). (ii) 부재 이름 → 기동 0 + 실행 후 그 경로 여전히 `IsNotExist` (**resolve 이며 create 아님**). (iii) 값 없는 `-w` → 기동 0 |
| AC-CLV-012 | **PASS** | `TestCodexLaunchVerb_GateInheritedByBareForm` PASS 5셀. 미배선 4상태(`absent`/`empty dir`/`hooks only`/`config only`) × 비대화형 → 프롬프트 호출 0, 기동 0, rc 비영, stderr 에 `non-interactive` + `codexWiringAction`. `wired` 상태는 막지 않고 기동 1 |
| AC-CLV-013 | **PASS** | `TestCodexLaunchPath_CrossPlatformPropertyHolds` PASS 2셀 — 실파일 findings 0 **AND** 심은 위반 4종(OS 빌드태그 / `"syscall"` import / `syscall.Exec` / `golang.org/x/sys/unix`)이 전부 잡힘(공허 아님). `GOOS=windows go vet ./internal/cli/...` exit 0(V3). 종료코드 전파 `TestCodexVerbRouting_ExitCodePropagation` 5셀(0,1,2,126,127) PASS |
| AC-CLV-014 | **PASS** | `TestLaunchers_BareInvocationConvention` PASS — `cc`·`glm`·`codex` 세 값 모두 참, **기존 교차 시험 파일 `codex_launcher_test.go` 확장**(새 교차 파일 없음). help 문안: `TestCodexCommand_HelpDescribesReversedDefault` PASS + 기존 `TestCodexCommand_HelpCopyGuidance` PASS + 중립성 `TestCodexCommand_NeutralityScan` PASS(SPEC-ID·REQ-ID·카드id·내부 날짜·SHA·비ASCII 0건) |

**전수 판정: 15/15 PASS · FAIL 0.**

`spec.md` REQ-CLV-004 정규 표 다섯 행 대응: 맨몸→AC-CLV-001/005 · `cli`→AC-CLV-003/005 · `status`→AC-CLV-002 · `app`→AC-CLV-006 · `-- <args>`→AC-CLV-015. 전부 PASS.

### 물려받은 기대값을 바꾼 자리 (전수)

커밋 `e33eeb93c` 본문에 7건 전부 이유가 적혀 있다. 요지:

| # | 위치 | 바뀐 이유 |
|---|---|---|
| 1 | `TestCodexVerbRouting_PassthroughTailExact` (구 :465) | `cli` 전달은 계약이 아니라 결함을 고정한 것 — codex 에 `cli` 서브커맨드 없음(help 인용) |
| 2 | `TestCodexSpawn_RealAssemblyThroughStubTmux` | 같은 이유 + `CODEX_HOME` 대입 추가(tmux 창은 tmux 서버 환경을 상속) |
| 3 | `TestCodexVerbRouting_LaunchCountsPerVerb` | 맨몸 행 0 → 1 |
| 4 | `TestCodexSpawn_RejectedOnReadoutForms` | 맨몸 행 제거 — 이제 기동 형태라 `--spawn` 이 정상 |
| 5 | `TestCodexVerbRouting_HelpAfterDashDashIsNotHelp` | 성질은 불변(`--help` after `--` 는 codex 것), 판정 형태만 "거절"→"자식에 도달"로 더 직접적으로 |
| 6 | `TestCodexVerbRouting_ClosedSets` | 삭제 — 집합 도출이 새 파일 한 자리로 이동(리터럴 집합 2벌 = 동시 갱신 부담) |
| 7 | 리드아웃/교차 파일 bare 셀 4곳 | 맨몸이 리드아웃 형태가 아니게 됨 → `status` 로 이동, "두 형태 동일" 셀 삭제 |

### 승계하는 Gap (닫지 않음 — 산문으로 해소하지 않는다)

| Gap | 상태 |
|---|---|
| 설치본 `~/go/bin/moai --help` 문안 | **미재측정** (`plan.md` P1). 이 세션은 트리 소스만 잰다 |
| `codex app` 실제 기동 거동 | **미관측** (`plan.md` P3). `app` 토큰이 자식 argv 에 실린 것까지만 관측했고 codex 자식을 실제로 띄우지 않았다 |
| codex 서브커맨드별·숨은 플래그 | **최상위 help 만 훑음** (`plan.md` P4). "`cli` 없음"과 "`-w` 없음" 두 부재 주장은 그만큼 좁다 |
| 대화형 tty 왕복 | **구조적 미관측** (0.8.0 승계, `plan.md` P5) |
| `plan.md`·`acceptance.md` 기계 검증 | **없음** — `moai spec lint` 는 `spec.md` 만 판정(`plan.md` P6). 이 문서의 수치는 세어 본 것이지 린터 판정이 아니다 |
| `-w` 절대경로 **수용** 분기 | 시험에서 직접 타지 않음. `resolveWorktreeL2Path` 가 실제 `findProjectRoot()` 기준 접두사로 판정하므로 `t.TempDir()` 프로젝트 루트 아래 절대경로는 구조상 접두사 밖이다. **거절** 분기는 판정했고 진단이 cc 와 바이트 일치함을 보였다 |
| 범위 밖 관측 1건 | `internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md` 76-80행이 `moai codex task` / `setup` / `job status` 를 "CLI equivalent" 로 적고 있으나 그런 CLI 동사는 존재하지 않는다(`codexVerbRouting` 은 이 카드 전에도 {"", status, cli, app} 닫힌 집합이었다). **이 카드가 만든 것도 바꾼 것도 아니어서 손대지 않았다** |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-01
run_commit_sha: e33eeb93c        # M1 (구현 + 시험 + spec.md status). M2 = 이 문서
run_status: complete
ac_pass_count: 15
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run    # 워크트리 격리 세션, 푸시 없음 — primary 체크아웃 미접촉
l44_post_push_fetch: not-run     # 이 카드는 푸시·PR·병합을 수행하지 않는다
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues, go vet exit 0
cross_platform_build:
  goos_windows_vet: pass         # GOOS=windows go vet ./internal/cli/... exit 0
  os_build_tags: 0
  syscall_imports: 0
  scan_mutation_controlled: true # 심은 위반 4종 전부 적발
total_run_phase_files: 6         # 구현 1 + 시험 4 + spec.md 1 (증거 파일 별도)
m1_to_mN_commit_strategy: "M1 = 구현+시험+status 전이 (e33eeb93c), M2 = progress.md §E + 증거 파일"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-01
sync_commit_sha: 2f28bc394   # 커밋이 자기 해시를 인용할 수 없어 바로 다음 커밋에서 backfill
sync_status: complete
b12_self_test_a: pass    # grep -c 'SPEC-CODEX-LAUNCH-VERB-001' CHANGELOG.md → 0 (중복 없음, 방출 가능)
b12_self_test_b: pass    # acceptance.md 의 AC 식별자 15건 = CHANGELOG 가 적은 15건
b12_self_test_c: pass    # CHANGELOG 가 이름을 댄 경로 전부 ls 로 존재 확인
changelog_entry_position: "[Unreleased] → ### Changed, 절 첫 항목"
frontmatter_status_transitions:
  spec_md: "in-progress → implemented → completed (이 sync 커밋 1건이 운반)"
  plan_md: "n/a — 상태 필드 없음(artifact statelessness). 이 문서 축은 spec.md 가 단독 보유"
  acceptance_md: "n/a — 같은 이유"
  progress_md: "n/a — 상태는 본문 절(§E.*)로 기록, frontmatter 없음"
canary_compliance_check:
  applicable: false      # 이 SPEC 은 자기 sync 가 시험할 전방 정책을 정의하지 않는다
supersession_pointer:
  target: SPEC-CODEX-LAUNCHER-001
  form: "frontmatter partially_superseded_by + HISTORY 1행 + REQ-CL-002 옆 승계 주석"
  untouched: "REQ 본문 · version · status: completed 무변경"
docs_sweep:
  readme_locales: [README.md, README.ko.md, README.ja.md, README.zh.md]
  codemaps: [entry-points.md, overview.md, modules.md, data-flow.md, dependencies.md, docs-truth.md]
  template_paths_touched: 0   # 따라서 make build 불필요
  docs_site_touched: 0        # 런처 계약을 적은 docs-site 페이지 없음
verification:
  spec_lint_launch_verb: 0     # moai spec lint --strict <이 SPEC>/spec.md
  spec_lint_launcher_001: 0    # 승계 포인터를 단 뒤의 SPEC-CODEX-LAUNCHER-001/spec.md
push_pr_merge: none            # 푸시·PR·병합 없음. 통합 창은 리드 몫
```

### 승계하는 Gap (sync 에서도 닫지 않는다 — 산문으로 해소하지 않는다)

§E.2 의 Gap 표를 그대로 승계한다. 축자로 다시 적는다:

- 설치본 `~/go/bin/moai --help` 문안은 **재측정하지 않았다**.
- `codex app` 의 실제 기동 거동은 **미관측**이다 — `app` 토큰이 자식 argv 에 실린 것까지만 관측했다.
- codex 는 **최상위 help 만** 훑었다. 따라서 "`cli` 없음"과 "`-w` 없음" 두 부재 주장은 그만큼 좁다.
- 대화형 tty 왕복은 **구조적 미관측**이다 (SPEC-CODEX-LAUNCHER-001 0.8.0 승계).
- `moai spec lint` 는 `spec.md` 만 판정하므로 `plan.md`·`acceptance.md` 는 **세어 본 구조**에 근거한다.
- `-w` 절대경로 **ACCEPT 분기는 시험에서 구조적으로 도달 불가**하다 — 운영자 판정에 따라 gap 으로 보고하며, 지어낸 시험으로 우회하지 않는다.

### 범위 밖 관측 (손대지 않음)

`moai-mcp-tools-catalogue.md`(로컬·템플릿 두 사본)와 docs-site MCP 페이지가 `moai codex task` / `setup` / `job status` 를 "CLI equivalent" 로 적고 있으나 그런 CLI 동사는 없다. 이 카드 이전부터의 결함이며 리드가 별도 카드로 등록했다.
