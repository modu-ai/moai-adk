---
id: SPEC-DOCS-LOCALE-PARITY-REPAIR-001
title: "docs-site locale-parity repair — e2e desktop-native en/zh stale + doctor ja/zh examples + legacy emphasis spacing + skill-guide SVG rules"
version: "0.2.0"
status: draft
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: "docs-site/content"
lifecycle: spec-anchored
tags: "docs, i18n, locale-parity, docs-site, t538"
tier: M
related_specs: [SPEC-DESKTOP-NATIVE-E2E-001, SPEC-DOCS-V313-CATCHUP-001, SPEC-DOCS-CODEX-WIRING-CALLOUT-001]
---

# SPEC: docs-site 로케일 패리티 수리 — e2e·doctor·skill-guide 4-로케일 정합 복원

## HISTORY

- 2026-09-08: 카드 t538 착수. 리드 배차 전제("docs-site 4-locale 에 v3.1.3 변경사항 없음")는 측정-선행 스윕에서 절반 정정됨 — 본 SPEC §B.1 참조.
- 2026-09-08: 측정 근거 `.moai/reports/t538/plan-phase.md` (이 워크트리 `bce6d7e08` = origin/develop 팁에서 실측). 본 SPEC 과 측정 파일이 어긋나면 **측정 파일이 이긴다**.
- 2026-09-08: plan 페이즈 아티팩트 4종 생성 (status: draft).
- 2026-09-08: plan-audit iter1 **FAIL 0.71** (Tier M 문턱 0.80) — `.moai/reports/t538/plan-audit-iter1.md` (커밋 `86c8023b6`) 결함 D1-D6 수리 → version 0.2.0.
  - D1: AC-006 기대값 정정(전체 파일 2→4 각) + 예시 블록 한정 행두-앵커 판정 추가(`^moai doctor (permission|sandbox)` — 표행 배제, ja/zh baseline 0 실측).
  - D2: REQ-005(호스트 OS 규칙 문단)에 실효 검증 부여 — AC-005에 원어-토큰 grep 병합(ko `호스트 OS 규칙`·ja `ホスト OS ルール` baseline 1 실측, en `host OS rule`, zh 파생 시점 확정 절차).
  - D3: plan.md AC 색인 전면 재정렬(acceptance.md 실번호).
  - D4: REQ-012 zh 토큰 `桌面原生`→`原生桌面` 계열(zh:170 실측 기존 용어).
  - D5: RED-now AC 계수 4건→6건 정정.
  - D6: CHANGELOG 배치 `[Unreleased] → ### Added` 최상단으로 정정(`### Docs` 섹션은 현행 [Unreleased] 부재 — t535 선례 = 직하 배치).
  - F1 수리: AC-010 `--quiet` 제거, WARN/ERROR 행 계수로 판정.
  - F2 수리: §C G1 문장을 "명시한 제거 대상"이 아닌 "지연 동안 생존" 서술로 정밀화.
  - F3 각하(기록): REQ-004/008/013의 shall+shall-not 복합은 GEARS 통합 복합절 형식 내 복합으로 유지 — MP-2 통과 판정과 일치.
  - F4 각하(기록): 측정 파일(plan-phase.md)의 codex-dual-harness "각 64행"은 63행이 실측 — 존재 주장 자체는 참. 측정 SSOT는 소관 밖이라 본 SPEC에서 정정하지 않으며, 후속 터치 시 정정 대상으로 기록.
  - F5 각하(기록, 대응 절차 추가): ko skill-guide :167-177의 기존 내부-경로 오염은 본 카드 불요 — plan.md §G에 G4 파생자가 해당 스타일을 모방하지 말라는 경고 추가. docs 위생 카드로 분리 권고.
  - F6 각하(기록): `hns-oss-docs-run` 문법-오류 전제는 배차 지시문 기재 사항이며 감사자 실측(find 히트 0)으로 본체 미확인 — 금지 자체는 방향상 안전하게 유지하고, run 페이즈 진입 시 리드가 러너 실물 확인으로 전제를 확정하는 것으로 기록.

