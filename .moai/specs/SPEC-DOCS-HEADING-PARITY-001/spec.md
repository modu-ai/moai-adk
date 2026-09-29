---
id: SPEC-DOCS-HEADING-PARITY-001
title: "docs-site 4로케일 헤딩 패리티 정렬 — 표적 3페이지 구조 재유도 + 래칫 프루닝"
version: "0.1.0"
status: draft
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "docs-site/content"
lifecycle: spec-anchored
tags: "docs, i18n, locale-parity, docs-site, heading-parity, t1328"
tier: M
related_specs: [SPEC-DOCS-LOCALE-PARITY-REPAIR-001, SPEC-AGENT-MODEL-INHERIT-DOCS-001]
---

# SPEC: docs-site 4로케일 헤딩 패리티 정렬 (card t1328)

## HISTORY

- v0.1.0 (2026-09-29) — manager-spec 초판. 카드 t1328(Class C) plan-phase 산출물. 레인 측정치를 재검증해 사전 결정 테이블(§2)과 패리티 카운터 정의(§4)를 확정했다.

## 1. 개요

docs-site의 ko/en/ja/zh 4로케일 페이지 가운데 헤딩 구조가 갈라진 표적 3페이지(`advanced/agent-guide.md`, `multi-llm/_index.md`, `advanced/tokenomics-overview.md`)를 페이지별로 정한 정렬 방향에 따라 재유도해, 고정된 패리티 카운터(§4) 기준 4로케일 헤딩 수를 일치시킨다. 정렬이 끝난 페이지는 래칫 파일 `docs-site/.locale-parity-baseline`에서 프루닝한다(수렴 항목 삭제만 허용, 추가 금지). 마감 검증은 `hns-oss-docs-verify` 레시피 전체로 한다.

이 SPEC의 1급 산출물은 §2의 **사전 결정 테이블**이다. 페이지별로 ko-캐노닉 원칙(2026-08-17 승격, ko가 원본이고 타 로케일은 파생)과 실제 파생 이력을 대조해 정렬 방향을 확정한다. 특히 `tokenomics-overview.md`는 en 단독 사실 수리 2건이 존재해, ko→타 로케일 나이브 덮어쓰기는 그 수리를 파괴하므로 내용 보존 정렬이 강제된다.

## 2. 사전 결정 테이블 (first-class)

아래 표가 이 SPEC의 핵심 결정이다. run phase는 이 표를 재확인한 뒤 착수하며, 표를 바꿔야 할 경우 orchestrator를 통해 사용자 확인을 받는다(mid-run D-NEW-1 패턴).

### 2.1 표적 3페이지

| 페이지 | 카운터 값 (ko/en/ja/zh) | 캐노닉 생성 결정 | 정렬 방향 | 근거 커밋 (git 이력 재검증 완료) |
|---|---|---|---|---|
| `advanced/agent-guide.md` | 22 / 30 / 30 / 30 | ko-캐노닉 (구조 기준 = ko) | en/ja/zh → ko 구조로 재유도 | 최신 커밋 `5f4f199ef` (2026-09-29, t1300 M1 A-cluster rewrite)이 4로케일을 한 커밋에 함께 손봤는데도 구조 불균형이 발생 — 스테일 파생이 아니라 재작성 자체가 불균등했다. `9a53efd24`(t1115)도 agent-guide는 4로케일 동시 수정(검증됨) → en 단독 수리 없음 |
| `multi-llm/_index.md` | 11 / 5 / 5 / 5 | ko-캐노닉 — ko가 5세대 앞섬 | en/ja/zh → ko 재유도 (en/ja/zh는 스테일) | 마지막 4로케일 터치 `ce79ef7ca` (2026-09-12, t649). 이후 ko 단독 커밋 5건(전부 ko 파일만 터치 — 검증됨): `7897e4d8a`(t918, 09-18), `a29db12ec`(t1094, 09-23), `81c10109e`(t1095, 09-23), `be43782e2`(t1118, 09-23), `fbd13bbfc`(t1300 M3, 09-29) |
| `advanced/tokenomics-overview.md` | 12 / 13 / 12 / 11 | ko가 구조 기준 — 단, en 단독 사실 수리 2건 보존 | ko→en/ja/zh 재유도하되 en 단독 사실 수정 이월(carry-forward) + 생존 검사 | en 단독: `9a53efd24`(t1115 sync-audit FAIL 수리, en만), `7ad954556`(t1095 Sonnet 4.5/4.6 1M 경계, en+ko). 4로케일: `a9d9779d9`(t1115 tier 서술), `81c10109e`(t1095 1M 문맥), `a29db12ec`(t1094), `5f4f199ef`(t1300 M1). 병합 `64868b3e4`는 en+ko |

