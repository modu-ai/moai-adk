---
id: SPEC-CODEMAPS-REFRESH-002
title: "Codemaps 최신성 재발 종결: 누락 단위 편입 · 변경 구간 재기술 · 재스탬프"
version: "0.1.4"
status: completed
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".moai/project/codemaps"
lifecycle: spec-anchored
tier: M
tags: "codemaps, freshness, regeneration, accuracy-verification, provenance, stamp, graph-check, t475"
related_specs: [SPEC-CODEMAPS-REFRESH-001, SPEC-GRAPH-FRESHNESS-CADENCE-001, SPEC-CODEMAPS-ACCURACY-001, SPEC-V3R6-GRAPH-FRESHNESS-001, SPEC-V3R6-GRAPH-FRESHNESS-002, SPEC-STAMP-REACHABILITY-001, SPEC-V3R6-DOCS-CODEMAPS-V3-001]
---

# SPEC-CODEMAPS-REFRESH-002 — codemaps 최신성 재발 종결

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-08 | 최초 plan-phase 저작(카드 t475). 기준선은 워크트리 `.claude/worktrees/t475` @ `52f863f36`에서 전수 재측정. 카드 텍스트의 수치 144는 재측정 결과 64로 정정(§A.1). | manager-spec |
| 0.1.1 | 2026-09-08 | plan-audit iter-1(FAIL 0.74) 수리 — D1~D9 전부 적용. D1: 손 열거한 6단위 집합을 폐기하고 §A.3(a)의 P층·F층 후보 규칙으로 교체(실측 62), omission→기록전용 공허 통과 경로 차단. D2: M2의 사전 사본을 재생성보다 앞 단계로 이동 + 선행 조건화. D3: `docs-truth.md`가 생성기 산출물이 아님을 실측으로 확정하고 REQ-CM2-014/AC-CM2-003a로 분리(REFRESH-001에서 계승된 결함을 여기서 닫음). D4: DoD "3항목"→"2항목". D5: §A.3(b) 컷오프 명시(≥6, 템플릿 콘텐츠 제외) → 7구간, `internal/settings` 편입. D6: `check_citations.go`의 정규식·구두점 절삭·blockquote 면제를 추출 규약으로 명명 + `citations` 계층을 판정면으로 연결. D7: `diff -u` 명명 + "변경 없음" 행에 빈 diff 증거 요구. D8: 판정 행에 부모 산문 위치·인용 요구. D9: MUST AC 전수에 4요소 RED-now 셀 부여 + RED-3에 수출 경로 인용. O4 채택(앵커 귀속 = t476 / 2026-09-03). | manager-spec |
| 0.1.2 | 2026-09-08 | D1 근본 원인 수리 + O1 + 날짜 있는 수치 표기. **입도**: 0.1.0의 6개가 재현 불가였던 이유는 패키지 3 + 파일 3의 **두 입도 혼재**임을 확정하고, 후보 규칙을 A층(패키지·앵커 이후 변경, 확장자 무관) / B층(파일·앵커 이후 ADDED) / C층(운영자 지명 이월 `internal/chain`)의 3우주로 재정의 — 실측 A 5 + B 14 + C 1 = **후보 20**, 새 후보 `internal/settings/yamlpatch`·`internal/core/git` 편입. 변경 필터는 후보 한정 전용이며 판정 근거가 아님을 명시. 잔여 42개의 처분(판정 안 함 + verdict.md 목록)을 REQ-CM2-013 ③으로 신설. **O1**: plan §E의 도달성 검증을 `/tmp` 중간 파일에서 `provenance.json` 직독으로 교체(의도가 아니라 효과를 검증) + REQ-CM2-011의 `Where` 전제에 `gate.yaml:72-78`(`blocking: false` @ `:74`) 좌표 인용. **수치 표기**: 64를 `52f863f36`의 날짜 있는 판독으로 [HARD] 명시하고 run 재측정이 대체함을 규정(`origin/develop` 이동분 실측 = described-worthy +2). | manager-spec |
| 0.1.4 | 2026-09-08 | **운영자 결정 — 20 채택(A5 + B14 + C1). 0.1.3의 62는 철회됐다.** 62 지시는 0.1.1의 2층 집합을 대상으로 발신돼 0.1.2의 3우주 재정의와 교차했고, 운영자가 3우주 유도를 채택하며 취소했다. 근거는 인과다 — 게이트를 붉게 만드는 것이 앵커 이후의 변경이므로 카드의 방아쇠와 범위가 일치하고, 무변경 42개는 이 붉음과 인과가 없는 오래된 부채다. **철회된 62에서 그대로 넘어온 것 셋**(수가 아니라 규율이었으므로): ① 판정은 20개 전수에 대해 §A.3(a1) 책임 질문으로 하며 필터는 한정만 한다(AC-CM2-002 조건 8이 "무변경이니 fold"를 FAIL) ② 판정 행 수 = 후보 수를 명령으로 대조(조건 2) ③ §A.3(a1) 생성 파일 처분(`fieldsets_codex_templ.go`). REQ-CM2-013 ③(잔여 42 전수 목록)은 **비선택**으로 복원 — A층 필터를 정당화하는 이관 조건이며 AC-CM2-012가 부재를 FAIL한다. C층은 명시된 예외로 유지(굽힌 규칙은 감사 불가, 명시된 예외는 감사 가능). | manager-spec |
| ~~0.1.3~~ | 2026-09-08 | ~~**운영자 결정 — 62개 전수 판정.**~~ **[SUPERSEDED by 0.1.4]** 이 행의 지시는 철회됐다. 아래 내용은 이력으로만 남긴다. 0.1.2가 P층에 건 변경 필터(48→5)를 폐기하고 규칙을 유도된 그대로 세운다: **P층 48(필터 없음) + F층 14 = 후보 62**, 전부 판정 행을 단다. C층은 해소(`internal/chain`이 P층에 자동 포함). 변경 여부는 **판정 순서를 정하는 보조**로만 남고, "무변경이니 fold"를 AC-CM2-002 조건 8이 명시 금지. 판정 행 수 = 후보 수를 **명령으로 대조**(AC-CM2-002 조건 2: `candidates.txt` 줄 수 vs 판정 표 행 수; 적은 쪽 통과 경로 없음). 62행이 Tier M에 크다는 사실을 **수용된 비용**으로 명기(손 좁힘이 D1을 만들었다는 이유 포함). §A.3(a1)에 **생성 파일 처분** 신설 — `internal/web/fieldsets_codex_templ.go`처럼 기계 생성되는 단위는 생성기 책임으로 접히며, omission일 때 편입 대상은 산물이 아니라 생성 관계다. REQ-CM2-013 ③(잔여 42 목록)은 해소. | manager-spec |

