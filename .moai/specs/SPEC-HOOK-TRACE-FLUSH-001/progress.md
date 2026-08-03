# SPEC-HOOK-TRACE-FLUSH-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

- plan-phase 산출물 저작 완료: `spec.md` / `plan.md` / `acceptance.md` / `progress.md` (Tier M).
- v0.2.0(감사 반영) 기준: 요구사항 13건(REQ-HTF-001 … REQ-HTF-013),
  인수 기준 16건(AC-HTF-001 … AC-HTF-014, 006b·007E 포함).
- 미커버 REQ 0건, 고아 AC 0건. 미해소 명확화 항목 없음.
- 근거는 세션 내 통제 실험(실험 A 121회 / 실험 B 10회)의 실측값이며 `spec.md` §1.1에 기재.
- 감사 반영(iteration 2): 판정 계층의 공허 통과 7건을 제거했다. 모든 AC 명령을 미수정
  트리에서 실행해 기준선을 인용했고, 각 AC는 그로부터의 델타를 판정한다.

## §E.2 Run-phase Evidence

### 커밋

| 마일스톤 | SHA | 내용 |
|---|---|---|
| (plan) | `ac2cf8825` | plan-phase 산출물 4종 |
| M1 | `ec27db5e8` | `CloseWithTimeout` + `ErrFlushTimeout` + 미배수 건수 방출 |
| M2 | `2c88eaa8c` | `DefaultTraceFlushTimeout` 상수화 |
| M3 | `6a0320778` | `registry.Shutdown` + CLI 두 지점 defer 배선 |
| M4 | `f8aaf7ea4` | 회귀 가드 3종 |
| M5 | `c8441840d` | @MX ANCHOR→NOTE 강등 + goleak 억제 제거 |
| M6 | (이 커밋) | 이식성·품질 게이트 + 본 기록 |

### AC 판정

| AC | 판정 | 명령 | 관측 출력 |
|---|---|---|---|
| AC-HTF-001 | PASS | `-list`/`-run '^TestTraceWriterCloseWithTimeoutFlushesAll$'` | `1` / `--- PASS` |
| AC-HTF-002 | PASS | `-list`/`-run '^TestTraceWriterCloseWithTimeoutAbandonsOnBudget$'` | `1` / `--- PASS` |
| AC-HTF-003 | PASS | `git diff` 삭제 함수 수 / `FlushPending` 존치 / 패키지 | `0` / `1` / `ok` |
| AC-HTF-004 | PASS | 상수 grep / 배선 함수 2종 / registry.go | `2` / `0` / `0` / `0` |
| AC-HTF-005 | PASS | `-list`/`-run` noop+idempotent 2건 | `2` / 둘 다 `--- PASS` |
| AC-HTF-006 | PASS | `Dispatch` 지점 수 / `defer .*Shutdown()` | `2` / `2` |
| AC-HTF-006b | PASS | `Registry` 메서드 수 / shutdown·close 언급 | `3` / `0` |
| AC-HTF-007 | PASS | `-list`/`-run '^TestRegistryShutdownFlushesLastHandlerEntry$'` | `1` / `--- PASS` |
| AC-HTF-007E | PASS | `-list`/`-run '^TestHookCommandFlushesLastHandlerEntry$' -count=3` | `1` / `--- PASS` ×3 |
| AC-HTF-008 | PASS | 반증 왕복 4개 출력 (아래 별도 절) | 사보타주 시 둘 다 `FAIL`, 원복 후 둘 다 `ok` |
| AC-HTF-009 | PASS | 프로덕션 호출자 grep | `4` (기준선 `0`) |
| AC-HTF-010 | PASS | `fan_in=24` / 거짓 서술 / `@MX:NOTE` / `@MX:WARN` | `0` / `0` / `1` / `1` |
| AC-HTF-011 | PASS — 분기 (a) 제거 | 억제 호출 라인 grep / 패키지 | `0` / `ok` |
| AC-HTF-012 | PASS | `GOOS=linux\|darwin\|windows go build ./...` | `LINUX_OK` `DARWIN_OK` `WINDOWS_OK` |
| AC-HTF-013 | PASS-WITH-DEBT | 테스트 / vet / lint | FAIL `0` / 출력 없음 / lint 1건(외부 귀속, 아래) |
| AC-HTF-014 | PASS | `-list`/`-run '^TestCloseWithTimeoutReportsUndrainedCount$'` | `1` / `--- PASS` |

