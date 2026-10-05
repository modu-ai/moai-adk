# SPEC-AUDIT-BUILD-IDENTITY-001 — 구현 계획

> 순서는 **되돌리기 어려운 결정 먼저**다. 자료형과 필드 이름이 맨 위에 오고, 배선과 회귀 가드가 뒤에 온다.

## §A 맥락

카드 t248. 감사 판정이 자기를 낸 바이너리의 커밋을 밝히지 않아, 259 커밋 뒤처진 설치본이 낸 판정을 사후에 가려낼 수 없었다. 근거 좌표는 `spec.md` §1.1~§1.3(이 워크트리 base `64bba61aa`).

## §B 결정해야 할 것 (변경 가능성 높은 순)

### D1 — 자료형과 필드 이름 [결정됨 · 운영자 2026-09-01]

JSON 키는 한 번 나가면 소비자가 붙으므로 먼저 정했다. **평탄 형제 필드 2개**를 `ReviewOutput`과 `ConvergenceResult`에 각각 더한다. 중첩 신원 객체는 만들지 않는다.

    // internal/cli/mcp_codex.go — ReviewOutput
    BuildCommit string `json:"build_commit,omitempty"` // 서빙 바이너리의 커밋 SHA (version.GetCommit())
    BuildLag    string `json:"build_lag,omitempty"`    // 지연 권고. StatusBehind 에서만 비지 않음

    // internal/cli/mcp_convergence.go — ConvergenceResult
    (같은 두 필드, 같은 키 이름, 같은 타입)

근거 세 가지:

1. **선례가 평탄이다.** 같은 구조체가 이미 `synthesis_note`·`gate_unmet`을 형제 필드로 달고 있다(`internal/cli/mcp_codex.go:262` 주변). 고치고 있는 파일의 모양을 따르는 것이 취향보다 우선한다.
2. **Tier S 최소 변경.** 필드 2개는 자료형 선언 2줄씩이고, 중첩 객체는 새 타입 + 마샬 경로 + 테스트 축을 하나 더 만든다.
3. **확장 여지를 살 두 번째 소비자가 없다.** 중첩이 사는 것은 미래의 확장 자리인데, 이 카드에는 그것을 쓸 소비자가 없다 — YAGNI.

**버전 필드(`build_version`)는 더하지 않는다.** 버전이 곧 대표 뮤턴트이므로(`spec.md` §1.4), 커밋 옆에 나란히 실으면 소비자가 틀린 쪽을 읽도록 초대하는 셈이다. 버전이 필요해지는 날이 오면 그때 카드가 더한다.

붙는 자리와 붙지 않는 자리:

| 자료형 | 좌표 | 처분 |
|---|---|---|
| `ReviewOutput` | `internal/cli/mcp_codex.go:262` | 두 필드 추가 (`codex_audit` / `glm_audit` 공유) |
| `ConvergenceResult` | `internal/cli/mcp_convergence.go:106` | 두 필드 추가 (`audit_multi` 반환 + 지속 기록이 같은 구조체를 마샬) |
| `PerBackendVerdict` | `internal/cli/mcp_convergence.go:89` | **무변경** — 한 바이너리가 세 백엔드를 서비스하므로 per-backend 반복은 같은 값의 3중 기록이 된다 |

`omitempty` 규율은 협상 대상이 아니다(REQ-ABI-004): 두 값이 모두 비면 두 키 다 사라져 기존 소비자의 파스가 바이트 동일하게 유지된다.

### D2 — 지연 권고의 거처 [결정됨 · 운영자 2026-09-01]

지연 권고는 신원 객체 안에 중첩하지 않고 **형제 필드 `build_lag`**으로 둔다. D1과 같은 선례(`synthesis_note`·`gate_unmet`)를 따르며, 두 구조체에서 모양이 일관되게 유지된다.

### D3 — 신원 조립 지점 (단일 생성자)

세 진입점이 각자 조립하면 드리프트한다(REQ-ABI-007). `internal/cli`에 생성자 하나를 둔다:

    func auditBuildIdentity(ctx context.Context, projectRoot string) (buildCommit, buildLag string)

