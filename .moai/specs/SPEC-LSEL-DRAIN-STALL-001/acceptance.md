# Acceptance — SPEC-LSEL-DRAIN-STALL-001

> **Format note:** 본 파일은 검증 레이어 — 모든 항목은 `AC-LDS-XXX` 라벨의 `Given … When … Then …` 이고 이진 판정 가능해야 한다. GEARS 의무는 spec.md §C(요구 레이어)에 있다. 여기에 GEARS를 재술하지 않고, GWT를 GEARS로 제시하지도 않는다.
> **Two-cell discipline** (`verification-completeness.md` §2): 각 AC는 RED-now 셀(사전구현 트리에서의 관측, 측정 트리 명시)과 green-path 셀(어느 마일스톤이 뒤집는지 + 통과 출력)을 쌍으로 갖는다. RED는 "왜 빨간지"가 서술된다.

## §A. Verification Lens

- **Surface 3** (결함/상태 주장): 기계 도구가 상태를 확인 (test, grep, jq, exit-code, 파일 존재). 본 SPEC 대부분.
- **Surface 1** (자기보고): §E 행렬이 명령+관측 출력을 인용. M2 live 드레인 증거가 여기 속한다 (progress.md §E.2).
- 양방향 no-claim-without-observation: PASS는 관측된 검사, FAIL도 동일.

## §B. Severity Conventions

- **MUST** — 마일스톤 출하 차단. **SHOULD** — 이탈 시 plan.md에 근거 기록. **MAY** — 선택.

## §C. Traceability (카드·spec 장 ↔ AC)

| AC | 근거 | 마일스톤 |
|---|---|---|
| AC-LDS-001..005 | 카드 범위 (1) 원인 대응 — 트리거 내구화 (REQ-LDS-001..005) | M1 |
| AC-LDS-006/007 | 카드 범위 (2) 정지 신호 (REQ-LDS-005/006) | M1 |
| AC-LDS-008/009 | 위생 (REQ-LDS-010) — §B.3 기여 결함 2·3 | M1 |
| AC-LDS-010 | 카드 범위 (3) 백로그 방침 (REQ-LDS-007/008) — **mutant guard** | M2 |
| AC-LDS-011 | 카드 경계 (spec.md §D) — CI guard·PRESERVE | M2 (M1에서 선행 관측) |
| AC-LDS-012 | 배선 정직성 (REQ-LDS-009) | M1 문서화 / M2 적용 |

## §D. AC Matrix

| AC ID | REQ | Severity | 요약 | Milestone |
|-------|-----|----------|------|-----------|
| AC-LDS-001 | REQ-LDS-001 | MUST | wrapper가 fixture 백로그를 드레인 — 오프셋 전진+후보 스테이징+상태 1행 | M1 |
| AC-LDS-002 | REQ-LDS-002 | MUST | 잠금 경쟁에서 skip + exit 0 + 경쟁 통지 | M1 |
| AC-LDS-003 | REQ-LDS-003 | MUST | clusters.json 후보 보존 — history 사본 후 덮어쓰기 | M1 |
| AC-LDS-004 | REQ-LDS-004 | MUST | offset==tail no-op — exit 0, 오프셋 불변, 실패 아님 | M1 |
| AC-LDS-005 | REQ-LDS-001..004, 007 | MUST | session_drain_test.sh 5경로+mutant probe 전부 초록 | M1 |
| AC-LDS-006 | REQ-LDS-005 | MUST | fail-open — wrapper 자체 오류가 세션 시작을 막지 않음 | M1 |
| AC-LDS-007 | REQ-LDS-006 | MUST | backlog advisory 임계 초과 발화 + 같은 SessionStart 표면 배선(로컬) | M1 |
| AC-LDS-008 | REQ-LDS-010 | MUST | loop 레시피 자기기술 진실 — "실행한다" 주장 0, wrapper 참조 ≥1 | M1 |
| AC-LDS-009 | REQ-LDS-010 | MUST | 죽은 §28 앵커 0 + SKILL.md 내구-운영 섹션 존재 | M1 |
| AC-LDS-010 | REQ-LDS-007/008 | MUST | live 일괄 드레인 — offset==live && candidates≥1 && 자기일관 (mutant guard) | M2 |
| AC-LDS-011 | spec.md §D | MUST | PRESERVE — 템플릿 빈 diff, applier 무변, `.moai/state/lsel/` 외 쓰기 0, CI guard 초록 | M2 |
| AC-LDS-012 | REQ-LDS-009 | SHOULD | 로컬 인도물 문서화(M1)+적용 관측(M2), PR 미포함 | M1/M2 |

