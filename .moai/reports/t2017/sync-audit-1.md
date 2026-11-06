auditor-model: claude-opus-5-5

# Sync 감사 판정 — SPEC-AUDIT-ENGINE-CONVERGENCE-001 (카드 t2017), 1회차

verdict: FAIL
audited_sha: 1008b915b27d314c96670a9803345dfbcdaba4b8
blocking_count: 1
blocking_defect: D1
required_backend_fail: codex
convergence_overall: fail
required_backend: codex fail

## 요약

- 최종 판정: **FAIL** (가중 조화평균 점수 73/100, 하드 must-pass 방화벽에 걸림)
- 감사한 트리: 브랜치 `WT-audit-engine-repairs`, HEAD `1008b915b27d314c96670a9803345dfbcdaba4b8`, 작업 트리 깨끗함(`git status --short` 출력 없음). 범위는 로컬 `main`(`cb30e05b3e5f64774b8dabd87e80c24332f4551d`, merge-base와 같음) 위의 47개 커밋 전부다.
- 평가 프로파일: `.moai/config/evaluator-profiles/default.md`(SPEC에 `evaluator_profile` 없음, `harness.default_profile: "default"`), flat 가중 모드.
- FAIL 근거는 두 갈래다.
  1. 교차 모델 필수 레그(codex, gate `required`)가 **fail을 보고했다**(P1 4건, P2 2건). 보고된 fail이 있으면 전체 판정은 fail로 남는다.
  2. 독립 검토에서 BLOCKING 결함 D1을 직접 재현했다. 중첩 Markdown fence 하나로 `§E.2` 비교를 우회해, 영수증이 신선한 상태로 남는다. REQ-AEC-010 / AC-AEC-003b 위반이다.
- 이 판정은 오케스트레이터의 `moai verify audit-plan --result-file` 검사를 통과해야 확정된다(아래 audit_multi 다이제스트 참고).

## 차원별 점수 (default 프로파일)

| 차원 | 점수 | 판정 | 근거 (원문 인용은 아래 Evidence) |
|---|---|---|---|
| Functionality (40%, must-pass) | 70/100 | FAIL | 16개 AC 하위 기준의 동작 테스트는 전부 통과했다(cli 212 PASS/0 FAIL, hook 4, homestate 40, template 38, 패키지 4개 ok). 그러나 REQ-AEC-010 "`§E.2` 아래 한 줄만 바뀌어도 stale"이 유효한 Markdown 입력에서 깨진다(D1, 재현함). must-pass는 부분 점수를 주지 않는다. |
| Security (25%, must-pass) | 62/100 | FAIL | D1은 영수증 신선도 게이트의 무결성 우회다(High). 그 밖에 SHOULD-FIX/MINOR 급: D3 저장한 verdict 다이제스트를 어디서도 검증하지 않음, D5 `.moai/reports` 자체가 symlink면 증거 경로 containment가 트리 밖을 허용함. 승인 동사의 거절 순서, 거절 시 무쓰기, MCP 경로 부재, 유일 호출자는 확인했다. |
| Craft (20%) | 82/100 | PASS | 측정한 패키지 커버리지: auditreceipt 86.8%, auditverdict 95.1%, closure 88.6%, runtime 85.7%. golangci-lint `0 issues.`, gofmt 출력 없음. 뮤턴트 2개를 표본으로 만들었고 둘 다 테스트가 잡았다. cli/hook/homestate 커버리지는 측정하지 못했다(Gap). homestate 80.7%는 run 단계의 자체 보고 수치다. |
| Consistency (15%) | 76/100 | PASS (주의) | 컨벤션·스킬·카탈로그와 각 미러는 바이트 일치한다(`cmp` exit 0 ×3). 감사자 정의 6개 파일 모두 잘못된 문장 0건이고, `blocking_defect`와 4-라벨 `AUDIT-VERDICT` 템플릿을 1건씩 갖는다. 남은 불일치: D7 codex_audit 도구 설명(`internal/cli/mcp_server.go:566`)이 아직 "verdict 'fail'"이라고 적혀 있음(기존 문구, REQ-AEC-014 대상 파일은 아님), D8 Go 주석 두 곳(§E.4 Gap 10·11), D4 문서 문장 모호성, D9 EV-01 green 셀이 문자 그대로는 미충족. |

