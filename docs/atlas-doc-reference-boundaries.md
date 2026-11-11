# #2450 C6d 수리 경계와 검증

범위: 감사 [6795423172](https://github.com/modu-ai/moai-adk/issues/2450#issuecomment-6795423172)의
D1·D2·D3. G-002만 생산하고 G-006/H1 HOLD는 유지한다. 기존 frozen fixture·truth·기준과
#2440 선택 계약은 바꾸지 않는다. 아래 상태는 저자의 시험 증거이며 독립 감사 판정이 아니다.

## 공식 근거와 채택/기각

공식 본문을 실제 읽고 별도 GET 원문 SHA256을 기록했다.

| 출처 | GET 확인 시각(UTC) | SHA256 |
| --- | --- | --- |
| [https://doc.rust-lang.org/reference/macros.html](https://doc.rust-lang.org/reference/macros.html) | 2026-11-11T12:14:04.925015+00:00 | `88ae535c579becdc4514772f26f874fdbdf6125fa99c16855b0eb4bc279698dc` |
| [https://doc.rust-lang.org/reference/items/modules.html](https://doc.rust-lang.org/reference/items/modules.html) | 2026-11-11T12:14:04.978586+00:00 | `4f1fa508cd687f7218ef0bc775adfa2bbbfcfc1f88247a675d262cbbcaab07ee` |
| [https://spec.commonmark.org/0.31.2/](https://spec.commonmark.org/0.31.2/) | 2026-11-11T12:14:05.445891+00:00 | `c85f56cc2101f761cc94215c677cd2c7c1ecca9c47be016479fdfe718af60f0d` |

- 채택: Rust Reference의 macro/attribute token tree를 선언으로 해석하지 않는다.
  세 종류 delimiter를 반복적 stack으로 추적해 선언·문서 후보를 가리고 기존 lexical call
  탐색은 유지한다. macro가 생성한 정의의 의미 분석·확장은 하지 않는다.
- 채택: doc 속성은 `#[path]` lookback에 투명하다. path attribute는 계속 지원 밖이며
  `module_path_attribute`를 남긴다. conventional decoy도 실제 path 파일도 따라가지 않는다.
- 채택: CommonMark §4.5·§5.1·§5.2를 근거로 fence의 marker·길이·열린 container prefix를
  보존한다. fence 내부의 추가 `>`/리스트 표식은 literal code이며 새 prefix로 재해석하지 않는다.
  block-comment 장식을 먼저 제거하되 Markdown indentation은 보존한다.
- 기각: macro가 있는 파일 전체 제외, 호출 추출기 비활성화, 모든 quote/list prose 제외,
  fixture/기준 하향, oracle 참조, 새 compiler/Markdown 의존성 도입. 양성 control로 실제 정의와
  container prose 및 fence 뒤 정상 참조가 살아 있음을 확인한다.
- extractor는 `code-refs=3`이다. 버전 2의 잘못 저장된 scan을 새 의미로 relabel/reuse하지 않는다.

## RED와 실제 소비 경로

`tests/doc_reference_boundaries.rs`의 D1~D3 fixture는 감사 본문의 리터럴이다.
각 case는 합성 git repository에서 실제 `index-build` → `atlas-graph`를 호출한다.
거부 coverage는 같은 generation의 native `index::index_records` API로 확인한다.
제품 프로세스는 별도 합성 HOME와 env allowlist를 사용한다.

수리 전 원본에서 D1/D2/D3 모두 fresh query의 false documents=1과 assertion failure를
직접 확인했다(`repair-red-literal.log`, exit101). 최초 helper의 존재하지 않는
`index-records` subcommand 오류는 준비 실패이며 RED·변이 검출 수에 포함하지 않는다.

수리 후 focused 실행(`repair-boundaries.log` 및 교차 경계의 `repair-cross-boundary-green.log`, 각각 exit0)에서 아래 셀을 모두 실제 CLI로 실행했다.
label은 같은 test 파일의 case id이며 각 행의 passed는 독립 PASS 판정이 아니다.

## D1 macro/attribute 경계

| cell | 불투명 입력 | 시험 결과 |
| --- | --- | --- |
| D1 literal | `const TEXT: &str = stringify!(fn parse() {});` | passed: edge0 + unresolved |
| M-paren | `discard!(fn parse() {});` | passed: edge0, 뒤 실제 정의 control edge1 |
| M-bracket | `discard![fn parse() {}];` | passed: 같은 음성·양성 control |
| M-brace | `discard!{fn parse() {}}` | passed: 같은 음성·양성 control |
| M-nested | parentheses → brackets → braces | passed: 같은 음성·양성 control |
| M-qualified | `tools::discard!(...)` | passed: 같은 음성·양성 control |
| M-attribute | attribute payload의 `fn parse` | passed: 같은 음성·양성 control |
| M-definition | `macro_rules!` transcriber의 `fn parse` | passed: 같은 음성·양성 control |

## D2 attribute/doc 순서와 기존 거부

| cell | 순서/형태 | 시험 결과 |
| --- | --- | --- |
| D2 literal / A-path-doc | path attribute → line doc → module | passed: decoy documents0, path 거부 유지 |
| A-doc-path | line doc → path attribute → module | passed: decoy imports/calls/documents 모두 없음 |
| A-mixed | doc → path → cfg attribute → doc → module | passed: 같은 거부·missing |
| A-block | path → 여러 줄 block doc → module | passed: 같은 거부·missing |
| A-restricted | path → doc → `pub(crate)` module | passed: 같은 거부·missing |
| A-normal | doc + 정상 conventional module | passed: documents1 양성 control |

## D3 container × fence marker

각 셀은 example에서 후보·edge·doc_reference_missing이 모두 0인 음성 control과,
정상 closing fence 뒤 prose에서 documents1인 양성 control을 함께 실행한다.

| container | tilde fence | backtick fence |
| --- | --- | --- |
| 없음 | F-plain-tilde: passed | F-plain-tick: passed |
| blockquote | F-quote-tilde / D3 literal: passed | F-quote-tick: passed |
| 중첩 blockquote | F-nested-tilde: passed | F-nested-tick: passed |
| bullet list | F-list-tilde: passed | F-list-tick: passed |
| ordered list | F-ordered: passed | F-ordered-tick: passed |
| list → quote | F-list-quote-tilde: passed | F-list-quote: passed |
| quote → list | F-quote-list: passed | F-quote-list-tick: passed |
| 중첩 list | F-nested-list-tilde: passed | F-nested-list: passed |

| 추가 cell | 경계 | 시험 결과 |
| --- | --- | --- |
| F-extra-quote / F-extra-list | fence body의 추가 container 유사 표식 | passed: 조기 종료·후속 false edge 없음 |
| F-plain-body-quote / F-plain-body-list | 일반 fence body의 quote/list 유사 표식 | passed: 같은 literal 처리 |
| F-short-close | opening보다 짧은 closing 유사 fence | passed: 후보·edge·doc missing0 |
| F-wrong-marker | opening과 다른 marker | passed: 같은 제외 |
| F-block-indent | block-comment 장식 뒤 indented 예시 | passed: 같은 제외 |
| F-block-fence / F-block-fence-tick | 장식된 block doc 안 quote fence | passed: 두 marker 모두 제외 |
| F-list-continuation | 앞 list item의 indentation을 유지한 fence | passed: 같은 제외 |
| F-attribute-between-docs / F-attribute-between-docs-tick | doc 주석 사이 일반 attribute, tilde/backtick | passed: fence 상태 보존, 후보·edge·doc missing0 |
| P-quote / P-nested-quote / P-list / P-ordered / P-list-quote | fence 밖 container prose | passed: 각 documents1, container 전체를 버리지 않음 |

## not_run과 제한

| 항목 | 상태와 이유 |
| --- | --- |
| 완전한 CommonMark 적합성·tab stop·HTML block 전체 조합 | not_run: 의존성 없는 lexical subset이며 이번은 위 space 기반 literal/container 경계만 검증 |
| macro expansion·procedural attribute의 생성/제거 semantics | not_run: compiler binding을 수행하지 않고 token tree를 불투명하게 처리 |
| static mut 이름 처리의 이전 자동 리뷰 제안 | not_run: 독립 감사 D1~D3의 이번 수리 범위 밖. 이 문서로 해결되었다고 세지 않음 |
| concept identity·alias / G-006 | blocked: 사용자 결정 H1 HOLD, 이유 있는 gap 유지 |
| live AI host·실제 사용자 HOME/키 | not_run: 합성 HOME/fixture만 허용된 범위 |

전체 검사·actual CLI 변이·exact-head CI 결과는 PR #2461 본문과 저자 인계 댓글에
head/base와 함께 기록한다. 위 focused 경계의 성공만으로 재감사 PASS를 주장하지 않는다.
