# 리서치 — 인용 재검증과 전제 확인

- 기준 트리: 작업 트리 `WT-review-gate-fix` @ `2aab5f797`.
- 설치 빌드: `moai version` 출력은 `v3.2.0-rc.29   moai_cp/20261007_195927-158-g4f8aba061   built 2026-10-09T16:02:30Z` 다. `git merge-base --is-ancestor HEAD 4f8aba061` 은 참이므로 설치 빌드는 트리보다 새 커밋이다. 이 문서의 측정 증거는 트리 소스에서 만든 `go test` 실행 결과와 파일 읽기다. 설치 빌드의 결과는 트리 증거로 쓰지 않는다.
- 세션 제약: 워크트리 가드가 다른 워크트리(t1543 등)를 대상으로 한 git 호출을 거부했다. 레인 저장소의 git 상태는 관측하지 못했고, 파일 읽기로만 확인했다.

## §1 방법

- 연구 입력 세 건(`.moai/reports/t1620/research-input/lens-1-review-gates.md`, `lens-3-codex-glm-tools.md`, `lens-5-records-statistics.md`)의 `파일:줄` 인용을 스크립트로 추출했다. 스크립트는 임시 디렉터리에서만 실행했고 트리에는 쓰지 않았다. 추출한 인용을 이 트리에서 해석하고, 인용된 줄을 읽어 판정했다.
- 판정 어휘: `확인`(인용 줄이 주장과 맞음), `이동`(주장은 맞으나 파일이나 줄이 다름, 새 위치를 적음), `미발견`(주장의 대상이 없음), `부분`(줄은 맞으나 주장의 일부만 읽음, 읽지 않은 부분을 적음), `관측불가`(이 트리와 이 세션에서 확인할 수 없음, 기계 로컬 데이터나 다른 워크트리).
- 스크립트가 잘못 해석한 인용(같은 이름의 파일이 여럿이거나 괄호 안에서 파일 이름이 생략된 경우)은 주장의 문맥으로 파일을 바로잡았고, 그 사실을 "파일 정정"으로 표시했다.
- 연구 입력의 절대 경로 인용은 모두 프로젝트 기준 상대 경로로 바꿨다.
- 건수: lens-1 약 62, lens-3 약 110(중복 포함), lens-5 코드 인용 약 12. 말뭉치 수치(lens-5 §A)는 기계 로컬 데이터이므로 `관측불가`로 묶었다.

## §2 카드가 적은 인용 경로

| 카드 인용 | 이 트리의 실제 위치 | 판정 | 근거 |
|---|---|---|---|
| `internal/hook/codex_review_delay.go:169-183` | `internal/cli/codex_review_delay.go:169-183` | 이동(디렉터리) | `git ls-files internal/hook` 에 `codex_review*` 원본이 없다. 있는 것은 `review_gate_directory_test.go`, `review_gate_resolver_test.go`, `review_gate_selfgate_test.go` 세 테스트 파일뿐이다 |
| `internal/hook/codex_review_cache.go:29` | `internal/cli/codex_review_cache.go:29` | 이동 | 29행: `const codexReviewCacheTTL = 24 * time.Hour` |
| `internal/hook/codex_review_scope.go:189-212` | `internal/cli/codex_review_scope.go:189-212` | 이동 | 189행: `func cardMergeBase(dir string) (string, error) {` |
| `internal/hook/mcp_review_material.go:156-186` | `internal/cli/mcp_review_material.go:156-186` | 이동 | 156행: `func resolveReviewBase(root string) (reviewBase, error) {` |

## §3 인용 등록부 (연구 입력 전수)

표기: 출처 `L1`(lens-1), `L3`(lens-3), `L5`(lens-5). "줄" 열은 이 트리에서 읽은 줄의 요지다.

### §3.1 lens-1 (턴 종료·진입·영수증·범위)

