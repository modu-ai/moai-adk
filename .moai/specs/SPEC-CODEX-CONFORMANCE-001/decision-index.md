# SPEC-CODEX-CONFORMANCE-001 — Decision Index

Carrier for decisions surfaced during plan-phase assembly that the operator has not settled in an interview (`interview.decision_gate: on`, `.moai/config/sections/interview.yaml:6`). Rows state what is unresolved and why; no judgment-call row carries an embedded recommendation beyond a published-Default application. `Operator verdict:` lines are empty for non-implementation-level rows at authoring.

### Q1: 0.160.0 fixture 세트를 교체하는가, 0.161.0과 병존 보존하는가?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — the operator card delegated the replace-vs-retain judgment explicitly; no committed policy row fixes fixture retention. The regeneration DISCIPLINE itself is committed: `internal/cli/testdata/codex-0.160.0/README.md` and `.moai/docs/factory-managed-session.md` — "최소 지원 codex 버전이 올라가면 사본을 다시 만든다".)
Why unresolved: 교체는 누적된 0.160.0 세트를 지우고(소비자 리터럴 2곳 이전) 0.161.0 세트로 대체하는 것이고, 보존은 두 세트를 함께 두는 것이다. 두 소비자는 디렉터명을 리터럴로 쥐므로 병존 시 0.160.0을 읽는 시험이 남지 않아 두 번째 진실의 원천이 된다.
Default: replace (rule: preserves current behavior — the committed vendored-dir contract is regeneration-on-minimum-version-move, so replacement is the behavior-preserving option and retention is the deviation)
Alternate: retain both directories (a backward-compat suite against the retired minimum)
Operator verdict: DEFAULT-APPLIED 2026-10-08T15:29:12Z manager-spec (plan-phase SPEC author, card t1607)

### Q2: 라이브 바이너리 측정을 자동 검증에 넣는가?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — no committed policy ranks fixture-based vs live-binary conformance for this adapter.)
Why unresolved: 0.161.0 실행이 필요한 확인 세 가지(login status 출력 형태, resume --help 옵션 집합, 재생 프레임 확장 필드)를 설치 바이너리 의존 자동 AC로 넣으면 CI가 운영자 환경에 종속된다. fixture가 판정 불가한 항목의 처분(수동 검증 기록 vs 라이브 게이트)은 명시적 판정이 필요하다.
Default: fixture-based mechanical conformance as the sole automated surface; operator-environment-dependent checks recorded as manual verification in progress.md §E.2, never as automated ACs (rule: preserves current behavior — the existing conformance suite is entirely fixture-based and CI-runnable)
Alternate: add a live-binary smoke gate to CI
Operator verdict: DEFAULT-APPLIED 2026-10-08T15:29:12Z manager-spec (plan-phase SPEC author, card t1607)

### Q3: auth.json 부재 시 어댑터는 무엇을 하는가 — keyring 전환(#49361)에 대한 어댑터 변경이 필요한가?

Label: DECIDED
Authority anchor: SPEC-CODEX-LAUNCHER-001 (the two-stage auth ladder discipline: a rejected file — unknown mode, blank credential material, unparseable, or absent — descends to stage 2, never an outcome; an unreadable probe is a gap, not a verdict), committed implementation `internal/cli/mcp_codex.go` `classifyCodexAuth` (stage-1 → stage-2 descent) and `codexAuthStatusLine` whole-line grammar; tests `internal/cli/codex_auth_ladder_test.go`.
Why unresolved: (none — the committed ladder already answers the keyring-absence question: absence descends to stage 2, which asks the codex binary directly, so moai needs no keyring client. Recorded here because the card named it as an open axis.)
Operator verdict:

### Q4: stage 2 문법이 0.161.0 `codex login status` 출력(keyring 변어 포함)을 여전히 분류하는가?

