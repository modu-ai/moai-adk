# progress.md — SPEC-PRECOMMIT-GATE-SCOPE-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-PRECOMMIT-GATE-SCOPE-001
card: t461
branch: WT-precommit-gate-scope
base: a239cf050
status: draft
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
verified_facts:
  install_call_site: "installPreCommitHookOptional(...) — internal/cli/update_template_sync.go (심볼 인용; 카드의 :574는 실측 :575로 drift)"
  twin_pair: "preCommitHookContent 상수 (internal/cli/hook_install_precommit.go) + internal/template/templates/.git_hooks/pre-commit (3,245 bytes)"
  twin_guard: "TestPreCommitTemplateMatchesConstant (internal/cli/hook_install_precommit_test.go) — AC-PGM-010 교차참조 (internal/cli/precommit_relocation_test.go)"
  defect_entry: "883d53852 (2026-07-28, SPEC-PRETOOL-GATE-MOVE-001, #1189) — 사전 커밋 52b5e4bf5 (2026-07-05, SPEC-PRECOMMIT-001)는 무해 fast subset"
  gate_defaults: "NewDefaultGateConfig(): Enabled=true, SkipTests=false (internal/config/defaults.go)"
  gate_keys: "GateConfig.enabled / skip_tests / disabled_steps (internal/config/types.go) — 카드 지목 3키 모두 존재 확인"
  gate_yaml_path: ".moai/config/sections/gate.yaml (loadGateSection, internal/config/loader_gate.go)"
  config_wipe: "CleanMoaiManagedPaths가 .moai/config/ 통째 삭제 후 템플릿 재배포 (internal/cli/update/deploy/deploy.go) — REQ-006의 근거"
  enabled_shared_switch: "QualityGate.Run이 Enabled=false에서 즉시 통과 (internal/hook/quality/gate.go) — 단독 moai gate와 공유 스위치"
t237_collision: "t237 / issue #1641 — 동일 twin 파일 편집 카드. 본 SPEC 소관 아님(REQ-008/AC-005). 병합 순서 메모는 run-phase 시 본 파일에 기록"
decisions_recorded:
  - "D1 해소 — 축 (b) 확정: pre-commit 맥락 heavy gate 기본 OFF, gate.yaml opt-in (Implementation Kickoff Approval, 운영자 결정 2026-09-03)"
  - "D1 해소 — 메커니즘 1 확정: 훅이 MOAI_PRECOMMIT=1 마커와 함께 moai gate 호출, 러너는 마커 하에서만 gate.pre_commit.enabled 존중. 단독 moai gate 불변, 새 서브커맨드 없음, 러너 분기점 1개"
  - "D1 해소 — 신설 REQ-009: gate.pre_commit.enabled를 moai web 설정 화면에서 편집 가능 (운영자 지시). 저장은 정확히 .moai/config/sections/gate.yaml"
web_axis_anchors:
  schema_gap: "settings 스키마에 gate 섹션 없음 (internal/settings/schema.go SectionID 목록) — SectionGate 신설 + FieldDef(PersistSeam, Section: gate) 등록 필요"
  render_wiring: "internal/web/schemaform.go schemaSectionMetas() 패널 배선 + fieldsets.templ fieldsetSchemaSection 스키마 주도 렌더"
  naming_convention: "폼 컨트롤 name=\"gate.pre_commit.enabled\" + bool hidden companion __present (parseSchemaForm EC-1)"
  hide_precedent: "workflow_agents 은닉은 레지스트리 수준 부재 (agent_settings_test.go:97 폼 컨트롤 미렌더, :230 TestWorkflowAgentsWebSubmissionIgnored) — 노출은 레지스트리 등록으로 충분, 별도 플래그 불필요"
  save_path: "ApplySchemaEdits → WriteSectionViaSeam (internal/settings/sectionwrite.go) — .moai/config/sections/<section>.yaml yamlpatch 기록, 주석 보존. gate.yaml 정확히 기록 확인"
  i18n: "internal/web/assets/i18n.js sec.*.title 4-locale 패턴 (en/ko 확인, ja/zh는 run-phase 확인)"
  naming_overlap: "git_strategy.<mode>.hooks.pre_commit (validation.go checkStringField 3곳)은 다른 subtree — plan.md 제약 8에 기록"
