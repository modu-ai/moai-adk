# SPEC-AUDIT-BUILD-IDENTITY-001 — sync-audit 보고서 (카드 t248)

- 감사자: sync-auditor (독립 판정)
- 측정 트리: `2d336d8545e7cb9adbaf2e2e8028907f3fc5f7c7` (branch `WT-audit-binary-sha`, base `64bba61aa`) — 모든 판정은 이 트리에서 직접 재측정했다
- 판정일: 2026-09-01
- MCP provenance: 이 감사는 moai MCP 서버(스스로 v3.1.2/`343399d2f` 라고 자기보고, 본 트리보다 뒤처진 빌드)의 측정값을 **하나도 인용하지 않았다** — 전부 트리 직접 판독 + 로컬 `go test`다. 인용된 MCP 기반 판정 0건이므로 provenance 귀속 조항은 적용 대상이 없다.

## 종합 판정

**PASS-WITH-DEBT** — 차단 결함 0건. 8/8 AC를 본 트리에서 재실행해 전부 PASS를 직접 관측했고, 반뮤턴트 M1/M2를 `go test -overlay`로 실제 심어 두 다 죽는 것을 관측했다. 남은 것은 전부 optional 성격의 부채 5건(아래 F1–F5)이다.

| Dimension | Score | Verdict | 핵심 근거 |
|---|---|---|---|
| Functionality (40%) | 100 | PASS | 8/8 AC 재실행 PASS(스웁 집합 공집합 아님: AC-001..007 = 7테스트+4서브테스트, AC-008 = 29테스트), M1/M2 뮤턴트 사살 관측, 단일 구현 3호출점 확인 |
| Security (25%) | 100 | PASS | 새 신뢰경계 0(`projectRoot`는 변경 전에도 백엔드/diff 수집기가 이미 받던 값), 신원 문자열은 빌드타임 상수, `exec`/`Getenv`/secret 토큰 0히트, fail-open 유지 |
| Craft (20%) | 92 | PASS | lint 0 issues, gofmt/vet clean, 손댄 함수 커버리지 93.3–100%(직접 재측정), 테스트 품질 높음(공허화 방어 선행 단언·대조군·정확 집합 스웁) — 패키지 커버리지 80.0%는 기존 부채(F4), 스웁 좌표 고정의 취약성(F2) |
| Consistency (15%) | 90 | PASS | `synthesis_note`/`gate_unmet` `omitempty` 선례 정확히 계승, 네이밍·주석(영어)·포맷 일관 — fan_in=3인데 `@MX:ANCHOR` 대신 `@MX:NOTE`(F3), GLM 핸들러 순서 변경(F1) |

조화평균: 4 / (1/100 + 1/100 + 1/92 + 1/90) ≈ **95.3** → must-pass(Functionality, Security) 양쪽 독립 통과.

## 1. AC 재판정 (mission 1) — 본 트리 `2d336d854`에서 직접 실행

명령: `go test ./internal/cli/ -run 'TestAuditVerdictCarriesBuildCommit|TestPersistedConvergenceCarriesBuildCommit|TestBuildIdentityVersionAloneIsRejected|TestBuildIdentityOmittedWhenAbsent|TestAuditCompletesWithoutBuildIdentity|TestAuditLagAdvisoryNamesBothCommits|TestAuditLagUsesBinlagSeam' -count=1 -v` → 종료코드 0. 출력(발췌, verbatim):

```
--- PASS: TestAuditVerdictCarriesBuildCommit (0.89s)
--- PASS: TestPersistedConvergenceCarriesBuildCommit (0.00s)
--- PASS: TestBuildIdentityVersionAloneIsRejected (0.92s)
--- PASS: TestBuildIdentityOmittedWhenAbsent (0.00s)
--- PASS: TestAuditCompletesWithoutBuildIdentity (1.89s)   # 서브테스트 4건 전부 실행: commit="none"/""/"unknown"/no_git_tree_anywhere
--- PASS: TestAuditLagAdvisoryNamesBothCommits (1.73s)
--- PASS: TestAuditLagUsesBinlagSeam (0.60s)
ok  	github.com/modu-ai/moai-adk/internal/cli	7.036s
```

