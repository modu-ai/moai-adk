---
id: SPEC-AGENT-PARALLEL-OPT-001
title: "Agent instruction diet + plan/run/sync parallelization maximization — Implementation Plan"
version: "0.6.0"
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: ".claude/agents/moai, .claude/skills/moai/workflows, .claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md, .claude/rules/moai/workflow, .claude/workflows, internal/template/templates, internal/template/internal_content_leak_test.go"
lifecycle: spec-anchored
tags: "agent-diet, parallelization, fan-out, write-concurrency, workflow-wiring, template-first"
tier: L
---

## §A Context

`spec.md` §A 참조. 본 문서는 구현 계획이며, 마일스톤은 **결정 가역성 내림차순**으로 배열한다 — 가장 바뀔 가능성이 높은 정책·설계 결정을 앞에 두고, 기계적 편집을 뒤에 둔다. 따라서 사람 리뷰는 M1~M2에 집중하면 된다.

v0.3.0에서 구 M1(write-concurrency 규칙 개정)이 철회되어 마일스톤은 M1~M5로 재번호되었다. 현재 M1은 구 M2(fan-out 배선 + 배포)다.

---

## §B 결정 사항 (RESOLVED — 사용자 승인 완료)

세 항목 모두 사용자 결정으로 해소되었다. 미해소 clarification 마커는 남아 있지 않다.

### B.1 D1 — dynamic workflow 스크립트 배포: **템플릿 미러(배포) 채택**

3개 fan-out 스크립트(`plan-research-fanout.js`, `sync-audit-4dim.js`, `codemaps-extract.js`)를 `internal/template/templates/.claude/workflows/`에 미러해 배포 사용자에게 전달한다. 배포 사용자도 병렬화 이득을 얻는 것이 목적이다.

**결정에 따른 필수 작업 7건** (Group 5 REQ-APO-069..073 + Group 6 REQ-APO-074..078):

1. **선행 AC supersession 명시** — `SPEC-DWF-CODEMAPS-PILOT-001`의 비배포 판정을 침묵 위반하지 않고 명시 superseded 처리. 정식 대상은 **`AC-DCP-010`**(`acceptance.md:79` / `progress.md:86`, grep 문구 `grep -r "codemaps-extract\|codemaps-pilot" internal/template/templates/` → nothing)이며 소유 요구사항은 **`REQ-DCP-009/010`**이다. 파일럿 SPEC은 `status: completed`이므로 그 아티팩트에 supersession 주석을 다는 작업은 **`manager-spec` 소관**이다(Status Transition Ownership Matrix — `manager-develop`/`manager-docs`는 SPEC body 수정 금지). 본 SPEC의 run-phase에서 해당 편집이 필요해지면 오케스트레이터 재위임 경로를 거친다.
2. **스크립트 §25 중립화** — 헤더·주석의 내부 토큰 제거(실행 로직 불변).
3. **중립성 스캐너 확장** — `leakTextExtensions`에 `.js` 추가(중립성 판정의 선행 조건).
4. **capability gate 존치** — 배포해도 gate는 제거하지 않는다.
5. **shipped rule 모순 해소 (D2)** — `dynamic-workflows.md` L80("MoAI does not ship any saved workflows by default; the user-owned `.claude/workflows/` directory is not template-managed")과 L131("the validated script lives in the local, user-owned `.claude/workflows/` directory (not template-managed…)")을 개정한다. 두 파일(로컬/템플릿)은 **0-diff 실측 확인**되었으므로 양쪽 동시 편집 + byte-parity 유지가 필수다. 미개정 시 배포 템플릿이 `.claude/workflows/*.js`를 담은 채 "그 디렉터리는 template-managed가 아니다"라고 선언하는 자기모순 상태가 된다.
6. **`.js` 헤더 자기모순 정정** — `plan-research-fanout.js:36` / `sync-audit-4dim.js:38`의 "user-owned workflows" 문구(실측 2건)를 정정한다.
7. **SSOT 방향 확정 (D12)** — 배포되는 3개 스크립트의 SSOT는 **템플릿**이다. 편집은 `internal/template/templates/.claude/workflows/`에서 시작하고 로컬은 파생본이다. 아울러 3개 스크립트가 user-owned 보존 집합 밖이므로 `moai update`가 **로컬 사본을 덮어쓴다**는 사실을 `dynamic-workflows.md` 개정문에 명시한다.

