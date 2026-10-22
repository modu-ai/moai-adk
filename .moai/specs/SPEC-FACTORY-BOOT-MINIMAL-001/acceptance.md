# SPEC-FACTORY-BOOT-MINIMAL-001 — 수용 기준

> 문서 수준 트리 고정: 모든 RED-now 관측의 코드 기준 트리는 `59959629908cda723c37e26c7c15718194d6b38a`다(verification-completeness.md §2.1 — 자체 고정이 없는 기준은 이 SHA에 묶인다). 0.2.0의 재측정은 HEAD `b805ac2d2429f414a7b3634d4196711d688289be`(plan 커밋)의 작업 트리에서 실행했고, 그 트리는 Go 소스·AGENTS.md·목적지 문서가 기준 트리와 바이트 동일하다 — `git diff --name-only 59959629908cda723c37e26c7c15718194d6b38a HEAD`의 출력 전부가 `.moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/` 아래다(SPEC 디렉터리 밖 0건). 원문 출력 파일은 `probe/`에 있으며 §C 증거 장부가 그 파일을 인용한다. Go 측정은 `moai slot acquire --resource go-heavy --max-duration 15m` 임대 아래 1분 부하 31.5에서 한 번에 하나씩 실행했다.

## §A 기준 목록

| AC | 분류 | 요구사항 | 한 줄 요약 | 증거 장부 |
|---|---|---|---|---|
| AC-FBM-001 | release-blocking | REQ-FBM-001, REQ-FBM-002, REQ-FBM-004, REQ-FBM-005 | 탐침: 72칸 고정분 ≤ 2,845·합본 ≤ 9,146, startup 칸의 두 채널 공지가 유지 원소 전부 + 포인터 줄 하나뿐, 유지 원소 문구 지문 불변 | E1 |
| AC-FBM-002 | release-blocking | REQ-FBM-008, REQ-FBM-004 | 경계 변이(영어 레인 고정분 2,846)에서 저장소 내 예산 테스트가 실패하고 위반 칸을 이름으로 보고 | E2 |
| AC-FBM-003 | release-blocking | REQ-FBM-003 | R1~R5: 목적지 두 면에 원소 존재, 제거된 곳에서 표지 부재, 그곳이 목적지 경로를 이름 | E3 |
| AC-FBM-004 | release-blocking | REQ-FBM-010 | AGENTS.md ≤ 24,276바이트 + 두 문단 밖 바이트 동일 + 의무 13개 보존 | E4 |
| AC-FBM-005 | regression-guard | REQ-FBM-010 | AGENTS.md의 배타성 문구 「live only under」 보존 | E5 |
| AC-FBM-006 | regression-guard | REQ-FBM-009, REQ-FBM-010 | Codex 상한 테스트 2개 PASS, skip 없음 | E6 |
| AC-FBM-007 | regression-guard | REQ-FBM-007 | 역할 게이트 규칙 2개(배포·템플릿) 무변경 | E7 |
| AC-FBM-008 | regression-guard | REQ-FBM-011 | 착지 바이너리로 6레인+리더 라이브 합본 ≤ 10,000, 넘침 0, 파싱 오류 0 | E8 |
| AC-FBM-009 | regression-guard | REQ-FBM-006, REQ-FBM-002, REQ-FBM-001, REQ-FBM-009 | 다른 SPEC이 고정한 공지·레인 규칙 테스트 13개가 수정 없이 PASS | E9 |
| AC-FBM-010 | regression-guard | REQ-FBM-003 | 목적지 표지가 기준 트리에 이미 있음 — 「제거 전 착지」의 커밋 그래프 증인 | E10 |
| AC-FBM-011 | release-blocking | REQ-FBM-008, REQ-FBM-004, REQ-FBM-005 | 저장소 내 예산 테스트가 72칸을 정해진 형식으로 보고하고 회계 분해를 출력하며 PASS | E11 |
| AC-FBM-012 | regression-guard | REQ-FBM-009 | `internal/hook`에서 수정·삭제된 파일이 허용 6개 안 | E12 |

regression-guard 분류 사유: AC-005·009·010·012는 보존·순서 단언이라 기준 트리에서 이미 녹색이다(RED-now 없음, 각 항목의 변이 탐침을 함께 둔다). AC-006은 카드가 인용한 CI RED(main@3f81377dc9)가 현재 트리에서 재현되지 않는다(§2.1 undecidable disposition — release 게이트 자격 상실, pass로 기록하지 않음). AC-007은 기준 트리가 범위 시작점이라 측정 대상이 없다. AC-008은 라이브 세션이라 plan 단계에서 재실행 가능한 RED가 아니다(역사적 관측만 있음).