## §A. Problem Statement

### §A.1 기준선 — 이 워크트리에서 재측정한 값

모든 수치는 워크트리 `.claude/worktrees/t475`, HEAD `52f863f36`(= `origin/develop`)에서 실행한 명령의 출력이다.

```
$ go build -o ./bin/moai ./cmd/moai && ./bin/moai graph check ; echo EXIT=$?
codemaps  metric=described-source-diff value=64 threshold=40 verdict=stale
mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=absent
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=absent
citations metric=positive-cited-path-absence value=0 threshold=0 verdict=fresh
  measured from: 25a3212a9 (working-tree-differs-from-stamp)
EXIT=1
```

- **[HARD] `64`는 현재값이 아니라 `52f863f36`에서 취한 날짜 있는 판독이다.** 이 카드의 이력 자체가 그 이유다 — 카드 텍스트의 `144`는 t478·t493 착지로 낡았고, 64도 같은 방식으로 낡는다. **run 시작 시 plan.md §C가 실제 작업 트리에서 다시 재며, 그 실측값이 기준선을 대체한다.** 64를 현재값처럼 이월하지 않는다.

  움직임의 크기도 실측했다: `52f863f36` 대비 `origin/develop`(측정 시점 `19cf21408`)의 described roots 변경은 **15개 파일**이고, described-worthy 술어(`.go` && !`_test.go` && !`testdata/`, `internal/mx/described_worthy.go:13-26`)를 통과하는 것은 **2개**뿐이다(`internal/config/defaults.go`, `internal/config/types.go`). 나머지 13개는 `internal/template/templates/**`의 markdown·yaml로 술어가 배제한다. 즉 **64는 여전히 유효한 판독이며 `+2` 규모로 움직였을 뿐**이다 — 다시 쓰지 않는다.
- mx-index / edges의 `verdict=absent`는 신규 워크트리의 예상 상태다(untracked runtime artifact). stale이 아니다.
- citations 계층은 `fresh` — 인용 경로의 **적극적 부재**는 현재 0건이다. 즉 이 SPEC이 다루는 부정확은 "없는 것을 인용한다"가 아니라 **"있는 것을 기술하지 않는다"** 쪽이다.

provenance 스탬프(`.moai/project/codemaps/provenance.json` 직독):

```
commit_sha:      25a3212a93b4c811cbb22e3c0b34d43571fa65b4
dirty:           false
described_roots: [internal, cmd, pkg]
generated_at:    2026-09-03T18:18:34Z
tree_root:       /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t476
```

### §A.2 이것은 재발이다 — 선행 판정을 다시 열지 않는다

같은 게이트, 같은 처방으로 이미 한 번 닫힌 카드가 있다.

- **SPEC-CODEMAPS-REFRESH-001**(`completed`, 2026-09-02, Tier M) — 재생성 → 정확성 검증 3항목 증거 수출 → merge-base 재스탬프 → 게이트 종결의 절차 정본. 본 SPEC은 그 절차를 **계승**하며 다시 설계하지 않는다.
- **SPEC-GRAPH-FRESHNESS-CADENCE-001**(`completed`) — 카드가 제기하는 임계값 질문을 이미 판정했다: 누적 지표이므로 큰 통합 하나가 이후 모든 측정을 적색으로 만들고, 사람이 재스탬프해야 풀린다. 그 위에서 **임계 40은 통합 축 재유도를 근거로 유지(RETAINED)** 됐다. 본 카드의 운영자 결정은 **보고만 하고 아무것도 바꾸지 않는다**이므로, 본 SPEC은 이 판정을 인용할 뿐 재개하지 않는다(§B.2).
- **SPEC-CODEMAPS-ACCURACY-001**(`completed`) — 팬텀 인용 수리. 실측상 팬텀 6개(`internal/{design,evaluator,factory,migrate,research,state}`)는 codemaps 히트 0이고 트리에도 부재다. **종결됐으며 본 SPEC과 얽히지 않는다.**

### §A.3 스탬프만 갱신하면 안 되는 이유 — 실측된 서술 공백

카드의 중심 요건은 "재생성 전에 codemaps가 현재 구조를 옳게 기술하는지 확인하라"이다. 실측된 공백은 셋이다.

**(a) codemaps가 한 번도 언급하지 않는 실재 단위들 — 손으로 고른 목록이 아니라 규칙이 뱉는 집합.**

> **[정정 — 0.1.1]** 0.1.0은 이 자리에 단위 **6개**를 손으로 열거했다. 그 열거는 재현 불가능했고, 카드가 고치려는 바로 그 종류의 단위를 빠뜨렸다(plan-audit iter-1 D1: `internal/template/agentemit`). **손 열거를 다시 하는 것은 같은 결함을 재생산한다** — 그래서 목록을 고치는 대신 **집합을 만드는 규칙**으로 교체했다. 아래 명령이 후보 집합을 계산하며, 6이라는 수는 폐기됐다.

#### 먼저 입도를 정한다 — 6개가 재현되지 않은 근본 원인

**0.1.0의 여섯은 서로 다른 두 입도를 섞어 놓은 것이었다.** 셋은 패키지(`internal/chain`, `internal/stateanchor`, `internal/template/commandemit`), 셋은 개별 파일(`internal/web/codexmirror.go`, `internal/template/published_skills.go`, `internal/template/skill_mirror_repair.go`)이다. 그래서 **어떤 단일 규칙으로도 재현될 수 없었다** — 파일 셋은 부모 패키지가 히트를 갖고 있어(`internal/web`=5, `internal/template`=19) 히트-0 패키지 집합에 **구조적으로 등장할 수 없고**, 파일 입도 규칙은 전혀 다른 집합을 낸다. 0.1.1의 2층 구성은 이 점을 부분적으로만 반영했다 — 패키지 층에 변경 필터가 없어 두 층의 기준이 서로 달랐다.

