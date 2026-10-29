# SPEC-AGENTIC-FACTORY-001 — 수용 기준 (0.6.7)

## §A 측정 규칙

- 수용 기준은 25개다(Tier L 상한 25). AC-AF-0NN의 주 REQ는 spec.md §E 표에 있다. 기준마다 주 REQ는 하나다.
- **셈 규칙 두 가지**(iteration-5 D6):
  1. 상한 셈은 이 파일의 `#### AC-` 제목 수다. 명령: `grep -c '^#### AC-' .moai/specs/SPEC-AGENTIC-FACTORY-001/acceptance.md` → 25여야 한다.
  2. 저장소의 AC 계수기는 정규식 `AC-(?:[A-Z0-9]+-)*[0-9]+`(`internal/spec/ac_count_clause_test.go:563`)로 서로 다른 id를 모두 센다. 다른 SPEC의 id(예: 하위 SPEC의 AC-AFI-001 [REF])도 세므로, 이 파일에서 다른 SPEC id를 쓸 때는 바로 뒤에 같은 줄 공백으로 `[REF]`를 붙여 셈에서 뺀다.
- **판정**: 단위 시험 명령은 시험 이름을 고른다. 출력에 고른 시험마다 `--- PASS: <이름>`이 있고 `[no tests to run]`·`[no test files]`가 없으며, PASS 줄 수가 기대 수와 같을 때만 PASS다. 종료 코드 0만으로는 PASS가 아니다(빈 선택은 아무것도 단언하지 않는다).
- **`-run` 모양**: Go의 `-run`은 괄호 밖의 `/`와 `|`에서만 수준·분기를 나눈다(Go 1.26.8 `src/testing/match.go` `splitRegexp`, 이번 실행에서 읽음). 그래서 하위 시험은 `'^TestAcceptanceManifest$/^(…)$'`처럼 수준마다 고정하고, 최상위 시험과 섞을 때는 괄호 밖 `|`로 분기를 나눈다. `'^(TestAcceptanceManifest/M01|…)$'`처럼 괄호 안에 `/`를 두면 최상위 이름과 맞지 않아 manifest 하위 시험이 하나도 선택되지 않는다(0.6.0 작성 중 발견해 고침).
- **정본 시험**: design §15.1의 정본 id는 manifest 실행기의 하위 시험 `TestAcceptanceManifest/<id>`로 돈다(제안, 이름은 lane-22가 정함). 기준의 "정본 주 시험" 줄이 기대 PASS 수다. 보조 근거는 셈에 넣지 않는다.
- **시험 이름**: "제안"이라고 적힌 이름은 lane-22가 실제 이름을 정하면 이 파일이 따라 고친다. lane-22는 이전 제안 이름을 쓰지 않았다고 보고했다.
- **두 칸 채택**: 모든 기준은 아직 **채택 전**이다. 대상 패키지가 없어서(§B EV-REF-0, `TOOL_FAILURE`) 올바른 이유의 RED가 관측되지 않았다. run 단계가 각 기준의 시험에서 **의도한 단언 때문에** 실패한 `EXPECTED_RED`(명령, 원문 출력, 종료 코드, 트리 SHA)를 기록하면 RED-now 칸이 채워진다. "초록 경로" 줄이 짝이 되는 칸이다. 패키지 부재·컴파일 오류는 올바른 이유가 아니다.
- **F2 시험 이름 점검(0.6.7)**: 각 기준의 명령은 고정 후보 `4cdaba5d7e`에 실제로 있는 시험만 고르도록 다시 썼다. 시험이 없는 사례는 "미충족"으로 적고, 옛 이름은 "목표 시험 이름(아직 없음, 실행 대상 아님)" 목록에 남겼다. 기대 PASS 수는 "시험이 있는 사례만 센다"로 바꿨다. `TestAcceptanceManifest`(정본 시험 실행기)는 후보에 없으므로, 아래 "정본 시험" 줄은 목표이며 지금 실행할 수 없다. 자세한 것은 §G.
- **초록 경로 원칙**(하위 감사 D5와 같은 원칙): 어떤 기준도 그것이 단언하는 것을 만드는 마일스톤보다 먼저 초록이 되지 않는다.
- **단일 실행 원칙**: 시나리오 묶음(AC-AF-025)은 실제 Issue 한 쌍에 대해 한 번의 실행으로 01~07을 돈다. 채택 실행은 M7(잠금 해제) 뒤의 실행뿐이다. M7 전의 실행은 진단 실행이며 05·06에서 `producer-not-allowed`로 막히는 것이 정상이고, 어떤 기준의 채택 근거도 되지 않는다.
- 명칭은 spec.md §A.4를 따른다. 기준은 이름이 아니라 권한 경계와 동작을 단언한다.

### A.1 시나리오 하네스 계약 (AC-AF-025)

- **하나의 묶음**: `internal/agenticfactory/e2e/`의 `TestScenarioSuite` 하나 안에서 순서 있는 하위 시험 `01_…`~`07_…`. 모두 같은 실제 Issue(`MOAI_AF_E2E_ISSUE`)를 쓰고, 07은 대기 중인 두 번째 실제 Issue(`MOAI_AF_E2E_NEXT_ISSUE`)를 쓴다.
- **명령**: `MOAI_AF_E2E_ISSUE=<n> MOAI_AF_E2E_NEXT_ISSUE=<m> go test -tags af_e2e ./internal/agenticfactory/e2e/ -run '^TestScenarioSuite$' -count=1 -v`
- **판정**: 하위 시험 01~07 각각의 `--- PASS: <하위 시험 전체 이름> ` 줄(전체 이름 뒤 공백까지 일치) 일곱 줄이 모두 있어야 한다. 묶음의 `ok`나 종료 코드 0만으로는 PASS가 아니다.
- **귀속 원천 사전 점검**: 묶음은 시작할 때 귀속 원천 종류를 기록한다. 제품 원천이 아니면(대역이면) 01 전에 FAIL한다. 잠긴 트리에서 05·06은 `producer-not-allowed`로 막힌다.
- **선행 조건**: 하위 시험은 앞 단계가 남긴 상태(같은 Issue의 claim·attempt·HEAD)를 확인한다. 하나만 따로 돌리면 FAIL(skip 아님). 환경 변수가 없어도 FAIL.

| 하위 시험 | 실제로 하는 것 | 주입하는 것 |
|---|---|---|
| 01 경쟁 claim | 실제 claimant 두 개가 동시에 claim | 없음 |
| 02 중복 dispatch | 살아 있는 attempt가 있는 Issue로 실제 dispatch 재전송. 미전달 dispatch 분기는 격리 사본 | 없음 |
| 03 ACK 유실 | 실제 명령 전송 | 전송 계층에서 ACK만 버림 |
| 04 중단 복구 | agent 프로세스를 정상 종료 | 시계 주입으로 lease 만료. checkpoint `remote-durable`에서 새 epoch로 재개 |
| 05 감사 예산 소진 | 실제 감사 라운드 기록 | `audit.max_rounds=1` + 고정 FAIL 판정. PARKED `budget-exhausted` 뒤 운영자 승인 라운드로 재개 |
| 06 병합 결과 미상 | 아래 증거 생산 단계를 실제로 거친 뒤 실제 병합 요청 | 병합 응답만 버림 → MERGING + PARKED `unknown-effect` → 원격 재조회 |
| 07 원격 반영 → 종료 → DONE → 다음 dispatch | E12, 로컬 동기화, Issue 종료, DONE, 두 번째 Issue dispatch | 배포 단계가 없으면 사유 적힌 `deploy: N/A` |

- **05→06 증거 생산 단계**(06의 선행 조건, 주입 없음): 구현 커밋과 PR(agent, 현재 epoch) → PR 인계 receipt와 leader 수락(E03) → 필수 검사 3종 → 역할 부여(하위 SPEC 절차, 코드 리뷰·보안 리뷰·감사·QA 네 역할을 서로 다른 네 세션에. 어느 세션도 그 Issue의 writer가 아님) → 코드·보안 리뷰(design §11.1) → 무부채 감사 PASS → QA PASS → E08. 06은 시작할 때 이 증거가 같은 HEAD로 등록되어 있고 귀속 조회가 맞는 역할로 `attributed`인지 확인하며, 아니면 FAIL.
- **격리 실행**: 실제 Issue에서 돌지 않는 분기·변형(AC-AF-006 로컬 동기화 변형 등)은 격리 사본에서 돈다. control ref 대신 시험용 임시 ref를 쓰고, GitHub 쓰기 대신 호출 기록기를 붙인다. 로컬 동기화 변형은 독립 gitdir·독립 인덱스의 별도 클론에서 만든다(디렉터리 복사·`git worktree add` 금지).
- **격리 불변 단언**(격리 사례마다 전후 비교): ① 클론의 `git rev-parse --absolute-git-dir`이 실제 트리와 다름 ② 실제 트리의 `git rev-parse HEAD`, `git symbolic-ref HEAD`, `git status --porcelain=v2 -z`가 바이트 동일 ③ 실제 gitdir `index` SHA-256 동일 ④ 실제 트리 미커밋 경로마다 내용 SHA-256·인덱스 항목·모드 동일 ⑤ control ref 순번 동일.
- **재실행**: 묶음은 실제 Issue 한 쌍에 한 번만. FAIL이면 새 Issue 한 쌍과 새 control 시험 ref가 필요하다. 이전 실행의 기록은 읽기 전용으로 남긴다. 06에서 실제 병합이 일어났다면 되돌리지 않는다.
- **두 번째 Issue**: 07은 두 번째 Issue가 E01까지 간 뒤 하네스가 PARKED `operator-hold`(근거 원문 `scenario-suite-end`)로 두고, attempt가 실행되지 않았음을 단언한다. `operator-hold`를 PARKED 사유로 둘지는 운영자 결정 대기 OD-12(spec.md §G.1)이며, 결정이 이 값을 두지 않는 쪽이면 이 단언은 manager-spec 재위임으로 다시 쓴다.
- **운영자 keep-set**: 실제 Issue 댓글 쓰기(01~05), 보호된 `main` 실제 병합과 Issue 종료(06·07)는 운영자 승인 대상이다(OD-6). 승인 기록이 없으면 PARKED `operator-ack-missing`이며 PASS가 아니다.
- **Codex**: `workflow.agentic_factory.codex.enabled` 미설정으로 돈다. 끝에 Codex 호출 계수기 = 0. 계수기가 실제로 센다는 증거는 AC-AF-015의 `true` 사례다.

## §B 증거 원장

### B.1 대상 부재 참고 관측 (채택 기준점 아님)

```
id: EV-REF-0
command: go test ./internal/agenticfactory/... -count=1
stdout:
pattern ./internal/agenticfactory/...: lstat ./internal/agenticfactory/: no such file or directory
FAIL	./internal/agenticfactory/... [setup failed]
exit_code: 1
tree_sha: 5c3f0ebf27584d2c45deb2d00f9d69a4eea932c4
classification: TOOL_FAILURE (setup failed) — EXPECTED_RED 아님
```

0.6.0 작성 트리(`3275706cc9`)에서는 다시 실행하지 않았다. 이 관측은 "패키지가 아직 없다"는 사실만 보이며 어느 기준의 RED-now 칸도 되지 못한다.

### B.2 AC-AF-008 기준값 (regression-guard, 0.5.0의 EV-AC019-BASE)

```
id: EV-AC008-BASE
command: gh api repos/modu-ai/moai-adk/branches/main/protection --jq '[.required_status_checks.contexts, .enforce_admins.enabled, .allow_force_pushes.enabled, .allow_deletions.enabled, .required_linear_history.enabled]'
stdout:
[["Lint","Build (linux/amd64)","Analyze (Go) (go)"],true,false,false,false]
exit_code: 0
measured_at: 2026-10-29T04:08:11Z
tree_sha: e14b5bbac2eb2499cbaa3ae9434595001c91cb39 (관측 대상은 원격 보호 설정. 측정 시점 표시용)
```

원격 보호 설정은 현재 트리에서 다시 실행해 RED를 만들 수 없으므로, 이 값에 기대는 단언은 regression-guard이며 release-blocking이 아니다. AC-AF-008의 release-blocking 부분은 거절 경로 시험이다.

### B.3 control ref 존재 (이 실행)

```
id: EV-CTRL-REF
command: git ls-remote origin refs/heads/moai-control
stdout:
ae3a4edcaa529bc927357b345051b96cd04605a1	refs/heads/moai-control
exit_code: 0
tree_sha: 3275706cc9 (작성 트리. 관측 대상은 원격 ref)
```

`control/schema.json`의 sha256은 리더 보고이며 이 실행에서 재지 않았다.

## §C 수용 기준

#### AC-AF-001 — 활성 claim·fencing·중복 dispatch

