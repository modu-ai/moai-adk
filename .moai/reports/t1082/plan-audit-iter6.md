# SPEC 감사 보고서: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001

Iteration: 6 (DELTA — iter-5 결함 N1-N6 + 리드 전달 idempotency 항목. 운영자 승인 최종 반복)
Revision: `6ad4824a2` (5 files, +75/-43)
Verdict: **PASS-WITH-DEBT**
Overall Score: 0.86 (Tier L 임계 0.85. iter-5 0.81 대비 상승 — STOP 신호 없음. 여유는 0.0125로 얇다)

작성자 추론 맥락(결정 기록의 근거 서술 포함)은 M1 Context Isolation 에 따라 판단 근거로 쓰지 않았다. 결정 기록은 "무엇을 택했는가"를 확인하는 데만 읽었다. 판단 근거는 SPEC 산출물, 착지 코드, 그리고 이번 감사에서 직접 실행한 명령 출력이다.

## Claim

개정 `6ad4824a2` 는 iter-5 의 N1(차단)과 N2-N6, 그리고 idempotency 문구를 모두 닫았다. 새로 찾은 결함은 5건이며 모두 optional 이다. 가장 무거운 N7(major)은 N2 의 거울상이다. reservation 쪽이 source 행을 자기 transaction 안에서 읽는지는 규범상 요구되지만(REQ-FLH-017 불변식, plan.md:40 「CAS로 고정」), 어느 AC 도 이를 판별하지 않는다. 필수 통과 7개는 모두 PASS 이고 차단 결함은 없다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: `grep -oE '^### REQ-FLH-[0-9]+' spec.md` → 001..018, 18건, 공백·중복 없음. AC 도 001..019 연속 19건(E1).
- [PASS] MP-2 GEARS 준수 (요구사항 층에서 판정): 개정된 REQ 다섯 개를 다시 읽었다. REQ-FLH-003(spec.md:77)은 When 문 두 개, REQ-FLH-009(:101)는 While/When + `SHALL NOT`, REQ-FLH-015(:125)는 While, REQ-FLH-017(:133)은 Ubiquitous + While + When 복합형, REQ-FLH-018(:137)은 While/When + 추가 문장 「This rejection SHALL NOT strand …」(Unwanted 형)이다. AC 의 Given-When-Then 은 검증 층 형식이라 MP-2 대상이 아니다.
- [PASS] MP-3 frontmatter: spec.md:1-17 에 12 필드 + `tier: L`. `version: "0.5.0"`(따옴표 semver), `created`/`updated` ISO 날짜, `priority: P1`, `lifecycle: spec-anchored`, `tags` 문자열. 거부된 별칭 없음(E4).
- [N/A] MP-4 언어 중립성: Go 내부 패키지(`internal/factorymsg`) 전용 SPEC, 템플릿 대상 아님.
- [PASS] MP-5 D7: 참조 SPEC 은 자기 자신과 `SPEC-FACTORY-MIXED-HOOK-001`(`status: completed`, version 0.1.4)뿐. retired/superseded/archived 없음(E5).
- [PASS] MP-6 D8: `grep -c syscall spec.md` → 0.
- [PASS] MP-7 clarification gate: `grep -rn 'NEEDS CLARIFICATION' plan.md research.md` → 출력 없음, exit 1.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | iter-5 의 규칙 개수 불일치(design.md:101 「네 가지」로 수정)와 :18 정책 긴장(acceptance.md:18 수정)은 해소. 남은 모호성은 요구사항 두 곳 — REQ-FLH-017(spec.md:133)의 무조건 「later launcher registration or bind SHALL leave the handoff-bound endpoint unchanged」(N8), REQ-FLH-003(spec.md:77)의 dirty target NACK 사유 부재(N10). 합리적 엔지니어는 AC 문맥으로 일관되게 풀 수 있다. |
| Completeness | 0.85 | 0.75–1.0 사이 | 필수 절·frontmatter·`### Out of Scope — …` H3 모두 존재(구조 기준 1.0). 내용 누락 두 건으로 감점: design.md:116 의 「한 transaction 안에 있어야 하는 것」 목록에 reservation 판독이 빠짐(N7), plan.md 마일스톤과 파일 소유 표에 REQ-FLH-016..018 작업(착지된 t1074 `RegisterPeer` transaction 수정 포함)이 없음(N11). |
| Testability | 0.85 | 0.75–1.0 사이 | 모든 AC 가 정확한 `jq -se` 판정식을 가진 이진 판정이다. AC-FLH-019 (iv)는 판독 위치를 실제로 판별함을 프로브로 확인(E6). 감점: reservation 쪽 판독 위치를 판별하는 순서가 없음(N7), AC-FLH-018/019 fixture 문구가 실제 배치 가능한 방법과 어긋남(N9). |
| Traceability | 1.0 | 1.0 | spec.md 표 `§ REQ-FLH-015 → AC-FLH-015, AC-FLH-018`, `§ REQ-FLH-017 → AC-FLH-018, AC-FLH-019`, `§ REQ-FLH-018 → AC-FLH-019`. acceptance.md:44-45 역방향 일치. strict lint `[]`(E2). 모든 REQ 에 AC, 모든 AC 가 실재 REQ 를 가리킴. |

평균 (0.75 + 0.85 + 0.85 + 1.0) / 4 = 0.8625.

## iter-5 결함 폐쇄 판정 (Regression Check)

