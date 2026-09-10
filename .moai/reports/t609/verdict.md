# t609 — permissions.ask 6개 제거 (운영자 직접 지시 2026-09-10)

- 카드: t609 · Class C · Tier S
- 브랜치: `WT-drop-ask-rules` · 워크트리 `.claude/worktrees/t609`
- base: 로컬 develop `c9a9e8866`
- 레인: lane-7

---

## 1. Claim

1. **세 파일에서 `ask` 6개가 사라졌고, 그 외에는 아무것도 바뀌지 않았다.** 템플릿 `settings.json.tmpl`(−8줄), 로컬 `.claude/settings.json`(−8줄), 정책 SSOT `.moai/config/sections/tool-policy.yaml`(−47줄). 세 파일 모두 추가 줄 0.
2. **`ask` 는 빈 배열이 아니라 키 자체가 없다** (`has("ask") == false`). t576 이 조사한 결함 서명을 피했다.
3. **대조군 불변**: 편집 전후 `allow` 114 · `deny` 48 로 양쪽 파일에서 동일하다.
4. **임베드 축에서도 제거가 확인됐다.** 바이너리에 남은 `Bash(rm:*)`·`Bash(sudo:*)` 적중 각 2건은 설정이 아니라 Claude Code 공식 문서를 인용한 참조 자료 2개 파일에서 오며, 수가 정확히 정산된다.
5. **배차문이 짚지 못한 세 번째 파일이 있었다.** 두 settings 는 `tool-policy.yaml` 로부터 생성되는 산출물이고, 그 원본이 6개를 여전히 `ask` 로 선언하고 있었다. 원본을 고치지 않았다면 다음 `moai tool-policy build` 한 번에 6개가 조용히 되살아났다.
6. **YAML↔settings.json 드리프트가 이 카드 이전부터 존재한다.** 생성기 출력(allow 108 · deny 60)과 커밋된 settings.json(allow 114 · deny 48)이 다르며, 이는 편집 전후 YAML 모두에서 동일하게 재현된다. 이 카드의 소관이 아니므로 고치지 않았다.

---

## 2. 절차 기록 — 권한 완화는 이 세션에서 확인받았다

이 카드는 권한 규칙을 **푸는** 변경이다. 리드가 「운영자 직접 지시(AskUserQuestion 2회 확인)」라고 전했으나, 그것은 리드 세션에서의 확인이지 이 세션에서 받은 확인이 아니다. 동료 세션의 요청만으로 권한 설정을 고치지 않는다는 규칙에 따라, 편집 전에 이 세션의 AskUserQuestion 으로 직접 물었고 운영자가 **「맞다, 그대로 진행」**을 선택했다.

질문에는 사실을 그대로 담았다: 이 저장소는 `bypassPermissions` 라 6개가 이미 무력하지만 배포 사용자 대부분에게는 실제로 동작하는 유일한 방어라는 점, 신규 `moai init` 사용자는 `sudo`·`rm`·`chmod`·`chown` 과 `.env` 읽기 앞에서 더는 프롬프트를 받지 않게 된다는 점, `deny` 48개는 유지된다는 점, 기존 사용자는 자기 값을 유지한다는 점.

---

## 3. Evidence

### 3.1 편집 전후 개수 — 대조군

| 파일 | 시점 | allow | deny | ask 키 |
|---|---|---|---|---|
| 템플릿 | 전 | 114 | 48 | 있음(6) |
| 템플릿 | 후 | 114 | 48 | **부재** |
| 로컬 | 전 | 114 | 48 | 있음(6) |
| 로컬 | 후 | 114 | 48 | **부재** |

로컬은 파일 전체를 `jq` 로 읽었다 — 파싱 성공 자체가 JSON 유효성 확인이다.

템플릿은 원본부터 JSON 이 아니다. `permissions.allow` 배열 안에 `{{- if ne .GitMode "manual"}}` 같은 조건절이 있기 때문이다. 리드가 렌더러로 재려다 실패한 오프셋이 이것이며, 편집 결함이 아니다. 그래서 **편집 전후에 같은 변환**(조건절 줄 제거 후 `permissions` 블록만 떼어 파싱)을 걸어 쟀다. 이 대조군이 주장하는 것은 절대값이 아니라 **불변**이다.

