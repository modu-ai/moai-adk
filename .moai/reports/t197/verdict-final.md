# t197 — 최종 확인 감사 판정

| 항목 | 값 |
|---|---|
| 대상 | `SPEC-CODEX-LAUNCHER-001` (spec / plan / acceptance) |
| 감사 내용 핀 | `01f5c531b` (직전 `417381af3`), 트리 깨끗함 — 감사 중 이동 없음 |
| 실행 | `mcp__moai__audit_multi`, `project_root=<worktree toplevel>`, `target=HEAD` |
| 반영 커밋 | `fa6bd6c54` (감사 이후) |
| 일자 | 2026-08-24 |

## VERDICT — **FAIL** (감사 시점 핀 기준), 지적 3건 전량 반영 후 재감사 필요

**점수 0.75 / 1.00**. Tier M 통과선 **0.80** 미달. iteration 3 의 0.63 에서 +0.12.

| 차원 | iter3 | final | 변화 |
|---|---:|---:|---|
| 명확성 | 0.75 | 0.75 | — |
| 완전성 | 0.50 | 0.75 | +0.25 |
| 시험 가능성 | 0.50 | 0.50 | — |
| 추적성 | 0.75 | 1.00 | +0.25 |

## 백엔드별 판정

| 백엔드 | 게이트 | verdict 필드 | 본문 판정 | 채택 |
|---|---|---|---|---|
| claude | required | pass | pass | pass |
| codex | required | fail | FAIL 0.75 · 차단 3건 | **FAIL** |
| glm | advisory | inconclusive | — | fail-open |

`overall_verdict: fail`, `disagreement_flag: true`. **이번 라운드는 verdict 필드와 본문이 일치했다.**

### 판정 필드 불일치 이력 정정

iteration 3 판정서에 "엔진이 구조화된 필드만 읽고 본문을 읽지 않는다" 고 적었다. **그 일반화는 과했다** — 이번 라운드에서 필드와 본문이 일치했으므로, 항상 필드만 읽는 구조적 결함이라고 단정할 근거가 없다. 관측된 사실은 이것뿐이다:

| 라운드 | verdict 필드 | 본문 | 일치 |
|---|---|---|---|
| iter2 | pass | FAIL 5건 | 불일치 |
| iter3 | pass | FAIL 5건 | 불일치 |
| final | fail | FAIL 3건 | 일치 |

**3회 중 2회 불일치, 방향은 둘 다 필드가 관대한 쪽.** 재현 조건은 미상이다. 카드 후보로 남기되, 원인 서술은 "필드가 본문보다 관대하게 나오는 경우가 관측됨(2/3)" 까지만 주장한다.

## 차단 지적 3건 — 전량 반영 (`fa6bd6c54`)

### F1 (High) — `nonEmpty` 판정이 원문 바이트 비교라 뚫림

감사가 제안된 술어를 그대로 실행해 보였다:

```
'{}'  -> False      '{ }' -> True     ← 같은 뜻인데 갈린다
'null'-> False      'false'-> True    ← 타입이 아예 다른데 통과
'""'  -> False      '0'   -> True
                    '[]'  -> True
```

`auth_mode=apikey` + `OPENAI_API_KEY: false` 같은 stale·불량 파일이 인증된 것처럼 보인다.

**반영**: 판정을 원문에서 **JSON 타입** 으로 옮겼다. 자격 재료는 비어 있지 않은 JSON 문자열이어야 하고, 다른 타입(불리언·수·배열·객체·null)은 전부 부재. `tokens` 는 객체가 아니면 0. AC-CL-008 표에 12행 추가 — `{ }` 와 `false` 두 행이 바이트 비교 구현과 타입 판정 구현을 가른다.

### F2 (Medium) — 근거 스크립트가 자기완결이 아니고 read-only 도 아님

`probe.sh` 가 `/tmp/moai-t197` 과 `/tmp/t197-doctor.json` 을 소비하면서 만들지 않아 신규 검토자가 재현할 수 없었고, `set +e` 때문에 선행 조건이 없어도 스크립트는 성공으로 끝났다. 게다가 codex 호출이 PATH 별칭 생성을 시도해 "읽기 전용" 주장이 사실과 달랐다 (`WARNING: proceeding, even though we could not create PATH aliases`).

