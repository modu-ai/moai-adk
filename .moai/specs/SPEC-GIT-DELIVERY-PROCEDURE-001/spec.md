---
id: SPEC-GIT-DELIVERY-PROCEDURE-001
title: "배포 지침의 git 전달 절차 기계적 수리 — fetch 순서, 병합 방식 해석, auto-merge 옵트인 단일 기준과 플래그 의미"
version: "0.2.6"
status: implemented
created: 2026-09-10
updated: 2026-09-12
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".claude/agents/moai/manager-git.md, .claude/rules/moai/core/agent-common-protocol.md, .claude/skills/moai/workflows/sync/delivery.md, .claude/skills/moai/workflows/sync/doc-execution.md, .claude/skills/moai/SKILL.md, .claude/skills/moai/references/reference.md, .claude/skills/moai/workflows/sync/quality-gates-context.md, .claude/skills/moai/workflows/sync.md, .claude/commands/moai/sync.md (template sync.md.tmpl) (+ internal/template/templates mirrors, internal/template/templates/.codex/agents/moai/manager-git.toml, internal/template/templates/.agents/skills/moai-sync/SKILL.md)"
lifecycle: spec-anchored
tags: "git, manager-git, sync-check, fetch-ordering, merge-method, auto-merge, sync-flags, commands-emit, template-mirror, instruction-audit"
era: V3R6
tier: M
related_specs: [SPEC-MERGE-METHOD-CONFIG-001]
---

# SPEC-GIT-DELIVERY-PROCEDURE-001 — 배포 지침의 git 전달 절차 기계적 수리

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-10 | manager-spec | 최초 작성. 칸반 카드 t622(지침 감사 G1). 결함 3건(AC-01·AC-11·SX-R04)을 한 SPEC으로 묶었다. 재현 근거 `.moai/reports/t622/repro.md`(측정 트리 `c352330d3`), 인용 줄번호는 `b412f8a33` 에서 확인. |
| 0.1.1 | 2026-09-10 | manager-spec | **plan-audit 1회차**(`.moai/reports/t622/plan-audit.md`) **FAIL 0.67** 반영 — 전체 범위(AC-01 포함) 대상. |
| 0.1.2 | 2026-09-10 | manager-spec | **plan-audit 2회차**(`.moai/reports/t622/plan-audit-iter2.md`) **FAIL 0.75** 반영(결함 N1~N6) — 전체 범위 대상. 운영자 결정 OD-1 = 선택지 1, OD-2 = 선택지 B 반영. |
| 0.1.3 | 2026-09-10 | manager-spec | 운영자 결정 B1~B3 반영. REQ 23·AC 24로 Tier M 상한 초과, B4·B6 기록. 커밋 `87988e946`. 이 판에 대한 plan-audit은 실행되지 않았다. |
| 0.2.0 | 2026-09-11 | manager-spec | 운영자 결정 T1 = 분할(리드 경유). 기계적 수리 세 가지와 그 파일들의 부수 의무만 남기고 late-branch 재설계는 카드 t658, amend 적용 도우미 스텁은 카드 t659로 옮겼다(§G). 번호 유지·자리표시 방식. Frozen 비접촉 가드 REQ-GDP-024·AC-GDP-025 추가. plan-audit 회차는 이 판부터 다시 센다. 커밋 `24e20ec50`. |
| 0.2.1 | 2026-09-11 | manager-spec | **축소판 plan-audit 1회차**(`.moai/reports/t622/plan-audit-reduced-iter1.md`) **FAIL 0.71** 반영. 운영자 하위 결정(플래그 의미)과 리드 공통 항목 판정(모드별 승인 조건)으로 REQ-GDP-025·026, AC-GDP-026~029 추가. 플래그 표면 네 파일을 범위에 넣고 소비자 조사 결과를 차단 항목 X1~X4로 보고. N1·N2·N4·N5·N6 반영. 커밋 `caa601d7c`. |
| 0.2.2 | 2026-09-11 | manager-spec | 리드 범위 판정(X1~X4) 반영. X1(`workflows/sync.md` 사용법 줄), X2(슬래시 명령 `argument-hint`, 로컬 `.claude/commands/moai/sync.md:3`·템플릿 `sync.md.tmpl:3`), X3(`delivery.md:404` 다음 단계 선택지)를 범위에 넣고, AC-GDP-026·027 조각을 이 세 자리로 넓혔다. X2는 명령 원본이라 `make commands-emit`·`make commands-emit-check` 와 게시본 변화 여부(두 경우 모두 판정)를 AC-GDP-030으로 새로 두고 REQ-GDP-014를 넓혔다. X4(docs-site 네 로케일)는 범위 밖에 두고 후속 문서 카드와 그 입력을 §D에 기록했다. 사본 일치·중립성·Frozen·순서 점검 목록을 새 파일로 넓혔다. 판정 대상 REQ 12·AC 16. 앞선 작성 시도는 세션 한도로 멈췄고 부분 편집은 WIP 커밋 `ea09ca650` 으로 보존됐다. 이 판은 그 커밋에서 이어 완성했다 — 부분 편집의 조각·대조 값을 트리에서 다시 쟀고, AC-GDP-025 레지스트리 검출식의 `manager-git` 을 워크트리 가드가 받는 `manager-[g]it` 으로 고쳤고, 빈 집합으로 통과하지 못하게 하는 대조를 AC-GDP-015·016·025·030에 더했다. |
| 0.2.3 | 2026-09-11 | manager-spec | **축소판 plan-audit 2회차**(`.moai/reports/t622/plan-audit-reduced-iter2.md`) **FAIL 0.75** 반영. 운영자 결정(리드 경유)에 따라 결함 D1~D8을 검출식 문자열과 문구만 고쳐 좁게 수리했다. 요구사항·결정·범위는 그대로이고 판정 대상 REQ 12·AC 16도 그대로다. D1: AC-GDP-026에 `--merge` 존재 검사 (iii)과 방향 검사, 읽기 기록 추가. D2: AC-GDP-028 (a)가 전원 승인을 요구하고 (b)가 부분 승인 문구를 잡으며 읽기 기록 추가. D3: AC-GDP-029 (b)가 올바른 문구를 떨어뜨리지 않고 리뷰 조건을 잡음. D4: AC-GDP-027 (ii)에 효과 낱말 추가. D5: 범위 파일 수를 아홉 개로 다시 세고 영향 파일 19(경우 B 21)로 정정. D6: §E.2 HEAD 인용 갱신. D7: 변수를 판정 명령과 같은 호출에서 지정하도록 명시. D8: AC-GDP-016 기준 커밋에 템플릿 사본 포함. 이어지는 3회차 감사는 D1~D8과 그로 인한 퇴행만 보며 마지막 회차다. |
| 0.2.4 | 2026-09-11 | manager-spec | **축소판 plan-audit 3회차**(`.moai/reports/t622/plan-audit-reduced-iter3.md`)의 선택 발견 N1~N3 반영. 운영자 결정(리드 경유): 구현 착수는 N1~N3을 먼저 적용하는 조건으로 승인됐고 재감사는 없다. 요구사항·결정·범위와 판정 대상 REQ 12·AC 16은 그대로다. N1: 명령 관례를 값을 글자 그대로 넣는 형태로 바꾸고, git 명령과 `manager-git.md` 를 담은 명령에는 변수 지정 형태를 쓰지 않는다고 적었다(가드 거부 관측). N2: 빈 결과가 PASS인 판정 열여섯 곳과 AC-GDP-030 경우 선택 한 곳 앞에 `test -e` 존재 확인을 두고 뮤턴트를 §D.2에 기록했다. N3: AC-GDP-028·029 읽기 목록 선택자에 `--auto-merge` 줄을 더하고 뮤턴트를 기록했다. |
| 0.2.5 | 2026-09-11 | manager-spec | **develop 흡수 뒤 재고정.** 카드가 로컬 develop 을 두 번 흡수했다(`7ac8b8491` ← develop `ee99507fb`, `255f88eb0` ← develop `f1f034bb4`). develop 카드 t635 가 `agent-common-protocol.md` Pre-Spawn 절을 Lane A(순서)/Lane B(독립) 형태로 다시 썼고 카드 dr0911 이 템플릿에 미러했다(두 사본 바이트 동일). **리드 판단 (a)(운영자 경유, 최종)**: t635 의 형태가 REQ-GDP-002 를 충족하므로 이 카드는 그 파일을 편집·미러하지 않는다. 요구사항의 실질·범위·판정 대상 REQ 12·AC 16은 그대로이고 Kickoff 승인도 유지된다. 반영: (1) 기준 트리를 `255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089` 로 옮기고, 인용 줄번호를 템플릿 사본 기준으로 밝히며 로컬 줄이 다른 곳을 §A.1 표로 적었다. (2) 로컬·템플릿 사본 차이 기준선을 BASE 에서 다시 재 REQ-GDP-013·AC-GDP-013·plan §C 6단계에 반영했다(바이트 동일은 `agent-common-protocol.md`·게시본 둘, 의도된 차이는 여덟). (3) REQ-GDP-002 를 착지한 성질(fetch 완료와 exit status 관측 뒤 rev-list)로 고쳤다 — t635·dr0911 이 충족, 이 카드는 편집하지 않는다. [0.2.6 정정: 이 문구는 exit status 관측을 새로 요구해 0.2.4 보다 좁았다. 이 항목에 한해 앞 문장의 "요구사항의 실질 … 그대로" 는 맞지 않았고, 0.2.6 에서 원래 실질로 되돌렸다.] (4) AC-GDP-002 를 Lane A/B 형태 판정으로 바꾸고 회귀 방지 판정으로 표시했다(뮤턴트 4개 FAIL·현재 블록 PASS 재현). (5) M5 를 편집 없는 확인 단계로 바꾸고 AC-GDP-016 을 "이 카드 커밋이 `agent-common-protocol.md` 두 사본을 바꾸지 않음" 판정으로 바꿨다(양성 대조 `b412f8a33..BASE` 2커밋). (6) M4·M6·AC-GDP-013 에 미러 테스트 비회귀 점검을 더했다(BASE 기준선 exit 1, PASS 17줄, FAIL {`TestRuleTemplateMirrorDrift`, `TestRuleTemplateMirrorDrift/spec-workflow.md`}). (7) 흡수로 밀린 로컬 줄 인용을 plan·acceptance 에서 고쳤다. 증거 `.moai/reports/t622/reanchor/`. |
| 0.2.6 | 2026-09-11 | manager-spec | **재고정 변경분 감사**(`.moai/reports/t622/plan-audit-reanchor.md`) **FAIL 0.75** 의 D1~D6 반영. 요구사항 수·범위·판정 대상 REQ 12·AC 16은 그대로다. **D1**: "이 카드가 바꾼 것" 판정의 왼쪽 끝을 리터럴 BASE 에서 읽는 시점의 merge-base 로 바꿨다 — 차이는 `git diff develop...HEAD -- <경로>`, 커밋 목록은 `git log --no-merges --format=%H HEAD --not develop -- <경로>`, 대조 `git diff --name-only develop...HEAD` 1줄 이상(0줄이면 "측정 불가"). 대상 AC-GDP-014·015·016·030 과 같은 형태였던 AC-GDP-025, plan §C 1단계·M5. 병합 전 전용임을 적었다. BASE 는 고정 스냅숏(기준 사본 반출·양성 대조·사본 덩어리·미러 기준선·Pre-Spawn/Pre-Edit 기준 절)에만 남기고, 판정 전 스냅숏 신선도 점검을 더했다. "재흡수 없음" 전제와 "BASE 를 흡수 병합으로 옮김" 대응을 지웠다(§E.2). **D2**: AC-GDP-002 를 (a) 순서 판정과 (b) Pre-Spawn 절 전체 보존 diff 로 나눴다. (a)는 fetch 줄을 줄 끝에 고정하고(백그라운드 `&`·파이프는 실패, `;`·`&&` 로 이은 한 줄은 순서 보장으로 받음) `exit` 판별을 `exit([[:space:]]|$)` 로 고쳤다. 뮤턴트 i~viii 모두 FAIL, 현재 두 블록 PASS. **D3**: REQ-GDP-002 를 원래 실질(fetch 가 끝난 뒤 rev-list 시작, 독립 병렬 항목 나열 금지)로 되돌리고 착지 블록의 상태 관측·실패 시 멈춤은 비규범 주석으로 옮겼다. §A.2 의 Pre-Edit 한 줄 형태 "정상" 은 이 문구와 맞아 유지했다. **D4**: AC-GDP-002 MUST-PASS 근거를 `verification-completeness.md` §4 보존 단언으로 적었다. **D5**: plan §C 8단계 집합 비교에 정렬을 명시했다. **D6**: AC-GDP-016 GIVEN 을 "이 카드의 커밋들(병합 제외)" 로 고쳤다. 증거 `.moai/reports/t622/reanchor-fix/`. |