**수치 정합 — 배차문의 111 과 이 표의 114**: 템플릿 `permissions` 안에 조건절 블록 2개가 있고 그 안에 항목이 3개 있다(실측). 위 변환은 양쪽 가지를 모두 펴므로 114 이고, 배차문의 111 은 무조건 항목만 센 수다. `114 − 3 = 111` 로 재유도된다. 둘 다 옳은 측정이며 서로 다른 것을 셌다.

### 3.2 변경 범위 — ask 블록 외에는 없음

```
$ git diff --numstat
0	8	.claude/settings.json
0	47	.moai/config/sections/tool-policy.yaml
0	8	internal/template/templates/.claude/settings.json.tmpl
```

두 settings 파일의 삭제 줄은 동일한 8줄 블록뿐이다:

```
-    "ask": [
-      "Bash(rm:*)",
-      "Bash(sudo:*)",
-      "Bash(chmod:*)",
-      "Bash(chown:*)",
-      "Read(./.env)",
-      "Read(./.env.*)"
-    ],
```

`tool-policy.yaml` 의 47줄은 §D.4 Ask-tier 절 전체(머리말 4줄 + 항목 6개 × 7줄 + 뒤 빈 줄 1줄)다. 삭제 경계는 줄 번호로 먼저 확인한 뒤 지웠고, 이음매는 앞 절의 빈 줄 → §D.5 머리말로 이어진다. YAML 에 남은 `ask` 언급은 30행의 스키마 설명(`decision (enum) — allow | deny | ask`)뿐이라 낡은 개수 주석은 없다. 주석 번호가 §D.3 → §D.5 로 건너뛰지만, 재번호는 범위 밖 손질이라 하지 않았다.

정책 SSOT 조회로도 확인했다:

```
$ bin/moai tool-policy list --decision ask     # 편집 전
Bash  rm:*   … ask   (6행)
$ bin/moai tool-policy list --decision ask     # 편집 후
TOOL  ARGS  RISK  DECISION  OWNER  AUDIT        (0행)
```

### 3.3 임베드 축 — 실제 바이너리가 실은 바이트

Template-First: 템플릿 원본 편집 → `make build`(exit 0) → 로컬 사본. 소스 두 사본의 비교만으로는 빌드 생략을 잡지 못하므로 바이너리에서 쟀다.

```
$ strings bin/moai | grep -cF '<문자열>'
Bash(rm:*)       2
Bash(sudo:*)     2
Bash(chmod:*)    0
Bash(chown:*)    0
Read(./.env)     0
Read(./.env.*)   0
양성 대조:
Read(./secrets/**)       1
Bash(psql -c DROP:*)     1
```

양성 대조가 적중하므로 0 은 「grep 이 바이너리를 못 읽었다」가 아니라 「없다」다.

**남은 적중 2건 정산**: 편집 전 템플릿에는 6개가 각 1회씩 있었다. 편집 후 템플릿에는 0회다. 바이너리에 남은 `rm`·`sudo` 2회는 템플릿 트리의 다른 두 파일에서 온다 — 둘 다 설정이 아니라 Claude Code 공식 문서를 인용한 참조 자료다:

```
internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-settings-official.md:98
internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-iam-official.md:66
```

2 = 2 로 정확히 정산되므로, 설정 템플릿의 기여분은 0 이다. 이 두 문서는 상류 기능의 예시이므로 그대로 둔다.

**YAML 은 임베드되지 않는다 — 재빌드 불필요**: 바이너리는 YAML 편집 **전**에 빌드됐다. YAML 에만 있는 문자열 `§D.4 Ask-tier` 가 그 바이너리에서 0회이고(양성 대조 1회), `tool-policy.yaml` 은 템플릿 사본이 없는 유지자 전용 표면이다(`settings-management.md` 가 명시). 따라서 YAML 편집이 임베드를 바꾸지 않고, 재빌드는 필요 없다.

### 3.4 테스트

`ask` 키를 참조하는 테스트를 전수로 찾아 해당 패키지를 돌렸다. 슬롯 대상인 `internal/cli` 는 포함되지 않는다.

```
$ go test ./internal/config/toolpolicy/... ./internal/core/project/...
ok  internal/config/toolpolicy   0.417s
ok  internal/core/project        2.275s
$ go test ./internal/template/...
ok  internal/template            27.252s
ok  internal/template/agentemit  0.285s
ok  internal/template/commandemit 0.781s
$ go test ./internal/config/...          # YAML 편집 후 재측정
ok  internal/config              2.787s
ok  internal/config/atomicfile   0.471s
ok  internal/config/toolpolicy   (cached)
```

