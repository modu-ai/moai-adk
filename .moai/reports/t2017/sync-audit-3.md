auditor-model: claude-opus-5-5
verdict: PASS-WITH-DEBT
audited_sha: 400caf47ba9ad1650148ec0b37aae23fb13fe304
overall_score: 0.85
functionality_score: 88
security_score: 82
craft_score: 82
consistency_score: 84
must_pass_failed: 0
blocking_count: 0
receipts: none

# Sync 감사 판정 — SPEC-AUDIT-ENGINE-CONVERGENCE-001 (카드 t2017), 3회차

## 요약

판정은 PASS-WITH-DEBT, 가중 점수 84.7(0.85)이다. 차단 결함은 없다. 1회차 D1과 2회차 D10은 둘 다 닫혔다. 수정 코드를 읽었고, 지정 테스트를 다시 돌렸고, 스크래치 사본에서 D10 뮤턴트를 직접 만들어 테스트가 잡는 것을 확인했다. main 흡수 머지(a1c76f1e99)는 카드 변경을 한 줄도 바꾸지 않았다. 다른 부분은 `internal/template/catalog.yaml`의 해시 4줄뿐이고, 해시와 파일이 일치한다. 400caf47ba의 테스트 수리도 REQ-AEC-003을 실제로 단언한다. diff 재료를 지운 뮤턴트에서 해당 테스트가 실패했다.

부채로 남긴 항목은 모두 비차단이다. 이전 감사의 D2~D9·D11·D12가 그대로 열려 있고, 이번에 새로 D13·D14(영수증 신선도 판정의 섹션 인식 한계, 가설 수준 노출)와 D15(progress.md §E.4 산문이 사실과 어긋남)를 올렸다.

이 판정은 리더 지시와 운영자 선택(리더 원장 `decisions.md` R503, 카드당 Claude sync-auditor 1회, Codex·유료 호출 없음)에 따라 Claude 단독으로 냈다. 이 트리의 감사 계획은 codex를 `required`로 둔다. 그래서 codex 필수 게이트는 이번 감사에서 충족되지 않았다. 이 판정을 그 게이트의 대체로 받을지는 운영자와 리더가 정할 일이다(아래 Gaps G1).

## 차원별 점수 (default 프로파일, flat 가중)

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%, must-pass) | 88/100 | PASS | AC 16개를 모두 테스트로 확인했다(아래 AC 표). 유일한 문자 편차는 EV-01 green 셀이다(D9, `codex_audit` 직접 경로이고 REQ-AEC-003 범위 밖). |
| Security (25%, must-pass) | 82/100 | PASS | D10 우회 경로가 닫혔다(Stop 경로 프로브 15종 PASS, 뮤턴트 kill). 새로 찾은 D13·D14는 판정 대상 커밋에 특정 문서 형태가 이미 있어야 하는 가설 수준 노출이다. 현재 코퍼스에서 D13 형태는 0건, D14 형태는 11개 파일이다. |
| Craft (20%) | 82/100 | PASS | golangci-lint `0 issues.`, gofmt 출력 없음, Windows 빌드 exit 0, `internal/auditreceipt` 커버리지 87.1%. 들여쓰기 규칙 뮤턴트는 여전히 살아남는다(D11). 낡은 주석 2곳도 그대로다(D8). |
| Consistency (15%) | 84/100 | PASS | 문서·미러 쌍이 동일하고, error_class 집합이 문서와 상수에서 같고, CHANGELOG 항목은 1건이며, catalog 해시가 맞다. 다만 §E.4 산문이 사실과 다르다(D15). |

가중 합: 88×0.40 + 82×0.25 + 82×0.20 + 84×0.15 = 84.7. 두 must-pass 차원 모두 기준을 넘는다.

## 결함 목록

차단(blocking) 결함 없음.

