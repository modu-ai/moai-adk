---
id: SPEC-AGENT-PARALLEL-OPT-001
title: "Agent instruction diet + plan/run/sync parallelization maximization — Research"
version: "0.2.0"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: ".claude/agents/moai, .claude/skills/moai/workflows, .claude/rules/moai/core, internal/template/templates"
lifecycle: spec-anchored
tags: "agent-diet, parallelization, fan-out, write-concurrency, workflow-wiring, template-first"
tier: L
---

## §A 조사 범위와 방법

두 차례의 read-only 코드베이스 스윕(에이전트 본문 축 / 워크플로우 병목 축)에서 수집한 관찰을, 본 plan-phase 세션에서 **전량 재실측**했다. 아래 모든 수치·경로·라인 번호는 본 세션에서 실행한 명령의 출력에 근거한다. 재실측 과정에서 브리프 대비 2건의 사실 정정이 발생했으며 §E에 별도 기록한다.

검증에 사용한 명령 계열: `wc -l`, `grep -rn`, `grep -c`, `grep -l`, `ls -la`, `diff` 대상 인벤토리, `sed -n` 구간 확인.

---

## §B 축 1 — 에이전트 본문 과복잡도

### B.1 라인 수 분포 (실측)

```
505 plan-auditor.md
317 manager-spec.md
311 manager-develop.md
221 sync-auditor.md
211 manager-git.md
201 manager-design.md
195 builder-harness.md
182 e2e-tester.md
167 manager-docs.md
107 super-advisor.md
─────
2417 total
```

상위 3개 파일이 전체의 **47%**(1,133/2,417)를 차지한다. 이 3개가 다이어트 효과의 대부분을 결정한다.

### B.2 유형별 과복잡도 관찰

#### (1) 선언된 SSOT의 본문 중복

| 위치 | 관찰 | SSOT |
|---|---|---|
| `plan-auditor.md` L132 (MP-3) | 12개 필드를 타입·enum까지 전문 열거 | `spec-frontmatter-schema.md` |
| `plan-auditor.md` L248-261 (FC-1..FC-12) | **같은 파일 안에서** 12개 필드를 다시 열거 | 동일 |
| `manager-spec.md` L182-231 | 12-field 스키마 YAML 블록 + 검증 규칙 재기술 | 동일 |

세 위치 모두 본문 안에서 `spec-frontmatter-schema.md`를 SSOT라고 **명시적으로 선언한 뒤** 그 내용을 복제한다. `plan-auditor.md` L248은 심지어 "Verify against the canonical 12-field schema in ... (the SSOT)"로 시작한 직후 12줄을 열거한다. 자기모순적 패턴이며, 스키마 변경 시 3곳이 동시에 드리프트할 수 있는 구조다.

`builder-harness.md`도 같은 계열이다: L156과 L183에 동일한 `Model/effort escalation` 문장이 두 번 등장하고, L162-167에서 `model-policy.md`의 표를 재진술한다.

#### (2) 금지된 방어적 스캐폴딩

`agent-authoring.md` L215는 다음을 명시 금지한다:

> Do NOT add Opus 4.6-era defensive scaffolding ("double-check N times", "verify before returning", "explicitly confirm before proceeding") — counterproductive given literal instruction following.

`plan-auditor.md`는 이 금지에 정면으로 저촉되는 구조를 3곳에 보유한다: L142 `### M6: Chain-of-Verification`, L153 second-pass 결과 기록 지시, L415 보고 템플릿의 `## Chain-of-Verification Pass` 섹션. 리터럴 지시 수행 모델에서 2-pass 재검증 요구는 토큰만 증가시키고 판정 품질에 기여하지 않는다.

`manager-spec.md` L145-180의 SPEC-ID 자가검사 프로토콜도 유사 계열이다. 실행 가능한 Bash 정규식 검사(L153-159)는 실질 가드로서 유효하지만, 그 주변의 4단계 의례(decomposition 출력 마커 강제, 예시 출력 표, AC sub-ID 혼동 방지 표, PASS/FAIL 라인엔드 마커 규정)가 약 36줄을 소비한다. 추가로 L235 Step 5 체크리스트가 같은 제약을 **세 번째로** 재진술한다.

#### (3) 동형 워크플로우의 전문 중복

