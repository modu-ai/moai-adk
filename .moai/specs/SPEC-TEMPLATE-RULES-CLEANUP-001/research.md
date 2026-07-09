# SPEC-TEMPLATE-RULES-CLEANUP-001 — 조사 기록 (research.md)

## 출처 및 재검증

- **1차 감사**: 2026-07-09, session 98a5197a — `internal/template/templates/.claude/rules/` 62파일 3-렌즈 감사 (memory: `project_template_rules_audit_2026_07.md`)
- **재검증**: 2026-07-09, 본 plan-phase 세션 (manager-spec) — 아래 각 finding의 load-bearing anchor를 현재 트리 대비 grep/diff/ls로 실측 재확인. 검증되지 않은 채 AC로 인코딩된 주장 없음 (verification-claim-integrity §1.1 surface 3 준수).
- **가드 현황 baseline**: `go test ./internal/template/ -run 'Mirror|Leak|Neutral|Language'` 전부 PASS (감사 시점) — 즉 아래 finding 전량이 현행 가드 사각지대.

## Finding A — 깨진 cross-reference 5건 (재검증 완료)

증거 명령: `grep -n 'dev-only-commands-isolation' <agent-authoring.md> <skill-authoring.md>` 외.

| # | 파일:라인 | 인용 대상 | 상태 |
|---|-----------|-----------|------|
| A1 | `development/agent-authoring.md:25,38` | `.moai/docs/dev-only-commands-isolation.md` + `97-*`/`98-*` wrapper 표기 | 확인 — 25행에 `97-*`/`98-*` per dev-only-commands-isolation.md 인용 실측 |
| A2 | `development/skill-authoring.md:357` | 동일 문서 | 확인 |
| A3 | `workflow/spec-workflow.md:25` | `.moai/docs/git-local-workflow-doctrine.md` | 확인 |
| A4 | `development/manager-develop-prompt-template.md:107` | `.moai/docs/git-workflow-doctrine.md` | 확인 |
| A5 | `workflow/spec-workflow.md:124` | `.moai/design/v3-redesign/synthesis/pattern-library.md` §O-6 | 확인 — dev repo에도 부재 (완전 사장 참조) |

비고: agent-authoring.md는 17/25/30/32/38행에 `.claude/agents/local/` 언급 — 이 중 30/32행의 `moai update` 보호 계약은 배포 프로젝트에도 유효한 normative 내용이므로 보존 (REQ-TRC-012).

## Finding B — 중립성 위반 (재검증 완료)

- **REQ/AC 토큰 7건**: `askuser-protocol.md:177`(AC-ADM-005..017), `settings-management.md:181,344,355,364`(REQ-MIG003-006, REQ-WF006-006/015/011), `sprint-round-naming.md:23,94`(AC-LR-009) — 전건 grep 실측. 추가 발견: `manager-develop-prompt-template.md:175`의 `AC-XXX-001`은 pedagogical placeholder (가드 allowlist 대상, 제거 아님 — §R9).
- **askuser-protocol provenance 프로즈**: :252-258 실측 — "Epic 7 TMC-001 plan-phase L51 도출 원천 해소", "§24 namespace align 후속" (Worked Example의 description/preview 내).
- **lessons #N / W# 11건**: `manager-develop-prompt-template.md:71`(lessons #21 W0), `:91`(W1/W2), `:221`(W3 케이스), `agent-common-protocol.md:274`(W3 meta-analysis), `session-handoff.md:291`(lessons #13,#12), `:332`(lessons #14), `spec-workflow.md:51`(lessons #13) — 실측; `manager-develop-prompt-template.md:132,222` + `agent-common-protocol.md:409-413`는 감사 기록 인계분 (run-phase에서 광패턴 grep으로 전수 재확인).
- **내부 날짜 9건**: `design/constitution.md:12-15`(HISTORY), `:423-424`(footer "Last Updated: 2026-05-20"/"Relocated: 2026-04-20"), `worktree-integration.md:44,46` + `spec-workflow.md:175,269,416` + `session-handoff.md:330`(반복 "2026-05-17 policy"), `zone-registry.md:629-630`(2026-05-04/05-09) — 전건 grep 실측.
- **CONST-V3R\***: `zone-registry.md` 121줄 (grep -c 실측, 감사 수치 일치), `manager-develop-prompt-template.md` 4줄(:27,28,32×2,36 — 5토큰/4줄), `worktree-integration.md` 2곳. `settings-management.md:174` standalone `MIG-003` heading 실측.
- **한국어 15/62 파일**: 사용자 결정 4에 따라 Out of Scope 이연 (기존 가드 `korean_leftover_audit_test.go`는 recommended-marker 등 협스코프만 커버 — 실측).

