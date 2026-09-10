# t357 plan 실측 — Tier L 산출물 상태 전이 규약 공백

카드: t357 (`[t338 별건] design.md·research.md 가 draft 로 남는다`)
측정 트리: `.claude/worktrees/t357` · HEAD `c6aa61346` (= origin/develop, 발산 `0 0`)
측정 시각: 2026-08-28 · 코퍼스: `.moai/specs/SPEC-*` 696개

---

## 1. Claim — 카드 전제 두 가지가 실측과 어긋난다

카드는 두 가지를 전제한다. 둘 다 이번 측정에서 뒤집혔다.

| 카드 전제 | 실측 |
|---|---|
| "상태 전이 규약이 4종(spec/plan/acceptance/progress)만 명시" | 규약은 **spec.md 1종만** 명시한다. `plan.md`·`acceptance.md`도 규약 밖이다 |
| "Tier L **6종**" | 규약상 Tier L은 **5종**(spec/plan/acceptance/design/research). `progress.md`는 별도 진행 기록 |

따라서 이 카드는 "4종 → 6종으로 넓힐 것인가"가 아니라 **"1종 → N종으로 넓힐 것인가"**를 정하는 카드다. 범위가 카드에 적힌 것보다 넓다.

## 2. Evidence — 규약과 검사가 각각 무엇을 보는가

### 2.1 규약 (문서)

`.claude/rules/moai/development/spec-frontmatter-schema.md:13`

> All SPEC documents (`spec.md`) MUST contain exactly these 12 fields in YAML frontmatter.

- 의무 대상이 **`spec.md`로 괄호까지 박혀 있다**. 나머지 산출물의 frontmatter는 스키마에 **존재하지 않는 필드**다.
- 같은 문서 §Status Transition Ownership Matrix는 산출물을 특정하지 않고 "SPEC의 status" 전이만 규정한다 — 어떤 파일의 status인지는 §Canonical 12 Required Fields가 답한다: spec.md.
- plan 워크플로도 같다. `.claude/skills/moai/workflows/plan/spec-assembly.md:72,80,313,519` — frontmatter 지시는 전부 spec.md 대상. design.md·research.md·plan.md·acceptance.md에 frontmatter를 쓰라는 지시는 **0건**.

### 2.2 검사 (코드)

| 지점 | 읽는 파일 | 근거 |
|---|---|---|
| `moai spec audit` | spec.md | `internal/spec/audit.go:347` `filepath.Join(specDir, "spec.md")` |
| lint 대상 발견 | spec.md | `internal/spec/lint.go:307,328` `discoverSPECs` — `SPEC-*/spec.md` 패턴 |
| 전이 소유권 lint | spec.md | `internal/spec/lint_ownership.go:11,200` — spec.md status 라인 변화만 추적 |
| `moai spec close` **쓰기** | spec.md + progress.md | `internal/spec/closer.go:312,335` `os.WriteFile` 2곳뿐. 444행 `rewriteSpecStatusCompleted`, 460행 `rewriteProgressStatusCompleted` |

`closer.go:627`이 acceptance.md를 읽긴 하나 **AC PASS 마커 판독용**이고 frontmatter status는 건드리지 않는다.

### 2.3 audit 무경고 실측

```
$ moai spec audit    # 트리 c6aa61346, 696 SPEC
rc=0
[INFO] 481 / [WARN] 0 / [ERROR] 0
'draft|design.md|research.md' 매치: 0
```
증거 파일: `.moai/cache/t357_audit.txt`

### 2.4 결론 — 리드 질문에 대한 답

> 규약이 4종만 명시해서인가, 검사가 4종만 보도록 짜여서인가, 둘 다인가?

**셋 다 아니다.** 규약은 1종(spec.md)만 명시하고, 검사도 1종만 본다 — **둘은 어긋나 있지 않고 정확히 일치한다.**
경고가 없는 것은 검사 누락이 아니라 **규약이 그 파일에 아무 의무도 지우지 않기 때문**이다. design.md가 `draft`로 남는 것은 규약 위반이 아니라, 규약에 없는 필드를 에이전트가 임의로 붙였다가 아무도 옮기지 않은 것이다.

## 3. Baseline-attribution — 코퍼스 계수

측정 스크립트 `.moai/cache/t357_measure.sh` (spec.md frontmatter의 `status:`·`tier:`, 그리고 design/research/plan/acceptance 각각의 frontmatter status를 추출), 원자료 `.moai/cache/t357_rows.tsv` 696행.

- `-` = 파일 없음 · `NOFM` = 파일은 있으나 frontmatter 자체 없음 · `NOSTATUS` = frontmatter는 있으나 status 필드 없음

### 3.1 spec.md 상태 분포 (696)

| status | 수 |
|---|---:|
| completed | 488 |
| implemented | 145 |
| archived | 31 |
| in-progress | 10 |
| draft | 10 |
| superseded | 9 |
| rejected | 1 |
| (spec.md 부재) | 2 |

**종결(closed) = completed + implemented = 633** — 이하 모든 계수의 모집단.

### 3.2 종결 SPEC의 비-spec 산출물 상태 (633)