| id | 심각도 | 분류 | 위치 태그 | 위치 | 내용 | 상태 |
|---|---|---|---|---|---|---|
| D1 | (High였음) | blocking였음 | in-diff | `internal/auditreceipt/store.go:1919-1964` | 중첩 fence가 §E.2를 가리던 문제 | 닫힘. `fenceRun`이 문자와 길이를 추적하고, 닫는 fence는 같은 문자·같거나 긴 길이·뒤 공백만 허용한다. Stop 경로 테스트 `D1-four-backtick-holding-three` 등 PASS |
| D10 | (High였음) | blocking였음 | in-diff | `store.go:1887-1905` | 파서가 헤딩을 놓치면 빈 목록끼리 같다고 판정하던 문제 | 닫힘. `e2DocsUnchanged`가 원문 헤딩 수와 파싱 결과 수를 양쪽에서 비교해 다르면 실패 처리한다. 조건을 `if false`로 바꾼 뮤턴트에서 `TestE2DocsUnchanged_E2BodyEditIsAlwaysAChange` FAIL |
| D2 | SHOULD-FIX | optional | in-diff | `internal/cli/mcp_convergence.go` (이어받은 claude 레그의 gate) | 이어받은 레그가 이전 호출의 gate를 유지한다 | 열림, 코드 변경 없음. 이 저장소는 claude gate가 `off`라 해당하지 않는다 |
| D3 | SHOULD-FIX | optional | in-diff | `internal/cli/mcp_audit_multi_record.go:122`, `internal/closure/records.go:71` | `VerdictDigests`를 기록만 하고 다시 읽지 않는다 | 열림. 재측정 결과 비테스트 사용처가 쓰기 1곳과 필드 정의 1곳뿐이다. §E B1 잔여 위험 범위 |
| D4 | MINOR | optional | in-diff | `internal/hook/audit_receipt_guard.go` | spawn 거부 사유에 error class가 없다 | 열림, 재확인하지 않음 |
| D5 | MINOR | optional | in-diff | `internal/cli/verify_clear_rejection.go` | `.moai/reports` 자체가 트리 밖 symlink이면 containment를 통과한다 | 열림. §E B4가 보안 경계가 아니라고 명시한 범위 |
| D6 | SHOULD-FIX | optional | in-diff | `mcp_convergence.go` (이어받은 codex 영수증) | 일부 레그만 재실행하면 오래된 영수증이 인용된다(fail-closed) | 열림, 실행하지 않음 |
| D7 | MINOR | optional | in-diff | `internal/cli/mcp_server.go:569` | codex_audit 설명 문구가 `applyGateUnmet`의 실제 동작과 다르다 | 열림. 머지 뒤 줄 번호가 566에서 569로 바뀌었고 문구는 그대로다 |
| D8 | MINOR | optional | in-diff | `mcp_convergence.go:134`, `:1352` | 주석이 코드와 맞지 않는다 | 열림, 재측정함 |
| D9 | MINOR | optional | in-diff | `internal/cli/audit_delta_scope.go:106`, `mcp_codex_audit.go:2414` | EV-01 green 셀(grep exit 1)이 문자 그대로는 충족되지 않는다. `codex_audit` 직접 adversarial 경로는 여전히 "prose-only scope"를 보고한다 | 열림. `audit_multi` 레그(`mcp_convergence.go:687-711`)는 서버가 수집한 diff를 싣고, 이 문구를 쓰지 않는다 |
| D11 | MINOR | optional | in-diff | `store.go:1953`, 테스트 | "4칸 이상 들여쓴 줄은 fence가 아니다" 규칙을 고정하는 테스트가 없다 | 열림. 조건을 지운 뮤턴트에서도 `go test ./internal/auditreceipt/` 결과가 `ok` |
| D12 | MINOR, 가설 | optional | touched-file | `store.go:1804` | `## §E.2.1` 같은 헤딩이 앞선 §E.2를 끝낸다 | 열림, 변경 없음 |
| D13 | MINOR (Low~Medium), 측정됨 | optional | in-diff | `store.go:1919-1946` | §E.2 안의 HTML `<pre>` 블록에 `## ` 줄이 있으면 파서는 거기서 §E.2를 끝낸다. CommonMark는 그 줄을 헤딩으로 보지 않는다. 그 아래 AC 행을 고쳐도 변경 없음으로 판정된다. 판정 대상 커밋에 이 형태가 이미 있어야 하고, 코퍼스에는 `^<pre`로 시작하는 줄이 있는 progress.md가 0개다 | 새로 발견. 원하면 HTML 블록 1·2형(`<pre`, `<!--`)도 fence처럼 추적하거나, §E.2 본문 끝을 "다음 `## ` 헤딩"에서 "원문 헤딩 수가 일치할 때만 신뢰"로 좁힌다 |
| D14 | MINOR, 측정됨 | optional | in-diff | `store.go:1804`, `:1887-1893` | 정규 형태(`## §E.2` + 공백)의 헤딩이 하나도 없는 progress.md는 보호할 §E.2가 없는 것으로 읽힌다. 그래서 파일 전체 수정이 증거 전용 커밋으로 받아들여진다. 해당하는 형태는 `### §E.2 …`(코퍼스 11개 파일), ` ## §E.2`(1~3칸 들여쓰기), `##\t§E.2`, 제목 없는 `## §E.2\r`이다. 섹션 맵 SSOT(`spec-frontmatter-schema.md`)가 정규 헤딩을 `## §E.2 Run-phase Evidence`로 정하고 있어 SPEC 위반으로 보지는 않는다 | 새로 발견. 원하면 `§E.2` 토큰이 들어간 헤딩 형태 줄이 있는데 정규 헤딩이 0개인 문서는 실패 처리로 닫는다 |
| D15 | MINOR | optional | in-diff | `.moai/specs/SPEC-AUDIT-ENGINE-CONVERGENCE-001/progress.md:274` | "No sync audit has run, and none is approved."라고 적혀 있지만 sync 감사 1·2회차가 실제로 실행되었다(`sync-audit-1.md`, `sync-audit-2.md`) | 새로 발견. `sync_commit_sha` 백필 커밋에서 함께 고친다 |