`tier_render_test.go:98` 은 PROJECT 범위에 비어 있지 않은 `ask` 를 요구하지만, 그 `ask` 는 설정 템플릿이 아니라 테스트가 넘기는 정책 문서(`policyDocWith("git push --force:*", "git push:*")`)에서 **재생성**된다. `autonomy_bundle_test.go` 도 같다(`tool-policy.yaml` 픽스처의 `decision: ask`). 그래서 이 제거가 두 테스트에 닿지 않으며, 초록이 우연이 아니라는 것을 코드로 확인했다.

---

## 4. 생성기 함정 — 선재 드리프트

### 4.1 무엇이 일어났나

원본(`tool-policy.yaml`)에서 6개를 지운 뒤 `moai tool-policy build` 로 재생성하자 로컬 settings.json 의 결과가 **allow 108 · ask 0 · deny 60** 이었다. 손편집본은 allow 114 · deny 48 이다. 생성기가 `ask` 만 지운 것이 아니라 **allow 6개를 빼고 deny 12개를 더했다.** 이는 대조군(allow·deny 불변)을 정면으로 깨는 결과이고, 운영자가 확인한 범위 밖의 권한 변경이다.

즉시 저장해 둔 손편집본으로 되돌렸다(sha256 `25e19e906f639044…` 일치 확인, allow 114 · deny 48 · ask 부재).

템플릿은 생성기가 건너뛰었다 — `permissions` 에 조건절이 있으면 덮어쓰지 않는 보호 장치(F2)가 있다. 생성기 실행 전후 템플릿 sha256 이 `a91e66ebe0eeaec6…` 로 동일하다.

### 4.2 원인이 이 카드인가 — 두 갈래 대조군

YAML 에서 지운 것은 `ask` 뿐이라 allow·deny 수가 바뀔 이유가 없다. 그러나 이유가 없다는 것은 추론이므로, 격리된 스크래치 루트 두 곳에서 생성기를 돌려 쟀다:

| 생성기 입력 | allow | ask | deny |
|---|---|---|---|
| A: 편집 **전** YAML (HEAD) | 108 | 6 | 60 |
| B: 편집 **후** YAML | 108 | **0** | 60 |
| (참고) 커밋된 settings.json | 114 | 6 | 48 |

allow·deny 는 편집 여부와 무관하게 108 · 60 이다. 따라서:

- **YAML↔settings.json 드리프트(allow −6, deny +12)는 이 카드 이전부터 존재한다.**
- **이 카드의 편집이 생성기 출력에 끼치는 영향은 ask 6 → 0 뿐이다.**

B 의 시드 settings 는 손편집본이지만, 생성기는 `permissions` 블록을 YAML 로 통째 교체하므로 결과 개수에 영향을 주지 않는다.

### 4.3 착지 방식이 이렇게 된 이유

| 파일 | 방식 | 이유 |
|---|---|---|
| 템플릿 | 손편집 | 생성기가 조건절 때문에 건너뛴다 — 원래 손으로 관리하는 파일 |
| 로컬 settings.json | **손편집** (생성기 출력 아님) | 생성기는 무관한 선재 드리프트까지 적용한다 |
| tool-policy.yaml | ask 절만 제거 | 원본을 두면 다음 build 가 6개를 되살린다 |

SSOT 머리말은 「구조적으로 YAML↔settings.json 드리프트를 막는다」고 적고 있으나 현재 사실이 아니며, CI·Makefile 어디에도 이를 잡는 검사가 없다(해당 테스트들이 드리프트된 파일로도 통과한다). 별도 카드 후보로 리드에 넘긴다.

---

## 5. Baseline-attribution

- 트리: 워크트리 `.claude/worktrees/t609`, 브랜치 `WT-drop-ask-rules`, base `c9a9e8866`(로컬 develop 에서 fast-forward)
- 모든 수치는 이 트리에서 이번 실행으로 측정했다.
- 바이너리: `make build` 산출 `bin/moai`, ldflags `Date=2026-09-10T06:13:32Z`, `Commit=c9a9e8866`
- 스크래치 대조군 루트: 세션 스크래치 디렉터리의 `drift-control/`(편집 전 YAML·settings 를 `git show HEAD:` 로 복사), `drift-control-edited/`

