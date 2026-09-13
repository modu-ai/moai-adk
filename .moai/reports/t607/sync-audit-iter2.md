# sync-audit 2회차 — SPEC-RESOURCE-SLOT-LEASE-001 (카드 t607)

- **판정: PASS**
- **점수: 83.35 / 100** (기본 평가 프로필 `.moai/config/evaluator-profiles/default.md`, flat 가중)
- **막는 결함: 0건.** 1회차의 막는 결함 F1·F2 는 둘 다 닫혔다. 남은 것은 비차단 2건(F3·F4, 영구 이탈로 기록됨)과 선택 4건(F5·F6·F7 + 이번에 새로 찾은 F8).
- 감사 트리: `WT-heavy-test-slot`, HEAD `3eb94228407efeadd68faa6ea19a1426a7b12293`, 카드 base `8d42587e695aee97cb4454e5efa564cd613343d3`. 1회차 감사 커밋은 `ece65109a`.
- 이번 회차의 범위: 리드 지시대로 **F1~F4 수리 검증 + F5~F7 존치 확인**에 한정했다. 1회차가 이미 건전하다고 판정한 것은 수리가 건드린 곳 외에는 다시 다투지 않았다.
- 판정 도구 provenance: `./bin/moai`(Commit `9c288daa9`). `git diff --stat 9c288daa9 HEAD -- internal/spec` → **출력 없음**을 이번 실행에서 내가 직접 확인했으므로, 이 바이너리가 싣고 있는 lint 엔진은 HEAD 의 소스와 같다(`verification-claim-integrity.md` §2.2 의 두 번째 좌표 충족). Go 툴체인·`golangci-lint` 는 이 세션의 것.

## 0. heavy-test 슬롯 준수

`internal/cli` 를 컴파일·링크하는 명령은 하나도 돌리지 않았다. 돌린 네 패키지는 실행 **전에** 의존을 재고 들어갔다.

```text
$ for p in ./internal/config ./internal/kanban ./internal/hook ./internal/template ./internal/template/...; do
      go list -test -deps $p | grep -c 'moai-adk/internal/cli$'; done
./internal/config 0
./internal/kanban 0
./internal/hook 0
./internal/template 0
./internal/template/... 0
```

`./bin/moai` 는 이미 만들어져 있는 바이너리를 경로로 부른 것이라 컴파일이 일어나지 않는다. `make build`·`go build ./cmd/moai`·`go test ./internal/cli/...`·`go test ./...` 는 돌리지 않았다.

## 1. 차원 점수

| 차원 | 점수 | 판정 | 증거 (이번 실행, 이 트리 HEAD `3eb942284`) |
|---|---|---|---|
| Functionality (40%) | 75/100 | PASS (must-pass 충족) | `go test ./internal/kanban/ -count=1` → 종료 코드 **0**, `ok  github.com/modu-ai/moai-adk/internal/kanban 157.798s`. `go test ./internal/hook/ -count=1` → **0**, `ok … internal/hook 154.034s`. 필터 없는 전체 패키지 판정이다. 1회차 대비 AC 자체는 달라진 것이 없고 `internal/cli` 공백도 그대로라 점수를 올리지 않았다. |
| Security (25%) | 95/100 | PASS (Critical/High 0) | 수리 3건 중 코드에 닿은 것은 `internal/config/testdata/shipped_key_inventory.yaml` 6줄뿐이고 프로덕션 경로는 무변경이다(`git diff --name-only ece65109a..HEAD` 7파일, 아래 §2.4). 1회차의 신뢰 경계 판정을 무너뜨린 변경이 없다. |
| Craft (20%) | 88/100 | PASS | `go test ./internal/config/ -count=1` → **0**, `ok … internal/config 2.268s` (1회차의 레드가 닫혔다). `go test ./internal/template/... -count=1` → **0**, 네 줄 전부 `ok`/`no test files`. `golangci-lint run --timeout=5m ./internal/kanban/... ./internal/hook/... ./internal/config/...` → **0**, `0 issues.`. `./bin/moai spec lint --strict …` → **0**, `✓ No findings`. 감점은 F8(새 결함)·F5·F7. |
| Consistency (15%) | 80/100 | PASS | 정정 3건이 전부 **원래 값을 남긴 채** 쓰였다(§2.2·§2.3). `acceptance.md`·`spec.md`·`plan.md` 무변경을 diff 로 확인. 감점은 F4 — `acceptance.md:263` 이 틀린 기작을 그대로 싣고 있고 정정은 `progress.md` 에만 있다(리드 결정이지만 문서 divergence 는 남는다) — 그리고 F3. |

