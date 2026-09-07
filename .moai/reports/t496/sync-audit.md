# sync-audit — SPEC-CODEX-EVENT-COVERAGE-001 (card t496, lens `--deep`)

- Audit date: 2026-09-07 · Auditor: sync-auditor (independent, lane-10 dispatch)
- Baseline-attribution: worktree `.claude/worktrees/t496`, branch `WT-codex-hook-events`, HEAD `732609dcf`, tree clean — 이 리포트의 모든 명령은 이 트리·이 HEAD에서 이번 감사 세션에 직접 실행해 관측한 출력이다. §E.2 인용치를 재사용한 항목은 없다(전 항목 재측정).

## Verdict

**PASS-WITH-DEBT — 97.4/100** (harmonic mean). 필수 통과 축(Functionality, Security) 양축 독립 통과. 부채 2건은 산출물(코드·문서·측정)의 결함이 아니라 마감 장부 보수 1건(F1, 수리 권고)과 요약 계수 정정 1건(F2, 공개된 채택 판정)이다.

## Dimension Scores

| Dimension | Score | Verdict | Evidence (verbatim) |
|-----------|-------|---------|---------------------|
| Functionality (40%, must-pass) | 100/100 | PASS | 전 11개 AC 재측정 PASS — 아래 AC 재검증 표 |
| Security (25%, must-pass) | 100/100 | PASS | 실 `~/.codex` 0변경 독립 재확인 + evidence 비밀 스캔 0 |
| Craft (20%) | 95/100 | PASS | `coverage: 87.3% of statements` / `coverage: 88.2% of statements` (≥85%), `go vet` rc=0, `gofmt -l` 공출력, `golangci-lint` `0 issues.` |
| Consistency (15%) | 95/100 | PASS | 기존 파일 관습 부합(doc comment en, `%w` 래핑, 테이블 기반 테스트), Conventional Commit + 카드 id, docs-site 4-로케일 패리티(hugo rc=0, `warn/error` grep 0) |

## AC 재검증 (본 감사가 직접 실행, 전 항목 재측정)

| AC | 판정 | 본 감사가 실행한 명령 | 관측 출력 (verbatim) |
|----|------|----------------------|---------------------|
| AC-CEV-001 (M1 시점) | PASS(역사 재측정) | `git show b4653524b:internal/codexadapter/events.go \| awk '/^var EventTable/,/^}/' \| grep -c 'true},'` + `'false},'` | `6` / `6` — M1 커밋 트리에서 합계 12·false=6 성립 |
| AC-CEV-002 | PASS | `grep -rn 'EventInterrupt' internal/ \| grep -v _test \| grep -v codexadapter` | 무출력, `rc=1` |
| AC-CEV-003 | PASS | `go test -count=1 ./internal/codexadapter/ -run 'TestResolve' -v` | `--- PASS: TestResolveInterruptNoCounterpart (0.00s)` 포함 5/5 PASS, `ok ... 0.408s` |
| AC-CEV-004 | PASS | `go test -count=1 ./internal/codexadapter/ ./internal/codexwiring/` | `ok ... codexadapter 0.947s` / `ok ... codexwiring 0.430s` (캐시 없는 fresh) |
| AC-CEV-005 | PASS | `go test ./internal/codexwiring/ -run 'TestRenderHooks_InterruptNeverInstalled' -v` | `--- PASS: TestRenderHooks_InterruptNeverInstalled (0.00s)` |
| AC-CEV-006 | PASS | `grep -c 'All eleven' ...` / `grep -c 'never an absence of' ...` | `0` / `0`; 갱신된 doc block이 Interrupt + "no MoAI dispatcher counterpart" 기재 (events.go:47-68) |
| AC-CEV-010 | PASS | 캠페인 기록 §2 6행 대조 + `evidence/captures/*.jsonl`·`evidence/runs/*` 실물 대조 | 6/6 행이 판정+명령+관측 운반. FIRED 3(SubagentStart/SubagentStop/Interrupt) 캡처 실물 확인 — `SubagentStop.jsonl`에 `agent_id`·`agent_type:"default"`·`last_assistant_message:"4"`, `Interrupt.jsonl`에 `hook_event_name:"Interrupt"`(모두 transcript_path가 tmp home 내부) |
| AC-CEV-011 | PASS | §0-§1 대조 + 독립 정밀검증(아래 Security 행) + `compact2.err` 직독 | `CODEX_HOME` SUPPORTED(`p0.rc`=0); `Error: ... Input exceeds the maximum length of 1048576 characters. ... "actual_chars":1583027` verbatim; perm 3런 `grep -c -i 'approval\|permission'` = `0/0/0` |
| AC-CEV-012 | PASS | `collab.jsonl` + `SubagentStop.jsonl` 실물 | `collab_tool_call` 아이템 + SubagentStop 캡처 1행 — 0.147.0 관측이 인용이 아니라 0.153.4 재측정으로 뒤집혔음을 확인 |
| AC-CEV-013 | PASS | 캠페인 기록 §4 처분표 대조 | adapt-now 2(SubagentStart/Stop→M3 착지) · follow-up 2(Interrupt, PermissionRequest) · trigger-not-achieved 2(PreCompact/PostCompact) — 미처분 0행 |
| AC-CEV-020 (분지 b) | PASS | `awk ... grep -c 'true},'` + `'false},'` (현재 트리) | `8` / `4` — 합계 12, adapt 2행 flip + `TestAdaptedRowCount` GREEN |