---

## §A 배경과 문제

### A.1 결함 요약

| 결함 | 내용 | 처리 |
|---|---|---|
| AC-11 | `git fetch` 와 그 결과를 읽는 `git rev-list` 를 서로 독립인 병렬 배치로 지시한다 | `manager-git.md` 는 기계적 수리 (이 SPEC). `agent-common-protocol.md` Pre-Spawn 절은 develop 카드 t635·dr0911 이 이미 고쳐 이 SPEC은 편집하지 않고 보존만 판정한다(0.2.5, 리드 판단 (a)) |
| SX-R04 (병합 방식) | `delivery.md` 와 `manager-git.md` 예시가 `gh pr merge --squash` 를 고정한다 | 기계적 수리 (이 SPEC) |
| SX-R04 (기본값) | 워크트리 문맥의 auto-merge 기본값이 `delivery.md`·`doc-execution.md` 와 `manager-git.md` 에서 반대로 적혀 있다 | OD-2 = 선택지 B (이 SPEC) |
| SX-R04 (플래그 의미) | `manager-git.md` 의 옵트인 플래그 `--auto-merge` 가 `/moai sync` 표면에 없고, 표면은 `--merge`(폐기 표시가 곳에 따라 다름)와 `--no-merge` 만 서술하며, 비-team 모드의 병합 조건이 없다 | OD-2 하위 결정, 리드 공통 항목 판정, 리드 범위 판정 (이 SPEC) |
| AC-01 | 배포 지침이 스스로 금지한 primary checkout 브랜치 변경을 실행하라고 안내한다 | 카드 t658로 이동 (§G) |

**기준 트리 (0.2.5 재고정).** `BASE=255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089` — 카드가 로컬 develop `f1f034bb4` 를 두 번째로 흡수한 병합 커밋이다. 이 카드는 아직 범위 파일을 고치지 않았으므로 이 커밋의 범위 파일이 run-phase 이전 상태다. 0.2.4 까지의 기준 `b412f8a33` 은 HISTORY 에만 남긴다. **BASE 는 고정 스냅숏이다(0.2.6).** 기준 사본 반출, 양성 대조, 사본 덩어리 기준선, 미러 테스트 기준선, AC-GDP-002·003 의 기준 절처럼 움직이면 안 되는 값에만 쓴다. "이 카드가 무엇을 바꿨는가·커밋했는가" 를 재는 범위 판정의 왼쪽 끝은 BASE 가 아니라 읽는 시점의 `git merge-base develop HEAD` 다 — 통합 창에서 로컬 develop 을 다시 흡수하는 일은 예정된 단계이기 때문이다(§E.2). 측정(증거 `.moai/reports/t622/reanchor/`):

- 0.2.5 작성 시점 HEAD `b24f2e184` 와 BASE 의 차이는 보고서 두 파일뿐이다: `git diff --stat 255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089 HEAD` → `.moai/reports/t622/absorb-ee99507fb.md`, `.moai/reports/t622/absorb2-mirror-baseline.txt`, 2 files changed (`base-vs-head-stat.txt`).
- 범위 루트 작업 트리 = BASE: `git diff --stat 255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089 -- internal/template/templates/ .claude/ .agents/ Makefile docs-site/content` → 빈 출력(`test -e` exit 0, `test -s` exit 1). 같은 형태를 `b412f8a33..255f88eb0` 에 돌리면 "113 files changed" — 양성 대조(`control-b412-to-base-scope-stat.txt`).
- `b412f8a33..BASE` 사이에 바뀐 보조 파일은 `Makefile`, `plan-auditor.md`, docs-site `moai-web-console.md` 네 로케일뿐이고, 이 문서가 인용하는 `Makefile:38-60` 네 타깃 줄과 `plan-auditor.md:138` MP-1 문장은 BASE 에서 같은 줄이다. 사본·발행 테스트 파일 여덟 개, `commandemit`·`agentemit`, `zone-registry.md`(로컬·템플릿), 생성물 `manager-git.toml`, 게시본(로컬·템플릿), `defaults.go`, `shipped_key_inventory.yaml`, `git-strategy.yaml.tmpl`, X4 docs-site 파일은 바뀌지 않았다(`support-files-b412-to-base-stat.txt`).

**인용 줄번호는 템플릿 사본 기준이다.** 템플릿 사본의 인용 줄은 `b412f8a33` 과 BASE 에서 같다. develop 이 로컬 사본에만 넣은 편집으로 일부 파일의 로컬 줄이 밀렸고, 다른 곳은 아래 표다(BASE 에서 두 사본을 `/usr/bin/grep -n` 으로 다시 잼, `lines-*.txt`). 이 문서의 나머지 줄번호도 이 표로 로컬 줄로 바꿔 읽는다. 판정 명령은 줄번호가 아니라 표지(제목·고정 문구)로 위치를 찾는다.

| 파일 | 템플릿 줄 (인용 값) | 로컬 줄 | 이유 |
|---|---|---|---|
| `manager-git.md` | 32 · 42 · 82 · 114 · 133-140(Personal Mode) · 142·144·148(Team Mode) · 154(`## Synchronization`) · 156 · 160 · 164-171(`## PR Auto-Merge`, 166 옵트인 문장) | 34 · 44 · 84 · 116 · 135-142 · 144·146·150 · 156 · 158 · 162 · 166-173(168) | 로컬 프론트매터 설명 +2줄(덩어리 `5,7c5`) → 6행부터 +2 |
| `delivery.md` | 330(`#### Step 3.4`) · 335-338 · 343 · 348-349 · 355 · 356 · 396(`#### Context-Aware Next Steps`) · 404 | 355 · 360-363 · 368 · 373-374 · 380 · 381 · 421 · 429 | 템플릿 279~421행 구간은 로컬 +25 (덩어리 `300c275`·`303c278` 과 `447,448c422` 사이) |
| `workflows/sync.md` | 85(사용법) · 104(`**Flags**:`) | 95 · 114 | 로컬 전용 65-74행 이후 +10 (0.2.4 부터 적은 값) |
| `agent-common-protocol.md` | 두 사본 바이트 동일: 13 · 17 · 52 · 290(`### Pre-Spawn Sync Check`) · 292(태그) · 296-297(Lane A/B 문장) · 299-312(bash 블록: fetch 301, `fetch_status=$?` 302, 검사 303-306, rev-list 307, session list 311) · 341(`### Pre-Edit Sync Check`) · 360 · 375(`#### The sweep prohibition`) | 같음 | t635 가 Pre-Spawn 절을 다시 써 `b412f8a33` 의 290-305 · 328 · 347 · 362 가 이렇게 바뀌었다 |
| 나머지 — `doc-execution.md` 29·34-36, `moai/SKILL.md` 140, `references/reference.md` 161, `quality-gates-context.md` 30·98·101, 명령 원본 3, `manager-git.toml` 26·108·142·148·150·158·160 | 인용 값 | 같음 | 밀리지 않음 |

### A.2 AC-11 — 결과를 읽는 명령과의 병렬 배치

| 파일 | 줄 | 문제 |
|---|---|---|
| `.claude/agents/moai/manager-git.md` (생성물 `manager-git.toml:150`) | 156 (로컬 158) | `git fetch`, `git status`, `git rev-list --count --left-right`, `gh pr checks --json` 를 "independent and read-only" 로 묶어 한 턴 병렬 배치로 지시 |
| `.claude/rules/moai/core/agent-common-protocol.md` (`b412f8a33`) | 292, 296, 299 | Pre-Spawn Sync Check 가 "parallel batch" 를 선언하고 `git fetch origin main`(296)과 `git rev-list`(299)를 코드 블록 안 별도 명령으로 나열 |
| 같은 파일 (BASE — 해소됨) | 296-297, 299-312 | develop 카드 t635 가 Pre-Spawn 절을 Lane A(순서)/Lane B(독립)로 다시 썼고 카드 dr0911 이 템플릿에 미러했다(두 사본 `cmp` exit 0). Lane A 블록은 `git fetch origin main 2>&1`(301) → `fetch_status=$?`(302) → 0이 아니면 `exit`(303-306) → `git rev-list --count --left-right origin/main...HEAD`(307) 순서이고, Lane B `moai session list --json --filter-spec=<SPEC-ID>`(311)는 fetch 와 함께 돌 수 있다. 이 SPEC은 이 파일을 편집하지 않는다(REQ-GDP-002, AC-GDP-002 회귀 방지) |
| 같은 파일 (정상) | 360 (`b412f8a33` 347) | Pre-Edit Sync Check 는 `git fetch origin main 2>&1; git rev-list ...` 를 한 명령으로 묶어 fetch 가 끝난 뒤 rev-list 가 시작됨 — 유지 대상. 0.2.6 에서도 "정상" 이 맞다: 이 모양은 REQ-GDP-002 가 요구하는 순서 성질을 충족하고, AC-GDP-002 (a) 도 `;`·`&&` 로 이은 한 줄을 순서 보장으로 받는다. 다만 Pre-Edit 절은 REQ-GDP-002 의 대상이 아니다 — REQ-GDP-003 으로 보존만 판정한다. t635 가 쓴 Pre-Spawn 절 문장이 "완료를 구분할 수 없는 한 줄" 을 피하라고 적는 것은 착지 블록의 추가 성질이며, AC-GDP-002 (b) 가 Pre-Spawn 절 보존으로 지킨다 |

