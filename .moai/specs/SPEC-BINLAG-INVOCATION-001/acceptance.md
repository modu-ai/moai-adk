# SPEC-BINLAG-INVOCATION-001 — 수락 기준

문서 수준 트리 핀: 아래 RED-now 셀은 워크트리 `.claude/worktrees/t366`, HEAD **`d7010f86a`**(plan-phase)에서 관측했다.
run-phase가 더한 GREEN 관측은 같은 워크트리 HEAD **`968c9caf8`**에서 얻었으며, 각 셀에 그 SHA를 명시한다.
개별 셀이 자기 SHA를 갖지 않으면 문서 핀이 구속한다.

> **[HARD] Option B 채택에 따른 재서술 (2026-08-31)** — 게이트가 **Option B(절차만, 코드 무변경)** 를
> 골랐다. 아래 8건 중 발화(경고 출력)를 전제하던 항목은 성격이 바뀌었다. 그 변화를 각 항목 머리에
> 명시하며, **자명 충족**으로 강등된 항목을 통과 건수로 합산하지 않는다 — 발화가 없어 자동으로 참이
> 되는 것은 성취가 아니라 미착수의 부산물이다(`spec.md` §2.0 · R-4).

읽는 법: 이 파일의 항목은 **Given-When-Then**이다. GEARS 문장은 요구 계층(`spec.md` §2)에 있고 여기 사본을 두지 않는다.

---

## §D 수락 기준 매트릭스

| ID | 요구 | 분류(원안) | **B 아래 분류** | 방향 | 판정 |
|---|---|---|---|---|---|
| AC-BLI-001 | REQ-BLI-001 | 릴리스 차단 | **릴리스 차단** (유지) | 증거 귀속 | PASS |
| AC-BLI-002 | REQ-BLI-002 | 릴리스 차단 | **릴리스 차단** (재서술) | 관측 — 갈림이 실재 | PASS |
| AC-BLI-003 | REQ-BLI-002, REQ-BLI-008 | 릴리스 차단 | **릴리스 차단** (유지) | 재현 — 격리 | PASS |
| AC-BLI-004 | REQ-BLI-003 | 회귀 가드 | **릴리스 차단** (재서술) | 관측 — 침묵의 대칭 | PASS |
| AC-BLI-005 | REQ-BLI-004 | 릴리스 차단 | **자명 충족** | 단일 구현 | 합산 제외 |
| AC-BLI-006 | REQ-BLI-005 | 릴리스 차단 | **자명 충족** | 표면 무해성 | 합산 제외 |
| AC-BLI-007 | REQ-BLI-006 | 릴리스 차단 | **자명 충족** | 뜨거운 경로 | 합산 제외 |
| AC-BLI-008 | REQ-BLI-007 | 릴리스 차단 | **자명 충족** | 거짓 보편 금지 | 합산 제외 |

실질 판정 대상은 **4건**(AC-BLI-001 · 002 · 003 · 004)이며 전부 PASS. 나머지 4건은 조건부 후속(A/C)이
착수될 때 판정 대상으로 복귀한다. 원안 표의 「분류(원안)」 열을 지우지 않은 것은 강등을 사후에
재구성하지 않기 위해서다.

**AC-BLI-004의 분류가 회귀 가드에서 릴리스 차단으로 *올라간* 이유**: 원안에서 그것이 회귀 가드였던
근거는 「지금은 어떤 권고도 발화하지 않으므로 침묵이 자동으로 참」이었다. B 아래에서는 그 문장이
성립하지 않는다 — 재서술된 AC-BLI-004는 침묵을 단언하지 않고 **침묵이 양쪽에서 동일하다는 것**을
관측하며, 그것은 오늘 실제로 재야 하는 값이고 이 카드의 결론 자체다(`spec.md` HISTORY 0.2.0).

---

### AC-BLI-001 — 증거로 인용된 측정은 자기 트리에서 빌드된 바이너리가 냈다

**Given** 어떤 판정 보고가 `moai` 하위 명령의 출력을 증거로 인용하고,
**When** 그 인용의 baseline 귀속을 읽으면,
**Then** 그 출력을 낸 바이너리의 빌드 커밋과 측정 대상 트리의 HEAD가 함께 적혀 있고, 전자가 후자의 진조상이 아니다.

