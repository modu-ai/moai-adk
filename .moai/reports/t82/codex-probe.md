# t82 — codex AGENTS.md 로딩 실측 (3개 전제 해소)

리드 판정(2026-08-22)에 따라 t91 대기 없이 t82 레인이 직접 실측했다.
바이너리 `codex-cli 0.147.0` (`/Users/goos/.local/bin/codex`).

**측정 수단**: `codex debug prompt-input <prompt>` — 모델에게 실제로 전달되는 prompt input 목록을
JSON 으로 렌더한다. **모델 호출이 0회**이므로 비용 없이 "무엇이 실제로 실렸는가"를 직접 관측할 수 있다.
문서 신뢰가 아니라 관측이다.

**픽스처**: 스크래치패드에 git 리포를 만들고 `AGENTS.md`(루트) · `area/AGENTS.md` · `area/deep/AGENTS.md`
3층을 배치. 루트 문서에는 110 바이트 간격으로 오프셋을 이름에 박은 눈금 마커(`MARK00000`, `MARK00110`, …)를
40,040 B 까지 채웠다.

---

## 결과 요약

| # | 전제 | 판정 | 근거 |
|---|---|---|---|
| 1 | `project_doc_max_bytes` 기본값 32,768 B | **확인** | 눈금 마지막 통과 `MARK32670`, 다음 눈금 `MARK32780` 부재 |
| 2 | 중첩 AGENTS.md 병합 범위 | **git 프로젝트 루트 → CWD 경로 체인 한정** | 루트에서 실행 시 `area/*` 마커 0건 |
| 3 | 잘림 경고 가시성 | **무음** | stderr 0 바이트, 경고·종료코드 이상 없음 |

## 1. 기본 한도 = 32,768 B (확인)

```
$ cd <fixture>            # git 리포 루트
$ codex debug prompt-input probe            # stderr 0 B
$ grep -o 'MARK[0-9]*' out.json | tail -2
MARK32560
MARK32670                                    # 마지막으로 실린 눈금
```

다음 눈금은 오프셋 32,780 으로 32,768 을 넘어 실리지 않았다. 40,040 B 중 298개 눈금만 통과.

키가 실제로 소비되는지도 확인했다:

```
$ codex debug prompt-input -c project_doc_max_bytes=4096 probe
$ grep -o 'MARK[0-9]*' out.json | tail -1
MARK04070                                    # 4,096 경계에서 잘림
```

→ `project_doc_max_bytes` 는 실동작하는 키이고, 무설정 기본값은 32,768 B 다.
**잘리는 쪽은 뒤(tail)** — 앞부분이 살아남는다.

## 2. [중대] 중첩 AGENTS.md 는 CWD 경로 체인에서만 병합된다

세 번의 실행이 서로를 가른다.

| 실행 위치 | 루트 문서 크기 | 실린 마커 |
|---|---:|---|
| `area/deep` (git 리포 **아님**) | 42,066 B | `MARKER_DEEP` 만 |
| `area/deep` (git 리포) | 42,066 B | `MARKER_ROOT_HEAD`, `MARKER_ROOT_AT_31K` — **`MARKER_AREA`·`MARKER_DEEP` 소실** |
| `area/deep` (git 리포) | 28 B | `MARKER_ROOT_HEAD`, `MARKER_AREA`, `MARKER_DEEP` **셋 다** |
| 리포 루트 | 40,040 B | 루트 눈금만 — **`MARKER_AREA`·`MARKER_DEEP` 0건** |

세 가지가 동시에 따라 나온다.

- **프로젝트 루트 판정은 git 기준**이다. git 리포가 아니면 상위로 올라가지 않고 CWD 문서 하나만 읽는다.
- **병합 대상은 루트→CWD 경로 위의 문서뿐**이다. 리포 루트에서 codex 를 띄우면
  `area/AGENTS.md` 는 **읽히지 않는다**. 하위 디렉터리에 규칙을 두어도 그 디렉터리에서
  세션을 시작하지 않는 한 모델은 그것을 보지 못한다.
