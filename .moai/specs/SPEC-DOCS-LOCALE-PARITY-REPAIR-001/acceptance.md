# Acceptance — SPEC-DOCS-LOCALE-PARITY-REPAIR-001

> 하니스: **standard** · 측정 근거: `.moai/reports/t538/plan-phase.md` (base `bce6d7e08`)
> 모든 검증 grep 은 `/usr/bin/grep` (REQ-012). RED-now AC 는 baseline(현재) vs after(수리 후)로 기술된다.

## HISTORY

- 2026-09-08: AC-008 검증식 정정 — 매치 **행** 계수에서 **출현 수** 계수로. 기준선(bar)은 그대로다.
  - before: `/usr/bin/grep -c "SVG060\|SVG070" docs-site/content/{ko,en,ja,zh}/advanced/skill-guide.md`
  - after: `/usr/bin/grep -o "SVG060\|SVG070" docs-site/content/<locale>/advanced/skill-guide.md | wc -l`
  - 사유: `grep -c` 는 매치되는 **행**의 수를 센다. 산문 마크다운에서는 한 문단이 한 행이므로, 두 계열이 한 문단 안에 있으면 이 명령의 상한이 1이 되어 `≥2` 기준은 산출물이 아무리 정확해도 도달할 수 없다. 계측기가 기준이 말하는 대상을 재지 못한 것이다.
  - **이 정정은 판정에 영향을 주지 않는다. 옛 읽기와 새 읽기 모두에서 기준이 충족되기 때문이다.** 이 트리·이 실행에서 레인 오케스트레이터가 실측한 값:
    - `/usr/bin/grep -c "SVG060\|SVG070" docs-site/content/{ko,en,ja,zh}/advanced/skill-guide.md` → ko `2` · en `2` · ja `2` · zh `2`
    - `/usr/bin/grep -o "SVG060\|SVG070" docs-site/content/<locale>/advanced/skill-guide.md | wc -l` → ko `2` · en `2` · ja `2` · zh `2`
  - 산출물이 두 행인 이유: 두 규칙 계열(접근성-이름 / 커넥터-기하)을 각각 한 문단으로 나눠 서술했기 때문이며, 이는 plan.md M3 이 스스로 지시한 "사실문 1-2행 추가"를 따른 것이다. 잘못 세는 검사에서 더 큰 수를 얻으려고 행을 나눈 것이 아니다.

## §D AC Matrix

| AC | 소관 | 검증 | 기대 (baseline → after) |
|---|---|---|---|
| AC-001 | G1 | en deferral 문장 제거 (RED-now) | 1 → 0 |
| AC-002 | G1 | zh deferral 문장 제거 (RED-now) | 1 → 0 |
| AC-003 | G1 | en·zh 플래그 행 desktop-native (RED-now) | 0 → ≥1 (각) |
| AC-004 | G1 | en·zh desktop-native 총 히트 (플래그+3-OS+자동감지) | 0 → ≥4 (각) |
| AC-005 | G1 | en·zh 표 행 수 = ko 패리티 | 불일치 → 동일 |
| AC-006 | G2 | ja·zh doctor 예시행 2줄 (RED-now, 예시 블록 한정) | 블록 한정 0 → 2 · 전체 파일 2 → 4 (각) |
| AC-007 | G3 | 전체-파일 bold-내부-괄호 스캔 (RED-now) | ko/ja/zh 1 → 0, en 0 유지 |
| AC-008 | G4 | skill-guide ×4 SVG060/070 언급 — 출현 수 계수 (`grep -o` + `wc -l`) (RED-now) | 0 → ≥2 (각) |
| AC-009 | 경계 | ko·ja e2e 무변경 + 배지 추가 없음 | — |
| AC-010 | 빌드 | hugo exit 0 · WARN/ERROR 0 | — |
| AC-011 | 착지 | 4-로케일 동일 착지 + CHANGELOG t538 항목 | — |

## §D.1 AC-001 — en deferral 문장 제거 (RED-now)

- **Given** base `bce6d7e08` 의 `docs-site/content/en/utility-commands/moai-e2e.md` :170 이 "OS-level native-desktop automation **is not yet provided**"를 포함한다 (2026-09-08 실측 1건).
- **When** `/usr/bin/grep -c "not yet provided" docs-site/content/en/utility-commands/moai-e2e.md`
- **Then** baseline `1` → after `0`. 라우팅 의미론(ko:176 파생)이 같은 절을 대체한다.

## §D.2 AC-002 — zh deferral 문장 제거 (RED-now)

- **Given** zh `moai-e2e.md` :170 이 "原生桌面自动化尚未提供"를 포함한다 (실측 1건).
- **When** `/usr/bin/grep -c "尚未提供" docs-site/content/zh/utility-commands/moai-e2e.md`
- **Then** baseline `1` → after `0`.

## §D.3 AC-003 — en·zh 플래그 행 (RED-now)

