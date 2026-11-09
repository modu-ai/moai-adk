# 부록 4 — 시험 ID 대응표 (DR-07 #2143)

출처: review-group3 X14(시험 ID 8체계) + 상위 §17 15군 + RC1 receipt 실물(`docs/e2e-receipts/receipts.jsonl`). X14 결론 그대로: **HA-02 하나만 문서 ID가 RC1 receipt에 쓰였고**, 나머지 문서 ID는 대응 시험이 없거나 이름이 다르다.

## 8체계 → RC1 대응

| 문서 | ID 체계 | 예시 | RC1 대응 |
| --- | --- | --- | --- |
| [83f] (02-page) | `HOME-`·`SETUP-`·`GENERATION-` | HOME-, SETUP-, GENERATION- | 미대응 — RC1에서는 `project-plan/apply`·`generation-plan/apply`·e2e fixture가 같은 계약을 측정(T1 receipts) |
| [dff] (host adapter) | `HA-` | HA-01~10 | **HA-02만 RC1 receipt에 존재**(`receipts.jsonl` AC-08C-1~8과 병기 — update diff 0), HA-04/08은 T1 fixture가 계약 측정, HA-03/05/06/07 not_run(운영자 T2) |
| [25f] (릴리스·Desktop) | `RL-`·`DT-` | RL-COMMAND-CUTOVER, DT- | 미대응 — RL-10(RL-09)은 지원표 RL-10 산술로 실현(`tools/check_support_matrix.py`) |
| [e08] (PLAN) | `PLAN-` | PLAN-01~ | 미대응 — `plan-admit`(`Decision ready/stale/blocked`+사유)이 같은 계약 |
| [3d9] (RUN) | `RUN-` | RUN-01~09 | 부분 — `run-verify` receipt가 verify 절 수행 |
| [487] | `ST-` | ST- | 미대응 |
| [66a] (설치·UPDATE) | `GH-` | GH- | 미대응 — claim/e2e fixture가 같은 계약(claim journal, update-plan/apply) |
| 상위 §17 | 15개 시험군 | §17 표 | 부모 어휘 — 아래 "15군" 참조 |

## 상위 §17 15군 ↔ RC1 시험 파일

15군의 개별 ID는 상위 §17 표를 따르고, RC1에서 실제로 돌아가는 대응 물은 다음과 같다(모두 `cargo test --offline --locked` 소속 — `docs/e2e-matrix.md`·`docs/support-matrix.md` 참조):

| RC1 시험 파일 | 측정하는 것 | 대응 문서 ID |
| --- | --- | --- |
| `tests/e2e_host_fixture.rs` (T1, 5시험) | 중단 복구·변조 거절·재실행 diff 0·HOME 쓰기 차단·clean clone 자산 | HA-04·HA-08(T1 절반)·[66a] UPDATE 계약·AC-08C-1/2/7/8 |
| `tests/e2e/host-claude.sh`·`host-codex.sh` (T2) | 호스트 로딩·권한 거부 | HA-03/05/06/07 — **not_run**(운영자 실측) |
| `tests/install_cli.rs` | 설치 receipt·inspect·Go feed 격리 | [66a]·[83f] 설치 절 |
| `tests/update_cli.rs` | update plan/apply·오염 표 | [66a]·[83f] UPDATE 절 |
| `tests/generation_ref_slash.rs` | workflow_reference 정규화 | [83f] GENERATION- |
| `tests/product_cli.rs` | profile·agents·generation 전반 | [83f] HOME-/SETUP- 상당 |
| `tests/support_matrix_test.py` | RL-10 산술 검증기 | [25f] RL-10 |
| `tests/claim_*.rs`·`tests/recovery_cli.rs` | claim·recovery 계약 | [66a] GH- 상당 |

## 규칙

1. 새 receipt는 AC-08C-n 형식을 따른다(스키마 `schemas/e2e-receipt.schema.json`).
2. 문서 ID는 RC1 receipt로 번역될 때 이 표의 대응 열을 거친다 — 번호로 누락 검사하지 않는다(S8: 같은 F 번호가 다른 기능).
3. 대응이 없는 문서 ID는 "미대응"으로 남긴다 — 임의의 RC1 시험에 붙이지 않는다(가설 금지).
