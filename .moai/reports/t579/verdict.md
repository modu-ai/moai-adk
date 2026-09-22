# t579 판정 — 빈 리뷰 판별식의 U+200B 우회 수리

- 카드: t579 (Class B, SPEC 없음, cycle_type=tdd) · 리드 발행 2026-09-08 · t551 감사 실측
- 트리: `.claude/worktrees/t579`, 브랜치 `WT-zerowidth-blank-review`, base `e6b1c3a1c` (로컬 develop)
- 측정 일자: 2026-09-10
- 선행 카드: t551 (`SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001`, status completed) — 같은 fail-open 계열의 잔여 축
- 커밋 구성: 테스트와 수리 전 RED 로그를 먼저 커밋하고(이 커밋의 `mcp_codex.go` 는 base 그대로다), 수리와 수리 후 로그·이 판정서를 그다음 커밋에 담았다. 테스트가 수리보다 먼저 존재했다는 것은 커밋 그래프로 확인된다.

## 주장 (Claim)

1. 코덱스 리뷰 경로의 유일한 공백 판별식 `codexReviewTextIsBlank` 는 수리 전 `strings.TrimSpace(s) == ""` 였고, U+200B ZERO WIDTH SPACE 만 담긴 본문을 **내용으로 읽었다.** 수리 전 운영 RPC 경로에서 그 본문은 **실제로 verdict=pass 로 합성됐다**(아래 §7, end-to-end 관측). 내용 없는 리뷰가 "문제 없음"으로 보고되는 fail-open 이다.
2. 수리 후 판별식은 문자를 손으로 나열하지 않고 **범주**(`unicode.IsSpace` 또는 일반 범주 `Cf`)로 판정한다. 제로폭 문자만 담긴 본문은 빈 본문으로 판정되고, end-to-end 로 inconclusive 에 떨어진다.
3. 수리는 과잉 수리가 아니다. 실제 내용에 제로폭 문자가 섞인 리뷰(앞·뒤·중간)는 여전히 내용으로 판정되고, 제로폭이 없는 평범한 리뷰의 판정도 변하지 않는다. 이 두 묶음은 수리 전후 모두 같은 결과를 냈다.
4. 새 테스트는 **제로폭 분기만** 붙잡는다. 수리 전 코드에서 떨어진 칸은 제로폭 4칸과 그 end-to-end 한 건뿐이고, 나머지 대조 칸과 기존 테스트는 그대로 통과했다.
5. t551·t580 착지본이 이 축을 이미 닫았을 가능성은 배제했다 — develop 의 Go 코드 전체에 제로폭 처리가 없었다.

## 증거 (Evidence)

### 1. 문자 성질 — 패키지 밖 독립 프로그램

```
U+0020 IsSpace=true  Cf=false Zs=true  TrimSpaceEmpty=true
U+00A0 IsSpace=true  Cf=false Zs=true  TrimSpaceEmpty=true
U+200B IsSpace=false Cf=true  Zs=false TrimSpaceEmpty=false
U+200C IsSpace=false Cf=true  Zs=false TrimSpaceEmpty=false
U+200D IsSpace=false Cf=true  Zs=false TrimSpaceEmpty=false
U+2060 IsSpace=false Cf=true  Zs=false TrimSpaceEmpty=false
U+FEFF IsSpace=false Cf=true  Zs=false TrimSpaceEmpty=false
```

U+200B 는 `unicode.IsSpace` 가 아니므로 `strings.TrimSpace` 가 깎지 못한다. 카드의 전제가 그대로 확인된다.

### 2. 선행 처리 부재 확인

`grep -rn "200B\|200b\|ZeroWidth\|zero-width\|zerowidth\|FEFF\|feff" --include='*.go' internal/` → 0건. 제로폭 처리는 이 트리 어디에도 없었다.

### 3. 범주 `Cf` 의 범위 실사

범주로 넓히는 선택은 나열보다 낫지만, 범주가 무엇을 담는지 모른 채 넓히면 그것도 미관측 주장이 된다. 그래서 `unicode.Cf` 를 전수로 셌다: **총 170 코드포인트.** 앞 40개:

