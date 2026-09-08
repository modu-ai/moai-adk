# t538 — sync-phase 독립 감사

SPEC: `SPEC-DOCS-LOCALE-PARITY-REPAIR-001` · 카드 t538 · Tier M
감사 트리: `.claude/worktrees/t538` · 브랜치 `WT-docs-v313-locales` · 감사 시점 tip `df653dd8e`
베이스 `bce6d7e08` · 기준선 blob `d787897ca` · 콘텐츠 커밋 `bbc37b729` · sync 종결 `7b951a240`
계측: 모든 grep 은 `/usr/bin/grep`. 셸 `grep` 은 조용히 건너뛰는 ugrep 래퍼이므로 감사 수치에 쓰지 않았다.
감사자: sync-auditor (독립) · 카드 산출물의 수치를 인용하지 않고 전건 자체 재유도했다.

## 판정

**PASS** · 가중 조화평균 **93.0**

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 96 | PASS (must-pass) | AC 11건 전건 자체 재유도로 재현. RED-now 6건 전건 기준선 blob 에서 RED 확인 |
| Security (25%) | 98 | PASS (must-pass) | 문서 전용. 코드·설정 0파일, 비밀값 0, 신규 외부 URL 0, 내부 식별자 유출 0 |
| Craft (20%) | 86 | PASS | 증거 규율은 강하나 증거 앵커가 tip 보다 2커밋 뒤, 잔여 위험 목록에 누락 2건 |
| Consistency (15%) | 88 | PASS | ko 정본 충실 파생. zh 행 레이블 주석 형태가 나머지 3로케일과 갈림 |

must-pass 방화벽(Functionality·Security) 양쪽 통과. 차단성(blocking) 결함 없음 — 아래 결함은 전부 재량(optional)이다.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

카드가 주장한 것: AC 11건 전건 PASS, RED-now 6건 편집 전 RED 관측, 쓰기 표면 정확히 10파일, 경계 3건(ko·ja e2e 무변경 / 배지 미추가 / t535 Codex Wiring 절 불변) 준수, 검증식 정정 2건은 계측기 수리이지 기준 완화가 아님.

감사 결론: **위 주장 전건이 독립 재측정에서 성립한다.**

### Evidence — 자체 재유도한 원출력

기준선은 카드의 수치를 인용하지 않고 `git show d787897ca:<경로>` 로 blob 에서 직접 재유도했다. 먼저 그 기준선 SHA 자체의 정당성을 확인했다.

```
$ git diff --name-only bce6d7e08..d787897ca -- docs-site CHANGELOG.md
(출력 없음)      # docs-site·CHANGELOG 는 베이스 이후 d787897ca 까지 무변경 — 기준선 SHA 로 유효
```

AC-001~004 (G1):

```
AC-001 en "not yet provided"   baseline 1 → after 0
AC-002 zh "尚未提供"            baseline 1 → after 0
AC-003/004 grep -ci desktop-native   baseline en 0 / zh 0 → after en 7 / zh 7
```

AC-005 (표 행·헤딩 패리티, 호스트-OS 토큰):

```
baseline rows  ko 46  en 42  ja 46  zh 42
after    rows  ko 46  en 46  ja 46  zh 46
after headings ko 18  en 18  ja 18  zh 18
after host-OS token  ko(호스트 OS 규칙) 1 · ja(ホスト OS ルール) 1 · en(host OS rule) 1 · zh(主机 OS 规则) 1
baseline host-OS token  en 0 · zh 0
```

AC-006~008:

```
AC-006 앵커 ^moai doctor (permission|sandbox)   baseline ja 0 / zh 0 → after ja 2 / zh 2
AC-006 전체 파일                                  baseline ja 2 / zh 2 → after ja 4 / zh 4
AC-007 \*\*[^*]*\([^)]*\)[^*]*\*\* on doctor.md  baseline ko 1 en 0 ja 1 zh 1 → after ko 0 en 0 ja 0 zh 0
AC-008 grep -o "SVG060\|SVG070" | wc -l          baseline 0/0/0/0 → after ko 2 en 2 ja 2 zh 2
AC-008 옛 읽기 grep -c (동일 트리)                 after ko 2 en 2 ja 2 zh 2
```

AC-009~011:

