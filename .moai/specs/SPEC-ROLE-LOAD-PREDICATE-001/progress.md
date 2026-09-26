---
id: SPEC-ROLE-LOAD-PREDICATE-001
title: "진행 기록 — Codex 역할 로드 판별식의 계약 중립화 (승계 노선, 축소판)"
version: "0.5.0"
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: High
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tags: "codex, role-load, acceptance-criteria, predicate-design"
---

> frontmatter에 `status:`를 두지 않는다 — SPEC의 생애 상태는 `spec.md` 한 곳에 산다(iter-1 D21).

## §E.1 Plan-phase Audit-Ready Signal

- Tier: M (spec.md + plan.md + acceptance.md + progress.md)
- 카드: t1171 · 워크트리 `.claude/worktrees/t1171` · 브랜치 `WT-role-load-predicate` · 기준 `develop 0356e8117`
- SPEC 버전 **0.5.0 — iter-5 결함 델타 소진판**(0.4.0의 축소를 유지하고 그 위에서 열거된 다섯 건만 고쳤다)
- plan-audit 이력: iter-1 `FAIL 0.65` → iter-2 `FAIL 0.68` → iter-3 `FAIL 0.75` → iter-4 델타 `FAIL 0.767`(critical 0, must-pass 7/7 PASS). 점수는 네 번 올랐다. 3회차 FAIL로 리드의 확장 종료 조건이 걸려 iter-4가 감사자의 범위 축소안을 적용했고, 4회차 FAIL에서 **운영자가 「다섯 건만 수정하고 Kickoff」(선택지 A)을 택했다** — 이 판이 그 다섯 건이며 재감사는 그 델타로 한정된다
- 노선: 승계 유지. 형제 SPEC 두 개의 파일을 고치지 않는다(세 회차 모두 Route A 준수가 측정으로 확인됐다). `AC-DHR-012`·`AC-DHR-023`은 영구 미충족
- 예산: **판정 대상 REQ 11 / AC 7** (Tier M 상한 16 / 16 이내)

### 축소 — 무엇을 잘랐고 대가를 어디에 적었는가

| 절단 | 대상 | 대가를 적은 자리 |
|---|---|---|
| ① | `AC-RLP-005`(LIVE 14회) → 후속 카드 후보 | `spec.md` §F.1 — 잃는 것은 「새 실행에서도 참인가」의 **재확인** 하나. 말하지 못하게 되는 셋(CLI 버전 차이, 새 실행의 본문 상이성, 감사 역할 쓰기 거부의 동반 관측)을 명시 |
| ② | `AC-RLP-010`·`AC-RLP-011` → **삭제** | `spec.md` §F.2 — 세 회차 변이 이동 표(D9 → E2 → F1)와, 「실제 방어」로 지명했던 항이 `REQ-RLP-012` 때문에 이 카드 안에서 결코 변하지 않았다는 구조적 한계 |
| ③ | `AC-RLP-007`(선고정 조상 판정) → 후속 카드 후보 | `spec.md` §F.3 — 여섯 케이스와 상속 증거 경로. F3(동결 복구 경로 없음)을 미해결로 함께 넘김 |

`REQ-RLP-009`·`REQ-RLP-013`은 **산문 의무로 유지**되고 검증이 plan-audit·sync-audit의 **읽기**에 위임됐다(`spec.md` §B의 두 처분 블록, `acceptance.md` §A.1). `REQ-RLP-006`·`REQ-RLP-008`은 판정식이 이관돼 **이 카드에 AC가 없다** — 넷 모두 「기계가 지킨다」는 주장을 뺐고, 그것이 공허 초록을 없애는 방법이다.

### 살아남은 AC 일곱

`AC-RLP-001`(파서) · `002`(기대표 적격성) · `003`(유도·선별·판별력) · `004`(계약 중립성) · `006`(A/B 분리 구조 집행) · `008`(형제 무수정) · `009`(AC 스냅숏). 번호 **005·007·010·011은 비어 있고 재사용하지 않는다.**

### iter-3 결함 처리 요지 (전문은 `plan.md` §H-3)

| 결함 | 처분 |
|---|---|
| F1 `AC-010` 변이 통과 | 범위에서 제거(판정식 삭제) |
| F2 `AC-011` 서술 반대 통과 | 범위에서 제거(판정식 삭제) |
| F3 동결 복구 경로 없음 | 범위 밖 — §F.3이 **미해결로** 넘김 |
| F4 상한 증인 열 과대 | 범위 밖 — §F.1이 미해결로 넘김 |
| F5 대장 과대 기재 셋 | **정정함**(아래) |
| F6 스테일 잔여 행 모순 | **정정함**(아래) |
| F7 스테일 `COUNT 15` + 유령 AC | **정정함** — 수를 단언하지 않고, 예시 경로를 바꿔 유령 제거(아래) |
| F8 `-S` 의미 오기 | 소멸 — 그 문장이 이관 대상 안에 있었다 |
| F9 `AC-009` 한 호출 거부 | **부분**(iter-5 정정) — iter-3의 2단계도 거부됐고, iter-5가 **네 단계**로 재분할했다. 전문은 `plan.md` §H-3 |
| F10 `REQ-008` 이분 대 삼분 | 범위 밖 — §F.1 |
| F11 M3b 제목·배치 불일치 | **정정함** — 실행 순서 `M1 → M2 → M3b → M3 → M4`를 절 안에 못 박음 |
| F12 스냅숏에 형제 두 행 | **기록함** — M3 [HARD] + 커밋 메시지 의무. Route A 위반 아님도 함께 |

### F5 정정 — 과대 기재 세 자리

- **(a) E19.** iter-3의 HISTORY와 §H-2 E19 행이 「§B.1 머리에 배치 사유 한 줄」을 완료로 적었으나 그 줄은 없었다(`grep -n '015' spec.md` → 두 줄뿐). **iter-4에서 실제로 넣었다** — `spec.md` §B.1 머리의 문장이 `REQ-RLP-015`가 주제순 배치임을 적는다. 이제 주장이 참이다.
- **(b) E2·E3.** 「닫음」 → **「부분 — 저자 변이는 막고 감사자 변이는 통과」**로 정정하고, iter-4 처분(판정식 삭제)을 같은 행에 적었다.
- **(c) 집계.** iter-3의 `RESOLVED 23 / PARTIAL 1 / UNRESOLVED 0` → 감사자 판정 **`RESOLVED 21 / PARTIAL 2 / UNRESOLVED 1`** 채택. D9 UNRESOLVED(세 회차 모두 뚫렸다), D12·D13 PARTIAL(D13은 구현이 없어 **실행되지 않았고** 설계 타당성만 확인됐다).

