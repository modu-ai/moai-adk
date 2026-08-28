---
id: SPEC-AC-COUNT-DISCRIMINATOR-001
title: "AC 개수 자가검사 판별자 — 폐기 기준을 문면 예약 토큰으로 가르고, 애매는 세지 않고 멈춘다"
version: "0.4.1"
status: in-progress
created: 2026-08-28
updated: 2026-08-28
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: ".claude/agents/moai, .claude/rules/moai/development, internal/template/templates/.claude, internal/template/templates/.codex, internal/spec"
lifecycle: spec-anchored
tags: "b12, changelog, acceptance-criteria, self-test, template-first, mutation-testing"
tier: L
era: V3R6
---

# SPEC: AC 개수 자가검사 판별자

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.4.1 | 2026-08-28 | manager-spec | iter-3 감사(PASS 0.91)의 차단 등급 2건 수리. **P1(major)**: `spec.md`·`research.md` 가 이 SPEC 자신의 계수기 아래에서 정지하고 있었다(둘 다 `AMBIGUOUS AC-DCP-010 AC-SYN-010`). **좁히지 않고 적용으로 풀었다** — §3.4 규칙 2 를 두 파일의 위반 지점에 적용했다: §3.1 의 스팬 안/밖 대비를 서로 다른 합성 식별자 두 개로 바꾸고(`acceptance.md` AC-ACD-001 비교표와 같은 이름·같은 뜻 — 이 행이 그 이름들을 다시 적지 않는 것도 같은 규칙이다), `AC-DCP-010` 은 두 파일에서 **모든 등장이 표시되지 않은 채**로 남겼으며(§2.2 결정적 사례 블록 포함), 정지 진단이 이름으로 담은 `AC-SYN-010` 은 §3.4·`research.md` §D.2 에서 표시된 모양으로 다시 쓰지 않도록 서술을 고쳤다. 실측: 두 파일 모두 `rc=0 · ambiguous=0`, corpus `halting=0` 불변. AC-ACD-006 주석도 이동이 숨을 곳이 아님을 문면에 담게 고쳤다. **P2(minor)**: `progress.md` 의 `module:` 에 `internal/template/templates/.codex` 를 더해 나머지 5개 산출물과 일치시켰다(Tier 이관에서 유일하게 빠져 있던 필드). **P4 앞 절반(선택)**: AC-ACD-006 항목 5 에 (e) 추가 — `HALT` 인 파일의 **정지시킨 식별자 집합이 달라지면** 실패(상태 전이만 열거하던 (b)(c)(d) 의 구멍). **P6(선택)**: 0.4.0 행의 N2 실측 표 인용을 §3.1 → `research.md` §D.1 로 정정. P3·P5 는 남긴다. 요구 8 / 수락 8 불변, Tier L 불변 |
| 0.4.0 | 2026-08-28 | manager-spec | iter-2 감사(FAIL 0.72) 수리. **N1(critical) → Tier 재분류**: B12 절의 **세 번째 배포 반송자** `internal/template/templates/.codex/agents/moai/manager-docs.toml` 을 §F 에 넣었다 — `.md` 미러에서 **기계 생성**되고 golden 게이트가 `make build` 앞에 서 있어(`Makefile:23` `build: agents-emit-check …`) 재생성을 빠뜨리면 빌드가 **정지**한다. 영향 파일 15 → **16**, Tier M 상한 15 초과 → **Tier L 로 재분류**(산출물 5종, 임계 0.85, 감사 상한 3, Route B). `plan.md` 가 이미 "여기서 파일이 더 붙으면 L" 이라고 적어 둔 경계가 실제로 넘어간 것이다. **N2(major)**: 행 끝 뮤턴트 기대값을 `52` → **`54`** 로 정정 — 52 는 이 SPEC 이 금지하는 행 단위 구현만이 내는 값이라 항목 2·3 과 동시에 만족 불가능했다(실측 표 `research.md` §D.1). **N3(major)**: 정지 파일의 corpus 처리(§3.5 + REQ-ACD-006)와 **규약을 예시하는 문서가 규약에 걸리지 않는 법**(§3.4)을 함께 세우고, 이 SPEC 자신의 표를 그 규칙대로 고쳤다. **N4·N5**: 인접을 정밀하게 못 박았다 — **같은 행, 공백·탭만, 닫는 백틱을 포함한 어떤 비공백 문자도 인접을 깬다**(둘 다 실측 분기). **N6**: AC-ACD-007 에 비공허 가드. **N7**: 얼린 123/56 을 재유도 + 방향 라벨로 교체(같은 HEAD 에서 이미 **125** 로 이동). **N8-N11** 정리. 요구 8 / 수락 8 불변 |
| 0.3.0 | 2026-08-28 | manager-spec | iter-1 감사(FAIL 0.67) 수리. **D1(critical)**: 예약 토큰을 **행이 아니라 식별자 등장**에 묶었다(인접 규칙) — 행 단위에서는 살아 있는 식별자와 아닌 것이 한 행에 오면 규약이 자기모순이었고 계수기가 살아 있는 기준을 조용히 삼켰다(재현 54→52 rc=0; 그런 행 123행/56파일, `.moai/reports/t338/collision-scan.sh`). REQ-ACD-002·§3.1·§3.2 재작성, 세 상태가 새 단위에서 망라적임을 세웠다. **D2(critical)**: 세 번째 축(같은 파일 안 **별칭**)을 §2.3 으로 세웠다 — 토큰을 늘리지 않고 `[REF]` 의 뜻을 "다른 곳에 선언된 기준을 가리킨다"로 넓혀 담는다. 별칭 축 상한 실측 **85 파일**(`alias-shape-scan.sh`, 참/오탐 각 1건 손 확인). **D3**: `SPEC-V3R2-ORC-001` 의 유일 플래그가 오탐임을 손으로 확인 → 폐기 축 오탐 4파일 / 진짜 14파일·21식별자로 정정, 그 파일은 별칭 축 대상으로 남는다. AC-ACD-007 은 목록 인용을 버리고 **대상 집합을 매 실행 재유도 + 플래그별 손 판정 기록**으로 바꿨다. **D4** 합계 134 삭제(하한+상한 합산 금지, 축별 기록). **D5** 글롭 고정(깊이 1, `_archive/` 제외) + 얼린 수 삭제. **D6** 센티널 앵커 + 추출 1건 단언. **D7** 스냅샷에 상태별 집계 + 재생성 규율, 남는 위험 명시. **D8** `manager-docs.md` 쌍 diff 단언 추가. **D9** M0 비용 근거 정정. **D10** 부채 목록·스냅샷 경로 명명. **D11-D14** 정리. **감사 정정 1건**: `catalog.yaml` 은 §F 에서 빠질 파일이 아니다 — 내용 해시가 박혀 있어 미러 편집 시 함께 바뀐다. 요구 8 / 수락 8 불변, Tier **M 유지**(파일 13-15) |
| 0.2.0 | 2026-08-28 | manager-spec | `plan.md` §E M0 의 미결 결정 종결 — 운영자가 **선택 A**(인용 축 규약만 확장, 소급 없음)를 결정했다. 근거 실측(운영자, 이 트리): 인용 축 후보 156건 중 **122건이 `completed`**, 종단 이전은 **34건**(`.moai/reports/t338/pre-terminal-scan.txt`). 폐기 축에서 세운 "`completed` 는 다시 돌 sync 가 없다" 가 그대로 일반화한다. REQ 층은 **추가 없이 확장**: REQ-ACD-003(정지 메시지가 해소 방법을 말한다)·REQ-ACD-008(부채 목록이 두 축을 담는다). 요구 8 / 수락 8 불변, Tier M 유지 |
| 0.1.0 | 2026-08-28 | manager-spec | 최초 작성 — t338 실측(`.moai/reports/t338/measurement.md`) 위에 작성. 실측이 남긴 두 제약(마크업 앵커 불가·행 단위 표지 매칭 불충분) 위에, 이 판이 세 번째 제약을 측정으로 추가한다: **어휘 매칭은 양방향으로 불건전하다**(플래그된 29개 중 7개가 폐기 주제 SPEC 의 살아 있는 기준). 요구 8 / 수락 8 |

