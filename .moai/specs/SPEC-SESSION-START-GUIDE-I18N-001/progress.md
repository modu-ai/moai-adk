# progress — SPEC-SESSION-START-GUIDE-I18N-001

card: t1603
spec: SPEC-SESSION-START-GUIDE-I18N-001
phase: plan
updated: 2026-10-08

## 카드·배차 맥락

- 카드 t1603 (운영자 지시 2026-10-08, Class C): 팩토리 리더/레인 SessionStart 안내 문구 다국어 렌더 + `/moai todo --auto-leader`·`--auto-lane` 2모드 설명 + https://adk.mo.ai.kr 문서 안내. 대상: 부트스트랩 주입 텍스트(standing spawn authority 포함)·템플릿 미러 양측.
- dispatch: `wt: .moai/worktrees/t1603`, 브랜치 `WT-session-start-i18n`, base `81786284e`. 직렬 슬롯 임대 거부(t1588 선점) 확인 — 리더 배차 선례 t1498(임대 불가 레인도 리더 배차면 카드 완수, 정산은 리더)로 레인 진행.
- depth-1 준수: 하위 에이전트 스폰 없음. 전체 스위트 미실행(레인-로컬 규율).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-08

- plan-phase 아티팩트 집합 착지 완료 — spec.md(draft)·plan.md·acceptance.md·decision-index.md 4종 + RED 테스트 원본(internal/hook/session_start_guide_i18n_red_test.go).
- [NEEDS CLARIFICATION] 0건 — Q1은 2026-10-08 레인 카드-우위 판정으로 해소(plan.md §D, decision-index.md Q1 RESOLVED).
- plan-audit iteration 1 FAIL 0.81(.moai/reports/t1603/plan-audit-1.md)의 차단 결함 D1·D2·D4 수리 완료: RED-now 셀 4요소 충족(LEDGER-RED-AC001/AC002/AC005 — command·verbatim stdout·exit code 1·tree SHA 81786284e6e5ee2fe5fbe7b498d649b27c3e0109), AC-006 regression-guard 강등(사유 셀 기재), AC-009 신설(REQ-007 매핑, §D.6 매핑표 정정), §E.1 갱신(본 절). D3(codex 게이트 영수증)은 레인 소관.
- plan-audit iteration 2 FAIL 0.94 — D1~D4 해소 독립 재검증 확인(MP-8 PASS, rcpt-fe989264c2d8eb547a8c37ed). 신규 차단 D5 수리: plan.md §F M4의 검증 셀렉터가 본 카드의 RED 테스트를 0건 선택하는 결함 — `go test ./internal/hook -run 'TestRed|Factory'` / `go test ./internal/cli -run 'TestRed|AutoDoc'` 패키지 분할 형태로 교체(TestRed 포함 + 결합 실행의 120초 초과 회피).
- plan-audit iteration 3 D6 수리: cli 측 셀렉터의 `AutoDoc`는 실제 계열 이름이 아니어서 0건 선택 — 실측 교정. `go test ./internal/cli -list 'AutoDoc'` → 0건, `-list 'AutoRank'` → 16건(TestAutoRankDoctrineAmendment·TestAutoRankMirrorParity·TestAutoRankMarkerDisclosure 등, internal/cli/todo_auto_doc_test.go), `-list 'TestRed'`(cli) → 0건. plan.md M4 cli 층을 `go test ./internal/cli -run 'TestRed|AutoRank'`로, acceptance.md AC-007 판정 명령을 `go test ./internal/cli -run 'TestAutoRank'`로 각각 교정하고 실측 산출을 셀에 기록. 교훈: 규약에 의존한 미실측 이름은 D5급 빈-스위프를 낳는다 — 검증 셀렉터는 `-list` 실측 이름만 사용.
- plan-audit iteration 6 D9 수리: RED 테스트 강화(internal/hook/session_start_guide_i18n_red_test.go) — AC-002에 4-로케일 매트릭스 경로 긍정 단정 + en 정준 문장 단정 추가(문장 삭제·빈 로케일 텍스트 변이 포획), AC-005에 설계-노면 마커 `t1600` 리터럴 긍정 단정 추가(마커 없는 플래그 서술 변이 포획). 강화판 재실행 RED: `go test ./internal/hook -run 'TestRedLaneJoinAuthorityLocalized|TestRedAutoModeGuidancePresent' -v -count=1` → exit 1(3행 영어-잔존 + 플래그 2종 누락 + 마커 누락, SHA 81786284e6e5ee2fe5fbe7b498d649b27c3e0109). 변이 3종 사전 실행(`/tmp/t1603-d9/mutation_check.go`, 테스트와 동일 술어): M1 문장 삭제 DETECTED(4-로케일 경로 누락), M2 빈 로케일 텍스트 DETECTED(ko 경로 누락), M3 마커 없는 플래그 텍스트 DETECTED. LEDGER-RED-AC002·AC005 셀을 강화판 출력으로 재기록.
- plan-audit iteration 7 D10 수리: RED 테스트 파일의 커밋 시점 고정 — `internal/hook/session_start_guide_i18n_red_test.go`는 M2 전환 커밋(TestRed 3종 전부 녹색이 되는 첫 커밋)과 함께만 커밋하며, plan-phase·M1 커밋에는 포함하지 않는다(plan.md §D 제약 + M2 셀에 기재). 사유: AC-001·AC-005가 M2까지 적색이므로 그 이전 커밋은 pre-commit 게이트(변경 패키지 테스트 통과)를 위반하고 CI를 적색으로 남긴다. M1 작업 동안 이 파일은 워크트리에 미커밋 상태로 존재한다.
- plan-audit iteration 9 D12 수리: (b) 설계-노면 마커를 로케일 무관 리터럴 `designed surface — t1600, not yet shipped`로 고정 — plan.md M1 셀·acceptance.md AC-005(Then)·테스트 단정 3곳 동일 리터럴(REQ-001 프로토콜 토큰 규율), 맨 카드 번호만 있는 변형은 시험 거부; (a) AC-002 셀에 검증 범위 배정 기재 — 하위 내용 품질(위임 매핑·depth-1 봉인·동료 불변식)은 AC-009 등급 게이트 소관, AC-002는 현지화 존재만 검증(plan 단계에서 미작성 텍스트에 긍정 하위 내용 단정을 박지 않음). 강화판 RED 재실행 exit 1(SHA 81786284e6e5ee2fe5fbe7b498d649b27c3e0109, gofmt 클린). 자기-변이 스윕 13행(기존 12 + D12-b 맨 카드 번호 변형) **전부 DETECTED, 0 undetected** (sweep exit 0). Exit-delta 제출 준비 완료.

