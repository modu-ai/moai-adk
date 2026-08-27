# plan.md — SPEC-CI-DOCTOR-BIN-001

## §A Context

- **브랜치/기준점**: `WT-ci-doctor-bin` @ `4fdbd55c1` (base = origin/develop). 카드 t346.
- **산출물**: `.moai/specs/SPEC-CI-DOCTOR-BIN-001/{spec,plan,progress}.md` (Tier S — AC 는 spec.md §3 인라인).
- **결함 한 줄**: `checkAgentEmitEmbedAgainst`(`internal/cli/doctor_agentemit_embed.go:119-124`)가 적용 가능 트리에서 `bin/moai` 부재를 `CheckFail`로 보고 → CI 체크아웃은 바이너리를 절대 갖지 않으므로 `TestRunDoctor_*` 전부가 `doctor: 1 check(s) failed`로 실패.
- **수리 방향 (운영자 결정, 옵션 (a))**: 바이너리 부재 → 정보성 skip (`ok` + 이유 메시지), fail-open. 선례: `checkBinaryFreshness` (카드 t184, `internal/cli/doctor.go:502` — 판정 불가 시 `ok` + 메시지, doctor 를 gate 하지 않음).
- **관련 인프라 (PRESERVE 대상)**: `internal/cli/doctor.go`의 레지스트리(`:205` 등록)·`countFailedChecks`/`doctorExitStatus`(`:101`, `:140-146`), `findEmbedCheckRoot` 상향 탐색, `MOAI_EMBED_CHECK_BIN` override, `uikit` 상태 열거 — 전부 수정 대상 아님.
- **plan-auditor verdict**: PASS 0.93 (Tier S 임계 0.75, skip-eligible; iter 1/1 — `.moai/reports/t346/verdict-plan.md`). minor 3건 처분 — **D1**(`related_specs` 비스키마 필드): 수용된 minor 로 문서화. 비차단 참조를 차단 의미의 `depends_on:` 으로 옮기면 관계가 왜곡되고 스키마 등록은 범위 확장이라 하지 않았다(감사 판정 "스킵해도 무방"). **D2**(supersession 선언의 검증 계층 파생 미지목): 적용 — spec.md REQ-CDB-001 주해에 파생물 한 문장 추가. **D3**(CI 전칭 과잉 일반화): 적용 — §1 전칭을 go test 잡 한정으로 축소 + `lint`/`constitution-check` 대조 각주. D2·D3 수리로 artifact-hash 불변이 깨지므로 감사 권고대로 Phase 1 재실행이 정규 경로다.

### §A.5 PRESERVE 목록 (범위 절제)

`internal/cli/doctor_agentemit_embed.go`와 `internal/cli/doctor_agentemit_embed_test.go` **이외의 모든 파일**. 특히: `internal/cli/doctor.go`, `internal/cli/doctor_codex*.go`, `internal/cli/uikit/**`, `internal/kanban/**`, `.github/workflows/**`, `internal/template/**`.

## §B Known Issues (본 SPEC 관련만)

- **B-상태열거**: `uikit.CheckStatus`는 `ok`/`warn`/`fail` 3값뿐(`internal/cli/uikit/types.go:12-17`) — skip 상태 없음. 스킵은 `ok` + 메시지로 표현하는 것이 코드베이스 선례(`checkBinaryFreshness`, 임베드 검사 자신의 not-applicable 분기)와 유일한 선택지다.
- **B-기대역전**: `TestAgentEmitEmbed_MissingBinaryFails`(`doctor_agentemit_embed_test.go:182-190`)가 현재 결함을 **단언하는 테스트**다. 수리는 이 테스트의 기대를 뒤집는 것을 포함한다 — 테스트 이름도 행동에 맞게 바꾼다(예: `MissingBinarySkips`).
- **B-MX계약**: `checkAgentEmitEmbedAgainst`의 `@MX:REASON`("branch ORDER is the contract")은 변경 후에도 유효하다 — 분기 추가/삭제 없이 기존 bin-absent 분기의 verdict 만 바뀐다. `@MX:SPEC` 태그는 run-phase 에서 새 SPEC ID 를 병기해 갱신한다(MX 태그는 자율 운영).
- **B-CI판정면**: 로컬 패키지 게이트 통과 후 원격 `origin/develop` CI가 전체 판정 주체다(git-flow 레인 규율). 수리가 착지해야 develop CI 가 녹색으로 돌아온다.
- **B-동반실패**: 카드 t346 기록의 `TestConcurrencyStress`(internal/kanban) 동반 실패는 귀속 미확립 — 본 SPEC 범위 밖(spec.md §4). 수리 후 재발 시 별도 조사.

## §C Pre-flight