- 커밋은 `pkg/version`의 `GetCommit()`에서 읽는다. 버전은 `binlag.Request.BinaryVersion`에만 넘기고(비교에 참여하지 않는 보고용 입력) **결과 JSON에는 싣지 않는다**.
- 지연은 `binlag.Evaluate(ctx, binlag.Request{Dir: projectRoot, BinaryCommit: version.GetCommit(), BinaryVersion: version.GetVersion()})` → `binlag.Advisory(v)`.
- **`projectRoot`가 비면 비교를 건너뛰지 않는다 — `os.Getwd()`로 폴백한다**(운영자 결정, D-3). 선례를 그대로 따르며 두 번째 해석 규칙을 만들지 않는다:

      internal/cli/doctor.go:518   func checkBinaryFreshness(...)
      internal/cli/doctor.go:521       cwd, err := os.Getwd()   // err 시 cwd = ""
                                       → binlag.Request{Dir: cwd, ...}

  `os.Getwd()`가 실패하면 빈 문자열을 그대로 넘긴다 — `binlag.gitCompare`가 빈 `Dir`와 비-git 디렉터리를 모두 `StatusNotApplicable`로 처리하므로 폴백은 fail-open을 깨지 않는다(REQ-ABI-005).
- [HARD] **폴백 값은 비교에만 쓰고 백엔드에 넘기지 않는다.** `resolveOptionalToolProjectRoot`가 빈 문자열을 돌려주는 것은 의도된 계약이다 — 인자를 생략한 기존 호출자의 백엔드가 받는 `cwd`를 바꾸지 않기 위해서다(`internal/cli/mcp_project_root.go:95-107`의 주석). `performCodexAudit` / `performGLMAudit`에 넘기는 인자는 무변경이며, `git diff`로 확인한다(`acceptance.md` §D.1).
- 새 패키지를 만들지 않는다(Tier S).

`internal/cli/mcp_server_runtime.go:63`의 기존 스탬프 조립과 **중복 구현이 되지 않도록** 한다 — 그쪽은 프로세스 기록, 이쪽은 판정 기록으로 목적이 다르지만, 읽는 원천(`pkg/version`)은 같아야 한다.

### D4 — 배선 (세 진입점)

| 진입점 | 핸들러 | 이미 손에 있는 트리 |
|---|---|---|
| `codex_audit` | `handleCodexAudit` (`mcp_codex.go:1479`) | `resolveToolProjectRoot(req)` (`:1493`) |
| `glm_audit` | `handleGLMAudit` (`mcp_glm.go:221`) | `resolveToolProjectRoot(req)` (`:245`) |
| `audit_multi` | `handleAuditMulti` (`mcp_audit_multi.go`) | `resolveOptionalToolProjectRoot(req)` (`:71`) — **생략 시 `""`**, 비교는 `os.Getwd()` 폴백 |

`audit_multi`는 `runMultiAudit`의 결과 조립 지점에서 채우는 편이 자연스럽다 — 지속(`persistConvergenceResult`, `mcp_convergence.go:646`)이 같은 구조체를 그대로 마샬하므로, 반환값에 채우면 파일 기록은 **자동으로 따라온다**(REQ-ABI-002가 별도 코드를 요구하지 않는 이유).

사람이 읽는 한 줄(`mcp_audit_multi.go:145`의 텍스트 폴백)은 마샬 실패 시에만 쓰이는 열화 경로다. 여기에 커밋을 덧붙일지는 선택 — 붙인다면 문자열 포맷만 바뀐다.

### D5 — 회귀 가드 (기계적)

`acceptance.md`의 8개 기준을 테스트로 옮긴다. 순서는 반뮤턴트(AC-ABI-003)를 **먼저** 쓰고, 그 테스트가 현재 트리에서 RED임을 확인한 뒤 구현한다.

## §C 마일스톤

