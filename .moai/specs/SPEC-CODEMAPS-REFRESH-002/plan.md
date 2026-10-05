---
id: SPEC-CODEMAPS-REFRESH-002
title: "plan.md — 누락 단위 판별 · 재생성 · 정확성 검증 · 재스탬프"
version: "0.1.4"
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
phase: "v3.2.0 target"
module: ".moai/project/codemaps"
tier: M
---

# plan.md — SPEC-CODEMAPS-REFRESH-002

## §A. Context

- 카드 t475: "codemaps stale — described-source-diff 144(임계 40). 처방: 재생성. 단 재생성 전에 codemaps가 현재 구조를 옳게 기술하는지 확인할 것 — 스탬프만 갱신하면 낡은 서술에 새 앵커를 붙이는 셈이다."
- 작업면: 워크트리 `.claude/worktrees/t475`, 브랜치 `WT-codemaps-stale`, HEAD `52f863f36`(= `origin/develop`).
- 기준선 전수는 spec.md §A에 있다. 요약: codemaps value=**64**(카드의 144는 낡음), 앵커 `25a3212a9` 이후 described roots 변경 78 A / 1 D / 121 M, 후보 규칙(§A.3(a))이 내는 후보 **A층 5 + B층 14 + C층 1 = 20**(히트-0 패키지 48 중 42는 앵커 이후 무변경 → 후보 아님, verdict.md 목록으로 이관), 재기술 컷오프(§A.3(b))가 내는 **7구간**.
- 절차 정본은 **SPEC-CODEMAPS-REFRESH-001**이다. 본 plan은 그 M1~M4를 재설계하지 않고, 그 앞에 **후보 산출 + 판별 단계(M1)** 를 하나 더 세운 형태다.
- development_mode는 tdd이나 이 SPEC은 Go 코드 변경이 0줄이므로 RED-GREEN-REFACTOR 대신 **판별 → 재생성 → 검증(증거 수출) → 재스탬프 → 게이트**의 검증 주도 절차를 따른다. "테스트"에 해당하는 것이 M3의 기계적 검증 명령과 그 증거 파일이다.

## §B. Known Issues

1. **후보 산출은 기계적이지만 판별은 아니다.** 후보 집합은 §A.3(a) 명령이 내지만, fold/omission은 부모 산문을 읽는 사람의 판단이다 — 히트 수도 변경량도 근거가 아니다. 이것이 M1을 M2 앞에 두는 이유이며, 이 카드에서 가장 뒤집히기 쉬운 판단이다. **후보 20개 전수를 판정해야 한다**는 것이 이 마일스톤의 실제 부피다.
2. **재생성이 판별 결과를 자동으로 반영한다는 보장이 없다.** `/moai codemaps --force`는 생성기의 규칙을 따를 뿐 M1의 판정을 모른다. 누락 판정 단위가 재생성 결과에도 없으면 직접 보정이 필요하다(REQ-CM2-004).

2a. **`docs-truth.md`는 재생성되지 않는다 — 계승된 결함.** 스킬이 선언하는 산출물은 5개이고 `docs-truth`는 그 문서에 0회 등장한다(실측). 그런데 REFRESH-001도 이 SPEC의 0.1.0도 "6문서 재생성"이라고 적었다 — 즉 `ls`가 7항목을 보여주며 통과하는 동안 11KB짜리 손-저작 문서가 조용히 낡는다(iter-1 D3). M2.2가 이 빈틈을 메우며, 지금 닫지 않으면 -003에서 다시 나온다.
3. **기준선이 시간에 민감하다.** 다른 카드가 described roots를 건드리면 value가 다시 오른다. §C 재측정과 §E 종결 직전 재판독으로 흡수한다. 게이트가 advisory이므로 종결 후 재적색이 되어도 본 SPEC의 판정은 종결 시점 측정에 귀속된다.
4. **스탬프 고아화 함정.** 워크트리 브랜치 HEAD(`WT-*`)를 그대로 스탬프하면 squash 병합 시 조상 단절로 reachability 가드가 적색이 된다(SPEC-STAMP-REACHABILITY-001이 닫은 계열). merge-base 명시가 필수다.
5. **48을 두 가지로 오독할 위험 — 양방향이다.** ① 결함 48건으로 읽으면 카드가 접힘 정책 재설계로 번진다(§B.2가 금지). ② 반대로 "관측이니 기록만"으로 읽으면 omission 판정 단위가 편입되지 않은 채 통과한다(iter-1 D1의 공허한 통과). 정확한 지위는 **판정 대기 후보**다 — fold면 기록만, omission이면 편입. 정책은 그대로 두고 단위만 분류한다.
6. **`tree_root`가 외래 워크트리를 가리키지만 결함이 아니다.** codemaps는 tracked 아티팩트라 스탬프를 찍은 트리 말고는 어디서 봐도 `tree_root`가 외래인 것이 기본 상태이며, `check.go:341-346` 주석이 "hard-matching it would disable the check everywhere"라고 그 설계 의도를 직접 적어 두었다(spec.md §A.4). **수리 항목도 관측 보고 항목도 아니다** — 실행자가 이 낯선 경로를 보고 조사를 다시 열지 않도록 여기 적어 둔다.