### F6 정정 — 세 자리를 고쳤으나 그것으로 닫히지 않았다 (iter-5 G6 정정)

> iter-4는 이 절을 「세 자리가 같은 이야기를 한다」로 적었다. **같은 §G 표의 다른 두 행이 LIVE 범위인 채 남아 있었으므로 그 문장은 과대 기재였다.** 아래 세 항은 참이고, 나머지 두 행과 B.7은 iter-5가 처분했다(위 G5·G6 행).

- `plan.md` §G 위험 표의 「위상 순서가 조상 관계를 함의하지 않는다 / 닫히지 않는 잔여」 행을 **폐기**하고, 「선고정 게이트가 없다 → 위험이 아니라 범위 밖」과 「동결 복구 경로 없음 → 후속 카드가 물려받는다」 두 행으로 교체했다.
- `plan.md` §H D8 행이 서술하던 폐기된 `max(감시)<min(LIVE)` 형태를 정정하고, 그 판정 자체가 §F.3으로 이관됐음을 적었다.
- §H 말미의 최종 기록을 「iter-2 잔여 사유는 거짓이었고(E7), iter-3이 조상 판정으로 없앴고, iter-4가 그 판정을 이관했다」는 한 흐름으로 정리했다.

### F7 — 유령 AC는 절단만으로 사라지지 않는다 (확인 결과를 명시한다)

유령 `AC-ANCHOR-SCOPE-001`은 iter-3이 양성 대조 예시로 넣은 경로 `.moai/specs/SPEC-AC-ANCHOR-SCOPE-001/acceptance.md`의 부분 문자열이 카운터 패턴에 걸려 생겼다. **절단은 이것을 제거하지 않는다** — 그 경로는 `AC-RLP-009`의 양성 대조에 있고 `AC-RLP-009`는 살아남기 때문이다.

그래서 절단과 **별도로** 예시 경로를 바꿨다: `.moai/specs/SPEC-ADVISOR-RUNG-001/acceptance.md`. 스냅숏에 존재하고(`COUNT 7 live=7`), 이름에 `AC-` 부분 문자열이 없다. 같은 판정식으로 GREEN 방향을 재확인했다:

```
GREEN(new path) pass=1 bad=0 absent=0 row=1
true
```

> 이 출력은 **iter-4 판(2단계) 형태**다. iter-5가 판정식을 네 단계로 재분할하고 항 둘(`fresh`·`ordered`)을 더했으므로 현재 형태의 출력은 아래 측정 표에 있다. 이 블록은 F7 처리 시점의 기록으로 남긴다.

**이관된 AC 번호 둘은 `[RETIRED]` 인접 표식으로 제외했다.** `acceptance.md` §A.1의 두 행이 `AC-RLP-007`·`AC-RLP-005`를 「이 REQ의 판정식이었다」는 설명으로 언급하므로, 표식 없이 두면 카운터가 둘을 **살아 있는 AC로 센다.** 그래서 코드 스팬 안에 `AC-RLP-007 [RETIRED]` 형태로 적었다(닫는 백틱이 인접을 끊으므로 표식은 스팬 **안**에 들어가야 한다). 각각 한 번만 나타나고 그 한 번이 표시됐으므로 「일부만 표시」로 인한 카운터 중단(ambiguous halt)은 발생하지 않는다 — 아래 실측이 그것을 확인한다.

**절단 후 실측 (이 실행):**

```
absent-from-snapshot .moai/specs/SPEC-ROLE-LOAD-PREDICATE-001/acceptance.md: COUNT 10
```

분해: 자기 AC **7**(001·002·003·004·006·008·009) + 인용한 형제 AC **3**(`AC-DHR-012`·`AC-DHR-023`·`AC-CAR-012b`) = 10. iter-3의 16(자기 11 + 인용 4 + 유령 1)에서 내려왔고, **유령 `AC-ANCHOR-SCOPE-001`은 사라졌다**(경로 교체), 이관된 둘은 표식으로 제외됐고, `AC-CAR-009`는 옛 `AC-RLP-008` 근거문이 재작성되며 인용이 사라졌다. 테스트는 `pass` 2 / `fail`·`skip` 0으로 정상 종료 — ambiguous halt 없음.

`acceptance.md`의 [HARD] 절은 그럼에도 **어떤 수도 단언하지 않는다.** 위 10은 이 실행의 관측으로 여기 기록하고, AC 본문은 「카운터 자신의 출력과 스냅숏 행의 존재로 판정하고 수를 손으로 재현하지 않는다」만 말한다 — M3에서 인용이 늘거나 줄면 또 달라지기 때문이다.

### iter-5 결함 델타 처리 (G1~G10) — 강화 회차가 아니라 열거된 다섯 건의 소진

iter-4 델타 감사 판정: **FAIL 0.767**(0.65 → 0.68 → 0.75 → 0.752 → 0.767), critical 0, must-pass 7/7. 차단 다섯 건은 전부 문서 한두 줄이고 코드가 없다. 전제(12/12)는 세 회차가 닫았고 **다시 재지 않았다**. F5·F6·F7·F11은 수리 확인분이므로 열지 않았다(F6의 **문면**만 G6에 따라 「부분」으로 정정 — 수리 자체를 되짚은 것이 아니다).

