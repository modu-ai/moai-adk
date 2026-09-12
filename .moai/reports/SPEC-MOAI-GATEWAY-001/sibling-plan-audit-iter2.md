# CG-RETIRE SP-B1 변경분 재감사
Iteration: 2
Verdict: PASS
Overall Score: 1.00
Scope: CG 로컬 이전·거절 경계 구현 준비; TEAMMATE native pane 완료 판정 제외

## Claim

SP-B1은 해소되었다. CG 0.1.0 보강 문서의 명령·저장 결과·소비 경계·고정 fixture는 로컬 구현을 시작할 만큼 구체적이다.
실제 capability가 없는 hybrid 이전/실행은 거절해야 한다. 이 PASS는 기존 CG 역할을 실제로 모두 보존했다는 판정이 아니다.
Reasoning context ignored per M1 Context Isolation. 현재 6종 문서와 sibling-plan-design.md를 읽고 이전 보고서의 SP-B1 및 직접 회귀만 판정했다.
검사한 실패 가능성은 명령 이름만 추가하고 저장/reader를 남겨 둔 경우, typed 저장으로 사용자 키를 잃는 경우,
기존 cg/GLM backend가 guard보다 먼저 실행되는 경우, 역할 변화 무음 승인, 사용자 verified 설정으로 hybrid 게이트를 우회하는 경우다.

## Regression Check

| 항목 | 판정 | 현재 문서 근거 |
|---|---|---|
| SP-B1 명령·명시 선택 | RESOLVED | design.md:11–22: `moai migrate cg`, preview 기본, 두 target, claude-only apply의 `--accept-role-change`, 잘못된 apply의 쓰기 0 |
| SP-B1 저장·reader | RESOLVED | design.md:20–28,30–40: team_mode와 gateway 두 키, 허용 조합, 새 reader/guard, 모든 launch 부작용 전 검증, mode:glm 충돌 거절 |
| SP-B1 고정 fixture | RESOLVED | acceptance.md:22–49: 이전 전후 YAML·원본 backup·credential 참조·주석 보존 및 hybrid의 정확한 delta |
| SP-A1 기존 CG guard 경합 | RESOLVED | acceptance.md:51–59: cc/glm/gpt·resume/spawn·-k/-f의 첫 guard, typed backend/credential/worktree/tmux/exec 계수 0, 정상 준비 seam 대조군과 guard 제거 뮤턴트 |
| 역할 보존 과장 회귀 | PASS | design.md:13–15,27–28: claude-only는 역할 변화 수락을 요구. :36–40: hybrid는 cc만 허용, apply와 매 launch 모두 실제 설치 capability 검사 |
| 사용자 데이터 보존 회귀 | PASS | design.md:42–49: typed 전체 저장 금지, YAML node 편집, 중복/alias 편집 실패, lock 후 원본 hash 재검사, owner-only 원본 backup·atomic replace·readback |

## Must-Pass Results

- MP-1 PASS: 현재 spec.md:31–45의 REQ 001…008과 AC 8개를 다시 추출했다. 누락·고아 연결 없음.
- MP-2 PASS: 요구사항 층의 When/Where/The system shall 계약은 유지된다. acceptance.md:20–59는 검증층의 고정 fixture 보강이다.
- MP-3 PASS: spec.md:2–14의 canonical 필드와 0.1.0 draft가 유지되며 지정 바이너리 lint가 통과했다.
- MP-4 N/A: 이 변경은 단일 저장소의 YAML/launcher 계약이다.
- MP-5 PASS (변경분): spec.md:15의 정식 형제 SPEC 참조는 iter1과 같으며 새 retired/superseded 참조를 도입하지 않았다.
- MP-6 PASS (변경분): 변경된 spec/design에 syscall 도입 없음.
- MP-7 PASS: plan.md/research.md의 clarification 표식 0. 이전의 본문 미결은 design.md:9–49로 대체되었다.
- MP-8 N/A: release-blocking AC/RED-now cell을 추가하지 않았다. acceptance.md:57–59의 미래 뮤턴트 요구를 현재 실행된 RED로 주장하지 않는다.

## Category Scores

| 항목 | 점수 | 근거 |
|---|---:|---|
| Clarity | 1.00 | design.md:11–40의 명령·결과·거절/소비 경계 |
| Completeness | 1.00 | design.md:42–49의 보존/오류와 plan.md:10–12의 선행 구현 순서 |
| Testability | 1.00 | acceptance.md:22–59의 입력/출력·계수·대조군·뮤턴트 |
| Traceability | 1.00 | 기존 REQ/AC 8쌍 유지, 새 fixture는 AC-CR-002/003에 명시 연결 |

