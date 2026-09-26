# SPEC-ALWAYS-LOADED-DIET-002 — 진행 기록

카드 t1175 · Tier L · 워크트리 `.claude/worktrees/t1175` · 브랜치 `WT-rules-diet`

---

## §E.1 Plan-phase Audit-Ready Signal

- 상태: plan phase 산출물 작성 완료 (`status: draft`)
- 산출물: `spec.md` · `plan.md` · `acceptance.md` · `design.md` · `research.md` · `progress.md` (Tier L 6종)
- SPEC ID 정규식 검사: Bash 실행, 출력 `PASS`
- 기준선(이 실행, 이 트리, 2026-09-25):
  - 18파일 `wc -m` 합계 = `246943 total`
  - 구속 조항 줄 = 170, sha256 `d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3`
- `[NEEDS CLARIFICATION]` **0건** — 작성 시점 2건이 리드 판정으로 해소됐다(`plan.md §B`):
  - B-1 `CLAUDE.md` 템플릿↔라이브 의미 분기 → **보존**, AC-ALD2-003 판정에서 제외. 실측으로 라이브 루트가 낡은 쪽임이 확정됐다(`moai cg --help` 가 은퇴 통지를 출력, `moai migrate cg` 존재). 별도 카드.
  - B-2 중복 제거와 동결의 경계 → 기본 입장 유지, **REQ-ALD2-013 으로 승격**.
- 리드 판정으로 추가된 항목: **범위 이탈** 잔여 위험(`acceptance.md §D.3`)과 재배치 표의 **범위 의존** 칸(AC-ALD2-004, `design.md §4`).
- **v0.2.0 (리드 판정)** — 파일당 40,000자 축 편입: REQ-014·015·016, AC-ALD2-009, `spec.md §D` 신규 Out of Scope 1건.
  - 이 축이 `design.md §2` 의 "새 companion 을 만들지 않는다" 결정과 충돌해 **설계를 개정**했다. 목적지는 존재가 아니라 **수용량**으로 고른다.
  - 실측: 계획대로면 셋이 한도 초과(`kanban-dispatch-detail` 57,036 · `session-handoff-examples` 51,616 · `agent-common-protocol-reference` 46,207). 셋째는 **이 카드가 만드는 신규 초과**이며 리드 보고에는 없던 항목이다.
  - 개정된 배정(`design.md §2.2`): 신규 companion 3개 생성, 이미 초과한 파일에는 쓰지 않는다. 마일스톤 M6.5.
  - 범위: 이미 초과한 룰 4개는 **수리하지 않고 악화만 금지**한다(REQ-015·016).
- 기준선 추가: 40,000자 초과 룰 파일 **4개**(두 트리 동일).
- **범위 이탈 공개를 양방향으로 확장**(`acceptance.md §D.3`) — 종전 공개는 축소 방향(참조 대상이 떠남)만 예시로 들었다. 확대 방향(범위를 **좁히던** 비구속 문장이 지워져 남은 조항이 넓어짐)을 방향 2 로 추가했고, 그것이 **결손이 아니라 증가**여서 손실 탐지기에 걸리지 않음을 명시했다. AC-ALD2-004 범위 의존 칸은 Q1(도달)/Q2(폭)의 두 질문을 묻도록 확장. 이 구멍은 감사가 아니라 **SPEC 작성자가 자기 공개에서 찾았고**, 발견 경위를 §D.3 에 남겼다.
- **v0.3.0 (plan-audit iter2 수리)** — 판정 FAIL 0.805 / Tier L 임계 0.85, MUST-PASS 7/7 통과, Clarity 0.65 가 하락 원인. 입력 커밋 `c95124a5a`, 보고서 `.moai/reports/t1175/plan-audit-iter2.md`.
  - **D5 는 설계 변경이다** — 감사의 "문서 편집, 설계 변경 없음" 분류를 받아들이지 않는다. 포인터 줄 재유입 항을 파일별 가중으로 실측해 투영을 gross/net 으로 재구성. **재유입은 이동량에 비례하므로 목표 집합마다 값이 다르다** — 종전 목표 8,796 / 개정 목표 9,019. 감사 추정(~16,300)은 과대계상이었고 예측한 5,443 부족은 성립하지 않으나, 종전 목표로는 여유가 10,857 → **2,061** 로 깎였다. 저재유입 파일 8개 목표를 상향해 복원: gross 117,200 · 재유입 9,019 · **net 108,181 · 여유 11,238**. 재유입이 실측 2배여도 통과(+2,220).
  - **D7** — 재배치 표 발동 조건이 "옮겼을 때"뿐이라 삭제 경로(`AGENTS.md` M2 · `CLAUDE.md` M1′, 계획의 11.1%)가 통째로 우회됐다. REQ-017(삭제 행 의무)·REQ-018(stub 포인터 줄 검사) 신설, AC-004 Given 을 제거 전반으로 확장. **Q1/Q2 는 이동 경로에 대해 옳았고 구멍은 질문이 아니라 발동 조건에 있었다.**
  - **D1** — 지시대로 한 건이 아니라 전수 스윕. 살아 있는 개정 전 진술 **2건**(감사가 찾은 `plan.md §D`, 작성자가 보류 중이던 `design.md §2`), 역사 서술 5건은 유지. 둘을 각각 다른 주체가 찾았고 **아무도 둘 다 찾지 못한 것**이 이 결함군의 성질이다.
  - D2(검사 허용 ≠ SPEC 허용) · D3(초과 파일을 허용 선례로 인용) · D4(트리 표기) · D6(§E ↔ MUST-PASS 일대일) · D8(표의 귀속 축 한정) · D10(앵커 수리 vs 초과 파일 선례 규칙) 수리.