## §C. Pre-flight (run 첫 턴에 한 배치로 실행)

```bash
go build -o ./bin/moai ./cmd/moai && ./bin/moai graph check ; echo EXIT=$?
cat .moai/project/codemaps/provenance.json
go list ./internal/... ./cmd/... ./pkg/... | wc -l          # 136 기대
git rev-parse --short HEAD
git branch --show-current
git merge-base HEAD origin/develop                          # M4 스탬프 리비전 후보
```

- 재측정에서 codemaps가 이미 `fresh`면(다른 행위자가 재스탬프) M2~M4를 실행하지 말고 리드에 보고한다 — 불필요한 재생성은 최근 착지분을 다시 섞을 위험만 산다.
- `go list` 수가 136이 아니면 값과 함께 리포트에 기록한다(차단 아님 — 문서가 따라가야 할 값이 바뀌었다는 뜻).
- codemaps value가 64가 아니면 실측값을 기준선으로 채택하고 차이를 기록한다. **명령이 이긴다.**

## §D. Constraints

- **변경 허용 경로 3개뿐**: `.moai/project/codemaps/**`, `.moai/reports/t475/**`, `.moai/specs/SPEC-CODEMAPS-REFRESH-002/**`(run-phase progress.md §E.2 갱신 포함). 그 외 전부 금지(AC-CM2-011).
- `gate.yaml`·임계값 설정·Go 코드·접힘 정책·SPEC-CODEMAPS-REFRESH-001 아티팩트: 수정 금지.
- 스탬프는 `--commit <merge-base>` 형식만 허용. bare HEAD 금지(REQ-CM2-009).
- 증거 파일은 `.moai/reports/t475/codemaps-accuracy-verification.md` 단일 파일에 **7개 섹션**으로 수출한다: ① 후보 집합(A·B·C층 산출) + 판별 ② 편입 ③ **`docs-truth.md` 손 갱신**(M2.1 증거와 분리) ④ 구간별 `diff -u` ⑤ 경로 실존 ⑥ 패키지 대조 ⑦ 식별자 hit·miss. 재생성 전 사본은 `.moai/reports/t475/pre-regen/`.
- **최종 증거는 `/tmp`에 두지 않는다** — 감사 시점까지 경로가 살아 있어야 한다. 중간 계산용 `/tmp` 사용(`/tmp/Zero.txt`·`/tmp/A.txt`·`/tmp/B.txt` 등)은 무방하되, **그 결과는 반드시 증거 파일로 수출한다.** 판정 근거로 인용되는 경로는 언제나 `.moai/reports/t475/` 아래여야 한다.
- 관측 리포트는 `.moai/reports/t475/verdict.md`(리드 소비).
- 재생성 실행면은 `/moai codemaps --force`. `--area` 부분 생성은 목적과 어긋나므로 쓰지 않는다.
- 신규 Go 코드·신규 서브커맨드 금지. 필요해지면 실행을 멈추고 blocker로 반환한다.

## §E. Self-Verification (§F 종결 직전 실행)

