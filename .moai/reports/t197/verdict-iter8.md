# SPEC Review Report: SPEC-CODEX-LAUNCHER-001

Iteration: 8
Verdict: **FAIL**
Overall Score: 0.82 (must-pass 실패가 있으므로 점수와 무관하게 FAIL)

M1 Context Isolation 고지: 배차 프롬프트가 전한 "무엇이 바뀌었다" 는 서술은 **주장으로만** 받았고, 판정은 내가 직접 읽고 실행한 것에만 근거한다.

## 0. 내가 검증한 content pin

| 항목 | 관측값 |
|---|---|
| `git rev-parse --short HEAD` | `219cbe19c` — 지시된 핀과 **일치** |
| `git branch --show-current` | `WT-codex-launcher` |
| `git status --porcelain` (감사 착수 시점) | 빈 출력 — **clean** |

감사 도중 트리가 한 번 더러워졌다. 원인은 내가 실행한 `gate-selftest.sh` 이며(D1 참조), 감사 종료 시 `git checkout -- .moai/reports/t197/citation-manifest.txt` 로 되돌려 `git status --porcelain` 빈 출력을 재확인했다.

## 1. 백엔드별 판정과 내가 채택한 것

| 백엔드 | gate | verdict | 채택 |
|---|---|---|---|
| claude (본 세션, anchor) | required | **FAIL** | 본문 전체 |
| codex | required | **fail** | **부분 채택** — 7건 중 3건만 |
| glm | advisory | `inconclusive` | 미채택 (fail-open) |

`audit_multi` overall_verdict = `fail`, `disagreement_flag` = false, `fail_open_backends` = [glm]. glm 은 리뷰 대상 문자열을 해석하지 못해 판정 없음(백엔드 부재와 동치, 차단하지 않음).

**codex 지적의 채택 내역** — 본문이 판정을 이긴다는 관례에 따라 각 지적의 본문을 내가 따로 재현했다:

| codex 지적 | 내 판정 |
|---|---|
| 1·2·3 (테스트 선택, 경로 거부 읽기, symlink 경쟁) | **미채택 — 범위 밖**. 셋 다 `SPEC-CODEX-INIT-001` 소관이며 본 감사 대상 SPEC 이 아니다. LAUNCHER 의 `-run Codex` 는 AC-CL-015 가 skip 수 0 + 시험 함수 이름 목록 대조로 이미 막아 두었으므로 지적 1을 그대로 옮길 수 없다 |
| 4 (AC-CL-012 격리 홈 밖 쓰기) | **채택** → D4. 내가 AC 문면을 다시 읽어 재현했다 |
| 5 (파싱 실패가 명령 프로브로 하강하지 않아도 통과) | **채택하되 확대** → D2. codex 는 파싱 실패 한 칸만 짚었으나, 실제 구멍은 **거부된 파일 전 사유**(빈 tokens·미지 모드·파싱 실패)로 넓다 |
| 6 (self-test 가 추적 파일을 오염) | **채택하되 격상** → D1. codex 는 위생 문제로 적었으나, 관측된 실체는 **self-test 의 복원 검사가 거짓 통과** 하는 것이고 spec.md 본문이 그 복원을 사실로 단언한다 |
| 7 (AC-CL-009 케이스 수) | **채택** → D5 (minor) |

**codex 가 놓치고 내가 찾은 것**: D3 (크로스 플랫폼 모순, must-pass MP-6).

## 2. Must-Pass 결과

