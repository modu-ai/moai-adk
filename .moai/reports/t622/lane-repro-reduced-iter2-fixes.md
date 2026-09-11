# t622 — 축소판 감사 2회차 수정(SPEC 0.2.3) 레인 재현 기록

- 카드: t622 · SPEC: SPEC-GIT-DELIVERY-PROCEDURE-001 0.2.3 — 재현은 커밋 전 워킹 사본(HEAD `0af445525` 위)에서 했고, 그 내용이 그대로 `8cf5aaf06` 으로 커밋됐다(재현 직후 `git add` 로 같은 파일을 스테이징)
- 지시: 리드 — 고친 검출식마다 감사자 반례가 FAIL, 올바른 문장이 PASS 임을 레인이 직접 재현
- 방법: 판정 명령은 `acceptance.md` 에서 글자 그대로 복사해 실행했다. 반례·양성 파일은 세션 스크래치에 `perl -CSD` 로 만들었다(파일 자체는 반출하지 않음 — 아래에 줄 내용을 그대로 적는다).

## D1 — AC-GDP-026 별칭 방향 판정 (ii), 존재 개수 가드 (iii)

판정 (ii) (acceptance.md 655행):

```
awk '/(^|[^A-Za-z-])--merge([^A-Za-z-]|$)/ && (!/--merge([^A-Za-z-].*)?[Dd]eprecated alias (of|for) .?--auto-merge([^A-Za-z-]|$)/ || /(not|no longer|never) (a |an |the )?[Dd]eprecated/ || /[Uu]n-?deprecat/) {print FNR ": " $0}'
```

| 파일 | 줄 | 결과 |
|---|---|---|
| 반례 | `` - `--merge`: auto-merge the PR after sync (the older `--auto-merge` spelling is deprecated) `` (감사자 줄) | 적중 → FAIL |
| 반례 | `` - `--auto-merge`: deprecated alias of `--merge` `` (방향 뒤집음) | 적중 → FAIL |
| 양성 | `` - `--merge`: deprecated alias of `--auto-merge` (logs a warning) `` | 미적중 → PASS |

판정 (iii) (acceptance.md 658행) `/usr/bin/grep -c -E '(^|[^A-Za-z-])--merge([^A-Za-z-]|$)'`:

| 파일 | 줄 | 결과 |
|---|---|---|
| 제거 반례 | `Modes: auto, force, status, project. Flags: --auto-merge, --skip-mx` (감사자 줄) | `0` → FAIL |
| 양성 | 위 양성 줄 | `1` → PASS |

## D2 — AC-GDP-028 전원 승인 (a), 부분 승인 금지 (b)

판정 (a) (acceptance.md 777행): `awk '/[Tt]eam mode/ && /--auto-merge/ && /(^|[^A-Za-z])[Aa]ll( [A-Za-z-]+)?( [A-Za-z-]+)? approv/ {c++} END{print c+0}'`
판정 (b) (acceptance.md 780행): 목록 전체를 글자 그대로 사용.

| 파일 | 줄 | (a) | (b) | 결과 |
|---|---|---|---|---|
| 반례 | `` In team mode, `--auto-merge` merges once CI passes; approvals are optional. `` (감사자 줄) | 파일 합계 `0` | 적중 | FAIL |
| 반례 | `` In team mode, `--auto-merge` merges after at least one approval. `` (감사자 줄) | (위와 같은 파일) | 적중 | FAIL |
| 양성 | `` In team mode, `--auto-merge` merges only after all approvals are obtained. `` | `1` | 미적중 | PASS |

## D3 — AC-GDP-029 (b)

판정 (acceptance.md 834행) 글자 그대로 사용.

| 줄 | 결과 |
|---|---|
| `` In personal and manual modes, `--auto-merge` merges without requiring approval. `` (감사자가 든 올바른 문장 형태) | 미적중 → PASS |
| `` In personal and manual modes, `--auto-merge` merges after a code review. `` (감사자 반례 형태) | 적중 → FAIL |