```
AC-009 git diff --quiet bce6d7e08 -- ko/ja moai-e2e.md   exit=0   (무변경)
AC-009 추가된 badge 행                                     0
AC-010 hugo --source docs-site                            exit=0 · WARN|ERROR 매치 행 0
       KO 187 / EN 185 / JA 185 / ZH 185 — 카드 기록 로그와 동일
AC-011 git diff --name-only bce6d7e08..HEAD -- docs-site CHANGELOG.md   10
       (CHANGELOG.md · e2e en,zh · doctor ko,ja,zh · skill-guide ko,en,ja,zh — 명명된 집합과 정확 일치)
AC-011 grep -c "t538" CHANGELOG.md    1  (12행, [Unreleased] → ### Added 최상단)
```

경계 3건:

```
git diff bce6d7e08..HEAD -- docs-site | grep -c '^+.*new-badge'   0
git diff bce6d7e08..HEAD -- docs-site | grep -ci "codex"          0   (t535 절 불변)
git diff --numstat  — doctor ko 1/1 · ja 3/1 · zh 3/1 · e2e en 8/2 · zh 8/2 · skill-guide 4/0 ×4 · CHANGELOG 2/0
```

### 검증식 정정 2건 — 각각 판정

**AC-008 — 계측기 수리다. 기준 완화가 아니다.**

`grep -c` 는 매치되는 **행** 수를 세므로 한 문단 안의 두 토큰은 상한이 1이고, 산문 문단에 두 계열을 쓰면 `≥2` 는 도달 불가다. 카드는 이 정정이 "옛 읽기와 새 읽기 양쪽에서 기준이 충족되므로 판정에 실리지 않는다"고 주장한다. 재측정 결과 **양쪽 다 4-로케일 전부 2** 로 주장이 성립한다.

정정을 정당화하는 결정적 사실은 따로 있다. 산출물이 두 문단인 이유가 잘못 세는 계측기에서 더 큰 수를 얻으려는 조정이 아님을 확인해야 하는데, `plan.md`:75 가 정정보다 **먼저** 쓰인 문서로서 "사실문 1-2행 추가"를 스스로 지시하고 있다:

```
$ /usr/bin/grep -n "사실문" .moai/specs/SPEC-DOCS-LOCALE-PARITY-REPAIR-001/plan.md
75:1. content-author: ko `skill-guide.md` … SVG060-064(…)·SVG070-074(커넥터 기하) 사실문 1-2행 추가.
```

문단 분할은 계획이 지시한 형태를 따른 것이다. 판정: **계측기 수리**.

**AC-011 — 계측기 결함이다. 골대 이동이 아니다.**

옛 읽기가 실패했다는 카드의 자기 신고를 그대로 받지 않고 재측정했다:

```
$ git diff --name-only bce6d7e08..bbc37b729 | wc -l      17
$ git diff --name-only bce6d7e08..fbc9cc6b1 | wc -l      18
$ git diff --name-only bce6d7e08..HEAD      | wc -l      20
```

카드 기록(17 → 18)과 정확히 일치하며, sync·backfill 두 커밋이 다시 20까지 밀어 올린다. 이 수는 카드가 자기 장부(SPEC 4종·보고서 4종)를 쓰는 동안 단조 증가하므로 **어떤 고정 수치도 이 식의 기준이 될 수 없다.** 이 카드만 실패하는 것이 아니라 규약을 따르는 어떤 카드도 통과할 수 없는 계측기다. 초과 8경로가 전부 규약상 필수 산출물이라는 주장도 재확인했다.

기준선(10파일)은 움직이지 않았고, 옛 식이 겨눈 대상(diff 범위)만 의도한 대상(콘텐츠 쓰기 표면)으로 교정됐다. 전체-브랜치 계수는 삭제되지 않고 **보고 항목(기준 아님)** 으로 §D.11 에 존치했다. 승인 주체가 저자에서 리드로 올라간 것도 옳은 처리다 — AC-008 에 쓰인 저자 측 안전 조건(양쪽 읽기 충족)이 여기서는 성립하지 않기 때문이다.

판정: **계측기 결함의 수리**. 다만 리드 재측정·승인 사실은 카드 내부 산문 외 별도 산출물이 없다(F4 참조) — 실질은 본 감사가 독립 재측정으로 대체 확인했다.

### RED-now 주장 — 가드 6건 전부 비공허

부재·존재 형 가드는 기능이 없어도 저절로 만족되므로 편집 전 RED 관측이 유일한 판별 증거다. 카드의 관측 기록을 신뢰하지 않고, 기준선 blob 에서 직접 재유도해 판별식을 다시 세웠다 — 위 Evidence 의 baseline 값이 그것이다.