---

## 1. 배경 — 이미 절반은 적혀 있는 규칙

B12 자가검사는 SPEC 의 수락 기준 개수를 토큰 스윕으로 센다.

```bash
grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/<SPEC-ID>/acceptance.md | sort -u | wc -l
```

스윕은 **유효한 기준**과 **폐기 각주 안에서만 살아남은 식별자**를 구별하지 못한다. 폐기 기준을 한 번이라도 낸 SPEC 은 과다 계상되고, 자가검사는 그대로 통과하며, 틀린 수가 CHANGELOG 에 실린다.

**같은 절이 이미 반대 방향을 다룬다.** `manager-docs.md` B12 는 "**A count of 0 is a RED flag, not a pass.** `0 == 0` is a vacuous comparison" 이라고 적는다. 저자는 *과소* 방향 — 아무것도 세지 않는 선택자가 공허하게 통과하는 것 — 을 의식했고, *과다* 방향은 의식하지 않았다. **둘은 하나의 규칙의 두 얼굴이다**: 자가검사가 내는 수는 근거 있는 수여야 하고, 근거를 못 세우면 통과가 아니라 정지여야 한다. 이 SPEC 은 절이 이미 반쯤 말한 그 규칙을 마저 적는다.

## 2. 실측이 이미 결정해 놓은 것 (다시 열지 않는다)

`.moai/reports/t338/measurement.md` 가 이 트리에서 낸 값이다. 재유도하지 않는다.

| 관측 | 값 |
|---|---|
| `acceptance.md` 전수 | 601 |
| 검출기가 플래그한 파일 | 18 |
| 검출기가 플래그한 유령 식별자 | 29 |
| 첫 시험대 | `SPEC-AGENT-EMIT-LINEAGE-001` — 스윕 8 / 유효 7 |

**제약 1 — 마크업 형태에 앵커할 수 없다.** 기준은 `### AC-RIL2-001 — …` 헤딩으로도, `| AC-HYG-001-001 |` 표 셀로도, `| AC-DFF-01 |` 두 자리 행으로도 나타난다(스윕 주석이 스스로 인정한다). 헤딩만 세는 계수기는 표 기반 SPEC 에서 **반대 방향으로** 틀린다.

**제약 2 — 행 단위 표지 매칭으로는 부족하다.** `.moai/reports/t338/overcount-detector.sh` 가 정확히 그 방식이고, 스스로 하한임을 인정한다. 폐기 기록이 여러 행에 걸치고 그중 한 행에 표지가 없으면 놓친다.

### 2.1 이 판이 실측으로 추가하는 제약 3 — 어휘 매칭은 양방향으로 불건전하다

플래그된 29개를 전수 대조하지는 않았으나, **폐기 주제 SPEC** 세 곳에서 7개가 살아 있는 기준임을 확인했다.

```console
$ grep -nE "AC-RA-02([^0-9-]|$)" .moai/specs/SPEC-V3R3-RETIRED-AGENT-001/acceptance.md
46:### AC-RA-02: manager-tdd.md retired stub has all 5 standardized fields (REQ-RA-002)

$ grep -nE "AC-CMR-002([^0-9-]|$)" .moai/specs/SPEC-COMPLETION-MARKER-RETIRE-001/acceptance.md
24:### AC-CMR-002 — Persistent-mode subsystem retired (REQ-CMR-002)
```

`AC-RA-02` 는 **유효한 기준**이다. 제목에 "retired" 라는 단어가 들어 있을 뿐이고, 그 단어는 이 SPEC 의 *주제*이지 이 기준의 *상태*가 아니다. 같은 이유로 `AC-RA-07`·`AC-RA-11`·`AC-RA-14`·`AC-RA-17`·`AC-CMR-002`·`AC-LSPMCP-RETIRE-007` 이 모두 오탐이다(§8 근거).

따라서 실측의 18/29 는 **깨끗한 하한이 아니다.** 과소 검출(제약 2)과 과대 검출(제약 3)을 동시에 담은 혼합 수치다. 오탐을 걷어내면 플래그 집합 안의 진짜 과다 계상은 **14 파일 / 21 식별자**이고(§8 — 오탐 4파일), 여기에 제약 2 가 놓친 미지의 건수가 더해진다.

**하한임이 실측으로 다시 확인됐다.** `SPEC-UPDATE-DOC-DRIFT-001` 은 폐기 식별자를 셋 담는데 검출기는 하나만 플래그했다.

```console
$ grep -nE 'AC-UDD-00[236]([^0-9]|$)' .moai/specs/SPEC-UPDATE-DOC-DRIFT-001/acceptance.md
304:#### AC-UDD-006 — **RETIRED at v0.3.0** (REQ-UDD-005 retired)
574:#### AC-UDD-002, AC-UDD-003 — **RETIRED at v0.3.0**
778:- AC-UDD-002, AC-UDD-003, AC-UDD-006 are retired; their retirement evidence is recorded above and
```

**그래서 검출기 출력은 목록이지 판정이 아니다.** 이 SPEC 은 검출기 목록을 어디에서도 그대로 결함 집합으로 쓰지 않는다 — 정규화 대상은 매 실행 재유도하고 플래그마다 손으로 판정한다(AC-ACD-007).

**이 관측이 설계를 결정한다.** 자연어 어휘를 읽는 판별자는 — 검출기든, 파서든, 사람이든 — 주제와 상태를 구별할 수 없다. 판별자는 산문이 만들어낼 수 없는 **예약된 문면 토큰**이어야 한다.

### 2.2 제약 4 — 인용된 남의 식별자도 똑같이 과다 계상된다 (이 SPEC 자신에서 관측)

이 SPEC 의 `acceptance.md` 를 기존 스윕에 넣으면 **18** 이 나온다. 실제 수락 기준은 **8** 이다.

```console
$ grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/acceptance.md | sort -u | wc -l
      18
```

차이 10건은 전부 **다른 SPEC 의 식별자를 인용**한 것이다(`AC-AEL-008`, `AC-RA-02`, `AC-CMR-002`, `AC-LSPMCP-RETIRE-007`, 합성 fixture id `AC-SYN-002` 등). 폐기와 무관하다 — 애초에 이 파일의 기준이 아니다.

즉 과다 계상에는 폐기 축 말고 **인용 축**이 있다. 카드가 겨눈 폐기 축과, 여기서 처음 관측된 인용 축이다. 그리고 인용 축이 더 클 가능성이 크다. (§2.3 이 **세 번째 축**을 하나 더 세운다 — 축은 셋이다.)

```console
$ bash .moai/reports/t338/multidomain-scan.sh | tail -3
acceptance.md scanned      : 602
files with >=2 AC prefixes : 156
```

156/602 는 **상한**이다 — 한 SPEC 이 정당하게 두 요구 계열을 담는 경우(`SPEC-AGENT-PARALLEL-OPT-001` 의 `AC-APO-*` + `AC-DCP-*`)가 같은 모양으로 읽히기 때문이다. 그래도 폐기 축의 15건과는 자릿수가 다르다. (602 는 이 SPEC 의 `acceptance.md` 가 포함된 값이다.)

