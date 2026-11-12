# 부록 3 — 인터페이스/명령 대응표 (DR-07 #2143)

출처: synthesis S11 + `src/main.rs` OPERATIONS(전수는 아래 "RC1 OPERATIONS 전수" 절의 목록이 전부다. 이름 집합은 `src/main.rs`의 시험 `operations_doc_matches_code`가 코드 배열과 대조한다). 문서가 그린 인터페이스명(`result.admit` 등)은 RC1에 없고, RC1 OPERATIONS는 문서 어디에도 대응 표가 없었다 — 이 표가 그 대응표다.

## 문서 인터페이스명 → RC1 OPERATIONS

| 문서 인터페이스명 (출처) | RC1 OPERATIONS | 대응 관계 |
| --- | --- | --- |
| `result.admit` ([43fa] 등) | `plan.admit` (`plan-admit`) | admission 판정 — RC1은 `ready/stale/blocked`+사유 코드로 응답 |
| `evidence.admit` | `plan.admit` + `run-verify` receipt | evidence 연결은 plan pin(`PinnedContract`) + verify receipt로 분리 |
| `verify.accept` | `run-verify` (`run.verify`) | verify receipt(6종 상태) 반환 |
| `owner.handoff` | `claim-handoff` (`claim.handoff`) | handoff 기록·인계 |
| `decision.propose` | `decision-record` (`decision.record`) | 결정 원장 append |
| `sync.*` | 없음 (RC1 미출하 — SYNC는 workflow 문서 절차) | `skills/moai/workflows/sync.md` 절차만; 상태 기계는 부록 1 |
| `artifact.pin` | `plan.admit` `PinnedContract` + `update` rollback_manifest | artifact pin은 plan pin/manifest로 실현 |
| `dispatch.*` | `forge-read`/`forge-report` + `effect-intend`/`effect-confirm`/`effects-reconcile` | dispatch는 forge 읽기·outbox effect 3-op로 분리 실현 |
| 배정·수락 관측 (#2390 S2) | `dispatch-record` / `dispatch-ack` / `dispatch-check` (`dispatch.record`·`dispatch.ack`·`dispatch.check`) | S1 위의 revision·수신/수락 구분·5분 경보. C3는 관측까지 충족, 감사 claim 소비·게이트 연동 미충족(#2403/#2388 후속). 계약은 `docs/rust-product-path.md`의 S2 절 |
| 레인↔리더 영속 메시지 (#2390 S1) | `message-record` / `message-ack` / `message-reconcile` (`message.record`·`message.ack`·`message.reconcile`) | 로컬+Issue 기록과 명시 수신 ACK. `queued`·`received`·`received_ack`를 구분하며 dispatch 수락·소유권 이전·업무 완료는 아님. 입력/결과 계약은 `docs/rust-product-path.md`의 S1 절 |
| 레인 통신 메타데이터 관측 (#2390 S6) | `lane-status` (`lane.status`)의 추가 `communication` 객체 | 기존 입력·출력 의미 유지. `available`, `observed_at`, `lanes`(세션별 `sent_reports`, `received_acks`, `unacknowledged_report_ids`, S2 `dispatches`), `errors: [{issue, reason}]`; payload 값은 노출하지 않음. 상세 필드는 `docs/rust-product-path.md`의 S6 절 |
| — (문서 인터페이스명 없음) | `codex-projection-plan` (`codex.projection_plan`) | 대응 없음 — [83f] §6 host adapter projection을 측정하는 op이며 문서 인터페이스 어휘가 이를 이름으로 부르지 않는다 |

S6의 `sent_reports`는 S1 wire의 sender session 기준이며, `received_acks`는 그 레인이 **보낸** 보고에 대한 명시적 receiver ACK다. GitHub 작성자 숫자 ID는 계정 증거이지 독립 세션 인증이 아니다. 기존 Issue/comment GET 스냅샷과 완전한 pagination을 재사용하고, 유효한 S1 marker가 있을 때만 저장소 identity GET을 추가한다. canonical packet·marker/digest·작성자·대상·실제 context를 검증한다. 선언 레인과 발견한 메시지 actor를 결정적으로 정렬하고, 정확히 같은 중복은 합친다. event/context 필터 전에 동일 report ID의 불변성 충돌을 검사하며, 관련 malformed packet·충돌·불완전 관측은 `communication.available=false`와 일반 오류 사유로 남긴다. 기존 상태 출력은 유지하고 ACK를 추정하지 않는다.

5분은 `Message.created_at`이 아니라 서버 record comment의 `created_at` 기준이다. 명시 ACK가 없으면 299초 `waiting`, 300초(N)·301초(N+1) `unacknowledged`; 일치하는 명시 ACK는 `received`다. 미래 시간은 unknown이며 malformed 시간·overflow는 unavailable/unknown이지 0초나 timeout으로 보정하지 않는다. S2 checker를 같은 frozen snapshot에 적용하며 receipt는 acceptance가 아니고 superseded revision은 current 배정이 아니다. GET-only·파일 쓰기 없음·관측 전용 경계를 유지한다. S3·S4 나머지·S5·host transport·claim/audit gate 연동은 **HOLD**다.

## C6d generation 문서 참조 (#2450)

`index-build`가 만드는 Graph와 이를 읽는 `index::index_records` API/`atlas-graph`에는 기존
`imports`/`calls` 외에 Rust 문서 주석의 qualified 식별자에서 얻은 파일→파일
`documents` edge가 포함된다. JSON 필드는 기존 온톨로지 형식을 그대로 사용한다.
각 edge는 문서 언급과 대상 정의의 `file-uri#span@digest`를 모두 보존하며,
`basis=inferred`, `verification=unreviewed`,
`uncertainty=doc-lexical-reference; rustdoc-binding-not-checked`로 한계를 표시한다.

지원 범위는 `//!`·`///`·`/*! */`·`/** */` 문서의 inline-code qualified 경로와
모듈 수준 선언이다. compiler/rustdoc의 완전한 이름 해석·가시성 검사는 수행하지
않는다. bare name·일반 주석·문자열·fenced 예시를 참조로 만들지 않으며,
alias/re-export·associated item은 이 문서 해석기의 지원 범위 밖이다.
상수 선언은 const generic과 구분하는 구문 해석을 구현하지 않아 지원하지 않는다.
사용자 결정 `6801346885`에 따라 물리적인 doc 블록 전체를 먼저 검사한다.
quote/list/fence/4space indentation/tab/HTML 시작 중 하나라도 있으면 해당 블록의
후보를 모두 제외하고 기존 `missing`에 `doc_reference_unsupported`의
`path`, block `line`, `count=1`, `reason`을 기록한다. 앞선 prose도 소급 제외하며
container 수명 추적은 하지 않는다. 물리적 blank/source/attribute나 별도 block comment는
새 블록 경계이고, 지원되는 plain 블록만 backtick qualified 경로를 수집한다.
대상 정의 0개/복수 후보는 `doc_reference_missing`, `reason=unresolved|ambiguous`다.
증분 빌드는 제외 계수도 저장 scan으로 재사용하며 `code-refs=5`로 이전 추출기 fragment를
재사용하지 않는다. 새 top-level 공개 필드는 추가하지 않는다.

리더 결정 `6793749743`에 따라 GRAPH는 기존 7개 코드 관계와 G-002 문서 관계를
확인하고 G-006만 이유 있는 gap으로 남긴다. concept identity/alias H1은 HOLD다.
fixture·ground truth·baseline과 #2440 answer 선택 계약은 변경하지 않는다.
수리 경계·공식 근거·실행/미실행 표는 [C6d 수리 경계](../../atlas-doc-reference-boundaries.md)에 기록한다.

## RC1 OPERATIONS 전수

- `agents-apply`
- `agents-plan`
- `apply-recover`
- `archive-apply`
- `archive-plan`
- `archive-restore`
- `archive-search`
- `atlas-graph`
- `atlas-map`
- `atlas-tag-scan`
- `audit-claim`
- `audit-clone-plan`
- `audit-reduce`
- `baseline-inventory`
- `claim-acquire`
- `claim-handoff`
- `claim-release`
- `claim-show`
- `claim-verify`
- `claude-doctor`
- `claude-settings-apply`
- `claude-settings-plan`
- `codex-doctor`
- `codex-projection-apply`
- `codex-projection-plan`
- `context-bundle`
- `context-verify`
- `cost-report`
- `decision-append`
- `decision-record`
- `dispatch-ack`
- `dispatch-check`
- `dispatch-record`
- `effect-confirm`
- `effect-intend`
- `effects-reconcile`
- `flow-gate-verdict`
- `flow-wip-check`
- `forge-audit-status`
- `forge-links-probe`
- `forge-read`
- `forge-report`
- `generation-apply`
- `generation-plan`
- `github-apply`
- `github-plan`
- `hook-dispatch`
- `hook-receipt`
- `hooks-apply`
- `hooks-plan`
- `identity`
- `index-build`
- `index-query`
- `install-inspect`
- `key-import-apply`
- `key-import-plan`
- `key-remove`
- `key-set`
- `key-status`
- `lane-handoff-check`
- `lane-status`
- `layout-apply`
- `layout-plan`
- `leader-tick`
- `message-ack`
- `message-reconcile`
- `message-record`
- `migrate-apply`
- `migrate-plan`
- `migrate-recover`
- `migrate-recovery-plan`
- `package-inspect`
- `package-resolve`
- `parity-check`
- `plan-admit`
- `pr-body-render`
- `pr-confirm`
- `project-apply`
- `project-plan`
- `project-skill-apply`
- `project-skill-plan`
- `project-skill-revert`
- `queue-next`
- `queue-query`
- `recovery-apply`
- `recovery-plan`
- `run-verify`
- `scope-check`
- `search`
- `snapshot-describe`
- `update-apply`
- `update-plan`
- `worktree-apply`
- `worktree-plan`
- `worktree-report`

(알파벳순, 한 줄에 operation 하나, 개수는 적지 않는다. `src/main.rs`의 시험 `operations_doc_matches_code`가 이 목록과 OPERATIONS 배열의 이름 집합을 비교해 빠진 것·남는 것·중복·순서 어긋남을 이름으로 보고한다 — op를 더하거나 빼는 PR은 알파벳 위치에 자기 줄만 넣거나 뺀다. 이력: tip `604792666` 47개에서 시작해 rc.2 카드들이 더했고, #2241에서 이 형식으로 바꿨다; `lane-handoff-check`은 A10 #2264 수리 r2에서 더했다. envelope 계약: `{protocol_version, operation, ok, result|error{kind, message, retryable, next_actions?}}`, exit 0/1/2, `kind` 14종 — `src/contract.rs`·`docs/protocol-v1.md`)

envelope 예외 하위 명령: `statusline`(#2350). Claude Code가 `statusLine` 명령의 stdout을 그대로 화면에 보여 주므로 JSON envelope 대신 상태줄 원문을 출력하고 항상 exit 0으로 끝난다. OPERATIONS에 들어가지 않으므로 위 목록에도 없다. 실행기 `cc`·`glm`·`codex`·`session`도 같은 이유로 목록 밖이다. 숨은 시험용 옵션 `--render-budget-ms`는 렌더 예산(기본 800ms, Go와 같음)을 바꾼다. golden 시험이 부하와 무관하게 결정적으로 돌도록 큰 값을 넘기고, `0`은 이미 지난 예산(git을 시작하지 않음)이다. `src/main.rs`의 시험 `statusline_is_the_documented_envelope_exception`이 이 문단과 코드를 대조한다.

## 문서별 내부 명령 체계 → RC1

| 문서 | 문서의 명령 체계 | RC1 대응 |
| --- | --- | --- |
| [25f] | `setup` / `project init` / `maintenance update` / `doctor` / `capabilities` | `project-plan/apply`(setup) · `update-plan/apply`(maintenance update) · `claude-doctor`·`codex-doctor`(doctor) · `capabilities`는 **RC1에 없음**(S11: "공통 명령은 capabilities뿐인데 RC1에 없다") |
| [e08] | `spec plan` / `task status` / `skill show` | `agents-plan` + `plan-admit`(spec plan) · `claim-show`(task status) · skill show 없음(`install-inspect`·`context-bundle`이 부분 대체) |
| [3d9] | `task submit/cancel` / `spec verify` / `artifacts list` | `claim-acquire`/`claim-release`(task) · `run-verify`(spec verify) · artifacts list 없음 |
| [43fa] §18 | 문서 명령 전부 | RC1 OPERATIONS 전수 목록(위 절)으로 교체(S11 결정) |

교체 규칙(S11 결정): 문서의 명령 표는 `src/main.rs` OPERATIONS 표로 교체하거나 "RC1 미출하 계약안" 배너를 단다. 번호·이름으로 누락 검사를 하지 않는다(S8 — 같은 F 번호가 다른 기능).