- **v0.4.0 (plan-audit iter3 수리)** — 판정 FAIL 0.845 / 임계 0.85 (+0.040), MUST-PASS 7/7, iter2 결함 7건 CLOSED. 입력 커밋 `19b1e63e5`.
  - **N6** 재배치 가능 풀 **172,363 → 170,847**. 원인은 측정 스크립트의 `len(para)+2` — 문단 구분자를 재배치 쪽에 귀속시켜 파일마다 합이 `현재+2` 가 됐다. `goal-directive` 재배치 6,877 이 파일 크기 6,875 를 넘던 것이 증상. §B 헤드라인 수치라 자릿수만 고치지 않고 주장을 재확인했다 — 풀이 요구치의 1.76배, 결론 유지.
  - **D4 회귀** — 종전 수리가 "모두 라이브 트리 실측"이라 **재측정 없이 단언했고 사실과 반대**였다(§2=템플릿, §2.1=라이브). 두 트리를 재측정해 라이브로 통일하고, **14개 중 6개가 두 트리에서 갈린다**는 선행 분기를 명시했다. 미측정 귀속보다 측정한 듯한 틀린 귀속이 더 나쁘다는 판정에 동의한다.
  - **N2** 조항 선택 술어가 여전히 "참조하는"이라 방향 2 가 빠져나갔다 — 범위를 좁히던 문장은 참조되지 않고 **위치로** 지배하기 때문. **지배 = 참조 ∪ 인접 ∪ 용어 의존**으로 교체하고 세 축 모두 확인한 뒤에만 `없음` 을 적게 했다. 3개 산출물 전파.
  - **N3** §4 표 **스키마**가 이동 전용이라 REQ-017 이 요구하는 삭제 행을 담을 수 없었다. `기제` 열 추가, M1′·M2 행의 목적지 3칸을 `해당없음` 으로 규정. D1 의 모양이 설계층에서 재현된 것.
  - **N4** `7개 → 8개` 를 15줄 아래 형제에서 놓쳤다. 이 패턴의 세 번째 출현(D1·N3·N4)이며, 이후로는 인용된 줄이 아니라 **클래스를 수리하고 형제를 전수 grep** 한다.
  - N5(AC 철자 통일 — 스냅숏 가드 적색화 방지) · N1/D9(§E 의 "일대일" 주장 정정 — 덮개는 성립, 주장이 거짓이었다) 수리.
  - **공개 2건 편입**: §D.3.1 율 표 복제 상태, §D.3.2 재유입 추정의 **입도** — 재유입은 문자량이 아니라 절 개수에 비례하는데 율이 기존 분리 입도를 대리 변수로 가정한다.
  - §D.3.1 은 작성 중 **내용이 뒤집혔다**. 리드가 최고가중 2개를 재도출한 시점의 "14개 중 2개 재현"으로 쓰기 시작했으나, iter3 감사 부록이 2층 재계산으로 **14개 전부**를 파일에서 재측정해 전부 일치시켰다(69줄 / 20,175자 / 평균 292.39 / 집계 8.6738%). 따라서 **잔여 위험이 아니라 복제 상태 기록**으로 다시 썼다. 남긴 것은 수치가 아니라 **경위** — 이 표는 한동안 D5 전체가 기대는 단일 측정이었고, 재현 전후의 문서는 수치가 한 글자도 다르지 않다. 복제되었다는 사실은 표를 봐서는 알 수 없고 적어야만 남는다.
- **v0.5.0 (plan-audit iter4 수리)** — 판정 **PASS-WITH-DEBT 0.890** (+0.045, 임계 0.85 초과), iter3 결함 전부 CLOSED, 회귀 0. 입력 커밋 `07eb0c703`. **범위 축소 없음** — 타당성 논거가 이번 실행에서 전수 재도출돼, 축소는 좁은 완화를 계획 전체의 실패처럼 다루는 것이 된다는 판단.
  - **N7 — 세 번째 통과 변이, 그리고 진단이 곧 수리의 형태였다.** `kanban-dispatch.md` 의 `## Scope — when this rule is live` 절은 구속 토큰 0이라 합법적 이동 대상인데, 그것이 떠나면 `[HARD] Never spawn background load` 가 kanban 한정에서 **매 세션 구속**으로 읽힌다. 참조·인접·용어 **셋 다 근접성을 요구하므로** 원거리 지배를 구조적으로 못 본다.
  - 수리는 네 번째 가지를 더하는 것이 **아니다** — 목록이 길어질 뿐 같은 실패가 재현된다. **닫힌 열거를 열린 정의로 교체**하고 셋을 예시로 내렸다. 판정 기준은 하나: "이 절이 사라지면 이 조항이 구속하는 대상이 달라지는가".
  - **이 카드가 낳은 넷째 교훈이 이 카드 안으로 되돌아온 자리다** — 탐색의 형식이 찾을 수 있는 것의 모양을 정하고, 닫힌 열거는 하나의 탐색 형식이다. 앞의 셋(거짓 안심·비독립 검증·닿지 않는 증거)은 더 주의 깊게 읽어 고쳐지지만 이것은 아니다. 목록을 아무리 성실히 돌려도 목록 밖은 구조적으로 안 보인다.
  - **D9 종결** — §E 는 행 5~8 에만 AC 번호가 있어 대조가 절반만 가능했다. 전용 `AC` 열로 여덟 행 전부 병기. 같은 수정에서 행 4 의 "옮긴 절"→"제거된 절"(N3 의 이동 전용 어법 잔재 — 삭제도 앵커를 끊고, 그쪽은 경유할 포인터가 없어 더 조용히 끊긴다). 앵커 무결성 절도 이동/삭제 두 갈래로 나눴다.
  - **M1 선행 조건 명시** — M1 의 산출물이 재배치 표 자체이므로 술어 수정은 표를 채우기 전에 끝나 있어야 한다. v0.5.0 이 이미 담고 있으나 run phase 가 잃지 않도록 `plan.md §C` 에 기록.
  - **§D.3.3 감사 자기 Gap 3건** — N7 은 검증된 사례 하나에 서 있고(`main-checkout-branch-guard.md` 축의 두 번째 변이는 미확인), 이번 실행의 율 재측정은 14개 중 3개, 동결 해시는 재계산 없이 인용됐다. **감사자의 Gap 이지 이 SPEC 의 것이 아니므로** 출처를 갈라 적었다 — 적지 않으면 "55개 검사 통과"가 전수 검증처럼 읽힌다.
