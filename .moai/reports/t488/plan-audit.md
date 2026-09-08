# SPEC 감사 판정서 — SPEC-PREMERGE-SETTINGS-DRIFT-001

카드 t488 · 반복 1/3 · Tier M(PASS 임계 0.80)
감사 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t488` · 브랜치 `WT-premerge-drift-assert` · HEAD `3084f1071`
감사자: plan-auditor (독립 감사, M1 문맥 격리 적용)

**판정: FAIL** · 종합 점수 **0.725** (임계 0.80 미달)
FAIL의 직접 원인은 두 가지가 독립적으로 성립한다 — MP-7 미해결 `[NEEDS CLARIFICATION]` 2건(점수와 무관한 관문), 그리고 아래 blocking 결함 8건.

작성자 추론 문맥은 받지 않았고 받았어도 무시한다(M1 Context Isolation). 판단 근거는 이 트리의 아티팩트 4개와 내가 직접 읽은 파일뿐이다.

---

## §0 사전 확인

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t488
$ git branch --show-current
WT-premerge-drift-assert
```

워크트리 이름이 `t488`임을 확인했다. 아티팩트 4종(spec.md 11391B / plan.md 15491B / acceptance.md 6761B / progress.md 806B) 전부 읽었다.

---

## §1 Must-Pass 결과

| 항목 | 판정 | 근거 |
|---|---|---|
| MP-1 REQ 번호 일관성 | PASS | `grep -o 'REQ-PSD-[0-9]\{3\}' spec.md \| sort -u` → 001..013 연속, 결번·중복 0. `grep -c '^- \*\*REQ-PSD-' spec.md` → `13` (정의 줄 수 = 고유 id 수) |
| MP-2 GEARS 형식 준수 | PASS | 요구층(`REQ-PSD-*`) 13개 전부 GEARS 5패턴 중 하나에 대응. 판정은 요구층에 대해서만 내렸고, `AC-PSD-*`의 Given-When-Then은 검증층 정격이므로 여기서 감점하지 않았다. 약한 항목 3건은 §6에 optional로 분리 |
| MP-3 프론트매터 유효성 | PASS | 12개 정본 필드 전부 존재(`spec.md:2-14`). 거부 별칭(`created_at`/`updated_at`/`labels`/`spec_id`) 0건. `tier: M`·`related_specs` 추가 필드는 허용 범위 |
| MP-4 언어 중립성 | N/A | 이 저장소 자체(Go)를 대상으로 한 단일 언어 SPEC. 16개 프로그래밍 언어 표면을 건드리지 않는다 |
| MP-5 D7 교차 SPEC | PASS | 참조 2건 모두 실재하고 어느 것도 retired/superseded/archived가 아니다. `SPEC-SETTINGS-ORIGIN-001` → `status: in-progress`, `SPEC-GIT-STATUS-FIXTURE-001` → `status: completed` |
| MP-6 D8 크로스 플랫폼 | PASS | `grep -c 'syscall' spec.md plan.md acceptance.md` → 전부 `0`. 검사 대상 없음 |
| **MP-7 clarification 관문** | **FAIL** | `plan.md:50`, `plan.md:52`에 미해결 `[NEEDS CLARIFICATION:` 2건. 점수와 무관하게 FAIL을 강제하며, Implementation Kickoff Approval 전에 `AskUserQuestion`으로 해소되어야 한다 |

`spec.md:105`의 `[NEEDS CLARIFICATION]`은 **마커가 아니라 마커 규약을 서술하는 문장**이다(대괄호 안에 topic이 없다). 관문 대상으로 세지 않았다.

**린트 확인(직접 실행).** `moai spec lint`가 지원하는 인자 형태는 파일 경로다(`--help`: `moai spec lint [spec.md...]`). 경로 인자형으로 직접 돌렸고 clean이다.

```
$ moai spec lint .moai/specs/SPEC-PREMERGE-SETTINGS-DRIFT-001/spec.md
✓ No findings — all SPEC documents are valid
```

이 결과는 MP-2 판정의 보조 근거이지 단독 근거가 아니다. 린트는 GEARS 패턴 대응을 기계적으로 훑을 뿐이므로, 요구 13개는 내가 하나씩 읽고 따로 판정했다(약한 항목은 §6에 optional로 분리).

---

## §2 차원별 점수

| 차원 | 점수 | 루브릭 대역 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 인용 규율이 이 저장소 평균보다 높다(코드 주장 전부에 경로+줄번호). 다만 `REQ-PSD-011`/`-012`는 행위 주체가 없는 수동태이고, `AC-PSD-006`의 "판정 지점을 훑으면"은 측정 방법이 없으며, `AC-PSD-005`는 관측 지점 자체가 정의되지 않았다(D2) |
| Completeness | 0.75 | 0.75 | 필수 절 전부 존재, Out of Scope 5개 H3 + 불릿. 그러나 acceptance.md `경계 사례` 22-26줄에 **행동 규칙 2건이 요구·수용 어느 층에도 없이** 산문으로만 있다(D6) |
| Testability | 0.70 | 0.50-0.75 사이, 하향 | 대부분의 AC는 일치 줄 수·바이트·해시로 이분 판정된다. 그러나 `AC-PSD-006`은 판단형이고, `AC-PSD-005`는 공허하게 초록일 수 있으며, `AC-PSD-007(c)`는 D3 파일명 규칙과 충돌해 자기 모순이다(D2/D4) |
| Traceability | 0.70 | 0.50-0.75 사이, 하향 | REQ 13개 전부 최소 1개 AC에 매핑되고 고아 AC는 0이다. 그러나 `REQ-PSD-007`은 **매핑된 AC가 그 요구를 위반해도 통과한다** — 명목상 커버, 실질 미검증(D1) |

종합 = (0.75 + 0.75 + 0.70 + 0.70) / 4 = **0.725**. Tier M 임계 0.80 미달.

---

## §3 발견된 결함 (blocking)

### D1. `REQ-PSD-007`을 반증하는 AC가 없다 — 명목 커버, 실질 미검증