모든 release-blocking 기준의 명령은 한 번의 읽기 전용 호출이다(파이프·리다이렉트·`&&` 없음). 판정은 그 호출의 stdout 원문과 종료 코드로만 한다.

## §B 시나리오 (Given-When-Then)

### AC-FBM-001
- **Given** 기준 트리 이후의 작업 트리와 `probe/overlay.json`
- **When** `go test -count=1 -overlay .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/overlay.json -run '^TestT1923BootBudgetProbe$' -v ./internal/hook/` 를 실행하면
- **Then** exit 0, stdout에 `--- FAIL` 0건, `--- PASS: TestT1923BootBudgetProbe ` (이름 뒤 공백), `PROBE swept=72 channel_checks=48`, `--- PASS: TestT1923BootBudgetProbe/notice_mutants ` (이름 뒤 공백), `MUTANT … rejected=…` 13줄이 있고, `PROBE <칸> fixed=N composite=N core=N` 72줄의 모든 fixed ≤ 2845·composite ≤ 9146이다.
- 탐침이 칸마다 단언하는 것: 고정분 ≤ 2,845, 합본 ≤ 9,146, core ≤ 4,092. startup 칸에서는 additionalContext 사본과 systemMessage 사본(`factoryBootstrapNotice`) 각각에 대해 — (i) REQ-FBM-001/002의 유지 원소 전부가 렌더된 그대로 있다(리더 12개, 레인 3개), (ii) `--auto-leader`·`--auto-lane`·`shipped — t1814`가 없다, (iii) 유지 원소를 걷어 낸 나머지가 정확히 한 줄이고 그 줄이 `.claude/skills/moai/workflows/gtd.md`를 담으며 200 단위 이하다(= 「그 밖에는 없음」), 그리고 (iv) 유지 원소 문구의 지문이 기준 트리 값과 같다.
- RED-now: E1 — exit 1, `--- FAIL` 25줄(실패 하위 테스트 24 + 부모 1). 사유는 둘 다 이 SPEC이 바꿀 렌더 텍스트다: 24개 startup 칸의 두 채널 48건이 모두 `[carries-removed:--auto-leader carries-removed:--auto-lane carries-removed:shipped — t1814 pointer-lacks-path]`(자동 모드 해설이 남아 있고 포인터 줄이 없음), 그리고 영어 레인 고정분 3,751 > 2,845·합본 10,006 > 9,146, 일본어 레인 2,852 > 2,845. 유지 원소 누락·지문 변경은 0건(기준 트리가 지문의 원천).
- Green 경로: M3(메시지 테이블 축소). 통과 출력은 72칸 모두 상한 이하 + 위 PASS 줄. 영어 칸의 실현 가능성은 plan.md §B.7 모의(영어 레인 startup 2,727, 영어 startup 6칸 계약 PASS)로 관측했다.
- 변이 탐침(실행 관측, E1 원문의 `MUTANT` 줄): 자동 모드 문구를 남김 → `carries-removed:--auto-leader`, 자동 모드만 빼고 포인터 줄을 안 넣음 → `residue-lines=0`, 목록에 없는 문장 추가 → `residue-lines=2`, 경로 없는 포인터 → `pointer-lacks-path`, 유지 원소 하나 삭제(배치 게이트 문장 등) → `missing-retained[i]`, 포인터에 산문 덧붙임 → `pointer-too-long`, 유지 원소 문구에 산문 덧붙임 → 지문 변경. 리더·레인 × 4개 언어 모두에서 각 변이가 자기 코드로 거부됐다. role core를 깎아 합본만 맞추는 변이는 AC-FBM-007이 잡는다.