```
U+00AD U+0600 U+0601 U+0602 U+0603 U+0604 U+0605 U+061C U+06DD U+070F U+0890 U+0891 U+08E2 U+180E
U+200B U+200C U+200D U+200E U+200F U+202A U+202B U+202C U+202D U+202E
U+2060 U+2061 U+2062 U+2063 U+2064 U+2066 U+2067 U+2068 U+2069 U+206A U+206B U+206C U+206D U+206E U+206F U+FEFF
```

계열: soft hyphen(U+00AD) · 아랍 숫자 표지(U+0600–0605, U+06DD, U+070F, U+0890/0891, U+08E2) · 몽골 모음 분리자(U+180E) · 제로폭군(U+200B–200D, U+2060) · 방향 제어(U+200E/200F, U+202A–202E, U+2066–2069) · 불가시 연산자(U+2061–2064) · 폐용 포맷(U+206A–206F) · BOM(U+FEFF) · 그 밖에 보조 평면의 태그 문자.

**판정**: 이 판별식이 묻는 것은 "본문에 리뷰할 내용이 있는가" 뿐이고, 위 어느 계열도 혼자서는 내용을 이루지 못한다(아랍 숫자 표지는 뒤따르는 숫자를 수식하는 접두 표지이고, 방향 제어는 배치만 바꾼다). 그러므로 양끝에서 깎이는 것이 옳다.

**과잉 수리 우려를 구조적으로 닫는 성질**: `strings.TrimFunc` 은 양끝만 깎는다. 내용 안쪽의 `Cf` 문자는 건드리지 않으므로, 판별식은 본문을 고치거나 거절하지 않고 "비었는가"만 답한다. 정상 리뷰에 제로폭이 섞여 들어와도 그 본문은 내용으로 판정되고 원문 그대로 흐른다. 이 논거는 측정에 기대지 않지만, 측정(아래 (iv) 칸)으로도 확인했다.

### 4. 수리 내용 (`internal/cli/mcp_codex.go`)

```go
func codexReviewTextIsBlank(s string) bool {
	return strings.TrimFunc(s, isReviewBlankFiller) == ""
}

func isReviewBlankFiller(r rune) bool {
	return unicode.IsSpace(r) || unicode.Is(unicode.Cf, r)
}
```

- 함수 이름·시그니처와 네 호출 지점(가드·수집·선택)은 그대로다. t551 이 "한 곳을 빠뜨리는 위험" 때문에 판별식을 한 함수로 모아둔 구조라, 이 한 함수만 고치면 네 지점이 함께 움직인다.
- 문서 주석에 범주 판정의 근거와 "양끝만 깎는다"는 성질을 적었다.

### 5. 소스 위생 — 투명 문자 잔존 0

저작 단계에서 `\u200b` 이스케이프가 실제 투명 문자로 들어가는 사고가 한 번 있었고 gofmt 가 `illegal byte order mark` 로 잡아 고쳐졌다. 보고를 믿지 않고 직접 셌다:

```
$ perl -CSD -ne '$n += () = /\p{Cf}/g; END { print ... }' <file>
mcp_codex.go literal_Cf=0
codex_blank_review_test.go literal_Cf=0
control_planted_U+200B detected=1
```

같은 계수기가 심어둔 U+200B 1개를 검출하므로 0 이 진짜다. 투명 문자가 남아 있었다면 테스트가 이름과 다른 입력을 검사했을 것이다.

### 6. 컴파일 ① — 수리 적용 트리 (`compile1-fixed.txt`)

**동시 실행 조건**: `internal/cli` 는 레인 사이에서 직렬화 중이었고, 리드가 스코프 실행 2회를 승인했다. 사전 확인과 컴파일을 한 호출 안의 조건문으로 묶어, lane-6 전체 스위트(`-timeout 60m`)가 아닌 `internal/cli` go 프로세스가 보이면 시작하지 않게 했다.

