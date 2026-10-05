# research.md — SPEC-INIT-HARNESS-001

> 측정 기준: 카드 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t585`, 브랜치 `WT-init-harness-q`, HEAD **`a404132e7a53e0e1689a425be426882f2f2455e4`**, 트리 `aebfa5bea6978e81d5c431e57113d2dec9f647ab`. 이 문서의 모든 file:line은 이 HEAD에서 직접 읽거나 실행해 확인했다. 인용 행번호는 HEAD가 움직이면 무효다.
> 원 조사: `.moai/reports/init-tui-audit-20260909.md` — primary 체크아웃의 2026-09-09 보고서(읽기전용 참조). 조사는 t583(2026-09-12 착지) 이전 측정이라 질문 표면 행번호는 낡았고, 이 문서는 그 축을 이 트리로 다시 재었다.

## §0 측정 방법

- 코드 판독: deployer/skill_mirror/questions/init/update/doctor/codexwiring 직접 Read + grep.
- 실행 확인: `git log --all --grep t523`, `git log -- internal/template/commandemit/`, 디렉터 나열(카탈로그 34·게시 16·codex TOML 11 — `ls | wc -l` 수치 37/19/14에서 헤더 3행 제외한 이름 수), SPEC ID 충돌 조회(디렉터 0건 + 카탈로그 grep 0건), SPEC ID 정규식 자가검증(`PASS` 출력).
- 실행하지 않은 것: `moai init` 실실행(실제 HOME 오염 위험 — t583의 홈 지문 슬롯 규약 REQ-IQW-014/016이 run 단계에서 이 SPEC의 실행 테스트에도 준용된다). 배포 무필터 판정은 코드 판독(`deployer.go` walk 순회에 하니스 분기 부재)으로 확정했다.

## §1 카드 전제 검증 표

| # | 카드 전제 (디스패치·조사) | 판정 | 이 트리 근거 |
|---|---|---|---|
| 1 | 배포기에 하네스 필터 없음 — deployer.go:160-277 무조건 전부 심음 | **VERIFIED** (행번호 드리프트 :160-284) | `internal/template/deployer.go:160-284` — walk 본문에 하니스 분기 없음. `DeployWithResult`가 임베디드 FS 전체를 기록 |
| 2 | `--agent codex`도 `.claude` 전체 배포 | **DRIFTED 명칭** — 플래그는 `--llm` | `internal/cli/init.go:132`(`--llm` 선언), `:176-178`(resolveAgentWiring → `getStringFlag(cmd, "llm")`). `--agent` 플래그는 이 트리에 없음. 실질 주장(`--llm codex`에도 `.claude` 전체 배포)는 전제 1로 VERIFIED |
| 3 | 템플릿 CLAUDE.md 이미 `@AGENTS.md` 임포트 (~90%) | **VERIFIED** | `internal/template/templates/CLAUDE.md:9` = `@AGENTS.md` |
| 4 | commandemit 16종이 미병합 워크트리(t553·t562)에만 존재 | **REFUTED — 착지됨** | `internal/template/commandemit/commandemit.go`(커밋 `e7d2a1658`, t503 SPEC-CODEX-COMMAND-SKILLS-001 M1) + 게시 16종 실파일. both 모드의 "최대 공사"는 소멸 |
| 5 | 스킬 미러가 `.claude/skills` 상대링크 — skill_mirror.go:49,151 | **VERIFIED** (행번호 드리프트, 방향 정정) | `internal/template/skill_mirror.go:174-176`(`MirrorLinkTarget` = `../../.claude/skills/<name>`), `:213`(`mirrorOneSkill`). 미러는 `.agents/skills/` 쪽에 있고 원본 `.claude/skills/`를 가리킨다(카드 문구는 방향이 뒤집혀 있었다). `:11-17` — go:embed가 심볼릭 링크를 누락하므로 원본-실파일/링크-배포시 생성 구조는 구속 조건 |
| 6 | `.mcp.json` 템플릿 파일 무조건 배포 | **VERIFIED** | `internal/template/templates/.mcp.json` 존재, walk가 기록. codex는 런타임 프로비저닝만 생략(`internal/cli/init.go:980-987` mcpDeclined 규칙) |
| 7 | `--agent` 의미 변경 vs 신규값 — AC-CW-004·init_agent_flag_test.go가 핀 | **VERIFIED** (명칭 드리프트) | `internal/cli/init_agent_flag_test.go:94-118` — AC-CW-004 = "플래그 부재 ≡ `--llm claude`, `.codex/` 배선 파일 0". 판정 축은 동일, 플래그명만 `--llm` |
| 8 | t523 "결속표 착지" vs 실측 0행 | **REFUTED — 표 존재** | `internal/template/templates/AGENTS.md:21-25` — `Capability / Claude implementation / If this harness lacks it` 표, 데이터 3행(question-channel·task-list·design-sync). t523 병합 `0755cc7f5`가 이 트리 조상. 조사의 "0행" 측정은 t523 이전 또는 낡은 트리 |
| 9 | t583 착지(2026-09-12), 4문항 | **VERIFIED** | `.moai/specs/SPEC-INIT-QUIET-WIZARD-001/` 존재, frontmatter `status: completed`. `questions.go:368-409` Page3Questions = agent_wiring + autonomy_tier |
| 10 | t583이 "하네스 3-way 배포(t585)" 명시 유보 | **VERIFIED** | `spec-compact.md` Exclusions — "자율 등급 재정의(t584) · 하네스 3-way 배포(t585)" |
| 11 | agent_wiring 흡수 시 위저드 4문항 유지 | **MOOT — 질문이 이미 존재** | `questions.go:375-387` — agent_wiring이 곧 하니스 질문(3옵션). 새 질문 불필요. 남은 일은 설명문·옵션 설명의 배포 결과화 |
| 12 | F1 `init_workflow_flags.go:41-44` | **VERIFIED** | `:41-44` — worktree-auto-create Changed 블록(플래그 전용 추적자, t753 범위) |
| 13 | F4 `init.go:542-567·911` | **DRIFTED 행번호** | InitOptions 리터럴 시작 `:565` 부근, 비대칭 서술 `:703-712`(`opts.MCPProvision = true`는 대화형 블록 `:712`), tail 규칙 `:980-987`. 실질(F4)은 유지 — t753 범위 |

## §2 코드 지도 (하니스 축 전수)

### §2.1 해석과 소비

```
--llm 플래그 선언      internal/cli/init.go:132   (closed set help, AC-CW-001)
validateInitFlags      internal/cli/init.go (fail-loud, :38-58 테스트)
normalizeAgentWiring   internal/cli/init.go:159-166  (닫힌 집합, 기본 claude)
resolveAgentWiring     internal/cli/init.go:176-178  (플래그 읽기 원시)
resolveAgentWiringWithWizard internal/cli/init.go:737-739  (단일 해석 지점 — 플래그 > 위저드)
소비자 1: .mcp.json    internal/cli/init.go:980-987  (codex 생략 / both 강제 / claude 대화형 기본 true :712)
소비자 2: codex 배선   internal/cli/init.go:994 → wireCodexUnlessClaude :191-198 → codexwiring.Wire
소비자 3: 배포기       (없음 — 이 SPEC이 추가한다)
```

### §2.2 배포기와 시접

- `DeployWithResult`(`internal/template/deployer.go:148-284`): `fs.WalkDir` 전체 순회, `.tmpl` 렌더, 기존 파일 보호(manifest provenance), R-011 게시 스킬 경로 보호(:238 `isPublishedSkillPath`). 하니스 매개변수 없음.
- 초기화 호출부: `internal/core/project/initializer.go:375-393` — `ResultDeployer` 타입 단언으로 `DeployWithResult`(미러 결과 수집) → 폴백 `Deploy`(:387).
- 슬림 선례: `internal/template/slim_fs.go`(slimFS — 카탈로그 등급 FS 래퍼, 비-코어 은닉), 진입 `internal/template/embed_catalog.go:37-58`(NewDeployerFromCatalog). **하니스 필터는 이 선례를 따라 FS 래퍼로** 만든다 — deployer 본문에 분기를 넣지 않는다.
- 미러: `WithSkillMirror(false)` 옵션이 이미 존재(`skill_mirror.go:138-140`) — codex 단독에서 링크 생성을 끄는 데 재사용.
- 게시 스킬: `templates/.agents/skills/moai-<command>/SKILL.md` 16종 실파일 — walk가 그대로 배포. `moai-<command>` 이름과 카탈로그 `.claude/skills` 34개 이름의 교집합 0(§4).

### §2.3 update · doctor · tool

- update 재배포: `runCleanReinstall` 경로(`internal/cli/update.go:340-393` 주석 — "runCleanReinstall already invoked deployer.Deploy", `:442`) + `internal/cli/update/deploy/deploy.go:51` CleanMoaiManagedPaths. 하니스 미인지 — codex-only 프로젝트에 전면 재배포하면 `.claude/` 부활.
- update codex 리프레시: `internal/cli/update.go:494-502`(REQ-CW-009, 존재 게이트).
- update 3-way 병합: `RestoreMoaiConfigRetained`(조사 §update 절 — init이 쓴 section yaml 선택 생존). `llm.harness` 생존 경로(AC-IH-015).
- doctor: `internal/cli/doctor/doctor.go`(38.5KB) + `doctor_codex_*` 8파일(미러 드리프트·스테일 스킬·enabled shape·path guard 등). claude 표면 검사와 codex 검사가 섞여 있고 하니스 조건 없음.
- `moai tool enable codex`: `internal/cli/tool.go`(기존 프로젝트 codex 추가, dry-run 프리뷰, system.yaml 존재 요구). t589와의 경계 — 이 SPEC에서 무변경.

### §2.4 위저드

- `agent_wiring` 질문: `internal/cli/wizard/questions.go:375-387` — 3옵션(Value claude/codex/both 동결 대상), 설명이 "연결(wiring)"만 서술하고 배포 결과를 서술하지 않음.
- 번역: `internal/cli/wizard/translations.go:107`(ko)·`:193`(ja)·`:279`(zh) — 갱신 대상 3블록.
- reconfigure: `DefaultQuestions(5) + GitQuestions(7)` — agent_wiring 미포함, 불변.

### §2.5 영속화 후보 판정

- `llm.yaml`(템플릿, plain yaml — `.tmpl` 아님): 단일 루트 키 `llm:`. LLM 설정의 자연 소재이고 update 3-way 병합 대상이라 **`llm.harness`를 여기에 둔다**(design.md D5).
- `harness.yaml`: effort 프로필·평가기 설정(하니스 선택 무관) — 부적합.
- `system.yaml.tmpl`: `moai:` 버전 메타 — 부적합.
- 존재 게이트(`.codex/` 유무로 추론): update의 codex 리프레시가 이미 쓰는 방식이지만 "claude 전용 트리를 사용자가 지운 프로젝트"와 구분 불가 — 명시 키가 맞다.

## §3 표면 소속 지도 (조사 :68-73 → 이 트리 재확인)

| 축 | 경로 | 이 트리 실측 |
|---|---|---|
| 보편 | `templates/AGENTS.md`(17,059B), `.moai/config/sections/*`(33파일 — llm/workflow/quality/harness/…), `.gitignore`, `.git_hooks/`, `.github/`, `.worktreeinclude` | 존재 확인 |
| codex 기착 | `.codex/agents/moai/*.toml`(11), `.codex/hooks.json`+`config.toml`+사이드카(codexwiring — 런타임 생성), `.agents/skills`(게시 16 실파일 + 카탈로그 미러 링크), config.toml 네이티브 statusline(codexwiring/statusline.go) | 존재 확인 |
| claude 전용 | `CLAUDE.md`(19,247B, `@AGENTS.md` 래퍼+claude 설명), `.claude/**`(settings.json.tmpl·agents·commands 16·hooks .sh·output-styles·rules·skills 카탈로그 34), `.claudeignore`, `.mcp.json`, `.moai/status_line.sh(.tmpl)` | 존재 확인 — codex 단독에서 전부 제외(단 skills는 REQ-IH-006 재매핑) |

## §4 충돌·경계 분석

- **이름 충돌**: 카탈로그 34개(`moai`, `moai-domain-*`, `moai-foundation-*`, `moai-harness-learner`, `moai-kanban-foreman`, `moai-meta-harness`, `moai-ref-*`, `moai-workflow-*`) ∩ 게시 16개(`moai-clean` … `moai-todo`) = **0개** (전수 비교, 2026-09-14). codex 단독 재매핑은 오늘 충돌 없음 — 그래도 REQ-IH-007로 skip+report 정책 고정(미래 카탈로그·게시 추가 대비).
- **`.agents/skills` 이중 성격**: 게시(실파일)와 미러(심볼릭 링크)가 같은 디렉터를 공유한다. codex 단독에서는 미러가 소멸하고 게시+재매핑 실디렉터만 남는다 — `moai update`의 R-011 게시 경로 보호(deployer.go:238)는 재매핑 실디렉터와 무관하게 유지된다(게시 경로만 보호).
- **AGENTS.md 공개 갭**: 동결 목록 8개 중 배포본 AGENTS.md에 명시적 근거가 확인된 것 — 질문 채널(표 :23), Task 도구군(표 :24), 훅 6/11(:268-271), 스킬 경로·로딩(:28-33), design-sync(표 :25). 확인 못한 것 — **Agent 소환, 스킬 로더(deferred-m1)의 미제공 명시, output style, 슬래시 명령, Workflow 스크립트**. 이 5개가 REQ-IH-008의 신규 저작 분량이다(조사 :99 "codex 보강 5건 여지"와 부합).
- **바이트 상한**: `internal/config/token_budget_guard.go:94` `CodexContractByteCeiling` 24,576B — 현 17,059B, 여유 ~7.5KB. 전면 저작 4건 + 스킬 로더 1절 포인트 보강 분량으로 충분. (iter1 정정 — 종전 `token_budget_guard.go:40` 표기은 template 패키지 앵커 오인이었고 정본은 `internal/config/` 쪽 상수다.)
- **의미 변경의 기존 사용자 영향**: `llm.harness`가 없는 기존 프로젝트는 claude로 해석 → update가 오늘과 같이 전면 재배포(그들의 `.claude/`는 이미 존재). codex-only 신규 배포는 새 init에만 적용 — 마이그레이션 절차 불필요(§4 제약의 근거).
- **조사 대비 착지 델타**: 조사(09-09) 이후 t583(4문항, 09-12)·t584(자율 모드 재정의)·t523(결속표)·t503(commandemit)·t393(질문+플래그) 착지로 C3 전제 5건 중 4건이 소멸/축소. 남은 본체는 codex 단독 4축 + 영속화 + 문서화.

## §5 미측정 항목 (Gaps)

- `moai init` 실실행 파일집(현재 트리에서 `--llm codex`가 심는 정확한 파일 목록) — 홈 오염 회피로 run 단계 첫 슬롯에서 홈 지문 규약으로 채집. AC-IH-002/003의 RED-now 관측이 이 자리를 채운다.
- doctor가 오늘 codex-only 유사 상태(사용자가 `.claude/`를 손으로 지운 프로젝트)에서 무엇을 보고하는지 — REQ-IH-011의 발견적(finding) 목록은 run 단계에서 doctor golden 스냅샷으로 확정.
- `.github/`·`.git_hooks/`의 codex-only 유지가 조직 템플릿 사용자에게 실제로 쓸모있는지 — 보편(git 인프라)로 분류했으나 사용성 검증은 sync 단계 문서 서술에서 다룬다.