| 결함 | 처분 | 무엇이 그것을 닫았는가 |
|---|---|---|
| **G3** 2단계가 지정 환경에서 거부된다 | **닫음** | `acceptance.md` §B의 판정을 **네 단계**(1a git만 / 1b go만 / 2a git만 / 2b git 없는 판정)로 재분할. 방아쇠를 문면에 적었다 — 복잡도가 아니라 **한 호출 안의 `git` + 명령 치환 공존**. `plan.md:112` pre-flight #6을 네 단계 점검으로, §H-3 F9 행을 **「부분」**으로, `acceptance.md` §A의 [HARD] 항에서 「2단계로 나눠 해결」 함의와 iter 라벨 오기를 함께 제거 |
| **G4** 위조 1단계 산출물에 `true` | **기계화(선택지 A) — 부분적으로만 닫힌다** | 1a가 측정 시점 HEAD, 2a가 판정 시점 HEAD를 적고 2b가 `fresh`·`ordered`를 **판정 항**으로 본다. 아래 변이 대조 참조. **닫은 것은 스테일 재사용이고, 고의 위조는 닫히지 않는다** — 그 잔여를 Residual-risk 5에 적었다 |
| **G1** 삭제된 AC를 판정자로 단언 | **닫음** | `spec.md` §C 머리의 뒷절을 **삭제 이전 사실**로 고쳤다(「iter-3판까지 판정하려 한 대상이었으나 iter-4에서 삭제됐다」) + `§B` 처분 블록과 같은 [HARD] 문장을 같은 자리에 둠 |
| **G2** lint 둘을 넷의 확인으로 읽는다 | **닫음** | 아래 측정 표의 lint 행에 **정직한 수(넷)**, 2 대 4 괴리, 그리고 iter-4가 변이 대조로 측정한 원인(이 SPEC 자신의 `[RETIRED]` 언급이 `006`·`008`의 커버리지 판독을 통과시킨다)을 적었다. `lint.skip`은 쓰지 않았고 경고 둘은 그대로 남긴다 |
| **G5**·**G6** §G 두 행 + 무표식 B.7 | **닫음** | B.7 제목과 처분 주석을 B.8과 같은 모양으로(「iter-4에서 범위 밖으로 이관」), §G의 LIVE 범위 두 행(기대표 버전 귀속 / LIVE 재실패 예산)을 「범위 밖」으로, 첫 행의 「LIVE 전 유도 규칙 재확인」 뒷절을 후속 카드로, `plan.md` §H-3 F6 행과 §H 말미 최종 기록을 **「부분 → iter-5에서 같아짐」**으로 정정 |

**기록된 부채로 실어 보낸다 (G7·G8·G10 — 고치지 않았다).**

| 부채 | 왜 실어 보내는가 |
|---|---|
| **G7** `spec.md:168` 매달린 포인터(`AC-RLP-007`을 현재 시제로 지목) | 바로 위 `:166` 처분 블록이 이관을 선언해 **인접이 스스로 무장해제한다**. G1과 달리 판정자 주장이 아니라 참조다 |
| **G8** 양성 대조 예시 경로가 `superseded` SPEC | **오늘 건전하다**(감사 실측: 스냅숏 10행 `COUNT 7 live=7`, `AC-` 부분 문자열 0, 같은 논리로 `true`). 판정식 본문이 아니라 산문 예시다. **선택 기준을 남긴다** — 대체 경로는 ① 스냅숏에 행이 있고 ② 경로 문자열에 `AC-` 부분 문자열이 없고 ③ `status: completed`인 SPEC이면 되고, 그 SPEC이 아카이브되면 대조가 조용히 RED로 바뀌는 것이 이 부채의 실현 형태다(VCI §2.1의 이동 좌표) |
| **G10** `plan.md` D9 행 본문 대 집계 | 집계(`RESOLVED 21 / PARTIAL 2 / UNRESOLVED 1`)가 그 행을 **명시적으로 뒤집으므로** 표식 없는 잔여가 아니다. 행만 읽으면 오독한다는 위험만 남는다 |

**G9는 부채 목록에서 빠졌다 — 지나가며 닫았다.** 한 토큰(`acceptance.md` §A의 「iter-2 F9」→「iter-3 F9」)이고 그 줄을 G3 때문에 이미 다시 쓰고 있었다. 요구된 수리가 아니었음을 여기 적는다.

### 저작 시점 측정 (Claim / Evidence / Baseline / Gaps)

| 주장 | 명령 | 관측 |
|---|---|---|
| SPEC ID 적격 | `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` | `PASS` |
| 네 아티팩트 lint (0.4.0) | `moai spec lint SPEC-ROLE-LOAD-PREDICATE-001` | `0 error(s), 2 warning(s)` — `REQ-RLP-009`·`REQ-RLP-013`의 `CoverageIncomplete`. **둘은 관측이고, 선언은 넷이다(iter-5 G2 정정).** `acceptance.md` §A.1 둘째 표가 무-AC로 선언하는 REQ는 `006`·`008`·`009`·`013` **넷**이며, lint는 그중 **둘만** 낸다. 이 괴리를 iter-4 감사가 변이 대조로 측정했다 — 처분표의 `AC-RLP-007 [RETIRED]`·`AC-RLP-005 [RETIRED]` 언급을 다른 토큰으로 바꾸자 `0 error(s), 4 warning(s)`로 `006`·`008`이 함께 나타났다. 기제는 `internal/spec/lint_coverage_sibling_table.go`의 행별 AC id 적중이며(`006`→1, `008`→1, `009`→0, `013`→0), 즉 **이 SPEC 자신의 `[RETIRED]` 언급이 두 REQ의 커버리지 판독을 통과시킨다.** 따라서 이 둘은 참인 경고이지만 **선언과 일치한다는 근거가 아니다** — 넷 중 둘은 기계가 침묵한 자리다. `lint.skip`으로 덮지 않는다(관측 없는 주장을 주제로 삼은 SPEC에서 참인 경고를 숨기는 것이 금지선이고, 막을 것은 경고가 아니라 그것을 「선언대로」로 읽은 문장이었다) |
| 절단 후 AC 코퍼스 수 | `go test -json ./internal/spec -run '^TestACCounterFullCorpusMatchesBaseline$'` | `COUNT 10`(자기 7 + 인용 3). 유령 AC 소멸, ambiguous halt 없음 — `pass` 2 / `fail`·`skip` 0 |
| `AC-RLP-009` 네 단계 발행 가능성 (G3) | 1a·1b·2a·2b를 **원문 그대로** 각각 한 번 발행 | 넷 다 **거부 없이 실행**, 각각 exit 0. 2a는 `git` 셋(리터럴 인자, 치환 없음)이 한 호출에 있어도 통과한다 — iter-4가 격리한 방아쇠가 치환 공존임을 이 실행이 다시 보인다 |
| `AC-RLP-009` GREEN 방향(예시 경로) | 2a·2b, `P=.moai/specs/SPEC-ADVISOR-RUNG-001/acceptance.md` | `pass=1 bad=0 absent=0 row=1 same_commit=1 fresh=1 ordered=1` → `true`, exit 0 |
| `AC-RLP-009` RED-now(이 SPEC 경로) | 같은 2a·2b | `pass=1 bad=0 absent=1 row=0 same_commit=1 fresh=1 ordered=1` → `false`, exit 1 — **사유는 스냅숏 미재생성이고 신선도 항이 아니다**(`fresh`·`ordered`가 둘 다 1이다). green path는 M3 |
| `AC-RLP-009` 변이 ① 스테일 산출물 (G4가 닫는 쪽) | GREEN 방향에서 `head` 표지만 다른 SHA로 바꿈 | `fresh=0 ordered=0` → **`false`, exit 1**. 다른 커밋에서 만든 산출물의 재사용을 판정 항이 실제로 기각한다 |
| `AC-RLP-009` 변이 ② 고의 위조 (G4가 닫지 **않는** 쪽) | GREEN 방향에서 jsonl을 손으로 쓴 한 줄로 교체(`go test` 미실행), 표지 쌍은 정상 | `pass=1 … fresh=1 ordered=1` → **`true`, exit 0 — 여전히 통과한다.** 이 판이 닫지 못하는 것을 실측으로 고정한다(Residual-risk 5) |
| `AC-RLP-008` 사전 상태 | 판정식 원문 | `changed_total=0 changed_siblings=0` → `false`, exit 1 — 양성 대조가 빈 범위를 부재로 읽지 않는다 |
| 새 예시 경로가 스냅숏에 있고 패턴 안전 | 스냅숏 행 조회 + `AC-` 부분 문자열 검사 | 행 존재(`COUNT 7`), `AC-` 적중 0 |