- **예산은 체인 공유이고 루트가 먼저 먹는다.** 루트가 42,066 B 이던 두 번째 실행에서
  중첩 문서 2개가 통째로 사라졌다 — 잘린 것도 아니고 아예 실리지 않았다. §3 대로 무음이다.

### 설계 귀결 — 카드의 중간 갈래는 그대로는 성립하지 않는다

카드안은 "영역별 규칙은 해당 디렉터리 중첩 AGENTS.md 각 ~4 KiB"로 예산을 넓히려 했다.
실측은 그 반대를 말한다:

1. 중첩 문서는 **예산을 넓히지 않는다**. 32,768 B 를 루트와 나눠 쓸 뿐이다.
2. 중첩 문서는 **평상시 로드되지 않는다**. 개발자가 리포 루트에서 `codex` 를 실행하는 통상 사용에서
   하위 문서는 전부 미로드다. 즉 **루트 AGENTS.md 하나가 자립해야 한다.**
3. 따라서 `[HARD]` 계약은 전량 **루트 32,768 B 안에** 들어가야 한다.
   중첩 문서는 "영역 전용 세션에서만 추가로 붙는 보너스"로 격하되며,
   그만큼 루트 예산을 갉아먹으므로 **개수를 늘릴수록 손해**다.

리드가 요청한 "중간 갈래 폐기 시나리오"는 시나리오가 아니라 **실측이 지시하는 기본안**이다.
plan 에는 다음 두 안을 병렬 수록하되 A 를 기본으로 둔다.

- **A(기본) — 루트 단일 계약**: 중첩 AGENTS.md 를 두지 않고 루트 하나에 계약 전량.
  예산 32,768 B 전부를 루트가 쓴다. 통상 사용에서 손실 0.
- **B(조건부) — 루트 + 소수 중첩**: 루트를 8 KiB 로 줄이고 나머지를 중첩에 두는 안.
  **채택 조건**: 해당 영역에서 세션을 시작하는 것이 실제 작업 습관이라는 근거가 있을 때만.
  근거 없이 채택하면 통상 사용에서 규칙의 3/4 가 사라진다.

## 3. 잘림은 무음이다 — 회귀 가드가 유일한 방어선

`codex debug prompt-input` 40,040 B 실행에서 **stderr 0 바이트**, 종료코드 0.
바이너리에 `project doc exceeds remaining budget; truncating` 문자열이 있으나 이는 `tracing` 이벤트
(`core/src/agents_md.rs:124`)로, 기본 로그 레벨에서 사용자에게 노출되지 않는다.

귀결: **예산 초과는 사용자에게 아무 신호도 주지 않는다.** 규칙 파일이 커져 계약 뒷부분이 잘려도
codex 는 조용히 계속 동작하고, 잘린 규칙은 그저 존재하지 않는 것처럼 행동한다.
따라서 t82 의 회귀 가드는 선택이 아니라 **필수**이며, 런타임이 아니라 **CI 바이트 가드**여야 한다
(`internal/config/token_budget_guard.go` 의 표면 열거를 재사용).

## 검증 재현

```
fixture=<scratchpad>/codexprobe          # git init 된 3층 픽스처
cd $fixture && codex debug prompt-input probe            # 루트만 로드, 32,768 에서 잘림
cd $fixture/area/deep && codex debug prompt-input probe  # 루트+area+deep (루트가 작을 때)
codex debug prompt-input -c project_doc_max_bytes=4096 probe   # 키 실동작 확인
```

## 미검증 / 잔여

- `AGENTS.override.md` 우선순위와 `project_doc_fallback_filenames` 는 심볼 존재만 확인했고
  동작을 실측하지 않았다 — t82 설계가 이 둘에 의존하지 않으므로 범위 밖으로 둔다.
- 글로벌 `~/.codex/AGENTS.md` 층은 픽스처에 두지 않았다. 실제 사용자 환경에 이 파일이 있으면
  **같은 32,768 예산을 더 먼저 먹는다**고 보아야 하며(체인 병합), 이는 루트 예산을 더 좁힌다.
  배포 문서에 이 경고를 넣을지는 plan 에서 결정한다.
- 측정은 macOS + codex-cli 0.147.0 단일 환경이다. 다른 OS·버전에서 기본값이 다를 가능성은 배제하지 않았다.