**설계상의 함의**: 축들은 같은 형태다 — "이 파일의 살아 있는 기준을 **새로 선언하지 않는** 식별자 등장". 그래서 §3.1 의 규약은 예약 토큰 **집합**으로 일반화한다: 폐기는 `[RETIRED]`, 그 밖의 비선언 등장은 `[REF]`. 계수 규칙(§3.2)은 그대로다.

**범위상의 함의는 종결됐다 — 선택 A(규약만, 소급 없음).** 운영자가 결정했고, 결정을 가른 것은 상한 156 을 상태별로 갈라 본 실측이다.

```console
$ bash .moai/reports/t338/pre-terminal-scan.sh | tail -4
multi-domain files      : 156
  status=completed      : 122
  pre-terminal          :  34
  no spec.md status     :   0
```

156건 중 **122건이 `completed`** 다. §3.3 이 폐기 축 12건에 대해 세운 판정 — B12 는 그 SPEC 자신의 sync 에서만 도므로 `completed` 파일을 고쳐도 미래의 자가검사 결과는 한 건도 달라지지 않는다 — 이 10배 큰 모집단에서 그대로 성립한다. 따라서 인용 축에서도 소급 정규화는 하지 않고, **122건은 폐기 축의 12건과 같은 종류의 상시 부채**로 기록한다(REQ-ACD-008). 기각된 두 선택과 각각의 값(정규화 대상·영향 파일 총계·Tier)은 `plan.md` §E M0 의 표에 남겨 두었다.

남는 34건은 각자 자기 sync 에서 규약을 처음 만난다 — 운영자가 명시적으로 받아들인 비용이다. 다만 **그 만남의 형태를 정확히 적어 둔다**: §3.2 의 세 번째 상태는 **부분 표시**를 잡으므로, 토큰이 하나도 없는 순수 인용은 `live` 로 읽혀 정지가 아니라 과다 계상으로 남는다. 34건 대부분에서 필요한 것은 계수기의 정지가 아니라 저자가 `[REF]` 를 다는 일이다. 이 잔여를 계수기 쪽에서 메우려면 접두사 단위 애매성 판정이 필요하고, 그것은 정당한 다중 도메인 SPEC 까지 정지시켜 **네이티브 접두사 선언**이라는 새 기제를 부른다 — 이 카드의 범위 밖이다(§6).

### 2.3 제약 5 — 같은 파일 안의 **별칭**도 똑같이 과다 계상된다 (감사가 세우고, 이 판이 넓혔다)

세 번째 축이 있다. 한 기준이 **같은 파일 안에서 두 철자로** 불리는 경우다. 정본 철자와 짧은 별칭이 함께 쓰이면 스윕은 둘을 서로 다른 기준으로 센다.

```console
$ grep -c '^## AC-ORC-001-' .moai/specs/SPEC-V3R2-ORC-001/acceptance.md
17
$ grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/SPEC-V3R2-ORC-001/acceptance.md | sort -u | wc -l
      34
```

`AC-01`…`AC-17` 은 `AC-ORC-001-01`…`-17` 의 짧은 철자다(그 파일의 추적 표가 짧은 철자를 쓴다 — `:423`, `:445`). 참값 **17**, 스윕 **34** — 100% 과다 계상이고, 종료코드는 0 이다. **이 축은 폐기도 인용도 아니다**: 어느 철자도 폐기되지 않았고, 남의 SPEC 것도 아니다.

**한 파일의 기벽이 아니다.** 별칭의 기계적 형태 — 숫자 꼬리가 같고 한쪽의 알파 접두사가 다른 쪽의 진접두사인 식별자 쌍 — 을 전수로 재어 상한을 세웠다.

```console
$ bash .moai/reports/t338/alias-shape-scan.sh > .moai/reports/t338/alias-shape-scan.txt
$ grep -c '^=== ' .moai/reports/t338/alias-shape-scan.txt
85
```

**85 는 상한이다.** 형태만 보므로 참 별칭과 우연한 모양이 섞여 있다. 양쪽을 각각 손으로 확인했다.

| 관측 | 파일 | 판정 |
|---|---|---|
| `AC-001 <~ AC-HUM-001` | `SPEC-HUMANIZE-001` (`:300` 이 짧은 철자로 같은 기준을 부른다) | **참 별칭** |
| `AC-SL-001 <~ AC-SL-NF-001` | `SPEC-STATUSLINE-001` | **오탐** — `NF` 는 별개 요구 계열이다 |

즉 별칭 축은 실재하고(둘 이상의 파일에서 확인) 상한은 85 이며, 참값은 재지 않았다 — **전수 판정 없이 이 85 를 결함 목록으로 쓰지 않는다**(§2.1 이 어휘 매칭에서 이미 겪은 오류다).

**규약은 토큰을 늘리지 않고 이 축을 담는다.** `[REF]` 의 뜻을 "남의 SPEC 것"에서 **"이 등장이 기준을 새로 선언하지 않고, 다른 곳에 선언된 기준을 가리킨다"** 로 넓히면 된다 — 다른 곳이 남의 SPEC 이든, 같은 파일 안의 정본 철자든. 별칭 등장에 `[REF]` 를 달면 정본만 세여 값은 17 이 된다. 세 번째 토큰도, 새 계수 규칙도 필요하지 않다.

**범위는 인용 축과 같다(선택 A).** 별칭 축에서도 소급 정규화는 하지 않는다. 상한 85 는 **미검증 후보 목록**으로만 남기고 부채로 단언하지 않는다(REQ-ACD-008).

## 3. 판별자 결정 — 문면 규약이지 파서가 아니다

카드가 물은 것: 판별자는 **문면 규약**인가 **파서**인가.

**파서를 기각한다.** 파서는 식별자가 *어디* 있는지를 알려주지, 그것이 *유효한지*를 알려주지 못한다. 유효성은 구조가 아니라 의미다 — 똑같은 `### AC-XXX-001 …` 헤딩이 한 SPEC 에서는 살아 있는 기준을, 다른 SPEC 에서는 폐기된 기준을 담을 수 있다. 그래서 어떤 파서든 결국 표지를 읽어야 하고, **규약을 없애는 게 아니라 규약을 숨긴 채 마크업 형태 의존성(제약 1)을 얹는다.** 실측이 이미 위험하다고 판정한 의존성이다. 규약 + 얇은 기계 필터가 지배한다.

### 3.1 규약: 예약 토큰은 **행이 아니라 식별자에** 붙는다

예약 토큰은 두 개다.

| 토큰 | 뜻 |
|---|---|
| `[RETIRED]` | 이 등장이 가리키는 기준은 **폐기**됐다 |
| `[REF]` | 이 등장은 기준을 **새로 선언하지 않고**, 다른 곳에 선언된 기준을 가리킨다 — 남의 SPEC 의 기준(인용 축)이거나, 같은 파일 안 정본 철자의 기준(별칭 축) |

**부착 규칙(인접):** 토큰은 자기가 표시하는 식별자 등장의 **바로 뒤**에 오고, 사이에는 공백만 둔다 — `AC-SYN-020 [REF]`. 앞에 두는 형태는 규약이 아니다.

> **이 절의 예시가 실제 식별자 대신 합성 식별자 두 개를 쓰는 이유는 §3.4 규칙 2 다.** 표시된 모양과 표시되지 않은 모양을 같은 문서에서 **함께** 보여야 하므로 두 모양에 서로 다른 이름을 준다 — 표시되는 쪽은 이 파일 안 모든 등장이 표시되고, 표시되지 않는 쪽은 한 등장도 표시되지 않는다. 실제 식별자 `AC-DCP-010` 은 이 파일 안에서 **모든 등장이 표시되지 않은 채**로 유지된다. 그래야 이 파일 자신이 계수기 아래에서 정지하지 않는다 — 규약을 적는 문서에도 규약을 적용한다는 §3.4 의 결정을 이 파일이 자기 문면에서 지키는 자리다. `acceptance.md` 의 AC-ACD-001 비교표가 같은 두 이름을 같은 뜻으로 쓴다.

