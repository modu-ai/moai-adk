# SPEC-V3R2-CON-001 Task Decomposition

> 구현 작업 분해 및 의존성 그래프
> 버전: 1.1.0
> 최종 갱신: 2026-04-25 (plan-audit 대응 — AC ID 3-digit + G3 직렬 반영 + DoD AC 참조 추가 + task count 재계산)
> Phase 매핑: plan.md §1 Phase 1-4
> Task 개수: **22**

---

## 범례

- **Owner-role**: `implementer` (구현), `tester` (테스트 전담), `reviewer` (리뷰), `researcher` (조사·annotation). 팀 모드 시 `workflow.yaml` role profile로 스폰.
- **Isolation**: 팀 모드 병렬 실행 시 `worktree` 필요 여부 (구현/테스트 파일 쓰기 → worktree, 읽기 전용 → 불필요).
- **Blocks**: 이 task 완료가 선행 조건인 후속 task ID.
- **Parallel Group**: 동일 그룹 내 task는 Agent Teams 병렬 실행 가능 (파일 소유권 중복 없음).

---

## Parallel Group 요약

| Group | Tasks | 사전 조건 |
|-------|-------|-----------|
| G0 (직렬) | T-01, T-02, T-03 | 없음 (Phase 1 annotation) |
| G1 (병렬) | T-04, T-05, T-06 | G0 완료 |
| G2 (직렬) | T-07, T-08, T-09 | G1 완료 |
| G3 (직렬) | T-10, T-11, T-12 | G2 완료 (plan-audit 2026-04-25: 병렬 → 직렬 확정; T-11 이 T-10 시그니처에 의존) |
| G4 (직렬) | T-13, T-14 | G3 완료 |
| G5 (병렬) | T-15, T-16, T-17 | T-14 완료 |
| G6 (직렬) | T-18, T-19 | G5 완료 |
| G7 (통합) | T-20, T-21, T-22 | G6 완료 |

---

## Phase 1: Zone Registry 작성

### T-01 HARD 조항 수집 및 annotation 표 작성
- **Owner-role**: researcher (read-only)
- **Isolation**: 불필요 (read-only)
- **File ownership**: 없음 (조사 단계, 산출물은 작업 노트)
- **의존성**: 없음 (루트 task)
- **Blocks**: T-02
- **설명**: 4개 load-bearing 규칙 파일(`CLAUDE.md`, `moai-constitution.md`, `agent-common-protocol.md`, `design/constitution.md`)에서 `[HARD]` 조항을 grep 스캔. 각 조항에 대해 잠정 Zone 분류(Frozen/Evolvable), 소스 파일 경로, 섹션 앵커, verbatim clause text 기록. 작업 노트는 `.moai/specs/SPEC-V3R2-CON-001/annotations-worksheet.md` (임시 파일, 최종 머지 전 정리 또는 삭제).
- **DoD**:
  - [ ] `[HARD]` 조항 ≥ 30개 수집 확인 (grep count 증빙)
  - [ ] 각 조항에 Zone 분류 근거 1줄 기록
  - [ ] SPEC §5.1 canonical 7 FROZEN 불변 조항이 모두 포함됨을 대조 (clause 필드가 verbatim 일치할 예정임을 확인)
  - [ ] OQ 4/5 는 plan.md §7 Decision Log 에서 이미 해소 — 본 task 는 해당 결정을 적용만 하면 됨

### T-02 CONST-V3R2-NNN ID 번호 할당 규칙 적용
- **Owner-role**: researcher
- **Isolation**: 불필요
- **File ownership**: 없음
- **의존성**: T-01
- **Blocks**: T-03
- **설명**: plan.md §7 Decision Log OQ5 에서 확정된 규칙 — `(file, anchor_line_number)` 오름차순, 4 파일 고정 순서 — 을 T-01 산출물에 적용. 001-050 = pre-existing, 051-099 = design mirror, 100-149 = design overflow 예약, 150+ = 미래 신규. 규칙을 zone-registry.md 상단 `## ID Allocation Policy` 섹션에 문서화.
- **DoD**:
  - [ ] ID 할당 규칙 문서화 완료 (spec.md §7 + plan.md §7 OQ5 참조)
  - [ ] T-01에서 수집된 모든 조항에 고유 ID 잠정 배정 (규칙 적용)
  - [ ] 중복 없음 확인 (bash sort + uniq)