**전제 실측 정정 3건** (`spec.md` §F.8 — 코디네이터 지시의 전제와 다름):

- **가드 개정은 불필요하다.** `split_namespace_test.go`는 `.claude/workflows/*.js` 전체를 막지 않고 `hns-*`/`harness-*` **접두사만** 막는다. 3개 generic 스크립트는 애초에 차단 대상이 아니다. 따라서 "가드를 좁힌다"는 작업은 존재하지 않으며, 요구되는 것은 **차단 유효성 불변식 단언**(AC-APO-072b)이다. 존재하지 않는 차단을 없애려 가드를 손대면 dev-only 격리가 약화된다.
- **중립성 AC는 그대로 두면 공허하다.** leak 스캐너의 `leakTextExtensions`에 `.js`가 없어 스크립트를 **읽지도 않고** green이 된다. 실제 위반은 존재한다: `plan-research-fanout.js` L35-36/L54와 `sync-audit-4dim.js` L37-38/L42가 `REQ-ATR-*`/`AC-ATR-*`/`SPEC-FOO-001`을 보유하고, `codemaps-extract.js`는 이미 클린(0건). 그러므로 `.js` 확장자 추가가 **중립성 판정의 선행 조건**이며, 이것이 D1이 실제로 요구하는 유일한 Go 변경이다(예상과 반대 방향).
- **`moai update` 보존 계약 변경은 불필요하다.** `update/plan/plan.go`의 user-owned 판정도 `hns-`/`harness-` 접두사 기반이라 generic 3개는 자동으로 template-managed가 된다 — 배포 자산에 요구되는 정확한 의미다. `catalog.yaml` 등록도 불필요(배포는 generic FS walk).

### B.2 D2 — sync Phase 12 형태: **read-only drafter + 단일 적용자 확정**

5개 read-only drafter(CHANGELOG / README+docs-site / project-docs / SPEC-artifacts / codemaps)가 병렬로 내용을 생성하고, 단일 `manager-docs`가 순차 적용한다. disjoint-writer 변형은 **채택하지 않으며** 문서화된 향후 선택지로만 보존한다(`spec.md` §C).

**write-concurrency 규칙 무의존 (중요).** drafter가 전원 read-only이고 적용자가 단일이므로 동시 write가 발생하지 않는다. 따라서 Phase 12 재구조화는 현행 `[HARD]` 절대 금지형 write-concurrency 규칙을 **그대로 둔 채** 성립하며(REQ-APO-024b), §C가 후속 SPEC으로 이관한 규칙 완화의 진행 여부와 무관하다. 규칙 완화 축은 본 SPEC에서 철회되었으므로 어떤 마일스톤의 blocker도 아니다.

### B.3 D3 — SPEC-ID 자가검사 마커: **선행 측정 후 결정 (게이트화)**

지금 제거 여부를 결정하지 않는다. M3의 **첫 작업**으로 아래 게이트를 실행하고 그 출력을 `progress.md`에 verbatim 기록한 뒤, 결과에 따라 분기한다.

**게이트 명령 (판별형):**

```bash
grep -rn 'decomposition:' --include='*.go' --include='*.sh' --include='*.yaml' \
  internal/ .github/ .claude/hooks/
```

**기대 출력량: 0-5줄** (실측 baseline **0건**). 출력 전량을 `progress.md`에 verbatim 기록한다.

> **비판별형 금지.** v0.2.0이 지시한 `grep -rn "decomposition\|segment match trace" internal/ .github/ .claude/` 는 **12,133건**을 반환한다(실측) — `decomposition`이 리포 전반의 일반 어휘이기 때문이다(`run/task-decomposition.md`, "hierarchical decomposition" 등). 이 형태는 (a) verbatim 기록 자체가 불가능하고 (b) 0-branch가 도달 불가능해 게이트가 항상 ≥1로 귀결되므로 판별력이 없다. 판별형은 콜론 앵커(`decomposition:` — 마커의 실제 출력 형태) + 실행 가능 파일 확장자 한정 + 마커 정의처(`.claude/agents/moai/manager-spec.md`) 제외의 3중 조건으로 좁힌다.

