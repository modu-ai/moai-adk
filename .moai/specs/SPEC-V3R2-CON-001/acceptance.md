# SPEC-V3R2-CON-001 Acceptance Criteria

> Given-When-Then 인수 조건 + 추적성 + 테스트 파일 매핑
> 최종 갱신: 2026-04-24
> AC 개수: **16** (SPEC §6 의 12개 + 추론된 엣지 케이스 4개)
> 출처: SPEC §6 Acceptance Criteria (AC-01 ~ AC-12) + 이 문서에서 파생된 AC-13 ~ AC-16

---

## 읽는 방법

- 각 AC는 고유 ID (AC-CON-001-NN)와 REQ 추적성, 기대 테스트 파일 경로, 테스트 함수 이름을 명시한다.
- "테스트 기대 위치"는 tasks.md에서 오너 task가 생성할 파일과 일치한다.
- `SPEC source`: SPEC §6 명시 (원본 verbatim) 또는 `INFERRED`: 이 문서에서 파생.
- 모든 AC는 Given-When-Then 형식을 확장한 Given-When-Then-Evidence 구조를 따른다.

---

## Part A: SPEC §6 원본 AC (12개)

### AC-CON-001-01 — 7개 FROZEN 불변 조항이 registry에 존재
- **Source**: SPEC §6 AC-CON-001-01
- **Traceability**: REQ-CON-001-005, REQ-CON-001-006
- **Given**: Fresh v3 설치 상태이며 `.claude/rules/moai/core/zone-registry.md`가 존재한다.
- **When**: 사용자가 `moai constitution list` 를 실행한다.
- **Then**: 출력에 `zone: Frozen`이 포함된 엔트리가 정확히 7개 이상 포함되며, 각 엔트리는 master-v3 §1.3 불변 조항(SPEC+EARS, TRUST 5, @MX TAG, 16-language neutrality, Template-First, AskUserQuestion monopoly, Claude Code substrate)과 verbatim 대응한다.
- **Evidence**:
  - 테스트 파일: `internal/cli/constitution_integration_test.go`
  - 테스트 함수: `TestConstitutionListContainsSevenFrozenInvariants`
  - 실행: `go test -tags=integration ./internal/cli/... -run TestConstitutionListContainsSevenFrozenInvariants`
  - 부가 검증: `moai constitution list --zone frozen | wc -l` ≥ 7

### AC-CON-001-02 — `--zone frozen` 필터 정확도
- **Source**: SPEC §6 AC-CON-001-02
- **Traceability**: REQ-CON-001-012
- **Given**: Registry가 Frozen 2개, Evolvable 3개를 포함한 fixture 상태이다.
- **When**: `moai constitution list --zone frozen` 을 실행한다.
- **Then**: 출력에는 Frozen 엔트리 2개만 표시되고 Evolvable 3개는 제외된다. Exit code 0.
- **Evidence**:
  - 테스트 파일: `internal/cli/constitution_test.go`
  - 테스트 함수: `TestConstitutionListFilterFrozen`
  - Fixture: `internal/constitution/testdata/valid_registry.md`

### AC-CON-001-03 — CI가 신규 HARD 조항의 registry 누락 탐지
- **Source**: SPEC §6 AC-CON-001-03
- **Traceability**: REQ-CON-001-010
- **Given**: 개발자가 `.claude/rules/moai/core/moai-constitution.md` 에 `[HARD]` 조항 1줄을 추가하되 `zone-registry.md`는 수정하지 않고 커밋한다.
- **When**: CI 파이프라인의 `constitution-check` job이 실행된다.
- **Then**: 빌드가 실패하며, 실패 메시지에는 누락된 조항의 파일 경로와 신규 `[HARD]` 라인이 명시된다.
- **Evidence**:
  - 테스트 파일: (Go 옵션 채택 시) `internal/cli/constitution_guard_test.go`; (bash 옵션 채택 시) `scripts/constitution_guard_test.sh`
  - 테스트 함수: `TestGuardDetectsMissingRegistryUpdate`
  - CI evidence: `.github/workflows/*.yml` constitution-check job failure log

