# SPEC-CODEX-SKILL-DISABLE-001 — 수용 기준

카드 t502 · 모든 AC는 이진 판정 가능해야 하며, 판정 근거는 실행한 명령과 그 출력이다.

> **[예산 초과 — 기록된 예외, 판정 완료] 현재 17개, Tier M 상한 16개.** 세는 명령은 표제 앵커 `grep -c '^### AC-CSD'` → **17**. (id 토큰을 훑는 `sort -u` 형태는 이 문서 자체가 그 패턴을 본문에 적으면 자기 자신을 세어 18을 낸다 — 표제 앵커가 오염되지 않는 쪽이다.) 기준을 넓혀 수를 맞추지 않았다: 그 수법이 감사 D-N1이 지적한 손상(수는 줄고 한 기준이 서명해야 할 표면은 늘어남) 그 자체다. **원인·검토하고 기각한 병합 후보·「완화가 아니라 예외」라는 성격은 `plan.md` §F.1 이 정본으로 담는다** — 이 문서 밖의 대화 기록이 아니라 그 절을 읽으면 된다.

> **판별 셀 규율**: 한 기준이 두 성질을 지면 그것을 가르는 셀도 둘이어야 한다. 아래에서 `뮤턴트` 로 시작하는 줄이 그 판별 셀이며, 성질 하나마다 하나씩 있다. 뮤턴트가 잡히지 않으면 그 성질은 무방비이고 기준은 채택 불가다.

## §A 발행 — 엔트리가 생기는가, 값이 맞는가, 실제로 꺼지는가

세 기준은 한 근본의 세 축이다. 부분만 검사하면 **조합이 검사되지 않는다** — 추가가 no-op인 구현은 훑는 집합이 0이라 조용히 초록이 된다.

### AC-CSD-001 — 없던 엔트리가 정확히 하나 생기고, 값이 `false` 다 (추가 축 + 값 축)
- **Given** 경로 P에 대한 엔트리를 **하나도** 갖지 않은 config
- **When** 병합 함수를 적용하면
- **Then** P를 가진 `[[skills.config]]` 엔트리가 **정확히 1개** 존재하고, 그 구간에 `path` 와 `enabled` 가 모두 있으며 `enabled` 값이 `false` 다.
- **뮤턴트 1 (추가 축)**: 추가 경로를 no-op으로 만든 구현에서 이 기준이 **RED** 여야 한다.
- **뮤턴트 2 (값 축)**: `enabled = true` 를 발행하는 구현에서 이 기준이 **RED** 여야 한다.
- **뮤턴트 3 (스키마 축)**: `enabled` 줄 발행을 제거한 구현에서 이 기준이 **RED** 여야 한다. (누락은 사용자 codex 전면 장애다.)
- maps REQ-CSD-020, REQ-CSD-021, REQ-CSD-030

### AC-CSD-002 — 발행 표기는 절대 리터럴 미러 경로다
- **Given** 심링크 미러 프로젝트와 복사 폴백 미러 프로젝트 각각에서 해석된 스킬
- **When** 발행이 일어나면
- **Then** 두 경우 모두 `path` 값이 `<projectRoot>/.agents/skills/<skill>/SKILL.md` (절대·파일 모양)이다.
- **뮤턴트**: `.claude/skills/…` 해소 표기를 발행하는 구현에서 이 기준이 **RED** 여야 한다. (그 표기는 복사 미러에서 조용히 무효 — `gate-path-shape.md` Cres.)
- maps REQ-CSD-012

### AC-CSD-003 — 실제로 꺼진다 (E2E, 결과 축)
- **Given** 격리 `CODEX_HOME` 과 미러 픽스처, 그리고 노출 상태의 프로브 스킬
- **When** `--force` 로 동사를 실행하면
- **Then** verb가 만든 config를 그대로 겨눈 프로브가 실행 전 `--expect exposed`, 실행 후 `--expect gated` 로 각각 **exit 0** 이다 (마커 1 → 0).
- **기구**: `.moai/reports/t502/probe.sh probe --codex-home <verb가 쓴 home> --project <픽스처> --skill <name> --expect exposed|gated` — `--entry-path`/`--enabled` 를 주지 않으면 그 config를 **다시 쓰지 않고 그대로 읽는다**(불일치 시 exit 3). 픽스처는 같은 스크립트의 `fixture` 서브커맨드로 만든다.
- **기구의 검증 상태**: 측정 에이전트가 `--selftest` 로 검증했고 증거는 `.moai/reports/t502/lab/instrument/` 에 보관돼 있다(셀별 판정 6건 — `symlink-lit.txt` = `verdict=gated marker=0`, `copy-res.txt` = `verdict=exposed marker=1` — 및 판정서의 `SELFTEST PASS`). **이 SPEC 저자가 독립적으로 재실행하지는 않았다.** 제3자 재실행은 더 강한 근거이므로 남겨 두되, 부채가 아니라 선택적 강화다. selftest의 채택 게이트는 차단 검출만이 아니라 **무해한 엔트리를 차단으로 오인하지 않음**(복사 모양에서 해소 표기가 여전히 노출)까지 요구하므로, 이 계측기는 「무엇이든 gated로 부르는 도구」가 아님이 셀로 확인된 상태다.
- maps REQ-CSD-013

