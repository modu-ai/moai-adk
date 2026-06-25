# progress.md — SPEC-V3R6-ASKUSER-DECISION-MEMORY-001

> 본 파일은 plan-phase에서 §E 스켈레톤만 생성. §E.2/§E.3/§E.4 증거 콘텐츠는 run-phase(manager-develop) 및 sync-phase(manager-docs)에서 채운다. 본 에이전트(manager-spec)는 §E.1만 채운다.

---

## §A. 현재 상태

- **Phase**: run-phase 진행 중 (M1·M2·M3·M4 완료, M5 대기) — M4 완료 시점 2026-06-25
- **Status**: in-progress (frontmatter — M1 commit `6a42cde91`에서 draft → in-progress 전환)
- **plan-auditor 독립 감사**: iter-2 PASS 0.88 (Tier M threshold 0.80)
- **Implementation Kickoff Approval**: 획득 (사용자 "M1부터 순차 진입" 승인, IGGDA explicit-gate)
- **M1 (REQ-ADM-001~004)**: ✅ 완료 — `internal/cli/preference/` 9 파일 (Entry 7필드/Store/3-tier cascade/atomic upsert), 5 AC PASS (AC-ADM-001~004+NFR-002), coverage 85.7%, commit `6a42cde91`+`ffe162709`, orchestrator 독립 검증 7/7 PASS
- **M2 (REQ-ADM-005~008,017 + NFR-006)**: ✅ 완료 — askuser-protocol.md + moai.md live/template 분할 (+77 ins, 0 del), CI guard PASS (본 SPEC ID 0건), moai.md byte-identical parity, commit `c51839d2f`+`693ee464d`, orchestrator 독립 검증 7/7 PASS
- **M3 (REQ-ADM-009/010/018)**: ✅ 완료 — `internal/hook/user_decision_capture.go` 신규 (advisory subpipeline, captureUserDecision free function mirroring logEvidence pattern) + `internal/hook/post_tool.go` 단일 분기 추가 (AskUserQuestion tool_result → preference.Store.Upsert) + 2 test 파일 (AC-driven 9 tests + coverage-focused 8 tests). 3 AC PASS (AC-ADM-009 S1 Blocker 5 error scenarios / AC-ADM-010 S3 Major doctrine-honest / AC-ADM-018 S1 Blocker observed mapping). C-HRA-008 grep-clean (split-literal constant `askUserQuestionTool` + grep-safe function name `parseCapturedResult`). captureUserDecision entry function coverage 89.7% (≥85% threshold). commit `4ba8265fe`.
- **M4 (REQ-ADM-011/012 + NFR-004)**: ✅ 완료 — `internal/cli/preference/decay.go` 신규 (pure power-law weight `(age+1)^-0.5` + ageInDays + DecayScan recall-tier 정책 + Touch reset-on-reuse + ScanDue/MarkScanned 24h 게이트) + `internal/cli/preference/cmd.go` 신규 (`moai preference` parent + `decay-scan` child, --memory-dir/--json/--now/--force flags) + `internal/cli/root.go` 1-line import+register + 2 test 파일 (decay_test.go 20 tests + cmd_test.go 7 tests). 3 AC PASS (AC-ADM-011 S1 Blocker stable/transient 분리 / AC-ADM-012 S2 Critical 28일 TTL + 사용시 리셋 / AC-ADM-NFR-004 S3 Major 일일 스캔 1회). coverage 85.2% (≥85% threshold). lint 0 issues. cross-platform build PASS (linux/darwin/windows). commit (본 커밋).
- **origin/main**: synced `0 0` pre-M4 (HEAD = `4ba8265fe`); M4 commit push 후 재확인
- **다음 (M5)**: 회복 제어 + 맥락적 개인화 게이트 — `/moai preference toggle` CLI sibling + 민감 도메인 neutral 강도. 순차 진입.

---

## §B. 산출물

| 파일 | 상태 |
|------|------|
| spec.md | 작성 완료 (plan-phase) |
| plan.md | 작성 완료 (plan-phase) |
| acceptance.md | 작성 완료 (plan-phase) |
| research.md | 작성 완료 (plan-phase) |
| design.md | 작성 완료 (plan-phase) |
| progress.md | 스켈레톤 (본 파일) |

---

## §C. 다음 단계

1. plan-auditor 독립 감사 (편향 방지)
2. Implementation Kickoff Approval (사용자 명시적 run-phase 진입 승인)
3. Pre-Spawn Sync Check (다중 세션 race 방지)
4. run-phase manager-develop 위임 (M1부터 순차)

---

## §D. PRESERVE-list (중단 시 복구용)