`manager-develop.md`(311줄)는 DDD(ANALYZE-PRESERVE-IMPROVE)와 TDD(RED-GREEN-REFACTOR) 두 사이클을 각각 전문 기술한 뒤, 후속 3개 섹션에서 다시 모드별로 분기 서술한다. 두 사이클은 "현 상태 파악 → 변경 → 검증"이라는 동일 골격을 공유하므로 공통 골격 1회 + 모드 차이 서술로 압축 가능하다.

같은 파일이 "one atomic change at a time"을 **전역 제약**으로 선언한다. 이 문장은 서로 의존하지 않는 별개 패키지에 대한 작업까지 직렬화하도록 읽히며, 축 2의 병렬화 목표와 직접 충돌한다.

#### (4) 이중 모델 / 이중 템플릿

`sync-auditor.md` L44-46: "Two scoring models exist; the selection rule is normative". 결과적으로 L64-88(평면 모델 report)과 L130-203(HRN-003 계층 모델 + 별도 report)이 공존해 약 90줄을 소비한다. 두 모델의 선택 규칙 자체도 본문에 존재해야 하므로 3중 비용이다.

#### (5) 소유 범위와 무관한 레거시

`manager-docs.md`는 실제로 CHANGELOG / README / docs-site / frontmatter 전이를 소유하는 sync-phase 에이전트다. 그러나 본문 L35-103은 Nextra 프레임워크 설정(`theme.config.tsx`, `next.config.js`, MDX, i18n, SSG), WCAG 2.1 접근성 준수, "Accessibility score > 95%", 모바일 반응형 테스트를 지시한다. 같은 파일이 서로 다른 두 에이전트를 기술하고 있다.

`e2e-tester.md`는 macOS(axcli) / Windows(FlaUI) / Linux(dogtail) 3개 OS 레시피를 모두 본문에 보유한다(L87-114). 실행 중인 세션의 호스트 OS는 하나이므로 최소 2/3은 매 spawn마다 도달 불가능한 컨텍스트다.

`manager-git.md`는 `merge_method` 해석 규칙을 L125 / L163 / L191 세 곳에서 반복한다.

### B.3 근원 — 병렬 규범의 미도달

가장 중요한 관찰이다. `verification-batch-pattern.md` 또는 `agent-common-protocol.md § Parallel Execution`을 참조하는 에이전트 본문은 **`plan-auditor.md` 단 1개**다(`grep -l` 실측, 10개 중 1개). 나머지 9개 에이전트는 다중 검증을 수행하면서도 단일 턴 다중 Bash 배칭 규범에 대한 참조를 갖지 않는다.

즉 병렬화 규범은 규칙 계층에 존재하지만 **실행 계층(에이전트 본문)에 배선되지 않았다**. 이것이 축 1과 축 2를 잇는 공통 근원이며, REQ-APO-054가 이를 직접 겨냥한다.

---

## §C 축 2 — 워크플로우 병목

### C.1 fan-out 자산 인벤토리 (실측)

`.claude/workflows/` 실재 파일 5개:

| 파일 | 크기 | 성격 |
|---|---:|---|
| `plan-research-fanout.js` | 8,914 B | read-only 렌즈 explorer 병렬 + synthesizer, 파일 미기록(문자열 반환) |
| `sync-audit-4dim.js` | 11,065 B | 4차원 read-only judge 병렬 + in-script harmonic mean |
| `codemaps-extract.js` | 3,629 B | 패키지별 Explore fan-out(`effort: 'low'` 고정) |
| `hns-oss-docs-run.js` | 10,530 B | user-owned harness Runner (본 SPEC 범위 밖) |
| `hns-release-update-run.js` | 5,267 B | dev-only harness Runner (범위 밖) |

앞의 3개가 본 SPEC의 배선 대상이다.

### C.2 배선 부재 실측

`.claude/skills/moai/` 하위 전체에서 3개 스크립트명을 검색한 결과, `plan.md` / `run.md` / `sync.md` / `codemaps.md` 및 그 하위 스킬 문서 어디에서도 **매치 0건**이다. (매치된 것은 `harness-builder.md` L81의 `codemaps-extract.js` 선례 인용 1건뿐이며, 이는 호출이 아니라 harness Runner 작성 지침의 참조다.)

### C.3 단계별 병목 (실측 phase 번호)