- 판정 방식: 보고 본문에서 인용된 명령 옆에 바이너리 빌드 커밋이 있는지 읽는다. 없으면 실패.
- RED-now — 지금 이 규율을 강제하는 것이 아무것도 없다:

      $ grep -rn "binlag" internal/cli/root.go
      (출력 없음)
      exit 1

  호출 시점에 지연을 알리는 코드가 트리에 존재하지 않으므로, 인용자가 지연을 모른 채 인용하는 것을 막는 기계적 장치는 0이다.

- **GREEN (Option B, 트리 `968c9caf8`)** — 규율이 always-loaded 규칙에 착지했다. 명령·출력·종료 상태:

      $ grep -c "### 2.2 Tool-provenance attribution" .claude/rules/moai/core/verification-claim-integrity.md
      1
      exit 0

      $ grep -c "### 2.2 Tool-provenance attribution" internal/template/templates/.claude/rules/moai/core/verification-claim-integrity.md
      1
      exit 0

  [HARD] **B안이 바꾸는 것은 강제력이 아니라 의무의 성문화다.** RED-now 셀이 적은 「기계적 장치 0」은
  **여전히 참이다** — 코드 변경이 0이므로. 이 GREEN은 「규율이 문서화된 의무로 존재하고, 인용이
  두 좌표(판정한 빌드의 커밋 + 측정 대상 트리 HEAD)를 담도록 요구한다」까지만 세운다. 규율을 모르는
  사람이 계속 밟는 것은 이 카드가 해소하지 못하며, 그것이 조건부 후속(A/C)의 발동 조건이다.

---

### AC-BLI-002 — 규칙 집합이 다르면 판정이 실제로 갈린다 (관측: 다른 쪽) · **B안 재서술**

> **원안** — 「stderr에 바이너리 커밋과 트리 HEAD를 둘 다 명명하는 권고가 나타난다」. B안은 코드 변경이
> 0이므로 발화 자체가 없다. 원문 그대로는 이 카드에서 **충족 불가**이며, 「호출이 알린다」는 조건부
> 후속(A/C)으로 이월한다. 아래는 B가 실제로 세우는 것 — **갈림이 실재한다는 관측**이다.

**Given** 트리에는 존재하고 비교 대상 바이너리에는 존재하지 않는 규칙이 있고(그 바이너리의 빌드 커밋이 트리 HEAD의 진조상),
**When** 두 바이너리로 각각 같은 트리를 읽는 같은 하위 명령을 호출하면,
**Then** 두 stdout이 실제로 다르고, 실행되지 않은 규칙을 코드 단위로 귀속할 수 있으며, 그 규칙이 비교 대상 바이너리의 빌드 커밋 **이후에** 착지했음이 조상 관계로 확인된다.

- 판정 방식: 두 stdout을 파일로 받아 `cmp`로 대조하고, 규칙 코드 집합의 차집합을 구해 그 규칙의 착지 커밋을 `git merge-base --is-ancestor`로 귀속한다.
- **GREEN (트리 `968c9caf8`)** — 명령·출력·종료 상태:

      $ cmp -s installed.stdout tree.stdout
      (출력 없음)
      exit 1                  # 1 = 두 출력이 다르다

      $ tail -1 installed.stdout ; tail -1 tree.stdout
      0 error(s), 64 warning(s)
      0 error(s), 177 warning(s)
      exit 0

      $ comm -13 rules.installed.txt rules.tree.txt
      MovingRefUnpinned
      exit 0

      $ git merge-base --is-ancestor 343399d2f 84b3b7949
      (출력 없음)
      exit 0                  # MovingRefUnpinned 는 설치본이 빌드된 뒤 착지했다

  설치본(`343399d2f`)이 **113개 warning 을 아예 발화하지 않았다**. 증거 원본:
  `.moai/reports/t366/evidence/` · 서사: `.moai/reports/t366/run-observation.md` §2.

- RED-now(원안, 트리 `d7010f86a`) — 발화 코드 부재는 **여전히 참이다**:

      $ grep -rn "binlag" internal/cli/root.go
      (출력 없음)
      exit 1

---

### AC-BLI-003 — 재현은 자기 바이너리를 빌드하고 공용 설치본을 건드리지 않는다