## §B 전제 정정 및 귀속 (audit 오귀속 방지 — 본 절은 REQ 보다 먼저 읽힌다)

### §B.1 배차 전제 정정

리드 배차문의 전제 "docs-site 4-locale 에 v3.1.3 변경사항이 없다(codex 듀얼 하네스 포함)"는 절반은 낡았다:

- **codex 듀얼 하네스는 이미 문서화돼 있다**: `docs-site/content/{ko,en,ja,zh}/advanced/codex-dual-harness.md` ×4 실재. 작성 t274 `SPEC-DOCS-V313-CATCHUP-001` (커밋 `175d63f3f`, 2026-08-26), 갱신 t496 `SPEC-CODEX-EVENT-COVERAGE-001` (커밋 `732609dcf`, 2026-09-07). v3.1.3 캐치업 축은 **이미 반영** — t535 verdict §C5 예고대로. 이 카드에서 codex 관련 신규 작업 없음.
- 실제 갭은 아래 §C 의 G1-G3 (+ 선택 G4) 뿐이다.
- 「카드 전제 정정 2026-09-08: 측정 결과 v3.1.3 반영은 완료돼 있었고(t274 #1662 + t496), 실제 갭은 G1–G3」
- G2 부속 기록: 「t535 관측 정정 2026-09-08: ja 도 해당」 — t535 sync-audit F2 의 zh-단일 기록은 **침묵으로 지워지지 않으며**, 본 정정 기록이 뒤에 덧붙는 방식으로만 정정된다.

### §B.2 G1 부채의 시대 귀속 [HARD]

G1 (e2e desktop-native en/zh 스테일)의 부채는 **v3.1.0대** 것으로, `SPEC-DESKTOP-NATIVE-E2E-001` 이 docs-site 동기화를 연기(defer)하면서 생겼다. 배차 카드가 느슞하게 "v3.1.3"이라 적었으나 desktop-native 는 v3.1.3 과 별개 축이다. **어떤 미래 감사도 이 부채를 v3.1.3 로 귀속해선 안 된다.** 본 SPEC 의 G1 수리는 착지 현실에 대한 문서 정렬이지 신규 기능이 아니며, 따라서 어떤 로케일에도 new-badge 를 붙이지 않는다 (ko·ja 기존 행이 배지 없음이 그 근거).

## §C 측정된 갭 (범위 — 재스코핑 금지)

모든 실측은 `.moai/reports/t538/plan-phase.md` 근거 (2026-09-08, `bce6d7e08`, `/usr/bin/grep`).

### G1 — e2e desktop-native en/zh 스테일 (실질 최대)

`docs-site/content/{ko,en,ja,zh}/utility-commands/moai-e2e.md`:

| 축 | ko (201행) | ja (201행) | en (195행) | zh (195행) |
|---|---|---|---|---|
| 플래그 행 `--platform …desktop-native` | ✅ :58 | ✅ :58 | ❌ `web\|mobile\|desktop` | ❌ 동일 |
| 툴체인 매트릭스 desktop-native 3-OS 행 | ✅ :78-80 | ✅ :78-80 | ❌ 0행 | ❌ 0행 |
| 자동감지 표 desktop-native 마커 행 | ✅ :98 | ✅ :98 | ❌ | ❌ |
| "대상 없음" 절 | ✅ :176 라우팅 | ✅ :176 | ❌ :170 deferral notice 잔존 | ❌ :170 동일(中文) |
| 호스트 OS 규칙 문단 | ✅ 매트릭스 직후 | ✅ | ❌ | ❌ |

en:170 은 기능이 착지한 뒤에도 "OS-level native-desktop automation **is not yet provided**"라고 기능 부재를 주장하는 부정확 문서다. zh:170 도 "原生桌面自动化尚未提供"로 동일. 이는 `SPEC-DESKTOP-NATIVE-E2E-001` 의 docs-site 동기화가 지연되는 동안 생존한 문장이다(plan-audit iter1 F2 — 동일 SPEC 이 제거를 지시한 verbatim 문장은 다른 표면의 것이므로 "명시한 제거 대상"이 아닌 "지연 동안 생존"으로 서술). **ko·ja 는 이미 완비 — 무변경.** ja 는 desktop-native 표 형태의 참조 렌더링으로 사용한다.

### G2 — doctor.md 예시 블록 ja·zh 2줄 부족

`docs-site/content/{ja,zh}/cli-reference/doctor.md` 예시 블록(ja:106-111, zh:106-111)은 4행에서 `hook` 으로 끝나며 `moai doctor permission`·`moai doctor sandbox` 행이 누락됐다. ko(:115-122, 예시 행 :120-121)·en(:118-119)은 보유. 명령 표(:32-33)는 4-locale 모두 정상 — 갭은 예시 블록 한정. t535 sync-audit F2 가 zh 만 지목했으나 **측정 정정: ja 도 누락** → 본 SPEC 의 G2 범위는 두 로케일이다.

### G3 — doctor.md 기존 절 강조 간격 위반 3곳

`ko:41` (`**권고(advisory)**`), `ja:39` (`**勧告 (advisory)**`), `zh:39` (`**建议 (advisory)**`) — 괄호가 강조 마커 안쪽. t535 run 이 신규(Codex Wiring) 절은 0위반으로 착지했고 기존 절은 REQ-DWC-015 경계상 기록만 했다. 본 SPEC 은 **기존 절만** 수리고 t535 가 착지한 Codex Wiring 절은 만지지 않는다.

### G4 — skill-guide.md SVG v3.1.3 축 (포함 결정)

`advanced/skill-guide.md` 는 `moai-domain-svg-infographic`(ko:158, :163)을 서술하지만 v3.1.3 접근성-이름 보장(role="img", aria-labelledby, title 최전단 — SVG060-064)과 커넥터-기하 검사(SVG070-074)를 언급하지 않는다. 기존 svg-infographic 문맥에 사실문 1-2행 추가, ko 정본 → 4-로케일 파생.

## §D 요구사항 (GEARS)

### REQ-001 — G1 플래그 행 (en·zh)

**When** 독자가 en 또는 zh `moai-e2e.md` 의 플래그 표를 읽을 때, 그 페이지는 ko:58 정본과 동일하게 `--platform` 행에 `desktop-native` 선택지를 나열해야 한다(shall).

### REQ-002 — G1 툴체인 매트릭스 3-OS 행 (en·zh)

en·zh `moai-e2e.md` 의 Platform-Toolchain Matrix 는 ko:78-80 정본(canonical)과 동일한 desktop-native 3-OS 행(macOS/Windows/Linux, 도구·폴백·비고 포함)을 가져야 한다(shall). 표 형태(열 구조)의 참조 렌더링은 ja:78-80 이다.

### REQ-003 — G1 자동감지 표 마커 행 (en·zh)

**When** en·zh `moai-e2e.md` 의 자동감지(autodetection) 표가 렌더링될 때, ko:98 정본과 동일한 desktop-native 마커 행을 포함해야 한다(shall).

### REQ-004 — G1 deferral 문장 제거 및 라우팅 의미론 치환 (en·zh)

en:170 ("…is not yet provided…") 및 zh:170 ("…尚未提供…")의 stale deferral 문장은 제거되어야 하며(shall not 잔존), 그 자리에 ko:176 정본의 라우팅-투-네이티브-레인 의미론(비 Electron/비 Tauri 네이티브 데스크톱 앱이 desktop-native 레인으로 분류·라우팅됨)이 각 로케일 자연어로 서술돼야 한다(shall).

### REQ-005 — G1 호스트 OS 규칙 문단 (en·zh)

en·zh `mo-e2e.md` 가 아니라 `moai-e2e.md` 는 매트릭스 직후에 ko 정본이 가진 호스트 OS 규칙 문단(desktop-native 제어가 로컬 호스트 OS 에서만 동작한다는 규칙)을 가져야 한다(shall).

### REQ-006 — G1 ko·ja 무변경 + 배지 금지

본 수리는 `docs-site/content/ko/utility-commands/moai-e2e.md` 와 `docs-site/content/ja/utility-commands/moai-e2e.md` 를 변경해선 안 되며(shall not), 수리 대상 en·zh 페이지 및 그 밖의 모든 페이지에 new-badge 를 추가해선 안 된다(shall not).

### REQ-007 — G2 doctor 예시 행 보강 (ja·zh)

**When** ja·zh `cli-reference/doctor.md` 의 예시 블록이 렌더링될 때, `moai doctor permission` 행과 `moai doctor sandbox` 행이 각 로케일 주석과 함께 ko(:120-121)·en(:118-119) 정본 구성과 동등하게 포함돼야 한다(shall).

### REQ-008 — G3 강조 간격 수리 (ko·ja·zh, 기존 절 한정)

ko:41·ja:39·zh:39 의 강조+괄호 패턴은 괄호가 강조 마커 **바깥**에 오도록 수리돼야 한다(shall) — `**권고** (advisory)` 형태(i18n 숙칙 §5: 괄호문은 마커 밖). 수리는 기존 Home Disk Usage 절 3곳에 한정되며, t535 가 착지한 Codex Wiring 절(ko:69·ja:67·zh:67)은 bold+괄호 패턴이 없어 rule §5 에 이미 준수 — 변경해선 안 된다(shall not). 판정은 3행 한정 검사가 아니라 파일 전체 bold-내부-괄호 패턴 스캔이 0건을 반환하는 것으로 한다(shall). 공유 어구("권고/advisory 성격" 계열)는 괄호 이동 후에도 생존하므로 파일-내부 일관성은 절 편집 없이 유지된다.

### REQ-009 — G4 skill-guide SVG 규칙 보강 (×4)

ko `advanced/skill-guide.md` 의 기존 svg-infographic 문맥에 v3.1.3 의 (a) 접근성-이름 보장(role="img", aria-labelledby, title 요소 최전단 — SVG060-064)과 (b) 커넥터-기하 검사(SVG070-074)를 서술하는 사실문 1-2행이 추가돼야 하며(shall), en·ja·zh 는 ko 정본에서 파생돼야 한다(shall).

### REQ-010 — CHANGELOG 항목

**When** 수리 집합이 착지할 때, `CHANGELOG.md` 의 `[Unreleased]` → `### Added` 최상단에 수리 집합을 커버하는 항목 1건이 삽입돼야 하며(shall — t535 선례의 `### Added` 직하 배치; `### Docs` 섹션은 현행 [Unreleased] 에 부재), 그 항목은 카드 id `t538` 과 G2 의 두-로케일(ja·zh) 정정 범위를 명시해야 한다(shall).

### REQ-011 — 4-로케일 동일 착지 [HARD]

**Where** 본 SPEC 의 변경 집합(e2e en·zh + doctor ja·zh + doctor ko 1행 간격 + skill-guide ×4 + CHANGELOG)에 대해, 전체 집합은 이 브랜치(`WT-docs-v313-locales`)의 **하나의 커밋 체인**으로 착지해야 하며(shall), 부분-로케일 착지는 금지된다(shall not).

### REQ-012 — 검증 도구 규율

본 SPEC 의 모든 검증 grep 은 `/usr/bin/grep` 을 사용해야 한다(shall — 셸 `grep` 은 ugrep 래퍼로 `-I`/`--ignore-files` 기본값 때문에 파일을 조용히 건너뛴다). ko·ja·zh 존재 단정에는 원어 토큰(ko `데스크탑-네이티브` / ja `デスクトップネイティブ` / zh `原生桌面` 계열)을 사용해야 한다(shall — ASCII 토큰 `desktop-native` 는 ko·ja 산문에 플래그 행 외엔 나오지 않고, zh 페이지의 기존 용어 계열은 `原生桌面`이다 — `桌面原生` 이 아님, zh:170 실측).

### REQ-013 — 실행 주체

run 페이즈 집행은 hns-oss-docs 스페셜리스트 스폰(content-author 가 ko 정본 편집 → locale-translator 가 en/ja/zh 파생)으로 수행돼야 하며(shall), 알려진 저장-스크립트 문법 오류가 있는 `hns-oss-docs-run` 러너는 사용해선 안 된다(shall not).

## §E 제약

- ko 정본 → en/zh 파생 (문서 체인: ko → en → ja/zh). ja 는 G1 표 형태 참조 렌더링으로 열람만.
- Mermaid TD-only, 본문 이모지 금지(icon shortcode 사용), 강조-마커 간격 규칙, `adk.mo.ai.kr` URL 화이트리스트 — Skill("hns-oss-docs-i18n-rules") 전 숙칙 적용.
- 버전 SSOT: `hugo.toml` — 본 SPEC 은 버전 표기 축을 건드리지 않는다(이미 t274 가 해소).
- 허깅 빌드 게이트: exit 0 + WARN/ERROR 0.

### Out of Scope — v3.1.3 캐치업 축

- codex 듀얼 하네스 문서화 등 v3.1.3 캐치업 — t274 `SPEC-DOCS-V313-CATCHUP-001` 이 이미 착지. 본 SPEC 은 codex 관련 신규 작업을 하지 않는다.

### Out of Scope — ko·ja e2e 페이지

- `ko/ja` `utility-commands/moai-e2e.md` 는 이미 정확하므로 일절 변경하지 않는다(REQ-006).

### Out of Scope — t535 가 착지한 Codex Wiring 절의 재작성

- `doctor.md` 의 Codex Wiring 절(ko:69·ja:67·zh:67)은 t535 가 착지시킨 콘텐츠로, bold+괄호 패턴이 없어 rule §5 에 이미 준수한다(코디네이터 확인 2026-09-08 — 본 SPEC plan 페이즈 전체-파일 스캔 실측과 일치) — 본 SPEC 은 그 절을 일절 편집하지 않는다. REQ-008 의 수리는 기존 Home Disk Usage 절 3곳뿐이다.

### Out of Scope — v3.1.3 Fixed 항목 전수 판정

- 동작 수리류의 문서 요구성 전수 판정은 하지 않는다(측정 파일 §Gaps 계승 — 동작 수리는 통상 문서 표면 불요).

### Out of Scope — 문서 파일 직접 편집 (plan 페이즈)

- 본 SPEC 의 plan 페이즈에서 문서 파일을 편집하지 않는다. 집행은 run 페이즈 위임(REQ-013)이다.

## §F 성공 기준 요약

- en·zh e2e 가 ko 와 구조적 패리티(플래그 행·3-OS 매트릭스·자동감지 행·라우팅 절·호스트 OS 문단)에 도달.
- ja·zh doctor 예시 블록이 6행(4+permission+sandbox)이 됨.
- doctor 강조 간격 위반 0 (기존 절 + 신규 절 합산).
- skill-guide ×4 가 SVG060-064·SVG070-074 를 언급.
- 허깅 빌드 경고 0, 4-로케일 패리티 체크 통과, CHANGELOG 항목 1건.

## §G 검증 전략 요약

상세 커맨드·기대출력은 `acceptance.md` §D AC 매트릭스. 공통: `/usr/bin/grep` 필수, RED-now AC 는 baseline-vs-after 로 기술, 빌드 게이트는 `hugo` (exit 0, WARN/ERROR 0).

## §H 교차 참조

- `.moai/reports/t538/plan-phase.md` — 측정 SSOT (본 SPEC 과 어긋나면 이 파일이 이김)
- `SPEC-DESKTOP-NATIVE-E2E-001` — G1 부채의 원천 SPEC (v3.1.0대, docs 동기화 연기)
- `SPEC-DOCS-V313-CATCHUP-001` — v3.1.3 캐치업 선례(t274) + CHANGELOG Docs/Added 배치 선례(t535)
- `SPEC-DOCS-CODEX-WIRING-CALLOUT-001` — t535, G3 경계(REQ-DWC-015)의 출처
- Skill("hns-oss-docs-i18n-rules") / Skill("hns-oss-docs-verify") — HARD 숙칙 + exit 게이트
