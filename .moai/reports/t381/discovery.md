# t381 — 무시되는 증거 경로를 인용하는 코드 (착수 전 전수 분류)

- 카드: t381 · 워크트리 `.claude/worktrees/t381` · 브랜치 `WT-ignored-evidence-cite`
- 기준: `origin/develop` `3f03d9c36` (`merge-base --is-ancestor` 참)
- 측정 트리: `3f03d9c36` (이 트리, 이 실행)

---

## 1. 결함의 형태

코드 주석이 **gitignore 되는 경로**를 어떤 주장의 근거로 인용한다. clone·CI·타 머신에서 그 경로는 해소되지 않으므로, 인용은 있는데 근거는 없는 상태가 된다.

```
$ git check-ignore -v .moai/state/verify/t225/ac-amp-006-glm-differential-attempt1.md
.gitignore:298:.moai/state/	.moai/state/verify/t225/ac-amp-006-glm-differential-attempt1.md

$ ls .moai/state/verify/
ls: .moai/state/verify/: No such file or directory
```

두 번째가 직접 증명이다 — 이 워크트리는 방금 `origin/develop` 에서 만들어졌고, 인용된 디렉터리는 **존재하지 않는다**. 원 저작 머신 밖에서는 어디서도 해소되지 않는다.

> **배차문 정정**: 배차문은 ignore 규칙을 `.gitignore:284` 로 지목했으나 이 트리에서는 **`.gitignore:298`** 이다. 규칙 내용(`.moai/state/`)은 동일하다. 행 번호는 움직이는 좌표라 인용 시점이 다르면 어긋난다.

## 2. 전수 분류 — 인용 vs 경로를 쓰는 코드

배차문이 준 `.go` 8건을 요청대로 갈랐다. 축은 **"그 경로가 무엇인가"**다: 코드가 **만들어 내는** 경로인가(기계 소비자), 아니면 이미 있다고 **주장하는** 경로인가(인용).

| # | 위치 | 형태 | 분류 |
|---|---|---|---|
| 1 | `internal/verify/store.go:15` | `const SnapshotDir = ".moai/state/verify/snapshots"` | **기계 소비자** — 이 상수가 경로를 *정의*한다. 인용이 아니라 API |
| 2 | `internal/verify/schema.go:44` | 스냅샷이 어디 쌓이는지 서술한 doc comment | **기계 소비자** — 1번 상수의 설명 |
| 3 | `internal/web/events.go:29` | `"verify": {".moai/state/verify"}` | **기계 소비자** — 감시 대상 디렉터리 지정 |
| 4-6 | `internal/verify/store_test.go:26,44,45` | 1번 상수의 모양을 단언 | **기계 소비자** — 상수 검증 |
| 7 | `internal/cli/audit_pin_live_test.go:32` | "Evidence lands in `<repo>/.moai/state/verify/t225/`" | **인용** |
| 8 | `internal/cli/mcp_glm.go:110` | 측정치의 근거로 `.moai/state/verify/t225/ac-amp-006-…md` 지목 | **인용** (카드가 지목한 그것) |

**1-6은 carve-out이다.** 가드를 분류 없이 넓히면 이 여섯을 오탐으로 잡는다 — 경로를 정의하는 코드에게 "그 경로가 해소되지 않는다"고 말하는 것은 무의미하다. 그 경로는 **런타임에 만들어지며**, 만들어지기 전에 없는 것이 정상이다.

### 2.1 7번은 스스로 반증한다

`audit_pin_live_test.go:32` 의 문장 전문:

> Evidence lands in `<repo>/.moai/state/verify/t225/` (the card-scoped verify directory) **so the cited paths still resolve at audit time.**

이 트리에서 그 디렉터리는 존재하지 않는다(§1). 주석이 주장하는 성질을 ignore 규칙이 부정한다 — 인용이 끊긴 것에 더해, **끊기지 않는다는 진술까지 함께 실려 있다.**

## 3. 배차문 목록 밖에서 3건 더

배차문의 8건은 `.moai/state/verify` 전반을 훑은 결과로 보인다. 인용 형태(카드 id 디렉터리 지목)로 좁혀 **추적 파일 전수**를 다시 쟀다:

```
$ git grep -nE "\.moai/state/[a-z-]+/t[0-9]+" -- . ':!*.md'
.moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt:3
internal/cli/audit_pin_live_test.go:32
internal/cli/mcp_glm.go:110
internal/hook/evidence_writer_zeroexec_test.go:10
internal/session/ignored_content_test.go:26
```

새로 나온 3건:

| 위치 | 내용 | 분류 |
|---|---|---|
| `internal/hook/evidence_writer_zeroexec_test.go:10` | "Raw captures live under `.moai/state/verify/t341/`" — 러너 버전 실측치의 출처로 지목 | **인용** (배차문 목록에 없음) |
| `internal/session/ignored_content_test.go:26` | `porcelain: "!! .moai/state/verify/t209/\n…"` — 테스트 **입력 픽스처** | **carve-out** — 파싱 대상 문자열이지 주장의 근거가 아니다 |
| `.moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt:3` | "Source runs live at `.moai/state/verify/t299-sync/0{2,3,4,5}-*.txt`" | **인용** — 그리고 `.go` 도 `.md` 도 아니다 |

## 4. 훑은 범위 / 안 훑은 범위

**훑음** — `git grep` 으로 **추적 파일 전체**, `.md` 만 제외. 확장자 목록이 아니라 전수이므로 확장자를 빠뜨릴 여지가 없다.

**안 훑음**:
- `.md` — t375 가드 소관이라 의도적 제외. t375 가 실제로 `.md` 한정인지는 **미확인**(lane-8 이 run 중이고 develop 에 없음). 배차문 진술을 그대로 옮긴 것이며 제가 잰 것이 아니다.
- **미추적 파일** — `git grep` 은 추적 파일만 본다. 미추적 산출물 안의 인용은 이 계수에 없다.
- **`.moai/state/` 밖의 다른 ignore 경로** — `bin/`, `.claude/settings.local.json` 등을 인용하는 코드는 이번에 재지 않았다. 같은 결함 계열일 수 있으나 카드 범위 밖으로 두었다.

## 5. 카드 경계

**t375 를 고쳐도 이 축은 닫히지 않는다** — t375 가드가 `.md` 한정이기 때문(배차문 진술, §4 미확인 표시). 여기서 문제되는 4건 중 3건은 `.go`, 1건은 `.txt` 다.

파일 겹침: 제 인용 4건은 `internal/cli/`, `internal/hook/`, `.moai/reports/t299/` 에 있다. t375 가 가드 구현을 어디에 두는지 모르므로 겹침 여부는 **미확인** — 리드에게 병합 순서 판단을 요청한다.

## 6. Gaps / 잔여 위험

- **Gaps**: t375 가드의 실제 범위 미확인. 미추적 파일 미scan. `.moai/state/` 외 ignore 경로 미scan. 수리 방향은 아직 미정(plan-phase 소관).
- **잔여 위험**: 인용을 지우면 측정치의 출처가 사라진다 — `mcp_glm.go:110` 의 숫자(3667 vs 3480, budgets 3072 vs 1024, ratio 1.02)는 주석 본문에 이미 실려 있어 경로 없이도 읽히지만, `evidence_writer_zeroexec_test.go:10` 의 "raw captures" 는 본문에 값이 없어 경로를 지우면 출처가 통째로 사라진다. 세 인용은 같은 처리를 받으면 안 될 수 있다.