### N1 REQ18-BLOCKS-T1074-REQUIRED-BINDER — **RESOLVED**

- 결정 ②가 규범으로 들어갔다. REQ-FLH-017(spec.md:133) 「When a t1074 launcher provisional registration commits on a lane whose handoff is non-final, that same registration transaction SHALL move the handoff to `NACK` with reason `STALE_GENERATION` … a registration rejected by the t1074 live-owner rule SHALL leave the handoff unchanged」. 불변식 「While a handoff is non-final … the lane endpoint row SHALL equal the reserved source endpoint … any committed change to that row other than the handoff rebind SHALL finalize the handoff in the same transaction」.
- REQ-FLH-018(spec.md:137) 추가 문장 「This rejection SHALL NOT strand a lane restarted through the launcher: that launcher's provisional registration finalizes the handoff first (REQ-FLH-017)」가 두 REQ 를 명시적으로 연결한다.
- REQ-FLH-015(spec.md:125) 결합자 서술 「bound by the first legitimate non-empty UserPromptSubmit (mandatory) or an earlier order-independent SessionStart (best-effort) per … REQ-FMH-001」은 t1074 원문(`.moai/specs/SPEC-FACTORY-MIXED-HOOK-001/spec.md:57` 「When startup `SessionStart` runs after provisional registration, it MAY perform the same idempotent early bind, but correctness SHALL NOT depend on that ordering. When the first legitimate non-empty `UserPromptSubmit` is observed, the registry SHALL atomically rebind any remaining provisional lane」)과 일치한다(E7). design.md:69 표 3·4번 행, 도식(:82-95)도 맞춰졌다.

**다섯 쓰기 경로 대조 (착지 코드, HEAD `6ad4824a2`).** 비종결 handoff 동안 행을 바꾸면서 handoff 를 종결하지 않는 경로가 있는가:

| # | 경로 | 착지 동작 | 비종결 handoff 중 | 판정 |
|---|---|---|---|---|
| 1 | `RegisterLaunchPending` (store.go:381) → `RegisterPeer` (:386) | launch-pending 토큰으로 행 교체. 다른 identity 이고 기존 owner 가 current 면 거절(:354-355) | REQ-FLH-017: commit 하면 같은 tx 에서 NACK, 거절되면 handoff 불변 | 규정됨 |
| 2 | `RollbackLaunchPending` (:455, `DELETE … WHERE session_uuid=<token> AND generation=…`) | exact provisional identity 일 때만 삭제 | 행은 bound source tuple 이라 토큰 불일치 → 영향 행 0 | 불변식으로 커버(design.md:103) |
| 3 | `BindLaunchPending` (:392) | 행이 launch-pending 이 아니면 no-op(:419-421) | 행이 bound 라 no-op | 불변식으로 커버 |
| 4 | `RegisterPeer` from UserPromptSubmit (factory_messages.go:93) | 같은 owner·새 session 이면 gen+1 교체(:353-359) | REQ-FLH-018 로 거부 | 규정됨 |
| 5 | handoff rebind | 새 코드 | CAS + tombstone + receipt | 규정됨 |

RollbackLaunchPending 이 행을 **삭제**하는 경우도 design.md:103 「2번(rollback)과 3번(SessionStart bind)은 pending 행에만 작동하는데 비종결 handoff 동안 행은 bound source tuple이므로 일치할 수 없고」로 커버된다. launcher 등록이 handoff 를 종결한 **뒤** rollback 이 행을 지우면 lane 에 행이 없지만, 그때 handoff 는 이미 종결이고 REQ-FLH-017 의 종료 조건은 「once the t1074 binder has run」으로 걸려 있어 모순이 아니다(t1074 launch 실패 동작 그대로).

단서: 경로 2·3 의 커버는 「launch-pending 행 + 비종결 handoff」가 도달 불가라는 전제에 기대고, 그 전제의 admission 쪽은 reservation 이 source 행을 자기 write transaction 안에서 읽어 고정할 때만 성립한다. 이것이 N7 이다.

**AC-FLH-018 강제 순서 (iv)·(v)의 세대 산술 (착지 코드로 계산).** launcher 와 hook 모두 `Generation: 1` 을 넘긴다(factory_launch_pending.go:60, factory_messages.go 의 `want`). `RegisterPeer` 는 `p.Generation <= oldGen` 이면 `oldGen+1`(store.go:356-358, :346-348), `BindLaunchPending` 은 `current.Generation + 1`(:425).
- (iv): source bound `g` → SessionStart `BindLaunchPending` no-op(행이 pending 아님) → A-register: `oldSession`(src) ≠ 토큰, identity 다름, source 가 fake process-start 라 not current → commit, `g+1` → UserPromptSubmit `RegisterPeer`: 행은 토큰, identity 는 A 와 같음 → live-owner 검사 없이 `g+2`. **`g+2` 일치.**
- (v): A-register `g+1` → SessionStart `BindLaunchPending` 이 alias 결합 `g+2` → 실제 session 의 UserPromptSubmit, 같은 PID/start, 다른 session → `g+3`. **`g+3` 일치.**
- 착지 alias 테스트 `TestFactorySessionStartCannotRotateAuthoritativeUserPromptBinding`(internal/hook/factory_messages_test.go:256)은 HEAD 에 그대로 있고, 단언이 alias = pending+1, authoritative = alias+1 이다(E8). (v)가 같은 산술을 handoff 종결 뒤에 재현한다. 개정은 코드를 건드리지 않았다. 다만 AC-FLH-015 게이트는 `TestFactoryLaneHandoffT1074Compatibility` 하나만 선택하므로, 이 착지 테스트 자체의 보존은 CI 전체 스위트에 맡겨진다(관측 메모, 결함 아님).

