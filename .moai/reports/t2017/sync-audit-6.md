auditor-model: claude-opus-5-5
verdict: FAIL
audited_sha: a50cf15d64d5a80ae259cbaf54ae2adbbfcdc01f
overall_score: 0.84
functionality_score: 86
security_score: 90
craft_score: 84
consistency_score: 70
must_pass_failed: 0
blocking_count: 1
blocking_defect: D20
required_backend: codex
required_backend_verdict: fail
receipts: rcpt-be71757522b8572b0da40751
delta_base: 985853c683565e6f765ad17608d4526760804333

# Sync 감사 판정 — SPEC-AUDIT-ENGINE-CONVERGENCE-001 (카드 t2017), 6회차 (B3 제외 델타 재감사)

## 요약

판정은 FAIL, 가중 점수 84.2(0.84)다. 차단 결함은 1건(D20)이다. 코드 쪽 제거는 완전하고 fail-closed다. 남은 문제는 문서 한 줄이다.

- **B3 코드 제거는 완전하다.** `internal/auditreceipt/store.go`와 `internal/hook/audit_receipt_guard.go`는 M5 이전 커밋 d15679fc78과 바이트 단위로 같다(`git diff --stat d15679fc78 HEAD` 출력에 두 파일이 없다). 지운 이름 여섯 개와 `UncommittedDigest`/`CheckCitedReceiptsForSpec`는 `.go` 파일 어디에도 없다.
- **신선도는 모든 커밋 형태에서 fail-closed다.** 코드, reports만, progress만(§E.1·§E.2·§E.3·§E.4), 빈 커밋, 커밋 뒤 미커밋 수정, HEAD 비후손 형태가 모두 `CauseReceiptTreeStateChanged`로 stale이다. HEAD가 그대로면 fresh이고, 같은 HEAD에서 미커밋 수정이 생기면 stale이다. "HEAD 이동 무시" 뮤턴트를 새 시험 6행이 모두 잡았다.
- **D20 (blocking, 새로 발견, codex가 독립적으로 같은 지적)**: `CHANGELOG.md:17`이 제외된 B3를 출시 기능으로 계속 설명한다. 제목 절 "a receipt survives evidence-only commits", 항목 (3) "Receipt binding across evidence-only commits … stays valid", 그리고 "16 requirements … 16 acceptance sub-criteria"가 남아 있다. SPEC은 14/14이고 REQ-MN-020은 완화하지 않았다고 적는다.
- codex(필수 게이트)는 `fail`이고 영수증은 `rcpt-be71757522b8572b0da40751`이다. 지적은 위 D20과 같은 1건(P2)이다. 필수 게이트가 fail이므로 PASS 계열 판정은 나올 수 없다.

## 차원별 점수 (default 프로파일, flat 가중)

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%, must-pass) | 86/100 | PASS | 지정 시험이 모두 통과했다(auditreceipt `ok 32.775s coverage: 86.2%`, hook `TestAEC` 최상위 3개, 하위 25개 PASS). B1·B2·B4·B5 쪽 시험 함수는 하나도 지워지지 않았다. 감점 사유는 사용자 대상 기록(CHANGELOG)이 실제 동작과 다르다는 점이다(D20). |
| Security (25%, must-pass) | 90/100 | PASS | 영수증 신선도는 REQ-MN-020 그대로 fail-closed다. 뮤턴트(HEAD 불일치 무시)를 시험 6행이 잡았다. 남은 fail-open 경로는 없다. D20은 fail-closed 방향의 문서 오기라서 보안 위험이 아니다. |
| Craft (20%) | 84/100 | PASS | golangci-lint `0 issues.`, go vet exit 0, 커버리지 86.2%다. 시험 도우미 실패 메시지 한 줄에 옛 문구가 남았다(D22). |
| Consistency (15%) | 70/100 | FAIL | SPEC 세 문서는 서로 맞는다(14/14, `[RETIRED]` 토큰 24곳 모두 붙어 있음, strict lint 무결함). 그러나 같은 카드가 쓴 CHANGELOG(D20)와 SPEC progress.md §E.3/§E.4(D21)가 제외 전 상태를 현재 사실처럼 적고 있다. |

