# plan.md — SPEC-PRECOMMIT-GATE-SCOPE-001

id: SPEC-PRECOMMIT-GATE-SCOPE-001
created: 2026-09-03
tier: M

## A. Context

- 카드: t461 (Class C — 설계 변경, Tier M). 브랜치 `WT-precommit-gate-scope`, base `a239cf050`.
- 결함: pre-commit heavy gate(`moai gate`)가 프로젝트 전역 vet/lint/test/typecheck를 커밋 허용
  조건으로 삼아, 무관한 기존 실패 1건이 모든 커밋을 막는다. 진입점은 `883d53852`
  (SPEC-PRETOOL-GATE-MOVE-001, 2026-07-28) — fast-subset(`52b5e4bf5`, SPEC-PRECOMMIT-001)에
  heavy gate 블록을 얹으면서 범위가 커밋 단위에서 프로젝트 전역으로 확대됐다.
- 요구사항·좌표: spec.md §A/§B 참조. 모든 좌표는 2026-09-03 본 워크트리에서 심볼 기준 재검증 완료.

## B. Known Issues

1. **remedy 표면 자체가 소멸 위험에 있다.** 안내할 remedy 키(`gate.enabled` / `gate.skip_tests` /
   `gate.disabled_steps`)는 `.moai/config/sections/gate.yaml`에 있는데, 이 디렉터리는
   `CleanMoaiManagedPaths`(`internal/cli/update/deploy/deploy.go`)가 `moai update`마다 통째 삭제 후
   템플릿판(`enabled: true`, `skip_tests: false`)으로 재배포한다. 사용자가 gate.yaml을 고쳐 opt-out해도
   다음 update에서 원상복구된다 — REQ-006이 이를 직접 다룬다.
2. **`QualityGate.Run`의 Enabled 스위치는 단독 CLI와 공유된다.**
   `internal/hook/quality/gate.go` `QualityGate.Run`은 `config.Enabled == false`이면 즉시 통과
   반환한다. 템플릿 gate.yaml의 `gate.enabled`를 false로 뒤집으면 단독 `moai gate` 실행까지 꺼진다 —
   축 (b)의 메커니즘은 이 공유 스위치를 겨냥해서는 안 된다(§C 제약).
3. **t237 충돌.** t237 / issue #1641이 같은 twin 파일(`preCommitHookContent` 상수 +
   `.git_hooks/pre-commit`)을 편집하는 go vet 모듈해석 수리이며, 검증된 패치를 가진 채 open 상태다
   (`.moai/specs/SPEC-PRECOMMIT-PRESERVE-001/plan.md`에도 동일 충돌 기록 존재). 병합 순서 충돌 가능성은
   progress.md와 아래 §C에 기록한다.

## C. 설계 결함 축과 권고 (Decision Axes)

### 축 (a) — heavy gate를 staged 범위로 좁힌다 (커밋 단위 계약 복원)

- 장점: 원래 fast-subset의 계약으로 복귀. 훅이 항상 켜져 "커밋은 최소한 스스로 건강하다"를 보장.
- 단점: 16개 프로그래밍 언어에서 "staged 파일만 검증"하는 균일한 모드가 존재하지 않는다. pytest /
  ruff / go test 등 도구별로 staged-파일 필터를 각각 구현해야 하고, 패키지 의존성 때문에 staged 파일만
  테스트하는 것이 이론적으로도 불완전하다(`go test ./affected/...` 정도가 상한). 구현 비용이 크고
  결함 표면이 넓어진다.
- 추가 비용: `moai gate`에 새 scope 모드(예: `--staged`)가 필요하고, 이는 gate 러너
  (`internal/hook/quality/gate.go`) 전반의 인터페이스 변경이다.

### 축 (b) — pre-commit 맥락의 heavy gate를 기본 OFF로, gate.yaml opt-in (권장)

- 장점: 이 코드베이스의 확립된 패턴과 정확히 일치한다 — `BranchGuard.Enabled`(기본 false,
  `main-checkout-branch-guard.md` § Mechanical Enforcement 명시), `agent_stop_guard`(기본 false),
  ast-grep 차단 모드(opt-in `block_on_error`). 기본값 1회 변경 + 안내 문구 추가로 구현이 좁다.
