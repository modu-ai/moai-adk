# Tasks: SPEC-V3R2-WF-006 — Output Styles Alignment (MoAI, Einstein)

- SPEC: SPEC-V3R2-WF-006
- Plan ref: `.moai/specs/SPEC-V3R2-WF-006/plan.md`
- Methodology: TDD (RED → GREEN → REFACTOR)
- Parallelism: T1 / T2 파일 독립. T3는 T1·T2의 감사 테스트와 independent. T4 통합만 sequential.

---

## 1. Task List

### T0. Precheck (block everything else)

- Owner role: analyst
- Files: read-only
- Scope:
  - [ ] `diff -rq .claude/output-styles/moai internal/template/templates/.claude/output-styles/moai` 공백 출력 확인
  - [ ] `head -5 .claude/output-styles/moai/moai.md` frontmatter 3 key 확인 (`name`, `description`, `keep-coding-instructions`)
  - [ ] `head -5 .claude/output-styles/moai/einstein.md` frontmatter 확인
  - [ ] `grep '"outputStyle"' .claude/settings.json internal/template/templates/.claude/settings.json.tmpl` 기본값 `"MoAI"` 일치
  - [ ] `make build && make install && moai version` commit hash == `git rev-parse HEAD` (MEMORY.md Hard Constraint)
  - [ ] `go test ./internal/template/... -count=1` baseline green
  - [ ] plan §Open Questions OPEN-A 후보 2(`runtime.Caller`) 로컬 smoke: 간단한 소스 파일에서 `..` 역산 가능 여부 확인
- DoD: OPEN-A 접근 결정 + 빌드/테스트 baseline green + 현재 template/live byte-identity 확인 완료.

### T1. Schema audit test — frontmatter validation (Phase 1)

- Owner role: tester + implementer (TDD)
- Files owned:
  - `internal/template/output_styles_audit_test.go` (신규)
- Scope:
  - [ ] RED: 스케치로 실패 메시지 기대 문자열 테이블 (`OUTPUT_STYLE_SCHEMA_ERROR`) 먼저 작성 (임시 skip 후 활성화)
  - [ ] GREEN — `TestOutputStylesFrontmatterSchema`:
    - embedded `fs.Sub(fsys, ".claude/output-styles/moai")` 로 파일 순회
    - 각 파일에 대해 `parseFrontmatterAndBody` (동일 패키지 unexported util 재사용) 호출
    - `name` 존재 및 non-empty
    - `description` 존재 및 non-empty
    - `keep-coding-instructions` 존재 + 값 `"true" | "false"` 만 허용 (그 외 → `OUTPUT_STYLE_SCHEMA_ERROR`)
    - `name == "MoAI"` ↔ `keep-coding-instructions == "true"`
    - `name == "Einstein"` ↔ `keep-coding-instructions == "false"`
  - [ ] GREEN — `TestOutputStylesEncoding`:
    - `bytes.Contains(data, []byte{0x0d})` → CR 없음
    - `utf8.Valid(data)` true
    - BOM(`EF BB BF`) 선두 없음
  - [ ] REFACTOR:
    - 리터럴 상수 블록: `styleNameMoAI = "MoAI"`, `styleNameEinstein = "Einstein"`, `keyName = "name"`, 등 단일 원천 (CLAUDE.local.md §14 하드코딩 금지)
    - 에러 메시지는 `fmt.Errorf("OUTPUT_STYLE_SCHEMA_ERROR: %s (%s): %s", path, key, reason)` 포맷 통일
- Dependencies: T0
- DoD:
  - REQ-WF006-001, 005, 007, 013 구현
  - `go test ./internal/template/... -run TestOutputStylesFrontmatterSchema -count=1 -race` green
  - 테이블 드리븐 + `t.Run(tt.name, ...)`
  - `t.Parallel()` 적용
  - 이모지/XML user-facing 태그 없음

### T2. Count + BC guard test (Phase 5 merged into Phase 1)

- Owner role: tester + implementer
- Files owned:
  - `internal/template/output_styles_audit_test.go` (T1에서 생성한 파일에 함수 추가)
- Scope:
  - [ ] GREEN — `TestOutputStylesExactlyTwo`:
    - `fs.ReadDir(fsys, ".claude/output-styles/moai")` 결과 길이 == 2
    - 파일명 집합 == `{"moai.md", "einstein.md"}`
    - 초과 시 `OUTPUT_STYLE_UNVERIFIED: expected exactly 2 styles, got N` 에러 메시지
  - [ ] REFACTOR: 허용 파일명 집합을 T1 상수 블록에 통합