가중 합: `0.40×75 + 0.25×95 + 0.20×88 + 0.15×80 = 83.35`.

**must-pass 방화벽**: Functionality·Security 둘 다 독립으로 통과. 막는 결함이 0 이므로 판정은 PASS.

## 2. 수리별 검증 (내가 직접 잰 것)

### 2.1 F1 — 인벤토리 등재 · class W vs R 판정

**전체 패키지 재측정(필터 없음), 전부 이번 실행·이 트리:**

```text
$ go test ./internal/config/ -count=1        → EXIT=0   ok … internal/config 2.268s
$ go test ./internal/kanban/ -count=1        → EXIT=0   ok … internal/kanban 157.798s
$ go test ./internal/hook/ -count=1          → EXIT=0   ok … internal/hook 154.034s
$ go test ./internal/template/... -count=1   → EXIT=0
ok  	github.com/modu-ai/moai-adk/internal/template	27.873s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.826s
ok  	github.com/modu-ai/moai-adk/internal/template/commandemit	1.114s
?   	github.com/modu-ai/moai-adk/internal/template/scripts	[no test files]
```

**class W vs R — 수리가 옳고, 1회차의 내 지시문이 틀렸다.**

인벤토리 파일 자신의 범례가 판별식이다(`internal/config/testdata/shipped_key_inventory.yaml:4-5`):

```text
# Classification: W=wire, P=prose-consumed, R=reserved, D=delete
# Evidence: 'reader' = Go production reader; prose file path = P consumer; 'none' = R/D
```

`R` 은 **reserved**(리더 없음)이지 read 가 아니다. 1회차 보고서 F1 의 “둘 다 실제 판독기가 있으므로 **R(read)** 가 옳다”는 문장은 범례를 잘못 읽은 것이다 — 근거로 든 사실(리더가 둘 다 존재한다)은 맞았는데 그 사실이 지목하는 등급을 반대로 적었다. **정정한다: `W` / `evidence: reader` 가 옳고 수리가 맞다.**

리더 실재를 이번 실행에서 다시 확인했다:

- `workflow.slot_lease.default_max_duration` → `internal/config/loader_slot_lease.go:19` `LoadSlotLeaseDefaultMaxDuration`, `wrapper.Workflow.SlotLease.DefaultMaxDuration` 직접 판독.
- `workflow.slot_lease.enabled` → `internal/hook/pre_tool.go:856-865` `slotLeaseConfig()` 가 `cfg.Workflow.SlotLease.Enabled` 를 판독하고, `:566` 에서 가드 분기를 연다.

그리고 **테스트 자신의 probe 가 내 판단이 아니라 기계로 이를 뒷받침한다.** `TestShippedConfigKeysHaveReaders` 는 dead/unresolved/unbound 키 640개를 `t.Logf` 로 열거하는데, 두 키는 그 목록에 **없다**:

```text
$ go test ./internal/config/ -run 'TestShippedConfigKeysHaveReaders' -count=1 -v | grep -c 'slot_lease'
0
$ …  shipped_key_reader_test.go:126: REQ-CKH-008 diagnostic: 640 triaged config key(s) are dead/unresolved/unbound …
$ …  --- PASS: TestShippedConfigKeysHaveReaders (1.47s)
```

