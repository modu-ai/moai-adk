# Research — SPEC-V3R4-SPECLINT-DEBT-001

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-15 | manager-spec | 초기 draft. `moai spec lint` 의 8개 규칙 (6 ERROR + 2 WARNING) 의 의미 분석 + SPEC lifecycle status convention 정리 + 분류 휴리스틱. |

---

## 1. 조사 범위

본 research 는 SPEC-V3R4-SPECLINT-DEBT-001 plan 작성에 필요한 다음 사실을 정리한다:

1. `moai spec lint --strict` 가 검사하는 8개 규칙의 의미와 검증 로직.
2. SPEC lifecycle status convention (draft / planned / in_progress / implemented / completed / superseded).
3. SPEC frontmatter 7개 mandatory 필드의 의미와 추론 가능성.
4. `depends_on` / `supersedes` / `related_specs` 의 의존성 그래프 구성 방식.
5. 자동화 가능 작업 vs 수동 검토 필요 작업의 분류 휴리스틱.

본 research 는 lint 도구 자체의 소스 코드 분석은 포함하지 않는다 (블랙박스로 취급). lint 출력의 카테고리 메시지에서 규칙 의미를 역추론한다.

---

## 2. `moai spec lint` 8개 규칙 분석

### 2.1 ERROR 카테고리 (6개, CI-blocking)

#### FrontmatterInvalid

**의미**: SPEC frontmatter YAML 블록에서 mandatory 필드가 누락되었거나 형식 오류가 있음.

**검증 로직 추정**:
1. SPEC 파일 첫 `---` ~ `---` 블록을 YAML parse.
2. 다음 7개 필드 존재 여부 확인: `title`, `created`, `updated`, `phase`, `module`, `lifecycle`, `tags`.
3. 누락 시 `FrontmatterInvalid: missing field <name>` emit.

**일반 패턴**: 초기 SPEC 작성자가 frontmatter schema 정착 전에 작성한 SPEC, 또는 PR 머지 시 schema 진화 후 backfill 되지 않은 SPEC.

**해소 방법**: 누락 필드를 식별하고 본문 분석 + git log + 인접 SPEC 컨벤션으로 값 추론하여 보충.

#### MissingExclusions

**의미**: SPEC 의 `## Out of Scope` 섹션이 비어있거나 부재.

**검증 로직 추정**:
1. SPEC 본문에서 `## Out of Scope` 또는 `## Non-Goals` 헤딩 검색.
2. 해당 섹션 내부의 bullet (`-`) 또는 list 항목 개수 카운트.
3. 0개면 emit.

**일반 패턴**: SPEC 작성 시 in-scope 만 정의하고 out-of-scope 명시를 잊은 경우.

**해소 방법**: SPEC 본문 검토 후 deferred / out-of-scope 항목 최소 1개 명시.

#### MissingDependency

**의미**: SPEC frontmatter 의 `depends_on` 리스트가 존재하지 않는 SPEC ID 를 참조.

**검증 로직 추정**:
1. SPEC frontmatter 의 `depends_on` 리스트 수집.
2. 각 참조 SPEC ID 가 `.moai/specs/SPEC-<ID>/` 디렉토리로 resolve 되는지 확인.
3. 부재 시 emit.

**일반 패턴**: 초기 SPEC 작성 시 미래 SPEC 을 참조했으나 그 미래 SPEC 이 실제로는 생성되지 않은 경우, 또는 supersede 되어 ID 가 변경된 경우.

**해소 방법**: depends_on 에서 제거, repoint, 또는 stub SPEC 작성. 본 SPEC 은 제거를 선택.

#### DependencyCycle

**의미**: SPEC 의존성 그래프에서 순환 cycle 검출.

**검증 로직 추정**:
1. 모든 SPEC 의 depends_on 으로 directed graph 구성.
2. DFS 또는 topological sort 로 cycle 검출.
3. cycle 검출 시 emit (cycle 에 포함된 모든 SPEC 보고).

**일반 패턴**: 양방향 SPEC 의존성 (드물지만 RT cluster 같은 sequential 그룹에서 잘못 backlink 추가).

**해소 방법**: cycle 의 한 edge 를 제거. 일반적으로 후행 SPEC → 선행 SPEC 의 forward edge 유지.

#### ModalityMalformed

**의미**: EARS-format requirement 본문이 modality 키워드 (SHALL / MUST / WILL / SHALL NOT / MUST NOT / WILL NOT) 를 포함하지 않음.

