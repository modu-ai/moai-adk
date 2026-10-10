# SPEC-ROLE-INJECTION-BUDGET-001 — Plan

> 카드 t1617 (3.2-1-5 · P0 · Class C) · 기준 트리 HEAD `2aab5f797` (WT-3-2-1) · 상위 SPEC `SPEC-ALWAYS-LOADED-BUDGET-001` 후속 수리
> 해석 결정: 「역할별 조립 ≤9,000자」 = 조립 합본 판독 (spec.md §A.4). 생산자 상한 4,797 (게이트 함의, §B — 직접 분해 4,376은 startup 참조점). core 예산 **3,946** / 설계 목표 3,800 / REQ-RIB-004 정지 밸브가 구속 후퇴선.
> 완료 기준: 경계 소스 봉투 (결정 d-20261010T072910Z-f1c3, option (a)) — spec.md §A.5 세 줄. 생산자 상한 4,797은 봉투 안 최중 소스(startup)에서 파생.

## §A. 범위

### §A.1 편집 대상 파일

| 축 | 파일 | 성격 |
|---|---|---|
| 규칙 | `.claude/rules/moai/workflow/factory-dispatch.md` + 템플릿 미러 | 역할 core 36영역 18,114/17,793 → 예산 3,946 안으로, 설계 목표 3,800 압축 재작성 (재배치 감사 동반) |
| 규칙 | `.claude/rules/moai/workflow/cross-session-messaging.md` + 템플릿 미러 | 빈 영역 유지 확인 (예산에 최유리) — 내용 변경 없음이 기본, 미러 정합만 |
| 규칙 | `internal/template/templates/.claude/rules/moai/core/hooks-system.md` (미러 원본, 34,482바이트 — 선행 편집) → `.claude/rules/moai/core/hooks-system.md` (paths:-scoped) | :131 두 한도 구분 기재 (REQ-RIB-009) — Template-First 순서로 미러 원본 먼저 |
| 로컬 | `.claude/rules/local/gitflow-lane-protocol.md` | REQ-RIB-010 정합 확인 — 원칙 무편집, 갈림 시에만 지목 문장 수정 (로컬 전용, 미러 없음) |
| 코드 | `internal/hook/role_rules.go` | doc 주석 사다리 중복 제거(:393–413), NOTE 지시문·로캘 경고 문구 통일(:538–544, :263–312), 버전 불일치 감지+안내(InjectionFailed 경로) |
| 코드 | `internal/config/role_markers.go` | 원칙 무편집 (표지·추출기 보존) |
| 테스트 | `internal/hook/role_rules_test.go` (+신규 파일 가능) | 조립 예산 테스트, 스텁 게이트(배치 장소는 internal/template 권장 — §C M1), 불일치 술어 테스트, 로캘 안내 테스트 |
| 테스트 | `internal/template/` 신규 테스트 | `*-core.md` 스텁 게이트 (REQ-RIB-005 — 기계 열거 glob) |
| 원장 | `internal/template/testdata/binding_ledger.json` | 36개 `role-core:` 행 after-text 갱신 (상시 유지보수 계약 — REQ-ALB-015 원장 테스트가 정합 검증) |
| SPEC | 본 SPEC dir `relocation-ledger.md` | 재배치 감사 원장 (M2 산출물) |
| 기록 | `progress.md` | §E.1 계획 신호, §E.2 런 증거, 항상 로드 델타 기록 |

### §A.2 PRESERVE (건드리지 않는 것)

- `roleRulesContextLimit = 10000` — 인상 금지 (Q4).
- `roleRuleSizeGate` 사다리 동작과 `FinalizeSessionStartOutput` 구조 — REQ-ALB-010/009 경로 보존.
- `buildRoleCore` 의 표지 검증 4단(균형·첫 표지 순서·열림-다음-시작·미닫힘) — REQ-RIB-003이 명시 보존.
- `internal/config/role_markers.go` 표지 레지스트리와 `ExtractRoleCoreRegions` — 소비자(hook·template 테스트) 공유 구현.
- `cross-session-messaging.md` 빈 역할 core 영역(포인터 전용) — 내용 변경 없음.
- 스텁 2개의 구속 절 세트 — REQ-RIB-005 는 게이트만 추가, 스텁 재분류 아님.
- 상위 SPEC 본문(`.moai/specs/SPEC-ALWAYS-LOADED-BUDGET-001/*`) — amendment 아님 (spec.md §F).
- 오버플로 로캘 경고의 의미(잘리지 않고 보낸다/파일 경로+2,000자 미리보기) — 문구 통일이 의미 변경이 아님.

