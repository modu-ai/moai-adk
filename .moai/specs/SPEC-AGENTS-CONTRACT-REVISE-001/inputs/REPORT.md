# ADK 제한 해제 후 공식 작업 재개 확인

## Claim

**2026-10-19 05:11 UTC 이후 리더와 기존 레인의 실제 재개를 확인했다.** 시간 경과만으로 해제를 판단하지 않았다. 최초 화면은 `Usage limit reached / Continuing shortly`였지만, 다음 화면에 05:11:24Z의 새 Bash 실행 결과와 갱신된 5H 2%가 나타났고 이후 정상 응답·배차가 이어졌다. 관측한 화면에 새 로그인·계정 선택·결제 요청은 없었다. 세션 재시작, 계정·모델 전환, 추가 크레딧 구매, 반복 입력으로 제한을 우회하지 않았다.

리더에게 최신 방침·기존 검토 자료를 한 번 전달했다. 그 뒤 새로 확인한 리뷰 전 병합과 실제 Codex 성공 응답은 각각 별도 사실 알림으로 한 번씩 전달했다. 기존 자동 재개를 중단하거나 다른 writer를 만들지 않았다. 리더에게 구현·수리·빌드·시험을 맡기지 않았다.

CI 전환·코드 수리·문서 적용·lessons 정리 완료를 뜻하는 보고는 아니다. 공식 담당에서 실제 착수한 단계와 남은 조건을 아래에 구분한다.

## Evidence

### 실제 시작한 공식 작업

| 카드·담당 | 확인한 실행 | 기준·상태 |
|---|---|---|
| t1790 / lane-22 | 공식 만료 lease 복구 후 기존 tree의 재현 시험 준비와 갱신. t1790의 수정된 시험 파일 2개를 보존 | `.moai/worktrees/t1790`, HEAD 74ad197380, 05:18:24Z lease 유효(05:32:05Z 만료). RED 실행은 슬롯 대기 |
| t1868 / lane-22 | 리더가 Codex 댓글 프로토콜 재설계를 지금 진행하도록 배차. 담당이 실제 성공 표본을 읽고 `fetch_codex_samples.py` 실행 | PR1828 b2af3db9a1은 그대로. 공식 t1868 lease는 아직 만료 상태이며, 코드 수정·시험 착수까지 확인한 것은 아님 |
| t1885 / lane-45 | 기존 tree 진입, release head 흡수와 착지 준비 시작 | `.moai/worktrees/t1885`; 담당 화면과 리더 수신에서 c34c709dd4 clean 보고. root가 이 최종 SHA를 별도 Git 조회한 것은 아님 |
| t1892 / lane-40 | kickoff 후 run M1 워크플로 YAML 작업을 manager-develop에 배정하여 시작 | 공식 state=run, lease 유효(05:27:30Z 만료). 무거운 Go 시험은 슬롯 순서 대기 |
| t1900 / lane-44 | 기존 t1884+t1886을 통합 재발급. tree 생성·진입, 검토 보고서와 patch 보존, plan 준비 시작 | `.moai/worktrees/t1900`, WT-agents-agents-md, 8249bff8c3 기준. lease 유효(05:30:09Z 만료). t1882 tree는 keep로 보존 |
| t1901 / lane-23 | lessons 안전 정리 카드 배정·임대·tree 진입, 기존 분석 2건 보존·읽기, 원본 소스 해시 고정 시작 | `.moai/worktrees/t1901`, WT-lessons-inbox-adk, 8249bff8c3 기준. lease 유효(05:32:01Z 만료). t1653과 겹침·경계 기록 |

t1901 담당 화면은 실제 inbox 23,031행/checkpoint23,339를 **읽기 전용** 확인했다고 보고했다. 이번 root가 그 행 수를 별도로 다시 계산하지 않았으므로 담당 관측으로만 귀속한다. 실제 inbox 제거·검증 완료는 확인되지 않았다.

`moai factory status --run tn1p4e --json`의 05:18:24Z 관측:

```text
t1790 leased lane-22 lease_expired=false
t1868 leased stage=merge-ready lane-22 lease_expired=true
t1892 run lane-40 lease_expired=false
t1900 leased lane-44 lease_expired=false
t1901 leased lane-23 lease_expired=false
```