- DoD#4 (REQ-7 불변): `git diff --stat ace1c5440..HEAD -- internal/hook/` → **공출력** (0행). 전체 diff도 `internal/hook/` 변경 0개로 확인.
- M3 RED 근거 검증: `git show b4653524b:internal/codexadapter/events_test.go`에 `TestAdaptedRowCount`가 `wantAdapted = 6`으로 M1 시점부터 존재 — §E.2의 "adapted rows = 6, want 8" RED가 실제 테스트의 pre-GREEN 상태임을 두 커밋 상태로 구조 확인.
- 디스패처 사전 존재: `git show ace1c5440:internal/cli/hook.go \| grep -c '"subagent-start"\|"subagent-stop"'` → `2` — M3가 디스패처 변경 없이 테이블 flip+렌더러 도출만으로 성립함을 확인(이것이 diff가 작은 이유이고, `TestDispatcherArgsExist` GREEN과 정합).

## Security (25%) — 세부

- **실 홈 무변경, 독립 재확인**: zero-write-verdict.txt가 올린 캠페인 창 내 rollout 8개 중 표본 1개를 직접 독读 → `session_meta`가 `"originator":"moai-codex-gate"`, `"cwd":"/var/folders/.../TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey..."` — 기록 §1의 병렬 레인 귀속 주장과 정확히 일치. 전체 8개 + 창 인접 rollout에 `grep -l 't496'` → `rc=1` (0매치). 3핵심 파일(config.toml/auth.json/hooks.json) 해시·mtime이 판정서에 기록돼 있고 mtime은 캠페인 이전.
- **§0 공시 정확성**: 첫 P0(d4IY5wt6)가 실홈으로 실행된 사실이 측정창 이전 사건으로 공시돼 있고, config 미변경 기술과 `originator: "codex_exec"` 검증 절차가 기록돼 있다. 공시 내용과 §1 결론("0 campaign-attributable home writes")이 상충하지 않는다.
- **비밀 스캔**: `grep -riE 'OPENAI_API_KEY|api_key|sk-...|Bearer|auth_token|"token"|PASSWORD'` over `evidence/` + 기록 + 스크립트 → 0매치. `evidence/config.toml`은 model+project trust 최소본, `evidence/hooks.json`은 로거 명령만 — auth는 symlink 레시피(사본 없음)로 실제로 지켜졌음이 evidence로 확인.
- `--dangerously-bypass-hook-trust`·`--skip-git-repo-check`는 캠페인 런 플래그로만 존재하고 제품 코드·설치 표면에 새 위험을 만들지 않는다(설치되는 것은 `moai hook subagent-start/stop --harness codex`로, 베이스부터 등록된 디스패처 하위커맨드).

## Docs close (sync 커밋 732609dcf)

- CHANGELOG: 신규 불릿 1개 추가, `SPEC-LEAD-DEPUTY-001` 불릿 포함 기존 행 삭제 0(diff에 `+` 2행뿐 — both-keep 성립).
- docs-site: 4-로케일(en/ko/ja/zh) `codex-dual-harness.md` 각 7개 헤딩·동일 구조, 0.153.4 기준, SubagentStart/Stop adapted+RenderHooks 설치 문구, "trigger-not-achieved ≠ does not fire" 정직 서술이 4개 로케일 모두에 존재.
- hugo 빌드 재실행(본 감사): `hugo --gc` → `Total in 2639 ms`, rc=0, `grep -ci 'warn\|error'` → `0`; `public/{en,ko,ja,zh}/advanced/codex-dual-harness/index.html` 4개 + `public/sitemap.xml` 생성 확인.
- README 4-로케일: codex hook-event 열거 부재(`grep -l` rc=1) — §E.4의 갱신 불요 판정 재확인.
- Frontmatter: sync 커밋에서 `status: in-progress → completed` (커밋 전 `in-progress` → 커밋 후 `completed`, `git show` 양측 직독).

## Findings