- Dependencies: T1 (동일 파일, 같은 커밋 범위로 묶어도 무방)
- Parallel: T1 내에서 함수 추가이므로 별도 병렬화 안 함 (같은 PR/커밋)
- DoD:
  - REQ-WF006-002, 014 구현
  - `TestOutputStylesExactlyTwo` green
  - 임의로 `tmp-style.md` 를 embedded fs 외부에서 상상하는 경우 테스트가 빨간불로 떨어지는지 합성 입력으로 확인 (subtest)

### T3. Drift guard test — template vs live parity

- Owner role: tester + implementer
- Files owned:
  - `internal/template/output_styles_audit_test.go` (동일 파일 내 `TestOutputStylesTemplateLiveParity` 추가)
- Scope:
  - [ ] OPEN-A 결정 반영 (plan 제안: `runtime.Caller(0)` → `filepath.Dir` → `../..` → `.claude/output-styles/moai/`)
  - [ ] GREEN:
    - live dir 파일 목록과 embedded dir 파일 목록 집합 동치
    - 각 파일명에 대해 `os.ReadFile(livePath)` vs `fs.ReadFile(fsys, templatePath)` byte 비교
    - 불일치 시 `OUTPUT_STYLE_DRIFT: %s differs between template and live` 포맷 에러
  - [ ] 로컬 리포지토리가 아닌 임시 checkout(CI clone 등)에서도 동작하도록 프로젝트 루트 탐색 resilience 확보
  - [ ] live dir 접근 실패 시 `t.Fatal` 대신 `t.Skip("live tree not available, CI-only check")` 로 graceful fallback — OPEN-A 최종 결정에 따라 조정
- Dependencies: T1 (동일 파일), T0 (경로 전략 확정)
- Parallel: T1/T2와 파일 공유이나 함수 분리로 병렬 편집은 conflict-free
- DoD:
  - REQ-WF006-003, 010 구현
  - `TestOutputStylesTemplateLiveParity` green (현재 byte-identical 상태에서 통과)
  - 의도적으로 live 파일 1 byte 변경 후 테스트 실패 확인(로컬, 커밋 전 원복)

### T4. Settings management docs — precedence + fallback (Phase 3+4)

- Owner role: implementer
- Files owned:
  - `.claude/rules/moai/core/settings-management.md`
  - `internal/template/templates/.claude/rules/moai/core/settings-management.md`
- Scope:
  - [ ] precedence 섹션 신설: project `.claude/settings.json` `outputStyle` > user `~/.claude/settings.json` `outputStyle` > hardcoded default `"MoAI"` — 표 형식
  - [ ] 예시 2개:
    1. project=Einstein, user=MoAI → Einstein 로드
    2. user=Einstein, project=MoAI → MoAI 로드 (REQ-015 복합 조건)
  - [ ] unknown style fallback 규칙: `outputStyle: "NonExistent"` → "MoAI" fallback + 경고 로그 emit (REQ-008)
  - [ ] frontmatter schema 계약 요약 (3 key 최소 집합, 추가 key tolerate — OPEN-B 결정 반영)
  - [ ] breaking change 정책: frontmatter schema 변경은 별도 SPEC 필요
  - [ ] `make build` 로 embedded 템플릿 재생성
  - [ ] live와 template diff 없음 확인 (동일 패턴의 drift guard가 `rules_audit_test.go` 에 아직 없으므로 수동 확인)
- Dependencies: T0
- Parallel: T1/T2/T3와 완전 독립 (파일 겹침 없음) — 병렬 착수 가능
- DoD:
  - REQ-WF006-004, 008, 009, 015 문서화 완료
  - 이모지/HTML 태그/자유형 질문 없음 (coding-standards)
  - `documentation: ko` 설정 준수 — 본 문서는 한국어로 기술
  - `make build` 후 embedded 파일 diff 없음

### T5. Verification sweep (Phase 6)

- Owner role: reviewer
- Files: read-only + 테스트 실행만
- Scope:
  - [ ] `go test ./... -race -count=1`
  - [ ] `go vet ./...`
  - [ ] `golangci-lint run ./internal/template/...`
  - [ ] `make build && make install && moai version` — commit hash 일치
  - [ ] `diff -rq .claude/output-styles/moai internal/template/templates/.claude/output-styles/moai` 공백 출력
  - [ ] `diff -q .claude/rules/moai/core/settings-management.md internal/template/templates/.claude/rules/moai/core/settings-management.md` 공백 출력 (T4 동기화 검증)
  - [ ] `audit_sweep_patterns.md` Pattern A — YAML frontmatter CSV/array 규칙 위반 없는지 grep
  - [ ] 3번째 임시 style 추가 smoke: `echo -e "---\nname: Foo\n---" > internal/template/templates/.claude/output-styles/moai/foo.md && make build && go test ./internal/template/... -run TestOutputStylesExactlyTwo` 가 FAIL, `OUTPUT_STYLE_UNVERIFIED` 메시지 포함 확인 후 즉시 `git restore` 복원
  - [ ] TRUST 5 체크리스트 (acceptance.md §5 DoD 참조)