**입도 결정: 두 입도를 모두 유지하며, 각각 자기 우주를 명시한다.** codemaps는 패키지 단위로 기술하지만 기술된 패키지 **안**에 앵커 이후 새 역량이 들어올 수 있으므로 둘 다 필요하다.

**[운영자 결정] A층은 "앵커 이후 변경"으로 후보를 한정한다 — 후보는 20개다.** 근거는 인과다: 게이트를 붉게 만드는 것은 **앵커 이후의 변경**이므로, 카드의 방아쇠와 그 범위가 여기서 일치한다. 앵커보다 오래된 히트-0 패키지는 실재하는 부채이되 이 붉음과 인과가 없다.

**필터는 한정할 뿐 분류하지 않는다.** 변경 파일 수는 양적 지표이므로 §A.3(a1)의 판정 근거가 될 수 없다 — 필터는 M1이 **무엇을 들여다볼지** 정하고, fold/omission은 그 안에서 **언제나 단위별 책임 질문**으로 정해진다. 따라서 "앵커 이후 무변경이므로 fold"는 필터를 판정으로 오용한 것이며 AC-CM2-002 조건 4가 실패시킨다.

**필터에서 제외되는 42개는 버려지지 않는다** — 전수 목록이 `verdict.md`에 남고, 그 목록이 없으면 AC-CM2-012가 FAIL한다(REQ-CM2-013 ③). 이것이 좁은 범위를 정당화하는 조건이다: 범위 밖으로 나간 것이 무엇인지 리드가 읽을 수 있어야 좁힘이 손실이 아니라 이관이 된다.

**변경 여부는 `.go` 한정이 아니라 파일 전체로 본다.** 근거는 실측이다: `internal/template/agentemit`의 앵커 이후 변경은 `agents-codex.yaml` **하나뿐**이고, `internal/core/git`은 `status_branch_test.go` 하나뿐이다. described-worthy 술어(`.go` && !`_test.go`)를 후보 필터로 쓰면 **감사가 D1의 물증으로 명명한 `agentemit`이 후보에서 빠진다.** 그 술어는 **게이트 지표(64)를 정의할 뿐 후보를 정의하지 않으며**, 방출 계약이 데이터 파일에 사는 패키지도 codemaps가 기술하지 못하는 역량을 질 수 있다.

#### 후보 규칙 — 3개 우주

히트 집계 규약은 6문서를 연결한 텍스트에 대한 `grep -c -F`, 즉 **적중 행 수**다(같은 행에 두 번 나와도 1).

```bash
cat .moai/project/codemaps/*.md > /tmp/cm.txt
D=.moai/reports/t475/described-roots-diff-since-anchor.txt   # 앵커 대비 변경 집합(수출본)

# 히트-0 패키지 전체 (A층·잔여의 모집합)
go list ./internal/... ./cmd/... ./pkg/... | sed 's|^[^/]*/[^/]*/[^/]*/||' \
  | while read -r p; do /usr/bin/grep -q -F "$p" /tmp/cm.txt || echo "$p"; done > /tmp/Zero.txt

# A층 (패키지 입도) — 히트 0 AND 앵커 이후 그 패키지 경로 하위에 변경 파일 1개 이상(확장자 무관)
while read -r p; do
  awk -v P="$p/" '$2 ~ "^"P {f=1} END{exit !f}' "$D" && echo "$p"
done < /tmp/Zero.txt > /tmp/A.txt

# B층 (파일 입도) — 앵커 이후 ADDED 된 비-테스트 .go 파일 중 히트 0이면서 그 패키지가 A층에 없는 것
#   (패키지가 이미 A층이면 그 패키지 판정에 흡수되므로 중복 계상하지 않는다)
awk '$1=="A" && $2 ~ /\.go$/ && $2 !~ /_test\.go$/ && $2 !~ /testdata\//{print $2}' "$D" \
  | while read -r f; do
      /usr/bin/grep -q -F "$f" /tmp/cm.txt && continue
      /usr/bin/grep -qx -F "$(dirname "$f")" /tmp/A.txt && continue
      echo "$f"
    done > /tmp/B.txt

wc -l /tmp/Zero.txt /tmp/A.txt /tmp/B.txt
```

**C층 (운영자 지명 이월).** 운영자의 in-scope 지시가 명명한 단위 중 A·B 어느 층도 들이지 않는 것을 **명시적 이름으로** 후보에 더한다. `52f863f36`에서 이에 해당하는 것은 **`internal/chain` 하나**다 — 히트 0이지만 앵커 이후 변경이 0이라 A층에 들어오지 않고(`/usr/bin/grep -c 'internal/chain' "$D"` → `0`), 패키지이므로 B층 대상도 아니다. **규칙을 넓혀 한 단위를 들이는 대신 이름으로 이월하고 그 사실을 여기 적는다** — 굽힌 규칙은 감사할 수 없지만 명시된 예외는 감사할 수 있다.

**`52f863f36` 실측: 히트-0 패키지 48, A층 5, B층 14, C층 1 → 후보 20.** 앵커 이후 추가된 비-테스트 `.go` 파일 18개 중 4개(`stateanchor/stateanchor.go`, `commandemit/{commandemit,emit,loader}.go`)는 그 패키지가 이미 A층이라 흡수되고, 남은 14개가 B층이다.

A층 5개와 그 변경 파일:

| A층 단위 | 앵커 이후 변경 |
|---|---|
| `internal/core/git` | `M internal/core/git/status_branch_test.go` |
| `internal/settings/yamlpatch` | `M internal/settings/yamlpatch/yamlpatch.go` |
| `internal/stateanchor` | `A internal/stateanchor/stateanchor.go`, `A .../stateanchor_test.go` |
| `internal/template/agentemit` | `M internal/template/agentemit/agents-codex.yaml` |
| `internal/template/commandemit` | `A .../commandemit.go`, `A .../emit.go`, `A .../loader.go` |

`internal/settings/yamlpatch`와 `internal/core/git`는 감사도 0.1.0도 언급하지 않았던 **새 후보**다.

B층 14개 — 부모는 기술돼 있으나 이 파일들은 히트 0인 단위:

```
internal/cli/{codex_skills_disable,codex_skills_prune,doctor_hook_delivery,
              integration_settings_drift,skills,update_mirror_heal}.go
internal/hook/quality/step_git_env.go
internal/kanban/{prlink_landedref,settings_drift}.go
internal/statusline/state_anchor.go
internal/template/{published_skills,skill_mirror_repair}.go
internal/web/{codexmirror,fieldsets_codex_templ}.go
```

**[HARD] 20개 전수에 판정 행을 단다.** 좁은 범위가 판정을 면제하지는 않는다 — 20은 후보 집합의 크기이지 표본이 아니며, **판정 행 수 = 후보 수**가 AC-CM2-002 조건 2의 기계 판정 대상이다. 20보다 적은 행으로 통과하는 경로는 D1이 명명한 공허한 통과가 다시 열리는 문이다.

이 20개는 **후보**이지 결함이 아니다. 각각을 §A.3(a1)의 책임 질문으로 판정하는 것이 M1이며, 판정 결과가 fold면 기록만 하고 omission이면 편입한다(REQ-CM2-002 / REQ-CM2-004). 0.1.0이 손으로 고른 6개는 전부 이 20 안에 있다(A층 2 + B층 3 + C층 1).

**감사의 쌍둥이 논거 정정.** iter-1 D1은 `agentemit`을 `commandemit`의 "구조적 쌍둥이"라 불렀다. 둘 다 `internal/template` 하위의 기계 방출 패키지인 것은 맞지만 **이력은 다르다** — `commandemit`은 앵커 이후 **추가**됐고(`A internal/template/commandemit/*.go` 3건), `agentemit`은 **수정**만 됐다(`M internal/template/agentemit/agents-codex.yaml`). 쌍둥이 관계는 성립하되 "구조적 쌍둥이"가 함의하는 것보다 약하다. 어느 쪽이든 판정 근거는 이력이 아니라 책임 질문이다.

**부모 히트는 어느 칸에서도 판별력이 없다.** 0.1.0이 인용한 부모 수치(`internal/template`=19, `internal/web`=5, `internal`=236, `internal/harness`=4)는 전부 정확하지만, 그중 어느 것도 fold/omission을 가르지 못한다 — `internal`(236)은 아무것도 가르지 않고, 19나 5는 "부모가 기술돼 있다"만 말할 뿐 **무엇을** 기술하는지 말하지 않는다. 판별식은 §A.3(a1)이다.

### §A.3(a1) 판별식 — 부모의 서술이 그 단위의 책임을 실제로 담고 있는가 (운영자 판정)

접힘/누락 판별의 기준은 **히트 수도 변경 파일 수도 아니다.** 둘 다 양(量)이고 이 판정은 질(質)의 문제다. 판별 질문은 하나이며, 단위마다 개별로 묻는다:

> **부모 패키지의 기존 서술이 이 단위가 지는 책임을 실제로 담고 있는가?**

- **담고 있으면 fold다.** 판정과 그 근거를 기록하고, 그 단위에 대해 codemaps 산문을 **바꾸지 않는다.**
- **담고 있지 않으면 omission이다.** 부모의 서술이 설명하지 못하는 역량(capability)을 그 단위가 지고 있다는 뜻이며, 문서에 편입한다.

명명된 사례 — `internal/template/commandemit`: 스탬프 앵커 시점에 **존재하지 않던** 패키지이고, 파일 5개이며, 명령 소스를 codex 스킬 아티팩트로 발행한다. 부모 `internal/template`의 19개 적중 행 중 어느 것도 이 발행 책임을 서술하지 않는다 → **omission.**

`internal/chain`과 `internal/stateanchor`는 부모가 `internal`(236)뿐이라 부모 히트 수가 애초에 아무 판별력을 갖지 않는다. **이 둘도 같은 책임 질문으로 판정한다** — 부모 히트 수를 근거로 삼지 않는다.

**생성 파일은 그 생성기의 책임으로 접힌다.** 후보 중 다른 파일에서 기계 생성되는 것 — `internal/web/fieldsets_codex_templ.go`가 `fieldsets_codex.templ`에서 생성되는 것이 실제 사례다 — 은 **스스로 책임을 지지 않는다.** 그 단위가 지는 것은 생성기가 정한 계약의 *산물*이므로, 책임 질문의 답은 "생성기(또는 그 생성기를 품은 패키지)의 서술이 이 산물의 존재 이유를 담고 있는가"가 된다. 담고 있으면 **fold**이며, 근거는 그 생성기를 기술한 줄을 인용한다. 담고 있지 않으면 **omission이되 편입 대상은 생성 파일이 아니라 생성 관계**다 — "이 패키지는 `X.templ`에서 `X_templ.go`를 생성한다"가 편입될 서술이지, 생성 산물의 목록이 아니다.

판별 방법: 파일 머리의 생성 표식(`// Code generated ... DO NOT EDIT.`)이나 `.templ`·`.tmpl` 같은 짝 소스의 존재를 확인한다. 이 구분이 없으면 실행자는 생성 파일 앞에서 멈추게 되고, 멈춘 자리에서 임의로 정하는 것이 이 SPEC이 막으려는 형태다.

**판정은 읽은 것을 인용해야 한다.** 책임 질문은 부모 **산문**에 대한 것이므로, 부모 산문을 읽지 않은 판정은 판정이 아니라 패키지 이름에서 유도한 추측이다. 따라서 각 판정 행은 검사한 부모 산문의 위치를 명명한다 — `fold`는 그 책임을 담는 줄을 **인용**하고(`<문서>.md:L<n>` + 인용문), `omission`은 검사한 부모 적중 행의 **범위**를 명시한다(예: "`internal/template` 19행 전수"). 이것은 저장소 자신의 `verification-claim-integrity.md` §1.1 surface 4(권고 전제 주장)를 이 판정에 적용한 것이다.

이 판별식은 기존 접힘 정책 **아래에서 단위를 분류**할 뿐 정책을 재정의하지 않는다. 정책 자체는 §B.2에 따라 소관 밖이다.

**(b) 스탬프 앵커 이후의 구조 드리프트.**

```
$ (스탬프 앵커 25a3212a9 대비 HEAD 의 described roots 변경 집합, 상태별 집계)
  78 A
   1 D
 121 M
```