- 위치: `spec.md:66`(REQ), `spec.md:84`(매핑 선언), `acceptance.md:15`(AC-PSD-007)
- 심각도: **critical** · 분류: **blocking**

`REQ-PSD-007`은 "보존 사본을 자동으로 커밋하거나 push하지 않는다"다. 이것이 매핑된 `AC-PSD-007`이 단정하는 것은 세 가지뿐이다 — (a) 보존 사본이 원본과 바이트 일치, (b) `ledger.jsonl`에 정확히 한 줄, (c) 두 번 적중 시 사본 두 개.

**셋 중 어느 것도 커밋·push의 부재를 단정하지 않는다.** 보존 사본을 만든 **뒤에 커밋까지 하는** 구현은 (a)(b)(c)를 모두 통과한다. 즉 `REQ-PSD-007`은 위반해도 빨간불이 켜지지 않는다.

`.moai/state/`가 gitignore 대상이라는 사실(`.gitignore:224` `**/.moai/state/`, `:311` `.moai/state/` — 내가 직접 확인했다)이 완화 요인이지만 방어가 아니다. AC가 gitignore 상태를 단정하지도 않고, push는 gitignore와 무관하다.

**필요한 수정**: `AC-PSD-007`에 (d)를 추가한다 — 게이트 실행 전후로 `git log --oneline -1` 출력이 동일하고, 보존 디렉터리가 `git status --porcelain --ignored`에서 ignored로 분류되며, 게이트가 실행한 명령 목록에 `commit`/`push`가 0건임을 기록된 argv로 단정한다.

### D2. `AC-PSD-005`에 관측 지점이 없다 — 뮤턴트 5가 공허하게 초록일 수 있다

- 위치: `acceptance.md:13`, `plan.md:99`(M1 서명), `plan.md:119`(M2 표 마지막 행)
- 심각도: **critical** · 분류: **blocking**

브리프가 특히 엄하게 보라고 한 항목이다. 결론부터: **argv 문자열 단정 자체는 반증 가능하다.** 기대 토큰열을 테스트에 리터럴로 적어 두면 구현에서 `--no-optional-locks`를 빼는 순간 RED가 된다. 구현을 되풀이해 적는 동어반복이 아니다 — 단, **조건이 하나 붙는데 SPEC이 그 조건을 적지 않았다.**

조건: *단정하는 argv가 실제로 실행된 argv여야 한다.* `plan.md:99`가 정한 M1 서명은 `(matchCount int, rawOutput string, err error)`이고 **argv를 밖으로 내보내지 않는다.** 그러면 테스트는 argv를 어디서 보는가? SPEC은 답하지 않는다. 남는 경로는 둘뿐이고, 한쪽은 함정이다.

- (가) 별도의 argv 빌더 함수를 노출하고 테스트가 그것을 호출해 단정 — 구현이 실행 지점에서 `exec.Command("git", …)`를 따로 적어 두면 **아무도 실행하지 않는 함수를 고정하는 꼴**이 된다. 빌더에서 플래그를 빼도 실행 경로는 멀쩡하고, 반대로 실행 경로에서 빼도 빌더가 멀쩡해 테스트는 초록이다. 후자가 정확히 t477이 관측한 "틀린 모양이 통과하는" 공허한 매치다.
- (나) 실행 경계에 러너를 주입해 **실제로 넘어간 argv를 기록**하고 그것을 단정 — 이쪽만 뮤턴트 5를 진짜로 잡는다.

SPEC은 (나)를 요구하지 않으므로 run 단계가 (가)로 갈 여지를 열어 둔 채다. 이 카드 전체가 "게이트가 조용히 무의미해지는 것"을 막으려는 카드인데, 다섯 뮤턴트 중 하나의 유일한 포착기가 그 결함 모양을 그대로 갖고 있다.

**필요한 수정**: `REQ-PSD-013` 또는 `AC-PSD-005`에 단일 원천 조항을 넣는다 — argv는 정확히 한 곳에서 구성되고, 실행기는 그 구성 결과를 그대로 받으며, `AC-PSD-005`가 단정하는 문자열은 **실행 경계에서 기록된 argv**여야 한다(빌더 반환값을 직접 부르는 형태를 금지한다).

### D3. `AC-PSD-005`의 "행동으로는 원리적으로 검출 불가" 주장이 과하다

- 위치: `acceptance.md:13`, `plan.md:121`
- 심각도: **major** · 분류: **blocking** (근거가 틀리면 D2의 설계 선택 근거가 흔들린다)

두 문서 모두 "플래그를 빼도 출력과 반환값이 동일하므로 행동 기반 검증으로는 **원리적으로** 검출되지 않는다"고 적었다. 출력이 같다는 부분은 맞다. 그러나 플래그를 빼는 것의 **요점 자체가 부작용의 차이**다 — `--no-optional-locks` 없는 `git status`는 인덱스를 쓸 수 있고, 그것이 카드 t485가 실측해 이 카드의 근거가 된 바로 그 현상이다(`spec.md:56`이 스스로 인용한다). 부작용이 다르면 관측 가능성이 원리적으로 배제되지 않는다: 픽스처의 `.git/index` mtime/inode를 술어 실행 전후로 재는 검증이 성립할 수 있다.

다만 나는 그것이 **신뢰할 만하다고는 말하지 않는다.** `git status`가 인덱스를 실제로 다시 쓰는지는 stat 캐시가 낡았는지에 달려 있어, 갓 만든 픽스처에서는 쓸 수도 쓰지 않을 수도 있다. 즉 정확한 서술은 "원리적으로 불가"가 아니라 **"신뢰성 있게는 불가"**다.

이 정정은 argv 단정을 폐기하자는 뜻이 아니다 — argv 단정은 그대로 두는 게 맞다. 틀린 것은 그 선택을 정당화하는 문장이고, 검증되지 않은 전제를 근거로 적은 것은 그 자체로 결함이다.

**필요한 수정**: 두 문장을 "행동만으로는 신뢰성 있게 구별되지 않는다(인덱스 쓰기 여부는 stat 캐시 상태에 좌우된다). 그래서 argv를 단정한다"로 바꾼다.