리더는 Go 시험 자제 창을 닫고 기존 슬롯 순서를 담당에게 전달했다. t1899/lane-29는 슬롯 대기 monitor를 두고 아직 Go 명령을 돌리지 않았다고 응답했다. root는 슬롯을 강제로 회수하거나 진행 시험을 종료하지 않았다. `slot.json`은 05:15:13Z `held=true, stale=false, expired=false`를 기록한다. 익명 session의 PID를 특정 레인으로 추정하지 않는다.

### 리뷰 전 병합 발견과 후속 정정

`gh pr view 1841 --repo modu-ai/moai-adk --json state,mergedAt,mergeCommit,headRefOid,reviews,comments,statusCheckRollup,autoMergeRequest`:

```text
state=MERGED
mergedAt=2026-10-19T05:13:31Z
mergeCommit=8249bff8c3933d647b4f99e59048595f9808e83a
headRefOid=d4ea064178412c0d8c1dc149fc83d4ff81470511
reviews=[]
```

05:13:48Z 관측에서 native Codex 댓글은 code/security 모두 running이었다. ready 전환으로 시작된 code review 시작05:13:37Z, security 시작05:13:38Z는 병합보다 늦다. 따라서 **병합 전에 실제 리뷰가 완료되어야 한다는 조건은 지켜지지 않았다.** root의 병합이 아니며 승인한 결과로 세지 않는다.

이를 리더에게 알렸고, 리더는 이후 ready 전환 → 실제 리뷰 완료 대기 → 병합 순서를 답했다. 옛 병합들도 점검하고 #1836의 옛 SHA P2 지적은 lane-26에 후속 수리를 배정했다. 그 결과 신규 draft PR1842가 열렸음은 원격에서 직접 확인했다. 다른 옛 PR에 대한 리더의 사후 점검 주장은 이번 root가 전부 재조회하지 않았다. revert·force push는 하지 않았다.

05:16:12Z 재조회에서는 같은 head의 code review가05:14:11Z, security review가05:15:04Z에 completed로 바뀌었으며 두 상태 모두 `has_any_findings=false`, `has_unresolved_findings=false`였다. `reviews=[]`는 그대로다. REST 댓글6260188619의 신원도 확인했다:

```json
{"id":199175422,"login":"chatgpt-codex-connector[bot]","type":"Bot"}
```