**재기술 대상 구간의 선정 컷오프(0.1.1에서 신설).** 0.1.0은 여기서도 구간을 손으로 골랐고, 동률인 `internal/settings`(6)를 빠뜨렸다(plan-audit iter-1 D5). 컷오프를 명시한다:

> **described roots 하위이며, `internal/template/templates/**`의 비-Go 템플릿 콘텐츠를 제외한, 앵커 대비 변경 파일 수 ≥ 6인 디렉터리.**

`internal/template/templates/**`를 빼는 이유는 그것이 배포 템플릿 **콘텐츠**이지 이 저장소의 Go 구조가 아니기 때문이다 — codemaps는 후자를 기술한다.

```bash
awk '{print $2}' .moai/reports/t475/described-roots-diff-since-anchor.txt \
  | xargs -n1 dirname | sort | uniq -c | sort -rn \
  | awk '$1>=6 && $2 !~ /^internal\/template\/templates\//'
```

`52f863f36`에서의 출력 — **7구간**:

```
  59 internal/cli
  13 internal/web
  12 internal/statusline
  11 internal/kanban
  10 internal/template
   6 internal/settings
   6 internal/codexwiring
```

`internal/template/commandemit`(5)는 이 컷오프에 걸리지 않지만 §A.3(a)의 후보 규칙 A·B층이 이미 잡는다 — 두 층은 서로 다른 질문에 답한다(여기는 "재기술", 저기는 "편입").

**(c) 패키지 커버리지 — 48은 결함 수가 아니라 히트-0 모집합(Zero)의 크기다.**

```
$ go list ./internal/... ./cmd/... ./pkg/... | wc -l
     136
$ (경로 정규화 후 6문서 연결 텍스트에 대해 exact-path grep -F, 미히트 집계)
missing=48        # 그중 internal/harness/* 하위 12개
```

> **[정정 이력]** 0.1.0은 이 48을 "관측으로만 보고하고 작업 항목이 아니다"라고 적었다. **운영자 결정으로 48은 셋으로 갈린다** — 판정되는 것, 이름으로 이월되는 것, 목록으로 이관되는 것.

`52f863f36`에서 48의 분해(§A.3(a)):

| 분류 | 수 | 처분 |
|---|---|---|
| A층 (앵커 이후 변경 있음) | 5 | M1이 §A.3(a1)로 판정 — fold면 기록, omission이면 편입 |
| C층 (운영자 지명 이월, `internal/chain`) | 1 | 동일하게 M1이 판정 |
| 잔여 (앵커 이후 변경 없음) | 42 | **후보 아님** — 전수 목록을 `verdict.md`에 남긴다(REQ-CM2-013 ③) |

- fold로 판정된 후보 → codemaps 산문 무변경, 증거 파일에 판정과 근거만 기록.
- omission으로 판정된 후보 → 문서에 편입(REQ-CM2-004).
- 잔여 42 → **판정하지 않되 버리지도 않는다.** 이 카드의 방아쇠는 앵커 이후의 드리프트이므로 앵커보다 오래된 서술 공백은 인과가 다르다 — 실재하는 부채이되 이 붉음의 원인이 아니다. 그중 12개가 `internal/harness/*` 하위이며 부모 `internal/harness`(히트 4)로 접힐 개연성이 높지만, **그 판단조차 이 카드가 하지 않는다**(부모 히트 수는 근거가 아니고, 판정하려면 42회의 책임 질문이 더 필요하다). 전수 목록이 `verdict.md`에 남아 리드가 후속 카드를 정한다 — **목록이 없으면 AC-CM2-012가 FAIL한다.** 좁은 범위를 정당화하는 것이 바로 이 이관 조건이다.

**[HARD] omission 판정 단위는 "기록 전용"으로 흘러갈 수 없다.** 0.1.0에는 M1이 6개만 판정하고 나머지가 AC-CM2-007의 기록 전용 분류로 빠지는 경로가 있었다 — 그 경로로 빠진 단위는 편입되지 않은 채 AC 전부가 통과했다(공허한 통과, iter-1 D1). 후보 전수를 M1이 판정하는 지금, 그 경로는 닫혔다: AC-CM2-007의 기록 전용 분류는 **fold로 판정된 후보에만** 적용되며, omission 판정 단위가 편입되지 않은 채 남으면 AC-CM2-004가 FAIL한다.

변하지 않은 것: **접힘 정책 자체는 여전히 소관 밖이다**(§B.2). M1은 기존 정책 아래에서 단위를 분류할 뿐 정책을 바꾸지 않는다.

### §A.4 관측 — 외래 `tree_root`는 신호가 아니다 (설계 의도)

`provenance.json`의 `tree_root`가 **다른 워크트리(t476)** 를 가리킨다. 눈에 띄는 값이지만 **결함도 Gap도 아니다** — codemaps 계층은 이 필드를 의도적으로 대조하지 않는다. `internal/graph/check.go:341-346`의 주석이 그 이유를 직접 말한다:

```
// TreeRoot is deliberately
// NOT matched against this checkout: codemaps is a TRACKED artifact, one
// file replicated to every checkout, so a different tree root is the
// normal state on every machine but the stamper's — hard-matching it
// would disable the check everywhere (the mx-index/edges layers are
// untracked and DO match TreeRoot; see checkMXIndex).
```

코드도 같은 말을 한다: `pv.TreeRoot != projectRoot` 비교는 `check.go:456`(`checkMXIndex`, 445-487)과 `:557`(`MXIndexNeedsRefresh`, 543-565)에만 있고 `checkCodemaps`(315-444) 경로에는 없다. 즉 codemaps에 대해 `MetricWrongTree`가 발화하지 않은 것이 **정상 동작**이며, tracked 아티팩트인 codemaps는 스탬프를 찍은 트리 말고는 어디서 봐도 `tree_root`가 외래인 것이 기본 상태다. 재생성 후 값이 t475를 가리키게 되는 것도 마찬가지다.

이 절은 나중에 이 낯선 경로를 보고 다시 파헤칠 독자를 위해 남긴다 — 위 인용이 그 조사를 미리 닫는다. 어떤 수리 항목도 여기서 파생되지 않는다.

## §B. Scope

### §B.1 In Scope

실행 순서 M1 → M5(상세는 plan.md §F):