### N2 REQ18-SAME-TX-UNDISCRIMINATED — **RESOLVED**

AC-FLH-019 (iv)-(vi)(acceptance.md:283-289)가 추가됐다. 질문 「read-before-BEGIN 구현이 여전히 통과할 수 있는가」의 답은 **아니오**다. 임시 프로브로 production `Open` 과 같은 DSN(`busy_timeout(2500)`, `journal_mode(WAL)`, `_txlock=immediate`, `SetMaxOpenConns(1)`)의 두 핸들을 열어 확인했다(E6).

```
PRE_BEGIN_READ count=0
H blocked: InUse=1 WaitCount=0 Open=1
IN_TX_READ count=1 vCommitBeforeHReturn=true
```

- V 가 미commit INSERT 를 쥔 동안 H 핸들의 `BEGIN` 전 판독은 0행을 본다. WAL 판독자는 막히지 않으므로 이 판독은 즉시 돌아온다. 따라서 read-before-BEGIN 구현은 H 가 막히기 **전에** 이미 「handoff 없음」을 결정한다.
- H 의 `BEGIN IMMEDIATE` 는 SQLite 잠금에서 막히고(`InUse=1`, `WaitCount=0`), V commit 뒤 transaction 안에서의 판독은 1행을 본다.
- 결과: read-inside 구현은 (iv)에서 `ENDPOINT_HANDOFF_PENDING` 으로 PASS, read-before-BEGIN 구현은 endpoint 를 옮겨 「byte-identical」 단언에서 FAIL 한다. (v)도 같은 구조라 launcher 쪽 판독 위치를 판별한다. (vi)는 판별력이 없다고 스스로 밝힌 대조군이다(:287).
- 경계 조건: 2500 ms 는 production `Open` 경로 값이다. `Open` → `OpenWithDeadline(…, 5*time.Second)`(store.go:122) → `busyMillis = deadline/2`(:207-212). AC 의 값과 일치한다.

barrier seam 을 「t1082가 추가한 코드 전체」로 넓힌 것과 N3 의 관계: 정합한다. (i)는 H 의 transaction 을 barrier 에서 쥐어야 하므로 seam 이 H 의 transaction **안**에 있어야 하고, 그 위치는 REQ-FLH-018 이 `RegisterPeer` 안에 추가하는 상태 판독이다(:291 「the handoff-state check that REQ-FLH-018 adds inside the registration transaction, which forced order (i) uses」). seam 이 `BEGIN` 뒤에 있어야 (i)가 성립하므로, seam 위치 자체가 「판독은 transaction 안」이라는 요구와 같은 방향을 가리킨다. 별도 test 전용 등록 경로를 두지 말라는 :18 문구와도 충돌하지 않는다. seam 은 production 함수 안의 주입점이지 대체 경로가 아니다. 다만 「Code landed by t1074 carries no seam」과 「seam 이 t1074 함수 `RegisterPeer` 본문 안에 들어간다」는 문장이 나란히 있어 처음 읽을 때 걸린다. 「t1074 가 착지한 줄에는 seam 이 없다」는 뜻으로 읽으면 일관된다(결함으로 올리지 않음).

### N3 AC19-DIRECT-REGISTRATION-POLICY — **RESOLVED**

acceptance.md:18 이 「production 경로를 우회한 fixture 준비(직접 peer 등록, 수동 DB seed)」로 좁혀졌고, 피시험 store entry point 를 명시적으로 제외한다. :291 이 같은 내용을 AC-FLH-019 쪽에서 반복한다.

### N4 HANDLE-PROOF-OPENCONN-VACUOUS — **RESOLVED**

acceptance.md:266 증명 (2)가 「`db.Stats().InUse == 1` and `db.Stats().WaitCount == 0`」으로 바뀌었고 :291 이 모든 racer 쌍(H/R, V/H, V/A)에 적용한다. E6 에서 SQLite 잠금 대기 핸들이 정확히 이 값을 보고함을 확인했다. 잔여 약점: `WaitCount` 는 핸들 생성 이후 누적값이므로 같은 핸들을 앞선 순서에서 동시 사용했다면 0 이 아닐 수 있다. racer 핸들이 goroutine 하나에서만 쓰이면 문제되지 않는다(결함으로 올리지 않음).

### N5 STALE-SIBLING-COUNTS — **RESOLVED**

- progress.md:41 「20개 t1082 named tests(AC 19 + 공통 gate-quality 1 …) → 20」. 재실행 결과 `20`(E1).
- design.md:101 「순서 규칙은 네 가지다」 뒤에 첫째~넷째 네 개.
- 잔여 사소: progress.md frontmatter `updated: 2026-09-22` 인데 본문은 2026-09-23 개정에서 바뀌었다(N-trivial, 아래 N12).

### N6 POST-NACK-FRESH-RESERVATION-ADMISSION — **RESOLVED (사소한 잔여 N10)**

