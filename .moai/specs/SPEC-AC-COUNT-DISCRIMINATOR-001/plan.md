---
id: SPEC-AC-COUNT-DISCRIMINATOR-001
title: "AC 개수 자가검사 판별자 — 구현 계획"
version: "0.5.0"
status: in-progress
created: 2026-08-28
updated: 2026-08-28
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: ".claude/agents/moai, .claude/rules/moai/development, internal/template/templates/.claude, internal/template/templates/.codex, internal/spec"
lifecycle: spec-anchored
tags: "b12, changelog, acceptance-criteria, self-test, template-first"
---

# 구현 계획 — SPEC-AC-COUNT-DISCRIMINATOR-001

M0 의 미결 결정은 **종결됐다**(선택 A — 인용 축 규약만 확장, 소급 없음). 마일스톤은 **되돌리기 어려운 결정부터** 놓았다. M1 과 M2 는 규약과 계약을 정하는 자리이고 — 한번 전수 corpus 와 배포 템플릿에 나가면 되돌리는 비용이 크다 — M4·M5 는 기계적이다. 검토는 M1·M2 에 몰려야 한다.

## §A 맥락

- 실측: `.moai/reports/t338/measurement.md` (이 트리, 이 카드의 변경 0인 상태에서 측정)
- 스캔 원본: `.moai/reports/t338/overcount-scan.txt` · 검출기: `.moai/reports/t338/overcount-detector.sh`
- 인용 축 범위 결정의 실측(운영자): `.moai/reports/t338/pre-terminal-scan.sh` · `.moai/reports/t338/pre-terminal-scan.txt` — 156 / 122 completed / 34 pre-terminal
- iter-1 감사 수리의 실측(이 트리): `.moai/reports/t338/collision-scan.sh` → `collision-scan.txt` (한 행 다중 접두사 — **수를 얼리지 않는다**; iter-1 시점 123, 같은 HEAD 재측정 125. 방향은 어느 쪽 경계도 아니다 — `spec.md` §3.1, 감사 N7) · `.moai/reports/t338/alias-shape-scan.sh` → `alias-shape-scan.txt` (별칭 형태 상한 85파일, D2)
- 판별자 결정과 그 근거: `spec.md` §3
- base: `da03d9188` (= origin/develop), branch `WT-ac-count-sweep`

## §B 알려진 문제 / 함정

1. **검출기의 오탐 4파일** — `SPEC-V3R3-RETIRED-AGENT-001` · `SPEC-COMPLETION-MARKER-RETIRE-001` · `SPEC-LSPMCP-RETIRE-001` 은 통째로 오탐이고, `SPEC-V3R2-ORC-001` 은 유일한 플래그 `AC-ORC-001-05` 가 오탐이라 **폐기 축에서 빠진다**(`spec.md` §8 — 넷째는 iter-1 감사가 세웠다). 폐기 축 정규화 대상에 넣으면 살아 있는 기준을 폐기로 표시하게 된다. **`SPEC-V3R2-ORC-001` 은 별칭 축에서 따로 대상이 된다**(`spec.md` §2.3) — 같은 파일이 한 축에서는 오탐, 다른 축에서는 진짜다. 축을 섞어 판단하지 않는다.
   - 일반화: 검출기 목록은 **판정이 아니라 후보**다. 놓치기도 하고(폐기 축은 하한 — `SPEC-UPDATE-DOC-DRIFT-001` 은 3건 중 1건만 잡혔다) 잘못 잡기도 한다(오탐 4파일). M5 는 플래그마다 손으로 판정한다.
2. **쌍둥이 드리프트** — 계수기 문면을 테스트에 복사해 두면 절이 바뀌어도 테스트가 통과한다. 반드시 절에서 **추출**한다(REQ-ACD-005).
3. **미러 중립화 되돌림** — `manager-develop-prompt-template.md` 미러는 171행에 SPEC-ID 중립화를 담고 있다. verbatim `cp` 는 이를 되돌린다.
4. **`moai update` 삭제 뿌리** — `.claude/agents/moai` 와 `.claude/rules/moai` 는 관리 대상 뿌리다. 로컬 편집만 하고 미러를 빼먹으면 다음 update 에서 통째로 되돌아간다(CLAUDE.local.md §2.3).
5. **fixture 중립성** — fixture 가 실재 SPEC ID 를 담으면 템플릿 중립성 가드에 걸린다. fixture 는 배포 템플릿 밖(`internal/spec/testdata/`)에 두고 합성 ID(`AC-SYN-00N`)를 쓴다.
6. **반송자를 둘로 세기 — 셋이다(감사 N1)** — B12 절은 `.codex/agents/moai/manager-docs.toml` 에도 통째로 실려 배포된다. 손으로 고치는 대상이 아니라 `make agents-emit` 이 미러에서 다시 내는 대상이고, golden 게이트가 `make build` **앞**에 있어 재생성을 빠뜨리면 빌드가 선다. M4 의 순서가 그래서 규약이다.
7. **규약을 예시하는 문서가 규약에 걸린다(감사 N3)** — 한 식별자를 표시된 모양과 표시 없는 모양으로 함께 보이면 그 문서는 `ambiguous` 로 정지한다. 이 SPEC 의 `acceptance.md` 가 실제로 그랬다. 해법은 예외가 아니라 적용이다(`spec.md` §3.4).