## §D.1..§D.12 — Given-When-Then + two-cell

### AC-LDS-001 — wrapper 드레인 (MUST)

- **Given** fixture 인박스(신호 stub N≥10, noise 다수)와 오프셋 0의 state dir, **When** `session_drain.sh --inbox <f> --state-dir <s>` 실행, **Then** exit 0 && drain-offset.json의 offset == fixture 행 수 && clusters.json에 후보 ≥1 && stdout/stderr에 1행 상태(stub 읽음/후보/신규 offset).
- **RED-now**: wrapper 부재 — `test -f .claude/skills/hns-lsel-curator/session_drain.sh` → rc=1 (2026-08-25, worktree `a739d04b4`). 빨간 이유: 대상 자체가 없음 (vacuous 아님 — green path가 이 파일을 만들어 뒤집음).
- **Green path**: M1 — session_drain_test.sh의 drain 경로 `PASS`.

### AC-LDS-002 — 잠금 경쟁 (MUST)

- **Given** 다른 프로세스가 state dir 잠금을 보유(테스트가 배경에서 lock 획득), **When** wrapper 호출, **Then** exit 0 && 드레인 미실행(clusters.json·offset 무변경) && 경쟁 통지 출력.
- **RED-now**: wrapper 부재 (동일 rc=1 관측). 경쟁 경로 자체가 현재 없다 — 현재는 잠금 없이 누구든 동시 실행 가능(드레인은 수동 실행 시에만 항상 단일).
- **Green path**: M1 — 경쟁 fixture 서브테스트 `PASS`.

### AC-LDS-003 — 보존 후 덮어쓰기 (MUST)

- **Given** clusters.json에 후보 1개 이상인 state dir, **When** wrapper 실행, **Then** `.moai/state/lsel/clusters-history/`에 이전 clusters.json 사본(타임스탬프/일련 명명) 존재 && 사본이 원 후보를 포함 && 신규 clusters.json은 drain 결과.
- **RED-now**: 보존 기구 부재 — history 디렉터리 없음(`test -d` rc=1). drain.sh는 no-op 경로(63-76행)조차 clusters.json을 덮어써 기존 후보를 유실시킨다(2026-08-25, drain.sh 판독 + §B.5). 빨린 이유: 유실 방지 코드가 없고 덮어쓰기가 무조건.
- **Green path**: M1 — 보존 서브테스트 `PASS` (no-op 경로에서도 보존됨 — 무조건 호출-전 보존).

### AC-LDS-004 — no-op (MUST)

- **Given** offset == 인박스 행 수인 state dir, **When** wrapper 실행, **Then** exit 0 && drain-offset.json 값 불변 && 에러 아님(상태 1행은 no-op임을 서술).
- **RED-now**: wrapper 부재 (rc=1). 참고: drain.sh 단독 no-op는 offset 불변이지만 clusters.json을 빈 후보로 덮어쓴다 — AC-LDS-003과 함께 뒤집는 대상.
- **Green path**: M1 — no-op 서브테스트 `PASS`.

### AC-LDS-005 — 테스트 하네스 + mutant probe (MUST)

