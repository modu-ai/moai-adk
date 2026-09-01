---
id: SPEC-AUDIT-BUILD-IDENTITY-001
title: 감사 판정이 자기를 낸 바이너리의 커밋을 밝힌다 — 259 커밋 뒤처진 판정을 사후에 가려내기
version: "0.1.2"
status: draft
created: 2026-09-01
updated: 2026-09-01
author: manager-spec
priority: High
phase: "v3.1.5 target"
module: internal/cli, internal/binlag, pkg/version
lifecycle: spec-anchored
tags: "audit, build-identity, binary-lag, attribution, fail-open, additive-json, mcp"
tier: S
---

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 |
|---|---|---|---|
| 0.1.0 | 2026-09-01 | manager-spec | 최초 작성(카드 t248, plan-phase). 이 워크트리(`WT-audit-binary-sha`, base `64bba61aa`)에서 직접 읽은 좌표 위에 세움. 요구 8 / 수락 9 |
| 0.1.1 | 2026-09-01 | manager-spec | 운영자 결정 2건 반영(plan-phase, 코드 변경 0). ① 신원은 **평탄 형제 필드** `build_commit` / `build_lag` — 중첩 객체 없음(같은 구조체의 `synthesis_note`·`gate_unmet` 선례, Tier S 최소 변경, 두 번째 소비자 부재). ② 지연 권고도 형제 필드(`build_lag`). `build_version` 필드는 **싣지 않는다** — 버전이 곧 뮤턴트이므로 커밋 옆에 두면 틀린 쪽을 읽게 된다. `PerBackendVerdict` 무변경. REQ-ABI-003/004 재서술. plan.md 미해결 표시 0건 |
| 0.1.2 | 2026-09-01 | manager-spec | plan-audit FAIL(t248, 1/1) 결함 수리 — 코드 변경 0. D-3: REQ-ABI-006에 빈 `projectRoot` → `os.Getwd()` 폴백 명시(`doctor.go:521` 선례). D-4: 지연 AC가 세 핸들러 전수 + `StatusFresh` 대조군. D-1: 소스 스윕 기대를 기준선 3좌표 정확 집합 술어로 교체(본 트리 실측 재확인). D-2+D-5: 모양 기준을 AC-ABI-001에 병합(비어있지 않음 전제) — 수락 9→8, Tier S 상한 회복. D-6: §1.4가 두 번째 뮤턴트(M2)를 명명. D-7: 좌표 3건 정정(1493/245/:32, 본 트리에서 각 줄 검증). D-8: `version.Commit` ∈ {`""`,`none`,`unknown`}에서 `build_commit` 생략 명시. plan.md 낡은 AC 번호 2건 재번호 |

---

## §1 문제 — 판정은 남았는데, 그 판정을 누가 냈는지가 남지 않았다

감사 한 라운드 전체가 소스 트리보다 **259 커밋 뒤처진 설치본**으로 서비스됐다. 그 자체는 이미 알려진 결함 축(카드 t366 / SPEC-BINLAG-INVOCATION-001)이고, 이 카드가 닫으려는 것은 그 축이 아니다. 이 카드가 닫는 것은 **사후 판독 불가**다.

어느 판정도 자기를 낸 바이너리를 밝히지 않았다. 그래서 라운드가 끝난 뒤 누구도 "이 판정이 현재 코드를 잰 것인가"를 되물을 수 없었고, **이미 고쳐진 결함이 살아 있는 것처럼 보였다** — 앞선 카드의 F2/F5가 정확히 이 방식으로 오판됐다. 결함은 판정의 정확성이 아니라 판정의 **귀속 가능성**에 있다.

### §1.1 이 트리에서 실제로 읽은 좌표