| AC | 기준선 | 편집 후 | 비공허 성립 |
|---|---|---|---|
| AC-001 | 1 | 0 | 성립 |
| AC-002 | 1 | 0 | 성립 |
| AC-003 | en 0 / zh 0 | en 7 / zh 7 | 성립 |
| AC-006 | ja 0 / zh 0 | ja 2 / zh 2 | 성립 |
| AC-007 | ko 1 / ja 1 / zh 1 (en 0) | 전부 0 | 성립 |
| AC-008 | 4-로케일 0 | 4-로케일 2 | 성립 |

여섯 건 모두 기준선에서 기대와 반대 값을 실제로 냈다. 공허한 가드는 하나도 없다.

### 내용 정확성 — grep 계수를 넘어서

계수는 토큰의 존재만 세지 산문이 옳은 말을 하는지는 세지 않는다. 그래서 카드가 하지 않은 검사를 했다: en/zh 산문을 ko 정본이 아니라 **코드 층위 정본**(`.claude/skills/moai/workflows/e2e.md`)에 직접 대조했다.

```
:41   --platform web|mobile|desktop|desktop-native            → 플래그 행 일치
:74   desktop-native 마커 정의(AppKit/.vcxproj/Qt/GTK)          → 자동감지 행 일치
:82   "does NOT take this branch" 라우팅 의미론                  → 라우팅 절 일치
:84-86 Host-OS Rule                                            → 호스트 OS 문단 일치
:148-153 axcli / appium-mac2 · FlaUI.WebDriver / pywinauto ·
         dogtail / ydotool+xdotool+스크린샷, EXPERIMENTAL,
         버전 PIN, Wayland GNOME 한정                            → 3-OS 매트릭스 도구·폴백·비고 전건 일치
```

3-OS 매트릭스 행이 나르는 도구명·폴백·버전 PIN·실험적 표시·Wayland 제약이 정본과 어긋나는 항목은 없다. G1 의 전제(데스크탑-네이티브 레인이 실제로 착지했다)도 성립한다.

G2 가 두 로케일에 전파한 명령이 실재하는지도 확인했다:

```
$ /usr/bin/grep -rn "Use:" internal/cli/doctor_permission.go internal/cli/doctor_sandbox.go
internal/cli/doctor_permission.go:13:	Use:   "permission",
internal/cli/doctor_sandbox.go:31:	Use:   "sandbox",
```

zh 자연스러움: `主机 OS 规则` 의 `主机` 는 zh 문서 트리에서 이미 16회 쓰이는 용어이고, `(experimental)`·`(mixed)`·`(回退)` 처럼 영문 원어를 괄호로 덧붙이는 방식은 이 파일의 기존 관행이다. 번역투 영어 직역이 아니라 사이트의 기존 zh 어휘를 따랐다.

### Baseline-attribution

이 보고서의 모든 수치는 감사 트리 `.claude/worktrees/t538`, tip `df653dd8e` 에서 본 감사가 직접 실행해 관측한 값이다. 기준선은 blob `d787897ca` 에서 `git show` 로 재유도했다. 카드 산출물(`ac-verification.md`·`progress.md`·`verdict.md`)의 수치는 **비교 대상으로만** 읽었고 근거로 인용하지 않았다.

### Gaps — 관측하지 않은 것

- **원격 CI 판정**: 깨끗한 환경·darwin/windows 매트릭스 판정을 관측하지 않았다. 이 감사는 전부 로컬 관측이다.
- **Vercel 배포 반응**: docs-site 변경이 develop 에 들어갈 때 프리뷰/프로덕션이 어떻게 반응하는지 관측하지 않았다.
- **렌더 결과의 육안 확인**: hugo exit 0 과 페이지 수는 쟀으나 생성된 HTML 을 열어 표가 의도대로 그려지는지는 보지 않았다.
- **변경 9파일 밖 docs-site**: i18n 게이트를 사이트 전체로 돌리지 않았다.
- **리드의 AC-011 재측정 행위 자체**: 리드가 실제로 재측정했다는 사실은 관측 불가(F4).

### Residual-risk

- AC-007 의 스캔 범위가 훗날 doctor.md 밖으로 넓어지면, 이 카드가 새로 쓴 e2e 행 6줄이 그 정규식에 걸린다(F1). 현재 판정에는 영향이 없다.
- SVG 규칙 구간 구분자(ko en dash vs 파생 3로케일 ASCII 하이픈)는 카드가 알면서 남긴 편차다. 어떤 AC 도 이 부호를 재지 않으므로 조용히 남는다.

## 결함 목록

`blocking` 은 정확성 또는 SPEC 이 실제로 요구한 사항을 건드리는 것, `optional` 은 그 밖의 전부. 차단성 결함은 없다.