**"공백만" 의 정확한 뜻 — 두 갈래를 여기서 닫는다(감사 N4·N5).** 종전 문면의 "공백만" 은 두 가지를 열어 두었고, 두 해석 모두 규약에 부합하면서 **서로 다른 수**를 냈다. 둘 다 실측으로 갈렸으므로 문면에서 못 박는다.

1. **같은 행에 한한다.** 식별자와 토큰 사이에 **줄바꿈이 오면 인접이 아니다.** 허용되는 개입 문자는 스페이스와 탭뿐이다. (실측 분기: 한 식별자가 행 끝에 오고 다음 행이 토큰으로 시작하는 6행 fixture 에서, 같은 행 해석은 live 3, 줄바꿈 허용 해석은 live 2 — `.moai/reports/t338/repair-scratch/crossline.md`.)
2. **공백·탭이 아닌 문자는 무엇이든 인접을 깬다 — 닫는 백틱을 포함한다.** 즉 코드 스팬 안에서 표시하려면 토큰이 **스팬 안에** 들어와야 한다(`` `AC-SYN-020 [REF]` ``). 스팬 밖에 두면(`` `AC-SYN-021` [REF] ``) 백틱이 개입해 **아무것도 표시하지 않는다.**

```console
$ python3 .moai/reports/t338/repair-scratch/counter.py inside.md  sameline   # 토큰을 코드 스팬 **안**에 둔 사본
COUNT 53 (live=53 excluded=1 ambiguous=0)
$ python3 .moai/reports/t338/repair-scratch/counter.py outside.md sameline   # 같은 등장, 토큰만 스팬 **밖**에 둔 사본
COUNT 54 (live=54 excluded=0 ambiguous=0)
```

이 구분이 사소하지 않은 이유: 이 SPEC 의 결정적 예시(`AC-DCP-010`)가 실제 파일에서 **정확히 그 자리**에 있다 — 코드 스팬 안, 바로 뒤에 백틱과 괄호가 따른다. 렌더된 마크다운을 읽는 사람이 자연스럽게 쓰는 배치는 스팬 **밖**이고, 그것은 규약상 표시가 아니다. AC-ACD-001 의 Given 이 배치를 문면 그대로 못 박는다.

**예약성:** 예약 토큰은 **살아 있는 기준의 식별자 등장에 인접하지 않는다.**

> **왜 행이 아니라 식별자인가 — 감사가 반증한 지점이다(D1).** 0.2.0 판은 토큰을 **행**에 묶었고("등장하는 모든 행에", "살아 있는 식별자가 있는 행에는 나타나지 않는다"), 그래서 **한 행이 살아 있는 식별자와 살아 있지 않은 식별자를 함께 담으면 두 절반이 서로를 부정했다.** 그 행은 표시할 방법이 없었고, 규약대로 표시하면 계수기가 **살아 있는 기준을 조용히 삼켰다**(54 → 52, 종료코드 0. 의도는 53).
>
> 가설이 아니라 실측이다 — 이 트리에서 재현했다. **수를 얼리지 않는다(감사 N7).**
>
> ```console
> $ bash .moai/reports/t338/collision-scan.sh
> lines carrying >=2 distinct AC prefixes: 125     # iter-1 시점 123 — 같은 HEAD 에서 이동했다
> files containing such a line          : 56
> ```
>
> **이 수는 재유도해서 읽고, 방향을 함께 읽는다.** 0.3.0 판은 `123` 을 문면에 얼렸는데, 같은 HEAD(`da03d9188`)에서 **125** 로 바뀌었다 — corpus 에 이 SPEC 자신의 `acceptance.md` 가 들어 있고, 0.3.0 개정이 그 파일의 충돌 행을 늘렸기 때문이다. **자기가 인용하는 수를 자기가 움직인 것**이며, REQ-ACD-008 이 다른 모든 수치에 요구하는 재유도 규율을 이 SPEC 이 자기 증거에만 적용하지 않고 있었다. 그래서 이 값은 스크립트 이름으로만 인용하고, 읽는 쪽이 다시 돌린다.
>
> **방향**: 이 수는 깨끗한 상한도 하한도 아니다 — 정당하게 두 계열을 인용하는 추적 행을 과다 계상하고(상한 방향), 같은 접두사 두 식별자가 한 행에 오는 충돌은 접두사 구분 스캔에 보이지 않아 과소 계상한다(하한 방향). **D1 의 근거는 이 수에 걸려 있지 않다** — 근거는 아래 결정적 사례 1건의 손 검증이고, 이 수는 그 사례가 외딴 것이 아님을 보이는 규모 지표일 뿐이다.
>
> 결정적 사례: `SPEC-AGENT-PARALLEL-OPT-001/acceptance.md:248` 은 자기 행의 살아 있는 기준 `AC-APO-070`(그 파일 안 유일 등장)과 남의 SPEC 의 `AC-DCP-010` 을 **한 행에** 담는다. 인접 규칙에서는 `AC-DCP-010` 등장에만 토큰이 붙어 그 등장만 표시되고 `AC-APO-070` 은 그대로 살아 센다 — 충돌이 사라진다. (부착된 모양 자체는 §3.1 이 합성 식별자로 보인다 — 이 파일 안에서 `AC-DCP-010` 을 표시된 모양으로 쓰면 이 파일이 정지한다.)

대괄호 형태는 산문이 자연스럽게 만들어내지 않는다. `retired` / `폐기` 같은 낱말과 달리, 주제로 언급되는 일과 상태로 표시되는 일이 형태로 갈린다 — 이것이 제약 3 에 대한 답이다.

### 3.2 계수기: 세 상태 — 판정 단위는 **식별자**다

정의를 먼저 못 박는다. **등장**은 파일 안에서 식별자 패턴이 일치한 한 자리다. 한 등장이 **표시됨**이라는 것은, 그 등장 바로 뒤에 공백만 두고 예약 토큰이 이어진다는 뜻이다(§3.1 의 인접 규칙). 상태는 식별자마다 정한다.

| 상태 | 조건 | 처리 |
|---|---|---|
| live | 그 식별자의 등장 중 **표시된 것이 하나도 없다** | **센다** |
| excluded | 그 식별자의 **모든 등장이 표시됐다** | **제외한다** (`[RETIRED]` 든 `[REF]` 든 동일) |
| ambiguous | **일부 등장만** 표시됐다 | **세지 않고 멈춘다** — 자가검사 실패 |

**세 상태는 이 단위에서 망라적이고 배타적이다.** 모든 식별자는 등장이 하나 이상이고, 각 등장은 표시됐거나 아니거나 둘 중 하나이므로 {하나도 없음, 전부, 일부}가 경우를 남김없이 가른다. 행 단위에서는 이것이 성립하지 않았다 — 한 행이 여러 식별자를 담을 수 있어 판정 단위와 표시 단위가 어긋났고, 그 어긋남이 D1 이다.

**세 번째 상태가 이 설계의 핵심이다.** 현행 계수기는 결과가 둘뿐이라(셈/안 셈) 모든 불확실을 틀린 수로 바꾼다. 세 번째 상태는 제약 2 를 — 여러 등장에 걸친 폐기 기록이 표지를 하나 빠뜨리는 경우를 — **조용한 오계상에서 시끄러운 실패로** 옮긴다. 자가검사의 실패 양식이 "틀린 답"에서 "답 없음"으로 이동하는 것이며, 이것이 §1 이 말한 "근거를 못 세우면 통과가 아니라 정지" 의 기계적 형태다.