반복 수리 중지 규칙: D10은 이전 수리(D1)에서 생긴 결함이었고, 이번 감사에서 새로 생긴 차단 결함은 없다. 따라서 연속 2회 조건은 충족되지 않는다.

## AC 판정 (16개)

| AC | 판정 | 근거 (이번 실행) |
|---|---|---|
| AC-AEC-001a | PASS | `TestAEC001a_*` 3개 PASS. 수리된 `TestPerformCodexAudit_DeltaBaseRidesThePrompt` PASS. EV-01 grep은 아직 `:106`을 출력한다(D9) |
| AC-AEC-001b | PASS | `TestAEC001b_*` 7개 PASS |
| AC-AEC-001c | PASS | `TestAEC001c_*` 2개 PASS (하위 테스트 포함) |
| AC-AEC-001d | PASS | `TestAEC001d_*` 5개 PASS |
| AC-AEC-001e | PASS | `TestAEC001e_*` 1개 PASS |
| AC-AEC-001f | PASS | `TestAEC001f_*` 9개 PASS |
| AC-AEC-002a | PASS | `TestAEC002a_*` 5개 PASS. error_class 상수 5개가 문서 집합과 같다 |
| AC-AEC-002b | PASS | auditverdict·runtime·auditreceipt 패키지 전체 `ok`. hook·homestate 선택 실행 `ok`(Inconclusive 테스트 26개 이름 확인). 규약 라벨 줄에 INCONCLUSIVE가 있다 |
| AC-AEC-002c | PASS | `TestAEC002c_*` 12개 PASS |
| AC-AEC-003a | PASS | hook `TestAEC003a_EvidenceOnlyCommits_KeepTheReceipt` 하위 4개와 `TestAEC003a_UncommittedStateAlreadyJudged_KeepsTheReceipt` PASS |
| AC-AEC-003b | PASS | hook `TestAEC003b_AnyOtherChange_MakesTheReceiptStale` 하위 16개, `TestAEC003b_E2FenceShapes_*` 하위 15개, `TestAEC003b_HeadUnchanged_*` 하위 3개 PASS |
| AC-AEC-004a | PASS | `TestAEC004a_*` 3개 PASS |
| AC-AEC-004b | PASS | `TestAEC004b_*` 3개와 `TestClearRejection_*` 6개 PASS |
| AC-AEC-004c | PASS | `TestAEC004c_*` 3개 PASS |
| AC-AEC-005 | PASS | template `TestAEC005_*` 4개 PASS. EV-10·EV-11 각 0(exit 1). EV-17 각 1. EV-20 세 명령을 HEAD에서 실행한 결과: 1·2번 출력 없음, 3번 네 파일 각 1. `make agents-emit-check` → `specs=12 copies=4 diff=0` |
| AC-AEC-006 | PASS | EV-12 두 파일 모두 `:210` 일치. `cmp` exit 0. 미러에는 `<SPEC-ID>` 자리표시자만 있다. homestate `TestFR_ConventionCarriesTheRepairScopeRule` 포함 선택 실행 `ok` |

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