미커버 REQ 0건. `git status --porcelain internal/template/` → 출력 없음.

### AC-HTF-008 반증 왕복 (요구된 4개 출력)

전제(0단계): `git status --porcelain -- internal/hook/registry.go internal/cli/hook.go` → 빈 출력.

사보타주 후(2단계):

```
--- FAIL: TestRegistryShutdownFlushesLastHandlerEntry (0.00s)
    shutdown_test.go:87: read trace file …/trace-shutdown-flush-last.jsonl: no such file or directory
--- FAIL: TestHookCommandFlushesLastHandlerEntry (3.33s)
    hook_flush_test.go:141: run 1: trace holds entries from 0 handlers [], want at least 3 …
    (run 2·3·4 동일)
```

가드 1E의 실패 양상이 §1.1 실험 A와 정확히 일치한다 — 파일은 만들어졌으나 항목이 0건이다.
run 0만 경쟁에서 이겨 통과했고 나머지 4회가 실패했다. 단발 실행에 의존하지 말라는
AC-HTF-007E의 `-count` 규율이 실측으로 정당화되었다.

원복(3단계)은 `git restore --source=HEAD --worktree -- internal/hook/registry.go internal/cli/hook.go`
로 수행했다(stash 미사용, 경로 한정).

원복 후(4단계):

```
--- PASS: TestRegistryShutdownFlushesLastHandlerEntry (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/hook	0.335s
--- PASS: TestHookCommandFlushesLastHandlerEntry (4.77s) ×3
ok  	github.com/modu-ai/moai-adk/internal/cli	15.170s
```

### 배수 소요 실측 (§E7 보고 의무)

일회용 측정 하네스로 반복 200회씩 측정한 뒤 삭제했다(로컬 SSD, macOS).

| 적재 건수 | p50 | p90 | p99 | max |
|---|---|---|---|---|
| 1 | 62µs | 82µs | 142µs | 317µs |
| 3 | 99µs | 111µs | 140µs | 159µs |
| 10 | 225µs | 252µs | 288µs | 380µs |
| 30 | 660µs | 787µs | 952µs | 1.19ms |
| 100 (채널 용량) | 2.33ms | 2.90ms | 3.39ms | 5.30ms |

**판단: 200ms 유지.** 실제 디스패치가 만드는 3건 기준 p99가 140µs이고, 채널을 가득
채운 최악 케이스도 max 5.3ms다. 200ms는 그 최악값의 약 40배 여유이며 가장 빡빡한 훅
timeout(5초)의 약 4%다. 예산은 정상 경로에서 지불하는 비용이 아니라 상한이므로
(배수가 끝나는 즉시 반환) 넉넉한 값의 대가가 없고, 느린 파일시스템에 대한 여유만 남는다.
측정이 로컬 SSD 기준이라는 한계는 REQ-HTF-013의 미배수 건수 신호가 운영 중 보완한다.

### 사전 존재 부채 (본 SPEC 귀속 아님)

`golangci-lint run` 1건: `internal/cli/launcher.go:146 errcheck`. 본 SPEC이 건드리지 않은
파일이며, 병렬 세션(SPEC-PROFILE-MEMORY-001)의 미커밋 편집이 `os.Stderr`를 변수
`launcherStderr`로 바꾸면서 errcheck 기본 제외에서 벗어나 표면화된 것이다
(실측: 기준선 `c907db541`에 `launcherStderr` 0건, 현재 워킹트리 5건).
본 SPEC 편집 파일 대상 lint는 `0 issues`.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-02
run_commit_sha: 30f52691f
run_status: complete
ac_pass_count: 16
ac_fail_count: 0
ac_pass_with_debt_count: 1   # AC-HTF-013 — 외부 귀속 lint 1건
preserve_list_post_run_count: 0
l44_pre_commit_fetch: performed
l44_post_push_fetch: not-applicable   # 미푸시 (main 직접 푸시 차단)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  linux_amd64: ok
  darwin_arm64: ok
  windows_amd64: ok