즉 path-resolved probe 가 두 키를 live 로 분류했다. `W` 는 저자의 주장이 아니라 도구가 독립으로 도달한 결론과 일치한다.

### 2.2 인벤토리 수리가 테스트를 약화하지 않았는가 — 뮤턴트 2종

**먼저 게이트의 판별 범위를 읽었다.** `internal/config/shipped_key_reader_test.go:100-112` 의 실패 조건은 **인벤토리 멤버십 하나뿐**이다 — class 토큰(W/P/R/D)은 실패에 관여하지 않고, liveness 는 `:123-127` 의 `t.Logf` 라 실패를 만들지 않는다. 따라서 “W 로 넣어서 R 로 넣었을 때보다 약해졌는가”는 **아니다**(어느 쪽이든 기계적 효과가 같다). 약화 여부는 멤버십 축에서만 물을 수 있고, 그 축을 뮤턴트 둘로 쟀다.

**M-A — 수리가 넣은 6줄을 도로 뺀다** (`sed '2892,2897d'`, 2966줄 → 2960줄):

```text
$ go test ./internal/config/ -run 'TestShippedConfigKeysHaveReaders' -count=1   → 종료 코드 1
--- FAIL: TestShippedConfigKeysHaveReaders (1.09s)
    shipped_key_reader_test.go:109: REQ-CKH-008 anti-rot: 2 shipped config key(s) are NOT in the triage inventory …
          workflow.slot_lease.default_max_duration
          workflow.slot_lease.enabled
복구: cp 백업 → cmp 종료 코드 0
```

⇒ 수리의 6줄은 load-bearing 이다. 빼면 즉시 레드로 돌아가고, 실패가 이름 대는 키도 정확히 그 둘이다.

**M-B — 아직 어느 인벤토리에도 없는 배포 키를 새로 심는다.** 템플릿 `internal/template/templates/.moai/config/sections/workflow.yaml` 의 `slot_lease:` 블록에 `audit_probe_unread_key: zzz` 한 줄을 넣었다(리더 없음):

```text
$ go test ./internal/config/ -run 'TestShippedConfigKeysHaveReaders' -count=1   → 종료 코드 1
    shipped_key_reader_test.go:109: REQ-CKH-008 anti-rot: 1 shipped config key(s) are NOT in the triage inventory …
          workflow.slot_lease.audit_probe_unread_key
복구: cp 백업 → cmp 종료 코드 0,  git status --porcelain → 무출력
```

⇒ 가드는 **수리 이후에도** 새로 들어오는 미분류 배포 키를 잡는다. 이 카드의 두 키를 등재한 것이 “앞으로 들어올 키까지 조용히 통과시키는” 일반적 완화가 아니라는 뜻이다.

**단, 판별 범위를 정확히 적어 둔다(과대평가 방지).** 이 게이트가 잡는 것은 **미분류(untriaged)** 키이지 **미판독(unread)** 키가 아니다. 리더가 없는 키를 `W` 로 잘못 등재해도 기계는 실패하지 않고 `t.Logf` 한 줄만 늘어난다. 이 카드에서는 위 §2.1 의 probe 결과가 그 공백을 메웠지만, 그것은 기계적 보증이 아니라 이번에 확인한 사실이다.

두 뮤턴트 뒤 작업 트리는 깨끗하다(`git status --porcelain` 무출력).

### 2.3 F2 — 정정이 정정으로 읽히는가, 같은 모양의 새 주장이 들어왔는가

**읽히는가 — 그렇다.** 세 자리 모두 원래 값을 지우지 않고 남겼다.