**Given** AC-BLI-002 / AC-BLI-004의 재현 절차가 실행되고,
**When** 절차가 사용한 바이너리 경로와 절차 전후의 공용 설치본 상태를 읽으면,
**Then** 사용된 바이너리는 세션 전용 임시 디렉터리 아래에 있고, `/Users/goos/go/bin/moai`의 빌드 커밋·mtime은 절차 전후로 동일하다.

- 판정 방식: 절차 시작 전과 종료 후에 설치본의 `version` 출력을 각각 파일로 남기고 대조한다. 두 값이 다르면 실패 — 아홉 레인이 그 경로를 공유하므로, 조용한 갱신은 다른 레인의 측정을 무효화한다.
- RED-now — 절차 자체가 아직 존재하지 않는다:

      $ grep -rn "binlag" internal/cli/root.go
      (출력 없음)
      exit 1

  구속할 절차가 없으므로 기준은 미충족 상태다. (이 셀은 AC-BLI-002와 같은 관측을 인용한다 — 같은 부재가 두 기준을 동시에 RED로 만든다.)

- **GREEN (트리 `968c9caf8`)** — 절차가 실행됐고 공용 설치본은 무변경이다:

      $ ~/go/bin/moai version | tail -1        # 절차 전 / 절차 후 동일
       v3.1.2   343399d2f   built 2026-08-27T14:07:38Z
      exit 0

      $ ls -l ~/go/bin/moai                    # 절차 전 / 절차 후 동일
      -rwxr-xr-x@ 1 goos  staff  68955858 Aug 27 23:09 /Users/goos/go/bin/moai
      exit 0

  재현 바이너리는 세션 전용 임시 디렉터리(`<session-scratchpad>/binlag/moai-tree`)에 빌드했고,
  `make install`을 호출하지 않았다. 서사: `run-observation.md` §1 · §6.

---

### AC-BLI-004 — 침묵은 양쪽에서 동일하다 (관측: 같은 쪽) · **B안 재서술 · 릴리스 차단으로 승격**

> **원안** — 「규칙 집합이 같으면 경고가 반드시 침묵한다」, 분류 **회귀 가드**. 그 분류의 근거는
> 「지금은 어떤 권고도 발화하지 않으므로 침묵이 자동으로 참」이었다. **B안 아래에서 그 문장이 이 기준의
> 내용을 무의미하게 만든다** — 영영 발화가 없으므로 침묵 단언은 영원히 자명하다. 그래서 이 기준을
> 침묵의 **단언**에서 침묵의 **대칭성 관측**으로 바꾸고, 오늘 실제로 재야 하는 값이 되므로 분류를
> 릴리스 차단으로 **올린다**.

**Given** 지연된 호출(AC-BLI-002의 설치본)과 일치한 호출(트리 빌드) 두 가지가 있고,
**When** 두 호출의 stderr와 종료 상태를 각각 읽으면,
**Then** 둘이 **구별되지 않는다** — 양쪽 모두 stderr 0바이트, 종료 상태 0. 즉 출력에는 지연 여부를 가르는 신호가 없다.

- 이 기준을 두는 이유는 그대로다. AC-BLI-002만 있으면 「출력이 다르네」로 끝나 결함의 성격에 닿지 못하고, AC-BLI-004만 있으면 「조용하니 정상」으로 끝난다. **두 방향을 함께 재야만** 「침묵이 아무것도 말하지 않는다」가 보인다.
- 판정 방식: 두 호출의 stderr 바이트 수와 종료 상태를 각각 기록해 대조한다. 두 쌍이 동일하면 이 기준은 충족이다 — 충족이 곧 **결함의 실증**이라는 것이 이 카드의 요지다.
- **GREEN (트리 `968c9caf8`)** — 명령·출력·종료 상태:

      $ wc -c installed.stderr tree.stderr
             0 installed.stderr
             0 tree.stderr
             0 total
      exit 0

      # 두 호출의 종료 상태
      installed exit=0
      tree      exit=0

  판정 자체는 존재하고 정확하다 — 다만 별도 명령을 타이핑해야 닿는다:

      $ ~/go/bin/moai doctor | grep -i 'Binary Freshness'
      warn  Binary Freshness  binary is behind source tree (binary: 343399d2f, HEAD: 968c9caf8)
      $ <tree-build> doctor | grep -i 'Binary Freshness'
      ok    Binary Freshness  binary matches source HEAD (968c9caf8)
      exit 0

  서사: `run-observation.md` §3-§5.