### AC-FBM-002
- **Given** M1~M3 적용 트리와 `probe/breach-overlay.json`(영어 auto 레인 startup 칸을 가장 짧은 생산 형태 기준 정확히 2,846으로 채우는 변이 — `probe/zz_t1923_breach_mutant_test.go.txt`)
- **When** `go test -count=1 -overlay .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/breach-overlay.json -run '^(TestRoleInjectionAssemblyBudget|TestCompositeBudgetRealAssembly)$' -v ./internal/hook/` 를 실행하면
- **Then** exit 1, stdout에 `--- FAIL: TestRoleInjectionAssemblyBudget ` (이름 뒤 공백) 이 있고, `FIXED BREACH role=lane dispatch=auto lang=en src=startup face=deployed fixed=N ceiling=2845` 와 같은 줄(face=template도)이 N ≥ 2846으로 있다.
- RED-now: E2 — exit 0, 두 테스트 모두 `--- PASS`, `FIXED BREACH` 줄 0. 사유: 기준 트리의 영어 레인 고정분은 이미 3,751로 상한 위인데(변이는 채울 필요가 없어 아무것도 하지 않음) 현재 예산 테스트는 합성 생산자 4,797을 쓰고 실조립 테스트는 10,000만 보므로 위반을 보지 못한다 — 올바른 이유의 RED.
- Green 경로: M1이 예산 테스트에 실제 렌더 고정분 측정과 상한 단언을 넣으면 이 변이에서 실패한다. M3 뒤에도 변이는 그 칸을 2,846으로 다시 채우므로 계속 실패해야 한다.
- 변이 탐침: (a) 칸별 수치만 출력하고 단언하지 않는 테스트 → 이 변이에서 exit 0 → 이 기준이 FAIL로 잡는다(codex 지적 반례). (b) 상한을 2,845보다 느슨하게 두는 테스트 → 2,846에서 통과 → 잡힌다. (c) 변이 자체가 위반을 만드는지: plan.md §B.7 모의 트리에 이 변이를 얹으면 영어 레인 startup 칸이 정확히 `fixed=2846`이 되고 상한 2,845 비교가 `factory fixed part 2846 > ceiling 2845`로 실패함을 관측했다(`probe/research/sim-breach-5995962990.txt`).
- 합본 경계(9,147)는 따로 변이하지 않는다: 합본 = 1,950 + 2 + 고정분 + 2 + (core + 255)이므로 core ≤ 4,092(기존 core 상한 단언)와 고정분 ≤ 2,845가 함께 성립하면 합본 ≤ 9,146이 산술로 따른다. 공지 문구 변이로는 core가 움직이지 않아 합본만의 위반을 만들 수 없다(Residual-risk).

### AC-FBM-003
- **Given** M3 적용 트리
- **When** `go test -count=1 -overlay .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/overlay.json -run '^TestT1923RelocationProbe$' -v ./internal/hook/` 를 실행하면
- **Then** exit 0, `--- PASS: TestT1923RelocationProbe ` (이름 뒤 공백), `RELOC swept=48`, `RELOC R1-auto-mode …` ~ `RELOC R5-assigned-card-lease …` 5줄.
- 탐침이 단언하는 것(spec.md §B.3 표 그대로): 재배치 R1~R5 각각에 대해 (i) 목적지 문서가 배포·템플릿 두 면 모두에서 목적지 표지 전부를 담는다(표지 12개 × 2면 = 24), (ii) 빠진 곳(R1: 리더·레인 공지, R2~R4: 레인 자동 규칙, R5: 레인 startup 공지 + 규칙)이 4개 언어 모두에서 사라져야 할 표지를 담지 않는다, (iii) 그곳이 목적지 경로를 저장소 경로 그대로 담는다(24).
- 「제거 전에 목적지에 있다」의 순서: 목적지 내용은 기준 커밋에 이미 있고(AC-FBM-010), 기준 커밋은 M3 커밋의 조상이다 — 커밋 그래프가 순서의 증인이다(verification-claim-integrity.md §2.3). 실행 단계가 목적지에 새 문안을 써야 한다면 그것은 범위 변경이다(plan.md M2).
- RED-now: E3 — exit 1, `RELOC swept=48`, 목적지 단언 24개는 모두 통과(`lacks` 0줄), 제거·경로 단언에서 실패 — R1 리더·레인 공지가 표지를 담고 gtd 경로가 없음, R2 레인 규칙이 `codex_review`·`Merge Risk`·`moai worktree sweep`을 담고 mechanics 경로가 없음, R3 `keep-set`, R4 `jev_ask`(auto-semantics 경로는 이미 있음), R5 `--card`를 담고 gtd 경로가 없음. 사유는 모두 이 SPEC이 바꿀 렌더 텍스트다.
- Green 경로: M3.
- 변이 탐침: 원소를 목적지 없이 그냥 지움 → 그 원소의 목적지는 이미 있으므로 존재 단언은 통과하지만 경로 단언(iii)이 빠진 곳에 경로가 없으면 잡는다. 경로만 적고 원소를 남김 → (ii)가 잡는다. 목적지 문서에서 원소를 지움 → (i)이 잡는다. 필드 이름만 바꾸고 텍스트를 남김 → 표지 기준이라 (ii)가 잡는다.