**남는 형태 하나(정직하게 적는다):** 어떤 식별자 등장에도 인접하지 않은 떠 있는 토큰은 아무것도 표시하지 않으므로 판정을 바꾸지 않는다. 계수기는 이를 정지 사유로 삼지 않는다.

### 3.3 601개 파일을 소급하지 않는 이유 — 측정으로 결정했다

카드는 "새 SPEC 만 따르는 규약은 실측된 18건을 고치지 않는다. 그 18건을 어떻게 할지 SPEC 이 말해야 한다"고 요구한다. 이 판은 **18건 전부를 고치지 않는다**, 그리고 그 판단은 다음 측정 위에 선다.

```console
$ grep -m1 -H '^status:' .moai/specs/SPEC-CONFIG-DEAD-SWEEP-001/spec.md
.moai/specs/SPEC-CONFIG-DEAD-SWEEP-001/spec.md:status: in-progress
$ grep -m1 -H '^status:' .moai/specs/SPEC-V3R2-ORC-001/spec.md
.moai/specs/SPEC-V3R2-ORC-001/spec.md:status: implemented
$ grep -m1 -H '^status:' .moai/specs/SPEC-UPDATE-DOC-DRIFT-001/spec.md
.moai/specs/SPEC-UPDATE-DOC-DRIFT-001/spec.md:status: draft
```

플래그된 18개 SPEC 중 **15개가 `completed`** 이고, 종단 이전 상태는 **3개**뿐이다(위 세 건). B12 자가검사는 **그 SPEC 의 sync 단계에서만** 돈다. `completed` SPEC 에는 다시 돌 sync 가 없다 — 그 파일을 고쳐도 미래의 자가검사 결과는 한 건도 달라지지 않고, 착지한 산출물만 다시 쓰인다.

**그 종단 이전 3건 중 하나는 오탐이다 — 손으로 판정했다.** `SPEC-V3R2-ORC-001` 의 유일한 플래그 `AC-ORC-001-05` 는 살아 있는 기준이다.

```console
$ sed -n '122,126p' .moai/specs/SPEC-V3R2-ORC-001/acceptance.md
## AC-ORC-001-05 — All 7 retired stubs exist with status=retired (REQ-006)

**Given** M2 has run for all 5 new retirements
**When** I list retired stubs
**Then** all 7 retired stubs exist in template tree:
```

온전한 Given/When/Then 이고, "retired" 는 그 기준의 *주제*다 — §2.1 이 이미 세운 오탐 부류와 같은 종류다. 나머지 둘은 참이다: `SPEC-CONFIG-DEAD-SWEEP-001` 의 세 플래그는 모두 표 셀에 `RETIRED` 판정을 달고 있고(`:13-15`), `SPEC-UPDATE-DOC-DRIFT-001` 은 위 §2.1 이 보인 대로 셋이 폐기다.

그래서 **폐기 축의 종단 이전 정규화 대상은 2건**(`SPEC-CONFIG-DEAD-SWEEP-001`, `SPEC-UPDATE-DOC-DRIFT-001`)이고, `completed` 는 **12건**(15 − 오탐 3파일)이다. `SPEC-V3R2-ORC-001` 은 폐기 축에서 빠지지만 **별칭 축의 대상으로 남는다**(§2.3 — 스윕 34 / 참값 17, 손으로 확인). 즉 종단 이전 정규화 대상은 여전히 3건이되, **셋째의 근거가 검출기 플래그에서 손으로 확인한 별칭 과다 계상으로 바뀌었다.**

그래서: **종단 이전 3건은 규약대로 정규화하고, `completed` 12건(오탐 3파일 제외)은 상시 목록으로 기록하되 편집하지 않는다.** 소급 편집의 값어치가 0임을 주장이 아니라 측정으로 세운 것이다.

### 3.4 규약을 **예시하는** 문서가 규약에 **걸리지** 않는 법 (감사 N3 — 앞 절반)

감사가 이 SPEC 자신에서 결함을 찾았다. 전수 corpus 를 계수기로 돌리면 602 파일 중 **정확히 한 건**이 정지하고, 그 한 건이 이 SPEC 의 `acceptance.md` 다.

```console
$ python3 .moai/reports/t338/repair-scratch/counter.py acceptance.md sameline
AMBIGUOUS AC-SYN-010 AC-SYN-012 (live=22 excluded=0)
```

원인은 규약 자체를 설명하는 표다. AC-ACD-004 의 인접 규칙 표는 한 셀(fixture 열)에서 `AC-SYN-010` 을 토큰이 붙은 **표시된 모양으로** 보이고, 다른 셀(기대 열)에서 같은 식별자를 **표시 없이** 부른다. 그것이 정확히 `ambiguous` 의 정의 — 일부 등장만 표시됨 — 이므로 계수기는 정수를 내지 않고 멈춘다. `AC-SYN-012` 도 같다.

**싼 우회로는 닫혀 있다.** 코드 스팬이나 fenced 블록을 계수에서 빼는 방식은 **마크업 앵커**이고, REQ-ACD-004 가 금지한다(제약 1). 표를 빼는 예외도 같은 부류다.

**결정 — 예시 문서는 자기 예시에 규약을 실제로 적용한다. 예외를 만들지 않는다.** 두 단계로 적는다.

1. **식별자 단위 균일성이 우선이다.** 예시 문서 안에서 한 합성 식별자는 **모든 등장이 표시되거나, 한 등장도 표시되지 않아야** 한다. 표시된 모양을 보이는 행이라면 그 행의 기대 열에서도 같은 식별자를 표시된 채로 부른다. 표시되지 않은 모양을 보이는 행(예: 토큰 앞 배치)이라면 양쪽 다 표시하지 않는다. 이것은 새 규칙이 아니라 §3.2 의 세 상태를 예시 문서에 그대로 적용한 것이다.
2. **한 식별자가 같은 문서에서 표시된 모양과 표시되지 않은 모양을 **동시에** 보여야 한다면, 두 모양에 **서로 다른 합성 식별자**를 쓴다.** 같은 개념을 두 이름으로 부르는 비용이, 예시 문서를 규약의 예외로 만드는 비용보다 싸다.

이 판이 그 규칙을 자기 표에 적용했다 — AC-ACD-004 의 표는 1번만으로 해소되며, 2번을 쓸 필요가 없었다. 적용 후 재측정은 acceptance.md 의 §검증에 있다.

**왜 예외가 아니라 적용인가.** 예시 문서를 corpus 에서 빼면 이 SPEC 은 "규약은 모두에게 적용되지만 규약을 적는 문서에는 적용되지 않는다"를 문면으로 담게 된다. 그 순간 규약은 자기 자신을 견디지 못하는 규약이 되고, 다음 저자가 자기 문서에도 같은 예외를 요구할 근거가 생긴다. 자기 예시에 적용해서 통과하는 규약만이 남의 파일에 요구할 자격이 있다.

### 3.5 corpus 실행이 **정지하는 파일**을 만나면 (감사 N3 — 뒤 절반)

§3.4 는 이 SPEC 자신의 정지를 없앤다. 그러나 **다음 정지가 없다는 보장은 없다** — corpus 는 602 파일이고 그중 어느 파일이든 부분 표시를 만들 수 있다. 그래서 corpus 실행이 정지 파일을 어떻게 다루는지를 문면이 말해야 한다. 0.3.0 판은 이것을 말하지 않았고, 그 침묵은 두 결과 중 하나를 뜻했다 — corpus 실행 자체가 실패해 AC-ACD-006 이 만족 불가능해지거나, 배포된 baseline 이 정지를 영구히 기록하거나.

