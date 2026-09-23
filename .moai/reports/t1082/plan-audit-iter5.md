# SPEC 감사 보고서: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001

Iteration: 5 (DELTA — iter-4 결함 D1-D5 대비, 운영자가 반복 상한 초과를 명시 승인)
Revision: `c29d35ea7`
Verdict: **FAIL**
Overall Score: 0.81 (Tier L 임계 0.85 미달, iter-4 0.81 과 동일 — 회귀 아님, STOP 신호 없음)

작성자 추론 맥락은 M1 Context Isolation 에 따라 배제했다. 판단 근거는 SPEC 산출물, 착지 코드, 그리고 이번 감사에서 직접 실행한 명령 출력뿐이다.

## Claim

개정 `c29d35ea7` 은 iter-4 의 D1(문서·AC 측면), D2, D3, D5 를 닫았다. 그러나 D1 을 닫으려고 새로 넣은 REQ-FLH-018 이 새 차단 결함 하나를 만들었다. t1074 착지 계약에서 UserPromptSubmit `RegisterPeer` 는 launch-pending 행을 bound 로 바꾸는 **필수** 결합 경로이고(t1074 spec.md:57, SessionStart 는 best-effort), REQ-FLH-018 은 비종결 handoff 동안 이 경로를 예외 없이 막는다. 그 결과 REQ-FLH-017 이 다루는 「handoff 비종결 중 launcher 재기동」 순서에서 SessionStart 가 launcher 등록보다 먼저 발화하면, 재기동된 lane 은 handoff 가 종결될 때까지 launch-pending 으로 남고 이를 풀 결합 경로가 없다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성 (요구사항 층): `grep -nE '^### REQ-FLH-[0-9]+' spec.md` → REQ-FLH-001(L66) … REQ-FLH-017(L130), REQ-FLH-018(L134), 18건 연속·중복 없음. AC 도 AC-FLH-001(acceptance.md:78) … AC-FLH-019(:266), 19건 연속.
- [PASS] MP-2 GEARS 준수 (요구사항 층에서 판정): REQ-FLH-018(spec.md:136) 은 State-driven 「While a lane has a handoff in a non-final state …, the t1074 UserPromptSubmit peer registration … SHALL be rejected …」에 When 절이 이어지는 compound 형이다. AC-FLH-019 의 Given-When-Then 은 검증 층 형식이며 MP-2 대상이 아니다.
- [PASS] MP-3 frontmatter: 개정은 `version: "0.3.0"` → `"0.4.0"` 만 바꿨다. 12 필드 + `tier: L` 존재(spec.md:1-17).
- [N/A] MP-4 언어 중립성: Go 내부 패키지 전용 SPEC, 템플릿 대상 아님.
- [PASS] MP-5 D7: spec.md 참조 SPEC 은 자기 자신과 `SPEC-FACTORY-MIXED-HOOK-001`(status `completed`) 뿐. retired/superseded/archived 없음.
- [PASS] MP-6 D8: `grep -c syscall spec.md` → 0.
- [PASS] MP-7 clarification gate: `grep -rn 'NEEDS CLARIFICATION' plan.md research.md` → 출력 없음, exit 1.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | design.md:99 「순서 규칙은 세 가지다」 뒤에 규칙 네 개(첫째~넷째). AC-FLH-019 의 H 직접 `RegisterPeer` 호출과 acceptance.md:18 「direct peer registration … FAIL」 정책의 긴장(N3). |
| Completeness | 0.75 | 0.75 | 쓰기 경로 목록(design.md:69-75)은 착지 코드와 전수 일치. 그러나 4번 행이 이 경로가 t1074 의 필수 launch-pending 결합자라는 역할을 빠뜨렸고, 그 누락이 N1 을 낳았다. |
| Testability | 0.75 | 0.75 | AC-FLH-019 강제 순서 셋은 t1074 현행 동작을 잡아내 비공허하다. 다만 separate-handle 증명 (2)의 `OpenConnections >= 1` 은 `Open` 직후 이미 참이라 공허하고(H1), REQ-FLH-018 의 「같은 transaction 안에서 읽기」 조항은 어느 순서로도 판별되지 않는다(N2). |
| Traceability | 1.0 | 1.0 | spec.md:159 `§ REQ-FLH-018 → AC-FLH-019`, acceptance.md:45 요약표, RED ledger 행, 정책 :15 `AC-FLH-001..019`. 번호 연속, heading 형식 `### REQ-FLH-NNN — <title>` 유지. |

## iter-4 결함 폐쇄 판정 (Regression Check)