### T-03 `zone-registry.md` 및 템플릿 트윈 작성
- **Owner-role**: implementer
- **Isolation**: 불필요 (단일 파일 2 복사본)
- **File ownership**:
  - `.claude/rules/moai/core/zone-registry.md`
  - `internal/template/templates/.claude/rules/moai/core/zone-registry.md`
- **의존성**: T-02
- **Blocks**: T-04, T-05, T-06 (Phase 2 전체)
- **설명**: T-01/T-02 산출물을 plan.md §3.2 결정된 YAML-in-markdown 포맷으로 전환. HISTORY 섹션, 사용 가이드, ID Allocation Policy, Entries (YAML code fence) 순서로 구성. 두 복사본을 정확히 일치시킴 (Template-First 규율). CanaryGate 기본값은 plan.md §7 OQ6 결정 적용 (Frozen=true, Evolvable=false).
- **DoD**:
  - [ ] 두 파일 byte-level 동일 (diff 결과 empty)
  - [ ] Frozen 엔트리 ≥ 7개 (SPEC §5.1 canonical 7)
  - [ ] 전체 엔트리 ≥ 30개
  - [ ] YAML 파싱 가능 (로컬 `yq` 또는 `python -c "import yaml;..."` 사전 검증)
  - [ ] REQ-CON-001-001, 002, 005, 007 예비 검증 통과
  - [ ] 사용자 수동 리뷰 완료 (어떤 HARD도 누락되지 않음)
  - [ ] AC-CON-001-017 (6-field schema) 선검증: 첫 번째 엔트리 key set 이 정확히 `{id, zone, file, anchor, clause, canary_gate}` 임을 yq 로 확인

---

## Phase 2: Go 타입 + 로더

### T-04 `internal/constitution/zone.go` 구현
- **Owner-role**: implementer
- **Isolation**: worktree (병렬 쓰기)
- **File ownership**: `internal/constitution/zone.go`
- **의존성**: T-03 (Phase 1 완료 후 시작 — 스키마 확정)
- **Blocks**: T-07, T-09
- **설명**: `Zone uint8` enum + 상수 2개 + `String()` + `ParseZone()`. REQ-CON-001-003 직접 구현. godoc 주석 필수.
- **DoD**:
  - [ ] `go vet ./internal/constitution/...` 통과
  - [ ] `golangci-lint run ./internal/constitution/...` 통과
  - [ ] Exported 심볼에 godoc
  - [ ] `Zone` 값이 정확히 2개 (`ZoneFrozen=0`, `ZoneEvolvable=1`)

### T-05 `internal/constitution/rule.go` 구현
- **Owner-role**: implementer
- **Isolation**: worktree
- **File ownership**: `internal/constitution/rule.go`
- **의존성**: T-03
- **Blocks**: T-07, T-08
- **설명**: `Rule` 구조체 + `Validate()` 메서드. REQ-CON-001-004 직접 구현. **6 exported fields**: `ID string`, `Zone Zone`, `File string`, `Anchor string`, `Clause string`, `CanaryGate bool` (순서 포함). ID 정규표현식 `^CONST-V3R2-\d{3,}$` 상수로 정의. `Orphan bool` 은 internal-only (unexported, loader 가 설정).
- **DoD**:
  - [ ] Rule struct 이 정확히 6 exported fields 를 가짐 (AC-CON-001-004 직접 검증)
  - [ ] 필드명이 registry YAML 키와 정확히 매핑됨 (yaml.v3 태그: `yaml:"id"`, `yaml:"zone"`, `yaml:"file"`, `yaml:"anchor"`, `yaml:"clause"`, `yaml:"canary_gate"`)
  - [ ] 빈 ID, 잘못된 형식, 빈 Clause 등 validation 실패 케이스 에러 반환
  - [ ] `orphan` 은 unexported (yaml tag 없음, loader 가 reflect 로 세팅 안 함 — 별도 internal struct 추천)