### D4. `AC-PSD-007(c)`가 D3 파일명 규칙과 충돌한다 — 같은 초 안에서 자기 모순

- 위치: `plan.md:58`(파일명 규칙), `acceptance.md:15`(c)
- 심각도: **major** · 분류: **blocking**

파일명은 `settings.json.<card>.<UTC RFC3339 압축형>.<md5 앞 8자>`이고 예시가 `settings.json.t488.20260906T031500Z.4f455d94` — **초 해상도**다. `AC-PSD-007(c)`는 "같은 카드로 두 번 적중시키면 보존 파일이 두 개가 되어 첫 사본이 덮이지 않는다"를 요구한다.

같은 카드 · 같은 내용(→ 같은 md5) · 같은 초에 두 번 적중하면 파일명 세 성분이 전부 같아 **한 파일로 겹친다**. 그리고 이 조건은 이론적 극단이 아니다 — AC-PSD-007(c)를 검증하는 테스트가 게이트를 연달아 두 번 부르면 대개 같은 초 안에 들어간다. 즉 이 AC는 **자기 자신을 실패시키거나 간헐적으로 실패시키기 가장 쉬운 형태**로 쓰여 있다.

**필요한 수정**: 둘 중 하나를 고른다 — (가) 시각 성분을 밀리초 이상으로 올린다, (나) 충돌 시 접미 카운터를 붙이는 규칙을 D3에 명시한다. 어느 쪽이든 규칙을 정한 뒤 AC를 그 규칙에 맞춰 다시 쓴다. "덮어쓰지 않는다"는 의도는 옳다(운영자가 t334를 대조 표본으로 남긴 것과 같은 이유). 결함은 의도가 아니라 이름 규칙이다.

### D5. `plan.md` M2의 뮤턴트 2 포착기 귀속이 acceptance.md와 어긋나고, 실제로 잡지 못한다

- 위치: `plan.md:116`(M2 표 "판정 반전 → F1"), `acceptance.md:17`(AC-PSD-009)
- 심각도: **major** · 분류: **blocking**

`plan.md` M2 표는 "판정 반전" 뮤턴트를 **F1(clean 픽스처)**이 잡는다고 적었다. 그런데 F1이 단정하는 것은 `matchCount == 0`이고, 이는 **술어의 반환값**이다. 판정 반전은 술어가 아니라 **게이트**에서 일어난다(`drift = matchCount > 0`을 뒤집는 것). 게이트를 뒤집어도 F1의 `matchCount`는 그대로 0이므로 **F1은 초록으로 남는다.**

실제 포착기는 `AC-PSD-009`이고, acceptance.md는 그것을 정확히 적어 두었다("판정 반전 뮤턴트는 여기서 lock 파일을 만들어 FAIL이 된다"). 두 문서가 같은 뮤턴트에 대해 서로 다른 포착기를 지목하고 있고, 그중 plan.md 쪽이 틀렸다.

이것이 blocking인 이유는 run 단계의 산출물 때문이다. `acceptance.md:32`는 뮤턴트 5종 각각에 대해 **RED를 실제로 관측한 출력을 판정서에 인용**하라고 요구한다. 구현자가 plan.md의 표를 따라 F1으로 뮤턴트 2를 뒤집어 보면 초록이 나오고, 그때 "잡히지 않는다"고 결론짓거나 — 더 나쁘게는 — 다른 무언가를 조정해 억지로 빨간불을 만들 수 있다. 틀린 좌표는 틀린 실험을 부른다.

**필요한 수정**: `plan.md:116` 행의 포착기를 `AC-PSD-009`(거절 경로에서 lock 파일 미생성)로 고친다. 겸해서 "판정 반전"이 술어층이 아니라 게이트층 뮤턴트임을 표에 명시한다.

### D6. `경계 사례`의 안전 규칙 2건이 요구·수용 어느 층에도 없다

- 위치: `acceptance.md:23`, `acceptance.md:25`
- 심각도: **major** · 분류: **blocking**

두 규칙이 산문으로만 있다.

1. **대상 트리에서 `git`이 실패하면 통과로 처리하지 않는다** — "신호 부재는 깨끗함의 증거가 아니다"라고 acceptance.md가 직접 적었다.
2. **보존 디렉터리 쓰기가 실패해도 거절 판정은 유지한다** — 보존 실패를 통과로 바꾸지 않는다.

둘 다 이 카드의 존재 이유와 정확히 같은 성질의 불변식인데, 대응하는 `REQ-PSD-*`도 `AC-PSD-*`도 없다. 즉 `git` 실패를 조용히 0으로 접어 통과시키는 구현이나, 보존에 실패했다고 창을 내주는 구현이 모든 AC를 통과한다. 게이트가 가장 위험하게 무력화되는 두 경로가 정확히 미검증 상태다.

**필요한 수정**: `REQ-PSD-014`(Unwanted: 술어 실행이 실패하면 통과 판정을 내리지 않는다)와 `REQ-PSD-015`(Unwanted: 보존 실패는 거절 판정을 뒤집지 않는다)를 추가하고, 각각에 반증 AC를 붙인다 — (a) `git`이 없는/저장소가 아닌 디렉터리를 대상으로 주면 통과가 아니라 오류가 나오는지, (b) 보존 디렉터리를 쓰기 불가로 만든 픽스처에서 `integration-lock.json`이 여전히 생성되지 않는지. Tier M 요구 상한은 16이므로 15개까지 늘려도 예산 안이다(현재 13개, 실측).

---

## §4 md5 결정 — 독립 검증과 판단

리드가 이미 낸 발견이지만, 신뢰하지 말고 직접 재라고 했으므로 인용된 4개 사실을 하나씩 이 트리에서 확인했다. **네 개 모두 사실이다.**