세 감사 진입점의 출력 자료형에 빌드 신원이 **없다**:

    internal/cli/mcp_codex.go:262   type ReviewOutput struct
        verdict / summary / findings / next_steps / synthesis_note / gate_unmet — 빌드 신원 없음
    internal/cli/mcp_convergence.go:89   type PerBackendVerdict struct   — 없음
    internal/cli/mcp_convergence.go:106  type ConvergenceResult struct   — 없음

지속되는 기록도 같은 결손을 상속한다. `persistConvergenceResult`(`internal/cli/mcp_convergence.go:646`)는 `ConvergenceResult`를 그대로 마샬해 `<projectRoot>/.moai/state/audit-multi/<sessionID>.json`에 쓴다 — 구조체에 없는 것은 파일에도 없다.

사람이 읽는 한 줄도 마찬가지다. `internal/cli/mcp_audit_multi.go:145`의 텍스트 결과는 `audit_multi: overall=<v> (<residual risk note>)`뿐이다.

### §1.2 신원은 이미 이 바이너리 안에 있다 — 판정 경로에 닿지 않을 뿐이다

없는 것은 신원 자체가 아니라 **배선**이다.

    internal/cli/mcp_server_runtime.go:63   {pid, version, commit, build_date, executable}
        → .moai/state/mcp-server/<pid>.json 에 서버 프로세스마다 이미 찍힌다
    pkg/version   GetVersion() / GetCommit() / GetBuildID()
        GetBuildID 는 태그+거리+해시이며, 의도적으로 Version 으로 폴백하지 않는다
        (pkg/version/version.go:37, 그 근거 문장은 :32)

`version.GetCommit()`의 현재 소비자는 `internal/cli/version.go:40`, `internal/cli/doctor_mcp_version.go:40`, `internal/cli/mcp_server_runtime.go:63`, `internal/cli/doctor.go:530`, `internal/hook/session_start_binary_lag.go:55` — **감사 경로는 하나도 없다**(측정: `grep -rn "GetCommit()\|GetBuildID()" internal/ --include='*.go' | grep -v _test.go`, 이 워크트리 base `64bba61aa`).

### §1.3 비교기도 이미 있다 — 두 번째 구현을 만들 이유가 없다

범위 3항(바이너리 커밋과 리뷰 대상 트리 HEAD의 이격 경고)의 비교는 `internal/binlag`이 이미 수행한다: `binlag.Evaluate(ctx, binlag.Request{Dir, BinaryCommit, BinaryVersion}) Verdict`, 상태 `StatusFresh | StatusBehind | StatusDivergent | StatusNotApplicable`, 렌더러 `binlag.Short` / `binlag.Advisory` / `binlag.RemedyCommand`. 소비자 선례는 `internal/cli/doctor.go:518`.

`binlag.Request.BinaryVersion`은 문서화된 대로 **판정에 참여하지 않는다** — 시맨틱 버전 문자열은 지연을 판정할 수 없다. 이 사실이 §3의 반뮤턴트 기준의 근거다.

세 진입점 모두 이미 리뷰 대상 트리를 이름으로 받는다(`resolveToolProjectRoot` / `resolveOptionalToolProjectRoot` — `mcp_glm.go:245`, `mcp_codex.go:1493`, `mcp_audit_multi.go:71`).

**다만 `audit_multi`에서는 그 값이 흔히 비어 있다.** `resolveOptionalToolProjectRoot`는 호출자가 `project_root`를 생략하면 `("", nil)`을 돌려주고(`internal/cli/mcp_project_root.go:101-107`), 생략은 primary 체크아웃 세션의 **문서화된 정상 호출**이다(`moai-mcp-tools.md` § The `project_root` input). 빈 값에서 비교를 건너뛰면, 낡은 감사를 서비스할 확률이 가장 높은 바로 그 호출에서 지연 권고가 영영 발화하지 않는다. 그래서 REQ-ABI-006은 프로세스 작업 디렉터리 폴백을 요구하며, 그 폴백은 이 트리에 이미 있는 선례를 그대로 따른다:

    internal/cli/doctor.go:518   func checkBinaryFreshness(...)
    internal/cli/doctor.go:521       cwd, err := os.Getwd()
                                     → binlag.Request{Dir: cwd, ...}