**검증 로직 추정**:
1. SPEC `## Requirements` 섹션 내의 REQ-XXX-NNN 패턴 각각 추출.
2. 본문에 modality 키워드 grep.
3. 없으면 emit (line number 포함).

**일반 패턴**: WHERE / WHEN / IF 같은 EARS prefix 만 작성하고 modality 키워드를 빠뜨림.

**해소 방법**: 본문에 SHALL 또는 동의어 삽입. 의미 보존이 핵심.

#### CoverageIncomplete

**의미**: SPEC 의 spec.md 에서 선언된 REQ-XXX-NNN 가 같은 SPEC 의 acceptance.md 의 어떤 AC 에서도 참조되지 않음.

**검증 로직 추정**:
1. spec.md 의 `## Requirements` 섹션에서 REQ-XXX-NNN 패턴 수집.
2. acceptance.md 본문 전체를 grep 하여 각 REQ-XXX-NNN 참조 여부 확인.
3. 0건 참조 시 emit.

**일반 패턴**:
- REQ 가 명확한 의도를 가졌으나 acceptance.md 작성 시 누락 (가장 흔함).
- REQ 가 orphan placeholder 로 작성되었으나 실제 검증할 의도가 없음.
- 기존 AC 가 implicit 으로 검증하나 명시적 `Reference: REQ-XXX-NNN` 라인이 없음.

**해소 방법**: case-by-case. plan.md §5.2 결정 트리 참조.

### 2.2 WARNING 카테고리 (2개, 비차단)

#### StatusGitConsistency

**의미**: SPEC frontmatter `status` 가 git 이력에서 추론한 lifecycle status 와 불일치.

**검증 로직 추정**:
1. SPEC frontmatter `status` 값 읽기.
2. git log 또는 GitHub PR API 로 lifecycle status 추론 (plan.md §5.3 참조).
3. 불일치 시 emit (frontmatter, git inferred).

**일반 패턴**:
- SPEC 작성 후 코드 머지 완료되었으나 frontmatter status 가 draft 로 잔존.
- supersede 관계로 status 가 superseded 로 변경되어야 하나 본 SPEC frontmatter 가 미반영.

**해소 방법**: 자동화 스크립트 + 수동 검토. plan.md T-SLD-007 참조.

**Note**: SPEC-V3R4-HARNESS-001 의 follow-up 노트에서 V3R3 status-transition manager-git instruction 이 언급됨. 본 SPEC 의 T-SLD-007 자동화 스크립트가 그 instruction 의 일반화 버전이라고 볼 수 있다.

#### OrphanBCID

**의미**: SPEC frontmatter 의 `breaking: false` 인데 `bc_id` 필드가 비어있지 않음.

**검증 로직 추정**:
1. SPEC frontmatter `breaking` 값 읽기.
2. `breaking: false` 이면 `bc_id` 필드 값 확인.
3. 비어있지 않으면 emit (BC ID 가 orphan 임을 알림).

**일반 패턴**: 초기에 breaking 으로 설계되었다가 non-breaking 으로 변경되었으나 bc_id 잔존.

**해소 방법**: bc_id 필드 제거 또는 `bc_id: []` 명시.

---

## 3. SPEC Lifecycle Status Convention

`spec-workflow.md` 및 다수 SPEC 의 HISTORY 섹션 분석 결과, 다음 status convention 이 사실상 표준이다:

| Status | 의미 | git 추론 조건 |
|--------|------|---------------|
| `draft` | 초기 작성 단계 | plan PR 미머지 |
| `planned` | plan 단계 완료, run 진입 대기 | plan PR 머지, run PR 미머지 |
| `in_progress` | run 진행 중 | run PR open (미머지) |
| `implemented` | run PR 머지 완료, 코드 반영됨 | run PR 머지, sync PR 미머지 |
| `completed` | run + sync PR 모두 머지 완료, 라이프사이클 종료 | run + sync PR 모두 머지 |
| `superseded` | 다른 SPEC 으로 대체됨 | `supersedes` 또는 `superseded_by` frontmatter 명시 |
| `in_review` | 일부 SPEC 에서 사용 (planned 와 유사 의미) | plan PR open |

**Note**: 일부 V3R2 / V3R3 SPEC 은 `in_review` 를 `planned` 와 혼용. plan.md T-SLD-007 자동화 스크립트는 이를 동의어로 처리.

---

## 4. Frontmatter Mandatory 필드 의미