- 주 REQ: maps REQ-AF-001. 초록: M1.
- 정본 주 시험 5: M01~M05. 전용: §12 재현 시험 `TestClaim_DeadOwnerExpiredAttempt_ReissuedUnderSameClaim`, `TestClaim_WorktreeIdentity_NormalizedPathPreventsMismatch`, `TestClaim_SameNameTree_NoDivergentTreeCreated`.
- 명령(F2, 있는 시험만 고름): `go test ./internal/agenticfactory/claim/ -run '^TestConcurrentAcquireExactlyOneWins$' -count=1 -v` 그리고 `go test ./internal/agenticfactory/e2e/ -run '^TestReal_DuplicateDispatch_NotReclaimed$' -count=1 -v`.
- 기대: 시험이 있는 사례만 센다 — 고른 시험 2개, 각 `--- PASS:` 한 줄. 존재만 확인했고 실행하지 않았다.
- 본문을 읽고 연결함: (a) 일부 — `TestConcurrentAcquireExactlyOneWins`는 동시 claim 16개 중 성공이 정확히 1개임을 단언한다(종료 10·충돌 기록은 단언하지 않음). (d) 일부 — `TestReal_DuplicateDispatch_NotReclaimed`는 같은 dispatch 재전송이 `ErrDuplicateDispatch`로 거절되고 기록의 epoch·seq가 그대로임을 단언한다.
- 미충족 — 시험 없음: (a)의 종료 10·충돌 기록, (b), (c), (e), 정본 M01~M05.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestAcceptanceManifest/M01`~`M05`, `TestClaim_DeadOwnerExpiredAttempt_ReissuedUnderSameClaim`, `TestClaim_WorktreeIdentity_NormalizedPathPreventsMismatch`, `TestClaim_SameNameTree_NoDivergentTreeCreated`.
- 사례: (a) 같은 Issue에 동시 claim 둘 → 정확히 하나 성공, 다른 하나 거절(종료 10)과 충돌한 claim 기록. (b) 재발급 → 같은 Issue 키 아래 새 attempt, 키 수 1. (c) 옛 epoch 쓰기 → 거절(종료 15), control ref 불변. (d) 살아 있는 attempt가 있는 Issue로 dispatch → 거절. (e) 미전달 dispatch가 있는 Issue로 dispatch → 거절.
- 변형을 죽이는 단언: 두 claim이 모두 성공하는 구현, 재발급이 새 키를 만드는 구현, epoch를 비교하지 않는 구현, 충돌 기록을 남기지 않는 구현은 각각 (a)(b)(c)(a)에서 FAIL.
- RED 이유: claim 저장 구현이 없어 (a)에서 단언 실패. 초록 경로: M1 lane-18 claim·fencing.

#### AC-AF-002 — 권위·재구축·checkpoint 복구

- 주 REQ: maps REQ-AF-002. 초록: M2.
- 정본 주 시험 2: SYNC12, M47. 전용(제안): `TestCheckpoint_*`, `TestAuthority_*`.
- 명령(F2, 있는 시험만 고름): `go test ./internal/agenticfactory/claim/ -run '^(TestAuthorityFailureLeavesTheCacheUntouched|TestStaleCacheCannotOverwriteAnotherWritersRevision|TestRebuildRestoresTheCacheFromTheAuthority|TestRebuildFailsClosedOnUnreadableAuthority)$' -count=1 -v` 그리고 `go test ./internal/agenticfactory/e2e/ -run '^TestCheckpoint_.*$' -count=1 -v`.
- 기대: 시험이 있는 사례만 센다 — 고른 시험 16개(claim 4 + e2e `TestCheckpoint_` 12), 각 `--- PASS:` 한 줄. 존재만 확인했고 실행하지 않았다.
- 본문을 읽고 연결함: A1 — `TestAuthorityFailureLeavesTheCacheUntouched`(권위 실패 → `ErrStoreUnavailable`, 캐시 불변), `TestRebuildFailsClosedOnUnreadableAuthority`(읽기 실패·깨진 자료 → `ErrStoreUnavailable`). A2 — `TestStaleCacheCannotOverwriteAnotherWritersRevision`(순번 충돌 → `ErrRevisionConflict`, 상태 불변). A3 — `TestRebuildRestoresTheCacheFromTheAuthority`(새 캐시가 권위에서 같은 상태로 복원). e2e `TestCheckpoint_*` 12개는 이름이 있어 명령에 남겼고, 본문을 CK1~CK10에 하나씩 대응시키지는 않았다.
- 미충족 — 시험 없음: A4, 정본 SYNC12·M47. CK1~CK10 중 어느 것이 위 12개로 덮이는지는 확인하지 않았다.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestAcceptanceManifest/SYNC12`, `TestAcceptanceManifest/M47`, `TestAuthority_*`(0개).
- 권위 사례: A1 control ref 읽기 실패 → 의존 전이 없음, `store-unavailable`. A2 CAS 실패 → 전이 없음. A3 로컬 캐시 삭제 → control ref에서 재구축한 상태가 삭제 전과 같음. A4 control ref에 없는 claim이 로컬 캐시에만 있음 → claim으로 인정하지 않음.
- checkpoint 사례(design §4.6): CK1 K1 뒤 ACK 전 충돌 → `pending`(`committed-unacked`), `remote-durable` 아님. CK2 ACK는 있으나 manifest readback 불일치 → `acked-manifest-unverified`, `recovered` 불가. CK3 K4 Issue 링크 실패 → 여전히 `remote-durable`, 링크 재시도가 예약됨. CK4 같은 ref·SHA로 두 번째 push 요청 → push 호출 0, 거절. CK5 이전 writer가 `live`인데 복구 요청 → 거절(`ownership-unverified`), 새 writer 0. CK6 사용자의 미커밋 변경이 있는 트리에서 checkpoint → 그 경로의 내용 digest·인덱스 항목 불변, 커밋에 포함 0. CK7 push 대상이 기본 branch → 거절. CK8 non-fast-forward → 강제 push 0, PARKED. CK9 manifest 후보에 C1 값 → 기록 거절 또는 `[REDACTED:C1]`, 원값 검색 0. CK10 상한 초과 → 가장 오래된 제거 가능 항목만 색인에서 빠지고, 제거 가능 항목이 없으면 `capacity-short`로 새 checkpoint 거절, 기존 항목 불변.
- 변형 단언: ACK 없이 `remote-durable`을 세우는 구현은 CK1, manifest를 검증하지 않는 구현은 CK2, 링크 실패로 durable을 취소하는 구현은 CK3, 중복 검사 없는 구현은 CK4(push 호출 계수 1), 활성 판정 없는 복구는 CK5, `git add -A`식 스테이징은 CK6, 상한 무시는 CK10에서 FAIL.
- RED 이유: control ref git 구현과 checkpoint 상태 기계가 없음. 초록 경로: M1 control ref(lane-18), M2 checkpoint·Issue 링크(lane-18·20).

#### AC-AF-003 — 소유 확인과 Issue 쓰기 권한

- 주 REQ: maps REQ-AF-003. 초록: M2.
- 정본 주 시험 2: M13, M15. 전용(제안): `TestIssueAuthority_*`.
- 명령(F2, 있는 시험만 고름): `go test ./internal/agenticfactory/e2e/ -run '^(TestAuthz_SameNameDifferentID_DecidedByIDOnly|TestAuthz_UnknownIdentity_BlocksEveryOperation|TestAuthz_RedirectOrRename_IsRefused|TestAuthz_RepositoryComparison_PerProvider|TestAuthz_OwnIssue_LabelAndAssignee_RefusedEvenWithEveryKey|TestPreflightPR_ForeignClosingKeyword_Refuses)$' -count=1 -v`.
- 기대: 시험이 있는 사례만 센다 — 고른 시험 6개, 각 `--- PASS:` 한 줄. 존재만 확인했고 실행하지 않았다.
- 본문을 읽고 연결함: IA3 일부 — `TestAuthz_SameNameDifferentID_DecidedByIDOnly`(요청 구조에 이름 필드가 없고 id가 숫자형임. 224행 뒤 판정 부분은 읽지 않음). IA4 — `TestAuthz_UnknownIdentity_BlocksEveryOperation`(id 0·음수 → 모든 동작 `identity-unknown`). IA9·IA10 — `TestAuthz_RedirectOrRename_IsRefused`(redirect·rename·응답 없음 → `repository-not-allowed`). IA12 — `TestAuthz_OwnIssue_LabelAndAssignee_RefusedEvenWithEveryKey`(자기 Issue 레이블·담당자 → `permission-not-granted`). IA14 일부 — `TestPreflightPR_ForeignClosingKeyword_Refuses`(`Closes`·`fixes`·`Resolved #99` → `ErrForeignClosing`, 상태 불변. 대소문자 혼합·URL 형태·PR effect 전송 0은 읽은 줄에 없음). IA15 — `TestAuthz_RepositoryComparison_PerProvider`(GitHub 대소문자 무시 허용, GitLab 거절).
- 미충족 — 시험 없음: IA1, IA2, IA5~IA8, IA11, IA13, IA16, 정적 단언, 정본 M13·M15. 같은 파일에 다른 후보 시험이 있으나 본문을 읽지 않아 연결하지 않았다.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestAcceptanceManifest/M13`, `TestAcceptanceManifest/M15`, `TestIssueAuthority_*`(0개).
- 사례(기대 사유는 design §2.1a):
  - IA1 claim 없음·holder 불일치·control ref 읽기 실패 → 쓰기 거절 `ownership-unverified`, 대체 쓰기 0.
  - IA2 자기 Issue 본문 편집 → 허용.
  - IA3 로그인 이름은 같고 숫자 id가 다름 → `not-issue-author`. 로그인 이름은 다르고 숫자 id가 같음 → 허용(로그인 비교 변형을 죽임).
  - IA4 작성자 숫자 id 조회 실패 → `identity-unknown`, 쓰기 0.
  - IA5 남의 Issue 관련 댓글 → 허용.
  - IA6 남의 Issue 본문 편집, 협업 꺼짐 → `not-issue-author`.
  - IA7 남의 Issue 본문 편집, 협업 켜짐·목록 안·`edit_body: false` → `permission-not-granted`. 같은 조건에서 `edit_body: true` → 허용.
  - IA8 남의 Issue claim, 협업 켜짐·목록 밖 → `repository-not-allowed`.
  - IA9 목록 안 저장소가 301로 답함 → `repository-not-allowed`.
  - IA10 이름이 바뀐 저장소의 옛 이름이 목록에 있음(응답 저장소 ≠ 요청 저장소) → `repository-not-allowed`.
  - IA11 남의 Issue claim, 기존 owner 인계 미확인 → `owner-unconfirmed`.
  - IA12 자기 Issue 레이블 변경, `permissions.labels: true` → `permission-not-granted`. 담당자도 `permissions.assignees: true`에서 같음.
  - IA13 자기 Issue 제목 편집·마일스톤 변경·다시 열기, 모든 `permissions.*` 키 `true` → 각각 `op-not-defined`.
  - IA14 PR 본문 `Fixes #N`(N은 닫으면 안 되는 Issue), 대소문자 혼합 `fIxEs #N`, URL 형태 → `foreign-closing-keyword`, PR effect 전송 0.
  - IA15 GitHub에서 목록 `Owner/Repo`, 요청 `owner/repo` → 허용. GitLab에서 같은 쌍 → `repository-not-allowed`.
  - IA16 같은 업무의 기존 Issue가 있을 때 자기 Issue 발행 요청 → 새 Issue 0.
- 정적 단언(IA 묶음 안): `internal/config`에 위 8개 사유 문자열이 0건, 판정 도우미 이름 `ForeignIssueAllows` 0건.
- 변형 단언: 로그인 비교(IA3), 권한 키가 자기 Issue 동작을 여는 구현(IA12·IA13), redirect를 따라가 허용하는 구현(IA9), 목록만 보고 인계를 보지 않는 구현(IA11), 키워드 검사 없는 구현(IA14)이 각각 FAIL.
- RED 이유: 판정 코드와 effect 경계의 연결이 없음. 초록 경로: M2(lane-18 판정 + lane-20 effect + 설정 연결).

#### AC-AF-004 — 활성 판정·인계·복구 조회

