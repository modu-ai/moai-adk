# #2450 C6d plain-doc 블록 계약과 검증

[사용자 결정6801346885](https://github.com/modu-ai/moai-adk/issues/2450#issuecomment-6801346885)이
이전 container 수명 해석을 대체한다. agent-17의
[빈 container 5개 반례](https://github.com/modu-ai/moai-adk/pull/2461#issuecomment-6800067035)를
또 다른 Markdown 상태 조건으로 수리하지 않는다. 새 Markdown 의존성도 추가하지 않는다.
G-002만 생산하고 G-006 concept identity/alias H1은 HOLD로 유지한다.
frozen fixture·ground truth·baseline·0.9212 기준·#2440 answer 선택기는 변경하지 않는다.

## 현재 계약

- 연속된 물리적 `///`·`//!` 줄 묶음, 또는 하나의 `/** */`·`/*! */`가 한 블록이다.
  주석이 없는 빈 줄, 일반 주석/소스/attribute, 별도 block comment는 경계다.
  문서 주석인 빈 `//!` 줄은 같은 블록에 속한다. Rustdoc의 속성 연결/이름 binding을 모방하지 않는다.
- 블록 전체를 먼저 읽는다. 한 줄에라도 quote, 목록, fence, 4칸 이상 indentation,
  tab, HTML 시작이 있으면 그 블록의 후보는 모두 0이다. 앞서 나온 평범한 참조도 제외한다.
  marker를 닫거나 container가 끝났다고 같은 블록의 뒤 prose를 복구하지 않는다.
- comment 표식 뒤 관례적 한 칸과 block 주석의 후속 줄 장식 `*`만 정규화한다.
  block 주석 첫 줄의 `*`는 내용이므로 목록 표식으로 남긴다.
  CRLF의 CR만 줄 종결자로 정규화하며 나머지 indentation과 tab은 보존한다.
- quote는 줄 앞 `>`, 목록은 `-`·`+`·`*` 또는 1~9자리 숫자 뒤 `.`·`)`와
  공백/줄끝, fence는 줄 앞 세 개 이상의 backtick/tilde다.
  4칸 indentation은 빈 줄도 제외하며 tab은 줄 안의 어느 위치든 제외한다.
  HTML은 줄 앞 tag-like 시작·comment·declaration·processing instruction을 보수적으로 제외한다.
  paragraph 안의 HTML marker와 `<https://example.test>` 같은 autolink는 HTML 블록 시작으로 세지 않는다.
- 위 제외 구문이 없는 블록에서만 backtick qualified 경로를 후보로 만든다.
  대상은 유일한 모듈 수준 선언이어야 하며 양쪽 줄/digest evidence를 보존한다.
  alias/re-export·associated item·compiler/rustdoc binding은 지원하지 않는다.

`DocState`·`DocFrame`·`DocContainer`·`DocFence` 및 pop/lazy continuation/상대 indent
수명 추적은 제거했다. 남는 것은 블록 경계와 정적인 제외 predicate뿐이다.

## 제외 계수와 저장

제외 블록마다 Graph의 기존 `missing` 경로에 다음 진단 하나를 기록한다.

```json
{"kind":"doc_reference_unsupported","path":"src/user.rs","line":1,"count":1,"reason":"unsupported doc block: blockquote"}
```

`count`의 합은 제외 블록 수다. 한 블록에 여러 marker가 있어도 한 번만 센다.
후보가 원래 없던 블록도 계수하므로 이것은 잃어버린 참조 수나 recall 비율이 아니다.
지원 밖 블록에는 `doc_reference_missing`을 만들지 않는다.
지원되는 경로의 정의 0개/복수 후보는 기존 unresolved/ambiguous 진단을 유지한다.
새 top-level 공개 필드나 edge kind는 추가하지 않는다.

`Scan.excluded_doc_blocks`는 block line·이유를 저장하고 canonical scan의 `x`행으로
encode/decode한다. 기존 generation의 저장 fragment를 재사용해도 계수는 유지되고,
plain 블록으로 바꾸면 해제된다. 증분/전체 build의 동일한 map/graph bytes를 시험한다.
extractor는 `code-refs=5`이며 이전 extractor fragment의 의미를 재라벨하지 않는다.

## D1·D2와 공식 근거

아래 원문은 앞 수리에서 실제 읽고 GET bytes의 SHA256을 기록했다. 이번에는 사용자
결정이 Markdown 인식 범위의 권위이며, 과거 CommonMark 수명 알고리즘을 다시 채택하지 않는다.

| 출처 | GET 확인 시각(UTC) | SHA256 |
| --- | --- | --- |
| [Rust Reference Macros](https://doc.rust-lang.org/reference/macros.html) | 2026-11-11T12:14:04.925015+00:00 | `88ae535c579becdc4514772f26f874fdbdf6125fa99c16855b0eb4bc279698dc` |
| [Rust Reference Modules](https://doc.rust-lang.org/reference/items/modules.html) | 2026-11-11T12:14:04.978586+00:00 | `4f1fa508cd687f7218ef0bc775adfa2bbbfcfc1f88247a675d262cbbcaab07ee` |
| [CommonMark0.31.2](https://spec.commonmark.org/0.31.2/) | 2026-11-11T12:14:05.445891+00:00 | `c85f56cc2101f761cc94215c677cd2c7c1ecca9c47be016479fdfe718af60f0d` |

채택 유지: macro/attribute token tree를 선언으로 해석하지 않는 D1과, Doc을 건너뛰어
`#[path]`를 거부하는 D2. macro 확장·procedural attribute semantics를 주장하지 않는다.
기각: 이전 Markdown container/fence 수명 구현, parser 의존성 추가, oracle 경로·별칭 하드코딩,
fixture·기대 기준 하향, 제외 블록의 뒤 prose만 다시 살리는 조건 추가.
과거 수명 표·측정은 [이전 d413f9b 문서](https://github.com/modu-ai/moai-adk/blob/d413f9b0d30002d046db2821c988a529be8ea7fc/docs/atlas-doc-reference-boundaries.md)에
보존되며 현재 계약의 PASS 근거로 승계하지 않는다.

## RED와 직접 검증 경계

- 감사 D3·N1 quote/list·N2·agent-17 빈 container5건과 뒤 marker의 소급 제외2건,
  합계11입력을 모두 실행한 후 새 계약 assertion FAILED(exit101)를 확인했다.
  N1의 뒤 prose도 이제 의도된 후보0·제외1 gap이다.
- 물리적 경계/목록/fence/tab/HTML/블록 장식·CRLF26입력을 추가했다.
  `/** * See ... */` 첫 줄을 장식으로 지우는 경우가 실제 CLI RED였고 첫 줄 내용을 보존해 수리했다.
- 이전251개 입력의 생성식·문자열 bytes는 유지했다. 사용자 결정에 따라 active expectation은
  대부분 후보0·제외1로 바뀐다. C-N1-plain-prose와 HTML/tab 뒤 별도 doc block reset2건,
  총3개 양성 control은 남긴다. 이전의 135 양성을 현재 계약의 양성이라고 세지 않는다.
- 각 case는 실제 native `index-build` → `atlas-graph`, 같은 generation의
  `index::index_records`, component/stored scan, 직접 documents·coverage·줄/digest를 비교한다.
  D1 opaque macro literal·괄호 이웃과 D2 path/doc 순서는 그대로 실행한다.
- 고정 snapshot의 G-002 문서 블록은 평범한 줄이며 제외 계수0이어야 한다.
  snapshot 전체 제외 블록 수도 별도 native records 시험에서 출력한다.
- 단순한 exclusion guard뿐 아니라 전체 블록 검사·생산·진단 계수를 각각 역변이한다.
  compile 오류나 시험0개를 killed로 세지 않고 실제 assertion failure와 정상/복원 control을 기록한다.

실제 명령·exit·source/binary SHA와 칸별 결과는 자기 `target/2450-evidence/narrow-*`에
보존하며 최종 전체 검사·benchmark 결과와 제외 블록 수는 PR 보고에 기록한다.
이 문서는 독립 감사 판정이 아니다.

## 남은 gap / not_run

제외 블록의 모든 참조는 이유 있는 coverage gap이다. 완전한 CommonMark/HTML/tab 적합성,
macro expansion, compiler/rustdoc binding, associated item, G-006 H1은 완료로 세지 않는다.
fresh clone·live AI host·실제 키·토큰 비용·전체 의미적 정확도는 not_run이다.
Linux CI의 apt DNS 설치 실패를 제품 시험 실패나 성공으로 바꾸어 기록하지 않는다.

## 251 historical input cells

Input bytes are unchanged. The user decision replaces the old document-count oracle; the following values are measured with the final plain-block source and native controls.

| Cell | Previous docs (historical) | Current docs | Excluded blocks | Native control | Killed operator |
| --- | ---: | ---: | ---: | --- | --- |
| `C-N1-closed-quote` | 1 | 0 | 1 | passed | exclusion |
| `C-N1-plain-prose` | 1 | 1 | 0 | passed | production |
| `C-N2-no-prior-list` | 0 | 0 | 1 | passed | exclusion |
| `C-N2-plain-indent` | 0 | 0 | 1 | passed | exclusion |
| `C-N2-valid-list-continuation` | 0 | 0 | 1 | passed | exclusion |
| `N1-list-dedent` | 1 | 0 | 1 | passed | exclusion |
| `N1-quote-exit` | 1 | 0 | 1 | passed | exclusion |
| `N2-quote-list-indent-leak` | 0 | 0 | 1 | passed | exclusion |
| `T-html-then-new-doc-group` | 1 | 1 | 1 | passed | exclusion,production |
| `T-indent-L-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-indent-L-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-indent-LL-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-indent-LL-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-indent-LQ-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-indent-LQ-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-indent-Q-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-indent-Q-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-indent-QL-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-indent-QL-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-indent-QQ-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-indent-QQ-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-new-list-child-LL-tick-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-child-LL-tick-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-child-LL-tilde-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-child-LL-tilde-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-child-LQ-tick-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-child-LQ-tick-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-child-LQ-tilde-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-child-LQ-tilde-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-child-QL-tick-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-child-QL-tick-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-child-QL-tilde-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-child-QL-tilde-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-child-QQ-tick-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-child-QQ-tick-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-child-QQ-tilde-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-child-QQ-tilde-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-L-tick-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-L-tick-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-L-tilde-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-L-tilde-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-LL-tick-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-LL-tick-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-LL-tilde-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-LL-tilde-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-LQ-tick-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-LQ-tick-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-LQ-tilde-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-LQ-tilde-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-Q-tick-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-Q-tick-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-Q-tilde-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-Q-tilde-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-QL-tick-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-QL-tick-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-QL-tilde-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-QL-tilde-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-QQ-tick-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-QQ-tick-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-QQ-tilde-b0` | 1 | 0 | 1 | passed | exclusion |
| `T-new-list-root-QQ-tilde-b1` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-LL-tick-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-pop-child-LL-tick-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-LL-tick-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-LL-tick-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-pop-child-LL-tick-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-LL-tick-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-LL-tilde-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-pop-child-LL-tilde-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-LL-tilde-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-LL-tilde-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-pop-child-LL-tilde-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-LL-tilde-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-LQ-tick-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-pop-child-LQ-tick-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-LQ-tick-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-LQ-tick-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-pop-child-LQ-tick-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-LQ-tick-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-LQ-tilde-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-pop-child-LQ-tilde-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-LQ-tilde-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-LQ-tilde-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-pop-child-LQ-tilde-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-LQ-tilde-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-QL-tick-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-pop-child-QL-tick-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-QL-tick-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-QL-tick-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-pop-child-QL-tick-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-QL-tick-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-QL-tilde-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-pop-child-QL-tilde-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-QL-tilde-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-QL-tilde-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-pop-child-QL-tilde-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-QL-tilde-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-QQ-tick-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-pop-child-QQ-tick-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-QQ-tick-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-QQ-tick-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-pop-child-QQ-tick-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-QQ-tick-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-QQ-tilde-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-pop-child-QQ-tilde-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-QQ-tilde-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-QQ-tilde-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-pop-child-QQ-tilde-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-pop-child-QQ-tilde-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-relative-LLQ-i0-tick-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LLQ-i0-tick-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LLQ-i0-tilde-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LLQ-i0-tilde-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LLQ-i1-tick-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LLQ-i1-tick-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LLQ-i1-tilde-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LLQ-i1-tilde-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LLQ-i2-tick-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LLQ-i2-tick-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LLQ-i2-tilde-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LLQ-i2-tilde-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LLQ-i3-tick-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LLQ-i3-tick-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LLQ-i3-tilde-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LLQ-i3-tilde-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQ-i0-tick-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQ-i0-tick-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQ-i0-tilde-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQ-i0-tilde-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQ-i1-tick-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQ-i1-tick-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQ-i1-tilde-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQ-i1-tilde-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQ-i2-tick-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQ-i2-tick-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQ-i2-tilde-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQ-i2-tilde-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQ-i3-tick-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQ-i3-tick-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQ-i3-tilde-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQ-i3-tilde-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQQ-i0-tick-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQQ-i0-tick-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQQ-i0-tilde-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQQ-i0-tilde-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQQ-i1-tick-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQQ-i1-tick-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQQ-i1-tilde-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQQ-i1-tilde-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQQ-i2-tick-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQQ-i2-tick-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQQ-i2-tilde-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQQ-i2-tilde-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQQ-i3-tick-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQQ-i3-tick-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQQ-i3-tilde-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-LQQ-i3-tilde-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-QLQ-i0-tick-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-QLQ-i0-tick-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-QLQ-i0-tilde-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-QLQ-i0-tilde-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-QLQ-i1-tick-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-QLQ-i1-tick-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-QLQ-i1-tilde-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-QLQ-i1-tilde-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-QLQ-i2-tick-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-QLQ-i2-tick-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-QLQ-i2-tilde-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-QLQ-i2-tilde-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-QLQ-i3-tick-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-QLQ-i3-tick-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-QLQ-i3-tilde-b0` | 0 | 0 | 1 | passed | exclusion |
| `T-relative-QLQ-i3-tilde-b1` | 0 | 0 | 1 | passed | exclusion |
| `T-root-L-tick-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-L-tick-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-L-tick-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-L-tick-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-L-tick-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-L-tick-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-L-tilde-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-L-tilde-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-L-tilde-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-L-tilde-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-L-tilde-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-L-tilde-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-LL-tick-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-LL-tick-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-LL-tick-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-LL-tick-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-LL-tick-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-LL-tick-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-LL-tilde-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-LL-tilde-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-LL-tilde-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-LL-tilde-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-LL-tilde-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-LL-tilde-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-LQ-tick-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-LQ-tick-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-LQ-tick-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-LQ-tick-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-LQ-tick-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-LQ-tick-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-LQ-tilde-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-LQ-tilde-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-LQ-tilde-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-LQ-tilde-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-LQ-tilde-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-LQ-tilde-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-Q-tick-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-Q-tick-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-Q-tick-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-Q-tick-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-Q-tick-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-Q-tick-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-Q-tilde-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-Q-tilde-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-Q-tilde-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-Q-tilde-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-Q-tilde-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-Q-tilde-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-QL-tick-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-QL-tick-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-QL-tick-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-QL-tick-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-QL-tick-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-QL-tick-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-QL-tilde-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-QL-tilde-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-QL-tilde-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-QL-tilde-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-QL-tilde-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-QL-tilde-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-QQ-tick-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-QQ-tick-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-QQ-tick-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-QQ-tick-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-QQ-tick-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-QQ-tick-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-QQ-tilde-b0-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-QQ-tilde-b0-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-QQ-tilde-b0-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-QQ-tilde-b1-code` | 0 | 0 | 1 | passed | exclusion |
| `T-root-QQ-tilde-b1-new-quote` | 1 | 0 | 1 | passed | exclusion |
| `T-root-QQ-tilde-b1-prose` | 1 | 0 | 1 | passed | exclusion |
| `T-root-fence-stays-tick` | 0 | 0 | 1 | passed | exclusion |
| `T-root-fence-stays-tilde` | 0 | 0 | 1 | passed | exclusion |
| `T-tab-then-new-doc-group` | 1 | 1 | 1 | passed | exclusion,production |
| `T-unsupported-html` | 0 | 0 | 1 | passed | exclusion |
| `T-unsupported-html-comment` | 0 | 0 | 1 | passed | exclusion |
| `T-unsupported-tab` | 0 | 0 | 1 | passed | exclusion |
