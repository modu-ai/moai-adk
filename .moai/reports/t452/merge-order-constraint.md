# t452 — 병합 순서 제약 (리드 확정)

이 문서는 **흡수·병합 시점에 읽으라고** 남긴다. run-phase 에서 할 일은 없다.

## 확정 사항

**t443 이 t452 보다 먼저 병합된다.** 창 순번에서 lane-8(t443)이 앞, lane-12(t452)가 마지막이다.

## 왜

두 브랜치가 같은 두 파일을 고친다.

| 카드 | 브랜치 | 겹치는 파일 |
|---|---|---|
| t443 | `WT-sync-auditor-derived` | `internal/template/catalog.yaml` · `internal/template/templates/.codex/agents/moai/sync-auditor.toml` |
| t452 | `WT-codex-skill-wiring` | 위 둘 + `internal/template/agentemit/agents-codex.yaml` |

t452 쪽의 두 파일 변경은 의무 명령(`make agents-emit` / `make build`)의 **부산물**이고,
t443 은 **그 수리가 카드 범위 전부**다. 소관 카드가 자기 수리를 착지시켜야 하며,
부산물이 먼저 들어가면 t443 이 진짜 no-op 이 된다.

## 흡수할 때 할 것

1. **커밋에서 그 파일들을 빼려 하지 않는다.** 되돌리면 AC-CSL-008 이 요구한 초록이
   워킹 트리에만 남는다. 귀속 기록(`m3-verdict.md` · `progress.md` §E.2)도 그대로 둔다.
2. t443 이 먼저 착지하면 그 두 파일은 develop 쪽이 정본이다. 충돌 시 **develop 값 채택**이
   기본 — 같은 수리라 내용이 수렴하기 때문이다.
3. **그래도 값이 다르면 강행하지 말고 보고한다.** `catalog.yaml` 해시는 whole-tree 함수라
   최종 병합 트리에서 재생성이 필요할 수 있다.
4. 흡수 후 **병합 트리에서** 세 가드를 다시 돌려 초록을 확인한다 —
   `TestCatalogHashParity` · `TestManifestHashFormat` · agentemit 골든.
   병합 전 초록을 병합 후 근거로 재사용하지 않는다.

## 참고 — 리드가 관측한 develop 상태

리드 실측: `develop`(당시 `bac2cf15b`)에서 `TestCatalogHashParity` 가 FAIL 이며
entry `sync-auditor` 해시가 드리프트했다. 이것이 t443 의 수리 대상이다.
이 값은 리드가 그 시점에 잰 것이고 develop 은 움직이는 ref 이므로,
흡수 시점에 **다시 재고** 이 줄을 근거로 쓰지 않는다.
