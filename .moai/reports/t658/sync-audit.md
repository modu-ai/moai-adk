# SPEC-TODO-QUEUE-HOME-CANON-001 — sync-audit 보고서 (카드 t658)

- 감사자: sync-auditor (독립 감사 — run/sync 수행 주체와 분리)
- 판정 기준 트리: worktree `t658`, branch `WT-home-queue-canonical`, HEAD `a2590d81a`
- 감사 범위: `git diff e37659c00..HEAD` (foreman SKILL.md 템플릿+로컬, catalog.yaml, 신규 가드 테스트 2종, 기존 watch 테스트 2종 수정, CHANGELOG, SPEC 산출물)
- 모든 근거는 본 감사 세션에서 직접 실행한 명령의 원문 출력이다 (VCI §1·§2).

## Evaluation Report

SPEC: SPEC-TODO-QUEUE-HOME-CANON-001
Overall Verdict: **PASS**

### Dimension Scores

| Dimension | Score | Verdict | Evidence (본 감사 실측) |
|-----------|-------|---------|--------------------------|
| Functionality (40%) | 100/100 | PASS | 6개 AC 전건 재관측 — 아래 §AC 재검증 상세. must-pass |
| Security (25%) | 100/100 | PASS | docs 문구 변경뿐, 신뢰경계 변화 없음. 테스트가 canary HOME + fixture 국소 MOAI_HOME 고정(운영자 home 미접촉 실측: `foreman_queue_statement_test.go:101-111`). 템플릿 중립성 grep 0건. must-pass |
| Craft (20%) | 95/100 | PASS | `go vet ./internal/kanban/` exit 0, `gofmt -l` 빈 출력. 두 가드 모두 변이 제어 RED 재관측 완료 + swept-set 계수(1218→1219) + 하한 단언(`scanned < 100`)으로 공허 초록 차단. 차감: 비ASCII basename byte↔rune 근사(F1), --separate-git-dir 등가 미측정(F2) |
| Consistency (15%) | 100/100 | PASS | 스킬 편집이 기존 목록 항목 서식·cksum 폴 유지. Conventional Commits + 카드 id + `🗿 MoAI`·`Authored-By-Agent` 트레일러. 3-phase close 정확 (아래 §close). `MOAI_TEST_QUEUE_DIR` 생산 리더 0건 실측 |

**조화평균: 98.7/100** — must-pass 방화벽(Functionality·Security) 양쪽 통과, 전 차원 기준 초과.

## AC 재검증 상세 (감사의무 1)

### AC-001 — 스킬이 정준 위치를 이름붙임 — PASS

```
$ grep -n 'state/todo' internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md
(no output) grep-exit=1
$ grep -c 'state/todo' .claude/skills/moai-kanban-foreman/SKILL.md
0
```
프래그먼트는 `${MOAI_HOME:-$HOME/.moai}/db/<project-key>/todo` 형태를 이름붙인다 (SKILL.md:98-106).

### AC-002 — 템플릿 우선 패리티 — PASS

```
$ diff -q .claude/skills/moai-kanban-foreman/SKILL.md internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md
diff-exit=0 (byte-identical)
```
- catalog 해시 갱신은 스킬 편집과 **같은 커밋** `d26091f5b`에 함께 실렸음 (`git show d26091f5b --stat` 실측 — catalog.yaml + 양쪽 SKILL.md 동봉).
- 카탈로그 해시 검증: `go run ./internal/template/scripts/gen-catalog-hashes.go --dry-run --entry moai-kanban-foreman` → `9e2ebb720734e002fcb47c3bb082a5694c368376a48405a01de079c379e66831` = catalog.yaml:39 값과 일치. 주의: 스킬 항목 해시는 파일 sha256이 아니라 **디렉터리 전체 트리 해시**(ComputeDirTreeHash, t323)다 — 생 sha256(`74f9ba11…`)과 다른 것은 결함이 아니다.
- `make build` exit 0, 재생성된 catalog.yaml이 커밋본과 **0-drift** (`git status --porcelain` 빈 출력) — 스테일 임베드 없음.

### AC-003a — docs-match-resolution 가드 — PASS