- **F1** [LOW] [optional-repair-recommended] `.moai/specs/SPEC-CODEX-EVENT-COVERAGE-001/progress.md:70` — §E.4 `sync_commit_sha: PENDING_THIS_COMMIT`. 판정: **수리 가능한 갭** (close contract의 실질은 충족, 문자는 미충족). 바인딩 자체는 커밋 메시지+파일 내 각주+병합 트리로 증명 가능하나, (a) 정식 placeholder 토큰 계열(`pending-backfill`/`pending-backfill-sync`, spec-frontmatter-schema § D3)이 아니고 (b) D3가 정한 후속 커밋 backfill이 아직 착지하지 않았다. 필수 수리: 후속 커밋 1개로 `sync_commit_sha: 732609dcfb36d67b598b4d1ee9c7190aac1a893f` backfill.
- **F2** [LOW] [optional] `progress.md:56` — §E.3 `ac_pass_count: 10`이 옆 줄의 열거(001..006, 010..013, 020 = 11)와 어긋남. §E.2의 AC별 PASS 행은 전부 옳고 완전하므로 요약 계수만의 오차. §E.4 각주에 공시돼 있음. 판정: **accept-as-disclosure 가능** — 단, 수리할 경우 §E.3은 run-phase 소유 구역이므로 sync 측 수정이 아니라 run-phase 소유자 경유 1줄 정정(10→11)으로 라우팅할 것. F1 수리 커밋에 같이 태우려면 소관 경유를 명시.
- **F3** [INFO] [optional] `codex-event-campaign.md` §2 — compact3 입력을 "~960,000 chars"로 표기했으나 실측 `.size` 파일은 `989855` 바이트. 하중 숫자(702,974 / 1,583,027 / 264,808 토큰)는 전부 정확. 산문 근사치 1곳, 수리 불요.
- **F4** [INFO] [optional] 기록 §1 산문이 창 내 신규 8 rollout은 명시 귀속하지만 `.tmp/git-*/HEAD` 2건 + arg0 lock 교체 1건은 명명하지 않는다(판정서 raw 목록에는 존재). 캠페인은 `CODEX_HOME=tmp`로 실행돼 이 파일들을 쓸 경로가 없고 시각도 병렬 레인 세션들과 일치 — 결론 반전 없음. 기록 완결성 메모.
- **스타일 게이트 판정 (dispatch 항목 7)**: gopls 제안 `slices.Contains` (internal/codexwiring/hooks_test.go:56 부근의 동등 루프). `golangci-lint run ./internal/codexadapter/... ./internal/codexwiring/...` → `0 issues.` rc=0 — **게이트가 플래그하지 않음 → optional polish, non-gating.**

## Repair list (compact)

1. **R1 (권장)** — 후속 커밋: progress.md §E.4 `sync_commit_sha` → `732609dcfb36d67b598b4d1ee9c7190aac1a893f` (D3 backfill).
2. **R2 (선택)** — run-phase 소유자 경유 1줄: §E.3 `ac_pass_count: 10` → `11`.
3. R3 이상 없음 — F3/F4는 기록 메모 수준.

## Gaps (본 감사가 관측하지 않은 것)

- codex 캠페인을 재실행하지 않았다(dispatch가 명시한 optional 사항). 대신 기록의 귀속 주장을 session_meta 독读 + 전 rollout 마커 스윕으로 독립 정밀검증했다.
- `go test ./internal/cli/ -run 'Codex|Hooks'` (§E.2 M3행, 45.954s) 재실행 안 함 — 미변경 패키지이고(전체 diff에 internal/cli 없음), 등록 사전존재는 베이스 커밋에서 직접 확인했다.
- `homecheck/*.gz` 원본 스냅샷 6개를 압축 해제해 재차 diff하지 않았다 — 유도본 zero-write-verdict.txt 판독 + 독립 정밀검증으로 갈음. 유도본의 3-스냅샷 대조 방법론 기술은 내부 정합.
- `go build` 크로스 플랫폼 미실행 — §E.3과 동일(빌드 시스템 파일 무변경), 최종 판정은 origin/develop CI의 몫(리드 일괄 push 대기).

## Residual-risk

- PreCompact/PostCompact/PermissionRequest의 판정은 "trigger-not-achieved"로 남는다(공시됨) — 대화형 세션 트리거 검증은 후속 카드 사항이며, 향후 실측이 이 판정을 뒤집을 수 있다.
- SubagentStart/SubagentStop 적응은 다음 `RenderHooks`/wire 시 사용자 `.codex/hooks.json`에 새 핸들러를 설치한다 — 이것이 본 SPEC의 목적이고 발화+페이로드가 실측됐지만, 사용자 표면 행태 변화는 첫 배포 후 관측 지점이다.
- 이 브랜치는 아직 origin/develop에 push되지 않았다(레인 프로토콜상 리드 일괄). CI 통합 판정은 병합 후 확정된다.