| # | 주장 (요지) | 인용 | 이 트리의 줄 (요지) | 판정 |
|---|---|---|---|---|
| 1 | 턴 종료 핸들러 본체 | `codex_review_gate.go:104-211` | 104행 `func HandleCodexReviewGate(...)` | 확인(시작 줄, 끝 211행은 읽지 않음) |
| 2 | 비활성화면 허용 | `:106` | `if !enabled {` | 확인 |
| 3 | stop_hook_active 면 허용 | `:109` | `if input != nil && input.StopHookActive {` | 확인 |
| 4 | stale 바이너리 건너뜀 | `:130` | `if staleBinarySkipApplies(scope.Dir) {` | 확인 |
| 5 | 트리·주 체크아웃 범위 건너뜀 | `:136` | `if treeScopeSkipApplies(scope, ...) {` | 확인 |
| 6 | 검토할 변경 없으면 허용 | `:139` | `if !reviewGateScopedChangeDetector(scope) {` | 확인 |
| 7 | codex 가 PATH 에 없으면 허용 | `:143-146` | 143행 `binaryPath, err := codexLookPath(codexBinaryName)` | 확인 |
| 8 | 현재 키의 수용 영수증 | `:159` | `if chk := verify.CheckReceipt(...); chk.Run {` | 확인 |
| 9 | fail 이면 상세와 함께 차단 | `:161-174` | 161행 `if isBlockVerdict(chk.Receipt.Verdict) {` | 확인 |
| 10 | 그 밖의 판정은 허용 | `:176-178` | 176행: "reusing the cached %s verdict" | 확인 |
| 11 | 진행 중 kick 표식이면 허용 | `:197-201` | 197행 `inFlight, markerPath := kickInFlight(...)`, 198행 `if inFlight {` | 확인 |
| 12 | kick 을 걸고 허용 | `:202-210` | 202행 `if kickErr := codexReviewBackgroundKick(scope.Dir); kickErr != nil {` | 확인 |
| 13 | 진입 핸들러 본체 | `codex_review_delay.go:140-190` | 140행 `func HandleCodexReviewEntry(...)` | 확인(시작 줄) |
| 14 | 수용된 fail 이면 진입에서 차단 | `codex_review_delay.go:169-183` | 169행 `if chk := verify.CheckReceipt(...); chk.Run {`, 170행 `if isBlockVerdict(...)` | 확인 |
| 15 | 영수증 상태 표 | `internal/verify/receipt.go:63-96` | 63행 `func CheckReceipt(...)`, 96행 `}` | 확인 |
| 16 | 유효기간 24시간 | `codex_review_cache.go:29` | 29행 `const codexReviewCacheTTL = 24 * time.Hour` | 확인 |
| 17 | UserPromptSubmit 은 `decision:"block"` 과 exit 2 로 차단 | `hooks-system.md:91` (`.claude/rules/moai/core/hooks-system.md`) | 91행 UserPromptSubmit 행: "Exit 2 blocks prompt; JSON `decision:"block"` also available" | 확인 |
| 18 | Stop 의 stdout 열은 `systemMessage` 만 | `hooks-system.md:103` | 103행 Stop 행: `systemMessage` | 확인 |
| 19 | 공식 훅 문서에 두 이벤트의 차단 의미 문장이 없다 | `claude-code-hooks-official.md` (파일 전체) | UserPromptSubmit 또는 Stop 과 block 을 함께 쓴 줄이 없다. block 이 나오는 줄은 32, 37, 222, 257, 312행으로 도구 훅과 스키마 항목이다 | 확인(부정 주장) |
| 20 | "Original prompt" 문구가 저장소에 없다 | (저장소 전체) | 연구 입력 lens-1 보고서 자체만 이 문구를 인용한다 | 확인(부정 주장) |
| 21 | 도구 트리 범위 판정이 트리 클래스에 따라 달라진다 | `codex_review_tree_scope.go:133` | 133행 `if scope.Class != reviewScopeTree || cardScopeFromBranch(scope.Branch) {` | 확인 |
| 22 | 실시간 재현(표준입력 파이프 + 설치 빌드) | `printf ... | moai hook codex-review-entry` | 해당 명령은 재실행하지 않았다(표준입력 파이프는 단일 호출 형식 밖이고, 설치 빌드는 트리와 다르다). 대신 레인 저장소의 스냅샷을 읽었다(§4) | 부분 |
| 23 | 15분 임대 상수 | `internal/config/defaults.go:65` | 65행 `DefaultFactoryLeaseDuration = 15 * time.Minute` | 확인(값). 임대와의 상호작용은 추론이라 관측불가 |
| 24 | cardMergeBase 가 IntegrationTarget 을 읽는다 | `codex_review_scope.go:189-212` | 189행 `func cardMergeBase`, 190행 `cfg := config.LoadGitFlowIntegrationConfig(dir)`, 191행 `target := strings.TrimSpace(cfg.IntegrationTarget)` | 확인 |
| 25 | 두 키의 값: main 과 develop | `git-strategy.yaml` (`.moai/config/sections/`) | 5행 `worktree_base_branch: main`, 16행 `develop_branch: develop` | 확인 |
| 26 | merge-base(main, develop) = db0c514d3, main 139 커밋 | (git 위상) | `merge-base main develop` = `db0c514d3b0ce16befa5a6d35126a35b6b0f73d0`. `rev-list --count db0c514d3..main` = 139. `..develop` = 149 | 확인 |
| 27 | 카드 HEAD 에 대해 `git diff --name-only db0c514d3` 가 168 파일 | (레인 t1543 저장소) | git 가드가 다른 워크트리 대상 호출을 거부했다 | 관측불가 |
| 28 | t1543 과 t1614 가 같은 키 `2aab5f797…-7f1e2657f0831d9f` | (레인 저장소) | t1543 의 스냅샷 키가 `2aab5f797b75…-7f1e2657f0831d9f` 로 일치한다. t1614 는 읽지 않았다 | 부분 |
| 29 | 발견 사항 두 건이 카드가 건드리지 않은 파일을 가리킨다 | `internal/homestate/card_transition.go:287` 과 `leader_approval.go:209` | 287행 `if err := req.BeforeCommit(c); err != nil {`, `internal/homestate/leader_approval.go:209` 는 `leader_approvals` 조회 SQL | 부분(줄은 있으나 "카드가 건드리지 않았다"는 주장은 관측불가) |
| 30 | resolveReviewBase 가 worktree_base_branch 를 먼저 본다 | `mcp_review_material.go:156-161` | 156행 `func resolveReviewBase`, 159행 `if name := config.LoadWorktreeBaseBranch(root); name != "" {` | 확인 |
| 31 | 코드는 origin/<target> 만 처리한다 | `:204-210` (파일 정정: `codex_review_scope.go`) | 204행 `remote, remoteErr := runReviewGit(dir, "merge-base", "refs/remotes/origin/"+target, "HEAD")`, 207행 `--is-ancestor` 검사 | 확인(파일 정정) |
| 32 | 두 테스트가 기준 동작을 고정 | `TestCardMergeBaseConfiguredMain`, `TestCodexReviewScope_AbsorbedDevelopRecomputesBase` | 두 테스트 모두 존재한다. 첫 번째는 셀 B-1 에서 PASS | 확인 |
| 33 | 턴 종료 재사용 분기가 inconclusive 를 허용 | `codex_review_gate.go:159-178` | 159행 `if chk := ...; chk.Run {`, 160행 `codexReviewCacheSkips.Add(1)` | 확인 |
| 34 | 생산자는 RPC 오류와 타임아웃, 미고정 출력을 inconclusive 로 기록 | `codex_review_receipt.go:160-186` | 160행 `case rpcErr != nil:`, 182행 `case !strings.HasPrefix(... codexReviewVerdictPass):` | 확인 |
| 35 | 래퍼가 stderr 를 리다이렉트하지 않는다 | `handle-codex-review-gate.sh` (마지막 줄) | 마지막 호출은 `printf '%s' "$INPUT" \| "$MOAI_BIN" hook codex-review-gate` 이고 `2>` 가 없다. 다만 Claude Code 가 그 stderr 를 버리는지는 관측하지 못했다 | 확인(래퍼) / 관측불가(버림) |
| 36 | 집계 48 inconclusive, 48 fail, 21 pass (117) | (레인 저장소 31개) | 기계 로컬 데이터 | 관측불가 |
| 37 | 키는 재검토 전에 측정된다 | `codex_review_receipt.go:149` | 149행 `state, err := codexReviewReceiptStateForScope(ctx, scope, binaryPath)` | 확인 |
| 38 | 주석이 "키 측정 후 검토" 이 중 상태를 설명한다 | `codex_review_receipt.go:121-123` | 121행 "REQ-CGS-009) and records the verdict. The key is measured before the review, ..." | 확인 |
| 39 | Codex 정지 체인에 연속 상한이 있다 | `codex_stop_chain.go:709` | 709행 `// ── §D3.8 consecutive-unmeasured cap (Codex-only)`. 상한은 측정되지 않은 결과의 연속 횟수이며 차단 횟수가 아니다. 값 3 은 `internal/codexwiring/stop_budget.go:153` | 부분(특성 정정: 차단이 아니라 미측정 결과의 상한) |
| 40 | Head 가 결합 필드다 | `receipt.go:71` (파일 정정: `internal/verify/receipt.go`) | 71행 `{"head", stored.Head, current.Head},` | 확인 |
| 41 | `--allow-empty` 커밋이 한 턴 우회 | (미실행) | 실행하지 않았다 | 관측불가 |
| 42 | 같은 digest 에 다른 HEAD 에서 두 번 검토 (t1595, t1613, t1585) | (레인 저장소) | 기계 로컬 데이터 | 관측불가 |
| 43 | stale 바이너리 건너뜀의 조건 | `codex_review_binary_age.go:56-68` | 56행 `func staleBinarySkipApplies(...)`. 62–67행: `verdict.Status != binlag.StatusBehind` 면 거짓, 아니면 건너뜀 | 부분(건너뜀 조건은 `StatusBehind` 이며, 그 상태의 조상 판정은 `internal/binlag` 에 있어 읽지 않았다) |
| 44 | 설치 빌드 4f8aba061 은 WT-web-haiku55-effort-agent-dedupe 에만 있다 | (브랜치 포함 질의) | 4f8aba061 이 HEAD 의 후손이라는 사실은 확인(§5 P-16). 브랜치 포함 질의는 실행하지 않았다 | 부분 |
| 45 | 진입 훅 timeout 5 | `.claude/settings.json` 과 `settings.json.tmpl:272` | `.claude/settings.json` 272행 `"timeout": 5,`(진입 훅 블록), 템플릿 272행 `"timeout": 5,` | 확인 |
| 46 | 버전 탐색이 `context.Background()` 를 쓴다 | `codex_review_delay.go:165` | 165행 `state, stateErr := codexReviewReceiptStateForScope(context.Background(), scope, binaryPath)` | 확인 |
| 47 | 측정 1.53 s 와 1.90 s, 부하 18 | (측정 기록) | 재측정하지 않았다 | 관측불가 |
| 48 | 형제 훅을 2026-10-09 에 5 s 에서 15 s 로 올렸다 | `hooks-system.md:368` | 368행은 timeout 표의 UserPromptSubmit 행 `5s`, `30s`. 2026-10-09 의 상향 기록은 `.claude/rules/moai/core/hooks-system.md` 와 `.moai/docs/` 에 없다 | 미발견 |
| 49 | 래퍼의 스위치 읽기 경로 | `handle-codex-review-entry.sh` (`CODEX_GATE_PROJECT_ROOT`) | 33행 `CODEX_GATE_PROJECT_ROOT="${CLAUDE_PROJECT_DIR:-$PWD}"`, 34행 `CODEX_GATE_WORKFLOW_YAML="$CODEX_GATE_PROJECT_ROOT/.moai/config/sections/workflow.yaml"` | 확인 |
| 50 | Go 층은 주 체크아웃 대신 저장 루트의 설정을 읽는다 | `codex_review_gate.go:469,489-495` | 469행 `enabled := readCodexReviewGateEnabled(reviewGateConfigRoot(projectDir))`. 489–495행은 `auditreceipt.StoreRoot` 호출 | 확인 |
| 51 | 저장 루트는 트리에 workflow.yaml 이 있으면 그 트리 자신이다 | `storeroot.go:243` (`internal/auditreceipt/`) | 243행 `if !IsConfigOrphanedRoot(root) {`. `IsConfigOrphanedRoot`(179–185행)는 `.moai/config/sections/workflow.yaml` 이 있으면 거짓을 돌려준다 | 확인 |
| 52 | 추적 파일이므로 각 레인은 자기 사본을 읽는다 | `workflow.yaml` | `git diff` 가 이 파일의 추적 상태와 한 줄 변경을 보여 준다 | 확인 |
| 53 | 발견 사항 항목이 Finding{Title: msg, Body: msg} 로 중복된다 | `mcp_codex.go:1838` | 1838행 `findings = append(findings, Finding{Severity: sev, Title: msg, Body: msg})` | 확인 |
| 54 | 요약과 발견 사항을 모두 출력한다 | `codex_review_cache.go:53-58` | 53행 `fmt.Fprintf(&b, "# codex review fail — %s\n\n%s\n\n## Findings\n\n", ...)`, 이어서 발견 사항 줄 형식 | 확인 |
| 55 | (:0) 항목과 절대 경로가 기계 로컬 파일에 있다 | (기계 로컬 .md) | 관측하지 않았다 | 관측불가 |
| 56 | `normalizeFindingPath` 가 있으나 여기에 적용되지 않는다 | (코드) | 확인하지 않았다 | 관측불가 |
| 57 | 두 리뷰 계층이 같은 카드 diff 를 두 번 검토한다 | `factory-dispatch-detail.md:157` | 157행 "**One review per diff.** ... enable one of the two." | 확인 |
| 58 | card-review 단계가 필수다 | `factory-dispatch.md:136` (파일 정정: 실제는 132행) | 136행은 빈 줄이다. 필수 문장은 132행: "[HARD] ... every lane runs a `card-review` stage" | 이동(132행) |
| 59 | workflow.yaml:301-302 의 `enabled: true` | `workflow.yaml:301-302` | HEAD 에서 301행 `review_gate:`, 302행 `enabled: true`. 작업 트리에서는 302행이 `enabled: false` 다(미커밋 변경, `git diff` 한 줄) | 확인(HEAD). 작업 트리는 다름 |
| 60 | 리뷰 범위 해석 호출 | `mcp_selfreview.go:119` | 119행 `rs = reviewScopeResolver(root)` | 확인 |
| 61 | 버킷이 15분 단위로 정렬된다 | `codex_review_delay.go:77` | 77행 `bucket := now.Unix() / int64(config.DefaultCodexReviewGateTimeout.Seconds())`. 값 900 s 는 `internal/config/defaults.go:807` | 확인 |
| 62 | 직전 버킷의 마커만 수집한다 | `codex_review_delay.go:98-100` | 98행 `prev := codexReviewKickMarkerPath(...)`, 99행 `now.Add(-config.DefaultCodexReviewGateTimeout))` | 확인 |
| 63 | 자식 프로세스에 Setsid/Setpgid 가 없다 | `codex_review_delay.go:52-62` | 52행 `cmd := exec.Command(self, "verify", "codex-review", "--project-root", dir)`. 파일 전체에 `Setsid` 와 `Setpgid` 가 없다 | 확인 |
| 64 | 마커 64개가 남아 있다 | (기계 로컬) | 관측하지 않았다 | 관측불가 |
| 65 | 추적 파일이 없어도 untracked 를 센다 | `codex_review_scope.go:310` | 310행 `untracked, err := reviewScopeGit(scope.Dir, "ls-files", "--others", "--exclude-standard")` | 확인 |
| 66 | t1582 가 같은 HEAD 에서 여섯 번 kick | (레인 저장소) | 관측불가 | 관측불가 |
| 67 | 카드 범위가 git 루트에 고정되지 않는다 | `codex_review_gate.go:123-125` | 123행 `if scope.Class == reviewScopeTree {`, 124행 `scope.Dir = reviewExclusionRoot(scope.Dir)` | 확인(트리 클래스에만 고정됨) |
| 68 | `.moai/logs/` 규칙은 루트 고정, `**/.moai/state/` 는 재귀 | `.gitignore` | 312행 `**/.moai/state/`, 400행 `.moai/logs/`(중간 구분자가 있어 루트 고정), 401행 `.moai/state/` | 확인 |
| 69 | 하위 디렉터리 테스트는 트리 클래스를 단언한다 | `TestReviewEntry_SubdirSessionReadsRootReceipt` 등 | 테스트 이름은 존재한다. 단언 본문은 읽지 않았다 | 부분 |
| 70 | 추적 diff 전체를 해시한다 | `codex_review_scope.go:346` | 346행 `diff, err := reviewScopeGit(scope.Dir, "diff", scope.MergeBase)` | 확인 |
| 71 | 런타임 접두는 untracked 쪽에서만 걸러진다 | `codex_review_scope.go:362` | 362행 `if name == "" \|\| isRuntimeManagedPath(name) {` | 확인 |
| 72 | 추적 런타임 파일이 1,949 개다 | `git ls-files .moai/harness .moai/reports` | 실제 개수 1948 | 이동(개수 차이: 1948) |
| 73 | 제외 pathspec 은 selfReviewMaterial 이 이미 쓴다 | `mcp_selfreview.go:168` | 168행 `for _, p := range reviewGateRuntimePrefixes {`, 169행 `":(exclude)"+p` | 확인 |
| 74 | FIFO 나 소켓이 키 계산을 실패시킨다 | `codex_review_scope.go:397` (파일 정정: `:397` 은 `codex_review_scope.go`) | 397행 `return fmt.Errorf("unsupported file type at %q", path)` | 확인(파일 정정) |
| 75 | `--version` 실패면 tool_version 이 비어 영수증이 받아들여지지 않는다 | (영수증 상태 계산) | 상태 계산 본문은 읽지 않았다 | 부분 |
| 76 | 저장 루트를 풀 수 없으면 비활성으로 읽힌다 | `codex_review_gate.go:491-493` | 491행 `if err != nil {`, 492행 `return ""`. `readCodexReviewGateEnabled("")` 는 `internal/cli/mcp_codex.go:2503-2505` 에서 거짓을 돌려준다 | 확인(연쇄) |
| 77 | 잘못된 표준입력 JSON 이면 기본 출력 | `codex_review_gate.go:463-467` | 464행 `fmt.Fprintf(cmd.ErrOrStderr(), "codex-review-gate: invalid stdin JSON (%v); emitting default output\n", err)` | 확인 |
| 78 | 변경 감지 중 git 오류는 허용 | `codex_review_gate.go:231` | `hasReviewableChanges`(226–233행)는 오류 시 거짓을 돌려준다 | 확인 |
| 79 | 같은 취지의 카드 범위 감지 | `scope.go:329` (파일 정정: `internal/cli/codex_review_scope.go:329`) | 329행 `if err != nil {`, 함수 `hasReviewableCardChanges` 는 오류 시 거짓 | 확인(파일 정정) |
| 80 | codex 가 없으면 허용 | `codex_review_gate.go:143-146` | 143행 `binaryPath, err := codexLookPath(codexBinaryName)` | 확인 |
| 81 | kick 실패면 허용 | `codex_review_gate.go:202-208` | 202행 `if kickErr := codexReviewBackgroundKick(scope.Dir); kickErr != nil {` | 확인 |
| 82 | 마커가 있으면 프로세스가 죽었어도 허용 | `codex_review_gate.go:198-201` | 198행 `if inFlight {` | 확인 |
| 83 | 상세 파일을 못 읽으면 "fail" 만 남는다 | `cache.go:72-76` (파일 정정: `internal/cli/codex_review_cache.go:72-76`) | `codexReviewCachedDetail` 의 72–76행, 읽기 오류면 빈 문자열을 돌려준다 | 확인(파일 정정) |
| 84 | 트리 범위 fail 이 설정 전용 발견 사항이면 통과로 재분류된다 | `receipt.go:176-178` | `internal/verify/receipt.go` 는 167행뿐이다. 실제 재분류는 `internal/cli/codex_review_receipt.go:176`: `if targets, ok := runtimeConfigOnlyFindings(out.Findings, scope.Dir); ok && scope.Class == reviewScopeTree {` | 이동(파일: `internal/cli/codex_review_receipt.go:176`) |
| 85 | 주 체크아웃은 기본적으로 건너뛴다 | `tree_scope.go:137` (파일 정정: `internal/cli/codex_review_tree_scope.go:137`) | 137행 `if scope.Primary && primaryScopeReader(root) != config.CodexReviewGatePrimaryScopeReview {`. 기본값은 읽지 않았다 | 이동(파일) / 부분(기본값) |
| 86 | 리뷰 diff 는 200,000 바이트에서 잘린다 | `mcp_review_material.go:35` | 35행 `const reviewDiffMaxBytes = 200_000` | 확인 |
| 87 | 턴 종료 결과는 세션 키로 묶인다 | `multi_review_gate.go:79-98` | 79행 `results := loadConvergenceResults(projectDir, sessionID)`, 80행 `if len(results) == 0 {` | 확인 |
| 88 | 다중 게이트 자체 점검은 커밋되지 않은 변경만 본다 | `multi_review_gate.go:75` | 75행 `if !reviewGateChangeDetector(projectDir) {`. 감지기는 `codex_review_gate.go:76` 의 `hasReviewableChanges`(`git status --porcelain`) | 확인 |
| 89 | Codex 하네스 소비자는 10분 TTL, ttl 0 | `codex_stop_chain.go:659` (`:660` 이 맞다) | 659행은 `}`. 660행 `chk := verify.CheckReceipt(verify.LoadReceipt(c.root, state), state, time.Now(), 0)`. ttl 0 은 기본 10분(`internal/verify/receipt.go:90`, `internal/verify/freshness.go:7`) | 이동(+1행) / 확인(실질) |
| 90 | fail 시 전체 재실행을 지시한다 | `codex_stop_chain.go:673-676` | 673행 `Reason: fmt.Sprintf("codex review gate: the codex review recorded for this tree failed. Run ...`) | 확인 |
| 91 | 테스트 이름 목록 (약 40개) | 연구 입력 §5 | 모든 이름이 `internal/cli` 또는 `internal/runtime` 에 존재한다(§6) | 확인 |
| 92 | codex_review_delay_test.go 에 StopHookActive 참조가 없다 | `codex_review_delay_test.go` | `grep -c StopHookActive` = 0 | 확인 |
| 93 | 결함 테스트가 허용과 exit 코드 ≠ 0 을 단언한다 | `TestReviewGate_InconclusiveReceiptIsNotPassEvidence` | 실행 결과 PASS 이고 stderr 에 "reusing the cached inconclusive verdict" 가 나온다(셀 B-2). 단언 본문은 읽지 않았다 | 부분 |