## §E.2 Run-phase Evidence

- 측정 기준: 본 워크트리(.moai/worktrees/t1603, 브랜치 WT-session-start-i18n), HEAD `19cd7218c`(plan 종결 커밋) 위 M1 구현 **미커밋 상태**(D10 — RED 테스트 파일은 M2 전환 커밋과 함께; M1 구현 파일은 레인이 M1 경계에서 스테이지). 측정일 2026-10-08.
- M1 변경 파일: `internal/hook/session_start_factory_i18n.go`(+54행 — factoryMessages 필드 4종 신설 `laneSpawnAuthority`·`autoModeGuideLeader`·`autoModeGuideLane`·`docsPointer`, 4-로케일 전 항목 작성), `internal/hook/session_start_factory.go`(리더 안내 신설 블록 (f) + 레인 join 결합점의 테이블 참조 전환 + 파일 머리말 주석 갱신), `internal/hook/session_start.go`(additionalContext 채널 langEnglish → operatorLang(h.cfg) 교체 + 채널 주석 갱신), `internal/hook/lane_spawn_authority.go`(영어 단일 상수 제거 — 설계 결정 문서로 전환, 문구는 테이블로 이항), `internal/hook/session_start_lang.go`(langEnglish 주석 갱신).

### AC-001 (TestRedAgentChannelFollowsConversationLanguage) — PASS