### T-06 testdata 픽스처 작성
- **Owner-role**: implementer
- **Isolation**: worktree
- **File ownership**:
  - `internal/constitution/testdata/valid_registry.md`
  - `internal/constitution/testdata/duplicate_ids.md`
  - `internal/constitution/testdata/orphan_reference.md`
  - `internal/constitution/testdata/malformed_yaml.md`
  - `internal/constitution/testdata/empty_frozen.md`
  - `internal/constitution/testdata/overflow_mirror.md` (50+ design mirrors for AC-008 overflow test)
- **의존성**: T-03
- **Blocks**: T-07, T-08
- **설명**: 로더 단위 테스트용 입력 파일 6종. 각 파일은 의도적 오류 유형 1가지씩 내포 (또는 valid case). T-03 실제 registry와 독립적 — 미래 스키마 변경이 테스트를 깨뜨리지 않도록 최소 예시. overflow_mirror 는 REQ-021 의 auto-extend 로직 검증용.
- **DoD**:
  - [ ] 6개 파일 생성
  - [ ] valid_registry.md는 엔트리 ≥ 3개 (Frozen 2, Evolvable 1)
  - [ ] 각 파일 상단에 설명 주석
  - [ ] overflow_mirror.md 는 51 개 design mirror 엔트리 포함 (auto-extend to 100-149 로직 검증)

### T-07 `internal/constitution/loader.go` 구현
- **Owner-role**: implementer
- **Isolation**: worktree
- **File ownership**: `internal/constitution/loader.go`
- **의존성**: T-04, T-05
- **Blocks**: T-08, T-10, T-13
- **설명**: `LoadRegistry`, `Registry`, `Get`, `FilterByZone` 구현. Markdown 에서 YAML code fence 추출 → yaml.v3 unmarshal → Rule 검증 (plan.md §7 OQ1 Decision). 중복 ID, orphan 파일, 빈 Frozen, design mirror overflow 감지 로직. REQ-001, 002, 012, 040, 021 구현. 성능 목표 <10ms for 200 entries.
- **DoD**:
  - [ ] YAML code fence 첫 매치 추출 로직 테스트 (여러 fence 있을 때 첫 번째만)
  - [ ] `filepath.Clean` + `projectDir` scope 제한 (경로 traversal 방지)
  - [ ] Orphan 처리 시 panic 금지 (REQ-040), `Orphan=true` 플래그만 세팅
  - [ ] 중복 ID는 fatal error (REQ-030 strict 모드 외에서도 거부)
  - [ ] Design mirror overflow 감지: ID 범위 051-099 초과 시 100-149 로 auto-extend + warning (REQ-021)
  - [ ] godoc 완비

### T-08 Phase 2 단위 테스트 (`loader_test.go`, `rule_test.go`, `zone_test.go`)
- **Owner-role**: tester
- **Isolation**: worktree
- **File ownership**:
  - `internal/constitution/zone_test.go`
  - `internal/constitution/rule_test.go`
  - `internal/constitution/loader_test.go`
- **의존성**: T-04, T-05, T-06, T-07
- **Blocks**: T-13
- **설명**: Table-driven tests. Zone parse/render, Rule validate 양/음성, Loader 6 시나리오 (valid, duplicate, malformed, orphan, empty frozen, overflow mirror). 성능 benchmark `BenchmarkLoadRegistry` 포함. AC-CON-001-015 (성능) 및 AC-CON-001-017 (6-field schema direct) 검증 포함.
- **DoD**:
  - [ ] 패키지 커버리지 ≥ 85% (`go test -cover ./internal/constitution/...`)
  - [ ] `go test -race` 통과
  - [ ] Benchmark 결과 기록 (주석 또는 별도 벤치 리포트)
  - [ ] **AC-CON-001-004 (Go struct schema), AC-CON-001-007 (orphan), AC-CON-001-009 (ID stability), AC-CON-001-015 (performance benchmark), AC-CON-001-017 (YAML 6-field direct)** 검증 (내부 API 수준)

### T-09 `internal/constitution/dangling.go` 구현 (behavior-verified skeleton)
- **Owner-role**: implementer
- **Isolation**: worktree
- **File ownership**:
  - `internal/constitution/dangling.go`
  - `internal/constitution/dangling_test.go`
