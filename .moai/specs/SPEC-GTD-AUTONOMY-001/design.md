---
id: SPEC-GTD-AUTONOMY-001
created: 2026-09-15
updated: 2026-09-15
---

# SPEC-GTD-AUTONOMY-001 설계

## 1. 시스템 경계

```text
사용자 / 허용 자료원
  → GTD Capture · Clarify · Organize · Reflect · Engage
  → 구조화 Proposal / Decision
  → deterministic policy executor
  → 기존 SQLite queue와 foreman/lead/lane
  → plan → run → sync
  → audit · CI · PR readback
  → mission completion 또는 blocked
```

GTD 계층은 실행 후보를 정리한다. 기존 queue는 승인된 개발 카드를 저장한다. `mission-governor`는 봉인된 정책 안에서 선택하되 상태를 직접 변경하지 않는다. 정책 실행기가 권한과 최신성을 검사한 뒤 기존 소유 component를 호출한다.

## 2. 명령·생성물 구조

- CLI에는 `gtd`를 canonical name으로 등록하고 `todo`가 같은 command construction과 handler를 사용하게 한다.
- slash/workflow/agent skill은 GTD 원본 하나를 두고 todo 표면은 얇은 compatibility entry로 둔다.
- template source를 먼저 수정하고 emitted artifact는 생성·golden 검사로 동기화한다.
- 내부 `kanban`, `BacklogStore`, `todo/backlog.db`는 compatibility boundary 뒤에 유지한다.

## 3. 저장 구조

기존 `meta/items/findings/archived_items/archived_findings` 계약은 유지한다. 별도 GTD metadata version을 두고 다음 논리 집합을 추가한다.

- inbox item과 classification
- project/action/reference/someday/waiting/scheduled 속성
- relation assertion과 provenance
- delegation policy와 sealed version/hash
- mission, event, decision, operation receipt
- logical source revision과 projection metadata

기존 schema-version을 신규 GTD schema-version으로 덮어쓰지 않는다. migration은 additive, transactional, idempotent여야 한다. old/new 왕복에서 구버전 write가 알 수 없는 데이터를 보존할 수 있는지 실측하고, 불가능하면 feature를 기본 off로 유지하며 손실 없는 경계가 마련되기 전 배포를 막는다.

## 4. 관계와 projection

관계 원본은 DB다. 각 assertion은 source, kind, target, source revision, observed time, assertion status, policy version을 가진다.

- `depends_on`: 후속→선행, verified completion만 차단 해제
- `part_of`: action→project, 비차단
- `supported_by`: decision/work→evidence, 비차단이나 evidence revision 변경 시 stale
- `related_to`: 대칭 탐색, 비차단
- `supersedes`: new→old, 자동 삭제 금지

비공개 projection은 홈 DB sibling인 `~/.moai/db/<project-key>/todo/gtd-edges.jsonl`과 `gtd-edges.meta.json`에만 둔다. `todo/` 디렉터리는 `0700`, projection·sidecar·임시 파일은 생성 순간부터 `0600`이어야 한다. 허용 독자는 현재 OS account와 봉인된 mission이 명시한 같은 account의 local MoAI process뿐이다.

한 read transaction의 logical revision으로 node namespace를 분리하고 tuple 정렬·dedupe 후 임시 파일과 sidecar를 atomic publish한다. privacy classifier가 하나라도 결정하지 못하면 전체 publish를 거절한다. sidecar는 source revision, hash, count만 담더라도 projection과 같은 민감도로 취급한다.

기본 노출 정책은 fail-closed다. repo, template, Git staging, application log, trace, telemetry, generic export, diagnostics bundle, crash report, repo graph federation output, 자동 backup에는 projection·sidecar·원문·민감 node identifier를 포함하지 않는다. backup은 사용자가 대상과 보존 기간을 명시한 opt-in에서만 암호화 또는 기존 private backup 경계를 통해 포함한다. 정책 철회는 신규 독자를 즉시 차단하되 audit에 필요한 최소 provenance를 보존하고, item 삭제는 연결 edge와 projection node를 다음 atomic rebuild에서 제거한다. 법적·운영 보존이 지정된 원본 assertion은 tombstone과 만료 정책만 남기며 projection에는 노출하지 않는다.

## 5. 정책과 agent 역할

`MissionSnapshot`은 sealed contract, policy version, queue revision, cards, relationships, lane ownership, evidence, PR SHA, resource remainder, recent failures, advisor recommendation을 포함한다.

`mission-governor`는 permissionMode plan과 읽기 전용 입력만 사용해 schema-valid `Decision`을 반환한다. Decision에는 decision/mission ID, snapshot hash, policy version, action, targets, rationale, confidence, required evidence, operation specs, expiry가 들어간다.

정책 executor는 다음 순서로 검사한다.

1. mission이 approved/running이고 철회되지 않았는지 확인
2. policy version과 contract hash 확인
3. snapshot과 queue/source revision의 현재성 확인
4. action·target·resource가 allowlist와 scope의 부분집합인지 확인
5. dependency, evidence, lane ownership, audit/CI gate 확인
6. stable operation identity와 기존 receipt 확인
7. prepared receipt를 commit한 뒤 소유 component 호출
8. authoritative readback 후 observed_applied와 reconciled 기록

