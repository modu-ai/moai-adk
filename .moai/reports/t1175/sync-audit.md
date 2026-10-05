auditor-model: claude-opus-5-5[1m]

# t1175 — SPEC-ALWAYS-LOADED-DIET-002 sync-phase 독립 감사

카드 t1175 · Tier L · 워크트리 `.claude/worktrees/t1175` · 브랜치 `WT-rules-diet`
감사 대상 ref: 요청 `4989ea6b0`(develop `b59a5d69c` 흡수 병합) — 감사 도중 HEAD 가 `805d44bed` 로 이동해 두 ref 를 모두 쟀다(아래 「프로세스 결함」).
평가 프로필: 기본 프로필(Functionality + Security must-pass), 평면 가중 모드. 판정: **FAIL** — 수리 범위는 좁다(F1~F4).

---

## 판정 요약

| 차원 | 점수 | 판정 | 근거 |
|---|---:|---|---|
| Functionality (40%) | 60/100 | FAIL | MUST-PASS AC-ALD2-005 를 이 카드가 만든 끊긴 앵커 2건이 위반한다(F1·F2). REQ-ALD2-001 은 미충족인데 REQ 층이 개정되지 않았다(F3). AC-ALD2-001 기록값은 병합 트리와 맞지 않는다(F4) |
| Security (25%) | 88/100 | PASS | 동결 다중집합 170줄 해시가 기준선·병합 전·병합 후·이동 후 HEAD 네 ref 에서 바이트 동일하다. 카드 범위 전체 트리에서 구속 줄 변경은 경로 재지정 1건뿐이고, 비밀 패턴 적중은 0건이다 |
| Craft (20%) | 72/100 | PASS | 영향 패키지 테스트와 예산 테스트는 통과했다. 다만 companion 안 중복 절 4곳, 거짓이 된 서술 1건(F5·F6)이 있고, 계획된 이동 R-70 은 실행되지 않았다 |
| Consistency (15%) | 88/100 | PASS | 건드린 규칙 쌍의 잔여 분기는 전부 중립성 클래스(SPEC ID·REQ 토큰)다. 카드가 만든 미미러 구조 변경은 없다 |

가중 조화평균 = 1 / (0.40/60 + 0.25/88 + 0.20/72 + 0.15/88) ≈ **71.5**. must-pass 방화벽에 따라 Functionality FAIL 이 전체 판정을 FAIL 로 확정한다.

### 질문별 답