## §B 이름 해석 — 세 실패는 서로 다른 종료 코드를 갖는다

### AC-CSD-010 — 이름 → 절대 경로 해석
- **Given** 미러에 `<name>/SKILL.md` 가 실존
- **When** `moai skills disable <name> --codex` 를 실행하면
- **Then** dry-run 보고가 `<projectRoot>/.agents/skills/<name>/SKILL.md` 를 명시하고 exit 0 이다.
- maps REQ-CSD-010

### AC-CSD-011 — 해석 불가: 무쓰기 + **0이 아닌** 종료 코드
- **Given** 미러는 있으나 그 이름의 `SKILL.md` 가 없는 상태
- **When** `--force` 를 포함해 실행하면
- **Then** 종료 코드가 **0이 아니고**, config sha256이 실행 전과 동일하다. (오타난 이름이 스크립트·CI에서 검출 가능해야 한다.)
- maps REQ-CSD-011

### AC-CSD-012 — 모호: 무쓰기 + **0이 아닌** 종료 코드
- **Given** 둘 이상의 후보 루트에서 같은 이름이 해석되는 상태
- **When** `--force` 로 실행하면
- **Then** 종료 코드가 **0이 아니고**, config sha256이 실행 전과 동일하다.
- maps REQ-CSD-011

### AC-CSD-013 — 미러 부재: 무쓰기 + 종료 코드 **0**
- **Given** `.agents/skills/` 자체가 존재하지 않는 프로젝트 (이 저장소의 실제 상태 — `ls .agents/skills` exit 1)
- **When** `--force` 로 실행하면
- **Then** 종료 코드가 **0** 이고 config sha256이 불변이며, 사유가 「미러 없음」으로 이름 붙어 출력된다.
- maps REQ-CSD-011

## §C 병합 — 멱등·비파괴

### AC-CSD-020 — 기존 엔트리 갱신 (중복 추가 없음)
- **Given** 대상 경로가 `enabled = true` 로 이미 등록된 config
- **When** 비활성화 병합을 적용하면
- **Then** 그 경로를 가진 엔트리 수가 여전히 1이고, 값이 `false` 다.
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

## §D 러너 — dry-run · 백업 · 모드 · fail-open

### AC-CSD-030 — 기본은 dry-run이고, 계층 명시 없이는 실행되지 않는다
- **Given** 유효한 config와 해석 가능한 스킬 이름
- **When** (a) `--force` 없이 실행, (b) `--codex` 없이 실행하면
- **Then** (a)는 「무엇을 쓸 것인지」만 출력하고, (b)는 명령이 실패하며, 두 경우 모두 config sha256이 실행 전과 동일하다.
- maps REQ-CSD-040, REQ-CSD-001

### AC-CSD-031 — 쓰기 전 백업, 백업 실패 시 무쓰기, **원본 모드 보존**
- **Given** 퍼미션이 `0644` 인 기존 config로 `--force` 실행, 그리고 백업 쓰기가 실패하도록 만든 상태(테스트 seam)
- **When** 각각 실행하면
- **Then** 정상 경로에서는 `<cfg>.bak-<UTC>` 가 원본과 동일한 내용·mode 0600으로 존재하고 원본 sha256이 출력되며, **대상 config의 모드가 `0644` 그대로**다. 백업 실패 경로에서는 대상 config가 바이트 불변이다.
- **뮤턴트**: 대상 config를 `0600` 으로 무조건 쓰는 구현에서 이 기준이 **RED** 여야 한다.
- maps REQ-CSD-040, REQ-CSD-032

### AC-CSD-032 — 입력 부재는 에러가 아니다
- **Given** codex home이 해석되지 않거나 config가 없는 상태
- **When** 실행하면
- **Then** 사유를 말하고 exit code 0으로 끝난다.
- maps REQ-CSD-041

### AC-CSD-033 — 모든 거절·건너뛰기 사유가 서로 구별된다
- **Given** 다섯 상태: 해석 불가 · 모호 · 미러 부재 · 미인식 줄 건너뛰기 · 이미 `false`(무변경)
- **When** 각각 실행하면
- **Then** 다섯 출력의 사유 문자열이 **서로 다르다**. (한 문구로 뭉뚱그리면 사용자는 무엇을 고쳐야 하는지 알 수 없다.)
- maps REQ-CSD-042

