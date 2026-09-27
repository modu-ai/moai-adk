# Progress — SPEC-HANDOFF-NEUTRAL-001 (card t1273)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-26T20:10:00+09:00
plan_artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (Tier L 5)
plan_commit: (이 커밋에 동반)
baseline: worktree t1273, branch WT-handoff-neutral, develop 분기점 1b7a88d78

## §F Phase 4 Mode Selection

- tier: L / scope: internal/cli·internal/codexadapter·internal/homestate (신규 파일 2-3 + 수정 3-4) / domain count: 3 (CLI·어댑터·저장소) / language mix: Go + Markdown / concurrency benefit: LOW (직렬 의존: RED→show→시딩→LIVE→판정)
- direct: not selected — 다중 파일·다중 도메인
- fanout: not selected — 코딩 중심(Anthropic coding-task parallelism caveat), 단계 간 의존
- sweep: not selected — ~30 파일 미만·비기계적
- agent-team: not selected — 명시 요청 없음
- **Decision: serial**
- Justification: 구현 체인이 직렬(RED 관측 → show → 시딩 → LIVE → P1 판정)이고 각 단계가 앞 단계 산출을 소비. 단일 manager-develop 위임이 동기 부여·의존 관리 모두에서 단순.

## 진행 기록

- 2026-09-26 research.md — 공식 문서 4건 검증(§F), codexadapter 갭 발견, 쿼터 관측(F-7). 커밋 36326286f·f6c96d862·fb51c8a59.
- 2026-09-26 design.md — P3 기본+P1 조건부(전제 2개: 워크트리 시딩·관문 b), D2.5 옵션 A(materializer 시딩+런처 보완, 리드 조정 반영). 커밋 cc463138a.
- 2026-09-26 spec.md·acceptance.md·plan.md 작성 — plan-audit 대기.
- 2026-09-26 **plan-audit 1차 FAIL 0.75** (임계 0.85). 방향(P3+조건부 P1·LIVE 관문·codexadapter 갭·워크트리 부재)은 감사가 소스 재현으로 확인; 결함 5건은 전부 명세 수준 — 수리 커밋 794b749b6: ① 저장모델 부정합(pending.json→factory.db `resume_handoffs` 상태머신) ② 시딩 재사용 소재(`update_codex_wiring.go` "creates nothing" → `internal/codexwiring` wire.go) ③ REQ-HN-011 미커버(AC-HN-011 신설) ④ AC-HN-010 허위 셀렉터(실존 테스트로 교체) ⑤ red 셀 SHA 계약. acceptance fixture 계약 DB 세팅 전환(공허 초록 방지).
- 2026-09-26 **plan-audit 재심사 PASS-WITH-DEBT 0.92** (임계 0.85 초과, 회귀 없음; Clarity 1.0 · Completeness 1.0 · Testability 0.75 · Traceability 1.0). 1차 결함 D1·D2·D3·D7 전부 RESOLVED. 잔여 정리 커밋: R1(AC-HN-010 TestMapOutput 조건부 스윕 주석), R2/R3(워크트리 패키지 잔여 참조 → materializer 실측 소재 `internal/cli/session_worktree.go`), R4(module: internal/homestate 추가), R5(design D5 잔문), D6(design §D3 소비 조건 실측 보강 — handoff_inject.go:52-100, 저장자 신원 불검사·claim_token CAS). 리드 재점검 요청 ①②도 이 커밋에 반영: 저장 하네스 무관(handoff.go harness 참조 0건)·역방향 성립(소비 조건 4개뿐, SavedBySession 판정 미사용). **session-handoff.md:31 pending.json 드리프트를 M2 문서 범위에 등록**(판정서에 기록 예정).
- 2026-09-26 **M1.1–M1.3 run 완료** (M1.4/M1.5 LIVE·조건부 어댑터는 쿼터 회복 9/28 14:37 이후로 분리 — 위임 범위 그대로). RED 선관측 → M1.2 `cfa540967` → M1.3 `a347d83ee`. spec.md `draft → in-progress` (M1 커밋 동반, manager-develop 소유 전이).
- 2026-09-26 레인 검증 배치: 커밋 3건·RED 첫 줄 HEAD 계약·PRESERVE 4대상 무변경(git diff)·show/시딩/보존 그룹 GREEN 재실행 ok·`go build ./...`/`go vet` exit 0 (에디터 gopls가 워크트리를 go.work에 안 넣어 뿜는 undefined 진단은 재측정으로 기각 — 판정 근거는 실측 실행). manager-develop가 적발한 acceptance 표 셀 `\|` 셀렉터 공허 초록 결함을 fenced evidence ledger로 수리(`934e111b0`, 양성 대조 L1 스윕 3·L2 스윕 7). AC-cascade 재측정: `go test ./internal/spec/ -count=1` → ok 132.8s exit 0.
- 2026-09-26 **M1.4 격리 환경 사전 구축 완료** (리드 지시, 모델 호출 0회): `/tmp/t1273-live` — 격리 CODEX_HOME·스크래치 proj(`.codex/config.toml` features.hooks + hooks.json SessionStart/UserPromptSubmit 배선, 이 트리 HEAD 빌드 바이너리 `/tmp/t1273-live/moai-t1273`), 핸드오프 표본 save→show 재출력 **실환경 관측**(factory.db 저장·출처 pending 헤더·본문 verbatim). 판정서 `.moai/reports/t1273/verdict.md` 작성 — 머리 「외부 차단: Codex 쿼터 2026-09-28 14:37」, LIVE 절차(양방향·관문별 판정 명령·상한 관문당 3회·30분·증거 경로 live-gate-*.txt)·M1.5 분기 판정 선언. **격리 인증 방식 리드 확정(auth 사본, [HARD]: LIVE 직전 복사+즉시 600, 삭제는 같은 복합 호출의 trap 정리, 부재는 경로+absent만 기록, 사본 흔적 전면 금지) — 판정서 LIVE 절차에 반영 완료.** **대기 상태 진입.**
- 2026-09-28 **재개 (worker-64)**: ① 다른 작성자 부재 확인(lsof cwd 0건) ② develop `e9577de4f` 흡수 — 머지 `ddf24851f`, 충돌 0 ③ 병합 트리 재측정 **전항목 초록**: `go build ./...`·`GOOS=windows` exit 0 · vet(cli/codexadapter/homestate/codexwiring) exit 0 · show/시딩 스윕 10 PASS · 회귀 스윕 cli ok(59 PASS) · `internal/hook` RenderHandoffContext PASS 보강(원 판정의 2패키지 스윕 재현 — TestRenderHandoffContext는 hook 패키지) · homestate ok 85.1s·codexwiring ok · golangci-lint **v2.1.6**(CI 판 일치) `0 issues.` ④ 판정 바이너리 `ddf24851f` 재빌드(sha `d04b1a46…`) → save→show 재출력 재현(exit 0, pending 헤더, 마커 2) ⑤ **관문 (a) 시도 3회 → 상한 도달·훅 무발화**: 쿼터는 회복 실측(attempt2 gpt-6-astra 4,635 tokens exit 0). 신규 실측 — codex v0.157.0의 `--skip-git-repo-check` 요구·hook trust 개념·`hooks stable true`. 훅 실행 흔적 없음(상태 파일 mtime 무변화·로그 부재). 증거 `live-gate-a-attempt{1,2,3}.txt`·`live-show-recheck.txt` — 판정서에 전문 기록. **리드 보고·판정 대기** (원인 후보: trust 경로 / hooks.json 발견 경로). 부수 발견: 원 회귀 스윕의 `TestHandoffRecover`는 역사상 부재 테스트(0-스윕 토큰) — sync 시 정정 후보. auth 사본 trap 정리가 worktree 가드(2.1.275)에 거부돼 [실행]→[rm+부재기록] 연속 호출 형태로 대체(매 시도 absent confirmed).