**반영**: 스크립트가 측정 대상 바이너리를 스스로 빌드하고 doctor JSON 도 스스로 만든다(자기 temp 디렉터리, 종료 시 삭제). 선행 조건 실패 시 명시적으로 중단한다. 인용하던 timing 과 codex 커맨드 수를 전사본 안으로 옮겼다. **"읽기 전용" 표기는 철회** 하고 무엇이 어디를 바꾸는지 명시했다.

### F3 (High) — AC-CL-015/016 이 필요한 상태를 시험하지 못함

배선 판정이 디렉터리 존재만 봐서, 빈 `.codex/` 나 파일 한쪽만 있는 상태가 "배선됨" 으로 통과하고 훅 0개로 기동될 수 있었다. AC-CL-016 은 "둘 다 있는데 import 줄만 없는" 상태를 빠뜨렸고, 로컬 지시 파일을 이름으로 지목하지 않았으며, import 구문을 정하지 않은 채 "멱등" 을 1회 실행의 줄 수로 주장했다.

**반영**: 배선 판정을 **파일 집합** 으로 바꾸고 5종 상태 행렬(없음 / 빈 디렉터리 / 한쪽만 ×2 / 둘 다)을 AC 에 넣었다. import 줄 형태를 `@AGENTS.md` 로 고정, 로컬 파일을 `CLAUDE.local.md` 로 명명, "둘 다 있는데 미연결" 상태 추가, **멱등은 2회 실행 후 바이트 비교** 로 판정. REQ-CL-006/015/016 도 같은 방향으로 재기술.

## 감사가 확인한 iteration 3 해소 항목

- 자격 필드가 `nonEmpty`/`tokenSet` 로 바뀌었고, 리플렉션 기준이 `string`·`[]byte`·`json.RawMessage` 를 재귀적으로 배제한다
- 실행 seam 이 stdout/stderr 를 분리 반환하고 `combineCodexStreams` 를 거친다
- AC-CL-008 이 커밋된 fixture 실행 파일을 요구하고 시험 바이너리 재실행을 명시적으로 금지한다
- `classifyCodexAuthFile` / `readCodexAuthFile` 이 오류를 반환해 sentinel 오류 문안 단언이 작성 가능해졌다
- 변경 범위는 SPEC/근거 마크다운 4개뿐, scope creep 없음
- 자격 증명 스캔에서 실제 비밀값 노출 0건

## 내 측정 오류 2건 (같은 커밋에서 정정)

| 항목 | 이전 기록 | 실측 | 원인 |
|---|---|---|---|
| codex 커맨드 수 | 24 | **28** | 눈으로 셈 |
| `codex doctor --json` 소요 | 46초 | 이번 실행 **18초** | 캐시 상태 의존 — 단일 값으로 못 박지 않고 두 값 다 기록. 판단("대화형 리드아웃에는 과하다")은 초 단위 정확도에 의존하지 않는다 |

## 근거 명령

```
$ /tmp/moai-t197 spec lint .moai/specs/SPEC-CODEX-LAUNCHER-001/spec.md
✓ No findings — all SPEC documents are valid

REQ=16  AC=16   (Tier M 상한 16/16, 정확히 상한)
```

감사 백엔드 측 실측: `probe run calls=27` / `transcript command lines=27` / `transcript rc lines=27` — 스크립트의 단계 수와 전사본의 명령·rc 줄 수가 일치한다.

## 잔여 위험

- 구현이 아직 없으므로 fixture 실행 파일·스트림 동작·리플렉션 시험·생성기 호출·크로스 플랫폼·초기화 보존은 모두 run-phase 게이트 항목이다.
- 리플렉션은 **파싱된 구조체 필드에 자격 값이 남지 않음** 만 막는다. 원본 JSON 바이트와 파싱 중 임시 버퍼가 로그·오류로 새지 않는지는 run-phase 리뷰가 따로 봐야 한다 (감사의 residual-risk 지적).
- `apikey` / `provider` 모드의 실제 `auth.json` 형태는 여전히 미관측 — M1 에서 관측한다.

## 상태

지적 3건은 `fa6bd6c54` 에 전량 반영했다. **이 판정은 반영 이전 핀(`01f5c531b`) 기준이므로 현재 트리는 미판정** 이다. 재감사 여부는 리드 판단.