- REQ 18개 / AC 9개(MUST-PASS 8) / 추적성 18행. `moai spec lint` exit 0, findings 0.
- plan-audit iteration 1 ABORTED(혼합 세대 읽기), iteration 2 FAIL 0.805 → 수리 완료, iteration 3 대기.

## §E.2 Run-phase Evidence

**이 절은 늦게 쓰였다.** 구현은 `3c40e0aa2`(M1)부터 `625679ace`까지 이미 커밋돼 있었고, 기록만 비어 있었다. 아래 수치는 전부 **이 record-closing 실행에서 이 트리(`WT-rules-diet` · HEAD `625679ace`)에서 다시 쟀다** — `verdict.md` 나 감사 보고서에서 옮겨 적은 값이 아니다. 옮겨 적은 항목은 그 자리에 **인용임을 명시**했다.

**[추가 실행 2026-09-26 — 재측정]** AC 매트릭스의 **두 행만** 더 나중 ref 에서 왔다: `AC-ALD2-006`(좁힌 Given 재측정)과 `AC-ALD2-008`(사후 재도출)은 **판정 ref `cb8466d09`** 에서 쟀다. 다른 모든 행의 귀속은 위 문단 그대로(HEAD `625679ace`)이며, 이 실행이 다시 확인한 것은 세 불변식(18파일 합계 `197897 total` · 동결 해시 `d97b33d9…18c6c3` 170줄 · AC 수 9)뿐이다.

### 실행 범위와 커밋

- 분기점 `a0b78213d`(로컬 `develop`), 기준선 ref `172ef22eb`(plan phase 마지막 커밋), 최종 ref `625679ace`. 미푸시.
- 카드 기여 파일 **85개**(`git diff --name-only a0b78213d HEAD | wc -l` = 85). **Go 소스 0개** — 라이브 룰 30 + 템플릿 미러 30 + 에이전트 정의 6(라이브 3 · 템플릿 3) + Codex 방출 3 + SPEC 산출물 6 + 나머지.
- 마일스톤 8개(M1 · M2 · M3 · M4 · M5 · M6 · M6.5 · M7)를 커밋 4개에 실었다: `3c40e0aa2`(M1 재배치 표) · `4ab0511b4`(M2) · `7bcdce760`(M3 + M6.5) · `276391646`(M4 · M5 · M6 · M7). 이후 `0acfa28e1`(미러 수리) + SPEC 산출물 커밋 4개(`14a5a3caf` · `f0893dc36` · `87fd59092` · `625679ace`).

### AC 판정 매트릭스

| AC | 심각도 | 판정 명령 | Actual Output | Status |
|---|---|---|---|---|
| AC-ALD2-001 | MUST-PASS | `acceptance.md §AC-ALD2-001` 의 18경로 `wc -m … \| tail -1` | `  197897 total` | **PASS** (기록 술어 — 아래 단서 참조) |
| AC-ALD2-002 | MUST-PASS | `acceptance.md §AC-ALD2-002` 의 `grep -rhE … \| sed \| sort \| shasum -a 256`, 줄 수 `wc -l` | `170` · `d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3` | **PASS** — 기준선과 바이트 동일 |
| AC-ALD2-003 | MUST-PASS | `§AC-ALD2-003.1` 2단계(hunk 계수 + 기준선 본문 바이트 대조) | `TOUCHED_LIVE=30` · `PAIRS_WITH_TEMPLATE=30 TOTAL_HUNKS=32 DIVERGING_PAIRS=11` · `TOTAL: inherited=29 card-created=3` | **PASS** — card-created 3건 전부 PERMITTED, UNCLASSIFIED 0. **수리로 달성됐다**(아래) |
| AC-ALD2-004 | MUST-PASS | `design.md §4` 재배치 표 판독 | 표는 `217` 행부터 존재, §4 구간 표 행 185, 빈 칸 스캔 적중 0 | **PASS (인용 + 구조 확인)** — 칸 내용의 **판단 축**(Q1/Q2 적정성)은 이 실행에서 재도출하지 않았다 |
| AC-ALD2-005 | MUST-PASS | 절 제목 역방향 grep(`acceptance.md §AC-ALD2-005`) | 이 실행에서 **전수 재실행하지 않았다**. 인용: 초과 파일 앵커 미수리 10건(선례 규칙 3 에 따라 FAIL 아닌 기록 사항) | **PASS-WITH-DEBT** — 네 인용 형식 전체에 대한 트리 전수 최종 스윕이 없다(Gap) |
| AC-ALD2-006 | MUST-PASS | `grep -m1 '^paths:' <companion>` — **좁힌 Given**(이 카드가 한 절을 옮겨 넣은 companion) 전수 11개 | 11개 모두 부모 stub 경로 하나만 담지 않는다. self-keyed 0건 | **PASS** (재측정, ref `cb8466d09`) — 종전 FAIL 은 아래 `### AC-ALD2-006 재측정` 이 보존한다 |
| AC-ALD2-007 | SHOULD-PASS | 신규 stub 2개의 3요소 판독 | (a) 포인터 줄 — `kanban-dispatch.md:9`, `session-handoff.md:7` 모두 절을 이름으로 호명 + 로드 트리거 있음. (b) companion 소유 경계 — 두 신규 companion 모두 `> Owns:` 줄 보유. (c) 분리를 기록하는 stub 푸터 — `kanban-dispatch.md` 푸터는 `kanban-dispatch-detail.md` 만 적고 신규 mechanics companion 을 적지 않으며, `session-handoff.md` 푸터(`Status: HARD operational rule …`)는 분리를 전혀 적지 않는다 | **PASS-WITH-DEBT** — (a)(b) 충족, (c) 2건 미충족 |
| AC-ALD2-008 | MUST-PASS | companion 11개 각각 `post − pre ≤ 옮긴절합계 + 500`. `pre`/`post` 는 `git show <ref>:<path> \| wc -m`, 옮긴절합계는 **기준선 stub 안의 그 절 본문** 문자수 | 11/11 통과. 최소 여유 `agent-common-protocol-reference` **+176**, 최대 `kanban-dispatch-mechanics` +4,250. 합계 delta 40,582 vs 옮긴절합계 47,980 | **PASS** (사후 재도출, ref `172ef22eb` → `cb8466d09`) — 종전 UNVERIFIED 는 아래 `### AC-ALD2-008 재도출` 이 보존한다 |
| AC-ALD2-009 | MUST-PASS | `acceptance.md §AC-ALD2-009` 의 두 `find … wc -m … awk '$1>=40000'` | 두 트리 동일하게 4개: `worktree-integration.md 61435` · `session-handoff-examples.md 41615` · `kanban-dispatch-detail.md 41034` · `spec-workflow.md 40797`. 기준선 대비 각 동일 또는 감소(41616→41615 · 41036→41034 · 40799→40797), 신규 초과 0 | **PASS** — (a) 이 카드가 쓴 파일 중 목록에 든 것 0, (b) 목록 길이 4 (래칫 유지) |