| M | 내용 | 닫는 AC |
|---|---|---|
| M1 | 평탄 필드 2개 + 단일 생성자(폴백 포함) + 반뮤턴트 테스트(RED 확인 포함) | AC-ABI-003, AC-ABI-004 |
| M2 | 세 진입점 배선 + 지속 경로 확인 | AC-ABI-001, AC-ABI-002 |
| M3 | `binlag` 심 소비(세 진입점 전수 + 빈 root 경로) + fail-open + 회귀 가드 | AC-ABI-005, AC-ABI-006, AC-ABI-007, AC-ABI-008 |

## §D 제약

- 추가·후방호환 JSON만. 기존 키 삭제·개명·의미 변경 0.
- fail-open: 신원 부재나 비-git 트리가 감사를 실패시키지 않는다.
- 새 MCP 도구 0, 새 verdict 열거값 0, 새 Go 패키지 0.
- 세 진입점 일관: 키 이름 상수 2개(`build_commit` / `build_lag`), 생성자 1개.
- 중첩 신원 객체 0, 버전 필드 0, `PerBackendVerdict` 무변경.
- `internal/binlag`은 **읽기만** 한다. 그 패키지를 수정해야 한다면 범위 재협상이다.
- 빈 `projectRoot`에서 비교를 건너뛰지 않는다(D-3). 폴백은 `os.Getwd()` 하나뿐이며, 백엔드 인자는 무변경이다.

## §E 자기 검증 (run-phase에서 실행할 것)

```bash
go test ./internal/cli/... ./internal/binlag/... -count=1
go vet ./internal/cli/... ./internal/binlag/...
golangci-lint run
grep -rn "merge-base\|is-ancestor" internal/cli --include='*.go' | grep -v _test.go
# 기준선(64bba61aa, 3건 정확 집합): graph_stamp.go:68 / graph_stamp.go:131 / mcp_review_material.go:95
```

[HARD] 판정은 트리에서 빌드한 테스트 바이너리로만 낸다. 설치본(`~/go/bin/moai`)의 출력은 이 카드의 증거가 아니다.

## §F 위험과 완화

| 위험 | 완화 |
|---|---|
| 반뮤턴트 테스트가 공허해짐(버전까지 다른 픽스처만 씀) | AC-ABI-004가 **버전 동일·커밋 상이** 케이스를 명시적으로 요구 |
| per-backend 필드까지 붙여 같은 값을 3중 기록 | D1에서 무변경으로 확정(운영자 결정) |
| `binlag` 비교가 감사마다 git을 두 번 부름(비용) | 감사 1회당 1회. 뜨거운 경로가 아님(감사는 이미 네트워크/외부 프로세스를 탄다) |
| 기존 테스트 기대값을 고쳐 통과시키는 유혹 | AC-ABI-008이 무수정 통과를 요구 |

## §G 안티패턴

- 버전 문자열만 기록하기(대표 뮤턴트).
- 커밋 옆에 버전 필드를 함께 싣기(REQ-ABI-003 위반).
- 두 필드를 중첩 객체로 묶기(D1 위반).
- 진입점마다 신원을 따로 조립하기.
- `binlag` 밖에 두 번째 조상 비교를 쓰기.
- 신원이 없을 때 감사를 에러로 만들기.
- 빈 `projectRoot`에서 비교를 건너뛰기(가장 흔한 `audit_multi` 호출에서 지연 권고가 죽는다).
- 폴백 작업 디렉터리를 백엔드 인자로 흘리기(`resolveOptionalToolProjectRoot` 계약 위반).
- `build_commit`에 `"none"` / `"unknown"`을 그대로 싣기(신원처럼 보이지만 아무것도 식별하지 못한다).

## §H 상호 참조

- `internal/binlag/binlag.go` — `Evaluate` / `Advisory` / `Short` / `RemedyCommand`
- `internal/cli/doctor.go:518` — 기존 소비자 선례
- `internal/cli/mcp_server_runtime.go:63` — 프로세스 신원 스탬프(다른 목적, 같은 원천)
- SPEC-BINARY-LAG-VISIBILITY-001 (t326), SPEC-BINLAG-INVOCATION-001 (t366) — 지연 축의 선행 카드
