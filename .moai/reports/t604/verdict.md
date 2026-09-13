# t604 — [hooks 감사 2026-09-11 · H09 · P2] Kotlin 프로젝트가 Java로 감지되어 검사 없이 종료됨

- 카드: t604
- 워크트리: `.claude/worktrees/t604` · 브랜치 `WT-gate-kotlin-detect`
- baseline: `eabce7444` (origin/develop 흡수 완료)
- 대상: `.claude/hooks/moai/sync-phase-quality-gate.sh` (+ `internal/template/templates/` 미러)
- 규범 근거: `.claude/rules/moai/workflow/language-routing-contract.md` — "`build.gradle.kts` 는 Kotlin 플러그인이나 **Kotlin 소스가 있을 때** Kotlin" (리드 인용 승인)

---

## 1. 착수 전 재현 재확인 — 카드의 재현은 재현되지 않는다

리드의 [HARD] 지시에 따라 최신 develop 에서 먼저 재측정했다.

**Claim 1.** 카드가 적은 재현(루트 `build.gradle.kts` + `Main.kt`)은 `eabce7444` 에서 성립하지 않는다.

**Evidence 1.** 픽스처 A — 루트에 `build.gradle.kts`(`plugins { kotlin("jvm") version "2.0.0" }`) 와 `Main.kt`. git `docs: sync` 커밋 후:

```
echo '{}' | CLAUDE_PROJECT_DIR=/private/tmp/t604fx/A bash .../sync-phase-quality-gate.sh
→ EXIT=0
.moai/logs/sync-quality-gate.log:
  ... language=kotlin snapshot_status=miss head=63cf8301...
  ... language=kotlin languages=kotlin mode=advisory decision=allow kotlinc=0 (none)=0 deps_modified=1
```

카드의 세 증상(kotlinc 미호출 · 게이트 로그 미생성 · exit 0 으로 조용히 통과) 중 앞의 둘이 불성립. 기여 커밋은 `de957342a` "fix: route Kotlin and monorepo languages in card t640" 이다.

> **귀속 정정.** 배차문은 이 수리를 t601 로 지목했으나, 해당 파일 이력상 Kotlin 선판별 분기를 넣은 것은 **t640**(`de957342a`)이다. t601(`2ae7c1c3f`)이 같은 파일에 넣은 것은 작업-트리 내용 식별자다. 두 커밋 모두 이 카드의 baseline 에 들어와 있다.

---

## 2. 그러나 같은 결함이 관용 배치에서 살아 있다

**Claim 2.** 동일한 결함이 **관용 Gradle 배치(`src/main/kotlin/`) + version-catalog 빌드스크립트** 조합에서 그대로 남아 있으며, 카드가 적은 세 증상을 글자 그대로 재현한다.

**Evidence 2.** 픽스처 C — `src/main/kotlin/Main.kt` + `build.gradle.kts` 본문이 `plugins { alias(libs.plugins.jvm) }`:

```
→ EXIT=0
cat .moai/logs/sync-quality-gate.log → (NO LOG)   # 파일 자체가 생성되지 않음
```

kotlinc 미호출 · 게이트 로그 미생성 · exit 0 — 카드의 3증상 전부 성립.

**Evidence 2-기전.** 107행 조건의 두 판별식을 각각 따로 측정:

| 측정 | C (결함) | B (대조군) |
|---|---|---|
| `grep -Eic 'kotlin\(\|org\.jetbrains\.kotlin\|kotlin-dsl' build.gradle.kts` | `0` (exit 1, 미적중) | `1` (exit 0, 적중) |
| `find <root> -maxdepth 3 -type f -name '*.kt' -print -quit` | 빈 출력 | **빈 출력** |
| `test -f build.gradle.kts` | 참 → `elif` 로 낙하 | — |
| 결과 | **java** | kotlin |

두 픽스처를 가르는 것은 **grep 한 줄뿐**이다. 이후 `code_delta_pattern java` = `\.java$` 가 변경된 `Main.kt` 를 잡지 못해 `CODE_DELTA=0` 이 되고, 218행에서 조용히 `exit 0` 한다 — 로그조차 남지 않는 이유가 이것이다.

