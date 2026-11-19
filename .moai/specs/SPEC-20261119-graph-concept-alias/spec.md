---
id: SPEC-20261119-graph-concept-alias
title: "G-006 개념 이름 대상 관계: 명시 alias 계약으로 documents edge 생산 (C-17)"
status: draft
revision: 2
created: 2026-11-19
---

# G-006 개념 이름 대상 관계: 명시 alias 계약으로 documents edge 생산 (C-17)

## Intent

카드 C-17 (p2) 의 남은 건은 G-006 하나예요: GRAPH fixture 의 관계 대상이 파일이 아니라 개념 이름 `"handoff protocol"` 이라서, #2447 이후 '못 찾음 + 이유' gap 으로 남아 있었어요(H1 HOLD). 사용자가 결정 A(2026-11-15T10:16:15Z, 카드 노트 sha256 `6d567695372616c7e76ac95ec4b1304c86c2faa26ead6ee4b34b8f364cb1d0a8`)로 **명시 alias 계약(출처·digest 포함)** 을 채택했어요. 이 SPEC 은 그 계약을 제품 데이터로 두고, **코드 문자열 리터럴**의 계약 토큰 언급(`const HANDOFF_PREFIX: &str = "moai-claim-handoff:"`)을 계약으로 해석해 **기존 `documents` edge**(`file:src/claim_verify.rs → file:src/handoff.rs`)로 생산하며, GRAPH 어댑터를 리더 결정 6793749743(H2 승인)대로 typed target + 직접 documents 근거로 확장해 9/9 found·gap 0건으로 닫아요. 언급 채널이 문서 블록이 아니라 코드 문자열인 이유: frozen snapshot 에서 claim_verify.rs 의 언급 문서 블록(L11 목록 표식)은 plain-doc 계약(사용자 결정 6801346885)으로 블록 전체가 제외되어 문서 채널로는 생산이 불가능하고(D-5), 코드 문자열은 그 계약과 무관한 채널이에요. G-002 는 PR #2461 로 이미 병합돼 이 SPEC 범위가 아니에요.

## Non-goals

- 새 concept 노드 kind·공개 identity — ontology HOLD 가 유지되고, 결정 A 가 정하지 않은 것이라 기존 `file:` 노드와 `documents` edge 로만 표현해요.
- G-002 생산자의 재변경 — D1(매크로 불투명 토큰)·D2(`#[path]` 거부) 수리와 plain-doc 블록 계약(사용자 결정 6801346885)은 그대로 유지해요.
- 새 의존성(pulldown-cmark 등) — 사용자 결정 6801346885 로 기각된 길이에요.
- `tests/fixtures/atlas/` 의 fixture·ground truth·`baseline.json`·precision 기준 0.9212·#2440 `literal-evidence-v1` 선택기 변경 — 전부 불변.
- `moai-claim-release:` 등 다른 토큰의 계약 행 추가 — 계약 표는 정확히 1행이에요.
- 평가 oracle(G round truth)에서 alias 를 읽는 것 — 기존 기각 유지; 계약은 제품 데이터이고 어댑터는 generation 의 실제 edge 와 snapshot 바이트 digest 를 독립 대조해요.

## Requirements