조화평균(가중치 미적용, 정보용): 4 / (1/70 + 1/62 + 1/82 + 1/76) ≈ 71.9. 판정은 must-pass 방화벽이 결정한다. Functionality(70, 기준: 모든 AC 충족)와 Security(62, 기준: High 없음)가 각자 기준에 미달하므로 다른 점수와 무관하게 FAIL이다.

## 결함 목록 (구조화)

| id | 심각도 | 분류 | 위치 | 한 줄 요약 | 필요한 수정 |
|---|---|---|---|---|---|
| D1 | BLOCKING (High) | blocking | `internal/auditreceipt/store.go:1893-1914` (`e2Sections`, 토글은 :1906) | fence 추적이 단순 boolean 토글이다. 판정 시점의 progress.md에 `````` ```` ```````(4 backtick) fence가 있고 그 안에 ```` ``` ```` 한 줄이 들어 있으면, 이후 내용이 계속 "fenced"로 읽힌다. 그래서 실제 `## §E.2` 헤딩이 구 버전·신 버전 모두에서 사라지고, 두 버전 비교가 빈 목록 대 빈 목록이라 같다고 판정된다. 그 결과 `§E.2` 본문을 고쳐 써도 영수증이 신선하게 남는다(REQ-AEC-010, AC-AEC-003b). | fence를 CommonMark식으로 추적한다(여는 문자, 길이 ≥3, 같은 문자이면서 길이가 같거나 긴 줄로만 닫힘). 또는 어느 한쪽 버전이라도 `## §E.2` 헤딩이 원문 줄에 있는데 섹션이 0개로 계산되면 stale로 판정한다. 4-backtick 안의 3-backtick 줄 fixture를 AC-AEC-003b 테스트에 추가한다. |
| D2 | SHOULD-FIX | optional | `internal/cli/mcp_convergence.go:1012-1016`, `:1067-1074` | 이어받은(carried) claude 레그가 이전 호출의 `gate`를 그대로 가진다. 재스탬프 루프가 claude를 건너뛰므로, 이전 호출이 advisory로 받은 claude FAIL이 현재 호출에서 required여도 converge가 이를 advisory로 취급한다(codex P1-1). 호출자는 원래 `gates` 인자로 gate를 낮출 수 있으므로 권한 상승은 아니다. 다만 결과가 현재 호출의 gate 계획과 맞지 않는다. | 이어받은 verdict에도 현재 호출의 gate를 다시 매긴다. 또는 persisted 결과의 gate 계획이 현재와 다르면 carry를 거부한다. 회귀 테스트를 추가한다. |
| D3 | SHOULD-FIX | optional | `internal/cli/audit_delta_chain.go:157-206`, `internal/closure/records.go:71` | 기록된 `verdict_digests`를 기록만 하고 다시 읽지 않는다(`grep VerdictDigests` 결과 비테스트 사용처는 쓰기 1곳뿐). 이전 링크가 의존한 FAIL 파일에서 id와 count를 지우면 closure 집합이 줄어든 채로 delta가 허용된다(codex P1-3). spec §E B1이 "감지 가능"이라는 잔여 위험으로 받아들인 범위이므로 SPEC 위반으로 판정하지는 않는다. 감사자 의견 불일치로 기록한다. | 이전 체인 레코드의 `verdict_digests`와 현재 파일의 SHA-256을 비교하고, 파일이 사라졌거나 달라졌으면 full scope로 보낸다. |
| D4 | MINOR | optional | `internal/hook/audit_receipt_guard.go:276-306`, `.moai/docs/audit-artifact-convention.md:193-194` | 기존 거절 표식이 없을 때 첫 INCONCLUSIVE 종료 후 phase-entry spawn이 허용된다(codex P1-4). FAIL 뒤의 hook 동작과 같아서 "exactly as after FAIL"에는 부합한다고 판단한다. 다만 문서의 "stays blocked … reason carries the error class"는 hook이 차단한다고 읽힐 수 있고, spawn 거부 사유에는 class가 들어가지 않는다. | 문서에 "차단은 verdict 파일 admission과 카드 리더가 맡고, hook은 미해결 거절을 지우지 않을 뿐"이라고 명시한다. 또는 SPEC 해석을 리더가 확정한다. |
| D5 | MINOR | optional | `internal/cli/verify_clear_rejection.go:195-205` | `.moai/reports` 자체를 symlink로 트리 밖에 연결하면 그 밖의 증거 파일이 containment를 통과한다(codex P2-5, 코드 읽기로 확인). spec §E B4가 "보안 경계 아님"이라고 명시했고, 경로 문자열상으로는 `.moai/reports/` 아래다. | 해석된 reports 루트가 `Canonical(tree)` 아래에 있는지 확인한다. 거절 시 저장소가 변하지 않음을 테스트한다. |
| D6 | SHOULD-FIX | optional | `internal/cli/mcp_convergence.go:1180-1186` | codex를 이어받고 다른 required 레그(예: glm)만 재실행해 overall이 pass가 되어도, 결과의 `audit_receipt`는 원래 실행의 영수증이다. 그 영수증에는 `gate_unmet=glm`이 남아 있어 `CheckCitedReceipts`가 `CauseReceiptAuditNotRun`으로 거절한다(codex P2-6, 코드 읽기로 확인). fail-closed라 안전 문제는 아니지만, B2의 목적(미실행 레그만 재실행)을 이 구성에서 무력화한다. 이 저장소처럼 codex만 required인 구성에서는 해당되지 않는다. | 재실행 결과를 증명할 영수증 의미론을 정한다. 예: carried codex 증거와 새 레그 결과를 묶는 영수증. 설계 판단이 필요하므로 리더/운영자 결정 사항이다. |
| D7 | MINOR | optional | `internal/cli/mcp_server.go:566`; `internal/cli/audit_delta_scope.go:27` | codex_audit 도구 설명이 "a review that produced no verdict returns verdict 'fail' with a non-empty 'gate_unmet'"이라고 적혀 있어 `applyGateUnmet`(:2446-2475, inconclusive 유지)과 맞지 않는다. codex_audit의 `delta_base` 설명도 "the audit_multi codex leg is told in the prompt only"라고 남아 있다. 둘 다 이 카드가 바꾸지 않은 기존 문구다. | 두 설명을 코드에 맞게 고친다. `audit-tools-schema-pre.json` golden도 함께 갱신한다. |
| D8 | MINOR | optional | `internal/cli/mcp_convergence.go:134-135`, `:1352-1353` | 주석이 코드와 맞지 않는다(`OverallVerdict ∈ {pass, fail}`; "The verdict now fails"). §E.4 Gap 10·11이 정직하게 기록해 두었다. | 주석을 고친다. |
| D9 | MINOR | optional | `acceptance.md` EV-01, `internal/cli/audit_delta_scope.go:106` | EV-01 green 셀("exit 1 with empty stdout")이 문자 그대로는 충족되지 않는다. `grep -n "prose-only scope"`가 아직 :106을 출력하고 exit 0이다(codex_audit 직접 경로, §F 범위 밖). progress.md는 이 편차를 조정·기록하지 않았다. | progress에 편차를 기록하거나, plan 측에서 EV-01 green 셀을 audit_multi 경로로 좁힌다. |