- command: `go test ./internal/hook -run 'TestRed' -v -count=1`
- verbatim 출력(결정 행):
  ```
  --- PASS: TestRedAgentChannelFollowsConversationLanguage (1.29s)
  --- PASS: TestRedLaneJoinAuthorityLocalized (0.00s)
  --- PASS: TestRedAutoModeGuidancePresent (0.00s)
  ok  	github.com/modu-ai/moai-adk/internal/hook	2.329s
  ```
  (동일 3종이 최종 고정 실행에서도 녹색 — 아래 M4 셀렉터 행 참조. RED→GREEN 전환: plan-phase LEDGER-RED 원장의 적색 3종이 본 M1+배선 구현으로 녹색 전환됨.)

### AC-002 (TestRedLaneJoinAuthorityLocalized) — PASS

- 상기 동일 실행의 verbatim 행: `--- PASS: TestRedLaneJoinAuthorityLocalized (0.00s)`. ko·ja·zh join에서 영어 접두 `Standing spawn authority:`와 영어 본문 단편 부재 + 4-로케일 매트릭스 경로 `.claude/rules/moai/development/spec-frontmatter-schema.md` 존재, en 정준 문장 유지 확인.
- 기존 회귀 가드도 녹색: `--- PASS: TestFactoryWorkerNoticeCarriesSpawnAuthority` / `--- PASS: TestLaneSpawnAuthorityFailOpenPreserved` / `--- PASS: TestFactoryWorkerNoticeNamesLabel` / `--- PASS: TestFactoryWorkerNoticeLocaleWordOrders`(동일 -v 실행).

### AC-005 (TestRedAutoModeGuidancePresent) — PASS

- 상기 동일 실행의 verbatim 행: `--- PASS: TestRedAutoModeGuidancePresent (0.00s)`. 리더 안내의 `--auto-leader` 블록 안에 온전한 리터럴 `designed surface — t1600, not yet shipped` 동일 블록 존재 확인.

### M4 레인-로컬 고정 셀렉터 — PASS

- command: `go test ./internal/hook -run 'TestRed|Factory' -count=1` → verbatim: `ok  	github.com/modu-ai/moai-adk/internal/hook	133.093s`
- command: `go test ./internal/cli -run 'TestRed|AutoRank' -count=1` → verbatim: `ok  	github.com/modu-ai/moai-adk/internal/cli	38.863s`
- 비고: plan-audit이 기록한 카드-트리 구조적 적색(환경 신호, t1350/t1542 계열)은 본 실행 두 번(사후 윤문 전 158.408s, 후 133.093s) 모두 재현되지 않았다 — 관측값은 `ok`. 구조적 적색은 재현-의존 환경 신호로 기록을 유지한다.

### 품질 게이트 (TRUST 5 — Unified/Tested/Secured)

- command: `gofmt -l internal/hook/` → verbatim 출력: (빈 출력, rc=0 — gofmt 클린)
- command: `go build ./...` → verbatim: `BUILD_OK`(echo 마커, rc=0)
- command: `GOOS=windows GOARCH=amd64 go build ./...` → verbatim: `windows_build_ok rc=0`
- Secured: 해당 없음(문자열 테이블 + 호출점 1곳 — 신뢰 경계 없음, acceptance §D.4 참조).

### AC-009 원어 윤문 게이트 (moai-domain-humanize 실행 기록)

