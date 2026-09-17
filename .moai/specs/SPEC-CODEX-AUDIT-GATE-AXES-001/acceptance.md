# SPEC-CODEX-AUDIT-GATE-AXES-001 — 인수 기준

카드 **t686**. 각 AC 는 명령과 관측 출력으로 판정한다. 테스트 이름은 run-phase 가 정하며, 여기서는 관측해야 할 행동만 고정한다.

## §D AC Matrix

### 축 (a) — 양방향 회귀

#### AC-CAG-001 — required + 바이너리 부재 → 차단
- **Given** 감사 대상 트리의 `workflow.yaml` 에 `workflow.audit.gates.codex: required` 가 쓰여 있고 codex 바이너리가 PATH 에 없다
- **When** `codex_audit` 를 그 트리의 `project_root` 로 호출한다
- **Then** 결과는 `isError: false`, 구조화 내용의 `verdict == "fail"`, `gate_unmet` 비어 있지 않음, `summary` 에 "required" 게이트 미충족과 "codex binary not found" 원인이 함께 들어 있다

#### AC-CAG-002 — required + RPC 실패 → 차단
- **Given** AC-CAG-001 과 같은 설정이고, codex 는 있으나 RPC seam 이 fail-open inconclusive 를 돌려준다
- **When** `codex_audit` 를 호출한다
- **Then** AC-CAG-001 과 같은 차단 결과이며 `summary` 에 RPC 쪽 원인이 보존된다

#### AC-CAG-003 — 비-required + inconclusive → 바이트 동일
- **Given** `gates.codex` 가 각각 `off`, `advisory`, 키 부재, 설정 파일 부재, YAML 손상인 다섯 트리와, 변경 전 코드에서 같은 입력으로 캡처한 직렬화 결과(골든)
- **When** 각 트리에서 codex 바이너리 부재 조건으로 `codex_audit` 를 호출한다
- **Then** 다섯 경우 모두 직렬화 JSON 이 골든과 바이트 동일하다(`verdict == "inconclusive"`, `gate_unmet` 키 없음)

#### AC-CAG-004 — 배포 기본값은 opt-in 이 아니다
- **Given** `workflow.yaml` 에 `audit` 블록이 없다(엔진 기본값은 codex required)
- **When** codex 바이너리 부재 조건으로 `codex_audit` 를 호출한다
- **Then** `verdict == "inconclusive"` 이고 `gate_unmet` 이 없다

#### AC-CAG-005 — 실제 판정은 건드리지 않는다
- **Given** `gates.codex: required` 이고 codex 세션 스크립트가 각각 pass 와 fail 을 돌려준다
- **When** `codex_audit` 를 호출한다
- **Then** verdict 는 각각 `pass`, `fail` 로 변경 전과 같고 `gate_unmet` 은 비어 있다

#### AC-CAG-006 — 다중 경로·Stop-hook 무변경
- **Given** 변경 전 초록인 수렴 엔진 테스트와 codex Stop-hook 게이트 테스트
- **When** `go test ./internal/cli/ -run 'RequiredGate|Convergence|CodexReviewGate|MultiReviewGate' -count=1` 을 변경 후 실행한다
- **Then** 테스트 파일 수정 없이 전부 통과한다(`git diff --stat` 에 해당 테스트 파일이 없음)

#### AC-CAG-007 — 서술 정합
- **Given** 변경 후 트리
- **When** `grep -n "inconclusive" internal/cli/mcp_server.go` 와 두 감사자 본문(로컬·템플릿)을 판독한다
- **Then** `codex_audit` 도구 설명과 네 본문이 모두 "명시적 required 의 무판정은 단일 도구에서도 `verdict: fail` + `gate_unmet`" 을 서술하고, `make agents-emit-check` 가 0 으로 끝난다

### 축 (b) — 호출 증거 (§B.1 확정 옵션에 맞춰 run-phase M1 에서 명령을 구체화)

#### AC-CAG-008 — 증거 없는 PASS 거부
- **Given** `gates.codex: required` 인 트리와, codex 감사 호출 기록이 전혀 없는 상태에서 작성된 PASS 판정
- **When** 확정된 집행 지점이 그 판정을 처리한다
- **Then** PASS 로 수용되지 않고, 거부 이유에 "codex 감사 호출 기록 없음"이 명시된다

