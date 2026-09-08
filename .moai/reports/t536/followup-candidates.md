# t536 — 후속 카드 후보 (반출 기록)

카드 t536(`SPEC-TODO-HOME-TEMP-GUARD-001`)의 plan-phase 도중 관측됐으나 **이 카드의 범위 밖**으로 판정된 항목이다. 카드 발행은 운영자 결재 사안이므로 여기서 발행하지 않는다 — 이 문서는 **조용히 잊히지 않게 하는 기록**이고, 리드가 다음 결재 때 함께 올린다.

기준 트리: `.claude/worktrees/t536`, 브랜치 `WT-home-fallback`, 기준 커밋 `075dd8d4b`.

---

## 후보 1 — 한 규칙의 두 변형이 세 자리에 산다

**관측 (읽어서 확인, 실행 아님)**

`internal/cli/todo_queue_root_test.go:243-253` `queueRootInsideTemp` 가 `REQ-THG-003` 의 정규화 규칙을 **이미 구현하고 있다**:

```go
p := filepath.Clean(root)
if resolved, err := filepath.EvalSymlinks(p); err == nil { p = resolved }
tmp := filepath.Clean(os.TempDir())
tmpResolved := tmp
if resolved, err := filepath.EvalSymlinks(tmp); err == nil { tmpResolved = resolved }
return underDir(p, tmp) || underDir(p, tmpResolved)
```

양쪽 `EvalSymlinks` + 실패 시 어휘적 `Clean`, 컴포넌트 포함 비교. 함수 주석까지 `/var/folders` → `/private/var/folders` 기제를 이 SPEC 과 같은 근거로 설명한다.

**다만 등가물이 아니다 — 이 구분이 이 항목의 핵심이다.**

| | `queueRootInsideTemp` | `REQ-THG-003` |
|---|---|---|
| 정규화 규칙 | 같음 | 같음 |
| 포함 비교 | 같음 | 같음 |
| **루트 집합** | **`os.TempDir()` 하나** | `{os.TempDir(), /tmp, /var/folders}` 셋 |

macOS 에서 `os.TempDir()` 은 `/var/folders/…` 를 돌려주므로 **이 헬퍼는 `/tmp` 기원을 잡지 못한다.** 그런데 이 카드가 확보한 **유일한 생산 오염 실측**(`~/.moai/todo/t203-probe-d7a16ea2` ← `/tmp/t203-probe`, `TodoQueueProjectKey` 역산으로 복원; `sha256("/tmp/t203-probe")[:4] = d7a16ea2` 일치, `/private/tmp/t203-probe` 는 `a8822e43` 로 불일치)이 **정확히 `/tmp` 기원**이다. 이 헬퍼를 그대로 승격했다면 **이 카드가 실증한 그 오염을 놓친 채 「통합했다」고 적었을 것**이다.

따라서 정확한 서술은 **「한 규칙 · 세 구현」이 아니라 「한 규칙의 두 변형 · 세 자리」**다.

**왜 이 카드 밖인가**

1. unexported 테스트 코드를 프로덕션 경로로 끌어올리는 것은 **새 설계 결정**이고, 5차 감사까지 온 카드에 얹을 일이 아니다.
2. 통합이 **단순 추출이 아니라 집합을 넓히는 동작 변경**을 수반한다 — 그 변경은 자기 몫의 검증을 요구한다.
3. 5차 감사가 이 out-of-scope 판정을 **옳다고 판정**했다.

**후속 카드가 다룰 것**: 세 자리(`queueRootInsideTemp` · `REQ-THG-003` 구현 · 그 밖에 같은 규칙이 재등장하는 자리)를 한 곳으로 모으되, **루트 집합을 넓히는 쪽으로** 통일하고 그 확대가 무엇을 새로 잡는지 실측으로 보일 것.

---

## 후보 2 — `§D.0` 커버리지 선언의 비대칭 (5차 감사 D28, optional)

**관측**

`acceptance.md §D.0` 은 §D 매트릭스의 요구사항 칸을 `AC-… maps REQ-…` 기계 판독 형식으로 한 번 더 적는다. 이 이중화는 드리프트를 **한 방향으로만** 잡는다:

| 방향 | 기계가 잡는가 |
|---|---|
| 매트릭스 → `§D.0` (매트릭스가 앞서 바뀜) | **잡는다** — `CoverageRule` 이 커버 안 된 REQ 를 보고 |
| `§D.0` → 매트릭스 (`§D.0` 이 앞서 바뀜) | **못 잡는다** |

그리고 **못 잡는 쪽이 D21 의 방향**이다 — D21 은 매트릭스 행이 실제 산출 요구사항과 어긋난 결함이었다.

**감사가 함께 준 해소책**: `CoverageRule` 의 정규식이 파일 전체를 읽으므로, `maps REQ-…` 를 **매트릭스 셀 안에** 넣으면 두 자리가 하나로 합쳐지고 비대칭이 소멸한다.

**왜 이 카드 밖인가**: 5차 감사가 optional 로 등급했고, 운영자가 이번 회차에서 닫지 말라고 결재했다. `§D.0` 자체는 감사가 **건전** 판정했다(전제를 소스에서 확인: `internal/spec/lint_coverage_sibling.go:114` → `internal/spec/ears.go:128` 이 `maps REQ-…` 만 매치).

**함께 기록하는 코퍼스 잔여** — 이 규칙을 발화시키는 47개 SPEC 중 **43개**가 `maps` 선언 없는 `acceptance.md` 를 갖는다. **t536 에서 쓸어담지 않는다**(5차 감사도 이 처분에 동의). 후속 카드가 이 코퍼스 축까지 볼지는 별도 판단이다.

---

## 이 문서가 주장하지 않는 것

- 두 항목 모두 **실행으로 확인한 것이 아니다** — 소스 판독과 기존 감사 산출물 인용이다. 후보 1 의 키 역산만이 계산으로 재유도된 값이다.
- 어느 항목도 **카드로 발행되지 않았다.** 발행은 운영자 결재 사안이다.
- 후보 1 의 「그 밖에 같은 규칙이 재등장하는 자리」는 **전수 조사하지 않았다.** 확인된 것은 두 자리이고, 세 번째는 `REQ-THG-003` 이 run-phase 에서 만들 구현이다.