**전제는 재측정하지 않았다 — 리드가 닫았다.** 12개 역할 / 12개 distinct 해시, 26 기록(표찰 12 / 미표찰 14), 자기 역할 지문 **12/12 mismatch 0**(`manager-lead` 27879 `714d3ca3…`, `mission-governor` 1724 `e3c7bded…`), 파서 1바이트 델타, 다른 트리 기대표 0/12 — 세 회차 감사가 모두 독립 재유도해 참으로 확인했고 iter-4는 다시 재지 않는다.

### Baseline·Gaps·Residual-risk

- Baseline 귀속: 워크트리 `.claude/worktrees/t1171`, HEAD `0356e8117`, 이 실행.
- 판정 도구 귀속: `moai` 설치본 `v3.2.0-rc.16 moai_cp/20260925_122548-14-ga8a9b9376`(build 2026-09-25). 트리 HEAD `0356e8117`와 다른 커밋이며 조상 관계를 재지 않았다 — `moai spec lint` 결과의 도구 귀속 gap이다(VCI §2.2). 0.4.0에서도 유지된다.
- Gaps:
  1. `AC-RLP-001`~`004`·`006`은 구현이 없어 아직 실행 불가다(M2). 이 판에서 실행한 것은 `AC-RLP-009`의 두 방향과 `AC-RLP-008`의 사전 상태뿐이다.
  2. 픽스처 원본(26개)은 primary 체크아웃의 미추적 경로이며 아직 워크트리로 고정되지 않았다. 합성 N1~N4도 없다 — `AC-RLP-003`의 `14 OF 28`은 M1 이후에 처음 측정된다(iter-3이 D13을 PARTIAL로 둔 이유도 이것이다).
  3. `AC-RLP-009` 4항(같은 커밋 재생성)은 카드 범위에 `P` 커밋이 0이어서 **양쪽 방향 모두 공허 통과**다. M3에서 처음 실제로 구속된다.
  4. `REQ-RLP-009`·`REQ-RLP-013`의 충족은 이 카드에 기계 증인이 없다. 감사 읽기가 유일한 확인 경로이며, 그 위임이 감사 절차 쪽에 성문화돼 있지는 않다(§F.2가 후보로 적는다).
- Residual-risk:
  1. 12개 역할 본문이 서로 다르다는 것은 이 트리의 성질이며 보장이 아니다(`AC-RLP-002`가 거른다).
  2. **호출자 층 후처리 결합**(D15 잔여) — 타입 경계는 함수 안을 막지만 호출자가 로드 결과를 행동값과 결합해 후처리하는 것을 막지 못한다. `AC-RLP-006`이 살아남았으므로 **축소 후에도 이 카드의 잔여다**(`spec.md` §C.5).
  3. `AC-RLP-008`은 리터럴 핀을 택했으므로 develop 흡수로 다른 카드가 형제 디렉터리를 건드리면 이 카드가 하지 않은 일로 거짓이 될 수 있다. 그때의 귀속 절차는 AC 본문에 있다.
  4. M3의 스냅숏 전체 재생성은 형제 두 행을 함께 싣는다(F12). Route A 위반은 아니나 커밋 메시지에 적지 않으면 예고 없는 변경이 된다.
  5. **`AC-RLP-009`의 두 성질은 절차이고 기계가 아니다(iter-5 G4).** ① 네 단계의 접속 — 종료 코드 넷을 AND하는 장치는 없고, 앞 단계가 비영이면 뒤 단계를 돌리지 않는 것은 사람이 지킨다. ② **고의로 위조한 한 벌** — `head`/`judge-head`/jsonl 세 파일을 손으로 쓰고 순서를 맞추면 일곱 항이 모두 성립한다(위 변이 ②에서 실측: `true`, exit 0). 신선도 항이 닫는 것은 **다른 커밋의 산출물 재사용**뿐이다(변이 ①에서 `false`). 이 잔여를 기계로 닫으려면 산출물을 서명하거나 테스트 자신이 HEAD를 출력에 실어야 하며, 그것은 이 카드 범위 밖이다 — 아래 후속 카드 후보 표의 §F.1 행이 상속 항으로 받는다.
- 감사 잔여물 주의: `.moai/reports/t1171/`의 `audit-iter2-spec.jsonl`, `probe-ac009.jsonl`, `ac-rlp-00*-*.txt` 일부는 감사자와 이전 회차가 남긴 것이다. run 단계 산출물로 읽지 않는다. **iter-5는 자기 실행 산출물 다섯(`ac-rlp-009{,-head,-judge-head,-p,-b}`)을 보고 전에 삭제했다** — 남겨 두면 그 자체가 G4의 스테일 함정 재고가 되기 때문이며, 감사자가 iter-4에서 같은 이유로 같은 처분을 했다.

### 후속 카드 후보 — 기록만 하고 발행하지 않는다

[HARD] 카드 발행은 큐 변경이며 **운영자 승인 뒤 리드가 수행한다.** 이 SPEC은 후보를 적을 뿐 `moai gtd add`를 부르지 않았다.