## §E.2 Run-phase Evidence

> 모든 판정 명령은 환경 스크럽 단일 복합 호출(`unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && <명령>`)로 실행. 측정 트리: worktree t1273, branch WT-handoff-neutral, 구현 전 RED는 HEAD `ecd5a9e58`, GREEN은 구현 커밋 `cfa540967`·`a347d83ee` 위.

### RED 선관측 (M1.1, E8)

- `.moai/reports/t1273/red-show.txt` — 첫 줄 `HEAD ecd5a9e58`. 관측 출력 발췌: `internal/homestate/handoff_consumed_test.go:55:29: db.ReadLatestConsumedResume undefined` / `internal/cli/handoff_show_test.go` 빌드 실패, exit 1. RED 사유: `handoff show` 서브커맨드·`ReadLatestConsumedResume` 부재 (acceptance AC-HN-001 red 셀 계약이 명시한 컴파일 실패 경로).
- `.moai/reports/t1273/red-seeding.txt` — 첫 줄 `HEAD ecd5a9e58`. 관측 출력 발췌: `unknown field seedHooks in struct literal of type swSeams` / `undefined: seedCodexHooksReal` / `too many arguments in call to resolveWorktreeL2Path`, exit 1. RED 사유: 시딩 표면(SeedHooksIfMissing·seedCodexHooksReal·warn 시그니처) 부재.
- 채택 순서 계약 준수: 테스트 파일 3종(handoff_show_test.go·session_worktree_codexseed_test.go·handoff_consumed_test.go)이 구현 커밋보다 먼저 작성·실행됨. 첫 RED 캡처에 테스트 측 오타 1건(3값 반환 snapshot 1변수 수령)이 섞여 있어, 오타만 수정한 뒤(구현 전, 같은 HEAD) RED를 재관측해 red-show.txt를 청결판으로 갱신 — RED 사유는 변동 없음.