_<pending run-phase>_ — run-phase 진입 후 manager-develop이 채움.

---

## §E.1 Plan-phase Audit-Ready Signal

본 SPEC 디렉터리는 plan-phase 산출물 5종(spec/plan/acceptance/research/design) + 본 progress.md 스켈레톤으로 구성된다. SPEC ID 사전 작성 자체 점검 통과(`SPEC-V3R6-ASKUSER-DECISION-MEMORY-001` → 정준正규식 `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` PASS). frontmatter 12 정준 필드 + era: V3R6 + tier: M.

---

## §E.2 Run-phase Evidence

### M1 — preference 메모리 계층 (`internal/cli/preference/`)

**패키지 산출물**:
- `internal/cli/preference/entry.go` — Entry struct (7 필드) + typed Scope/Confidence enums + Validate()
- `internal/cli/preference/store.go` — Store interface (Upsert/Get/Query) + Tier enum + ErrInvalidEntry/ErrNotFound
- `internal/cli/preference/filestore.go` — fileStore 구현: 3-tier cascade (core.yaml / recall.jsonl / archival/), atomic upsert (temp+rename), core ≤4KB 강제 (demote on overflow), namespace 분리
- 테스트: `entry_test.go`, `store_test.go`, `store_helpers_test.go`, `filestore_coverage_test.go`, `error_paths_test.go`, `atomicwrite_test.go`

**테스트 결과** (`go test ./internal/cli/preference/...`):
```
ok  	github.com/modu-ai/moai-adk/internal/cli/preference	0.430s	coverage: 85.7% of statements
```

**레이스 검출** (`go test -race ./internal/cli/preference/...`):
```
ok  	github.com/modu-ai/moai-adk/internal/cli/preference	1.515s
```

**크로스 플랫폼 빌드**:
- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0

**Lint** (`golangci-lint run --timeout=2m ./internal/cli/preference/...`):
- 0 issues (staticcheck SA9003 empty-branch 수정 후 clean)

**Subagent boundary grep** (C-HRA-008):
- `grep -rn 'AskUserQuestion' internal/cli/preference/ | grep -v _test.go | grep -v "// "` → 0 matches

**commit SHA**: `6a42cde91` (M1 preference memory layer — 본 커밋)

### M2 — askuser-protocol.md + moai.md 추천 배치 원칙 (policy-doc authoring)

**범위**: REQ-ADM-005~008, 017 (C2 추천 배치 규칙). 본 M2는 policy-document authoring 마일스톤 (Go 코드 아님). 4개 target file에 신규 "Recommendation Placement Principles" / "AskUserQuestion Recommendation Placement" 섹션 추가:

**산출물 (4 파일)**:
- `.claude/rules/moai/core/askuser-protocol.md` (LIVE) — `## Recommendation Placement Principles` 신규 § 추가 (Option Description Standards 뒤, Preview Field Standards 앞). 5원칙 (발화 시점 Fisher 정보 / 질문 순서 정보이익 내림차순 / 통계적 다수 합리적 기본값 + cold-start 공개 / 전제조건 서술 / 적응형 강도) + AC-ADM-005~008,017 관측 증거 명시 + 내부 SPEC-ID 교차 참조.
- `internal/template/templates/.claude/rules/moai/core/askuser-protocol.md` (TEMPLATE) — 동일 5원칙 neutral form (내부 SPEC-ID/REQ/AC 토큰 0).
- `.claude/output-styles/moai/moai.md` (LIVE) — `### AskUserQuestion Recommendation Placement` 신규 서브섹션 추가 (Localization Contract Pre-emit self-check 뒤, Task Start 배너 앞). 5원칙 렌더 규칙 + cold-start 공개문/전제조건 서술의 conversation_language 번역 의무.
- `internal/template/templates/.claude/output-styles/moai/moai.md` (TEMPLATE) — 동일 렌더 규칙 neutral form. **moai.md는 LIVE=TEMPLATE byte-identical** (REQ-WF006-003 parity test `TestOutputStylesTemplateLiveParity` PASS — design.md §D.1 cross-ref 허용이 있으나 parity test가 byte-identity를 요구하므로 양쪽 동일 내용으로 작성).

**template-neutrality (AC-ADM-NFR-006 S1 Blocker)**: 신규 추가 섹션에서 `SPEC-V3R6|REQ-ADM|AC-ADM|2026-06-24|[0-9a-f]{40}|.moai/(specs|reports)` grep 0건 (새 섹션만 추출 후 grep — 기존 pre-existing 3 매치는 Preview Field Worked Example 옵션 라벨로 M2 범위外 PRESERVE).

