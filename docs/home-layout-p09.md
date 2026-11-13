# P09 결정안 — 홈 트리 실물 정의 (`~/.moai-cli/projects/<project_id>/`, #2175)

상태: **결정안(제안)** — 기본값은 현행(프로젝트 쪽) 유지이며, 홈 이동은 opt-in이다. 확정은 운영자가 본 문서를 승인하는 것으로 한다(REQ-B8-01의 이동은 이 결정안의 구현이다).

## 홈 트리 실물 정의

```
~/.moai-cli/
├── installations/<install_id>.json        # 04-A (기존, 불변)
├── installations/<id>/hosts/<host>/known-good/<sha256>   # pre-A14 legacy store (preserved)
├── installations/<id>/hosts/moai/<host>/known-good/<sha256>   # A14 (#2286)
├── bin/moai-cli                            # 04-A (기존, 불변)
└── projects/<project_id>/                  # 이 결정안이 정의하는 새 실물
    ├── layout.json                         # 포인터 — 존재 자체가 opt-in 상태
    ├── receipts/<request-id>.json          # apply receipt (01-B)
    ├── forge/outbox/<op-id>.json           # outbox (07-B)
    └── claims/                             # claim journal (06-A: journal.log·<task_key>.json·.lock)
```

`<project_id>`는 `.moai-cli/project.yaml`의 `project_id`(profile)다. 같은 repo라도 profile의 project_id가 같으면 같은 홈 프로젝트 디렉터리를 공유한다.

## 이동 대상과 근거 위치(현행)

| store | 현행(프로젝트 쪽) | 홈 목표 | 코드 위치(tip 실측) |
|---|---|---|---|
| receipts | `<project>/.moai-cli/receipts/` | `projects/<id>/receipts/` | `src/apply.rs` `receipts_dir()` |
| outbox | `<project>/.moai-cli/forge/outbox/` | `projects/<id>/forge/outbox/` | `src/forge/outbox.rs` `outbox_dir()` |
| claim journal | `<git common dir>/.moai-cli/claims/` | `projects/<id>/claims/` | `src/claim.rs` `journal_root()` |

## opt-in 절차 (REQ-B8-01)

1. `moai layout-plan PROJECT` — 세 store를 측정(파일 수·바이트·`moai-tree-v1` digest)해 plan_digest로 묶는다. 아무것도 쓰지 않는다.
2. `moai layout-apply PROJECT --plan-digest <digest> --permission-reference <approved-ref>` — 권한 참조와 digest를 검증하고 claim journal 잠금을 잡는다. 잠금 전과 획득 뒤에 전환 상태·모든 홈 대상의 존재를 검사한다. 일치하면: (a) 원본을 **지우지 않고** 배타적으로 생성한 대상에 복사, (b) 대상 트리 digest가 원본과 일치하는지 검증, (c) 검증 통과 뒤 `layout.json` 포인터를 쓴다. 기존 대상은 덮어쓰거나 실패 정리로 삭제하지 않는다.
   - 원본의 디렉터리 열기·항목 조회·metadata·digest 읽기 오류는 store별로 집계해 plan/apply를 거부한다. 읽지 못한 store를 빈 것으로 확인하거나 건너뛰어 포인터를 전환하지 않는다.
3. 되돌리기 = `layout.json` 삭제. 프로젝트 쪽의 전환 시점 원본으로 복귀하며, 전환 중 홈에 새로 기록한 데이터는 홈에 남고 자동 병합하지 않는다. 기존 홈 대상이 남아 있으면 재적용을 거부하므로 새 데이터를 별도로 보존·병합한 뒤 다시 확인한다.

## 판독·기록 순서 (포인터 전환 뒤)

`layout.json`이 존재하는 동안 세 store의 읽기·쓰기는 홈 쪽을 같은 권위 경로로 해석한다. outbox의 읽기는 없는 디렉터리를 만들지 않으며 홈에 없는 항목을 프로젝트 쪽의 옛 항목으로 대신하지 않는다. 프로젝트 쪽은 전환 시점의 동결 이력으로 남아 직접 판독할 수 있다. 포인터를 지우면 프로젝트 쪽으로 즉시 복귀한다(REQ-B8-01).

## 알려진 제한

- claim journal은 클론 단위(common dir)였으나 홈은 프로젝트 단위다. 같은 project_id를 가진 두 클론이 모두 opt-in하면 홈 저널을 공유하게 되며, 이 경우 저널의 advisory lock이 직렬화한다. 단일 클론 가정이 RC1 기본이다.
- 원본이 남으므로 전환 뒤 프로젝트 쪽 저널은 동결 이력이다 — 양쪽에 동시 기록되는 일은 없다(포인터가 한쪽만 가리킴).
- 이 결정안은 RC1 기본값을 바꾸지 않는다. 기본은 프로젝트 쪽(사용자 결정, decisions)이며 홈 이동은 명시적 `layout-apply` 실행자의 책임으로 한다.
- claim journal이 전혀 없는 프로젝트의 첫 apply는 잠금 파일 초기화 뒤 plan이 달라져 `stale_plan`으로 거부될 수 있다. 다시 plan을 확인하면 해소되며, 이 경우도 포인터를 먼저 전환하지 않는다.
- 전환 상태에서 profile의 `project_id`를 바꾸면 이전 홈에 저장한 backup과 새 경로의 receipt가 갈릴 수 있다. 현재 복구는 backup이 없으면 오류로 끝난다. 이 부차 제한은 이번 세 차단 결함 수리에서 변경하지 않았다.
