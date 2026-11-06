auditor-model: claude-opus-5-5
verdict: PASS-WITH-DEBT
audited_sha: 479c826792a753a99a1f742cba52301ba6679f44
overall_score: 0.88
functionality_score: 92
security_score: 90
craft_score: 84
consistency_score: 82
must_pass_failed: 0
blocking_count: 0
blocking_defect: none
required_backend: codex
required_backend_verdict: pass
receipts: rcpt-c28b5a8270c83157c7012186
delta_base: a50cf15d64d5a80ae259cbaf54ae2adbbfcdc01f

# Sync 감사 판정 — SPEC-AUDIT-ENGINE-CONVERGENCE-001 (카드 t2017), 7회차 (D20 확인 감사)

## 요약

판정은 PASS-WITH-DEBT, 가중 점수 88.4(0.88)다. 차단 결함은 없다. 부채는 optional 3건(D21·D22·D23)이고 이번 카드에서 일부러 손대지 않았다.

- **D20은 닫혔다.** `CHANGELOG.md:17`의 SPEC-AUDIT-ENGINE-CONVERGENCE-001 항목에서 제목 절 "a receipt survives evidence-only commits"가 빠졌다. 항목 (3)은 "**Receipt freshness unchanged**: … excluded from this SPEC by operator decision; after an audit, any commit that moves HEAD still leaves the receipt stale (SPEC-MCP-NEUTRALITY-001 REQ-MN-020)"로 바뀌었다. 개수는 "14 live requirements of REQ-AEC-001..016 with REQ-AEC-009/010 retired, 14 live acceptance sub-criteria"다. 이 감사가 직접 센 값(REQ 14, AC 14, 은퇴 AC 2)과 맞는다.
- B3를 출시 기능으로 적은 문장은 사용자 대상 기록 어디에도 없다. 남은 언급은 SPEC progress.md의 실행 이력(136행)과 §E.3 표(212행)뿐이다. 둘 다 D21로 이미 기록된 optional 부채다.
- 항목 (1), (2), (4), (5)는 인용한 파일과 대조했고 모두 맞다(아래 Evidence).
- codex(필수 게이트)는 `pass`, 지적 0건, 영수증 `rcpt-c28b5a8270c83157c7012186`이다. 운영자가 허락한 4번째이자 마지막 호출이다(d-20261106T163140Z-5ef9).
- 델타 a50cf15d64..HEAD에는 `.go` 파일과 `.moai/specs` 변경이 없다. 빌드, vet, auditreceipt 시험, strict lint 모두 기준선 그대로 통과한다.

## 차원별 점수 (default 프로파일, flat 가중)

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%, must-pass) | 92/100 | PASS | `go test ./internal/auditreceipt/... -count=1` → `ok … 27.497s`. `go build ./...` exit 0. CHANGELOG 항목이 실제 동작(HEAD 이동 = stale)과 일치한다. 6회차 감점 사유(D20)가 사라졌다. |
| Security (25%, must-pass) | 90/100 | PASS | 코드 무변경(`git diff --stat a50cf15d64..HEAD -- '*.go' '.moai/specs'` 출력 없음). 영수증 신선도는 6회차에 뮤턴트로 확인한 fail-closed 상태 그대로다. 이번 델타는 문서 한 줄이라 새 공격면이 없다. |
| Craft (20%) | 84/100 | PASS | `go vet ./internal/auditreceipt/... ./internal/hook/...` exit 0. 시험 실패 메시지 문구(D22)와 옛 영수증 로드 시험 부재(D23)는 그대로다. |
| Consistency (15%) | 82/100 | PASS | CHANGELOG, spec.md, acceptance.md, plan.md가 14/14로 서로 맞는다. strict lint 무결함. progress.md §E.3/§E.4가 아직 제외 전 값(M5 PASS 2 of 2, `live=16`)을 적는다(D21). |

가중 합: 92×0.40 + 90×0.25 + 84×0.20 + 82×0.15 = 36.8 + 22.5 + 16.8 + 12.3 = 88.4. must-pass 두 차원이 모두 기준을 넘는다. 필수 백엔드 codex는 pass이고 영수증이 있다. optional 부채 3건이 남아서 PASS가 아니라 PASS-WITH-DEBT로 적는다.

