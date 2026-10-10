# SPEC-ROLE-INJECTION-BUDGET-001 — Plan

> 카드 t1617 (3.2-1-5 · P0 · Class C) · 기준 트리 HEAD `2aab5f797` (WT-3-2-1) · 상위 SPEC `SPEC-ALWAYS-LOADED-BUDGET-001` 후속 수리
> 해석 결정: 「역할별 조립 ≤9,000자」 = 조립 합본 판독 (spec.md §A.4). 생산자 기준선 4,376 (lane-15 직접 분해, §B). core 예산 4,367 / 설계 목표 4,000.

## §A. 범위

### §A.1 편집 대상 파일

| 축 | 파일 | 성격 |
|---|---|---|
| 규칙 | `.claude/rules/moai/workflow/factory-dispatch.md` + 템플릿 미러 | 역할 core 36영역 18,114/17,793 → ≤4,000 압축 재작성 (재배치 감사 동반) |
| 규칙 | `.claude/rules/moai/workflow/cross-session-messaging.md` + 템플릿 미러 | 빈 영역 유지 확인 (예산에 최유리) — 내용 변경 없음이 기본, 미러 정합만 |
| 규칙 | `.claude/rules/moai/core/hooks-system.md` (paths:-scoped, 미러 확인 후 편집) | :131 두 한도 구분 기재 (REQ-RIB-009) |
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

plan 단계 전수 재측정 완료 (커밋 `2aab5f797`, 원문 `.moai/reports/t1617/measurements.md`):

| 수치 | 값 | 명령 요지 |
|---|---|---|
| 배포본 core | 36영역 join 18,114 UTF-16 | python3 region split + UTF-16 합산 |
| 템플릿본 core | 36영역 join 17,793 UTF-16 (마지막 영역 1,150 vs 배포 1,471 = 로컬 꼬리 +321) | 〃 |
| 헤더 | lane 123 / leader 125 (뒤 `\n\n` 포함) | 형식 문자열 직접 계수 |
| 포인터 | 128, 결합자 2 | 〃 |
| 스텁 | 8,583 / 8,072 UTF-16, role-core 표지 0, `role-rules-required` 각 1 | 〃 + grep |
| 상위 원장 | `role-core:` 행 정확히 36개, 전부 `workflow/factory-dispatch.md` | json 행수 |
| 배포−템플릿 갈림 | 마지막 영역 꼬리 문장 1개뿐 (sweep 처분) | diff |
| 생산자 기준선 | 4,376 (lane-15 직접 분해, dispatch 귀속 — 라이브 자산 재측정 불가; 합본 교차 검산 23,066 vs 23,166, ±100) | dispatch 기록 + 본 트리 산술 |

재고정 절차: 런 단계에서 규칙 파일을 고치기 전 이 표의 명령을 같은 방법으로 재실행해 수치가 다르면 (흡수 등으로 트리가 움직였으면) AC 의 RED 셀 수치를 재고정하고 그 이유를 progress.md §E.2 에 적는다. AC 가 인용하는 RED 명령은 전부 단일 호출 형태라 재실행 가능하다.

## §C. 마일스톤

### M1 — RED: 테스트·고정물 (규칙·코드 무편집)

1. **조립 예산 테스트** `internal/hook` (package hook — `buildRoleCore` 직접 접근, 조립 공식 중복 방지): 모듈 루트 해석 헬퍼(runtime.Caller 기반)로 템플릿 트리(`internal/template/templates/.claude/rules/moai/workflow/`)와 배포 트리(`.claude/rules/moai/workflow/`)를 각각 읽는다. 레지스트리 역할마다: `buildRoleCore` → 역할 블록(헤더+core+결합자+포인터) 조립 → `4,376 + 2 + 블록 ≤ 9,000` 단정, 실패 메시지에 breakdown(생산자/결합자/헤더/core/포인터/합) 출력; 진단 하위 단정 core ≤ 4,367. 예상 RED: 템플릿 17,793 > 4,367 → `--- FAIL` (이유: 현재 core가 예산 초과 — 옳은 RED).
2. **스텁 게이트** `internal/template` 신규 테스트: 배포·템플릿 규칙 트리에서 `*-core.md` glob 기계 열거 → 각 파일 (a) role-core 표지 0, (b) ≤10,000 UTF-16. 자기 모터 서브테스트: 표지를 심은 임시 스텁 픽스처와 10,000 초과 픽스처로 게이트가 실제로 FAIL 하는 것을 스위트 안에서 관측(§1.1 관측-실패 완료 — 오늘 이미 참인 성질이라 RED-now 대신 이 관측이 채택 증거).
3. **불일치 술어·안내 테스트** `internal/hook`: 술어(같음/다름/판독 불가 3케이스, 픽스처) + 4개 로캘 경고에 `moai update` 안내 존재 단정. 예상 RED: 구현 부재 → 컴파일 실패 or 0히트 단정 FAIL.
4. 런 단계 재고정 절차(§B) 실행 — RED 원문(command+stdout+exit+SHA)을 progress.md §E.2 와 acceptance.md 증거 원장에 기록.

### M2 — GREEN(규칙 축): 역할 core 재작성 + 재배치 감사