### 불변식 행

| 불변식 | 판정 명령 | Actual Output | Status |
|---|---|---|---|
| 구속 조항 축자 동결 (REQ-ALD2-002 · -003 · -011 · -013) | AC-ALD2-002 파이프라인 | 170줄 · `d97b33d9…18c6c3` — 기준선과 동일 | **HOLDS** |
| 파일당 40,000자 래칫 (REQ-ALD2-015 · -016) | AC-ALD2-009 파이프라인 | 두 트리 4개, 신규 초과 0, 크기 비증가 | **HOLDS** |
| 이 카드가 만든 구조 변경의 미러 (REQ-ALD2-007) | AC-ALD2-003 census | card-created 3, UNCLASSIFIED 0 | **HOLDS**(수리 후) |
| 영향 패키지 테스트 | `go test ./internal/config/... ./internal/template/... ./internal/spec/...` | `ok internal/config 2.644s` · `ok internal/template 59.096s` · `ok internal/template/agentemit 0.366s` · `ok internal/spec (cached)` · 나머지 `ok`/`(cached)`, 실패 0 | **HOLDS** |
| 빌드 + 임베드 | `make build` → exit 0; `git status --short` | `catalog.yaml updated successfully (13408 bytes)` · `go build -ldflags … -o bin/moai ./cmd/moai`, exit 0. `git status --short` **빈 출력** → `catalog.yaml` 내용 불변 | **HOLDS** |
| 크로스 플랫폼 | `GOOS=windows GOARCH=amd64 go build ./...` | 출력 없음, exit 0 | **HOLDS** |

### 헤드라인 AC 를 정직하게 적는다 — PASS 가 무엇이 아닌가

**AC-ALD2-001 의 PASS 는 「런타임 한도를 지켰다」가 아니다.** 이 AC 는 0.7.0(iter5 수리)에서 **문턱이 아니라 기록 술어**로 제자리 개정됐다 — 측정하고, 기준선 `B = 246,943` 대비 감축을 적고, 런타임 한도 150,000 에 대한 잔여를 귀속과 함께 채무로 남기는 것이 판정 대상 전부다.

- 측정 `total` = **197,897** · 감축 = **49,046** · 런타임 한도 대비 잔여 = **47,897**.
- **잔여 47,897 의 성격은 확립되지 않았다.** 노력 부족으로도 구조적 불가로도 **아니다** — 허용 풀 `A_adm` 이 구성상 미측정이기 때문이며(`acceptance.md §AC-ALD2-001.3`), 어느 쪽이라고 적으면 관측 없는 주장이 된다.
- FAIL 조건 4(`B − total > 0`)는 **래칫 방향만 본다.** 1자만 줄인 실행도 통과하므로 이 PASS 는 노력의 충분성을 재지 않는다.
- 폐기된 두 수치를 함께 남긴다: 종전 문턱 `< 150,000`(0.6.0 이전) → 구조적 하한 `< 171,695`(0.6.0) → 기록 술어(0.7.0). 조용히 바뀐 목표는 골대 이동과 구별되지 않는다.

**AC-ALD2-003 의 깨끗함은 관측된 것이 아니라 수리로 달성됐다.** 순서를 적어 둔다 — 이 구별이 사라지면 다음 감사가 「처음부터 0건이었다」로 읽는다.

1. `0acfa28e1` 이 `main-checkout-branch-guard-detail.md` 의 미분류 3 hunk(런처 동사 축)와 `cross-session-messaging.md` 의 `42,43` 분기를 수리했다.
2. `f0893dc36` 이 `cross-session-messaging-detail.md` 템플릿 사본의 `150a151` 중복 빈 줄을 제거했다.

또한 **승계 분기는 이 AC 의 판정 대상 밖이며, 그 결정이 분기를 해소했다는 뜻은 아니다.** 승계 29 hunk 중 셋(`verification-claim-integrity.md` 의 `3c3` · `5c5` · `21c21`)은 열거된 18개 허용 클래스로 **덮이지 않는다** — 숨기지 않고 이름을 적는다. 처분은 후속 카드다.

### AC-ALD2-006 FAIL — 넓은 Given 아래의 종전 기록 (재측정으로 **대체됐으나 지우지 않는다**)

> **이 절은 종전 상태 그대로 남는다.** 아래 `### AC-ALD2-006 재측정` 이 좁힌 Given 으로 다시 재어 PASS 를 얻었지만, 좁힘이 정당한지는 이 FAIL 이 보이는 동안에만 판단될 수 있다. 좁힘 뒤 남는 PASS 만 있는 기록은 「처음부터 통과했다」와 구별되지 않는다.

`verdict.md` 는 AC-ALD2-006 을 판정하지 않았다. 이 record-closing 실행이 이 카드가 수정한 companion 전수에 `grep -m1 '^paths:'` 를 돌려 **자기 참조 2건**을 관측했고, 둘 다 `acceptance.md §AC-ALD2-006` 의 FAIL 예(부모 stub 경로 하나만 담는 값)와 같은 모양이다.

- `core/verification-claim-integrity-detail.md` — `paths: "**/verification-claim-integrity*.md"` (부모 stub 계열만)
- `workflow/session-handoff-examples.md` — `paths: "**/session-handoff.md"` (부모 stub 만)