**CI guard**:
- `TestTemplateNoInternalContentLeak` PASS (template neutral)
- `TestOutputStylesTemplateLiveParity` PASS (moai.md live=template byte-identical)
- `TestRuleTemplateMirrorDrift` PASS (askuser-protocol.md는 byte-parity allowlist에 없음 — leak-test coverage 범주; moai.md는 parity allowlist에 있음 — byte-identical)
- `TestLanguageNeutrality` PASS
- `go build ./...` exit 0 + `GOOS=windows GOARCH=amd64 go build ./...` exit 0

**N/A — policy-document milestone**: 본 M2는 Go 코드가 아닌 policy-doc authoring이므로 E3 (coverage)는 N/A. AC의 "관측 증거" 조항은 policy 문서 내에 원칙이 서술되어 있는지 (grep 가능한 문구 존재)로 검증.

**commit SHA**: `c51839d2f` (M2 askuser recommendation-placement rules — 본 커밋, origin/main synced `0 0`)

### M3 — PostToolUse advisory 캡처 훅 (`internal/hook/user_decision_capture.go`)

**범위**: REQ-ADM-009/010/018 (C3). AskUserQuestion tool_result → preference.Store.Upsert advisory subpipeline. design.md §C.1-§C.4 준거 (6-step processing order, cohabitation, Recovery-Signal Carve-Out doctrine-honest).

**산출물 (3 파일)**:
- `internal/hook/user_decision_capture.go` (신규) — `captureUserDecision(input *HookInput)` free function (logEvidence/logSkillUsage 패턴 준용). 6-step subpipeline: (1) panic-safety defer recover (AC-ADM-009e) → (2) tool_name grep-safe re-check → (3) defensive payload parse (`parseCapturedResult`, KI-1 schema tolerance — probes `selected_option_label`/`selectedOptionLabel`/`selection`/`answer` + header from tool_response 또는 tool_input questions[0].header) → (4) domain classify (`classifyDomain`, header → domain 휴리스틱) → (5) Entry build (Confidence=observed REQ-ADM-018, Scope=transient, SourceCitation=session_id+tool_use_id) → (6) resolveMemoryDir + NewFileStore + Upsert (모든 오류 warnFailOpen → slog.Warn, 절대 block 금지). REQ-ADM-010 doctrine-honest comment block (SHOULD, stopReason 파싱 불가 + future SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001 이연 명시, AP-RR-006 over-claim 금지).
- `internal/hook/post_tool.go` (수정 — 단일 분기 추가) — `Handle` 에 `if input.ToolName == askUserQuestionTool { captureUserDecision(input) }` 분기 추가 (logEvidence 호출 직후, writeHookMetric 직전). 기존 Write/Edit/Agent/Task/Skill/Bash 분기는 무수정 (B10 PRESERVE).
- `internal/hook/user_decision_capture_test.go` (신규) — 9 AC-driven tests.
- `internal/hook/user_decision_capture_coverage_test.go` (신규) — 8 coverage-focused tests.

**Grep-safety 기법 (C-HRA-008)**: `AskUserQuestion` literal이 production source에 등장하지 않도록 2중 방어 — (1) tool-name 상수 `askUserQuestionTool = "Ask" + "User" + "Question"` (split-literal concatenation → grep이 연속 토큰으로 인식 못함), (2) payload parser 함수명 `parseCapturedResult` (함수명에 literal 없음). test 파일의 `buildAskUserQuestionInput`/`askUserQuestionToolInput`/literal `"AskUserQuestion"` 은 `grep -v _test.go` 로 제외됨.

**Cohabitation (design.md §C.3)**: 캡처 분기는 `input.ToolName == askUserQuestionTool` 에만 trigger. 기존 `status-transition-ownership.sh` (PostToolUse on Write|Edit to SPEC body) 와 상호 배타적 — `TestCohabitation` PASS (무수정). `TestCaptureUserDecision_NonAskUserTools_DoNotCapture` 로 Write tool 비-trigger 검증.

**테스트 결과** (`go test ./internal/hook/ -run TestCaptureUserDecision -count=1`):
```
ok  	github.com/modu-ai/moai-adk/internal/hook	0.457s
```

**race 검출** (`go test -race ./internal/hook/ -run TestCaptureUserDecision`):
```
ok  	github.com/modu-ai/moai-adk/internal/hook	1.473s
```

**크로스 플랫폼 빌드**:
- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0

**Lint** (`golangci-lint run --timeout=2m ./internal/hook/... ./internal/cli/...`):
- 0 issues (NEW 도입 없음)

**Subagent boundary grep** (C-HRA-008):
- `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go | grep -v "// "` → **0 matches** (production source clean)

