# SPEC-AUDIT-BUILD-IDENTITY-001 — 수락 기준

> 모든 기준은 **이름 붙은 명령 또는 이름 붙은 테스트가 이름 붙은 필드를 단언**하는 형태로 쓴다. 판정 수단이 없는 기준은 이 문서에 두지 않는다.
>
> **채택된 필드 모양(운영자 결정, `plan.md` §B D1/D2)**: `ReviewOutput`과 `ConvergenceResult`에 **평탄 형제 필드 2개** — `build_commit`(커밋 SHA, 반뮤턴트 기준이 판정하는 유일한 필드)과 `build_lag`(지연 권고 문자열, 조상 빌드에서만 비지 않음). 중첩 객체 없음, **버전 필드 없음**, `PerBackendVerdict` 무변경.
>
> 좌표와 기준선은 이 워크트리(`WT-audit-binary-sha`)의 `git rev-parse --short HEAD` = `64bba61aa`에서 직접 측정한 값이다.

## §D 수락 매트릭스 (8건 — Tier S 상한 8 준수)

| AC | 요구 | 성격 | 판정 수단 |
|---|---|---|---|
| AC-ABI-001 | REQ-ABI-001, REQ-ABI-007 | 존재 + 모양(비어있지 않음 전제) | `go test ./internal/cli/ -run TestAuditVerdictCarriesBuildCommit` |
| AC-ABI-002 | REQ-ABI-002 | 지속 | `go test ./internal/cli/ -run TestPersistedConvergenceCarriesBuildCommit` |
| AC-ABI-003 | REQ-ABI-003 | **반뮤턴트**(버전 필드 부재 포함) | `go test ./internal/cli/ -run TestBuildIdentityVersionAloneIsRejected` |
| AC-ABI-004 | REQ-ABI-004 | 후방호환 | `go test ./internal/cli/ -run TestBuildIdentityOmittedWhenAbsent` |
| AC-ABI-005 | REQ-ABI-005 | fail-open | `go test ./internal/cli/ -run TestAuditCompletesWithoutBuildIdentity` |
| AC-ABI-006 | REQ-ABI-006 | 지연 경고(`build_lag`), 세 진입점 전수 | `go test ./internal/cli/ -run TestAuditLagAdvisoryNamesBothCommits` |
| AC-ABI-007 | REQ-ABI-006 | 단일 구현 | `go test ./internal/cli/ -run TestAuditLagUsesBinlagSeam` + 소스 스윕 |
| AC-ABI-008 | REQ-ABI-008 | 회귀 | 기존 `TestConverge*` / `TestRunMultiAudit*` 무수정 통과 |

---

## AC-ABI-001 — 판정 출력이 빌드 커밋을 담고, 세 진입점의 모양이 같다

**Given** `version.Commit = "abc123def"`(**비어 있지 않은** 픽스처)로 설정된 빌드에서
**When** `codex_audit`, `glm_audit`, `audit_multi` 각각의 핸들러가 결과를 낼 때
**Then** 세 결과 모두에 대해 다음이 성립한다:

1. **존재·값**: `build_commit` 키가 존재하고 값이 `"abc123def"`와 같다.
2. **[HARD] 비어있지 않음(선행 전제)**: `build_commit`이 빈 문자열이거나 JSON에서 생략되어 있으면 **테스트는 즉시 실패한다**. 이 단언은 아래 3의 비교보다 **먼저** 평가한다.
3. **모양 일치**: `build_commit`과 `build_lag`의 키 이름·타입이 세 결과에서 동일하고, `build_commit` 값이 서로 같다.

판정: `go test ./internal/cli/ -run TestAuditVerdictCarriesBuildCommit -count=1`