가중 합: 86×0.40 + 90×0.25 + 84×0.20 + 70×0.15 = 84.2. must-pass 두 차원은 기준을 넘는다. 그래도 FAIL인 이유는 둘이다. 필수 백엔드 codex가 `fail`을 보고했고, 차단 결함 D20이 열려 있다.

## 백엔드별 결과

- **codex** (gate `required`, `mcp__moai__codex_audit` 1회, mode `adversarial`, target `baseBranch`, `delta_base: 985853c683565e6f765ad17608d4526760804333`, `project_root` = 이 워크트리, resolved_model `gpt-6.1-sol`, build_commit `04d8c27a7e`): verdict **`fail`**, receipt **`rcpt-be71757522b8572b0da40751`**, 지적 1건.
  - 반환된 `review_base`는 `delta_base 985853c683565e6f765ad17608d4526760804333 (prose-only scope: codex adversarial mode cannot be bounded mechanically)`이다.
  - [P2] `CHANGELOG.md:17` "B3 제거 후에도 CHANGELOG가 삭제된 영수증 재사용을 현재 기능으로 안내한다" → **D20. 동의한다. 이 감사도 grep으로 따로 찾았다.**
  - codex는 "집중 테스트는 임시 디렉터리 생성 권한 거부로 NOT_RUN"이라고 스스로 밝혔다. 이 감사의 Go 시험이 그 빈자리를 채운다.
  - 이번 호출이 운영자가 허락한 3번째이자 마지막 호출이다(d-20261106T074340Z-19c9). 재호출은 하지 않았다.
- **claude**: 계획에서 gate `off`. 이 감사(Claude, 자체 검토)가 아래 증거를 독립적으로 모았다.
- **glm**: gate `off`, 호출하지 않았다.
- `moai verify audit-plan`은 `cross_model_active: true`를 보고했다. 지시에 따라 `audit_multi`는 호출하지 않았다. 그래서 convergence digest가 없다(Gap G1).

## 결함 목록

