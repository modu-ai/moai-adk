# t877 판정문 — 게이트웨이 400 조사의 교란 요인 재검증 (Gap #4 측정)

카드: t877 (Class B, Tier S) · 브랜치: `WT-gateway-400-recheck` (base develop `90b0a32bc`) · 날짜: 2026-09-18

선행 문서는 이동하는 ref 가 아니라 커밋 한정 경로로 인용한다:
`git show 54fcf2e1b:.moai/reports/t851/verdict.md` (lane-4 receipt-400 RCA. 이 커밋은 `WT-receipt-400-rca` 브랜치에만 있고 develop 에 착지하지 않았다 — §2 근거 E5).

---

## 1. Claim (주장)

1. **2026-09-14 표본 두 건은 Claude Code 2.1.270 에서 발생했다.** 상류 결함 구간(2.1.265 도입 ~ 2.1.268 수리)의 **밖**이다.
2. 따라서 **그 표본에 상류 Artifact input-schema 결함이 교란 요인으로 들어갔을 가능성은 배제된다.** t851 의 단일 원인 전제는 이 축에서는 깨지지 않는다.
3. 이 판정은 카드의 [HARD] 를 지킨다 — 「상류 버그가 우리 400 을 일으켰다」고 주장하지 않으며, 그 반대 방향(무관함)을 **버전 일치 여부 한 가지 근거로만** 말한다.
4. **Gap #4 는 닫힌다**: 표본의 CC 버전은 이제 측정값으로 존재한다. 선행 분석 문서에는 버전 기재가 아예 없었다(§2 E4).
5. 파생: t707·t708·t851 의 판정 중 **이 표본군에 근거한 부분**은 상류 교란 요인 때문에 재측정할 필요가 없다. 다른 표본군에 근거한 판정은 이 카드가 보지 않았다(§4 G2).

---

## 2. Evidence (증거 — 명령과 관측 출력)

**E1. 표본 1 트랜스크립트의 CC 버전**

```
python3 -c "... json.loads(line); d.get('version') ..." \
  ~/.moai/state/gateway-conversations/families/2e343aa0-5a47-4c1f-bc7e-c3046531713b/native/projects/-Users-goos-MoAI-moai-adk-go--claude-worktrees-t850/2e343aa0-5a47-4c1f-bc7e-c3046531713b.jsonl
→ lines 168
→ versions {'2.1.270': 105}
→ first_ts 2026-09-13T17:26:47.288Z  last_ts 2026-09-13T17:30:30.780Z
```

**E2. 표본 2 트랜스크립트의 CC 버전**

```
→ {'2.1.270': 68}  2026-09-13T17:33:24.448Z  2026-09-13T18:14:04.353Z
```

**E3. 같은 가족의 형제 트랜스크립트 2건(교차 확인)**

```
2e343aa0…/native/projects/-Users-goos-MoAI-moai-adk-go/5558bdaa-….jsonl → {'2.1.270': 21}
e2f5ba68…/native/projects/-Users-goos-MoAI-moai-adk-go/ccda9de6-….jsonl → {'2.1.270': 23}
```

네 파일 모두 `version` 값이 `2.1.270` 단일이고 다른 값이 섞이지 않았다.

**E4. 상류 결함 구간 — 표본 세션이 캐시한 CC changelog 원문**

`~/.moai/state/gateway-conversations/families/2e343aa0…/native/cache/changelog.md:117` (`## 2.1.268` 절 안):

```
- Fixed every turn failing with HTTP 400 on third-party Anthropic-compatible endpoints
  (`ANTHROPIC_BASE_URL`) since 2.1.265: a regex in the Artifact tool's input schema that
  those endpoints reject
```

도입 2.1.265, 수리 2.1.268 — 카드 서술과 일치하며, 이 문장은 **카드가 아니라 원문에서** 읽었다.

**E5. 선행 판정문의 소재**

```
git merge-base --is-ancestor 54fcf2e1b HEAD → false
git branch --contains 54fcf2e1b → WT-receipt-400-rca (develop 아님)
ls .moai/reports/t851/ → No such file or directory (이 워크트리·primary 양쪽)
```