- 세 백엔드는 스텁으로 대체한다(네트워크·외부 바이너리 없음).
- 테스트는 키 이름을 **문자열 상수 두 개**(`build_commit` / `build_lag`)에서 읽어 세 결과 각각에 적용한다. 진입점마다 키를 따로 적는 테스트는 이 기준을 충족하지 않는다 — 그렇게 쓰면 드리프트를 못 본다.
- `audit_multi`는 `ConvergenceResult` 최상위에서 읽는다(한 바이너리가 세 백엔드를 서비스하므로 per-backend 반복은 요구하지 않는다).
- **2번 절이 이 기준의 공허화를 막는다.** `omitempty` 아래에서는 모든 결과가 빈 `build_commit`을 내면 키가 **부재**하게 되고, 그러면 3의 이름 비교는 빈 피연산자끼리, 값 비교는 `"" == "" == ""`가 되어 아무것도 주장하지 않은 채 참이 된다. 비어있지 않음을 먼저 못박아야 3이 무언가를 판정한다.

## AC-ABI-002 — 지속된 기록이 같은 커밋을 담는다

**Given** `SessionID`와 `ProjectRoot`가 주어진 `audit_multi` 실행에서
**When** 실행이 끝난 뒤 `<ProjectRoot>/.moai/state/audit-multi/<SessionID>.json`을 되읽을 때
**Then** 파일의 `build_commit`이 반환된 결과의 `build_commit`과 바이트 동일하고, 빈 문자열이 아니다.

판정: `go test ./internal/cli/ -run TestPersistedConvergenceCarriesBuildCommit -count=1`

- `t.TempDir()`를 `ProjectRoot`로 쓴다. 프로젝트 트리에 쓰는 테스트는 금지(CLAUDE.local.md §6).

## AC-ABI-003 — 반뮤턴트: 버전만으로는 통과하지 못한다

**Given** 다음 두 뮤턴트 각각에 대해

- **M1(카드가 지목한 것)**: `build_commit`을 비운 채 버전 문자열만 기록하는 구현(예: `build_version: "v3.1.3"`).
- **M2(더 미묘한 것)**: 필드 **이름은 `build_commit`인데 값이 버전 문자열**인 구현.

**When** 이 기준을 실행할 때
**Then** 두 뮤턴트 모두에서 테스트는 실패한다.

정상 구현에 대한 판정 절:

**Given** `Version = "v3.1.3"`, `Commit = "abc123def"`인 빌드와 `Version = "v3.1.3"`, `Commit = "999fedcba"`인 빌드 두 가지 — **버전은 동일하고 커밋만 다르다**
**When** 두 빌드가 각각 감사 결과를 낼 때
**Then**
1. 두 결과의 `build_commit`은 **서로 다르다**.
2. `build_commit`이 빈 문자열이면 테스트는 실패한다.
3. 결과 JSON에 **버전 문자열을 담은 키가 존재하면 테스트는 실패한다**(REQ-ABI-003 — 커밋 옆에 버전을 실으면 소비자가 틀린 쪽을 읽는다).

판정: `go test ./internal/cli/ -run TestBuildIdentityVersionAloneIsRejected -count=1`

- [HARD] 픽스처는 **버전이 같고 커밋만 다른** 쌍이어야 한다. 이 규율이 죽이는 것은 M1이 아니라 **M2**다: M1은 위 2번(비어있지 않음)만으로 픽스처와 무관하게 죽지만, M2는 버전까지 다른 픽스처에서는 두 값이 서로 달라 1번을 통과해 살아남는다. 버전이 동일한 쌍에서만 M2의 두 값이 같아져 죽는다.
- 근거: `pkg/version/version.go:8`의 기본값 `Version = "v3.1.3"`은 ldflags 없는 **모든** 빌드에서 동일하다. 그리고 `internal/binlag/binlag.go`의 `Request.BinaryVersion` 주석은 버전이 판정에 참여하지 않는 이유를 이미 명시한다.

## AC-ABI-004 — 신원이 없으면 JSON이 변하지 않는다

**Given** `build_commit`과 `build_lag`이 모두 빈 결과에서
**When** 결과를 마샬할 때
**Then** 산출 JSON의 키 집합이 이 변경 이전의 키 집합과 같다(두 키 모두 나타나지 않는다).

판정: `go test ./internal/cli/ -run TestBuildIdentityOmittedWhenAbsent -count=1`

