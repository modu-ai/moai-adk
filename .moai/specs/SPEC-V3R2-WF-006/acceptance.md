# Acceptance Criteria: SPEC-V3R2-WF-006 — Output Styles Alignment (MoAI, Einstein)

- SPEC: SPEC-V3R2-WF-006
- Plan ref: `.moai/specs/SPEC-V3R2-WF-006/plan.md`
- Tasks ref: `.moai/specs/SPEC-V3R2-WF-006/tasks.md`
- 총 AC: 12 (spec.md §6 전부 승격) + Edge/DoD 확장
- Traceability: 모든 REQ(15개)에 최소 1개 AC 매핑. 역매핑은 §3 Traceability Matrix 참조.
- 성격: 본 SPEC은 "감사 테스트 + 정책 문서화" 가 본질이므로 일부 AC(precedence/fallback)는 문서 수용 + EXT-002 deferral 로 분할됨.

---

## 1. Given–When–Then Acceptance Scenarios

### AC-WF006-01 — MoAI frontmatter schema

- Maps: REQ-WF006-001, REQ-WF006-005
- **Given** `.claude/output-styles/moai/moai.md` 파일의 frontmatter
- **When** `TestOutputStylesFrontmatterSchema` 가 해당 파일을 파싱한다
- **Then** `name == "MoAI"`, `description` non-empty string, `keep-coding-instructions == true` 가 모두 만족되고, 누락/타입 불일치 시 `OUTPUT_STYLE_SCHEMA_ERROR` 접두사 에러가 emit 된다
- Test path: `internal/template/output_styles_audit_test.go::TestOutputStylesFrontmatterSchema` (subtest: `moai.md`)

### AC-WF006-02 — Einstein frontmatter schema

- Maps: REQ-WF006-001, REQ-WF006-005
- **Given** `.claude/output-styles/moai/einstein.md` 파일의 frontmatter
- **When** `TestOutputStylesFrontmatterSchema` 가 해당 파일을 파싱한다
- **Then** `name == "Einstein"`, `description` non-empty, `keep-coding-instructions == false`
- Test path: `internal/template/output_styles_audit_test.go::TestOutputStylesFrontmatterSchema` (subtest: `einstein.md`)

### AC-WF006-03 — Exactly two styles

- Maps: REQ-WF006-002
- **Given** embedded template tree `.claude/output-styles/moai/`
- **When** `TestOutputStylesExactlyTwo` 가 디렉터리를 순회한다
- **Then** 파일 수 == 2, 이름 집합 == `{"moai.md", "einstein.md"}`
- Test path: `internal/template/output_styles_audit_test.go::TestOutputStylesExactlyTwo`

### AC-WF006-04 — Template vs live byte-identity

- Maps: REQ-WF006-003, REQ-WF006-010
- **Given** `.claude/output-styles/moai/*.md` (live) 와 `internal/template/templates/.claude/output-styles/moai/*.md` (template)
- **When** `TestOutputStylesTemplateLiveParity` 실행 또는 `diff -rq` 수동 실행
- **Then** 양 디렉터리의 파일 집합 동치 + 각 파일 byte 동치. 불일치 시 `OUTPUT_STYLE_DRIFT: <path>` 에러 emit
- Test path:
  - `internal/template/output_styles_audit_test.go::TestOutputStylesTemplateLiveParity`
  - 보조 수동: `diff -rq .claude/output-styles/moai internal/template/templates/.claude/output-styles/moai` (CI `make build` 후 실행 권장)

### AC-WF006-05 — Project precedence (project=Einstein, user=MoAI → Einstein)

- Maps: REQ-WF006-004, REQ-WF006-006
- **Given** `.claude/settings.json` 의 `outputStyle: "Einstein"` AND `~/.claude/settings.json` 의 `outputStyle: "MoAI"`
- **When** Claude Code 세션이 시작된다
- **Then** Einstein 스타일이 로드된다
- Test path:
  - **문서 레벨 AC**: `.claude/rules/moai/core/settings-management.md` precedence 섹션에 예시 표 존재
  - **자동화 deferral**: SPEC-V3R2-EXT-002 Go loader 구현 시 precedence 테스트로 자동화 — 본 SPEC 범위 외
- Verification (본 SPEC): docs grep으로 precedence 표/예시 존재 확인 (reviewer 수동 체크리스트 포함)

### AC-WF006-06 — Complex precedence (user=Einstein, project=MoAI → MoAI)

- Maps: REQ-WF006-015
- **Given** user-level `outputStyle: "Einstein"` AND project-level `outputStyle: "MoAI"`
- **When** Claude Code 세션이 시작된다
- **Then** project-level 이 승리하여 MoAI 로드
- Test path:
  - **문서 레벨 AC**: `settings-management.md` 의 precedence 표 + 예시 2(본 케이스)
  - **자동화 deferral**: EXT-002