REQ-FLH-003(spec.md:77) 둘째 문장, design.md:214 §9 새 행, AC-FLH-001(acceptance.md:80) 확장. AC 에는 「target HEAD moved off the fresh pin returns `BASE_DRIFT`」 음성 대조가 있고, materializer 호출 수 0 을 단언한다. 잔여: 조건이 「clean, reserved `WT-*` branch, fresh develop pin」 세 가지인데 실패 사유는 `BASE_DRIFT`/`BRANCH_COLLISION` 두 개뿐이라 dirty target 의 사유가 정해지지 않았다(N10). 또 develop 이 그 사이 전진했다면 fresh pin 과 target HEAD 가 달라 매번 `BASE_DRIFT` 가 되고, 탈출은 §9 의 operator `ABANDONED`(REQ-FLH-011, 자동 삭제 금지)뿐이다. broker 교착은 아니므로 결함으로 올리지 않았다.

### 리드 전달 — Idempotency — **RESOLVED**

- design.md:204, REQ-FLH-009(spec.md:101), plan.md:58 세 곳이 같은 말을 한다. 키는 현행 t1074 스키마를 바꾸지 않고, handoff generation 은 키가 아니며 stale NACK 판정에만 쓰고, 키 기준 결정은 t1100 소관이다.
- 착지 스키마는 `messages(… idem_key …, UNIQUE(sender_session,idem_key))`(store.go:282)이다. lead→lane dispatch 의 `sender_session` 은 lead 쪽이라 lane BOUND 전후에 변하지 않으므로, AC-FLH-008(acceptance.md:150) 「identical idempotency keys before/after BOUND」와 일치한다.
- schema 요구 추가 여부: `idempoten` grep 결과(E9)에 새 컬럼·키 요구가 없다. design.md:52 의 envelope 필드 `handoff_generation` 은 payload 필드이지 key 가 아니다. 이전 문구 `(run,lane,card,handoff_generation,idempotency_key)` 는 plan.md 에서 사라졌다.

## Defects Found

D-N7. RESERVATION-SIDE-READ-UNDISCRIMINATED — design.md:116, design.md:103, acceptance.md:283-289 (AC-FLH-019), acceptance.md:244 (AC-FLH-017) — N1 을 닫은 불변식에서 「launch-pending 행 + 비종결 handoff」가 도달 불가라는 결론의 admission 쪽(design.md:103 「admission 쪽은 REQ-FLH-016이 … 막는다」)은 reservation 이 source 행 판독, REQ-FLH-016 판정, `RESERVED` 삽입을 하나의 write transaction 에서 할 때만 성립한다. design.md:116 의 「각각 한 transaction 안에 있어야」 목록은 rebind·launcher 등록·UserPromptSubmit 판독만 들고 reservation 은 빠졌다. 규범 근거는 있다 — REQ-FLH-017 불변식, plan.md:40 「CAS로 고정」, design.md:125 「admission + CAS reservation」. 그러나 어느 AC 도 판별하지 않는다: AC-FLH-019 (v)는 V 가 쥐고 A 가 기다리는 방향만 검사하고, 반대 방향(A 가 미commit 등록을 쥔 동안 V 가 `BEGIN` 전에 source 행을 읽는 순서)은 없다. read-before-BEGIN reservation 은 bound 행을 읽은 뒤 A 의 등록 commit 을 기다렸다가 오래된 tuple 로 `RESERVED` 를 commit 한다. 그러면 경로 3(`BindLaunchPending`)이 handoff 를 종결하지 않고 행을 바꿀 수 있고, SessionStart 가 먼저 오지 않으면 첫 UserPromptSubmit 이 REQ-FLH-018 로 거부되어 N1 의 고립이 더 좁은 순서로 재발한다. — Severity: major — Class: optional (SHOULD-FIX; 규범은 옳고 판별 AC 만 없음 — iter-5 N2 와 같은 분류) — Required fix: design.md:116 목록에 「reservation 의 source 행 판독·REQ-FLH-016 판정·`RESERVED` 삽입」을 추가하고, AC-FLH-019 에 순서 (vii)을 넣는다. A(launcher 등록)가 REQ-FLH-017 이 `RegisterPeer` 안에 추가하는 t1082 seam 에서 미commit 등록을 쥔 동안 V 가 별도 핸들에서 막힘(`InUse==1`, `WaitCount==0`)을 관측하고, A 가 commit 한 뒤 V 는 `ENDPOINT_LAUNCH_PENDING` NACK, reservation 0 이어야 PASS. run 단계에서 채무로 이월해도 된다.

D-N8. REQ17-UNCONDITIONAL-LATER-LAUNCHER — spec.md:133 (REQ-FLH-017) — 「When the rebind commits first, a later launcher registration or bind SHALL leave the handoff-bound endpoint unchanged」. 0.5.0 에서 「bind」가 「registration or bind」로 넓어져 무조건 문장이 됐다. t1074 착지 동작은 handoff-bound owner 가 current 가 아니면 relaunch 등록이 행을 덮는다(store.go:354-358). AC-FLH-018 Given(acceptance.md:256)도 괄호 안에서 「with the same identity as the source, or a non-current new owner, the launcher overwrites the row as launch-pending instead」라고 그 분기를 인정한다. 문자 그대로 구현하면 BOUND 이후 lane 이 죽었을 때의 정상 relaunch 를 막아 REQ-FLH-015(t1074 재사용)와 부딪힌다. — Severity: minor — Class: optional — Required fix: 문장 끝에 「while the handoff-bound owner is current under the t1074 live-owner rule」을 붙인다.