## 판정 근거 서술 (5-section)

### Claim

1. 빌드(네이티브·Windows), `agents-emit-check`, SPEC lint(strict)는 HEAD에서 통과한다.
2. 범위 테스트는 전부 통과하고, 표본 뮤턴트 2개를 잡는다.
3. 그러나 REQ-AEC-010을 위반하는 입력이 존재한다(D1).
4. 교차 모델 필수 레그는 fail을 보고했다.

따라서 판정은 FAIL이다.

### Evidence (명령과 원문 출력, 트리 `1008b915b2`)

```
$ git rev-parse HEAD
1008b915b27d314c96670a9803345dfbcdaba4b8
$ git status --short
(출력 없음)
$ git rev-parse main ; git merge-base main HEAD
cb30e05b3e5f64774b8dabd87e80c24332f4551d
cb30e05b3e5f64774b8dabd87e80c24332f4551d

$ go build ./...                               → build exit=0
$ GOOS=windows GOARCH=amd64 go build ./...     → win exit=0
$ make agents-emit-check
agents-emit-check: specs=12 copies=4 diff=0    → emit exit=0
$ <트리 빌드 바이너리> spec lint .moai/specs/SPEC-AUDIT-ENGINE-CONVERGENCE-001 --strict
✓ No findings — all SPEC documents are valid   → tree-lint exit=0
  (설치된 moai로도 같은 결과, exit=0)

$ moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2017
"config_status": "ok", "model": "codex", claude gate off / codex gate required / glm gate off,
"cross_model_active": true, "cross_model_required": true, "enforced_required": ["codex"]   exit=0

$ go test -count=1 -timeout 590s -run '^(TestAEC|TestClearRejection_|TestRunMultiAudit_|TestAuditMulti_|TestSelfReview_|TestCodexAudit_|TestApplyGateUnmet|TestEnforceRequiredGateUnmet)' -v ./internal/cli/
--- PASS: TestAEC004c_StoreClearFunctionHasExactlyOneNonTestCaller_TheVerb (0.44s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	243.307s
  최상위 --- PASS 212 / --- FAIL 0 / --- SKIP 0
  TestAEC 계열: 001a 3, 001b 7, 001c 2, 001d 5, 001e 1, 001f 9, 002a 5, 002c 12, 004a 3, 004b 3, 004c 3, 005 2
$ go test -count=1 -timeout 590s ./internal/auditreceipt/ ./internal/auditverdict/ ./internal/closure/ ./internal/runtime/
ok  	github.com/modu-ai/moai-adk/internal/auditreceipt	13.057s
ok  	github.com/modu-ai/moai-adk/internal/auditverdict	0.277s
ok  	github.com/modu-ai/moai-adk/internal/closure	14.416s
ok  	github.com/modu-ai/moai-adk/internal/runtime	5.263s
$ go test -count=1 -timeout 590s -run 'TestFR_' -v ./internal/homestate/
ok  	github.com/modu-ai/moai-adk/internal/homestate	31.464s   (PASS 40, FAIL 0)
$ go test -count=1 -timeout 590s -run '^(TestAEC003a_|TestAEC003b_)' -v ./internal/hook/
ok  	github.com/modu-ai/moai-adk/internal/hook	16.754s       (PASS 4, FAIL 0)
$ go test -count=1 -timeout 590s -run 'Catalog|Mirror|AuditPlanDocSurface|ClaudeAuditSurface|Neutral|SPECEditedRegionsSynced' -v ./internal/template/
ok  	github.com/modu-ai/moai-adk/internal/template	1.315s    (PASS 38, FAIL 0)
  "no tests to run" 0건(hook/homestate/template) — 선택된 테스트 집합이 비어 있지 않음

$ go test -count=1 -timeout 300s -cover ./internal/auditreceipt/ ./internal/auditverdict/ ./internal/closure/ ./internal/runtime/
ok  	.../internal/auditreceipt	13.379s	coverage: 86.8% of statements
ok  	.../internal/auditverdict	0.327s	coverage: 95.1% of statements
ok  	.../internal/closure	13.560s	coverage: 88.6% of statements
ok  	.../internal/runtime	4.447s	coverage: 85.7% of statements
$ golangci-lint run --timeout=8m ./internal/{cli,auditreceipt,auditverdict,hook,homestate,runtime,closure}/...
0 issues.                                       lint exit=0
$ gofmt -l <같은 7개 패키지>                     (출력 없음) exit=0
```

