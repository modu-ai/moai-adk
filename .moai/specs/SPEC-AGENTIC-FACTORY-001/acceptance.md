# SPEC-AGENTIC-FACTORY-001 — 수용 기준 (0.6.4)

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
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(M01|M02|M03|M04|M05)$|^(TestClaim_DeadOwnerExpiredAttempt_ReissuedUnderSameClaim|TestClaim_WorktreeIdentity_NormalizedPathPreventsMismatch|TestClaim_SameNameTree_NoDivergentTreeCreated)$' -count=1 -v` → PASS 줄 8.
- 사례: (a) 같은 Issue에 동시 claim 둘 → 정확히 하나 성공, 다른 하나 거절(종료 10)과 충돌한 claim 기록. (b) 재발급 → 같은 Issue 키 아래 새 attempt, 키 수 1. (c) 옛 epoch 쓰기 → 거절(종료 15), control ref 불변. (d) 살아 있는 attempt가 있는 Issue로 dispatch → 거절. (e) 미전달 dispatch가 있는 Issue로 dispatch → 거절.
- 변형을 죽이는 단언: 두 claim이 모두 성공하는 구현, 재발급이 새 키를 만드는 구현, epoch를 비교하지 않는 구현, 충돌 기록을 남기지 않는 구현은 각각 (a)(b)(c)(a)에서 FAIL.
- RED 이유: claim 저장 구현이 없어 (a)에서 단언 실패. 초록 경로: M1 lane-18 claim·fencing.

#### AC-AF-002 — 권위·재구축·checkpoint 복구

- 주 REQ: maps REQ-AF-002. 초록: M2.
- 정본 주 시험 2: SYNC12, M47. 전용(제안): `TestCheckpoint_*`, `TestAuthority_*`.
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(SYNC12|M47)$|^(TestAuthority_.*|TestCheckpoint_.*)$' -count=1 -v` → 아래 사례마다 PASS 줄 하나(SYNC12·M47 2 + A1~A4 4 + CK1~CK10 10 = 16).
- 권위 사례: A1 control ref 읽기 실패 → 의존 전이 없음, `store-unavailable`. A2 CAS 실패 → 전이 없음. A3 로컬 캐시 삭제 → control ref에서 재구축한 상태가 삭제 전과 같음. A4 control ref에 없는 claim이 로컬 캐시에만 있음 → claim으로 인정하지 않음.
- checkpoint 사례(design §4.6): CK1 K1 뒤 ACK 전 충돌 → `pending`(`committed-unacked`), `remote-durable` 아님. CK2 ACK는 있으나 manifest readback 불일치 → `acked-manifest-unverified`, `recovered` 불가. CK3 K4 Issue 링크 실패 → 여전히 `remote-durable`, 링크 재시도가 예약됨. CK4 같은 ref·SHA로 두 번째 push 요청 → push 호출 0, 거절. CK5 이전 writer가 `live`인데 복구 요청 → 거절(`ownership-unverified`), 새 writer 0. CK6 사용자의 미커밋 변경이 있는 트리에서 checkpoint → 그 경로의 내용 digest·인덱스 항목 불변, 커밋에 포함 0. CK7 push 대상이 기본 branch → 거절. CK8 non-fast-forward → 강제 push 0, PARKED. CK9 manifest 후보에 C1 값 → 기록 거절 또는 `[REDACTED:C1]`, 원값 검색 0. CK10 상한 초과 → 가장 오래된 제거 가능 항목만 색인에서 빠지고, 제거 가능 항목이 없으면 `capacity-short`로 새 checkpoint 거절, 기존 항목 불변.
- 변형 단언: ACK 없이 `remote-durable`을 세우는 구현은 CK1, manifest를 검증하지 않는 구현은 CK2, 링크 실패로 durable을 취소하는 구현은 CK3, 중복 검사 없는 구현은 CK4(push 호출 계수 1), 활성 판정 없는 복구는 CK5, `git add -A`식 스테이징은 CK6, 상한 무시는 CK10에서 FAIL.
- RED 이유: control ref git 구현과 checkpoint 상태 기계가 없음. 초록 경로: M1 control ref(lane-18), M2 checkpoint·Issue 링크(lane-18·20).