| 후보 | 내용 | 상속 증거 |
|---|---|---|
| §F.1 | LIVE 14회 측정(구 `AC-RLP-005`) + `REQ-RLP-006`·`REQ-RLP-008` | 미해결로 물려받는 항: iter-3 F4·F10, D12, **그리고 iter-5 G4의 잔여**(1단계 산출물 위조를 기계로 닫는 길 — 산출물 서명 또는 테스트 자신이 HEAD를 출력에 싣기. 이 카드는 스테일 재사용까지만 닫았다) |
| §F.2 | `REQ-RLP-009`·`REQ-RLP-013` 검증 위임의 성문화 | 세 회차 변이 이동 표(`spec.md` §F.2) |
| §F.3 | 선고정 조상 판정 알고리즘(구 `AC-RLP-007`) | **`.moai/reports/t1171/plan-audit-iter3.md`** §4 — 검증된 여섯 케이스(선형 조상 `true` / 평행 가지 `false` / 한 LIVE 표지의 조상이나 다른 표지의 조상은 아님 `false` / 같은 커밋 `false` / 빈 입력 `false` / **실제 병합 이력(병합 둘 통과) `true`**). 다시 쓰지 말고 물려받는다. F3(동결 복구 경로 없음)을 미해결로 함께 |

- plan-audit 델타 감사 대기

## §E.2 Run-phase Evidence

카드 t1171 run-phase. 워크트리 `.claude/worktrees/t1171`, 브랜치 `WT-role-load-predicate`, 기준 `develop 0356e8117`. cycle_type=tdd. 실행 순서 `M1 → M2 → M3b → M3 → M4`(plan.md §H-3 F11이 못 박은 순서).

### M1 — 픽스처 고정과 선별

`.moai/reports/t1143/m4-live-evidence/ac012-sessions/2026/09/24/`의 실물 기록 26개 전부를 `internal/cli/testdata/codex-rollouts-t1171/real/`로, 그 시점 역할 TOML 12개(이 트리 `internal/template/templates/.codex/agents/moai/*.toml`)를 `.../roles/`로 복사했다. 26개 중 표찰(`agent_role`) 보유는 **정확히 12개**이고, 나머지 14개는 표찰 없는 부모 세션이다(N5 실물 음성 대조). 12개 표찰 보유 기록 각각의 developer response_item 본문 sha256이 자기 역할 TOML의 `developer_instructions`(TOML 파싱값) sha256과 **정확히 일치**함을 Python(`hashlib`+`tomllib`)으로 독립 재확인했다 — `manager-lead`(ordinal 8, 27879+α bytes급)와 `mission-governor`(ordinal 8) 포함 12/12, 매칭 위치는 12개 전부 동일하게 ordinal 8. 14개 부모 세션은 어느 역할 본문과도 일치하지 않음(0/14)을 같은 스크립트로 확인했다.

- `roles-other-version/`(N1): `roles/`의 12개 TOML을 `developer_instructions` 여는 구분자 직후에 마커 라인 하나(`OTHER_VERSION_MARKER_T1171`)를 삽입해 만든, 이름은 같고 본문만 다른 "다른 버전 트리" 기대표. 실물 12개 기록에 대해 0/12 매치를 확인했다.
- `synthetic/n3-label-without-body.jsonl`(N3): `builder-harness` 세션(표찰 보유)에서 자기 본문과 일치하는 `type=="response_item", payload.role=="developer", ordinal==8` 항목 1건을 제거. 표찰은 `builder-harness`로 남고, 매치는 공집합.
- `synthetic/n4-crossed-roles.jsonl`(N4): 같은 `builder-harness` 세션에서 그 항목을 `manager-git` 세션의 해당 항목으로 치환. 표찰은 `builder-harness`인데 매치는 `{manager-git}`.

선별 대상 집합은 26(실물) + 2(N3·N4) = **28개**이고, 표찰 보유는 12(실물) + 2(N3·N4) = **14개**다(§A.0 SSOT와 일치). N1·N2는 별도 파일이 아니라 **기대표 자체**(각각 다른 디렉터리)이며, N2는 M2 테스트가 `t.TempDir()`으로 빈 디렉터리를 즉석 생성한다.

**픽스처 사본 sha256 전수 (52개: 실물 26 + roles 12 + roles-other-version 12 + synthetic 2):**