뮤턴트 표본(`git archive HEAD`로 만든 스크래치 사본에서만 수정, 감사 트리는 건드리지 않음):

```
뮤턴트 A — store.go e2Unchanged: `return true` (cited progress.md 전체를 면제)
$ go test -count=1 -run '^TestAEC003b_' ./internal/hook/
--- FAIL: TestAEC003b_AnyOtherChange_MakesTheReceiptStale (8.38s)
FAIL	github.com/modu-ai/moai-adk/internal/hook	10.140s     → 잡힘
뮤턴트 B — audit_delta_chain.go: 이후 재명명을 무시하고 id 단위로 상속
$ go test -count=1 -run '^TestAEC001d_ClosureConfirmation$' -v ./internal/cli/
    --- FAIL: TestAEC001d_ClosureConfirmation/re-report:_R1's_FAIL_names_B1_again_and_no_commit_after_R1's_head_closes_it (1.78s)
    (나머지 하위 테스트 12개 PASS, 대조군 'control: re-report with a closing commit ...' PASS)
FAIL	github.com/modu-ai/moai-adk/internal/cli	25.527s       → 잡힘
```

D1 재현(`e2Sections`를 원문 그대로 복사한 독립 프로그램, 스크래치):

```
입력: "## §E.1 Plan\n````text\nA fence opens with a line like:\n```\n````\n" 다음에
      "## §E.2 Run-phase Evidence\nAC-1 PASS"(구) / "...AC-1 FAIL (rewritten)"(신)
