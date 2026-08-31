# t371 — verdict (진행 중)

카드: t371 · SPEC-SPECLINT-GITBLIND-001 · 워크트리 `.claude/worktrees/t371` · 브랜치 `WT-lint-shallow-clone`

| 단계 | 상태 | 근거 |
|---|---|---|
| 보존 커밋 | 완료 | `3b637a290` — SPEC 4종 + 증거 6종, 명시 pathspec |
| develop 흡수 | 완료 | `git merge origin/develop`(`9328a5242`) rc=0 충돌 0 → HEAD `35bc0715f` |
| 인용 재측정 | 완료 | `.moai/reports/t371/remeasure-35bc0715f.md` |
| 린트 기준값 재산출 | 완료 | `0 error / 1096 warning` — `.moai/reports/t371/lint-local-35bc0715f.txt` |
| SPEC 인용 갱신 (v0.4.0) | 진행 | manager-spec 위임 |
| v0.4.0 plan-audit (Tier M 문턱 0.80) | 미착수 | — |
| kickoff 승인 | 미착수 | — |

---

## 관측 보존 — cwd 오해로 생긴 경로 (삭제하지 않음)

리드 지시로 **삭제하지 않고 증거로 남긴다**. 경로 자체가 증거이므로 지우면 사라진다.

```
.moai/reports/t371/.moai/state/config-cache.json                                       8612 bytes, 0 newlines
.moai/reports/t371/.moai/state/context-usage/387e58f8-17c6-4ee2-916a-43c271573db7.json  327 bytes
```

둘 다 mtime `2026-08-31 01:55`. `.gitignore:216` 의 `**/.moai/state/` 에 걸려 미추적 목록에도 뜨지 않으므로, 이 항목이 유일한 기록이다.

**정정 하나** — 최초 보고에서 `config-cache.json` 을 "0 bytes" 라 적었으나 틀렸다. `wc -l` 이 낸 0 은 **개행 수**였고(단일 행 JSON, 끝 개행 없음), 실제 크기는 **8612 bytes** 다. 빈 파일이 아니라 온전한 설정 캐시 한 벌이다.

### 무엇이 썼는가

`context-usage` 파일 내용:

```json
{ "schema_version": 2,
  "session_id": "387e58f8-17c6-4ee2-916a-43c271573db7",
  "writer_pid": 33339,
  "captured_at": "2026-08-31T01:55:02.656137+09:00",
  "context_window_size": 1000000, "tokens_used": 220000, "raw_pct": 22,
  "stage": "none", "band": "large", "model": "Opus 5 (1M context)", "effort": "medium" }
```

- **다른 세션이다.** `387e58f8-…` 는 현재 세션(`b2e58561-…`)이 아니다. `/clear` 이전의 t371 세션이 남긴 것이다.
- **statusline 쓰기 경로다.** context-usage 스냅샷은 statusline 이 렌더마다 `<projectDir>/.moai/state/context-usage/<session-id>.json` 에 쓴다. 그것이 `.moai/reports/t371/` 아래에 생겼다는 것은, 그 시점 프로젝트 루트가 **`.moai/reports/t371` 로 해소됐다**는 뜻이다.

### 같은 방향을 가리키는 두 번째 신호

`config-cache.json` 이 담은 값이 이 프로젝트의 설정이 아니다:

```
"User":{"Name":""}   "Language":{"ConversationLanguage":"en","ConversationLanguageName":"English", …}
```

이 프로젝트는 `.moai/config/sections/user.yaml` 이 `GOOS 오라버니~`, `language.yaml` 이 `ko` 다. 캐시에 박힌 것은 **컴파일 기본값**이다 — 그 루트에 `.moai/config/` 가 없어 로더가 설정을 못 찾고 기본값으로 채운 뒤 그 자리에 캐시를 썼다.

두 신호가 같은 결론을 가리킨다: **어떤 명령이 cwd 를 프로젝트 루트로 삼았고, 그 cwd 가 `.moai/reports/t371` 이었다.** 워크트리 루트를 거슬러 올라가 찾는 대신 서 있는 자리를 루트로 읽었다.