**plan**
- Phase 2 Project Exploration + Phase 6 Deep Research — 둘 다 단일 Explore. `plan-research-fanout.js`가 정확히 이 형태를 대체하도록 작성되어 있다.
- Phase 10 SPEC Document Creation — 단일 writer. 산출물이 복수이므로 단일 턴 병렬 Write가 적정. `manager-spec.md` L120이 "3-file creation"이라 서술하나 실제 열거는 spec/plan/acceptance/progress 4개다(불일치).
- Phase 11 Independent SPEC Review — 단일 plan-auditor + 재시도 루프.

**run**
- Phase 13 Quality Validation → Phase 16 Active Quality Evaluation(sync-auditor) → Phase 17 TRUST 5 Static Verification(sync-auditor). 3회 직렬 감사 패스이며 각각 반복 상한을 가진다.
- Phase 15 Pre-Review Quality Gate — **참조 GOOD 패턴**. 단일 턴 다중 Bash 배치 + 파일 리다이렉트 계약 + `moai verify` 스냅샷 record/consume을 모두 갖춘 유일한 단계다. 다른 단계가 따라야 할 형태가 이미 리포 안에 존재한다.
- Phase 8 File Structure Scaffolding / Phase 9 Pre-Implementation MX Context Scan — 파일 단위 순회.

**sync**
- Phase 12 Execute Document Synchronization — **최대 병목**. Step 2.1 백업, Step 2.2 CHANGELOG·README, Step 2.2.5 structure/tech/product.md + codemaps 재생성, Step 2.3 post-sync 품질, Step 2.4 SPEC status, Step 2.4.1 GitHub issue sync. 산출물 대부분이 서로 **경로 disjoint**인데 단일 `manager-docs`가 직렬 기록한다.
- Phase 10 Coverage Analysis — 패키지별 독립 작업을 단일 에이전트로 깔때기.
- Phase 1(`gate-sync-1`) / Phase 7 Quality Check — run Phase 15가 이미 기록한 스냅샷을 소비하지 않고 재실행.

### C.4 정책 차단 요인 (실측 3표면)

| 표면 | 라인 | 형태 |
|---|---|---|
| `agent-common-protocol.md` | 191 / 193 / 198 | 절대 금지 |
| `CLAUDE.md` | 250 | 절대 금지 |
| `internal/template/templates/CLAUDE.md` | 250 | 절대 금지(byte-parity) |
| `.claude/skills/moai/workflows/e2e.md` | 251 | **스코프 한정** — "never run concurrently on overlapping scope" |

`agent-common-protocol.md` L193은 안전장치의 의도를 스스로 정확히 서술한다: "This targets the actual hazard — a **file-write race** between agents". 실제 위험이 파일 쓰기 레이스라면 방어선은 "동시 실행 금지"가 아니라 "겹치는 경로에 대한 동시 쓰기 금지"여야 한다. `e2e.md`가 이미 그 형태를 쓰고 있으므로, 개정은 새 정책 도입이 아니라 **기존 두 표현의 정합화**다.

---

## §D Template-First 스코프 실측

| 경로 | 미러 |
|---|---|
| `.claude/agents/moai` | 존재(10 파일) |
| `.claude/skills/moai/workflows`(+ `plan`/`run`/`sync` 하위) | 존재 |
| `.claude/rules/moai/core` | 존재 |
| `.claude/rules/moai/workflow` | 존재 |
| `CLAUDE.md` | 존재 |
| `.claude/workflows/` | **부재 (조사 시점)** → D1 결정으로 M2에서 generic fan-out 3개만 신설 |

조사 시점의 이 비대칭이 본 SPEC의 핵심 설계 제약이었다. 배선 대상 문서는 배포되는데 배선되는 자산은 배포되지 않는 상태였다. `internal/template/split_namespace_test.go` L93-104는 `.claude/workflows/` 하위 **dev-only harness Runner**의 템플릿 유입을 접두사 기반으로 차단하며(§H.1 — generic fan-out은 애초에 대상 아님), SPEC-DWF-CODEMAPS-PILOT-001은 `codemaps-extract.js`의 비배포를 명시 AC(`grep -r "codemaps-extract" internal/template/templates/` → nothing)로 고정했다.

