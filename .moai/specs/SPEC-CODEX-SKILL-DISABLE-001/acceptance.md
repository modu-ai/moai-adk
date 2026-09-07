# SPEC-CODEX-SKILL-DISABLE-001 — 수용 기준

카드 t502 · Tier M 예산: 수용 기준 16개 이하. 아래는 정확히 16개다. 모든 AC는 이진 판정 가능해야 하며, 판정 근거는 실행한 명령과 그 출력이다.

## §A 발행 안전 (스키마 하드 에러 방지)

### AC-CSD-001 — 발행 엔트리는 `path` + `enabled = false` 를 함께 가진다
- **Given** 스킬 이름이 실존 `SKILL.md` 로 해석된 상태
- **When** 병합 함수가 새 엔트리를 발행하면
- **Then** 출력의 그 엔트리 구간에 `path` 와 `enabled` 가 모두 있고 `enabled` 값이 `false` 다.
- maps REQ-CSD-020

### AC-CSD-002 — 가드의 물성 (뮤턴트)
- **Given** AC-CSD-001을 검사하는 테스트
- **When** 구현에서 `enabled` 줄 발행을 제거한 뮤턴트를 넣으면
- **Then** 그 테스트가 **실패**한다. (뮤턴트가 잡히지 않으면 AC-CSD-001은 공허하다 — 채택 불가.)
- maps REQ-CSD-021

### AC-CSD-003 — 디렉터 모양 경로 거절
- **Given** 해석 결과가 `SKILL.md` 로 끝나지 않는 경로
- **When** 발행이 시도되면
- **Then** 동사가 거절하고 대상 내용은 바이트 불변이다.
- maps REQ-CSD-012

## §B 이름 해석

### AC-CSD-010 — 이름 → 파일 경로 해석
- **Given** 후보 루트에 `<name>/SKILL.md` 가 실존
- **When** 사용자가 `moai skills disable <name> --codex` 를 실행하면
- **Then** dry-run 보고가 그 절대 경로를 명시한다.
- maps REQ-CSD-010

### AC-CSD-011 — 해석 불가·모호 시 무쓰기
- **Given** (a) 어떤 후보 루트에도 `<name>/SKILL.md` 가 없는 상태, (b) 둘 이상의 루트에서 해석되는 상태
- **When** 각각 `--force` 를 포함해 실행하면
- **Then** 두 경우 모두 거절 사유가 출력되고, config의 sha256이 실행 전과 동일하다.
- maps REQ-CSD-011

### AC-CSD-012 — 미러 부재는 이름 붙은 거절 사유이지 에러가 아니며, 세 사유는 서로 구별된다
- **Given** `.agents/skills/` 자체가 존재하지 않는 프로젝트 (이 저장소의 실제 상태 — `ls .agents/skills` exit 1)
- **When** `moai skills disable <name> --codex --force` 를 실행하고, 이어서 AC-CSD-011의 (a)·(b) 상태에서도 실행하면
- **Then** 미러 부재 실행이 종료 코드 0으로 끝나고 config sha256이 불변이며, 세 실행의 사유 문자열이 **서로 다르다**. (한 문구로 뭉뚱그리면 사용자는 무엇을 고쳐야 하는지 알 수 없다.)
- maps REQ-CSD-011, REQ-CSD-042

## §C 병합 — 멱등·비파괴

### AC-CSD-020 — 기존 엔트리 갱신 (중복 추가 없음)
- **Given** 대상 경로가 `enabled = true` 로 이미 등록된 config
- **When** 비활성화 병합을 적용하면
- **Then** 그 경로를 가진 `[[skills.config]]` 엔트리 수가 여전히 1이고, 그 값이 `false` 다.
- maps REQ-CSD-030

### AC-CSD-021 — 재실행 바이트 불변 (멱등)
- **Given** 대상 경로가 이미 `enabled = false` 인 config
- **When** 같은 병합을 다시 적용하면
- **Then** 출력 바이트가 입력과 **완전히 동일**하다(`bytes.Equal`).
- maps REQ-CSD-031

### AC-CSD-022 — 기존 엔트리·주석·줄끝 형식 보존
- **Given** 대상과 무관한 엔트리 N개(주석·빈 줄 포함)를 가진 config, 그리고 CRLF 줄끝 config와 말미 개행이 없는 config
- **When** 병합이 일어나면
- **Then** N개 엔트리의 줄들이 순서·내용 그대로 남고, 원래의 줄끝·말미 개행 상태가 유지된다.
- maps REQ-CSD-032

### AC-CSD-023 — 미인식 줄이 있는 엔트리는 불가침
- **Given** 대상 경로 엔트리 구간에 파서가 인식하지 못한 줄이 있는 config (`FirstUnrecognizedLine >= 0`)
- **When** 병합이 시도되면
- **Then** 그 엔트리는 수정되지 않고, 건너뛴 사유가 보고된다.
- maps REQ-CSD-033