**귀속(provenance)**: 두 값은 **기준선 ref `172ef22eb` 에서 바이트 동일**하다(`git show 172ef22eb:<path> | grep -m1 '^paths:'` — 두 건 모두 현재 값과 같은 출력). 즉 **이 카드가 만든 값이 아니라 승계된 값**이고, 이 카드는 그 파일들의 본문만 고쳤다.

그럼에도 **FAIL 로 적는다.** AC-ALD2-006 의 Given 은 「이 SPEC 이 수정한 각 companion 파일」이고, 두 파일은 수정됐다. AC-ALD2-003 이 0.7.0 에서 얻은 `card-created` 한정어를 AC-ALD2-006 은 갖고 있지 않으므로, 문언대로 읽으면 통과하지 않는다. 이 기록을 「승계라서 대상 밖」으로 스스로 완화하는 것은 기준을 자기 산출물에 맞춰 재단하는 일이다(같은 판단을 `§AC-ALD2-003.2` 가 `codification` 토큰 제거로 이미 한 번 내렸다).

**따라서 후속 판정이 필요하고, 그것은 이 에이전트의 소관이 아니다.** AC-ALD2-006 에 `card-created` 대칭 한정어를 넣을지(AC-ALD2-003 과의 정합) 아니면 두 `paths:` 값을 domain-keyed 로 고칠지는 SPEC 본문 개정 또는 범위 확대이며, 어느 쪽도 run phase 의 소유 범위 밖이다 — blocker 로 올린다.

### AC-ALD2-006 재측정 — 좁힌 Given 아래에서 무엇을 쟀는가

**순서를 먼저 적는다.** 종전 측정은 **넓은 Given**(「이 SPEC 이 **수정한** 각 companion」) 아래에서 돌았고 self-keyed 2건으로 **FAIL** 이었다(위 절). `cb8466d09` 이 Given 을 `REQ-ALD2-004` 의 When 에 맞춰 **「한 절을 옮겨 넣은」** 으로 좁혔고, 그 좁힘이 이 재측정보다 **먼저** 커밋됐다 — 순서를 커밋 그래프가 증언하게 하려는 의도이며(`verification-claim-integrity.md §2.3`), 그 성질을 지키려고 `cb8466d09` 을 amend 하지 않고 이 재측정을 별도 커밋에 실었다.

**판정 ref 는 `cb8466d09`, 기준선 ref 는 `172ef22eb`.**

#### 대상 집합을 스스로 판정했다 — 이름 목록이 아니라 순증으로 가른다

`cb8466d09` 저자가 보고한 「순증 11개 · `1/1` 4개」를 **주장으로 받아 재현**했다. 발동 조건은 「diff 파일 목록에 이름이 있는가」가 아니라 **내용의 순증**이다.

- `git show <ref>:<path> | wc -m` 을 두 ref 에서 돌려 companion 15개의 문자 델타를 쟀다. 순증 **11개**(+850 ~ +7,432), 나머지 **4개**는 `−1` · `−2` · `+10` · `+64`.
- 그 4개는 diff 본문을 읽어 **한 줄 편집**임을 확인했다 — 앵커 수리 3건(`verification-claim-integrity-detail` `agent-common-protocol.md`→`agent-common-protocol-reference.md` · `session-handoff-examples` `session-handoff.md`→`session-handoff-format.md` · `kanban-dispatch-detail` 같은 형태)과 **`paths:` 확장 1건**(`cache-aware-execution-reference`, +64 의 정체가 그 한 줄). 절의 유입은 없다.
- 가장 작은 순증 둘(`skill-routing-detail` +850 · `goal-directive-detail` +976)도 diff 를 읽어 **`##` 절 통째 유입**임을 확인했다 — 문자량만으로 절 유입을 단정하지 않았다.

#### 11개 전수 판정 — `grep -m1 '^paths:'`

| companion | `paths:` | 판정 |
|---|---|---|
| `core/agent-common-protocol-reference` | `**/agent-common-protocol.md,**/.claude/agents/moai/*.md,**/.claude/skills/moai/workflows/*.md` | PASS |
| `core/askuser-protocol-reference` | `**/askuser-protocol.md,**/.claude/skills/moai/workflows/*.md,**/.claude/output-styles/moai/*.md` | PASS |
| `workflow/session-handoff-format` | `**/session-handoff*.md,**/.claude/output-styles/moai/*.md,**/.moai/state/handoff/**` | PASS |
| `workflow/cross-session-messaging-detail` | `**/cross-session-messaging*.md,**/kanban-dispatch*.md` | PASS (단서 — 아래) |
| `workflow/context-window-management-detail` | `**/context-window-management*.md,**/session-handoff*.md,**/internal/statusline/**` | PASS |
| `workflow/kanban-dispatch-mechanics` | `**/kanban-dispatch*.md,**/.claude/agents/moai/manager-lead.md,**/.moai/state/integration/**` | PASS |
| `core/moai-mcp-tools-catalogue` | `**/moai-mcp-tools.md,**/internal/cli/mcp_server.go,**/.claude/agents/moai/*.md` | PASS |
| `workflow/main-checkout-branch-guard-detail` | `**/main-checkout-branch-guard*.md,**/internal/hook/branch_guard.go,**/internal/hook/pre_tool.go` | PASS |
| `core/native-idiom-and-register-detail` | `**/native-idiom-and-register.md,**/.claude/skills/moai-domain-humanize/**` | PASS |
| `workflow/skill-routing-detail` | `**/skill-routing.md,**/skill-authoring.md,**/agent-authoring.md` | PASS (단서 — 아래) |
| `workflow/goal-directive-detail` | `**/.moai/state/goal/**,**/.claude/skills/moai/workflows/goal.md,**/goal-directive.md,**/goal-directive-detail.md` | PASS (선례) |

**11개 중 self-keyed 0건 → PASS.** AC 가 적은 FAIL 형태(부모 stub 경로 **하나만** 담는 값)는 11개 어디에도 없다.