- **D1 M2-THIRD-WRITER — 부분 폐쇄(문서·AC 요구는 충족, 새 차단 결함 N1 유발).**
  - (a) 쓰기 경로 목록: design.md:69-75 표가 비-테스트 호출부 전수와 일치한다(E2). `DELETE/UPDATE/INSERT … peers` 원시 SQL 도 store.go 의 세 지점(:367 RegisterPeer, :429 BindLaunchPending, :464 RollbackLaunchPending)뿐이다. 「둘로 고정」 문장은 삭제됐다.
  - (b) 동작 규정: 옵션 (a) 「비종결 동안 같은 transaction 안에서 거부」를 REQ-FLH-018(spec.md:136)로 규정하고 근거를 design.md:101 에 적었다. REQ-FLH-008(검증 후 한 transaction 만 교체)·REQ-FLH-010(stale 거부)·REQ-FLH-013(BOUND 전 무쓰기)과는 정합한다. REQ-FLH-016 과도 충돌하지 않는다(launch-pending 이면 reservation 자체가 없음).
  - **그러나 REQ-FLH-017 과 충돌한다(N1).** REQ-FLH-017 은 「in every interleaving … no orphan launch-pending endpoint」를 요구하고, REQ-FLH-015 는 t1074 의 결합 수명주기 재사용을 요구한다. t1074 에서 launch-pending 행의 필수 결합자는 UserPromptSubmit 이다(아래 N1).
  - (c) AC: AC-FLH-019(acceptance.md:266-281)가 rebind 전/후 커밋 두 순서와 검증 실패 후 H 순서를 강제 interleaving 으로 넣었다. 비공허성은 확인했다 — t1074 현행 `RegisterPeer` 는 (i)에서 거절 없이 행을 옮기고(probe B1: `got.session=post-cd-uuid got.gen=3`), (ii)의 src-uuid 호출도 거절 없이 되가져간다(probe B2: `got.session=src-uuid got.gen=4`). 따라서 REQ-FLH-018 미구현은 (i)·(ii)에서 FAIL 한다. 「항상 거부」 구현은 (iii)에서 FAIL 한다. 200회 비강제 반복은 end state 가 (i)/(ii) 뿐이라는 단언으로, H 가 먼저 옮기면 R 이 STALE_GENERATION NACK 이 되어 FAIL 한다.
  - 복구 경로: NACK/ABANDONED 뒤 UserPromptSubmit 이 t1074 의미로 복귀하므로 broker 수준 교착은 없다. BOUND 뒤에는 slot PRIMARY KEY(행 하나)라 중복 endpoint 가 구조적으로 불가능하고, 같은 identity 재호출은 store.go:351 `p.Generation = oldGen` 경로로 세대 불변, 묘비 UUID 는 STALE_ENDPOINT 로 거부된다(REQ 요구). 다만 「비종결 handoff 가 종결로 가는 사건」이 없는 순서가 N1 이다.
- **D2 AC18-SINGLE-HANDLE — 폐쇄.** acceptance.md:256 Given 에 「separate production `Open` call … distinct `*Store` and distinct underlying `*sql.DB`」, :258 에 공유 핸들 FAIL 조건과 3단 증명, jq 게이트에 `RACER_HANDLES_DISTINCT=2` 가 들어갔다. 질문 「Go 풀에서만 직렬화되면서도 통과할 수 있는가」의 답은 **아니오**다 — (1) `*sql.DB` 포인터 부등은 두 풀이 따로임을 확정하고, 따로인 두 풀 사이의 직렬화는 SQLite 잠금 말고는 없다(H2: `s2 blocked; s2.Open=1 s2.InUse=1 s2.WaitCount=0`). 단 (2)의 `OpenConnections >= 1` 부분은 증거 가치가 없다: `Open` 직후 경쟁자 실행 전에도 두 핸들 모두 `Open=1`(H1)이고, 공유 핸들에서도 `Open=1`(H3)이다. 풀 대기와 SQLite 대기를 가르는 값은 `WaitCount`(공유 1 vs 분리 0)와 `InUse` 다. 잔여 약점으로 N4(MINOR) 기록.
- **D3 AC18-OWNER-FIXTURE — 폐쇄.** acceptance.md:256 Given 에 「new endpoint owner … identity different from the source and injected as current (the only fixture under which the launcher-loses branch reaches the t1074 live-owner rejection …)」. store.go:354-355 의 분기 조건 `(oldPID != p.PID || oldStart != p.ProcessStart) && s.ownerCurrent(oldPID, oldStart)` 와 일치한다.
- **D4 — 해당 없음(iter-4 에서 결함 아님 판정, 조치 불요).**
- **D5 PROGRESS-COUNT — 폐쇄(인접 줄 하나 잔존, N5).** progress.md:42 「18 REQ/19 AC」 = `grep -cE '^### REQ-FLH-[0-9]+' spec.md` 18, `grep -cE '^### AC-FLH-[0-9]+' acceptance.md` 19(E1). 같은 목록의 progress.md:41 「16개 t1082 named tests」는 갱신되지 않았다.