```
$ cat -n .moai/astgrep-rules/security/crypto.yml
10  id: sec-weak-hash-md5
11  language: go
12  severity: error
18  rule:
19    pattern: md5.New()
   (머리말 3-6줄: "deliberately narrow: it flags the constructor call, not incidental references")

$ sed -n '157,165p' internal/hook/security/patterns.go
    Name:        "weak-crypto",
    Severity:    SevMedium,
    Patterns: []*regexp.Regexp{ mp(`(?i)\b(MD5|SHA1)\b`), ... }

$ cat internal/hook/security/testdata/scan-corpus/go_deny_md5.go
package sample
import "crypto/md5"
func Digest() any { return md5.New() }

$ grep -rn "crypto/md5" --include='*.go' internal/ pkg/ cmd/
internal/astgrep/testdata/fixtures/go/violation.go:11:  "crypto/md5"
internal/hook/security/testdata/scan-corpus/go_deny_md5.go:3:import "crypto/md5"
```

마지막 줄이 결정적이다. **이 저장소의 프로덕션 Go 코드에 `crypto/md5`는 0건이다** — 히트 2건은 둘 다 위반 픽스처/거부 코퍼스, 즉 "이것이 나쁜 예시"라고 선언된 파일이다. 그리고 인용된 다섯 개 내용 지문 지점은 전부 sha256임을 직접 확인했다(`internal/manifest/hasher.go:4,28,41` / `internal/graph/citation.go:4,72-73` / `internal/template/catalog_tree_hash.go:4,49,58` / `internal/verify/key.go:5,47` / `internal/runtime/audit_cache.go:6`).

### 리드가 던진 세 가지에 대한 내 판단

**(1) 의도상 오탐인가, 그리고 게이트가 의도를 모른다는 사실이 그 판단을 바꾸는가.**

의도만 보면 오탐에 가깝다 — 변경 검출용 내용 지문이지 보안 원시요소가 아니다. 그러나 그것으로 md5를 정당화하려면 게이트가 의도를 읽어야 하는데 읽지 못한다. `sec-weak-hash-md5`는 `md5.New()`를 무조건 `error`로 잡고, 정규식 층은 `\b(MD5|SHA1)\b`로 마크다운 안의 토큰까지 잡는다(추가 확인: `ScanBuffer`는 확장자 게이트를 거치지 않는다 — `internal/hook/security/scan.go:26` 이하가 내용만 받고, `internal/hook/post_tool_guardian.go:65`가 `ExtractToolInputContent`로 파일명 없이 내용만 넘긴다. 확장자 게이트는 ast-grep 경로에만 있다 — `internal/hook/security/scanner.go:92`, `ast_grep.go:416`).

다만 **내 판단의 무게는 게이트가 아니라 관례에 있다.** 게이트 쪽 실제 강도를 재 보면 생각보다 약하다:

- `internal/config/defaults.go:607-611` — `AstGrepGate{Enabled: true, BlockOnError: false, WarnOnlyMode: true}`. 주석이 명시한다: *"findings reported, commits never blocked; blocking is opt-in via gate.yaml."*
- 정규식 층은 `SevMedium`이고 `HandleSecurityScan`은 *"NEVER blocks the edit (REQ-SG-014)"*(`internal/hook/security/guardian.go:36-38`).
- CI(`.github/workflows/ci.yml`)의 ast-grep은 `sg` 바이너리를 rewrite guard용으로 설치하는 것이지 프로덕션 Go에 대한 저장소 전체 보안 스캔이 아니다. `sec-weak-hash-md5`를 참조하는 테스트 2개(`internal/astgrep/coverage_matrix_test.go:449`, `internal/hook/security/patterns_test.go`)를 열어 보니 커버리지 매트릭스 문서와 패턴 테이블을 고정하는 테스트이지 프로덕션 코드를 스캔하지 않는다.

**따라서 정직하게 말하면: md5를 넣어도 CI는 빨개지지 않는다.** 경고만 나온다. 나는 이 사실을 숨기고 심각도를 부풀리지 않는다.

그럼에도 이것을 blocking으로 올리는 이유는 게이트가 아니라 **예외 없는 관례**다. 이 저장소의 프로덕션 Go에서 내용 지문은 100% sha256이고 md5는 0%다 — 표본이 아니라 전수다. 여기에 md5를 넣으면 이 저장소 최초의 프로덕션 md5가 되고, 그것을 도입하는 카드가 하필 **"게이트가 조용히 무의미해지는 것"을 막으러 온 카드**다. 자기 경고 층이 상시 경고를 뱉는 코드를 심는 게이트는 그 경고 층의 신호 대 잡음을 스스로 깎는다.

**(2) `md5.Sum()`이 `md5.New()` 패턴을 피해 간다는 점 — 허용의 근거인가, SPEC이 가드를 회피하는 모양을 고르고 있다는 논거인가.**

후자다. 다만 정확히 해 둘 것이 있다. **SPEC은 지금 철자를 고르고 있지 않다** — spec.md도 plan.md도 `md5.New()`인지 `md5.Sum()`인지 적지 않았다. 그러니 "SPEC이 회피를 선택했다"는 현재 사실이 아니고, run 단계에서 생길 수 있는 위험이다.

그 위험을 미리 못 박아야 하는 이유는 이렇다. `md5.Sum()`을 고르면 `severity: error` 규칙은 피하지만 `\b(MD5|SHA1)\b` 정규식은 여전히 잡는다 — 그런데 그 층은 어차피 차단하지 않는다. 즉 **철자를 바꿔서 얻는 것은 아무것도 없고, 잃는 것은 관례다.** 그리고 규칙을 피해 가는 철자를 고른다는 행위 자체가, 규칙이 말하려던 것을 규칙의 구현 한계로 대체하는 것이라 방향이 거꾸로다. 규칙이 좁은 것은 그 머리말이 스스로 밝힌 설계(`crypto.yml:3-6`)이지 빠져나가라는 초대가 아니다.

**(3) 인용된 역사와 새 코드가 계산하는 해시를 내 권고가 구별하는가 — 구별한다.**

구별한다. 그리고 그 구별이 이 결정의 핵심이다.

`spec.md:35-42`의 세 줄은 카드 t487이 실제로 실행한 측정의 기록이다. 나는 그 셋을 이 트리에서 독립적으로 재현했다(t334는 읽기만 했고 수정·git 실행 0건):

