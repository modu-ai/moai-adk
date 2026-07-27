---
id: SPEC-AGENT-PARALLEL-OPT-001
title: "Agent instruction diet + plan/run/sync parallelization maximization — Progress"
version: "0.5.0"
status: completed
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: ".claude/agents/moai, .claude/skills/moai/workflows, .claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md, .claude/rules/moai/workflow, .claude/workflows, internal/template/templates, internal/template/internal_content_leak_test.go"
lifecycle: spec-anchored
tags: "agent-diet, parallelization, fan-out, workflow-wiring, template-first"
tier: L
---

## §E.1 Plan-phase Audit-Ready Signal

- 산출물 (Tier L 5-artifact): `spec.md` / `plan.md` / `acceptance.md` / `design.md` / `research.md`. `progress.md`는 모든 Tier 공통 파일이며 5-artifact 계수에 포함되지 **않는다**.
- SPEC ID 사전 자가검사: `decomposition: SPEC ✓ | AGENT ✓ | PARALLEL ✓ | OPT ✓ | 001 ✓ → PASS`.
- 규모: **54 REQ (Group 1~6) / 56 AC**.
- Ground truth: `spec.md` §F 전량 실측. 브리프 정정 2건은 `research.md` §E, 배포 전제 정정 3건은 `spec.md` §F.8 / `research.md` §H.
- 미해소 clarification 마커 **0건**.

### 결정 사항 (v0.2.0 — 사용자 승인)

| 결정 | 값 | 반영 위치 |
|---|---|---|
| D1 `.js` 배포 | 템플릿 미러(배포) 채택 | Group 5 REQ-069..073, Group 6 REQ-074..078, `plan.md` M1 |
| D2 sync P12 형태 | read-only drafter 5 + 단일 적용자 확정 | REQ-024 / 024b, `design.md` §B |
| D3 SPEC-ID 마커 | 선행 게이트 후 결정 | `plan.md` §B.3 + M3 작업 0 |

### v0.3.0 — plan-audit FAIL 0.69 대응

**스코프 결정 (사용자 승인): 구 Group 1 (write-concurrency) 전면 철회.**

| 항목 | 조치 |
|---|---|
| REQ-APO-001..005 / AC-APO-001..005 | 삭제 |
| 시나리오 1·2 (write-concurrency 시나리오) | 삭제 — 남은 시나리오 6개로 재번호 |
| 구 M1 (규칙 개정 마일스톤) | 삭제 — M2~M6 → M1~M5 재번호 |
| `spec.md` §A.2 | "정책 차단 요인" → "완화가 필요하지 않다"로 재작성 |
| `spec.md` §C | `### Out of Scope — write-concurrency rule relaxation` 신설 (이관 5항목 명시) |
| `spec.md` §F.4 | 후속 SPEC 입력 자료로 재범위화 (본 SPEC 무편집) |
| `spec.md` §G | 후속 SPEC `SPEC-WRITE-CONCURRENCY-SCOPE-001`(가칭) 명명 — 본 SPEC에서 저작하지 않음 |
| REQ-024b / M2 문구 | "M1 의존 없음" → "현행 write-concurrency 규칙 무의존"으로 교정 |

**결함 해소 매핑**

| 결함 | 조치 | 위치 |
|---|---|---|
| D2 (critical) — shipped rule 모순 | Group 6 신설 REQ-074..077 + AC-074..077. `dynamic-workflows.md` L80/L131 양 표면 개정 + byte-parity + `.js` 헤더 "user-owned" 정정 | `spec.md` §B.6, `research.md` §H.6 |
| D4 (critical) — 게이트 grep 비판별 | **v0.3.0에서 서사만 반영** — `plan.md` §B.3과 `spec.md` §D.2는 교체됐으나 binding surface 2곳(`plan.md` M3 작업 0, AC-APO-043)은 v0.2.0 명령을 유지. **v0.4.0 N1에서 전파 완료** | `plan.md` §B.3 + M3 작업 0, `spec.md` §D.2, AC-APO-043 |
| D5 (major) — 공허 AC | AC-023을 실문구 `Maximum 3 (fix-evaluate cycles \| review iterations)` == 2 로 교정 + 경로 인자 부여 | AC-APO-023 |
| D6 (major) — design.md 부재 | `design.md` 신규 저작 (§B 데이터 흐름 / §C SSOT / §D gate degradation) | `design.md` |
| D7 (major) — C1 정규식 사실오류 | **v0.3.0에서 서사만 반영** — §F.8.3-a / §H.2-a는 정정됐으나 AC-APO-071 셀은 v0.2.0과 byte-identical로 잔존했고, manual-only 표기 의무는 선언만 되고 이행되지 않음. **v0.4.0 N2에서 AC-071 CI 정렬 + AC-071b 신설로 전파 완료** | `spec.md` §F.8.3-a, `research.md` §H.2-a, AC-APO-071 / 071b |
| D8 (major) — 정보 손실 강제 AC | AC-052를 `squash \| merge \| rebase` == 1 로 교정. L163/L191 운용 경로 보존 명시 | AC-APO-052 |
| D9 (minor) — AC-049 임계 미상 | 3항 복합 판정(모델 서술 0 / report 섹션 1 / 잔존 모델 정확히 1) | AC-APO-049 |
| D10 (minor) — 플레이스홀더 | `git diff --name-only ... \| xargs grep` 형태로 실행 가능화 + CI 클래스 정렬 | AC-APO-061 |
| D11 (minor) — e2e 목적지 미명명 | `.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md` 확정 + REQ-078 + `module:` + §E.2 인벤토리 | `spec.md` §B.6, §E.2 |
| D12 (minor) — SSOT 방향 미정의 | 템플릿 = SSOT 확정(REQ-077), M1 작업 순서 template-first로 교정, `moai update` 덮어쓰기 귀결 명시 | `design.md` §C, `plan.md` M1 |
| D13 (minor) — supersession 부정확 | `AC-DCP-010` + `REQ-DCP-009/010` ID 인용 + 정식 grep 문구 + 파일럿(`status: completed`) 편집 소유자 `manager-spec` 명시 | AC-APO-070, `plan.md` §B.1 |
| D14 | 구 Group 1과 함께 삭제. 관찰(L193/L198 이중 서술)은 후속 SPEC 입력으로 보존 | `spec.md` §F.4 |
| D15 | 구 Group 1과 함께 삭제 | — |

### v0.4.0 — plan-audit iter-2 FAIL 0.77 대응

iter-2에서 iter-1 결함 15건 중 **12건이 실행 검증으로 RESOLVED 확인**되었고, `design.md`는 "the cleanest artifact in the set"로 평가되었다. 잔여 결함은 **단일 형태**를 공유한다 — **결정이 서사 섹션에는 반영되었으나 binding surface(MUST AC 셀 / 마일스톤 실행 지시)에 전파되지 않음**. v0.3.0 보고에서 "전파 완료"로 기재했던 3건은 실제로는 서사 전용이었으며, 위 매핑 표의 D4·D7 행을 그에 맞게 정정했다.