AC-ABI-008: `go test ./internal/cli/ -run 'TestConverge|TestRunMultiAudit|TestAuditMulti' -count=1` → `ok ... 0.749s`. 공허 초록 배제: 동일 셀렉터 `-v`에서 `--- PASS` 29건(셀렉터가 실제로 29개 테스트/서브테스트를 훑었음을 관측). 기존 테스트 기대값 수정 0건(diff로 확인 — 이 카드의 커밋은 기존 테스트 파일을 건드리지 않았다).

DoD 범위 판정:
- `go test ./internal/binlag/... -count=1` → `ok ... 2.946s`
- `go vet ./internal/cli/... ./internal/binlag/...` → 출력 없음, exit 0
- `golangci-lint run ./internal/cli/... ./internal/binlag/...` → `0 issues.`
- `gofmt -l` (신규 2파일 + 수정 3파일) → 출력 없음

## 2. 반뮤턴트 공격 (mission 2) — M1/M2를 실제로 심어 관측

트리 파일을 건드리지 않기 위해 `go test -overlay`(뮤턴트 사본은 전부 `/tmp`)으로 심었다. 뮤턴트는 원본과 정확히 1줄만 다르다(`diff`로 확인).

**M2** (필드 이름 `build_commit`에 버전 문자열 — `return buildCommit, ...` → `return version.GetVersion(), ...`):

```
--- FAIL: TestBuildIdentityVersionAloneIsRejected (0.29s)
    mcp_build_identity_test.go:369: codex_audit build A: "build_commit" = "v3.1.3", want "abc123def456789"
FAIL
```

**M1** (커밋 생략 — `return "", ...`):

```
--- FAIL: TestBuildIdentityVersionAloneIsRejected (0.28s)
    mcp_build_identity_test.go:369: codex_audit build A: "build_commit" key ABSENT — the verdict carries no build identity
FAIL
```

두 뮤턴트 모두 죽는다 — 관측된 사실. 다만 **죽인 주체**에 대한 정밀 판정이 mission 2의 후반부 질문이다:

**"버전 동일·커밋 상이" 픽스처 규율은 과연 하중을 지니는가** — 실험: (a) 정확-값 단언을 SPEC 절 2의 문자 그대로(비어있지 않음만)로 완화하고, (b) 픽스처를 버전-상이(`"v3.1.2"` vs `"v3.1.3"`)로 바꾼 변형 테스트를 `/tmp`에 만들어 M2와 함께 실행:

```
--- FAIL: TestBuildIdentityVersionAloneIsRejected (1.01s)
    mcp_build_identity_test.go:377: codex_audit build B: key "build_commit" carries the version string "v3.1.3" — build identity keyed on version, not commit (M2 mutant)
    mcp_build_identity_test.go:377: glm_audit build B: (동일)
    mcp_build_identity_test.go:377: audit_multi build B: (동일)
FAIL
```

**build A의 버전-기록이 스캔을 통과해 살아남았다.** 원인: `assertNoVersionString`의 값 비교가 `version.GetVersion()`과의 대조인데, 스캔 시점에는 두 번째 `withVersionIdentity`(build B)가 전역을 이미 덮어써 build A 시점의 버전이 아니다. 같은 변형 테스트를 올바른 구현에 돌리면 PASS(대조군 — 변형 자체가 항상-적색이 아님을 확인).

판정:
- **구현된 테스트 기준**: 픽스처 규율은 하중을 지니지 **않는다**. M2는 정확-값 단언(테스트 :244)이 픽스처와 무관하게 죽인다(위 M2 실행이 그 증거 — 동일-버전 픽스처에서 죽었고 죽은 지점이 정확-값 절이다). 3겹 방어(정확-값 + 라이브 버전 값 스캔 + 픽스처 규율) 중 셋째는 이중 안전장치다.
- **SPEC 절 그 자체(1/2/3) 기준**: 규율은 하중을 지닌다. 버전-상이 픽스처에서는 절 1이 M2를 죽이지 못하고(두 값이 다름), 절 3의 스캔은 **스캔 순서의 우연**(두 번째 identity가 아홉 살아 있음) 덕에 build B만 잡는다 — build A의 위조 신원은 관측되지 않은 채 남는다(위 출력이 실측). 동일-버전 픽스처에서만 절 1이 두 값을 같게 만들어 결정적으로 죽인다. SPEC §1.4/AC-ABI-003 [HARD]의 논거는 자기 절 집합 안에서 정확하다.
- 종합: 구현이 SPEC이 요구한 것보다 엄격해서 규율이 만족 상태로 남고, 공허한 규율은 아니다(변형 실험이 남는 구멍을 보여줬다). 결함 아님.

