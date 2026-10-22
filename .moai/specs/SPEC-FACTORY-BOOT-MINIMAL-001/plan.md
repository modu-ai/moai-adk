# SPEC-FACTORY-BOOT-MINIMAL-001 — 구현 계획

> 카드 t1923 · 기준 트리 `59959629908cda723c37e26c7c15718194d6b38a` · Tier M · 개발 방식 TDD(`quality.yaml` `constitution.development_mode: tdd`)

## §A 맥락

- 대상 파일(변경, 이 여섯 개와 AGENTS.md가 전부): `internal/hook/session_start_factory.go`, `internal/hook/session_start_factory_i18n.go`, `internal/hook/role_injection_budget_test.go`, `internal/hook/role_composite_e2e_test.go`, `internal/hook/session_start_guide_i18n_red_test.go`, `internal/hook/session_start_guide_i18n_render_test.go`, `AGENTS.md`. 다시 쓰는 테스트는 REQ-FBM-009가 이름으로 나열한 자동 모드 guide 테스트 7개와 예산 테스트 2개뿐이다(AC-FBM-012가 `internal/hook` 수정 파일 집합을 기계 확인).
- 보존 대상(PRESERVE, 손대지 않고 그대로 통과해야 하는 고정 테스트): `internal/hook/session_start_leader_gate_notice_test.go`(`TestLeaderNoticeBatchGatePointer` — SPEC-AUTONOMY-BATCH-GATE-001 REQ-BGS-016/017), `internal/hook/session_start_factory_test.go`(`TestFactoryLeadNoticeCarriesDispatchDiscipline`, `TestFactoryLeadNoticeDispatchDisciplineKorean` — t85, `TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide`, `TestFactoryWorkerNoticeLocaleWordOrders`), `internal/hook/session_start_factory_worker_test.go`(`TestFactoryGuideTeachesLaneFormsInEveryLocale`), `internal/hook/session_start_factory_rule_test.go`(`TestSD_AC019_NextCardRuleInjection`), `internal/hook/lane_recheck_rule_test.go`, `internal/hook/lane_spawn_authority_test.go`, `internal/hook/session_start_factory_autodispatch_test.go`, `internal/hook/session_start_factory_failfast_test.go`(Codex 레인 노트), `internal/hook/role_core_elements_test.go`(core — REQ-FBM-007), `internal/cli/factory_m6_test.go`(`TestSD_M6_ClearPolicyCarrierConstantOnly` — 두 hook 소스에 env 이름 원문 리터럴이 없음을 스캔), 그리고 `session_start_guide_i18n_render_test.go` 안의 `TestFactoryGuideDocsPointerInBothNotices`·`TestFactoryGuideStaleRunStaysEnglishOnAgentChannel`·`TestCodexLaneBootNoteNamesAutoLane`(같은 파일이지만 다시 쓰지 않는다).
- 보존 대상(PRESERVE, 파일): `.claude/rules/moai/workflow/factory-dispatch.md`, `.claude/rules/moai/workflow/cross-session-messaging.md`, 그리고 두 파일의 `internal/template/templates/` 미러(REQ-FBM-007). 재배치 목적지 3쌍(`gtd.md`, `factory-dispatch-mechanics.md`, `auto-semantics.md`, 배포·템플릿) — 이미 같은 내용을 가지고 있으므로 편집하지 않는다(spec.md §B.3). `internal/hook/role_rules.go`의 상수·로직(`roleRulesContextLimit`, `roleRuleSizeGate`, `buildRoleCore`). `internal/config/token_budget_guard.go`의 `CodexContractByteCeiling`. `internal/template/templates/AGENTS.md.tmpl`.
- Template-First: 이번 변경에는 템플릿 대상 문서 편집이 없다(목적지는 기존 문서). Go 소스(`internal/hook`)는 템플릿 대상이 아니다. AGENTS.md 두 문단은 루트 전용이다. 실행 단계가 목적지 문안을 새로 써야 한다고 판단하면 그것은 R1~R5 밖의 범위 변경이므로 blocker report로 돌려보낸다.