#### AC-AF-003 — 소유 확인과 Issue 쓰기 권한

- 주 REQ: maps REQ-AF-003. 초록: M2.
- 정본 주 시험 2: M13, M15. 전용(제안): `TestIssueAuthority_*`.
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(M13|M15)$|^(TestIssueAuthority_.*)$' -count=1 -v` → PASS 줄 18(M13·M15 + IA1~IA16).
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
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(M06|M07|M08|M09|M10|M11|M12)$|^(TestLiveness_.*)$' -count=1 -v` → PASS 줄 = 7 + 사례 4.
- 사례: (a) lease 만료 + 판정 `live` → 인계 거절, attempt `lapsed`. (b) `indeterminate` → 같음. (c) 같은 pid·다른 fingerprint → `dead` → epoch+1, 이전 worktree·branch·SHA·dirty 목록 보존·참조. (d) 재시작·끊김 뒤 복구 → 실제 실행 조회가 재개·재발급·PARKED 결정보다 먼저(호출 순서 기록), 강제 종료 0, 중복 attempt 0.
- 변형 단언: PID만 보는 판정은 (c)에서 FAIL(재사용된 pid를 `live`로 봄). TTL만으로 인계하는 구현은 (a)에서 FAIL.
- RED 이유: 판정 연결 없음. 초록 경로: M1(lane-18, `homestate.ClassifyOwnerWith` 사용).

#### AC-AF-005 — effect 불명 재조회·전달 상태

- 주 REQ: maps REQ-AF-005. 초록: M2.
- 정본 주 시험 10: M18, M19, M20, M29, M30, M31, SYNC03, SYNC04, SYNC05, SYNC11.
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(M18|M19|M20|M29|M30|M31|SYNC03|SYNC04|SYNC05|SYNC11)$|^(TestDelivery_.*)$' -count=1 -v` → PASS 줄 = 10 + 사례 4.
- 사례: (a) 전송 성공만 있음 → `sent`만, `received` 이상 0. (b) idle 알림 → `completed` 아님. (c) effect 응답 불명 → 재발행 전 원격 조회 1회 이상, 재발행 0, PARKED `unknown-effect`. (d) intent 기록 전에 전송을 시도하는 경로 → 거절(intent 없는 전송 0).
- 변형 단언: 답장으로 `completed`를 세우는 구현은 (b), 불명 시 재전송하는 구현은 (c)에서 FAIL.
- RED 이유: effect 경계 없음. 초록 경로: M2(lane-20).

#### AC-AF-006 — DONE 순서

- 주 REQ: maps REQ-AF-006. 초록: M3.
- 정본 주 시험 4: M33, M34, M35, M36.
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(M33|M34|M35|M36)$|^(TestDone_.*)$' -count=1 -v` → PASS 줄 = 4 + 사례 5.
- 사례: (a) E12 뒤 Issue 종료 effect의 control ref 순번 < DONE 기록 순번. (b) 종료 실패 → MERGING + PARKED `pending-effect`, revert effect 0. (c) 로컬 동기화 변형: 동기화 전 수정 경로의 내용이 바뀌고 경로 목록은 같음 → DONE 거절. 인덱스만 바뀜, 모드만 바뀜도 각각 거절. (d) `deploy: N/A` 사유 없음 → DONE 거절. (e) DONE 전 같은 자리 다음 dispatch → 거절. 이관 기록은 DONE 0.
- 변형 단언: DONE 뒤 종료하는 구현(0.5.0 순서)은 (a)에서 FAIL. 경로 목록만 비교하는 구현은 (c)에서 FAIL.
- RED 이유: 상태 엔진의 E13·E14 없음. 초록 경로: M3(lane-18). lane-18 기존 시험(이름 미전달)은 보조 근거다.

#### AC-AF-007 — 병합 게이트·PR 적체 상한