d5_adjudication: "기각(변경 없음) — Event-detected는 GEARS 정캐논: .claude/skills/moai-workflow-spec/SKILL.md:59 (GEARS Five Patterns 표 행 'Event-detected (replaces IF/THEN)'), .claude/skills/moai-foundation-core/SKILL.md:113 (5패턴 서술에 'Event-detected (replaces the deprecated conditional modality)' 명시)"
```

## §E.2 Run-phase Evidence

### 구현 요약 (확정 설계 = 축 (b) + 메커니즘 1 + moai web 편집)

- **M1** (`8347656b8`): `config.GateConfig`에 `PreCommit GatePreCommitConfig`(yaml `pre_commit`) 신설,
  `NewDefaultGateConfig()` 기본 `enabled: false`. `internal/config/envkeys.go`에
  `EnvPreCommitMarker = "MOAI_PRECOMMIT"` 상수화. 러너 분기점 1개 — `internal/cli/gate.go`
  `runGate`가 마커 `== "1"` && `!cfg.PreCommitEnabled`일 때 project-wide heavy 단계를 건너뛰고
  통과(스킵 안내에 remedy 키 명시). `config.Enabled`는 절대 뒤집지 않음(Known Issue 2 준수).
  `quality.GateConfig`에 `PreCommitEnabled` 필드 추가(런너 `Run`은 읽지 않음 — 전달 전용).
  템플릿 `gate.yaml`에 `pre_commit:` 블록 + 안내 주석. `shipped_key_inventory.yaml`에
  `gate.pre_commit.enabled` (class W, reader internal/cli/gate.go) 등재.
- **M2** (`b18de874c`): settings `SectionGate` 신설 + seam FieldDef 1개
  (`gate.pre_commit.enabled`, PersistSeam Section `gate` Path `[gate pre_commit enabled]`).
  `sectionroute.go` `"gate": RouteSeam` + `sectionwrite.go` `sectionRootKeys` `"gate": {"gate": true}`
  — plan M2 항목 5a의 명시 등록 2건(누락 시 `not seam-writable` 실패). web: `gate` 탭(13번째) +
  `schemaSectionMetas()` 패널, bool companion `gate.pre_commit.enabled__present`,
  `wantTabOrder`에 gate 추가, i18n.js 4-locale(sec.gate.title/desc + f.gate.pre_commit.enabled.title/desc).
- **M3+M4** (`8ab11ed99`): twin 동시 편집 — heavy gate 블록을 `MOAI_PRECOMMIT=1 moai gate` 호출로
  전환, 실패 안내 5문자열(`.moai/config/sections/gate.yaml`, `gate.pre_commit.enabled`,
  `gate.enabled`, `gate.skip_tests`, `gate.disabled_steps`) + `SKIP_MOAI_PRECOMMIT=1` 유지(AC-001).
  fast-subset 구간(STAGED_GO 수집, gofmt, `go vet $BT_TAGS $PKGS`, build-tags) 무변경 — AC-005.
  E2E 훅 테스트 4건(실제 git repo + moai 테스트 더블).
- **M5**: CHANGELOG·사용자 문서는 manager-docs 소관(에이전트 소관 매트릭스 — run-phase 변경 금지 표면).
  sync 커밋에 반입할 항목: 신설 키 `gate.pre_commit.enabled`(기본 false, pre-commit 맥락 한정),
  `moai web` Gate 패널 경로, remedy 안내 확장. 훅 본문·gate.yaml 주석·web 패널에 이미 사용자
  도달 표면 존재.
- **M6**: catalog hash 재계산 완료(`go run ./internal/template/scripts/gen-catalog-hashes.go --all`,
  `catalog.yaml updated successfully` — make build가 t443 드리프트로 선행 단계에서 중단되어 직접 실행).

### RED-GREEN 증거

| 단계 | RED (구현 전) | GREEN |
|------|---------------|-------|
| config 스키마 | `cfg.PreCommit undefined` build fail 3건 — `.moai/reports/t461/red_config_precommit.log` | `go test ./internal/config/ -run 'TestNewDefaultGateConfigPreCommitDefaultOff\|TestLoadGateSectionPreCommit'` PASS |
| 러너 분기 | `TestRunGatePreCommitMarkerSkipsHeavyGate` FAIL — 마커 하에서도 heavy 실행돼 `quality gate failed` — `.moai/reports/t461/red_gate_precommit.log` | 5건 전부 PASS |
| twin 편집 | `TestUpdateReplacesMarkerHookWithCurrentContent` FAIL — 교체된 훅에 `MOAI_PRECOMMIT=1 moai gate` 부재 — `.moai/reports/t461/gate_yaml_preserve_run1.log` | twin identity + AC-008 PASS |
| web 패널 | `data-tab="gate"` 등 4 probe 미렌더 + i18n 키 4locale 부재 — `.moai/reports/t461/red_web_gate.log` | `go test ./internal/web/` 전량 PASS |

### 뮤턴트 프로브 (공허 초록 방지 — AC-002)

구 hook(`a7f30b373`의 `.git_hooks/pre-commit`, 마커 export 없음 — 6건 히트 모두 `SKIP_MOAI_PRECOMMIT`)을
같은 시나리오(가짜 moai = 실패하는 project-wide 검사, opt-in 없음)에 넣으면 **커밋 차단 exit 1** —
`.moai/reports/t461/mutant_probe_old_hook_blocks.txt`. 새 훅은 같은 시나리오에서 커밋 성공
(`TestPrecommitE2EDefaultAllowsUnrelatedCommit`). 즉 AC-002 테스트의 초록은 마커 export에 의존한다.

### REQ-006/AC-004 소관 판정 (기계 존재 검증)

`.moai/config` 소거 위에서 gate.yaml의 생존은 **기존 update 파이프라인**(Backup → Clean Managed Paths →
Deploy Templates → Restore Settings, `RestoreMoaiConfigRetained`의 3-way node merge,
SPEC-UPDATE-YAML-PRESERVE-001)이 이미 담당한다 — 신설 기계 장치 없이 keep+pin 테스트로 계약을 고정했다
(`update_gate_yaml_preserve_test.go`, llm.yaml 선례와 동일 드라이버 `runTemplateSyncAt`):
손편집값(`skip_tests`, `disabled_steps`, 사용자 키·주석)과 web 작성값(`pre_commit.enabled: true`) 모두
update 후 유지, 신규 키 전달도 확인. 카드가 가정한 "gate.yaml 원상복구 결함"은 실측에서 재현되지 않았다.

### t237 충돌 메모 (REQ-008/AC-005)

t237 / issue #1641(go vet 모듈해석 수리)과 본 카드는 **같은 twin 파일**(`preCommitHookContent` +
`.git_hooks/pre-commit`)을 편집한다. 본 카드의 diff는 fast-subset 구간에 대해 주석 수준(헤더 2줄)이며
`go vet`/`STAGED_GO`/`BT_TAGS`/`PKGS` 줄은 무변경(`git diff a7f30b373 -- …` 그렙 실측, AC-005).
t237 착지 시 본 카드 브랜치(`WT-precommit-gate-scope`)와의 재충돌은 t237 쪽 rebase 소관이고, 충돌 지점은
heavy-gate 블록(본 카드) vs go vet 블록(t237)으로 인접하지만 겹치지 않는다.

### t443 빌드 드리프트 메모

`make build`는 `agents-emit-check`에서 **t443 소관 드리프트로 실패**한다(본 카드 무관 — agent .md 편집 0):
`TestGoldenCommittedArtifactsMatchEmission: .codex/agents/moai/sync-auditor.toml: committed artifact
differs from emission (sha256 mismatch)`. 레인 지시에 따라 수리하지 않았고 `make agents-emit`도
실행하지 않았다. 대체: `go build ./...` exit 0, `GOOS=windows GOARCH=amd64 go build ./...` exit 0,
catalog hash는 직접 재계산. `TestPreCommitTemplateMatchesConstant`는 소스 파일을 직접 읽어 무영향.

### 상속 적색 (본 카드 변위 아님 — 귀속)

- `internal/config` `TestAlwaysLoadedTokenBudget`: always-loaded 표면 77,104 tokens > 예산 76,400
  (overflow 704). 본 카드 diff는 `.claude/`·`templates/.claude/` 아래 **0 파일** 변경
  (`git diff a7f30b373 --stat -- .claude/` → 0행) — 측정 대상 표면이 base와 동일하므로 base에서도
  동일하게 적색이다. 수리는 본 카드 소관 아님(별도 카드: 예산 상향 or 룰 다이어트).

### 훅 본문 변경의 필수 형제 수리 — AC-PCP-005 sub-case (c) 플립

본 카드의 legitimate hook-body 변경은 SPEC-PRECOMMIT-PRESERVE-001 acceptance.md §D.4 point 2가
규정한 절차를 트리거한다: 새 본문을 실은 릴리스 태그가 존재하기 전까지는 어떤 태그도 pin 규칙을
만족하지 못하며, 그 창에서 v3.1.2-era 훅은 incoming 본문과 genuinely user-different다 — 즉 sub-case (c)의
올바른 기대는 backup AND notice, sub-case (a)의 형태다. 이에 따라 `TestPreCommitLegacyNoRecord/
c_no_record_pinned_released_body`의 기대를 `hookUnmodified` → `hookUserModified`로 플립하고
(플립 사유·재 pin 절차를 테스트 주석에 기록), 새 본문을 실은 첫 릴리스 컷 시점에
`pinnedReleasedHookTag`를 그 태그로 재지정하고 원래 단정을 복원하는 것이 릴리스 체크리스트 항목이다.
(c)는 삭제하지 않았다(§D.4 point 1). 최초 전체 실행에서 관측된 적색
(`attribution = user-modified, want unmodified` + backup 생성)이 바로 이 플립의 RED 관측이다
(`.moai/reports/t461/internal_cli_full.log`).

### 검증 부하 측정 (리드 요구 통제군)

`go list -deps ./internal/cli/... ./internal/config/... ./internal/hook/... ./internal/settings/... ./internal/web/...`
→ moai 패키지 116개, 리드가 제시한 흡수-delta 합집합(internal/core/project, internal/mx, internal/profile)
을 **포함**하고 추가 의존자 없음(graph, merge, spec, statusline, template가 추가로 포함 — 전부 본 축의 의존 방향).
internal/cli는 단독 실행(`-timeout 1800s`).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-03
run_commit_sha: pending-backfill-run
run_status: complete
ac_pass_count: 10
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a (로컬 레인 — push 금지, develop 병합은 리드 창 경유)
l44_post_push_fetch: n/a (동일 — push는 리드 일괄)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: pass (go build ./... exit 0)
  windows: pass (GOOS=windows GOARCH=amd64 go build ./... exit 0)
total_run_phase_files: 28
m1_to_mN_commit_strategy: M1 8347656b8 / M2 b18de874c / M3+M4 8ab11ed99 / M6+evidence 최종 커밋
inherited_reds:
  - "internal/config TestAlwaysLoadedTokenBudget — always-loaded 77,104 > 76,400; 본 카드 diff가 측정 표면(.claude/)을 건드리지 않아 base 동일 적색 (귀속 완료)"
external_blockers:
  - "make build agents-emit-check — t443 sync-auditor.toml 드리프트 (본 카드 무관, 수리 금지 지시 준수)"
m5_handoff:
  - "CHANGELOG + 4-locale 사용자 문서는 manager-docs sync 소관 — §E.2 M5 절의 반입 항목 목록 참조"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-03
sync_commit_sha: b0cd51195
sync_status: complete
changelog_entry_position: "CHANGELOG.md [Unreleased] → Added 섹션 말미 (SPEC-PRECOMMIT-GATE-SCOPE-001 1건, 사전 중복 grep 0 히트)"
b12_self_test_a: "grep -c 'SPEC-PRECOMMIT-GATE-SCOPE-001' CHANGELOG.md → 0 (커밋 전, 중복 없음 확인)"
b12_self_test_b: "acceptance.md AC 식별자 10건 (AC-001..AC-010, grep + sort -u 실측) — CHANGELOG 항목은 키 단위 서술로 10 AC 전체를 커버"
b12_self_test_c: "CHANGELOG가 언급하는 경로 실존 확인 — .moai/config/sections/gate.yaml (internal/template/templates/.moai/config/ 동일), .git_hooks/pre-commit 마커 호출 실측"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (sync 커밋 탑승, updated: 2026-09-03 유지)"
  plan_md: "상태축 없음 (Artifact Statelessness — status 필드 부재, 무변경)"
  acceptance_md: "상태축 없음 (동일, 무변경)"
canary_compliance_check:
  docs_code_match: "gate.pre_commit.enabled 기본 false / MOAI_PRECOMMIT=1 마커 한정 / moai web Gate 패널이 정확히 .moai/config/sections/gate.yaml 기록 — §E.2 M1/M2 구현과 문서 서술 일치"
  template_neutrality: "sync 단계에서 internal/template/templates/** 변경 0 (git status 실측) — 중립 CI 가드 무자극"
```