1. HEAD는 400caf47ba이고 트리는 깨끗하다.
2. D1과 D10은 닫혔다.
3. 머지 a1c76f1e99는 카드 변경을 보존했고, catalog 해시는 파일과 일치한다.
4. 400caf47ba의 테스트는 REQ-AEC-003의 diff 재료를 단언한다.
5. AC 16개가 통과한다.
6. lint, 포맷, 빌드가 깨끗하다.

### Evidence (명령 → exit → 판정에 쓴 출력)

- `git rev-parse HEAD` → `400caf47ba9ad1650148ec0b37aae23fb13fe304`; `git status --short | wc -l` → `0` (판정 파일 작성 직전 재확인)
- `go test -count=1 ./internal/auditreceipt/` → exit 0, `ok github.com/modu-ai/moai-adk/internal/auditreceipt 33.677s`
- `go test -count=1 -v -run 'TestAEC003a_|TestAEC003b_|TestE2' ./internal/hook/` → exit 0. 출력에 `--- PASS: TestAEC003b_E2FenceShapes_RewriteUnderE2MakesTheReceiptStale (20.84s)`가 있고 하위 15개 모두 PASS(`D1-four-backtick-holding-three`, `two-inline-code-lines`, `backtick-info-strings` 포함). 마지막 줄 `ok github.com/modu-ai/moai-adk/internal/hook 56.665s`
- D10 뮤턴트(스크래치 사본, `store.go:1889`를 `if false {`로 교체) → `--- FAIL: TestE2DocsUnchanged_E2BodyEditIsAlwaysAChange`, 패키지 `FAIL`. 원본 복원
- D11 뮤턴트(스크래치, `store.go:1953`에서 `len(line)-len(t) > 3 || ` 삭제) → `ok github.com/modu-ai/moai-adk/internal/auditreceipt 28.436s` (살아남음). 원본 복원
- 머지 보존 확인: merge-base `cb30e05b3e`. 다음 네 범위의 `git diff --name-only`를 구했다. 카드 쪽(`cb30e05b3e..c38c9bdf96`) 80개, main 쪽(`cb30e05b3e..9b632c1eb3`) 1064개, `c38c9bdf96..a1c76f1e99` 1064개, `9b632c1eb3..a1c76f1e99` 80개. 정렬 후 `diff`는 두 쌍 모두 exit 0이다. `--numstat` 비교도 exit 0이다. `-U0` 패치의 +/- 줄 비교에서는 10078줄 중 차이가 `internal/template/catalog.yaml`의 `hash:` 4줄뿐이었다(카드 쪽 `9365f5e4…`, `17caf205…`; 머지 쪽 `bc986f75…`, `00b91183…`)
- `make embed-manifest-check` → exit 0, `ok github.com/modu-ai/moai-adk/internal/template/embedemit 0.171s`
- `make agents-emit-check` → exit 0, `agents-emit-check: specs=12 copies=4 diff=0`
- 400caf47ba 뮤턴트(스크래치, `audit_delta_scope.go:186`의 diff 재료 삭제) → `--- FAIL: TestPerformCodexAudit_DeltaBaseRidesThePrompt`, `audit_delta_scope_test.go:282: codex leg prompt does not carry the delta diff material (canary_c_delta)`. 원본 복원 후 `ok`. 수정 전 단언(`git diff <sha>...HEAD` 문자열)은 이 뮤턴트를 통과시켰을 것이다. 프롬프트 산문에 `` `<sha>...HEAD` ``만 있고 `git diff `는 없다(`audit_delta_scope.go:180-182`, 읽어서 확인)
- `go test -count=1 -v -timeout 1500s -run 'TestAEC|TestClearRejection_|Delta|AuditMulti|CodexAudit|SchemaSnapshot|Convergence' ./internal/cli/` → exit 0, `ok github.com/modu-ai/moai-adk/internal/cli 525.836s`. 최상위 `--- PASS` 227개, `--- FAIL` 0개, `--- SKIP` 2개, `no tests to run` 0개. SKIP 2개는 `TestCodexAuditLaunchLiveContract`와 `TestCodexAuditLaunchLiveReadOnlyRoles`이며 사유는 `MOAI_CODEX_ROLE_LIVE is not 1`
- `go test -count=1 ./internal/auditverdict/ ./internal/closure/ ./internal/runtime/` → exit 0 (각 `ok`)
- `go test -count=1 -run 'TestAEC|TestFR_|Inconclusive|INCONCLUSIVE' ./internal/hook/ ./internal/homestate/` → exit 0 (`ok … hook 72.060s`, `ok … homestate 89.064s`)
- `go test -count=1 -v -run 'TestAEC005_' ./internal/template/` → exit 0, PASS 4개
- `golangci-lint run --timeout=5m` (auditreceipt, cli, hook, auditverdict, closure, homestate, runtime) → exit 0, `0 issues.`
- `gofmt -l` (같은 7개 패키지) → 출력 없음. `GOOS=windows GOARCH=amd64 go build ./...` → exit 0. `go build ./...` → exit 0
- `go test -count=1 -cover ./internal/auditreceipt/` → `coverage: 87.1% of statements`
- 문서와 코드 일치: `audit_error_class.go:19-23`의 상수 5개(`quota`, `timeout`, `auth`, `unavailable`, `other`)가 카탈로그 규칙 `:86` 및 스킬 `:160`, `:176`, `:364`의 집합과 같다. `grep -c SPEC-AUDIT-ENGINE-CONVERGENCE-001 CHANGELOG.md` → `1`
- §E.4: `progress.md:232`는 `sync_status: complete-pending-audit`, `:235`는 `sync_commit_sha: pending-backfill`, spec.md `status: in-progress`. 셋 다 PASS 전 기대 상태다. `:274`의 산문은 D15에 해당한다
- D13·D14 프로브: 스크래치 사본에 `zz_auditor_probe_test.go`를 넣고 `e2DocsUnchanged`를 직접 호출했다. `html-pre-hash-line-inside-E2 want=false got=true MISMATCH raw(old)=1 secs(old)=1`, `indented-E2-heading-edit … got=true MISMATCH raw(old)=0`, `tab-after-hashes-E2-edit … got=true MISMATCH raw(old)=0`. 대조군 `control-edit-inside-E2 got=false ok`, `fenced-hash-line-inside-E2 got=false ok`, `crlf-E2-edit got=false ok`. (`setext` 행은 내가 기대값을 잘못 적었다. CommonMark에서는 그 수정이 §E.2 밖이므로 기대값은 true여야 하고, 결과 false는 더 보수적인 쪽이다.)
- D14 코퍼스: `### §E.2 (Run-phase Evidence|Sync-phase)` 형태가 있는 progress.md 12개 중 정규 `## §E.2` 헤딩이 0개인 파일이 11개다(SPEC-COMPLETION-MARKER-RETIRE-001, SPEC-INTERNAL-TEST-003, SPEC-SEC-HARDEN-003 등)
- `moai verify audit-plan --project-root <toplevel>` → exit 0, `config_status: ok`, `model: codex`, codex gate `required`, `cross_model_active: true`, `enforced_required: ["codex"]`

