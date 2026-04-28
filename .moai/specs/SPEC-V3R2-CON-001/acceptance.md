# SPEC-V3R2-CON-001 Acceptance Criteria

> Given-When-Then 인수 조건 + 추적성 + 테스트 파일 매핑
> 버전: 1.1.0
> 최종 갱신: 2026-04-25 (plan-audit 대응)
> AC 개수: **17** (SPEC §6 의 12개 + Part B 추론 엣지 케이스 4개 + plan-audit 추가 AC-017)
> 출처: SPEC §6 Acceptance Criteria (AC-001 ~ AC-012) + 이 문서에서 파생된 AC-013 ~ AC-016 + plan-audit 추가 AC-017

---

## 읽는 방법

- 각 AC는 고유 ID (AC-CON-001-NNN, 3-digit zero-padded)와 REQ 추적성, 기대 테스트 파일 경로, 테스트 함수 이름을 명시한다.
- "테스트 기대 위치"는 tasks.md에서 오너 task가 생성할 파일과 일치한다.
- `SPEC source`: SPEC §6 명시 (원본 verbatim) 또는 `INFERRED`: 이 문서에서 파생.
- 모든 AC는 Given-When-Then 형식을 확장한 Given-When-Then-Evidence 구조를 따른다.
- 2026-04-25 플랜 감사 대응: AC ID 형식을 2자리에서 3자리 zero-padded로 표준화 (AC-01 → AC-001 등). REQ ID / CONST ID와 정렬.

---

## Part A: SPEC §6 원본 AC (12개)

### AC-CON-001-001 — 7개 FROZEN 불변 조항이 registry에 존재
- **Source**: SPEC §6 AC-CON-001-001
- **Traceability**: REQ-CON-001-005, REQ-CON-001-006
- **Given**: Fresh v3 설치 상태이며 `.claude/rules/moai/core/zone-registry.md`가 존재한다.
- **When**: 사용자가 `moai constitution list` 를 실행한다.
- **Then**: 출력에 `zone: Frozen`이 포함된 엔트리가 정확히 7개 이상 포함되며, 각 엔트리의 `clause` 필드는 SPEC §5.1 "Canonical 7 FROZEN Invariants" 표 column 2 의 verbatim 식별자(SPEC+EARS format, TRUST 5, @MX TAG protocol, 16-language neutrality, Template-First discipline, AskUserQuestion monopoly, Claude Code substrate)와 byte-level 일치한다.
- **Evidence**:
  - 테스트 파일: `internal/cli/constitution_integration_test.go`
  - 테스트 함수: `TestConstitutionListContainsSevenFrozenInvariants`
  - 실행: `go test -tags=integration ./internal/cli/... -run TestConstitutionListContainsSevenFrozenInvariants`
  - 부가 검증: `moai constitution list --zone frozen | wc -l` ≥ 7
  - Oracle: SPEC §5.1 표 (SPEC 자체가 canonical source; plan-audit 2026-04-25 이전 사용된 `master-v3 §1.3` 앵커는 해당 파일 §1.3 이 "Success Metrics" 이라 무효였으며 SPEC 내부 인라인으로 교체됨)

### AC-CON-001-002 — `--zone frozen` 필터 정확도
- **Source**: SPEC §6 AC-CON-001-002
- **Traceability**: REQ-CON-001-012
- **Given**: Registry가 Frozen 2개, Evolvable 3개를 포함한 fixture 상태이다.
- **When**: `moai constitution list --zone frozen` 을 실행한다.
- **Then**: 출력에는 Frozen 엔트리 2개만 표시되고 Evolvable 3개는 제외된다. Exit code 0.
- **Evidence**:
  - 테스트 파일: `internal/cli/constitution_test.go`
  - 테스트 함수: `TestConstitutionListFilterFrozen`
  - Fixture: `internal/constitution/testdata/valid_registry.md`