### §3.2 lens-3 (모델 도구·감사·잡)

| # | 주장 (요지) | 인용 | 이 트리의 줄 (요지) | 판정 |
|---|---|---|---|---|
| 94 | 기본 타임아웃 600 s | `I/config/defaults.go:888` (`internal/config/defaults.go`) | 888행 `var DefaultCodexTaskTimeout = 600 * time.Second` | 확인 |
| 95 | 전경 턴에 적용 | `C/codex_task.go:172` (`internal/cli/`) | 172행 `ctx, cancel := context.WithTimeout(parent, config.DefaultCodexTaskTimeout)` | 확인 |
| 96 | 백그라운드 세션 전체(핸드셰이크 포함)에 적용 | `:460-461` | 460행 `sessionCtx, cancelSession = context.WithTimeoutCause(`, 461행 `context.WithoutCancel(ctx), config.DefaultCodexTaskTimeout, errCodexTaskSessionDeadline)` | 확인 |
| 97 | stdin EOF 가 r.cancel() 을 호출 | `C/mcp_server.go:144-161` | 144행 `func (r *mcpEOFReader) Read(...)`, 147행 `r.cancel()` | 확인 |
| 98 | EOF 는 정상 종료로 0 을 낸다 | `:129-131` | 130행 `return nil // EOF is normal MCP shutdown, not a command failure.` | 확인 |
| 99 | 정리 스크럽은 거부 목록 방식 | `C/mcp_claude.go:266-282` | 266행 `func scrubClaudeAuditEnv(environ []string) []string {` | 확인 |
| 100 | 429 를 구독 용량으로 표시 | `:151-158` | 151행 `if claudeAuditAPIStatus(stdout) == 429 {` | 확인 |
| 101 | 버전 점검 | `C/doctor_mcp_version.go:39-95` | 39행 `func checkMCPServerVersion(...)` | 확인 |
| 102 | 프로젝트 루트 해석 | `C/mcp_project_root.go:104-109` | 104행 `func resolveToolProjectRootWithSource(...)` | 확인 |
| 103 | CLAUDE_PROJECT_DIR 대체 | `C/session.go:316` | 316행 `if dir := os.Getenv("CLAUDE_PROJECT_DIR"); dir != "" {` | 확인 |
| 104 | 서버 기동 시 고정 경고는 rootProvenanceMap 에만 있다 | `:122-131` (파일 정정: `C/mcp_project_root.go`) | 122행 `func rootProvenanceMap(root, source string) map[string]any {` | 확인(파일 정정) |
| 105 | 하위 프로세스 stderr 는 버려진다 | `C/mcp_codex.go:542` | 542행 `cmd.Stderr = &bytes.Buffer{}` | 확인 |
| 106 | applyGateUnmet 가 fail 을 만든다 | `:2084-2114` | 2084행 `func applyGateUnmet(out ReviewOutput, projectDir string) ReviewOutput {` | 확인 |
| 107 | 허브 훅의 메시지 청구 30 s 임대 | `I/hook/factory_messages.go:268` | 268행 `claims, err := s.Claim(ctx, p, factorymsg.MaxBatch, 30*time.Second)` | 확인 |
| 108 | 청구마다 토큰이 바뀐다 | `I/factorymsg/store.go:969-970` | 969행 `c.ClaimToken = newID()`, 970행 `UPDATE messages SET state='claimed',claim_token=?...` | 확인 |
| 109 | 목록이 messages 를 null 로 직렬화한다 | `C/mcp_factory_msg.go:88` | 88행 `return toolJSON("factory_msg_list", map[string]any{"messages": claims, ...}), nil` | 확인 |
| 110 | 해당 없음이면 nil, nil 을 돌려준다 | `:935-937` (파일 정정: `internal/factorymsg/store.go`) | 935행 `if eligible == 0 {`, 936행 `return nil, nil` | 확인(파일 정정) |
| 111 | 토큰 불일치는 claimMismatch | `I/factorymsg/handoff_bind.go:445` | 445행 `return errors.New("claim identity mismatch")` | 확인 |
| 112 | 비 2xx 응답이 "z.ai returned HTTP %d" 로 접힌다 | `C/mcp_glm.go:356-360` | 356행 `if resp.StatusCode < 200 \|\| resp.StatusCode >= 300 {` | 확인 |
| 113 | 같은 접힘이 glm_task 에도 있다 | `C/glm_task.go:338-341` | 338행 동일 조건 | 확인 |
| 114 | 턴 종료 게이트와 scope card 가 같은 해석기를 쓴다 | `C/mcp_selfreview.go:119` | 119행 `rs = reviewScopeResolver(root)` | 확인 |
| 115 | cardMergeBase 가 develop 을 읽는다 | `C/codex_review_scope.go:189-212` | 189행 `func cardMergeBase`, 190행 `cfg := config.LoadGitFlowIntegrationConfig(dir)` | 확인 |
| 116 | resolveReviewBase 가 worktree_base_branch 를 먼저 읽는다 | `C/mcp_review_material.go:156-186` | 156행 `func resolveReviewBase` | 확인 |
| 117 | 같은 읽기 | `:159` | 159행 `if name := config.LoadWorktreeBaseBranch(root); name != "" {` | 확인 |
| 118 | 버전 커밋을 차이 기준으로 쓰지 않는다 | `C/codex_review_binary_age.go:45` | 45행 `return version.Commit, version.Version` | 확인(사용처가 점검용임) |
| 119 | 빌드 식별 점검 | `C/mcp_build_identity.go:59` | 59행 `rawCommit := version.GetCommit()` | 확인 |
| 120 | 두 해석기가 다른 키를 읽는다 | `C/codex_review_scope.go:190` 과 `C/mcp_review_material.go:159` | 190행과 159행 | 확인 |
| 121 | 적대적 리뷰 프롬프트가 "proposed change" 를 말한다 | `C/mcp_convergence.go:643` 과 `C/mcp_codex.go:1175-1200`, `:1954` | 643행 `out, _ := codexReviewRPC(legCtx, binaryPath, codexMethodTurnStart, params)`, 1175행 `case codexMethodTurnStart:`, 1954행 `prompt := "Perform an adversarial code review of the proposed change. " +` | 확인 |
| 122 | 감사 다중 기록이 Target 과 BaseSHA, DiffSHA256 을 쓴다 | `C/mcp_audit_multi_record.go:61,93-101` | 61행 `Target: target,`, 93–101행 `rec.Scope.BaseSHA`, `DiffSHA256` 계산 | 확인 |
| 123 | 수신 영수증 검사는 gate_unmet 과 inconclusive 만 거부한다 | `I/auditreceipt/store.go:1265-1279` | 1265–1266행 `case strings.TrimSpace(r.GateUnmet) != "", strings.TrimSpace(r.CodexVerdict) == codexVerdictInconclusive:`. 1267–1279행(fail 이 통과로 되는 경로)은 읽지 않았다 | 부분 |
| 124 | 영수증 구조체에 HEAD 등이 없다 | `Receipt (:143-151)` | 143–151행: receipt_id, tool, tree_root, root_source, created_at, codex_verdict, gate_unmet | 확인 |
| 125 | 생산자가 RPC 오류를 inconclusive 로 기록 | `C/codex_review_receipt.go:160-164` | 160행 `case rpcErr != nil:`, 161행 "Inconclusive records a non-zero exit (2)" | 확인 |
| 126 | 미인식 판정 본문도 inconclusive | `:182-185` | 182행 `case !strings.HasPrefix(strings.ToLower(strings.TrimSpace(out.Verdict)), codexReviewVerdictPass):` | 확인 |
| 127 | 턴 종료 게이트가 재사용 검사를 한다 | `C/codex_review_gate.go:157-179` | 157행 `state, stateErr := codexReviewReceiptStateForScope(...)`, 158행 `if stateErr == nil {` | 확인 |
| 128 | 24 시간 유효기간 | `C/codex_review_cache.go:29` | 29행 | 확인 |
| 129 | 조상 검사 kick 은 stale 로 건너뜀 | `C/codex_review_gate.go:130` | 130행 `if staleBinarySkipApplies(scope.Dir) {` | 확인 |
| 130 | Stop 은 항상 허용하고 kick 한다 | `C/codex_review_gate.go:197-210` | 197행 `inFlight, markerPath := kickInFlight(...)`, 198행 `if inFlight {` | 확인 |
| 131 | 진입 훅은 현재 키의 fail 에서만 차단 | `C/codex_review_delay.go:165-190` | 165행 `state, stateErr := codexReviewReceiptStateForScope(context.Background(), ...)` | 확인 |
| 132 | 다중 게이트 결과 기록은 평문 쓰기 | `C/mcp_convergence.go:1069-1092` | 1069행 `stateDir, name := convergenceStateDirFor(projectRoot), sessionID+".json"`, 쓰기는 `os.WriteFile`(약 1090행) | 확인 |
| 133 | 프로젝트 루트 생략 시 기본 상태 디렉터리 | `(:1098-1103)` (파일 정정: `internal/cli/mcp_convergence.go`) | `convergenceStateDirFor` 가 있는 구간. 생략 시 패키지 로드 때의 기본 디렉터리를 쓴다 | 확인(파일 정정) |
| 134 | 훅은 세션 cwd 에서 읽는다 | `C/multi_review_gate.go` | `loadConvergenceResults(projectDir, sessionID)`(79행) | 확인 |
| 135 | session_id 가 경로에 그대로 이어진다 | `sessionID+".json"` | 1069행 | 확인(검증 로직은 없음을 읽음) |
| 136 | 클로드 게이트의 scrub 키 집합 | `C/mcp_claude.go:251-282` | 251행 `var claudeAuditScrubKeySet = map[string]struct{}{` | 확인 |
| 137 | RunAuthStatus 가 `--restricted` 없이 실행 | `C/mcp_claude_runner.go:71-73` | 71행 `func (realClaudeRunner) RunAuthStatus(...)`, 72행 `return runClaudeCommand(ctx, binary, "", []string{"auth", "status", "--json"}, env, nil)` | 확인 |
| 138 | 바이너리 고정 우회 | `C/claude_binary.go:55-71` | 55행 `func resolveClaudeBinaryAt(projectRoot string) (string, error) {`, 56행 `if pin := os.Getenv(config.EnvClaudeBin); pin != "" {` | 확인 |
| 139 | claudeLookPath("claude") 가 :85 에 있다 | `:85` (파일 정정: `internal/cli/mcp_claude.go:85`) | 85행 `binary, err := claudeLookPath(claudeBinaryName)`. 리터럴 `"claude"` 가 아니라 상수 `claudeBinaryName` 을 쓴다 | 부분(파일 정정, 리터럴 주장은 틀림) |
| 140 | 감사에 sandboxPolicy 가 실리는 조건 | `C/mcp_codex.go:1194-1200` | 1194행 "sandboxPolicy is forwarded ONLY when the caller supplied one" 주석 | 확인 |
| 141 | codex_review 는 ReadOnlyHint(true) | `C/mcp_server.go:520` | 520행 `mcp.WithReadOnlyHintAnnotation(true),` | 확인 |
| 142 | exec.CommandContext 에 cmd.Env 를 설정하지 않는다 | `C/mcp_codex.go:449` | 449행 `cmd := exec.CommandContext(ctx, binaryPath, args...)`. `internal/cli/mcp_codex.go` 에 `cmd.Env` 가 없다 | 확인 |
| 143 | 필수 게이트 미충족을 fail 로 바꾼다 | `C/mcp_codex.go:2084-2114` 과 `C/mcp_convergence.go:977-1001` | 2084행 `applyGateUnmet`, 977행 `func enforceRequiredGateUnmet(...)` | 확인 |
| 144 | 미인식 판정은 inconclusive | `codex_review_cache_test.go:245` (파일 정정: 문자열은 246행) | 245행은 주석 "producer records inconclusive.", 246행 `withCodexSession(t, codexSessionScript("Looks good to me — no blocking findings."))`. 253–254행이 inconclusive 를 단언 | 이동(+1행) |
| 145 | codexUnrecognizedVerdict 가 inconclusive 를 돌려준다 | `:1671-1673` (파일 정정: `internal/cli/mcp_codex.go`) | 1671–1673행 `func codexUnrecognizedVerdict(method string) string { return VerdictInconclusive }` | 확인(파일 정정) |
| 146 | 발견 사항 상한 200,000 바이트 | `C/mcp_review_material.go:35` | 35행 `const reviewDiffMaxBytes = 200_000` | 확인 |
| 147 | git diff HEAD 와 base...HEAD | `:84` 과 `:82` | 84행 `return []string{"diff", "HEAD"}, nil`, 82행 `return []string{"diff", base.MergeBase + "...HEAD"}, nil` | 확인 |
| 148 | 감사 핸들러가 빈 diff 를 확인하지 않고 호출 | `C/mcp_codex.go:2028-2058` | 2028행 `method := codexMethodReviewStart`, 2058행 `out, _ := codexReviewRPC(ctx, binaryPath, method, params)` | 확인 |
| 149 | GLM 은 빈 diff 를 막는다 | `C/mcp_glm.go:308-310` | 308행 `if strings.TrimSpace(diff) == "" {` | 확인 |
| 150 | Claude 는 빈 diff 를 막는다 | `C/mcp_claude.go:101-104` | 102행 `if err != nil \|\| strings.TrimSpace(diff) == "" {` | 확인 |
| 151 | 자체 리뷰도 빈 diff 를 막는다 | `C/mcp_selfreview.go:137-144` | 137행 `if strings.TrimSpace(diff) == "" && (backend == selfReviewBackendGLM \|\| len(untracked) == 0) {` | 확인 |
| 152 | 필수 백엔드 미확정이면 Claude 값이 전체 pass 가 된다 | `C/mcp_convergence.go:249` | 249행 `overall = claudeVerdictOrDefault(verdicts, overallVerdictPass)` | 확인 |
| 153 | 감사 핸들러가 rootSource 를 계산한다 | `C/mcp_codex.go:1997` | 1997행 `root, rootSource, rootErr := resolveToolProjectRootWithSource(req)` | 확인 |
| 154 | 생략된 root 에서 codex 는 cwd 를 받지 않는다 | `C/mcp_convergence.go:634-636` | 634행 `if root := strings.TrimSpace(projectRoot); root != "" {`, 635행 `params["cwd"] = root` | 확인 |
| 155 | GLM 과 Claude 는 resolveProjectDir 로 대체 | `C/mcp_claude.go:71-74` | 71행 `root := strings.TrimSpace(req.ProjectRoot)`, 72행 `if root == "" {` | 확인 |
| 156 | GLM 쪽 대체가 :673-676 에 있다 | `:673-676` (파일 정정: `internal/cli/mcp_convergence.go:671-677`) | 671–677행: `root := strings.TrimSpace(projectRoot)`, `if root == "" { root = resolveProjectDir() }` | 확인(파일 정정) |
| 157 | CLI 쌍둥이도 같은 대체 | `C/verify.go:83` | 83행 `root := resolveProjectDir()` | 확인 |
| 158 | 서버 종료 시 Codex 백그라운드 정리 | `C/mcp_server.go:119` | 119행 `defer stopCodexBackgroundJobs()` | 확인 |
| 159 | job_status 가 생존 검사 없이 돌려준다 | `C/codex_job_control.go:148-154` 와 `C/glm_job_control.go:106-112` | 148행 `func handleCodexJobStatus(...)`, 106행 `func handleGLMJobStatus(...)` | 확인 |
| 160 | 역할 감사 작업 맵이 정리되지 않는다 | `C/codex_audit_mcp.go:80` | 80행 `codexRoleAuditJobs = map[string]*codexRoleAuditJob{}` | 확인 |
| 161 | 역할 감사가 context.Background 로 돈다 | `:123` | 123행 `res := plan.run(context.Background())` | 확인 |
| 162 | EOF 시 "cancelled by the caller" 로 기록 | `C/codex_task.go:193-195` | 193행 `if parentErr := parent.Err(); parentErr != nil && ...` 분기. 문자열 "cancelled by the caller" 는 145행 `reason := "cancelled by the caller"` | 확인(분기 줄, 문자열은 145행) |
| 163 | 전경 타임아웃은 "" 를 돌려준다 | `C/mcp_codex.go:1304-1306` | 1304행 `if err := ctx.Err(); err != nil {`, 1305행 `return "", fmt.Errorf("codex review turn ended by context: %w", err)` | 확인 |
| 164 | 스레드는 백그라운드 작업에만 기록 | `C/codex_jobs.go:389-390` | 389–390행 주석: "REQ-CX2-003 creates them for BACKGROUND jobs only" | 확인 |
| 165 | codex_audit 는 자체 제한이 없다 | `C/mcp_codex.go:2058` | 2058행 `codexReviewRPC(ctx, ...)` | 확인 |
| 166 | recv 는 타이머 없이 막힌다 | `:586-589` | 586행 `func (c *realCodexConn) recv() (string, bool) {` | 확인 |
| 167 | 네이티브 리뷰가 effort 핀을 빠뜨린다 | `C/mcp_codex.go:809-811` | 809행 `if me.Model != "" {`, 810행 `threadParams["model"] = me.Model` | 확인 |
| 168 | 감사 모델 기록은 Claude 에만 있다 | `C/mcp_claude.go:169-183` | 169행 `out.Provenance = &AuditProvenance{` | 확인 |
| 169 | GLM 응답을 ReviewOutput 으로 직접 해석 | `C/mcp_glm.go:398-408` | 398행 `var out ReviewOutput` | 확인 |
| 170 | glm_audit 는 4096 토큰, 120 s | `C/mcp_glm.go:76,80` | 76행 `glmAuditMaxTokens = 4096`, 80행 `glmAuditHTTPTimeout = 120 * time.Second` | 확인 |
| 171 | glm_review 는 900 s 문맥을 쓴다 | `C/mcp_selfreview.go:224-228` | 224행 `detached, cancel := context.WithTimeout(context.Background(), config.DefaultCodexReviewGateTimeout)` | 확인 |
| 172 | Content[0].Text 를 읽는다 | `C/glm_task.go:353` | 353행 `if len(env.Content) == 0 \|\| env.Content[0].Text == "" {` | 확인 |
| 173 | 감사 캐시는 잠기지 않은 맵 | `I/runtime/audit_cache.go` | `InMemoryCache`(108행). 파일에 `sync.` 와 `Mutex` 가 없다 | 확인 |
| 174 | 빈 디렉터리 해시가 오류 없이 끝난다 | `I/runtime/audit_cache.go` (ComputeHash) | `TestComputeHashEmptyDirFails` 존재(이름 확인). 본문은 읽지 않았다 | 부분 |
| 175 | Codex 표준출력 스캐너 상한 8 MB | `C/mcp_codex.go:566` | 566행 `sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)` | 확인 |
| 176 | job_id 가 경로에 이어진다 | `C/codex_jobs.go:234-236` 과 `C/glm_jobs.go:204-206` | 235행 `return filepath.Join(r.dir, id+".json")`, 205행 동일 | 확인 |
| 177 | Stop 체인 멤버의 TTL 0 | `C/codex_stop_chain.go:660` | 660행 `chk := verify.CheckReceipt(..., time.Now(), 0)` | 확인 |
| 178 | 잡 메모 파일명(`audit-multi/<session>.json`, `second-review.jsonl`) | (파일 이름) | 이름 규칙의 기술이며 파일 확인은 생략 | 관측불가 |
| 179 | glm_task 가 MOAI_FACTORY_WORKERS 일 때 model 무시 | `C/glm_task.go:163-175` | 163행 `} else if os.Getenv(config.EnvMoaiFactoryWorkers) != "" {` | 확인 |
| 180 | 파일 존재 (읽지 않음 목록) | `codex_launcher.go`, `codex_sync_gate.go`, `codex_role_fingerprint.go`, `codex_readiness.go`, `glm.go`, `glm_tools.go`, `verify_run.go`, `codex_stop_chain.go` | 모두 `internal/cli/` 에 존재한다 | 확인(존재만) |
| 181 | 테스트 공백: stopCodexBackgroundJobs, handleCodexRoleAudit, moaiAuditCache | (테스트 파일) | 세 이름을 참조하는 테스트 파일이 없다 | 확인(부정 주장) |
| 182 | "cmux" 문자열이 없다 | (`internal/`, `.claude/`) | 없다. "shim" 은 확인하지 않았다 | 부분 |
| 183 | 레인 메시지 훅의 경로 설명 | `I/hook/factory_messages.go`, `I/factorymsg/store.go` | 위 107–108행과 같다 | 확인 |

