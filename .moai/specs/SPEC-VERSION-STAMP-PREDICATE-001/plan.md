# SPEC-VERSION-STAMP-PREDICATE-001 — 구현 계획

카드 t392 · Tier M · 트리 `9a3e2dabe`(워크트리 `.claude/worktrees/t392`, 브랜치
`WT-version-stamp-predicate`)에서 작성.

---

## §A 맥락

- 선행: `SPEC-VERSION-STAMP-GUARD-001`(t388, `status: completed`, develop `9a3e2dabe`).
  그 카드가 목록 **안**을 닫았고 목록 **밖**을 이 카드에 넘겼다.
- 술어와 등록부 형태는 운영자 판정이며 재론하지 않는다: 권위 토큰 + 정확-경로 등록부,
  golden 6개는 등록이 아니라 pin으로 제거.
- 산출 위치: `internal/cli/`(검사 · pin) · `internal/cli/testdata/`(픽스처) ·
  `.moai/docs/version-management.md`(문서 절).

### A.1 PRESERVE 목록 (건드리지 않는다)

- `internal/cli/version_sync_list_test.go` — t388의 검사. 헬퍼(`parseVersionStampEntries`,
  `isBoldLabel`)를 **읽어 쓰기만** 하고 시그니처를 바꾸지 않는다. 상수
  `expectedVersionStampEntries = 7` 도 그대로다.
- `.moai/docs/version-management.md` 의 「Version Stamps:」/「Release Artifacts:」 소제목
  문자열 — t388 검사의 앵커다. 항목은 늘지 않고 소제목은 바이트 그대로 유지한다.
- `internal/template/templates/**` — 이 SPEC의 산출물이 없다. Template-First 미러 의무 없음.
- `pkg/version/version.go` — 읽기만 한다. 값을 고치지 않는다.
- `origin/release/v3.1.4` — 이 카드가 손대지 않는다.

---

## §B 착수 전 재측정 (M0)

[HARD] 아래 세 수는 `spec.md` 가 트리 `9a3e2dabe` 에서 잰 값이다. run-phase는 **작업 트리에서
다시 재고**, 다르면 상수와 문서 표를 그 값으로 고친 뒤 진행한다(§8 R-7).

```bash
git rev-parse HEAD
git grep -lF "$(sed -n 's/.*Version = "\(v[0-9.]*\)".*/\1/p' pkg/version/version.go)" -- . \
  ':!.moai/reports' ':!.moai/specs' ':!.moai/release-notes' ':!CHANGELOG.md' \
  ':!*_test.go' ':!docs-site/content/*/changelog*' | wc -l
```

| 값 | `9a3e2dabe` 측정 | 의미 |
|---|---|---|
| 거부 목록 적용 파일 | 34 | pin 전 스윕 크기 |
| pin 후 스윕 | 28 | green 시점의 스윕 크기. **상수가 아니다**(범프마다 움직인다) |
| 등록부 항목 수 | 28 | **상수 1** — 범프에 불변 |
| `stamp` 분류 수 | 7 | **상수 2** — 범프에 불변 |
| 제외가 감추는 파일 | 121 | `121 + 34 = 155`(전체 **추적** 파일) 귀속 확인 |

[HARD] 위 명령이 `git grep` 인 것은 **의도적이다.** 스윕의 모집단이 추적 파일 집합이므로
(`spec.md` §2.0, REQ-VSP-012) 측정 도구와 판정 대상이 같은 집합을 본다. `grep -rl` 로 재면
162가 나오고 그 차이 7은 전부 이 카드 자신의 미추적 산출물이다 — 그 값을 상수로 쓰지 않는다.

[HARD] **커밋 후 값은 규칙으로만 적는다 — 정수를 예측하지 않는다**(iter-3, `spec.md` §2.3).
이 카드가 커밋하는 산출물 파일 하나마다 그것이 속한 제외 그룹의 수와 합·전체 트리 값이 1씩
는다. **`34` 는 불변이다**(산출물이 전부 제외 그룹 안). 0.2.0판은 여기에 「121 → 128,
155 → 162」를 박았고, 그것을 검사한 감사 보고서 자신이 여덟 번째 산출물이 되면서 그 자리에서
거짓이 됐다 — 같은 형태로 세 번 낡았다. run-phase는 병합 트리에서 재고, 늘어난 값을 드리프트로
읽지 않는다.