- 첫 시도(`compile1-attempt1-wait.txt`): 다른 레인의 스코프 컴파일(`go test ./internal/cli -run ^(TestHomeState.*|…)$ -coverpkg=…`)이 보여 `foreign_scoped=1`, **시작하지 않았다.** 이때 래퍼는 exit 0 으로 끝났으나 테스트 로그는 생성되지 않았다 — 래퍼 종료코드는 판정이 아니다.
- 재시도(`ps-before-compile1-cli.txt`): 대기 반복 1회 후 직전 확인에 lane-6 전체 스위트(`go test ./internal/cli/ -count=1 -timeout 60m`) 하나만 보였고 `foreign_scoped=0`. 15:41:17 시작, 15:41:25 종료, **go test exit=0**.

```
$ unset MOAI_SESSION_ID MOAI_PROJECT_ROOT MOAI_WORKTREE_ROOT CLAUDE_PROJECT_DIR && go test ./internal/cli/ -run 'TestCodexBlankReview' -count=1 -v -timeout 900s
상위 테스트 PASS 16 / FAIL 0 · 하위 테스트 PASS 28 / FAIL 0
ok  	github.com/modu-ai/moai-adk/internal/cli	0.924s
```

12칸 표의 하위 테스트 12개가 이름별로 모두 찍혔다(셀렉터가 0건에 맞아 공허하게 통과한 것이 아니다). end-to-end 두 건과 기존 t551 테스트(AC001–AC009, M5, 3상태 대조 행렬)도 전부 통과했다.

### 7. 컴파일 ② — 수리 전 운영 코드 (`compile2-prefix.txt`, 재현 겸 뮤턴트)

재현과 뮤턴트는 같은 증거다 — 둘 다 "수리 안 된 판별식에서 새 테스트가 실패한다"를 보인다. 그래서 판별식 한 줄이 아니라 **base 판본 `mcp_codex.go` 전체**(`git show HEAD:internal/cli/mcp_codex.go`)로 바꿔 끼웠다. 주석·import·보조 함수 차이는 동작이 없으므로, 이쪽이 "수리 전 운영 코드"의 더 정직한 재현이다. 테스트 파일은 그대로 두었다.

**동시 실행 조건**(`ps-before-compile2-cli.txt`): 대기 반복 1회, 직전 확인에 lane-6 전체 스위트 하나만 보였고 `foreign_scoped=0`.

```
swapped-in: 28b436563500a998526358a16e55857840ec069cd6e4d720c9078fb8ffaea1b1  internal/cli/mcp_codex.go
compile2 start 2026-09-10T15:42:46+0900
compile2 go-test exit=1
compile2 end   2026-09-10T15:42:53+0900
상위 테스트 PASS 14 / FAIL 2 · 하위 테스트 PASS 24 / FAIL 4
```

| 칸 | 수리 전 (컴파일 ②) | 수리 후 (컴파일 ①) |
|---|---|---|
| i-empty · ii-whitespace · ii-nbsp | PASS | PASS |
| **iii-zwsp · iii-zwsp-run · iii-zw-mix · iii-zw-plus-ws** | **FAIL** | PASS |
| iv-zw-lead · iv-zw-trail · iv-zw-inner | PASS | PASS |
| v-normal-clean · v-normal-finding | PASS | PASS |
| end-to-end: 제로폭 단독 본문 | **FAIL** | PASS |
| end-to-end: 제로폭 + 실제 리뷰 | PASS | PASS |
| 기존 t551 테스트 전부 | PASS | PASS |

수리 전 코드에서 떨어진 칸의 메시지:

```
cell iii-zwsp: codexReviewTextIsBlank("\u200b") = false, want true
cell iii-zwsp-run: codexReviewTextIsBlank("\u200b\u200b\u200b") = false, want true
cell iii-zw-mix: codexReviewTextIsBlank("\u200b\u200c\u200d\u2060\ufeff") = false, want true
cell iii-zw-plus-ws: codexReviewTextIsBlank(" \u200b\n") = false, want true
verdict = pass for a zero-width-only body "\u200b" — a review that produced no verdict was reported as one that found nothing wrong
verdict = "pass", want "inconclusive" for body "\u200b"
```