```
$ md5 -q .../.claude/worktrees/t334/.claude/settings.json
4f455d9425a396d38c202f2614bc918f          ← spec.md:37과 일치
$ git cat-file -p 4837b05211fc7288747742f380de4048bc43de19:.claude/settings.json | md5 -q
437834679fcc2435761e264cec2ffc15          ← spec.md:39와 일치
$ md5 -q .moai/reports/t487/preserved-copies/settings.json.t334-dirty
4f455d9425a396d38c202f2614bc918f          ← spec.md:41과 일치 (23556 bytes)
```

세 줄 모두 참이다. **이것을 sha256으로 고쳐 쓰는 것은 실행된 적 없는 명령을 실행했다고 적는 것이고, 기록의 위조다.** 상류인 t487 판정서 §Q3도 마찬가지다 — 원문이 "사본 보존(path+md5) + 리드 blocker 보고"이고(`.moai/reports/t487/verdict.md:158`), SPEC의 md5는 거기서 물려받은 것이지 임의 선택이 아니다. 그 문장도 손대면 안 된다.

바뀌어야 하는 것은 **새 Go 코드가 계산하는 해시**뿐이고, 그것이 사는 자리는 넷이다: `spec.md:64`(REQ-PSD-005), `spec.md:67`(REQ-PSD-008), `plan.md:58-59`(파일명 접미 + 원장 필드), `plan.md:73`(`--json`의 `"md5"` 키).

계속성을 잃지 않는다는 점도 확인했다. t487의 보존 사본은 파일 자체가 남아 있으므로(23556 bytes, 위 실측), 미래에 3차 인스턴스가 나와 원장 행과 대조할 때 어느 해시로든 다시 계산하면 된다. 바이트 비교는 해시 함수와 무관하게 성립한다. 즉 sha256으로 옮기는 비용은 재계산 한 줄이고 잃는 능력은 없다.

### D7. Go 구현이 계산하는 해시를 sha256으로 옮긴다

- 위치: `spec.md:64`, `spec.md:67`, `plan.md:58-59`, `plan.md:73`
- 심각도: **major** · 분류: **blocking**
- **수정하지 말 것**: `spec.md:35-42`의 측정 기록, `.moai/reports/t487/verdict.md:158`의 인용. 둘 다 실행된 명령의 기록이다.

**필요한 수정**: 위 네 지점의 `md5`를 `sha256`으로 바꾸고, `--json` 키를 `"sha256"`으로, 파일명 접미를 `<sha256 앞 8자>`로 한다. 원장 행에 t487 기록과의 대조용으로 `md5` 필드를 **함께** 남기는 것도 가능하지만, 그 경우 md5가 계산되므로 위 판단이 그대로 적용된다 — 남기지 않는 쪽을 권한다. 지금 고치면 상수 몇 개지만, 원장 행이 한 줄이라도 쌓인 뒤에는 형식이 호환성 표면이 된다.

---

## §5 나머지 감사 차원

### 뮤턴트 반증 5종 — 개별 판정

| 뮤턴트 | SPEC이 지목한 포착기 | 내 판정 |
|---|---|---|
| 술어 삭제(항상 0) | F2 / AC-PSD-001 | **진짜로 RED가 된다.** F2는 정확히 1을 요구하고 뮤턴트는 0을 낸다 |
| 판정 반전 | plan.md: F1 / acceptance.md: AC-PSD-009 | **두 문서가 어긋나고 plan.md가 틀렸다** — D5 |
| 경로를 `settings.local.json`으로 오지정 | F2(0이 나옴) | **진짜로 RED가 된다.** F2 픽스처에 그 파일이 없으므로 pathspec이 0줄을 내고 AC-PSD-001의 "정확히 1"이 깨진다 |
| 경로 인자 누락 | F4(2가 나옴) | **진짜로 RED가 되고, 다섯 중 가장 강하다.** "제거" 뮤턴트가 아니라 "틀린 모양" 뮤턴트라서 t477 정격에 맞는다. 두 파일이 모두 dirty인 픽스처에서만 1과 2가 갈린다 |
| `--no-optional-locks` 제거 | argv 문자열 단정 | **단정 자체는 반증 가능하나 관측 지점이 정의되지 않아 공허할 수 있다** — D2. 그리고 그 선택을 정당화하는 문장이 과하다 — D3 |

F1..F4 픽스처 설계는 정확하다. 특히 F4가 있다는 점(둘 다 dirty → 1을 요구)이 경로 인자 누락을 잡는 유일한 형태이고, F3만 있었다면 그 뮤턴트는 빠져나갔다. `plan.md:108`의 `git init -b main` 못박기와 픽스처 로컬 `user.name`/`user.email` 지정도 옳다 — 인용된 선례 `SPEC-GIT-STATUS-FIXTURE-001`이 실재하고 `status: completed`임을 확인했다.

### `[NEEDS CLARIFICATION]` 2건 — 미룰 일이었나

**마커 1(기본 거절 vs opt-in 선례): 미룬 것은 옳다. 다만 근거 제시가 불완전해서 운영자가 잘못된 선택지 앞에 서게 된다.**

plan.md는 이 결정을 "네 개 가드가 전부 기본 false인데(`defaults.go:877-900`) 이 표면만 뒤집는다"로 틀 지었다. 네 개 값은 내가 확인했다 — `BranchGuard`/`IntegrationLock`/`AgentModelGuard`/`AgentStopGuard` 전부 `Enabled: false`. PreToolUse 훅이라 폭발 반경이 다르다는 논거도 타당하다.

그런데 **이 트리에는 SPEC이 언급하지 않은 세 번째 계열이 있다.** 기본 ON인 게이트들이다:

```
internal/config/defaults.go:607-611   AstGrepGate{Enabled: true, BlockOnError: false, WarnOnlyMode: true}
   주석: "ON by default in advisory mode (findings reported, commits never blocked); blocking is opt-in via gate.yaml"
internal/config/defaults.go:616-619   GraphFreshness{Enabled: true, Blocking: false, ...}
   주석: "the graph-freshness step mirrors the ast-grep posture: ON and advisory by default"
```