- **M1 후보 산출 + 판별** — §A.3(a) 명령으로 후보 집합을 산출하고(저작 시점 실측 20), 그 **전수**를 fold / omission으로 판정한다. 편입 대상과 기술 입도를 확정한다. 이 판정이 뒤집히면 이후 전부가 바뀌므로 가장 먼저 온다.
- **M2 사본 → 재생성 → 편입 확인** — **재생성 전 사본이 첫 단계다**(잃으면 복구 불가). 그다음 `/moai codemaps --force`로 생성기 5문서를 재생성하고, `docs-truth.md`는 손으로 갱신하며, M1이 omission으로 판정한 단위가 실제로 편입됐는지 확인한다. 자동 편입이 안 됐으면 해당 문서를 직접 보정한다.
- **M3 정확성 검증** — SPEC-CODEMAPS-REFRESH-001 REQ-CMR-002~004의 3항목(인용 경로 실존 / `go list` 대조 / 인용 식별자 실존)을 계승해 증거 파일로 수출한다.
- **M4 재스탬프** — merge-base 명시 재스탬프(REQ-CM2-009).
- **M5 종결 + 관측 보고** — 게이트 재판독 + 리드용 관측 리포트(`.moai/reports/t475/verdict.md`).

변경 허용 경로는 `.moai/project/codemaps/**`, `.moai/reports/t475/**`, `.moai/specs/SPEC-CODEMAPS-REFRESH-002/**` 뿐이다.

### §B.2 Boundary decisions

- 이 SPEC은 **문서 재생성 + 검증** SPEC이다. Go 프로덕션 코드 변경은 0줄이다(REQ-CM2-012).
- 검증은 기존 도구(`grep`, `go list`, `test -d/-f`, `merge-base`)로 수행 가능하며 신규 툴링을 요구하지 않는다 — 필요해지는 순간 실행을 멈추고 blocker로 반환한다.
- 스탬프는 콘텐츠 해시가 아니라 커밋 SHA를 기록하므로, 재스탬프는 반드시 병합 생존 리비전을 명명해야 한다(REQ-CM2-009).

### Out of Scope — 접힘 정책 자체의 변경

- `internal/harness/*` 를 포함한 어떤 하위 패키지 묶음의 접힘/전개 정책도 바꾸지 않는다.
- §A.3(c)의 48개 미히트 패키지는 **관측**으로 보고할 뿐 작업 항목이 아니다. 이유: 48 중 상당수가 의도된 접힘일 개연성이 높고(harness 12개가 그 사례), 접힘 정책 변경은 codemaps 문서 설계의 문제이지 최신성 카드의 문제가 아니다.

### Out of Scope — 임계값 40과 게이트 설정

- `gate.yaml`, `internal/graph/check.go` `DefaultThresholds()`, 그 밖의 임계값 관련 설정을 일절 수정하지 않는다.
- 이유: SPEC-GRAPH-FRESHNESS-CADENCE-001이 통합 축 재유도로 40을 이미 유지 판정했고(§A.2), 본 카드의 운영자 결정은 "보고만 하라"이다. 관측(정상 재생성 주기에도 값이 임계를 넘는가 등)은 `.moai/reports/t475/verdict.md`에 기록한다.

### Out of Scope — Go 프로덕션 코드와 신규 툴링

- `internal/`, `pkg/`, `cmd/` 의 Go 코드 변경 0줄. `moai graph` / `moai spec` CLI 자체도 소관 밖이다.
- 검증을 위해 새 Go 코드·새 서브커맨드를 만들지 않는다.

### Out of Scope — 선행 SPEC 재개

- SPEC-CODEMAPS-REFRESH-001을 다시 열지 않는다(`completed`). 본 SPEC은 별개 id를 갖는 후속이다.
- SPEC-CODEMAPS-ACCURACY-001의 팬텀 6개는 §A.2에서 종결 실측됐다 — 다시 다루지 않는다.

## §C. Requirements (GEARS)

- **REQ-CM2-001** (Ubiquitous) — **기준선 재측정.** run 시작 시 the executor shall re-measure `moai graph check`, `provenance.json`, `go list` 패키지 수를 실행하고 그 출력을 진행 기록에 남긴다. §A의 값은 저작 시점 측정이며 재측정이 이를 대체한다.

- **REQ-CM2-002** (event-driven) — **후보 산출과 접힘/누락 판별.** **When** M1 시작 시, the executor shall compute the candidate set by running the §A.3(a) A층·B층 명령 그대로 and adding the C층 named carry-over — 손으로 열거하지 않고, 표본을 뽑지 않는다 — 결과 집합을 증거 파일에 수출하고, **its every member** shall be classified as fold or omission **by the §A.3(a1) discriminator**(부모의 기존 서술이 그 단위의 책임을 담는가). 판정마다 그 책임을 명명하고 **검사한 부모 산문의 위치를 인용**한 근거를 기록한다. **히트 수만 또는 변경 파일 수만을 근거로 삼는 판정은 금지한다** — 둘 다 양적 지표이며 책임 귀속을 말하지 않는다. 특히 **"앵커 이후 무변경이므로 fold"는 A층 필터를 판정으로 오용한 것이며 금지된다** — 필터는 후보를 한정할 뿐 아무것도 분류하지 않는다. 규칙이 내는 후보 수가 저작 시점 관측(A 5 / B 14 / C 1 = 20)과 다르면 실측값을 채택하고 차이를 기록한다 — 어느 경우든 **판정 행 수 = 후보 수**다.

- **REQ-CM2-003** (event-driven) — **재생성 완전성 — 생성기 5문서.** **When** `/moai codemaps --force` 가 실행되면, the regeneration shall produce the generator's five documents — `overview.md`, `modules.md`, `dependencies.md`, `entry-points.md`, `data-flow.md` — from the current tree under described_roots `[internal, cmd, pkg]`. **`docs-truth.md`는 이 집합에 속하지 않는다** — REQ-CM2-014가 별도로 다룬다.

