# SPEC Review Report: SPEC-EVIDENCE-CITATION-CANON-001

Iteration: 2/2 (Tier M ceiling)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.85** (Tier M PASS 임계 0.80)
Delta: iter1 **0.70 (FAIL)** → iter2 **0.85** — 단조 증가(+0.15). 점수 퇴행 없음, STOP 에스컬레이션 해당 없음.

측정 트리: 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t375`, 브랜치 `WT-state-evidence-canon`, HEAD `b64043481`. primary 체크아웃에서 잰 값은 그렇게 표시한다. 아래 모든 수치는 이 감사가 이 트리에서 직접 실행한 명령의 출력이다.

범위: 리드 지시대로 **수리 델타와 수리가 새로 들여온 것**에 한정했다. iter1에서 재측정해 통과 확인한 항목(must-pass 7건, §1.1의 나머지 측정치, t373 선례 충실도, 미러 정합성 D14, 베이스라인)은 다시 감사하지 않았다. 예외는 두 가지 — must-pass는 수리가 깨뜨렸을 수 있어 값싼 재확인만 했고, `go test ./internal/web/...`는 이번 판에서 **새로 추가된** 명령이라 실행했다.

Reasoning context ignored per M1 Context Isolation.

---

## Must-Pass Results (수리가 깨뜨리지 않았는지 확인만)

- **[PASS] MP-1** — `REQ-ECC-001` … `REQ-ECC-013` 13건, 결번·중복 없음, zero-padding 일관. (정의 **순서**는 013이 010 뒤·011 앞에 놓여 비단조 — 집합은 완전하므로 MP-1 위반 아님, N5로 별기.)
- **[PASS] MP-2** — 신설·개정 요구 전부 GEARS 대응: 009 = Where + `shall`/`shall not` 복합, 013 = ubiquitous `shall`, 004·011·012 = `shall` + `shall not`, 008 = ubiquitous. AC 층(Given-When-Then)은 검증 층이므로 여기서 감점하지 않음.
- **[PASS] MP-3** — version `"0.2.0"`으로 상향된 뒤에도 정본 12필드 전부 존재·타입 정합, `tier: M` 부가. 거부 별칭 없음.
- **[N/A] MP-4** — 변동 없음.
- **[PASS] MP-5** — 변동 없음.
- **[N/A] MP-6** — `grep -c 'syscall' *.md` → 4파일 전부 0.
- **[PASS] MP-7** — `[NEEDS CLARIFICATION` 마커 재귀 검색 → 0행.

---

## 증거 기반 재생성 검증 — 판별식은 실행 가능한가

리드 질문: "예측이 *서술만* 되고 실행되지 않으면 D3/D4와 같은 결함이 옷만 갈아입은 것이다."

**답: 실행 가능하다.** 두 스크립트를 이 트리에서 직접 돌렸고, SPEC §1.1 표의 일곱 값이 **전부 정확히 재생산**됐다.

```
(citers 목록 생성) → 184
python3 .moai/reports/t375/measure_citations.py <citers>
  files scanned: 184 | path occurrences: 515 | concrete occurrences: 346
  files carrying >=1 concrete citation: 124 | distinct concrete cited paths: 231
python3 .moai/reports/t375/measure_resolution.py <citers>
  rows 231 | primary 42 | worktree 0
```

파생 표도 대조했다: `diff` 결과 `cited-path-resolution.txt`는 새로 돌린 출력과 **바이트 동일**, `citation-figures.txt`는 스크립트 출력과 일치, `citing-files.txt`는 184행(= citers 모집단)이다. §1.1이 적은 "232행 = 헤더 1 + 231"도 실측 232로 일치.

판별식(`^[A-Za-z0-9]`)이 **제외 목록이 아니라 양성 문자 클래스**로 코드에 박혀 있는 것은 D3 재발 방지로서 옳은 형태다 — 무엇을 빼야 하는지 기억할 필요가 없다.

**§1.1.1은 수행적이지 않고 정확하다.** 세 변형이 129/131/132로 갈렸다는 서술 자체는 이 감사가 재현할 수 없지만(그 변형들이 남아 있지 않다), 검증 가능한 부분은 전부 참이다: 옛 명령 `grep -r`이 이 트리에서 187·562를 내는 것, 그 차이가 SPEC 자신의 산출물 3개인 것, 131이 어느 명령으로도 재현되지 않는 것. 지우지 않고 남긴 판단은 옳다.

**다만 이 증거 기반에 남은 문제 둘이 있다 — N1·N2 참조.**

---

## Category Scores

| 차원 | iter1 | iter2 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | **0.85** | 해석을 요구하던 세 지점이 전부 텍스트로 결정됐다: 허용목록 단위(REQ-ECC-009), `.gitkeep` 여부(§4.2 판정 + REQ-ECC-011의 `shall not`), Tier 범위(§B 11~18 + M2 재확인). §1.1.1·§4.3은 이례적으로 명시적이다. 감점: N2 과잉 주장, N3·N4 단위 미끄러짐. |
| Completeness | 0.80 | **0.90** | D10 닫힘 — REQ-ECC-002 주어가 doctrine 표면 문서로 넓어져 `manager-lead.md`·출력 스타일이 요구 층에 대응하고, REQ-ECC-013 신설. carve-out 2건 완비. §5가 부채 2건을 후속 카드로 라우팅. Out of Scope H3 3개 + 구체 불릿. 감점: N1(증거 기반의 추적 상태). |
| Testability | 0.55 | **0.80** | D6 닫힘 — AC-ECC-002의 고정 문구 3종이 오늘 전부 0이고 AC-ECC-001의 둘은 전부 1임을 **직접 확인**했다(아래 표). D2 닫힘(AC-ECC-015). D11 닫힘. 감점: D7 잔여(단언 #1에 판정 명령 없음), D1 잔여(작은 하위트리 구멍). |
| Traceability | 0.70 | **0.85** | REQ 13건 전부 ≥1 AC, AC 15건 전부 REQ에 매핑, **plan.md 절 번호를 가리키는 매트릭스 행 0개**(실측). REQ-ECC-008이 AC-ECC-015라는 실판정자를 얻었다. 감점: N1, N5. |

**총점**: (0.85 + 0.90 + 0.80 + 0.85) / 4 = **0.85** ≥ 0.80.

### AC 판정력 실측 (수리 이전 트리, HEAD `b64043481`)

15건 중 오늘 확인 가능한 12건을 전부 돌렸고, **SPEC이 적은 기대 상태와 12/12 일치**했다.

| AC | 명령 | SPEC 기대 | 실측 |
|---|---|---|---|
| 001a/b | `SHALL be persisted under` / `evidence is persisted under` | 오늘 1 / 1 | **1 / 1** ✓ |
| 002a/b/c | `machine-local scratch` / `export before citing` / `.moai/reports/<card-id>` | 오늘 0 / 0 / 0 | **0 / 0 / 0** ✓ |
| 003a/b | `names one file` / `never the directory\|wholesale` | 오늘 0 / 0 | **0 / 0** ✓ |
| 004 | `residual-risk` | 오늘 0 | **0** ✓ |
| 005a/b | `state/verify/snapshots` / `internal/web/events.go` | 오늘 0 / 0 | **0 / 0** ✓ |
| 006 | `go test ./internal/verify/... ./internal/web/...` | rc=0 | **rc=0** ✓ (신규 `web` 포함) |
| 007a/b | `canonical persistence location` / `machine-local scratch` | 오늘 1 / 0 | **1 / 0** ✓ |
| 008 | `state/verify` in `moai.md` | 오늘 3 | **3** ✓ |
| 009 | progress.md `gate.md\|loop.md\|run.md` | 사후 ≥3 | **0** (오늘 RED, 판정력 있음) ✓ |
| 014 | `.gitignore` navigator / observability | 각 0 | **각 0** ✓ |

AC-ECC-014의 3·4번은 오늘 이미 참이지만 이는 D6 결함이 **아니다** — REQ-ECC-011/012가 `shall not`이므로 이 단언들은 M5 도중의 회귀를 막는 보존 단언이다. AC-ECC-002가 새 편집의 착지를 증명해야 했던 것과 성격이 다르다.

---

## 결함별 종결 표 (D1-D13)

| # | 상태 | 근거 |
|---|---|---|
| **D1** 가드 하한 공허 | **부분 종결** | 하한이 실측 모집단에서 도출됨(루트 **363**, 미러 **340** — 둘 다 이 감사가 재측정), 트리별 300. 최대 단일 하위트리 붕괴는 막힌다(`.claude/skills` 251 < 300). **그러나 잔여 구멍 있음 — 아래 §D1 잔여.** |
| **D2** REQ-ECC-008 판정자 부재 | **종결** | AC-ECC-015 신설: 방문 트리 루트 목록 단언 + 트리별 하한 + 미러 정합 + 방문 뮤테이션. 매트릭스에서 REQ-ECC-008에 매핑. |
| **D3** 131 미귀속 | **종결** | 131 제거, 124로 대체, 스크립트가 정확히 재생산. §1.1.1이 경위 기록. |
| **D4** 표 명령이 값을 못 만듦 | **종결** | 표 출처가 tracked-set grep과 두 스크립트로 교체. 7행 전부 재생산 확인. |
| **D5** carve-out 누락 | **종결** | §1.4가 둘을 명명, REQ-ECC-006이 둘을 열거, AC-ECC-005가 둘을 grep, AC-ECC-006이 `./internal/web/...`로 확장(rc=0 확인). `events.go:29`가 디렉터리 전체를 감시한다는 서술도 소스에서 확인. |
| **D6** AC-ECC-002 판정력 | **종결** | 고정 문구 3종 전부 오늘 0. 넓은 단어 검색 폐기. |
| **D7** 허용목록 단위 | **부분 종결** | REQ-ECC-009가 파일+리터럴을 의무화하고 파일 단독 항목을 `shall not`으로 금지. AC-ECC-013이 두 단언을 적음. **그러나 판정 명령이 단언 #1을 검사하지 않음 — 아래 §D7 잔여.** |
| **D8** `.gitkeep` 뿔 | **종결** | 판정 역전. 근거를 이 감사가 확인: 텔레메트리 형제 `.gitkeep`이 **추적되고** **템플릿에도 실려 있다**(`internal/template/templates/.moai/evolution/telemetry/.gitkeep`). `os.Stat` opt-in 관문도 소스 확인(`post_tool_duration.go:118-119`, REQ-CC2122-HOOK-001-003 인용). 불활성/활성 구분이 반대 뿔에 답한다. AC-ECC-014의 2·3번이 단언. |
| **D9** 판별식 미적용 | **반증으로 종결** | 아래 §D9 판정 참조 — **이 감사의 iter1 발견이 틀렸다.** |
| **D10** 요구 층 공백 | **종결** | REQ-ECC-002 주어 확장 + REQ-ECC-013 신설. plan.md 절 번호를 가리키는 매트릭스 행 **0개**(실측). |
| **D11** 인용 넓이 상한 | **종결** | REQ-ECC-004가 "파일 하나를 이름 붙인다 / 디렉터리는 안 된다"로 넓이를 구속하고, 인용 뒤 노트가 004·005를 **하나의 상한**으로 묶는다. AC-ECC-003이 양쪽을 grep. |
| **D12** Tier vs 18파일 | **종결** | §B가 11~18 조건부 + [HARD] M2 직후 Tier 재확인. 요구/AC 13/15로 상한 16/16 내. |
| **D13** 124를 세션으로 셈 | **종결** | "최상위 124개, 그중 1개는 `snapshots`, 세션 123개". |
| (D15) 부채 서술 | **종결** | §5가 124/231/189로 갱신, 전부 재도출 확인(231−42=189). |

### D1 잔여 — 하한 300은 큰 붕괴는 막고 작은 절단은 막지 못한다

리드 질문("300은 363의 83%인데 여유가 실제로 구속하는가")에 대한 실측 답이다.

```
.claude/rules 88 | .claude/agents 21 | .claude/output-styles 3 | .claude/skills 251   (합 363)
mirror: rules 83 | agents 11 | output-styles 3 | skills 241                            (합 338, +2 = 340)
```

- **막힌다**: 최대 단일 하위트리로 붕괴 → `.claude/skills` 251 < 300 ✗. 초판 하한 7이 통과시키던 붕괴는 전부 걸린다.
- **막히지 않는다**: 루트에서 300을 유지하려면 최대 63개까지 뺄 수 있다. `.claude/agents`(21)를 빼면 **342 ≥ 300 통과**, `.claude/output-styles`(3)를 빼면 **360 통과**. 미러도 같다(340−11=329 통과, 340−3=337 통과).

빠져도 통과하는 두 하위트리가 하필 **이 SPEC이 자기 최강 사례로 지목한 곳**이다 — `.claude/agents/moai/manager-lead.md`(§1.3 "가장 강한 형태")와 `.claude/output-styles/moai/moai.md`(AC-ECC-008의 3지점). 범위에서 `.claude/agents/`가 빠지고 `manager-lead.md`가 되돌려지면 가드는 342개를 스캔하고 위반 0을 보고하며 초록이다.

AC-ECC-015는 **트리** 루트 방문을 단언하지 하위트리 방문을 단언하지 않으므로 이것을 잡지 못한다. 닫는 법은 이미 문서 안에 있다: AC-ECC-015가 세운 "방문한 루트 목록을 반환해 단언한다" 패턴을 하위트리 4개로 한 단계 내리면 된다.

Severity: **major** · Class: **blocking** (수리는 AC 한 줄 확장)

### D7 잔여 — 단언 #1에 판정 명령이 없다

AC-ECC-013은 두 가지가 **둘 다** 성립해야 한다고 적는다: (1) 항목이 파일+리터럴 쌍이고 파일 단독 항목이 없을 것, (2) 항목 수가 상수와 일치할 것. 그런데 적힌 판정 명령은 하나뿐이다.

```
grep -c 'allowlistSize\|wantAllowlist' internal/template/evidence_citation_guard_test.go   # 기대 >=1
```

이 grep은 **개수 단언의 존재**만 본다. 단언 #1(파일 단독 항목 부재)을 검사하는 명령은 없고, AC-ECC-009처럼 "기계 판정 명령 없음 + 대체 판독법"을 밝히지도 않는다. acceptance.md 자신이 서두에 세운 기준 — "명령이 없는 항목은 그 사실과 대체 판독법을 적는다" — 을 이 항목이 지키지 않는다.

D7의 실질(단위 미지정)은 REQ-ECC-009로 닫혔다. 남은 것은 그 단위가 **기계로 확인되는가**이고, 그 부분이 열려 있다.

Severity: **major** · Class: **blocking** (단언 #1을 검사하는 명령 한 줄, 또는 대체 판독법 한 문장)

---

## D9 판정 — 이 감사의 iter1 발견이 틀렸다

리드가 요구한 대로 근거만 보고 판정한다.

**1. 숨은 전제 추출은 §4.1에 충실한가 — 충실하다. 필요 이상으로.**

iter1에서 이 감사가 **직접 인용한** t373 원문에 그 전제가 들어 있었다: "처분되지 않았다는 뜻이기 때문이다 … `git status`에 뜨는 것이 **처분이 안 됐다는 유일한 신호**". 신호를 만드는 것이 처분이라는 말이 인용문 안에 있었고, 이 감사는 그것을 인용해 놓고 적용하지 않았다. §4.3의 추출은 편의적 재구성이 아니라 원문에 명시된 것을 읽어낸 것이다.

**2. 반증이 실제로 발견을 무너뜨리는가 — 무너뜨린다. SPEC보다 넓은 범위에서 확인했다.**

SPEC이 적은 범위(`internal/navigator/fix/` + `navigator_fix.go`)는 좁아서 무언가를 가릴 수 있으므로 넓혀서 다시 쟀다.

```
grep -rn 'RemoveAll' internal/navigator/ internal/cli/navigator_fix.go   (비-테스트) → 0행
grep -rnE 'os\.Remove|os\.Rename|\.Remove\(' internal/navigator/ …      (비-테스트)
  → os.Rename 히트 전부 tmp→final 원자적 쓰기(route/write.go:131, tiers/overlay.go:258,
    tiers/symbol_narrative.go:153, fix/request.go:459, sync/write.go:51). os.Remove 0건.
```

처분 코드가 **어떤 원시 동사로도** 없다. 넓힌 범위가 좁힌 범위와 같은 답을 준다.

그리고 비대칭의 반대편도 확인했다 — `chain/`은 실제로 처분한다:

```
internal/cli/chain.go:104  disposeLegacyChainDir(...)
  → migrateChainEvents(...)
internal/cli/chain.go:151  if err := os.Remove(src); err != nil { … }
```

**성공 경로가 원본을 지운다.** 따라서 `chain/`의 잔존은 미완의 신호이고, `fix-drafts/<id>/`의 잔존은 성공한 실행에서도 똑같이 남아 아무 신호가 아니다. **모양이 같지 않다.** iter1 D9의 중심 추론 — "`chain/`·`migrate-tx`와 같은 모양" — 은 **틀렸다.**

**3. 살아남은 두 근거가 넓힌 결론을 지탱하는가 — 지탱한다. 과잉 확장이 아니다.**

- 근거 1(`applied.json`이 진짜 완료 신호이고 `.gitignore`는 경로에 걸리지 내용에 걸리지 않는다)은 `fix-drafts/` 한정이고 독립적으로 옳다. `types.go:77`이 승인 후 `applied.json`을 쓴다는 서술을 확인했다.
- 근거 2(존재-부재 비대칭)는 **일반적**이고, 초판이 이미 나머지 navigator 산출물 전부에 적용하던 논거다. `fix-drafts/`로 확장한 것은 새 주장을 더한 것이 아니라 **정당화되지 않은 예외를 없앤 것**이다. 오늘 navigator 아래 파일은 추적 템플릿 `symbols/narrative.template.md` **1개뿐**임을 재확인했다 — `fix-drafts/`도 나머지도 똑같이 존재하지 않는다.
- 방향도 보수적이다. 잘못된 not-ignore의 비용은 `git status` 소음이고 스스로 드러난다. 잘못된 ignore는 상태를 조용히 감춘다. 이 SPEC 자신의 독트린(침묵은 증거가 아니다)과 일관된다.
- `capability-map.md`가 **입력**으로 읽히는 것도 재확인했다(`navigator_enrich.go:75`가 그 경로를 기본값으로 열고 출력은 `.moai/project/codemaps`). 디렉터리 통째 ignore를 막는 판정은 옳다.

**남는 관찰 하나(결함 아님)**: REQ-ECC-012는 아직 존재하지 않는 파일들에 대해 앞으로 구속하는 규칙이므로, 구조적으로는 §4.3이 반대하는 "예측"의 거울상이다. §4.2의 불활성/활성 구분이 이에 답하지만 §4.3이 그것을 다시 적지는 않는다. 한 문장이면 닫힌다.

**정리: SPEC의 반박은 옳고, iter1 D9는 철회한다.** SPEC이 감사에 동조하지 않고 측정으로 반박한 것이 이 판의 가장 좋은 대목이다.

---

## 수리가 새로 들여온 결함

### N1 — §1.1의 증거 기반은 추적되고 있지 않고, 추적되는 순간 표의 모든 값이 움직인다 [BLOCKING]

- **파일**: `spec.md` §1.1, §5 / `plan.md` §A, §D
- **주장**: "아래 수치는 전부 **커밋된** 스크립트 2개가 만든다."
- **실측**: `.moai/reports/t375`와 `.moai/specs/SPEC-EVIDENCE-CITATION-CANON-001`의 추적 파일 목록 → **출력 없음**. 작업 트리 상태에서도 두 디렉터리 모두 `??`. 스크립트도 SPEC도 **추적되지 않는다**. 현재형 단언이 트리와 어긋난다.
- **더 무거운 절반**: §1.1은 tracked-set grep을 고른 이유를 "`grep -r`은 미추적 파일까지 훑으므로 **이 SPEC 자신의 산출물을 세고**, 카드가 진행될수록 값이 움직인다"라고 적는다. 그 논거는 SPEC이 미추적인 동안만 성립한다. plan 단계가 닫히며 SPEC 3파일이 추적 집합에 들어가면 tracked-set grep이 그것을 세기 시작한다. 실측했다:

  ```
  # SPEC 3파일(전부 이 문자열을 담는다)을 모집단에 넣고 같은 스크립트 실행
  184 → 187 | 515 → 539 | 346 → 351 | 124 → 127 | 231 → 234
  ```

  §1.1 표의 다섯 값과 §5의 부채 수치(124 / 231 / 189)가 **plan-close 커밋 하나로 전부 스테일해진다.** 이는 §1.1.1이 고백한 결함과 같은 계열이다 — 값이 읽히는 트리에서 그 값을 만들지 못한다 — 다만 기제가 명령 오기가 아니라 **시간**이다.
- **왜 blocking인가**: §1.1의 재귀속이 이번 판의 표제 수리이고, 그것이 이 SPEC 자신의 정상적인 다음 동작을 견디지 못한다. 귀속 무결성을 입법하는 문서에서 두 번째로 같은 자리다.
- **Severity**: major · **Class**: blocking
- **Required fix**: (a) 스크립트를 실제로 커밋해 §1.1의 "커밋된"이 참이 되게 한다. (b) 모집단에서 이 SPEC 자신의 디렉터리를 배제해 값이 커밋 경계를 넘어 안정되게 한다(pathspec 제외 `:!.moai/specs/SPEC-EVIDENCE-CITATION-CANON-001/*`, 또는 스크립트 내 필터 한 줄) — 그러면 §1.1이 이미 세운 논거가 끝까지 성립한다. (c) 또는 표의 값들을 HEAD `b64043481`에 고정된 것으로 명시하고 커밋 후 +3 이동을 함께 적는다.

### N2 — "두 출력은 갈라질 수 없다"는 주장을 코드가 뒷받침하지 않는다

- **파일**: `.moai/reports/t375/measure_resolution.py` docstring
- **주장**: "Shares the `concrete` predicate with measure_citations.py … so the two outputs **cannot** drift apart."
- **실측**: 공유가 아니라 **복제**다. 두 파일이 각자 `PATH_RE`·`CONCRETE_RE`를 리터럴로 정의하고(citations.py:26-27, resolution.py:24-25), import 관계가 없다. 오늘 문자열이 동일한 것은 사실이나, 갈라지는 것을 막는 기제는 없다. 참인 문장은 "지금 동일하다"이지 "갈라질 수 없다"가 아니다.
- **Severity**: minor · **Class**: optional (한쪽이 다른 쪽에서 import하면 주장이 참이 된다)

### N3 — AC-ECC-010의 두 하한 도출 명령이 서로 다른 범위를 잰다

- **파일**: `acceptance.md` AC-ECC-010 / `plan.md` §D.2
- **실측**: 루트 명령은 doctrine 하위트리 **4개를 열거**해 363을 낸다. 미러 명령은 `internal/template/templates/.claude` **전체**를 훑어 340을 낸다. 4개 하위트리만 세면 미러는 **338**이고, 차이 2는 범위 밖 파일이다:

  ```
  internal/template/templates/.claude/loop.md
  internal/template/templates/.claude/commands/moai/todo.md
  ```

  §D.1이 정한 스캔 범위는 양쪽 트리 모두 4개 하위트리인데, 미러 하한은 그보다 넓은 모집단에서 도출됐다. 하한 결정은 바뀌지 않지만(338도 300 초과), 라벨과 명령의 모집단이 어긋난다 — 이 SPEC이 입법하는 바로 그 단위 정합 문제다.
- **Severity**: minor · **Class**: optional

### N4 — 532 → 515 전환에 다리가 없다

- **파일**: `spec.md` §1.1 표 2행 vs §1.1.1
- **실측**: 맨 문자열 출현 → **532**. 뒤에 `/`가 오는 출현 → **515**. 차이 17은 `/`가 뒤따르지 않는 출현이다.
- **불일치**: §1.1.1은 "184·**532**는 추적 집합에 대해 옳고"라고 적어 532를 옳은 값으로 확인해 주는데, §1.1 표의 같은 라벨("경로 출현") 행은 이제 **515**를 싣는다. 단위가 좁아진 것(맨 문자열 → `/`가 뒤따르는 문자열)인데 그 사실이 적혀 있지 않아, 두 절을 나란히 읽으면 모순으로 보인다.
- **Severity**: minor · **Class**: optional

### N5 — REQ-ECC-013이 번호 순서 밖에 정의됨

`§2.2` 끝에 013이 놓여 정의 순서가 001…010, **013**, 011, 012다. 집합은 완전하므로 MP-1 위반이 아니고 매핑도 정확하다. 읽는 순서만의 문제다. **Severity**: minor · **Class**: optional

---

## Regression Check (iter1 결함의 잔존 여부)

iter1 13건(D14 확인 항목 제외) 중 **11건 종결, 2건 부분 종결**. 전 항목에서 **미변경 잔존은 0건** — 정체(stagnation) 신호 없음. 부분 종결 2건(D1·D7)은 방향이 옳고 폭이 줄었으며, 남은 것은 각각 한 줄 수리다.

---

## Recommendation

**PASS-WITH-DEBT (0.85).** must-pass 7항목 전부 통과 또는 N/A, D7·D8 BLOCKING 없음, 총점이 Tier M 임계 0.80을 상회한다. iter1 대비 +0.15 단조 상승. Tier M 재감사 상한(2회)에 도달했으므로 이것이 종결 판정이다.

**부채로 넘기는 항목(우선순위 순)** — 넷 다 SPEC 재작성이 아니라 국소 수리다:

1. **N1 (증거 기반)** — 스크립트를 커밋하고, 모집단에서 이 SPEC 자신의 디렉터리를 배제한다. 배제 한 줄이면 §1.1의 값이 plan-close를 견딘다. **run 단계 진입 전에 닫는 것을 권한다** — 닫지 않으면 M1이 시작되는 순간 §1.1과 §5가 스테일해진다.
2. **D1 잔여 (가드)** — AC-ECC-015의 "방문 루트 목록 단언" 패턴을 하위트리 4개로 내린다. 지금은 `.claude/agents/`나 `.claude/output-styles/`가 범위에서 빠져도 초록이고, 그 둘이 이 SPEC의 최강 사례를 담고 있다.
3. **D7 잔여 (허용목록)** — 단언 #1(파일 단독 항목 부재)에 판정 명령을 붙이거나, AC-ECC-009처럼 대체 판독법을 밝힌다.
4. **N2·N3·N4·N5** — optional. 고쳐도 좋고 넘겨도 판정에 영향이 없다.

**iter1 D9는 철회한다.** SPEC의 반박이 옳다 — `fix-drafts/`를 처분하는 코드가 없어 그 잔존은 `chain/`과 같은 신호가 아니며, `chain.go:151`의 `os.Remove(src)`가 그 비대칭을 확정한다. 이 감사가 iter1에서 인용한 t373 원문에 이미 그 전제가 들어 있었는데 적용하지 않았다.