total_run_phase_files: 9
m1_to_mN_commit_strategy: per-milestone
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-02
sync_commit_sha: 9d976c95b
sync_status: complete
b12_self_test_a: pass          # grep -c 'SPEC-HOOK-TRACE-FLUSH-001' CHANGELOG.md → 0 (no duplicate)
b12_self_test_b: pass          # acceptance.md AC 총수 16 (heading grep '^## AC-HTF' → 16; 토큰 grep은 007E·006b를 놓쳐 14) — CHANGELOG 기재 16과 일치
b12_self_test_c: pass          # CHANGELOG 인용 경로 전수 ls 확인 (writer.go / defaults.go / registry.go / hook.go / shutdown_test.go / hook_flush_test.go / main_test.go / retention.go / deps.go / launcher.go)
changelog_entry_position: "[Unreleased] → ### Fixed, 섹션 최상단"
frontmatter_status_transitions:
  spec.md: "in-progress → completed"
  plan.md: n/a                 # 프론트매터·상태 마커 없음 (grep '^status:|Status:' → 무출력)
  acceptance.md: n/a           # 동일
  progress.md: n/a             # 동일
docs_updated:
  readme: none                 # 사용자 표면 변경 없음 (내부 런타임 동작)
  docs_site: none              # docs-site/content 에 trace/observability 언급 0건
  template_tree: none          # git status --porcelain internal/template/ 무출력
mx_tag_validation:
  demotions: 1                 # writer.go Close: @MX:ANCHOR → @MX:NOTE (실측 fan_in 0 → 3 미만)
  preserved: 1                 # NewTraceWriter 의 고루틴 수명 @MX:WARN 보존
sync_verification:
  go_build: exit 0
  go_vet: exit 0               # ./internal/hook/... ./internal/cli/...
  spec_lint: "no findings"
  named_guards: PASS           # TestRegistryShutdownFlushesLastHandlerEntry / TestHookCommandFlushesLastHandlerEntry
  goleak_exemptions: 0
pushed: false                  # main 직접 푸시 차단 (enforce_admins: true) — PR 은 별도 인간 게이트
origin_divergence_at_sync: "1 28"   # origin/main 이 1 커밋 앞섬 (998744216, PR #1285). 해당 커밋도 CHANGELOG.md 를 편집하므로 PR 시 CHANGELOG 충돌 가능
```

### 후속 항목 (본 SPEC 범위 밖 — 구현하지 않음)

1. `internal/cli/deps.go` `enableObservabilityIfConfigured` 가 `observability.yaml` 의 **존재**만 검사하고 `enabled` 키를 읽지 않는다 → `enabled: false` 로도 트레이싱이 꺼지지 않는다. 수정 시 AC-HTF-007E 픽스처가 `enabled: true` 를 담도록 함께 갱신해야 한다(그러지 않으면 가드 1E 가 조용히 공허해진다).
2. 미검증 훅 4계열(`security-*`, sync-gate, workflow 훅 5종, ci-watch) 재조사 — 관측성이 신뢰 가능해진 지금에야 의미를 가진다.
3. 플러시 예산 측정 공백: macOS 로컬 SSD 단일 환경. 네트워크 파일시스템·Windows·부하 상태 미측정. REQ-HTF-013 의 `undrained_entries` 가 운영 중 이 잔여를 보완한다.
4. 가드 1E(`TestFlushBarrierHasProductionCaller` 포함)는 핸들러 **3개 이상**이라는 하한을 단언할 뿐, 프로덕션 배선에서 도출한 정확한 수를 단언하지 않는다.
5. `internal/cli/launcher.go:147` errcheck 1건 — 병렬 세션(SPEC-PROFILE-MEMORY-001, `525b3c723`/`90dca4f4c`)이 `os.Stderr` 를 `launcherStderr` 로 바꾸며 표면화. 본 SPEC 8커밋 중 해당 파일을 건드린 것은 없다(기준선 `c907db541` 에 `launcherStderr` 0건, 현재 5건). 소유는 병렬 세션.
6. `writer.go` `Write` 앞 `@MX:ANCHOR fan_in=20` 역시 미검증 수치 — 본 SPEC 승인 범위는 `Close` 주석뿐이었다.