```
$ go test ./internal/kanban/ -run 'TestForemanQueueWatchResolvesCanonicalStateDir|TestBacklogJSONLiteralStaysSeamScoped|TempOrigin' -count=1 -v
--- PASS: TestForemanQueueWatchResolvesCanonicalStateDir (1.27s)
--- PASS: TestBacklogJSONLiteralStaysSeamScoped (0.36s)
--- PASS: TestTempOrigin_* 5종 (TempOriginReason/StateDirForRoot 분기 유지)
PASS  ok  github.com/modu-ai/moai-adk/internal/kanban  3.559s
```
-run 선택자로 지정한 3계열 테스트 이름이 전부 `=== RUN`으로 관측됐다(무성 선택자 드롭 없음).

### AC-003b — repo-scope seam 스캔 — PASS

위 실행에 포함. `swept 1218 non-test Go files` 로그, 예외 2곳(`internal/kanban/state_dir.go` — seam const, `cmd/t657-merge/main.go` — t835 경계) 역방향 검증(`hits[rel]==0 → error`) 통과.

### AC-003c — 변이 제어 (감사자 직접 재관측) — PASS (RED 2회 관측)

변이 1 — 템플릿 SKILL.md에 `d=.moai/state/todo` 복원 후:
```
--- FAIL: TestForemanQueueWatchResolvesCanonicalStateDir (0.00s)
    foreman_queue_statement_test.go:92: the skill names the project-local .moai/state/todo path — the dead watch target this SPEC removed
```
변이 2 — `internal/cli/t658_scratch_violation.go` 주입(`filepath.Join(".", "backlog.json")`) 후:
```
    queue_path_seam_scan_test.go:89: internal/cli/t658_scratch_violation.go constructs the legacy queue-path literal "backlog.json" (1 occurrence(s))
    queue_path_seam_scan_test.go:99: swept 1219 non-test Go files ...; 3 file(s) carry the literal
--- FAIL: TestBacklogJSONLiteralStaysSeamScoped (0.43s)
go-test-exit=1
```
양변 모두 원복 후 재실행 → `ok ... 1.063s`, `git status --porcelain` 빈 출력(트리 원복 확인). 주입 파일의 swept 계수 1218→1219 증가가 계수의 생존까지 함께 증명했다.

### AC-004 — 처분 원장 — PASS

progress.md §E.2 M1: B1-B9 전 행에 disposition 보유(`keep-read-only`/`keep`/`boundary-exception`/`already-converged`), 공란 0행. B6은 t835 소유 boundary-exception으로 seam 가드 예외와 일치. run 단계 신규 발견 없음.

### AC-005 — 유지 fallback 무변경 — PASS

`git diff e37659c00..HEAD -- internal/kanban/state_dir.go` → **0행** (state_dir.go:29 temp-guard, :33 fail-open 분기 byte-unchanged). `TempOrigin` 5/5 PASS.

### AC-006 — 신규 거짓 문구 정화 — PASS

```
$ grep -rn "state/todo\|state/kanban" internal/template/templates/
→ 1 hit: templates/.moai/docs/todo-queue-storage.md (t704 소유, REQ-006 명시 제외)
```
정화 대상 0건 — 정화할 행 없음으로 원장 기록과 일치.

## load-bearing 주장 독립 검증 (감사의무 2)

스킬 프래그먼트의 파생 논리를 실제 git 저장소 + 링크드 워크트리 피처에서 직접 실행:

```
피처: /tmp/t658-audit-primary (git init) + 연결 워크트리 /tmp/t658-audit-wt
실행 위치: /tmp/t658-audit-wt (워크트리 쪽)
출력: /tmp/t658-audit-moaihome/db/t658-audit-primary-71cf6bad/todo
```

- 키 해시 검증: `printf '%s' /private/tmp/t658-audit-primary | sha256sum | cut -c1-8` → `71cf6bad` 일치. 프래그먼트가 **심볼릭 링크 해소된 정준 루트**를 해시했음을 뜻한다.
- git 측 정규화 일치: `git -C /tmp/t658-audit-primary worktree list --porcelain` 첫 항목 = `/private/tmp/t658-audit-primary` — git 자체가 해소된 경로를 출력하며, Go `CanonicalProjectRoot`의 `EvalSymlinks`+`Clean`과 동일 정규화. (`pwd -P` 이중 정규화도 합의.)
- 워크트리 수렴: 워크트리에서 실행했음에도 키 basename이 `t658-audit-primary`(주 체크아웃) — 프래그먼트가 주 체크아웃 큐를 감시하는 REQ-001의 본질을 직접 관측.
- resolver 측 등가는 가드 테스트가 본 감사 실행에서 PASS로 담보 (StateDirForRoot == 프래그먼트 출력).