폴백은 fail-open을 깨지 않는다 — `binlag.gitCompare`가 비-git 디렉터리를 `StatusNotApplicable`로 처리한다. 그리고 폴백 값은 **비교에만** 쓰고 백엔드에 넘기지 않는다: `resolveOptionalToolProjectRoot`의 빈 반환은 「인자를 생략한 기존 호출자의 백엔드가 받는 것을 바꾸지 않는다」는 계약이며(같은 파일의 주석), 폴백을 백엔드까지 흘리면 그 계약을 깬다.

### §1.4 대표 뮤턴트 — 버전만 적는 구현

이 SPEC이 반드시 죽여야 하는 구현은 이것이다: **판정에 버전 문자열만 기록하고 커밋을 빠뜨리는 것**(예: `"v3.1.2"`). 같은 태그가 뒤처진 빌드와 현재 빌드를 똑같이 이름하므로, 버전만 있는 필드는 아무것도 주장하지 않는다. 이 트리의 기본값 `Version = "v3.1.3"`(`pkg/version/version.go:8`)은 ldflags 없는 모든 빌드에서 동일하게 나온다 — 즉 버전 필드는 커밋이 다른 두 빌드에서 바이트 동일할 수 있다. §3의 AC-ABI-004가 이 뮤턴트에 대해 실패하는 기준이다.

이 때문에 채택된 모양은 버전 필드를 **아예 싣지 않는다**(REQ-ABI-003): 커밋 옆에 버전이 함께 실리면 소비자가 틀린 쪽을 읽을 여지가 생긴다. 버전이 필요해지는 날이 오면 그때 카드가 더한다.

뮤턴트는 사실 **둘**이고, `acceptance.md`의 `[HARD]` 픽스처 규율이 값하는 것은 두 번째다. 위에 적은 M1(커밋을 비우고 버전만 싣는 구현)은 「`build_commit`이 비면 실패」라는 단언 하나로 픽스처와 무관하게 죽는다. 픽스처 규율이 죽이는 것은 **M2 — 필드 이름은 `build_commit`인데 값이 버전 문자열인 구현**이다: 버전까지 다른 픽스처 쌍에서는 두 값이 서로 달라 「두 커밋이 다르다」를 통과해 살아남고, **버전이 같고 커밋만 다른** 쌍에서만 두 값이 같아져 죽는다. 그래서 픽스처 쌍의 모양이 협상 대상이 아니다.

---

## §2 요구 (GEARS)

**REQ-ABI-001** — The `audit_multi`, `codex_audit`, and `glm_audit` verdict outputs shall each carry the build commit of the binary that produced the verdict.

**REQ-ABI-002** — **When** a convergence result is persisted to `<projectRoot>/.moai/state/audit-multi/<sessionID>.json`, the persisted record shall carry the same build commit the returned verdict carried.

**REQ-ABI-003** — The recorded build identity shall be keyed on the commit hash, carried in a field named `build_commit`. The verdict output shall not carry a version-string field, because one version string names both a lagging build and a current one, and a version field shipped beside the commit invites a consumer to read the wrong one.

**REQ-ABI-004** — The build identity shall be carried as flat sibling fields on `ReviewOutput` and on `ConvergenceResult` — `build_commit` and `build_lag` — each additive and JSON-omittable, following the existing `synthesis_note` / `gate_unmet` precedent on the same struct (`internal/cli/mcp_codex.go:262`), so that every existing consumer's parse of an audit result remains valid. A nested build-identity object shall not be introduced.

**REQ-ABI-005** — An audit shall not fail, block, error, or change its verdict because the build identity is unavailable. **Where** the binary carries no commit metadata (a dev build, `commit` unset or `"none"`) or the reviewed tree is not a git working tree, the audit shall complete exactly as it does today and shall emit no identity-derived warning.

