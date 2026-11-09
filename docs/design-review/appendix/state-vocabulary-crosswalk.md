# 부록 1 — 상태 어휘 대응표 (DR-07 #2143, design-docs-r2)

기준: `schemas/protocol/v1/status-crosswalk.json`(7개 어휘 69토큰, variant A·B)와 `docs/protocol-v1.md`(HOLD-00D-1 — 상태 어휘 통합 수준은 P06 미결). 이 표는 검토 파일이 열거한 문서별 어휘를 RC1 소스 어휘에 대응시킨다. RC1 열은 모두 tip `604792666` 소스에서 확인했다(확인 근거는 각 행; 가설 없음).

## RC1 어휘 (소스 확정)

| 도메인 | RC1 어휘 | 소스 |
| --- | --- | --- |
| update receipt | `pending` \| `committed` \| (`recovered` — recovery) | `src/update.rs` HostReceipt.status, `tests/e2e_host_fixture.rs:319` |
| update host 결과 | `applied` \| `unchanged` \| `conflict` \| `blocked` \| `failed` \| 전체 `partial` \| `unchanged` | `src/update.rs` HostOutcome.status, apply_update overall |
| update 파일 action | `create` \| `update` \| `unchanged` \| `conflict` \| `preserved` \| `remove_blocked` | `src/update.rs` FilePlan.action |
| admission | `ready` \| `stale` \| `blocked` (+사유 `snapshot-digest-mismatch`·`approval-reference-mismatch`·`expected-revision-mismatch`·`not-approved`·`approval-revoked`·`approval-scope`) | `src/plan.rs` Decision enum, 사유 코드 |
| claim 상태 | 활성 8: `assigned`·`running`·`verifying`·`pr_ready`·`in_review`·`blocked`·`handoff_pending`·`uncertain`; 비활성 3: `unclaimed`·`integrated`·`cancelled` | `src/claim.rs` ACTIVE_STATES |
| claim outcome | `already_owned`·`expired` 등 acquire/release 결과 | `src/claim.rs` outcome 필드 |
| sync | 진행 7: `ready→reconciling→document_proposed→authorized→publishing→confirmed→closed`; 정지 5: `needs_rebase`·`needs_amendment`·`conflict`·`blocked`·`publish_unknown` | `skills/moai/workflows/sync.md:13` |
| verify | `passed` \| `failed` \| `not_run` \| `blocked` \| `stale` \| `unknown` | `skills/moai/workflows/verify.md` Output |

## 문서별 어휘 → RC1 대응

### [83f] (02-page, 8단계 + 결과 6) — crosswalk `lifecycle` 14토큰과 동일

원문(02:148, 사본 L1223): `discovered → selected → planned → previewed → authorized → staged → applied → host_verified`, 도중 `unsupported·conflict·permission_denied·partial·blocked·unknown`.

- `staged/applied` ↔ update receipt `pending/committed`(stage→publish 구조 동형). `host_verified` ↔ RC1 대응 없음 — T2 운영자 실측 전이며 `host_loaded:false` 원칙(`docs/rust-product-path.md`) 유지. `unknown` ↔ verify `unknown`·sync `publish_unknown`(재합의 불가 상태의 도메인별 표기).

### [25f] (릴리스·Desktop 상세, 10개 — S4 집계)

원문은 단일 목록 없이 분산(`blocked` L20·L632, `stale` L73·L459, `unknown` L424, `release blocked` L489). 토큰 전수는 원문 재검토 필요 — 검토 파일(review-group3 §7.2 등)이 열거한 것은 위 4토큰이며 S4가 10개로 집계했다. `blocked` ↔ admission·update·sync `blocked` 동명 이의 주의(도메인 표 참조).

### [e08] (PLAN 상태 전이, 10)

`Draft·Investigating·NeedsInput·Planned·Reviewed·SnapshotReady·Ready·Running·Stale·Cancelled` (사본 L58-67).