$ go -C <scratch>/fence run .
old sections: 1 new sections: 1
e2Unchanged-equivalent (should be false): false
odd inner fence: old sections: 0 new sections: 0 equal (should be false): true
control without nested fence (should be false): false
exit=0
```

`e2Unchanged`는 `slices.Equal(e2Sections(oldDoc), e2Sections(newDoc))`를 그대로 반환한다(store.go:1880). 따라서 이 입력에서 `evidenceOnlyRange`는 progress.md 항목을 허용한다. 실제 git 범위를 통한 end-to-end 재현은 내가 하지 않았다. codex 레그는 overlay 재현으로 `evidenceOnlyRange`가 true를 반환했다고 보고했다.

AC-AEC-005/006 문서 검사:

```
$ grep -c -E 'returns .verdict: fail. with a non-empty|makes the overall verdict .fail.|the overall verdict is .fail.' <6개 파일>
.claude/agents/moai/plan-auditor.md:0  .claude/agents/moai/sync-auditor.md:0
.codex/agents/moai/plan-auditor.toml:0 .codex/agents/moai/sync-auditor.toml:0
internal/template/agentemit/source/sync-auditor.md:0  internal/template/agentemit/source/plan-auditor.md:0
$ grep -c 'blocking_defect' <6개 파일>          → 각 1
$ grep -c 'AUDIT-VERDICT: <PASS|PASS-WITH-DEBT|FAIL|INCONCLUSIVE>' <6개 파일>   → 각 1
$ cmp .moai/docs/audit-artifact-convention.md internal/template/templates/.moai/docs/audit-artifact-convention.md   → conv cmp=0
$ cmp <skill> <mirror> → skill cmp=0 ; cmp <catalogue> <mirror> → catalogue cmp=0
$ grep -n 'may only remove' .moai/docs/audit-artifact-convention.md
210:may only remove the defective path, tighten an existing check, or add tests. A
$ grep -n 'verdict: <PASS' .moai/docs/audit-artifact-convention.md
87:  verdict: <PASS|PASS-WITH-DEBT|FAIL|INCONCLUSIVE>
```

RED-now 원장 재실행(HEAD에서 green 관찰):

```
EV-01 grep -n "prose-only scope" internal/cli/audit_delta_scope.go
      106:	return deltaBaseRef + " " + sha + " (prose-only scope: codex adversarial mode cannot be bounded mechanically)"   exit=0  (D9)