## §C 사전 점검

```bash
git rev-parse --show-toplevel        # 이 워크트리여야 한다
git branch --show-current            # WT-ac-count-sweep
git rev-parse --short HEAD
ls .moai/reports/t338/               # 실측 3종 존재 확인
```

## §D 제약

- Template-First [HARD] — `spec.md` §7
- 이 SPEC 은 배포되는 새 실행 파일을 만들지 않는다(범위 밖: Go 린트 엔진 편입)
- 시간 예측 금지. 우선순위 라벨과 순서로만 적는다

---

## §E 마일스톤 (되돌리기 어려운 순)

### M0 — 인용 축의 범위: **결정 종결** (선택 A — 규약만 확장, 소급 없음) [Priority: High · kickoff 게이트 선행]

`spec.md` §2.2 가 관측한 두 번째 과다 계상 축 — 남의 SPEC 식별자 인용 — 을 이 카드에서 어디까지 처리할지가 미결이었다. **운영자가 선택 A 로 결정했다**: 예약 토큰 규약(`[RETIRED]` + `[REF]`)과 세 상태 계수기는 그대로 나가되, **인용 축에서는 기존 파일을 한 건도 정규화하지 않는다.** 폐기 축의 처리(종단 이전 정규화 + `completed` 부채 기록)는 이미 세운 대로 유지한다.

**결정을 가른 실측** (운영자 측정, 이 트리, 변경 0 상태. 스크립트 `.moai/reports/t338/pre-terminal-scan.sh`, 출력 `.moai/reports/t338/pre-terminal-scan.txt`)

```
multi-domain files      : 156
  status=completed      : 122
  pre-terminal          :  34
  no spec.md status     :   0
```

**근거**: 폐기 축에서 세운 사실이 인용 축에도 그대로 일반화한다 — **B12 는 그 SPEC 자신의 sync 에서만 돈다.** 156건 중 122건은 `completed` 이고, 다시 돌 sync 가 없다. 그 122건을 지금 정규화해도 미래의 자가검사 결과는 한 건도 달라지지 않고, 착지한 산출물만 다시 쓰인다 — `spec.md` §3.3 이 폐기 축 12건에 대해 세운 것과 동일한 판정이며, 같은 측정이 10배 큰 모집단에서 되풀이된 것이다. 남는 34건만이 다시 센다.

**기각된 두 선택과 각각의 값** (다시 열 때 맨손으로 열지 않도록 값을 남긴다)

| 선택 | 정규화 대상 | 영향 파일 총계 | Tier | 기각 사유 |
|---|---|---|---|---|
| A — 규약만, 소급 없음 **(채택)** | 종단 이전 3건(폐기 축 2 + 별칭 축 1) | 16 | **L** (iter-2 재산정 — 아래) | — |
| B — 인용 축 종단 이전도 정규화 | 3건 + 인용 축 34건 | 45+ | M → **L** | 34건은 각자 자기 sync 에서 저자가 규약을 적용하면 되는 일이다. 지금 한 카드에 몰면 Tier 만 올리고 34건의 저작 맥락은 잃는다 |
| C — 인용 축 전수 소급 | 3건 + 156건 | 167+ | **L** + 카드 분할 | 122건은 값어치 0 이 측정으로 서 있다(위). 카드를 쪼개도 그 122건의 값어치는 0 그대로다 |

**받아들인 비용**: 34건의 종단 이전 파일이 각자 다음 sync 에서 규약을 처음 만난다. 지금 한 번에 몰아서 치르는 대신 시간에 걸쳐 34번 나누어 치른다. 운영자가 이 비용을 명시적으로 받아들였다 — **누락이 아니라 채택된 비용이다.**

