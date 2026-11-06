auditor-model: claude-opus-5-5
verdict: FAIL
score: 0.93
audited_sha: 9bce7aeb30169f5be15270470ec455821ab20a0f
code verdict: PASS

## Evaluation Report
SPEC: 없음 (카드 t2132, #1969 ② 3/3 — Windows POSIX 권한 비트 테스트 실패 14건, 8개 패키지)
Overall Verdict: FAIL — 코드 판정은 PASS(0.93)이다. FAIL이 된 이유는 오직 하나, 감사 계획상 `codex` 게이트가 `required`인데 리더 지시로 Codex를 호출하지 않아 영수증이 없기 때문이다. 이 게이트의 처분은 리더가 따로 결정한다.

### Dimension Scores
| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 95/100 | PASS | 대상 14개 테스트를 호스트에서 실행하니 8개 패키지가 모두 `ok`, Windows 게이트를 강제한 overlay 실행도 모두 `ok`(`--- SKIP: TestErrorPaths/save-readonly-dir`). 아래 Evidence E2·E3 참조 |
| Security (25%) | 100/100 | PASS | `git show --stat HEAD`: 변경 파일은 `_test.go` 10개와 `.moai/reports/t2132/` 2개뿐이고 운영 코드는 0개. diff 전체를 읽었으며 비밀값·입력 처리 변경 없음 |
| Craft (20%) | 88/100 | PASS | `golangci-lint run --new-from-rev=HEAD~1 <8 pkgs>` → `0 issues.`; `gofmt -l` 출력 없음; `GOOS=windows go vet` exit 0. 감점: progress.md의 운영 코드 결함 원인 서술이 불완전(F2) |
| Consistency (15%) | 92/100 | PASS | 게이트 15곳 모두 같은 `runtime.GOOS != "windows" && <기존 비교>` 형태이고 같은 주석을 단다. 같은 파일의 기존 Windows skip 관례(yamlpatch `read-only directory`)와도 일치 |

Must-pass 방화벽(Functionality·Security): 둘 다 PASS. 전체 FAIL은 코드 차원이 아니라 미충족 `codex` 필수 게이트 때문이다.

### Findings (structured defect-list)
- F1 [Low] [optional] [touched-file 아님, outside] `internal/graph/gtd_private.go:72-93,117` — 운영 코드 `AuthorizePrivateGTDProjection`은 시스템 checker로 Windows에서 항상 `permission_denied`를 낸다. 이 결함은 사이트 3(`TestBuildPrivateGTDProjectionPrivacy`)에는 닿지 않는다(해당 테스트는 `AuthorizePrivateGTDProjection`을 호출하지 않음). 드러나는 곳은 별도 테스트 `TestPrivateGTDProjectionAccountAndCanonicalCollectorIsolation`(`gtd_private_test.go:226`)이고, CI 분류표에서는 `OTHER`(1205행)로 이 카드 범위 밖이다. 별도 카드로 미루는 것은 타당하다. 확신도 높음. 필요 조치: 후속 카드를 발행한다(아래 F2의 원인 정정 포함).
- F2 [Low] [optional] [in-diff] `.moai/reports/t2132/progress.md:51` — 원인을 `mode.Perm() != item.mode` 비교 하나로만 적었다. 실제로는 `OwnerUID`도 Windows에서 실패한다. `info.Sys()`가 `*syscall.Win32FileAttributeData`이고 `Uid` 필드가 없어 `owner_unavailable`를 돌려주며, `user.Current().Uid`는 SID 문자열이다. 따라서 모드 비교만 고치면 이 경로는 여전히 거부된다. 확신도 높음(`GOOS=windows go doc syscall.Win32FileAttributeData` 출력으로 확인). 필요 조치: 후속 카드 본문에 소유자 확인 경로도 Windows 비이식 지점으로 적는다. 이 카드의 코드는 고칠 필요가 없다.
- F3 [Info] [optional] [in-diff] `internal/graph/meta_tmp_test.go:75-77` — 조기 `return`이 테스트 끝의 read-only 단계만 건너뛴다. 지금은 그 뒤에 이식 가능한 단계가 없다. 나중에 아래쪽에 단계를 추가하면 Windows에서 조용히 빠진다(progress.md Residual-risk에 이미 적혀 있음). 확신도 높음. 조치 불필요.
- F4 [Info] [optional] [in-diff] `internal/homestate/profile_lease_test.go:30-37` — 분리 전 코드는 `os.Stat` 오류 시 `info.Mode()`에서 nil 역참조 panic이 날 수 있었다. 이번 분리로 그 위험이 함께 사라졌고 POSIX 의미는 같다. 개선 사항이며 결함이 아니다.

blocking finding 없음.

### 사이트별 판정 (a)(b)(c)
(a) = 게이트 아래에 weak chmod 후 거부를 기대하는 단계가 남았는가 · (b) = 조기 skip/return이 이식 가능한 단언을 가리는가 · (c) = POSIX 중립인가(조건과 기대값 불변)

| # | 사이트 | (a) | (b) | (c) |
|---|---|---|---|---|
| 1 | civerdict `TestErrorPaths/save-readonly-dir` | 해당 subtest의 전부가 read-only 단계이고 skip 사유가 출력됨. 형제 subtest 3개는 chmod 없음 | skip 범위가 subtest 하나로 한정, 형제 subtest는 계속 실행(sim 로그 `--- PASS: TestErrorPaths/load-directory-as-file`) | `runtime.GOOS == "windows"`일 때만 skip |
| 2 | cli/update `TestSafeWriteFilePreservesExistingMode` | 이후 단계는 내용 확인뿐 | 해당 없음(내용 단언이 Windows에서도 실행) | `0o600` 유지 |
| 3 | graph `TestBuildPrivateGTDProjectionPrivacy` | 이후 단계(backup 목록, ambiguous 거부, revoke, fresh)에 chmod 없음 | stat 오류는 여전히 Fatal | `check.perm` 그대로, `err` 분리는 의미 동일 |
| 4 | graph `TestWriteEdgesMeta_FailureBranches` | read-only 단계 자체를 게이트 | `return` 뒤에는 read-only 단계만 있음. mkdir 실패 단계는 앞에서 실행 | Windows에서만 return |
| 5·6 | factory 2건 | 비교가 테스트 마지막 단계 | 해당 없음 | `0o600` 유지 |
| 7–10 | homestate 4건(게이트 5곳) | 모든 루프 반복이 `Errorf`/`Fatalf` 게이트 대상이고 이후 chmod 없음 | stat 오류는 여전히 Fatal | `0o700`/`0o600` 유지 |
| 11·12 | yamlpatch 2건 | `read-only directory` subtest는 기존부터 Windows skip | 내용 단언 실행 | `0o644` 유지 |
| 13 | template `file_mode_stays_0600` | 마지막 단계 | 해당 없음 | `0o600` 유지 |
| 14 | web `TestHandleSaveGreenfieldSectionCreation` | 마지막 단계 | 내용 단언은 앞에서 실행 | `0o644` 유지 |

## Claim
1. 14개 사이트 모두 운영자 정책(2026-11-06)대로 처리됐다. Windows에서는 이식 가능한 부분(존재·내용·stat 성공)만 단언하고, POSIX는 기존 단언을 그대로 유지한다. skip은 이식 가능한 부분이 남지 않는 civerdict subtest 한 곳뿐이며 사유를 출력한다.
2. 게이트를 Windows로 강제해도 대상 테스트는 모두 통과하며, 게이트 때문에 이식 가능한 단언이 사라진 곳은 없다.
3. 운영 코드 결함(`gtd_private.go`)은 사이트 3을 Windows에서 다시 실패시키지 않는다. 다른 테스트(`OTHER` 범주)에 속하므로 별도 카드로 미루는 것이 타당하다.

## Evidence
E1. `git rev-parse HEAD` → `9bce7aeb30169f5be15270470ec455821ab20a0f`. `git show --stat HEAD` → `12 files changed, 100 insertions(+), 14 deletions(-)`. Go 파일은 모두 `_test.go`.

E2. 호스트 실행 (`go test -count=1 <8 pkgs> -run '^(<14 names>)$' -v`, exit=0), 발췌:
```
    --- PASS: TestErrorPaths/save-readonly-dir (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/civerdict	0.186s
ok  	github.com/modu-ai/moai-adk/internal/cli/update	0.273s
ok  	github.com/modu-ai/moai-adk/internal/factory	0.515s
ok  	github.com/modu-ai/moai-adk/internal/graph	0.782s
ok  	github.com/modu-ai/moai-adk/internal/homestate	0.569s
    --- PASS: TestYAMLPatchAtomicWriteErrors/read-only_directory (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.116s
    --- PASS: TestSyncLocalSettingsPATH/file_mode_stays_0600 (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/template	0.371s
ok  	github.com/modu-ai/moai-adk/internal/web	0.583s
```

E3. Windows 게이트 시뮬레이션. 커밋 diff에서 추가된 줄 중 `runtime.GOOS`가 들어간 줄만 `"windows"`로 치환한 overlay를 만들었다(스크립트 출력: civerdict 1, cli/update 1, factory 2, graph 1+1, homestate 4+1, yamlpatch 2, template 1, web 1 = 15곳). 이 overlay로 `go test -count=1 -overlay <overlay.json> <8 pkgs> -run '^(<14 names>)$' -v`를 실행했다(exit=0), 발췌:
```
    --- SKIP: TestErrorPaths/save-readonly-dir (0.00s)
--- PASS: TestSafeWriteFilePreservesExistingMode (0.00s)
--- PASS: TestBacklogSQLiteArtifactsArePrivateWhileOpen (0.02s)
--- PASS: TestRelocateQueueArtifactsCopiesLogicalDataAndKeepsRollbackSource (0.03s)
--- PASS: TestBuildPrivateGTDProjectionPrivacy (0.05s)
--- PASS: TestWriteEdgesMeta_FailureBranches (0.10s)
--- PASS: TestHomeLayoutAndProjectRuntimePermissions (0.04s)
--- PASS: TestProjectLayoutPermissions (0.11s)
--- PASS: TestFactorySQLiteArtifactsArePrivateWhileOpen (0.12s)
--- PASS: TestProfileLeasesAreGlobalAndPrivate (0.01s)
    --- PASS: TestYAMLPatchAtomicWriteErrors/stat_missing_target (0.00s)
--- PASS: TestSyncLocalSettingsPATH (0.11s)
--- PASS: TestHandleSaveGreenfieldSectionCreation (0.06s)
```
SKIP은 의도된 civerdict 1건뿐이다(`grep -c SKIP sim.log` → `1`). 게이트 외의 이유로 실패한 테스트는 없다.

E4. `GOOS=windows go vet <8 pkgs>` → `vet_exit=0`. 호스트 `go vet` → `hostvet_exit=0`. `gofmt -l <8 dirs>` → 출력 없음. `golangci-lint run --new-from-rev=HEAD~1 <8 pkgs>` → `0 issues.`

E5. POSIX 중립성: `git diff -U0 HEAD~1 HEAD -- '*.go' | grep -E '^[-+].*0o[0-7]{3}'`로 확인한 결과, 삭제된 비교 12줄과 추가된 비교 12줄의 기대값(`0o600`/`0o700`/`0o644`)이 하나하나 같다. 나머지 3곳(gtd `check.perm`, civerdict skip, meta_tmp return)은 diff를 직접 읽어 확인했다.

E6. 운영 코드 경로: `grep -rn 'AuthorizePrivateGTDProjection\|SystemPrivateGTDPermissionChecker' --include='*.go'` 결과, 시스템 checker 호출은 `gtd_private_test.go:225,230`(`TestPrivateGTDProjectionAccountAndCanonicalCollectorIsolation`) 한 곳뿐이다. 사이트 3의 본문(61–111행)은 `BuildPrivateGTDProjection`·`PrivateGTDProjectionFresh`·`PrivateGTDBackupFiles`·`RevokePrivateGTDProjection`만 호출하고, 이 넷은 모드·소유자를 비교하지 않는다(`gtd_private.go:211-350` 확인). CI 분류표 1205행: `OTHER internal/graph TestPrivateGTDProjectionAccountAndCanonicalCollectorIsolation gtd_private_test.go:226: kernel owner/mode authorization: gtd projection: permission_denied`. `GOOS=windows go doc syscall.Win32FileAttributeData` → 필드는 `FileAttributes, CreationTime, LastAccessTime, LastWriteTime, FileSizeHigh, FileSizeLow`이며 `Uid`는 없다.

E7. 감사 계획: `moai verify audit-plan --project-root <toplevel>` → `"config_status": "ok"`, `"model": "codex"`, `"cross_model_required": true`, `"enforced_required": ["codex"]`, codex 게이트 `required`.

## Baseline-attribution
모든 측정은 이번 실행에서 워크트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2132`(HEAD `9bce7aeb30`, `git status` clean)를 대상으로 했다. 판정 도구는 로컬 `go` 툴체인과 `golangci-lint`이며 moai 빌드로 판정한 것은 없다. `moai verify audit-plan`은 설치된 바이너리로 실행했고, 감사 계획을 읽는 데만 썼다.

## Gaps
- **Codex 필수 게이트 미충족**: 감사 계획은 `codex: required`인데 리더 지시(이 카드에서 Codex 미승인)에 따라 `codex_audit`/`audit_multi`를 호출하지 않았다. 영수증이 없으므로 전체 판정을 FAIL로 기록했고, 코드 판정은 위 `code verdict: PASS` 줄과 같다. 처분은 리더 몫이다.
- 실제 Windows에서 실행하지 않았다(호스트는 darwin). E3 시뮬레이션은 게이트 분기만 Windows로 바꾼 것이고, 파일시스템 의미(chmod 무시, 0666/0777 보고)는 여전히 POSIX다. 실제 Windows 동작의 판정 근거는 통합 브랜치 CI다.
- 처음 시도한 `go test` 명령 두 개는 워크트리 격리 가드가 거부했다("command too complex to verify"). 같은 명령을 스크래치 스크립트(`run.sh`)로 옮겨 다시 실행했고 결과는 E2·E3과 같다. 측정을 대체한 것이 아니라 실행 방식만 바꾼 것이다. overlay 생성 스크립트도 같은 이유로 `git -C <toplevel>` 형태로 바꿔 실행했다.
- 전체 스위트는 실행하지 않았다(범위 정책). 대상 14개 테스트와 그 subtest만 실행했다.
- 대상 테스트 중 CI에서 게이트 줄보다 아래가 한 번도 실행되지 않은 구간(예: 사이트 3의 85–111행)은 코드를 읽어 판단했을 뿐, Windows에서 관측하지 않았다.

## Residual-risk
- 사이트 3처럼 기존에 `Fatalf`로 멈추던 테스트는, 게이트를 통과한 뒤 Windows에서 처음 실행되는 단계가 다른 이유로 실패할 수 있다. 코드 판독상 위험은 낮다(chmod·모드 비교가 없음).
- F1 운영 코드 결함은 남아 있다. `TestPrivateGTDProjectionAccountAndCanonicalCollectorIsolation`은 이 카드와 무관하게 Windows CI에서 계속 실패한다.
- 기계 로컬 스크래치(`.../scratchpad/sim.log`, `host.log`, `ov/`)는 이 파일에 옮긴 발췌 외에는 인용하지 않는다. 그 밖의 내용은 알려진 손실이다.

## Iteration history
- 1차(이번): 전체 FAIL(Codex 게이트) / 코드 PASS 0.93 @9bce7aeb30.

## Recommendations
- `gtd_private.go`의 Windows 권한 확인(소유자 + 모드 둘 다)을 다룰 후속 카드를 발행한다.
- Codex 게이트는 리더가 처분한다. 승인되면 이 SHA에 `codex_audit`를 한 번 실행해 영수증을 받는다.