- **Given** `session_drain_test.sh`(mktemp -d fixture, trap cleanup — 기존 drain_test.sh 패턴 계승), **When** `bash .claude/skills/hns-lsel-curator/session_drain_test.sh` 실행, **Then** exit 0 && 5경로(drain/경쟁/보존/no-op/fail-open) 전부 `PASS` && **mutant probe**: "오프셋만 전진시키는 가짜 드레인"을 주입하면 검증 predicate(offset==live && candidates≥1 && 자기일관)가 FAIL로 거부함을 입증.
- **RED-now**: 테스트 파일 부재 (rc=1 관측).
- **Green path**: M1 — 전 경로+probe `PASS`. mutant가 predicate를 통과하면 이 AC는 채택 불능(verification-completeness §2 — probe가 뚫리면 기준 자체가 얕음).

### AC-LDS-006 — fail-open (MUST)

- **Given** 드레인 불가 상태(인박스 부재 또는 state dir 쓰기 불가 등 wrapper 오류 유발 fixture), **When** wrapper가 SessionStart 훅 문맥에서 실행, **Then** exit 0 && stderr에 1행 notice && 세션 시작 차단 없음.
- **RED-now**: wrapper 부재 (rc=1). 기준 근거: coding-standards.md § Advisory-Check Discipline — advisory 실패가 경로를 막으면 안 된다.
- **Green path**: M1 — fail-open 서브테스트 `PASS`.

### AC-LDS-007 — 정지 신호 (MUST)

- **Given** fixture: unread backlog > 임계(기본 25), **When** `backlog_check.sh --inbox <f> --state-dir <s> --threshold 25` 실행, **Then** stderr에 `<system-reminder>`로 unread 수·offset·드레인 지시(교체된 텍스트 — wrapper 지시) 출력 && exit 0. 아래 임계에서는 침묵.
- **Given** (로컬 인도물 적용 후, 유지자 머신) `.claude/settings.local.json` `.hooks.SessionStart`에 wrapper+backlog_check 항목 존재, **When** `jq '.hooks.SessionStart' .claude/settings.local.json` 관측, **Then** 2개 lsel 항목 포함 (관측을 progress.md §E.2에).
- **RED-now**: (a) fixture 스크립트는 존재하고 `backlog_check_test.sh`는 현재도 초록이지만 — **발화 표면이 없다**: live settings.json+settings.local.json hooks에 backlog/lsel grep → 0 matches (2026-08-25, primary). 즉 트리거가 다시 죽어도 아무 표면에 나타나지 않는다(3주 정지가 증명). (b) 배선 부재 동일.
- **Green path**: M1 — fixture 회귀 초록 + 교체 텍스트 반영; M2 — 배선 관측 §E.2 기록.

### AC-LDS-008 — 레시피 자기기술 진실 (MUST)

- **Given** `.claude/workflows/lsel-drain-loop.js`, **When** 헤더·본문 판독, **Then** "레시피가 drain.sh를 실행/수행한다"는 주장 0건 && 본문이 모델 매개 리마인더(출력만)임을 명시 && 내구 트리거로 `session_drain.sh` 참조 ≥1.
- **RED-now**: `grep -c "The recipe runs drain.sh" .claude/workflows/lsel-drain-loop.js` → **1** (12행) — 본문은 45-49행 `console.log`뿐 (2026-08-25, worktree `a739d04b4`). 빨간 이유: 주장≠본문 — 이 문서 허위가 §B.2 직접 원인 서술을 3주간 은폐.
- **Green path**: M1 — `grep -c` → 0 && `grep -c "session_drain"` → ≥1.

### AC-LDS-009 — 죽은 앵커 제거 + SKILL 미러 (MUST)