## 3. 단일 구현 (mission 3, AC-ABI-007/REQ-ABI-007)

`grep -rn "auditBuildIdentity" internal/ --include='*.go' | grep -v _test.go` — 프로덕션 호출점 정확히 3곳: `mcp_codex.go:1517`, `mcp_glm.go:239`, `mcp_convergence.go:541`(파일당 1회씩). 그 외에는 주석 언급뿐. 세 핸들러가 전부 binlag.Comparer seam을 지난다 — `TestAuditLagUsesBinlagSeam` 관측 1이 진입점별 카운터 증가를 단언하고 PASS.

소스 스웁(직접 실행, 본 트리):

```
$ grep -rn "merge-base\|is-ancestor" internal/cli --include='*.go' | grep -v _test.go
internal/cli/graph_stamp.go:68:  moai graph stamp codemaps --commit "$(git merge-base HEAD origin/main)"
internal/cli/graph_stamp.go:131:		`explicit commit anchor ... Use "$(git merge-base HEAD origin/main)" ...`)
internal/cli/mcp_review_material.go:95:		out, err := runReviewGit(root, "merge-base", ref, "HEAD")
```

히트 집합 = 기준선 3좌표와 정확히 동일(추가 0, 삭제 0). `mcp_review_material.go:95`의 비조상성 판정(D-1)대로 기준선에 남는다 — 이것은 리뷰 diff 기준점 해석이지 바이너리↔트리 조상 비교가 아니다.

`ConvergenceResult` 구축점 전수 조사: `converge()`(:236, 반환 직후 `runMultiAudit`가 신원을 채움 — :638 부근) + DQ-2 거부 리터럴(:550, 인라인 채움) + `multi_review_gate.go`의 zero-value 오류 반환 3건(소비자 쪽 파싱 실패 경로, 판정 출력이 아님). 판정을 만드는 모든 경로가 신원을 싣는다.

## 4. D-3 계약 (mission 4)

코드 판독 + diff: `auditBuildIdentity`의 cwd 폴백은 함수 내부에서만 살고 반환값은 문자열 2개뿐이다. 백엔드 인자 경로는 무변경 — `backendCall(gctx, s.name, target, focus, cfg.ProjectRoot)`(`mcp_convergence.go:609`, 폴백 토큰 없음), `performCodexAudit`/`performGLMAudit` 호출부(`:431`/`:439`, 이 diff가 건드리지 않음), codex `params["cwd"] = root`(해석된 호출자 값 그대로). 빈 `projectRoot` 호출에서 백엔드는 변경 전과 같은 `""`을 받는다. 테스트 쪽에서는 D-3 절이 `stub.lastReq.Dir == cwd`로 폴백 디렉터리에서의 비교 **수행**을 관측하고 PASS. ("비교 전용" 성질은 DoD가 지정한 검증 수단인 diff 판독으로 확인했다 — 위와 같다.)

## 5. fail-open · 생략 · 모양 (mission 5)

- `normalizeBuildCommit`(`mcp_build_identity.go:75-81`)이 `""`/`none`/`unknown`을 정확히 빈 문자열로 만들고, binlag 자체의 not-applicable 집합(`binlag.go:108-110`)과 동일하다 — `pkg/version/version.go`에서 직접 확인(`Commit = "none"` 기본값). 비교는 정규화 성공 **이후에만** 실행되므로 CHANGELOG의 "short-circuit happens BEFORE any git subprocess" 주장도 코드와 일치한다.
- `omitempty`로 빈 신원이 JSON 키에서 사라진다 — `TestBuildIdentityOmittedWhenAbsent`가 명시적 기대 키 집합(`verdict, summary, findings, next_steps` / `per_backend_verdicts, overall_verdict, disagreement_flag, residual_risk_note, fail_open_backends`)을 실제 구조체와 대조하며 PASS. 실패-open 3상태 + 비-git 상태에서 verdict 불변·양 필드 부재를 상태별로 각각 판정(서브테스트 4건 전부 실행 관측).
- 감사 출력 모양에 `version` JSON 태그 0건(`grep 'json:' ... | grep version` → 히트 없음). `PerBackendVerdict` 본문 무변경(diff로 확인 — 주석 언급 1건은 ConvergenceResult 쪽 문서).

