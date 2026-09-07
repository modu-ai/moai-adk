# Acceptance — SPEC-DOCS-LOCALE-PARITY-REPAIR-001

> 하니스: **standard** · 측정 근거: `.moai/reports/t538/plan-phase.md` (base `bce6d7e08`)
> 모든 검증 grep 은 `/usr/bin/grep` (REQ-012). RED-now AC 는 baseline(현재) vs after(수리 후)로 기술된다.

## §D AC Matrix

| AC | 소관 | 검증 | 기대 (baseline → after) |
|---|---|---|---|
| AC-001 | G1 | en deferral 문장 제거 (RED-now) | 1 → 0 |
| AC-002 | G1 | zh deferral 문장 제거 (RED-now) | 1 → 0 |
| AC-003 | G1 | en·zh 플래그 행 desktop-native (RED-now) | 0 → ≥1 (각) |
| AC-004 | G1 | en·zh desktop-native 총 히트 (플래그+3-OS+자동감지) | 0 → ≥4 (각) |
| AC-005 | G1 | en·zh 표 행 수 = ko 패리티 | 불일치 → 동일 |
| AC-006 | G2 | ja·zh doctor 예시행 2줄 (RED-now) | 0 → 2 (각) |
| AC-007 | G3 | 전체-파일 bold-내부-괄호 스캔 (RED-now) | ko/ja/zh 1 → 0, en 0 유지 |
| AC-008 | G4 | skill-guide ×4 SVG060/070 언급 (RED-now) | 0 → ≥2 (각) |
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

## §D.5 AC-005 — 표 행 수 패리티 (ko 정본 대비)

- **Given** ko `moai-e2e.md` 는 201행·완비 상태다.
- **When** `/usr/bin/grep -c '^|' docs-site/content/{ko,en,zh}/utility-commands/moai-e2e.md` 를 각 로케일별 비교
- **Then** after 상태에서 en·zh 의 표 행 총수가 ko 와 동일하다(±0). 호스트 OS 규칙 문단 존재는 헤딩 패리티로 보강: `/usr/bin/grep -c '^#'` 가 4-로케일 동일.

## §D.6 AC-006 — ja·zh doctor 예시행 (RED-now)

- **Given** ja:106-111·zh:106-111 예시 블록이 `moai doctor hook` 에서 끝나 4행이다 (실측).
- **When** `/usr/bin/grep -c "moai doctor permission\|moai doctor sandbox" docs-site/content/ja/cli-reference/doctor.md docs-site/content/zh/cli-reference/doctor.md`
- **Then** 각 `0` → `2`. 구성은 ko:120-121·en:118-119 정본과 동등, 주석은 각 로케일 자연어.

## §D.7 AC-007 — 전체-파일 강조 간격 스캔 (RED-now)

- **Given** base 에서 위반 패턴(`**…(…)…**` — 괄호가 bold 마커 안)이 ko:41·ja:39·zh:39 에 각 1건, en 에 0건이다 (2026-09-08 전체-파일 스캔 실측). t535 신규 절(ko:69·ja:67·zh:67)은 마커 없는 서술로 이미 통과.
- **When** `/usr/bin/grep -cE '\*\*[^*]*\([^)]*\)[^*]*\*\*' docs-site/content/{ko,en,ja,zh}/cli-reference/doctor.md`
- **Then** after: ko `0`, en `0`, ja `0`, zh `0`. 수리 방향은 `**권고** (advisory)` 형태(i18n 숙칙 §5 — 괄호 밖).

## §D.8 AC-008 — skill-guide SVG 규칙 (RED-now)

- **Given** base 에서 `advanced/skill-guide.md` ×4 가 SVG060-064·SVG070-074 를 언급하지 않는다 (ko 실측 0건 — svg-infographic 서술은 ko:158·:163 에 있음).
- **When** `/usr/bin/grep -c "SVG060\|SVG070" docs-site/content/{ko,en,ja,zh}/advanced/skill-guide.md`
- **Then** 각 `0` → `≥2` (접근성-이름 계열 1 + 커넥터-기하 계열 1). ko 정본 → en/ja/zh 파생.

## §D.9 AC-009 — 경계 준수 (ko·ja e2e 무변경 + 배지 금지)

- **Given** ko·ja `moai-e2e.md` 는 이미 정확하며 배지가 없다.
- **When** `git diff --quiet bce6d7e08 -- docs-site/content/ko/utility-commands/moai-e2e.md docs-site/content/ja/utility-commands/moai-e2e.md` (exit 0 기대) 및 `git diff bce6d7e08 -- docs-site/content/en/utility-commands/moai-e2e.md docs-site/content/zh/utility-commands/moai-e2e.md | /usr/bin/grep -c '^+.*badge'`
- **Then** 첫 명령 exit 0 (무변경), 둘째 `0` (배지 마크업 추가 없음 — 현실 정렬이므로).

## §D.10 AC-010 — hugo 빌드 게이트

- **When** `hugo --source docs-site --quiet; echo $?` (또는 hns-oss-docs-verify 레시피의 빌드 절차)
- **Then** exit `0`, WARN `0`, ERROR `0`. 추가로 verify 레시피의 Mermaid TD-only grep·URL 블랙리스트 grep·본문 이모지 스캔 통과.

## §D.11 AC-011 — 4-로케일 동일 착지 + CHANGELOG

- **When** `git diff --name-only bce6d7e08..HEAD` 및 `/usr/bin/grep -n "t538" CHANGELOG.md`
- **Then** 변경 파일 집합이 10개로 한정된다: e2e en·zh(2) + doctor ko·ja·zh(3) + skill-guide ×4 + `CHANGELOG.md`(1) — ko·ja e2e 및 Codex Wiring 절 미포함. CHANGELOG 에는 `[Unreleased]` → `### Docs` → `### Added` 최상단에 `t538` 카드 id + G2 두-로케일 정정 명시 항목이 ≥1건 있다. 전체 집합은 하나의 커밋 체인(REQ-011).

## §E Edge Cases

1. **en vs zh 표현 분기**: zh 파생 시 en 문장 복사 금지 — zh 자연어 검증(원어 토큰 桌面原生 계열 존재 확인).
2. **표 열 구조**: en/zh 매트릭스는 ja:78-80 의 열 구조(도구/폴백/비고)와 ko 내용의 합성 — 열 수 불일치 시 AC-005 가 잡는다.
3. **간격 수리의 과잉 적용**: REQ-008 패턴이 doctor.md 밖 다른 페이지에서 발견돼도 본 SPEC 범위 밖이다 (파일-내부 정규화 한정).
4. **`grep -c` 종료코드**: 히트 0이면 grep 은 exit 1 — 판정은 출력 카운트로, exit 코드로 하지 않는다.

## §F Quality Gate (harness: standard)

- grep 카운트 AC 전부 통과 (AC-001~009)
- hugo 빌드 경고 0 (AC-010)
- 4-로케일 동일 착지 + 파일 집합 한정 (AC-011)
- Conventional Commit 체인, 커밋 메시지에 `t538` 포함

## §G Definition of Done

1. AC-001~AC-011 전부 PASS (커맨드 출력 인용).
2. RED-now AC 4건(AC-001·002·003·006·007·008)이 baseline 값과 after 값 쌍으로 기록됨.
3. `.moai/specs/SPEC-DOCS-LOCALE-PARITY-REPAIR-001/progress.md` §E.2 에 run 페이즈 증거 적재.