이 둘을 넣고 보면 이 저장소의 실제 축은 **켜짐/꺼짐이 아니라 차단/권고**다. 기본 ON인 것은 전부 권고이고, 차단하는 것은 전부 기본 OFF다. 내가 `defaults.go`에서 `Blocking: true` / `BlockOnError: true`를 찾아봤을 때 히트는 0이었다.

SPEC이 제안하는 것은 **기본 ON이면서 거절**이라 두 계열 어디에도 선례가 없다. 이것은 SPEC이 그린 그림("가드는 OFF지만 우리 표면은 훅이 아니니 예외")보다 더 강한 이탈이다. 결정은 여전히 운영자 몫이 맞다 — 다만 운영자에게 **정확한 지형**을 보여준 뒤 물어야 한다.

권고 선택지(운영자에게 이 형태로 물을 것): (가) 기본 ON + 거절(SPEC 현안, 두 계열 모두에서 이탈), (나) 기본 ON + 권고 + 보존·원장, 거절은 `gate.yaml` opt-in(ast_grep_gate 정격 그대로 — 내가 보기에 이 저장소 관례와 가장 정합적이다), (다) 기본 OFF + 거절, 로컬 config에서 켬(네 가드 정격).

**마커 2(킬 스위치 키 위치): 지금 결정할 수 있었고, 미룰 만한 사안이 아니다.**

`workflow.settings_drift_gate.enabled` 새 블록 대 `workflow.integration_lock` 하위 키의 선택인데, 트리의 증거가 이미 답을 준다. `internal/config/types.go:650-659`의 `IntegrationLockConfig` 주석이 그 키의 계약을 명시한다 — *"only the DENY layer is gated, exactly as BranchGuard gates only its deny"*, 그리고 그 deny 층은 **PreToolUse 훅의 `git merge` 거부**다. 새 게이트의 거절은 `moai integration acquire`라는 다른 표면이고 다른 시점이며, plan.md D1 스스로 두 축을 섞지 않겠다고 적었다(`plan.md:36` — `resolveIntegrationTarget()`과 섞지 않는다는 같은 취지).

같은 논리를 설정 키에도 적용하면 답은 새 블록이다. 기존 키 아래로 넣으면 하나의 `integration_lock.enabled`가 서로 다른 두 층을 게이트하게 되어, 어느 쪽을 끄려던 것인지 설정이 말하지 못한다 — REQ-PSD-010이 `--force`에 우회를 얹지 않는 이유와 **정확히 같은 모양의 중복 적재**다. SPEC이 플래그 축에서는 그 논거를 세워 두고 설정 키 축에서는 미룬 것이 일관성 없다.

**필요한 수정**: 마커 2를 해소하고 `workflow.settings_drift_gate.enabled`로 확정한다(근거: REQ-PSD-010과 동일한 중복 적재 회피 + `types.go:650-659`의 deny 층 계약). 마커 1은 남기되 위 (가)/(나)/(다) 세 선택지와 ast_grep_gate 계열 증거를 함께 제시해 다시 묻는다.

### 범위 규율 — 유출 없음 (PASS)

감시 대상이 `.claude/settings.json` 하나로 유지되는지 전수 확인했다.

```
$ grep -rn "described-source-diff\|settings.local.json" .moai/specs/SPEC-PREMERGE-SETTINGS-DRIFT-001/
plan.md:117        (뮤턴트 대상으로만 등장)
acceptance.md:11   (뮤턴트 대상으로만 등장)
spec.md:105        (Out of Scope 선언)
spec.md:115        (Out of Scope 제목: described-source-diff 술어 개선)
```

`settings.local.json`은 **오지정 뮤턴트의 대상**으로만 나오고 감시 대상으로는 한 번도 등장하지 않는다 — 오히려 그 파일을 감시하면 AC-PSD-001이 뒤집히도록 설계돼 있어, 조용한 확대가 **기계적으로 차단**된다. 좋은 형태다. t478의 described-source-diff 개선은 Out of Scope 제목에서만 언급되고 본문 어디에서도 그 코드를 건드리지 않는다. `internal/`에 `settings_drift`/`settings-drift` 식별자 충돌도 0건이다.

### 자동 복원 금지 — 검증 가능한 요구로 표현됨 (PASS)

`REQ-PSD-006`은 "shall not" 형태의 Unwanted로 제대로 쓰였고, `AC-PSD-008`이 **실제로 반증한다**: dirty 파일의 사전 md5와 사후 md5를 비교하므로, 복원하는 구현은 워킹 사본을 HEAD blob 내용으로 되돌려 md5가 바뀌고 FAIL이 된다. 산문 의도가 아니라 진짜 검증이다.

부수 단정 두 개 중 "대상 트리에 새로 생기거나 사라진 파일이 없다"도 기계적이다. 세 번째 "게이트가 실행한 명령 중 대상 트리를 수정하는 것은 하나도 없다"는 측정 방법이 없어 판단형이지만, md5 단정이 이 AC의 무게를 지탱하므로 차원 판정은 PASS로 둔다. D2에서 요구한 argv 기록 러너를 도입하면 이 단정도 자동으로 기계화된다 — 한 번의 수정이 두 곳을 고친다.

### 미검증 전제 스윕 — 인용 정확도는 예외적으로 높다

plan.md의 코드 인용을 **전부** 이 트리에서 대조했다. 결과: 인용 13건 중 **틀린 것 0건.**

| 인용 | 실측 |
|---|---|
| `integration.go:46` `integrationLockRoot()` | ✓ 46행이 함수 선언, 36-45행 주석이 주장대로 primary 루트 성질을 명시 |
| `integration.go:80` `integrationSessionID()` | ✓ |
| `integration.go:117` `resolveIntegrationTarget()` | ✓ |
| `integration.go:167` 세 동사 | ✓ |
| `integration.go:264-268` "recorded, never silent" | ✓ displaced 출력 |
| `integration.go:275-276` `--card` / `--force` | ✓ 275 `--card`, 276 `--force` |
| `integration_lock.go:53` `ErrIntegrationLockHeld` | ✓ |
| `integration_lock.go:109` `IntegrationLock` 구조체 | ✓ |
| `integration_lock.go:151` `integrationLockPath` | ✓ `<root>/.moai/state/integration-lock.json` |
| `.gitignore:224`, `:311` | ✓ `**/.moai/state/`, `.moai/state/` |
| `types.go:650` `IntegrationLockConfig` | ✓ |
| `types.go:670-679` `AgentStopGuardConfig` 주석 | ✓ "Observation is not gated" |
| `defaults.go:877-900` 네 가드 전부 false | ✓ |