## §B 연구 결과 (plan 단계 실측)

모든 측정은 기준 트리에서 이 세션이 직접 실행했다. 원문 출력은 `probe/`·`probe/research/`에 커밋 대상으로 보관한다(scratch 경로는 인용하지 않는다). Go 측정은 `moai slot acquire --resource go-heavy` 임대 아래, load 27~32에서 실행했다.

### §B.1 측정 도구

- `probe/research/measure_t1923_test.go.txt` — `go test -overlay`로 `internal/hook`에 가상 주입해 실제 렌더러(`factoryLeaderNotice`, `factoryBootstrapNoticeForSource`, `factoryLaneRuleForSource`, `roleRuleInjectionFor`, `assembleInjectionComposite`)의 UTF-16 길이를 출력한다. 트리에 파일을 쓰지 않는다.
- `probe/zz_t1923_boot_probe_test.go.txt` + `probe/overlay.json` — 수용 기준 탐침 두 개: `TestT1923BootBudgetProbe`(AC-FBM-001 — 72칸 크기 + 두 채널 공지 계약 + 변이 자가검사)와 `TestT1923RelocationProbe`(AC-FBM-003 — R1~R5 목적지·제거·경로). overlay 경로는 저장소 루트 기준 상대 경로라 어느 클론에서도 재실행된다.
- `probe/zz_t1923_breach_mutant_test.go.txt` + `probe/breach-overlay.json` — AC-FBM-002의 경계 변이. 영어 auto 레인 startup 칸을 가장 짧은 생산 형태 기준으로 정확히 2,846(상한 + 1)이 되게 레인 규칙을 채운다.
- `probe/agents_md_check.py` — AC-FBM-004. AGENTS.md 크기·두 문단 밖 바이트 동일성·의무 13개를 한 번에 판정한다. 변이 사본 생성기는 `probe/research/mk_agents_mutants.py`.
- `probe/research/zz_t1923_a_sim_test.go.txt` + `sim-overlay.json` / `sim-breach-overlay.json` — 녹색 경로 실현 가능성 측정용 모의(§B.7). 수용 기준 도구가 아니다.

### §B.2 필드별·조합별 크기 (UTF-16, 기준 트리)

명령: `go test -v -overlay .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/research/measure-overlay.json -run '^TestT1923Measure$' -count=1 ./internal/hook/` → exit 0. 원문: `probe/research/measure-5995962990.txt` (0.2.0에서 앵커 패턴으로 재실행 — FIELDS·MEASURE 52줄이 0.1.0 원문과 `diff` 차이 없음, HEAD `b805ac2d24`, 코드는 기준 트리와 동일). 발췌(원문 그대로):

```text
FIELDS lang=en leaderHeader=37 leaderIdentity=319 leaderManual=202 entryGuide=140 agentFanout=60 leaderSocket=17 leaderClasses=375 leaderStagger=444 gateSummary=226 operationalStatus=99 settingsAuto=69 settingsVerify=113 laneJoin=58 laneJoinNoCount=55 laneNextCardRule=1926 laneManualDispatchRule=317 laneRecheckRule=779 laneSpawnAuthority=467 autoModeGuideLeader=464 autoModeGuideLane=457 docsPointer=51
MEASURE role=leader face=deployed lang=en source=startup bootstrap=2573 laneRule=0 roleBlock=4303 composite_factory=6878
MEASURE role=lane face=deployed lang=en source=startup bootstrap=1036 laneRule=2713 roleBlock=4301 composite_factory=8054
MEASURE role=lane face=deployed lang=en source=clear bootstrap=0 laneRule=2713 roleBlock=4301 composite_factory=7016
MEASURE role=lane face=deployed lang=en source=compact bootstrap=0 laneRule=0 roleBlock=4301 composite_factory=4301
MEASURE role=leader face=deployed lang=ko source=startup bootstrap=1650 laneRule=0 roleBlock=4303 composite_factory=5955
MEASURE role=lane face=deployed lang=ko source=startup bootstrap=759 laneRule=2074 roleBlock=4301 composite_factory=7138
MEASURE role=lane face=deployed lang=ja source=startup bootstrap=766 laneRule=2084 roleBlock=4301 composite_factory=7155
MEASURE role=lane face=deployed lang=zh source=startup bootstrap=584 laneRule=1680 roleBlock=4301 composite_factory=6569
```