| 게이트 결과 | 조치 | 적용 라인 상한 |
|---|---|---|
| 기계적 소비자 **0건** (예상) | 마커 강제 제거 + 주변 산문 축약 | `manager-spec.md` ≤ **230** (분기 A) |
| 기계적 소비자 **≥1건** | 마커의 **출력 계약을 보존**한 채 주변 산문(예시 표, AC sub-ID 혼동 표)만 축약 | `manager-spec.md` ≤ **250** (분기 B) |

어느 분기든 실행 가능한 Bash 정규식 검사는 존치한다. 게이트 출력 기록 없이 마커를 제거하는 것은 가정에 근거한 조치이며 금지된다.

**참고 (게이트 결과 해석 보조):** `manager-spec.md:166`은 마커가 "downstream grep verification을 가능하게 한다"고 주장하나, 그 주장이 인용하는 grep 자체가 마커 정의처를 가리키는 **자기참조**이며 실제 소비자는 실측되지 않는다. 게이트가 0을 반환하면 이 주장은 근거 없는 것으로 확정된다.

---

## §C Pre-flight (M1 착수 전 필수 실측)

run-phase 담당 에이전트는 아래를 **실행하고 출력을 인용**한 뒤 M1을 시작한다. 어느 항목이든 기대와 다르면 blocker report로 반환한다.

| # | 명령 | 기대 |
|---|---|---|
| 1 | `wc -l .claude/agents/moai/*.md` | 합계 2417 (baseline 재확인) |
| 2 | `diff .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md` | 0-diff (실측 확인됨 — M1 개정 후에도 유지되어야 함) |
| 3 | `diff .claude/skills/moai-workflow-testing/references/ internal/template/templates/.claude/skills/moai-workflow-testing/references/` | 기존 차이 목록 확정(D11 목적지 미러) |
| 4 | `grep -rn "plan-research-fanout\|sync-audit-4dim\|codemaps-extract" internal/template/templates/` | 0건 (배포 前 baseline — M2 이후 3건으로 전환) |
| 5 | `go test ./internal/template/...` | green (배포 前 baseline) |
| 6 | `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` | 병렬 세션 레이스 부재 확인 |

특히 #2·#3은 mirror byte-parity의 baseline이다. 기존 차이가 있다면 그것은 본 SPEC이 만든 것이 아님을 먼저 고정해야 AC-APO-060 / AC-APO-075 판정이 유효해진다.

D3 게이트 grep은 Pre-flight가 아니라 **M3의 첫 작업**이다(§B.3) — 결과가 M3 작업 범위를 분기시키므로 M3 착수 시점에 실행하고 기록한다.

---

## §D 제약

- **Template-First**: 미러 존재 파일은 `internal/template/templates/` 먼저 → `make build` → 로컬 동기화. 역방향 편집 금지.
- **템플릿 중립성**: 미러 산출물에 SPEC ID / REQ 토큰 / 내부 날짜 / commit SHA 금지.
- **평면 계층**: 모든 fan-out은 오케스트레이터 launch. subagent nesting 의존 금지.
- **동시 spawn 상한**: Mode 4 기준 3-5 concurrent `Agent()`.
- **게이트 불변**: DP1 / Implementation Kickoff Approval / `gate-sync-1` / `gate-sync-2` 보존.
- **verdict 소유권 불변**: plan-auditor / sync-auditor.
- **커밋 규약**: plan-phase 산출물 커밋은 `feat(SPEC-AGENT-PARALLEL-OPT-001): ...` 접두사 사용(`docs(` 금지 — status 오분류 유발).
- **브랜치**: 본 리포는 all-tier PR 경유(main direct push 차단).

---

## §E 자가 검증

각 마일스톤 종료 시 담당 에이전트는 아래를 단일 턴 다중 Bash 배치로 실행하고 exit code + bounded tail을 인용한다.

```
1. wc -l .claude/agents/moai/*.md                       # 라인 상한
2. diff <dynamic-workflows.md pair>                     # shipped rule byte-parity
3. grep -rn "<3 script names>" .claude/skills/moai/     # zero-orphan
4. ls internal/template/templates/.claude/workflows/       # 배포 3개 존재 + hns-/harness- 0개
5. for f in <mirrored pairs>; do diff local template; done   # byte-parity
6. go test ./internal/template/...                      # 가드 테스트(leak + split_namespace)
7. go test ./...                                        # 전체
```

M2 이후에는 항목 4가 "3개 존재 AND `hns-*`/`harness-*` 0개" 이중 판정이 된다 — 배포 확인과 격리 유지를 한 명령으로 본다.