- **의존성**: T-04, T-07
- **Blocks**: 없음 (SPC-003 에서 CLI wiring 추가)
- **설명**: `ValidateRuleReferences(registry *Registry, refs []string) []string` 시그니처 + **실제 behavior 구현** (plan-audit 2026-04-25 강화). 각 ref 에 대해 registry.Get 호출, 존재하지 않으면 `"dangling reference: <ID> not found in registry"` 형식의 warning 문자열 slice 에 추가. REQ-CON-001-041 에서 behavior-verified 수준.
- **DoD**:
  - [ ] 빌드 통과 (시그니처 확정)
  - [ ] Body 가 실제 동작 구현 (빈 return 금지) — plan-audit 2026-04-25 요구
  - [ ] `@MX:NOTE` 태그로 SPC-003 wiring 계획 명시, `@MX:TODO` 제거
  - [ ] 단위 테스트 포함: `TestValidateRuleReferencesReturnsWarningForUnknownID` — AC-CON-001-011 직접 검증 (fixture call 로 ≥1 warning string 반환 확인)

---

## Phase 3: CLI 통합

### T-10 `internal/cli/constitution.go` 구현
- **Owner-role**: implementer
- **Isolation**: worktree
- **File ownership**: `internal/cli/constitution.go`
- **의존성**: T-07
- **Blocks**: T-11, T-12, T-13
- **설명**: cobra 서브커맨드 2개 (`constitution`, `constitution list`). `--zone`, `--file`, `--format` 플래그. Registry 경로 해석 (env → CLAUDE_PROJECT_DIR → cwd). table/json 포맷터. research.go 패턴 준수.
- **DoD**:
  - [ ] `go vet` + `golangci-lint` 통과
  - [ ] `runConstitutionList`는 `io.Writer` 인자로 테스트 친화
  - [ ] `--zone` 파싱 실패 시 usage 메시지 출력 후 non-zero exit
  - [ ] 레지스트리 미발견 시 명확한 에러 (REQ 목표: 개발자 유도)

### T-11 `internal/cli/root.go` 서브커맨드 등록
- **Owner-role**: implementer
- **Isolation**: 불필요 (한 줄 추가, 순차 처리)
- **File ownership**: `internal/cli/root.go`
- **의존성**: T-10
- **Blocks**: T-12, T-13
- **설명**: `rootCmd.AddCommand(newConstitutionCmd())` 한 줄 추가 + `GroupID: "tools"` 확인. research.go 패턴 따름.
- **DoD**:
  - [ ] `moai --help` 출력에 constitution 커맨드 표시
  - [ ] 기존 테스트 (root_test.go) 모두 통과
  - [ ] git diff 최소 (한 줄 추가)

### T-12 `internal/cli/constitution_test.go` 작성
- **Owner-role**: tester
- **Isolation**: worktree
- **File ownership**: `internal/cli/constitution_test.go`
- **의존성**: T-10, T-11
- **Blocks**: T-13
- **설명**: CLI 단위 테스트. `TestConstitutionListAllEntries`, `TestConstitutionListFilterFrozen`, `TestConstitutionListFilterByFile`, `TestConstitutionListJSON`, `TestConstitutionListRegistryMissing_FileNotFound`, `TestConstitutionListRegistryMissing_PermissionDenied`. 기존 misc_coverage_test.go 서브커맨드 발견 패턴 준수.
- **DoD**:
  - [ ] 커버리지 ≥ 85%
  - [ ] **AC-CON-001-001 (list output), AC-CON-001-002 (filter frozen), AC-CON-001-010 (ID sync), AC-CON-001-013 (missing/perm-denied registry)** 예비 검증 통과
  - [ ] `t.TempDir()` 사용 (CLAUDE.local.md §6 test isolation)
  - [ ] AC-013 은 file-missing 과 permission-denied 2 subtest 로 분리 (plan-audit LOW 권고 반영)