- **Given** tracked lsel 표면 전체(`.claude/skills/hns-lsel-curator/**`, `.claude/workflows/lsel-drain-loop.js`), **When** `grep -rn "§28" <표면>`, **Then** 0건. **When** SKILL.md 판독, **Then** 내구-운영 섹션(트리거·드레인·검증·로컬 배선 안내) 존재.
- **RED-now**: `grep -rn "§28" .claude/skills/hns-lsel-curator/` → **2건** (backlog_check.sh:6 헤더 주석 + backlog_check.sh:50 리마인더 본문 — iter-1 D1 정정; 종전 1건 기록은 :50만 집계한 과소계상) — live CLAUDE.local.md(35,355B, 활성 유지)에는 §28/LSEL 0회: 앵커가 가리키는 실체 없음 (2026-08-25 재측정).
- **Green path**: M1 — grep 0건 + SKILL.md 섹션 `grep -c` ≥1.

### AC-LDS-010 — live 일괄 드레인 + mutant guard (MUST — 카드 명명 mutant 봉쇄)

- **검증 절차(순서 보장 — iter-1 D2)**: (i) 드레인 **직전** 캡처 — `OFFSET_BEFORE=$(jq -r .offset .moai/state/lsel/drain-offset.json)` && `LIVE=$(wc -l < .moai/lessons-inbox.jsonl | tr -d ' ')` (ii) wrapper로 일괄 드레인 실행 (kickoff 승인 이후만) (iii) **직후** 검증. 근거: 인박스는 계속 자라므로 "현시점 wc -l"이 아니라 **캡처한 `$LIVE`**에 대해 판정하고, 배선 적용 후 다른 세션 시작이 끼어들면 no-op 드레인이 live clusters.json을 `candidates: [], total_read: 0`으로 덮어쓰므로(drain.sh 63-76행) 조건 (2)·(3)은 **`clusters-history/`의 일괄 드레인 archived 사본**으로 판정한다 — wrapper는 no-op 덮어쓰기 전에도 무조건 보존하므로(REQ-LDS-003) 해당 사본이 반드시 존재한다.
- **Given** 유지자 머신의 live 백로그 (plan-phase 기측: offset 629 vs 인박스 4,204행 — M2 실행 직전 재측정, moving target), **When** wrapper로 일괄 드레인 실행, **Then** 동시 충족: (1) `.moai/state/lsel/drain-offset.json`의 offset == 캡처한 `$LIVE` (2) **archived 사본**에서 `jq '.candidates | length'` ≥ 1 (3) **archived 사본**에서 `jq --argjson n "$LIVE" --argjson b "$OFFSET_BEFORE" '.offset_after == $n and .total_read == ($n - $b)'` — 자기일관(읽은 만큼만 전진; `629` 하드코딩 폐지·재측정값 파라미터화 — D2).
- **Mutant probe (fixture)**: 오프셋만 live로 전진시키고 클러스터링을 건너뛰는 구현은 (2)·(3)에서 FAIL — AC-LDS-005의 probe가 이를 기계 입증.
- **RED-now**: `jq` predicate 현재 거짓 — offset 629 ≠ 4,204 (2026-08-25, primary live state). 빨른 이유: 드레인이 21일간 미실행 (트리거 부재, spec.md §B.2) — wrong-reason 아님(본 카드가 정확히 이 전표를 뒤집는다).
- **Green path**: M2 — 캡처→드레인→검증 순서로 3조건 동시 관측을 §E.2에 (명령+출력+측정 시점). fixture probe가 통과 못하는 mutant는 live에서도 금지.

### AC-LDS-011 — PRESERVE·경계 (MUST)