`git fetch` 는 원격 추적 ref 를 갱신하고 `git rev-list` 는 그 ref 를 읽으므로 둘은 독립이 아니다. 병렬로 돌면 원격이 앞선 상태(`N 0`)가 `0 0` 으로 보고될 수 있다.

### A.3 SX-R04 — 병합 방식 고정, auto-merge 기본값 충돌, 플래그 의미 공백

줄은 템플릿 사본 기준이고, 로컬 줄이 다르면 괄호에 적는다(§A.1 표).

| 파일 | 줄 | 내용 |
|---|---|---|
| `delivery.md` | 335-338 (로컬 360-363) | 트리거: `is_worktree_context == true` 이고 `--no-merge` 가 없으면 병합(기본 병합), 또는 `--merge` 명시(deprecated) |
| `delivery.md` | 343, 355 (로컬 368, 380) | `gh pr merge --squash --delete-branch` 고정 |
| `delivery.md` | 348-349 (로컬 373-374) | "`--no-merge`: Skip auto-merge even in worktree context.", "`--merge`: Deprecated. … Auto-merge is now the default for worktree contexts." |
| `delivery.md` | 404 (로컬 429) | Phase 14 다음 단계 선택지 "Auto-Merge PR (/moai sync --merge)" (X3) |
| `sync/doc-execution.md` | 34-36 | "This affects auto-merge behavior: worktree contexts default to auto-merge." |
| `moai/SKILL.md` | 140 | "Modes: auto, force, status, project. Flags: --merge, --skip-mx" — `--merge` 를 폐기 표시 없이 나열 |
| `moai/references/reference.md` | 161 | "- --merge: Auto-merge PR and clean up branch after sync" |
| `sync/quality-gates-context.md` | 30, 101 | "  - Flag: --merge", "- --merge: After sync, auto-merge PR and clean up branch. …" |
| `workflows/sync.md` | 로컬 95 / 템플릿 85 | 사용법 줄 `/moai sync [mode] [--pr] [--merge] [--skip-mx]` (X1) |
| `workflows/sync.md` | 로컬 114 / 템플릿 104 | "**Flags**: `--pr` (PR 생성) \| `--merge` (deprecated, auto-merge) \| `--skip-mx` (MX 검증 스킵)" |
| `.claude/commands/moai/sync.md` (로컬) · `sync.md.tmpl` (템플릿) | 3 | 슬래시 명령 힌트 `argument-hint: "[SPEC-XXX] [--merge] [--skip-mx]"` (X2) |
| `manager-git.md` (생성물 `.toml:108`) | 114 (로컬 116) | 예시 `gh pr merge <PR> --squash --delete-branch   # squash default` |
| `manager-git.md` (생성물 `.toml:26`) | 32 (로컬 34) | `merge_method` 해석 규칙의 기본값 설명 — **유지 대상** |
| `manager-git.md` (생성물 `.toml:142`) | 148 (로컬 150) | Team Mode "Auto-merge: only with the `--auto-merge` flag, per § PR Auto-Merge (Team Mode)" — 유지 |
| `manager-git.md` (생성물 `.toml:158-166`) | 164-171 (로컬 166-173) | "## PR Auto-Merge (Team Mode)" 절: "Execute only with `--auto-merge` flag AND all approvals obtained:" 와 다섯 단계. 절 제목이 team 한정이고, "### Personal Mode"(133-140, 로컬 135-142)에는 auto-merge 규칙이 없다 |

team 모드이면서 워크트리 문맥이면 두 지침이 반대를 지시한다. 명령 표면은 `--merge` 를 auto-merge 플래그로 소개하며 폐기 여부가 서로 다르고, `--auto-merge` 는 어느 표면에도 없다. 병합 방식 치환은 SPEC-MERGE-METHOD-CONFIG-001 REQ-MMC-007/008이 이미 요구한 것이 반영되지 않은 상태다. 워크트리 기본 병합 문구는 커밋 `b7447cb90`(`sync.md`)에서 들어와 `980ccdc56` 이 `delivery.md` 로 옮겼다(`git log -S`, `delivery.md:349` 문장).

### A.4 설정·사본·테스트·생성물 사실 (측정)

| 항목 | 측정 결과 |
|---|---|
| `merge_method` | 템플릿 `git-strategy.yaml.tmpl` 세 모드 모두 `squash` · Go 기본값 `internal/config/defaults.go` 738·751·767행 `MergeMethod: "squash"` |
| 폐기 키 `workflow.worktree.auto_merge` | `internal/config/testdata/shipped_key_inventory.yaml:2940-2943` 에 `class: D`, `evidence: none`, `deprecate_after: "v3.1.0"`. 이 SPEC은 건드리지 않는다 |
| 로컬·템플릿 사본 (BASE, `diff <로컬> <템플릿>` 덩어리 머리, `.moai/reports/t622/reanchor/pair-hunks.txt`) | diff exit 0(바이트 동일): `agent-common-protocol.md`, 게시본 `.agents/skills/moai-sync/SKILL.md`. diff exit 1(의도된 차이 — develop 이 로컬에만 넣은 편집 포함): `manager-git.md` — `5,7c5`(로컬 프론트매터 설명 +2줄); `delivery.md` — `9,11c9,11`, `49,54c49,50`, `154,163c150,152`, `169,188c158,163`, `300c275`, `303c278`, `447,448c422`, `450,458c424,427`, `510,511c479`; `doc-execution.md` — `9,11c9,11`, `79,91d78`, `118c105`, `128,134c115,116`, `144c126`, `156,161d137`, `173,174c149`, `180c155`, `182,189d156`; `quality-gates-context.md` — `9,11c9,11`, `48,49c48,49`, `51c51`, `138c138`, `142,149c142`, `151c144`, `165c158`; `moai/SKILL.md` — 20개 덩어리(`125c125` … `273,274c273,274`, `392d391`, `b412f8a33` 과 같음), 140행은 덩어리 밖; `references/reference.md` — `229d228`; `workflows/sync.md` — `29,31c29,31`, `65,74d64`, `81c71`(로컬 사본에 10줄이 더 있어 75행부터 번호가 10 밀림); 명령 원본 `.claude/commands/moai/sync.md` 대 `sync.md.tmpl` — `2c2`(2행 `description` 만 다름: 템플릿은 로케일별 Go 템플릿 조건문, 로컬은 영어 문장). 3행 `argument-hint` 는 두 사본에서 같다. 편집 자리(§A.5)는 모두 이 덩어리 밖이다 |
| 바이트 동일 미러 테스트 | `rule_template_mirror_test.go` 의 `workflowOptMirroredPaths`·`lateBranchMirroredPaths` 는 범위 파일 아홉 개 중 어느 것도 담지 않는다. `agent-common-protocol.md`·`manager-git.md`·`moai/SKILL.md` 는 주석으로 "byte-parity 허용 목록에서 제거" 가 명시돼 있다 |
| 기타 사본·청결 가드 | `sanitized_pair_parity_test.go:71` 이 `agent-common-protocol.md` 를 담는다(`TestSanitizedPairParity`). `internal_content_leak_test.go:1535` `TestTemplateNoInternalContentLeak` 는 템플릿 루트 전체를 걷는다. 나머지 여덟 파일의 로컬·템플릿 사본 일치를 지키는 테스트는 없다 — diff만이 가드다. 새 범위 파일을 이름으로 가리키는 테스트는 사본 일치가 아닌 다른 성질을 본다: `backlog_json_disclosure_mirror_test.go:24`(`moai/SKILL.md` 임베드 사본 = 템플릿 원본), `template_neutrality_audit_test.go:141`(`moai/SKILL.md` C2 허용 목록), `agent_frontmatter_audit_test.go:399`·`:407`(프론트매터), `agentless_audit_test.go:44`(`sync.md` 를 구현 스킬 목록에 둠), `commandemit/golden_test.go`·`published_skills_deploy_test.go`(게시본과 명령 원본의 발행 관계) |
| 사본 테스트 기준선 (BASE) | `go test ./internal/template/ -count=1 -run 'TestSanitizedPairParity\|TestRuleTemplateMirrorDrift' -v` (표 안이라 파이프 앞에 역슬래시를 붙였다 — 실행할 때는 뺀다) → exit 1. PASS 줄 17(하위 테스트 16 + 최상위 `TestSanitizedPairParity`), FAIL 집합 = {`TestRuleTemplateMirrorDrift`(최상위), `TestRuleTemplateMirrorDrift/spec-workflow.md`}. `spec-workflow.md` 는 범위 밖이고 다른 카드가 고친다. 원본 `.moai/reports/t622/absorb2-mirror-baseline.txt`, 0.2.5 재실행 `.moai/reports/t622/reanchor/mirror-baseline-rerun.txt`(PASS/FAIL 집합이 원본과 `diff` exit 0), 집합 파일 `reanchor/mirror-baseline-sets.txt`. `TestHookWrapperCopiesStayIdentical`(`internal/hook/wrapper_copies_contract_test.go:73`)은 `.claude/hooks/moai` 의 훅 래퍼 스크립트 여섯 개와 그 템플릿 사본만 읽어 이 카드 파일이 닿지 않으므로 넣지 않는다 |
| 생성물 (에이전트) | 템플릿 `manager-git.md` 가 바뀌면 `manager-git.toml` 을 `make agents-emit` 으로 재생성한다. 기준 트리에서 `.toml` 의 `## Synchronization` 절(11줄)과 `## PR Auto-Merge` 절(9줄)은 템플릿 `manager-git.md` 의 같은 절과 diff exit 0 |
| 생성물 (명령 게시본) | `make commands-emit` 은 `COMMAND_EMIT_UPDATE=1 go test ./internal/template/commandemit/... -run TestGoldenCommittedArtifactsMatchEmission`, `make commands-emit-check` 는 같은 테스트를 읽기 전용으로 실행(`Makefile:51-52`·`:58-60`). 발행기는 템플릿 트리(`internal/template/templates`, `golden_test.go:28` `templatesDir`)의 명령 원본(`commandemit.go:49-53` `DefaultOptions` 의 `.claude/commands/moai` → `.agents/skills`)에서 게시본을 만들고, 갱신 분기(`golden_test.go:57-74`)는 템플릿 트리 게시본만 쓴다. 게시본 프론트매터는 생성 머리말·`name`·영어 `description` 뿐이고 본문은 원본 본문을 그대로 옮긴다(`emit.go:122-130` `renderSkill`). `TestCommandSourcesUnmodified`(`golden_test.go:94-106`)는 한 실행 안에서 발행 전후 원본 해시를 비교하므로 원본 편집만으로는 실패하지 않는다. 게시본 두 사본은 각각 7줄이고, 명령 원본 두 사본과 함께 추적 파일이다(`git ls-files` 네 경로 모두 출력). `argument-hint` 와 `allowed-tools` 는 "Claude-only keys and are NOT carried into the published skill" (`loader.go:4-6`). 로컬 게시본 `.agents/skills/moai-sync/SKILL.md` 도 추적 파일이며 템플릿 게시본과 바이트 동일하다. 두 게시본에는 `merge` 가 없다(`grep -i merge` exit 1) |

