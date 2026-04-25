# SPEC-V3R2-CON-001 Implementation Plan

> FROZEN/EVOLVABLE 영역 코드화 구현 로드맵
> 최종 갱신: 2026-04-24
> 상태: ready (plan-auditor 통과)
> 종속 SPEC: 없음 (루트). 이 SPEC이 구현되면 CON-002, CON-003, SPC-003, RT-005가 해금됨.

---

## 0. 계획 개요 (Executive Summary)

본 계획은 SPEC-V3R2-CON-001을 4개 Phase로 분해한다. 핵심 산출물은 (a) 손으로 작성되는 단일 zone registry 마크다운, (b) 이를 로드·검증하는 Go 패키지 `internal/constitution`, (c) registry를 조회하는 `moai constitution list` CLI, (d) CI 가드레일 (신규 HARD 규칙 등록 누락 거부)이다.

이 SPEC은 **코드화 패스** (annotation pass)이며 **어떤 HARD 조항도 재작성하지 않는다** — 모든 조항은 verbatim 보존된다. 조항 개정 프로토콜(5계층 안전 게이트)은 SPEC-V3R2-CON-002의 영역이다.

- Phase 수: **4**
- 예상 산출 파일: `zone-registry.md` 1개 (템플릿 + 로컬 트윈 2 복사본) + Go 구현 3개 (`zone.go`, `rule.go`, `loader.go`) + CLI 1개 (`constitution.go`) + doctor 훅 1개 + CI 가드 스크립트 1개
- 핵심 의존: cobra (이미 사용), goldmark 또는 직접 파서 (의존 최소화 검토 필요 → OPEN QUESTION 1)

---

## 1. Phase 분해

### Phase 1: Zone Registry 마크다운 초안 작성 (Annotation Pass)

**목표**: `.claude/rules/moai/core/zone-registry.md`를 손으로 작성하여 Wave 2/3 SPEC이 의존할 수 있는 **단일 진실 공급원**을 확립한다.

**구현 단계**:
1. 4개 load-bearing 규칙 파일을 스캔하여 모든 `[HARD]` 조항을 식별
   - `CLAUDE.md` (§1 Hard Rules + §7 Development Safeguards + §8 User Interaction)
   - `.claude/rules/moai/core/moai-constitution.md` (§MoAI Orchestrator, §Response Language, §Parallel Execution, §Opus 4.7, §Output Format, §Worktree Isolation, §Quality Gates, §MX Tag Quality Gates, §URL Verification, §Tool Selection, §Error Handling, §Security Boundaries, §Lessons, §Agent Core Behaviors 1-6)
   - `.claude/rules/moai/core/agent-common-protocol.md` (User Interaction Boundary, Language Handling, Output Format, MCP Fallback, Agent Invocation Pattern, Background Agent Execution, Tool Usage Guidelines, Time Estimation)
   - `.claude/rules/moai/design/constitution.md` (v3.3.0 기존 Zone 마커 — 미러링만 수행, 원본은 수정하지 않음)
2. 각 HARD 조항에 `CONST-V3R2-NNN` ID 부여 (0-padded 3자리)
   - 001-050: 위 4개 파일에서 발견된 pre-existing 조항
   - 051-099: design constitution 미러 엔트리
   - 100+: 향후 신규 추가용 (Phase 1에서는 비워둠)
3. master-v3 §1.3 FROZEN 7대 불변 조항을 반드시 Frozen zone으로 매핑 확인
   - SPEC+EARS format, TRUST 5, @MX TAG protocol, 16-language neutrality, Template-First discipline, AskUserQuestion monopoly, Claude Code substrate
4. 각 엔트리에 필드 6개 채움: `id`, `zone` (Frozen|Evolvable), `file`, `anchor` (`#섹션명` 또는 `L45` 라인 번호), `clause` (verbatim), `canary_gate` (boolean)
5. 파일 상단 HISTORY 섹션 및 사용 가이드 섹션 포함
6. 템플릿 트윈 동기화: `internal/template/templates/.claude/rules/moai/core/zone-registry.md`

