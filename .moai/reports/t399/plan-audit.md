# SPEC Review Report: SPEC-CODEX-REVIEW-TARGET-001

- Iteration: 1/2 (Tier M ceiling)
- **Verdict: FAIL**
- **Overall Score: 0.75** (Tier M PASS threshold 0.80 — `spec-workflow.md:140-152`)
- 감사 트리: `.claude/worktrees/t399` @ `442da4f06`, 브랜치 `WT-codex-native-branch`
- 감사 입력(Tier M 계약): `spec.md` · `plan.md` · `acceptance.md` (+ `progress.md` 참고)
- Reasoning context ignored per M1 Context Isolation. 아래 모든 판정은 파일 관측이다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — `REQ-CRT-001..006` 연속, 결번·중복 없음(`spec.md:97,101,105,109,113,117`). AC 도 `AC-CRT-001..009` 연속(`acceptance.md:51..112`).
- **[PASS] MP-2 GEARS 형식** — 6건 전부 GEARS 패턴. 001/006 Ubiquitous, 002 event-driven(`When ... shall`), 003 state-driven(`While ... shall prefer`), 004 legacy Unwanted(`If ... shall not / shall`), 005 canonical Unwanted(`shall not`). 판정 층위: **requirement layer(`spec.md` 의 REQ-CRT-*)** 에 대해 매겼다. `acceptance.md` 의 Given-When-Then 은 verification layer 이므로 여기서 감점하지 않았다(Group 4 소관). 004 의 legacy `If` 형에 `[DEPRECATED — use shall not]` 주석이 없다 → backward-compat 창(2026-11-22) 안이라 MINOR(D8)이며 FAIL 아님.
- **[PASS] MP-3 프론트매터 유효성** — 12 정본 필드 전부 존재·타입 정합(`spec.md:2-15`): `id` / `title` / `version`("0.1.0" 인용) / `status`(draft) / `created`·`updated`(ISO) / `author` / `priority`(P0) / `phase` / `module` / `lifecycle`(spec-anchored) / `tags`(CSV 문자열). 거부되는 snake_case alias(`created_at` / `updated_at` / `labels` / `spec_id`) 없음. `tier: M` 은 선택 필드로 추가 존재.
- **[N/A] MP-4 언어 중립성** — 단일 프로그래밍 언어(Go) 범위 SPEC. `internal/cli` 한정이며 16개 언어 도구 표면을 다루지 않는다.
- **[PASS] MP-5 D7 교차-SPEC** — 참조된 외부 SPEC 은 `SPEC-CODEX-VERDICT-SYNTH-001` 하나. `grep '^status:' .moai/specs/SPEC-CODEX-VERDICT-SYNTH-001/spec.md` → `status: completed`. retired / superseded / archived 아님 ⇒ BLOCKING 없음.
- **[PASS] MP-6 D8 크로스플랫폼** — `grep -c syscall spec.md` → `0`. 자동 PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-REVIEW-TARGET-001/` → 0 hit (rc=1). 미해결 마커 없음.

Must-pass 7건 중 실패 0. **이 FAIL 은 must-pass 위반이 아니라 Tier M 임계 미달 + 아래 blocking 결함에서 나온다.**

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 (한둘의 요구가 해석을 요구) | REQ-CRT-006 "byte-identical **in shape**"(`spec.md:119`)가 자기모순(바이트 동일 ≠ 형태 동일)이고 AC-CRT-003 은 형태만 잰다(`acceptance.md:70`). REQ-CRT-005(`spec.md:115`)는 금지만 있고 양성 귀결이 없다. |
| Completeness | 0.75 | 0.75 (비핵심 누락) | 필수 섹션·프론트매터 완비, Out of Scope H3 5개 각각 `-` 불릿 보유(`spec.md:142,149,155,160,166`). 누락: 저장소 안의 라이브 왕복 선례 미인용(D1), config↔origin/HEAD 해석 분기 미기재(D2), AC-CRT-009 를 뒷받침하는 REQ 부재(D6). |
| Testability | 0.70 | 0.50↔0.75 straddle | 대부분의 AC 가 `sess.sent[2]` 직렬화 바이트를 관측해 이례적으로 강하다. 그러나 AC-CRT-005 ↔ AC-CRT-006 이 commit/custom 행에서 모순(D3), AC-CRT-004 의 "원인 문자열"이 어느 필드인지 미지정(D4), REQ-CRT-003 3·4 단계가 원리상 구별 불가(D5), AC-CRT-007/009 는 실행 검사가 아닌 판독 판정. |
| Traceability | 0.80 | 0.75 (한 건이 간접 매핑) | REQ 6건 전부 ≥1 AC(`acceptance.md:39-47`), 존재하지 않는 REQ 를 참조하는 AC 없음. 단 AC-CRT-008(RED 증거 보존)은 REQ-CRT-002 에 매핑돼 있으나 실제로는 `acceptance.md §C` 규율이지 REQ-002 의 행동 검증이 아니고, AC-CRT-009 는 대응 REQ 가 없다(D6). |

산술평균 (0.75 + 0.75 + 0.70 + 0.80) / 4 = **0.75** < 0.80.

---

## 리드가 지정한 네 질문에 대한 답

### Q1 — 회귀선이 완화됐는가: **그렇다. 완화이며, 정당화가 없다. (D1, blocking)**

카드의 [HARD] 문언(큐 원문 재판독): "회귀 [HARD]: native 경로가 **실제로 codex 에 도달했는지**를 단언할 것."
SPEC §0(`spec.md:33`)은 이를 "직렬화된 요청이 측정된 스키마의 required 집합을 만족한다"로 바꾸고, `acceptance.md:131` 은 "이 SPEC 의 어떤 AC 도 실제 codex 바이너리와의 왕복을 관측하지 않는다 … 계약이 곧 코덱스의 수락을 뜻한다는 것은 스키마 문서의 주장이지 이 SPEC 의 측정이 아니다"라고 스스로 인정한다.

완화라고 판정하는 근거는 셋이고 전부 이 트리에서 측정했다.

1. **환경이 가능하다.** `/Users/goos/.local/bin/codex --version` → `codex-cli 0.150.1`, rc=0.
2. **선례가 저장소 안에 이미 있다.** `internal/cli/codex_live_protocol_probe_test.go:507` `TestCodexLive_ReviewStartEmitsTurnStarted` 가 **실제 codex 세션을 열어 `review/start` 를 라이브로 보낸다**(`:533`, `target: codexTargetUncommitted`). 부속 키트도 완비다 — `probeLiveCodex` / `probeSeedRepo`(`:548`, 임시 git repo 생성) / `probeInstallRunner` / `probeWriteTranscript`. 별도로 `internal/cli/codex_review_gate_live_test.go:33` 이 skip 조건 3종(바이너리 부재 / `--version` 비정상 / `MOAI_SKIP_LIVE_CODEX=1`)까지 확립해 뒀다. 즉 라이브 왕복은 **단순성 사다리 2단(이미 있는 것 재사용)**이지 새 발명이 아니다.
3. **비용 논거가 성립하지 않는다.** 스키마 위반은 turn 이 시작되기 전에 JSON-RPC 오류로 즉시 거절된다. 리뷰 turn 을 끝까지 돌릴 필요가 없다 — 위 probe 가 `turn/started` 에서 세션을 끊어 "full review turn 을 청구서에서 뺀다"고 명시한 것과 같은 기법이다(`:505-507`).

SPEC 은 이 선례를 **어디에서도 인용하지 않는다**(spec / plan / acceptance 3파일에 `live` 언급 0). §A.3 에서 해석기에 대해서는 "새 해석기를 발명할 필요가 없다"는 재사용 규율을 정확히 적용해 놓고, **검증 표면에는 같은 규율을 적용하지 않았다.** 그래서 이것은 "라이브가 불가능해서 계약으로 대신했다"가 아니라 "라이브가 가능한데 계약으로 대신했다"이며, 후자에는 근거 문장이 필요한데 없다.

**닫는 AC (이 SPEC 안에 두어야 한다):**

> **AC-CRT-010 — 라이브 codex 가 baseBranch 요청을 거절하지 않는다**
> **Given** 실 codex 바이너리(PATH 존재 + `--version` 정상, 부재 시 skip, `MOAI_SKIP_LIVE_CODEX=1` 로 opt-out)와 base 브랜치 및 그로부터 갈라진 HEAD 를 가진 픽스처 git repo 가 주어지고,
> **When** `mode=native`, `target=baseBranch`, `project_root=<픽스처>` 로 감사가 실행되면,
> **Then** `review/start` 에 대한 응답이 JSON-RPC error 가 **아니고**(특히 missing field `branch` 계열이 아니고), 그 turn 이 `turn/started` 에 도달한다. 판정은 `inconclusive` 여부가 아니라 **거절 부재라는 양성 사실**이다.

**이 SPEC 에 두어야 하는 이유:** (a) 카드의 [HARD] 문언이 정확히 이것을 지목한다 (b) 현행 트리에서 이 AC 는 **RED 다** — 0.150.1 이 `{"type":"baseBranch"}` 를 거절하므로, 이 SPEC 이 확보 가능한 증거 중 가장 강한 RED 이며 `acceptance.md §C` 의 RED 규율이 요구하는 관측을 스키마 문서가 아니라 실물에서 얻는다 (c) `acceptance.md §F` 의 "별도 관측 항목으로 남긴다"는 처분은, 같은 SPEC 이 §A.4 에서 "기존 검사가 초록인 것은 교정의 근거가 아니다"라며 배격한 바로 그 종류의 미룸이다.

부가로 이 AC 는 **현재 아무도 재지 않은 전제** 하나를 함께 잰다 — codex 가 `baseBranch` 의 값으로 **로컬에 존재하지 않는 이름**(예: `origin/develop` 에서 접두사를 뗀 `develop`)을 수용하는지. plan §C 는 그 이름을 보내기로 했으면서 수용 여부를 관측한 적이 없다(D5 와 연결).

### Q2 — §A.4 의 전제 정정: **확인. 정확하다. (표 좌표 1건만 부정확)**

행동 기준 주장은 4행 전부 참이다.

| §A.4 주장 | 검증 | 결과 |
|---|---|---|
| 리프트 검사가 `uncommittedChanges` 로만 잰다 | `codex_review_rpc_test.go:63` `TestCodexRPC_TargetIsTaggedObject_BareStringLifted` — 인자 `"target": codexTargetUncommitted`(`:67`), 단언 `target["type"] != codexTargetUncommitted`(`:83`) | **확인** |
| bare string 미직렬화 검사도 같은 variant 로만 | `:96` `TestCodexRPC_BareStringTargetNotSentInReviewRequest` — `bad := "\"target\":\"" + codexTargetUncommitted + "\""`(`:105`) | **확인** |
| native dispatch 검사도 같은 variant 로만 | `mcp_codex_test.go:183` — `"target": codexTargetUncommitted`(`:188`) | **확인** |
| `baseBranch` 를 든 유일한 codex 검사가 adversarial 경로라 target 을 아무것도 단언하지 않는다 | `mcp_codex_test.go:216` `TestCodexAudit_AdversarialDispatchesTurnStart` — `"target": codexTargetBaseBranch`(`:231`), 단언은 `sess.sent[2]` 에 `turn/start` 포함(`:236`)과 verdict fail(`:239`) 둘뿐. target 형태 단언 0 | **확인** |
| 저장소 안 유일성 | `grep -rn 'codexTargetBaseBranch\|"baseBranch"' --include=*.go internal/` → 테스트 파일 hit 은 `mcp_codex_test.go:231` **한 줄뿐**. 나머지는 프로덕션(`mcp_codex.go:125,992`, `mcp_server.go:255,396`, `mcp_review_material.go:75`) | **확인** |

즉 리드의 원 전제("`coerceCodexReviewTarget` 커버리지 없음")를 SPEC 이 뒤집은 것은 **옳고**, 뒤집은 방향의 함의("공허한 초록 — 고정 지점이 하필 리프트가 옳은 유일한 variant")도 옳다. 이것은 이 SPEC 의 가장 좋은 부분이다.

좌표 정확도만 어긋난다(D7, MINOR): §A.4 표는 `codex_review_rpc_test.go:62` / `:94` 를 인용하나 실제 `func` 선언은 `:63` / `:96` 이다(인용된 줄은 doc comment). 같은 표의 `mcp_codex_test.go:183` 은 `func` 줄, `:231` 은 인자 줄 — 한 표 안에서 좌표 관례가 셋으로 갈린다.

### Q3 — 축-2 경계: **코드상 분리는 실재한다. 남는 것은 이슈 응답 문언이다.**

**분리 검증(확인):**

- 이 SPEC 이 건드리는 지점: `mcp_codex.go:995-1005` `coerceCodexReviewTarget`, `mcp_codex.go:939-946` `buildCodexReviewParams`, `mcp_server.go:255` 도구 표면.
- t284 가 건드릴 지점: `mcp_convergence.go` Step 2 disagreement 유도부 — `distinctRequired := distinctVerdicts(required, "pass", "fail")` / `disagreement := len(distinctRequired) > 1`. SPEC §E 가 인용한 `:168-227` 범위와 실제 위치가 일치한다.
- 같은 파일도 같은 함수도 아니다 ⇒ §E 의 분리 주장 **성립**.

**핵심 기제 확인(왜 t284 가 별개인지):** `distinctVerdicts(required, "pass", "fail")` 는 `inconclusive` 를 애초에 세지 않는다. 따라서 codex 가 required 인데 `inconclusive` 를 내고 claude 가 pass 면 `distinctRequired == {pass}` → `disagreement == false`. "합의함"과 "비교할 것이 없음"이 같은 값이 되는 지점은 정확히 여기이고, `mcp_codex.go` 의 요청 조립부가 아니다.

**t284 카드 실물 확인:** `moai todo` 에서 t284 는 `queued` 이고 본문이 (1) on-target 백엔드 수 노출 (2) 참여자 2 미만이면 `disagreement_flag=false` 금지 (3) 대표 mutant 를 이미 명세한다 — §E 의 서술과 일치. 다만 §E 는 "근거(측정)"으로 `.moai/reports/t229/succession.md` 를 든다. 그 문서는 축의 존재와 2026-08-26 운영자 처분("별도 신규 카드")을 기록하지만 **t284 라는 이름은 담고 있지 않다** — 그 연결은 큐 텍스트에서 온다. discovery.md 의 Gaps 가 이 점을 정직하게 적어 둔 것과 §E 의 "근거(측정)" 표기가 어긋난다(D7 에 병합, MINOR).

**"제보자에게 오도되는가"에 대한 판정: 아니다, 조건부로.** 이 SPEC 이 닫는 것은 #1632 의 1축이고, 착지하면 `mode=native, target=baseBranch` 가 **더는 스키마 위반으로 거절되지 않는다** — 제보자가 첨부한 원문 JSON-RPC 오류가 사라진다. 실질적 수리다. 다만 착지 후에도 "required native gate 가 조용히 아무것도 안 낸다"는 증상은 **다른 원인들로는 남는다**: 바이너리 부재(`mcp_codex.go:1502`), codex 오류, 그리고 이 SPEC 이 새로 만드는 base 해석 불가(REQ-CRT-004). 셋 다 `applyGateUnmet`(`mcp_codex.go:1535-1545`)의 주석을 받되 verdict 는 `inconclusive` 이고, 위 기제에 의해 `disagreement_flag` 를 올리지 않는다.

⇒ SPEC 은 오도하지 않는다. **오도는 sync 단계에서 이슈에 "고쳤다"고만 답할 때 발생한다.** DoD 에 응답 문언 구속이 없다(D9, MINOR).

### Q4 — AC-CRT-004 의 미해석 케이스: **유예 자체는 정당하다. 다만 AC 가 관측 표면을 지정하지 않아 실질적으로 미지정 기준이다. (D4)**

**정당한 부분:** `spec.md:176` 이 반환 형태를 run 으로 미루지만 REQ-CRT-004 는 **금지 3종을 이미 고정**했다(review/start 미전송 · uncommittedChanges 미대체 · 원인 명명). plan §B(`plan.md:20-31`)는 후보 2개를 표로 제시하고 (가)를 권장하며 근거·대가까지 적었고, M1 에서 운영자 확인으로 고정한다고 배치했다. "결정하지 않았다"가 아니라 "결정을 마일스톤에 배치하고 그 전후 계약을 고정했다"이며, 되돌리기 어려운 결정을 앞에 두라는 원칙에도 맞는다.

**정당하지 않은 부분:** AC-CRT-004 의 Then 은 "결과는 base 를 해석할 수 없다는 원인을 **문자열로 명명**하며"(`acceptance.md:78`)라고만 쓴다. **어느 필드인지 말하지 않는다.** `ReviewOutput` 은 `Verdict` / `Summary` / `GateUnmet` 등 복수의 문자열 필드를 가지므로, 이 AC 는 반환 JSON 어디에든 "base" 를 포함한 문자열이 있으면 통과한다 — `acceptance.md §A` 가 "반환된 verdict 는 어떤 AC 의 근거도 되지 못한다"며 세운 관측 규율이 정작 이 AC 에서 가장 헐겁다.

**`applyGateUnmet` 과의 상호작용(리드 지적 확인):** (가)를 택하면 반환은 `inconclusiveReview(...)` 이고 `handleCodexAudit:1522` 의 `applyGateUnmet(out, root)` 를 통과한다. `applyGateUnmet` 은 `out.Verdict != VerdictInconclusive` 면 즉시 반환하므로(`:1536`) — (가)는 주석을 받고, **(나)(도구 오류)는 `toolErr` 로 조기 반환해 `applyGateUnmet` 을 아예 지나지 않는다.** 즉 required 게이트를 켠 프로젝트가 보는 화면이 두 후보에서 다르다. plan §B 의 대가 칸에도, `plan.md:31` 의 "(나)를 택하면 … AC-CRT-004 의 Then 첫 두 항목은 **문구 수정**"에도 이 차이가 없다 — (나)는 관측 표면 자체가 구조화 결과에서 Go 오류로 바뀌므로 "문구 수정"보다 크다.

**요구되는 수정:** AC-CRT-004 의 Then 을 후보별로 관측 필드까지 고정할 것. 예: (가)라면 "`Verdict == "inconclusive"` 이고 `Summary` 가 base 해석 불가를 명명하며 `"codex binary not found"` 와 서로 다르다"; (나)라면 "`res.IsError == true` 이고 오류 텍스트가 원인을 명명한다". plan §B 가 이미 "원인 문자열이 반드시 서로 구별 가능해야 한다"(`plan.md:29`)고 옳게 지적했으므로, AC 가 그 구별 가능성을 **단언**하기만 하면 된다.

---

## Defects Found

**D1. 카드의 [HARD] 회귀선이 근거 없이 스키마 계약으로 대체됨** — `acceptance.md:131` / `spec.md:33` — 카드는 "native 경로가 실제로 codex 에 도달했는지 단언"을 요구하나 어떤 AC 도 라이브 왕복을 관측하지 않는다. 불가능해서가 아니다: codex-cli 0.150.1 이 PATH 에 있고(`codex --version` rc=0), 라이브 `review/start` 를 실제로 보내는 선례가 `internal/cli/codex_live_protocol_probe_test.go:507,533` 에 있으며, skip 3조건 확립본이 `internal/cli/codex_review_gate_live_test.go:33-47` 에 있다. SPEC 3파일 어디에도 이 선례 인용이 없다. — **Severity: critical — Class: blocking** — Required fix: 위 Q1 의 `AC-CRT-010` 을 acceptance.md 에 추가하고 §D 매트릭스에 `RED: 예`로 등재, `plan.md` 에 마일스톤(권장: M2 RED 확립과 함께) 배치. §F 의 "별도 관측 항목으로 남긴다" 문장을 삭제하거나 "0.149.0 차이"만 남기도록 축소.

**D2. `git_strategy.worktree_base_branch` 우선이 GLM 경로와 해석 분기를 만드는데 기록되지 않음** — `spec.md:107`(REQ-CRT-003) vs `internal/cli/mcp_review_material.go:91-100` — REQ-CRT-003 은 설정 키를 1순위로 두지만 GLM 이 쓰는 `resolveReviewMergeBase` 는 설정 키를 **읽지 않고** `origin/HEAD → origin/main → main` 만 본다. 설정 키와 `origin/HEAD` 가 갈리는 트리에서는 `audit_multi(target=baseBranch)` 의 두 백엔드가 **서로 다른 변경분을 리뷰**하고 그 판정이 `mcp_convergence.go` 에서 하나로 수렴된다. §A.6 이 현행 비대칭을 결함으로 규정해 놓고 새 비대칭을 도입하면서 잔여 위험에도 없다. (이 트리에서는 `worktree_base_branch: develop` 과 `origin/HEAD → origin/develop` 이 우연히 일치해 증상이 보이지 않는다 — 일치는 측정했고, 일반화는 성립하지 않는다.) — **Severity: major — Class: blocking** — Required fix: 둘 중 하나를 **결정으로** 기록. (a) REQ-CRT-003 우선순위를 `resolveReviewMergeBase` 와 정렬하고 설정 키 우선을 포기하거나, (b) 분기를 의도된 것으로 선언하고 `spec.md §F` + `acceptance.md §F` 에 잔여 위험으로 명시.

**D3. AC-CRT-005 와 AC-CRT-006 이 commit/custom 행에서 모순** — `acceptance.md:82-88` vs `:90-96` — AC-CRT-005 는 `{"type":"commit"}` / `{"type":"custom"}` 이 **직렬화되지 않을 것**을 요구한다. AC-CRT-006 은 "`§B 표의 각 행`에 대해 그 variant 를 유발하는 입력이 주어지고 … 직렬화된 `target` 의 키 집합이 required 집합을 포함한다"를 요구하며 §B 표는 4행 전부를 담는다(`:20-25`). 두 variant 는 직렬화되지 않으므로 관측할 `target` 객체가 존재하지 않는다 ⇒ 그 두 행은 (i) 조용히 건너뛰어 **0매칭 초록**이 되거나(=`acceptance.md:31` 이 명시적으로 막으려는 사고) (ii) AC-CRT-005 와 직접 충돌한다. — **Severity: major — Class: blocking** — Required fix: AC-CRT-006 의 순회를 "직렬화되는 variant" 로 한정하고, 나머지 행에 대해서는 **부재를 단언**하도록 속성을 둘로 쪼갤 것(`serializable → required 포함` / `non-serializable → target 객체 미출현`). 그러면 다섯째 variant 가 늘어도 표만 느는 원래 의도가 보존된다.

**D4. AC-CRT-004 가 "원인 문자열"의 관측 필드를 지정하지 않음 + (나) 후보의 대가가 과소기술** — `acceptance.md:78` / `plan.md:31` — 반환 JSON 어디에든 원인처럼 보이는 문자열이 있으면 통과한다. 또한 (나)를 택하면 `toolErr` 조기 반환으로 `applyGateUnmet`(`internal/cli/mcp_codex.go:1535`)을 지나지 않아 required 게이트 소비자가 보는 것이 달라지는데, plan 은 이를 "문구 수정"으로 적었다. — **Severity: major — Class: blocking** — Required fix: Q4 에 적은 대로 후보별 관측 필드를 AC 에 고정하고, `plan.md §B` 의 (나) 대가 칸에 `applyGateUnmet` 미경유를 명시.

**D5. REQ-CRT-003 의 3·4 단계가 이름 층위에서 구별 불가이고, 반환하는 이름의 수용 가능성이 미측정** — `spec.md:107` / `plan.md:42-49` — plan §C 3단계("`origin/main` 이 존재하면 `main`")와 4단계("로컬 `main` 이 존재하면 `main`")는 **같은 문자열**을 낸다. 어떤 AC 도 둘을 구별할 수 없다(AC-CRT-002 는 1·2 단계만 잰다). 더 무거운 문제: 2단계는 `origin/develop` 에서 접두사를 떼 `develop` 을, 3단계는 `origin/main` 존재를 근거로 `main` 을 반환하는데, **그 이름이 로컬에 존재한다는 확인은 하지 않는다** — plan §C 의 "존재 확인 없이 이름을 반환하지 않는다" 주의(`plan.md:49`)가 자기 사슬에 의해 위반된다. codex 가 `baseBranch` 값을 어떻게 해석하는지(로컬 브랜치 / 임의 revision)는 스키마 설명("Review changes between the current branch and the given base branch")에 없고 이 트리에서 측정되지 않았다. — **Severity: major — Class: blocking** — Required fix: 3·4 단계를 하나로 합치거나, 각 단계가 반환하는 **이름 자체의 해석 가능성**을 확인하도록 §C 를 고칠 것. codex 의 수용 여부는 D1 의 `AC-CRT-010` 이 함께 잰다.

**D6. AC-CRT-009 를 뒷받침하는 REQ 가 없고 AC-CRT-008 의 매핑이 부정확** — `acceptance.md:46-47,104-116` — AC-CRT-009(도구 표면 서술)는 REQ-CRT-002 로 매핑돼 있으나 REQ-CRT-002 는 요청 조립 요구이지 도구 설명 요구가 아니다. 어떤 REQ 도 "도구 표면이 서버 해석을 서술한다"를 요구하지 않는다. AC-CRT-008(RED 증거 보존)도 REQ-CRT-002 의 행동 검증이 아니라 `§C` 프로세스 규율이다. — **Severity: minor — Class: blocking** — Required fix: 도구 표면 요구를 REQ-CRT-007 로 신설하거나 AC-CRT-009 를 REQ-CRT-002 의 하위 관측으로 재서술. AC-CRT-008 은 매핑 칸을 `§C 규율` 로 표기.

**D7. 좌표 관례 불일치 2건** — (a) `spec.md:78-81` §A.4 표에서 `codex_review_rpc_test.go:62` / `:94` 는 doc comment 줄이고 `func` 은 `:63` / `:96`; 같은 표의 `mcp_codex_test.go:183` 은 `func` 줄, `:231` 은 인자 줄 — 한 표 안에서 관례가 셋. (b) `spec.md:145` 가 "그 카드가 t284 다"의 "근거(측정)"로 `succession.md` 를 들지만 그 문서는 t284 를 명명하지 않는다(연결은 큐 텍스트에서 옴 — discovery.md Gaps 가 이를 이미 인정). — **Severity: minor — Class: optional** — Required fix: (a) 심볼 줄로 통일. (b) 출처를 `succession.md`(축의 존재·처분) + `moai todo`(카드 번호)로 분리 표기.

**D8. REQ-CRT-004 가 legacy EARS `If` 형인데 deprecation 주석이 없다** — `spec.md:111` — GEARS canonical Unwanted 는 `shall not`. backward-compat 창(2026-11-22) 안이라 MP-2 는 통과. — **Severity: minor — Class: optional** — Required fix: `[DEPRECATED — use shall not]` 주석을 달거나 `shall not` 형으로 재작성(금지 2 + 양성 1 이 한 REQ 에 묶여 있으므로 분할해도 좋다).

**D9. DoD 에 #1632 응답 문언 구속이 없다** — `acceptance.md:120-126` — Q3 참조. 착지 후에도 "required native gate 가 조용히 아무것도 안 낸다"는 증상이 다른 원인들로 남는데, 이슈 응답이 이를 말하지 않으면 SPEC 이 아니라 응답이 오도한다. — **Severity: minor — Class: optional** — Required fix: DoD 6항 추가 — "#1632 응답이 1축 닫힘 / 2축 t284 / 제보 #3 정책 질문의 세 상태를 구별해 말한다".

**D10. REQ-CRT-006 의 "byte-identical in shape" 가 자기모순이고 기준 시점이 고정되지 않았다** — `spec.md:119` — 바이트 동일과 형태 동일은 다른 강도이고, AC-CRT-003 은 형태만 단언한다(`acceptance.md:70`). "pre-change form" 의 기준 트리도 이 REQ 안에 고정돼 있지 않다(SPEC 다른 곳의 `442da4f06` 로 복원 가능하지만 REQ 자체는 움직이는 참조). — **Severity: minor — Class: optional** — Required fix: "shall remain shape-identical to its form at `442da4f06`" 로 문언과 기준 SHA 를 함께 고정.

---

## 확인했고 결함이 아닌 것 (감사자가 의심했다가 기각한 항목)

- **Tier M 분류** — 정당하다. `spec-workflow.md:148-150` 의 Tier S AC 상한은 8, 이 SPEC 은 AC 9건 ⇒ 초과 ⇒ tier-up. `progress.md:8` 이 "LOC·파일 수만 보면 S 로도 읽힌다"고 스스로 적어 근거를 예산 축에 정확히 한정했다. 임계도 그에 맞춰 0.80 을 적용했다.
- **`commit` / `custom` 을 enum 에 노출하지 않는 결정**(`spec.md:149-153`) — 편의가 아니라 근거 있는 결정이다. 제보된 결함은 기능 부재가 아니라 잘못된 요청이고, 값 추가는 `sha` / `instructions` 파라미터까지 끌어오는 확장이다. 카드의 "baseBranch 로 조용히 좁히지 말라"는 지시는 REQ-CRT-005 의 안전 축으로 답해져 있다. 다만 그 REQ 에 **양성 귀결이 없다**는 점은 D3 로 별도 계상했다 — 노출 거부는 옳고, 거부했을 때 무엇이 일어나는지가 비어 있다.
- **§A.1 스키마 표** — `.moai/reports/t399/schema/v2/ReviewStartParams.json` 을 직접 파싱해 4행 전부 대조: `uncommittedChanges [type]` / `baseBranch [branch,type]` / `commit [sha,type]` / `custom [instructions,type]`. **완전 일치.** 최상위 `required: [target, threadId]` 도 확인했고 `buildCodexReviewParams` 가 `threadId` 를 항상 싣는다(`mcp_codex.go:943`).
- **plan §C 배치 판단** — 옳다. `handleCodexAudit:1508-1512` 가 `params["cwd"] = root` 를 넘기고 그 map 이 `buildCodexReviewParams` 까지 도달하므로, 해석에 필요한 루트가 조립부에 이미 닿아 있다. "새 배관이 필요한지 먼저 읽는다"(`plan.md:71`)는 hedge 는 실제로 불필요한 배관을 막았다.
- **§F 의 `codex_review_gate.go:90` 주장** — 확인. `runCodexReviewRPC(..., map[string]any{"target": codexTargetUncommitted, "cwd": projectDir})`(`:89-93`). 하드코딩이 맞고 이 변경의 영향을 받지 않으며 REQ-CRT-006 회귀선이 그 경로를 함께 지킨다.

---

## Recommendation

이 SPEC 은 측정 품질이 높다 — §A.1 스키마 표는 원문과 4행 전부 일치하고, §A.4 의 전제 정정은 옳으며 그 함의(공허한 초록)까지 정확하다. §E 의 t284 분리도 코드 위치·카드 본문 양쪽에서 성립한다. FAIL 은 이 SPEC 이 나쁘다는 판정이 아니라, **자기가 세운 규율을 자기 검증 표면에는 적용하지 않았다**는 결함이 중심에 있다는 판정이다.

수정 순서(반영 후 iteration 2 는 아래 델타만 재감사한다):

1. **D1** — `AC-CRT-010`(라이브 baseBranch 왕복) 추가 + `plan.md` 마일스톤 배치 + `acceptance.md §F` 축소. 저장소 안의 `codex_live_protocol_probe_test.go:507` / `codex_review_gate_live_test.go:33` 를 §G 참조에 넣어, 이것이 새 발명이 아니라 재사용임을 문서에 남길 것.
2. **D3** — AC-CRT-006 을 `serializable → required 포함` / `non-serializable → 부재` 두 속성으로 분할.
3. **D4** — AC-CRT-004 Then 을 후보별 관측 필드로 고정 + `plan.md §B` (나) 대가에 `applyGateUnmet` 미경유 추가.
4. **D5** — plan §C 사슬의 3·4 단계 병합 또는 이름 해석 가능성 확인 추가.
5. **D2** — GLM 경로와의 해석 분기에 대해 (a) 정렬 또는 (b) 잔여 위험 명시 중 하나를 **결정으로** 기록.
6. **D6** — 도구 표면 요구를 REQ 로 승격하거나 AC-009 를 재서술; AC-008 매핑 정정.
7. D7 · D8 · D9 · D10 — optional. 운영자 재량.

D2 와 D5 는 운영자 결정을 요구할 수 있다(D5 의 codex 이름 해석 규칙은 D1 의 라이브 AC 가 함께 답하므로 D1 을 먼저 두었다).

---

## Evidence / Gaps / Residual-risk

**Evidence** — 이 감사가 실제로 실행한 것: 트리·HEAD·브랜치 확인(`442da4f06`, `WT-codex-native-branch`) · SPEC 4파일 전문 판독 · `ReviewStartParams.json` python 파싱(4 variant required 집합 대조) · `codex --version`(0.150.1, rc=0) · `grep -rn 'codexTargetBaseBranch|"baseBranch"' --include=*.go internal/` · `grep -n '^func Test'` 2파일 · `mcp_codex.go` 985-1015 / 930-960 / 1470-1570 판독 · `mcp_convergence.go` 160-230 판독 · `mcp_review_material.go` 55-130 판독 · `loader_worktree_base.go` 판독 · `codex_review_gate.go` 84-96 판독 · `codex_live_protocol_probe_test.go` 498-560 판독 · `codex_review_gate_live_test.go` 1-80 판독 · `origin/HEAD` 판독 → `origin/develop` · `.moai/config/sections/git-strategy.yaml` 판독 → `worktree_base_branch: develop` · `moai todo` 에서 t284 / t399 카드 본문 판독 · `grep '^status:' SPEC-CODEX-VERDICT-SYNTH-001/spec.md` → completed · `grep -rn 'NEEDS CLARIFICATION'` → 0 · `grep -c syscall` → 0 · `spec-workflow.md:136-158` Tier 표 판독.

**Baseline-attribution** — 전부 `.claude/worktrees/t399` @ `442da4f06`, 이번 실행. primary 체크아웃은 읽지 않았다.

**Gaps (관측하지 않은 것)** — (1) 어떤 테스트도 **실행하지 않았다**; 기존 검사가 현재 초록인지는 미관측이며 AC-CRT-003 의 "변경 전 초록" 전제도 미확인이다. (2) 라이브 codex 왕복을 이 감사에서 **시도하지 않았다** — D1 의 RED 예측(`{"type":"baseBranch"}` 거절)은 스키마 판독에 근거한 추론이지 관측이 아니다. (3) codex 가 `baseBranch` 값으로 로컬 미존재 이름을 수용하는지 미측정(D5). (4) codex-cli 0.149.0 과의 스키마 차이 미관측. (5) `resolveToolProjectRoot` 의 폴백 동작을 읽지 않아 AC-CRT-004 픽스처 구성 가능성은 추론이다.

**Residual-risk** — Tier M 임계 0.80 대비 0.75 는 근접 구간이고, 차원 점수 중 Testability(0.70)는 밴드 사이 판단이다. D3 를 "구현 시 자연히 해소될 표현 문제"로 읽는 견해가 가능하며 그 경우 총점이 0.80 에 닿을 수 있다. 다만 D1 은 그 해석에 영향받지 않는다 — 카드의 [HARD] 문언과 저장소 안 선례의 존재가 둘 다 관측된 사실이므로, D1 하나만으로도 이번 라운드의 재작업은 정당하다.

---

# SPEC Review Report: SPEC-CODEX-REVIEW-TARGET-001 — Iteration 2

- Iteration: 2/2 (Tier M ceiling)
- **Verdict: PASS-WITH-DEBT**
- **Overall Score: 0.81** (iter1 0.75 → iter2 0.81, Tier M 임계 0.80 — 회귀 없음, LEAN STOP escalation 조건 미해당)
- 감사 대상: `spec.md` v0.2.0 · `plan.md` v0.2.0 · `acceptance.md` v0.2.0, 트리 `442da4f06`
- 범위: **iter1 결함 D1-D10 의 델타 재감사 + 회귀 검사.** 전면 재감사가 아니다(Retry Loop Contract).
- Reasoning context ignored per M1 Context Isolation.

**PASS-WITH-DEBT 의 조건:** 아래 **D2-RESIDUAL 이 M2a 착수 전에 해소되는 것**이 이 판정의 전제다. 그 편집이 없으면 이 판정은 성립하지 않는다 — 이유는 해당 항목에 적었다.

---

## §1 Regression Check — iter1 결함 10건의 처분

| # | iter1 severity | 처분 | 근거 (관측) |
|---|---|---|---|
| D1 | critical | **RESOLVED** (파생 결함 N1 있음) | AC-CRT-010 신설(`acceptance.md:139-156`), §D 매트릭스 등재(`:53`), §C 에 RED 종류 구분(`:33`), §G 선례 블록(`spec.md:233-238`), §0 이 방어선을 2층으로 재서술(`:34-39`), plan M2b(`:63-74`) + 자기검증 행(`:100`) + 안티패턴 7·8(`:118-119`), DoD 1·2 항이 skip 을 **미관측**으로 처리(`acceptance.md:162-163`). 내가 지정한 것보다 넓게 닫았다. |
| D2 | major | **PARTIAL — 결정은 착지, 전파는 미착지** | §A.7 신설(`spec.md:101-116`), REQ-CRT-003 재작성(`:130-134`), §F 2건(`:218-222`), plan §C:39 + 안티패턴 9. **그러나 `acceptance.md` AC-CRT-002(`:63-70`)가 철회된 설계를 그대로 요구한다.** → D2-RESIDUAL |
| D3 | major | **RESOLVED** | AC-CRT-006(`:103-107`, 직렬화 가능 행만) / AC-CRT-006b(`:109-113`, 부재 단언)로 분할. §B 표에 `분류` 열 추가(`:20-27`)로 순회 대상을 데이터가 가른다. 내가 요구하지 않은 행-수 계수 가드까지 추가(`:117`). 파생 결함 N2 있음. |
| D4 | major | **RESOLVED** | AC-CRT-004 에 후보별 관측 필드 표(`:88-91`) — (가) `Verdict`+`Summary`+구별 가능성, (나) `res.IsError`+오류 텍스트. "반환 JSON 어딘가" 판정 금지 명문화(`:86`). plan §B (나) 대가 칸에 `applyGateUnmet` **조기 반환** 명시(`:25` ②) + `:27` 권장 근거에 반영 + `:31` 이 "문구 수정" 표현을 폐기. |
| D5 | major | **RESOLVED** | plan §C 사슬 2단계로 병합(`:41-45`) + 병합 근거를 "구별할 수 없는 두 단계는 어떤 AC 도 가를 수 없다"로 명시. ref 해석 가능성 [HARD](`:49`) + REQ-CRT-003 본문에 편입(`spec.md:132-134`) + 안티패턴 5. `spec.md §F:221` 이 codex 쪽 해석 규칙 미관측을 제약으로 기록. |
| D6 | minor | **RESOLVED** | REQ-CRT-007 신설(`spec.md:150-154`), AC-CRT-009 → REQ-CRT-007 재매핑(`acceptance.md:52`), AC-CRT-008 → `§C 규율` 로 정정(`:51`, "특정 REQ 의 행동 검증이 아니다" 명시). |
| D7 | minor | **RESOLVED** | (a) §A.4 좌표를 `func` 선언 줄로 통일(`:63` / `:96` / `:183` / `:216`)하고 관례를 표 위에 선언(`spec.md:82`); `:231` 은 인자 줄임을 본문에 표기. **네 좌표 전부 트리에서 재확인, 일치.** (b) §E 출처를 succession.md(축·처분)와 큐(카드 번호)로 분리(`:186-188`)하고 "이 문서는 카드 번호를 담고 있지 않다"를 명문화. §G:230 도 정정. |
| D8 | minor | **PARTIAL** | canonical `shall not` 은 착지(`spec.md:138`). 그러나 modifier 가 `Where` 로 바뀌었다 → 새 패턴 오용. D8-RESIDUAL |
| D9 | minor | **RESOLVED** | DoD 6항 신설(`acceptance.md:167-169`) — (a) 닫힌 것 (b) 열린 채 남는 것(원인 3종 열거 + t284 귀속) (c) 제보 #3 정책 질문. 구속을 SPEC 이 아니라 응답에 거는 이유까지 적었다. |
| D10 | minor | **RESOLVED** | REQ-CRT-006 을 "shape-identical to its form at `442da4f06`"(`spec.md:146`)로 재작성 + 근거 2문장(`:148`) — 움직이는 참조 문제와 바이트 동일이 회귀선을 거짓으로 붉게 만드는 문제 양쪽. |

**10건 중 8건 완전 해소, 2건 부분 해소.** 파생 결함 2건(N1 · N2)이 새로 발생했다.

---

## §2 Must-Pass Results (재확인)

- **[PASS] MP-1** — `REQ-CRT-001..008` 연속, 결번·중복 없음(`spec.md:122,126,130,136,140,144,150,156`). AC 는 11건(`001..010` + `006b`); `006b` 는 짝 표기이며 매트릭스가 그 관계를 명시한다 — 결번이 아니다(N3 로 minor 계상).
- **[PASS] MP-2** — 8건 전부 GEARS 패턴. 신규 2건: REQ-CRT-007 / 008 모두 Ubiquitous(`shall ...`) ✓. REQ-CRT-004 의 modifier 오용은 D8-RESIDUAL(MINOR)이며 canonical `shall not` 은 유지되므로 MP-2 는 통과. 판정 층위는 iter1 과 동일하게 requirement layer.
- **[PASS] MP-3** — 프론트매터 12 필드 불변, `version: "0.2.0"` 로만 갱신(`spec.md:4`).
- **[N/A] MP-4** — 단일 프로그래밍 언어 범위, 변동 없음.
- **[PASS] MP-5 D7** — 참조 SPEC 집합 불변(`SPEC-CODEX-VERDICT-SYNTH-001` 단 하나, `status: completed`). 신규 참조 없음.
- **[PASS] MP-6 D8** — `grep -c syscall spec.md` → `0`.
- **[PASS] MP-7** — `grep -rn '\[NEEDS CLARIFICATION'` → 0 hit (rc=1).

예산: REQ 8 ≤ 16, AC 11 ≤ 16 (Tier M 상한). 초과 없음.

---

## §3 Category Scores (iter1 → iter2)

| Dimension | iter1 | iter2 | Band | 이동 근거 |
|---|---|---|---|---|
| Clarity | 0.75 | **0.75** | 0.75 | REQ-006 자기모순 해소, REQ-004 를 금지/양성으로 분리, REQ-008 신설, §A.7 이 결정과 근거를 분리해 서술 — 전부 개선. 그러나 REQ-CRT-003 ↔ AC-CRT-002 의 정면 충돌(D2-RESIDUAL)이 발생했고, 그것은 iter1 에 없던 종류의 불명확이다. 상쇄되어 제자리. |
| Completeness | 0.75 | **0.85** | 0.75↔1.0 (1.0 쪽) | §A.7(결정 기록) · §F 제약 4건(codex 해석 규칙 미관측 포함) · §G 선례 블록 · DoD 6항 · §F 잔여위험 3건이 모두 추가됐다. 남은 결손은 REQ-CRT-005 의 양성 귀결 부재(N2) 하나. |
| Testability | 0.75 | **0.75** | 0.75 | AC-006/006b 분할 + 행-수 가드, AC-004 필드 고정 표, AC-010 의 양성 사실 판정과 skip≠통과 — 큰 개선. 그러나 AC-CRT-002 가 현 REQ 로는 구현 불가하고(D2-RESIDUAL), AC-006b 의 부재 단언이 조용한 대체를 통과시킨다(N2). |
| Traceability | 0.80 | **0.90** | 0.75↔1.0 (1.0 쪽) | REQ 8건 전부 ≥1 AC(001→006·006b·007 / 002→001·010 / 003→002 / 004→004 / 005→005·006b / 006→003 / 007→009 / 008→004). 존재하지 않는 REQ 를 참조하는 AC 없음. AC-008 의 오매핑과 AC-009 의 근거 REQ 부재가 둘 다 해소됐다. |

산술평균 (0.75 + 0.85 + 0.75 + 0.90) / 4 = **0.81** ≥ 0.80.

**점수 회귀 없음** (0.75 → 0.81). LEAN STOP escalation 조건(iter(N+1) < iter(N)) 미해당.

---

## §4 리드가 지정한 네 질문

### Q1 — AC-CRT-010 은 양성 사실을 관측하는가, 그리고 RED 주장은 예측인가 관측인가

**관측 대상은 옳다.** `acceptance.md:143` 의 Then 은 "응답이 JSON-RPC error 가 **아니고** … 그 turn 이 `turn/started` 에 **도달한다**" — 두 절 중 후자가 양성 사실이고, `:145` 가 "판정은 거절 부재라는 양성 사실이다. `inconclusive` 여부는 보지 않는다"로 못 박는다. 실패 부재만으로 통과시키지 않는 구조가 맞다. skip 처리도 정확하다(`:156` [HARD] skip ≠ 통과, DoD 1항, plan 안티패턴 7).

**RED 주장은 과잉이다.** 세 곳이 이를 **이미 관측된 사실**로 적는다.

- `acceptance.md:53` §D 매트릭스 — "예 — 현행 트리에서 **실물로 붉다**"
- `acceptance.md:150` — "…거절하므로, 이 SPEC 이 확보할 수 있는 RED 중 가장 강하다 — 스키마 판독에서 유도한 것이 아니라 **실물에서 관측한 실패다**"
- `plan.md:67` M2b — "(스키마 판독에서 유도한 예측이 아니라 실 codex 가 요청을 거절하는 **관측**)"

아무도 이 왕복을 돌리지 않았다. 내 iter1 Gaps 가 명시적으로 기록한 바다: "라이브 codex 왕복을 이 감사에서 시도하지 않았다 — D1 의 RED 예측은 스키마 판독에 근거한 추론이지 관측이 아니다." `acceptance.md:150` 은 한 문장 안에서 자기를 반박한다 — 근거절("0.150.1 이 `{"type":"baseBranch"}` 를 required field `branch` 누락으로 거절하므로")이 **바로 그 스키마 판독**인데, 결론절이 "스키마 판독에서 유도한 것이 아니다"라고 말한다.

SPEC 자신의 규칙과도 충돌한다 — `plan.md:117` 안티패턴 6: "**RED 를 서술로 대체하기.** '이 검사는 현행 구현에서 실패할 것이다'는 예측이지 관측이 아니다."

`acceptance.md:33` 은 정확히 옳게 썼다("그 거절 응답(JSON-RPC error 본문)을 그대로 남긴다" — run-phase 의무). 문제는 그 의무를 이미 이행한 것처럼 적은 세 곳이다. → **N1 (major)**

### Q2 — D2 정렬이 새 문제를 만드는가, §A.7 은 정당화인가 기록인가

**§A.7 은 기록이 아니라 정당화다.** 후보 2개를 결과로 대비시키고(`spec.md:105-108`), 결정을 명시하고(`:110`), 근거를 원리로 진술한다 — "설정 키가 옳다면 그것은 **두 백엔드 모두에게** 옳다. 한쪽만 읽게 만드는 것은 이 SPEC 이 진단한 것과 같은 종류의 비대칭을 반대 방향으로 하나 더 만드는 일"(`:112`). §A.6 이 세운 진단과 정합한다.

특히 좋은 두 가지: (1) 이 트리에서 두 값이 우연히 일치한다는 사실을 **어느 쪽 근거도 아니라고 스스로 배제한다**(`:114`) — iter1 에서 지적한 바로 그 함정을 명시적으로 피했다. (2) 뒤집는 경로를 "GLM 경로까지 함께 바꾸는 별도 카드"로 고정한다(`:116`, `§F:222`, plan 안티패턴 9) — 되돌릴 수 있음을 보이면서 되돌리는 방법을 한쪽만 바꾸지 못하게 묶었다.

**"사용자가 의도적으로 설정한 값을 버리는가"라는 새 문제는 실재하나, 잘못된 축이다.** 이 SPEC 이 그 키를 무시해도 **GLM 이 원래 무시하고 있었으므로** 사용자 관점의 동작은 변하지 않는다 — `baseBranch` 리뷰는 이전에도 `origin/HEAD` 기준이었다. 즉 채택안은 현상 유지이고, 기각안이 새 동작이었다. `worktree_base_branch` 가 존중되어야 한다는 주장은 옳을 수 있지만 그것은 **두 백엔드 공통의 미구현**이지 이 카드가 만든 결손이 아니다. §F:222 가 정확히 그렇게 적었다. 새 결함으로 계상하지 않는다.

**그러나 정렬이 acceptance 층에 전파되지 않았다.** → **D2-RESIDUAL (critical)**

### Q3 — AC-CRT-006b 는 0매칭 위험을 닫는가, 부재 단언에서 재생산하는가

**닫는 쪽이 하나, 열린 쪽이 하나.**

닫은 것 — **순회 행 수의 0매칭**: `:117` [HARD] "두 AC 모두 순회한 행의 수를 먼저 세고 판정한다. 분류 결과가 빈 집합이면 그 AC 는 아무것도 단언하지 않은 것이며, **통과가 아니라 결함으로 보고한다**." 내가 요구하지 않은 가드이고, `§B` 표의 `분류` 열이 데이터로 순회 대상을 가르므로 기계적으로 셀 수 있다. 셀렉터 0매칭 계열의 사고는 이것으로 막힌다.

열린 것 — **부재 단언의 만족 방식**: Then 이 "**그 variant 의** `target` 객체는 출현하지 않는다"(`:113`)로 variant 한정이다. 따라서 구현이 `commit` 입력에 대해 `{"type":"uncommittedChanges"}` 를 **대신 내보내도** 이 AC 는 통과한다 — `commit` 의 target 객체는 출현하지 않았으므로. 이는 AC-CRT-004 가 `baseBranch` 에 대해 명시적으로 닫은 조용한 대체(`:84` "…`target.type == "uncommittedChanges"` 인 요청도 대신 전송되지 않으며")를 `commit`/`custom` 에 대해서는 열어 두는 것이다.

뿌리는 REQ-CRT-005 다: 금지만 있고 양성 귀결이 없다(iter1 Clarity 항목에서 지적했으나 결함으로는 계상하지 않았던 것). 현행 구현의 default 분기가 이미 `uncommittedChanges` 로 떨어지므로(`internal/cli/mcp_codex.go:1004`), 이 구멍은 가설이 아니라 **현행 코드가 실제로 하는 동작**이다 — AC-006b 는 그 동작에 대해 붉지 않다. §D 매트릭스가 AC-006b 를 "RED 로 시작: 예"(`:49`)로 등재한 것과 어긋날 소지가 있다. → **N2 (major)**

### Q4 — REQ 6 → 8: 번호 연속성 · 신규 2건의 GEARS · AC 매핑

- **번호 연속성**: `REQ-CRT-001..008` 연속, 결번 0, 중복 0, 제로패딩 일관(3자리). ✓
- **REQ-CRT-007** (`spec.md:152`): "The `codex_audit` tool description **shall state** that … and **shall name** the resolution source." — Ubiquitous ✓. **제 자리를 번다**: iter1 D6 이 지적한 "근거 REQ 없는 AC-009" 를 정확히 메우고, `:154` 가 왜 필요한지를 §A.2 에 연결한다.
- **REQ-CRT-008** (`spec.md:158`): "The codex audit **shall report** an unresolvable base branch **in a named output field**, distinguishable from every other fail-open cause." — Ubiquitous ✓. **제 자리를 번다**: REQ-004 에서 양성 의무를 떼어내 독립시킨 결과이고(`:160` 이 그 분담을 명시), D4 가 요구한 "관측 필드 고정"의 요구 층 근거가 된다. 이 분리 덕에 REQ-004 는 순수 unwanted 가 됐다.
- **AC → 존재하지 않는 REQ 매핑**: 없음. 11개 AC 의 매핑 대상은 `REQ-CRT-001/002/003/004/005/006/007/008` 및 `§C 규율` 하나뿐이며 전부 실재한다. ✓
- **AC-CRT-010 → REQ-CRT-002** 는 처음에 편의 매핑을 의심했으나 **기각한다**: codex 는 정확히 `branch` 누락일 때 거절하므로, 라이브 비거절은 "non-empty `branch` 필드를 실었다"(REQ-002)의 종단 검증이다. 결함 아님.
- **REQ-CRT-004 의 modifier**: `Where` 는 GEARS 에서 capability gate / feature flag / static config 를 가리킨다. "base 브랜치를 해석할 수 없다"는 런타임 상태이지 capability gate 가 아니므로 `When`(event-driven) 또는 `While`(state-driven)이 맞다. iter1 의 legacy `If` 를 고치면서 modifier 를 잘못 갈아끼웠다. → **D8-RESIDUAL (minor)**

---

## §5 Defects Found (iteration 2)

**D2-RESIDUAL. §A.7 의 철회 결정이 acceptance 층에 전파되지 않아, SPEC 이 서로 모순되는 두 동작을 요구한다** — `acceptance.md:63-70` vs `spec.md:130-134` — REQ-CRT-003 은 "`resolveReviewMergeBase` 와 같은 사슬"만 쓰고 `worktree_base_branch` 를 읽지 않는다고 규정하고, `spec.md §F:222` 는 "그 키는 이 경로에서 읽히지 않는다"를 제약으로 못 박으며, `plan.md` 안티패턴 9 는 그 키를 codex 경로에 배선하는 것을 금지한다. **그런데 AC-CRT-002 는 그 키를 요구한다**: Given "`git_strategy.worktree_base_branch` 가 설정된 프로젝트 루트"(`:65`), Then "전송된 `target.branch` 는 그 **설정값과 같다**"(`:67`), And-절도 그 키의 부재를 조건으로 둔다(`:69-70`). §D 매트릭스는 이 AC 를 여전히 REQ-CRT-003 에 매핑한다(`:44`).

전파 누락임이 확인된다: 같은 개정에서 AC-CRT-004 의 Given 은 "설정 키 부재 + …"에서 설정 키 절이 **제거됐다**(`:82`). 즉 편집이 AC-004 에는 닿았고 AC-002 에는 닿지 않았다.

이것이 critical 인 이유는 **run-phase 가 AC 로부터 검사를 쓰기 때문**이다. M2a 가 AC-CRT-002 를 문면대로 구현하면 `LoadWorktreeBaseBranch` 를 1순위로 읽는 해석기를 요구하는 RED 가 서고, M3 가 그것을 초록으로 만들면 §A.7 이 기각한 설계가 정확히 착지한다. HISTORY(`spec.md:23`)의 "**D2** — `worktree_base_branch` 우선을 **철회**하고 … 정렬" 이라는 주장은 산출물 집합 전체에 대해서는 참이 아니다. — **Severity: critical — Class: blocking** — Required fix: AC-CRT-002 를 정렬된 사슬로 재작성. 예: Given `origin/HEAD` 가 해석되는 루트 → Then `target.branch` 는 접두사를 뗀 그 이름; And Given `origin/HEAD` 부재이고 `main` 이 ref 로 해석되는 루트 → Then `main`; And 각 단계가 반환 전에 ref 해석 가능성을 확인한다(REQ-CRT-003 후반절). 제목("해석 우선순위가 관측된다")은 그대로 유효하다.

**N1. AC-CRT-010 의 RED 가 관측된 것으로 서술됐다 — 실제로는 예측이다** — `acceptance.md:53`, `acceptance.md:150`, `plan.md:67` — 이 트리에서 라이브 왕복은 **한 번도 실행되지 않았다**(iter1 Gaps 에 기록). 세 곳 모두 "실물로 붉다" / "실물에서 관측한 실패다" / "예측이 아니라 관측"으로 완료형 서술을 쓴다. `acceptance.md:150` 은 근거절이 스키마 판독인데 결론절이 스키마 판독이 아니라고 말해 한 문장 안에서 모순된다. SPEC 자신의 `plan.md:117` 안티패턴 6("RED 를 서술로 대체하기 — 예측이지 관측이 아니다")과도 충돌한다. 위험은 실질적이다: 이미 관측된 것으로 적힌 RED 는 run-phase 가 `acceptance.md:33` 의 보존 의무를 이행했다고 착각할 근거가 된다. — **Severity: major — Class: blocking** — Required fix: 세 곳을 예측형으로 되돌릴 것. §D 매트릭스 `예 — 현행 트리에서 실물로 붉다` → `예 — 스키마상 거절이 예측되며, M2b 가 그 거절을 관측해 §C 규율대로 기록한다`. `:150` / `plan.md:67` 도 동일하게. `acceptance.md:33` 은 이미 올바른 문언이므로 손대지 않는다.

**N2. AC-CRT-006b 의 부재 단언이 조용한 대체를 통과시킨다** — `acceptance.md:113` — Then 이 "**그 variant 의** `target` 객체는 출현하지 않는다"로 variant 한정이라, `commit` 입력에 대해 `{"type":"uncommittedChanges"}` 를 대신 내보내는 구현이 통과한다. 이는 AC-CRT-004 가 `baseBranch` 에 대해 명시적으로 닫은 구멍(`:84`)을 `commit`/`custom` 에 대해 열어 두는 것이다. 현행 코드가 이미 그렇게 동작하므로(`internal/cli/mcp_codex.go:1004` default 분기가 `uncommittedChanges` 반환) 가설이 아니라 실제 동작이고, §D 매트릭스가 AC-006b 를 "RED 로 시작: 예"(`:49`)로 등재한 것과 어긋날 소지가 있다. 뿌리는 REQ-CRT-005 가 금지만 있고 양성 귀결이 없다는 점이다. — **Severity: major — Class: blocking** — Required fix: AC-CRT-006b 의 Then 에 대체 부재를 추가 — "그 variant 의 `target` 객체가 출현하지 않으며, **다른 variant 의 `target` 을 실은 `review/start` 가 대신 전송되지도 않는다**". REQ-CRT-005 에 양성 귀결 한 절을 붙이는 것이 더 근본적이나(무엇이 일어나야 하는지), AC 층 보강만으로도 이 구멍은 닫힌다.

**D8-RESIDUAL. REQ-CRT-004 의 GEARS modifier 가 `Where` 로 잘못 바뀌었다** — `spec.md:138` — canonical `shall not` 은 착지했으나(D8 의 요구), modifier 가 legacy `If` 에서 `Where` 로 갔다. GEARS 에서 `Where` 는 capability gate / feature flag / static config 를 가리키며, "base 브랜치를 해석할 수 없다"는 런타임 상태다. `When`(event-driven) 또는 `While`(state-driven)이 맞다. — **Severity: minor — Class: optional** — Required fix: `Where` → `When`.

**N3. AC 식별자에 접미사 형태(`006b`)가 섞였다** — `acceptance.md:109` — `AC-CRT-001..010` 중 유일하게 접미사를 쓴다. 매트릭스가 짝 관계를 명시하고(`:49`) 본문이 "두 AC 는 한 쌍"(`:115`)이라 읽는 데 지장은 없으므로 MP-1 위반으로 계상하지 않았다. 다만 식별자 형식 일관성 축에서는 이탈이다. — **Severity: minor — Class: optional** — Required fix: 선택. `AC-CRT-011` 로 재번호하되 본문의 짝 서술을 유지하거나, 현행 유지.

---

## §6 Recommendation

iter1 대비 개선 폭이 크다. 특히 D1 은 내가 지정한 범위보다 넓게 닫혔고(방어선을 2층으로 재서술 + skip≠통과 + DoD 미관측 처리 + 안티패턴 2건), D5 의 사슬 병합 근거("구별할 수 없는 두 단계는 어떤 AC 도 가를 수 없다")와 §A.7 의 결정 서술은 이 SPEC 에서 가장 잘 쓰인 부분이다. 점수도 0.75 → 0.81 로 임계를 넘겼고 회귀는 없다.

그럼에도 **PASS 가 아니라 PASS-WITH-DEBT** 인 이유는 하나다: D2-RESIDUAL 이 남은 상태에서 M2a 가 착수되면 §A.7 이 기각한 설계가 그대로 착지한다. 이것은 미룰 수 있는 부채가 아니라 **한 AC 의 재작성**으로 끝나는 편집이며, 그 편집이 없으면 이 판정은 성립하지 않는다.

**run-phase 진입 전 반드시 해소 (blocking):**

1. **D2-RESIDUAL** — AC-CRT-002 를 정렬된 사슬로 재작성. iter2 의 다른 어떤 항목보다 먼저.
2. **N1** — RED 완료형 서술 3곳을 예측형으로. 이것을 두면 M2b 가 관측 없이 이행됐다고 기록될 수 있다.
3. **N2** — AC-CRT-006b 의 Then 에 대체 부재 절 추가.

**선택 (운영자 재량):** D8-RESIDUAL(`Where` → `When`), N3(식별자 형식).

세 항목 전부 문서 편집이고 새 설계 결정을 요구하지 않는다 — §A.7 과 plan §B 가 필요한 결정을 이미 내려 뒀다. 따라서 iteration 3 을 여는 것보다 M1 착수 전 교정으로 처리하는 편이 싸며, 교정 여부는 오케스트레이터가 세 지점을 읽어 기계적으로 확인할 수 있다.

**Tier M 반복 상한(2)에 도달했다.** 위 세 건이 해소되지 않은 채 run-phase 로 진입해야 한다면 그것은 운영자 판단 사항이며, 그 경우 D2-RESIDUAL 은 부채가 아니라 **알려진 오설계의 착지**로 기록되어야 한다.

---

## §7 Evidence / Gaps / Residual-risk (iteration 2)

**Evidence** — 이번 라운드에 실제로 실행한 것: 개정 3파일 전문 판독(spec.md 238줄 / acceptance.md 175줄 / plan.md 128줄) · `grep -n '^### REQ-'` (8건, 연속 확인) · `grep -n '^### AC-'` (11건, `006b` 포함) · `grep -c syscall spec.md` → 0 · `grep -rn '\[NEEDS CLARIFICATION'` → 0 hit(rc=1) · `grep -Eoh 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'` → 참조 집합 불변 · `grep -n 'worktree_base_branch|설정값|설정 키' acceptance.md` → `:65,:67,:70` 3줄(D2-RESIDUAL 의 직접 증거).