## Defects Found

D-N1. REQ18-BLOCKS-T1074-REQUIRED-BINDER — spec.md:136 (REQ-FLH-018), spec.md:132 (REQ-FLH-017), design.md:74 (§2.1 표 4번 행), design.md:82 (수명주기 도식), design.md:99 (넷째 규칙) — t1074 착지 계약에서 launch-pending 행을 bound 로 바꾸는 **필수** 경로는 첫 정상 UserPromptSubmit 이며 SessionStart 는 순서가 보장되지 않는 best-effort 다(t1074 spec.md:57 「correctness SHALL NOT depend on that ordering. When the first legitimate non-empty UserPromptSubmit is observed, the registry SHALL atomically rebind any remaining provisional lane …」; t1074 design.md:32·:46). 착지 코드도 그렇다 — UserPromptSubmit 모드는 `BindLaunchPending` 을 건너뛰고 `RegisterPeer` 로 결합한다(factory_messages.go:83-93, 테스트 `TestFactoryUserPromptSubmitRebindsLaunchPendingPeer` factory_messages_test.go:345). REQ-FLH-018 은 비종결 handoff 동안 이 경로를 행 상태와 무관하게 거부하고, design.md:99 는 그 동안 endpoint 를 옮길 수 있는 writer 를 「launcher 경로(1-3번)와 handoff rebind 뿐」으로 한정한다. 따라서 REQ-FLH-017 이 다루는 순서 — handoff `SWITCH_PENDING_INTERACTIVE` 중 lane 재기동, 재기동 프로세스의 SessionStart(startup)가 launcher 등록보다 먼저 발화해 `BindLaunchPending` 이 no-op(행이 아직 pending 이 아님, store.go:419-421) — 에서 재기동된 lane 은 launch-pending 으로 남고, 첫 사용자 prompt 의 UserPromptSubmit 결합은 `ENDPOINT_HANDOFF_PENDING` 으로 거부된다. handoff 는 사용자가 `/cd` 를 다시 해서 다음 turn SessionStart evidence 가 오기 전까지 종결되지 않으므로(design.md:116 「next normal-turn SessionStart → BOUND, NACK, ABANDONED」), 그 사이 lane 은 결합자가 없는 launch-pending 상태(송신·수신 모두 `ErrEndpointLaunchPending`)에 머문다. 이는 REQ-FLH-017 의 「no orphan launch-pending endpoint」 및 REQ-FLH-015 의 t1074 결합 수명주기 재사용 의무와 모순이다. 부수적으로 REQ-FLH-015(spec.md:124)와 design.md:82 는 t1074 결합자를 「first-normal-turn SessionStart」로 서술해 t1074 0.1.4(t1074 spec.md:23)와 어긋나며, AC-FLH-018 의 A-bind 도 `BindLaunchPending` 만 쓰므로 이 순서는 어느 AC 에서도 재현되지 않는다. — Severity: critical — Class: blocking — Required fix: 둘 중 하나를 택해 적는다. (a) REQ-FLH-018 에 carve-out 추가: 현재 행이 launch-pending 이고 호출자 PID/process-start 가 그 provisional 행의 owner 와 정확히 같으면 UserPromptSubmit 은 t1074 provisional→bound 결합을 수행한다(이미 launcher 가 이겼으므로 handoff rebind 는 REQ-FLH-017 대로 `STALE_GENERATION`). (b) launcher 등록 transaction 이 같은 slot 의 비종결 handoff 를 같은 transaction 안에서 `STALE_GENERATION` NACK 으로 종결한다. 어느 쪽이든 design.md:74 행 4 의 「효과」 칸에 「t1074 필수 launch-pending 결합자」 역할을 추가하고, design.md:82·spec.md:124 의 결합자 서술을 t1074 0.1.4 와 맞추며, AC-FLH-018 에 「A-register → UserPromptSubmit bind(SessionStart 선행 no-op)」 강제 순서를 넣는다.

D-N2. REQ18-SAME-TX-UNDISCRIMINATED — acceptance.md:268-275 (AC-FLH-019), spec.md:136 — REQ-FLH-018 은 거부를 「inside the same broker write transaction that reads the handoff state」로 요구하지만, AC-FLH-019 의 세 강제 순서는 모두 H 대 rebind 경합이다. handoff 상태를 `BEGIN` 전에 읽는 TOCTOU 구현도 (i)(H 가 먼저라 상태가 비종결), (ii)(요구가 「행 불변」뿐이라 오래된 읽기로 거부해도 통과), (iii)(종결 뒤 읽기)를 모두 통과한다. 이 조항이 실제로 막는 경합은 H 대 **reservation 생성**(RESERVED 진입)인데 그 순서는 AC 에 없다. — Severity: major — Class: optional (SHOULD-FIX) — Required fix: AC-FLH-019 에 강제 순서 (iv) 「H 의 상태 판독과 행 쓰기 사이에 reservation transaction 이 커밋을 시도」를 추가해, H 가 트랜잭션 밖에서 상태를 읽었다면 RESERVED 인 채로 endpoint 가 옮겨지는 것을 FAIL 로 잡는다.