### AC-CON-001-04 — Go 타입 노출 시그니처
- **Source**: SPEC §6 AC-CON-001-04
- **Traceability**: REQ-CON-001-003, REQ-CON-001-004
- **Given**: `internal/constitution` 패키지가 빌드 가능한 상태이다.
- **When**: Go 프로그램에서 `import "github.com/modu-ai/moai-adk/internal/constitution"` 한다.
- **Then**: 패키지는 정확히 2개의 `Zone` 값(`ZoneFrozen`, `ZoneEvolvable`)과 `Rule` struct(exported fields: ID, Zone, File, Anchor, Clause, CanaryGate)를 노출한다.
- **Evidence**:
  - 테스트 파일: `internal/constitution/zone_test.go`, `internal/constitution/rule_test.go`
  - 테스트 함수: `TestZoneEnumValuesExactlyTwo`, `TestRuleStructFieldsMatchRegistrySchema`
  - 컴파일타임 검증: `go doc ./internal/constitution Zone` → 2 consts 표시

### AC-CON-001-05 — Clause verbatim 보존
- **Source**: SPEC §6 AC-CON-001-05
- **Traceability**: REQ-CON-001-007
- **Given**: `CONST-V3R2-001` 엔트리가 `zone-registry.md`에 존재하고, 소스 파일 `.claude/rules/moai/workflow/spec-workflow.md` 의 해당 앵커 텍스트가 읽을 수 있다.
- **When**: Registry entry의 `clause` 필드와 소스 파일의 해당 라인을 비교한다.
- **Then**: 두 텍스트가 공백·줄바꿈 포함 byte-level 일치한다.
- **Evidence**:
  - 테스트 파일: `internal/cli/constitution_integration_test.go`
  - 테스트 함수: `TestClauseMatchesSourceVerbatim`
  - 구현 방법: Registry 로드 → 각 엔트리 `file`/`anchor` 해석 → 소스 읽어 비교
  - 실패 허용 조건: `Orphan=true`인 엔트리는 스킵

### AC-CON-001-06 — Strict 모드에서 중복 ID 감지 시 doctor 실패
- **Source**: SPEC §6 AC-CON-001-06
- **Traceability**: REQ-CON-001-030
- **Given**: `MOAI_CONSTITUTION_STRICT=1` 환경 변수 설정 + registry에 중복 `CONST-V3R2-042` 엔트리 2개.
- **When**: `moai doctor` 가 실행된다.
- **Then**: Exit code가 non-zero이며, stderr에 중복 ID가 포함된 경고 라인이 나타난다.
- **Evidence**:
  - 테스트 파일: `internal/cli/doctor_constitution_test.go`
  - 테스트 함수: `TestDoctorConstitutionStrictModeFailsOnDuplicate`
  - Fixture: `internal/constitution/testdata/duplicate_ids.md`
  - 환경: `t.Setenv("MOAI_CONSTITUTION_STRICT", "1")` — 단 CLAUDE.local.md §6 OTEL 룰과 무관한 단순 env이므로 허용

### AC-CON-001-07 — Orphan 파일 참조 시 graceful degradation
- **Source**: SPEC §6 AC-CON-001-07
- **Traceability**: REQ-CON-001-040
- **Given**: Registry entry `CONST-V3R2-099` 가 존재하지 않는 파일 `non-existent.md` 를 참조한다.
- **When**: `LoadRegistry()` 가 호출된다.
- **Then**: 로더는 panic하지 않고, 해당 엔트리의 `Orphan` 필드를 `true`로 설정한 뒤 structured error(레벨 warning)를 반환한다. 다른 정상 엔트리는 모두 로드된다.
- **Evidence**:
  - 테스트 파일: `internal/constitution/loader_test.go`
  - 테스트 함수: `TestLoadRegistryMarksOrphanWithoutPanic`
  - Fixture: `internal/constitution/testdata/orphan_reference.md`

### AC-CON-001-08 — Design subsystem 미러링
- **Source**: SPEC §6 AC-CON-001-08
- **Traceability**: REQ-CON-001-021
- **Given**: `.claude/rules/moai/design/constitution.md` v3.3.0 이 존재하고 §2에 FROZEN 리스트가 있다.
- **When**: 핵심 registry가 로드된다.
- **Then**: design constitution의 각 FROZEN 조항이 핵심 registry에 미러 엔트리로 포함되며, `file` 필드는 `.claude/rules/moai/design/constitution.md`를 가리킨다. 미러 엔트리 ID 범위는 051-099.
- **Evidence**:
  - 테스트 파일: `internal/cli/constitution_integration_test.go`
  - 테스트 함수: `TestDesignSubsystemMirrorsPresent`
  - 검증: `moai constitution list --file .claude/rules/moai/design/constitution.md` 결과에서 ID가 `CONST-V3R2-05[1-9]` 또는 `CONST-V3R2-0[6-9][0-9]` 범위 확인