---

## §C 마일스톤 — 되돌리기 비싼 결정부터

순서는 **변경 가능성이 높은 결정을 먼저** 놓는다. 리뷰가 가장 값싸게 개입할 수 있는 자리가
앞에 오고, 기계적인 작업이 뒤로 간다. 실행 순서는 M1 → M5 그대로이며, M3 → M4 의 선후는
[HARD] 제약이다(§D).

### M1 — 등록부의 자료 모형과 분류 어휘 (Priority High · 가장 되돌리기 비쌈)

이 SPEC에서 **가장 바뀔 법한 결정**이다. 코드보다 먼저 리뷰에 걸린다.

- 분류 어휘를 `stamp` / `prose` 둘로 고정한다. 셋째 값(`fixture` 등)을 도입하지 않는다 —
  golden 은 pin으로 사라지므로 분류될 대상이 없다.
- 분류 판별식을 문장으로 못박는다: **「범프가 이 파일을 다시 써야 하는가, 혹은 이 파일을
  읽는 검사가 범프 누락으로 깨지는가」** → 예면 `stamp`, 아니면 `prose`(REQ-VSP-002).
  이 판별식은 사람이 읽어 적용한다. 기계로 결정하지 않으며, 오분류가 열린 구멍으로 남는
  것을 `spec.md` §3.1(나)가 이름으로 받는다.
- 28항목을 정확 경로로 열거한다(스탬프 7 + 서술 21). 글로브 금지.
- **상수 둘을 등록부의 성질로 고정한다**: 항목 수 28 · `stamp` 수 7. 스윕 개수에 대한 상수는
  **두지 않는다**(REQ-VSP-005 — 사유는 `spec.md` §6.1).
- 등록부의 거처를 Go 리터럴로 둔다(`spec.md` §6의 사유 셋). 문서에 두지 않는다.
- 판정부를 **순수 함수**로 설계한다. 이 결정이 M2의 합성 RED을 가능하게 하므로 여기서 못박는다.

산출: `internal/cli/version_stamp_registry_test.go` 의 자료 선언부 + 함수 시그니처.

### M2 — 순수 판정 코어 + 합성 RED 넷 (Priority High)

AC-VSP-004 · 005 · 006 · 009 · 013 · 015의 RED을 각각 **자기 입력**으로 관측한다. 서로의
빨강을 증거로 쓰지 않는다(t388 §5.1).

| 단언 | 합성 입력이 만드는 상태 |
|---|---|
| 신선도(004) | `stamp` 항목 하나의 내용이 권위 토큰을 담지 않음 |
| 비공허성(005) | 스윕 결과가 빈 목록 — 스탬프 7개가 전부 빠졌다고 이름으로 불러야 한다 |
| 내용-대-경로(006) | 이름에만 토큰이 있고 내용에는 없는 경로 |
| 등록부 ⇄ 문서(009) | 한쪽에만 있는 경로 하나 |
| 유령 항목(013) | 등록부에 실재하지 않는 경로 하나 |
| 도달 범위(015-가) | 등록부 경로 하나가 빠진 모집단 — 그 경로를 이름으로 불러야 한다 |
| 도달 범위(015-나) | 코어가 받은 경로 일부를 건너뛴 상태 — `judged=N examined_of=M` 로 두 수를 함께 내야 한다 |

RED 일곱을 각각 관측하고 verbatim 출력을 `progress.md §E.2` 에 기록한 뒤 GREEN으로 넘어간다.

[HARD] **양방향으로 돌린다.** 각 합성 입력마다 (a) 잡아야 하는 경우와 (b) **여전히 통과해야
하는 경우**를 함께 먹인다. 특히 004는 같은 상태의 `prose` 항목이 finding을 내지 **않는** 것을
같은 실행에서 확인한다 — 면제가 살아 있어야 §3의 4행이 정상으로 남는다.

### M3 — 모집단 드라이버 + 거부 목록 열거, 그리고 완결성 RED (Priority High)