**Baseline-attribution** — 전부 `.claude/worktrees/t399` @ `442da4f06`, 이번 실행. iter1 에서 확인한 코드 측 사실(§A.4 좌표 4건, 스키마 4행, `mcp_convergence.go:168-227`, `applyGateUnmet`, `codex_review_gate.go:90`, 라이브 선례 2파일, codex 0.150.1)은 같은 트리·같은 SHA 이며 이번 개정이 코드를 건드리지 않았으므로 재측정하지 않고 iter1 관측을 인용했다. 이 인용은 트리 불변을 전제로 하며, 그 전제는 SHA 동일성으로 성립한다.

**Gaps (이번에도 관측하지 않은 것)** — (1) 테스트를 실행하지 않았다. AC-CRT-003 의 "변경 전 초록" 전제는 iter1 에 이어 여전히 미확인이다. (2) 라이브 codex 왕복을 시도하지 않았다 — N1 의 판정은 "SPEC 이 관측을 주장하는데 그 관측 기록이 어디에도 없다"는 것이지, 거절이 실제로 일어나지 **않는다**는 주장이 아니다. 스키마상 거절이 예측되는 것은 여전히 타당하다. (3) `internal/cli` 코드는 이번 라운드에 재판독하지 않았다(위 baseline 참조) — 단 N2 의 근거인 `mcp_codex.go:1004` default 분기는 iter1 에서 직접 읽은 것이다. (4) codex-cli 0.149.0 차이는 여전히 관측 불가.

**Residual-risk** — 0.81 은 임계 0.80 에 근접하며, Completeness 0.85 와 Traceability 0.90 은 밴드 사이 판단이다. 그 둘을 각각 0.75 로 보수적으로 읽으면 총점은 0.75 로 임계 아래가 되므로, 이 PASS-WITH-DEBT 는 차원 판단에 민감하다. 다만 방향은 바뀌지 않는다 — 어느 읽기에서도 D2-RESIDUAL 은 run-phase 진입 전 해소 대상이고, 그것이 이 판정의 실질적 내용이다. 반대 방향의 위험도 적는다: N2 를 "구현이 상식적으로 그렇게 하지 않을 것"이라 읽으면 major 가 아니라 minor 이며, 그 경우 이번 라운드의 blocking 은 D2-RESIDUAL 과 N1 둘뿐이다.