## §B. 측정 원장과 재고정

plan 단계 전수 재측정 완료 (커밋 `2aab5f797`, 원문 `.moai/reports/t1617/measurements.md`; F1 수리로 §4 산술 정정):

| 수치 | 값 | 명령 요지 |
|---|---|---|
| 배포본 core | 36영역 join 18,114 UTF-16 (로컬 꼬리 +321 포함) | python3 region split + UTF-16 합산 |
| 템플릿본 core | 36영역 join 17,793 UTF-16 (마지막 영역 1,150 vs 배포 1,471 = 로컬 꼬리 +321뿐) | 〃 |
| 헤더 | lane 123 / leader 125 (뒤 `\n\n` 포함) | 형식 문자열 직접 계수 |
| 포인터 | 128, 결합자 각 2 | 〃 |
| 조립 산식 | 생산자 + `"\n\n"` + 헤더 + core + `"\n\n"` + 포인터 (오버헤드 257, leader 최악) | role_rules.go:102–123·:373–380 코드 인용 |
| 생산자 상한 | **4,797** = 합본 실측 23,166 − 본 트리 역할 블록 18,369; 직접 분해 4,376(startup 참조점)과 차 421 | dispatch 기록 + 코드 산식 |
| core 예산 | **3,946** = 9,000 − 4,797 − 257 (leader 구속); lane 3,948. 설계 목표 3,800 | §B 파생 |
| 스텁 | 8,583 (여유 14.2%) / 8,072 (19.3%) UTF-16, role-core 표지 0, `role-rules-required` 각 1 | 〃 + grep |
| 상위 원장 | `role-core:` 행 정확히 36개, 전부 `workflow/factory-dispatch.md` | json 행수 |
| 배포−템플릿 갈림 | 마지막 영역 꼬리 문장 1개뿐 (sweep 처분) | diff |
| hooks-system 미러 | `internal/template/templates/.claude/rules/moai/core/hooks-system.md` 존재 (34,482바이트) | ls |

재고정 절차: 런 단계에서 규칙 파일을 고치기 전 이 표의 명령을 같은 방법으로 재실행해 수치가 다르면 (흡수 등으로 트리가 움직였으면) AC 의 RED 셀 수치를 재고정하고 그 이유를 progress.md §E.2 에 적는다. AC 가 인용하는 RED 명령은 전부 단일 호출 형태라 재실행 가능하다.

## §C. 마일스톤

### M1 — RED: 테스트·고정물 (규칙·코드 무편집)

1. **조립 예산 테스트** `internal/hook` (package hook — `buildRoleCore` 직접 접근, 조립 공식 중복 방지): 모듈 루트 해석 헬퍼(runtime.Caller 기반)로 템플릿 트리(`internal/template/templates/.claude/rules/moai/workflow/`)와 배포 트리(`.claude/rules/moai/workflow/`)를 각각 읽는다. 레지스트리 역할마다: `buildRoleCore` → 역할 블록(헤더+core+결합자+포인터, role_rules.go:102–123·:373–380 산식) 조립 → `4,797(생산자 상한) + 257(오버헤드) + core ≤ 9,000` 단정, 실패 메시지에 breakdown(생산자/결합자/헤더/core/포인터/합) 출력; 진단 하위 단정 core ≤ 3,946. 상수 옆에 출처 주석(합본 실측 23,166 − 역할 블록 18,369; startup 최중 소스 근거 — spec.md REQ-RIB-002) 필수. 예상 RED: 템플릿 17,793 > 3,946 → `--- FAIL` (이유: 현재 core가 예산 초과 — 옳은 RED).
2. **스텁 게이트** `internal/template` 신규 테스트: 배포·템플릿 규칙 트리에서 `*-core.md` glob 기계 열거 → 각 파일 (a) role-core 표지 0, (b) ≤10,000 UTF-16. 자기 모터 서브테스트: 표지를 심은 임시 스텁 픽스처와 10,000 초과 픽스처로 게이트가 실제로 FAIL 하는 것을 스위트 안에서 관측(§1.1 관측-실패 완료 — 오늘 이미 참인 성질이라 RED-now 대신 이 관측이 채택 증거).
3. **불일치 술어·안내 테스트** `internal/hook`: 술어(같음/다름/판독 불가 3케이스, 픽스처) + 4개 로캘 경고에 `moai update` 안내 존재 단정. 예상 RED: 구현 부재 → 컴파일 실패 or 0히트 단정 FAIL.
4. 런 단계 재고정 절차(§B) 실행 — RED 원문(command+stdout+exit+SHA)을 progress.md §E.2 와 acceptance.md 증거 원장에 기록.

