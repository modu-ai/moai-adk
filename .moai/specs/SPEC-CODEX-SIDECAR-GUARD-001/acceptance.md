# SPEC-CODEX-SIDECAR-GUARD-001 — 수용 기준

> 카드 t405 · 기준 트리 `64bba61aa` · 대상 파일 `internal/cli/init_agent_flag_test.go`

모든 좌표(`:70` / `:82` / `:97` / `:109`)는 **기준 트리 `64bba61aa` 시점의 줄 번호**다. 단언이
추가되면 뒤쪽 줄 번호가 밀리므로, run-phase에서는 줄 번호가 아니라 **시험 함수명**으로 대상을
지목한다.

## §A. 단언 결손 봉합

**AC-CSG-001** (MUST, REQ-CSG-001 · REQ-CSG-003) — *Given* 커밋 `64bba61aa` 기준의
`internal/cli/init_agent_flag_test.go`, *When* `TestRunInit_AgentClaudeLeavesNoCodexFiles`의 경로
슬라이스를 읽으면, *Then* 그 원소가 정확히 `.codex/hooks.json` · `.codex/config.toml` ·
`.moai/state/codex-wiring.json` 3개이고, `TestRunInit_AgentAbsentLeavesNoCodexFiles`(`:97`)의 경로
집합과 원소 단위로 동일하다. 단언 대상은 전부 **파일 경로**이며 `.codex/` 디렉터리 자체를
가리키는 원소는 0개다.

**AC-CSG-002** (MUST, REQ-CSG-002 · REQ-CSG-003) — *Given* 동일 파일, *When*
`TestRunInit_AgentBothWiresBothSides`의 경로 슬라이스를 읽으면, *Then* 그 원소가 정확히
`.codex/hooks.json` · `.codex/config.toml` · `.moai/state/codex-wiring.json` 3개이고,
`TestRunInit_AgentCodexWiresAndSkipsMCPProvisioning`(`:70`)의 경로 집합과 원소 단위로 동일하다.

**AC-CSG-003** (MUST, REQ-CSG-003) — *Given* 수정된 시험 파일, *When*
`grep -cE '"\.codex"|"\.codex/"' internal/cli/init_agent_flag_test.go`를 실행하면, *Then*
출력이 정확히 `0`이다.

> 두 대안 모두 `.codex` 또는 `.codex/` **바로 뒤에 따옴표**를 요구한다. 산문 주석은 이 패턴을
> 만들 수 없으므로 주석/단언 판별이 개입할 여지가 없다 — 판정은 숫자 하나로 끝난다.
>
> **이 AC의 성격을 분명히 해둔다**: 기준 트리 `64bba61aa`에서 이 값은 **이미 0**이다(실측).
> 따라서 이 AC는 기존 결함을 탐지하지 않으며, 결함이 고쳐졌다는 증거도 아니다. 구현이 디렉터리
> 단위 단언을 **새로 도입하지 않도록** 묶어두는 금지 제약이다. 이 AC에는 격리 뮤턴트를 두지
> 않는다 — 적발할 대상이 애초에 없는 검사에 적색 능력 입증 장치를 붙이는 것은 Tier S에 과하다.

## §B. 적색 능력 입증 — 격리 뮤턴트

**AC-CSG-004** (MUST, REQ-CSG-004) — *부재 방향 격리 뮤턴트.*
*Given* AC-CSG-001이 통과한 트리, *When* 임시 프로덕션 편집 2건으로 뮤턴트를 심고 관측이 끝나는
즉시 둘 다 되돌린다 — (1) `internal/codexwiring`에 비exported `writeSidecar`(`wire.go:177-192`)
를 호출만 하는 임시 export 래퍼(제안명 `WriteSidecarOnly`)를 추가하고, (2) `internal/cli/init.go`
의 `claude` 조기 반환 분기(`:167-169`)가 그 래퍼를 호출해 `--agent claude`가 **sidecar만** 기록하
게 만든다(`.codex/hooks.json`·`.codex/config.toml`은 여전히 기록하지 않는다). 그 상태에서
`go test ./internal/cli/ -run 'TestRunInit_Agent'`를 실행하면, *Then*
`TestRunInit_AgentClaudeLeavesNoCodexFiles`가 RED이고, **그 RED를 만든 실패 메시지가 오직
`.moai/state/codex-wiring.json` 한 경로에 대한 것**이다 — 같은 시험의 기존 두 단언
(`hooks.json` · `config.toml`)은 실패 메시지를 내지 않는다.