- `progress.md:316-321` `new_warnings_or_lints_introduced: 1` 에 “이 줄은 원래 `0`이었고 그 값은 자기가 명시한 baseline(HEAD `272521e45`)에서 거짓이었다 … 원래 0은 그 테스트를 선택하지 않는 `-run` 필터 아래에서 잰 값이다”가 붙어 있다. 조용한 덮어쓰기가 아니다.
- `progress.md:520-537` 「정정 — sync-audit F1/F2가 §E.4에 남긴 것」이 CHANGELOG 원문(“**all PASS**”)을 인용한 뒤 왜 거짓이었는지를 적는다.
- `progress.md:493` `ac_remeasured_here` 의 config 줄은 **그대로 두고** 옆에 “이 초록은 해당 AC에 대한 초록일 뿐 패키지 초록이 아니다”를 달았다. 기록 삭제가 아니라 한정이다.

**CHANGELOG 재진술 — 검증.** `all PASS` 가 t607 항목에서 사라졌다. 남은 한 건은 이웃 항목 소관이다:

```text
$ awk 'NR>=24 && NR<=28 {n=gsub(/all PASS/,"&"); if(n>0) printf "line %d: %d hit(s); %s\n", NR,n,substr($0,1,60)}' CHANGELOG.md
line 28: 1 hit(s); - **[SPEC-UPDATE-ADD-CODEX-001](…)** — sync-
```

현재 t607 문구는 (a) 기준 수 16 은 유지하되 일괄 통과를 주장하지 않고, (b) 패키지 판정을 **필터 없이 잰 것만** 트리 `714bf8a7e` 와 함께 나열하며, (c) `internal/cli` 는 그 목록에 없고 판정이 없다고 명시한다. (b)의 네 패키지를 내가 HEAD `3eb942284` 에서 다시 재어 전부 종료 코드 0 을 관측했다(§2.1) — 인용된 트리와 다른 트리지만 같은 결론이고, 진술이 자기 트리를 명시하고 있으므로 귀속은 성립한다.

**같은 모양의 새 주장이 들어왔는가 — 한 건 있다(F8, 선택).** `progress.md:524` 의 “이 §E.4의 `ac_remeasured_here` **여덟 줄은 전부** `-run` 필터를 건 측정이다”는 세지 않고 쓴 일반화다. 실제 그 블록은 항목 14개이고(직접 측정 9 + 증거 판독 5), 직접 측정 9 중 `-run` 필터를 건 것은 5(`kanban`·`hook`·`config`·`IntegrationLock`·`template` neutrality)뿐이다. 나머지는 `git diff` 프로브(AC-RSL-013(c)), `grep`(AC-RSL-014 (a)-(f), AC-RSL-015), `./bin/moai spec lint` 라 `-run` 필터와 무관하다. 방향이 **자기 증거를 과소평가하는 쪽**이라 초록을 만들어 내지는 않으므로 차단하지 않는다. 그러나 1회차가 지적한 바로 그 형태(세지 않은 집합에 대한 일괄 진술)라 기록한다.

그 밖에는 없다. `progress.md` 전체에서 `전부 PASS`/`all PASS`/`모두 통과`/`무결함` 계열 적중은 위 §2.3 의 인용문 한 줄(정정 서술 안의 원문 인용)뿐이다.

### 2.4 F3·F4 기록 여부와 `acceptance.md` 무변경

```text
$ git diff --name-only ece65109a..HEAD
.moai/reports/t607/f1/f1-config.txt
.moai/reports/t607/f1/f1-hook.txt
.moai/reports/t607/f1/f1-kanban.txt
.moai/reports/t607/f1/f1-template.txt
.moai/specs/SPEC-RESOURCE-SLOT-LEASE-001/progress.md
CHANGELOG.md
internal/config/testdata/shipped_key_inventory.yaml

$ git diff --stat ece65109a..HEAD -- …/acceptance.md …/spec.md …/plan.md
(출력 없음)
```