### AC-CON-001-003 — CI가 신규 HARD 조항의 registry 누락 탐지
- **Source**: SPEC §6 AC-CON-001-003
- **Traceability**: REQ-CON-001-010
- **Given**: 개발자가 `.claude/rules/moai/core/moai-constitution.md` 에 `[HARD]` 조항 1줄을 추가하되 `zone-registry.md`는 수정하지 않고 커밋한다.
- **When**: CI 파이프라인의 `constitution-check` job이 실행된다.
- **Then**: 빌드가 실패하며, 실패 메시지에는 누락된 조항의 파일 경로와 신규 `[HARD]` 라인이 명시된다.
- **Evidence**:
  - 테스트 파일: (Go 구현 채택, plan.md Decision 3 참조) `internal/cli/constitution_guard_test.go`
  - 테스트 함수: `TestGuardDetectsMissingRegistryUpdate`
  - CI evidence: `.github/workflows/*.yml` constitution-check job failure log

### AC-CON-001-004 — Go 타입 노출 시그니처 (6-field schema)
- **Source**: SPEC §6 AC-CON-001-004
- **Traceability**: REQ-CON-001-003, REQ-CON-001-004
- **Given**: `internal/constitution` 패키지가 빌드 가능한 상태이다.
- **When**: Go 프로그램에서 `import "github.com/modu-ai/moai-adk/internal/constitution"` 한다.
- **Then**: 패키지는 정확히 2개의 `Zone` 값(`ZoneFrozen`, `ZoneEvolvable`)과 `Rule` struct(exported fields: **ID, Zone, File, Anchor, Clause, CanaryGate — 정확히 6개**, 순서도 일치)를 노출한다.
- **Evidence**:
  - 테스트 파일: `internal/constitution/zone_test.go`, `internal/constitution/rule_test.go`
  - 테스트 함수: `TestZoneEnumValuesExactlyTwo`, `TestRuleStructFieldsMatchRegistrySchema`
  - 컴파일타임 검증: `go doc ./internal/constitution Zone` → 2 consts 표시
  - 필드 카운트 검증: reflect 를 사용하여 `reflect.TypeOf(Rule{}).NumField() == 6` 확인 (Orphan 등 internal-only 필드는 unexported 처리)

### AC-CON-001-005 — Clause verbatim 보존
- **Source**: SPEC §6 AC-CON-001-005
- **Traceability**: REQ-CON-001-007
- **Given**: `CONST-V3R2-001` 엔트리가 `zone-registry.md`에 존재하고 `clause: "SPEC+EARS format"`, `file: .claude/rules/moai/workflow/spec-workflow.md` 이다.
- **When**: Registry entry의 `clause`/`file` 필드를 SPEC §5.1 표 row 1 과 비교한다.
- **Then**: `clause`와 `file` 모두 byte-level 일치한다.
- **Evidence**:
  - 테스트 파일: `internal/cli/constitution_integration_test.go`
  - 테스트 함수: `TestClauseMatchesSpecCanonicalTable`
  - 구현 방법: Registry 로드 → entry.clause 와 SPEC 표의 canonical name 비교 (하드코딩된 7-tuple)
  - 실패 허용 조건: `Orphan=true`인 엔트리는 스킵

### AC-CON-001-006 — Strict 모드에서 중복 ID 감지 시 doctor 실패
- **Source**: SPEC §6 AC-CON-001-006
- **Traceability**: REQ-CON-001-030
- **Given**: `MOAI_CONSTITUTION_STRICT=1` 환경 변수 설정 + registry에 중복 `CONST-V3R2-042` 엔트리 2개.
- **When**: `moai doctor` 가 실행된다.
- **Then**: Exit code가 non-zero이며, stderr에 중복 ID가 포함된 경고 라인이 나타난다.
- **Evidence**:
  - 테스트 파일: `internal/cli/doctor_constitution_test.go`
  - 테스트 함수: `TestDoctorConstitutionStrictModeFailsOnDuplicate`
  - Fixture: `internal/constitution/testdata/duplicate_ids.md`
  - 환경: `t.Setenv("MOAI_CONSTITUTION_STRICT", "1")` — `t.Parallel()` 은 비활성화 (이 테스트는 env 변이, CLAUDE.local.md §6 OTEL 룰과 무관하지만 race + parallel 조합 회피). non-parallel subtest 로 실행.