advisor는 권고만 하며 auditor FAIL과 policy deny를 뒤집지 못한다.

## 6. `/moai:goal --auto`

slash workflow는 approval 전 read-only 조사로 versioned mission contract를 만들고 전용 mission create 경로에 전달한다. 자연어는 condition language와 shell command가 아니다.

```text
existing: progression_mode = autonomous | semi-autonomous
new:      mission_mode = auto
```

첫 값은 일반 goal 안의 진행 방식이다. 두 번째는 카드 관리, 배차, commit, PR, merge를 포함하는 전체 임무 state machine이다. `--auto`에는 별도 단축 별칭을 추가하지 않는다.

상태는 `draft / approved / running / blocked / revoking / completed / failed`로 제한한다. approval 후 scope 안의 결정은 질문 없이 진행하며 scope 밖 상황은 blocked report로 종료한다.

## 7. MissionRuntime

런타임 adapter는 `start`, `reconnect`, `replace`, credential refresh, process identity를 각각 probe한다.

| 능력 | 허용 상태 | 전이 |
|---|---|---|
| full: start+reconnect+replace+유효 credential | durable | lease와 heartbeat 아래 동일 sealed snapshot으로 재개 |
| partial: start 가능, reconnect 또는 replace 일부 불가 | 조건부 active session | 지원되지 않은 transition 전에 `active-session-only`로 전이하고 durable claim 금지 |
| unsupported 또는 credential 복구 불가 | active-session-only | 현재 session 밖 지속 실행 거절 |

reconnect 실패나 credential expiry는 새 session을 곧바로 중복 기동하지 않는다. 먼저 lease와 process identity를 조정하고, replace가 지원되며 이전 owner가 더는 effect를 만들 수 없음이 확인된 경우에만 동일 `mission_id`, contract hash, snapshot hash로 replacement를 시작한다. lease loss는 현재 writer를 정지시키고 takeover candidate가 authoritative receipt를 replay한 뒤 하나의 active owner만 얻도록 한다. restart는 마지막 reconciled operation부터 같은 snapshot lineage로 replay한다. PID, idle, timeout, exit 0은 실제 작업 상태의 보조 신호일 뿐 completion evidence가 아니다.

## 8. 배차와 idempotency

foreman은 계속 picked card만 배차한다. GTD manager와 governor가 발행·선택 Proposal을 만들 수 있지만 executor가 standing delegation과 fresh revision을 확인한 때만 기존 queue mutation을 호출한다.

각 논리 작업은 재시작에도 같은 `operation_id`를 사용한다. receipt lifecycle은 `prepared → invoked → observed_applied → reconciled`다. invoked 이후 결과 기록 전에 종료되면 재호출 전에 queue/dispatch/commit/PR/merge readback을 수행한다.

## 9. repo-local commit·통합·release

```text
local_develop_base_sha
  → launcher-entered WT-gtd-autonomy / card_head_sha
  → single local develop integration worktree / local_develop_merge_sha (--no-ff)
  → lead batch push / origin_develop_sha
  → CI PASS on origin_develop_sha
  → release_head_sha created from verified develop
  → release→main PR / main_landed_sha
```

lane/manager-develop은 local develop에서 launcher로 들어간 `WT-gtd-autonomy`에서 explicit path만 stage하고 카드 ID를 commit에 남긴다. lane은 push나 card PR을 만들지 않는다. manager-git만 single local develop integration worktree의 lease를 얻고 branch/HEAD를 다시 읽은 뒤 `--no-ff` merge한다. lead가 local merge들을 batch-push하고 `origin_develop_sha` readback과 그 SHA의 CI PASS를 확보한다. release branch는 이 검증된 develop SHA에서만 만들고 protected main은 release PR로만 진입한다. 각 owner, lease, SHA, audit, CI, review를 별도 receipt에 기록한다. head drift, conflict, missing review, FAIL, protection refusal은 blocked다.

## 10. 실패와 복구

- 정책 위반: side effect 전 deny receipt와 blocked reason
- governor timeout/schema failure/repeated decision: bounded retry 후 blocked
- ambiguous external result: 동일 operation 재호출 금지, readback/reconciliation 우선
- graph publish crash: 기존 complete projection 유지, 불완전 temp 무시
- revocation: 새 prepared 작업 금지, invoked 작업 readback, safe recovery point 기록
- 완료: 계약의 모든 evidence와 `main_landed_sha` ancestry가 현재 상태에서 확인된 경우만 기록

## 11. 보안 경계

- 외부 content와 tool output은 evidence이며 instruction authority가 아니다.
- mission text는 parser나 shell 실행 문자열에 삽입하지 않는다.
- governor는 Bash, write, queue, git tool을 가지지 않는다.
- 개인 메모는 repo graph와 commit 대상으로 자동 승격하지 않는다.
- policy와 decision schema validation은 fail-closed다.
- 비밀정보, 결제, 배포, 권한 변경, 외부 메시지는 명시 계약이 없는 한 금지한다.