## Defects Found

변경분에서 blocking 또는 optional 결함을 발견하지 않았다. iter1의 CG FAIL 0.81은 이 변경분 PASS로 대체한다.
TEAMMATE의 iter1 후보/preflight 한정 판정은 다시 감사하거나 확대하지 않았다.

## Evidence

지정 WT에서 직접 실행한 명령과 출력:

```text
$ git rev-parse --short HEAD
81c1d58f9
$ /tmp/moai-gateway-81c1d58f9 spec lint SPEC-MOAI-CG-RETIRE-001
✓ No findings — all SPEC documents are valid
```

각 exit 0. 바이너리는 부모가 해당 HEAD에서 빌드한 기존 파일이며 이번 감사가 재빌드한 것은 아니다.
Python pathlib/re로 현재 정의와 참조 차집합을 재추출한 출력(exit 0):

```text
REQ 8 AC 8 uncovered [] orphan []
clarification_markers 0
```

`rg -n 'Use:.*migrate|migrateCmd|"migrate"' internal/cli -g '*.go' -g '!**/*_test.go'` 출력 중 실제 명령 등록 근거(exit 0):

```text
internal/cli/migrate_agency.go:675:var migrateCmd = &cobra.Command{
internal/cli/migrate_agency.go:676:	Use:     "migrate",
internal/cli/migrate_agency.go:702:	rootCmd.AddCommand(migrateCmd)
internal/cli/migrate_agency.go:703:	migrateCmd.AddCommand(migrateAgencyCmd)
```

처음 추정한 internal/cli/migrate.go에 대한 rg는 파일 부재로 exit 2였다. 위 전체 CLI 검색으로 실제 파일 위치를 찾았으며 이를 명령 부재로 판정하지 않았다.
`sed -n '265,288p' internal/config/types.go`(exit 0)는 LLMConfig의 `Mode string`과 `TeamMode string` 및 cg/GLM backend 주석을 확인했다.
`rg -n 'GuardLegacyCG|ReadGatewayTeammatePolicy' internal/cli internal/config`는 출력 없이 exit 1이었다. 설계의 “구현 예정”과 일치하며 reader 도달성을 검증했다고 주장하지 않는다.
6종 문서는 nl/cat으로 읽었으며 기존 구현은 실행하지 않았다.

## Baseline-attribution

WT: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified
HEAD: 81c1d58f9. dirty WT의 현재 CG 0.1.0 문서만 변경분 판정했다.
직접 hashlib.sha256으로 측정한 파일 hash:

```text
acceptance.md 413ad375976932bb01c1df7faaedb139fefd1a504f75556cafaeb33b47fb4658
design.md 00e633f7be9e40012f58a1dc0a551fbd311ee60e3325382c6a94ae8abf16ea06
plan.md c94529b01094e6f42bfb821de967eb08e2e6095854efe290ac7c540e799f9a94
progress.md ca3c0cbab19068786079985e795dd1a18293bb959b47a308f9860b3007963a63
research.md 6e6193f9bf1ffac2c54c9818d676684d5825ffd25270a386dffb48c297a84710
spec.md 0d6e0150b211f1ed5daea1447aa3b3fa7225c8f859adde33c001f5a3702e270f
```

## Gaps

로컬 transaction/guard/reader 구현 및 fixture 실행은 아직 검증하지 않았다. Claude·provider·tmux 실행 없음.
실제 TEAMMATE native capability, 설치 버전별 지원, hybrid 역할 보존은 미검증이다. 코어·TEAMMATE 전체 재감사나 작성 중 CLI 코드 감사는 수행하지 않았다.

## Residual-risk

구현은 모든 launcher의 첫 부작용보다 앞에서 guard를 실행해야 한다. helper 존재나 설정 파일 결과만으로 AC-CR-002를 충족하지 않는다.
hybrid capability가 미확인/닫힘이면 apply는 쓰기 0이고, 이미 저장된 hybrid도 launch 실패여야 한다.
claude-only의 의미 변화 수락을 hybrid 동등성으로 기록할 수 없다. provider 실측은 기존 19:00 Asia/Seoul 이후 실행 제약을 따른다.

## Recommendation

CG 로컬 migration·첫 guard·reader 구현과 실패/정상 대조군 검증을 진행한다. 실제 TEAMMATE 게이트가 닫힌 동안 hybrid는 거절 동작을 구현하고 동등 이전 완료 표시는 보류한다.