- **[PASS] MP-1 REQ 번호 일관성** — `REQ-CL-001` ~ `REQ-CL-014` 연속, 결번·중복 0, 제로패딩 일관. `spec.md` §D 에서 전수 확인. `REQ-CL-015/016` 문자열이 spec.md 에 나타나지만 전부 HISTORY 0.5.0/0.6.0/0.7.0 의 **이력 서술**이며 §D 정의부에는 없다(0.7.0 이 `SPEC-CODEX-INIT-001` 로 분리한 기록) — 결번 아님.
- **[PASS] MP-2 GEARS 준수** — 판정 층위: **요구사항 층(`REQ-CL-xxx`, spec.md §D)**. 14건 전부 `The system shall …` / `Where …, the system shall …` / `When …, the system shall …` 형태. 예: `spec.md:110` "The system shall provide a top-level `moai codex` command…", `spec.md:118` "Where the project's `.codex/` wiring is incomplete — …". acceptance.md 의 Given-When-Then 은 **검증 층**이므로 이 항목에서 감점하지 않는다.
- **[PASS] MP-3 YAML frontmatter** — 12필드 전수 존재: `id`·`title`·`version`("0.7.0" 인용됨)·`status`(draft)·`created`/`updated`(2026-08-24, ISO)·`author`·`priority`(P1)·`phase`·`module`·`lifecycle`(spec-anchored)·`tags`(CSV 문자열). 거부 별칭(`created_at`/`updated_at`/`labels`/`spec_id`) 0건.
- **[N/A] MP-4 언어 중립성** — 이 SPEC 은 moai 자체 Go 코드(`internal/cli`) 단일 언어 범위다. 16언어 도구 서술이 아니므로 N/A(자동 통과). 템플릿 표면 중립성은 REQ-CL-014 + AC-CL-013 이 별도로 기계화하고 있으며 그 판정은 오히려 강하다(비-ASCII 0자 단언).
- **[PASS] MP-5 D7 교차 SPEC** — 참조된 5개 전부 `.moai/specs/` 에 존재하고 `retired`/`superseded`/`archived` 0건: WIRING-001=completed, DUAL-AGENTS-001=completed, HOOK-ADAPTER-001=completed, SKILLS-CANONICAL-001=completed, INIT-001=draft. BLOCKING 없음.
- **[FAIL] MP-6 D8 크로스 플랫폼** — D3 참조. `syscall` 이 SPEC 본문에 등장(acceptance.md:287 `syscall.Exec`)하는데 `//go:build` 제약도 명시적 크로스 플랫폼 면제도 없고, **AC-CL-014 가 OS 빌드 태그를 0건으로 금지**한다. 교훈 #21(Windows `syscall.Flock` 빌드 태그 누락)과 같은 형태다.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-LAUNCHER-001/` 히트 0 (rc 1). `research.md` 는 Tier M 이라 부재.

## 3. 반복 7 차단 2건이 실제로 닫혔는가 — 내가 관측한 rc

지시받은 대로 보고서가 주장하는 rc 가 아니라 **내가 실행해 관측한 rc** 를 적는다.

**게이트 (`citation-sweep.sh`)** — 관측 `rc=0`, 말미 `CITATION GATE PASS — 230 checks, 0 failures`. 다섯 검사(C1 핀, C2 범위, C3 manifest 결속, C4 근거 포함, C5 고아 행)가 전부 돌았고, C4 가 실제로 정규식으로 줄 내용을 본다는 것을 출력에서 확인했다(예: `L24-26 supported by /^1ed61e4ac$/`).

**self-test (`gate-selftest.sh`)** — 관측 `rc=0`, `SELFTEST PASS`. 다섯 주입이 전부 rc 1 로 잡혔다:

```
caught: [C1 stale pin] rc=1
caught: [C2 out-of-bounds range] rc=1
caught: [C3 undeclared citation] rc=1
caught: [C4 unsupported citation] rc=1
caught: [C5 stale manifest row] rc=1
restore check:
        ok: both files byte-identical to the backup
SELFTEST PASS — every injected fault was rejected
```

**판정**:

- **F1(줄 범위가 주장을 담지 않는 인용) — 닫혔다.** C4 가 그 정확한 형태를 잡는 검사이고, self-test 가 C4 주입을 실제로 rc 1 로 거부하는 것을 관측했다. 230 검사 전수 통과.
- **F2(판정하지 않고 보고만 하는 스크립트) — 게이트 층에서는 닫혔고, self-test 층에서 재발했다.** `citation-sweep.sh` 는 rc 로 판정한다. 그러나 `gate-selftest.sh` 의 **복원 검사 자체가 판정하지 않고 통과를 보고**한다 — 아래 D1. F2 는 한 겹 위로 옮겨갔을 뿐 소멸하지 않았다.

