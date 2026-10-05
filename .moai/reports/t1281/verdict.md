# 카드 t1281 판정서 — SPEC-STATUSLINE-LANDED-LABEL-001

상태줄 TODO 세그먼트의 「착지」 표기를 고친 카드다. 집계 기준을 `moai todo auto-done` 과 같은 주제줄 귀속 판정으로 바꾸고, 표기 기호를 `✓N` 에서 `⚑N` 으로 바꿨다.

- 브랜치: `WT-statusline-landed-label`
- run 커밋: `7803fa4c0`(구현), `dece05e34`(run_commit_sha 기입)
- 원본 증거: `.moai/specs/SPEC-STATUSLINE-LANDED-LABEL-001/progress.md` §E.2, 원출력 `.moai/state/verify/t1281/`(로컬, 미추적)

## 1. Claim (주장)

1. 착지 집계는 이제 카드 id 가 **커밋 주제줄에 귀속된** 착지 커밋만 센다. 판정식은 `moai todo auto-done` 과 같은 kanban 패키지 기계(`LandedScanArgs` / `LandedAttributions` / `AutoDoneSubjectFresh`)이며, 카드의 `added_at` 보다 오래된 커밋은 세지 않는다(`added_at` 이 비어 있으면 세지 않는 쪽으로 닫힌다). 카드 수와 무관하게 `git log` 는 한 번만 호출한다.
2. 표기 기호는 `⚑`(U+2691)이다. 뜻은 「착지 커밋이 있다 — done 처리 전에 확인하라」이며 「닫아도 된다」가 아니다.
3. 옛 기준으로 쓰인 캐시는 「모름」으로 취급해 표기를 내지 않고, 곧바로 갱신 대상이 된다. 첫 갱신 전까지는 표기가 비어 있다.
4. 「착지는 picked 수에서 빼지 않는다」는 규칙은 **유지**한다(§1.2).
5. 수용 기준 18행(AC ID 12개 AC-SLL-001..012, b/c 하위 기준 포함)이 모두 통과했다.

### 1.1 리드 결정 ③

리드가 Jev 판정(0.60)을 근거로 결정 ③ 을 택했다 — **집계 기준과 기호를 둘 다 바꾼다.** 기준만 고치면 주제줄 귀속 수도 여전히 「완료」가 아니라는 점(측정된 4건이 모두 plan 단계만 착지)이 `✓` 에 가려지고, 기호만 고치면 과대 집계가 남는다. 이 카드는 그 결정을 다시 열지 않았다.

### 1.2 「착지는 picked 에서 빼지 않는다」 유지 판단

유지한다. 카드는 `auto-done` 이나 `done` 이 닫을 때까지 picked 로 남는 것이 큐의 정의이고, 착지 커밋은 plan 단계만 착지한 경우일 수 있다. 착지를 picked 에서 빼면 상태줄이 아직 닫히지 않은 카드를 끝난 것처럼 보이게 만들어, 바로 이 카드가 없애려는 오독을 되살린다. `TestRenderer_LandedNeverSubtracts` 가 이 규칙을 고정한다.

### 1.3 기호 후보

| 후보 | 코드포인트 | 결과 | 이유 |
|---|---|---|---|
| **⚑** | U+2691 BLACK FLAG | **채택** | 「주의 표시」로 읽히며 완료의 뜻이 없다. 한 칸 폭(East Asian Width N)이고 Emoji 속성이 없어 터미널이 넓히지 않는다. |
| ⇡ | U+21E1 UPWARDS DASHED ARROW | 기각 | 화살표는 상태가 아니라 「수치 상승 추세」로 읽힌다. |
| ⚐ | U+2690 WHITE FLAG | 기각 | 속이 빈 윤곽이라 작은 글꼴에서 잘 보이지 않고, 「항복·비어 있음」으로 읽힐 수 있다. |

구조적으로 제외한 부류:

- 체크 표시(✓ ✔ ☑) — 「완료」로 읽힌다.
- ⚠ U+26A0, ⤴ U+2934 — Emoji 속성이 있어 터미널마다 폭이 달라진다.
- ◆ U+25C6, ⊙ U+2299 — East Asian Width 가 A(모호)라 CJK 설정 터미널에서 두 칸이 된다.