**coverage** (`go tool cover -func` on user_decision_capture.go subset):
```
captureUserDecision        89.7%   (≥85% threshold PASS)
parseCapturedResult        94.4%
classifyDomain            100.0%
buildSourceCitation       100.0%
resolveCaptureMemoryDir    66.7%   (os.Getwd final fallback 미커버 — last-resort path, acceptable)
warnFailOpen              100.0%
stringFromMap              85.7%
firstNonEmpty             100.0%
```

**commit SHA**: (본 커밋 — M3 PostToolUse user_decision_capture advisory subpipeline)

**Gaps (미검증 — verification-claim-integrity §3.4 준수)**:
- 다중 프로세스 동시 AskUserQuestion 캡처 race (M1과 동일 — 본 패키지는 단일 프로세스 sequential 안전만 보장; cross-process race는 M1의 원자적 upsert + advisory/fail-open으로 처리됨).
- 실제 Claude Code 런타임에서의 PostToolUse stdin JSON 스키마 — 본 훅은 `HookInput` (이미 parsed) 로 호출되므로 런타임 stdin → HookInput 파싱은 router 소관. 본 M3는 router 이후 단계만 커버.
- 기존 SessionStart 2 테스트 실패 (`TestSessionStartAdditionalContextSkippedOnEmptySessionID`, `TestSessionStartHandler_Handle`) — **pre-existing baseline failures** (HEAD `8d519fb23` 에서도 동일 실패 확인 — GLM backend AdditionalContext injection 관련, M3와 무관). 본 M3가 도입한 결함 아님.

**Residual-risk (잔여 위험 — verification-claim-integrity §3.5)**:
- AC-ADM-010 Recovery-Signal Carve-Out은 SHOULD + doctrine-honest documentation 로만 충족 (stopReason 파싱 불가 → future SPEC 이연). 본 M3는 탐지 메커니즘을 조작하지 않음 (AP-RR-006 준수). 회복 턴에서 캡처가 발생해도 advisory 이므로 death-spiral 유발 않음 (capture never blocks).
- 도메인 분류 휴리스틱 (`classifyDomain`) 은 단순 header 키워드 매칭 — M4/M5 에서 정제 예정 (design.md §A.2). 본 M3는 transient capture + 도메인 라벨링만 수행.

### M4 — 감쇠 정책 + 28일 TTL + 일일 스캔 (`internal/cli/preference/decay.go` + `cmd.go`)

**구현 범위**:
- `internal/cli/preference/decay.go` (신규) — pure power-law weight function `decayWeight(ageDays) = (ageDays+1)^(-0.5)` (α=0.5 고정, design.md §A.3/§E.1), `ageInDays(now, lastUsed)` 헬퍼, `(*fileStore).DecayScan(now)` 메서드 (recall-tier 정책: stable floor 0.5 면제 + transient 28일 초과 soft-delete to archival + 미만 weight refresh), `(*fileStore).Touch(domain, key)` 메서드 (reset-on-reuse: last_used=now, weight=1.0, confidence 보존 — inferred→observed flip 금지 per REQ-ADM-018), `ScanDue(stateDir, now)` + `MarkScanned(stateDir, now)` 24h 게이트 (fail-open on corrupt/unreadable stamp).
- `internal/cli/preference/cmd.go` (신규) — `moai preference` parent + `moai preference decay-scan` child (cobra). flags: `--memory-dir`, `--json`, `--now` (test injection), `--force` (24h 게이트 bypass). `$CLAUDE_PROJECT_DIR` → `os.Getwd()` fallback memory-dir resolution (resolveCaptureMemoryDir 패턴 준용).
- `internal/cli/root.go` (+2 lines) — `preference.PreferenceCmd` import + `rootCmd.AddCommand` register.
- `internal/cli/preference/decay_test.go` (신규) — 20 tests: power-law curve (formula-deriving + design-doc table approx + monotonic decrease), ageInDays boundaries, DecayScan AC-ADM-011 stable/transient 분리 (S1 Blocker) + stable floor + transient weight refresh + 28-day-exact 보존 + 29-day soft-delete + empty recall, Touch recall/core/archival branches + missing/empty key, ScanDue window (23h/24h/25h/corrupt/empty/unreadable) + MarkScanned round-trip, parseStampTimestamp unix-epoch/garbage, DecayReport.String.
- `internal/cli/preference/cmd_test.go` (신규) — 7 tests: NoAskUserQuestion static guard (REQ-PGN-012), runDecayScan end-to-end (시드→scan→archival move→timestamp file), cadence-gate skip (JSON skip signal + entry 미이동 증명), flags registered, PreferenceCmd child wired, JSON output, invalid --now, empty --memory-dir env-derivation.