- 위험(반드시 설계에서 처리): `QualityGate.Run`의 `Enabled` 스위치가 단독 `moai gate`와 공유되므로
  (Known Issue 2), opt-in 표면은 pre-commit 호출 맥락만 겨냥해야 한다. 후보 메커니즘 3가지 —
  run-phase에서 하나를 확정한다:
  1. 훅이 `MOAI_PRECOMMIT=1` 환경마커와 함께 `moai gate`를 호출하고, `moai gate`는 이 마커 하에서만
     새 키(예: `gate.pre_commit.enabled`)를 존중한다. gate 러너에 맥락 분기가 하나 늘어난다.
  2. 새 `GateConfig` 키(예: `gate.pre_commit_heavy`)만 두고, 훅은 이 키가 true일 때만 `moai gate`를
     실행한다. 훅이 shell이므로 gate.yaml을 직접 파싱하는 대신 `moai gate --check pre-commit`류의
     질의 서브커맨드가 필요할 수 있다.
  3. 기존 `disabled_steps`의 반전 규약(FALSE=skip, issue #667 Fix 3)을 재활용해 pre-commit 맥락
     기본 비활성 목록을 코드 기본값으로 주는 방식 — 새 키 없이 defaults.go만 변경. 단, 단독 실행과의
     구분이 없어 Known Issue 2를 해소하지 못해 기각 경향.
- 권고 근거: 결함의 본질은 "사용자가 동의하지 않은 비용을 매 커밋에 부과"하는 것이고, 축 (b)는 최소
  변경으로 이를 제거하면서 opt-in 사용자의 전역 검증 능력(spec.md Out of Scope에서 보존 선언)을
  유지한다. 축 (a)의 언어별 staged-필터 구현은 이 카드 범위를 넘는다.

### 축 (c) — 실패 메시지 remedy 안내 (무조건, 어느 축이든 시행)

- 훅 실패 메시지에 `.moai/config/sections/gate.yaml` 경로와 `gate.enabled`, `gate.skip_tests`,
  `gate.disabled_steps` 세 키를 명시한다. 기존 `SKIP_MOAI_PRECOMMIT=1` 안내는 유지한다.
- twin 동시 편집 대상이므로 M2에서 같은 커밋으로 들어간다.

### [NEEDS CLARIFICATION: 축 (a) vs (b) 최종 확정]

run-phase 진입 전 Implementation Kickoff Approval에서 운영자가 확정한다. 본 plan은 (b)를 권고하며,
(c)는 어느 쪽이든 시행된다. 확정 전까지 AC-002/AC-006의 통과 조건은 axis-conditional로 남는다.
운영자가 (b)를 확정하면 위 후보 메커니즘 1~3 중 하나도 같은 게이트에서 확정한다.

## D. Constraints

1. **Twin 동시 편집 [HARD]** — `preCommitHookContent`(`internal/cli/hook_install_precommit.go`)와
   `internal/template/templates/.git_hooks/pre-commit`은 반드시 같은 커밋에서 함께 편집한다.
   `TestPreCommitTemplateMatchesConstant`(internal/cli/hook_install_precommit_test.go)가 위반을 잡는다.
2. **t237 무변경 [HARD]** — fast-subset go vet 단계(staged 파일 수집 + `go vet $BT_TAGS $PKGS`)의
   의미를 변경하지 않는다(REQ-008, AC-005).
3. **gate.yaml 생존 설계 [HARD]** — `.moai/config/`가 `CleanMoaiManagedPaths` 소거 뿌리에 포함된
   현실(Known Issue 1) 위에서 REQ-006을 충족하는 메커니즘을 설계한다. 후보: (i) gate.yaml을 소거
   제외·병합 대상으로 전환, (ii) opt-in 표면을 소거 뿌리 밖(예: settings.local.json류 런티임 표면)에
   두기, (iii) 기본값 자체를 OFF로 바꿔 opt-out 편집이 필요 없게 하되 opt-in 편집의 생존은 별도 처리.
   전례: `git-strategy.yaml`은 템플릿 미러를 하지 않고 update마다 수동 재적용하는 **인지된 한계**로
   운영 중 — 이 SPEC은 같은 상태를 만들지 않는다(REQ-006이 기계 판정 AC를 갖는 이유).
4. **검증 부하 [HARD]** — 로컬 전체 스위트 금지. 대상 패키지만: `go test ./internal/cli/...`,
   `./internal/config/...`, `./internal/hook/...`. 템플릿 편집 후 `make build` 필수.
5. **템플릿 중립성 [HARD]** — 훅 템플릿 본문에 SPEC ID, 내부 날짜, 커밋 SHA를 넣지 않는다
   (`.moai/docs/template-internal-isolation-doctrine.md` §25.1; CI 가드
   `.github/workflows/template-neutrality-check.yaml`).
6. **단독 `moai gate` 보존 [HARD]** — REQ-004 메커니즘은 단독 실행을 비활성화하지 않는다(Known Issue 2).

## E. Self-Verification

- M1 종료 시: `go test ./internal/config/... -count=1` 통과 출력 인용.
- M2 종료 시: `TestPreCommitTemplateMatchesConstant` 통과 출력 인용 + `go test ./internal/cli/... -count=1`.
- M3 종료 시: AC-002 재현 테스트의 RED→GREEN 출력 인용.
- 최종: `make build` 성공 + 템플릿 중립성 자가 점검 5항목(§25.3) + spec.md checklist.

## F. Milestones (의사결정 역전가능성 순 — 변경 가능성 높은 결정을 앞에)

### M1 — gate.yaml opt-in 메커니즘 + update 생존 (Priority High, data-model 결정)

-REQ-004/REQ-006의 메커니즘 확정(운영자 게이트에서)과 구현. `internal/config/types.go` GateConfig,
`internal/config/defaults.go`, `internal/config/loader_gate.go`, 템플릿
`internal/template/templates/.moai/config/sections/gate.yaml`, (메커니즘에 따라)
`internal/cli/gate.go` / `internal/cli/update/deploy/deploy.go`. 새 키 추가 시
`audit_struct_yaml_symmetry_test.go` symmetryCases 동기화(settings-management.md 5-step procedure).
이 마일스톤이 스키마와 배포 계약을 바꾸는 가장 역전하기 어려운 결정이므로 최전단.

### M2 — twin 편집: 훅 heavy gate 블록 + 실패 메시지 (Priority High, user-facing 결정)

- `preCommitHookContent` 상수와 `.git_hooks/pre-commit` 템플릿을 **같은 커밋에서** 편집:
  축 (b) 확정 시 heavy gate 블록을 opt-in 조건부로 전환, 축 (a) 확정 시 staged 범위 호출로 전환.
  어느 쪽이든 REQ-003의 remedy 안내 4요소를 실패 메시지에 추가.
- `make build` 재생성. `TestPreCommitTemplateMatchesConstant` 통과 확인.

### M3 — 사고 재현 테스트 (Priority Medium)

- AC-002의 형태를 테스트로 고정: 기존 실패 1건 + 무관한 staged 변경 → 기본 동작으로 커밋 성공.
  RED(현재 결함 재현) → GREEN(M1+M2 적용 후). TDD 사이클 진입점.

### M4 — 문서·안내 (Priority Medium)

- CHANGELOG 항목. gate.yaml opt-in 방법을 사용자 문서 표면(README 또는 docs-site 해당 절)에 반영 —
  문서 언어는 4-locale 동기화 규칙 준수. 범위 최소화: 이 SPEC이 도입하는 키/동작만.

### M5 — 기계적 마무리 (Priority Low)

- 카탈로그/해시 재계산(`make build` 체인), go vet ./internal/... 영향 패키지, 커밋 정리.
  t237과의 병합 순서 메모를 progress.md에 최종 기록.

## G. Anti-Patterns

- 훅 셸 스크립트에서 gate.yaml을 grep/ad-hoc 파싱하지 않는다 — 설정 판독은 Go 로더 경로로.
- 템플릿과 상수를 "나중에 맞추겠다"며 따로 커밋하지 않는다 — byte identity 테스트가 즉시 적색이 된다.
- `gate.enabled` 기본값만 뒤집고 끝내지 않는다 — 단독 `moai gate`까지 꺼지는 회귀(Known Issue 2).
- 검증을 `go test ./...`로 넓히지 않는다 — 제약 4.

## H. Cross-References

- spec.md — 요구사항/좌표 SSOT
- acceptance.md — AC 매트릭스(AC-001~AC-008)
- 진입 커밋: `52b5e4bf5`, `883d53852` (git show로 재확인 가능)
- t237 / issue #1641 — twin 파일 경합 카드, 본 SPEC 소관 아님
- 전례 SPEC: SPEC-PRECOMMIT-001, SPEC-PRETOOL-GATE-MOVE-001, SPEC-PRECOMMIT-PRESERVE-001