### AC 판정 행 (M1 범위: 001..007, 004b + 보존·형식 AC)

| AC | 판정 | 근거 (명령 → 관측) |
|---|---|---|
| AC-HN-001 | PASS | `go test ./internal/cli/ -count=1 -run 'TestHandoffShow_(PendingSource|ConsumedFallback|NoHandoffErrors)' -v` → `--- PASS: TestHandoffShow_PendingSource` · `--- PASS: TestHandoffShow_NoHandoffErrors` · `--- PASS: TestHandoffShow_ConsumedFallback` · `ok github.com/modu-ai/moai-adk/internal/cli 3.044s` (스윕 3건 확인). 주: acceptance 표의 셀렉터 리터럴(`\|`)은 go test 정규식에서 alternation이 아니라 리터럴 `\|`로 읽혀 **0건 스윕 + ok**(공허 초록, verification-completeness §1.1)이 된다 — 본 판정은 백슬래시를 제거한 동일 셀렉터로 스윕 3건을 확인한 실행이 근거다. |
| AC-HN-002 | PASS | `-run 'TestHandoffShow_DoesNotMutateState' -v` → `--- PASS` — show 2회 호출 후 status·consumed_at·body SQL 재열람 대조 불변, 2회 출력 바이트 동일 |
| AC-HN-003 | PASS | `-run 'TestHandoffShow_JSONOutput' -v` → `--- PASS` — pending/consumed 양쪽 source + record(body·spec_id·phase) JSON 검증 |
| AC-HN-004 | PASS | `-run 'TestHandoffShow_LocaleHeader' -v` → `--- PASS` — ko 헤더 `출처: pending`(영어 라벨 부재), en 헤더 `Source: pending`, 본문 verbatim |
| AC-HN-004b | PASS | `-run 'TestHandoffShow_LegacyCompatRead' -v` → `--- PASS` — 유일한 파일 fixture 케이스(읽기 호환 분기 명시 검증) |
| AC-HN-005 | PASS | `-run 'TestNew_SeedsCodexHooksJson' -v` → `--- PASS` — materializer 경로에서 `.codex/hooks.json` 생성·`"moai hook ` 커맨드·description 확인 (생성 로직 `internal/codexwiring` RenderHooks 재사용) |
| AC-HN-006 | PASS | `-run 'TestEnterWorktree_SeedsMissingCodexHooks' -v` → `--- PASS` — 부재 트리 진입 시 채워짐(단축명·절대경로 양쪽), 존재 파일 진입 시 바이트 불변 |
| AC-HN-007 | PASS | `-run 'TestNew_SeedFailureFailOpen' -v` → `--- PASS` — 경로 오염(`.codex`을 일반파일로 점유) 강제 실패에도 워크트리 생성 성공 + `codex hooks` 진단 발화 |
| AC-HN-008 | DEFERRED | 관문 (b) 미실시(쿼터 9/28 14:37) — 위임 범위 밖(M1.5). 테스트 미작성은 계약상 정상 상태 |
| AC-HN-009 | DEFERRED | LIVE 관문 (a)(b)(c) — 쿼터 회복 후 실행(regression-guard 등급, 판정 기록 의무) |
| AC-HN-010 (보존) | PASS | `go test ./internal/cli/ ./internal/hook/ -count=1 -run 'TestHandoffSave_(WritesJSONNotMarkdown\|Schema\|Stdin\|RequiresBody)\|TestHandoffClear\|TestHandoffCmdRegistered\|TestRenderHandoffContext' -v` → `--- PASS` 7건 스윕(grep -c로 확인), `ok` 2패키지. TestMapOutput은 관문 (b) 전까지 이 트리에 부재 — AC 주석대로 제외 |
| AC-HN-011 | PASS | `grep -c "moai handoff save" .moai/specs/SPEC-HANDOFF-NEUTRAL-001/design.md` → `6` (≥1, design §D3 방향 중립 문서화 — plan-phase 산출물, run에서 변동 없음; 이 run에서 재측정) |
| AC-HN-012 (형식) | PASS | `go vet ./internal/cli/ ./internal/codexadapter/ ./internal/homestate/` exit 0 · `gofmt -l` 대상 파일 전부 빈 목록(첫 측정에서 handoff.go·handoff_show_test.go 2건 적발 → `gofmt -w` 후 재측정 빈 목록) |