- `Ready/Running/Stale/Cancelled` ↔ plan admission `ready/stale` + claim `running/cancelled`. `SnapshotReady` ↔ plan `expected_revision` pin. `NeedsInput` ↔ claim `blocked`(사유 명시) 또는 HOLD 보고.

### [3d9] (RUN, 4+3)

`Ready → Running → Review → Verified`, 실패 시 `Blocked·Ready(되돌림)·Cancelled` (사본 L165).

- `Running` ↔ claim `running`. `Review` ↔ claim `verifying`/`in_review`. `Verified` ↔ RC1 상태 아님 — verify receipt `passed` + 독립 감사로 분리("process exit 0만으로 Verified 선언 금지"는 RC1 원칙과 일치).

### [66a] (설치·UPDATE, 11) — RC1 claim과 거의 동일

`unclaimed → assigned → running → verifying → pr_ready → in_review → integrated` + `blocked·handoff_pending·uncertain·cancelled` (사본 L1256).

- claim.rs ACTIVE_STATES 8 + 비활성 3과 **1:1 일치**. `integrated`는 RC1 상태 아님 — 리더 게이트(통합 큐)가 소유(S4 결정: "[66a]를 기준으로 삼고 integrated는 '리더 게이트, RC1 상태 아님' 표시").

### [0259] (attempt, 14) — crosswalk `attempt` 14토큰과 동일

`prepared·authorized·submitted·running·unknown·result_ready·cancel_requested·reconciling·blocked·verifying·accepted·needs_fix·cancelled·closed` (사본 L46-51 확인).

- `submitted` ↔ forge outbox intent→dispatched(ACK는 완료 아님 — `src/forge/outbox.rs`). `result_ready` ↔ effects `reconcile` 전 단계. `needs_fix` ↔ verify `failed`+`needs_fix` 사유.

### [0d45] (SYNC owner, 효과 상태 3벌)

L18 `proposed, applied, not_required, blocked, unknown` / L29 `"status": "proposed"` / L71 상태 기계 `ready→…→closed`.

- L71 기계는 RC1 sync.md와 동일 + RC1이 `needs_rebase`를 추가. L18의 `applied` ↔ update `committed`로 정규화 권장(S4: `applied`↔`confirmed` 대응 부재 해소). `not_required` ↔ RC1 대응 없음(범위 밖 표기).

### Core Brain (상태 기계 3개)

`SetupRun`(L854)·context(L4334)·archive(L3776-3786)가 `blocked`·`stale`을 각자 정의 — RC1에서는 도메인이 다르다: `blocked`는 admission(사유 코드 있음)·sync(게이트 부재)·verify(실행 불가)로 각기 다른 사유를 가지며, `stale`은 plan admission(`snapshot-digest-mismatch` 등)과 verify(측정 시점 경과)로 나뉜다. 교차 매핑은 `schemas/protocol/v1/status-crosswalk.json` variants A·B 참조.

## 금지 승격 (REQ-00D-06, crosswalk forbidden_promotions)

`not_run·unknown·blocked·stale·uncertain·publish_unknown·inherited_uninspected` → `passed·applied·host_verified·verified·completed·ok` 매핑은 금지다. 임의의 도메인에서 측정되지 않은 토큰은 성공 값으로 승격되지 않는다.

## 누락 0 확인 방법 (AC-1)

검토 파일이 열거한 어휘 목록 = synthesis S4 행의 7문서 집계 + review-group2 §에지·상태 절 + review-group3 문서별 절. 위 표는 그 7문서([83f]·[25f]·[e08]·[3d9]·[66a]·[0259]·[0d45]) + Core Brain 3기계를 각각 한 절로 다루며, [25f]의 분산 어휘는 위와 같이 한계를 명시했다. 재확인 명령: `grep -c '"' schemas/protocol/v1/status-crosswalk.json` 로 69토큰 존재 확인, 각 사본 행 번호는 표 기재.