### AC-CON-001-09 — ID 안정성 (append-only)
- **Source**: SPEC §6 AC-CON-001-09
- **Traceability**: REQ-CON-001-011
- **Given**: 초기 빌드에서 `CONST-V3R2-042` 가 조항 X에 할당되었다.
- **When**: 후속 빌드에서 신규 조항 Y, Z가 추가되어 registry가 업데이트된다.
- **Then**: `CONST-V3R2-042`는 여전히 조항 X를 가리키며, Y/Z는 미사용 다음 번호를 부여받는다.
- **Evidence**:
  - 테스트 파일: `internal/constitution/loader_test.go`
  - 테스트 함수: `TestIDStabilityAppendOnly`
  - 검증 방식: git history에서 registry의 CONST-V3R2-042 변경 여부 검사 (또는 fixture 기반 시뮬레이션)

### AC-CON-001-10 — 출력 ID와 파일 내용 동기성
- **Source**: SPEC §6 AC-CON-001-10
- **Traceability**: REQ-CON-001-001, REQ-CON-001-006
- **Given**: `moai constitution list` 출력이 확보된 상태이다.
- **When**: 각 출력 Rule ID를 `zone-registry.md` 에서 grep한다.
- **Then**: 100%의 출력 ID가 파일 내에 존재한다 (역 매핑 완전성).
- **Evidence**:
  - 테스트 파일: `internal/cli/constitution_integration_test.go`
  - 테스트 함수: `TestListOutputIDsAllPresentInRegistryFile`
  - 실행: `comm -23` 등 set diff 도구 사용하여 검증

### AC-CON-001-11 — Dangling reference 경고 (SPC-003 skeleton)
- **Source**: SPEC §6 AC-CON-001-11
- **Traceability**: REQ-CON-001-041
- **Given**: SPEC 문서가 YAML frontmatter에 `related_rule: [CONST-V3R2-999]` (등록되지 않은 ID) 를 선언한다.
- **When**: SPEC-V3R2-SPC-003 linter가 실행된다 (향후 구현; 본 SPEC에서는 skeleton API 제공).
- **Then**: linter가 "dangling reference" 경고를 발행하며 알 수 없는 ID를 citation한다.
- **Evidence**:
  - 테스트 파일: `internal/constitution/dangling_test.go` (skeleton 단계 — SPC-003에서 실제 테스트 채움)
  - 테스트 함수: `TestValidateRuleReferencesSignatureOnly` (skeleton 현시점)
  - 상태: **이 SPEC에서는 API 시그니처만 확정**; 실제 동작 검증은 SPC-003에서 수행
  - **주의**: AC-11은 CON-001에서는 "skeleton pass" 수준으로 완화됨 — 구현은 SPC-003 의 작업

### AC-CON-001-12 — AskUserQuestion monopoly 엔트리 정확성
- **Source**: SPEC §6 AC-CON-001-12
- **Traceability**: REQ-CON-001-005, REQ-CON-001-007
- **Given**: Registry가 로드된 상태이다.
- **When**: master-v3 §1.3 "AskUserQuestion monopoly" 조항의 엔트리를 조회한다.
- **Then**: 정확히 1개 Frozen-zone 엔트리가 반환되며, `file: .claude/rules/moai/core/agent-common-protocol.md`이고 `clause`는 원본 텍스트와 verbatim 일치한다.
- **Evidence**:
  - 테스트 파일: `internal/cli/constitution_integration_test.go`
  - 테스트 함수: `TestAskUserQuestionMonopolyEntryExact`
  - 검증: clause 필드를 `.claude/rules/moai/core/agent-common-protocol.md` line 7 텍스트와 byte-level 비교

---

## Part B: 추론된 엣지 케이스 AC (4개)

SPEC §6 외에 코드 리뷰 관점에서 추가로 검증해야 하는 케이스. `INFERRED` 표기.

