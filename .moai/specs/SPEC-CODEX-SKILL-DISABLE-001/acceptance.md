# SPEC-CODEX-SKILL-DISABLE-001 — 수용 기준

카드 t502 · 모든 AC는 이진 판정 가능해야 하며, 판정 근거는 실행한 명령과 그 출력이다.

## §A 발행 안전 (스키마 하드 에러 방지)

### AC-CSD-001 — 발행 엔트리는 `enabled` 를 항상 가진다
- **Given** 스킬 이름이 실존 `SKILL.md` 로 해석된 상태
- **When** 병합 함수가 새 엔트리를 발행하면
- **Then** 출력 내용의 그 엔트리 구간에 `path` 와 `enabled` 두 키가 모두 존재한다.

### AC-CSD-002 — 발행 값은 `false`
- **Given** 위와 같은 상태
- **When** 비활성화 발행이 일어나면
- **Then** 해당 엔트리의 `enabled` 값이 `false` 다.

### AC-CSD-003 — 가드의 물성 (뮤턴트)
- **Given** AC-CSD-001을 검사하는 테스트
- **When** 구현에서 `enabled` 줄 발행을 제거한 뮤턴트를 넣으면
- **Then** 그 테스트가 **실패**한다. (뮤턴트가 잡히지 않으면 AC-CSD-001은 공허하다 — 채택 불가.)

### AC-CSD-004 — 디렉터 모양 경로 거절
- **Given** 해석 결과가 `SKILL.md` 로 끝나지 않는 경로
- **When** 발행이 시도되면
- **Then** 동사가 거절하고 대상 내용은 바이트 불변이다.

## §B 이름 해석

### AC-CSD-010 — 이름 → 파일 경로 해석
- **Given** 후보 루트에 `<name>/SKILL.md` 가 실존
- **When** 사용자가 `moai skills disable <name> --codex` 를 실행하면
- **Then** dry-run 보고가 그 절대 경로를 명시한다.

### AC-CSD-011 — 해석 실패 시 무쓰기
- **Given** 어떤 후보 루트에도 `<name>/SKILL.md` 가 없는 상태
- **When** `--force` 를 포함해 실행하면
- **Then** 동사가 거절 사유를 출력하고, config의 sha256이 실행 전과 동일하다.

### AC-CSD-012 — 다중 매치 시 무쓰기
- **Given** 둘 이상의 후보 루트에서 같은 이름이 해석되는 상태
- **When** `--force` 로 실행하면
- **Then** 모호성이 보고되고 config는 바이트 불변이다.

## §C 병합 — 멱등·비파괴

### AC-CSD-020 — 기존 엔트리 갱신 (중복 추가 없음)
- **Given** 대상 경로가 `enabled = true` 로 이미 등록된 config
- **When** 비활성화 병합을 적용하면
- **Then** 그 경로를 가진 `[[skills.config]]` 엔트리 수가 여전히 1이고, 그 값이 `false` 다.

### AC-CSD-021 — 재실행 바이트 불변 (멱등)
- **Given** 대상 경로가 이미 `enabled = false` 인 config
- **When** 같은 병합을 다시 적용하면
- **Then** 출력 바이트가 입력과 **완전히 동일**하다(`bytes.Equal`).

### AC-CSD-022 — 기존 엔트리 보존
- **Given** 대상과 무관한 엔트리 N개(주석·빈 줄 포함)를 가진 config
- **When** 병합이 일어나면
- **Then** 그 N개 엔트리의 줄들이 순서·내용 그대로 출력에 존재한다.

### AC-CSD-023 — 줄끝 형식 보존
- **Given** CRLF 줄끝 config, 그리고 말미 개행이 없는 config
- **When** 병합이 일어나면
- **Then** 두 경우 모두 원래의 줄끝·말미 개행 상태가 유지된다.

### AC-CSD-024 — 미인식 줄이 있는 엔트리는 불가침
- **Given** 대상 경로 엔트리 구간에 파서가 인식하지 못한 줄이 있는 config (`FirstUnrecognizedLine >= 0`)
- **When** 병합이 시도되면
- **Then** 그 엔트리는 수정되지 않고, 건너뛴 사유가 보고된다.

## §D 러너 — dry-run · 백업 · fail-open

### AC-CSD-030 — 기본은 dry-run
- **Given** 유효한 config와 해석 가능한 스킬 이름
- **When** `--force` 없이 실행하면
- **Then** 「무엇을 쓸 것인지」가 출력되고, config의 sha256이 실행 전과 동일하다.

### AC-CSD-031 — 쓰기 전 백업
- **Given** `--force` 실행
- **When** 쓰기가 일어나면
- **Then** `<cfg>.bak-<UTC>` 파일이 존재하고 그 내용이 원본과 동일하며 mode가 0600이고, 원본 sha256이 출력에 나타난다.

