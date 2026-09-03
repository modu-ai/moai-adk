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
  (Known Issue 2), opt-in 표면은 pre-commit 호출 맥락만 겨냥해야 한다. 후보 메커니즘 3가지를 검토했고
  **메커니즘 1이 확정됐다**(§F 상단의 운영자 결정): 훅이 `MOAI_PRECOMMIT=1` 환경마커와 함께
  `moai gate`를 호출하고, 러너는 이 마커 하에서만 새 키 `gate.pre_commit.enabled`를 존중한다.
  단독 `moai gate` 동작은 불변이며 새 서브커맨드는 없고 러너 분기점은 1개다. (기각: 후보 2는
  shell 훅이 gate.yaml을 파싱해야 하는 규약 위반, 후보 3은 단독 실행과의 구분이 없어
  Known Issue 2를 해소하지 못함.)
- 권고 근거: 결함의 본질은 "사용자가 동의하지 않은 비용을 매 커밋에 부과"하는 것이고, 축 (b)는 최소
  변경으로 이를 제거하면서 opt-in 사용자의 전역 검증 능력(spec.md Out of Scope에서 보존 선언)을
  유지한다. 축 (a)의 언어별 staged-필터 구현은 이 카드 범위를 넘는다.

### 축 (c) — 실패 메시지 remedy 안내 (무조건, 어느 축이든 시행)

- 훅 실패 메시지에 `.moai/config/sections/gate.yaml` 경로와 `gate.enabled`, `gate.skip_tests`,
  `gate.disabled_steps` 세 키를 명시한다. 기존 `SKIP_MOAI_PRECOMMIT=1` 안내는 유지한다.
- twin 동시 편집 대상이므로 M2에서 같은 커밋으로 들어간다.

### 확정된 설계 (Implementation Kickoff Approval — 운영자 결정, 2026-09-03)

Implementation Kickoff Approval이 통과했고 운영자가 다음 3건을 기록했다. 이 절에 남아 있던
미확정 표시(축 선택)는 아래 기록으로 해소·제거됐다.

1. **축 (b)** — pre-commit 맥락의 heavy gate 기본 OFF, gate.yaml opt-in.
2. **메커니즘 1** — 훅이 `MOAI_PRECOMMIT=1` 환경 마커와 함께 `moai gate`를 호출하고, 게이트 러너는
   **새 키 `gate.pre_commit.enabled`를 오직 이 마커 하에서만** 존중한다. 단독 `moai gate` 동작은
   불변. 새 서브커맨드 없음. 러너 분기점 1개.
3. **신설 요구 (운영자 지시)** — `gate.pre_commit.enabled`는 `moai web` 설정 화면에서 편집 가능해야
   한다 (REQ-009). 편집 저장 경로는 정확히 `.moai/config/sections/gate.yaml`이다.

AC-002/AC-006의 axis-conditional 표기는 해제된다: 통과 조건은 축 (b) + 메커니즘 1로 확정.

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
6. **단독 `moai gate` 보존 [HARD]** — `gate.pre_commit.enabled`는 `MOAI_PRECOMMIT=1` 마커 하에서만
   존중된다(운영자 결정 2). 마커 없는 단독 실행은 이 키를 읽지 않고 기존 `gate.enabled` 계약 그대로다.
7. **web 저장 경로 [HARD]** — REQ-009의 저장은 `WriteSectionViaSeam` 경로로 정확히
   `.moai/config/sections/gate.yaml`에 기록한다(yamlpatch — 주석·미모델링 키 보존). 다른 섹션 파일로
   흘리지 않는다.
8. **명명 중복 [HARD]** — `git_strategy.<mode>.hooks.pre_commit`(문자열,
   `internal/config/validation.go` `checkStringField`)은 본 SPEC의 `gate.pre_commit.enabled`와
   무관한 다른 subtree다. 구현·테스트·문서 어디서도 두 키를 혼용하지 않는다.

## E. Self-Verification