### AC-CON-001-007 — Orphan 파일 참조 시 graceful degradation
- **Source**: SPEC §6 AC-CON-001-007
- **Traceability**: REQ-CON-001-040
- **Given**: Registry entry `CONST-V3R2-099` 가 존재하지 않는 파일 `non-existent.md` 를 참조한다.
- **When**: `LoadRegistry()` 가 호출된다.
- **Then**: 로더는 panic하지 않고, 해당 엔트리의 `Orphan` 필드를 `true`로 설정한 뒤 structured error(레벨 warning)를 반환한다. 다른 정상 엔트리는 모두 로드된다.
- **Evidence**:
  - 테스트 파일: `internal/constitution/loader_test.go`
  - 테스트 함수: `TestLoadRegistryMarksOrphanWithoutPanic`
  - Fixture: `internal/constitution/testdata/orphan_reference.md`

### AC-CON-001-008 — Design subsystem 미러링 + overflow 처리
- **Source**: SPEC §6 AC-CON-001-008
- **Traceability**: REQ-CON-001-021
- **Given**: `.claude/rules/moai/design/constitution.md` v3.3.0 이 존재하고 §2 + §3.1/§3.2/§3.3 에 [FROZEN] 마커가 있다.
- **When**: 핵심 registry가 로드된다.
- **Then**: design constitution §2 + §3.1/§3.2/§3.3 의 각 [FROZEN] 조항이 핵심 registry에 미러 엔트리로 포함되며, `file` 필드는 `.claude/rules/moai/design/constitution.md`를 가리킨다. 미러 엔트리 ID 범위는 **051-099**; 만약 미러링할 [FROZEN] 조항이 49 개를 초과하면 **100-149** 로 자동 확장되고 `moai doctor` 는 warning 을 발행한다.
- **Evidence**:
  - 테스트 파일: `internal/cli/constitution_integration_test.go`
  - 테스트 함수: `TestDesignSubsystemMirrorsPresent`, `TestDesignMirrorOverflowAutoExtends`
  - 검증: `moai constitution list --file .claude/rules/moai/design/constitution.md` 결과에서 ID가 `CONST-V3R2-05[1-9]` 또는 `CONST-V3R2-0[6-9][0-9]` 범위 확인 (기본); overflow fixture에서는 `CONST-V3R2-1[0-4][0-9]` 확인 + doctor warning

### AC-CON-001-009 — ID 안정성 (append-only)
- **Source**: SPEC §6 AC-CON-001-009
- **Traceability**: REQ-CON-001-011
- **Given**: 초기 빌드에서 `CONST-V3R2-042` 가 조항 X에 할당되었다.
- **When**: 후속 빌드에서 신규 조항 Y, Z가 추가되어 registry가 업데이트된다.
- **Then**: `CONST-V3R2-042`는 여전히 조항 X를 가리키며, Y/Z는 미사용 다음 번호를 부여받는다.
- **Evidence**:
  - 테스트 파일: `internal/constitution/loader_test.go`
  - 테스트 함수: `TestIDStabilityAppendOnly`
  - 검증 방식: git history에서 registry의 CONST-V3R2-042 변경 여부 검사 (또는 fixture 기반 시뮬레이션)

### AC-CON-001-010 — 출력 ID와 파일 내용 동기성
- **Source**: SPEC §6 AC-CON-001-010
- **Traceability**: REQ-CON-001-001, REQ-CON-001-006
- **Given**: `moai constitution list` 출력이 확보된 상태이다.
- **When**: 각 출력 Rule ID를 `zone-registry.md` 에서 grep한다.
- **Then**: 100%의 출력 ID가 파일 내에 존재한다 (역 매핑 완전성).
- **Evidence**:
  - 테스트 파일: `internal/cli/constitution_integration_test.go`
  - 테스트 함수: `TestListOutputIDsAllPresentInRegistryFile`
  - 실행: `comm -23` 등 set diff 도구 사용하여 검증

