# SPEC-DRIFT-CLOSE-BODY-001 — 진행 기록

plan_status: audited (PASS-WITH-DEBT 0.80) — run 진입 승인 2026-09-03

## §E.1 Plan-phase Audit-Ready Signal

- 카드: t410 (Class C) · 브랜치 `WT-drift-false-positive` · 트리 `.claude/worktrees/t410`
- Tier: S (spec.md + plan.md; AC는 spec.md §3 인라인). 근거는 `plan.md` §A Tier S 판정 근거
- REQ 7 / AC 7 — Tier S 상한(각 8) 안
- SPEC-ID 정규식 자체 점검: `[[ "SPEC-DRIFT-CLOSE-BODY-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`. `.moai/specs/` 충돌 없음
- 조사 원장: `.moai/reports/t410/discovery.md` + `r1-walker-trace.log` · `r2-blast-radius.log` · `r3-drift-row.log`
- 수리 후보 판정: **B 채택**(close 선언 커밋에 한해 본문 조회), A 기각(측정된 위험), C 기각·D 범위 밖 — `spec.md` §5
- 미확정으로 남긴 것: 파급 건수. 조사의 LOOSE 33 / TIGHT 15는 경계값이며 TIGHT에 알려진 오탐 1건 포함. 실제 수는 run-phase M3 전수 대조의 **결과**로 정해진다(AC-DCB-005)
- plan-audit: **PASS-WITH-DEBT · 0.80** (Tier S PASS 임계 0.75, must-pass 7/7, iteration 1/1). 판정본 `.moai/reports/t410/plan-audit-verdict.md`. 남은 부채: blocking 3건은 아래 §E.1.1대로 상환됨, D4·D6은 run-phase 원장 처리로 이월

### §E.1.1 감사 고정본 이후 편집 델타 실측 (리드 [HARD], 2026-09-03)

**Claim** — 감사 고정 이후의 spec.md 편집은 감사가 본 AC/REQ 집합을 바꾸지 않는다. 판정서 결함 상환에 국한된다.

**Evidence**

- 해시 대조: 고정본 `ff4489f3140aaf514d2f9cbc3e4bd6e39c81a12f5c6b07ccb69364d03f46ae96` (판정서 pin, 2026-09-03T06:27:42Z) vs 현재본 `shasum -a 256` → `0bf04b1d05522e1ab30271b2738a830122f7bd79504c99850fc25bad2e926fca` — **다름**. `plan.md`은 `9b69231a…`로 **고정본과 동일**(편집 없음)
- 집합 불변: `grep -c '^\*\*REQ-DCB-00'` → `7`; REQ id `001~007` 연속; AC id `001~007` — 판정서 MP-1 측정과 동일
- 좌표 불변: 판정서 인용 `spec.md:2`(프론트매터 12필드+`tier: S`), `spec.md:84~96`(GEARS 블록, REQ-001~007 형식 동일) — 현재본에서 같은 줄에 그대로
- 델타 위치: 판정서 인용 `spec.md:115`(AC-DCB-002 Given) → 현재 116, `spec.md:164`(AC-DCB-007 술어) → 현재 172. 삽입 +1 / +7, 전부 §3 AC 본문 안
- 델타 방향: spec.md HISTORY `0.3.0` 행이 자기기술 — "plan-audit(t410, PASS-WITH-DEBT 0.80) 차단 결함 3건 + 문서 내 모순 2건 상환". D1 술어 `^[-+].*^status:` → `^[-+]status:` 교체(현재본 AC-DCB-007에서 확인), D2 모양-B 자격+전수 훑기 명시·픽스처 10→12줄, D3은 **AC 신설 없이** AC-DCB-003 (d) 케이스로 병합(AC 7개 유지), D5 §5 표·§5.2 정정, D7 AC-DCB-002 Given 보강. `version:` 0.1.0→0.3.0 정합 수리 동반

**Baseline-attribution** — 전부 이 트리(`.claude/worktrees/t410`), origin/develop `7835148d3` + 로컬 develop `6765a75c0` 흡수 후 HEAD `460a7e16d`에서 실측. `git status --porcelain` 0행

**Gaps** — 리드가 지정한 `git diff <감사시점 SHA>..HEAD -- spec.md` 는 **성립하지 않는다**. 고정본은 커밋된 적이 없는 워킹트리 판본이고(`git show HEAD:…spec.md | shasum` → `0bf04b1d…` = 현재본), 감사 시각의 트리 HEAD `4e4607abe`에는 이 SPEC 파일 자체가 없다. 따라서 델타 **원문**은 재구성 불가이며, 위 증거는 원문 diff가 아니라 등가 검사(집합·좌표·자기기술 HISTORY)다

**Residual-risk** — 등가 검사는 AC/REQ의 **개수·id·좌표**와 자기기술 기록을 확인할 뿐, 감사가 읽은 AC **본문 문장**이 그 밖에서 바뀌지 않았음을 증명하지 못한다. 판정서가 좌표를 인용한 3개 AC(002·006·007)는 직접 읽어 대조했고 나머지 4개는 대조하지 않았다

**판정** — 델타가 AC·REQ 집합을 건드리지 않았으므로 리드가 건 정지 조건(집합 변경 시 재감사 요청)은 불성립. run 진행

## §E.2 Run-phase Evidence

_\<pending run-phase\>_

## §E.3 Run-phase Audit-Ready Signal

_\<pending run-phase\>_

## §E.4 Sync-phase Audit-Ready Signal

_\<pending sync-phase\>_