**파생 발견 (근본).** `has_suffix '*.kt'` 의 `-maxdepth 3` 은 **관용 Gradle 배치에서 원리상 절대 적중하지 않는다**. B 에서도 빈 출력이었고, B 가 kotlin 으로 잡힌 것은 순전히 grep 덕이다. 즉 Kotlin 판별이 **문자열 grep 단일 다리** 위에 있었고, version catalog(현대 Gradle 기본)에서 그 다리가 끊긴다.

**대조군.** 픽스처 D(진짜 Java: `id("java")` + `Main.java`) → java. 오탐 아님.

---

## 3. 수리

세 곳. 리드 승인 순서대로 (1) 접미사 탐색 확대를 **근본**, (2) 별칭 grep 을 **보조**.

| # | 위치 | 변경 |
|---|---|---|
| R1 | `detect_languages` 내부 | `has_kotlin_source()` 신설 — 깊이 상한 대신 무거운 디렉터리(`.git` `build` `target` `node_modules`)를 prune 하고 `*.kt` 만 탐색. `*.kts` 는 제외(빌드 DSL 이므로 Java Gradle 프로젝트가 Kotlin 으로 오분류되지 않도록) |
| R2 | 107행 조건 | `has_suffix '*.kt'` → `has_kotlin_source`; grep 패턴에 `libs\.plugins\.kotlin` 추가 |
| R3 | 144행 | `code_delta_pattern kotlin`: `'\.kt\|\.kts$'` → `'\.(kt\|kts)$'` |

R3 은 **카드 문구 밖, 리드 승인**(같은 언어 감지 축 · 한 줄 · 재현의 CODE_DELTA 경로에 직접 얽힘). 앞쪽 `\.kt` 에 끝 앵커가 없어 ".kt" 를 **포함하기만 한** 임의 경로가 Kotlin 코드 변경으로 계수되던 문제. 테스트 1개 + 뮤턴트 1개로 고정.

템플릿 미러 반영 후 `diff -q` → IDENTICAL.

---

## 4. 검증

**테스트.** `.claude/hooks/tests/test-language-routing-contract.sh` (t640 이 만든 기존 하네스)에 4개 주장 추가:

1. version-catalog + 깊은 Kotlin 소스 → `kotlin` — R1 을 지킴
2. Kotlin 별칭(`alias(libs.plugins.kotlin.jvm)`) 단독, `.kt` 소스 없음 → `kotlin` — R2 를 지킴 (1번 픽스처는 별칭에 kotlin 토큰이 아예 없어 R2 를 지킬 수 없다)
3. Java Gradle(`build.gradle.kts` + `.java`, `.kt` 없음) → `java` — R1 이 과도하게 넓어지는 것을 막는 대조군
4. 델타 패턴 **양방향**: `Main.kt`·`build.gradle.kts` 는 적중해야 하고 `docs/README.kt.md` 는 적중하면 안 됨 — 아무것도 매치하지 않는 패턴이 "오탐 없음"으로 통과하지 못하도록

```
bash .claude/hooks/tests/test-language-routing-contract.sh
→ PASS: language routing distinguishes Kotlin and preserves monorepo candidates
→ EXIT=0
```

**뮤턴트 3/3 — 각자의 주장에 잡힘.**

| 뮤턴트 | 되돌린 것 | 관측된 실패 |
|---|---|---|
| M1 | `has_kotlin_source` → `has_suffix '*.kt'` | `FAIL: version-catalog Kotlin project detected as java, want kotlin` (EXIT 1) |
| M2 | grep 에서 `\|libs\.plugins\.kotlin` 제거 | `FAIL: Kotlin catalog alias detected as java, want kotlin` (EXIT 1) |
| M3 | 델타 패턴 앵커 되돌림 | `FAIL: kotlin delta pattern \.kt\|\.kts$ matches a non-Kotlin path` (EXIT 1) |

M1·M2 가 **서로 다른 주장**에 잡힌 것이 두 다리가 독립임을 세운다. 되돌린 뒤 다시 PASS/EXIT 0 확인.

**종단 재현(수리 후).**