## 6. sync close 무결성 (mission 6)

- **spec.md frontmatter 전이**: `git show b60ca5583 -- .../spec.md` — 변경은 정확히 1행 `status: in-progress → completed`. `updated:`는 plan-phase(v0.1.2)때 이미 2026-09-01이라 무변경이고 본문 무변경. 소유자 규정(manager-docs, sync 커밋에 completed 병합) 합법.
- **CHANGELOG 사실성**: 추가된 단일 항목의 검증 가능한 주장을 전부 본 트리에서 대조 — 평탄 형제 필드+omitempty(✓ diff), 단일 생성자 3진입점(✓ §3), 버전 필드 부재(✓ §5), binlag.Evaluate 재사용(✓), cwd 폴백 비교-전용(✓ §4), 3상태 생략+fail-open(✓ §5), `PerBackendVerdict`/수렴 정책/새 도구·열거값·패키지 무변경(✓ diff), README 4개국어 `build_commit`/`buildCommit` 0히트(✓ 재실행, exit 1 = 매치 없음), "버전은 태그 전까지 불신뢰" 근거(✓ `pkg/version` 주석과 일치). 8/8 PASS 재검증 트리 표기는 `feb272f70`(run 트리) — 본 감사는 `2d336d854`에서 재확인했고 두 트리의 `internal/`/`cmd/`/`pkg/`는 `git diff 1c3adc4d5..HEAD --stat`가 공집합으로 내용 동일이라 run 시점 §E.2/§E.3 수치의 귀속도 성립한다.
- **§E.4 신호**: `sync_commit_sha: "b60ca5583"`(백필 완료), `sync_status: complete`, B12 셀프테스트 3건 기록, docs-scope 결정, push 미수행 명시. 완전하다.
- **백필 패턴**: `2d336d854`는 progress.md 1행(placeholder → 실제 SHA)만 변경 — commit이 자기 해시를 못 쓰는 물리 제약의 정식 D3 면제 패턴. sync 커밋 `b60ca5583`은 코드 파일 0건 변경(3파일: spec.md/progress.md/CHANGELOG만). 정합.
- **§E.2 증거 귀속**: AC 매트릭스 각 행이 명령+실제 출력+커밋 SHA를 다는데, 본 감사의 재실행이 같은 결과를 재현했다(§1). RED 증거(E8)는 구현 전 트리 `fd26c6cf2`에서 캡처됐다고 기록 — AC-ABI-004의 RED 불가능성(회귀 가드 본성)을 RED 대신 "변경 전 상태 GREEN 관측"으로 정직하게 처분한 것도 합격.

## 7. 결함 목록 (전부 optional — 차단 0)