### AC-FBM-004
- **Given** M5 적용 트리
- **When** `python3 -I .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/agents_md_check.py` 를 실행하면(저장소 루트에서, 인자 없음 = `AGENTS.md`와 기준 SHA)
- **Then** exit 0, 마지막 줄 `result PASS`, `size N ceiling 24276 OK`(N ≤ 24276), `scope prefix identical suffix identical`, `O1`~`O13` 모두 `found`.
- 검사 내용: 크기 ≤ 24,276; 「No main-variant」 문단 앞과 「Three traceability carriers」 문단부터의 바이트가 기준 트리 AGENTS.md와 같음(다른 의무를 지워 바이트를 사는 변이 차단); 두 문단 안에서 의무 13개(plan.md §B.5 O1~O13 — `WT-*main*` 금지, `main` 보유 통합 워크트리 금지, PR 병합으로만 전진, 착지마다 재동기화, 앞서 있으면 결함, primary HEAD detach 금지, 점유 브랜치 거절, `.moai/worktrees/` 아래**에만**, `moai worktree new`로 생성, `.claude/worktrees/`, `agent-*`, 제자리 `WT-` 개명, 생성자 경로 유지)가 현재 문구나 운영자 패치 문구 중 하나로 남아 있음.
- RED-now: E4 — exit 1, `size 24501 ceiling 24276 OVER`, `scope prefix identical suffix identical`, O1~O13 `found`, `result FAIL`. 사유: 여유 75바이트(< 300) — 이 SPEC이 바꿀 크기 하나뿐.
- Green 경로: M5 — 운영자 패치 문단 + `only` 복원 사본(`cand.md`)이 `size 24221 … OK`, `result PASS`, exit 0(`probe/research/agents-mutants-5995962990.txt`).
- 변이 탐침(실행 관측, 같은 파일): `only` 누락 24,218바이트 → `O8 MISSING` exit 1; main-variant 문단 삭제 23,888바이트 → `scope MISSING-ANCHOR` exit 1; 무관한 의무 문단 삭제 23,872바이트 → `scope prefix CHANGED` exit 1; 두 문단을 제목만 남기고 비움 23,625바이트 → O1~O13 `MISSING` exit 1. 네 변이 모두 크기 기준만으로는 통과했을 사본이다.

### AC-FBM-005
- **When** `grep -c "live only under" AGENTS.md` → **Then** 출력 `1`, exit 0 (기준 트리에서도 `1` — 보존 단언). AC-FBM-004의 O8과 같은 의무를 한 줄 명령으로 다시 본다.

### AC-FBM-006
- **When** `go test -count=1 -run '^(TestCodexContractByteCeiling|TestCodexNestedTemplateDiscoveryBudget)$' -v ./internal/config/` → **Then** exit 0, 두 테스트 `--- PASS`(SKIP 아님), `contract document AGENTS.md = <≤24276> bytes (ceiling 24576, ...)`. `CodexContractByteCeiling` 값 24576 무변경.

### AC-FBM-007
- **When** `git log --oneline 59959629908cda723c37e26c7c15718194d6b38a..HEAD -- .claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/cross-session-messaging.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md` → **Then** 카드 t1923 커밋이 0건(main 흡수 전 카드 브랜치에서 판정; 흡수 뒤에는 다른 카드 커밋이 나타날 수 있으므로 커밋 메시지의 카드 id로 귀속을 확인).

### AC-FBM-008
- **Given** 착지 바이너리가 설치되고 6레인 + 리더가 startup으로 기동된 뒤
- **When** `python3 -I .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/live_scan.py <claude projects dir> <설치 시각 ISO>` 를 실행하면
- **Then** 행 ≥ 7(lane ≥ 6, leader ≥ 1), 모든 행의 합본 ≤ 10000, `overflow 0`, 마지막 줄 `errors 0`.
- 0.2.0 수정: 스캐너는 이제 세션 파일 전부를 읽고(앞 16줄 제한 제거), 읽기·파싱에 실패한 파일·레코드를 `errors N`으로 센다 — 넘친 세션이 조용히 빠지는 반례(codex·감사 D7)를 막는다. 합성 입력 관측(scratch, 커밋하지 않음): 17번째 줄에 10,668단위 넘침 레코드를 둔 파일과 깨진 JSON 레코드 파일 하나로 실행하면 `2026-10-23T00:00:00Z lane 10668 10530 138 True p/late.jsonl` / `count 1` / `overflow 1` / `errors 1` — 0.1.0 스캐너가 놓치던 두 경우가 모두 보인다.
- 역사 관측(재실행 RED 아님): E8 — 10-21 기준 `count 15` / `overflow 10`(모두 05:42Z 이전, 역할 블록 18,367). 이 관측은 16줄 제한이 있던 0.1.0 스캐너로 얻었다.

