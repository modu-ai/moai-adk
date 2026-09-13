# t530 sync-audit — SPEC-DOCS-TABCOUNT-DRIFT-001

- 판정: **FAIL** (blocking 1건: F1)
- 가중 점수: **0.80 / 1.00** (조화평균 0.75) — 점수만으로는 통과선이며, FAIL 을 만드는 것은 must-pass 방화벽이 아니라 F1 이다
- 평가 프로필: `default` (`.moai/config/evaluator-profiles/default.md`; `harness.yaml default_profile: "default"`, spec.md 에 `evaluator_profile` 없음)
- must-pass 방화벽 (Functionality + Security): **둘 다 PASS** — 방화벽이 발동한 것이 아니다
- 감사 트리: `.claude/worktrees/t530`, 브랜치 `WT-web-tab-docs`, HEAD `c5b72f275`, merge-base `1d150a27d`
- 감사 일자: 2026-09-13 / 감사자: sync-auditor (독립 재실행, progress.md 기록은 주장으로만 취급)

---

## 1. Claim

이 카드는 다음을 주장한다.

1. `acceptance.md` 의 AC-TCD-001~012 전부 PASS, RG-TCD-001~003 전부 유지.
2. `internal/web/docs_tab_contract_test.go` 가드가 공허하게 통과하는 것이 아니라 실제로 판별한다(변이 6종).
3. CHANGELOG 항목의 사실 주장이 실제 diff 와 일치한다.
4. 3-phase close 가 단일 sync 커밋 + SHA backfill 로 올바르게 수행됐다.

감사는 이 넷을 각각 **직접 실행해** 판정했다. `progress.md §E.2`·`§E.4` 의 기록은 CLAIM 으로만 취급했고, 아래 Evidence 는 전부 이 감사 세션이 이 트리에서 다시 잰 값이다.

---

## 2. Evidence

### 2.1 AC 매트릭스 — 12/12 독립 재실행

| AC | 판정 | 명령 | 관측 출력 |
|---|---|---|---|
| AC-TCD-001 | PASS | `head -4 count-literals.txt > /tmp/a; wc -l < /tmp/a; grep -rnF -f /tmp/a docs-site/content/{ko,en,ja,zh}/cli-reference/web.md \| wc -l` | `4` / `0` (패턴 파일 4줄 확인 후 0적중 — 공허한 0 아님) |
| AC-TCD-002 | PASS | `grep -rnF -f .moai/reports/t530/count-literals.txt <12파일> \| wc -l` + `wc -l count-literals.txt` | `0` / `16` |
| AC-TCD-003 | **PASS (단서)** | `grep -rn '3rd Party LLM\|서드파티 LLM\|サードパーティ LLM\|第三方 LLM' <8파일> \| wc -l` | `0`. 대체 라벨 합계 **12행** (README 4본 `GLM Settings` 각 1 + en console 2 + ko `GLM 설정` 2 + ja `GLM設定` 2 + zh `GLM设置` 2). 단서는 F3 참조 |
| AC-TCD-004 | PASS | `go test ./internal/web/ -run 'TestDocsTabContract' -v` | `--- PASS: TestDocsTabContract/{literals,allowlist,names}` 세 줄 모두 존재, `--- FAIL` 없음, `no tests to run` 없음 |
| AC-TCD-005 | PASS | 위와 동일 | `swept 12 files against 16 enumerated literals` |
| AC-TCD-006 | PASS (3/6 직접 재현) | 아래 §2.2 | V1·V4·V6 각각 FAIL 관측 + 되돌린 뒤 `ok` |
| AC-TCD-007 | PASS | 위와 동일 | `allowed rules 1` · `allowed lines 16` |
| AC-TCD-008 | PASS | `grep -c 'measured nine forms' README.md` 외 3건 + 가드 재실행 | `1` / `1` / `1` / `1`, 가드 `ok` |
| AC-TCD-009 | PASS | 위와 동일 | 8파일 각 `extracted 14 tab names` |
| AC-TCD-010 | PASS | `CARD_BASE=$(git merge-base develop HEAD)` → `git diff --name-only "$CARD_BASE" -- <12파일>` | `12` / `^README` `4` / 상대경로 `4 advanced/moai-web-console.md` + `4 cli-reference/web.md` |
| AC-TCD-011 | PASS | `grep '^\| [ABCD][0-9]' tab-count-sites.md \| wc -l` 외 2건 | 행 `32` / 경로 `12` / MISSING 출력 없음 |
| AC-TCD-012 (셸 축) | PASS | `tab-count-sites.md § 재현 명령 5` 축어 실행 | 출력 없음 (`rc=1`). **공허 아님 확인**: 마지막 두 필터를 뗀 원시 스윕은 이 트리에서 `1`행을 내며, 그 1행이 codex 허용 규칙에 걸리는 `en/advanced/moai-web-console.md:149` 다 |
| AC-TCD-012 (가드 축) | PASS | §2.3 (M1 트리 복원 후 재측정) | `word-axis hit` **정확히 6줄**, 집합이 A2·A4·B4·C1·C2·C4 와 일치 |