| 결함 | 성격 | 조치 | binding surface |
|---|---|---|---|
| N1 (critical) | D4 게이트가 서사에만 반영 | 판별형 명령 + 기대 출력량(0-5줄) + 비판별형 금지 규범을 실행 지시와 AC에 전파 | `plan.md` M3 작업 0, AC-APO-043 |
| N2 (critical) | D7이 AC에 미전파 + manual-only 표기 의무 미이행 | AC-071을 CI 클래스(C1/C2/S2)에 정렬하고 일반형 `SPEC-[A-Z0-9-]+-[0-9]{3}` 제외; manual-only 클래스(내부 날짜 / `/Users/` / SHA `{9,40}`) 표기를 AC-071b로 신설 | AC-APO-071, AC-APO-071b |
| N3 (major) | 분기 상한이 AC에 미전파 | 합계 상한을 분기 조건부(A ≤ 1907 / B ≤ 1927)로 전환, 분기 B가 정상 적용임을 명시 | AC-APO-055 |
| N4 (major) | 구-넘버링 잔재 4곳 | §E Tier 근거를 "배포 위험"으로 재작성(구 근거는 신 Group 1에 대해 거짓), §C 모순 문장 정정, §G 라벨 재지정, `plan.md:53` 재작성 | `spec.md` §C·§E·§G, `plan.md` M2 |
| N5 (minor) | AC-049 (a) 과잉 / (c) 비이진 | (a)를 선택 규칙 산문으로 한정, (c)를 2-마커 합 == 1 명령으로 이진화 | AC-APO-049 |
| N6 (minor) | `module:`에 디렉터리만 존재 | 전체 파일 경로로 확장(4개 아티팩트 frontmatter 동시) | frontmatter |
| N7 (minor) | 다중 파일 `grep -c` 오용 + glob 범위 초과 | 대상 3개 파일 명시 열거 + `grep -l \| wc -l` == 0 형태 | AC-APO-076 |
| N8 (minor) | `"not template-managed" == 0`이 R5 완화책과 충돌 | L80 전칭 주장으로 한정, `hns-*`/`harness-*` 한정 서술 및 L131 보존 허용 | AC-APO-074 |
| N9 (minor) | Phase 귀속 오기 | :190 = Phase 16, :230 = Phase 17, Phase 13은 상한 없음으로 정정 | AC-APO-023 |

**검증 절차 추가**: v0.4.0부터 제출 전 `git diff a1aa064b2 -- acceptance.md`로 수정 주장한 AC 셀이 실제로 달라졌는지 확인한다. 이 검사는 N1/N2/N3를 iter-1에서 잡아낼 수 있었다. 실행 결과 — 수정 주장 13개 셀 전량 CHANGED 확인.

### v0.5.0 — plan-audit iter-3 **0.847** 마감 편집 (F1-F4)

iter-3 점수 **0.847**(Tier L 임계 0.85 대비 **-0.003**). 궤적 0.69 → 0.77 → 0.847 단조 상승, 정체 플래그 없음. must-pass 7/7 PASS. iter-1 critical 4건 + iter-2 결함 9건 **전량 실행 검증으로 RESOLVED 확인**. 변경 AC baseline 12건 재실행 중 11건 정확. `module:` 수정과 자가 발견 silent-replace 결함은 독립 검증 통과. AC-071b는 1차 리뷰에서 3개 사실 주장 전량 정확 판정.

| 결함 | 조치 | binding surface |
|---|---|---|
| F1 (major) | AC-049(c) `M` 마커를 **정의 heading 앵커**(`^## HRN-003 …`)로 교체하고 기재 baseline을 합 2로 정정. 앵커 없는 형태는 :49 산문 cross-reference까지 세어 M=2가 되고, "계층형 유지 + 평면형 제거" 분기가 합 2로 **부당 FAIL** 했다(오직 "평면형 유지" 분기만 통과 가능). 앵커 적용 후 두 분기 모두 합 1 → PASS | AC-APO-049 |
| F2 (major) | AC-024b가 v0.2.0 상태로 잔존 — **구-넘버링 stale의 4번째 사례**. REQ-024b 문구("현행 `[HARD]` 절대 금지형 규칙을 그대로 둔 채")를 반영하고 규칙 완화 전제 서술 0건 grep을 추가 | AC-APO-024b |
| F3 (minor) | REQ-071이 금지한 5개 클래스 중 3개(내부 날짜 / SHA `{9,40}` / `/Users/`)가 SHOULD 등급 AC-071b의 **기록 의무**로만 커버되어 DoD 면제 가능했던 구조적 누수 → AC-071b에 `== 0` 판정 추가(manual-only / CI-unenforced 라벨은 유지). 감사 실측상 3개 클래스 모두 0 매치라 실제 노출은 없었음 | AC-APO-071b |
| F4 (minor) | 미한정 "Group 1" 2곳(`spec.md` §C, `research.md` §F)에 `구` 한정어 적용. **추가로** 감사가 허용 판정한 7곳도 동일 조치 — 현 넘버링에서 Group 1 = fan-out 배선(철회되지 않음)이므로 "Group 1 철회"는 신규 독자에게 오독 위험이 있다 | `spec.md`·`plan.md`·`acceptance.md`·`design.md`·`progress.md` |

**근본 원인 확정 (F2)**: assert 가드는 실패하지 않았다 — 식별한 13개 셀은 13/13 실제 변경되었다. 결함은 **replacement 실패가 아니라 identification 누락**이다. AC-024b는 애초에 후보 목록에 없었으므로 가드가 도울 수 없었다. 감사의 체계적 스윕(REQ-touched ∖ AC-changed)이 후보 5쌍(024/024b/062/069/073)을 산출했고 수동 대조 결과 **024b만 진짜 stale**로 확인 — 고립된 사례이며 더 큰 문제의 징후가 아니다.

### 표준 사전 제출 검사 (2종, v0.5.0부터 필수)

제출 직전 아래 **두 검사를 모두** 실행하고 출력을 보고한다. 검사 1은 v0.4.0부터, 검사 2는 v0.5.0부터 필수다.

**검사 1 — 셀 변경 확인** (수정 주장한 AC 셀이 실제로 달라졌는가):

```bash
git diff <prev-commit> -- .moai/specs/<SPEC-ID>/acceptance.md | grep -E '^[+-]\| AC-APO-'
```

**검사 2 — 집합차 후보 도출** (REQ 본문은 고쳤는데 AC 셀은 안 고친 쌍이 있는가):

```bash
git diff <prev-commit> -- spec.md       | grep -oE 'REQ-APO-[0-9]+b?'      | sort -u
git diff <prev-commit> -- acceptance.md | grep -oE '^\+\| AC-APO-[0-9]+b?' | sed 's/^+| //' | sort -u
```

두 집합의 차를 구하고 **진짜 본문 재작성만 필터링**한 것이 후보 목록이다. 검사 1은 "고쳤다고 주장한 것"을 검증하고, 검사 2는 "고쳐야 하는데 놓친 것"을 발굴한다 — 두 방향이 상보적이며 어느 하나로는 N1/N2/N3(검사 1이 포착)와 F2(검사 2가 포착)를 동시에 잡을 수 없다.

### 미결 관측 항목 (run-phase에서 기록)