**왜 이 비용이 견딜 만한가 — 정정된 근거(감사 D9)**: 종전 문면은 "실패가 **틀린 수**가 아니라 **멈춤**이기 때문" 이라고 적었다. **그 근거는 이 34건의 기본형에는 적용되지 않는다.** 멈춤 성질은 §3.2 의 세 번째 상태가 잡는 **부분 표시**에만 해당하고, 34건의 기본형은 토큰이 하나도 없는 순수 인용이라 `live` 로 읽혀 **조용한 과다 계상으로 남는다**(바로 아래 "남는 것"). 두 문단이 서로를 부정하고 있었다.

정정한 근거는 셋이다. 멈춤 성질은 이 중 어디에도 들어가지 않는다.

1. **저작 맥락** — 각 파일에 `[REF]` 를 달 위치를 아는 사람은 그 SPEC 의 저자다. 지금 한 카드에 몰면 34개 파일의 맥락을 모르는 한 사람이 판정하게 되고, 그것이 §2.1 이 이미 겪은 오탐의 발생 조건이다
2. **값어치의 시점** — 이 34건은 종단 이전이므로 각자 다음 sync 를 돈다. 그때 규약을 적용하면 같은 효과를 맥락과 함께 얻는다. 지금 몰아서 하면 Tier 만 오르고(선택 B: 45+ 파일 → L) 효과는 같다
3. **잔여의 성격이 국지적이다** — 정규화하지 않은 파일의 과다 계상은 **그 파일 자신의 자가검사에만** 나타난다. 다른 SPEC 의 계수에 번지지 않으므로, 미룬 비용이 복리로 불어나지 않는다

**미래의 독자에게**: 34건이 만나는 것은 대부분 "정지" 가 아니라 "저자가 `[REF]` 를 달아야 할 파일" 이다. 실제로 정지가 걸리는 경우(부분 표시)를 "조용히 세도록" 고쳐 없애는 것은 수리가 아니라 이 SPEC 이 막으려던 결함을 되돌리는 일이다 — 정지는 규약을 적용해서 푸는 것이지, 계수기를 무르게 해서 푸는 것이 아니다.

**남는 것(정직하게 적는다)**: 세 번째 상태는 **부분 표시**를 잡는다 — 한 식별자의 등장 일부에만 토큰이 인접한 경우. 토큰이 **하나도 없는** 순수 인용(현재 34건의 기본형)은 §3.2 의 판정에서 `live` 로 읽히므로 정지가 아니라 과다 계상으로 남는다. 즉 34건이 다음 sync 에서 만나는 것은 대부분 "정지"가 아니라 "저자가 `[REF]` 를 달아야 할 파일"이다. 이 잔여를 계수기 쪽에서 메우려면 접두사 단위 애매성 판정이 필요하고, 그것은 정당한 다중 도메인 SPEC(`AC-APO-*` + `AC-DCP-*`)까지 정지시키므로 **네이티브 접두사 선언**이라는 새 기제를 부른다. 이 카드는 그 기제를 만들지 않는다 — 후속 카드 후보로만 기록한다(`spec.md` §6).

**M1 선행 조건 해제**: 이 결정으로 M1 의 문면 범위(`[RETIRED]` + `[REF]` 두 토큰, 소급 없음)와 M5 의 크기(정규화 대상은 재유도로 확정, 부채는 축별 기록)가 함께 고정됐다. 미측정 값은 남아 있지 않다.

### M1 — 규약의 문면을 확정한다 [Priority: High · 가장 되돌리기 어려움]

**결정 대상**: 예약 토큰의 정확한 형태와 배치 규칙.