**D1 결정으로 비대칭은 해소된다** — 3개 generic fan-out을 배포하고 선행 SPEC의 비배포 AC를 명시 supersede한다(REQ-APO-070). harness Runner(`hns-*` / `harness-*`)의 비배포는 그대로 유지된다(REQ-APO-062). capability-gate는 배포 이후에도 존치한다 — 파일 존재가 런타임 지원을 보장하지 않기 때문이다(§H.4).

---

## §E 브리프 대비 실측 정정

### E.1 정정 1 — "zero references"는 정확하지 않다

브리프는 3개 스크립트가 "ORPHANED (zero references from plan.md/run.md/sync.md, grep-verified)"라고 기술했다. 워크플로우 문서 기준으로는 정확하다. 그러나 리포 전체 grep에서 다음이 드러났다:

- `docs-site/content/{en,ko,ja,zh}/claude-code/agentic/workflows.md:120` — 4개 로케일 **전부**가 "MoAI-ADK ... puts them into its **actual pipeline** — the sync-phase 4-dimension quality evaluation (sync-audit-4dim) and the plan-phase parallel research fan-out (plan-research-fanout) are implemented as workflow scripts"라고 공개 서술한다.
- `.claude/rules/moai/workflow/dynamic-workflows.md:105` — `codemaps-extract.js`를 "canonical pattern for mechanical read-only fan-out"의 worked example로 문서화한다.

즉 **공개 문서가 배선 상태를 주장하는데 실제 호출 경로가 없다**. 이는 단순 고아 자산 문제가 아니라 `verification-claim-integrity.md` §1.1 관점의 미검증 주장이다. 본 관찰이 REQ-APO-016(주장 정합화)을 신설하게 했다. 브리프의 4-axis 스코프에는 없던 항목이다.

### E.2 정정 2 — 템플릿 미러 비대칭이 배선 형태를 강제한다

브리프는 "Verify which touched files have template mirrors and scope accordingly"라고만 지시했다. 실측 결과는 단순 확인 이상의 설계 제약이었다: 배선 대상(`plan.md` 등)은 미러 대상이고 배선 자산(`.js`)은 의도적 비배포이므로, **무조건 참조는 배포 사용자에게 dead reference를 만든다**. capability-gate가 선택이 아니라 필수 조건이 되며, 이것이 REQ-APO-011 / REQ-APO-063을 별도 요구사항으로 승격시킨 근거다.

---

## §F 결정 사항 (RESOLVED)

세 항목 모두 사용자 결정으로 해소되었다. 미해소 clarification 마커는 남아 있지 않다.

- **D1 — 배포 채택.** 3개 스크립트를 템플릿 미러한다. SPEC-DWF-CODEMAPS-PILOT-001의 비배포 AC를 명시 supersede하고 §25 중립화를 수행한다. capability gate는 존치(§H.4).
- **D2 — read-only drafter + 단일 적용자 확정.** disjoint-writer 변형은 문서화된 향후 선택지로만 보존. Phase 12는 Group 1 결과와 독립이다.
- **D3 — 선행 측정 후 결정.** M4 첫 작업으로 게이트 grep을 실행하고 출력을 기록한 뒤 분기한다. 현 시점에 제거 여부를 단정하지 않는다.

---

## §H 배포 결정(D1) 전제 실측 — 코디네이터 지시 대비 정정 3건

D1 지시에 포함된 세 전제를 실행 검증한 결과, 그중 셋이 사실과 달랐다. 잘못된 전제로 작업을 정의하면 (a) 불필요한 가드 완화로 dev-only 격리가 약해지고 (b) 중립성 판정이 공허해지므로 정정 내용을 요구사항에 반영했다.

### H.1 `split_namespace_test.go` 가드는 이미 prefix-scoped — 개정 불요

지시 전제는 "가드가 `.claude/workflows/*.js`를 기계적으로 차단하므로 좁혀야 한다"였다. 실제 코드(L93-104)의 차단 조건은 `splitHarnessAgentPrefixes` 6종 접두사 매칭이다. 5개 스크립트명을 접두사 규칙에 대입한 결과:

```
plan-research-fanout.js          NOT blocked
sync-audit-4dim.js               NOT blocked
codemaps-extract.js              NOT blocked
hns-oss-docs-run.js              NOT blocked   (user-owned harness — 의도적 비대상)
hns-release-update-run.js        BLOCKED (prefix=hns-release)
```