**생성/수정 파일**:
- 생성: `.claude/rules/moai/core/zone-registry.md`
- 생성: `internal/template/templates/.claude/rules/moai/core/zone-registry.md` (Template-First 규율 준수)

**테스트 전략**:
- 수동 검수: 4개 원본 규칙 파일을 grep `[HARD]`로 훑어 registry에 누락된 조항이 없는지 대조
- 스크립트 검증: Phase 2에서 생성할 `go test` 가 pre-existing 조항 수를 카운트하여 threshold(≥30) 만족 여부 검사

**롤아웃**: Phase 1은 마크다운 단독 변경이므로 독립 커밋. Go 코드 변경 없음 → 바이너리 영향 0.

**롤백**: 단일 파일 revert. 리스크 최소.

---

### Phase 2: Go 타입 + 로더 구현 (`internal/constitution/`)

**목표**: registry를 읽어 Go 타입으로 변환하는 파서와 검증기를 제공한다. CLI/doctor/SPEC linter가 공통 사용할 API 계층.

**구현 단계**:
1. `internal/constitution/zone.go` — `Zone` enum 타입 정의
   - `type Zone uint8` (master-v3 §4 Layer 1 code listing + 제약조건 §7 준수)
   - 상수 2개: `ZoneFrozen Zone = iota` (=0), `ZoneEvolvable` (=1)
   - `func (z Zone) String() string` — "Frozen"/"Evolvable" 렌더링
   - `func ParseZone(s string) (Zone, error)` — 대소문자 무시 파싱, 미지 값은 에러
2. `internal/constitution/rule.go` — `Rule` 구조체 + registry 엔트리 스키마
   - 필드: `ID string`, `Zone Zone`, `File string`, `Anchor string`, `Clause string`, `CanaryGate bool`, `Orphan bool` (REQ-040 지원)
   - 유효성 검증 메서드 `func (r Rule) Validate() error` — 빈 필드, 잘못된 ID 형식 거부
   - ID 정규표현식: `^CONST-V3R2-\d{3,}$`
3. `internal/constitution/loader.go` — 마크다운 파서
   - `func LoadRegistry(path string) (*Registry, error)` — 파일을 읽고 엔트리 테이블을 파싱
   - `Registry` 구조체: `Entries []Rule`, `lookup map[string]int` (ID→index)
   - 파싱 전략: goldmark 대신 자체 라인 기반 파서 채택 (OPEN QUESTION 1 해소 근거는 §3에서 논의)
   - 중복 ID 탐지, ID 정규표현식 검증, 소스 파일 존재 확인 (비존재 시 `Orphan=true`, 에러로 중단하지 않음 — REQ-040)
   - `func (r *Registry) Get(id string) (Rule, bool)` — O(1) 조회
   - `func (r *Registry) FilterByZone(z Zone) []Rule` — REQ-012 지원
4. 단위 테스트: `internal/constitution/{zone,rule,loader}_test.go`
   - Zone 파싱/렌더링 테이블 기반 테스트
   - Rule validate 양성/음성 케이스
   - 로더: 유효 registry, 중복 ID, 빈 zone, orphan 파일, 형식 오류 각 케이스

**생성/수정 파일**:
- 생성: `internal/constitution/zone.go`
- 생성: `internal/constitution/rule.go`
- 생성: `internal/constitution/loader.go`
- 생성: `internal/constitution/zone_test.go`
- 생성: `internal/constitution/rule_test.go`
- 생성: `internal/constitution/loader_test.go`
- 생성: `internal/constitution/testdata/valid_registry.md`
- 생성: `internal/constitution/testdata/duplicate_ids.md`
- 생성: `internal/constitution/testdata/orphan_reference.md`

**패키지 구조 결정**:
- 경로 `internal/constitution` — 현 SPEC은 내부 도구용, 외부 모듈 공개 계획 없음 (master-v3 §4 Layer 1 언급과 일치)
- `internal/config`와 분리: config는 runtime 설정, constitution은 규칙 메타데이터 → 책임 경계 명확화
- 순환 의존 금지: `internal/constitution`은 stdlib + 필요시 `errors`/`fmt`만 의존