1. 36영역 전수 재배치 감사 표 작성(영역 id·before 요지·after 위치 {core-압축|companion §|스텁}·after 요지·이미-동반본-확인). `relocation-ledger.md` 로 커밋. 원칙(§D): 구속 절은 core 압축 재작성(의미 보존) 또는 스텁; 이미 companion 이 본문을 가진 절차 중복·포인터 꼬리만 core에서 정리; 삭제 금지.
2. `factory-dispatch.md` 템플릿본 재작성 → core ≤ 4,000 (예산 4,367 −367 여유). 지압 지점: 영역 내 포인터 꼬리(스텁의 companion 지도와 중복) 제거, 1,300–1,500자 대형 복합절을 항목형 [HARD] 목록으로 압축(절차 본문은 mechanics/gates/cards 가 소유 — 감사가 선행 확인), 인접 동종 영역 병합.
3. 배포본 동일 재작성 — 로컬 +321 꼬리는 감사 행으로 처리(`worktree-integration.md` § Hoist 본문 확인 후 압축·정리). 두 트리 모두 ≤ 예산.
4. 상위 원장 36행 after-text 갱신 → `internal/template` 원장 테스트 GREEN 확인.
5. gitflow 정합(REQ-RIB-010): 재작성 후 `Isolation` 절 존재 + 이동 금지 [HARD] 생존 확인, 로컬 규칙 지목 문장과 대조 (갈림 시에만 로컬 파일 수정).
6. 템플릿 미러 정합 + 스텁 델타 기록(변동 시): rule-authoring (a)/(b) 비용 문을 progress.md 에 적는다. 예상: 스텁 무변동 또는 소폭.
7. 예산 테스트 GREEN 플립. cross-session-messaging.md 미러·배포 정합 확인 (무변경이 기본).

### M3 — GREEN(코드·문서 축) + 최종 검증

1. `role_rules.go`: doc 주석 사다리 1회화(:393–413), NOTE 지시문 천단위·중복 정리(:538–544), 로캘 경고 중복 서술 정리(:263–312), 버전 불일치 감지(배포 `system.yaml` `template_version` vs 바이너리 버전 — 판독 불가 시 조용히 생략, fail-open) + 상세와 4개 로캘에 `moai update` 안내. M1 테스트 GREEN 플립.
2. `hooks-system.md`: :131 두 한도 구분 기재 (50K 총 stdout→디스크 저장 / 10,000자 문자열당 전달 한도, 각 실측 출처 — Q4·lane-15 공지).
3. 전체 영향계열 검증: `go test ./internal/hook/... ./internal/template/...` (카드 트리 레인-로컬, env 스크럽, 컴파운드 1호출 — 스텁 게이트 포함 전 Green) + `golangci-lint run --timeout=2m ./internal/hook/... ./internal/template/...`.
4. **라이브 재진입 확인(완료 기준 1의 실측 반쪽)**: 팩토리 레인 세션 1회 재진입(re-launch) — 시작 공지에 "역할 규칙 주입 초과" 부재 + `additionalContext` 조립 합본 ≤10,000 관측을 progress.md §E.2 에 기록.
5. 크로스플랫폼 빌드 `GOOS=windows GOARCH=amd64 go build ./...` (role_rules_read_windows.go 계열 보존 확인).
6. Byte-delta 문 완성(REQ-RIB-011): 상시 파일(스텁) 전후 UTF-16 기록 — 증가 시 1,000바이트 초과분 비산 문.

### 마일스톤↔AC 대응

| AC | 플립 마일스톤 |
|---|---|
| AC-RIB-001/002 (예산 테스트 RED→GREEN) | M1 작성 → M2 플립 |
| AC-RIB-003 (재배치 원장) | M2 |
| AC-RIB-004 (스텁 게이트) | M1 작성(자기 모터 관측) → M2 확정 |
| AC-RIB-005/006 (안내·불일치) | M1 작성 → M3 플립 |
| AC-RIB-007 (문구 정리) | M3 |
| AC-RIB-008 (hooks-system) | M3 |
| AC-RIB-009 (gitflow 정합, 회귀 가드) | M2 생존 확인 |
| AC-RIB-010 (미러·델타 문) | M2/M3 |
| AC-RIB-011 (오버플로 경로 보존, 회귀 가드) | M3 전체 스위트 |
| AC-RIB-012 (라이브 재진입, 회귀 가드) | M3 |

## §D. 검증 계획

- 레인-로컬 원칙: 영향계열만(`internal/hook`, `internal/template`), env 스크럽 컴파운드 1호출, `go test ./...` 전체 금지(레인 규율). CI 가 전체를 돌린다.
- 모든 RED 명령은 단일 호출(파이프·`&&`·`;` 체이닝·서브셸 없음) — acceptance.md 증거 원장의 것을 그대로 재실행해 재현한다.
- 최종 턴 종료 리뷰: `codex_review scope=uncommitted` (레인 소관) — card-review.md 는 리더 정산 단계의 것.
- 커밋 규율: 마일스톤별 Conventional Commit, 카드 id t1617 본문 포함, `Authored-By-Agent` 트레일러, 로컬 착지(push 없음 — 리더 3f34).