| 사본 | sha256 |
|---|---|
| `real/rollout-2026-09-24T18-39-36-01a0d2c8-e466-74a1-afe9-dc96ba281e67.jsonl` | `db11ab925f4c61ae12e903352be22fb27c33adfd405ca8bec7efcd1b7f7f748c` |
| `real/rollout-2026-09-24T18-39-44-01a0d2c9-0503-7700-becc-1bc8c98a8198.jsonl` | `d9b737b3ab31ea949282f1ffefa923bedf740a6049140422ba19af6b42c79d65` |
| `real/rollout-2026-09-24T18-39-53-01a0d2c9-27d1-7402-9c3f-50839825dd81.jsonl` | `421ffb08e03b5c3f29746ce518f4bed6cfa6ca1bb420223a6153f5b51a74fbd4` |
| `real/rollout-2026-09-24T18-40-01-01a0d2c9-4759-7b62-98e2-e0ee3cd75452.jsonl` | `78776cf0557ff4bc876a40f0707d9aa168985609204248da8e11cbc4218e87c2` |
| `real/rollout-2026-09-24T18-40-11-01a0d2c9-6dae-7fc1-ac6c-f3a0b455df0c.jsonl` | `4cc4491ada4a77244401de8450ffb5053e4e0f2163d45d70d6c9380ef4d8c926` |
| `real/rollout-2026-09-24T18-40-19-01a0d2c9-8d51-7623-a979-b364671a0205.jsonl` | `ab5d37cf63e255aa2ef86cf8c901dd98de6cf6e7ff50de733c660391129df6a4` |
| `real/rollout-2026-09-24T18-40-28-01a0d2c9-b00c-7f31-8926-420f59ee9ebd.jsonl` | `17f40735e83aba275ae908b1fd627675882e877991af43ee04b8f087ed7be9a3` |
| `real/rollout-2026-09-24T18-40-36-01a0d2c9-cf1a-7bb0-adeb-60c2392aea26.jsonl` | `b86592544285a94be9c941abaf365a71e73be515b69aabbe827c111d76195d9e` |
| `real/rollout-2026-09-24T18-40-43-01a0d2c9-ec75-74a3-af66-5d022f8da3d5.jsonl` | `11dad51246c6a709d3c60ed746fde9b13eecfa7339d10386b1c45b39805a2888` |
| `real/rollout-2026-09-24T18-40-54-01a0d2ca-15aa-7bd2-a616-d7aea311e307.jsonl` | `733995b9c988c331241e410782a77dc41a11ce176a8e8a09845872083501106f` |
| `real/rollout-2026-09-24T18-41-08-01a0d2ca-4c55-7fe2-800b-cd390747a59a.jsonl` | `36149ea66185190af53dc22da4ebee5fa126915ce00c311728d13c14b212b8dc` |
| `real/rollout-2026-09-24T18-41-16-01a0d2ca-6afa-7553-ad97-70d3de5e0931.jsonl` | `5586c34ed35292c06cd1b041a40c5fb70709353bc604d431f55c2a93182edfda` |
| `real/rollout-2026-09-24T18-41-27-01a0d2ca-988c-7b22-897e-2a90133891bc.jsonl` | `e81490deb100fd0d893f84ae4f2b5038db77d2867a5b5db8ae4d3da89b153aab` |
| `real/rollout-2026-09-24T18-41-35-01a0d2ca-b7e2-7e50-8f7f-0ab053beb07f.jsonl` | `eff20609126ffc9a513d0cb949516a3c8cd65a024d9cc9041348f63a8977e5bb` |
| `real/rollout-2026-09-24T18-41-46-01a0d2ca-e10b-7c22-8f98-f111fbbcb2e5.jsonl` | `c2452cce0381ba04969d81479c53a5dd749eb598782702450c32e70e5f820aeb` |
| `real/rollout-2026-09-24T18-41-55-01a0d2cb-037e-7e53-b0b3-f8145b13d586.jsonl` | `16b5585d05f84a256b956bd9e4c3fbe16175c467257ec1796a2c8589d1ea3fab` |
| `real/rollout-2026-09-24T18-42-05-01a0d2cb-2a52-7b11-ad80-44b5165f509d.jsonl` | `9e25eded96b5e8d78517978374b91f0e5726cf3eb2ce31544f2f6ba21274b510` |
| `real/rollout-2026-09-24T18-42-12-01a0d2cb-482d-7951-b192-17e99a3102ed.jsonl` | `cb7438a8a1671f85c3892857409e36bde8ec52d65d4a3d25e846a56a560e29d4` |
| `real/rollout-2026-09-24T18-42-26-01a0d2cb-7f3f-7d73-99ba-cf44079a8d71.jsonl` | `7a923217a8fddea34112ddff9fd287bdba9fbfb6f2177d4e5cf5f0866162beef` |
| `real/rollout-2026-09-24T18-42-34-01a0d2cb-9cc0-7570-b6dc-15bcb55f7ec8.jsonl` | `ed905d72d9109e3e468ac469c4a23e0b8391d2b79b100b3d879afb812bc86ae5` |
| `real/rollout-2026-09-24T18-42-46-01a0d2cb-cadc-7282-841c-39f23441ee7e.jsonl` | `dff2c8137e7bc753be09872afb3a9d14c854e9d52b9c1146a8d30a9aa329a299` |
| `real/rollout-2026-09-24T18-42-54-01a0d2cb-ead4-7471-97fd-9dd09d814dab.jsonl` | `a9ae3c37d7ac04b4993d64496b65ea0726035bdca8e30497919624b55654bd3d` |
| `real/rollout-2026-09-24T18-43-02-01a0d2cc-0a94-74e3-9ee7-a9b4909c9297.jsonl` | `e3fc9d2ae4c999f7b37c1d9fb49b7e7225cb5ec2d757b40406798d6c79c45859` |
| `real/rollout-2026-09-24T18-43-10-01a0d2cc-2968-7d41-a26e-df219b0ba0c6.jsonl` | `68bb89f31cfcbad56fce1ff6922b0f09efdcada961139da4ce4c2ea20c14918d` |
| `real/rollout-2026-09-24T18-43-24-01a0d2cc-5e3b-7f00-b3af-340e8af13aa4.jsonl` | `34e96222a098c89a8c501ccef015e428b5470af6b66d5a0e1db53f8e9a65e56c` |
| `real/rollout-2026-09-24T18-43-37-01a0d2cc-9306-7092-b939-e81d10728dee.jsonl` | `9c22275ddf2073d2c243f47886ec961f3a8fd3b2fe370f7d4156406860fe5b5f` |
| `roles/builder-harness.toml` | `64980f85862eff7c3ad0e7a1f71ff492ea237065d9a453bca63c860650e789a8` |
| `roles/e2e-tester.toml` | `bf7a2214687c22c82d4d1ec7286ae6daf42f8def6d3b9f876d6c54c3f5e839c7` |
| `roles/manager-design.toml` | `2b44f4693f9918e3c9ff7f8a1a7b975f94976c1163fb1bc3eae7d2820d4b4d83` |
| `roles/manager-develop.toml` | `61e9f97723d3e105ac42e46e4723bec98b7f613c0387f548a56ada1d657b5009` |
| `roles/manager-docs.toml` | `e403880986d58f2703c646fe3ad03fa8d72e8244911b2ef0ad5f51864eded9cf` |
| `roles/manager-git.toml` | `c45ccd183909752cf81d5d80715fe6547169c5844bbceb0196e2b7971521286b` |
| `roles/manager-lead.toml` | `33591e406747406da90455fa21c5e890cbf616b4c9a2ad8a9147da6c6a043cff` |
| `roles/manager-spec.toml` | `67b2fe2a4770199a8c58f910788dbe2322926573fefe37c290e45a89380aefff` |
| `roles/mission-governor.toml` | `6c89526a8eb1c8dca0bd1cf85792ac9208fb82011f52ac6516ead3adeb526a8e` |
| `roles/plan-auditor.toml` | `2a20caf3c7e9e2fca1f5f6a89af244178ca85f5f82083acc815aae0689635953` |
| `roles/super-advisor.toml` | `73d08754792332df76d72fe3a17584caf09a5a439f019fb8022f7a698ece54a2` |
| `roles/sync-auditor.toml` | `8aee58a3aa0b853a9bfe4dc0081eebc13af5c860558998addb57df7e7e7dcf8e` |
| `roles-other-version/builder-harness.toml` | `5f04e32280b4b5eb6afbe8586cad9ce55896f0166dd5dcf8126e36916c95a8a1` |
| `roles-other-version/e2e-tester.toml` | `a8957a946497e56f54063e2251a8354543f0c8bcc293d7d080d9519982dc953b` |
| `roles-other-version/manager-design.toml` | `7400b1274ef365dff3db680321912890eeaef9d082dc48e346ebd73d279582f3` |
| `roles-other-version/manager-develop.toml` | `214bbd2b13c1a1301ae8da699af0951e0d8ee1a5ed7751b22ff523ef6f7a86d0` |
| `roles-other-version/manager-docs.toml` | `dae453e347d963fa878a9a1b976cf121bdd79992b285604830c9e4f48bebc248` |
| `roles-other-version/manager-git.toml` | `d997f2d8a875370e1bc211e96634f35cea48239b5f9ffb6c19ab48094d2528c2` |
| `roles-other-version/manager-lead.toml` | `a854866f79a8b1edbca5e477727afc78b6a54257f86aff25ca2cbaa3a6aae41e` |
| `roles-other-version/manager-spec.toml` | `294886680982d67a6a5a30e4e3287476f5b6a3f82f6a50d0abd2fa81e874a4ef` |
| `roles-other-version/mission-governor.toml` | `f443e14acbd2da3893b44be597ca80e4bb198841494d6d7e6cb61f833551f373` |
| `roles-other-version/plan-auditor.toml` | `baf064ff99be240edc06cc28604f9c3ac59126c13bd2fc1118da8573f3adeb51` |
| `roles-other-version/super-advisor.toml` | `157e178e83d71106702e6e6e29c62555b6c79dd21f904a5f94e911a09779f98d` |
| `roles-other-version/sync-auditor.toml` | `0f19e3afe31646b47790501530af50651ab2c2ab9e63d6166489d7a4f70ad925` |
| `synthetic/n3-label-without-body.jsonl` | `b3175823d2e5df48851ab1698c61e17e9576f1a8a8d76b84977d14a2b5c687a8` |
| `synthetic/n4-crossed-roles.jsonl` | `c79eda71345679d861450036a574e8a0fc0a0498a211b575ef7f664ef2257295` |

