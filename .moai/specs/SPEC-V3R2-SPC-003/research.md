---
id: SPEC-V3R2-SPC-003
document: research
version: "0.1.0"
status: backfilled
created: 2026-05-10
updated: 2026-05-10
author: manager-spec (Batch 3 backfill)
related_spec: SPEC-V3R2-SPC-003
phase: plan
language: ko
---

# SPEC-V3R2-SPC-003 — Research (코드베이스 분석 + 외부 레퍼런스)

> **BACKFILL NOTICE**: 본 research.md는 Wave 5에서 PR #745로 이미 머지된 `moai spec lint` 구현(`internal/spec/lint.go` 외 4 파일)의 design rationale을 사후 정리한 retroactive 문서이다. 코드는 FROZEN 상태이며 본 문서는 **실제 채택된 설계 결정의 근거**를 외부 레퍼런스 + 사내 SPEC 의존 관계로 backfill한다. 새 요구사항은 추가하지 않는다.

---

## 1. 작성 맥락 (Why this research is retroactive)

| 사실 | 출처 |
|------|------|
| spec.md v0.1.0 | 2026-04-23 Wave 4 작성 (SPEC 본문만) |
| RED 커밋 | `44019ec4d test(spec): SPEC-V3R2-SPC-003 RED phase — failing lint tests` |
| GREEN 커밋 | `4d5699f27 feat(spec): SPEC-V3R2-SPC-003 GREEN — moai spec lint linter engine` |
| REFACTOR 커밋 | `1587e6e3a refactor(spec): SPEC-V3R2-SPC-003 REFACTOR — Karpathy simplicity + MX tags` |
| 머지 커밋 | `03146d1ae feat(spec): SPEC-V3R2-SPC-003 — moai spec lint CLI (Wave 5) (#745)` |
| 머지일 | 2026-04-30 17:51 KST (Wave 5 batch) |
| Plan-phase 산출물 부재 | spec.md + progress.md만 존재했음 — research/plan/acceptance/tasks/spec-compact/issue-body 부재 |
| 본 backfill 작성일 | 2026-05-10 (Batch 3 normalization) |

본 research는 이미 코드로 굳어진 다음 결정을 외부 레퍼런스에 매핑한다:

1. EARS modality 정규식 (Conservative pattern + lint.skip escape hatch)
2. Tarjan SCC 알고리즘 (cycle detection)
3. SARIF 2.1.0 출력 형식 (CI annotation tool 호환)
4. Rule interface + cross-SPEC interface 분리 패턴
5. lint.skip frontmatter (per-SPEC suppress 메커니즘)

---

## 2. 외부 레퍼런스 매핑

### 2.1 Linter UX 설계 — eslint / golangci-lint / ast-grep 비교

| 도구 | 출력 형식 기본 | 코드 식별자 | severity 분리 | suppress 메커니즘 |
|------|---------------|-------------|--------------|-------------------|
| ESLint (JS) | stylish (table) | rule name (`no-unused-vars`) | error/warn/info | `// eslint-disable-line` |
| golangci-lint (Go) | colored line list | linter name + check | error/warn | `//nolint:linter` |
| ast-grep | YAML / SARIF | rule id | error/warn/info | per-rule config |
| **moai spec lint** (채택) | table (default) → JSON → SARIF | finding `Code` (`CoverageIncomplete`) | error/warn/info | frontmatter `lint.skip: [CODE]` |

**채택 근거**:

- 3-tier severity (error / warn / info)는 ESLint 컨벤션 (스코어카드 식 1-tier보다 표현력 우수)
- 코드 식별자(`CoverageIncomplete`, `ModalityMalformed`)는 golangci-lint 스타일 (각 rule이 고유 ID; 사용자가 suppress 가능)
- `--strict` 플래그는 warning을 error로 승격 (ESLint `--max-warnings=0`와 동등)
- `lint.skip`은 frontmatter-scoped (line-scoped가 아님) → SPEC 본문에 주석을 남기지 않으므로 가독성 보존

**구현 매핑**: `internal/spec/lint.go` §`type Severity`, `type Finding`, `applylintSkip()` (line 205).

### 2.2 Cycle Detection — Tarjan SCC algorithm

**문헌**: Robert Tarjan, "Depth-first search and linear graph algorithms", SIAM Journal on Computing, 1972. 정통 reference로 Sedgewick *Algorithms* 4판 §4.2 + Cormen *CLRS* 3판 §22.5 (Strongly Connected Components).

**복잡도**: O(V+E) — V=SPEC 수, E=dependencies 엣지 수.

**채택 근거**:

- 자기 참조(self-loop) + 다중 cycle 동시 검출 가능
- DFS 한 번으로 SCC + topological order 동시 계산 → 향후 의존 순서 출력 기능 확장 시 재사용 가능
- spec.md §4 가정: <100 SPECs → Tarjan O(100+200) = 300 step → < 1ms (실측 시간은 의미 없음)

**대안 비교**:
| 알고리즘 | 복잡도 | 적합성 |
|----------|--------|--------|
| Tarjan SCC (채택) | O(V+E) | 다중 cycle 동시 검출, topological order 부수산출 |
| Kosaraju SCC | O(V+E), DFS 두 번 | 단순하지만 DFS 두 번 필요 |
| Naive DFS w/ visited | O(V·E) | 단일 cycle만 검출, 다중 cycle 시 누락 위험 |

**구현 매핑**: `internal/spec/dag.go` (100 LOC, Tarjan iterative variant).

### 2.3 SARIF 2.1.0 — Static Analysis Results Interchange Format

**문헌**: OASIS Static Analysis Results Interchange Format (SARIF) Version 2.1.0. https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html

**채택 근거**:

- GitHub Code Scanning UI가 SARIF 2.1.0 표준 형식을 입력으로 받음 → CI에서 `gh api /code-scanning/sarifs` 업로드 시 자동 PR annotation
- VS Code SARIF Viewer 확장도 동일 표준 사용 → 로컬 개발자도 동일 출력으로 디버깅 가능
- JSON Schema 명세가 안정적 (2026-05 기준 2.1.0이 latest)

**최소 필수 필드** (린터가 채워야 하는 minimum subset):
- `version`: "2.1.0"
- `runs[]`: 단일 element 배열
- `runs[0].tool.driver`: `name` ("moai spec lint"), `version` (moai 바이너리 버전), `informationUri`
- `runs[0].results[]`: 각 finding 하나당 element. `ruleId` (Finding.Code), `level` (`error`/`warning`/`note`), `message.text`, `locations[].physicalLocation.artifactLocation.uri`, `locations[].physicalLocation.region.startLine`

**구현 매핑**: `internal/spec/sarif.go` (165 LOC, `Report.ToSARIF()` writer).

### 2.4 EARS modality keywords — Mavin et al. (2009)

**문헌**: Alistair Mavin, Philip Wilkinson, Adrian Harwood, Mark Novak, "Easy Approach to Requirements Syntax (EARS)", IEEE International Requirements Engineering Conference, 2009.

**5가지 modality form** (EARS canonical):

| Form | Trigger | Pattern |
|------|---------|---------|
| Ubiquitous | (always) | The [system] **shall** [response] |
| Event-Driven | event | **When** [event], the [system] **shall** [response] |
| State-Driven | state | **While** [condition], the [system] **shall** [response] |
| Optional | feature presence | **Where** [feature], the [system] **shall** [response] |
| Unwanted | undesired | **If** [undesired], **then** the [system] **shall not** [response] |
| Complex (composite) | state+event 등 | **While** ..., **when** ..., the [system] **shall** ... |

**채택된 modality 정규식 (보수적 — false-negative 우선)**:

```
^(The [a-z][a-zA-Z\s]*\s+)?(SHALL|shall)\b           # Ubiquitous
^(WHEN|When)\s+.*,.*?\b(SHALL|shall)\b               # Event-driven
^(WHILE|While)\s+.*,.*?\b(SHALL|shall)\b             # State-driven
^(WHERE|Where)\s+.*,.*?\b(SHALL|shall)\b             # Optional
^(IF|If)\s+.*,.*?\b(THEN|then)\b.*?\b(SHALL|shall)\b # Unwanted
```

**채택 근거**: spec.md §8 Risks 표 row 1에서 식별한 false-positive 위험을 보수적 패턴 + `lint.skip` 탈출구로 완화. 자연어 변형(e.g., "The user can ...")은 modality 미충족으로 간주되어 `ModalityMalformed` 발화 → 작성자가 명시적 EARS 형으로 다시 쓰도록 유도.

**구현 매핑**: `internal/spec/lint.go` §`isModalityMalformed()` (line 415).

### 2.5 Frontmatter validation — gopkg.in/yaml.v3 + go-playground/validator/v10

**Go ecosystem standard**:
- `gopkg.in/yaml.v3`: YAML 1.2 컴플라이언트, 가장 널리 채택된 Go YAML 파서
- `go-playground/validator/v10`: struct tag 기반 validator, struct field에 `validate:"required,..."` 태그로 schema 강제

**채택 근거** (spec.md §4 Assumptions에서 사전 결정):