- 주 REQ: maps REQ-AF-004. 초록: M1.
- 정본 주 시험 7: M06~M12.
- 명령(F2, 있는 시험만 고름): `go test ./internal/agenticfactory/claim/ -run '^(TestLiveOwnerExpiredLockStaysBusy|TestUnreadableOrIndeterminateOwnerStaysBusy|TestReusedPidWithDifferentFingerprintIsDead)$' -count=1 -v`.
- 기대: 시험이 있는 사례만 센다 — 고른 시험 3개, 각 `--- PASS:` 한 줄. 존재만 확인했고 실행하지 않았다.
- 본문을 읽고 연결함: (a) 일부 — `TestLiveOwnerExpiredLockStaysBusy`(`live` + 오래된 잠금 → 재발급 후보 아님, 재발급 `ErrBusy`. `lapsed` 기록은 단언하지 않음). (b) 일부 — `TestUnreadableOrIndeterminateOwnerStaysBusy`(판정 불가·읽을 수 없는 잠금 → 후보 아님, `ErrBusy`). (c) 일부 — `TestReusedPidWithDifferentFingerprintIsDead`(같은 pid·다른 fingerprint → 재발급 후보. epoch+1·보존은 단언하지 않음).
- 미충족 — 시험 없음: (d), (e), 정본 M06~M12.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestAcceptanceManifest/M06`~`M12`, `TestLiveness_*`(0개), `TestLiveness_CrashHandover_ResponsibilityPassesWithEpoch`.
- 사례: (a) lease 만료 + 판정 `live` → 인계 거절, attempt `lapsed`. (b) `indeterminate` → 같음. (c) 같은 pid·다른 fingerprint → `dead` → epoch+1, 이전 worktree·branch·SHA·dirty 목록 보존·참조. (d) 재시작·끊김 뒤 복구 → 실제 실행 조회가 재개·재발급·PARKED 결정보다 먼저(호출 순서 기록), 강제 종료 0, 중복 attempt 0.
- (e) **시나리오 "crash 인계"**(0.6.7, DEC-AF-23, REQ-AF-009의 책임 기록을 함께 단언, 제안 이름 `TestLiveness_CrashHandover_ResponsibilityPassesWithEpoch`): PR을 연 Issue의 책임 agent 프로세스가 죽음(판정 `dead`) → 새 agent에게 epoch+1로 책임이 넘어감, control ref의 책임 기록이 새 agent를 가리킴, 책임자 없는 시간 0(책임 기록이 비는 순번 0), 이전 worktree·branch·변경 보존, 같은 Issue의 attempt 동시 둘 0. 판정이 `live`·`indeterminate`면 책임 이전 0.
- 변형 단언: PID만 보는 판정은 (c)에서 FAIL(재사용된 pid를 `live`로 봄). TTL만으로 인계하는 구현은 (a)에서 FAIL. crash 뒤 책임 기록을 비워 두는 구현은 (e)에서 FAIL.
- RED 이유: 판정 연결 없음. 초록 경로: M1(lane-18, `homestate.ClassifyOwnerWith` 사용).

#### AC-AF-005 — effect 불명 재조회·전달 상태

- 주 REQ: maps REQ-AF-005. 초록: M2.
- 정본 주 시험 10: M18, M19, M20, M29, M30, M31, SYNC03, SYNC04, SYNC05, SYNC11.
- 명령(F2, 있는 시험만 고름): `go test ./internal/agenticfactory/effect/ -run '^(TestAmbiguousSendIsReconciledNeverResent|TestIntentWriteFailureMeansNoSend)$' -count=1 -v`.
- 기대: 시험이 있는 사례만 센다 — 고른 시험 2개, 각 `--- PASS:` 한 줄. 존재만 확인했고 실행하지 않았다.
- 본문을 읽고 연결함: (c) 일부 — `TestAmbiguousSendIsReconciledNeverResent`(응답 불명 → 조회 1회·전송 1회로 재전송 없음, 조회 불확실·오류 → `StatusParked`. 사유 문자열 `unknown-effect`는 읽은 줄에 없음). (d) — `TestIntentWriteFailureMeansNoSend`(intent 기록 실패 → 전송 0).
- 미충족 — 시험 없음: (a), (b), (e), 정본 M18~M20·M29~M31·SYNC03~SYNC05·SYNC11.
- 목표 시험 이름(아직 없음, 실행 대상 아님): 위 정본 id의 `TestAcceptanceManifest/<id>`, `TestDelivery_*`(0개), `TestDelivery_ReplyLoss_RecoveredFromRecordNotReply`.
- 사례: (a) 전송 성공만 있음 → `sent`만, `received` 이상 0. (b) idle 알림 → `completed` 아님. (c) effect 응답 불명 → 재발행 전 원격 조회 1회 이상, 재발행 0, PARKED `unknown-effect`. (d) intent 기록 전에 전송을 시도하는 경로 → 거절(intent 없는 전송 0).
- (e) **시나리오 "답장 유실 복구"**(0.6.7, DEC-AF-23, 제안 이름 `TestDelivery_ReplyLoss_RecoveredFromRecordNotReply`): agent가 일을 끝냈지만 완료 답장이 전송 계층에서 버려짐 → 진행은 control ref·증거 기록을 읽어 복구되고 다음 단계로 감, 같은 명령 재전송 0, 답장 대기 중에 남긴 대기 기록은 design §2.1c 여섯 필드를 모두 가짐. 답장만 잃고 기록도 없으면 `completed` 0이고 대기 기록의 재확인 시각에 상승 대상으로 넘어감.
- 변형 단언: 답장으로 `completed`를 세우는 구현은 (b), 불명 시 재전송하는 구현은 (c), 답장이 올 때까지 멈추는 구현은 (e)에서 FAIL.
- RED 이유: effect 경계 없음. 초록 경로: M2(lane-20).

#### AC-AF-006 — DONE 순서

- 주 REQ: maps REQ-AF-006. 초록: M3.
- 정본 주 시험 4: M33, M34, M35, M36.
- 명령(F2): 없음 — 이 기준의 사례를 단언하는 시험을 찾지 못해, 아무것도 고르지 않는 명령을 두지 않는다.
- 미충족 — 시험 없음: (a)~(e) 전부, 정본 M33~M36.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestAcceptanceManifest/M33`~`M36`, `TestDone_*`(0개).
- 사례: (a) E12 뒤 Issue 종료 effect의 control ref 순번 < DONE 기록 순번. (b) 종료 실패 → MERGING + PARKED `pending-effect`, revert effect 0. (c) 로컬 동기화 변형: 동기화 전 수정 경로의 내용이 바뀌고 경로 목록은 같음 → DONE 거절. 인덱스만 바뀜, 모드만 바뀜도 각각 거절. (d) `deploy: N/A` 사유 없음 → DONE 거절. (e) DONE 전 같은 자리 다음 dispatch → 거절. 이관 기록은 DONE 0.
- 변형 단언: DONE 뒤 종료하는 구현(0.5.0 순서)은 (a)에서 FAIL. 경로 목록만 비교하는 구현은 (c)에서 FAIL.
- RED 이유: 상태 엔진의 E13·E14 없음. 초록 경로: M3(lane-18). lane-18 기존 시험(이름 미전달)은 보조 근거다.

#### AC-AF-007 — 병합 게이트·PR 적체 상한

- 주 REQ: maps REQ-AF-007. 초록: M4.
- 정본 주 시험 8: M14, M24, M25, M26, M27, M32, M37, M55.
- 명령(F2): 없음 — 표 행·적체 사례를 단언하는 시험을 본문으로 확인하지 못했다.
- 미충족 — 시험 없음: 게이트 표 8행, 적체 (a)(b), 정본 M14·M24~M27·M32·M37·M55.
- 목표 시험 이름(아직 없음, 실행 대상 아님): 위 정본 id의 `TestAcceptanceManifest/<id>`, `TestGate_ReasonTable_*`(0개), `TestPRQueue_*`(0개).
- 게이트 사례(게이트 단위 시험, 귀속 대역 허용): design §2.1b 표의 8행마다 하나. 각 행의 기대 사유는 표와 같고 한 결과는 사유 하나만 남긴다(기록 수 = 1). 여기에 skipped·누락·실패·취소·다른 HEAD 검사 → `check-not-passed`, 감사 예산 소진 → `budget-exhausted`.
- 적체 사례: (a) 열린 인계 + 수리 적체 = `wip_limit` → 새 Issue dispatch 0, 종결·수리 effect는 계속. (b) 키 값 `0`, `-1`, `"x"`, 범위 밖 정수 → 설정 로드가 키 이름과 허용 범위를 밝혀 거절(`0`을 끔으로 읽는 변형을 죽임). 키 미설정 사례는 두지 않는다. 미설정 동작은 운영자 결정 대기(OD-7)이며, 결정이 나면 manager-spec 재위임으로 사례 (c)를 넣고 기대 PASS 수를 고친다.
- 변형 단언: `PASS-WITH-DEBT`를 통과로 세는 구현, 귀속 실패를 판정값보다 늦게 보는 구현(한 결과에 사유 둘), `operator-ack-missing`을 역할 부재에 쓰는 구현은 표 행에서 FAIL.
- RED 이유: 게이트 없음. 초록 경로: M4(lane-18 판정 + lane-20 effect).

#### AC-AF-008 — 보호 비약화

- 주 REQ: maps REQ-AF-008. 초록: M2.
- 정본 주 시험 0. 전용: lane-22 `TestProtectionGuard_*` 3개 + 약화 거절 시험(제안 `TestProtectionGuard_RefusesEveryWeakeningOperation`).
- 명령(F2, 있는 시험만 고름): `go test ./internal/agenticfactory/e2e/ -run '^TestProtectionGuard_.*$' -count=1 -v`.
- 기대: 시험이 있는 사례만 센다 — 고른 시험 3개(레이블 endpoint, 쓰기 기본값 꺼짐, 제품 소스 용어), 각 `--- PASS:` 한 줄. 존재만 확인했고 실행하지 않았다.
- 미충족 — 시험 없음: W1~W7 약화 거절.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestProtectionGuard_RefusesEveryWeakeningOperation`.
- 사례: design §11.3 W1~W7 각각 → 전송 0, 상태 불변, 사유 기록 정확히 1개(`permission-denied`). 레이블 쓰기 경로는 악의적 이름으로도 레이블 endpoint를 벗어나지 않음. 쓰기 기본값 꺼짐에서 전송 0. 제품 소스에 보호 설정 변경 용어 0.
- regression-guard(release-blocking 아님): EV-AC008-BASE의 다섯 값이 착지 전후 같음.
- 변형 단언: 일곱 중 하나라도 보내는 구현, 사유를 두 개 남기는 구현이 FAIL.
- RED 이유: 거절 경로가 코드에 없음(lane-22 보고). 초록 경로: M2(lane-20 집행, lane-18 사유 타입).

#### AC-AF-009 — PR 인계·수리 단일 소유

- 주 REQ: maps REQ-AF-009. 초록: M3.
- 정본 주 시험 8: HANDOFF01~HANDOFF08.
- 명령(F2): 없음 — `runner`·`claim`에 인계·수리 후보 시험이 있으나 본문을 읽지 않아 연결하지 않았다.
- 미충족 — 시험 없음: (a)~(g) 전부, 정본 HANDOFF01~HANDOFF08.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestAcceptanceManifest/HANDOFF01`~`08`, `TestHandoff_*`·`TestRepair_*`·`TestResponsibility_*`(AF 범위 0개), `TestResponsibility_CIFailure_SameAgentRepairsThroughDone`, `TestResponsibility_SecondIssue_OnlyWhenConditionsHold`.
- 사례: (a) receipt 필드(design §4.8 PRHandoff — lane-18 이름으로 새 필수 필드 `PR URL`·`base`·`adopted SPEC`·`task result`·`check status`와 유지 필드 `handoff id`·`LeaderAccepted`·`PR number`·`HEAD`·`branch`·`worktree`·`WriterStopped`·`BoundaryHead`) 하나씩 빼기 → E03 거절, 상태 BUILDING 유지. `PR URL`이 http/https가 아니면 거절. (b) PR 발행만 있음 → DONE도 REVIEWING도 아님. (c) E03 뒤 agent의 실행 자리를 비움(새 Issue를 받을지는 acquire 검사가 판정), 책임 기록(Holder)은 그대로 같은 agent, worktree·branch·변경 보존(경로 digest 불변). (d) E06 수리 → 원 Issue·branch에 배타 claim, 같은 worktree·epoch 연속성, 그 Issue의 책임 agent에 배차, 그 agent의 다른 Issue 기록 불변. (e) MERGING 중 HEAD 변경 → E10으로 REVIEWING, 이어 E06.
- (f) **시나리오 "CI 실패 수리부터 종료까지"**(0.6.7, DEC-AF-23, 제안 이름 `TestResponsibility_CIFailure_SameAgentRepairsThroughDone`): 일부러 실패하게 만든 필수 검사 → 같은 agent가 수리 → 다시 CI → 리뷰 → 병합 → DONE. 단언: 수리·리뷰 대응 attempt의 writer가 모두 처음 배정된 agent와 같음, leader(Coordinator) 쓰기 주체의 구현·수리 커밋 0, 각 단계 순번이 CI 실패 < 수리 커밋 < 재검사 PASS < 리뷰 PASS < 병합 < Issue 종료 확인 < DONE, 책임 기록은 DONE 전까지 비지 않음. 시나리오는 격리 사본(acceptance §A.1 격리 실행 규칙)에서 돈다.
- (g) **다른 Issue를 맡는 조건**(0.6.7, 제안 이름 `TestResponsibility_SecondIssue_OnlyWhenConditionsHold`): PR을 연 agent에게 두 번째 Issue 배차 요청 → ① 원 Issue 수리가 대기 중이면 거절, 수리가 먼저 배차됨 ② `wip_limit`에 닿았으면 거절(설정값 N 사례만. 미설정 사례는 두지 않는다 — OD-7) ③ 두 번째 Issue의 쓰기 worktree가 첫 Issue의 것과 같으면 거절 ④ 첫 Issue의 책임 기록이 control ref에 없으면 거절 ⑤ 넷이 모두 맞으면 허용, 두 Issue 모두 같은 agent가 책임자.
- 변형 단언: 필드 하나 없이 인계를 받는 구현은 (a), 수리를 새 Issue로 여는 구현은 (d), 바쁜 agent의 Issue를 덮는 구현은 (d), PR 인계로 책임을 끝내는 구현은 (c)·(f), 수리를 다른 agent에 주는 구현은 (d)·(f), worktree를 같이 쓰는 구현은 (g) ③에서 FAIL.
- RED 이유: E03·E06·E10 없음. 초록 경로: M3(lane-18).

#### AC-AF-010 — 독립성과 역할 쓰기 경계