### 아직 확인되지 않은 것

- 어느 명령이었는지 특정하지 못했다. 이 디렉터리의 증거 생성 스크립트 `walker-input.sh` 는 `.moai/reports/t371/statusgit-18-ids.txt` 를 **상대 경로로** 읽으므로 워크트리 루트에서 실행되도록 쓰였고 자체적으로 `cd` 하지 않는다 — 이 스크립트 자신은 용의자가 아니다.
- statusline 이 그 cwd 를 어디서 받았는지(훅 입력의 `cwd` 필드인지, `os.Getwd()` 폴백인지)는 미확인. `input.CWD` 빈 값 → `os.Getwd()` 폴백은 알려진 형태이나, **이 사건이 그 형태라고 단정할 근거는 아직 없다.** 가설이지 관측이 아니다.

후속 카드 재료로 남긴다.

---

## iter-3 감사 결과 — FAIL 0.79 (Tier M 문턱 0.80)

판정문: `.moai/reports/t371/plan-audit-iter-3.md`. 감사 트리 `bff12c54d`.

차원: Clarity 0.85 · Completeness 0.80 · **Testability 0.68** · Traceability 0.90 → 조화평균 0.7987.
MUST-PASS 7종 실패 없음 — FAIL 은 점수와 결함 목록이 만든 판정이다.

### 차단 3건

| ID | 내용 |
|---|---|
| D1 | AC-SLGB-001/004 가 지목한 픽스처 선례가 RED 기준선을 깬다 — 그 픽스처는 `MissingExclusions` 경고를 내어 `✓ No findings` 줄이 애초에 안 찍히고, 두 AC 의 "그 줄 부재" 단언이 공허해진다. 도달 불가는 아니다(Out of Scope 절을 넣으면 찍힌다) |
| D2 | `verification-completeness.md` §2 의 2-cell 채택 규율을 11개 AC 중 **0건**이 만족. RED-now 셀·green-path 셀·트리 핀 전부 부재 |
| D3 | **iter-3 의 "인용 좌표 전량 재측정" 주장이 거짓** — `plan.md:170` 괄호 안 3개가 스테일. `plan.md:16` 도 흡수 전 기준점 |

### D3 귀속 — 왜 두 겹의 검증이 함께 놓쳤나

오케스트레이터의 V2 검증과 manager-spec 의 사후 스윕이 **같은 형태의 패턴**을 썼다:

```
오케스트레이터 V2:   grep -rn 'lint\.go:1299-1301' …   → rc=1 (무매치)
manager-spec 스윕:   grep -rn 'spec/lint\.go:' …       → 잔여 0 보고
접두사 없이 재검색:  grep -rn ':1299-1301' …           → plan.md:170 적발
```

둘 다 **파일 접두사에 앵커**했는데, `plan.md:170` 의 좌표는 앞 줄이 파일명을 이미 밝힌 뒤의 **연속 표기**라 접두사가 없다:

```
169:- `internal/spec/lint.go:1306-1342` — `StatusGitConsistencyRule.Check`
170:  (terminal 조기반환 `:1299-1301`, err skip `:1305-1308`, emission `:1310-1319`)
```

접두사 앵커 스윕은 이 형태를 **원리상 볼 수 없다**. 게다가 앞의 둘은 바로 윗줄이 선언한 부모 범위 `:1306-1342` **바깥**이라, 트리를 열지 않고도 모순이 보이는 값이었다 — 범위 포함 관계라는 더 싼 검사가 있었는데 쓰지 않았다.

실측 정답: `:1318-1320` / `:1324-1327` / `:1331-1340`. `plan.md` §F M1.4 는 이미 옳은 값을 써서 파일이 자기모순 상태다.

교훈: 인용 갱신 스윕은 접두사가 아니라 **좌표 모양**(`` `:NNNN` `` / `` `:NNNN-NNNN` ``)으로 훑고, 부모 범위와의 포함 관계를 함께 본다.

### iter-1 차단 5건 개별 판정