**REQ-ABI-006** — **When** the binary's build commit is a strict ancestor of the reviewed tree's HEAD, the verdict shall carry a lag advisory naming both commits. **Where** no reviewed tree was named by the caller, the comparison shall fall back to the process working directory, following the precedent already in the tree at `internal/cli/doctor.go:521` (`checkBinaryFreshness`, `:518`), rather than skipping the comparison. The fallback directory shall be used for the comparison only and shall not be passed to any audit backend, because `resolveOptionalToolProjectRoot` returns the empty string deliberately so that an absent `project_root` keeps handing `audit_multi`'s fan-out no `cwd` at all (`internal/cli/mcp_project_root.go:101-107`). The comparison shall be obtained from `internal/binlag` through its existing `Evaluate` seam; a second ancestry comparison shall not be implemented.

**REQ-ABI-007** — The three audit entry points shall report build identity through one shared constructor and one field shape. A per-entry-point identity assembly shall not be written.

**REQ-ABI-008** — The change shall not introduce a new MCP tool, shall not introduce a new verdict enum value, and shall not alter `overall_verdict`, `disagreement_flag`, or the 4-case convergence policy.

### §2.1 요구 ↔ 수락 추적

| 요구 | 수락 |
|---|---|
| REQ-ABI-001 | AC-ABI-001 |
| REQ-ABI-002 | AC-ABI-002 |
| REQ-ABI-003 | AC-ABI-003 (반뮤턴트) |
| REQ-ABI-004 | AC-ABI-004 |
| REQ-ABI-005 | AC-ABI-005 |
| REQ-ABI-006 | AC-ABI-006, AC-ABI-007 |
| REQ-ABI-007 | AC-ABI-001 (모양 절) |
| REQ-ABI-008 | AC-ABI-008 |

---

## §3 수락 기준

전체 Given-When-Then 서술과 판정 명령은 `acceptance.md`에 있다. 이 절은 기준 목록과 그 기계적 판정 수단만 든다.

| ID | 기준 | 판정 수단 |
|---|---|---|
| AC-ABI-001 | 세 진입점의 결과에 `build_commit`이 **비어있지 않게** 실재하고, 세 결과의 키 모양이 같다 | 비어있지 않음을 먼저 단언하는 Go 테스트 |
| AC-ABI-002 | 지속된 `<sessionID>.json`이 반환값과 같은 커밋을 담는다 | 파일을 되읽어 비교하는 테스트 |
| AC-ABI-003 | **반뮤턴트** — 버전만 기록한 구현(M1)과 이름만 `build_commit`인 구현(M2) 둘 다에서 실패한다 | 버전 동일·커밋 상이 픽스처 쌍 |
| AC-ABI-004 | `build_commit`·`build_lag`이 빈 결과의 JSON 키 집합이 변경 전과 같다 | `omitempty` 왕복 테스트 |
| AC-ABI-005 | 커밋 미상(`""`/`none`) 3상태에서 감사가 완주하고 `build_commit`이 생략된다 | 상태별 fail-open 테스트 |
| AC-ABI-006 | 조상 빌드에서 `build_lag`이 **세 진입점 모두** 두 커밋을 명시한다(빈 `projectRoot` 경로 포함) | 테이블 구동 `binlag.Comparer` 스텁 |
| AC-ABI-007 | 감사 경로에 `binlag` 외의 조상 비교가 없다 | 심 카운터 + 기준선 3좌표 정확 집합 스윕 |
| AC-ABI-008 | 도구 수·verdict 열거값·수렴 정책이 불변이다 | 기존 수렴 테스트 무수정 통과 |

수락 기준은 **8건**으로, Tier S 상한(요구 8 / 수락 8)을 지킨다. 전체 Given-When-Then 서술은 `acceptance.md`에 있다.