[HARD] **이 마일스톤은 M4보다 먼저 실행되어야 하며, 그 이유가 §D다.**

- **모집단 드라이버는 `git ls-files` 다** — `filepath.WalkDir` 이 아니다(REQ-VSP-012, 사유는
  `spec.md` §2.0). `git` 호출은 이 드라이버 한 자리에만 있고, 판정 코어에는 없다.
  `.git/` 제외 조항은 **두지 않는다** — 모집단에 애초에 들어오지 않는다.
- [HARD] **argv는 이름 붙인 리터럴 하나다** — `var gitLsFilesArgv = []string{"git", "ls-files"}`
  를 파일에 **정확히 한 번** 두고 드라이버가 그것만 쓴다. AC-VSP-010(b)가 이 형태를 허용
  목록으로 판정하므로(iter-3, N5), 흩어 놓으면 판정이 거부 목록으로 되돌아간다(AP-12).
- [HARD] **도달 범위 단언 둘을 여기서 배선한다**(REQ-VSP-015): 드라이버가 넘긴 경로 집합에
  등록부 28경로가 전부 들어 있는지, 그리고 코어가 훑은 수가 넘겨받은 수와 같은지. 어느 쪽도
  기대 **크기**를 들지 않는다(AP-11).
- 제외 그룹 **여섯**을 리터럴로 열거한다(`spec.md` §2.3 표).
- 권위 토큰을 `pkg/version/version.go` 에서 뽑는다(REQ-VSP-001).
- **AC-VSP-003의 RED을 pin 이전 트리에서 관측한다.** golden 6개가 이름으로 지목되어야 한다.
- 이 시점 스윕은 34, 등록부는 28이지만 **개수 단언은 울지 않는다** — 상수는 등록부의
  성질(28 · 7)이고 스윕 개수에 대한 상수가 없기 때문이다. 우는 것은 완결성 단언 하나이며,
  그것이 옳다: 결함은 「6개가 미등록」이지 「개수가 어긋남」이 아니다.
  초판은 여기서 개수 단언이 함께 울 것으로 적었고 AP-6으로 그 빨강을 격리했다. 상수를
  없앤 지금 격리할 대상이 사라졌다 — AP-6은 M2의 005 RED에만 남는다.

### M4 — pin + 픽스처 재생성 (Priority Medium · 기계적)

- `status_golden_test.go` · `doctor_golden_test.go` 에 `version.Version` pin 을 넣는다.
  선례를 그대로 따른다(`version_test.go:180-186`: 원값 보존 → 대입 → `defer` 복원).
- `UPDATE_GOLDEN=1 go test ./internal/cli/ -run "TestStatus_Current|TestStatus_NoColor|TestDoctorGolden" -count=1`
  로 6개를 재생성한다.
- 재생성 diff를 **읽고** 커밋한다 — 버전 줄 하나만 바뀌어야 한다(각 golden 의 v3.1.3 출현은
  실측 1회).
- 이후 스윕은 28이 되고 M3의 빨강이 GREEN으로 넘어간다.

### M5 — 문서 절 (Priority Medium · 기계적)

`.moai/docs/version-management.md` 에:

- 거부 목록 열거 표(그룹 **여섯** · 사유 · 감추는 파일 수, 트리 SHA에 못박음) — REQ-VSP-007.
- 3축 표와 **닫지 못하는 것 둘**(미등록 + 낡음 · `prose` 오분류) — REQ-VSP-011.
- **등록부 유지 계약** 한 문단 — 누가 · 언제 · 그 사이 검사가 무엇을 말하는가, 그리고
  **「범프는 등록부를 건드리지 않는다」** — REQ-VSP-014. 등록부 **파일 이름 하나**만 적고
  28경로를 옮겨 적지 않는다(§6 사유 1 보존).
- t388 서술의 **정확한 교체 문안**(§E).

[HARD] **이 절이 쓰는 문서는 영문이다**(실측 `grep -cP '[가-힣]' … ` → 0). 위 네 항목 전부
영문으로 쓴다 — §E의 교체 문안은 그대로 붙여 넣고, 나머지 셋도 같은 문체로 영문 저작한다.
run-phase는 §E 문안을 **재집필하지 않는다**(§D.2). 판정 grep은 `acceptance.md` AC-VSP-011 ·
014에 영문 구절로 못박혀 있으며 현재 문서에서 전부 0건이다.

