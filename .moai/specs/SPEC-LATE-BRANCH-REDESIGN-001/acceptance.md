# SPEC-LATE-BRANCH-REDESIGN-001 — 인수 기준

> 모든 기준은 `develop @ 0cca34439` 에서 측정한 RED-now 를 짝으로 갖는다.
> 프로브는 영어 원문·식별자로만 잡는다. 각 회차는 양성 대조를 함께 관측한다.

## §A 측정 규약

- 명령은 **단일 호출**로 적고, 종료 코드는 별도 필드로 기록한다.
- 0 적중을 보고하는 프로브는 같은 회차에 **존재가 확인된 문자열**을 함께 잡아 관측기가 살아 있음을 보인다.
- 빈 스윕(`no test files` · `no tests to run` · 선택자 0건)은 통과로 읽지 않는다.

---

## AC-LBR-001 — SPEC 당 PR 1개 (REQ-LBR-001)

**Given** `spec-workflow.md` 두 사본
**When** 단계 당 PR 을 규정하는 문구를 찾을 때
**Then** 적중이 0 이고, SPEC 당 1개 PR 을 규정하는 문구가 적중한다.

| 셀 | 내용 |
|---|---|
| RED-now 명령 | `grep -c "per phase" .claude/rules/moai/workflow/spec-workflow.md` |
| RED-now 출력 | (M1 착수 시 측정해 기입) |
| RED 인 이유 | `:26` "opens a PR per phase" · `:47` "one squash commit per phase" 가 현재 살아 있다 |
| 양성 대조 | 같은 파일에서 `Route B` 가 적중(현재 8행) |
| 녹색 경로 | M1 이 두 문장을 SPEC 당 1개 서술로 바꾼다 |
| 통과 출력 | `per phase` 적중 0, 새 문구 적중 ≥1 |
| 트리 핀 | `0cca34439` |

---

## AC-LBR-002 — plan 단계 워크트리 진입 (REQ-LBR-002)

**Given** `spec-workflow.md` 두 사본과 `zone-registry.md` 의 `CONST-V3R5-027`
**When** plan 단계의 워크트리 금지 문구를 찾을 때
**Then** 본문과 등재 문언 **둘 다** 금지를 담지 않는다.

| 셀 | 내용 |
|---|---|
| RED-now 명령 | `grep -c "NO L2/L3 worktree at this step" .claude/rules/moai/workflow/spec-workflow.md` |
| RED-now 출력 | (M3 착수 시 기입) |
| RED 인 이유 | `:50` 본문과 `zone-registry.md` `CONST-V3R5-027` `clause:` 양쪽에 살아 있다 |
| 양성 대조 | 같은 파일에서 `Step 1 (plan)` 적중 |
| 녹색 경로 | M3 이 amend 로 등재를 개정하고 본문을 정렬한다 |
| 통과 출력 | 본문 적중 0 **그리고** 등재 `clause:` 적중 0 |
| 트리 핀 | `0cca34439` |

> **둘 다** 인 것이 핵심이다. 본문만 고치면 등재와 어긋난 채 남고, 그 어긋남은 어떤 빌드도 잡지 않는다.

---

## AC-LBR-003 — 폐기 조건 단일화 (REQ-LBR-003)

**Given** `worktree-integration.md` · `spec-workflow.md` · `zone-registry.md`
**When** 두-PR 전제 문구를 찾을 때
**Then** 세 곳 모두 적중이 0 이다.

| 셀 | 내용 |
|---|---|
| RED-now 명령 | `grep -rc "run PR AND sync PR" .claude/rules/moai/workflow/worktree-integration.md` |
| RED-now 출력 | (M2 착수 시 기입) |
| RED 인 이유 | `worktree-integration.md:45`(both run + sync PRs merge) · `:623`(Frozen disposal) · `spec-workflow.md:53` · `CONST-V3R5-028` `clause:` 에 살아 있다 |
| 양성 대조 | 같은 파일에서 `moai worktree done` 적중 |
| 녹색 경로 | M2 가 `CONST-V3R5-028` 을 개정하고 세 본문을 정렬한다 |
| 통과 출력 | 네 지점 모두 적중 0 |
| 트리 핀 | `0cca34439` |

---

## AC-LBR-004 — Step 3.3.5 은퇴 (REQ-LBR-004)

**Given** `delivery.md` 두 사본과 이를 참조하는 `quality-gates-context.md`
**When** Step 3.3.5 을 찾을 때
**Then** 정의와 참조가 **둘 다** 사라진다.