1. **수명주기 종결의 정합성.** 이 SPEC 은 `implemented` 에서 멈춰 있다. "작업은 인도됐고 채무가 기록됐다"는 뜻으로 읽으면 이 상태는 정합적이다. 그러나 이 상태를 `completed` 로 올릴 근거는 되지 않는다. AC-ALD2-001 을 문턱에서 기록 술어로 바꾼 개정은 투명하게 공개됐고, 폐기 문안도 보존됐으며, 원 문턱이 동결 규칙 아래서 도달 불가였다는 논증도 서 있다. AC 층의 개정은 정당하다고 판정한다. 하지만 **REQ-ALD2-001 은 여전히 "shall 150,000자 미만"이다**(`spec.md:90`). 제목도 같은 목표를 내걸고 있고, 카드 본문도 150k 를 [HARD] 로 적었다. 기준을 개정하면서 그 기준이 인용하는 요구사항을 함께 훑지 않은 것은 `verification-completeness.md §3`(교차 층 개정 스윕)이 말하는 바로 그 결함이다. 정직한 처분은 **FAIL**(완료 전이 불가)이다. REQ 개정(F3)과 F1·F2·F4 수리 뒤 델타 재감사를 통과하면 **PASS-WITH-DEBT** 가 된다. 채무의 후속 카드 `t1226`(A_adm 실측)은 이미 큐에 있다.
2. **구속 조항의 손실·약화.** 16파일 동결 해시는 네 ref 모두에서 동일했다. 카드 범위(`172ef22eb..3a48485af`)의 `.claude`·`CLAUDE.md`·`AGENTS.md` 전체에서 구속 줄 변경은 `plan-auditor.md:237` 의 경로 재지정 1건뿐이고, 의무 내용은 같다. 병합 해소 쪽을 보면, develop 이 `moai-mcp-tools.md` 에 더한 사실 12개 항목은 두 트리의 catalogue 에 모두 실렸다. 반면 `4989ea6b0` 의 `CLAUDE.md §15` 는 develop 이 추가한 마이그레이션 플래그 요건(`--target claude-only --apply --accept-role-change`)과 "mixed-role 대상 차단" 문장, 그리고 `glm-web-tooling.md` 포인터를 떨어뜨렸다. 이 결손은 `805d44bed` 가 `model-policy.md` § Legacy CG Configuration 포인터를 달아 메웠다. 그 절은 두 트리 모두에서 플래그를 축자로 담고 있다. 떨어진 문장 가운데 `[HARD]`/`MUST` 토큰을 가진 줄은 없었다. 해시에 드러나지 않는 약화는 F7(잔여 위험)로 따로 적었다.
3. **앵커 무결성.** 이동된 절 51개(라이브)를 대상으로 두 방식의 트리 전수 스윕을 돌렸다. **이 카드가 만든, 어느 경로로도 해소되지 않는 끊긴 앵커는 2건**이고, 미러까지 세면 파일 위치 5곳이다(F1·F2). stub 포인터 줄을 경유해서만 해소되는 미수리 인용은 10건이다. 그중 초과 파일(40k 초과)에 있는 것은 2건뿐이다(F8).
4. **미러 동등성.** 건드린 규칙 30쌍에서 hunk 수를 세면 기준선 대비 증가한 쌍이 0이다. `main-checkout-branch-guard.md` 는 4→1, `skill-routing.md` 는 1→0 으로 줄었다. 잔여 분기(최대 `verification-claim-integrity.md` 12 hunk)는 기준선에서 승계된 것이고, 확인한 표본은 모두 SPEC ID·REQ 토큰 제거(중립성)다. 카드가 만든 미미러 구조 변경은 관측되지 않았다. 결함 F1·F2·F5 는 두 트리에 똑같이 들어 있어서 미러는 맞지만, 둘 다 틀렸다.
5. **알려진 후속 두 건.** (a) 템플릿 `main-checkout-branch-guard-detail.md` 의 맨손 `git worktree add` 는 **이미 해소됐다.** `0acfa28e1` 이 런처 안내(`moai worktree new` / `moai cc -w`)로 교체했고, 두 트리에서 grep 적중이 0이다. `moai worktree new` 동사가 실재함은 바이너리 도움말로 확인했다. 따라서 `verdict.md` 의 잔여 위험 항목과 큐 카드 `t1191` 은 낡은 기록이다. (b) AC-ALD2-003 재작성은 카드 안에서 이미 이뤄졌다(0.6.0·0.7.0·0.8.0, `card-created` 한정어). 큐 카드 `t1192` 도 낡은 기록이다. 두 카드의 처분은 운영자가 정할 일이다.

---

## Findings (structured defect-list)