| 필드 | 의미 | 추론 가능성 |
|------|------|-------------|
| `title` | 사람이 읽는 SPEC 제목 | spec.md 의 H1 헤딩 텍스트로 추론 가능 |
| `created` | SPEC 최초 작성일 | `git log --reverse --follow spec.md \| head -1` 의 commit 날짜 |
| `updated` | SPEC 최종 수정일 | `git log -1 spec.md` 의 commit 날짜 |
| `phase` | 워크플로우 phase 라벨 | 인접 SPEC 참조 + supersedes 관계로 추론 |
| `module` | 영향 받는 디렉토리 또는 skill | spec.md 본문에서 언급된 파일 경로 grep 으로 추론 |
| `lifecycle` | spec-first / spec-anchored / spec-as-source | 일반적으로 spec-anchored (SPEC + 구현 동시 유지보수) |
| `tags` | 도메인 태그 (comma-separated string 또는 YAML array) | spec.md 본문 키워드 기반 |

---

## 5. 자동화 vs 수동 작업 분류

### 5.1 완전 자동화 가능

- T-SLD-002 (MissingDependency 제거): grep 으로 detection + sed/Edit 으로 제거.
- T-SLD-004 (DependencyCycle 끊기): 1건 단일 라인 Edit.
- T-SLD-005 (ModalityMalformed 키워드 삽입): 1건 단일 라인 Edit.
- T-SLD-008 (OrphanBCID 제거): 1건 단일 라인 Edit.

### 5.2 부분 자동화 + 수동 검토

- T-SLD-001 (FrontmatterInvalid): 자동으로 누락 필드 식별 + git log 로 created/updated 추론, 그러나 phase/module/lifecycle/tags 는 SPEC 본문 분석 필요.
- T-SLD-007 (StatusGitConsistency): 자동화 스크립트가 base case 처리, false-positive 수동 검토.

### 5.3 수동 분석 필수

- T-SLD-003 (MissingExclusions): SPEC 본문 의미 분석 후 out-of-scope 항목 작성.
- T-SLD-006 (CoverageIncomplete): 각 REQ 별 spec.md/acceptance.md 본문 분석 후 case-by-case 결정.

---

## 6. 위험 분석 (Research-level)

### R-RE-001: lint 도구 자체의 false-positive

`moai spec lint --strict` 가 실제로 잘못된 ERROR/WARNING 을 emit 할 가능성. 대표 케이스:
- CoverageIncomplete: AC 가 REQ 를 implicit 으로 검증하나 `Reference: REQ-XXX-NNN` 라인이 없는 경우 false-positive.
- StatusGitConsistency: 다중 PR 경로로 머지된 SPEC 의 status 추론 실패.

**Mitigation**: run-phase 시작 시 lint 도구 source code 를 한 차례 확인 (필요 시). 본 SPEC plan-phase 에서는 블랙박스 가정.

### R-RE-002: lint 규칙이 향후 강화될 가능성

본 SPEC 머지 후 새로운 lint rule 이 추가되면 또 debt 누적. 미티게이션: 별도 SPEC 으로 처리하되 본 SPEC 의 패턴 (categorical wave + automation + manual) 재활용.

### R-RE-003: 대규모 메타데이터 변경의 reviewer 부담

본 SPEC 의 run PR 이 188 SPEC 전반에 걸친 메타데이터 변경을 포함하면 reviewer 가 수천 라인 diff 를 검토해야 함. 미티게이션: plan.md §6 R8 참조 (카테고리별 5 commit 분할).

---

## 7. 외부 참조

- `.moai/specs/SPEC-V3R4-HARNESS-001/spec.md` — V3R4 series frontmatter schema reference.
- `.claude/rules/moai/workflow/spec-workflow.md` § Status Lifecycle.
- `.claude/rules/moai/workflow/mx-tag-protocol.md` § MX Tag Types.
- PR #913 (commit `2e27c14f8`) — spec-lint CI 도입 baseline.
- `gh pr list --state merged` 검색 패턴 — T-SLD-007 자동화 스크립트 데이터 소스.

---

## 8. 결론 (Plan 진입 신호)

- 본 SPEC 의 plan.md / spec.md / acceptance.md 작성에 필요한 사실적 근거는 충분하다.
- lint 도구 source code 분석은 plan-phase 에서 deferred. run-phase 시작 시 필요 시 확인.
- CoverageIncomplete 의 실제 케이스 수는 raw output 추정 25건 → run-phase 초입 재실행으로 확정.
- 본 SPEC 자체가 self-coverage 를 만족하므로 (REQ-SLD-010 → AC-SLD-010), meta-circularity issue 없음.

다음 단계: plan-auditor 호출 + Run PR 생성.