#### AC-CAG-009 — 에이전트 텍스트로는 충족 불가
- **Given** AC-CAG-008 의 조건에서, 판정 텍스트가 실제 저장소에 없는 영수증/호출을 인용한다
- **When** 집행 지점이 처리한다
- **Then** 거부된다. 반대로 실제 `codex_audit` 호출로 생긴 기록을 인용한 PASS 는 수용된다

#### AC-CAG-010 — 비-required 무영향
- **Given** `gates.codex` 가 `off` / `advisory` / 부재인 트리와 호출 기록 없는 PASS
- **When** 집행 지점이 처리한다
- **Then** 변경 전과 같은 결과로 수용되며 경고·차단 출력이 없다

#### AC-CAG-011 — 증거 저장소 판독 불가
- **Given** 증거 저장소 디렉터리 부재 또는 레코드 손상
- **When** required 트리와 비-required 트리에서 각각 PASS 를 처리한다
- **Then** 비-required 는 수용, required 는 "증거 판독 불가" gap 으로 거부된다(PASS 로 보고되지 않음)

### 축 (c) — 재현 우선

#### AC-CAG-012 — 관측 형태 characterization (RED 먼저)
- **Given** §B.2 에서 확보한 제보자 구성 형태(값은 가짜 자리표시자)
- **When** 분류기를 고치기 **전** 트리에서 해당 테스트를 실행한다
- **Then** `auth_provider == "unknown"` 이 관측되어 테스트가 실패하며, 그 실패 출력이 progress.md §E.2 에 인용된다. 형태를 확보하지 못하면 이 AC 와 AC-CAG-013 은 **미검증 gap** 으로 기록하고 분류기는 고치지 않는다

#### AC-CAG-013 — 수리 후 분류
- **Given** AC-CAG-012 의 테스트
- **When** 최소 수리 후 실행한다
- **Then** 관측 형태에 맞는 provider(`chatgpt` 또는 `api key`)가 나온다

#### AC-CAG-014 — unknown 유지
- **Given** 판독 불가 JSON, 서로 다른 provider 를 말하는 두 줄, 자격증명 없는 파일
- **When** 분류한다
- **Then** 모두 `unknown` 이며 기존 `codex_auth_ladder_test.go` 표 기반 사례가 전부 통과한다

#### AC-CAG-015 — 비밀값 비노출
- **Given** 자격증명 값에 식별 가능한 표식 문자열을 넣은 입력
- **When** `codex_setup` 결과와 분류 경로의 오류 문자열을 수집한다
- **Then** 표식 문자열이 어디에도 나타나지 않는다

## §D.1 Edge Cases

- 설정 파일이 `gates.codex: "required "`(공백 포함)나 `REQUIRED` 같은 변형을 담은 경우 — 현재 판별은 정확 일치다. 동작을 바꾸지 않고 정확 일치를 유지함을 테스트로 고정한다(비-required 로 취급).
- `project_root` 가 없고 서버 cwd 가 다른 트리인 경우 — 게이트는 감사 대상 트리의 설정을 따른다(기존 계약 유지).
- 축 (b) 에서 같은 세션에 `audit_multi` 가 codex 를 **제외**하고 돈 경우 — codex 호출 기록으로 치지 않는다.

## §D.2 Quality Gate

- `go vet` / `golangci-lint` 대상 패키지 0 건
- 신규·변경 코드 커버리지 85% 이상
- 템플릿 중립성 가드 통과, `make agents-emit-check` 0

## §D.3 Definition of Done

- AC-CAG-001 ~ 011 PASS, AC-CAG-012 ~ 015 는 PASS 또는 §B.2 에 따른 명시적 gap 기록
- 축 (b) 옵션 선택 근거가 progress.md 에 기록됨
- 미러 쌍과 방출본 동기화, `make build` 성공
- 이슈 착지 댓글·종료는 이 SPEC 의 DoD 가 아니다(리드 몫)