## 백엔드별 결과

- **codex** (gate `required`, `mcp__moai__codex_audit` 정확히 1회, mode `adversarial`, target `baseBranch`, `delta_base: a50cf15d64d5a80ae259cbaf54ae2adbbfcdc01f`, `project_root` = 이 워크트리): `verdict: pass`, `findings: []`, `audit_receipt: rcpt-c28b5a8270c83157c7012186`, `resolved_model: gpt-6.1-sol`, `requested_effort: high`, `build_commit: 04d8c27a7e`, `review_base: "delta_base a50cf15d64d5a80ae259cbaf54ae2adbbfcdc01f (prose-only scope: codex adversarial mode cannot be bounded mechanically)"`. 재호출하지 않았다.
- **claude**: 계획에서 gate `off`. 이 감사(Claude, 자체 검토)가 아래 증거를 따로 모았다.
- **glm**: gate `off`, 호출하지 않았다.
- `audit_multi`, `claude_audit`, `glm_audit`은 지시대로 호출하지 않았다. 그래서 convergence digest가 없다(Gap G1).

## 결함 목록

| id | 심각도 | 분류 | 위치 태그 | 위치 | 내용 | Required fix / 상태 |
|---|---|---|---|---|---|---|
| D20 | Medium | 닫힘 | in-diff | `CHANGELOG.md:17` | 479c826792가 제목 절, 항목 (3), 개수를 고쳤다. B3를 출시 기능으로 적는 문장이 없다. | 닫힘 |
| D21 | Low | optional (유지) | touched-file, 이번 델타 밖 | `.moai/specs/SPEC-AUDIT-ENGINE-CONVERGENCE-001/progress.md:212`, `:243` | §E.3 "M5 … PASS (2 of 2)"와 §E.4 `b12_self_test_b: "… live=16 … (non-zero, matches the CHANGELOG count)"`가 그대로다. CHANGELOG가 14로 바뀌었으므로 "matches the CHANGELOG count"는 이제 틀린 말이다. 그래도 내부 실행 기록이고 사용자 대상 문서가 아니다. SPEC 요구사항이 이 값을 요구하지도 않는다. 그래서 optional로 둔다. | 다음 sync 갱신 때 §E.3 M5 행에 retired 표시를 달고 §E.4를 14로 다시 잰다. |
| D22 | Low | optional (유지) | 이번 델타 밖 | `internal/hook/audit_receipt_range_test.go:134` | 실패 메시지가 아직 "was refused after evidence-only commits"다. 단언은 맞고 메시지만 실제와 다르다. | 재량 사항이다. |
| D23 | Low | optional (유지) | 이번 델타 밖 | `internal/hook/audit_receipt_range_test.go` | `uncommitted_digest`를 가진 옛 영수증 로드를 저장소 시험이 고정하지 않는다. 6회차가 스크래치 사본에서 로드 가능함을 확인했다. 이번 델타에서 코드가 바뀌지 않았으므로 그 결과는 유효하다(재측정 아님). | 재량 사항이다. |
| D24 | Low | optional (새로 발견) | in-diff | `CHANGELOG.md:17` | 개수 절이 "REQ-AEC-009/010 retired"는 적지만 은퇴한 AC 두 개(AC-AEC-003a/b)는 적지 않는다. 결과값 14는 맞다. | 원하면 "(AC-AEC-003a/b retired)"를 덧붙인다. 재량 사항이다. |
| D2–D19 | — | 6회차 기록 그대로 | — | — | 이번에 다시 재지 않았다 | 이전 기록 인용(이번 실행의 증거 아님) |

반복 수리 중지 규칙: 이번 회차에 blocking 결함이 없으므로 repair-born 판정 대상이 없다.

## 지시 항목별 결과

1. **개수 일치**: spec.md 살아 있는 REQ 14개, acceptance.md 살아 있는 AC 행 14개, 은퇴 AC 행 2개. CHANGELOG는 "14 live requirements … 14 live acceptance sub-criteria"다. 일치한다.
2. **항목 (3)**: "Receipt freshness unchanged … excluded … any commit that moves HEAD still leaves the receipt stale". B3를 출시 기능으로 적는 문장이 없다.
3. **항목 (1), (2), (4), (5)**: 인용한 파일이 모두 있고, 설명한 식별자와 동작이 코드에 있다(Evidence 참조). 위반은 없다.
4. **도구 측정**: strict lint exit 0, `go build ./...` exit 0, `go vet` exit 0, `go test ./internal/auditreceipt/... -count=1` ok.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