**읽는 법**:
- 떨어진 것은 **제로폭 4칸과 그 end-to-end 한 건뿐**이다. (i)(ii)(iv)(v) 와 기존 테스트는 수리 전 코드에서도 통과했다. 새 테스트가 제로폭 분기 하나만 붙잡고 있다는 뜻이다 — 뮤턴트가 전부를 떨어뜨리거나 아무것도 떨어뜨리지 않았다면 이 결론은 서지 않는다.
- end-to-end 한 건이 결함의 실체다. 판별식 단위의 오판이 아니라, **수리 전 운영 RPC 경로가 제로폭 본문에 대해 실제로 `pass` 를 반환했다.**
- (iv) 칸과 "제로폭 + 실제 리뷰" end-to-end 가 수리 전후 모두 통과한 것이 과잉 수리 부재의 측정 근거다.

**원복**: 컴파일 직후 수리본 스냅샷으로 되돌렸고, 래퍼가 찍은 값과 별도로 원복 후 새로 읽은 값이 `31ddc0922a98539b1386358edca59d5da12fb45711b42a924b5e47dffcf078b3` 로 수리본과 바이트 동일하다. `git diff --stat` 도 컴파일 전과 같다(2 files, 98 insertions, 3 deletions). 원복 확인은 세 번째 컴파일 대신 해시로 했다(리드 승인).

## 기준선 귀속 (Baseline-attribution)

- 트리: base `e6b1c3a1c`. 수리 전 측정은 base 판본 `mcp_codex.go`(sha256 `28b436563500…`), 수리 후 측정은 수리본(sha256 `31ddc0922a98…`)에서 했다. 두 측정 모두 같은 테스트 파일을 썼다.
- 테스트 바이너리는 매 실행마다 이 트리에서 새로 컴파일됐다(`go test`). 설치된 moai 바이너리는 판정에 쓰지 않았다.
- 문자 성질·`Cf` 실사는 로컬 Go 툴체인의 `unicode` 표로 잰 값이다.
- 모든 수치는 이 실행에서 관측했다. t551 SPEC 의 기록에서 옮겨온 값이 아니다.

## 미검증 (Gaps)

- **`go vet ./internal/cli/` 는 이 슬롯에서 돌리지 않았다.** 패키지 전체를 타입체크해 직렬화 중인 패키지에 부하를 얹으므로, 리드 승인에 따라 패키지 전체 판정 슬롯으로 미뤘다. `gofmt -l` 은 두 파일에서 출력 없음.
- **`internal/cli` 패키지 전체는 돌리지 않았다.** `-run 'TestCodexBlankReview'` 스코프만 실행했다.
- **병합 트리 재측정은 아직이다.** 통합 창에서 로컬 develop 흡수 후 수행한다.
- **Summary 문자열에 앞의 U+200B 가 남는다는 점은 소스 판독이다.** `synthesizeReviewOutput` 이 `strings.TrimSpace` 만 적용하므로 앞의 제로폭이 Summary 에 남는다고 읽었고, 그래서 end-to-end 테스트는 Summary 를 정확 일치가 아니라 포함으로 검사한다. Summary 의 실제 바이트는 측정하지 않았다.
- **동시 실행 확인은 필터 결과만 반출했다.** 판정 근거는 `internal/cli` 를 컴파일하는 go 프로세스 목록(`*-cli.txt`)이고 그것은 반출했다. 원본 전체 프로세스 목록은 호스트의 모든 명령줄을 담아 자격 증명·경로가 섞일 수 있으므로 저장소에 반출하지 않았다.
- darwin 밖(linux·windows)에서는 측정하지 않았다.

## 잔여 위험 (Residual-risk)

- **Summary 에 투명 문자가 남는다.** 판정(verdict)에는 영향이 없지만, Summary 를 사람이 읽거나 문자열로 비교하는 소비자에게는 보이지 않는 문자가 섞인 채 전달된다. 이 카드의 범위(빈 본문 판정) 밖이다.
- **판별식 밖의 `TrimSpace` 사용처는 바꾸지 않았다.** 코덱스 경로의 공백 판정은 이 한 함수로 모여 있지만, 다른 경로에 같은 모양의 판정이 있는지는 전수로 훑지 않았다.
- **`Cf` 범주는 유니코드 버전에 따라 늘어날 수 있다.** 범주 판정이므로 새로 추가되는 서식 문자도 자동으로 빈 채움으로 분류된다 — 이것이 나열 대신 범주를 고른 이유이며, 방향은 fail-closed(inconclusive)다.