D-N3. AC19-DIRECT-REGISTRATION-POLICY — acceptance.md:268 vs acceptance.md:18 — AC-FLH-019 는 racer H 가 `RegisterPeer` 를 직접 호출하도록 규정하는데, 전역 정책 :18 은 「direct peer registration … 은 vacuous PASS 가 아니라 FAIL」이다. 정책의 의도가 「준비 단계 우회 금지」라면 H 호출은 피시험 대상 그 자체이므로 예외지만, 문언상 구분이 없다. 게이트 명령이 `./internal/hook` 도 포함해 테스트 위치도 열려 있다. — Severity: minor — Class: optional — Required fix: :18 의 「direct peer registration」을 「production 경로를 우회한 fixture 준비」로 좁히거나, AC-FLH-019 에 「H 의 `RegisterPeer` 호출은 피시험 store entry point 이므로 :18 적용 대상이 아니다」를 명시한다.

D-N4. HANDLE-PROOF-OPENCONN-VACUOUS — acceptance.md:258 증명 (2) — `db.Stats().OpenConnections >= 1` 은 `Open` 이 스키마 실행 후 유휴 연결을 남기므로 경쟁자 실행 전에도 참이고(H1 `s2.Open=1 s2.InUse=0`), 공유 핸들에서도 참이다(H3 `s1.Open=1`). 증명 (1) 포인터 부등이 실질 증거를 이미 제공하므로 결론을 뒤집지는 않는다. — Severity: minor — Class: optional — Required fix: (2)를 「대기 중인 쪽 핸들의 `InUse == 1` 이고 `WaitCount == 0`(풀 대기 아님)」으로 바꾼다.

D-N5. STALE-SIBLING-COUNTS — progress.md:41 「16개 t1082 named tests」, design.md:99 「순서 규칙은 세 가지다」(뒤에 넷째까지) — D5 수리의 인접 줄 누락과 규칙 개수 불일치. — Severity: minor — Class: optional — Required fix: progress.md:41 을 현재 named test 수(요약표 19 + gate-quality 1)로, design.md:99 를 「네 가지」로 고친다.

D-N6. POST-NACK-FRESH-RESERVATION-ADMISSION — design.md:103, design.md:204 (§9 새 행 「재시도는 fresh reservation만」), spec.md:76 (REQ-FLH-003) — AC-FLH-019 (iii) 끝 상태에서 사용자는 이미 `/cd` 로 card worktree 안에 있고 target 경로도 존재한다. 이 상태의 fresh reservation 이 REQ-FLH-003 의 「untrusted cwd」「conflicting target path」 NACK 에 걸리는지, §9 「target 존재, exact pin/branch/clean → WT_READY 재구성」이 fresh reservation 에도 적용되는지 명시돼 있지 않다. broker 교착은 아니지만 design.md:103 의 「교착 방지」 주장이 이 경로의 성공에 기대고 있다. — Severity: minor — Class: optional — Required fix: §9 새 행 또는 REQ-FLH-003 에 「NACK 된 동일 card 의 clean·pinned target 이 이미 있고 lane cwd 가 그 target 이면 fresh reservation 은 WT_READY 재구성으로 진입」 여부를 한 줄로 적는다.

## Evidence

E0. 트리 확인
```
$ git rev-parse --show-toplevel; git branch --show-current; git rev-parse --short HEAD; git status --short
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1082
WT-factory-lane-worktree-handoff
c29d35ea7
(출력 없음)
```

E1. heading 수와 progress 주장
```
$ grep -cE '^### REQ-FLH-[0-9]+' spec.md
18
$ grep -cE '^### AC-FLH-[0-9]+' acceptance.md
19
progress.md:41:- 16개 t1082 named tests는 현재 모두 부재하여 plan RED 상태다.
progress.md:42:- Strict SPEC lint는 빈 finding 배열을 반환했고, 18 REQ/19 AC heading과 양방향 trace reference를 확인했다.
$ moai spec lint SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001 --strict --json; echo lint_exit=$?
[]
lint_exit=0
```