미검증 전제로 남은 것은 §4의 md5 계열(D7으로 처리)과 위 마커 1의 기본값 선례 불완전 제시뿐이다.

### D8. M6 미러 대상 판정 — **지금 결정했다** (deferral 불필요)

- 위치: `plan.md:134` — *"어느 파일이 미러 대상이고 어느 것이 로컬 전용인지는 착수 시점에 확인한다"*
- 심각도: **minor** · 분류: **blocking**(미룰 필요가 없는 것을 미뤘고, 답이 나왔으므로 계획을 고칠 수 있다)

브리프가 판정하라고 한 항목이다. 실측:

```
$ ls internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md
-rw-r--r--  34268  ...    ← 미러 존재
$ ls internal/template/templates/CLAUDE.local.md
ls: No such file or directory      ← 미러 없음
```

판정: **`kanban-dispatch.md`는 미러가 있고 Template-First 규율이 걸린다**(`.claude/rules/moai/`는 `moai update`의 관리 대상 뿌리이므로 로컬만 고치면 다음 update에서 템플릿판으로 덮인다 — `CLAUDE.local.md` §2.3의 "경로 충돌" 유실 모양 그대로다). **`CLAUDE.local.md`는 미러가 없다** — 그 파일 자신이 Local-Only 목록에 자기를 올려 둔 의도적 로컬 전용 파일이므로 템플릿 작업이 없다.

**필요한 수정**: `plan.md:134`의 유보 문장을 위 판정으로 바꾼다. M6의 실제 작업은 3파일이다 — 로컬 `.claude/rules/.../kanban-dispatch.md`, 그 템플릿 미러, 그리고 `CLAUDE.local.md` §4.1(미러 없음). 그리고 M6에는 미러 동일성 검증을 붙인다(`diff` 한 줄). 다만 템플릿 미러 쪽 문구는 §2.1 템플릿 중립성 대상이므로 카드 id(`t488`)와 SPEC ID를 넣지 않는다 — SPEC은 이 제약을 언급하지 않았다.

---

## §6 optional 분류 발견 (수정 여부는 운영자·오케스트레이터 재량)

M6 규율에 따라 분리한다. 아래는 정확성·내부 정합성·이 문서가 명시한 기준 어느 것도 건드리지 않는다. **이것들만으로 FAIL을 만들지 않았고, 만들어서도 안 된다.**

- **D9 (minor/optional)** — `REQ-PSD-011`(`spec.md:70`)의 `Where` 절이 능력 게이트가 아니라 사용 사례다("사람이 손으로 확인하려 하면"). GEARS의 `Where`는 정적 설정·기능 플래그 축이다. 술어는 통과시켰다(수동태 "표면이 제공된다"는 Ubiquitous로 읽힌다).
- **D10 (minor/optional)** — `REQ-PSD-009`(`spec.md:68`)도 같은 문제의 다른 형태다. 호출별 CLI 플래그는 정적 능력 게이트보다 Event-driven(`When 호출자가 …을 지정하면`)에 가깝다.
- **D11 (minor/optional)** — `REQ-PSD-012`(`spec.md:71`)의 두 번째 문장 "리드는 이 필드로 판정을 읽는다"는 시스템 요구가 아니라 사람의 행동 서술이다. 요구층에서 덜어내고 plan.md 산문으로 옮기는 편이 깨끗하다.
- **D12 (minor/optional)** — `REQ-PSD-003`(`spec.md:62`)의 "어떤 수용 기준도 종료 코드에 의존하지 않는다"와 `REQ-PSD-013`(`:72`) 전체는 **검증층을 구속하는 요구**다. 그 결과 AC-PSD-003/004/005/009가 "AC가 존재한다"는 요구를 검증하는 순환이 생긴다. 실질 피해는 없다(그 AC들이 실제 요구 004/002/008도 함께 매핑한다). acceptance.md 머리말과 품질 게이트 절이 같은 규율을 이미 담고 있으므로 중복이기도 하다.
- **D13 (minor/optional)** — `AC-PSD-006`(`acceptance.md:14`)의 "판정 지점을 훑으면"은 측정 방법이 없어 판단형이다. 기계화 가능하다: 드리프트 관련 테스트·구현 파일에서 `ExitCode`/`ExitError`/`$?` 히트가 0임을 grep으로 단정하면 이분 판정이 된다.
- **D14 (minor/optional)** — `REQ-PSD-001`/`-004`가 정확한 명령줄을 요구층에 못박는다. 통상 이것은 요구에 구현을 넣은 것이지만, 여기서는 argv 고정 자체가 REQ-PSD-013의 반증 대상이라 방어 가능하다. 성질(인덱스 쓰기 락 금지)이 `REQ-PSD-002`에 따로 서술돼 있으므로 실질 결함은 아니다.
- **D15 (관측, 결함 아님)** — 상류 `SPEC-SETTINGS-ORIGIN-001`이 `status: in-progress`다. D7 차단 조건(retired/superseded/archived)이 아니므로 MP-5는 PASS다. 다만 이 카드가 그 SPEC의 권고를 구현하므로, 상류가 아직 열려 있다는 사실을 리드가 알고 있는 편이 좋다.

---

## §7 권고 (수정 순서)

되돌리기 어려운 순, 그리고 앞의 결정이 뒤를 바꾸는 순으로 놓는다.

