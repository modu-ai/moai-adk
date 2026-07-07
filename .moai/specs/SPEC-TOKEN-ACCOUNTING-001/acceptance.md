# Acceptance Criteria — SPEC-TOKEN-ACCOUNTING-001

> 각 AC는 **기계적 검증 가능**(단위 테스트 / grep / CLI 출력)해야 한다. vacuous AC 금지.

## §D AC Matrix

| AC | REQ | 검증 방식 | Severity |
|----|-----|-----------|----------|
| AC-TA-001 | REQ-TA-001,002 | 단위 테스트 (fixture 합산) | MUST |
| AC-TA-002 | REQ-TA-013 | 단위 테스트 (usage 부재 레코드) | MUST |
| AC-TA-003 | REQ-TA-013 | 단위 테스트 (malformed 라인) | MUST |
| AC-TA-004 | REQ-TA-003 | 단위 테스트 (cache-hit ratio 경계) | MUST |
| AC-TA-005 | REQ-TA-004 | 단위 테스트 (read-only 불변 — 파일 mtime/개수) | MUST |
| AC-TA-006 | REQ-TA-005,007 | 단위 테스트 (session-set 합산 + high 신뢰도) | MUST |
| AC-TA-007 | REQ-TA-006,007 | 단위 테스트 (lineage 부재 → low 폴백) | MUST |
| AC-TA-008 | REQ-TA-008,010 | grep (progress.md `## §I` + `tokens_spent:`) | MUST |
| AC-TA-009 | REQ-TA-009 | 회귀 테스트 (§I 추가 전후 ClassifyEra 불변) | MUST |
| AC-TA-010 | REQ-TA-011,012 | CLI 통합 테스트 (`moai spec audit --json` grep `tokens_spent`) | MUST |
| AC-TA-011 | REQ-TA-009 | 소스 가드 (era.go/audit.go의 §E.N + SHA 토큰 미개명) | SHOULD |
| AC-TA-012 | REQ-TA-014 | 중립성 가드 (token-accounting 소스 0개가 template tree에 존재) | MUST |

## §D.1 Given-When-Then Scenarios

### Scenario 1 — transcript 합산 (핵심 happy path) [AC-TA-001]

- **Given** assistant 턴 3개를 가진 transcript JSONL fixture, 각 `message.usage` =
  `{input_tokens:100, output_tokens:20, cache_creation_input_tokens:0, cache_read_input_tokens:500}`
- **When** 파서가 세션 합산을 수행
- **Then** `tokens_input==300`, `tokens_output==60`, `tokens_cache_read==1500`,
  `tokens_cache_creation==0`, `tokens_spent==1860` (4필드 산술 합) 이어야 한다.
- **검증**: `go test ./internal/tokenusage/ -run TestSumSession` 통과.

### Scenario 2 — malformed/부재 관용 (robustness) [AC-TA-002, AC-TA-003]

- **Given** (a) `message.usage`가 없는 assistant 레코드 1개, (b) JSON 파싱 불가 라인 1개,
  (c) 정상 레코드 1개를 섞은 transcript
- **When** 파서가 합산
- **Then** (a)는 0 기여, (b)는 skip, (c)만 반영되며 **panic 없이** 정상 반환.
- **검증**: `go test ./internal/tokenusage/ -run TestMalformedTolerant` 통과 (패닉 시 실패).

### Scenario 3 — 귀속 신뢰도 (attribution honesty) [AC-TA-006, AC-TA-007]

- **Given** progress.md lineage에 SPEC-전용 session-UUID 2개가 기록된 SPEC
- **When** session-set 합산 실행
- **Then** `token_attribution: session-set`, `token_session_count: 2`,
  `token_attribution_confidence: high`.
- **And Given** lineage가 `not-available`(environment-fallback)인 SPEC
- **When** 합산 실행
- **Then** 활성 세션 1개로 폴백 + `token_attribution_confidence: low`.
- **검증**: `go test ./internal/tokenusage/ -run TestAttributionConfidence` 통과.

### Scenario 4 — 파서 무충돌 회귀 (parser safety) [AC-TA-009, AC-TA-011]

- **Given** `§E.2/§E.3/§E.4` skeleton + `sync_commit_sha` 를 가진 progress.md
- **When** `## §I Token Accounting` 섹션 + 토큰 필드를 추가
- **Then** 추가 전후로 `spec.ClassifyEra()` 반환 era가 **동일**해야 한다 (§I는 era.go가 grep하지
  않는 신규 letter이므로 분류 불변).
- **검증**: `go test ./internal/spec/ -run TestEraUnchangedByTokenSection` 통과.

### Scenario 5 — audit 표면 노출 [AC-TA-010, AC-TA-012]

- **Given** `## §I Token Accounting` + `tokens_spent: 1860` 이 채워진 SPEC 디렉터리
- **When** `moai spec audit --json` 실행
- **Then** 해당 SPEC의 JSON 항목에 `tokens_spent` 필드가 `1860` (또는 파싱된 정수)로 노출.
- **And** §I 미기록 SPEC은 `tokens_spent` 가 `null` 또는 omit (fabricate 금지).
- **And** `grep -rl "message.usage\|tokens_spent" internal/template/templates/` 결과가 **0** (중립성).
- **검증**: `go test ./internal/cli/ -run TestSpecAuditTokensSpent` 통과 + 중립성 grep 0.

## §D.2 Edge Cases

- 빈 transcript(assistant 턴 0개) → `tokens_spent==0`, ratio==0, panic 없음.
- 분모 0(input+creation+read==0) → `cache_hit_ratio==0` (0-division 방어).
- transcript 파일 부재(세션 UUID는 있으나 JSONL 없음) → 해당 세션 0 기여 + 로그, 중단 없음.
- 동일 UUID가 2개 SPEC lineage에 등장(세션 공유) → 해당 기여 세션 `low` 신뢰도로 강등.
- `~/.claude/projects/**` read-only 불변 — 실행 전후 디렉터리 파일 개수·mtime 불변 [AC-TA-005].

## §D.3 Quality Gate

- `go test ./internal/tokenusage/... ./internal/spec/... ./internal/cli/...` 전부 통과.
- 신규 패키지 coverage ≥ 85% (critical 경로 90% 목표).
- `go vet ./...` + `golangci-lint run` 무경고 (신규 코드 한정).
- 중립성 CI 가드 통과(token-accounting 소스가 template tree에 없음).
- 파서 무충돌 회귀 테스트 통과.

## §D.4 Definition of Done

- [ ] M1–M4 전 마일스톤 구현 + AC-TA-001~012 전부 PASS.
- [ ] `moai spec audit --json` 이 실제로 `tokens_spent` 를 노출(실측 출력 인용).
- [ ] progress.md §I 필드가 실제 SPEC에 기록되고 grep으로 확인.
- [ ] era.go/audit.go 의 `§E.N`/SHA 토큰 미개명 (소스 diff로 확인).
- [ ] 귀속 값이 신뢰도 qualifier와 함께 기록되어 미검증 정밀도를 주장하지 않음.
- [ ] run-phase가 생성하는 **유일한 신규 코드**는 `internal/tokenusage` 패키지 + `internal/spec/audit.go`
      / `internal/cli/spec_audit.go` 확장뿐이다. SPEC 디렉터리 밖의 **유일한 비(非)코드 문서 편집**은
      Section Map SSOT 문서 1편집(`spec-frontmatter-schema.md` §I 행 + 그 template mirror)뿐이며,
      그 외 신규 파일은 없다. template tree 무접촉(REQ-TA-014).
