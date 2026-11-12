# RC 릴리스 검토 (08-D #2015 — 독립 감사 양식)

verdict 어휘: **`release blocked`**(필수 항목 중 failed/blocked/not_run 존재 — REQ-08D-04) 또는 **`ready for operator decision`**(필수 항목 전부 실측, 게시는 여전히 운영자 결정).

## reviewer 기록 (REQ-08D-05 — 감사가 채운다)

| 필드 | 값 |
| --- | --- |
| reviewer | (독립 세션 명 — 미기입) |
| 공유 context 범위 | 이 저장소 워크트리; 구현 세션의 세션 메모리·대화·미커밋 상태는 읽지 않음 |
| audited SHA | (PR head — 감사 시점 실측) |
| 대조 digest | SPEC(Issue #2015 본문) · patch(diff) · `docs/e2e-receipts/receipts.jsonl` · `docs/e2e-receipts/T1-cargo-test-output.txt` · sha256.sum 출력 |
| 감사 일시 | (실측) |

## verdict

**`release blocked`** — 초기값(08-D 레인 기록, REQ-08D-04). 아래 gap이 운영자 결정·실측으로 해소되면 독립 감사가 verdict를 다시 쓴다. 레인이 스스로 verdict를 `ready for operator decision`으로 승격하지 않는다.

## gap 표 (감사가 갱신)

| # | gap | REQ/AC | 상태 | 해소 조건 |
| --- | --- | --- | --- | --- |
| 1 | release artifact·checksum 목록 미생성 | AC-08D-3 후단 / REQ-08D-02(d) | not_run | 게시 단계 artifact 생성 |
| 2 | attestation 미생성·`gh attestation verify` 미실행 | AC-08D-5 / REQ-08D-08 | HOLD/not_run | attestation workflow 운영 결정·실행 |
| 3 | `/releases/latest` 격리 실측 없음 | AC-08D-4 / REQ-08D-03 | not_run | Release 게시 후 확인 |
| 4 | `moai-cli-v*` tag ruleset 미결 | REQ-08D-09 | HOLD | 운영자 저장소 설정 |
| 5 | T2 호스트 로딩(3건) not_run — certified 지원 행 0 | REQ-08D-01/RL-10 | not_run | 운영자 자격증명 실측 |
| 6 | P12 OS matrix · SBOM · macOS 서명 | REQ-08D-01 등 | HOLD | 사용자 결정 |

## 대조 방법 (감사 절차)

1. PR head SHA와 본 문서 audited SHA 대조. 2. Issue #2015 REQ-08D-01..09 각각을 지원표·체크리스트 행과 대조. 3. `receipts.jsonl` 각 passed receipt를 지원표 지원 행과 대조(RL-10 산술 — `tools/check_support_matrix.py` 재실행). 4. not_run/HOLD가 숨김 없이 적혀 있는지 grep. 5. 필수 gap 잔존 시 verdict는 `release blocked` 유지.