- **D3 게이트 출력** — M3 착수 시 판별형 명령을 실행하고 verbatim 출력을 §E.2에 기록한다. plan-phase 참고 실측은 0건이나, **AC 판정 근거는 run-phase 실행 출력**이며 현 시점 값을 인용하지 않는다.
- **`manager-spec.md` 적용 상한** — D3 게이트 분기 확정 후 230(분기 A) 또는 250(분기 B) 중 하나를 기록한다.

## §E.2 Run-phase Evidence

### M3 작업 0 — D3 게이트 (판별형) verbatim 기록

실행 시각: 2026-07-25 (M3 착수 시점, 브랜치 `feat/SPEC-AGENT-PARALLEL-OPT-001`, HEAD `533859188`).

실행한 명령 (`plan.md` §F M3 작업 0 / `acceptance.md` AC-APO-043 판별형, 축자 동일):

```bash
grep -rn 'decomposition:' --include='*.go' --include='*.sh' --include='*.yaml' \
  internal/ .github/ .claude/hooks/
```

verbatim 출력:

```
(no output — exit status 1)
```

출력 줄 수: **0줄** (`| wc -l` → `0`). 기대 범위 0-5줄 이내이므로 명령은 올바르게 좁혀졌다 — 재작성 불요. 비판별형(`decomposition\|segment match trace` 전역 grep, 12,133-match)은 사용하지 않았다.

**게이트 판정: 기계적 소비자 0건 → 분기 A.**

| 항목 | 값 |
|---|---|
| 마커 소비자 수 | **0** |
| 선택된 분기 | **분기 A** |
| `manager-spec.md` 적용 라인 상한 | **≤ 230** |
| 합계 적용 상한 (AC-APO-055, M4/M5 판정) | **≤ 1,907** |
| REQ-APO-043 조치 | 마커 **강제 제거** + 주변 산문 축약 (실행 Bash 정규식 검사는 존치) |

`plan.md` §B.3 각주 확인: `manager-spec.md:166`이 마커가 "downstream grep verification을 가능하게 한다"고 주장했으나, 게이트가 0을 반환했으므로 그 주장은 **근거 없는 것으로 확정**된다 — 인용된 grep은 마커 정의처를 가리키는 자기참조였다.

### M3 Pre-flight 실측 (편차 2건 포함)

| # | 명령 | 기대 | 실측 | 판정 |
|---|---|---|---|---|
| 1 | `wc -l .claude/agents/moai/*.md` 합계 | 2417 | **2444** | **편차 +27** — 전량 `builder-harness.md`(실측 222 vs `spec.md` §F.1 기재 195). `git log` 최종 수정 `eee1c4fc1`(2026-07-20)로 본 SPEC plan-phase(2026-07-25) **이전**이며 `git diff origin/main...HEAD` 상 무변경 → **plan-phase §F.1 측정 오류**이지 후속 drift가 아니다. M3 대상 3파일은 기재값과 **정확히 일치**(plan-auditor 505 / manager-spec 317 / manager-develop 311) |
| 2 | M3 대상 3쌍 mirror `diff` | 0-diff | plan-auditor **0-diff**, manager-develop **0-diff**, manager-spec **1줄 상이** | manager-spec `:139`은 로컬에 `per SPEC-V3R6-LIFECYCLE-REDESIGN-001`, 템플릿에서는 §25 중립화로 해당 토큰 제거 — **의도된 선행 baseline**(`plan.md` §C 항목 3이 예상한 "기존 차이")이며 본 SPEC이 만든 것이 아니다. M3 편집 중 **보존** 대상 |
| 3 | `go test ./internal/skills/... ./internal/template/...` | green | `ok internal/skills` / `ok internal/template 1.457s`, exit 0 | PASS |
| 5 | SSOT 교차참조 목적지 실재 | 3/3 | `spec-frontmatter-schema.md`(`## Canonical 12 Required Fields` @:12) / `model-policy.md`(203줄) / `moai-workflow-spec/SKILL.md`(`### GEARS Format (current)` @:140) | PASS |

편차 2건 모두 M3 편집 대상 3파일의 baseline·분기 판정을 무효화하지 않으므로 blocker 반환 없이 진행했다. 편차 1은 `spec.md` §F.1 / §D.2 합계 정정이 필요하며 이는 `manager-spec` 소관이다(run-phase 에이전트는 SPEC body 수정 금지).

### M3 AC 판정 매트릭스 (REQ 040-048 + 068)

증거 로그: `.moai/state/verify/apo-m3/` (`/tmp` 미사용 — evidence persistence obligation 준수).

| AC | 등급 | 판정 | 명령 | 실제 출력 |
|---|---|---|---|---|
| AC-APO-040 | MUST | **PASS** | `grep -c '^- FC-[0-9]' plan-auditor.md` / `grep -c 'The 12 required fields are'` | `0` / `1` → 열거 블록 1개(MP-3만) ≤ 1. SSOT 교차참조 2건 |
| AC-APO-041 | MUST | **PASS** | `grep -c 'lifecycle: spec-anchored  *#' manager-spec.md` / `grep -c 'spec-frontmatter-schema.md'` | `0` / `3` |
| AC-APO-042 | MUST | **PASS** | `grep -c "Chain-of-Verification" plan-auditor.md` | `0` |
| AC-APO-043 | MUST | **PASS** (분기 A) | 게이트 0건 → 마커 제거; `grep -c 'decomposition:'` / `'segment match trace'` / Bash 정규식 | `0` / `0` / `1` — 실행 Bash 검사 존치 |
| AC-APO-044 | MUST | **PASS** | Step 5 항목 중 Step 4 축자 재진술 grep | `0` |
| AC-APO-045 | MUST | **PASS** | `grep -c '^- \*\*Ubiquitous\*\*' manager-spec.md` / `grep -c 'moai-workflow-spec'` | `0` / `2` |
| AC-APO-046 | MUST | **PASS** | Step 4 개수 서술 vs 열거 블록 수 | 서술 "The four files enumerated below are the Tier M set" / 열거 `4` — 일치. Tier S·L 변형은 `spec-workflow.md` 교차참조 |
| AC-APO-047 | MUST | **PASS** | `grep -c '^## DDD Cycle\|^## TDD Cycle'` / `grep -c '^## Implementation Cycle'` | `0` / `1` — 공통 5-step 골격 + 모드별 차이 |
| AC-APO-048 | MUST | **PASS** | `grep -n 'atomic structural change' manager-develop.md` | `:133 ... scoped **within a single package**. Independent packages MAY progress concurrently — the one-change-at-a-time constraint bounds the package, not the repository.` |
| AC-APO-068 | MUST | **PASS** | 제거 블록별 SSOT 목적지 실측 (아래 표) | 5개 외부 목적지 전량 내용 보유 확인 |

### AC-APO-068 정보 무성 소실 방지 — 제거 블록별 교차참조 실증