## 4. Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.90 | 0.75–1.0 | `acceptance.md:5-12` 의 판정 어휘 정의(정확 일치/폐집합/포착 seam/추가 파일)가 모호성을 선제 제거. 감점은 D6·D9 의 REQ↔AC·seam 의미 흔들림 |
| Completeness | 0.95 | 1.0 근접 | HISTORY·WHY(§B)·WHAT(§C)·HOW(plan)·REQUIREMENTS(§D)·ACCEPTANCE(16건)·Out of Scope 전부 존재. `spec.md:97-104` 의 제외 6항은 각각 구체 bullet |
| Testability | 0.72 | 0.50–0.75 | 대부분 이름 붙은 상수와의 정확 일치·폐집합·100% 커버리지로 기계화. 그러나 16건 중 **3건이 mutant-writable**(D2·D4) |
| Traceability | 1.00 | 1.0 | REQ 14건 전부 최소 1개 AC 가 인용, AC 16건이 인용한 REQ 전부 §D 에 존재(스크립트 대조 결과 UNDEFINED 0건) |

## 5. 전 AC mutant 표 (16/16 전수 — 표본 아님)

각 AC 에 대해 "명시된 단언을 **전부** 만족하면서 인용된 REQ 를 **위반**하는 구현" 을 실제로 작성 시도했다.

| AC | REQ | 판정 | 구체 mutant (writable 인 경우) |
|---|---|---|---|
| AC-CL-001 | 001 | **MUTANT-FREE** | 심볼 동등(`codexCmd.GroupID == ccCmd.GroupID`, `Parent()==rootCmd`) + 헤딩 출현 정확히 1 + 같은 블록 토큰 집합. 별도 동명 그룹을 만드는 우회가 헤딩 카운트에서 죽는다 |
| AC-CL-002 | 002 | **MUTANT-FREE** | 라우팅 표 폐집합 등식 + 미지 토큰 6칸 부정 방향 + cwd 3위치 교차(하위 디렉터리 칸이 `os.Getwd()` 판을 가른다) + `--` 꼬리 정확 일치. 빈틈을 찾지 못함 |
| AC-CL-003 | 003 | **MUTANT-FREE** (단 D9 모호성) | 진단 바이트 동등 + 리터럴 비시험 원본 전체 1회 + spawn↔exec 꼬리 상호 비교. 꼬리를 버리는 spawn 구현이 상호 비교에서 죽는다 |
| AC-CL-004 | 004 | **MUTANT-FREE** | 6상태 전문 정확 일치 + 두 partial 칸의 **배타** 단언 + 상태 토큰 폐집합. 고정 문안 출력 구현이 배타 단언에서 반드시 깨진다 |
| AC-CL-005 | 005 | **MUTANT-FREE** | 4칸(빈 문자열·공백뿐 포함) + 출처 폐집합 + 끝-구분자 `filepath.Join` 칸 + 홈 seam 호출 정확히 1회 + `HOME` 삭제 칸 |
| AC-CL-006 | 006 | **MUTANT-FREE** | 10칸 × rc 0 × 전문 정확 일치 × stderr 0바이트 × 오류 어휘 6낱말 합 0 × 완료 시 조치 부재 |
| AC-CL-007 | 007,010 | **MUTANT-FREE** | sentinel 을 **소비자 축(런처·MCP·web)** 으로 교차 + provider 리터럴 보유 파일 집합 폐집합 등식. 두 번째 분류 경로가 리터럴 등식에서 걸린다 |
| AC-CL-008 | 008 | **MUTANT-WRITABLE** | **D2**. `readCodexAuthFile` 이 파일을 읽어 거부(빈 `tokens`·미지 모드·파싱 실패)했을 때 `codexLoginStatusRunner` 를 **부르지 않고 즉시 `unknown` 을 반환**하는 구현. 1단 순수함수 표(30행)는 `classifyCodexAuthFile` 만 보므로 전부 통과, 통합 3칸은 (유효 파일→러너 0회) / (파일 **부재**→러너 호출) 뿐이라 전부 통과, 비밀값 4채널도 통과. REQ-CL-008 의 "an unparseable file … shall fall back to the command probe" 를 정면 위반 |
| AC-CL-009 | 009 | **MUTANT-FREE** | 13행 표 + 정규화 축 전조합 + **양방향 동치 속성 단언**(고정 seed 1,000건, 참조 문법을 시험이 독립적으로 보유). 룩업 표 구현이 속성 축에서 죽는다. 가장 강한 AC |
| AC-CL-010 | 009 | **MUTANT-WRITABLE** | **D2 와 동일 mutant**. 4번째 축(`auth.json` 파싱 실패)에서 러너의 반환값을 규정하지 않은 채 auth 행이 `unknown` 이기를 요구하므로, 하강을 생략하는 구현이 오히려 이 칸을 **더 쉽게** 만족한다. REQ-CL-009 만 놓고 보면 충분하나 REQ-CL-008 과 결합하면 잠재 모순 |
| AC-CL-011 | 012 | **MUTANT-FREE** | 진단 **줄 수 정확히 1** + 설치 상수 정확 일치 + 행 라벨 집합 폐집합 등식(조기 반환이 여기서 죽는다) + 배선 2상태 교차 |
| AC-CL-012 | 013 | **MUTANT-WRITABLE** | **D4**. 하드코딩 절대 경로(`os.WriteFile("/var/tmp/moai-codex.lock", …)` — `os.TempDir()` 를 거치지 않아 `TMPDIR` 리디렉션에 걸리지 않는다)에 쓰는 구현. 스냅샷 대상이 "격리 홈 트리 전체" 라는 **열거**이므로 그 밖의 쓰기가 관측되지 않는다. REQ-CL-013 의 "it does not write" 는 열거보다 넓다 |
| AC-CL-013 | 014 | **MUTANT-FREE** | 리플렉션으로 **모든** string/[]string 필드 + 모든 플래그 usage 순회 + 금지 8패턴 + **비-ASCII 0자**. 한국어 도움말·내부 식별자 어느 쪽도 빠져나갈 수 없다 |
| AC-CL-014 | 전 REQ | **MUTANT-FREE** (단 D3 로 자기모순) | 빌드 태그 0건 + GOOS skip 정확히 3(이름 상수 고정) + `GOOS=windows go test -c`. 스텁 갈아끼우기를 실제로 막는다. 다만 이 단언 자체가 §C.5 설계와 충돌 → D3 |
| AC-CL-015 | 010 | **MUTANT-FREE** | skip 수 0 + 시험 함수 **이름 목록** 삭제·개명 0 + 케이스 수 하한(표 행 수 연동) + 세 순수함수 **구문 커버리지 100%** + `SENTINEL-VER-9x9` 정확 일치. 몰래 넣은 분기가 커버리지에서 드러난다 |
| AC-CL-016 | 011 | **MUTANT-FREE** | 기동 실행 파일 basename 집합 폐집합 `{codex}` 를 **두 축**(런타임 포착 + 정적 0번 인자)으로 + 실패 seam rc 전파 + fixture 출력 바이트 동일 passthrough. `open -b` 번들 ID 우회가 폐집합에서 걸린다 |