- **F3** — `progress.md:453-455` 에 “§E.3 본문은 `170988694` 에서 작성됐고 그 시점에 `11fa72743` 이 이미 `status: completed` 를 세운 뒤였다 … **영구 이탈로 기록한다 — 이력을 다시 쓰지 않는다**”로 들어가 있다. SHA backfill 면제가 SHA 필드 한정이라는 근거도 함께 적혔다. 1회차 요구(§E.4 한 줄 기록)를 충족한다.
- **F4** — `progress.md:457-459` 에 관측된 기작(`Expired()` 가 `!Held()` 에서 false 라 만료 가지로 가지 않고 보유자가 빈 `guard-deny` 로 떨어진다)이 적혔고, `acceptance.md` 는 고치지 않는다고 명시했다. 무변경은 위 diff 로 확인.

수리 범위 밖 파일은 하나도 건드리지 않았다. 7파일 전부 F1·F2 처분과 그 증거다.

## 3. 1회차 결함별 처분

| 결함 | 1회차 등급 | 이번 회차 판정 | 근거 |
|---|---|---|---|
| **F1** 배포 키 2개 미분류 → `internal/config` 레드 | High · blocking | **닫힘** | 인벤토리 6줄 등재(`714bf8a7e`), `go test ./internal/config/ -count=1` → 0(§2.1). 뮤턴트 M-A/M-B 로 약화 없음 확인(§2.2). **class 는 `W` 가 옳고 1회차 내 지시문(`R`)이 틀렸다 — 정정함**(§2.1) |
| **F2** §E.3·CHANGELOG 의 무결함 주장이 자기 baseline 에서 거짓 | Medium · blocking | **닫힘** | `new_warnings_or_lints_introduced: 0 → 1` + 경위 주석, CHANGELOG 일괄 주장 철회 + 필터 없는 패키지 판정 나열, §E.4 config 줄 한정(§2.3). 모두 **legible correction** |
| **F3** run 신호를 sync 종결 뒤에 작성 | Medium · non-blocking | **기록됨 (영구 이탈)** | `progress.md:453-455`. 이력은 되돌릴 수 없으므로 1회차 요구가 곧 기록이었고 충족됨. 재발 방지는 다음 카드 소관 |
| **F4** `acceptance.md:263` n-held 뮤턴트 기작 오기 | Medium · non-blocking | **기록됨 · 원문 미수정 (존치)** | 관측 기작이 `progress.md:457-459` 에 적혔다. 다만 `acceptance.md:263` 자체는 틀린 채 남아 있다 — 리드 결정이나 divergence 는 존치하므로 Consistency 감점으로 반영 |
| **F5** `pid_source` 가 상수라 해석 실패를 구분 못 함 | Low · optional | **존치** | `internal/kanban/slot_lease.go:344` `PIDSource: PIDSourceSessionOwner` 무변경 확인 |
| **F6** CLI 와 가드의 루트 폴백 방향 불일치 | Low · optional | **존치** | `internal/cli/slot.go:78-81` `if root, err := kanban.ResolveSlotLeaseRoot(start); err == nil { return root }; return start` 무변경(읽기만 — 컴파일하지 않음) |
| **F7** 감사 로그 회전·상한 없음 | Low · optional | **존치** | `internal/kanban/slot_lease.go:527` `AppendSlotLeaseAudit` + `:542` `os.O_APPEND\|os.O_CREATE\|os.O_WRONLY` 무변경 |

## 4. 이번 회차의 새 결함

### F8 [Low] [optional] — 정정문이 세지 않은 일반화를 하나 들여왔다

- 위치: `.moai/specs/SPEC-RESOURCE-SLOT-LEASE-001/progress.md:524`.
- 문장: “이 §E.4의 `ac_remeasured_here` **여덟 줄은 전부** `-run` 필터를 건 측정이다.”
- 관측: 그 블록의 항목은 **14개**(`awk '/^ac_remeasured_here:/,/^gaps:/' … | grep -c '^  - "'` → 14; 직접 측정 9 + 증거 판독 5). 직접 측정 9 중 `-run` 필터를 건 것은 5뿐이고, AC-RSL-013(c)(`git diff` 프로브)·AC-RSL-014 (a)-(f)(`grep`)·AC-RSL-015(`grep`)·SPEC lint(`./bin/moai`)는 `-run` 과 무관하다.
- 왜 선택인가: 방향이 자기 증거를 **깎는** 쪽이라 거짓 초록을 만들지 않는다. 그러나 1회차가 겨눈 것과 같은 형태(집합을 세지 않고 일괄로 말하기)이므로 남긴다.
- 요구되는 수리(선택): 문장을 “직접 측정 9줄 중 `-run` 필터를 건 5줄(`kanban`·`hook`·`config`·`IntegrationLock`·`template` neutrality)은 해당 AC 에 대한 초록이지 패키지 판정이 아니다”로 좁힌다.