EV-02 grep -l "audited_sha" … → internal/cli/mcp_audit_multi.go  exit=0
EV-05 grep 'cfg.CardID != "" && cfg.DeltaBase == ""' → exit=1
EV-06 grep "error_class" mcp_convergence.go → 118: ErrorClass string `json:"error_class,omitempty"`
EV-13 store.go:178  `^AUDIT-VERDICT: (PASS|PASS-WITH-DEBT|FAIL|INCONCLUSIVE) …`
EV-14 card_evidence_readers.go:23 `^verdict: (PASS|PASS-WITH-DEBT|FAIL|INCONCLUSIVE)$`
EV-15 grep "return len(files), nil" audit_ceiling.go → exit=1
EV-18 grep 'cfg.DeltaBase == "" || result.OverallVerdict == overallVerdictFail' → exit=1
```

보안 측 읽기 검증:

- `verify_clear_rejection.go:99-185`의 거절 순서는 마커 → TTY → 표식 존재 → 증거 경로 → 증거가 SPEC을 명명하는지 → 입력한 확인 문구 → git 신원 → HEAD 순이다. 이 순서는 REQ-AEC-012와 일치하고, `ClearRejectionOnApproval` 전에는 아무것도 쓰지 않는다.
- `store.go:1587`의 `ownTree` 검사는 다른 트리의 표식이면 `ErrMarkNotFound`로 거절한다.
- 비테스트 호출자를 훑은 결과: `internal/cli/verify_clear_rejection.go:177` 한 곳뿐이다.
- 증거 파일 symlink는 `EvalSymlinks` 후 `Rel` 검사로 막힌다. `..`도 정규화된다. 예외는 D5다.
- `evidenceOnlyRange`의 나머지 부분: `--no-renames`로 rename을 경로별로 판정하고, A/M/D 외의 상태(T 포함)는 거절하며, 경로 정규식에 앵커가 걸려 있어 접두 트릭을 막는다. 미커밋 상태 다이제스트도 비교한다. 이 부분들은 의도대로 동작한다고 읽혔다.

### Baseline-attribution

- 모든 명령은 이번 실행에서 감사 트리 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2017` HEAD `1008b915b27d314c96670a9803345dfbcdaba4b8`(깨끗한 트리)를 대상으로 실행했다.
- 도구 출처(§2.2):
  - 설치된 `moai`는 `v3.2.0-rc.39 v3.2.0-1041-gdad3603de1`이다. `git merge-base --is-ancestor dad3603de1 HEAD` exit=1이므로 HEAD의 엄격한 조상이 아니다(별도의 rc.39 후보 라인).
  - 그래서 SPEC lint는 이 트리에서 빌드한 바이너리(스크래치 `moai-tree`)로 경로를 지정해 다시 실행했다.
  - `audit-plan`과 `factory renew` / `slot`은 설치된 바이너리로 실행했다.
  - 교차 모델 레그를 실행한 MCP 서버는 `build_commit: b1806498fe`(rc.38)다. 이 서버는 이 카드의 엔진 코드를 담고 있지 않지만, 여기서는 검토 주체일 뿐 검토 대상이 아니다.

### 교차 모델 레그 (audit_multi, 정확히 1회 호출)

호출 인자는 `project_root=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2017`, `target=baseBranch`, `session_id=t2017-sync-audit-1`이고 `delta_base`, `card_id`, `gates`는 넘기지 않았다. 결과 다이제스트(오케스트레이터 검사용 멤버만):

- `overall_verdict`: `fail`
- `gate_unmet`: (비어 있음, 필드 없음)
- `plan_source`: `config`
- `per_backend_verdicts`: `[{backend: codex, gate: required, verdict: fail}]`
- `audit_receipt`: `rcpt-d49281ed03fe058170314137`
- `state_notice`: (없음)
- `build_commit`: `b1806498fe`
- `residual_risk_note`: `required-backend FAIL: codex`

codex가 보고한 결함과 이 감사의 처리:

| codex | 이 감사의 판단 |
|---|---|
| P1 carried claude gate | D2로 확인(코드 읽기). SHOULD-FIX |
| P1 fence 우회 | D1로 확인(독립 재현). BLOCKING |
| P1 verdict 해시 미검증 | D3로 확인(코드 읽기). SPEC 잔여 위험 범위라서 SHOULD-FIX. **의견 불일치** |
| P1 INCONCLUSIVE 후 spawn 허용 | D4. "as after FAIL"에 부합한다고 판단해 MINOR. **의견 불일치** |
| P2 reports symlink | D5로 확인(코드 읽기). MINOR |
| P2 재실행 후 이전 영수증 | D6로 확인(코드 읽기). SHOULD-FIX |

보고된 fail은 전체 fail을 유지한다. 이 판정은 오케스트레이터가 위 다이제스트를 `<toplevel>/.moai/state/audit-plan-result.json`에 새로 쓰고 `moai verify audit-plan --project-root <toplevel> --result-file <path>`를 실행해 `convergence_check.ok`가 확인되기 전까지는 최종이 아니다.

### 판단 항목 (리더가 지정한 사항)

- **§E.4 Gap 10-13**:
  - Gap 10·11은 실제 주석 불일치다(D8, MINOR). 정직하게 기록되어 있다.
  - Gap 12의 두 예외 문장은 코드(`mcp_convergence.go:1346-1349`, `mcp_codex_audit.go:2456-2461`)와 일치함을 확인했다.
  - Gap 13("How to invoke" 문장을 요청 범위를 넘어 고친 것)은 코드에 맞는 정정이라 수용한다. 결함이 아니다.
