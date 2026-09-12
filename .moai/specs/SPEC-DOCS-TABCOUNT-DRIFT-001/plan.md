---
id: SPEC-DOCS-TABCOUNT-DRIFT-001
title: 구현 계획 — 설정 탭 수·이름 드리프트 차단
version: "0.1.0"
status: draft
created: 2026-09-12
updated: 2026-09-12
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: "README.{md,ko,ja,zh}, docs-site/content/**, internal/web"
lifecycle: spec-anchored
tags: "docs, guard, i18n, drift, t530"
---

# 구현 계획

아래는 **바뀔 가능성이 큰 결정부터** 놓았다. §A(가드 설계)와 §B(자리별 처분)가 사람이 검토해야 할 지점이고,
§D 이하는 기계적인 편집이다.

---

## §A 가드 설계 — 가장 되돌리기 어려운 결정

### A.1 어디에 두는가

`internal/web/docs_tab_contract_test.go` (신규). 근거:

- `consoleTabs()` 는 `package web` 의 비공개 함수다. 같은 패키지 안이어야 export 를 새로 만들지 않고 읽는다.
- 이 패키지에는 이미 소스·문서 파일을 패키지 상대 경로로 읽는 전례가 있다 —
  `internal/web/restyle_test.go:18` 의 `osReadFile`, 그리고 `docs-site/static/moai-brand.css` 토큰을 대조하는 `TestConsoleCSSEmbedded`.
  즉 "웹 패키지 테스트가 저장소 문서를 읽어 대조한다" 는 패턴은 새로 만드는 것이 아니라 **이미 있는 것을 재사용**한다.
- 테스트의 cwd 는 패키지 디렉터리이므로 저장소 루트는 `../../` 다.

### A.2 무엇을 읽는가

| 읽는 대상 | 용도 |
|---|---|
| `consoleTabs()` | 탭 개수와 순서·id (정본) |
| `internal/web/assets/i18n.js` | 로케일별 렌더 라벨 (이름 정본) |
| 대상 12파일 (명시 목록) | 문서에 적힌 수·이름 |

대상 파일 목록은 테스트 안의 리터럴 슬라이스로 둔다. 목록이 코드에 있어야 "무엇을 지키고 있는지" 를 읽을 수 있다.

### A.3 무엇을 단정하는가 — 두 갈래

- **N1 (음성 단정)**: 대상 파일에 **수 + 탭 명사가 인접한 구절이 하나도 없다**.
  정규식은 숫자와 낱말 수사를 모두 본다 — `[0-9]+`, `nine|ten|…|fifteen`, `九|十|十四`, `아홉|열|열네`, 그리고
  탭 명사 `tabs?|탭|タブ|标签页`. 사이에 허용하는 글자는 소수(공백·`개`·`-`·`の`·`个`·`가지` 정도)로 좁힌다.
  - 다만 `must-stay` 로 판정된 자리가 하나라도 남으면 N1 은 "0건" 이 아니라 "허용 목록과 정확히 일치" 로 바뀐다.
    허용 목록의 각 항목은 그 자리의 수가 `len(consoleTabs())` 와 같은지도 함께 본다.
- **N2 (양성 단정)**: 이름 목록을 담은 자리(README 4본 411~414 구역, `advanced/moai-web-console.md` 의 번호 목록)에서
  추출한 이름 배열이 `consoleTabs()` 순서의 렌더 라벨 배열과 **순서까지 같다**.

N1 만으로는 부족하다 — 수를 다 지워도 이름이 어긋날 수 있고, 실제로 지금 어긋나 있다(D군).
N2 만으로도 부족하다 — 이름 목록이 없는 자리(README:750 표 행)가 있다.

### A.4 오탐 방지

- 대상 파일 명시 목록 → `plugins.md:100` 의 `4 タブ` 는 애초에 스캔되지 않는다.
- 수 + **탭 명사 인접** 조건 → `README.md:422` 의 `nine forms`, `:418` 의 `Eleven ref skills`, `:486` 의 `twelve` 는 걸리지 않는다.
- 두 장치를 **함께** 둔다. 하나만으로는 각각 다른 쪽 오탐을 놓친다.