| id | 심각도 | 분류 | 위치 태그 | 위치 | 내용 | Required fix / 상태 |
|---|---|---|---|---|---|---|
| D20 | Medium | **blocking**, repair-born 아님(위치 기준) | 카드 diff 안(642d049cc6), 수리 델타 밖 | `CHANGELOG.md:17` | 제외된 B3를 출시 기능으로 설명한다. 제목 절 "and a receipt survives evidence-only commits", 항목 (3) "**Receipt binding across evidence-only commits** (`internal/auditreceipt/store.go`, `internal/hook/audit_receipt_guard.go`): an audit receipt stays valid when the commits since the audited head change only audit evidence …", 그리고 "16 requirements REQ-AEC-001..016, 16 acceptance sub-criteria"가 그대로다. 실제 동작(HEAD가 움직이면 항상 stale)과 반대이고, SPEC의 14/14 및 REQ-MN-020 "not relaxed"와도 맞지 않는다. codex도 같은 줄을 지적했다. | CHANGELOG 17행만 고친다. 제목 절에서 "a receipt survives evidence-only commits"를 빼고, 항목 (3)을 지우거나 "B3(evidence-only 커밋 뒤 영수증 신선도 완화)는 운영자 결정으로 제외됐고, 감사 뒤 HEAD를 옮기는 커밋은 영수증을 stale로 만든다(SPEC-MCP-NEUTRALITY-001 REQ-MN-020 그대로)"로 바꾼다. 개수는 "14 live requirements (REQ-AEC-009/010 retired), 14 live acceptance sub-criteria (AC-AEC-003a/b retired)"로 고친다. 뒤 항목 번호 (4)·(5)는 필요하면 다시 매긴다. 기존 경로를 없애는 수리라서 A1 규칙 범위 안이다. 확인 감사 범위는 a50cf15d64 이후 델타(이 파일 한 줄)다. |
| D21 | Low | optional | 카드 diff 안, 수리 델타 밖 | `.moai/specs/SPEC-AUDIT-ENGINE-CONVERGENCE-001/progress.md:212`, `:243` | §E.3 표의 "M5 … PASS (2 of 2)"와 §E.4 `b12_self_test_b: "… live=16 … matches the CHANGELOG count"`가 제외 전 상태를 현재 값처럼 적는다. §E.2 M5 절(136행~)은 실행 이력이라 남겨도 된다. | 다음 sync 갱신 때 §E.3 M5 행에 "retired (B3 excluded 2026-11-06)" 표시를 달고, §E.4 개수를 14로 다시 잰다. 판정에는 영향이 없다. |
| D22 | Low | optional | in-diff 파일, 변경되지 않은 줄 | `internal/hook/audit_receipt_range_test.go:134` | `requireFresh`의 실패 메시지가 "was refused after evidence-only commits"다. 지금은 HEAD가 그대로인 대조군에서만 쓰이므로 문구가 실제와 다르다. 단언 자체는 맞다. | 메시지를 "was refused at an unchanged head" 정도로 바꾼다. 재량 사항이다. |
| D23 | Low | optional | in-diff | `internal/hook/audit_receipt_range_test.go`(지운 행) | "uncommitted 기록이 없는 옛 영수증 + 증거 커밋 → stale" 행이 지워졌다. 지금 규칙에서는 모든 커밋이 stale이므로 일반 행이 같은 결과를 덮는다. `uncommitted_digest`를 가진 영수증의 로드는 저장소 시험이 고정하지 않는다(이 감사가 스크래치 사본에서 따로 확인함). | 원하면 `uncommitted_digest` 키를 가진 영수증의 로드·판정 시험을 하나 둔다. 재량 사항이다. |
| D19, D18, D16, D13, D1, D10 | — | 해소(대상 소멸) | — | `store.go` e2 함수군 | 모든 결함이 B3 함수군 안에 있었다. 그 코드가 통째로 지워졌다(d15679fc78 대비 diff 없음). | 닫힘 |
| D2–D9, D11, D12, D14, D15, D17 | 3·4회차 기록 | optional | — | 3·4회차 표 참조 | 이번에 다시 재지 않았다 | 이전 기록을 그대로 옮긴다(이번 실행의 증거 아님) |

### D20을 blocking으로 보는 근거

- 이번 수리의 목적은 B3를 "남김없이" 빼는 것이었다. CHANGELOG는 같은 카드가 sync 단계에서 쓴 산출물이다(642d049cc6). 그런데 지금 사용자에게 없는 동작을 있다고 알린다. 문서가 정확한지의 문제이고, 카드 diff 안에 있다.
- 필수 게이트 codex가 같은 줄을 독립적으로 짚었다.
- 수리 비용은 한 줄이다. A1 규칙("결함 경로 제거, 기존 검사 조이기, 시험 추가만")에도 맞는다.
- 반복 수리 중지 규칙: D20은 985853c683..HEAD diff 밖에 있으므로(이 델타는 CHANGELOG를 건드리지 않았다) 위치 기준으로 repair-born이 아니다. 5회차 D19는 repair-born이었지만 연속 2회 조건은 성립하지 않는다. 다만 원인은 이번 제거 작업이 문서 갱신을 빠뜨린 데 있다.

## 판정 항목별 결과 (리더 지시 1~6)

