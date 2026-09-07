# Progress — SPEC-DOCS-LOCALE-PARITY-REPAIR-001

- 카드: t538 · 브랜치: `WT-docs-v313-locales` · 베이스: `bce6d7e08`
- 상태: **in-progress** (run 페이즈 M1~M3 착지 — 검증 완료, sync 미개시)

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
status: draft
spec: SPEC-DOCS-LOCALE-PARITY-REPAIR-001
tier: M
artifacts:
  - spec.md
  - plan.md
  - acceptance.md
  - progress.md
measurement_ssot: .moai/reports/t538/plan-phase.md
base: bce6d7e08
scope: G1 (e2e en/zh stale) + G2 (doctor ja/zh examples) + G3 (legacy emphasis spacing, 3 lines) + G4 (skill-guide SVG rules x4) + CHANGELOG
ac_count: 11
red_now_acs: [AC-001, AC-002, AC-003, AC-006, AC-007, AC-008]
milestones: [M1-G1, M2-G2G3, M3-G4-CHANGELOG]
kickoff_gate: pending (Implementation Kickoff Approval — HUMAN GATE)
```

## §E.2 Run-phase Evidence

측정 주체는 레인 오케스트레이터이며, 아래 값은 모두 이 워크트리에서 명시된 SHA 기준으로 관측됐다. 본 절은 그 관측을 귀속해 기록한 것이지, 이 기록을 쓰면서 다시 잰 값이 아니다.

- 베이스: `bce6d7e08`
- baseline(RED-now) 측정 시점: HEAD `d787897ca` (해당 시점까지 docs 무변경)
- after 측정 시점: HEAD `bbc37b729`
- 측정 도구: `/usr/bin/grep` (이 셸의 `grep` 은 ugrep 래퍼로 조용히 건너뛴다 — 절대경로로 고정)

### E.2.1 AC 판정 행렬

| AC | 검증식 | baseline | after | 판정 |
|---|---|---|---|---|
| AC-001 | `/usr/bin/grep -c "not yet provided" docs-site/content/en/utility-commands/moai-e2e.md` | 1 | 0 | PASS |
| AC-002 | `/usr/bin/grep -c "尚未提供" docs-site/content/zh/utility-commands/moai-e2e.md` | 1 | 0 | PASS |
| AC-003 | `/usr/bin/grep -ci "desktop-native" <en·zh e2e>` | en 0 · zh 0 | en 7 · zh 7 | PASS |
| AC-004 | 같은 식, 총 히트 ≥4 | en 0 · zh 0 | en 7 · zh 7 | PASS |
| AC-005 | 표 행 `/usr/bin/grep -c '^\|'` · 제목 `'^#'` · 호스트 OS 토큰 | en·zh 42행 | ko·en·ja·zh 46행 각 · 제목 18 각 | PASS |
| AC-006 | `/usr/bin/grep -cE '^moai doctor (permission\|sandbox)' <ja·zh doctor>` | 블록 한정 ja 0 · zh 0 / 전체 파일 ja 2 · zh 2 | 블록 한정 ja 2 · zh 2 / 전체 파일 ja 4 · zh 4 | PASS |
| AC-007 | `/usr/bin/grep -cE '\*\*[^*]*\([^)]*\)[^*]*\*\*' <ko·en·ja·zh doctor>` | ko 1 · en 0 · ja 1 · zh 1 | ko 0 · en 0 · ja 0 · zh 0 | PASS |
| AC-008 | `/usr/bin/grep -o "SVG060\|SVG070" <skill-guide ×4> \| wc -l` | 0 각 | 2 각 (4-로케일 전부) | PASS |
| AC-009 | `git diff --quiet bce6d7e08 -- <ko·ja e2e>` · `git diff bce6d7e08 -- <en·zh e2e> \| /usr/bin/grep -c '^+.*badge'` | — | exit 0 · 배지 추가 0 | PASS |
| AC-010 | `hugo --source docs-site` | — | exit 0 · WARN/ERROR 0 · sitemap.xml 존재 · 4-로케일 187/185/185/185 | PASS |
| AC-011 | `git diff --name-only bce6d7e08..HEAD -- docs-site CHANGELOG.md` (개정식 — E.2.3 참조) | — | 10파일(명명된 집합과 정확히 일치) · CHANGELOG t538 항목 1 | PASS |

AC-005 의 호스트 OS 토큰은 로케일별 원어로 각 1회 확인됐다 — ko `호스트 OS 규칙`, ja `ホスト OS ルール`, en `host OS rule`, zh `主机 OS 规则`. zh 토큰은 plan.md §D.5 가 문서화한 유도 절차를 따라 유도 시점에 확인했다.

AC-001·AC-002 는 부수 확인으로 `deferral` 토큰도 en 0 · zh 0 임을 함께 관측했다.

AC-006 의 증가분 2건은 양쪽 로케일 모두 예시 블록 안에서만 발생했고, 표 행은 변하지 않았다.

AC-008 은 개정된 출현-계수식과 폐기된 `grep -c` 식 **양쪽 모두에서** 4-로케일 전부 2를 냈다.

### E.2.2 i18n 게이트 — 변경된 문서 9파일

파일별로 URL 블랙리스트 0, Mermaid LR/RL 0, diff 내 `^+.*new-badge` 0, diff 내 본문 이모지 0.

이 측정에는 catch-all 대조군을 함께 걸었다 — `/usr/bin/grep -c "moai"` 가 9파일에서 각각 24/24/31/31/31/55/76/74/75 로 전부 0이 아니었고, 이로써 grep 이 실제로 파일에 도달했음이 성립한다.

**거짓 0 기록.** 이 두 게이트의 앞선 실행은 거짓 0을 냈다. 원인은 zsh 가 인용되지 않은 변수를 단어 분할하지 않는다는 점이다 — 파일 목록 전체가 파일명 하나로 grep 에 전달됐고, grep 은 "No such file or directory" 로 실패했다. 실패한 grep 의 빈 출력이 "위반 0건" 과 구분되지 않았다. 대조군을 세우고 나서야 도달 여부가 드러났다.

### E.2.3 검증식 결함 2건 — 양쪽 수리 완료

**(1) AC-008 검증식 결함 — 카드 안에서 수리됨.**
원래 식 `/usr/bin/grep -c "SVG060\|SVG070" … → ≥2` 는 도달 불가능했다. `grep -c` 는 매칭된 **행** 수를 세므로, 한 산문 문단 안에 토큰이 둘 있으면 이 명령의 상한은 1이다. ko 서술은 규칙 계열마다 한 문단씩, 두 문단으로 저작됐다 — 이는 plan.md M3 자신이 규정한 "사실문 1-2행 추가" 를 따른 것이지, 잘못 세는 검사가 더 큰 수를 내게 만들려는 조정이 아니다. 이 구분을 명시해 둔다. acceptance.md 의 검증식은 manager-spec 이 출현-계수 형태로 개정했고, 그 개정은 판정에 영향을 주지 않는다 — 두 읽기 모두에서 기준이 충족되기 때문이다(양쪽 다 2, 4-로케일 전부). 개정 기록은 acceptance.md 자신의 HISTORY 가 보유한다.

**(2) AC-011 검증식 결함 — 카드 안에서 수리됨(리드 승인).**
원래 식 `git diff --name-only bce6d7e08..HEAD` 에 "변경 파일 집합이 10개로 한정" 이라는 기준을 건 형태는 쓰인 그대로는 도달 불가능하다. 브랜치 전체 diff 에는 베이스 이후 커밋된 계획·증거 산출물이 필연적으로 포함되기 때문이다. 이 식은 `bbc37b729` 에서 17을, `fbc9cc6b1` 에서 18을 잰다.

**결정적 근거는 이 17→18 이동이다.** 계수는 카드가 자기 장부(SPEC 4종·보고서)를 써 나가는 동안 계속 증가하므로, 어떤 고정 수치도 이 식의 기준이 될 수 없다. 처음에는 이 사실을 "이 카드가 검사를 통과하지 못한다" 로 읽었으나, 계수의 이동이 드러내는 것은 그보다 넓다 — 규약을 따르는 **어떤** 카드도 이 기준에 도달할 수 없으며, 따라서 결함이 있는 쪽은 카드가 아니라 계측기다. 판정 대상의 실패가 아니라 측정 도구의 결함이라는 점이 개정의 근거다.

`fbc9cc6b1` 에서 초과분 8파일(`git diff --name-only bce6d7e08..HEAD -- ':!docs-site' ':!CHANGELOG.md'`)은 전부 규약을 따르는 카드가 구조적으로 산출하는 파일이다 — `.moai/reports/t538/` 의 4파일(`hugo-build.log`·`plan-audit-iter1.md`·`plan-audit-iter2.md`·`plan-phase.md`)과 SPEC 4종(`acceptance.md`·`plan.md`·`progress.md`·`spec.md`). 어느 하나도 이 카드에 고유한 잉여가 아니다.

**수리 내용.** acceptance.md §D 행과 §D.11 절이 범위 한정식 `git diff --name-only bce6d7e08..HEAD -- docs-site CHANGELOG.md` → 명명된 집합 10파일을 규정하도록 개정됐다. 기준선(bar) 10 자체는 바뀌지 않았다. 전체-브랜치 계수는 폐기되지 않고 §D.11 에 **보고 항목(기준 아님)** 으로 명시 표시돼 남는다 — 참고 수치로서의 값은 있으나 판정에는 쓰지 않는다.

**승인 주체는 저자가 아니라 리드다.** AC-008 과 달리 이 개정은 옛 읽기에서의 판정을 바꾼다 — 옛 식으로 읽으면 18 대 10 으로 **FAIL** 이다. AC-008 에서 저자 측 개정을 정당화한 안전 조건(두 읽기 모두에서 기준이 충족된다)이 여기에는 성립하지 않는다. 그래서 카드가 스스로 개정하지 않고 리드에게 보고했고, 리드가 직접 재측정한 뒤 승인했다. 경위 전체 — 옛 읽기가 FAIL 이었다는 사실과 승인 주체가 리드라는 사실을 포함해 — 는 acceptance.md 자신의 HISTORY 가 보유한다.

**개정 후 판정.** 개정식 아래에서 AC-011 은 `fbc9cc6b1` 에서 10을 재며 명명된 집합과 정확히 일치한다. CHANGELOG 쪽 절반도 충족된다: `/usr/bin/grep -c "t538" CHANGELOG.md` → 1, 항목은 `[Unreleased]` 최상단 `### Added` 에 놓였다. AC-011 은 **PASS** 로 기록한다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
phase: run
status: in-progress
spec: SPEC-DOCS-LOCALE-PARITY-REPAIR-001
tier: M
base: bce6d7e08
baseline_measured_at: d787897ca
run_commit_sha: bbc37b729
measurement_tool: /usr/bin/grep (shell grep is a ugrep wrapper — absolute path pinned)
ac_total: 11
ac_pass_count: 11
ac_pass_on_intended_subject: 0
ac_fail_count: 0
ac_matrix:
  AC-001: PASS
  AC-002: PASS
  AC-003: PASS
  AC-004: PASS
  AC-005: PASS
  AC-006: PASS
  AC-007: PASS
  AC-008: PASS
  AC-009: PASS
  AC-010: PASS
  AC-011: PASS