### A.5 Frozen 비접촉 확인 (0.2.2 작성 시점, 템플릿·로컬 사본 — 0.2.5 에서 BASE 에 다시 잼)

0.2.5 재측정: `[ZONE:` 줄 수 — 나머지 여덟 파일 로컬·템플릿 모두 0(`reanchor/zone-lines-others.txt`, `manager-git.md` 두 사본 0), `agent-common-protocol.md` 는 `[ZONE:Frozen]` 17행 1줄·`[ZONE:Evolvable]` 16줄(두 사본), 레지스트리 항목 — `agent-common-protocol.md` 13·13, 나머지 여덟 파일 0·0(`reanchor/registry-others.txt`). 아래 표의 값과 같다.

| 확인 | 명령과 결과 | 판정 |
|---|---|---|
| 범위 파일의 `[ZONE:]` 태그 | `manager-git.md`·`delivery.md`·`doc-execution.md`·`moai/SKILL.md`·`references/reference.md`·`quality-gates-context.md`·`workflows/sync.md`·명령 원본(`sync.md`, `sync.md.tmpl`) 0줄(로컬·템플릿). `agent-common-protocol.md` 는 `[ZONE:Frozen]` 1줄(17행, User Interaction Boundary)과 `[ZONE:Evolvable]` 16줄 | Pre-Spawn Sync Check 태그는 292행 `[ZONE:Evolvable]` |
| 레지스트리 등록 절 | `zone-registry.md`(로컬·템플릿)에서 `file:` 이 위 여덟 파일인 항목 0개. `file:` 이 `agent-common-protocol.md` 인 항목 13개(양성 대조). 그중 Frozen은 `CONST-V3R2-006`(13행), `CONST-V3R2-036`·`038`(17행), `CONST-V3R2-037`(52행) — 모두 `#user-interaction-boundary` | 등록 절은 11~110행 구간에만 있다 |
| 유지 편집 자리 | `manager-git.md` 114·156·164-171(와 필요하면 133-148) — 로컬 116·158·166-173(135-150), `delivery.md` 335-338·343·348-349·355·404 — 로컬 360-363·368·373-374·380·429, `doc-execution.md` 34-36, `moai/SKILL.md` 140, `references/reference.md` 161, `quality-gates-context.md` 30·101, `workflows/sync.md` 사용법 줄과 플래그 줄, 명령 원본 3행. `agent-common-protocol.md` 는 0.2.5부터 편집 자리가 없다 | 어느 자리도 `[ZONE:Frozen]` 블록이나 등록 절 안에 있지 않다 |

### A.6 선행 SPEC 관계

| SPEC | 상태 | 관계 |
|---|---|---|
| SPEC-MERGE-METHOD-CONFIG-001 | completed | `merge_method` 도입. REQ-MMC-007/008이 `delivery.md`·`manager-git.md` 의 `--squash` 고정 제거를, REQ-MMC-009가 기본값에서 명령 바이트 동일을 요구 |

### A.7 `/moai sync` 명령 표면 소비자 조사와 범위 판정

조사 명령(0.2.1): `/usr/bin/grep -rn -e '--merge' -e '--no-merge' -e '--auto-merge' .claude internal/template/templates/.claude internal/template/templates/.agents docs-site` → 167줄. 줄마다 뜻을 읽어 나눴고, 0.2.2에서 리드 범위 판정을 반영했다.

| 갈래 | 위치 | 판정 |
|---|---|---|
| 범위 (0.2.1 결정 목록) | `moai/SKILL.md:140`, `references/reference.md:161`, `quality-gates-context.md:30`·`:101`, `workflows/sync.md` 플래그 줄(로컬 114 / 템플릿 104), `delivery.md` 337-338·348-349, `manager-git.md` 148·166 (모두 로컬·템플릿; 줄은 템플릿 사본, 로컬 줄은 §A.1 표) | REQ-GDP-006·025·026 |
| 범위 (0.2.2 리드 범위 판정) | X1 `workflows/sync.md` 사용법 줄(로컬 95 / 템플릿 85), X2 `.claude/commands/moai/sync.md:3`·템플릿 `sync.md.tmpl:3` `argument-hint`, X3 `delivery.md:404` | REQ-GDP-025, X2는 REQ-GDP-014(명령 게시본)도 |
| 범위 밖 — 후속 문서 카드 | X4 docs-site 네 로케일(입력은 §D) | §D |
| 다른 뜻 — 범위 밖 | `gh pr merge --merge`(병합 커밋 방식): `.claude/agents/harness/hns-release-specialist.md` 74·232·371·379(로컬 전용), `moai-ref-git-workflow/SKILL.md:134`(로컬·템플릿), docs-site `worktree/examples.md` 네 로케일. `--merged-only`·`git branch --merged`(워크트리 정리·브랜치 가드 문서, docs-site worktree 문서). `docs-site/README.md:348` `moai update --merge`(다른 명령) | §D |

---

## §B 요구사항 (GEARS)

요구사항 식별자는 `REQ-GDP-NNN`. **번호 방식**: 0.1.3의 번호를 그대로 유지한다. 카드 t658로 옮긴 번호와 이전에 철회된 번호는 원래 자리에 한 줄 자리표시로 남기고, 새 요구사항은 다음 번호를 받는다. 이유는 §C.2.

### B.1 AC-11 — fetch 순서 보장

**REQ-GDP-001** (Ubiquitous)
`manager-git` 에이전트 정의의 동기화 절은 `git fetch` 가 끝난 뒤에 그 결과를 읽는 `git rev-list --count --left-right` 가 실행되도록 순서를 지시해야 하며, `git fetch` 를 그 결과를 읽는 명령과 서로 독립인 병렬 배치 항목으로 분류해서는 안 된다. fetch 결과를 읽지 않는 명령(`git status`, `gh pr checks --json`)을 병렬로 두는 일은 이 요구사항이 제한하지 않는다.

**REQ-GDP-002** (Ubiquitous)
`agent-common-protocol.md` 의 Pre-Spawn Sync Check 코드 블록은 `git fetch origin main` 이 끝난 뒤에 `git rev-list --count --left-right origin/main...HEAD` 가 시작되도록 지시해야 하며, 두 명령을 서로 독립인 병렬 항목으로 나열해서는 안 된다. 세 번째 명령(`moai session list --json --filter-spec=<SPEC-ID>`)의 병렬 실행 허용과 두 해석 표의 내용은 바뀌어서는 안 된다.

> 0.2.5: 이 요구사항은 BASE 에서 이미 충족돼 있다 — develop 카드 t635 가 Pre-Spawn 절을 Lane A(순서: fetch → `fetch_status=$?` → 0이 아니면 `exit` → rev-list)/Lane B(독립: session list)로 다시 썼고, 카드 dr0911 이 템플릿 사본에 미러했다(두 사본 바이트 동일). 리드 판단 (a)에 따라 이 카드는 이 파일을 편집하지 않으며, AC-GDP-002 는 이 성질이 이 카드의 작업 동안 유지되는지 보는 회귀 방지 판정이다.
>
> 0.2.6 (비규범): 요구사항 문구는 0.2.4 의 실질 — fetch 가 끝난 뒤에 rev-list 가 시작되고, 둘을 독립 병렬 항목으로 나열하지 않는다 — 로 되돌렸다. 0.2.4 의 "순서가 보장되는 한 명령" 은 그 실질을 한 가지 모양으로 적은 것이었고, 0.2.5 의 "exit status 를 관측한 뒤" 는 실질을 좁힌 것이었다(재고정 감사 D3). 착지한 블록은 여기에 더해 fetch 의 exit status 를 `fetch_status=$?` 로 관측하고 0이 아니면 `exit` 로 멈춘다. 이 추가 성질은 REQ-GDP-002 가 요구하지 않으며, AC-GDP-002 (b) 가 Pre-Spawn 절 전체 보존으로 지킨다.

**REQ-GDP-003** (Unwanted)
`agent-common-protocol.md` 의 Pre-Edit Sync Check 절(`### Pre-Edit Sync Check` 제목 — BASE 341행, `b412f8a33` 328행 — 부터 `#### The sweep prohibition` 소절 앞까지)은 이 SPEC의 변경으로 수정되어서는 안 된다.

### B.2 SX-R04 — 병합 방식 해석

**REQ-GDP-004** (Ubiquitous)
`sync/delivery.md` 의 auto-merge 실행 단계는 병합 명령을 활성 모드의 `git_strategy.<mode>.merge_method`(기본 `squash`)로 해석한 `gh pr merge --<merge_method> --delete-branch` 로 지시해야 한다. 기본값에서 실제 실행 명령은 지금과 바이트 동일해야 한다(SPEC-MERGE-METHOD-CONFIG-001 REQ-MMC-009 연속성).

**REQ-GDP-005** (Unwanted)
범위 파일 아홉 개(로컬·템플릿)와 생성물 `manager-git.toml` 의 어떤 실행 예시도 `--squash` 를 고정한 `gh pr merge` 명령을 실행할 명령으로 제시해서는 안 된다. `manager-git.md:114`(로컬 116) 예시는 `--<merge_method>` 로 적혀야 하고, `manager-git.md:32`(로컬 34) 의 기본값 설명 문장은 남아야 한다.

### B.3 SX-R04 — auto-merge 옵트인 단일 기준 (OD-2 = B)

**REQ-GDP-006** (Event-driven)
When sync 단계가 PR 병합 여부를 판단할 때, `delivery.md` 의 Auto-Merge 절(335-338 트리거, 348-349 플래그 설명 — 로컬 360-363, 373-374)과 `doc-execution.md` 의 워크트리 문맥 감지 절(34-36)은 `manager-git.md` 의 옵트인(`--auto-merge` 플래그와 REQ-GDP-026의 모드별 승인 조건)을 유일한 기준으로 따라야 하고, 워크트리 문맥이라는 이유만으로 병합하는 기본값을 적어서는 안 되며, 기준이 `manager-git.md` 임을 이름으로 밝혀야 한다. `manager-git.md` 의 옵트인 문장(148행, 166행 — 로컬 150행, 168행)의 `--auto-merge` 조건은 남아야 한다.