| 픽스처 | 구성 | 수리 전 | 수리 후 |
|---|---|---|---|
| A | 루트 `build.gradle.kts` + `Main.kt` | kotlin | kotlin |
| B | `src/main/kotlin/` + `kotlin("jvm")` | kotlin | kotlin |
| **C** | `src/main/kotlin/` + `alias(libs.plugins.jvm)` | **java, 로그 없음** | **kotlin, 로그 2줄 생성** |
| D | Java Gradle | java | java |

**저장소 검증.**

```
bash -n .claude/hooks/moai/sync-phase-quality-gate.sh   → SYNTAX-OK
make build                                              → EXIT 0, catalog.yaml 드리프트 0 (git status 미출현)
go test ./internal/template/...                         → ok (template 34.3s, agentemit, commandemit)
```

---

## 5. Gaps (명시적 미검증)

- ~~`go test ./internal/hook/...` 의 FAIL 1건이 미귀속이다.~~ → **§5.1 에서 해소.** 최초 측정 때 `tail -10` 이 실패 패키지 이름을 잘랐고 `echo $?` 가 파이프 때문에 tail 의 종료코드(0)를 찍었다 — 둘 다 계측 결함이었고, 그 상태의 출력으로는 어느 쪽으로도 판정할 수 없었다.
- shellcheck 미설치(`command not found`)라 정적 검사를 돌리지 못했다.
- 멀티모듈 배치, `settings.gradle.kts` 기반 구성, 나머지 15개 언어의 감지 경로는 측정하지 않았다.
- 로그의 `kotlinc=0` 이 "검사 통과"인지 "도구 부재 스킵"인지 분해하지 않았다. 이번 판정 축(언어 감지)에는 영향이 없다.

## 5.1 `internal/hook` 레드 귀속 (홀드 해제 후 재측정)

부하 해제 뒤 `uptime` 확인(1분 14.69 < 30) 후 루트 패키지만 직렬로 재측정. 이번에는 출력을 파일로 받고 종료코드를 **파이프 밖에서** 읽었다.

```
go test ./internal/hook/ > .moai/reports/t604/hook-root-test.log 2>&1; echo "GO-TEST-EXIT=$?"
→ GO-TEST-EXIT=1
→ --- FAIL: TestSessionStart_DeferredScanDoesNotBlockReturn (0.58s)
→ FAIL   github.com/modu-ai/moai-adk/internal/hook   280.974s
```

**Claim.** 이 레드는 이번 카드의 변경과 인과가 없다.

**Evidence (본 레인 자체 측정).**

1. 실패는 **한 건뿐**이다 — 로그에 `--- FAIL` 줄이 하나다.
2. 실패 테스트의 소재는 `internal/hook/session_start_parallel_test.go` (대상 `session_start.go`) 이다 — `grep -rln` 으로 확정.
3. 이번 변경 대상 심볼(`sync-phase-quality-gate` / `detect_languages` / `code_delta_pattern`)을 참조하는 `internal/hook/*.go` 는 5개이고 — `stopchain_ac004_006b_test.go`, `stopchain_trim_test.go`, `sync_gate_failstate_test.go`, `user_decision_capture.go`, `wrapper_copies_contract_test.go` — **전부 통과했다**. 실패 파일은 이 목록에 없다(참조 0건).
4. 특히 `wrapper_copies_contract_test.go`(로컬↔템플릿 사본 계약)와 `sync_gate_failstate_test.go` 가 초록인 것은, 템플릿 미러 반영이 계약을 깨지 않았다는 **적극적 근거**이기도 하다.

**인용(본 레인 미측정, 리드 경유).** lane-4 가 `internal/hook` 전체에서 실패가 이 한 건뿐임을, lane-3·lane-7 이 base 에서도 이 테스트가 실패함을 각각 독립으로 실측했다고 보고되었다 — 즉 develop 선재 레드. 본 레인은 **이름 확정과 인과 독립성까지만 자체 측정**했고, base 대조는 이 세 관측의 인용으로 갈음한다(리드 승인). 인용은 인용으로 표시하며, 본 레인의 실측으로 계산하지 않는다.

## 5.2 병합 트리 재측정 (통합 창, lane-10)