- **Given** base 에서 en·zh 의 `--platform` 플래그 행이 `web|mobile|desktop` 으로 끝나 desktop-native 선택지가 없다 (`desktop-native` 히트 0 — 실측).
- **When** `/usr/bin/grep -c "desktop-native" docs-site/content/en/utility-commands/moai-e2e.md docs-site/content/zh/utility-commands/moai-e2e.md`
- **Then** 각 파일 `0` → `≥1` (플래그 행). ko 정본 = :58 형태.

## §D.4 AC-004 — en·zh desktop-native 축 전체 (RED-now)

- **Given** base 에서 en·zh 에 데스크톱-네이티브 관련 행이 전무하다 (플래그 0 + 3-OS 매트릭스 0 + 자동감지 0).
- **When** `/usr/bin/grep -ci "desktop-native" <en|zh moai-e2e.md>`
- **Then** 각 `0` → `≥4` (플래그 1 + 3-OS 매트릭스 3 + 자동감지 마커 행 1 이상). 내용은 ko:78-80·:98 정본 파생, 표 형태 참조는 ja:78-80.

## §D.5 AC-005 — 표 행 수 패리티 + 호스트 OS 규칙 문단 (ko 정본 대비)

- **Given** ko `moai-e2e.md` 는 201행·완비 상태다. 헤딩 수는 4-로케일 이미 18로 동일하므로 헤딩 패리티는 호스트-OS 문단(평문, ko:84)을 못 잡는다 — 문단 존재는 전용 원어-토큰 grep 으로 판정한다 (plan-audit iter1 D2 정정).
- **When** `/usr/bin/grep -c '^|' docs-site/content/{ko,en,zh}/utility-commands/moai-e2e.md` 로케일별 비교 · `/usr/bin/grep -c '^#'` 4-로케일 비교 · 호스트-OS 문단: ko `/usr/bin/grep -c "호스트 OS 규칙" docs-site/content/ko/utility-commands/moai-e2e.md`, ja `/usr/bin/grep -c "ホスト OS ルール" docs-site/content/ja/utility-commands/moai-e2e.md`, en `/usr/bin/grep -c "host OS rule" docs-site/content/en/utility-commands/moai-e2e.md`, zh — ko 정본 파생 문구 기준, 파생 시점 트리에서 확정 후 기록 (zh 자연어 문장이므로 이 SPEC 시점엔 존재 검증 불가 — 문서화된 확정 절차)
- **Then** after 상태에서 en·zh 표 행 총수가 ko 와 동일(±0), 헤딩 수 4-로케일 동일 유지, 호스트-OS 토큰: ko `1` (baseline 유지 — 무변경 증명), ja `1` (동일), en `≥1` (0 → 신규), zh `≥1` (0 → 신규, 확정 토큰으로).

## §D.6 AC-006 — ja·zh doctor 예시행 (RED-now, 예시 블록 한정)

- **Given** ja:106-111·zh:106-111 예시 블록이 `moai doctor hook` 에서 끝나 4행이다 (실측). **주의**: 파일 전체 스캔은 명령 표 행(ja·zh :32-33, 행두 `|`)도 세므로 기대값이 다르다 — 전체 파일 기준 ja·zh baseline 2, 표행은 수리 후에도 불변 (plan-audit iter1 D1 정정).
- **When** (블록 한정 — 행두 앵커로 fenced 예시행만 계수, 표 행 배제) `/usr/bin/grep -cE '^moai doctor (permission|sandbox)' docs-site/content/ja/cli-reference/doctor.md docs-site/content/zh/cli-reference/doctor.md` 및 (전체 파일 — 표행 포함) `/usr/bin/grep -c "moai doctor permission\|moai doctor sandbox" <같은 두 파일>`
- **Then** 앵커 패턴: 각 `0` → `2`. 전체 파일: 각 `2` → `4` (증분 +2는 전부 예시 블록 — 표행 불변의 증명). 구성은 ko:120-121·en:118-119 정본과 동등, 주석은 각 로케일 자연어.

## §D.7 AC-007 — 전체-파일 강조 간격 스캔 (RED-now)

- **Given** base 에서 위반 패턴(`**…(…)…**` — 괄호가 bold 마커 안)이 ko:41·ja:39·zh:39 에 각 1건, en 에 0건이다 (2026-09-08 전체-파일 스캔 실측). t535 신규 절(ko:69·ja:67·zh:67)은 마커 없는 서술로 이미 통과.
- **When** `/usr/bin/grep -cE '\*\*[^*]*\([^)]*\)[^*]*\*\*' docs-site/content/{ko,en,ja,zh}/cli-reference/doctor.md`
- **Then** after: ko `0`, en `0`, ja `0`, zh `0`. 수리 방향은 `**권고** (advisory)` 형태(i18n 숙칙 §5 — 괄호 밖).