- "변경 이전의 키 집합"은 테스트 안에 **명시적 기대 집합**으로 적는다(`verdict, summary, findings, next_steps` + 값이 있을 때만 나타나는 `synthesis_note` / `gate_unmet`). 변경 전 트리를 읽는 방식은 쓰지 않는다 — 테스트가 그 트리에 접근할 수 없다.
- `synthesis_note` / `gate_unmet`의 `omitempty` 선례를 그대로 따른다.

## AC-ABI-005 — fail-open: 신원 없이도 감사는 완주한다

**Given** 다음 세 상태 각각에서
1. `version.Commit`이 `"none"`(ldflags 없는 dev 빌드)
2. `version.Commit`이 `""`
3. `ProjectRoot`와 프로세스 작업 디렉터리가 모두 git 워킹 트리가 아닌 경우

**When** 세 진입점의 핸들러를 호출할 때
**Then**
1. 어떤 호출도 Go 에러를 반환하지 않는다.
2. verdict 값이 신원 도입 전과 같다.
3. `build_lag`이 비어 있다(따라서 JSON에서 생략된다).
4. **[D-8]** `version.Commit` ∈ {`""`, `"none"`, `"unknown"`}인 **모든** 상태에서 `build_commit`은 **빈 문자열이며 JSON에서 생략된다**. `"none"` / `"unknown"`을 그대로 싣는 구현은 이 기준에서 실패한다 — 그 문자열은 읽는 쪽에 신원처럼 보이면서 아무것도 식별하지 못해, 이 카드가 닫으려는 오귀속을 되살린다. 정규화 선례: `internal/binlag/binlag.go:108-110`이 `""` / `"none"` / `"unknown"`을 한 덩어리로 not-applicable 처리한다.

판정: `go test ./internal/cli/ -run TestAuditCompletesWithoutBuildIdentity -count=1`

- 세 상태를 **각각** 판정한다. 한 상태만 재고 다른 쪽을 추론하면 미측정이다.

## AC-ABI-006 — 지연 권고가 세 진입점 모두에서, 두 커밋을 모두 명시한다

**Given** `binlag.Comparer`를 `StatusBehind`(`BinaryCommit=<조상>`, `SourceHead=<HEAD>`)를 반환하는 스텁으로 대체한 상태에서
**When** `codex_audit`, `glm_audit`, `audit_multi` **세 핸들러 각각**을 호출할 때
**Then** 세 결과 모두에서 `build_lag`이 두 커밋의 짧은 형태를 **모두** 포함하고, 재설치 명령(`binlag.RemedyCommand`)을 포함한다.

**대조군**: 같은 스텁이 `StatusFresh`를 반환하면 세 결과 모두에서 `build_lag`은 빈 문자열이다.

**[D-3] 빈 `projectRoot` 경로**: `audit_multi`를 `project_root` 인자 **없이**(→ `resolveOptionalToolProjectRoot`가 `""` 반환) 호출하고 같은 `StatusBehind` 스텁을 심었을 때에도 `build_lag`이 비어 있지 않다. 즉 비교가 프로세스 작업 디렉터리로 폴백해 실제로 수행됐음을 관측한다.

판정: `go test ./internal/cli/ -run TestAuditLagAdvisoryNamesBothCommits -count=1`

- [HARD] 테스트는 **세 핸들러를 테이블 구동으로 전수** 돈다. 한 진입점만 도는 테스트는 이 기준을 충족하지 않는다 — 그렇게 쓰면 `codex_audit`에만 배선한 구현이 통과한다.
- 대조군이 없으면 "항상 문자열을 낸다"는 구현이 통과한다. 두 방향을 같은 테스트에서 관측한다.
- 빈 `projectRoot` 절이 없으면, `audit_multi`의 **가장 흔한 호출**(인자 생략)에서 지연 권고가 영영 발화하지 않는 구현이 통과한다.

## AC-ABI-007 — 비교 구현은 하나뿐이다

**Given** 감사 경로에서
**When** 아래 두 관측을 수행할 때
**Then** 둘 다 참이다.