1. **B3 완전 제거**: 코드, 시험, SPEC, 에이전트 프롬프트에는 relaxation을 설명하거나 구현하는 현재형 문장이 없다. `.go`에서 지운 이름 0건이다. 템플릿과 `.claude`에도 0건이다. 남은 것은 CHANGELOG(D20, blocking)와 SPEC progress.md 이력·요약(D21, optional)뿐이다.
2. **신선도 fail-closed**: 코드, reports만, progress만, 빈 커밋, 커밋 뒤 미커밋 수정, HEAD 그대로(fresh/미커밋 수정 stale)를 시험으로 재현했다. HEAD 비후손과 이름 바꾸기 형태도 stale이다.
3. **B1/B2/B4/B5 보존**: 지워진 시험 함수는 모두 B3 전용 파일이거나 B3 전용 함수다. 그 파일들은 M5 이후 커밋(83f776335f, c38c9bdf96, 985853c683, 75876111d8)이 만들었다. `TestTreeState_DigestFormatIsUnchanged`는 지워진 뒤 새 파일로 그대로 옮겨졌다. `TestCheckCitedReceiptsSince_AllowsNoProgressFile`의 stale 단언은 새 시험 행이 덮는다. `TestAEC003b_AnyOtherChange…`와 `TestAEC003b_HeadUnchanged…`는 `TestAEC_…`로 이름만 바뀌었고 행과 단언이 같다(옛 영수증 행 하나는 D23). fresh를 기대하던 `TestAEC003a_*` 단언은 `TestAEC_EvidenceOnlyCommits_LeaveTheReceiptStale`에서 stale로 뒤집혔다. 이것은 의도된 변화다. B1/B2/B4/B5 코드 파일은 이 델타에서 바뀌지 않았다.
4. **`uncommitted_digest` 영수증 로드**: 스크래치 사본에서 확인했다. 로드는 오류 없이 되고, HEAD가 그대로면 fresh, 증거 커밋 뒤에는 stale이다. `readJSON`은 `json.Unmarshal`이고 `DisallowUnknownFields`를 쓰지 않는다.
5. **SPEC 일관성**: 살아 있는 REQ 14개(`**REQ-AEC-0NN**` 줄 14개, `[RETIRED]` 2개 별도), plan.md 18·19행 14/14다. 은퇴한 ID 네 개는 세 문서 24곳 모두 `[RETIRED]`가 바로 붙어 있다. REQ-MN-020 행은 "Not relaxed by this SPEC"이다. strict lint는 무결함이다.
6. **A1 수리 규칙**: 델타는 결함 경로 제거(store.go, guard.go 호출 복원), 시험 삭제와 추가, 운영자 결정에 따른 SPEC 제외 개정, 레인 기록뿐이다. 새 verb, flag, 입력, 저장소, 규칙은 없다.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

1. 감사 대상은 HEAD a50cf15d6408이고, 트리는 시작과 끝 모두 깨끗하며 바뀌지 않았다.
2. store.go와 guard.go는 M5 이전 상태로 정확히 돌아갔고, B3 식별자는 코드에 남아 있지 않다.
3. 영수증 신선도는 모든 커밋 형태에서 fail-closed이고, 시험이 그것을 고정한다(뮤턴트로 확인).
4. 옛 `uncommitted_digest` 영수증도 로드된다.
5. SPEC 세 문서는 14/14로 서로 맞고 lint는 무결함이다.
6. CHANGELOG.md:17은 제외된 기능을 현재 기능으로 설명한다(D20).
7. codex 필수 게이트는 fail을 보고했다.

### Evidence (명령 → exit → 판정에 쓴 출력)