- 실행: moai-domain-humanize 스킬 로드 → 한국어 모듈(modules/korean.md) 산문 카탈로그 A–J + ja/zh 모듈 대조 검토 대상: 본 M1 신설 ko·ja·zh 산문 전체(spawn authority 번역·2모드 설명 리더/레인 변형·문서 안내 문장).
- 교정 반영(1차 초안에서 발견된 계어·AI 흔적 수정): (a) ko — 의무 대명사 `당신` 제거(house style `이 세션은`), `그래서` 접속 계어 제거, under-후위사 계어 `승인 하나 아래에서` → `승인 하나로`, 접속 종결 뒤 쉼표(C-11) 제거; (b) ja — `すなわち` 재진술 계어 제거, 목적어를 가로막던 이중 대시 삽입구 해체(괄호 보충형으로 전환); (c) zh — `你` 의무 대명사 → `本会话`(house style 정합), 같은 대시 삽입구 해체. 프로토콜 토큰(플래그명·`moai factory next`·매트릭스 경로·URL·설계-노면 리터럴)은 전 로케일에서 원문 유지 확인(REQ-001/REQ-005).
- 잔여 판정: 잔여 S1 0(로케일별). 잔여 S2 — em-dash 부가(J 계열)가 기존 테이블 전체의 기조 밀도와 동일 수준으로 로케일당 ≤2. 등급 판정 행(AC-009 판정 형식, 로케일당 1행):

AC-009-verdict: ko grade A S1 0

AC-009-verdict: ja grade A S1 0

AC-009-verdict: zh grade A S1 0

### Gaps

- M4 신설 필드의 렌더 테스트 확장(4-로케일 전수 비공백·레이아웃 불변식·원문 리터럴 핀)과 AC-006/AC-007/AC-008 판정 테스트는 plan §F-M4의 후속 단계다 — 본 위임은 M1 내용 + additionalContext 배선만 운반하며 세 TestRed 종료가 판정 기준이었다.
- 커밋 미실행: D10 고정과 위임 지시(7번 항목)에 따라 본 위임은 아무것도 커밋하지 않는다. `git status --short` 기준 변경 5파일 + 미커밋 RED 테스트 파일 1파일이 워크트리에 남는다. §E.3(run_commit_sha 포함)은 커밋 착지 후 채움.

### Residual-risk

- ja/zh 산문의 원어성 판정은 스킬 카탈로그 기반 수동 검토다(자동 검출기 미사용 — 스킬 자체의 한계 고지대로). 판정 행의 등급은 이 검토의 관측값이다.
- `설계 노면` 용어는 본 SPEC 아티팩트(plan/spec 본문)가 확립한 프로젝트 내 용어로 ko 산문에 사용했다 — t1600 착지 후 M3 재확인에서 마커·용어 제거 판정과 함께 재검토 대상이다.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-10-08
run_commit_sha: "9f94e9903" (M1-M4 커밋 — D10 고정에 따라 RED 테스트 파일 `internal/hook/session_start_guide_i18n_red_test.go` 를 포함하는 M2 전환 커밋을 겸한다. 선행 plan 커밋 `19cd7218c`)
run_status: audit-ready

- AC 판정: AC-001·AC-002·AC-005 (TestRed 3종) 전부 녹색 — plan-phase LEDGER-RED 원장의 적색이 전환됨. AC-003·AC-004·AC-006·AC-007·AC-008 판정 테스트 M4 착지·녹색 (아래 명령). AC-009 grade A S1 0 ×3 (ko·ja·zh — §E.2 판정 행).
- 레인-로컬 검증(관측값): `go test ./internal/hook -run 'TestRed|Factory|LeaderNotice' -count=1` → `ok  github.com/modu-ai/moai-adk/internal/hook  122.599s` / `go test ./internal/cli -run 'TestRed|AutoRank' -count=1` → `ok  github.com/modu-ai/moai-adk/internal/cli  19.996s`. gofmt -l (internal/hook, internal/cli) → 빈 출력. `go build ./...` rc=0, `GOOS=windows GOARCH=amd64 go build ./...` rc=0. 레인의 포맷 수리 뒤 재측정: hook 계열 117.7s·cli 21.1s 녹색 (레인 관측, §Gaps).
- t1480 계열 수리: TestLeaderNoticeBatchGatePointer/handler_level 의 영어-고정 기대(session_start.go :518 배선이 뒤집은 정책)를 로케일-적합성 단정으로 갱신 — 갱신된 M4 셀렉터 `TestRed|Factory|LeaderNotice` 로 가족 전체 녹색. plan.md §F-M4 셀에 셀렉터 갱신 기록.
- 환경 신호: TestStaleRunNoticeFactoryLegacyLabel 가 전체 계열 부하 하에서 1회 적색("factory messaging degraded: context deadline exceeded") — 단독 실행 녹색(`--- PASS ... (0.69s)`), stale-run 게이트는 본 SPEC 변경 면 밖. 재현-의존 환경 신호로 기록.