### B.4 옮긴 번호와 철회된 번호 (자리표시)

아래 번호는 이 SPEC에서 요구사항으로 판정하지 않는다. 원문은 커밋 `87988e946` 의 spec.md에 있고, 추적 표는 §G.

**REQ-GDP-007** — **[MOVED → 카드 t658]** AC-01 명령 형태 실행 안내 불변식.
**REQ-GDP-008** — **[MOVED → 카드 t658]** late-branch 대체 절차 없는 예시 삭제 금지와 절차 참조 유지.
**REQ-GDP-009** — **[MOVED → 카드 t658]** 네 지침 절의 흐름 일치와 `feat/SPEC-` 단일 접두.
**REQ-GDP-010** — **[MOVED → 카드 t658]** 런처 워크트리 흐름, `manager-git.md` 42·160, Frozen 줄 수정 기록.
**REQ-GDP-011** — **[WITHDRAWN 2026-09-10]** OD-1 선택지 2(조건부 유지) 경로. 기록은 §G.
**REQ-GDP-012** — **[WITHDRAWN 2026-09-10]** OD-1 선택지 3(`main_late_branch` 은퇴) 경로. 기록은 §G.

### B.5 사본·생성물·중립성

**REQ-GDP-013** (Ubiquitous)
이 SPEC이 바꾼 줄은 로컬 사본(`.claude/…`)과 템플릿 사본(`internal/template/templates/.claude/…`)에서 같아야 한다. 두 사본 사이의 의도된 차이 — BASE 에서 잰 `diff <로컬> <템플릿>` 덩어리(§A.4): `manager-git.md` 1개(`5,7c5`), `delivery.md` 9개, `doc-execution.md` 9개, `quality-gates-context.md` 7개, `moai/SKILL.md` 20개, `references/reference.md` 1개(`229d228`), `workflows/sync.md` 3개(`29,31c29,31`, 로컬 전용 65-74행, 81행), 명령 원본 `sync.md`/`sync.md.tmpl` 1개(2행 `description`) — 는 그대로 남아야 하고, 바이트 동일한 두 쌍(`agent-common-protocol.md`, 게시본 `moai-sync/SKILL.md`)은 바이트 동일하게 남아야 한다.

**REQ-GDP-014** (Ubiquitous)
템플릿 `manager-git.md` 가 바뀌면 `internal/template/templates/.codex/agents/moai/manager-git.toml` 은 `make agents-emit` 으로 재생성되어 같은 카드에 커밋되어야 하며, 손으로 편집해서는 안 되고, `make agents-emit-check` 가 exit 0 이어야 한다. 템플릿 명령 원본 `sync.md.tmpl` 이 바뀌면 `make commands-emit` 을 실행하고 `make commands-emit-check` 가 exit 0 이어야 하며, 그로 인해 바뀐 게시본이 있으면 명령 원본을 바꾼 커밋과 같은 커밋에 들어가야 하고, 로컬 게시본 사본은 템플릿 게시본과 같아야 한다.

**REQ-GDP-015** (Unwanted)
`internal/template/templates/` 아래에 추가되는 문구는 SPEC ID, REQ 토큰, 내부 날짜, 커밋 SHA, `CLAUDE.local` 참조, 16개 지원 프로그래밍 언어 중 특정 언어로 치우친 표현을 담아서는 안 된다.

### B.6 옮긴 번호 (자리표시, 계속)

**REQ-GDP-016** — **[MOVED → 카드 t658]** Route B SPEC당 PR 하나(B1).
**REQ-GDP-017** — **[MOVED → 카드 t658]** Route B Step 1~3 같은 워크트리와 서술형 안내 금지(B2).
**REQ-GDP-018** — **[MOVED → 카드 t658]** `delivery.md` Step 3.2 github-flow 전달 경로(B2).
**REQ-GDP-019** — **[MOVED → 카드 t658]** `delivery.md` Step 3.3.5 퇴역과 `ExitWorktree`(B2).
**REQ-GDP-020** — **[MOVED → 카드 t658]** `CONST-V3R5-027`·`028` amend 기록(B3).
**REQ-GDP-021** — **[MOVED → 카드 t658]** HumanOversight 승인 대행 금지(B3).
**REQ-GDP-022** — **[MOVED → 카드 t658]** 트리 빌드로 잰 헌법 검증 drift 0(B3).
**REQ-GDP-023** — **[MOVED → 카드 t658]** `zone-registry.md` 사본 일치와 새 clause(B3).

### B.7 Frozen 비접촉

**REQ-GDP-024** (Unwanted)
이 SPEC의 변경은 범위 파일 아홉 개(로컬·템플릿)의 `[ZONE:Frozen]` 줄과 `zone-registry.md` 에 등록된 Frozen clause 문장을 바꾸거나 지워서는 안 된다.

### B.8 `/moai sync` 플래그 의미와 모드별 승인 조건 (OD-2 하위 결정)

**REQ-GDP-025** (Ubiquitous)
`/moai sync` 의 플래그 표면 — `moai/SKILL.md:140` 의 Flags 줄, `references/reference.md:161` 의 sync 플래그 목록, `quality-gates-context.md:30` 의 `$ARGUMENTS` 목록과 `:101` 의 Supported Flags 절, `workflows/sync.md` 의 사용법 줄과 `**Flags**:` 줄, 명령 원본(`.claude/commands/moai/sync.md`, `sync.md.tmpl`)의 `argument-hint`, `delivery.md` Step 3.4 절과 Phase 14 다음 단계 선택지 절 — 은 병합 옵트인 플래그를 `--auto-merge` 로 노출해야 하고, `--merge` 는 `--auto-merge` 의 폐기된 별칭(사용하면 폐기 경고)으로만, `--no-merge` 는 동작을 바꾸지 않고 호환을 위해 남긴 폐기된 no-op(사용하면 폐기 경고)으로만 서술해야 한다. 이 표면은 `--merge` 를 `--auto-merge` 와 별개이거나 폐기되지 않은 auto-merge 플래그로 서술해서는 안 되며, `--no-merge` 를 병합을 건너뛰게 하거나 병합 조건을 바꾸는 플래그로 서술해서는 안 된다.

**REQ-GDP-026** (Event-driven)
When `--auto-merge`(또는 폐기된 별칭 `--merge`)가 주어져 sync 단계가 PR 병합을 판단할 때, team 모드는 전원 승인을 조건으로 병합해야 하고, personal·manual 모드는 승인 조건 없이 병합해야 한다(승인할 팀원이 없다). `manager-git.md` 의 PR Auto-Merge 절과 `delivery.md` Step 3.4 절은 각각 이 두 조건을 모드 이름을 담은 문장으로 적어야 하며, 두 문서의 조건은 같아야 한다. CI 통과와 충돌 없음 확인은 모든 모드에 그대로 남는다.

---

## §C 결정 기록

### C.1 결정

| 결정 | 선택 | 날짜 | 결정자 | 이력 |
|---|---|---|---|---|
| OD-2 — 워크트리 문맥 auto-merge 기본값의 단일 기준 | **선택지 B: `manager-git.md` 옵트인(`--auto-merge`)** | 2026-09-10 | 운영자(리드 경유) | 처음에는 선택지 C(설정 키 `workflow.worktree.auto_merge`)가 선택됐다. 그 키가 `class: D`, `evidence: none`, `deprecate_after: "v3.1.0"` 으로 기록돼 있다는 사실이 제시된 뒤 B로 바뀌었다. 이 SPEC은 그 키를 건드리지 않는다 |
| T1 — SPEC 분할 | **분할.** 이 SPEC은 AC-11, SX-R04 병합 방식, OD-2 = B와 그 파일들의 부수 의무만 남긴다. late-branch 재설계는 카드 t658, amend 적용 도우미 스텁은 카드 t659 | 2026-09-11 | 운영자(리드 경유) | 0.1.3에서 REQ 23·AC 24로 Tier M 상한을 넘은 뒤 결정 |
| OD-2 하위 결정 — `/moai sync` 플래그 의미 | **`/moai sync` 에 새 플래그 `--auto-merge`(manager-git 옵트인과 같은 이름)를 노출한다. `--merge` 는 `--auto-merge` 의 폐기된 별칭으로 남기고 폐기 경고를 낸다. `--no-merge` 는 병합하지 않는 것이 이제 기본이므로 호환용 no-op으로 남기고 폐기 경고를 낸다.** 범위: `moai/SKILL.md:140`, `references/reference.md:161`, `quality-gates-context.md:30`·`:101`, `workflows/sync.md` 플래그 줄, `delivery.md` 337-338·348-349(로컬·템플릿, 줄 단위) | 2026-09-11 | 운영자(리드 경유) | 축소판 plan-audit 1회차 B1이 세 가지 구현이 모두 가능하다고 지적한 뒤 결정 |
| 모드별 승인 조건 (공통 항목) | **personal·manual 모드: `--auto-merge` 는 승인 조건 없이 병합(승인할 팀원 없음). team 모드: 전원 승인 조건 유지.** 모드마다 수용 기준을 따로 둔다. personal·manual 규칙을 `manager-git.md` 절차가 있는 곳에 적어 `delivery.md` 와 같은 말을 하게 하고 `.toml` 은 기존 단계로 재생성한다 | 2026-09-11 | 리드 판정 | — |
| 소비자 범위 판정 (X1~X4) | **운영자의 "`--auto-merge` 노출" 결정을 일관되게 적용한다. X1(`workflows/sync.md` 사용법 줄), X2(슬래시 명령 `argument-hint`, 로컬 `.md`·템플릿 `.md.tmpl`), X3(`delivery.md:404`)는 범위에 넣는다(로컬·템플릿, 줄 단위). X2는 명령 원본이므로 run-phase가 `make commands-emit` 과 `make commands-emit-check`(exit 0)를 실행하고, 다시 만든 산출물은 원본 편집과 같은 커밋에 넣는다. X4(docs-site 네 로케일)는 범위 밖에 두고 리드가 큐 이관 뒤 별도 문서 카드로 발행한다. 같은 문자열·다른 뜻의 §D 구분은 유지한다. Tier M 유지(요구사항·수용 기준 수로 정하고 파일 수 안내는 참고)** | 2026-09-11 | 리드 판정 | 0.2.1 차단 항목 X1~X4 보고 뒤 |
| 흡수 뒤 REQ-GDP-002 처리 | **선택지 (a): develop 카드 t635 의 Lane A/B Pre-Spawn 형태가 REQ-GDP-002 를 충족한다. 이 카드는 `agent-common-protocol.md` 를 편집하지 않고 미러도 하지 않는다(카드 dr0911 이 미러했다). 요구사항의 실질·범위는 그대로이고 Kickoff 승인은 유지된다. 이어지는 plan-auditor 는 REQ-GDP-002·AC-GDP-002 변경분과 그 파생 항목만 본다** | 2026-09-11 | 리드 판단(운영자 경유, 최종) | 카드가 develop `ee99507fb`·`f1f034bb4` 를 흡수한 뒤(`.moai/reports/t622/absorb-ee99507fb.md`) |