## Finding C — 백포트 (재검증 + **스코프 정제**)

- C1 `spec-frontmatter-schema.md`: local :92에 `§I Token Accounting` row 실존, template 측 grep 0 hit — 백포트 확정. row 내용은 SPEC ID 무포함 (중립).
- C2 `runtime-recovery-doctrine.md`: **감사 기술과 실측의 차이를 확인, 스코프 축소** — §R1/§R2 참조.

## Finding D — retired 어휘 (재검증 완료)

- D1 `design/constitution.md:323-350`: "Sprint Contract Protocol" 섹션 + `:346-348` `[ZONE:Frozen] [HARD]` 조항 + `:350` `.moai/sprints/` 인용 실측. 파일 헤더는 RETIRED 배너 보유하나 해당 블록은 미주석.
- D2 `orchestration-mode-selection.md:190`: "SPEC cohorts" 실측 (AP-SRN-005 위반).
- **주의**: `.moai/sprints`는 `design.yaml gan_loop.sprint_contract.artifact_dir`로 config에 살아 있고 `LoadDesignConfig`(loader_design.go)가 소비 — rules 프로즈만 정리 대상 (§R6과 함께 판단).

## Finding E — design 역드리프트 (재검증 + **스코프 정제**)

- `bda9a3f33` stat 실측: template constitution.md 18줄 변경, template design.yaml 7줄 삭제(= `design_docs:` 블록), `.moai/design/` 트리 삭제, `design_folder.go` 등 Go 삭제. `grep -rn 'design_docs|DesignDocs' internal/ --include='*.go'`(non-test) → 0 hit 확인.
- local `design/constitution.md` 21,766B vs template 22,581B — stale 확인.
- local `design.yaml` 1,614B vs template 2,867B — **diff 실측: local에 `design_docs:` 블록 잔존** (auto_load_on_design_command / dir: .moai/design / priority / token_budget). 이것이 "stale remnant"의 실체 — §R6.

## Group F 설계 근거 — 기존 가드 구조 실측

- **§R3 date tier-ownership**: `internal_content_leak_test.go:857` `TestLeakClassNoDateShaInDefaultTier` — default-tier `leakClasses`는 날짜/SHA probe에 매치되면 즉시 FAIL하는 계약 ("date detection is owned by the strict tier (ISOLATION-001), not this SPEC"). → **가드 (c)는 별도 테스트 함수로 구현해야 함** (REQ-TRC-062).
- **§R4 REQ/AC 패턴 기존재**: `internal_content_leak_test.go:246-248` `S3-req-ac-token-any-prefix` class가 `\b(REQ|AC)-[A-Z][A-Z0-9]*-[0-9]+\b` 완전 일반형 패턴을 이미 보유하되 `skillBodyScoped: true` + 주석 "REQ/AC tokens in agents/rules are owned elsewhere (EXCL-SBN-002)". → 가드 (a)는 신규 regex가 아니라 **rules 스코프 확장**. (감사가 언급한 ATR/WO/COORD/UNP/LNC/TII prefix allowlist의 정확한 코드 위치는 `agent_frontmatter_audit_test.go`/`contract_schema_test.go`/`implementation_kickoff_approval_preservation_test.go` 중에 있음 — grep으로 후보 파일 4개 확인, 정확 seam은 run-phase M1에서 확정.)
- **§R8 allowlist 메커니즘 기존재**: `pedagogicalAllowlistEntry{File, LineStart, LineEnd, SpecID, Rationale}` 구조 + literal-substring 매칭 — rules 스코프 확장 시 재사용.
- 기존 recurrence backstop 선례: `TestSkillBodyLeakClassRecurrenceBackstop` (synthetic re-leak probe + clean replacement) — REQ-TRC-066 모델.

