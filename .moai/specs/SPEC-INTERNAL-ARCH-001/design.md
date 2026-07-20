---
id: SPEC-INTERNAL-ARCH-001
status: draft
updated: 2026-07-08
---

# SPEC-INTERNAL-ARCH-001 — Design

> 본 문서는 WHAT(spec.md)을 만족하는 구조 설계와 이행 순서를 다룬다. 함수 시그니처·파일명은 **후보안**이며 run-phase에서 코드 실상에 맞춰 조정 가능하다(단, AC의 기계 검증 조건은 불변).

## §A DI seam 설계 (REQ-ARCH-001)

### A.1 현상 구조

```
internal/cli (거대 flat 패키지)
 ├─ deps.go:76  var deps *Dependencies   ← package-global mutable singleton
 ├─ update.go / hook.go / ...            ← 커맨드 파일들이 deps를 클로저 캡처
 └─ agentlint/, specid/, preference/     ← subpackage들: internal/cli를 import하면
                                            cycle (cli가 subpackage를 import해 등록하므로)
```

subpackage가 공유 의존성이 필요하면 (a) 파라미터로 낱개 전달받거나 (b) 회피 주석과 함께 자체 재구현하는 우회를 써 왔다 — 4곳의 cycle 회피 주석이 그 흔적.

### A.2 목표 구조 — leaf seam + 좁은 인터페이스 + constructor injection

```
internal/cli/clideps  (leaf — internal/cli를 import하지 않음)   ← 후보명, 충돌 시 조정
 ├─ 좁은 capability 인터페이스들 (인터페이스 분리 원칙)
 │    예: ConfigProvider / TemplateDeployer / HookRunner / OutputSink
 └─ (필요시) Deps 집합 struct — 배선 편의용

internal/cli
 ├─ root.go: 부팅 시 구체 구현을 조립해 각 subpackage 생성자에 주입
 │    agentlint.NewAgentCmd(d clideps.XxxProvider) *cobra.Command
 └─ var deps 전역 제거 (또는 M1 중간 단계에서 wiring 함수 지역 변수로 격하 후 삭제)

internal/cli/agentlint (pilot)
 └─ clideps만 import — cycle 소멸, 회피 주석 삭제
```

설계 원칙:

1. **좁은 인터페이스 우선**: subpackage마다 필요한 능력만 받는다. god-interface(`Dependencies` 전체 노출)는 후속 cluster 추출 시 다시 결합을 만들므로 금지. seam 패키지에는 "cluster별 요구 능력" 단위의 소형 인터페이스를 둔다.
2. **구체 타입은 internal/cli에 잔류**: `Dependencies` struct 자체와 그 조립(부팅 배선)은 internal/cli(root.go 인근)에 남긴다. seam은 인터페이스와 최소 자료형만 가진다 → seam이 무엇도 import할 필요가 없어 leaf가 보장된다.
3. **위치 후보**: 1안 `internal/cli/clideps`(subpackage — cli 네임스페이스 응집 유지), 2안 `internal/clideps`(top-level — 향후 internal/hook·internal/template과의 cycle까지 흡수 가능). **1안을 기본**으로 하되, run-phase에서 preference↔hook류 교차 cycle 흡수가 필요해지면 2안 승격. AC-ARCH-002a는 어느 위치든 성립하도록 "seam 패키지 deps에 internal/cli 부재"로 정의되어 있다.

### A.3 M1 마이크로 스텝 (green-to-green)

1. seam 패키지 신설 + 인터페이스 정의 (기존 코드 무접촉 — build/test green)
2. `Dependencies`가 seam 인터페이스들을 구현함을 컴파일 타임 검증 (`var _ clideps.ConfigProvider = (*Dependencies)(nil)` 류 assertion 추가)
3. pilot subpackage(agentlint 우선 — SUBPKG-SPLIT M1에서 이미 clean cluster 검증됨)의 생성자 시그니처에 seam 주입 추가, root.go 배선 갱신
4. pilot의 cycle 회피 주석 삭제 (AC-ARCH-002c)
5. 잔여 커맨드 파일들의 `deps` 전역 참조를 배선 지점 경유로 치환 → `var deps` 삭제 (AC-ARCH-002b). 치환 규모가 크면 파일 그룹 단위 커밋 분할.

## §B monolith 분할 concern map (REQ-ARCH-002)

동일 패키지 내 **파일 단위** 분할 — 심볼 이동은 export 변화가 없어 행위 중립. 후보 파일명:

| 현재 소재 | 관심사 | 분할 후보 파일 |
|-----------|--------|----------------|
| update.go | 커맨드 정의 + 오케스트레이션 | update.go (잔류, ≤800줄) |
| update.go | 바이너리 self-update | update_binary.go |
| update.go L1637-2177 | config 3-way merge + backup | update_merge.go |
| update.go | archive-drift 처리 | update_archive.go |
| update.go | namespace 보호 | update_namespace.go |
| update.go ~L2199-2400 | SEC-HARDEN-003/004/005 path-guard (restoreTargetContained / parentChainContained) | update_pathguard.go (전용 — AC-ARCH-003c) |
| hook.go | 이벤트 dispatcher | hook.go (잔류, ≤500줄) |
| hook.go | CLI 서브커맨드 등록 | hook_commands.go |
| hook.go | harness-classify | hook_classify.go |
| hook.go | DB-sync 경로 구성 | hook_dbsync.go |

주의: L 앵커는 저작 시점 근사치 — run-phase에서 심볼 기준으로 재확정한다(line-number drift 방지, content-token 앵커 우선). path-guard 가족은 보안 회귀 표면이므로 characterization(경로 탈출/symlink 시나리오 픽스처) 선행 필수.

## §C internal/core 해체 매핑 (REQ-ARCH-003)