Gaps:
- 최종 게이트 실행 도중 429로 이 세션이 중단되었다 — 중단 시점의 gofmt -l 이 render 테스트 파일 1건을 적려했고, 레인이 포맷 수리(commit 9f94e9903 포함)로 닫고 디스크에서 계열 전체 녹색·gofmt 클린·빌드 클린을 재확인했다. 포맷 수리 뒤의 재측정 수치(hook 117.7s / cli 21.1s)는 레인의 관측값이며 본 세션이 직접 관측하지 않았다.

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: complete (3-phase close — 단일 sync 커밋에 `implemented → completed` 전이 탑재)
- sync_commit_sha: "pending-backfill-sync" (커밋은 자신의 해시를 인용할 수 없음 — D3 관례, 다음 커밋에서 backfill)
- sync scope: spec.md frontmatter 상태 전이 + §E.4 시그널 + CHANGELOG [Unreleased] 엔트리. plan/acceptance 본문 무변경(갱신할 `updated:` 필드 부재 실측).
- CHANGELOG: `grep -c 'SPEC-SESSION-START-GUIDE-I18N-001' CHANGELOG.md` → 0 (사전 실측, 중복 없음) → `### Added` 신규 엔트리 1건 발행.
- 구현 요약: factoryMessages 4신규 필드(laneSpawnAuthority·autoModeGuideLeader·autoModeGuideLane·docsPointer) 4-로케일, agent-facing additionalContext 채널 conversation_language 배선(Q1 — two-audience 규칙은 이 면에 한해 카드가 개정), 2모드 설명 블록 + 설계-노면 마커 리터럴 `designed surface — t1600, not yet shipped` + adk.mo.ai.kr 문서 안내, 템플릿 미러 2면 갱신, RED 3종 + 렌더/레이아웃/doc-parity 테스트.
- lane-direct 기록: manager-docs 위임 2회 연속 429 사망(착지 0) → t1495 선례로 레인이 sync 소관 편집을 직접 수행. 편집 내용은 위 범위에 한정.

## 조사 기록 (본 워크트리, base 81786284e, 2026-10-08)

### 방출기 (런타임)

- `internal/hook/session_start_factory_i18n.go:32-72` — factoryMessages 구조체(20 필드). `:77-302` — factoryLocales 4-로케일(en/ko/ja/zh). `:304-312` — factoryMessagesFor 영어 폴백. 산문 전용 테이블·프로토콜 토큰 제외·개행 불변식의 기존 규율 확인.
- `internal/hook/session_start_factory.go:119,132,173,232` — 안내 빌더. `:243` — 레인 join + laneSpawnAuthority 결합점.
- `internal/hook/lane_spawn_authority.go:38-41` — laneSpawnAuthority 영어 단일 상수 + 영어 고정의 근거(두-청중 규칙) 주석. 카드가 뒤집는 대상.
- `internal/hook/session_start.go:518` — additionalContext 채널 langEnglish 하드코딩. `:528` — systemMessage 채널 operatorLang. `:503` 주석은 agent-facing 사본이 agent_prompt_language를 따른다고 서술(코드와 어긋남 — Q1 표적).
- `internal/hook/session_start_lang.go:14-30` — operatorLang, conversation_language 읽고 nil/빈 값 fail-open 영어.

