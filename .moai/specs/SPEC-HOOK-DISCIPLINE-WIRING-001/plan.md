# Implementation Plan — SPEC-HOOK-DISCIPLINE-WIRING-001

> Discipline hook wiring (Phase-2 realization). Tier M. cycle_type per quality.yaml.
> Status: draft. This plan derives from spec.md §D requirements.

## A. Context

세 discipline 훅 중 둘(`status-transition-ownership.sh`, `sync-phase-quality-gate.sh`)을 settings.json.tmpl에 advisory/warn-first로 wiring한다. 핵심 엔지니어링은 sync-gate의 Go-하드코딩을 16개 언어 중립 자동 감지로 일반화하는 작업이다. `team-ac-verify.sh`는 의도적으로 제외(파일 보존, 미등록)한다.

본 작업은 **SPEC 본문 작성이 아니라 구현(run-phase)** 계획이다. 실제 편집 대상:
- `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` (언어 일반화)
- `internal/template/templates/.claude/settings.json.tmpl` (2개 엔트리 추가)
- `.claude/hooks/moai/sync-phase-quality-gate.sh` (local mirror, make build 후 동기화)
- `.claude/settings.json` (local git-tracked, 2개 엔트리 추가 — dev-intent 키 미접촉)

## B. Known Issues / Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| sync-gate 언어 일반화가 중립성 CI guard를 깨뜨림 (Go 키워드 잔존) | High (central technical risk) | M1에서 언어별 toolchain을 변수/케이스로 분기, Go 키워드는 Go 케이스 안으로만 한정. M3에서 neutrality guard 실행 검증 |
| settings.json.tmpl 렌더링 깨짐 (platform-conditional 패턴 오류) | Med | 기존 wrapper 엔트리 패턴을 verbatim 미러, `moai init` 렌더 smoke test |
| local↔template 불일치 (make build 누락) | Med | M2 말미에 make build + diff 검증 |
| dev-intent 키(defaultMode/env.PATH/teammateMode) 교란 | Med | local settings.json은 hook 배열 엔트리만 ADD, 다른 키 미접촉 — git diff로 검증 |
| advisory 훅이 기존 핸들러와 충돌 | Low | exit-0 advisory 훅은 공존 가능 (배열 추가). coexistence AC로 검증 |
| sync-gate가 비-Go 프로젝트에서 도구 부재로 오류 종료 | Med | graceful skip(command -v 가드) 구현, 비-Go fixture로 검증 |

## C. Pre-flight Checks

- [ ] 세 훅 파일이 local + template mirror 양쪽에 존재 (확인됨: 3625/4379/3066 bytes)
- [ ] settings.json.tmpl이 PostToolUse/Stop/TaskCompleted 블록 보유 (확인됨)
- [ ] neutrality guard 테스트가 현재 GREEN인지 baseline 측정 (`go test ./internal/template/ -run 'TestTemplateNeutralityAudit|InternalContentLeak'`)
- [ ] CLAUDE.md §7 언어 매트릭스 재확인 (Go/Node/Python/Rust 4종)

## D. Constraints

- [HARD] SPEC 본문 외 코드/settings/hook 편집은 run-phase(manager-develop)에서 수행. 본 plan은 그 작업 분해.
- [HARD] 16개 언어 중립성이 central technical risk. sync-gate 일반화 후 neutrality guard 통과 필수.
- [HARD] `team-ac-verify.sh` 파일 삭제 금지 — 미등록만.
- [HARD] local `.claude/settings.json`의 dev-intent 키(defaultMode/env.PATH/teammateMode) 미접촉.
- [HARD] Template-First: template source 먼저 편집 → `make build`(EMBEDDED 재생성 only) → local `.claude/` 트리는 **수동 미러(manual copy)** (CLAUDE.local.md §2). **D5 주의**: `make build`는 `go:embed → embedded.go`만 갱신하며 local `.claude/` 작업 트리는 건드리지 않는다. local sync는 `moai update` 또는 수동 복사. byte-identity(AC-HDW-007)는 manual mirror로 달성한다.

## E. Self-Verification

run-phase 완료 시 §F의 각 마일스톤 exit 기준 + acceptance.md 9개 AC(AC-HDW-001..009)가 모두 PASS여야 한다. 검증은 read-only batch(go test, grep, bash -n, moai init smoke)로 병렬 수행.

## F. Milestones

### M1 — sync-gate 언어 일반화 (the real engineering)

**대상**: `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh`

작업:
1. 프로젝트 언어 감지 함수 추가: 마커 우선순위 `go.mod`(Go) / `package.json`(Node.js) / `pyproject.toml` OR `requirements.txt`(Python) / `Cargo.toml`(Rust). **[HARD — D1 testability] `detect_language()`는 직접 호출 가능(source-able)한 셸 함수로 구현** — sync-phase git 게이트를 거치지 않고 `source <hook> && detect_language "$dir"`로 단위 검증 가능해야 하며, 부수효과 없이 언어 문자열만 stdout 반환(AC-HDW-002(a), AC-HDW-009의 case-block 추출 전제).
2. 언어별 toolchain case 분기 (CLAUDE.md §7 매트릭스):
   - Go: `go vet` → `golangci-lint`(opt) → `go test`
   - Node.js: `eslint`(opt) → `npm test`
   - Python: `ruff`(opt) → `pytest`
   - Rust: `cargo clippy`(opt) → `cargo test`