## CHANGELOG·3-phase close 정합 (감사의무 4)

- CHANGELOG 항목(카드 t658)은 본 감사 관측과 정확히 일치: 죽은 경로 서술, 해석 방식, RED→변이 RED→GREEN 서술, "6 acceptance criteria, all PASS", run commit `d26091f5b`.
- 전이 위치: `d26091f5b` = draft→in-progress (`Authored-By-Agent: manager-develop` 트레일러 확인), `da3e72ea5` = in-progress→completed (sync 커밋에 병합된 3-phase close — 별도 Mx 커밋 없음, 정규), `a2590d81a` = `sync_commit_sha` 백필(D3 면제 규정).
- plan-audit 이력(`.moai/reports/t658/plan-audit.md`): 3회 반복 → PASS 0.95, skip-eligible — run 진입 정당성 성립.

## Findings (전부 optional — blocking 0건)

- F1 [Low] [optional] internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md:103 - 프래그먼트의 키 sanitize는 byte 단위 `tr -c`, resolver(`homestate.ProjectKey`)는 rune 단위 `strings.Map`. basename에 비ASCII가 포함된 저장소에서 키가 갈라져 프래그먼트가 다른 디렉터리를 폴링할 수 있다(본 SPEC이 제거한 무성 무감시 결함의 협소한 재발 경로). progress.md Gaps에 정직하게 기록됨. Required fix(착수 시): 비ASCII 피처 테스트로 갈라짐을 고정하거나, SKILL 산문에 ASCII-basename 전제를 명기. sh에서 rune 안전 sanitize는 비용이 커서 현 단계 문서화가 합리적 판단.
- F2 [Low] [optional] internal/kanban/foreman_queue_statement_test.go - `--separate-git-dir` 체크아웃에 대한 프래그먼트↔resolver 등가는 미측정(Gaps 기록). 정적 독해상 양측 모두 체크아웃 루트를 내놓아 일치가 기대되지만, 그것은 측정이 아니라 가설이다. Required fix: separate-git-dir 피처 1건 추가.
- F3 [Low] [optional] Windows(git-bash)에서 프래그먼트의 `cksum`/`sha256sum` 가용성 미측정 — 가드는 LookPath로 건전하게 skip하며 CI darwin/linux가 판정면. progress.md Gaps 기록과 일치.
- F4 [Info] [optional] 감사 세션의 프래그먼트 수동 실행은 가드가 복합 git·HOME 주입 명령을 거부하여 /tmp 스크립트 본체에 env 고정을 넣는 형태로 수행됐다(t586 교훈의 "스크립트 파일은 가드를 우회한다" 해당). 실행 내용은 /tmp 피처 전용으로 이 트리에 영향 없음을 명시한다.

## Gaps (감사자가 관측하지 못한 것)

- 전체 스위트(`go test ./...`) 로컬 미실행 — CLAUDE.local.md §4 레인-로컬 검증 규율 준수. 전체 판정은 develop 일괄 push 후 origin/develop CI의 몫.
- Windows 레그 미측정(F3).
- 패키지 coverage % 재측정 안 함 — 생산 Go 코드 변경 0행(state_dir.go 포함 전부 무변경, 테스트+md+catalog만 변경)이므로 생산 coverage는 정의상 불변.
- 수동 프래그먼트 검증은 ASCII basename 피처에서 수행(F1 범위 밖 미관측).

## Residual-risk

- F1의 비ASCII basename 갈라짐은 해당 저장소 클래스에서 무감시 감시를 재현할 수 있으며, 현 가드(ASCII 피처)는 잡지 못한다. 기록된 한계로 남는다.
- SKILL.md 재구조화 시 docs-match 가드는 fence/`last=init` 센티넬로 요란하게 실패한다(조용한 무효화 없음) — 가드 생존성 양호. seam 가드도 예외 역방향 검증을 갖는다.

## Recommendations

- 후속 카드(신규 등록 불요, t705/t706 계열에 흡수 가능): 비ASCII basename 키 갈라짐(F1)을 피처 테스트로 고정하고 당면 동작을 결정.
- develop 일괄 push 시 origin/develop CI 전체 판정을 리드가 읽어 카드 done 처리(레인 규율대로 본 감사는 로컬 근거만 확정).

🗿 MoAI