| 셀 | 내용 |
|---|---|
| RED-now 명령 | `grep -rn "Step 3.3.5" .claude/skills/moai/workflows/sync/` |
| RED-now 출력 | `delivery.md:344: #### Step 3.3.5: Return to Base Branch (Post-PR Cleanup)` 외 참조 |
| RED 인 이유 | 정의(`delivery.md:344`)와 참조(`quality-gates-context.md:100`)가 함께 살아 있다 |
| 양성 대조 | 같은 회차에 `Step 3.3` 적중(상위 단계는 남는다) |
| 녹색 경로 | M1 이 정의를 은퇴시키고 참조를 정렬한다 |
| 통과 출력 | `Step 3.3.5` 적중 0, 상위 단계 문구는 유지 |
| 트리 핀 | `0cca34439` |

> **한국어 프로브 금지.** 카드 문구 `기준 브랜치 복귀` 는 이 트리에서 0 적중이며, 그것을 통과로 읽으면 이 AC 는 수리 없이 초록이 된다. 프로브는 `Step 3.3.5` 또는 `Return to Base Branch` 로 잡는다.

---

## AC-LBR-005 — Frozen 개정의 정규 경로 (REQ-LBR-005)

**Given** `canary_gate: true` 인 등재 항목 2건
**When** 개정을 수행할 때
**Then** `moai constitution amend` 가 5단 게이트를 통과해 종료 코드 0 으로 끝나고, 그 로그가 증거로 남는다.

| 셀 | 내용 |
|---|---|
| RED-now | 개정 전 등재 `clause:` 축자(spec.md §D.4 에 기록됨) |
| RED 인 이유 | 두 항목이 개정 전 문언을 담고 있다 |
| 녹색 경로 | M2 · M3 이 각각 amend 를 수행한다 |
| 통과 출력 | amend 종료 코드 0 + FrozenGuard → Canary → ContradictionDetector → RateLimiter → HumanOversight 5단 통과 기록 |
| 트리 핀 | `0cca34439` |

**직접 편집은 실패로 간주한다.** `zone-registry.md` 를 손으로 고쳐 통과시키는 경로는 이 AC 를 만족하지 않는다 — 게이트를 거치지 않은 개정은 기록이 남지 않기 때문이다.

---

## AC-LBR-006 — 두 사본 동반 (REQ-LBR-006)

**Given** 대상 4파일의 로컬·템플릿 사본
**When** 변경 후 두 사본을 비교할 때
**Then** 이 SPEC 이 바꾼 문장에 한해 두 사본이 같은 방향이고, 남은 차이는 **의도된 분기**로 판정돼 기록된다.

| 셀 | 내용 |
|---|---|
| RED-now 명령 | `diff .claude/rules/moai/workflow/spec-workflow.md internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` |
| RED-now 출력 | (M5 착수 시 기입 — 착수 전 기준 차이) |
| 녹색 경로 | M1~M3 이 두 사본을 각각 편집하고 M5 가 diff 를 판정한다 |
| 통과 출력 | 이 SPEC 이 바꾼 문장에서 두 사본 일치. 남은 차이는 각각 의도 판정과 근거가 기록됨 |
| 트리 핀 | `0cca34439` |

`cp` 통째 미러는 실패로 간주한다 — 의도된 분기를 지우기 때문이다.

---

## AC-LBR-007 — 생성물 두 축 (M5)

**Given** 템플릿 사본이 변경된 상태
**When** 생성물 검사를 돌릴 때
**Then** **두 축 모두** 초록이다.

| 축 | 명령 | 통과 |
|---|---|---|
| 방출 (`.md` → `.codex` toml) | `make agents-emit-check` | 종료 코드 0 |
| 카탈로그 해시 | `go test -run 'TestCatalogHashParity\|TestManifestHashFormat' ./internal/spec/... ./internal/template/...` | 종료 코드 0, 빈 스윕 아님 |

> 한 축의 초록을 다른 축의 근거로 쓰지 않는다. t784 에서 `agents-emit-check` 가 초록인 채로 카탈로그 해시가 적색이었다.

**임베드 축은 이 SPEC 의 범위가 아니다** — `make build` 는 레인이 돌리지 않으므로, 빌드된 바이너리가 싣는 내용은 **Gap 으로 명시**하고 통과로 기록하지 않는다.

---

## §B 이 SPEC 이 통과로 기록하지 않는 것

| 항목 | 이유 |
|---|---|
| `REQ-WBG-011` | 소멸 확인(적중 0). 못 잰 것이 아니라 대상이 없다 — §F.1 에 근거 기록 |
| 임베드 축 | `make build` 금지로 미관측. 범위 밖과 초록은 다른 칸이다 |
| docs-site 4로케일 | 후속 문서 카드 소유 |

---

🗿 MoAI
