# AC 카운트 baseline 재생성 — 수명주기 연동 cascade 절차

> SPEC-AC-BASELINE-REFRESH-001 (card t1068, 2026-09-23) 이 정본. `internal/spec/ac_count_clause_test.go` 의 `TestACCounterFullCorpusMatchesBaseline` 이 지키는 스냅샷 `.moai/reports/t338/ac-count-baseline.txt` 을 **언제, 어떻게** 다시 찍는지의 절차다. 이 문서가 존재하기 전에는 그 방법이 트리에 없었다 — 스크래치 스크립트로 손수 재생성하다 그 스크립트가 유실되는 사고가 두 번 반복됐다(아래 §5 사고 기록).

---

## 1. 재생성 명령

```
MOAI_AC_BASELINE_REGENERATE=1 go test ./internal/spec -run TestACCounterBaselineRegenerate -count=1
```

- 재생성 장치는 테스트 패키지 안에 있다. `extractCounterCommand` 으로 B12 절에서 카운터를 **추출**해 쓰므로, 카운터가 바뀌면 재생성 결과도 자동으로 따라온다 — 계측기가 두 벌로 갈라질 여지가 없다.
- `MOAI_AC_BASELINE_REGENERATE=1` 없이는 **아무것도 쓰지 않는다**(기본 off, `TestACRegenerationGateDefaultsOff` 가 단언). CI 가 이 변수를 설정하는 일은 없다.
- 재생성 출력은 `parseACBaseline` 으로 왕복 round-trip 된다(`TestACBaselineEmitterRoundTrip`). 파일 형식이 소비자와 어긋날 수 없다.

## 2. 방아쇠 이벤트 — 이것들이 일어나면 cascade 의무가 생긴다

다음 세 가지는 공통점이 있다: **커밋된 스냅샷의 관측 대상 집단이나 관측값을 움직인다.**

| 이벤트 | 게이트에 나타나는 모양 | 실제 전례 |
|---|---|---|
| `acceptance.md` 의 superseded/split 로 인한 삭제 | `:479` vanish — "present in the snapshot but no longer matched by the corpus glob" (하드 실패) | `20cdeb6bd` — SPEC-MODEL-PROFILE-MATRIX-002 를 네 successor 로 분해하며 삭제 |
| SPEC 디렉터리의 `_archive/` 이동 | 같은 vanish — depth-1 glob 이 더는 그 파일을 매치하지 않는다 | (아직 없음 — 첫 발생이 이 절차의 첫 시험이다) |
| AC 개수에 영향을 주는 corpus 재작성 (B12 절 카운터 문법 변경, corpus glob 변경) | 광범위한 count-move 하드 실패 | t573 (`d9b472409`) — corpus 기준 재작성 후 cascade 누락 |

**새 `acceptance.md` 의 추가는 방아쇠가 아니다.** 부재(absent) 행은 v0.5.0 협정상 "report, not fail" 이고 다음 재생성에서 흡수되는 것이 문서화된 수명주기다. 다만 부재 행은 적체로 불어나기만 하므로(아래 §5), 추가만 있던 기간에도 가끔은 소모성 재생성을 해 주는 것이 보고의 가독성을 지킨다.

## 3. 같은 커밋 규칙 (same-commit rule)

[HARD] **방아쇠 이벤트의 커밋에 재생성된 스냅샷이 같이 들어가야 한다.** 커밋을 나누면 사이에 있는 어떤 HEAD 에서든 게이트가 빨간불이고, 그 빨간불은 §5 에서 봤듯이 실제 결함 신호를 잡는 채널을 오염시킨다.

순서:

1. HEAD 를 다시 읽는다 (`git rev-parse --short HEAD`) — 재생성 헤더에 기록되는 SHA 가 곧 커밋되는 트리여야 한다.
2. §1 의 명령을 실행한다.
3. **diff 를 줄마다 읽는다** (`git diff -- .moai/reports/t338/ac-count-baseline.txt`). §4 의 귀속 술어를 통과할 때까지 커밋하지 않는다.
4. 같은 커밋에 넣는다. 스냅샷 커밋은 **스냅샷만** 운반한다(무관한 변경 혼입 금지).