### 2.2 RG 매트릭스 — 3/3

| RG | 판정 | 명령 | 관측 |
|---|---|---|---|
| RG-TCD-001 | PASS | `git diff --quiet 1d150a27d4c5cdeedb37df19b7a4a025e5dd2c09 -- assets/images/` | 종료코드 `0` (무변경). `grep -c '## 5. 스크린샷 범위 판단' spec.md` = `1` |
| RG-TCD-002 | PASS | `cd docs-site && hugo --gc --minify` | `rc=0`, `grep -ciE 'warn\|error'` = `0`, hugo `v0.160.1+extended`, 4로케일 빌드 완료 |
| RG-TCD-003 | PASS | `grep -c '^\| [ABCD][0-9]' tab-count-sites.md` | `32` (기준 32 이상) |

### 2.3 판별력 검증 — 변이 3종 직접 주입·관측·복원

progress.md 가 6종을 주장한다. 감사는 그중 **3종을 직접 재현**했고, 세 결과가 기록과 축어로 일치했다.

**V1 (`README.ko.md` 이름 축, `Audit` → `Audits`)**
```
docs_tab_contract_test.go:377: README.ko.md: tab 7 name = "Audits", console renders "Audit"
--- FAIL: TestDocsTabContract/names
```

**V6 (`zh/cli-reference/web.md`, `设置九个标签页` 되살림 — zh 낱말 축)**
```
docs_tab_contract_test.go:142: hand-written tab count survives: docs-site/content/zh/cli-reference/web.md contains "设置九个标签页"
word-axis hit docs-site/content/zh/cli-reference/web.md: 九个标签页
--- FAIL: TestDocsTabContract/literals
--- FAIL: TestDocsTabContract/allowlist
--- PASS: TestDocsTabContract/names
```

**V4 (`en/advanced/moai-web-console.md`, `unfolds fourteen settings tabs below …` — 리터럴 집합에 없는 재작성문)** — 이 변이가 가장 무겁다. 리터럴 층은 **통과**하고 낱말 축만 잡는다:
```
word-axis hit docs-site/content/en/advanced/moai-web-console.md: fourteen settings tabs
docs_tab_contract_test.go:300: word axis: …:128 writes a settings tab count by hand: "fourteen settings tabs"
--- PASS: TestDocsTabContract/literals
--- FAIL: TestDocsTabContract/allowlist
```
셋 모두 되돌린 뒤 `git status --porcelain` 무출력, `go test ./internal/web/` = `ok … 14.168s`.

### 2.4 AC-TCD-012 가드 축 RED 재현 (M1 트리 복원)

기록된 RED 는 M1 직후·M2 이전 트리에서 잰 값이다. 감사는 그 트리를 복원해 **직접 재측정**했다 — `git show d85ac3e9e:<path> > <path>` 로 대상 12파일만 되돌린 뒤 가드 실행, 이후 `git show HEAD:<path> > <path>` 로 복원.

```
word-axis hit README.md: fourteen tabs
word-axis hit README.ko.md: 열네 개 탭
word-axis hit README.zh.md: 十四个标签页
word-axis hit docs-site/content/en/cli-reference/web.md: nine settings tabs
word-axis hit docs-site/content/zh/cli-reference/web.md: 九个标签页
word-axis hit docs-site/content/en/advanced/moai-web-console.md: fourteen tabs
```
행 수 `6`. 같은 트리에서 가드 전체는 `--- FAIL` (세 서브테스트 전부). 다섯 토큰(`nine`·`九`·`fourteen`·`열네`·`十四`) 모두 1줄 이상. `progress.md:202` 의 기록과 **축어 일치**. 복원 후 `git status --porcelain` 무출력.

### 2.5 Craft / Consistency / Security 기계 검증