### A.5 공허한 초록 방지

가드가 실제로 무엇을 잡는지 증명하지 않으면 통과는 근거가 못 된다. run-phase 는 대상 문서 한 자리를
일부러 틀리게 바꿔(예: README:414 의 이름 하나를 `Audits` 로) 가드가 FAIL 하는 것을 보이고, 되돌린 뒤 PASS 를 보인다.
두 출력 모두 `progress.md` §E.2 에 기록한다. (AC-TCD-006)

---

## §B 자리별 처분 — 사람이 검토할 문구 결정

REQ-TCD-003 이 요구하는 per-site 판정. 전체 20자리의 파일·행은 `.moai/reports/t530/tab-count-sites.md`.

| 군 | 자리 | 처분 | 근거 / 대체 문구 방향 |
|---|---|---|---|
| A1~A4 | `cli-reference/web.md:53` × 4로케일 | **removable** | `/settings` 엔드포인트 행에서 수는 정보가 없다. "설정 탭 화면" 수준으로 줄이고 `?tab=` 설명만 남긴다 |
| B1,B3,B5,B7 | `advanced/moai-web-console.md:25` × 4 | **removable** | 화면 목록 표의 설명 칸. "프로필 선호도와 프로젝트 섹션 편집" 만으로 충분하다 |
| B2,B4,B6,B8 | 같은 파일 `:128` × 4 | **removable** | 바로 다음 줄부터 이름 14개가 번호 목록으로 나온다. "아래 탭들이 세로 목록으로 펼쳐집니다" 로 가리킨다 |
| C1~C4 | `README*:414` × 4 | **removable (수만)** | 이미 이름 14개를 전부 적고 있다. 수를 빼고 목록만 남긴다 — README 가 일반화할 본보기 |
| C5~C8 | `README*:750` × 4 | **removable** | 명령 표 행. `14-tab settings` → `설정 탭` 수준. 표에서 수는 읽는 이에게 쓸모가 없다 |

**판정 결과: 20자리 전부 removable.** 즉 `must-stay` 자리는 없고, 가드의 N1 은 허용 목록 없이 "0건" 으로 단정한다.

그렇다면 가드가 필요 없는 것 아닌가 — 아니다. 수를 지우면 **이름**이 그 자리를 대신 진다(C1~C4, B2/B4/B6/B8 아래 목록).
이름은 지울 수 없는 정보이므로, 드리프트가 이름 쪽으로 옮겨갈 뿐이다. N2 가 그 자리를 지킨다.
D군이 이미 그 드리프트의 실물이다.

---

## §C 결정 기록 — llm 탭 이름 (2026-09-12, 해소됨)

**결정: 갈래 1 — 문서를 코드에 맞춘다. 본 SPEC 범위 안에서 끝낸다.**

4번 탭의 문서 이름은 D군 12자리 전부 `3rd Party LLM` 계열인데, 콘솔이 렌더하는 라벨은
`GLM Settings` / `GLM 설정` / `GLM設定` / `GLM设置` 다(`internal/web/assets/i18n.js:229,1096,1852,2608`).
문서 쪽을 렌더 라벨로 고친다.

리드가 밝힌 근거:

- 이 카드의 축은 **문서 정확성**이고, 사용자가 실제로 보는 것은 렌더 라벨이다.
- 코드 쪽 두 표면이 서로 어긋나 있지 않다 — `i18n.js:229` 와 `schemaform.go` 의 `Baseline: "GLM Settings"` 가
  일치하므로 "코드 안에서 무엇이 현재 의도인가" 를 먼저 가려야 하는 모호함이 없다. 현재 의도는 `GLM Settings` 다.