---

## §F 마일스톤 (결정 가역성 내림차순)

### M1 — fan-out 배선 + 스크립트 배포 (D1 확정)

**대상**: `plan.md`, `run.md`, `sync.md`, `codemaps.md` + 하위 스킬 문서 + 템플릿 미러; `internal/template/templates/.claude/workflows/`(신규 3파일); `internal/template/internal_content_leak_test.go`; `SPEC-DWF-CODEMAPS-PILOT-001` 아티팩트.

**작업 순서** (2→3→4가 순서 의존 — 스캐너 확장이 중립화 판정의 선행 조건):

1. `leakTextExtensions`에 `".js": true` 추가 (REQ-APO-072). 이 단계 없이 진행하면 이후 중립성 판정이 전부 공허해진다.
2. **템플릿 원본 생성 (SSOT 우선)** — 3개 스크립트를 `internal/template/templates/.claude/workflows/`에 **먼저 생성**하고 그 자리에서 §25 중립화한다: `plan-research-fanout.js` L35-36/L54와 `sync-audit-4dim.js` L37-38의 `REQ-ATR-*`/`AC-ATR-*`/`design.md §C/§D`/`acceptance.md` 참조 제거 및 일반화, 두 파일의 "user-owned workflows" 문구 정정(REQ-APO-076). `codemaps-extract.js`는 이미 클린(실측 0건) — 무수정 이관. **실행 로직 불변**. 참고: `sync-audit-4dim.js:42`의 `spec_id: "SPEC-FOO-001"`은 CI 클래스 미매치(`spec.md` §F.8.3-a)이며 일반 플레이스홀더이므로 제거 대상이 아니다.
3. `make build` → 로컬 `.claude/workflows/`를 템플릿 산출물로 동기화. **로컬을 먼저 고쳐 템플릿에 복사하는 방향은 금지**(§G anti-pattern) — 로컬은 파생본이다.
4. RED/GREEN 왕복 검증 — 미중립 버전을 잠시 심어 leak 테스트 FAIL 관측 후 중립 버전으로 복구해 PASS 관측(스캐너가 실제로 `.js`를 읽었다는 증거).
5. `TestSplitHarnessNamespaceNoLeak` 통과 확인 + 차단 유효성 확인(`hns-release-update-run.js`를 임시로 심어 FAIL 관측 후 제거). **가드 소스는 수정하지 않는다** — 이미 prefix-scoped이다.
6. `SPEC-DWF-CODEMAPS-PILOT-001`에 supersession 주석 추가 — 대상은 **`AC-DCP-010`** + **`REQ-DCP-009/010`**, 정식 grep 문구 `grep -r "codemaps-extract\|codemaps-pilot" internal/template/templates/` → nothing 이 무효화됨을 명시. 파일럿은 `status: completed`이므로 이 편집은 `manager-spec` 소관이며, run-phase에서 다른 에이전트가 만나면 blocker report 후 오케스트레이터 재위임 경로를 탄다. 본 SPEC frontmatter `partially_supersedes`와 상호 참조 성립.
6b. **shipped rule 개정 (D2)** — `internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md` L80/L131을 먼저 개정 → `make build` → 로컬 미러 동기화 → `diff` 0-diff 확인. 개정문은 (a) MoAI-shipped generic fan-out과 user-owned `hns-*`/`harness-*`를 구분하고, (b) `moai update`가 generic 3개의 로컬 사본을 덮어쓴다는 사실을 명시한다.
7. 워크플로우 문서 배선: `plan-research-fanout.js`(plan Phase 2+6 통합), `sync-audit-4dim.js`(run Phase 13/16/17 + sync Phase 7), `codemaps-extract.js`(codemaps high-count 증강, pilot 스코핑 유지).
8. 모든 참조를 capability-gate 형태로 작성 — gate 조건은 **파일 존재 AND 런타임 dynamic-workflow 지원** 두 가지. 어느 하나라도 부재 시 기존 단일 에이전트 경로 fallback, 오류·경고 없음.
9. verdict 소유권 문장 삽입: 스크립트는 증거 수집 수단이며 verdict는 auditor 소유.
10. docs-site 4-locale "실제 파이프라인 투입" 주장 — 배선+배포 완료로 **참이 되었음을 검증**(AC-APO-015 + AC-APO-069 동시 PASS). 4-locale 동시 처리.