- **F1 [Low] [optional]** — `.moai/reports/t538/verdict.md` 잔여 위험 절 · 잔여 위험 목록 불완전: 이 카드가 새로 쓴 e2e 행이 AC-007 정규식의 계수를 en·zh 각각 2 → 5 로 올린다(추가 행 중 6줄이 매치). ko·ja 기준선이 이미 각 5이므로 이는 4-로케일 공통의 기존 행-레이블 관행(`**데스크탑 (Electron)**` 형태)을 그대로 따른 것이며 §5 가 겨눈 번역 주석과는 다른 구조다 — 그래서 결함이 아니라 **미기록**이 문제다. `acceptance.md` §E.3 이 이 스캔의 범위 확대를 이미 우려 항목으로 적어 두었으므로 상호작용이 기록됐어야 한다. — 필요한 수리: 잔여 위험 절에 한 줄 추가(재측정: `for f in $(git diff --name-only bce6d7e08..HEAD -- docs-site); do grep -cE '\*\*[^*]*\([^)]*\)[^*]*\*\*' $f; done` → e2e en 5 · zh 5 · 나머지 0).

- **F2 [Low] [optional]** — `docs-site/content/zh/utility-commands/moai-e2e.md`:78-80 · 로케일 간 행-레이블 형태 분기: zh 만 `**原生桌面 (macOS, desktop-native)**` 로 ASCII 토큰을 레이블 안에 넣고, ko `(macOS)` · ja `(macOS)` · en `(macOS)` 는 넣지 않는다. AC-004 가 중국어 파일에서 ASCII 토큰 히트 ≥4 를 요구하므로 계측기가 이 형태를 유도한 면이 있다. 파일의 기존 주석 관행과는 부합하나, 중국어 문어에서 원어 주석은 첫 등장 1회가 통례인데 여기서는 같은 용어를 5회 주석했다. — 필요한 수리(재량): 78-80 행 레이블을 `**原生桌面 (macOS)**` 로 되돌리고 토큰은 58·84·98·176 행에서 이미 4회 확보되므로 AC-004 기준(≥4)은 유지된다.

- **F3 [Low] [optional]** — `.moai/reports/t538/{verdict.md:8, ac-verification.md:4}` · 증거 앵커가 tip 보다 2커밋 뒤: 두 파일 모두 `after SHA: 91fd27bb6` 로 고정돼 있고 브랜치 tip 은 `df653dd8e` 다. 이 저장소의 규율은 "병합할 tip 에서 재측정"이다. 실질 영향은 없다 — `git diff --name-only 91fd27bb6..HEAD` 가 `.moai/` 장부 4파일뿐이라 콘텐츠 AC 는 불변이고, 본 감사가 tip 에서 11건 전건을 재측정해 성립을 확인했다. 다만 전체-브랜치 보고 수치는 18 → **20** 으로 움직였고 어느 파일에도 갱신되지 않았다. — 필요한 수리: 두 파일의 앵커를 tip 으로 갱신하거나, "tip 에서 재측정했고 콘텐츠 AC 불변" 한 줄 명기.

- **F4 [Info] [optional]** — `acceptance.md` `## HISTORY` AC-011 항목 · 리드 승인 증거가 카드 내부 산문뿐: "리드가 직접 재측정한 뒤 승인했다"를 뒷받침하는 리드 측 산출물이 없어 이 진술 자체는 검증 불가다. 실질은 닫혀 있다 — 정정의 근거(계측기 도달 불가)를 본 감사가 독립적으로 재측정해 확인했으므로 승인의 정당성은 증언에 의존하지 않는다. 기록으로만 남긴다.

- **F5 [Info] [optional]** — 카드의 내용 검증은 ko 정본 대조까지이고 코드 층위 정본 대조가 없다. "ko 정본 자체가 낡았을 수 있다"는 위험이 잔여 위험 목록에 없다. 본 감사가 `.claude/skills/moai/workflows/e2e.md` 대조로 닫았고(위 § 내용 정확성), 어긋난 항목은 없었다. 남는 것은 기록의 공백이지 살아 있는 위험이 아니다.

## 점수 산출

가중 조화평균 = 1 / (0.40/96 + 0.25/98 + 0.20/86 + 0.15/88)
= 1 / (0.0041667 + 0.0025510 + 0.0023256 + 0.0017045)
= 1 / 0.0107478 = **93.0**

Tier M 기준(≥0.80 / 80) 상회. must-pass 2차원 독립 통과. **PASS**.