### AC-WF006-07 — Default fallback when no setting

- Maps: REQ-WF006-009
- **Given** project `.claude/settings.json` 과 user `~/.claude/settings.json` 중 어느 쪽에도 `outputStyle` key 가 명시되지 않음
- **When** Claude Code 세션이 시작된다
- **Then** hardcoded default "MoAI" 가 로드된다
- Test path:
  - **문서 레벨 AC**: `settings-management.md` fallback 섹션
  - **보조 검증**: `settings.json.tmpl` 에 `"outputStyle": "MoAI"` 라인 존재 — `grep '"outputStyle"' internal/template/templates/.claude/settings.json.tmpl` 로 수동 확인
  - **자동화 deferral**: EXT-002

### AC-WF006-08 — Unknown style fallback with warning

- Maps: REQ-WF006-008
- **Given** `.claude/settings.json` 의 `outputStyle: "NonExistent"`
- **When** Claude Code 세션이 시작된다
- **Then** "MoAI" 가 fallback 으로 로드되고, 경고 로그가 emit 된다
- Test path:
  - **문서 레벨 AC**: `settings-management.md` fallback 정책 섹션 (unknown → "MoAI" + warn)
  - **자동화 deferral**: EXT-002

### AC-WF006-09 — Schema violation rejection

- Maps: REQ-WF006-007, REQ-WF006-013
- **Given** 한 style 파일이 `keep-coding-instructions` key 를 누락했거나 boolean 이 아닌 값(`"yes"`, `1` 등)을 가진다
- **When** CI 에서 `go test ./internal/template/... -run TestOutputStylesFrontmatterSchema` 실행
- **Then** 테스트가 실패하고 `OUTPUT_STYLE_SCHEMA_ERROR: <path> (keep-coding-instructions): <reason>` 메시지 출력
- Test path: `internal/template/output_styles_audit_test.go::TestOutputStylesFrontmatterSchema` (synthetic subtest: missing/non-bool 케이스)

### AC-WF006-10 — Drift detection blocks build

- Maps: REQ-WF006-010
- **Given** template `internal/template/templates/.claude/output-styles/moai/moai.md` 또는 live `.claude/output-styles/moai/moai.md` 중 하나만 수정됨
- **When** `make build` 후 `go test ./internal/template/... -run TestOutputStylesTemplateLiveParity`
- **Then** 테스트 실패 + 메시지 `OUTPUT_STYLE_DRIFT: moai.md differs between template and live`
- Test path: `internal/template/output_styles_audit_test.go::TestOutputStylesTemplateLiveParity`
- 수동 검증 (T5): 의도적으로 1 byte 변경 → 실패 확인 → `git restore` 원복

### AC-WF006-11 — Third style addition blocked

- Maps: REQ-WF006-014
- **Given** PR 이 `internal/template/templates/.claude/output-styles/moai/foo.md` 를 추가 (스키마 검증 없이)
- **When** `go test ./internal/template/... -run TestOutputStylesExactlyTwo` 실행
- **Then** 테스트 실패 + 메시지 `OUTPUT_STYLE_UNVERIFIED: expected exactly 2 styles, got 3`
- Test path: `internal/template/output_styles_audit_test.go::TestOutputStylesExactlyTwo`
- 수동 smoke: T5 verification 단계에서 임시 파일 추가 → 실패 확인 → 원복 (명령 로그 캡처)

### AC-WF006-12 — Future Go loader consumes schema (deferral)

- Maps: REQ-WF006-012 (Optional)
- **Given** SPEC-V3R2-EXT-002 의 Go loader 가 구현된다
- **When** loader 가 output-style 파일 frontmatter 를 파싱한다
- **Then** 본 SPEC에서 정의한 3-key 스키마(`name`/`description`/`keep-coding-instructions`) 를 consume 하고 위반 시 동일 에러 포맷(`OUTPUT_STYLE_SCHEMA_ERROR`) 을 emit 한다
- Test path:
  - **본 SPEC 범위 외** — EXT-002 acceptance.md 에서 자동화
  - 본 SPEC 은 "스키마 확립 + 상수 위치 고정" 으로 preconditional 충족

---

## 2. Edge Cases