- 토큰: `[RETIRED]` 와 `[REF]` (대괄호 리터럴). 대소문자 고정, 대괄호 필수 — 산문이 만들어내지 않는 형태여야 제약 3 을 막는다. **토큰은 둘뿐이다**: 별칭 축은 세 번째 토큰이 아니라 `[REF]` 의 뜻을 넓혀 담는다(`spec.md` §2.3).
- 배치 — **인접**: 토큰은 자기가 표시하는 **식별자 등장 바로 뒤**에 오고 사이에는 공백만 둔다. 행이 아니라 등장에 묶는다. 앞에 두는 형태는 규약이 아니다.
- **인접의 두 경계를 M1 에서 못 박는다(감사 N4·N5)** — M1 이 되돌리기 가장 어려운 자리이고, 두 갈래 모두 실측에서 서로 다른 수를 냈다. ① **같은 행에 한한다**: 줄바꿈은 인접을 깬다(허용 개입 문자는 스페이스·탭뿐). 실측 분기 live 3 vs 2. ② **공백·탭이 아닌 문자는 무엇이든 인접을 깬다 — 닫는 백틱 포함**: 코드 스팬 안에서 표시하려면 토큰이 스팬 **안**에 들어와야 한다(`` `AC-SYN-010 [REF]` ``). 실측 분기 53 vs 54. 이 SPEC 의 결정적 예시가 정확히 그 자리에 있으므로 문면이 침묵하면 AC 의 판정이 뒤집힌다
- 예약성: 예약 토큰은 살아 있는 기준의 식별자 등장에 **인접하지 않는다**.
- 세 상태(live / excluded / ambiguous)의 판정 규칙과 각 상태의 처리(`spec.md` §3.2 표)를 그대로 문면화한다. 판정 단위가 **식별자**임을 문면이 명시한다.

**산출**: `.claude/agents/moai/manager-docs.md` B12 절의 개정 문면(초안). 이 시점에서는 아직 커밋하지 않고 M2 의 계약과 함께 확정한다.

**되돌리기 비용이 큰 이유**: 토큰 형태나 부착 규칙이 바뀌면 이미 정규화한 파일과 배포된 절이 모두 갈린다. **인접 규칙은 특히 그렇다** — 행 단위로 되돌리면 이미 정규화한 파일에서 살아 있는 기준이 조용히 사라진다(감사 D1, 재현 54→52). 그래서 이 결정이 M1 에 있고 후속 카드에 있지 않다.

### M2 — 계수기 출력 계약을 확정한다 [Priority: High]

**결정 대상**: 계수기가 무엇을 출력하는가 — 자가검사를 읽는 것은 사람이 아니라 에이전트다.

- live 개수는 정수 1개, 종료코드 0
- ambiguous 는 `AMBIGUOUS <id>` + 종료코드 ≠ 0. **정수를 내지 않는다** — 이것이 §3.2 의 "세지 않고 멈춘다" 의 기계적 형태다. 출력은 애매한 식별자를 하나도 빠뜨리지 않고 이름으로 지목하고, **해소 방법**(그 식별자의 남은 등장 바로 뒤에도 예약 토큰을 단다)을 함께 적는다 — 푸는 법을 말하지 않는 정지는 결함 하나를 다른 결함으로 바꿀 뿐이다(REQ-ACD-003)
- 0 은 기존 절이 이미 규정한 RED 플래그 거동을 유지한다(과소 방향은 이 카드 범위 밖이지만 문면을 훼손하지 않는다)
- 계수기는 절 안의 셸 명령으로 산다. 배포되는 새 스크립트 파일을 만들지 않는다(범위 밖 결정)
- **추출 앵커를 문면에 심는다**: 계수기 명령을 예약 센티널 주석 줄 **한 쌍 사이**에 둔다. 절을 다시 쓰는 사람은 그 두 줄 사이만 건드리지 않으면 된다. 앵커를 산문 구조("B12 헤딩 다음 fenced 블록")에 두면 M1·M2 가 세 상태 표·정지 의무·해소 메시지·출력 예시를 넣는 순간 깨지고, 절이 담은 인라인 코드 명령까지 집을 수 있다(감사 D6). 센티널 문자열 자체를 M2 에서 확정하고, 검증자는 **추출 결과가 정확히 1건이며 비어 있지 않음**을 먼저 단언한다(AC-ACD-005 항목 1)
- **개정 문면에 실재 SPEC ID 를 쓰지 않는다**: `[RETIRED]`/`[REF]` 예시는 합성 ID(`AC-SYN-00N`)로 적는다. 미러는 오늘 SPEC ID 를 0건 담고 있고, CI 가드는 좁은 접두사 부류만 본다(`spec.md` §7)

**산출**: B12 절 개정 문면 확정 + `manager-develop-prompt-template.md` B12 불릿의 대응 개정(현행은 산문 한 줄 — "Verify AC count in CHANGELOG matches `acceptance.md` (SSOT)" — 이므로 세 상태와 정지 의무를 여기에도 적는다).

**AC**: AC-ACD-001 · AC-ACD-002 · AC-ACD-003 · AC-ACD-004

### M3 — 검증자를 세운다 [Priority: High]