### Baseline-attribution

- 측정 트리: 400caf47ba9ad1650148ec0b37aae23fb13fe304 (작업 트리 깨끗, 이번 실행).
- 뮤턴트와 프로브: 같은 커밋을 `git archive -o head.tar HEAD`로 스크래치에 풀어서 실행했다. 추적 트리는 건드리지 않았다.
- Go 테스트·빌드·lint: 이 트리의 소스를 로컬 `go` 툴체인으로 컴파일했다. `moai` 측정값은 `verify audit-plan` 하나만 인용한다. 그 바이너리는 설치본이고, 실행 중인 MCP 서버는 빌드 9b632c1eb3이다. 9b632c1eb3은 HEAD의 조상이다. 이 값은 감사 계획을 읽은 결과일 뿐 코드 판정의 근거로 쓰지 않았다.

### Gaps

- G1. **codex 필수 게이트를 실행하지 않았다.** 감사 계획은 codex `required`인데, 리더 지시와 운영자 선택 R503에 따라 `audit_multi`, `codex_audit` 등 백엔드 도구를 하나도 호출하지 않았다. 이 판정에는 codex 영수증이 없다(`receipts: none`). 관측 결과: 감사자 종료 시 SubagentStop 훅이 `AUDIT_RECEIPT_VIOLATION: sync-auditor reported PASS-WITH-DEBT but the audit could not be corroborated — no receipt cited.`로 거부했다. 감사자는 리더 지시(R503)에 따라 codex를 호출하지 않았고, 판정 내용도 바꾸지 않았다. 거절 표식 해제는 운영자 경로(`moai verify clear-rejection sync-auditor SPEC-AUDIT-ENGINE-CONVERGENCE-001 --evidence .moai/reports/t2017/sync-audit-3.md`)로 하거나, 운영자 승인을 받아 codex 감사를 한 번 돌리는 것으로 한다.
- G2. 워크트리 가드가 두 명령을 거부했다. 첫째는 `$S` 변수와 `;`가 섞인 `go test -run 'TestAEC003a_|…'` 복합 명령이었고, 같은 명령을 단순 형태로 다시 실행했다. 둘째는 변수를 쓰는 `git diff` 묶음이었고, 각 `git diff`를 단독 명령으로 나눠 실행한 뒤 비교는 스크래치 스크립트로 했다. 다른 명령으로 대체하지는 않았다.
- G3. 실행하지 않은 것: `internal/cli`·`internal/hook`·`internal/template` 전체 스위트, race detector, Windows 테스트 실행, `/verify`. 또 `internal/cli` 등 다른 패키지의 커버리지는 재지 않았다.
- G4. D4 spawn 거부 사유와 D6 영수증 시나리오는 실행으로 재현하지 않았다(코드 읽기만 함).
- G5. D13·D14는 순수 함수 `e2DocsUnchanged` 수준에서만 측정했다. Stop 경로 전 과정으로 재현하지는 않았다.
- G6. D1 뮤턴트(닫는 조건 강제)는 레인 기록만 읽었고 내가 다시 만들지 않았다. 대신 D10 뮤턴트를 직접 만들었다.
- G7. CHANGELOG 문구 전체를 코드와 줄 단위로 대조하지 않았다. 확인한 것은 error_class 집합, 항목 수, 카탈로그·스킬 언급뿐이다.