- M1 종료 시: `go test ./internal/config/... -count=1` 통과 출력 인용.
- M2 종료 시: `go test ./internal/settings/... ./internal/web/... -count=1` 통과 출력 인용 +
  web 저장이 gate.yaml에 기록함을 확인하는 테스트 출력.
- M3 종료 시: `TestPreCommitTemplateMatchesConstant` 통과 출력 인용 + `go test ./internal/cli/... -count=1`.
- M4 종료 시: AC-002 재현 테스트의 RED→GREEN 출력 인용.
- 최종: `make build` 성공 + 템플릿 중립성 자가 점검 5항목(§25.3) + spec.md checklist.

## F. Milestones (의사결정 역전가능성 순 — 변경 가능성 높은 결정을 앞에)

> 축 (b) + 메커니즘 1은 §C에서 확정됐다. 이하 마일스톤은 확정 설계 기준이다.

### M1 — `gate.pre_commit.enabled` 스키마 + 러너 분기 + update 생존 (Priority High, data-model 결정)

- `internal/config/types.go` `GateConfig`에 `PreCommit GatePreCommitConfig`(yaml:
  `pre_commit`, 내부 `enabled`) 추가. `internal/config/defaults.go` 기본값 `enabled: false`
  (기본 OFF가 이 SPEC의 요구다). 템플릿 `internal/template/templates/.moai/config/sections/gate.yaml`
  에 `pre_commit:` 섹션과 안내 주석 추가.
- 게이트 러너 분기점 1개: `internal/cli/gate.go` `runGate`가 `MOAI_PRECOMMIT=1` 마커를 감지하면
  `pre_commit.enabled`가 거짓일 때 heavy 단계를 건너뛰고 통과 (AC-009). 마커가 없으면(단독 실행)
  이 키를 읽지 않는다 — 단독 동작 불변(운영자 결정 2).
- REQ-006 메커니즘: `CleanMoaiManagedPaths`(`internal/cli/update/deploy/deploy.go`)가
  gate.yaml을 소거·재배포하기 전에 사용자 값(`pre_commit.enabled` 포함 전 키)을 보존·병합하거나,
  gate.yaml을 소거 제외 대상으로 전환한다. 어느 쪽이든 AC-004가 기계 판정한다.
- 새 키 추가 시 `audit_struct_yaml_symmetry_test.go` symmetryCases 동기화
  (settings-management.md 5-step procedure).

### M2 — `moai web` 설정 표면 (Priority High, 운영자 지시 REQ-009)

- 조사 결과(2026-09-03 실측): settings 스키마에는 `gate` 섹션이 **존재하지 않는다**
  (`internal/settings/schema.go` `SectionID` 목록에 gate 없음). 따라서:
  1. `internal/settings/schema.go`에 `SectionGate SectionID = "gate"` 신설 +
     `internal/settings/schema_sections.go`에 `gate.pre_commit.enabled` FieldDef 등록 —
     `Persist: PersistTarget{Kind: PersistSeam, Section: "gate", Path: "pre_commit.enabled"}`.
  2. 렌더 배선: `internal/web/schemaform.go` `schemaSectionMetas()`에 gate 패널 항목 추가
     (신규 탭 또는 기존 탭 배치 — `isWorktreeFieldName`류 이름 기반 배치 선례 준수,
     영속화 경로는 그대로). `fieldsetSchemaSection`(fieldsets.templ)이 스키마 주도 렌더.
  3. 폼 컨트롤 네이밍 규약: `name="gate.pre_commit.enabled"` + bool hidden companion
     `gate.pre_commit.enabled__present` (unchecked↔미제출 구분, `parseSchemaForm` EC-1 계약).
  4. 라벨: `internal/web/assets/i18n.js`에 `sec.gate.title` 등 4-locale 등록
     (`sec.workflow.title` 선례 — en/ko 외 ja/zh 확인).
  5. 저장 경로 확인: `ApplySchemaEdits` → `WriteSectionViaSeam`
     (`internal/settings/sectionwrite.go`)가 `.moai/config/sections/<section>.yaml`에
     yamlpatch 기록(주석·미모델링 키 보존) — gate.yaml에 정확히 기록됨(실측 완료).
  6. `schemaEditableField`는 `PersistSeam`/`PersistTypedSection`만 편집 대상으로 삼으므로
     위 FieldDef는 그 자격을 갖는다.