| 제거된 블록 | 교차참조 목적지 | 목적지 내용 보유 실측 |
|---|---|---|
| `plan-auditor` FC-1..FC-12 | `spec-frontmatter-schema.md` § Canonical 12 Required Fields / § Field Reference / § Status Enum / § Rejected Snake_Case Aliases | heading 1/1/1, Field Reference 표 행 37, 거부 alias 4종(`created_at`/`updated_at`/`labels`/`spec_id`) 전량 실재 |
| `plan-auditor` Tool Selection Priority + Mandatory Parallel Batching | `agent-common-protocol.md` § Tool Selection by Task / § Parallel Execution | heading 1/1, Grep·Glob·Read 우선 행 3, `single-turn multi-Bash call` HARD 문장 1 |
| `plan-auditor` Tier PASS threshold 표 | `spec-workflow.md` § SPEC Complexity Tier | `plan-auditor PASS threshold` 컬럼 실재 + S 0.75 / M 0.80 / L 0.85 3행 실측 |
| `plan-auditor` M6 Chain-of-Verification | (제거 명령 — REQ-APO-042) 검사 항목은 동일 파일 Audit Checklist Group 2(SC-6) / 3(RQ-1,2) / 4(AC-4,5) / 6(CN-1)에 이미 열거. 유일한 비중복 지시(표본 아닌 전수 점검)는 Audit Checklist 서두 1줄로 보존 | 동일 파일 내 도달 — 외부 참조 불요 |
| `manager-spec` 12-field 스키마 블록 | 위 `spec-frontmatter-schema.md` 동일 | 동일 |
| `manager-spec` GEARS/EARS 패턴 표 | `moai-workflow-spec/SKILL.md` § GEARS Format / § EARS Format + `references/ears-deep-dive.md`; 추가로 `manager-spec` frontmatter가 **preload** 하는 `moai-foundation-core` 가 5-패턴 전문 보유 | GEARS/EARS heading 1/1, `ears-deep-dive.md` 실재(yes), `moai-foundation-core` 5-패턴 열거 1 |
| `manager-spec` SPEC-ID 마커 강제 + 예시 표 | (제거 명령 — REQ-APO-043 분기 A). 유효/무효 예시는 산문으로 보존(`SPEC-AUTH-001`·`SPEC-V3R6-SPEC-ID-VALIDATION-001`·`SPEC-RETIRED-DDD-001` 및 무효 3종) | 동일 파일 내 보존 — `pedagogicalAllowlist` 등재 2토큰 **의도적 유지** |
| `manager-spec` Scope Boundaries | 동일 파일 frontmatter `description` 의 `NOT for:` 절 | 동일 파일 내 도달 (부수 효과: stale `manager-develop/tdd` 참조 해소) |
| `manager-spec` Delegation 도메인 3-bullet | 동일 파일 Step 6 + `archived-agent-rejection.md` §C rows 7-10 | §C backend/frontend/devops 행 4건 실측 |
| `manager-develop` DDD/TDD 전문 2회 기술 | 동일 파일 § Implementation Cycle 통합 골격 | LARGE_SCALE·LSP baseline·loop prevention(max 100 / stale 5)·completion marker(type errors == 0)·coverage 85%·checkpoint 경로 전량 이관 확인 |
| `manager-develop` Migration 표 + archived 목록 | 동일 파일 산문 1문장 (동일 매핑의 2회 기술 제거) | 동일 파일 내 도달 |
| `manager-develop` Cycle Selection Decision Guide | 동일 파일 § Required Input Parameter 의 `Focus:` 2행 + § Implementation Cycle 선택 문장 | 동일 파일 내 도달 |
| `manager-develop` Scope Boundaries / Delegation Protocol 이중 열거 | 동일 파일 통합 표 (경계와 라우팅을 1회 기술) | 6개 라우팅 전량 보존 |

### M3 라인 수 결과 (`wc -l` 실측)

| 파일 | 착수 전 | M3 후 | `spec.md` §D.2 상한 | 판정 |
|---|---:|---:|---:|---|
| `plan-auditor.md` | 505 | **454** | 340 | **미달성 (+114)** — 아래 blocker B2 참조 |
| `manager-spec.md` | 317 | **236** | 230 (분기 A) | **미달성 (+6)** |
| `manager-develop.md` | 311 | **238** | 240 | **충족 (-2)** |
| 10파일 합계 | 2444 | **2239** | (M4·M5 판정 대상) | M4 대상 5파일 미착수 |

AC-APO-055는 M4 소관(`plan.md` M4 작업 7)이므로 본 M3에서 판정하지 않는다. 다만 M3 대상 3파일은 M4가 손대지 않으므로, 위 2건의 미달성은 M4에서 해소 불가하며 blocker B2로 보고한다.

### 본 세션 run-phase 커밋 이력 (M3 잔여 + M4)

| SHA | 마일스톤 | 파일 | diff |
|---|---|---:|---|
| `87ea1bc18` | M3 작업 0-4 — core-agent 본문 다이어트 | 8 | +336 / −621 |
| `03d42eeb3` | M3 작업 1b — 파일 내 재진술 제거 | 5 | +8 / −72 |
| `cdeceb251` | M4 작업 1-5 — 5개 에이전트 다이어트 + recipe 추출 | 13 | +233 / −599 |
| `25231bd6d` | M4 작업 6 — parallel-batching 교차참조 스윕 | 9 | +12 / −4 |

**커밋 방식 (방법 기록 — blocker B1 해소 경로)**: 병렬 트랙이 동일 에이전트 파일에 미커밋 편집을 보유한 상태였으므로, 본 세션의 run-phase 커밋 전량을 격리된 `git worktree`에서 저작하고 `git hash-object` + `git update-index`로 index에 직접 스테이징했다 — 공유 워킹트리는 건드리지 않았다. 그 결과 동일 파일 6개에 걸친 병렬 트랙의 frontmatter 변경(`effort: xhigh` → `high`)이 **본 SPEC의 모든 커밋에서 배제**되어 REQ-APO-065 스코프 순도가 보존됐다.

### AC-APO-055 라인 수 판정 — `wc -l .claude/agents/moai/*.md` @ `25231bd6d`

| 파일 | 실측 | 상한 | 판정 |
|---|---:|---:|---|
| `plan-auditor.md` | 430 | 430 | PASS |
| `manager-spec.md` | 229 | 230 | PASS |
| `manager-develop.md` | 239 | 240 | PASS |
| `sync-auditor.md` | 146 | 150 | PASS |
| `manager-git.md` | 190 | 190 | PASS |
| `manager-design.md` | 202 | 205 | PASS |
| `builder-harness.md` | 168 | 170 | PASS |
| `e2e-tester.md` | 146 | 150 | PASS |
| `manager-docs.md` | 116 | 120 | PASS |
| `super-advisor.md` | 108 | 112 | PASS |
| **합계** | **1974** | **1997** | **PASS** |

**감축률 — 두 baseline을 분리해 기록한다 (혼동 방지)**:

| baseline | 기준점 | 값 | → 1974 | 감축 |
|---|---|---:|---|---|
| 본 브랜치 직전 | `c48faf6ed` 시점 10파일 합계 | 2212 | 2212 → 1974 | **−238 (−10.8%)** |
| plan-phase §F.1 | `origin/main` 10파일 합계 | 2417 | 2417 → 1974 | **−443 (−18.3%)** |

`spec.md` §D.2가 명시한 "≥16%" 주장은 **plan-phase `origin/main` baseline 2417 기준**이며 위 표의 두 번째 행(−18.3%)이 그 판정 대상이다. 첫 번째 행(−10.8%)은 본 브랜치에서 실제로 발생한 감축량이며 서로 다른 기준점이므로 상호 대체 불가하다. 분기 A가 적용된다(D3 게이트가 기계적 소비자 0건을 반환).