### §3.3 lens-5 (기록 통계, 코드 인용)

| # | 주장 (요지) | 인용 | 이 트리의 줄 (요지) | 판정 |
|---|---|---|---|---|
| 184 | 레인 임대 상수 15분 | `internal/config/defaults.go:65` | 65행 `DefaultFactoryLeaseDuration = 15 * time.Minute`. 주석의 "chosen value" 문구는 읽지 않았다 | 확인(값) |
| 185 | 기록 scrub 이 `scrubClaudeAuditEnv` 에 있다 | `internal/cli/mcp_claude.go:266` | 266행 `func scrubClaudeAuditEnv(environ []string) []string {` | 확인 |
| 186 | 팩토리 카드 분류 수정 | `internal/cli/factory_card.go:1229-1238` | 1229행 `classOf := func(id string) factory.CardClassification { return factoryQueueClassification(queueRec, id) }`. 분기 본문은 읽지 않았다 | 부분 |
| 187 | 리뷰 노트 파일 경로 | `internal/cli/plan.go:111` | 111행 `dir := filepath.Join(root, ".moai", "reports", "plan-audit")` | 확인 |
| 188 | audit_report.go 가 같은 디렉터리를 쓴다 | `internal/runtime/audit_report.go` | 15행 "Daily report path: .moai/reports/plan-audit/<SPEC-ID>-<YYYY-MM-DD>.md" | 확인 |
| 189 | DefaultCodexTaskTimeout 600 s | `defaults.go:888` | 888행 | 확인 |
| 190 | 로컬 EOF 취소 | `internal/cli/mcp_server.go:144-156` | 144행 `func (r *mcpEOFReader) Read(...)`, 147행 `r.cancel()` | 확인 |
| 191 | "branch" 키를 보낸다 | `internal/cli/codex_review_scope.go:247` | 247행 `"target": map[string]any{"type": codexTargetBaseBranch, "branch": scope.MergeBase},` | 확인 |
| 192 | tree_scope=skip 사유 | `internal/cli/codex_stop_chain.go:639` | 639행 `return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Status: stopStatusNotApplicable, Reason: "tree_scope=sk...` | 확인 |
| 193 | 엔진 파일 존재 | `internal/runtime/audit_ceiling.go` | 존재한다 | 확인(존재만) |
| 194 | 말뭉치 수치 (783, 307, 397, 348, 2,268, 115,724, 20,229 등) | (기계 로컬 말뭉치) | 재계산하지 않았다 | 관측불가 |
| 195 | 6회 `--skip-hook`, 482회 누락 훅 | `.moai/logs/hook-skip.log`, `.moai/logs/hook-missing.log` | 기계 로컬 로그 | 관측불가 |
| 196 | 기억 파일 이름 (feedback_* 등) | `~/.moai/claude-profiles/…/memory/` | 저장소 밖 기계 로컬 저장소 | 관측불가 |
| 197 | 팩토리 메시지와 DB 읽기 | `factory.db`, `hook-runtime.log` 등 | 읽지 않았다(lens-5 §D 목록 그대로) | 관측불가 |