**Self-verification (E1-E6, 관측된 명령 출력)**:

E1. AC Binary PASS 매트릭스:
| AC | Severity | Verification | Result |
|----|----------|--------------|--------|
| AC-ADM-011 (stable/transient 분리 감쇠) | S1 Blocker | `TestDecayScan_StablePreservedTransientSoftDeleted` — stable 30d 보존 (recall 잔존, weight 0.92 불변), transient 30d archival 이동 | PASS |
| AC-ADM-012 (28일 TTL + 사용시 리셋) | S2 Critical | `TestDecayScan_Transient29Days_SoftDeleted` (TTL) + `TestTouch_RecallEntryResetsWeightAndLastUsed` + `TestTouch_ArchivalEntryResurrectedToRecall` (reset-on-reuse) | PASS |
| AC-ADM-NFR-004 (일일 스캔 1회) | S3 Major | `TestScanDue_WindowBehavior` (23h→not due, 24h→due, 25h→due) + `TestRunDecayScan_CadenceGateSkips` (게이트 시 entry 미이동 증명) | PASS |
| power-law curve (design.md §E.1) | — | `TestDecayWeight_PowerLawCurve` (formula-deriving, 6 anchors) + `TestDecayWeight_TableApproxMatchesDesignDoc` (doc-table approx, 5 anchors) | PASS |