red_now_acs: [AC-001, AC-002, AC-003, AC-006, AC-007, AC-008]
red_now_confirmed_before_edit: true
red_now_green_after: true
milestones:
  M1-G1: complete
  M2-G2G3: complete
  M3-G4-CHANGELOG: complete
build:
  command: hugo --source docs-site
  exit_code: 0
  warn_error_count: 0
  sitemap_present: true
  locales_built: 4
  pages: [187, 185, 185, 185]
  log: .moai/reports/t538/hugo-build.log  # 작성 시점 untracked
i18n_gates:
  scope: 9 changed docs files
  url_blacklist: 0
  mermaid_lr_rl: 0
  new_badge_added: 0
  body_emoji_added: 0
  catch_all_control: "/usr/bin/grep -c \"moai\" → 24/24/31/31/31/55/76/74/75 (all non-zero — grep reached the files)"
  false_zero_recorded: "앞선 실행이 거짓 0. 원인: zsh 가 인용되지 않은 변수를 단어 분할하지 않아 파일 목록 전체가 파일명 하나로 전달돼 grep 이 실패했고, 실패의 빈 출력이 위반 0건과 구분되지 않았다."
open_items:
  - id: AC-008-expression-defect
    status: repaired-in-card
    detail: "grep -c 는 행을 세므로 한 문단 두 토큰의 상한이 1 — 원식 ≥2 는 도달 불가. acceptance.md 검증식을 출현-계수형으로 개정. 두 읽기 모두에서 2 → 판정 불변. 기록은 acceptance.md HISTORY."
  - id: AC-011-expression-defect
    status: repaired-in-card
    approved_by: lead
    detail: "git diff --name-only bce6d7e08..HEAD 에 10파일 기준을 건 형태는 도달 불가 — 브랜치 전체 diff 가 계획·증거 산출물을 필연 포함해 bbc37b729 에서 17, fbc9cc6b1 에서 18을 잰다. 이 17→18 이동이 결정적 근거다: 계수가 카드의 자기 장부 기록과 함께 증가하므로 어떤 고정 수치도 기준이 될 수 없고, 결함은 카드가 아니라 계측기에 있다. 초과분 8파일은 전부 규약을 따르는 카드의 구조적 산출물(.moai/reports/t538/ 4파일 + SPEC 4종). acceptance.md §D·§D.11 을 범위 한정식(-- docs-site CHANGELOG.md → 10)으로 개정, 기준선 10 불변, 전체-브랜치 계수는 '보고 항목(기준 아님)'으로 §D.11 에 존치. 옛 읽기에서는 18 대 10 으로 FAIL 이었고 AC-008 의 저자 측 안전 조건(두 읽기 모두 충족)이 성립하지 않으므로, 저자가 아니라 리드가 재측정 후 승인. 개정식 아래 fbc9cc6b1 에서 10 — 명명된 집합과 정확히 일치. 경위는 acceptance.md HISTORY."
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## Phase Log