### T-13 Phase 3 통합 검증
- **Owner-role**: reviewer
- **Isolation**: 불필요
- **File ownership**: 없음
- **의존성**: T-08, T-11, T-12
- **Blocks**: T-14
- **설명**: `make build` → `./moai constitution list` 수동 실행하여 실제 registry 읽기 확인. 출력 시각 확인, 이모지/컬러 코드 적절성, --zone frozen이 7개 이상 표시되는지 확인.
- **DoD**:
  - [ ] `moai constitution list` 정상 동작 스크린샷 또는 터미널 출력 기록
  - [ ] `moai constitution list --zone frozen` 에서 SPEC §5.1 canonical 7 조항 모두 표시
  - [ ] PR-3 생성 준비 완료

---

## Phase 4: 검증 훅 + Doctor + CI

### T-14 `internal/cli/doctor.go` 체크 추가
- **Owner-role**: implementer
- **Isolation**: worktree
- **File ownership**: `internal/cli/doctor.go` (기존 파일 수정)
- **의존성**: T-13
- **Blocks**: T-15, T-16, T-17
- **설명**: `runDiagnosticChecks`에 `ConstitutionCheck` 추가 (~30 LOC). (a) registry 존재, (b) Frozen ≥ 1, (c) 중복 ID 없음, (d) orphan 경고, (e) design mirror overflow 경고. `MOAI_CONSTITUTION_STRICT=1` 환경 변수로 strict 모드.
- **DoD**:
  - [ ] `moai doctor` 출력에 Constitution 섹션 포함
  - [ ] `MOAI_CONSTITUTION_STRICT=1` 설정 시 violation에서 non-zero exit (AC-CON-001-006)
  - [ ] 기존 doctor_test.go 모두 통과 (regression 없음)
  - [ ] REQ-CON-001-020, 030 직접 구현

### T-15 Doctor 단위 테스트
- **Owner-role**: tester
- **Isolation**: worktree
- **File ownership**: `internal/cli/doctor_constitution_test.go`
- **의존성**: T-14
- **Blocks**: T-18
- **설명**: Constitution doctor check의 4가지 케이스 테스트. valid, duplicate, empty-frozen, strict mode에서의 exit code. Env 변이 테스트 (`t.Setenv`) 는 `t.Parallel()` 비활성화 + non-parallel subtest 로 격리 (plan-audit LOW 권고).
- **DoD**:
  - [ ] 커버리지 ≥ 85%
  - [ ] **AC-CON-001-006 (strict mode duplicate), AC-CON-001-014 (empty Frozen warns)** 직접 검증
  - [ ] `t.Parallel()` OFF 확인 (t.Setenv 사용 테스트)

### T-16 `internal/cli/constitution_guard.go` 작성 (Go 기반, OQ3 해소 확정)
- **Owner-role**: implementer
- **Isolation**: worktree
- **File ownership**:
  - `internal/cli/constitution_guard.go` (구현)
  - `internal/cli/constitution_guard_test.go` (단위 테스트)
- **의존성**: T-14
- **Blocks**: T-18
- **설명**: plan.md §7 OQ3 Decision 확정 — Go 기반만 채택. `moai constitution guard --base <ref>` 서브커맨드 추가. `git diff --name-only $BASE HEAD` 를 `exec.Command` 로 실행 → `.claude/rules/moai/**/*.md` 변경 hunk 에서 `[HARD]` 추가 탐지 → `zone-registry.md` 미변경 시 non-zero exit. Code fence 내부 예시는 skip. REQ-CON-001-010 구현.
- **DoD**:
  - [ ] CI에서 실행 시 지정 조건에서 non-zero exit
  - [ ] 커밋 메시지에 [skip-constitution-check] 있을 때 skip (우회 가드)
  - [ ] False positive 테스트 (code fence 내부 `[HARD]` 예시 무시)
  - [ ] **AC-CON-001-003 (CI guard detects missing registry update)** 직접 검증

### T-17 Makefile 타겟 추가
- **Owner-role**: implementer
- **Isolation**: 불필요 (단일 파일)
- **File ownership**: `Makefile`
- **의존성**: T-16
- **Blocks**: T-18
- **설명**: `constitution-check` 타겟 추가 — `moai constitution guard --base origin/main` 실행. 로컬 개발자가 커밋 전 수동 실행 가능. `make lint` 타겟에도 옵션 포함 검토.
- **DoD**:
  - [ ] `make constitution-check` 실행 가능
  - [ ] help 출력에 타겟 설명 포함