### M2 — GREEN(규칙 축): 역할 core 재작성 + 재배치 감사

1. 36영역 전수 재배치 감사 표 작성(영역 id·before 요지·after 위치 {core-압축|companion §|스텁}·after 요지·이미-동반본-확인). `relocation-ledger.md` 로 커밋. 원칙(§D): 구속 절은 core 압축 재작성(의미 보존) 또는 스텁; 이미 companion 이 본문을 가진 절차 중복·포인터 꼬리만 core에서 정리; 삭제 금지.
2. `factory-dispatch.md` 템플릿본 재작성 → core ≤ 3,800 (예산 3,946 −146 여유). 지압 지점: 영역 내 포인터 꼬리(스텁의 companion 지도와 중복) 제거, 1,300–1,500자 대형 복합절을 항목형 [HARD] 목록으로 압축(절차 본문은 mechanics/gates/cards 가 소유 — 감사가 선행 확인), 인접 동종 영역 병합. 바닥이 3,946을 넘으면 REQ-RIB-004 정지 밸브 — 편집 전 정지·보고.
3. 배포본 동일 재작성 — 로컬 +321 꼬리는 감사 행으로 처리(`worktree-integration.md` § Hoist 본문 확인 후 압축·정리). 두 트리 모두 ≤ 예산.
4. 상위 원장 36행 after-text 갱신 → `internal/template` 원장 테스트 GREEN 확인.
5. gitflow 정합(REQ-RIB-010): 재작성 후 `Isolation` 절 존재 + 이동 금지 [HARD] 생존 확인, 로컬 규칙 지목 문장과 대조 (갈림 시에만 로컬 파일 수정).
6. 템플릿 미러 정합 + 스텁 델타 기록(변동 시): rule-authoring (a)/(b) 비용 문을 progress.md 에 적는다. 예상: 스텁 무변동 또는 소폭.
7. 예산 테스트 GREEN 플립. cross-session-messaging.md 미러·배포 정합 확인 (무변경이 기본).

### M3 — GREEN(코드·문서 축) + 최종 검증