1. 감사 대상은 HEAD 479c826792a7이고, 트리는 시작과 판정 파일 작성 직전 모두 깨끗했으며 바뀌지 않았다.
2. a50cf15d64 이후 델타는 CHANGELOG.md 한 줄(+1/-1)과 6회차 판정 파일 커밋뿐이다. 코드와 SPEC은 바뀌지 않았다.
3. CHANGELOG의 SPEC-AUDIT-ENGINE-CONVERGENCE-001 항목은 B3를 출시 기능으로 적지 않고, 개수 14/14가 SPEC과 맞는다.
4. 다른 사용자 대상 기록(README, docs-site, `.claude`, `internal/template`, `.moai/docs`, `.codex`)에도 B3를 출시 기능으로 적는 문장이 없다.
5. 항목 (1), (2), (4), (5)는 코드와 맞는다.
6. 빌드, vet, 시험, strict lint가 통과한다.
7. codex 필수 게이트는 pass이고 영수증이 있다.

### Evidence (명령 → exit → 판정에 쓴 출력)

- `git rev-parse HEAD` → `479c826792a753a99a1f742cba52301ba6679f44`(시작, 판정 파일 작성 직전 모두). `git status --short` → 출력 없음(두 번 모두). `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2017`.
- `git diff --stat a50cf15d64..HEAD` → `.moai/reports/t2017/sync-audit-6.md | 162 +++`, `CHANGELOG.md | 2 +-`, `2 files changed, 163 insertions(+), 1 deletion(-)`.
- `git log --oneline a50cf15d64..HEAD` → `479c826792 … CHANGELOG no longer presents the excluded B3 as shipped (D20, card t2017)`, `e07d601adf … sync audit 6 verdict FAIL 0.84 …`.
- `git diff --stat a50cf15d64..HEAD -- '*.go' '.moai/specs'` → 출력 없음.
- `git diff a50cf15d64..HEAD -- CHANGELOG.md` → 17행 한 줄 교체. 지운 줄: "… and a receipt survives evidence-only commits (card t2017, Tier M; 16 requirements REQ-AEC-001..016, 16 acceptance sub-criteria …) … (3) **Receipt binding across evidence-only commits** … an audit receipt stays valid when the commits since the audited head change only audit evidence …". 새 줄: "… is no longer read as a FAIL (card t2017, Tier M; 14 live requirements of REQ-AEC-001..016 with REQ-AEC-009/010 retired, 14 live acceptance sub-criteria AC-AEC-001a..006 counted by the B12 counter, …) … (3) **Receipt freshness unchanged**: the evidence-only freshness relaxation (REQ-AEC-009/010) was excluded from this SPEC by operator decision; after an audit, any commit that moves HEAD still leaves the receipt stale (SPEC-MCP-NEUTRALITY-001 REQ-MN-020). (4) …". 항목 (1), (2), (4), (5) 본문은 글자 그대로 같다.
- `grep -cE '^\s*-\s+\*\*REQ-AEC-0[0-9]+\*\*' .moai/specs/SPEC-AUDIT-ENGINE-CONVERGENCE-001/spec.md` → `14`(REQ-AEC-009/010은 `**REQ-AEC-009 [RETIRED]**` 형태라 패턴에 걸리지 않는다. 101·102행에서 따로 확인).
- `grep -cE '^\| AC-AEC-[0-9]+[a-z]? \|' …/acceptance.md` → `14`. `grep -cE '^\| AC-AEC-[0-9]+[a-z]? \[RETIRED\] \|' …/acceptance.md` → `2`.
- `sed -n '17p' CHANGELOG.md | tr '.;' '\n\n' | grep -nE 'evidence-only|16 requirements|16 acceptance|survives|stays valid|Receipt binding|stays fresh|retired|excluded|14 live|stale'` → `4: 14 live requirements of REQ-AEC-001`, `6:016 with REQ-AEC-009/010 retired, 14 live acceptance sub-criteria AC-AEC-001a`, `27: (3) **Receipt freshness unchanged**: the evidence-only freshness relaxation (REQ-AEC-009/010) was excluded from this SPEC by operator decision`, `28: after an audit, any commit that moves HEAD still leaves the receipt stale (SPEC-MCP-NEUTRALITY-001 REQ-MN-020)`.
- `git grep -n -i -E 'survives evidence-only|receipt binding across evidence-only|receipt survives' HEAD -- . ':!.moai/reports'` → `progress.md:136:### M5 — B3 receipt binding across evidence-only commits`, `progress.md:212:| M5 | B3 receipt binding across evidence-only commits | AC-AEC-003a, 003b | PASS (2 of 2) | …`(둘 다 D21). CHANGELOG 0건.
- `git grep -n -i -E 'evidence-only' HEAD -- 'README*' 'docs-site' '.claude' 'internal/template' '.moai/docs' '.codex' 'CHANGELOG.md'` → `CHANGELOG.md:17` 한 줄뿐이고, 그 줄의 해당 문구는 제외를 알리는 항목 (3)이다.
- `git grep -n -E '16 requirements REQ-AEC|16 acceptance sub-criteria|REQ-AEC-001\.\.016' HEAD -- . ':!.moai/reports'` → `plan.md:18`("14 live (REQ-AEC-001..016 less REQ-AEC-009 [RETIRED] and REQ-AEC-010 [RETIRED])")과 CHANGELOG 17행("14 live requirements of REQ-AEC-001..016")만 나온다. 16/16을 현재 값으로 적는 줄은 없다.
- `grep -nE 'live=16|matches the CHANGELOG count' …/progress.md` → `243:b12_self_test_b: "moai-ac-counter on acceptance.md: live=16 excluded=0 ambiguous=0, stdout 16, exit 0 (non-zero, matches the CHANGELOG count)"`(D21).
- `sed -n '134p' internal/hook/audit_receipt_range_test.go` → `t.Errorf("the receipt judged at %s was refused after evidence-only commits (spec=%s): %v", …)`(D22).
- 항목 (1) 대조: `ls` → `internal/cli/audit_delta_default.go`, `audit_delta_chain.go`, `audit_delta_scope.go` 존재. `grep` → `audit_delta_default.go:25: deltaBaseFull = "full"`, `mcp_audit_multi.go:192: const blockingClosedArg = "blocking_closed"`, `audit_delta_chain.go:184: return refuse("the %s confirmation does not name every BLOCKING id of the closure set (missing: %s)", "blocking_closed", …)`.
- 항목 (2) 대조: `audit_error_class.go:19-23` → `errClassQuota = "quota"`, `errClassTimeout = "timeout"`, `errClassAuth = "auth"`, `errClassUnavailable = "unavailable"`, `errClassOther = "other"`. `audit_rerun_legs.go:26: const rerunLegsArg = "rerun_legs"`, `mcp_convergence.go:226: CarriedLegs []string \`json:"carried_legs,omitempty"\``. `audit_counter.go:90: return auditverdict.Parse(data).Label == "INCONCLUSIVE"`, `:100 func CountAuditRounds`, `audit_ceiling.go:1186: func CountPlanAuditRounds`. `INCONCLUSIVE`을 다루는 비시험 Go 파일: `internal/homestate/card_evidence_readers.go`, `internal/runtime/audit_counter.go`, `audit_ceiling.go`, `audit_gate.go`, `internal/hook/audit_receipt_guard.go` 등 12개.
- 항목 (4) 대조: `verify_clear_rejection.go:72-74` "The verb refuses, writing and removing nothing, unless: neither CLAUDECODE nor CLAUDE_CODE_SESSION_ID is set; standard input is an interactive terminal; the named mark exists for this tree; the evidence file is under .moai/reports/ and …", `:111 if !contractStdinIsTerminalFn()`, `:155 "the typed confirmation did not match %q"`, `:205 "the evidence file %q is outside .moai/reports/"`. `grep -rn 'clear-rejection\|clear_rejection\|ClearRejection' internal/cli/mcp_server.go internal/cli/mcp_*.go | grep -v _test` → 출력 없음(MCP 도구 없음).
- 항목 (5) 대조: `grep -c INCONCLUSIVE` → 소스 plan-auditor 6, 소스 sync-auditor 5, 로컬 `.claude` plan/sync 6/5, `.codex` plan/sync 6/5, 규약 문서 13. `internal/template/agentemit/source/sync-auditor.md:213` "… the result is `inconclusive` with a non-empty `gate_unmet`, yields no PASS, and the verdict file you export says `INCONCLUSIVE` … it returns `fail` only when the gate configuration cannot be read". `audit-artifact-convention.md:210-214` "may only remove the defective path, tighten an existing check, or add tests … a SPEC amendment with its own acceptance criterion on a separate card … work already specified by a SPEC amendment". `grep -rlE 'blocking_closed|rerun_legs|carried_legs'` → `.claude/rules/moai/core/moai-mcp-tools-catalogue.md`, `.claude/skills/moai-ref-cross-model-audit/SKILL.md`. `mcp_convergence.go:1325 func enforceRequiredGateUnmet`, `mcp_codex_audit.go:2455 func applyGateUnmet`.
- `go build -o <scratch>/moai-tree ./cmd/moai` → exit 0.
- `<scratch>/moai-tree spec lint --strict SPEC-AUDIT-ENGINE-CONVERGENCE-001` → exit 0, `✓ No findings — all SPEC documents are valid`(단독 실행 1회, 이어서 `; echo "exit=$?"`로 다시 실행해 `exit=0` 확인. 파이프는 쓰지 않았다).
- `go build ./...` → exit 0, 출력 없음.
- `go vet ./internal/auditreceipt/... ./internal/hook/...` → exit 0, 출력 없음.
- `go test ./internal/auditreceipt/... -count=1` → exit 0, `ok  github.com/modu-ai/moai-adk/internal/auditreceipt 27.497s`.
- `<scratch>/moai-tree verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2017` → `config_status: ok`, `model: codex`, claude `off`, codex `required`, glm `off`(모두 config.gates, explicit), `cross_model_active: true`, `enforced_required: ["codex"]`.
- `<scratch>/moai-tree decision read --scope card:t2017` → `id=d-20261106T163140Z-5ef9 scope=card:t2017 kind=ceiling-exception`, 본문 "One fourth codex_audit approved for t2017, scoped to the delta after sync-audit-6 (the one CHANGELOG line). No further calls."
- `mcp__moai__codex_audit`(1회) → `verdict: pass`, `findings: []`, `audit_receipt: rcpt-c28b5a8270c83157c7012186`, `build_commit: 04d8c27a7e`.
- `git check-ignore -v .moai/reports/t2017/sync-audit-7.md` → `.gitignore:235:.moai/reports/*`.