**테스트 전략**:
- 패키지 단위 커버리지 ≥ 85% (TRUST 5 Tested 준수)
- Table-driven tests (Go 컨벤션 `CLAUDE.local.md §6` 준수)
- 로드 성능 벤치마크: `BenchmarkLoadRegistry` → 200 엔트리 cold load < 10ms 검증 (제약 §7)

**롤아웃**: Phase 1이 merge된 후 별도 커밋. Phase 2만으로는 CLI/doctor 변화 없음 → 사용자 관찰 변화 없음, 안전.

**롤백**: `internal/constitution/` 디렉토리 삭제 + `go.mod`/`go.sum` 의존성 변화 없음 확인 후 복원.

---

### Phase 3: CLI 통합 (`moai constitution list`)

**목표**: 개발자가 registry를 터미널에서 읽을 수 있도록 CLI 서브커맨드 제공. 필터 옵션 `--zone` 지원.

**구현 단계**:
1. `internal/cli/constitution.go` 생성 — research.go 패턴 따름
   - `func newConstitutionCmd() *cobra.Command` — 루트 서브커맨드 (`Use: "constitution"`, `GroupID: "tools"`)
   - `func newConstitutionListCmd() *cobra.Command` — `list` 서브커맨드
     - Flag: `--zone frozen|evolvable` (REQ-012)
     - Flag: `--file <path>` (제약 §7의 registry 가독성 보조)
     - Flag: `--format table|json` (기본 table)
   - `runConstitutionList(w io.Writer, projectDir string, zoneFilter *Zone, fileFilter string, format string) error` — 순수 함수, 테스트 친화
2. `internal/cli/root.go`에 `rootCmd.AddCommand(newConstitutionCmd())` 등록
3. Registry 위치 해석 순서: `$MOAI_CONSTITUTION_REGISTRY` → `$CLAUDE_PROJECT_DIR/.claude/rules/moai/core/zone-registry.md` → 현재 디렉토리 기준 같은 경로
4. Output 포맷:
   - table: ID, Zone, File, Anchor (Clause는 -v 옵션에서만 표시 → 터미널 가독성)
   - json: 전체 엔트리 직렬화 (자동화용)

**생성/수정 파일**:
- 생성: `internal/cli/constitution.go`
- 생성: `internal/cli/constitution_test.go`
- 수정: `internal/cli/root.go` (1 줄 AddCommand 호출 추가)

**테스트 전략**:
- `TestConstitutionListAllEntries` — registry fixture로 전체 렌더링 검증
- `TestConstitutionListFilterFrozen` — `--zone frozen` 플래그 동작 (AC-02 매핑)
- `TestConstitutionListFilterByFile` — `--file` 필터
- `TestConstitutionListJSON` — JSON 출력 스키마
- `TestConstitutionListRegistryMissing` — 파일 부재 시 non-zero 종료 + 명확한 에러 메시지
- 기존 `misc_coverage_test.go` 패턴 따라 서브커맨드 발견 테스트 추가

**롤아웃**: Phase 1+2 merge 후. 사용자 입장에서 신규 커맨드 추가만 — 기존 동작 무영향.

**롤백**: `constitution.go` + `root.go`의 AddCommand 한 줄 revert.

---

### Phase 4: 검증 훅 + Doctor 통합 + CI 가드

**목표**: registry가 코드베이스와 동기화된 상태를 유지하도록 자동 검증 계층 추가. REQ-010, REQ-020, REQ-030 수행.

**구현 단계**:
1. **Doctor 서브체크**: `moai doctor constitution`
   - `internal/cli/doctor.go`의 `runDiagnosticChecks` 체크 목록에 `ConstitutionCheck` 추가
   - 검사 항목: (i) registry 파일 존재, (ii) Frozen zone 엔트리 ≥ 1개 (REQ-020), (iii) 중복 ID 없음, (iv) Orphan 엔트리 경고 리스팅
   - `MOAI_CONSTITUTION_STRICT=1` 환경 변수 처리 — 기본 warn, strict일 때 non-zero exit (REQ-030)