### AC-CON-001-13 — Registry 파일 부재 시 CLI 동작
- **Source**: INFERRED (사용성 요구 — REQ-001 의 존재 전제)
- **Traceability**: REQ-CON-001-001 (negative path)
- **Given**: `.claude/rules/moai/core/zone-registry.md`가 존재하지 않거나 권한 없음.
- **When**: `moai constitution list` 를 실행한다.
- **Then**: Exit code 1, stderr에 명확한 에러 메시지("registry not found at <path>, run `moai init` or create manually") 출력. panic 없음.
- **Evidence**:
  - 테스트 파일: `internal/cli/constitution_test.go`
  - 테스트 함수: `TestConstitutionListRegistryMissing`

### AC-CON-001-14 — 빈 Frozen zone 경고 (warn level)
- **Source**: INFERRED (REQ-020 명시 "doctor-level warning")
- **Traceability**: REQ-CON-001-020
- **Given**: Registry 파일이 존재하지만 `zone: Frozen`인 엔트리가 0개.
- **When**: `moai doctor constitution` 이 실행된다 (strict 모드 아님).
- **Then**: Exit code 0 (warning만), stdout에 "Constitution registry has no Frozen entries — this is invalid" 경고 표시.
- **Evidence**:
  - 테스트 파일: `internal/cli/doctor_constitution_test.go`
  - 테스트 함수: `TestDoctorConstitutionEmptyFrozenWarns`
  - Fixture: `internal/constitution/testdata/empty_frozen.md`

### AC-CON-001-15 — Registry 로드 성능 (cold start)
- **Source**: INFERRED (SPEC 제약 §7 "registry loader performance budget: <10ms for 200 entries")
- **Traceability**: 제약 §7 (비기능 요구)
- **Given**: 200개 엔트리를 포함한 registry fixture.
- **When**: 콜드 상태에서 `LoadRegistry()` 호출 (벤치마크 `b.ResetTimer()` 이후 측정).
- **Then**: 1회 호출 당 실행 시간이 10ms 미만 (median, -benchtime=10s 옵션).
- **Evidence**:
  - 테스트 파일: `internal/constitution/loader_test.go`
  - 테스트 함수: `BenchmarkLoadRegistry200Entries`
  - 실행: `go test -bench=BenchmarkLoadRegistry200Entries -benchtime=10s ./internal/constitution/`
  - Fixture: 자동 생성 (테스트 setup 단계에서 임시 파일로 200 엔트리 합성)

### AC-CON-001-16 — 바이너리 크기 회귀 없음
- **Source**: INFERRED (SPEC 제약 §7 "<50KB 증가")
- **Traceability**: 제약 §7 (비기능 요구)
- **Given**: `main` 브랜치 기준 `bin/moai` 크기 B0.
- **When**: Phase 1-4 완료 후 동일 빌드 플래그로 `bin/moai` 생성 — 크기 B1.
- **Then**: `B1 - B0 < 50 KiB` (50 * 1024 바이트).
- **Evidence**:
  - 테스트 파일: `scripts/binary_size_check.sh` 또는 CI workflow step
  - 테스트 함수: CI job `binary-size-regression`
  - 증빙: 머지 PR description에 before/after bytes 기록

---

## REQ → AC 매핑 표

| REQ | 커버 AC |
|-----|---------|
| REQ-CON-001-001 | AC-01, AC-10, AC-13 |
| REQ-CON-001-002 | AC-01, AC-04 (간접) |
| REQ-CON-001-003 | AC-04 |
| REQ-CON-001-004 | AC-04 |
| REQ-CON-001-005 | AC-01, AC-12 |
| REQ-CON-001-006 | AC-01, AC-10 |
| REQ-CON-001-007 | AC-05, AC-12 |
| REQ-CON-001-010 | AC-03 |
| REQ-CON-001-011 | AC-09 |
| REQ-CON-001-012 | AC-02 |
| REQ-CON-001-020 | AC-14 |
| REQ-CON-001-021 | AC-08 |
| REQ-CON-001-030 | AC-06 |
| REQ-CON-001-040 | AC-07 |
| REQ-CON-001-041 | AC-11 (skeleton 수준) |
| 제약 §7 성능 | AC-15 |
| 제약 §7 바이너리 크기 | AC-16 |

모든 REQ에 최소 1개 AC 매핑 완료 (100% 커버리지).

---

## Definition of Done

다음 체크리스트를 모두 만족할 때 SPEC-V3R2-CON-001 구현 완료:

### 기능 완성도
- [ ] AC-CON-001-01 ~ 12 모두 PASS (SPEC §6 원본 AC)
- [ ] AC-CON-001-13 ~ 16 모두 PASS (추론 AC)
- [ ] REQ 15개 (Ubiquitous 7 + Event 3 + State 2 + Optional 1 + Complex 2) 중 AC 매핑 100%
- [ ] AC-11은 skeleton 수준만 요구 (실제 동작은 SPC-003에서 완성; 시그니처 공개만 확인)