### 설정·로케일

- `.moai/config/sections/language.yaml` — conversation_language: ko(옵션 ko/en/ja/zh/es/fr/de), agent_prompt_language: en("Always en" 주석). 템플릿 원본 `internal/template/templates/.moai/config/sections/language.yaml.tmpl`.

### 템플릿 미러

- 안내 산문의 템플릿 리터럴 사본 없음(훅은 바이너리 내장). 미러 정합면: `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md` § Lane spawn authority (standing) · `factory-dispatch-detail.md:135`("the SessionStart join notice carries the authority sentence verbatim") · `internal/template/templates/.claude/skills/moai/workflows/gtd.md` § --auto.
- 패리티 테스트 선례: `internal/cli/todo_auto_doc_test.go` — live+미러 쌍 원문 리터럴 고정 + 비고정 문구 공백 정규화 비교(SPEC-TODO-AUTO-PRIORITY-001 AC-TAP-011/-013/-014).

### --auto 2모드

- `--auto-leader`/`--auto-lane`: 본 base에 부재(전트리 grep 0적중). 플래그 등록은 `internal/cli/todo.go:323`의 통합 `--auto` 단일. 레인 면 `internal/cli/todo_auto_lane.go`(t1554). 2모드 분리는 형제 카드 t1600(미병합) 소관 — 본 SPEC은 설계 의미론 기준 안내 텍스트 + REQ-005 설계-노면 명시로 운반.

### SPEC ID 사전 점검

- `SPEC-SESSION-START-GUIDE-I18N-001` → Bash 정규식 `PASS`, `.moai/specs/` 중복 0건 실측.

## plan 해소 기록 (2026-10-08 추가)

- Q1 판정(레인 카드-우위 판정, decision-index.md Q1 반영): agent-facing additionalContext 채널(session_start.go:518)은 conversation_language를 따른다. 권위: 카드 본문(운영자 지시 2026-10-08) — 리더·레인 SessionStart 안내 전체의 conversation_language별 렌더 + standing spawn authority 문구의 다국어 대상 명시. two-audience 규칙(lane_spawn_authority.go:38-40)과 `agent_prompt_language: en` 정책은 본 카드가 SessionStart 부트스트랩 안내 면에 한해 개정; 그 외 agent-facing 면에는 `agent_prompt_language`가 계속 지배. 가역적 텍스트 렌더 결정, keep-set 게이트 아님. 레인 결정 기록은 리더가 `.moai/reports/t1603/`에 착지.
- 이에 따라 plan.md §D의 [NEEDS CLARIFICATION] 제거(Q1 RESOLVED로 대체), §F M1/M2의 Q1 대기 조항 해소, spec.md §1에 설계 결정(카드에 의한 규칙 개정) 절 추가, REQ-001에 additionalContext 채널 포함 명시, acceptance.md AC-001에 Q1 판정 반영. AC-001·AC-005의 RED-now 기록은 유지 — 영어 하드코딩은 어느 AC의 생존 조건이 아니다.

## plan-phase 아티팩트 목록

- `.moai/specs/SPEC-SESSION-START-GUIDE-I18N-001/spec.md` (status: draft, Tier M)
- `.moai/specs/SPEC-SESSION-START-GUIDE-I18N-001/plan.md`
- `.moai/specs/SPEC-SESSION-START-GUIDE-I18N-001/acceptance.md`
- `.moai/specs/SPEC-SESSION-START-GUIDE-I18N-001/decision-index.md` (decision_gate: on — interview.yaml:6)
- `.moai/specs/SPEC-SESSION-START-GUIDE-I18N-001/progress.md` (본 파일)