| Edge | 기대 동작 | Test |
|------|-----------|------|
| Style 파일에 추가 key (e.g. `model: claude-opus-4-7`) 존재 | Tolerate (superset 허용) — 필수 3 key 만 엄격 검증 | `TestOutputStylesFrontmatterSchema_ExtraKeyTolerated` (synthetic) — OPEN-B 결정 반영 |
| `keep-coding-instructions: True` (capitalized Python-style) | Reject — JSON/YAML standard boolean 만 허용 (`"true"`/`"false"` 소문자) | synthetic subtest |
| `keep-coding-instructions: "true"` (quoted string) | Reject — 문자열 `"\"true\""` 과 boolean `true` 구분 | synthetic subtest |
| Frontmatter 자체 누락 | Reject — `OUTPUT_STYLE_SCHEMA_ERROR: missing frontmatter` | synthetic subtest |
| 파일명이 `MoAI.md` (대문자) | 테스트 실패 — name/파일명 대응 규칙 상 소문자 관례 유지 | (현재 live는 소문자 `moai.md` — 보존 검증) |
| BOM 선두 (`EF BB BF`) | Reject — `OUTPUT_STYLE_SCHEMA_ERROR: BOM detected` | `TestOutputStylesEncoding` |
| CRLF line endings | Reject | `TestOutputStylesEncoding` |
| Live 파일 접근 불가 (sandbox/CI 환경 제약) | `t.Skip("live tree not available")` — graceful fallback | `TestOutputStylesTemplateLiveParity` |
| Symlink로 치환된 `moai.md` | follow 하지 않음, skip + error log | 수동 (본 Wave 범위 외, 후속 hardening) |
| Project `settings.json` 에 `outputStyle: ""` (빈 문자열) | default "MoAI" fallback 과 동등 취급 (문서화) | AC-07 문서 level |
| 2개 style 중 하나만 template 에 있고 live 에는 없음 | `OUTPUT_STYLE_DRIFT: <filename> exists only in template` | `TestOutputStylesTemplateLiveParity` |

---

## 3. Traceability Matrix (REQ ↔ AC)

| REQ | 커버 AC | Test 경로 |
|-----|---------|-----------|
| REQ-WF006-001 | AC-01, AC-02 | `output_styles_audit_test.go::TestOutputStylesFrontmatterSchema` |
| REQ-WF006-002 | AC-03 | `output_styles_audit_test.go::TestOutputStylesExactlyTwo` |
| REQ-WF006-003 | AC-04 | `output_styles_audit_test.go::TestOutputStylesTemplateLiveParity` |
| REQ-WF006-004 | AC-05, AC-07 | `settings-management.md` precedence 섹션 (문서 AC) |
| REQ-WF006-005 | AC-01, AC-02 | `output_styles_audit_test.go::TestOutputStylesFrontmatterSchema` |
| REQ-WF006-006 | AC-05 | 문서 AC + EXT-002 deferral |
| REQ-WF006-007 | AC-09 | `output_styles_audit_test.go::TestOutputStylesFrontmatterSchema` (synthetic) |
| REQ-WF006-008 | AC-08 | `settings-management.md` fallback 섹션 (문서 AC) + EXT-002 deferral |
| REQ-WF006-009 | AC-07 | 문서 AC + `settings.json.tmpl` grep 보조 |
| REQ-WF006-010 | AC-04, AC-10 | `output_styles_audit_test.go::TestOutputStylesTemplateLiveParity` |
| REQ-WF006-011 | (Optional — v3.1+ 범위 외) | 본 SPEC에서 "schema 재사용 의무" 로 문서화; 테스트 없음 |
| REQ-WF006-012 | AC-12 (deferral) | EXT-002 acceptance.md 에서 자동화 |
| REQ-WF006-013 | AC-09 | `output_styles_audit_test.go::TestOutputStylesFrontmatterSchema` (synthetic) |
| REQ-WF006-014 | AC-11 | `output_styles_audit_test.go::TestOutputStylesExactlyTwo` |
| REQ-WF006-015 | AC-06 | 문서 AC + EXT-002 deferral |

Optional REQ-WF006-011 ("Where future v3.x adds a 3rd style, it shall go through this SPEC's schema validation") 는 "미래 조건" 이므로 현재 Wave 1 deliverable 에 테스트 없음. `settings-management.md` 의 precedence/schema 섹션에 "3번째 style 추가 시 본 SPEC의 스키마 재검증 필수" 로 명문화함으로써 충족.

---

## 4. Test File Paths (Summary)

단위 (Go):

- `internal/template/output_styles_audit_test.go` (신규 — 4개 Test 함수 + synthetic subtests)

문서 (Markdown):

- `.claude/rules/moai/core/settings-management.md` (precedence + fallback 섹션 확장)
- `internal/template/templates/.claude/rules/moai/core/settings-management.md` (template mirror)

수동/보조:

- `diff -rq .claude/output-styles/moai internal/template/templates/.claude/output-styles/moai`
- `grep '"outputStyle"' internal/template/templates/.claude/settings.json.tmpl`
- T5 verification 시 3rd-style smoke 테스트 (임시 파일 생성 → 테스트 실패 확인 → 원복)

---

## 5. Definition of Done Checklist

### 5.1 기능 (Functional)

- [ ] 15개 REQ 중 13개(REQ-011/012 제외 — 미래 조건/deferral) AC 매핑 완료, 관련 테스트 녹색
- [ ] `output_styles_audit_test.go` 의 4개 Test 함수 모두 `go test -count=1 -race` 통과
- [ ] `.claude/rules/moai/core/settings-management.md` 에 precedence 표(project > user > default), 예시 2개, fallback 정책, schema 계약 섹션 모두 존재
- [ ] template mirror 와 live 문서 diff 없음 (`diff -q`)
- [ ] `diff -rq .claude/output-styles/moai internal/template/templates/.claude/output-styles/moai` 공백 출력

### 5.2 비기능 (Quality Gates — TRUST 5)

- [ ] **Tested**: `output_styles_audit_test.go` 4개 함수 녹색, synthetic 실패 케이스 포함, 기존 `commands_audit_test.go` 회귀 없음
- [ ] **Readable**: 테스트 상수 블록 단일 원천(`styleNameMoAI` 등), godoc 영문, golangci-lint 경고 0
- [ ] **Unified**: `gofmt`/`goimports` 적용, `commands_audit_test.go` 와 동일한 구조·네이밍 패턴 유지
- [ ] **Secured**: 파일 시스템 수정 0, 테스트는 read-only, path traversal 방어(정적 경로), symlink follow 금지
- [ ] **Trackable**: Conventional Commits + `Refs: SPEC-V3R2-WF-006` 트레일러 포함

### 5.3 워크플로우 (Process)

- [ ] `make build && make install` 수행, `moai version` commit hash = `git rev-parse HEAD`
- [ ] `go vet ./...` zero warnings
- [ ] `go test ./... -race -count=1` zero failures (단일 패키지 아닌 전체 범위에서도 회귀 없음)
- [ ] `golangci-lint run ./internal/template/...` zero issues
- [ ] 이모지 없음, 산문형 사용자 질문 없음 (AskUserQuestion 전용)
- [ ] 하드코딩 금지 — `"MoAI"`, `"Einstein"`, 파일명 리터럴이 테스트 파일 내부 const 블록 단일 원천 (CLAUDE.local.md §14)
- [ ] 템플릿 언어 중립성 (CLAUDE.local.md §15) — 문서에 특정 언어 편향 없음

### 5.4 수동 수용 (Precedence / Fallback — 자동화 불가)

- [ ] reviewer 수동 체크리스트: `settings-management.md` 에 precedence 표 존재 확인
- [ ] reviewer 수동 체크리스트: unknown style fallback 정책 문단 존재 확인
- [ ] reviewer 수동 체크리스트: schema 계약(3 key 최소 집합, 추가 key tolerate) 섹션 존재 확인
- [ ] 3rd-style smoke 테스트 수행 로그 캡처 (PR 코멘트 또는 커밋 본문)

### 5.5 롤백 준비

- [ ] 본 SPEC 변경은 test-only + docs-only (프로덕션 코드 수정 없음) — `git revert` 로 무손실 철회 가능
- [ ] output style body 파일(`moai.md`, `einstein.md`)은 본 SPEC에서 수정 없음 — 롤백 시 사용자 정의 변경 보존
- [ ] Phase별 단일 커밋 유지 (CLAUDE.local.md §4 Commit Message Format)

### 5.6 Open Question 해소

- [ ] OPEN-A (live tree 경로 탐색 전략) — `runtime.Caller(0)` 기반 해결 또는 대안 결정을 PR 설명에 기록
- [ ] OPEN-B (추가 key tolerate 정책) — `settings-management.md` 에 명문화
- [ ] OPEN-C (오류 코드 emit 위치) — "Go test failure message only, Wave 1 범위" 로 acceptance.md 명시 완료
- [ ] OPEN-D (precedence 자동화 분할) — EXT-002 에서 자동화하기로 결정 기록
- [ ] OPEN-E (명시적 MoAI vs 생략 default 구분) — `settings-management.md` 에 두 케이스 동일 효과 명시

### 5.7 Wave 1 조율

- [ ] CON-001 의 `settings-management.md` 편집 계획 확인 — 충돌 시 CON-001 merge 후 rebase
- [ ] EXT-001 / WF-001 과 파일 겹침 0 재확인
- [ ] PR 본문에 Wave 1 sibling SPEC 링크 포함

---

End of Acceptance.