- `internal/spec/` 에 저장소 로컬 테스트 1개. 배포 대상 아님.
- (a) 두 `manager-docs.md`(로컬·미러)에서 **센티널 사이** 계수기 명령을 추출 → 각 추출이 **정확히 1건·비어 있지 않음**을 먼저 단언 → 바이트 동일 단언
- (b) fixture corpus 실행: 3 마크업 형태 × 3 상태 + 폐기 주제 산문 함정 + **인접 규칙 6경우**(인접 / 같은 행 비인접 / 앞 배치 / 떠 있는 토큰 / **줄바꿈 개입** / **닫는 백틱 개입**). 뒤의 두 경우는 iter-2 감사(N4·N5)가 실측으로 갈라낸 분기이고, fixture 가 담지 않으면 구현이 자유롭게 고르게 된다. 기대값은 손계산.
- (c) 깊이 1 글롭 `.moai/specs/*/acceptance.md` 전수 실행(`_archive/` 제외) → 커밋된 baseline 스냅샷 `.moai/reports/t338/ac-count-baseline.txt` 와 대조. corpus 크기는 재유도하고 얼리지 않는다
- (d) `manager-develop-prompt-template.md` 쌍의 차이가 171행 1건뿐임을 단언
- (e) `manager-docs.md` 쌍의 `diff` 가 **아무 출력도 내지 않음**을 단언 — (a) 가 보는 것은 명령뿐이라 절의 **산문 절반**(세 상태 표·정지 의무·해소 메시지)이 미러에 실렸는지는 이 단언에만 걸린다(감사 D8). 현 트리 실측 baseline: 두 파일은 바이트 동일
- (f) 개정 절과 미러가 실재 SPEC ID 를 0건 담음을 단언

**뮤턴트로 RED 를 세운다** — (a) 미러 한 글자 변경, (b) fixture 부분 표시 심기 **+ 인접 토큰을 행 끝으로 옮기기**, (c) 임의 파일에 토큰 심기. 모두 죽는 것을 관측한 뒤에만 GREEN 을 주장한다.

> **(b) 의 행 끝 뮤턴트 기대값은 `54` 다 — 종전 문면의 `52` 는 뒤집혀 있었다(감사 N2).** 토큰을 행 끝으로 옮기면 어떤 식별자에도 인접하지 않으므로, **올바른** 인접 계수기는 아무것도 제외하지 않고 스윕과 같은 `54` 를 낸다. `52` 는 이 SPEC 이 존재 이유로 삼아 금지하는 **행 단위 구현만이** 내는 값이다. 즉 종전 문면대로면 올바른 구현이 뮤턴트 절을 만족시킬 수 없고, 뮤턴트 절을 만족시키는 구현은 AC-ACD-001 항목 2(`53`)를 만족시킬 수 없었다 — 동시 만족이 불가능한 요구였다. 실측:
>
> ```console
> $ python3 .moai/reports/t338/repair-scratch/counter.py eol.md sameline    # 행 끝 뮤턴트, 인접 구현
> COUNT 54 (live=54 excluded=0 ambiguous=0)
> $ python3 .moai/reports/t338/repair-scratch/counter.py eol.md line        # 행 끝 뮤턴트, 행 단위 구현
> LINE live=52
> ```
>
> 판독표(세 값이 각각 다른 것을 뜻한다): **54 → 인접 구현이 맞다, 뮤턴트 사망**(AC 는 53 을 요구하므로 실패한다 — 그것이 RED 다) · **53 → 뮤턴트가 살아남았다**, 계수기가 배치 이동을 무시했으므로 인접을 구현하지 않은 것이다 · **52 → D1 회귀**, 행 단위 구현이다. `52` 라는 회귀 서명은 AC-ACD-001 **항목 3** 의 것이고 거기서는 그대로 옳다.

**corpus 실행의 정지 파일 처리**(§3.5): (c) 의 스냅샷은 `COUNT` 항목과 `HALT` 항목을 함께 담고, 검증자는 **상태 변화**를 판정한다. 정지 파일을 corpus 에서 빼거나 0 으로 세는 구현은 이 SPEC 이 없애려는 병을 corpus 층에 되들이는 것이므로 금지다.

**AC**: AC-ACD-001 · AC-ACD-003 · AC-ACD-005 · AC-ACD-006

### M4 — 미러 반영 + 생성물 재생성 + build [Priority: Medium · 기계적이되 **순서가 규약이다**]

**순서를 지키지 않으면 빌드가 선다.** `Makefile:23` 은 `build: agents-emit-check templ-generate` 이고, `agents-emit-check` 는 의도적으로 **읽기 전용**이라 스스로 재생성하지 않는다(`Makefile:31-38`). 그래서 `.codex` TOML 재생성을 빠뜨린 채 `make build` 를 부르면 컴파일에 닿기 전에 sha256 불일치로 **정지**한다. 0.3.0 판의 M4 는 `make build` 만 적었고, 그대로는 완주하지 못한다(감사 N1).

