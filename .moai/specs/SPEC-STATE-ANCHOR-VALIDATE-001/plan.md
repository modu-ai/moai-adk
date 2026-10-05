# Plan — SPEC-STATE-ANCHOR-VALIDATE-001

## §A 맥락

카드 **t537**. 기원: t510 sync-audit **F1** (`.moai/reports/t510/sync-audit.md:120` — develop 워크트리에서 전문 재확인 완료). 부모 SPEC: `SPEC-STATE-ANCHOR-001` (`status: completed` — `depends_on` 충족). 작업 트리 `.claude/worktrees/t537`, 브랜치 `WT-resolve-validate`, base `52f863f36`. 본 SPEC이 소유하는 요구사항 변경: 부모 plan §D1이 "중간 단계 삽입·순서 변경은 요구사항 변경(REQ-SA-002)"이라고 고정했으므로, 체인 1·2단의 반환 조건 강화(무조건 반환 → 검증-통과-폴스루)를 **명시적 요구사항 변경으로 선언**한다(spec.md §4 전언 블록). 부모 SPEC에 영향을 주는 부분: 부모 AC 중 체인 1·2단의 무조건 반환을 고정한 것은 없다 — 부모 acceptance.md는 B1/B2/B3의 착지 위치만 단언하고 체인 내부 조건을 고정하지 않는다(`TestResolve_ProjectDirWins`는 부모 AC가 아니라 **테스트 파일**의 자산이며, 아래 §D8이 갱신을 다룬다).

### Tier 판정 — S (판단 근거와 예외 1건)

- **S**: 생산 변경 1파일(`stateanchor.go`) + 그 테스트 1파일. LOC 10~20 수준. REQ 7 / AC 8 — Tier S 상한 8/8 이내. 폭발 반경이 패키지 안이다.
- **예외**: 카드 배차가 **acceptance.md를 명시 산출물로 지시**했다. Tier S 표준 산출물은 spec.md + plan.md(AC 인라인)이지만, 배차 지시를 따라 **spec.md + plan.md + acceptance.md + progress.md** 4파일로 세운다. plan-auditor 임계는 Tier S 기준 0.75를 적용한다. 이 예외는 frontmatter `tier: S`와 함께 기록된다.

### 개발 방법론

**TDD (RED-GREEN-REFACTOR)** — 부모와 동일 선택. 결함 수리(무검증 → 검증)이며, RED-first가 §D9로 구속된다. 기존 커버리지: 패키지에 테스트 7건 존재(`stateanchor_test.go` 직접 판독) — brownfield ≥10% 형태로 TDD 적합.

## §B 알려진 결함 (이미 측정됨 — 재론 금지)

1. `Resolve` 체인 1·2단이 `!= ""`만 확인하고 후보를 그대로 반환한다 (`stateanchor.go:64-76`, 본 트리 직접 판독).
2. 앵커-유도 MkdirAll 3곳(`context_usage.go:197`, `landed.go:256`, `github.go:365`)이 검증 없는 앵커를 **디렉터 생성**으로 승격시킨다 — stale/거짓 `project_dir`의 임의 경로 부활 기전.
3. **정정**: 배차문 근거가 앵커-유도로 나열한 `model_cache.go:55`는 홈-앵커다(`WriteModelCache(homeDir, …)` → `~/.moai/state/last-model.txt`, `model_cache.go:47-60` + `metrics.go:55` 직접 판독). 소비자 지도는 §1.3의 3곳 + B3 읽기가 정확하다.
4. `FromDirectory`(3단·B4)는 git 해석으로 존재성이 검증돼 있다 — 결함 아님, 건드리지 않는다.

## §C 사전 점검 (run-phase 진입 시 재측정 — 값이 다르면 멈추고 보고)

| 명령 | 기대값 |
|---|---|
| `git rev-parse --short HEAD` + `git branch --show-current` | `52f863f36` 계열 + `WT-resolve-validate` (다르면 흡수 후 재판정) |
| `go test ./internal/stateanchor/ -count=1` | ok — 기존 7테스트 전부 GREEN (baseline; §D8 재작성 전 상태) |
| `sed -n '64,76p' internal/stateanchor/stateanchor.go` | `if s.ProjectDir != "" { return s.ProjectDir }` 형태 존재 (무검증 재확인) |
| `grep -rn "stateanchor.Resolve" internal/ --include="*.go" \| grep -v _test` | `internal/statusline/state_anchor.go:34` 1매치 (호출자 단일성 재확인) |
| `grep -n "MkdirAll" internal/statusline/context_usage.go internal/statusline/landed.go internal/statusline/github.go` | 각 1매치 (부활 기전 표면 재확인) |

## §D 구속 조건 (재논의 금지)