### 품질 게이트 (TRUST 5)
- [ ] **Tested**: `go test -race ./internal/constitution/...` 통과, 패키지 커버리지 ≥ 85%
- [ ] **Readable**: 모든 exported 심볼에 godoc, ID 상수 정규표현식 명시
- [ ] **Unified**: `gofmt -l ./...` 출력 비어있음, `golangci-lint run ./...` 통과
- [ ] **Secured**: `filepath.Clean` + projectDir scope 적용, registry 경로 traversal 불가 검증 완료
- [ ] **Trackable**: 모든 커밋 메시지에 `SPEC-V3R2-CON-001` 레퍼런스

### 비기능 요구
- [ ] Binary size 증가 < 50 KiB (AC-16)
- [ ] Registry cold load < 10ms (AC-15)
- [ ] Frozen 엔트리 수 ≥ 7 (master-v3 §1.3)
- [ ] 전체 registry 엔트리 수 ≥ 30 (Phase 1 annotation 목표)

### 프로세스 준수
- [ ] Template-First 규율: `.claude/rules/moai/core/zone-registry.md`와 `internal/template/templates/.claude/rules/moai/core/zone-registry.md` byte-level 일치
- [ ] `make build` 후 embedded 파일 재생성 확인
- [ ] CHANGELOG.md 업데이트
- [ ] spec.md status 필드 `draft → implemented` (본문 FROZEN 준수)

### 통합 검증
- [ ] `moai constitution list` 수동 실행 성공 (사용자 시연 영상 또는 스크린샷)
- [ ] `moai doctor` 출력에 Constitution 섹션 포함
- [ ] CI constitution-check job 녹색 (warn 모드부터 시작)
- [ ] 의존 SPEC (CON-002, CON-003, SPC-003, RT-005) 준비팀에 해금 통지

### Open Questions 해소
- [ ] plan.md §7 OPEN QUESTIONS 6개 항목 모두 결정 또는 후속 SPEC으로 이관 기록
- [ ] 특히 OPEN QUESTION 1 (YAML 파서), OPEN QUESTION 3 (CI 가드 언어), OPEN QUESTION 5 (ID 할당 순서)는 필수 해소

---

## 수동 검증 실행 순서 (최종 acceptance run)

QA 단계에서 다음 순서로 수동 실행:

1. `make build && make install` — 최신 바이너리
2. `moai --version` — 버전 출력 확인
3. `moai constitution list` — 전체 registry 표시
4. `moai constitution list --zone frozen` — Frozen 7+ 확인 (AC-01)
5. `moai constitution list --zone evolvable` — Evolvable 엔트리 표시 (AC-02 역방향)
6. `moai constitution list --format json | jq '.entries | length'` — 엔트리 수 확인
7. `moai doctor` — Constitution 섹션 포함 확인
8. `MOAI_CONSTITUTION_STRICT=1 moai doctor` — Strict 모드에서 동일 출력 (위반 없을 시)
9. 임의로 `.claude/rules/moai/core/moai-constitution.md` 에 dummy `[HARD] 테스트` 라인 추가 → registry 미수정 → `make constitution-check` → 실패 확인 (AC-03)
10. 라인 제거 → `make constitution-check` → 통과 확인
11. `go test -tags=integration ./internal/cli/...` — 통합 테스트 전체 통과
12. `go test -bench=BenchmarkLoadRegistry -benchtime=10s ./internal/constitution/` — 성능 목표 확인 (AC-15)
13. `ls -la bin/moai` 전후 비교 — 바이너리 크기 확인 (AC-16)

모든 단계 통과 시 DoD 완료로 간주.

---

## 요약

- **AC 총 개수**: 16 (§6 원본 12 + 추론 4)
- **REQ 커버리지**: 100% (모든 REQ에 ≥1 AC)
- **테스트 파일 수**: 6 (zone_test.go, rule_test.go, loader_test.go, constitution_test.go, constitution_integration_test.go, doctor_constitution_test.go)
- **TRUST 5 게이트**: 5/5 명시적 검증 항목 포함
- **OPEN QUESTIONS**: plan.md §7 참조 — 구현 착수 전/중 해소 요구
