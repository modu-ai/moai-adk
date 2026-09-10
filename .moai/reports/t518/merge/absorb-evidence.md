# 병합 창 증거 — 카드 t518 (lane-4)

## 창

```
moai integration acquire --name lane-4   → acquired (b98c9746… on main)
  ※ settings drift 검출 1건 (아래 별도 절)
moai integration release                 → released (수리 미완 → 창 반납, 병합 미수행)
```

## 흡수

흡수 직전 워크트리 안에서 재측정 — 리드가 준 값과 일치.

```
로컬 develop tip          = 08609c198
origin/develop            = 91d25bc61
origin/develop..develop   = 26
카드 브랜치               = WT-spec-lint-axes @ cc63fa54f
git merge develop         → CONFLICT (CHANGELOG.md 1건)
흡수 커밋                 = 27317a6dd
```

## 충돌 1건 — CHANGELOG.md, 양쪽 보존

독립 카드끼리의 충돌. 마커 3줄만 제거(lane-3 선례와 동일).

```
충돌 전 1338줄 → 해결 후 1335줄 (마커 3줄 차)
마커 제거본 vs 해결본 diff → IDENTICAL (내용 손실·추가 0)

SPEC-SPEC-LINT-BLIND-AXES-001      : 1   ← HEAD 측
SPEC-SPEC-LINT-ID-ARG-001          : 1   ← HEAD 측
SPEC-DOCS-LOCALE-PARITY-REPAIR-001 : 1   ← develop 측
SPEC-CTX-BLIND-DOUBLE-001          : 1   ← develop 측
SPEC-NONEXISTENT-CONTROL-999       : 0   ← 대조군, 0 이 침묵이 아니라 측정
```

## 병합 트리 재측정 — 미완 (BLOCKER 로 중단)

```
go vet ./internal/spec/... ./internal/cli/... ./internal/kanban/...
  → exit 0, 출력 0줄

go test ./internal/spec/... ./internal/kanban/... -count=1
  → exit 1
     ok   internal/kanban  142.695s
     FAIL internal/spec     72.277s
       --- FAIL: TestTableCollection_CorpusListFindingsUnchanged (0.27s)
         code CoverageIncomplete: narrow-path findings = 16,
         live non-advisory findings = 0 (scanned 812 docs)

go test ./internal/cli/... -timeout 900s
  → 미실행 (블로커로 중단)
```

## 블로커 귀속 — 코드 충돌이 아니라 코퍼스 성장

`TestTableCollection_CorpusListFindingsUnchanged` (AC-SLB-003) 는
`.moai/specs/SPEC-*/spec.md` 코퍼스를 읽어 **코드별 `narrow(전체) == live(비자문)`** 을 단언한다.

일회용 프로브(`internal/spec/zz_t518_probe_test.go`, 측정 후 삭제)로 narrow 측
`CoverageIncomplete` 16건의 출처를 열거한 결과 — **문서 하나에서 전부 나온다**:

```
PROBE 16 .moai/specs/SPEC-AC-COLLECTOR-ANCHOR-001/spec.md
(그 외 문서 0건)

git cat-file -e cc63fa54f:.moai/specs/SPEC-AC-COLLECTOR-ANCHOR-001/spec.md
  → fatal: exists on disk, but not in 'cc63fa54f'
```

즉 t528 이 develop 으로 들여온 **자기 SPEC 문서**가 코퍼스에 들어오면서 narrow 측
계수가 0 → 16 으로 움직였다. t528 의 **코드** 변경(`parseSingleACLine` 확장)과의
의미적 충돌이 아니다 — 새 문서 1개가 추가된 결과다.

### 단언 자체의 결함 (내 카드 소관)

`CoverageRule.Check` 는 `internal/spec/lint.go:1107` 에서 **무조건** `Advisory: true` 를
붙인다(SPEC-COVERAGE-RULE-SCOPE-001 M3 의 의도된 설계). 그러므로

- live 측은 `!f.Advisory` 로 거르므로 이 코드에서 **구조적으로 항상 0**
- narrow 측은 자문 여부를 거르지 않으므로 **findings 가 생기는 즉시 > 0**

두 계수는 **같은 양을 재고 있지 않다**. 병합 전까지 등식이 성립한 것은 narrow 측이
코퍼스 전체에서 `LegacyEARSKeyword` 7건만 냈기 때문이며(그 제한은 progress.md
M-A1 절에 명시돼 있다), **우연한 0 == 0 이었다.** 문서 하나가 늘자 드러났다.

### 판단이 필요한 지점

수리 후보는 like-for-like 비교 — narrow 측에도 `!f.Advisory` 필터를 적용하는 것이다.
그러면 `CoverageIncomplete` 는 양쪽 0 으로 비교 대상에서 빠지고, 변별력은
`LegacyEARSKeyword` 7건 + 순서 의존 `DuplicateREQID` 로 남는다(비공허, 기존 M-A1
제한 서술과 동일 범위).

다만 이는 **AC-SLB-003 이 단언하는 내용을 바꾸는 편집**이므로 레인이 창 안에서
단독으로 집행하지 않고 리드 판정으로 올린다.

## settings drift (창 획득 시 검출 — 처분은 사람 몫)

```
file:      /Users/goos/MoAI/moai-adk-go/.claude/settings.json
sha256:    f509fffaddce38fa5a70357b9548a8214d60d883113a82214e2e9b5f89145cde
size:      23556 bytes
preserved: /Users/goos/moai/moai-adk-go/.moai/state/settings-drift/settings.json.main.20260908T031051.626Z.f509fffa
```

복원·되돌림·삭제 없음.

## 잔여 미검증

- `internal/cli` 재측정 미실행(블로커로 중단) — 수리 확정 후 `-timeout 900s` 로 실행 필요
- CI 판정 없음 (미푸시)
- `internal/spec` 의 나머지 테스트는 이 1건 외 전부 통과했는지 미분해 — FAIL 요약만 관측

---

## 후속 — 블로커 해소 (리드 승인 후)

리드가 수리를 승인했다(승인 주체는 리드). 조건은 **비공허성 증명**이었고, 뮤턴트 C 로
RED 를 관측해 충족했다. 경위·뮤턴트 3종(못 잡은 2종 포함)·잔존 변별력 계수·수리 후
재측정은 같은 디렉터리의 `revision-record.md` 가 정본이다.

수리 후 재측정 요약: `go vet` exit 0/무출력 · `internal/spec` ok 72.688s ·
`internal/kanban` ok 142.460s · `internal/cli` ok 461.510s(+하위 16패키지 ok).