**단서 2건을 PASS 안에 적는다 — 반올림하지 않는다.** `cross-session-messaging-detail` 과 `skill-routing-detail` 은 Then 의 첫째 절(부모 경로만 담지 않는다)을 만족하지만, 부모 밖 항목이 **전부 형제 룰·독트린 문서**다(`kanban-dispatch*.md` · `skill-authoring.md` · `agent-authoring.md`). 즉 그 룰이 지배하는 **작업 세션**이 아니라 **문서를 편집하는 세션**에 도달한다 — self-keyed 의 해악이 사라진 것이 아니라 희석된 형태다. 선례 `goal-directive-detail` 은 `.moai/state/goal/**` 이라는 **작업 산출물 경로**를 담아 그 축을 넘는다. 둘째 절(「그 companion 이 다루는 작업이 건드리는 도메인 경로」)이 이 둘에 대해 확립됐는지는 이 실행이 판정하지 않았다 — Gap 으로 적는다.

#### 범위 밖 2건 — 고치지 않고 후속 후보로 이름만 남긴다

`core/verification-claim-integrity-detail.md`(`**/verification-claim-integrity*.md`)와 `workflow/session-handoff-examples.md`(`**/session-handoff.md`)는 절의 유입이 없어(각 `1/1`, 문자 델타 `+10` · `−1`) 좁힌 Given 의 대상에서 내려왔다. **두 값은 여전히 self-keyed 이고, 이 실행은 고치지 않았다** — `acceptance.md §AC-ALD2-006.4` 가 이름 붙인 **선존 결함 후속 후보** 그대로다. 발행은 운영자·리드 소관이다.

### AC-ALD2-008 재도출 — 「불가」를 철회하고 사후에 쟀다

**종전 기록은 UNVERIFIED 였고, 그 근거는 「옮긴 절별 문자 합계는 이동 시점의 측정이라 다시 만들 수 없다」였다.** 리드가 그 근거를 받지 않았고, 근거가 실제로 틀렸다 — 기준선 ref `172ef22eb` 와 재배치 커밋이 **모두 그래프에 있으므로** 절 본문은 사후에 복원된다. 종전 판정과 그 근거를 지우지 않고 여기에 적어 둔다.

**방법**: companion 이 **획득한 절 제목**을 diff 에서 뽑고(`git diff <base> <judge> -- <companion> | grep '^+#'`), 각 제목을 **기준선 stub 사본**(`git show 172ef22eb:<stub>`)에서 찾아, 그 절 본문(제목 줄부터 같거나 상위 레벨의 다음 제목 전까지)의 문자수를 센다. `pre`/`post` 는 같은 두 ref 에서 `wc -m`.

| companion | pre | post | delta | 옮긴절합계 | 허용(+500) | 여유 |
|---|---|---|---|---|---|---|
| `core/agent-common-protocol-reference` | 31,207 | 38,639 | 7,432 | 7,108 | 7,608 | **+176** |
| `core/askuser-protocol-reference` | 15,646 | 21,080 | 5,434 | 5,124 | 5,624 | +190 |
| `workflow/session-handoff-format` | 0 | 5,266 | 5,266 | 6,926 | 7,426 | +2,160 |
| `workflow/cross-session-messaging-detail` | 15,221 | 19,006 | 3,785 | 4,295 | 4,795 | +1,010 |
| `workflow/context-window-management-detail` | 7,486 | 11,742 | 4,256 | 5,120 | 5,620 | +1,364 |
| `workflow/kanban-dispatch-mechanics` | 0 | 5,224 | 5,224 | 8,974 | 9,474 | +4,250 |
| `core/moai-mcp-tools-catalogue` | 11,424 | 15,495 | 4,071 | 5,429 | 5,929 | +1,858 |
| `workflow/main-checkout-branch-guard-detail` | 9,984 | 11,348 | 1,364 | 1,272 | 1,772 | +408 |
| `core/native-idiom-and-register-detail` | 2,509 | 4,433 | 1,924 | 1,920 | 2,420 | +496 |
| `workflow/skill-routing-detail` | 1,351 | 2,201 | 850 | 847 | 1,347 | +497 |
| `workflow/goal-directive-detail` | 20,283 | 21,259 | 976 | 965 | 1,465 | +489 |
| **합계** | | | **40,582** | **47,980** | | |

**11/11 통과 → PASS.** 이 AC 가 재는 것은 「원본에 없던 내용의 획득」이며, 전체적으로 companion 은 옮긴 절의 합보다 **7,398자 적게** 자랐다(이동 중 압축의 흔적이고, 압축은 이 AC 의 위험 방향이 아니다).

#### 절 두 개는 제목 대조로 보이지 않았다 — 그 둘을 따로 쟀다

제목 키 조회가 `NOT_FOUND` 를 낸 절이 넷 있었고, **그중 둘이 판정을 뒤집을 수 있는 자리**였다(그 둘을 0으로 두면 `cross-session-messaging-detail` −158, `context-window-management-detail` −60 으로 FAIL 로 읽힌다). 원인은 획득이 아니라 **이동 단위가 절이 아니라 절 안의 문단 블록**이라는 것이다 — 제목이 목적지에서 새로 붙었으므로 기준선에 그 제목이 없다.

- `cross-session-messaging-detail ## The stopped-teammate registry` ← 기준선 `cross-session-messaging.md` 의 `## Rules` 안 `**Mechanism layer (landed).**` 블록(`git diff -U0` 훅 `@@ -51,15 +39,9 @@` → 기준선 51~65행). **1,168자.**
- `context-window-management-detail ## Snapshot confidence` ← 기준선 `context-window-management.md` 의 `## Detection Heuristics` 안 `**Where the number comes from…**` 한 줄(훅 `@@ -108 +66,9 @@` → 기준선 108행). **1,424자.**

두 값을 더한 뒤 각각 여유 **+1,010** · **+1,364** 로 통과한다. 남은 `NOT_FOUND` 둘은 판정에 영향이 없다: `## Relocated bodies …` 2건은 목적지에서 새로 쓴 **묶음 제목**(접합부이므로 500 여유에 귀속되고 크레딧 0), `kanban-dispatch-mechanics ## Factory Mode mechanics` 는 기준선의 `## Factory Mode — the card travels whole` 이 이동 중 **개명**된 것이라 크레딧 0 으로 **과소** 계상됐고 그래도 여유 +4,250 이다.

