# Plan — SPEC-V3R6-DEV-HARNESS-SPLIT-001

> Tier S (minimal). devkit 단일 진입을 3개 독립 harness 커맨드로 분리.
> 본 plan은 WHAT/HOW의 *순서*만 정의한다. 산출물 파일 생성·수정은 run-phase(manager-develop) 책임.

## §A. Context

SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001(completed, `9ef450f74`)이 통합한 단일 `/harness:devkit` 진입을 3개 독립 harness로 분리한다. CONSOLIDATION-001의 통합-진입 *결정만* 번복하고, 그것이 전달한 3개 specialist body는 verbatim 재사용한다.

## §B. Known Issues / 사전 검증 사실 (re-investigate 금지)

- 통합 진입 + Runner + manifest 존재 확인: `.claude/commands/harness/devkit.md`, `.claude/commands/harness/manifest.json`, `.claude/workflows/harness-devkit-run.js`.
- 3개 specialist body 존재 확인: `.claude/agents/harness/harness-devkit-{release-update,github,release}-specialist.md`. 이식된 multi-phase workflow 본문 포함.
- CI guard 존재 확인: `internal/template/devkit_namespace_test.go` (`TestDevkitNamespaceNoLeak`) — embedded tree에 `commands/harness/` / `harness-devkit*` / `workflows/harness-devkit-*` 부재 단언, sentinel `DEVKIT_NAMESPACE_LEAK`.
- `.claude/agents/harness/`에는 무관한 Layer B specialist(cli-template/hook-ci/quality/workflow)도 존재 — **건드리지 않는다**.
- Runner는 처음부터 release-update의 fan-out만 모델링 (github/release는 fan-out 없음).

## §C. Pre-flight (run-phase 진입 전 확인)

- [ ] `git rev-list --count --left-right origin/main...HEAD` — 병렬 세션 race 점검 (공유 main orphan 방지)
- [ ] `.claude/agents/harness/` Layer B 4개 specialist 파일 목록 캡처 (touch 금지 대상 고정)
- [ ] `go test ./internal/template/... -run TestDevkitNamespaceNoLeak` 현재 PASS 확인 (baseline)

## §D. Constraints

- Tier S — 정확히 3 harness, 추가 추상화 없음.
- github/release: Runner 없음 + manifest 없음 (fan-out 부재 → 단순성 우선). thin command → specialist 직접 라우팅.
- release-update: thin command + `release-update/manifest.json` + `harness-release-update-run.js` Runner + specialist.
- Scope: devkit→3-split + CI guard + doctrine만. Layer B / 사용자 template / CONSOLIDATION-001 closed artifacts 불가침.

## §E. Self-Verification (plan-phase audit-ready)

- [ ] frontmatter 12 필드 + `tier: S` + `depends_on` 존재
- [ ] SPEC ID regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` 매칭 (decomposition PASS)
- [ ] §E Exclusions에 `### Out of Scope — <topic>` H3 + `-` bullet 존재 (OutOfScopeRule)
- [ ] 7개 scope 항목이 각각 REQ + AC로 매핑됨
- [ ] spec.md에 구현 세부(함수명/클래스 구조) 없음

## §F. Milestones (priority-ordered, no time estimate)

### M1 — specialist 이름 변경 (3개) — REQ-DHS-002
- `harness-devkit-release-update-specialist.md` → `harness-release-update-specialist.md`
- `harness-devkit-github-specialist.md` → `harness-github-specialist.md`
- `harness-devkit-release-specialist.md` → `harness-release-specialist.md`
- 각 파일: frontmatter `name` 갱신, 진입 자기-참조(`/harness:devkit X` → `/harness:X`) 갱신, Migration Provenance에 본 SPEC 인용 추가. body workflow 본문은 verbatim.
- `[DEV-ONLY]` 배너 유지.
- (manager-develop: `git mv`로 이름 변경 후 Edit — git 이력 보존)