### M2 — 추출·유도 함수와 타입 경계

구현 파일: `internal/cli/codex_role_fingerprint.go`(`codexRoleBodyExtractTOML`·`codexBuildRoleExpectationTable`·`codexRoleSessionLabel`·`codexRoleFingerprintDerive`·`codexRoleLoadInput`·`codexRoleLoadPredicate`·`codexRoleBehaviour`). 테스트 파일 넷: `codex_role_fingerprint_test.go`(AC-001·002), `codex_role_derive_test.go`(AC-003), `codex_role_contract_test.go`(AC-004), `codex_role_behaviour_test.go`(AC-006).

**RED (구현 삭제 상태에서 이 다섯 테스트를 먼저 발행, 구현 파일 부재로 컴파일 실패):**

```
$ go test ./internal/cli -run '^TestCodexRoleBodyFingerprintParse$|^TestCodexRoleBodyExpectationTable$|^TestCodexRoleBodyFingerprintDerive$|^TestCodexRoleLoadPredicateContractNeutral$|^TestCodexRoleLoadPredicateStructurallyIndependent$' -count=1
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/codex_role_behaviour_test.go:22:57: undefined: codexRoleBehaviour
internal/cli/codex_role_behaviour_test.go:42:24: undefined: codexBuildRoleExpectationTable
internal/cli/codex_role_behaviour_test.go:57:26: undefined: codexRoleSessionLabel
internal/cli/codex_role_behaviour_test.go:73:19: undefined: codexRoleFingerprintDerive
internal/cli/codex_role_behaviour_test.go:77:17: undefined: codexRoleLoadPredicate
internal/cli/codex_role_behaviour_test.go:77:40: undefined: codexRoleLoadInput
internal/cli/codex_role_behaviour_test.go:80:34: undefined: codexRoleBehaviour
internal/cli/codex_role_behaviour_test.go:87:31: undefined: codexRoleLoadInput
internal/cli/codex_role_behaviour_test.go:109:21: undefined: codexRoleFingerprintDerive
internal/cli/codex_role_behaviour_test.go:113:12: undefined: codexRoleLoadPredicate
internal/cli/codex_role_behaviour_test.go:113:12: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
exit status 1
```

**GREEN (구현 복원 후, 다섯 테스트를 acceptance.md §B의 원문 명령으로 각각 발행):**

| AC | 명령 | 관측(원문 발췌) | 종료 |
|---|---|---|---|
| AC-RLP-001 | `go test -json ./internal/cli -run '^TestCodexRoleBodyFingerprintParse$' …` + jq | `PARSE_LEADING_DELTA 1 0x0a` / 5개 서브테스트 pass 1 각각 / `true` | 0 |
| AC-RLP-002 | `go test -json ./internal/cli -run '^TestCodexRoleBodyExpectationTable$' …` + jq | `TABLE_DISTINCT_REASONS 3` / 4개 서브테스트 pass 각각 / `true` | 0 |
| AC-RLP-003 | `go test -json ./internal/cli -run '^TestCodexRoleBodyFingerprintDerive$' …` + jq | `SELECTED_BY_LABEL 14 OF 28` / `FINGERPRINT_POSITIVE_MATCHES 12` / `PARENT_SESSIONS_FALSE 14` / 8개 서브테스트 pass 각각 / `true` | 0 |
| AC-RLP-004 | `go test -json ./internal/cli -run '^TestCodexRoleLoadPredicateContractNeutral$' …` + jq | `CONTRACT_NEUTRAL_LOAD_TRUE 12 NONCE_TRUE 10` / 4개 서브테스트 pass 각각 / `true` | 0 |
| AC-RLP-006 | `go test -json ./internal/cli -run '^TestCodexRoleLoadPredicateStructurallyIndependent$' …` + jq | `BEHAVIOUR_FIELDS_MUTATED 2 OF 2` / `COUPLED_MUTANT_DIVERGED 24` / 4개 서브테스트 pass 각각 / `true` | 0 |

다섯 명령 전부 `acceptance.md` §B의 jq 판정식을 **원문 그대로** 발행했고, 다섯 다 `true`를 stdout에 냈다(jq `-se` 판정이므로 그 자체가 exit 0). 판정 대상 다섯 함수 전부 컴파일·발행 가능함이 이 실행에서 처음 확인됐다(iter-5 감사 §6이 "구현 부재로 미실행"으로 남겼던 목록이 이 M2로 해소됐다).

`AC-RLP-006`의 `codexRoleBehaviour` 필드는 2개(`NonceReturned`, `ContractRefusalObserved`) — 반사 수 2와 변이 수 2가 일치. 결합 변이(`codexRoleLoadPredicateCoupled`, 테스트 전용)는 24회(2필드 × 12역할) 전부에서 실제 발산해 `COUPLED_MUTANT_DIVERGED 24`를 냈다.