- **적용 불가 관용은 이 카드에서 재지 않았다.** 원안이 요구한 「비-git 디렉터리에서의 침묵」은 코드
  변경이 0이므로 `internal/binlag`의 관용이 그대로 보존되지만, 이 세션은 그것을 **실행해 확인하지
  않았다**. 무변경의 부산물이지 관측이 아니다 — 조건부 후속(A/C) 착수 시 실측 대상이다.

---

### AC-BLI-005 — 지연 판정의 구현은 하나뿐이다 · **B안: 자명 충족(합산 제외)**

> B안은 코드 변경이 0이므로 새 소비자가 생길 수 없고, 따라서 중복 구현도 생길 수 없다. 이 기준은 **미착수의 부산물로 참**이며 통과 건수에 합산하지 않는다. 조건부 후속(A/C) 착수 시 실제 판정 대상으로 복귀한다.

**Given** 수리가 착지한 트리에서,
**When** 조상 비교를 수행하는 코드를 전수 조사하면,
**Then** `internal/binlag` 밖에서 `merge-base --is-ancestor`를 호출하거나 자체 조상 판정을 수행하는 비테스트 코드가 0건이다.

- 판정 방식: `merge-base` 문자열을 비테스트 Go 파일에서 검색하고, 매치를 `internal/binlag` 소속으로 전수 귀속한다.
- RED-now — 착지 전이라 새 소비자가 없다. 이 기준은 수리가 만들 수 있는 **중복 구현**을 막는 항목이므로, 오늘의 RED는 「새 소비자 0」이라는 자명한 상태다:

      $ grep -rln "binlag" --include='*.go' internal/cli/
      internal/cli/doctor.go
      exit 0

  트리 `d7010f86a`. 현재 `internal/cli`의 유일한 `binlag` 소비자는 `doctor.go`다. 수리 후 이 집합에 추가되는 파일은 전부 seam 소비여야 하며, 자체 비교를 들고 오면 실패다.

---

### AC-BLI-006 — 권고는 stdout도 종료 상태도 건드리지 않는다 · **B안: 자명 충족(합산 제외)**

> 권고가 존재하지 않으므로 오염시킬 대상이 없다. 미착수의 부산물로 참이며 합산하지 않는다.

**Given** 지연 상태에서 기계 판독 출력을 내는 하위 명령(`spec status` 계열, JSON 출력 경로)을 호출하고,
**When** stdout만 파서에 넘기면,
**Then** 파싱이 성공하고, 그 stdout은 지연 없는 동일 호출의 stdout과 바이트 동일하며, 종료 상태도 동일하다.

- 판정 방식: 지연 있는 호출과 없는 호출의 stdout을 각각 파일로 받아 바이트 비교한다. 권고 문자열이 stdout 쪽 파일에 나타나면 실패.
- RED-now — 권고가 존재하지 않으므로 「stdout 오염 없음」이 자동으로 참이다. 이 항목은 수리가 도입할 수 있는 회귀를 막는 성격이 강하지만, 권고 발화 자체가 없는 지금은 발화 경로가 stderr임을 확정할 수단도 없다:

      $ grep -rn "binlag" internal/cli/root.go
      (출력 없음)
      exit 1

---

### AC-BLI-007 — 뜨거운 경로는 호출당 저장소 비교를 얻지 않는다 · **B안: 자명 충족(합산 제외)**

> 붙일 비교가 없으므로 뜨거운 경로는 아무것도 얻지 않는다. 면제 집합 결정(`hook` 추가 여부) 자체가 불요가 됐다. 미착수의 부산물로 참이며 합산하지 않는다 — 이것이 게이트가 B를 고른 주된 근거이기도 하다(`spec.md` §6 결정 배너).

**Given** 수리가 착지한 트리에서,
**When** `moai hook <event>` 및 런처(`cc` / `cg` / `glm`) 호출 경로가 지연 비교를 수행하는지 확인하면,
**Then** 그 경로들은 비교를 수행하지 않으며, 면제 집합은 기존 `trivialCommands`에서 파생됐고 별도 병렬 목록이 새로 만들어지지 않았다.

