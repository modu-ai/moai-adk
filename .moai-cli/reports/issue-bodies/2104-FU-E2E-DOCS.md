## 배경
- #2015(PR #2102) 레인 보고: docs/e2e-matrix.md가 `docs/host-support-matrix.md`를 가리키지만 실제 지원표는 `docs/support-matrix.md`다(명칭 불일치, 08-C 소유 문서).
- #2014(PR #2100) 감사 비차단: e2e_clean_clone_carries_portable_read_plan_assets가 git 메타데이터 없는 트리에서 "NOT RUN"을 stderr로만 남기고 return해 cargo가 ok로 집계한다(빈 sweep이 통과로 읽힘).

## 범위
- (r2 정정 — 전제 오류: 지원표가 둘이다. `docs/host-support-matrix.md` = Host 지원표(ZCode·Claude Code·Codex 절, tests/host_matrix_test.py가 검사), `docs/support-matrix.md` = 08-D receipt 기반 지원표.) **08-D 지원표를 가리키는 곳만** `docs/support-matrix.md`로 맞춘다 — docs/e2e-matrix.md, schemas/e2e-receipt.schema.json description. Host 지원표를 가리키는 곳(docs/zcode-support.md, docs/codex-adapter-contract.md 등)은 그대로 둔다.
- NOT RUN 경우가 통과로 읽히지 않게 한다 — 예: 그 조건을 `#[ignore = "..."]` 사유로 드러내거나 receipt/출력에 not_run 필드를 남긴다. 선택과 근거를 PR에 기록. git 저장소 안에서는 지금처럼 실행.

## 수락 기준
- AC-1 (r2 정정): 08-D 지원표를 뜻하는 곳에 `host-support-matrix`가 남지 않는다 — `grep -n 'host-support-matrix' docs/e2e-matrix.md schemas/e2e-receipt.schema.json` 0건, Host 지원표를 가리키는 docs/zcode-support.md·docs/codex-adapter-contract.md는 base와 같다. (원문: `grep -rn 'host-support-matrix' docs` 0건 — 두 표를 하나로 본 전제)
- AC-2: git 메타데이터 없는 트리에서 해당 시험이 passed로 집계되지 않는다(ignored 또는 명시 not_run, 관측 기록).
- AC-3: check set 전부 exit 0. AC-4: 독립 AUDIT PASS.