### T-18 CI workflow 통합 (OQ2 해소 확정)
- **Owner-role**: implementer
- **Isolation**: worktree (yml 수정)
- **File ownership**: `.github/workflows/ci.yml` (plan.md §7 OQ2 Decision — 기존 파일에 job 추가, 신규 파일 생성 안 함)
- **의존성**: T-15, T-16, T-17
- **Blocks**: T-19, T-20
- **설명**: CI에 `constitution-check` job 추가. 초기에는 warn 모드 (`continue-on-error: true`). 의존 SPEC 준비 후 blocking 전환.
- **DoD**:
  - [ ] PR에서 HARD 추가 + registry 미수정 시 경고 생성 (PR check 표시)
  - [ ] 정상 PR에서는 녹색 체크 유지
  - [ ] 워크플로우 문법 검증 (`actionlint` 또는 비슷한 도구)

### T-19 통합 테스트 작성
- **Owner-role**: tester
- **Isolation**: worktree
- **File ownership**: `internal/cli/constitution_integration_test.go`
- **의존성**: T-14, T-16
- **Blocks**: T-20
- **설명**: `//go:build integration` 빌드 태그. 실제 `.claude/rules/moai/core/zone-registry.md`를 입력으로 `moai constitution list` 실행 → Frozen ≥ 7 확인. `moai doctor constitution` 실행 → exit 0 확인.
- **DoD**:
  - [ ] `go test -tags=integration ./internal/cli/...` 통과
  - [ ] AC-CON-001-001, AC-CON-001-005, AC-CON-001-008, AC-CON-001-010, AC-CON-001-012 검증

---

## Phase 5: 완료 검증

### T-20 전체 Acceptance Criteria 검증
- **Owner-role**: reviewer
- **Isolation**: 불필요
- **File ownership**: 없음
- **의존성**: T-18, T-19
- **Blocks**: T-21
- **설명**: acceptance.md 기반으로 AC-001~017 체크리스트 하나씩 수동 검증. 각 AC마다 테스트 파일:테스트 함수 대응 확인.
- **DoD**:
  - [ ] 17개 AC 모두 PASS 마킹 (12 원본 + 4 INFERRED + 1 plan-audit 추가)
  - [ ] 실패 AC는 새 task로 이관 또는 SPEC follow-up issue 생성
  - [ ] acceptance.md §Definition of Done 체크리스트 모두 ✓

### T-21 Binary size / performance 회귀 확인
- **Owner-role**: reviewer
- **Isolation**: 불필요
- **File ownership**: 없음
- **의존성**: T-20
- **Blocks**: T-22
- **설명**: Phase 시작 전 `bin/moai` 크기 ↔ 완료 후 `bin/moai` 크기 비교. 50KB 증가 이하 확인 (SPEC 제약 §7). Benchmark 결과 <10ms 재확인.
- **DoD**:
  - [ ] Before/after binary size 기록
  - [ ] `go test -bench=BenchmarkLoadRegistry` 결과 기록
  - [ ] 제약 §7 준수 증빙 첨부

### T-22 PR 머지 및 SPEC 상태 업데이트
- **Owner-role**: implementer
- **Isolation**: 불필요 (git operation)
- **File ownership**: `.moai/specs/SPEC-V3R2-CON-001/spec.md` (status 필드만 수정 허용 — 본문 FROZEN)
- **의존성**: T-21
- **Blocks**: 없음 (최종 task)
- **설명**: spec.md frontmatter `status: draft → implemented`, `updated: 2026-xx-xx` 업데이트. CHANGELOG.md에 엔트리 추가. CON-002, CON-003, SPC-003, RT-005 의존 SPEC에 "CON-001 해금" 코멘트 (각 spec.md §9.1).
- **DoD**:
  - [ ] spec.md status 필드 갱신 (본문은 수정 금지)
  - [ ] CHANGELOG.md에 SPEC-V3R2-CON-001 엔트리 추가
  - [ ] 의존 SPEC 해금 커뮤니케이션 완료
  - [ ] git tag 또는 버전 마커 (선택)

---

## 병렬 실행 전략 (팀 모드)