- `git rev-parse HEAD` → `a50cf15d64d5a80ae259cbaf54ae2adbbfcdc01f`(시작, 판정 파일 작성 직전 모두). `git status --short` → 출력 없음(두 번 모두). `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2017`.
- `git diff --stat 985853c683..HEAD` → `15 files changed, 673 insertions(+), 1305 deletions(-)`(store.go `284 +-------------`, 지운 시험 파일 4개, 새 시험 파일 3개, SPEC 3개, 레인 기록 2개).
- `git log --oneline -15 -- internal/auditreceipt/store.go` → M5 첫 커밋 `83f776335f`, 직전 커밋 `d15679fc78`.
- `git diff --stat d15679fc78 HEAD -- internal/auditreceipt/ internal/hook/audit_receipt_guard.go` → `store_stale_after_commit_test.go | 106 +++`, `store_treestate_format_test.go | 25 +++`, `2 files changed, 131 insertions(+)`(store.go와 guard.go 없음). `git diff d15679fc78 HEAD -- internal/auditreceipt/store.go internal/hook/audit_receipt_guard.go` → 출력 없음.
- `git diff 985853c683..HEAD -- internal/hook/audit_receipt_guard.go` → `-CheckCitedReceiptsForSpec(g.store, start, boundary, cited, specID)` / `+CheckCitedReceiptsSince(g.store, start, boundary, cited)`.
- `grep -rnE 'evidenceOnlyRange|e2RawRegion|e2Unchanged|e2DocsUnchanged|e2HeadingLike|e2BoundaryPattern|UncommittedDigest|uncommitted_digest|CheckCitedReceiptsForSpec' --include='*.go' .` → 출력 없음.
- `git grep -nE '<같은 패턴>' HEAD -- . ':!.moai/reports'` → `4 .moai/specs/SPEC-AUDIT-ENGINE-CONVERGENCE-001/progress.md`(실행 이력 줄 153, 155, 222, 227)만 나온다.
- `git grep -nE -i 'stays fresh|stays valid|survives evidence|receipt binding across|across evidence-only|…' HEAD -- . ':!.moai/reports' ':!…/progress.md'` → 관련 줄: `CHANGELOG.md:17`, 그리고 SPEC 문서의 `[RETIRED]`/제외 문장들. `.claude/`, `internal/template/`에는 관련 줄이 없다.
- `sed -n '17p' CHANGELOG.md` → "… and a receipt survives evidence-only commits (card t2017, Tier M; 16 requirements REQ-AEC-001..016, 16 acceptance sub-criteria …) … (3) **Receipt binding across evidence-only commits** (`internal/auditreceipt/store.go`, `internal/hook/audit_receipt_guard.go`): an audit receipt stays valid when the commits since the audited head change only audit evidence …". `git log --oneline -S'Receipt binding across evidence-only commits' -- CHANGELOG.md` → `642d049cc6 [t2017] docs(SPEC-AUDIT-ENGINE-CONVERGENCE-001): sync-phase artifacts (card t2017)`.
- `grep -cE '^\s*-\s+\*\*REQ-AEC-0[0-9]+\*\*' spec.md` → `14`. `grep -noE '(REQ-AEC-009|REQ-AEC-010|AC-AEC-003a|AC-AEC-003b)( \[RETIRED\])?' spec.md plan.md acceptance.md | …` → `6 acceptance.md tok`, `8 plan.md tok`, `10 spec.md tok`(BARE 0). `spec.md:227` → "| SPEC-MCP-NEUTRALITY-001 REQ-MN-020 (freshness) | … | Not relaxed by this SPEC: the rule stands as written, and any commit that moves HEAD after an audit makes the receipt stale (B3 excluded, 2026-11-06) |".
- `<scratch>/moai-tree spec lint --strict SPEC-AUDIT-ENGINE-CONVERGENCE-001` → exit 0, `✓ No findings — all SPEC documents are valid`.
- `<scratch>/moai-tree verify audit-plan --project-root <toplevel>` → `config_status: ok`, `model: codex`, codex `required`(config.gates, explicit), claude/glm `off`, `cross_model_active: true`, `enforced_required: ["codex"]`.
- `<scratch>/moai-tree decision read --scope card:t2017` → `id=d-20261106T074340Z-19c9 scope=card:t2017 kind=ruling`, 본문 "운영자 결정 2026-11-06 (원문 답 "B3 빼고 착지 (권장)" …): t2017에서 B3(…)를 제외하고 감사 뒤 커밋이면 영수증을 항상 stale로 처리, B1·B2·B4·B5로 착지. 제외 델타 재감사에 codex_audit 1회(3회차) 허용."
- `go test ./internal/auditreceipt/... -count=1 -cover` → exit 0, `ok github.com/modu-ai/moai-adk/internal/auditreceipt 32.775s coverage: 86.2% of statements`.
- `go test ./internal/hook/ -run TestAEC -count=1 -v -timeout 20m` → exit 0, `ok github.com/modu-ai/moai-adk/internal/hook 24.043s`. 최상위: `--- PASS: TestAEC_AnyChangeAfterTheAudit_MakesTheReceiptStale (14.97s)`(하위 15개 PASS: `a_.go_file`, `spec.md`, `a_report_outside_the_audit-artifact_family`, `a_path_that_only_begins_with_the_allowed_prefix_text`, `exactly_one_line_under_E.2`, `the_E.2_heading_deleted`, 이름 바꾸기 2종, `the_E.2_section_moved_below_E.3`, `the_progress.md_of_another_SPEC`, `a_progress.md_while_the_citation_names_a_card_id`, rename 2종, `HEAD_is_not_a_descendant_of_the_judged_commit,_whatever_its_tree`, `an_uncommitted_code_edit_exists_after_an_evidence-only_commit`), `--- PASS: TestAEC_HeadUnchanged_IsJudgedAsBefore_Control (2.45s)`(하위 3개: `nothing_changed`, `an_uncommitted_code_edit_at_the_same_head`, `a_legacy_receipt_at_the_same_head`), `--- PASS: TestAEC_EvidenceOnlyCommits_LeaveTheReceiptStale (5.62s)`(하위 7개: `a_verdict_file_and_a_second-review_record`, `every_member_of_the_audit-artifact_family`, `the_E.1_and_E.4_lines_of_the_cited_SPEC`, `a_full_rewrite_of_E.3_with_E.2_byte-unchanged`, `a_reports-only_commit_while_the_citation_names_a_card_id`, `an_empty_commit`, `a_receipt_that_judged_an_uncommitted_edit,_then_an_evidence-only_commit_with_the_edit_untouched`).
- `go vet ./internal/auditreceipt/... ./internal/hook/` → exit 0, 출력 없음.
- `GOOS=windows go build ./...` → exit 0, 출력 없음.
- `golangci-lint run --timeout=5m ./internal/auditreceipt/... ./internal/hook/...` → exit 0, `0 issues.`.
- 시험 함수 비교(`git diff 985853c683..HEAD -- '*_test.go'`의 `-func`/`+func`): 지움 = `TestAEC003a_EvidenceOnlyCommits_KeepTheReceipt`, `TestAEC003a_UncommittedStateAlreadyJudged_KeepsTheReceipt`, `TestAEC003b_AnyOtherChange_MakesTheReceiptStale`, `TestAEC003b_E2FenceShapes_*` 2개, `TestAEC003b_HeadUnchanged_IsJudgedAsBefore_Control`, `TestCheckCitedReceiptsSince_AllowsNoProgressFile`, `TestE2DocsUnchanged_*` 8개, `TestE2RawRegion_*` 5개, `TestEvidenceOnlyRange_Edges`, `TestTreeState_DigestFormatIsUnchanged`, `TestUncommittedState_IsHeadIndependent`, `TestWriteReceipt_RecordsUncommittedStateOnlyWhenTheBindingHolds`와 도우미. 추가 = `TestAEC_AnyChangeAfterTheAudit_MakesTheReceiptStale`, `TestAEC_EvidenceOnlyCommits_LeaveTheReceiptStale`, `TestAEC_HeadUnchanged_IsJudgedAsBefore_Control`, `TestCheckCitedReceipts_AnyCommitAfterTheAuditLeavesTheReceiptStale`, `TestTreeState_DigestFormatIsUnchanged`와 도우미.
- `git diff --stat 75876111d8^ HEAD -- internal/hook/audit_receipt_range_test.go …` → `audit_receipt_range_test.go | 288 +++`(M5 RED 커밋 전에는 파일이 없었다).
- 프로브(스크래치 사본 `go -C <copy> test -count=1 -v -run TestAuditor6_ ./internal/auditreceipt/`) → exit 0:
  - `moveHead=false ReadReceipt err=<nil> head_equal=true digest_equal=true`
  - `moveHead=false CheckCitedReceiptsSince=(true, "")`
  - `moveHead=true ReadReceipt err=<nil> head_equal=true digest_equal=true`
  - `moveHead=true CheckCitedReceiptsSince=(false, "receipt judged a different tree state (head or digest moved)")`