- `workflow_agents` 은닉 선례 참고: 그 은닉은 스키마 레지스트리 수준의 부재
  (`internal/web/agent_settings_test.go:97` — 폼 컨트롤 미렌더, :230
  `TestWorkflowAgentsWebSubmissionIgnored` — 제출 무시)다. 본 키는 정반대(노출)이므로
  레지스트리 등록이 곧 노출이며, 별도 hide/expose 플래그는 필요 없다.
- 템플릿 편집 시 `make build`(fieldsets_templ.go 재생성 포함).

### M3 — twin 편집: 훅 heavy gate 블록 + 실패 메시지 (Priority High, user-facing 결정)

- `preCommitHookContent` 상수와 `.git_hooks/pre-commit` 템플릿을 **같은 커밋에서** 편집:
  heavy gate 블록을 `MOAI_PRECOMMIT=1 moai gate` 호출로 전환(마커가 opt-in 판정의 트리거).
  REQ-003의 remedy 안내 5요소(`.moai/config/sections/gate.yaml` 경로 +
  `gate.pre_commit.enabled` / `gate.enabled` / `gate.skip_tests` / `gate.disabled_steps`)를
  실패 메시지에 추가. `SKIP_MOAI_PRECOMMIT=1` 안내는 유지.
- `make build` 재생성. `TestPreCommitTemplateMatchesConstant` 통과 확인.

### M4 — 사고 재현 테스트 (Priority Medium)

- AC-002의 형태를 테스트로 고정: 기존 실패 1건 + 무관한 staged 변경 → 기본 동작(opt-in 없음)으로
  커밋 성공. RED(현재 결함 재현) → GREEN(M1+M3 적용 후). TDD 사이클 진입점.
- AC-004 생존 테스트: 손편집값과 M2 web 작성값 모두 `moai update` 후 유지됨을 fixture로 확인
  (하나의 REQ-006/AC-004로 흡수 — 중복 요구 아님).

### M5 — 문서·안내 (Priority Medium)

- CHANGELOG 항목. gate.yaml opt-in 방법(`moai web` 화면 경로 포함)을 사용자 문서 표면에 반영 —
  문서 언어는 4-locale 동기화 규칙 준수. 범위 최소화: 이 SPEC이 도입하는 키/동작만.

### M6 — 기계적 마무리 (Priority Low)

- 카탈로그/해시 재계산(`make build` 체인), go vet ./internal/... 영향 패키지, 커밋 정리.
  t237과의 병합 순서 메모를 progress.md에 최종 기록.

## G. Anti-Patterns

- 훅 셸 스크립트에서 gate.yaml을 grep/ad-hoc 파싱하지 않는다 — 설정 판독은 Go 로더 경로로.
- 템플릿과 상수를 "나중에 맞추겠다"며 따로 커밋하지 않는다 — byte identity 테스트가 즉시 적색이 된다.
- `gate.enabled` 기본값만 뒤집고 끝내지 않는다 — 단독 `moai gate`까지 꺼지는 회귀(Known Issue 2).
- 검증을 `go test ./...`로 넓히지 않는다 — 제약 4.

## H. Cross-References

- spec.md — 요구사항/좌표 SSOT
- acceptance.md — AC 매트릭스(AC-001~AC-010)
- 진입 커밋: `52b5e4bf5`, `883d53852` (git show로 재확인 가능)
- t237 / issue #1641 — twin 파일 경합 카드, 본 SPEC 소관 아님
- 전례 SPEC: SPEC-PRECOMMIT-001, SPEC-PRETOOL-GATE-MOVE-001, SPEC-PRECOMMIT-PRESERVE-001
- web 설정 표면 아키텍처 선례: `internal/settings/schema_sections.go`(레지스트리),
  `internal/web/schemaform.go` `schemaSectionMetas`(패널 배선),
  `internal/settings/sectionwrite.go` `WriteSectionViaSeam`(seam 저장)