### Baseline-attribution

- 측정 트리: 479c826792a753a99a1f742cba52301ba6679f44(작업 트리 깨끗, 이번 실행).
- 비교 기준: a50cf15d64d5a80ae259cbaf54ae2adbbfcdc01f(6회차 감사 SHA, delta_base).
- Go 빌드, vet, 시험은 이 트리 소스를 로컬 `go` 툴체인과 기본 GOCACHE로 쟀다(디스크 여유 9.4Gi라 비공개 캐시를 만들지 않았고, 지울 것도 없다).
- `moai` 측정(spec lint, audit-plan, decision read)의 판정 빌드: 이 트리에서 `go build -o <scratch>/moai-tree ./cmd/moai`로 만든 바이너리를 경로로 직접 실행했다(§2.2 첫째 방식). 빌드 출처는 트리 HEAD 479c826792다. 설치본은 판정에 쓰지 않았다.
- codex는 MCP 서버 빌드 `04d8c27a7e`(v3.2.1-rc.2)로 실행됐다. 6회차 기록대로 04d8c27a7e는 이 트리 HEAD의 조상이 아니다(이번에 다시 재지 않음). codex 결과의 출처일 뿐이고, 코드와 문서 판정 근거는 위 측정이다.

### Gaps

- G1. 계획은 `cross_model_active: true`를 보고했지만, 지시에 따라 `audit_multi`를 호출하지 않았다. convergence digest(`overall_verdict`, `gate_unmet`, `plan_source`)와 orchestrator의 `--result-file` 확인이 없다. 단일 백엔드 codex 결과와 영수증이 필수 게이트의 근거다.
- G2. codex는 adversarial 모드라 범위가 기계적으로 묶이지 않았다(`prose-only scope`). 재호출하지 않았다(운영자 허용 4회 소진).
- G3. 거부된 명령이 1건 있다. `git diff a50cf15d64..HEAD -- CHANGELOG.md > <scratch>/a7-changelog.diff; echo exit=$?; wc -c …`를 워크트리 격리 가드가 "this command names git in a form too complex to verify that it stays inside the worktree. Refusing to run it"로 거부했다. 같은 `git diff`를 리다이렉트 없이 단독 명령으로 다시 실행했고 그 출력을 썼다. 대체한 측정은 없다.
- G4. 실행하지 않은 것: `internal/hook` 시험(이번에는 vet만), golangci-lint, 커버리지, race detector, Windows 빌드, `/verify`, 넓은 패키지 시험. 델타에 코드 변경이 없어서 6회차 측정을 기준선으로 본다(6회차 결과의 인용이고 이번 재실행이 아니다).
- G5. D2–D19와 D23의 로드 확인은 다시 재지 않았다(이전 회차 기록 인용).
- G6. 문서 검색은 저장소 추적 파일(`git grep`)에 한정했다. 배포된 docs-site 사이트나 저장소 밖 문서는 보지 않았다. 다른 말로 B3를 표현한 문장은 이번 패턴에 걸리지 않을 수 있다.
- G7. 항목 (1)·(2)·(4)·(5)는 식별자와 문구를 코드와 대조한 정적 확인이다. 각 동작의 시험은 이번에 다시 돌리지 않았다(auditreceipt 제외).