성공 산출물은 [실제 Codex 댓글](https://github.com/modu-ai/moai-adk/pull/1841#issuecomment-6260188619)의 `codex-review-state:v1`이다. `native-codex-rest-comment.json`, `pr1841-followup.json`에 보존했다. **사후 성공은 병합 시점의 누락을 소급해 통과시키지 않는다.** PR1828 gate는 현재 `reviews` 배열의 `APPROVED`를 요구하므로, 담당 lane-22가 실제 댓글 성공 형식에 대한 계약 수정·재현·감사·시험을 진행하도록 리더가 배차했다. root는 gate를 수정하거나 제품 시험을 실행하지 않았다.

### 최신 GitHub 목록과 병합 방침

05:18:07Z GraphQL은 issue totalCount3/nodes3, PR totalCount6/nodes6, 두 연결 모두 hasNextPage=false다.

| 열린 PR | head | 작성자·상태 |
|---|---|---|
| #1828 | b2af3db9a1a7c3554110a3b4bd0dc588ed6d29ac | GoosLab, draft |
| #1830 | 8249bff8c3933d647b4f99e59048595f9808e83a | GoosLab, draft |
| #1833 | 54be77d2fd474258f610dad8f33ca1dd8b5fb085 | GoosLab, draft |
| #1839 | 43d0340ad5b54e37783eddccb75129babcbf8f33 | dependabot, non-draft |
| #1840 | f96d3f3910ae81dc5a4603d8e5dad7c4b28dbd7c | dependabot, non-draft |
| #1842 | 89dd874419f410ed62556a3e5b00f0896772b060 | GoosLab, draft, #1836 P2 후속 |

전부 autoMergeRequest=null이다. 이슈1813/1817/1827은 계속 열려 있다. 다른 작성자·fork·bot PR은 사용자 테스트 확인 **및** 명시 승인이 모두 필요하며, #1839/#1840 자동 병합 금지를 리더가 명시했다. 기존 auto-merge의 작성자/fork 필터 부재·자동 branch 삭제 설정은 여전히 미정합으로 기록됐고 변경하지 않았다. 보호와 전체 카드 완료 전 릴리스 조건을 유지했다.

### 입력 영수증과 변경 보존

일회성 입력 requestId:

- 재개 인계: 48736552-c000-4667-91b7-8602f9a5b123
- 새 병합 누락 사실: 6dbc69e4-1981-45ce-80f9-a8d606956b6e
- 새 실제 성공 결과: 11d08a6d-7021-4d51-87ae-404fd9fcc7ae

각 영수증은 accepted=true/input_accepted다. 영수증 자체에는 turn_started가 없었으므로 이를 제출 완료 증거로 바꾸지 않았다. 이후 화면에서 첫 인계의 자료 읽기·t1900/t1901 배정, 누락 알림의 실제 댓글 조회·후속 병합 절차 정정, 담당 lane-22의 실제 표본 조회를 관측했다. 재전송하지 않았다.

05:12:52Z primary는 detached478285ab4c, M14·미추적2였고 기존 경로를 유지했다. t1790의 미커밋 시험2개, t1862의 앞선 로컬 커밋af0de9e48c, t1882 tree를 관측·보존했다. 공유 checkout의 branch 이동·stash·reset·clear·변경 삭제를 하지 않았다. 제품 수정·PR 생성/병합·설정 변경·inbox 정리는 root가 수행하지 않았다.

## Baseline-attribution

모든 결과는 `/tmp/adk-resume-20261019-0510`의 이번 실행 자료에 귀속한다. Orca 화면05:11~05:18Z, 공식 factory 상태05:18:24Z, GitHub 끝 분모05:18:07Z를 구분했다. PR1830은 새 head8249bff8c3로 바뀌었으므로 이전5122060f의 검사 결과를 새 head에 전용하지 않는다.

`moai todo show t1900/t1901`은 실제 SQLite backlog store의 응답이다. 이번 도구는 옆의 `backlog.json`이 export/legacy로 오래됐을 수 있다고 명시했다. 이전 보고서에서 JSON 파일에 대해 센 336/1,274 숫자는 그 파일 관측 수이며 현재 공식 queue 분모로 사용하면 안 된다. 공식 담당·lease는 이번 factory/카드 조회에 귀속한다.

## Gaps

- CI 전환 PR1828의 수정·현 head CI·새 gate 성공·원격 main 적용은 아직 미완료다.
- t1868은 재설계 표본 조회가 시작됐으며 쓰기 전에 공식 lease·대상 worktree를 정리해야 한다. 현재 유효 lease는 lane-22의 t1790에 있다.
- t1900은 plan 준비, t1901은 조사·plan 준비 단계다. 적용·시험·inbox 제거는 완료하지 않았다.
- 새 PR1839/1840/1842의 전체 리뷰·CI·수정 내용은 이번 재개 점검에서 전수 상세 검토하지 않았다. 새 head1830의 CI 완료도 확인하지 않았다.
- 화면의 기존 실행과 공식 카드 소유자는 확인했지만 OS 프로세스 전체와 모든 writer의 파일 접근을 추적한 것은 아니다. 리더의 직접 구현·시험 부재는 관측 화면 범위에 한정한다.
- Q18 질문 화면은 답변이 들어간 상태로 관측됐다. root가 대신 선택하거나 해당 답의 승인 주체를 추정하지 않았다.

## Residual-risk

리더는 실행을 재개했지만 리뷰 완료 전 병합이 실제 발생했다. 이후에는 모든 대상 브랜치에서 같은 검증 조건을 지켜야 한다. native Codex 성공 댓글은 봇 신원·정확한 SHA·현재 본문·완료 상태·지적 해결을 검증해야 하고, 누락/running/실패/옛 SHA/변조·중복 마커는 통과시키면 안 된다. 기존 기능 검사와 보호를 대체하지 않는다. 관련 계약·gate 수정과 fail-closed 시험은 기존 담당에게 남아 있다.

새 로그인·한도·결제 결정이 나타나면 해당 단계만 멈춘다. 이번 점검에서 기존 거부된 광범위 Orca memory 진단이나 금지된 sender 경로를 재시도하지 않았다.