```bash
./bin/moai graph check ; echo EXIT=$?         # codemaps fresh, citations fresh, 타 계층 stale 없음
cat .moai/project/codemaps/provenance.json    # commit_sha 를 읽어 조상 여부를 확인(M4 참조)
git status --porcelain                        # 허용 3경로 밖 변경 0
ls .moai/reports/t475/                        # 증거 7섹션 파일 + pre-regen/ + verdict.md
```

각 항목은 verification-claim-integrity 5절 형식(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)으로 보고한다. 증거는 명령 + 그 명령의 출력 + 측정한 HEAD SHA를 함께 명명한다.

## §F. Milestones

마일스톤은 **뒤집힐 가능성이 큰 판단부터** 배치했다. M1의 판정이 바뀌면 M2~M5 전부가 바뀐다.

### M1 — 후보 산출 + 접힘/누락 판별 (가장 뒤집히기 쉬운 결정)

**M1.0 — 후보를 규칙으로 산출한다. 손으로 열거하지 않는다.** spec.md §A.3(a)의 A층·B층 명령을 그대로 실행하고 C층(운영자 지명 이월, `internal/chain`)을 더한 뒤, 그 합집합(= 후보)을 `.moai/reports/t475/candidates.txt`(한 줄에 하나)로 수출한 뒤 같은 집합을 `.moai/reports/t475/codemaps-accuracy-verification.md` §① 상단에 수출한다.

```bash
wc -l /tmp/Zero.txt /tmp/A.txt /tmp/B.txt   # 저작 시점 실측: 48 / 5 / 14 (+C층 1 → 후보 20)
```

값이 다르면 **실측값을 채택하고 차이를 기록한다** — 트리가 움직였다는 뜻이지 규칙이 틀렸다는 뜻이 아니다.

**[운영자 결정] A층 필터는 후보를 한정할 뿐 아무것도 분류하지 않는다.** 게이트를 붉게 만드는 것이 앵커 이후의 변경이므로 카드의 방아쇠와 범위가 일치한다 — 그것이 필터의 근거다. 그러나 필터는 M1이 **무엇을 들여다볼지**만 정하고, fold/omission은 그 안에서 언제나 단위별 책임 질문으로 정해진다. **"무변경이니 fold"는 필터를 판정으로 오용한 것**이며 AC-CM2-002 조건 4와 8이 실패시킨다.

**[HARD] 20개 전수에 판정 행을 단다 — 좁은 범위가 판정을 면제하지 않는다.** 20은 후보 집합의 크기이지 표본이 아니다. 판정 행이 20보다 적으면 AC-CM2-002 조건 2가 기계적으로 FAIL한다(D1의 공허한 통과가 다시 열리는 문).

**필터에서 빠진 42개는 버려지지 않는다.** 전수 목록을 `verdict.md`에 남긴다(REQ-CM2-013 ③) — 목록이 없으면 AC-CM2-012가 FAIL한다. **이 이관이 좁은 범위를 정당화하는 조건이다**: 범위 밖으로 나간 것이 무엇인지 리드가 읽을 수 있어야 좁힘이 손실이 아니라 이관이 된다.

**M1.1 — 후보 전수를 판정한다.** 각 후보에 대해 fold / omission을 판정하고 근거를 기록한다. **판별식은 운영자 판정으로 확정돼 있다(spec.md §A.3(a1)) — 여기서 다시 정하지 않는다:**

> 부모 패키지의 기존 서술이 이 단위가 지는 책임을 실제로 담고 있는가?

- 담고 있으면 **fold** — 판정과 근거만 기록하고 codemaps 산문은 그 단위에 대해 손대지 않는다.
- 담고 있지 않으면 **omission** — 문서에 편입한다.

`internal/template/commandemit`은 운영자가 명명한 **omission 사례**다(앵커 시점 부재, 5파일, 명령 소스를 codex 스킬 아티팩트로 발행 — 부모 `internal/template`의 19개 적중 행이 이 발행 책임을 서술하지 않음). 종전 초안이 "부모는 기술돼 있으나 변경량이 크니 보수적으로 omission"이라고 적었던 것은 **폐기됐다** — 변경량은 근거가 아니다. 판정 근거는 책임 귀속이다.

`internal/chain` / `internal/stateanchor`는 부모가 `internal`(236)뿐이라 부모 히트가 판별력을 갖지 않는다 — 같은 책임 질문으로 개별 판정한다.