D-N9. AC18-19-FIXTURE-WORDING — acceptance.md:256, :258, :274 Given, :286 — (a) AC-FLH-018 Given 은 source owner 를 not current 로, 새 endpoint owner 를 current 로 「injected」한다고 쓴다. 그런데 (iv)·(v)는 `registerFactorySessionStartPeer`/`registerFactoryUserPromptPeer`(package `hook` 비공개 함수)를 거쳐야 하므로 test 는 package `hook` 안에 있어야 한다. 그 위치에서는 `Store.ownerCurrent`(비공개 필드, 유일한 주입 사례는 internal/factorymsg/store_test.go:489)를 주입할 수 없고, package `factorymsg` 내부 test 는 import cycle 때문에 `hook` 을 부를 수 없다. 실제로는 실행 중인 identity 와 fake process-start 로 배치하는 방법만 가능하다(:258 이 source 쪽은 그렇게 적었다). (b) :258 「registering the source row with a fake process-start」가 어느 경로로 등록하는지 적지 않아, 직접 `RegisterPeer` seed 로 읽히면 :18 에 따라 FAIL 이다(AC-FLH-019 는 production launcher 경로를 명시). (c) AC-FLH-019 Given 은 source 를 「live PID `p`」로 두고 (iv)-(vi)는 「the same bound source row」에서 시작한다고 하는데, (v)는 「source owner not current」를 요구한다. — Severity: minor — Class: optional — Required fix: 「injected」를 「arranged (by injection in `factorymsg` or by real/fake process identity in `hook`)」로 바꾸고, source seed 는 production launcher 경로(`RegisterLaunchPending`+`BindLaunchPending`, fake process-start 허용)로 한다고 명시하며, (v)에는 별도 fixture(또는 A 가 source 와 같은 identity)를 적는다.

D-N10. REQ3-DIRTY-TARGET-REASON — spec.md:77 (REQ-FLH-003 둘째 문장), design.md:214 — 진입 조건은 clean·reserved branch·fresh pin 세 가지인데 실패 사유는 `BASE_DRIFT`/`BRANCH_COLLISION` 둘뿐이다. dirty target 은 둘 중 어느 쪽에도 의미상 맞지 않아 첫 문장의 「reason-specific NACK」 원칙과 어긋난다. — Severity: minor — Class: optional — Required fix: dirty target 사유(예: 첫 문장 목록의 dirty checkout 사유 재사용)를 한 단어로 지정한다.

D-N11. PLAN-OMITS-REQ16-18-WORK — plan.md:37-73 (M1-M5), plan.md:78-86 (file ownership) — 마일스톤과 파일 소유 표에 REQ-FLH-016(launch-pending NACK), REQ-FLH-017(착지된 t1074 `RegisterPeer`/`RegisterLaunchPending` transaction 에 handoff 종결 추가), REQ-FLH-018(UserPromptSubmit 거부)의 작업 항목이 없다. `internal/hook/` 행은 「SessionStart evidence 전달」만 적는다. 0.5.0 으로 t1074 착지 코드 수정이 명시적 요구가 됐으므로 누락의 무게가 커졌다. 이번 개정이 만든 것은 아니다(0.3.0 이후 누락). plan.md:92 「각 AC의 exact named test를 먼저 작성」이 AC-FLH-017..019 를 통해 작업을 끌어오므로 구현 누락 위험은 낮다. — Severity: minor — Class: optional — Required fix: M3 또는 M4 에 세 REQ 의 작업 한 줄씩, 파일 소유 표의 `internal/factorymsg/`·`internal/hook/` 행에 「t1074 `RegisterPeer` transaction 안 handoff 판독·종결·거부」를 추가한다.

D-N12. PROGRESS-UPDATED-DATE — progress.md:5 `updated: 2026-09-22` — 본문(:41)은 2026-09-23 개정에서 바뀌었다. — Severity: minor — Class: optional — Required fix: `updated: 2026-09-23`.

## Part B

해당 없음 — 판정이 임계 이상(PASS-WITH-DEBT)이다. 채무 처리 권고: N7 은 run 단계 첫 RED 작성 때 AC-FLH-019 (vii)로 흡수하는 것이 가장 싸다(named test 는 그대로, 순서 하나 추가). N8·N9·N10·N12 는 한 줄씩이다. N11 은 run 착수 전 plan 한 줄 보강으로 충분하다.

## Part C — 범위 축소 입력

기준: CORE 는 「안전한 최소 lane worktree handoff」에 필요한 것. DETACHABLE 은 떼어 후속 카드로 옮겨도 남은 SPEC 이 안전하지 않거나 모순되지 않는 것. factory lane 은 대화형 Codex 세션이므로 최소 형태는 interactive 경로라고 보았다.