**요약: 13 MUTANT-FREE / 3 MUTANT-WRITABLE (AC-CL-008, AC-CL-010, AC-CL-012).**

## 6. Defects Found

D1. **self-test 의 복원 검사가 거짓 통과하고, spec.md 가 그 복원을 사실로 단언한다** — `.moai/reports/t197/gate-selftest.sh` / `spec.md:40` — Severity: **critical** — Class: **blocking**
관측: self-test 실행 직후 `git status --porcelain` 이 ` M .moai/reports/t197/citation-manifest.txt` 를 보고했다(실행 전은 clean). 남은 diff 는 C4 주입이 append 한 manifest 한 줄이다:
```
+.moai/specs/SPEC-CODEX-INIT-001/spec.md ~ L211-214 ~ 경로 상수 ~ writeAtomic\(
```
그런데 같은 실행이 `restore check: ok: both files byte-identical to the backup` 과 `SELFTEST PASS` 를 출력했다 — **복원 검사가 백업한 2개 파일만 보고, 실제로 건드린 파일 집합(최소 3개)을 보지 않는다.** 결과가 특히 나쁘다: 그 오염된 트리에서 게이트를 다시 돌리면 `rc=1` 이다(내가 관측). 즉 **게이트를 검증한다는 스크립트가 게이트를 깨진 상태로 남긴다.**
`spec.md:40` 은 "이 게이트 자체는 `gate-selftest.sh` 가 검증한다 — … 건드린 파일을 바이트 동일로 되돌린다" 고 단언한다. 관측이 이 단언을 반증하므로 미관측 주장(`verification-claim-integrity.md` §1)이다. 이것이 반복 7 F2("판정하지 않고 보고하는 스크립트")의 재발 지점이다.
부수 위험: 공유 체크아웃의 **추적 파일**에 결함을 주입하므로, 동시 세션이 그 오염을 스테이지·커밋할 수 있다.
**Required fix**: (1) 복원 검사를 "건드린 파일 집합" 이 아니라 **코퍼스 전체의 `git status --porcelain` 이 빈 출력** 인지로 판정하고, 아니면 rc 1 로 실패시킨다. (2) 주입을 추적 파일이 아니라 `mktemp -d` 로 복제한 임시 트리에 하고 게이트가 그 사본을 읽도록 매개변수화한다. (3) `spec.md:40` 의 복원 단언은 (1)(2) 가 실제로 관측된 뒤에만 유지한다.

