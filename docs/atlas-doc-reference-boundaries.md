# #2450 C6d 수리 경계와 검증

범위: 감사 [6795423172](https://github.com/modu-ai/moai-adk/issues/2450#issuecomment-6795423172)의
D1·D2·D3와 재감사 [6797284062](https://github.com/modu-ai/moai-adk/issues/2450#issuecomment-6797284062)의
N1·N2. G-002만 생산하고 G-006/H1 HOLD는 유지한다. 기존 frozen fixture·truth·기준과
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
- 채택: 리더 [6797331091](https://github.com/modu-ai/moai-adk/issues/2450#issuecomment-6797331091)의
  container 스택. quote/list item frame마다 fence·paragraph 상태와 상대 indentation을 소유한다.
  접두어/indent가 끝나면 해당 frame과 자손을 pop하고 같은 줄을 남은 바깥 문맥에서 처리한다.
  정상 paragraph만 lazy continuation을 허용하고 fence/code에는 허용하지 않는다.
  빈 줄이 없는 paragraph continuation과 실제 바깥 indented code를 구별한다.
- CommonMark §4.5의 containing-block 종료와 §5.1 예236·237·248 및 paragraph continuation을
  다시 읽었다. 2026-11-11T13:55:19.177209Z~13:55:19.514283Z GET 원문514698bytes,
  SHA256 `c85f56cc2101f761cc94215c677cd2c7c1ecca9c47be016479fdfe718af60f0d`.
- 기각: fence만 지우는 개별 조건 추가, quote의 list indentation을 바깥으로 승계,
  종료 줄을 버리거나 뒤 prose를 통째로 제외하는 수리. 이번에도 완전한 Markdown parser 도입은 하지 않는다.
- extractor는 `code-refs=4`이다. 버전 2/3의 잘못 저장된 scan을 새 의미로 relabel/reuse하지 않는다.

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

tab이나 HTML block을 만나면 완전한 문법을 추측하지 않고 해당 doc group의 나머지 후보를
보수적으로 제외한다. 알려진 fence 내부의 HTML 유사 문자열은 code 그대로 취급한다.
다음 source declaration으로 새 doc group이 시작되면 정상 참조 수집을 재개한다.
이 제외 경계의 소수 회귀는 전체 tab stop/HTML 적합성 검증을 뜻하지 않는다.

## Container 수명 전이

감사자의8입력은 원문 Rust string을 추출해 기대값·줄과 함께 byte 동일을 대조했다
(`lifetime-literal-proof.json`). 수정 전 실제 CLI/generation의 후보·stored scan·edge를 모두 읽고
N1 quote exit/list dedent·N2 quoted-list indent leak의3건 assertion 실패와5control 통과를
수집했다(`lifetime-red.log`, exit101). 첫 실패에서 loop를 끝내지 않는다.

`auditor_eight_container_lifetime_literals`와 `container_lifetime_transition_matrix`는 모든 셀을
실제 index-build → atlas-graph 및 같은 generation의 index::index_records로 대조한다.
정상 prose의 source/target·줄·SHA256, 후보/저장 후보/직접 edge 수와 missing0을 함께 확인한다.
Q=quote, L=list, QQ/LL/QL/LQ=2단계 중첩, b0/b1=종료 전 빈 줄 없음/있음이다.
root 행은 전체 container 종료, pop-child 행은 바깥 container를 남기고 자식만 종료한다.
indent 행은 열린 paragraph의 lazy continuation과 빈 줄 뒤 indentation 소유권을 구별한다.

자체 교차 검사에서 list 안 quote paragraph가 끝난 뒤 바깥 fence를 여는 상대 문맥도
추가했다. root 줄의 들여쓰기로 lazy 여부를 미리 정하면 안 되며 각 frame의 접두어를
제거한 현재 줄로 판단해야 한다. 4context(LQ/LLQ/QLQ/LQQ) × 상대0~3spaces × marker2 ×
빈 줄 유무의64입력을 실제 native CLI로 실행해16오탐·48control을 관측했고, 같은 입력의
Rust 시험에서도64셀을 끝까지 실행한 뒤 assertion exit101을 확인했다
(`relative-fence-rust-red.log`). 첫 native probe의 집계/command receipt 이름이 겹친 실행은
최종 집계에 쓰지 않고 이름을 분리한 재실행과 Rust RED를 보존했다.

최종 전이 셀은 감사 literal8 + 기존 스택 전이139 + 상대 fence64 + 새 list 전이40 =251개다.
새 container는 quote뿐 아니라 list도 시험하며, 새 list40셀에는 같은 pop/생산 제거 변이를
재적용한다. 이 추가 양성 control은 production source 변경 없이 확장했다.
어느 변이에서도 첫 실패로 loop를 중단하지 않고 전 셀 결과를 저장한다.
다음 독립 재감사에서 같은 부류의 새 차단이 나오면 리더6797331091에 따라 추가 수작업
문법 수리 대신 parser 의존성 채택/인식 범위 축소 중 사용자 결정을 요청한다.
두 대안 모두 이번에 임의 채택하지 않았다.

<!-- lifetime-cell-results:begin -->
셀별 실제 변이 대조는 진행 중이며 아직 이 절의 완료 상태를 주장하지 않는다.
<!-- lifetime-cell-results:end -->

전체 검사·actual CLI 변이·exact-head CI 결과는 PR #2461 본문과 저자 인계 댓글에
head/base와 함께 기록한다. 위 focused 경계의 성공만으로 재감사 PASS를 주장하지 않는다.