정리(배포·템플릿 두 면 동일값):

| 조합 | 부팅 공지 | 레인 규칙 | 팩토리 고정분 | 역할 블록 |
|---|---:|---:|---:|---:|
| leader · startup · en / ko / ja / zh | 2,573 / 1,650 / 1,651 / 1,229 | 0 | 같음 | 4,303 |
| lane(auto) · startup · en / ko / ja / zh | 1,036 / 759 / 766 / 584 | 2,713 / 2,074 / 2,084 / 1,680 | 3,751 / 2,835 / 2,852 / 2,266 | 4,301 |
| lane(auto) · clear · en / ko / ja / zh | 0 | 2,713 / 2,074 / 2,084 / 1,680 | 같음 | 4,301 |
| 모든 역할 · compact | 0 | 0 | 0 | 4,301 / 4,303 |

clear·compact에서 부팅 공지 0, compact에서 레인 규칙 0은 소스 게이트(`session_start_factory.go:93-95`, `:128-130`)와 일치한다.

### §B.3 23,145자 넘침의 분해 (카드 관측의 원인)

명령: `python3 -I .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/live_scan.py /Users/goos/.moai/claude-profiles/moai-adk/projects 2026-10-21T00:00:00Z` → exit 0. 원문: `probe/research/live-scan-20261021.txt`. 열 = 시각, 역할, 합본, 역할 블록 앞, 역할 블록, 넘침 여부. 발췌:

```text
2026-10-21T05:41:49.333Z lane 23569 5202 18367 True -Users-goos-MoAI-moai-adk-go--moai-worktrees-t1929/38c8307d-6f95-4adf-
2026-10-21T06:14:54.768Z lane 8068 3775 4293 False -Users-goos-MoAI-moai-adk-go--moai-worktrees-t1737/87d69d3d-984a-4963-
2026-10-21T20:54:29.502Z lane 9071 4778 4293 False -Users-goos-MoAI-moai-adk-go--moai-worktrees-t1909/4a6d0ae6-fa22-4f63-
2026-10-22T18:26:15.251Z leader 5762 1467 4295 False -Users-goos-MoAI-moai-adk-go/3936623d-62a7-4ce8-b44e-74ba0ebe94b8.json
count 15
overflow 10
```