- **F1** [Medium] [blocking] `.claude/agents/moai/sync-auditor.md:31` · `internal/template/templates/.claude/agents/moai/sync-auditor.md:31` · `internal/template/templates/.codex/agents/moai/sync-auditor.toml:20` — `7bcdce760` 이 이 줄을 `agent-common-protocol-reference.md` 로 재지정하면서 같은 줄 뒤쪽의 `§Language Handling` 까지 reference 파일을 가리키게 됐다. 그 절은 reference 파일에 없고 `agent-common-protocol.md:89` 에 남아 있다. 초과 파일이 아닌 곳의 끊긴 앵커이므로 AC-ALD2-005 선례 규칙에 따라 예외 없이 수리 대상이다. 확신도 높음. — Required fix: C2(템플릿 `.md`)와 C1(라이브 `.md`)에서 두 절에 각자의 파일을 붙인다 — "`agent-common-protocol-reference.md` §Skeptical Evaluation Stance … and `agent-common-protocol.md` §Language Handling". 그 뒤 `make agents-emit` 으로 C3(`.toml`)를 재생성한다. C3 는 손으로 고치지 않는다.
- **F2** [Medium] [blocking] `.claude/rules/moai/core/moai-constitution.md:37` · 템플릿 동일 줄 — `276391646` 이 `moai-constitution-detail.md` § Parallel Execution 포인터를 넣었다. 그러나 detail 파일에는 그 절이 없고(제목은 `Opus 5.5 Prompt Philosophy` · `Lessons Protocol` 둘뿐), 카드 범위에서 detail 파일은 수정되지 않았다. 재배치 표 R-70 은 이 행을 M1(이동)으로 적었지만 실제로는 삭제 뒤 끊긴 포인터만 남았다. 설명 내용은 `dynamic-workflows.md:19-27` 에 살아 있으므로 정보 손실은 작다. 확신도 높음. — Required fix: 첫 포인터를 지우고 `dynamic-workflows.md` 포인터만 남긴다(길이 감소). 재배치 표 R-70 을 실제 처리(삭제, 정본 `dynamic-workflows.md`)로 정정한다. 두 트리를 같은 커밋에서 고친다.
- **F3** [High] [blocking — `completed` 전이 조건] `.moai/specs/SPEC-ALWAYS-LOADED-DIET-002/spec.md:90` (REQ-ALD2-001) · `spec.md` 제목 — REQ 는 "shall 150,000자 미만"인데 병합 뒤 실측은 **198,413**(`805d44bed`)이다. AC-001 이 기록 술어로 바뀌어 AC 층은 통과하지만, 요구사항 층은 그대로 미충족 상태로 남아 있다. 확신도 높음. — Required fix(manager-spec 재위임): REQ-ALD2-001 을 개정 표시와 함께 둘로 나눈다. (a) 이 카드 몫은 "감축하고 잔여를 귀속과 함께 채무로 기록한다", (b) 런타임 한도 150k 달성은 후속 카드 `t1226` 으로 넘긴다. 제목에서 "150,000자 미만"을 목표가 아닌 채무 표기로 고치고, `§AC-ALD2-001.3` 의 "후속 카드"에 `t1226` 을 적는다. 이 개정이 끝날 때까지 `completed` 로 전이하지 않는다.
- **F4** [Medium] [blocking] `acceptance.md §AC-ALD2-001.2` · `progress.md §E.4`·재개 지점 — 기록값 197,897 은 병합 전 트리(`3a48485af`)의 값이다. 재개 절차 2단계는 병합 트리에서 다시 재라고 요구하지만, `remeasure/` 에는 18파일 합계와 동결 해시가 없다(`go-test.txt`·`lint.txt`·전역 `spec-lint.txt` 와 develop 추적 파일뿐). 병합 트리 실측은 `4989ea6b0` = 198,351, `805d44bed` = 198,413 이다. 통합될 트리에 대해 197,897 은 더 이상 참이 아니다. 확신도 높음. — Required fix: 기록 대상 ref 에서 18경로 `wc -m` 합계와 동결 해시를 다시 재서 명령·ref·출력을 붙여 기록한다. `805d44bed` 기준으로 감축 48,530, 잔여 48,413 이다. 이 값은 FAIL 조건 4(감축 > 0)를 여전히 만족한다.
- **F5** [Low] [optional] `.claude/rules/moai/core/agent-common-protocol-reference.md:266`(두 트리) — "The 4-step pattern and the asymmetric-retry summary remain inline there"는 기존 SPEC(`3f31135c3`) 때는 참이었다. 이 카드가 Error Recovery Pattern 본문을 모두 옮기면서 거짓이 됐다. stub 에서 `retry` 를 grep 하면 복구 신호 줄 1건만 나온다. 확신도 높음. — Required fix: 문장을 "The full pattern now lives here; AGENTS.md §7 keeps the summary" 식으로 사실에 맞춘다.
- **F6** [Low] [optional] companion 중복 절: `context-window-management-detail.md` 에 "Multi-session work" 절이 두 번 있고(115 `##` 는 기준선부터, 144 `###` 는 카드가 이동), 144절 안에는 자기 자신을 가리키는 "Detail: `context-window-management-detail.md`"가 있다. `native-idiom-and-register-detail.md` 의 "Why calques survive"(13·48), `main-checkout-branch-guard-detail.md` 의 "Why the race is quiet"(14)와 "Why This Matters"(154), `session-handoff.md:65` 와 `session-handoff-format.md:58` 의 "Auto-Injected Resume Flow"도 겹친다. 마지막 건은 stub 포인터가 "relocated"라고 적는데 stub 에도 절이 남아 있다. SSOT 가 갈라질 위험이 있다. 확신도 중간(내용 동일성까지는 대조하지 않았다). — Required fix: 기준선 절과 이동해 온 절을 하나로 합치고 자기 참조를 지운다.
- **F7** [Low] [optional] `.claude/rules/moai/core/agent-common-protocol.md` `## Tool Usage Guidelines` — 상시 로드 파일에는 `[HARD] Agents must follow tool usage patterns …` 한 줄만 남았다. 그 목적어인 패턴 목록은 paths 스코프 companion 에만 있다. 재배치 표 R-02 는 "Q1 도달(포인터 경유)"로 판단했고, `AGENTS.md §7` 이 요지를 담아 일부 완화된다. 동결 해시가 원리상 볼 수 없는 범위 축의 약화다. 확신도 중간. — Required fix(선택): 절 아래에 AGENTS.md §7 과 companion 을 가리키는 한 줄을 둔다.
- **F8** [Low] [optional] 포인터 줄 경유로만 해소되는 미수리 인용 10건: `output-styles/moai/moai.md:688·739`, `agent-common-protocol-reference.md:266·275`, `context-window-management.md:45`, `verification-batch-pattern.md:65·70`, `session-handoff-examples.md:8·15`, `CLAUDE.md:85`. AC-ALD2-005 경로 (a)가 허용하므로 FAIL 은 아니다. 다만 선례 규칙 2(길이 중립 불가 시 미수리)가 적용되는 초과 파일은 `session-handoff-examples.md` 2건뿐이다. 나머지 8건은 초과 파일 밖에 있어 경로를 바로잡는 것이 쉽다. 확신도 높음. — Required fix(선택): companion 을 직접 가리키도록 경로만 바꾼다.
- **F9** [Info] [optional] 큐 카드 `t1191`(맨손 `git worktree add`)과 `t1192`(AC-003 재작성)는 카드 안에서 이미 해소됐다(질문별 답 5). `verdict.md` 잔여 위험의 해당 항목도 낡았다. — Required fix: 통합 뒤 운영자에게 두 카드의 처분(done/drop)을 상신한다.
- **F10** [Info] [process] 감사 창 안에 외부 쓰기가 있었다 — 아래 「프로세스 결함」.