**결정 — 정지는 조용히 건너뛰지 않고, corpus 실행을 무조건 실패시키지도 않는다. 스냅샷의 일급 상태로 기록하고 *변화*를 판정한다.**

| 스냅샷의 파일별 항목 | 뜻 |
|---|---|
| `COUNT <n>` + 상태별 집계 | 그 파일은 정수를 내며, 값이 이것이다 |
| `HALT <식별자…>` + 사유 | 그 파일은 정지하며, 정지시키는 식별자가 이것이다 |

판정 규칙은 셋이다.

1. **판정 대상은 상태의 *변화*다.** 스냅샷이 `COUNT` 인 파일이 `HALT` 로 바뀌면 실패, `HALT` 이던 파일이 `COUNT` 로 바뀌어도 실패한다(정규화가 일어났는데 스냅샷이 따라오지 않은 것이다). 스냅샷에 **없던** 정지는 언제나 실패다.
2. **건너뛰기는 금지한다.** 정지 파일을 corpus 에서 제외하거나 계수 0 으로 처리하는 구현은 이 조항 위반이다 — 그것이 이 SPEC 이 없애려는 병(정지를 통과로 읽기)을 corpus 층에 다시 들이는 경로다.
3. **정지 항목은 사유와 소유자를 함께 적는다.** 각 `HALT` 줄은 그 파일을 소유한 SPEC 과, 정지가 남아 있는 이유(그 SPEC 이 아직 자기 sync 를 돌지 않았다 등)를 담는다. 사유 없는 `HALT` 항목은 스냅샷을 정지의 쓰레기통으로 만든다.

**이 분리가 정당한 이유**: corpus 검증자와 B12 자가검사는 **다른 것을 판정한다.** B12 는 *그 SPEC 이* 옳은 수를 내는지를 묻고, 정지는 거기서 통과가 아니다 — 그 성질은 그대로다. corpus 검증자는 *계수기가* 트리 전체에서 안정적으로 분류하는지를 묻는다. 후자에게 정지는 정당한 관측 결과이며, 판정해야 할 것은 그 관측이 **움직였는가**다. 정지가 통과로 읽히는 지점은 어디에도 생기지 않는다.

## 4. 요구사항 (GEARS)

**REQ-ACD-001** — The AC-count self-test shall classify every acceptance-criterion identifier found in `acceptance.md` into exactly one of three states — live, retired, or ambiguous — and shall never resolve an ambiguous identifier by counting it.

**REQ-ACD-002** — An acceptance-criterion identifier occurrence that does not declare a live criterion of the file it occurs in shall be marked by a reserved literal token placed immediately after that occurrence **on the same line, separated from it by space and tab characters only, so that a newline or any other intervening character — a closing backtick included — breaks the adjacency** — `[RETIRED]` where the criterion was retired, `[REF]` where the occurrence instead refers to a criterion declared elsewhere, whether in another SPEC or under its canonical spelling in the same file — and no reserved token shall be adjacent to an occurrence that declares a live criterion.

> **표시 단위는 행이 아니라 등장이다.** 행에 묶으면 한 행이 살아 있는 식별자와 살아 있지 않은 식별자를 함께 담을 때 두 절반이 서로를 부정하고, 규약대로 표시한 파일에서 계수기가 **살아 있는 기준을 조용히 삼킨다**(실측 재현: 54 → 52, 종료코드 0). 그런 행은 이 트리에 여러 건 있다 — 수를 얼리지 않고 `collision-scan.sh` 로 재유도하며, 그 수가 어느 쪽 방향으로도 깨끗한 경계가 아니라는 점까지 §3.1 에 적었다.
>
> `[REF]` 는 **세 번째 축(별칭)까지 담도록 뜻을 넓힌 것**이지 토큰을 늘린 것이 아니다(§2.3). 어느 토큰이든 계수 규칙(§3.2)은 동일하다. `[REF]` 축의 **소급 범위는 종결됐다**: 소급하지 않는다(선택 A — §2.2 + `plan.md` §E M0). 규약은 앞으로 저작되거나 다시 sync 를 도는 파일에만 적용된다.

**REQ-ACD-003** — When the self-test detects an identifier some but not all of whose occurrences carry the reserved token, the self-test shall halt CHANGELOG emission and return a blocker report naming every such identifier and stating that adding the reserved token to that identifier's remaining occurrence lines clears the halt, rather than emitting a count.

> 해소 방법을 함께 적는 것이 이 조항의 확장분이다. 정지는 규약을 만나 본 적 없는 파일에서 처음 걸리므로, 무엇이 애매한지와 어떻게 푸는지를 말하지 않는 정지는 **결함 하나를 다른 결함으로 바꿀 뿐이다** — 그리고 그 정지를 "조용히 세도록" 무르게 고치는 것이 이 SPEC 이 막으려는 바로 그 결함이다(`plan.md` §G).

**REQ-ACD-004** — The discriminator shall not decide liveness by matching natural-language retirement vocabulary, and shall not anchor on markdown markup shape.

> 이 요구가 없으면 구현자가 제약 1·3 을 각각 다시 밟는다. 둘 다 실측이 이미 반증한 접근이므로 요구 층에서 금지한다.

**REQ-ACD-005** — The B12 clause and its template mirror shall each delimit the counter command with a reserved sentinel comment line, and the project shall provide a repository-local test that extracts the delimited command verbatim from both files, asserts each extraction yields exactly one non-empty command, asserts the two extractions are byte-identical, and runs the extracted command against a fixture corpus whose expected counts are derived by hand.

> 추출이 load-bearing 이다. 사본을 테스트에 적어 두면 절이 바뀌어도 테스트는 통과한다 — 이 저장소가 `.sh` / `.sh.tmpl` 쌍에서 이미 겪은 쌍둥이 드리프트다.
>
> **앵커를 문면에 못 박는 이유**: 오늘은 절 안에 fenced bash 블록이 하나뿐이라 "B12 헤딩 다음 블록" 이 우연히 통한다. 그러나 M1·M2 가 바로 그 절을 다시 쓰면서 세 상태 표·정지 의무·해소 메시지를 넣고 출력 예시를 더 둘 수 있으므로, 개정 직후에 그 우연이 깨진다. 절은 인라인 코드 명령(`grep -c '<SPEC-ID>' CHANGELOG.md` 등)도 담으므로 느슨한 앵커는 엉뚱한 것을 집는다. **문면이 바뀌어도 앵커가 살아남게 하려면 앵커가 문장이 아니라 센티널이어야 한다** — 절을 다시 쓰는 사람은 센티널 두 줄 사이만 건드리지 않으면 된다. "정확히 1건, 비어 있지 않음" 단언이 앵커가 조용히 0건이나 2건을 집는 경우를 잡는다.

**REQ-ACD-006** — The verifier shall additionally run the extracted counter across every `acceptance.md` matched by the depth-1 glob `.moai/specs/*/acceptance.md`, excluding `.moai/specs/_archive/`, compare the per-file result against a committed baseline snapshot that records each file's live count and its per-state identifier tallies, and fail on any difference; the corpus size shall be re-derived at run time rather than asserted as a frozen literal. The snapshot shall record a halting file as a first-class `HALT` entry naming the halting identifiers, its owning SPEC, and the reason the halt remains, the verifier shall fail on any change of a file's recorded state in either direction and on any halt absent from the snapshot, and the verifier shall not exclude a halting file from the corpus nor record it as a zero count.