### AC-FBM-009
- **When** `go test -count=1 -run '^(TestSD_AC019_NextCardRuleInjection|TestLaneRuleCarriesStandingRecheckCron|TestFactoryWorkerNoticeCarriesSpawnAuthority|TestInjectedCoreCarriesObligationElements|TestLeaderNoticeBatchGatePointer|TestFactoryLeadNoticeCarriesDispatchDiscipline|TestFactoryLeadNoticeDispatchDisciplineKorean|TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide|TestFactoryGuideTeachesLaneFormsInEveryLocale|TestLaneBootstrapCarriesDispatchMode|TestFactoryGuideDocsPointerInBothNotices|TestFactoryWorkerNoticeLocaleWordOrders|TestFactoryGuideStaleRunStaysEnglishOnAgentChannel)$' -v ./internal/hook/` → **Then** exit 0, 13개 테스트 모두 `--- PASS`(SKIP 아님, 13개 이름 전부 출력에 나타남 — 빈 순회 아님), 그리고 AC-FBM-012가 이 테스트들의 파일이 수정되지 않았음을 보인다.
- 이 기준은 다른 완료 SPEC의 고정 계약 보존이다: 배치 게이트 문장(SPEC-AUTONOMY-BATCH-GATE-001 REQ-BGS-016/017), t85 디스패치 규율, 레인 규칙 토큰·문구(REQ-SD-019, SPEC-WORKFLOW-TASKS-001, SPEC-FACTORY-LANE-NO-CLEAR-001), 스폰 권한, docs 문장.
- 변이 탐침: 리더 공지에서 배치 게이트 문장을 빼는 변이 → `TestLeaderNoticeBatchGatePointer`가 실패. 테스트 쪽을 고쳐 통과시키는 변이 → AC-FBM-012가 잡는다.

### AC-FBM-010
- **When** `git grep -c -F -e '--auto-leader' -e '--auto-lane' -e 'stays for compatibility' -e 'moai factory next --card <id>' -e "## The lane's standard landing" -e 'codex_review' -e 'Merge Risk' -e 'moai worktree sweep' -e "## The leader's remaining role" -e 'Keep-set gate relay' -e '## 6. The decision ladder' -e 'jev_ask' 59959629908cda723c37e26c7c15718194d6b38a -- .claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/skills/moai/workflows/gtd.md .claude/rules/moai/workflow/factory-dispatch-mechanics.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md .claude/rules/moai/workflow/auto-semantics.md internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md` → **Then** exit 0, 여섯 파일 모두 출력에 나타나고 각 줄 수가 E10과 같다(gtd 4, mechanics 6, auto-semantics 5 — 두 면 동일).
- 이 명령은 고정 SHA를 읽으므로 언제 실행해도 같은 답을 낸다 — 목적지가 제거 이전에 있었다는 사실의 증인이며, HEAD에서 목적지가 계속 원소를 담는지는 AC-FBM-003 (i)이 본다.

### AC-FBM-011
- **Given** M1~M3 적용 트리
- **When** `go test -count=1 -run '^(TestRoleInjectionAssemblyBudget|TestCompositeBudgetRealAssembly)$' -v ./internal/hook/` 를 실행하면
- **Then** exit 0, 두 테스트 `--- PASS`(SKIP 아님), 그리고 stdout에 다음이 모두 있다:
  - 정규식 `FIXED role=(leader|lane) dispatch=(-|auto|manual) lang=(en|ko|ja|zh) src=(startup|clear|compact) face=(deployed|template) fixed=[0-9]+ composite=[0-9]+` 에 맞는 줄 정확히 72개(리더는 `dispatch=-`, 레인은 auto·manual), 모든 fixed ≤ 2845·composite ≤ 9146
  - `FIXED swept=72`
  - `FIXED accounting: producer 4797 = allowance 1950 + joiner 2 + fixed ceiling 2845`
- RED-now: E11 — exit 0, 두 테스트 PASS, 그러나 `FIXED` 줄 0개(형식 줄·순회 줄·회계 줄 모두 없음). 기준이 요구하는 출력이 없으므로 기준 자체는 거짓이다. 출력 줄만 추가하는 변이로 이 기준이 녹색이 될 수 있다는 약점은 AC-FBM-002가 막는다 — 두 기준은 짝이다(정상 트리에서 72칸 보고 + 위반 트리에서 실패).
- Green 경로: M1(테스트 재도출) 직후 기준 렌더에서 영어·일본어 레인 칸 위반으로 RED(E8 run 단계 증거), M3에서 녹색.
- 변이 탐침: 순회가 일부 칸을 건너뜀 → 72줄·`swept=72` 단언이 잡는다. 칸 이름만 출력하고 측정하지 않음 → 수치가 실제 렌더와 다르면 AC-FBM-001 탐침의 같은 칸 수치와 대조해 드러난다(Residual-risk).

