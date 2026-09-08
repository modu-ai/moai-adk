# t528 — 병합 트리 재측정 (통합 창, 2026-09-08)

창 소유: lane-4 (`moai integration acquire --name lane-4`).

## 흡수

- 흡수 직전 재유도: `git fetch origin develop` → `git rev-parse --short origin/develop` = **`c72dc1baf`**
  (리드가 지명한 대상과 일치. 앞서 잰 `124 7`의 시점 이후 develop 이 더 나갔으므로 재유도했다.)
- `git merge origin/develop` → **CHANGELOG.md 충돌 1건**.
  충돌 양쪽은 서로 다른 카드의 `### Fixed` 항목이었다 — HEAD 쪽 t528,
  develop 쪽 t563 / t544 / t537. **어느 한쪽이 최신인 것이 아니라 둘 다 필요한
  내용**이므로 양쪽을 모두 보존하고 마커만 제거했다.
  해소 검증: 잔존 마커 0, 네 SPEC id 각각 1회 존재.
- 병합 커밋 **`759b08420`**.

## 재측정 — 패키지 테스트 (5요소)

```
$ go test ./internal/spec/... -count=1 -timeout 1200s
ok  github.com/modu-ai/moai-adk/internal/spec  76.773s
EXIT=0 · --- FAIL 0 · panic/timed out 0 · (cached) 0
```

카드 자체 트리(`97da8be8b`)에서는 `ok 71.168s`, 같은 세 계수 0. 판정 동일.

## 재측정 — 도달률 (판별식 B, 동결 앵커 대조 열 동반)

```
DENOMINATOR spec.md read = 815
IN-SECTION declarations = 1167
  accepted by FROZEN baseline anchor = 216
  accepted by LIVE parseSingleACLine = 1085
  rejected by LIVE parseSingleACLine = 82
NEWLY accepted (live yes, baseline no) = 869
POSITIVE-NEEDLE files (live) = 110
BASELINE-NEEDLE files (frozen anchor) = 18
FULLY-BLIND files = 9
```

## 움직인 것 하나 — 분모 807 → 815

**보고 대상이므로 새 값으로 조용히 갈아끼우지 않고 원인을 실측했다.**

흡수가 `spec.md` 8개를 들여왔다:

```
SPEC-CODEMAPS-REFRESH-002        SPEC-CTX-BLIND-DOUBLE-001
SPEC-DOCS-CODEX-WIRING-CALLOUT-001  SPEC-DOCTOR-STAT-SEAM-001
SPEC-JUDGMENT-FIRST-MODE-001     SPEC-SEAM-GREENFIELD-001
SPEC-STATE-ANCHOR-VALIDATE-001   SPEC-TODO-HOME-TEMP-GUARD-001
```

그런데 **절 안 선언 기여는 0**이다. 두 이유를 각각 열어 확인했다:

- 7개는 AC 불릿 줄이 파일 전체에 **0개**다.
- 1개(`SPEC-TODO-HOME-TEMP-GUARD-001`)만 AC 불릿 형태의 줄을 1개 갖는데,
  그 줄은 **선언이 아니라 산문**이다 —
  `163:  - **AC-WTQ-008의 주어는 「비git」이지 「임시」가 아니다.**` — 어떤 AC 를
  *설명하는* 문장이지 AC 를 *선언하는* 줄이 아니다. 게다가 그 파일에는
  `acceptance` 를 포함한 `##` 이상 헤딩이 없어 `findACSectionStart` 가 -1 을
  내므로 절 자체가 열리지 않는다.

따라서 **분모만 움직이고 분자·도달률·대조군·blind 는 전부 불변**이다:
`216 → 1085 / 1167 (18.5% → 93.0%)`, 대조 열 `216 → 216`, 파일 `18 → 110`,
blind `9`. 카드 자체 트리의 판정이 병합 트리에서 그대로 성립한다.

## Gaps

- 도달률은 `.moai/specs/**/spec.md` 만 잰다. 흡수가 가져온 다른 패키지의
  변화는 이 수치가 말하지 않는다 — 패키지 테스트가 그 자리를 덮는다.
- 로컬 전체 스위트는 돌리지 않았다(레포 규율). 전 패키지 판정은 CI 몫이다.