```
$ go test ./internal/web/ -cover
ok  github.com/modu-ai/moai-adk/internal/web  14.278s  coverage: 67.5% of statements

$ gofmt -l internal/web/docs_tab_contract_test.go   → 무출력 (exit 0)
$ go vet ./internal/web/                            → 무출력 (exit 0)
$ golangci-lint run --timeout=3m ./internal/web/... → 0 issues.

$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/web/docs_tab_contract_test.go        → 무적중
$ grep -nE 'os\.(Remove|WriteFile|Create)|exec\.Command|os/exec' <같은 파일>            → 무적중
$ git diff --name-only 1d150a27d… -- go.mod go.sum                                       → 무출력
```
비밀정보 스캔 적중 1건은 오탐이다 — `docs_tab_contract_test.go:225` 의 `token: "codex"` 는 허용 규칙 구조체 필드이지 자격증명이 아니다.

### 2.6 3-phase close

```
$ git show --stat 4d214b66b
docs(SPEC-DOCS-TABCOUNT-DRIFT-001): sync-phase artifacts (t530)
 progress.md | 149 ++++-   spec.md | 2 +-   CHANGELOG.md | 2 +
$ git show 4d214b66b -- .../spec.md | grep -E '^[+-](status|updated)'
-status: in-progress
+status: completed
```
단일 sync 커밋이 `in-progress → completed` 전이 + CHANGELOG + §E.4 를 함께 운반했고, `sync_commit_sha: 4d214b66b` 는 후속 커밋 `c5b72f275` 가 backfill 했다(스키마 §D3 예외 경로). placeholder 잔존 없음. **PASS.**

---

## 3. Baseline-attribution