D1 닫힘 · D2 닫힘 · D4 닫힘 · D7 닫힘 · **D8 부분 닫힘** — 지목된 두 사례는 닫혔으나 결함 부류는 열려 있다(관측된 red 기록 전무, 같은 형태의 새 사례를 AC-001/004 에서 실측).

### 절차 — 운영자 판단 필요

`harness.yaml` 의 `plan_audit_tier_ceilings` 는 Tier M = **2**. 이번이 iter-3 라 이미 Tier 상한 초과이며, 수정 후 재감사(iter-4)는 재시도 계약의 **명시적 사용자 override** 분기에 해당한다. 레인은 이 게이트를 열 수 없어 리드에 보고했다.

점수 회귀는 없다(0.75 → 0.79). 문턱이 0.75 → 0.80 으로 멀어졌을 뿐이다.

---

## 보류 — 운영자 판정 (t376 → t382 착지 후 재개)

> ### [HARD] 병합 금지 — 현재 HEAD 는 귀속 분열 상태다
>
> **재개 3단계(`plan.md` 좌표 재측정 + `:16` 라벨 갱신)를 마치기 전에는 이 브랜치를 통합 창에 넣지 않는다.** 5단계를 돌리다 앞질러 병합하는 것을 막는 표식이다.
>
> 분열은 **3겹**이다 (실측 — `git diff bff12c54d^ bff12c54d -- plan.md`, `git log -1 35bc0715f`):
>
> | 대상 | 어느 트리 값인가 | 근거 |
> |---|---|---|
> | `plan.md:16` 라벨 `1e5199b88` | 흡수 **이전** | 흡수 커밋 `35bc0715f` 의 2번 부모가 `9328a5242` |
> | `plan.md:101` `:1324-1327` · `:169` `:1306-1342` | 흡수 **이후** | `bff12c54d` 가 이 둘만 갱신 |
> | `plan.md:170` × 3 (`:1299-1301` / `:1305-1308` / `:1310-1319`) | 흡수 **이전** (D3) | `bff12c54d` 가 이 줄은 건드리지 않음 |
>
> 두 가지가 여기서 따라 나온다.
>
> 1. **라벨을 어느 값으로 정해도 지금은 틀린다.** `1e5199b88` 은 `:170` 3건에는 맞고 `:101`·`:169` 에는 틀리며, `9328a5242` 는 그 반대다. 좌표를 **전부** 한 트리로 맞춘 뒤에 라벨을 찍는 것 외에 해소 경로가 없다.
> 2. **같은 사실이 plan.md 안에 두 번 적혀 있고 한쪽만 갱신됐다.** `:101` 의 `:1324-1327`(err skip, 갱신됨)과 `:170` 의 `:1305-1308`(같은 err skip, 스테일)이 같은 코드 위치를 가리킨다. 재개 시 `:170` 을 고치면 두 값이 일치해야 하며, 일치하지 않으면 재측정이 틀린 것이다 — 공짜로 얻는 교차검증이다.


리드 지명 대기. **문서는 지금 고치지 않는다** — t376(rule 등록부 `:137`)과 t382(`era.go` + `lint.go:272-275`)가 착지하면 `lint.go` 좌표가 다시 밀린다. 특히 D2 의 2-cell RED 셀에 적을 트리 SHA 는 그 뒤라야 최종값이 되며, 지금 채우면 **스테일한 트리를 가리키는 RED 셀** — D1 이 지적한 바로 그 형태 — 을 새로 만든다.

보존 상태: 워크트리 `.claude/worktrees/t371`, 브랜치 `WT-lint-shallow-clone`, 보존 커밋 `3b637a290`, 흡수 `35bc0715f`, 인용 갱신 `bff12c54d`. 통합 창 미획득.

### 재개 시 처리할 수정 목록 (확정)

**D3 — 좌표 (기계적)**

| 위치 | 현재 | 정정 |
|---|---|---|
| `plan.md:170` | `:1299-1301` | `:1318-1320` |
| `plan.md:170` | `:1305-1308` | `:1324-1327` |
| `plan.md:170` | `:1310-1319` | `:1331-1340` |
| `plan.md:16` | `develop 1e5199b88` | 재개 시점 흡수 HEAD |