- 주 REQ: maps REQ-AF-007. 초록: M4.
- 정본 주 시험 8: M14, M24, M25, M26, M27, M32, M37, M55.
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(M14|M24|M25|M26|M27|M32|M37|M55)$|^(TestGate_ReasonTable_.*|TestPRQueue_.*)$' -count=1 -v` → PASS 줄 = 8 + 표 행 8 + 적체 사례 2.
- 게이트 사례(게이트 단위 시험, 귀속 대역 허용): design §2.1b 표의 8행마다 하나. 각 행의 기대 사유는 표와 같고 한 결과는 사유 하나만 남긴다(기록 수 = 1). 여기에 skipped·누락·실패·취소·다른 HEAD 검사 → `check-not-passed`, 감사 예산 소진 → `budget-exhausted`.
- 적체 사례: (a) 열린 인계 + 수리 적체 = `wip_limit` → 새 Issue dispatch 0, 종결·수리 effect는 계속. (b) 키 값 `0`, `-1`, `"x"`, 범위 밖 정수 → 설정 로드가 키 이름과 허용 범위를 밝혀 거절(`0`을 끔으로 읽는 변형을 죽임). 키 미설정 사례는 두지 않는다. 미설정 동작은 운영자 결정 대기(OD-7)이며, 결정이 나면 manager-spec 재위임으로 사례 (c)를 넣고 기대 PASS 수를 고친다.
- 변형 단언: `PASS-WITH-DEBT`를 통과로 세는 구현, 귀속 실패를 판정값보다 늦게 보는 구현(한 결과에 사유 둘), `operator-ack-missing`을 역할 부재에 쓰는 구현은 표 행에서 FAIL.
- RED 이유: 게이트 없음. 초록 경로: M4(lane-18 판정 + lane-20 effect).

#### AC-AF-008 — 보호 비약화

- 주 REQ: maps REQ-AF-008. 초록: M2.
- 정본 주 시험 0. 전용: lane-22 `TestProtectionGuard_*` 3개 + 약화 거절 시험(제안 `TestProtectionGuard_RefusesEveryWeakeningOperation`).
- 명령: `go test ./internal/agenticfactory/... -run '^TestProtectionGuard_.*$' -count=1 -v` → PASS 줄 4 이상, 거절 시험 안의 하위 시험 W1~W7 각 PASS.
- 사례: design §11.3 W1~W7 각각 → 전송 0, 상태 불변, 사유 기록 정확히 1개(`permission-denied`). 레이블 쓰기 경로는 악의적 이름으로도 레이블 endpoint를 벗어나지 않음. 쓰기 기본값 꺼짐에서 전송 0. 제품 소스에 보호 설정 변경 용어 0.
- regression-guard(release-blocking 아님): EV-AC008-BASE의 다섯 값이 착지 전후 같음.
- 변형 단언: 일곱 중 하나라도 보내는 구현, 사유를 두 개 남기는 구현이 FAIL.
- RED 이유: 거절 경로가 코드에 없음(lane-22 보고). 초록 경로: M2(lane-20 집행, lane-18 사유 타입).

#### AC-AF-009 — PR 인계·수리 단일 소유

- 주 REQ: maps REQ-AF-009. 초록: M3.
- 정본 주 시험 8: HANDOFF01~HANDOFF08.
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(HANDOFF0[1-8])$|^(TestHandoff_.*|TestRepair_.*)$' -count=1 -v` → PASS 줄 = 8 + 사례 5.
- 사례: (a) receipt 필드(design §4.8 PRHandoff) 하나씩 빼기 → E03 거절, 상태 BUILDING 유지. (b) PR 발행만 있음 → DONE도 REVIEWING도 아님. (c) E03 뒤 agent claim·slot 반환, worktree·branch·변경 보존(경로 digest 불변). (d) E06 수리 → 원 Issue·branch에 배타 claim, 같은 worktree·epoch 연속성, 가용 agent에 배차, 그 agent의 다른 Issue 기록 불변. (e) MERGING 중 HEAD 변경 → E10으로 REVIEWING, 이어 E06.
- 변형 단언: 필드 하나 없이 인계를 받는 구현은 (a), 수리를 새 Issue로 여는 구현은 (d), 바쁜 agent의 Issue를 덮는 구현은 (d)에서 FAIL.
- RED 이유: E03·E06·E10 없음. 초록 경로: M3(lane-18).