- 두 라이브러리 모두 이미 moai 모듈 그래프에 포함 (의존성 추가 없음)
- struct tag 방식은 schema가 코드와 한 곳에 위치 → drift 방지
- `validator.Struct()` 호출 시 모든 필드 검증을 한 번에 수행 → linter rule 자체는 단순

**구현 매핑**: `internal/spec/lint.go` §`type SPECFrontmatter` (line 255), `parseSPECDoc()` (line 302), `extractFrontmatter()` (line 333), `FrontmatterSchemaRule` (line 503).

---

## 3. 사내 SPEC 의존 관계

### 3.1 Hard dependencies (BLOCKING)

| 의존 SPEC | 역할 | 본 SPEC에서 어떻게 소비되는가 |
|----------|------|-------------------------------|
| SPEC-V3R2-CON-001 | Constitution Zone Registry (Frozen/Evolvable) | `ZoneRegistryRule.Check()` (lint.go:701)에서 `CONST-V3R2-NNN` ID들이 zone registry에 존재하는지 cross-reference 검증. registry 미적재 시 `DanglingRuleReference` warning 발화. |
| SPEC-V3R2-SPC-001 | Hierarchical Acceptance Criteria parser | `parseSPECDoc()`가 acceptance.md (또는 spec.md §6) 트리를 파싱할 때 SPC-001의 canonical tree를 소비. AC depth/uniqueness 검증은 SPC-001 invariant이므로 본 린터는 그것을 *trust*하고 REQ 매핑 coverage만 확인. |

### 3.2 Soft dependencies (RELATED, not blocking)

- **SPEC-V3R2-SPC-002** (@MX TAG validation): `/moai mx --verify`로 별도 검증되며 본 린터의 scope 외부.
- **skill moai-workflow-spec**: required body sections (Goal/Scope/Environment/Assumptions/Requirements/Acceptance/Constraints/Risks/Dependencies/Traceability)의 authoritative source. 본 린터는 그 중 §2.2 Out of Scope 존재만 강제 (REQ-SPC-003-009).
- **CI workflow**: `moai spec lint --strict` 실행을 PR gate로 등록할 때 본 린터의 exit code가 game changer. 별도 SPEC으로 분리.

---

## 4. 채택된 9개 Lint Rule 일람

> 코드 진실: `internal/spec/lint.go` line 394 ~ 786. 실측 결과 9개 single-SPEC rule + 1개 cross-SPEC rule = **10개 rule**이며 progress.md "9 rules" 표기는 single-SPEC만 카운트한 것이다.

| # | Rule struct | Code (Finding ID) | Scope | REQ 매핑 |
|---|-------------|-------------------|-------|----------|
| 1 | `EARSModalityRule` | `ModalityMalformed` | per-SPEC | REQ-SPC-003-003, REQ-SPC-003-050 |
| 2 | `REQIDUniquenessRule` | `DuplicateREQID` | per-SPEC | REQ-SPC-003-004 |
| 3 | `CoverageRule` | `CoverageIncomplete` | per-SPEC | REQ-SPC-003-005 |
| 4 | `FrontmatterSchemaRule` | `FrontmatterInvalid` | per-SPEC | REQ-SPC-003-006 |
| 5 | `DependencyExistsRule` | `MissingDependency` | per-SPEC | REQ-SPC-003-007 |
| 6 | `OutOfScopeRule` | `MissingExclusions` | per-SPEC | REQ-SPC-003-009 |
| 7 | `BreakingChangeIDRule` | `BreakingChangeMissingID` | per-SPEC | REQ-SPC-003-052 |
| 8 | `ZoneRegistryRule` | `DanglingRuleReference` | per-SPEC | REQ-SPC-003-010 |
| 9 | `DependencyCycleRule` | `DependencyCycle` | **cross-SPEC** | REQ-SPC-003-008 |
| 10 | `DuplicateSPECIDRule` | `DuplicateSPECID` | **cross-SPEC** | REQ-SPC-003-031 |

**Rule interface 분리** (`type Rule` line 198 vs `type crossSPECRule` line 192):
- `Rule.Check(doc *SPECDoc, _ []*SPECDoc)`: 단일 SPEC 검증 (per-file 병렬화 가능)
- `crossSPECRule.Check(allDocs []*SPECDoc)`: 트리 전체 cross-validation (DAG cycle, SPEC ID 중복 등 — 모든 SPEC을 동시에 알아야 함)

이 분리로 향후 `--parallel` 옵션 추가 시 per-SPEC rule만 worker pool에 분배하면 되며 cross-SPEC rule은 마지막에 일괄 실행.

---

## 5. 구현 sketch 사실 확인 (코드 vs spec.md)