### Residual-risk

- progress.md §E.4가 "matches the CHANGELOG count"(16)라고 적고 있어서, CHANGELOG가 14로 바뀐 지금은 두 기록이 어긋난다(D21). 나중에 이 줄을 근거로 개수를 판단하면 혼선이 생길 수 있다.
- CHANGELOG 개수 절은 은퇴한 AC id를 적지 않는다(D24). 결과값은 맞다.
- 이 카드는 판정 파일을 `.gitignore` 대상인데도 커밋해 왔다(sync-audit-1, 3, 4, 5, 6이 추적됨). 이 감사는 7회차 파일을 커밋하지 않는다. 판단은 리더 몫이다.
- 판정 파일의 `blocking_defect:` id를 감사자 자신이 쓴다는 점은 바뀌지 않았다.
- 스크래치의 바이너리와 출력은 기계 로컬이다. 판정에 쓴 줄은 위 Evidence에 옮겨 적었다.

## 반복 이력

| 회차 | 감사 커밋 | 판정 | 주요 변화 |
|---|---|---|---|
| 1 | 1008b915b2 | FAIL (D1 BLOCKING) | fence boolean 토글이 §E.2를 가렸다 |
| 2 | d9665c52c5 | FAIL 0.73 (D10 BLOCKING) | D1은 닫혔지만 빈 목록 비교가 남았다 |
| 3 | 400caf47ba | PASS-WITH-DEBT 0.85, receipts none | Claude 단독이라 영수증 가드가 거부했다(codex required) |
| 4 | 7ea9b20555 | FAIL 0.73 (D16, D13 BLOCKING), codex fail `rcpt-919c8ec5ced34525ba7663b0` | D16 신규, D13 재분류 |
| 5 | 985853c683 | FAIL 0.75 (D19 BLOCKING, repair-born), codex fail `rcpt-7911463ec3cd51004d46c029` | 인용 경계 fail-open(D19), 가용성 회귀 D18 |
| 6 | a50cf15d64 | FAIL 0.84 (D20 BLOCKING), codex fail `rcpt-be71757522b8572b0da40751` | 운영자 결정으로 B3 제외, 코드는 M5 이전으로 복원. CHANGELOG에 B3 설명이 남음(D20) |
| 7 | 479c826792 | PASS-WITH-DEBT 0.88, codex pass `rcpt-c28b5a8270c83157c7012186` | D20 닫힘. optional D21–D24 남음 |