다음 병렬 그룹은 `workflow.yaml` role profile로 동시 스폰 가능:

```
G1 Wave: T-04 (implementer A) + T-05 (implementer B) + T-06 (implementer C) 
  - 각 서로 다른 파일 쓰기, worktree isolation 필수
  - 완료 후 T-07 단일 구현자 (dependency 수렴)

G3 Wave (직렬 확정, plan-audit 2026-04-25):
  - T-10 → T-11 → T-12 순차 (T-11 이 T-10 의 cobra 시그니처에 의존)
  - 병렬 spawn 불가 — 라벨 "병렬" 제거

G5 Wave: T-15 (tester) + T-16 (implementer) + T-17 (implementer)
  - T-15는 internal/cli/ 내부, T-16은 internal/cli/constitution_guard.go, T-17은 Makefile — 파일 충돌 없음
  - worktree 3개 병렬 가능
```

**솔로 모드 권장 경로**: T-01 → T-02 → T-03 → T-04 → T-05 → T-06 → T-07 → T-08 → T-09 → T-10 → T-11 → T-12 → T-13 → T-14 → T-15 → T-16 → T-17 → T-18 → T-19 → T-20 → T-21 → T-22 (순수 직렬 실행 시 22 step).

**팀 모드 권장 경로**: G0 (직렬 3) → G1 (병렬 3) → G2 (직렬 3) → G3 (직렬 3) → G4 (직렬 2) → G5 (병렬 3) → G6 (직렬 2) → G7 (직렬 3) ≈ 약 20 논리 step (G3 가 직렬로 수정되어 기존 18 → 20).

---

## 파일 소유권 매트릭스

병렬 실행 시 충돌 방지용 — 각 파일을 수정하는 task는 최대 1개:

| 파일/경로 | Owner Task |
|-----------|------------|
| `.claude/rules/moai/core/zone-registry.md` | T-03 |
| `internal/template/templates/.claude/rules/moai/core/zone-registry.md` | T-03 |
| `internal/constitution/zone.go` | T-04 |
| `internal/constitution/rule.go` | T-05 |
| `internal/constitution/loader.go` | T-07 |
| `internal/constitution/dangling.go` | T-09 |
| `internal/constitution/dangling_test.go` | T-09 |
| `internal/constitution/testdata/*.md` | T-06 |
| `internal/constitution/zone_test.go` | T-08 |
| `internal/constitution/rule_test.go` | T-08 |
| `internal/constitution/loader_test.go` | T-08 |
| `internal/cli/constitution.go` | T-10 |
| `internal/cli/constitution_test.go` | T-12 |
| `internal/cli/root.go` | T-11 |
| `internal/cli/doctor.go` (수정) | T-14 |
| `internal/cli/doctor_constitution_test.go` | T-15 |
| `internal/cli/constitution_guard.go` | T-16 |
| `internal/cli/constitution_guard_test.go` | T-16 |
| `internal/cli/constitution_integration_test.go` | T-19 |
| `Makefile` | T-17 |
| `.github/workflows/ci.yml` | T-18 |
| `CHANGELOG.md` | T-22 |
| `spec.md` (status만) | T-22 |

---

## 요약

- Phase 수: 4 구현 + 1 검증 = **5 Phase**
- Task 수: **22**
- 병렬 그룹: **8 그룹** (G0-G7); G3 는 plan-audit 2026-04-25 후 직렬로 변경
- Worktree 필수 task: **13개** (T-04, T-05, T-06, T-07, T-08, T-09, T-10, T-12, T-14, T-15, T-16, T-18, T-19) — 병렬 실행 대상 또는 파일 쓰기 작업. 기존 "11개" 라벨은 오계산 (plan-audit LOW 권고로 재계산).
- 읽기 전용 / isolation 불필요 task: **5개** (T-01, T-02, T-11, T-13, T-20, T-21, T-22) — 하지만 T-22 는 spec.md 수정이므로 실제로는 "isolation 불필요, 쓰기 있음" 임. 순수 read-only 는 T-01, T-02, T-13, T-20, T-21 = **5개**. 기존 "3개" 라벨은 오계산 (plan-audit LOW 권고로 재계산).