### M2 — release-update Runner + manifest 재범위화 — REQ-DHS-003
- `git mv .claude/workflows/harness-devkit-run.js .claude/workflows/harness-release-update-run.js`
- Runner 내부 `MANIFEST_PATH` + 주석을 release-update 전용으로 재지향 (`.claude/commands/harness/release-update/manifest.json` 경로).
- `.claude/commands/harness/release-update/manifest.json` 생성 (release-update single-specialist manifest — `specialists` 1개 entry).
- **Justify**: Runner는 처음부터 release-update fan-out 전용이므로 삭제+재생성이 아니라 rename+재범위화가 정직하다 (이력 보존 + 의미 동일).

### M3 — 3개 독립 thin command 생성 — REQ-DHS-001
- `.claude/commands/harness/release-update.md` — `argument-hint: "[--since vX.Y.Z | --dry]"`, body는 Runner/specialist 라우팅.
- `.claude/commands/harness/github.md` — `argument-hint: "issues|pr [...]"`, body는 `harness-github-specialist` 직접 라우팅 (Runner 없음).
- `.claude/commands/harness/release.md` — `argument-hint: "[VERSION] [--hotfix]"`, body는 `harness-release-specialist` 직접 라우팅 (Runner 없음).
- 각 20 LOC 미만, `[DEV-ONLY]` 문구 + `(dev-only). NOT distributed to user projects.` 유지.

### M4 — 통합 자산 제거 — REQ-DHS-003
- `git rm .claude/commands/harness/devkit.md`
- `git rm .claude/commands/harness/manifest.json`

### M5 — CI guard 재지향 — REQ-DHS-004 / REQ-DHS-005
- `internal/template/devkit_namespace_test.go`의 단언을 3개 분리 namespace로 갱신 (필요시 test fn/file rename — 예: `split_namespace_test.go` / `TestSplitHarnessNamespaceNoLeak`).
- 단언: embedded tree에 `.claude/commands/harness/` 경로 부재 + `harness-{release-update,github,release}*` agent 부재 + `.claude/workflows/harness-*` 부재.
- RED→GREEN 절차로 검증 (누출 심기 → FAIL → 제거 → PASS).

### M6 — doctrine 갱신 (5개 surface) — REQ-DHS-006
- `.moai/docs/dev-only-commands-isolation.md`: 배포 금지 파일 목록 표 + 검증 체크리스트를 3개 분리 자산으로 갱신, migration note에 본 SPEC 인용.
- `CLAUDE.local.md` §2 Local-Only Files + §21 stub: `commands/harness/devkit*` → 3개 분리 패턴.
- `.claude/rules/moai/development/skill-authoring.md` Deprecated Skill Slots table: harness-devkit 참조 갱신.
- `.claude/skills/moai-foundation-core/modules/INDEX.md`: harness:devkit 참조 갱신.
- `.claude/skills/moai/references/reference.md`: harness-devkit 참조 갱신.

## §G. Anti-Patterns (피할 것)

- github/release에 정당화 없이 Runner/manifest 추가 (없는 fan-out에 대한 over-engineering).
- specialist body workflow 본문 재작성 (verbatim 보존 위반 — 이름/라우팅/provenance만 변경).
- Layer B harness specialist touch.
- CONSOLIDATION-001을 superseded로 표시.
- `internal/template/templates/` 하위에 harness 자산 누출.
- doctrine에서 stale `harness-devkit` 참조 잔존.

## §H. Cross-References

- spec.md §C (REQ-DHS-001..007), acceptance.md (AC 매트릭스)
- `.claude/skills/moai/workflows/harness-builder.md` — v4 artifact shapes
- `.moai/docs/dev-only-commands-isolation.md` — dev-only 격리 doctrine
- `.claude/rules/moai/development/coding-standards.md` § Thin Command Pattern

## §I. Sync-phase 메모리 갱신 (REQ-DHS-007 — plan-phase에서 작성 금지)

sync-phase에서 manager-docs가 처리할 항목:
- `project_*` 메모리 entry 생성 (split 완결 기록 + supersession-of-decision 관계).
- MEMORY.md index 1줄 갱신.
- CONSOLIDATION-001 관련 기존 entry가 있으면 `[SUPERSEDED by ...]` 마커가 아니라 cross-link만 (CONSOLIDATION-001은 completed 유지이므로 supersede 마커 금지).