재개 시 위 값도 **다시 재야 한다** — t376/t382 착지가 `lint.go` 를 또 밀기 때문에 위 표는 정정 대상의 목록이지 최종값이 아니다.

#### 좌표 분류 — 갱신 대상인가, 고정 대상인가 (재개 전 확정)

리드 정정(전방 인용 → 갱신 / 역사 앵커 → base SHA 고정, VCI §2.1 R1)을 이 카드의 4행에 적용한 결과.
**이 절은 분류만 기록한다 — 문서는 지금 고치지 않는다.**

| 위치 | 분류 | 처리 | 근거 |
|---|---|---|---|
| `plan.md:170` × 3 (`:1299-1301` / `:1305-1308` / `:1310-1319`) | **전방 인용** | 갱신 | §H 는 「상호참조」 절이고, 세 값은 이 카드가 **고칠 대상 코드**의 구조(terminal 조기반환 · err skip · emission)를 짚는 길잡이다. "고치기 전에 이랬다"는 관측을 담고 있지 않다 |
| `plan.md:16` `develop 1e5199b88` | **역사 아님 — baseline 귀속 라벨** | 좌표와 함께 이동 | 아래 별항 |
| `plan.md:17` `사전 A/B 측정 … b9149857c` | **역사 앵커** | 고정 (D3 표에 없음, 접촉 금지) | 그 시점 측정이 어디서 수행됐는지의 기록. 갱신하면 관측이 사라진다 |

**`plan.md:16` 에 대한 이견 — 역사로 고정하면 안 된다.**
그 줄은 "기준 트리: … develop `1e5199b88` 기준" 이다. 이것은 사건의 기록이 아니라 **이 문서의 전방 좌표들이 어느 트리에서 재는가**를 선언하는 귀속 라벨이다. 좌표는 흡수(`35bc0715f`)에서 이미 갱신됐으므로, 라벨만 `1e5199b88` 에 고정하면 **라벨이 가리키는 트리와 좌표가 사는 트리가 갈린다** — VCI §2 의 미귀속 주장이 된다. 재개 시 이 줄은 좌표를 재는 그 트리로 함께 옮긴다.

> **정정 — 처음 쓴 판별식은 순환이었다.** 최초 서술은 「`:17` 만 D3 표에 없다」는 대비를 근거로 삼았는데, 표 등재 여부는 앞선 분류의 *결과*이지 근거가 아니다. 리드 지적을 받아 아래 3분·판별식으로 대체한다.

**3분 분류와 판별식 (리드 채택 — 다음 흡수 때 재논쟁 금지)**

판별식은 **그 줄이 지금 무엇을 주장하는가** 하나다. 시제가 그 신호를 겉으로 드러낸다.

| 종류 | 그 줄이 지금 주장하는 것 | 시제 신호 | 처리 |
|---|---|---|---|
| **전방 좌표** | "그 코드는 지금 여기에 있다" | 현재 | 트리가 움직이면 재측정·갱신 |
| **귀속 라벨** | "이 문서의 좌표가 지금 어느 트리에서 유효한가" | 시제 없는 선언 | 좌표와 **함께** 이동. 좌표만 옮기고 라벨을 두면 미귀속 |
| **역사 기록** | "그때 그 트리에서 이랬다 / 이렇게 쟀다" | 과거 | base SHA 고정. 갱신하면 관측이 사라짐 |

적용: `plan.md:16` "기준 트리: … develop `1e5199b88` 기준" = 시제 없는 선언 → **귀속 라벨**. `plan.md:17` "사전 A/B 측정은 develop `b9149857c` 에서 **수행됐다**" = 과거 시제 → **역사 기록**.

**[HARD] 라벨은 최종 커밋 직전에 다시 읽는다.** 지금 아는 값(`9328a5242`)을 미리 적으면 병합 창에서 develop 을 한 번 더 흡수할 때 또 썩는다. 커밋 직전 `git rev-parse origin/develop` 재판독값으로 맞춘다.