### AC-CSD-032 — 실패 시 무쓰기
- **Given** 백업 쓰기가 실패하도록 만든 상태(테스트 seam)
- **When** `--force` 실행하면
- **Then** 대상 config가 바이트 불변이다.

### AC-CSD-033 — 입력 부재는 에러가 아니다
- **Given** codex home이 해석되지 않거나 config가 없는 상태
- **When** 실행하면
- **Then** 사유를 말하고 exit code 0으로 끝난다.

### AC-CSD-034 — 건너뛴 항목은 사유와 함께 보고
- **Given** 건너뛰기 사유가 발생하는 어떤 경로(AC-CSD-011/012/024)
- **When** 실행하면
- **Then** 출력에 대상 식별자와 사유가 함께 나타난다.

### AC-CSD-035 — 계층 명시 없이는 실행되지 않는다
- **Given** `--codex` 없이 `moai skills disable <name>`
- **When** 실행하면
- **Then** 명령이 실패하고 config는 바이트 불변이다.

## §E 경계 검증

### AC-CSD-040 — 파서는 읽기 전용으로 남는다
- **Given** 이 SPEC의 구현 diff
- **When** `git diff --name-only` 로 확인하면
- **Then** `internal/codexwiring/skills.go` 가 목록에 없다.

### AC-CSD-041 — prune 미변경
- **Given** 같은 diff
- **When** 확인하면
- **Then** `internal/cli/codex_skills_prune.go` 가 목록에 없고, 기존 prune 테스트가 전부 통과한다.

## §F 선행 측정 게이트 [HARD]

### AC-CSD-050 — 심링크 대 실경로 게이트 모양이 측정으로 확정된다
- **Given** 미러 심링크(`.agents/skills/<name>` → `../../.claude/skills/<name>`)로 노출되는 살아있는 스킬
- **When** `CODEX_HOME` 격리 환경에서 (a) 심링크 경로 표기 + `enabled=false`, (b) 해석 실경로 표기 + `enabled=false`, (c) 양성 통제 `enabled=true` 세 셀을 돌리면
- **Then** 각 셀의 스킬 노출 마커 수가 기록되고, **어느 표기가 게이트에 묶이는지**가 판정서에 명시된다.
- **[HARD] 차단 조건**: 이 판정 이전에는 발행 경로 모양을 코드에 고정하지 않는다. 세 셀 모두 통제군이 유효(양성 통제가 노출됨)해야 행렬이 채택된다.

### AC-CSD-051 — 측정 버전 스탬프
- **Given** M1 판정서
- **When** 읽으면
- **Then** 측정 시점에 직접 실행한 `codex --version` 출력이 인용돼 있다(세션 반입 값 금지).

## §F.1 추적성 (AC → REQ)

| AC | 검증하는 REQ |
|---|---|
| AC-CSD-001 | maps REQ-CSD-020 |
| AC-CSD-002 | maps REQ-CSD-022 |
| AC-CSD-003 | maps REQ-CSD-021 |
| AC-CSD-004 | maps REQ-CSD-013 |
| AC-CSD-010 | maps REQ-CSD-010 |
| AC-CSD-011 | maps REQ-CSD-011 |
| AC-CSD-012 | maps REQ-CSD-012 |
| AC-CSD-020 | maps REQ-CSD-030 |
| AC-CSD-021 | maps REQ-CSD-031 |
| AC-CSD-022 | maps REQ-CSD-032 |
| AC-CSD-023 | maps REQ-CSD-032 |
| AC-CSD-024 | maps REQ-CSD-033 |
| AC-CSD-030 | maps REQ-CSD-040, REQ-CSD-001 |
| AC-CSD-031 | maps REQ-CSD-041 |
| AC-CSD-032 | maps REQ-CSD-042 |
| AC-CSD-033 | maps REQ-CSD-043 |
| AC-CSD-034 | maps REQ-CSD-044 |
| AC-CSD-035 | maps REQ-CSD-003, REQ-CSD-002 |
| AC-CSD-040 | maps REQ-CSD-050, REQ-CSD-052 |
| AC-CSD-041 | maps REQ-CSD-051 |
| AC-CSD-050 | maps REQ-CSD-013 (발행 경로 모양의 근거 측정) |
| AC-CSD-051 | maps REQ-CSD-013 |

## §G Definition of Done

- [ ] AC-CSD-050 판정서가 `.moai/reports/t502/` 에 존재하고 통제군이 유효
- [ ] §A~§E 전 AC가 테스트로 표현되고 통과
- [ ] AC-CSD-003 뮤턴트가 실제로 잡힘(잡힌 사실을 출력으로 인용)
- [ ] `go test ./internal/cli/... ./internal/codexwiring/...` 통과 (전체 스위트는 CI 몫)
- [ ] `go vet ./...` · `golangci-lint run` 무경고
- [ ] 미해결 `[NEEDS CLARIFICATION]` 0건