## §D 러너 — dry-run · 백업 · fail-open

### AC-CSD-030 — 기본은 dry-run이고, 계층 명시 없이는 실행되지 않는다
- **Given** 유효한 config와 해석 가능한 스킬 이름
- **When** (a) `--force` 없이 실행, (b) `--codex` 없이 실행하면
- **Then** (a)는 「무엇을 쓸 것인지」만 출력하고, (b)는 명령이 실패하며, 두 경우 모두 config sha256이 실행 전과 동일하다.
- maps REQ-CSD-040, REQ-CSD-001

### AC-CSD-031 — 쓰기 전 백업, 백업 실패 시 무쓰기
- **Given** `--force` 실행, 그리고 백업 쓰기가 실패하도록 만든 상태(테스트 seam)
- **When** 각각 실행하면
- **Then** 정상 경로에서는 `<cfg>.bak-<UTC>` 가 원본과 동일한 내용·mode 0600으로 존재하고 원본 sha256이 출력되며, 백업 실패 경로에서는 대상 config가 바이트 불변이다.
- maps REQ-CSD-040, REQ-CSD-041

### AC-CSD-032 — 입력 부재는 에러가 아니다
- **Given** codex home이 해석되지 않거나 config가 없는 상태
- **When** 실행하면
- **Then** 사유를 말하고 exit code 0으로 끝난다.
- maps REQ-CSD-041

## §E 경계 검증

### AC-CSD-040 — 파서 읽기 전용 유지 · prune 미변경
- **Given** 이 SPEC의 구현 diff
- **When** `git diff --name-only` 로 확인하면
- **Then** `internal/codexwiring/skills.go` 와 `internal/cli/codex_skills_prune.go` 가 모두 목록에 없고, 기존 prune 테스트가 전부 통과한다.
- maps REQ-CSD-050, REQ-CSD-051

## §F 선행 측정 게이트 [HARD]

### AC-CSD-050 — 게이트 경로 모양이 측정으로 확정된다 (Q-a + Q-b)
- **Given** 미러 심링크(`.agents/skills/<name>` → `../../.claude/skills/<name>`)로 노출되는 살아있는 스킬, 그리고 같은 스킬이 **사본 폴백**(`MirrorModeCopy` — 경로에 심링크가 없는 진짜 디렉터리)으로 배포된 트리
- **When** `CODEX_HOME` 격리 환경에서 5셀을 돌리면 — S+(심링크 표기 + `enabled=true`, 양성 통제) · S-lit(심링크 표기 + `false`) · S-res(해석 실경로 표기 + `false`) · C+(사본 경로 + `true`, 양성 통제) · C-x(S에서 묶인 것과 같은 표기 + `false`)
- **Then** 각 셀의 스킬 노출 마커 수가 기록되고, 판정서가 **(Q-a) 어느 표기가 게이트에 묶이는지**와 **(Q-b) 그 표기가 copy 모드에서도 묶는지**에 각각 예/아니오로 답한다.
- **[HARD] 차단 조건 1**: 이 판정 이전에는 발행 경로 모양을 코드에 고정하지 않는다.
- **[HARD] 차단 조건 2**: 양성 통제 S+ 또는 C+ 가 노출되지 않으면 그 행렬은 채택 불가(공허한 음성).
- **[HARD] 차단 조건 3**: 두 모드 모두에서 묶이는 표기가 없으면 그 사실을 판정으로 기록하고, 발행 대상을 좁히는 범위 재협상을 리드에게 blocker로 올린다 — 한쪽에서만 유효한 모양을 발행하지 않는다.
- maps REQ-CSD-012, REQ-CSD-013

### AC-CSD-051 — 측정 버전 스탬프 + 미측정 미러 결과의 Gaps 명시
- **Given** M1 판정서
- **When** 읽으면
- **Then** 측정 시점에 직접 실행한 `codex --version` 출력이 인용돼 있고(세션 반입 값 금지), `mirrorOneSkill` 의 네 결과 중 측정한 둘(symlink·copy)과 측정하지 않은 둘(skipped·failed)이 Gaps 절에 구별돼 적혀 있으며, 후자가 AC-CSD-012의 거절 경로로 처리된다는 사실이 함께 명시된다.
- maps REQ-CSD-013, REQ-CSD-011

## §G Definition of Done

- [ ] AC-CSD-050 판정서(`.moai/reports/t502/gate-path-shape.md`)가 존재하고 양쪽 양성 통제(S+·C+)가 유효
- [ ] §A~§E 전 AC가 테스트로 표현되고 통과
- [ ] AC-CSD-002 뮤턴트가 실제로 잡힘(잡힌 사실을 출력으로 인용)
- [ ] `go test ./internal/cli/... ./internal/codexwiring/...` 통과 (전체 스위트는 CI 몫)
- [ ] `go vet ./...` · `golangci-lint run` 무경고
- [ ] 미해결 `[NEEDS CLARIFICATION]` 0건