| REQ | 분류 | 근거 한 줄 |
|---|---|---|
| REQ-FLH-001 stable lane / replaceable endpoint | CORE | 주소 모델 자체. 모든 상태·receipt 범위가 여기에 걸린다. |
| REQ-FLH-002 interactive/headless 분리 | CORE (분할) | interactive 전이는 CORE. headless 절은 REQ-FLH-007 과 함께 뗄 수 있다 — 떼면 「headless 는 범위 밖」 한 줄로 바꾼다. |
| REQ-FLH-003 fail-closed admission | CORE (분할) | 첫 문장은 안전 핵심. 둘째 문장(NACK 뒤 재구성 진입)은 DETACHABLE — 없으면 재진입이 conflicting target 으로 NACK 되어 operator 정리가 필요하지만 fail-closed 라 안전하다. |
| REQ-FLH-004 develop pin / base drift | CORE | 실제 관측된 사고(main→develop drift)를 막는 조항. |
| REQ-FLH-005 traceability | CORE | branch rename·collision 검증 전 `SWITCH_PENDING` 금지가 안전 조건이다. envelope 필드는 싸다. |
| REQ-FLH-006 interactive relocation boundary | CORE | 최소 형태의 relocation 경로 그 자체. |
| REQ-FLH-007 headless official relocation | DETACHABLE | 떼도 interactive 경로의 안전 조건은 변하지 않는다. app-server fork/start 계약이 통째로 빠진다. |
| REQ-FLH-008 atomic validated rebind | CORE (분할) | 원자 교체·tombstone·BOUND receipt 는 핵심. headless evidence 절은 REQ-FLH-007 과 함께 뗄 수 있다. |
| REQ-FLH-009 dispatch ordering / idempotency | CORE | BOUND 전 body 거부와 1회 release 가 무쓰기 안전의 집행점이다. |
| REQ-FLH-010 tombstone / stale rejection | CORE | 옛 endpoint 가 계속 ACK·송신하는 것을 막는다. |
| REQ-FLH-011 crash / abandoned recovery | CORE | 자동 삭제 금지와 live owner 추측 금지는 데이터 보존 조건이다(AC-FLH-009 failpoint 행렬의 폭은 줄일 여지가 있다). |
| REQ-FLH-012 authority / product-boundary truth | DETACHABLE | 안전 관련 금지(slash 자동화·tmux·private socket)는 REQ-FLH-006 이, launcher 관리 worktree 는 REQ-FLH-004 가 이미 싣는다. 남는 것은 설명의 정확성이다. |
| REQ-FLH-013 zero-write before BOUND | CORE | 이 기능의 존재 이유. |
| REQ-FLH-014 LIVE cross-harness proof | DETACHABLE | 검증 층이다. 떼도 SPEC 은 모순되지 않지만 interactive `/cd`→SessionStart 경계의 실측 증거가 사라진다. interactive Codex↔Codex 한 행만 남기는 부분 분리가 가능하다. headless LIVE 행은 REQ-FLH-007 을 떼면 함께 빠져야 한다. |
| REQ-FLH-015 reuse / compatibility | CORE | 새 broker·store 금지와 t1074 결합자 계약. REQ-FLH-017/018 의 전제. |
| REQ-FLH-016 launch-pending admission NACK | CORE | N1 불변식의 admission 쪽. 떼면 reservation 이 provisional 행에 대해 생겨 불변식이 깨진다. |
| REQ-FLH-017 launcher vs rebind 직렬화 | CORE | lane 재기동 경합의 안전성과 N1 폐쇄의 등록 쪽. |
| REQ-FLH-018 UserPromptSubmit 거부 | CORE | interactive `/cd` 뒤 turn 의 UserPromptSubmit 이 tombstone·receipt 없이 endpoint 를 옮기는 것을 막는다. interactive 경로에 필수다. |

**떼어낼 수 있는 집합**: REQ-FLH-007, REQ-FLH-012, REQ-FLH-014(전부 또는 headless 행), 그리고 REQ-FLH-002·008 의 headless 절과 REQ-FLH-003 둘째 문장. 함께 빠지는 AC: AC-FLH-004, AC-FLH-013, 공통 gate-quality test(AC-FLH-012 를 남기면 유지), AC-FLH-014·005 의 headless 부분, AC-FLH-001 의 재구성 확장.

**떼면 사라지는 결함**: N10(REQ-FLH-003 둘째 문장이 빠지므로), design.md:114 잔여 위험 「NACK 뒤 고아 headless fork thread」. **남는 결함**: N7·N8·N9·N11 은 모두 CORE 인 REQ-FLH-016..018 경합 묶음과 그 plan 에 있어 범위를 줄여도 그대로다. N12 도 남는다. 즉 범위 축소는 남은 채무를 거의 줄이지 않는다. 경합 묶음은 떼면 안전성이 깨지는 CORE 이기 때문이다.

**의존 간선 (자르는 선을 제약하는 것)**:

- REQ-FLH-017 ↔ REQ-FLH-018: 018 의 「고립시키지 않는다」 문장은 017 의 종결에 기대고, 017 의 불변식은 018 이 경로 4 를 거부해야 성립한다. 둘은 함께만 움직인다.
- REQ-FLH-016 → REQ-FLH-017: 불변식의 admission 쪽. 016 없이 017 은 참이 아니다.
- REQ-FLH-015 → REQ-FLH-017/018: t1074 결합자(UserPromptSubmit 필수) 서술이 두 REQ 의 전제다.
- REQ-FLH-006, REQ-FLH-007 → REQ-FLH-008: mode 별 evidence. 007 을 떼면 008 의 headless 절도 떼야 한다.
- REQ-FLH-008 → REQ-FLH-009, 010, 013, 011: BOUND receipt 가 release·tombstone·무쓰기 해제·idempotent finalize 의 기준점이다.
- REQ-FLH-002 → REQ-FLH-006, 007: 분리 원칙. 007 을 떼면 002 를 interactive 단독으로 다시 써야 한다.
- REQ-FLH-014 → REQ-FLH-006, 007, 013: LIVE 행이 두 mode 를 실행한다. 007 을 떼면 014 의 headless 행도 반드시 빠진다.
- REQ-FLH-003 둘째 문장 → REQ-FLH-004(fresh pin), REQ-FLH-011(§9 재구성).
- REQ-FLH-012 ⊂ REQ-FLH-006 ∪ REQ-FLH-004 (중복 — 떼도 빈틈 없음).