> 계수기의 정확성을 주장하지 않고 매 실행 재유도한다 — 카드가 물은 "판별자 자신을 무엇이 검증하는가" 의 답이 이 조항이다.
>
> **글롭을 문면에 못 박는 이유**: 같은 트리에서 세 모집단이 동시에 참이다 — 깊이 1 이 602, 재귀가 603(`.moai/specs/_archive/SPEC-SKILL-001/acceptance.md`), 실측이 낸 값이 601(이 SPEC 자신의 파일 이전). 스냅샷의 판정 기준이 "어떤 차이든 실패" 이므로, 저자와 나중 독자가 모집단을 다르게 읽으면 **테스트가 무엇을 단언하는지가 조용히 달라진다.** 그래서 수를 얼리지 않고 글롭을 얼린다.
>
> **스냅샷이 고무도장이 되는 경로(정직하게 적는다)**: 스냅샷은 재생성할 수 있고, `plan.md` M5 는 재생성을 정상 절차로 둔다. 나쁜 이유로 수가 움직인 실행도 재생성 한 번으로 통과한다 — 이 조항을 실현하는 바로 그 기제가 이 조항의 목적을 무를 수 있다. 상태별 집계를 함께 담게 한 것이 그 완화책이다: 총계만 담으면 재생성 diff 가 "숫자가 바뀌었다" 로만 읽히지만, 상태별로 담으면 **live 가 excluded 로 옮겨 갔는지**가 diff 에 보여 검토가 가능해진다. 재생성 커밋은 어느 파일이 왜 움직였는지 적는다. 이것은 방지가 아니라 **가시화**다 — 남는 위험을 여기 적어 둔다.

**REQ-ACD-007** — When the B12 clause or the prompt-template B12 bullet changes, **every distributed carrier of that clause — each hand-maintained template mirror and each machine-generated artifact emitted from a mirror** — shall change in the same commit, each generated artifact shall be produced by running its own generator rather than hand-edited, the mirror's existing neutralization shall be preserved rather than overwritten by a verbatim copy, and the revised clause shall itself carry no real SPEC identifier.

> **반송자는 둘이 아니라 셋이다(감사 N1).** 0.3.0 판은 "두 정의 지점 모두 템플릿 미러를 가진다"고 적었으나, B12 절은 **세 곳**에 배포된다 — 로컬 `.claude/agents/moai/manager-docs.md`, 템플릿 미러, 그리고 그 미러에서 기계 생성되는 `internal/template/templates/.codex/agents/moai/manager-docs.toml`. 셋째는 계수기 명령까지 통째로 담고 있고, 손으로 고치는 대상이 아니라 `make agents-emit` 이 다시 내는 대상이다.
>
> **빠뜨리면 조용히 지나가지 않는다 — 빌드가 선다.** golden 게이트가 `build` 의 선행 타깃이다(`Makefile:23` `build: agents-emit-check templ-generate`), 그리고 그 게이트는 의도적으로 읽기 전용이라 스스로 재생성하지 않는다(`Makefile:31-38`). 그래서 순서가 **미러 편집 → `make agents-emit` → `make build`** 여야 하고, 재생성을 빠뜨리면 `make build` 가 sha256 불일치로 멈춘다. 감사가 미러 한 글자를 바꿔 재현했다: `golden_test.go:109: .codex/agents/moai/manager-docs.toml: committed artifact differs from emission (sha256 mismatch)`.
>
> 이 조항이 "미러" 가 아니라 "모든 배포 반송자" 를 말하는 이유가 그것이다. 반송자를 손으로 세는 문면은 반송자가 하나 늘 때 조용히 틀리고, 그 침묵이 배포된 에이전트에 낡은 절을 남긴다 — `plan.md` §G 가 첫 번째 안티패턴으로 적은 쌍둥이 드리프트가 한 표면 건너에서 되풀이된 것이다.

**REQ-ACD-008** — The acceptance files whose over-count has been adjudicated by hand and whose `status` is not terminal shall be normalized to the reserved-token convention; every other over-count candidate shall be recorded, without being edited, in a standing list that keeps each axis as its own section, labels each section's figure with the direction of its bound — verified, lower, or upper — and re-derives every figure at record time rather than restating a literal.

> **축별 수치는 더할 수 없다.** 폐기 축의 12건은 검출기가 낸 **하한**이고(제약 2 — `SPEC-UPDATE-DOC-DRIFT-001` 에서 3건 중 1건만 잡혔다), 인용 축의 122건과 별칭 축의 85건은 형태만 보는 스캔이 낸 **상한**이다(정당한 다중 계열 SPEC 과 우연한 모양이 섞인다). 방향이 반대인 두 종류를 더한 합계는 어느 쪽도 뜻하지 않는 수다 — 그래서 이 조항은 합계를 만들지 않고 축별로 적는다.
>
> **상한 축은 "부채"가 아니라 "후보"다.** 전수 판정 없이 122·85 를 결함으로 단언하면 검증되지 않은 결함 주장이 된다 — §2.1 이 어휘 매칭에서 이미 겪은 오류를 목록 층에서 되풀이하는 일이다. 목록은 그 사실을 문면으로 담는다.
>
> 세 축의 `completed` 는 같은 종류의 항목이다: 다시 돌 sync 가 없어 정규화의 값어치가 0 인 파일. 인용 축·별칭 축의 종단 이전 파일은 이 조항의 정규화 대상이 **아니다** — 선택 A(§2.2 · §2.3).

### 요구 ↔ 수락 추적

| REQ | AC |
|---|---|
| REQ-ACD-001 | AC-ACD-001, AC-ACD-003 |
| REQ-ACD-002 | AC-ACD-001, AC-ACD-004 |
| REQ-ACD-003 | AC-ACD-003 |
| REQ-ACD-004 | AC-ACD-002, AC-ACD-004 |
| REQ-ACD-005 | AC-ACD-005 |
| REQ-ACD-006 | AC-ACD-006 |
| REQ-ACD-007 | AC-ACD-005 |
| REQ-ACD-008 | AC-ACD-007, AC-ACD-008 |

## 5. 완료 조건 (카드가 못 박은 것 — 문면 그대로 옮긴다)

> 새 계수기는 **폐기 기준이 있는 SPEC 과 없는 SPEC 양쪽에서 돌려 두 값이 갈리는 것을 관측하기 전까지 미완성**이다. 폐기 이력이 없는 SPEC 에서만 맞는 계수기는 현행과 구별되지 않는다.

첫 시험대는 `SPEC-AGENT-EMIT-LINEAGE-001`(스윕 8 / 유효 7). 대조군은 어느 스캔에도 플래그되지 않은 파일 중 아무거나 — 개수는 얼리지 않고 run-phase 가 재유도해 고른 경로를 기록한다. 기계적 판정 절차는 AC-ACD-001·AC-ACD-002 에 있다.

## 6. 범위 밖

### Out of Scope — `spec.md` 의 REQ 개수 스윕
- 같은 형태의 과다 계상이 `spec.md` 의 REQ 스윕에도 있는지는 실측이 보지 않았다(measurement.md Gaps). **후속 카드 후보로만 기록**하며, 이 SPEC 의 요구사항이 아니다.
- 판별자 규약을 REQ 식별자로 확장하는 일도 하지 않는다.

### Out of Scope — 과소 계상 형제(t241)
- 스윕이 기준을 **덜** 세는 경우는 별도 카드 소관이다. 이 SPEC 은 그 방향을 고치지 않는다.
- 다만 §1 이 기록하듯 둘은 하나의 규칙이므로, 이 SPEC 이 B12 절에 적는 문면이 그 방향을 막지 않도록만 한다.

### Out of Scope — 수락 기준 저작 방식 일반
- 기준을 헤딩으로 쓸지 표로 쓸지, 식별자 자리수를 몇으로 할지는 건드리지 않는다. 판별자가 필요로 하는 것 — 폐기 표시의 형태 — 만 규정한다.
- 마크업 형태 통일은 명시적으로 기각한다(제약 1: 통일 시도 자체가 표 기반 SPEC 을 과소 계상시키는 경로였다).