2. **CI 가드 스크립트**: `scripts/constitution_guard.sh`
   - 목적: PR에서 `.claude/rules/moai/`의 `[HARD]` 조항이 추가되었는데 `zone-registry.md`가 수정되지 않았다면 실패
   - 동작:
     1. `git diff --name-only $BASE HEAD` 로 변경 파일 수집
     2. `.claude/rules/moai/**/*.md` 중 `[HARD]` 추가된 hunk를 grep
     3. `zone-registry.md`가 같은 diff에 포함되지 않았다면 실패 메시지 + 변경된 HARD 리스트 출력
   - Makefile 타겟 `make constitution-check` 추가
3. **SPEC linter 훅 stub**: 향후 SPEC-V3R2-SPC-003에서 REQ-041 구현하도록 TODO 주석 + `@MX:TODO` 태그 붙임
   - `internal/constitution/dangling.go` (skeleton만) — `func ValidateRuleReferences(registry *Registry, refs []string) []string` 시그니처 확정, 실제 호출부는 SPC-003에서 연결
4. **Integration test**: `internal/cli/constitution_integration_test.go`
   - 실제 `.claude/rules/moai/core/zone-registry.md`를 테스트 하네스 입력으로 사용 (`//go:build integration`)
   - `moai constitution list` 실행 → Frozen 엔트리 ≥ 7개 확인 (AC-01 매핑)

**생성/수정 파일**:
- 수정: `internal/cli/doctor.go` (체크 1개 추가 — ~30 LOC)
- 생성: `internal/cli/doctor_constitution_test.go`
- 생성: `scripts/constitution_guard.sh`
- 수정: `Makefile` (`constitution-check` 타겟 추가)
- 생성: `internal/constitution/dangling.go` (skeleton)
- 생성: `internal/cli/constitution_integration_test.go`
- 수정: `.github/workflows/*.yml` 또는 `.github/workflows/ci.yml` (constitution-check 잡 추가 — OPEN QUESTION 2)

**테스트 전략**:
- Doctor 체크 단위 테스트 (valid, duplicate, empty, strict mode 변형)
- CI 가드 스크립트를 bash-shellspec 대신 Go 통합 테스트로 래핑하여 크로스 플랫폼 일관성 확보 검토 (OPEN QUESTION 3)
- Integration test: `go test -tags=integration ./internal/cli/...`

**롤아웃**: Phase 1-3 merge 후. CI 가드는 기본 warn 모드로 도입 → 1 sprint 관찰 후 blocking 전환 (점진 롤아웃).

**롤백**:
- Doctor 체크: `runDiagnosticChecks` 체크 목록에서 1줄 제거
- CI 가드: 워크플로우 잡 제거 + Makefile 타겟 제거
- Dangling skeleton: `internal/constitution/dangling.go` 삭제 (SPC-003 구현 전까지 기능 미접속)

---

## 2. 롤아웃 시퀀스 (Incremental Shipping)

다음 순서로 main에 배포하되 각 단계는 독립된 PR:

1. **PR-1 (Phase 1)**: zone-registry.md 초안 + 템플릿 트윈. 코드 변경 없음 → zero-risk merge.
2. **PR-2 (Phase 2)**: `internal/constitution` Go 패키지 단독. CLI/사용자 가시 영향 없음.
3. **PR-3 (Phase 3)**: `moai constitution list` CLI. 순수 신규 기능, 기존 동작 무영향.
4. **PR-4a (Phase 4a)**: doctor 서브체크 (warn 모드). 사용자에게 경고만, blocking 아님.
5. **PR-4b (Phase 4b)**: CI 가드 스크립트 (warn 모드). 1 sprint 관찰.
6. **PR-4c (Phase 4c)**: strict 모드 활성화 (`MOAI_CONSTITUTION_STRICT=1` 기본값, blocking). 의존 SPEC (CON-002, SPC-003) 준비 완료 시점.

각 PR은 직전 PR merge 완료를 전제로 한다. PR-1과 PR-2는 이론상 병렬 가능하지만 PR-2 테스트가 PR-1 fixture에 의존하지 않도록 `testdata/` 별도 관리 필요.

---

## 3. 기술적 결정 사항

