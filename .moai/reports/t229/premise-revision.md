# t229 — 전제 정정과 현재 트리 실측

`cause.md` 는 조사 시점에 두 번 스테일했다. 이 문서가 그 정정과, **현재 main 에서 실제로 남아 있는 결함**의 실측 기록이다.

| 항목 | 값 |
|---|---|
| 측정 트리 | `294b4b6ab` (= `origin/main`, worktree base) |
| 측정 일자 | 2026-08-24 |
| 측정 방법 | `internal/cli` 에 임시 테스트를 넣어 `synthesizeReviewOutput` 직접 호출 후 삭제 |

---

## 1. 라이브 프로브가 `pass` 를 받은 이유는 코드 결함이 아니라 **바이너리 랙**이다

MCP 서버를 띄우는 것은 설치 바이너리 `~/go/bin/moai` 다. 그 바이너리가 어느 커밋에서 빌드됐는지 직접 물었다.

```
$ ~/go/bin/moai version
 v3.1.2   a1b1ca696   built 2026-08-24T11:19:07Z
```

그리고 그 커밋이 관련 수정 2건을 담고 있는지 기계로 판정했다.

```
$ git merge-base --is-ancestor f505955a9 a1b1ca696 ; echo $?   # t178
1
$ git merge-base --is-ancestor 4505df411 a1b1ca696 ; echo $?   # t186
1
$ git merge-base --is-ancestor f505955a9 HEAD ; echo $?        # 현재 트리
0
```

- `f505955a9` — `fix(t178): stop two audit backends verdicting code they never read` (2026-08-23 16:59)
- `4505df411` — `fix(t186): read a stated inconclusive verdict instead of synthesizing pass` (2026-08-23 18:03)

**둘 다 설치 바이너리의 조상이 아니다.** 즉 프로브는 t178·t186 **이전** 코드를 측정했다. t197 의 감사들도 같은 바이너리를 거쳤다.

바이너리가 얼마나 뒤처졌는지:

```
$ git rev-list --count a1b1ca696..origin/main
259
```

**259 커밋.** 이것은 메모리에 이미 적힌 위험(`primary 워킹트리는 기준이 아니다`, t245)이 같은 라운드에서 두 번째로 실현된 것이다. 리드 primary 체크아웃이 origin/main 보다 259 커밋 뒤처져 있고, 설치 바이너리가 거기서 빌드됐다.

> 참고 — 이 절의 근거는 **바이너리의 커밋 스탬프**다. 별도로 제기된 "장수 `moai mcp-server` 프로세스가 구 코드를 붙들고 있다"는 설명은 내가 재현하지 못했다: `ps` 에 `moai mcp-server` 프로세스가 한 건도 없고, 비슷한 이름으로 보이는 것들(`moai-mcp-imweb` / `moai-mcp-smartstore` / `moai-mcp-cafe24` / `moai-mcp-threads-poster`)은 Claude 데스크톱 플러그인으로 **이 저장소와 무관**하다. 프로세스 랙 가설은 채택하지 않는다.

### 이에 따라 철회하는 주장

| cause.md | 상태 |
|---|---|
| F1 "codex 는 구조화된 verdict 를 반환하지 않는다" | **유지** (참) |
| F2 "합성 규칙은 정규식 1개" | **정정** — 현재 트리는 정규식 **2개**다. `codexStatedVerdict`(`mcp_codex.go:1130`)가 본문 명시 verdict 를 읽는다 |
| F5 "본문에 verdict 단어가 있는데 어댑터가 버린다" | **철회** — 현재 트리는 `Verdict: <word>` 라벨 형태를 읽는다 |
| F4 "라이브 재현" | **관측은 유효하나 대상이 다르다** — 구 바이너리(a1b1ca696)를 측정한 것이지 현재 main 이 아니다 |

---

## 2. 그런데 결함은 현재 main 에 **살아 있다** — 실측