### 동기화 산출물

- CHANGELOG.md — [Unreleased] Added에 사용자 가시 변경 1항목 추가 (신설 opt-in 키, 기본 동작 변경, `moai web` 토글).
- docs-site `/moai gate` 페이지 4-locale 확장 (ko/en/ja/zh — ko 캐논, 동일 구조): 새 `##` 섹션 1개 —
  `gate.pre_commit.enabled` 키, `.moai/config/sections/gate.yaml` opt-in, `MOAI_PRECOMMIT=1`
  마커 한정 동작, `moai web` Gate 패널 경로, `SKIP_MOAI_PRECOMMIT=1` 우회. 기존 섹션 재작성 없음(범위 최소).
  - ko: `docs-site/content/ko/utility-commands/moai-gate.md` (섹션 7 → 8)
  - en: `docs-site/content/en/utility-commands/moai-gate.md` (7 → 8)
  - ja: `docs-site/content/ja/utility-commands/moai-gate.md` (7 → 8)
  - zh: `docs-site/content/zh/utility-commands/moai-gate.md` (7 → 8)
- README 4-locale: gate 설정 키를 나열하는 섹션 부재 실측 (`gate.yaml` 언급은 ast_grep_gate 버전노트 1곳뿐) —
  건드리지 않음(부재는 유효 결과).
- 검증 증거: `.moai/reports/t461/hugo_build_sync.log` — `hugo --gc --minify` exit 0, WARN/ERROR 0행,
  4-locale 패리티 grep `gate.pre_commit.enabled` 4파일 × 2히트 동일.