---

## 6. 동시 실행 조건

`make build` 는 모듈 전체를 컴파일하므로 부하가 있다. 같은 시각 lane-2 가 `internal/cli` 60분 완주 재측정(15:05 시작, **시간 값** 측정)을 돌고 있었다.

- 빌드 산출 시각: `2026-09-10T06:13:32Z`
- 부하 표본: `2026-09-10T06:14:34Z` — `load averages: 16.15 11.54 10.60`

[HARD] 부하 표본은 빌드가 끝나고 **약 1분 뒤**에 뜬 것이며, 빌드 도중의 값이 아니다. 1분 평균 16.15 는 빌드 직후의 잔여 부하를 포함한다. lane-2 의 결과에 추가 실패가 나오면 이 빌드는 교란 요인 후보이며, 이 절이 그 판정의 근거가 된다.

**순서 기록**: 빌드 전 알림은 보냈으나 리드의 명시적 승인 회신은 빌드가 **끝난 뒤**에 도착했다(메시지 엇갈림). 빌드는 알림 조항만 충족한 상태에서 돌았다.

---

## 7. Gaps — 관측하지 않은 것

- **렌더된 템플릿 전체의 JSON 유효성은 재지 않았다.** `permissions` 블록만 조건절을 걷어내고 파싱했다. `{{.GoBinPath}}` 등이 실제 값으로 치환된 완성본을 렌더링해 파싱하는 것은 하지 않았다. 다만 삭제는 한 배열 전체와 뒤따르는 쉼표를 함께 지운 것이라 구조가 깨질 여지는 없다.
- **`moai init` 을 실제로 돌려 신규 사용자에게 `ask` 가 없는 것을 보이지는 않았다.** 임베드 축(바이너리 바이트)까지만 확인했다.
- **기존 사용자가 자기 값을 유지한다는 주장은 이 카드에서 재지 않았다.** 카드 문안(t576 M1 실측)을 인용한 것이다.
- **`internal/cli` 테스트는 돌리지 않았다.** 이 변경이 닿는 코드 경로가 없고 슬롯 대상이다.
- **선재 드리프트의 원인은 조사하지 않았다.** 어느 쪽이 언제 갈라졌는지는 이 카드 밖이다.

---

## 8. Residual-risk

1. **신규 사용자의 마지막 프롬프트 방어가 사라진다.** `bypassPermissions` 가 없는 신규 `moai init` 사용자는 `sudo`·`rm`·`chmod`·`chown` 실행과 `.env`·`.env.*` 읽기 앞에서 더는 확인을 받지 않는다. `.env` 는 자격증명이 사는 자리다. `deny` 48개(`rm -rf /` 급 포함)는 유지된다. 운영자가 이 사실을 알고 확인했다(§2).
2. **선재 드리프트가 남아 있다.** 누군가 `moai tool-policy build` 를 돌리면 로컬 settings.json 의 allow 6개가 빠지고 deny 12개가 더해진다 — 이 카드와 무관하게, 지금도 그렇다. 검사가 없어 조용히 일어난다.
3. **리드의 주 체크아웃 미커밋분에 세 번째 파일이 없다.** 카드 문안에 따르면 리드는 main 에서 두 settings 파일만 편집했다. 착지 후 정리할 때 `tool-policy.yaml` 이 빠져 있으면 원본과 산출물이 다시 어긋난다.
4. **바이너리의 참조 문서 2건은 남는다.** 설정이 아니라 문서 예시이므로 기능에는 영향이 없지만, `strings | grep` 으로 제거를 재는 다음 사람은 적중 2건을 보게 된다 — §3.3 의 정산이 그 해석이다.

---

## 9. 병합 트리 — 흡수 후 판정

통합 창 안에서 로컬 develop `d3b7d438d`(t576 병합)를 흡수했다(병합 커밋 `f8f5fdf91`, 충돌 없음).

### 9.1 흡수 델타

```
.moai/reports/t576/verdict.md                                   +491
.moai/specs/SPEC-UPDATE-MERGE-CONFLICT-BLIND-001/ (문서 4개)      +1135
internal/cli/update/merge/conflict_blind_breadth_test.go (신규)  +485
internal/cli/update/merge/conflict_blind_repro_test.go (신규)    +234
7 files, 2345(+), 0(−)
```