## Evidence

E0. 트리 확인
```
$ git rev-parse --show-toplevel; git branch --show-current; git rev-parse --short HEAD; git status --short
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1082
WT-factory-lane-worktree-handoff
6ad4824a2
(출력 없음)
```

E1. heading 과 named test 수
```
$ grep -E '^unset ' .moai/specs/SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001/acceptance.md | grep -oE 'Test[A-Za-z]+' | sort -u | wc -l
      20
$ grep -cE '^### REQ-FLH-[0-9]+' spec.md
18
$ grep -cE '^### AC-FLH-[0-9]+' acceptance.md
19
$ grep -oE '^### REQ-FLH-[0-9]+' spec.md | tr '\n' ' '
### REQ-FLH-001 … ### REQ-FLH-018   (연속, 공백 없음)
$ grep -oE '^### AC-FLH-[0-9]+' acceptance.md | tr '\n' ' '
### AC-FLH-001 … ### AC-FLH-019   (연속, 공백 없음)
```

E2. strict lint
```
$ moai spec lint SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001 --strict --json; echo lint_exit=$?
[]
lint_exit=0
```

E3. RED 재확인 (HEAD `6ad4824a2`)
```
$ rg -n -F 'func TestFactoryLaneHandoffRebindVsLaunchBindRace(' internal --glob '*_test.go'; echo exit18=$?
exit18=1
$ rg -n -F 'func TestFactoryLaneHandoffRebindVsUserPromptRegisterRace(' internal --glob '*_test.go'; echo exit19=$?
exit19=1
```

E4. frontmatter·MP-6·MP-7
```
$ grep -rn 'NEEDS CLARIFICATION' plan.md research.md; echo mp7_exit=$?
mp7_exit=1
$ grep -c syscall spec.md
0
$ head -20 spec.md
id: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001 / title: "Factory lane card worktree handoff" / version: "0.5.0" / status: draft / created: 2026-09-22 / updated: 2026-09-23 / author: manager-spec / priority: P1 / phase: "v3.0.0" / module: "internal/factorymsg" / lifecycle: spec-anchored / tags: "factory,lane,worktree,codex,app-server,rebind" / tier: L
```

E5. D7
```
$ grep -oE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u
SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
SPEC-FACTORY-MIXED-HOOK-001
$ grep -n '^status:\|^version:' .moai/specs/SPEC-FACTORY-MIXED-HOOK-001/spec.md
4:version: "0.1.4"
5:status: completed
```

E6. 임시 프로브 `internal/factorymsg/zz_auditprobe_t1082_iter6_test.go` (실행 후 삭제). production `OpenWithDeadline` 과 같은 DSN 파라미터로 한 DB 파일에 핸들 두 개. V 가 `BEGIN IMMEDIATE` 안에서 INSERT 후 미commit, H 가 `BEGIN` 전 판독 → 별도 goroutine 에서 `BEGIN` + 판독, 300 ms 뒤 통계 관측, V commit.
```
$ unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=<scratchpad>/gc go test ./internal/factorymsg -run '^TestZZAuditProbeT1082Iter6$' -count=1 -v
=== RUN   TestZZAuditProbeT1082Iter6
    zz_auditprobe_t1082_iter6_test.go:48: PRE_BEGIN_READ count=0
    zz_auditprobe_t1082_iter6_test.go:70: H blocked: InUse=1 WaitCount=0 Open=1
    zz_auditprobe_t1082_iter6_test.go:77: IN_TX_READ count=1 vCommitBeforeHReturn=true
--- PASS: TestZZAuditProbeT1082Iter6 (0.35s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/factorymsg	0.627s
$ rm internal/factorymsg/zz_auditprobe_t1082_iter6_test.go
$ git status --short
(출력 없음)
```

E7. t1074 결합자 원문
```
$ sed -n '57p' .moai/specs/SPEC-FACTORY-MIXED-HOOK-001/spec.md
… When startup `SessionStart` runs after provisional registration, it MAY perform the same idempotent early bind, but correctness SHALL NOT depend on that ordering. When the first legitimate non-empty `UserPromptSubmit` is observed, the registry SHALL atomically rebind any remaining provisional lane to the actual session UUID and resolved owner PID/process-start as `bound` before that hook processes the inbox. …
```