| 현재 | LOC(감사) | fan-in | 이동 후보 | 비고 |
|------|-----------|--------|-----------|------|
| internal/core/git | 1,305 | 5 files | `internal/gitops` | `internal/git` 단독명은 도메인 혼동 여지 — run-phase에서 `ls internal/` 충돌 pre-check 후 확정 |
| internal/core/project | 2,299 | 3 files | `internal/project` | 충돌 pre-check 필수 |
| internal/core/quality | 1,350 | 2 files | `internal/quality` | 충돌 pre-check 필수 |
| internal/core/integration | 0 (.gitkeep만) | 0 | **삭제** | 죽은 스텁 |
| internal/core/migration | 0 (.gitkeep만) | 0 | **삭제** | 죽은 스텁 |

이행 규칙: (1) 스텁 삭제 1커밋 선행, (2) subpackage당 1커밋 — `git mv` + 동일 커밋 내 import call-site 전량 갱신 + package clause 갱신, (3) 각 커밋 단독 green. fan-in이 disjoint하므로 3커밋은 상호 순서 무관.

기각 대안: (a) internal/core에 doc.go만 추가해 grouping을 문서화 — bare namespace 자체가 잔존해 finding 미해소, (b) 스텁을 채워 populate — 신규 기능 scope라 Out of Scope.

## §D config pipeline 단일화 결정 구조 (REQ-ARCH-004)

**Primary — resolver 은퇴**:

1. characterization: 현재 `moai doctor` config 진단 출력을 fixture로 고정 (단, env-override **불일치 결함 자체는 보존 대상이 아님** — acceptance.md 시나리오 2가 목표 행위를 정의; "doctor 표시 항목/형식"만 보존)
2. `doctor_config.go`를 ConfigManager 조회 기반으로 재구축. flat key-path 표시가 필요하면 `flattenStruct`/`flattenStructInto`(현 resolver.go:562-626)를 doctor 측 helper로 이전(로직 보존 이동)
3. `config.NewResolver` 소비처 소멸 확인 (grep) → resolver.go + 전속 테스트 삭제
4. 주의: doctor의 진단 값이 env override를 반영하게 되는 것은 **결함 수정**이며, 이는 REQ-ARCH-004가 명시적으로 요구하는 정합화다 — "행위 보존" 불변식의 예외로 spec.md §B에 이미 계약됨(diagnostics가 runtime 의미를 따르게 됨)

**Decision gate → fallback**: run-phase에서 resolver의 8-tier 의미(SrcPolicy/Plugin/Skill/Session 등)가 doctor 이외의 예정된 소비처를 가진 증거(로드맵/코멘트/테스트 계약)가 발견되면 은퇴를 중단하고 reconcile 경로로 전환: resolver에 4종 env-var override를 추가 + 두 pipeline의 역할 차이를 CLAUDE.md에 문서화. AC-ARCH-005b의 대체 충족 조항이 이 경로를 커버한다.

## §E table-driven loader 설계 (REQ-ARCH-005)

```go
type sectionSpec struct {
    name     string                     // "user", "language", ...
    filename string                     // "user.yaml", ...
    target   func(*Config) any          // 로드 대상 포인터 반환
}

var sectionRegistry = []sectionSpec{ /* 13 entries */ }

// Load: registry 순회 → loadYAMLFile(dir, s.filename, s.target(cfg))
//       → loadedSections[s.name] = ok
// Save: 동일 registry 순회 → saveSection(dir, s.filename, s.target(cfg))
// getSectionLocked/setSectionLocked: registry 기반 map lookup으로 switch 대체
```

행위 보존 포인트: (1) 로드 순서 유지(registry 순서 = 기존 호출 순서), (2) 부분 실패 시 관용 정책(존재하지 않는 파일 skip, 손상 YAML 처리) 기존과 동일 — characterization fixture에 손상 YAML 케이스 포함(acceptance.md E4), (3) `detectSchemaDrift` 경유 여부 등 per-section 특이 분기가 있으면 spec 필드로 표현(범용 루프에 예외 하드코딩 금지).

## §F 이행 순서 + rollback 전략

```
M0 gate ── M6(문서, 선착지 가능)
   │
   ▼
M1 seam (5 마이크로 스텝, §A.3) ── CHECKPOINT-1 ──▶ M2 분할 (concern map §B)
   │
   ├─ M3 core 해체 (스텁삭제 → 3× 이동커밋)     [M1과 독립 — 병행 가능하나 동일
   └─ M4 resolver 은퇴 (§D) → M5 table-driven (§E)  세션 순차 착지 권장]
```

- **rollback 단위 = 커밋**: 모든 커밋이 단독 green이므로 `git revert <sha>` 1회로 임의 스텝 철회 가능. milestone 간 커밋 혼합 금지가 이 속성의 전제.
- **suite red 발견 시**: 즉시 revert (REQ-ARCH-007 revert-on-red) → 실패 접근을 progress.md §E.2에 기록 → 더 작은 스텝으로 재시도.
- **병렬 세션 레이스**: 격리 worktree 사용 시 landing 직전 `git fetch` + rebase 재확인 (worktree가 main 전진을 가리는 함정 방지).

## §G 후속 SPEC 연결점 (본 SPEC 범위 밖, 설계상 인지)

- update/hook cluster의 **패키지 추출**: M1 seam + M2 파일 분할이 완료되면 SUBPKG-SPLIT 후속 SPEC의 tri-axis coupling 중 deps-축이 해소된 상태가 된다.
- preference↔hook, cli↔template 교차 cycle: seam 2안(top-level) 승격 또는 별도 인터페이스 역전 SPEC.
- `MOAI_USER_NAME`/`MOAI_CONVERSATION_LANG` env override 구현 여부 결정.