1. `internal/template/templates/.claude/agents/moai/manager-docs.md` 갱신 — **생성 원본은 이 미러다**(로컬 `.claude/` 가 아니다; `agentemit/golden_test.go:30-34`)
2. `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md` 갱신 — **171행 중립화 보존**(verbatim 복사 금지, 해당 절만 편집)
3. **`make agents-emit`** — `.codex/agents/moai/manager-docs.toml` 을 미러에서 다시 낸다. 이 TOML 을 손으로 고치지 않는다
4. **`make build`** — 이제 `agents-emit-check` 가 통과하고, 이어서 `gen-catalog-hashes.go --all` 이 `catalog.yaml` 해시를 다시 낸다
5. `diff` 로 prompt-template 쌍 차이가 171행 1건뿐임을 재확인
6. `git status --short` 로 **셋이 함께 움직였는지** 확인 — 미러 `.md` 2건 + `.codex` TOML 1건 + `catalog.yaml` 1건

**빠뜨렸을 때의 관측 형태**(감사가 미러 한 글자를 바꿔 재현):

```console
golden_test.go:109: .codex/agents/moai/manager-docs.toml: committed artifact differs from emission (sha256 mismatch) — regenerate or stop hand-editing
```

**AC**: AC-ACD-005 항목 3 · 항목 4 · 항목 5 · **항목 6**(codex 생성 표면)

### M5 — 종단 이전 3건 정규화 + 부채 목록 [Priority: Medium · 기계적]

- **대상 목록을 이 문면에서 읽어오지 않는다.** 세 스캔을 다시 돌려 후보를 재유도하고, 종단 상태를 빼고, **남은 플래그를 하나씩 손으로 판정**해 오탐을 뺀 뒤 남는 파일이 대상이다(AC-ACD-007 의 Given). 판정 근거는 식별자 단위로 progress.md `§E.2` 에 적는다
- **plan-phase 가 예상하는 결과(구속 아님)**: 폐기 축 `SPEC-CONFIG-DEAD-SWEEP-001`·`SPEC-UPDATE-DOC-DRIFT-001`, 별칭 축 `SPEC-V3R2-ORC-001`. `SPEC-V3R2-ORC-001` 의 폐기 축 플래그 `AC-ORC-001-05` 는 **오탐이므로 손 판정에서 빠지고**(`spec.md` §8), 같은 파일이 별칭 축에서 스윕 34 → 17 로 들어온다(`spec.md` §2.3). 재유도 결과가 다르면 기대를 고치고 근거를 남긴다
- 각 파일 정규화 전/후 스윕값·계수기값을 기록
- 부채·후보 목록 파일 **`.moai/reports/t338/debt-list.md`** 에 **축별 절**로 기록만 하고 **편집하지 않는다**. 각 절은 자기 수치의 **방향**을 함께 적는다 — 폐기 축 `completed` 는 **하한**(SPEC ID + 식별자), 인용 축 `completed` 는 **상한**(SPEC ID + 실측 출처), 별칭 축 후보는 **상한**(SPEC ID + 실측 출처). **합계를 만들지 않는다**: 하한과 상한을 더한 수는 어느 쪽도 뜻하지 않는다(감사 D4). 상한 두 절은 절 머리에 미검증 후보임을 명시한다. 종단 경로들의 diff 가 비어 있음을 좁힌 `git diff --stat` 으로 확인한다
- baseline 스냅샷 **`.moai/reports/t338/ac-count-baseline.txt`**(M3-c)은 M5 의 변경을 반영해 재생성한다 — 스냅샷 갱신이 M5 뒤에 오는 순서를 지킨다. 스냅샷은 파일별 live 개수와 **상태별 집계**를 함께 담고, 재생성 커밋 메시지에 **어느 파일이 왜 움직였는지** 적는다. 이것은 재생성이 회귀를 흡수하는 것을 **막는** 장치가 아니라 **보이게 하는** 장치다(감사 D7 — 남는 위험으로 기록한다)

**AC**: AC-ACD-007 · AC-ACD-008

---

## §F 영향 파일 전수 열거 (Tier 산정 근거)

