# t366 run-phase — 양방향 관측 (SPEC-BINLAG-INVOCATION-001, Option B)

측정 트리: 워크트리 `.claude/worktrees/t366`, 브랜치 `WT-lint-binary-lag`, HEAD **`968c9caf8`**.
측정 시각: 2026-08-31. 아래 모든 수치는 이 트리에서 이 세션이 직접 실행해 얻었다.

Option B(절차만)가 채택됐으므로 코드 변경은 0이다. 따라서 이 관측은 「경고가 뜬다/안 뜬다」가 아니라
**절차를 지킨 측정과 지키지 않은 측정이 실제로 갈리는가, 그리고 그 둘을 가르는 신호가 존재하는가**를 잰다.

---

## §1 비교 대상 두 바이너리

| | 설치본 | 트리 빌드 |
|---|---|---|
| 경로 | `/Users/goos/go/bin/moai` | `<session-scratchpad>/binlag/moai-tree` |
| 빌드 커밋 | `343399d2f` | `968c9caf8` |
| `strings … grep -ci binlag` | `0` | `9` |

    $ git merge-base --is-ancestor 343399d2f 968c9caf8 ; echo $?
    0                       # 설치본은 트리 HEAD의 진조상 — 지연 확정

트리 빌드는 세션 전용 임시 디렉터리에 만들었고 `make install` 을 부르지 않았다(REQ-BLI-008).

---

## §2 다른 쪽 — 규칙 집합이 갈리는 경우

같은 명령 `moai spec lint` 를 같은 트리에 대고 두 바이너리로 각각 호출했다.

| | 설치본 `343399d2f` | 트리 빌드 `968c9caf8` |
|---|---|---|
| stdout 마지막 줄 | `0 error(s), 64 warning(s)` | `0 error(s), 177 warning(s)` |
| stdout 줄 수 | 68 | 181 |
| **종료 상태** | **0** | **0** |
| **stderr 바이트** | **0** | **0** |

    $ cmp -s installed.stdout tree.stdout ; echo $?
    1                       # stdout 은 다르다

**113개 warning 이 설치본에서는 아예 발화하지 않았다.** 실행되지 않은 규칙을 코드 단위로 귀속하면:

    $ comm -13 rules.installed.txt rules.tree.txt
    MovingRefUnpinned

    $ git log --oneline -1 84b3b7949
    84b3b7949 merge(WT-moving-ref-guard): SPEC-MOVING-REF-GUARD-001 … (t342)
    $ git merge-base --is-ancestor 343399d2f 84b3b7949 ; echo $?
    0                       # 그 규칙은 설치본이 빌드된 뒤에 착지했다

즉 t342 가 심은 `MovingRefUnpinned` 규칙은 설치본 안에 존재하지 않으며, 설치본의 `spec lint` 는
그 규칙을 **한 번도 돌리지 않고** 초록을 냈다.

---

## §3 같은 쪽 — 규칙 집합이 일치하는 경우

트리 빌드는 정의상 트리와 일치한다. 그 실행의 결과:

- stdout: `0 error(s), 177 warning(s)` — 규칙 전부 실행됨
- 종료 상태: `0`
- **stderr 바이트: 0**

---

## §4 이 카드가 닫으려는 결함 — 침묵이 대칭이다

§2 와 §3 을 나란히 두면 결함의 모양이 정확히 드러난다.

| | 지연된 호출(§2) | 일치한 호출(§3) |
|---|---|---|
| 종료 상태 | 0 | 0 |
| stderr | 비어 있음 | 비어 있음 |
| stdout 형태 | `0 error(s), N warning(s)` | `0 error(s), N warning(s)` |

**두 경우를 가르는 신호가 출력 어디에도 없다.** 초록을 증거로 인용하는 사람은 자기가 113개 규칙을
재지 않았다는 사실을 알 수 없다. 이것이 「초록이 통과인지 비실행인지 갈리지 않는다」의 실측이다.

한 방향만 봤다면 이 결론에 닿지 못한다. §2 만 보면 "출력이 다르네" 로 끝나고, §3 만 보면 "조용하니
정상" 으로 끝난다. **침묵이 양쪽에서 동일하다**는 사실은 두 쪽을 함께 재야만 보인다.

---

## §5 판정은 이미 존재하고 정확하다 — 도달만 없다

    $ ~/go/bin/moai doctor | grep -i 'Binary Freshness'
    warn  Binary Freshness  binary is behind source tree (binary: 343399d2f, HEAD: 968c9caf8)

    $ <tree-build> doctor | grep -i 'Binary Freshness'
    ok    Binary Freshness  binary matches source HEAD (968c9caf8)

`internal/binlag` 의 판정은 양방향 모두 맞다. 결함은 판정의 정확성이 아니라 **호출 시점에 그 판정이
소비되지 않는 것**이다(`grep -rn binlag internal/cli/root.go` → 출력 없음, exit 1). Option B 는 그
간극을 코드가 아니라 인용 규율로 덮는다.

---

## §6 공용 설치본 무변경 (AC-BLI-003)

| | 절차 전 | 절차 후 |
|---|---|---|
| `moai version` | `v3.1.2 343399d2f built 2026-08-27T14:07:38Z` | 동일 |
| size / mtime | `68955858` / `Aug 27 23:09` | 동일 |

아홉 레인이 공유하는 경로이므로 이 세션은 그것을 읽기만 했다.

---

## §7 미검증으로 남는 것

- 「이 결함이 실제로 오염시킨 판정의 수」 — 여전히 미측정(`spec.md` §8 R-1). 이 세션이 잰 것은
  **한 명령에서 갈림이 실재한다**는 것이지, 과거 판정 몇 건이 오염됐는지가 아니다.
- 호출당 저장소 비교 비용 — Option A 미채택으로 측정 불요(R-2).
- 명령 단위 census — Option C 미채택으로 미작성.
- `spec lint` 외 다른 하위 명령에서의 갈림 폭 — 재지 않았다. 한 명령으로 결함의 존재를 세웠을 뿐,
  그 폭을 계층 전체로 일반화하지 않는다.
- Option B 의 전제(「규율은 이미 있는데 발생했다」, R-3) — **여전히 미검증**. 리드 배차문의 세 조우
  (t343 · t362 · t357)는 이 세션이 측정하지 않았고, 각 레인 보고서를 직접 열어 확인하지도 않았다.
  따라서 이 SPEC 은 그 셋을 근거로 인용하지 않는다.