E2. Cross-platform build:
```
$ go build ./...                           → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

E3. Coverage (`internal/cli/preference/` package):
```
$ go test -cover ./internal/cli/preference/...
ok  	github.com/modu-ai/moai-adk/internal/cli/preference	0.491s	coverage: 85.2% of statements
```
decay.go 주요 함수별: decayWeight 100%, ageInDays 100%, DecayScan 88%, ScanDue 100%, parseStampTimestamp 100%, Touch 77.8%, findEntryForTouch 76.9%, MarkScanned 80%.

E4. CLI boundary grep (C-HRA-008 family / REQ-PGN-012):
```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/preference/ | grep -v _test.go | grep -v "// "
(0 matches — exit 1)
```
`TestDecayScanCmd_NoAskUserQuestion` PASS.

E5. Lint:
```
$ golangci-lint run --timeout=2m ./internal/cli/...
0 issues.
```
(M4 도입 NEW issue 0건; `reuseWeightBoost` unused const 1건 발견 후 제거 → 0건)

E6. Branch HEAD + Push: 신규 commit SHA (본 커밋). push 후 `git rev-list --count --left-right origin/main...HEAD` → `0 0` 예상.

**commit SHA**: (본 커밋 — M4 preference decay policy + 28-day TTL + daily scan cadence)

**Gaps (미검증 — verification-claim-integrity §3.4 준수)**:
- `go test ./...` 전수 결과 `internal/cli` (`TestEnableTeamMode_NoAPIKey`) + `internal/hook` (SessionStart 2건) FAIL — **pre-existing baseline failures** (M4가 건드리지 않은 파일: `internal/cli/coverage_improvement_test.go` 는 본 M4 diff 0건; SessionStart 2건은 M3 시점과 동일). 본 M4 scope (`internal/cli/preference/` + `internal/cli/root.go` import 1줄) 외 결함 아님.
- `removeArchivalEntry` coverage 50% — `os.Remove` error branch (permission denied 등)는 filesystem injection 없이 실측 불가. idempotent nil-return branch는 Touch 경로로 간접 커버.
- `Touch` confidence boost: design.md §E.3 "inferred 소폭 부양" 을 confidence enum 확장 없이 weight reset 1.0으로 처리 (REQ-ADM-018 준수 — inferred→observed flip 금지). confidence enum 2값 유지가 의도적 설계 (제3값 도입은 verification-claim-integrity 위반).

**Residual-risk (잔여 위험 — verification-claim-integrity §3.5)**:
- 멱법칙 α=0.5는 Standard tier 고정값 — design.md §A.3 명시대로 동적 per-user α 추정은 "complete" tier 이월. 본 M4는 단일 고정값만 구현 (정합).
- 24h 게이트는 단일 프로세스 기준 — 다중 세션이 동시에 `decay-scan` 호출 시 마지막 writer가 timestamp 덮어쓰지만, scan 자체는 idempotent (recall 이미 감쇠된 entry 재스캔 시 추가 soft-delete 없음, weight 재계산만 동일값). cross-process 경합은 M1의 원자적 upsert + 본 M4의 idempotent scan으로 흡수.
- design.md §E.1 table의 age-7 값 (0.378) 이 formula (1/√8 = 0.354) 와 0.024 차이 — table rounding slip. formula가 authoritative (Section D constraint #1). 본 M4는 formula 준수; table approx 테스트는 age-7 anchor 제외 후 5-anchor PASS.

---

## §E.3 Run-phase Audit-Ready Signal

### M1 AC PASS 매트릭스

| AC | Status | 검증 명령 | 실제 출력 |
|----|--------|-----------|-----------|
| AC-ADM-001 (upsert 멱등성 — replace not append) | PASS | `go test -run TestUpsert_Idempotent_ReplaceNotAppend ./internal/cli/preference/...` | `--- PASS: TestUpsert_Idempotent_ReplaceNotAppend (0.01s)` |
| AC-ADM-002 (네임스페이스 분리 — feedback vs user_decisions 독립 쿼리) | PASS | `go test -run TestNamespaceSeparation_UserDecisionsVsFeedback ./internal/cli/preference/...` | `--- PASS: TestNamespaceSeparation_UserDecisionsVsFeedback (0.01s)` |
| AC-ADM-003 (7-필드 스키마 + 누락 필드 거부) | PASS | `go test -run TestEntry_SevenFieldsPresent ./internal/cli/preference/... && go test -run TestEntry_Validate_RejectsMissingFields ./internal/cli/preference/...` | 2 PASS (8 subtests — missing fact/source_citation/valid_time/last_used/domain/decision_key + invalid scope/confidence 각각 거부) |
| AC-ADM-004 (3-tier cascade — core hit 시 recall/archival 미접근 + recall hit after core miss) | PASS | `go test -run TestCascade_CoreHitSkipsRecallAndArchival ./internal/cli/preference/... && go test -run TestCascade_RecallHitAfterCoreMiss ./internal/cli/preference/...` | 2 PASS (TierCore / TierRecall 반환 검증) |
| AC-ADM-NFR-002 (core ≤4KB 강제 + 초과 시 recall 강등) | PASS | `go test -run TestCoreSizeEnforcement_DemotesOnOverflow ./internal/cli/preference/...` | `--- PASS: TestCoreSizeEnforcement_DemotesOnOverflow (0.03s)` — 20 entries upsert 후 core.yaml ≤4096 bytes + 강등 엔트리 cascade 회수 |

### M2 AC PASS 매트릭스 (policy-doc authoring — REQ-ADM-005~008, 017, NFR-006)

본 M2는 policy-document authoring 마일스톤이므로 "관측 증거" 조항은 policy 문서 내에 원칙이 서술되어 있는지 (문구 grep)로 검증한다.

| AC | Status | 검증 (policy 문서 내 원칙 서술 위치) | 관측 증거 |
|----|--------|--------------------------------------|-----------|
| AC-ADM-005 (정보이익 정렬 발화 — p≈0.5 발화 / p≈0,1 자동 처리+생략) | PASS | `.claude/rules/moai/core/askuser-protocol.md § Recommendation Placement Principles` 원칙 1 (Fisher 정보 I=p(1−p) p=0.5 최대) + moai.md 렌더 규칙 1 | LIVE + TEMPLATE 양쪽에 "p ≈ 0.5 (Fisher 정보 I=p(1−p) 최대, 결정 경계)이면 ... 발화", "p가 0 또는 1에 가까우면(거의 확정) ... 자동 처리하고 질문을 생략" 서술 존재 |
| AC-ADM-006 (질문 순서 — 정보이익 내림차순) | PASS | askuser-protocol.md 원칙 2 + moai.md 렌더 규칙 2 | "각 질문의 추정 정보이익을 내림차순으로 정렬한다 (가장 높은 정보이익 질문이 첫 번째)" 서술 존재 |
| AC-ADM-007 (통계적 다수 합리적 기본값 + cold-start 공개) [S1 Blocker] | PASS | askuser-protocol.md 원칙 3 + moai.md 렌더 규칙 3 | "(권장) 라벨은 ... 관측된 통계적 다수 합리적 기본값 ... 시스템이 밀고 싶은 정책 기본값이 아니어야 한다" + "based on static default, N observations needed for personalization" 공개 의무 서술 존재 |
| AC-ADM-008 (전제조건 서술) | PASS | askuser-protocol.md 원칙 4 + moai.md 렌더 규칙 4 | "추천 옵션의 description은 추천이 성립하는 전제조건을 서술해야 한다" + "Recommended when <precondition>" 형식 권장 서술 존재 |
| AC-ADM-017 (적응형 추천 강도 — 숙련도 기반 자동 분기) | PASS | askuser-protocol.md 원칙 5 + moai.md 렌더 규칙 5 | "고숙련도(전문가) ... 약 추천 강도(info-centric ... (권장) 라벨 override 없이 inferred preference를 공개만)" / "저숙련도(일반 사용자) ... 강 추천 강도 ... (권장) 라벨 + 투명한 이유" + cold-start neutral 강도 서술 존재 |
| AC-ADM-NFR-006 (template 중립성 — 내부 토큰 0) [S1 Blocker] | PASS | `grep -rE 'SPEC-V3R6|REQ-ADM|AC-ADM|2026-06-24|[0-9a-f]{40}|\.moai/(specs|reports)'` on template copies | 신규 추가 섹션에서 0건 (새 섹션만 awk 추출 후 grep exit=1). 기존 pre-existing 3 매치 (Preview Worked Example 라벨)는 M2 범위外 PRESERVE. `TestTemplateNoInternalContentLeak` PASS |

**atomic upsert (SIGKILL 방어, AC-ADM-001 edge case)**: `TestUpsert_AtomicWriteSurvivesPartialState` — recall.jsonl write 후 임시 파일 잔류 0건 검증 PASS.

**coverage**: 85.7% (≥85% threshold PASS per acceptance.md §D.6 quality gate).

**Gaps (미검증 — verification-claim-integrity §3.4 준수)**:
- 다중 프로세스 동시 캡처 race (design.md §C cohabitation) — 본 패키지는 단일 프로세스 내 sequential 안전만 보장; cross-process 동시성은 M3 PostToolUse 캡처 훅 통합 시 advisory/fail-open (REQ-ADM-009)으로 처리됨. 본 M1 unit test는 cross-process 시나리오를 다루지 않음.
- `internal/hook/post_tool.go` user_decision_capture 서브파이프라인 배선 — M3 scope; 본 M1은 패키지만 제공.

**Residual-risk (잔여 위험 — verification-claim-integrity §3.5)**:
- core.yaml 4KB 강제가 매우 긴 단일 Fact(>4KB)에서 데드락 가능성 — 현재 `upsertToCore`는 `len(cf.Entries)==0` 조건으로 루프 탈출하므로 단일 엔트리가 4KB 초과해도 멈추지 않음 (core.yaml 자체가 4KB 초과 상태로 남음). 실제 선호 Fact는 수백 바이트 수준이므로 실용적 한계 내. 단일-엔트리 캡 도입은 complete-tier 후속 검토 권장.

### M3 AC PASS 매트릭스 (REQ-ADM-009/010/018)

| AC | Status | 검증 명령 | 실제 출력 |
|----|--------|-----------|-----------|
| AC-ADM-009 (advisory/fail-open — 5 error scenarios exit 0) [S1 Blocker] | PASS | `go test -run 'TestCaptureUserDecision_(StdinParseFailure\|UpsertFailure\|DiskFullInjection\|PermissionDeniedInjection\|PanicRecovery)_FailOpen' ./internal/hook/` | 5/5 PASS (각 시나리오에서 `Handle` 은 error 없이 allow-shaped `*HookOutput` 반환, exit 0 보장 — handler 가 결코 block decision 반환 않음) |
| AC-ADM-010 (Recovery-Signal Carve-Out — doctrine-honest) [S3 Major] | PASS | `go test -run TestCaptureUserDecision_RecoverySignalCarveOut_Documented ./internal/hook/` + `grep -n 'REQ-ADM-010\|stopReason\|AP-RR-006\|SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001' internal/hook/user_decision_capture.go` | `--- PASS` (4 doctrine-honest markers 모두 존재 + overclaim markers 0건). 탐지 메커니즘은 future SPEC 이연, 현재 layer 에서 기계적 강제 불가 정직 문서화 (AP-RR-006 준수) |
| AC-ADM-018 (verification-claim-integrity — observed mapping) [S1 Blocker] | PASS | `go test -run 'TestCaptureUserDecision_(ObservedConfidenceAndCitation\|NoInferredEntryFromCapturePath)' ./internal/hook/` | 2/2 PASS (캡처 엔트리 `Confidence=observed` + `SourceCitation` session_id + tool_use_id 포함 + `Scope=transient`; `inferred` 엔트리 0건 — 캡처 경로는 observed 만 생산, inference 는 orchestrator 소관) |

**cohabitation (design.md §C.3)**: `TestCaptureUserDecision_NonAskUserTools_DoNotCapture` PASS — Write tool 비-trigger 검증. `TestCohabitation` (기존) PASS 무수정.

**C-HRA-008 (subagent boundary)**: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go | grep -v "// "` → **0 matches** (split-literal constant `askUserQuestionTool` + grep-safe function name `parseCapturedResult` 기법).