---

## §D [HARD] 순서 제약 — RED은 pin 이전 트리에서만 관측된다

pin은 golden 6개를 **술어에서 제거**한다(34 → 28). 제거된 뒤의 트리에 검사를 걸면 완결성
단언은 그 6개를 **영원히 보지 못한다**. 따라서 AC-VSP-003의 RED은 golden 이 아직 권위 토큰을
담고 있는 트리에서 관측되어야 한다.

**그 트리를 이름으로 못박는다: 이 워크트리의 `9a3e2dabe`(및 M1~M3 커밋이 얹힌 그 후속),
즉 M4 이전의 작업 트리다.**

운영자 지시가 제시한 `origin/release/v3.1.4` 는 **쓰지 않는다.** 사유는 `spec.md` §1.1이며
요지는 둘이다.

1. 그 트리의 권위 토큰은 `v3.1.4` 이고 낡은 golden 은 `v3.1.3` 을 담는다. 현재-토큰 스윕에
   걸리지 않으므로 **RED이 나오지 않는다** — 드리프트는 보이지만 검사는 침묵한다.
2. 그 tip(`26898312e`)은 HEAD의 조상이 아니며(`git merge-base --is-ancestor` rc=1) CI의
   `fetch-depth: 0` 으로도 들어오지 않는다. t388 §7 R-1의 도달 불가 문제는 **재발한다** —
   다만 (1) 때문에 애초에 쓰지 않으므로 걸리지 않는다.

### D.0 M4 → M5 는 순서 제약이 **아니다** (iter-2)

감사가 D5로 지목한 파손을 여기에 기록한다. 실측:

    grep -n 'v3\.1\.3' .moai/docs/version-management.md   → 90행 1건

그 90행이 §E가 갈아 끼우는 바로 그 문장이고, 교체 문안에는 버전 토큰이 없다. 그러므로 M5
이후 그 문서는 스윕에서 빠진다. **초판 설계에서는 이것이 파손이었다** — 「스윕 = 28」 상수가
27을 만나 SPEC 자신의 마지막 마일스톤이 자기 GREEN을 깼을 것이다.

수리는 M4/M5 순서를 정하는 것이 아니라 **그 상수를 없애는 것**이었다(`spec.md` §6.1 —
D2 수리). 상수가 등록부의 성질(28 · 7)이 된 지금:

- 문서가 토큰을 담으면 → 등록돼 있으므로 REQ-VSP-003 통과.
- 문서가 토큰을 잃으면 → 등록된 `prose` 이고 REQ-VSP-004는 `prose` 를 면제하므로 통과.
  REQ-VSP-002가 등록부를 스윕의 **상위집합**으로 정의하므로 이것이 위반이 아니다.
- 파일 자체는 남으므로 REQ-VSP-013(유령 가드)도 통과.

**따라서 `.moai/docs/version-management.md` 는 등록부에 `prose` 로 남고 등록부는 28을
유지하며, M4와 M5 사이에 순서 제약이 없다.** 파손이 설계 결함이었지 순서 문제가 아니었다는
것이 요지다 — 순서로 고쳤다면 같은 결함이 다음 문서 편집에서 되살아났을 것이다.

### D.1 CI 적색 구간의 처리

M3 커밋만 push되면 그 head의 CI는 빨갛다(설계상 그래야 한다). 처리:

- M3와 M4를 **같은 push로 올린다**. 조용한 head는 M4 이후에만 만든다.
- M3의 RED은 CI가 아니라 **작업 트리 실행**으로 관측하고 verbatim 을 `progress.md §E.2` 에
  남긴다. CI를 RED 관측의 주체로 삼지 않는다.

### D.2 단언 메시지 계약 (측정 **전에** 못박는다)

RED을 관측하기 전에 기대 실패 문자열을 고정한다. 예측이 틀려도 고정되어 있으면 조사가
강제된다.