**전방 분류의 근거 — 현재 트리에서 실측 (`35bc0715f`)**
```
grep -n 'func (r \*StatusGitConsistencyRule) Check' internal/spec/lint.go
→ 1306:func (r *StatusGitConsistencyRule) Check(...)
sed -n '1316,1342p' internal/spec/lint.go
→ terminal 조기반환 1318-1320 · err skip 1324-1327 · emission 1331-1340
```
부모 범위 `:1306-1342` 는 현재 트리에서 정확하고, D3 표의 정정값 3개는 전부 그 안에 든다. (t376/t382 착지 후 재측정 의무는 위 문단 그대로 유효)

**재개 순서 (리드 확정 — 이 순서대로 기계적으로 처리)**

1. t376 · t382 착지 확인 — `git log origin/develop` 조상 판정
2. develop 흡수
3. `plan.md` D3 좌표 재측정 + `:16` 라벨을 **커밋 직전 `git rev-parse origin/develop` 재판독값**으로
4. 세 축 스윕 재실행 — **이때의 증감이 판정 근거**
5. 완료 보고

> **보류 중에 돈 스윕의 증감 0 은 근거가 아니다.** `plan.md` 를 안 고쳤으니 안 변한 것이 당연하다. 실효 스윕은 3번 편집 직후의 4번뿐이며, 보류 중 값(백틱 36 · 평문 0 · `file.go:N` 48)은 그때의 대조 기준선으로만 쓴다.

**세 축 스윕 — 이 SPEC 에서의 현재 상태 (`35bc0715f`)**
```
grep -rnE '`:[0-9]+(-[0-9]+)?`' .moai/specs/SPEC-SPECLINT-GITBLIND-001/          → 36건
grep -rnoE '\b(line|lines|L)[ ]?[0-9]+(-[0-9]+)?' .moai/specs/SPEC-SPECLINT-GITBLIND-001/ → 0건
grep -rnoE '[a-z_]+\.go:[0-9]+(-[0-9]+)?' .moai/specs/SPEC-SPECLINT-GITBLIND-001/         → 48건
```
lane-1 이 지적한 평문 축(`line 28` / `L28`)은 이 SPEC 에 **0건**이다 — 가정이 아니라 위 명령의 출력이다. 재개 시 편집 후 세 축 모두 재실행한다.

**손대지 않을 것 (역사 앵커, 재개 시에도 고정)**
- `.moai/reports/t371/classification-18.md` 의 18건 — 수정 이전 코퍼스의 집합
- D1 재현(`noscope` → `✓ No findings` 미발화) — 수정 이전 상태의 관측
- `plan-audit-iter-1/3.md` 의 좌표 — 그 시점 트리를 가리키는 판정 기록

스윕 방법 [HARD]: 파일 접두사(`lint\.go:`)가 아니라 **좌표 모양**으로 훑는다.
```
grep -rnE '`:[0-9]+(-[0-9]+)?`' .moai/specs/SPEC-SPECLINT-GITBLIND-001/
```
그리고 각 좌표가 같은 항목이 선언한 **부모 범위 안에 드는지** 확인한다. `plan.md:170` 의 두 값은 윗줄 부모 범위 `:1306-1342` 밖이었으므로, 이 검사만으로 트리를 열지 않고 잡혔을 결함이다.

**D1 — AC-SLGB-001 / 004 에 픽스처 조건 명시**

RED 기준선이 성립하려면 픽스처가 `MissingExclusions` 를 내지 않아야 한다. 두 AC 에 "픽스처는 Out of Scope 절을 포함한다"는 전제를 명시하면 닫힌다. 도달 불가가 아니다(아래 재현 B).

**D2 — 11개 AC 전부 2-cell 채택**

`verification-completeness.md` §2 대로 RED-now 셀(명령 + 원문 출력 + exit code + 트리 SHA, 그리고 **왜 red 인지**) + green-path 셀(어느 마일스톤이 뒤집고 통과 출력이 무엇이 되는지). 좌표가 최종값이 된 뒤에 채운다.

### D1 재현 — 이 세션이 직접 실행