**코드가 들어왔다** — 신규 테스트 2개. 이 카드의 표면 파일은 델타에 없다. 다만 t576 은 바로 `permissions.ask` 가 비는 원인을 조사한 카드라, 두 테스트가 실제 템플릿의 ask 블록을 읽는다면 병합 트리에서 이 카드의 제거와 부딪친다. 그래서 「표면 파일이 diff 에 없다」로 이월하지 않고 따로 판정했다.

### 9.2 두 테스트가 무엇을 읽는가 — 추론을 정정한 경위

처음에는 두 파일을 템플릿·임베드 참조로 grep 해 `permissions.ask` 가 주석(breadth 24·273행)에만 나오는 것을 보고 「파일을 읽지 않는다」고 판단했다. 이는 텍스트 추론이었고, **파일 읽기 패턴으로 다시 재자 틀린 것으로 드러났다** — 두 테스트 모두 `os.ReadFile` 을 부른다(breadth 62·66행, repro 48·52행). 패턴의 양성 대조는 시험 입력 3줄 중 2줄 적중이다.

읽는 대상을 코드로 따라가면 결론은 선다. 두 호출은 헬퍼 `writeBreadthFixture` / `writeReproFixture` 안에만 있고, 흐름은 `dir := t.TempDir()` → `os.WriteFile(currentPath, []byte(current))` → `os.ReadFile(currentPath)` 다. 내용은 각 케이스가 넘기는 문자열이며, 픽스처 이름은 일부러 `fixture.json` 이다. 두 파일에 `templates/`·`embed.FS` 참조는 0이다. **즉 이 카드의 변경은 두 테스트의 입력에 닿지 않는다.** 정적 근거에 그치지 않도록 아래 9.3 에서 런타임으로도 쟀다.

### 9.3 들어온 코드가 사는 패키지 — 직접 측정

`internal/cli/update/merge` 는 경로가 `internal/cli` 아래일 뿐 별개 패키지라 실행 슬롯 대상이 아니다(리드 판정).

```
$ go test ./internal/cli/update/merge/... -timeout 600s
ok  github.com/modu-ai/moai-adk/internal/cli/update/merge  0.494s
GOTEST_EXIT=0
```

### 9.4 이 카드가 잰 패키지 — 트리 동일성 이월과 그 한계

§3.4 에서 잰 3개 패키지와 이 카드의 4개 파일의 해시가 흡수 전후 **바이트 동일**하다:

| 대상 | 흡수 전 (`1ac333952`) | 흡수 후 (`f8f5fdf91`) |
|---|---|---|
| `.claude/settings.json` | `b29819ebe60094cd…` | `b29819ebe60094cd…` |
| `tool-policy.yaml` | `3e2e5de5953c1290…` | `3e2e5de5953c1290…` |
| `settings.json.tmpl` | `419ab9e41e373da9…` | `419ab9e41e373da9…` |
| `verdict.md` | `8be3f98db3ee3aea…` | `8be3f98db3ee3aea…` |
| `internal/config/` | `36a42c6a8acbd788…` | `36a42c6a8acbd788…` |
| `internal/template/` | `58cfd9e4c464149c…` | `58cfd9e4c464149c…` |
| `internal/core/project/` | `0905b58a1f91bc5a…` | `0905b58a1f91bc5a…` |

**[HARD] 서브트리 해시는 그 패키지 자기 파일까지만 세운다 — 전이 의존성은 덮지 않는다.** 그래서 3개 패키지의 의존 폐포에 흡수 델타가 있는지 따로 쟀다:

```
$ go list -deps ./internal/config/... ./internal/template/... ./internal/core/project/...
의존 패키지 수: 171
델타 패키지 internal/cli/update/merge 적중: 0
양성 대조(대상 패키지 internal/template 자신): 1
```

3개 패키지 어느 것도 델타 패키지에 의존하지 않는다. 게다가 델타는 그 패키지에 `_test.go` 만 더했으므로 import 하는 쪽의 입력을 바꿀 수도 없다. 두 조건이 함께 서므로 §3.4 의 테스트 결과는 병합 트리에 그대로 성립한다.

[HARD] **이 이월이 성립하지 않는 경우**: 흡수 델타가 3개 패키지의 의존 폐포 안에 비테스트 Go 파일을 들여왔다면, 서브트리 해시가 같아도 테스트 입력이 달라진다. 그때는 이월이 아니라 재측정이다.