### AC-FBM-012
- **When** `git diff --name-only --diff-filter=MD 59959629908cda723c37e26c7c15718194d6b38a HEAD -- internal/hook` → **Then** exit 0, 출력의 모든 줄이 다음 여섯 중 하나다: `internal/hook/session_start_factory.go`, `internal/hook/session_start_factory_i18n.go`, `internal/hook/role_injection_budget_test.go`, `internal/hook/role_composite_e2e_test.go`, `internal/hook/session_start_guide_i18n_red_test.go`, `internal/hook/session_start_guide_i18n_render_test.go`. 새 파일 추가(A)는 허용한다.
- 기준 트리 관측: E12 — 출력 없음, exit 0.
- 변이 탐침: t85·배치 게이트·레인 규칙 고정 테스트를 고쳐 통과시킴 → 그 파일 이름이 출력에 나타나 실패. 고정 테스트를 지움 → `--diff-filter=D`가 같은 방식으로 잡는다. 허용된 `session_start_guide_i18n_render_test.go` 안의 보존 테스트 3개를 고치는 변이는 이 명령이 보지 못한다 — AC-FBM-009가 그 3개의 PASS를, sync 감사가 diff를 본다(Residual-risk).

## §C 증거 장부

### E1 — AC-FBM-001 RED-now
```text
command: go test -count=1 -overlay .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/overlay.json -run '^TestT1923BootBudgetProbe$' -v ./internal/hook/
exit: 1
tree: 59959629908cda723c37e26c7c15718194d6b38a (measured at HEAD b805ac2d24, code-identical)
stdout (raw file): .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/red-now-5995962990-probe.txt
excerpt (verbatim lines from that file):
    zz_t1923_boot_probe_test.go:210: operator channel notice violates the contract: [carries-removed:--auto-leader carries-removed:--auto-lane carries-removed:shipped — t1814 pointer-lacks-path]
    zz_t1923_boot_probe_test.go:210: agent channel notice violates the contract: [carries-removed:--auto-leader carries-removed:--auto-lane carries-removed:shipped — t1814 pointer-lacks-path]
    zz_t1923_boot_probe_test.go:191: factory fixed part 3751 > ceiling 2845
    zz_t1923_boot_probe_test.go:194: composite 10006 > budget 9146
    zz_t1923_boot_probe_test.go:269: PROBE swept=72 channel_checks=48
    --- PASS: TestT1923BootBudgetProbe/notice_mutants (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/hook	4.004s
counts on the raw file: "--- FAIL" 25 (24 failing subtests + 1 parent; "^--- FAIL" 1);
  "violates the contract" 48, all with the identical code list above;
  "factory fixed part 3751 > ceiling 2845" 2, "factory fixed part 2852 > ceiling 2845" 2, "composite 10006 > budget 9146" 2;
  "retained elements" 0; "MUTANT " 13.
selected cells: PROBE leader//en/startup/deployed fixed=2573 composite=8830 core=4048
                PROBE lane//en/startup/deployed fixed=3751 composite=10006 core=4048
                PROBE lane//ja/startup/deployed fixed=2852 composite=9107 core=4048
```

### E2 — AC-FBM-002 RED-now
```text
command: go test -count=1 -overlay .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/breach-overlay.json -run '^(TestRoleInjectionAssemblyBudget|TestCompositeBudgetRealAssembly)$' -v ./internal/hook/
exit: 0
tree: 59959629908cda723c37e26c7c15718194d6b38a (measured at HEAD b805ac2d24, code-identical)
stdout (raw file): .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/red-now-5995962990-breach.txt
excerpt:
    role_composite_e2e_test.go:90: RESULT lane_en_glm: composite utf16=9946 (cap 10000, margin=54)
    role_injection_budget_test.go:163: deployed tree: role factory-lane composite=9100 (budget 9146) core=4048 (ceiling 4092, target 3800)
--- PASS: TestCompositeBudgetRealAssembly (0.45s)
--- PASS: TestRoleInjectionAssemblyBudget (0.02s)
ok  	github.com/modu-ai/moai-adk/internal/hook	1.519s
red reason: exit 0 and no "FIXED BREACH" line although the English lane fixed part of this tree is 3,751 (E1) — the in-tree tests do not measure the factory fixed part.
supporting observation (mutant is a breach): .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/research/sim-breach-5995962990.txt
    zz_t1923_boot_probe_test.go:189: PROBE lane//en/startup/deployed fixed=2846 composite=9101 core=4048
    zz_t1923_boot_probe_test.go:191: factory fixed part 2846 > ceiling 2845
```