#### AC-AF-010 — 독립성과 역할 쓰기 경계

- 주 REQ: maps REQ-AF-010. 초록: M4.
- 정본 주 시험 2: M22, M23.
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(M22|M23)$|^(TestIndependence_.*)$' -count=1 -v` → PASS 줄 = 2 + 사례 4.
- 사례: (a) ControlService가 아닌 쓰기 주체의 claim·lease·liveness·resume 쓰기 → 거절. (b) Coordinator의 직접 상태 쓰기 → 거절. (c) 같은 Issue의 어느 attempt writer였던 세션의 리뷰 증거 → 세지 않음. (d) Coordinator가 직접 구현한 변경의 리뷰 → 세지 않음. (e) 다른 Issue의 writer였던 세션이 이 Issue를 리뷰 → 셈(정본 [2]§21).
- 변형 단언: run 전체를 막는 과잉 구현은 (e)에서 FAIL, Issue 범위를 보지 않는 구현은 (c)에서 FAIL.
- RED 이유: 역할 경계 없음. 초록 경로: M4. (e)는 귀속 대역으로 게이트 단위 시험 안에서만 돈다.

#### AC-AF-011 — 역할 귀속과 잠금

- 주 REQ: maps REQ-AF-011. 초록: M4(잠긴 형태). M7에서 L1·L2를 다시 쓴다.
- 정본 주 시험 0. 전용(제안): `TestGate_RoleEvidence_RequiresChildAttribution_AndLockHolds`, `TestLock_*`.
- 명령: `go test ./internal/agenticfactory/... -run '^(TestGate_RoleEvidence_RequiresChildAttribution_AndLockHolds|TestLock_.*)$' -count=1 -v` → PASS 줄 = 사례 6.
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
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(PREP(0[1-9]|10))$|^(TestIntake_.*)$' -count=1 -v` → PASS 줄 = 10 + 사례 5.
- 사례: (a) control ref 기록 전 충돌 → 수락 응답 0. (b) 상태가 design §4.2의 7행 밖으로 이동 요청 → 거절. 6단계 이름만 가진 구현은 `BRANCH_REGISTERED`가 없어 FAIL. (c) READY ACK 전 개발 launch → 0. (d) `intake.max_concurrent`를 N으로 설정했을 때 동시 준비 N+1번째 → 대기. (e) 기한 초과 → 예약 해제. 키 미설정 사례는 두지 않는다. 미설정 동작은 운영자 결정 대기(OD-9)이며, 결정이 나면 manager-spec 재위임으로 사례를 넣고 기대 PASS 수를 고친다.
- 변형 단언: 수락을 먼저 돌려주는 구현은 (a), 무한 하위 작업은 (d)에서 FAIL.
- RED 이유: 접수 서비스 없음. 초록 경로: M3(lane-18 상태, lane-20 Issue 발행).

#### AC-AF-013 — 비밀·공개 검토

- 주 REQ: maps REQ-AF-013. 초록: M5.
- 정본 주 시험 3: M39, M40, CONTEXT04. 전용: `TestRedact_*` 13개(C1 3, C2 1, C3 2, C4 2, C5 1, Restricted 1, Idempotent 1, Negative 2).
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(M39|M40|CONTEXT04)$|^(TestRedact_.*|TestLedger_FieldOwner_.*)$' -count=1 -v` → PASS 줄 = 3 + 13 + 2.
- 사례: 범주별 가림 출력, `RESTRICTED`는 상태·요약·SHA-256 세 필드만, 범주 하나씩 꺼 둔 변형 refiner에서 게시 거절이 다섯 번 모두 일어남, 게시 산출물 전체 원값 검색 0. 필드 owner가 아닌 역할의 장부 쓰기 → 거절. 관찰 읽기 경로에 쓰기 호출 0.
- 변형 단언: 범주 하나를 빠뜨린 구현은 Negative 시험에서 FAIL.
- RED 이유: refiner 없음. 초록 경로: M5(lane-20).