## §4 재현 근거 — 레인 t1543

관측 대상은 이 트리의 상대 경로로 적는다. 레인 저장소는 `.moai/worktrees/t1543/` 아래에 있다.

- 디렉터리: `.moai/worktrees/t1543/.moai/state/verify/` 에 `codex-review/` 와 `snapshots/` 가 있다(수정 시각 2026-10-10 01:37).
- `snapshots/2aab5f797b75-7f1e2657f0831d9f.json`(448 B, 권한 0600). 내용의 요지:
  - `key`: `2aab5f797b75983e132af451da68f69e3426557b:7f1e2657f0831d9f`
  - `recorded_at`: `2026-10-10T01:37:46.265725+09:00`
  - 검사 `codex-review`: `command` `moai verify codex-review`, `exit_code` 1, `verdict` `fail`, `tool_version` `codex-cli 0.161.0`, `config_digest` `cc6efc5bda54a3c3`
- `codex-review/2aab5f797b75983e132af451da68f69e3426557b-7f1e2657f0831d9f.md`(3,011 B): 판정 `fail`, 발견 사항 두 건.
  - P1: `internal/homestate/card_transition.go:287`. 이 트리의 287행은 `if err := req.BeforeCommit(c); err != nil {` 이다. 발견 사항은 두 저장소의 디스패치 갱신이 부분적으로 남을 수 있다고 적는다.
  - P2: `internal/cli/factory_card.go:2894`. 이 트리의 2894행은 `ApprovalUUID: approvalUUID,` 이다. 발견 사항은 PR 완료 전이에서 승인 영수증 버전을 갱신하지 않는다고 적는다.
  - 상세 파일은 검토 기준 HEAD `2aab5f797…` 을 명시하고, "정적 코드 검토만 수행, 테스트·빌드·재현 명령은 실행하지 않음"이라고 적는다.