E2. slot 쓰기 경로 전수 (`^func` 정의 + 호출부, 비-테스트)
```
$ grep -rnE '\b(RegisterPeer|RegisterLaunchPending|BindLaunchPending|RollbackLaunchPending)\(' internal --include='*.go' | grep -v '_test.go'
internal/factorymsg/store.go:305:func (s *Store) RegisterPeer(ctx context.Context, p Peer) (Peer, error) {
internal/factorymsg/store.go:381:func (s *Store) RegisterLaunchPending(ctx context.Context, p Peer) (Peer, error) {
internal/factorymsg/store.go:386:	return s.RegisterPeer(ctx, p)
internal/factorymsg/store.go:392:func (s *Store) BindLaunchPending(ctx context.Context, p Peer) (Peer, bool, error) {
internal/factorymsg/store.go:455:func (s *Store) RollbackLaunchPending(ctx context.Context, p Peer) (bool, error) {
internal/cli/factory_launch_pending.go:58:	return s.RegisterLaunchPending(ctx, factorymsg.Peer{
internal/cli/factory_launch_pending.go:73:	_, err = s.RollbackLaunchPending(ctx, pending)
internal/hook/factory_messages.go:84:		p, bound, bindErr := s.BindLaunchPending(ctx, want)
internal/hook/factory_messages.go:93:	p, err := s.RegisterPeer(ctx, want)
$ grep -rn 'registerFactoryUserPromptPeer\|registerFactorySessionStartPeer\|registerLaunchPending\|rollbackLaunchPending' internal --include='*.go' | grep -v _test.go
internal/hook/session_start.go:458:	if factoryNotice := registerFactorySessionStartPeer(ctx, input); factoryNotice != "" {
internal/hook/user_prompt_submit.go:112:		bindNotice := registerFactoryUserPromptPeer(bindCtx, input)
internal/hook/factory_messages.go:34:func registerFactorySessionStartPeer(ctx context.Context, input *HookInput) string {
internal/hook/factory_messages.go:38:func registerFactoryUserPromptPeer(ctx context.Context, input *HookInput) string {
$ grep -rnEi '(insert (or [a-z]+ )?into|update|delete from|replace into) +peers' internal --include='*.go' | grep -v _test.go
internal/factorymsg/store.go:367:	_, err = tx.ExecContext(ctx, `INSERT INTO peers(... ON CONFLICT(slot) DO UPDATE ...` (생략)
internal/factorymsg/store.go:429:	result, err := tx.ExecContext(ctx, `UPDATE peers SET session_uuid=?,generation=?,updated_at=?
internal/factorymsg/store.go:464:	result, err := s.db.ExecContext(ctx, `DELETE FROM peers
$ grep -rlnE 'FROM peers|INTO peers|UPDATE peers' internal pkg cmd --include='*.go' | grep -v _test.go
internal/factorymsg/store.go
```

E3. 임시 프로브 `internal/factorymsg/zz_auditprobe_t1082_iter5_test.go` (실행 후 삭제). 두 번의 production `Open` 으로 같은 broker 경로에 핸들 둘을 열고, `ownerCurrent` 는 현재 프로세스만 live 로 주입.
```
$ unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=<scratchpad>/gc go test ./internal/factorymsg -run '^TestZZAuditProbeT1082Iter5$' -count=1 -v
=== RUN   TestZZAuditProbeT1082Iter5
    zz_auditprobe_t1082_iter5_test.go:38: B0 pending.gen=1 bind ok=true err=<nil> bound.session=src-uuid bound.gen=2
    zz_auditprobe_t1082_iter5_test.go:42: B1 same-owner new-uuid err=<nil> got.session=post-cd-uuid got.gen=3
    zz_auditprobe_t1082_iter5_test.go:44: B2 same-owner old-uuid again err=<nil> got.session=src-uuid got.gen=4
    zz_auditprobe_t1082_iter5_test.go:48: H1 storeDistinct=true dbDistinct=true s1.Open=1 s1.InUse=0 s2.Open=1 s2.InUse=0
    zz_auditprobe_t1082_iter5_test.go:66: H2 distinct: s2 blocked; s2.Open=1 s2.InUse=1 s2.WaitCount=0
    zz_auditprobe_t1082_iter5_test.go:69: H2 distinct: s2 after release err=<nil>
    zz_auditprobe_t1082_iter5_test.go:84: H3 shared: blocked; s1.Open=1 s1.InUse=1 s1.WaitCount=1
    zz_auditprobe_t1082_iter5_test.go:87: H3 shared: after release err=<nil>
--- PASS: TestZZAuditProbeT1082Iter5 (0.45s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/factorymsg	0.730s
$ rm internal/factorymsg/zz_auditprobe_t1082_iter5_test.go
$ git status --short
(출력 없음)
```
B0 은 production launcher 경로(`RegisterLaunchPending` → `BindLaunchPending`)로 bound 를 만든 것이다. H2 는 s1 이 `BEGIN IMMEDIATE` 쓰기 transaction 을 쥔 동안 s2 의 `RegisterPeer` 가 자기 연결을 잡은 채(`InUse=1`, 풀 대기 `WaitCount=0`) SQLite 잠금에서 대기함을 보인다.

E4. AC-FLH-019 게이트의 RED 판정과 두-패키지 형태의 건전성
```
$ unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=<scratchpad>/gc go test -json ./internal/factorymsg ./internal/hook -run '^TestFactoryLaneHandoffRebindVsUserPromptRegisterRace$' -count=1 -timeout=180s > <scratchpad>/ac19red.jsonl
$ jq -c 'select(.Action!="output")|{Action,Package,Test}' ac19red.jsonl
{"Action":"start","Package":"github.com/modu-ai/moai-adk/internal/factorymsg","Test":null}
{"Action":"pass","Package":"github.com/modu-ai/moai-adk/internal/factorymsg","Test":null}
{"Action":"start","Package":"github.com/modu-ai/moai-adk/internal/hook","Test":null}
{"Action":"pass","Package":"github.com/modu-ai/moai-adk/internal/hook","Test":null}
$ jq -se '<acceptance.md:279 판정식 그대로>' ac19red.jsonl; echo jq_exit=$?
false
jq_exit=1
```
매치가 없는 패키지는 `skip` 이 아니라 패키지 수준 `pass` 를 내므로, 테스트가 한 패키지에만 있어도 게이트가 오탐 FAIL 하지 않는다. 테스트 부재는 `false` 로 FAIL 한다.

E5. RED ledger 재측정
```
$ rg -n -F 'func TestFactoryLaneHandoffRebindVsUserPromptRegisterRace(' internal --glob '*_test.go'; echo rg_exit=$?
rg_exit=1
$ rg -n -F 'func TestFactoryLaneHandoffRebindVsLaunchBindRace(' internal --glob '*_test.go'; echo rg_exit=$?
rg_exit=1
```

E6. MP-5/6/7
```
$ grep -rn 'NEEDS CLARIFICATION' plan.md research.md; echo nc_exit=$?
nc_exit=1
$ grep -c syscall spec.md
0
$ grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u
SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
SPEC-FACTORY-MIXED-HOOK-001
$ grep -n '^status:' .moai/specs/SPEC-FACTORY-MIXED-HOOK-001/spec.md
5:status: completed
```

E7. 코드 좌표 (N1·Part B 근거)
```
$ grep -n 'factory session UUID has a live owner\|factory logical lane has a live owner\|p.Generation = oldGen + 1\|p.Generation = oldGen$\|A bound row is never rotated\|separate policy' internal/factorymsg/store.go
345:					return Peer{}, errors.New("factory session UUID has a live owner")
348:					p.Generation = oldGen + 1
351:				p.Generation = oldGen
355:				return Peer{}, errors.New("factory logical lane has a live owner")
358:				p.Generation = oldGen + 1
390:// row. A bound row is never rotated here; authoritative turn hooks use
391:// RegisterPeer for that separate policy.
$ grep -n 'mode == factoryPeerBindSessionStart\|s.Peer(ctx, input.SessionID)' internal/hook/factory_messages.go
76:	if current, peerErr := s.Peer(ctx, input.SessionID); peerErr == nil {
83:	if mode == factoryPeerBindSessionStart {
116:	p, err := s.Peer(ctx, input.SessionID)
$ grep -n 'func TestFactoryUserPromptSubmitRebindsLaunchPendingPeer' internal/hook/factory_messages_test.go
345:func TestFactoryUserPromptSubmitRebindsLaunchPendingPeer(t *testing.T) {
```

---

## Part B — t1074 자체 계약 판정 (SPEC 과 독립)

**질문**: `Store.RegisterPeer` 가 살아 있는 같은 owner(PID + process-start 동일)의 새 session UUID 등록에 대해 거절 없이 slot endpoint 를 generation+1 로 교체하는 것은 t1074 계약 위반인가.

**판정: UNSPECIFIED** (DEFECT 아님). 한 하위 사례는 t1074 가 의도·검증한 동작이며, 일반 사례에 대해서는 t1074 가 침묵한다. 해로움은 t1082 가 존재할 때에만 생긴다.

### 코드가 하는 일 (착지 코드 기준, 재측정)

- store.go:353-359: 기존 행의 session 과 다르고, 호출자 identity 가 기존 owner 와 **같으면** `s.ownerCurrent` 를 부르지도 않고(단락 평가) `p.Generation = oldGen + 1` 로 교체한다. 다른 identity 이고 기존 owner 가 live 일 때만 「factory logical lane has a live owner」로 거절한다(:355).
- store.go:367: `ON CONFLICT(slot) DO UPDATE` 로 행 하나를 덮는다. 묘비·receipt 없음.
- 재측정(E3): production launcher 경로로 bound 된 행에서 B1 `post-cd-uuid` gen 3 교체, B2 옛 `src-uuid` 가 다시 gen 4 로 lane 을 되찾는다.
- 호출부: UserPromptSubmit 모드는 `BindLaunchPending` 을 건너뛰고 이 `RegisterPeer` 를 부른다(factory_messages.go:83-93). 식별자가 같으면 쓰기 없이 반환한다(:76-80).

### t1074 텍스트 대조

1. **위반 후보 — 해당 없음.**
   - REQ-FMH-002(t1074 spec.md:61) 「SHALL reject stale generation/session ownership. While an owner's PID and process-start identity remain live, the registry SHALL NOT displace that owner because of heartbeat expiry alone」: 교체 주체가 같은 live owner 이므로 owner 를 밀어낸 것이 아니다. 같은 프로세스의 현재 session 은 「stale session ownership」도 아니다.
   - AC-FMH-003(t1074 acceptance.md:19, :152) 「the lane resolves only the current endpoint and only the exact session UUID/generation/process-start tuple succeeds」: 교체 뒤 옛 endpoint 는 세대가 낮아 정확한 tuple 이 아니므로 거부된다 — 이 AC 와 모순되지 않는다.
   - REQ-FMH-001 의 동일-endpoint 조항(t1074 spec.md:57 「Once the same actual endpoint is already bound, later prompts SHALL leave peer generation and persisted fields, including `updated_at`, unchanged」)은 **같은** endpoint 에만 적용되고, 코드도 :76-80 에서 지킨다.
2. **의도 근거 — 하위 사례에 한정.**
   - t1074 design.md:32 「the first legitimate non-empty `UserPromptSubmit` is the required fallback. It binds the observed hook session UUID … After the identical endpoint is bound, later prompts skip the registry write entirely」, :46 「If the identical endpoint is already bound, skip the bind write」: skip 은 identical 일 때뿐이라는 구조가 「non-identical 이면 관측된 session 으로 결합한다」를 함의한다.
   - t1074 spec.md:57 「The logical lane ID SHALL remain the broker address while the session UUID/generation is a replaceable physical endpoint」.
   - 착지 테스트 `TestFactorySessionStartCannotRotateAuthoritativeUserPromptBinding`(internal/hook/factory_messages_test.go:256-281)은 SessionStart 가 먼저 다른 session(alias)으로 bound 시킨 **bound 행**을 UserPromptSubmit 이 같은 owner 의 실제 session 으로 gen+1 교체하도록 단언한다(:275-280 「UserPromptSubmit did not rotate alias to actual session」). 이 테스트는 t1074 최종 판정서의 「Atomic owner/bind race gate」 증거에 PASS 로 기록돼 있다(.moai/reports/t1074/verdict.md:15, :20). 코드 주석도 store.go:390-391 「authoritative turn hooks use RegisterPeer for that separate policy」라고 적는다.
   - 즉 「bound 행을 같은 owner 의 다른 session 으로 교체」는 alias 교정이라는 하위 사례에서 t1074 가 의도하고 검증한 동작이다. 그러나 이 alias 사례를 명시한 REQ/AC 는 없고(REQ-FMH-001 은 「any remaining provisional lane」의 재결합만 규정), 테스트가 증거로 쓰였을 뿐이다.
3. **침묵 — 일반 사례.**
   - `/cd`·`/new`·resume 처럼 **UserPromptSubmit 이 이미 bound 시킨 endpoint** 를 같은 프로세스의 다른 session 이 대체하는 경우, 옛 session 의 재탈환(B2) 모두 t1074 REQ/AC/design 어디에도 규정이 없다. 코드는 alias 사례와 이 사례를 구분할 수단이 없어 같은 분기로 처리한다.
   - t1074 는 이 영역을 명시적으로 t1082 에 넘겼다: spec.md:50 「Worktree 생성, interactive `/cd`, headless `cwd` handoff, 그 이후 endpoint rebind는 `t1082`가 소유한다」, spec.md:123 Out of Scope 「endpoint transition state, and atomic old→new session rebind (`t1082`)」, design.md:36.

### t1082 없이도 해로운가

해롭지 않다고 판단한다. t1074 만 있을 때 교체 뒤 옛 endpoint 의 claim token·세대는 stale 이 되어 ACK 가 거부되고(AC-FMH-003·AC-FMH-004), 미확인 claim 은 lease 만료로 재전달된다(REQ-FMH-005 「When a claim lease expires, the broker SHALL redeliver」, at-least-once). 한 프로세스에는 활성 UI session 이 하나뿐이므로 「가장 최근 turn 의 session 이 lane 을 가진다」는 결과가 t1074 의 어떤 불변식도 깨지 않는다.

t1082 가 생기면 해로워진다: 비종결 handoff 가 source tuple 을 CAS 로 고정한 상태에서 이 교체는 묘비·BOUND receipt·검증 없이 endpoint 를 옮겨 REQ-FLH-008·010·013 과 충돌한다. REQ-FLH-018 이 존재하는 이유가 이것이며, 그 수리는 t1074 결함 수정이 아니라 t1082 의 계약 확장으로 보는 것이 맞다. 다만 alias 교정 사례가 t1074 의 검증된 동작이므로, REQ-FLH-018 구현이 비종결 handoff **밖**에서 이 동작을 바꾸면 `TestFactorySessionStartCannotRotateAuthoritativeUserPromptBinding` 이 회귀 가드로 걸린다(AC-FLH-015 호환성 범위).

---

## Baseline-attribution

- 트리 `.claude/worktrees/t1082`, 브랜치 `WT-factory-lane-worktree-handoff`, HEAD `c29d35ea7`. 감사 시작·종료 시 `git status --short` 비어 있음(E0, E3).
- 코드 좌표는 이 HEAD 의 `internal/factorymsg/store.go`, `internal/hook/factory_messages.go`, `internal/hook/factory_messages_test.go`, `internal/hook/user_prompt_submit.go`, `internal/cli/factory_launch_pending.go` 에서 직접 읽었다.
- t1074 텍스트는 이 트리의 `.moai/specs/SPEC-FACTORY-MIXED-HOOK-001/`(마지막 변경 커밋 `875efc28b`)과 `.moai/reports/t1074/verdict.md` 에서 읽었다.
- 프로브는 이 트리에 임시로 두었다가 실행 직후 삭제했다. 빌드 캐시는 세션 scratchpad 에 두었다.

## Gaps

- `audit_multi` / `codex_audit` / `glm_audit` 교차 모델 의견은 호출하지 않았다(Claude 단독 델타 감사).
- N1 의 순서(재기동 프로세스 SessionStart 가 launcher 등록보다 먼저)는 t1074 가 「보장되지 않는다」고 명시한 순서에 기대는 논증이며, 실제 Codex 에서 그 순서가 나는 것을 관측하지 않았다. handoff 구현이 없어 종단 재현도 하지 않았다.
- `-race` 로는 프로브를 돌리지 않았다(프로브는 경합 검출이 아니라 상태·통계 관측용).
- 개정이 건드리지 않은 절(REQ-FLH-001..017 본문, AC-FLH-001..017 본문)은 재감사하지 않았다. N1 판정에 필요한 REQ-FLH-015·017 문언만 다시 읽었다.
- Part B 에서 t1074 의 `operational-lane-status-addendum.md`, `plan.md`, `research.md`, `progress.md` 전문은 읽지 않았고 키워드 grep 으로만 확인했다.

## Residual-risk

- REQ-FLH-018 의 묘비 UUID `STALE_ENDPOINT` 는 BOUND 뒤에도 영구적이다. 사용자가 BOUND 뒤 같은 프로세스에서 옛 session 으로 resume 하면 그 session 은 lane 에 결합되지 못하고 메시지는 사용하지 않는 `post-cd-uuid` endpoint 에 머문다. 이는 설계 선택이며 결함으로 보지 않았다.
- 비강제 200회 반복의 두 end state((i)·(ii))는 끝 상태가 같아 어떤 순서가 실제로 일어났는지를 가르지 못한다. 비공허성은 강제 순서 쪽이 떠받친다.

## Recommendation

1. **N1 을 먼저 고친다(차단).** REQ-FLH-018 에 launch-pending 행의 exact-owner 결합 carve-out 을 넣거나, launcher 등록이 비종결 handoff 를 같은 transaction 에서 `STALE_GENERATION` 으로 종결하도록 REQ-FLH-017 을 보강한다. design.md:74 행 4 효과 칸, design.md:82 도식, spec.md:124 의 결합자 서술을 t1074 0.1.4(UserPromptSubmit 필수, SessionStart best-effort)와 맞추고, AC-FLH-018 에 UserPromptSubmit 결합 순서를 추가한다.
2. N2(SHOULD-FIX): AC-FLH-019 에 H 대 reservation 생성 강제 순서를 추가한다.
3. N3-N6 은 선택 사항이다.
4. 재감사는 N1(과 선택한 경우 N2) 델타로 한정할 수 있다. 반복 상한은 이미 넘었으므로 진행 방식은 오케스트레이터가 사용자에게 확인한다.