### E3 — AC-FBM-003 RED-now
```text
command: go test -count=1 -overlay .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/overlay.json -run '^TestT1923RelocationProbe$' -v ./internal/hook/
exit: 1
tree: 59959629908cda723c37e26c7c15718194d6b38a (measured at HEAD b805ac2d24, code-identical)
stdout (raw file): .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/red-now-5995962990-reloc.txt
excerpt (verbatim):
    zz_t1923_boot_probe_test.go:334: R1-auto-mode leader-notice/en still carries removed element marker "--auto-leader"
    zz_t1923_boot_probe_test.go:338: R1-auto-mode leader-notice/en does not name the destination path .claude/skills/moai/workflows/gtd.md
    zz_t1923_boot_probe_test.go:334: R2-landing lane-rule/en still carries removed element marker "codex_review"
    zz_t1923_boot_probe_test.go:338: R2-landing lane-rule/en does not name the destination path .claude/rules/moai/workflow/factory-dispatch-mechanics.md
    zz_t1923_boot_probe_test.go:334: R3-leader-duties lane-rule/en still carries removed element marker "keep-set"
    zz_t1923_boot_probe_test.go:334: R4-judgment-ladder lane-rule/en still carries removed element marker "jev_ask"
    zz_t1923_boot_probe_test.go:334: R5-assigned-card-lease lane-startup/en still carries removed element marker "--card"
    zz_t1923_boot_probe_test.go:338: R5-assigned-card-lease lane-startup/en does not name the destination path .claude/skills/moai/workflows/gtd.md
    zz_t1923_boot_probe_test.go:348: RELOC swept=48
--- FAIL: TestT1923RelocationProbe (1.14s)
FAIL	github.com/modu-ai/moai-adk/internal/hook	1.976s
counts on the raw file: "lacks" 0 (all 24 destination assertions pass); per relocation, the failing lines across 4 locales:
  R1 leader-notice 12 carries + 4 path, lane-notice 12 carries + 4 path; R2 12 carries + 4 path; R3 4 carries + 4 path;
  R4 4 carries (path already present); R5 4 carries + 4 path.
```

### E4 — AC-FBM-004 RED-now
```text
command: python3 -I .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/agents_md_check.py
exit: 1
tree: 59959629908cda723c37e26c7c15718194d6b38a (AGENTS.md byte-identical at HEAD b805ac2d24)
stdout (raw file): .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/red-now-5995962990-agents.txt
first lines (verbatim):
size 24501 ceiling 24276 OVER
scope prefix identical suffix identical
O1 found no WT-*main* branch
last line: result FAIL   (O1-O13 all "found")
green path + mutants: .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/research/agents-mutants-5995962990.txt
```

### E5 — AC-FBM-005 기준 관측
```text
command: grep -c "live only under" AGENTS.md
exit: 0
stdout: 1
tree: 59959629908cda723c37e26c7c15718194d6b38a
```

### E6 — AC-FBM-006 기준 관측
```text
command: go test -count=1 -run '^(TestCodexContractByteCeiling|TestCodexNestedTemplateDiscoveryBudget)$' -v ./internal/config/
exit: 0
tree: 59959629908cda723c37e26c7c15718194d6b38a (re-run 0.2.0 at HEAD b805ac2d24, code- and AGENTS.md-identical)
stdout (raw file): .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/red-now-5995962990-ceiling.txt
excerpt (verbatim):
    token_budget_guard_test.go:109: contract document AGENTS.md = 24501 bytes (ceiling 24576, headroom 75)
    token_budget_guard_test.go:109: contract document internal/template/templates/AGENTS.md.tmpl = 23784 bytes (ceiling 24576, headroom 792)
--- PASS: TestCodexContractByteCeiling (0.04s)
--- PASS: TestCodexNestedTemplateDiscoveryBudget (0.10s)
ok  	github.com/modu-ai/moai-adk/internal/config	0.494s
```

### E7 — AC-FBM-007
기준 트리 자체가 범위 시작점이라 plan 단계 관측 대상 없음(빈 범위). run 단계에서 측정.

### E8 — AC-FBM-008 역사 관측
```text
command: python3 -I .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/live_scan.py /Users/goos/.moai/claude-profiles/moai-adk/projects 2026-10-21T00:00:00Z
exit: 0
scanner version: 0.1.0 (first-16-lines cap, silent exceptions; no "errors" line)
stdout (raw file): .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/research/live-scan-20261021.txt
last lines:
2026-10-22T18:26:15.251Z leader 5762 1467 4295 False -Users-goos-MoAI-moai-adk-go/3936623d-62a7-4ce8-b44e-74ba0ebe94b8.json
count 15
overflow 10
```

### E9 — AC-FBM-009
plan 단계 미실행(§D Gaps). 0.1.0의 4개 테스트는 plan-audit 1회차가 기준 트리에서 실행해 4/4 PASS를 관측했다(`.moai/reports/t1923/plan-audit.md`). 0.2.0이 더한 9개는 run 단계 E1 매트릭스에서 관측.