## 2. Evidence (증거)

아래는 모두 `progress.md` §E.2 에 run 단계가 남긴 원출력의 전사다.

RED — 구현 변경 전, 옛 `countNamed` / `--format=%B` 코드에서 행위 실패:

```
$ go test ./internal/statusline/ -run '^TestRefreshLandedCounts_NonAttributingMentionDoesNotCount$' -count=1
--- FAIL: TestRefreshLandedCounts_NonAttributingMentionDoesNotCount (0.09s)
    landed_test.go:573: non-attributing mention: {Landed:1 Ref:origin/main Measured:true FetchedAt:1790502302 Available:true}, want measured landed=0 (old criterion would give 1)
FAIL
```

GREEN 과 게이트:

```
$ go test ./internal/statusline/... ./internal/kanban/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/statusline	20.428s
ok  	github.com/modu-ai/moai-adk/internal/kanban	185.537s

$ golangci-lint version
golangci-lint has version v2.1.6 built with go1.26.8 ...
$ golangci-lint run ./internal/statusline/...
0 issues.

$ go vet ./internal/statusline/...
(no output, exit 0)
```

문서·잔재 grep:

- `grep -c '✓'` docs-site statusline 페이지 → ko 0 / en 0 / ja 0 / zh 0
- `grep -c '⚑N'` 같은 페이지 → ko 3 / en 2 / ja 3 / zh 3
- `grep -c -- '--format=%B' internal/statusline/landed.go` → 0

커버리지: `internal/statusline` 문장 기준 90.7% (run 단계가 `7803fa4c0` 에서 측정). sync-audit 재측정은 90.4% (`f4ca82858` 에서 측정).

## 3. Baseline-attribution (기준선 귀속)

- 변경 전 기준선(레인 측정, 이 세션): 상태줄 캐시 `landed=9`(본문 언급 기준) 대 `moai todo auto-done --dry-run` 의 `closed=4 scanned=29`, ref `origin/develop`. 이 차이가 과대 집계의 근거다. 출처는 `spec.md` §A 와 `plan.md` §B 에 기록된 레인 측정이다.
- 새 집계의 정답 기준은 위 `closed=4` 가 아니라 AC 픽스처다. 상태줄 집계는 증거 형식 1(기록된 SHA 도달성)과 M1/M2 가드를 쓰지 않으므로 `auto-done` 의 `closed=N` 과 같을 필요가 없다(`spec.md` §D).
- 테스트·lint 출력은 run 단계가 `7803fa4c0` 트리에서 잰 것이다. `dece05e34` 는 `progress.md` 한 파일만 바꾼 커밋이다.

## 4. Gaps (미검증)

- sync 단계는 테스트·lint 를 다시 돌리지 않았다. §2 의 출력은 run 단계 원출력의 전사다.
- 이 트리에서 새 기준의 실제 상태줄 수치(`⚑N` 의 N)는 재지 않았다. 첫 갱신 이후의 실 캐시 값은 관측되지 않았다.
- darwin 외 플랫폼(linux/windows) 빌드는 로컬에서 재지 않았다. CI 매트릭스가 판정한다.
- `maybeRefreshLandedCounts` 의 프로세스 생성 경로는 `go test` 에서 막혀 있어 66.7% 에 머문다(기존 상태).

## 5. Residual-risk (잔여 위험)

- 재발급된 카드 id 의 이전 세대가 새 카드의 `added_at` 이후에 주제줄 귀속 커밋을 착지시키면 세대 경계로 구별되지 않아 한 건 과대 집계된다. 드물고, 「확인 후 done」 기호가 흡수한다고 보고 수용했다.
- 주제줄 귀속 착지도 plan 단계만일 수 있다. `⚑` 은 확인을 요구할 뿐이며 운영자가 여전히 `✓` 처럼 읽을 가능성은 문서로만 막는다.
- 옛 기준 캐시가 남은 환경에서는 첫 갱신 전까지 표기가 사라진다. 결함이 아니라 설계이지만 사용자가 「기능이 없어졌다」로 오해할 수 있다.