### 2.2 결정 유형별 처리 규칙

- **스테일 파생**(multi-llm): canonical(ko)이 명백히 앞서므로 파생 로케일 전체를 ko의 최신 구조로 재유도한다. 파생 쪽에 보존할 단독 수정이 없음을 이력으로 확인했다.
- **동시 재작성 불균형**(agent-guide): 4로케일 동시 커밋이 구조를 불균등하게 만든 경우다. ko-캐노닉 원칙에 따라 ko를 기준으로 삼되, 30→22 접기 과정에서 en/ja/zh에만 있는 사실이 사라지지 않는지 §3 REQ-006의 내용 보존 검사로 확인한다.
- **혼합 이력**(tokenomics): 구조는 ko 기준으로 맞추되, en 단독 사실 수리(`9a53efd24`, `7ad954556`가 만든 문장)는 실시간 재유도 도중 유실되지 않아야 한다. ja/zh는 재유도를 통해 이 수정 사실을 처음 얻게 된다(현재 ja=12, zh=11은 수정 미반영 스테일).

## 3. 요구사항 (GEARS)

- **REQ-001** — The run phase shall align the three target pages exactly per the §2 pre-decision table: `advanced/agent-guide.md` en/ja/zh re-derived to ko's heading structure, `multi-llm/_index.md` en/ja/zh re-derived to ko, `advanced/tokenomics-overview.md` re-derived with ko's structure as the base while carrying forward the en-only factual corrections.
- **REQ-002** — The run phase shall use the pinned parity counter (§4) for every section count it produces, measures, or reports in this SPEC's scope; no other counting method substitutes for it.
- **REQ-003** — When the alignment of a target page completes, the page shall satisfy ko == en == ja == zh under the pinned counter.
- **REQ-004** — When an alignment would remove or alter a factual correction introduced by `9a53efd24` or `7ad954556` in `advanced/tokenomics-overview.md`, the run phase shall carry that correction forward into the realigned text of all four locales before declaring the page aligned.
- **REQ-005** — The run phase shall record, per target page, a before/after heading table (per-locale counts + section-title list) under `.moai/reports/t1328/`, together with the canonical-generation verdict that restates the §2 decision with its evidence commit SHAs.
- **REQ-006** — The run phase shall perform a content-preservation check on every collapsing alignment (agent-guide 30→22, tokenomics) that diffs the section inventory of each derived locale before and after, and shall document any section removed as intentional with its content disposition (merged / moved / confirmed-absent-from-ko).
- **REQ-007** — When a target page reaches 4-locale parity under the pinned counter, the run phase shall remove exactly that page's line from `docs-site/.locale-parity-baseline`; the ratchet update shall contain no added lines.
- **REQ-008** — The closing verification shall execute the `hns-oss-docs-verify` recipe in full and shall satisfy every must_pass dimension: warning-free `hugo --minify --gc` build with sitemap present, zero NEW section-count divergence (`comm -23` prints nothing), README 4-file heading parity unchanged, TD-only Mermaid, URL blacklist clean.
- **REQ-009** — When a structural fix requires touching navigation config (`hugo.toml`, `_meta.yaml`, `data/menu/main.yaml`), the run phase shall route that fragment to the structure-curator specialist and keep content rewrite with the content-author / locale-translator specialists.

## 4. 패리티 카운터 정의 (pinned)

이 SPEC이 고정하는 유일한 카운터다. 출처: `.claude/skills/hns-oss-docs-verify/SKILL.md` §4 (recipe 원문). 카드가 적어 둔 "공식 수치"(22/30, 11/5, 12/13/12/11)가 바로 이 카운터 출력과 일치함을 2026-09-29 본 워크트리에서 재측정해 확인했다.