| 일시 | 페이즈 | 내용 |
|---|---|---|
| 2026-09-08 | plan | 측정 SSOT(`plan-phase.md`) 인용 기반 SPEC 4종 생성. baseline 재관측 완료(en:170 deferral 1건, zh:170 1건, ja/zh 예시 4행, 강조위반 ko:41·ja:39·zh:39 각 1건/en 0건, skill-guide SVG0 언급 0건). 리드 추가지시 2건(전제-정정 기록·G4 1행 근거)과 코디네이터 정정 2회(G3 = 3행 한정·괄호 마커 밖, t535 신규 절 불가침) 반영. |
| 2026-09-08 | plan (iter1 수리) | plan-audit iter1 **FAIL 0.71**(`.moai/reports/t538/plan-audit-iter1.md` @ `86c8023b6`) → D1-D6 전건 수리 + F1·F2 수리, F3~F6 각하 기록(spec.md HISTORY). version 0.2.0. 핵심 실측 정정: doctor 예시행 블록-한정 앵커(ja/zh 0, 표행 2), zh 원어 토큰 `原生桌面`, CHANGELOG 배치 `### Added` 직하. 재감사는 결함 delta 범위로 iteration 2/2. |
| 2026-09-08 | run | M1-G1·M2-G2G3·M3-G4-CHANGELOG 착지(HEAD `bbc37b729`, 베이스 `bce6d7e08`). RED-now 6건 편집 전 RED 확인·편집 후 GREEN. AC 11건 중 10건 PASS, AC-011 은 의도된 대상 한정 PASS. hugo exit 0·WARN/ERROR 0·4-로케일 빌드. i18n 게이트 9파일 전건 0(catch-all 대조군으로 도달 확인, 앞선 거짓 0 원인 기록). 미해결 2건: AC-008 검증식 결함(카드 내 수리, 판정 불변) · AC-011 검증식 결함(미수리, 리드 보고). |
| 2026-09-08 | run (AC-011 검증식 개정) | AC-011 검증식을 전체-브랜치 diff 에서 범위 한정 diff(`-- docs-site CHANGELOG.md`)로 개정, 기준선 10 불변. 근거는 전체-브랜치 계수의 17(`bbc37b729`)→18(`fbc9cc6b1`) 이동 — 카드가 자기 장부를 쓰는 동안 계수가 증가하므로 규약을 따르는 어떤 카드도 도달할 수 없고, 결함은 카드가 아니라 계측기에 있다. 옛 읽기에서는 18 대 10 으로 FAIL 이었고 AC-008 의 저자 측 안전 조건이 성립하지 않아 **리드가 재측정 후 승인**했다. 전체-브랜치 계수는 §D.11 에 보고 항목(기준 아님)으로 존치. 개정식 아래 AC-011 은 10 — 명명된 집합과 정확히 일치. AC 11/11 PASS. |