## D4 — AC-GDP-027 (i)·(ii)

판정 (i)·(ii) (acceptance.md 727·730행) 글자 그대로 사용.

| 줄 | (i) | (ii) | 결과 |
|---|---|---|---|
| `` - `--no-merge`: Deprecated no-op; disables auto-merge. `` (감사자 반례 형태) | 미적중 | 적중 | FAIL |
| `` - `--no-merge`: Deprecated no-op kept for compatibility; changes nothing. `` | 미적중 | 미적중 | PASS |

## 함께 확인한 것

- 바뀐 파일: `git diff --stat HEAD` → SPEC 네 파일뿐(acceptance 189, plan 24, progress 19, spec 23 줄 변경).
- 번호·개수: `**REQ-GDP-NNN**` 26개, 판정 대상 AC 제목 001–006·013–016·025–030(자리표시 007·017 별도).
- 보이지 않는 문자: `perl -CSD` `\p{Cf}` → SPEC 네 파일 0, 심어 둔 대조 파일 1.

## 추가 — 착수 전 선반영 N2·N3 (SPEC 0.2.4)

- 지시: 리드 — 운영자가 Implementation Kickoff 를 승인하면서 N1·N2·N3 선반영을 조건으로 둠. 레인은 N2·N3 각각 반례 하나로 판정이 바뀌는 것을 재현.
- 측정: 커밋 전 워킹 사본(HEAD `06c243b49` 위). 재현 직후 같은 세 파일을 명시 경로로 스테이징해 `ccfe3005e`(SPEC 0.2.4)로 커밋했다.

### N3 — AC-GDP-028/029 읽기 목록 선택식 (acceptance.md 805행)

반례 파일 세 줄: `` In team mode, `--auto-merge` merges only after all approvals are obtained. `` · `` `--auto-merge` merges as soon as CI checks pass. `` · `Unrelated line about release notes.`

| 선택식 | 뽑힌 줄 |
|---|---|
| 옛 선택식 `awk '/[Aa]pprov/ \|\| /[Rr]eview/ \|\| /[Tt]eam mode/ \|\| /[Pp]ersonal/ \|\| /[Mm]anual/'` | 1행만 — 모드 이름 없는 병합 조건(2행)을 놓침 |
| 새 선택식 (805행, `\|\| /--auto-merge/` 추가) | 1행, 2행 — 2행이 읽기 목록에 들어감 |

### N2 — 빈 파일이면 PASS 인 판정의 존재 확인

```
missing: test -e exit=1 (new check, expect 1 -> FAIL)
missing: test -s exit=1 (old check alone read 1 as PASS)
empty: test -e exit=0 (expect 0)
empty: test -s exit=1 (expect 1 -> PASS)
```

없는 파일은 새 존재 확인에서 FAIL(판정 불가)이 되고, 옛 빈 파일 검사만으로는 PASS 로 읽혔다. 판정 명령이 실제로 돌아 빈 파일을 남긴 경우는 두 검사를 모두 통과한다.

### 요구사항·수용 기준 불변

- `git diff -U0 HEAD -- acceptance.md spec.md` → 바뀐 줄 61개 중 `^[-+]\| AC-GDP-` 또는 `^[-+]\*\*REQ-GDP-` 형태 0개.
- 같은 패턴의 대조: 원본에서 `^\| AC-GDP-` 23줄, `^\*\*REQ-GDP-` 26줄이 잡힘 — 검출식이 표 행과 정의 줄을 실제로 잡는다.

## 미검증 (Gaps)

- D5–D8(표기·인용·환경 변수 안내·AC-016 기준 커밋)은 문구 수정이라 이 기록에서 반례를 돌리지 않았다. 3회차 감사 범위에 포함된다.
- AC-016 순서 반례는 커밋을 만들어야 해서 작성자도 레인도 실행하지 않았다.
- 판정식은 기록한 반례 계열만 막는다. 목록 밖 표현은 AC-026·028·029 의 읽기 기록이 떠받친다.