3. 각 도구 호출 전 `command -v <tool>` 가드 → 부재 시 graceful skip(exit 0, 해당 step만 skip, 로그에 skipped 기록).
4. 언어 마커 0개 → silent pass(exit 0, toolchain 미실행).
5. 기존 sync-phase 커밋 감지 / `--skip-hook` / Go-delta skip 로직은 일반화된 형태로 유지(언어별 delta 감지).
6. coverage 측정은 advisory로 유지(언어별 가능 시), regression 차단 미구현.

**Exit 기준**: bash -n 통과; `detect_language()` 직접 호출 가능; **실 git-repo**(sync-phase HEAD 커밋) Go fixture에서 Go toolchain 실행; 실 git-repo 비-Go fixture(package.json만)에서 Go toolchain skip + exit 0; 마커 없는 fixture에서 silent pass exit 0. **Go-tool 토큰이 Go case 블록 밖에 잔존하지 않음(AC-HDW-009 `total == inblk`)**.

### M2 — settings.json.tmpl wiring (양 훅) + local manual mirror

**대상**: `internal/template/templates/.claude/settings.json.tmpl` + `.claude/settings.json` (local) + local `.claude/` 트리 **수동 미러**(make build는 EMBEDDED만 갱신, local 트리 미접촉 — D5)

작업:
1. PostToolUse hooks 배열에 `status-transition-ownership.sh` command 엔트리 추가 (platform-conditional 패턴, `$CLAUDE_PROJECT_DIR`, timeout 5s). matcher는 기존 PostToolUse 블록 matcher와 정합 (Write|Edit; status-transition 훅 자체가 SPEC 아티팩트 경로로 self-filter).
2. Stop hooks 배열에 (일반화된) `sync-phase-quality-gate.sh` command 엔트리 추가 (platform-conditional, timeout 10s — 테스트 실행 시간 고려).
3. `team-ac-verify.sh`는 어떤 블록에도 추가하지 않음 (의도적 제외).
4. local `.claude/settings.json`에 동일 2개 엔트리 추가 — dev-intent 키 미접촉(git diff로 ADD-only 검증).
5. `make build`로 EMBEDDED 재생성(`go:embed → embedded.go`) **+ 별도로** local `.claude/hooks/moai/sync-phase-quality-gate.sh`를 template 소스에서 **수동 복사**(make build는 local 트리를 갱신하지 않음 — D5; `diff -q`로 byte-identity 확인).

**Exit 기준**: settings.json.tmpl이 `moai init` 렌더 smoke test 통과(유효 JSON); template/tmpl + local settings.json 양쪽에 2개 엔트리 존재; team-ac-verify grep 0; dev-intent 키 git diff 무변경.

### M3 — neutrality / coexistence 검증

작업:
1. neutrality CI guard 실행 (**D2 — bare-pipe + 정확한 함수명**): `go test ./internal/template/ -run 'TestTemplateNeutralityAudit|TestTemplateNoInternalContentLeak' -v -count=1` → 두 named test 모두 `--- PASS`, `no tests to run` 미출력 (escaped-pipe `\|`는 리터럴 매칭 → vacuous 금지). 이 guard는 **내부-콘텐츠 누출만** 검증함(AC-HDW-006).
2. **Go-bias boundedness (D3 — 실제 언어중립성 자동 guard)**: AC-HDW-009 실행 — `awk` case-block 추출 후 Go-tool 토큰 `total == inblk` 확인. neutrality CI guard에는 Go-bias 클래스가 없으므로 이 단계가 언어중립성 입증의 핵심.
3. coexistence: 기존 `handle-post-tool.sh` / `handle-stop.sh` / `handle-harness-observe-stop.sh` / `handle-task-completed.sh` 엔트리가 settings.json.tmpl에 모두 잔존(회귀 없음) 확인.
4. local↔template parity (**manual mirror** 결과): sync-gate local vs template `diff -q` = 0; settings 엔트리 parity.
5. advisory-only 확증: 등록된 sync-gate 구성이 본 SPEC 기본 경로에서 exit 2를 강제하지 않음(warn-first, 실 git-repo passing fixture로 검증).
6. 전체 회귀: `go test ./...` GREEN.

**Exit 기준**: acceptance.md AC-HDW-001..009 전부 PASS.

## G. Anti-Patterns (avoid)

- sync-gate에서 Go 키워드를 default/공통 경로에 두어 비-Go 프로젝트에서 실행 → 중립성 위반.
- `team-ac-verify.sh` 파일 삭제 (미등록만 해야 함).
- local settings.json에서 hook 외 키 편집 (dev-intent 교란).
- 신규 `handle-sync-gate.sh` wrapper 생성 (불필요 — self-contained 직접 등록).
- `make build`가 local `.claude/` 트리를 동기화한다고 가정 (D5 — make build는 EMBEDDED만; local은 수동 미러 필요). 수동 미러 누락 시 local↔template drift.
- neutrality 검증에 escaped-pipe `-run 'TestX\|TestY'` 사용 (D2 — 리터럴 `\|` 매칭 → `no tests to run` vacuous PASS). bare-pipe alternation 필수.
- AC-HDW-006(내부누출) 통과를 언어중립성 증명으로 오인 (D3 — 별개 속성; 언어중립성은 AC-HDW-009/002).
- non-git temp dir로 sync-gate 런타임 검증 (D1 — git 게이트 단락으로 detect_language 미도달, vacuous). 실 git-repo 또는 detect_language 직접 호출 필수.
- exit-2 blocking 경로 활성화 (본 SPEC scope 밖, warn-first 위반).

## H. Cross-References

- spec.md §D (REQ-HDW-001..010), §Exclusions
- design.md (언어 감지 설계 + 삽입 설계)
- acceptance.md (AC-HDW-001..009)
- CLAUDE.md §7, CLAUDE.local.md §2/§15/§22/§25