OD-1·B1·B2·B3의 결정 기록은 카드 t658의 입력(§G)으로 옮겼다. 결정은 Implementation Kickoff Approval을 대신하지 않는다.

### C.2 번호 방식 — 원래 번호 유지

plan-auditor MP-1은 "REQ numbers must be sequential (REQ-001, REQ-002, ... REQ-N) with no gaps, no duplicates, and consistent zero-padding. Even one gap or duplicate = FAIL." 로 판정한다(`.claude/agents/moai/plan-auditor.md:138`).

- **택한 방식**: 0.1.3 번호를 유지하고, 옮긴 번호(007-010, 016-023)와 철회된 번호(011, 012)를 §B의 원래 자리에 한 줄 자리표시로 남긴다. §B에 REQ-GDP-001부터 026까지 모든 번호가 정확히 한 번씩 나타난다. 축소판 plan-audit 1회차는 이 방식을 결함으로 보지 않았다.
- **다시 매기지 않은 이유**: 다시 매기면 같은 토큰이 이 문서와 커밋 `87988e946`(t658의 입력)에서 서로 다른 요구사항을 가리킨다.
- **수용 기준도 같은 방식**: AC-GDP-001~024를 유지하고 새 기준은 AC-GDP-025~030을 받는다(acceptance.md §D).

### C.3 OD-2 검토한 선택지 (결정 당시 자료)

| 선택지 | 영향 파일 (L·T = 로컬·템플릿) | 배포 사용자에게서 철회·변경되는 기능 | 기본값 영향 | 뒤집거나 확장하는 선행 기록 |
|---|---|---|---|---|
| **A. `delivery.md` 워크트리 기본값** (미채택) | `manager-git.md` L·T 148·166 + 재생성 `.toml` | team 모드 워크트리 문맥에서 플래그 없이 병합 | 플래그로 결정 | `b7447cb90` 동작을 확정 |
| **B. `manager-git.md` 옵트인** (채택) | `delivery.md` L·T 335-338·348-349; `doc-execution.md` L·T 34-36 | 워크트리 문맥의 sync가 기본으로 병합하지 않는다. `--merge` 폐기 경고의 논리가 뒤집힌다 | 워크트리 문맥 동작 기본값 "병합" → "병합 안 함" | `b7447cb90` 동작을 뒤집음(SPEC 기원 없음) |
| **C. 설정 키 `workflow.worktree.auto_merge`** (처음 선택 후 B로 교체) | `delivery.md`, `doc-execution.md`, `manager-git.md` + 사용처 분류 파일 | 기본 병합이 멈춤 | 키 값 `false` 유지. 키는 `class: D`, `deprecate_after: "v3.1.0"` | 폐기 예정으로 분류된 키에 첫 소비자를 만든다 |

B 행이 적은 "`--merge` 폐기 경고의 논리가 뒤집힌다" 는 §C.1의 OD-2 하위 결정으로 해소했다.

### C.4 티어 재분류

| 신호 | 측정값 | Tier S 기준 | Tier M 기준 |
|---|---|---|---|
| 판정 대상 요구사항 수 | 12 (001-006, 013-015, 024-026) | 8 이하 | 16 이하 |
| 판정 대상 수용 기준 수 | 16 (001-006, 013-016, 025-030) | 8 이하 | 16 이하 |
| 영향 파일 수 | 19 (범위 파일 아홉 개 × 로컬·템플릿 18 + 생성물 `manager-git.toml` 1). 명령 게시본이 바뀌면(AC-GDP-030 경우 B) 2개 더해 21. 세는 규칙: 로컬·템플릿 사본 한 쌍을 범위 파일 하나로 센다(명령 원본 `sync.md`/`sync.md.tmpl` 도 한 쌍). 범위 파일 아홉 개는 `manager-git.md`, `agent-common-protocol.md`, `delivery.md`, `doc-execution.md`, `moai/SKILL.md`, `references/reference.md`, `quality-gates-context.md`, `workflows/sync.md`, 명령 원본이다. 생성물·게시본은 범위 파일로 세지 않고 영향 파일로만 센다. 0.2.2까지의 "열 개"·"21" 은 이 목록을 잘못 센 값이다. **0.2.5**: `agent-common-protocol.md` 는 범위 파일(사본·Frozen·보존 판정 대상)로 남지만 편집하지 않으므로 이 카드가 실제로 바꾸는 파일은 17(여덟 쌍 16 + `.toml` 1), 경우 B 19 | 5개 미만 | 5~15개 (안내) |
| 예상 변경량 | 수십 줄 규모 | 300줄 미만 | 300~1000줄 |

요구사항·수용 기준 수는 Tier S를 넘고 Tier M 상한 안이다(수용 기준은 상한과 같은 16). 영향 파일 수는 Tier M 파일 수 안내를 넘지만, 리드가 "요구사항·수용 기준 수로 정하고 파일 수 안내는 참고" 로 판정했다(§C.1). **`tier: M` 을 유지한다.** 자리표시 번호(옮김 12·철회 2)는 판정 대상이 아니므로 상한 계산에 넣지 않았다.

---

## §D 제외 범위 (Out of Scope)

### Out of Scope — 카드 t658·t659로 옮긴 작업

- late-branch 재설계 전체: AC-01 명령·서술형·단계별 PR 안내, OD-1 선택지 1, B1·B2·B3·B6, `delivery.md` Step 3.2 github-flow와 Step 3.3.5, `spec-workflow.md`·`spec-assembly.md`·`zone-registry.md` 편집 → 카드 t658 (§G).
- `internal/constitution/pipeline.go` 의 amend 적용 도우미 스텁(B4) → 카드 t659 (t658이 의존).

### Out of Scope — docs-site 문서 (X4, 후속 문서 카드)

- docs-site 네 로케일의 `workflow-commands/moai-sync.md` 와 `core-concepts/what-is-moai-adk.md`. 리드 판정(2026-09-11)으로 범위 밖에 두며, 리드가 큐 이관 뒤 별도 문서 카드로 발행한다. 이 문서 공백은 그 후속 카드가 소유한다.
- **이 SPEC이 착지한 뒤 후속 카드가 끝날 때까지 배포된 문서는 옛 플래그와 옛 기본값을 설명한다** — `--merge` 를 자동 병합 플래그로, `--no-merge` 를 자동 병합 건너뜀으로, 워크트리 문맥의 자동 병합을 기본값으로.
- 후속 카드 입력 — 줄 수 측정 명령(기준 트리, 0.2.5에서 BASE 로 재측정 — 값은 0.2.2 `b412f8a33` 측정과 같다, `.moai/reports/t622/reanchor/x4-docs-count.txt`·`x4-docs-lines.txt`): `git grep -c -e --merge -e --no-merge 255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089 -- docs-site/content/en/workflow-commands/moai-sync.md docs-site/content/ko/workflow-commands/moai-sync.md docs-site/content/ja/workflow-commands/moai-sync.md docs-site/content/zh/workflow-commands/moai-sync.md docs-site/content/en/core-concepts/what-is-moai-adk.md docs-site/content/ko/core-concepts/what-is-moai-adk.md docs-site/content/ja/core-concepts/what-is-moai-adk.md docs-site/content/zh/core-concepts/what-is-moai-adk.md`. 줄번호는 같은 경로에 `git grep -n -e --merge -e --no-merge -e --auto-merge 255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089 -- …` 로 쟀다(`--auto-merge` 적중 0).

| 파일 | en | ko | ja | zh |
|---|---|---|---|---|
| `workflow-commands/moai-sync.md` 줄 수 | 5 (80, 82, 188, 303, 480) | 4 (80, 82, 299, 473) | 6 (80, 82, 299, 476, 477, 482) | 9 (77, 92, 97, 108, 189, 304, 485, 486, 491) |
| `core-concepts/what-is-moai-adk.md` 줄 수 | 1 (388) | 1 (391) | 1 (391) | 1 (389) |

- 워크트리 기본 병합을 설명하는 줄: ko 473 "워크트리 컨텍스트에서는 플래그를 따로 주지 않아도 자동 머지가 기본입니다", en 480 "In a worktree context, auto-merge is the default behavior with no extra flag". 둘 다 "`/moai sync` 가 받는 플래그는 `--pr` / `--merge` (deprecated) / `--skip-mx`" 로 끝난다.
- `--no-merge` 를 자동 병합 건너뜀으로 적은 표와 조건: ja 477 "`--no-merge` | N/A | 自動マージのスキップ" 와 482 "`--no-merge` フラグ未設定"(트리거 조건), zh 486 "`--no-merge` | N/A | 跳过自动合并" 와 491 "未设置 `--no-merge` 标志"(트리거 조건).
- `--merge` 를 자동 병합 플래그로 소개하는 줄: zh 77 "`/moai sync --merge`" 표, 92 "### --merge 标志", 97, 108, 189 흐름도, 304; en 188 흐름도·303.
- `what-is-moai-adk.md` 네 로케일의 sync 행은 플래그를 `--merge`, `--skip-mx` 로 적는다.

### Out of Scope — 같은 문자열이지만 다른 뜻

- `gh pr merge --merge`(병합 커밋 방식 선택): `.claude/agents/harness/hns-release-specialist.md` 74·232·371·379(로컬 전용 하네스), `moai-ref-git-workflow/SKILL.md:134`(로컬·템플릿), docs-site `worktree/examples.md` 네 로케일. `/moai sync` 플래그가 아니다.
- `--merged-only`(`moai worktree clean`), `git branch --merged`(브랜치 가드·gitflow 규칙), `docs-site/README.md:348` `moai update --merge`(다른 명령).

### Out of Scope — 이 SPEC의 편집 대상이 아닌 것

- `spec-workflow.md`, `spec-assembly.md`, `zone-registry.md`, `worktree-integration.md`.
- `agent-common-protocol.md`(로컬·템플릿) 편집과 미러. REQ-GDP-002 는 develop 카드 t635·dr0911 로 이미 충족돼 있다(0.2.5, 리드 판단 (a)). 이 파일은 사본 일치·Frozen·보존 판정 대상으로만 남는다.
- `manager-git.md` 의 Late-Branch Invocation Pattern 절에서는 114행(로컬 116) 병합 예시만 고친다. 나머지 절차 줄과 42·160·171행(로컬 44·162·173)은 t658 소관이다.
- `delivery.md` Step 3.4 안의 356행(로컬 381) "Checkout target branch, fetch latest" 와 `manager-git.md:171`(로컬 173) 은 t658 입력으로 넘긴다(§G.2).
- 명령 원본의 `description` 줄과 본문, 발행기 코드(`internal/template/commandemit`), 게시본 손편집.
- 설정 키 `workflow.worktree.auto_merge` 와 사용처 분류 파일, Go 코드, `git-strategy.yaml.tmpl` 값.
- `[ZONE:Frozen]` 줄과 등록된 Frozen clause 문장(REQ-GDP-024).