### AC-APO-054 — M4 작업 6 parallel-batching 교차참조 스윕

```bash
grep -l "verification-batch-pattern\|Parallel Execution" .claude/agents/moai/*.md | wc -l
```

실제 출력: **10** (임계 ≥ 8 → PASS). 본 SPEC 착수 전 값은 **1**이었다.

### M5 검증 배치 — 격리 워크트리 `25231bd6d` 실행 결과

| # | 검증 항목 | 결과 |
|---|---|---|
| 1 | `make build` | exit 0. 실행 후 워킹트리 clean — 즉 커밋된 `catalog.yaml`이 생성기 산출물과 **바이트 동일** |
| 2 | 미러 바이트 parity (에이전트 10개 전량) | 선행 SPEC-ID 중립화 차이만 잔존 — `builder-harness` 4줄 / `manager-docs` 1줄 / `manager-spec` 1줄 / `sync-auditor` 1줄, 나머지 6개는 동일. **신규 차이 0** |
| 3 | 템플릿 중립성 | `internal/template/templates/` 하위에서 `REQ-APO\|AC-APO\|SPEC-AGENT-PARALLEL` 매치 파일 **0건**. 범용 fan-out 스크립트 3종 배포 확인: `codemaps-extract.js`, `sync-audit-4dim.js`, `plan-research-fanout.js` |
| 4 | 하네스 비배포 invariant | 템플릿 트리에 `hns-*` / `harness-*` / `my-harness-*` 스킬 디렉터리 **없음**, `.claude/agents/harness/` **없음**. 로컬에는 user-owned `hns-*` 스킬 **7개** 존재 — 의도된 상태 |
| 5 | `go test ./...` | exit 0 — **107 패키지 ok, 0 FAIL**. 명명 가드 개별 실행: `TestTemplateNeutralityAudit` PASS / `TestTemplateNoInternalContentLeak` PASS / `TestSplitHarnessNamespaceNoLeak` PASS / `TestRuleTemplateMirrorDrift` PASS / `TestCatalogHashParity` PASS |
| 6 | archived 에이전트 명 | `.claude/agents/moai/` 전량에서 매치 **0건** |
| 7 | frontmatter invariant (REQ-APO-065) | `git diff origin/main...HEAD -- .claude/agents/moai/ internal/template/templates/.claude/agents/moai/ \| grep -E '^[-+](effort\|model\|tools\|color\|name\|description\|permissionMode\|memory\|hooks\|skills):'` → **출력 없음** (SPEC 전 구간) |
| 8 | supersession (AC-APO-070) | 4개 조건 전량 충족 — `partially_supersedes: [SPEC-DWF-CODEMAPS-PILOT-001]` 실재 / `AC-DCP-010`·`REQ-DCP-009`·`REQ-DCP-010` ID 인용 / 무효화된 grep 문구 명시 / 파일럿 SPEC의 `acceptance.md`+`progress.md`가 `REQ-APO-070`·`AC-APO-070`을 역참조하는 SUPERSESSION 주석 보유 |
| 9 | `dynamic-workflows.md` 로컬 vs 템플릿 | 동일 |
| 10 | `moai-workflow-testing/references/e2e-desktop-native-recipes.md` | 양쪽 실재, **27줄**, 미러 0-diff |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-25
run_commit_sha: pending-backfill-blocked   # 커밋 미수행 — blocker B1 참조
run_status: blocked-at-commit
milestone: M3
ac_pass_count: 10        # AC-APO-040..048 + 068 (M3 범위 전량)
ac_fail_count: 0
preserve_list_post_run_count: 0            # PRESERVE 목록 파일 무변경
l44_pre_commit_fetch: not-run              # 커밋 미수행
l44_post_push_fetch: not-run
new_warnings_or_lints_introduced: 0        # golangci-lint 0 issues
cross_platform_build:
  host: pass                               # go build ./... exit 0
  windows: pass                            # GOOS=windows GOARCH=amd64 go build ./... exit 0
full_test_suite: fail-preexisting-cascade  # TestCatalogHashParity / TestManifestHashFormat — blocker B1
total_run_phase_files: 6                   # 로컬 3 + 템플릿 미러 3
m1_to_mN_commit_strategy: not-applied      # blocker B1 로 커밋 보류
```

### Blocker B1 — catalog.yaml 해시 재생성이 병렬 세션 미커밋 변경집합과 충돌 (커밋 보류)

본 세션 편집과 **무관한** 병렬 변경이 07:54:11에 워킹트리에 유입되었다(본 세션 최초 편집 08:28:28보다 선행, mtime 실측). 내용은 profile 매트릭스 frontmatter 재작성(`effort:` xhigh→high / high→medium)이며 **에이전트 10개 × 로컬·템플릿 2트리 = 20파일 + `internal/template/catalog.yaml`** 에 걸쳐 있다. 본 세션이 건드리지 않은 7개 파일도 동일하게 변경돼 있어 귀속이 명확하다.

충돌 구조:

- catalog 해시는 파일 바이트의 평문 SHA256이며 **frontmatter를 포함**한다(실측: 워킹트리 파일 shasum = 테스트가 보고한 `computed` 값과 일치, staged blob은 상이).
- 본 세션 body 편집으로 3개 entry가 drift → `TestCatalogHashParity` / `TestManifestHashFormat` FAIL.
- 재생성(`gen-catalog-hashes --all`)은 워킹트리를 읽으므로 병렬 세션의 `effort` 변경을 본 SPEC 커밋에 흡수하게 된다 → REQ-APO-065(frontmatter 무변경) 위반 + L46 귀속 위반.
- body-only 스테이징(frontmatter hunk 제외)은 검증 완료했으나(staged frontmatter 변경 0줄, mirror parity 유지), 그 경우 커밋된 트리의 catalog 해시가 불일치해 테스트가 red로 남는다.

따라서 자기정합적 커밋 경로가 없어 **커밋을 수행하지 않았다**. 편집 산출물은 워킹트리에 온전히 존재하며 index는 원상 복구했다(병렬 세션의 `git add`/commit 간섭 방지).

해소 경로(오케스트레이터 결정 필요): (a) 병렬 세션이 profile 변경을 먼저 커밋 → 이후 `gen-catalog-hashes --all` + 본 M3 커밋, (b) 본 SPEC 커밋에 `effort` 변경 흡수를 명시 승인(REQ-APO-065 예외 기록 필요), (c) 병렬 변경 revert 후 본 M3 단독 커밋.

### Blocker B2 — `spec.md` §D.2 상한 2건이 M3 인가 범위로 도달 불가

- `plan-auditor.md` 454 vs 상한 340(-114 부족). `plan.md` M3 작업 1이 인가한 편집(frontmatter 열거 1회화 + M6 CoVe 제거)의 산출 감축량은 약 27줄이며, 추가로 SSOT 중복(Tool Selection / Parallel Batching / Tier threshold 표)을 제거해 총 51줄을 감축했다. 잔여 114줄은 rubric band 정의, D7/D8 검증 verb, Audit Checklist Group 3-6, Output Format 템플릿 등 **plan-auditor 고유 소유 내용**이며, 제거 시 REQ-APO-068(정보 무성 소실 금지) 및 §G 안티패턴 "라인 수만 맞추는 압축" 위반이다.
- `manager-spec.md` 236 vs 분기 A 상한 230(-6 부족). 동일 사유.
- 아울러 `spec.md` §F.1 / §D.2의 `builder-harness.md` 기재값 195는 실측 222와 불일치하며(합계 2417 vs 2444), 이는 plan-phase 측정 오류다. §D.2 합계 상한(1907/1927)의 근거 수치 정정이 필요하다.
- 상한 조정 또는 M3 범위 확대는 `spec.md` / `plan.md` body 수정이므로 **manager-spec 소관**이다.

### 최종 run-phase 신호 (M1-M5 완료 — 위 M3 시점 신호를 갱신)

```yaml
run_complete_at: 2026-07-25
run_commit_sha: c538a6bc6
run_status: audit-ready
milestone: M1-M5 (전량 완료)
ac_pass_count: 13        # 본 progress.md에 실측 기록된 AC 한정 — M3 매트릭스 10건 + M4/M5 실측 3건(054/055/070). SPEC 전량 AC 집계가 아님
ac_fail_count: 0
preserve_list_post_run_count: not-measured-at-M5   # M5 배치 미포함 (미러 parity 검증 #2는 별개 항목)
l44_pre_commit_fetch: not-recorded-in-this-delegation
l44_post_push_fetch: not-recorded-in-this-delegation
new_warnings_or_lints_introduced: 0                # go test ./... exit 0, 107 pkg ok / 0 FAIL
cross_platform_build:
  host: pass                                       # make build exit 0 @ c538a6bc6, 이후 워킹트리 clean
  windows: not-re-run-at-M5                        # M5 검증 배치에 미포함 (M3 시점 신호의 pass 기록 참조)