기록할 근거는 두 가지를 **함께** 담아야 한다(AC-CM2-002가 이분 판정한다):

1. 그 단위가 지는 **책임의 명명**. "히트 0" 또는 "5파일 변경"만 적힌 행은 근거가 아니다.
2. **검사한 부모 산문의 위치** — `fold`는 그 책임을 담는 줄을 인용(`<문서>.md:L<n>` + 인용문), `omission`은 검사한 부모 적중 행의 범위를 명시(예: "`internal/template` 19행 전수"). 부모 산문을 읽지 않은 판정은 패키지 이름에서 유도한 추측이며, 저장소 자신의 `verification-claim-integrity.md` §1.1 surface 4가 금지하는 형태다.

**후보가 20이면 판정 행도 20이다.** 규모가 부담이면 그것은 규칙을 좁힐 이유가 아니라 리드에게 보고할 사실이다 — 손으로 좁히는 순간 iter-1 D1이 재발한다. 같은 부모 산문을 공유하는 묶음(예: `internal/cli` 하위 B층 파일 6개)은 **하나의 부모 인용을 공유하는 여러 행**으로 기록할 수 있다(행은 각각, 인용은 1회).

편입 입도(패키지 1행인지 섹션인지)도 여기서 정한다.

M1 종료 조건: 후보 집합 파일이 수출됐고, 그 집합의 **모든** 원소에 대한 판정 행이 증거 파일 §판별에 존재한다.

### M2 — 재생성 + 편입 확인 (사본이 **먼저**다)

> **[HARD] 단계 순서는 목록 순서가 아니라 의무다.** 0.1.0은 사본 명령을 재생성 **아래**에 두었고, 위에서 아래로 따르면 "재생성 전 사본"이 재생성 **후**에 떠졌다(plan-audit iter-1 D2). AC-CM2-005는 사본 부재를 **복구 불가 FAIL**로 규정한다 — 잃으면 되돌릴 방법이 없다.

**M2.0 — 사본 (재생성보다 반드시 먼저).**
```bash
mkdir -p .moai/reports/t475/pre-regen
cp .moai/project/codemaps/*.md .moai/reports/t475/pre-regen/
ls .moai/reports/t475/pre-regen/ | wc -l        # 6 이어야 한다
```
**이 명령의 출력이 6이 아니면 M2.1을 실행하지 않는다.** 사본 확인은 재생성의 선행 조건이다.

**M2.1 — 생성기 5문서 재생성.**
```bash
# /moai codemaps --force  (스킬 실행)
ls .moai/project/codemaps/                       # 7개 항목(문서 6 + provenance.json)
```
스킬이 산출하는 것은 `overview` / `modules` / `dependencies` / `entry-points` / `data-flow` **5개**다. `docs-truth.md`는 산출물이 아니다(M2.2).

**M2.2 — `docs-truth.md` 손 갱신.** 스킬이 손대지 않으므로 실행자가 직접 갱신한다(REQ-CM2-014). 검증 경계는 REFRESH-001 REQ-CMR-004 계승 — **§1 에이전트 카탈로그 표 전수**를 `.claude/agents/` 트리 나열과 대조한다(표본 추출 없음). 갱신 여부와 대조 결과를 증거 파일의 **별도 섹션**에 남긴다 — M2.1의 증거와 섞지 않는다.

**M2.3 — 편입 확인.** M1이 omission으로 판정한 단위 전부에 대해:
```bash
/usr/bin/grep -rl -F "<단위>" .moai/project/codemaps/ | wc -l    # ≥1 이어야 한다
```
여전히 0이면 해당 문서를 직접 보정하고 보정 사실을 증거 파일에 기록한다(REQ-CM2-004).

**M2.4 — 구간별 전후 대조.** §A.3(b) 컷오프 명령이 산출한 구간(저작 시점 7개)마다:
```bash
diff -u .moai/reports/t475/pre-regen/<doc>.md .moai/project/codemaps/<doc>.md
```
출력을 그대로 증거 파일에 붙인다. **"변경 없음" 행은 빈 diff 출력을 증거로 첨부해야 한다** — 산문만 적힌 행은 반증 불가능하다(AC-CM2-005).