#### AC-AF-014 — 인증·rate limit·예산

- 주 REQ: maps REQ-AF-014. 초록: M2.
- 정본 주 시험 15: M17, M28, M43, M44, GH05, GH06, SYNC09, SYNC10, WT15, SYNC13~SYNC18. 전용(제안): `TestForgeClassify_*`.
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(M17|M28|M43|M44|GH05|GH06|SYNC09|SYNC10|WT15|SYNC1[3-8])$|^(TestForgeClassify_.*|TestBudget_.*)$' -count=1 -v` → PASS 줄 = 15 + 분류 14 + 예산 4.
- 분류 사례: design §4.5 G1~G8, L1~L6 각각 → 표의 분류와 재시도 수. G6 → `unclassified-403`, 재시도 0.
- 예산 사례: (a) 재개 시각 전 같은 자원 요청 0. (b) 헤더 없음 → `UNKNOWN`, 남은 몫 있음으로 보지 않음. (c) 설정 기본값을 바꾸면 대기 동작이 따라 바뀜(코드 상수 변형을 죽임). (d) 예비 몫 아래에서 안전 쓰기 > 안전 알림 > 보고 댓글 순서. 다른 자격·endpoint로 전환 0.
- 변형 단언: 403을 모두 권한으로 보는 구현은 G1·G2, 모두 rate limit으로 보는 구현은 G5·G6에서 FAIL.
- RED 이유: 분류기 없음. 초록 경로: M2(lane-20).

#### AC-AF-015 — 공급자 관문

- 주 REQ: maps REQ-AF-015. 초록: M6.
- 정본 주 시험 1: M21.
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(M21)$|^(TestProvider_.*)$' -count=1 -v` → PASS 줄 = 1 + 사례 4. 이어 이전 단계 기록 점검(design §9.7 CX-1~CX-3).
- 사례: (a) `codex.enabled` 미설정 → Codex 호출 계수 0. (b) `false` → 0. (c) `true` → 계수 1 이상(계수기가 실제로 셈). (d) GLM 백엔드 호출 0.
- 이관 점검: `progress.md` §J의 CX 행 세 개가 있고, 커밋 그래프로 CX-1 ⊂ CX-2 ⊂ CX-3, 그 사이 모든 커밋에서 Codex 기본값 짝이 `run_base_sha`와 같음.
- 변형 단언: 계수기가 늘 0인 구현은 (c)에서 FAIL.
- RED 이유: 관문 없음. 초록 경로: M6(lane-17 키, lane-20 호출 경계).

#### AC-AF-016 — 전이표

- 주 REQ: maps REQ-AF-016. 초록: M1.
- 정본 주 시험 6: M16, M41, M42, M48, M54, M56. 전용: `TestTransition_Table_*`, §12 `TestState_SpeclessIssue_TransitionsWithoutSpecHash`, `TestState_OperatorDecision_RequiresOperatorEvidence_CoordinatorRefused`.
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(M16|M41|M42|M48|M54|M56)$|^(TestTransition_Table_.*|TestState_SpeclessIssue_TransitionsWithoutSpecHash|TestState_OperatorDecision_RequiresOperatorEvidence_CoordinatorRefused)$' -count=1 -v`.
- 사례: (a) 상태 7 × 사건 전부의 표 구동 시험: E01~E27에 있는 쌍만 적용, 나머지는 거절하고 상태 불변(쌍 수를 출력에 기록하고 0이 아님을 단언). (b) 요청자가 맞지 않음 → 거절. (c) PARKED 사유 집합 밖 값 → 거절. (d) E20 → 단계·owner 불변.
- 변형 단언: 표에 없는 간선 하나를 허용하는 구현은 (a), owner를 바꾸는 재개는 (d)에서 FAIL.
- RED 이유: 상태 엔진 없음. 초록 경로: M1(lane-18).