임시 테스트로 `synthesizeReviewOutput` 을 직접 호출해 관측했다.

| 입력 | `statedMatch` | `bulletMatch` | **합성된 verdict** | 판정 |
|---|---|---|---|---|
| 라이브 프로브 본문 (`Verdict: inconclusive` 명시) | true | false | `inconclusive` | 정상 — t178/t186 이 닫음 |
| **점수 표기 `FAIL 0.75 / 1.00` + 차단 2건** | false | false | **`pass`** | **결함** |
| 점수 표기 `PASS 0.88 / 1.00` | false | false | `pass` | 우연히 맞음 |
| **미인식 서식 — `Blocking` 행 표 + `merge_status: blocked`** | false | false | **`pass`** | **결함** |
| 미인식 서식 — 산문 1줄 | false | false | `pass` | adversarial 에서는 `inconclusive` 여야 함 |
| native 정상 리뷰 (무불릿) | false | false | `pass` | 정상 — **보존 대상** |
| review-mode 불릿 `- [P1]` | false | true | `fail` | 정상 |

**핵심**: t197 기록에 나타난 바로 그 서식(`FAIL 0.75 · 차단 N건`)이 현재 main 에서도 `pass` 로 합성된다. 카드의 핵심 결함은 좁아졌을 뿐 사라지지 않았다.

두 결함 행이 공유하는 뿌리는 하나다 — **아는 서식이 안 맞으면 `pass`** (`verdict := "pass"`, `mcp_codex.go:1145`). t178·t186 은 서식을 **하나 더 알게** 했을 뿐, 모르는 서식이 통과로 떨어지는 구조는 그대로다. 그래서 교정의 형태는 "정규식을 하나 더 추가한다"가 아니라 **모르면 `inconclusive`** 여야 한다.

---

## 3. 남은 결함 정리 (현재 트리 기준)

| # | 결함 | 실측 근거 | 성격 |
|---|---|---|---|
| G1 | 점수 표기(`FAIL 0.75`)를 명시 verdict 로 인식하지 못함 | 위 표 2행 | **live defect** |
| G3 | 아는 서식이 하나도 안 맞으면 `pass` (adversarial 포함) | 위 표 4·5행 | **live defect — 구조적 원인** |
| G4 | 두 신호가 갈렸다는 사실이 결과 어디에도 기록되지 않음 | `converge()` 가 `Summary` 를 판정에 안 씀 (`mcp_convergence.go:135`) | live gap |
| G2 | 보수 채택이 명시적 순서 테이블이 아니라 대입 순서로 구현됨 | `mcp_codex.go:1144-1156` | **결함 아님** — 현재 조합은 이미 fail 편향으로 올바르게 동작한다(위 표에 반례 없음). 유지보수성 사안으로만 다룰 것 |

G2 를 "결함" 으로 적으면 관측하지 않은 결함 주장이 된다. 실측에서 잘못된 값을 낸 조합이 없었으므로 **구현 형태에 대한 지적**으로만 남긴다.

---

## 4. 파생 관측 — 이 카드 범위 밖, 파급은 더 넓다

설치 바이너리가 259 커밋 뒤처진 상태로 이 팩토리 라운드의 **모든 audit 호출**을 서빙했다. t197 의 불일치 3회도, 내 라이브 프로브도 그 바이너리를 통과했다. 즉:

- 이 라운드에서 나온 audit 판정 중 **어느 것이 현재 코드를 측정한 것인지 아무도 모른다**.
- 그리고 그 사실이 판정 출력 어디에도 드러나지 않는다 — 바이너리 커밋을 따로 묻기 전까지는.

카드 후보: **audit 결과에 그것을 산출한 바이너리의 커밋을 함께 기록한다.** 대표 mutant — 버전 문자열(`v3.1.2`)만 적고 커밋을 안 적는 구현. `v3.1.2` 는 259 커밋 뒤처진 빌드와 최신 빌드를 구분하지 못한다.