### 3.1 Markdown 파싱 전략

**결정**: goldmark 등 외부 파서 미사용, 자체 라인 기반 파서 채택.

**근거**:
- Registry는 고정 스키마 (파이프 테이블 또는 `key: value` 블록 리스트) — 전체 CommonMark 지원 불필요
- Binary size 제약 §7 "<50KB 증가" 준수 — goldmark 추가 시 ~300KB 증가 위험
- 로더 성능 제약 §7 "<10ms for 200 entries" 준수 — 순수 문자열 split이 AST 생성보다 빠름
- 의존 최소화로 공급망 공격 표면 축소

**대안 거부 근거**: `go.mod`에 goldmark 추가 시 다른 SPEC이 markdown 파싱을 필요로 할 때까지 지연. 지금은 YAGNI 원칙 적용.

### 3.2 Registry 포맷 최종 선택

**결정**: YAML list embedded in markdown code fence.

```markdown
## Entries

```yaml
- id: CONST-V3R2-001
  zone: Frozen
  file: .claude/rules/moai/workflow/spec-workflow.md
  anchor: "#phase-overview"
  clause: "SPEC+EARS format ..."
  canary_gate: true
- id: CONST-V3R2-002
  ...
```
```

**근거**:
- 사람 읽기 쉬움 (마크다운 contextual narrative 가능) + 기계 파싱 용이 (`gopkg.in/yaml.v3` 이미 프로젝트 의존성)
- 파이프 테이블은 `clause` verbatim 보존 시 줄바꿈 이스케이프 필요 → 취약
- YAML frontmatter 단독 사용 시 narrative 배치 불가

### 3.3 Zone 타입 underlying representation

**결정**: `type Zone uint8` (SPEC 제약 §7 직접 준수).

**근거**: master-v3 §4 Layer 1 code listing과 일치. 향후 Zone 값 확장 (예: `ZoneFrozenConstitutional`, `ZoneEvolvableHeuristic`) 여유 공간 확보 — 그러나 이번 SPEC은 2값 제한 유지 (REQ-003).

---

## 4. 의존성 그래프 (Phase 간)

```
Phase 1 (Registry MD)
    |
    v
Phase 2 (Go package) --+
    |                   |
    v                   |
Phase 3 (CLI list) ----+ 공유 의존
    |                   |
    v                   v
Phase 4 (Doctor + CI Guard)
```

Phase 1 → 2 강결합 (타입이 registry 스키마 반영). Phase 3 → 2 강결합. Phase 4 → 2+3 병렬 가능.

---

## 5. 리스크 및 완화 (§8 확장)

SPEC §8의 리스크에 더해 구현 수준 리스크:

| 리스크 | 영향 | 완화책 |
|--------|------|--------|
| Phase 1 annotation 오류 (HARD 누락/오분류) | 중 | Phase 1 PR에 수동 grep 명령 결과 첨부 요구 + 2명 이상 리뷰 |
| Zone 타입 호환성 — 향후 3번째 Zone 필요 시 breaking | 중 | `Zone` 타입 패키지 외부 노출 제한 (internal/), API는 `String()` 기반 → 내부 값 변경 허용 |
| CI 가드 false positive (문서성 HARD 예시가 실제 조항으로 오인) | 중 | Phase 1에서 예시 HARD는 4 space indent 또는 code fence 내에 배치 — guard 스크립트는 code fence 내부 skip |
| `moai doctor constitution` 성능 — 모든 `moai doctor` 호출에서 registry 로드 | 저 | 200 엔트리 < 10ms 이미 제약 (§7), doctor 전체 실행시간 대비 무시 가능 |
| 템플릿 트윈 drift | 중 | CLAUDE.local.md §2 Template-First rule 적용, `make build` 이후 diff 확인 |

---

## 6. 테스트 전략 종합

| 계층 | 도구 | 커버리지 목표 |
|------|------|---------------|
| Zone enum | `go test` table-driven | 100% (간단) |
| Rule struct | `go test` table-driven | 100% |
| Registry loader | `go test` + testdata fixtures | ≥ 90% |
| CLI command | `go test` with `cobra.Command.ExecuteC` | ≥ 85% |
| Doctor integration | `go test` | ≥ 85% |
| CI guard script | `go test -tags=integration` 래핑 | critical path 100% |
| Performance | `go test -bench` | <10ms 임계 통과 |