| spec.md § | 명세 | 실제 코드 위치 | 일치 |
|-----------|------|---------------|------|
| §2.1 In Scope item 1 | `moai spec lint` CLI subcommand | `internal/cli/spec_lint.go:newSpecLintCmd()` | ✅ |
| §2.1 In Scope item 2 | EARS compliance + REQ uniqueness + 단일 modality | `EARSModalityRule` + `REQIDUniquenessRule` | ✅ |
| §2.1 In Scope item 3 | AC→REQ coverage ≥ 100% | `CoverageRule` | ✅ |
| §2.1 In Scope item 4 | Frontmatter schema 14 필드 | `SPECFrontmatter` + `FrontmatterSchemaRule` | ✅ |
| §2.1 In Scope item 5 | Dependency DAG cycle 검출 | `DependencyCycleRule` + `dag.go` Tarjan SCC | ✅ |
| §2.1 In Scope item 6 | Hierarchical AC structural validation | `parseSPECDoc()` invariants from SPC-001 | ✅ (delegated) |
| §2.1 In Scope item 7 | Zone registry cross-reference | `ZoneRegistryRule` | ✅ |
| §2.1 In Scope item 8 | Out of Scope 존재 강제 | `OutOfScopeRule` | ✅ |
| §2.1 In Scope item 9 | table / JSON / SARIF 출력 | `Report.ToJSON()` + `Report.ToSARIF()` + table writer | ✅ |
| §2.2 Out of Scope | --fix 미구현, @MX 별도 SPEC | 코드에 --fix 플래그 부재; @MX 검증 코드 없음 | ✅ (의도적 부재) |

**Drift**: 0% — 모든 In-Scope 항목이 코드로 구현됨. progress.md의 "Drift: ~0%"가 사실 확인됨.

---

## 6. Go 패키지 구조 (실측)

```
internal/spec/
├── lint.go         (816 LOC, REFACTOR 후)  — Linter struct, Rule interface, 10 rules
├── lint_test.go    (572 LOC)              — 16 TestLinter_AC* 테이블 테스트
├── dag.go          (100 LOC)              — Tarjan SCC iterative
├── sarif.go        (165 LOC)              — SARIF 2.1.0 writer
└── testdata/                              — 11 fixture SPEC 디렉터리

internal/cli/
└── spec_lint.go    (164 LOC)              — cobra subcommand (newSpecLintCmd)
```

총 sources: 1245 LOC + tests 572 LOC. progress.md "internal/spec coverage 86.6%" 검증됨.

---

## 7. 참고 출처 (External Citations)

- Mavin, A., Wilkinson, P., Harwood, A., Novak, M. (2009). *Easy Approach to Requirements Syntax (EARS)*. IEEE International Requirements Engineering Conference. — EARS 5-modality 표준의 1차 출처.
- Tarjan, R. (1972). *Depth-first search and linear graph algorithms*. SIAM Journal on Computing, 1(2), 146-160. — SCC 알고리즘 1차 출처.
- Sedgewick, R., Wayne, K. (2011). *Algorithms* (4th ed.), §4.2 Directed Graphs. — Tarjan SCC 교과서 reference.
- Cormen, T., Leiserson, C., Rivest, R., Stein, C. (2009). *Introduction to Algorithms* (CLRS, 3rd ed.), §22.5 Strongly Connected Components.
- OASIS (2020). *Static Analysis Results Interchange Format (SARIF) Version 2.1.0*. https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html
- ESLint Configuration Reference. https://eslint.org/docs/latest/use/configure/ — `--max-warnings`, suppress 컨벤션 (`// eslint-disable-line`).
- golangci-lint Documentation. https://golangci-lint.run/ — colored line list 출력, `//nolint:linter` directive.
- ast-grep documentation. https://ast-grep.github.io/ — YAML/SARIF 출력 + per-rule config 패턴.

---

## 8. 본 backfill의 한계

- **Retroactive 작성**: 본 research.md는 Wave 4 spec.md 작성 *시점*에 존재하지 않았으며, Wave 5 구현이 끝난 뒤 사후에 작성됨. 따라서 "research가 plan을 결정했다"는 인과적 주장이 아니라 "이미 채택된 결정의 외부 정당성을 보증한다"는 회고적 정당화이다.
- 외부 reference의 URL은 backfill 시점(2026-05-10) 기준 유효성 가정. 향후 link rot 발생 가능.
- `internal/spec/testdata/` 11개 fixture SPEC 디렉터리의 정확한 contents는 코드 진실(FROZEN)이므로 본 문서에서 재명세하지 않음.

---

End of research.md (backfill v0.1.0).
