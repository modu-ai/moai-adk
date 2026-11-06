auditor-model: claude-opus-5-5
verdict: PASS
score: 0.93
audited_sha: 9bce7aeb30169f5be15270470ec455821ab20a0f
audit_receipt: rcpt-4e0458009463c95dd20cd4a9

## Evaluation Report
SPEC: 없음 (카드 t2132, #1969 ② 3/3: Windows POSIX 권한 비트 테스트 실패 14건, 8개 패키지)
Overall Verdict: PASS. 1차의 코드 판정 PASS(0.93)는 그대로 유지한다. 1차를 FAIL로 만든 유일한 사유는 `codex` 필수 게이트에 영수증이 없다는 것이었고, 이번에 충족됐다. 운영자 결정 d-20261106T163140Z-56c1로 승인된 `codex_audit` 1회 호출이 `pass`를 반환했다(findings 0건, 영수증 `rcpt-4e0458009463c95dd20cd4a9`).

### Dimension Scores
| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 95/100 | PASS | 1차 측정(E2·E3)을 유지한다. 대상 14개 테스트를 호스트에서 돌리면 8개 패키지 모두 `ok`이고, Windows 게이트 overlay로 돌려도 모두 `ok`(의도된 `--- SKIP: TestErrorPaths/save-readonly-dir` 1건)다. Codex: `"findings":[]` |
| Security (25%) | 100/100 | PASS | 변경 파일은 `_test.go` 10개와 `.moai/reports/t2132/` 2개뿐이고, 운영 코드 변경은 없다(E1) |
| Craft (20%) | 88/100 | PASS | `golangci-lint run --new-from-rev=HEAD~1 <8 pkgs>` → `0 issues.`, `gofmt -l`는 출력 없음, `GOOS=windows go vet` exit 0(E4). progress.md의 원인 서술이 불완전한 점(F2)으로 감점했다 |
| Consistency (15%) | 92/100 | PASS | 15개 게이트가 모두 `runtime.GOOS != "windows" && <기존 비교>` 형태이고 같은 주석을 단다 |

Must-pass 방화벽(Functionality·Security): 둘 다 PASS. 필수 게이트 `codex`: 충족(verdict `pass`, 영수증 발급).

### Findings (structured defect-list)
Codex가 새로 낸 finding은 0건이라 재현할 대상이 없다. 1차 finding은 모두 optional이며 그대로 유지한다.
- F1 [Low] [optional] [outside] `internal/graph/gtd_private.go:72-93,117`: 운영 코드 `AuthorizePrivateGTDProjection`은 시스템 checker를 쓰므로 Windows에서 항상 `permission_denied`를 낸다. 사이트 3과는 무관하다(CI 분류 `OTHER`). 조치: 후속 카드를 발행한다.
- F2 [Low] [optional] [in-diff] `.moai/reports/t2132/progress.md:51`: 원인 서술에 소유자 확인 경로(`Win32FileAttributeData`에 `Uid` 필드 없음)가 빠져 있다. 조치: 후속 카드 본문에 적는다. 이 카드의 코드는 고칠 필요가 없다.
- F3 [Info] [optional] [in-diff] `internal/graph/meta_tmp_test.go:75-77`: 조기 `return` 뒤에 나중에 단계를 추가하면 Windows에서 조용히 빠진다. 조치 불필요.
- F4 [Info] [optional] [in-diff] `internal/homestate/profile_lease_test.go:30-37`: stat 오류를 분리하면서 nil 역참조 위험이 사라졌다. 결함이 아니라 개선이다.

blocking finding 없음.

## Claim
1. 14개 사이트 모두 운영자 정책(2026-11-06)대로 처리됐다. Windows에서는 이식 가능한 단언만 남기고, POSIX에서는 기존 단언을 그대로 유지한다(1차 판정 유지).
2. 독립 Codex 리뷰가 카드 커밋 범위(부모 `04d8c27a7e` 대비 delta)에서 결함을 찾지 못했다.
3. `codex` 필수 게이트가 충족됐으므로 전체 판정은 PASS다.

## Evidence
E1~E7: 1차 측정을 그대로 유지한다(리더 지시로 긴 스위트는 다시 돌리지 않았다). 1차 요지는 다음과 같다. 호스트 `go test` 8개 패키지 `ok`, overlay 시뮬레이션 `ok`(`grep -c SKIP sim.log` → `1`), `golangci-lint` `0 issues.`, `GOOS=windows go vet` exit 0, POSIX 기대값(`0o600`/`0o700`/`0o644`)이 삭제·추가 12줄씩 일치, 감사 계획 `"enforced_required": ["codex"]`.

E8 (이번 실행). 사전 확인:
```
$ git status --short          → (출력 없음)
$ git rev-parse HEAD          → e628fef6d04ef1a1e90b7102087244cc0fbc2509
$ git show --format='%H %P' 9bce7aeb30 → 9bce7aeb30169f5be15270470ec455821ab20a0f 04d8c27a7ed77afc09217f8623519806d48df3c4
```

E9 (이번 실행). `mcp__moai__codex_audit(project_root=<toplevel>, target=baseBranch, delta_base=04d8c27a7ed77afc09217f8623519806d48df3c4, mode=native)` 결과 발췌:
```
"verdict":"pass"
"summary":"... 지정한 기준 커밋 대비 변경에서 보고할 결함은 찾지 못했어. 정적 검토만 했고, ... 테스트·빌드·재현 명령은 실행하지 않았어."
"findings":[]
"resolved_model":"gpt-6.1-sol","requested_effort":"high","tool_surface":"codex_audit"
"audit_receipt":"rcpt-4e0458009463c95dd20cd4a9"
"build_commit":"04d8c27a7e"
"review_base":"delta_base (merge base 04d8c27a7ed77afc09217f8623519806d48df3c4)"
```

## Baseline-attribution
이번 실행의 측정 대상은 워크트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2132`(HEAD `e628fef6d0`, 작업 트리 clean)이다. Codex 리뷰 범위는 `04d8c27a7e...HEAD`이며, 카드 코드 커밋 `9bce7aeb30`과 그 위의 문서 전용 1차 판정 커밋 `e628fef6d0`을 포함한다. 리뷰를 수행한 MCP 서버 빌드는 `04d8c27a7e`로, 카드 커밋의 부모다. E1~E7은 1차 실행에서 같은 트리(HEAD `9bce7aeb30`)를 측정한 값이다.

## Gaps
- **Codex는 정적 검토만 했다**(요약에 직접 명시됨). 테스트 실행 근거는 1차 E2·E3이다.
- **긴 스위트는 다시 돌리지 않았다**(리더 지시). E1~E7은 이번 실행에서 다시 관측한 것이 아니라 1차 관측값이다. 그 사이 코드 변경은 없다. `9bce7aeb30` 이후 커밋은 문서 전용인 `e628fef6d0` 하나뿐이다.
- **도구 빌드 지연**: 서버가 `build_lag`를 보고했다(설치 바이너리 `04d8c27a7`는 HEAD의 조상). 카드 변경이 테스트 파일뿐이라 moai 바이너리의 동작이 판정 대상에 들어가지 않는다. 그래도 영수증 발급 경로는 이전 빌드에서 실행됐다.
- **전송 경로**: `codex_audit`의 provenance가 `"transport":"codex app-server"`이다. 이것은 승인된 MCP 도구가 내부에서 쓰는 경로이고, 감사자가 app-server 경로를 따로 호출한 것은 아니다. 리더가 금지한 "app-server 경로"에 해당하는지는 리더가 판단할 몫으로 남긴다.
- **판정 파일 추적 상태**: 리더 지시는 이 파일이 git-ignored라고 전제했다. 그러나 `git ls-files .moai/reports/t2132/`가 `sync-audit.md`를 출력했고 `git check-ignore`는 아무것도 출력하지 않았다. 즉 이 파일은 **추적되는 파일**이다. 이 덮어쓰기는 영수증 발급 뒤에 작업 트리를 바꾼다. 영수증의 트리 다이제스트가 미커밋 변경을 포함하는지는 확인하지 않았다.
- 실제 Windows에서는 실행하지 않았다(1차와 같음).

## Residual-risk
- 사이트 3처럼 그동안 Windows에서 `Fatalf`로 멈추던 테스트는, 게이트를 통과한 뒤 처음 실행되는 단계가 다른 이유로 실패할 수 있다(코드 판독상 위험은 낮다). 최종 판단 근거는 통합 브랜치 Windows CI다.
- F1 운영 코드 결함이 남아 있어 `TestPrivateGTDProjectionAccountAndCanonicalCollectorIsolation`은 Windows CI에서 계속 실패한다.
- 이 파일을 커밋하면 문서 전용 커밋이 하나 더 생긴다. 영수증에 묶인 트리와 달라지는지는 확인하지 않았다(Gaps 참조).

## Iteration history
| 회차 | 감사 SHA | 판정 | 점수 | 사유 |
|---|---|---|---|---|
| 1차 | 9bce7aeb30 | FAIL (코드 판정 PASS) | 0.93 | `codex` 필수 게이트 미충족(영수증 없음, 당시 리더가 미승인) |
| 2차(이번) | 9bce7aeb30 | PASS | 0.93 | 승인된 `codex_audit` 1회 → `pass`, findings 0건, 영수증 `rcpt-4e0458009463c95dd20cd4a9` |

## Recommendations
- `gtd_private.go`의 Windows 권한 확인(소유자 + 모드)을 다룰 후속 카드를 발행한다(F1·F2).
- 리더는 이 파일이 추적 대상이라는 점을 감안해 커밋 여부와 영수증 바인딩을 확인한다.