1. **[운영자 관문 · High]** 마커 1을 §5의 (가)/(나)/(다) 세 선택지로 다시 묻는다. ast_grep_gate/GraphFreshness 계열(기본 ON + 권고)을 반드시 함께 제시한다 — 이 증거 없이 물으면 운영자는 실제로는 세 갈래인 것을 두 갈래로 본다. 이 답이 M5와 킬 스위치 설계를 정한다.
2. **[High]** 마커 2를 해소한다 — `workflow.settings_drift_gate.enabled` 새 블록. 근거는 `types.go:650-659`의 deny 층 계약과 REQ-PSD-010의 중복 적재 회피 논리다.
3. **[High]** D2 — `AC-PSD-005`에 관측 지점을 명시한다. argv 단일 원천 + 실행 경계 기록 러너. `plan.md:99`의 M1 서명도 함께 고친다(러너 주입 지점이 서명에 드러나야 한다).
4. **[High]** D1 — `AC-PSD-007`에 (d)를 추가해 `REQ-PSD-007`을 실제로 반증 가능하게 만든다.
5. **[High]** D6 — `REQ-PSD-014`/`-015`를 추가하고 각각에 반증 AC를 붙인다(`git` 실패 ≠ 통과, 보존 실패 ≠ 통과). 요구 15개는 Tier M 상한 16 안이다.
6. **[High]** D7 — `spec.md:64`, `spec.md:67`, `plan.md:58-59`, `plan.md:73` 네 지점의 해시를 sha256으로 옮긴다. **`spec.md:35-42`와 t487 §Q3 인용은 손대지 않는다.**
7. **[Medium]** D4 — D3 파일명 규칙의 시각 해상도를 올리거나 충돌 접미 규칙을 명시하고, `AC-PSD-007(c)`를 그 규칙에 맞춰 다시 쓴다.
8. **[Medium]** D5 — `plan.md:116`의 뮤턴트 2 포착기를 `AC-PSD-009`로 고치고, 술어층/게이트층 구분을 표에 넣는다.
9. **[Medium]** D3 — `acceptance.md:13` / `plan.md:121`의 "원리적으로 검출 불가"를 "신뢰성 있게 구별되지 않는다"로 정정한다.
10. **[Low]** D8 — `plan.md:134`의 미러 유보를 실측 판정으로 대체한다(kanban-dispatch.md 미러 있음 / CLAUDE.local.md 없음). 템플릿 미러 문구의 §2.1 중립성 제약을 함께 적는다.

optional(D9-D14)은 오케스트레이터 재량이다. 내가 보기에는 D13(AC-PSD-006 기계화)만 값이 있고 나머지는 문구 정리라 이번 회차에 넣지 않는 편을 권한다.

---

## §8 검사하지 않은 것 (Gaps)

부재를 통과로 읽지 않도록 명시한다.

- **구현 코드를 검사하지 않았다** — 이 카드는 구현 전이고 드리프트 술어·게이트·`preflight` 동사는 아직 존재하지 않는다. 뮤턴트 판정은 **설계 반증 가능성**에 대한 것이지 실행된 RED의 관측이 아니다. `acceptance.md:32`가 요구하는 5종 RED 실측은 run 단계의 몫이며 이 감사가 대신할 수 없다.
- **`t334` 워크트리는 `.claude/settings.json` 한 파일의 md5만 읽었다.** 수정·삭제·`git` 실행 0건. 그 트리의 다른 파일, 브랜치 상태, HEAD는 보지 않았다.
- **`spec.md:38`의 `git show 4837b052…` 명령을 그 트리에서 재실행하지 않았다.** 공유 오브젝트 저장소에서 `git cat-file -p`로 같은 blob을 읽어 해시를 대조했다. 같은 값이 나왔지만 같은 명령은 아니다.
- **테스트를 실행하지 않았다.** `go test` / `go vet` / `golangci-lint` 어느 것도 돌리지 않았다 — 감사 대상이 문서이고, 이 머신에서 여러 레인이 도는 중이라 부하를 만들지 않는 것이 §4 제약이다.
- **`.claude/rules/moai/workflow/kanban-dispatch.md`와 그 템플릿 미러의 내용을 대조하지 않았다.** 미러의 **존재**만 확인했다(34268 bytes). 두 파일이 현재 동일한지는 재지 않았다.
- **`moai integration acquire`를 실행하지 않았다.** 플래그 정의를 소스에서 읽었을 뿐 런타임 동작은 관측하지 않았다.
- **`sec-weak-hash-md5` 규칙을 `sg`로 실행하지 않았다.** 규칙 파일과 게이트 기본값과 CI 워크플로를 읽어 "차단하지 않는다"를 판단했다. `moai gate`를 실제로 돌려 md5 코드에 대한 출력을 본 것은 아니다.
- **`.moai/reports/t487/verdict.md`를 전부 읽지 않았다.** §Q3(156-162행)과 AC 표(165-175행)만 읽었다.
- **cross-model 2차 의견을 받지 않았다.** 이 감사는 Claude 단독 판정이다.

## §9 잔여 위험

- 위 수정 10건을 모두 반영해도, 이 게이트의 실효성은 **레인이 `acquire`를 자기 카드 워크트리에서 친다**는 절차적 사실에 의존한다. `plan.md:34`가 이 한계를 정직하게 적어 두었고(코드가 강제하지 않으며 잰 트리를 밝힌다) 그것이 옳은 처분이지만, 다른 트리에서 친 `acquire`는 깨끗한 트리를 재고 통과한다. 남는 위험이다.
- md5→sha256 이관을 **부분적으로** 하면(예: REQ만 고치고 `--json` 키를 놓치면) 원장 형식과 요구 문장이 갈린다. 네 지점을 한 번에 고쳐야 한다.
- D2의 러너 주입을 도입하면 M1 서명이 바뀌고 M3/M5 테스트가 그 서명에 의존한다. 3번 권고를 1·2번 답 이후 **가장 먼저** 처리해야 뒤의 마일스톤이 두 번 고쳐지지 않는다.

---

**반복 1/3 종료. 판정 FAIL.** 위 defect 목록이 기계 소비 가능한 수정 경로이며, 확인 재감사는 이 델타 범위로 한정된다(전면 재감사가 아니다). 다음 회차 진입 전에 §7의 1·2번(운영자 관문)이 먼저 닫혀야 한다 — MP-7은 점수와 무관한 관문이라 나머지 8건을 다 고쳐도 이것 없이는 PASS가 나오지 않는다.