- 뮤턴트(스크래치 사본 `store.go:1712`의 `r.HeadCommit != head || r.TreeDigest != digest`를 `r.HeadCommit == head && r.TreeDigest != digest`로 바꿈): `go -C <copy> test -count=1 -run 'TestCheckCitedReceipts_AnyCommit' ./internal/auditreceipt/` → exit 1, `--- FAIL: TestCheckCitedReceipts_AnyCommitAfterTheAuditLeavesTheReceiptStale`(하위 6개 모두 FAIL: verdict 파일, sync-audit+second-review, §E.3, §E.2, 코드 파일, 빈 커밋).
- `mcp__moai__codex_audit`(1회) → `verdict: fail`, `audit_receipt: rcpt-be71757522b8572b0da40751`, `build_commit: 04d8c27a7e`, 지적 `[P2] … CHANGELOG.md:17`.
- `git check-ignore -v .moai/reports/t2017/sync-audit-6.md` → `.gitignore:235:.moai/reports/*`.

### Baseline-attribution

- 측정 트리: a50cf15d64d5a80ae259cbaf54ae2adbbfcdc01f(작업 트리 깨끗, 이번 실행).
- 비교 기준: 985853c683565e6f765ad17608d4526760804333(5회차 감사 SHA)와 d15679fc78(M5 직전 커밋, 복원 기준).
- Go 시험, vet, lint, 빌드는 이 트리 소스를 로컬 `go` 툴체인과 비공개 GOCACHE(스크래치 아래)로 컴파일해 쟀다.
- `moai` 측정(spec lint, audit-plan, decision read)의 판정 빌드: 이 트리에서 `go build -o <scratch>/moai-tree ./cmd/moai`로 만든 바이너리를 경로로 직접 실행했다(§2.2 첫째 방식). 이 빌드는 ldflags가 없어서 `moai-tree version`이 `v3.2.0 none built unknown`을 출력한다. 빌드의 출처는 트리 HEAD a50cf15d64다. 설치본 `/Users/goos/go/bin/moai`(`v3.2.1-rc.2 v3.2.0-1356-g04d8c27a7e`)는 판정에 쓰지 않았다. 04d8c27a7e는 HEAD의 조상이 아니다(`git merge-base --is-ancestor` exit 1).
- codex는 MCP 서버 빌드 `04d8c27a7e`(v3.2.1-rc.2)로 호출했다. 이 값은 codex 결과의 출처일 뿐이고, 코드 판정 근거는 위의 Go 측정이다.
- 프로브와 뮤턴트는 `git archive -o <scratch>/head.tar HEAD`를 새 빈 디렉터리 `<scratch>/a6copy`에 풀어 만든 사본에서만 돌렸다. 추적 트리에는 판정 파일 말고는 아무것도 쓰지 않았다.