---

## 프로세스 결함 — 감사 중 외부 쓰기

`agent-common-protocol.md` § Background Agent Execution 은 "감사 중인 워크트리의 작성자는 하나"라고 정한다. 이번 감사 중 다음 쓰기가 관측됐다.

- 18:12 에 워킹 트리의 `CLAUDE.md`·`internal/template/templates/CLAUDE.md` 가 수정됐고, 18:14:08 에 커밋 `805d44bed`(`docs(t1175): point CLAUDE.md §15 CG line at model-policy migration flags`)가 들어왔다. 감사 시작 때 HEAD 는 `4989ea6b0`, 상태는 clean 이었다.
- 18:12 에 `remeasure/` 에 `develop-added.diff`·`develop-lines-trace.txt` 가 새로 생겼고, `spec-lint.txt`(0바이트)는 1.5MB 전역 lint 출력으로 덮어써졌다.

리드는 이 사실을 진행 기록에 남겨야 한다. 이 감사는 두 ref 를 모두 쟀으므로 판정의 기준이 무너지지는 않았다. 다만 `805d44bed` 는 감사 대상 요청 밖에서 들어온 커밋이다.

---

## Claim

1. 16파일 구속 조항 동결 다중집합은 기준선·병합 전·병합 후·현재 HEAD 에서 바이트 동일하다(170줄, `d97b33d9…c6c3`).
2. 18파일 `wc -m` 합계는 기준선 246,943 → 병합 전 197,897 → 병합 후 198,351 → 현재 HEAD 198,413 이다. 런타임 한도 150,000 은 미충족이다.
3. 이 카드가 만든 끊긴 앵커가 2건 있고(미러 포함 5곳), 모두 초과 파일 밖에 있다(F1·F2).
4. 건드린 규칙 30쌍의 미러 분기는 기준선 대비 증가하지 않았고, 카드가 만든 미미러 구조 변경은 0건이다.
5. 40k 파일당 래칫은 병합 뒤에도 유지된다(두 트리 모두 같은 4개, 같은 크기).
6. 영향 패키지 테스트, 예산 테스트, SPEC lint 는 현재 HEAD 에서 통과한다.
7. 알려진 후속 두 건은 카드 안에서 이미 해소됐다.