**리스크**: 작업 1을 건너뛰면 중립성 green이 공허해져 내부 토큰이 사용자에게 배포된다. 작업 6b를 건너뛰면 배포 템플릿이 자기모순(스크립트를 담은 채 "template-managed가 아니다"라고 선언) 상태로 출하된다. `moai update` 보존 계약과 catalog 등록은 변경 불필요(실측 확인).

**REQ**: 010-016, 069-078. **AC**: 010-016, 069-078, 072b.

### M2 — read-only fan-out + single-writer 재구조화 [설계 결정 다수]

**대상**: `plan/spec-assembly.md`, `run/task-decomposition.md`, `run/phase-execution.md`, `sync/doc-execution.md`, `sync/quality-gates-quality.md`, `sync/quality-gates-context.md` + 템플릿 미러.

**작업**:
1. plan Phase 11 — 복수 read-only 심사 렌즈 병렬 + plan-auditor 단일 verdict.
2. plan Phase 10 — 단일 writer 유지 + 산출물 전량 단일 턴 병렬 Write 명시.
3. run RED 단계 — 테스트 초안 drafter pool + 단일 manager-develop 적용.
4. run Phase 13/16/17 — 3회 직렬 감사를 1회 병렬 증거 + 1회 verdict로 축약(최대 3회 반복 상한 보존).
5. sync Phase 12 — **D2 확정 설계**: 5개 read-only drafter(CHANGELOG / README+docs-site / project-docs / SPEC-artifacts / codemaps) 병렬 + 단일 manager-docs 순차 적용. disjoint-writer 변형 서술은 넣지 않는다. **현행 write-concurrency 규칙과 독립임을 문서에 명시** — drafter 전원 read-only + 단일 적용자이므로 동시 write가 없다(REQ-APO-024b). 데이터 흐름은 `design.md` §B 참조.
6. sync Phase 10 — 패키지별 테스트 생성 fan-out.
7. sync Phase 1/7 — run Phase 15 `moai verify` 스냅샷 소비(신선도 규칙 보존).
8. MX 스캔 샤딩(run Phase 9/18, sync Phase 9).
9. 게이트·AskUserQuestion 경계 보존 문장을 각 변경 지점에 유지.

**리스크**: 게이트 문장을 실수로 이동·삭제하면 [HARD] 위반. 각 편집 후 게이트 토큰 grep으로 존재를 재확인한다.

**write-concurrency 규칙 무의존**: 본 마일스톤의 모든 항목은 read-only fan-out + 단일 writer로 설계되어 write 동시성을 요구하지 않는다. 현행 `[HARD]` 절대 금지형 규칙을 **그대로 둔 채** 성립하며, §C가 후속 SPEC으로 이관한 규칙 완화의 진행 여부와 무관하다(REQ-APO-024b).

**REQ**: 020-030 (024b 포함). **AC**: 020-030 (024b 포함).

### M3 — 본문 다이어트: SSOT 중복 및 금지 스캐폴딩 제거

**대상**: `plan-auditor.md`, `manager-spec.md`, `manager-develop.md` + 템플릿 미러.

**작업**:

0. **[게이트 — M3 첫 작업, D3]** 아래 **판별형** 명령을 실행하고 출력을 `progress.md`에 **verbatim 기록**한다.

   ```bash
   grep -rn 'decomposition:' --include='*.go' --include='*.sh' --include='*.yaml' \
     internal/ .github/ .claude/hooks/
   ```

   **기대 출력량 0-5줄**(plan-phase 참고 실측 0건). 이 범위를 넘으면 명령이 잘못 좁혀진 것이므로 재작성한다. **비판별형 금지** — `grep -rn "decomposition\|segment match trace" internal/ .github/ .claude/` 는 12,133건을 반환해 verbatim 기록이 불가능하고 0-branch가 도달 불가능하다(§B.3). 결과에 따라 작업 2의 범위와 `manager-spec.md` 적용 라인 상한이 분기된다. 이 게이트 없이 마커를 건드리는 것은 금지된다.