1. `role_rules.go`: doc 주석 사다리 1회화(:393–413), NOTE 지시문 천단위·중복 정리(:538–544), 로캘 경고 중복 서술 정리(:263–312), 버전 불일치 감지(배포 `system.yaml` `template_version` vs 바이너리 버전 — 판독 불가 시 조용히 생략, fail-open) + 상세와 4개 로캘에 `moai update` 안내. M1 테스트 GREEN 플립.
2. `hooks-system.md`: :131 두 한도 구분 기재 (50K 총 stdout→디스크 저장 / 10,000자 문자열당 전달 한도, 각 실측 출처 — Q4·lane-15 공지). **템플릿 미러 편집 선행**: `internal/template/templates/.claude/rules/moai/core/hooks-system.md` (존재 확인됨, 34,482바이트)를 먼저 고치고 배포본에 미러 — REQ-RIB-011의 Template-First 순서.
3. 전체 영향계열 검증: `go test ./internal/hook/... ./internal/template/...` (카드 트리 레인-로컬, env 스크럽, 컴파운드 1호출 — 스텁 게이트 포함 전 Green) + `golangci-lint run --timeout=2m ./internal/hook/... ./internal/template/...`.
4. **라이브 재진입 확인(§A.5 완료 기준의 실측 반쪽 — 경계 소스 봉투)**: 팩토리 레인 세션 재진입 2회 — (a) 새 startup, (b) clear 재진입(핸드오프 대기 없음 — 본 리포 handoff.mode=manual이 기본 참; `factoryLaneRuleForSource` 가 clear 에서 발화하는 소스) — 각 시작 공지에 "역할 규칙 주입 초과" 부재 + `additionalContext` 조립 합본 ≤10,000 관측을 progress.md §E.2 에 기록. 봉투 밖 재주입 세션(핸드오프 본문 클레임 clear · armed-goal compact)의 게이트 가시 발화는 AC-RIB-013(REQ-RIB-012)의 테스트 단언이 담당한다.
5. 크로스플랫폼 빌드 `GOOS=windows GOARCH=amd64 go build ./...` (role_rules_read_windows.go 계열 보존 확인).
6. 스텁 델타 기록 완성(REQ-RIB-011): progress.md §E.2 에 REQ-RIB-011이 고정한 라벨 행으로 상시 파일(스텁) 전후 크기를 UTF-16 코드 단위와 UTF-8 바이트 둘 다 기록 — 바이트 증가가 1,000바이트를 넘으면 rule-authoring 비산 문 첨부. (이 라벨 행이 작성되기 전까지 progress.md 는 라벨을 갖지 않는다 — AC-RIB-009 의 RED E8 이 그 부재를 관측한다.)

### 마일스톤↔AC 대응 (acceptance.md 번호와 정렬 — F2 수리)

| AC | 내용 / 분류 | 플립 마일스톤 |
|---|---|---|
| AC-RIB-001 | 조립 예산 테스트 / release-blocking (E1+E2) | M1 작성 → M2 플립 |
| AC-RIB-002 | 재배치 감사 원장 / release-blocking (E3) | M2 |
| AC-RIB-003 | 스텁 게이트 / regression-guard (모터 관측 채택) | M1 작성(자기 모터 관측) → M2 확정 |
| AC-RIB-004 | InjectionFailed 안내 / release-blocking (E4) | M1 작성 → M3 플립 |
| AC-RIB-005 | 불일치 술어 / release-blocking (E5) | M1 작성 → M3 플립 |
| AC-RIB-006 | 사다리 주석·문구 정리 / release-blocking (E6) | M3 |
| AC-RIB-007 | hooks-system 두 한도 / release-blocking (E7) | M3 (템플릿 미러 선행) |
| AC-RIB-008 | gitflow 포인터 정합 / regression-guard | M2 생존 확인 |
| AC-RIB-009 | 미러 정합 + `stub-delta:` 기록 / release-blocking (E8) | M2 미러 → M3 기록 |
| AC-RIB-010 | 오버플로 경로 보존 / regression-guard | M3 전체 스위트 |
| AC-RIB-011 | 상위 원장 36행 정합 / regression-guard (§2 미결정 처분 — 편집 후 재관측이 채택 증거) | M2 |
| AC-RIB-012 | 라이브 재진입(경계 봉투: startup + clear-핸드오프대기없음) / regression-guard | M3 |
| AC-RIB-013 | 봉투 밖 재주입 세션의 게이트 가시 발화 + 폴백 보존(REQ-RIB-012) / regression-guard | M1 작성(기존 사다리 서브테스트) → M3 봉투 밖 서브테스트 |

## §D. 검증 계획

- 레인-로컬 원칙: 영향계열만(`internal/hook`, `internal/template`), env 스크럽 컴파운드 1호출, `go test ./...` 전체 금지(레인 규율). CI 가 전체를 돌린다.
- 모든 RED 명령은 단일 호출(파이프·`&&`·`;` 체이닝·서브셸 없음) — acceptance.md 증거 원장의 것을 그대로 재실행해 재현한다.
- 최종 턴 종료 리뷰: `codex_review scope=uncommitted` (레인 소관) — card-review.md 는 리더 정산 단계의 것.
- 커밋 규율: 마일스톤별 Conventional Commit, 카드 id t1617 본문 포함, `Authored-By-Agent` 트레일러, 로컬 착지(push 없음 — 리더 3f34).