| 상태 | design | research | plan | acceptance |
|---|---:|---:|---:|---:|
| 파일 없음 | 527 | 470 | 51 | 74 |
| **frontmatter 없음(NOFM)** | **81** | **131** | **402** | **379** |
| NOSTATUS | 2 | 3 | 25 | 25 |
| draft | 12 | 12 | 44 | 45 |
| completed | 9 | 9 | 61 | 57 |
| implemented | 0 | 0 | 28 | 26 |
| in-progress | 2 | 3 | 12 | 8 |
| 기타(backfilled/planned/audit-ready 등) | 0 | 8 | 10 | 19 |

**읽는 법**: design/research의 지배적 형태는 `draft`가 아니라 **frontmatter 부재**(81/131)다. `draft`로 남은 것은 12/12뿐. 카드가 지목한 형태는 소수 패턴이다.
그리고 `plan.md`·`acceptance.md`도 402/379가 frontmatter 부재 — 카드가 "규약 안"이라고 가정한 두 종이 실제로는 design/research와 같은 상태다.

### 3.3 Tier 분포 (696 / closed 633)

| tier | 전체 | closed |
|---|---:|---:|
| L | 89 | 81 |
| M | 234 | 221 |
| S | 112 | 105 |
| (필드 없음) | 247 | 226 |
| `2` / `3` (스키마 위반값) | 12 | — |

`tier: 2` 7건, `tier: 3` 5건은 열거값(S/M/L) 밖 값이다 — 이 카드 범위 밖이나 기록해 둔다.

## 4. 세 안의 영향 범위 — (a)(b)(c)

리드 요청 3항목. 모집단은 종결 SPEC 633.

| | **안 A** 전수 백필 | **안 B** 시점 이후만 | **안 C** 무상태 선언 |
|---|---|---|---|
| 내용 | 규약을 5종으로 넓히고 기존 종결 SPEC까지 소급 정정 | 규약을 5종으로 넓히되 소급 없음 (기존은 grandfather) | spec.md 1종 유지 명문화 + 나머지는 무상태로 확정 |
| **(a) 해당 완료 SPEC 수** | **544** (633 중 하나 이상의 산출물 상태가 spec 상태와 불일치) | **0** | **170** (규약 밖 status 필드를 실제로 보유) |
| **(b) 그중 Tier L (design/research 보유)** | Tier L closed **81** 중 design 보유 67 · research 보유 70 · **둘 다 67** | 0 | Tier L closed 중 design/research가 draft: **9** |
| **(c) 실제 편집될 파일 수** | **1,251** (design 97 + research 154 + plan 509 + acceptance 491) | **0** | **362** (status 라인 제거 대상) |

보조 계수:
- design/research를 보유하지만 `tier: L`이 아닌 종결 SPEC: **106** — 규약을 "Tier L 한정"으로 쓰면 이 106건이 규약 밖에 남는다.
- 안 A에서 파일 수 1,251 > SPEC 수 544인 것은 한 SPEC이 여러 산출물을 동시에 위반하기 때문.
- 안 C의 362는 "status 라인만 제거" 기준. frontmatter 블록 통째 제거로 정의하면 대상이 달라지므로 결정 시 정의를 못박아야 한다.

## 5. Gaps — 관측하지 않은 것

- **카드가 지목한 실례 `SPEC-AC-COUNT-DISCRIMINATOR-001`은 이 트리에 없다.** lane-7 t338 브랜치 소관이고 병합이 동결 중이라 develop에 아직 없다. 즉 **원 사례를 직접 재현하지 못했다** — 이 보고의 계수는 develop 코퍼스 696건 기준이며, 그 SPEC은 포함되지 않는다.
- 안 A를 택했을 때 **백필의 정당성**(닫힌 SPEC의 산출물 상태를 사후에 바꾸는 것이 이력 왜곡인지)은 측정 대상이 아니다 — 운영자 판단.
- `progress.md`의 status 라인은 이 측정에서 제외했다. closer.go가 유일하게 쓰는 두 번째 파일이지만 형태가 frontmatter가 아니라 본문 라인이라 축이 다르다.
- 검사를 넓힐 때의 **구현 비용**은 미측정(lint 규칙 신설 범위, era demotable 코드 편입 여부).

## 6. Residual-risk

- 상태 추출은 frontmatter 첫 `status:` 라인 기준이다. 본문에 `status:`가 먼저 나오는 변칙 파일이 있으면 오분류될 수 있다 — 다만 awk가 `---` 블록 안으로 범위를 좁히므로 가능성은 낮다.
- `NOFM` 판정은 파일 1행이 `---`가 아닌 경우 전부를 포함한다. 프론트매터가 2행부터 시작하는 파일이 있다면 NOFM으로 잘못 셀 수 있다.
- 안 C를 택해도 **에이전트가 다시 임의 frontmatter를 붙이는 것을 막는 장치는 없다**. 무상태 선언은 문서 선언이지 기계 가드가 아니므로, 재발 방지까지 원하면 lint 규칙(규약 밖 frontmatter 거부)이 별도로 필요하다 — 이것이 안 C의 숨은 작업량이다.
