<!--
RC1 PR 본문 양식 (00-E #2020 1차, 확장은 06-C #2007).
GitHub는 기본 브랜치에 있는 PR 템플릿만 자동으로 채운다. `v2026.11.0` 대상 PR에는 이 파일이
자동 적용되지 않으므로, 이 양식을 복사해 채운 뒤 `gh pr create --body-file <파일>`로 연다.
Issue는 `Refs #<n>`으로만 연결한다. `Closes`·`Fixes` 같은 자동 종료 키워드는 쓰지 않는다
(기본 브랜치가 아닌 곳의 merge는 Issue를 닫지 않고, 종료는 리더 통합 기록으로 판단한다).
-->

Refs #<Issue 번호>

## 대상과 claim

- 대상 Issue: #<n> (<논리 키, 예: 00-E>)
- claim 토큰: `Moai-Claim: <n>#<claim 댓글 id>`
- merge train: `<순번 또는 끼워 넣기(문서 전용)>` · base 병합은 자기 차례에 1회만 · 미리 쌓기라면 앞 PR #<n>의 감사된 head `<40자 SHA>` (#2388)
- base: `v2026.11.0` @ `<base 40자 SHA>`
- candidate head: `<head 40자 SHA>` (exact SHA, 독립 review가 본 SHA와 같아야 한다)
- SPEC: `<SPEC id>` · revision `<revision>` · digest `<digest 또는 null>`

## 변경

| 경로 | 무엇을 왜 |
| --- | --- |
| `<path>` | <설명> |

## 실행한 검사와 exit

| 명령 | exit | 요약 |
| --- | --- | --- |
| `cargo test --offline --locked` | <exit> | <통과 수> |
| `cargo clippy --offline --locked --all-targets -- -D warnings` | <exit> | |
| `cargo fmt --check` | <exit> | |
| `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p '*_test.py'` | <exit> | <Ran N> |
| `git diff --check` | <exit> | |
| <Issue의 AC 명령> | <exit> | <AC 번호와 결과> |

실행하지 않은 명령은 표에서 빼지 말고 exit 칸에 `not_run`과 이유를 적는다.

## 독립 review

- 판정: <PASS / FAIL> · 검토 세션: <세션 이름> · 본 SHA: `<40자 SHA>`
- 판정 댓글 링크: <https://github.com/modu-ai/moai-adk/issues/<n>#issuecomment-...>
- 재감사라면: 수리 확인 범위 <항목> · 새 지적 분류 <차단 / 후속 Issue #<n>> · 3회째 FAIL이면 범위 분할 상신 (#2388)

## not_run / blocked

- not_run: <항목과 이유>
- blocked: <항목, 기다리는 결정이나 HOLD>

## 남은 gap과 위험

- <관측하지 못한 것, 관측했지만 여전히 틀릴 수 있는 것>