- **정의**: `grep -rc '^#\{2,\} '` — 2개 이상의 `#` 뒤 공백이 오는 줄(H2 이하 전 계층)을 센다. H1(`# `)은 제외되고, 코드 펜스 내부의 패턴 적중 줄도 포함된다(펜스 제외 없음 — 전후 측정이 같은 정의를 쓰는 한 일관적이다).
- **측정 단위**: 페이지별, 로케일별. 트리 합계 비교는 반대 방향 편차가 상쇄되므로 패리티 판정으로 쓰지 않는다.
- **레시피 (hns-oss-docs-verify §4 원문)**:

```bash
cd docs-site/content
grep -rc '^#\{2,\} ' ko en ja zh --include='*.md' \
| awk -F: '
    { i=index($1,"/"); loc=substr($1,1,i-1); page=substr($1,i+1)
      n[page,loc]=$2; pages[page]=1 }
    END { for (p in pages)
            if (n[p,"en"]!=n[p,"ko"] || n[p,"ja"]!=n[p,"ko"] || n[p,"zh"]!=n[p,"ko"])
              print p }' \
| sort > /tmp/parity-now.txt
grep -v '^#' ../.locale-parity-baseline | grep -v '^[[:space:]]*$' | sort > /tmp/parity-base.txt
comm -23 /tmp/parity-now.txt /tmp/parity-base.txt   # NEW divergence  -> FAIL
comm -13 /tmp/parity-now.txt /tmp/parity-base.txt   # converged pages -> prune baseline
```

- **FAIL 조건**: 첫 번째 `comm`이 한 줄이라도 출력하면 FAIL(신규 발산). 두 번째 `comm`은 수렴 페이지 목록으로, 래칫 프루닝(REQ-007)의 입력이다.
- **래칫 파일 규약**: `docs-site/.locale-parity-baseline` 헤더 문언 그대로 — 재생성은 수렴 페이지 삭제(프루닝) 목적으로만 하며, 줄을 추가하는 것은 새 부채를 인정하는 행위다.

## 5. 제약

- 본 SPEC의 집행(run phase)은 `hns-oss-docs-run` Runner의 author → translate → verify 파이프라인으로 라우팅할 것을 전제로 한다. 스폰 형태(author 전용 스폰 / Runner 위임)는 orchestrator가 run 시점에 결정한다.
- ko-캐노닉 원칙(2026-08-17 승격): ko가 원본이고 en/ja/zh는 파생이다. 단 §2.2의 세 유형대로 이력이 원칙을 수정하는 페이지는 표의 결정이 우선한다.
- 4로케일 동일 PR 묶음 유지, Mermaid TD-only, URL 블랙리스트(`adk.mo.ai.kr`만 유효), 본문 이모지 금지(`{{</* icon */>}}` 사용) — `hns-oss-docs-i18n-rules` 규칙을 그대로 적용한다.
- 증거 경로: `.moai/reports/t1328/` (canonical-generation verdict + 전/후 헤딩 테이블). plan phase에서는 이곳에 아무것도 만들지 않는다.

## 6. Out of Scope

### Out of Scope — 래칫 파일에 등록된 나머지 발산 페이지 (51페이지)

- 표적 3페이지 외에 `.locale-parity-baseline`에 이름 올라와 있는 페이지는 이번 카드에서 정렬하지 않는다. 후속 카드로 처리한다.

### Out of Scope — README 4파일 헤딩 패리티

- README.ko.md / README.md / README.ja.md / README.zh.md의 섹션 정렬은 docs-site 페이지와 별개 표면이며, 이 SPEC은 README 본문을 수정하지 않는다. verify 게이트가 README 패리티를 회귀 검사로 확인하는 것까지만 포함한다.

### Out of Scope — 내비게이션 재설계

- `hugo.toml`, `_meta.yaml`, `data/menu/main.yaml`의 구조 변경은 REQ-009의 조건부 라우팅(패리티 달성에 필요한 최소 수정)까지만 허용하며, 메뉴 구조 재설계·아이콘 체계 변경은 하지 않는다.

### Out of Scope — 패리티 검사 도구화

- 카운터를 스크립트·CI 워크플로로 도구화하지 않는다. 레시피는 hns-oss-docs-verify 인라인 형태를 유지한다. (도구화는 별도 카드.)

### Out of Scope — 게시(publish) 행위

- 커밋·push·PR 생성은 orchestrator/사용자 게이트 소관이며, run phase 전문가는 docs-site 트리 안의 파일 수정과 검증까지만 수행한다.