D2. **REQ-CL-008 이 요구한 "거부된 파일 → 명령 프로브로 하강" 을 어떤 AC 도 단언하지 않는다** — `acceptance.md` AC-CL-008 통합 3칸 / AC-CL-010 4번째 축 — Severity: **critical** — Class: **blocking**
REQ-CL-008(`spec.md:123`)은 "In every other case — an unknown mode, empty or missing credential material, an unparseable file, or no file at all — it shall **fall back to the command probe**" 를 요구한다. 그러나 통합 3칸은 (a) 유효 파일 → 러너 0회, (b) 파일 **부재** + stderr, (c) 파일 **부재** + stdout 뿐이다. **파일이 존재하되 거부되는** 경우에 러너가 호출되는지를 세는 칸이 없다. AC-CL-010 의 4번째 축(파싱 실패)은 러너 반환값을 규정하지 않은 채 `unknown` 을 기대하므로 하강 생략 구현을 오히려 통과시킨다.
Mutant: `readCodexAuthFile` 이 `ok=false` 를 돌려주면 러너를 부르지 않고 `unknown` 을 반환 — 16개 AC 전부 통과, REQ 위반. 로그인된 머신에서 `auth.json` 이 stale 하면 §F 성공 판정("`auth: chatgpt` 를 보고")도 함께 깨진다.
**Required fix**: AC-CL-008 통합 표에 거부 사유별 칸을 추가한다 — 최소 3행(빈 `tokens` / 미지 `auth_mode` / 파싱 실패) 각각에 대해 **러너가 긍정 결과를 돌려주도록 스텁**하고, 기대를 `AuthProvider == "chatgpt"` **이고 러너 호출 1회** 로 못 박는다. 동시에 AC-CL-010 4번째 축의 러너 상태를 "두 스트림 비어 있음" 으로 명시해 D2 칸과 충돌하지 않게 한다.

D3. **프로세스 교체 설계와 "OS 빌드 태그 0건 + Windows 컴파일" 단언이 양립 불가 (MP-6 / D8)** — `plan.md` §C.5 · `acceptance.md` AC-CL-003 · AC-CL-016 말미 vs `acceptance.md` AC-CL-014 — Severity: **critical** — Class: **blocking**
`plan.md` §C.5 는 `cli` 를 "`exec` 로 `codex` 를 프로젝트 루트에서 **교체 실행**" 으로 규정하고, AC-CL-003("현재 프로세스는 exec 으로 교체되지 않는다" — spawn 경로 한정)과 AC-CL-016 말미("exec 교체 경로에서는 출력 경로 자체가 관측되지 않기 때문")가 교체 의미론을 확인한다. Go 에서 프로세스 교체는 `syscall.Exec` 이며 이는 **unix 전용** 이다(Windows 에 해당 심볼이 없다). acceptance.md:287 이 `syscall.Exec` 을 명시적으로 언급한다.
그런데 AC-CL-014 는 이 SPEC 이 추가한 파일에서 **OS 빌드 태그 0건**과 `_windows.go`/`_unix.go` 접미 0건을 요구하면서 동시에 `GOOS=windows go vet` 과 `GOOS=windows go test -c ./internal/cli/` 가 rc 0 이기를 요구한다. 세 요구는 동시에 만족될 수 없다 — 교체를 구현하면 Windows 컴파일이 깨지고, 빌드 태그로 가르면 태그 0건 단언이 깨진다.
`syscall` 이 SPEC 본문에 등장하는데 `//go:build` 제약도 명시적 크로스 플랫폼 면제 조항(`EXCL-…`)도 없으므로 **D8-3 BLOCKING** 이며 MP-6 실패다. 교훈 #21(Windows `syscall.Flock` 빌드 태그 누락)과 같은 형태다.
**Required fix**: 셋 중 하나를 택해 SPEC 에 명시한다 — (a) 교체를 버리고 전 플랫폼에서 `os/exec` + 종료코드 전파로 통일(AC-CL-016 의 rc 동등 단언과 이미 정합적이며 빌드 태그 0건을 유지한다), 또는 (b) 교체를 유지하되 AC-CL-014 에 **명시적 크로스 플랫폼 면제 조항**과 허용 빌드 태그 파일 집합을 상수로 열거하고 Windows 판의 동작(교체 대신 대기)을 별도 AC 로 판정한다. 어느 쪽이든 §C.5·AC-CL-003·AC-CL-016 의 교체 서술을 그 결정에 맞춰 일치시킨다.