#### AC-AF-017 — record와 수집기

- 주 REQ: maps REQ-AF-017. 초록: M1.
- 정본 주 시험 11: M45, M49, M50, M51, M52, M53, SYNC01, SYNC02, SYNC06, SYNC07, SYNC08.
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(M45|M49|M50|M51|M52|M53|SYNC01|SYNC02|SYNC06|SYNC07|SYNC08)$' -count=1 -v` → PASS 줄 11. 이어 `-run '^TestCollector_.*$'`의 사례 2.
- 사례: (a) 같은 저장소에 수집기 둘 시작 → 하나만 활성. (b) agent 패키지에서 polling 경로 호출 0(정적).
- RED 이유: record 없음. 초록 경로: M1(lane-18).

#### AC-AF-018 — 맥락 장부와 증거 위치

- 주 REQ: maps REQ-AF-019. 초록: M5.
- 정본 주 시험 5: CONTEXT01, CONTEXT02, CONTEXT03, CONTEXT05, CONTEXT06. 전용: `TestLedger_*`, `TestEvidenceStage_*`.
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(CONTEXT0[12356])$|^(TestLedger_.*|TestEvidenceStage_.*)$' -count=1 -v`.
- 사례: (a) 미추적·ignore 경로만 있음 → `inaccessible-local`, 링크 게시 0. (b) 로컬에만 있는 커밋 → `inaccessible-unpushed`, 오프라인 `accessible` 0. (c) HEAD 변경 → `STALE`, 항목 수 불변, 원 판정값 보존. (d) blob 바뀜 → `STALE`. (e) 원격 조회 실패 → `UNKNOWN`. (f) 단계 `PUSHED`에서 원격 조회 전 `AVAILABLE` 0. (g) 링크가 branch 이름 URL → 거절(전체 커밋 SHA만).
- 변형 단언: 로컬 브랜치 정보로 `accessible`을 내는 구현은 (b), STALE 항목을 지우는 구현은 (c)에서 FAIL.
- RED 이유: 장부 없음. 초록 경로: M5(lane-20).

#### AC-AF-019 — 관찰 보기

- 주 REQ: maps REQ-AF-020. 초록: M5.
- 정본 주 시험 0. 전용(제안): `TestObservationView_ShowsPrepAndPRBacklogStatesDistinctly_ReadOnly`.
- 명령: `go test ./internal/agenticfactory/... -run '^TestObservationView_.*$' -count=1 -v` → PASS 줄 1 이상, 안의 사례 4.
- 사례: (a) design §8.1 항목마다 값·원천·읽은 시각. (b) 준비 상태 7행과 PR 적체가 BUILDING 이후 작업과 다른 구역에 표시. (c) 실행 전후 control ref 순번과 트리 키 동일, 쓰기 호출 0. (d) 출력에 C1~C5 원값 0. 종료 코드는 design §8.
- 변형 단언: 준비 중인 Issue를 시작한 작업으로 세는 구현은 (b), 읽다가 상태를 고치는 구현은 (c)에서 FAIL.
- RED 이유: 관찰 보기 없음. lane-20의 monitor는 포화·적체 값만 보인다는 보고가 있다(미검증). 초록 경로: M5(lane-20).

#### AC-AF-020 — 설정·브라우저

- 주 REQ: maps REQ-AF-021. 초록: M6.
- 정본 주 시험 9: M38, M46, M46a~M46g.
- 명령: `go test ./internal/agenticfactory/... ./internal/config/... -run '^TestAcceptanceManifest$/^(M38|M46|M46[a-g])$|^(TestAgenticFactoryConfig_.*|TestBrowserSelect_.*)$' -count=1 -v`.
- 사례: (a) 범위 밖 값 → 키·범위·파일을 밝혀 거절, 깎기 0. (b) 읽기 오류 → 미설정으로 처리 0. (c) 선택지는 정확히 세 개. (d) 미선택 → `chrome`. (e) `aside` 선택·미설치 → PARKED `not-installed`, `chrome` 호출 0. (f) 선택 저장 뒤 다시 읽기 → 같은 값.
- 변형 단언: 미설치 시 대체하는 구현은 (e)에서 FAIL.
- RED 이유: 설정 연결 없음(읽기 코드는 lane-17 브랜치에 있으나 이 트리에 없음). 초록 경로: M6.