- 판정 방식: 면제 집합이 정의된 위치를 읽어 `trivialCommands`와의 관계(확장/참조)를 확인하고, `hook`이 포함되었는지 본다. 새 리터럴 목록이 별도로 선언되어 있으면 실패.
- RED-now — `hook`은 지금 건너뛰기 목록에 없다:

      $ grep -n '"hook"' internal/cli/root.go
      (출력 없음)
      exit 1

  트리 `d7010f86a`. 즉 루트 수준 훅을 지금 그대로 달면 43개 래퍼 중 39개가 쓰는 `moai hook <event>` 경로가 전부 비교를 얻는다.

---

### AC-BLI-008 — 보편 문장은 예외를 명명한다 (거짓 보편 금지) · **B안: 자명 충족(합산 제외)**

> 루트 `PersistentPreRunE`를 설치하지 않으므로 「모든 하위 명령이 알린다」를 쓸 자리가 없다. `worktree` 하위 트리의 비연쇄 성질(RED-now 셀)은 사실로 남지만, 이 카드가 그것에 걸리지 않는다. 미착수의 부산물로 참이며 합산하지 않는다.

**Given** 수리가 루트 수준 `PersistentPreRunE`를 사용하고(Option A 계열),
**When** SPEC과 구현 문서에서 「모든 / 어떤 하위 명령이든 권고한다」 형태의 문장을 찾으면,
**Then** 그런 문장이 0건이거나, 존재한다면 같은 문단에서 `worktree` 하위 트리를 예외로 명명하거나 연쇄를 명시적으로 만들었음을 진술한다.

- 판정 방식: cobra의 `PersistentPreRunE` 비연쇄 성질을 실제로 확인하는 테스트(루트 훅 설치 후 `moai worktree` 계열 호출에서 권고가 나오는지)를 함께 둔다. 문서 문장 검사만으로는 통과시키지 않는다.
- RED-now — 대체될 부모 훅이 이미 존재한다:

      $ grep -rn "PersistentPreRun" internal/cli/root.go
      internal/cli/root.go:127:	worktree.WorktreeCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
      exit 0

  트리 `d7010f86a`. `worktree` 하위 트리는 자기 `PersistentPreRunE`를 가지므로 루트 것을 **대체**한다. 「모든 하위 명령」이라 쓰면 그 문장은 쓴 대로 거짓이다.

---

## §D.1 Definition of Done (B안 판정 기준)

- **실질 판정 대상 4건**(AC-BLI-001 · 002 · 003 · 004)이 전부 PASS다. 자명 충족 4건(005-008)은 합산하지 않으며 각 항목에 사유를 유지한다.
- 양방향 관측(AC-BLI-002 다른 쪽 + AC-BLI-004 같은 쪽)이 **둘 다** 실행됐다. 한쪽만 관측한 상태는 완료가 아니다.
- 공용 설치본 `/Users/goos/go/bin/moai`가 절차 전후로 동일하다(AC-BLI-003).
- 규율이 always-loaded 규칙과 그 템플릿 미러 **양쪽에** 착지했다(AC-BLI-001).
- 코드 변경이 0이므로 테스트·lint 판정 대상이 없다. **이것은 통과가 아니라 미해당이다** — 「테스트 통과」로 보고하면 재지 않은 것을 잰 것으로 세는 일이다.

## §D.2 미검증으로 남기는 것

- 「이 결함이 실제로 오염시킨 판정의 수」 — 미측정(`spec.md` §8 R-1).
- 「호출당 비교 비용」 — 미측정. Option A 미채택으로 불요(R-2).
- 명령 단위 census — 존재하지 않음. Option C 미채택으로 만들지 않았다.
- **Option B의 전제**(「규율은 이미 있는데 발생했다」) — 미검증(R-3). 리드 배차문의 세 조우(t343 · t362 · t357)를 각 레인 보고서에서 직접 확인하지 않았고, 따라서 SPEC 근거로 인용하지 않는다.
- **적용 불가 관용의 실측** — 비-git 디렉터리에서의 침묵을 실행해 확인하지 않았다(AC-BLI-004 말미).
- **갈림의 폭 일반화** — `spec lint` 한 명령에서만 쟀다(113 warning). 다른 트리 판독 명령의 폭은 미측정(R-5).