- 측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t530` (`git rev-parse --show-toplevel` 로 확인)
- HEAD: `c5b72f275` (감사 시작·종료 시점 동일, `git status --porcelain` 무출력)
- 범위 왼쪽 끝: `git merge-base develop HEAD` = `1d150a27d4c5cdeedb37df19b7a4a025e5dd2c09` — 읽는 시점에 재산출, 리터럴로 고정하지 않음(`gitflow-lane-protocol.md` §8)
- RED 재현용 참조 커밋: `d85ac3e9e` (M1)
- 모든 명령은 이 실행에서 이 트리를 대상으로 직접 수행했으며, `progress.md` 에 기록된 수치를 근거로 재사용하지 않았다
- 도구: go (프로젝트 툴체인), hugo v0.160.1+extended, golangci-lint (설치 확인됨)
- 판정 근거로 인용하는 산출물(`tab-count-sites.md`, `count-literals.txt`, `progress.md`)은 전부 추적 파일이며 이 브랜치에 커밋돼 있다

---

## 4. Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 0.75 | PASS | AC 12/12 + RG 3/3 재실행 통과(§2.1·§2.2). 1.00 이 아닌 이유: AC-TCD-003 의 And 대조군이 문자 그대로는 성립하지 않는다(F3) |
| Security (25%) | 1.00 | PASS | §2.5 — 의존성 무변경, 비밀정보·쓰기·실행 표면 없음, 서브에이전트 경계 무적중 |
| Craft (20%) | 0.50 | **FAIL (임계 미달, 카드 귀속 아님)** | `coverage: 67.5% of statements` < 프로필 하드 임계 85%. 상세는 F4 |
| Consistency (15%) | 1.00 | PASS | `gofmt -l` 무출력 · `go vet` exit 0 · `golangci-lint` `0 issues.` |

가중 합: `0.75×0.40 + 1.00×0.25 + 0.50×0.20 + 1.00×0.15 = 0.80`. 조화평균 `0.75`.

must-pass 방화벽(Functionality, Security)은 **양쪽 다 통과**했다. Craft FAIL 은 프로필상 전체 판정을 뒤집지 않는다(Security 만이 하드 오버라이드). 따라서 **이 FAIL 은 점수나 방화벽이 아니라 blocking finding F1 이 만든 것**이다.

---

## 5. Findings

- **F1** [Medium] [**blocking**] `CHANGELOG.md:12` — 항목이 *"Two of the 20 sites were already wrong (`9` instead of `14`, both in `cli-reference/web.md`, one numeral and one spelled out in English and Chinese)"* 라고 적는다. 이 카드 자신의 열거 산출물 `tab-count-sites.md §3` (A1~A4) 과 `acceptance.md` AC-TCD-001 은 **4자리**(ko·en·ja·zh, 전부 `cli-reference/web.md:53`)를 기록하고, 감사가 잰 A군 리터럴 4줄도 그 4자리다. "both"·"Two" 는 사실과 다르다. 뒤따르는 표기 설명도 어긋난다 — A1(ko)·A3(ja)이 숫자, A2(en)·A4(zh)가 낱말이므로 "one numeral and one spelled out" 은 성립하지 않는다. 확신도: 높음(카드 자신의 산출물과 감사 실측이 함께 반증). **손으로 적힌 수가 어긋난다는 것이 이 카드의 논지인데 카드가 배달한 사용자 대면 문서가 바로 그 오류를 담고 있다.** — Required fix: 해당 문장을 `Four of the 20 sites were already wrong (9 instead of 14 — all four locale copies of cli-reference/web.md:53; ko/ja write the digit, en/zh spell the numeral out)` 취지로 정정한다.

- **F2** [Low] [optional] `CHANGELOG.md:12` — *"a one-rule/16-line allowlist for a legitimate rhetorical count inside `ja/advanced/moai-web-console.md`"*. 허용 규칙의 파일 범위는 접미사 `advanced/moai-web-console.md`(`docs_tab_contract_test.go:224`)라서 **4로케일 사본 전부**를 덮고, 실제 면제 줄 수도 로케일당 4줄 × 4 = 16 이다(`allowed lines 16`). ja 하나만 지목하면 면제 표면을 축소해 읽힌다. 게다가 현재 트리에서 실제로 살아남는 스윕 적중은 `en/advanced/moai-web-console.md:149` 다(§2.1). 확신도: 높음. — Required fix: "in the four `advanced/moai-web-console.md` locale copies" 로 범위를 정정.

- **F3** [Low] [optional] `acceptance.md` AC-TCD-003 And 절 — 각 README 에서 **그 로케일의** 라벨(`GLM 설정`/`GLM設定`/`GLM设置`)이 1행 이상일 것을 요구하지만, 실측은 `README.ko.md`·`README.ja.md`·`README.zh.md` 모두 그 라벨 **0행**이고 영어 `GLM Settings` 1행이다. 다만 README 의 탭 이름 나열은 14개 전부를 영어로 적는 것이 정본이며 가드의 `names` 층이 그 영어 표기를 강제하므로, **구현이 아니라 기준 문구가 틀렸다.** 실질(정본 라벨 합계 12행, 구 라벨 0행)은 충족된다. 확신도: 높음. — Required fix: AC 문구를 "README 4본은 영어 라벨 `GLM Settings` 1행 이상" 으로 정정(문서 쪽을 바꾸지 말 것 — 바꾸면 `names` 가드가 붉어진다).

- **F4** [Low] [optional] `internal/web` 패키지 커버리지 `67.5%` 가 프로필 하드 임계 85% 아래다. **이 카드에 귀속되지 않는다** — 카드가 추가한 Go 파일은 `docs_tab_contract_test.go` 하나뿐이고 프로덕션 문장을 늘리지 않으므로 커버리지를 낮출 수 없다(오히려 올린다). 카드 이전 baseline 은 재지 않았다(Gaps 참조). 확신도: 중간(귀속 논증은 확실, 정확한 이전 값은 미측정). — Required fix: 이 카드에서는 없음. 패키지 커버리지 개선은 별도 카드.

- **F5** [Info] [optional] `acceptance.md §D.2` DoD 5번("후속 카드 2건이 리드에게 전달됐다")은 이 트리에서 검증 불가능하다 — 산출물이 아니라 메시지다. 확신도: 높음(검증 불가라는 사실 자체는 확실). — Required fix: 리드가 큐(`moai todo`)에서 두 카드의 실재를 확인.

- **F6** [Info] [optional] `tab-count-sites.md §7-A` 이 허용 규칙이 덮는 자리로 `ja/…:149` 를 지목하지만, 현재 트리에서 실제로 규칙에 걸리는 자리는 `en/…:149`(낱말 축)다. 산출물이 스스로 행 번호는 base 기준임을 밝히고 있으므로 결함이라기보다 시점 차이다. 확신도: 중간. — Required fix: 없음(정보).

---

## 6. Gaps (관측하지 않은 것)

- **변이 6종 중 3종(V2·V3·V5)은 직접 주입하지 않았다.** V1·V4·V6 만 재현했다. 다만 세 축(이름/숫자/낱말)과 세 로케일(ko·en·zh), 두 파일 부류(README·docs-site)를 모두 덮으므로 판별력 결론에는 충분하다고 판단했다.
- **카드 이전 `internal/web` 커버리지 baseline 미측정.** F4 의 "카드 귀속 아님" 은 구조적 논증(테스트 파일만 추가)에 근거하며, `1d150a27d` 에서 실제로 재지는 않았다.
- **CHANGELOG 의 나머지 서술은 전수 대조하지 않았다.** F1·F2 는 카드 자신의 산출물과 정면으로 어긋나 잡힌 것이고, "20 sites"·"8 files"·"32 rows / 12 paths"·"12 acceptance criteria" 는 확인했으나 문장 단위 전수 검증은 아니다.
- **원격 CI 미판정.** 이 감사는 로컬 트리 판정이다. darwin/windows 매트릭스와 전체 스위트는 `origin/develop` push 가 만드는 실행에 남는다.
- **`go test ./...` 미실행** — 지시대로 `./internal/web/` 로 범위를 한정했다. 다른 패키지에 대한 회귀 여부는 이 감사가 말하지 않는다.
- **docs-site Vercel 바인딩 영향 미검증** — 카드도 잔여 위험으로 기록하고 있고, 감사도 재지 않았다.
- **`moai spec audit` 재실행 안 함** — progress.md §E.4 가 INFO 1건(EraAutoDetected)을 기록하나, 감사는 이를 재관측하지 않았다.

---

## 7. Residual-risk

- **낱말 클래스는 손으로 유지되는 목록이다.** `wordNumeralRe`(`docs_tab_contract_test.go:192-199`)에 없는 새 로케일·새 표기는 조용히 빠져나간다. 가드 자신의 주석(`:190`)과 `spec.md §7` 이 이 비용을 이미 명시하고 있으나, 구조상 남는 위험이다.
- **허용 규칙 1개가 16줄을 면제한다.** 규칙 수는 가드가 단정하지만, `advanced/moai-web-console.md` 4본에서 `codex` 토큰이 늘면 면제 표면이 커진다. `allowedLinesBaseline = 16` 이 그 이동을 잡도록 설계돼 있으나, 정당한 codex 문장이 늘어날 때마다 baseline 을 손으로 올려야 한다.
- **`allowedLinesBaseline` 은 base 시점 상수다.** 다른 카드가 codex 문장을 건드리면 이 가드가 무관한 이유로 붉어진다. 실패 메시지가 이유를 밝히긴 하지만, 이 파일을 만지는 사람에게는 예상 밖 실패로 보인다.
- **D9~D12 산문 4자리는 여전히 가드 밖이다** — 카드가 스스로 기록한 잔여 구멍이며 감사도 이를 확인했다(수 자체가 제거돼 재발 확률은 낮다).
- **패키지 커버리지 67.5%** — 이 카드의 책임은 아니지만, `default` 프로필의 Craft 임계를 계속 밑돌므로 이 패키지를 만지는 다음 카드마다 같은 FAIL 이 재보고된다.

---

## 8. 재감사 범위 (FAIL 후속)

blocking finding 은 F1 하나다. 재감사는 전면 재실행이 아니라 다음 델타로 한정한다.

1. `CHANGELOG.md` 의 해당 문장 정정 확인(F1). F2 도 같은 줄이라 함께 고치는 것이 자연스럽다.
2. 정정 커밋이 `CHANGELOG.md` 외 파일을 건드리지 않았음 확인 — `git diff --name-only <정정전>..<정정후>`.
3. 가드 재실행 1회(`go test ./internal/web/ -run 'TestDocsTabContract'`) — 회귀 없음 확인용.

AC 12건·RG 3건·변이 재현·3-phase close 는 이 판정서가 이미 통과로 확정했으므로 다시 재지 않는다.

---

## Operational Notes (unverified)

- [inferred] F3 을 "문서를 고쳐서" 닫으려 하면 `names` 가드가 붉어진다 — README 의 탭 이름 나열은 `consoleTabs()` 의 영어 `Baseline` 과 대조되기 때문이다(`docs_tab_contract_test.go:386-412`). 코드를 읽고 추론한 것이며, 로케일 라벨로 바꿔 보는 변이는 주입하지 않았다. 확인하려면 `README.ko.md:414` 의 `GLM Settings` 를 `GLM 설정` 으로 바꾸고 `go test ./internal/web/ -run 'TestDocsTabContract/names' -v` 를 돌린 뒤 되돌린다.
- [measured] hugo 로컬 빌드는 무경고다(`rc=0`, `grep -ciE 'warn|error'` = `0`, v0.160.1+extended). Vercel 이 쓰는 hugo 버전과 대조하지는 않았다.

---

🗿 MoAI