| 단언 | 기대 실패 문자열(형태) |
|---|---|
| 완결성(003) | `unregistered file carries the authoritative token: <path>` |
| 신선도(004) | `registered stamp does not carry the authoritative token: <path>` |
| 비공허성(005) | `registry entries=<n> expected=28` / `stamp entries=<n> expected=7` / `registered stamp missing from sweep: <path>` |
| 내용-대-경로(006) | `<path>` 가 스윕 결과에 **없어야** 한다 |
| 등록부 ⇄ 문서(009) | `stamp set differs from documentation list: <path>` |
| 유령 항목(013) | `registry entry does not resolve to a file: <path>` |

[HARD] 005의 세 문자열 어디에도 **스윕 개수의 기대값이 없다.** 스윕은 「스탬프 7개를 전부
담고 있는가」로만 판정되며, 그 판정이 실패할 때 개수가 아니라 **빠진 경로**를 낸다.

### D.3 앵커·상수 리터럴 계약

| 리터럴 | 값 | 소유 |
|---|---|---|
| 문서 스탬프 앵커 | `**Version Stamps:**` | t388 (재사용, 변경 금지) |
| 권위 토큰 추출 대상 | `pkg/version/version.go` 의 `Version` 대입 줄 | 이 SPEC |
| 모집단 획득 | `git ls-files`(드라이버 한 자리) | 이 SPEC |
| 등록부 기대 항목 수 | 28 (M0 재측정으로 확정) | 이 SPEC |
| `stamp` 기대 분류 수 | 7 (M0 재측정으로 확정) | 이 SPEC |
| ~~스윕 기대 개수~~ | **없음** — iter-2에서 제거 | — |

두 개수 상수는 **각자 따로** 보유한다. 하나를 다른 하나에서 유도하면 서로를 비교하는
자기 참조가 되어 아무것도 단언하지 않는다.

[HARD] 스윕 개수 상수를 다시 넣지 않는다. 그것은 범프마다 28 → 7 로 움직이며, 움직이는
이유가 이 SPEC이 §3 4행에서 **정상**이라고 판정한 상태다. 상수를 되살리는 것은 정상을
결함으로 판정하는 일이다(`spec.md` §6.1).

---

## §E t388 서술의 교체 문안 (REQ-VSP-011)

`.moai/docs/version-management.md` 의 현재 문장:

> The guarantee it establishes is **partial**. It catches the list naming a path that does not
> exist. It **does not detect** a stamp site that is absent from the list — which is the
> direction that actually bit us: the v3.1.3 bump missed `docs-site/hugo.toml`, and nothing
> said so. Closing that direction is card t392.

**판정: 좁힐 자격이 있다.** 두 번째 문장(「목록에 없는 스탬프 사이트를 탐지하지 못한다」)은
REQ-VSP-003이 착지하면 **현재 토큰을 담은 경우에 한해** 거짓이 된다. 그러나 좁히는 폭을
정확히 지켜야 한다 — 이 카드가 존재하는 이유가 문서가 아는 것보다 더 많이 주장했기
때문이므로, 수리 과정에서 같은 실수를 되풀이하지 않는다.

[HARD] **교체 문안은 영문이다.** 대상 문서는 영문 전용이다 — 트리 `9a3e2dabe` 에서
`grep -cP '[가-힣]' .moai/docs/version-management.md` → **0**(104줄 전부). 0.2.0판은 이
자리에 한국어 문안을 못박아, run-phase에 「영문 문서에 한국어를 붙이거나 / 번역하라(§D.2가
금지한 재집필)」는 서로 모순되는 지시를 남겼다. 아래는 그 수리이며, 주변 문단의 문체
(설명적 산문 + `**bold**` 강조 + 백틱 경로)를 따른다. **다음 편집자도 한국어를 되넣지 않는다.**

교체 문안(M5가 이 문장으로 갈아 끼운다 — 영문, 그대로 붙여 넣는다):