### Gaps

- G1. 계획은 `cross_model_active: true`를 보고했지만, 지시에 따라 `audit_multi`를 호출하지 않았다. convergence digest(`overall_verdict`, `gate_unmet`, `plan_source`)와 orchestrator의 `--result-file` 확인이 없다.
- G2. codex는 adversarial 모드라 범위가 기계적으로 묶이지 않았다(`prose-only scope`). codex는 집중 시험이 권한 거부로 NOT_RUN이었다고 스스로 밝혔다. 재호출은 하지 않았다(운영자 허용 3회 소진).
- G3. 거부된 명령이 1건 있다. GOCACHE를 `export`하고 이어서 빌드와 vet를 도는 복합 명령을 워크트리 격리 가드가 "construct too complex to verify" 사유로 거부했다. 같은 명령을 `GOCACHE=<literal> go …` 형태의 단일 명령으로 나눠 다시 실행했고, 그 결과를 썼다. 대체한 측정은 없다.
- G4. 실행하지 않은 것: 두 패키지 전체 스위트 중 hook은 `-run TestAEC`만, race detector, Windows 시험 실행(빌드만), `/verify`, `internal/cli` 넓은 선택.
- G5. D2–D9, D11, D12, D14, D15, D17은 다시 재지 않았다(3·4회차 기록 인용). 이 결함들이 B3 함수군 밖에 있는지도 이번에 따로 확인하지 않았다.
- G6. "B1/B2/B4/B5 동작 보존"은 시험 함수 비교, 코드 파일 무변경(이 델타가 B1/B2/B4/B5 파일을 건드리지 않음), 지정 시험 통과로 판단했다. 그 기능들의 시험 스위트 전체를 다시 돌리지는 않았다.
- G7. 문서 검색 범위는 저장소 `git grep`과 `.claude`/`internal/template` grep에 한정했다. docs-site 등 저장소 밖 문서는 보지 않았다.