1. plan-auditor — frontmatter 스키마 열거 1회화, M6 Chain-of-Verification 절 및 보고 템플릿 CoVe 섹션 제거.
1b. **plan-auditor 잔여 — 파일 내 재진술 제거**(상한 430 재보정에 대응, `spec.md` §D.2.1). 작업 1의 렌즈는 *파일 간* SSOT 중복이었고 잔여 구간은 *파일 내* 동일 규칙 반복이다. Group 7/8이 각 차원을 5중 진술(① 요약 불릿 ② 유래 산문 ③ 번호 체크 ④ bash 검증 동사 ⑤ Severity rubric)하는 것에서 **①②⑤만 제거하고 ③④는 존치**한다 — ③④가 조작적 내용이므로 이를 건드리면 REQ-APO-068 위반이다. 추가로 AP-VEM-003(`verification-batch-pattern.md` AP-VBP-002 중복), Delegation Note 말미의 `NOT for:` 재진술, Invocation Examples 3→1, Output Format 2-stream 산문을 정리한다. 제거 가능 28행 중 24행 필요. **공백 전용 압축 금지**(§D.2.1 [HARD]).
2. manager-spec — frontmatter 스키마 블록을 SSOT 교차참조로 대체, Step 5 중복 제거, GEARS/EARS 표를 스킬 교차참조로 대체, 산출물 개수 서술 정정. **잔여 6행**: § Status Responsibility Matrix의 1행 표가 § SPEC Artifact Ownership "Status transitions owned"와 동일 사실을 진술하므로 두 구간을 병합한다(상한 230 유지 — `spec.md` §D.2.1). **SPEC-ID 자가검사 블록은 작업 0의 결과에 따라**: 소비자 0건 → 마커 강제 제거 + 축약; 소비자 ≥1건 → **출력 계약 보존** + 주변 산문(예시 표, AC sub-ID 혼동 표)만 축약. 두 분기 모두 실행 Bash 검사는 존치.
3. manager-develop — DDD/TDD 동형 워크플로우 통합, "one atomic change" 제약을 패키지 내부로 한정.
4. 제거된 각 블록 자리에 SSOT 교차참조 삽입(정보 무성 소실 금지).

**상한 분기 (D4 reconciliation)**: 작업 0의 결과가 소비자 ≥1건이면 `manager-spec.md`에 적용되는 상한은 **분기 B(250)**로 전환된다(`spec.md` §D.2). 이는 "MUST 상한 미달성 + 사유 기록"이 아니라 **분기 조건부 상한의 정상 적용**이다 — AC-APO-055는 어느 분기든 MUST로 판정되며, 적용 상한만 게이트 결과에 따라 결정된다. 마커 계약을 깨서 분기 A 상한을 억지로 맞추는 것은 금지된다.

**REQ**: 040-048, 068. **AC**: 040-048, 068.

### M4 — 본문 다이어트: 구조 분리 및 병렬 규범 확산

**대상**: `sync-auditor.md`, `manager-docs.md`, `e2e-tester.md`, `manager-git.md`, `builder-harness.md`, 신규/기존 스킬 레퍼런스 + 템플릿 미러.

**작업**:
1. sync-auditor — scoring model 1개·report template 1개로 축약.
2. manager-docs — Nextra/WCAG/page-speed 레거시 제거(실제 소유 범위와 무관).
3. e2e-tester — 비-호스트 OS 레시피를 **`.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md`** 로 이관(REQ-APO-078). 해당 디렉터리는 템플릿 미러가 존재하므로 Template-First 순서(템플릿 먼저 → `make build` → 로컬) + byte-parity 의무가 적용된다. `e2e-tester.md`는 해당 경로를 Conditional Skill Loading 절에서 참조한다.
4. manager-git — merge_method 해석 1회화.
5. builder-harness — model-policy 재진술 및 stale 마이그레이션 안내 제거.
6. 다중 검증을 수행하는 모든 retained 에이전트에 병렬 배칭 교차참조 1줄 삽입(현재 1/10 → 목표 전량).
7. 라인 상한 판정(`spec.md` §D.2).

**REQ**: 049-055, 068. **AC**: 049-055.

### M5 — Template-First 동기화 및 전량 검증 [가장 기계적]