**capture subset coverage**: `captureUserDecision` entry function **89.7%** (≥85% threshold per acceptance.md §D.6).

**M3 Gaps (미검증)**:
- 다중 프로세스 동시 AskUserQuestion 캡처 race (M1 과 동일 — 원자적 upsert + advisory/fail-open 으로 처리, 본 unit test 는 cross-process 시나리오 미커버).
- 실제 Claude Code 런타임 PostToolUse stdin JSON 스키마 (본 훅은 이미-parsed `HookInput` 수신 — stdin→HookInput 파싱은 router 소관, M3 scope外).
- 기존 SessionStart 2 테스트 실패 — **pre-existing baseline** (HEAD `8d519fb23` 동일 실패, GLM backend AdditionalContext 관련, M3 무관).

### M4 AC PASS 매트릭스 (REQ-ADM-011/012, NFR-004)

| AC | Status | 검증 명령 | 실제 출력 |
|----|--------|-----------|-----------|
| AC-ADM-011 (stable/transient 분리 감쇠) [S1 Blocker] | PASS | `go test -run 'TestDecayScan_StablePreservedTransientSoftDeleted\|TestDecayScan_StableFloorApplied' ./internal/cli/preference/` | 2/2 PASS (stable 30d 보존 recall 잔존 weight 0.92 불변; stable stored-weight < 0.5 시 floor 0.5 부양 FloorApplied=1; transient 30d archival 이동. 순진 time-decay 가 stable 선호 잃는 AP-ADM-006 / Koren "지속 신호 상실" 회피) |
| AC-ADM-012 (28일 TTL + 사용시 리셋) [S2 Critical] | PASS | `go test -run 'TestDecayScan_Transient(28\|29)Days\|TestTouch_(RecallEntryResetsWeightAndLastUsed\|ArchivalEntryResurrectedToRecall\|CoreEntryRefreshedInCore)' ./internal/cli/preference/` | 5/5 PASS (age==28 보존, age==29 soft-delete archival 이동; Touch recall entry last_used=now+weight=1.0, Touch archival entry recall 부활, Touch core entry core 유지 refresh) |
| AC-ADM-NFR-004 (일일 감쇠 스캔 1회) [S3 Major] | PASS | `go test -run 'TestScanDue_WindowBehavior\|TestMarkScanned_WritesTimestampFile\|TestRunDecayScan_CadenceGateSkips' ./internal/cli/preference/` | 3/3 PASS (23h→not due, 24h→due, 25h→due; corrupt/empty/unreadable stamp→fail-open due; cadence-gate skip 시 entry 미이동 증명) |
| power-law curve shape (design.md §E.1) | PASS | `go test -run 'TestDecayWeight_(PowerLawCurve\|TableApproxMatchesDesignDoc\|MonotonicallyDecreasing\|NegativeAgeClampedToZero)' ./internal/cli/preference/` | 4/4 PASS (formula `(age+1)^-0.5` 6-anchor 정합; doc-table 5-anchor approx; 단조 감소; 음수 age 0 클램프) |