#### AC-AF-021 — CLI

- 주 REQ: maps REQ-AF-022. 초록: M6.
- 정본 주 시험 23: CLI01~CLI23.
- 명령: `go test ./internal/cli/... ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(CLI(0[1-9]|1[0-9]|2[0-3]))$|^(TestCompat_.*|TestLauncher_.*)$' -count=1 -v`.
- 사례: (a) S1: golden 파일과 도움말·거절 문구·출력·종료 코드 바이트 동일. (b) S2: stderr 한 줄만 추가. (c) 버전 선언 없이 새 문법 → 호환 오류(종료 2). (d) 새·옛 문법 혼합 → 호환 오류. (e) 역할 뒤바뀜 0: 모든 플래그 조합에서 세션 역할이 선언과 같음. (f) 조용한 흡수 0: 모르는 인자는 거절. (g) `--` 뒤 인자는 해석 없이 전달.
- golden 순서 점검: golden 커밋이 S1 등록 커밋의 엄격한 조상.
- RED 이유: 새 명령 없음. 초록 경로: M6(launcher 담당). OD-2의 4건이 정해지면 (c)(d)의 문법을 고친다.

#### AC-AF-022 — template·분류·setup

- 주 REQ: maps REQ-AF-023. 초록: M6.
- 정본 주 시험 28: WT01~WT14, WT16, GH01~GH04, GH07~GH15.
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(WT(0[1-9]|1[0-46])|GH(0[1-4]|0[7-9]|1[0-5]))$' -count=1 -v` → PASS 줄 28.
- 추가 사례: 공급자가 지원하지 않는 계층 기능 → `not supported` 보고. 보안 실패 댓글은 묶음 window를 기다리지 않음.
- RED 이유: template 연결 없음. 초록 경로: M6.

#### AC-AF-023 — 환경 행렬

- 주 REQ: maps REQ-AF-024. 초록: M6.
- 정본 주 시험 0. 전용(제안): `TestEnvMatrix_StepEvidenceRequired_NotRunNeverSuccess`.
- 명령: `go test ./internal/agenticfactory/... -run '^TestEnvMatrix_.*$' -count=1 -v`.
- 사례: (a) 단계 칸에 명령·종료 코드·출력 중 하나라도 없음 → 행 거절. (b) 실행 안 한 단계 → `not-run`, `PASS` 0. (c) 전부 `BLOCKED`인 행렬 → 행렬 FAIL. (d) Codex 미설정 환경 → `BLOCKED(not-configured)`.
- RED 이유: 구현 없음(lane-22 보고). 초록 경로: M6.

#### AC-AF-024 — 효율 측정 기록

- 주 REQ: maps REQ-AF-025. 초록: M5.
- 정본 주 시험 6: EFF01~EFF06.
- 명령: `go test ./internal/agenticfactory/... -run '^TestAcceptanceManifest$/^(EFF0[1-6])$|^(TestEfficiencyRecord_.*)$' -count=1 -v`.
- 사례: (a) design §4.7 필드 누락 → 기록 거절. (b) `estimated` 값을 `measured` 칸에 넣음 → 거절. (c) 품질 회귀가 있는 비교 → 효율 향상 보고 0. (d) checkpoint 기록의 비용 필드는 측정값이거나 Gap 표시, 추정값 0.
- RED 이유: 기록 형식 없음. 초록 경로: M5(lane-22).

#### AC-AF-025 — 실제 Issue 시나리오

- 주 REQ: maps REQ-AF-006(보조: REQ-AF-001·004·005·011). 초록: M7 뒤 단일 실행.
- 정본 주 시험 0. 전용: `TestScenarioSuite`(§A.1).
- 명령: §A.1.
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
| 009 | 필드 빠진 receipt 수락 | (a) |
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