D4. **AC-CL-012 의 쓰기 관측 범위가 열거라서 격리 홈 밖 쓰기를 보지 못한다** — `acceptance.md` AC-CL-012 — Severity: **major** — Class: **blocking**
AC 는 `HOME`·`XDG_*`·`TMPDIR` 을 격리 홈 아래로 돌리고 **그 트리 전체**를 스냅샷한다 — 앞선 "세 트리 열거" 보다 넓어졌으나 여전히 열거다. Mutant: `os.WriteFile("/var/tmp/moai-codex.lock", …)` 처럼 `os.TempDir()` 을 거치지 않는 하드코딩 절대 경로에 쓰는 구현. `TMPDIR` 리디렉션에 걸리지 않고 스냅샷 밖이므로 네 형태 전부 통과하면서 REQ-CL-013 의 "it does not write" 를 위반한다. AC 자신이 "REQ 후단 'it does not write' 는 열거보다 넓다" 고 적어 놓고 한 겹 넓은 열거로 답했다.
**Required fix**: 파일 쓰기를 계수 seam(예: `codexWriteFile` 변수)으로 일원화하고 **경로와 무관하게 쓰기 모드 호출 횟수 0** 을 단언한다. 스냅샷은 보조 축으로 남긴다. 추가로 이 SPEC 이 추가한 파일에서 `os.WriteFile`·`os.Create`·`os.MkdirAll`·`os.OpenFile`(쓰기 플래그) 호출 지점 집합이 **공집합**임을 정적으로 단언하면 열거가 아니라 폐집합이 된다.

D5. AC-CL-009 산문의 케이스 수가 표와 어긋난다 — `acceptance.md` AC-CL-009 정규화 축 문단 — Severity: minor — Class: **optional**
산문은 "고정 입력 **11칸**을 외운 룩업 표는 이 축에서 죽는다" 인데 표의 데이터 행은 **13** 이다(내가 세었다: 충돌 2행이 추가되면서 11→13 이 됐고 산문이 따라오지 않았다). AC-CL-015 의 케이스 하한은 "AC-CL-009 표의 행 수" 로 표에 연동돼 있으므로 **어떤 단언도 약화되지 않았다** — 순수 산문 드리프트다. 다만 0.7.0 이 "산문-전사본 드리프트 정정" 을 표방한 라운드라는 점에서 같은 부류의 잔존이다. **Required fix**: "11칸" → "13칸", 또는 수를 적지 않고 "위 표의 고정 입력" 으로 바꿔 재발을 구조적으로 막는다.

D6. REQ-CL-004 가 5행을 열거하는데 AC-CL-011 은 6개 라벨 폐집합을 요구한다 — `spec.md:116` vs `acceptance.md` AC-CL-011 — Severity: minor — Class: **optional**
REQ-CL-004 는 "codex binary path, codex version, resolved `CODEX_HOME`, auth provider, and project wiring state" 5행이다. AC-CL-011 은 라벨 집합이 `{codex, home, auth, wiring, agents, harness}` 와 **같다**(6개)를 요구하고 `plan.md` §C.4 리드아웃도 6행이다. REQ 를 문자 그대로 만족하는 5행 구현이 AC 에서 실패한다. **Required fix**: REQ-CL-004 에 `agents`·`harness` 두 행을 추가하거나, AC-CL-011 의 폐집합을 REQ 가 열거한 5행 + 명시적으로 허용하는 부가행으로 정리한다.

D7. plan.md 의 문법과 acceptance.md 의 참조 문법이 다르다 — `plan.md` §C.2 vs `acceptance.md` AC-CL-009 — Severity: minor — Class: **optional**
plan 은 `(?i)^logged in using (chatgpt|api key)$`, acceptance 는 `(?i)^[ \t]*logged in using (chatgpt|api key)[ \t]*\r?$`. acceptance 의 정규화 축(후행 `\r`·공백·탭)은 plan 문법으로는 통과할 수 없다. acceptance 가 구속력을 가지므로 실무 영향은 없으나 두 문서가 어긋나 있다. **Required fix**: plan §C.2 의 문법을 acceptance 판본으로 교체한다.