### Out of Scope — 금지 목록과 경고표

- 범위 파일 안의 금지 목록·경고표·인용 예시에 들어 있는 `git` 명령 문구는 이 SPEC이 판정하지 않는다.

---

## §E 미검증과 잔여 위험

### E.1 미검증 (Gaps)

- `git fetch`/`git rev-list` 경합을 실제로 일으키지 않았다.
- 범위 파일 중 `agent-common-protocol.md` 를 뺀 여덟 파일(`manager-git.md`, `delivery.md`, `doc-execution.md`, `moai/SKILL.md`, `references/reference.md`, `quality-gates-context.md`, `workflows/sync.md`, 명령 원본 `sync.md`/`sync.md.tmpl`)은 로컬·템플릿 사본 일치를 지키는 테스트가 없다. 사본 일치는 이 SPEC의 diff 판정으로만 확인된다.
- `TestSanitizedPairParity` 는 토큰 정규화 뒤 줄 차이를 허용 오차(`structuralDriftToleranceLines = 4`) 안에서 받아들여, `agent-common-protocol.md` 의 한 줄짜리 사본 차이는 이 테스트만으로 드러나지 않을 수 있다.
- `/moai sync` 플래그는 Go 코드가 해석하지 않고 스킬 지침을 읽은 오케스트레이터가 해석한다. 이 SPEC의 수용 기준은 지침 문구를 판정할 뿐, 실제 sync 실행에서 `--merge` 가 경고를 내고 `--no-merge` 가 아무것도 바꾸지 않는지는 관측하지 않는다. 슬래시 명령 힌트가 실제 입력 창에 어떻게 보이는지도 관측하지 않는다.
- `argument-hint` 편집이 게시본을 바꾸지 않는다는 예상(AC-GDP-030 경우 A)은 발행기 소스 읽기(`loader.go:4-6`, `emit.go` `renderSkill`)에 근거한다. plan 작성 시점에는 `make commands-emit`·`make commands-emit-check` 를 실행하지 않았다(리드 제약). 발행기는 템플릿 트리 게시본만 쓰므로, 경우 B가 되면 로컬 게시본 사본을 템플릿 게시본으로 맞추는 일이 따로 필요하다.
- `make agents-emit-check` 기준선(재생성 전 exit 0)도 plan 작성 시점에 실행하지 않았다.
- 워크트리 기본 병합 문구의 `git log -S` 추적은 `delivery.md:349` 문장에만 수행했다.
- (0.2.5) REQ-GDP-002 충족 여부는 t635 형태의 지침 문구를 판정할 뿐, 그 블록을 실제로 실행해 fetch 실패 시 멈추는지는 관측하지 않았다.
- (0.2.5) develop 이 로컬 사본에만 넣은 편집이 의도된 분기인지 템플릿 미러 누락인지는 판정하지 않았다 — 해당 develop 카드 소관이다. 이 SPEC은 그 차이를 BASE 기준선으로 받아들여 보존만 판정한다.

### E.1b 0.2.5 재고정 증거

`.moai/reports/t622/reanchor/` — `base-vs-head-stat.txt`, `lines-*.txt`(인용 줄), `pair-*.diff`·`pair-hunks.txt`(사본 덩어리), `mut013/`(사본 본문 비교 뮤턴트), `ac002-*`(AC-GDP-002 판정·뮤턴트), `ac003-*`, `ac016-*`(AC-GDP-016 판정·양성 대조), `mirror-*`·`mutmirror/`(미러 테스트 기준선과 집합 비교 뮤턴트), `frag/`(AC-GDP-026 조각 로컬·템플릿 비교), `x4-docs-*`, `cf-*`, `index.md`(명령과 관측 값 목록).

0.2.6 수리 증거는 `.moai/reports/t622/reanchor-fix/` — 범위 판정 재현(현재 트리와 흉내 낸 흡수 커밋 `7fae4764e`: 참조를 옮기지 않은 `git merge-tree --write-tree` + `git commit-tree` 객체), 양성 대조(t645·dr0911 카드 이력), 스냅숏 신선도 점검, `ac002/`(AC-GDP-002 판정기·뮤턴트 i~ix 표), `index.md`(명령과 관측 값 목록).

### E.2 잔여 위험

- **t658과의 겹침.** `manager-git.md:114`(로컬 116, 이 SPEC)는 t658이 다시 쓸 Late-Branch 절 안에 있다. t658은 이 SPEC이 병합된 뒤의 줄을 입력으로 받아야 하며, 순서가 뒤바뀌면 `--<merge_method>` 치환이 되돌려질 수 있다.
- **사용자에게 보이는 기본값 변경과 문서 공백.** OD-2 = B와 하위 결정으로 워크트리 문맥의 기본 병합이 없어지고 `--no-merge` 는 효과가 없어진다. X4 후속 문서 카드가 끝날 때까지 공개 문서는 옛 동작을 안내한다(§D).
- **검출식 한계.** AC-GDP-001·006·026~029의 검출식은 줄 단위이며, 한 조건을 여러 줄에 나눠 적거나 검출식이 예상하지 않은 표현을 쓰면 통과할 수 있다. AC-GDP-001·006·026·028·029는 읽기 기록(`ac001`·`ac006`·`ac026`·`ac028` reading)을 PASS 전제로 둬 틈을 좁히지만 오독 가능성이 남는다. AC-GDP-027은 자동 검출만으로 판정하므로 효과 낱말 목록 밖의 표현으로 `--no-merge` 에 동작을 부여한 줄은 통과할 수 있다.
- **발행 대조의 일시 변경.** AC-GDP-030의 양성 대조는 run-phase 사전 점검에서 템플릿 게시본 한 파일을 잠시 바꿨다가 백업으로 되돌린다. 되돌림은 `cmp` 와 `git status` 로 확인하지만, 그 사이 다른 작업이 같은 파일을 쓰면 섞일 수 있다.
- **인용 줄번호 전제.** 범위 파일이 BASE(`255f88eb0`)와 바이트 동일하다는 전제에 기댄다(0.2.5 작성 시점 HEAD `b24f2e184` 에서 확인 — BASE 와 HEAD 의 차이는 보고서 두 파일, 범위 루트 작업 트리 차이 없음, §A.1). 이 전제는 아래 스냅숏 신선도 점검이 빈 결과인 동안만 선다.
- **통합 창의 develop 재흡수와 판정 범위 (0.2.6).** 이 저장소 절차는 통합 창에서 로컬 develop 을 카드 워크트리에 흡수한 뒤 병합 트리에서 재측정한다(`CLAUDE.local.md` §4.1). 재흡수는 선택이 아니라 예정된 단계이고, 로컬 develop 은 계획 시점에 이미 카드보다 앞서 있다(읽을 때 `git rev-list --count --left-right develop...HEAD` 로 다시 잰다 — 0.2.6 작성 시점 참고값 `17 27`). 그래서 이 SPEC 의 값은 두 종류로 나뉜다.
  - **고정 스냅숏 — `$BASE` 를 쓰고 움직이지 않는다.** 기준 사본 반출(plan §C 2단계: 범위 파일 로컬·템플릿, 생성물 `.toml`, 게시본), 양성 대조(`b412f8a33..$BASE` 구간, 반출본에 돌리는 "대조" 명령), 사본 덩어리 기준선(§A.4, AC-GDP-013 (b)), 미러 테스트 기준선(AC-GDP-013 (e)), AC-GDP-002·003 의 기준 절, AC-GDP-025 의 clause 개수 기준.
  - **읽는 시점 범위 — `$CARD_BASE` 를 쓰고 핀하지 않는다.** "이 카드가 무엇을 바꿨는가·커밋했는가" 를 재는 AC-GDP-014·015·016·025·030 과 plan M5. 왼쪽 끝은 읽을 때마다 `git merge-base develop HEAD` 로 다시 구한다(`.claude/rules/local/gitflow-lane-protocol.md` §8). 차이는 `git diff develop...HEAD -- <경로>`(= `git diff $CARD_BASE HEAD -- <경로>`), 커밋 목록은 `git log --no-merges --format=%H HEAD --not develop -- <경로>` 로 잰다. 흡수 뒤 develop 이 더 앞서가도 merge-base 는 마지막으로 흡수한 develop 커밋에 머물러 계속 옳다 — `verification-completeness.md` §4 가 경계하는 "위쪽이 움직여 생긴 가짜 변경" 이 이 형태에서는 생기지 않는다. 대조: `git diff --name-only develop...HEAD` 가 1줄 이상이어야 하고, 0줄이면 "변경 없음" 이 아니라 "측정 불가" 로 보고한다.
  - **한계 — 병합 전 전용.** 카드가 develop 에 병합되면 merge-base 가 카드 tip 자신이 되어 범위가 빈다. 병합 뒤 근거는 범위 판정이 아니라 병합 트리 동일성(병합 커밋의 `git rev-parse <merge>^{tree}` 가 재측정한 트리와 같음)이다.
  - **BASE 를 흡수 병합으로 옮기지 않는다.** 0.2.5 의 대응(재흡수 뒤 BASE 를 흡수 병합으로 옮김)은 카드 커밋을 새 BASE 의 조상으로 만들어 범위에서 빼 버린다 — 흉내 낸 흡수에서 그 범위는 0줄이었다(`.moai/reports/t622/reanchor-fix/old-remedy-moved-base-range.txt`). 리터럴 `$BASE..HEAD` 를 그대로 두는 것도 틀린다 — 같은 흉내 흡수에서 비병합 커밋이 3개에서 12개로 늘었고, 늘어난 9개는 모두 develop 커밋이었다(`reanchor-fix/literal-base-range-emulated-foreign.txt`).
  - **판정 전 스냅숏 신선도 점검.** 판정을 시작하기 전에 `git diff --name-only $BASE $CARD_BASE -- <스냅숏 경로>` 가 빈 결과인지 확인한다(acceptance.md 관례). 비어 있지 않으면 흡수가 스냅숏 파일을 바꾼 것이다. 그 파일의 고정 스냅숏은 낡았으므로 그대로 판정하지 않는다 — 기준 사본을 `$CARD_BASE` 에서 다시 반출하고, 그 파일이 쓰는 기준선(§A.1 줄, §A.4 덩어리, AC-GDP-002·003 기준 절, AC-GDP-025 clause 개수)을 다시 재어 기록한 뒤 리드에 올린다. 미러 테스트 기준선은 `git diff --name-only $BASE $CARD_BASE -- internal/template/ .claude/rules/moai/` 가 비어 있지 않으면 낡았을 수 있다. 그때 기준선 밖 FAIL 은, 그 하위 테스트가 읽는 두 사본이 `git diff --name-only develop...HEAD` 에 없으면 흡수 탓으로 기록하고 리드에 올린다(이 카드의 FAIL 로 세지 않는다). 0.2.6 작성 시점 측정: 현재 merge-base `f1f034bb4` 와 흉내 낸 흡수 뒤 merge-base `00ae57ad7` 모두에서 스냅숏 경로 점검은 빈 결과였고, 같은 형태를 `b412f8a33..$BASE` 에 돌리면 8개 파일이 나온다(양성 대조, `reanchor-fix/snapshot-stale-*.txt`).