## §E 경계 검증

### AC-CSD-040 — 파서 읽기 전용 유지 · prune 미변경
- **Given** 이 SPEC의 구현 diff
- **When** `git diff --name-only bf779ecf2..HEAD` 로 확인하면 — `bf779ecf2` 는 이 카드의 plan-phase 분기점(첫 커밋 `9a0e9a364` 의 부모이자 `develop` 과의 merge-base)이며, **여기에 박아 고정한다. 자리표시자로 남기지 않는다**: 실행 시점에 값을 고르게 두면 가장 손에 잡히는 후보가 `develop` 이고, 그것이 바로 아래에서 틀렸다고 이름 붙인 값이다. 두 대체 형태가 각각 반대 방향으로 틀린다:
  - **인자 없는 형태 → 거짓 음성.** 스테이지되지 않은 변경만 보고하므로 커밋된 트리에서 항상 빈 목록을 내고, 파서를 실제로 고쳐 커밋한 뒤에도 무조건 통과한다.
  - **브랜치명을 base로 쓰면 → 거짓 양성.** 브랜치는 움직이므로 분기 이후 그쪽에 들어온 남의 커밋이 내 변경으로 보고된다. 가설이 아니라 **실측된 사례다**: 리드가 이 카드의 트리에서 base를 `develop` 으로 두고 `internal/` 을 훑어 파일 3개(`internal/cli/codex_task.go`, 그 failcause 테스트, `internal/template/agentemit/agents-codex.yaml`)를 얻었고 경계 위반으로 보였으나, 같은 범위의 커밋 목록이 비어 있어 이 브랜치의 어떤 커밋도 `internal/` 을 건드리지 않았음이 확인됐다 — 세 파일은 분기 이후 `develop` 이 앞서간 11개 커밋의 것이었다. 통합 브랜치가 앞서갈수록 이 오탐은 잦아진다.
- **Then** 목록에 `internal/codexwiring/skills.go` 와 `internal/cli/codex_skills_prune.go` 가 **없다**.
- **그리고** 기존 prune 테스트가 통과한다: `go test -run 'TestPruneCodexSkillEntries|TestJudgeCodexSkillEntry|TestRunCleanCodexSkills' ./internal/cli/...` 의 **훑은 테스트 수가 0이 아님**을 먼저 확인한다(셀렉터가 0개를 고르면 `ok` 를 찍는다 — 그 초록은 아무것도 주장하지 않는다). 기대 모집단 수는 run-phase 시작 시점에 `go test -list` 로 실측해 여기에 적어 고정한다.
- maps REQ-CSD-050, REQ-CSD-051

## §F 재측정 게이트 [HARD]

### AC-CSD-050 — 발행 코드를 넣는 시점의 codex 버전에서 2셀 재측정
- **Given** `gate-path-shape.md` 의 판정은 `codex-cli 0.153.4` 한 버전 관측이고, realpath 정규화는 구현 세부라 버전 간 안정성이 보증되지 않는다
- **When** 발행 코드를 착지시키기 전에, 그 시점 버전에서 최소 2셀을 다시 돌리면 — 심링크 미러 차단(Slit 대응) · 복사 미러 차단(Clit 대응), 각각 양성 통제 동반
- **Then** 두 셀이 모두 차단(marker 0)으로 재현되고, 그 실행의 `codex --version` 출력이 판정 기록에 인용된다(세션 반입 값 금지).
- **[HARD] 차단 조건**: 재현되지 않으면 발행 표기를 그대로 두지 않는다 — 리드에게 blocker로 올린다.
- maps REQ-CSD-012, REQ-CSD-013

## §G Definition of Done

- [ ] §A~§E 전 AC가 테스트로 표현되고 통과
- [ ] AC-CSD-001의 뮤턴트 3종, AC-CSD-002·AC-CSD-031의 뮤턴트가 실제로 잡힘(잡힌 사실을 출력으로 인용)
- [ ] AC-CSD-003 E2E가 `probe.sh` 로 마커 1 → 0 을 보임
- [ ] AC-CSD-040의 prune 테스트 모집단 수가 `go test -list` 로 실측돼 기준에 적혀 있음
- [ ] AC-CSD-050 재측정 2셀 통과 + 버전 스탬프
- [ ] `go test ./internal/cli/... ./internal/codexwiring/...` 통과 (전체 스위트는 CI 몫)
- [ ] `go vet ./...` · `golangci-lint run` 무경고
- [ ] plan.md 에 미해결 clarification 마커 0건 (`grep -c 'NEEDS CLARIFICATION' plan.md` → 0)