전체 TRUST 5:
- **Tested**: 위 표의 커버리지 목표 준수
- **Readable**: Go 관용 네이밍, exported 타입에 godoc
- **Unified**: gofmt + goimports, 기존 `internal/cli/` 패턴 일관
- **Secured**: registry 파일 경로 traversal 방지 (`filepath.Clean` + `projectDir` scope 제한)
- **Trackable**: 모든 커밋 메시지에 `SPEC-V3R2-CON-001` 참조

---

## 7. OPEN QUESTIONS

이 SPEC에서 해소되지 않아 구현 착수 전 또는 중에 결정 필요한 항목:

1. **[OPEN] Registry 파서 라이브러리**: §3.1에서 자체 파서 선호하나, `gopkg.in/yaml.v3`가 이미 프로젝트에 있다면 YAML block 추출 후 yaml.v3로 unmarshal하는 hybrid 접근이 더 단순할 수 있음. 착수 시 `go mod why gopkg.in/yaml.v3`로 확인 필요.
2. **[OPEN] CI 워크플로우 파일 위치**: `.github/workflows/` 하위 기존 워크플로우 구조 파악 후 — (a) 기존 `ci.yml`에 job 추가 vs (b) 전용 `constitution-check.yml` 신설. 기존 관례 따름.
3. **[OPEN] CI 가드 구현 언어**: bash vs Go. Windows CI 지원 여부에 따라 결정 — 현재 프로젝트 CI가 Linux-only라면 bash, cross-platform이면 Go `cmd/constitution-check` 작성. CLAUDE.local.md §6 test isolation 규율과 교차 참조 필요.
4. **[OPEN] REQ-CON-001-021 "design subsystem mirror" 상세 범위**: design/constitution.md의 어느 하위 조항까지 미러링할지 (§2 Zones 리스트만인지, §5 Safety Architecture 5-Layer까지 포함할지) SPEC §2.1 "design-subsystem FROZEN clauses"의 문자적 해석으로는 양쪽 모두 가능 — Phase 1 annotation 시 수동 판단 후 주석에 근거 기록.
5. **[OPEN] CONST-V3R2-NNN ID 번호 할당 순서**: 발견 순서 vs 파일 순서 vs Zone 우선. 제약 §7은 "첫 50 IDs는 pre-existing에 예약"만 규정. Phase 1 착수 시 규칙 고정 후 HISTORY에 기록.
6. **[OPEN] `CanaryGate` 필드의 기본값**: Frozen은 대부분 true여야 하나 (CON-002 canary 게이트 대상), Evolvable은 false인지 case-by-case인지 SPEC 미지정. 임시 휴리스틱: Frozen=true, Evolvable=false 통일 후 CON-002에서 재검토.

---

## 8. 완료 기준 체크리스트

- [ ] Phase 1: zone-registry.md 작성 + 템플릿 트윈 (Frozen 엔트리 ≥ 7, 전체 엔트리 ≥ 30)
- [ ] Phase 2: `internal/constitution/` 패키지 커버리지 ≥ 85%
- [ ] Phase 3: `moai constitution list` + `--zone frozen` 동작
- [ ] Phase 4: `moai doctor constitution` + CI 가드 (warn → blocking 전환 준비)
- [ ] AC-CON-001-01 ~ 12 모두 통과 (acceptance.md 참조)
- [ ] TRUST 5 quality gates (go vet, golangci-lint, go test -race, coverage ≥ 85%)
- [ ] Binary size 증가 < 50KB (bin/moai 전후 비교)
- [ ] Registry cold load < 10ms (benchmark 결과 첨부)
- [ ] 문서 동기화: CLAUDE.local.md §17 (해당 없음, docs-site 무영향), SPEC-V3R2-CON-002 해금 noti
- [ ] OPEN QUESTION 1-6 모두 해소 또는 후속 SPEC으로 이관