### F9 [Low] [optional] — F1 증거 파일이 명령도 종료 코드도 싣지 않는다

- 위치: `.moai/reports/t607/f1/f1-{config,kanban,hook}.txt`(각 1줄), `f1-template.txt`(4줄).
- 관측: 내용이 `ok  github.com/modu-ai/moai-adk/internal/config	2.294s` 같은 결과 줄뿐이다. 어떤 명령이 그 줄을 냈는지, 종료 코드가 무엇이었는지는 파일 안에 없고 커밋 메시지와 `progress.md` 표에만 있다.
- 왜 선택인가: `ok` 줄 자체가 go test 의 성공 신호이고 명령은 두 군데에 적혀 있어 재유도가 가능하다. 다만 증거 파일 단독으로는 귀속이 닫히지 않는다.
- 요구되는 수리(선택): 다음부터 증거 파일 첫 줄에 명령을, 마지막 줄에 `EXIT=<n>` 을 적는다.

## 5. 5절 증거 형식

### Claim

1. 1회차의 막는 결함 F1(`internal/config` 레드)은 이 트리에서 닫혔고, 그 수리는 게이트를 약화하지 않는다.
2. F1 수리의 class 선택(`W` / `evidence: reader`)이 옳고, 1회차 보고서가 지시한 `R` 은 범례 오독이었다.
3. F2 의 정정 3건은 모두 원래 값을 남긴 legible correction 이며, CHANGELOG 의 일괄 통과 주장은 철회됐다.
4. F3·F4 는 영구 이탈로 기록됐고 `acceptance.md`·`spec.md`·`plan.md` 는 무변경이다.
5. F5·F6·F7 은 그대로 존치하며 전부 선택 등급이다. 새 결함 F8·F9 도 선택 등급이다.
6. 따라서 막는 결함이 0 이고 must-pass 두 차원이 통과하므로 **PASS**.

### Evidence

§0·§1·§2 의 명령과 원문. 요약: 필터 없는 전체 패키지 4건 전부 종료 코드 0(`config 2.268s` / `kanban 157.798s` / `hook 154.034s` / `template …` 4줄) · `golangci-lint … → 0 issues.` · `./bin/moai spec lint --strict → ✓ No findings`(종료 코드 0) · 뮤턴트 M-A(6줄 제거 → 종료 코드 1, 두 키를 이름 댐) · M-B(미분류 배포 키 주입 → 종료 코드 1, 그 키를 이름 댐) · 두 뮤턴트 복구 `cmp` 0, `git status --porcelain` 무출력 · probe 의 640행 진단 목록에 `slot_lease` 적중 0.

### Baseline-attribution

전부 **이번 실행, 이 트리** `WT-heavy-test-slot` HEAD `3eb94228407efeadd68faa6ea19a1426a7b12293` 에서 잰 값이다. 카드 base 는 `8d42587e695aee97cb4454e5efa564cd613343d3`. `./bin/moai`(Commit `9c288daa9`)의 lint 판정은 `git diff --stat 9c288daa9 HEAD -- internal/spec` 이 이번 실행에서 빈 출력임을 확인한 뒤에만 근거로 썼다 — 도구가 실은 코드와 판정 대상 트리의 lint 엔진이 같다. `progress.md` 정정문이 인용하는 `714bf8a7e` 트리의 네 줄은 **읽은 값**이며, 나는 같은 네 패키지를 HEAD 에서 **다시 재어** 같은 결론(전부 0)을 얻었다.