### M3 — 정확성 검증 3항목 증거 수출

REFRESH-001 REQ-CMR-002~004 계승. 세 항목 전부 유니크 기준 전수이며 표본 추출을 하지 않는다. **세 항목 모두 추출 명령이 명명돼 있다** — 명명 없는 추출은 느슨해도 판별할 근거가 없다(iter-1 D6).

- **(a) 인용 경로 실존** — 추출 규약은 `internal/graph/check_citations.go`의 정본 3요소다: 정규식 `:23` `\b(?:internal|pkg|cmd)/[A-Za-z0-9_/.-]*`, 후행 구두점 절삭 `:35` `".,;:)]}\"'"`, **blockquote(`>`로 시작하는 줄) 면제**. 코드펜스·mermaid는 면제가 아니다. 면제를 빠뜨리면 일부러 부존재를 인용한 줄이 전부 거짓 `absent`가 된다.
  판정면은 게이트가 이미 갖고 있다:
  ```bash
  ./bin/moai graph check --json | grep -A4 '"layer": "citations"'   # value 0 / verdict fresh 기대
  ```
  표는 그 판정의 **증거**이지 대체물이 아니다 — 계층은 수를 주고 표는 어느 경로인지를 준다.

- **(b) 패키지 구조 대조** — §A.3(a) 히트-0 패키지 명령(`/tmp/Zero.txt`)을 재생성 후 다시 돌려 잔여 히트-0 패키지를 전수 열거하고, 각각에 M1 판정(fold / omission)을 붙인다. **omission인데 여전히 0이면 기록 대상이 아니라 REQ-CM2-004 미이행이다**(AC-CM2-004 FAIL).

- **(c) 인용 식별자 실존** — "식별자"는 REQ-CM2-008이 명명한 명령의 출력으로 정의된다(백틱 인라인 코드 중 Go 식별자 형태). 재생성 전 `52f863f36`에서 이 명령은 **10행**을 낸다 — 0행이면 추출식이 문서 형식과 어긋난 것이므로 통과가 아니라 blocker다.

세 표 모두 `.moai/reports/t475/codemaps-accuracy-verification.md`에 수출한다. **게이트 판정과 무관하게 이 증거로 닫는다**(REQ-CM2-011).

### M4 — 재스탬프 (도달성 보장)

```bash
REV="$(git merge-base HEAD origin/develop)"
./bin/moai graph stamp codemaps --commit "$REV"
```
bare HEAD 스탬프는 금지다(REQ-CM2-009).

**도달성 확인은 `provenance.json`에서 읽는다 — 중간 파일이 아니라.** 0.1.0은 리비전을 `/tmp/t475-stamp-rev`에 저장하고 §E가 그것을 되읽었다. 그것은 §D의 "`/tmp`는 판정 근거가 아니다"와 부딪히고, 무엇보다 **실제로 스탬프에 들어간 값이 아니라 들어갔다고 믿는 값**을 검사한다. 스탬프가 기록한 값을 직접 읽는 쪽이 강하다:

```bash
./bin/moai graph check --json | grep 'content_anchor'
git merge-base --is-ancestor "$(/usr/bin/grep -o '"commit_sha": "[a-f0-9]*"' .moai/project/codemaps/provenance.json | cut -d'"' -f4)" origin/develop
echo ANCESTOR=$?    # 0 기대
```

**주의 — 지금은 `merge-base`와 `HEAD`가 같다**(둘 다 `52f863f36`). 그래서 이 시점에는 bare HEAD 스탬프와 merge-base 스탬프가 구분되지 않고 AC-CM2-009가 물지 않는다. run이 자체 커밋을 쌓는 순간 갈라지므로 **"지금 같으니 아무거나"로 읽지 않는다** — 명시 형식이 규정이다.

### M5 — 게이트 종결 + 관측 리포트

```bash
./bin/moai graph check ; echo EXIT=$?
```
codemaps `verdict=fresh`(value < 40, 기대 0), 다른 계층에 `stale` 없음. mx-index/edges의 `absent`는 예상 상태다.

`.moai/reports/t475/verdict.md`에 관측 3항목을 수출한다(REQ-CM2-013). **어떤 설정도 바꾸지 않는다** — 이 리포트가 임계값 질문에 대한 본 카드의 유일한 산출물이다.