D8. plan §C.3 의 `CODEX_HOME` 부재 조치(`codex login`)를 어떤 AC 도 판정하지 않는다 — `plan.md` §C.3 — Severity: minor — Class: **optional**
AC-CL-005 는 4칸 모두 경로 해석만 보고 `missing` 상태나 그 조치 문구를 보지 않는다. AC-CL-012 는 "생성하지 않음" 만 본다. **Required fix**: AC-CL-005 에 `missing` 상태 행과 조치 문구 정확 일치 칸을 추가하거나, plan §C.3 에서 그 조치 서술을 뺀다.

D9. AC-CL-003 의 spawn↔exec 포착 동등이 seam 의미에 따라 작성 불가능할 수 있다 — `acceptance.md` AC-CL-003 — Severity: minor — Class: **optional**
"spawn 에 포착된 `(program, argv)` 가 같은 입력의 **exec 경로 포착값과 토큰 단위로 동일**" 을 요구한다. 그런데 spawn 경로가 실제로 띄우는 프로세스는 tmux 이므로, 포착 seam 이 `program` 에 실행 대상(tmux)을 기록하면 이 등식은 성립할 수 없고, 기록 대상이 "최종 기동 대상(codex)" 이어야만 성립한다. `acceptance.md:11` 의 포착 seam 정의는 `(program, argv, cwd, 형태)` 라고만 적어 어느 쪽인지 고정하지 않는다. **Required fix**: 포착 seam 정의에 "spawn 경로에서 `program`/`argv` 는 tmux 가 아니라 **새 창에서 실행될 대상** 을 기록한다" 를 명시한다.

## 7. Regression Check (반복 7 결함 대비)

- **F1 — 인용의 줄 범위가 주장을 담지 않음**: **RESOLVED**. `citation-sweep.sh` C4 가 정규식으로 줄 내용을 판정하고, 내가 관측한 230 검사 전수 통과 + self-test 의 C4 주입 거부(rc 1)가 근거.
- **F2 — 판정하지 않고 보고만 하는 스크립트**: **PARTIALLY RESOLVED**. 게이트 본체는 rc 로 판정한다(RESOLVED). 그러나 그 게이트를 검증하는 self-test 의 **복원 검사**가 같은 결함을 재현한다 — D1. 정체(stagnation)는 아니다: 결함이 같은 자리에 남은 것이 아니라 한 겹 위로 이동했고, 아래층은 실제로 고쳐졌다.

## 8. Gaps — 내가 확인하지 **않은** 것

미관측을 명시한다. 아래 항목에 대해 나는 아무것도 주장하지 않는다.