**작업**:
1. 모든 편집이 `internal/template/templates/` 우선이었는지 감사, `make build` 실행.
2. mirror byte-parity 전량 확인.
3. 템플릿 중립성 grep(SPEC ID / REQ 토큰 / 날짜 / SHA 부재) — 배포된 3개 `.js` 포함.
4. harness Runner 비배포 불변식 grep(`hns-*` / `harness-*` 부재) + 배포 존재 grep(generic 3개 존재).
5. `go test ./...`, template-neutrality CI guard, `split_namespace_test.go`, `internal_content_leak_test.go`.
6. archived agent 이름 부재 grep.
7. frontmatter 무변경 확인(에이전트 `.md` frontmatter 블록 `git diff` 0줄).
8. supersession 상호 참조 성립 확인(`AC-DCP-010` + `REQ-DCP-009/010` ID 인용 포함).
9. shipped rule 개정 확인 — `dynamic-workflows.md` 로컬/템플릿 0-diff + 모순 문구 부재.
10. e2e 목적지 실재 + 미러 0-diff 확인.

**REQ**: 060-067, 069-078. **AC**: 060-067, 069-078.

---

## §G Anti-Patterns

- **미러 역방향 편집** — 로컬 `.claude/`를 먼저 고치고 템플릿에 복사. §2 Template-First 위반이며 중립성 검사를 통과한 적 없는 내용이 유입된다.
- **capability gate 없는 무조건 참조** — 템플릿 미러 문서가 비배포 `.js`를 무조건 참조하면 배포 사용자에게 dead reference.
- **verdict 소유권 이전** — 스크립트 출력의 harmonic mean을 그대로 최종 verdict로 채택. REQ-APO-013 위반.
- **게이트 문장 소실** — 단계 재구조화 중 HUMAN GATE 서술이 함께 삭제/이동.
- **정보 무성 소실** — 중복 블록을 지우면서 교차참조를 남기지 않아 SSOT 도달 경로가 끊김.
- **라인 수만 맞추는 압축** — 의미 있는 제약을 지워 상한을 충족. AC는 라인 수와 **동시에** 교차참조 존재를 요구한다.
- **frontmatter 동시 튜닝** — 본문 작업 중 `model`/`effort`를 함께 바꿈. REQ-APO-065 위반이며 프로파일 축과 충돌.
- **근사 카운트** — "약 N줄 감소" 같은 서술. 실측 `wc -l` 출력 인용만 유효하다.
- **공허 green (vacuous pass)** — 스캐너가 대상 파일 확장자를 읽지 않는 상태에서 "가드 green"을 중립성 근거로 인용. 스캐너가 실제로 읽었다는 RED/GREEN 왕복 증거가 없으면 판정은 무효다.
- **존재하지 않는 차단을 제거** — `split_namespace_test.go`가 generic fan-out을 막는다고 가정하고 가드를 완화. 실측상 이미 prefix-scoped이며, 손대면 dev-only 격리만 약화된다.
- **선행 SPEC AC 침묵 위반** — DWF-CODEMAPS-PILOT-001의 비배포 판정을 supersession 기록 없이 어김. 상호 참조가 성립해야 한다.
- **D3 게이트 생략** — grep 출력 기록 없이 마커 제거 여부를 판단. 가정 기반 조치이며 금지된다.
- **로컬 선편집 후 템플릿 복사 (SSOT 역방향)** — 배포되는 `.js`와 `dynamic-workflows.md`는 템플릿이 SSOT다. 로컬 라인 번호를 근거로 로컬을 먼저 고치면 §2 Template-First 위반이며, `moai update` 시 로컬 수정이 소실된다.
- **비판별형 게이트 grep 사용** — `decomposition\|segment match trace` 전역 grep은 12,133건을 반환해 verbatim 기록이 불가능하고 0-branch가 도달 불가능하다. 판별형(콜론 앵커 + 확장자 한정 + 정의처 제외)만 유효하다.
- **write-concurrency 규칙을 본 SPEC에서 손대기** — 구 Group 1(write-concurrency)은 철회되었다. `agent-common-protocol.md` / `CLAUDE.md`의 write-concurrency 문장은 `git diff` 0줄이어야 한다.
- **자기모순 배포** — `.claude/workflows/*.js`를 템플릿에 넣으면서 `dynamic-workflows.md`의 "not template-managed" 서술을 그대로 두는 것.

---

## §H Cross-References

- `spec.md` §B(요구사항), §D.2(라인 상한), §F(Ground Truth).
- `acceptance.md` §D(AC 매트릭스 및 판정 명령).
- `.claude/rules/moai/core/agent-common-protocol.md` — M1 대상.
- `.claude/rules/moai/development/agent-authoring.md` § Prompt Craft — M4 근거.
- `CLAUDE.local.md` §2 / §25 — Template-First / 중립성.