`[NEEDS CLARIFICATION: llm-tab-label]` 마커는 이 결정으로 **해소되었고 남겨두지 않는다.**
D군은 이제 미룬 관측이 아니라 M4 의 범위 안 작업이며, N2 가드는 렌더 라벨을 기준으로 단정한다 —
한시 허용 목록도, 후속 카드 주석도 두지 않는다 — 둘 다 "코드 라벨을 고친다" 는 반대 갈래에 딸린 조건부 장치였고,
그 갈래는 채택되지 않았다. 그 갈래가 나중에 되살아날 경우의 처리는 `spec.md §7` 잔여 위험에 남겼다.

---

## §D 마일스톤

우선순위 순. 시간 예측은 두지 않는다.

### M1 (High) — 가드 먼저 (실패하는 상태로)

- `internal/web/docs_tab_contract_test.go` 작성. 지금 트리에서 **FAIL 해야 한다** (A군 4자리 + D군 8자리).
- FAIL 출력을 `progress.md` §E.2 에 기록 — 이 카드가 실재하는 결함을 겨눈다는 증거.

### M2 (High) — A군: 틀린 수 제거

- `cli-reference/web.md:53` 4로케일. 한 커밋.

### M3 (High) — B군 + C군: 남은 수 제거

- `advanced/moai-web-console.md` 8자리, README 8자리. 로케일 패리티를 커밋 단위로 유지한다.

### M4 (High) — D군: 이름 정합 (§C 결정 반영, 범위 안)

- 12자리를 렌더 라벨로 고친다 — 실측 `grep '3rd Party LLM|서드파티 LLM|サードパーティ LLM|第三方 LLM'` 12행:
  목록 8자리(README ×4 :414, `advanced/moai-web-console.md` ×4 :133) + 산문 4자리(같은 파일 ×4 :165).
- 대상 라벨: en `GLM Settings` / ko `GLM 설정` / ja `GLM設定` / zh `GLM设置`.
- 4로케일을 같은 커밋에 담는다(REQ-TCD-007).
- 한시 허용 목록 없음, 후속 카드 주석 없음 — N2 가드는 처음부터 렌더 라벨을 기준으로 들어간다.

### M5 (Medium) — 가드 GREEN + 변이 시험

- 가드 통과 확인. 일부러 깨뜨려 FAIL 재현 → 되돌려 PASS (§A.5).

### M6 (Low) — 빌드·패리티 검증

- `hugo --gc --minify` 경고 없음, 로케일 파일 존재·섹션 수 패리티, README 4본 heading 패리티.
- 후속 카드 2건(§5)을 리드에게 카드 요청으로 올린다 — 이 카드에서 만들지 않는다.

---

## §E 위험

| 위험 | 완화 |
|---|---|
| 가드가 오탐을 내 CI 를 막는다 | A.4 두 장치 + 오탐 3자리를 테스트 안의 음성 케이스로 고정 |
| 수를 지우다 문장이 어색해진다(특히 ja/zh) | 문장 재작성이 아니라 최소 구절만 손댄다. 4로케일을 같은 커밋에서 나란히 본다 |
| `consoleTabs()` 가 다른 카드에서 바뀌어 가드가 깨진다 | 그것이 가드의 목적이다. 깨지면 문서를 고치라는 신호이고, 테스트 실패 메시지가 고칠 자리를 가리키게 쓴다 |
| 이름 추출 파서가 문서 서식에 취약하다 | 번호 목록의 `**이름**` 굵은 글씨 패턴에만 의존하고, 추출 개수가 14가 아니면 그 자체로 FAIL 시켜 조용한 0건 통과를 막는다 |
| 워크트리에서 `../../README.md` 가 안 읽힌다 | M1 에서 첫 실행으로 확인. 못 읽으면 `t.Fatal` — skip 하지 않는다 |

---

## §F Cross-references

- 전수 열거: `.moai/reports/t530/tab-count-sites.md`
- 기존 전례(웹 패키지 테스트가 docs-site 를 읽는다): `internal/web/restyle_test.go` `TestConsoleCSSEmbedded`
- 정본 순서 테스트: `internal/web/tab_layout_test.go` `TestConsoleTabsOrder`
- 경계: 카드 t509 (codex 패널 반경)