| # | 경로 | 성격 |
|---|---|---|
| 1 | `.claude/agents/moai/manager-docs.md` | B12 절 개정 |
| 2 | `internal/template/templates/.claude/agents/moai/manager-docs.md` | 미러 |
| 3 | `.claude/rules/moai/development/manager-develop-prompt-template.md` | B12 불릿 개정 |
| 4 | `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md` | 미러(중립화 보존) |
| 5 | `internal/spec/ac_count_clause_test.go` | 검증자 |
| 6 | `internal/spec/testdata/ac_count/` fixture 3종 | 손계산 기대값(인접 규칙 6경우 포함 — 줄바꿈·백틱 개입 2경우는 감사 N4·N5). 종전 문면의 `account/` 는 오기다(감사 N9) — 검증자 파일명 `ac_count_clause_test.go` 와 맞춘다 |
| 7 | `.moai/reports/t338/ac-count-baseline.txt` | 전수 대조 기준(파일별 live + 상태별 집계) |
| 8-10 | 종단 이전 `acceptance.md` 3건 | 정규화(재유도로 확정 — M5) |
| 11 | `.moai/reports/t338/debt-list.md` | 축별 부채·후보 기록(합계 없음) |
| 12 | `internal/template/catalog.yaml` | `manager-docs.md` 미러의 내용 해시 재생성(`make build`) |
| 13 | `.moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/progress.md` | run-phase 가 `§E.2`/`§E.3` 를 채운다 |
| 14 | `internal/template/templates/.codex/agents/moai/manager-docs.toml` | **B12 절의 세 번째 배포 반송자** — 미러에서 기계 생성. 손으로 고치지 않고 `make agents-emit` 이 다시 낸다(감사 N1) |
| 15 | `.moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/design.md` | Tier L 산출물 — 계수기 계약과 스냅샷 스키마 |
| 16 | `.moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/research.md` | Tier L 산출물 — 실측·선행 조사 |

### Tier 산정 — 산술을 그대로 보인다

행이 아니라 **파일**을 센다(fixture 행은 3파일).

| 행 | 파일 수 | 누계 |
|---|---|---|
| 1-5 (절 2 + 불릿 2 + 검증자 1) | 5 | 5 |
| 6 fixture 3종 | 3 | 8 |
| 7 baseline 스냅샷 | 1 | 9 |
| 8-10 종단 이전 `acceptance.md` | 3 | 12 |
| 11 debt-list | 1 | 13 |
| 12 catalog.yaml | 1 | 14 |
| 13 progress.md | 1 | 15 |
| **14 `.codex` TOML(신규 — 감사 N1)** | **1** | **16** |

**16 파일 > 15 → Tier L.** (`spec-workflow.md:141-142` — M 은 `5-15 files`, L 은 `> 15 files`.) 여기에 Tier L 이 요구하는 산출물 2종(행 15·16)이 더해지므로 최종 열거는 **16 행 / 18 파일**이고, 어느 쪽으로 세도 L 이다.

> **경계는 이미 이름 붙여져 있었다.** 0.3.0 판의 이 자리와 `progress.md §E.1` 은 둘 다 "상한 15 에 닿아 있다 — 여기서 파일이 더 붙으면 L 이다" 라고 적었다. 감사 N1 이 찾은 누락 1건이 정확히 그 한 건이다. **13-15 라는 범위 표기도 함께 버린다**(감사 N8): §F 자신의 계수 규칙으로는 단일 값 **15** 였고, 13 은 파일 수가 아니라 행 수였다. 범위로 적으면 있지도 않은 여유가 있는 것처럼 읽히고, 그 오독이 이번 한 건의 누락을 결정적으로 만들었다.

**Tier L 이 바꾸는 것**(cosmetic 이 아니다):

| 축 | Tier M (종전) | Tier L (현재) |
|---|---|---|
| 산출물 | 3종 + progress | **5종**(+ `design.md` + `research.md`) + progress |
| plan-auditor PASS 임계 | 0.80 | **0.85** |
| plan-audit 반복 상한 | 2 | **3** (`harness.yaml:78`) |
| git 경로 | Route A(main 직행) | **Route B**(phase 별 PR — `manager-git`) |
| 요구/수락 상한 | 각 16 | 각 **25** |

LOC 는 여전히 300 미만이지만 **파일 축이 상위 축**이므로 판정은 파일 수가 낸다 — 0.3.0 판이 M 을 지킬 때 쓴 것과 같은 규칙을, 이번에는 반대 방향으로 적용한 것이다.