### Residual-risk

- D20을 고친 뒤에도 같은 카드의 다른 sync 산출물(MCP 도구 카탈로그, cross-model-audit 스킬)에 B3를 다른 말로 표현한 문장이 남아 있을 수 있다. 이번 grep 패턴("evidence-only", "stays fresh/valid", "survives evidence", "receipt binding across")에는 걸리지 않았다.
- SPEC progress.md의 M5 이력(§E.2 136행~)은 실행 기록으로 남는다. 읽는 사람이 현재 동작으로 오해할 여지가 있다(D21).
- `uncommitted_digest` 영수증의 로드는 저장소 시험이 고정하지 않는다. 나중에 `DisallowUnknownFields`가 들어오면 옛 영수증이 깨질 수 있다(D23).
- 판정 파일의 `blocking_defect:` id를 감사자 자신이 쓴다는 점(§E B1)은 바뀌지 않았다.
- 스크래치의 프로브·뮤턴트 출력은 기계 로컬이다. 판정에 쓴 줄은 위 Evidence에 옮겨 적었고, 나머지는 소실을 감수한다.

## 반복 이력

| 회차 | 감사 커밋 | 판정 | 주요 변화 |
|---|---|---|---|
| 1 | 1008b915b2 | FAIL (D1 BLOCKING) | fence boolean 토글이 §E.2를 가렸다 |
| 2 | d9665c52c5 | FAIL 0.73 (D10 BLOCKING) | D1은 닫혔지만 빈 목록 비교가 남았다 |
| 3 | 400caf47ba | PASS-WITH-DEBT 0.85, receipts none | Claude 단독이라 영수증 가드가 거부했다(codex required) |
| 4 | 7ea9b20555 | FAIL 0.73 (D16, D13 BLOCKING), codex fail `rcpt-919c8ec5ced34525ba7663b0` | D16 신규, D13 재분류 |
| 5 | 985853c683 | FAIL 0.75 (D19 BLOCKING, repair-born), codex fail `rcpt-7911463ec3cd51004d46c029` | 인용 경계 fail-open(D19), 가용성 회귀 D18 |
| 6 | a50cf15d64 | FAIL 0.84 (D20 BLOCKING, repair-born 아님), codex fail `rcpt-be71757522b8572b0da40751` | 운영자 결정으로 B3 제외, 코드는 M5 이전으로 복원, D19·D18 등 대상 소멸. CHANGELOG에 B3 설명이 남음(D20) |

## 다음 단계 (리더·운영자 판단 사항)

- D20은 CHANGELOG.md 17행만 고치면 닫힌다(위 Required fix). 확인 감사 범위는 a50cf15d64 이후 델타이고, blocking id는 D20이다.
- 이 카드의 codex 호출은 운영자 허용분까지 모두 썼다(3회). 확인 감사에 codex 영수증이 필요하면 4번째 호출이 되므로 운영자 승인이 다시 필요하다. 승인 없이는 필수 게이트를 채울 수 없고, PASS 계열 판정도 나올 수 없다.
- D21~D23은 optional이다. D20 수리와 함께 D21(progress.md §E.3/§E.4 표기)을 정리하면 다음 감사의 혼선이 줄어든다.