- **REQ-CM2-014** (Ubiquitous) — **`docs-truth.md`는 손으로 유지된다.** `docs-truth.md` is NOT a generator output; the executor shall refresh it by hand and record that refresh separately from REQ-CM2-003's evidence. 검증 경계는 REFRESH-001 REQ-CMR-004가 정한 것을 계승한다 — **§1 에이전트 카탈로그 표 전수**를 `.claude/agents/` 트리 나열과 대조하며 표본 추출은 없다.

  근거(실측): `.claude/skills/moai/workflows/codemaps.md`가 선언하는 산출 파일은 5개이고(`:113-117`), 같은 파일 전체에 `docs-truth`는 **0회** 등장한다(`/usr/bin/grep -c 'docs-truth' … → 0`, exit 1). 파일 자신의 머리말도 손 저작임을 말한다 — *"Canonical Facts Checklist for the Docs-v3 Cohort — Navigation aid, NOT a new SSOT"*(`docs-truth.md:1-3`), 저작 SPEC은 `SPEC-V3R6-DOCS-CODEMAPS-V3-001`.

  **이것은 계승된 결함이며 여기서 닫는다.** REFRESH-001도 같은 "6문서 재생성" 문구를 썼고 `completed`로 닫혔다(plan-audit iter-1 D3 · Residual-risk 5). 지금 닫지 않으면 -003에서 다시 나타난다.

- **REQ-CM2-004** (event-driven) — **누락 단위 편입.** **When** REQ-CM2-002가 어떤 단위를 누락으로 판정했으면, the regenerated document set shall cite that unit; 재생성이 자동으로 편입하지 못한 경우 the executor shall 해당 문서를 직접 보정하고 보정 사실을 증거 파일에 기록한다.

- **REQ-CM2-005** (event-driven) — **변경 구간 재기술.** **When** §A.3(b)의 컷오프 명령이 구간을 산출하면(저작 시점 실측 **7구간**: `internal/cli`, `internal/web`, `internal/statusline`, `internal/kanban`, `internal/template`, `internal/settings`, `internal/codexwiring`), the regenerated documents shall describe those areas from the current tree state, and the executor shall record the pre/post 서술 차이를 **구간마다 `diff -u` 출력으로** 증거 파일에 남긴다. 구간 목록은 손으로 열거하지 않고 컷오프 명령의 출력을 채택한다 — 실행 시점 출력이 7과 다르면 실측값을 쓰고 차이를 기록한다.

- **REQ-CM2-006** (accuracy a) — **인용 경로 실존 — 정본 추출 규약을 명명한다.** **While** the regenerated documents cite paths under `(internal|pkg|cmd)/`, the accuracy verification shall extract those paths **under the canonical convention already implemented in this repository** and export a (path → exists/absent) table over the unique set. **When** a cited path is absent, the executor shall record it in the new-findings 섹션 — 인용 본문을 임의로 지우지 않는다.

  정본 추출 규약(`internal/graph/check_citations.go`), 세 요소 전부:
  1. 정규식 `:23` — `\b(?:internal|pkg|cmd)/[A-Za-z0-9_/.-]*`
  2. 후행 구두점 절삭 `:35` — `citedPathTrailingPunct = ".,;:)]}\"'"`
  3. **blockquote(부정 인용) 면제** — `>` 로 시작하는 줄은 스캔 대상이 아니다. 코드펜스와 mermaid 블록은 **면제가 아니다**(주석 `:18-22`가 그 이유를 적는다: 인용은 그 안에도 산다).

  **면제를 빠뜨리면 표가 거짓말을 한다.** 부존재를 일부러 인용한 줄(제거 기록·rename 이력)이 전부 `absent` new-finding으로 잘못 분류되기 때문이다.

  **기계 측정면이 이미 존재한다.** `moai graph check`의 `citations` 계층이 같은 규약으로 `positive-cited-path-absence`를 threshold 0으로 측정한다. 따라서 이 요건의 **판정**은 그 계층이 담당하고(`--json`의 `citations` 행), 표는 사람이 감사할 수 있는 **증거**다. 둘은 서로를 대체하지 않는다 — 계층은 수를 주고 표는 어느 경로인지를 준다.

- **REQ-CM2-007** (accuracy b) — **패키지 구조 대조.** The accuracy verification shall re-run the §A.3(a)의 히트-0 모집합(Zero) 명령 against the regenerated documents and record every remaining zero-hit package with its M1 판정(fold / omission). **기록 전용 처분은 fold 판정 단위에만 적용된다** — omission으로 판정된 단위가 재생성 후에도 히트 0으로 남아 있으면 그것은 기록 대상이 아니라 REQ-CM2-004 미이행이며 AC-CM2-004가 FAIL한다. 접힘 정책 자체는 여전히 수정 대상이 아니다(§B.2).