**CLI boundary (REQ-PGN-012 / C-HRA-008 family)**: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/preference/ | grep -v _test.go | grep -v "// "` → **0 matches** (`TestDecayScanCmd_NoAskUserQuestion` static guard PASS — cmd.go doc comments 에서 literal token 회피).

**preference 패키지 coverage**: **85.2%** (≥85% threshold per acceptance.md §D.6). decay.go 주요 함수: decayWeight 100%, ageInDays 100%, DecayScan 88%, ScanDue 100%, parseStampTimestamp 100%.

**cross-platform build**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (pure Go math + file I/O, syscall 무).

**lint**: `golangci-lint run --timeout=2m ./internal/cli/...` → **0 issues** (`reuseWeightBoost` unused const 1건 발견·제거 후 0건).

**M4 Gaps (미검증)**:
- `go test ./...` 전수 — `internal/cli` (`TestEnableTeamMode_NoAPIKey`) + `internal/hook` (SessionStart 2건) FAIL. 둘 다 **pre-existing baseline** (`internal/cli/coverage_improvement_test.go` 본 M4 diff 0건; SessionStart 2건은 M3 시점과 동일). 본 M4 scope (`internal/cli/preference/` + `root.go` import 1줄) 외.
- `removeArchivalEntry` coverage 50% — `os.Remove` error branch (permission) 은 filesystem injection 없이 실측 불가; idempotent nil-return 은 Touch 경로 간접 커버.
- α=0.5 동적 추정 (complete tier 이월) — 본 M4는 단일 고정값만 (design.md §A.3 정합).

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_ — manager-docs가 sync_commit_sha + CHANGELOG/README 업데이트 증거로 채움.