- **D1 — 요구사항 변경의 소유.** 체인 순서(`project_dir` → `original_cwd` → git 워크업)와 3단 의미론은 불변이다. 변경되는 것은 1·2단의 **반환 조건**뿐이다. 새 단 삽입·재정렬은 본 SPEC의 위반이다(REQ-SAV-002의 체인 언급 참조).
- **D2 — 술어는 (a)+(b)다.** `filepath.IsAbs` AND `os.Stat` 성공(디렉터 모드). §5 격자의 채택 결정. IsAbs-only·exists-only·git-membership은 전부 기각됐다 — 기각 근거는 spec.md §5와 §3 뮤턴트 1~3에 있다. 운영자가 Implementation Kickoff Approval에서 (a) 단독(F1 최소선)으로 좁히길 원하면 AC-SAV-002(존재 셀)를 제거하는 범위 축소로 반영한다 — 술어 교체는 blocker 보고로 plan을 갱신한 뒤에만 가능하다.
- **D3 — 실패 동작은 폴스루다.** 첫 실패에서 ""를 반환하는 hard-"" 변형은 금지(§3 뮤턴트 4, REQ-SAV-002).
- **D4 — 빈 후보는 무검증 폴스루다.** 빈 문자열은 부재다(REQ-SAV-003). 기존 `TestResolve_EmptySessionIsEmpty`는 수정 없이 GREEN으로 남아야 한다(AC-SAV-005).
- **D5 — 값 재작성 금지.** 검증 통과 후보는 그대로 반환한다. `EvalSymlinks` 정규화 없음(REQ-SAV-006). `os.Stat`이 symlink를 따르는 것으로 dual-spelling 존재 판정은 자연 해결된다.
- **D6 — FromDirectory 불변.** 3단·B4 진입의 git 해석 코드는 무변경이다. 기존 `TestFromDirectory_*` 2건은 수정 없이 GREEN으로 남는다(AC-SAV-006).
- **D7 — 생산 변경 2파일 한정.** `internal/stateanchor/stateanchor.go` + `stateanchor_test.go`. `internal/statusline/**`, `internal/kanban/**`(t536 전용), `internal/cli/**` 무변경. `git diff --stat`으로 M2에서 확인한다.
- **D8 — 기존 테스트 2건은 계약 갱신이다 (계획된 RED-first 변경).** `TestResolve_ProjectDirWins`(`stateanchor_test.go:23-34`)와 `TestResolve_OriginalCwdSecond`(`:38-48`)는 현재 **존재하지 않는 경로**(`/proj`, `/primary`)로 무조건 반환을 고정한다. 새 계약에서는 이 입력이 폴스루되므로 두 테스트는 `t.TempDir()` 실재 디렉터 기반으로 재작성된다 — `TestResolve_ProjectDirWins` → `TestResolve_ProjectDirValidDirWins`(유효 후보 우선 + **워크트리 하위 케이스**: 후보가 common-dir 부모가 아닌 워크트리 디렉터여도 반환 — §3 뮤턴트 3 판별), `TestResolve_OriginalCwdSecond` → 실재 디렉터 기반. 갱신 사실과 이유를 §E.2에 문서화한다 — "수리가 깨뜨린 것"이 아니라 "고정하던 계약이 바뀐 것"이다.
- **D9 — RED-first는 커밋 증거다.** AC-SAV-001/002/004의 신규 테스트는 **구현 이전에 커밋되는 RED**로 시작하고, RED 실패 출력 전문이 progress.md §E.2에 남는다(E8 항목). "오늘의 코드가 stale-경로 페이로드를 그대로 반환한다"가 RED의 이유다(§2 두-셀 규율 — RED 이유 명시).
- **D10 — 로컬 검증은 패키지 단위다.** `go test ./internal/stateanchor/ -count=1` + `go vet ./internal/stateanchor/`. `go test ./...` 로컬 전체 금지 — 전 패키지 판정은 CI 몫이다.
- **D11 — MX 태그 계승.** 기존 `@MX:ANCHOR`(stateanchor.go:61-63)는 유지하고, 검증 로직에 `@MX:SPEC: SPEC-STATE-ANCHOR-VALIDATE-001` 보조선을 추가한다. ANCHOR는 절대 자동 삭제하지 않는다.

## §E 자가 검증

각 마일스톤 종료 시 §C 표 해당 행과 acceptance.md 판정 명령을 재측정하고 **실제 출력을 그대로** progress.md §E.2에 인용한다. 귀속 삼인조(명령 + 관측 출력 + baseline attribution `(this run, this tree)` + HEAD SHA)를 채운다.

## §F 마일스톤