### AC-CON-001-011 — Dangling reference 경고 (skeleton API, 이 SPEC에서 완결 검증)
- **Source**: SPEC §6 AC-CON-001-011 (plan-audit 2026-04-25 강화: skeleton-pass-only 에서 behavior verification 으로 승격)
- **Traceability**: REQ-CON-001-041
- **Given**: Registry에 `CONST-V3R2-001` 엔트리 1개 이상 존재하고 `CONST-V3R2-999` 는 존재하지 않는다.
- **When**: skeleton API `ValidateRuleReferences(registry *Registry, refs []string) []string` 를 `refs: ["CONST-V3R2-999"]` 로 호출한다.
- **Then**: 반환 slice 길이 ≥ 1 이며, 첫 번째 element 는 비어 있지 않은 string 으로 literal substring `CONST-V3R2-999` 를 포함한다 (예: `"dangling reference: CONST-V3R2-999 not found in registry"`).
- **Evidence**:
  - 테스트 파일: `internal/constitution/dangling_test.go`
  - 테스트 함수: `TestValidateRuleReferencesReturnsWarningForUnknownID`
  - 상태: **이 SPEC 에서 behavioral verification 완료**. SPC-003 는 CLI wiring 과 SPEC frontmatter scanning 을 담당 (별도 SPEC 범위).
  - 구현 힌트: skeleton body 는 `for _, ref := range refs { if _, ok := registry.Get(ref); !ok { warnings = append(warnings, fmt.Sprintf("dangling reference: %s not found in registry", ref)) } }` 수준이면 충분.

### AC-CON-001-012 — AskUserQuestion monopoly 엔트리 정확성
- **Source**: SPEC §6 AC-CON-001-012
- **Traceability**: REQ-CON-001-005, REQ-CON-001-007
- **Given**: Registry가 로드된 상태이다.
- **When**: SPEC §5.1 표 row 6 "AskUserQuestion monopoly" 조항의 엔트리를 조회한다.
- **Then**: 정확히 1개 Frozen-zone 엔트리가 반환되며, `file: .claude/rules/moai/core/agent-common-protocol.md`이고 `clause`는 SPEC §5.1 표의 `AskUserQuestion monopoly` 와 verbatim 일치한다.
- **Evidence**:
  - 테스트 파일: `internal/cli/constitution_integration_test.go`
  - 테스트 함수: `TestAskUserQuestionMonopolyEntryExact`
  - 검증: clause 필드를 SPEC §5.1 canonical name 과 byte-level 비교

---

## Part B: 추론된 엣지 케이스 AC (4개)

SPEC §6 외에 코드 리뷰 관점에서 추가로 검증해야 하는 케이스. `INFERRED` 표기.

### AC-CON-001-013 — Registry 파일 부재 시 CLI 동작 (file-missing)
- **Source**: INFERRED (사용성 요구 — REQ-001 의 존재 전제)
- **Traceability**: REQ-CON-001-001 (negative path)
- **Given**: `.claude/rules/moai/core/zone-registry.md`가 존재하지 않는다 (파일 삭제 또는 미생성 상태).
- **When**: `moai constitution list` 를 실행한다.
- **Then**: Exit code 1, stderr에 명확한 에러 메시지("registry not found at <path>, run `moai init` or create manually") 출력. panic 없음.
- **Evidence**:
  - 테스트 파일: `internal/cli/constitution_test.go`
  - 테스트 함수: `TestConstitutionListRegistryMissing_FileNotFound`

### AC-CON-001-014 — 빈 Frozen zone 경고 (warn level)
- **Source**: INFERRED (REQ-020 명시 "doctor-level warning")
- **Traceability**: REQ-CON-001-020
- **Given**: Registry 파일이 존재하지만 `zone: Frozen`인 엔트리가 0개.
- **When**: `moai doctor constitution` 이 실행된다 (strict 모드 아님).
- **Then**: Exit code 0 (warning만), stdout에 "Constitution registry has no Frozen entries — this is invalid" 경고 표시.
- **Evidence**:
  - 테스트 파일: `internal/cli/doctor_constitution_test.go`
  - 테스트 함수: `TestDoctorConstitutionEmptyFrozenWarns`
  - Fixture: `internal/constitution/testdata/empty_frozen.md`