**관측 1 — 심 통과**: `binlag.Comparer`를 호출 횟수 카운터가 달린 스텁으로 대체하면, **세 진입점 각각의** 감사 1회당 카운터가 1 이상 증가한다.

**관측 2 — 정확 집합 술어**: 아래 스윕의 히트 집합이 기준선과 **정확히 같다**(추가도 삭제도 없다).

```bash
grep -rn "merge-base\|is-ancestor" internal/cli --include='*.go' | grep -v _test.go
```

기준선(이 워크트리 `64bba61aa`에서 실측, 3건):

| # | 좌표 | 정체 |
|---|---|---|
| 1 | `internal/cli/graph_stamp.go:68` | 도움말 문자열 안의 `git merge-base` 예시 |
| 2 | `internal/cli/graph_stamp.go:131` | 플래그 설명 문자열 안의 같은 예시 |
| 3 | `internal/cli/mcp_review_material.go:95` | `runReviewGit(root, "merge-base", ref, "HEAD")` |

PASS 조건: 변경 후 히트 집합 == 위 3개 좌표 집합. 히트 수만 세지 않는다 — 하나가 사라지고 다른 하나가 생겨도 수는 같기 때문이다.

**[D-1] 3번 항목에 대한 명시적 판정**: `mcp_review_material.go:95`의 `resolveReviewMergeBase`는 `handleGLMAudit` → `collectReviewDiff` 경로에서 **리뷰 대상 diff의 기준점**을 구하는 호출이다. 바이너리 커밋과 트리 HEAD의 조상 비교가 **아니므로**, REQ-ABI-006이 금지하는 "두 번째 조상 비교"에 해당하지 않는다. 이 항목은 기준선에 남으며, 제거 대상이 아니다.

판정: `go test ./internal/cli/ -run TestAuditLagUsesBinlagSeam -count=1` + 위 스윕

- 두 관측이 함께 필요하다: 스텁만으로는 "다른 곳에도 사본이 있는지"를 말하지 못하고, 스윕만으로는 "감사가 실제로 심을 지나는지"를 말하지 못한다.

## AC-ABI-008 — 수렴 정책은 변하지 않았다

**Given** 이 SPEC의 변경이 적용된 트리에서
**When** 기존 수렴 테스트를 **수정 없이** 실행할 때
**Then** 전부 통과한다.

판정: `go test ./internal/cli/ -run 'TestConverge|TestRunMultiAudit|TestAuditMulti' -count=1`

- [HARD] 기존 테스트의 기대값을 고쳐서 통과시키는 것은 이 기준의 위반이다. 기대값 수정이 필요하다면 REQ-ABI-008이 깨진 것이다.

---

## §D.1 완료 정의 (Definition of Done)

- [ ] AC-ABI-001 ~ AC-ABI-008 전부 PASS, 각 판정 명령의 실제 출력을 §E.2에 인용
- [ ] `go test ./internal/cli/... ./internal/binlag/... -count=1` 통과(범위: 건드린 패키지)
- [ ] `go vet ./internal/cli/... ./internal/binlag/...` 통과
- [ ] `golangci-lint run` 신규 지적 0
- [ ] 새 MCP 도구 0, 새 verdict 열거값 0, 새 Go 패키지 0, 중첩 신원 객체 0, 버전 필드 0
- [ ] `PerBackendVerdict` 무변경 (`git diff` 로 확인)
- [ ] 폴백 작업 디렉터리가 백엔드에 전달되지 않았음 확인(REQ-ABI-006 단서 — `git diff`로 `performCodexAudit` / `performGLMAudit` 인자 무변경)
- [ ] 측정 트리의 커밋 SHA를 모든 판정 인용에 붙임

## §D.2 판정에 쓰지 않는 것

- 설치본(`~/go/bin/moai`)의 출력. 이 카드의 판정은 트리에서 빌드한 테스트 바이너리로만 낸다 — 설치본으로 재는 것은 이 카드가 닫으려는 결함 그 자체다.
- 전체 스위트(`go test ./...`) 로컬 실행. 전 패키지 판정은 CI 몫(CLAUDE.local.md §4).