- **F1** [Low] [optional] `internal/cli/mcp_glm.go:222-237` — GLM 핸들러 재배치: `resolveToolProjectRoot`가 키 검사보다 앞으로 옮겨졌다. base `64bba61aa`에서는 (무효 root + GLM 키 부재) 조합이 fail-open inconclusive를 반환했으나 이제 tool error를 반환한다. REQ-ABI-001이 "모든 판정 출력이 신원을 싣는다"를 요구하므로(키-부재 inconclusive도 판정 출력이고 신원엔 root가 필요) 필연적 선택이고 codex_audit과 정렬되며 인라인 주석으로 문서화됐다 — 그러나 SPEC이 명시하지 않은 관측 가능한 행동 변경이다. REQ-ABI-005의 문자(신원 **불가능** 때문인 오류)는 위반하지 않는다. — 필요한 수리 없음. 행동 변경을 되돌리고 싶으면 키 검사 뒤에서 root를 해석하고 키-부재 판정에 신원을 안 싣는 대가를 SPEC 수준에서 결정할 것.
- **F2** [Low] [optional] `internal/cli/mcp_build_identity_test.go:588-592` — 스웁 기준선이 줄 좌표(`graph_stamp.go:68`/`:131`, `mcp_review_material.go:95`)에 고정돼 있어, 해당 파일들의 무해한 줄 이동(주석 한 줄 추가 등)만으로 `TestAuditLagUsesBinlagSeam`이 오탐 적색이 된다. 실패 방향은 안전한 쪽(침묵하는 초록이 아니라 시끄러운 빨강)이고 정확-집합 술어는 acceptance.md가 명시한 설계다. — 후속 카드에서 내용-앵커 매칭으로 바꾸거나 비용을 수용.
- **F3** [Low] [optional] `internal/cli/mcp_build_identity.go:37-41` — 프로덕션 호출점 3개(codex/glm/convergence)인 함수가 `@MX:NOTE`만 싣고 있다. fan_in ≥ 3에는 `@MX:ANCHOR`가 MUST인데(mx-tag-protocol.md + moai-constitution.md § MX Tag Quality Gates), sync 커밋 메시지의 "MX tag pass: no changes required"는 이 점에서 과대 서술이다. — 다음 파일 접촉 시 `@MX:ANCHOR` + `@MX:REASON`으로 승격(MX 태그는 자율 운용).
- **F4** [Info] [optional] `internal/cli` 패키지 커버리지 80.0%(본 트리 직접 재측정, `go test -coverprofile` → `coverage: 80.0% of statements`) — 문서된 패키지 하한 85%, critical 패키지 상한 90%(CLAUDE.local.md §6)에 못 미친다. 기존 baseline 수준이며 카드 자체가 pre-change baseline 미측정을 Gap으로 기록했다. 손댄 함수는 93.3–100%: `auditBuildIdentity`/`normalizeBuildCommit`/`handleCodexAudit`/`runMultiAudit` 100%, `handleAuditMulti` 93.3%, `handleGLMAudit` 96%. 이 카드 소관 밖의 상설 부채.
- **F5** [Info] [optional] AC-ABI-003의 [HARD] 픽스처 규율은 구현 테스트에서 하중을 지니지 않는다(§2 실험 — M2의 실제 사살자는 정확-값 단언). 규율 자체는 SPEC 절 집합 안에서 유효하고 이중 안전장치로 남는 것이 합리적이다. 행동 요구 없음.

## 권고

- F3은 다음 카드가 이 파일을 만질 때 1줄로 닫는다. 지금 전용 커밋을 만들 가치는 없다.
- F1의 행동 변경은 CHANGELOG/§E.4에 기록된 바 없으므로, 통합 전에 §E.4에 한 줄 부기("root 해석 선행으로 무효-root+무키 조합이 inconclusive→tool error")를 남기면 인용자 혼동을 막는다. 코드 변경은 불필요.
- F2/F4는 별도 카드 재료(스웁 내용-앵커화, internal/cli 커버리지 상향).

## Gaps (명시적으로 관측하지 않은 것)

- 설치본(`~/go/bin/moai`) 동작 — acceptance §D.2가 이 카드의 판정 대상에서 명시적으로 제외.
- CI(원격) 판정 — 브랜치 미푸시 상태이며 로컬 트리 판정만 존재한다.
- `internal/cli`·`internal/binlag` 밖의 전체 스위트 — mission 범위 제한에 따라 미실행(원격 CI 몫).
- 지속 파일 쓰기 **실패** 분기에서의 REQ-ABI-002(성공 경로만 테스트됨 — 실패 분기의 best-effort 성질은 기존 동작).

## 잔여 위험

- MCP 서버는 오래 산다 — 신원은 서빙 프로세스 시작 시점의 빌드를 말하며, 재설치 후 재접속 전까지는 옛 커밋이 정직하게 기록된다(SPEC §5가 이미 명시).
- `project_root` 생략 호출에서 `build_lag`은 "작업 디렉터리 기준 지연"이지 리뷰 대상 트리 기준이 아니다(SPEC §5 명시 — 인용자가 읽어야 할 차이).
- F2의 줄-좌표 스웁은 언젠가 무해한 리네임으로 오탐 적색을 낸다.