- `codex-review/…kick-1990626`(20 B, 01:31): 진행 중 표식으로 보인다. 프로세스 식별자는 파일 이름에서 읽었고, 프로세스의 생존은 확인하지 않았다.

판정:

- 카드의 재현 사례(fail 영수증이 main 쪽 발견 사항으로 생성되고, 카드가 바꾼 것이 없다)는 관측과 맞는다. 키가 이 트리의 HEAD 와 같고, 상세는 카드가 건드리지 않은 main 쪽 코드 두 곳을 가리킨다.
- 24시간 잠금: 기록 시각 2026-10-10 01:37:46 KST 에 유효기간 24시간(`internal/cli/codex_review_cache.go:29`)을 더한 2026-10-11 01:37 KST 는 산술 결과이며 관측이 아니다. 진입 훅이 실제로 막았는지는 이 세션에서 재현하지 않았다(§3 #22).
- 관측불가: 레인 자체의 커밋 수(git 가드 거부), 발견 사항의 옳고 그름. 기계 로컬 파일이므로 이 저장소의 커밋 근거로 쓸 수 없다.

## §5 전제 확인 (P 항목)

| 번호 | 전제 | 관측 | 판정 |
|---|---|---|---|
| P-1 | 두 키의 값이 다르다 | `git-strategy.yaml` 5행 `main`, 16행 `develop` | 확인 |
| P-2 | 턴 종료 기준은 IntegrationTarget 이며 git-flow 수동 프로필에서 develop 이다 | `internal/config/loader_integration_branch.go:123-125`(git-flow 면 `DevelopBranch`, 아니면 `WorkflowIntegrationTarget`) | 확인 |
| P-3 | 위상: main 과 develop 의 공통 조상은 db0c514d3, 139 와 149 커밋 | git 측정(§3 #26) | 확인 |
| P-4 | 현재 HEAD 는 main 과 같다 | `git rev-parse --short main` = `2aab5f797` | 확인 |
| P-5 | develop 의 HEAD 는 main 과 다르다 | `git rev-parse --short develop` = `966f1b158` | 확인 |
| P-6 | doctor 에 두 계층 동시 활성 경고가 없다 | `internal/cli/doctor.go`, `doctor_codex.go`, `doctor_harness.go`, `doctor_sandbox.go`, `doctor_agentemit_embed.go`, `doctor_settings_defaultmode.go` 에서 review 를 찾으면 `doctor_sandbox.go:101`(역할 목록)과 `doctor_harness.go:48`(주석)뿐이다. `internal` 전체에서 card-review·turn-end 와 warn·doctor 를 함께 찾으면 주석만 나온다 | 미발견 |
| P-7 | 문서에는 두 계층 중복 문장이 있다 | `.claude/rules/moai/workflow/factory-dispatch-detail.md:157` | 확인 |
| P-8 | 설정 `interview.decision_gate` 는 on 이다 | `.moai/config/sections/interview.yaml:6` `decision_gate: on`, `:9` `recommendation_mode: pull` | 확인 |
| P-9 | 추적 workflow.yaml 의 턴 종료 스위치는 HEAD 에서 true 다 | HEAD 302행 `enabled: true` | 확인 |
| P-10 | 작업 트리의 workflow.yaml 은 리더의 미커밋 변경이 있다 | `git diff -- .moai/config/sections/workflow.yaml`: `-            enabled: true` / `+            enabled: false`(한 줄) | 확인 |
| P-11 | 영수증 결합 필드가 다섯 개다 | `internal/verify/receipt.go:70-76`: head, tree_digest, config_digest, command, tool_version | 확인 |
| P-12 | 영수증 유효기간 0 은 기본값 10분이다 | `internal/verify/receipt.go:61,90`(`ttl <= 0` 이면 DefaultTTL), `internal/verify/freshness.go:5-7`(`DefaultTTL = 10 * time.Minute`) | 확인 |
| P-13 | 진입 훅이 추가 맥락을 실을 수 있다 | `HookOutput.AdditionalContext` 사용처: `internal/hook/subagent_start.go:79`, `internal/hook/session_start_compact.go:72` | 확인 |
| P-14 | 진입 훅은 설계상 차단을 쓴다 | `internal/cli/codex_review_delay.go` 1–14행: "a fresh FAIL blocks the prompt" 와 `async:true` 를 채택하지 않은 이유 | 확인 |
| P-15 | 진행 중 SPEC 이 같은 차단을 명시한다 | `.moai/specs/SPEC-GATE-BOTTLENECK-001/spec.md`: `status: in-progress`, 43–46행 REQ-GBN-002, 67–68행 AC-GBN-004 | 확인 |
| P-16 | 설치 빌드는 트리의 후손이다 | `git merge-base --is-ancestor HEAD 4f8aba061` 참 | 확인 |
| P-17 | 완료된 SPEC 의 결정이 코드에서 뒤집혔다 | `.moai/specs/SPEC-CODEX-REVIEW-TARGET-001/spec.md` 102–111행 §A.7(완료), 테스트 `internal/cli/codex_review_target_test.go` AC-CRT-002 의 주석("Card t1426 reverses ..."), 커밋 `f82a078bb`(2026-10-03 11:40:39 +0900, 제목 "fix(mcp): audit tools review the card diff against the configured integration base (card t1426)"), SPEC 본문의 `t1426` 개수 = 0 | 확인 |
| P-18 | 카드 t1426 은 config 키를 step 0 으로 올렸다 | 커밋 `f82a078bb`, `resolveReviewBase` 159행 | 확인 |
| P-19 | 턴 종료 게이트는 현재 키의 fail 영수증이 있을 때만 차단한다 | `codex_review_gate.go:159-174` | 확인 |
| P-20 | 인수 기준 가드 테스트는 모두 존재하고 셀 B-1 에서 PASS 한다 | §6 과 acceptance.md §C | 확인 |

## §6 테스트 이름과 부정 주장 확인

- 존재 확인(`grep -rhoE 'func Test…'` 로 찾은 이름): `TestReviewEntry_DisabledAllows`, `TestReviewEntry_FreshFailBlocks`, `TestReviewEntry_FreshPassAllows`, `TestReviewEntry_NoEditTurnAllows`, `TestReviewEntry_NoReceiptAllows`, `TestReviewEntry_StaleReceiptAllows`, `TestReviewEntry_SubdirSessionReadsRootReceipt`, `TestReviewGate_BackgroundKickFailureAllows`, `TestReviewGate_CacheHitReusesFailVerdict`, `TestReviewGate_CacheMissKicksBackgroundReview`, `TestReviewGate_DisabledAllows`, `TestReviewGate_FailedKickIsRetryable`, `TestReviewGate_FailOpenOnMissingCodex`, `TestReviewGate_InconclusiveReceiptIsNotPassEvidence`, `TestReviewGate_InFlightKickNotRepeated`, `TestReviewGate_LoopPreventionAllows`, `TestReviewGate_MissingCodexDoesNotKick`, `TestReviewGate_NoEditTurnAllows`, `TestReviewGate_StaleKeyKicksBackgroundReview`, `TestReviewGate_SweptMarkerRekicks`, `TestReviewGateReaders_AgreeWithConfigLoader`, `TestReviewGateReaders_HonourNestedWorkflowKeyPath`, `TestCardMergeBaseConfiguredMain`, `TestCodexReviewScope_AbsorbedDevelopRecomputesBase`, `TestCodexReviewScope_CardDetectorRuntimePathsExcluded`, `TestCodexReviewScope_CardUntrackedLegCountsNonRuntimeFiles`, `TestCodexAudit_AdversarialBaseBranchClaimsNoReviewBase`, `TestCodexAudit_BaseBranchResolutionChain`, `TestKickInFlight_ExclusiveAcquisition`, `TestProducerFailReceiptPreservesDetailForDelayedBlock`, `TestProduceCodexReviewReceipt_SubdirRootStoresAtGitRoot`, `TestResolveProjectDirFromInput_Order`, `TestStaleBinarySkip_BehindSkips`, `TestStaleBinarySkip_DevBuildReviews`, `TestStaleBinarySkip_NonBehindVerdictsReview`, `TestHandleCodexReviewGate_StaleBinaryAllowsWithRow`, `TestComputeHashEmptyDirFails`(`internal/runtime`).
- 부정 주장: `TestReviewGate_*`/`TestReviewEntry_*` 가운데 `StopHookActive` 를 참조하는 파일은 `codex_review_delay_test.go` 에 없다(0). `stopCodexBackgroundJobs`, `handleCodexRoleAudit`, `moaiAuditCache` 를 참조하는 `_test.go` 파일은 없다.
- 셀 B-1 과 B-2 의 PASS 는 §3 과 acceptance.md §C 에서 확인했다.

## §7 F3 주장 검증 (factory 두 건)

연구 입력 lens-1 의 F3 후보를 이 트리에서 확인했다. 두 주장 모두 코드에서 관측했으나, 일부는 주장과 다르다.

### F3(a) — 빈 SQLite 가 JSON 큐 이관 전에 만들어질 수 있다

- 인용: `internal/factory/todo_dispatch_current.go:53`. 53행은 `eng, err := openBacklogEngine(backlogSQLitePath(s.path))` 다(확인).
- 앞선 가드: 같은 함수는 먼저 읽기 전용 로드(`LoadPure`)로 해당 카드의 행이 있는지 본다. 행이 없으면 쓰기 엔진을 열지 않는다(40–52행). 함수 상단 주석이 이유를 적는다. 쓰기 열기는 레거시 JSON 큐를 SQLite 로 이관하고, 없는 DB 를 만들고, 스키마 DDL 을 실행하기 때문이다.
- 열기 순서: `openBacklogEngine`(`internal/factory/backlog_sqlite.go:357`)은 `sql.Open` → `PingContext` → `ensureSchema` 순서로 DB 파일과 스키마를 만든 뒤에 권한을 조정하고 반환한다. 이 열기 자체는 이관을 수행하지 않는다. 따라서 파일이 생기고 이관 전의 빈 DB 가 존재하는 순간은 실제로 있다.
- 방어: `backlog_sqlite.go:333-345` 의 읽기 경로는 "uninitialized database beside legacy queue"를 `ErrBacklogCorrupt` 로 거부한다. 주석이 이유를 적는다: "Older releases published an empty DB before copying legacy records. Neither a pure read nor an adopting read may call that empty state real." 즉 빈 DB 가 레거시 큐 옆에 남으면 읽기가 실패하고 점검을 요구한다. 조용히 빈 큐로 읽히지 않는다.
- 판정: **부분 확인.** 빈 SQLite 가 이관 전에 생길 수 있는 순서는 확인되었다. 그러나 결과는 조용한 손실이 아니라 읽기 경로의 실패(닫힌 쪽 실패)다. 중단 시점의 실제 도달 가능성은 측정하지 않았다(관측불가). 주장 문구 "may create an empty SQLite before the JSON queue is migrated"는 맞고, 그 결과 "데이터가 사라진다"는 함축은 확인되지 않았다.

### F3(b) — 훅은 할당 트랜잭션 안에서 실행되며, 외부 저장소 변경 뒤 내부 커밋이 실패하면 복구되지 않는다

- 인용: `internal/factory/todo_runtime.go:364`. 364행은 `if err := hook(); err != nil {` 이다(확인). 커밋은 368행 `return tx.Commit()` 이다(확인).
- 트랜잭션의 출처: 288행 `tx, err := eng.db.BeginTx(ctx, nil)`. `eng` 는 백로그 SQLite 엔진이다.
- 훅의 쓰기 경로: 호출 체인 가운데 하나(`internal/cli/factory_card.go:1010`)는 `todoStoreAt(root).RefreshDispatchCurrentLockHeld(...)` 를 호출한다. 이 함수(`internal/factory/todo_dispatch_current.go:53`)는 `openBacklogEngine` 으로 **자기 자신의 연결**을 열고 `eng.db.ExecContext` 로 `UPDATE todo_dispatch_current` 를 실행한다. 이 실행은 `tx` 와 다른 연결의 자동 커밋이다. 따라서 훅의 쓰기는 `tx` 의 롤백 범위에 들지 않는다.
- 주석의 보장 범위: 358–363행 주석은 **훅 실패**가 할당을 롤백한다고만 보장한다. 훅이 성공한 뒤 커밋이 실패하는 경우는 다루지 않는다.
- 판정: **메커니즘 확인, 도달성 미측정.** 훅 쓰기가 커밋 실패 뒤에도 남는다는 구조는 코드로 확인되었다. 실제 실패 창이 열리는지는 SQLite 잠금 순서에 달려 있고(같은 DB 파일에 두 연결이 쓰므로 잠금 대기 또는 시간 초과가 먼저 날 수 있다), 이 세션에서 측정하지 않았다(관측불가). 추적한 훅 호출자는 `factory_card.go:1010` 하나이며, 다른 호출자는 추적하지 않았다.
- 참고: 연구 입력의 P1 발견 사항(`internal/homestate/card_transition.go:287`)은 같은 메커니즘을 설명한다. `BeforeCommit` 의 `RefreshDispatchCurrent` 가 별도 백로그 DB 에 즉시 커밋한다는 내용이며, 이 트리의 코드와 일관된다.

## §8 Gaps (관측하지 못한 것)

1. 레인 t1543 저장소의 git 상태: 자체 커밋 수, 168 파일 diff, 현재 브랜치(`WT-t1534-hold-1` 주장). 워크트리 가드가 다른 워크트리 대상의 git 호출을 거부했다.
2. 설치 빌드의 라이브 재현(표준입력 파이프와 `moai hook codex-review-entry`). 단일 호출 형식 밖이고, 설치 빌드는 트리와 다르다.
3. 레인 저장소 말뭉치(lens-1 §1 집계 48/48/21, lens-5 §A 전체 수치, 마커 64개, `--allow-empty` 우회, 같은 digest 재검토). 모두 기계 로컬이며 이 세션에서 재계산하지 않았다.
4. 훅 타임아웃 측정(1.53 s, 1.90 s)과 부하 18 의 재현.
5. `hooks-system.md:368` 의 "5 s → 15 s (2026-10-09)" 기록: 저장소에서 찾지 못했다(§3 #48, 미발견).
6. 발견 사항 P1·P2 의 옳고 그름. codex 출력의 주장이며 이 세션은 검증하지 않았다.
7. `TestReviewGate_CacheHitReusesFailVerdict`(AC-RG-004 가드)와 `TestReviewGate_InconclusiveReceiptIsNotPassEvidence` 의 단언 끝부분. 본문 앞부분만 읽었다.
8. `auditreceipt/store.go` 1267–1279행(fail 이 통과로 되는 경로), 영수증 상태 계산의 `--version` 실패 처리, `internal/binlag` 의 `StatusBehind` 정의, 주 체크아웃 기본 판독 값.
9. 브랜치 포함 질의(`git branch --contains 4f8aba061`)와 WT-* 브랜치 목록.
10. factory 큐 쓰기 경로의 실제 실패 창 도달성(§7 F3(a), F3(b)).
11. 다른 플랫폼(Windows 등)의 동작과 `.codex` 훅 연결.
12. lens-5 의 `factory_card.go:1229-1238` 분기 본문과 `feedback_*` 기억 파일들(저장소 밖).