---

## §4 범위 밖

### Out of Scope — 바이너리 지연 자체의 수리
- `make build && make install`을 자동 수행하지 않는다. 이 SPEC은 **판정에 신원을 남기는 것**까지이며, 재설치는 사람의 몫이다.
- 공용 설치본(`~/go/bin/moai`) 갱신은 이 SPEC의 어떤 절차에도 포함되지 않는다.

### Out of Scope — 이미 닫힌 지연 표면
- `moai doctor`의 Binary Freshness(`internal/cli/doctor.go:518`), 세션 시작 권고(`internal/hook/session_start_binary_lag.go`), 실행 중 MCP 서버 대조(`internal/cli/doctor_mcp_version.go`)는 이미 존재한다. 재구현하지 않고 손대지 않는다.
- `internal/binlag`의 판정 논리, 관용 경로, 권고 문구는 카드 t326이 닫았다. 다시 열지 않는다.

### Out of Scope — 감사 품질과 수렴 정책
- 4-케이스 수렴 정책, `overall_verdict` 의미, `disagreement_flag` 산출은 건드리지 않는다.
- 백엔드 프롬프트, 모델 핀, 게이트 기본값은 이 카드의 범위가 아니다.

### Out of Scope — 새 표면의 신설
- 새 MCP 도구를 만들지 않는다. 새 verdict 열거값을 만들지 않는다.
- 새 Go 패키지를 만들지 않는다(Tier S). 신원 조립은 기존 `internal/cli` 안에 둔다.
- 중첩 신원 객체를 만들지 않는다. `PerBackendVerdict`(`internal/cli/mcp_convergence.go:89`)에는 신원을 붙이지 않는다 — 한 바이너리가 세 백엔드를 서비스하므로 같은 값의 3중 기록이 된다.
- 판정 출력에 버전 문자열 필드를 싣지 않는다(REQ-ABI-003).

### Out of Scope — 감사 외 도구의 신원 기록
- `verify_snapshot`, `spec_audit`, `graph_*` 등 다른 MCP 도구의 출력에 신원을 더하는 것은 이 카드가 다루지 않는다. 필요하면 별도 카드다.

---

## §5 잔여 위험

- **신원이 있다는 것과 인용자가 읽는다는 것은 다르다.** 필드를 더해도, 판정을 인용하는 사람이 그 필드를 보지 않으면 오귀속은 되풀이된다. 이 SPEC은 판독 가능성만 세우며, 인용 규율은 `verification-claim-integrity.md` §2 소관이다.
- **MCP 서버는 오래 산다.** 서버 프로세스는 시작 시점의 빌드를 계속 서비스하므로, 재설치 후에도 재연결 전까지는 옛 커밋이 정직하게 기록된다. 이는 결함이 아니라 정확한 기록이지만, 읽는 쪽이 "고쳤는데 왜 옛 커밋인가"로 오독할 수 있다.
- **트리 HEAD는 감사 도중에도 움직일 수 있다.** 비교는 호출 시점의 스냅숏이며, 그 이상을 주장하지 않는다.
- **폴백 디렉터리는 리뷰 대상 트리와 다를 수 있다.** `project_root`가 생략된 호출에서 비교는 프로세스 작업 디렉터리를 쓴다. 두 트리가 다르면 `build_lag`은 「작업 디렉터리 기준 지연」을 말하는 것이지 리뷰 대상 트리 기준이 아니다. 이는 비교를 아예 하지 않는 것보다 낫지만, 인용자는 이 차이를 알아야 한다. 비-git 작업 디렉터리에서는 `StatusNotApplicable`로 조용히 침묵한다.
- **`build_lag`이 비었다는 것은 두 가지를 뜻한다** — 비교했고 최신이거나, 비교 자체가 불가능했거나. 이 SPEC은 둘을 구별하는 신호를 만들지 않는다. 구별이 필요해지면 별도 카드다.