- **레인이 정한 이름**: `blocking_closed`, `rerun_legs`, `carried_legs`, `audited_head/audited_digest/audited_scope`, `uncommitted_digest`, `clear-rejection`, `approvals.jsonl`, Latest-verdict 읽기 방식. 모두 카탈로그·스킬·컨벤션에 문서화되어 있고 코드 상수와 일치한다(`rerunLegsArg`, `blockingClosedArg`, `ErrorClass` json 태그 확인). 결함 아님.
- **`spec.md status: in-progress`, `sync_commit_sha: pending-backfill`**: 감사 전 상태로서 정직하다. `sync_status: complete-pending-audit`과 일관되고, placeholder는 인정된 `pending-backfill` 형태다. 다만 스키마 원칙(`completed` 전이는 sync 커밋에 실린다)과 달리 전이를 감사 뒤로 미룬 것은 의도적 편차다. FAIL 판정이므로 지금은 in-progress가 맞다.
- **Binding run conditions**: progress.md에 해당 헤딩이 없다(`grep -n -i 'binding run conditions'` exit=1). 처리할 항목이 없다.

### Gaps (관찰하지 않은 것)

1. `internal/cli` 전체 스위트와 `internal/hook` 전체 스위트는 지시에 따라 실행하지 않았다. Windows 테스트는 실행하지 않았고 빌드만 했다. CI 결과는 없다(푸시 없음).
2. cli/hook/homestate/template 패키지 커버리지는 측정하지 않았다. homestate 80.7%는 run 단계의 자체 보고이며 이번 실행에서 관찰하지 않았다.
3. D2·D3·D5·D6은 코드 읽기로 확인했고 직접 실행하지는 않았다. codex 레그는 overlay 재현을 했다고 보고했다. D1은 함수 원문 복사본으로 재현했고, git 범위를 거치는 end-to-end 재현은 하지 않았다.
4. 뮤턴트는 표본 2개뿐이다. AC별 뮤턴트 행 전부를 확인하지는 않았다.
5. docs-site와 README(§E.4 Gap 4)는 읽지 않았다.
6. 거절된 도구 호출(§3.1): 여러 Bash 호출이 worktree-isolation guard에 거절되었다. 사유는 복합 명령, 런타임 변수, `git grep` 인자 검증 불가였다. 대체 방법으로 명령을 단순 호출로 쪼개고, `git grep` 대신 깨끗한 작업 트리에 `grep -c`를 썼다(HEAD와 동일함을 `git status --short`로 확인). 측정 대상과 결과는 같다. 전용 Grep 도구는 이 세션에 없었다.
7. 실제 `.moai/state/audit-receipts`는 열람하지 않았고, `moai verify clear-rejection`은 실행하지 않았다(지시대로).

### Residual-risk (관찰했음에도 남는 위험)

- D1을 고쳐도 `§E.2` 경계 판정은 텍스트 휴리스틱이다. 다른 Markdown 구성(HTML 블록, 들여쓴 코드 등)에서 비슷한 어긋남이 남을 수 있다.
- 체인 게이트의 closure 집합은 감사자가 쓰고 gitignore된 verdict 파일에서 온다(D3). 협조적 프로세스 위협 모델 밖에서는 위조나 편집이 가능하다.
- 손으로 관리하는 로컬 감사자 정의 4개는 이후 소스와 어긋날 수 있다(spec §E B5 잔여 위험). 이 감사 시점에는 일치한다.
- 교차 모델 레그를 실행한 서버(rc.38)의 도구 설명에는 아직 옛 "fail" 문구가 있다(D7). 이를 읽는 에이전트가 오해할 수 있다.
- 중간 산출물(스크래치의 테스트 로그, 뮤턴트 사본)은 머신 로컬이고 이 파일에 인용한 줄 말고는 보존되지 않는다.

## 반복 이력

- 1회차(이번): FAIL. BLOCKING D1 하나와 codex 필수 레그의 보고된 fail이 근거다. 확인용 재감사는 D1과 D2-D6 처분의 델타로 범위를 잡을 수 있다(블로킹 id: D1).