## 스코프 정제 (본 세션 발견 — 감사 원문 대비 델타)

- **R1 — runtime-recovery-doctrine template 실존**: template mirror 존재 (17,984B, Jul 7). 구 기억("template-ABSENT")은 stale — 파일이 이후 추가됨. byte-parity allowlist에는 **의도적으로 미등재** (mirror test 주석 실측: "deliberately NOT added ... its template mirror strips internal provenance ... byte-parity cannot hold").
- **R2 — C2 백포트는 용어 한정**: 감사가 "누락"으로 기록한 `.moai/research/dive-into-claude-code-archive.md` 참조와 Version/Origin footer의 template 부재는 **sanitized-pair 계약의 의도적 산출** (mirror test 주석 + `sanitized_pair_parity_test.go:71-76` registry 실측). 백포트 대상은 `MoAI`→`moai-adk` 용어 ~8곳 + local 측의 더 중립적인 표현("The orchestrator-interrupt-ledger contract owns...")만. local 측 `CONST-V3R6-001` 토큰은 역방향으로 template에 절대 복사 금지 (REQ-TRC-032).
- **R5 — zone-registry 런타임 소비자 3곳**: `internal/cli/constitution.go:24`(constitutionRegistryRelPath), `internal/cli/doctor.go:580`(not-found 메시지 — graceful), `internal/cli/spec_lint.go:159`(`detectRegistryPath` → 부재 시 `""` 반환 — graceful), `internal/constitution/validator.go`. 템플릿 제거는 사용자 프로젝트의 `moai constitution list` 동작에 영향 — scratch 실측 검증 필수 (REQ-TRC-027). `internal/template/catalog.yaml`에는 zone-registry 미등재 (grep 0 hit) — 카탈로그 변경 불요.
- **R6 — design.yaml은 라이브 config**: template이 여전히 배포 (2,867B, Jul 8), `LoadDesignConfig` + `Loader.Load()` 체인(loader.go:95) 활성, `TestStructYAMLSymmetry_Design`은 **template 측** design.yaml을 읽음 (audit_struct_yaml_symmetry_test.go:190-191 실측 — repoRoot/internal/template/templates 경로). → Finding E의 design.yaml 처리는 **파일 삭제가 아니라 stale `design_docs:` 블록 제거 + template baseline 정렬** (REQ-TRC-051). local diff 실측으로 `design_docs:` 블록 잔존 확인.
- **R7 — 트리 분류 3종**: mirror test는 명시 allowlist 기반 byte-parity (`workflowOptMirroredPaths`: spec-workflow.md, session-handoff.md, hooks-system.md, model-policy.md + evaluator profiles). 편집 대상 중 spec-workflow.md/session-handoff.md는 **양 트리 동일 편집 필수**. manager-develop-prompt-template.md/runtime-recovery-doctrine.md는 sanitized pair — **template 측만**. 나머지는 unenrolled. (plan.md §A.1 매트릭스로 정식화 — memory lesson `template-tree-is-subset-of-live` 적용.)
- **R9 — pedagogical placeholder**: `manager-develop-prompt-template.md:175` `AC-XXX-001` (예시표) — 가드 (a)의 `-XXX-` regex 제외 또는 allowlist 등재 필요. 제거 대상 아님.

## 잔여 미검증 (Gaps — run-phase 이관)

- `manager-develop-prompt-template.md:132,222`, `agent-common-protocol.md:409-413`의 lessons/W# 인스턴스는 감사 기록 인계분 — M3에서 광패턴 grep으로 전수 확정 (본 세션은 대표 인스턴스만 실측).
- ATR/WO/COORD/UNP/LNC/TII prefix allowlist의 정확한 코드 위치 — M1에서 확정 (후보 4파일 좁힘 완료).
- `moai constitution list`의 registry-부재 동작 — M3 scratch 실측 전까지 미확정 (doctor/spec_lint는 코드 레벨 graceful 확인).
- 내부 날짜 low-confidence 6건 (감사 기록) — M3에서 date 가드 RED 출력으로 전수 표면화 후 판정.