### plan-audit 이력 — 점수가 내려간 구간을 남긴다

| iteration | 판정 | 입력 커밋 | 처분 |
|---|---|---|---|
| iter1 | ABORTED (혼합 세대 읽기) | — | 재실행 |
| iter2 | FAIL 0.805 | `c95124a5a` | 수리 → v0.3.0 |
| iter3 | FAIL 0.845 | `19b1e63e5` | 수리 → v0.4.0 |
| iter4 | **PASS-WITH-DEBT 0.890** | `07eb0c703` | 수리 → v0.5.0 |
| iter5 | **FAIL 0.71** | — | 수리 → v0.7.0 (`f0893dc36`) |
| iter6 | **PASS-WITH-DEBT 0.85** | — | 채무 D16~D19 해소 → v0.8.0 (`625679ace`) |

iter4 0.890 에서 iter5 0.71 로 **점수가 내려갔다.** 원인은 새 결함이 아니라 AC-ALD2-001·003 의 문턱이 구성상 만족 불가였음이 iter5 에서 드러난 것이고, 그 수리가 문턱 형식 자체의 폐기(기록 술어 교체)였다. 하락을 지우지 않고 남기는 이유는, 두 PASS-WITH-DEBT 만 남기면 이 SPEC 이 단조롭게 개선된 것으로 읽히기 때문이다.

- **iter6 채무 D16~D19 는 `625679ace` 에서 닫혔다**(문구 층 한정, AC/REQ id 불변).
- **D15 는 열려 있다** — `verdict.md` 의 마일스톤 분해 합이 실측 델타와 20 어긋난다. 어떤 AC 도 그 파일을 읽지 않으므로 이 카드에서 다루지 않았고, 채무로 남는다.

### Gaps — 관측하지 않은 것을 이름으로 적는다

- **허용 풀 `A_adm` 미측정.** 잔여 47,897 이 노력 부족인지 구조적 불가인지 가를 근거가 없다(`§AC-ALD2-001.3`).
- **승계 29 hunk 미분류.** 그중 셋(`3c3` · `5c5` · `21c21`)은 18개 허용 클래스 **밖**임이 확인됐고, 나머지 26 은 분류 자체를 하지 않았다.
- **`design.md §4.0` 열 재조정 미확정** — 구조적 하한 `F` 가 단일값이 아니라 세 값 구간 `{170,528 · 171,695 · 172,863}` 으로 남는다.
- **전체 테스트 스위트를 로컬에서 돌리지 않았다** — 상시 금지(`CLAUDE.local.md §4`, 2026-08-15 load 413 사고). 전 패키지 판정은 통합 브랜치 CI 몫이다.
- **AC-ALD2-005 의 트리 전수 최종 스윕 없음** — 앵커 스윕은 절이 움직이는 시점의 절 단위였고, 네 인용 형식(절 제목 · 파일+절 · 줄번호 · 산문 언급) 전체에 대한 최종 스윕은 돌지 않았다.
- ~~**AC-ALD2-008 재도출 불가** — 이 기록 실행에서 옮긴 절별 문자 합계를 다시 만들 수 없어 미검증으로 남긴다.~~ **[철회 — 재도출 실행 2026-09-26]** 근거가 틀렸다(기준선 ref 와 재배치 커밋이 그래프에 있으므로 절 본문은 사후 복원된다). `### AC-ALD2-008 재도출` 로 대체. 지우지 않고 남기는 이유는 **재도출 가능한 것을 불가로 적은 것 자체가 결함**이고, 그 결함이 보이지 않으면 다음 실행이 같은 이유로 또 닫기 때문이다.
- **AC-ALD2-006 Then 의 둘째 절이 2건에서 미확립** — `cross-session-messaging-detail` · `skill-routing-detail` 은 부모 밖 항목이 전부 형제 룰·독트린 문서다. 「그 companion 이 다루는 작업이 건드리는 도메인 경로」인지는 이 실행이 판정하지 않았고, 선례(`goal-directive-detail`)가 담은 작업 산출물 경로에 해당하는 항목은 없다.
- **AC-ALD2-008 의 절 경계는 제목 레벨로 잘랐다** — 제목 줄부터 같거나 상위 레벨 다음 제목 전까지. 문단 블록 이동 2건은 `git diff -U0` 훅의 행 범위로 잘랐고(정확한 경계), 그 둘이 판정을 가르는 자리였다는 사실을 본문에 적었다.
- **AC-ALD2-008 의 `## Cross-references` 크레딧은 과대 계상일 수 있다** — 기준선 stub 의 그 절을 통째로 크레딧했으나 stub 은 자기 `Cross-references` 를 유지한다. 두 건(`session-handoff-format` 1,549 · `kanban-dispatch-mechanics` 842)을 **0 으로 두고 다시 봐도 각각 여유 +611 · +3,408** 로 통과하므로 판정은 이 선택에 의존하지 않는다.
- **AC-ALD2-008 은 이 카드가 절을 옮겨 넣은 11개만 쟀다** — 유입 없는 companion 4개는 대상이 아니다(delta 최대 +64).
- **린트 미실행** — 이 카드의 Go 파일 기여가 0개(실측)여서 새 린트 표면이 없다는 판단이며, `golangci-lint` 를 이 실행에서 돌리지는 않았다. 판단이지 측정이 아니다.

### 프로세스 결함 1건 — 기록해 반복을 막는다