Label: EVIDENCE-NEEDED
Authority anchor: (partial — version/help output captured and iter-2 re-verified: `codex-cli 0.161.0` and `generate-json-schema` presence, lane 2026-10-08T16:14-16:15Z; the login-status output this row needs is not yet captured — its first capture is AC-CONF-005's mandatory M2 minimum observation.)
Why unresolved: #49361이 인증 문서를 keyring 저장으로 갱신하며 출력 문구가 바뀌었을 가능성이 있다. 정통 라인 문법 `logged in using (chatgpt|api key)`가 0.161.0의 실제 출력과 일치하는지는 0.161.0 실행 관측으로만 확정된다. M2가 정화 샘플을 포획해 fixture화하거나, 불가하면 수동 검증 갭으로 기록한다(REQ-CONF-007의 두 경로).
Operator verdict:

### Q5: Tier M 분류가 올바른가?

Label: POLICY-COVERED
Authority anchor: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (M = 300–1000 LOC guidance, 5–15 files, 3-file artifact set, 16/16 REQ/AC ceilings).
Why unresolved: (none — the classification criteria are committed policy; the measured shape fits M: one vendored data directory + 2 test files + optional 1–2 test additions + 1 doc touch, 8 REQ / 8 AC within ceilings.)
Operator verdict:

### Q6: filesystem escalation #49353에 어댑터 변경이 필요한가, 아니면 기록된 NO-OP로 충분한가?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — no committed policy fixes the disposition of upstream sandbox-escalation semantics for this adapter.)
Why unresolved: escalation 의미 확장은 codex가 승인을 받았을 때의 권한 폭 변화다. 소유(owner) 턴의 요청은 관리 소유자가 기각하므로 소유자 경로는 무영향이지만(`TestManagedCodexServerRequestPolicy` 검증), 오퍼레이터가 붙은 턴의 요청은 `leavesForOperator`(`internal/cli/managed_codex_tui.go:654`, 분기 `managed_codex_factory.go:435`, 로그 `managed_codex_tui.go:670`)에 따라 오퍼레이터 TUI에 남고 #49353은 그 **인간 승인**이 풀 수 있는 권한을 넓힐 수 있다. 검증 수위(기존 정책 시험 재실행 vs 확장 시나리오 추가)는 판정이 필요하다.
Default: documented NO-OP — no adapter change and no new approval handling; REQ-CONF-008 scopes the decline-all guarantee to owner turns, records the operator-approval path as a known surface, and M4 records the operator-path exposure as an observation item, verified via the existing `TestManagedCodexServerRequestPolicy` family on the refreshed fixtures (rule: preserves current behavior — the owner's answer policy is version-independent at the protocol level and escalation-aware handling would be new capability, out of this card's scope)
Alternate: extend the adapter with escalation-aware approval handling
Operator verdict: DEFAULT-APPLIED 2026-10-08T16:36:03Z manager-spec (plan-phase SPEC author, card t1607, iter-2 repair)

### Q7: 0.161.0 제너레이터 출력(39파일, 4.3MB)을 전량 vendoring하는가, 소비 부분집합만 하는가?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — no committed policy row fixes vendoring composition; the committed dir's own composition — 8 consumed Response schemas + resume-help + README — is the precedent being mirrored, not a policy.)
Why unresolved: 0.161.0 `generate-json-schema`는 39파일을 v1/v2 분할 레이아웃으로 내놓고 통합 번들 2종(codex_app_server_protocol.schemas.json 716KB / .v2 615KB)을 포함한다. 신규 31파일(request-side Params 동반, JSONRPC 봉투 계열, ToolRequestUserInput*, FuzzyFileSearch*, Attestation*/ChatgptAuthTokensRefresh*, RequestId)은 본 트리에 소비자가 없다. 전량 벤더링은 소비자 8종에 4.3MB를 싣는 일이다.
Default: vendor the consumed subset — the 8 named Response schemas via mechanical named-file copy from the generator output, plus resume-help.txt capture and README regeneration, mirroring the committed dir's composition (rule: preserves current behavior — the vendored dir's committed composition is exactly this subset; the copy is mechanical with no content edits, so the fixture hand-editing ban is untouched). Future need for the request-side Params companions re-vendors per the README discipline.
Alternate: vendor the full generator output (4.3MB incl. two redundant consolidated bundles and a v1/v2 split no consumer reads)
Operator verdict: DEFAULT-APPLIED 2026-10-08T17:12:42Z manager-spec (plan-phase SPEC author, card t1607, plan amendment round — lane measurement 2026-10-08T16:56Z+)