### AC-CON-001-015 — Registry 로드 성능 (cold start)
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

### AC-CON-001-016 — 바이너리 크기 회귀 없음
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

## Part C: plan-audit 추가 AC (1개)

### AC-CON-001-017 — Registry YAML 6-field schema 직접 검증 (REQ-002 direct mapping)
- **Source**: plan-audit 2026-04-25 — Finding #4 (REQ-002 의 indirect 매핑 해소)
- **Traceability**: REQ-CON-001-002 (direct, not indirect)
- **Given**: `.claude/rules/moai/core/zone-registry.md` 이 YAML-in-markdown 포맷으로 존재한다.
- **When**: 첫 번째 registry 엔트리를 파싱하고 YAML key set 을 추출한다.
- **Then**: key set 이 정확히 `{id, zone, file, anchor, clause, canary_gate}` (6 개, 소문자, 순서 무관, 별칭 금지) 이다. 추가 key 나 누락 key 가 있으면 FAIL.
- **Evidence**:
  - 테스트 파일: `internal/constitution/loader_test.go`
  - 테스트 함수: `TestRegistryEntryHasExactSixFieldsWithCanonicalNames`
  - 구현: yaml.v3 로 raw map[string]any unmarshal 후 key sort → `{"anchor", "canary_gate", "clause", "file", "id", "zone"}` 과 비교
  - 실행: `go test ./internal/constitution/ -run TestRegistryEntryHasExactSixFieldsWithCanonicalNames`

---

## REQ → AC 매핑 표

| REQ | 커버 AC |
|-----|---------|
| REQ-CON-001-001 | AC-001, AC-010, AC-013 |
| REQ-CON-001-002 | AC-017 (direct), AC-004 (Go struct side) |
| REQ-CON-001-003 | AC-004 |
| REQ-CON-001-004 | AC-004 |
| REQ-CON-001-005 | AC-001, AC-012 |
| REQ-CON-001-006 | AC-001, AC-010 |
| REQ-CON-001-007 | AC-005, AC-012 |
| REQ-CON-001-010 | AC-003 |
| REQ-CON-001-011 | AC-009 |
| REQ-CON-001-012 | AC-002 |
| REQ-CON-001-020 | AC-014 |
| REQ-CON-001-021 | AC-008 |
| REQ-CON-001-030 | AC-006 |
| REQ-CON-001-040 | AC-007 |
| REQ-CON-001-041 | AC-011 (behavior-verified, no longer skeleton-only) |
| 제약 §7 성능 | AC-015 |
| 제약 §7 바이너리 크기 | AC-016 |

모든 REQ에 최소 1개 AC 매핑 완료 (100% 커버리지). REQ-002 는 plan-audit 2026-04-25 이후 AC-017 의 direct mapping 을 가진다.

---

## Definition of Done

다음 체크리스트를 모두 만족할 때 SPEC-V3R2-CON-001 구현 완료:

### 기능 완성도
- [ ] AC-CON-001-001 ~ 012 모두 PASS (SPEC §6 원본 AC)
- [ ] AC-CON-001-013 ~ 016 모두 PASS (추론 AC)
- [ ] AC-CON-001-017 PASS (plan-audit 추가 AC, REQ-002 direct)
- [ ] REQ 15개 (Ubiquitous 7 + Event 4 [REQ-010/011/012/041] + State 2 + Optional 1 + Complex 1) 중 AC 매핑 100%
- [ ] AC-011 은 behavior-verified 수준 요구 (fixture 호출이 비어 있지 않은 warning string 반환 확인)

### 품질 게이트 (TRUST 5)
- [ ] **Tested**: `go test -race ./internal/constitution/...` 통과, 패키지 커버리지 ≥ 85%
- [ ] **Readable**: 모든 exported 심볼에 godoc, ID 상수 정규표현식 명시
- [ ] **Unified**: `gofmt -l ./...` 출력 비어있음, `golangci-lint run ./...` 통과
- [ ] **Secured**: `filepath.Clean` + projectDir scope 적용, registry 경로 traversal 불가 검증 완료
- [ ] **Trackable**: 모든 커밋 메시지에 `SPEC-V3R2-CON-001` 레퍼런스