full_test_suite: pass                              # go test ./... exit 0 — 107 패키지 ok, 0 FAIL
total_run_phase_files: "8 / 5 / 13 / 9"            # 본 세션 4커밋의 커밋별 파일 수. 합산은 중복 포함이므로 미기재
m1_to_mN_commit_strategy: multi-commit (M1..M5)    # 본 세션 기여분 4커밋, 격리 워크트리 + index-level 스테이징
```

### blocker 해소 상태

- **B1 (catalog 해시 재생성 충돌로 커밋 보류)** — **해소**. 격리 워크트리 저작 + `git hash-object` / `git update-index` index-level 스테이징으로 병렬 트랙의 frontmatter 변경을 흡수하지 않고 커밋했다(§E.2 커밋 방식 기록). REQ-APO-065 invariant는 SPEC 전 구간에서 **출력 없음**으로 실측 확인됐다(§E.2 M5 배치 #7).
- **B2 (`spec.md` §D.2 상한 2건 도달 불가)** — **해소**. `spec.md` v0.10.0(`c48faf6ed`)의 §D.2 상한 재보정 이후 실측 10파일 전량이 상한을 충족한다(§E.2 AC-APO-055 표: `plan-auditor` 430/430, `manager-spec` 229/230).

### 잔여 부채 — sync-phase 이관

`builder-harness.md`는 커밋 시점 **168/170**으로 충족하나, 병렬 트랙이 동일 파일에 하네스 model/effort 정책 **+27줄** 추가를 미커밋 상태로 보유하고 있다. 양쪽이 모두 랜딩하면 해당 파일은 약 **195줄**이 되어 상한 170을 초과하며, 그 경우 `spec.md` §D.2.1 절차에 따른 상한 재보정이 필요하다. 본 run-phase에서는 해소하지 않고 **sync-phase 부채 항목으로 기록**한다.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-25
sync_commit_sha: 52c5dba79
sync_status: audit-ready
ac_pass_count: 56
ac_fail_count: 0
ac_debt_count: 0
ac_total: 56
```

### 4-dimension score + verdict

독립 sync-auditor 패스(격리 워크트리 `.claude/worktrees/sync-apo001`, HEAD `8f0426f4b`, read-only, git 변경 없음)가 `.moai/reports/sync-audit/SPEC-AGENT-PARALLEL-OPT-001-sync-audit.md`에 전체 리포트를 남겼다.

| Dimension | Weight | Score | Verdict | Evidence |
|---|---:|---:|---|---|
| Functionality | 40% | 0.92 | PASS | `go test ./...` → `exit=0`, 107 packages `ok`, 0 FAIL. 52/56 AC PASS(교정 전); 54개 REQ 전량 실질 충족 |
| Security | 25% | 0.95 | PASS | 배포된 `.js` 3종에 대한 위험 구문 프로브(`eval\|new Function\|child_process\|exec(\|execSync\|spawn(\|require(\|process.env\|fs.(write\|unlink\|rm)`) → 매치 0. Credential 프로브 → 0. Subagent boundary: `AskUserQuestion` 2건 전량 부정형 주석("No AskUserQuestion") |
| Craft | 20% | 0.84 | PASS | `golangci-lint run --timeout=3m` → `exit=0`, `0 issues.` 라인 상한 전량 충족(1974/1997). AC 판정 명령 2건의 저작 결함(scope-mismatch·formatting-brittleness) |
| Consistency | 15% | 0.93 | PASS | SPEC 전 구간 mirror-parity 회귀 스윕 → REGRESSION 0건. frontmatter-field diff → 0 |

**Harmonic mean = 4 / (1/0.92 + 1/0.95 + 1/0.84 + 1/0.93) = 0.908** — Tier L 임계 0.85 대비 PASS.

### AC-050 / AC-052 교정 기록 (MUST FAIL → PASS)

sync-audit 최초 패스는 must-pass firewall 위반으로 **FAIL** 판정했다(`acceptance.md` §A: "MUST 등급 AC 1건이라도 FAIL이면 SPEC은 close 불가"). 두 결함 모두 **판정 명령 저작 결함**이며 기반 REQ는 이미 충족되어 있었다.

**AC-APO-050 (MUST)** — 구 명령 `grep -ciE -e 'nextra' -e 'wcag' -e 'page.?speed' -e 'lighthouse' manager-docs.md`(전체 파일) → `1`(요구 `== 0`). 유일한 매치는 frontmatter `description:` 블록(:6) — AC-065가 **변경 금지**로 명시한 필드. 본문(body)만 스코프한 재실행:
```bash
B=$(awk 'NR>1 && /^---$/{f=1;next} f' .claude/agents/moai/manager-docs.md); grep -ciE -e 'nextra' -e 'wcag' -e 'page.?speed' -e 'lighthouse' <<< "$B"
```
결과: **0**(origin/main 동일 파이프라인 baseline: 7). 구 명령은 AC-065와 동시 충족 불가능한 형태였다. 교정: `acceptance.md`에서 body-scope 명령으로 대체(manager-spec 소관, 커밋 `6f5a28ff0`). 위 인용은 `acceptance.md` AC-APO-050 셀의 축자 명령이다 — `acceptance.md` 표 셀은 파이프 문자가 markdown 표 행을 깨뜨리므로 `|` 파이프 대신 here-string(`<<< "$B"`)으로 awk 출력을 grep에 전달한다.