E8. 착지 코드 좌표 (HEAD `6ad4824a2`)
```
store.go:122   return OpenWithDeadline(projectRoot, runID, 5*time.Second)
store.go:207-214 busyMillis := deadline.Milliseconds() / 2 … busy_timeout, journal_mode(WAL), _txlock=immediate
store.go:354-358 if (oldPID != p.PID || oldStart != p.ProcessStart) && s.ownerCurrent(oldPID, oldStart) { … "factory logical lane has a live owner" } if p.Generation <= oldGen { p.Generation = oldGen + 1 }
store.go:386   return s.RegisterPeer(ctx, p)
store.go:419-421 if !isLaunchPendingSession(current.SessionUUID) { return current, false, nil }
store.go:425   bound.Generation = current.Generation + 1
store.go:464   DELETE FROM peers WHERE slot=? AND … session_uuid=? AND generation=? AND pid=? AND process_start=?
internal/cli/factory_launch_pending.go:60  Role: role, Slot: slot, Generation: 1, PID: pid, ProcessStart: processStart,
internal/hook/factory_messages.go:84  p, bound, bindErr := s.BindLaunchPending(ctx, want)
internal/hook/factory_messages.go:93  p, err := s.RegisterPeer(ctx, want)
internal/hook/factory_messages_test.go:256 func TestFactorySessionStartCannotRotateAuthoritativeUserPromptBinding(t *testing.T) {
  … afterAlias.Generation != pending.Generation+1 … authoritative.Generation != afterAlias.Generation+1 …
$ grep -rn 'ownerCurrent' internal --include='*.go' | grep -v 'store.go'
internal/factorymsg/store_test.go:489:	s.ownerCurrent = func(pid int, start string) bool { return pid == os.Getpid() && start == currentStart }
```

E9. idempotency 문구와 t1074 스키마
```
$ grep -n -i 'idempoten' .moai/specs/SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001/*.md
plan.md:58: … 현행 t1074 스키마의 idempotency key를 바꾸지 않고 멱등 처리한다. handoff generation은 key에 넣지 않고 stale-generation NACK 판정에만 쓴다 …
design.md:204: … 현행 t1074 스키마의 idempotency key를 바꾸지 않고 그대로 써서 판정한다. … 이 SPEC은 schema 요구를 추가하지 않는다.
acceptance.md:150: **Given** identical idempotency keys before/after BOUND …
spec.md:101: … detecting duplicates by the current t1074 schema's idempotency key unchanged. The handoff generation SHALL NOT be part of that key …
(그 밖: design.md:219, acceptance.md:160, spec.md:26/99/109 — 재시도·finalize 문맥, key 정의 아님)
$ grep -n 'UNIQUE' internal/factorymsg/store.go
282: CREATE TABLE IF NOT EXISTS messages(… idem_key TEXT NOT NULL, … UNIQUE(sender_session,idem_key));
```

## Baseline-attribution

- 트리 `.claude/worktrees/t1082`, 브랜치 `WT-factory-lane-worktree-handoff`, HEAD `6ad4824a2`. 감사 시작·종료 시 `git status --short` 비어 있음(E0, E6).
- 코드 좌표는 이 HEAD 의 `internal/factorymsg/store.go`, `internal/hook/factory_messages.go`, `internal/hook/factory_messages_test.go`, `internal/cli/factory_launch_pending.go` 에서 직접 읽었다.
- t1074 텍스트는 이 트리의 `.moai/specs/SPEC-FACTORY-MIXED-HOOK-001/spec.md` 에서 읽었다.
- 프로브 빌드 캐시는 세션 scratchpad 에 두었다.

## Gaps

- `audit_multi` / `codex_audit` / `glm_audit` 교차 모델 의견은 호출하지 않았다(Claude 단독 델타 감사).
- E6 프로브는 production `Open` 이 아니라 같은 DSN 파라미터로 연 raw 핸들 위의 두-테이블 모형이다. `RegisterPeer` 자체나 handoff 테이블(아직 없음)로 한 종단 재현이 아니다. `-race` 로는 돌리지 않았다.
- N1 의 다섯 경로 판정은 코드 읽기와 세대 산술이다. REQ-FLH-017/018 이 구현되지 않았으므로 AC-FLH-018 (iv)·(v)를 실행하지 않았다.
- N7 의 재발 순서는 read-before-BEGIN reservation 을 가정한 논증이며, 그런 구현이 존재하지 않으므로 관측하지 않았다.
- N9(a)의 import cycle 주장은 `internal/hook` 이 `factorymsg` 를 import 한다는 사실(factory_messages.go 의 `factorymsg.Open`)과 Go 규칙에서 나온 추론이다. 실제로 컴파일해 보지 않았다.
- `session.ResolveOwnerPID` 가 `go test` 프로세스 안에서 무엇으로 해석되는지는 확인하지 않았다(AC-FLH-018 (iv)·(v) fixture 의 배치 가능성에 영향).
- 개정이 건드리지 않은 AC 본문(AC-FLH-002..007, 009..017)은 재감사하지 않았다. plan.md 는 N11 판단에 필요한 마일스톤·파일 소유 절만 다시 읽었다.

## Residual-risk

- 점수 여유가 0.0125 다. Clarity 를 0.75 로, Completeness·Testability 를 0.85 로 매겼는데, 둘 중 하나가 0.75 로 내려가면 평균은 0.8375 로 FAIL 이 된다. 판정은 차단 결함 0건과 필수 통과 7/7 에 기대고 있으며, 점수 자체는 경계에 있다.
- N7 이 run 단계에서 흡수되지 않으면, reservation 을 판독-후-트랜잭션으로 구현한 코드가 모든 AC 를 통과한 채 N1 계열 고립을 좁은 순서로 재도입할 수 있다.
- 결정 ②의 잔여 위험 두 가지 — launcher 밖 재기동(F②-3)과 NACK 뒤 고아 headless fork thread — 는 design.md:113-114 에 기록돼 있고 어느 AC 도 다루지 않는다. 이는 결정 기록의 선택이므로 결함으로 올리지 않았다.
- REQ-FLH-018 의 「묘비 UUID `STALE_ENDPOINT`」는 BOUND 뒤에도 영구적이다(iter-5 에서 지적된 잔여, 이번 개정 범위 밖).