> **감사 대비 정정 2건**(iter-1). ① `progress.md` 누락 — run-phase 가 `§E.2`/`§E.3` 를 쓰고 AC-ACD-002·AC-ACD-007 이 그 기록을 요구하므로 영향 파일이다(감사 D13, 수용). ② **`catalog.yaml` 은 감사가 "올바르게 제외됐다" 고 판정했으나 그 판정이 틀렸다.** 엔트리가 이미 있다는 사실은 근거가 되지 못한다 — 엔트리가 담은 것은 **내용 해시**이고, 미러를 고치면 그 값이 바뀐다. 실측: `catalog.yaml:125` 의 `hash:` 가 미러 파일의 `shasum -a 256` 과 일치하며, `make build`(`Makefile:24`)가 `gen-catalog-hashes.go --all` 로 이 값을 다시 낸다. 즉 추적 파일이 함께 바뀐다. (`manager-develop-prompt-template.md` 는 `catalog.yaml` 엔트리가 없어 이 경로에 해당하지 않는다 — grep 0건.)
>
> M0 의 선택 A 는 §F 에 파일을 한 건도 더하지 않았다는 판정은 그대로 유효하다(선택 B 는 45+, C 는 167+ 로 15 를 넘어 L 이 됐을 것이다). 위 두 건은 선택 A 와 무관한 **누락**이다.

## §G 안티패턴

- 계수기 문면을 테스트에 복사하기 → 쌍둥이 드리프트(§B-2)
- 마크업 형태 통일로 문제를 풀려 하기 → 제약 1 을 다시 밟는다
- `retired` / `폐기` 어휘 매칭 → 제약 3 을 다시 밟는다
- 전수 소급 정규화 → §3.3 의 측정이 값어치 0 임을 세웠다
- fixture 에 실재 SPEC ID 넣기 → 템플릿 중립성
- 인용 축 34건이 다음 sync 에서 걸린다고 계수기를 무르게 고치기 → M0 이 채택한 비용을 되돌리는 일이다. 규약을 적용해서 푼다
- 토큰을 **행**에 묶기 → 감사 D1 의 재현이다. 한 행이 살아 있는 식별자와 아닌 것을 함께 담으면 살아 있는 기준이 조용히 사라진다(54→52). 판정 단위는 식별자다
- 별칭 축에 **세 번째 토큰**을 새로 내기 → `[REF]` 의 뜻을 넓히면 토큰도 계수 규칙도 그대로다(`spec.md` §2.3). 토큰을 늘리면 규약 표면만 커지고 잔여(상한 85 미소급)는 그대로다
- 검출기·스캔 출력을 **그대로 결함 목록으로 쓰기** → 세 스캔 모두 불건전하다(폐기 축은 하한, 인용·별칭 축은 상한). 플래그마다 손 판정을 거친다(AC-ACD-007)
- 축이 다른 수치를 **더해서 총계로 적기** → 하한 + 상한은 어느 쪽도 뜻하지 않는다(감사 D4)
- 추출 앵커를 **산문 구조에 걸기** → M1·M2 가 바로 그 산문을 다시 쓴다. 센티널 주석 쌍에 건다(감사 D6)
- 미러 둘만 고치고 **`.codex` TOML 을 잊기** → 배포된 codex 에이전트가 낡은 절을 계속 담는다. 빌드가 서서 알려 주지만, 순서를 모르면 M4 가 실패로 멈춘다(감사 N1)
- **`.codex` TOML 을 손으로 고치기** → 다음 `make agents-emit` 이 덮어쓴다. 원본은 템플릿 미러 `.md` 다
- 정지 파일을 corpus 에서 **빼거나 0 으로 세기** → 정지를 통과로 읽는 것이고, 이 SPEC 이 없애려는 병 그 자체다. 스냅샷의 `HALT` 항목으로 기록하고 **변화**를 판정한다(`spec.md` §3.5, 감사 N3)
- 예시 문서를 규약의 **예외**로 두기 → 자기 예시를 견디지 못하는 규약은 남의 파일에 요구할 자격이 없다(`spec.md` §3.4)
- baseline 스냅샷을 **총계만** 담게 하기 → 재생성 diff 가 "숫자가 바뀌었다" 로만 읽혀 회귀를 흡수한다. 상태별 집계를 담는다(감사 D7)

## §H 교차 참조

- `.moai/reports/t338/measurement.md` — 실측
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier 경계
- `CLAUDE.local.md` §2 / §2.3 — Template-First 와 update 삭제 뿌리
- `.claude/rules/moai/core/verification-claim-integrity.md` — 뮤턴트로 RED 를 세우는 의무의 근거