- **REQ-GDP-002 의 충족을 develop 에 기댄다.** 이후 develop 이 Pre-Spawn 절을 다시 바꾸고 이 카드가 그것을 흡수하면 스냅숏 신선도 점검이 `agent-common-protocol.md` 를 낡은 스냅숏으로 잡는다. 원인은 이 카드 편집이 아니라 흡수이므로, 기준 절을 `$CARD_BASE` 에서 다시 반출하고 원인을 기록해 리드에 올린다. 절 밖 문장으로 Pre-Spawn 의 순서를 뒤집는 편집(AC-GDP-002 뮤턴트 ix)은 AC-GDP-002 가 보지 않는다 — 이 카드의 커밋이면 AC-GDP-016, 흡수로 들어오면 신선도 점검이 잡는다.

---

## §F 수용 기준

`acceptance.md` 가 판정 대상 기준(AC-GDP-001~006, 013~016, 025~030)과 검증 명령, 양성 대조, 뮤턴트 재실행 결과, 옮긴 번호의 자리표시를 담는다.

---

## §G 카드 t658로 옮긴 항목 (Moved to card t658)

원문 출처: 커밋 `87988e946` 의 `.moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/` (spec.md·acceptance.md·plan.md·progress.md). 표의 번호는 그 판의 번호이며, 이 문서의 §B·acceptance.md에는 같은 번호가 자리표시로만 남는다. §G.2 의 줄번호는 템플릿 사본 기준이고(로컬은 §A.1 표 — `manager-git.md` +2, `delivery.md` 279~421행 +25), "기준 트리 측정(`b412f8a33`)" 은 t658 입력으로 그 트리에서 잰 값이라 0.2.5 재고정과 무관하게 그대로 둔다.

### G.1 요구사항·수용 기준

| 원래 ID | 제목 | 한 줄 내용 | 출처 커밋 |
|---|---|---|---|
| REQ-GDP-007 | AC-01 불변식 | 명령 형태 안내 위치가 primary checkout 브랜치 변경을 지시하지 않음 | `87988e946` |
| REQ-GDP-008 | 대체 절차 없는 삭제 금지 | 예시만 지우지 않고 `manager-git.md` 절차 참조와 대상 제목 유지 | `87988e946` |
| REQ-GDP-009 | 흐름 일치·단일 접두 | 네 지침 절이 같은 진입·PR·종결 명령, PR 접두 `feat/SPEC-` 하나 | `87988e946` |
| REQ-GDP-010 | 런처 워크트리 흐름 | 워크트리 안 커밋, 맨손 `git worktree add` 금지, `manager-git.md` 42·160 처리, Frozen 줄 기록 | `87988e946` |
| REQ-GDP-011 | [WITHDRAWN] 조건부 유지 | OD-1 선택지 2 경로 — 0.1.2에서 철회 | `87988e946` |
| REQ-GDP-012 | [WITHDRAWN] `main_late_branch` 은퇴 | OD-1 선택지 3 경로 — 0.1.2에서 철회 | `87988e946` |
| REQ-GDP-016 | SPEC당 PR 하나 (B1) | 단계별 PR과 그 브랜치 금지, 동기화 커밋은 같은 SPEC PR 브랜치 | `87988e946` |
| REQ-GDP-017 | 같은 워크트리·서술형 안내 금지 (B2) | Route B Step 1~3 같은 워크트리, 서술형 primary 브랜치 안내 금지 | `87988e946` |
| REQ-GDP-018 | Step 3.2 github-flow (B2) | 워크트리 안 PR 브랜치만 전달, primary 비-main 브랜치에서 멈춤 | `87988e946` |
| REQ-GDP-019 | Step 3.3.5 (B2) | 기준 브랜치 복귀 대신 `ExitWorktree`, 병합 착지 뒤 폐기 | `87988e946` |
| REQ-GDP-020 | amend 기록 (B3) | `CONST-V3R5-027`·`028` 을 `--before`·`--after`·`--evidence` 로 개정, 증거 추적 | `87988e946` |
| REQ-GDP-021 | 승인 대행 금지 (B3) | 레인은 HumanOversight 승인을 대신하지 않고 멈춰 올림 | `87988e946` |
| REQ-GDP-022 | 트리 빌드 검증 (B3) | 트리 빌드 `constitution validate` ok·drift 0, 빌드 커밋 = HEAD | `87988e946` |
| REQ-GDP-023 | 레지스트리 일치 (B3) | `zone-registry.md` L·T 동일, 새 clause 1회·이전 clause 0회 | `87988e946` |
| AC-GDP-007 | 명령 검출식 분류 장부 | 금지 집합 검출식 적중 분류, primary 실행 안내 0 (기준 트리 20줄) | `87988e946` |
| AC-GDP-008 | 절차 참조 유지 | 파일별 형식 참조 하한·느슨한 개수 일치·제목 조회 | `87988e946` |
| AC-GDP-009 | 흐름 표지·접두 합집합 | 네 절 표지 매트릭스, 접두 합집합 = {feat/SPEC-} (기준 트리 {chore, feat, plan, sync}) | `87988e946` |
| AC-GDP-010 | 워크트리 흐름·Frozen 기록 | `git worktree add` 부재, §E.2 Frozen·Kickoff 기록 | `87988e946` |
| AC-GDP-011 | [철회] | OD-1 선택지 2 경로 | `87988e946` |
| AC-GDP-012 | [철회] | OD-1 선택지 3 경로 | `87988e946` |
| AC-GDP-017 | 단계별 PR 분류 장부 | 단계별 PR 검출식 적중 분류 (기준 트리 19줄) | `87988e946` |
| AC-GDP-018 | Step 3.2 github-flow | Feature 경로 0·런처 표지·primary checkout 멈춤 경로 | `87988e946` |
| AC-GDP-019 | Step 3.3.5 결과 | `ExitWorktree`·명령 0·"return to base branch" 0·병합 착지 표지 | `87988e946` |
| AC-GDP-020 | amend 명령 기록 | 근거 문서·명령 원문·dry-run·운영자 실행 출력 | `87988e946` |
| AC-GDP-021 | 서술형 분류 장부 | 서술형 검출식 적중 분류 (기준 트리 11줄) | `87988e946` |
| AC-GDP-022 | 트리 빌드 헌법 검증 | 빌드 출처 기록과 `status: ok`·`drift_count: 0` (짝 대조 측정 포함) | `87988e946` |
| AC-GDP-023 | 레지스트리 사본·새 clause | L·T diff 0, 새 clause 1·이전 clause 0 | `87988e946` |
| AC-GDP-024 | 승인 대행 금지 | 명령 기록에 표준입력 주입 없음, 운영자 `(Y/N)`·리드 에스컬레이션 기록 | `87988e946` |

### G.2 t658 입력 메모 (이 SPEC의 위험·사실에서 옮긴 것)

- **결정 기록**: OD-1 = 선택지 1(런처 워크트리 흐름, 2026-09-10), B1 = SPEC당 PR 하나 `feat/SPEC-*`, B2 = primary feature 브랜치 경로도 워크트리 흐름으로, B3 = `moai constitution amend` 로 `CONST-V3R5-027`·`028` 개정(HumanOversight는 운영자, 레인은 멈춰 올림). 모두 운영자 결정(리드 경유). OD-1 선택지 표는 `87988e946` spec.md §C.6.
- **REQ-WBG-011 근거 소실**: Phase D가 사라지면 SPEC-WORKTREE-BRANCH-GUARD-001 REQ-WBG-011의 근거 문장("Phase D is the ONLY place…", spec.md 210행)이 성립하지 않고 `internal/hook/branch_guard.go:455` 의 manager-git 신원 면제가 근거 없이 남는다 — 후속 카드 후보.
- **git-flow와 B1 접두**: `delivery.md` Step 3.2 git-flow는 `feature/*` 만 PR로 받아 `feat/SPEC-*` 브랜치는 "Any other branch → stop and report" 로 떨어진다. 배포 `AGENTS.md:125` 의 `WT-<slug>` 명명과 C1 이름 변경이 겹친다.
- **`feature/SPEC-` 형제**: `manager-git.md:82`·`:144`, `spec-assembly.md:388`(`--branch` 경로), `:547` 시나리오.
- **브랜치 변경 형제**: `manager-git.md:171`, `delivery.md:356`, `generic-patterns-guide.md`(템플릿 195행~, 로컬 207행~) Phase D Recovery.
- **Frozen 레지스트리 사실**: spec-workflow.md `[ZONE:Frozen]` 23·49·166행 중 등록 clause는 49행 블록의 `CONST-V3R5-027`·`028` 뿐, 23행 블록·166행은 미등록. `validator.go` 는 `ZONE_UNREGISTERED` 를 만들지 않는다. 검증기는 프로젝트 밖 레지스트리를 거부한다("escapes project dir"). worktree-integration.md 45·543-556행이 B1·B2와 모순되며 545·556행 `[ZONE:Frozen]` 은 미등록(B6).
- **amend 동작**: `--before` 바이트 대조(`internal/cli/constitution.go:528-529`), 레지스트리 해석 두 갈래(`:143-153` 대 `internal/constitution/pipeline.go:66`), 게이트 위치(`frozen_guard.go:22`, `canary.go:16/18/38`, `contradiction.go:87/110/124`, `rate_limiter.go:10/12/92`, `human_oversight.go:32/44-45`), 적용 도우미 스텁(`pipeline.go:256-266`, `pipeline_test.go:334-356`·`406-423`) → 카드 t659.
- **기준 트리 측정(`b412f8a33`)**: 명령 검출식 20줄(10/5/1/4), 서술형 검출식 11줄, 단계별 PR 검출식 19줄, 접두 합집합 {chore, feat, plan, sync}, delivery Step 3.2 github-flow 30줄, Step 3.3.5 11줄.
- **이 SPEC과의 겹침**: `manager-git.md:114` 의 `--<merge_method>` 치환은 이 SPEC이 먼저 착지한다(§E.2).