- **Given** 본 SPEC 브랜치 diff (기준 ref는 **명시적으로 `origin/main`에 결속** — 사전 `git fetch origin main`, three-dot 형식 `origin/main...HEAD` = merge-base..HEAD. 스테일 로컬 `main` ref에 대고 재면 무관 파일 8+개(internal/template/** 포함)가 뜨는 false-FAIL이 재현된다 — iter-1 D6), **When** 각 검사 실행, **Then** 동시 충족:
  - `git diff --stat origin/main...HEAD -- internal/template/templates` → 빈 출력 (미러링 0)
  - `git diff origin/main...HEAD -- internal/harness/applier.go internal/harness/curator_dispatch.go` → 빈 출력
  - 드레인 후 `find memory -newer <drain-start-timestamp> -name 'feedback_*' | wc -l` → 0 (M1 불변식 계승, SKILL.md §Verification 3번)
  - wrapper/drain 쓰기 대상이 `.moai/state/lsel/` 하위만 (`clusters-history/` 포함), 인박스 mtime·내용 불변
  - CI guard 3종(template-neutrality-check, lsel-leak-guard, internal_content_leak_test) PR에서 초록
- **RED-now**: 항목들은 "보존" 성격 — 사전 트리에서 이미 초록(템플릿에 lsel 0, applier 무변). 채택 근거는 green path 유지: 본 작업이 실수로 뒤집는 순간 FAIL. (기측: CI guard 오늘 초록 — 리드 브리프.)
- **Green path**: M2 — 5검사 관측을 §E.2에.

### AC-LDS-012 — 로컬 인도물 (SHOULD)

- **Given** spec.md §E의 로컬 인도물 3종, **When** (M1) 문서 확인 / (M2) 유지자 머신 적용, **Then** (M1) 3종이 "PR 미탑재·적용 단계 명시"로 기록됨 && (M2) 적용 관측(jq SessionStart 항목, CLAUDE.local.md 섹션 grep ≥1)을 §E.2에 && PR diff에 settings.local.json·CLAUDE.local.md 0건.
- **RED-now**: 문서화는 plan-phase에서 완료 상태로 반영(spec.md §E 존재 — `grep -c "Local Deliverables" spec.md` ≥1). 적용 관측은 미실행(M2).
- **Green path**: M2 — 적용 관측 기록. SHOULD인 이유: 적용은 유지자 머신 행위라 지연될 수 있으나, 미적용이면 내구 트리거가 실제로 산 상태가 아니므로 M2 완료 보고에 명시적 잔여위험으로 기록한다.

## §E. Edge Cases

- 인박스가 드레인 중간에 적재됨(append-only이므로 안전): wrapper의 offset은 drain.sh가 읽은 시점 tail로 — 차애는 다음 세션 시작이 소진(드레인은 never-blocking).
- wrapper가 잠금 획득 대기 중 긴 세션 시작: 경쟁은 즉시 skip(REQ-LDS-002) — 대기 없음.
- drain.sh가 jq 부재 등으로 실패: fail-open(AC-LDS-006) — 오프셋 전진 없음, 다음 세션이 재시도. **오프셋만 전진하는 부분 실패는 없다**(drain.sh는 offset을 마지막에 쓴다, 124-126행).
- 백로그 0인 저장소(신규 clone): no-op(AC-LDS-004) — 훅이 조용히 통과.
- clusters-history/ 무한 증가: 사본은 작다(현재 9,320B 스케일) — 상한(예: 최근 N개 보존)은 run-phase 구현 재량, SPEC은 보존 의무만 규정.

## §F. Quality Gates / Definition of Done

- AC 12종 판정: 11 MUST 전 PASS + 1 SHOULD 판정 기록 (M2 보고서 행렬).
- 기존 회귀: drain_test.sh + backlog_check_test.sh 초록 유지.
- §E.2 증거: M2 live 항목(AC-LDS-007 배선 관측, AC-LDS-010 3조건, AC-LDS-011 5검사, AC-LDS-012 적용 관측)이 전부 명령+출력+측정 시점으로 기록됨.
- CI: PR에서 guard 3종 + go test 초록 (Go 코드 무변경이므로 회귀 표면은 최소).
- 종료 상태: 드레인이 다음 세션 시작에서 다시 돌고(내구), 안 돌면 다음 세션 시작에 보인다(가시성) — 3주 무음 정지 재발 구조적 차단.