**E6. 2026-09-14 분석 문서와 표본의 연결**

```
grep -o '2e343aa0…|e2f5ba68…|2\.1\.2[0-9][0-9]' \
  /Users/goos/MoAI/moai-adk-go/.moai/reports/gateway-400-analysis-20260914.html
→ 2e343aa0-5a47-4c1f-bc7e-c3046531713b (3건), 버전 문자열 0건
```

그 문서는 이 표본을 인용하면서 **CC 버전을 한 번도 적지 않았다** — Gap #4 가 실재했다는 직접 근거다.

---

## 3. Baseline-attribution (이번 런 귀속)

- 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t877`, 브랜치 `WT-gateway-400-recheck`, HEAD `90b0a32bc` (원격 착지 develop).
- 표본 아티팩트: `~/.moai/state/gateway-conversations/families/{2e343aa0-5a47-4c1f-bc7e-c3046531713b, e2f5ba68-68b4-4c57-b24b-9fab37d5ff14}/` — lane-9 계측 산출물, 디렉터리 mtime 2026-09-14 02:26 / 02:33 (KST). 트랜스크립트 타임스탬프(UTC 09-13 17:26~18:14)와 같은 시각이다.
- 모든 수치는 이번 런에서 위 명령으로 읽은 값이다. **카드 본문의 수치·행번호는 인용하지 않았다.**

---

## 4. Gaps (미검증)

- **G1 — 버전 필드의 의미.** `version` 은 트랜스크립트를 기록한 CC 가 자기 버전으로 적은 값이다. 요청을 실제로 보낸 바이너리가 같은 빌드라는 것은 합리적 추정이지만 별도로 계측하지 않았다.
- **G2 — 다른 표본군.** t707·t708 이 이 두 가족 밖의 표본에 근거한 부분이 있다면 그 표본의 버전은 재지 않았다. 이 카드는 t851 이 명시한 라이브 표본 2건만 측정했다.
- **G3 — 상류 결함의 관측적 부재.** 「2.1.270 이므로 그 결함이 없었다」는 changelog 의 수리 선언에 기댄 추론이다. 표본 트랜스크립트에서 Artifact input-schema 400 의 흔적을 직접 찾지는 않았다.
- **G4 — 코드 변경 없음.** 이 카드는 조사이며 `internal/` 을 수정하지 않았다. 따라서 테스트·vet·lint 를 돌리지 않았다 — 판정 대상이 없기 때문이며, 측정 실패가 아니다.
- **G5 — 파일 부재.** 카드가 지목한 `.moai/reports/gateway-400-analysis-20260914.md` 는 존재하지 않는다. 존재하는 것은 같은 이름의 `.html`(미추적, 다른 카드 산출물)뿐이다. 리드 판정에 따라 이 판정서 한 곳에만 결론을 쓰고 그 파일들은 건드리지 않았다.

---

## 5. Residual-risk (잔여 위험)

- **R1.** 표본 두 건이 같은 세션 시각대(약 48분)에서 나왔다. 버전이 같은 것은 당연하므로, 이 측정은 「그 시각대의 CC 가 2.1.270 이었다」를 말할 뿐 「모든 400 관측이 2.1.270 이었다」를 말하지 않는다.
- **R2.** 상류 결함이 매 턴 400 을 냈다는 서술이 맞다면, 2.1.265~267 구간에서 게이트웨이를 쓴 세션은 조사 자체가 불가능했을 것이다. 즉 교란 요인이 실제로 섞이기 어려운 구조인데, 이 역방향 논증은 changelog 서술에만 기댄다.
- **R3.** t851 의 미착지 상태는 이 카드가 해결하지 않았다. 그 판정문이 develop 에 없으므로, 다음 사람이 `WT-receipt-400-rca` 브랜치를 모른 채 찾으면 부재로 읽는다 — 카드 발행 후보(이 카드 범위 밖).

---

## 6. 결론 요약 (한 줄)

2026-09-14 표본은 CC **2.1.270** — 상류 결함 구간(2.1.265~2.1.267) 밖이므로, 그 결함은 이 조사의 교란 요인이 **아니다**. t851 의 원인 판정은 이 축에서 그대로 선다.