- 주 REQ: maps REQ-AF-010. 초록: M4.
- 정본 주 시험 2: M22, M23.
- 명령(F2): 없음 — 이 기준의 사례를 단언하는 시험을 찾지 못했다.
- 미충족 — 시험 없음: (a)~(f) 전부, 정본 M22·M23.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestAcceptanceManifest/M22`, `TestAcceptanceManifest/M23`, `TestIndependence_*`(0개), `TestIndependence_LeaderDoesNotTakeOverAgentWork`.
- 사례: (a) ControlService가 아닌 쓰기 주체의 claim·lease·liveness·resume 쓰기 → 거절. (b) Coordinator의 직접 상태 쓰기 → 거절. (c) 같은 Issue의 어느 attempt writer였던 세션의 리뷰 증거 → 세지 않음. (d) Coordinator가 직접 구현한 변경의 리뷰 → 세지 않음. (e) 다른 Issue의 writer였던 세션이 이 Issue를 리뷰 → 셈(정본 [2]§21).
- (f) leader 역할 경계(0.6.7, DEC-AF-23, 제안 이름 `TestIndependence_LeaderDoesNotTakeOverAgentWork`): Coordinator가 agent에 배정된 Issue의 구현·수리·리뷰 대응 쓰기를 요청 → 거절, 그 Issue의 attempt 기록 불변. 변경 알림·실패 복구·배정 조정·통합 지원 요청 → ControlService 명령으로 받아들여짐. 이 사례는 위 PASS 줄 수에 1을 더한다.
- 변형 단언: run 전체를 막는 과잉 구현은 (e)에서 FAIL, Issue 범위를 보지 않는 구현은 (c)에서 FAIL, leader가 agent 일을 대신하는 구현은 (f)에서 FAIL.
- RED 이유: 역할 경계 없음. 초록 경로: M4. (e)는 귀속 대역으로 게이트 단위 시험 안에서만 돈다.

#### AC-AF-011 — 역할 귀속과 잠금

- 주 REQ: maps REQ-AF-011. 초록: M4(잠긴 형태). M7에서 L1·L2를 다시 쓴다.
- 정본 주 시험 0. 전용(제안): `TestGate_RoleEvidence_RequiresChildAttribution_AndLockHolds`, `TestLock_*`.
- 명령(F2, 있는 시험만 고름): `go test ./internal/agenticfactory/e2e/ -run '^TestRoleLock_NoAttributor_BlocksACompletePassSet$' -count=1 -v`.
- 기대: 시험이 있는 사례만 센다 — 고른 시험 1개, `--- PASS:` 한 줄. 존재만 확인했고 실행하지 않았다.
- 본문을 읽고 연결함: L2 일부(L1 일부 포함) — 잠긴 runner에 전부 통과인 증거 묶음 → `verifying` + `producer-not-allowed` 관문, 관문 문장에 `identity-spec-incomplete`가 있음. 증거 종류별 여섯 사례는 아니다.
- 미충족 — 시험 없음: L1 종류별 6, L3, L4, P4a, P4b(같은 파일의 writer 인계 후보 시험은 단언 줄을 읽지 않아 연결하지 않음).
- 미충족 — 구현 없음: 해제 뒤 형태(L1·L2). 감사 보고상 귀속 조회 구현(`runner.Attributor`, `runner.Facts`)은 0개다(리더 전달, 이 작성자는 그 밖을 읽지 않음).
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestGate_RoleEvidence_RequiresChildAttribution_AndLockHolds`, `TestLock_*`(0개).
- 사례(잠긴 형태):
  - L1 제품 귀속 원천에 증거 종류 6개(code-review, security-review, audit, qa, writer-handoff, coordinator-acceptance)를 각각 넣음 → 모두 `unattributable{identity-spec-incomplete}`.
  - L2 같은 증거로 게이트를 돌림 → 종류마다 `producer-not-allowed` 하나.
  - L3 설정 키(임의의 `workflow.agentic_factory.*` 잠금 이름 후보 포함)와 환경 변수(잠금 해제를 뜻하는 이름 후보)를 바꿔 같은 L1·L2 실행 → 결과 바이트 동일. 정적 검사: 귀속 원천 패키지가 환경 변수·설정·파일을 읽는 호출 0, 잠금 상수를 읽는 분기 정확히 1곳.
  - L4 시나리오 하네스 사전 점검에 귀속 대역을 주입 → 01 전에 FAIL. 제품 원천이면 사전 점검 통과.
  - P4a 다른 세션이 현재 epoch로 writer 인계 등록 → `producer-not-allowed`(잠금 중에도, 해제 뒤에도).
  - P4b 같은 세션이 옛 epoch로 writer 인계 등록 → 귀속 조회 전에 epoch 거절(종료 15), 귀속 조회 호출 0.
- 해제 뒤 형태(M7에서 L1·L2만 바꿈): L1 → 하위 SPEC 부여로 등록된 세션의 증거가 `attributed{필요한 역할}`. L2 → 그 증거가 §2.1b "통과로 셈" 행. L3·L4·P4a·P4b는 그대로.
- 변형 단언: 환경 변수로 풀리는 구현(L3), 시나리오에서 대역을 쓰는 구현(L4), fencing보다 귀속을 먼저 보는 구현(P4b 귀속 호출 1)이 FAIL.
- RED 이유: 잠금 상수도 귀속 원천도 없음(lane-22 보고). 초록 경로: M4 잠금 도입(lane-18). M7 해제(lane-18, 하위 SPEC 완료와 같은 착지).

#### AC-AF-012 — 접수·준비

- 연결: maps REQ-AF-012, REQ-AF-018. 주 REQ는 REQ-AF-012이고, REQ-AF-018은 보조 연결이다(REQ-AF-018에는 주 AC가 없다 — spec.md §E). 초록: M3.
- 정본 주 시험 10: PREP01~PREP10.
- 명령(F2, 있는 시험만 고름): `go test ./internal/agenticfactory/e2e/ -run '^TestIntake_.*$' -count=1 -v`.
- 기대: 시험이 있는 사례만 센다 — 고른 시험 12개, 각 `--- PASS:` 한 줄. 존재만 확인했고 실행하지 않았다. 12개를 (a)~(e)에 하나씩 대응시키지는 않았다.
- 미충족 — 시험 없음: 정본 PREP01~PREP10. (a)~(e) 중 어느 것이 위 12개로 덮이는지는 확인하지 않았다.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestAcceptanceManifest/PREP01`~`PREP10`.
- 사례: (a) control ref 기록 전 충돌 → 수락 응답 0. (b) 상태가 design §4.2의 7행 밖으로 이동 요청 → 거절. 6단계 이름만 가진 구현은 `BRANCH_REGISTERED`가 없어 FAIL. (c) READY ACK 전 개발 launch → 0. (d) `intake.max_concurrent`를 N으로 설정했을 때 동시 준비 N+1번째 → 대기. (e) 기한 초과 → 예약 해제. 키 미설정 사례는 두지 않는다. 미설정 동작은 운영자 결정 대기(OD-9)이며, 결정이 나면 manager-spec 재위임으로 사례를 넣고 기대 PASS 수를 고친다.
- 변형 단언: 수락을 먼저 돌려주는 구현은 (a), 무한 하위 작업은 (d)에서 FAIL.
- RED 이유: 접수 서비스 없음. 초록 경로: M3(lane-18 상태, lane-20 Issue 발행).

#### AC-AF-013 — 비밀·공개 검토

- 주 REQ: maps REQ-AF-013. 초록: M5.
- 정본 주 시험 3: M39, M40, CONTEXT04. 전용: `TestRedact_*` 13개(C1 3, C2 1, C3 2, C4 2, C5 1, Restricted 1, Idempotent 1, Negative 2).
- 명령(F2, 있는 시험만 고름): `go test ./internal/agenticfactory/core/ -run '^(TestRedactMasksSecretsAndCounts|TestRedactMasksEmailsAndAbsoluteHomePaths)$' -count=1 -v`.
- 기대: 시험이 있는 사례만 센다 — 고른 시험 2개, 각 `--- PASS:` 한 줄. 존재만 확인했고 실행하지 않았다.
- 본문을 읽고 연결함: C1 일부 — `TestRedactMasksSecretsAndCounts`(토큰·Bearer·`api_key=` 세 값 가림, 가린 수 3, 일반 문장 불변. 표시는 `[REDACTED]`이며 `[REDACTED:C1]`이 아님). C2·C3 일부 — `TestRedactMasksEmailsAndAbsoluteHomePaths`(이메일·ssh 원격·macOS·Linux·Windows 홈 경로가 남지 않음. C2의 이름·전화번호는 없음).
- 미충족 — 시험 없음: C4, C5, Restricted, 범주별 게시 거절(Negative), 필드 owner 쓰기 거절, 관찰 읽기 쓰기 0, 정본 M39·M40·CONTEXT04.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestRedact_*` 13개(0개 있음), `TestLedger_FieldOwner_*`(0개), `TestAcceptanceManifest/M39`·`M40`·`CONTEXT04`.
- 사례: 범주별 가림 출력, `RESTRICTED`는 상태·요약·SHA-256 세 필드만, 범주 하나씩 꺼 둔 변형 refiner에서 게시 거절이 다섯 번 모두 일어남, 게시 산출물 전체 원값 검색 0. 필드 owner가 아닌 역할의 장부 쓰기 → 거절. 관찰 읽기 경로에 쓰기 호출 0.
- 변형 단언: 범주 하나를 빠뜨린 구현은 Negative 시험에서 FAIL.
- RED 이유: refiner 없음. 초록 경로: M5(lane-20).

#### AC-AF-014 — 인증·rate limit·예산

- 주 REQ: maps REQ-AF-014. 초록: M2.
- 정본 주 시험 15: M17, M28, M43, M44, GH05, GH06, SYNC09, SYNC10, WT15, SYNC13~SYNC18. 전용(제안): `TestForgeClassify_*`.
- 명령(F2, 있는 시험만 고름): `go test ./internal/agenticfactory/forge/ -run '^(TestClassifyTable|TestUnsignalled403IsUnclassifiedNotPermissionNotRateLimit)$' -count=1 -v`.
- 기대: 시험이 있는 사례만 센다 — 고른 시험 2개, 각 `--- PASS:` 한 줄. 존재만 확인했고 실행하지 않았다.
- 본문을 읽고 연결함: G6 — `TestUnsignalled403IsUnclassifiedNotPermissionNotRateLimit`(신호 없는 403 → `KindUnclassified`, 호출 1회로 재시도 0, intent는 park). 분류 표 일부 — `TestClassifyTable`(200·304·429·403 rate limit 행 등. G1~G8·L1~L6에 행별로 대응시키지 않음).
- 미충족 — 시험 없음: 예산 (a)~(d)(`forge`의 예산 후보 시험은 본문을 읽지 않음), G6 밖 분류 행의 행별 대응, 정본 15개.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestForgeClassify_*`(0개), `TestBudget_*`(AF 범위 0개; `scripts/i18n-validator`의 같은 접두는 관계없음), 위 정본 id의 `TestAcceptanceManifest/<id>`.
- 분류 사례: design §4.5 G1~G8, L1~L6 각각 → 표의 분류와 재시도 수. G6 → `unclassified-403`, 재시도 0.
- 예산 사례: (a) 재개 시각 전 같은 자원 요청 0. (b) 헤더 없음 → `UNKNOWN`, 남은 몫 있음으로 보지 않음. (c) 설정 기본값을 바꾸면 대기 동작이 따라 바뀜(코드 상수 변형을 죽임). (d) 예비 몫 아래에서 안전 쓰기 > 안전 알림 > 보고 댓글 순서. 다른 자격·endpoint로 전환 0.
- 변형 단언: 403을 모두 권한으로 보는 구현은 G1·G2, 모두 rate limit으로 보는 구현은 G5·G6에서 FAIL.
- RED 이유: 분류기 없음. 초록 경로: M2(lane-20).

#### AC-AF-015 — 공급자 관문

- 주 REQ: maps REQ-AF-015. 초록: M6.
- 정본 주 시험 1: M21.
- 명령(F2, 있는 시험만 고름): `go test ./internal/agenticfactory/review/ -run '^TestImportGraph_NoCodexSurfaceNoGLM$' -count=1 -v`. 이어 이전 단계 기록 점검(design §9.7 CX-1~CX-3).
- 기대: 시험이 있는 사례만 센다 — 고른 시험 1개, `--- PASS:` 한 줄. 존재만 확인했고 실행하지 않았다.
- 본문을 읽고 연결함: (d) 근사 — `review` 패키지의 의존 목록에 codex 어댑터·glm 패키지가 없음을 단언한다(정적 import 검사이며 호출 계수가 아님).
- 미충족 — 시험 없음: (a), (b), (c), 정본 M21.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestAcceptanceManifest/M21`, `TestProvider_*`(0개).
- 사례: (a) `codex.enabled` 미설정 → Codex 호출 계수 0. (b) `false` → 0. (c) `true` → 계수 1 이상(계수기가 실제로 셈). (d) GLM 백엔드 호출 0.
- 이관 점검: `progress.md` §J의 CX 행 세 개가 있고, 커밋 그래프로 CX-1 ⊂ CX-2 ⊂ CX-3, 그 사이 모든 커밋에서 Codex 기본값 짝이 `run_base_sha`와 같음.
- 변형 단언: 계수기가 늘 0인 구현은 (c)에서 FAIL.
- RED 이유: 관문 없음. 초록 경로: M6(lane-17 키, lane-20 호출 경계).

#### AC-AF-016 — 전이표