1. **구현 코드 일체**. 이 트리에는 `internal/cli/codex_launcher.go`·`codex_readiness.go` 가 아직 없다(plan-phase). 16개 AC 가 실제로 작성 가능한지는 SPEC 문면으로만 판단했고, 컴파일·실행으로 확인하지 않았다. D3 의 Windows 컴파일 불가는 `syscall.Exec` 의 플랫폼 가용성에 대한 정적 지식에 근거하며, `GOOS=windows` 빌드를 실제로 돌려 관측하지는 않았다.
2. **`probe-output.txt` 전사본 내용의 해석 타당성**. 게이트는 인용이 그 줄 범위에 있고 그 줄이 주장 형태를 담는지까지만 판정한다(spec.md:39 가 스스로 그렇게 적는다). 나는 전사본 316줄을 처음부터 끝까지 읽어 산문이 그 근거의 **올바른 해석**인지 독립 검증하지 않았다. §A 의 67초·11 TOML·24바이트 같은 수치는 게이트가 결속한 범위 안에 있다는 것까지만 확인했다.
3. **`SPEC-CODEX-INIT-001` 의 내용**. codex 백엔드 지적 1·2·3 은 그 SPEC 소관이라 범위 밖으로 두었고, 그 지적들이 옳은지 재현하지 않았다. INIT-001 은 별도 감사 대상이다.
4. **`mutants-launcher.md` / `rules-applied.md` / `measurement.md` 의 내용**. 지시대로 보고서가 아니라 SPEC 을 판정했다. 이 문서들이 주장하는 mutant 분석과 내 §5 표가 일치하는지 대조하지 않았다 — 내 표는 SPEC 문면에서 독립적으로 작성했다.
5. **게이트·self-test 의 코드 자체**. 두 스크립트를 실행해 rc 와 출력을 관측했으나, 소스를 줄 단위로 읽어 검사 로직이 우회 가능한지 감사하지 않았다. D1 은 스크립트를 읽어서가 아니라 **실행 후 `git status` 관측**으로 발견했다.
6. **동시 세션 유무**. `git fetch` / `rev-list --left-right` 로 origin 대비 발산을 확인하지 않았다. codex 백엔드는 자기 실행에서 `FETCH_HEAD: Operation not permitted` 로 원격 기준선 갱신에 실패했다고 보고했다 — 그 제약이 내 세션에도 적용되는지 시험하지 않았다.
7. **Tier M 반복 상한**. 이 SPEC 은 `tier: M` 이며 `plan_audit_tier_ceilings` 상 상한은 2회인데 현재 8회차다. 상한 초과 자체는 이미 별건(procedure defect)으로 기록돼 있다고 전해 들었으나, `.moai/config/sections/harness.yaml` 을 열어 상한값을 확인하지 않았다. Tier M 의 PASS 임계값도 `spec-workflow.md` 를 열어 확인하지 않았다 — 다만 must-pass(MP-6) 실패이므로 임계값과 무관하게 FAIL 이다.
8. **Windows 실행**. AC-CL-014 가 스스로 Gap 으로 선언한 항목이며, 나 역시 관측하지 않았다.

## 9. Recommendation

**FAIL — 반복 9 로 넘긴다.** 다만 두 가지를 함께 기록한다.

**(가) 진전은 실재한다.** F1 은 닫혔고, 16개 AC 중 13개가 mutant-free 다. AC-CL-009 의 양방향 동치 속성 단언과 AC-CL-015 의 100% 구문 커버리지 요구는 이 저장소 SPEC 중 가장 강한 형태에 속한다. 남은 결함은 "약한 SPEC" 이 아니라 **세 군데의 구체적 구멍**이다.

**(나) 차단 4건의 수정 순서** — 서로 독립이므로 병렬 가능:

1. **D3 먼저** — 유일한 must-pass 실패이고, 결정(교체 유지/포기)이 §C.5·AC-CL-003·AC-CL-014·AC-CL-016 넷을 동시에 움직이므로 나머지 수정보다 먼저 확정해야 재작업이 없다. 권고는 (a) 전 플랫폼 `os/exec` — AC-CL-016 의 rc 전파 단언과 이미 정합하고 빌드 태그 0건을 지킨다.
2. **D2** — AC-CL-008 통합 표에 거부 사유 3행 추가 + AC-CL-010 4번째 축의 러너 상태 명시. 이 SPEC 의 존재 이유(§F 성공 판정)와 직결된다.
3. **D4** — 쓰기 계수 seam + 정적 공집합 단언으로 열거를 폐집합으로 교체.
4. **D1** — self-test 복원 검사를 `git status --porcelain` 빈 출력 판정으로 바꾸고 주입을 임시 트리로 옮긴 뒤, `spec.md:40` 의 단언을 관측으로 뒷받침한다.

**(다) optional 5건(D5–D9)** 은 오케스트레이터 재량이다. 어느 것도 AC 를 mutant-writable 로 만들지 않는다. 다만 D6(REQ-CL-004 5행 vs AC 6행)은 run-phase 에서 실제 마찰을 만들 가능성이 있어 D3 수정 시 함께 처리하는 편이 싸다. **optional 5건을 이유로 FAIL 을 만들지 않았다** — 판정은 전적으로 MP-6(D3)과 mutant-writable 3건(D2·D4)에 근거한다.

**(라) 절차 관측**: 8회차는 Tier M 상한(2)을 크게 넘었고, 점수는 반복 7 대비 하락하지 않았으나 상한 초과 자체가 이미 별도 결함으로 제기돼 있다. D3 는 **8회차에 처음 제기되는 must-pass 실패**다 — 이는 SPEC 이 나빠져서가 아니라 이전 회차들이 크로스 플랫폼 축을 보지 않았기 때문이며, 반복을 더 도는 것보다 **범위 축소 또는 운영자 판단**이 유효할 수 있는 신호로 읽는 것이 타당하다.