측정 트리 `bff12c54d`, 판정 도구 트리 빌드 `./bin/moai`. 픽스처는 `.moai/reports/t371/repro/{noscope,withscope}/spec.md` (두 파일은 `## 4. Scope` / `### 4.1 Out of Scope` 절 유무만 다르다).

```
$ ./bin/moai spec lint .moai/reports/t371/repro/noscope/spec.md
SEVERITY  CODE               FILE                                      LINE  MESSAGE
--------  ----               ----                                      ----  -------
WARNING   MissingExclusions  .moai/reports/t371/repro/noscope/spec.md  1     'Out of Scope' section missing — minimum one item in Out of Scope subsection required [grandfathered era — downgraded to warning]

0 error(s), 1 warning(s)
rc=0

$ ./bin/moai spec lint .moai/reports/t371/repro/withscope/spec.md
✓ No findings — all SPEC documents are valid
rc=0
```

**A**: `✓ No findings` 줄이 찍히지 않는다 → AC-SLGB-001/004 의 "그 줄 부재" 단언이 구현과 무관하게 성립 → 공허.
**B**: 같은 픽스처에 Out of Scope 절만 넣으면 그 줄이 찍힌다 → 도달 가능. AC 를 픽스처 조건까지 명시하도록 고치면 닫힌다.

부수 관측 — A 의 메시지 꼬리 `[grandfathered era — downgraded to warning]`. 오늘 날짜(`created: 2026-08-31`)로 만든 픽스처가 grandfathered 로 분류됐다. 다만 이것은 **t382 가 지목한 H-3 이 아니라 H-1** 이다(이 픽스처에는 `progress.md` 자체가 없어 H-1 이 먼저 매치한다). 형태가 닮았다고 t382 사례로 세지 않는다.

### 재현하지 못한 것 — 감사자 귀속으로 남긴다

감사자가 관측한 *"`main` ref 없는 저장소에서 `StatusGitConsistency` 가 침묵하고 도구가 전 문서 유효를 선언했다"* 는 이 세션에서 재현하지 못했다. `cachedMainBranch` 가 **cwd 의존**이라 `main` 이 없는 저장소를 cwd 로 삼아야 하는데, 이 세션은 워크트리 격리 가드가 cwd 변경(`cd` / 서브셸 / `env -C`)을 거부한다. 위 재현 A/B 는 `main` 이 존재하는 이 워크트리에서 돌았으므로 그 조건을 만들지 못한다.

따라서 이 관측은 **감사자 귀속**이며 이 verdict 의 자체 측정이 아니다. 재개 시 in-package 테스트 시드(`drift_characterization_test.go:55` `chdirForTest`)로 재현하는 것이 옳은 경로이고, 그것이 이 SPEC 의 M2 가 세우려는 seam 과 같다.

### 순환 해소 — 리드 판정 (재개 시 재론 금지)

앞 절이 "관측을 근거로 쓰려면 M2 를 먼저 세워야 하는 순환"이라 적었으나, 순환이 아니다. **M2 가 세우려는 seam 이 곧 그 관측을 측정 가능하게 만드는 장치**이므로, 그 관측은 M2 의 **입력이 아니라 산출**이다.

SPEC 의 논지는 코드 판독만으로 이미 선다 — `cachedMainBranch` 가 cwd 의존이고(`gitquery_cache.go:89`, fallback `:103`/`:114`), `main` 부재 시 `getGitImpliedStatus` 가 실패해 `StatusGitConsistencyRule.Check` 가 skip 한다(`lint.go:1324-1327`). 실행 관측은 그 논지의 전제가 아니라 **M2 가 성립했음을 보이는 증거**다.

재개 시 순서:
```
M1/M2 로 seam 을 세운다
  → chdirForTest 시드로 main 없는 저장소를 만든다
  → StatusGitConsistency 침묵을 관측한다        ← M2 의 AC 증거
```

[HARD] 이 관측을 M1 이전으로 끌어올리지 않는다. 끌어올리면 seam 없이 관측할 방법을 찾느라 워크트리 격리 가드와 싸우게 되고, 그것은 이 카드가 풀 문제가 아니다.
