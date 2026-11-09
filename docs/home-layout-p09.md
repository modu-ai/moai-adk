# P09 결정안 — 홈 트리 실물 정의 (`~/.moai-cli/projects/<project_id>/`, #2175)

상태: **결정안(제안)** — 기본값은 현행(프로젝트 쪽) 유지이며, 홈 이동은 opt-in이다. 확정은 운영자가 본 문서를 승인하는 것으로 한다(REQ-B8-01의 이동은 이 결정안의 구현이다).

## 홈 트리 실물 정의

```
~/.moai-cli/
├── installations/<install_id>.json        # 04-A (기존, 불변)
├── installations/<id>/hosts/<host>/known-good/<sha256>   # 04-B (기존, 불변)
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

1. `moai-cli layout-plan PROJECT` — 세 store를 측정(파일 수·바이트·`moai-tree-v1` digest)해 plan_digest로 묶는다. 아무것도 쓰지 않는다.
2. `moai-cli layout-apply PROJECT --plan-digest <digest>` — digest가 어긋나면 `stale_plan` 거부. 일치하면: (a) 원본을 **지우지 않고** 전 파일 복사, (b) 대상 트리 digest가 원본 digest와 일치하는지 검증(불일치 시 부분 복사를 지우고 실패), (c) 검증 통과 뒤 `layout.json` 포인터를 쓴다.
3. 되돌리기 = `layout.json` 삭제. 원본이 남아 있으므로 즉시 프로젝트 쪽으로 복귀하며 홈 복사본은 무해한 이력이다.

## 판독·기록 순서 (포인터 전환 뒤)

`layout.json`이 존재하는 동안 세 store의 읽기·쓰기는 홈 쪽을 우선한다. 프로젝트 쪽 디렉터리는 전환 시점의冻结 이력으로 남는다(읽기 호환 — 도구가 프로젝트 쪽을 직접 읽어도 전환 시점까지의 내용은 유효). 포인터를 지우면 프로젝트 쪽으로 즉시 복귀한다(REQ-B8-01).

## 알려진 제한

- claim journal은 클론 단위(common dir)였으나 홈은 프로젝트 단위다. 같은 project_id를 가진 두 클론이 모두 opt-in하면 홈 저널을 공유하게 되며, 이 경우 저널의 advisory lock이 직렬화한다. 단일 클론 가정이 RC1 기본이다.
- 원본이 남으므로 전환 뒤 프로젝트 쪽 저널은 동결 이력이다 — 양쪽에 동시 기록되는 일은 없다(포인터가 한쪽만 가리킴).
- 이 결정안은 RC1 기본값을 바꾸지 않는다. 기본은 프로젝트 쪽(사용자 결정, decisions)이며 홈 이동은 명시적 `layout-apply` 실행자의 책임으로 한다.