```bash
git -C <worktree> branch --show-current        # WT-ci-doctor-bin
git -C <worktree> rev-parse --short HEAD       # 4fdbd55c1 (진행 중 재확인)
ls bin/moai                                    # absent — RED 전제
go test ./internal/cli/ -run 'TestRunDoctor_WithExport$' -count=1   # RED 재현 (2026-08-28 실측: FAIL, doctor: 1 check(s) failed)
go vet ./internal/cli/                         # baseline
```

## §D Constraints

- PRESERVE 목록(§A.5) 밖 쓰기 금지. 특히 `doctor.go`·`uikit`·CI 워크플로·`internal/kanban` 비접촉.
- `--no-verify`, `--amend`, force-push 금지. Conventional Commits + 카드 id 병기(`(t346)`).
- 커밋 전 재판독: `git rev-parse --short HEAD` + `git branch --show-current` (AGENTS.md §2).
- 테스트는 영향 패키지 한정(`go test ./internal/cli/...`) — 전체 스위트 로컬 금지.
- REQ-CDB-003 위반(바이너리 있을 때 fail 경로 약화)은 수리 실패로 본다 — 스킵은 "판정 불가"에만 적용.

## §E Self-Verification (manager-develop 보고 양식)

E1 AC 매트릭스(AC-CDB-001~004, 명령+출력 전문) / E2 `GOOS=windows GOARCH=amd64 go build ./...` / E3 `go test -cover ./internal/cli/...` (패키지 임계 85%) / E4 서브에이전트 경계 grep(해당 없음 — 표면 미변경, 해당 시 보고) / E5 `golangci-lint run --timeout=2m` (신규 vs baseline 구분) / E6 커밋 SHA + push 상태 / E7 blocker / E8 RED 전문(§C 의 재현 출력 — 구현 전 캡처본).

## §F Milestones (결정 가역성 순 — 바뀔 가능성이 큰 결정이 앞선다)

### M1 — verdict 역전 + 스킵 관측성 (Priority High, AC-CDB-001·002)

`internal/cli/doctor_agentemit_embed.go`의 bin-absent 분기(:119-124)를 `CheckFail` → `CheckOK` 정보성 skip 으로:
- 메시지는 부재한 대상 경로와 처방(`make build` 또는 `MOAI_EMBED_CHECK_BIN` override)을 이름으로 넣는다 (REQ-CDB-002).
- Detail(verbose)엔 판정 대상이 무엇이었는지 한 줄.
- `TestAgentEmitEmbed_MissingBinaryFails` → 스킵 기대로 갱신: status `ok` + extractor 미호출 플래그 + 메시지 내용 단언 (AC-CDB-001·002의 잠금 테스트).
- 검증: `go test ./internal/cli/ -run 'TestAgentEmitEmbed' -count=1`.

### M2 — 수렴 검증 + 패키지 게이트 (Priority High, AC-CDB-003·004)

- bin-absent 수렴: `go test ./internal/cli/ -run 'TestRunDoctor_WithExport$' -count=1` → PASS (`bin/moai` 없이).
- 판정 보존: `go test ./internal/cli/ -run 'TestAgentEmitEmbed_(ExtractionErrorFails|DriftFailsAndNamesPath|PartialExtractionFails)$' -count=1` → PASS (수정 없이 — 비회귀).
- bin-present 대조: `make build` 후 동일 명령 재실행 → PASS.
- 패키지 게이트: `go vet ./internal/cli/` + `go test ./internal/cli/...` + `golangci-lint run --timeout=2m`, `GOOS=windows GOARCH=amd64 go build ./...`.
- 커밋 + push → origin/develop CI 판독 (전체 스위트 판정은 CI 몫).

## §G Anti-Patterns (금지)

- 스킵을 `CheckWarn`으로 바꾸는 것 — 선례(`checkBinaryFreshness`)는 `ok`다. warn 은 "판정 불가"가 아니라 "주의"라는 다른 의미를 갖고, 배포 프로젝트가 아닌 개발 트리에서 매 실행 노이즈가 된다.
- bin-absent 스킵을 이유로 추출 실행 시도 — 존재하지 않는 바이너리를 exec 하면 오류가 난다. 스킵은 실행 전에 끝난다.
- 스킵 메시지를 "ok" 한 단어로 남기는 것 — REQ-CDB-002 위반 (스킵과 비활성화를 못 가른다).
- `uikit`에 새 상태 추가 — 범위 밖(§A.5), 3값 열거는 코드베이스 전체가 전제한다.

## §H Cross-References

- `.moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/spec.md` — 검사의 기원 SPEC. REQ-AEL-004 bin-absent-failure 절을 본 SPEC REQ-CDB-001이 부분 대체.
- `internal/cli/doctor.go:502` — `checkBinaryFreshness` (t184 fail-open 선례).
- 카드 t346 본문 — CI 실측 전문, 9종 테스트 목록, make build 대조 실측.