① **임계 40 대비 값의 거동 — 재료는 누적 속도의 귀속이다.** 현재 앵커 `25a3212a9`를 찍은 것은 REFRESH-001(2026-09-02)이 아니라 워크트리 **t476이 2026-09-03T18:18:34Z**에 찍은 스탬프다(`provenance.json` 직독). 즉 64는 **5일** 누적분이며, CADENCE-001의 "corrected-40이 약 1.6일에 교차"와 정합한다. 임계값 판단에 필요한 것은 값이 아니라 이 속도이므로, 값과 함께 **앵커를 누가 언제 찍었는가**를 적는다.

② **후보(A·B·C층) 20개 전수의 fold/omission 분류 요약** — 몇 개가 fold였고 몇 개가 omission이었는지, 그리고 `internal/harness/*` 하위 12개가 어느 쪽으로 갈렸는지.

③ **후보에서 제외된 잔여 42개 패키지의 전수 목록** — 앵커 이후 무변경이라 이 카드가 판정하지 않은 히트-0 패키지 전부(그중 12개가 `internal/harness/*` 하위). 판정이 아니라 목록이며, 리드가 후속 카드를 정할 재료다. **선택 항목이 아니다** — 이것이 A층 필터를 정당화하는 이관 조건이고, 없으면 AC-CM2-012가 FAIL한다.

`tree_root`는 이 리포트에 싣지 않는다(§B 6번). 설계 의도가 코드 주석으로 확정된 사항을 관측으로 올리면 리드에게 판단할 것이 없는 항목을 넘기는 셈이고, 닫힌 조사를 다시 여는 신호가 된다.

## §G. Anti-Patterns

- **스탬프만 갱신하고 끝내기.** 게이트는 녹색이 되지만 낡은 서술에 새 앵커를 붙인 것뿐이다 — 카드가 명시적으로 금지한 실패 형태다.
- **후보를 손으로 열거.** iter-1 D1이 정확히 이것이었다 — 손 열거는 재현 불가능하고, `internal/template/agentemit`처럼 카드가 고치려는 종류를 빠뜨린다. 규칙을 돌리고 나오는 것을 받는다.
- **히트 0을 곧바로 누락으로 단정, 또는 부모 히트 수로 fold 단정.** 둘 다 양적 지표다. M1이 존재하는 이유다.
- **부모 산문을 읽지 않고 판정.** 패키지 이름에서 책임을 유도한 행은 판정이 아니라 추측이다(AC-CM2-002 FAIL).
- **재생성보다 먼저 사본을 뜨지 않음.** M2.4의 `diff -u`를 영원히 낼 수 없게 된다 — 사후 복구 불가(iter-1 D2).
- **`docs-truth.md`를 재생성 산출물로 간주.** `ls`가 7항목을 보여주며 통과하는 동안 그 문서만 낡는다(iter-1 D3).
- **후보를 "관측이니 기록만"으로 흘려보냄.** fold 판정 후보만 기록 전용이다. omission 판정 단위가 편입되지 않은 채 남으면 AC-CM2-004 FAIL이다.
- **임계값을 "관측된 값이 크니까" 올리기.** CADENCE-001이 통합 축 재유도로 이미 유지 판정했다. 관측은 리드에게 보고하고 판단은 운영자 몫이다.
- **게이트 녹색을 정확성 증거로 대체.** REQ-CM2-011이 금지한다 — 두 층은 독립적으로 성립해야 한다.
- **재생성 전 사본을 안 뜨고 시작.** 구간별 서술 차이(REQ-CM2-005)를 낼 수 없게 되고, M2 종료 후에는 복구할 수 없다.

## §H. Cross-References

- `.moai/specs/SPEC-CODEMAPS-REFRESH-001/` — 절차 정본(plan.md §F가 M1~M4로 서술).
- `.moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/` — 임계 40 유지 판정.
- `.claude/skills/moai/workflows/codemaps.md` — `/moai codemaps --force` 실행면.
- `internal/graph/check.go` — 게이트 구현(읽기 전용 참조; 수정 금지).
- 증거 경로: `.moai/reports/t475/`.