> Two checks now stand side by side. One catches the list naming a **path that does not exist**
> (t388). The other catches a file carrying the authoritative version token that the **registry
> does not name**, and a registered stamp that **does not carry the authoritative token** (t392).
> The registry lives in `internal/cli/version_stamp_registry_test.go`.
>
> Things still go uncaught. **At least the following remain, and this list is not exhaustive.**
>
> 1. A file that is neither in the registry nor carrying the authoritative token — an
>    unregistered site left holding only an aged-out token matches no predicate.
> 2. A genuine stamp site registered as `prose` — completeness passes because it is registered,
>    freshness is skipped because it is `prose`, and the documentation cross-check passes because
>    it is in neither stamp set. All three assertions are blind to it.
> 3. A stamp inlined inside a file the exclusion set hides — had the same fixture lived in a
>    `*_test.go` file rather than in `testdata/*.golden`, this predicate would not see it.
> 4. A stamp that is not a version token at all, such as `releaseDate` in `docs-site/hugo.toml`.
> 5. A site that renders the version rather than carrying it, such as
>    `internal/template/templates/.moai/config/sections/system.yaml.tmpl`.
> 6. A file the repository does not track — outside the sweep's population.
>
> None of this means the list can no longer rot.
>
> Who maintains the registry, and when: the author of the commit that adds or removes a file
> edits the registry in that same commit. **A version bump does not touch the registry** — it
> rewrites the seven stamp files and adds or removes nothing. Between a file landing and the
> registry edit, the check fails naming the path: a new token-carrying file is reported as
> unregistered, and a deleted registered path is reported as unresolved.

**판정 grep이 이 문안에만 걸리도록 하는 규율.** `_test.go` · `system.yaml.tmpl` 같은 토큰은
문서의 **다른 절에 이미 존재하므로**(현 L83·L90), 그것들로 (b)를 판정하면 M5가 아무것도 안
써도 통과한다. 그래서 AC-VSP-011(b)는 위 문안에만 존재하는 **여섯 구절**로 못박고, 그 여섯이
현재 문서에 **0건**임을 실측해 두었다(= 미리 못박은 RED). 목록은 `acceptance.md` AC-VSP-011.

[HARD] **판정 구절은 줄바꿈으로 쪼개지지 않아야 한다.** `grep` 은 줄 단위이므로 구절 가운데에
개행이 들어가면 문안이 정확히 들어갔는데도 판정이 0건이 된다 — 판정을 무르게 하는 것이 아니라
**거짓 빨강**을 만드는 방향의 취약성이다. M5는 위 문안을 붙여 넣을 때 각 판정 구절(총 11개,
`acceptance.md` AC-VSP-011 · 014)을 **한 줄 안에** 유지한다. plan-phase에서 이 문안이 그
조건을 만족함을 실측했다: 11개 구절이 §E 안에서 각각 1줄에 있다.

**왜 열린 형태여야 하는가.** 초판 문안은 「셋이다」로 닫았고, 그 닫힌 수가 위 2·3·6을
빠뜨렸다. 이 카드가 존재하는 이유가 **문서가 아는 것보다 더 많이 주장했기 때문**이므로,
과잉 주장을 더 작은 과잉 주장으로 바꾸는 것은 수리가 아니다. 남은 경우의 수를 안다고
주장하지 않는 것이 유일하게 참인 형태다.

**그대로 참인 것**: 「목록이 더는 썩지 않는다」에 해당하는 서술은 문서에도 SPEC에도 쓰지
않는다는 t388의 규율. 이 카드가 그 규율을 유지한다.

---

## §F 안티패턴

- **AP-1 — 포괄 면제.** 거부 목록을 「릴리스 부산물 일체」 같은 한 조항으로 적는 것. t388이
  이름 붙인 첫 안티패턴이며 REQ-VSP-007이 금지한다.
- **AP-2 — 경로에서 토큰 뽑기.** `git grep -n` 출력이나 모집단 목록의 **경로**에서 버전
  문자열을 세는 것. 거부 목록 안에 이름-토큰 파일이 여덟 있다(§spec.md 2.4).
- **AP-3 — 보고를 판정으로 착각.** finding 을 출력하고 exit 0 으로 끝나는 검사.
  판정은 `t.Errorf` 로 낸다. AC의 증거는 인쇄된 목록이 아니라 **비영 종료**다.
- **AP-4 — pin 이후 트리에서 완결성 RED을 찾기.** §D가 금지한다.
- **AP-5 — golden 재생성 diff를 읽지 않고 커밋.** 버전 줄 외의 변화가 섞이면 pin이 아니라
  다른 회귀다.