## Evidence

**동결 해시·합계**(측정 스크립트는 `git show <ref>:<path>` 로 18경로 `wc -m`(UTF-8)를 합산하고, AC-ALD2-002 파이프라인을 그대로 적용한다. 기준선 재현으로 스크립트를 검증했다):

```
$ bash measure.sh 172ef22eb
ref=172ef22eb total18=246943
frozen_lines=170 sha=d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3
$ bash measure.sh 3a48485af && bash measure.sh 4989ea6b0 && bash measure.sh b59a5d69c
ref=3a48485af total18=197897
frozen_lines=170 sha=d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3
ref=4989ea6b0 total18=198351
frozen_lines=170 sha=d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3
ref=b59a5d69c total18=247131
frozen_lines=170 sha=d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3
$ bash measure.sh 805d44bed
ref=805d44bed total18=198413
frozen_lines=170 sha=d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3
```

AC-ALD2-001 블록 명령을 워킹 트리(당시 `805d44bed` 와 같은 내용)에 그대로 실행한 결과의 마지막 줄: `  198413 total`.

**카드 범위 구속 줄 다중집합 델타**(동결 16파일 밖 포함):

```
$ bash binding_delta.sh 172ef22eb 3a48485af .claude CLAUDE.md AGENTS.md
removed_binding=1 added_binding=1
== removed-not-readded == [HARD] Read-only verification during audit … `.claude/rules/moai/core/agent-common-protocol.md` § Tool Selection by Task …
== added-not-from-removed == [HARD] Read-only verification during audit … `.claude/rules/moai/core/agent-common-protocol-reference.md` § Tool Selection by Task …
```

(`.claude/agents/moai/plan-auditor.md:237` — 앞 경로만 바뀌었고 의무 본문은 같다.)

**앵커 스윕**(인용 중 `§` 가 같은 줄의 직전 `.md` 토큰에 결합하는 것을 대상 파일의 제목과 대조했다. 기준선과 HEAD 의 끊김 집합 차이를 수작업으로 판독했다):

```
$ python3 anchor_resolve.py 172ef22eb  → ref=172ef22eb citations=635 broken=55
$ python3 anchor_resolve.py 4989ea6b0  → ref=4989ea6b0 citations=701 broken=75
NEW-at-HEAD 20줄 → 수작업 판독: 파서 분절 잡음 12, 포인터 경유 해소 4, 기존 SPEC 이력 서술 1(F5),
                   실제 끊김 2종 × 미러 = F1(sync-auditor .md×2, .toml×1), F2(moai-constitution.md×2)
$ grep -n '^#.*Language Handling' .claude/rules/moai/core/agent-common-protocol*.md
.claude/rules/moai/core/agent-common-protocol.md:89:## Language Handling
$ git show 172ef22eb:.claude/agents/moai/sync-auditor.md | grep -n 'Skeptical Evaluation Stance'
31:> See `.claude/rules/moai/core/agent-common-protocol.md` §Skeptical Evaluation Stance (…) and §Language Handling (…)
$ grep -n '^#' .claude/rules/moai/core/moai-constitution-detail.md
6:# MoAI Constitution — Detail Companion
13:## Opus 5.5 Prompt Philosophy
39:## Lessons Protocol
$ git log --format='%h %s' -1 -S'moai-constitution-detail.md` § Parallel Execution' -- .claude/rules/moai/core/moai-constitution.md
276391646 refactor(SPEC-ALWAYS-LOADED-DIET-002): finish M4, M5, M6 and M7 (card t1175)
```

이동 절 기준 좁은 스윕(`anchor_sweep.py`): `hits=10 strict_broken=10 lenient_broken(no pointer line in stub)=0`. 10건 목록은 F8 에 적었다.

**미러 hunk 수**(live vs template, 쌍별 `diff | grep -cE '^[0-9]'`), 기준선 → 병합 전 → HEAD. 발췌:

```
12 → 12 → 12  verification-claim-integrity.md      6 → 6 → 6  verification-claim-integrity-detail.md
 4 → 4 → 4   cross-session-messaging-detail.md    4 → 1 → 1  main-checkout-branch-guard.md
 3 → 3 → 3   main-checkout-branch-guard-detail.md 1 → 0 → 0  skill-routing.md