### Gaps (이번 감사에서 관측하지 못한 것)

- **`go test ./internal/cli/` 전체 패키지 판정** — heavy-test 슬롯을 다른 레인이 쥐고 있어 잴 수 없다. 1회차와 동일하게 공백으로 남는다. develop push 가 일으키는 CI 가 첫 판정이다.
- **AC-RSL-003c 와 `moai slot` CLI 동작 전부** — 같은 이유. `.moai/reports/t607/m3/*.txt` 를 읽었을 뿐 재실행하지 않았다.
- **AC-RSL-014(i) `make build`** — `internal/cli` 를 링크하므로 잴 수 없다. §E.4 의 “오케스트레이터가 슬롯 안에서 쟀고 sync 에이전트가 관측한 값이 아니다”라는 표기를 읽었을 뿐이다.
- **두 플랫폼 빌드 / Windows 행동** — 같은 이유로 못 쟀고, 이 머신에 Windows 판정 수단도 없다.
- **커버리지 재집계** — 이번 회차에서는 다시 뽑지 않았다. 수리가 테스트 파일도 프로덕션 파일도 건드리지 않았으므로 1회차 값(`slot_lease.go 85.9%`, `slot_lease_guard.go 94.4%`)이 유효하다고 **추론**했을 뿐 재측정하지 않았다.
- **AC-RSL-011 뮤턴트 표 10행** — 이번 회차 범위 밖이라 다시 재현하지 않았다. 1회차가 1행을 재현하고 나머지는 `summary.tsv` 로 읽었으며, 그 상태가 그대로다.
- **F6 의 실행 확인** — `internal/cli/slot.go` 는 **읽기만** 했다. 그 폴백 경로를 실행해 본 적은 없다.

### Residual-risk

- `internal/cli` 전체 판정 공백은 F1 수리와 무관하게 남는다. 이 카드가 CLI 에 넣은 `moai slot` 표면은 단위 테스트와 M3 증거 밖에서 한 번도 판정된 적이 없다.
- 배포 키 게이트는 **미분류**만 잡고 **미판독**은 잡지 않는다(§2.2). 이 카드의 두 키는 probe 가 live 로 분류해 공백이 메워졌지만, 다음 카드가 리더 없는 키를 `W` 로 등재하면 기계는 조용하다.
- `acceptance.md:263` 은 틀린 기작을 그대로 싣고 다닌다. 정정은 `progress.md` 에만 있어, `acceptance.md` 만 읽는 사람은 여전히 잘못된 예측을 읽는다(F4).
- 가드는 이 저장소 어디에서도 켜져 있지 않다. deny 경로가 실제 사용자 명령을 거절한 장면은 단위 테스트 밖에 존재하지 않는다.
- 감사 로그 무한 증가(F7)와 CLI/가드 폴백 비대칭(F6)은 그대로다.
- 1회차 보고서 `ece65109a` 는 F1 의 수리 지시를 `R` 로 적은 채 이력에 남는다. 이 파일의 §2.1 이 그 정정이며, 두 문서를 함께 읽어야 한다.

## 6. 리드가 받아 갈 것

- 판정 **PASS 83.35**, 막는 결함 0. 카드는 통합 창을 요청할 수 있다.
- **class W-vs-R 는 수리가 옳다.** 1회차 지시문을 따라 `R` 로 되돌리지 말 것 — 되돌리면 범례와 어긋난다(테스트는 어느 쪽이든 그린이라 기계가 잡아 주지 않는다).
- 남는 비차단 2건(F3·F4)과 선택 4건(F5·F6·F7·F8·F9 중 선택 5건)은 이 카드에서 닫지 않아도 된다. F8 은 한 줄 수정이라 통합 전에 같이 넣을 수도 있다.
- `internal/cli` 판정은 develop push 가 일으키는 CI 가 첫 근거다.