- **AP-6 — 한 단언의 빨강을 다른 단언의 증거로.** M3에서 개수 단언과 완결성 단언이 같은
  트리에서 함께 우는 것이 정상이지만, 005의 RED은 M2의 빈-스윕 입력이 낸 것을 쓴다.
- **AP-7 — 시간 추정.** 우선순위 라벨과 단계 순서만 쓴다.
- **AP-8 — 파일시스템 walk로 모집단 잡기.** primary 체크아웃에서 gitignore된 형제 워크트리
  183개로 내려가고, 그중 하나만으로 토큰 보유 파일이 144다(실측). prune 목록은 곧 썩는 두
  번째 등록부가 된다(`spec.md` §2.0). REQ-VSP-012가 금지한다.
- **AP-9 — 스윕 개수 상수 되살리기.** 「스윕 개수 = 28」은 범프 직후 7이 되고, 그 7은 이
  SPEC이 정상이라고 판정한 상태다. 개수가 아니라 **경로**를 내는 단언으로만 판정한다
  (§D.3, `spec.md` §6.1).
- **AP-11 — 모집단 크기 리터럴 넣기. [iter-3 신설]** 「추적 파일 = 10,048」류의 상수는 범프에
  불변이지만 개발에 불변이 아니다. 무관한 커밋에서 맨 숫자로 실패하고 값을 깎게 만든다 —
  AP-9가 금지한 것과 같은 형태다. 도달 범위는 REQ-VSP-015의 **양변 실행 시점 단언 둘**로만
  묶는다(`spec.md` §6.2).
- **AP-12 — `git` argv를 여러 자리에 흩기. [iter-3 신설]** 모집단 획득의 argv는 **이름 붙인
  리터럴 하나**여야 한다 — `var gitLsFilesArgv = []string{"git", "ls-files"}`. 흩어 놓으면
  AC-VSP-010(b)의 허용 목록 판정이 불가능해지고 거부 목록으로 되돌아간다.
- **AP-13 — 양방향 확인에 실제 색인 건드리기. [iter-3 신설]** `git add` / `git rm` 으로 RED을
  만드는 것. §6의 순수 코어 설계가 **정확히 그것을 피하려고** 있다. 두 방향은 합성 모집단
  둘로 낸다(AC-VSP-012). 공유 체크아웃에서 색인을 만지는 것은 이 저장소의 독트린 위반이기도
  하다.
- **AP-10 — 판정 코어에 `git` 넣기.** 모집단 획득은 드라이버 한 자리에만 있다. 코어가
  외부 프로세스를 부르면 합성 입력 RED이 불가능해지고 REQ-VSP-010이 깨진다.

---

## §G 검증 (run-phase 자가 확인)

```bash
# 1. 새 검사 + t388 검사가 함께 돈다
go test ./internal/cli/ -run 'TestVersionSyncList|TestVersionStampRegistry' -count=1

# 2. 패키지 전체 (golden 재생성 이후)
go test ./internal/cli/... -count=1

# 3. 픽스처에 버전 토큰이 남지 않았는지 (M4 이후 0이어야 한다)
grep -c 'v3\.1\.3' internal/cli/testdata/*.golden || echo "no matches"

# 4. 크로스 플랫폼 컴파일
go vet ./internal/cli/...
GOOS=windows GOARCH=amd64 go vet ./internal/cli/...
```

[HARD] `go test ./...` 를 로컬에서 돌리지 않는다. 전 패키지 판정은 CI 몫이다.

---

## §H 교차 참조

- `spec.md` §1.1(운영자 전제 반증) · §2.0(모집단) · §2.3(거부 목록 열거) · §3(3축 표 +
  §3.1 닫지 못하는 것 둘) · §5(Tier 도출) · §6.1(등록부 유지 계약) · §8(잔여 위험)
- `acceptance.md` §D — AC 14건과 각자의 RED 셀
- `.claude/rules/moai/development/verification-completeness.md` §1.1 · §1.3 · §2 · §2.1
- `.moai/specs/SPEC-VERSION-STAMP-GUARD-001/` — 선행 카드 전체
- `.moai/reports/t392/baseline.md`