### M3b — 기계적 정리

```
$ go vet ./internal/cli/... ./internal/spec/...
(무출력, 둘 다 exit 0)
$ golangci-lint run --timeout=2m ./internal/cli/...
0 issues. (exit 0)
```

테스트 이름은 acceptance.md §B 명령의 정규식과 정확히 일치하도록 저작했다 — 별도 이름 수정 불필요.

### M3 — 판별식 확정 커밋 (같은 커밋에 AC 스냅숏 재생성)

```
$ MOAI_AC_BASELINE_REGENERATE=1 go test ./internal/spec -run TestACCounterBaselineRegenerate -count=1
ok  	github.com/modu-ai/moai-adk/internal/spec	12.655s
```

`.moai/reports/t338/ac-count-baseline.txt`에 `SPEC-ROLE-LOAD-PREDICATE-001/acceptance.md  COUNT 10  live=10 excluded=2 ambiguous=0` 행이 추가됐다. **전체 재생성이 형제 두 행도 함께 실었다(F12 예고대로)**: `SPEC-ACSNAPSHOT-COMMIT-GUARD-001`(COUNT 16), `SPEC-CODEX-AUDIT-READONLY-001`(COUNT 18)이 새로 나타났다 — 이 두 SPEC의 `acceptance.md` 자체는 이 카드가 건드리지 않았고(Route A 판정은 디렉터리 diff만 본다, `AC-RLP-008`), 스냅숏이라는 별개 파일에 예고 없이 실리는 사실만 커밋 메시지에 남긴다.

이 SPEC의 acceptance.md와 이 스냅숏 재생성이 **같은 커밋**에 들어간다 — REQ-RLP-014, `AC-RLP-009` 4항의 요구.

### M4 — 판정·기록

**AC PASS/FAIL 이진 매트릭스 (§B 원문 명령):**

| AC | Status | 명령 | 관측 |
|---|---|---|---|
| AC-RLP-001 | **PASS** | 위 표 | `true`, exit 0 |
| AC-RLP-002 | **PASS** | 위 표 | `true`, exit 0 |
| AC-RLP-003 | **PASS** | 위 표 | `true`, exit 0 |
| AC-RLP-004 | **PASS** | 위 표 | `true`, exit 0 |
| AC-RLP-006 | **PASS** | 위 표 | `true`, exit 0 |
| AC-RLP-008 | **FAIL(사전 상태, 예상됨)** | `git diff --name-only 0356e8117..HEAD` 전/후 조합(§B 원문) | 커밋 전: `changed_total=0 changed_siblings=0` → `false`, exit 1(§E 자기 검증이 명시한 사전 상태 그대로). **M3 커밋 뒤 재판정 필요 — 아래 잔여 참조** |
| AC-RLP-009 | **PASS(예상, M3 커밋 뒤 재판정 필요)** | 1a·1b·2a·2b(§B 원문, 네 단계 각각 별도 Bash 호출) | M3 커밋 이전에는 스냅숏 파일이 아직 커밋되지 않아 `absent=1 row=0`(`false`); **M3 커밋 뒤에는 `absent=0 row=1`(양쪽 다 이 커밋에 실림)으로 전환되어야 한다 — 아래 잔여 참조** |

**형제 세 AC(`AC-DHR-012`·`AC-DHR-023`·`AC-CAR-012b`)는 이 트리에서 실행 불가다.** 판정 명령이 리터럴로 읽는 `.moai/reports/t1100/`이 이 워크트리에 없다(`ls .moai/reports/t1100 2>&1` → `No such file or directory`, 확인함). 실행 결과를 이 SPEC의 완료 조건으로 삼지 않는다(§C.6, `AC-DHR-012`·`AC-DHR-023`은 영구 미충족으로 남는다).

**`REQ-RLP-009`·`REQ-RLP-013`은 이 카드에 판정식이 없다.** 충족 여부는 plan-audit·sync-audit의 읽기가 판정한다(§A.1). 이 자기 검증은 그 사실을 숨기지 않는다.

### 예산 정책 준수 확인 (리드 지시)

이 카드가 손댄 파일은 `internal/cli`(구현 5, 테스트 4) · `internal/spec` 재측정(스냅숏 파일만, 코드 변경 없음) · `.moai/specs/SPEC-ROLE-LOAD-PREDICATE-001/` · `internal/cli/testdata/`뿐이다. `.claude/rules/**`, `CLAUDE.md`, `AGENTS.md`(로컬·템플릿 미러 포함) 중 어떤 파일도 이 카드에서 수정하지 않았다 — **watched-path 파일 미접촉, 예산 테스트(`TestAlwaysLoadedTokenBudget`·`TestCodexContractByteCeiling`) 적용 대상 아님.**

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-09-26
run_commit_sha: pending-backfill-m3
ac_pass_count: 5
ac_fail_count: 0
ac_deferred_count: 2  # AC-RLP-008(사전 상태 기록, 커밋 후 재판정 필요), AC-RLP-009(RED-now, 커밋 후 재판정 필요)
sibling_ac_status: "이 트리에서 실행 불가 (.moai/reports/t1100/ 부재) — AC-DHR-012, AC-DHR-023, AC-CAR-012b"
preserve_list_post_run_count: 0  # 형제 SPEC 두 디렉터리 무수정, git diff로 확인
new_warnings_or_lints_introduced: 0
cross_platform_build:
  linux_darwin: "go build ./... exit 0 (native darwin/arm64)"
  windows: "미실행 — 재측정 범위가 internal/cli·internal/spec로 한정되어 CI의 windows 매트릭스에 위임"
```

**커밋 후 재판정 대기 항목 (M3 커밋 직후 리드/후속 세션이 재확인):**

1. `AC-RLP-008`을 M3 커밋 뒤 재발행 — `changed_total>=1 changed_siblings==0`을 기대한다.
2. `AC-RLP-009`의 2b를 M3 커밋 뒤 재발행 — `absent=0 row=1`을 기대한다(같은 커밋이 acceptance.md와 스냅숏을 함께 실었으므로).

두 항목은 이 진행 기록을 쓰는 시점(커밋 전)에는 **관측할 수 없다** — 자기 자신을 담을 커밋이 아직 없기 때문이다. 이것은 verification-completeness §2의 "green path"이지 gap이 아니다: RED-now의 사유(스냅숏·커밋 부재)가 정확히 서술돼 있고, green으로 넘어가는 조건(이 커밋 자체)도 명시돼 있다.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