- **REQ-CM2-008** (accuracy c) — **인용 식별자 실존 — "식별자"를 추출 명령으로 정의한다.** The accuracy verification shall extract identifiers from `entry-points.md` and `data-flow.md` **by the named command below**, resolve each against its named file/package, and export a (identifier → 명명 위치 → hit/miss) table. 미적중은 기록만 한다.

  **식별자의 정의는 이 명령의 출력이다** — 산문의 형용사가 아니라 명령이 정의한다:

  ```bash
  # 백틱 인라인 코드 중, Go 식별자 형태(대문자 시작 · 또는 pkg.Sym 점표기)만
  for f in entry-points data-flow; do
    /usr/bin/grep -o '`[A-Za-z0-9_.]*`' ".moai/project/codemaps/$f.md" \
      | tr -d '`' | /usr/bin/grep -E '^([a-z][A-Za-z0-9_]*\.)?[A-Z][A-Za-z0-9_]*$'
  done | sort -u
  ```

  경로 형태(`internal/...`)는 REQ-CM2-006의 소관이라 이 정규식이 `/`를 배제해 자동으로 갈라진다. 재생성 **전** 이 명령을 `52f863f36`에서 실행한 결과는 **10행**이다(`AddCommand`, `BacklogPathForRoot`, `ExitCoder`, `PreToolUse`, `RunE`, `Shutdown`, `cli.ResolveExitCode`, `hook.EventType`, …) — 즉 이 추출식은 현재 문서 형식에서 비어 있지 않다. 실행 시점에 0행을 내면 그것은 통과가 아니라 **빈 집합 위의 공허한 통과**이므로, 추출식이 문서 형식과 어긋난 것으로 보고 blocker를 반환한다.

- **REQ-CM2-009** — **스탬프 도달성.** **While** the working branch is not the integration branch, the restamp shall name a merge-surviving revision — `moai graph stamp codemaps --commit <merge-base of HEAD and origin/develop>` — and the executor shall not stamp the branch-local HEAD. (SPEC-STAMP-REACHABILITY-001 계승 — squash 머지 시 스탬프 고아화 방지.)

- **REQ-CM2-010** — **게이트 종결.** **When** `moai graph check` runs after 재생성과 재스탬프, the codemaps layer shall report verdict=fresh (value < 40, expected 0), and no other layer shall report verdict=stale. 신규 워크트리에서 mx-index/edges의 `verdict=absent`는 예상 상태이며 stale이 아니다.

- **REQ-CM2-011** — **증거 독립성.** **Where** the graph freshness gate runs advisory, the run shall not substitute a green gate verdict for the accuracy evidence; REQ-CM2-006~008 shall close with exported evidence regardless of the gate verdict.

  **이 `Where` 절의 전제는 공허하지 않다 — 실측 좌표가 있다**: `.moai/config/sections/gate.yaml:72-78`의 `graph_freshness` 블록이 `enabled: true`(`:73`), **`blocking: false`(`:74`)**, `codemaps_changed_files: 40`(`:78`)이다. 즉 게이트는 지금 차단 없는 권고 모드로 돌고 있으며, 그것이 이 요건이 존재하는 이유다.

- **REQ-CM2-012** (Unwanted) — **범위 금지.** The executor shall not modify Go production code (`internal/`, `pkg/`, `cmd/`), `gate.yaml` 또는 어떤 임계값 설정, 접힘 정책, 또는 SPEC-CODEMAPS-REFRESH-001의 아티팩트.

- **REQ-CM2-013** — **관측 보고.** The run shall export 세 가지 관측 — ① 임계값 40 대비 관측된 값의 거동, ② 후보(A·B·C층) 20개 전수의 fold/omission 분류 요약, ③ **후보에서 제외된 잔여 42개 패키지의 전수 목록** — to `.moai/reports/t475/verdict.md` for the lead. 설정 변경은 동반하지 않는다.

  **③은 선택 항목이 아니다.** 이것이 A층 변경 필터를 정당화하는 조건이다 — 42개는 범위 밖으로 **이관**되는 것이지 사라지는 것이 아니며, 목록이 없으면 "히트-0 48"과 "후보 20"의 간극을 다음 독자가 처음부터 다시 발견한다. AC-CM2-012가 목록 부재를 FAIL로 판정한다. §A.4의 `tree_root`는 설계 의도가 코드 주석으로 확정돼 있으므로 **보고 항목이 아니다**; §A.4의 한 줄이 그 조사를 닫는 유일한 산출물이다.

  ①의 재료는 **누적 속도의 귀속**이다(실측): 현재 앵커 `25a3212a9`를 찍은 것은 REFRESH-001(2026-09-02)이 **아니라** 워크트리 t476이 **2026-09-03T18:18:34Z**에 찍은 스탬프다(`provenance.json`의 `tree_root` + `generated_at` 직독). 즉 64는 **5일**만에 누적된 값이며, CADENCE-001이 산출한 "corrected-40이 약 1.6일에 교차"와 정합한다. 임계값 판단에 필요한 것은 값 자체가 아니라 이 속도이므로, verdict.md는 값과 함께 그 귀속(앵커를 누가 언제 찍었는가)을 적는다.

## §D. Acceptance Criteria

전체 Given-When-Then 시나리오와 RED-now 원장은 `acceptance.md`에 있다. 요약:

| AC | 내용 | MUST |
|----|------|------|
| AC-CM2-001 | 기준선 재측정 기록 (명령 + 출력 + HEAD SHA) | ✓ |
| AC-CM2-002 | 후보 집합을 규칙으로 산출 + 전수 판별 + 부모 산문 인용 | ✓ |
| AC-CM2-003 | 재생성 완전성 — 생성기 5문서 | ✓ |
| AC-CM2-003a | `docs-truth.md` 손 갱신 + 별도 증거 | ✓ |
| AC-CM2-004 | omission 판정 단위가 재생성 결과에 인용됨 | ✓ |
| AC-CM2-005 | 컷오프 산출 구간의 `diff -u` 전후 기록 | ✓ |
| AC-CM2-006 | 정본 규약 인용 경로 표 + `citations` 계층 대조 (accuracy a) | ✓ |
| AC-CM2-007 | 히트-0 모집합 재실행 + 잔여 히트 0 패키지의 판정 분류 (accuracy b) | ✓ |
| AC-CM2-008 | 명명된 추출 명령 기반 식별자 hit/miss 표 (accuracy c) | ✓ |
| AC-CM2-009 | 스탬프 도달성 — `origin/develop` 조상 | ✓ |
| AC-CM2-010 | 게이트 fresh, 타 계층 stale 없음 | ✓ |
| AC-CM2-011 | 범위 위생 — 변경 집합이 허용 3경로에 한정 | ✓ |
| AC-CM2-012 | 관측 리포트 2항목 수출 (설정 무변경 동반) | ✓ |

## §E. Cross-References

- **SPEC-CODEMAPS-REFRESH-001** — 절차 정본(재생성 → 3항목 정확성 검증 → merge-base 재스탬프 → 게이트 종결). REQ-CM2-006~008은 그 REQ-CMR-002~004의 계승이다.
- **SPEC-GRAPH-FRESHNESS-CADENCE-001** — 임계 40 유지 판정의 근거. §B.2가 임계값을 소관 밖으로 두는 이유.
- **SPEC-CODEMAPS-ACCURACY-001** — 팬텀 인용 수리. 실측상 종결.
- **SPEC-STAMP-REACHABILITY-001** — 스탬프 고아화 가드. REQ-CM2-009의 직접 선행.
- **SPEC-V3R6-GRAPH-FRESHNESS-001 / -002** — 신선도 게이트 3계층 모델과 "branch-local HEAD 재스탬프 금지"의 원천.
- **SPEC-V3R6-DOCS-CODEMAPS-V3-001** — codemaps SSOT 문서 집합 원천 생성.
- 카드 t475 — 본 SPEC의 발주 카드. 증거 경로 `.moai/reports/t475/`.
