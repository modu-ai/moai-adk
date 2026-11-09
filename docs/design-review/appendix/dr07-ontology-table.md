# DR-07 부속 — 온톨로지 노드·엣지·basis/verification 통합 근거표 (#2247)

#2180(온톨로지 노드/엣지 schema)의 선행 근거표입니다. 모든 행은 아래 출처 문서의
행 번호를 인용하며, 출처로 검증되지 않은 항목은 추정하지 않고 "확인 불가" 절에
모았습니다.

## 출처 문서 (2026-11-09 읽음, sha256 병기)

| 약칭 | 파일 | sha256 |
| --- | --- | --- |
| [TO] | `references/chatgpt-space/downloads-design-2026-11-09-textonly.md` (§8 온톨로지와 데이터 계약, L326-400) | `b4242abda1b116f7a38ac81530961ce53a0ded9216a95f92b23de8ba8be7d542` |
| [CB] | `references/chatgpt-space/page_136c8203bf2881919eb1f543f72f87ea.md` (Core Brain 상세 설계, L741·L752-810·L2413-2491) | `49e445f464e088d3e2a842cf8622b368139fd16e4bc87d66cebaca11411547fa` |

## 노드 통합 근거표

### [TO] §8 노드와 식별자 (L354-369)

| 노드 | identity와 revision | 주요 속성 | 출처 |
| --- | --- | --- | --- |
| Project·Repository | project UUID와 repository UUID, provider object ID·canonical origin 별도 | visibility·ACL·profile | [TO] L358 |
| Workspace·Snapshot | repository UUID+clone/worktree UUID, commit+dirty overlay digest | base·tree·observed time·parser config | [TO] L359 |
| File | project UUID+repository UUID+normalized path의 논리 ID, content hash별 revision | language·size·generated·sensitivity | [TO] L360 |
| Symbol | repository UUID·언어·package·qualified name·signature·scope로 논리 ID, AST/location별 revision | kind·visibility·span·resolver | [TO] L361 |
| Tag | file revision+comment span+normalized content digest | namespace·type·text·explicit refs | [TO] L362 |
| SPEC·Requirement·AC | stable ID와 immutable revision digest | status·author·approved scope·supersedes | [TO] L363 |
| Test·Evidence | test logical ID, run receipt UUID+candidate/toolchain digest | selection·result·duration·artifact | [TO] L364 |
| Issue·PR/MR | provider+host+repo object ID+object ID | URL·remote version·updated time | [TO] L365 |
| Attempt·Effect | task+attempt UUID, effect UUID | owner·epoch·scope·receipt·state | [TO] L366 |
| Decision·Approval | immutable decision ID+revision | actor·target·allowed effect·expiry | [TO] L367 |
| Document·Report·Log | source URI+content hash+logical record ID | sensitivity·retention·time range | [TO] L368 |
| MemoryClaim | claim UUID+revision | proposed/accepted/rejected/stale, source refs | [TO] L369 |

식별 규칙([TO] L371): 경로 변경은 새 파일 논리 ID와 `renamed_from` 후보 또는 Git
rename 증거로 연결. fuzzy symbol match는 후보일 뿐 기존 identity를 무조건
재사용하지 않음. unsaved buffer는 저장 snapshot과 구분한 ephemeral overlay이며
원격 context에 자동 포함하지 않음.

### [CB] L741 신규 node 추가 (Core Brain 상세 설계)

| 노드 | 정의 | 출처 |
| --- | --- | --- |
| ProjectProfileRevision | Core Brain 기존 node에 추가(또는 DocumentRevision 역할) | [CB] L741 |
| SetupPlan | 동상 | [CB] L741 |
| SkillPackageRevision | 동상 | [CB] L741 |
| RoleDefinitionRevision | 동상 | [CB] L741 |
| HostCapabilitySnapshot | 동상 | [CB] L741 |
| GenerationReceipt | 동상 | [CB] L741 |
| DocumentRevision(역할 표현) | "적절한 DocumentRevision 역할로 표현한다" — 별도 node 승격은 실제 query 요구가 있을 때만 | [CB] L741 |

[CB] L741 후반: "타입 증가는 실제 query 요구가 있을 때만 별도 node로 승격한다."

## 엣지 통합 근거표

### [TO] §8 최소 관계 집합 (L375) + edge JSON schema (L377-390)

`contains, declares, imports, calls, annotated_by, references_requirement,
implements, verifies, tracks, changes, generated_from, supports, contradicts,
supersedes, derived_from, authorized_by, archived_as` — 17개를 최소 관계 집합으로
제안([TO] L375). RDF·graph DB 서버를 필수로 요구하지 않음.