> 순서 구속: 결정이 먼저, 착지가 나중이다. 가장 바뀔 가능성이 큰 결정(검증 술어)은 §D2와 spec.md §5로 plan-phase에서 확정했다. M1 이후에는 결정이 아니라 기계적 착지만 남는다.

### M1 — RED: 검증 테스트 커밋

- AC-SAV-001(상대경로 폴스루)·AC-SAV-002(부재 절대경로 폴스루)·AC-SAV-004(OriginalCwd 동일 처리)의 신규 테스트를 작성한다. `t.TempDir()` 기반 — 실재하는/삭제된 디렉터 fixture.
- 현재 트리에서 실행 → **FAIL (RED)** 관측, 출력 전문 §E.2 보존(E8). RED 이유: 무검증 반환(§B1).
- 폴스루 단언 형태: 후보 실패 시 git 워크업이 빈 CurrentDir이면 ""가 나온다 — 단언은 "실패 후보 값 != 반환값"과 "폴스루가 일어났음(CurrentDir 없음 → "")" 둘 다로 한다. 단, CurrentDir을 실제 git fixture로 주는 하위 케이스는 AC-SAV-004에서 다룬다.

**M1 종료 조건**: 신규 테스트 RED 출력 전문 §E.2 보존. 기존 7테스트 중 6건은 아직 무변경 GREEN.

### M2 — GREEN: 술어 구현 + 계약 갱신 테스트 재작성

- `Resolve`에 §D2 술어(`isAnchorCandidate`: `filepath.IsAbs` && `os.Stat` 성공 && 디렉터 모드)를 적용하고 1·2단을 검증-통과-폴스루로 바꾼다. `PathError`(부재)와 기타 Stat 오류는 동일하게 실패 처리 — 후보 기각이지 프로세스 오류가 아니다.
- §D8의 2건 재작성(`TestResolve_ProjectDirValidDirWins` + 워크트리 하위 케이스, `TestResolve_OriginalCwdSecond` 실재 디렉터판). 갱신 문서화.
- 신규 테스트 GREEN 전환 관측(RED 출력과 동일 명령 재실행).
- MX 보조선 추가(§D11).

**M2 종료 조건**: AC-SAV-001~004 GREEN + 패키지 테스트 전수 GREEN(§D4/D6의 무수정 테스트 포함).

### M3 — 회귀 스윕 + 인계

- §C 전 행 재측정 + AC 매트릭스 전수 판정, 출력 전문 §E.2.
- `go vet ./internal/stateanchor/` 클린.
- `git diff --stat`으로 생산 변경이 `internal/stateanchor/` 2파일로 한정됐음을 확인(§D7, AC-SAV-007의 코드 검사와 함께).
- coverage 측정(`go test -cover ./internal/stateanchor/`) — TRUST 5 Tested 85% 기준 확인.

**M3 종료 조건**: AC 전수 판정 + §E.3 audit-ready 신호.

## §G Anti-Patterns (이 SPEC이 금지하는 것)

- **IsAbs-only 착지** — F1 최소선 그대로 이행하는 것. 존재 검증이 빠지면 부활 기전이 생존한다(§3 뮤턴트 1, AC-SAV-002).
- **git-membership 과잉 검증** — 후보 경로에 git spawn을 추가하거나 워크트리 `project_dir`을 거절하는 것(§3 뮤턴트 3, REQ-SAV-004).
- **hard-"" 조기 반환** — 폴스루 없는 첫 실패 즉시 ""(§3 뮤턴트 4, REQ-SAV-002).
- **값 정규화** — `EvalSymlinks`로 반환값을 재작성하는 것(§D5, REQ-SAV-006).
- **호출자 건드리기** — statusline 어댑터·소비자·표시 유도의 "정리" 접촉(§D7, REQ-SAV-005).
- **kanban 접촉** — §7 관측 대상 2건을 "같이 고치는" 것 — t536 전용 소관이다.
- **전체 스위트 로컬 실행** — `go test ./...` 금지(§D10).

## §H 상호 참조

- 부모 SPEC: `SPEC-STATE-ANCHOR-001` (시접 창설 — REQ-SA-002 체인, REQ-SA-003 skip, D13 스키마 불변을 계승)
- 기원: `.moai/reports/t510/sync-audit.md:120` F1 (develop 워크트리 — 전문 재확인 완료)
- 병렬 카드: t536 (`internal/kanban` root-의미 결함 — 본 SPEC §7 관측, 상호 인용은 판정서 착지 시 갱신, Gaps §8)
- 규칙: `.claude/rules/moai/development/verification-completeness.md` §2 (두-셀 채택 — RED-now 셀에 RED 이유 명시), `.claude/rules/moai/core/verification-claim-integrity.md` §3 (§E 귀속 삼인조)