### Residual-risk

- 영수증 신선도 완화(B3)는 파서가 정의하는 "§E.2 섹션"에 기대고 있다. CommonMark와 다르게 읽는 경우 중 헤딩을 놓치는 방향은 D10 수리로 실패 처리된다. 하지만 섹션을 짧게 끊는 방향(D13, HTML 블록)과 정규 헤딩이 아예 없는 방향(D14)은 남아 있다. 둘 다 판정 대상 커밋에 그 형태가 이미 있어야 하고, 협력 프로세스 위협 모델(§E B1) 아래에 있다.
- 판정 파일에 적힌 `blocking_defect:` id를 감사자 자신이 쓴다는 점(§E B1)은 이번 수리로 바뀌지 않았다.
- 이 판정은 단일 모델 판정이다. G1 때문에 교차 모델의 독립 확인이 없다.

## 반복 이력

| 회차 | 감사 커밋 | 판정 | 주요 변화 |
|---|---|---|---|
| 1 | 1008b915b2 | FAIL (D1 BLOCKING) | fence boolean 토글이 §E.2를 가릴 수 있었다 |
| 2 | d9665c52c5 | FAIL 0.73 (D10 BLOCKING) | D1의 입력은 닫혔지만 근본 원인(빈 목록 비교)이 남았다 |
| 3 | 400caf47ba | PASS-WITH-DEBT 0.85 | D10 수리 c38c9bdf96, main 흡수 a1c76f1e99, 테스트 수리 400caf47ba. 차단 결함 없음. D13~D15 신규(optional) |

## Operational Notes (unverified)

- `assumption`(측정 지시): 영수증 가드가 codex required 트리에서 `receipts=none`인 PASS 계열 판정을 어떻게 처리하는지는 아직 측정하지 않았다. 확인하려면 이 감사자 종료 뒤 `ls .moai/state/audit-receipts/rejections/`를 실행해 sync-auditor 거절 표식이 생겼는지 본다. 감사 종료 뒤에야 볼 수 있는 상태다.
- `inferred`: D15는 `sync_commit_sha` 백필 커밋에서 같이 고치면 커밋 하나가 더 늘지 않는다. 근거 규칙은 `spec-frontmatter-schema.md` § SHA placeholder backfill exemption이다.