## §D.8 AC-008 — skill-guide SVG 규칙 (RED-now)

- **Given** base 에서 `advanced/skill-guide.md` ×4 가 SVG060-064·SVG070-074 를 언급하지 않는다 (ko 실측 0건 — svg-infographic 서술은 ko:158·:163 에 있음).
- **When** `/usr/bin/grep -o "SVG060\|SVG070" docs-site/content/<locale>/advanced/skill-guide.md | wc -l` 을 ko·en·ja·zh 각각에 적용한다 (계수 단위는 **출현 수**이며 매치 행 수가 아니다 — 아래 HISTORY 2026-09-08 항목 참조).
- **Then** 각 `0` → `≥2` (접근성-이름 계열 1 + 커넥터-기하 계열 1). ko 정본 → en/ja/zh 파생.

## §D.9 AC-009 — 경계 준수 (ko·ja e2e 무변경 + 배지 금지)

- **Given** ko·ja `moai-e2e.md` 는 이미 정확하며 배지가 없다.
- **When** `git diff --quiet bce6d7e08 -- docs-site/content/ko/utility-commands/moai-e2e.md docs-site/content/ja/utility-commands/moai-e2e.md` (exit 0 기대) 및 `git diff bce6d7e08 -- docs-site/content/en/utility-commands/moai-e2e.md docs-site/content/zh/utility-commands/moai-e2e.md | /usr/bin/grep -c '^+.*badge'`
- **Then** 첫 명령 exit 0 (무변경), 둘째 `0` (배지 마크업 추가 없음 — 현실 정렬이므로).

## §D.10 AC-010 — hugo 빌드 게이트

- **When** `hugo --source docs-site 2>&1 | /usr/bin/grep -ciE "WARN|ERROR"; hugo --source docs-site >/dev/null 2>&1; echo $?` (plan-audit iter1 F1 — `--quiet` 은 WARN/ERROR 계수 출력을 억제하므로 미사용)
- **Then** WARN/ERROR 매치 행 `0`, hugo exit `0`. 추가로 verify 레시피의 Mermaid TD-only grep·URL 블랙리스트 grep·본문 이모지 스캔 통과.

## §D.11 AC-011 — 4-로케일 동일 착지 + CHANGELOG

- **When** `git diff --name-only bce6d7e08..HEAD` 및 `/usr/bin/grep -n "t538" CHANGELOG.md`
- **Then** 변경 파일 집합이 10개로 한정된다: e2e en·zh(2) + doctor ko·ja·zh(3) + skill-guide ×4 + `CHANGELOG.md`(1) — ko·ja e2e 및 Codex Wiring 절 미포함. CHANGELOG 에는 `[Unreleased]` → `### Added` **최상단**에 `t538` 카드 id + G2 두-로케일 정정 명시 항목이 ≥1건 있다(t535 선례 = `### Added` 직하 배치 — `### Docs` 섹션은 현행 [Unreleased] 에 부재, plan-audit iter1 D6 정정). 전체 집합은 하나의 커밋 체인(REQ-011).

## §E Edge Cases

1. **en vs zh 표현 분기**: zh 파생 시 en 문장 복사 금지 — zh 자연어 검증(원어 토큰 `原生桌面` 계열 존재 확인 — D4 정정 토큰).
2. **표 열 구조**: en/zh 매트릭스는 ja:78-80 의 열 구조(도구/폴백/비고)와 ko 내용의 합성 — 열 수 불일치 시 AC-005 가 잡는다.
3. **간격 수리의 과잉 적용**: REQ-008 패턴이 doctor.md 밖 다른 페이지에서 발견돼도 본 SPEC 범위 밖이다 (파일-내부 정규화 한정).
4. **`grep -c` 종료코드**: 히트 0이면 grep 은 exit 1 — 판정은 출력 카운트로, exit 코드로 하지 않는다.
5. **zh 원어 토큰**: zh 페이지의 기존 용어 계열은 `原生桌面`(zh:170 `原生桌面应用`·`原生桌面自动化` 실측 1행)이다 — `桌面原生` 이 아님 (plan-audit iter1 D4 정정). zh 파생은 이 기존 용어를 따른다.

## §F Quality Gate (harness: standard)

- grep 카운트 AC 전부 통과 (AC-001~009)
- hugo 빌드 경고 0 (AC-010)
- 4-로케일 동일 착지 + 파일 집합 한정 (AC-011)
- Conventional Commit 체인, 커밋 메시지에 `t538` 포함

## §G Definition of Done

1. AC-001~AC-011 전부 PASS (커맨드 출력 인용).
2. RED-now AC 6건(AC-001·002·003·006·007·008)이 baseline 값과 after 값 쌍으로 기록됨.
3. `.moai/specs/SPEC-DOCS-LOCALE-PARITY-REPAIR-001/progress.md` §E.2 에 run 페이즈 증거 적재.