- Dependencies: T1, T2, T3, T4 전부
- DoD:
  - 모든 명령 zero error
  - 3rd style smoke 테스트 실행 로그 캡처 (커밋 본문 또는 PR 코멘트에 포함)
  - acceptance.md DoD 5 섹션 모두 체크

---

## 2. Dependency Graph

```
T0 ──┬──► T1 (schema test)
     │     └──► T2 (count/BC guard, same file)
     │           └──► T3 (drift test, same file)
     │                 └──► T5 (verification)
     └──► T4 (docs, independent) ────────────┘
```

- T4 는 T1/T2/T3 와 완전 독립 — 병렬 착수 가능
- T5 는 T1/T2/T3/T4 모두 완료 후 단독 실행

---

## 3. Parallelizable Groups

| Group | Tasks | 조건 |
|-------|-------|------|
| G1 | T1, T4 | T0 완료 후 병렬 시작 가능 (파일 겹침 없음) |
| G2 | T2 | T1 완료 후 즉시(동일 파일, 순차) |
| G3 | T3 | T1/T2 완료 후 즉시(동일 파일, 순차) |
| G4 | T5 | T1+T2+T3+T4 전부 완료 후 단독 실행 |

실질 임계 경로: T0 → T1 → T2 → T3 → T5 (5 단계). T4는 병렬로 묻어감.

---

## 4. File Ownership Map (team 모드용)

| Path pattern | Owner |
|--------------|-------|
| `internal/template/output_styles_audit_test.go` | T1, T2, T3 (동일 파일, 순차 편집) |
| `.claude/rules/moai/core/settings-management.md` | T4 |
| `internal/template/templates/.claude/rules/moai/core/settings-management.md` | T4 |
| `.claude/output-styles/moai/*.md` | 읽기만 (OWNER: 없음, 본 SPEC 범위 외) |
| `internal/template/templates/.claude/output-styles/moai/*.md` | 읽기만 (OWNER: 없음) |

team 모드에서 구현 teammates는 `.claude/rules/moai/workflow/worktree-integration.md §Implementation Agents` 를 따라 `isolation: "worktree"` 사용 필수. T4와 T1-T3 를 다른 teammate에 배정할 경우 worktree 격리로 충돌 방지.

---

## 5. Global DoD (모든 Task 공통)

- [ ] Conventional Commits 포맷 (`test`/`docs`/`chore`), 한국어 본문 (`git_commit_messages: ko`)
- [ ] 이모지 없음, XML 태그 user-facing 없음 (coding-standards)
- [ ] Go code, godoc, comment: English (CLAUDE.local.md §3 "All code, comments, godoc in English")
- [ ] 테스트 격리: `t.Parallel()`, `t.TempDir()` (필요 시), `t.Setenv("HOME", ...)` 금지, OTEL env 금지
- [ ] 하드코딩 금지 — `"MoAI"`, `"Einstein"`, `"moai.md"`, `"einstein.md"` 는 테스트 파일 내부 const 블록 단일 원천
- [ ] TRUST 5 5개 항목 녹색 (acceptance.md §5.2 참조)
- [ ] `make build && make install` 수행 후 바이너리 stale 검증 완료 (MEMORY.md Hard Constraint)
- [ ] 템플릿 언어 중립성 (CLAUDE.local.md §15) — 문서에 특정 언어 편향 없음

---

## 6. Integration with Wave 1 Siblings

Wave 1 parallel SPECs (모두 leaf):

- **SPEC-V3R2-CON-001** (FROZEN/EVOLVABLE codification): 파일 경로 `.claude/rules/moai/core/*`, `.moai/config/*` — 본 SPEC T4의 `settings-management.md` 와 겹칠 가능성 있음. 병렬 실행 시 동일 파일 동시 편집 방지 필요 → CON-001 착수자와 T4 오너가 file-lock 조율(순차 merge).
- **SPEC-V3R2-EXT-001** (Typed Memory Taxonomy): `internal/hook/memo/**`, `moai-memory.md` — 겹침 없음.
- **SPEC-V3R2-WF-001** (Skill Consolidation 48→24): `.claude/skills/**` — 본 SPEC은 `.claude/output-styles/` 만 read — 겹침 없음.

[HARD] T4 머지 전 CON-001 의 `settings-management.md` 편집 여부 확인. CON-001 이 동일 파일을 수정할 경우 본 SPEC은 CON-001 머지 후 rebase하여 precedence 섹션을 추가하는 방식으로 진행.

---

End of Tasks.