**AC-APO-052 (MUST)** — 구 명령 `grep -c 'squash | merge | rebase' manager-git.md`(bare pipe, BRE literal-string 해석) → `0`(요구 `== 1`, baseline 3). origin/main baseline은 정확히 `3`으로 검증됨. 잔존 유일 서술이 백틱으로 재포맷되어(`:32`) 리터럴 매치가 깨진 것 — **REQ-APO-052("규칙을 한 번만 명시")는 실질 충족**. 포맷-관용 재실행:
```bash
grep -cE 'squash.{0,6}merge.{0,6}rebase' .claude/agents/moai/manager-git.md
```
결과: **1**(origin/main 동일 패턴: 3) — AC가 의도한 3→1 축약이 실제로 발생했음을 확인. 운용 경로 2곳(`gh pr merge --squash`, `gh pr merge --<merge_method>`) 보존 확인. 교정: `acceptance.md`에서 포맷-관용 정규식으로 대체(manager-spec 소관, 커밋 `6f5a28ff0`, 3줄 변경). 위 인용은 `acceptance.md` AC-APO-052 셀의 축자 명령이다 — 리터럴 `\|` 교대(alternation) 대신 `.{0,6}` 스팬으로 세 토큰의 순차 근접을 판정하며, 이는 markdown 표 셀 안에서 리터럴 파이프가 행을 깨뜨리는 문제도 함께 회피한다.

**교정 후 AC 매트릭스: 56 PASS / 0 FAIL / 0 PASS-WITH-DEBT / 0 UNVERIFIED (총 56).** MUST 등급 FAIL 0건 — `acceptance.md` §A close firewall 해제. (2026-07-25 amendment로 AC-APO-072/072b PASS-WITH-DEBT 2건이 해소되어 최종 56 PASS로 갱신 — 아래 "PASS-WITH-DEBT 2건 — RESOLVED" 참조.)

### PASS-WITH-DEBT 2건 — RESOLVED (2026-07-25 amendment)

원래 두 AC는 "템플릿 트리 변형이 필요해 read-only 감사 범위 밖"이라는 이유로 RED/GREEN round-trip 미실행 상태로 PASS-WITH-DEBT 판정되었다. 이 이연 사유는 본 amendment에서 무효화되었다 — 격리 워크트리(`fix/SPEC-APO-001-ac072-amendment`, base `c7309aeb6`)에서 실제 RED/GREEN round-trip을 수행하고 직접 관측했다.

- **AC-APO-072 (MUST)** — 원 이연 사유: "비중립 스크립트 주입 → guard FAIL → 중립화 → guard PASS round-trip은 템플릿 트리 변형이 필요해 범위 밖". 사용 프로브: `internal/template/templates/.claude/workflows/redgreen-probe.js` (SPEC-ID 리터럴 `SPEC-V3R6-REDGREEN-PROBE` 포함). 관측된 FAIL: `TestTemplateNoInternalContentLeak` → `internal_content_leak_test.go:735: [1] templates/.claude/workflows/redgreen-probe.js | class=C1-spec-id-prefix | match=SPEC-V3R6-REDGREEN-PROBE`. 프로브 제거 후 GREEN: `ok  github.com/modu-ai/moai-adk/internal/template  0.672s`. → **RESOLVED, PASS**.
- **AC-APO-072b (MUST)** — 원 이연 사유: "`hns-release-update-run.js` 주입 → `SPLIT_HARNESS_NAMESPACE_LEAK` FAIL 기대 확인은 템플릿 트리 변형이 필요해 재실행하지 않음". 사용 프로브: `internal/template/templates/.claude/workflows/hns-release-update-run.js`. 관측된 FAIL: `TestSplitHarnessNamespaceNoLeak` → `split_namespace_test.go:96: SPLIT_HARNESS_NAMESPACE_LEAK: dev-only split-harness Runner ".claude/workflows/hns-release-update-run.js" found in embedded template tree.` 프로브 제거 후 GREEN: `ok  github.com/modu-ai/moai-adk/internal/template  0.227s`. → **RESOLVED, PASS**.

두 프로브 모두 매처를 먼저 읽은 뒤 저작되어 vacuous RED가 아님을 확인했다: C1 클래스 패턴(`internal_content_leak_test.go:163`)은 always-on(strict-gate 아님)이고 `.js`는 `leakTextExtensions`(:554)에 등록되어 있으며, `splitHarnessAgentPrefixes`에는 `hns-release-update`가 포함되어 있다. 프로브 제거 후 워크트리 clean 상태(`git status --short` → 0 dirty)도 확인됨. 전체 round-trip 로그: `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/57efc472-b30c-47d3-8baf-87bb4d01d8d3/scratchpad/verify/ac072-roundtrip.log` (세션 로컬 경로 — 재현 시 동일 절차로 재생성).

**Amendment-close 기록**: `amendment_of: SPEC-AGENT-PARALLEL-OPT-001` (self-referential in-place amendment). manager-spec 재오픈 커밋 `bf683eb7b`(spec.md/plan.md/acceptance.md → `status: in-progress`). 본 sync 커밋(§ 아래 SHA)이 4개 아티팩트를 `status: completed`로 재클로즈한다. `sync_commit_sha`는 원본 클로즈 값(`52c5dba79`)을 그대로 보존 — 이 amendment는 원본 sync 커밋을 재작성하지 않으며, 새 커밋이 자기 자신의 SHA를 참조할 수 없다는 물리적 제약과 `internal/spec/era.go` H-4가 필드의 non-empty 여부만 요구한다는 점(`moai spec audit --json` 재확인 결과 이 SPEC 유일 finding은 `EraAutoDetected` INFO, repo-wide MUST-FIX 0)에 근거한다.

### CI-unenforced 라벨 (AC-APO-071b)