(나머지 쌍 0 또는 1 로 불변; 신규 companion kanban-dispatch-mechanics.md · session-handoff-format.md 는 0)
$ diff <live> <template> main-checkout-branch-guard-detail.md
36c36 < - **Opt-in gate (v1.2.0, SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 REQ-1/REQ-3)**:  > - **Opt-in gate (v1.2.0)**:
142,146d141 < Origin: SPEC-WORKTREE-BRANCH-GUARD-001 (REQ-WBG-001 through REQ-WBG-013). …
```

**파일당 40k 래칫**(HEAD, 두 트리 같은 출력):

```
   61435 …/worktree-integration.md
   41034 …/kanban-dispatch-detail.md
   40797 …/spec-workflow.md
   41615 …/session-handoff-examples.md
```

**테스트·예산·lint**(`805d44bed`, 환경 스크럽 단일 호출):

```
$ unset MOAI_KANBAN … && go test -count=1 -run 'TestAlwaysLoadedTokenBudget' -v ./internal/config/
    token_budget_guard_test.go:70: always-loaded surface = 65416 tokens (budget 77600, headroom 12184, 16 entries)
--- PASS: TestAlwaysLoadedTokenBudget (0.01s)
--- PASS: TestAlwaysLoadedTokenBudget_OverBudgetFails (0.01s)
$ unset MOAI_KANBAN … && go test -count=1 ./internal/template/ ./internal/config/   → exit=0
ok  	github.com/modu-ai/moai-adk/internal/template	63.739s
ok  	github.com/modu-ai/moai-adk/internal/config	2.963s
$ go run ./cmd/moai spec lint .moai/specs/SPEC-ALWAYS-LOADED-DIET-002   → exit=0
INFO  OwnershipTransitionUnmeasured  …/spec.md  1  … commit 5bfa134ca … has no Authored-By-Agent trailer …
0 error(s), 0 warning(s)
```

**병합 해소의 develop 사실 반영**(catalogue, 라이브/템플릿 적중 수):

```
Codex read-only roles 2/2 · codex_role_audit_result 2/2 · jev_ask 2/2 · Factory messaging 2/2 · factory_msg_status 2/2
graph_shortest_path 1/1 · claude_audit 4/4 · overall_verdict: fail 1/1 · goal_arm…intentionally wired 1/1
restart is procedure step zero 1/1 · Linked worktrees 1/1 · 39 tools 3/3
```

`CLAUDE.md §15` 비교: develop(`b59a5d69c`)은 "Removing automatic GLM teammate assignment requires `--target claude-only --apply --accept-role-change`; the mixed-role target remains blocked … See `glm-web-tooling.md` § CG Retirement and Migration."을 담는다. `4989ea6b0` 는 이 문장을 담지 않는다. `805d44bed` 는 "Migration flags: `model-policy.md` § Legacy CG Configuration."를 더했고, 그 절(`model-policy.md:220-222`, 두 트리)은 플래그와 차단 문장을 축자로 담는다.

**후속 건 해소 확인**:

```
$ grep -n 'git worktree add' .claude/rules/moai/workflow/main-checkout-branch-guard*.md internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard*.md
(출력 없음)
$ ./bin/moai worktree --help | grep 'new <name>'
    new <name>                    Create a worktree through the shared MoAI materializer
```

**AC-ALD2-006 좁힘 판정 근거**:

```
$ git diff --stat 172ef22eb 3a48485af -- …/verification-claim-integrity-detail.md …/session-handoff-examples.md
 2 files changed, 2 insertions(+), 2 deletions(-)
