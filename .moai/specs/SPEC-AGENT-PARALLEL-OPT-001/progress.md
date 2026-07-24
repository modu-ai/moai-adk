---
id: SPEC-AGENT-PARALLEL-OPT-001
title: "Agent instruction diet + plan/run/sync parallelization maximization — Progress"
version: "0.5.0"
status: draft
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

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