### Out of Scope — Go 린트 엔진 편입
- `internal/spec` 의 린트 엔진은 `spec.md` 만 읽는다(`parser.go:33,67` — `acceptance.md` 를 로드하는 경로가 없다). 이 SPEC 은 `acceptance.md` 를 린트 엔진의 문서 모델에 편입하지 않는다. 로더·규칙·CLI 표면·테스트를 새로 내는 일은 이 카드의 크기를 넘는다.
- 검증자는 저장소 로컬 테스트로 두고, 배포되는 새 실행 파일을 만들지 않는다.

### Out of Scope — 과다 계상 파일의 소급 정규화(세 축)
- `completed` 상태의 과다 계상 파일은 어느 축이든 편집하지 않는다(§3.3 · §2.2 · §2.3 의 측정 근거 — 폐기 축 12건(하한), 인용 축 122건(상한), 별칭 축은 상한 85 파일 중 미분류). 목록으로만 남긴다.
- 인용 축의 **종단 이전 34건도 이 카드에서 정규화하지 않는다**(선택 A). 각 파일은 자기 다음 sync 에서 저자가 규약을 적용한다.
- **별칭 축도 소급하지 않는다.** 이 카드가 손대는 별칭 파일은 `SPEC-V3R2-ORC-001` 하나뿐이고, 그것은 이미 종단 이전 정규화 대상이라서다. 상한 85 의 나머지는 전수 판정조차 하지 않는다 — 후보 목록으로만 기록한다(REQ-ACD-008).
- 순수 인용(예약 토큰이 하나도 없는 상태)을 계수기 쪽에서 잡기 위한 **접두사 단위 애매성 판정과 네이티브 접두사 선언 기제**는 만들지 않는다. 정당한 다중 도메인 SPEC 을 정지시키지 않으려면 새 선언 문법이 필요하고, 그것은 이 카드의 크기를 넘는다 — 후속 카드 후보로만 기록한다.

## 7. 제약

- **Template-First [HARD]**: 두 정의 지점 모두 템플릿 미러를 가진다. 로컬 편집은 미러 + `make build` 를 같은 커밋에 동반한다. `make build` 는 `internal/template/catalog.yaml` 의 내용 해시를 다시 낸다(`Makefile:24` → `gen-catalog-hashes.go --all`); `manager-docs.md` 미러의 해시가 `catalog.yaml:125` 에 박혀 있으므로 **catalog.yaml 도 같은 커밋에 바뀐다**(`plan.md` §F 12행).
- **B12 절의 배포 반송자는 셋이다 [HARD]** — 로컬 `.claude/agents/moai/manager-docs.md`, 템플릿 미러 `internal/template/templates/.claude/agents/moai/manager-docs.md`, 그리고 미러에서 기계 생성되는 `internal/template/templates/.codex/agents/moai/manager-docs.toml`. 셋째는 계수기 명령을 통째로 담는다(실측: `manager-docs.toml:68` 이하가 "AC count match" 자가검사를 담는다). 생성 원본은 **템플릿 미러** 쪽이다(`agentemit/golden_test.go:30-34` — `templatesDir = "../templates"`, `agentMDRoot = ".claude/agents/moai"`), 로컬 `.claude/` 가 아니다.
  - **순서가 규약이다**: 로컬 편집 → 템플릿 미러 반영 → `make agents-emit` → `make build`. `agents-emit-check` 가 `build` 의 선행 타깃이고 읽기 전용이므로, 재생성을 건너뛴 빌드는 **컴파일 전에 정지**한다. 조용히 낡은 절이 배포되는 경로는 없되, 순서를 모르면 M4 가 빌드 실패로 멈춘다.
  - `.codex` TOML 은 `catalog.yaml` 에 엔트리가 없다(실측: `grep -c '\.codex' internal/template/catalog.yaml` → `0`). 따라서 catalog 해시 경로와는 **별개의 파일**이며, §F 에서 별도 행으로 센다.
- **중립성은 fixture 만이 아니라 개정 문면에도 적용된다**: 미러는 오늘 SPEC ID 를 하나도 담지 않는다(측정: `grep -coE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' internal/template/templates/.claude/agents/moai/manager-docs.md` → `0`). 개정 절이 `[RETIRED]`/`[REF]` 예시를 실재 SPEC ID 로 적으면 배포 템플릿이 내부 상태를 담게 된다. CI 가드는 좁은 부류만 본다(`internal_content_leak_test.go` — `SPEC-(V3R[2-6]|AGENCY|WORKTREE)-`)므로 다른 접두사는 잡히지 않는다. **예시는 fixture 와 같은 합성 ID(`AC-SYN-00N`)를 쓴다.** fixture 도 배포 템플릿 밖(`internal/spec/testdata/`)에 둔다.
- **미러의 기존 중립화 보존**: `manager-develop-prompt-template.md` 의 로컬↔미러 차이는 171행 SPEC-ID 중립화 1건뿐이다(실측). verbatim 복사는 그 중립화를 되돌린다 — 금지.
- **비용**: 검증자는 601개 파일을 훑는다. 전수 스캔의 실측 비용이 기존 테스트 스위트 규모를 넘어서면 baseline 스냅샷 방식을 다시 본다.

## 8. 오탐 판정 근거 (§2.1 의 7건)

| 식별자 | 파일 | 관측된 등장 행 | 판정 |
|---|---|---|---|
| AC-RA-02 | `SPEC-V3R3-RETIRED-AGENT-001` | `:46 ### AC-RA-02: manager-tdd.md retired stub has all 5 standardized fields` | live — "retired" 는 주제 |
| AC-RA-07 | 〃 | `:174 ### AC-RA-07: retired-rejection guard returns proper JSON + exit 2` | live |
| AC-RA-11 | 〃 | `:272 ### AC-RA-11: retired stub body describes reason + replacement + migration` | live |
| AC-RA-14 | 〃 | `:348 ### AC-RA-14: 'moai agents list --retired' subcommand surfaced or deferred` | live |
| AC-RA-17 | 〃 | `:423 ### AC-RA-17: manager-tdd retired stub spawn via Agent() blocked …` | live |
| AC-CMR-002 | `SPEC-COMPLETION-MARKER-RETIRE-001` | `:24 ### AC-CMR-002 — Persistent-mode subsystem retired` | live |
| AC-LSPMCP-RETIRE-007 | `SPEC-LSPMCP-RETIRE-001` | `:84 \| AC-LSPMCP-RETIRE-007 \| REQ-007 \| predecessor superseded \| …` | live — 표 행의 "superseded" 는 판정 대상 서술 |

| AC-ORC-001-05 | `SPEC-V3R2-ORC-001` | `:122 ## AC-ORC-001-05 — All 7 retired stubs exist with status=retired (REQ-006)` + `:124-126` 의 온전한 Given/When/Then | live — "retired" 는 이 기준이 검사하는 *대상*이다 |

앞 세 파일은 통째로 오탐이고, `SPEC-V3R2-ORC-001` 은 유일한 플래그가 오탐이라 폐기 축에서 빠진다 — **오탐 4파일**. 따라서 플래그 집합 안의 진짜 과다 계상은 **14 파일 / 21 식별자**다.

`SPEC-V3R2-ORC-001` 은 폐기 축에서만 빠진다. 별칭 축(§2.3)에서는 손으로 확인한 과다 계상 파일로 남으며, 그 근거는 검출기 플래그가 아니라 §2.3 의 직접 측정이다(스윕 34 / 정본 헤딩 17).