### E10 — AC-FBM-010 기준 관측
```text
command: (AC-FBM-010의 git grep 명령 그대로)
exit: 0
tree: 59959629908cda723c37e26c7c15718194d6b38a (read from the object store at that SHA)
stdout (verbatim):
59959629908cda723c37e26c7c15718194d6b38a:.claude/rules/moai/workflow/auto-semantics.md:5
59959629908cda723c37e26c7c15718194d6b38a:.claude/rules/moai/workflow/factory-dispatch-mechanics.md:6
59959629908cda723c37e26c7c15718194d6b38a:.claude/skills/moai/workflows/gtd.md:4
59959629908cda723c37e26c7c15718194d6b38a:internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md:5
59959629908cda723c37e26c7c15718194d6b38a:internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md:6
59959629908cda723c37e26c7c15718194d6b38a:internal/template/templates/.claude/skills/moai/workflows/gtd.md:4
```

### E11 — AC-FBM-011 RED-now
```text
command: go test -count=1 -run '^(TestRoleInjectionAssemblyBudget|TestCompositeBudgetRealAssembly)$' -v ./internal/hook/
exit: 0
tree: 59959629908cda723c37e26c7c15718194d6b38a (measured at HEAD b805ac2d24, code-identical)
stdout (raw file): .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/red-now-5995962990-intree.txt
excerpt:
    role_composite_e2e_test.go:90: RESULT lane_en_glm: composite utf16=9946 (cap 10000, margin=54)
    role_injection_budget_test.go:163: deployed tree: role factory-lane composite=9100 (budget 9146) core=4048 (ceiling 4092, target 3800)
--- PASS: TestRoleInjectionAssemblyBudget (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/hook	1.305s
red reason: "FIXED" occurs 0 times on the raw file — no per-cell line, no swept line, no accounting line.
```

### E12 — AC-FBM-012 기준 관측
```text
command: git diff --name-only --diff-filter=MD 59959629908cda723c37e26c7c15718194d6b38a HEAD -- internal/hook
exit: 0
stdout: (empty)
tree: HEAD b805ac2d24
```

## §D Gaps / Residual-risk

- Gaps: AC-FBM-009의 13개 테스트는 plan 단계에서 실행하지 않았다(0.1.0의 4개는 감사가 관측). 녹색 경로 모의(plan.md §B.7)는 영어 테이블만 바꿨다 — ko·ja·zh의 녹색 경로는 Python 길이 추정(커밋하지 않은 scratch)만 있고 실행 관측이 없다. 리더 수동 디스패치·Codex(gpt) owned-card 규칙 변형은 탐침에 없다(gpt 레인 규칙은 owned-card 경로로 lane 자동 규칙보다 짧다 — 측정 안 함). AC-FBM-011의 「칸 수치가 실제 렌더인가」는 AC-FBM-001 탐침 수치와의 대조로만 드러나며 자동 대조는 없다.
- Residual-risk: 비-팩토리 허용분 1,950은 GLM 레인의 단일 라이브 세션 분해 기반이다. 다른 생산자 조합(예: degraded 알림이 여러 개 겹치는 경우)은 1,950을 넘을 수 있다. 봉투 밖 생산자는 여전히 넘침을 만들 수 있으며 폴백이 백스톱이다. 합본 상한만의 위반(core가 4,092를 넘지 않는데 합본이 넘는 경우)은 산술상 생기지 않으므로 별도 변이를 두지 않았다. AC-FBM-002 변이는 레인 자동 규칙의 현재 조립 구조(다음 카드 규칙 + `\n\n` + 정지 복구 규칙)에 기댄다. 탐침의 유지 원소 지문은 기준 트리 렌더 문구에 묶여 있으므로, 유지 원소의 정당한 수정(다른 카드의 문구 변경)이 착지 전에 들어오면 지문을 다시 측정해야 한다.

## §E 완료 정의

- release-blocking 5개(AC-FBM-001·002·003·004·011)가 Then 그대로 관측됨, regression-guard 7개(AC-FBM-005~010, 012) 관측 기록.
- run 단계 E8(RED 출력)에 M1 직후의 예산 테스트 실패 원문 포함.
- sync 단계에서 AC-FBM-008 라이브 확인 기록. **AC-FBM-008 관측이 `overflow` ≥ 1, `errors` ≥ 1, 또는 행 수 미달(lane < 6 또는 leader < 1)이면 카드는 sync를 떠나지 않는다** — 넘친 세션의 분해(plan.md §B.3 방식)를 progress 기록에 남기고, 원인이 이 SPEC의 고정분이면 run으로 되돌리고 봉투 밖 생산자면 별도 카드로 분리한다.