창 확보(`moai integration acquire --name lane-10` → `acquired ... on develop`) 후 로컬 `develop`(`e9aedd128`, t603 착지본)을 흡수. **충돌 없음** — t603 은 `:574` 주변, 본 카드는 `99-127` + `144` 행이라 겹치지 않았고 `ort` 전략이 자동 병합했다. 병합 커밋 `63fa51cf3`.

각 측정 직전 `uptime` 으로 부하를 확인했다(리드 게이트 30).

| 측정 | 결과 |
|---|---|
| 로컬↔템플릿 미러 `diff -q` | IDENTICAL |
| `bash -n` | SYNTAX-OK |
| `bash .claude/hooks/tests/test-language-routing-contract.sh` | PASS, EXIT 0 |
| 픽스처 A / B / C / D 종단 | kotlin / kotlin / **kotlin** / java |
| `go test ./internal/template/...` | **EXIT 0** — ok 3패키지 (t603 이 추가한 `hook_cpp_gate_behavior_test.go` 포함) |
| `go test ./internal/hook/` | **EXIT 0 — ok** (210.999s) |
| `make build` | EXIT 0, catalog.yaml 드리프트 0 |

### 새 사실 — §5.1 의 레드가 이번엔 재현되지 않았다

§5.1 에서 부하 52.63 아래 측정했을 때 `TestSessionStart_DeferredScanDoesNotBlockReturn` 이 실패했으나, 병합 트리에서 부하 18.13 아래 재측정하니 **패키지 전체가 `ok`, EXIT 0** 이다. 같은 테스트가 두 조건에서 다른 결과를 냈다.

두 가설이 남고, 본 레인은 **어느 쪽도 확정하지 않았다**:

- (a) **부하 의존 flake.** 테스트 이름이 `DoesNotBlockReturn` 으로 반환 지연을 보는 성격이고, 실패 시점 부하는 52.63, 통과 시점은 18.13 이었다. 타이밍 단언이 부하에 밀렸을 수 있다.
- (b) **t603 흡수가 고쳤다.** 병합으로 들어온 변경이 원인을 제거했을 수 있다.

가르려면 동일 부하 조건에서 base(`eabce7444`)와 병합본(`63fa51cf3`)을 각각 재야 한다 — 본 카드 범위 밖이라 하지 않았다.

**팀 차원의 함의.** lane-3·lane-7·lane-4 가 이 레드를 "develop 선재 레드"로 귀속했고 §5.1 은 그것을 인용했다. 그러나 부하가 낮을 때 통과한다면 **결정적 레드가 아니라 부하 유발 flake** 일 수 있고, 그렇다면 "base 에서도 실패" 라는 관측들 역시 높은 부하 아래에서 나왔을 가능성을 검토해야 한다. 본 레인의 이번 관측은 그 재검토를 요구하는 **반례 1건**이며, 판정이 아니다. 리드에게 별도 보고했다.

본 카드의 판정에는 영향이 없다 — 병합 트리에서 관련 패키지가 모두 초록이므로 어느 가설이 맞든 본 카드는 레드를 들여오지 않았다.

## 6. Residual-risk

- 픽스처가 4개뿐이다. 떠올리지 못한 배치가 더 있을 수 있다.
- `has_kotlin_source` 는 깊이 상한 대신 prune 을 쓴다. `.kt` 가 하나도 없는 대형 저장소에서는 전체 순회 비용이 발생한다(첫 적중에서 `-quit` 하므로 Kotlin 프로젝트에서는 저렴하다). Stop 훅 60초 예산 안이라고 보나 대형 저장소에서 측정하지는 않았다.
- 공유 `has_suffix` 의 `-maxdepth 3` 자체는 **손대지 않았다**. 같은 구조적 한계가 다른 언어(예: `src/main/java/com/x/Main.java`)에도 있으나 매니페스트 마커가 가려주고 있다. 16개 언어 전반의 감지 깊이는 별도 카드 소관으로 남긴다.
- t603(lane-9)이 같은 파일 `:574` 주변을 수정 중이다. 먼저 창을 받는 쪽이 병합하고 나중 쪽이 흡수한다.