- REQ-1: The `src/code_refs.rs` alias contract table shall carry exactly one row — label `handoff protocol`, token `moai-claim-handoff:`, definition file `src/handoff.rs`, confirm symbol `claim_handoff`, and a source string naming decision A of 2026-11-15T10:16:15Z and its record digest — as product data rather than an evaluation fixture.
- REQ-2: When the scanner reads a string literal token (plain, byte or raw) in a Rust file whose body starts with a contracted token, the scanner shall record one concept reference (stored scan kind `concept`) at that token's line carrying the token. A string inside an opaque macro token tree records nothing.
- REQ-3 (REQ-2): IF a string literal's body does not start with a contracted token, or the token appears only in a doc block, an ordinary comment or a doc backtick span, THEN the scanner shall store no concept reference for it.
- REQ-4: The stored-scan codec shall round-trip every `concept` row with a contracted token unchanged.
- REQ-5: The stored-scan codec shall return None for a `concept` row whose token is not a contract token.
- REQ-6: When resolving, the alias resolver shall produce one `documents` edge (kind `documents`, source the mentioning file, target the contract's definition file, both-ends line and digest evidence, `inferred`) only when the definition file is in the indexed file set and its scan declares exactly one item whose name equals the confirm symbol.
- REQ-7 (REQ-6): IF the definition file is absent from the indexed file set or declares zero or several items named by the confirm symbol, THEN the resolver shall record one `doc_reference_missing` diagnostic naming the token with reason `alias_definition_absent` or `alias_definition_ambiguous`, and no edge.
- REQ-8 (REQ-6): IF the mentioning file equals the definition file, THEN the resolver shall produce neither an edge nor a missing diagnostic.
- REQ-9: The stored extractor version shall rise from `5` to `6` so stored extractor-5 scans re-extract.
- REQ-10: The general imports/calls resolver shall ignore `concept` rows.
- REQ-11: The GRAPH adjudication in `tests/atlas_query_cli.rs` and `tests/search_graph.rs` shall treat a non-file target as a typed concept target resolved through the contract's public accessor, requiring for it a direct `documents` edge whose evidence digests match the snapshot's real file digests.
- REQ-12: The seven original code relations and G-002's exact edge shall keep their endpoints, kind, basis and evidence lines.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-6, REQ-7): frozen snapshot 의 실제 generation 에서 G-006 의 직접 근거가 `documents file:src/claim_verify.rs file:src/handoff.rs inferred L48,L165` 로 관측돼요.
  verify: `cargo test --locked --offline --test atlas_query_cli graph_cases_have_direct_code_reference_edges`
- AC-2 (REQ-11, REQ-12): GRAPH 9 케이스 전부 직접 근거로 found, gap 0건(검색 일반·expanded 양쪽)이고 `atlas-graph` 질의는 fresh 여요.
  verify: `cargo test --locked --offline --test atlas_query_cli graph_suite_relations_are_found_or_their_gap_is_stated`
  verify: `cargo test --locked --offline --test search_graph`
- AC-3 (REQ-2, REQ-3, REQ-4, REQ-5, REQ-9, REQ-10): 스캔·codec — 코드 문자열 리터럴 토큰만 개념 참조가 되고(plain·byte·raw, 매크로 불투명 영역 제외), 문서 블록·주석·비계약 토큰은 없으며, 알 수 없는 토큰 행은 decode 거부, 계약 토큰 행은 왕복 동일해요.
  verify: `cargo test --locked --offline --lib code_refs`
- AC-4 (REQ-7, REQ-8): 반례 — 정의 파일 부재, confirm 기호 불일치·복수, 언급 파일==정의 파일, 비계약 토큰(`moai-claim-release:`)에서 edge 0과 이유 있는 missing(자기 참조는 진단 없음)이 관측돼요.
  verify: `cargo test --locked --offline --test doc_references`
- AC-5: three-arm 계약 불변 — rg 재실행이 committed baseline 과 일치하고 세 arm 이 통과해요(답은 패턴이 파일 바이트에 있어야 선택되므로 handoff.rs 는 G-006 답에 안 들어가요).
  verify: `cargo test --locked --offline --test atlas_three_arm`
  verify: `cargo test --locked --offline --test atlas_query_cli`
- AC-6: 이웃 검사 — lib 전체·fmt·clippy·Python unittest·home_isolation_scan·trailer-check.
  verify: `cargo test --locked --offline --lib`
  verify: `cargo test --locked --offline --test home_isolation_scan`
  verify: `cargo fmt --check && cargo clippy --locked --offline --all-targets -- -D warnings`
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -I -m unittest discover -s tests -p '*_test.py'`(heavy-slot 안에서)
  verify: `~/.moai-cli/path/moai-cli trailer-check origin/v2026.11.0..HEAD`(failing 0)

## Scope

- write: `src/code_refs.rs`, `tests/atlas_query_cli.rs`, `tests/search_graph.rs`, `tests/doc_references.rs`, `docs/atlas-doc-reference-boundaries.md`, `.moai/specs/SPEC-20261119-graph-concept-alias/spec.md`
- read-only: `tests/fixtures/atlas/**`(불변), `src/ontology.rs`, `src/index_records.rs`(호출부 확인), `tools/atlas_three_arm.py`
- forbidden: `tests/fixtures/atlas/**` 쓰기, `docs/design-review/appendix/interfaces.md`(카드 C-58 예약 — 그 파일의 G-006 HOLD 문장 갱신은 C-58 몫), 실제 HOME·키·원장 수정, 강제 push·reset --hard·hooks 비활성화, 타 레인 예약 파일, 새 의존성

## Plan

1. RED: AC-1..AC-4 의 시험을 먼저 만들고 현 tip 코드에서 실패를 기록해요(합성 프로젝트와 frozen snapshot 만; 실제 checkout 파일 불변).
2. GREEN: `src/code_refs.rs` — 계약 표(REQ-1), 문자열 리터럴 스캔 수집(REQ-2·3), codec(REQ-4·5), `resolve_concepts`(REQ-6..8), extractor 버전·일반 해석 제외(REQ-9·10).
3. 어댑터·기대값(REQ-7): `tests/atlas_query_cli.rs`(typed target, 9/9, G-006 직접 근거)와 `tests/search_graph.rs`(matrix 9/9).
4. 문서: `docs/atlas-doc-reference-boundaries.md`(HOLD 문장을 계약 절로). `docs/design-review/appendix/interfaces.md` 의 HOLD 문장은 C-58 예약이라 이 SPEC 이 고치지 않고 C-58 에 남긴다.
5. AC-6 전체 검사 → push(push만, PR 없음) → 다른 문맥의 moai-audit 자체 감사(읽기 전용, 자기 clone) → 판정을 카드 노트로 → `LANE-READY-INTEGRATION`.

## Risks

- three-arm 정밀도 — 새 edge 가 G-006 답에 handoff.rs 를 추가하면 precision 이 떨질 수 있어요. 완화: 답 선택은 패턴이 파일 바이트에 있어야 하고(`select_answers` 의 `pattern in files[p]`) handoff.rs 바이트에 토큰이 없어 답이 불변인 것을 AC-5 가 확인해요.
- 어댑터 순환성 — 어댑터가 제품 계약을 읽으면 생산자와 같은 데이터를 쓰는 순환처럼 보일 수 있어요. 완화: 어댑터는 generation 의 실제 edge(evidence digest)를 snapshot 파일 digest 와 독립 대조하고, REQ-1 의 계약은 제품 데이터임을 SPEC 이 고정해요.
- 구 인덱스 호환 — extractor 5 저장 스캔은 버전 불일치로 재추출돼요(기존 버전 상승 패턴과 같음); 첫 빌드 비용이 커질 뿐 결과 동등은 증분=전체 시험이 확인해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 명시 alias 계약 표를 제품 데이터(`src/code_refs.rs`)로 채택 — 라벨 `handoff protocol` ↔ 토큰 `moai-claim-handoff:`, 정의 파일 `src/handoff.rs`, confirm `claim_handoff` | 사용자 결정 A(2026-11-15T10:16:15Z, sha256 `6d567695372616c7e76ac95ec4b1304c86c2faa26ead6ee4b34b8f364cb1d0a8`)·권고 6793132456·결정 시트 6813237309 #62 | 계약 행을 제거하면 해석이 missing 으로 복귀 |
| D-2 | 새 concept 노드 kind·공개 identity 없음 — 기존 `documents` edge 로만 표현 | 결정 A 가 정하지 않은 것(ontology HOLD 유지) | ontology HOLD 해제 시 별도 카드 |
| D-3 | 기대값 변경은 승인된 두 곳만: `hits 8 → 9`, `gaps ["G-006"] → []` (G-006 의 truth relation 자체와 나머지 8건의 기대값은 불변) | 결정 A + 리더 배차(2026-11-19, "남은 2건 관계 생산"); 카드 AC-3 의 "기대값 변경 없이"는 truth relation 기준으로 읽음(결정 1349 가 예고한 본문 정리) | 기대값 되돌리면 G-006 이 gap 으로 복귀 |
| D-4 | 계약 해석의 확인 조건: 정의 파일 존재 + confirm 기호와 정확히 일치하는 선언 1개 | 계약만으로 edge 를 만들지 않기 위한 실측 고리(H1 권고의 '근거' 요구) | confirm 조건 완화는 본 SPEC 개정으로만 |
| D-5 | 개념 언급 채널을 문서 블록에서 코드 문자열 리터럴로 변경(rev 1 → 2) | RED 실행이 드러낸 사실: frozen snapshot 의 claim_verify.rs 언급 문서 블록(L11 `- ` 목록 표식)은 사용자 결정 6801346885 로 블록 전체 제외라 문서 채널로는 생산 불가; 코드 문자열(`L48` const 선언)은 그 계약과 무관 | 문서 채널로 되돌리려면 plain-doc 계약 자체를 여는 사용자 결정이 필요 |