- 주 REQ: maps REQ-AF-016. 초록: M1.
- 정본 주 시험 6: M16, M41, M42, M48, M54, M56. 전용: `TestTransition_Table_*`, §12 `TestState_SpeclessIssue_TransitionsWithoutSpecHash`, `TestState_OperatorDecision_RequiresOperatorEvidence_CoordinatorRefused`.
- 명령(F2, 있는 시험만 고름): `go test ./internal/agenticfactory/core/ -run '^(TestEdgeTableIsExactlyE01ToE27|TestEveryEdgeAcceptsExactlyItsFromStages)$' -count=1 -v`.
- 기대: 시험이 있는 사례만 센다 — 고른 시험 2개, 각 `--- PASS:` 한 줄. 존재만 확인했고 실행하지 않았다.
- 본문을 읽고 연결함: (a) 일부 — `TestEdgeTableIsExactlyE01ToE27`(간선 id가 정확히 E01~E27), `TestEveryEdgeAcceptsExactlyItsFromStages`(간선마다 시작 단계만 받음, E20 제외). 쌍 수 출력은 단언하지 않음.
- 미충족 — 시험 없음: (b)~(e), 정본 M16·M41·M42·M48·M54·M56. `core`에 다른 후보가 있으나 본문을 읽지 않아 연결하지 않았다.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestTransition_Table_*`(0개), `TestTransition_Table_WaitRecordRequiresAllSixFields`, `TestState_SpeclessIssue_TransitionsWithoutSpecHash`, `TestState_OperatorDecision_RequiresOperatorEvidence_CoordinatorRefused`, 위 정본 id의 `TestAcceptanceManifest/<id>`.
- 사례: (a) 상태 7 × 사건 전부의 표 구동 시험: E01~E27에 있는 쌍만 적용, 나머지는 거절하고 상태 불변(쌍 수를 출력에 기록하고 0이 아님을 단언). (b) 요청자가 맞지 않음 → 거절. (c) PARKED 사유 집합 밖 값 → 거절. (d) E20 → 단계·owner 불변. (e) 대기 기록 여섯 필드(design §2.1c, 0.6.7, 제안 이름 `TestTransition_Table_WaitRecordRequiresAllSixFields`): 필드를 하나씩 뺀 대기 기록 여섯 개 → 각각 무효로 거절, 이전 상태 불변. 여섯 필드가 모두 있는 기록 → 받아들여짐.
- 변형 단언: 표에 없는 간선 하나를 허용하는 구현은 (a), owner를 바꾸는 재개는 (d), 필드가 빠진 대기 기록을 받는 구현은 (e)에서 FAIL.
- RED 이유: 상태 엔진 없음. 초록 경로: M1(lane-18).

#### AC-AF-017 — record와 수집기

- 주 REQ: maps REQ-AF-017. 초록: M1.
- 정본 주 시험 11: M45, M49, M50, M51, M52, M53, SYNC01, SYNC02, SYNC06, SYNC07, SYNC08.
- 명령(F2, 있는 시험만 고름): `go test ./internal/agenticfactory/e2e/ -run '^TestCollector_.*$' -count=1 -v`.
- 기대: 시험이 있는 사례만 센다 — 고른 시험 6개, 각 `--- PASS:` 한 줄. 존재만 확인했고 실행하지 않았다. 6개를 (a)(b)에 대응시키지는 않았다.
- 미충족 — 시험 없음: 정본 11개(M45·M49~M53·SYNC01·SYNC02·SYNC06~SYNC08). (a)(b)가 위 6개로 덮이는지는 확인하지 않았다.
- 목표 시험 이름(아직 없음, 실행 대상 아님): 위 정본 id의 `TestAcceptanceManifest/<id>`.
- 사례: (a) 같은 저장소에 수집기 둘 시작 → 하나만 활성. (b) agent 패키지에서 polling 경로 호출 0(정적).
- RED 이유: record 없음. 초록 경로: M1(lane-18).

#### AC-AF-018 — 맥락 장부와 증거 위치

- 주 REQ: maps REQ-AF-019. 초록: M5.
- 정본 주 시험 5: CONTEXT01, CONTEXT02, CONTEXT03, CONTEXT05, CONTEXT06. 전용: `TestLedger_*`, `TestEvidenceStage_*`.
- 명령(F2): 없음 — `effect/ctxledger`에 후보 시험이 있으나 본문을 읽지 않아 연결하지 않았다.
- 미충족 — 시험 없음: (a)~(g) 전부, 정본 CONTEXT01·02·03·05·06.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestLedger_*`·`TestEvidenceStage_*`(AF 범위 0개; `internal/jevledger`의 `TestLedger_`는 관계없음), 위 정본 id의 `TestAcceptanceManifest/<id>`.
- 사례: (a) 미추적·ignore 경로만 있음 → `inaccessible-local`, 링크 게시 0. (b) 로컬에만 있는 커밋 → `inaccessible-unpushed`, 오프라인 `accessible` 0. (c) HEAD 변경 → `STALE`, 항목 수 불변, 원 판정값 보존. (d) blob 바뀜 → `STALE`. (e) 원격 조회 실패 → `UNKNOWN`. (f) 단계 `PUSHED`에서 원격 조회 전 `AVAILABLE` 0. (g) 링크가 branch 이름 URL → 거절(전체 커밋 SHA만).
- 변형 단언: 로컬 브랜치 정보로 `accessible`을 내는 구현은 (b), STALE 항목을 지우는 구현은 (c)에서 FAIL.
- RED 이유: 장부 없음. 초록 경로: M5(lane-20).

#### AC-AF-019 — 관찰 보기

- 주 REQ: maps REQ-AF-020. 초록: M5.
- 정본 주 시험 0. 전용(제안): `TestObservationView_ShowsPrepAndPRBacklogStatesDistinctly_ReadOnly`.
- 명령(F2, 있는 시험만 고름): `go test ./internal/agenticfactory/e2e/ -run '^TestObservation_Reading_ChangesNothing$' -count=1 -v`.
- 기대: 시험이 있는 사례만 센다 — 고른 시험 1개, `--- PASS:` 한 줄. 존재만 확인했고 실행하지 않았다.
- 본문을 읽고 연결함: (c) 일부 — 세 번 읽기 전후 요청 목록이 같음(control ref 순번·트리 키는 읽은 줄에 없음).
- 미충족 — 시험 없음: (a), (b), (d). 같은 파일의 다른 관찰 시험은 본문을 읽지 않아 연결하지 않았다.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestObservationView_ShowsPrepAndPRBacklogStatesDistinctly_ReadOnly`(`TestObservationView_*` 0개).
- 사례: (a) design §8.1 항목마다 값·원천·읽은 시각. (b) 준비 상태 7행과 PR 적체가 BUILDING 이후 작업과 다른 구역에 표시. (c) 실행 전후 control ref 순번과 트리 키 동일, 쓰기 호출 0. (d) 출력에 C1~C5 원값 0. 종료 코드는 design §8.
- 변형 단언: 준비 중인 Issue를 시작한 작업으로 세는 구현은 (b), 읽다가 상태를 고치는 구현은 (c)에서 FAIL.
- RED 이유: 관찰 보기 없음. lane-20의 monitor는 포화·적체 값만 보인다는 보고가 있다(미검증). 초록 경로: M5(lane-20).

#### AC-AF-020 — 설정·브라우저

- 주 REQ: maps REQ-AF-021. 초록: M6.
- 정본 주 시험 9: M38, M46, M46a~M46g.
- 명령(F2, 있는 시험만 고름): `go test ./internal/agenticfactory/browser/ -run '^TestChoices_ExactlyThree$' -count=1 -v`.
- 기대: 시험이 있는 사례만 센다 — 고른 시험 1개, `--- PASS:` 한 줄. 존재만 확인했고 실행하지 않았다.
- 본문을 읽고 연결함: (c) — `Choices()`가 정확히 egolite·aside·chrome 셋.
- 미충족 — 시험 없음: (a), (b), (d), (e), (f), 정본 M38·M46·M46a~M46g. `browser`의 해석 시험과 `internal/config`의 설정 시험은 단언 줄을 읽지 않아 연결하지 않았다.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestAgenticFactoryConfig_*`(0개), `TestBrowserSelect_*`(0개), 위 정본 id의 `TestAcceptanceManifest/<id>`.
- 사례: (a) 범위 밖 값 → 키·범위·파일을 밝혀 거절, 깎기 0. (b) 읽기 오류 → 미설정으로 처리 0. (c) 선택지는 정확히 세 개. (d) 미선택 → `chrome`. (e) `aside` 선택·미설치 → PARKED `not-installed`, `chrome` 호출 0. (f) 선택 저장 뒤 다시 읽기 → 같은 값.
- 변형 단언: 미설치 시 대체하는 구현은 (e)에서 FAIL.
- RED 이유: 설정 연결 없음(읽기 코드는 lane-17 브랜치에 있으나 이 트리에 없음). 초록 경로: M6.

#### AC-AF-021 — CLI

- 주 REQ: maps REQ-AF-022. 초록: M6.
- 정본 주 시험 23: CLI01~CLI23.
- 명령(F2): 없음 — 이 기준의 사례를 단언하는 시험을 찾지 못했다.
- 미충족 — 시험 없음: (a)~(g) 전부, golden 순서 점검, 정본 CLI01~CLI23.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestAcceptanceManifest/CLI01`~`CLI23`, `TestCompat_*`(0개), `TestLauncher_*`(`internal/cli`·AF 범위 0개; `internal/lsp`·`internal/sandbox`의 같은 접두는 관계없음).
- 사례: (a) S1: golden 파일과 도움말·거절 문구·출력·종료 코드 바이트 동일. (b) S2: stderr 한 줄만 추가. (c) 버전 선언 없이 새 문법 → 호환 오류(종료 2). (d) 새·옛 문법 혼합 → 호환 오류. (e) 역할 뒤바뀜 0: 모든 플래그 조합에서 세션 역할이 선언과 같음. (f) 조용한 흡수 0: 모르는 인자는 거절. (g) `--` 뒤 인자는 해석 없이 전달.
- golden 순서 점검: golden 커밋이 S1 등록 커밋의 엄격한 조상.
- RED 이유: 새 명령 없음. 초록 경로: M6(launcher 담당). OD-2의 4건이 정해지면 (c)(d)의 문법을 고친다.

#### AC-AF-022 — template·분류·setup

- 주 REQ: maps REQ-AF-023. 초록: M6.
- 정본 주 시험 28: WT01~WT14, WT16, GH01~GH04, GH07~GH15.
- 명령(F2): 없음 — manifest 실행기가 없다.
- 미충족 — 시험 없음: 정본 28개 전부, 추가 사례 둘.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestAcceptanceManifest/WT01`~`WT14`·`WT16`, `GH01`~`GH04`·`GH07`~`GH15`.
- 추가 사례: 공급자가 지원하지 않는 계층 기능 → `not supported` 보고. 보안 실패 댓글은 묶음 window를 기다리지 않음.
- RED 이유: template 연결 없음. 초록 경로: M6.

#### AC-AF-023 — 환경 행렬

- 주 REQ: maps REQ-AF-024. 초록: M6.
- 정본 주 시험 0. 전용(제안): `TestEnvMatrix_StepEvidenceRequired_NotRunNeverSuccess`.
- 명령(F2): 없음 — 이 기준의 사례를 단언하는 시험을 찾지 못했다.
- 미충족 — 시험 없음: (a)~(d) 전부.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestEnvMatrix_StepEvidenceRequired_NotRunNeverSuccess`(`TestEnvMatrix_*` 0개).
- 사례: (a) 단계 칸에 명령·종료 코드·출력 중 하나라도 없음 → 행 거절. (b) 실행 안 한 단계 → `not-run`, `PASS` 0. (c) 전부 `BLOCKED`인 행렬 → 행렬 FAIL. (d) Codex 미설정 환경 → `BLOCKED(not-configured)`.
- RED 이유: 구현 없음(lane-22 보고). 초록 경로: M6.

#### AC-AF-024 — 효율 측정 기록

- 주 REQ: maps REQ-AF-025. 초록: M5.
- 정본 주 시험 6: EFF01~EFF06.
- 명령(F2): 없음 — 이 기준의 사례를 단언하는 시험을 찾지 못했다.
- 미충족 — 시험 없음: (a)~(d) 전부, 정본 EFF01~EFF06.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestAcceptanceManifest/EFF01`~`EFF06`, `TestEfficiencyRecord_*`(0개).
- 사례: (a) design §4.7 필드 누락 → 기록 거절. (b) `estimated` 값을 `measured` 칸에 넣음 → 거절. (c) 품질 회귀가 있는 비교 → 효율 향상 보고 0. (d) checkpoint 기록의 비용 필드는 측정값이거나 Gap 표시, 추정값 0.
- RED 이유: 기록 형식 없음. 초록 경로: M5(lane-22).

#### AC-AF-025 — 실제 Issue 시나리오