`f0893dc36`(`docs(…): repair plan-audit iter5 defects`)이 SPEC 산출물 5개와 함께 **구현 편집 1건**을 실었다 — `internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging-detail.md` 에서 빈 줄 1개 삭제(`git show --stat f0893dc36` 의 6번째 파일, `1 -`). 그 결과 **미러 수리가 문서 수리와 커밋 경계로 분리되지 않는다.** 착지를 막는 결함은 아니고, AC-ALD2-003 의 「수리로 달성」 순서를 이 절이 따로 적어야 하는 이유가 바로 그것이다. 다음에는 미러 수리를 자기 커밋으로 낸다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-26
run_commit_sha: 13e6adb1b               # run 기록 종결 커밋. 그 커밋 자신은 자기 해시를 인용할 수 없어 `pending-backfill-run` 이었고, 이 재측정 커밋이 backfill 했다
run_status: pass-with-debt              # FAIL 0 · UNVERIFIED 0. MUST-PASS AC-ALD2-005 가 PASS-WITH-DEBT 로 남아 `pass` 가 아니다
ac_total_count: 9
ac_pass_count: 7                        # AC-ALD2-001 · -002 · -003 · -004 · -006(재측정) · -008(재도출) · -009
ac_fail_count: 0                        # 종전 1건(AC-ALD2-006)은 좁힌 Given 재측정으로 PASS — 종전 FAIL 기록은 §E.2 에 보존
ac_pass_with_debt_count: 2              # AC-ALD2-005 (MUST-PASS) · AC-ALD2-007 (SHOULD-PASS)
ac_unverified_count: 0                  # 종전 1건(AC-ALD2-008)은 사후 재도출로 PASS — 종전 UNVERIFIED 근거는 §E.2 Gaps 에 철회 표기로 보존
remeasurement_ref: cb8466d09            # AC-ALD2-006 · -008 재측정 판정 ref (기준선 172ef22eb). 좁힘 커밋을 amend 하지 않아 순서가 커밋 그래프에 남는다
preserve_list_post_run_count: 0         # plan.md 는 §A.5 PRESERVE 목록을 선언하지 않는다(범위는 18파일 + companion + 템플릿 미러로 서술)
l44_pre_commit_fetch: not-run           # 워크트리 격리 레인; 커밋 직전 재판독은 `git rev-parse --short HEAD`(625679ace) + `git branch --show-current`(WT-rules-diet)로 수행
l44_post_push_fetch: not-run            # push 자체를 하지 않는다 — 레인 push 금지(운영자 지시 2026-09-01), develop 일괄 push 는 리드 소관
new_warnings_or_lints_introduced: unmeasured   # Go 기여 0개라 새 린트 표면 없음으로 판단; golangci-lint 는 이 실행에서 미실행
cross_platform_build:
  darwin_arm64: pass                    # make build → exit 0, catalog.yaml 내용 불변(git status 빈 출력)
  windows_amd64: pass                   # GOOS=windows GOARCH=amd64 go build ./... → exit 0, 출력 없음
  go_files_changed: 0                   # git diff --name-only a0b78213d HEAD -- '*.go' → 0
total_run_phase_files: 85               # git diff --name-only a0b78213d HEAD | wc -l
m1_to_mN_commit_strategy: >
  마일스톤 8개(M1·M2·M3·M4·M5·M6·M6.5·M7)를 커밋 4개에 묶었다 —
  3c40e0aa2(M1) · 4ab0511b4(M2) · 7bcdce760(M3+M6.5) · 276391646(M4·M5·M6·M7).
  이어 0acfa28e1(미러 수리) + SPEC 산출물 4개(14a5a3caf · f0893dc36 · 87fd59092 · 625679ace).
  이 기록 커밋이 draft → in-progress 전이를 뒤늦게 적용한다(M1 시점에 누락됐다).
evidence_paths:
  - .moai/reports/t1175/verdict.md              # run 판정서(이 절의 수치는 인용이 아니라 재측정)
  - .moai/reports/t1175/run-baseline-pre.md      # 기준선 B = 246,943 (ref 172ef22eb)
  - .moai/reports/t1175/pool-census.md           # 재배치 가능 풀 census
  - .moai/reports/t1175/plan-audit-iter5.md       # FAIL 0.71
  - .moai/reports/t1175/plan-audit-iter6.md       # PASS-WITH-DEBT 0.85
blockers: []                            # 종전 2건 모두 닫혔다 — 이력은 아래 closed_blockers 에 남긴다
closed_blockers:
  - id: BLK-ALD2-006
    ac: AC-ALD2-006
    summary: >
      이 카드가 수정한 companion 2개(verification-claim-integrity-detail.md ·
      session-handoff-examples.md)의 paths: 가 self-keyed 다. 두 값은 기준선 ref 172ef22eb 에서
      바이트 동일한 승계 값이며 이 카드가 만들지 않았다. AC-ALD2-006 은 AC-ALD2-003 이 가진
      card-created 한정어가 없어 문언대로 FAIL 한다.
    owner: manager-spec
    resolution: >
      cb8466d09 이 AC-ALD2-006 의 Given 을 REQ-ALD2-004 의 When(「한 절이 옮겨질 때」)에 맞춰
      좁혔고, 이 실행이 좁힌 대상 11개를 전수 재측정해 self-keyed 0건으로 PASS 를 얻었다.
      두 self-keyed 값은 **고쳐지지 않았다** — 범위에서 내려왔을 뿐이며
      acceptance.md §AC-ALD2-006.4 의 선존 결함 후속 후보로 남는다(발행은 운영자·리드 소관).
      좁힘의 정당성 판정은 sync phase 감사의 것이다.
  - id: BLK-ALD2-008
    ac: AC-ALD2-008
    summary: 옮긴 절별 문자 합계를 이 실행에서 재도출할 수 없어 MUST-PASS 가 미검증으로 남는다.
    owner: lead
    resolution: >
      근거가 틀렸다 — 기준선 ref 172ef22eb 와 재배치 커밋(7bcdce760 및 형제)이 모두 그래프에
      있으므로 절 본문은 사후에 복원된다. 절 제목을 기준선 stub 사본에서 조회해 11개 전수를
      재도출했고 11/11 통과. 제목으로 보이지 않는 문단 블록 이동 2건은 git diff -U0 훅의 행
      범위로 따로 쟀다(1,168자 · 1,424자) — 그 둘이 판정을 가르는 자리였다.
follow_up_candidates:
  - "self-keyed paths 2건 — core/verification-claim-integrity-detail.md · workflow/session-handoff-examples.md (선존, 이 카드가 만들지 않음)"
  - "AC-ALD2-006 Then 둘째 절 미확립 2건 — cross-session-messaging-detail · skill-routing-detail (부모 밖 항목이 전부 형제 문서)"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

🗿 MoAI