### 비기능 요구
- [ ] Binary size 증가 < 50 KiB (AC-016)
- [ ] Registry cold load < 10ms (AC-015)
- [ ] Frozen 엔트리 수 ≥ 7 (SPEC §5.1 canonical 7)
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

### Open Questions 해소 (plan-audit 2026-04-25 대응: plan.md Decision Log 참조)
- [ ] OQ1 (YAML 파서) — **해소**: `gopkg.in/yaml.v3` 를 사용한 hybrid (YAML code fence 추출 + yaml.v3 unmarshal). plan.md §3.1 / Decision Log 2026-04-25 참조.
- [ ] OQ2 (CI workflow 파일 위치) — **해소**: 기존 `ci.yml` 에 `constitution-check` job 추가 (신규 파일 미생성). plan.md Decision Log 2026-04-25.
- [ ] OQ3 (CI 가드 구현 언어) — **해소**: Go 기반 (`internal/cli/constitution_guard.go` + `moai constitution guard` subcommand). Windows CI 호환성 및 TRUST 5 Tested 준수. plan.md Decision Log 2026-04-25.
- [ ] OQ4 (REQ-021 mirror 범위) — **해소**: SPEC §5.3 REQ-021 에서 §2 + §3.1/§3.2/§3.3 만 미러링, overflow 시 100-149 auto-extend.
- [ ] OQ5 (ID 할당 순서) — **해소**: SPEC §7 에 기록 — `(file, anchor_line_number)` 오름차순 (4 파일 고정 순서).
- [ ] OQ6 (CanaryGate 기본값) — **해소**: Frozen=true, Evolvable=false (plan.md Decision Log 2026-04-25; CON-002 에서 재평가 가능).

---

## 수동 검증 실행 순서 (최종 acceptance run)

QA 단계에서 다음 순서로 수동 실행:

1. `make build && make install` — 최신 바이너리
2. `moai --version` — 버전 출력 확인
3. `moai constitution list` — 전체 registry 표시
4. `moai constitution list --zone frozen` — Frozen 7+ 확인 (AC-001)
5. `moai constitution list --zone evolvable` — Evolvable 엔트리 표시 (AC-002 역방향)
6. `moai constitution list --format json | jq '.entries | length'` — 엔트리 수 확인
7. `moai doctor` — Constitution 섹션 포함 확인
8. `MOAI_CONSTITUTION_STRICT=1 moai doctor` — Strict 모드에서 동일 출력 (위반 없을 시)
9. 임의로 `.claude/rules/moai/core/moai-constitution.md` 에 dummy `[HARD] 테스트` 라인 추가 → registry 미수정 → `make constitution-check` → 실패 확인 (AC-003)
10. 라인 제거 → `make constitution-check` → 통과 확인
11. `go test -tags=integration ./internal/cli/...` — 통합 테스트 전체 통과
12. `go test -bench=BenchmarkLoadRegistry -benchtime=10s ./internal/constitution/` — 성능 목표 확인 (AC-015)
13. `ls -la bin/moai` 전후 비교 — 바이너리 크기 확인 (AC-016)

모든 단계 통과 시 DoD 완료로 간주.

---

## 요약

- **AC 총 개수**: 17 (§6 원본 12 + 추론 4 + plan-audit 추가 1)
- **REQ 커버리지**: 100% (모든 REQ에 ≥1 AC; REQ-002 는 AC-017 direct)
- **테스트 파일 수**: 7 (zone_test.go, rule_test.go, loader_test.go, dangling_test.go, constitution_test.go, constitution_integration_test.go, doctor_constitution_test.go)
- **TRUST 5 게이트**: 5/5 명시적 검증 항목 포함
- **OPEN QUESTIONS**: plan.md Decision Log 2026-04-25 에서 6개 모두 해소 완료
- **ID 형식**: AC/REQ/CONST 모두 3-digit zero-padded (AC-CON-001-001 ~ AC-CON-001-017, REQ-CON-001-001 ~ REQ-CON-001-041, CONST-V3R2-001 ~ CONST-V3R2-149)