- 주 REQ: maps REQ-AF-006(보조: REQ-AF-001·004·005·011). 초록: M7 뒤 단일 실행.
- 정본 주 시험 0. 전용: `TestScenarioSuite`(§A.1).
- 명령: §A.1. **F2**: `TestScenarioSuite`가 후보에 없어 §A.1 명령은 지금 아무것도 고르지 않는다 — 실행 대상이 아니다.
- 미충족 — 시험 없음: 01~07 전부. `e2e`의 `TestScenario_*` 7개는 실제 Issue 한 쌍의 단일 묶음이 아니고 본문을 읽지 않아 연결하지 않았다.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestScenarioSuite`와 하위 시험 `01_…`~`07_…`.
- 판정: §A.1의 판정 규칙(하위 시험 01~07의 전체 이름 PASS 줄 일곱 개). 07은 Issue 종료 순번 < DONE 순번 < 다음 dispatch 순번을 단언한다.
- RED 이유: 하네스 없음. M7 전에는 잠금 때문에 05·06이 `producer-not-allowed`로 막힌다(올바른 이유의 RED 후보). 초록 경로: M7 잠금 해제 + 운영자 keep-set 승인(OD-6).

## §D 변형 점검표 (문서 추론)

이 표는 이 판에서 문서만 보고 만든 **추론**이다. 실제 변형 코드를 돌려 본 것이 아니다. run 단계가 각 변형을 실제로 만들어 해당 사례가 FAIL하는지 관측해야 한다.

| AC | 변형(요구를 어기면서 얕은 기준을 통과할 수 있는 구현) | 죽이는 사례 |
|---|---|---|
| 001 | 동시 claim 둘 다 성공 | (a) |
| 001 | 재발급이 새 키 생성 | (b) |
| 002 | ACK 없이 durable | CK1 |
| 002 | 링크 실패로 durable 취소 | CK3 |
| 002 | `git add -A` 스테이징 | CK6 |
| 003 | 로그인 이름 비교 | IA3 |
| 003 | 권한 키가 자기 Issue 레이블을 엶 | IA12 |
| 003 | redirect 추종 | IA9·IA10 |
| 004 | PID만 비교 | (c) |
| 005 | 답장으로 `completed` | (b) |
| 006 | DONE 뒤 Issue 종료 | (a) |
| 006 | 경로 목록만 비교 | (c) |
| 007 | 한 결과에 사유 둘 | 표 행 기록 수 = 1 |
| 007 | `wip_limit: 0`을 끔으로 해석 | 적체 (b) |
| 008 | 약화 연산 하나 전송 | W1~W7 |
| 004 | crash 뒤 책임 기록 비움 | (e) |
| 005 | 답장이 올 때까지 멈춤 | (e) |
| 009 | 필드 빠진 receipt 수락 | (a) |
| 009 | PR 인계로 책임 종료 | (c)·(f) |
| 009 | 수리를 다른 agent에 배차 | (d)·(f) |
| 009 | 두 Issue가 쓰기 worktree 공유 | (g) ③ |
| 010 | leader가 agent 일을 대신 | (f) |
| 016 | 필드 빠진 대기 기록 수락 | (e) |
| 010 | run 전체 금지 | (e) |
| 011 | 환경 변수로 잠금 해제 | L3 |
| 011 | 시나리오에서 대역 사용 | L4 |
| 011 | 귀속을 fencing보다 먼저 | P4b |
| 012 | 6단계 상태 | (b) |
| 013 | 범주 하나 누락 | Negative |
| 014 | 403을 모두 권한으로 | G1·G2 |
| 014 | 코드 상수 대기 | (c) |
| 015 | 늘 0인 계수기 | (c) |
| 016 | 표 밖 간선 허용 | (a) |
| 018 | 로컬 정보로 `accessible` | (b) |
| 019 | 읽다가 상태 수정 | (c) |
| 020 | 미설치 시 대체 | (e) |
| 021 | 모르는 인자 흡수 | (f) |
| 023 | 미실행을 성공으로 | (b) |
| 024 | 추정값을 측정 칸에 | (b) |
| 025 | 종료 전 DONE | 07 순번 단언 |

## §E 완료 정의

이 SPEC의 run 단계가 끝났다고 말하려면 다음이 모두 있어야 한다.

1. AC-AF-001~024가 같은 HEAD에서 PASS(각 AC의 기대 PASS 줄 수 일치, 빈 선택 0). 보조 근거는 셈에 넣지 않는다.
2. M7 착지에서 AC-AF-011의 L1·L2가 해제 뒤 형태로 다시 쓰이고 그 형태로 PASS.
3. M7 뒤 시나리오 묶음 단일 실행에서 AC-AF-025 PASS.
4. 하위 SPEC-AGENTIC-FACTORY-IDENTITY-001이 `completed`이고 그 AC 전부 PASS.
5. 운영자 기록: OD-4(중계 경로 위험 수용), OD-6(시나리오 keep-set 승인).
6. 각 AC의 RED-now 칸이 올바른 이유의 `EXPECTED_RED`로 채워져 있음.
7. 독립 감사 무부채 PASS(새 감사 라운드는 운영자 승인 대상, OD-5).

사람이 보는 설정 화면 시험(OQ-18)은 이 정의에 들지 않으며 완료로 세지 않는다.

## §G 시험 이름 점검 (4cdaba5d7e 기준)

**읽는 법 (F2d)**: IDENTITY·귀속에 기대는 기준과 사례는 행동 축과 시험 축을 따로 적는다(G.0, G.1, 하위 §F). 표지는 셋이다.

- `잠금 유지 — REQ-AF-011에 따른 의도된 상태(미충족 아님)`: 계약이 잠금 동안 원하는 동작이다(귀속 fail-closed, 게이트 `producer-not-allowed`). 리뷰어에게 이 표지는 통과도 실패도 아니다. 그 동작의 시험이 생긴 뒤에야 잠금이 계약대로 버틴다는 증거가 된다. 이 표지는 코드가 그렇게 동작한다는 관측이 아니다. 계약이 그것을 의도하고, 감사가 그렇게 보고했을 뿐이다(아래).
- `미충족 — 구현 없음`: 잠금이 풀린 뒤에만 있는 동작이다(하위 SPEC의 실제 귀속 조회, `runner.Attributor`·`runner.Facts` 구현). lane-18이 구현 중이다(`6b945fd844` HostFacts, 리더 전달, 이 작성자는 읽지 않음).
- `시험 있음` / `시험 없음`: 시험 축. 지금까지처럼 `git grep`으로 존재만 확인한 값이다. 의도된 잠금 동작이라도 그 시험이 없으면 `시험 없음`으로 남는다 — 동작이 의도된 것이라고 시험 공백이 지워지지 않는다.
- 잠금과 상관없는 사례는 "나누지 않음"으로 두고 기존 표지를 그대로 쓴다.
- 근거 1 — REQ-AF-011(`spec.md` 150행) 원문: "Until the lock is lifted, the production attribution source SHALL be a fixed implementation whose behaviour is selected by a single compile-time constant and by no configuration key, environment variable, file, or flag, and SHALL return `unattributable` with reason `identity-spec-incomplete` for every evidence kind, so that the gate records `producer-not-allowed`"; "the lock SHALL cover every caller including the scenario harness, attribution stubs SHALL be confined to gate unit tests"; "the lock SHALL be lifted only by a single code change that lands together with SPEC-AGENTIC-FACTORY-IDENTITY-001 `completed`, its acceptance criteria green, and the operator's recorded risk acceptance of its relay-path residual risk".
- 근거 2 — 감사 판정 문구(audit verdict wording relayed by the leader, not read by this SPEC's author): 'F2: production runner.guard always rejects identity-unknown and the merge gate is always blocked by AttributionLocked; fail-closed'.
- 분류 출처: leader ruling (relayed via the leader session; lane-19 observation) — 운영에서 병합 게이트가 늘 `AttributionLocked`로 막힌다는 감사 지적은 rc.39 범위에서 REQ-AF-011이 의도한 fail-closed 상태이며 결함이 아니다. 이 판정은 감사 문구의 병합 게이트 쪽만 다룬다. 같은 문구의 앞쪽(`runner.guard`가 identity-unknown을 늘 거절함)은 REQ-AF-011이 말하지 않으므로 이 표들에서 분류하지 않았다.

- 배경: 고정 후보 `4cdaba5d7e`의 독립 코드 감사(판정 FAIL 0.67)가, 이 파일의 명령이 고르는 시험 이름·접두 48개 중 47개가 `internal/` 어디에도 `func <이름>(`으로 없음을 찾았다(audit verdict relayed by the leader). 그러면 §E 1번의 "빈 선택 0" 때문에 어느 기준도 PASS 줄을 낼 수 없다.
- 택한 길: 선택지 (나) — 명령을 실제로 있는 시험만 고르도록 다시 쓰고, manifest 시험은 만들지 않는다. 출처: lane-17 decision under the autonomous-judgment procedure (leader dispatch relayed).
- 방법: 이름마다 `git grep -n -E '^func (<이름>)\(' 4cdaba5d7e -- '*_test.go'`(접두는 `git grep -c`)로 존재를 확인했다. 없는 이름은 같은 동작을 다루는 실제 시험 후보를 찾아 **본문을 읽고**, 기준의 동작을 단언할 때만 연결했다. 연결한 시험의 패키지 경로는 그 시험 파일의 디렉터리와 같다.
- **아무것도 실행하지 않았다.** 존재는 `git grep`으로만 확인했으므로 "시험 있음"은 "시험 통과"가 아니다. `4cdaba5d7e` 뒤에 트리가 움직였을 수 있다.
- "일부"는 연결한 시험이 그 기준 사례의 일부만 단언한다는 뜻이다. 각 기준 본문의 "미충족" 줄이 덮이지 않은 사례이며, 그것은 PASS가 아니고 PASS로 고를 수 없다.

| AC | 명령이 고르는 시험(존재 확인) | 상태 | 근거(명령·출력 한 줄) |
|---|---|---|---|
| 001 | `claim` 1, `e2e` 1 = 2 | 시험 있음(일부) | `git grep -n -E '^func (TestAcceptanceManifest\|TestClaim_DeadOwnerExpiredAttempt_ReissuedUnderSameClaim\|…\|TestConcurrentAcquireExactlyOneWins\|TestReal_DuplicateDispatch_NotReclaimed)\(' …` → `…claim/claim_test.go:74:func TestConcurrentAcquireExactlyOneWins(t *testing.T) {` (옛 이름 4개는 출력 없음) |
| 002 | `claim` 4, `e2e` `TestCheckpoint_` 12 = 16 | 시험 있음(일부) | `git grep -c -E '^func (TestAuthority_\|TestCheckpoint_)' … -- 'internal/agenticfactory/*_test.go'` → `…e2e/real_checkpoint_test.go:12` |
| 003 | `e2e` 6 | 시험 있음(일부) | `git grep -c -E '^func TestIssueAuthority_' …` → 출력 없음; 연결 시험 첫 줄 `…e2e/real_authz_test.go:177:func TestAuthz_UnknownIdentity_BlocksEveryOperation(t *testing.T) {` |
| 004 | `claim` 3 | 시험 있음(일부) | `…claim/lock_test.go:63:func TestLiveOwnerExpiredLockStaysBusy(t *testing.T) {` |
| 005 | `effect` 2 | 시험 있음(일부) | `…effect/intent_test.go:78:func TestIntentWriteFailureMeansNoSend(t *testing.T) {` |
| 006 | 없음 | 미충족-시험 없음 | `git grep -c -E '^func (TestLiveness_\|TestDelivery_\|TestDone_\|TestGate_\|TestPRQueue_)' …` → 출력 없음(exit 1) |
| 007 | 없음 | 미충족-시험 없음 | 위와 같은 명령 → 출력 없음 |
| 008 | `e2e` `TestProtectionGuard_` 3 | 시험 있음(일부) | `…e2e/protection_guard_test.go:3` |
| 009 | 없음 | 미충족-시험 없음 (사례 (f)의 두 축: G.0) | `git grep -c -E '^func (TestProtectionGuard_RefusesEveryWeakeningOperation\|TestHandoff_\|TestRepair_\|TestResponsibility_\|TestIndependence_\|…\|TestLock_)' … -- 'internal/agenticfactory/*_test.go'` → 출력 없음(exit 1) |
| 010 | 없음 | 미충족-시험 없음 (사례 (c)(d)(e)의 두 축: G.0) | 위와 같은 명령 → 출력 없음 |
| 011 | `e2e` 1 | 시험 축: 시험 있음(일부). 행동 축: 잠긴 형태(L1·L2·L3·L4·P4a·P4b)는 잠금 유지 — REQ-AF-011에 따른 의도된 상태(미충족 아님), 해제 뒤 형태(L1·L2)는 미충족 — 구현 없음. 사례별: G.0 | `…e2e/real_role_attribution_test.go:55:func TestRoleLock_NoAttributor_BlocksACompletePassSet(t *testing.T) {` |
| 012 | `e2e` `TestIntake_` 12 | 시험 있음(일부) | `…e2e/real_intake_test.go:12` |
| 013 | `core` 2 | 시험 있음(일부) | `…core/core_test.go:166:func TestRedactMasksSecretsAndCounts(t *testing.T) {` |
| 014 | `forge` 2 | 시험 있음(일부) | `…forge/forge_test.go:42:func TestClassifyTable(t *testing.T) {` |
| 015 | `review` 1 | 시험 있음(일부) | `…review/call_test.go:65:func TestImportGraph_NoCodexSurfaceNoGLM(t *testing.T) {` |
| 016 | `core` 2 | 시험 있음(일부) | `…core/stage_model_test.go:94:func TestEdgeTableIsExactlyE01ToE27(t *testing.T) {` |
| 017 | `e2e` `TestCollector_` 6 | 시험 있음(일부) | `…e2e/real_collector_test.go:6` |
| 018 | 없음 | 미충족-시험 없음 | `git grep -c -E '^func (TestRedact_\|TestLedger_\|…\|TestScenarioSuite)' … -- 'internal/agenticfactory/*_test.go' 'internal/config/*_test.go' 'internal/cli/*_test.go'` → 출력 없음(exit 1) |
| 019 | `e2e` 1 | 시험 있음(일부) | `…e2e/real_observation_test.go:156:func TestObservation_Reading_ChangesNothing(t *testing.T) {` |
| 020 | `browser` 1 | 시험 있음(일부) | `…browser/resolve_test.go:177:func TestChoices_ExactlyThree(t *testing.T) {` |
| 021 | 없음 | 미충족-시험 없음 | 018 행과 같은 명령 → 출력 없음 |
| 022 | 없음 | 미충족-시험 없음 | manifest 실행기 없음(001 행의 `TestAcceptanceManifest` 출력 없음) |
| 023 | 없음 | 미충족-시험 없음 | 018 행과 같은 명령 → 출력 없음 |
| 024 | 없음 | 미충족-시험 없음 | 018 행과 같은 명령 → 출력 없음 |
| 025 | 없음(§A.1 명령은 실행 대상 아님) | 미충족-시험 없음 (진단 실행 05·06과 채택 실행의 두 축: G.0) | 018 행과 같은 명령(`TestScenarioSuite` 포함) → 출력 없음 |

셈: 시험이 있는 기준 15개(전부 "일부"), 시험이 없는 기준 10개. 고른 시험은 모두 60개다(2+16+6+3+2+3+1+12+2+2+1+2+6+1+1). 이 수는 존재를 센 것이며 통과 수가 아니다.

### G.0 IDENTITY·귀속에 기대는 기준 — 행동 축과 시험 축

대상은 본문이 귀속·역할 증거를 말하는 기준이다(이 파일에서 `귀속`·`producer-not-allowed`·`identity-spec-incomplete`를 찾아 고름). 시험 축은 위 표의 값을 옮겼고 바꾸지 않았다.

| 기준·사례 | 행동 축 | 시험 축 | 근거 |
|---|---|---|---|
| 011 L1 제품 귀속 원천, 증거 종류 6개 | 잠금 유지 — REQ-AF-011에 따른 의도된 상태(미충족 아님) | 시험 있음(일부: `TestRoleLock_NoAttributor_BlocksACompletePassSet`가 관문 문장의 `identity-spec-incomplete`까지 봄). 종류별 6사례는 시험 없음 | REQ-AF-011 "SHALL return `unattributable` with reason `identity-spec-incomplete` for every evidence kind"; 감사 문구(리더 전달) |
| 011 L2 게이트 기록 | 잠금 유지 — REQ-AF-011에 따른 의도된 상태(미충족 아님) | 시험 있음(일부, 같은 시험). 종류마다 하나씩은 시험 없음 | REQ-AF-011 "so that the gate records `producer-not-allowed`"; 감사 문구 "the merge gate is always blocked by AttributionLocked"(리더 전달) |
| 011 L3 설정·환경 변수로 풀리지 않음 | 잠금 유지 — REQ-AF-011에 따른 의도된 상태(미충족 아님) | 시험 없음 | REQ-AF-011 "selected by a single compile-time constant and by no configuration key, environment variable, file, or flag" |
| 011 L4 하네스의 귀속 대역 거절 | 잠금 유지 — REQ-AF-011에 따른 의도된 상태(미충족 아님). 해제 뒤에도 같음 | 시험 없음 | REQ-AF-011 "the lock SHALL cover every caller including the scenario harness, attribution stubs SHALL be confined to gate unit tests" |
| 011 P4a 다른 세션의 writer 인계 | 잠금 유지 — REQ-AF-011에 따른 의도된 상태(미충족 아님). AC 본문상 해제 뒤에도 같은 거절 | 시험 없음 | AC 본문 P4a "(잠금 중에도, 해제 뒤에도)" |
| 011 P4b 옛 epoch, 귀속 조회 전 거절 | 잠금 유지 — REQ-AF-011에 따른 의도된 상태(미충족 아님). AC 본문상 해제 뒤 그대로 | 시험 없음 | AC 본문 "L3·L4·P4a·P4b는 그대로" |
| 011 해제 뒤 형태 L1·L2(`attributed{필요한 역할}`, "통과로 셈") | 미충족 — 구현 없음 | 시험 없음 | REQ-AF-011 "the lock SHALL be lifted only by a single code change that lands together with SPEC-AGENTIC-FACTORY-IDENTITY-001 `completed`, its acceptance criteria green, and the operator's recorded risk acceptance"; lane-18 구현 중(`6b945fd844`, 리더 전달, 읽지 않음) |
| 025 진단 실행(M7 전)의 05·06 | 잠금 유지 — REQ-AF-011에 따른 의도된 상태(미충족 아님): 05·06이 `producer-not-allowed`로 막힘 | 시험 없음 | §A.1 "M7 전의 실행은 진단 실행이며 05·06에서 `producer-not-allowed`로 막히는 것이 정상"; REQ-AF-011 "including the scenario harness" |
| 025 채택 실행(M7 뒤) 01~07 | 미충족 — 구현 없음 | 시험 없음 | §A.1 "채택 실행은 M7(잠금 해제) 뒤의 실행뿐" |
| 009 (f) CI 실패 수리부터 DONE — 잠금 중 운영 경로 | 잠금 유지 — REQ-AF-011에 따른 의도된 상태(미충족 아님): 리뷰 증거가 `producer-not-allowed`라 병합 단계에 닿지 않음 | 시험 없음 | REQ-AF-011 "for every evidence kind" |
| 009 (f) 리뷰 PASS → 병합 → DONE(해제 뒤) | 미충족 — 구현 없음 | 시험 없음 | 같은 REQ의 해제 조건 |
| 009 (a)~(e)·(g) | 나누지 않음(귀속에 기대지 않음) | 시험 없음 | — |
| 010 (c)(d)(e) 리뷰 증거 산입 — 잠금 중 운영 경로 | 잠금 유지 — REQ-AF-011에 따른 의도된 상태(미충족 아님): 리뷰 증거는 모두 `producer-not-allowed` | 시험 없음 | REQ-AF-011 "for every evidence kind" |
| 010 (c)(d)(e) 실제 생산자로 세고 빼는 형태(해제 뒤) | 미충족 — 구현 없음 | 시험 없음 | 같은 REQ의 해제 조건 |
| 010 (c)(d)(e) 게이트 단위 시험(귀속 대역) 형태 | 나누지 않음(대역은 게이트 단위 시험에서 허용되므로 잠금이 정하지 않음) | 시험 없음 | REQ-AF-011 "attribution stubs SHALL be confined to gate unit tests"; AC 본문 "(e)는 귀속 대역으로 게이트 단위 시험 안에서만 돈다" |
| 010 (a)(b)(f) | 나누지 않음 | 시험 없음 | — |
| 007 게이트 표 8행(귀속 대역 허용 단위 시험) | 나누지 않음(같은 이유) | 시험 없음 | AC 본문 "게이트 단위 시험, 귀속 대역 허용" |

`잠금 유지` 칸은 계약이 지금 원하는 상태라는 분류다. 그 동작을 코드에서 관측한 것은 감사 보고(리더 전달)뿐이며, 그 보고는 병합 게이트 쪽만 말한다. L3·L4·P4a·P4b와 025 진단 실행의 코드 동작은 이 작성자가 관측하지 않았다.

### G.1 하위 SPEC-AGENTIC-FACTORY-IDENTITY-001 [REF] 명령

F2 작업 당시에는 하위 SPEC의 `acceptance.md`를 고치지 않고 여기에 적었다. 그 뒤 F2b에서 하위 `acceptance.md` §F에 같은 점검과 원장을 적었다. 하위 SPEC의 아홉 기준 명령은 모두 `./internal/agenticfactory/identity/` 패키지의 `TestIdentity_*` 시험을 고른다.

- `git grep -c -E '^func TestIdentity_' 4cdaba5d7e -- '*_test.go'` → 출력 없음(exit 1).
- `git ls-tree --name-only 4cdaba5d7e internal/agenticfactory/identity` → 출력 없음.
- 그래서 아홉 명령(`TestIdentity_GrantAuthority_ControlRefOnly`, `TestIdentity_OperatorAuth_ServerObservedOnly`, `TestIdentity_Ancestry_RegisteredDescendantRefused`, `TestIdentity_Ancestry_OperatorAnchorAndOrphans`, `TestIdentity_Ancestry_IndeterminateAndBodyPid`, `TestIdentity_SessionBinding_ReobservedOnEveryRegistration`, `TestIdentity_RoleExclusion_PerIssueWriter`, `TestIdentity_AttributionMatrix_Generated`·`_ExtraConditions`, `TestIdentity_AttributionQuery_ReadOnlyClosedReasons`)은 지금 아무것도 고르지 않으며, 아홉 기준 모두 미충족이다. 귀속 조회(AC-AFI-008 [REF], AC-AFI-009 [REF])는 미충족-구현 없음이다: 감사 보고상 `runner.Attributor`·`runner.Facts` 구현이 0개다(리더 전달. 이 작성자는 그 밖을 읽지 않았다). 나머지 일곱은 대상 패키지 경로가 후보에 없다는 사실만 확인했다. (F2b에서 이 작성자가 `runner/runner.go`를 읽었다. 두 인터페이스는 선언만 있고 메서드 구현은 `_test.go`에만 있다 — 하위 §F 원장 L13·L14.)

두 축(F2d). 아홉 기준은 모두 하위 SPEC 자신의 동작이며 상위 잠금의 동작이 아니다. 잠금 상태(`identity-spec-incomplete`)는 상위 REQ-AF-011과 AC-AF-011이 가진다(G.0). 하위 AC-AFI-009 [REF]의 Q4도 그 값을 하위 구현이 스스로 내지 않는다는 단언이다.

| 하위 기준 | 행동 축 | 시험 축 |
|---|---|---|
| AC-AFI-001 [REF] 부여 권한 | 미충족 — 구현 없음 | 시험 없음 |
| AC-AFI-002 [REF] 운영자 인증 | 미충족 — 구현 없음 | 시험 없음 |
| AC-AFI-003 [REF] 등록 세션 자손 거절 | 미충족 — 구현 없음 | 시험 없음 |
| AC-AFI-004 [REF] 기준점·고아 | 미충족 — 구현 없음 | 시험 없음 |
| AC-AFI-005 [REF] 판정 불가·본문 pid | 미충족 — 구현 없음 | 시험 없음 |
| AC-AFI-006 [REF] 세션 결속 | 미충족 — 구현 없음 | 시험 없음 |
| AC-AFI-007 [REF] 역할 배제 | 미충족 — 구현 없음 | 시험 없음 |
| AC-AFI-008 [REF] 귀속 행렬 | 미충족 — 구현 없음(`runner.Attributor` 테스트 밖 구현 없음) | 시험 없음 |
| AC-AFI-009 [REF] 읽기 전용 귀속 조회 | 미충족 — 구현 없음(같음) | 시험 없음 |

근거: 하위 `acceptance.md` §F(F2b 원장 L01~L14), lane-18 구현 중(`6b945fd844` HostFacts, 리더 전달, 읽지 않음).

### G.2 아직 없는 시험 이름 — 담당 lane별

- **배정**: lane-17이 리더 판정(relayed)에 따라 정했다. 기준은 각 이름을 고르는 AC의 `maps REQ-…` 줄과, 고정 후보 `4cdaba5d7e`의 패키지 배치(`git ls-tree -d --name-only 4cdaba5d7e internal/agenticfactory/`와 형제 시험 파일 이름)다.
- **lane 표(리더 판정)**: lane-17 = 계약·설정(`internal/config`, SPEC). lane-18 = core, claim, runner, mirror, checkpoint, ctxmodel, gitref. lane-20 = effect(effect/* 포함), monitor(monitor/* 포함), forge(forge/webhook 포함), resident, reconcile. lane-21 = web, e2e, browser, intake(intake는 추론). lane-22 = `internal/cli`와 통합.
- **패키지 칸은 제안이다.** lane은 다른 곳에 둘 수 있다. 경로는 `internal/agenticfactory/` 아래이며, 그 밖이면 전체 경로를 적었다.
- 두 lane에 걸친 이름은 주된 동작의 lane에 두고, 다른 lane을 "zusätzlich"로 적었다.
- "동작" 칸은 AC 본문 사례 문장에서 옮겼다. AC가 사례 없이 이름만 적은 경우에는 "사례 문장 없음"이라고 적고, 동작을 지어내지 않았다.
- 이 목록은 사실로서 각 lane에 보낸다.
- 배정 불가: 없음. 모든 이름의 REQ에 놓을 동작이 있다.

#### lane-18

| 시험 이름 | AC | REQ | 패키지(제안) | 동작 |
|---|---|---|---|---|
| `TestClaim_DeadOwnerExpiredAttempt_ReissuedUnderSameClaim` | AC-AF-001 | REQ-AF-001 | `claim/`(`claim_test.go` 옆) | 사례 문장 없음(§12 재현 시험 이름만 적힘) |
| `TestClaim_WorktreeIdentity_NormalizedPathPreventsMismatch` | AC-AF-001 | REQ-AF-001 | `claim/` | 사례 문장 없음(§12 재현 시험 이름만 적힘) |
| `TestClaim_SameNameTree_NoDivergentTreeCreated` | AC-AF-001 | REQ-AF-001 | `claim/` | 사례 문장 없음(§12 재현 시험 이름만 적힘) |
| `TestAuthority_*` | AC-AF-002 | REQ-AF-002 | `claim/`(`controlref_test.go` 옆) | A4: control ref에 없는 claim이 로컬 캐시에만 있음 → claim으로 인정하지 않음 |
| `TestIssueAuthority_*` | AC-AF-003 | REQ-AF-003 | `core/`(`authz_test.go` 옆) | IA1, IA2, IA5~IA8, IA11, IA13, 정적 단언(`internal/config`에 사유 문자열 0건, `ForeignIssueAllows` 0건). zusätzlich lane-20: IA14 PR effect 전송 0, IA16 새 Issue 0 |
| `TestLiveness_*` | AC-AF-004 | REQ-AF-004 | `runner/`(`takeover_test.go` 옆) | (d) 재시작·끊김 뒤 실제 실행 조회가 재개·재발급·PARKED 결정보다 먼저, 강제 종료 0, 중복 attempt 0 |
| `TestLiveness_CrashHandover_ResponsibilityPassesWithEpoch` | AC-AF-004 | REQ-AF-004 | `runner/` | (e) 책임 agent `dead` → 새 agent에 epoch+1로 책임, 책임 기록이 비는 순번 0, 이전 worktree·branch·변경 보존, `live`·`indeterminate`면 이전 0 |
| `TestDone_*` | AC-AF-006 | REQ-AF-006 | `core/`(`closing_test.go` 옆) | (a) 종료 effect 순번 < DONE 순번 (b) 종료 실패 → MERGING + PARKED `pending-effect` (c) 로컬 동기화 변형 → DONE 거절 (d) `deploy: N/A` 사유 없음 → 거절 (e) DONE 전 다음 dispatch → 거절. zusätzlich lane-20: Issue 종료 effect |
| `TestGate_ReasonTable_*` | AC-AF-007 | REQ-AF-007 | `core/`(`gate_test.go` 옆) | design §2.1b 8행마다 기대 사유 하나, 기록 수 1. 검사 미통과 → `check-not-passed`, 감사 예산 소진 → `budget-exhausted` |
| `TestPRQueue_*` | AC-AF-007 | REQ-AF-007 | `runner/`(`next_test.go` 옆) | 적체 (a) 열린 인계 + 수리 적체 = `wip_limit` → 새 dispatch 0, 종결·수리 effect는 계속. zusätzlich lane-17: 적체 (b) 설정 로드 거절 |
| `TestHandoff_*`, `TestRepair_*`, `TestResponsibility_*` | AC-AF-009 | REQ-AF-009 | `runner/`(`handoff_test.go`·`repair_test.go` 옆) | (a)~(e): receipt 필드 하나씩 빼기 → E03 거절, PR 발행만 → DONE·REVIEWING 아님, E03 뒤 책임 유지·보존, E06 수리 배타 claim, MERGING 중 HEAD 변경 → E10 |
| `TestResponsibility_CIFailure_SameAgentRepairsThroughDone` | AC-AF-009 | REQ-AF-009 | `runner/` | (f) 실패한 필수 검사 → 같은 agent가 수리부터 DONE까지, leader 커밋 0, 단계 순번 |
| `TestResponsibility_SecondIssue_OnlyWhenConditionsHold` | AC-AF-009 | REQ-AF-009 | `runner/` | (g) ①~⑤ 두 번째 Issue 배차 조건 |
| `TestIndependence_*` | AC-AF-010 | REQ-AF-010 | `core/`(`authz_test.go`·`gate_test.go` 옆) | (a)~(e): ControlService 밖 쓰기 거절, Coordinator 직접 상태 쓰기 거절, 같은 Issue writer의 리뷰 불산입, Coordinator 구현의 리뷰 불산입, 다른 Issue writer의 리뷰 산입 |
| `TestIndependence_LeaderDoesNotTakeOverAgentWork` | AC-AF-010 | REQ-AF-010 | `core/` | (f) Coordinator의 구현·수리·리뷰 대응 쓰기 요청 → 거절, attempt 기록 불변 |
| `TestGate_RoleEvidence_RequiresChildAttribution_AndLockHolds` | AC-AF-011 | REQ-AF-011 | `runner/`(`attribution_test.go` 옆) | L1 증거 종류 6개 각각 `unattributable{identity-spec-incomplete}`, L2 종류마다 `producer-not-allowed` 하나 |
| `TestLock_*` | AC-AF-011 | REQ-AF-011 | `runner/` | L3 설정·환경 변수 변경에도 결과 바이트 동일 + 정적 검사, P4a, P4b. zusätzlich lane-21: L4 시나리오 하네스 사전 점검 |
| `TestRedact_*`(13) | AC-AF-013 | REQ-AF-013 | `core/`(`core_test.go` 옆) | C1 3, C2 1, C3 2, C4 2, C5 1, Restricted 1, Idempotent 1, Negative 2(범주 하나를 뺀 변형에서 FAIL). zusätzlich lane-20: AC 초록 경로의 refiner(M5) |
| `TestLedger_FieldOwner_*` | AC-AF-013 | REQ-AF-013 | `ctxmodel/` | 필드 owner가 아닌 역할의 장부 쓰기 → 거절 |
| `TestTransition_Table_*` | AC-AF-016 | REQ-AF-016 | `core/`(`stage_model_test.go` 옆) | (b) 요청자 불일치 → 거절 (c) PARKED 사유 집합 밖 → 거절 (d) E20 → 단계·owner 불변 |
| `TestTransition_Table_WaitRecordRequiresAllSixFields` | AC-AF-016 | REQ-AF-016 | `core/`(`wait_test.go` 옆) | (e) 필드 하나씩 뺀 대기 기록 여섯 → 각각 거절, 여섯 필드 기록 → 받아들임 |
| `TestState_SpeclessIssue_TransitionsWithoutSpecHash` | AC-AF-016 | REQ-AF-016 | `core/` | 사례 문장 없음(§12 이름만 적힘) |
| `TestState_OperatorDecision_RequiresOperatorEvidence_CoordinatorRefused` | AC-AF-016 | REQ-AF-016 | `core/` | 사례 문장 없음(§12 이름만 적힘) |

#### lane-20

| 시험 이름 | AC | REQ | 패키지(제안) | 동작 |
|---|---|---|---|---|
| `TestDelivery_*` | AC-AF-005 | REQ-AF-005 | `effect/`(`intent_test.go` 옆) | (a) 전송 성공만 → `sent`만, `received` 이상 0 (b) idle 알림 → `completed` 아님 |
| `TestDelivery_ReplyLoss_RecoveredFromRecordNotReply` | AC-AF-005 | REQ-AF-005 | `effect/` | (e) 완료 답장 유실 → 기록으로 복구, 재전송 0, 대기 기록 여섯 필드. zusätzlich lane-18: runner의 진행 복구 |
| `TestProtectionGuard_RefusesEveryWeakeningOperation` | AC-AF-008 | REQ-AF-008 | `effect/`(같은 접두 시험 3개는 `e2e/`) | design §11.3 W1~W7 각각 → 전송 0, 상태 불변, `permission-denied` 사유 1개. zusätzlich lane-18: 사유 타입 |
| `TestForgeClassify_*` | AC-AF-014 | REQ-AF-014 | `forge/`(`forge_test.go` 옆) | design §4.5 G1~G8·L1~L6 각각 → 표의 분류와 재시도 수 |
| `TestBudget_*` | AC-AF-014 | REQ-AF-014 | `forge/`(`budget_test.go` 옆) | 예산 (a) 재개 시각 전 요청 0 (b) 헤더 없음 → `UNKNOWN` (c) 설정 기본값 변경에 대기 동작이 따라감 (d) 예비 몫 아래 우선순위, 다른 자격·endpoint 전환 0 |
| `TestProvider_*` | AC-AF-015 | REQ-AF-015 | `review/`(`call_test.go` 옆; review는 lane 표에 없음) | (a) `codex.enabled` 미설정 → Codex 호출 0 (b) `false` → 0 (c) `true` → 계수 1 이상 (d) GLM 호출 0. zusätzlich lane-17: `codex.enabled` 키 |
| `TestLedger_*`, `TestEvidenceStage_*` | AC-AF-018 | REQ-AF-019 | `effect/ctxledger/` | (a)~(g): `inaccessible-local`, `inaccessible-unpushed`, HEAD 변경 → `STALE`, blob 변경 → `STALE`, 조회 실패 → `UNKNOWN`, `PUSHED`에서 조회 전 `AVAILABLE` 0, branch 이름 URL 거절 |
| `TestObservationView_ShowsPrepAndPRBacklogStatesDistinctly_ReadOnly` | AC-AF-019 | REQ-AF-020 | `monitor/`(`view_ext_test.go` 옆) | (a) 항목마다 값·원천·읽은 시각 (b) 준비 7행·PR 적체를 BUILDING 이후와 다른 구역에 (c) 순번·트리 키 동일, 쓰기 0 (d) C1~C5 원값 0 |

#### lane-21

| 시험 이름 | AC | REQ | 패키지(제안) | 동작 |
|---|---|---|---|---|
| `TestScenarioSuite`(하위 `01_…`~`07_…`) | AC-AF-025 | REQ-AF-006(보조 REQ-AF-001·004·005·011) | `e2e/` | §A.1 하위 시험 01~07, 실제 Issue 한 쌍, 한 번의 실행 |
| `TestBrowserSelect_*` | AC-AF-020 | REQ-AF-021 | `browser/`(`resolve_test.go` 옆) | (d) 미선택 → `chrome` (e) `aside` 선택·미설치 → PARKED `not-installed`, `chrome` 호출 0 (f) 선택 저장 뒤 다시 읽기 → 같은 값. zusätzlich lane-17: (f)의 저장 |

#### lane-17

| 시험 이름 | AC | REQ | 패키지(제안) | 동작 |
|---|---|---|---|---|
| `TestAgenticFactoryConfig_*` | AC-AF-020 | REQ-AF-021 | `internal/config/`(`agentic_factory_test.go` 옆) | (a) 범위 밖 값 → 키·범위·파일을 밝혀 거절, 깎기 0 (b) 읽기 오류 → 미설정 처리 0 |

#### lane-22

| 시험 이름 | AC | REQ | 패키지(제안) | 동작 |
|---|---|---|---|---|
| `TestCompat_*`, `TestLauncher_*` | AC-AF-021 | REQ-AF-022 | `internal/cli/`(`launch_exec_test.go`·`gtd_compat_test.go` 옆) | (a)~(g): golden 바이트 동일, stderr 한 줄, 버전 선언 없는 새 문법 → 종료 2, 혼합 → 호환 오류, 역할 뒤바뀜 0, 모르는 인자 거절, `--` 뒤 전달. golden 순서 점검 |
| `TestEnvMatrix_StepEvidenceRequired_NotRunNeverSuccess` | AC-AF-023 | REQ-AF-024 | 기존 패키지 없음(새 위치) | (a) 명령·종료 코드·출력 중 하나 없음 → 행 거절 (b) 실행 안 한 단계 → `not-run` (c) 전부 `BLOCKED` → 행렬 FAIL (d) Codex 미설정 → `BLOCKED(not-configured)` |
| `TestEfficiencyRecord_*` | AC-AF-024 | REQ-AF-025 | 기존 패키지 없음(새 위치) | (a) 필드 누락 → 거절 (b) `estimated`를 `measured` 칸에 → 거절 (c) 품질 회귀 비교 → 향상 보고 0 (d) checkpoint 비용 필드는 측정값 또는 Gap. zusätzlich lane-18: (d) checkpoint |
| `TestAcceptanceManifest`(실행기) | 전부 | — | 기존 패키지 없음(새 위치) | design §15.1 정본 id를 하위 시험으로 돈다(§A). 각 id의 동작은 AC 본문에 따로 적혀 있지 않다. zusätzlich: 각 id는 그 AC의 lane |

`TestAcceptanceManifest`의 하위 시험(정본 id, 각 AC의 "정본 주 시험" 줄):
- AC-AF-001: M01~M05
- AC-AF-002: SYNC12, M47
- AC-AF-003: M13, M15
- AC-AF-004: M06~M12
- AC-AF-005: M18~M20, M29~M31, SYNC03~SYNC05, SYNC11
- AC-AF-006: M33~M36
- AC-AF-007: M14, M24~M27, M32, M37, M55
- AC-AF-009: HANDOFF01~HANDOFF08
- AC-AF-010: M22, M23
- AC-AF-012: PREP01~PREP10
- AC-AF-013: M39, M40, CONTEXT04
- AC-AF-014: M17, M28, M43, M44, GH05, GH06, SYNC09, SYNC10, WT15, SYNC13~SYNC18
- AC-AF-015: M21
- AC-AF-016: M16, M41, M42, M48, M54, M56
- AC-AF-017: M45, M49~M53, SYNC01, SYNC02, SYNC06~SYNC08
- AC-AF-018: CONTEXT01~03, CONTEXT05, CONTEXT06
- AC-AF-020: M38, M46, M46a~M46g
- AC-AF-021: CLI01~CLI23
- AC-AF-022: WT01~WT14, WT16, GH01~GH04, GH07~GH15
- AC-AF-024: EFF01~EFF06

#### IDENTITY

하위 SPEC-AGENTIC-FACTORY-IDENTITY-001 [REF]의 항목은 lane-18이 구현한다(커밋 `6b945fd844` HostFacts. relayed, not read). 하위 SPEC의 `acceptance.md`에는 자체 목록이 아직 필요하며, 다음 작업에서 만든다. G.1의 `TestIdentity_*` 아홉은 이 표에 넣지 않았다.