## 4. diff 검토 + 명명-원인 커밋 규율

[HARD] **이름 붙일 수 없는 diff 행은 정규화하지 않고 멈춘다.** 재생성은 측정이지 승인이 아니므로, 변경된 모든 행의 원인을 커밋 메시지가 명명해야 한다. 숙어:

- **추가 행 (+N)**: 각 행이 구 스냅샷에 없던 SPEC 디렉터리로 귀속되는지 확인한다 — 구 스냅샷의 디렉터리 목록과 겹침이 0이어야 한다. HALT 행이 새로 생겼다면 그것부터 의심한다: 정규화된 것인지(마킹 정리 전) 실제 결함인지 커밋 전에 판정하고 메시지에 명명한다. 절대 조용히 COUNT 로 넘기지 않는다.
- **삭제 행 (−1)**: 어떤 커밋이 그 `acceptance.md` 를 지웠는지 `git log --diff-filter=D -- <path>` 로 답을 내고 그 커밋을 메시지에 적는다.
- **count/state 이동**: 원인을 알 수 있는 이동(그 SPEC 의 실제 본문 편집)만 허용되며, 커밋 메시지가 파일별 원인을 명명한다. 원인을 알 수 없는 이동은 재생성이 아니라 **조사 대상**이다.
- **예상 총수를 미리 못 박지 않는다**: corpus 집단은 `acceptance.md` 가 작성될 때마다 움직인다. 멈춰야 할 판정식은 "숫자가 맞는가"가 아니라 **귀속 술어**(위 세 항)다.

커밋 메시지 형식 모델: 이 SPEC 의 M2 커밋 (`chore(SPEC-AC-BASELINE-REFRESH-001): M2 catch-up cascade`) — 행 클래스별 원인, 귀속 근거(겹침 0·중복 0·HALT 0), 제거 행의 원인 커밋 SHA 를 본문에 적었다.

## 5. 사고 기록 — 이 절차가 없어서 반복된 실패 클래스

- **t348** (`23df21c9e`): 게이트와 최초 스냅샷 생성. 재생성 경로가 git 추적 밖의 임시 스크래치 스크립트였다 — 저장소에 커밋되지 않은 경로.
- **t573** (`d9b472409` → 수리 `5f546af2c`): corpus 기준을 재작성하고 cascade 를 빠뜨렸다. **develop tip 자체에서 게이트가 빨개졌고**, 수리는 손수 재생성으로 했는데 그때 쓴 스크래치 스크립트는 유실됐다.
- **t1068 의 발단** (`20cdeb6bd`): superseded-split 이 `acceptance.md` 를 지우고 cascade 를 빠뜨렸다 — 같은 클래스 두 번째. 방치 비용의 실측: 부재 행은 68(primary@main, 09-21) → 75(t1058 병합 트리, 09-21) → 83(cd99336bf, 09-22) → 84(동일 트리, 본 SPEC plan 산출물이 집단에 편입된 직후) → 91(본 SPEC 실행 트리, 09-22)로 단조 증가했고, 줄일 수 있는 유일한 행위는 재생성뿐이다. 게이트가 benign 한 이유로 빨간 상태인 동안, 게이트가 잡으려는 실제 회귀(count/halt/vanish)는 구별 불가능한 빨간불로 묻힌다.

## 6. 판정 금지선 — 재생성이 바꾸지 않는 것

스냅샷 재생성은 **관측을 다시 찍는 것**이지 판정을 다시 쓰는 것이 아니다. 부재는 report-only 로 남고(필수 출력), 기록된 파일의 vanish / count-move / state-move / halt 식별자 집합 이동은 하드 실패로 남는다 (REQ-ABR-006; 원본 계약 `SPEC-AC-COUNT-DISCRIMINATOR-001` spec.md §3.5 rules 1–4). 자동 흡수는 없다 — 축복 행위는 변수를 걸고 diff 를 읽고 커밋하는 **사람의 검토된 실행**이다.