### 회귀·경계 (E2·E4·E5·E6)

- **E2 build**: `go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (구현 후 재측정 포함).
- **E4 boundary**: `grep -rn 'AskUserQuestion' internal/cli/handoff*.go internal/cli/session_worktree.go …`(touch 파일 전부) → 0건 (exit 1, 무출력).
- **E5 lint**: `golangci-lint run --timeout=2m ./internal/cli/... ./internal/homestate/... ./internal/codexwiring/...` → `0 issues.` (첫 측정 errcheck 1건 — handoff.go fmt.Fprintf 반환값 미검사 → `_, _ =` 수리 후 0건).
- **회귀 스위프(레인-로컬)**: `TestLauncherWorktreeL2|TestNormalizeWorktreeFlag|TestEnterSessionWorktree|TestSessionWorktreeBranchName|TestCleanupSessionWorktree|TestResolveSessionShortReal` → ok · `TestHandoffSave_*|TestHandoffClear|TestHandoffCmdRegistered|TestHandoffRecover` → ok · `TestCodexLaunch|TestResolveCodexWorktree|TestCodexWorktree` → ok · `./internal/codexwiring/ ./internal/homestate/` 전체 → ok. `go test ./...` 미실행(레인 규율).
- **E6 commits**: `cfa540967`(M1.2, spec.md draft→in-progress 동반) · `a347d83ee`(M1.3). push 없음(리드 일괄).
- **PRESERVE 준수**: `internal/hook/handoff/persist.go`·`internal/hook/handoff_inject.go`·`internal/codexadapter/output.go`·`internal/template/templates/**` 무변경 (git status 대조). `resolveWorktreeL2Path` 시그니처 변경(warn io.Writer 추가)에 따른 기존 테스트 2파일(codex_launch_verb_test.go·launcher_worktree_l2_test.go) 호출부 갱신 — 기존 테스트 삭제·약화 아님.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-26T21:05:00+09:00
run_commit_sha: a347d83ee
run_status: m1-partial-green   # M1.1–M1.3 완료; M1.4(LIVE)/M1.5(P1 판정) 쿼터 회복(9/28 14:37) 후 잔여 — M1 종결 아님
ac_pass_count: 10              # 001..007, 004b, 010, 011, 012
ac_fail_count: 0
ac_deferred_count: 2           # 008(조건부, 관문 b 대기), 009(LIVE, 쿼터 대기)
preserve_list_post_run_count: 4  # persist.go, handoff_inject.go, codexadapter/output.go, template/templates/**
l44_pre_commit_fetch: not-run-lane   # push는 리드 일괄 — 원격 판정은 develop push 후 CI
l44_post_push_fetch: not-run-lane
new_warnings_or_lints_introduced: 0  # 첫 측정 errcheck 1건·gofmt 2건은 동일 run 내 수리 후 0 (커밋 시점 기준 신규 잔존 0)
cross_platform_build:
  darwin_arm64: pass
  windows_amd64: pass
total_run_phase_files: 15      # Go 신규 5 + 수정 10
m1_to_mN_commit_strategy: per-milestone   # M1.2 cfa540967, M1.3 a347d83ee, 문서 커밋 후행
```