```

두 self-keyed companion 은 앵커 1줄만 바뀌었고, 옮겨 온 절은 없다. REQ-ALD2-004 의 When("절이 옮겨질 때")은 이 둘에서 발동하지 않는다. 따라서 이 좁힘은 결과에 맞춘 재단이 아니라 REQ 에 근거한 수리라고 판정한다.

**비밀·비마크다운 변경 탐침**: 카드 범위 추가 줄의 비밀 패턴 적중은 `0`이다. 비마크다운 변경은 `internal/template/catalog.yaml` 해시 5줄(에이전트·스킬 파일 편집에 따른 재계산, `7bcdce760`)뿐이다.

## Baseline-attribution

- 모든 수치는 이 감사 실행에서 이 워크트리를 대상으로 쟀다(2026-09-27, 18:0x–18:2x KST). ref 는 SHA 로 고정했고 브랜치 이름으로 적지 않았다.
- 기준선 `172ef22eb` 의 `246943` / `d97b33d9…c6c3` / `170` 을 내 측정 스크립트로 먼저 재현한 뒤 다른 ref 에 적용했다. 스크립트가 AC 명령과 같은 값을 낸다는 것을 기준선으로 확인한 것이다.
- `golangci-lint` 결과(`0 issues.`)와 `internal/spec` 테스트 결과는 오케스트레이터의 `4989ea6b0` 측정(`remeasure/lint.txt`·`go-test.txt`)을 **인용**했다. 다시 실행하지는 않았다. 이 카드의 Go 기여는 0개이고, `805d44bed` 는 마크다운 2개만 바꾸므로 lint 표면은 변하지 않는다. 다만 이것은 판단이지 측정이 아니다.
- 템플릿·설정 테스트와 예산 테스트는 `805d44bed` 에서 다시 실행했다.

## Gaps

- **교차 모델 2차 의견을 돌리지 않았다.** 이 스폰에는 `mcp__moai__audit_multi`/`codex_audit`/`glm_audit` 도구가 노출되지 않았다. `audit_model` 설정 키도 설정 파일에서 찾지 못했다(grep 0건).
- **F6 의 중복 절은 내용 동일성까지 대조하지 않았다.** 제목과 첫 문단만 봤다.
- **앵커 스윕은 "파일명 + § 제목" 형식에 한정된다.** 줄번호 인용, 파일명 없는 산문 언급(예: `session-handoff.md:60` 의 자기 파일 내부 "§ Localization Table")은 전수로 보지 않았다. 파서 분절 잡음은 수작업으로 걸렀으므로 사람의 판독 오류 가능성이 남는다.
- **승계 hunk 26건의 중립성 클래스 분류**는 이번에도 하지 않았다. 표본 2쌍만 확인했다.
- **재배치 표(AC-ALD2-004)의 Q1/Q2 판단**은 R-01·R-02·R-70 세 행만 다시 읽었다. 나머지 행은 재도출하지 않았다.
- **전체 테스트 스위트는 로컬에서 돌리지 않았다**(상시 금지). 전 패키지 판정은 develop CI 가 한다.
- **`internal/spec`·`agentemit` 테스트는 `805d44bed` 에서 다시 돌리지 않았다.** 오케스트레이터의 `4989ea6b0` 측정을 인용했다.

## Residual-risk

- **도달 범위 축의 약화는 기계 검사에 잡히지 않는다.** Skeptical Evaluation Stance·Error Recovery Pattern·Tool Selection by Task 가 상시 로드 파일에서 paths 스코프 companion 으로 옮겨졌다. companion 의 `paths:` 는 에이전트 정의나 워크플로 스킬을 **읽을 때** 발동하는데, 서브에이전트는 자기 정의 파일을 읽지 않으므로 이 내용이 실제 작업 세션에 닿지 않을 수 있다. 해시는 이것을 원리상 볼 수 없다. AGENTS.md §7 과 에이전트 본문의 직접 포인터가 일부를 막는다.
- **REQ-ALD2-001 의 잔여(48,413자)가 노력 부족인지 구조적 불가인지는 여전히 확립되지 않았다.** 판정은 `t1226` 에 달려 있다.
- **이 감사가 권고한 수리(F1·F2) 자체가 새 앵커를 끊을 수 있다.** 델타 재감사에서 같은 스윕을 다시 돌려야 한다.
- **감사 창 안에서 외부 쓰기가 있었다.** 이 보고서가 관측하지 못한 추가 쓰기가 있었을 가능성을 배제할 수 없다. 마지막으로 다시 읽은 HEAD 는 `805d44bed`, 상태는 clean 이었다.

VERDICT: FAIL score=0.71