> 래퍼가 필요한 이유: `writeSidecar`는 다른 패키지(`internal/codexwiring`)의 비exported 함수라
> `internal/cli`의 `claude` 분기가 직접 호출할 수 없다 — 직접 형태("claude 분기가
> `writeSidecar`를 직접 호출")의 뮤턴트는 컴파일되지 않는다. 래퍼는 관측 중에만 존재하는 관측
> 도구이며, 원복 검증은 AC-CSG-007의 `git status --porcelain internal/cli/init.go internal/codexwiring/` 빈 출력 검사가 이미 두 경로를 함께 덮어 수행한다. 영구 export는 C1 위반이므로 하지 않는다.

> 동반 관측(기록 의무, 판정 기준 아님): 플래그 부재도 `resolveAgentWiring`에서 `claude`로
> 해소되므로 `TestRunInit_AgentAbsentLeavesNoCodexFiles`(`:97`)도 같은 이유로 RED가 될 것으로
> **예상**한다. 예상과 다르게 `:97`이 초록으로 남으면 그것은 통과가 아니라 **조사 대상**이며,
> 뮤턴트가 의도한 지점을 건드리지 못했다는 신호로 보고한다.

**AC-CSG-005** (MUST, REQ-CSG-004) — *존재 방향 격리 뮤턴트.*
*Given* AC-CSG-002가 통과한 트리, *When* `Wire`에서 **`writeSidecar` 호출만** 제거하고
(`hooks.json`·`config.toml` 기록 경로는 그대로 둔 채) `go test ./internal/cli/ -run
'TestRunInit_Agent'`를 실행하면, *Then* `TestRunInit_AgentBothWiresBothSides`가 RED이고, 그 RED를
만든 실패 메시지가 오직 `.moai/state/codex-wiring.json` 한 경로에 대한 것이다 — 같은 시험의
기존 두 단언은 실패 메시지를 내지 않는다.

**AC-CSG-006** (MUST, REQ-CSG-004) — *존재 방향 뮤턴트의 생사 교차 확인.*
*Given* AC-CSG-005와 동일한 뮤턴트가 심어진 상태, *When* 같은 실행의 결과에서
`TestRunInit_AgentCodexWiresAndSkipsMCPProvisioning`(`:70`)의 판정을 읽으면, *Then* 그 시험도
RED이다.

> 이 교차 확인은 부가 정보가 아니라 **판정에 필수**다. `:70`은 기준 트리에서 이미 sidecar 존재를
> 단언하고 있으므로, 살아 있는 뮤턴트라면 반드시 이 시험도 깨뜨린다. `:70`이 초록으로 남으면
> 뮤턴트가 아무것도 건드리지 못한 **죽은 뮤턴트**이며, 그 경우 AC-CSG-005의 RED는 아무것도
> 입증하지 않는다 — 이 상황은 AC-CSG-005의 통과로 계산하지 않고 뮤턴트를 다시 설계한다.

**AC-CSG-007** (MUST, REQ-CSG-004 · §E.2) — *뮤턴트 원복.*
*Given* AC-CSG-004 · AC-CSG-005 · AC-CSG-006의 관측이 끝난 상태, *When*
`git status --porcelain internal/cli/init.go internal/codexwiring/`를 실행하면, *Then* 출력이
비어 있다(프로덕션 코드에 남은 변경 0). 커밋에 포함되는 변경은
`internal/cli/init_agent_flag_test.go` 한 파일뿐이며, `git diff --stat`이 이를 보인다.

## §C. 검증 범위

**AC-CSG-008** (MUST, REQ-CSG-005) — *Given* 뮤턴트가 전부 원복된 최종 트리, *When* 두 실행을
관측하면 — (i) 패키지 녹색 명령 `go test ./internal/cli/... ./internal/codexwiring/...`, (ii)
시험별 판정 명령 `go test ./internal/cli/ -run 'TestRunInit_Agent' -v` — *Then* (i)은 exit code
0이고, (ii)의 관측 출력에는 네 시험 각각의 `--- PASS` 행이 하나씩 있다:
`TestRunInit_AgentCodexWiresAndSkipsMCPProvisioning` · `TestRunInit_AgentBothWiresBothSides` ·
`TestRunInit_AgentAbsentLeavesNoCodexFiles` · `TestRunInit_AgentClaudeLeavesNoCodexFiles`.

(ii)가 판정 명령이고 (i)은 패키지 녹색 명령이다. 비상세(non-verbose) 실행은 시험별 판정 행을
출력하지 않고 패키지당 `ok` 한 줄만 남긴다 — 네 시험 중 하나가 파일에서 삭제된 트리도 그 `ok`
한 줄은 같아서, 삭제 여부를 그 출력만으로는 판정할 수 없다. 반면 (ii)는 삭제된 시험의 `--- PASS`
행이 사라져 셋으로 줄고 판정이 뒤집힌다. `ok` 한 줄만으로는 이 구별이 불가능하다는 것이 곧
`spec.md` §A.4가 고발하는 공허 초록 모양이며, 이 AC가 (ii)로 그 재발을 막는다.

> 측정 사실(트리 `64bba61aa`): `-run 'TestRunInit_Agent'` 필터는 위 네 시험만 선택한다. 같은
> 파일의 나머지 시험 4종(`TestInitAgentFlag_RegisteredWithClosedSet` ·
> `TestValidateInitFlags_AgentClosedSet` · `TestRunInit_CodexProvisioningDeclineIsolated` ·
> `TestRunInit_CallsCodexWiring`)은 접두사가 일치하지 않아 선택되지 않는다. 즉 (ii)의 `--- PASS`
> 행 수는 네 시험의 존재 그 자체를 재는 카운터다.

(i)(ii) 이 쌍의 실행이 본 SPEC의 지역 검증 전부이며, `go test ./...`(전체 스위트) 지역 실행 기록은
0건이다 — 전 패키지 판정은 PR head에 대한 CI가 소유한다.

## §D. 추적 행렬

| AC | 요구 | 방향 | 확인 대상 |
|---|---|---|---|
| AC-CSG-001 | REQ-CSG-001 · REQ-CSG-003 | 부재 | `:109` 경로 집합 = `:97` 경로 집합 |
| AC-CSG-002 | REQ-CSG-002 · REQ-CSG-003 | 존재 | `:82` 경로 집합 = `:70` 경로 집합 |
| AC-CSG-003 | REQ-CSG-003 | 양방향 | 디렉터리 단언 부재 |
| AC-CSG-004 | REQ-CSG-004 | 부재 | 격리 뮤턴트가 새 단언만 RED로 |
| AC-CSG-005 | REQ-CSG-004 | 존재 | 격리 뮤턴트가 새 단언만 RED로 |
| AC-CSG-006 | REQ-CSG-004 | 존재 | `:70` 동반 RED = 뮤턴트 생존 증명 |
| AC-CSG-007 | REQ-CSG-004 | — | 프로덕션 코드 원복, 시험 파일 1개만 변경 |
| AC-CSG-008 | REQ-CSG-005 | — | 범위 한정 검증 초록 |

## §E. Definition of Done

- AC-CSG-001 … AC-CSG-008 전부 PASS
- 변경 파일: `internal/cli/init_agent_flag_test.go` 1개 (+ SPEC 산출물)
- 커밋 메시지에 카드 id `t405` 포함
- 뮤턴트 관측 결과(명령 + 관측된 출력)가 `progress.md` §E.2에 기록됨