가드는 이미 "dev-only Runner는 막고 generic fan-out은 통과"라는 원하는 분리를 정확히 수행한다. 따라서 **가드 소스 수정 작업은 존재하지 않으며**, 남는 것은 배포 이후에도 차단이 유효함을 확인하는 불변식 단언(AC-APO-072b: 임시로 `hns-release-update-run.js`를 심어 FAIL을 관측한 뒤 제거)이다.

### H.2 중립성 leak 스캐너가 `.js`를 읽지 않는다 — 공허 green 위험

`internal_content_leak_test.go`의 `leakTextExtensions`는 `.md` / `.tmpl` / `.yaml` / `.yml` / `.sh` / `.json` 6종이며 `.js`가 없다. 스크립트를 템플릿에 추가한 뒤 "중립성 가드 green"을 근거로 삼으면 **스캐너가 파일을 열지도 않은 채 통과한 판정**이 된다.

실제 위반은 존재한다:

| 파일 | 라인 | 매치 내용 |
|---|---|---|
| `codemaps-extract.js` (62줄) | — | 0건 — 이미 중립 |
| `plan-research-fanout.js` (132줄) | 35-36 | `REQ-ATR-018/019/020`, `AC-ATR-023/024/025`, `design.md §D`, `acceptance.md` |
| | 54 | `design.md §D + REQ-ATR-020` |
| `sync-audit-4dim.js` (173줄) | 37-38 | `REQ-ATR-015/016/017`, `AC-ATR-020/021/022`, `design.md §C` |
| | 42 | `spec_id: "SPEC-FOO-001"` — 플레이스홀더지만 C1 정규식 `SPEC-[A-Z0-9-]+-[0-9]{3}` 매칭 대상 |

따라서 `.js` 확장자 추가(REQ-APO-072)는 중립성 판정(REQ-APO-071)이 유효해지기 위한 **선행 조건**이며, D1이 실제로 요구하는 유일한 Go 변경이다. 예상되었던 "가드 축소"와 방향이 반대다.

참고: `pedagogicalAllowlist` 기구가 존재하므로 `SPEC-FOO-001` 같은 교육용 플레이스홀더는 allowlist 등재로도 해결 가능하나, 스크립트 예시는 일반 식별자(`<spec-id>` 등)로 바꾸는 편이 단순하다.

### H.3 `moai update` 보존 계약도 prefix-scoped — 변경 불요

`internal/cli/update/plan/plan.go` L135-145 / L189-196은 `.claude/workflows/hns-` 및 `.claude/workflows/harness-` 접두 경로만 user-owned로 판정한다. generic 3개는 이 집합 밖이므로 자동으로 **template-managed**(덮어쓰기 가능)가 되며, 이는 배포 자산에 필요한 정확한 의미다. `update_preserve_inventory.go` 헤더 주석(L23)도 동일하게 `hns-*.js` / `harness-*.js`만 보존 대상으로 명시한다.

`internal/template/catalog.yaml`에는 `.claude/workflows` 항목이 없다(유일한 "workflows" 매치는 스킬 설명 문자열 L232). 배포는 embedded 트리의 generic FS walk이므로 catalog 등록도 불필요하다.

### H.4 배포해도 capability gate는 필요하다

배포는 **파일 존재**를 보장하지만 **런타임 지원**을 보장하지 않는다. dynamic workflow 실행에는 Claude Code 최소 버전 요구가 있으므로 구버전 사용자는 파일을 받고도 실행할 수 없다. gate 조건을 "파일 존재 AND 런타임 지원"으로 확장해 유지한다(REQ-APO-011).

---

## §G 참고 자산

- `.claude/skills/moai/workflows/run/task-decomposition.md` L158-172 — Phase 15 참조 GOOD 패턴(배치 + 리다이렉트 + 스냅샷).
- `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution — 7항목 canonical 배치 + file-redirect 계약.
- `.claude/rules/moai/workflow/verification-batch-pattern.md` — 배치 클래스 분류(read-only 안전 판정 기준).
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` §C.2 — Mode 4 동시 spawn 3-5 상한.
- `.claude/rules/moai/development/agent-authoring.md` § Prompt Craft L210-215 — 방어적 스캐폴딩 금지 근거.
- `internal/template/split_namespace_test.go` L93-98 — `.claude/workflows/` 템플릿 유입 차단 가드 선례.