edge JSON 필드([TO] L377-390): `edge_id`(type·endpoints·evidence의 hash),
`type`, `source_id`, `target_id`, `basis`, `verification`, `evidence_refs`
(file-uri#span@content-hash), `extractor`{name, version: pinned}, `snapshot_id`,
`observed_at`(RFC3339), `valid_from`(revision-or-time), `valid_to`(null),
`uncertainty`(예: `semantic-match-not-proven`).

머리글-노드 대응([TO] mermaid L330-348): Project→contains→Repository,
Repository→has revision→Snapshot, Snapshot→contains→File revision,
File→declares→Symbol, Symbol→calls or imports→Symbol,
Symbol→annotated by→Tag, Tag→explicit reference→Requirement,
SPEC revision→contains→Requirement, Symbol→implements candidate→Requirement,
Test→verifies→Requirement, Forge Issue→tracks→Requirement,
Attempt→produces→Patch·Commit, Patch→changes→File revision,
Attempt→produces→Evidence receipt, Evidence→supports or contradicts→Requirement,
Document→derived from→Evidence, Evidence→bound to→Snapshot,
Decision·Approval→authorizes scope→Attempt.

### [CB] L752-810 관계 정의

| 관계 | 정의 | 출처 |
| --- | --- | --- |
| derived_from | profile 필드와 역할이 어떤 사용자 답변·repo 사실에서 나왔는가 | [CB] L752 |
| generated_from | 파일이 어떤 profile·template·generator·source digest에서 만들어졌는가 | [CB] L754 |
| targets_host | 어떤 host/version/capability를 대상으로 검증했는가 | [CB] L756 |
| governed_by | 적용되는 승인 SPEC·policy reference. 링크 자체가 권한을 부여하지 않음 | [CB] L758 |
| verifies | 어떤 smoke test·검증 receipt가 어떤 exact 파일 revision을 검사했는가 | [CB] L760 |
| tracks | Issue의 작업 범위와 안정된 local task ID 연결 | [CB] L762 |
| supersedes / conflicts_with | 새 profile·사용자 수정·생성안의 대체와 충돌 | [CB] L764 |

### [CB] L2413-2491 관계 시맨틱

| 관계 | 시맨틱 | 주의 | 출처 |
| --- | --- | --- | --- |
| revision_of / supersedes | 동일 entity의 버전·대체 관계 | 오래된 버전이 잘못됐다는 뜻은 아님 | [CB] L2413 |
| generated_by / derived_from | run 또는 원문에서 산출 | 에이전트 생성 보고와 trusted test receipt의 근거 등급 구분 | [CB] L2417 |
| documents / references_requirement | 문서·코드가 설명하거나 참조 | 해당 요구를 충족한다는 선언과 다름 | [CB] L2421 |
| verifies / observed_execution | 특정 test/run이 검증·관측한 범위 | 테스트 선택과 assertion·환경까지 기록. 전체 correctness로 확대 금지 | [CB] L2425 |
| reports_on / mentions_issue | 보고·로그·문서가 작업/이슈를 언급 | Issue status 변경 권한이나 해결 증거 아님 | [CB] L2429 |
| archived_as / restored_from | 보관 위치와 복구 이력 | 삭제 또는 새 entity 생성과 다름 | [CB] L2433 |
| included_in_context | 해당 bundle에 어떤 revision과 범위가 포함됐는지 | (정의는 [CB] L2491까지 확인) | [CB] L2491 |

## basis / verification 정의 ([TO] L392)

- `basis`: **explicit** / **static_resolved** / **inferred** / **runtime_observed**
- `verification`: **unreviewed** / **checked** / **contradicted** / **stale**
- 명시 연결은 작성 의도의 근거이고, checked는 검토 범위의 확인이다. 숫자
  confidence가 필요하면 calibration 근거·평가 집합·버전을 함께 저장하고 그렇지
  않으면 범주형 불확실성을 사용한다([TO] L392).

## 동적 복원·삭제 한계 ([TO] L394-396)

- 동적 dispatch·reflection·macro·generated code·DI·runtime loading은 정적 graph로
  완전 복원되지 않는다. 미해결 call을 삭제하지 않고 unresolved node·후보
  집합·resolver 제한을 남긴다. runtime trace는 관측 실행 경로만 증명하며 미관측
  경로 부재를 증명하지 않는다.
- 삭제는 현재 snapshot에서 tombstone 또는 `valid_to`로 표시한다. 이전
  source·evidence가 참조하는 immutable revision은 보존한다. 보존 정책상 지워야
  하는 민감정보는 별도 승인·규정에 따라 제거하고, 단순 index cache GC와 혼동하지
  않는다.

## 확인 불가·충돌

| 항목 | 상태 |
| --- | --- |
| 리더 지시 "노드 8종"의 정확한 8종 목록 — DR-07 본문이 이 작업에 제공되지 않아 [TO] 12행·[CB] L741 7종과 수가 일치하는지 대조 불가 | 확인 불가 — DR-07 본문 제공 시 대조 필요 |
| [TO] mermaid의 `Other symbol`(calls or imports 자기 참조 대상)이 노드 표에 독립 행이 없음 — Symbol의 자기 참조로 해석되나 명시 정의 없음 | 확인 불가 |
| [CB] L2491 `included_in_context`의 상세 정의 — L2491이 표의 마지막 확인 행이며 이후 내용은 제공 범위 밖 | 확인 불가(L2492 이후 미제공) |
| [TO]와 [CB]의 관계 목록 간 명시적 상충 — 없음(derived_from·generated_from·verifies·tracks·supersedes가 양쪽에 공존, 시맨틱도 상충 없음) | 상충 없음 실측 |