AC-APO-071b가 요구하는 3개 manual-only 누출 클래스(내부 날짜 / `/Users/` 경로 / SHA `{9,40}`)는 **CI-unenforced**로 명시 라벨링한다 — `MOAI_TEMPLATE_LEAK_STRICT=1`이 `.github/` 또는 `Makefile` 어디에도 설정되어 있지 않다(sync-audit 실측, `.moai/reports/sync-audit/SPEC-AGENT-PARALLEL-OPT-001-sync-audit.md` § Residual Risk #3). 감사 실측상 3개 클래스 전량 매치 0이나, 이는 **"CI green"이 이 3클래스의 증거가 아님**을 의미한다 — 향후 회귀는 CI가 포착하지 못한다. 이는 문서화·라벨링된 한계이며 결함이 아니다.

### Gaps

1. ~~AC-072 RED/GREEN round-trip + AC-072b blocking-validity 주입~~ — **RESOLVED (2026-07-25 amendment)**: 격리 워크트리에서 실제 프로브 주입 → FAIL 관측 → 제거 → GREEN 확인까지 완주했다. 상세는 위 "PASS-WITH-DEBT 2건 — RESOLVED" 참조.
2. `make build` embedded-tree lookup(AC-069 2번째 조건) — 미실행(변형 빌드 스텝). 정적 대체 관측: `embed.go:28`의 `//go:embed all:templates`가 `.claude/workflows/`를 포괄하며, `go test ./internal/template/`(catalog/manifest 해시 parity 검사 포함)는 통과.
3. Coverage 측정 — 미실행. 이 SPEC은 구조적으로 무관(`git diff --name-only a1aa064b2^..8f0426f4b | grep '\.go$'` → M3-M5 구간 없음; 유일한 Go 파일 변경은 M1의 `internal_content_leak_test.go`뿐 — 프로덕션 Go 코드 무변경).
4. docs-site 렌더 출력(AC-016) — 소스 텍스트만 검증(en/ko/ja/zh `workflows.md` 각 1건 fan-out-script 언급 확인); Hugo 빌드 미실행.
5. 산문-판정 잔여 항목(AC-020 lens adequacy, AC-028(b)/AC-030(b) per-site 전칭) — 명령은 per-file 개수만 기계화, reviewer-read 설계 그대로 유지.

### Residual risk

1. **브랜치가 `origin/main` 대비 189개 파일 변경 보유** — 27개만 이 SPEC의 M3-M5 구간 소속. `origin/main`을 baseline으로 삼는 향후 머지-타임 검증은 무관 작업을 이 SPEC에 오귀속할 위험(CFP-1이 실증한 바로 그 함정) — SPEC 자체 커밋 구간을 baseline으로 삼아야 한다.
2. **`builder-harness.md`에 상속된 mirror drift**(이 SPEC의 결함 아님, 그러나 이 브랜치에 잠복) — 로컬 `:75`/`:131`이 템플릿의 `Use WebSearch / WebFetch`와 다름(`:131`은 의미상 중복 오탈). `make build`가 *템플릿*을 embed하므로 배포 사용자는 정상 텍스트를 받으며, 로컬 dev 사본만 오류 — 별도 1줄 수정을 PR #1141 트랙에 권고.
3. **CI-unenforced 3클래스**(위 § 참조) — 오늘 0매치이나 향후 회귀 미포착 가능.
4. **F1/F2 결함은 반복 클래스의 신규 2가지 형태다** — `acceptance.md` §A.1이 이미 5개 저작 규칙을 카탈로그하고 있으나, F1(scope mismatch: 파일 전체 vs 본문)과 F2(formatting brittleness: 리터럴 vs 백틱)는 그 5개에 없던 신규 형태. 다음 SPEC이 반복하지 않도록 §A.1에 추가 권고.
5. **`plan-auditor.md`가 상한 정확히(430/430)에 위치**, `manager-git.md`도 190/190 — 향후 1줄 추가가 즉시 AC-055를 깨뜨릴 여유 없음.

### 브랜치 재정렬 기록 (cherry-pick onto origin/main)

**왜(Why)** — 원 브랜치 `feat/SPEC-AGENT-PARALLEL-OPT-001`은 이미 `origin/main`에 스쿼시 머지된 3개의 무관 SPEC(`SPEC-CLI-TUX-INIT-UPDATE-001` PR #1145, `SPEC-CONFIG-AUDIT-REPAIR-001` PR #1142, `SPEC-CLI-WIZARD-RESTRUCTURE-001`) 커밋 21개를 함께 보유한 채 오래된 시점을 base로 삼고 있었다. 이 상태로 PR을 열었다면 **189개 파일 변경**(그중 **105개가 본 SPEC 스코프 밖**)이 표시되고, trial merge는 **20건 충돌**(그중 **15건이 이미 머지된 파일**에서 발생)을 냈을 것이다.

**무엇을 했나(What)** — 본 SPEC의 21개 커밋을 `cherry-pick -x`(각 신규 커밋 본문에 `cherry picked from commit <old>` 출처 라인 보존)로 `origin/main`(`758624d6e`) 위에 재적용해 브랜치 `feat/SPEC-AGENT-PARALLEL-OPT-001-clean`을 생성했다. 결과: **56개 파일, 스코프 밖 0건**.

**유일한 실질 충돌과 해소** — `.claude/agents/moai/builder-harness.md`: `origin/main` 쪽은 3줄에 걸친 `WebSearch / WebFetch`(스페이스 포함) 타이포그래피 정정을, 본 SPEC은 동일 블록을 2줄로 압축했다. 양쪽을 모두 반영해 해소했다 — main의 스페이싱 정정을 본 SPEC의 2줄 압축형에 적용. `internal/template/catalog.yaml`은 수동 병합 대신 `make build`로 재생성했다.

**손실 없음(Nothing lost)** — 원 브랜치는 변경 없이 그대로 존재하며, 추가로 태그 `backup/SPEC-AGENT-PARALLEL-OPT-001-pre-rebase`(`6e3677ac4`)로 고정되어 있다. 아래 표의 모든 구 SHA는 여전히 reachable하다.

**old → new SHA 매핑표** — 아래 표는 이 파일(§E.2 커밋 이력 표, AC-APO-055 표제, M5 검증 배치 표제, §E.3 blocker 산문)에서 이미 인용된 구 SHA를 재정렬 후 SHA로 해석하기 위한 것이다. **위 이력 산문은 재작성하지 않는다** — 원 브랜치에서 실제로 일어난 일의 정확한 기록이며, 매핑표는 독자가 그것을 재정렬 후 좌표로 해석하는 도구다.

| old | new |
|---|---|
| `a1aa064b2` | `8c224c887` |
| `e3a3eb4e4` | `a5423e62b` |
| `3925074bb` | `d8b7a8b51` |
| `b4eb07721` | `e5575c219` |
| `4aff0cb3d` | `7aef61f60` |
| `af0ce0195` | `c471790cf` |
| `d8815a722` | `d5c7627a9` |
| `731c40875` | `a6f47b40f` |
| `874c403ef` | `ec729ed83` |
| `c401c3ad2` | `a91861f02` |
| `29edb0aa9` | `f2360b6df` |
| `533859188` | `874fbc9ff` |
| `c48faf6ed` | `4250f9855` |
| `87ea1bc18` | `4fc2e1a70` |
| `03d42eeb3` | `acbed5b98` |
| `cdeceb251` | `c9739aa62` |
| `25231bd6d` | `c538a6bc6` |
| `8f0426f4b` | `787bb02c7` |
| `6f5a28ff0` | `e56d5b85f` |
| `aa24273aa` | `52c5dba79` |
| `6e3677ac4` | `05ebec300` |

**재정렬 후 재검증(실측)** — 재정렬된 브랜치(`feat/SPEC-AGENT-PARALLEL-OPT-001-clean`, HEAD `05ebec300`) 위에서 아래 항목을 재실행하고 관측값을 기록한다:

| 항목 | 명령 | 관측값 |
|---|---|---|
| 전체 테스트 | `go test ./...` | exit 0, 105 패키지 ok, 0 FAIL |
| AC-APO-050 교정 명령 | (위 § "AC-050 / AC-052 교정 기록" 참조) | `0` (임계 `== 0`) |
| AC-APO-052 교정 명령 | (위 § "AC-050 / AC-052 교정 기록" 참조) | `1` (임계 `== 1`) |
| 10파일 라인 수 합계 | `wc -l .claude/agents/moai/*.md` | `1970` (상한 1997) |
| frontmatter invariant (REQ-APO-065) | `git diff origin/main...HEAD -- .claude/agents/moai/ internal/template/templates/.claude/agents/moai/ \| grep -E '^[-+](effort\|model\|tools\|color\|name\|description\|permissionMode\|memory\|hooks\|skills):'` | 출력 없음 |
| 템플릿 빌드 | `make build` | exit 0 |