- 23,569 = 넘침 지시문 422 + 결합자 2 + **23,145**(카드 수치와 일치). 역할 블록 18,367은 t1617 압축 전 core(budget 테스트 주석의 「this-tree role block 18,369」와 같은 계열)다.
- 해당 세션의 `cwd`는 primary checkout(`/Users/goos/MoAI/moai-adk-go`)이었고, 현재 그 경로의 `factory-dispatch.md`에는 압축 전 문장(「The leader is the queue's sole producer」)이 0건이다(`grep -c` 0). 즉 10-21 당시 primary checkout이 낡은 규칙을 서빙했다(559커밋 뒤 detach 사고와 같은 날).
- 넘침 10건은 모두 10-21T05:42Z 이전. 이후 5건은 넘침 0, 최대 9,071.

라이브 lane·ko·startup 세그먼트 분해(`probe/research/decompose.py`로 t1909 세션 분해, 세그먼트 머리만 인용): attribution 509 · messaging bound 111 · GLM 359 · 합류 줄 38 · 스폰 권한 347 · 자동 모드 370 · 다음 카드 규칙 1,573 · 정지 복구 499 · binary-lag 421 · guard-liveness 118 · 버전 불일치 안내 411 · 역할 블록 4,293. 비-팩토리 합계 = 509 + 111 + 359 + 421 + 118 + 411 + 결합자 6×2 = **1,941** → 허용분 1,950.

### §B.4 현재 테스트가 위험을 못 보는 이유

명령: `go test -count=1 -run '^(TestRoleInjectionAssemblyBudget|TestCompositeBudgetRealAssembly)$' -v ./internal/hook/` → exit 0. 원문: `probe/red-now-5995962990-intree.txt` (0.2.0 재실행). 발췌:

```text
    role_composite_e2e_test.go:90: RESULT lane_en_glm: composite utf16=9946 (cap 10000, margin=54)
    role_composite_e2e_test.go:90: RESULT lane_ko_glm: composite utf16=9030 (cap 10000, margin=970)
    role_injection_budget_test.go:163: deployed tree: role factory-lane composite=9100 (budget 9146) core=4048 (ceiling 4092, target 3800)
```

`TestRoleInjectionAssemblyBudget`는 생산자를 `strings.Repeat("a", 4797)`로 대체하므로 공지·규칙 텍스트가 커져도 결과가 변하지 않는다. `TestCompositeBudgetRealAssembly`는 실제 렌더를 쓰지만 en·ko 두 언어, startup 한 소스만 보고, 상한이 10,000(여유 판단 없음)이다. 영어 lane이 여유 54자로 통과 중이다.

### §B.5 AGENTS.md 바이트 상한

| 명령 | exit | 출력(원문) |
|---|---|---|
| `go test -count=1 -run '^(TestCodexContractByteCeiling\|TestCodexNestedTemplateDiscoveryBudget)$' -v ./internal/config/` | 0 | `contract document AGENTS.md = 24501 bytes (ceiling 24576, headroom 75)` / `contract document internal/template/templates/AGENTS.md.tmpl = 23784 bytes (ceiling 24576, headroom 792)` / `--- PASS: TestCodexContractByteCeiling (0.04s)` / `--- PASS: TestCodexNestedTemplateDiscoveryBudget (0.10s)` (원문 `probe/red-now-5995962990-ceiling.txt`, 0.2.0 재실행) |
| `git apply --check /tmp/adk-pr-recovery-20261021-xlr15jib/proposal-1862-byte-ceiling/BYTE-CEILING-PROPOSAL.patch` | 1 | `error: patch failed: AGENTS.md:164` / `error: AGENTS.md: patch does not apply` |
| `git cat-file -s 3f81377dc9:AGENTS.md` | 0 | `24807` |
| `grep -n "No main-variant\|WT worktrees live\|Three traceability" internal/template/templates/AGENTS.md.tmpl` | 0 | `169:**Three traceability carriers are then mandatory**, because the branch name no longer identifies` (교체 대상 두 문단은 템플릿에 없음) |

- 상한 상수: `internal/config/token_budget_guard.go:101` `const CodexContractByteCeiling = 24576`. 두 테스트는 `internal/config/token_budget_guard_test.go:84`, `:245`.
- CI FAIL(24,807바이트, main@3f81377dc9) 이후 `a1762cabe1`(#1869)이 AGENTS.md를 24,501로 줄여 기준 트리에서는 녹색이다. 카드의 RED는 현재 트리에서 재현되지 않으므로 해당 기준은 regression-guard로 분류한다(verification-completeness.md §2.1 undecidable disposition).
- 운영자 패치를 기준 트리의 같은 두 문단에 옮긴 후보 측정(scratch 스크립트, 트리 무변경): `head_bytes 24501` / `old_block_bytes 927 new_block_bytes 644` / `candidate_bytes 24218 ceiling 24576 headroom 358`. `only` 복원(+3바이트)을 더해도 24,221.
- 의미 대조(운영자 패치 vs 현재 문단) — 의무 13개, `probe/agents_md_check.py`의 O1~O13과 같은 목록이다: O1 `WT-*main*` 브랜치 금지, O2 `main`을 통합 보유자로 쥐는 워크트리 금지, O3 `main`은 PR 병합으로만 전진, O4 착지마다 local `main`을 `origin/main`과 재동기화, O5 착지 뒤 local `main`이 앞서 있으면 결함, O6 도구는 primary checkout의 HEAD를 detach하지 않음, O7 primary가 점유한 브랜치는 거절, O8 WT 워크트리는 `.moai/worktrees/` 아래에**만** 있음, O9 `moai worktree new`로 생성, O10 `.claude/worktrees/`는 런타임 전용, O11 런타임 `agent-*` 격리 트리, O12 카드 작업에 채택된 런타임 트리는 제자리에서 `WT-`로 개명, O13 그 트리는 생성자의 경로에 그대로 둠. 패치는 O8만 약화(「live **only** under」→「belong under」) → `only` 복원으로 수리. 제거 1건: 2026-10-21 사고 괄호(근거 문장, 의무 아님).
- 후보 측정(`probe/research/mk_agents_mutants.py`로 만든 사본, `probe/research/agents-mutants-5995962990.txt`): 패치 문단 + `live only under` = **24,221바이트**, `agents_md_check.py` exit 0. 변이 4종 — `only` 누락(24,218), main-variant 문단 삭제(23,888), 무관한 의무 문단 삭제(23,872), 두 문단을 제목만 남기고 비움(23,625) — 은 모두 크기 기준을 통과하지만 검사는 각각 O8 MISSING / 앵커 소실 / prefix CHANGED / O1~O13 MISSING으로 exit 1을 낸다.

### §B.6 연결 카드

- t1637 (hold): (d) 「role-core 밖 SessionStart 고정분 4,797자 축소 — 줄어든 만큼 t1617 OUT 행을 IN으로 되돌림」. 본 SPEC은 앞 절반(고정분 축소)을 흡수하고 뒤 절반(core 행 복원, core 상한 개정)은 t1637에 남긴다. 4,797 자체는 숫자로 유지하되 「허용분 1,950 + 결합자 2 + 팩토리 고정분 2,845」로 분해해 실측 가능하게 만든다 — 이후 고정분 실측이 2,845보다 작아진 만큼이 t1637이 청구할 여유다.
- SPEC-AUTONOMY-BATCH-GATE-001(completed): REQ-BGS-016/017이 리더 공지의 배치 게이트 문장을 네 로케일·두 채널에 요구한다. 본 SPEC은 그 문장을 그대로 둔다(decision-index Q8). 0.1.0 초안은 이 문장을 재배치 대상에 넣었었다 — plan-audit 1회차 D1.
- t1780 (picked): 훅 비동기 전환. 출력 텍스트 크기와 무관한 별개 축이라 범위 겹침 없음.
- t1617(SPEC-ROLE-INJECTION-BUDGET-001, completed): 예산 계약 원류. 0.6.0 수치 유지.

### §B.7 녹색 경로 실현 가능성 — 모의 측정

목적: AC-FBM-001의 상한 2,845가 R1~R5만으로 영어 레인 칸에서 달성 가능한지(불가능 방향의 기준이 아닌지), 그리고 AC-FBM-002의 경계 변이가 녹색 경로 트리에서 정확히 2,846을 만드는지 plan 단계에서 관측한다. 모의는 영어 메시지 테이블만 메모리에서 바꾼다(`probe/research/zz_t1923_a_sim_test.go.txt`: R1 자동 모드 해설 → 포인터 한 줄, R2~R4 문장 삭제 + 경로 절 하나, R5 `--card` 절 삭제). 문안은 예시이며 실행 단계가 고른다. 기준 트리(HEAD `b805ac2d24`, Go·AGENTS.md는 `5995962990`과 동일)에서 `go-heavy` 임대 아래 측정.

| 명령 | exit | 원문 발췌(파일) |
|---|---|---|
| `go test -count=1 -overlay .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/research/sim-overlay.json -run '^TestT1923BootBudgetProbe$' -v ./internal/hook/` | 1 (영어 외 칸은 모의하지 않았으므로 ja 레인 2,852 등이 남음) | `PROBE lane//en/startup/deployed fixed=2727 composite=8982 core=4048` / `PROBE leader//en/startup/deployed fixed=2162 composite=8419 core=4048` / `--- PASS: TestT1923BootBudgetProbe/lane//en/startup/deployed (0.37s)` (영어 startup 6칸 모두 공지 계약·크기 PASS) — `probe/research/sim-5995962990.txt` |
| 같은 명령, `sim-breach-overlay.json`(모의 + 경계 변이) | 1 | `PROBE lane//en/startup/deployed fixed=2846 composite=9101 core=4048` / `factory fixed part 2846 > ceiling 2845` (2칸, 두 면) — `probe/research/sim-breach-5995962990.txt` |

결론: 영어 레인 startup 고정분은 R1~R5 예시 문안으로 2,727(여유 118)이다. R5 절은 영어에서 104 UTF-16 단위다. 경계 변이는 녹색 경로 트리에서 정확히 상한 + 1을 만들고, 같은 상한 비교를 하는 탐침이 그것을 잡는다.

## §C 사전 점검 (run 단계 진입 시)

1. `git rev-parse HEAD`가 이 SPEC의 기준 트리 또는 그 후손인지 확인.
2. `go test -count=1 -overlay .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/overlay.json -run '^(TestT1923BootBudgetProbe|TestT1923RelocationProbe)$' -v ./internal/hook/` 가 RED(exit 1)임을 재관측 — run 단계의 RED 증거(E8).
3. 공지·레인 규칙 문구를 고정한 테스트를 먼저 뽑는다. 필드 이름만이 아니라 **렌더된 문자열**을 고정한 테스트까지 잡아야 한다(0.1.0의 필드 이름 grep은 t85 테스트와 배치 게이트 테스트를 놓쳤다 — plan-audit D1):
   - 필드·렌더러 참조: `grep -rln "factoryLeaderNotice\|factoryLaneNotice\|factoryBootstrapNotice\|factoryMessagesFor\|factoryLocales\|factoryLaneRuleForSource\|autoModeGuide\|laneNextCardRule\|laneRecheckRule" internal/hook`
   - 렌더 문자열 표지: `grep -rln -e "batch gate summary" -e "Card routing" -e "routed WHOLE" -e "Stagger" -e "started producing output" -e "Entry points" -e "up to 10 agents" -e "auto-leader" -e "auto-lane" -e "shipped — t1814" -e "adk.mo.ai.kr" -e "operator-authorized" -e "at card takeover" -e "ExitWorktree(keep)" -e "moai-lane-watchdog" internal/hook internal/cli`
   - 두 목록의 테스트 파일이 §A의 「다시 쓰는 9개」와 「보존 대상」 둘 중 하나에 모두 들어가는지 대조한다. 어느 쪽에도 없는 파일이 나오면 진행하지 말고 blocker report.
4. 무거운 `go test` 앞에는 `moai slot acquire --resource go-heavy` 임대, `go test ./...` 금지.

## §D 제약

- role core 무손실: 역할 게이트 규칙 2개와 그 템플릿 미러는 편집 금지(REQ-FBM-007). 재배치는 spec.md §B.3의 R1~R5뿐이고 목적지는 이미 같은 내용을 가진 기존 문서 3쌍이다. 그 밖의 원소를 빼야 하면 실행을 멈추고 blocker report(범위 변경).
- 리더 공지의 유지 원소(spec.md REQ-FBM-001 목록)와 레인 공지의 유지 원소(REQ-FBM-002 목록)는 문구까지 그대로 둔다. 탐침은 유지 원소의 지문을 기준 트리 값과 대조한다.
- 검사 삭제·skip·상한 상향 금지(REQ-FBM-009). 다시 쓰는 테스트는 REQ-FBM-009가 이름으로 나열한 9개뿐이다.
- 프로토콜 토큰 번역 금지 규칙(i18n 파일 머리 주석) 유지 — 명령·경로·도구 이름은 4개 언어 모두 원문.
- AGENTS.md 편집은 상시 로드 파일이므로 작업 마지막에 한 번(cache-aware-execution 지시 3).

## §F 마일스톤 (결정 가능성이 큰 것부터)

### M1 — 공지 계약 결정과 테스트 RED (Priority High)

- 공지 계약(REQ-FBM-001/002)을 테스트로 먼저 고정: REQ-FBM-009가 나열한 자동 모드 guide 테스트 7개를 새 계약(자동 모드 설명 부재 + 재배치 경로를 담은 포인터 줄 하나 + docs 문장 유지)으로 다시 쓴다. 삭제가 아니라 계약 반전이다. 리더 공지의 배치 게이트·라우팅·stagger 테스트는 손대지 않는다(§A 보존 목록).
- 예산 테스트 재도출(REQ-FBM-008): `TestRoleInjectionAssemblyBudget`의 합성 생산자 4,797을 「허용분 1,950 + 결합자 2 + 실제 렌더 고정분」으로 바꾸고 고정분 ≤ 2,845·합본 ≤ 9,146 단언, 역할(리더, 레인 auto, 레인 manual) × 소스 3 × 언어 4 × 면 2 = 72칸 순회, 칸마다 `FIXED role=… dispatch=… lang=… src=… face=… fixed=N composite=N` 한 줄, 끝에 `FIXED swept=72`와 `FIXED accounting: producer 4797 = allowance 1950 + joiner 2 + fixed ceiling 2845` 출력, 위반 칸은 `FIXED BREACH role=… dispatch=… lang=… src=… face=… fixed=N ceiling=2845`(합본이면 `composite=N budget=9146`)를 담은 실패로 보고(AC-FBM-002·011). `TestCompositeBudgetRealAssembly`는 같은 칸을 10,000 전달 한도로 감시하도록 언어·소스를 넓힌다. 헤더 주석에 0.6.0 → 본 SPEC 회계 변경 이력 기록.
- 이 단계 끝에서 새 테스트는 기준 렌더에서 RED여야 한다(E8 증거).

### M2 — 재배치 목적지 확인 (Priority High)

- 목적지는 이미 있다(spec.md §B.3, acceptance.md E10의 기준 트리 `git grep`). 이 단계는 쓰기 없이 `TestT1923RelocationProbe`의 목적지 단언 24개가 HEAD에서 통과함을 확인하는 것으로 끝난다.
- 「제거 전에 목적지에 있다」는 기준 커밋 `5995962990`이 M3 커밋의 조상이라는 사실로 커밋 그래프가 증명한다(verification-claim-integrity.md §2.3). 실행 단계가 목적지에 새 문안을 써야 한다고 판단하면 R1~R5 밖의 범위 변경이므로 멈추고 blocker report — 그 경우의 개정안은 목적지 커밋을 M3 커밋 앞의 별도 커밋으로 두고 `git merge-base --is-ancestor <목적지 커밋> <M3 커밋>` exit 0을 §E.2에 기록하는 AC를 더한다.

### M3 — i18n 메시지 테이블 축소 (Priority High)

- `session_start_factory_i18n.go` 4개 언어에서 자동 모드 해설(R1)을 재배치 경로 포인터 한 줄로 바꾸고, 레인 자동 규칙에서 R2~R5 문장을 빼고 목적지 경로를 담은 짧은 절로 대체(REQ-FBM-006 토큰·문구 보존).
- `session_start_factory.go`의 `factoryLeaderNotice`/`factoryLaneNotice`는 자동 모드 블록 자리에 포인터 줄을 조립한다. 운영자 사본(`factoryBootstrapNotice`)도 같은 렌더러라 자동으로 맞춰진다.
- 목표 크기: 영어 lane startup 고정분 3,751 → ≤ 2,845. 모의 측정(§B.7)은 예시 문안으로 2,727(여유 118)을 보였다. 다른 칸은 이미 여유가 크다.

### M4 — 검증 (Priority Medium)

- 탐침 두 개(AC-FBM-001, 003) 녹색, 예산 테스트 정상 실행(AC-FBM-011) 녹색, 경계 변이에서 예산 테스트가 실패(AC-FBM-002), 보존 테스트 녹색(AC-FBM-009).
- 역할 게이트 파일 diff 0(AC-FBM-007), `internal/hook` 수정 파일 집합이 §A 여섯 개 안(AC-FBM-012).

### M5 — AGENTS.md 문단 교체 (Priority Medium, 마지막)

- §B.5 후보(`probe/research/mk_agents_mutants.py`의 `cand.md`와 같은 문단) 적용 → 24,221바이트, `probe/agents_md_check.py` exit 0(AC-FBM-004), Codex 상한 테스트 녹색(AC-FBM-006).

### M6 — 라이브 확인 (sync 단계, Priority Medium)

- 착지 바이너리 설치 뒤 6레인 + 리더 기동, `probe/live_scan.py`로 합본·넘침·파싱 오류 확인(AC-FBM-008).

## §G 위험

- 레인 규칙 축소 중 REQ-SD-019 등 다른 SPEC이 고정한 토큰을 잘라낼 위험 → REQ-FBM-006과 보존 테스트(AC-FBM-009)가 잡는다.
- 공지가 짧아져 레인이 자동 모드나 착지 절차를 모를 위험 → 같은 내용이 목적지 문서에 이미 있고, 공지·규칙이 그 경로를 이름으로 가리킨다. 착지·판단 절차는 role core(「Before any work, read factory-dispatch.md to the end」)가 읽게 하는 문서 사슬 안에도 있다.
- 영어 레인 여유는 모의 기준 118로 작다. 실행 단계 문안이 길어지면 상한에 닿는다 → AC-FBM-001·011이 칸별 수치로 잡는다. R1~R5 밖의 추가 재배치는 SPEC 개정 대상이다.
- AC-FBM-002의 경계 변이는 레인 자동 규칙이 「다음 카드 규칙 + `\n\n` + 정지 복구 규칙」으로 조립된다는 현재 구조에 기댄다. 실행 단계가 조립 구조를 바꾸면 변이가 2,846보다 크게 채워질 수는 있어도(여전히 위반) 작게 채워지지는 않도록, 변이는 가장 짧은 생산 형태로 계산한다. 조립 구조 변경 시 변이의 채움 계산을 다시 확인한다.
- 비-팩토리 허용분 1,950은 GLM 세션 실측 기준. 새 생산자가 SessionStart에 추가되면 허용분을 넘을 수 있다 → 실조립 테스트(`TestCompositeBudgetRealAssembly`)가 10,000 상한으로 계속 감시한다.
- AGENTS.md 여유 300바이트 목표는 판단값이다(rule-authoring의 1,000바이트 단일 편집 고지 의무보다 작은 완충). 다음 규칙 추가가 여유를 다시 잠식할 수 있다.

## §H 참조

- SPEC-ROLE-INJECTION-BUDGET-001 (예산 계약 원류, 0.6.0)
- SPEC-SESSION-START-GUIDE-I18N-001 (자동 모드 안내·docs 포인터의 기존 계약 — REQ-004/005는 조건부라 문장 대체는 없고, 자동 모드 설명의 존재를 고정한 테스트 7개를 다시 쓴다)
- SPEC-AUTONOMY-BATCH-GATE-001 REQ-BGS-016/017 (리더 공지의 배치 게이트 문장 — 유지)
- SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-019, SPEC-WORKFLOW-TASKS-001, SPEC-FACTORY-LANE-NO-CLEAR-001 (레인 규칙 토큰·문구 — 유지)
- `internal/hook/role_rules.go:52`, `:638` / `internal/hook/session_start.go:600-680` / `internal/config/token_budget_guard.go:101`
